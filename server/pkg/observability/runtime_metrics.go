package observability

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"toolplane/pkg/trace"
)

type requestMetricsSource interface {
	RequestMetricsSnapshot() (pending, claimed, running, done, failed, stalled, deadLetter int)
}

type machineMetricsSource interface {
	MachineMetricsSnapshot() (active, draining, inflight int)
}

type taskMetricsSource interface {
	TaskMetricsSnapshot() (pending, running, completed, failed, cancelled, deadLetter int)
}

// RuntimeMetricsCollector maintains the supported operator metrics surface.
type RuntimeMetricsCollector struct {
	mu       sync.RWMutex
	requests requestMetricsSource
	machines machineMetricsSource
	tasks    taskMetricsSource

	requestRequeues    atomic.Int64
	requestDeadLetters atomic.Int64
	taskRetries        atomic.Int64
	taskDeadLetters    atomic.Int64
}

// NewRuntimeMetricsCollector creates an empty collector ready to bind to live services.
func NewRuntimeMetricsCollector() *RuntimeMetricsCollector {
	return &RuntimeMetricsCollector{}
}

// Bind attaches live service sources used for gauge snapshots.
func (c *RuntimeMetricsCollector) Bind(requests requestMetricsSource, machines machineMetricsSource, tasks taskMetricsSource) {
	c.mu.Lock()
	c.requests = requests
	c.machines = machines
	c.tasks = tasks
	c.mu.Unlock()
}

// Handler returns an HTTP handler serving Prometheus-style metrics at /metrics.
func (c *RuntimeMetricsCollector) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", c.serveMetrics)
	return mux
}

// Record implements trace.SessionTracer so lifecycle events can update counters.
func (c *RuntimeMetricsCollector) Record(event trace.SessionEvent) {
	if c == nil {
		return
	}
	switch event.Event {
	case trace.EventRequestRequeued:
		c.requestRequeues.Add(1)
	case trace.EventRequestDeadLettered:
		c.requestDeadLetters.Add(1)
	case trace.EventTaskRetryScheduled:
		c.taskRetries.Add(1)
	case trace.EventTaskDeadLettered:
		c.taskDeadLetters.Add(1)
	}
}

func (c *RuntimeMetricsCollector) serveMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requests, machines, tasks := c.snapshotSources()
	requestPending, requestClaimed, requestRunning, _, _, _, requestDeadLetterCurrent := 0, 0, 0, 0, 0, 0, 0
	if requests != nil {
		requestPending, requestClaimed, requestRunning, _, _, _, requestDeadLetterCurrent = requests.RequestMetricsSnapshot()
	}

	machineActive, machineDraining, machineInflight := 0, 0, 0
	if machines != nil {
		machineActive, machineDraining, machineInflight = machines.MachineMetricsSnapshot()
	}

	taskPending, taskRunning, _, _, _, taskDeadLetterCurrent := 0, 0, 0, 0, 0, 0
	if tasks != nil {
		taskPending, taskRunning, _, _, _, taskDeadLetterCurrent = tasks.TaskMetricsSnapshot()
	}

	var builder strings.Builder
	writeMetric(&builder, "toolplane_request_queue_depth", "gauge", "Number of pending requests waiting for dispatch.", int64(requestPending))
	writeMetric(&builder, "toolplane_request_inflight", "gauge", "Number of claimed or running requests.", int64(requestClaimed+requestRunning))
	writeMetric(&builder, "toolplane_request_dead_letter_current", "gauge", "Number of requests currently marked dead letter.", int64(requestDeadLetterCurrent))
	writeMetric(&builder, "toolplane_request_requeues_total", "counter", "Total number of request requeues after lease expiry or capacity rejection.", c.requestRequeues.Load())
	writeMetric(&builder, "toolplane_request_dead_letters_total", "counter", "Total number of requests dead-lettered after retry exhaustion.", c.requestDeadLetters.Load())
	writeMetric(&builder, "toolplane_machine_active", "gauge", "Number of active registered machines.", int64(machineActive))
	writeMetric(&builder, "toolplane_machine_draining", "gauge", "Number of machines currently draining.", int64(machineDraining))
	writeMetric(&builder, "toolplane_machine_inflight_load", "gauge", "Total in-flight machine load reserved by running requests.", int64(machineInflight))
	writeMetric(&builder, "toolplane_task_pending", "gauge", "Number of tasks waiting to run or retry.", int64(taskPending))
	writeMetric(&builder, "toolplane_task_running", "gauge", "Number of tasks currently running.", int64(taskRunning))
	writeMetric(&builder, "toolplane_task_dead_letter_current", "gauge", "Number of tasks currently marked dead letter.", int64(taskDeadLetterCurrent))
	writeMetric(&builder, "toolplane_task_retries_total", "counter", "Total number of task retries scheduled.", c.taskRetries.Load())
	writeMetric(&builder, "toolplane_task_dead_letters_total", "counter", "Total number of tasks dead-lettered after retry exhaustion.", c.taskDeadLetters.Load())

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	_, _ = w.Write([]byte(builder.String()))
}

func (c *RuntimeMetricsCollector) snapshotSources() (requestMetricsSource, machineMetricsSource, taskMetricsSource) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.requests, c.machines, c.tasks
}

func writeMetric(builder *strings.Builder, name, metricType, help string, value int64) {
	builder.WriteString(fmt.Sprintf("# HELP %s %s\n", name, help))
	builder.WriteString(fmt.Sprintf("# TYPE %s %s\n", name, metricType))
	builder.WriteString(name)
	builder.WriteByte(' ')
	builder.WriteString(strconv.FormatInt(value, 10))
	builder.WriteByte('\n')
}
