# Cancelled requests wake their waiters

Date: 2026-09-16

Third-party cancellation now reaches every waiter synchronously, and the
cache-only cancel path writes the first-class `CANCELLED` terminal status.

- **The waiter bug**: `WaitForRequestTerminal`'s status switch handled
  done/failed/stalled only. A `CancelRequest` from any other party (gRPC,
  MCP `notifications/cancelled`, `tasks/cancel`) woke the waiter, matched
  no case, and the waiter spun back to sleep until its own deadline — then
  the task layer mapped that deadline error to "task timed out after Ns"
  for a request that had been cancelled seconds earlier. The switch now
  checks `model.IsTerminalStatus` first, so the class closes for any
  future terminal state: cancelled returns `ErrRequestCancelled`
  (surfacing as `CANCELLED` on the wire) promptly, and an unknown terminal
  state exits loudly instead of spinning.
- **Task completion**: a task whose underlying request was cancelled by a
  third party now settles as `CANCELLED` with reason "underlying request
  cancelled" — durable, with the cancellation trace event — instead of
  misreporting a timeout or scheduling retries against a dead request.
- **Cache-only cancel parity**: the store path's `CancelRequestFenced`
  already wrote the first-class `CANCELLED` status; the cache-only path
  still rendered the legacy rejection→FAILED shape. It now writes
  `CANCELLED` too, with the same error text. Production deployments (which
  always run with a store, memory or Postgres) were already on the
  first-class path; this closes the last divergence, visible only to
  direct service-level consumers.
- Noise fix in passing: the waiter's departure path (caller deadline)
  durably cancels the request; when the request had already reached a
  terminal state on another replica the cancel refuses with
  `ErrRequestNotCancellable` — that outcome is now treated as success and
  no longer logged as an error.

Notes: SDKs already understand `REQUEST_STATUS_CANCELLED` (the typed
errors work); Python raises `ToolplaneCancelledError` and TypeScript maps
to `CancelledError` for it. No wire-shape change.
