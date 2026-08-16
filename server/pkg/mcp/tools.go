package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	gw "toolplane/proto"
)

const serverInstructions = "Toolplane MCP gateway: durable tool execution backed by a Postgres-persisted " +
	"queue. Tool calls are executed as Toolplane tasks with retries and a retained streaming chunk " +
	"window. Clients that advertise the io.modelcontextprotocol/tasks extension receive a task handle " +
	"from tools/call and poll it with tasks/get; other clients are served synchronously."

// handleDiscover implements server/discover, the required stateless
// advertisement of supported versions and capabilities.
func (s *Server) handleDiscover() (any, *Error) {
	return map[string]any{
		"resultType":        ResultTypeComplete,
		"ttlMs":             0,
		"supportedVersions": []string{ProtocolVersion},
		"capabilities": map[string]any{
			"tools": map[string]any{},
			"extensions": map[string]any{
				TasksExtensionID: map[string]any{},
			},
		},
		"instructions": serverInstructions,
	}, nil
}

// listToolsParams is the tools/list params object. Cursor pagination is
// accepted but Toolplane returns the full session tool list in one page.
type listToolsParams struct {
	baseParams
	Cursor string `json:"cursor"`
}

func (s *Server) handleToolsList(ctx context.Context, meta requestMeta, apiKey string) (any, *Error) {
	sessionID, err := s.resolveSession(ctx, meta, apiKey)
	if err != nil {
		return nil, errInternal("session resolution failed: "+err.Error(), nil)
	}

	response, err := s.tools.ListTools(ctx, &gw.ListToolsRequest{SessionId: sessionID})
	if err != nil {
		return nil, backendError("list tools", err)
	}

	tools := make([]any, 0, len(response.Tools))
	for _, tool := range response.Tools {
		entry := map[string]any{
			"name":        tool.Name,
			"inputSchema": normalizeToolSchema(tool.Schema),
		}
		if strings.TrimSpace(tool.Description) != "" {
			entry["description"] = tool.Description
		}
		tools = append(tools, entry)
	}

	return map[string]any{
		"resultType": ResultTypeComplete,
		"ttlMs":      0,
		"tools":      tools,
	}, nil
}

// normalizeToolSchema converts a stored Toolplane schema (a JSON string) into
// an MCP inputSchema: a root object schema. Mirrors normalizeToolSchema in
// clients/typescript-mcp-adapter/src/bridge.ts.
func normalizeToolSchema(schemaText string) map[string]any {
	fallback := map[string]any{"type": "object", "properties": map[string]any{}}
	trimmed := strings.TrimSpace(schemaText)
	if trimmed == "" {
		return fallback
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return fallback
	}
	schemaType, hasType := parsed["type"]
	if !hasType || schemaType == nil || schemaType == "object" {
		parsed["type"] = "object"
		return parsed
	}
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{"input": parsed},
		"required":   []string{"input"},
	}
}

// callToolParams is the tools/call params object.
type callToolParams struct {
	baseParams
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

func (s *Server) handleToolsCall(ctx context.Context, req *Request, meta requestMeta, apiKey string) (any, *Error) {
	var params callToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, errInvalidParams("tools/call params must be a JSON object: " + err.Error())
	}
	if strings.TrimSpace(params.Name) == "" {
		return nil, errInvalidParams("tools/call requires a non-empty tool name")
	}

	sessionID, err := s.resolveSession(ctx, meta, apiKey)
	if err != nil {
		return nil, errInternal("session resolution failed: "+err.Error(), nil)
	}

	input, err := json.Marshal(params.Arguments)
	if err != nil {
		return nil, errInvalidParams("tool arguments must be JSON-serializable: " + err.Error())
	}

	task, err := s.tasks.CreateTask(ctx, &gw.CreateTaskRequest{
		SessionId: sessionID,
		ToolName:  params.Name,
		Input:     string(input),
	})
	if err != nil {
		return nil, backendError("create task for tool "+params.Name, err)
	}

	if meta.clientSupportsTasks() {
		return createTaskResult(task), nil
	}
	return s.awaitSyncResult(ctx, sessionID, task.Id)
}

// awaitSyncResult serves clients that did not advertise the Tasks extension:
// poll the task until it reaches a terminal state or the sync timeout elapses,
// aggregating streaming chunks from the underlying request into one MCP
// CallToolResult.
func (s *Server) awaitSyncResult(ctx context.Context, sessionID, taskID string) (any, *Error) {
	deadline := time.Now().Add(s.syncTimeout)
	var lastRequestID string
	for {
		task, err := s.tasks.GetTask(ctx, &gw.GetTaskRequest{SessionId: sessionID, TaskId: taskID})
		if err != nil {
			return nil, backendError("poll task "+taskID, err)
		}
		if task.CurrentRequestId != "" {
			lastRequestID = task.CurrentRequestId
		}
		if isTerminalStatus(mapTaskStatus(task.Status)) {
			return s.buildSyncCallToolResult(ctx, sessionID, task, lastRequestID), nil
		}
		if time.Now().After(deadline) {
			return nil, errInternal(
				fmt.Sprintf("tool call did not complete within %s; the task is still running", s.syncTimeout),
				map[string]any{TaskIDDataKey: taskID},
			)
		}
		select {
		case <-time.After(s.pollInterval):
		case <-ctx.Done():
			return nil, errInternal("request cancelled while waiting for tool execution", map[string]any{TaskIDDataKey: taskID})
		}
	}
}

// buildSyncCallToolResult renders the terminal task as a CallToolResult with
// the underlying request's retained chunks surfaced as leading text blocks.
// The exact chunk sequence is also carried in _meta so clients can assert
// ordering and completeness without parsing content blocks.
func (s *Server) buildSyncCallToolResult(ctx context.Context, sessionID string, task *gw.Task, requestID string) any {
	result := callToolResultFromTask(task)
	meta := metaMap{TaskIDDataKey: task.Id}
	if requestID != "" {
		if window, err := s.requests.GetRequestChunks(ctx, &gw.GetRequestChunksRequest{
			SessionId: sessionID,
			RequestId: requestID,
		}); err == nil {
			content, _ := result["content"].([]any)
			chunkBlocks := make([]any, 0, len(window.Chunks))
			for _, chunk := range window.Chunks {
				chunkBlocks = append(chunkBlocks, textBlock(chunk))
			}
			result["content"] = append(chunkBlocks, content...)
			meta[ChunksMetaKey] = map[string]any{
				"requestId": requestID,
				"startSeq":  window.StartSeq,
				"nextSeq":   window.NextSeq,
				"chunks":    window.Chunks,
			}
		}
	}
	result["_meta"] = meta
	return result
}

// backendError converts a gRPC backend failure into an internal JSON-RPC
// error, preserving the backend message for diagnostics.
func backendError(op string, err error) *Error {
	message := err.Error()
	if trimmed := strings.TrimSpace(message); trimmed != "" {
		message = trimmed
	}
	return errInternal(op+" failed: "+message, nil)
}

var errNoSession = errors.New("no Toolplane session available")
