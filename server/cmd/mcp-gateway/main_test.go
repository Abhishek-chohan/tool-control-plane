package main

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
)

// TestUnaryDeadlineScopedByMethod pins the facade's mirror of cmd/proxy's
// per-method dial policy: the wait entrypoints get the backstop above the
// server's 3600s wait ceiling, everything else keeps the 30s default.
func TestUnaryDeadlineScopedByMethod(t *testing.T) {
	for _, method := range []string{
		"/api.v1.ToolService/InvokeTool",
		"/api.v1.ToolService/ExecuteTool",
	} {
		if got := unaryDeadline(method); got != waitMethodBackstop {
			t.Fatalf("unaryDeadline(%s) = %v, want the wait backstop %v", method, got, waitMethodBackstop)
		}
	}
	if got := unaryDeadline("/api.v1.RequestsService/GetRequest"); got != gatewayUnaryDeadline {
		t.Fatalf("unaryDeadline(GetRequest) = %v, want the %v default", got, gatewayUnaryDeadline)
	}
	if waitMethodBackstop <= time.Hour {
		t.Fatalf("wait backstop %v must sit above the server's 1h wait ceiling", waitMethodBackstop)
	}
}

// TestUnaryDeadlineInterceptorDefaultsOnly: the policy is a default — an
// existing deadline (the facade's sync-timeout budget) is never extended
// or shortened.
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

	if got := run(context.Background(), "/api.v1.ToolService/InvokeTool"); got <= time.Hour || got > waitMethodBackstop {
		t.Fatalf("wait method without deadline got %v, want ~%v", got, waitMethodBackstop)
	}

	existing, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if got := run(existing, "/api.v1.ToolService/InvokeTool"); got > 2*time.Second {
		t.Fatalf("existing 2s deadline was extended to %v; the policy is a default, not an override", got)
	}
}
