# Capacity

## Purpose

This document is the measured capacity envelope of the control plane:
which hot paths dominate, what they cost on reference hardware, and how
to reproduce every number. It exists so claims about load behavior
point at evidence instead of intuition. It is an envelope, not an SLA —
absolute numbers are hardware- and deployment-relative.

## How the numbers were produced

Reference environment: the maintained toolchain container (Go 1.24) on
a developer workstation, with Postgres 16 one docker-network hop away
(`sslmode=disable`), 2026-09-20, default server configuration (25-conn
pool, SERIALIZABLE store transactions). Every figure below is
reproducible:

```bash
cd server
make bench                                   # micro-benchmarks, memory store
TOOLPLANE_DATABASE_URL=<throwaway-db> make bench-postgres   # micro-benchmarks, Postgres
make load SHAPE=agent-turn DURATION=30s      # end-to-end agent-turn load
go run ./cmd/loadgen --soak --duration 30m --report bin/soak-report.json   # sustained window
```

Point `TOOLPLANE_DATABASE_URL` at a throwaway database — benchmark and
load rows accumulate by design.

## Measured envelope

### Micro-benchmarks (single operation, steady state)

| Operation | memory store | Postgres |
| --- | --- | --- |
| Durable insert (SERIALIZABLE count + insert) | ~115 µs | ~4.4–5.2 ms |
| Request lifecycle (create → claim → resolve) | ~110 µs | ~15.1 ms |
| Chunk append, 8×1KiB into a full 100-chunk window | ~13 µs | ~10.6 ms |
| Replay read, full 100×1KiB retained window | ~2.4 µs | ~513 µs |
| Status read (`GetRequest` path) | ~0.4 µs | ~321 µs |
| Discovery, 100 tools in session | ~1.5 µs | ~630 µs |

The shape to internalize: **the memory store is cache-speed for
everything, while on Postgres the two SERIALIZABLE write transactions
dominate** — the durable insert (~4.4 ms) and, more, the full-window
chunk append (~10.6 ms, because every append re-locks the request row,
inserts the batch, trims the window, and re-reads the retained window
in one transaction). Reads are sub-millisecond everywhere.

### End-to-end agent-turn load (embedded server, real gRPC contract)

Single session, burst of 3, in-process fake providers polling claims
every 50 ms:

- synchronous invoke end-to-end (create → claim → execute → resolve →
  waiter wake): **p50 ≈ 51 ms, p95 ≈ 53 ms** on the memory store — the
  fixed per-op floor through the whole stack including the provider
  poll interval
- provider claim poll: p50 ≈ 2–5 ms; resolve: p50 < 1 ms
- dispatch keeps up with sustained agent-turn load: zero
  wait-expiries in every healthy run

### Contention behavior observed under load

- **Serialization retries are the contention signal.** Under a
  lease-expiry requeue storm on Postgres, claim-poll transactions
  occasionally exhaust their 4 retry attempts and surface UNAVAILABLE —
  transient, and absorbed by the SDK retry behavior providers already
  have. Persistent exhaustion (the
  `toolplane_storage_serialization_exhausted_total` counter climbing)
  is the sign that same-session write contention exceeds what the
  retry budget absorbs.
- **Cross-replica load (active-active)** at agent-turn volume showed
  zero serialization exhaustion with consumers on one replica and
  providers on another.
- **Drain under load** completes within bounds (in-flight work
  finishes; nothing dead-letters).
- **Sustained windows** (soak mode) show bounded goroutine growth and a
  memory plateau — leak assertions run in every soak report.

## Operator guidance

- **Postgres write throughput is transaction-bound, not connection
  bound.** Each durable create is one SERIALIZABLE transaction (~4.4 ms
  on reference hardware → a rough ceiling of ~220 creates/s per replica
  for pending-heavy mixes, before retry overhead). If creates are the
  bottleneck, the levers are the transaction shape and the per-session
  serialization they contend on — not pool size.
- **Streaming tools set the write floor.** A tool streaming 1 KiB
  chunks at 10/s holds ~10 ms of serialized append transaction per
  second on Postgres; tens of concurrent streams multiply that. Budget
  provider streaming cadence accordingly or batch chunks.
- **Scrape `/metrics` at ≥5 s intervals**: the gauges walk the request
  cache under a lock; scraping is itself load at high cardinality.
- **Watch two counters first**: `toolplane_storage_serialization_retries_total`
  (rising = contention) and `toolplane_request_dead_letters_total`
  (rising = providers not keeping up).
- **Provider fleet sizing**: each machine executes at most 4 concurrent
  requests; a fleet of N machines serves ~4N concurrent executions.
  Claim polls (one per provider every 50 ms when idle) are cheap reads.

## What this document does not claim

- No multi-node or geographically distributed topology was measured;
  single-replica and dual-replica-on-one-host only.
- Absolute latencies are relative to the reference environment —
  co-located Postgres on the same host will shift every fixed cost
  down. Re-run the benchmarks on the serving hardware before quoting
  numbers.
- No latency or throughput service levels are promised anywhere; the
  soak workflow asserts leak bounds only.
- SDK client behavior (retries, backoff) is modeled by the driver's
  fake providers, not measured from the Python provider process.
