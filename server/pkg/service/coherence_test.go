package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage/memory"
	"toolplane/pkg/trace"
)

// newSessionPair builds a SessionsService over a fresh store. The replica
// counterpart is constructed explicitly inside each test AFTER the shared
// state exists, so its startup hydration behaves like a replica that booted
// with the state already present.
func newSessionPair(t *testing.T) (*memory.Store, *SessionsService) {
	t.Helper()
	store := memory.New()
	return store, NewSessionsService(trace.NopTracer(), store)
}

func TestAPIKeyRevocationPropagatesAcrossReplicas(t *testing.T) {
	store, svcA := newSessionPair(t)

	session, err := svcA.CreateSession("user-coherence", "revocation", "coherence", "", "tenant")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	apiKey, err := svcA.CreateApiKey(session.ID, "shared-key", "user-coherence", []string{"read"})
	if err != nil {
		t.Fatalf("create api key: %v", err)
	}

	// Replica B boots after the key exists, so its startup hydration
	// caches the grant; it authenticates.
	svcB := NewSessionsService(trace.NopTracer(), store)
	if _, err := svcB.AuthenticateAPIKey(apiKey.Key); err != nil {
		t.Fatalf("replica B should authenticate the live key: %v", err)
	}

	// Replica A revokes it (persisted). B's cached grant is still fresh, so it
	// keeps authenticating until the revalidation interval passes.
	if err := svcA.RevokeApiKey(session.ID, apiKey.ID); err != nil {
		t.Fatalf("revoke api key: %v", err)
	}

	// Backdate B's verification timestamp: the next authentication must
	// re-read the store and reject the revoked key.
	svcB.apiKeysMutex.Lock()
	svcB.keyVerifiedAt[apiKey.KeyHash] = time.Now().Add(-apiKeyRevalidationInterval - time.Second)
	svcB.apiKeysMutex.Unlock()

	if _, err := svcB.AuthenticateAPIKey(apiKey.Key); err == nil {
		t.Fatal("replica B accepted a key revoked on replica A")
	}
}

func TestSessionDeletionPropagatesToAuthAcrossReplicas(t *testing.T) {
	store, svcA := newSessionPair(t)

	session, err := svcA.CreateSession("user-coherence", "deletion", "coherence", "", "tenant")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	apiKey, err := svcA.CreateApiKey(session.ID, "shared-key", "user-coherence", []string{"read"})
	if err != nil {
		t.Fatalf("create api key: %v", err)
	}
	svcB := NewSessionsService(trace.NopTracer(), store)
	if _, err := svcB.AuthenticateAPIKey(apiKey.Key); err != nil {
		t.Fatalf("replica B should authenticate the live key: %v", err)
	}

	// Deleting the session cascades its keys in the store; B's revalidation
	// must treat the vanished key as authentication failure.
	if err := svcA.DeleteSession(session.ID); err != nil {
		t.Fatalf("delete session: %v", err)
	}

	svcB.apiKeysMutex.Lock()
	svcB.keyVerifiedAt[apiKey.KeyHash] = time.Now().Add(-apiKeyRevalidationInterval - time.Second)
	svcB.apiKeysMutex.Unlock()

	if _, err := svcB.AuthenticateAPIKey(apiKey.Key); err == nil {
		t.Fatal("replica B accepted a key whose session was deleted on replica A")
	}
}

func TestCreateSessionDedupAcrossReplicas(t *testing.T) {
	store, svcA := newSessionPair(t)
	svcB := NewSessionsService(trace.NopTracer(), store)

	const requestedID = "session-coherence-dup"
	if _, err := svcA.CreateSession("user-a", "first", "", requestedID, "tenant"); err != nil {
		t.Fatalf("first create: %v", err)
	}

	// Replica B has no local copy of the session, so only the store insert can
	// detect the collision. It must report AlreadyExists rather than
	// overwriting replica A's row.
	_, err := svcB.CreateSession("user-b", "second", "", requestedID, "tenant")
	if err == nil {
		t.Fatal("duplicate session create across replicas succeeded")
	}

	existing, getErr := svcB.GetSessionByID(requestedID)
	if getErr != nil {
		t.Fatalf("replica B should read the existing session through the store: %v", getErr)
	}
	if existing.CreatedBy != "user-a" {
		t.Fatalf("duplicate create overwrote the original row: createdBy=%s", existing.CreatedBy)
	}
}

func TestListRequestsSeesOtherReplicaCreates(t *testing.T) {
	store := memory.New()
	toolSvc := NewToolService(trace.NopTracer(), store)
	machineSvc := NewMachinesService(context.Background(), toolSvc, trace.NopTracer(), store)
	_ = NewRequestsService(context.Background(), toolSvc, machineSvc, trace.NopTracer(), store)
	svcB := NewRequestsService(context.Background(), toolSvc, machineSvc, trace.NopTracer(), store)

	// A request created and persisted by replica A, never mirrored on B.
	foreign := model.NewRequest("sess-coherence-list", "echo", `{}`)
	foreign.Status = model.RequestStatusPending
	if err := store.SaveRequest(context.Background(), foreign); err != nil {
		t.Fatalf("seed request: %v", err)
	}

	// A local-only request on B must not shadow the store: the read-through
	// lists both.
	local := model.NewRequest("sess-coherence-list", "echo", `{}`)
	if err := store.SaveRequest(context.Background(), local); err != nil {
		t.Fatalf("seed local request: %v", err)
	}

	requests, _, err := svcB.ListRequests("sess-coherence-list", "", "", 10, 0)
	if err != nil {
		t.Fatalf("list requests: %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("replica B listed %d requests, want 2 (store read-through missing)", len(requests))
	}
	seen := map[string]bool{}
	for _, req := range requests {
		seen[req.ID] = true
	}
	if !seen[foreign.ID] || !seen[local.ID] {
		t.Fatal("listed request ids do not include both the foreign and local requests")
	}
}

func TestReserveMachineSlotEnforcesStoreBackedCapacity(t *testing.T) {
	store := memory.New()
	toolSvc := NewToolService(trace.NopTracer(), store)
	machineSvc := NewMachinesService(context.Background(), toolSvc, trace.NopTracer(), store)
	_ = NewRequestsService(context.Background(), toolSvc, machineSvc, trace.NopTracer(), store)

	const sessionID = "sess-coherence-cap"
	const machineID = "machine-coherence-cap"
	ctx := context.Background()

	// Register through the service so the local machine map is populated;
	// the capacity check must still consult the store for foreign load. The
	// tool registration matters too: claims enforce machine-tool ownership.
	if _, err := machineSvc.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine: %v", err)
	}
	if ok := machineSvc.ReserveMachineSlot(sessionID, machineID); !ok {
		t.Fatal("reserve on an idle store-backed machine should succeed")
	}

	// Claims by OTHER replicas against the same machine row count toward
	// capacity here. The transitioning request is already among the counted
	// claimed work, so the reserve denies only when the machine is over the
	// cap, not at it.
	for i := 0; i < maxMachineConcurrentRequests+1; i++ {
		r := model.NewRequest(sessionID, "echo", `{}`)
		r.Status = model.RequestStatusPending
		if err := store.SaveRequest(ctx, r); err != nil {
			t.Fatalf("seed request %d: %v", i, err)
		}
		if _, ok, err := store.ClaimRequest(ctx, sessionID, r.ID, machineID, 30*time.Second); err != nil || !ok {
			t.Fatalf("seed claim %d: ok=%v err=%v", i, ok, err)
		}
	}

	if ok := machineSvc.ReserveMachineSlot(sessionID, machineID); ok {
		t.Fatal("reserve succeeded with the machine at store-backed capacity")
	}
}

func TestDrainFlagVisibleAcrossReplicas(t *testing.T) {
	store := memory.New()
	toolSvc := NewToolService(trace.NopTracer(), store)
	machineSvcA := NewMachinesService(context.Background(), toolSvc, trace.NopTracer(), store)

	const sessionID = "sess-coherence-drain"
	const machineID = "machine-coherence-drain"
	ctx := context.Background()

	if _, err := machineSvcA.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	// Replica B starts after the registration, so its startup hydration sees
	// the machine row and its heartbeats for the machine are accepted.
	machineSvcB := NewMachinesService(context.Background(), toolSvc, trace.NopTracer(), store)

	// Hold one claimed request so the drain waits instead of finishing
	// immediately.
	held := model.NewRequest(sessionID, "echo", `{}`)
	held.Status = model.RequestStatusPending
	if err := store.SaveRequest(ctx, held); err != nil {
		t.Fatalf("seed held request: %v", err)
	}
	claimedHeld, ok, err := store.ClaimRequest(ctx, sessionID, held.ID, machineID, 30*time.Second)
	if err != nil || !ok {
		t.Fatalf("seed held claim: ok=%v err=%v", ok, err)
	}

	drainDone := make(chan error, 1)
	drainCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() { drainDone <- machineSvcA.DrainMachine(drainCtx, sessionID, machineID) }()

	// The persisted flag must become visible to replica B's drain check.
	deadline := time.Now().Add(5 * time.Second)
	for {
		draining, err := store.IsMachineDraining(ctx, machineID)
		if err != nil {
			t.Fatalf("is machine draining: %v", err)
		}
		if draining {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("drain flag was never persisted")
		}
		time.Sleep(5 * time.Millisecond)
	}

	if !machineSvcB.IsMachineDraining(sessionID, machineID) {
		t.Fatal("replica B does not observe the drain through the store")
	}

	// A heartbeat through replica B mid-drain is a column-scoped ping and
	// must not erase the persisted flag: a full upsert here once re-admitted
	// the machine for dispatch on every replica.
	if _, err := machineSvcB.UpdateMachinePing(sessionID, machineID); err != nil {
		t.Fatalf("heartbeat through replica B: %v", err)
	}
	draining, err := store.IsMachineDraining(ctx, machineID)
	if err != nil {
		t.Fatalf("post-heartbeat drain check: %v", err)
	}
	if !draining {
		t.Fatal("heartbeat through replica B erased the persisted drain flag")
	}

	// Release the held work so the drain can complete and unregister.
	if _, err := store.RequeueRequestFenced(ctx, sessionID, held.ID, machineID, claimedHeld.LeaseEpoch, "test release", time.Second); err != nil {
		t.Fatalf("release held request: %v", err)
	}

	select {
	case err := <-drainDone:
		if err != nil {
			t.Fatalf("drain failed: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("drain did not complete after the held request was released")
	}

	// The machine row is gone, so the flag is gone with it.
	draining, err = store.IsMachineDraining(ctx, machineID)
	if err != nil {
		t.Fatalf("post-drain check: %v", err)
	}
	if draining {
		t.Fatal("drain flag outlived the drained machine")
	}
}

// TestGetRequestByIDSeesOtherReplicaCompletion pins the cache-coherence
// contract: instance B cached the request as pending, instance A completes
// it, and B's very next read observes DONE — not the stale cached copy.
// Under the cache-first getter the stale pending entry won forever and
// lifecycle waiters on B never saw the terminal state.
func TestGetRequestByIDSeesOtherReplicaCompletion(t *testing.T) {
	store, svcA, machA, svcB, _ := newActiveActiveStacks(t)
	seedSessionForAA(t, store)

	echoTool := model.NewTool("sess-aa", "machine-a", "echo", "d", `{}`, nil, nil)
	if _, err := machA.RegisterMachine("sess-aa", "machine-a", "1.0", "go", "127.0.0.1", []*model.Tool{echoTool}, ""); err != nil {
		t.Fatalf("RegisterMachine on A: %v", err)
	}

	req, err := svcA.CreateRequest("sess-aa", "echo", `{"x":1}`, 0, "")
	if err != nil {
		t.Fatalf("CreateRequest on A: %v", err)
	}

	// B reads the request (caching the pending state) before A executes it.
	stale, err := svcB.GetRequestByID("sess-aa", req.ID)
	if err != nil || stale.Status != model.RequestStatusPending {
		t.Fatalf("B initial read: status=%v err=%v", stale, err)
	}

	claimed, err := svcA.ClaimRequest("sess-aa", req.ID, "machine-a")
	if err != nil {
		t.Fatalf("ClaimRequest on A: %v", err)
	}
	if err := svcA.SubmitRequestResult("sess-aa", req.ID, "machine-a", claimed.LeaseEpoch,
		map[string]string{"echo": "done"}, model.ResultTypeResolution, nil); err != nil {
		t.Fatalf("SubmitRequestResult on A: %v", err)
	}

	done, err := svcB.GetRequestByID("sess-aa", req.ID)
	if err != nil {
		t.Fatalf("B read after completion: %v", err)
	}
	if done.Status != model.RequestStatusDone {
		t.Fatalf("B read status=%s want done (stale cache won)", done.Status)
	}
	if done.Result == nil {
		t.Fatal("B read lost the delivered result")
	}
}

// TestWaitForRequestTerminalObservesCrossReplicaCompletion drives the
// waiter's ticker path: replica B waits on the request while replica A
// completes it — no local signal fires on B, so only the store-first poll
// observes the terminal state.
func TestWaitForRequestTerminalObservesCrossReplicaCompletion(t *testing.T) {
	store, svcA, machA, svcB, _ := newActiveActiveStacks(t)
	seedSessionForAA(t, store)

	echoTool := model.NewTool("sess-aa", "machine-a", "echo", "d", `{}`, nil, nil)
	if _, err := machA.RegisterMachine("sess-aa", "machine-a", "1.0", "go", "127.0.0.1", []*model.Tool{echoTool}, ""); err != nil {
		t.Fatalf("RegisterMachine on A: %v", err)
	}

	req, err := svcA.CreateRequest("sess-aa", "echo", `{"x":1}`, 0, "")
	if err != nil {
		t.Fatalf("CreateRequest on A: %v", err)
	}

	resultCh := make(chan *model.ToolResult, 1)
	errCh := make(chan error, 1)
	go func() {
		result, err := svcB.WaitForRequestTerminal(context.Background(), "sess-aa", req.ID)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- result
	}()

	// Let B enter its wait, then complete the request on A. B has no local
	// signal for this; only the store-first poll observes it.
	time.Sleep(300 * time.Millisecond)
	claimed, err := svcA.ClaimRequest("sess-aa", req.ID, "machine-a")
	if err != nil {
		t.Fatalf("ClaimRequest on A: %v", err)
	}
	if err := svcA.SubmitRequestResult("sess-aa", req.ID, "machine-a", claimed.LeaseEpoch,
		map[string]string{"echo": "cross-replica"}, model.ResultTypeResolution, nil); err != nil {
		t.Fatalf("SubmitRequestResult on A: %v", err)
	}

	select {
	case result := <-resultCh:
		if result == nil || result.RequestID != req.ID {
			t.Fatalf("waiter result=%+v want request %s", result, req.ID)
		}
	case err := <-errCh:
		t.Fatalf("waiter failed: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("waiter did not observe the cross-replica completion within 5s")
	}
}

// TestClaimOwnershipServicePath pins the claim-path ownership rule at the
// service level: the leasable set is the machine's registered tools
// intersected with the caller's filter, and explicit claims reject foreign
// tools with ErrMachineNotToolOwner.
func TestClaimOwnershipServicePath(t *testing.T) {
	store := memory.New()
	toolSvc := NewToolService(trace.NopTracer(), store)
	machineSvc := NewMachinesService(context.Background(), toolSvc, trace.NopTracer(), store)
	requestSvc := NewRequestsService(context.Background(), toolSvc, machineSvc, trace.NopTracer(), store)

	const sessionID = "sess-claim-ownership"

	if _, err := machineSvc.RegisterMachine(sessionID, "mach-alpha", "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, "mach-alpha", "alpha", "d", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register mach-alpha: %v", err)
	}
	if _, err := machineSvc.RegisterMachine(sessionID, "mach-beta", "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, "mach-beta", "beta", "d", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register mach-beta: %v", err)
	}

	if _, err := requestSvc.CreateRequest(sessionID, "alpha", `{}`, 0, ""); err != nil {
		t.Fatalf("create alpha request: %v", err)
	}
	if _, err := requestSvc.CreateRequest(sessionID, "beta", `{}`, 0, ""); err != nil {
		t.Fatalf("create beta request: %v", err)
	}
	// mach-beta's empty filter leases only its own tool's request.
	claimed, err := requestSvc.ClaimPendingRequest(sessionID, "mach-beta", nil)
	if err != nil {
		t.Fatalf("mach-beta claim: %v", err)
	}
	if claimed.ToolName != "beta" {
		t.Fatalf("mach-beta claimed tool %q, want beta", claimed.ToolName)
	}

	// A name the machine does not own matches nothing.
	if _, err := requestSvc.CreateRequest(sessionID, "alpha", `{}`, 0, ""); err != nil {
		t.Fatalf("create second alpha request: %v", err)
	}
	if _, err := requestSvc.ClaimPendingRequest(sessionID, "mach-beta", []string{"alpha"}); !errors.Is(err, ErrNoPendingRequests) {
		t.Fatalf("mach-beta claim for alpha: err=%v (want ErrNoPendingRequests)", err)
	}

	// Explicit claim of a foreign tool's request is rejected. A fresh beta
	// request: the original was claimed above, and only pending requests are
	// claimable at all.
	foreignBeta, err := requestSvc.CreateRequest(sessionID, "beta", `{}`, 0, "")
	if err != nil {
		t.Fatalf("create beta request for foreign claim: %v", err)
	}
	if _, err := requestSvc.ClaimRequest(sessionID, foreignBeta.ID, "mach-alpha"); !errors.Is(err, ErrMachineNotToolOwner) {
		t.Fatalf("mach-alpha claiming beta request: err=%v (want ErrMachineNotToolOwner)", err)
	}

	// Without a machine identity nothing is leasable.
	if _, err := requestSvc.ClaimPendingRequest(sessionID, "", nil); !errors.Is(err, ErrNoPendingRequests) {
		t.Fatalf("claim without machine identity: err=%v", err)
	}

	// The owner still claims its own.
	alphaClaimed, err := requestSvc.ClaimPendingRequest(sessionID, "mach-alpha", nil)
	if err != nil || alphaClaimed.ToolName != "alpha" {
		t.Fatalf("mach-alpha claim: req=%+v err=%v", alphaClaimed, err)
	}
}
