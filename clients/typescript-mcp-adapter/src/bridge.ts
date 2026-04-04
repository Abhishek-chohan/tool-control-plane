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
  ADAPTER_NAME,
  CONCEPT_MAP_MARKDOWN,
  CONCEPT_MAP_RESOURCE_URI,
  SESSION_RESOURCE_URI,
  parseRequestResourceId,
  requestResourceUri,
} from './resources';

type JSONObject = Record<string, unknown>;

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