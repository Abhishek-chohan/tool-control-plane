# Changelog

Notable changes to this repository. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); detailed
release notes live in `server/docs/release-notes/`.

## [Unreleased]

### Changed

- **Error taxonomy completed across all handlers:** the remaining
  handler-local status calls (list/session/machine/task handlers,
  page-token decode) funnel through the shared translation. Narrow
  wire changes: `ListAuditEvents` page-encode failures move
  INVALID_ARGUMENT → INTERNAL (server-side fault, consistent with other
  lists); cancelled contexts keep CANCELED (previously surfaced as
  INTERNAL, e.g. a `DrainMachine` aborted by its caller); `CreateSession`
  collisions keep bare ALREADY_EXISTS with no payload. See
  `server/docs/release-notes/2026-09-20-error-taxonomy-completion.md`.

- **Streaming RPC failures route through the shared error taxonomy:**
  client disconnects mid-stream (`StreamExecuteTool`, `ResumeStream`)
  still report CANCELED and chunk-delivery failures INTERNAL, but via
  domain sentinels (`ErrClientDisconnected`, `ErrStreamSendFailed`,
  `ErrPrincipalRequired`) so codes and messages are consistent across
  handlers. `ResumeStream` no longer coerces every lookup error to
  NOT_FOUND — a genuine miss still returns NOT_FOUND (same shape as the
  cross-session guard, no existence oracle), while other store failures
  keep their real code. See
  `server/docs/release-notes/2026-09-20-error-taxonomy-stream-paths.md`.

### Added

- **Measured capacity envelope** (`server/docs/capacity.md`): hot-path
  costs on both stores from the benchmarks (the durable insert and
  full-window chunk append dominate Postgres costs; reads are
  sub-millisecond), end-to-end agent-turn numbers from the load driver,
  contention signals and operator guidance, and the reproduction
  command for every figure. Linked from the README's documentation
  index; the observability contract cross-references the serialization
  counters to it. See
  `server/docs/release-notes/2026-09-20-capacity-doc.md`.

- **Soak mode and nightly load workflow**: `loadgen --soak` samples
  goroutines and resident memory during a run and asserts bounded
  growth afterwards; `make soak` wraps it (default 30 minutes). A new
  `Load` workflow runs the soak nightly against Postgres (also
  `workflow_dispatch` with duration/shape inputs) and uploads the
  report artifact — load evidence is scheduled, never a per-PR gate.
  See `server/docs/release-notes/2026-09-20-soak-workflow.md`.

- **Reliability drills under load**: `loadgen --drill provider-kill |
  drain-under-backlog | multi-instance-contention` injects the failure
  semantics from the reliability drill matrix (lease-expiry requeue
  after provider death, drain under saturation, cross-replica
  contention) while agent-shaped load runs, with machine-checked
  assertions in the report and an exit code that reflects them. See
  `server/docs/release-notes/2026-09-20-loadgen-drills.md`.

- **`loadgen` load driver**: `make load` (or `server/cmd/loadgen`)
  drives the control plane with agent-shaped workload — synchronous
  invoke bursts, long token streams, discovery churn, provider
  heartbeats — with in-process fake providers speaking the real
  claim/execute loop, and reports per-op latency percentiles, error
  counts, and server metric deltas as JSON. Embedded mode boots its own
  server (memory, or Postgres via `TOOLPLANE_DATABASE_URL`); external
  mode drives any running server. See
  `server/docs/release-notes/2026-09-20-loadgen.md`.

- **Core-server trusted-transport declaration**: `TOOLPLANE_SERVER_TRUSTED_TRANSPORT=1`
  declares an upstream TLS terminator (mesh, sidecar, terminating proxy)
  and is the only production-legal way to boot the core gRPC server
  without certificate files — the server-side counterpart of the
  gateways' `TOOLPLANE_TRUSTED_PROXY`. A new `toolplane_server_tls_enabled`
  gauge makes served plaintext visible to alerts instead of boot logs.
  See `server/docs/release-notes/2026-09-20-server-trusted-transport.md`.

- **`toolplane status` and `toolplane doctor`**: status shows
  reachability, the server's build version, and its resolved storage
  mode; doctor validates configuration (environment contract +
  production gates, without booting), connectivity, version skew, and —
  with `--database-url` — database reachability and schema freshness,
  naming the exact fix for each failure. Additive
  `HealthCheckResponse.storage` field carries the resolved mode. See
  `server/docs/release-notes/2026-09-17-status-and-doctor.md`.

- **`toolplane-provider` console script**: serve, validate, test, and
  inspect tools files — Python modules using a bare `@tool` decorator
  that imports no session and performs no I/O at import. `serve` binds
  the file to a session and runs the provider loop;
  `validate` checks schemas without executing the module; `call` runs a
  tool locally; `schema` prints the generated JSON schema.
  `ProviderRuntime.tool()` now accepts `session_id=None` for file-based
  tools (bound at start; ambiguous setups fail loudly).

- **`toolplane invoke`**: run a tool from the command line —
  fire-and-forget with the request ID, blocking `--wait` (server-side
  long-poll to terminal state), or `--stream` to print chunks as they
  land; `--idempotency-key` dedups retries; `--format json` for scripts.
  Exit codes map the failure's gRPC status (an unreachable server prints
  the `toolplane serve` next step). See
  `server/docs/release-notes/2026-09-17-toolplane-invoke.md`.

- **`SessionsService/ListAuditEvents`**: the durable audit trail becomes
  readable — newest-first, paged via the standard ListPage trailer,
  filterable by session, actor key, or event type, gated to the admin
  capability. Additive per the compatibility policy; stubs ship for all
  three SDKs. See
  `server/docs/release-notes/2026-09-17-list-audit-events.md`.

- **`toolplane serve --config`**: a YAML config file supplies the base
  configuration layer — flags override environment variables, which
  override the file, which override defaults. Unknown keys are a boot
  failure naming the key; values expand `${VAR}`;
  `storage.database_url_file` reads a secret from a mounted file; a
  world-readable config file logs a warning; and `--dry-run` prints the
  resolved configuration with per-key provenance (`(flag)`/`(env)`/
  `(file)`/`(default)`), secrets masked. See
  `server/docs/release-notes/2026-09-17-serve-config-file.md`.

- **HealthCheck reports the real build version**: the `version` field on
  `HealthCheckResponse` returned a hardcoded `"1.0.0"` placeholder; it
  now carries the ldflags-injected build identity (`git describe`, `dev`
  for source builds), and `toolplane --version` prints the same
  identity — client/server skew is visible by comparing two commands.
  No wire change (the field already existed). See
  `server/docs/release-notes/2026-09-17-build-version-on-health.md`.

- **`toolplane` unified command**: `toolplane serve` runs the gRPC
  control plane with the same flags, `TOOLPLANE_*` environment contract,
  dev-posture banner, and production gates as the standalone
  `toolplane-server` binary (`--version` reports the build; generated
  shell completions via `toolplane completion ...`). Client verbs exit
  with a stable code mapped from the failure's gRPC status (2 invalid
  input, 3 not found, 4 denied, 5 precondition, 6 exhausted,
  7 unavailable, 8 deadline, 9 cancelled), pinned by tests. The
  standalone binaries are unchanged. See
  `server/docs/release-notes/2026-09-17-unified-cli.md`.

### Changed

- **Task results as durable JSON**: completed tasks record and serve their
  result as the JSON encoding of the tool's return value — identical bytes
  to the synchronous InvokeTool long-poll — instead of Go's `%v` rendering,
  which corrupted structured results in `tasks.result` (`{"a":1}` landed as
  `map[a:1]`). String results now carry their JSON encoding, matching what
  provider SDKs submit. See
  `server/docs/release-notes/2026-09-16-task-result-json.md`.

- **Failure fidelity end to end (SWE toolkit)**: toolkit wrappers raise on
  failure — including nonzero command exits — instead of returning error
  text. Provider submissions now record FAILED with the failure message,
  and MCP clients see `isError: true`, so retry heuristics and eval
  scoring read the failure from the status rather than parsing output.
  See `server/docs/release-notes/2026-09-16-tool-failure-fidelity.md`.

- **MCP edge as a durable client**: tools/call derives a request
  idempotency key from the JSON-RPC id (identical retries dedup instead
  of double-executing), `_meta.dev.toolplane/timeout_seconds` forwards a
  per-call execution timeout, `notifications/cancelled` fences the
  executing request (scoped to the authenticated caller), and sync tool
  results render their content exactly once with a 256 KiB cap. See
  `server/docs/release-notes/2026-09-16-mcp-durable-knobs.md`.

- **Typed errors on the wire**: cancellation is the first-class
  `REQUEST_STATUS_CANCELLED` status (plus a legacy error-string
  fallback), and failures carry stable machine-readable reasons as
  google.rpc.ErrorInfo details (`TIMEOUT_ABOVE_MAX`,
  `REPLAY_WINDOW_EXPIRED`, `CAPACITY_EXHAUSTED`, `SESSION_BACKLOG_FULL`)
  with RetryInfo on capacity rejections. The Go client exposes the
  reason on errors; Python raises `ToolplaneCancelledError` for
  cancelled requests; TypeScript maps cancellation to `CancelledError`.
  See `server/docs/release-notes/2026-09-15-typed-errors-and-cancelled-status.md`.

- **Synchronous InvokeTool**: `ExecuteToolRequest.wait_timeout_seconds`
  makes InvokeTool block server-side until the request is terminal —
  returning result/error for the first time — or return the in-flight
  state on wait expiry without cancelling the work. SDKs send the wait
  and keep local polling only as the in-flight fallback.

### Fixed

- **The demo reaps its provider**: `toolplane demo` leaked the provider
  subprocess on every happy-path run (the teardown only ran in the
  `--drill` branch) — the orphan also held inherited stdout pipes open,
  so a piped demo never finished. The provider is now killed and reaped
  on every exit path, and it writes straight to the terminal fds, which
  also removes a data race between the output copier and the demo's own
  prints. See
  `server/docs/release-notes/2026-09-18-demo-provider-teardown.md`.

- **Tool schemas clear a validity bar at registration**: a non-empty
  schema must be well-formed JSON with an object root — anything else is
  `INVALID_ARGUMENT` on both RegisterTool and machine registration (a
  bad schema in a machine's tool list fails the registration instead of
  being silently skipped). Semantic JSON-Schema validation stays out of
  scope and documented on the proto field. The TypeScript provider
  runtime applies the same check client-side, and the dead
  `GetOpenAITools`/`GetToolsSummary` helpers (zero callers; the former
  double-encoded schemas) are removed. See
  `server/docs/release-notes/2026-09-16-tool-schema-bar.md`.

- **Same-tick listings are deterministic**: `ListRequests` sorted by
  `CreatedAt` alone with an unstable sort over map iteration, so requests
  created in the same clock tick could reshuffle between pages and between
  identical calls. The store now orders by `(created_at, id)` and the
  service sorts with the same tiebreak — pages are reproducible on both
  storage modes. The pagination comments also stop overclaiming: the v1
  token is a reversible offset (not "non-enumerable"), and it addresses the
  live ordering, not a snapshot — pages may skip or repeat under mutation.
  See `server/docs/release-notes/2026-09-16-listing-determinism.md`.

- **Audit events name the acting key**: `audit_events` gains a nullable
  `actor_key_id` — API-key creation/revocation, session deletion/bulk
  deletion, and the session kill switch now record the acting API key
  (CreateApiKey previously attributed the session owner; the kill switch
  recorded no actor at all). System-driven events stay unattributed; the
  append-only trail cannot be backfilled. The audit contract is documented
  in `server/docs/observability.md`, and the dead legacy
  `sessions.api_key` column is dropped (written always-empty, read by
  nothing; the retirement drain went with it). See
  `server/docs/release-notes/2026-09-16-audit-actor-attribution.md`.

- **Serialization retries: jittered, counted, and retryable when
  exhausted**: the SERIALIZABLE transaction wrapper (behind every claim
  and fenced write) retried with deterministic 2/4/8ms backoff, no
  metrics, no logs, and a raw SQLSTATE error on exhaustion that failed
  closed to INTERNAL. Retries now sleep full-jitter draws
  (`toolplane_storage_serialization_retries_total` /
  `..._exhausted_total` on the operator surface), exhaustion logs, and a
  new `storage.ErrSerializationConflict` sentinel maps exhaustion to
  `UNAVAILABLE` — which every SDK already retries — since the work was
  never applied. See
  `server/docs/release-notes/2026-09-16-serialization-retry-observability.md`.

- **Machine re-register reconciles in place**: re-registering no longer
  hard-deletes the machine's tools before re-creating them — consumers
  saw NOT_FOUND mid-re-register and every tool gained a fresh ID. Kept
  names now upsert in place (IDs preserved), dropped names are removed,
  rival-owned names keep their owner, and the registry lock no longer
  spans the tool transactions. A cold-cache replica re-registering an
  existing machine verifies the presented credential against the durable
  row instead of minting a second credential over it; `SaveMachine` is
  first-registration-wins on identity fields (token binds once,
  `created_at` never moves) on both backends, and the legacy pre-token
  bind goes through the CAS. See
  `server/docs/release-notes/2026-09-16-reregister-reconcile.md`.

- **The gateways own their deadline policy**: the HTTP gateway's blanket
  30s unary deadline silently truncated server-side waits (expiry
  converted to a normal in-flight 200 — no error anywhere) and was
  client-overridable in either direction via the `Grpc-Timeout` header.
  Deadlines are now method-scoped — the wait entrypoints
  (`InvokeTool`/`ExecuteTool`) dial with a 1h30m backstop above the
  server's 3600s wait ceiling, everything else keeps the 30s fast-call
  default — and `Grpc-Timeout` is stripped at the edge (which also
  protects streaming replays). The MCP gateway mirrors the scoping.
  Deliberate max-length waits no longer surface as 504s feeding the
  circuit breaker. See
  `server/docs/release-notes/2026-09-16-gateway-deadline-policy.md`.

- **A full replay window is readable everywhere**: the server retains an
  8 MiB chunk window, but both gateways and every SDK default capped
  messages at 4 MiB — stored chunk data was unreadable through any
  mediated path, and the Python client silently swallowed the chunk-read
  failure. The gateways' `max-msg-size` default is now sized from the
  ladder (16 MiB batch + headroom), all three SDKs set explicit receive
  (9 MiB) / send (17 MiB) channel limits, `AppendRequestChunks` rejects
  over-ceiling batch totals with a clear `INVALID_ARGUMENT` (an
  exactly-full 16 MiB batch was always over the wire limit once the
  envelope rode on top — the old "a full max-size batch fits" comment
  was wrong), and Python surfaces chunk-read failures
  (`streamResultsError`) instead of dropping `streamResults`. A new
  `request_recovery_full_window` conformance fixture proves the ladder
  end to end. See
  `server/docs/release-notes/2026-09-16-message-size-ladder.md`.

- **wait_timeout_seconds gets a ceiling (3600s)**: the server-side
  long-poll on ExecuteTool/InvokeTool had no bound — an unclaimed PENDING
  request is immortal, so an over-max wait could pin an RPC forever.
  Values above 3600s (the timeout_seconds ceiling) are now rejected with
  `OUT_OF_RANGE` / `TIMEOUT_ABOVE_MAX` before anything is created. The
  Python SDK clamps its derived wait (`timeout_seconds + 15`) to the
  ceiling; explicit waits pass through and surface the rejection. The
  proto field comment documents the max and the per-ingress transport
  budgets. See
  `server/docs/release-notes/2026-09-16-wait-timeout-ceiling.md`.

- **Cancelled requests wake their waiters**: `WaitForRequestTerminal`
  handled done/failed/stalled only, so a third-party `CancelRequest`
  (gRPC, MCP `notifications/cancelled`, `tasks/cancel`) left the waiter
  spinning to its own deadline and the task recorded "timed out" for
  work that was cancelled. The waiter now checks `IsTerminalStatus`
  first: cancellation returns promptly with `ErrRequestCancelled`
  (`CANCELLED` on the wire), tasks settle `CANCELLED` with "underlying
  request cancelled", and the cache-only cancel path writes the
  first-class `CANCELLED` status like the store path always did. See
  `server/docs/release-notes/2026-09-16-cancelled-waiter-fidelity.md`.

- **RegisterTool error contract**: tool registration is an upsert —
  same-machine re-register and stale-owner takeover return `OK` with the
  tool ID preserved — and the one real conflict (a name owned by another
  live machine) now surfaces as `FAILED_PRECONDITION` naming the owner,
  never `ALREADY_EXISTS`. The old handler catch-all mislabeled *every*
  failure (Postgres outage, deadline) as `ALREADY_EXISTS`; failures now
  keep their real codes, with `DEADLINE_EXCEEDED` added to the taxonomy
  for persistence deadlines. Both storage modes share one sentinel, two
  memory-store divergences from the Postgres claim were fixed, and
  Python's reconnect walk no longer string-match-swallows registration
  failures (conflicts log as warnings). See
  `server/docs/release-notes/2026-09-16-register-tool-error-contract.md`.

- **Backlog cap enforced durably across replicas**: the per-session
  pending-request ceiling (512) is now counted and enforced inside the
  store's insert transaction on both backends — previously the check ran
  only against the per-replica cache, which production store-mode creates
  never populated, so the cap held nowhere but tests. Store-mode creates
  now mirror into the replica cache (with clone discipline; four
  un-cloned mirror sites fixed), and `toolplane_request_queue_depth` is
  sourced from a durable pending count refreshed every 5s instead of the
  cache. See
  `server/docs/release-notes/2026-09-16-backlog-cap-store-enforced.md`.

### Changed (earlier)

- **SDK timeout cliff closed (Python)**: `invoke`/`stream`/`ainvoke`/
  `astream` accept `timeout_seconds`, wired through to the request's
  absolute per-attempt timeout on the wire. `invoke` now returns the
  tool's result value (previously the full status envelope, which
  contradicted the documented contract), typed server errors keep their
  identity instead of being flattened, and a lapsed local wait raises
  `ToolplaneTimeoutError` carrying the request ID and last status. See
  `server/docs/release-notes/2026-09-13-sdk-timeout-cliff.md`.

- **Atomic provider claim**: new `ClaimNextRequest` RPC — one round-trip
  leases the oldest claimable pending request for a machine and tool set,
  replacing the list-then-claim race the Python and TypeScript provider
  runtimes used for polling. An idle queue is `claimed=false`, not an
  error. Both provider runtimes now poll through it. See
  `server/docs/release-notes/2026-09-13-atomic-claim.md`.

- **v1 contract**: the protobuf package is now `api.v1` with every RPC
  under `/api.v1.*`. Request/task statuses are enums
  (`REQUEST_STATUS_DONE`, `TASK_STATUS_COMPLETED`, …), every entity time
  is a `google.protobuf.Timestamp`, and `ListRequests` /
  `ListUserSessions` share one opaque-token pagination contract
  (`page_size` + `page_token`, responses carry a `ListPage` trailer).
  `GetTool` resolves a tool by ID or name in one RPC and `InvokeTool` is
  the v1 invocation name; `ExecuteTool` and the by-ID/by-name lookups
  remain as deprecated aliases. `Session.api_key` is gone — session
  credentials are minted explicitly via `CreateApiKey`. The Python, Go,
  and TypeScript SDKs are migrated and keep their friendly surfaces
  (lowercase status names, text timestamps); SDK pagination parameters
  changed from `offset` to an opaque `page_token`. See
  `server/docs/release-notes/2026-09-12-contract-v1.md`.
- HTTP gateway: resource reads bind to templated POST routes
  (`/api.v1/sessions/{session_id}`, `/api.v1/sessions/{session_id}/requests/{request_id}`,
  `/api.v1/users/{user_id}/sessions`, …) instead of per-method paths;
  status filters on the wire are enum names or numbers.

### Fixed

- **Request read coherence**: `GetRequestByID` /
  `GetRequestByIDAnySession` were cache-first with read-through only on
  a miss, so a replica that had cached a request kept reporting its
  stale state after another replica claimed, completed, or requeued
  it. Reads are now store-first on store-backed replicas and overwrite
  the cache mirror unconditionally; `ExecuteRequest` gained a poll
  fallback so cross-replica completions are observed without a local
  signal. A chunk-table gap raced through the trim now maps to
  `OUT_OF_RANGE` (`RequestStreamExpiredError`) instead of an internal
  error. See
  `server/docs/release-notes/2026-09-13-request-read-coherence.md`.
- **Task execution without self-claim**: tasks claimed their underlying
  request for an advisory machine, holding the lease while they merely
  waited — polling providers never saw the request until lease expiry
  (~30s dead time per attempt, attempts burned without execution, tools
  slower than the remaining budget never completed). Tasks now leave
  the request pending for providers to claim and observe it through the
  new non-claiming wait; task adoption refuses terminal tasks. Storage
  hardening: migration passes serialize behind an advisory lock and
  retry deadlocks; fenced writes retry deadlock victims (`40P01`);
  startup cache hydration gets a 30s bound. See
  `server/docs/release-notes/2026-09-13-task-execution-without-self-claim.md`.
- **ListRequests pagination trailer**: the response never carried its
  `ListPage` — the trailer was built and dropped, so clients had no
  `next_page_token` or `total_size`. The trailer is now attached, the
  full-page heuristic matches `ListUserSessions` exactly (shared helper:
  emit a token only when `offset+returned < total`, so a full final page
  no longer sends clients chasing an empty extra page), and the shared
  conformance fixtures gained a `request_list` pagination case exercised
  over both gRPC and HTTP by the Python and TypeScript harnesses. See
  `server/docs/release-notes/2026-09-13-request-list-pagination.md`.
- **Guarded cancel**: request cancellation now goes through a row-locked
  store primitive (`CancelRequestFenced`) instead of a stale-cache read
  plus blind upsert, so a cancel racing a result submission can no longer
  overwrite a delivered terminal outcome — the cancel is refused with
  `REQUEST_NOT_CANCELLABLE` and the result survives. Task persistence
  gained the same terminal guard: `SaveTask` drops stale snapshots that
  would rewrite a completed/failed/cancelled task, and request creation
  is insert-only (`InsertRequest`) so a colliding insert fails instead of
  rewriting history. See
  `server/docs/release-notes/2026-09-13-guarded-cancel.md`.
- `UpdateRequest` cast the v1 status enum's numeric value into the
  model's string status, storing `"3"` instead of `"running"`: requests
  read back as `UNSPECIFIED` from claim until result submission and were
  invisible to status-filtered listings.
- `GetTool` and `InvokeTool` were missing from the fixed-auth policy
  table and failed closed before the capability check.

### Removed

- Dead Python SDK layers: `toolplane/factories/` (an alternative
  construction API imported by nothing), `toolplane/session/` (superseded by
  `core/session_context.py`), and `MODULAR_ARCHITECTURE.md`. The live module
  graph is documented in `clients/python-client/ARCHITECTURE.md`.
- Demo math helpers from the Go (`Add`/`Subtract`/`Multiply`/`Divide`) and
  TypeScript (`add`/`subtract`/`multiply`/`divide`) clients — they hardcoded
  a fictional server-side calculator into the client surface. Use
  `ExecuteTool`/`executeTool` with your own registered tools.

### Changed

- SDK surface cleanup: Python `ainvoke`/`astream` are now real coroutines
  (previously `ainvoke` was a synchronous submit misusing the "a" prefix and
  `astream` was a plain alias for the blocking `stream`); the facades'
  `get_request_status` argument order is now `(session_id, request_id)`,
  consistent with every other facade method; seven byte-identical toolkit
  files in `toolkits/swe/` became re-export shims of the
  `toolkits/standalone_tools/` originals. See
  `server/docs/release-notes/2026-09-11-sdk-surface-cleanup.md`.
- Quickstart: `make demo` (in `server/`) runs the README "First Offload
  Path" — in-memory server, provider example, consumer example — as one
  verified command.

### Fixed

- `example.py`'s LangChain converter crashed with `AttributeError` on any
  tool whose `args_schema` was a plain dict; it now handles pydantic v2/v1
  models, dicts, and empty values.

### Added

- MCP initialize compatibility: the gateway answers the standard
  `initialize` handshake for the 2025-03-26 / 2025-06-18 / 2025-11-25
  revisions and serves their `tools/list` / `tools/call` with legacy shapes,
  so the installed base of initialize-based MCP clients can connect to the
  2026-07-28 gateway (previously they could not connect at all). Conformance
  now includes the official `@modelcontextprotocol/sdk` client connecting
  over StreamableHTTP through initialize and executing a tool end to end.
  See `server/docs/release-notes/2026-09-11-mcp-interop.md`.

### Added

- Prometheus-native metrics: `prometheus/client_golang` replaces the
  hand-rolled text renderer (same metric names, plus Go runtime and process
  collectors) and adds `toolplane_grpc_requests_total{method,code}` and a
  per-method `toolplane_grpc_request_duration_seconds` histogram recorded by
  interceptors wrapped outermost — auth rejections count like any other
  request. The server previously combined a singular auth interceptor with
  a chained metrics one, silently dropping the chain; it now builds one
  explicit chain. See
  `server/docs/release-notes/2026-09-11-observability-lifecycle.md`.
- Durable audit trail: session/API-key/machine lifecycle events and
  request/task dead-letters persist to a new `audit_events` table
  (best-effort — failures log with event identity and never block the
  operation).
- `grpc.health.v1`: the gRPC server reports SERVING once listening and
  NOT_SERVING on shutdown, so load balancers probe the standard service.
- `TOOLPLANE_LOG_FORMAT=json` switches the server to structured `slog`
  JSON records; a std-log bridge routes the services' existing `log.Printf`
  output through the same handler. `traceparent` is forwarded by the JSON
  proxy and surfaced on RPC log lines for trace correlation.

### Fixed

- Lifecycle: GracefulStop is bounded by a 20s deadline (a stuck stream no
  longer hangs shutdown forever); the metrics server reports its outcome on
  a channel instead of `log.Fatalf` in a goroutine (which skipped the
  deferred storage close); and the proxy and MCP gateway handle SIGTERM
  with a bounded 10s drain instead of dying mid-stream, plus idle-connection
  timeouts.

### Changed

- Chunk storage redesign: stream chunks moved out of the `requests` row's
  JSONB array (rewritten whole on every append) into an append-only
  `request_chunks` table. Fenced appends insert chunk rows and update only
  the request row's sequence bookkeeping, trimming the table to the newest
  100 chunks / 8 MiB payload; chunk windows are served from the table, so
  replica B sees replica A's chunks without cache mirroring. Worst-case
  retained window per request drops from ~400 MiB to ~8 MiB. Chunks above
  512 KiB are rejected with `INVALID_ARGUMENT`. See
  `server/docs/release-notes/2026-09-11-chunk-storage.md`.
- `StreamExecuteTool` and `ResumeStream` deliver chunks on append signal
  (microsecond latency) instead of a 200 ms poll, with a 2 s fallback
  timer; the request signal map now releases entries on terminal
  transitions instead of leaking one per request.
- The gRPC server sets explicit `MaxRecvMsgSize`/`MaxSendMsgSize` (16 MiB,
  sized for a full chunk batch) and a keepalive enforcement policy
  (MinTime 5s — below the proxy/gateway 10s ping cadence; stricter values
  kill proxy connections via GOAWAY).
- Task execution no longer self-assigns. Tasks are enqueued pending and a
  fenced adoption model runs them: a `tasks.adopted_by` column records the
  owning instance, the executing instance renews its lease every 10s and
  cancels itself if ownership is lost (no more double-execution when an
  instance restarts under a long-running task), and an adoption sweep on
  every instance claims due tasks nobody owns — recovering a dead
  instance's work within ~35s instead of at the next restart, and spreading
  retries across replicas. Retries release ownership after scheduling so
  any instance can pick them up. `persistTask` retries and returns its
  error, with call sites logging the concrete consequence of an undurable
  write; a periodic sweeper prunes terminal tasks older than 24h, and
  `CleanupOldTasks` now removes only terminal tasks (previously it deleted
  any old row, including pending/running work). See
  `server/docs/release-notes/2026-09-11-tasks-redesign.md`.

### Added

- Idempotency keys on creates: `CreateRequest`, `ExecuteTool`,
  `StreamExecuteTool`, and `CreateTask` accept an optional `idempotency_key`
  that dedups within a session — a retry returns the original request (or
  task, without re-executing it) instead of enqueueing duplicate work.
  Backed by `requests`/`tasks.idempotency_key` columns with partial unique
  indexes; a cross-replica race is settled by the index and the loser
  re-reads the winner's row. All three SDKs expose the key, and the Python
  `stream()` fallback re-invokes under one so a broken stream can no longer
  double-execute a tool. See
  `server/docs/release-notes/2026-09-10-idempotency-and-resume.md`.
- `ResumeStream` wrappers in all three SDKs (Python gRPC + HTTP
  `resume_stream()`, Go `ResumeStream()`, TypeScript `resumeStream()`):
  replay retained chunks after `last_seq` and stream live until the final
  marker. Conformance case `request_idempotency_retry` (grpc + http).
- Go client: `WithExecutionTimeout(d)` replaces the hardcoded 30s
  ExecuteTool/stream ceiling; caller deadlines still win.

### Fixed

- The Python `stream()` fallback re-invoked the tool on any stream failure —
  including a completed-but-failed tool — executing it twice. It now resumes
  from the last received sequence via `ResumeStream`, or re-invokes under an
  idempotency key when the stream died before the first chunk; tool failures
  surface instead of re-executing.
- The request reaper mutated cached requests after releasing the read lock
  while drain polls counted them under it (found by the re-enabled race
  detector); the mutation now holds the write lock.

### Changed

- Multi-instance coherence: with a shared store configured, the store is now
  the point of coherence for the paths that previously decided from
  per-replica memory. `ListRequests` reads through the store (requests from
  other replicas appear; new `ListRequestsBySession`); `CreateSession` dedups
  via `InsertSessionIfAbsent` so a cross-replica race yields one winner and
  an `ALREADY_EXISTS` instead of a silent overwrite; machine capacity
  consults `MachineInFlightCount` so the per-machine limit holds across
  replicas; `DrainMachine` persists the drain flag (and re-registration
  clears a stale one); cached API-key grants re-verify against the store
  (new `GetAPIKeyByHash`) so revocations and session deletions propagate
  within ~5s on every replica; `withSerializableTx` retries SQLSTATE 40001
  serialization failures with backoff. See
  `server/docs/release-notes/2026-09-09-multi-instance-coherence.md`.
- The service package no longer hands out live cached `Request`/`Task`
  pointers: read APIs return clones, store-fresh objects are cached as
  clones, and nothing mutates shared state under a read lock. The CI race
  leg covers every server package again (`make test-race`, no exclusions).

### Changed (earlier)

- Domain-error taxonomy: server handlers map failures to gRPC codes through
  one table (`statusFromDomainError`) instead of inline choices that
  collapsed distinct conditions into `NOT_FOUND`/`INTERNAL`. Claim
  contention, draining machines, terminal-state cancels, and missing
  providers are now `FAILED_PRECONDITION`; capacity requeues are
  `RESOURCE_EXHAUSTED`; over-maximum timeouts are `OUT_OF_RANGE`; fenced
  storage misses are `NOT_FOUND`. See
  `server/docs/release-notes/2026-09-07-typed-errors.md`.
- Typed SDK errors in all three clients: Python
  (`ToolplaneAPIError` + per-code subclasses with
  `code`/`retryable`/`request_id`/`status`, shared by the gRPC and HTTP
  transports), Go (`client.Error` with `errors.Is` sentinels and
  `status.Code` passthrough), and TypeScript (`APIError` + per-code
  subclasses). `retryable` is true only for `UNAVAILABLE` and
  `RESOURCE_EXHAUSTED`.
- Python SDK logs through `logging.getLogger("toolplane.…")` instead of
  writing to stdout; the provider poll loop logs claim contention at debug
  and unexpected claim failures at warning instead of silently passing.
- Python gRPC connection resets are restricted to `UNAVAILABLE`: a bad or
  revoked API key no longer triggers reconnect + re-registration churn.
- Python HTTP transport retries only transport failures and retryable
  typed errors — not deterministic 4xx rejections.

### Fixed

- Python `delete_tool` sent an empty `machine_id` (read the normalized tool
  dict under `machineId` instead of `machine_id`), which the per-machine
  credential gate rejects.

### Security

- Capability split: `invoke` (consumer) and `provide` (provider) replace the
  overloaded `execute` in the authz policy table; `execute` remains a legacy
  alias that expands to invoke+provide, so existing keys and rows keep
  working.
- Per-machine identity: `RegisterMachine` mints a once-only machine
  credential (`Machine.machine_token`); provide-scoped RPCs must present it
  (x-toolplane-machine-token metadata / X-Toolplane-Machine-Token header) in
  session-key mode, and re-registering an existing machine ID requires it —
  machine-ID takeover is closed. Python and TypeScript SDKs capture and send
  the credential automatically.
- `CreateApiKey` requires an explicit capability list (`INVALID_ARGUMENT` on
  empty); the implicit read+execute+admin default is gone, and all SDK
  wrappers take a required non-empty list.
- `InvalidateSession` is a real session-wide kill switch: revokes every live
  API key of the session and reports the count
  (`InvalidateSessionResponse.revoked_api_keys`).
- API-key auth is an O(1) SHA-256 hash-index lookup (no all-keys scan); new
  key format `toolplane_key_<uuid>` stops embedding the session ID; previews
  show a masked tail only.
- Authorizer helpers fail closed without a principal (auth-disabled dev mode
  gets anonymous interceptors; production still refuses disabled auth).
- `ResumeStream` returns identical `NOT_FOUND` for other sessions' requests
  (cross-session existence oracle closed).
- JSON proxy: no `?api_key=` URL credentials; XFF trusted only behind
  `TOOLPLANE_TRUSTED_PROXY=1`; rate-limiter/throttle keys hashed. Both HTTP
  edges gained TLS flags and refuse production plaintext without a declared
  trusted proxy.
- MCP gateway: rate limiting, hashed TTL-bounded capped session cache, and
  generic client errors with correlation IDs (backend detail logged
  server-side only).
- Toolkits quarantined: bash timeout + optional workspace root instead of the
  cosmetic blocklist, per-user editor state file, no `os.chdir`, loud
  UNSANDBOXED warnings, `example.py` defaults to a safe echo demo (SWE behind
  `--toolkit swe`).

### Removed

- `SessionsService.RefreshSessionToken` RPC (fabricated a token that
  authenticated nothing) and its Python wrappers; rotation is create+revoke.
  See `server/docs/rpc-retirement.md`.

### Added

- Conformance case `multi_instance_list_visibility` (grpc, two real server
  processes sharing one Postgres): a request created on instance A appears
  in instance B's `ListRequests` listing through the store read-through.
- Go multi-replica suites: `pkg/service/coherence_test.go` (revocation and
  session-deletion propagation, session dedup, list read-through,
  store-backed capacity, drain-flag visibility) and
  `pkg/storage/serialization_test.go` (SQLSTATE 40001 classifier).
- Conformance integrity guarantees: bootstrap failures are hard failures
  (never skips) in auto-boot mode, connectivity-skip conversion is opt-in via
  `TOOLPLANE_CONFORMANCE_ALLOW_SKIP`, and a session guard fails runs where an
  expected transport executed tests but passed nothing.
- CI: Python SDK unit suites (`python-unit` job / `make python-unit`) and the
  Go race detector (`make test-race`; CI runs every package except
  `pkg/service` under `-race` pending the clone-at-the-boundary fix).
- `gosec` and `staticcheck` enabled in `.golangci.yml`; advisory (non-blocking)
  mypy step in the Python lint job.
- HTTP `ReadHeaderTimeout`/`ReadTimeout` on the JSON proxy, MCP gateway, and
  metrics server (Slowloris hardening).

- `RequestsService.RenewRequestLease` RPC: providers renew the execution
  lease of claimed/running requests so long-running tools are not reclaimed
  mid-flight. Renewal never extends a lease past the request's absolute
  timeout. (`lease_epoch`, `leased_by`, `lease_expires_at`, `timeout_seconds`
  are now exposed on the `Request` message.)
- Per-request `timeout_seconds` override on `CreateRequest` and `ExecuteTool`
  (server default 45s, maximum 3600s).
- Lease fencing on provider writes: `UpdateRequest`, `SubmitRequestResult`,
  and `AppendRequestChunks` require the current lease grant
  (`machine_id` + `lease_epoch` from the claim response) and reject stale or
  forged writers with `FAILED_PRECONDITION`.
- Python and TypeScript provider runtimes renew in-flight leases
  automatically and present the lease grant on every fenced write.
- Conformance case `provider_runtime_fenced_submission` (grpc + http, Python
  and TypeScript runners); Go lease/fencing suites wired into
  `release-gate-runtime`.

### Changed

- Reference deployment compose refuses to start without an explicit
  `POSTGRES_PASSWORD`, `TOOLPLANE_DATABASE_URL`, and (for the bootstrap
  profile) `TOOLPLANE_BOOTSTRAP_FIXED_API_KEY`; Postgres and metrics ports are
  bound to 127.0.0.1 by default. `.env.example` is now fill-in-the-blank with
  no working default credentials.
- Python client `requirements.txt` trimmed to actual runtime + dev
  dependencies; toolkit-only deps (langchain, chardet, google-api-python-client,
  ipython) moved out — install `toolplane/toolkits/swe/requirements.txt` when
  using the toolkits. Stub `uv.lock` removed; README install paths and
  pyproject URLs fixed.
- The lease reaper reclaims on the unrenewed lease deadline (30s TTL) or the
  absolute per-attempt timeout, and scans by `visible_at` instead of
  `leased_at`.
- Line endings pinned to LF via `.gitattributes` on all platforms.

### Deprecated

- Calling the fenced provider RPCs without `machine_id`/`lease_epoch` is
  rejected; custom clients must present the claim's lease grant.

See `server/docs/release-notes/2026-09-04-lease-renewal-and-fencing.md` for
migration details.
