import { McpError } from '@modelcontextprotocol/sdk/types.js';

/**
 * MCP 2026-07-28 stateless-core wire contract shared by the adapter's request
 * handlers. Mirrors server/pkg/mcp/protocol.go: per-request _meta carries the
 * protocol version and client capabilities (no initialize handshake), every
 * result carries a resultType discriminator, and the Tasks extension
 * (io.modelcontextprotocol/tasks) maps Toolplane's durable requests to
 * pollable task handles.
 */

export const PROTOCOL_VERSION = '2026-07-28';

export const TASKS_EXTENSION_ID = 'io.modelcontextprotocol/tasks';

export const META_PROTOCOL_VERSION = 'io.modelcontextprotocol/protocolVersion';

export const META_CLIENT_CAPABILITIES = 'io.modelcontextprotocol/clientCapabilities';

/** Chunk cursor sent by clients on tasks/get: "first chunk seq I still need". */
export const LAST_SEQ_META_KEY = 'dev.toolplane/last_seq';

/** Server-to-client chunk window carried in tasks/get _meta while working. */
export const CHUNKS_META_KEY = 'dev.toolplane/chunks';

/** Error data key carrying the Toolplane task (request) id. */
export const TASK_ID_DATA_KEY = 'toolplaneTaskId';

export const RESULT_TYPE_COMPLETE = 'complete';

export const RESULT_TYPE_TASK = 'task';

export const TASK_STATUS_WORKING = 'working';

export const TASK_STATUS_COMPLETED = 'completed';

export const TASK_STATUS_FAILED = 'failed';

export const TASK_STATUS_CANCELLED = 'cancelled';

/** Polling interval suggested to MCP clients in every task payload. */
export const TASK_POLL_INTERVAL_MS = 500;

export const ERROR_INVALID_REQUEST = -32600;

export const ERROR_METHOD_NOT_FOUND = -32601;

export const ERROR_INVALID_PARAMS = -32602;

export const ERROR_INTERNAL = -32603;

export const ERROR_UNSUPPORTED_PROTOCOL_VERSION = -32022;

/**
 * Exact error marker CancelRequest records on a cancelled Toolplane request;
 * distinguishes user cancellation from other failures (which surface as the
 * generic failed task status). Mirrors requestCancelledError in tasks.go.
 */
export const REQUEST_CANCELLED_ERROR = 'Request was cancelled';

export type TaskStatus =
  | typeof TASK_STATUS_WORKING
  | typeof TASK_STATUS_COMPLETED
  | typeof TASK_STATUS_FAILED
  | typeof TASK_STATUS_CANCELLED;

/**
 * Wire shape shared by the flat CreateTaskResult returned from tools/call
 * (resultType "task") and the DetailedTask variants returned by tasks/get
 * (resultType "complete"). Field names follow the Tasks extension schema.
 * The task identifier is a Toolplane request id: requests are the durable
 * units that provider machines claim and execute.
 */
export interface TaskPayload {
  resultType: string;
  taskId: string;
  status: TaskStatus;
  statusMessage?: string;
  createdAt: string;
  lastUpdatedAt: string;
  ttlMs: number | null;
  pollIntervalMs?: number;
  result?: Record<string, unknown>;
  error?: { code: number; message: string; data?: unknown };
  _meta?: Record<string, unknown>;
  [key: string]: unknown;
}

/** Retained chunk window for a working task, sliced at the client's cursor. */
export interface ChunkWindow {
  requestId: string;
  startSeq: number;
  nextSeq: number;
  chunks: unknown[];
}

/** Decoded, validated _meta of a single 2026 request. */
export interface RequestMeta {
  capabilities: Record<string, unknown>;
  lastSeq?: number;
}

export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

/**
 * Reports whether the params object carries a 2026 protocol version in _meta,
 * which switches a tools/list or tools/call handler into the stateless surface
 * (resultType discriminators, task handles). Params without it are served with
 * legacy semantics negotiated at initialize.
 */
export function has2026Meta(params: unknown): boolean {
  if (!isRecord(params) || !isRecord(params._meta)) {
    return false;
  }

  return META_PROTOCOL_VERSION in params._meta;
}

/**
 * Validates the stateless per-request metadata. The 2026-07-28 spec requires
 * protocolVersion and clientCapabilities on every request. Mirrors
 * parseRequestMeta in server/pkg/mcp/protocol.go.
 */
export function parseRequestMeta(params: unknown): RequestMeta {
  const meta = isRecord(params) && isRecord(params._meta) ? params._meta : undefined;
  if (!meta) {
    throw new McpError(
      ERROR_INVALID_REQUEST,
      `missing _meta: MCP 2026-07-28 is stateless and requires _meta with ${META_PROTOCOL_VERSION} and ${META_CLIENT_CAPABILITIES} on every request`,
    );
  }

  const version = typeof meta[META_PROTOCOL_VERSION] === 'string' ? meta[META_PROTOCOL_VERSION] : '';
  if (!version) {
    throw new McpError(
      ERROR_INVALID_REQUEST,
      `missing _meta.${META_PROTOCOL_VERSION}: every request must declare its MCP protocol version`,
    );
  }
  if (version !== PROTOCOL_VERSION) {
    throw new McpError(ERROR_UNSUPPORTED_PROTOCOL_VERSION, 'Unsupported protocol version', {
      supported: [PROTOCOL_VERSION],
      requested: version,
    });
  }

  const capabilities = meta[META_CLIENT_CAPABILITIES];
  if (!isRecord(capabilities)) {
    throw new McpError(
      ERROR_INVALID_REQUEST,
      `missing _meta.${META_CLIENT_CAPABILITIES}: every request must declare client capabilities (use {} for none)`,
    );
  }

  let lastSeq: number | undefined;
  if (LAST_SEQ_META_KEY in meta) {
    const cursor = meta[LAST_SEQ_META_KEY];
    if (typeof cursor !== 'number' || !Number.isInteger(cursor) || cursor < 0 || cursor > 2147483647) {
      throw new McpError(ERROR_INVALID_PARAMS, `_meta.${LAST_SEQ_META_KEY} must be a non-negative 32-bit integer`);
    }
    lastSeq = cursor;
  }

  return { capabilities, lastSeq };
}

/**
 * Reports whether the client advertised the Tasks extension in this request's
 * capabilities. Servers MUST NOT return a task handle to a client that did not
 * declare support. Mirrors requestMeta.clientSupportsTasks.
 */
export function clientSupportsTasks(capabilities: Record<string, unknown>): boolean {
  const extensions = capabilities.extensions;
  return isRecord(extensions) && TASKS_EXTENSION_ID in extensions;
}

export function isTerminalTaskStatus(status: TaskStatus): boolean {
  return (
    status === TASK_STATUS_COMPLETED ||
    status === TASK_STATUS_FAILED ||
    status === TASK_STATUS_CANCELLED
  );
}
