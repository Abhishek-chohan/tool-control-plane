package service

import (
	"context"
	"testing"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage/memory"
	"toolplane/pkg/trace"
)

// newAdoptionStack builds a store-backed task service pair sharing one
// store, mirroring two replicas behind one database.
func newAdoptionStack(t *testing.T) (*memory.Store, *TasksService, *TasksService) {
	t.Helper()
	store := memory.New()
	mk := func() *TasksService {
		toolSvc := NewToolService(trace.NopTracer(), store)
		machineSvc := NewMachinesService(context.Background(), toolSvc, trace.NopTracer(), store)
		requestsSvc := NewRequestsService(context.Background(), toolSvc, machineSvc, trace.NopTracer(), store)
		return NewTasksService(context.Background(), toolSvc, machineSvc, requestsSvc, trace.NopTracer(), store)
	}
	return store, mk(), mk()
}

// ageTask rewrites the stored row with an old UpdatedAt, simulating a lease
// that has elapsed since the owner last touched it.
func ageTask(t *testing.T, store *memory.Store, taskID string, age time.Duration) {
	t.Helper()
	ctx := context.Background()
	tasks, err := store.AllTasks(ctx)
	if err != nil {
		t.Fatalf("load tasks: %v", err)
	}
	for _, stored := range tasks {
		if stored.ID != taskID {
			continue
		}
		stored.UpdatedAt = time.Now().Add(-age)
		if err := store.SaveTask(ctx, stored); err != nil {
			t.Fatalf("age task: %v", err)
		}
		return
	}
	t.Fatalf("task %s not found for aging", taskID)
}

func getStoredTask(t *testing.T, store *memory.Store, taskID string) *model.Task {
	t.Helper()
	tasks, err := store.AllTasks(context.Background())
	if err != nil {
		t.Fatalf("load tasks: %v", err)
	}
	for _, stored := range tasks {
		if stored.ID == taskID {
			return stored
		}
	}
	return nil
}

func TestTaskAdoptionFencing(t *testing.T) {
	store, svcA, svcB := newAdoptionStack(t)
	ctx := context.Background()

	task := model.NewTask("sess-adoption", "echo", `{}`)
	task.Status = model.StatusPending
	if err := store.SaveTask(ctx, task); err != nil {
		t.Fatalf("seed task: %v", err)
	}

	// Unowned: the first claimant wins, the second loses while the lease is
	// live.
	claimed, ok, err := store.ClaimTaskForAdoption(ctx, task.ID, "instance-a", taskAdoptionLeaseTTL)
	if err != nil || !ok {
		t.Fatalf("first claim: ok=%v err=%v", ok, err)
	}
	if claimed.ID != task.ID {
		t.Fatalf("claimed the wrong task: %s", claimed.ID)
	}
	if _, ok, err := store.ClaimTaskForAdoption(ctx, task.ID, "instance-b", taskAdoptionLeaseTTL); err != nil || ok {
		t.Fatalf("live lease stolen: ok=%v err=%v", ok, err)
	}

	// Only the owner renews.
	if renewed, err := store.RenewTaskAdoption(ctx, task.ID, "instance-a"); err != nil || !renewed {
		t.Fatalf("owner renewal: renewed=%v err=%v", renewed, err)
	}
	if renewed, err := store.RenewTaskAdoption(ctx, task.ID, "instance-b"); err != nil || renewed {
		t.Fatalf("non-owner renewal succeeded: renewed=%v err=%v", renewed, err)
	}

	// Only the owner releases; after release, anyone can claim.
	if err := store.ReleaseTaskAdoption(ctx, task.ID, "instance-b"); err != nil {
		t.Fatalf("non-owner release errored: %v", err)
	}
	if _, ok, _ := store.ClaimTaskForAdoption(ctx, task.ID, "instance-b", taskAdoptionLeaseTTL); ok {
		t.Fatal("non-owner release made the task claimable")
	}
	if err := store.ReleaseTaskAdoption(ctx, task.ID, "instance-a"); err != nil {
		t.Fatalf("owner release: %v", err)
	}
	if _, ok, err := store.ClaimTaskForAdoption(ctx, task.ID, "instance-b", taskAdoptionLeaseTTL); err != nil || !ok {
		t.Fatalf("post-release claim: ok=%v err=%v", ok, err)
	}

	_ = svcA
	_ = svcB
}

func TestTaskAdoptionStealsExpiredLease(t *testing.T) {
	store, _, _ := newAdoptionStack(t)
	ctx := context.Background()

	task := model.NewTask("sess-adoption-expiry", "echo", `{}`)
	task.Status = model.StatusPending
	if err := store.SaveTask(ctx, task); err != nil {
		t.Fatalf("seed task: %v", err)
	}
	if _, ok, err := store.ClaimTaskForAdoption(ctx, task.ID, "instance-a", taskAdoptionLeaseTTL); err != nil || !ok {
		t.Fatalf("initial claim: ok=%v err=%v", ok, err)
	}

	// Simulate the owner going silent: its last touch ages past the TTL.
	ageTask(t, store, task.ID, taskAdoptionLeaseTTL+5*time.Second)

	if _, ok, err := store.ClaimTaskForAdoption(ctx, task.ID, "instance-b", taskAdoptionLeaseTTL); err != nil || !ok {
		t.Fatalf("expired lease not stealable: ok=%v err=%v", ok, err)
	}
	// The previous owner's renewal now fails: it must stop executing.
	if renewed, err := store.RenewTaskAdoption(ctx, task.ID, "instance-a"); err != nil || renewed {
		t.Fatalf("stale owner still renews: renewed=%v err=%v", renewed, err)
	}
}

func TestTaskCreateDispatchesThroughFenceAndCompletes(t *testing.T) {
	store := memory.New()
	toolSvc := NewToolService(trace.NopTracer(), store)
	machineSvc := NewMachinesService(context.Background(), toolSvc, trace.NopTracer(), store)
	requestsSvc := NewRequestsService(context.Background(), toolSvc, machineSvc, trace.NopTracer(), store)
	tasksSvc := NewTasksService(context.Background(), toolSvc, machineSvc, requestsSvc, trace.NopTracer(), store)

	const sessionID = "sess-task-dispatch"
	const machineID = "machine-task-dispatch"
	if _, err := machineSvc.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	created, err := tasksSvc.CreateTask(sessionID, "echo", `{"m":"dispatch"}`, "")
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	// The dispatch claims the fresh task immediately and executes it; drive
	// the underlying request to completion as the executor would.
	deadline := time.Now().Add(5 * time.Second)
	var requestID string
	for {
		task, getErr := tasksSvc.GetTask(created.ID)
		if getErr != nil {
			t.Fatalf("get task: %v", getErr)
		}
		requestID = task.CurrentRequestID
		if requestID != "" {
			break
		}
		if task.Status == model.StatusCompleted || task.Status == model.StatusFailed {
			t.Fatalf("task settled without executing: status=%s", task.Status)
		}
		if time.Now().After(deadline) {
			t.Fatalf("dispatch never started the task: status=%s", task.Status)
		}
		time.Sleep(10 * time.Millisecond)
	}

	attemptReq, err := requestsSvc.GetRequestByID(sessionID, requestID)
	if err != nil {
		t.Fatalf("get attempt request: %v", err)
	}
	if err := requestsSvc.SubmitRequestResult(
		sessionID, requestID, attemptReq.LeasedBy, attemptReq.LeaseEpoch,
		"dispatched-done", model.ResultTypeResolution, nil,
	); err != nil {
		t.Fatalf("submit attempt result: %v", err)
	}

	for {
		task, getErr := tasksSvc.GetTask(created.ID)
		if getErr != nil {
			t.Fatalf("get task: %v", getErr)
		}
		if task.Status == model.StatusCompleted {
			if task.Result != "dispatched-done" {
				t.Fatalf("task completed with the wrong result: %q", task.Result)
			}
			// Terminal: ownership released so the sweep does not consider it.
			// Terminal states release ownership: renewal by the finished
			// owner must now fail, proving the sweep will not skip this task
			// as someone else's live work.
			renewed, err := store.RenewTaskAdoption(context.Background(), created.ID, tasksSvc.instanceID)
			if err != nil {
				t.Fatalf("post-terminal renewal errored: %v", err)
			}
			if renewed {
				t.Fatal("terminal task still holds its adoption lease")
			}
			break
		}
		if task.Status == model.StatusFailed || time.Now().After(deadline) {
			t.Fatalf("task did not complete: status=%s err=%s", task.Status, task.Error)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestTaskSweepAdoptsOrphanedPendingTask(t *testing.T) {
	store := memory.New()
	toolSvc := NewToolService(trace.NopTracer(), store)
	machineSvc := NewMachinesService(context.Background(), toolSvc, trace.NopTracer(), store)
	requestsSvc := NewRequestsService(context.Background(), toolSvc, machineSvc, trace.NopTracer(), store)
	tasksSvc := NewTasksService(context.Background(), toolSvc, machineSvc, requestsSvc, trace.NopTracer(), store)

	const sessionID = "sess-task-sweep"
	const machineID = "machine-task-sweep"
	if _, err := machineSvc.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	// Seed a task that was enqueued by an instance that then died: pending,
	// unowned, and its last touch is older than the sweep cares about
	// (aged so the adoption claim sees an expired lease).
	orphan := model.NewTask(sessionID, "echo", `{"m":"sweep"}`)
	orphan.Status = model.StatusPending
	orphan.UpdatedAt = time.Now().Add(-2 * taskAdoptionLeaseTTL)
	if err := store.SaveTask(context.Background(), orphan); err != nil {
		t.Fatalf("seed orphan: %v", err)
	}
	tasksSvc.tasksMutex.Lock()
	tasksSvc.tasks[orphan.ID] = orphan.Clone()
	tasksSvc.tasksMutex.Unlock()

	// Drive the sweep directly (the ticker is 5s; the logic is what
	// matters). Mirror adoptDueTasks' loop body for one pass.
	ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
	defer cancel()
	candidates, err := store.FindNonTerminalTasks(ctx)
	if err != nil {
		t.Fatalf("sweep scan: %v", err)
	}
	claimedAny := false
	for _, task := range candidates {
		if task.ID != orphan.ID || !tasksSvc.taskAttemptDue(task) {
			continue
		}
		claimed, ok, claimErr := store.ClaimTaskForAdoption(ctx, task.ID, tasksSvc.instanceID, taskAdoptionLeaseTTL)
		if claimErr != nil || !ok {
			continue
		}
		claimedAny = true
		go tasksSvc.executeTask(claimed)
	}
	if !claimedAny {
		t.Fatal("sweep did not adopt the orphaned pending task")
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		task, getErr := tasksSvc.GetTask(orphan.ID)
		if getErr != nil {
			t.Fatalf("get task: %v", getErr)
		}
		if task.CurrentRequestID != "" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("adopted task never started: status=%s", task.Status)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestCleanupOldTasksRetainsNonTerminal(t *testing.T) {
	store := memory.New()
	toolSvc := NewToolService(trace.NopTracer(), store)
	machineSvc := NewMachinesService(context.Background(), toolSvc, trace.NopTracer(), store)
	requestsSvc := NewRequestsService(context.Background(), toolSvc, machineSvc, trace.NopTracer(), store)
	tasksSvc := NewTasksService(context.Background(), toolSvc, machineSvc, requestsSvc, trace.NopTracer(), store)

	old := time.Now().Add(-2 * taskRetentionAge)

	terminal := model.NewTask("sess-cleanup", "echo", `{}`)
	terminal.Status = model.StatusCompleted
	terminal.CreatedAt = old
	terminal.UpdatedAt = old
	if err := store.SaveTask(context.Background(), terminal); err != nil {
		t.Fatalf("seed terminal: %v", err)
	}

	// A pending task older than the retention age must survive: its row is
	// the only record of work the sweep could still recover.
	pending := model.NewTask("sess-cleanup", "echo", `{}`)
	pending.Status = model.StatusPending
	pending.UpdatedAt = old
	if err := store.SaveTask(context.Background(), pending); err != nil {
		t.Fatalf("seed pending: %v", err)
	}

	tasksSvc.tasksMutex.Lock()
	tasksSvc.tasks[terminal.ID] = terminal.Clone()
	tasksSvc.tasks[pending.ID] = pending.Clone()
	tasksSvc.tasksMutex.Unlock()

	tasksSvc.CleanupOldTasks(taskRetentionAge)

	if getStoredTask(t, store, terminal.ID) != nil {
		t.Fatal("old terminal task survived cleanup")
	}
	if getStoredTask(t, store, pending.ID) == nil {
		t.Fatal("old non-terminal task was cleaned up")
	}
}
