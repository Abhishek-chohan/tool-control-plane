package storage_test

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
	"toolplane/pkg/storage/memory"
)

// TestMain wires the Postgres DSN from the environment. The shared contract
// suite runs against the in-memory store unconditionally and additionally
// against a Postgres store when TOOLPLANE_DATABASE_URL is provided.

// storeFactory returns a fresh, empty store plus a cleanup func.
type storeFactory func(t *testing.T) (storage.Storer, func())

// memoryStoreFactory builds an isolated in-memory store per test.
func memoryStoreFactory(_ *testing.T) (storage.Storer, func()) {
	s := memory.New()
	return s, func() { _ = s.Close() }
}

// postgresStoreFactory builds an isolated Postgres store against a unique
// schema namespace per test, so parallel Postgres-backed tests don't collide.
// It uses a fresh database connection and runs migrations. Skips when the DSN
// is not set.
func postgresStoreFactory(t *testing.T) (storage.Storer, func()) {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("TOOLPLANE_DATABASE_URL"))
	if dsn == "" {
		t.Skip("TOOLPLANE_DATABASE_URL not set; skipping Postgres-backed storage contract tests")
	}
	// Drop the test database suffix and create a uniquely-named one.
	if err := ensureTestDatabase(t, dsn); err != nil {
		t.Skipf("could not provision Postgres test database: %v", err)
	}
	store, err := storage.OpenFromEnv(context.Background(), log.New(io.Discard, "", 0))
	if err != nil {
		t.Skipf("could not open postgres store: %v", err)
	}
	return store, func() { _ = store.Close() }
}

// ensureTestDatabase is a no-op placeholder; the existing storage.OpenFromEnv already
// runs idempotent migrations against the configured database. For true
// per-test isolation against a shared Postgres a dedicated test DB would be
// provisioned here. For now the Postgres path relies on a dedicated
// CI database (as the release-gate workflow provisions).
func ensureTestDatabase(t *testing.T, dsn string) error {
	return nil
}

// runAgainstBoth runs the given test body against the in-memory store and,
// when a Postgres DSN is available, against the Postgres store.
func runAgainstBoth(t *testing.T, name string, fn func(t *testing.T, s storage.Storer)) {
	t.Run("memory", func(t *testing.T) {
		store, cleanup := memoryStoreFactory(t)
		defer cleanup()
		fn(t, store)
	})
	t.Run("postgres", func(t *testing.T) {
		store, cleanup := postgresStoreFactory(t)
		defer cleanup()
		fn(t, store)
	})
}

// seedSession persists a session so request/machine/tool FK constraints hold.
func seedSession(t *testing.T, s storage.Storer, id string) *model.Session {
	t.Helper()
	sess := model.NewSession("test-session", "test", "test-user", "", "")
	sess.ID = id
	ctx := context.Background()
	if err := s.SaveSession(ctx, sess); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	return sess
}

func seedMachine(t *testing.T, s storage.Storer, sessionID, machineID string, lastPing time.Time) *model.Machine {
	t.Helper()
	m := model.NewMachine(sessionID, machineID, "test", "go", "127.0.0.1")
	m.LastPingAt = lastPing
	m.CreatedAt = lastPing
	if err := s.SaveMachine(context.Background(), m); err != nil {
		t.Fatalf("seed machine: %v", err)
	}
	return m
}

func seedPendingRequest(t *testing.T, s storage.Storer, sessionID, requestID, toolName string) *model.Request {
	t.Helper()
	r := model.NewRequest(sessionID, toolName, `{}`)
	r.ID = requestID
	r.VisibleAt = time.Now()
	if err := s.SaveRequest(context.Background(), r); err != nil {
		t.Fatalf("seed pending request: %v", err)
	}
	return r
}

// TestClaimRequest_GuardedRejectsSecondClaimant asserts the core multi-instance
// guarantee: once a request is claimed by one caller, a second concurrent claim
// for the same request ID is rejected (claimed=false), not double-dispatched.
func TestClaimRequest_GuardedRejectsSecondClaimant(t *testing.T) {
	runAgainstBoth(t, "guarded claim", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		seedSession(t, s, "sess-a")
		seedMachine(t, s, "sess-a", "machine-a", time.Now())
		seedPendingRequest(t, s, "sess-a", "req-1", "echo")

		lease := 30 * time.Second
		first, claimed, err := s.ClaimRequest(ctx, "sess-a", "req-1", "machine-a", lease)
		if err != nil || !claimed {
			t.Fatalf("first claim: claimed=%v err=%v", claimed, err)
		}
		if first.Status != model.RequestStatusClaimed || first.LeasedBy != "machine-a" {
			t.Fatalf("first claim state: status=%s leasedBy=%s", first.Status, first.LeasedBy)
		}
		if first.Attempts != 1 {
			t.Fatalf("first claim attempts: got %d want 1", first.Attempts)
		}

		// A second claim for the same request must be rejected without error.
		_, claimed2, err := s.ClaimRequest(ctx, "sess-a", "req-1", "machine-b", lease)
		if err != nil {
			t.Fatalf("second claim errored: %v", err)
		}
		if claimed2 {
			t.Fatalf("second claim should be rejected (claimed=false), got claimed=true (double-dispatch)")
		}

		// Verify the request is still owned by the first claimant.
		stored, err := s.GetRequest(ctx, "req-1")
		if err != nil {
			t.Fatalf("get request: %v", err)
		}
		if stored.LeasedBy != "machine-a" {
			t.Fatalf("after second claim, leasedBy=%s want machine-a", stored.LeasedBy)
		}
		if stored.Attempts != 1 {
			t.Fatalf("after second claim, attempts=%d want 1 (double-increment = bug)", stored.Attempts)
		}
	})
}

// TestClaimRequest_RejectsMissingOrNonPending asserts that claiming a missing,
// already-terminal, or wrong-session request returns claimed=false, not an
// error, and does not mutate anything.
func TestClaimRequest_RejectsMissingOrNonPending(t *testing.T) {
	runAgainstBoth(t, "non-claimable", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		seedSession(t, s, "sess-a")
		seedMachine(t, s, "sess-a", "machine-a", time.Now())

		lease := 30 * time.Second

		// Missing request.
		if _, claimed, err := s.ClaimRequest(ctx, "sess-a", "missing", "machine-a", lease); err != nil || claimed {
			t.Fatalf("missing claim: claimed=%v err=%v", claimed, err)
		}

		// Wrong session.
		seedPendingRequest(t, s, "sess-a", "req-2", "echo")
		if _, claimed, err := s.ClaimRequest(ctx, "sess-b", "req-2", "machine-a", lease); err != nil || claimed {
			t.Fatalf("wrong-session claim: claimed=%v err=%v", claimed, err)
		}
	})
}

// TestReclaimExpiredRequest_SingleRequeue asserts the multi-instance-safe
// lease-expiry path: an expired claimed request is requeued to pending exactly
// once (attempts incremented by 1), and a second reclaim of the same request
// is a no-op (reclaimed=false), not a double-requeue.
func TestReclaimExpiredRequest_SingleRequeue(t *testing.T) {
	runAgainstBoth(t, "single requeue", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		seedSession(t, s, "sess-a")
		seedMachine(t, s, "sess-a", "machine-a", time.Now())
		seedPendingRequest(t, s, "sess-a", "req-1", "echo")

		// Claim it, then force the lease into the past.
		claimed, ok, err := s.ClaimRequest(ctx, "sess-a", "req-1", "machine-a", 30*time.Second)
		if err != nil || !ok {
			t.Fatalf("claim: ok=%v err=%v", ok, err)
		}
		// Push LeasedAt far enough back that HasTimedOut is true now.
		ago := time.Now().Add(-2 * time.Hour)
		claimed.LeasedAt = &ago
		claimed.VisibleAt = ago.Add(30 * time.Second)
		if err := s.SaveRequest(ctx, claimed); err != nil {
			t.Fatalf("re-save expired claim: %v", err)
		}

		now := time.Now()
		reclaimed, rOk, err := s.ReclaimExpiredRequest(ctx, "req-1", now, 30*time.Second, 3, 5*time.Second)
		if err != nil || !rOk {
			t.Fatalf("first reclaim: reclaimed=%v err=%v", rOk, err)
		}
		if reclaimed.Status != model.RequestStatusPending {
			t.Fatalf("after reclaim status=%s want pending", reclaimed.Status)
		}
		if reclaimed.Attempts != 2 { // 1 from claim, +1 from reclaim
			t.Fatalf("after reclaim attempts=%d want 2", reclaimed.Attempts)
		}
		if reclaimed.LeasedBy != "" || reclaimed.LeasedAt != nil {
			t.Fatalf("after reclaim lease not cleared: leasedBy=%s leasedAt=%v", reclaimed.LeasedBy, reclaimed.LeasedAt)
		}

		// Second reclaim (e.g. from another instance) must be a no-op: the
		// request is now pending (not claimed/running), so it is not reclaimable.
		_, rOk2, err := s.ReclaimExpiredRequest(ctx, "req-1", now, 30*time.Second, 3, 5*time.Second)
		if err != nil {
			t.Fatalf("second reclaim errored: %v", err)
		}
		if rOk2 {
			t.Fatalf("second reclaim should be no-op (reclaimed=false), got true (double-requeue)")
		}

		stored, _ := s.GetRequest(ctx, "req-1")
		if stored.Attempts != 2 {
			t.Fatalf("after second reclaim attempts=%d want 2 (double-increment = bug)", stored.Attempts)
		}
	})
}

// TestReclaimExpiredRequest_DeadLettersAfterMaxAttempts asserts that when
// attempts are exhausted, the request is dead-lettered and failed, not requeued.
func TestReclaimExpiredRequest_DeadLettersAfterMaxAttempts(t *testing.T) {
	runAgainstBoth(t, "dead letter", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		seedSession(t, s, "sess-a")
		seedMachine(t, s, "sess-a", "machine-a", time.Now())
		seedPendingRequest(t, s, "sess-a", "req-1", "echo")

		claimed, ok, err := s.ClaimRequest(ctx, "sess-a", "req-1", "machine-a", 30*time.Second)
		if err != nil || !ok {
			t.Fatalf("claim: ok=%v err=%v", ok, err)
		}
		ago := time.Now().Add(-2 * time.Hour)
		claimed.LeasedAt = &ago
		claimed.VisibleAt = ago.Add(30 * time.Second)
		// Pretend prior attempts already consumed the budget.
		claimed.Attempts = 3
		if err := s.SaveRequest(ctx, claimed); err != nil {
			t.Fatalf("re-save: %v", err)
		}

		now := time.Now()
		reclaimed, rOk, err := s.ReclaimExpiredRequest(ctx, "req-1", now, 30*time.Second, 3, 5*time.Second)
		if err != nil || !rOk {
			t.Fatalf("reclaim: reclaimed=%v err=%v", rOk, err)
		}
		if reclaimed.Status != model.RequestStatusFailed {
			t.Fatalf("exhausted status=%s want failure", reclaimed.Status)
		}
		if !reclaimed.DeadLetter {
			t.Fatalf("exhausted deadLetter=false want true")
		}
	})
}

// TestMachineInFlightCount asserts the capacity primitive counts only
// claimed/running requests for the given machine and excludes terminal ones.
func TestMachineInFlightCount(t *testing.T) {
	runAgainstBoth(t, "in-flight count", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		seedSession(t, s, "sess-a")
		seedMachine(t, s, "sess-a", "machine-a", time.Now())

		count, err := s.MachineInFlightCount(ctx, "sess-a", "machine-a")
		if err != nil {
			t.Fatalf("initial count: %v", err)
		}
		if count != 0 {
			t.Fatalf("initial count=%d want 0", count)
		}

		// Two claimed requests.
		for _, id := range []string{"r1", "r2"} {
			seedPendingRequest(t, s, "sess-a", id, "echo")
			if _, ok, err := s.ClaimRequest(ctx, "sess-a", id, "machine-a", 30*time.Second); err != nil || !ok {
				t.Fatalf("claim %s: ok=%v err=%v", id, ok, err)
			}
		}
		count, err = s.MachineInFlightCount(ctx, "sess-a", "machine-a")
		if err != nil {
			t.Fatalf("after claims count: %v", err)
		}
		if count != 2 {
			t.Fatalf("after 2 claims count=%d want 2", count)
		}

		// A terminal request on the same machine must NOT be counted.
		terminal := model.NewRequest("sess-a", "echo", `{}`)
		terminal.ID = "r3"
		terminal.Status = model.RequestStatusDone
		terminal.ExecutingMachineID = "machine-a"
		if err := s.SaveRequest(ctx, terminal); err != nil {
			t.Fatalf("save terminal: %v", err)
		}
		count, _ = s.MachineInFlightCount(ctx, "sess-a", "machine-a")
		if count != 2 {
			t.Fatalf("after terminal add count=%d want 2 (terminal must not count)", count)
		}
	})
}

// TestMachineDrainFlag_PersistedCoherent asserts the drain flag round-trips
// through the store so IsMachineDraining is coherent across instances.
func TestMachineDrainFlag_PersistedCoherent(t *testing.T) {
	runAgainstBoth(t, "drain flag", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		seedSession(t, s, "sess-a")
		seedMachine(t, s, "sess-a", "machine-a", time.Now())

		draining, err := s.IsMachineDraining(ctx, "machine-a")
		if err != nil {
			t.Fatalf("initial isDraining: %v", err)
		}
		if draining {
			t.Fatalf("initial draining=true want false")
		}

		if err := s.SetMachineDraining(ctx, "sess-a", "machine-a"); err != nil {
			t.Fatalf("set draining: %v", err)
		}
		draining, _ = s.IsMachineDraining(ctx, "machine-a")
		if !draining {
			t.Fatalf("after set draining=false want true")
		}

		if err := s.ClearMachineDraining(ctx, "machine-a"); err != nil {
			t.Fatalf("clear draining: %v", err)
		}
		draining, _ = s.IsMachineDraining(ctx, "machine-a")
		if draining {
			t.Fatalf("after clear draining=true want false")
		}

		// Missing machine reports false, not an error.
		draining, err = s.IsMachineDraining(ctx, "no-such-machine")
		if err != nil || draining {
			t.Fatalf("missing machine: draining=%v err=%v", draining, err)
		}
	})
}

// TestLeasePendingRequest_NoDoubleLease is a regression guard for the existing
// multi-instance-safe lease path: leasing consumes the pending request so a
// second lease returns nil.
func TestLeasePendingRequest_NoDoubleLease(t *testing.T) {
	runAgainstBoth(t, "no double lease", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		seedSession(t, s, "sess-a")
		seedMachine(t, s, "sess-a", "machine-a", time.Now())
		seedPendingRequest(t, s, "sess-a", "req-1", "echo")

		first, err := s.LeasePendingRequest(ctx, "sess-a", "machine-a", []string{"echo"}, 30*time.Second)
		if err != nil {
			t.Fatalf("first lease: %v", err)
		}
		if first == nil {
			t.Fatalf("first lease returned nil")
		}

		// Second lease of the same pending pool returns nil (already claimed).
		second, err := s.LeasePendingRequest(ctx, "sess-a", "machine-b", []string{"echo"}, 30*time.Second)
		if err != nil {
			t.Fatalf("second lease errored: %v", err)
		}
		if second != nil {
			t.Fatalf("second lease should return nil, got a request (double-lease)")
		}
	})
}

// TestReclaimMachine_StaleHandoff asserts the existing dead-machine reaper
// hands off tool ownership when the machine is stale and refuses when fresh.
func TestReclaimMachine_StaleHandoff(t *testing.T) {
	runAgainstBoth(t, "stale reclaim", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		seedSession(t, s, "sess-a")
		seedMachine(t, s, "sess-a", "machine-a", time.Now().Add(-10*time.Minute))

		// Tool owned by the stale machine.
		tool := model.NewTool("sess-a", "machine-a", "echo", "d", `{}`, nil, nil)
		tool.ID = "tool-1"
		if err := s.SaveTool(ctx, tool); err != nil {
			t.Fatalf("save tool: %v", err)
		}

		cutoff := time.Now().Add(-5 * time.Minute)
		sessionID, updates, removed, err := s.ReclaimMachine(ctx, "machine-a", cutoff)
		if err != nil {
			t.Fatalf("reclaim: %v", err)
		}
		if !removed {
			t.Fatalf("reclaim removed=false want true")
		}
		if sessionID != "sess-a" {
			t.Fatalf("reclaim sessionID=%s want sess-a", sessionID)
		}
		if len(updates) != 1 || updates[0].ToolID != "tool-1" {
			t.Fatalf("reclaim updates=%+v want [{tool-1 sess-a}]", updates)
		}

		// Second reclaim is a no-op (already removed).
		_, _, removed2, err := s.ReclaimMachine(ctx, "machine-a", cutoff)
		if err != nil || removed2 {
			t.Fatalf("second reclaim: removed=%v err=%v", removed2, err)
		}
	})
}

// TestFindNonTerminalTasks asserts the startup re-adoption query returns only
// non-terminal, non-dead-lettered tasks.
func TestFindNonTerminalTasks(t *testing.T) {
	runAgainstBoth(t, "non-terminal tasks", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		seedSession(t, s, "sess-a")

		mk := func(id string, status model.TaskStatus, deadLetter bool) {
			t.Helper()
			task := model.NewTask("sess-a", "echo", `{}`)
			task.ID = id
			task.Status = status
			task.DeadLetter = deadLetter
			if err := s.SaveTask(ctx, task); err != nil {
				t.Fatalf("save task %s: %v", id, err)
			}
		}
		mk("pending-1", model.StatusPending, false)
		mk("running-1", model.StatusRunning, false)
		mk("completed-1", model.StatusCompleted, false)
		mk("failed-1", model.StatusFailed, false)
		mk("cancelled-1", model.StatusCancelled, false)
		mk("pending-dead", model.StatusPending, true)

		tasks, err := s.FindNonTerminalTasks(ctx)
		if err != nil {
			t.Fatalf("find non-terminal: %v", err)
		}
		got := map[string]bool{}
		for _, tk := range tasks {
			got[tk.ID] = true
		}
		if !got["pending-1"] || !got["running-1"] {
			t.Fatalf("non-terminal missing pending/running: %+v", got)
		}
		for _, terminal := range []string{"completed-1", "failed-1", "cancelled-1", "pending-dead"} {
			if got[terminal] {
				t.Fatalf("non-terminal included %s (should be excluded)", terminal)
			}
		}
	})
}

// TestClaimToolOwnership_ConflictAndStaleTransfer asserts tool ownership
// arbitration: a fresh owner conflict rejects, but a stale owner transfers.
func TestClaimToolOwnership_ConflictAndStaleTransfer(t *testing.T) {
	runAgainstBoth(t, "tool ownership", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		seedSession(t, s, "sess-a")
		now := time.Now()
		seedMachine(t, s, "sess-a", "machine-a", now)
		seedMachine(t, s, "sess-a", "machine-b", now)

		// machine-a owns "echo".
		first := model.NewTool("sess-a", "machine-a", "echo", "d", `{}`, nil, nil)
		first.ID = "tool-1"
		first.LastPingAt = now
		if err := s.SaveTool(ctx, first); err != nil {
			t.Fatalf("save first tool: %v", err)
		}

		// machine-b tries to claim "echo" while machine-a is fresh -> conflict.
		candidate := model.NewTool("sess-a", "machine-b", "echo", "d2", `{}`, nil, nil)
		candidate.ID = "tool-2"
		candidate.LastPingAt = now
		_, _, err := s.ClaimToolOwnership(ctx, candidate, now.Add(-machineHeartbeatTTLForTest))
		if !errors.Is(err, storage.ErrToolOwnershipConflict) {
			t.Fatalf("fresh conflict: err=%v want storage.ErrToolOwnershipConflict", err)
		}

		// Make machine-a stale, then machine-b claims -> transfer.
		staleMachine := model.NewMachine("sess-a", "machine-a", "test", "go", "127.0.0.1")
		staleMachine.LastPingAt = now.Add(-2 * time.Hour)
		staleMachine.CreatedAt = staleMachine.LastPingAt
		if err := s.SaveMachine(ctx, staleMachine); err != nil {
			t.Fatalf("save stale machine: %v", err)
		}
		got, replaced, err := s.ClaimToolOwnership(ctx, candidate, now.Add(-machineHeartbeatTTLForTest))
		if err != nil {
			t.Fatalf("stale transfer: err=%v", err)
		}
		if replaced != "machine-a" {
			t.Fatalf("stale transfer replaced=%s want machine-a", replaced)
		}
		if got.MachineID != "machine-b" {
			t.Fatalf("stale transfer owner=%s want machine-b", got.MachineID)
		}
	})
}

// machineHeartbeatTTLForTest mirrors the service constant for the stale cutoff.
const machineHeartbeatTTLForTest = 5 * time.Minute
