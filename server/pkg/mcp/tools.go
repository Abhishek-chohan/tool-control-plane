package mcp

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
		if errors.Is(err, ErrSessionNotBound) {
			return nil, errInvalidParams(
				"no session bound: set dev.toolplane/session_id in each request's _meta " +
					"(or enable gateway session auto-provisioning for development)")
		}
		return nil, internalErrorRef("session resolution", err)
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
		if errors.Is(err, ErrSessionNotBound) {
			return nil, errInvalidParams(
				"no session bound: set dev.toolplane/session_id in each request's _meta " +
					"(or enable gateway session auto-provisioning for development)")
		}
		return nil, internalErrorRef("session resolution", err)
	}

	input, err := json.Marshal(params.Arguments)
	if err != nil {
		return nil, errInvalidParams("tool arguments must be JSON-serializable: " + err.Error())
	}

	// Retry dedup: a client retrying the same JSON-RPC id for the same tool
	// with identical arguments lands on the original (possibly still
	// running) request instead of double-executing a side-effecting tool.
	// The digest keeps an intentionally different call from ever colliding.
	request, err := s.requests.CreateRequest(ctx, &gw.CreateRequestRequest{
		SessionId:      sessionID,
		ToolName:       params.Name,
		Input:          string(input),
		IdempotencyKey: deriveCallIdempotencyKey(req.ID, params.Name, input),
		TimeoutSeconds: meta.timeoutSeconds,
	})
	if err != nil {
		return nil, backendError("create request for tool "+params.Name, err)
	}
	s.trackCall(apiKey, sessionID, req.ID, request.Id)

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

// buildSyncCallToolResult renders the terminal request as one CallToolResult.
// Content appears exactly once: the final result text when present, otherwise
// the retained stream chunks joined in order (a provider that died mid-stream
// still hands back its partial output). The chunk window stays in _meta —
// machine-readable metadata for resume cursors, not duplicated content — and
// text is capped so provider output cannot flow into a model context
// unbounded.
func (s *Server) buildSyncCallToolResult(ctx context.Context, sessionID string, request *gw.Request) any {
	result := callToolResultFromRequest(request)
	meta := metaMap{TaskIDDataKey: request.Id}

	// One chunk-window read serves both needs: content replay when the
	// request has no final result text, and the resume cursor in _meta.
	window, windowErr := s.requests.GetRequestChunks(ctx, &gw.GetRequestChunksRequest{
		SessionId: sessionID,
		RequestId: request.Id,
	})
	if windowErr != nil {
		window = nil
	}

	if !hasTextContent(result["content"]) && window != nil && len(window.Chunks) > 0 {
		text := strings.Join(window.Chunks, "\n")
		result["content"] = []any{textBlock(capRenderedText(text))}
	}
	if window != nil {
		meta[ChunksMetaKey] = map[string]any{
			"requestId": request.Id,
			"startSeq":  window.StartSeq,
			"nextSeq":   window.NextSeq,
		}
	}
	result["_meta"] = meta
	return result
}

// maxRenderedTextBytes bounds one rendered CallToolResult text block: the
// gateway must not forward unbounded provider output into a model context.
const maxRenderedTextBytes = 256 * 1024

// hasTextContent reports whether the content array carries any non-empty text
// block.
func hasTextContent(content any) bool {
	blocks, ok := content.([]any)
	if !ok {
		return false
	}
	for _, block := range blocks {
		m, ok := block.(map[string]any)
		if !ok {
			continue
		}
		if m["type"] == "text" {
			if text, ok := m["text"].(string); ok && strings.TrimSpace(text) != "" {
				return true
			}
		}
	}
	return false
}

// backendError converts a gRPC backend failure into an internal JSON-RPC
// error. Backend detail (session/request IDs, store errors, auth messages) is
// logged server-side and replaced with a short correlation reference, so
// gateway clients never receive internal error text they could probe with.
func backendError(op string, err error) *Error {
	return internalErrorRef(op, err)
}

// internalErrorRef logs the backend detail and returns a client-safe internal
// error carrying a correlation reference.
func internalErrorRef(op string, err error) *Error {
	ref := correlationRef()
	log.Printf("mcp gateway: %s failed (ref %s): %v", op, ref, err)
	return errInternal(op+" failed (ref "+ref+")", nil)
}

// correlationRef generates a short random reference tying a client-visible
// error back to its server-side log line.
func correlationRef() string {
	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "00000000"
	}
	return hex.EncodeToString(buf[:])
}

// handleLegacyToolsList serves initialize-handshake clients: the same tool
// catalog without the 2026 resultType wrapper.
func (s *Server) handleLegacyToolsList(ctx context.Context, apiKey string) (any, *Error) {
	sessionID, err := s.resolveSession(ctx, requestMeta{}, apiKey)
	if err != nil {
		if errors.Is(err, ErrSessionNotBound) {
			return nil, errInvalidParams(
				"no session bound: set dev.toolplane/session_id in each request's _meta " +
					"(or enable gateway session auto-provisioning for development)")
		}
		return nil, internalErrorRef("session resolution", err)
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

	return map[string]any{"tools": tools}, nil
}

// legacyCallToolParams is the legacy tools/call params object: identical to
// the 2026 one minus the required _meta.
type legacyCallToolParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// handleLegacyToolsCall serves initialize-handshake clients: same execution
// path (create request, await terminal state), rendered as a plain
// CallToolResult without the 2026 _meta chunk decoration.
func (s *Server) handleLegacyToolsCall(ctx context.Context, req *Request, apiKey string) (any, *Error) {
	var params legacyCallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, errInvalidParams("tools/call params must be a JSON object: " + err.Error())
	}
	if strings.TrimSpace(params.Name) == "" {
		return nil, errInvalidParams("tools/call requires a non-empty tool name")
	}

	sessionID, err := s.resolveSession(ctx, requestMeta{}, apiKey)
	if err != nil {
		if errors.Is(err, ErrSessionNotBound) {
			return nil, errInvalidParams(
				"no session bound: set dev.toolplane/session_id in each request's _meta " +
					"(or enable gateway session auto-provisioning for development)")
		}
		return nil, internalErrorRef("session resolution", err)
	}

	input, err := json.Marshal(params.Arguments)
	if err != nil {
		return nil, errInvalidParams("tool arguments must be JSON-serializable: " + err.Error())
	}

	request, err := s.requests.CreateRequest(ctx, &gw.CreateRequestRequest{
		SessionId:      sessionID,
		ToolName:       params.Name,
		Input:          string(input),
		IdempotencyKey: deriveCallIdempotencyKey(req.ID, params.Name, input),
	})
	if err != nil {
		return nil, backendError("create request for tool "+params.Name, err)
	}
	s.trackCall(apiKey, sessionID, req.ID, request.Id)

	result, rpcErr := s.awaitSyncResult(ctx, sessionID, request.Id)
	if rpcErr != nil {
		return nil, rpcErr
	}

	// Strip the 2026 decorations (resultType discriminator and _meta):
	// legacy revisions carry neither on results and unknown keys only
	// invite strict-client rejections.
	if resultMap, ok := result.(map[string]any); ok {
		delete(resultMap, "_meta")
		delete(resultMap, "resultType")
		return resultMap, nil
	}
	return result, nil
}

// deriveCallIdempotencyKey builds the dedup key for a tools/call so a client
// retrying the same JSON-RPC id for the same tool with identical arguments
// lands on the original request instead of double-executing it. The tool name
// guards a reused JSON-RPC id across different tools; the input digest keeps
// an intentionally different call from ever colliding.
func deriveCallIdempotencyKey(jsonrpcID json.RawMessage, toolName string, input []byte) string {
	digest := sha256.Sum256(input)
	return fmt.Sprintf("mcp:%s:%s:%s",
		normalizeJSONRPCID(jsonrpcID),
		toolName,
		hex.EncodeToString(digest[:])[:16],
	)
}

// normalizeJSONRPCID renders a JSON-RPC id (string or number) as a stable
// string key.
func normalizeJSONRPCID(raw json.RawMessage) string {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return ""
	}
	return string(trimmed)
}

// trackCall records the JSON-RPC id -> (session, request) mapping so a later
// notifications/cancelled can fence the executing request. The key is scoped
// to the authenticated caller: JSON-RPC ids are only unique per client, and
// two sessions can concurrently use the same id. The map is bounded with FIFO
// eviction.
func (s *Server) trackCall(apiKey, sessionID string, jsonrpcID json.RawMessage, requestID string) {
	key := trackingKey(apiKey, jsonrpcID)
	if key == "" || requestID == "" {
		return
	}
	s.callsMu.Lock()
	defer s.callsMu.Unlock()
	if _, exists := s.trackedCalls[key]; !exists && len(s.trackedCalls) >= maxTrackedCalls {
		delete(s.trackedCalls, s.trackedOrder[0])
		s.trackedOrder = s.trackedOrder[1:]
	}
	if _, exists := s.trackedCalls[key]; !exists {
		s.trackedOrder = append(s.trackedOrder, key)
	}
	s.trackedCalls[key] = trackedCall{sessionID: sessionID, requestID: requestID}
}

// trackedCallFor resolves the session and request recorded for a JSON-RPC id
// from the same authenticated caller.
func (s *Server) trackedCallFor(apiKey string, jsonrpcID json.RawMessage) (string, string) {
	key := trackingKey(apiKey, jsonrpcID)
	s.callsMu.Lock()
	defer s.callsMu.Unlock()
	call, ok := s.trackedCalls[key]
	if !ok {
		return "", ""
	}
	return call.sessionID, call.requestID
}

// trackingKey scopes a JSON-RPC id to its authenticated caller using a short
// hash of the credential, the stable per-caller scope available on both the
// tools/call and the cancelled notification.
func trackingKey(apiKey string, jsonrpcID json.RawMessage) string {
	id := normalizeJSONRPCID(jsonrpcID)
	if id == "" {
		return ""
	}
	digest := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(digest[:])[:16] + ":" + id
}

// handleCancelledNotification maps notifications/cancelled to a fenced
// CancelRequest. The MCP spec's params.requestId names the JSON-RPC id of the
// call being stopped; the gateway's tracking resolves it to the durable
// request. Unknown ids (already completed, foreign, or never tracked) are
// dropped silently — the notification is still acknowledged with 202.
func (s *Server) handleCancelledNotification(ctx context.Context, req *Request, callerCredential string) {
	var params struct {
		RequestId json.RawMessage `json:"requestId"`
	}
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return
		}
	}

	sessionID, requestID := s.trackedCallFor(callerCredential, params.RequestId)
	if requestID == "" {
		return
	}
	if _, err := s.requests.CancelRequest(ctx, &gw.CancelRequestRequest{
		SessionId: sessionID,
		RequestId: requestID,
	}); err != nil {
		fmt.Printf("mcp-gateway: notifications/cancelled for %s failed: %v\n", requestID, err)
	}
}
