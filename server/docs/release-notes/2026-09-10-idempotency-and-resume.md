# Release Note — Idempotency Keys and Stream Resume (2026-09-10)

Branch: `feat/idempotency-and-resume`

## What changed

### Idempotency keys on creates

`CreateRequest`, `ExecuteTool`, `StreamExecuteTool`, and `CreateTask` accept
an optional `idempotency_key`. Within a session, re-creating with a key that
already exists returns the original request (or task) instead of enqueueing
duplicate work — the semantic foundation for safe client retries. An empty
key keeps the old behavior exactly.

- Requests: `requests.idempotency_key` column with a partial unique index on
  `(session_id, idempotency_key)`. `CreateRequest` checks the local cache
  and the store (`Storer.GetRequestByIdempotencyKey`) before creating; when
  two replicas race, the unique index rejects the loser's insert
  (`storage.IsUniqueViolation`, SQLSTATE 23505) and the loser re-reads and
  returns the winner's request.
- Tasks: the same pattern (`tasks.idempotency_key`, partial unique index,
  `Storer.GetTaskByIdempotencyKey`). A retry returns the existing task
  **without scheduling it again** — no duplicate execution, no extra
  attempt.
- Keys live as long as their row: the dedup window is the request/task
  lifetime, not a fixed TTL.

### ResumeStream in the SDKs

The server RPC existed; no client exposed it. All three SDKs now ship
wrappers that replay the retained chunks after `last_seq` and stream live
chunks until the final marker:

- **Python**: `request_manager.resume_stream(session_id, request_id,
  last_seq)` on both transports (gRPC stub call; HTTP `api/ResumeStream`
  with result-frame unwrapping), yielding normalized chunk dicts. Raises
  `ToolplaneInvalidArgumentError` (OUT_OF_RANGE) when the retained window
  has moved past `last_seq`.
- **Go**: `client.ResumeStream(ctx, requestID, lastSeq, onChunk)` returning
  the full in-order chunk slice, typed errors via `client.FromGRPC`.
- **TypeScript**: `resumeStream(requestId, lastSeq, onChunk?)` resolving to
  normalized `ExecuteToolChunkModel[]`.

### The Python `stream()` double-execution fix

`SessionContext.stream()` (gRPC and HTTP) caught every exception from the
direct stream — including the deliberate raise for a tool that completed
with an error — and fell back to `_stream_via_polling`, whose first act was
a fresh `ainvoke()`: a new `CreateRequest`, re-executing the tool. The
fallback is now three-way and side-effect safe:

1. A transport failure mid-stream resumes via `ResumeStream` from the last
   received sequence number.
2. A stream that died before the first chunk re-invokes under the same
   caller-generated idempotency key, so the server returns the request the
   broken stream already created instead of executing the tool again; the
   polling fallback then polls that request.
3. A tool that completed with an error surfaces the failure — it is never
   re-executed.

When resume answers OUT_OF_RANGE (the retained window moved past the
position), the fallback polls the original request's retained window
instead, suppressing chunks the caller already received.

### Configurable Go execution timeout

The Go client hardcoded a 30s ceiling on `ExecuteTool` waits and execution
streams. `WithExecutionTimeout(d)` overrides it; a caller-supplied deadline
always wins, and clients constructed without the option keep the 30s cap.

### Reaper race fix (found by the new CI coverage)

With `pkg/service` under `-race` in CI (B5), the reaper surfaced a real
race: `handleExpiredRequest` mutated the cached request after releasing the
read lock while the drain poller counted active requests under it. The
mutation now happens under the write lock.

## Compatibility

- Proto: `idempotency_key` added to `ExecuteToolRequest` (field 5),
  `CreateRequestRequest` (field 5), and `CreateTaskRequest` (field 4) —
  additive; older clients sending the previous shapes are unaffected.
- Postgres deployments run the new migrations (`idempotency_key` columns +
  partial unique indexes) on startup.
- No existing behavior changes for callers that omit the key.

## Known limits

- The dedup window is the row lifetime: a key minted months ago still
  dedups. Callers wanting a shorter window should embed a window marker in
  the key (e.g. `key-2026-09-10`).
- Idempotency dedups creates; it does not dedup cancel/submit operations
  (those are already fenced by lease grants).

## Verification

- `pkg/service/idempotency_test.go`: key dedup (same key → same request;
  different key → distinct; empty key never dedups), cross-replica dedup
  through the store, and task dedup returning the completed task with the
  attempt count unchanged.
- Conformance case `request_idempotency_retry` (grpc + http, Python and
  TypeScript runners).
- Full server suite under `-race` in memory and Postgres modes (zero data
  races after the reaper fix); Python lint + units + full conformance
  (grpc/http/mcp); TypeScript build, lint, units, conformance 44/44;
  `make check-proto-drift` clean on the committed stubs.
