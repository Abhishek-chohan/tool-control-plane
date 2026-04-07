/**
 * Protocol types supported by the Toolplane client
 */
export enum ClientProtocol {
  GRPC = 'grpc'
}

/**
 * Tool definition for gRPC protocol
 */
export interface Tool {
  id: string;
  name: string;
  description: string;
  schema: string;
  config: Record<string, string>;
  createdAt: string;
  lastPingAt?: string;
  sessionId: string;
  tags: string[];
}

/**
 * Session definition for gRPC protocol
 */
export interface Session {
  id: string;
  name: string;
  description: string;
  createdAt: string;
  createdBy: string;
  apiKey: string;
  namespace: string;
}

/**
 * API key definition for gRPC protocol
 */
export interface ApiKey {
  id: string;
  name: string;
  key: string;
  keyPreview?: string;
  sessionId: string;
  createdAt: string;
  createdBy: string;
  capabilities: string[];
  revokedAt?: string;
}

/**
 * Machine definition for gRPC protocol
 */
export interface Machine {
  id: string;
  sessionId: string;
  sdkVersion: string;
  sdkLanguage: string;
  ip: string;
  createdAt: string;
  lastPingAt?: string;
}

/**
 * Task definition for gRPC protocol
 */
export interface Task {
  id: string;
  sessionId: string;
  toolName: string;
  status: string;
  input: string;
  result: string;
  resultType: string;
  error: string;
  createdAt: string;
  updatedAt: string;
  completedAt?: string;
}

/**
 * Request definition for gRPC protocol
 */
export interface Request {
  id: string;
  sessionId: string;
  toolName: string;
  status: string;
  input: string;
  createdAt: string;
  updatedAt: string;
  executingMachineId: string;
  result?: unknown;
  resultType?: string;
  error?: string;
  streamResults?: unknown[];
}

/**
 * Client configuration options
 */
export interface ClientConfig {
  protocol: ClientProtocol;
  serverHost: string;
  serverPort: number;
  sessionId: string;
  userId: string;
  apiKey?: string;
  timeout?: number;
  retryAttempts?: number;
  retryDelay?: number;
}

/**
 * Connection status
 */
export interface ConnectionStatus {
  connected: boolean;
  protocol: ClientProtocol;
  serverUrl: string;
  lastPing?: Date;
  error?: string;
}

/**
 * Request execution result
 */
export interface ExecutionResult<T = any> {
  success: boolean;
  result?: T;
  error?: string;
  requestId?: number | string;
  duration?: number;
}

/**
 * Tool registration request
 */
export interface RegisterToolRequest {
  sessionId: string;
  name: string;
  description: string;
  schema: string;
  config: Record<string, string>;
  tags: string[];
}

/**
 * Optional overrides for direct tool registration.
 */
export interface RegisterToolOptions {
  sessionId?: string;
  machineId?: string;
}

/**
 * Tool execution request
 */
export interface ExecuteToolRequest {
  sessionId: string;
  toolName: string;
  input: string;
}

/**
 * Tool execution response
 */
export interface ExecuteToolResponse {
  requestId: string;
  status: string;
  result: string;
  resultType: string;
  error?: string;
}

/**
 * Session creation request
 */
export interface CreateSessionRequest {
  userId: string;
  name: string;
  description: string;
  sessionId: string;
  namespace: string;
}

export interface CreateApiKeyOptions {
  name: string;
  capabilities?: string[];
}

/**
 * Machine registration request
 */
export interface RegisterMachineRequest {
  sessionId: string;
  machineId?: string;
  sdkVersion: string;
  sdkLanguage: string;
  tools?: RegisterToolRequest[];
}

/**
 * Health check response
 */
export interface HealthCheckResponse {
  status: string;
  version: string;
}

/**
 * Request list filter options.
 */
export interface RequestListOptions {
  status?: string;
  toolName?: string;
  limit?: number;
  offset?: number;
}

/**
 * Request update payload used by provider runtimes.
 */
export interface RequestUpdate {
  status?: string;
  result?: string;
  resultType?: string;
}

/**
 * Provider runtime configuration.
 */
export interface ProviderRuntimeOptions {
  pollIntervalMs?: number;
  heartbeatIntervalMs?: number;
  sdkVersion?: string;
}

/**
 * Provider session creation options.
 */
export interface ProviderSessionCreateOptions {
  sessionId?: string;
  name: string;
  description: string;
  namespace?: string;
  registerMachine?: boolean;
  machineId?: string;
  sdkVersion?: string;
}

/**
 * Provider session attach options.
 */
export interface ProviderSessionAttachOptions {
  registerMachine?: boolean;
  machineId?: string;
  sdkVersion?: string;
}

/**
 * Context passed to provider tool handlers.
 */
export interface ProviderToolContext {
  sessionId: string;
  requestId: string;
  toolName: string;
  machineId: string;
  input: Record<string, unknown>;
  appendChunk(chunk: unknown): Promise<void>;
  heartbeat(): Promise<Machine>;
}

export type ProviderToolResult =
  | unknown
  | Iterable<unknown>
  | AsyncIterable<unknown>
  | Promise<unknown | Iterable<unknown> | AsyncIterable<unknown>>;

export type ProviderToolHandler = (
  input: Record<string, unknown>,
  context: ProviderToolContext,
) => ProviderToolResult;

/**
 * Public provider tool registration.
 */
export interface ProviderToolRegistration {
  sessionId: string;
  name: string;
  description: string;
  handler: ProviderToolHandler;
  schema?: string | Record<string, unknown>;
  config?: Record<string, string>;
  tags?: string[];
  stream?: boolean;
}

/**
 * Session-scoped client surface required by the provider runtime.
 */
export interface ProviderRuntimeSessionClient {
  connect(): Promise<void>;
  disconnect(): Promise<void>;
  isConnected(): boolean;
  createSession(name: string, description: string, namespace?: string, requestedSessionId?: string): Promise<Session>;
  getSession(): Promise<Session>;
  registerMachine(machineId?: string, sdkVersion?: string, tools?: RegisterToolRequest[]): Promise<Machine>;
  registerTool(
    name: string,
    description: string,
    schema: string,
    config?: Record<string, string>,
    tags?: string[],
    options?: RegisterToolOptions,
  ): Promise<Tool>;
  listRequests(options?: RequestListOptions): Promise<Request[]>;
  claimRequest(requestId: string, machineId?: string): Promise<Request>;
  updateRequest(requestId: string, update: RequestUpdate): Promise<Request>;
  appendRequestChunks(requestId: string, chunks: unknown[], resultType?: string): Promise<boolean>;
  submitRequestResult(
    requestId: string,
    result: unknown,
    resultType?: string,
    meta?: Record<string, string>,
  ): Promise<boolean>;
  updateMachinePing(machineId?: string): Promise<Machine>;
  unregisterMachine(machineId?: string): Promise<boolean>;
  drainMachine(machineId?: string): Promise<boolean>;
}

/**
 * Root client surface required by the provider runtime.
 */
export interface ProviderRuntimeClient extends ProviderRuntimeSessionClient {
  forkSession(sessionId: string): ProviderRuntimeSessionClient;
}