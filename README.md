# Toolplane

Toolplane is a remote tool-execution control plane for tools that outlive a single caller, process, or deploy.

> Public project name: `Toolplane`
> Intended repository slug: `tool-control-plane`

It exposes a protobuf/gRPC contract, a maintained HTTP gateway compatibility layer, and optional ecosystem adapters. The canonical API is `server/proto/service.proto`, and the Go server owns the runtime semantics.

If your tools are long-running, streaming, restart-sensitive, or deploy-sensitive, Toolplane gives you explicit provider runtimes, request lifecycle control, retained-window stream recovery, and drain-safe rollout behavior over one canonical contract. If your tools are short-lived local functions, direct tool calling is usually simpler.

## Use Toolplane When

- Work may outlive one HTTP request, websocket, model response, or process.
- Providers need explicit machine ownership, heartbeats, and drain semantics.
- Callers need request inspection or bounded stream recovery after disconnect.
- Restarts or deploys must preserve control-plane state well enough to resume routing safely.

## Direct Tool Calling Is Probably Enough When

- Tools run in-process and finish quickly.
- Queueing, retained-window recovery, and drain are unnecessary.
- The caller and the tool worker share the same lifecycle and failure domain.

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

## Canonical Flow

The maintained first-touch path is a Python provider and consumer pair because it shows the full control-plane lifecycle on the richest SDK surface:

1. **Provider** (`clients/python-client/example_client.py`): connects via gRPC, creates a session, registers machine-backed tools, and starts the explicit provider loop.
2. **Consumer** (`clients/python-client/example_user.py`): joins the same session, lists tools, invokes provider-backed work, and polls request state.

Run the provider first. Copy the printed `TOOLPLANE_SESSION_ID`, then run the consumer with that value. See `clients/python-client/README_EXAMPLES.md` for environment defaults and the full example flow.

For runtime semantics behind this flow, including request lifecycle, streaming, retained-window recovery, and machine drain, see `server/DOCUMENTATION.md`.

## Start Here

- `server/DOCUMENTATION.md`: runtime semantics, request lifecycle, drain behavior, and retained-window recovery.
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
