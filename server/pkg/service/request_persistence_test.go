package service

import (
	"context"
	"io"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
	"toolplane/pkg/trace"
)

func TestRequestsServicePersistentRecoveryRequeuesExpiredRequest(t *testing.T) {
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

	tracer := &recordingTracer{}
	sessionSvc := NewSessionsService(tracer, store)
	toolSvc := NewToolService(tracer, store)
	machineSvc := NewMachinesService(context.Background(), toolSvc, tracer, store)
	// Cancellable context so the first instance's background lease sweep can be
	// stopped before the expired lease is forged below. Only the restarted
	// instance should reclaim the request and record the asserted trace events.
	requestCtx, stopRequestSweep := context.WithCancel(context.Background())
	defer stopRequestSweep()
	requestSvc := NewRequestsService(requestCtx, toolSvc, machineSvc, tracer, store)

	session, err := sessionSvc.CreateSession("persistent-user", "Persistent Recovery", "tier 4 persistence validation", "", "", "")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	t.Cleanup(func() {
		_ = sessionSvc.DeleteSession(session.ID)
	})

	const machineID = "machine-persistent-recovery"
	if _, err := machineSvc.RegisterMachine(session.ID, machineID, "1.0.0", "python", "127.0.0.1", []*model.Tool{
		model.NewTool(session.ID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	request, err := requestSvc.CreateRequest(session.ID, "echo", `{"message":"persist"}`, 0, "")
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	// ClaimRequest and UpdateRequest are store-first: they return new objects
	// and do not mutate the passed-in request in place. Capture the returned
	// objects so the expired-lease re-save below reflects the claimed/running
	// state rather than the stale pre-claim state.
	claimed, err := requestSvc.ClaimRequest(session.ID, request.ID, machineID)
	if err != nil {
		t.Fatalf("claim request: %v", err)
	}
	running, err := requestSvc.UpdateRequest(session.ID, request.ID, machineID, claimed.LeaseEpoch, model.RequestStatusRunning, nil, "")
	if err != nil {
		t.Fatalf("mark request running: %v", err)
	}
	request = running
	if request == nil {
		request = claimed
	}

	// Stop the first instance's background lease sweep and let any in-flight
	// tick drain. No request is expired yet, so the draining sweep reclaims
	// nothing; once it returns, only the restarted instance below can reclaim
	// the expired lease this test forges.
	stopRequestSweep()
	time.Sleep(200 * time.Millisecond)

	expiredAt := time.Now().Add(-30 * time.Second)
	request.TimeoutSeconds = 1
	request.LeasedAt = &expiredAt
	request.VisibleAt = expiredAt
	request.UpdatedAt = expiredAt
	persistCtx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
	if err := store.SaveRequest(persistCtx, request); err != nil {
		cancel()
		t.Fatalf("persist expired request: %v", err)
	}
	cancel()

	restartedStore := openPersistentStoreForTest(t)
	defer func() {
		if err := restartedStore.Close(); err != nil {
			t.Fatalf("close restarted store: %v", err)
		}
	}()

	restartedTracer := &recordingTracer{}
	restartedToolSvc := NewToolService(restartedTracer, restartedStore)
	restartedMachineSvc := NewMachinesService(context.Background(), restartedToolSvc, restartedTracer, restartedStore)
	restartedRequestSvc := NewRequestsService(context.Background(), restartedToolSvc, restartedMachineSvc, restartedTracer, restartedStore)

	// Reclaim is driven by markStalledRequests. Retry it within a bounded window:
	// the expired-lease scan and the guarded reclaim each take their own time
	// samples, so under CI load a single pass can occasionally observe the
	// request a moment before it is eligible. Re-scanning is idempotent, and if
	// the request is genuinely reclaimable one of the passes requeues it.
	var updated *model.Request
	reclaimDeadline := time.Now().Add(10 * time.Second)
	for {
		restartedRequestSvc.markStalledRequests()
		updated, err = restartedRequestSvc.GetRequestByID(session.ID, request.ID)
		if err != nil {
			t.Fatalf("get recovered request: %v", err)
		}
		if updated.Status == model.RequestStatusPending {
			break
		}
		if time.Now().After(reclaimDeadline) {
			t.Fatalf("request status = %q, want %q (after retrying markStalledRequests)", updated.Status, model.RequestStatusPending)
		}
		time.Sleep(100 * time.Millisecond)
	}
	if updated.ExecutingMachineID != "" {
		t.Fatalf("executing machine = %q, want empty", updated.ExecutingMachineID)
	}
	if updated.LastError != "request lease expired" {
		t.Fatalf("last error = %q, want request lease expired", updated.LastError)
	}
	if updated.NextAttemptAt == nil {
		t.Fatal("expected retry to be scheduled after persisted lease expiry")
	}

	for _, eventType := range []trace.SessionEventType{
		trace.EventRequestLeaseExpired,
		trace.EventRequestRequeued,
	} {
		if !restartedTracer.hasEvent(eventType) {
			t.Fatalf("expected trace event %q to be recorded", eventType)
		}
	}
}

func openPersistentStoreForTest(t *testing.T) *storage.Store {
	t.Helper()
	store, err := storage.OpenFromEnv(context.Background(), log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatalf("open persistent store: %v", err)
	}
	return store
}
