package service

import (
	"context"
	"testing"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage/memory"
	"toolplane/pkg/trace"
)

func TestCreateRequestIdempotencyKeyDedups(t *testing.T) {
	requestService, _, sessionID, _ := newLeaseTestStack()

	first, err := requestService.CreateRequest(sessionID, "echo", `{"x":1}`, 0, "key-alpha")
	if err != nil {
		t.Fatalf("first create: %v", err)
	}
	retry, err := requestService.CreateRequest(sessionID, "echo", `{"x":1}`, 0, "key-alpha")
	if err != nil {
		t.Fatalf("retry create: %v", err)
	}
	if retry.ID != first.ID {
		t.Fatalf("retry returned a new request %s, want the original %s", retry.ID, first.ID)
	}

	other, err := requestService.CreateRequest(sessionID, "echo", `{"x":1}`, 0, "key-beta")
	if err != nil {
		t.Fatalf("different-key create: %v", err)
	}
	if other.ID == first.ID {
		t.Fatal("a different idempotency key returned the original request")
	}

	// An empty key never dedups: identical calls create distinct requests.
	plainA, err := requestService.CreateRequest(sessionID, "echo", `{"x":1}`, 0, "")
	if err != nil {
		t.Fatalf("plain create A: %v", err)
	}
	plainB, err := requestService.CreateRequest(sessionID, "echo", `{"x":1}`, 0, "")
	if err != nil {
		t.Fatalf("plain create B: %v", err)
	}
	if plainA.ID == plainB.ID {
		t.Fatal("empty idempotency key deduplicated")
	}
}

func TestCreateRequestIdempotencyDedupsAcrossReplicas(t *testing.T) {
	store := memory.New()
	toolSvc := NewToolService(trace.NopTracer(), store)
	machineSvc := NewMachinesService(context.Background(), toolSvc, trace.NopTracer(), store)
	svcA := NewRequestsService(context.Background(), toolSvc, machineSvc, trace.NopTracer(), store)
	svcB := NewRequestsService(context.Background(), toolSvc, machineSvc, trace.NopTracer(), store)

	const sessionID = "sess-idem-aa"
	const machineID = "machine-idem-aa"
	if _, err := machineSvc.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	onA, err := svcA.CreateRequest(sessionID, "echo", `{}`, 0, "key-cross-replica")
	if err != nil {
		t.Fatalf("create on A: %v", err)
	}
	// Replica B has no local copy of A's request: only the store lookup can
	// dedup the key.
	onB, err := svcB.CreateRequest(sessionID, "echo", `{}`, 0, "key-cross-replica")
	if err != nil {
		t.Fatalf("create on B: %v", err)
	}
	if onB.ID != onA.ID {
		t.Fatalf("cross-replica retry returned %s, want the original %s", onB.ID, onA.ID)
	}
}

func TestCreateTaskIdempotencyKeyReturnsExistingWithoutReexecution(t *testing.T) {
	taskCtx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	tracer := trace.NopTracer()
	toolService := NewToolService(tracer, nil)
	machineService := NewMachinesService(taskCtx, toolService, tracer, nil)
	requestService := NewRequestsService(taskCtx, toolService, machineService, tracer, nil)
	tasksService := NewTasksService(taskCtx, toolService, machineService, requestService, tracer, nil)

	const sessionID = "sess-task-idem"
	const machineID = "machine-task-idem"
	if _, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	first, err := tasksService.CreateTask(sessionID, "echo", `{"m":"hi"}`, "key-task-1")
	if err != nil {
		t.Fatalf("first task create: %v", err)
	}

	// The task's attempt runs a request that needs a submitted result to
	// settle. Wait for the attempt's request, then complete it as the
	// executor would.
	deadline := time.Now().Add(5 * time.Second)
	var currentRequestID string
	for {
		task, getErr := tasksService.GetTask(first.ID)
		if getErr != nil {
			t.Fatalf("get task: %v", getErr)
		}
		currentRequestID = task.CurrentRequestID
		if currentRequestID != "" {
			break
		}
		if task.Status == model.StatusCompleted || task.Status == model.StatusFailed {
			t.Fatalf("task settled without an attempt request: status=%s", task.Status)
		}
		if time.Now().After(deadline) {
			t.Fatalf("task never started an attempt: status=%s", task.Status)
		}
		time.Sleep(10 * time.Millisecond)
	}

	attemptReq, err := requestService.GetRequestByID(sessionID, currentRequestID)
	if err != nil {
		t.Fatalf("get attempt request: %v", err)
	}
	if err := requestService.SubmitRequestResult(
		sessionID, currentRequestID, attemptReq.LeasedBy, attemptReq.LeaseEpoch,
		"done", model.ResultTypeResolution, nil,
	); err != nil {
		t.Fatalf("submit attempt result: %v", err)
	}

	for {
		task, getErr := tasksService.GetTask(first.ID)
		if getErr != nil {
			t.Fatalf("get task: %v", getErr)
		}
		if task.Status == model.StatusCompleted || task.Status == model.StatusFailed {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("task did not settle after the result: status=%s", task.Status)
		}
		time.Sleep(10 * time.Millisecond)
	}

	firstFinal, err := tasksService.GetTask(first.ID)
	if err != nil {
		t.Fatalf("get first final: %v", err)
	}

	retry, err := tasksService.CreateTask(sessionID, "echo", `{"m":"hi"}`, "key-task-1")
	if err != nil {
		t.Fatalf("retry task create: %v", err)
	}
	if retry.ID != first.ID {
		t.Fatalf("retry returned a new task %s, want the original %s", retry.ID, first.ID)
	}

	retryFinal, err := tasksService.GetTask(first.ID)
	if err != nil {
		t.Fatalf("get retry final: %v", err)
	}
	if retryFinal.Attempts != firstFinal.Attempts {
		t.Fatalf("retry re-executed the task: attempts %d -> %d", firstFinal.Attempts, retryFinal.Attempts)
	}
}
