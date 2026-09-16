# MCP edge as a durable client

Date: 2026-09-16

`tools/call` now carries the durable knobs instead of dropping them:

- **Retry dedup**: the request idempotency key derives from the JSON-RPC id,
  tool name, and input digest. A client retrying the same call lands on the
  original request (even while it executes) instead of double-executing a
  side-effecting tool. An intentionally different call never collides.
- **Per-call timeout**: `_meta.dev.toolplane/timeout_seconds` (positive
  integer) forwards as the request's absolute execution timeout, so long
  tools are not re-executed by the lease reaper at the default.
- **Cancellation**: `notifications/cancelled` maps to a fenced
  `CancelRequest` through per-caller tracking of JSON-RPC id → request.
  The 202 acknowledgement is unchanged. Tracking is scoped to the
  authenticated caller, bounded (4096 entries, FIFO eviction), and
  per-instance.
- **Rendering**: sync tool results carry the output exactly once — the
  final result text, or retained chunks replayed only when there is no
  final result — with `_meta.chunks` removed and text capped at 256 KiB
  (truncation marker points at tasks/get paging).

Notes: clients reusing a JSON-RPC id for a genuinely new call with
identical arguments will dedup; per the JSON-RPC spec, new calls should
use new ids. Cancellation tracking is per gateway instance behind an LB.
