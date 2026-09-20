package main

import (
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// collector accumulates per-op durations and error counts across all
// driver goroutines. Zero-duration samples (heartbeats, idle polls) are
// recorded for counts only.
type collector struct {
	mu       sync.Mutex
	samples  map[string][]float64
	errors   map[string]map[string]int
	chunkSum int64
}

func newCollector() *collector {
	return &collector{
		samples: map[string][]float64{},
		errors:  map[string]map[string]int{},
	}
}

func (c *collector) recordOp(name string, durationMS float64, code string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.samples[name] = append(c.samples[name], durationMS)
	if code != "" {
		if c.errors[name] == nil {
			c.errors[name] = map[string]int{}
		}
		c.errors[name][code]++
	}
}

func (c *collector) recordChunkCount(chunks int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.chunkSum += int64(chunks)
}

type opReport struct {
	Name   string         `json:"name"`
	Count  int            `json:"count"`
	P50MS  float64        `json:"p50_ms"`
	P95MS  float64        `json:"p95_ms"`
	P99MS  float64        `json:"p99_ms"`
	MaxMS  float64        `json:"max_ms"`
	Errors map[string]int `json:"errors,omitempty"`
}

type loadReport struct {
	Mode           string             `json:"mode"`
	Shape          string             `json:"shape"`
	Sessions       int                `json:"sessions"`
	DurationSec    float64            `json:"duration_sec"`
	Ops            []opReport         `json:"ops"`
	ChunksFollowed int64              `json:"chunks_followed"`
	MetricsDelta   map[string]float64 `json:"metrics_delta,omitempty"`
	Failed         bool               `json:"failed"`
}

func (c *collector) report(cfg config, metrics map[string]float64) loadReport {
	c.mu.Lock()
	defer c.mu.Unlock()

	report := loadReport{
		Mode:           cfg.mode,
		Shape:          cfg.shape,
		Sessions:       cfg.sessions,
		DurationSec:    cfg.duration.Seconds(),
		ChunksFollowed: c.chunkSum,
		MetricsDelta:   metrics,
	}
	for name, samples := range c.samples {
		sorted := append([]float64(nil), samples...)
		sort.Float64s(sorted)
		op := opReport{
			Name:   name,
			Count:  len(sorted),
			P50MS:  percentile(sorted, 0.50),
			P95MS:  percentile(sorted, 0.95),
			P99MS:  percentile(sorted, 0.99),
			MaxMS:  sorted[len(sorted)-1],
			Errors: c.errors[name],
		}
		if len(op.Errors) > 0 {
			report.Failed = true
		}
		report.Ops = append(report.Ops, op)
	}
	sort.Slice(report.Ops, func(i, j int) bool { return report.Ops[i].Name < report.Ops[j].Name })
	return report
}

// scrapeMetrics reads a Prometheus text exposition into name{labels} ->
// value. An empty URL yields a nil map (external runs without a metrics
// endpoint still report op statistics).
func scrapeMetrics(baseURL string) map[string]float64 {
	if strings.TrimSpace(baseURL) == "" {
		return nil
	}
	response, err := http.Get(strings.TrimRight(baseURL, "/") + "/metrics")
	if err != nil {
		return nil
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil
	}

	values := map[string]float64{}
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		value, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			continue
		}
		values[fields[0]] = value
	}
	return values
}

// metricsDelta keeps counter deltas (end - start) and end values for
// gauges, filtered to the toolplane surface so reports stay readable.
func metricsDelta(start, end map[string]float64) map[string]float64 {
	if end == nil {
		return nil
	}
	delta := map[string]float64{}
	for key, endValue := range end {
		if !strings.HasPrefix(key, "toolplane_") {
			continue
		}
		// Gauge families the report carries as levels, not deltas.
		if strings.Contains(key, "queue_depth") || strings.Contains(key, "tls_enabled") ||
			strings.Contains(key, "inflight") || strings.Contains(key, "_current") ||
			strings.Contains(key, "machine_active") || strings.Contains(key, "machine_draining") ||
			strings.Contains(key, "task_pending") || strings.Contains(key, "task_running") {
			continue
		}
		if start != nil {
			if startValue, ok := start[key]; ok {
				delta[key] = endValue - startValue
				continue
			}
		}
		delta[key] = endValue
	}
	// Final levels for the gauges operators read first.
	for _, gauge := range []string{
		"toolplane_request_queue_depth", "toolplane_request_inflight",
		"toolplane_machine_active", "toolplane_machine_inflight_load",
	} {
		for key, value := range end {
			if key == gauge || strings.HasPrefix(key, gauge+"{") {
				delta["final_"+key] = value
			}
		}
	}
	return delta
}
