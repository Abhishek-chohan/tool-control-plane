# Release Note — Lease Renewal and Fenced Provider Writes (2026-09-04)

Branch: `feat/lease-renewal-and-fencing`

## What changed

Two related runtime guarantees land together:

1. **Lease renewal.** A new `RequestsService.RenewRequestLease` RPC extends the
   execution lease of a claimed/running request. Each claim grants a fresh
   `lease_epoch`; the lease deadline (`visible_at`, exposed as
   `lease_expires_at`) moves forward on renewal but never past the request's
   absolute per-attempt timeout (`leased_at + timeout_seconds`). The lease
   reaper now reclaims on either the unrenewed lease deadline or the absolute
   timeout. `CreateRequest`/`ExecuteTool` accept an optional `timeout_seconds`
   override (server default stays 45s; maximum 3600s) so long-running tools
   can run beyond the default without being reclaimed mid-flight.

2. **Fenced provider writes.** `UpdateRequest`, `SubmitRequestResult`,
   `AppendRequestChunks`, and `RenewRequestLease` now require the caller to
   present the current lease grant (`machine_id` + `lease_epoch` from the
   claim response). Missing, stale, or mismatched grants are rejected with
   `FAILED_PRECONDITION`. Fenced writes are checked and persisted atomically
   against the store row (no more blind whole-row upserts on these paths), so
   a stale executor whose lease was reclaimed can no longer write results or
   stream chunks into a request another machine now owns.

The `Request` proto message additionally exposes `leased_by`, `lease_epoch`,
`lease_expires_at`, and `timeout_seconds`.

## Classification

- Contract: **additive** (new RPC, new fields; no field renumbering or
  removals). All three SDK proto copies regenerated from the canonical
  `server/proto/service.proto`.
- Behavior: **behavior-tightening** for the four fenced provider RPCs —
  see migration below.

## Affected surfaces

- Server: `server/pkg/service` (fenced write paths, renewal, reaper),
  `server/pkg/storage` + memory store (guarded fenced primitives),
  `server/pkg/model` (`LeaseEpoch`, `LeaseExpired`), authz policy entry for
  the new RPC, `server/pkg/service/server.go` error mapping
  (`FAILED_PRECONDITION` for lease conflicts).
- Python SDK: `RequestManager` threads the lease grant through all provider
  writes, gains `renew_request_lease()` and a background renewal loop started
  by `ProviderRuntime`; HTTP mirror updated identically; `create_request()`
  accepts `timeout_seconds`.
- TypeScript SDK: `ToolplaneClient.renewRequestLease()`, lease fields on the
  `Request` model, lease context parameters on provider writes, automatic
  renewal loop in `ProviderRuntime` (`leaseRenewalIntervalMs` option).
- Go SDK: regenerated protos only (no provider surface exists).
- Conformance: new `conformance/cases/provider_runtime_fenced_submission.json`
  (grpc + http transports, Python and TypeScript runners); new Go proofs in
  `server/pkg/service/request_lease_test.go` and
  `server/pkg/storage/requests_fenced_test.go`, wired into
  `release-gate-runtime`.

## Migration action

- **Provider authors using maintained SDKs:** none. The Python and TypeScript
  provider runtimes capture the lease from the claim response, present it on
  every write, and renew in-flight leases automatically (every 10s).
- **Custom clients calling the fenced RPCs directly** (`UpdateRequest`,
  `SubmitRequestResult`, `AppendRequestChunks`): you must now send
  `machine_id` and `lease_epoch` exactly as returned by `ClaimRequest`.
  Requests that do not present a valid grant receive `FAILED_PRECONDITION`.
  Long-running executions should call `RenewRequestLease` before the lease
  deadline (30s TTL by default) passes.
- **Operators:** none. The new `lease_epoch` column migrates idempotently on
  startup. In-flight requests claimed before the upgrade carry epoch 0 and
  keep working for their original holder; new claims start at epoch 1+.

## Known limits

- Renewal keeps a lease alive but never past the absolute timeout; a reclaimed
  request re-executes from the start under a fresh epoch (at-least-once). The
  contract still has no client-supplied idempotency key, so non-idempotent
  tools need their own dedupe until that lands.
- Rows claimed pre-upgrade keep epoch 0; the fencing invariant is exact for
  all claims made after the upgrade.
