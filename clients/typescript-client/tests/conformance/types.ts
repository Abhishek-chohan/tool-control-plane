export type Transport = 'http' | 'grpc';

export type SupportedFeature =
  | 'session_create'
  | 'session_list'
  | 'invoke_unary'
  | 'invoke_stream'
  | 'tool_discovery'
  | 'session_update'
  | 'request_create'
  | 'request_recovery'
  | 'api_key_lifecycle'
  | 'machine_lifecycle'
  | 'provider_runtime';

export interface ConformanceCase {
  id: string;
  feature: SupportedFeature;
  description: string;
  request: Record<string, unknown>;
  expected: Record<string, unknown>;
  tags?: string[];
}

export interface ConformanceAdapter {
  connect(): Promise<void>;
  close(): Promise<void>;
  createSession(request: Record<string, unknown>): Promise<string>;
  getSessionContext(sessionId: string): Promise<Record<string, unknown> | null>;
  updateSession(sessionId: string, request: Record<string, unknown>): Promise<Record<string, unknown>>;
  listUserSessions(request: Record<string, unknown>): Promise<Record<string, unknown>>;
  registerUnaryEchoTool(sessionId: string, toolName: string, description: string): Promise<void>;
  registerStreamTool(sessionId: string, toolName: string, description: string): Promise<void>;
  listTools(sessionId: string): Promise<Record<string, unknown>[]>;
  getToolById(sessionId: string, toolId: string): Promise<Record<string, unknown>>;
  getToolByName(sessionId: string, toolName: string): Promise<Record<string, unknown>>;
  deleteTool(sessionId: string, toolId: string): Promise<boolean>;
  createRequest(sessionId: string, toolName: string, params: Record<string, unknown>): Promise<string>;
  startProviderRuntime(sessionId: string): Promise<void>;
  startRequestProcessing(sessionId: string, requestId: string): Promise<void>;
  startStreamingRequest(sessionId: string, toolName: string, params: Record<string, unknown>): Promise<string>;
  getRequestStatus(sessionId: string, requestId: string): Promise<Record<string, unknown>>;
  getRequestChunksWindow(sessionId: string, requestId: string): Promise<Record<string, unknown>>;
  resumeStream(requestId: string, lastSeq: number): Promise<Record<string, unknown>>;
  waitForRequestCompletion(sessionId: string, requestId: string): Promise<Record<string, unknown>>;
  listRequests(sessionId: string, request: Record<string, unknown>): Promise<Record<string, unknown>[]>;
  createApiKey(sessionId: string, name: string, capabilities?: string[]): Promise<Record<string, unknown>>;
  listApiKeys(sessionId: string): Promise<Record<string, unknown>[]>;
  revokeApiKey(sessionId: string, keyId: string): Promise<boolean>;
  registerMachine(sessionId: string, request: Record<string, unknown>): Promise<Record<string, unknown>>;
  listMachines(sessionId: string): Promise<Record<string, unknown>[]>;
  getMachine(sessionId: string, machineId: string): Promise<Record<string, unknown>>;
  unregisterMachine(sessionId: string, machineId: string): Promise<boolean>;
  drainMachine(sessionId: string, machineId: string): Promise<boolean>;
  invoke(sessionId: string, toolName: string, params: Record<string, unknown>): Promise<unknown>;
  stream(sessionId: string, toolName: string, params: Record<string, unknown>): Promise<[unknown[], boolean]>;
}