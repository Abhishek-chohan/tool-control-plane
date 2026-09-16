package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"toolplane/pkg/model"
)

// TestWaitForRequestTerminalExitsPromptlyOnThirdPartyCancel pins the waiter
// contract: a CancelRequest from any party wakes the waiter with
// ErrRequestCancelled immediately — not at its own 45s deadline with a
// bogus timeout error.
func TestWaitForRequestTerminalExitsPromptlyOnThirdPartyCancel(t *testing.T) {
	tracer := &recordingTracer{}
	toolService := NewToolService(tracer, nil)
	machineService := NewMachinesService(context.Background(), toolService, tracer, nil)
	requestService := NewRequestsService(context.Background(), toolService, machineService, tracer, nil)

	const sessionID = "session-g5-waiter"
	const machineID = "machine-g5-waiter"
	if _, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	request, err := requestService.CreateRequest(sessionID, "echo", `{}`, 0, "")
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	waitDone := make(chan error, 1)
	started := time.Now()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		_, waitErr := requestService.WaitForRequestTerminal(ctx, sessionID, request.ID)
		waitDone <- waitErr
	}()
	time.Sleep(150 * time.Millisecond)

	if err := requestService.CancelRequest(sessionID, request.ID); err != nil {
		t.Fatalf("third-party cancel: %v", err)
	}

	select {
	case waitErr := <-waitDone:
		if !errors.Is(waitErr, ErrRequestCancelled) {
			t.Fatalf("waiter error = %v, want ErrRequestCancelled", waitErr)
		}
		if elapsed := time.Since(started); elapsed > time.Second {
			t.Fatalf("waiter returned after %v on cancellation, want < 1s", elapsed)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("waiter did not return within 3s of the request cancellation")
	}
}

// TestTaskRecordsCancelledWhenRequestCancelledThirdParty pins the task-side
// completion: cancelling the underlying request (not the task, not the
// waiter context) settles the task as CANCELLED with the cancellation
// reason — the old path mapped the waiter's deadline error to
// "task timed out".
func TestTaskRecordsCancelledWhenRequestCancelledThirdParty(t *testing.T) {
	taskCtx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	tracer := &recordingTracer{}
	toolService := NewToolService(tracer, nil)
	machineService := NewMachinesService(context.Background(), toolService, tracer, nil)
	requestService := NewRequestsService(context.Background(), toolService, machineService, tracer, nil)
	tasksService := NewTasksService(taskCtx, toolService, machineService, requestService, tracer, nil)

	const sessionID = "session-g5-task-cancel"
	const machineID = "machine-g5-task-cancel"
	if _, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	task, err := tasksService.CreateTask(sessionID, "echo", `{}`, "")
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	requestID := waitForActiveTaskRequestID(t, tasksService, task.ID, time.Second)

	if err := requestService.CancelRequest(sessionID, requestID); err != nil {
		t.Fatalf("third-party request cancel: %v", err)
	}

	waitForTaskStatus(t, tasksService, sessionID, task.ID, model.StatusCancelled, 2*time.Second)
	settled, err := tasksService.GetTaskByID(sessionID, task.ID)
	if err != nil {
		t.Fatalf("get settled task: %v", err)
	}
	if settled.Error != "underlying request cancelled" {
		t.Fatalf("task error = %q, want the cancellation reason", settled.Error)
	}
}
