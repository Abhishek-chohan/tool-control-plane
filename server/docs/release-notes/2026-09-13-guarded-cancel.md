# 2026-09-13 — Guarded cancel and insert-only create

## Summary

Request cancellation, task persistence, and request creation are now
guarded writes at the storage layer. The common failure shape they share:
a service path read the per-instance cache, decided from that possibly
stale copy, and persisted the decision through an unconditional upsert —
so under active-active replicas (or any concurrent writer) the stale
decision rewrote authoritative state.

## Changes

- `Storer.CancelRequestFenced(session_id, request_id, message)` — new
  guarded primitive on both stores. It loads the request row under
  `FOR UPDATE` inside a serializable transaction: terminal requests are
  returned untouched with `cancelled=false` (the caller reports the
  authoritative state), non-terminal requests get the rejection result,
  `dead_letter=true`, and the lease release, persisted in the same
  transaction. A cancel racing `SubmitRequestResultFenced` serializes on
  the row; exactly one of the two writes the terminal state.
- `RequestsService.CancelRequest` is store-first when a store is
  attached and mirrors the winner into the local cache. The no-store
  dev/test path keeps the cache-only behavior under the cache lock.
- `Storer.InsertRequest` — request creation is now insert-only. A
  colliding insert (same id, or the idempotency-key partial unique
  index) fails instead of silently overwriting the persisted row; the
  existing idempotent-retry handling on `CreateRequest` is unchanged.
- `SaveTask` gained a terminal guard on both stores: an upsert against a
  task already in `completed`/`failed`/`cancelled` is dropped. Late
  cache snapshots (e.g. `updateTaskWithError` firing after a
  `CancelTask` landed) can no longer rewrite a terminal task outcome.

## Tests

- Store-parity (`pkg/storage/requests_fenced_test.go`, both backends):
  cancel of a pending request persists `failed` + `dead_letter` and
  releases the lease; cancel of a done request returns `cancelled=false`
  with the result intact; missing request returns `ErrNotFound`;
  `InsertRequest` on an existing id fails.
- Active-active (`pkg/service/request_activeactive_test.go`):
  `TestActiveActive_CancelCannotClobberTerminalResult` — instance A
  completes a claimed request, instance B (whose cache never saw the
  completion) cancels; the cancel is refused with
  `ErrRequestNotCancellable` and the delivered result survives in the
  store.

## Limits

- `CancelTask` still validates cancellability from the instance cache;
  the `SaveTask` terminal guard makes a stale cancel persist inert
  against the authoritative row, but the task service does not yet read
  back before cancelling. The in-memory (no-store) service path is
  single-instance by construction and unchanged.
