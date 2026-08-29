import {
  ClientProtocol,
  ToolplaneClient,
  type Request as ToolplaneRequest,
  type Session as ToolplaneSession,
  type Tool as ToolplaneTool,
} from 'toolplane-typescript-client';

import type {
  CallToolResult,
  ReadResourceResult,
  Resource,
  Tool as MCPTool,
} from '@modelcontextprotocol/sdk/types.js';

import type { AdapterOptions } from './config';
import { debugLog } from './debug';
import {
  CHUNKS_META_KEY,
  ERROR_INTERNAL,
  PROTOCOL_VERSION,
  REQUEST_CANCELLED_ERROR,
  RESULT_TYPE_COMPLETE,
  RESULT_TYPE_TASK,
  TASKS_EXTENSION_ID,
  TASK_ID_DATA_KEY,
  TASK_POLL_INTERVAL_MS,
  TASK_STATUS_CANCELLED,
  TASK_STATUS_COMPLETED,
  TASK_STATUS_FAILED,
  TASK_STATUS_WORKING,
  isTerminalTaskStatus,
  type TaskPayload,
  type TaskStatus,
} from './protocol';
import {
  ADAPTER_NAME,
  CONCEPT_MAP_MARKDOWN,
  CONCEPT_MAP_RESOURCE_URI,
  SESSION_RESOURCE_URI,
  parseRequestResourceId,
  requestResourceUri,
} from './resources';

import { McpError } from '@modelcontextprotocol/sdk/types.js';

type JSONObject = Record<string, unknown>;

/** Cadence for polling a request on the 2026 sync tools/call path. */
const SYNC_POLL_INTERVAL_MS = 250;

const ADAPTER_DISCOVER_INSTRUCTIONS =
  'Toolplane MCP adapter: durable tool execution bridged from a Toolplane session. ' +
  'A tools/call enqueues a durable request that a provider machine claims and executes, with ' +
  'retries and a retained streaming chunk window. Clients that advertise the ' +
  'io.modelcontextprotocol/tasks extension receive a task handle from tools/call and poll it with ' +
  'tasks/get; other clients are served synchronously.';

function describeError(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

function isRecord(value: unknown): value is JSONObject {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function renderValue(value: unknown): string {
  if (typeof value === 'string') {
    return value;
  }

  return JSON.stringify(value, null, 2);
}

function normalizeToolSchema(schemaText: string): { type: 'object'; [key: string]: unknown } {
  const fallback: { type: 'object'; [key: string]: unknown } = {
    type: 'object',
    properties: {},
  };

  const trimmed = schemaText.trim();
  if (!trimmed) {
    return fallback;
  }

  try {
    const parsed = JSON.parse(trimmed);
    if (!isRecord(parsed)) {
      return fallback;
    }

    if (!('type' in parsed) || parsed.type === undefined || parsed.type === 'object') {
      return {
        type: 'object',
        ...parsed,
      };
    }

    return {
      type: 'object',
      properties: {
        input: parsed,
      },
      required: ['input'],
    };
  } catch {
    return fallback;
  }
}

function sanitizeSession(session: ToolplaneSession): JSONObject {
  return {
    id: session.id,
    name: session.name,
    description: session.description,
    namespace: session.namespace,
    createdAt: session.createdAt,
    createdBy: session.createdBy,
    translatedBy: ADAPTER_NAME,
    nativeProtocol: 'toolplane',
    adaptedProtocol: 'mcp',
  };
}

function buildTranslationDetails(request: ToolplaneRequest): JSONObject {
  const streamResults = request.streamResults ?? [];

  return {
    sessionScope: 'one adapter process is bound to one native Toolplane session',
    requestLifecycle: 'the adapter waits for native request completion before responding to the MCP client',
    streaming:
      streamResults.length > 0
        ? 'native Toolplane stream chunks were aggregated into this MCP tool result'
        : 'no native stream chunks were emitted for this request',
  };
}

function buildRequestPayload(request: ToolplaneRequest): JSONObject {
  return {
    adapter: ADAPTER_NAME,
    requestId: request.id,
    sessionId: request.sessionId,
    toolName: request.toolName,
    status: request.status,
    input: request.input,
    result: request.result ?? null,
    resultType: request.resultType ?? null,
    error: request.error ?? null,
    streamResults: request.streamResults ?? [],
    createdAt: request.createdAt,
    updatedAt: request.updatedAt,
    executingMachineId: request.executingMachineId,
    translation: buildTranslationDetails(request),
  };
}

function buildRequestSummary(request: ToolplaneRequest): string {
  const streamResults = request.streamResults ?? [];
  const lines: string[] = [
    `Toolplane request ${request.id} completed with status ${request.status}.`,
  ];

  if (streamResults.length > 0) {
    lines.push(`Aggregated ${streamResults.length} native stream chunk(s) into this MCP tool response.`);
  }

  if (request.result !== undefined) {
    lines.push(`Native result:\n${renderValue(request.result)}`);
  }

  if (request.error) {
    lines.push(`Native error: ${request.error}`);
  }

  return lines.join('\n\n');
}

function firstNonEmpty(...values: Array<string | undefined>): string {
  for (const value of values) {
    if (value !== undefined && value.trim() !== '') {
      return value;
    }
  }
  return '';
}

/**
 * Maps a Toolplane request status to an MCP Tasks status.
 * pending/claimed/running/stalled all mean work is in flight; MCP has a
 * single working state. A failure caused by CancelRequest maps to cancelled;
 * every other failure maps to failed. Mirrors requestTaskStatus in
 * server/pkg/mcp/tasks.go.
 */
function requestTaskStatus(request: ToolplaneRequest): TaskStatus {
  switch (request.status) {
    case 'done':
      return TASK_STATUS_COMPLETED;
    case 'failure':
      return request.error === REQUEST_CANCELLED_ERROR ? TASK_STATUS_CANCELLED : TASK_STATUS_FAILED;
    default:
      return TASK_STATUS_WORKING;
  }
}

function statusMessageForRequest(request: ToolplaneRequest): string {
  switch (request.status) {
    case 'pending':
      return 'queued for execution';
    case 'claimed':
    case 'running':
      return 'executing';
    case 'stalled':
      return 'stalled, awaiting lease reclaim';
    case 'failure':
      return firstNonEmpty(request.error, request.resultType);
    case 'cancelled':
      return REQUEST_CANCELLED_ERROR;
    default:
      return '';
  }
}

/** Fields every task payload carries, before the resultType is set. */
interface BaseTaskFields {
  taskId: string;
  status: TaskStatus;
  statusMessage?: string;
  createdAt: string;
  lastUpdatedAt: string;
  ttlMs: number | null;
  pollIntervalMs?: number;
}

/**
 * Fills the fields every task payload carries. Toolplane retains requests
 * indefinitely, so ttlMs is null (unlimited). Mirrors baseTaskPayload.
 */
function baseTaskPayload(request: ToolplaneRequest): BaseTaskFields {
  const statusMessage = statusMessageForRequest(request);
  return {
    taskId: request.id,
    status: requestTaskStatus(request),
    statusMessage: statusMessage || undefined,
    createdAt: request.createdAt,
    lastUpdatedAt: request.updatedAt,
    ttlMs: null,
    pollIntervalMs: TASK_POLL_INTERVAL_MS,
  };
}

export class ToolplaneMcpBridge {
  private readonly client: ToolplaneClient;

  private readyPromise?: Promise<void>;

  private sessionId = '';

  private recentRequestIds: string[] = [];

  constructor(private readonly options: AdapterOptions) {
    this.client = new ToolplaneClient({
      protocol: ClientProtocol.GRPC,
      serverHost: options.grpcHost,
      serverPort: options.grpcPort,
      sessionId: options.sessionId ?? '',
      userId: options.userId,
      apiKey: options.apiKey,
      timeout: options.timeoutMs,
    });
  }

  async connect(): Promise<void> {
    await this.ensureReady();
  }

  async close(): Promise<void> {
    await this.client.disconnect();
  }

  async listTools(): Promise<MCPTool[]> {
    await this.ensureReady();

    const tools = await this.client.listTools();
    return tools
      .slice()
      .sort((left, right) => left.name.localeCompare(right.name))
      .map((tool) => this.toMcpTool(tool));
  }

  async callTool(name: string, args: Record<string, unknown> = {}): Promise<CallToolResult> {
    await this.ensureReady();

    try {
      const request = await this.client.executeTool(name, args);
      this.rememberRequest(request.id);
      return this.buildCallToolResult(request);
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);

      return {
        content: [
          {
            type: 'text',
            text: `Toolplane tool call failed before a translated request record could be returned. ${message}`,
          },
        ],
        structuredContent: {
          adapter: ADAPTER_NAME,
          error: message,
          nativeProtocol: 'toolplane',
          adaptedProtocol: 'mcp',
        },
        isError: true,
      };
    }
  }

  /**
   * Returns the server/discover advertisement required by the 2026-07-28
   * stateless core: supported versions plus capabilities, including the Tasks
   * extension. Mirrors handleDiscover in server/pkg/mcp/tools.go.
   */
  discoverPayload(): JSONObject {
    return {
      resultType: RESULT_TYPE_COMPLETE,
      ttlMs: 0,
      supportedVersions: [PROTOCOL_VERSION],
      capabilities: {
        tools: {},
        extensions: {
          [TASKS_EXTENSION_ID]: {},
        },
      },
      instructions: ADAPTER_DISCOVER_INSTRUCTIONS,
    };
  }

  /**
   * tools/list for 2026 clients: plain resultType-bearing result with the
   * gateway's tool entry shape (name, inputSchema, optional description).
   */
  async listTools2026(): Promise<JSONObject> {
    await this.ensureReady();

    const tools = await this.client.listTools();
    const entries = tools
      .slice()
      .sort((left, right) => left.name.localeCompare(right.name))
      .map((tool) => {
        const entry: JSONObject = {
          name: tool.name,
          inputSchema: normalizeToolSchema(tool.schema),
        };
        if (tool.description.trim() !== '') {
          entry.description = tool.description;
        }
        return entry;
      });

    return {
      resultType: RESULT_TYPE_COMPLETE,
      ttlMs: 0,
      tools: entries,
    };
  }

  /**
   * tools/call for 2026 clients that advertise the Tasks extension: enqueue a
   * pending request for a provider machine to claim and return the task handle
   * immediately. The request id is the task id. Mirrors handleToolsCall in
   * server/pkg/mcp/tools.go.
   */
  async createTaskHandle(name: string, args: Record<string, unknown> = {}): Promise<TaskPayload> {
    await this.ensureReady();

    const request = await this.client.createRequest(name, JSON.stringify(args));
    this.rememberRequest(request.id);

    return {
      ...baseTaskPayload(request),
      resultType: RESULT_TYPE_TASK,
    };
  }

  /**
   * tools/call for 2026 clients that did not advertise the Tasks extension:
   * enqueue a pending request, then poll it to a terminal state and aggregate
   * the outcome into one CallToolResult, mirroring awaitSyncResult. A timeout
   * surfaces the still-running task id so the client can switch to polling.
   */
  async callToolSync(name: string, args: Record<string, unknown> = {}): Promise<JSONObject> {
    await this.ensureReady();

    const request = await this.client.createRequest(name, JSON.stringify(args));
    this.rememberRequest(request.id);

    const terminal = await this.waitForTerminalRequest(request.id);
    return this.decorateSyncResult(terminal);
  }

  /**
   * tasks/get: the DetailedTask payload for the request's current status.
   * While the task is working the retained chunk window is attached under
   * _meta, sliced at the client's last_seq cursor. Mirrors handleTasksGet.
   */
  async getTask(taskId: string, lastSeq?: number): Promise<TaskPayload> {
    await this.ensureReady();

    const request = await this.client.getRequest(taskId);
    this.rememberRequest(request.id);

    const payload: TaskPayload = {
      ...baseTaskPayload(request),
      resultType: RESULT_TYPE_COMPLETE,
    };

    switch (payload.status) {
      case TASK_STATUS_COMPLETED:
        payload.result = {
          ...this.buildCallToolResult(request),
          resultType: RESULT_TYPE_COMPLETE,
        };
        break;
      case TASK_STATUS_FAILED:
        payload.error = {
          code: ERROR_INTERNAL,
          message: firstNonEmpty(request.error, 'tool execution failed'),
          data: { [TASK_ID_DATA_KEY]: request.id },
        };
        break;
      default:
        break;
    }

    if (!isTerminalTaskStatus(payload.status)) {
      const chunksMeta = await this.readChunkWindow(request.id, lastSeq);
      if (chunksMeta) {
        payload._meta = chunksMeta;
      }
    }

    return payload;
  }

  /**
   * tasks/cancel: cooperative cancellation of the underlying request. A task
   * already in a terminal state is acknowledged rather than rejected. Throws
   * the client's not-found error for unknown task ids. Mirrors
   * handleTasksCancel.
   */
  async cancelTask(taskId: string): Promise<void> {
    await this.ensureReady();

    const request = await this.client.getRequest(taskId);
    if (isTerminalTaskStatus(requestTaskStatus(request))) {
      return;
    }

    await this.client.cancelRequest(taskId);
  }

  /**
   * Renders the terminal request as the adapter's translated CallToolResult:
   * a textual summary, a request resource link, and the full translated
   * request payload as structuredContent.
   */
  private buildCallToolResult(request: ToolplaneRequest): CallToolResult {
    const structuredContent = buildRequestPayload(request);
    const isError = request.status !== 'done' || Boolean(request.error);

    return {
      content: [
        {
          type: 'text',
          text: buildRequestSummary(request),
        },
        {
          type: 'resource_link',
          uri: requestResourceUri(request.id),
          name: `toolplane-request-${request.id}`,
          title: `Toolplane request ${request.id}`,
          mimeType: 'application/json',
          description: 'Translated Toolplane request record for this MCP tool call',
        },
      ],
      structuredContent,
      isError: isError || undefined,
    };
  }

  /**
   * Polls the request until it reaches a terminal task status or the adapter
   * timeout elapses. Mirrors the gateway's awaitSyncResult polling loop.
   */
  private async waitForTerminalRequest(requestId: string): Promise<ToolplaneRequest> {
    const deadline = Date.now() + this.options.timeoutMs;

    for (;;) {
      const request = await this.client.getRequest(requestId);
      if (isTerminalTaskStatus(requestTaskStatus(request))) {
        return request;
      }

      if (Date.now() >= deadline) {
        throw new McpError(
          ERROR_INTERNAL,
          `tool call did not complete within ${this.options.timeoutMs}ms; the request is still running`,
          { [TASK_ID_DATA_KEY]: requestId },
        );
      }

      await new Promise((resolve) => {
        setTimeout(resolve, SYNC_POLL_INTERVAL_MS);
      });
    }
  }

  /**
   * Adds the 2026 resultType discriminator and task id to a sync tool result,
   * surfacing retained stream chunks as leading text blocks and in _meta.
   * Mirrors buildSyncCallToolResult.
   */
  private async decorateSyncResult(request: ToolplaneRequest): Promise<JSONObject> {
    const result: JSONObject = {
      ...this.buildCallToolResult(request),
      resultType: RESULT_TYPE_COMPLETE,
    };

    const meta: JSONObject = { [TASK_ID_DATA_KEY]: request.id };
    try {
      const window = await this.client.getRequestChunksWindow(request.id);
      if (window.chunks.length > 0) {
        const content = Array.isArray(result.content) ? result.content : [];
        const chunkBlocks = window.chunks.map((chunk) => ({ type: 'text', text: renderValue(chunk) }));
        result.content = [...chunkBlocks, ...content];
        meta[CHUNKS_META_KEY] = {
          requestId: request.id,
          startSeq: window.startSeq,
          nextSeq: window.nextSeq,
          chunks: window.chunks,
        };
      }
    } catch (error) {
      debugLog(`sync chunk window read failed for ${request.id}: ${describeError(error)}`);
    }

    result._meta = meta;
    return result;
  }

  /**
   * Best-effort chunk-window read for a working task: read failures degrade to
   * a payload without chunks. Mirrors readChunkWindow in server/pkg/mcp.
   */
  private async readChunkWindow(requestId: string, lastSeq?: number): Promise<JSONObject | undefined> {
    try {
      const window = await this.client.getRequestChunksWindow(requestId);

      let startSeq = window.startSeq;
      let chunks = window.chunks;
      if (lastSeq !== undefined && lastSeq > startSeq) {
        const drop = lastSeq - startSeq;
        if (drop >= chunks.length) {
          chunks = [];
          startSeq = window.nextSeq;
        } else {
          chunks = chunks.slice(drop);
          startSeq = lastSeq;
        }
      }

      return {
        [CHUNKS_META_KEY]: {
          requestId,
          startSeq,
          nextSeq: window.nextSeq,
          chunks,
        },
      };
    } catch (error) {
      debugLog(`chunk window read failed for ${requestId}: ${describeError(error)}`);
      return undefined;
    }
  }

  async listResources(): Promise<Resource[]> {
    const session = await this.refreshSession();
    const resources: Resource[] = [
      {
        uri: SESSION_RESOURCE_URI,
        name: `toolplane-session-${session.id}`,
        title: 'Current Toolplane session',
        description: 'Sanitized Toolplane session context bound to this adapter process',
        mimeType: 'application/json',
      },
      {
        uri: CONCEPT_MAP_RESOURCE_URI,
        name: 'toolplane-mcp-concept-map',
        title: 'Toolplane to MCP concept map',
        description: 'Native versus translated semantics for the optional MCP adapter',
        mimeType: 'text/markdown',
      },
      ...this.recentRequestIds.map((requestId) => ({
        uri: requestResourceUri(requestId),
        name: `toolplane-request-${requestId}`,
        title: `Toolplane request ${requestId}`,
        description: 'Translated request record returned by a recent MCP tool call',
        mimeType: 'application/json',
      })),
    ];

    return resources;
  }

  async readResource(uri: string): Promise<ReadResourceResult> {
    const normalizedUri = uri.trim();

    if (normalizedUri === CONCEPT_MAP_RESOURCE_URI) {
      return {
        contents: [
          {
            uri: CONCEPT_MAP_RESOURCE_URI,
            mimeType: 'text/markdown',
            text: CONCEPT_MAP_MARKDOWN,
          },
        ],
      };
    }

    if (normalizedUri === SESSION_RESOURCE_URI) {
      const session = await this.refreshSession();
      return {
        contents: [
          {
            uri: SESSION_RESOURCE_URI,
            mimeType: 'application/json',
            text: JSON.stringify(sanitizeSession(session), null, 2),
          },
        ],
      };
    }

    const requestId = parseRequestResourceId(normalizedUri);
    if (requestId) {
      await this.ensureReady();
      const request = await this.client.getRequest(requestId);
      this.rememberRequest(request.id);

      return {
        contents: [
          {
            uri: requestResourceUri(request.id),
            mimeType: 'application/json',
            text: JSON.stringify(buildRequestPayload(request), null, 2),
          },
        ],
      };
    }

    throw new Error(`Unsupported resource URI: ${uri}`);
  }

  private async ensureReady(): Promise<void> {
    if (!this.readyPromise) {
      this.readyPromise = this.initialize();
    }

    await this.readyPromise;
  }

  private async initialize(): Promise<void> {
    debugLog(`connecting to Toolplane at ${this.options.grpcHost}:${this.options.grpcPort}`);
    await this.client.connect();

    if (this.options.sessionId) {
      this.sessionId = this.options.sessionId;
      debugLog(`binding adapter to existing session ${this.sessionId}`);
      await this.client.getSession();
      debugLog(`verified session ${this.sessionId}`);
      return;
    }

    const session = await this.client.createSession(
      this.options.sessionName,
      this.options.sessionDescription,
      this.options.sessionNamespace,
    );

    this.sessionId = session.id;
    debugLog(`created session ${this.sessionId}`);
  }

  private async refreshSession(): Promise<ToolplaneSession> {
    await this.ensureReady();
    return this.client.getSession();
  }

  private toMcpTool(tool: ToolplaneTool): MCPTool {
    return {
      name: tool.name,
      title: tool.name,
      description: tool.description,
      inputSchema: normalizeToolSchema(tool.schema),
    };
  }

  private rememberRequest(requestId: string): void {
    this.recentRequestIds = [
      requestId,
      ...this.recentRequestIds.filter((candidate) => candidate !== requestId),
    ].slice(0, this.options.requestResourceLimit);
  }
}