package service

import (
	"context"
	"strings"
	"testing"
	"time"

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

// TestInvokeToolLongPollReturnsTerminalResult pins the server-side
// long-poll: with wait_timeout_seconds set, InvokeTool blocks until the
// provider completes and returns the terminal payload with result/error
// populated — the ExecuteToolResponse fields that used to be dead on the
// wire.
func TestInvokeToolLongPollReturnsTerminalResult(t *testing.T) {
	server, _, machineService, requestService, _, machineID := newTaxonomyTestServer(t)
	const sessionID = "sess-taxonomy"

	// Provider: claim the request and complete it shortly after.
	go func() {
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			claimed, err := requestService.ClaimPendingRequest(sessionID, machineID, []string{"echo"})
			if err != nil || claimed == nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			_, _ = requestService.UpdateRequest(sessionID, claimed.ID, machineID, claimed.LeaseEpoch, model.RequestStatusRunning, nil, "")
			time.Sleep(100 * time.Millisecond)
			_ = requestService.SubmitRequestResult(sessionID, claimed.ID, machineID, claimed.LeaseEpoch,
				map[string]string{"echo": "hello"}, model.ResultTypeResolution, nil)
			return
		}
	}()

	resp, err := server.InvokeTool(context.Background(), &proto.ExecuteToolRequest{
		SessionId:          sessionID,
		ToolName:           "echo",
		Input:              `{}`,
		WaitTimeoutSeconds: 5,
	})
	if err != nil {
		t.Fatalf("long-poll invoke: %v", err)
	}
	if resp.Status != proto.RequestStatus_REQUEST_STATUS_DONE {
		t.Fatalf("long-poll status=%v want DONE", resp.Status)
	}
	if resp.Result == "" {
		t.Fatal("long-poll returned empty result")
	}
	if !strings.Contains(resp.Result, "hello") {
		t.Fatalf("result %q missing provider payload", resp.Result)
	}

	// Sanity: the machine is available again for a second call.
	if _, err := machineService.GetMachineByID(sessionID, machineID); err != nil {
		t.Fatalf("machine lookup after long-poll: %v", err)
	}
}

// TestInvokeToolLongPollDeadlineReturnsInFlight pins the deadline path: when
// the wait elapses before completion, InvokeTool returns the in-flight state
// and leaves the request running — the deadline never cancels the work.
func TestInvokeToolLongPollDeadlineReturnsInFlight(t *testing.T) {
	server, _, _, _, _, _ := newTaxonomyTestServer(t)
	const sessionID = "sess-taxonomy"

	resp, err := server.InvokeTool(context.Background(), &proto.ExecuteToolRequest{
		SessionId:          sessionID,
		ToolName:           "echo",
		Input:              `{}`,
		WaitTimeoutSeconds: 1,
	})
	if err != nil {
		t.Fatalf("long-poll deadline invoke: %v", err)
	}
	if resp.Status != proto.RequestStatus_REQUEST_STATUS_PENDING {
		t.Fatalf("deadline status=%v want PENDING", resp.Status)
	}
	if resp.RequestId == "" {
		t.Fatal("deadline response missing request id")
	}
}
