package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
	"toolplane/pkg/trace"
)

var (
	errTaskCancelled = errors.New("task cancelled")
	errTaskTimedOut  = errors.New("task timed out")
)

type taskExecutionState struct {
	requestID string
	cancel    context.CancelFunc
}

// TasksService handles task scheduling and execution
type TasksService struct {
	ctx context.Context // Context for cancellation

	// In-memory storage for tasks
	tasks      map[string]*model.Task // map[taskID]Task
	tasksMutex sync.RWMutex
	executions map[string]*taskExecutionState
	execMutex  sync.RWMutex

	// References to other services
	toolService     *ToolService
	machinesService *MachinesService
	requestsService *RequestsService
	tracer          trace.SessionTracer
	store           *storage.Store
}

// NewTasksService creates a new tasks service
func NewTasksService(ctx context.Context, toolService *ToolService, machinesService *MachinesService, requestsService *RequestsService, tracer trace.SessionTracer, store *storage.Store) *TasksService {
	if tracer == nil {
		tracer = trace.NopTracer()
	}

	service := &TasksService{
		ctx:             ctx,
		tasks:           make(map[string]*model.Task),
		executions:      make(map[string]*taskExecutionState),
		toolService:     toolService,
		machinesService: machinesService,
		requestsService: requestsService,
		tracer:          tracer,
		store:           store,
	}
	if store != nil {
		loadCtx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if tasks, err := store.AllTasks(loadCtx); err != nil {
			log.Printf("task persistence load failed: %v", err)
		} else {
			for _, task := range tasks {
				service.tasks[task.ID] = task
			}
		}
	}
	return service
}

// CreateTask creates a new task and schedules it for execution
func (s *TasksService) CreateTask(sessionID, toolName, input string) (*model.Task, error) {
	// Create the task
	task := model.NewTask(sessionID, toolName, input)

	// Store the task
	s.tasksMutex.Lock()
	s.tasks[task.ID] = task
	s.tasksMutex.Unlock()

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if err := s.store.SaveTask(ctx, task); err != nil {
			log.Printf("persist task create failed: %v", err)
		}
	}

	s.recordTaskEvent(task, trace.EventTaskCreated, nil)

	// Schedule task for execution asynchronously
	go s.executeTask(task)

	return task, nil
}

// GetTask gets a task by ID
func (s *TasksService) GetTask(taskID string) (*model.Task, error) {
	s.tasksMutex.RLock()
	defer s.tasksMutex.RUnlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("task with ID %s not found", taskID)
	}

	return task, nil
}

// GetTaskByID gets a task by ID for a specific session
func (s *TasksService) GetTaskByID(sessionID, taskID string) (*model.Task, error) {
	s.tasksMutex.RLock()
	defer s.tasksMutex.RUnlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("task with ID %s not found", taskID)
	}

	// Verify the task belongs to the specified session
	if task.SessionID != sessionID {
		return nil, fmt.Errorf("task with ID %s not found in session %s", taskID, sessionID)
	}

	return task, nil
}

// ListTasks lists all tasks for a session
func (s *TasksService) ListTasks(sessionID string) ([]*model.Task, error) {
	s.tasksMutex.RLock()
	defer s.tasksMutex.RUnlock()

	var tasks []*model.Task
	for _, task := range s.tasks {
		if task.SessionID == sessionID {
			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}

// CancelTask cancels a running task
func (s *TasksService) CancelTask(sessionID, taskID string) error {
	s.tasksMutex.Lock()

	task, ok := s.tasks[taskID]
	if !ok {
		s.tasksMutex.Unlock()
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	// Verify the task belongs to the specified session
	if task.SessionID != sessionID {
		s.tasksMutex.Unlock()
		return fmt.Errorf("task with ID %s not found in session %s", taskID, sessionID)
	}

	if task.Status != model.StatusPending && task.Status != model.StatusRunning {
		s.tasksMutex.Unlock()
		return fmt.Errorf("cannot cancel task with status %s", task.Status)
	}

	task.Status = model.StatusCancelled
	task.Error = "Task cancelled by user"
	task.Result = ""
	task.ResultType = ""
	task.UpdatedAt = time.Now()
	completedAt := time.Now()
	task.CompletedAt = &completedAt
	task.NextAttemptAt = nil
	task.DeadLetter = false
	task.LastError = ""
	s.tasksMutex.Unlock()

	s.persistTask(task)

	requestID, cancel := s.taskExecutionSnapshot(taskID)
	s.recordTaskEvent(task, trace.EventTaskCancelled, map[string]any{
		"requestID": requestID,
	})
	if requestID != "" {
		_ = s.requestsService.CancelRequest(sessionID, requestID)
	}
	if cancel != nil {
		cancel()
	}

	return nil
}

// executeTask executes a task by finding a suitable machine and sending a request
func (s *TasksService) executeTask(task *model.Task) {
	taskCtx, cancel := context.WithCancel(s.ctx)
	s.setTaskExecution(task.ID, cancel)
	defer func() {
		s.clearTaskExecution(task.ID)
		cancel()
	}()

	for {
		if err := taskCtx.Err(); err != nil {
			if s.isTaskCancelled(task.ID) {
				return
			}
			s.updateTaskWithError(task, "task cancelled due to server shutdown")
			return
		}
		if s.isTaskCancelled(task.ID) {
			return
		}

		err := s.runTaskAttempt(taskCtx, task)
		if err == nil {
			return
		}
		if errors.Is(err, errTaskCancelled) {
			return
		}
		if errors.Is(err, errTaskTimedOut) {
			s.updateTaskWithError(task, err.Error())
			return
		}
		if errors.Is(err, context.Canceled) {
			if s.isTaskCancelled(task.ID) {
				return
			}
			s.updateTaskWithError(task, "task cancelled due to server shutdown")
			return
		}
		if s.isTaskCancelled(task.ID) {
			return
		}

		s.tasksMutex.Lock()
		if task.Status == model.StatusCancelled {
			s.tasksMutex.Unlock()
			return
		}
		task.Error = err.Error()
		task.Result = ""
		task.ResultType = ""
		task.Status = model.StatusPending
		task.UpdatedAt = time.Now()
		task.CompletedAt = nil
		s.tasksMutex.Unlock()

		if task.Attempts >= task.MaxAttempts {
			s.tasksMutex.Lock()
			task.MarkDeadLetter(err.Error())
			completedAt := time.Now()
			task.CompletedAt = &completedAt
			s.tasksMutex.Unlock()
			s.persistTask(task)
			s.recordTaskEvent(task, trace.EventTaskDeadLettered, map[string]any{
				"reason": err.Error(),
			})
			return
		}

		delay := time.Duration(task.BackoffSeconds*task.Attempts) * time.Second
		next := time.Now().Add(delay)
		s.tasksMutex.Lock()
		if task.Status == model.StatusCancelled {
			s.tasksMutex.Unlock()
			return
		}
		task.NextAttemptAt = &next
		s.tasksMutex.Unlock()
		s.persistTask(task)
		s.recordTaskEvent(task, trace.EventTaskRetryScheduled, map[string]any{
			"reason":        err.Error(),
			"nextAttemptAt": next,
		})

		if delay <= 0 {
			continue
		}

		select {
		case <-taskCtx.Done():
			if s.isTaskCancelled(task.ID) {
				return
			}
			s.updateTaskWithError(task, "task cancelled due to server shutdown")
			return
		case <-time.After(delay):
		}
	}
}

func (s *TasksService) runTaskAttempt(taskCtx context.Context, task *model.Task) error {
	if s.isTaskCancelled(task.ID) {
		return errTaskCancelled
	}

	s.tasksMutex.Lock()
	if task.Status == model.StatusCancelled {
		s.tasksMutex.Unlock()
		return errTaskCancelled
	}
	task.Attempts++
	task.Status = model.StatusRunning
	task.Error = ""
	task.Result = ""
	task.ResultType = ""
	task.UpdatedAt = time.Now()
	task.CompletedAt = nil
	task.NextAttemptAt = nil
	s.tasksMutex.Unlock()
	s.persistTask(task)

	machine, err := s.selectMachine(task.SessionID, task.ToolName)
	if err != nil {
		return err
	}
	if err := taskCtx.Err(); err != nil {
		if s.isTaskCancelled(task.ID) {
			return errTaskCancelled
		}
		return err
	}

	request, err := s.requestsService.CreateRequest(task.SessionID, task.ToolName, task.Input)
	if err != nil {
		return err
	}
	s.setTaskRequestID(task.ID, request.ID)
	defer s.setTaskRequestID(task.ID, "")
	s.recordTaskEvent(task, trace.EventTaskExecutionStarted, map[string]any{
		"machineID": machine.ID,
		"requestID": request.ID,
	})

	attemptCtx := taskCtx
	cancel := func() {}
	if task.TimeoutSeconds > 0 {
		attemptCtx, cancel = context.WithTimeout(taskCtx, time.Duration(task.TimeoutSeconds)*time.Second)
	}
	defer cancel()

	result, err := s.requestsService.ExecuteRequest(attemptCtx, task.SessionID, machine.ID, request.ID)
	if err != nil {
		if s.isTaskCancelled(task.ID) {
			return errTaskCancelled
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("%w after %ds", errTaskTimedOut, task.TimeoutSeconds)
		}
		return err
	}
	if s.isTaskCancelled(task.ID) {
		return errTaskCancelled
	}

	s.tasksMutex.Lock()
	if task.Status == model.StatusCancelled {
		s.tasksMutex.Unlock()
		return errTaskCancelled
	}
	task.Status = model.StatusCompleted
	task.Result = result.Result
	task.ResultType = result.ResultType
	task.Error = ""
	completedAt := time.Now()
	task.CompletedAt = &completedAt
	task.UpdatedAt = completedAt
	task.NextAttemptAt = nil
	task.DeadLetter = false
	task.LastError = ""
	s.tasksMutex.Unlock()
	s.persistTask(task)
	s.recordTaskEvent(task, trace.EventTaskExecutionCompleted, map[string]any{
		"requestID":  request.ID,
		"resultType": result.ResultType,
	})
	return nil
}

func (s *TasksService) selectMachine(sessionID, toolName string) (*model.Machine, error) {
	machines, err := s.machinesService.FindMachinesWithTool(sessionID, toolName)
	if err != nil {
		return nil, err
	}
	if len(machines) == 0 {
		return nil, fmt.Errorf("No machines available with tool %s", toolName)
	}
	machines = s.machinesService.OrderMachinesByLoad(sessionID, machines)
	for _, machine := range machines {
		load, capacity := s.machinesService.MachineLoadInfo(sessionID, machine.ID)
		if load < capacity {
			return machine, nil
		}
	}
	return nil, fmt.Errorf("all machines at capacity for tool %s", toolName)
}

func (s *TasksService) persistTask(task *model.Task) {
	if s.store == nil || task == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
	if err := s.store.SaveTask(ctx, task); err != nil {
		log.Printf("persist task save failed: %v", err)
	}
	cancel()
}

// updateTaskWithError updates a task with an error status
func (s *TasksService) updateTaskWithError(task *model.Task, errorMsg string) {
	s.tasksMutex.Lock()
	defer s.tasksMutex.Unlock()
	if task.Status == model.StatusCancelled {
		return
	}

	task.Status = model.StatusFailed
	task.Error = errorMsg
	task.UpdatedAt = time.Now()
	completedAt := time.Now()
	task.CompletedAt = &completedAt
	task.NextAttemptAt = nil

	s.persistTask(task)
	s.recordTaskEvent(task, trace.EventTaskExecutionFailed, map[string]any{
		"reason": errorMsg,
	})
}

func (s *TasksService) setTaskExecution(taskID string, cancel context.CancelFunc) {
	s.execMutex.Lock()
	defer s.execMutex.Unlock()
	s.executions[taskID] = &taskExecutionState{cancel: cancel}
}

func (s *TasksService) setTaskRequestID(taskID, requestID string) {
	s.execMutex.Lock()
	defer s.execMutex.Unlock()
	if execution, ok := s.executions[taskID]; ok {
		execution.requestID = requestID
	}
}

func (s *TasksService) taskExecutionSnapshot(taskID string) (string, context.CancelFunc) {
	s.execMutex.RLock()
	defer s.execMutex.RUnlock()
	if execution, ok := s.executions[taskID]; ok {
		return execution.requestID, execution.cancel
	}
	return "", nil
}

func (s *TasksService) clearTaskExecution(taskID string) {
	s.execMutex.Lock()
	defer s.execMutex.Unlock()
	delete(s.executions, taskID)
}

func (s *TasksService) isTaskCancelled(taskID string) bool {
	s.tasksMutex.RLock()
	defer s.tasksMutex.RUnlock()
	task, ok := s.tasks[taskID]
	return ok && task.Status == model.StatusCancelled
}

func (s *TasksService) recordTaskEvent(task *model.Task, event trace.SessionEventType, metadata map[string]any) {
	if s == nil || s.tracer == nil || task == nil {
		return
	}
	timestamp := task.UpdatedAt
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	eventMetadata := map[string]any{
		"taskID":     task.ID,
		"toolName":   task.ToolName,
		"status":     task.Status,
		"attempts":   task.Attempts,
		"deadLetter": task.DeadLetter,
	}
	for key, value := range metadata {
		eventMetadata[key] = value
	}
	s.tracer.Record(trace.SessionEvent{
		SessionID: task.SessionID,
		RequestID: requestIDFromMetadata(metadata),
		TaskID:    task.ID,
		Event:     event,
		Timestamp: timestamp,
		Metadata:  eventMetadata,
	})
}

// TaskMetricsSnapshot returns current task counts for observability scrapes.
func (s *TasksService) TaskMetricsSnapshot() (pending, running, completed, failed, cancelled, deadLetter int) {
	s.tasksMutex.RLock()
	defer s.tasksMutex.RUnlock()

	for _, task := range s.tasks {
		if task == nil {
			continue
		}
		switch task.Status {
		case model.StatusPending:
			pending++
		case model.StatusRunning:
			running++
		case model.StatusCompleted:
			completed++
		case model.StatusFailed:
			failed++
		case model.StatusCancelled:
			cancelled++
		}
		if task.DeadLetter {
			deadLetter++
		}
	}

	return pending, running, completed, failed, cancelled, deadLetter
}

func requestIDFromMetadata(metadata map[string]any) string {
	if len(metadata) == 0 {
		return ""
	}
	value, ok := metadata["requestID"]
	if !ok {
		return ""
	}
	requestID, ok := value.(string)
	if !ok {
		return ""
	}
	return requestID
}

// CleanupOldTasks removes tasks older than a certain age
func (s *TasksService) CleanupOldTasks(maxAge time.Duration) {
	s.tasksMutex.Lock()
	defer s.tasksMutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for id, task := range s.tasks {
		if task.CreatedAt.Before(cutoff) {
			delete(s.tasks, id)
			if s.store != nil {
				ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
				if err := s.store.DeleteTask(ctx, id); err != nil {
					log.Printf("persist task cleanup delete failed: %v", err)
				}
				cancel()
			}
		}
	}
}
