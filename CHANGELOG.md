# Changelog

Notable changes to this repository. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); detailed
release notes live in `server/docs/release-notes/`.

## [Unreleased]

### Changed

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
