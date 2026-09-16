package service

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/trace"
)

func TestTasksServiceCancelTaskCancelsUnderlyingRequest(t *testing.T) {
	taskCtx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	tracer := &recordingTracer{}
	toolService := NewToolService(tracer, nil)
	machineService := NewMachinesService(context.Background(), toolService, tracer, nil)
	requestService := NewRequestsService(context.Background(), toolService, machineService, tracer, nil)
	tasksService := NewTasksService(taskCtx, toolService, machineService, requestService, tracer, nil)

	const sessionID = "session-task-cancel"
	const machineID = "machine-task-cancel"

	_, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, "")
	if err != nil {
		t.Fatalf("register machine: %v", err)
	}

	task, err := tasksService.CreateTask(sessionID, "echo", `{"message":"cancel me"}`, "")
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	requestID := waitForActiveTaskRequestID(t, tasksService, task.ID, time.Second)
	waitForRequestStatus(t, requestService, sessionID, requestID, model.RequestStatusPending, time.Second)

	if err := tasksService.CancelTask(sessionID, task.ID); err != nil {
		t.Fatalf("cancel task: %v", err)
	}

	waitForTaskStatus(t, tasksService, sessionID, task.ID, model.StatusCancelled, time.Second)
	waitForRequestStatus(t, requestService, sessionID, requestID, model.RequestStatusFailed, time.Second)

	// Present the cancelled request's own lease grant: even a correctly-fenced
	// late submission must be rejected once the request is terminal.
	cancelledReq, err := requestService.GetRequestByID(sessionID, requestID)
	if err != nil {
		t.Fatalf("get cancelled request: %v", err)
	}
	if err := requestService.SubmitRequestResult(sessionID, requestID, machineID, cancelledReq.LeaseEpoch, map[string]string{"echo": "late"}, model.ResultTypeResolution, nil); err == nil {
		t.Fatal("expected late request completion to be rejected after task cancellation")
	}

	updatedTask, err := tasksService.GetTaskByID(sessionID, task.ID)
	if err != nil {
		t.Fatalf("get task after cancellation: %v", err)
	}
	if updatedTask.Status != model.StatusCancelled {
		t.Fatalf("task status = %q, want %q", updatedTask.Status, model.StatusCancelled)
	}
	if updatedTask.Error != "Task cancelled by user" {
		t.Fatalf("task error = %q, want Task cancelled by user", updatedTask.Error)
	}

	request, err := requestService.GetRequestByID(sessionID, requestID)
	if err != nil {
		t.Fatalf("get request after cancellation: %v", err)
	}
	if request.Error != "Request was cancelled" {
		t.Fatalf("request error = %q, want Request was cancelled", request.Error)
	}

	for _, eventType := range []trace.SessionEventType{
		trace.EventTaskCreated,
		trace.EventTaskExecutionStarted,
		trace.EventTaskCancelled,
	} {
		if !tracer.hasEvent(eventType) {
			t.Fatalf("expected trace event %q to be recorded", eventType)
		}
	}

	startedEvent, ok := tracer.event(trace.EventTaskExecutionStarted)
	if !ok {
		t.Fatal("expected task execution started event")
	}
	if startedEvent.TaskID != task.ID {
		t.Fatalf("task execution event task id = %q, want %q", startedEvent.TaskID, task.ID)
	}
	if startedEvent.RequestID != requestID {
		t.Fatalf("task execution event request id = %q, want %q", startedEvent.RequestID, requestID)
	}
}

func TestTasksServiceTimeoutCancelsUnderlyingRequest(t *testing.T) {
	taskCtx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	tracer := &recordingTracer{}
	toolService := NewToolService(tracer, nil)
	machineService := NewMachinesService(context.Background(), toolService, tracer, nil)
	requestService := NewRequestsService(context.Background(), toolService, machineService, tracer, nil)
	tasksService := NewTasksService(taskCtx, toolService, machineService, requestService, tracer, nil)

	const sessionID = "session-task-timeout"
	const machineID = "machine-task-timeout"

	_, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, "")
	if err != nil {
		t.Fatalf("register machine: %v", err)
	}

	task := model.NewTask(sessionID, "echo", `{"message":"slow"}`)
	task.TimeoutSeconds = 1
	tasksService.tasksMutex.Lock()
	tasksService.tasks[task.ID] = task
	tasksService.tasksMutex.Unlock()

	go tasksService.executeTask(task)

	requestID := waitForActiveTaskRequestID(t, tasksService, task.ID, time.Second)
	waitForRequestStatus(t, requestService, sessionID, requestID, model.RequestStatusPending, time.Second)
	waitForTaskStatus(t, tasksService, sessionID, task.ID, model.StatusFailed, 3*time.Second)
	waitForRequestStatus(t, requestService, sessionID, requestID, model.RequestStatusFailed, time.Second)

	updatedTask, err := tasksService.GetTaskByID(sessionID, task.ID)
	if err != nil {
		t.Fatalf("get timed-out task: %v", err)
	}
	if updatedTask.Error != "task timed out after 1s" {
		t.Fatalf("task error = %q, want task timed out after 1s", updatedTask.Error)
	}
	if updatedTask.Attempts != 1 {
		t.Fatalf("task attempts = %d, want 1", updatedTask.Attempts)
	}
	if updatedTask.Status != model.StatusFailed {
		t.Fatalf("task status = %q, want %q", updatedTask.Status, model.StatusFailed)
	}

	requests, _, err := requestService.ListRequests(sessionID, "", "echo", 10, 0)
	if err != nil {
		t.Fatalf("list requests: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("request count = %d, want 1", len(requests))
	}

	request, err := requestService.GetRequestByID(sessionID, requestID)
	if err != nil {
		t.Fatalf("get timed-out request: %v", err)
	}
	if request.Error != "Request was cancelled" {
		t.Fatalf("request error = %q, want Request was cancelled", request.Error)
	}

	for _, eventType := range []trace.SessionEventType{
		trace.EventTaskExecutionStarted,
		trace.EventTaskExecutionFailed,
	} {
		if !tracer.hasEvent(eventType) {
			t.Fatalf("expected trace event %q to be recorded", eventType)
		}
	}
}

func waitForTaskStatus(t *testing.T, tasksService *TasksService, sessionID, taskID string, want model.TaskStatus, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		task, err := tasksService.GetTaskByID(sessionID, taskID)
		if err == nil && task.Status == want {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	task, err := tasksService.GetTaskByID(sessionID, taskID)
	if err != nil {
		t.Fatalf("get task %s: %v", taskID, err)
	}
	t.Fatalf("task %s status = %q, want %q", taskID, task.Status, want)
}

func waitForRequestStatus(t *testing.T, requestsService *RequestsService, sessionID, requestID string, want model.RequestStatus, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		request, err := requestsService.GetRequestByID(sessionID, requestID)
		if err == nil && request.Status == want {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	request, err := requestsService.GetRequestByID(sessionID, requestID)
	if err != nil {
		t.Fatalf("get request %s: %v", requestID, err)
	}
	t.Fatalf("request %s status = %q, want %q", requestID, request.Status, want)
}

func waitForActiveTaskRequestID(t *testing.T, tasksService *TasksService, taskID string, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		requestID, _ := tasksService.taskExecutionSnapshot(taskID)
		if requestID != "" {
			return requestID
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for active request for task %s", taskID)
	return ""
}

// TestTasksServiceTaskCompletesViaProvider pins the corrected execution
// shape: the task leaves its request pending, an independent provider claims
// and executes it, and the task completes with the task-level attempt counter
// 1:1 against the request-level claim. Under the old self-claim the task held
// the lease itself, the provider never saw the request, and the task timed out.
func TestTasksServiceTaskCompletesViaProvider(t *testing.T) {
	taskCtx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	tracer := &recordingTracer{}
	toolService := NewToolService(tracer, nil)
	machineService := NewMachinesService(context.Background(), toolService, tracer, nil)
	requestService := NewRequestsService(context.Background(), toolService, machineService, tracer, nil)
	tasksService := NewTasksService(taskCtx, toolService, machineService, requestService, tracer, nil)

	const sessionID = "session-task-provider"
	const machineID = "machine-task-provider"

	_, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "slow", "slow tool", `{"type":"object"}`, nil, nil),
	}, "")
	if err != nil {
		t.Fatalf("register machine: %v", err)
	}

	// Independent provider: poll-claim, mark running, execute (5s of work),
	// and submit the result — the same shape as the SDK provider runtime.
	go func() {
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			claimed, err := requestService.ClaimPendingRequest(sessionID, machineID, []string{"slow"})
			if err != nil || claimed == nil {
				time.Sleep(20 * time.Millisecond)
				continue
			}
			if _, err := requestService.UpdateRequest(sessionID, claimed.ID, machineID, claimed.LeaseEpoch, model.RequestStatusRunning, nil, ""); err != nil {
				t.Errorf("provider running update: %v", err)
				return
			}
			time.Sleep(5 * time.Second)
			_ = requestService.SubmitRequestResult(sessionID, claimed.ID, machineID, claimed.LeaseEpoch,
				map[string]string{"done": "true"}, model.ResultTypeResolution, nil)
			return
		}
	}()

	task, err := tasksService.CreateTask(sessionID, "slow", `{"message":"work"}`, "")
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	// Generous ceiling: 10s task timeout against 5s of provider work.
	waitForTaskStatus(t, tasksService, sessionID, task.ID, model.StatusCompleted, 15*time.Second)

	completed, err := tasksService.GetTaskByID(sessionID, task.ID)
	if err != nil {
		t.Fatalf("get completed task: %v", err)
	}
	if completed.Attempts != 1 {
		t.Fatalf("task attempts=%d want 1 (task and request counters must stay 1:1)", completed.Attempts)
	}
	requests, _, err := requestService.ListRequests(sessionID, "", "slow", 10, 0)
	if err != nil {
		t.Fatalf("list requests: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("request count=%d want 1", len(requests))
	}
	if requests[0].Attempts != 1 {
		t.Fatalf("request attempts=%d want 1", requests[0].Attempts)
	}
}

// TestTasksServiceDoesNotClaimRequest asserts the task path leaves a fresh
// request pending when no provider exists: under the old self-claim the task
// grabbed the lease itself, starving every polling provider until lease expiry.
func TestTasksServiceDoesNotClaimRequest(t *testing.T) {
	taskCtx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	tracer := &recordingTracer{}
	toolService := NewToolService(tracer, nil)
	machineService := NewMachinesService(context.Background(), toolService, tracer, nil)
	requestService := NewRequestsService(context.Background(), toolService, machineService, tracer, nil)
	tasksService := NewTasksService(taskCtx, toolService, machineService, requestService, tracer, nil)

	const sessionID = "session-task-pending"
	const machineID = "machine-task-pending"

	_, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, "")
	if err != nil {
		t.Fatalf("register machine: %v", err)
	}

	task, err := tasksService.CreateTask(sessionID, "echo", `{"message":"nobody home"}`, "")
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	requestID := waitForActiveTaskRequestID(t, tasksService, task.ID, 2*time.Second)
	request, err := requestService.GetRequestByID(sessionID, requestID)
	if err != nil {
		t.Fatalf("get task request: %v", err)
	}
	if request.Status != model.RequestStatusPending {
		t.Fatalf("task request status=%s want pending (self-claim = bug)", request.Status)
	}
	if request.LeasedBy != "" {
		t.Fatalf("task request leasedBy=%q want empty (no provider has claimed it)", request.LeasedBy)
	}
	// A provider must still be able to claim it: the task never held the lease.
	claimed, err := requestService.ClaimPendingRequest(sessionID, machineID, []string{"echo"})
	if err != nil || claimed == nil || claimed.ID != requestID {
		t.Fatalf("provider claim after task create: claimed=%v err=%v", claimed, err)
	}
}

// TestTasksServiceTaskResultIsJSON pins the durable task result as JSON: a
// structured tool result must land in tasks.result as the JSON encoding of
// the submitted value — the same bytes the ExecuteTool long-poll serves —
// not Go's %v rendering of the decoded map.
func TestTasksServiceTaskResultIsJSON(t *testing.T) {
	taskCtx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	tracer := &recordingTracer{}
	toolService := NewToolService(tracer, nil)
	machineService := NewMachinesService(context.Background(), toolService, tracer, nil)
	requestService := NewRequestsService(context.Background(), toolService, machineService, tracer, nil)
	tasksService := NewTasksService(taskCtx, toolService, machineService, requestService, tracer, nil)

	const sessionID = "session-task-result-json"
	const machineID = "machine-task-result-json"

	_, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "report", "report tool", `{"type":"object"}`, nil, nil),
	}, "")
	if err != nil {
		t.Fatalf("register machine: %v", err)
	}

	// Provider submits a structured result, exactly as the SDK runtimes do
	// after json.dumps-ing the tool's return value.
	go func() {
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			claimed, err := requestService.ClaimPendingRequest(sessionID, machineID, []string{"report"})
			if err != nil || claimed == nil {
				time.Sleep(20 * time.Millisecond)
				continue
			}
			if _, err := requestService.UpdateRequest(sessionID, claimed.ID, machineID, claimed.LeaseEpoch, model.RequestStatusRunning, nil, ""); err != nil {
				t.Errorf("provider running update: %v", err)
				return
			}
			_ = requestService.SubmitRequestResult(sessionID, claimed.ID, machineID, claimed.LeaseEpoch,
				map[string]interface{}{"items": []interface{}{"a", "b"}, "count": 2}, model.ResultTypeResolution, nil)
			return
		}
	}()

	task, err := tasksService.CreateTask(sessionID, "report", `{}`, "")
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	waitForTaskStatus(t, tasksService, sessionID, task.ID, model.StatusCompleted, 10*time.Second)

	completed, err := tasksService.GetTaskByID(sessionID, task.ID)
	if err != nil {
		t.Fatalf("get completed task: %v", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(completed.Result), &decoded); err != nil {
		t.Fatalf("task result is not JSON: %q (err %v)", completed.Result, err)
	}
	if decoded["count"] != float64(2) {
		t.Fatalf("task result count = %v, want 2", decoded["count"])
	}
	items, ok := decoded["items"].([]interface{})
	if !ok || len(items) != 2 || items[0] != "a" || items[1] != "b" {
		t.Fatalf("task result items = %v, want [a b]", decoded["items"])
	}
}

// TestTasksServiceTaskResultStringRoundTrip pins the string-result identity:
// a tool returning a plain string is submitted by the SDKs as a JSON string,
// so the task result must carry the same JSON encoding the ExecuteTool
// long-poll serves, not the raw %v rendering.
func TestTasksServiceTaskResultStringRoundTrip(t *testing.T) {
	taskCtx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	tracer := &recordingTracer{}
	toolService := NewToolService(tracer, nil)
	machineService := NewMachinesService(context.Background(), toolService, tracer, nil)
	requestService := NewRequestsService(context.Background(), toolService, machineService, tracer, nil)
	tasksService := NewTasksService(taskCtx, toolService, machineService, requestService, tracer, nil)

	const sessionID = "session-task-result-string"
	const machineID = "machine-task-result-string"

	_, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "greet", "greet tool", `{"type":"object"}`, nil, nil),
	}, "")
	if err != nil {
		t.Fatalf("register machine: %v", err)
	}

	go func() {
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			claimed, err := requestService.ClaimPendingRequest(sessionID, machineID, []string{"greet"})
			if err != nil || claimed == nil {
				time.Sleep(20 * time.Millisecond)
				continue
			}
			if _, err := requestService.UpdateRequest(sessionID, claimed.ID, machineID, claimed.LeaseEpoch, model.RequestStatusRunning, nil, ""); err != nil {
				t.Errorf("provider running update: %v", err)
				return
			}
			_ = requestService.SubmitRequestResult(sessionID, claimed.ID, machineID, claimed.LeaseEpoch,
				"hello world", model.ResultTypeResolution, nil)
			return
		}
	}()

	task, err := tasksService.CreateTask(sessionID, "greet", `{}`, "")
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	waitForTaskStatus(t, tasksService, sessionID, task.ID, model.StatusCompleted, 10*time.Second)

	completed, err := tasksService.GetTaskByID(sessionID, task.ID)
	if err != nil {
		t.Fatalf("get completed task: %v", err)
	}
	// json.dumps("hello world") on the provider side, json.Marshal on the
	// waiter side: identical bytes.
	if completed.Result != `"hello world"` {
		t.Fatalf("task result = %q, want the JSON encoding %q", completed.Result, `"hello world"`)
	}
}

// TestTasksServiceTaskResultIsJSONPersisted repeats the structured-result
// check against the durable row: tasks.result in the store must hold the
// same JSON, since the task API serves the persisted value after restarts.
func TestTasksServiceTaskResultIsJSONPersisted(t *testing.T) {
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

	taskCtx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	tracer := &recordingTracer{}
	sessionService := NewSessionsService(tracer, store)
	toolService := NewToolService(tracer, store)
	machineService := NewMachinesService(context.Background(), toolService, tracer, store)
	requestService := NewRequestsService(context.Background(), toolService, machineService, tracer, store)
	tasksService := NewTasksService(taskCtx, toolService, machineService, requestService, tracer, store)

	session, err := sessionService.CreateSession("task-result-user", "Task Result Persistence", "", "", "")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	t.Cleanup(func() {
		_ = sessionService.DeleteSession(session.ID)
	})
	sessionID := session.ID
	const machineID = "machine-task-result-persisted"

	_, err = machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "report", "report tool", `{"type":"object"}`, nil, nil),
	}, "")
	if err != nil {
		t.Fatalf("register machine: %v", err)
	}

	go func() {
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			claimed, err := requestService.ClaimPendingRequest(sessionID, machineID, []string{"report"})
			if err != nil || claimed == nil {
				time.Sleep(20 * time.Millisecond)
				continue
			}
			if _, err := requestService.UpdateRequest(sessionID, claimed.ID, machineID, claimed.LeaseEpoch, model.RequestStatusRunning, nil, ""); err != nil {
				t.Errorf("provider running update: %v", err)
				return
			}
			_ = requestService.SubmitRequestResult(sessionID, claimed.ID, machineID, claimed.LeaseEpoch,
				map[string]interface{}{"answer": 42}, model.ResultTypeResolution, nil)
			return
		}
	}()

	task, err := tasksService.CreateTask(sessionID, "report", `{}`, "")
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	waitForTaskStatus(t, tasksService, sessionID, task.ID, model.StatusCompleted, 15*time.Second)

	readCtx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
	defer cancel()
	stored, err := store.GetTaskByID(readCtx, task.ID)
	if err != nil {
		t.Fatalf("read stored task: %v", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(stored.Result), &decoded); err != nil {
		t.Fatalf("stored task result is not JSON: %q (err %v)", stored.Result, err)
	}
	if decoded["answer"] != float64(42) {
		t.Fatalf("stored task result answer = %v, want 42", decoded["answer"])
	}
}
