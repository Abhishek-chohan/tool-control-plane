# Toolplane Server Documentation

## Purpose

This document explains the current server behavior that is implemented in code today.

- The canonical external contract is `server/proto/service.proto`.
- `server/pkg/service/server.go` is the gRPC and HTTP-gateway adapter layer.
- Runtime behavior lives in the service-owned files under `server/pkg/service/`.
- Request and task runtime state is richer than the public proto messages; this document calls out both the public contract and the internal behavior.

## Boundary

What this is:

- The runtime documentation for Toolplane's distributed tool-execution control plane.
- The place to understand session lifecycle, provider registration, request execution, streaming or recovery, machine drain, and task orchestration.

What this is not:

- Not a generic RPC tutorial server.
- Not the place to infer equal support across every SDK; use `SDK_MAP.md` for that.
- Not an MCP protocol specification. MCP is a comparison point and possible adapter target, not the core protocol documented here.

## Release Gate

The authoritative end-to-end release scenario is documented in `server/docs/release-gate.md`. It proves secure startup, provider registration, session creation, unary and streaming execution, request inspection, and graceful drain. Run it locally with:

```bash
cd server && make release-gate
```

See `server/docs/local-development.md` for the development env contract and `.github/workflows/release-gate.yml` for the CI workflow.

The maintained operator-facing observability contract is summarized separately in `server/docs/observability.md` so signals, identifiers, redaction rules, and playbooks can be reviewed without reading the runtime implementation files directly.

## Production Baseline

Production mode is now an explicit runtime contract, not an implied deployment preference.

- `TOOLPLANE_ENV_MODE=production` requires a maintained production auth path. In practice that means `TOOLPLANE_AUTH_MODE=postgres`; disabled auth and fixed fixture auth are rejected at startup.
- `TOOLPLANE_ENV_MODE=production` now also requires durable Postgres-backed storage. In-memory mode remains a development or test choice only.
- The HTTP gateway in production requires explicit `TOOLPLANE_PROXY_ALLOWED_ORIGINS` and rejects insecure backend dialing.
- Proxy guardrails such as per-API-key and per-IP throttling are configured explicitly through `toolplane-gateway` flags like `--api-rate`, `--api-burst`, `--ip-rate`, and `--ip-burst`.

## Durable Storage And Recovery Contract

When a `storage.Store` is configured, the server treats Postgres as the durable recovery source for the following state families:

- sessions, API keys, and user-to-session mappings
- machine registrations and tool ownership
- requests, including lease metadata, retry state, dead-letter state, and retained stream windows
- tasks, including retry state, timeout state, terminal result, and dead-letter markers

The current recovery behavior is:

- Store-backed services load persisted sessions, machines, tools, requests, tasks, API keys, and user-session mappings during startup.
- Persisted request leases are the authoritative expiry source across restarts. The store-backed expired-request scan drives requeue or dead-letter behavior after a restart-like reload.
- Retained streaming chunk windows live on the request record. A restarted server can continue to serve retained replay windows from persisted request state.
- `TOOLPLANE_STORAGE_MODE=memory` is intentionally non-durable. It is valid for local development and CI fixtures, but restart recovery guarantees do not apply.

Some maintenance paths still log persistence failures instead of surfacing them directly to clients. Operators should treat those log lines as durability degradation and use the release gate plus the runtime diagnostics below to catch them before release.

## Durable Execution Contract

This section is the canonical summary of the durability guarantees supported by the repo today. It answers four questions directly: what a request lease means, what survives disconnect or restart, what stream resume guarantees, and what machine drain guarantees.

### Supported Guarantees

- **Request leases are explicit and time-bounded.** A request moves into `claimed` or `running` with a machine-owned lease. If that lease expires, the server emits `request_lease_expired`, releases any reserved machine slot, and either requeues the request with linear backoff or dead-letters it after `MaxAttempts`.
- **Postgres-backed restart recovery reuses persisted lease state.** When a `storage.Store` is attached, request lease metadata survives restart-like reloads. Expired persisted leases are scanned after startup and follow the same requeue-or-dead-letter path as in-memory leases.
- **Consumer disconnect does not cancel work.** A broken `StreamExecuteTool()` or `ResumeStream()` client connection leaves the request record intact. Callers can inspect final request state later and may resume from the retained chunk window if the missing sequence range is still retained.
- **Resume is retained-window replay, not durable full-history replay.** `GetRequestChunks()` and `ResumeStream()` expose only the retained chunk window. Resuming from within that window replays ordered remaining chunks plus the terminal marker. Resuming before the retained window fails with `OUT_OF_RANGE` rather than silently skipping missing data.
- **Machine drain blocks new work immediately and lets in-flight work resolve normally.** `DrainMachine()` removes tool ownership and marks the machine draining before unregister. New request creation and new claims stop targeting that machine immediately. Claimed or running requests already assigned to the machine may finish or age out through the normal lease-expiry path. The machine unregisters only after no active `claimed` or `running` requests remain.
- **Pending requests are not migrated automatically during drain.** If a request was never claimed, it remains pending until another machine registers the same tool or some other control path resolves it.

### Proof Matrix

| Guarantee | Automated proof | Current limit |
| --- | --- | --- |
| Lease expiry requeues or dead-letters claimed/running work and frees machine load | `server/pkg/service/request_runtime_test.go`, `server/pkg/service/request_persistence_test.go`, `server/pkg/service/machine_test.go` | Retries stop after `MaxAttempts`; in-memory mode still has no restart guarantee |
| Consumer reconnect can inspect state and replay the retained chunk window deterministically | `server/pkg/service/request_stream_test.go`, `conformance/cases/request_recovery_chunk_window.json`, `conformance/cases/request_recovery_resume.json`, `conformance/cases/request_recovery_resume_trimmed_window.json`, `conformance/cases/request_recovery_expired_window.json` | The retained window is capped at 100 chunks; there is no durable full-history replay API |
| Graceful drain stops new routing immediately and lets in-flight work finish or age out before unregister | `server/pkg/service/machine_test.go`, `conformance/cases/machine_lifecycle_drain_under_load.json`, `conformance/cases/provider_runtime_drain_under_load.json` | Pending requests are not automatically reassigned during drain |
| The authoritative release signal covers the maintained durability slice | `cd server && make release-gate`, plus `release-gate-runtime` focused Go tests | The gate remains intentionally narrow and does not directly prove live Postgres-backed auth or every deployment topology |

### Failure-Mode Matrix

| Failure mode | What happens now | Signals and proof | Important limit |
| --- | --- | --- | --- |
| Consumer disconnect during `StreamExecuteTool()` or `ResumeStream()` | The request continues. The caller can inspect request state later and may resume from the retained chunk window. | `request_chunks_appended`, `request_execution_completed` / `request_execution_failed`; `server/pkg/service/request_stream_test.go`; `conformance/cases/request_recovery_resume.json` | Replay is limited to the retained window only |
| Provider disconnect or crash after a request is claimed or running | When the lease expires, the server releases machine capacity and either requeues the request with linear backoff or dead-letters it after retry exhaustion. | `request_lease_expired`, `request_requeued`, `request_dead_lettered`; `server/pkg/service/request_runtime_test.go`; `server/pkg/service/machine_test.go` | Tool ownership may still exist until drain or inactive-machine cleanup completes |
| Server restart with Postgres-backed storage | Persisted request state reloads on startup, and expired persisted leases are requeued or dead-lettered by the same expiry path. | `server/pkg/service/request_persistence_test.go`; `cd server && make release-gate` | `TOOLPLANE_STORAGE_MODE=memory` does not carry this guarantee |
| Machine drain with claimed or running work | New routing and new claims stop immediately. In-flight work either completes normally or ages out through lease expiry. The machine unregisters after the active request count reaches zero. | `machine_drain_started`, `machine_drain_completed`, request lifecycle trace events; `server/pkg/service/machine_test.go`; drain-under-load conformance cases | Pending requests remain pending until another machine exists |
| Retained-window overflow before a caller resumes | `GetRequestChunks()` exposes the current `start_seq` / `next_seq` window. Resume from inside that window replays the retained tail plus the final marker. Resume from before `start_seq` fails with `OUT_OF_RANGE`. | `server/pkg/service/request_stream_test.go`; `conformance/cases/request_recovery_resume_trimmed_window.json`; `conformance/cases/request_recovery_expired_window.json` | The runtime does not promise durable replay of every chunk ever emitted |

The detailed lifecycle sections below expand these guarantees. If a future behavior change is not reflected here and in the tests or fixtures named above, it should not be treated as part of the maintained durability contract.

## Tenant And Policy Boundary

The maintained isolation boundary is session-scoped.

- Session ownership is keyed by `CreatedBy` plus the persisted user-to-session mapping.
- Machines, tools, requests, and tasks are all scoped to a single session ID.
- API keys are session-scoped credentials, not global tenancy credentials.
- Session-owned API keys carry explicit `read`, `execute`, or `admin` capabilities, and server authorization binds them to the owning session or user scope per RPC.
- API-key secrets are returned only from `CreateApiKey`; `ListApiKeys` is metadata-only and exposes `key_preview` plus `capabilities` instead of replaying the secret.
- Python-only admin helpers such as bulk delete, stats, token refresh, and invalidation remain explicitly labeled `admin` scope in `SDK_MAP.md`; they are not implied portable SDK guarantees.
- Proxy throttling provides the current platform guardrails for request volume by API key and client IP. Per-machine load protection remains enforced by the server runtime's machine-capacity tracking.

## Canonical Runtime Flow

The maintained first-touch path follows this lifecycle:

1. **Provider runtime attaches**: the maintained Python provider path uses the explicit `ProviderRuntime` surface to create or attach a session, register a machine, advertise tools, and start heartbeat plus polling.
2. **Session created**: `CreateSession` gives the operator a session scope for tool discovery and request routing.
3. **Request executed**: `CreateRequest` queues work; the provider claims it via `ClaimRequest`, moves it to `running`, and submits a result or streams chunks.
4. **Streaming or recovery**: `StreamExecuteTool` delivers ordered chunks. If the consumer disconnects, `ResumeStream` replays from the retained chunk window (default: 100 chunks).
5. **Drain**: `DrainMachine` stops new request routing to the machine, waits for in-flight work to finish, then unregisters.

The Python `example_client.py` (provider) and `example_user.py` (consumer) in `clients/python-client/` exercise this exact flow. See the sections below for detailed semantics of each stage.

## Provider Lifecycle Contract

The maintained provider lifecycle is now an explicit runtime surface rather than an incidental side effect of generic client startup.

1. Establish a transport client (`Toolplane` or `ToolplaneHTTP`) without implying machine ownership.
2. Create or reuse an explicit provider runtime.
3. Create a provider-owned session or attach the runtime to an existing session.
4. Register a machine and advertise tool metadata for that session.
5. Start heartbeats and the polling loop.
6. Claim work, move requests to `running`, execute the local callable, and submit final results or append stream chunks.
7. Honor cancellation, lease-expiry, and drain semantics owned by the server runtime.
8. Stop the provider runtime so heartbeats and polling terminate cleanly.

Consumer-facing `connect()` and `create_session()` calls do not imply these provider responsibilities. Python currently ships the only maintained provider runtime harness. Go and TypeScript expose narrower direct wrappers for session, tool, machine, request, and task operations, but they do not yet ship a maintained claim-and-submit worker loop.

## Runtime Map

| Layer | File | Responsibility |
| --- | --- | --- |
| Public contract | `server/proto/service.proto` | Canonical gRPC and HTTP-gateway API |
| Transport adapter | `server/pkg/service/server.go` | Converts proto requests to service calls and maps model objects back to proto |
| Tool domain | `server/pkg/service/tool.go` | Tool registration, ownership, lookup, and heartbeats |
| Session domain | `server/pkg/service/session.go` | Sessions, API keys, and user-to-session mapping |
| Machine domain | `server/pkg/service/machine.go` | Machine registration, cleanup, load tracking, and tool ownership cleanup |
| Request domain | `server/pkg/service/requests.go` | Request queueing, claiming, retries, result submission, and chunk retention |
| Task domain | `server/pkg/service/task.go` | Higher-level async orchestration on top of RequestsService |
| Internal models | `server/pkg/model/` | Rich runtime state for tools, sessions, machines, requests, and tasks |

The server can run on pure in-memory maps or with an attached `storage.Store`. That matters for some behaviors, especially request lease expiration and retry handling.

## Public Service Surface

These are the service groups currently exposed from `service.proto`.

| Service | Key RPCs |
| --- | --- |
| `ToolService` | `RegisterTool`, `ListTools`, `GetToolById`, `GetToolByName`, `DeleteTool`, `UpdateToolPing`, `ExecuteTool`, `StreamExecuteTool`, `ResumeStream`, `HealthCheck` |
| `SessionsService` | `CreateSession`, `GetSession`, `ListSessions`, `UpdateSession`, `DeleteSession`, plus API-key lifecycle RPCs (`CreateApiKey`, `ListApiKeys`, `RevokeApiKey`) |
| `SessionsService` (admin scope) | `ListUserSessions`, `BulkDeleteSessions`, `GetSessionStats`, `RefreshSessionToken`, `InvalidateSession` — currently surfaced only in the Python SDK; see `SDK_MAP.md` for cross-SDK support |
| `MachinesService` | `RegisterMachine`, `ListMachines`, `GetMachine`, `UpdateMachinePing`, `UnregisterMachine`, `DrainMachine` |
| `RequestsService` | `CreateRequest`, `GetRequest`, `ListRequests`, `UpdateRequest`, `ClaimRequest`, `CancelRequest`, `SubmitRequestResult`, `AppendRequestChunks`, `GetRequestChunks` |
| `TasksService` | `CreateTask`, `GetTask`, `ListTasks`, `CancelTask` |

## Core Timings And Limits

The current runtime constants come from `server/pkg/service/constants.go` and the request and task models.

| Setting | Current value | Notes |
| --- | --- | --- |
| Machine heartbeat TTL | 5 minutes | Machines older than this are reclaimed as inactive |
| Machine cleanup scan | 5 minutes | Inactive-machine cleanup goroutine interval |
| Max concurrent requests per machine | 4 | Used by request running-state reservation and task scheduling |
| Request lease duration | 30 seconds | Claim visibility window for leased work |
| Request timeout | 45 seconds | Default per-request timeout when `RequestsService.ExecuteRequest()` is called without its own deadline |
| Request dispatch scan | 2 seconds | Cleanup ticker interval for expired or stalled requests |
| Request retry backoff | 5 seconds base | Linear backoff is multiplied by attempt count |
| Request max attempts | 3 | Internal model default |
| Task retry backoff | 5 seconds base | Linear backoff is multiplied by attempt count |
| Task max attempts | 3 | Internal model default |
| Task timeout field | 60 seconds | Default per-attempt deadline stored on the task model and enforced when `TimeoutSeconds > 0` |
| Retained stream chunk window | 100 chunks | `model.Request.AddStreamChunk()` keeps only the newest 100 while tracking absolute start and next sequence numbers |

## Machine Lifecycle

`MachinesService` owns machine registration, heartbeat updates, cleanup of inactive machines, and per-machine in-flight capacity tracking.

### Registration And Ownership

- `RegisterMachine()` creates or updates a machine record, sets its default capacity, and tracks the machine as an active tool owner.
- Embedded tool definitions in `RegisterMachineRequest.tools` are re-registered on every machine registration.
- Existing tool registrations for that machine are cleaned up first, then each new tool is registered through `ToolService.RegisterTool()`.
- Tool ownership is machine-specific. The current lookup path for `FindMachinesWithTool()` resolves the active tool owner and returns that machine, not an arbitrary pool of equivalent providers.

### Heartbeats And Inactive Cleanup

- `UpdateMachinePing()` updates `LastPingAt`.
- A background cleanup goroutine runs every 5 minutes.
- Machines whose `LastPingAt` is older than the 5-minute TTL are reclaimed.
- Reclaiming a machine removes the machine record, clears its capacity counters, and detaches or deletes the tools owned by that machine.

### Load Tracking

- The service tracks per-machine in-flight work in a separate capacity map.
- `ReserveMachineSlot()` is used when a request moves into `running`.
- `ReleaseMachineSlot()` is called when requests complete, fail, cancel, or expire.
- `OrderMachinesByLoad()` and `MachineLoadInfo()` exist for scheduling decisions, but with the current tool-ownership model there is usually only one eligible machine per tool name.

### DrainMachine Semantics

The public proto comment says `DrainMachine` should finish in-flight work and then unregister the machine. The current implementation now follows that contract more closely.

- `GRPCServer.DrainMachine()` delegates to `MachinesService.DrainMachine()` instead of directly unregistering the machine.
- Drain starts by removing the machine's tool ownership and untracking it as an active owner, so new `CreateRequest()`, `ClaimRequest()`, and `ClaimPendingRequest()` flows stop targeting that machine immediately.
- Requests already assigned to that machine in `claimed` or `running` state are treated as in-flight work and are allowed to finish.
- A background drain watcher polls the request state for that machine and only calls `UnregisterMachine()` after no active requests remain.
- The operation remains effectively idempotent for already-missing machines because draining a missing machine returns success and unregister is still tolerant of absent records.
- Pending requests that were never claimed are not migrated automatically; they remain pending until another machine registers the same tool or some other control path resolves them.

## Request Lifecycle

`RequestsService` is the queue and result-tracking layer for concrete tool execution.

### Public Status And Result Values

The public request statuses are:

- `pending`
- `claimed`
- `running`
- `done`
- `failure`
- `stalled`

The public result types are:

- `resolution`
- `rejection`
- `interrupt`
- `streaming`

### Creation, Claiming, And Execution

- `CreateRequest()` verifies that the tool exists and that at least one machine currently owns that tool.
- New requests start as `pending` and are initialized with internal retry and lease fields such as `attempts`, `maxAttempts`, `timeoutSeconds`, `backoffSeconds`, `visibleAt`, and `nextAttemptAt`.
- `ClaimRequest()` and `ClaimPendingRequest()` move a request to `claimed`, stamp lease metadata, and update internal attempt bookkeeping.
- `UpdateRequest(..., running, ...)` reserves one of the machine's four slots. If the machine is already at capacity, the request is returned to `pending`, `nextAttemptAt` is scheduled, and the error is set to `machine at capacity`.
- `RequestsService.ExecuteTool()` is a convenience wrapper that creates a request and then delegates to `ExecuteRequest()`.
- `RequestsService.ExecuteRequest()` is the synchronous helper used by task execution. It claims an existing request for a chosen machine, marks it `running`, then waits for a terminal result, a stalled result, or either the caller deadline or the default 45-second request timeout.

### Result Submission, Cancellation, And Terminal State

- `SubmitRequestResult()` sets the final result and updates the request to `done` or `failure`.
- Rejection results populate `LastError` and leave `DeadLetter=false` unless cancellation or lease-expiration logic explicitly marks dead letter.
- `CancelRequest()` sets a rejection result, stores `Request was cancelled` as `Error` and `LastError`, marks `DeadLetter=true`, and releases the machine slot.
- Completed and failed requests release their reserved machine slot.

### Lease Expiration, Retries, And Dead Letter

The server currently has two different expiration paths depending on whether a persistence store is attached.

#### Store-backed path

- Every 2 seconds, `cleanupStalledRequests()` asks the store for expired leased requests.
- `handleExpiredRequest()` releases the machine slot and then either requeues the request with linear backoff or dead-letters it.
- Requeueing resets the request to `pending`, clears the lease, stores `request lease expired` as the error, and schedules `NextAttemptAt` using linear backoff.
- If the request has already exhausted `MaxAttempts`, `MarkDeadLetter()` sets `DeadLetter=true`, status `failure`, and the result becomes a rejection payload.

#### In-memory-only path

- Without a store, the fallback scanner now looks for expired `claimed` and `running` requests every 2 seconds.
- Expired leases are routed through the same `handleExpiredRequest()` path used by the store-backed implementation.
- That means the in-memory path now releases machine capacity and either requeues with linear backoff or dead-letters using the same request-state transition logic as the persistence-backed path.
- Store-backed mode still remains the authoritative durable expiry source, but the retry and dead-letter semantics are now materially aligned across both modes.

## Runtime Signals

Provider behavior is now observable through server-owned trace events instead of only through SDK-local logs.

- Server traces now expose stable top-level correlation fields: `sessionId`, `machineId`, `toolId`, `requestId`, `taskId`, `event`, and `timestamp`.
- Session and API-key audit tracing emits `session_created`, `session_updated`, `session_deleted`, `api_key_created`, and `api_key_revoked`.
- Auth tracing emits `auth_validated`, `auth_rejected`, and `auth_policy_denied` so operators can distinguish invalid tokens from scope or capability failures.
- Drain emits `machine_drain_started` and `machine_drain_completed`.
- Request lifecycle tracing emits `request_created`, `request_claimed`, `request_execution_started`, `request_execution_completed`, `request_execution_failed`, `request_cancelled`, `request_chunks_appended`, `request_lease_expired`, `request_requeued`, and `request_dead_lettered`.
- Task lifecycle tracing emits `task_created`, `task_execution_started`, `task_retry_scheduled`, `task_execution_completed`, `task_execution_failed`, `task_cancelled`, and `task_dead_lettered`.
- `toolplane-server --metrics-listen` now exposes a Prometheus-style `/metrics` endpoint for queue depth, in-flight load, retry totals, dead-letter totals, active machines, and drain progress.
- The supported bootstrap paths pin that metrics listener to a stable scrape address: `127.0.0.1:9102` for `server/run.sh` and `:9102` for the container entrypoint. The binary default remains `127.0.0.1:0` for ad hoc starts.
- The HTTP gateway keeps `/health` for circuit-breaker and throttle state and emits structured throttle logs with redacted `client_fingerprint` instead of raw client IPs.

These signals live in `server/pkg/trace/tracer.go`, `server/pkg/observability/runtime_metrics.go`, and the service-owned session, machine, request, task, and tool code paths. When the server starts with `--trace-sessions`, trace events are written as structured JSON log lines prefixed with `session_trace`.

The compact operator contract, supported metric names, throttle reasons, and redaction rules are documented in `server/docs/observability.md`.

## Operator Diagnostics

The maintained first-debug path now starts from the supported observability contract in `server/docs/observability.md`.

The shortest version is:

- Startup rejected in production mode: inspect env validation first. Production requires Postgres-backed auth and storage, explicit proxy origins, and a secure backend path.
- Queue or dispatch stall: inspect `toolplane_request_queue_depth`, `toolplane_machine_active`, then `request_created` and `request_claimed` traces.
- Repeated lease expiry or retries: inspect `toolplane_request_requeues_total` and correlate `request_lease_expired` plus `request_requeued` by `requestId`.
- Dead-letter churn: inspect `toolplane_request_dead_letters_total` or `toolplane_task_dead_letters_total`, then follow the matching `requestId` or `taskId` traces.
- Stuck drains: inspect `toolplane_machine_draining`, then look for `machine_drain_started` without `machine_drain_completed` and the in-flight `requestId` values for that machine.
- Proxy throttling or circuit-open symptoms: inspect `/health`, `Retry-After`, and proxy throttle logs keyed by stable throttle `reason` and redacted `client_fingerprint`.

### Streaming, Chunk Retrieval, And ResumeStream

The streaming path is split between `ToolService` streaming RPCs and `RequestsService` chunk storage.

- `StreamExecuteTool()` creates a request, then polls every 200 milliseconds.
- Provider-side chunk updates are appended through `AppendRequestChunks()` and stored in `Request.StreamResults`.
- Each request tracks a bounded retained window using absolute sequence metadata: `StreamStartSeq` marks the first retained chunk and `NextStreamSeq` marks the next chunk sequence to assign.
- `GetRequestChunks()` returns the current retained chunk window plus `start_seq` and `next_seq` metadata.
- `StreamExecuteTool()` and `ResumeStream()` send retained chunks by absolute sequence number and do not clear the retained window after delivery.
- Final completion is sent as a last `ExecuteToolChunk` with `is_final=true`; its sequence number is the request's current `next_seq`, plus either the marshaled final result or the final error.

`ResumeStream()` is a deterministic retained-window resume, not a durable exact replay API.

- The RPC takes only `request_id` and `last_seq`.
- It replays any currently retained chunks whose absolute sequence number is greater than `last_seq`, then keeps polling for new chunks.
- If `last_seq + 1` is older than the earliest retained chunk, `ResumeStream()` fails with `OUT_OF_RANGE` instead of silently skipping missing data.
- Because the retained chunk window is capped at 100 entries, `ResumeStream()` cannot guarantee replay of the full history of a long-lived stream.
- `GetRequestChunks()` has the same limitation, but it exposes `start_seq` and `next_seq` explicitly so callers can reason about the current retained window before reconnecting.

## Task Lifecycle

`TasksService` is a higher-level asynchronous wrapper around request execution.

### Public Task States

The public task statuses are:

- `pending`
- `running`
- `completed`
- `failed`
- `cancelled`

### CreateTask And Scheduling

- `CreateTask()` stores the new task immediately and starts `executeTask()` in a goroutine.
- The task model includes internal scheduling fields: `Attempts`, `MaxAttempts`, `BackoffSeconds`, `TimeoutSeconds`, `NextAttemptAt`, `DeadLetter`, and `LastError`.
- Session scoping is enforced on `GetTask()` and `CancelTask()` through `GetTaskByID(sessionID, taskID)` and `CancelTask(sessionID, taskID)`.

### How A Task Attempt Runs

For each attempt, the task runner does the following:

1. Increment `Attempts` and set the task status to `running`.
2. Select a machine using the current tool owner plus machine load and capacity checks.
3. Create the underlying request, store its request ID in the task execution state, and call `RequestsService.ExecuteRequest()` with a task-scoped context.
4. On success, store `result`, `resultType`, `completedAt`, and clear `NextAttemptAt` and `DeadLetter`.

The current selection logic is ownership-first, not broad cluster balancing. `FindMachinesWithTool()` resolves the active machine that owns the tool registration, and only then does `OrderMachinesByLoad()` and `MachineLoadInfo()` decide whether that machine has available capacity.

### Retry And Dead-Letter Behavior

If an attempt fails:

- The task is reset to `pending`.
- `Error` is set from the attempt failure.
- `Result`, `ResultType`, and `CompletedAt` are cleared.
- If `Attempts >= MaxAttempts`, `MarkDeadLetter()` sets `DeadLetter=true`, `LastError`, and a terminal `failed` status.
- Otherwise, `NextAttemptAt` is scheduled using linear backoff: `BackoffSeconds * Attempts`.

With the current defaults, the retry delays are 5 seconds after attempt 1, 10 seconds after attempt 2, and 15 seconds after attempt 3 if another retry were still allowed.

### Cancellation And Timeout Semantics

Task execution now keeps a live execution handle for each running task so task-level control can reach the underlying request path.

- `CancelTask()` still only accepts `pending` and `running` tasks, but it now does more than mutate the task record. It marks the task `cancelled`, clears result and retry state, snapshots the active request ID, calls `CancelRequest()` when a live request exists, and trips the task-scoped cancel function.
- Each live attempt records both the current request ID and a cancel function in an internal execution map, so the task layer can stop in-flight work instead of only marking metadata.
- `runTaskAttempt()` applies `TimeoutSeconds` as a per-attempt deadline with `context.WithTimeout(...)` when the field is greater than zero.
- `RequestsService.ExecuteRequest()` cancels the underlying request whenever the passed context is cancelled or its deadline expires.
- When a task deadline expires, the task ends terminally as `failed`, clears `NextAttemptAt`, and does not schedule further retries.
- Late request completion no longer resurrects cancelled or timed-out work. The task runner re-checks cancellation before writing success, and `updateTaskWithError()` refuses to convert an already-cancelled task into `failed`.

### Shutdown Behavior

- The task runner checks the service context before starting a new attempt and before sleeping for retry backoff.
- If shutdown is detected at those points, the task is marked `failed` with `task cancelled due to server shutdown`.
- Because each live attempt now passes a task-scoped context into `RequestsService.ExecuteRequest()`, shutdown during an in-flight attempt also cancels the underlying request path instead of waiting for it to finish independently.

## Public Contract Versus Internal Model State

The internal request and task models expose more state than the current proto messages do.

| Entity | Public proto fields | Internal-only fields currently used by runtime |
| --- | --- | --- |
| `Machine` | ID, session, SDK version/language, IP, created/ping timestamps | Capacity and in-flight load are tracked separately in `MachinesService` |
| `Request` | ID, session, tool, status, input, result, result type, error, timestamps, executing machine, `stream_results` | `Attempts`, `MaxAttempts`, `TimeoutSeconds`, `BackoffSeconds`, `VisibleAt`, `NextAttemptAt`, `LeasedBy`, `LeasedAt`, `LastError`, `DeadLetter` |
| `Task` | ID, session, tool, status, input, result, result type, error, created/updated/completed timestamps | `Attempts`, `MaxAttempts`, `TimeoutSeconds`, `BackoffSeconds`, `NextAttemptAt`, `LastError`, `DeadLetter` |

This means the retry and dead-letter mechanics described above are real runtime behavior even though the current gRPC and HTTP task and request messages do not surface all of that state directly.

## Compatibility Policy

The written compatibility and upgrade policy for the protobuf contract, HTTP gateway behavior, and maintained SDK wrappers now lives in `server/docs/compatibility-policy.md`.

## Operational Flow

The current interaction between sessions, machines, requests, and tasks is:

1. A session is created and one or more machines register themselves plus their tools.
2. Tool ownership ties a tool name to an active machine registration.
3. A client either creates a request directly or creates a task.
4. A task selects the tool's current owning machine, creates an underlying request, and runs it through `RequestsService.ExecuteRequest()` with task-scoped cancellation and timeout control.
5. Providers claim work, append streaming chunks, and submit final results.
6. Requests track leases, retries, chunk retention, and dead-letter behavior; tasks translate that lower-level execution into higher-level async task state.
7. Machine removal and heartbeat expiry still take a machine out of service immediately, while graceful drain removes tool ownership first, blocks new claims, waits for already-leased requests to finish, and only then unregisters the machine.

## Validation Anchors

When this behavior changes, verify the document against the code in these files first:

- `server/proto/service.proto`
- `server/pkg/service/server.go`
- `server/pkg/service/machine.go`
- `server/pkg/service/requests.go`
- `server/pkg/service/task.go`
- `server/pkg/model/request.go`
- `server/pkg/model/task.go`
