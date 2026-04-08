# Reliability Drills

## Purpose

This document packages Toolplane's maintained failure semantics into a small set of evaluator-facing drills.

The source of truth remains:

- `server/DOCUMENTATION.md` for the runtime contract and limits
- `server/docs/release-gate.md` for the authoritative release signal
- `conformance/cases/` for transport-neutral shared fixtures
- `server/pkg/service/*_test.go` for focused runtime proof of edge cases and limits

## Named Drill Matrix

| Drill | Trigger | Expected behavior | Primary runnable proof | Explicit limit |
| --- | --- | --- | --- | --- |
| D1. Provider crash or lease expiry mid-run | A request is claimed or running and the provider stops making progress until its lease expires | The lease expires, machine capacity is released, and the request is requeued with backoff or dead-lettered after retry exhaustion | `cd server && make release-gate-runtime`; `TestRequestsServiceInMemoryLeaseExpiryRequeuesRunningRequest`; `TestMachinesServiceDrainMachineWaitsForClaimedRequestUntilLeaseExpiryRequeuesIt` | Retries stop after `MaxAttempts`; this is focused runtime proof rather than a shared conformance fixture |
| D2. Caller disconnect during streaming | A caller drops after receiving only part of a stream | The request continues, retained chunks remain inspectable, and replay can resume from the last acknowledged sequence while it is still inside the retained window | `cd server && make release-gate`; `conformance/cases/request_recovery_chunk_window.json`; `conformance/cases/request_recovery_resume.json`; `TestGRPCServerResumeStreamReplaysRetainedWindowAndFinalMarker` | Replay is limited to the retained window |
| D3. Replay behind the retained window | A caller tries to resume from before `start_seq` after the retained window has advanced | Replay fails explicitly with `OUT_OF_RANGE`; missing data is not silently skipped | `cd server && make release-gate`; `conformance/cases/request_recovery_expired_window.json`; `conformance/cases/request_recovery_resume_trimmed_window.json`; `TestGRPCServerResumeStreamReturnsOutOfRangeWhenRetainedWindowExpired`; `TestGRPCServerResumeStreamReplaysTrimmedRetainedWindowAndFinalMarker` | Toolplane does not promise durable full-history replay beyond the retained 100-chunk window |
| D4. Deploy drain with in-flight work | A machine is drained while requests are already claimed or running | New routing stops immediately, in-flight work completes or ages out normally, and the machine disappears only after active work is gone | `cd server && make release-gate`; `conformance/cases/provider_runtime_drain_under_load.json`; `conformance/cases/machine_lifecycle_drain_under_load.json`; `TestMachinesServiceDrainMachineWaitsForInflightRequestAndBlocksNewWork`; `TestMachinesServiceDrainMachineWaitsForClaimedRequestUntilLeaseExpiryRequeuesIt` | Pending requests are not automatically migrated during drain |
| D5. Restart recovery with durable storage | The server reloads while Postgres-backed request state contains persisted leases and retained chunk windows | Persisted leases remain authoritative, expired work follows the same requeue or dead-letter path, and retained replay windows remain available from persisted request state | `cd server && TOOLPLANE_DATABASE_URL=<postgres-url> make release-gate`; `TestRequestsServicePersistentRecoveryRequeuesExpiredRequest` | `TOOLPLANE_STORAGE_MODE=memory` is intentionally non-durable; restart recovery guarantees do not apply there |
| D6. Claim-state and capacity safety | A claim or create path targets work while a machine is draining, or scheduling pushes a machine to overlap ownership or exceed capacity | Drain-blocked create or claim paths are rejected and machine-capacity enforcement stays native runtime behavior instead of silent overlapping ownership | `cd server && make release-gate-runtime`; `TestMachinesServiceDrainMachineWaitsForInflightRequestAndBlocksNewWork`; runtime enforcement in `server/pkg/service/requests.go` and `server/pkg/service/machine.go` | Drain-blocked routing is covered directly; standalone machine-capacity rejection is not yet a named shared fixture or release-gate drill |

## Core Validation Paths

### First Runnable Path

Run the authoritative gate first:

```bash
cd server && TOOLPLANE_DATABASE_URL=postgres://postgres:postgres@localhost:5432/toolplane?sslmode=disable make release-gate
```

That single command gives the shortest evaluator-facing proof path.

- With `TOOLPLANE_DATABASE_URL` set, it covers D1 through D5 directly and the drain-blocked slice of D6.
- Without `TOOLPLANE_DATABASE_URL`, the same command still covers D1 through D4 plus the drain-blocked slice of D6, but D5 remains out of scope because restart recovery is a durable-storage drill.

### Focused Runtime Slice

When you want the narrow Go-side proof without running the full shared-fixture leg:

```bash
cd server && make release-gate-runtime
```

This is the best path for exact lease-expiry, retained-replay, drain-waiting, and persisted-recovery assertions.

### Broader Shared-Fixture Sweep

When you want the broader maintained transport-neutral proof surface outside the authoritative release gate:

```bash
cd server && make conformance-python
```

This is the primary shared-fixture sweep for portable request recovery, provider-runtime behavior, and drain-under-load behavior.

## Proof Layers

### Layer 1: Runtime Contract

- `server/DOCUMENTATION.md` defines what survives disconnect or restart, what retained replay means, and what drain guarantees.
- Treat that document as the prose contract, not as a marketing summary.

### Layer 2: Shared Conformance

- Request-recovery fixtures cover retained-window inspection, replay from the last acknowledged sequence, replay from inside a trimmed retained window, and the explicit expired-window failure path.
- Provider-runtime and machine-lifecycle drain-under-load fixtures cover session-scoped work surviving drain until active requests resolve.

### Layer 3: Focused Runtime Tests

- Lease-expiry requeue behavior, persisted-recovery behavior, retained-window metadata, replay success, replay failure, trace events, and drain waiting behavior are covered most precisely in `server/pkg/service/*_test.go`.
- These tests are the right place to prove narrow edge cases that do not yet belong in transport-neutral shared fixtures.

## What The Drill Matrix Does Not Claim

- No durable full-history replay beyond the retained 100-chunk window.
- No restart durability in memory mode.
- No automatic migration of pending requests during drain.
- No claim that every production-auth path is exercised by the release gate.
- No claim that standalone machine-capacity rejection is already part of the named public proof surface.

If a future reliability claim cannot be mapped back to this matrix, the release gate, shared fixtures, or focused runtime tests, it should not be treated as part of the maintained proof story.
