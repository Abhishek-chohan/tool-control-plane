# Release Gate

This document defines the authoritative release signal for the maintained Toolplane architecture. The gate now runs the canonical end-to-end scenario against Postgres in CI, then follows with a narrow Postgres-backed platform slice so the repo proves more than a development-only happy path.

## What It Proves

The release gate exercises one canonical end-to-end path and one narrow platform-validation slice:

1. **Secure startup**: server and gateway start through the env-driven configuration path with fixed-key auth and Postgres-backed storage in CI — no committed secrets, no hardcoded bootstrap.
2. **Provider registration**: a machine registers and advertises tools via the maintained gRPC control-plane surface.
3. **Session creation**: a session is created to scope subsequent operations.
4. **Unary execution**: a tool invocation completes synchronously and returns a matching result.
5. **Streaming execution**: a streaming tool invocation delivers ordered chunks with a terminal marker.
6. **Request inspection**: the request lifecycle is visible to the consumer after execution.
7. **Graceful drain**: `DrainMachine` stops new routing, waits for in-flight work, then unregisters the machine.
8. **Persistence-backed recovery and policy checks**: focused Go tests verify that production mode rejects in-memory storage, session and task audit events are emitted, and expired persisted requests are requeued after a restart-like service reload.

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

This target always runs the Python shared-fixture conformance suite, which exercises the first seven steps above through the auto-boot harness in `clients/python-client/tests/conformance/conftest.py`. When `TOOLPLANE_DATABASE_URL` is set, the Makefile infers `TOOLPLANE_STORAGE_MODE=postgres` so the full conformance leg runs on Postgres rather than falling back to in-memory storage. The maintained provider path in that suite now runs through the explicit Python `ProviderRuntime` surface rather than ad hoc polling threads embedded in the adapters.

If `TOOLPLANE_DATABASE_URL` points at a reachable Postgres instance, the same target also runs focused Go tests that validate the production storage guardrail and persisted request recovery path. If the variable is unset, the conformance leg falls back to in-memory storage and the persisted recovery test is skipped for local convenience.

## CI

The `.github/workflows/release-gate.yml` workflow runs the same `make release-gate` command on every push to `main` and on pull requests. CI now provisions a Postgres service for that job and sets `TOOLPLANE_STORAGE_MODE=postgres`, so both the canonical Python conformance leg and the focused persistence-backed recovery slice execute against the same database-backed runtime. It remains the repo's authoritative release signal, distinct from the broader SDK Conformance & Verification workflow which covers all three maintained SDKs.

## Relationship To Other Workflows

| Workflow | Role |
| --- | --- |
| `release-gate.yml` | Authoritative release gate: one canonical secure end-to-end scenario |
| `conformance-python.yml` | SDK Conformance & Verification: cross-SDK shared-fixture coverage for Python, Go, and TypeScript |

The release gate remains intentionally narrow and fast. Shared conformance is broader and verifies parity across SDKs. Both must pass before a release is trusted.

Provider-runtime coverage in shared conformance now explicitly includes claim-and-submit, chunk append, and drain-under-load behavior on the maintained Python provider harness.

## What It Does Not Prove

- Live Postgres-backed API-key validation. The gate now proves the production storage guardrail and a Postgres-backed recovery path, but it does not exercise that production auth backend directly.
- The separate server-side `/rpc` reference path documented in `server/docs/rpc-retirement.md`.
- Tool-discovery RPCs without shared fixture coverage (see `SDK_MAP.md` for current `partial` claims).
- Every transport permutation (HTTP gateway behavior is covered by shared conformance, not duplicated here).
