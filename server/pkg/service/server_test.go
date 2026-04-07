package service

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"toolplane/pkg/trace"
	proto "toolplane/proto"
)

func TestGRPCServerCreateApiKeyReturnsInvalidArgumentForUnsupportedCapabilities(t *testing.T) {
	sessionService := NewSessionsService(trace.NopTracer(), nil)
	session, err := sessionService.CreateSession("user-audit", "Auth Session", "session auth coverage", "", "", "tenant-a")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	server := NewGRPCServer(nil, sessionService, nil, nil, nil)
	_, err = server.CreateApiKey(context.Background(), &proto.CreateApiKeyRequest{
		SessionId:    session.ID,
		Name:         "invalid-capability",
		Capabilities: []string{"bogus"},
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("CreateApiKey error = %v, want invalid argument", err)
	}
}
