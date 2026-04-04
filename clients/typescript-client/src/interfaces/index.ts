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
  sessionId: string;
  createdAt: string;
  createdBy: string;
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
  apiKey: string;
  sessionId: string;
  namespace: string;
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