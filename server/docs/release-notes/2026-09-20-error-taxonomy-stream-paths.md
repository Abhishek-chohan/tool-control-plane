# Streaming RPC failures route through the error taxonomy

Date: 2026-09-20

The streaming handlers (`StreamExecuteTool`, `ResumeStream`) were the
last handlers translating failures with handler-local status calls.
They now funnel through the shared taxonomy like every other handler,
so one condition always surfaces as the same code regardless of which
RPC produced it.

- **Client disconnect mid-stream is `ErrClientDisconnected` → CANCELED**
  (unchanged code). The work keeps running; followers reattach with
  `ResumeStream`.
- **Chunk delivery failure is `ErrStreamSendFailed` → INTERNAL**
  (unchanged code). The failure is between server and client; execution
  and the retained window are unaffected.
- **A handler reached without an authenticated principal is
  `ErrPrincipalRequired` → PERMISSION_DENIED** (unchanged code), still
  failed closed before any state is consulted.
- **`ResumeStream` no longer coerces every lookup error to NOT_FOUND.**
  A genuine miss still returns NOT_FOUND with the same message shape
  the cross-session check returns (no existence oracle). Any other
  store failure now keeps its real code — a persistence deadline
  surfaces as DEADLINE_EXCEEDED instead of a misleading NOT_FOUND. This
  is a rare-path wire change; clients that only branch on codes see a
  difference only when the store is already failing.
- Stream error messages now use the shared `failed to <action>: …`
  shape (e.g. `failed to follow tool stream: client disconnected`).

Notes: the cross-session existence guard is preserved and now pinned by
tests: a session-bound caller gets the same NotFound shape for a missing
request and a request in another session. The CLI exit-code mapping and
SDK typed-error classes are unchanged on the disconnect, send-failure,
and principal paths.
