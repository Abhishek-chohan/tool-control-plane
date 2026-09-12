# Toolplane

[![Lint](https://github.com/Abhishek-chohan/tool-control-plane/actions/workflows/lint.yml/badge.svg)](https://github.com/Abhishek-chohan/tool-control-plane/actions/workflows/lint.yml)
[![Proto Drift](https://github.com/Abhishek-chohan/tool-control-plane/actions/workflows/proto-drift.yml/badge.svg)](https://github.com/Abhishek-chohan/tool-control-plane/actions/workflows/proto-drift.yml)
[![SDK Conformance & Verification](https://github.com/Abhishek-chohan/tool-control-plane/actions/workflows/conformance-python.yml/badge.svg)](https://github.com/Abhishek-chohan/tool-control-plane/actions/workflows/conformance-python.yml)
[![Release Gate](https://github.com/Abhishek-chohan/tool-control-plane/actions/workflows/release-gate.yml/badge.svg)](https://github.com/Abhishek-chohan/tool-control-plane/actions/workflows/release-gate.yml)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

Toolplane is a gRPC control plane for tool execution that has to outlive the caller that started it. It owns the request lifecycle — durable requests, provider machine ownership with leases and heartbeats, bounded stream replay after disconnects, and drain-safe rollouts — so callers don't rebuild that machinery around every remote tool.

A model being able to call a tool is not a reason to use Toolplane. If the work is quick, in-process, and shares the caller's lifecycle, call it directly.

## Status

Pre-1.0 and under active development; there are no tagged releases yet, and notable changes land under CHANGELOG `[Unreleased]` until the first tag — expect breaking changes. The wire contract is versioned: `api.v1` is the compatibility boundary per the [compatibility policy](server/docs/compatibility-policy.md). None of the packages are published to a registry, so install from source and pin to a commit.

## When to use it

Use Toolplane when at least one of these is true:

- The tool may outlive the model turn, HTTP request, or caller process.
- Execution must survive consumer disconnects, with results inspectable or replayable afterwards.
- The provider needs explicit machine ownership, heartbeats, or drain behavior for safe rollouts.
- Streaming output needs bounded replay or later inspection.
- Restarts and deploys must not blindly drop in-flight remote work.
- The workload is queue-backed, multi-worker, or capacity-limited enough that request lifecycle control is load-bearing.

When none of these hold, a plain queue with idempotent workers — or a direct function call — is the simpler right answer.

## Quickstart

A Toolplane deployment is three cooperating processes: the **server** (owns state), a **provider** (registers machine-backed tools and executes requests), and a **consumer** (discovers tools and submits work). The quickstart paths below all follow that shape.

Prerequisites: Go 1.24+, Python 3.8–3.12, and `make`/bash (on Windows, use WSL). Docker is only needed for the reference deployment. Install the Python client dependencies once:

```bash
pip install -r clients/python-client/requirements.txt
```

### One command

From `server/`:

```bash
make demo
```

This boots an in-memory server, starts the Python provider example (registers tools and serves them), then runs the consumer example against the provider's session — discovery, invocation, and result polling, end to end. Everything is torn down on exit; logs land in `server/.tmp-demo/`.

### Manually

Development defaults are explicit and non-secret: in-memory storage, fixed auth, no TLS.

```bash
cd server
export TOOLPLANE_ENV_MODE=development
export TOOLPLANE_AUTH_MODE=fixed
export TOOLPLANE_AUTH_FIXED_API_KEY=dev-key
export TOOLPLANE_STORAGE_MODE=memory

make build && ./bin/toolplane-server --port 9001 &
go run ./cmd/proxy --listen :8080 --backend localhost:9001
```

The server and gateway alone don't execute any tools — start a provider (see the SDK snippets below) to register tools, then invoke from a consumer sharing the same session. These defaults are for local work and CI only; production mode requires Postgres storage, Postgres-backed auth, and gRPC TLS, and both gateways require explicit allowed origins (`TOOLPLANE_PROXY_ALLOWED_ORIGINS`, `TOOLPLANE_MCP_ALLOWED_ORIGINS`). See [server/docs/local-development.md](server/docs/local-development.md).

### Reference deployment (Docker Compose)

The production-shaped stack lives at `server/deploy/reference/compose.yaml`: Postgres 16, a one-shot migrate service, the server with TLS, the HTTP gateway, and the MCP gateway. There are no default credentials — copy `server/deploy/reference/.env.example` and set a Postgres password first. See [server/docs/reference-deployment.md](server/docs/reference-deployment.md).

## Architecture

The canonical API contract is `server/proto/service.proto` (package `api.v1`, 44 RPCs across five services). The Go server owns the runtime semantics; every other component is a projection of them.

| Component | Binary | Default port | Purpose |
| --- | --- | --- | --- |
| Control plane server | `server/cmd/server` | `9001` (gRPC) | Contract, request/machine/task lifecycle, storage, auth; Prometheus metrics on a separate listener (loopback-random by default, `:9102` in the reference deployment) |
| HTTP/JSON gateway | `server/cmd/proxy` (`toolplane-gateway`) | `8080` | grpc-gateway v2 transcoding with CORS, rate limiting, and circuit breaker |
| Model Context Protocol (MCP) gateway | `server/cmd/mcp-gateway` (`toolplane-mcp-gateway`) | `8081` | Stateless MCP JSON-RPC facade over the gRPC backend |

## Runtime guarantees

Details in [server/DOCUMENTATION.md](server/DOCUMENTATION.md).

- **Request lifecycle**: `PENDING → CLAIMED → RUNNING → DONE/FAILED` with explicit claim, cancel, and result submission.
- **Leases and fencing**: 30s renewable execution leases; storage-level fenced writes (machine ID + lease epoch) reject stale executors after a reclaim. Reclaimed requests re-execute under a fresh lease, so tools should be idempotent.
- **Active-active safe**: no-double-claim and fenced requeue hold across instances; proven by a two-replica suite sharing one Postgres in the release gate.
- **Bounded stream replay**: the newest 100 chunks (up to 8 MiB) are retained with absolute sequence numbers; `ResumeStream` replays them after a disconnect and returns `OUT_OF_RANGE` once trimmed or expired.
- **Machine ownership**: per-session registration, heartbeat TTL, per-machine in-flight capacity, and `DrainMachine` that stops new routing while in-flight work finishes.
- **Tasks**: fire-and-forget orchestration with retries, backoff, dead-lettering, and cancellation that propagates to the underlying request.
- **Idempotency**: `idempotency_key` on request creation deduplicates.
- **Auth**: API keys with explicit capabilities (`read`/`invoke`/`provide`/`admin`), scoped to their session; the dev-only fixed key is shared full access and is rejected in production mode.

## SDKs

`SDK_MAP.md` is the truth source for per-RPC parity — the SDK surfaces are intentionally different, so don't infer parity from folder names.

| SDK | Install | Transports | Provider runtime |
| --- | --- | --- | --- |
| [Python](clients/python-client/README.md) | `pip install -e clients/python-client` | gRPC + HTTP, sync + async | Yes |
| [Go](clients/go-client/README.md) | `cd clients/go-client && go mod tidy` | gRPC | No (raw machine wrappers only) |
| [TypeScript](clients/typescript-client/README.md) | `cd clients/typescript-client && npm install && npm run build` | gRPC | Yes |
| [MCP adapter](clients/typescript-mcp-adapter/README.md) | build typescript-client first, then `cd clients/typescript-mcp-adapter && npm install && npm run build` | stdio | — |

A provider and consumer in one process, using the manual server from the quickstart:

```python
from toolplane import Toolplane

client = Toolplane(server_host="localhost", server_port=9001, api_key="dev-key")
runtime = client.provider_runtime()
session = runtime.create_session(user_id="demo", name="demo")

@runtime.tool(session_id=session.session_id, name="add", description="Add two numbers")
def add(a: int, b: int) -> int:
    return a + b

runtime.start_in_background()
print(client.invoke("add", session.session_id, a=2, b=3))  # -> 5
runtime.stop()
```

Tools are session-scoped: a consumer can only invoke tools registered into the *same* session, so separate provider and consumer processes must share a session ID (this is what `make demo` arranges). The runnable walkthroughs are [clients/python-client/example_client.py](clients/python-client/example_client.py) (provider) and [example_user.py](clients/python-client/example_user.py) (consumer) — see [clients/python-client/README_EXAMPLES.md](clients/python-client/README_EXAMPLES.md). The examples default to `TOOLPLANE_API_KEY=toolplane-conformance-fixture-key`, matching `server/.env.example`; if you started the server with a different fixed key, export `TOOLPLANE_API_KEY` to match.

## Use with MCP clients

`toolplane-mcp-gateway` is a stateless JSON-RPC facade that exposes Toolplane tools over MCP at `POST /mcp` (health: `GET /health`). It holds no request state, so any number of instances can serve any request.

- **MCP 2026-07-28 stateless clients** connect directly: each request declares the protocol version in `_meta`; no handshake is needed.
- **Initialize-based clients** (2025-03-26 through 2025-11-25 revisions) connect through the built-in compatibility path: standard `initialize` handshake, then plain `tools/list` / `tools/call`.
- **Session binding**: 2026 clients set `dev.toolplane/session_id` in `_meta`; otherwise the gateway provisions one session per API key.
- **Tasks extension**: clients advertising `io.modelcontextprotocol/tasks` get durable task handles from `tools/call`, poll with `tasks/get` (including bounded stream replay behind a `dev.toolplane/last_seq` cursor), and cancel with `tasks/cancel`. Other clients are served synchronously.
- **Auth**: forward the Toolplane API key as `Authorization` or `X-API-Key`.

```bash
toolplane-mcp-gateway --listen :8081 --backend localhost:9001
```

For stdio-only environments, `clients/typescript-mcp-adapter` wraps one Toolplane session behind a stdio transport and handles both protocol generations.

## Development and testing

`conformance/cases/` holds 22 transport-neutral JSON fixtures covering sessions, requests, bounded stream replay, invocation, machines, provider runtime, multi-instance, and MCP. Python and TypeScript run every fixture over both gRPC and HTTP; Go has opt-in live integration tests instead of the shared harness.

```bash
cd server
make test-race                    # full Go server suite under -race
make python-unit                  # Python SDK unit tests
make conformance-python           # shared-fixture conformance (auto-boots a server)
make conformance-python-mcp       # adds the MCP transport (boots the MCP gateway)
make release-gate                 # authoritative gate: conformance + observability + runtime slices
make check-proto-drift            # regenerated stubs must match what's committed
```

CI runs Lint, SDK Conformance & Verification, and Release Gate on every push to `main` and pull request; Proto Drift runs when proto inputs change. Release Gate is the authoritative gate: a Postgres-backed end-to-end scenario plus multi-instance and MCP slices. The pin set for proto regeneration is in [server/docs/proto_regeneration.md](server/docs/proto_regeneration.md).

## Contributing

Issues are welcome for bugs, questions, and design discussion. Before pushing a PR, run the affected checks above — at minimum `make test-race`, `make conformance-python`, and `make check-proto-drift` from `server/`. Proto changes must follow [server/docs/proto_regeneration.md](server/docs/proto_regeneration.md); SDK surface changes update [SDK_MAP.md](SDK_MAP.md) and `CHANGELOG.md` in the same PR, and regenerate the README API sections with `python tools/gen_sdk_readmes.py` (CI enforces this via the drift check). Local setup: [server/docs/local-development.md](server/docs/local-development.md).

Report security issues privately via [SECURITY.md](SECURITY.md) — please don't open public issues for them.

## Repository layout

```text
server/                      Go control plane: cmd/server, cmd/proxy, cmd/mcp-gateway,
                             pkg/service, pkg/storage (memory + postgres), pkg/mcp,
                             deploy/reference (production compose stack)
clients/
  python-client/             Primary SDK (gRPC + HTTP, sync + async)
  go-client/                 Go SDK (separate Go module)
  typescript-client/         TypeScript SDK
  typescript-mcp-adapter/    MCP stdio adapter (depends on typescript-client via file:)
conformance/                 Shared transport-neutral fixtures + JSON schema
SDK_MAP.md                   Per-RPC SDK parity map
CHANGELOG.md                 Keep-a-Changelog; detailed notes under server/docs/release-notes/
LICENSE                      Apache-2.0
```

## Documentation

- [server/DOCUMENTATION.md](server/DOCUMENTATION.md) — runtime semantics, request lifecycle, drain, bounded stream recovery.
- [server/docs/local-development.md](server/docs/local-development.md) — supported local bootstrap.
- [server/docs/reference-deployment.md](server/docs/reference-deployment.md) — production topology, rollout, drain, rollback, validation.
- [server/docs/operator-runbook.md](server/docs/operator-runbook.md) — symptom-first day-2 workflows.
- [server/docs/reliability-drills.md](server/docs/reliability-drills.md) — named failure drills, expected outcomes, and what the drills do not claim.
- [server/docs/compatibility-policy.md](server/docs/compatibility-policy.md) — protobuf, gateway, and SDK compatibility rules; `api.v1` is the version boundary.
- [server/docs/agent-runtime-integration-seam.md](server/docs/agent-runtime-integration-seam.md) — the integration seam for external agent runtimes and adapters.
- [server/docs/incremental-adoption.md](server/docs/incremental-adoption.md) — stepwise first-tool migration guide.
- [server/docs/economic-case.md](server/docs/economic-case.md) — the operational-simplification argument and decision rubric (background).
- [SDK_MAP.md](SDK_MAP.md) — per-RPC parity and support tiers.

## License

[Apache License 2.0](LICENSE)
