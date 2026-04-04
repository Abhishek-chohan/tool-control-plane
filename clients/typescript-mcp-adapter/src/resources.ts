export const ADAPTER_NAME = 'toolplane-typescript-mcp-adapter';
export const ADAPTER_VERSION = '1.0.0';

export const CONCEPT_MAP_RESOURCE_URI = 'toolplane://adapter/concept-map';
export const SESSION_RESOURCE_URI = 'toolplane://session/current';

export function requestResourceUri(requestId: string): string {
  return `toolplane://requests/${encodeURIComponent(requestId)}`;
}

export function parseRequestResourceId(uri: string): string | null {
  const prefix = 'toolplane://requests/';
  if (!uri.startsWith(prefix)) {
    return null;
  }

  return decodeURIComponent(uri.slice(prefix.length));
}

export const ADAPTER_INSTRUCTIONS = [
  'This server translates one Toolplane session into MCP tools and read-only resources.',
  'The native Toolplane contract remains authoritative.',
  'Each MCP tool call waits for the underlying Toolplane request lifecycle to reach a terminal state.',
  'When a native Toolplane tool emits stream chunks, this adapter aggregates those chunks into the returned MCP tool result.',
  'Machine registration, request claiming, retained-window recovery, and drain behavior remain native Toolplane concerns rather than MCP-native behavior.',
].join(' ');

export const CONCEPT_MAP_MARKDOWN = `# Toolplane to MCP concept map

- Session scope: one adapter process is bound to one Toolplane session.
- Tool discovery: MCP tools/list maps to native Toolplane tool discovery inside that session.
- Tool invocation: MCP tools/call maps to native Toolplane request creation plus terminal-state waiting.
- Streaming: native Toolplane stream chunks are aggregated into one synchronous MCP tool result because stdio MCP tool calls do not expose the same request lifecycle.
- Request inspection: the adapter exposes read-only MCP resources for the current session, this concept map, and recent translated request records.
- Runtime ownership: provider machines, request claiming, heartbeats, drain, and retained-window recovery remain native Toolplane responsibilities.
- Source of truth: server/proto/service.proto and the Go server runtime still define native semantics. This adapter is a translated compatibility layer, not a replacement protocol.`;