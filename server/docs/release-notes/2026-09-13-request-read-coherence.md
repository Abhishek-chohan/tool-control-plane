# 2026-09-13 — Request read coherence

## Summary

`GetRequestByID` and `GetRequestByIDAnySession` read the per-instance
cache first and only consulted the store on a cache miss. On a
store-backed replica the cache can be arbitrarily stale — another
replica may have claimed, completed, requeued, or dead-lettered the
request — and the stale copy won every read, forever. Every lifecycle
waiter (synchronous execution, task attempts, stream terminal markers,
replay bookkeeping) reads through these getters, so one stale entry
poisoned all of them.

## Changes

- Both getters are store-first when a store is attached. The store row
  overwrites the cache mirror unconditionally (the old `!exists` guard
  is what let stale entries win). Cache-only reads remain the dev-mode
  path and the fallback for a transient store read error.
- `ExecuteRequest`'s wait loop gained a 250ms poll fallback alongside
  the local update signal: a provider on another replica completes the
  request without any local notification, and the store-first read
  observes it on the next tick.
- A chunk-table gap (window bookkeeping racing the retained-window trim
  across replicas) is now the typed `storage.ErrChunkWindowGap` and maps
  to `RequestStreamExpiredError` — clients see `OUT_OF_RANGE`, not an
  internal error.

## Tests

- `TestGetRequestByIDSeesOtherReplicaCompletion` — B caches the request
  as pending, A completes it, B's next read observes DONE with the
  delivered result intact.
- Full suite plus race detector green against Postgres and memory
  backends.

## Migration

- None. Read latency on store-backed replicas gains one indexed
  row lookup per request read; correctness now matches the
  active-active guarantees the storage layer already provides.
