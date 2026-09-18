# Server Architecture

The server is the contract-owning runtime for Toolplane's remote tool-execution control plane: a canonical gRPC surface generated from `proto/service.proto`, an HTTP gateway compatibility edge, and an MCP JSON-RPC edge. This file is the module map; trace order, conventions, and commands live in the repo root `AGENTS.md` and `server/Makefile` (`make help`).

## Module Graph

```text
proto/service.proto                  canonical contract (single editable source)
  -> internal/server                 control-plane lifecycle, config, auth wiring, storage boot
  -> internal/gateway                HTTP/JSON gateway (grpc-gateway transcode edge)
  -> internal/mcpgateway             MCP JSON-RPC facade edge
  -> internal/cli                    shared CLI plumbing (exit codes, rendering, connections)
  -> internal/auth                   principal extraction and capability checks
cmd/server, cmd/proxy, cmd/mcp-gateway   thin mains over the internal packages
cmd/toolplane                            unified binary: serve/gateway/mcp + invoke/status/doctor/demo verbs
pkg/service                           domain services behind the transport adapter (server.go)
pkg/storage                           Storer interface + Postgres and memory backends
pkg/model, pkg/trace, pkg/observability, pkg/mcp   shared model, tracing, metrics, MCP types
```

## Entry Points

- `cmd/toolplane`: the documented operator surface — `toolplane serve`, `gateway serve`, `mcp serve`, plus invoke/session/key/machine/request/task/tool/audit verbs, `status`, `doctor`, `demo`.
- `cmd/server`, `cmd/proxy`, `cmd/mcp-gateway`: the standalone binaries (used by the compose reference and CI); thin wrappers over `internal/server`, `internal/gateway`, `internal/mcpgateway`.
- `proto/service.proto`: the canonical API contract; all client stubs are regenerated from it.

## Transport To Domain Flow

1. `internal/server` constructs the domain services (`ToolService`, `SessionsService`, `MachinesService`, `RequestsService`, `TasksService`) over the configured storage.
2. `pkg/service/server.go` is the single gRPC adapter implementing every proto service; handlers convert transport payloads and delegate to the owning service file.
3. Domain services operate on `pkg/model` entities and persist through `pkg/storage` (the Storer interface is the durability boundary; per-replica caches hold clones, never shared pointers).
4. Responses convert back through the adapter; errors funnel through the domain-error taxonomy in `pkg/service/errors.go` — never handler-local status codes.

## Notes For Agents

- Do not edit generated protobuf outputs; start contract changes in `service.proto` and regenerate.
- Behavior bugs are usually split between the transport adapter and the owning domain service; inspect both.
- The exit-code contract for `cmd/toolplane` verbs lives in `internal/cli`.
