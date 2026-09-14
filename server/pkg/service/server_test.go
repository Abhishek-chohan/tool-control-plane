package service

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"toolplane/pkg/model"
	"toolplane/pkg/trace"
	proto "toolplane/proto"
)

func TestGRPCServerCreateApiKeyReturnsInvalidArgumentForUnsupportedCapabilities(t *testing.T) {
	sessionService := NewSessionsService(trace.NopTracer(), nil)
	session, err := sessionService.CreateSession("user-audit", "Auth Session", "session auth coverage", "", "tenant-a")
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

// TestGRPCServerListRequestsPagination pins the shared ListPage contract on
// request listings: the trailer is attached to the response, total_size is
// the filtered total, and the next-page token disappears exactly when the
// offset passes the end of the list — including on a page that is full but
// final.
func TestGRPCServerListRequestsPagination(t *testing.T) {
	ctx := context.Background()
	toolSvc := NewToolService(trace.NopTracer(), nil)
	machineSvc := NewMachinesService(ctx, toolSvc, trace.NopTracer(), nil)
	requestSvc := NewRequestsService(ctx, toolSvc, machineSvc, trace.NopTracer(), nil)
	server := NewGRPCServer(toolSvc, nil, machineSvc, requestSvc, nil)

	echoTool := model.NewTool("sess-page", "machine-page", "echo", "d", `{}`, nil, nil)
	if _, err := machineSvc.RegisterMachine("sess-page", "machine-page", "1.0", "go", "127.0.0.1", []*model.Tool{echoTool}, ""); err != nil {
		t.Fatalf("RegisterMachine: %v", err)
	}
	for i := 0; i < 15; i++ {
		if _, err := requestSvc.CreateRequest("sess-page", "echo", `{"n":1}`, 0, ""); err != nil {
			t.Fatalf("CreateRequest %d: %v", i, err)
		}
	}

	first, err := server.ListRequests(ctx, &proto.ListRequestsRequest{SessionId: "sess-page", PageSize: 10})
	if err != nil {
		t.Fatalf("ListRequests page 1: %v", err)
	}
	if len(first.Requests) != 10 {
		t.Fatalf("page 1 rows=%d want 10", len(first.Requests))
	}
	if first.Page == nil || first.Page.TotalSize != 15 {
		t.Fatalf("page 1 total_size=%v want 15", first.Page)
	}
	if first.Page.GetNextPageToken() == "" {
		t.Fatal("page 1 must carry a next_page_token")
	}

	second, err := server.ListRequests(ctx, &proto.ListRequestsRequest{
		SessionId: "sess-page",
		PageSize:  10,
		PageToken: first.Page.GetNextPageToken(),
	})
	if err != nil {
		t.Fatalf("ListRequests page 2: %v", err)
	}
	if len(second.Requests) != 5 {
		t.Fatalf("page 2 rows=%d want 5", len(second.Requests))
	}
	if second.Page == nil || second.Page.TotalSize != 15 {
		t.Fatalf("page 2 total_size=%v want 15", second.Page)
	}
	if second.Page.GetNextPageToken() != "" {
		t.Fatal("page 2 is the last page: next_page_token must be empty")
	}
}

// TestGRPCServerListRequestsFullFinalPage asserts the exact-multiple edge: a
// full page that is also the last page must not emit a next-page token (the
// old page-length heuristic sent clients chasing an empty extra page).
func TestGRPCServerListRequestsFullFinalPage(t *testing.T) {
	ctx := context.Background()
	toolSvc := NewToolService(trace.NopTracer(), nil)
	machineSvc := NewMachinesService(ctx, toolSvc, trace.NopTracer(), nil)
	requestSvc := NewRequestsService(ctx, toolSvc, machineSvc, trace.NopTracer(), nil)
	server := NewGRPCServer(toolSvc, nil, machineSvc, requestSvc, nil)

	echoTool := model.NewTool("sess-page2", "machine-page2", "echo", "d", `{}`, nil, nil)
	if _, err := machineSvc.RegisterMachine("sess-page2", "machine-page2", "1.0", "go", "127.0.0.1", []*model.Tool{echoTool}, ""); err != nil {
		t.Fatalf("RegisterMachine: %v", err)
	}
	for i := 0; i < 10; i++ {
		if _, err := requestSvc.CreateRequest("sess-page2", "echo", `{"n":1}`, 0, ""); err != nil {
			t.Fatalf("CreateRequest %d: %v", i, err)
		}
	}

	page, err := server.ListRequests(ctx, &proto.ListRequestsRequest{SessionId: "sess-page2", PageSize: 10})
	if err != nil {
		t.Fatalf("ListRequests: %v", err)
	}
	if len(page.Requests) != 10 {
		t.Fatalf("rows=%d want 10", len(page.Requests))
	}
	if page.Page == nil || page.Page.GetNextPageToken() != "" {
		t.Fatalf("full final page must not carry a next_page_token (page=%+v)", page.Page)
	}
}

// TestGRPCServerCreateSessionCollisionBareAlreadyExists pins the
// existence-oracle guard: a duplicate CreateSession gets a bare
// AlreadyExists with NO session payload (the response must be nil), so a
// caller cannot harvest another tenant's session metadata by guessing IDs.
func TestGRPCServerCreateSessionCollisionBareAlreadyExists(t *testing.T) {
	sessionService := NewSessionsService(trace.NopTracer(), nil)
	first, err := sessionService.CreateSession("owner-user", "Private Session", "owner metadata", "sess-collision", "tenant-a")
	if err != nil {
		t.Fatalf("create initial session: %v", err)
	}

	server := NewGRPCServer(nil, sessionService, nil, nil, nil)
	resp, err := server.CreateSession(context.Background(), &proto.CreateSessionRequest{
		SessionId: "sess-collision",
		UserId:    "attacker-user",
		Name:      "Attacker Session",
	})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("collision error = %v, want AlreadyExists", err)
	}
	if resp != nil && resp.Session != nil {
		t.Fatalf("collision response leaked session payload: %+v", resp.Session)
	}

	// The owner's session is untouched and still readable by its ID.
	stored, err := sessionService.GetSessionByID("sess-collision")
	if err != nil {
		t.Fatalf("get session after collision: %v", err)
	}
	if stored.ID != first.ID || stored.Name != "Private Session" {
		t.Fatalf("session mutated by collision: %+v", stored)
	}
}
