# 2026-09-13 — Atomic provider claim (`ClaimNextRequest`)

## Summary

Provider runtimes polled for work by listing pending requests and then
claiming the first match — two round-trips and a race: between the list
and the claim another claimant can take the request, so every poll tick
could burn FAILED_PRECONDITION noise and the queue drained at list
frequency. `ClaimNextRequest` collapses the poll into one atomic
round-trip.

## Contract changes (additive)

- `RequestsService.ClaimNextRequest(ClaimNextRequestRequest)
  returns (ClaimNextRequestResponse)` — `POST /api.v1/ClaimNextRequest`.
  `ClaimNextRequestRequest{session_id, machine_id, tool_names}` (empty
  `tool_names` matches every tool in the session).
  `ClaimNextRequestResponse{request, claimed}` — `claimed=false` with an
  empty request is the idle outcome, deliberately not an error, so
  providers poll without exception traffic.
- Authorizer: `provide` capability + session binding, matching
  `ClaimRequest`.

## Implementation

- Backed by the guarded `LeasePendingRequest` (serializable, SKIP LOCKED
  on Postgres), so cross-replica "exactly one winner" is preserved.
- The Python and TypeScript provider runtimes now poll through it: the
  claim response carries the lease grant the fenced provider writes
  present. The TypeScript runtime's `handleRequest` no longer claims, and both runtimes skip polling when no tools are registered —
  the lease arrives with the poll.

## Tests

- Full suites green: Go (service + storage, Postgres and memory),
  Python conformance (40 passed over gRPC and HTTP), TypeScript
  conformance 46/46 including the provider-runtime cases that exercise
  the new poll path, proto drift clean.

## Migration

- None for consumers. Provider runtimes vendored from older commits keep
  working: `ClaimRequest` and the list path are unchanged; the new RPC
  is purely additive.
