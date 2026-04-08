# Agent-Runtime Integration Seam

## Purpose

This document defines the stable, provider-agnostic seam between an external agent runtime and Toolplane.

Toolplane remains the durable remote tool-execution control plane. Edge adapters remain thin, session-bound translation layers that map another protocol onto that control plane without redefining request lifecycle, provider ownership, retained-window recovery, or drain semantics.

## Source Of Truth

- `server/proto/service.proto` defines the canonical external contract.
- `server/pkg/service/` owns the runtime semantics behind that contract.
- `SDK_MAP.md` is the truth source for which maintained SDK projection currently exposes which RPC family.
- The maintained Python and TypeScript `ProviderRuntime` surfaces package provider-side orchestration so adapters do not have to invent their own claim-submit lifecycle.

## Four-Layer Seam Model

| Layer | Owned by | Responsibilities |
| --- | --- | --- |
| Canonical control plane | `server/proto/service.proto`, `server/pkg/service/` | Session scope, request lifecycle, leases, retries, retained chunk windows, machine ownership, drain, and policy |
| Maintained SDK projections | Python, Go, and TypeScript plus `SDK_MAP.md` | Public wrappers over the control plane with explicit support and gap caveats |
| Provider runtime packaging | Python and TypeScript `ProviderRuntime` surfaces | Session attach or create, machine registration, tool registration, polling, claim, heartbeat, chunk append, result submission, stop, and drain |
| Edge adapters | Optional repo-edge adapters such as `clients/typescript-mcp-adapter/` | Session binding, protocol-specific discovery or invocation mapping, inspection presentation, and output shaping |

## Layer Responsibilities

### Layer 1: Canonical Control Plane

- The server owns request state, machine ownership, retry and lease behavior, retained chunk windows, and drain behavior.
- `ToolService`, `SessionsService`, `MachinesService`, `RequestsService`, and `TasksService` are the native capability families.
- No adapter should redefine these semantics locally.

### Layer 2: Maintained SDK Projections

- Python remains the richest maintained end-to-end reference surface.
- Go and TypeScript remain narrower maintained projections over the same control plane.
- `SDK_MAP.md` stays authoritative for which projection exposes consumer, provider, and admin responsibilities today.

### Layer 3: Provider Runtime Packaging

- Provider runtimes exist so integrations do not reimplement session attach or create, machine registration, tool registration, polling, claiming, heartbeats, result submission, streaming chunk append, stop, and drain.
- Python is the richest provider reference because it spans the maintained gRPC and HTTP gateway paths.
- TypeScript is the maintained JavaScript-family gRPC provider path.
- Go does not currently ship a maintained provider runtime harness, even though it exposes selected machine and tool lifecycle wrappers.

### Layer 4: Edge Adapters

- Edge adapters translate a foreign tool or agent protocol into Toolplane requests, streams, and inspection calls.
- They should stay bound to one explicit Toolplane session instead of inventing a hidden global runtime.
- They should be cheap to rewrite when upstream APIs change because the durable semantics live below them.

## Minimal Adapter Contract

An edge adapter should answer these questions without owning core runtime semantics itself.

| Concern | Native Toolplane mapping | Adapter rule |
| --- | --- | --- |
| Tool descriptors | `ListTools` for discovery or `RegisterTool` for provider registration | Preserve session scope and schema fidelity; do not invent a separate global catalog |
| Invocation | `CreateRequest` or `ExecuteTool` | Turn each foreign invocation into request-backed execution and keep the native request ID visible when possible |
| Cancellation | `CancelRequest` | If the foreign protocol supports cancellation, map it to the native request ID rather than a local task handle |
| Streaming | `StreamExecuteTool` on the consumer side, `AppendRequestChunks` and `SubmitRequestResult` on the provider side | Present chunks directly or aggregate them to fit the foreign protocol, but do not redefine replay semantics |
| Inspection | `GetRequest`, `ListRequests`, `GetSession`, `ListTools`, `GetMachine`, `ListMachines`, or adapter-facing read-only resources | Reuse native session, request, and machine records instead of creating a parallel model |
| Session binding | Explicit session ID | Keep session scope explicit per adapter process or connection |

## Current Maintained Gaps

- `ResumeStream` and `GetRequestChunks` exist at the server contract level, but they are not exposed as maintained public wrappers across every SDK projection.
- TypeScript provider mode is maintained, but the broader TypeScript public surface remains narrower than Python.
- The TypeScript MCP adapter is a useful edge example, not proof of general SDK parity.
- The current MCP adapter maps `tools/list`, `tools/call`, and read-only inspection resources, but it does not currently expose a separate MCP-native cancellation or replay surface.

## Reference Integration Pattern

1. Bind the adapter to one explicit Toolplane session.
2. Map foreign tool discovery or registration onto session-scoped Toolplane discovery or registration.
3. Translate each foreign invocation into request-backed execution on the control plane.
4. Keep request IDs visible so cancellation or inspection can reuse native control-plane records.
5. Translate streaming to the foreign protocol's presentation model without redefining retained-window replay.
6. Reuse the maintained provider runtime packaging instead of rebuilding claim, heartbeat, result submission, and drain logic inside the adapter.

## Reference Edge Example

The optional TypeScript MCP adapter is the repo's reference Layer 4 example.

- It binds one MCP stdio server process to one Toolplane session.
- It translates MCP `tools/list` to session-scoped native tool discovery.
- It translates MCP `tools/call` to request-backed execution and waits for a terminal native request state.
- It exposes read-only resources for the bound session, a concept map, and recent translated request records.
- It aggregates native stream chunks because MCP stdio tool calls do not expose the same long-lived request lifecycle as native Toolplane streams.

This is intentionally an edge translation layer. The source of truth stays in the native control plane and maintained SDK surfaces below it.

## Validation Trail

- `server/DOCUMENTATION.md` describes the native runtime behavior the seam must preserve.
- `SDK_MAP.md` documents which maintained SDK projection owns which part of the seam.
- `conformance/README.md` documents the maintained provider-runtime coverage and shared transport-neutral fixtures.
- `clients/typescript-client/tests/unit/provider_runtime.test.ts` exercises TypeScript provider-runtime unary, streaming, and drain behavior.
- `clients/typescript-mcp-adapter/tests/integration.test.ts` exercises the optional MCP edge adapter against a live Toolplane server and provider-backed tools.

If a future integration claim cannot be traced through those files, it should not be treated as part of the maintained seam.
