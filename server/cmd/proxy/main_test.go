package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"google.golang.org/grpc"
)

// TestUnaryDeadlineScopedByMethod pins the per-method dial policy: the two
// wait entrypoints get the wait backstop (which must sit above the
// server's 3600s wait ceiling), everything else keeps the 30s fast-call
// default.
func TestUnaryDeadlineScopedByMethod(t *testing.T) {
	for _, method := range []string{
		"/api.v1.ToolService/InvokeTool",
		"/api.v1.ToolService/ExecuteTool",
	} {
		if got := unaryDeadline(method); got != waitMethodBackstop {
			t.Fatalf("unaryDeadline(%s) = %v, want the wait backstop %v", method, got, waitMethodBackstop)
		}
	}
	if got := unaryDeadline("/api.v1.RequestsService/ListRequests"); got != proxyUnaryDeadline {
		t.Fatalf("unaryDeadline(ListRequests) = %v, want the %v default", got, proxyUnaryDeadline)
	}
	if waitMethodBackstop <= time.Hour {
		t.Fatalf("wait backstop %v must sit above the server's 1h wait ceiling", waitMethodBackstop)
	}
}

// TestUnaryDeadlineInterceptorDefaultsOnly pins the interceptor contract:
// with no deadline it applies the method's bound; an existing deadline
// (an HTTP client's own timeout) is never extended or shortened.
func TestUnaryDeadlineInterceptorDefaultsOnly(t *testing.T) {
	run := func(ctx context.Context, method string) time.Duration {
		var observed time.Duration
		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			deadline, _ := ctx.Deadline()
			observed = time.Until(deadline)
			return nil
		}
		_ = unaryDeadlineInterceptor(ctx, method, nil, nil, nil, invoker)
		return observed
	}

	if got := run(context.Background(), "/api.v1.ToolService/ExecuteTool"); got <= time.Hour || got > waitMethodBackstop {
		t.Fatalf("wait method without deadline got %v, want ~%v", got, waitMethodBackstop)
	}
	if got := run(context.Background(), "/api.v1.RequestsService/GetRequest"); got <= 25*time.Second || got > proxyUnaryDeadline {
		t.Fatalf("fast method without deadline got %v, want ~%v", got, proxyUnaryDeadline)
	}

	existing, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if got := run(existing, "/api.v1.ToolService/ExecuteTool"); got > 2*time.Second {
		t.Fatalf("existing 2s deadline was extended to %v; the policy is a default, not an override", got)
	}
}

// TestDeadlinePolicyMiddlewareStripsGrpcTimeout pins the edge ownership of
// deadlines: the gateway runtime converts Grpc-Timeout into the context
// deadline before the dial interceptor runs, so any HTTP caller could
// otherwise raise or shrink the effective bound (or kill a streaming
// replay). The middleware removes the header before the gateway mux sees
// the request.
func TestDeadlinePolicyMiddlewareStripsGrpcTimeout(t *testing.T) {
	var seen http.Header
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api.v1/ExecuteTool", nil)
	request.Header.Set("Grpc-Timeout", "24H")
	request.Header.Set("Content-Type", "application/json")

	deadlinePolicyMiddleware(inner).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("inner handler status = %d, want 200", recorder.Code)
	}
	if seen.Get("Grpc-Timeout") != "" {
		t.Fatal("Grpc-Timeout must be stripped before the gateway mux parses it")
	}
	if seen.Get("Content-Type") != "application/json" {
		t.Fatal("unrelated headers must pass through untouched")
	}
}
