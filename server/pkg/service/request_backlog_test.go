package service

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
)

func registerBacklogTool(t *testing.T, machineService *MachinesService, sessionID, machineID string) {
	t.Helper()
	_, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, "")
	if err != nil {
		t.Fatalf("register machine: %v", err)
	}
}

// TestCreateRequestBacklogCapRejectsPastCeiling pins the per-session backlog
// cap on the cache-only path: the ceiling accepts exactly
// MaxPendingRequestsPerSession pending creates, rejects the next with
// ErrTooManyPendingRequests, and releases a slot once a request completes.
func TestCreateRequestBacklogCapRejectsPastCeiling(t *testing.T) {
	tracer := &recordingTracer{}
	toolService := NewToolService(tracer, nil)
	machineService := NewMachinesService(context.Background(), toolService, tracer, nil)
	requestService := NewRequestsService(context.Background(), toolService, machineService, tracer, nil)

	const sessionID = "session-backlog-cap-cache"
	const machineID = "machine-backlog-cap-cache"
	registerBacklogTool(t, machineService, sessionID, machineID)

	for i := 0; i < storage.MaxPendingRequestsPerSession; i++ {
		if _, err := requestService.CreateRequest(sessionID, "echo", `{}`, 0, ""); err != nil {
			t.Fatalf("create %d below cap: %v", i, err)
		}
	}
	_, err := requestService.CreateRequest(sessionID, "echo", `{}`, 0, "")
	if !errors.Is(err, ErrTooManyPendingRequests) {
		t.Fatalf("create past cap: err=%v, want ErrTooManyPendingRequests", err)
	}
	if code := status.Code(statusFromDomainError("create request", err)); code != codes.ResourceExhausted {
		t.Fatalf("past-cap status = %v, want ResourceExhausted", code)
	}
	// The storage sentinel (the store-mode verdict) maps to the same wire
	// shape as the service sentinel.
	storeErr := statusFromDomainError("create request", storage.ErrTooManyPendingRequests)
	if code := status.Code(storeErr); code != codes.ResourceExhausted {
		t.Fatalf("storage sentinel status = %v, want ResourceExhausted", code)
	}

	// Completing one request frees backlog capacity: the queue drains, the
	// cap releases.
	claimed, err := requestService.ClaimPendingRequest(sessionID, machineID, []string{"echo"})
	if err != nil || claimed == nil {
		t.Fatalf("claim: claimed=%v err=%v", claimed, err)
	}
	if _, err := requestService.UpdateRequest(sessionID, claimed.ID, machineID, claimed.LeaseEpoch, model.RequestStatusRunning, nil, ""); err != nil {
		t.Fatalf("mark running: %v", err)
	}
	if err := requestService.SubmitRequestResult(sessionID, claimed.ID, machineID, claimed.LeaseEpoch, "done", model.ResultTypeResolution, nil); err != nil {
		t.Fatalf("submit result: %v", err)
	}
	if _, err := requestService.CreateRequest(sessionID, "echo", `{}`, 0, ""); err != nil {
		t.Fatalf("create after terminal transition: %v", err)
	}
}

// TestCreateRequestBacklogCapSharedAcrossReplicas is the authoritative cap
// test: two service stacks sharing one store each count only their own
// creates, so only the store's count-in-transaction verdict holds the global
// ceiling. Both stacks exist before any create, so neither cache can see the
// other's rows.
func TestCreateRequestBacklogCapSharedAcrossReplicas(t *testing.T) {
	store, svcA, machA, svcB, _ := newActiveActiveStacks(t)
	seedSessionForAA(t, store)

	registerBacklogTool(t, machA, "sess-aa", "machine-backlog-aa")

	accepted := 0
	for i := 0; i < storage.MaxPendingRequestsPerSession/2; i++ {
		if _, err := svcA.CreateRequest("sess-aa", "echo", `{}`, 0, ""); err != nil {
			t.Fatalf("replica A create %d: %v", i, err)
		}
		accepted++
	}
	// Replica B's cache counts only B's own creates; the store must reject
	// the create that crosses the shared ceiling.
	for i := 0; i < storage.MaxPendingRequestsPerSession; i++ {
		_, err := svcB.CreateRequest("sess-aa", "echo", `{}`, 0, "")
		if err == nil {
			accepted++
			continue
		}
		if !errors.Is(err, ErrTooManyPendingRequests) {
			t.Fatalf("replica B create %d: err=%v, want ErrTooManyPendingRequests", i, err)
		}
		break
	}
	if accepted != storage.MaxPendingRequestsPerSession {
		t.Fatalf("accepted %d creates across replicas, want exactly %d", accepted, storage.MaxPendingRequestsPerSession)
	}
	// The verdict is global: replica A is rejected too, not just the replica
	// that happened to cross the line.
	if _, err := svcA.CreateRequest("sess-aa", "echo", `{}`, 0, ""); !errors.Is(err, ErrTooManyPendingRequests) {
		t.Fatalf("replica A create after shared cap: err=%v, want ErrTooManyPendingRequests", err)
	}
}

// TestCreateRequestBacklogCapPostgres repeats the ceiling check against the
// durable Postgres insert path: the count and the INSERT share one
// serializable transaction.
func TestCreateRequestBacklogCapPostgres(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("TOOLPLANE_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("TOOLPLANE_DATABASE_URL not set")
	}

	t.Setenv("TOOLPLANE_STORAGE_MODE", "postgres")
	t.Setenv("TOOLPLANE_DATABASE_URL", databaseURL)

	store := openPersistentStoreForTest(t)
	defer func() {
		if err := store.Close(); err != nil {
			t.Fatalf("close store: %v", err)
		}
	}()

	svcCtx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	tracer := &recordingTracer{}
	sessionService := NewSessionsService(tracer, store)
	toolService := NewToolService(tracer, store)
	machineService := NewMachinesService(svcCtx, toolService, tracer, store)
	requestService := NewRequestsService(svcCtx, toolService, machineService, tracer, store)

	session, err := sessionService.CreateSession("backlog-user", "Backlog Cap", "", "", "")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	t.Cleanup(func() {
		_ = sessionService.DeleteSession(session.ID)
	})
	registerBacklogTool(t, machineService, session.ID, "machine-backlog-pg")

	for i := 0; i < storage.MaxPendingRequestsPerSession; i++ {
		if _, err := requestService.CreateRequest(session.ID, "echo", `{}`, 0, ""); err != nil {
			t.Fatalf("create %d below cap: %v", i, err)
		}
	}
	if _, err := requestService.CreateRequest(session.ID, "echo", `{}`, 0, ""); !errors.Is(err, ErrTooManyPendingRequests) {
		t.Fatalf("create past cap: err=%v, want ErrTooManyPendingRequests", err)
	}
}

// TestRequestQueueDepthSourcedFromStore pins the gauge source: the durable
// pending count, not the per-replica cache mirror. A fresh service with no
// local creates still reports the store's count after a refresh.
func TestRequestQueueDepthSourcedFromStore(t *testing.T) {
	store, svcA, machA, svcB, _ := newActiveActiveStacks(t)
	seedSessionForAA(t, store)
	registerBacklogTool(t, machA, "sess-aa", "machine-backlog-depth")

	for i := 0; i < 3; i++ {
		if _, err := svcA.CreateRequest("sess-aa", "echo", `{}`, 0, ""); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}

	// Replica B never created anything: its cache is empty, its store-sourced
	// depth is the authoritative count.
	svcB.refreshQueueDepth()
	if got := svcB.PendingQueueDepth(); got != 3 {
		t.Fatalf("replica B queue depth = %d, want 3 (store count)", got)
	}
	if got := svcA.PendingQueueDepth(); got == -1 {
		t.Fatal("store-backed service reports -1 before any refresh")
	}

	// Cache-only services have no store count and report the fallback.
	tracer := &recordingTracer{}
	toolService := NewToolService(tracer, nil)
	machineService := NewMachinesService(context.Background(), toolService, tracer, nil)
	cacheOnly := NewRequestsService(context.Background(), toolService, machineService, tracer, nil)
	if got := cacheOnly.PendingQueueDepth(); got != -1 {
		t.Fatalf("cache-only queue depth = %d, want -1 (fallback signal)", got)
	}
}
