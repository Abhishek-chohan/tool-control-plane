package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
	"toolplane/pkg/trace"
)

var (
	errTaskCancelled = errors.New("task cancelled")
	errTaskTimedOut  = errors.New("task timed out")
)

// Task-adoption fencing intervals. A task's executing instance owns it for
// taskAdoptionLeaseTTL and renews at taskAdoptionRenewInterval; any
// instance's sweep can claim a task whose lease expired. Due tasks are
// re-dispatched by whichever instance's sweep sees them first, so retries
// distribute across replicas instead of piling onto the creator.
const (
	taskAdoptionLeaseTTL      = 30 * time.Second
	taskAdoptionRenewInterval = 10 * time.Second
	taskAdoptionSweepInterval = 5 * time.Second
	// taskAdoptionSweepBatch bounds the work one sweep tick takes on.
	taskAdoptionSweepBatch = 32
	taskCleanupInterval    = 30 * time.Minute
	taskRetentionAge       = 24 * time.Hour
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

	// instanceID identifies this process to the adoption fence: task
	// ownership in the store is recorded per instance so replicas cannot
	// execute the same task, and a lost lease is detected on renewal.
	instanceID string

	// References to other services
	toolService     *ToolService
	machinesService *MachinesService
	requestsService *RequestsService
	tracer          trace.SessionTracer
	store           storage.Storer
}

// NewTasksService creates a new tasks service
func NewTasksService(ctx context.Context, toolService *ToolService, machinesService *MachinesService, requestsService *RequestsService, tracer trace.SessionTracer, store storage.Storer) *TasksService {
	if tracer == nil {
		tracer = trace.NopTracer()
	}

	service := &TasksService{
		ctx:             ctx,
		instanceID:      "toolplane-task-" + uuid.New().String(),
		tasks:           make(map[string]*model.Task),
		executions:      make(map[string]*taskExecutionState),
		toolService:     toolService,
		machinesService: machinesService,
		requestsService: requestsService,
		tracer:          tracer,
		store:           store,
	}
	if store != nil {
		loadCtx, cancel := context.WithTimeout(ctx, defaultPersistenceTimeout)
		defer cancel()
		if tasks, err := store.AllTasks(loadCtx); err != nil {
			log.Printf("task persistence load failed: %v", err)
		} else {
			for _, task := range tasks {
				service.tasks[task.ID] = task
			}
		}

		// Re-adopt in-flight work: launch the execution loop for any task that
		// was pending or running when this instance started (e.g. after a
		// restart) and is due. Without this, a non-terminal task loaded above
		// would stall forever.
		//
		// Multi-instance safety: ownership is acquired through the fenced
		// adoption claim (unowned or lease-expired wins), so two replicas
		// starting simultaneously cannot both adopt — and a replica whose
		// adoption lease lapses loses the task on renewal.
		adoptCtx, adoptCancel := context.WithTimeout(ctx, defaultPersistenceTimeout)
		if nonTerminal, err := store.FindNonTerminalTasks(adoptCtx); err != nil {
			log.Printf("task re-adoption scan failed: %v", err)
		} else {
			for _, task := range nonTerminal {
				if !service.taskAttemptDue(task) {
					continue
				}
				// Claim this task for adoption; only one instance wins.
				claimed, ok, claimErr := store.ClaimTaskForAdoption(adoptCtx, task.ID, service.instanceID, taskAdoptionLeaseTTL)
				if claimErr != nil {
					log.Printf("task re-adoption claim %s failed: %v", task.ID, claimErr)
					continue
				}
				if !ok {
					// Another instance owns this task.
					continue
				}
				service.recordTaskEvent(claimed, trace.EventTaskRetryScheduled, map[string]any{
					"reason": "instance restart re-adoption",
				})
				go service.executeTask(claimed)
			}
		}
		adoptCancel()

		// The adoption sweep is the task-side twin of the request reaper: it
		// claims due tasks nobody owns (new enqueues, retries released by
		// their scheduler, tasks orphaned by a dead instance) so work
		// distributes across replicas instead of depending on the creator.
		go service.adoptDueTasks()
		// Terminal tasks are retained for inspection, then pruned.
		go service.cleanupTerminalTasks()
	}
	return service
}

// taskAttemptDue reports whether the task's next attempt is scheduled for
// now: tasks waiting out a backoff are not adoptable before their slot.
func (s *TasksService) taskAttemptDue(task *model.Task) bool {
	if task == nil || task.DeadLetter {
		return false
	}
	switch task.Status {
	case model.StatusCompleted, model.StatusFailed, model.StatusCancelled:
		return false
	}
	if task.NextAttemptAt == nil {
		return true
	}
	return !task.NextAttemptAt.After(time.Now())
}

// dispatchTask acquires ownership of a freshly enqueued task on this
// instance (it is unowned, so the claim is uncontended) and runs it. When
// the claim unexpectedly loses, the sweep picks the task up within one
// interval.
func (s *TasksService) dispatchTask(taskID string) {
	ctx, cancel := context.WithTimeout(s.ctx, defaultPersistenceTimeout)
	defer cancel()
	claimed, ok, err := s.store.ClaimTaskForAdoption(ctx, taskID, s.instanceID, taskAdoptionLeaseTTL)
	if err != nil {
		log.Printf("task dispatch claim %s failed: %v", taskID, err)
		return
	}
	if !ok {
		return
	}
	s.executeTask(claimed)
}

// adoptDueTasks periodically claims due tasks that no live instance owns.
// New tasks dispatch immediately on their creating instance; this loop is
// what recovers them (and scheduled retries) when that instance dies or
// releases the task for retry.
func (s *TasksService) adoptDueTasks() {
	ticker := time.NewTicker(taskAdoptionSweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
		}

		ctx, cancel := context.WithTimeout(s.ctx, defaultPersistenceTimeout)
		// The targeted adoptable query answers due-ness and ownership in the
		// store: a task this instance is executing holds a fresh adoption
		// lease (renewed every taskAdoptionRenewInterval), so it is filtered
		// out server-side and no local execution map is consulted.
		candidates, err := s.store.FindAdoptableTasks(ctx, time.Now(), taskAdoptionLeaseTTL, taskAdoptionSweepBatch)
		if err != nil {
			log.Printf("task adoption sweep scan failed: %v", err)
			cancel()
			continue
		}
		for _, task := range candidates {
			claimed, ok, claimErr := s.store.ClaimTaskForAdoption(ctx, task.ID, s.instanceID, taskAdoptionLeaseTTL)
			if claimErr != nil {
				log.Printf("task adoption sweep claim %s failed: %v", task.ID, claimErr)
				continue
			}
			if !ok {
				continue
			}
			go s.executeTask(claimed)
		}
		cancel()
	}
}

// cleanupTerminalTasks prunes old terminal tasks on a schedule so the
// tasks table does not grow without bound.
func (s *TasksService) cleanupTerminalTasks() {
	ticker := time.NewTicker(taskCleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.CleanupOldTasks(taskRetentionAge)
		}
	}
}

// CreateTask creates a new task and schedules it for execution. With a
// non-empty idempotencyKey, a retry returns the existing task without
// scheduling it again.
func (s *TasksService) CreateTask(sessionID, toolName, input, idempotencyKey string) (*model.Task, error) {
	if idempotencyKey != "" {
		if existing := s.findTaskByIdempotencyKey(sessionID, idempotencyKey); existing != nil {
			return existing, nil
		}
	}

	// Create the task
	task := model.NewTask(sessionID, toolName, input)
	task.IdempotencyKey = idempotencyKey

	// Persist-first so the store stays authoritative: a persist failure
	// surfaces instead of leaving the local cache divergent, and a
	// cross-replica key race (unique violation) returns the winner's task
	// without scheduling this duplicate.
	if s.store != nil {
		ctx, cancel := context.WithTimeout(s.ctx, defaultPersistenceTimeout)
		defer cancel()
		if err := s.store.SaveTask(ctx, task); err != nil {
			if idempotencyKey != "" && storage.IsUniqueViolation(err) {
				if existing := s.findTaskByIdempotencyKey(sessionID, idempotencyKey); existing != nil {
					return existing, nil
				}
			}
			return nil, fmt.Errorf("persist task create failed: %w", err)
		}
	}

	s.tasksMutex.Lock()
	s.tasks[task.ID] = task
	s.tasksMutex.Unlock()

	s.recordTaskEvent(task, trace.EventTaskCreated, nil)

	// Enqueue, don't self-run: dispatch acquires ownership (uncontended for
	// a fresh task) so execution starts immediately here, while the sweep on
	// any instance recovers the task if this one dies before finishing. In
	// the in-memory dev mode there is no fence, so the loop runs directly.
	if s.store != nil {
		go s.dispatchTask(task.ID)
	} else {
		go s.executeTask(task)
	}

	return task, nil
}

// findTaskByIdempotencyKey resolves a dedup key to the session's task: the
// local cache first, then the store, so a key minted on another replica also
// dedups.
func (s *TasksService) findTaskByIdempotencyKey(sessionID, idempotencyKey string) *model.Task {
	s.tasksMutex.RLock()
	for _, task := range s.tasks {
		if task != nil && task.SessionID == sessionID && task.IdempotencyKey == idempotencyKey {
			cloned := task.Clone()
			s.tasksMutex.RUnlock()
			return cloned
		}
	}
	s.tasksMutex.RUnlock()

	if s.store == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(s.ctx, defaultPersistenceTimeout)
	defer cancel()
	found, err := s.store.GetTaskByIdempotencyKey(ctx, sessionID, idempotencyKey)
	if err != nil || found == nil {
		return nil
	}
	found = found.Clone()
	s.tasksMutex.Lock()
	if _, exists := s.tasks[found.ID]; !exists {
		s.tasks[found.ID] = found.Clone()
	}
	s.tasksMutex.Unlock()
	return found
}

// GetTask gets a task by ID
func (s *TasksService) GetTask(taskID string) (*model.Task, error) {
	s.tasksMutex.RLock()
	defer s.tasksMutex.RUnlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return nil, wrapf(ErrNotFound, "task with ID %s not found", taskID)
	}

	// Hand out a clone: callers read outside the lock while task
	// execution mutates the cached original.
	return task.Clone(), nil
}

// GetTaskByID gets a task by ID for a specific session
func (s *TasksService) GetTaskByID(sessionID, taskID string) (*model.Task, error) {
	s.tasksMutex.RLock()
	defer s.tasksMutex.RUnlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return nil, wrapf(ErrNotFound, "task with ID %s not found", taskID)
	}

	// Verify the task belongs to the specified session
	if task.SessionID != sessionID {
		return nil, wrapf(ErrNotFound, "task with ID %s not found in session %s", taskID, sessionID)
	}

	return task.Clone(), nil
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

	cloned := make([]*model.Task, 0, len(tasks))
	for _, task := range tasks {
		cloned = append(cloned, task.Clone())
	}
	return cloned, nil
}

// CancelTask cancels a running task
func (s *TasksService) CancelTask(sessionID, taskID string) error {
	s.tasksMutex.Lock()

	task, ok := s.tasks[taskID]
	if !ok {
		s.tasksMutex.Unlock()
		return wrapf(ErrNotFound, "task with ID %s not found", taskID)
	}

	// Verify the task belongs to the specified session
	if task.SessionID != sessionID {
		s.tasksMutex.Unlock()
		return wrapf(ErrNotFound, "task with ID %s not found in session %s", taskID, sessionID)
	}

	if task.Status != model.StatusPending && task.Status != model.StatusRunning {
		s.tasksMutex.Unlock()
		return wrapf(ErrTaskNotCancellable, "cannot cancel task with status %s", task.Status)
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

	if err := s.persistTask(task); err != nil {
		// Cancel is terminal: an undurable cancel row means a restart
		// re-adopts a task the user believes is gone.
		log.Printf("task %s cancel not durable; may re-adopt after restart: %v", task.ID, err)
	}

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

// executeTask executes a task by finding a suitable machine and sending a request.
// With a store configured, this instance holds the task's adoption lease for
// the duration: a renewal ticker keeps it alive, and losing it (another
// instance claimed the expired task) cancels the attempt so the task is
// never executed twice. Retries are released for any instance's sweep to
// pick up when due.
func (s *TasksService) executeTask(task *model.Task) {
	// Callers hand over the store-claimed copy; make it the cached
	// authoritative object so status transitions are visible to readers
	// instead of mutating a detached clone next to a stale cache entry.
	s.tasksMutex.Lock()
	s.tasks[task.ID] = task
	s.tasksMutex.Unlock()

	taskCtx, cancel := context.WithCancel(s.ctx)
	s.setTaskExecution(task.ID, cancel)
	defer func() {
		s.clearTaskExecution(task.ID)
		cancel()
	}()

	// Keep the adoption lease alive while this instance executes. Renewal
	// failure means ownership was lost — stop rather than double-execute.
	var leaseLost atomic.Bool
	if s.store != nil {
		renewTicker := time.NewTicker(taskAdoptionRenewInterval)
		go func() {
			defer renewTicker.Stop()
			for {
				select {
				case <-taskCtx.Done():
					return
				case <-renewTicker.C:
					ctx, renewCancel := context.WithTimeout(s.ctx, defaultPersistenceTimeout)
					renewed, err := s.store.RenewTaskAdoption(ctx, task.ID, s.instanceID)
					renewCancel()
					if err != nil {
						// Store errors keep the lease (it may still be live);
						// the next tick retries.
						continue
					}
					if !renewed {
						log.Printf("task %s adoption lease lost; aborting local execution", task.ID)
						leaseLost.Store(true)
						cancel()
						return
					}
				}
			}
		}()
	}
	// releaseTask drops the adoption lease once, when this loop is done
	// holding the task.
	releaseOnce := sync.Once{}
	releaseTask := func() {
		if s.store == nil {
			return
		}
		releaseOnce.Do(func() {
			ctx, releaseCancel := context.WithTimeout(context.WithoutCancel(s.ctx), defaultPersistenceTimeout)
			defer releaseCancel()
			if err := s.store.ReleaseTaskAdoption(ctx, task.ID, s.instanceID); err != nil {
				log.Printf("task adoption release %s failed: %v", task.ID, err)
			}
		})
	}
	defer releaseTask()

	for {
		if err := taskCtx.Err(); err != nil {
			if s.isTaskCancelled(task.ID) {
				return
			}
			if leaseLost.Load() {
				// Ownership moved elsewhere; do not write a failure state
				// that would race with the new owner.
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
			if leaseLost.Load() {
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
			if err := s.persistTask(task); err != nil {
				// Dead-letter is terminal: an undurable row means the exhausted
				// task re-runs from attempt one after a restart.
				log.Printf("task %s dead-letter not durable; may re-adopt after restart: %v", task.ID, err)
			}
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
		if err := s.persistTask(task); err != nil {
			// The retry slot lives only in memory now; a restart forgets the
			// backoff and the sweep adopts the task immediately.
			log.Printf("task %s retry state not durable; backoff may be lost: %v", task.ID, err)
		}
		s.recordTaskEvent(task, trace.EventTaskRetryScheduled, map[string]any{
			"reason":        err.Error(),
			"nextAttemptAt": next,
		})

		if s.store != nil {
			// The retry is scheduled durably; release ownership so any
			// instance's sweep claims the task when its slot arrives,
			// distributing retries across replicas instead of pinning them
			// to this one.
			return
		}

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
	if err := s.persistTask(task); err != nil {
		// The attempt count is memory-only now; a restart may re-run this
		// attempt against a stale counter.
		log.Printf("task %s attempt-start state not durable: %v", task.ID, err)
	}

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

	// Give the underlying request the same absolute timeout as the task
	// attempt so the lease reaper does not reclaim the request before the task
	// deadline is reached (0 keeps the server default).
	request, err := s.requestsService.CreateRequest(task.SessionID, task.ToolName, task.Input, task.TimeoutSeconds, "")
	if err != nil {
		return err
	}
	s.setTaskRequestID(task.ID, request.ID)
	s.setCurrentRequestID(task, request.ID)
	defer func() {
		s.setTaskRequestID(task.ID, "")
		s.setCurrentRequestID(task, "")
	}()
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
	if err := s.persistTask(task); err != nil {
		// The completed result exists only in memory; a restart re-adopts and
		// re-executes the task.
		log.Printf("task %s completion not durable; may re-execute after restart: %v", task.ID, err)
	}
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
		return nil, wrapf(ErrNoProviderAvailable, "no machines available with tool %s", toolName)
	}
	machines = s.machinesService.OrderMachinesByLoad(sessionID, machines)
	for _, machine := range machines {
		load, capacity := s.machinesService.MachineLoadInfo(sessionID, machine.ID)
		if load < capacity {
			return machine, nil
		}
	}
	return nil, wrapf(ErrMachineAtCapacity, "all machines at capacity for tool %s", toolName)
}

// persistTask writes the task through to the store with one retry, deriving
// its context from the service so it participates in graceful shutdown. The
// error is returned (and logged with task identity) so callers can react;
// terminal-state callers treat a failed persist as a data-loss risk and log
// loudly rather than swallowing it.
func (s *TasksService) persistTask(task *model.Task) error {
	if s.store == nil || task == nil {
		return nil
	}
	var err error
	for attempt := 0; attempt < 2; attempt++ {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(s.ctx), defaultPersistenceTimeout)
		err = s.store.SaveTask(ctx, task)
		cancel()
		if err == nil {
			return nil
		}
		time.Sleep(time.Duration(attempt+1) * 50 * time.Millisecond)
	}
	log.Printf("persist task save failed after retry: id=%s status=%s err=%v", task.ID, task.Status, err)
	return err
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

	if err := s.persistTask(task); err != nil {
		// The failure exists only in memory; a restart re-adopts the task.
		log.Printf("task %s failure state not durable; may re-execute after restart: %v", task.ID, err)
	}
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

// CurrentRequestID returns the ID of the request executing the task's current
// attempt, or "" when no attempt is in flight. The durable copy lives on the
// task record itself (model.Task.CurrentRequestID) and is authoritative across
// replicas; this live lookup covers the brief window before the attempt's
// first persist on this instance.
func (s *TasksService) CurrentRequestID(taskID string) string {
	requestID, _ := s.taskExecutionSnapshot(taskID)
	return requestID
}

// setCurrentRequestID records (or clears) the request executing the task's
// current attempt on the durable task record, so any replica can serve
// chunk-replay reads for a running task — not just the replica executing it.
func (s *TasksService) setCurrentRequestID(task *model.Task, requestID string) {
	s.tasksMutex.Lock()
	if task.Status == model.StatusCancelled {
		s.tasksMutex.Unlock()
		return
	}
	task.CurrentRequestID = requestID
	task.UpdatedAt = time.Now()
	s.tasksMutex.Unlock()
	if err := s.persistTask(task); err != nil {
		// Chunk-replay reads on other replicas fall back to the previous
		// request id until the next durable write.
		log.Printf("task %s current-request link not durable: %v", task.ID, err)
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

// CleanupOldTasks removes terminal tasks older than maxAge. Non-terminal
// tasks are never eligible regardless of age: deleting the row of a task
// that is pending, running, or awaiting retry would orphan work the
// adoption sweep could otherwise recover. A periodic sweeper invokes this
// with taskRetentionAge.
func (s *TasksService) CleanupOldTasks(maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)

	s.execMutex.RLock()
	executing := make(map[string]struct{}, len(s.executions))
	for id := range s.executions {
		executing[id] = struct{}{}
	}
	s.execMutex.RUnlock()

	var removals []string
	s.tasksMutex.Lock()
	for id, task := range s.tasks {
		if task == nil || task.CreatedAt.After(cutoff) {
			continue
		}
		if _, busy := executing[id]; busy {
			continue
		}
		switch task.Status {
		case model.StatusCompleted, model.StatusFailed, model.StatusCancelled:
		default:
			continue
		}
		removals = append(removals, id)
		delete(s.tasks, id)
	}
	s.tasksMutex.Unlock()

	if s.store == nil {
		return
	}
	for _, id := range removals {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(s.ctx), defaultPersistenceTimeout)
		if err := s.store.DeleteTask(ctx, id); err != nil {
			log.Printf("persist task cleanup delete failed: id=%s err=%v", id, err)
		}
		cancel()
	}
}
