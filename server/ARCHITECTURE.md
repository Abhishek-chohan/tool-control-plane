# Server Architecture

The server is the contract-owning runtime for Toolplane's remote tool-execution control plane. The maintained product boundary is a protobuf/gRPC contract, an HTTP gateway compatibility layer, and optional ecosystem adapters that sit outside the core runtime. The deprecated `/rpc` endpoint remains a documented server-side migration surface on the path to `v2.0.0`, but no maintained SDK in this repo depends on it.

## Boundary

- This server is the core control plane for sessions, machines, requests, tasks, and tool ownership.
- The gRPC contract is the primary architecture surface.
- The HTTP gateway is a maintained compatibility layer over the same contract.
- `/rpc` is outside the maintained SDK surfaces and remains only as a documented server-side retirement path.

## Module Graph

```text
proto/service.proto
  -> cmd/server/main.go
    -> pkg/service/server.go
      -> pkg/service/tool.go
      -> pkg/service/session.go
      -> pkg/service/machine.go
      -> pkg/service/requests.go
      -> pkg/service/task.go
    -> pkg/storage/*
    -> pkg/trace/*
  -> cmd/proxy/main.go
    -> proto/service.pb.gw.go
```

## Entry Points

- `cmd/server/main.go`: boots the gRPC server, auth interceptors, storage, tracing, and the aggregated service adapter.
- `cmd/proxy/main.go`: runs the HTTP gateway generated from protobuf HTTP annotations.
- `proto/service.proto`: canonical API contract for all services and generated stubs.

## Transport To Domain Flow

1. `cmd/server/main.go` constructs `ToolService`, `SessionsService`, `MachinesService`, `RequestsService`, and `TasksService`.
2. `pkg/service/server.go` exposes one `GRPCServer` adapter that implements every protobuf service.
3. Each handler converts transport payloads into domain-friendly arguments and delegates to the owning service file.
4. Domain services operate on `pkg/model` entities and optionally persist through `pkg/storage`.
5. Responses are converted back into protobuf messages and returned through gRPC or the gateway.

## Execution Flow

For a typical tool execution:

1. `ToolService.RegisterTool` or machine registration makes a tool discoverable for a session.
2. `RequestsService.CreateRequest` creates queue state for a tool invocation.
3. `RequestsService.ClaimRequest` binds the request to a machine.
4. `RequestsService.ExecuteTool` waits on request updates until completion, failure, stall, or timeout.
5. `TasksService` builds on top of `RequestsService` for higher-level scheduled execution.

## Handler Ownership

| Proto service | Adapter file | Domain file |
| --- | --- | --- |
| `ToolService` | `pkg/service/server.go` | `pkg/service/tool.go` |
| `SessionsService` | `pkg/service/server.go` | `pkg/service/session.go` |
| `MachinesService` | `pkg/service/server.go` | `pkg/service/machine.go` |
| `RequestsService` | `pkg/service/server.go` | `pkg/service/requests.go` |
| `TasksService` | `pkg/service/server.go` | `pkg/service/task.go` |

## Key Files

| File | Responsibility |
| --- | --- |
| `pkg/service/server.go` | Transport adapter that implements all protobuf services and translates errors into gRPC status codes |
| `pkg/service/tool.go` | Tool lifecycle, lookup, ping updates, and registration bookkeeping |
| `pkg/service/session.go` | Session CRUD, user session indexing, API key handling, and namespace metadata |
| `pkg/service/machine.go` | Machine registration, heartbeat, draining, and machine-to-tool association |
| `pkg/service/requests.go` | Request queueing, claiming, result submission, chunk handling, and timeout/stall detection |
| `pkg/service/task.go` | Higher-level task orchestration built on request execution |
| `pkg/service/persistence.go` | Service persistence glue shared by domain services |
| `pkg/storage/` | Optional persistent storage backend used when environment configuration is present |
| `pkg/model/` | Shared in-memory and persisted entity definitions |
| `pkg/trace/` | Session tracing hooks used by tool and machine flows |

## Notes For Agents

- Do not edit generated protobuf outputs directly unless the task is specifically about generated files.
- Contract changes should usually start in `proto/service.proto`, not in `pkg/service/server.go`.
- Behavior bugs are often split between the transport adapter and the owning domain service, so inspect both before patching.
