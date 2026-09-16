# wait_timeout_seconds gets a ceiling: 3600s, rejected above it

Date: 2026-09-16

`wait_timeout_seconds` — the server-side long-poll on `ExecuteTool` /
`InvokeTool` — had no bound at all. An unclaimed PENDING request can be
immortal (the reaper only reclaims leased rows), so an over-max wait could
genuinely never terminate and would pin an RPC goroutine for the duration.

- **The rule**: caller-supplied durations cap at one hour — the same
  ceiling as `timeout_seconds`, since a wait can never usefully outlive
  the request's own maximum execution time. Values above 3600 are rejected
  with `OUT_OF_RANGE` / `TIMEOUT_ABOVE_MAX` before anything is created:
  the `timeout_seconds` precedent verbatim, reason code and test
  infrastructure included. Wait-expiry-returns-in-flight is unchanged —
  the max only bounds the expiry, it does not change semantics.
- **Python SDK**: the derived wait (`timeout_seconds + 15`, else 60) now
  clamps to the ceiling — at `timeout_seconds=3600` the derivation would
  otherwise send 3615 and trip the rejection — while an explicit
  `wait_timeout` passes through untouched: over-max explicit values
  surface the server's loud rejection instead of the SDK silently
  shrinking a caller's chosen budget. The derivation lives in one shared
  helper used by both the gRPC and HTTP transports.
- **Ingress reality** (documented in the proto field comment): direct
  gRPC callers are bounded by this ceiling; the HTTP gateway additionally
  bounds calls at its own transport deadline; the MCP gateway applies its
  sync timeout. The gateway deadline policy itself is a separate change.

Notes: no in-repo SDK default exceeds the ceiling (Python derives
`timeout + 15` — now clamped; TypeScript derives from its own deadline;
the HTTP transport clamps to its request budget). Hand-rolled clients
sending over-max waits now receive `OUT_OF_RANGE` instead of a goroutine
pin — the documented tightening.
