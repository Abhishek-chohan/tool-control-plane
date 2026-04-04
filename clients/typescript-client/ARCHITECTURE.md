# TypeScript Client Architecture

The TypeScript client is the maintained JavaScript-family projection of Toolplane's remote tool-execution control plane. Its public story is the grpc-js wrapper layer over the protobuf contract, and the shipped SDK surface is gRPC-only.

## Module Graph

```text
src/index.ts
  -> src/core/toolplane_client.ts
  -> src/interfaces/*
  -> src/errors/*
src/client.ts
  -> src/core/toolplane_client.ts
src/examples/*.ts
  -> src/core/toolplane_client.ts
```

## Entry Points

- `src/index.ts`: public exports for the SDK.
- `src/core/toolplane_client.ts`: main implementation.
- `src/client.ts`: executable entry used by scripts and defaulted to the maintained gRPC control-plane path.
- `src/examples/`: maintained gRPC-first example flows.

## Public API Surface

`ToolplaneClient` exposes:

- Lifecycle: `connect()`, `disconnect()`, `isConnected()`, `getConnectionStatus()`.
- Connectivity and execution helpers: `executeTool()`, `add()`, `subtract()`, `multiply()`, `divide()`, `ping()`.
- Tool helpers: `registerTool()`, `listTools()`, `getToolById()`, `getToolByName()`, `deleteTool()`.
- Session and API-key helpers: `createSession()`, `getSession()`, `listSessions()`, `updateSession()`, `createApiKey()`, `listApiKeys()`, `revokeApiKey()`.
- Machine helpers: `registerMachine()`, `listMachines()`, `getMachine()`, `unregisterMachine()`, `drainMachine()`.
- Request helpers: `createRequest()`, `getRequest()`, `listRequests()`, `cancelRequest()`.
- Task helpers: `createTask()`, `getTask()`, `listTasks()`, `cancelTask()`.
- Factory: `createGRPCClient()`.
- No maintained provider runtime loop that claims requests and submits results.

## Transport Model

- The public SDK transport is gRPC only and initializes `ToolServiceClient`, `SessionsServiceClient`, `MachinesServiceClient`, `RequestsServiceClient`, and `TasksServiceClient` from the generated protobuf surface.
- The shared-fixture conformance runner under `tests/conformance/` still includes an internal HTTP adapter so fixture behavior can be checked against the maintained HTTP gateway. That adapter is not part of the public SDK surface.
- Public gRPC helpers in `src/core/toolplane_client.ts` now follow the live request lifecycle: auth metadata, health-check connect, request polling, and machine-aware tool registration.
- Broader protobuf coverage still remains narrower than Python; verify missing public wrappers against `server/proto/service.proto` before assuming parity.
- Provider-mode ownership remains explicit: TypeScript exposes direct machine and tool wrappers but not the maintained Python-style provider runtime harness.

## Key Files

| File | Responsibility |
| --- | --- |
| `src/core/toolplane_client.ts` | Main client implementation, connection logic, request polling, and shipped live gRPC wrappers |
| `src/interfaces/` | Shared transport and domain types |
| `src/errors/` | Error classes for connection, timeout, protocol, and validation failures |
| `src/examples/` | Example usage flows for basic and advanced scenarios |
| `tests/conformance/` | Repository-internal fixture runners for gRPC and HTTP transport coverage |
| `src/proto/` | Generated or protobuf-related TypeScript assets |

## Notes For Agents

- Use this SDK as a good TypeScript navigation target, and verify any broader gRPC coverage work against the Python or Go clients.
- When changing the public API, update `src/index.ts` first so the export surface remains discoverable.
- Do not confuse the internal HTTP conformance adapter with a public HTTP SDK surface.
