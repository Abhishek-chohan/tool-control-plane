# Capacity: the measured envelope

Date: 2026-09-20

`server/docs/capacity.md` records what the benchmarks, load driver,
drills, and soak actually measured on reference hardware, so load
behavior claims point at evidence:

- **Micro-benchmarks on both stores**: the durable insert and the
  full-window chunk append are the two SERIALIZABLE write transactions
  that dominate Postgres costs (~4.4 ms and ~10.6 ms respectively on
  the reference environment); reads are sub-millisecond everywhere and
  the memory store is cache-speed throughout.
- **End-to-end agent-turn numbers**: synchronous invoke p50 ≈ 51 ms
  through the full stack (create → claim → execute → resolve → waiter
  wake) with fake providers polling at the real cadence.
- **Contention signals**: serialization-retry exhaustion under a
  requeue storm is transient and SDK-retry-absorbed; the exhaustion
  counter is the operator signal that same-session write contention
  exceeds the retry budget. Active-active load showed zero exhaustion
  at agent-turn volume.
- **Operator guidance**: Postgres throughput is transaction-bound, not
  pool-bound; streaming tools set the write floor (budget their chunk
  cadence); scrape `/metrics` at ≥5 s; watch the serialization-retry
  and dead-letter counters first; provider fleets serve ~4 concurrent
  executions per machine.

Every number carries its reproduction command, and the document closes
with what it does not claim (no multi-node topologies, no SLAs,
hardware-relative absolutes). The README's documentation index links
it, and the observability contract cross-references the serialization
counters to it.
