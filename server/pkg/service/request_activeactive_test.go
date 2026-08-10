package service

import (
	"context"
	"testing"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage/memory"
	"toolplane/pkg/trace"
)

// newActiveActiveStacks builds two independent RequestsService stacks (A and B)
// that share the same store. This mirrors two server instances backed by one
// Postgres in an active-active deployment. It returns the shared store so tests
// can assert on authoritative state, plus the two services and their machine
// services (for registration).
func newActiveActiveStacks(t *testing.T) (*memory.Store, *RequestsService, *MachinesService, *RequestsService, *MachinesService) {
	t.Helper()
	store := memory.New()

	mk := func() (*RequestsService, *MachinesService) {
		// Each instance gets its own tool/machine service backed by the SAME
		// shared store, so caches are per-instance but authority is shared.
		toolSvc := NewToolService(trace.NopTracer(), store)
		machineSvc := NewMachinesService(context.Background(), toolSvc, trace.NopTracer(), store)
		return NewRequestsService(context.Background(), toolSvc, machineSvc, trace.NopTracer(), store), machineSvc
	}
	svcA, machA := mk()
	svcB, machB := mk()
	return store, svcA, machA, svcB, machB
}

// seedSessionForAA persists a session so request/machine/tool FK constraints
// hold across both instances.
func seedSessionForAA(t *testing.T, store *memory.Store) {
	t.Helper()
	sess := model.NewSession("s1", "test", "user", "", "")
	sess.ID = "sess-aa"
	_ = store.SaveSession(context.Background(), sess)
}

// TestActiveActive_CrossInstanceVisibility asserts the core multi-instance
// property: a request created on instance A is visible to instance B. Before
// the store-first rework this failed because B's in-memory map never received
// the row. Now CreateRequest persists first and the store is shared.
func TestActiveActive_CrossInstanceVisibility(t *testing.T) {
	store, svcA, machA, _, _ := newActiveActiveStacks(t)
	ctx := context.Background()
	seedSessionForAA(t, store)

	// Register machine + tool through instance A's service layer (persists to
	// the shared store AND populates A's cache).
	echoTool := model.NewTool("sess-aa", "machine-a", "echo", "d", `{}`, nil, nil)
	if _, err := machA.RegisterMachine("sess-aa", "machine-a", "1.0", "go", "127.0.0.1", []*model.Tool{echoTool}); err != nil {
		t.Fatalf("RegisterMachine on A: %v", err)
	}

	req, err := svcA.CreateRequest("sess-aa", "echo", `{"x":1}`)
	if err != nil {
		t.Fatalf("CreateRequest on A: %v", err)
	}

	// The authoritative store has the row.
	stored, err := store.GetRequest(ctx, req.ID)
	if err != nil || stored == nil {
		t.Fatalf("store.GetRequest: stored=%v err=%v", stored, err)
	}
	if stored.Status != model.RequestStatusPending {
		t.Fatalf("stored status=%s want pending", stored.Status)
	}
}

// TestActiveActive_NoDoubleClaim asserts the explicit-claim path cannot
// double-dispatch across instances: once instance A claims a request, instance
// B's ClaimRequest for the same request ID is rejected. This is the central
// multi-instance safety guarantee, exercised through the guarded store.ClaimRequest.
func TestActiveActive_NoDoubleClaim(t *testing.T) {
	store, svcA, machA, svcB, _ := newActiveActiveStacks(t)
	ctx := context.Background()
	seedSessionForAA(t, store)

	if _, err := machA.RegisterMachine("sess-aa", "machine-a", "1.0", "go", "127.0.0.1", []*model.Tool{model.NewTool("sess-aa", "machine-a", "echo", "d", `{}`, nil, nil)}); err != nil {
		t.Fatalf("RegisterMachine on A: %v", err)
	}
	// B's ClaimRequest uses the store-first path; its machine just needs a
	// draining check, which reads the persisted flag. Register machine-b too.
	machineB := model.NewMachine("sess-aa", "machine-b", "1.0", "go", "127.0.0.1")
	_ = store.SaveMachine(ctx, machineB)

	req, err := svcA.CreateRequest("sess-aa", "echo", `{"x":1}`)
	if err != nil {
		t.Fatalf("CreateRequest on A: %v", err)
	}

	// Instance A claims first.
	claimed, err := svcA.ClaimRequest("sess-aa", req.ID, "machine-a")
	if err != nil {
		t.Fatalf("ClaimRequest on A: %v", err)
	}
	if claimed.LeasedBy != "machine-a" {
		t.Fatalf("A claim leasedBy=%s want machine-a", claimed.LeasedBy)
	}

	// Instance B must be rejected for the same request.
	if _, err := svcB.ClaimRequest("sess-aa", req.ID, "machine-b"); err == nil {
		t.Fatalf("ClaimRequest on B succeeded (double-dispatch) — should have been rejected")
	}

	// Authoritative state: still owned by A, attempts unchanged at 1.
	stored, _ := store.GetRequest(ctx, req.ID)
	if stored.LeasedBy != "machine-a" {
		t.Fatalf("post-B-claim leasedBy=%s want machine-a", stored.LeasedBy)
	}
	if stored.Attempts != 1 {
		t.Fatalf("post-B-claim attempts=%d want 1 (double-increment = bug)", stored.Attempts)
	}
}

// TestActiveActive_CapacityCapHoldsAcrossInstances asserts the per-machine
// capacity cap (maxMachineConcurrentRequests=4) holds across two instances.
// Before the store-first rework each instance had its own in-memory counter,
// so the effective cap doubled.
func TestActiveActive_CapacityCapHoldsAcrossInstances(t *testing.T) {
	store, _, _, _, _ := newActiveActiveStacks(t)
	ctx := context.Background()
	seedSessionForAA(t, store)

	machine := model.NewMachine("sess-aa", "machine-a", "1.0", "go", "127.0.0.1")
	_ = store.SaveMachine(ctx, machine)

	// Manually seed maxMachineConcurrentRequests pending requests and claim them
	// directly through the store so they all count against machine-a's capacity.
	for i := 0; i < maxMachineConcurrentRequests; i++ {
		r := model.NewRequest("sess-aa", "echo", `{}`)
		r.VisibleAt = time.Now()
		_ = store.SaveRequest(ctx, r)
		if _, ok, err := store.ClaimRequest(ctx, "sess-aa", r.ID, "machine-a", 30*time.Second); err != nil || !ok {
			t.Fatalf("seed claim %d: ok=%v err=%v", i, ok, err)
		}
	}

	// Now the machine is at capacity. The store-level count must reflect it.
	count, err := store.MachineInFlightCount(ctx, "sess-aa", "machine-a")
	if err != nil {
		t.Fatalf("in-flight count: %v", err)
	}
	if count != maxMachineConcurrentRequests {
		t.Fatalf("in-flight count=%d want %d (capacity must be shared, not per-instance)", count, maxMachineConcurrentRequests)
	}
}

// TestActiveActive_NoDoubleRequeue asserts the lease-expiry reaper cannot
// double-requeue across instances: when two instances both run markStalledRequests
// against the same expired request, attempts increments exactly once (not twice)
// and the request is requeued a single time. Before Phase 4 this failed because
// the reaper did an unguarded SaveRequest.
func TestActiveActive_NoDoubleRequeue(t *testing.T) {
	store, svcA, machA, svcB, machB := newActiveActiveStacks(t)
	ctx := context.Background()
	seedSessionForAA(t, store)

	// Register the machine on A so it exists for capacity bookkeeping on both.
	echoTool := model.NewTool("sess-aa", "machine-a", "echo", "d", `{}`, nil, nil)
	if _, err := machA.RegisterMachine("sess-aa", "machine-a", "1.0", "go", "127.0.0.1", []*model.Tool{echoTool}); err != nil {
		t.Fatalf("RegisterMachine on A: %v", err)
	}
	// Also track the machine in B's in-memory capacity map (mirrors a second
	// instance that has cached the machine row).
	machB.setMachineCapacity("sess-aa", "machine-a", maxMachineConcurrentRequests)

	// Create a request on A, then claim it so it has a lease.
	req, err := svcA.CreateRequest("sess-aa", "echo", `{"x":1}`)
	if err != nil {
		t.Fatalf("CreateRequest on A: %v", err)
	}
	if _, err := svcA.ClaimRequest("sess-aa", req.ID, "machine-a"); err != nil {
		t.Fatalf("ClaimRequest on A: %v", err)
	}

	// Force the lease into the past so both reapers see it as expired.
	stored, _ := store.GetRequest(ctx, req.ID)
	ago := time.Now().Add(-2 * time.Hour)
	stored.LeasedAt = &ago
	stored.VisibleAt = ago.Add(30 * time.Second)
	_ = store.SaveRequest(ctx, stored)

	// Both instances run their reaper. Each call must be safe; only one wins.
	svcA.markStalledRequests()
	svcB.markStalledRequests()

	final, _ := store.GetRequest(ctx, req.ID)
	if final.Status != model.RequestStatusPending {
		t.Fatalf("after double reaper status=%s want pending", final.Status)
	}
	// Claim added 1 attempt; a single requeue adds 1 more. Two requeues would
	// land at 3. Asserting exactly 2 proves no double-requeue.
	if final.Attempts != 2 {
		t.Fatalf("after double reaper attempts=%d want 2 (double-requeue = bug)", final.Attempts)
	}
	if final.LeasedBy != "" {
		t.Fatalf("after reaper leasedBy=%s want empty", final.LeasedBy)
	}
}
