package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	gw "toolplane/proto"
)

const serverInstructions = "Toolplane MCP gateway: durable tool execution backed by a Postgres-persisted " +
	"queue. A tools/call enqueues a durable request that a provider machine claims and executes, with " +
	"retries and a retained streaming chunk window. Clients that advertise the " +
	"io.modelcontextprotocol/tasks extension receive a task handle from tools/call and poll it with " +
	"tasks/get; other clients are served synchronously."

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

	// Create a request in the pending state so a provider machine can claim and
	// execute it. Claiming it here (as the task execution path does) would hold
	// the lease and block the polling provider until lease expiry, so tool calls
	// must enter the queue the same way provider-driven invocations do.
	request, err := s.requests.CreateRequest(ctx, &gw.CreateRequestRequest{
		SessionId: sessionID,
		ToolName:  params.Name,
		Input:     string(input),
	})
	if err != nil {
		return nil, backendError("create request for tool "+params.Name, err)
	}

	if meta.clientSupportsTasks() {
		return createTaskResult(request), nil
	}
	return s.awaitSyncResult(ctx, sessionID, request.Id)
}

// awaitSyncResult serves clients that did not advertise the Tasks extension:
// poll the request until it reaches a terminal state or the sync timeout
// elapses, then aggregate any streaming chunks into one MCP CallToolResult. If
// the caller's context already carries an earlier deadline, it naturally bounds
// the wait through the ctx.Done select.
func (s *Server) awaitSyncResult(ctx context.Context, sessionID, requestID string) (any, *Error) {
	deadline := time.Now().Add(s.syncTimeout)
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		request, err := s.requests.GetRequest(ctx, &gw.GetRequestRequest{SessionId: sessionID, RequestId: requestID})
		if err != nil {
			return nil, backendError("poll request "+requestID, err)
		}
		if isTerminalStatus(requestTaskStatus(request)) {
			return s.buildSyncCallToolResult(ctx, sessionID, request), nil
		}
		if time.Now().After(deadline) {
			return nil, errInternal(
				fmt.Sprintf("tool call did not complete within %s; the request is still running", s.syncTimeout),
				map[string]any{TaskIDDataKey: requestID},
			)
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return nil, errInternal("request cancelled while waiting for tool execution", map[string]any{TaskIDDataKey: requestID})
		}
	}
}

// buildSyncCallToolResult renders the terminal request as a CallToolResult with
// the request's retained chunks surfaced as leading text blocks. The exact chunk
// sequence is also carried in _meta so clients can assert ordering and
// completeness without parsing content blocks.
func (s *Server) buildSyncCallToolResult(ctx context.Context, sessionID string, request *gw.Request) any {
	result := callToolResultFromRequest(request)
	meta := metaMap{TaskIDDataKey: request.Id}
	if window, err := s.requests.GetRequestChunks(ctx, &gw.GetRequestChunksRequest{
		SessionId: sessionID,
		RequestId: request.Id,
	}); err == nil && len(window.Chunks) > 0 {
		content, _ := result["content"].([]any)
		chunkBlocks := make([]any, 0, len(window.Chunks))
		for _, chunk := range window.Chunks {
			chunkBlocks = append(chunkBlocks, textBlock(chunk))
		}
		result["content"] = append(chunkBlocks, content...)
		meta[ChunksMetaKey] = map[string]any{
			"requestId": request.Id,
			"startSeq":  window.StartSeq,
			"nextSeq":   window.NextSeq,
			"chunks":    window.Chunks,
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
