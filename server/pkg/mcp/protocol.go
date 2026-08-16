// Package mcp implements a stateless MCP 2026-07-28 JSON-RPC facade over the
// Toolplane gRPC API. It exposes the tool registry and the durable task
// lifecycle through the MCP Tasks extension (io.modelcontextprotocol/tasks),
// translating JSON-RPC requests into Toolplane TasksService/ToolService calls.
//
// The wire shapes are pinned against the published 2026-07-28 core schema
// (schema/2026-07-28/schema.ts in modelcontextprotocol/modelcontextprotocol)
// and the Tasks extension schema (schema/draft/schema.ts in
// modelcontextprotocol/ext-tasks).
package mcp

import (
	"encoding/json"
	"fmt"
)

const (
	// JSONRPCVersion is the only JSON-RPC version MCP uses.
	JSONRPCVersion = "2.0"

	// ProtocolVersion is the MCP revision this facade speaks. There is no
	// initialize handshake in this revision: every request carries the
	// version in _meta and the server accepts or rejects each request
	// independently.
	ProtocolVersion = "2026-07-28"

	// TasksExtensionID identifies the MCP Tasks extension in capabilities.
	TasksExtensionID = "io.modelcontextprotocol/tasks"
)

// Reserved _meta keys defined by the 2026-07-28 core spec.
const (
	metaProtocolVersion    = "io.modelcontextprotocol/protocolVersion"
	metaClientCapabilities = "io.modelcontextprotocol/clientCapabilities"
)

// Toolplane-namespaced _meta keys. The prefix follows the spec's reverse-DNS
// naming rules for non-reserved vendor keys.
const (
	// SessionIDMetaKey lets a client bind a request to an existing Toolplane
	// session (the spec's explicit-handle guidance). When absent, the gateway
	// auto-provisions one session per API key.
	SessionIDMetaKey = "dev.toolplane/session_id"

	// LastSeqMetaKey carries the chunk cursor on tasks/get: the sequence
	// number of the first chunk the client still needs. Chunks with a
	// sequence >= last_seq are returned.
	LastSeqMetaKey = "dev.toolplane/last_seq"

	// ChunksMetaKey carries streaming chunks on tasks/get results while the
	// underlying request is still executing.
	ChunksMetaKey = "dev.toolplane/chunks"

	// TaskIDDataKey appears in error data when a synchronous tools/call times
	// out; the task keeps running and can be tracked with this ID.
	TaskIDDataKey = "toolplaneTaskId"
)

// JSON-RPC and MCP error codes.
const (
	CodeParseError                  = -32700
	CodeInvalidRequest              = -32600
	CodeMethodNotFound              = -32601
	CodeInvalidParams               = -32602
	CodeInternalError               = -32603
	CodeUnsupportedProtocolVersion  = -32022
	CodeMissingRequiredCapability   = -32021
)

// MCP Tasks extension task statuses.
const (
	TaskStatusWorking       = "working"
	TaskStatusInputRequired = "input_required"
	TaskStatusCompleted     = "completed"
	TaskStatusFailed        = "failed"
	TaskStatusCancelled     = "cancelled"
)

// Result types carried by the resultType discriminator field. Servers
// implementing 2026-07-28 MUST include resultType on every result.
const (
	ResultTypeComplete      = "complete"
	ResultTypeInputRequired = "input_required"
	ResultTypeTask          = "task"
)

// Request is a JSON-RPC 2.0 request. ID stays raw because JSON-RPC ids may be
// strings, numbers, or null.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// IsNotification reports whether the request carries no id and therefore
// expects no response.
func (r *Request) IsNotification() bool {
	return len(r.ID) == 0 || string(r.ID) == "null"
}

// Error is a JSON-RPC error object.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("jsonrpc error %d: %s", e.Code, e.Message)
}

// Response is a JSON-RPC 2.0 response envelope.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

func newResultResponse(id json.RawMessage, result any) Response {
	return Response{JSONRPC: JSONRPCVersion, ID: id, Result: result}
}

func newErrorResponse(id json.RawMessage, err *Error) Response {
	return Response{JSONRPC: JSONRPCVersion, ID: id, Error: err}
}

func errInvalidRequest(msg string) *Error {
	return &Error{Code: CodeInvalidRequest, Message: msg}
}

func errMethodNotFound(msg string) *Error {
	return &Error{Code: CodeMethodNotFound, Message: msg}
}

func errInvalidParams(msg string) *Error {
	return &Error{Code: CodeInvalidParams, Message: msg}
}

func errInternal(msg string, data any) *Error {
	return &Error{Code: CodeInternalError, Message: msg, Data: data}
}

// unsupportedVersionError builds the 2026-07-28 UnsupportedProtocolVersionError.
func unsupportedVersionError(requested string) *Error {
	return &Error{
		Code:    CodeUnsupportedProtocolVersion,
		Message: "Unsupported protocol version",
		Data: map[string]any{
			"supported": []string{ProtocolVersion},
			"requested": requested,
		},
	}
}

// baseParams captures the fields common to every MCP request params object.
type baseParams struct {
	Meta map[string]any `json:"_meta"`
}

// requestMeta is the decoded, validated _meta of a single request.
type requestMeta struct {
	protocolVersion string
	capabilities    map[string]any
	sessionID       string
	lastSeq         int32
	hasLastSeq      bool
}

// parseRequestMeta validates the modern per-request metadata. The 2026-07-28
// spec requires protocolVersion and clientCapabilities on every request.
func parseRequestMeta(params json.RawMessage) (requestMeta, *Error) {
	var base baseParams
	if len(params) > 0 {
		if err := json.Unmarshal(params, &base); err != nil {
			return requestMeta{}, errInvalidParams("params must be a JSON object: " + err.Error())
		}
	}
	if base.Meta == nil {
		return requestMeta{}, errInvalidRequest("missing _meta: MCP 2026-07-28 is stateless and requires _meta with " + metaProtocolVersion + " and " + metaClientCapabilities + " on every request")
	}

	meta := requestMeta{}

	version, _ := base.Meta[metaProtocolVersion].(string)
	if version == "" {
		return requestMeta{}, errInvalidRequest("missing _meta." + metaProtocolVersion + ": every request must declare its MCP protocol version")
	}
	meta.protocolVersion = version
	if version != ProtocolVersion {
		return meta, unsupportedVersionError(version)
	}

	caps, _ := base.Meta[metaClientCapabilities].(map[string]any)
	if caps == nil {
		return requestMeta{}, errInvalidRequest("missing _meta." + metaClientCapabilities + ": every request must declare client capabilities (use {} for none)")
	}
	meta.capabilities = caps

	if sessionID, ok := base.Meta[SessionIDMetaKey].(string); ok {
		meta.sessionID = sessionID
	}

	switch cursor := base.Meta[LastSeqMetaKey].(type) {
	case float64:
		meta.lastSeq = int32(cursor)
		meta.hasLastSeq = true
	case int64:
		meta.lastSeq = int32(cursor)
		meta.hasLastSeq = true
	}

	return meta, nil
}

// clientSupportsTasks reports whether the client advertised the Tasks
// extension in this request's capabilities. Servers MUST NOT return a task
// handle to a client that did not declare support.
func (m requestMeta) clientSupportsTasks() bool {
	extensions, _ := m.capabilities["extensions"].(map[string]any)
	if extensions == nil {
		return false
	}
	_, ok := extensions[TasksExtensionID]
	return ok
}
