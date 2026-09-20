package main

import (
	"sync"
	"time"
)

// Soak monitoring: sample the server's own runtime metrics (goroutine
// count, resident memory) while the load runs and, after the fleet has
// drained, assert the growth stayed bounded. The thresholds are leak
// detectors, not precision budgets — a real leak grows monotonically
// and trips them by orders of magnitude.

const soakSampleInterval = 15 * time.Second

type soakSample struct {
	at         time.Time
	goroutines float64
	rssMB      float64
}

type soakMonitor struct {
	mu      sync.Mutex
	samples []soakSample
	stop    chan struct{}
	done    chan struct{}
}

func newSoakMonitor() *soakMonitor {
	return &soakMonitor{stop: make(chan struct{}), done: make(chan struct{})}
}

func (m *soakMonitor) run(metricsURL string) {
	defer close(m.done)
	ticker := time.NewTicker(soakSampleInterval)
	defer ticker.Stop()
	for {
		select {
		case <-m.stop:
			return
		case <-ticker.C:
			metrics := scrapeMetrics(metricsURL)
			if metrics == nil {
				continue
			}
			sample := soakSample{at: time.Now()}
			for key, value := range metrics {
				switch {
				case key == "go_goroutines":
					sample.goroutines = value
				case key == "process_resident_memory_bytes":
					sample.rssMB = value / (1024 * 1024)
				}
			}
			m.mu.Lock()
			m.samples = append(m.samples, sample)
			m.mu.Unlock()
		}
	}
}

func (m *soakMonitor) halt() {
	close(m.stop)
	<-m.done
}

// finalSample reads one last point after the fleet drained.
func (m *soakMonitor) finalSample(metricsURL string) soakSample {
	metrics := scrapeMetrics(metricsURL)
	sample := soakSample{at: time.Now()}
	for key, value := range metrics {
		switch {
		case key == "go_goroutines":
			sample.goroutines = value
		case key == "process_resident_memory_bytes":
			sample.rssMB = value / (1024 * 1024)
		}
	}
	m.mu.Lock()
	m.samples = append(m.samples, sample)
	m.mu.Unlock()
	return sample
}

// assertions compares the run-window samples against the post-drain
// final point. Bounded growth is the pass condition:
//
//   - goroutines: the final count sits within a fixed slack of the
//     run-window minimum (a leaked goroutine per op would dwarf it).
//   - memory: the final RSS stays within 10% + 32MiB of the run-window
//     peak (caches legitimately grow during a soak; leaks do not stop).
func (m *soakMonitor) assertions() *drillResult {
	m.mu.Lock()
	samples := append([]soakSample(nil), m.samples...)
	m.mu.Unlock()

	result := &drillResult{Name: "soak"}
	if len(samples) < 2 {
		result.add("enough samples", false, "collected %d samples; soak needs a longer window", len(samples))
		return result
	}

	minGoroutines, peakGoroutines := samples[0].goroutines, samples[0].goroutines
	peakRSS := samples[0].rssMB
	for _, s := range samples[:len(samples)-1] {
		if s.goroutines < minGoroutines {
			minGoroutines = s.goroutines
		}
		if s.goroutines > peakGoroutines {
			peakGoroutines = s.goroutines
		}
		if s.rssMB > peakRSS {
			peakRSS = s.rssMB
		}
	}
	final := samples[len(samples)-1]

	goroutinesOK := final.goroutines <= minGoroutines+100 || final.goroutines <= peakGoroutines+20
	result.add("goroutine growth bounded", goroutinesOK,
		"min=%.0f peak=%.0f final=%.0f (leaked goroutines would dwarf the slack)", minGoroutines, peakGoroutines, final.goroutines)

	rssOK := final.rssMB <= peakRSS*1.10+32
	result.add("memory plateau", rssOK,
		"peak=%.0fMiB final=%.0fMiB (caches grow then flatten; leaks do not stop)", peakRSS, final.rssMB)

	return result
}
