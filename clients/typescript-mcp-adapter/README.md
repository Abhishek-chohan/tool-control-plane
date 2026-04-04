# Toolplane TypeScript MCP Adapter

This package is an optional in-repo adapter that exposes one Toolplane session through an MCP stdio server. It is built on the maintained TypeScript client path, and the source of truth remains `server/proto/service.proto` and the Go server runtime.

## Support Status

- Optional adapter at the repo edge.
- Built on the maintained TypeScript gRPC client path.
- MCP transport: stdio.
- Session ownership, machine lifecycle, request claiming, drain, and retained-window recovery stay aligned with the underlying Toolplane session.

## Adapter Behavior

- One adapter process binds to one Toolplane session.
- `tools/list` returns session-scoped tool discovery.
- `tools/call` creates a request and waits for completion.
- Stream chunks are aggregated into the returned tool result.
- Read-only resources expose the bound session and recent request records.

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

The integration suite boots a real Toolplane server, registers live provider-backed tools, spawns the stdio MCP adapter, and validates tool discovery, unary invocation, streaming aggregation, and request-resource inspection through the adapter.
