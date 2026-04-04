# Go Client Architecture

The Go client is a narrower maintained client projection of Toolplane's remote tool-execution control plane. The protobuf/gRPC contract is the source of truth. This folder keeps a mixed library-and-demo shape, but the maintained public surface is now gRPC-only.

## Module Graph

```text
client/toolplane_client.go
  -> proto/*
  -> google/*
client.go
  -> client/toolplane_client.go
examples/basic/main.go
examples/advanced/main.go
  -> client/toolplane_client.go
```

## Entry Points

- `client/toolplane_client.go`: main reusable client implementation.
- `client.go`: demo executable that defaults to the maintained gRPC path.
- `examples/basic/main.go` and `examples/advanced/main.go`: runnable usage references.

## Public API Surface

`ToolplaneClient` exposes:

- Lifecycle: `Connect()`, `Disconnect()`.
- Connectivity and execution helpers: `Ping()`, `ExecuteTool()`, `StreamExecuteTool()`, `Add()`, `Subtract()`, `Multiply()`, `Divide()`.
- Tool helpers: `RegisterTool()`, `ListTools()`, `GetToolByID()`, `GetToolByName()`, `DeleteTool()`.
- Session and API-key helpers: `CreateSession()`, `GetSession()`, `ListSessions()`, `UpdateSession()`, `CreateAPIKey()`, `ListAPIKeys()`, `RevokeAPIKey()`.
- Machine helpers: `RegisterMachine()`, `ListMachines()`, `GetMachine()`, `UnregisterMachine()`, `DrainMachine()`.
- Request helpers: `CreateRequest()`, `GetRequest()`, `ListRequests()`, `CancelRequest()`.
- Task helpers: `CreateTask()`, `GetTask()`, `ListTasks()`, `CancelTask()`.
- No maintained provider runtime loop that claims requests and submits results.

## Transport Model

- gRPC is the only public transport. `NewToolplaneClient(...)` rejects non-gRPC protocols.
- The client initializes protobuf service stubs directly and uses `ToolService.HealthCheck` during connect.
- Public execution helpers poll `RequestsService` to follow the live request lifecycle instead of returning demo-only transport results.
- Provider-mode ownership remains explicit: Go exposes direct machine and tool wrappers but not the maintained Python-style provider runtime harness.

## Key Files

| File | Responsibility |
| --- | --- |
| `client/toolplane_client.go` | Library implementation for gRPC connection management, request polling, and public lifecycle wrappers |
| `client.go` | Runnable demonstration of the maintained gRPC control-plane path |
| `examples/basic/main.go` | Minimal usage example for the client package |
| `examples/advanced/main.go` | Broader sample flow showing richer client usage |
| `proto/` | Generated protobuf types and gRPC client interfaces |

## Notes For Agents

- Treat the Go client as narrower than Python: it surfaces selected contract operations rather than the full protobuf surface.
- If a feature is missing here, verify whether it is absent from the Go SDK or simply implemented only in Python.
- Do not reintroduce `/rpc` or JSON-RPC compatibility helpers into the maintained public surface.
