package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	resultMeta, _ := result["_meta"].(map[string]any)
	chunksMeta, _ := resultMeta[mcp.ChunksMetaKey].(map[string]any)
	if chunksMeta == nil {
		t.Fatalf("sync result _meta missing %s: %v", mcp.ChunksMetaKey, resultMeta)
	}
	if chunks, _ := chunksMeta["chunks"].([]any); len(chunks) != 2 {
		t.Fatalf("chunks = %v, want 2 entries", chunks)
	}
	<-providerDone
}
