package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	gw "toolplane/proto"

	"toolplane/pkg/mcp"
	"toolplane/pkg/service"
	"toolplane/pkg/storage/memory"
	"toolplane/pkg/trace"
)

// serveBackend registers the Toolplane gRPC services on a bufconn listener and
// returns a dialed client connection.
func serveBackend(
	t *testing.T,
	toolService *service.ToolService,
	sessionService *service.SessionsService,
	machineService *service.MachinesService,
	requestService *service.RequestsService,
	tasksService *service.TasksService,
) *grpc.ClientConn {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	backend := service.NewGRPCServer(toolService, sessionService, machineService, requestService, tasksService)
	gw.RegisterToolServiceServer(grpcServer, backend)
	gw.RegisterSessionsServiceServer(grpcServer, backend)
	gw.RegisterMachinesServiceServer(grpcServer, backend)
	gw.RegisterRequestsServiceServer(grpcServer, backend)
	gw.RegisterTasksServiceServer(grpcServer, backend)
	go func() {
		if err := grpcServer.Serve(listener); err != nil && ctx.Err() == nil {
			t.Errorf("bufconn serve failed: %v", err)
		}
	}()
	t.Cleanup(func() {
		cancel()
		grpcServer.Stop()
	})

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial bufconn: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

// startBackend boots an in-memory Toolplane gRPC server behind a bufconn and
// returns a dialed connection plus the raw services for test fixture setup.
func startBackend(t *testing.T) (*grpc.ClientConn, *service.SessionsService, *service.MachinesService) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	store := memory.New()
	tracer := trace.NopTracer()
	toolService := service.NewToolService(tracer, store)
	sessionService := service.NewSessionsService(tracer, store)
	machineService := service.NewMachinesService(ctx, toolService, tracer, store)
	requestService := service.NewRequestsService(ctx, toolService, machineService, tracer, store)
	tasksService := service.NewTasksService(ctx, toolService, machineService, requestService, tracer, store)

	conn := serveBackend(t, toolService, sessionService, machineService, requestService, tasksService)
	return conn, sessionService, machineService
}

func newFacade(t *testing.T, conn *grpc.ClientConn, opts ...mcp.Option) http.Handler {
	t.Helper()
	return mcp.NewServer(conn, opts...).Handler()
}

// validMeta returns a minimal modern _meta object; callers mutate it.
func validMeta() map[string]any {
	return map[string]any{
		"io.modelcontextprotocol/protocolVersion":    mcp.ProtocolVersion,
		"io.modelcontextprotocol/clientCapabilities": map[string]any{},
	}
}

func metaWithTasks() map[string]any {
	meta := validMeta()
	meta["io.modelcontextprotocol/clientCapabilities"] = map[string]any{
		"extensions": map[string]any{mcp.TasksExtensionID: map[string]any{}},
	}
	return meta
}

// postRPC sends one JSON-RPC request through the HTTP facade and returns the
// recorded status code and decoded envelope.
func postRPC(t *testing.T, handler http.Handler, method string, params map[string]any) (int, map[string]any) {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  params,
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	var envelope map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response %q: %v", recorder.Body.String(), err)
	}
	return recorder.Code, envelope
}

func rpcError(t *testing.T, envelope map[string]any) map[string]any {
	t.Helper()
	errObj, ok := envelope["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error response, got %v", envelope)
	}
	return errObj
}

func rpcResult(t *testing.T, envelope map[string]any) map[string]any {
	t.Helper()
	result, ok := envelope["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result response, got %v", envelope)
	}
	return result
}

func TestDiscoverAdvertisesTasksExtension(t *testing.T) {
	conn, _, _ := startBackend(t)
	handler := newFacade(t, conn)

	status, envelope := postRPC(t, handler, "server/discover", map[string]any{"_meta": validMeta()})
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %v)", status, envelope)
	}
	result := rpcResult(t, envelope)

	if result["resultType"] != mcp.ResultTypeComplete {
		t.Errorf("resultType = %v, want %s", result["resultType"], mcp.ResultTypeComplete)
	}
	versions, _ := result["supportedVersions"].([]any)
	if len(versions) != 1 || versions[0] != mcp.ProtocolVersion {
		t.Errorf("supportedVersions = %v, want [%s]", versions, mcp.ProtocolVersion)
	}
	capabilities, _ := result["capabilities"].(map[string]any)
	extensions, _ := capabilities["extensions"].(map[string]any)
	if _, ok := extensions[mcp.TasksExtensionID]; !ok {
		t.Errorf("capabilities.extensions missing %s: %v", mcp.TasksExtensionID, capabilities)
	}
}

func TestUnsupportedProtocolVersionRejected(t *testing.T) {
	conn, _, _ := startBackend(t)
	handler := newFacade(t, conn)

	meta := validMeta()
	meta["io.modelcontextprotocol/protocolVersion"] = "2025-03-26"
	status, envelope := postRPC(t, handler, "server/discover", map[string]any{"_meta": meta})

	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", status)
	}
	errObj := rpcError(t, envelope)
	if code, _ := errObj["code"].(float64); int(code) != mcp.CodeUnsupportedProtocolVersion {
		t.Fatalf("error code = %v, want %d", errObj["code"], mcp.CodeUnsupportedProtocolVersion)
	}
	data, _ := errObj["data"].(map[string]any)
	if data["requested"] != "2025-03-26" {
		t.Errorf("data.requested = %v, want 2025-03-26", data["requested"])
	}
}

func TestLegacyAndMissingMetaRouting(t *testing.T) {
	conn, _, _ := startBackend(t)
	handler := newFacade(t, conn)

	// A meta-less request to a 2026-only method gets a JSON-RPC
	// method-not-found (HTTP 200: the request itself was well-formed).
	status, envelope := postRPC(t, handler, "tasks/get", map[string]any{"taskId": "t1"})
	if status != http.StatusOK {
		t.Fatalf("tasks/get without meta: status = %d, want 200", status)
	}
	errObj := rpcError(t, envelope)
	if code, _ := errObj["code"].(float64); int(code) != mcp.CodeMethodNotFound {
		t.Fatalf("tasks/get error code = %v, want %d", errObj["code"], mcp.CodeMethodNotFound)
	}

	// A request declaring the 2026 version but omitting required
	// clientCapabilities is still validated by the 2026 path.
	metaLessCaps := map[string]any{
		"io.modelcontextprotocol/protocolVersion": mcp.ProtocolVersion,
	}
	status, envelope = postRPC(t, handler, "tools/list", map[string]any{"_meta": metaLessCaps})
	if status != http.StatusBadRequest {
		t.Fatalf("tools/list without capabilities: status = %d, want 400", status)
	}
	errObj = rpcError(t, envelope)
	if code, _ := errObj["code"].(float64); int(code) != mcp.CodeInvalidRequest {
		t.Fatalf("error code = %v, want %d", errObj["code"], mcp.CodeInvalidRequest)
	}

	// A meta-less tools/list is a legacy client: served, not rejected.
	status, envelope = postRPC(t, handler, "tools/list", map[string]any{})
	if status != http.StatusOK {
		t.Fatalf("legacy tools/list status = %d, want 200", status)
	}
	result := rpcResult(t, envelope)
	if _, ok := result["tools"].([]any); !ok {
		t.Fatalf("legacy tools/list missing tools array: %v", result)
	}
}

func TestUnknownMethodAndTasksUpdateRejected(t *testing.T) {
	conn, _, _ := startBackend(t)
	handler := newFacade(t, conn)

	for _, method := range []string{"bogus/method", "tasks/update"} {
		status, envelope := postRPC(t, handler, method, map[string]any{"_meta": validMeta()})
		if status != http.StatusOK {
			t.Fatalf("%s: status = %d, want 200 JSON-RPC error", method, status)
		}
		errObj := rpcError(t, envelope)
		if code, _ := errObj["code"].(float64); int(code) != mcp.CodeMethodNotFound {
			t.Errorf("%s: error code = %v, want %d", method, errObj["code"], mcp.CodeMethodNotFound)
		}
	}
}

func TestNotificationAcceptedWithoutResponse(t *testing.T) {
	conn, _, _ := startBackend(t)
	handler := newFacade(t, conn)

	body := `{"jsonrpc":"2.0","method":"notifications/cancelled","params":{"requestId":"x"}}`
	request := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202", recorder.Code)
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("expected empty body for notification, got %q", recorder.Body.String())
	}
}

func TestToolsListNormalizesSchema(t *testing.T) {
	conn, sessionService, _ := startBackend(t)
	handler := newFacade(t, conn)

	session, err := sessionService.CreateSession("test-user", "fixture", "", "", "", "")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	machines := gw.NewMachinesServiceClient(conn)
	if _, err := machines.RegisterMachine(context.Background(), &gw.RegisterMachineRequest{
		SessionId:   session.ID,
		MachineId:   "machine-fixture",
		SdkVersion:  "1.0.0",
		SdkLanguage: "go",
		Tools: []*gw.RegisterToolRequest{{
			SessionId: session.ID,
			MachineId: "machine-fixture",
			Name:      "echo",
			Schema:    `{"type":"object","properties":{"message":{"type":"string"}}}`,
		}},
	}); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	meta := validMeta()
	meta[mcp.SessionIDMetaKey] = session.ID
	status, envelope := postRPC(t, handler, "tools/list", map[string]any{"_meta": meta})
	if status != http.StatusOK {
		t.Fatalf("status = %d, body %v", status, envelope)
	}
	result := rpcResult(t, envelope)
	tools, _ := result["tools"].([]any)
	if len(tools) != 1 {
		t.Fatalf("tools = %v, want one entry", tools)
	}
	tool, _ := tools[0].(map[string]any)
	if tool["name"] != "echo" {
		t.Errorf("tool name = %v, want echo", tool["name"])
	}
	schema, _ := tool["inputSchema"].(map[string]any)
	if schema["type"] != "object" {
		t.Errorf("inputSchema.type = %v, want object", schema["type"])
	}
}

func TestToolsCallReturnsTaskHandleForTasksClients(t *testing.T) {
	conn, sessionService, _ := startBackend(t)
	handler := newFacade(t, conn)

	session, err := sessionService.CreateSession("test-user", "fixture", "", "", "", "")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	machines := gw.NewMachinesServiceClient(conn)
	if _, err := machines.RegisterMachine(context.Background(), &gw.RegisterMachineRequest{
		SessionId:   session.ID,
		MachineId:   "machine-fixture",
		SdkVersion:  "1.0.0",
		SdkLanguage: "go",
		Tools: []*gw.RegisterToolRequest{{
			SessionId: session.ID,
			MachineId: "machine-fixture",
			Name:      "echo",
			Schema:    `{"type":"object"}`,
		}},
	}); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	meta := metaWithTasks()
	meta[mcp.SessionIDMetaKey] = session.ID
	status, envelope := postRPC(t, handler, "tools/call", map[string]any{
		"_meta":     meta,
		"name":      "echo",
		"arguments": map[string]any{"message": "hello"},
	})
	if status != http.StatusOK {
		t.Fatalf("status = %d, body %v", status, envelope)
	}
	result := rpcResult(t, envelope)
	if result["resultType"] != mcp.ResultTypeTask {
		t.Fatalf("resultType = %v, want %s", result["resultType"], mcp.ResultTypeTask)
	}
	taskID, _ := result["taskId"].(string)
	if taskID == "" {
		t.Fatalf("missing taskId in %v", result)
	}
	if result["status"] != mcp.TaskStatusWorking {
		t.Errorf("status = %v, want %s", result["status"], mcp.TaskStatusWorking)
	}
	if result["ttlMs"] != nil {
		t.Errorf("ttlMs = %v, want null (unlimited)", result["ttlMs"])
	}

	// tasks/get on the handle reports the durable state.
	getStatus, getEnvelope := postRPC(t, handler, "tasks/get", map[string]any{
		"_meta":  meta,
		"taskId": taskID,
	})
	if getStatus != http.StatusOK {
		t.Fatalf("tasks/get status = %d, body %v", getStatus, getEnvelope)
	}
	task := rpcResult(t, getEnvelope)
	if task["taskId"] != taskID {
		t.Errorf("tasks/get taskId = %v, want %s", task["taskId"], taskID)
	}
	if got := task["status"]; got != mcp.TaskStatusWorking && got != mcp.TaskStatusFailed {
		t.Errorf("tasks/get status = %v, want working or failed", got)
	}

	// tasks/cancel acknowledges and the task reaches cancelled.
	cancelStatus, cancelEnvelope := postRPC(t, handler, "tasks/cancel", map[string]any{
		"_meta":  meta,
		"taskId": taskID,
	})
	if cancelStatus != http.StatusOK {
		t.Fatalf("tasks/cancel status = %d, body %v", cancelStatus, cancelEnvelope)
	}
	ack := rpcResult(t, cancelEnvelope)
	if ack["resultType"] != mcp.ResultTypeComplete {
		t.Errorf("cancel ack resultType = %v, want %s", ack["resultType"], mcp.ResultTypeComplete)
	}

	deadline := time.Now().Add(3 * time.Second)
	for {
		_, checkEnvelope := postRPC(t, handler, "tasks/get", map[string]any{"_meta": meta, "taskId": taskID})
		check := rpcResult(t, checkEnvelope)
		if check["status"] == mcp.TaskStatusCancelled {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("task never reached cancelled: %v", check)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func TestTasksGetUnknownTaskRejected(t *testing.T) {
	conn, sessionService, _ := startBackend(t)
	handler := newFacade(t, conn)

	session, err := sessionService.CreateSession("test-user", "fixture", "", "", "", "")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	meta := validMeta()
	meta[mcp.SessionIDMetaKey] = session.ID

	status, envelope := postRPC(t, handler, "tasks/get", map[string]any{"_meta": meta, "taskId": "missing-task"})
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200 JSON-RPC error", status)
	}
	errObj := rpcError(t, envelope)
	if code, _ := errObj["code"].(float64); int(code) != mcp.CodeInvalidParams {
		t.Fatalf("error code = %v, want %d", errObj["code"], mcp.CodeInvalidParams)
	}
}

func TestSyncCallTimesOutWithTaskReference(t *testing.T) {
	conn, sessionService, _ := startBackend(t)
	handler := newFacade(t, conn, mcp.WithSyncTimeout(300*time.Millisecond), mcp.WithPollInterval(50*time.Millisecond))

	session, err := sessionService.CreateSession("test-user", "fixture", "", "", "", "")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	machines := gw.NewMachinesServiceClient(conn)
	if _, err := machines.RegisterMachine(context.Background(), &gw.RegisterMachineRequest{
		SessionId:   session.ID,
		MachineId:   "machine-fixture",
		SdkVersion:  "1.0.0",
		SdkLanguage: "go",
		Tools: []*gw.RegisterToolRequest{{
			SessionId: session.ID,
			MachineId: "machine-fixture",
			Name:      "slow",
			Schema:    `{"type":"object"}`,
		}},
	}); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	// No tasks capability advertised: the facade must block, then surface the
	// still-running task in the error data.
	meta := validMeta()
	meta[mcp.SessionIDMetaKey] = session.ID
	status, envelope := postRPC(t, handler, "tools/call", map[string]any{
		"_meta": meta,
		"name":  "slow",
	})
	if status != http.StatusOK {
		t.Fatalf("status = %d, body %v", status, envelope)
	}
	errObj := rpcError(t, envelope)
	if code, _ := errObj["code"].(float64); int(code) != mcp.CodeInternalError {
		t.Fatalf("error code = %v, want %d", errObj["code"], mcp.CodeInternalError)
	}
	data, _ := errObj["data"].(map[string]any)
	if taskID, _ := data[mcp.TaskIDDataKey].(string); taskID == "" {
		t.Fatalf("error data missing %s: %v", mcp.TaskIDDataKey, errObj)
	}
}

func TestAutoProvisionsSessionPerAPIKey(t *testing.T) {
	conn, sessionService, _ := startBackend(t)
	handler := newFacade(t, conn)

	// No session in _meta: the facade provisions one per API key.
	status, envelope := postRPC(t, handler, "tools/list", map[string]any{"_meta": validMeta()})
	if status != http.StatusOK {
		t.Fatalf("status = %d, body %v", status, envelope)
	}
	sessions, err := sessionService.ListSessions(mcp.DefaultUserID)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(sessions) == 0 {
		t.Fatal("expected the facade to auto-provision a session")
	}
}
