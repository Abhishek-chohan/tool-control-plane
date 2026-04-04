package main

import (
	"log"
	"math"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

type ThrottleReason string

const (
	ThrottleReasonNone         ThrottleReason = ""
	ThrottleReasonAPIRate      ThrottleReason = "api_rate_limit"
	ThrottleReasonIPRate       ThrottleReason = "ip_rate_limit"
	ThrottleReasonConcurrency  ThrottleReason = "concurrency_limit"
	ThrottleReasonCircuitOpen  ThrottleReason = "circuit_open"
	ThrottleReasonCircuitProbe ThrottleReason = "circuit_probe_limit"
	ThrottleReasonUnknown      ThrottleReason = "throttle_unknown"
)

// ThrottleSnapshot captures counters for throttled requests.
type ThrottleSnapshot struct {
	Total        int64 `json:"total"`
	APIRate      int64 `json:"apiRate"`
	IPRate       int64 `json:"ipRate"`
	Concurrency  int64 `json:"concurrency"`
	CircuitOpen  int64 `json:"circuitOpen"`
	CircuitProbe int64 `json:"circuitProbe"`
	Unknown      int64 `json:"unknown"`
}

// ThrottleTracker tracks throttled responses and emits structured logs.
type ThrottleTracker struct {
	apiRate      atomic.Int64
	ipRate       atomic.Int64
	concurrency  atomic.Int64
	circuitOpen  atomic.Int64
	circuitProbe atomic.Int64
	unknown      atomic.Int64
	total        atomic.Int64
}

func NewThrottleTracker() *ThrottleTracker {
	return &ThrottleTracker{}
}

func (t *ThrottleTracker) Record(reason ThrottleReason, retryAfter time.Duration, apiKey, clientIP, detail string) {
	if t == nil {
		return
	}
	switch reason {
	case ThrottleReasonAPIRate:
		t.apiRate.Add(1)
	case ThrottleReasonIPRate:
		t.ipRate.Add(1)
	case ThrottleReasonConcurrency:
		t.concurrency.Add(1)
	case ThrottleReasonCircuitOpen:
		t.circuitOpen.Add(1)
	case ThrottleReasonCircuitProbe:
		t.circuitProbe.Add(1)
	default:
		t.unknown.Add(1)
		reason = ThrottleReasonUnknown
	}
	t.total.Add(1)

	log.Printf("throttle reason=%s retry_after=%s api_key_present=%t client_ip=%s detail=%s",
		reason,
		retryAfter.String(),
		apiKey != "",
		clientIP,
		detail,
	)
}

func (t *ThrottleTracker) Snapshot() ThrottleSnapshot {
	if t == nil {
		return ThrottleSnapshot{}
	}
	return ThrottleSnapshot{
		Total:        t.total.Load(),
		APIRate:      t.apiRate.Load(),
		IPRate:       t.ipRate.Load(),
		Concurrency:  t.concurrency.Load(),
		CircuitOpen:  t.circuitOpen.Load(),
		CircuitProbe: t.circuitProbe.Load(),
		Unknown:      t.unknown.Load(),
	}
}

func applyRetryAfterHeader(w http.ResponseWriter, wait time.Duration) time.Duration {
	if wait <= 0 {
		wait = time.Second
	}
	seconds := int(math.Ceil(wait.Seconds()))
	if seconds < 1 {
		seconds = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(seconds))
	return time.Duration(seconds) * time.Second
}
