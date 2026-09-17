package service

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
	"toolplane/pkg/storage/memory"
	"toolplane/pkg/trace"
)

func toolFor(sessionID, machineID, name string) *model.Tool {
	return model.NewTool(sessionID, machineID, name, name+" tool", `{}`, nil, nil)
}

// runReregisterReconcile pins the reconcile contract against one backend:
// a re-register upserts the tool set in place — kept names preserve their
// tool IDs, added names are registered, dropped names are removed — with
// no delete-and-recreate window.
func runReregisterReconcile(t *testing.T, store storage.Storer) {
	t.Helper()
	tracer := &recordingTracer{}
	toolService := NewToolService(tracer, store)
	machineService := NewMachinesService(context.Background(), toolService, tracer, store)

	const sessionID = "session-g9-reconcile"
	const machineID = "machine-g9"

	registered, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		toolFor(sessionID, machineID, "alpha"),
		toolFor(sessionID, machineID, "beta"),
	}, "")
	if err != nil {
		t.Fatalf("initial register: %v", err)
	}
	token := registered.Token
	if token == "" {
		t.Fatal("first registration must return the minted token")
	}
	alpha, err := toolService.GetToolByName(sessionID, "alpha")
	if err != nil {
		t.Fatalf("get alpha: %v", err)
	}
	beta, err := toolService.GetToolByName(sessionID, "beta")
	if err != nil {
		t.Fatalf("get beta: %v", err)
	}

	// Re-register with one kept, one kept, one added.
	if _, err := machineService.RegisterMachine(sessionID, machineID, "1.1.0", "go", "127.0.0.1", []*model.Tool{
		toolFor(sessionID, machineID, "alpha"),
		toolFor(sessionID, machineID, "beta"),
		toolFor(sessionID, machineID, "gamma"),
	}, token); err != nil {
		t.Fatalf("re-register: %v", err)
	}
	for name, wantID := range map[string]string{"alpha": alpha.ID, "beta": beta.ID} {
		got, err := toolService.GetToolByName(sessionID, name)
		if err != nil {
			t.Fatalf("get %s after re-register: %v", name, err)
		}
		if got.ID != wantID {
			t.Fatalf("re-register changed tool ID for %s: %s -> %s (kept names must upsert in place)", name, wantID, got.ID)
		}
	}
	if _, err := toolService.GetToolByName(sessionID, "gamma"); err != nil {
		t.Fatalf("added tool gamma missing after re-register: %v", err)
	}

	// Re-register dropping a name: the dropped tool is removed, the kept
	// one still preserves its ID.
	if _, err := machineService.RegisterMachine(sessionID, machineID, "1.2.0", "go", "127.0.0.1", []*model.Tool{
		toolFor(sessionID, machineID, "alpha"),
	}, token); err != nil {
		t.Fatalf("shrinking re-register: %v", err)
	}
	if _, err := toolService.GetToolByName(sessionID, "beta"); err == nil {
		t.Fatal("dropped tool beta still present after shrinking re-register")
	}
	got, err := toolService.GetToolByName(sessionID, "alpha")
	if err != nil {
		t.Fatalf("get alpha after shrink: %v", err)
	}
	if got.ID != alpha.ID {
		t.Fatalf("shrinking re-register changed alpha's ID: %s -> %s", alpha.ID, got.ID)
	}
}

func TestReregisterReconcileCacheOnly(t *testing.T) {
	runReregisterReconcile(t, nil)
}

func TestReregisterReconcileMemoryStore(t *testing.T) {
	runReregisterReconcile(t, memory.New())
}

// TestReregisterNoVisibilityGap pins the defect class directly: while a
// machine re-registers its tool set repeatedly, a concurrent consumer
// reading the tool by name must never observe a missing tool. The old
// delete-first flow removed the row before re-registering it.
func TestReregisterNoVisibilityGap(t *testing.T) {
	store := memory.New()
	tracer := &recordingTracer{}
	toolService := NewToolService(tracer, store)
	machineService := NewMachinesService(context.Background(), toolService, tracer, store)

	const sessionID = "session-g9-gap"
	const machineID = "machine-g9-gap"
	tools := []*model.Tool{toolFor(sessionID, machineID, "alpha")}
	registered, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", tools, "")
	if err != nil {
		t.Fatalf("initial register: %v", err)
	}
	token := registered.Token
	if token == "" {
		t.Fatal("first registration must return the minted token")
	}

	stop := make(chan struct{})
	var misses, reads int64
	var counterMu sync.Mutex
	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		for {
			select {
			case <-stop:
				return
			default:
				if _, err := toolService.GetToolByName(sessionID, "alpha"); err != nil {
					counterMu.Lock()
					misses++
					counterMu.Unlock()
				}
				counterMu.Lock()
				reads++
				counterMu.Unlock()
			}
		}
	}()

	for i := 0; i < 25; i++ {
		if _, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", tools, token); err != nil {
			t.Fatalf("re-register %d: %v", i, err)
		}
	}
	close(stop)
	<-readerDone

	counterMu.Lock()
	defer counterMu.Unlock()
	if misses > 0 {
		t.Fatalf("concurrent reader saw the tool missing %d times across %d reads during re-registers (delete-and-recreate gap)", misses, reads)
	}
	if reads == 0 {
		t.Fatal("concurrent reader never ran")
	}
}

// runColdReplicaCredentialContract pins the identity contract across
// replicas sharing one store: a cold-cache replica re-registering an
// existing machine must verify the presented credential against the
// durable row and must never mint a second credential or reset identity
// age.
func runColdReplicaCredentialContract(t *testing.T, store storage.Storer) {
	t.Helper()
	// Durable stores enforce the machines→sessions foreign key: the
	// session row must exist before either replica registers.
	sessionID := "session-g9-cold"
	if store != nil {
		sessionService := NewSessionsService(trace.NopTracer(), store)
		session, err := sessionService.CreateSession("g9-cold-user", "G9 Cold Replica", "", "", "")
		if err != nil {
			t.Fatalf("create session: %v", err)
		}
		t.Cleanup(func() {
			_ = sessionService.DeleteSession(session.ID, "")
		})
		sessionID = session.ID
	}
	toolA := NewToolService(trace.NopTracer(), store)
	replicaA := NewMachinesService(context.Background(), toolA, trace.NopTracer(), store)
	toolB := NewToolService(trace.NopTracer(), store)
	replicaB := NewMachinesService(context.Background(), toolB, trace.NopTracer(), store)

	const machineID = "machine-g9-cold"

	registered, err := replicaA.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", nil, "")
	if err != nil {
		t.Fatalf("register on replica A: %v", err)
	}
	token := registered.Token
	if token == "" {
		t.Fatal("first registration must return the minted token exactly once")
	}

	storedBefore, err := store.GetMachine(context.Background(), machineID)
	if err != nil || storedBefore == nil {
		t.Fatalf("load stored machine: %v (%v)", err, storedBefore)
	}

	// Cold replica B re-registers with the legitimate credential.
	rereg, err := replicaB.RegisterMachine(sessionID, machineID, "1.1.0", "go", "127.0.0.1", nil, token)
	if err != nil {
		t.Fatalf("cold re-register with the machine's credential: %v", err)
	}
	if rereg.Token != "" {
		t.Fatal("re-registration must not mint or return a new credential")
	}
	storedAfter, err := store.GetMachine(context.Background(), machineID)
	if err != nil || storedAfter == nil {
		t.Fatalf("reload stored machine: %v (%v)", err, storedAfter)
	}
	if storedAfter.TokenHash != storedBefore.TokenHash {
		t.Fatal("cold re-register rotated the machine's stored credential hash")
	}
	if !storedAfter.CreatedAt.Equal(storedBefore.CreatedAt) {
		t.Fatal("cold re-register reset the machine's creation time")
	}

	// A different credential cannot take the identity.
	if _, err := replicaB.RegisterMachine(sessionID, machineID, "1.1.0", "go", "127.0.0.1", nil, "not-the-token"); !errors.Is(err, ErrMachineCredentialRejected) {
		t.Fatalf("re-register with a foreign credential: err=%v, want ErrMachineCredentialRejected", err)
	}
}

func TestColdReplicaCredentialContractMemoryStore(t *testing.T) {
	runColdReplicaCredentialContract(t, memory.New())
}

func TestColdReplicaCredentialContractPostgres(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("TOOLPLANE_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("TOOLPLANE_DATABASE_URL not set")
	}
	t.Setenv("TOOLPLANE_STORAGE_MODE", "postgres")
	t.Setenv("TOOLPLANE_DATABASE_URL", databaseURL)
	runColdReplicaCredentialContract(t, openPersistentStoreForTest(t))
}

// TestReregisterSurvivesRivalTakeoverDuringReconcile: a tool name taken
// over by another live machine is left with its new owner — the
// re-registering machine logs and skips instead of deleting the row.
func TestReregisterSurvivesRivalTakeoverDuringReconcile(t *testing.T) {
	store := memory.New()
	tracer := &recordingTracer{}
	toolService := NewToolService(tracer, store)
	machineService := NewMachinesService(context.Background(), toolService, tracer, store)

	const sessionID = "session-g9-rival"
	const machineA = "machine-g9-rival-a"
	const machineB = "machine-g9-rival-b"

	registeredA, err := machineService.RegisterMachine(sessionID, machineA, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		toolFor(sessionID, machineA, "shared"),
	}, "")
	if err != nil {
		t.Fatalf("register A: %v", err)
	}
	tokenA := registeredA.Token
	if tokenA == "" {
		t.Fatal("A's first registration must return the minted token")
	}
	shared, err := toolService.GetToolByName(sessionID, "shared")
	if err != nil {
		t.Fatalf("get shared: %v", err)
	}

	// A stops heartbeating (goes stale) and B takes the name over.
	stored, err := store.GetMachine(context.Background(), machineA)
	if err != nil || stored == nil {
		t.Fatalf("load A: %v (%v)", err, stored)
	}
	stored.LastPingAt = time.Now().Add(-2 * machineHeartbeatTTL)
	if err := store.SaveMachine(context.Background(), stored); err != nil {
		t.Fatalf("age A: %v", err)
	}
	if _, err := machineService.RegisterMachine(sessionID, machineB, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		toolFor(sessionID, machineB, "shared"),
	}, ""); err != nil {
		t.Fatalf("register B with takeover: %v", err)
	}

	// A re-registers (heartbeat fresh again — it just pinged): the name is
	// owned by live machine B, so A's claim conflicts and is skipped. The
	// tool row must survive with B's ownership and the original ID.
	if _, err := machineService.RegisterMachine(sessionID, machineA, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		toolFor(sessionID, machineA, "shared"),
	}, tokenA); err != nil {
		t.Fatalf("A re-register under conflict: %v", err)
	}
	after, err := toolService.GetToolByName(sessionID, "shared")
	if err != nil {
		t.Fatalf("get shared after A's conflicting re-register: %v (the reconcile must not delete a rival-owned row)", err)
	}
	if after.ID != shared.ID {
		t.Fatalf("shared tool ID changed: %s -> %s", shared.ID, after.ID)
	}
	if after.MachineID != machineB {
		t.Fatalf("shared tool owner = %s, want %s (B's live ownership must stand)", after.MachineID, machineB)
	}
}
