package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/trace"
)

type recordingTracer struct {
	mu     sync.Mutex
	events []trace.SessionEvent
}

func (t *recordingTracer) Record(event trace.SessionEvent) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, event)
}

func (t *recordingTracer) hasEvent(eventType trace.SessionEventType) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, event := range t.events {
		if event.Event == eventType {
			return true
		}
	}
	return false
}

func TestRequestsServiceInMemoryLeaseExpiryRequeuesRunningRequest(t *testing.T) {
	tracer := trace.NopTracer()
	toolService := NewToolService(tracer, nil)
	machineService := NewMachinesService(toolService, tracer, nil)
	requestService := NewRequestsService(toolService, machineService, tracer, nil)

	const sessionID = "session-requeue"
	const machineID = "machine-requeue"

	_, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "python", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	})
	if err != nil {
		t.Fatalf("register machine: %v", err)
	}

	request, err := requestService.CreateRequest(sessionID, "echo", `{"message":"lease"}`)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if _, err := requestService.ClaimRequest(sessionID, request.ID, machineID); err != nil {
		t.Fatalf("claim request: %v", err)
	}
	if _, err := requestService.UpdateRequest(sessionID, request.ID, model.RequestStatusRunning, nil, ""); err != nil {
		t.Fatalf("mark running: %v", err)
	}

	request.VisibleAt = time.Now().Add(-time.Second)
	request.UpdatedAt = time.Now().Add(-time.Second)

	requestService.markStalledRequests()

	updated, err := requestService.GetRequestByID(sessionID, request.ID)
	if err != nil {
		t.Fatalf("get request: %v", err)
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
		t.Fatal("expected retry to be scheduled after lease expiry")
	}
	load, _ := machineService.MachineLoadInfo(sessionID, machineID)
	if load != 0 {
		t.Fatalf("machine load = %d, want 0", load)
	}
}

func TestRequestsServiceRecordsProviderLifecycleEvents(t *testing.T) {
	tracer := &recordingTracer{}
	toolService := NewToolService(tracer, nil)
	machineService := NewMachinesService(toolService, tracer, nil)
	requestService := NewRequestsService(toolService, machineService, tracer, nil)

	const sessionID = "session-trace"
	const machineID = "machine-trace"

	_, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "python", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	})
	if err != nil {
		t.Fatalf("register machine: %v", err)
	}

	request, err := requestService.CreateRequest(sessionID, "echo", `{"message":"trace"}`)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if _, err := requestService.ClaimRequest(sessionID, request.ID, machineID); err != nil {
		t.Fatalf("claim request: %v", err)
	}
	if _, err := requestService.UpdateRequest(sessionID, request.ID, model.RequestStatusRunning, nil, ""); err != nil {
		t.Fatalf("mark running: %v", err)
	}
	if err := requestService.AppendRequestChunks(sessionID, request.ID, []string{"chunk-1", "chunk-2"}, model.ResultTypeStreaming); err != nil {
		t.Fatalf("append chunks: %v", err)
	}
	if err := requestService.SubmitRequestResult(sessionID, request.ID, map[string]string{"echo": "trace"}, model.ResultTypeResolution, nil); err != nil {
		t.Fatalf("submit request result: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := machineService.DrainMachine(ctx, sessionID, machineID); err != nil {
		t.Fatalf("drain machine: %v", err)
	}

	for _, eventType := range []trace.SessionEventType{
		trace.EventRequestCreated,
		trace.EventRequestClaimed,
		trace.EventRequestExecutionStarted,
		trace.EventRequestChunksAppended,
		trace.EventRequestExecutionCompleted,
		trace.EventMachineDrainStarted,
		trace.EventMachineDrainCompleted,
	} {
		if !tracer.hasEvent(eventType) {
			t.Fatalf("expected trace event %q to be recorded", eventType)
		}
	}
}
