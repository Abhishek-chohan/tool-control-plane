export interface AdapterOptions {
  grpcHost: string;
  grpcPort: number;
  userId: string;
  apiKey?: string;
  sessionId?: string;
  sessionName: string;
  sessionDescription: string;
  sessionNamespace: string;
  timeoutMs: number;
  requestResourceLimit: number;
}

const DEFAULT_TIMEOUT_MS = 30_000;
const DEFAULT_REQUEST_RESOURCE_LIMIT = 20;

function trimmed(value: string | undefined): string {
  return value?.trim() ?? '';
}

function parsePositiveInteger(name: string, value: string | undefined, fallback: number): number {
  if (!value || !value.trim()) {
    return fallback;
  }

  const parsed = Number.parseInt(value, 10);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    throw new Error(`${name} must be a positive integer, received ${value}`);
  }

  return parsed;
}

export function createAdapterOptionsFromEnv(env: NodeJS.ProcessEnv = process.env): AdapterOptions {
  const sessionId = trimmed(env.TOOLPLANE_MCP_SESSION_ID);
  const apiKey =
    trimmed(env.TOOLPLANE_MCP_API_KEY)
    || trimmed(env.TOOLPLANE_CONFORMANCE_API_KEY)
    || trimmed(env.TOOLPLANE_AUTH_FIXED_API_KEY);

  return {
    grpcHost: trimmed(env.TOOLPLANE_MCP_GRPC_HOST) || trimmed(env.TOOLPLANE_CONFORMANCE_GRPC_HOST) || 'localhost',
    grpcPort: parsePositiveInteger(
      'TOOLPLANE_MCP_GRPC_PORT',
      env.TOOLPLANE_MCP_GRPC_PORT ?? env.TOOLPLANE_CONFORMANCE_GRPC_PORT,
      9001,
    ),
    userId: trimmed(env.TOOLPLANE_MCP_USER_ID) || trimmed(env.TOOLPLANE_CONFORMANCE_USER_ID) || 'mcp-adapter-user',
    apiKey: apiKey || undefined,
    sessionId: sessionId || undefined,
    sessionName: trimmed(env.TOOLPLANE_MCP_SESSION_NAME) || 'mcp-stdio-adapter',
    sessionDescription:
      trimmed(env.TOOLPLANE_MCP_SESSION_DESCRIPTION)
      || 'Session created by the optional Toolplane MCP stdio adapter',
    sessionNamespace: trimmed(env.TOOLPLANE_MCP_SESSION_NAMESPACE) || 'mcp',
    timeoutMs: parsePositiveInteger('TOOLPLANE_MCP_TIMEOUT_MS', env.TOOLPLANE_MCP_TIMEOUT_MS, DEFAULT_TIMEOUT_MS),
    requestResourceLimit: parsePositiveInteger(
      'TOOLPLANE_MCP_REQUEST_RESOURCE_LIMIT',
      env.TOOLPLANE_MCP_REQUEST_RESOURCE_LIMIT,
      DEFAULT_REQUEST_RESOURCE_LIMIT,
    ),
  };
}