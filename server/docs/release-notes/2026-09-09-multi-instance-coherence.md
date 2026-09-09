# Release Note — Multi-Instance Coherence (2026-09-09)

Branch: `fix/multi-instance-coherence`

## What changed

Several server paths made decisions from per-replica in-memory state even
when a shared Postgres store was configured, so two replicas behind one
database could disagree: listings missed each other's requests, capacity
counted only local work, drain flags were never written where other
replicas could read them, revoked keys kept authenticating until restart,
and racing session creates overwrote each other. This branch makes the
store the point of coherence for those paths.

### Store-backed reads and writes

- **`ListRequests` reads through the store.** Previously it iterated only
  the local map, so a request created on replica A was invisible to
  replica B's listing until some local write happened to mirror it. The
  store result now mirrors into the local cache to keep it warm.
  (New `Storer.ListRequestsBySession`.)
- **`CreateSession` dedups across replicas.** Two instances racing to
  create the same client-supplied session ID previously both "succeeded"
  and the second `SaveSession` upsert silently overwrote the first row.
  Creates now go through `InsertSessionIfAbsent`
  (`INSERT ... ON CONFLICT DO NOTHING`): exactly one winner, the loser
  reports `ALREADY_EXISTS` and the handler returns the existing session as
  before.
- **Machine capacity counts cross-replica work.** `ReserveMachineSlot`
  enforced its limit on a per-replica in-memory counter, so N replicas
  each allowed the full per-machine limit. With a store configured, the
  promotion check now consults `MachineInFlightCount` (claimed+running
  rows for the machine), which sees claims made by every replica. The
  transitioning request is already counted (it was claimed), so the limit
  holds at cap total in-flight; a store error keeps the local grant —
  capacity is a soft limit and the fenced store write remains the hard
  serialization point.
- **Drain flags are persisted.** `SetMachineDraining` existed but had no
  production caller, so the flag that `IsMachineDraining` reads (including
  its store read-through) was never set in practice. `DrainMachine` now
  persists the flag before waiting; unregister removes the row (and the
  flag with it), and a fresh `RegisterMachine` clears any stale flag via
  the upsert (`draining = false`), so a provider that comes back is not
  stuck draining. The in-memory store now also drops its drain marker on
  `DeleteMachine`, matching the Postgres row semantics.
- **API-key revocation propagates.** A key revoked (or a session deleted,
  which cascades its keys) on replica A kept authenticating on replica B
  until B restarted, because the O(1) hash index was only populated at
  startup. Cached grants are now re-verified against the store (new
  `Storer.GetAPIKeyByHash`) when older than `apiKeyRevalidationInterval`
  (5s), so revocations take effect on every replica within roughly that
  interval. A store error keeps the cached grant and retries on the next
  authentication (availability over revocation latency).
- **Serialization failures retry.** `withSerializableTx` aborted on
  SQLSTATE 40001, which is routine contention between replicas under
  SERIALIZABLE isolation. It now retries up to 4 attempts with exponential
  backoff (2ms base) before surfacing the error.

### Race detector: the service package is clean

The service cached `*model.Request` / `*model.Task` pointers and handed
the live objects to callers, so executor goroutines read fields while
request-lifecycle writers mutated them under the lock (23 distinct
reports under `-race`, three failing tests). Read APIs now hand out
clones (`model.Request.Clone`, `model.Task.Clone` — deep copies of the
mutable reference fields), store-fresh objects are cached as clones so
post-mirror reads stay private, and nothing mutates shared state under a
read lock (`ensureRequestDefaults` now normalizes the private copy). The
CI race leg covers every server package again; the `pkg/service`
exclusion and its explanatory comment are gone.

Tests that previously forged lease deadlines by mutating the returned
pointer now go through a same-package test seam
(`mutateCachedRequestForTest`), which applies the mutation under the
write lock.

## Compatibility

- No proto changes. The observable differences are: `ListRequests`
  answers from the store when one is configured (requests from other
  replicas appear); a cross-replica duplicate session create now fails
  with `ALREADY_EXISTS` instead of silently overwriting; promotions to
  running can be denied (`RESOURCE_EXHAUSTED`) when the machine is over
  capacity globally; revoked keys stop authenticating within ~5s on all
  replicas.
- Performance: `ListRequests` performs one store query per call when a
  store is configured (the provider poll loop's per-tick cost class is
  unchanged — claim/renew were already store roundtrips); authentication
  performs at most one store read per key per 5s.

## Known limits

- Revocation propagation is bounded by the 5s revalidation interval, not
  immediate: a key revoked on replica A keeps authenticating on B until
  B's cached grant goes stale. Fail-closed on store error was traded for
  availability deliberately.
- Session *reads* (`GetSessionByID`) on a replica still serve a cached
  copy of a session deleted elsewhere until restart; the security impact
  is bounded because the session's keys stop authenticating via
  revalidation.
- Cross-replica capacity is a soft limit: the check races with concurrent
  claims on other replicas (TOCTOU), so brief overshoot is possible; the
  fenced write remains the hard serialization point.

## Verification

- New Go suite `pkg/service/coherence_test.go`: revocation propagation,
  session-deletion propagation, cross-replica session dedup, ListRequests
  read-through, store-backed capacity, drain-flag visibility across two
  service instances sharing one store.
- New storage test `serialization_test.go` for the 40001 classifier
  (including wrapped errors).
- New conformance case `multi_instance_list_visibility` (grpc, two real
  server processes sharing one Postgres): a request created on instance A
  appears in instance B's listing.
- Full server suite under `-race`, memory mode and Postgres-backed;
  multi-instance conformance (both cases pass); full Python conformance
  (grpc + http + mcp); TypeScript conformance; proto drift clean.
