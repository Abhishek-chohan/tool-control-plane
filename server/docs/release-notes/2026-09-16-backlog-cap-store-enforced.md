# Backlog cap enforced durably across replicas

Date: 2026-09-16

The per-session pending-request ceiling (`MaxPendingRequestsPerSession` =
512) is now enforced where the invariant actually lives: inside the store's
insert transaction, on both the Postgres and memory backends.

- **Why it was broken:** the only cap check counted the per-replica request
  cache. In production storage modes the create path persisted to the store
  and never wrote the cache, so the count stayed at zero — the cap held in
  no production configuration, and even in cache mode R replicas each
  counting their own map admitted R × 512.
- **The fix:** `InsertRequest` counts the session's pending, non-dead-letter
  rows and inserts inside one SERIALIZABLE transaction (Postgres) or under
  the store lock (memory), rejecting with `ErrTooManyPendingRequests`, which
  surfaces as `RESOURCE_EXHAUSTED` with the `SESSION_BACKLOG_FULL` reason
  and a retry hint — the same wire shape the taxonomy already defined. The
  service's cache check stays as an optimistic fast-fail; the store verdict
  is authoritative.
- **Cache mirror:** store-mode creates now mirror the durable row into the
  replica cache (cloned, per the established mirror discipline), so
  cache-first reads and the optimistic check see the replica's own creates.
  Four existing mirror sites that stored shared pointers without cloning
  (two in `CancelRequest`, one in `waitForRequestState`, one in the
  requeue path) were fixed in passing — they let concurrent transitions
  mutate objects a reader was holding.
- **Gauge:** `toolplane_request_queue_depth` is now sourced from a durable
  pending count refreshed on a 5s ticker on store-backed replicas; the
  per-replica cache snapshot remains the fallback for cache-only
  deployments. A store hiccup keeps the last value rather than zeroing the
  gauge.

Notes: the cap releases as before when requests reach terminal states —
the queue drains, capacity frees. Deployments that relied on the
un-enforced cap to queue unbounded backlogs per session will now see
`RESOURCE_EXHAUSTED`/`SESSION_BACKLOG_FULL` past 512 pending requests,
with the documented retry hint.
