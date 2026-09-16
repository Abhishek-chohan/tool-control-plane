# Typed errors and first-class cancellation

Date: 2026-09-15

Request cancellation is a first-class terminal status instead of FAILED
plus a conventioned error string, and failures carry stable
machine-readable reasons on the wire.

## CANCELLED status

- `RequestStatus` gains `REQUEST_STATUS_CANCELLED = 6` (model:
  `"cancelled"`). `CancelRequestFenced` sets it; the MCP gateway maps it
  to the cancelled task status, keeping the legacy error-string check as
  a fallback for one release; `IsTerminalStatus` covers it everywhere
  (claims, retention, waiters, slot release).
- SDKs: Python raises `ToolplaneCancelledError` when a waited request
  ends cancelled; TypeScript maps it to the new `CancelledError` (also
  replacing the previous mis-classification of the gRPC CANCELLED
  transport status as a connection error); the Go client maps it to
  `codes.Canceled` in the typed error.

## Machine-readable reasons

`statusFromDomainError` attaches `google.rpc.ErrorInfo` details (domain
`toolplane`) so clients branch on semantics instead of parsing messages:

| Condition | gRPC code | Reason |
| --- | --- | --- |
| Request timeout above the server maximum | OUT_OF_RANGE | `TIMEOUT_ABOVE_MAX` |
| Stream replay before the retained window | OUT_OF_RANGE | `REPLAY_WINDOW_EXPIRED` |
| Machine concurrency cap reached | RESOURCE_EXHAUSTED | `CAPACITY_EXHAUSTED` |
| Per-session pending backlog full | RESOURCE_EXHAUSTED | `SESSION_BACKLOG_FULL` |

Capacity rejections also carry `google.rpc.RetryInfo` with a suggested
backoff. The Go client surfaces the reason as `Error.Reason`, decoded
from the ErrorInfo detail.
