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
	machineSvc := NewMachinesService(toolSvc, tracer, store)
	requestSvc := NewRequestsService(toolSvc, machineSvc, tracer, store)

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
	}); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	request, err := requestSvc.CreateRequest(session.ID, "echo", `{"message":"persist"}`)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if _, err := requestSvc.ClaimRequest(session.ID, request.ID, machineID); err != nil {
		t.Fatalf("claim request: %v", err)
	}
	if _, err := requestSvc.UpdateRequest(session.ID, request.ID, model.RequestStatusRunning, nil, ""); err != nil {
		t.Fatalf("mark request running: %v", err)
	}

	expiredAt := time.Now().Add(-2 * time.Second)
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
	restartedMachineSvc := NewMachinesService(restartedToolSvc, restartedTracer, restartedStore)
	restartedRequestSvc := NewRequestsService(restartedToolSvc, restartedMachineSvc, restartedTracer, restartedStore)

	restartedRequestSvc.markStalledRequests()

	updated, err := restartedRequestSvc.GetRequestByID(session.ID, request.ID)
	if err != nil {
		t.Fatalf("get recovered request: %v", err)
	}
	if updated.Status != model.RequestStatusPending {
		t.Fatalf("request status = %q, want %q", updated.Status, model.RequestStatusPending)
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
