package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

func TestProxyHealthReflectsRateLimitRejectsAndThrottleCounters(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	breaker := NewCircuitBreakerManager(8)
	rateLimiter := NewRateLimiterManager(ctx, rate.Limit(1), 1, 0, 0)
	tracker := NewThrottleTracker()
	handler := newProxyRootHandler(
		proxyConfig{environment: "test", allowAnyOrigin: true, allowInsecureBackend: true},
		breaker,
		rateLimiter,
		tracker,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	)

	first := httptest.NewRecorder()
	firstReq := httptest.NewRequest(http.MethodPost, "/api/CreateSession", strings.NewReader(`{}`))
	firstReq.Header.Set("Grpc-Metadata-api_key", "fixture-key")
	firstReq.RemoteAddr = "198.51.100.9:1234"
	handler.ServeHTTP(first, firstReq)
	if first.Code != http.StatusNoContent {
		t.Fatalf("first request status = %d, want %d", first.Code, http.StatusNoContent)
	}

	second := httptest.NewRecorder()
	secondReq := httptest.NewRequest(http.MethodPost, "/api/CreateSession", strings.NewReader(`{}`))
	secondReq.Header.Set("Grpc-Metadata-api_key", "fixture-key")
	secondReq.RemoteAddr = "198.51.100.9:1234"
	handler.ServeHTTP(second, secondReq)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want %d", second.Code, http.StatusTooManyRequests)
	}
	if got := second.Header().Get("Retry-After"); got != "1" {
		t.Fatalf("Retry-After = %q, want 1", got)
	}

	healthRecorder := httptest.NewRecorder()
	handler.ServeHTTP(healthRecorder, httptest.NewRequest(http.MethodGet, "/health", nil))
	if healthRecorder.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", healthRecorder.Code, http.StatusOK)
	}

	var payload healthResponse
	if err := json.NewDecoder(healthRecorder.Body).Decode(&payload); err != nil {
		t.Fatalf("decode health payload: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("health status payload = %q, want ok", payload.Status)
	}
	if payload.RateLimitRejects != 1 {
		t.Fatalf("rate limit rejects = %d, want 1", payload.RateLimitRejects)
	}
	if payload.Throttle.Total != 1 || payload.Throttle.APIRate != 1 {
		t.Fatalf("throttle snapshot = %#v, want total=1 apiRate=1", payload.Throttle)
	}
	if payload.Circuit.State != gobreaker.StateClosed.String() {
		t.Fatalf("circuit state = %q, want %q", payload.Circuit.State, gobreaker.StateClosed.String())
	}
}

func TestBuildHealthResponseReportsDegradedOpenCircuit(t *testing.T) {
	breaker := NewCircuitBreakerManager(1)
	breaker.onStateChange("proxy-circuit-breaker", gobreaker.StateClosed, gobreaker.StateOpen)

	response, statusCode := buildHealthResponse(breaker, nil, nil, time.Unix(1_700_000_000, 0).UTC())
	if statusCode != http.StatusServiceUnavailable {
		t.Fatalf("status code = %d, want %d", statusCode, http.StatusServiceUnavailable)
	}
	if response.Status != "degraded" {
		t.Fatalf("health status = %q, want degraded", response.Status)
	}
	if response.Circuit.State != gobreaker.StateOpen.String() {
		t.Fatalf("circuit state = %q, want %q", response.Circuit.State, gobreaker.StateOpen.String())
	}
}

func TestThrottleTrackerRecordRedactsClientIdentity(t *testing.T) {
	tracker := NewThrottleTracker()
	var buffer bytes.Buffer
	originalWriter := log.Writer()
	log.SetOutput(&buffer)
	defer log.SetOutput(originalWriter)

	tracker.Record(ThrottleReasonIPRate, 2*time.Second, "fixture-key", "203.0.113.10", "IP rate limit exceeded")

	snapshot := tracker.Snapshot()
	if snapshot.Total != 1 || snapshot.IPRate != 1 {
		t.Fatalf("snapshot = %#v, want total=1 ipRate=1", snapshot)
	}
	output := buffer.String()
	if !strings.Contains(output, "client_fingerprint=") {
		t.Fatalf("expected client fingerprint in log output: %s", output)
	}
	if strings.Contains(output, "203.0.113.10") {
		t.Fatalf("expected redacted client identity, got raw ip in log output: %s", output)
	}
	if !strings.Contains(output, "api_key_present=true") {
		t.Fatalf("expected api_key_present in log output: %s", output)
	}
}
