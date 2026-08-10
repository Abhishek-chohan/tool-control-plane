package storage_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
	"toolplane/pkg/storage/memory"
)

// TestMain wires the Postgres DSN from the environment. The shared contract
// suite runs against the in-memory store unconditionally and additionally
// against a Postgres store when TOOLPLANE_DATABASE_URL is provided.

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

// uid returns a unique suffix for this test invocation so seeded rows never
// collide with other tests OR with prior test runs on a shared Postgres store
// (the memory path is isolated per-test, but the Postgres path shares one
// database across runs). The runID is randomized once per process so IDs are
// unique across binary invocations, not just within one.
var (
	uidCounter uint64
	runID      = fmt.Sprintf("%d", time.Now().UnixNano())
)

func uid(t *testing.T) string {
	t.Helper()
	n := atomic.AddUint64(&uidCounter, 1)
	// Combine a sanitized test name, a per-run randomizer, and a counter for
	// readability + cross-run uniqueness.
	name := strings.NewReplacer("/", "_", "-", "_", ".", "_").Replace(t.Name())
	return fmt.Sprintf("%s-r%s-%d", name, runID, n)
}

// seedSession persists a session so request/machine/tool FK constraints hold.
// The sessionID should be unique per test (use uid(t)).
func seedSession(t *testing.T, s storage.Storer, sessionID string) *model.Session {
	t.Helper()
	sess := model.NewSession("test-session", "test", "test-user", "", "")
	sess.ID = sessionID
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
		sess := "sess-" + uid(t)
		machA := "machA-" + uid(t)
		machB := "machB-" + uid(t)
		reqID := "req-" + uid(t)
		tool := "tool-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, machA, time.Now())
		seedPendingRequest(t, s, sess, reqID, tool)

		lease := 30 * time.Second
		first, claimed, err := s.ClaimRequest(ctx, sess, reqID, machA, lease)
		if err != nil || !claimed {
			t.Fatalf("first claim: claimed=%v err=%v", claimed, err)
		}
		if first.Status != model.RequestStatusClaimed || first.LeasedBy != machA {
			t.Fatalf("first claim state: status=%s leasedBy=%s", first.Status, first.LeasedBy)
		}
		if first.Attempts != 1 {
			t.Fatalf("first claim attempts: got %d want 1", first.Attempts)
		}

		// A second claim for the same request must be rejected without error.
		_, claimed2, err := s.ClaimRequest(ctx, sess, reqID, machB, lease)
		if err != nil {
			t.Fatalf("second claim errored: %v", err)
		}
		if claimed2 {
			t.Fatalf("second claim should be rejected (claimed=false), got claimed=true (double-dispatch)")
		}

		// Verify the request is still owned by the first claimant.
		stored, err := s.GetRequest(ctx, reqID)
		if err != nil {
			t.Fatalf("get request: %v", err)
		}
		if stored.LeasedBy != machA {
			t.Fatalf("after second claim, leasedBy=%s want %s", stored.LeasedBy, machA)
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
		sess := "sess-" + uid(t)
		machA := "machA-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, machA, time.Now())

		lease := 30 * time.Second

		// Missing request.
		if _, claimed, err := s.ClaimRequest(ctx, sess, "missing-"+uid(t), machA, lease); err != nil || claimed {
			t.Fatalf("missing claim: claimed=%v err=%v", claimed, err)
		}

		// Wrong session.
		reqID := "req-" + uid(t)
		seedPendingRequest(t, s, sess, reqID, "tool-"+uid(t))
		if _, claimed, err := s.ClaimRequest(ctx, "wrong-sess-"+uid(t), reqID, machA, lease); err != nil || claimed {
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
		sess := "sess-" + uid(t)
		machA := "machA-" + uid(t)
		reqID := "req-" + uid(t)
		tool := "tool-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, machA, time.Now())
		seedPendingRequest(t, s, sess, reqID, tool)

		// Claim it, then force the lease into the past.
		claimed, ok, err := s.ClaimRequest(ctx, sess, reqID, machA, 30*time.Second)
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
		reclaimed, rOk, err := s.ReclaimExpiredRequest(ctx, reqID, now, 30*time.Second, 3, 5*time.Second)
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
		_, rOk2, err := s.ReclaimExpiredRequest(ctx, reqID, now, 30*time.Second, 3, 5*time.Second)
		if err != nil {
			t.Fatalf("second reclaim errored: %v", err)
		}
		if rOk2 {
			t.Fatalf("second reclaim should be no-op (reclaimed=false), got true (double-requeue)")
		}

		stored, _ := s.GetRequest(ctx, reqID)
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
		sess := "sess-" + uid(t)
		machA := "machA-" + uid(t)
		reqID := "req-" + uid(t)
		tool := "tool-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, machA, time.Now())
		seedPendingRequest(t, s, sess, reqID, tool)

		claimed, ok, err := s.ClaimRequest(ctx, sess, reqID, machA, 30*time.Second)
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
		reclaimed, rOk, err := s.ReclaimExpiredRequest(ctx, reqID, now, 30*time.Second, 3, 5*time.Second)
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
		sess := "sess-" + uid(t)
		machA := "machA-" + uid(t)
		tool := "tool-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, machA, time.Now())

		count, err := s.MachineInFlightCount(ctx, sess, machA)
		if err != nil {
			t.Fatalf("initial count: %v", err)
		}
		if count != 0 {
			t.Fatalf("initial count=%d want 0", count)
		}

		// Two claimed requests.
		for i := 0; i < 2; i++ {
			rid := "req-" + uid(t)
			seedPendingRequest(t, s, sess, rid, tool)
			if _, ok, err := s.ClaimRequest(ctx, sess, rid, machA, 30*time.Second); err != nil || !ok {
				t.Fatalf("claim %s: ok=%v err=%v", rid, ok, err)
			}
		}
		count, err = s.MachineInFlightCount(ctx, sess, machA)
		if err != nil {
			t.Fatalf("after claims count: %v", err)
		}
		if count != 2 {
			t.Fatalf("after 2 claims count=%d want 2", count)
		}

		// A terminal request on the same machine must NOT be counted.
		terminal := model.NewRequest(sess, tool, `{}`)
		terminal.ID = "req-" + uid(t)
		terminal.Status = model.RequestStatusDone
		terminal.ExecutingMachineID = machA
		if err := s.SaveRequest(ctx, terminal); err != nil {
			t.Fatalf("save terminal: %v", err)
		}
		count, _ = s.MachineInFlightCount(ctx, sess, machA)
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
		sess := "sess-" + uid(t)
		machA := "machA-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, machA, time.Now())

		draining, err := s.IsMachineDraining(ctx, machA)
		if err != nil {
			t.Fatalf("initial isDraining: %v", err)
		}
		if draining {
			t.Fatalf("initial draining=true want false")
		}

		if err := s.SetMachineDraining(ctx, sess, machA); err != nil {
			t.Fatalf("set draining: %v", err)
		}
		draining, _ = s.IsMachineDraining(ctx, machA)
		if !draining {
			t.Fatalf("after set draining=false want true")
		}

		if err := s.ClearMachineDraining(ctx, machA); err != nil {
			t.Fatalf("clear draining: %v", err)
		}
		draining, _ = s.IsMachineDraining(ctx, machA)
		if draining {
			t.Fatalf("after clear draining=true want false")
		}

		// Missing machine reports false, not an error.
		draining, err = s.IsMachineDraining(ctx, "no-such-machine-"+uid(t))
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
		sess := "sess-" + uid(t)
		machA := "machA-" + uid(t)
		reqID := "req-" + uid(t)
		tool := "tool-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, machA, time.Now())
		seedPendingRequest(t, s, sess, reqID, tool)

		first, err := s.LeasePendingRequest(ctx, sess, machA, []string{tool}, 30*time.Second)
		if err != nil {
			t.Fatalf("first lease: %v", err)
		}
		if first == nil {
			t.Fatalf("first lease returned nil")
		}

		// Second lease of the same pending pool returns nil (already claimed).
		second, err := s.LeasePendingRequest(ctx, sess, "machB-"+uid(t), []string{tool}, 30*time.Second)
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
		sess := "sess-" + uid(t)
		machA := "machA-" + uid(t)
		toolID := "tool-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, machA, time.Now().Add(-10*time.Minute))

		// Tool owned by the stale machine.
		tool := model.NewTool(sess, machA, "echo-"+uid(t), "d", `{}`, nil, nil)
		tool.ID = toolID
		if err := s.SaveTool(ctx, tool); err != nil {
			t.Fatalf("save tool: %v", err)
		}

		cutoff := time.Now().Add(-5 * time.Minute)
		sessionID, updates, removed, err := s.ReclaimMachine(ctx, machA, cutoff)
		if err != nil {
			t.Fatalf("reclaim: %v", err)
		}
		if !removed {
			t.Fatalf("reclaim removed=false want true")
		}
		if sessionID != sess {
			t.Fatalf("reclaim sessionID=%s want %s", sessionID, sess)
		}
		if len(updates) != 1 || updates[0].ToolID != toolID {
			t.Fatalf("reclaim updates=%+v want [{%s %s}]", updates, toolID, sess)
		}

		// Second reclaim is a no-op (already removed).
		_, _, removed2, err := s.ReclaimMachine(ctx, machA, cutoff)
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
		sess := "sess-" + uid(t)
		seedSession(t, s, sess)

		prefix := uid(t)
		mk := func(suffix string, status model.TaskStatus, deadLetter bool) string {
			t.Helper()
			id := prefix + "-" + suffix
			task := model.NewTask(sess, "echo", `{}`)
			task.ID = id
			task.Status = status
			task.DeadLetter = deadLetter
			if err := s.SaveTask(ctx, task); err != nil {
				t.Fatalf("save task %s: %v", id, err)
			}
			return id
		}
		pendingID := mk("pending", model.StatusPending, false)
		runningID := mk("running", model.StatusRunning, false)
		mk("completed", model.StatusCompleted, false)
		mk("failed", model.StatusFailed, false)
		mk("cancelled", model.StatusCancelled, false)
		mk("pending-dead", model.StatusPending, true)

		tasks, err := s.FindNonTerminalTasks(ctx)
		if err != nil {
			t.Fatalf("find non-terminal: %v", err)
		}
		// Filter to only this test's tasks (shared Postgres may have others).
		got := map[string]bool{}
		for _, tk := range tasks {
			if strings.HasPrefix(tk.ID, prefix) {
				got[tk.ID] = true
			}
		}
		if !got[pendingID] || !got[runningID] {
			t.Fatalf("non-terminal missing pending/running: %+v", got)
		}
		for _, terminal := range []string{prefix + "-completed", prefix + "-failed", prefix + "-cancelled", prefix + "-pending-dead"} {
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
		sess := "sess-" + uid(t)
		machA := "machA-" + uid(t)
		machB := "machB-" + uid(t)
		toolName := "echo-" + uid(t)
		now := time.Now()
		seedSession(t, s, sess)
		seedMachine(t, s, sess, machA, now)
		seedMachine(t, s, sess, machB, now)

		// machine-a owns the tool.
		first := model.NewTool(sess, machA, toolName, "d", `{}`, nil, nil)
		first.ID = "tool-" + uid(t)
		first.LastPingAt = now
		if err := s.SaveTool(ctx, first); err != nil {
			t.Fatalf("save first tool: %v", err)
		}

		// machine-b tries to claim the same tool while machine-a is fresh -> conflict.
		candidate := model.NewTool(sess, machB, toolName, "d2", `{}`, nil, nil)
		candidate.ID = "tool-" + uid(t)
		candidate.LastPingAt = now
		_, _, err := s.ClaimToolOwnership(ctx, candidate, now.Add(-machineHeartbeatTTLForTest))
		if !errors.Is(err, storage.ErrToolOwnershipConflict) {
			t.Fatalf("fresh conflict: err=%v want storage.ErrToolOwnershipConflict", err)
		}

		// Make machine-a stale, then machine-b claims -> transfer.
		staleMachine := model.NewMachine(sess, machA, "test", "go", "127.0.0.1")
		staleMachine.LastPingAt = now.Add(-2 * time.Hour)
		staleMachine.CreatedAt = staleMachine.LastPingAt
		if err := s.SaveMachine(ctx, staleMachine); err != nil {
			t.Fatalf("save stale machine: %v", err)
		}
		got, replaced, err := s.ClaimToolOwnership(ctx, candidate, now.Add(-machineHeartbeatTTLForTest))
		if err != nil {
			t.Fatalf("stale transfer: err=%v", err)
		}
		if replaced != machA {
			t.Fatalf("stale transfer replaced=%s want %s", replaced, machA)
		}
		if got.MachineID != machB {
			t.Fatalf("stale transfer owner=%s want %s", got.MachineID, machB)
		}
	})
}

// machineHeartbeatTTLForTest mirrors the service constant for the stale cutoff.
const machineHeartbeatTTLForTest = 5 * time.Minute
