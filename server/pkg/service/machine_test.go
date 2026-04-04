package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/trace"
)

func TestMachinesServiceDrainMachineWaitsForInflightRequestAndBlocksNewWork(t *testing.T) {
	toolService := NewToolService(trace.NopTracer(), nil)
	machineService := NewMachinesService(toolService, trace.NopTracer(), nil)
	requestService := NewRequestsService(toolService, machineService, trace.NopTracer(), nil)

	const sessionID = "session-drain"
	const machineID = "machine-drain"

	_, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{\"type\":\"object\"}`, nil, nil),
	})
	if err != nil {
		t.Fatalf("register machine: %v", err)
	}

	activeRequest, err := requestService.CreateRequest(sessionID, "echo", `{"message":"active"}`)
	if err != nil {
		t.Fatalf("create active request: %v", err)
	}
	queuedRequest, err := requestService.CreateRequest(sessionID, "echo", `{"message":"queued"}`)
	if err != nil {
		t.Fatalf("create queued request: %v", err)
	}

	if _, err := requestService.ClaimRequest(sessionID, activeRequest.ID, machineID); err != nil {
		t.Fatalf("claim active request: %v", err)
	}
	if _, err := requestService.UpdateRequest(sessionID, activeRequest.ID, model.RequestStatusRunning, nil, ""); err != nil {
		t.Fatalf("mark active request running: %v", err)
	}

	drainDone := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		drainDone <- machineService.DrainMachine(ctx, sessionID, machineID)
	}()

	select {
	case err := <-drainDone:
		t.Fatalf("drain returned before in-flight work finished: %v", err)
	case <-time.After(150 * time.Millisecond):
	}

	if !machineService.IsMachineDraining(sessionID, machineID) {
		t.Fatal("expected machine to be marked draining")
	}

	if _, err := requestService.CreateRequest(sessionID, "echo", `{"message":"new"}`); err == nil {
		t.Fatal("expected new requests to fail once drain starts")
	}

	if _, err := requestService.ClaimRequest(sessionID, queuedRequest.ID, machineID); err == nil || !strings.Contains(err.Error(), "draining") {
		t.Fatalf("expected draining claim error, got %v", err)
	}

	if _, err := requestService.ClaimPendingRequest(sessionID, machineID, []string{"echo"}); err == nil || !strings.Contains(err.Error(), "draining") {
		t.Fatalf("expected draining auto-claim error, got %v", err)
	}

	if err := requestService.SubmitRequestResult(sessionID, activeRequest.ID, map[string]string{"echo": "done"}, model.ResultTypeResolution, nil); err != nil {
		t.Fatalf("submit active request result: %v", err)
	}

	select {
	case err := <-drainDone:
		if err != nil {
			t.Fatalf("drain failed: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("drain did not finish after in-flight work completed")
	}

	if machineService.IsMachineDraining(sessionID, machineID) {
		t.Fatal("expected draining flag to clear after unregister")
	}

	if _, err := machineService.GetMachineByID(sessionID, machineID); err == nil {
		t.Fatal("expected machine to be unregistered after graceful drain")
	}

	queuedSnapshot, err := requestService.GetRequestByID(sessionID, queuedRequest.ID)
	if err != nil {
		t.Fatalf("load queued request: %v", err)
	}
	if queuedSnapshot.Status != model.RequestStatusPending {
		t.Fatalf("expected queued request to remain pending, got %q", queuedSnapshot.Status)
	}
}

func TestMachinesServiceDrainMachineIsIdempotentForMissingMachine(t *testing.T) {
	toolService := NewToolService(trace.NopTracer(), nil)
	machineService := NewMachinesService(toolService, trace.NopTracer(), nil)
	_ = NewRequestsService(toolService, machineService, trace.NopTracer(), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := machineService.DrainMachine(ctx, "missing-session", "missing-machine"); err != nil {
		t.Fatalf("expected idempotent drain for missing machine, got %v", err)
	}
}
