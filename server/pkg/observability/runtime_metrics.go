package observability

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

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

// RuntimeMetricsCollector maintains the supported operator metrics surface
// on a private prometheus registry: runtime gauges and lifecycle counters
// plus per-RPC request counters and duration histograms recorded by the
// gRPC interceptors.
type RuntimeMetricsCollector struct {
	mu       sync.RWMutex
	requests requestMetricsSource
	machines machineMetricsSource
	tasks    taskMetricsSource

	registry *prometheus.Registry

	requestRequeues    prometheus.Counter
	requestDeadLetters prometheus.Counter
	taskRetries        prometheus.Counter
	taskDeadLetters    prometheus.Counter

	grpcRequests  *prometheus.CounterVec
	grpcDurations *prometheus.HistogramVec
}

// NewRuntimeMetricsCollector creates an empty collector ready to bind to live services.
func NewRuntimeMetricsCollector() *RuntimeMetricsCollector {
	c := &RuntimeMetricsCollector{
		registry: prometheus.NewRegistry(),
	}
	// Go runtime and process collectors come for free and are what
	// operators expect from a /metrics endpoint.
	c.registry.MustRegister(collectors.NewGoCollector())
	c.registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	c.requestRequeues = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "toolplane_request_requeues_total",
		Help: "Total number of request requeues after lease expiry or capacity rejection.",
	})
	c.requestDeadLetters = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "toolplane_request_dead_letters_total",
		Help: "Total number of requests dead-lettered after retry exhaustion.",
	})
	c.taskRetries = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "toolplane_task_retries_total",
		Help: "Total number of task retries scheduled.",
	})
	c.taskDeadLetters = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "toolplane_task_dead_letters_total",
		Help: "Total number of tasks dead-lettered after retry exhaustion.",
	})
	c.registry.MustRegister(c.requestRequeues, c.requestDeadLetters, c.taskRetries, c.taskDeadLetters)

	// Labels are bounded by construction: methods by the proto surface,
	// codes by the gRPC code set.
	c.grpcRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "toolplane_grpc_requests_total",
		Help: "Total gRPC requests by full method and resulting status code.",
	}, []string{"method", "code"})
	c.grpcDurations = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "toolplane_grpc_request_duration_seconds",
		Help:    "gRPC request latency by full method.",
		Buckets: []float64{.001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30},
	}, []string{"method"})
	c.registry.MustRegister(c.grpcRequests, c.grpcDurations)

	c.registerGauges()
	return c
}

func (c *RuntimeMetricsCollector) registerGauges() {
	gauge := func(name, help string, value func() float64) {
		c.registry.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: name,
			Help: help,
		}, value))
	}

	gauge("toolplane_request_queue_depth",
		"Number of pending requests waiting for dispatch.",
		func() float64 {
			if s := c.requestSource(); s != nil {
				pending, _, _, _, _, _, _ := s.RequestMetricsSnapshot()
				return float64(pending)
			}
			return 0
		})
	gauge("toolplane_request_inflight",
		"Number of claimed or running requests.",
		func() float64 {
			if s := c.requestSource(); s != nil {
				_, claimed, running, _, _, _, _ := s.RequestMetricsSnapshot()
				return float64(claimed + running)
			}
			return 0
		})
	gauge("toolplane_request_dead_letter_current",
		"Number of requests currently marked dead letter.",
		func() float64 {
			if s := c.requestSource(); s != nil {
				_, _, _, _, _, _, dead := s.RequestMetricsSnapshot()
				return float64(dead)
			}
			return 0
		})
	gauge("toolplane_machine_active",
		"Number of active registered machines.",
		func() float64 {
			if s := c.machineSource(); s != nil {
				active, _, _ := s.MachineMetricsSnapshot()
				return float64(active)
			}
			return 0
		})
	gauge("toolplane_machine_draining",
		"Number of machines currently draining.",
		func() float64 {
			if s := c.machineSource(); s != nil {
				_, draining, _ := s.MachineMetricsSnapshot()
				return float64(draining)
			}
			return 0
		})
	gauge("toolplane_machine_inflight_load",
		"Total in-flight machine load reserved by running requests.",
		func() float64 {
			if s := c.machineSource(); s != nil {
				_, _, inflight := s.MachineMetricsSnapshot()
				return float64(inflight)
			}
			return 0
		})
	gauge("toolplane_task_pending",
		"Number of tasks waiting to run or retry.",
		func() float64 {
			if s := c.taskSource(); s != nil {
				pending, _, _, _, _, _ := s.TaskMetricsSnapshot()
				return float64(pending)
			}
			return 0
		})
	gauge("toolplane_task_running",
		"Number of tasks currently running.",
		func() float64 {
			if s := c.taskSource(); s != nil {
				_, running, _, _, _, _ := s.TaskMetricsSnapshot()
				return float64(running)
			}
			return 0
		})
	gauge("toolplane_task_dead_letter_current",
		"Number of tasks currently marked dead letter.",
		func() float64 {
			if s := c.taskSource(); s != nil {
				_, _, _, _, _, dead := s.TaskMetricsSnapshot()
				return float64(dead)
			}
			return 0
		})
}

func (c *RuntimeMetricsCollector) requestSource() requestMetricsSource {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.requests
}

func (c *RuntimeMetricsCollector) machineSource() machineMetricsSource {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.machines
}

func (c *RuntimeMetricsCollector) taskSource() taskMetricsSource {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.tasks
}

// Bind attaches live service sources used for gauge snapshots.
func (c *RuntimeMetricsCollector) Bind(requests requestMetricsSource, machines machineMetricsSource, tasks taskMetricsSource) {
	c.mu.Lock()
	c.requests = requests
	c.machines = machines
	c.tasks = tasks
	c.mu.Unlock()
}

// Handler returns an HTTP handler serving Prometheus metrics at /metrics.
func (c *RuntimeMetricsCollector) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(c.registry, promhttp.HandlerOpts{}))
	return mux
}

// Record implements trace.SessionTracer so lifecycle events can update counters.
func (c *RuntimeMetricsCollector) Record(event trace.SessionEvent) {
	if c == nil {
		return
	}
	switch event.Event {
	case trace.EventRequestRequeued:
		c.requestRequeues.Inc()
	case trace.EventRequestDeadLettered:
		c.requestDeadLetters.Inc()
	case trace.EventTaskRetryScheduled:
		c.taskRetries.Inc()
	case trace.EventTaskDeadLettered:
		c.taskDeadLetters.Inc()
	}
}

// observeGRPC records one completed RPC for the interceptor metrics.
func (c *RuntimeMetricsCollector) observeGRPC(method, code string, seconds float64) {
	c.grpcRequests.WithLabelValues(method, code).Inc()
	c.grpcDurations.WithLabelValues(method).Observe(seconds)
}
