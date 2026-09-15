package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestProxyBreakerCountsOnlyGatewayFailures drives the root handler against a
// backend that answers 500 (client-induced/app failure) and 503 (backend
// unavailable): 500s must not feed the breaker, while 503s trip it.
func TestProxyBreakerCountsOnlyGatewayFailures(t *testing.T) {
	backendStatus := 500
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(backendStatus)
	})

	cfg := proxyConfig{environment: "development"}
	breaker := NewCircuitBreakerManager(10)
	handler := newProxyRootHandler(cfg, breaker, nil, nil, backend)

	// 500s are app failures, not backend-health failures: the breaker stays
	// closed across far more than the trip threshold (5 consecutive).
	for i := 0; i < 10; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api.v1/InvokeTool", nil))
		if rec.Code != 500 {
			t.Fatalf("request %d: status=%d want 500", i, rec.Code)
		}
	}
	if breaker.IsOpen() {
		t.Fatal("breaker opened on 500 responses (500s must not count)")
	}

	// A single 503 is a gateway-visible backend failure; the 500s above
	// latched the cold-start guard, so 503s count toward the breaker and
	// five consecutive ones trip.
	backendStatus = 503
	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api.v1/InvokeTool", nil))
	}
	if !breaker.IsOpen() {
		t.Fatal("breaker did not trip after 5 consecutive 503s")
	}
}

// TestExtractAPIKeyBearerCasing pins RFC 7235 case-insensitive scheme
// handling for the Bearer fallback.
func TestExtractAPIKeyBearerCasing(t *testing.T) {
	cases := []struct {
		header string
		want   string
	}{
		{"Bearer abc", "abc"},
		{"bearer abc", "abc"},
		{"BEARER abc", "abc"},
		{"bEaReR  spaced  ", "spaced"},
		{"Basic abc", ""},
		{"Bearer", ""},
		{"", ""},
	}
	for _, tc := range cases {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		if tc.header != "" {
			r.Header.Set("Authorization", tc.header)
		}
		if got := extractAPIKey(r); got != tc.want {
			t.Errorf("extractAPIKey(%q) = %q, want %q", tc.header, got, tc.want)
		}
	}
}

// TestProxyBreakerColdStartDoesNotTrip pins the cold-start latch: before the
// first successful backend response, gateway-visible failures (503) must not
// open the breaker — a cold gateway that answers 503 while the channel dials
// would otherwise fail its own readiness probe and loop restarts.
func TestProxyBreakerColdStartDoesNotTrip(t *testing.T) {
	backendStatus := http.StatusServiceUnavailable
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(backendStatus)
	})

	cfg := proxyConfig{environment: "development"}
	breaker := NewCircuitBreakerManager(10)
	handler := newProxyRootHandler(cfg, breaker, nil, nil, backend)

	// Far more than the post-latch trip threshold (5 consecutive), with no
	// success ever observed: the breaker stays closed and traffic keeps
	// reaching the backend.
	for i := 0; i < 15; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api.v1/InvokeTool", nil))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("request %d: status=%d want 503 (backend must still receive traffic)", i, rec.Code)
		}
	}
	if breaker.IsOpen() {
		t.Fatal("breaker opened on cold-start failures before any success")
	}

	// The first success latches the backend as up; from then on the usual
	// rule applies and 5 consecutive gateway-visible failures trip.
	backendStatus = http.StatusInternalServerError
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api.v1/InvokeTool", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("latch request: status=%d want 200-family pass-through", rec.Code)
	}

	backendStatus = http.StatusServiceUnavailable
	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api.v1/InvokeTool", nil))
	}
	if !breaker.IsOpen() {
		t.Fatal("breaker did not trip after 5 consecutive 503s once latched")
	}
}
