package service

import (
	"context"
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
	// the capacity check must still consult the store for foreign load.
	if _, err := machineSvc.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", nil, ""); err != nil {
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
	machineSvcB := NewMachinesService(context.Background(), toolSvc, trace.NopTracer(), store)

	const sessionID = "sess-coherence-drain"
	const machineID = "machine-coherence-drain"
	ctx := context.Background()

	if _, err := machineSvcA.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine: %v", err)
	}

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
	draining, err := store.IsMachineDraining(ctx, machineID)
	if err != nil {
		t.Fatalf("post-drain check: %v", err)
	}
	if draining {
		t.Fatal("drain flag outlived the drained machine")
	}
}
