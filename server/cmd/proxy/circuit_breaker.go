package main

import (
	"errors"
	"log"
	"math"
	"sync/atomic"
	"time"

	"github.com/sony/gobreaker"
)

var errTooManyConcurrent = errors.New("too many concurrent requests")

// CircuitBreakerStats captures breaker health for diagnostics.
type CircuitBreakerStats struct {
	State    string           `json:"state"`
	Inflight int64            `json:"inflight"`
	Rejected int64            `json:"rejected"`
	Accepted int64            `json:"accepted"`
	Counts   gobreaker.Counts `json:"counts"`
}

// CircuitBreakerManager wraps gobreaker with proxy-specific concurrency tracking.
type CircuitBreakerManager struct {
	breaker       *gobreaker.TwoStepCircuitBreaker
	maxConcurrent int64

	openTimeout  time.Duration
	state        atomic.Uint32
	stateChanged atomic.Int64

	inflight      atomic.Int64
	totalRejected atomic.Int64
	totalAccepted atomic.Int64
}

// NewCircuitBreakerManager configures a two-step breaker with half-open probing.
func NewCircuitBreakerManager(maxConcurrent int64) *CircuitBreakerManager {
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}

	manager := &CircuitBreakerManager{
		maxConcurrent: maxConcurrent,
		openTimeout:   30 * time.Second,
	}

	halfOpenProbes := math.Max(1, float64(maxConcurrent)/10)
	if halfOpenProbes > float64(^uint32(0)) {
		halfOpenProbes = float64(^uint32(0))
	}

	settings := gobreaker.Settings{
		Name:        "proxy-circuit-breaker",
		Timeout:     manager.openTimeout,
		Interval:    time.Minute,
		MaxRequests: uint32(halfOpenProbes),
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: manager.onStateChange,
	}

	manager.breaker = gobreaker.NewTwoStepCircuitBreaker(settings)
	manager.state.Store(uint32(gobreaker.StateClosed))
	manager.stateChanged.Store(time.Now().UnixNano())

	return manager
}

// Begin reserves breaker + concurrency slot; caller must call done(success) and release().
func (m *CircuitBreakerManager) Begin() (func(success bool), func(), time.Duration, error) {
	if m == nil {
		return nil, func() {}, 0, nil
	}

	current := m.inflight.Add(1)
	if m.maxConcurrent > 0 && current > m.maxConcurrent {
		m.inflight.Add(-1)
		m.totalRejected.Add(1)
		return nil, nil, time.Second, errTooManyConcurrent
	}

	done, err := m.breaker.Allow()
	if err != nil {
		m.inflight.Add(-1)
		m.totalRejected.Add(1)
		wait := m.retryAfter()
		if wait == 0 && errors.Is(err, gobreaker.ErrTooManyRequests) {
			wait = 5 * time.Second
		}
		return nil, nil, wait, err
	}

	m.totalAccepted.Add(1)
	release := func() {
		m.inflight.Add(-1)
	}

	return done, release, 0, nil
}

// Stats exposes aggregate breaker metrics for health endpoints.
func (m *CircuitBreakerManager) Stats() CircuitBreakerStats {
	if m == nil {
		return CircuitBreakerStats{}
	}

	counts := gobreaker.Counts{}
	if m.breaker != nil {
		counts = m.breaker.Counts()
	}

	return CircuitBreakerStats{
		State:    gobreaker.State(m.state.Load()).String(),
		Inflight: m.inflight.Load(),
		Rejected: m.totalRejected.Load(),
		Accepted: m.totalAccepted.Load(),
		Counts:   counts,
	}
}

// IsOpen reports if breaker currently rejects all traffic.
func (m *CircuitBreakerManager) IsOpen() bool {
	if m == nil {
		return false
	}
	return gobreaker.State(m.state.Load()) == gobreaker.StateOpen
}

func (m *CircuitBreakerManager) retryAfter() time.Duration {
	if m == nil {
		return 0
	}
	if gobreaker.State(m.state.Load()) != gobreaker.StateOpen {
		return 0
	}
	elapsed := time.Since(time.Unix(0, m.stateChanged.Load()))
	remaining := m.openTimeout - elapsed
	if remaining < time.Second {
		remaining = time.Second
	}
	return remaining
}

func (m *CircuitBreakerManager) onStateChange(name string, from, to gobreaker.State) {
	m.state.Store(uint32(to))
	m.stateChanged.Store(time.Now().UnixNano())
	log.Printf("%s transition: %s → %s", name, from.String(), to.String())
}
