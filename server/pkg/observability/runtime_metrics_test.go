package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/service"
	"toolplane/pkg/trace"
)

type staticTaskSource struct {
	pending    int
	running    int
	completed  int
	failed     int
	cancelled  int
	deadLetter int
}

func (s staticTaskSource) TaskMetricsSnapshot() (pending, running, completed, failed, cancelled, deadLetter int) {
	return s.pending, s.running, s.completed, s.failed, s.cancelled, s.deadLetter
}

func TestRuntimeMetricsCollectorRendersCurrentRuntimeStateAndCounters(t *testing.T) {
	collector := NewRuntimeMetricsCollector()
	toolService := service.NewToolService(collector, nil)
	machineService := service.NewMachinesService(toolService, collector, nil)
	requestService := service.NewRequestsService(toolService, machineService, collector, nil)
	collector.Bind(requestService, machineService, staticTaskSource{pending: 1, deadLetter: 1})

	const sessionID = "metrics-session"
	const machineID = "metrics-machine"

	_, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	})
	if err != nil {
		t.Fatalf("register machine: %v", err)
	}

	activeRequest, err := requestService.CreateRequest(sessionID, "echo", `{"message":"active"}`)
	if err != nil {
		t.Fatalf("create active request: %v", err)
	}
	if _, err := requestService.ClaimRequest(sessionID, activeRequest.ID, machineID); err != nil {
		t.Fatalf("claim active request: %v", err)
	}
	if _, err := requestService.UpdateRequest(sessionID, activeRequest.ID, model.RequestStatusRunning, nil, ""); err != nil {
		t.Fatalf("mark active request running: %v", err)
	}

	pendingRequest, err := requestService.CreateRequest(sessionID, "echo", `{"message":"pending"}`)
	if err != nil {
		t.Fatalf("create pending request: %v", err)
	}

	cancelledRequest, err := requestService.CreateRequest(sessionID, "echo", `{"message":"cancelled"}`)
	if err != nil {
		t.Fatalf("create cancelled request: %v", err)
	}
	if err := requestService.CancelRequest(sessionID, cancelledRequest.ID); err != nil {
		t.Fatalf("cancel request: %v", err)
	}
	collector.Record(trace.SessionEvent{Event: trace.EventRequestDeadLettered})
	collector.Record(trace.SessionEvent{Event: trace.EventTaskRetryScheduled})

	drainDone := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		drainDone <- machineService.DrainMachine(ctx, sessionID, machineID)
	}()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if machineService.IsMachineDraining(sessionID, machineID) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !machineService.IsMachineDraining(sessionID, machineID) {
		t.Fatal("expected machine to enter draining state")
	}

	recorder := httptest.NewRecorder()
	collector.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("metrics status = %d, want 200", recorder.Code)
	}
	body := recorder.Body.String()
	for _, fragment := range []string{
		"toolplane_request_queue_depth 1",
		"toolplane_request_inflight 1",
		"toolplane_request_dead_letter_current 1",
		"toolplane_request_requeues_total 0",
		"toolplane_request_dead_letters_total 1",
		"toolplane_machine_active 1",
		"toolplane_machine_draining 1",
		"toolplane_machine_inflight_load 1",
		"toolplane_task_pending 1",
		"toolplane_task_running 0",
		"toolplane_task_dead_letter_current 1",
		"toolplane_task_retries_total 1",
		"toolplane_task_dead_letters_total 0",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("metrics output missing %q:\n%s", fragment, body)
		}
	}

	if err := requestService.SubmitRequestResult(sessionID, activeRequest.ID, map[string]string{"echo": "done"}, model.ResultTypeResolution, nil); err != nil {
		t.Fatalf("submit active request result: %v", err)
	}
	select {
	case err := <-drainDone:
		if err != nil {
			t.Fatalf("drain machine: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("expected drain to complete after active request finished")
	}

	if _, err := requestService.GetRequestByID(sessionID, pendingRequest.ID); err != nil {
		t.Fatalf("pending request should remain accessible: %v", err)
	}
}
