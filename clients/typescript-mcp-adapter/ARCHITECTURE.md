# TypeScript MCP Adapter Architecture

This component is an optional compatibility layer that translates one Toolplane session into an MCP stdio server. It lives at the repo edge on purpose: the native Toolplane control-plane contract remains the protobuf API under `server/proto/service.proto`.

## Module Graph

```text
src/cli.ts
  -> src/server.ts
    -> src/bridge.ts
      -> toolplane-typescript-client
    -> src/protocol.ts
    -> @modelcontextprotocol/sdk/server/*
src/protocol.ts
  -> MCP 2026-07-28 wire contract: meta validation, task payloads, error codes
src/resources.ts
  -> static concept-map and resource URI definitions
tests/integration.test.ts
  -> src/resources.ts
  -> src/protocol.ts
  -> clients/typescript-client/tests/conformance/*
  -> @modelcontextprotocol/sdk/client/*
```

## Entry Points

- `src/index.ts`: exports configuration, bridge, protocol, resource constants, and server wrapper types.
- `src/cli.ts`: stdio MCP server entry point.
- `src/server.ts`: MCP request dispatcher and adapter lifecycle.
- `src/bridge.ts`: native Toolplane client orchestration and semantic translation.
- `src/protocol.ts`: MCP 2026-07-28 stateless wire contract shared by the dispatcher and the bridge.

## Translation Model

- Session scope: one adapter process is bound to one native Toolplane session.
- Protocol surface: requests carrying `_meta.io.modelcontextprotocol/protocolVersion: 2026-07-28` are served with the stateless surface (`resultType` discriminators, task handles, `tasks/*`); requests without it are served with legacy `initialize`-negotiated semantics. Both modes run in the same process.
- Tool discovery: session-scoped `listTools()` becomes MCP `tools/list`; `server/discover` advertises the 2026 surface to stateless clients.
- Tool invocation: `tools/call` always enqueues a durable request for a provider machine. Tasks-capable 2026 clients receive a task handle immediately (`createRequest()`); all other callers are served synchronously after terminal completion (`executeTool()` for legacy clients, request polling for 2026 sync clients).
- Tasks: a task id is a Toolplane request id. `tasks/get` maps request status to task status and attaches the retained chunk window (honoring the `dev.toolplane/last_seq` cursor) while the task is working; `tasks/cancel` performs cooperative request cancellation.
- Streaming: native stream chunks surface through the chunk window for task polling and are aggregated into the returned MCP tool result for synchronous callers.
- Inspection: session context, the concept map, and recent request records are exposed as read-only MCP resources.

## Notes For Agents

- Treat this component as a translated compatibility surface, not as proof that MCP semantics are native to the server.
- Keep native lifecycle behavior in the Go server and maintained SDKs. Only translation logic belongs here.
- Validate adapter behavior with live integration tests rather than README examples alone.