package auth

import (
	"context"
	"sync"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"toolplane/pkg/model"
	"toolplane/pkg/trace"
	proto "toolplane/proto"
)

type recordingTracer struct {
	mu     sync.Mutex
	events []trace.SessionEvent
}

func (t *recordingTracer) Record(event trace.SessionEvent) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, event)
}

func (t *recordingTracer) hasEvent(eventType trace.SessionEventType) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, event := range t.events {
		if event.Event == eventType {
			return true
		}
	}
	return false
}

func TestRequireSessionCapabilityEnforcesCapabilityAndScope(t *testing.T) {
	principal := &model.AuthPrincipal{
		Mode:         model.AuthModeSessionKey,
		SessionID:    "session-1",
		Capabilities: []model.APIKeyCapability{model.APIKeyCapabilityExecute},
	}
	ctx := context.WithValue(context.Background(), authPrincipalContextKey, principal)

	if err := RequireSessionCapability(ctx, "session-1", model.APIKeyCapabilityExecute); err != nil {
		t.Fatalf("RequireSessionCapability returned unexpected error: %v", err)
	}

	err := RequireSessionCapability(ctx, "session-2", model.APIKeyCapabilityExecute)
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("RequireSessionCapability mismatch error = %v, want permission denied", err)
	}

	err = RequireSessionCapability(ctx, "session-1", model.APIKeyCapabilityAdmin)
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("RequireSessionCapability capability error = %v, want permission denied", err)
	}
}

func TestAPIKeyAuthorizerUnaryInterceptorRecordsValidationAndDenial(t *testing.T) {
	tracer := &recordingTracer{}
	principal := &model.AuthPrincipal{
		Mode:         model.AuthModeSessionKey,
		SessionID:    "session-1",
		UserID:       "user-1",
		KeyID:        "key-1",
		TokenPreview: "toolplane_1234",
		Capabilities: []model.APIKeyCapability{model.APIKeyCapabilityAdmin},
	}
	authorizer := NewAPIKeyAuthorizer(func(ctx context.Context, token string) (*model.AuthPrincipal, error) {
		if token != "token-1" {
			t.Fatalf("authenticate token = %q, want token-1", token)
		}
		return principal, nil
	}, tracer)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("api_key", "token-1"))
	handlerCalled := false
	_, err := authorizer.UnaryInterceptor()(ctx, &proto.CreateSessionRequest{UserId: "user-1", SessionId: "session-2"}, &grpc.UnaryServerInfo{FullMethod: "/api.SessionsService/CreateSession"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		resolved, ok := PrincipalFromContext(ctx)
		if !ok || resolved == nil || resolved.KeyID != "key-1" {
			t.Fatalf("PrincipalFromContext = %#v, want key-1 principal", resolved)
		}
		return &proto.CreateSessionResponse{Session: &proto.Session{Id: "session-2"}}, nil
	})
	if err != nil {
		t.Fatalf("UnaryInterceptor returned unexpected error: %v", err)
	}
	if !handlerCalled {
		t.Fatal("UnaryInterceptor did not call handler")
	}
	if !tracer.hasEvent(trace.EventAuthValidated) {
		t.Fatal("expected auth_validated event")
	}

	deniedTracer := &recordingTracer{}
	deniedAuthorizer := NewAPIKeyAuthorizer(func(ctx context.Context, token string) (*model.AuthPrincipal, error) {
		return &model.AuthPrincipal{
			Mode:         model.AuthModeSessionKey,
			SessionID:    "session-1",
			UserID:       "user-1",
			KeyID:        "key-2",
			TokenPreview: "toolplane_5678",
			Capabilities: []model.APIKeyCapability{model.APIKeyCapabilityRead},
		}, nil
	}, deniedTracer)

	_, err = deniedAuthorizer.UnaryInterceptor()(ctx, &proto.CreateApiKeyRequest{SessionId: "session-1", Name: "denied"}, &grpc.UnaryServerInfo{FullMethod: "/api.SessionsService/CreateApiKey"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler should not be called when authorization fails")
		return nil, nil
	})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("denied UnaryInterceptor error = %v, want permission denied", err)
	}
	if !deniedTracer.hasEvent(trace.EventAuthPolicyDenied) {
		t.Fatal("expected auth_policy_denied event")
	}
}
