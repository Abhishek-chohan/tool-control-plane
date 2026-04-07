package service

import (
	"context"
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
	machineService := NewMachinesService(toolService, tracer, nil)
	requestService := NewRequestsService(toolService, machineService, tracer, nil)
	tasksService := NewTasksService(taskCtx, toolService, machineService, requestService, tracer, nil)

	const sessionID = "session-task-cancel"
	const machineID = "machine-task-cancel"

	_, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	})
	if err != nil {
		t.Fatalf("register machine: %v", err)
	}

	task, err := tasksService.CreateTask(sessionID, "echo", `{"message":"cancel me"}`)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	requestID := waitForActiveTaskRequestID(t, tasksService, task.ID, time.Second)
	waitForRequestStatus(t, requestService, sessionID, requestID, model.RequestStatusRunning, time.Second)

	if err := tasksService.CancelTask(sessionID, task.ID); err != nil {
		t.Fatalf("cancel task: %v", err)
	}

	waitForTaskStatus(t, tasksService, sessionID, task.ID, model.StatusCancelled, time.Second)
	waitForRequestStatus(t, requestService, sessionID, requestID, model.RequestStatusFailed, time.Second)

	if err := requestService.SubmitRequestResult(sessionID, requestID, map[string]string{"echo": "late"}, model.ResultTypeResolution, nil); err == nil {
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
	machineService := NewMachinesService(toolService, tracer, nil)
	requestService := NewRequestsService(toolService, machineService, tracer, nil)
	tasksService := NewTasksService(taskCtx, toolService, machineService, requestService, tracer, nil)

	const sessionID = "session-task-timeout"
	const machineID = "machine-task-timeout"

	_, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	})
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
	waitForRequestStatus(t, requestService, sessionID, requestID, model.RequestStatusRunning, time.Second)
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

	requests, err := requestService.ListRequests(sessionID, "", "echo", 10, 0)
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
