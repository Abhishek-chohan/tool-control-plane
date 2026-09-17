# Serialization retries: jittered, counted, and retryable when exhausted

Date: 2026-09-16

`withSerializableTx` — the wrapper behind every claim, fenced write,
adoption, and ownership operation — retried SQLSTATE 40001/40P01 aborts
with a deterministic 2/4/8ms backoff, no metrics, no logs, and a raw
last error on exhaustion that failed closed to INTERNAL with SQLSTATE
text. Three changes:

- **Full jitter**: each retry sleeps a uniform draw over the exponential
  base (`math/rand/v2` — predictability is harmless here; the draw only
  de-synchronizes replicas). Deterministic backoff lets contending
  instances re-abort in lockstep.
- **Counted**: `toolplane_storage_serialization_retries_total` and
  `toolplane_storage_serialization_exhausted_total` on the operator
  metrics surface, wired through a `SerializationObserver` the server
  attaches to the Postgres store. Exhaustion also logs with the attempt
  budget and last error.
- **Retryable when exhausted**: a new `storage.ErrSerializationConflict`
  sentinel wraps the last error when the attempt budget is spent on
  serialization failures, and the taxonomy maps it to `UNAVAILABLE` —
  which every SDK already retries — instead of INTERNAL. The work was
  never applied, so telling clients to retry is the honest verdict. The
  attempt budget stays at 4; nothing pins it, leaving room to tune.

Notes: only the Postgres store has a serialization retry loop (memory
storage serializes under a lock), so the observer is attached in
Postgres mode only and the counters stay at zero elsewhere.
