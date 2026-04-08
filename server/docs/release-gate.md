# Release Gate

This document defines the authoritative release signal for the maintained Toolplane architecture. The gate now runs the canonical end-to-end scenario against Postgres in CI, then follows with a narrow runtime slice so the repo proves more than a development-only happy path.

The maintained production topology itself is documented in `server/docs/reference-deployment.md` and materialized in `server/deploy/reference/compose.yaml`.

The relationship is deliberate:

- `reference-deployment.md` is the operator story for running, bootstrapping, upgrading, draining, and rolling back the split `Postgres + migrate + server + gateway` stack.
- `release-gate.md` is the runtime validation story for that same stack shape and Postgres-backed contract.
- The gate intentionally does not prove one-time TLS certificate provisioning or the bootstrap fixed-auth bridge used to mint the first Postgres-backed admin key; use `make reference-deployment-integration` for that deployment-specific rehearsal.

## Durability Proof Surface

The release signal is intentionally split into three categories so the repo can state what is proven directly, what is proven by focused runtime tests, and what remains a documented limit.

## Named Failure Drill Coverage

See `server/docs/reliability-drills.md` for the full drill matrix. The release gate maps onto that matrix as follows.

| Drill | Release-gate coverage | Note |
| --- | --- | --- |
| D1. Provider crash or lease expiry mid-run | `release-gate-runtime` | Focused lease-expiry requeue proof, not a shared fixture |
| D2. Caller disconnect during streaming | Shared conformance leg plus `release-gate-runtime` | Portable request-recovery fixtures plus retained-replay runtime assertions |
| D3. Replay behind the retained window | Shared conformance leg plus `release-gate-runtime` | Explicit `out_of_range` checks remain part of the authoritative gate |
| D4. Deploy drain with in-flight work | Shared conformance leg plus `release-gate-runtime` | Provider-runtime and machine-lifecycle drain-under-load fixtures plus drain waiting tests |
| D5. Restart recovery with durable storage | Postgres-backed `release-gate-runtime` | Skipped when `TOOLPLANE_DATABASE_URL` is unset |
| D6. Claim-state and capacity safety | Focused runtime slice only | Drain-blocked create or claim paths are covered; standalone machine-capacity rejection is not yet a named gate leg |

## Observability Coverage

Observability is now part of the maintained runtime contract, and the release gate now includes one live scrape leg in addition to the focused tests.

### Live Endpoint Scrape

- `server/scripts/release_gate_observability.sh` starts a live server with an explicit metrics listener, starts a live proxy, scrapes `/metrics`, and scrapes `/health`
- that leg asserts the maintained runtime metric names plus the maintained top-level `/health` payload fields and nested circuit and throttle structures

### Proven By Focused Tests

- stable `requestId` and `taskId` trace correlation fields are asserted in `server/pkg/service/request_runtime_test.go` and `server/pkg/service/task_test.go`
- the supported runtime metrics names and values are asserted in `server/pkg/observability/runtime_metrics_test.go`
- proxy `/health`, throttle counter accounting, redacted throttle logs, and `Retry-After` behavior are asserted in `server/cmd/proxy/observability_test.go`

### Documented Operator Contract

- `server/docs/observability.md` is the maintained summary of server traces, runtime metrics, proxy `/health`, proxy throttle logs, correlation keys, redaction rules, and the operator playbook

### Proven Directly By `make release-gate`

The shared-fixture conformance leg proves one canonical provider-backed flow and the portable recovery behavior that goes with it:

1. **Secure startup**: server and gateway start through the env-driven configuration path with fixed-key auth and Postgres-backed storage in CI — no committed secrets, no hardcoded bootstrap.
2. **Provider registration**: a machine registers and advertises tools via the maintained gRPC control-plane surface.
3. **Session creation**: a session is created to scope subsequent operations.
4. **Unary execution**: a tool invocation completes synchronously and returns a matching result.
5. **Streaming execution**: a streaming tool invocation delivers ordered chunks with a terminal marker.
6. **Request inspection and bounded replay**: `GetRequestChunks()` exposes retained-window metadata, `ResumeStream()` replays from within the retained window, and replay before the retained window fails with the canonical `out_of_range` error.
7. **Graceful drain**: drain-under-load fixtures prove that new routing stops while in-flight work still completes before the machine disappears.

These map directly onto drills D2 through D4 in `server/docs/reliability-drills.md`, and they provide the shared-fixture portion of the authoritative release story.

### Proven By Focused Go Tests In `release-gate-runtime`

The focused runtime slice keeps the gate narrow while making the durability claims precise:

- Production-mode startup rejects in-memory storage and allows the maintained Postgres-backed path.
- In-memory lease expiry requeues leased work and frees machine load.
- Postgres-backed restart-like reload requeues an expired persisted lease.
- Graceful drain waits for both running and claimed in-flight work to resolve before unregistering the machine.
- Stream replay returns ordered retained chunks plus the terminal marker, including replay from inside a trimmed retained window.
- Replay before the retained window still fails with `OUT_OF_RANGE` instead of silently skipping lost chunks.

These focused tests provide the narrow runtime proof for drills D1 through D5 and the drain-blocked routing slice of D6.

### Documented Limits, Not Durability Guarantees

- Toolplane does **not** currently promise durable full-history stream replay beyond the retained 100-chunk window.
- `TOOLPLANE_STORAGE_MODE=memory` is intentionally non-durable; restart recovery guarantees do not apply there.
- Pending requests are not automatically migrated during drain; they remain pending until another machine owns the tool.
- The release gate still does not directly exercise live Postgres-backed API-key validation.

## Env Contract

The release gate uses the same env contract as shared conformance. All values below are intentionally non-secret fixture defaults suitable for CI and local reproduction only.

| Variable | Value | Purpose |
| --- | --- | --- |
| `TOOLPLANE_ENV_MODE` | `development` | Allows fixed auth and insecure backend |
| `TOOLPLANE_AUTH_MODE` | `fixed` | Fixed API-key authentication |
| `TOOLPLANE_AUTH_FIXED_API_KEY` | `toolplane-conformance-fixture-key` | Non-secret fixture key |
| `TOOLPLANE_STORAGE_MODE` | `postgres` in CI, else inferred from `TOOLPLANE_DATABASE_URL` | Runs the canonical gate against Postgres when a database is configured |
| `TOOLPLANE_DATABASE_URL` | required in CI, optional locally | Connects the canonical gate and recovery tests to Postgres |
| `TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND` | `1` | Allow plaintext gRPC backend in dev |
| `TOOLPLANE_CONFORMANCE_AUTO_BOOT` | `1` | Auto-start server and proxy |
| `TOOLPLANE_CONFORMANCE_API_KEY` | `toolplane-conformance-fixture-key` | Conformance runner auth token |
| `TOOLPLANE_CONFORMANCE_USER_ID` | `conformance-user` | Fixture user identity |

## Local Reproduction

From the repository root:

```bash
cd server && TOOLPLANE_DATABASE_URL=postgres://postgres:postgres@localhost:5432/toolplane?sslmode=disable make release-gate
```

If you want the Postgres instance to come from the maintained reference topology instead of an ad hoc local database, start it first with:

```bash
cd server/deploy/reference && docker compose up -d postgres
```

Then run the same `make release-gate` command from `server/`, but point `TOOLPLANE_DATABASE_URL` at the host-published port from `server/deploy/reference/.env.example`. With the default reference env file that is:

```bash
cd server && TOOLPLANE_DATABASE_URL=postgres://toolplane:toolplane@localhost:5432/toolplane?sslmode=disable make release-gate
```

This target always runs the Python shared-fixture conformance suite, then the live observability scrape leg in `server/scripts/release_gate_observability.sh`, and finally the focused runtime slice. The conformance portion exercises the first seven steps above through the auto-boot harness in `clients/python-client/tests/conformance/conftest.py`. When `TOOLPLANE_DATABASE_URL` is set, the Makefile infers `TOOLPLANE_STORAGE_MODE=postgres` so the full conformance leg runs on Postgres rather than falling back to in-memory storage. The maintained provider path in that suite now runs through the explicit Python `ProviderRuntime` surface rather than ad hoc polling threads embedded in the adapters.

If `TOOLPLANE_DATABASE_URL` points at a reachable Postgres instance, the same target also runs focused Go tests that validate the production storage guardrail, lease-expiry requeue behavior, bounded replay semantics, drain waiting behavior, and persisted request recovery path. If the variable is unset, the conformance leg falls back to in-memory storage and the persisted recovery test is skipped for local convenience.

## CI

The `.github/workflows/release-gate.yml` workflow runs the same `make release-gate` command on every push to `main` and on pull requests. CI now provisions a Postgres service for that job and sets `TOOLPLANE_STORAGE_MODE=postgres`, so the canonical Python conformance leg, the live observability scrape leg, and the focused persistence-backed recovery slice execute against the same database-backed runtime. It remains the repo's authoritative release signal, distinct from the broader SDK Conformance & Verification workflow which covers all three maintained SDKs.

## Relationship To Other Workflows

| Workflow | Role |
| --- | --- |
| `release-gate.yml` | Authoritative release gate: one canonical secure end-to-end scenario |
| `conformance-python.yml` | SDK Conformance & Verification: cross-SDK shared-fixture coverage for Python, Go, and TypeScript |

The release gate remains intentionally narrow and fast. Shared conformance is broader and verifies parity across SDKs. Both must pass before a release is trusted.

Provider-runtime coverage in shared conformance now explicitly includes claim-and-submit, chunk append, bounded replay recovery, and drain-under-load behavior on the maintained Python provider harness.

## What It Does Not Prove

- Live Postgres-backed API-key validation. The gate now proves the production storage guardrail and a Postgres-backed recovery path, but it does not exercise that production auth backend directly.
- Reference deployment TLS certificate provisioning or the proxy's custom CA bundle wiring. Use `make reference-deployment-integration` when you need that exact deployment path exercised.
- The separate server-side `/rpc` reference path documented in `server/docs/rpc-retirement.md`.
- Tool-discovery RPCs without shared fixture coverage (see `SDK_MAP.md` for current `partial` claims).
- Every transport permutation (HTTP gateway behavior is covered by shared conformance, not duplicated here).
