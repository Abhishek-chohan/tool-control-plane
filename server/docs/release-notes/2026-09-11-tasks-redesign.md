# Release Note — Task Execution Redesign (2026-09-11)

Branch: `feat/tasks-redesign`

## What changed

Task execution previously self-assigned: the instance whose client created
the task ran it, picked the machine itself, and retried in a private loop.
The startup-only re-adoption claim was a one-shot 30-second touch with no
ownership, no renewal, and no sweeper — so a long-running task could be
re-adopted (and double-executed) by any instance that restarted during it,
a dead instance's tasks waited for the *next* instance restart to be
recovered, and retries never left the creating instance.

### Fenced adoption with renewal

`ClaimTaskForAdoption` is now a real ownership fence: a `tasks.adopted_by`
column records the owning instance, a claim succeeds only when the task is
unowned or the owner's lease (its last touch + a TTL, 30s) has expired, and
`RenewTaskAdoption` / `ReleaseTaskAdoption` (new `Storer` methods, Postgres
and in-memory) keep ownership accurate:

- The executing instance renews every 10 seconds; a failed renewal means
  ownership was lost, and the local execution **cancels itself** instead of
  double-executing.
- Terminal states and scheduled retries release ownership; only the
  recorded owner may release.

### Enqueue, don't self-assign

`CreateTask` no longer runs the task itself: it enqueues the pending task
and a dispatch acquires ownership (uncontended for a fresh task) so
execution still starts immediately on the creating instance. A background
**adoption sweep** (every 5 seconds, the task-side twin of the request
reaper) on every instance claims due tasks that no live instance owns —
recovering tasks from dead instances within one lease TTL and distributing
retries across replicas instead of pinning them to the creator. Retries are
scheduled durably and ownership released, so any instance's sweep picks the
task up when its backoff slot arrives.

### persistTask surfaces failures

`persistTask` derives its context from the service (graceful shutdown),
retries once, and returns its error; the cancel, dead-letter, and
retry-schedule call sites now log the concrete consequence of an undurable
write (a cancelled or exhausted task may re-adopt and re-run after a
restart; a lost retry row forgets its backoff) instead of a bare save
failure.

### CleanupOldTasks is wired — and terminal-only

A periodic sweeper (every 30 minutes) prunes terminal tasks older than 24
hours. `CleanupOldTasks` itself now removes **only** terminal tasks and
never ones currently executing locally: previously any task older than the
cutoff was deleted, including pending/running rows whose deletion would
orphan recoverable work.

In-memory dev mode (no store) keeps the previous direct behavior: the
creating instance runs the task in a private retry loop, unfenced.

## Compatibility

- No proto or client changes: the task API surface is identical; the
  execution model underneath changed.
- Postgres deployments run the `adopted_by` migration on startup.
- `CleanupOldTasks(maxAge)` keeps its signature; callers relying on it to
  remove non-terminal tasks (none exist in-repo) see them retained now.

## Known limits

- Recovery latency for a dead instance's tasks is bounded by the adoption
  lease TTL (30s) plus the sweep interval (5s) — not instantaneous.
- In-memory mode remains unfenced by design (single-process).
- The renewal interval (10s) means an owner can technically keep renewing a
  task whose attempt is stuck without progress; the attempt's own timeout
  and the request-side lease reaper bound that.

## Verification

- New suite `pkg/service/task_adoption_test.go`: claim/renew/release
  fencing (non-owner renew, non-owner release, post-release claim),
  expired-lease steal with stale-owner renewal failure, create-dispatch
  completing end to end through the fence, sweep adoption of an orphaned
  pending task, and cleanup retaining non-terminal tasks past the retention
  age while removing terminal ones.
- Full server suite under `-race`, memory and Postgres modes: zero data
  races; golangci-lint and gofmt clean.
- Full Python conformance (grpc + http + mcp, multi-instance against
  Postgres) and TypeScript conformance 44/44; proto untouched (no drift).
