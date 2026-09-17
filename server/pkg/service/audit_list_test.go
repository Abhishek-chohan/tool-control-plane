package service

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"toolplane/internal/auth"
	"toolplane/pkg/model"
	"toolplane/pkg/storage/memory"
	proto "toolplane/proto"
)

// TestListAuditEventsHandler exercises the gRPC handler through a memory
// store: recorded events come back newest-first with their actor, filters
// narrow the set, pagination totals are stable, and a store-less
// deployment serves an empty trail (audit is store-backed by definition).
func TestListAuditEventsHandler(t *testing.T) {
	store := memory.New()
	tracer := &recordingTracer{}
	sessionSvc := NewSessionsService(tracer, store)
	svc := NewGRPCServer(nil, sessionSvc, nil, nil, nil)

	principal := &model.AuthPrincipal{
		Mode:         model.AuthModeSessionKey,
		SessionID:    "sess-audit-handler",
		KeyID:        "key-admin",
		Capabilities: []model.APIKeyCapability{model.APIKeyCapabilityAdmin},
	}
	handlerCtx := auth.NewContext(context.Background(), principal)

	// The audit recorder writes asynchronously; handler coverage records
	// through the store directly (the same sink the recorder persists to).
	record := func(event string) {
		t.Helper()
		if err := store.RecordAuditEvent(context.Background(), &model.AuditEvent{
			Event:      event,
			SessionID:  "sess-audit-handler",
			ActorKeyID: "key-admin",
		}); err != nil {
			t.Fatalf("record: %v", err)
		}
	}
	record("api_key_created")
	record("api_key_revoked")

	resp, err := svc.ListAuditEvents(handlerCtx, &proto.ListAuditEventsRequest{
		SessionId: "sess-audit-handler",
	})
	if err != nil {
		t.Fatalf("list audit events: %v", err)
	}
	if len(resp.Events) != 2 || resp.Page.TotalSize != 2 {
		t.Fatalf("events=%d total=%d, want 2/2", len(resp.Events), resp.Page.TotalSize)
	}
	if resp.Events[0].Event != "api_key_revoked" {
		t.Fatalf("newest = %q, want api_key_revoked", resp.Events[0].Event)
	}
	if resp.Events[0].ActorKeyId != "key-admin" {
		t.Fatalf("actor = %q, want key-admin", resp.Events[0].ActorKeyId)
	}

	// A filter matching nothing returns an empty page, not an error.
	empty, err := svc.ListAuditEvents(handlerCtx, &proto.ListAuditEventsRequest{
		SessionId: "sess-other",
	})
	if err != nil {
		t.Fatalf("filtered list: %v", err)
	}
	if len(empty.Events) != 0 || empty.Page.TotalSize != 0 {
		t.Fatalf("filtered events=%d total=%d, want 0/0", len(empty.Events), empty.Page.TotalSize)
	}

	// Store-less deployments serve an empty trail.
	bare := NewGRPCServer(nil, NewSessionsService(tracer, nil), nil, nil, nil)
	bareResp, err := bare.ListAuditEvents(handlerCtx, &proto.ListAuditEventsRequest{})
	if err != nil {
		t.Fatalf("store-less list: %v", err)
	}
	if len(bareResp.Events) != 0 {
		t.Fatalf("store-less events=%d, want 0", len(bareResp.Events))
	}
}

// TestListAuditEventsRequiresAdminCapability pins the authz row through
// the real interceptor: a read-only session key is denied; the admin key
// passes through to the handler.
func TestListAuditEventsRequiresAdminCapability(t *testing.T) {
	denyAuthorizer := auth.NewAPIKeyAuthorizer(func(ctx context.Context, token string) (*model.AuthPrincipal, error) {
		return &model.AuthPrincipal{Mode: model.AuthModeSessionKey, SessionID: "session-1", KeyID: "key-read", Capabilities: []model.APIKeyCapability{model.APIKeyCapabilityRead}}, nil
	}, nil)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("api_key", "token-1"))
	req := &proto.ListAuditEventsRequest{SessionId: "session-1"}
	_, err := denyAuthorizer.UnaryInterceptor()(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/api.v1.SessionsService/ListAuditEvents"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler should not run for a read-only principal")
		return nil, nil
	})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("read-only principal: %v, want PermissionDenied", err)
	}

	allowAuthorizer := auth.NewAPIKeyAuthorizer(func(ctx context.Context, token string) (*model.AuthPrincipal, error) {
		return &model.AuthPrincipal{Mode: model.AuthModeSessionKey, SessionID: "session-1", KeyID: "key-admin", Capabilities: []model.APIKeyCapability{model.APIKeyCapabilityAdmin}}, nil
	}, nil)
	adminCtx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("api_key", "admin-token"))
	_, err = allowAuthorizer.UnaryInterceptor()(adminCtx, req, &grpc.UnaryServerInfo{FullMethod: "/api.v1.SessionsService/ListAuditEvents"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return &proto.ListAuditEventsResponse{}, nil
	})
	if err != nil {
		t.Fatalf("admin principal denied: %v", err)
	}
}
