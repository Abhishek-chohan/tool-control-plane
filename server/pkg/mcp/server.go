package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	gw "toolplane/proto"
)

const maxRequestBodyBytes = 4 * 1024 * 1024

// DefaultUserID is the user ID applied to auto-provisioned sessions when no
// override is supplied via WithDefaultUserID.
const DefaultUserID = "mcp-gateway"

// Server is the stateless MCP facade. It holds gRPC clients for the Toolplane
// backend and per-API-key session caching for auto-provisioning; no other
// request state is kept, so any number of gateway instances can serve any
// request (the 2026-07-28 stateless-core requirement).
type Server struct {
	tools    gw.ToolServiceClient
	sessions gw.SessionsServiceClient
	tasks    gw.TasksServiceClient
	requests gw.RequestsServiceClient

	syncTimeout   time.Duration
	pollInterval  time.Duration
	defaultUserID string

	sessionsMu   sync.Mutex
	sessionByKey map[string]string
}

// Option customizes Server behavior.
type Option func(*Server)

// WithSyncTimeout bounds how long a tools/call for a non-Tasks client blocks
// before returning an error that references the still-running task.
func WithSyncTimeout(d time.Duration) Option {
	return func(s *Server) {
		if d > 0 {
			s.syncTimeout = d
		}
	}
}

// WithPollInterval sets the backend task-poll cadence used by the sync path
// and by chunk reads.
func WithPollInterval(d time.Duration) Option {
	return func(s *Server) {
		if d > 0 {
			s.pollInterval = d
		}
	}
}

// WithDefaultUserID sets the user ID used when auto-provisioning sessions for
// API keys that do not pass an explicit session via _meta.
func WithDefaultUserID(userID string) Option {
	return func(s *Server) {
		if strings.TrimSpace(userID) != "" {
			s.defaultUserID = userID
		}
	}
}

// NewServer builds a facade over the given backend connection.
func NewServer(conn grpc.ClientConnInterface, opts ...Option) *Server {
	server := &Server{
		tools:         gw.NewToolServiceClient(conn),
		sessions:      gw.NewSessionsServiceClient(conn),
		tasks:         gw.NewTasksServiceClient(conn),
		requests:      gw.NewRequestsServiceClient(conn),
		syncTimeout:   60 * time.Second,
		pollInterval:  250 * time.Millisecond,
		defaultUserID: DefaultUserID,
		sessionByKey:  make(map[string]string),
	}
	for _, opt := range opts {
		opt(server)
	}
	return server
}

// Handler exposes POST /mcp, the single Streamable HTTP endpoint.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /mcp", s.serveMCP)
	return mux
}

func (s *Server) serveMCP(w http.ResponseWriter, r *http.Request) {
	// Read one byte past the limit so oversized bodies get a clear 413
	// instead of a silent truncation surfacing as a parse error.
	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBodyBytes+1))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, newErrorResponse(nil, errInvalidRequest("failed to read request body: "+err.Error())))
		return
	}
	if len(body) > maxRequestBodyBytes {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		if encErr := json.NewEncoder(w).Encode(newErrorResponse(nil, errInvalidRequest(
			fmt.Sprintf("request body exceeds the %d byte limit", maxRequestBodyBytes)))); encErr != nil {
			fmt.Printf("mcp-gateway: failed to encode response: %v\n", encErr)
		}
		return
	}

	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, newErrorResponse(nil, &Error{Code: CodeParseError, Message: "parse error: " + err.Error()}))
		return
	}
	if req.JSONRPC != JSONRPCVersion {
		writeJSON(w, http.StatusBadRequest, newErrorResponse(req.ID, errInvalidRequest("jsonrpc must be \"2.0\"")))
		return
	}
	if req.IsNotification() {
		// Notifications (e.g. notifications/cancelled) expect no response.
		w.WriteHeader(http.StatusAccepted)
		return
	}

	ctx := metadata.NewOutgoingContext(r.Context(), authMetadata(r))
	response := s.dispatch(ctx, &req, apiKeyForCache(r))

	httpStatus := http.StatusOK
	if response.Error != nil {
		switch response.Error.Code {
		case CodeParseError, CodeInvalidRequest, CodeUnsupportedProtocolVersion, CodeMissingRequiredCapability:
			httpStatus = http.StatusBadRequest
		}
	}
	writeJSON(w, httpStatus, response)
}

// dispatch validates the per-request _meta once, then routes the method.
func (s *Server) dispatch(ctx context.Context, req *Request, apiKey string) Response {
	meta, metaErr := parseRequestMeta(req.Params)
	if metaErr != nil {
		return newErrorResponse(req.ID, metaErr)
	}

	result, rpcErr := s.route(ctx, req, meta, apiKey)
	if rpcErr != nil {
		return newErrorResponse(req.ID, rpcErr)
	}
	return newResultResponse(req.ID, result)
}

func (s *Server) route(ctx context.Context, req *Request, meta requestMeta, apiKey string) (any, *Error) {
	switch req.Method {
	case "server/discover":
		return s.handleDiscover()
	case "tools/list":
		return s.handleToolsList(ctx, meta, apiKey)
	case "tools/call":
		return s.handleToolsCall(ctx, req, meta, apiKey)
	case "tasks/get":
		return s.handleTasksGet(ctx, req, meta, apiKey)
	case "tasks/cancel":
		return s.handleTasksCancel(ctx, req, meta, apiKey)
	case "tasks/update":
		return nil, errMethodNotFound("tasks/update is not supported: Toolplane tasks never request client input")
	default:
		return nil, errMethodNotFound("unknown method: " + req.Method)
	}
}

// authMetadata mirrors cmd/proxy's header forwarding: Authorization and
// X-API-Key become the authorization / api_key metadata keys the backend's
// auth interceptor reads.
func authMetadata(r *http.Request) metadata.MD {
	md := metadata.MD{}
	if header := r.Header.Get("Authorization"); header != "" {
		md.Set("authorization", header)
	}
	if header := r.Header.Get("X-API-Key"); header != "" {
		md.Set("api_key", header)
	}
	return md
}

func apiKeyForCache(r *http.Request) string {
	if key := r.Header.Get("X-API-Key"); key != "" {
		return key
	}
	if auth := r.Header.Get("Authorization"); auth != "" {
		return auth
	}
	return "<anonymous>"
}

// resolveSession picks the Toolplane session for a request: an explicit
// dev.toolplane/session_id in _meta wins; otherwise one session per API key
// is auto-provisioned and cached, matching the TS adapter's behavior. The
// lock spans lookup and creation so concurrent requests for the same key
// cannot provision duplicate sessions.
func (s *Server) resolveSession(ctx context.Context, meta requestMeta, apiKey string) (string, error) {
	if meta.sessionID != "" {
		return meta.sessionID, nil
	}

	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	if cached, ok := s.sessionByKey[apiKey]; ok {
		return cached, nil
	}

	session, err := s.sessions.CreateSession(ctx, &gw.CreateSessionRequest{
		UserId:      s.defaultUserID,
		Name:        "mcp-gateway",
		Description: "Auto-provisioned by toolplane-mcp-gateway",
		Namespace:   "mcp",
	})
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	if session.Session == nil {
		return "", fmt.Errorf("create session: empty session in response")
	}

	s.sessionByKey[apiKey] = session.Session.Id
	return session.Session.Id, nil
}

// taskIDParams is shared by tasks/get and tasks/cancel.
type taskIDParams struct {
	baseParams
	TaskID string `json:"taskId"`
}

func (s *Server) handleTasksGet(ctx context.Context, req *Request, meta requestMeta, apiKey string) (any, *Error) {
	var params taskIDParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, errInvalidParams("tasks/get params must be a JSON object: " + err.Error())
	}
	if strings.TrimSpace(params.TaskID) == "" {
		return nil, errInvalidParams("tasks/get requires taskId")
	}

	sessionID, err := s.resolveSession(ctx, meta, apiKey)
	if err != nil {
		return nil, errInternal("session resolution failed: "+err.Error(), nil)
	}

	task, err := s.tasks.GetTask(ctx, &gw.GetTaskRequest{SessionId: sessionID, TaskId: params.TaskID})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errInvalidParams("task not found: " + params.TaskID)
		}
		return nil, backendError("get task "+params.TaskID, err)
	}

	payload := detailedTask(task)
	if !isTerminalStatus(payload.Status) && task.CurrentRequestId != "" {
		if chunksMeta := s.readChunkWindow(ctx, sessionID, task.CurrentRequestId, meta); chunksMeta != nil {
			payload.Meta = chunksMeta
		}
	}
	return payload, nil
}

// readChunkWindow fetches the underlying request's retained chunk window and
// slices it at the client's last_seq cursor. Chunks are best-effort while a
// task is working: read failures degrade to a payload without chunks.
func (s *Server) readChunkWindow(ctx context.Context, sessionID, requestID string, meta requestMeta) metaMap {
	window, err := s.requests.GetRequestChunks(ctx, &gw.GetRequestChunksRequest{
		SessionId: sessionID,
		RequestId: requestID,
	})
	if err != nil {
		return nil
	}

	var lastSeq int32
	if meta.hasLastSeq {
		lastSeq = meta.lastSeq
	}
	startSeq := window.StartSeq
	chunks := window.Chunks
	if lastSeq > startSeq {
		drop := int(lastSeq - startSeq)
		if drop >= len(chunks) {
			chunks = nil
			startSeq = window.NextSeq
		} else {
			chunks = chunks[drop:]
			startSeq = lastSeq
		}
	}

	return metaMap{
		ChunksMetaKey: map[string]any{
			"requestId": requestID,
			"startSeq":  startSeq,
			"nextSeq":   window.NextSeq,
			"chunks":    chunks,
		},
	}
}

func (s *Server) handleTasksCancel(ctx context.Context, req *Request, meta requestMeta, apiKey string) (any, *Error) {
	var params taskIDParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, errInvalidParams("tasks/cancel params must be a JSON object: " + err.Error())
	}
	if strings.TrimSpace(params.TaskID) == "" {
		return nil, errInvalidParams("tasks/cancel requires taskId")
	}

	sessionID, err := s.resolveSession(ctx, meta, apiKey)
	if err != nil {
		return nil, errInternal("session resolution failed: "+err.Error(), nil)
	}

	task, err := s.tasks.GetTask(ctx, &gw.GetTaskRequest{SessionId: sessionID, TaskId: params.TaskID})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errInvalidParams("task not found: " + params.TaskID)
		}
		return nil, backendError("get task "+params.TaskID, err)
	}

	// Cancellation is cooperative: a task already in a terminal state is
	// acknowledged rather than rejected.
	if isTerminalStatus(mapTaskStatus(task.Status)) {
		return ackResult(), nil
	}

	if _, err := s.tasks.CancelTask(ctx, &gw.CancelTaskRequest{SessionId: sessionID, TaskId: params.TaskID}); err != nil {
		return nil, backendError("cancel task "+params.TaskID, err)
	}
	return ackResult(), nil
}

func ackResult() map[string]any {
	return map[string]any{"resultType": ResultTypeComplete}
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		fmt.Printf("mcp-gateway: failed to encode response: %v\n", err)
	}
}
