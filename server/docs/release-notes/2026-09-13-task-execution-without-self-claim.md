# 2026-09-13 — Task execution without self-claim; storage contention fixes

## Summary

Tasks executed their underlying requests by claiming them for an
advisory machine selection. The claim held the lease while the task
only waited for a terminal state: polling providers could never see the
request until the 30-second lease expired and the reaper requeued it.
The observable damage: ~30s of dead time per attempt, two of three
task attempts burned without execution, and any tool slower than the
remaining attempt budget never completed.

## Changes

- `runTaskAttempt` no longer claims. It creates the request, keeps
  `selectMachine` as an advisory fail-fast preflight, and waits through
  the new `RequestsService.WaitForRequestTerminal` — a non-claiming
  observer driven by request-update signals with a 250ms poll fallback
  for completions on other replicas. Task-level attempts and
  request-level attempts are now 1:1 (one provider claim per task
  attempt).
- `ClaimTaskForAdoption` refuses terminal tasks under the row lock on
  both backends, closing the race where a completion or cancellation
  landing between the adoption sweep's candidate query and the adoption
  claim resurrected finished work.
- Storage contention hardening found while validating the change:
  - Migration passes serialize behind a database advisory lock and
    retry deadlock victims — parallel passes (test binaries, replicas
    booting) deadlocked against each other and against concurrent DML.
  - `withSerializableTx` retries deadlock victims (`40P01`) in addition
    to serialization failures (`40001`); both are "system aborted the
    transaction, retry it" outcomes.
  - Startup cache hydration (tools, machines, sessions) gets a 30s
    bound instead of the 2s per-call persistence timeout; under load
    the short bound failed silently and left caches empty.

## Tests

- `TestTasksServiceTaskCompletesViaProvider` — an independent provider
  claims and executes the task's request (5s of work against a 10s
  budget); the task completes with task attempts == 1 and request
  attempts == 1.
- `TestTasksServiceDoesNotClaimRequest` — with no provider running, the
  task's request stays pending, unleased, and claimable.
- `TestClaimTaskForAdoptionRejectsTerminal` (both backends) — terminal
  tasks are never adopted; pending tasks are.
- 16 consecutive full-suite runs against Postgres and memory backends
  pass after the contention fixes (previously ~2 of 3 runs flaked).

## Migration

- None. Provider-visible behavior improves: task requests are visible
  to provider polls immediately instead of after a lease expiry.
