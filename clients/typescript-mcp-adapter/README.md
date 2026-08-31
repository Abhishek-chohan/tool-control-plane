# Toolplane TypeScript MCP Adapter

This package is an optional in-repo adapter that exposes one Toolplane session through an MCP stdio server. It is built on the maintained TypeScript client path, and the source of truth remains `server/proto/service.proto` and the Go server runtime.

## Support Status

- Optional adapter at the repo edge.
- Built on the maintained TypeScript gRPC client path.
- MCP transport: stdio.
- Dual protocol surface: MCP 2026-07-28 stateless core with the Tasks extension for modern clients, and the legacy initialize-based protocol for pre-2026 clients.
- Session ownership, machine lifecycle, request claiming, drain, and retained-window recovery stay aligned with the underlying Toolplane session.

## Role In The Seam

This adapter is the repo's reference Layer 4 edge adapter in the maintained agent-runtime seam.

- It binds one explicit Toolplane session to one MCP stdio server process.
- It reuses the native Toolplane control plane for request lifecycle, provider ownership, retained-window recovery, and drain behavior.
- It translates MCP discovery, invocation, and read-only inspection onto existing Toolplane session and request concepts instead of inventing a parallel runtime.
- It is an edge example, not proof of general SDK parity. The source of truth remains `server/proto/service.proto`, the Go server runtime, and `SDK_MAP.md`.

See `server/docs/agent-runtime-integration-seam.md` for the full four-layer seam model and the minimal adapter contract.

## Incremental Adoption Role

Use this adapter as coexistence proof after one Toolplane-backed session already exists for the remote tool you are offloading.

- An existing runtime can keep its own tool-selection logic and expose selected Toolplane-backed tools outward through MCP.
- Direct local tools can remain outside Toolplane entirely.
- Session ownership, request lifecycle, retained replay, and drain remain native Toolplane behavior underneath the adapter.

This adapter is supporting evidence for the incremental-adoption story, not a requirement for the first migration and not a replacement for the control plane itself. See `../../server/docs/incremental-adoption.md` for the maintained first-tool migration guide.

## Adapter Behavior

- One adapter process binds to one Toolplane session.
- `tools/list` returns session-scoped tool discovery.
- `tools/call` creates a durable request that a provider machine claims and executes.
- Requests carrying a 2026 `_meta.io.modelcontextprotocol/protocolVersion` are served with `resultType`-bearing results; requests without it are served with legacy semantics negotiated at `initialize`, so pre-2026 clients keep working against the same process.
- 2026 clients that advertise the `io.modelcontextprotocol/tasks` extension receive a task handle from `tools/call` and poll it with `tasks/get`; other clients are served synchronously.
- Stream chunks surface through the retained chunk window on `tasks/get` (while working) and are aggregated into the result for synchronous callers.
- Read-only resources expose the bound session and recent request records.

## Protocol Mapping

| MCP surface | Native Toolplane mapping | Current note |
| --- | --- | --- |
| `server/discover` | Static advertisement | Announces `2026-07-28`, the tools capability, and the Tasks extension |
| `tools/list` | Session-scoped tool discovery | No global catalog is introduced by the adapter |
| `tools/call` (Tasks-capable 2026 client) | Durable request creation | Returns a task handle immediately; the task id is the Toolplane request id |
| `tools/call` (other clients) | Request-backed execution plus terminal-state waiting | The returned payload keeps the native request visible through a request resource link |
| `tasks/get` | Request status plus retained chunk window | Statuses map pending/claimed/running/stalled → `working`, done → `completed`, failure → `failed` or `cancelled`; chunks replay honors the `dev.toolplane/last_seq` cursor |
| `tasks/cancel` | Cooperative request cancellation | Terminal tasks are acknowledged, not rejected |
| `tasks/update` | Not supported | Toolplane tasks never request client input (`-32601`) |
| Streaming tool output | Native chunk stream | Chunks are aggregated for synchronous callers and surfaced through the chunk window for task polling |
| `resources/list` and `resources/read` | Session and request inspection | Resources remain read-only and point back to native Toolplane state |

Every 2026 request must carry `_meta` with `io.modelcontextprotocol/protocolVersion` (`2026-07-28`) and `io.modelcontextprotocol/clientCapabilities`; an unsupported version is rejected with `-32022` and its supported list. Request IDs remain visible through request resources so inspection can stay on the native control plane.

## Installation

```bash
cd clients/typescript-client && npm install
cd clients/typescript-client && npm run build
cd clients/typescript-mcp-adapter && npm install
cd clients/typescript-mcp-adapter && npm run build
```

## Environment

The adapter reads these variables:

- `TOOLPLANE_MCP_GRPC_HOST`
- `TOOLPLANE_MCP_GRPC_PORT`
- `TOOLPLANE_MCP_USER_ID`
- `TOOLPLANE_MCP_API_KEY`
- `TOOLPLANE_MCP_SESSION_ID`
- `TOOLPLANE_MCP_SESSION_NAME`
- `TOOLPLANE_MCP_SESSION_DESCRIPTION`
- `TOOLPLANE_MCP_SESSION_NAMESPACE`
- `TOOLPLANE_MCP_TIMEOUT_MS`
- `TOOLPLANE_MCP_REQUEST_RESOURCE_LIMIT`

If `TOOLPLANE_MCP_SESSION_ID` is omitted, the adapter creates a new Toolplane session at startup. For tests and local fixture bootstraps it also falls back to the `TOOLPLANE_CONFORMANCE_*` variables.

## Running The Adapter

```bash
cd clients/typescript-mcp-adapter
TOOLPLANE_MCP_GRPC_HOST=localhost \
TOOLPLANE_MCP_GRPC_PORT=9001 \
TOOLPLANE_MCP_USER_ID=adapter-user \
TOOLPLANE_MCP_API_KEY=toolplane-conformance-fixture-key \
TOOLPLANE_MCP_SESSION_ID=<existing-session-id> \
npm start
```

This launches a stdio MCP server process. Any MCP client that can spawn a local command can use `node dist/cli.js` as the server command.

## Validation

```bash
cd clients/typescript-mcp-adapter && npm test
```

The integration suite boots a real Toolplane server, registers live provider-backed tools, starts the native provider runtime, spawns the stdio MCP adapter, and validates tool discovery, unary invocation, streaming aggregation, and request-resource inspection through the adapter. A second suite speaks the 2026-07-28 stateless protocol with a raw JSON-RPC stdio client (no initialize handshake) and validates `server/discover`, task handles, `tasks/get` chunk-window replay with the `last_seq` cursor, `tasks/cancel`, the synchronous fallback for non-Tasks clients, and protocol-version rejection. The goal is to prove the adapter stays a thin translation layer over native Toolplane semantics rather than redefining them.
