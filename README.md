# Toolplane

Toolplane is a remote tool-execution control plane for tools that outlive a single caller, process, or deploy.

> Public project name: `Toolplane`
> Intended repository slug: `tool-control-plane`

It exposes a protobuf/gRPC contract, a maintained HTTP gateway compatibility layer, and optional ecosystem adapters. The canonical API is `server/proto/service.proto`, and the Go server owns the runtime semantics.

## Decision Rule

Use Toolplane when at least one of the following is true and direct in-process tool calling would otherwise force the caller to own remote execution concerns itself:

- The tool may outlive the model turn, HTTP request, websocket, or caller process.
- Execution must survive or recover cleanly from consumer disconnects.
- The provider needs explicit machine ownership, heartbeats, or drain behavior.
- Streaming output needs bounded replay or later inspection after disconnect.
- Deploys or restarts must avoid blindly dropping in-flight remote work.
- The workload is queue-backed, multi-worker, or capacity-limited enough that request lifecycle control matters.

Do not use Toolplane just because a model can call a tool. If the work is quick, in-process, same-lifecycle, and not operationally sensitive, direct tool calling is simpler.

## Quick Comparison

| Approach | Best fit | Not enough when |
| --- | --- | --- |
| Direct tool calling | Work is quick, local, and shares the caller lifecycle | The caller would need to add request persistence, replay, provider ownership, or drain behavior |
| Queue-only execution | Deferred work is enough and operators do not need request inspection or replay | The workload also needs retained-window recovery, explicit machine ownership, or deploy-safe drain |
| Thin adapter layers | You are mapping an external tool or agent protocol onto an existing Toolplane session | The adapter itself is being treated as the product wedge |
| Toolplane | One remote tool is long-running, stateful, streaming, restart-sensitive, deploy-sensitive, or multi-worker | The work is already simple enough for a local function or a thin synchronous network call |

## Reference Workload

A strong first offload candidate is one sandboxed code-execution worker. It runs in a separate environment, may outlive the original caller, benefits from request inspection or cancellation, and often needs drain-safe rollout plus later recovery after disconnect. The bundled examples stay intentionally simple, but they follow the same control-plane shape.

## What Toolplane Is Not

- Not a generic RPC tutorial layer.
- Not a thin model-SDK tool wrapper.
- Not an MCP specification or a one-for-one MCP replacement.

## Repository Shape

- A Go server that owns the contract, request lifecycle, machine lifecycle, and task orchestration for distributed tool execution.
- A multi-SDK repo where Python is the richest current client surface, and Go and TypeScript are narrower maintained gRPC clients.
- A conformance-driven codebase with shared transport-neutral fixtures in `conformance/` for supported public behaviors.

## Validated Wedge

- Explicit provider runtimes in Python and TypeScript that own machine-backed execution.
- Request inspection plus retained-window replay after disconnect.
- Postgres-backed restart recovery for request state and lease expiry handling.
- Machine drain that stops new routing immediately and lets in-flight work finish or age out.
- Maintained SDKs with intentionally different public surfaces. Read `SDK_MAP.md` before assuming parity from folder names alone.

## Support Snapshot

| Surface | Status | Notes |
| --- | --- | --- |
| Go server + protobuf contract | Primary | Source of truth lives in `server/proto/service.proto` and `server/pkg/service/` |
| Python client | Primary maintained SDK | Richest end-to-end surface across gRPC and the maintained HTTP gateway |
| Go client | Supported secondary SDK | Maintained gRPC lifecycle, request, and task helpers; no provider runtime harness |
| TypeScript client | Supported secondary SDK | Maintained JavaScript-family gRPC client with an explicit `ProviderRuntime`; repository-internal HTTP adapters remain conformance-only |
| TypeScript MCP adapter | Optional ecosystem adapter | Stdio adapter for one Toolplane session with tool and resource access |

## Agent-Runtime Seam

Toolplane keeps external runtime integrations on a stable four-layer seam so adapters stay thin and replaceable:

- Layer 1: `server/proto/service.proto` plus the Go server runtime own session scope, request lifecycle, retained replay, machine ownership, and drain semantics.
- Layer 2: the maintained SDK projections expose the public wrappers available today; `SDK_MAP.md` is the truth source for current surface area and gaps.
- Layer 3: Python and TypeScript `ProviderRuntime` surfaces package provider-side session attach or create, machine and tool registration, polling, claim, heartbeat, result submission, and drain.
- Layer 4: edge adapters bind one explicit Toolplane session and translate foreign discovery, invocation, and inspection shapes without redefining lifecycle semantics.

See `server/docs/agent-runtime-integration-seam.md` for the full seam model, minimal adapter contract, current gap caveats, and the reference edge pattern.

## Reliability Proof

The maintained failure story is packaged as six drills: provider crash or lease expiry mid-run, caller disconnect during streaming, replay behind the retained window, drain under load, durable restart recovery, and claim-state or capacity safety. See `server/docs/reliability-drills.md` for triggers, expected outcomes, explicit limits, and the first runnable validation path.

## First Offload Path

The maintained first-touch path is a Python provider and consumer pair because it shows how to offload one painful remote tool first on the richest SDK surface. Use it as the scaffold for a sandboxed code-execution worker or another environment-bound tool that needs explicit provider ownership, request inspection, and drain-safe rollout:

1. **Provider** (`clients/python-client/example_client.py`): connects via gRPC, creates a session, registers machine-backed tools, and starts the explicit provider loop.
2. **Consumer** (`clients/python-client/example_user.py`): joins the same session, lists tools, invokes provider-backed work, and polls request state.

Run the provider first. Copy the printed `TOOLPLANE_SESSION_ID`, then run the consumer with that value. See `clients/python-client/README_EXAMPLES.md` for environment defaults and the full example flow. The sample tools stay intentionally simple so the lifecycle is easy to trace; replace them with the first remote tool you want to offload without replacing the rest of the caller stack.

The intended migration shape is explicit: bind one painful remote tool to one Toolplane session, keep the surrounding agent runtime or orchestration layer as the caller of record, and exercise inspection, cancellation, and drain before moving a second tool. See `server/docs/incremental-adoption.md` for the stepwise first-tool guide.

For runtime semantics behind this flow, including request lifecycle, streaming, retained-window recovery, and machine drain, see `server/DOCUMENTATION.md`.

## Economic Case

Toolplane's economic case is not raw cost savings. For wedge workloads, the repo argues for operational simplification: request lifecycle, replay, provider drain, operator signals, rollout checks, and compatibility expectations move into one maintained control plane instead of being rebuilt around each remote tool family.

That case is intentionally narrow. The control-plane layer is justified only when one remote tool already forces the team to own retry, recovery, inspection, or rollout logic itself. Direct tool calling remains simpler for short-lived same-lifecycle work. See `server/docs/economic-case.md` for the full simplification memo, before-and-after scenario, and decision rubric.

## Start Here

- `server/DOCUMENTATION.md`: runtime semantics, request lifecycle, drain behavior, and retained-window recovery.
- `server/docs/agent-runtime-integration-seam.md`: four-layer integration seam, minimal adapter contract, current gap caveats, and the reference edge pattern.
- `server/docs/reliability-drills.md`: named failure drills, expected outcomes, explicit limits, and the shortest validation path.
- `server/docs/incremental-adoption.md`: stepwise first-tool migration, coexistence model, reference workload shape, and validation path.
- `server/docs/economic-case.md`: operational simplification memo, before-and-after first-tool scenario, and the decision rubric against simpler options.
- `server/docs/operator-runbook.md`: symptom-first operator workflows for queue stalls, stuck requests, drain, throttling, policy denial, and safe control actions.
- `server/docs/reference-deployment.md`: maintained production topology, bootstrap, rollout, drain, rollback, and validation path.
- `server/docs/compatibility-policy.md`: protobuf, HTTP gateway, and maintained SDK compatibility rules.
- `server/docs/local-development.md`: explicit local bootstrap path with env-based auth, storage, and proxy settings.
- `SDK_MAP.md`: per-RPC parity plus support-tier caveats.
- `clients/typescript-mcp-adapter/README.md`: optional stdio adapter usage, session binding, and validation path.

## Local Development

Use `server/.env.example` plus `server/docs/local-development.md` for the supported bootstrap path. The development default is explicit and intentionally non-secret:

- `TOOLPLANE_ENV_MODE=development`
- `TOOLPLANE_AUTH_MODE=fixed`
- `TOOLPLANE_AUTH_FIXED_API_KEY=toolplane-conformance-fixture-key`
- `TOOLPLANE_STORAGE_MODE=memory`

That path exists for local work and CI fixtures only. Production-oriented startup should move to explicit auth and storage configuration.

The maintained production reference path is the split `Postgres + migrate + server + gateway` stack documented in `server/docs/reference-deployment.md`.

Production mode requires maintained auth, Postgres-backed storage, explicit proxy origins, and operator-visible runtime diagnostics.
