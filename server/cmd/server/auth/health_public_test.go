package auth

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toolplane/pkg/model"
	"toolplane/pkg/trace"
)

// authenticatingAuthorizer builds an authorizer whose authenticator only
// accepts the fixed bootstrap token — session-key mode with a live
// credential check, the strictest configuration.
func authenticatingAuthorizer() *APIKeyAuthorizer {
	return NewAPIKeyAuthorizer(
		func(_ context.Context, token string) (*model.AuthPrincipal, error) {
			if token != "bootstrap-secret" {
				return nil, status.Error(codes.Unauthenticated, "unknown token")
			}
			return &model.AuthPrincipal{Mode: model.AuthModeFixed}, nil
		},
		trace.NopTracer(),
	)
}

func TestPublicMethodsSkipAuthentication(t *testing.T) {
	authorizer := authenticatingAuthorizer()
	interceptor := authorizer.UnaryInterceptor()

	// No credentials in the context at all: the standard health probe must
	// still reach its handler, or every load balancer in auth-enabled mode
	// would see the server as down.
	invoked := false
	_, err := interceptor(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/grpc.health.v1.Health/Check"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			invoked = true
			return &healthCheckResponse{}, nil
		},
	)
	if err != nil {
		t.Fatalf("unauthenticated health check rejected: %v", err)
	}
	if !invoked {
		t.Fatal("health check handler never ran")
	}

	// A wrong credential on a public method must not change the outcome.
	_, err = interceptor(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/grpc.health.v1.Health/Check"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return &healthCheckResponse{}, nil
		},
	)
	if err != nil {
		t.Fatalf("health check with (bad) credentials rejected: %v", err)
	}
}

func TestNonPublicMethodsStillRequireAuthentication(t *testing.T) {
	authorizer := authenticatingAuthorizer()
	interceptor := authorizer.UnaryInterceptor()

	_, err := interceptor(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/api.ToolService/ListTools"},
		func(ctx context.Context, req interface{}) (interface{}, error) { return nil, nil },
	)
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("unauthenticated API call code = %v, want Unauthenticated", status.Code(err))
	}
}

// healthCheckResponse mirrors grpc.health.v1's response without importing
// the health package here (the interceptor never inspects it).
type healthCheckResponse struct{}
