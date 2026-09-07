# Release Note — Typed Errors and Domain-Error Taxonomy (2026-09-07)

Branch: `fix/typed-errors`

## What changed

### Server: one mapping from domain condition to gRPC code

Every request-lifecycle handler previously chose its gRPC code inline, and
several collapsed unrelated conditions into `NOT_FOUND`: `ClaimRequest`
returned it for a draining machine, a lost claim race, a missing request,
and a store outage alike; `CancelRequest` did the same for terminal-state
conflicts; capacity requeues surfaced as `NOT_FOUND` through the fenced-write
helper; a tool with no registered provider failed as `INTERNAL`.

Domain failures now carry sentinels (`server/pkg/service/errors.go`) and a
single function — `statusFromDomainError` — decides the code, so one
condition always surfaces the same way regardless of which RPC produced it:

| Condition | Code |
|---|---|
| Missing session / request / machine / tool / key / task | `NOT_FOUND` |
| Create collision (session, tool name) | `ALREADY_EXISTS` |
| Invalid input (capability lists, empty values) | `INVALID_ARGUMENT` |
| Wrong / missing machine credential | `PERMISSION_DENIED` |
| Machine draining | `FAILED_PRECONDITION` |
| Request not claimable (already claimed / running / terminal) | `FAILED_PRECONDITION` |
| Cancel of a terminal request or task | `FAILED_PRECONDITION` |
| Stale or forged lease grant | `FAILED_PRECONDITION` |
| Write against a terminal request | `FAILED_PRECONDITION` |
| Tool with no registered provider | `FAILED_PRECONDITION` |
| Machine concurrency limit reached (work requeued) | `RESOURCE_EXHAUSTED` |
| `timeout_seconds` above the maximum | `OUT_OF_RANGE` |
| Anything unmapped | `INTERNAL` (fail closed) |

Notes:

- The guarded store claim returns the same `ok=false` for a missing request
  and a lost race; `ClaimRequest` now re-reads on that path to distinguish
  `NOT_FOUND` from `FAILED_PRECONDITION`.
- Fenced storage primitives wrap misses with `storage.ErrNotFound` (new) so
  a missing request no longer surfaces as `INTERNAL` through them.
- `CreateSession`'s "already exists" detection now uses `errors.Is` instead
  of matching on the error string.

### SDKs: typed errors

All three SDKs translate transport failures into typed errors carrying the
status code, the request id (when the wrapper knows it), and a `retryable`
flag. Retryable means exactly `UNAVAILABLE` and `RESOURCE_EXHAUSTED` — the
only codes where re-issuing the call can help. Credentials, arguments,
state conflicts, and missing entities are deterministic and never retried.

- **Python** (`toolplane.core.errors`): `ToolplaneAPIError` base plus
  per-code subclasses (`ToolplaneNotFoundError`,
  `ToolplaneFailedPreconditionError`, `ToolplaneUnauthenticatedError`, …)
  with `code`, `retryable`, `request_id`, `status`. The HTTP transport maps
  gateway error bodies (`{"code": N, "message": ...}` — the numeric gRPC
  code is authoritative) and bare HTTP statuses through the same classes.
- **Go** (`client` package): `client.Error` wraps stub errors with op,
  request id, code, and `Retryable`; it implements `GRPCStatus()` so
  `status.Code` keeps working through the wrapper, `Unwrap()` so the
  original error stays reachable, and `Is()` matching so
  `errors.Is(err, client.ErrNotFound)` works. `client.FromGRPC` wraps;
  `client.IsRetryableCode` answers for a bare code.
- **TypeScript** (`src/errors`): `APIError` base plus per-code subclasses
  (`NotFoundError`, `FailedPreconditionError`, `UnauthenticatedError`, …)
  with `code`, `codeName`, `retryable`, `requestId`, `status`. All unary
  RPCs funnel through one translation point, and wrappers that know the
  request id attach it.

### Client bugs fixed

- The Python gRPC client treated `UNAUTHENTICATED` (and `INTERNAL`,
  `DEADLINE_EXCEEDED`) as channel-level failures, triggering reconnect and
  full machine re-registration on every call with a bad or revoked key.
  Connection resets are now restricted to `UNAVAILABLE`.
- The Python HTTP client retried **every** `POST` on every error class
  (4xx included) because `with_retry` caught `Exception`; it now retries
  only transport failures and the retryable typed errors.
- The Python provider poll loop swallowed all claim errors with
  `except Exception: pass`. Expected contention (`FAILED_PRECONDITION`,
  `NOT_FOUND` from a lost race) is logged at debug; everything else is
  logged at warning and transport failures still mark the connection
  unhealthy.
- The Python SDK wrote library output to stdout via `print()` (~25 sites:
  connection lifecycle, heartbeats, lease renewal, session context,
  provider runtime, event handlers). It now logs through
  `logging.getLogger("toolplane.…")`, so host applications control output
  and level.
- The Python `delete_tool` wrapper read the normalized tool dict under the
  wrong key (`machineId` vs `machine_id`), always sending an empty
  `machine_id` — which the per-machine credential gate rejects.
- The TypeScript client collapsed seven distinct codes (including
  `UNAUTHENTICATED` and `FAILED_PRECONDITION`) into a single
  `ProtocolError`; they are now distinct classes.

## Compatibility

- Error **messages** stay in the `failed to <action>: <detail>` shape; the
  gRPC **codes** change for previously-misclassified conditions (see the
  table above). Clients that matched on `NOT_FOUND` for claim contention or
  capacity should match on the new codes; the conformance suite's pinned
  expectations (`failed_precondition` for fenced writes, `out_of_range` for
  expired replay) are unchanged.
- Python SDK callers that caught `RequestError`/`ToolError`/… from RPC
  failures should catch `ToolplaneError` (the typed errors subclass it) or
  the specific per-code classes. The domain categories remain for
  client-side failures (marshalling, connection setup).
- Go SDK callers that string-matched error text should switch to
  `errors.Is` with the `client.Err…` sentinels or `status.Code`.
- TypeScript callers that matched `ProtocolError` for the affected codes
  should switch to the specific subclasses (all still extend
  `ToolplaneError`).

## Verification

- New server suite `pkg/service/error_taxonomy_test.go`: the full
  sentinel→code table, plus handler-level assertions for claim
  contention/drain/miss, terminal cancels, unknown tools, over-range
  timeouts, no-provider, and lookup misses.
- New SDK unit suites: `tests/test_errors.py` (Python), `client/errors_test.go`
  (Go), `tests/unit/errors.test.ts` (TypeScript).
