package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	pb "toolplane-go-client/proto"
)

type requestsServiceClientStub struct {
	createRequestFunc func(ctx context.Context, in *pb.CreateRequestRequest, opts ...grpc.CallOption) (*pb.Request, error)
	getRequestFunc    func(ctx context.Context, in *pb.GetRequestRequest, opts ...grpc.CallOption) (*pb.Request, error)
	listRequestsFunc  func(ctx context.Context, in *pb.ListRequestsRequest, opts ...grpc.CallOption) (*pb.ListRequestsResponse, error)
	cancelRequestFunc func(ctx context.Context, in *pb.CancelRequestRequest, opts ...grpc.CallOption) (*pb.CancelRequestResponse, error)
}

type toolServiceClientStub struct {
	getToolByIDFunc   func(ctx context.Context, in *pb.GetToolByIdRequest, opts ...grpc.CallOption) (*pb.GetToolResponse, error)
	getToolByNameFunc func(ctx context.Context, in *pb.GetToolByNameRequest, opts ...grpc.CallOption) (*pb.GetToolResponse, error)
	deleteToolFunc    func(ctx context.Context, in *pb.DeleteToolRequest, opts ...grpc.CallOption) (*pb.DeleteToolResponse, error)
	executeToolFunc   func(ctx context.Context, in *pb.ExecuteToolRequest, opts ...grpc.CallOption) (*pb.ExecuteToolResponse, error)
}

type sessionsServiceClientStub struct {
	createSessionFunc func(ctx context.Context, in *pb.CreateSessionRequest, opts ...grpc.CallOption) (*pb.CreateSessionResponse, error)
	listSessionsFunc  func(ctx context.Context, in *pb.ListSessionsRequest, opts ...grpc.CallOption) (*pb.ListSessionsResponse, error)
	updateSessionFunc func(ctx context.Context, in *pb.UpdateSessionRequest, opts ...grpc.CallOption) (*pb.Session, error)
	createAPIKeyFunc  func(ctx context.Context, in *pb.CreateApiKeyRequest, opts ...grpc.CallOption) (*pb.ApiKey, error)
	listAPIKeysFunc   func(ctx context.Context, in *pb.ListApiKeysRequest, opts ...grpc.CallOption) (*pb.ListApiKeysResponse, error)
	revokeAPIKeyFunc  func(ctx context.Context, in *pb.RevokeApiKeyRequest, opts ...grpc.CallOption) (*pb.RevokeApiKeyResponse, error)
}

func unexpectedRequestCall(method string) error {
	return errors.New("unexpected " + method + " call")
}

func unexpectedToolCall(method string) error {
	return errors.New("unexpected " + method + " call")
}

func unexpectedSessionCall(method string) error {
	return errors.New("unexpected " + method + " call")
}

func (s *requestsServiceClientStub) CreateRequest(ctx context.Context, in *pb.CreateRequestRequest, opts ...grpc.CallOption) (*pb.Request, error) {
	if s.createRequestFunc != nil {
		return s.createRequestFunc(ctx, in, opts...)
	}
	return nil, unexpectedRequestCall("CreateRequest")
}

func (s *requestsServiceClientStub) GetRequest(ctx context.Context, in *pb.GetRequestRequest, opts ...grpc.CallOption) (*pb.Request, error) {
	if s.getRequestFunc != nil {
		return s.getRequestFunc(ctx, in, opts...)
	}
	return nil, unexpectedRequestCall("GetRequest")
}

func (s *requestsServiceClientStub) ListRequests(ctx context.Context, in *pb.ListRequestsRequest, opts ...grpc.CallOption) (*pb.ListRequestsResponse, error) {
	if s.listRequestsFunc != nil {
		return s.listRequestsFunc(ctx, in, opts...)
	}
	return nil, unexpectedRequestCall("ListRequests")
}

func (s *requestsServiceClientStub) UpdateRequest(ctx context.Context, in *pb.UpdateRequestRequest, opts ...grpc.CallOption) (*pb.Request, error) {
	return nil, unexpectedRequestCall("UpdateRequest")
}

func (s *requestsServiceClientStub) ClaimRequest(ctx context.Context, in *pb.ClaimRequestRequest, opts ...grpc.CallOption) (*pb.Request, error) {
	return nil, unexpectedRequestCall("ClaimRequest")
}

func (s *requestsServiceClientStub) CancelRequest(ctx context.Context, in *pb.CancelRequestRequest, opts ...grpc.CallOption) (*pb.CancelRequestResponse, error) {
	if s.cancelRequestFunc != nil {
		return s.cancelRequestFunc(ctx, in, opts...)
	}
	return nil, unexpectedRequestCall("CancelRequest")
}

func (s *requestsServiceClientStub) SubmitRequestResult(ctx context.Context, in *pb.SubmitRequestResultRequest, opts ...grpc.CallOption) (*pb.SubmitRequestResultResponse, error) {
	return nil, unexpectedRequestCall("SubmitRequestResult")
}

func (s *requestsServiceClientStub) AppendRequestChunks(ctx context.Context, in *pb.AppendRequestChunksRequest, opts ...grpc.CallOption) (*pb.AppendRequestChunksResponse, error) {
	return nil, unexpectedRequestCall("AppendRequestChunks")
}

func (s *requestsServiceClientStub) GetRequestChunks(ctx context.Context, in *pb.GetRequestChunksRequest, opts ...grpc.CallOption) (*pb.GetRequestChunksResponse, error) {
	return nil, unexpectedRequestCall("GetRequestChunks")
}

func (s *toolServiceClientStub) RegisterTool(ctx context.Context, in *pb.RegisterToolRequest, opts ...grpc.CallOption) (*pb.RegisterToolResponse, error) {
	return nil, unexpectedToolCall("RegisterTool")
}

func (s *toolServiceClientStub) ListTools(ctx context.Context, in *pb.ListToolsRequest, opts ...grpc.CallOption) (*pb.ListToolsResponse, error) {
	return nil, unexpectedToolCall("ListTools")
}

func (s *toolServiceClientStub) GetToolById(ctx context.Context, in *pb.GetToolByIdRequest, opts ...grpc.CallOption) (*pb.GetToolResponse, error) {
	if s.getToolByIDFunc != nil {
		return s.getToolByIDFunc(ctx, in, opts...)
	}
	return nil, unexpectedToolCall("GetToolById")
}

func (s *toolServiceClientStub) GetToolByName(ctx context.Context, in *pb.GetToolByNameRequest, opts ...grpc.CallOption) (*pb.GetToolResponse, error) {
	if s.getToolByNameFunc != nil {
		return s.getToolByNameFunc(ctx, in, opts...)
	}
	return nil, unexpectedToolCall("GetToolByName")
}

func (s *toolServiceClientStub) DeleteTool(ctx context.Context, in *pb.DeleteToolRequest, opts ...grpc.CallOption) (*pb.DeleteToolResponse, error) {
	if s.deleteToolFunc != nil {
		return s.deleteToolFunc(ctx, in, opts...)
	}
	return nil, unexpectedToolCall("DeleteTool")
}

func (s *toolServiceClientStub) UpdateToolPing(ctx context.Context, in *pb.UpdateToolPingRequest, opts ...grpc.CallOption) (*pb.Tool, error) {
	return nil, unexpectedToolCall("UpdateToolPing")
}

func (s *toolServiceClientStub) StreamExecuteTool(ctx context.Context, in *pb.ExecuteToolRequest, opts ...grpc.CallOption) (pb.ToolService_StreamExecuteToolClient, error) {
	return nil, unexpectedToolCall("StreamExecuteTool")
}

func (s *toolServiceClientStub) ResumeStream(ctx context.Context, in *pb.ResumeStreamRequest, opts ...grpc.CallOption) (pb.ToolService_ResumeStreamClient, error) {
	return nil, unexpectedToolCall("ResumeStream")
}

func (s *toolServiceClientStub) ExecuteTool(ctx context.Context, in *pb.ExecuteToolRequest, opts ...grpc.CallOption) (*pb.ExecuteToolResponse, error) {
	if s.executeToolFunc != nil {
		return s.executeToolFunc(ctx, in, opts...)
	}
	return nil, unexpectedToolCall("ExecuteTool")
}

func (s *toolServiceClientStub) HealthCheck(ctx context.Context, in *pb.HealthCheckRequest, opts ...grpc.CallOption) (*pb.HealthCheckResponse, error) {
	return nil, unexpectedToolCall("HealthCheck")
}

func (s *sessionsServiceClientStub) CreateSession(ctx context.Context, in *pb.CreateSessionRequest, opts ...grpc.CallOption) (*pb.CreateSessionResponse, error) {
	if s.createSessionFunc != nil {
		return s.createSessionFunc(ctx, in, opts...)
	}
	return nil, unexpectedSessionCall("CreateSession")
}

func (s *sessionsServiceClientStub) GetSession(ctx context.Context, in *pb.GetSessionRequest, opts ...grpc.CallOption) (*pb.Session, error) {
	return nil, unexpectedSessionCall("GetSession")
}

func (s *sessionsServiceClientStub) ListSessions(ctx context.Context, in *pb.ListSessionsRequest, opts ...grpc.CallOption) (*pb.ListSessionsResponse, error) {
	if s.listSessionsFunc != nil {
		return s.listSessionsFunc(ctx, in, opts...)
	}
	return nil, unexpectedSessionCall("ListSessions")
}

func (s *sessionsServiceClientStub) UpdateSession(ctx context.Context, in *pb.UpdateSessionRequest, opts ...grpc.CallOption) (*pb.Session, error) {
	if s.updateSessionFunc != nil {
		return s.updateSessionFunc(ctx, in, opts...)
	}
	return nil, unexpectedSessionCall("UpdateSession")
}

func (s *sessionsServiceClientStub) DeleteSession(ctx context.Context, in *pb.DeleteSessionRequest, opts ...grpc.CallOption) (*pb.DeleteSessionResponse, error) {
	return nil, unexpectedSessionCall("DeleteSession")
}

func (s *sessionsServiceClientStub) ListUserSessions(ctx context.Context, in *pb.ListUserSessionsRequest, opts ...grpc.CallOption) (*pb.ListUserSessionsResponse, error) {
	return nil, unexpectedSessionCall("ListUserSessions")
}

func (s *sessionsServiceClientStub) BulkDeleteSessions(ctx context.Context, in *pb.BulkDeleteSessionsRequest, opts ...grpc.CallOption) (*pb.BulkDeleteSessionsResponse, error) {
	return nil, unexpectedSessionCall("BulkDeleteSessions")
}

func (s *sessionsServiceClientStub) GetSessionStats(ctx context.Context, in *pb.GetSessionStatsRequest, opts ...grpc.CallOption) (*pb.GetSessionStatsResponse, error) {
	return nil, unexpectedSessionCall("GetSessionStats")
}

func (s *sessionsServiceClientStub) RefreshSessionToken(ctx context.Context, in *pb.RefreshSessionTokenRequest, opts ...grpc.CallOption) (*pb.RefreshSessionTokenResponse, error) {
	return nil, unexpectedSessionCall("RefreshSessionToken")
}

func (s *sessionsServiceClientStub) InvalidateSession(ctx context.Context, in *pb.InvalidateSessionRequest, opts ...grpc.CallOption) (*pb.InvalidateSessionResponse, error) {
	return nil, unexpectedSessionCall("InvalidateSession")
}

func (s *sessionsServiceClientStub) CreateApiKey(ctx context.Context, in *pb.CreateApiKeyRequest, opts ...grpc.CallOption) (*pb.ApiKey, error) {
	if s.createAPIKeyFunc != nil {
		return s.createAPIKeyFunc(ctx, in, opts...)
	}
	return nil, unexpectedSessionCall("CreateApiKey")
}

func (s *sessionsServiceClientStub) ListApiKeys(ctx context.Context, in *pb.ListApiKeysRequest, opts ...grpc.CallOption) (*pb.ListApiKeysResponse, error) {
	if s.listAPIKeysFunc != nil {
		return s.listAPIKeysFunc(ctx, in, opts...)
	}
	return nil, unexpectedSessionCall("ListApiKeys")
}

func (s *sessionsServiceClientStub) RevokeApiKey(ctx context.Context, in *pb.RevokeApiKeyRequest, opts ...grpc.CallOption) (*pb.RevokeApiKeyResponse, error) {
	if s.revokeAPIKeyFunc != nil {
		return s.revokeAPIKeyFunc(ctx, in, opts...)
	}
	return nil, unexpectedSessionCall("RevokeApiKey")
}

func newConnectedRequestClient(requestsClient pb.RequestsServiceClient) *ToolplaneClient {
	conn := &grpc.ClientConn{}
	return &ToolplaneClient{
		protocol:       ProtocolGRPC,
		sessionID:      "session-1",
		grpcConn:       conn,
		toolClient:     pb.NewToolServiceClient(conn),
		sessionClient:  pb.NewSessionsServiceClient(conn),
		machineClient:  pb.NewMachinesServiceClient(conn),
		requestsClient: requestsClient,
	}
}

func newConnectedToolClient(toolClient pb.ToolServiceClient) *ToolplaneClient {
	conn := &grpc.ClientConn{}
	return &ToolplaneClient{
		protocol:       ProtocolGRPC,
		sessionID:      "session-1",
		grpcConn:       conn,
		toolClient:     toolClient,
		sessionClient:  pb.NewSessionsServiceClient(conn),
		machineClient:  pb.NewMachinesServiceClient(conn),
		requestsClient: pb.NewRequestsServiceClient(conn),
	}
}

func newConnectedExecutionClient(toolClient pb.ToolServiceClient, requestsClient pb.RequestsServiceClient) *ToolplaneClient {
	conn := &grpc.ClientConn{}
	return &ToolplaneClient{
		protocol:       ProtocolGRPC,
		sessionID:      "session-1",
		grpcConn:       conn,
		toolClient:     toolClient,
		sessionClient:  pb.NewSessionsServiceClient(conn),
		machineClient:  pb.NewMachinesServiceClient(conn),
		requestsClient: requestsClient,
	}
}

func newConnectedSessionClient(sessionClient pb.SessionsServiceClient) *ToolplaneClient {
	conn := &grpc.ClientConn{}
	return &ToolplaneClient{
		protocol:       ProtocolGRPC,
		sessionID:      "session-1",
		userID:         "user-1",
		grpcConn:       conn,
		toolClient:     pb.NewToolServiceClient(conn),
		sessionClient:  sessionClient,
		machineClient:  pb.NewMachinesServiceClient(conn),
		requestsClient: pb.NewRequestsServiceClient(conn),
	}
}

func TestRequireSessionID(t *testing.T) {
	client := &ToolplaneClient{sessionID: "session-123"}

	sessionID, err := client.requireSessionID("ListMachines")
	if err != nil {
		t.Fatalf("requireSessionID returned unexpected error: %v", err)
	}
	if sessionID != "session-123" {
		t.Fatalf("requireSessionID returned %q, want session-123", sessionID)
	}

	client.sessionID = ""
	_, err = client.requireSessionID("ListMachines")
	if err == nil {
		t.Fatal("requireSessionID succeeded without a session ID")
	}
	if err.Error() != "ListMachines requires a session ID" {
		t.Fatalf("requireSessionID error = %q, want %q", err.Error(), "ListMachines requires a session ID")
	}
}

func TestResolveMachineID(t *testing.T) {
	client := &ToolplaneClient{machineID: "cached-machine"}

	machineID, err := client.resolveMachineID("DrainMachine", "explicit-machine")
	if err != nil {
		t.Fatalf("resolveMachineID returned unexpected error: %v", err)
	}
	if machineID != "explicit-machine" {
		t.Fatalf("resolveMachineID returned %q, want explicit-machine", machineID)
	}

	machineID, err = client.resolveMachineID("DrainMachine", "")
	if err != nil {
		t.Fatalf("resolveMachineID returned unexpected error: %v", err)
	}
	if machineID != "cached-machine" {
		t.Fatalf("resolveMachineID returned %q, want cached-machine", machineID)
	}

	client.machineID = ""
	_, err = client.resolveMachineID("DrainMachine", "")
	if err == nil {
		t.Fatal("resolveMachineID succeeded without a machine ID")
	}
	if err.Error() != "DrainMachine requires a machine ID or a registered machine" {
		t.Fatalf("resolveMachineID error = %q, want %q", err.Error(), "DrainMachine requires a machine ID or a registered machine")
	}
}

func TestGRPCContextAddsAPIKeyAndTimeout(t *testing.T) {
	client := &ToolplaneClient{apiKey: "secret-key"}

	ctx, cancel := client.grpcContext(context.TODO(), 250*time.Millisecond)
	defer cancel()

	metadataValues, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("grpcContext did not attach outgoing metadata")
	}
	if got := metadataValues.Get("api_key"); len(got) != 1 || got[0] != "secret-key" {
		t.Fatalf("grpcContext api_key metadata = %v, want [secret-key]", got)
	}
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		t.Fatal("grpcContext did not apply a timeout deadline")
	}
	if ctx.Err() != nil {
		t.Fatalf("grpcContext returned an already-cancelled context: %v", ctx.Err())
	}
}

func TestGRPCContextPreservesExistingDeadline(t *testing.T) {
	parent, parentCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer parentCancel()

	parentDeadline, hasDeadline := parent.Deadline()
	if !hasDeadline {
		t.Fatal("parent context missing deadline")
	}

	client := &ToolplaneClient{}
	ctx, cancel := client.grpcContext(parent, 5*time.Second)
	defer cancel()

	ctxDeadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		t.Fatal("grpcContext removed the parent deadline")
	}
	if !ctxDeadline.Equal(parentDeadline) {
		t.Fatalf("grpcContext changed deadline from %v to %v", parentDeadline, ctxDeadline)
	}
}

func TestParseNumericResult(t *testing.T) {
	testCases := []struct {
		name        string
		payload     string
		want        float64
		wantErrText string
	}{
		{name: "number", payload: "42", want: 42},
		{name: "numeric string", payload: `"42.5"`, want: 42.5},
		{name: "empty", payload: "", wantErrText: "empty result payload"},
		{name: "invalid", payload: `{"value":42}`, wantErrText: "failed to parse numeric result payload: {\"value\":42}"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := parseNumericResult(testCase.payload)
			if testCase.wantErrText != "" {
				if err == nil {
					t.Fatalf("parseNumericResult(%q) succeeded, want error", testCase.payload)
				}
				if err.Error() != testCase.wantErrText {
					t.Fatalf("parseNumericResult(%q) error = %q, want %q", testCase.payload, err.Error(), testCase.wantErrText)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseNumericResult(%q) returned unexpected error: %v", testCase.payload, err)
			}
			if got != testCase.want {
				t.Fatalf("parseNumericResult(%q) = %v, want %v", testCase.payload, got, testCase.want)
			}
		})
	}
}

func TestNewToolplaneClientRejectsUnsupportedProtocol(t *testing.T) {
	_, err := NewToolplaneClient(ClientProtocol("http"), "localhost", 9001, "session-1", "user-1", "")
	if err == nil {
		t.Fatal("NewToolplaneClient succeeded for unsupported protocol")
	}
	if err.Error() != "unsupported protocol: http" {
		t.Fatalf("NewToolplaneClient error = %q, want %q", err.Error(), "unsupported protocol: http")
	}
}

func TestExecuteToolUsesGRPCRequestLifecycle(t *testing.T) {
	requestPolls := 0
	client := newConnectedExecutionClient(
		&toolServiceClientStub{
			executeToolFunc: func(ctx context.Context, in *pb.ExecuteToolRequest, opts ...grpc.CallOption) (*pb.ExecuteToolResponse, error) {
				if in.SessionId != "session-1" {
					t.Fatalf("ExecuteTool SessionId = %q, want session-1", in.SessionId)
				}
				if in.ToolName != "demo_tool" {
					t.Fatalf("ExecuteTool ToolName = %q, want demo_tool", in.ToolName)
				}
				if in.Input != `{"message":"hello"}` {
					t.Fatalf("ExecuteTool Input = %q, want JSON payload", in.Input)
				}
				return &pb.ExecuteToolResponse{RequestId: "request-99"}, nil
			},
		},
		&requestsServiceClientStub{
			getRequestFunc: func(ctx context.Context, in *pb.GetRequestRequest, opts ...grpc.CallOption) (*pb.Request, error) {
				requestPolls++
				status := "running"
				result := ""
				if requestPolls > 1 {
					status = "done"
					result = `{"echo":"hello"}`
				}
				return &pb.Request{
					Id:        in.RequestId,
					SessionId: in.SessionId,
					ToolName:  "demo_tool",
					Status:    status,
					Result:    result,
				}, nil
			},
		},
	)

	request, err := client.ExecuteTool(context.Background(), "demo_tool", map[string]interface{}{"message": "hello"})
	if err != nil {
		t.Fatalf("ExecuteTool returned unexpected error: %v", err)
	}
	if request.Id != "request-99" || request.Status != "done" || request.Result != `{"echo":"hello"}` {
		t.Fatalf("ExecuteTool returned %#v, want request-99/done/{\"echo\":\"hello\"}", request)
	}
}

func TestCreateRequestUsesSessionID(t *testing.T) {
	client := newConnectedRequestClient(&requestsServiceClientStub{
		createRequestFunc: func(ctx context.Context, in *pb.CreateRequestRequest, opts ...grpc.CallOption) (*pb.Request, error) {
			if in.SessionId != "session-1" {
				t.Fatalf("CreateRequest SessionId = %q, want session-1", in.SessionId)
			}
			if in.ToolName != "echo" {
				t.Fatalf("CreateRequest ToolName = %q, want echo", in.ToolName)
			}
			if in.Input != `{"message":"hello"}` {
				t.Fatalf("CreateRequest Input = %q, want JSON payload", in.Input)
			}
			return &pb.Request{Id: "request-1", SessionId: in.SessionId, ToolName: in.ToolName, Status: "pending"}, nil
		},
	})

	request, err := client.CreateRequest("echo", `{"message":"hello"}`)
	if err != nil {
		t.Fatalf("CreateRequest returned unexpected error: %v", err)
	}
	if request.Id != "request-1" || request.ToolName != "echo" || request.Status != "pending" {
		t.Fatalf("CreateRequest returned %#v, want request-1/echo/pending", request)
	}
}

func TestGetRequestUsesSessionID(t *testing.T) {
	client := newConnectedRequestClient(&requestsServiceClientStub{
		getRequestFunc: func(ctx context.Context, in *pb.GetRequestRequest, opts ...grpc.CallOption) (*pb.Request, error) {
			if in.SessionId != "session-1" {
				t.Fatalf("GetRequest SessionId = %q, want session-1", in.SessionId)
			}
			if in.RequestId != "request-42" {
				t.Fatalf("GetRequest RequestId = %q, want request-42", in.RequestId)
			}
			return &pb.Request{Id: in.RequestId, SessionId: in.SessionId, ToolName: "echo", Status: "done"}, nil
		},
	})

	request, err := client.GetRequest("request-42")
	if err != nil {
		t.Fatalf("GetRequest returned unexpected error: %v", err)
	}
	if request.Id != "request-42" || request.Status != "done" {
		t.Fatalf("GetRequest returned %#v, want request-42/done", request)
	}
}

func TestListRequestsPassesFilters(t *testing.T) {
	client := newConnectedRequestClient(&requestsServiceClientStub{
		listRequestsFunc: func(ctx context.Context, in *pb.ListRequestsRequest, opts ...grpc.CallOption) (*pb.ListRequestsResponse, error) {
			if in.SessionId != "session-1" {
				t.Fatalf("ListRequests SessionId = %q, want session-1", in.SessionId)
			}
			if in.Status != "running" || in.ToolName != "echo" || in.Limit != 5 || in.Offset != 2 {
				t.Fatalf("ListRequests request = %#v, want status=running toolName=echo limit=5 offset=2", in)
			}
			return &pb.ListRequestsResponse{
				Requests: []*pb.Request{{Id: "request-1", Status: "running"}, {Id: "request-2", Status: "running"}},
			}, nil
		},
	})

	requests, err := client.ListRequests("running", "echo", 5, 2)
	if err != nil {
		t.Fatalf("ListRequests returned unexpected error: %v", err)
	}
	if len(requests) != 2 || requests[0].Id != "request-1" || requests[1].Id != "request-2" {
		t.Fatalf("ListRequests returned %#v, want two request IDs", requests)
	}
}

func TestCancelRequestReturnsSuccessFlag(t *testing.T) {
	client := newConnectedRequestClient(&requestsServiceClientStub{
		cancelRequestFunc: func(ctx context.Context, in *pb.CancelRequestRequest, opts ...grpc.CallOption) (*pb.CancelRequestResponse, error) {
			if in.SessionId != "session-1" {
				t.Fatalf("CancelRequest SessionId = %q, want session-1", in.SessionId)
			}
			if in.RequestId != "request-9" {
				t.Fatalf("CancelRequest RequestId = %q, want request-9", in.RequestId)
			}
			return &pb.CancelRequestResponse{Success: true}, nil
		},
	})

	success, err := client.CancelRequest("request-9")
	if err != nil {
		t.Fatalf("CancelRequest returned unexpected error: %v", err)
	}
	if !success {
		t.Fatal("CancelRequest returned false, want true")
	}
}

func TestGetToolByIDUsesSessionID(t *testing.T) {
	client := newConnectedToolClient(&toolServiceClientStub{
		getToolByIDFunc: func(ctx context.Context, in *pb.GetToolByIdRequest, opts ...grpc.CallOption) (*pb.GetToolResponse, error) {
			if in.SessionId != "session-1" {
				t.Fatalf("GetToolById SessionId = %q, want session-1", in.SessionId)
			}
			if in.ToolId != "tool-1" {
				t.Fatalf("GetToolById ToolId = %q, want tool-1", in.ToolId)
			}
			return &pb.GetToolResponse{Tool: &pb.Tool{Id: in.ToolId, Name: "echo"}}, nil
		},
	})

	tool, err := client.GetToolByID("tool-1")
	if err != nil {
		t.Fatalf("GetToolByID returned unexpected error: %v", err)
	}
	if tool.Id != "tool-1" || tool.Name != "echo" {
		t.Fatalf("GetToolByID returned %#v, want tool-1/echo", tool)
	}
}

func TestGetToolByNameUsesSessionID(t *testing.T) {
	client := newConnectedToolClient(&toolServiceClientStub{
		getToolByNameFunc: func(ctx context.Context, in *pb.GetToolByNameRequest, opts ...grpc.CallOption) (*pb.GetToolResponse, error) {
			if in.SessionId != "session-1" {
				t.Fatalf("GetToolByName SessionId = %q, want session-1", in.SessionId)
			}
			if in.ToolName != "echo" {
				t.Fatalf("GetToolByName ToolName = %q, want echo", in.ToolName)
			}
			return &pb.GetToolResponse{Tool: &pb.Tool{Id: "tool-1", Name: in.ToolName}}, nil
		},
	})

	tool, err := client.GetToolByName("echo")
	if err != nil {
		t.Fatalf("GetToolByName returned unexpected error: %v", err)
	}
	if tool.Id != "tool-1" || tool.Name != "echo" {
		t.Fatalf("GetToolByName returned %#v, want tool-1/echo", tool)
	}
}

func TestDeleteToolReturnsSuccessFlag(t *testing.T) {
	client := newConnectedToolClient(&toolServiceClientStub{
		deleteToolFunc: func(ctx context.Context, in *pb.DeleteToolRequest, opts ...grpc.CallOption) (*pb.DeleteToolResponse, error) {
			if in.SessionId != "session-1" {
				t.Fatalf("DeleteTool SessionId = %q, want session-1", in.SessionId)
			}
			if in.ToolId != "tool-1" {
				t.Fatalf("DeleteTool ToolId = %q, want tool-1", in.ToolId)
			}
			return &pb.DeleteToolResponse{Success: true}, nil
		},
	})

	success, err := client.DeleteTool("tool-1")
	if err != nil {
		t.Fatalf("DeleteTool returned unexpected error: %v", err)
	}
	if !success {
		t.Fatal("DeleteTool returned false, want true")
	}
}

func TestListSessionsUsesUserID(t *testing.T) {
	client := newConnectedSessionClient(&sessionsServiceClientStub{
		listSessionsFunc: func(ctx context.Context, in *pb.ListSessionsRequest, opts ...grpc.CallOption) (*pb.ListSessionsResponse, error) {
			if in.UserId != "user-1" {
				t.Fatalf("ListSessions UserId = %q, want user-1", in.UserId)
			}
			return &pb.ListSessionsResponse{Sessions: []*pb.Session{{Id: "session-1"}, {Id: "session-2"}}}, nil
		},
	})

	sessions, err := client.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions returned unexpected error: %v", err)
	}
	if len(sessions) != 2 || sessions[0].Id != "session-1" || sessions[1].Id != "session-2" {
		t.Fatalf("ListSessions returned %#v, want two session IDs", sessions)
	}
}

func TestUpdateSessionUsesCurrentSession(t *testing.T) {
	client := newConnectedSessionClient(&sessionsServiceClientStub{
		updateSessionFunc: func(ctx context.Context, in *pb.UpdateSessionRequest, opts ...grpc.CallOption) (*pb.Session, error) {
			if in.SessionId != "session-1" {
				t.Fatalf("UpdateSession SessionId = %q, want session-1", in.SessionId)
			}
			if in.Name != "updated" || in.Description != "desc" || in.Namespace != "ns" {
				t.Fatalf("UpdateSession request = %#v, want updated fields", in)
			}
			return &pb.Session{Id: in.SessionId, Name: in.Name, Description: in.Description, Namespace: in.Namespace}, nil
		},
	})

	session, err := client.UpdateSession("updated", "desc", "ns")
	if err != nil {
		t.Fatalf("UpdateSession returned unexpected error: %v", err)
	}
	if session.Name != "updated" || session.Namespace != "ns" {
		t.Fatalf("UpdateSession returned %#v, want updated/ns", session)
	}
}

func TestCreateSessionOmitsLegacyAPIKey(t *testing.T) {
	client := newConnectedSessionClient(&sessionsServiceClientStub{
		createSessionFunc: func(ctx context.Context, in *pb.CreateSessionRequest, opts ...grpc.CallOption) (*pb.CreateSessionResponse, error) {
			if in.UserId != "user-1" {
				t.Fatalf("CreateSession UserId = %q, want user-1", in.UserId)
			}
			if in.Name != "created" || in.Description != "desc" || in.Namespace != "ns" {
				t.Fatalf("CreateSession request = %#v, want created session fields", in)
			}
			if in.ApiKey != "" {
				t.Fatalf("CreateSession ApiKey = %q, want empty legacy field", in.ApiKey)
			}
			return &pb.CreateSessionResponse{Session: &pb.Session{Id: in.SessionId, Name: in.Name, Description: in.Description, Namespace: in.Namespace}}, nil
		},
	})
	client.apiKey = "transport-auth-token"

	session, err := client.CreateSession("created", "desc", "ns")
	if err != nil {
		t.Fatalf("CreateSession returned unexpected error: %v", err)
	}
	if session.Id != "session-1" || session.Name != "created" || client.sessionID != "session-1" {
		t.Fatalf("CreateSession returned %#v and client session %q, want session-1/created", session, client.sessionID)
	}
}

func TestCreateAPIKeyUsesSessionIDAndCapabilities(t *testing.T) {
	client := newConnectedSessionClient(&sessionsServiceClientStub{
		createAPIKeyFunc: func(ctx context.Context, in *pb.CreateApiKeyRequest, opts ...grpc.CallOption) (*pb.ApiKey, error) {
			if in.SessionId != "session-1" {
				t.Fatalf("CreateApiKey SessionId = %q, want session-1", in.SessionId)
			}
			if in.Name != "cli" {
				t.Fatalf("CreateApiKey Name = %q, want cli", in.Name)
			}
			if len(in.Capabilities) != 2 || in.Capabilities[0] != "read" || in.Capabilities[1] != "execute" {
				t.Fatalf("CreateApiKey Capabilities = %v, want [read execute]", in.Capabilities)
			}
			return &pb.ApiKey{Id: "key-1", Name: in.Name, SessionId: in.SessionId, Capabilities: in.Capabilities, KeyPreview: "toolplane_1234"}, nil
		},
	})

	apiKey, err := client.CreateAPIKey("cli", "read", "execute")
	if err != nil {
		t.Fatalf("CreateAPIKey returned unexpected error: %v", err)
	}
	if apiKey.Id != "key-1" || apiKey.Name != "cli" || apiKey.KeyPreview != "toolplane_1234" {
		t.Fatalf("CreateAPIKey returned %#v, want key-1/cli/toolplane_1234", apiKey)
	}
	if len(apiKey.Capabilities) != 2 || apiKey.Capabilities[0] != "read" || apiKey.Capabilities[1] != "execute" {
		t.Fatalf("CreateAPIKey capabilities = %v, want [read execute]", apiKey.Capabilities)
	}
}

func TestListAPIKeysUsesSessionID(t *testing.T) {
	client := newConnectedSessionClient(&sessionsServiceClientStub{
		listAPIKeysFunc: func(ctx context.Context, in *pb.ListApiKeysRequest, opts ...grpc.CallOption) (*pb.ListApiKeysResponse, error) {
			if in.SessionId != "session-1" {
				t.Fatalf("ListApiKeys SessionId = %q, want session-1", in.SessionId)
			}
			return &pb.ListApiKeysResponse{ApiKeys: []*pb.ApiKey{
				{Id: "key-1", Key: "", KeyPreview: "toolplane_1234", Capabilities: []string{"execute"}},
				{Id: "key-2", Key: "", KeyPreview: "toolplane_5678", Capabilities: []string{"read", "admin"}},
			}}, nil
		},
	})

	apiKeys, err := client.ListAPIKeys()
	if err != nil {
		t.Fatalf("ListAPIKeys returned unexpected error: %v", err)
	}
	if len(apiKeys) != 2 || apiKeys[0].Id != "key-1" || apiKeys[1].Id != "key-2" {
		t.Fatalf("ListAPIKeys returned %#v, want two API key IDs", apiKeys)
	}
	if apiKeys[0].Key != "" || apiKeys[0].KeyPreview != "toolplane_1234" {
		t.Fatalf("ListAPIKeys[0] returned %#v, want redacted key with preview", apiKeys[0])
	}
	if len(apiKeys[1].Capabilities) != 2 || apiKeys[1].Capabilities[0] != "read" || apiKeys[1].Capabilities[1] != "admin" {
		t.Fatalf("ListAPIKeys[1] capabilities = %v, want [read admin]", apiKeys[1].Capabilities)
	}
}

func TestRevokeAPIKeyReturnsSuccessFlag(t *testing.T) {
	client := newConnectedSessionClient(&sessionsServiceClientStub{
		revokeAPIKeyFunc: func(ctx context.Context, in *pb.RevokeApiKeyRequest, opts ...grpc.CallOption) (*pb.RevokeApiKeyResponse, error) {
			if in.SessionId != "session-1" {
				t.Fatalf("RevokeApiKey SessionId = %q, want session-1", in.SessionId)
			}
			if in.KeyId != "key-1" {
				t.Fatalf("RevokeApiKey KeyId = %q, want key-1", in.KeyId)
			}
			return &pb.RevokeApiKeyResponse{Success: true}, nil
		},
	})

	success, err := client.RevokeAPIKey("key-1")
	if err != nil {
		t.Fatalf("RevokeAPIKey returned unexpected error: %v", err)
	}
	if !success {
		t.Fatal("RevokeAPIKey returned false, want true")
	}
}
