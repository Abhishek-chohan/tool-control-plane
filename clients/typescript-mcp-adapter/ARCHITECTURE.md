# TypeScript MCP Adapter Architecture

This component is an optional compatibility layer that translates one Toolplane session into an MCP stdio server. It lives at the repo edge on purpose: the native Toolplane control-plane contract remains the protobuf API under `server/proto/service.proto`.

## Module Graph

```text
src/cli.ts
  -> src/server.ts
    -> src/bridge.ts
      -> toolplane-typescript-client
    -> @modelcontextprotocol/sdk/server/*
src/resources.ts
  -> static concept-map and resource URI definitions
tests/integration.test.ts
  -> src/resources.ts
  -> clients/typescript-client/tests/conformance/*
  -> @modelcontextprotocol/sdk/client/*
```

## Entry Points

- `src/index.ts`: exports configuration, bridge, resource constants, and server wrapper types.
- `src/cli.ts`: stdio MCP server entry point.
- `src/server.ts`: MCP request handlers and adapter lifecycle.
- `src/bridge.ts`: native Toolplane client orchestration and semantic translation.

## Translation Model

- Session scope: one adapter process is bound to one native Toolplane session.
- Tool discovery: session-scoped `listTools()` becomes MCP `tools/list`.
- Tool invocation: `executeTool()` becomes MCP `tools/call`, with the adapter waiting for terminal request completion.
- Streaming: native stream chunks are aggregated into the returned MCP tool result and mirrored in translated request resources.
- Inspection: session context, the concept map, and recent request records are exposed as read-only MCP resources.

## Notes For Agents

- Treat this component as a translated compatibility surface, not as proof that MCP semantics are native to the server.
- Keep native lifecycle behavior in the Go server and maintained SDKs. Only translation logic belongs here.
- Validate adapter behavior with live integration tests rather than README examples alone.