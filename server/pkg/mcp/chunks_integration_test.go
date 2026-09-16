package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"

	"toolplane/pkg/mcp"
	"toolplane/pkg/model"
	"toolplane/pkg/service"
	"toolplane/pkg/storage/memory"
	"toolplane/pkg/trace"
)

// startBackendWithServices boots the in-memory backend like startBackend but
// also returns the raw services so tests can act as the provider runtime.
func startBackendWithServices(t *testing.T) (*grpc.ClientConn, *service.MachinesService, *service.RequestsService) {
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
	return conn, machineService, requestService
}

// rpcPost sends one JSON-RPC request through the facade handler and returns
// the decoded envelope.
func rpcPost(t *testing.T, handler http.Handler, method string, params map[string]any) map[string]any {
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
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body)))

	var envelope map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response %q: %v", recorder.Body.String(), err)
	}
	return envelope
}

func TestTasksLifecycleOverRequestModel(t *testing.T) {
	conn, machineService, requestService := startBackendWithServices(t)
	handler := mcp.NewServer(conn, mcp.WithPollInterval(20*time.Millisecond)).Handler()

	const sessionID = "session-request-model"
	const machineID = "machine-request-model"

	if _, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	meta := func(withTasks bool) map[string]any {
		capabilities := map[string]any{}
		if withTasks {
			capabilities["extensions"] = map[string]any{mcp.TasksExtensionID: map[string]any{}}
		}
		return map[string]any{
			"io.modelcontextprotocol/protocolVersion":    mcp.ProtocolVersion,
			"io.modelcontextprotocol/clientCapabilities": capabilities,
			mcp.SessionIDMetaKey:                         sessionID,
		}
	}
	tasksGet := func(taskID string, lastSeq any) map[string]any {
		params := map[string]any{"_meta": meta(false), "taskId": taskID}
		if lastSeq != nil {
			params["_meta"].(map[string]any)[mcp.LastSeqMetaKey] = lastSeq
		}
		return rpcPost(t, handler, "tasks/get", params)
	}

	// 1) tools/call with the Tasks extension advertised returns an immediate
	// handle whose ID is the queued request.
	callEnvelope := rpcPost(t, handler, "tools/call", map[string]any{
		"_meta": meta(true), "name": "echo", "arguments": map[string]any{"message": "hi"},
	})
	handle := rpcResult(t, callEnvelope)
	if handle["resultType"] != mcp.ResultTypeTask {
		t.Fatalf("resultType = %v, want %s", handle["resultType"], mcp.ResultTypeTask)
	}
	taskID, _ := handle["taskId"].(string)
	if taskID == "" {
		t.Fatalf("missing taskId in %v", handle)
	}
	if handle["status"] != mcp.TaskStatusWorking {
		t.Fatalf("handle status = %v, want %s", handle["status"], mcp.TaskStatusWorking)
	}

	// 2) Act as the provider machine: claim, run, stream chunks.
	claimed, err := requestService.ClaimRequest(sessionID, taskID, machineID)
	if err != nil {
		t.Fatalf("claim request: %v", err)
	}
	if _, err := requestService.UpdateRequest(sessionID, taskID, machineID, claimed.LeaseEpoch, model.RequestStatusRunning, nil, ""); err != nil {
		t.Fatalf("mark running: %v", err)
	}
	if err := requestService.AppendRequestChunks(sessionID, taskID, machineID, claimed.LeaseEpoch, []string{"chunk-1", "chunk-2"}, model.ResultTypeStreaming); err != nil {
		t.Fatalf("append chunks: %v", err)
	}

	// 3) tasks/get surfaces the working state with the chunk window.
	getEnvelope := tasksGet(taskID, nil)
	task := rpcResult(t, getEnvelope)
	if task["status"] != mcp.TaskStatusWorking {
		t.Fatalf("tasks/get status = %v, want %s", task["status"], mcp.TaskStatusWorking)
	}
	taskMeta, _ := task["_meta"].(map[string]any)
	if taskMeta == nil {
		t.Fatalf("tasks/get result missing _meta with chunk window: %v", task)
	}
	chunksMeta, _ := taskMeta[mcp.ChunksMetaKey].(map[string]any)
	if chunksMeta == nil {
		t.Fatalf("tasks/get _meta missing %s: %v", mcp.ChunksMetaKey, taskMeta)
	}
	chunks, _ := chunksMeta["chunks"].([]any)
	if len(chunks) != 2 || chunks[0] != "chunk-1" || chunks[1] != "chunk-2" {
		t.Fatalf("chunks = %v, want [chunk-1 chunk-2]", chunks)
	}

	// 4) Cursor replay: resuming from nextSeq must not redeliver seen chunks.
	replayEnvelope := tasksGet(taskID, chunksMeta["nextSeq"])
	replayTask := rpcResult(t, replayEnvelope)
	replayMeta, _ := replayTask["_meta"].(map[string]any)
	replayChunksMeta, _ := replayMeta[mcp.ChunksMetaKey].(map[string]any)
	if replayChunks, _ := replayChunksMeta["chunks"].([]any); len(replayChunks) != 0 {
		t.Fatalf("replay from nextSeq returned chunks %v, want none", replayChunks)
	}

	// 5) Provider submits the final result; tasks/get reports completed with
	// the CallToolResult-shaped payload.
	if err := requestService.SubmitRequestResult(sessionID, taskID, machineID, claimed.LeaseEpoch, map[string]string{"echo": "hi"}, model.ResultTypeResolution, nil); err != nil {
		t.Fatalf("submit result: %v", err)
	}
	finalEnvelope := tasksGet(taskID, nil)
	finalTask := rpcResult(t, finalEnvelope)
	if finalTask["status"] != mcp.TaskStatusCompleted {
		t.Fatalf("final status = %v, want %s", finalTask["status"], mcp.TaskStatusCompleted)
	}
	finalResult, _ := finalTask["result"].(map[string]any)
	if finalResult == nil {
		t.Fatalf("completed task missing result payload: %v", finalTask)
	}
	if structured, _ := finalResult["structuredContent"].(map[string]any); structured["echo"] != "hi" {
		t.Fatalf("structuredContent = %v, want echo=hi", finalResult["structuredContent"])
	}

	// 6) tasks/cancel on a second in-flight request reaches cancelled.
	cancelCall := rpcPost(t, handler, "tools/call", map[string]any{
		"_meta": meta(true), "name": "echo", "arguments": map[string]any{"message": "cancel me"},
	})
	cancelHandle := rpcResult(t, cancelCall)
	cancelTaskID, _ := cancelHandle["taskId"].(string)

	cancelEnvelope := rpcPost(t, handler, "tasks/cancel", map[string]any{"_meta": meta(false), "taskId": cancelTaskID})
	if ack := rpcResult(t, cancelEnvelope); ack["resultType"] != mcp.ResultTypeComplete {
		t.Fatalf("cancel ack = %v, want resultType complete", ack)
	}
	cancelledEnvelope := tasksGet(cancelTaskID, nil)
	cancelledTask := rpcResult(t, cancelledEnvelope)
	if cancelledTask["status"] != mcp.TaskStatusCancelled {
		t.Fatalf("cancelled task status = %v, want %s", cancelledTask["status"], mcp.TaskStatusCancelled)
	}
}

func TestSyncCallSurfacesChunksAndResult(t *testing.T) {
	conn, machineService, requestService := startBackendWithServices(t)
	handler := mcp.NewServer(conn, mcp.WithPollInterval(20*time.Millisecond), mcp.WithSyncTimeout(10*time.Second)).Handler()

	const sessionID = "session-sync-chunks"
	const machineID = "machine-sync-chunks"

	if _, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	// Drive the provider side concurrently: claim the request as soon as it
	// exists, stream chunks, then submit the result.
	providerDone := make(chan struct{})
	go func() {
		defer close(providerDone)
		var requestID string
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			requests, _, err := requestService.ListRequests(sessionID, model.RequestStatusPending, "", 10, 0)
			if err == nil && len(requests) > 0 {
				requestID = requests[0].ID
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if requestID == "" {
			t.Errorf("provider never saw a pending request")
			return
		}
		claimed, err := requestService.ClaimRequest(sessionID, requestID, machineID)
		if err != nil {
			t.Errorf("claim: %v", err)
			return
		}
		if _, err := requestService.UpdateRequest(sessionID, requestID, machineID, claimed.LeaseEpoch, model.RequestStatusRunning, nil, ""); err != nil {
			t.Errorf("mark running: %v", err)
			return
		}
		if err := requestService.AppendRequestChunks(sessionID, requestID, machineID, claimed.LeaseEpoch, []string{"part-1", "part-2"}, model.ResultTypeStreaming); err != nil {
			t.Errorf("append chunks: %v", err)
			return
		}
		if err := requestService.SubmitRequestResult(sessionID, requestID, machineID, claimed.LeaseEpoch, map[string]string{"echo": "sync"}, model.ResultTypeResolution, nil); err != nil {
			t.Errorf("submit result: %v", err)
		}
	}()

	// No Tasks capability advertised: the facade blocks and aggregates.
	envelope := rpcPost(t, handler, "tools/call", map[string]any{
		"_meta": map[string]any{
			"io.modelcontextprotocol/protocolVersion":    mcp.ProtocolVersion,
			"io.modelcontextprotocol/clientCapabilities": map[string]any{},
			mcp.SessionIDMetaKey:                         sessionID,
		},
		"name": "echo", "arguments": map[string]any{"message": "sync"},
	})
	result := rpcResult(t, envelope)
	if result["resultType"] != mcp.ResultTypeComplete {
		t.Fatalf("resultType = %v, want %s (envelope %v)", result["resultType"], mcp.ResultTypeComplete, envelope)
	}
	if result["isError"] == true {
		t.Fatalf("sync call reported isError: %v", result)
	}
	if structured, _ := result["structuredContent"].(map[string]any); structured["echo"] != "sync" {
		t.Fatalf("structuredContent = %v, want echo=sync", result["structuredContent"])
	}
	// The final result text is the single copy of the output: streaming
	// chunks are no longer duplicated into the content array or _meta.
	resultMeta, _ := result["_meta"].(map[string]any)
	if chunksMeta, exists := resultMeta[mcp.ChunksMetaKey]; exists {
		t.Fatalf("sync result _meta still duplicates chunks: %v", chunksMeta)
	}
	content, _ := result["content"].([]any)
	textBlocks := 0
	for _, block := range content {
		m, _ := block.(map[string]any)
		if m["type"] == "text" && strings.Contains(fmt.Sprint(m["text"]), "sync") {
			textBlocks++
		}
	}
	if textBlocks != 1 {
		t.Fatalf("result content carries %d copies of the result text, want 1: %v", textBlocks, content)
	}
	<-providerDone
}

// postRaw sends one JSON-RPC envelope with an explicit id (or none, for
// notifications) and returns the recorder for status inspection.
func postRaw(t *testing.T, handler http.Handler, id any, method string, params map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	envelope := map[string]any{"jsonrpc": "2.0", "method": method, "params": params}
	if id != nil {
		envelope["id"] = id
	}
	body, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body)))
	return recorder
}

// TestSyncCallRetryDedupsOnJSONRPCID pins retry dedup: a tools/call retry
// carrying the same JSON-RPC id, tool, and arguments reuses the original
// request instead of creating a second one (double execution).
func TestSyncCallRetryDedupsOnJSONRPCID(t *testing.T) {
	conn, machineService, requestService := startBackendWithServices(t)
	handler := mcp.NewServer(conn, mcp.WithPollInterval(20*time.Millisecond), mcp.WithSyncTimeout(200*time.Millisecond)).Handler()

	const sessionID = "session-dedup"
	if _, err := machineService.RegisterMachine(sessionID, "machine-dedup", "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, "machine-dedup", "slow", "slow tool", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	call := func() *httptest.ResponseRecorder {
		return postRaw(t, handler, 7, "tools/call", map[string]any{
			"_meta": map[string]any{
				"io.modelcontextprotocol/protocolVersion":    mcp.ProtocolVersion,
				"io.modelcontextprotocol/clientCapabilities": map[string]any{},
				mcp.SessionIDMetaKey:                         sessionID,
			},
			"name": "slow", "arguments": map[string]any{"n": 1},
		})
	}

	if rec := call(); rec.Code != http.StatusOK {
		t.Fatalf("first call status=%d body=%s", rec.Code, rec.Body.String())
	}
	if rec := call(); rec.Code != http.StatusOK {
		t.Fatalf("retry status=%d body=%s", rec.Code, rec.Body.String())
	}

	requests, _, err := requestService.ListRequests(sessionID, "", "", 100, 0)
	if err != nil {
		t.Fatalf("list requests: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("retry created %d requests, want exactly 1 (dedup failed)", len(requests))
	}
}

// TestMetaTimeoutForwardsToRequest pins _meta.dev.toolplane/timeout_seconds
// reaching the durable request as its absolute execution timeout.
func TestMetaTimeoutForwardsToRequest(t *testing.T) {
	conn, machineService, requestService := startBackendWithServices(t)
	handler := mcp.NewServer(conn, mcp.WithPollInterval(20*time.Millisecond), mcp.WithSyncTimeout(150*time.Millisecond)).Handler()

	const sessionID = "session-timeout-meta"
	if _, err := machineService.RegisterMachine(sessionID, "machine-timeout", "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, "machine-timeout", "slow", "slow tool", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	// Provider completes the call so the response carries _meta.toolplaneTaskId.
	providerDone := make(chan struct{})
	go func() {
		defer close(providerDone)
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			claimed, err := requestService.ClaimPendingRequest(sessionID, "machine-timeout", []string{"slow"})
			if err != nil || claimed == nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			_, _ = requestService.UpdateRequest(sessionID, claimed.ID, "machine-timeout", claimed.LeaseEpoch, model.RequestStatusRunning, nil, "")
			time.Sleep(50 * time.Millisecond)
			_ = requestService.SubmitRequestResult(sessionID, claimed.ID, "machine-timeout", claimed.LeaseEpoch,
				map[string]string{"ok": "true"}, model.ResultTypeResolution, nil)
			return
		}
	}()
	defer func() { <-providerDone }()

	rec := postRaw(t, handler, 3, "tools/call", map[string]any{
		"_meta": map[string]any{
			"io.modelcontextprotocol/protocolVersion":    mcp.ProtocolVersion,
			"io.modelcontextprotocol/clientCapabilities": map[string]any{},
			mcp.SessionIDMetaKey:                         sessionID,
			mcp.TimeoutMetaKey:                           120,
		},
		"name": "slow", "arguments": map[string]any{},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("call status=%d body=%s", rec.Code, rec.Body.String())
	}

	envelope := map[string]any{}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	result := envelope["result"].(map[string]any)
	meta := result["_meta"].(map[string]any)
	taskID := meta[mcp.TaskIDDataKey].(string)

	stored, err := requestService.GetRequestByID(sessionID, taskID)
	if err != nil {
		t.Fatalf("load request: %v", err)
	}
	if stored.TimeoutSeconds != 120 {
		t.Fatalf("request timeout=%d want 120 (_meta timeout not forwarded)", stored.TimeoutSeconds)
	}
}

// TestCancelledNotificationStopsRequest pins notifications/cancelled: "user
// hit stop" fences the executing request instead of being discarded.
func TestCancelledNotificationStopsRequest(t *testing.T) {
	conn, machineService, requestService := startBackendWithServices(t)
	handler := mcp.NewServer(conn, mcp.WithPollInterval(20*time.Millisecond), mcp.WithSyncTimeout(150*time.Millisecond)).Handler()

	const sessionID = "session-cancel-note"
	if _, err := machineService.RegisterMachine(sessionID, "machine-cancel", "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, "machine-cancel", "slow", "slow tool", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	rec := postRaw(t, handler, 9, "tools/call", map[string]any{
		"_meta": map[string]any{
			"io.modelcontextprotocol/protocolVersion":    mcp.ProtocolVersion,
			"io.modelcontextprotocol/clientCapabilities": map[string]any{},
			mcp.SessionIDMetaKey:                         sessionID,
		},
		"name": "slow", "arguments": map[string]any{},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("call status=%d body=%s", rec.Code, rec.Body.String())
	}

	envelope := map[string]any{}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	// The call hit the sync timeout: the request id surfaces in the error
	// data, exactly where a real client finds it before sending "stop".
	errObj := envelope["error"].(map[string]any)
	errData := errObj["data"].(map[string]any)
	requestID := errData[mcp.TaskIDDataKey].(string)

	// "User hit stop": the standard MCP cancellation notification.
	notification := postRaw(t, handler, nil, "notifications/cancelled", map[string]any{
		"requestId": 9,
		"reason":    "user pressed stop",
	})
	if notification.Code != http.StatusAccepted {
		t.Fatalf("notification status=%d want 202", notification.Code)
	}

	deadline := time.Now().Add(3 * time.Second)
	for {
		stored, err := requestService.GetRequestByID(sessionID, requestID)
		if err != nil {
			t.Fatalf("load cancelled request: %v", err)
		}
		if model.IsTerminalStatus(stored.Status) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("request still %q after notifications/cancelled", stored.Status)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// TestSyncCallFailureCarriesIsError pins the failure contract end to end: a
// provider tool that raises surfaces as a rejection → FAILED request → an
// MCP result with isError=true carrying the failure text.
func TestSyncCallFailureCarriesIsError(t *testing.T) {
	conn, machineService, requestService := startBackendWithServices(t)
	handler := mcp.NewServer(conn, mcp.WithPollInterval(20*time.Millisecond)).Handler()

	const sessionID = "session-failure-fidelity"
	if _, err := machineService.RegisterMachine(sessionID, "machine-fail", "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, "machine-fail", "explode", "raises intentionally", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	providerDone := make(chan struct{})
	go func() {
		defer close(providerDone)
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			claimed, err := requestService.ClaimPendingRequest(sessionID, "machine-fail", []string{"explode"})
			if err != nil || claimed == nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			_, _ = requestService.UpdateRequest(sessionID, claimed.ID, "machine-fail", claimed.LeaseEpoch, model.RequestStatusRunning, nil, "")
			_ = requestService.SubmitRequestResult(sessionID, claimed.ID, "machine-fail", claimed.LeaseEpoch,
				map[string]interface{}{"error": "intentional tool failure"}, model.ResultTypeRejection, nil)
			return
		}
	}()
	defer func() { <-providerDone }()

	envelope := rpcPost(t, handler, "tools/call", map[string]any{
		"_meta": map[string]any{
			"io.modelcontextprotocol/protocolVersion":    mcp.ProtocolVersion,
			"io.modelcontextprotocol/clientCapabilities": map[string]any{},
			mcp.SessionIDMetaKey:                         sessionID,
		},
		"name": "explode", "arguments": map[string]any{},
	})
	result := rpcResult(t, envelope)
	if result["isError"] != true {
		t.Fatalf("failed tool call isError=%v, want true: %v", result["isError"], result)
	}
	content := fmt.Sprint(result["content"])
	if !strings.Contains(content, "intentional tool failure") {
		t.Fatalf("content %q missing the failure message", content)
	}
}
