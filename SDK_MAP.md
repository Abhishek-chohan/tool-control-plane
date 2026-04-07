# SDK Map

This file is the fastest contract-to-wrapper map from `server/proto/service.proto` to the current Python, Go, and TypeScript SDK surfaces. Use it to trace maintained wrappers, support labels, and validation scope.

## Support Labels

- `full`: explicit public wrapper exists and the behavior family is part of the maintained validation story.
- `partial`: only an indirect, internal, provider-mode, transport-limited, or intentionally narrower path exists.
- `mock`: wrapper exists, but the relevant SDK path is mock-backed rather than a live protobuf transport.
- `unsupported`: no current wrapper targets that RPC.

## Scope Categories

RPCs fall into three scope categories that explain why some wrappers are Python-only or internal:

- **consumer**: remote invocation, session or machine or task lifecycle, and tool discovery. These RPCs are portable across maintained SDKs.
- **provider**: tool registration, request claiming, heartbeat, and result submission. These RPCs are exercised through the explicit maintained provider runtimes in Python and TypeScript. Go exposes selected direct registration wrappers but does not ship a maintained provider runtime harness.
- **admin**: session administration helpers such as bulk delete, stats, token refresh, and invalidation. These RPCs are currently exposed only in the Python SDK.

## Important Caveats

- `server/proto/service.proto` is the canonical contract.
- `server/docs/compatibility-policy.md` defines the written compatibility and deprecation rules for the protobuf contract, HTTP gateway, and maintained SDK wrappers.
- `clients/python-client/` is the richest SDK and the baseline for intended end-to-end capability.
- `clients/go-client/` and `clients/typescript-client/` expose selected live gRPC wrappers, but both remain narrower public surfaces than Python.
- `clients/typescript-client/` is the maintained JavaScript-family parity surface. Its HTTP adapter under `tests/conformance/` exists only to exercise the maintained HTTP gateway against shared fixtures; it is not a public HTTP SDK surface.
- `/rpc` remains a server-side removal-path surface documented in `server/docs/rpc-retirement.md`; it stays outside maintained SDK support and parity.
- `clients/typescript-mcp-adapter/` is an optional stdio adapter for one Toolplane session. Keep it outside the SDK parity tables.

## Provider Mode Support Decision

- Python: maintained provider mode exists through the explicit `ProviderRuntime` surface and is the canonical provider harness for claim, heartbeat, result submission, and drain behavior.
- Go: direct `RegisterMachine`, `RegisterTool`, and `DrainMachine` wrappers exist, but there is no maintained provider runtime loop that claims requests and submits results. Treat provider mode as unsupported as a named runtime surface.
- TypeScript: maintained provider mode exists through the explicit `ProviderRuntime` surface plus direct gRPC lifecycle wrappers for claim, heartbeat, chunk append, and result submission.

## Support Tiers

- Primary maintained surface: the Go server contract in `server/proto/service.proto`, the server runtime under `server/pkg/service/`, the shared conformance fixtures, and the Python SDK.
- Supported secondary SDKs: Go and TypeScript on their documented maintained public surfaces.
- Compatibility surfaces: the HTTP gateway. The server-side `/rpc` path remains documented separately during its removal window to `v2.0.0`.

### `/rpc` Reference Note

The HTTP JSON-RPC `/rpc` endpoint remains a server-side reference surface during the documented removal path to `v2.0.0`. It sits outside the maintained parity story, stays out of required CI, and should not be used as the basis for new SDK development. Use `server/docs/rpc-retirement.md` for the removal path and `.plans/roadmap-latest/tier-0-rpc-inventory.md` for the historical inventory.

## SDK Reality Snapshot

| SDK | Primary entry | Real transports | Contract notes |
| --- | --- | --- | --- |
| Python | `clients/python-client/toolplane/__init__.py` | gRPC + HTTP | Primary maintained SDK and current completeness baseline; the explicit `ProviderRuntime` is the maintained provider harness |
| Go | `clients/go-client/client/toolplane_client.go` | live gRPC | Supported secondary SDK with maintained gRPC lifecycle helpers and no provider runtime harness |
| TypeScript | `clients/typescript-client/src/index.ts` | live gRPC | Supported secondary SDK and maintained JavaScript-family path; includes an explicit gRPC `ProviderRuntime`, while repository-internal HTTP adapters under `tests/conformance/` are not part of the public SDK surface |

## Ecosystem Adapter Snapshot

| Surface | Path | Support label | Notes |
| --- | --- | --- | --- |
| TypeScript MCP adapter | `clients/typescript-mcp-adapter/` | `full` | Optional stdio adapter that exposes MCP `tools/list`, `tools/call`, `resources/list`, and `resources/read` for one Toolplane session. Keep it outside the SDK parity tables. |

## ToolService

| RPC | Python | Go | TypeScript | Notes / conformance |
| --- | --- | --- | --- | --- |
| `RegisterTool` | `partial`: provider registration via explicit `ProviderRuntime` | `full`: `RegisterTool()` | `full`: `registerTool()` plus `ProviderRuntime.registerTool()` / `ProviderRuntime.tool()` | Direct TypeScript registration still requires a machine; the explicit runtime now owns the maintained provider path |
| `ListTools` | `full`: `get_available_tools()` / `list_tools()` | `full`: `ListTools()` | `full`: `listTools()` | Covered by `conformance/cases/tool_discovery.json` |
| `GetToolById` | `full`: `get_tool_by_id()` | `full`: `GetToolByID()` | `full`: `getToolById()` | Covered by `conformance/cases/tool_discovery.json` |
| `GetToolByName` | `full`: `get_tool_by_name()` | `full`: `GetToolByName()` | `full`: `getToolByName()` | Covered by `conformance/cases/tool_discovery.json` |
| `DeleteTool` | `full`: `delete_tool()` | `full`: `DeleteTool()` | `full`: `deleteTool()` | Covered by `conformance/cases/tool_discovery.json` |
| `UpdateToolPing` | `partial`: explicit provider heartbeat path | `unsupported` | `unsupported` | No standalone public ping wrapper (provider scope) |
| `StreamExecuteTool` | `full`: `stream()` / `astream()` | `full`: `StreamExecuteTool()` | `unsupported` | Covered by `conformance/cases/invoke_stream.json` |
| `ResumeStream` | `unsupported` | `unsupported` | `unsupported` | No client currently exposes stream resumption; the server returns `OUT_OF_RANGE` when replay falls behind the retained window |
| `ExecuteTool` | `full`: `invoke()` / `ainvoke()` | `full`: `ExecuteTool()` plus math helpers | `full`: `executeTool()` plus math helpers | Covered by `conformance/cases/invoke_unary.json`; live execution still requires a provider loop |
| `HealthCheck` | `partial`: `ToolplaneHTTP.health()` plus connect probes | `full`: gRPC `Ping()` / `Connect()` | `full`: gRPC `ping()` / `connect()` | TypeScript and Go treat health checks as part of the maintained gRPC connection path |

## SessionsService

| RPC | Python | Go | TypeScript | Notes / conformance |
| --- | --- | --- | --- | --- |
| `CreateSession` | `full`: `create_session()` | `full`: `CreateSession()` | `full`: `createSession()` | Covered by `conformance/cases/session_create.json` |
| `GetSession` | `full`: `get_session()` | `full`: `GetSession()` | `full`: `getSession()` | Public in Python, Go, and TypeScript |
| `ListSessions` | `full`: `list_sessions()` | `full`: `ListSessions()` | `full`: `listSessions()` | Covered by `conformance/cases/session_list.json` |
| `UpdateSession` | `full`: `update_session()` | `full`: `UpdateSession()` | `full`: `updateSession()` | Covered by `conformance/cases/session_update.json` |
| `DeleteSession` | `partial`: internal `_delete_session_on_server()` | `unsupported` | `unsupported` | Public bulk invalidation exists instead of direct delete |
| `ListUserSessions` | `full`: `list_user_sessions()` | `unsupported` | `unsupported` | Python-only session admin helper (admin scope) |
| `BulkDeleteSessions` | `full`: `bulk_delete_sessions()` | `unsupported` | `unsupported` | Python-only session admin helper (admin scope) |
| `GetSessionStats` | `full`: `get_session_stats()` | `unsupported` | `unsupported` | Python-only session admin helper (admin scope) |
| `RefreshSessionToken` | `full`: `refresh_session_token()` | `unsupported` | `unsupported` | Python-only session admin helper (admin scope) |
| `InvalidateSession` | `full`: `invalidate_session()` | `unsupported` | `unsupported` | Python-only session admin helper (admin scope) |
| `CreateApiKey` | `full`: `create_api_key()` | `full`: `CreateAPIKey()` | `full`: `createApiKey()` | Covered by `conformance/cases/api_key_lifecycle.json` |
| `ListApiKeys` | `full`: `list_api_keys()` | `full`: `ListAPIKeys()` | `full`: `listApiKeys()` | Covered by `conformance/cases/api_key_lifecycle.json` |
| `RevokeApiKey` | `full`: `revoke_api_key()` | `full`: `RevokeAPIKey()` | `full`: `revokeApiKey()` | Covered by `conformance/cases/api_key_lifecycle.json` |

## MachinesService

| RPC | Python | Go | TypeScript | Notes / conformance |
| --- | --- | --- | --- | --- |
| `RegisterMachine` | `partial`: provider lifecycle via explicit `ProviderRuntime` | `full`: `RegisterMachine()` | `full`: `registerMachine()` plus `ProviderRuntime.createSession()` / `ProviderRuntime.attachSession()` | Python and TypeScript both own provider registration through explicit runtime surfaces; direct TypeScript machine wrappers remain public |
| `ListMachines` | `full`: `list_machines()` | `full`: `ListMachines()` | `full`: `listMachines()` | Covered by `conformance/cases/machine_lifecycle.json` |
| `GetMachine` | `full`: `get_machine()` | `full`: `GetMachine()` | `full`: `getMachine()` | Covered by `conformance/cases/machine_lifecycle.json` |
| `UpdateMachinePing` | `partial`: explicit provider heartbeat thread | `unsupported` | `full`: `updateMachinePing()` plus `ProviderRuntime` heartbeat loop | TypeScript now exposes the provider heartbeat RPC directly and uses it in the maintained runtime |
| `UnregisterMachine` | `full`: `unregister_machine()` | `full`: `UnregisterMachine()` | `full`: `unregisterMachine()` | Explicit machine cleanup is public in Python, Go, and TypeScript |
| `DrainMachine` | `full`: `drain_machine()` | `full`: `DrainMachine()` | `full`: `drainMachine()` | Covered by the shared machine lifecycle fixtures, including drain-under-load |

## RequestsService

| RPC | Python | Go | TypeScript | Notes / conformance |
| --- | --- | --- | --- | --- |
| `CreateRequest` | `full`: `create_request()` | `full`: `CreateRequest()` | `full`: `createRequest()` | Covered by `conformance/cases/request_create.json` |
| `GetRequest` | `full`: `get_request_status()` | `full`: `GetRequest()` | `full`: `getRequest()` | Python keeps the `get_request_status()` name, while Go and TypeScript expose direct request lookup wrappers |
| `ListRequests` | `full`: `list_requests()` | `full`: `ListRequests()` | `full`: `listRequests()` | Public across Python, Go, and TypeScript |
| `UpdateRequest` | `partial`: internal result-status updates | `unsupported` | `full`: `updateRequest()` | TypeScript exposes the provider-running/status transition helper used by the maintained runtime |
| `ClaimRequest` | `partial`: explicit provider poll loop | `unsupported` | `full`: `claimRequest()` plus `ProviderRuntime` | Python and TypeScript provider runtimes claim queued work; Go still lacks a maintained runtime loop |
| `CancelRequest` | `full`: `cancel_request()` | `full`: `CancelRequest()` | `full`: `cancelRequest()` | Public across Python, Go, and TypeScript |
| `SubmitRequestResult` | `partial`: explicit provider result submission | `unsupported` | `full`: `submitRequestResult()` plus `ProviderRuntime` | TypeScript now exposes the final provider result RPC directly and uses it in the maintained runtime |
| `AppendRequestChunks` | `partial`: explicit streaming result submission | `unsupported` | `full`: `appendRequestChunks()` plus `ProviderRuntime` | TypeScript now exposes streaming chunk submission directly and uses it in the maintained runtime |
| `GetRequestChunks` | `unsupported` | `unsupported` | `unsupported` | No current wrapper; the server returns retained chunks plus `start_seq` / `next_seq` metadata for the bounded replay window |

## TasksService

| RPC | Python | Go | TypeScript | Notes / conformance |
| --- | --- | --- | --- | --- |
| `CreateTask` | `full`: `create_task()` | `full`: `CreateTask()` | `full`: `createTask()` | Public across the maintained SDKs |
| `GetTask` | `full`: `get_task()` | `full`: `GetTask()` | `full`: `getTask()` | Task lookup is public across the maintained SDK surfaces |
| `ListTasks` | `full`: `list_tasks()` | `full`: `ListTasks()` | `full`: `listTasks()` | Python covers both transports; Go and TypeScript cover the maintained gRPC path |
| `CancelTask` | `full`: `cancel_task()` | `full`: `CancelTask()` | `full`: `cancelTask()` | Cancellation follows the underlying request lifecycle and prevents late completion from reviving cancelled work |

## Conformance Coverage Summary

| Case | RPC families exercised | Closest public wrappers |
| --- | --- | --- |
| `session_create` | `SessionsService.CreateSession` | Python `create_session()`, Go `CreateSession()`, TypeScript `createSession()` |
| `session_list` | `SessionsService.ListSessions` | Python `list_sessions()`, Go `ListSessions()`, TypeScript `listSessions()` |
| `session_update` | `SessionsService.UpdateSession` | Python `update_session()`, Go `UpdateSession()`, TypeScript `updateSession()` |
| `request_create` | `RequestsService.CreateRequest`, `RequestsService.GetRequest`, `RequestsService.ListRequests` | Python `create_request()`, `get_request_status()`, `list_requests()`, Go `CreateRequest()` / `GetRequest()` / `ListRequests()`, TypeScript `createRequest()` / `getRequest()` / `listRequests()` |
| `tool_discovery` | `ToolService.ListTools`, `ToolService.GetToolById`, `ToolService.GetToolByName`, `ToolService.DeleteTool` | Python `list_tools()` / `get_tool_by_id()` / `get_tool_by_name()` / `delete_tool()`, Go `ListTools()` / `GetToolByID()` / `GetToolByName()` / `DeleteTool()`, TypeScript `listTools()` / `getToolById()` / `getToolByName()` / `deleteTool()` |
| `api_key_lifecycle` | `SessionsService.CreateApiKey`, `SessionsService.ListApiKeys`, `SessionsService.RevokeApiKey` | Python `create_api_key()`, `list_api_keys()`, `revoke_api_key()`, Go `CreateAPIKey()` / `ListAPIKeys()` / `RevokeAPIKey()`, TypeScript `createApiKey()` / `listApiKeys()` / `revokeApiKey()` |
| `machine_lifecycle` | `MachinesService.RegisterMachine`, `MachinesService.ListMachines`, `MachinesService.GetMachine`, `MachinesService.DrainMachine`, plus in-flight request completion during drain | Python `list_machines()`, `get_machine()`, `drain_machine()`, and `invoke()`, plus the matching Go and TypeScript machine wrappers |
| `invoke_unary` | `ToolService.RegisterTool`, `ToolService.ExecuteTool` | Python `tool()` + `invoke()`, Go `ExecuteTool()`, TypeScript `executeTool()` |
| `invoke_stream` | `ToolService.RegisterTool`, `ToolService.StreamExecuteTool` | Python `tool()` + `stream()` / `astream()`, Go `StreamExecuteTool()` |
| `provider_runtime_unary_claim_submit` | `MachinesService.RegisterMachine`, `RequestsService.CreateRequest`, `RequestsService.ClaimRequest`, `RequestsService.SubmitRequestResult` | Python `ProviderRuntime` plus `create_request()`, TypeScript `ProviderRuntime` plus `createRequest()` |
| `provider_runtime_stream_append_chunks` | `MachinesService.RegisterMachine`, `RequestsService.CreateRequest`, `RequestsService.ClaimRequest`, `RequestsService.AppendRequestChunks`, `RequestsService.SubmitRequestResult` | Python `ProviderRuntime` plus `create_request()`, TypeScript `ProviderRuntime` plus `createRequest()` |
| `provider_runtime_drain_under_load` | `MachinesService.RegisterMachine`, `RequestsService.CreateRequest`, `MachinesService.DrainMachine`, plus in-flight request completion during drain | Python `ProviderRuntime`, `create_request()`, and `drain_machine()`, plus TypeScript `ProviderRuntime`, `createRequest()`, and `drainMachine()` |

## Navigation Order For Contract Work

1. Start at `server/proto/service.proto`.
2. Trace the RPC handler in `server/pkg/service/server.go`.
3. Trace the owning domain file in `server/pkg/service/`.
4. Update the richest public wrapper first in `clients/python-client/`.
5. Update any narrower SDK wrappers in Go and TypeScript.
6. Confirm whether the change should also affect `conformance/cases/` or `conformance/schema/test_case.schema.json`.
