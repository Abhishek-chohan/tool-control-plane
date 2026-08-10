# Concerns, Risks & Technical Debt

> Items below are derived from code, config, and git history — not from roadmap
> aspirations. Test-only TODOs and vendored dependencies are separated from
> production debt per the skill guidance.

## Core Sections (Required)

### 1) Production TODO / FIXME / HACK

A recursive scan of production source (excluding generated proto dirs, tests,
conformance, and venvs) found **no `TODO`/`FIXME`/`HACK` markers in maintained
Go, Python, or TypeScript source**. The only matches are inside
`server/.proto-venv/` (a local, **gitignored** Python toolchain used for proto
regeneration — third-party code, not production source).

Evidence: `grep -rniE "TODO|FIXME|HACK" server/ clients/` filtered to
`.go/.ts/.py` excluding `proto/gen/google/dist/node_modules/tests/conformance`.

### 2) High-Churn Files

The repo has **24 total commits, all between 2026-04-04 and 2026-04-08**
(verified via `git log --format=%ci`). Because every commit is older than 90
days from this mapping date (2026-08-09), the standard 90-day window is empty;
the churn list below uses a 365-day window. Most-touched files:

| Count | File |
|-------|------|
| 8 | `server/docs/release-gate.md` |
| 8 | `server/DOCUMENTATION.md` |
| 6 | `server/cmd/server/main.go` |
| 5 | `server/docs/local-development.md` |
| 5 | `clients/typescript-client/src/core/toolplane_client.ts` |
| 5 | `clients/python-client/README.md` |
| 5 | `SDK_MAP.md` |

Theme: the highest churn is in **documentation and the server entrypoint
(`main.go`)**, followed by the TypeScript client core and Python client facade.
This is consistent with the recent commit messages (docs, TLS, auth,
observability). No single source file shows runaway churn; the doc files
dominate, which is expected for a project in its docs-hardening phase.

Evidence: `git log --since="365 days ago" --name-only --pretty=format:`.

### 3) Architecture / Complexity Concerns

> **Status (2026-08): store-first rework COMPLETE.** The concerns below were the
> pre-rework baseline. The multi-instance store-first rework addressed them:
> - **Durability gap (dual-write log-and-continue)** → request write methods
>   (`CreateRequest`, `ClaimRequest`, `SubmitRequestResult`, `CancelRequest`,
>   `AppendRequestChunks`) now persist-first and **surface** persist errors
>   instead of logging them (`server/pkg/service/requests.go`).
> - **`ClaimRequest` not multi-instance safe** → now routes through the guarded
>   `store.ClaimRequest` primitive (`SELECT ... FOR UPDATE` in a serializable tx)
>   so two instances cannot double-claim (`server/pkg/service/requests.go`,
>   `server/pkg/storage/requests.go`).
> - **Lease-expiry double-requeue** → the reaper now uses the guarded
>   `store.ReclaimExpiredRequest` so `attempts` increments exactly once across
>   instances (`markStalledRequests` / `reclaimExpired`).
> - **No external broker** → confirmed (see ARCHITECTURE §8).
> - **Multi-instance active-active** → confirmed intended and proven by
>   `TestActiveActive_*` in `server/pkg/service/request_activeactive_test.go`,
>   added to `make release-gate-runtime`. See `server/docs/reliability-drills.md`
>   drill D7.
>
> The historical analysis is retained below as the pre-rework record.

- **Dual-write in-memory + Postgres with inconsistent failure handling** *(pre-rework record; resolved — see banner above)*.
  `RequestsService` mutates the in-memory `requests` map first and writes
  through to Postgres for durability. A trace of all 30 `s.store.*` call sites
  in `pkg/service/` shows **three different, inconsistent failure patterns**,
  which is the real durability gap (not the uniform "only logs" I first stated):
  1. **Log-and-continue** (most paths): `CreateRequest`, `ClaimRequest`,
     session/tool/task/machine saves — e.g.
     `log.Printf("persist request create failed: %v", err)` (`requests.go:144`,
     also `:412`, `session.go:156`, `task.go:406`, `tool.go:422`, `machine.go:153`).
     The RPC succeeds; in-memory and Postgres silently diverge.
  2. **Silently discard**: `_ = s.store.SaveRequest(ctx, request)`
     (`requests.go:295`, the in-memory `ClaimPendingRequest` requeue branch) —
     the error is not even logged.
  3. **Surface + partial compensation** (one path): `UpdateRequest` returns the
     error to the caller and releases the reserved machine slot
     (`requests.go:342-347`). But the in-memory status mutation already
     happened *before* the save, so even here the local map and Postgres
     diverge on failure.
  Combined with the multi-instance decision above, this strengthens the case
  for the store-first rework: once the store is the source of truth, the
  "write-through after local mutation" pattern and its three failure modes go
  away. `server/docs/reliability-drills.md` implies stronger durability than
  the log-and-continue paths deliver, so the docs should be reconciled with
  that rework.
- **Mixed multi-instance safety between claim paths**. The two claim entry
  points have different consistency models:
  - `ClaimPendingRequest` (the provider poll loop) — when a store is present it
    claims via `store.LeasePendingRequest`, which runs
    `SELECT ... FOR UPDATE SKIP LOCKED LIMIT 1` inside a `SERIALIZABLE`
    transaction (`storage/requests.go:113`, `storage/store.go:96`). This is the
    canonical multi-instance-safe queue pattern, so the **storage layer is
    multi-instance safe** on the Postgres path.
  - `ClaimRequest` (claim by explicit `request_id`) — reads from the
    **in-process** `s.requests[sessionID][requestID]` map first
    (`service/requests.go:383-391`) before persisting. That map is per-process
    and is populated by a load-on-startup + write-through path, not by polling
    the store, so a request created on instance A is **not visible** to
    `ClaimRequest` on instance B until that instance restarts/reloads.
  - Background dispatch, lease-expiry requeue, and retry are in-process
    goroutines operating on the same in-memory map
    (`cleanupStalledRequests`, `requests.go:100`).
  Net: the **Postgres storage primitives are ready for multiple replicas**, but
  the **service layer's in-memory-first state is single-instance by
  construction**. `server/docs/reference-deployment.md:146` confirms the
  reference artifact is "single-instance for clarity" and treats multiple
  replicas as the platform's concern.
- **Multi-instance active-active is implemented** (confirmed with owner, merged
  in PR #5). The service-layer `ClaimRequest`, `CreateRequest`, the lease-expiry
  reaper, capacity check, drain check, and task re-adoption are all store-first
  now: they treat the in-process maps as read caches and the store as the source
  of truth. The storage layer is multi-replica safe (`FOR UPDATE SKIP LOCKED` +
  `SERIALIZABLE`), and the service layer wires it: `ClaimRequest` routes through
  the guarded `store.ClaimRequest`; `MachineLoadInfo` uses
  `store.MachineInFlightCount`; `IsMachineDraining` consults the persisted flag;
  `markStalledRequests` uses `store.ReclaimExpiredRequest`; task re-adoption
  uses `store.ClaimTaskForAdoption` to prevent duplicate execution. Proven by
  `TestActiveActive_*` and the storage contract suite, both in
  `make release-gate-runtime`, plus the two-process conformance case in CI.
- **No external message broker** (confirmed with owner: "queue-backed" always
  meant the internal Postgres-backed queue). A `git grep` for
  kafka/rabbitmq/nats/sqs/redis finds zero references in the server, `go.mod`,
  or storage layer; the single hit is an aspirational note in
  `clients/python-client/toolplane/toolkits/README.md`, not a runtime
  dependency. The "queue" is the pending-request table in Postgres
  (`SELECT ... FOR UPDATE SKIP LOCKED`, `storage/requests.go:113`) plus an
  in-process cache. No external broker is planned.
- **`server.go` size**: the transport adapter is a single struct embedding 5
  unimplemented servers and implementing every RPC. It is large by design (one
  file per domain holds the logic), but any contract change touches it. The
  `.github/instructions` correctly pushes logic out of it — the risk is drift
  over time.

### 4) Security Risks

| Risk | Severity | Evidence | Note |
|------|----------|----------|------|
| `langchain` in `requirements.txt` but not the SDK core | Low–Medium | `requirements.txt:20-21`; `pyproject.toml` has **no** langchain entry; `toolplane/provider_runtime.py` and `__init__.py` import neither `toolkits` nor `langchain` | `langchain`/`langchain-core` are pinned only in the dev/example `requirements.txt` and used solely by `toolplane/toolkits/{standalone_tools,swe}/*` (bundled example tools). The installed SDK package surface (`toolplane/__init__.py`) does **not** expose `toolkits/`. So the core SDK is langchain-free, but `pip install -r requirements.txt` pulls a large transitive graph. **Decision (confirmed with owner): move langchain/langchain-core to an `[project.optional-dependencies]` extra** (e.g. `toolkits`) so `pip install toolplane-python-client` stays lean and the example tools are installed via `pip install "toolplane-python-client[toolkits]"`. |
| `.proto-venv` / `.tmp` / `attached_assets` / `.pytest_cache` are **untracked** | OK | `git ls-files` returns 0 entries for each | Correction of an earlier draft: these were never committed. `.tmp/` (`.gitignore:54`), `attached_assets/` (`:55`), and `.pytest_cache/` (`:62`) are covered by the root `.gitignore`. `.proto-venv` is **not** matched by the root `.gitignore` (its `.venv*/` rule at `:71` matches dirs named `.venv*`, not `.proto-venv`) — it stays untracked because the venv tool wrote its own nested `server/.proto-venv/.gitignore` containing `*`. Optional hardening: add `.proto-venv/` to the root `.gitignore` so protection doesn't depend on a generated nested file. |
| Fixed fixture key in `server/.env.example` | Low (by design) | `TOOLPLANE_AUTH_FIXED_API_KEY=toolplane-conformance-fixture-key` | Intentionally non-secret, dev/CI only; production guards forbid `fixed` mode (`config.go`). |
| Token redaction present | OK | `authorizer.go:220-230` | Tokens are redacted in trace events — good. |
| Production guards present | OK | `config.go:42-72`, `proxy/config.go:26-28` | Production correctly forbids disabled auth, fixed auth, in-memory storage, insecure backend, wildcard CORS. |
| gRPC TLS enforced in prod | OK | `main.go:42-45`; `compose.yaml` mounts certs | Maintained. |

No input-validation gaps were obvious in the scanned hot paths (session/tool
IDs validated; capabilities normalized). A dedicated security audit is
recommended but is out of scope for this mapping.

### 5) Performance / Scalability

- **`ListRequests` in-memory scan**: filters and paginates over a per-session
  slice in memory (`requests.go:194-250`). Fine for the intended workload;
  Postgres-backed queries exist in `storage/requests.go` but the authoritative
  path is in-memory.
- **Serializable transactions**: claim/capacity operations use
  `sql.LevelSerializable` (`store.go:96`) — correct for safety, but under high
  contention this isolation level serializes and can become a bottleneck.
- **Connection pool** fixed at 25 open/idle conns (`store.go:62-63`) — not
  tunable via env today. [TODO] — confirm whether operator tuning is intended.
- **Retained chunk window capped at 100** (`maxRequestStreamChunks`) — bounded
  memory per request; good. No global bound on number of requests is visible.

### 6) Operational / Tooling Debt

- `attached_assets/`, `.tmp/`, `server/.proto-venv`, and `.pytest_cache/` are
  all **gitignored and untracked** (`git ls-files` returns 0; `.gitignore:54,55,62,71`).
  The stray logs under `attached_assets/` (`ts_conformance_full.log`, `*.exit`)
  are local-only debug output, not committed. No repo cleanup required; safe to
  delete locally if desired.
- `clients/python-client/Dockerfile` is **stale and unreferenced**. Verified:
  it `COPY`s `dist/toolplane-0.1.0-py3-none-any.whl`, but the only wheel
  actually built in `dist/` is `mrpc_python_client-1.0.0-py3-none-any.whl`
  (both the **name** and **version** mismatch). It is **not referenced** by any
  maintained compose/CI/docs file (`server/deploy/reference/compose.yaml` uses
  `server/Dockerfile`; CI workflows build Go, not this image). Its only git
  touch was the "Initial public release" commit. This Dockerfile is dead code;
  recommend deletion or a documented refresh.
- No enforced coverage threshold (see TESTING.md §7).

### 7) Coupling (optional diagram)

```mermaid
graph LR
  Tool["ToolService"]
  Session["SessionsService"]
  Machine["MachinesService"]
  Request["RequestsService"]
  Task["TasksService"]

  Request -- "GetToolByName, FindMachinesWithTool, ReserveMachineSlot" --> Tool
  Request --> Machine
  Task -- "orchestrates requests" --> Request
  Machine -- "SetRequestTracker (bidirectional)" --> Request
  Session -- "AuthenticateAPIKey (auth dependency)" -.-> Request
```

The `Machine ↔ Request` bidirectional dependency (`machineService.SetRequestTracker`)
is the one notable coupling; it exists to let drain observe in-flight request
state. It is manageable but is the first place to look if either service's
initialization order changes.

### 8) Evidence

- `server/pkg/service/requests.go:141-147` (persist failure logged only),
  `requests.go:100` (cleanup goroutine), `requests.go:194-250` (list scan)
- `server/pkg/storage/store.go:62-63,96` (pool + serializable)
- `server/cmd/server/config.go:42-72`, `server/cmd/proxy/config.go:26-28` (prod guards)
- `clients/python-client/requirements.txt:20-21`, `clients/python-client/Dockerfile`
- `attached_assets/` listing
- `server/pkg/service/machine.go` (`SetRequestTracker`), `server/cmd/server/main.go:86-88` (init order)
