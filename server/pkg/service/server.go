package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"toolplane/cmd/server/auth"
	"toolplane/pkg/model"
	proto "toolplane/proto"
)

// defaultListPageSize is the page size the list handlers apply when the
// caller omits page_size; the service layers share the same default.
const defaultListPageSize = 10

// GRPCServer is the adapter between our service implementation and the gRPC interface
type GRPCServer struct {
	proto.UnimplementedToolServiceServer
	proto.UnimplementedSessionsServiceServer
	proto.UnimplementedMachinesServiceServer
	proto.UnimplementedRequestsServiceServer
	proto.UnimplementedTasksServiceServer
	toolService    *ToolService
	sessionService *SessionsService
	machineService *MachinesService
	requestService *RequestsService
	tasksService   *TasksService
}

// NewGRPCServer creates a new gRPC server adapter
func NewGRPCServer(
	toolService *ToolService,
	sessionService *SessionsService,
	machineService *MachinesService,
	requestService *RequestsService,
	tasksService *TasksService,
) *GRPCServer {
	return &GRPCServer{
		toolService:    toolService,
		sessionService: sessionService,
		machineService: machineService,
		requestService: requestService,
		tasksService:   tasksService,
	}
}

// ======================
// Tool Management Methods (Belongs to ToolService)
// ======================

// RegisterTool implements the gRPC RegisterTool method
func (s *GRPCServer) RegisterTool(ctx context.Context, req *proto.RegisterToolRequest) (*proto.RegisterToolResponse, error) {
	// Convert config map from string to interface{}
	config := make(map[string]interface{})
	for k, v := range req.Config {
		config[k] = v
	}

	// Register the tool with our service
	tool, err := s.toolService.RegisterTool(req.SessionId, req.MachineId, req.Name, req.Description, req.Schema, config, req.Tags)
	if err != nil {
		// Return the tool even if it already exists
		return &proto.RegisterToolResponse{
			Tool: convertModelToolToProto(tool),
		}, status.Errorf(codes.AlreadyExists, "tool registration: %v", err)
	}

	return &proto.RegisterToolResponse{
		Tool: convertModelToolToProto(tool),
	}, nil
}

// ListTools implements the gRPC ListTools method
func (s *GRPCServer) ListTools(ctx context.Context, req *proto.ListToolsRequest) (*proto.ListToolsResponse, error) {
	// List tools from our service
	tools, err := s.toolService.ListTools(req.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list tools: %v", err)
	}

	// Convert models to proto
	protoTools := make([]*proto.Tool, 0, len(tools))
	for _, tool := range tools {
		protoTools = append(protoTools, convertModelToolToProto(tool))
	}

	return &proto.ListToolsResponse{
		Tools: protoTools,
	}, nil
}

// GetToolById implements the gRPC GetToolById method
func (s *GRPCServer) GetToolById(ctx context.Context, req *proto.GetToolByIdRequest) (*proto.GetToolResponse, error) {
	// Get tool by ID
	tool, err := s.toolService.GetToolByID(req.SessionId, req.ToolId)
	if err != nil {
		return nil, statusFromDomainError("get tool", err)
	}

	return &proto.GetToolResponse{
		Tool: convertModelToolToProto(tool),
	}, nil
}

// GetToolByName implements the gRPC GetToolByName method
func (s *GRPCServer) GetToolByName(ctx context.Context, req *proto.GetToolByNameRequest) (*proto.GetToolResponse, error) {
	// Get tool by name
	tool, err := s.toolService.GetToolByName(req.SessionId, req.ToolName)
	if err != nil {
		return nil, statusFromDomainError("get tool", err)
	}

	return &proto.GetToolResponse{
		Tool: convertModelToolToProto(tool),
	}, nil
}

// DeleteTool implements the gRPC DeleteTool method
func (s *GRPCServer) DeleteTool(ctx context.Context, req *proto.DeleteToolRequest) (*proto.DeleteToolResponse, error) {
	// Delete tool
	err := s.toolService.DeleteTool(req.SessionId, req.ToolId)
	if err != nil {
		return nil, statusFromDomainError("delete tool", err)
	}

	return &proto.DeleteToolResponse{
		Success: true,
	}, nil
}

// UpdateToolPing implements the gRPC UpdateToolPing method
func (s *GRPCServer) UpdateToolPing(ctx context.Context, req *proto.UpdateToolPingRequest) (*proto.Tool, error) {
	// Update tool ping
	tool, err := s.toolService.UpdateToolPing(req.SessionId, req.ToolId)
	if err != nil {
		return nil, statusFromDomainError("update tool ping", err)
	}

	return convertModelToolToProto(tool), nil
}

// ======================
// Session Management Methods (Belongs to SessionsService)
// ======================

// CreateSession implements the gRPC CreateSession method
func (s *GRPCServer) CreateSession(ctx context.Context, req *proto.CreateSessionRequest) (*proto.CreateSessionResponse, error) {
	// Create session with optional client-specified ID
	session, err := s.sessionService.CreateSession(req.UserId, req.Name, req.Description, req.SessionId, req.Namespace)
	if err != nil {
		// If session already exists, return existing session
		if errors.Is(err, ErrAlreadyExists) {
			existing, getErr := s.sessionService.GetSessionByID(req.SessionId)
			if getErr != nil {
				return nil, status.Errorf(codes.Internal, "session %s exists but failed to retrieve: %v", req.SessionId, getErr)
			}
			return &proto.CreateSessionResponse{Session: convertPublicSessionToProto(existing)}, status.Errorf(codes.AlreadyExists, "session %s already exists", req.SessionId)
		}
		return nil, statusFromDomainError("create session", err)
	}

	return &proto.CreateSessionResponse{Session: convertPublicSessionToProto(session)}, nil
}

// GetSession implements the gRPC GetSession method
func (s *GRPCServer) GetSession(ctx context.Context, req *proto.GetSessionRequest) (*proto.Session, error) {
	// Get session
	session, err := s.sessionService.GetSessionByID(req.SessionId)
	if err != nil {
		return nil, statusFromDomainError("get session", err)
	}

	return convertPublicSessionToProto(session), nil
}

// ListSessions implements the gRPC ListSessions method
func (s *GRPCServer) ListSessions(ctx context.Context, req *proto.ListSessionsRequest) (*proto.ListSessionsResponse, error) {
	// List sessions
	sessions, err := s.sessionService.ListSessions(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list sessions: %v", err)
	}

	// Convert models to proto
	protoSessions := make([]*proto.Session, 0, len(sessions))
	for _, session := range sessions {
		protoSessions = append(protoSessions, convertPublicSessionToProto(session))
	}

	return &proto.ListSessionsResponse{
		Sessions: protoSessions,
	}, nil
}

// UpdateSession implements the gRPC UpdateSession method
func (s *GRPCServer) UpdateSession(ctx context.Context, req *proto.UpdateSessionRequest) (*proto.Session, error) {
	// Update session
	session, err := s.sessionService.UpdateSession(req.SessionId, req.Name, req.Description, req.Namespace)
	if err != nil {
		return nil, statusFromDomainError("update session", err)
	}

	return convertPublicSessionToProto(session), nil
}

// DeleteSession implements the gRPC DeleteSession method
func (s *GRPCServer) DeleteSession(ctx context.Context, req *proto.DeleteSessionRequest) (*proto.DeleteSessionResponse, error) {
	// Delete session
	err := s.sessionService.DeleteSession(req.SessionId)
	if err != nil {
		return nil, statusFromDomainError("delete session", err)
	}

	return &proto.DeleteSessionResponse{
		Success: true,
	}, nil
}

// ListUserSessions implements the gRPC ListUserSessions method
func (s *GRPCServer) ListUserSessions(ctx context.Context, req *proto.ListUserSessionsRequest) (*proto.ListUserSessionsResponse, error) {
	// List user sessions with pagination and filtering. The v1 cursor is an
	// opaque token; the offset it encodes never leaves this handler.
	offset, err := decodePageOffset(req.PageToken)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid page_token: %v", err)
	}

	sessions, totalCount, err := s.sessionService.ListUserSessions(
		req.UserId,
		int(req.PageSize),
		offset,
		req.Filter,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list user sessions: %v", err)
	}

	// Convert models to proto
	protoSessions := make([]*proto.Session, 0, len(sessions))
	for _, session := range sessions {
		protoSessions = append(protoSessions, convertPublicSessionToProto(session))
	}

	page := &proto.ListPage{TotalSize: int32(totalCount)}
	nextStart := offset + len(sessions)
	if len(sessions) > 0 && nextStart < totalCount {
		token, tokErr := encodePageOffset(nextStart)
		if tokErr != nil {
			return nil, status.Errorf(codes.Internal, "page token: %v", tokErr)
		}
		page.NextPageToken = token
	}

	return &proto.ListUserSessionsResponse{
		Sessions: protoSessions,
		Page:     page,
	}, nil
}

// BulkDeleteSessions implements the gRPC BulkDeleteSessions method
func (s *GRPCServer) BulkDeleteSessions(ctx context.Context, req *proto.BulkDeleteSessionsRequest) (*proto.BulkDeleteSessionsResponse, error) {
	// Bulk delete sessions
	deletedCount, failedDeletions, err := s.sessionService.BulkDeleteSessions(
		req.UserId,
		req.SessionIds,
		req.Filter,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to bulk delete sessions: %v", err)
	}

	return &proto.BulkDeleteSessionsResponse{
		DeletedCount:    int32(deletedCount),
		FailedDeletions: failedDeletions,
	}, nil
}

// GetSessionStats implements the gRPC GetSessionStats method
func (s *GRPCServer) GetSessionStats(ctx context.Context, req *proto.GetSessionStatsRequest) (*proto.GetSessionStatsResponse, error) {
	// Get session statistics
	totalSessions, activeSessions, expiredSessions, err := s.sessionService.GetSessionStats(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get session stats: %v", err)
	}

	return &proto.GetSessionStatsResponse{
		TotalSessions:   int32(totalSessions),
		ActiveSessions:  int32(activeSessions),
		ExpiredSessions: int32(expiredSessions),
	}, nil
}

// InvalidateSession implements the gRPC InvalidateSession method
func (s *GRPCServer) InvalidateSession(ctx context.Context, req *proto.InvalidateSessionRequest) (*proto.InvalidateSessionResponse, error) {
	// Revoke every live API key for the session (session-wide kill switch).
	revoked, err := s.sessionService.InvalidateSession(req.SessionId, req.Reason)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to invalidate session: %v", err)
	}

	return &proto.InvalidateSessionResponse{
		Success:        true,
		RevokedApiKeys: int32(revoked),
	}, nil
}

// ======================
// API Key Management Methods (Belongs to SessionsService)
// ======================

// CreateApiKey implements the gRPC CreateApiKey method
func (s *GRPCServer) CreateApiKey(ctx context.Context, req *proto.CreateApiKeyRequest) (*proto.ApiKey, error) {
	// Get the user ID from the session
	session, err := s.sessionService.GetSessionByID(req.SessionId)
	if err != nil {
		return nil, statusFromDomainError("get session", err)
	}

	// Create API key
	apiKey, err := s.sessionService.CreateApiKey(req.SessionId, req.Name, session.CreatedBy, req.Capabilities)
	if err != nil {
		return nil, statusFromDomainError("create API key", err)
	}

	return convertPublicAPIKeyToProto(apiKey, true), nil
}

// ListApiKeys implements the gRPC ListApiKeys method
func (s *GRPCServer) ListApiKeys(ctx context.Context, req *proto.ListApiKeysRequest) (*proto.ListApiKeysResponse, error) {
	// List API keys
	apiKeys, err := s.sessionService.ListApiKeys(req.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list API keys: %v", err)
	}

	// Convert models to proto
	protoApiKeys := make([]*proto.ApiKey, 0, len(apiKeys))
	for _, apiKey := range apiKeys {
		protoApiKeys = append(protoApiKeys, convertPublicAPIKeyToProto(apiKey, false))
	}

	return &proto.ListApiKeysResponse{
		ApiKeys: protoApiKeys,
	}, nil
}

// RevokeApiKey implements the gRPC RevokeApiKey method
func (s *GRPCServer) RevokeApiKey(ctx context.Context, req *proto.RevokeApiKeyRequest) (*proto.RevokeApiKeyResponse, error) {
	// Revoke API key
	err := s.sessionService.RevokeApiKey(req.SessionId, req.KeyId)
	if err != nil {
		return nil, statusFromDomainError("revoke API key", err)
	}

	return &proto.RevokeApiKeyResponse{
		Success: true,
	}, nil
}

// ======================
// Machine Management Methods (Belongs to MachinesService)
// ======================

// RegisterMachine implements the gRPC RegisterMachine method
func (s *GRPCServer) RegisterMachine(ctx context.Context, req *proto.RegisterMachineRequest) (*proto.Machine, error) {
	// Convert registered tools
	var tools []*model.Tool
	if len(req.Tools) > 0 {
		tools = make([]*model.Tool, 0, len(req.Tools))
		for _, toolReq := range req.Tools {
			// Convert config map from string to interface{}
			config := make(map[string]interface{})
			for k, v := range toolReq.Config {
				config[k] = v
			}

			tool := &model.Tool{
				Name:        toolReq.Name,
				Description: toolReq.Description,
				Schema:      toolReq.Schema,
				Config:      config,
				Tags:        toolReq.Tags,
			}
			tools = append(tools, tool)
		}
	}

	// Get remote IP from context
	remoteIP := "127.0.0.1" // Default to localhost if not available

	// Register machine. The re-registration credential travels in metadata
	// (validated by the machine-token interceptor before this handler runs,
	// and again here so the takeover check sees it).
	presentedToken := auth.MachineTokenFromContext(ctx)

	machine, err := s.machineService.RegisterMachine(
		req.SessionId,
		req.MachineId,
		req.SdkVersion,
		req.SdkLanguage,
		remoteIP,
		tools,
		presentedToken,
	)
	if err != nil {
		// Credential rejection (takeover attempt / mismatch) is an authz
		// outcome, not a precondition; a draining machine refuses work;
		// everything else is a server fault.
		return nil, statusFromDomainError("register machine", err)
	}

	protoMachine := convertModelMachineToProto(machine)
	// The minted credential is returned exactly once, on the registration
	// that created it.
	protoMachine.MachineToken = machine.Token
	return protoMachine, nil
}

// ListMachines implements the gRPC ListMachines method
func (s *GRPCServer) ListMachines(ctx context.Context, req *proto.ListMachinesRequest) (*proto.ListMachinesResponse, error) {
	// List machines
	machines, err := s.machineService.ListMachines(req.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list machines: %v", err)
	}

	// Convert models to proto
	protoMachines := make([]*proto.Machine, 0, len(machines))
	for _, machine := range machines {
		protoMachines = append(protoMachines, convertModelMachineToProto(machine))
	}

	return &proto.ListMachinesResponse{
		Machines: protoMachines,
	}, nil
}

// GetMachine implements the gRPC GetMachine method
func (s *GRPCServer) GetMachine(ctx context.Context, req *proto.GetMachineRequest) (*proto.Machine, error) {
	// Get machine
	machine, err := s.machineService.GetMachineByID(req.SessionId, req.MachineId)
	if err != nil {
		return nil, statusFromDomainError("get machine", err)
	}

	return convertModelMachineToProto(machine), nil
}

// UpdateMachinePing implements the gRPC UpdateMachinePing method
func (s *GRPCServer) UpdateMachinePing(ctx context.Context, req *proto.UpdateMachinePingRequest) (*proto.Machine, error) {
	// Update machine ping
	machine, err := s.machineService.UpdateMachinePing(req.SessionId, req.MachineId)
	if err != nil {
		return nil, statusFromDomainError("update machine ping", err)
	}

	return convertModelMachineToProto(machine), nil
}

// UnregisterMachine implements the gRPC UnregisterMachine method
func (s *GRPCServer) UnregisterMachine(ctx context.Context, req *proto.UnregisterMachineRequest) (*proto.UnregisterMachineResponse, error) {
	// Unregister machine
	err := s.machineService.UnregisterMachine(req.SessionId, req.MachineId)
	if err != nil {
		return nil, statusFromDomainError("unregister machine", err)
	}

	return &proto.UnregisterMachineResponse{
		Success: true,
	}, nil
}

// ======================
// Request Management Methods (Belongs to RequestsService)
// ======================

// CreateRequest implements the gRPC CreateRequest method
func (s *GRPCServer) CreateRequest(ctx context.Context, req *proto.CreateRequestRequest) (*proto.Request, error) {
	// Create request
	request, err := s.requestService.CreateRequest(req.SessionId, req.ToolName, req.Input, int(req.TimeoutSeconds), req.IdempotencyKey)
	if err != nil {
		return nil, statusFromDomainError("create request", err)
	}

	return convertModelRequestToProto(request), nil
}

// GetRequest implements the gRPC GetRequest method
func (s *GRPCServer) GetRequest(ctx context.Context, req *proto.GetRequestRequest) (*proto.Request, error) {
	// Get request
	request, err := s.requestService.GetRequestByID(req.SessionId, req.RequestId)
	if err != nil {
		return nil, statusFromDomainError("get request", err)
	}

	return convertModelRequestToProto(request), nil
}

// ListRequests implements the gRPC ListRequests method
func (s *GRPCServer) ListRequests(ctx context.Context, req *proto.ListRequestsRequest) (*proto.ListRequestsResponse, error) {
	// List requests. The v1 cursor is opaque; the offset it encodes stays
	// inside this handler.
	offset, err := decodePageOffset(req.PageToken)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid page_token: %v", err)
	}

	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = defaultListPageSize
	}

	requests, totalCount, err := s.requestService.ListRequests(
		req.SessionId,
		requestStatusFromProto(req.Status),
		req.ToolName,
		pageSize,
		offset,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list requests: %v", err)
	}

	// Convert models to proto
	protoRequests := make([]*proto.Request, 0, len(requests))
	for _, request := range requests {
		protoRequests = append(protoRequests, convertModelRequestToProto(request))
	}

	page := &proto.ListPage{TotalSize: int32(totalCount)}
	if len(protoRequests) == pageSize && len(protoRequests) > 0 {
		token, tokErr := encodePageOffset(offset + len(protoRequests))
		if tokErr != nil {
			return nil, status.Errorf(codes.Internal, "page token: %v", tokErr)
		}
		page.NextPageToken = token
	}

	return &proto.ListRequestsResponse{
		Requests: protoRequests,
	}, nil
}

// UpdateRequest implements the gRPC UpdateRequest method
func (s *GRPCServer) UpdateRequest(ctx context.Context, req *proto.UpdateRequestRequest) (*proto.Request, error) {
	// Convert result to interface{}
	var result interface{}
	if req.Result != "" {
		// Try to unmarshal as JSON, if it fails, use as string
		if err := json.Unmarshal([]byte(req.Result), &result); err != nil {
			result = req.Result
		}
	}

	// Update request
	request, err := s.requestService.UpdateRequest(
		req.SessionId,
		req.RequestId,
		req.MachineId,
		req.LeaseEpoch,
		requestStatusFromProto(req.Status),
		result,
		model.ResultType(req.ResultType),
	)
	if err != nil {
		return nil, statusFromDomainError("update request", err)
	}

	return convertModelRequestToProto(request), nil
}

// ClaimRequest implements the gRPC ClaimRequest method
func (s *GRPCServer) ClaimRequest(ctx context.Context, req *proto.ClaimRequestRequest) (*proto.Request, error) {
	// Claim request
	request, err := s.requestService.ClaimRequest(req.SessionId, req.RequestId, req.MachineId)
	if err != nil {
		return nil, statusFromDomainError("claim request", err)
	}

	return convertModelRequestToProto(request), nil
}

// CancelRequest implements the gRPC CancelRequest method
func (s *GRPCServer) CancelRequest(ctx context.Context, req *proto.CancelRequestRequest) (*proto.CancelRequestResponse, error) {
	// Cancel request
	err := s.requestService.CancelRequest(req.SessionId, req.RequestId)
	if err != nil {
		return nil, statusFromDomainError("cancel request", err)
	}

	return &proto.CancelRequestResponse{
		Success: true,
	}, nil
}

// SubmitRequestResult implements the gRPC SubmitRequestResult method
func (s *GRPCServer) SubmitRequestResult(ctx context.Context, req *proto.SubmitRequestResultRequest) (*proto.SubmitRequestResultResponse, error) {
	// Convert result to interface{}
	var result interface{}
	if req.Result != "" {
		// Try to unmarshal as JSON, if it fails, use as string
		if err := json.Unmarshal([]byte(req.Result), &result); err != nil {
			result = req.Result
		}
	}

	// Convert meta to map
	meta := make(map[string]string)
	for k, v := range req.Meta {
		meta[k] = v
	}

	// Submit request result
	err := s.requestService.SubmitRequestResult(
		req.SessionId,
		req.RequestId,
		req.MachineId,
		req.LeaseEpoch,
		result,
		model.ResultType(req.ResultType),
		meta,
	)
	if err != nil {
		return nil, statusFromDomainError("submit request result", err)
	}

	return &proto.SubmitRequestResultResponse{
		Success: true,
	}, nil
}

// AppendRequestChunks implements the gRPC AppendRequestChunks method
func (s *GRPCServer) AppendRequestChunks(ctx context.Context, req *proto.AppendRequestChunksRequest) (*proto.AppendRequestChunksResponse, error) {
	// Append request chunks
	err := s.requestService.AppendRequestChunks(
		req.SessionId,
		req.RequestId,
		req.MachineId,
		req.LeaseEpoch,
		req.Chunks,
		model.ResultType(req.ResultType),
	)
	if err != nil {
		return nil, statusFromDomainError("append request chunks", err)
	}

	return &proto.AppendRequestChunksResponse{
		Success: true,
	}, nil
}

// RenewRequestLease implements the gRPC RenewRequestLease method
func (s *GRPCServer) RenewRequestLease(ctx context.Context, req *proto.RenewRequestLeaseRequest) (*proto.Request, error) {
	renewed, err := s.requestService.RenewRequestLease(
		req.SessionId,
		req.RequestId,
		req.MachineId,
		req.LeaseEpoch,
	)
	if err != nil {
		return nil, statusFromDomainError("renew request lease", err)
	}

	return convertModelRequestToProto(renewed), nil
}

// GetRequestChunks implements the gRPC GetRequestChunks method
func (s *GRPCServer) GetRequestChunks(ctx context.Context, req *proto.GetRequestChunksRequest) (*proto.GetRequestChunksResponse, error) {
	// Get request chunks
	window, err := s.requestService.GetRequestChunks(req.SessionId, req.RequestId)
	if err != nil {
		return nil, statusFromDomainError("get request chunks", err)
	}

	return &proto.GetRequestChunksResponse{
		Chunks:   window.Chunks,
		StartSeq: window.StartSeq,
		NextSeq:  window.NextSeq,
	}, nil
}

func sendExecuteToolSnapshot(send func(*proto.ExecuteToolChunk) error, snapshot *RequestStreamSnapshot, lastSeq *int32) error {
	for index, chunk := range snapshot.Window.Chunks {
		seq := snapshot.Window.StartSeq + int32(index)
		if seq <= *lastSeq {
			continue
		}
		if err := send(&proto.ExecuteToolChunk{
			Seq:       seq,
			RequestId: snapshot.RequestID,
			Chunk:     chunk,
			IsFinal:   false,
		}); err != nil {
			return err
		}
		*lastSeq = seq
	}

	if snapshot.IsTerminal() && *lastSeq < snapshot.FinalSeq() {
		if err := send(&proto.ExecuteToolChunk{
			Seq:       snapshot.FinalSeq(),
			RequestId: snapshot.RequestID,
			Chunk:     marshalExecuteToolResult(snapshot.Result),
			IsFinal:   true,
			Error:     snapshot.Error,
		}); err != nil {
			return err
		}
		*lastSeq = snapshot.FinalSeq()
	}

	return nil
}

func marshalExecuteToolResult(result interface{}) string {
	if result == nil {
		return ""
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return fmt.Sprintf("%v", result)
	}
	return string(encoded)
}

// ======================
// Execution Methods (Belongs to ToolService)
// ======================

// ExecuteTool implements the gRPC ExecuteTool method
func (s *GRPCServer) ExecuteTool(ctx context.Context, req *proto.ExecuteToolRequest) (*proto.ExecuteToolResponse, error) {
	// Create a request for the tool execution
	request, err := s.requestService.CreateRequest(req.SessionId, req.ToolName, req.Input, int(req.TimeoutSeconds), req.IdempotencyKey)
	if err != nil {
		return nil, statusFromDomainError("execute tool", err)
	}

	// Return initial response
	return &proto.ExecuteToolResponse{
		RequestId: request.ID,
		Status:    protoRequestStatus(request.Status),
	}, nil
}

// StreamExecuteTool implements the gRPC StreamExecuteTool method
func (s *GRPCServer) StreamExecuteTool(req *proto.ExecuteToolRequest, stream proto.ToolService_StreamExecuteToolServer) error {
	if err := auth.RequireSessionCapability(stream.Context(), req.SessionId, model.APIKeyCapabilityInvoke); err != nil {
		return err
	}

	// Create a request for the tool execution
	request, err := s.requestService.CreateRequest(req.SessionId, req.ToolName, req.Input, int(req.TimeoutSeconds), req.IdempotencyKey)
	if err != nil {
		return statusFromDomainError("execute tool", err)
	}

	// Signal-driven delivery: chunk appends broadcast on the request's
	// signal, so chunks stream within microseconds of the append instead of
	// on the next poll tick. The slow fallback timer only guards against a
	// missed broadcast in an in-memory edge case.
	fallbackTick := time.NewTicker(2 * time.Second)
	defer fallbackTick.Stop()
	watch := s.requestService.subscribeRequest(request.ID)
	lastSeq := int32(0)

	for {
		snapshot, err := s.requestService.GetRequestReplayStream(request.SessionID, request.ID, lastSeq)
		if err != nil {
			return statusFromDomainError("stream request", err)
		}
		if err := sendExecuteToolSnapshot(stream.Send, snapshot, &lastSeq); err != nil {
			return status.Errorf(codes.Internal, "failed to send chunk: %v", err)
		}
		if snapshot.IsTerminal() {
			s.requestService.releaseRequestSignal(request.ID)
			return nil
		}

		select {
		case <-watch:
			watch = s.requestService.subscribeRequest(request.ID)
		case <-fallbackTick.C:
		case <-stream.Context().Done():
			return status.Errorf(codes.Canceled, "client disconnected")
		}
	}
}

// ResumeStream allows clients to resume a broken stream
func (s *GRPCServer) ResumeStream(req *proto.ResumeStreamRequest, stream proto.ToolService_ResumeStreamServer) error {
	// Fail closed before touching state: without a principal (wiring bug or a
	// test driving the handler directly), the capability check below would
	// only reject existing requests, which itself leaks existence.
	principal, ok := auth.PrincipalFromContext(stream.Context())
	if !ok || principal == nil {
		return status.Error(codes.PermissionDenied, "missing authenticated principal")
	}

	request, err := s.requestService.GetRequestByIDAnySession(req.RequestId)
	if err != nil {
		return status.Errorf(codes.NotFound, "failed to resolve request %s: not found", req.RequestId)
	}
	// Session-scoped check before the capability check, returning the same
	// NotFound as a missing request: a session-bound caller must not be able
	// to distinguish "exists in another session" from "does not exist".
	if principal.Mode != model.AuthModeFixed &&
		principal.SessionID != "" && principal.SessionID != request.SessionID {
		return status.Errorf(codes.NotFound, "failed to resolve request %s: not found", req.RequestId)
	}
	if err := auth.RequireSessionCapability(stream.Context(), request.SessionID, model.APIKeyCapabilityInvoke); err != nil {
		return err
	}

	// Signal-driven, like StreamExecuteTool: replays fast on append, with a
	// slow fallback timer as broadcast insurance.
	fallbackTick := time.NewTicker(2 * time.Second)
	defer fallbackTick.Stop()
	watch := s.requestService.subscribeRequest(req.RequestId)
	lastSeq := req.LastSeq

	for {
		snapshot, err := s.requestService.GetRequestReplayStreamAnySession(req.RequestId, lastSeq)
		if err != nil {
			return statusFromDomainError("resume stream", err)
		}
		if err := sendExecuteToolSnapshot(stream.Send, snapshot, &lastSeq); err != nil {
			return status.Errorf(codes.Internal, "failed to send resumed chunk: %v", err)
		}
		if snapshot.IsTerminal() {
			s.requestService.releaseRequestSignal(req.RequestId)
			return nil
		}

		select {
		case <-watch:
			watch = s.requestService.subscribeRequest(req.RequestId)
		case <-fallbackTick.C:
		case <-stream.Context().Done():
			return status.Errorf(codes.Canceled, "client disconnected during resume")
		}
	}
}

// ======================
// Health Check Method (Belongs to ToolService)
// ======================

// HealthCheck implements the gRPC HealthCheck method
func (s *GRPCServer) HealthCheck(ctx context.Context, req *proto.HealthCheckRequest) (*proto.HealthCheckResponse, error) {
	return &proto.HealthCheckResponse{
		Status:  "ok",
		Version: "1.0.0",
	}, nil
}

// ======================
// Task Management Methods (Belongs to TasksService)
// ======================

// CreateTask implements the gRPC CreateTask method
func (s *GRPCServer) CreateTask(ctx context.Context, req *proto.CreateTaskRequest) (*proto.Task, error) {
	// Create task
	task, err := s.tasksService.CreateTask(req.SessionId, req.ToolName, req.Input, req.IdempotencyKey)
	if err != nil {
		return nil, statusFromDomainError("create task", err)
	}

	protoTask := convertModelTaskToProto(task)
	if protoTask.CurrentRequestId == "" {
		protoTask.CurrentRequestId = s.tasksService.CurrentRequestID(task.ID)
	}
	return protoTask, nil
}

// GetTask implements the gRPC GetTask method
func (s *GRPCServer) GetTask(ctx context.Context, req *proto.GetTaskRequest) (*proto.Task, error) {
	// Get task
	task, err := s.tasksService.GetTaskByID(req.SessionId, req.TaskId)
	if err != nil {
		return nil, statusFromDomainError("get task", err)
	}

	protoTask := convertModelTaskToProto(task)
	if protoTask.CurrentRequestId == "" {
		protoTask.CurrentRequestId = s.tasksService.CurrentRequestID(task.ID)
	}
	return protoTask, nil
}

// ListTasks implements the gRPC ListTasks method
func (s *GRPCServer) ListTasks(ctx context.Context, req *proto.ListTasksRequest) (*proto.ListTasksResponse, error) {
	// List tasks
	tasks, err := s.tasksService.ListTasks(req.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list tasks: %v", err)
	}

	// Convert models to proto
	protoTasks := make([]*proto.Task, 0, len(tasks))
	for _, task := range tasks {
		protoTask := convertModelTaskToProto(task)
		if protoTask.CurrentRequestId == "" {
			protoTask.CurrentRequestId = s.tasksService.CurrentRequestID(task.ID)
		}
		protoTasks = append(protoTasks, protoTask)
	}

	return &proto.ListTasksResponse{
		Tasks: protoTasks,
	}, nil
}

// CancelTask implements the gRPC CancelTask method
func (s *GRPCServer) CancelTask(ctx context.Context, req *proto.CancelTaskRequest) (*proto.CancelTaskResponse, error) {
	// Cancel task
	err := s.tasksService.CancelTask(req.SessionId, req.TaskId)
	if err != nil {
		return nil, statusFromDomainError("cancel task", err)
	}

	return &proto.CancelTaskResponse{
		Success: true,
	}, nil
}

// ======================
// Drain Machine RPC
// ======================

// DrainMachine handles graceful machine drain and deregistration
func (s *GRPCServer) DrainMachine(ctx context.Context, req *proto.DrainMachineRequest) (*proto.DrainMachineResponse, error) {
	if err := s.machineService.DrainMachine(ctx, req.SessionId, req.MachineId); err != nil {
		return &proto.DrainMachineResponse{Drained: false}, status.Errorf(codes.Internal, "drain failed: %v", err)
	}
	return &proto.DrainMachineResponse{Drained: true}, nil
}

// encodePageOffset wraps a numeric list offset as the v1 opaque page token.
func encodePageOffset(offset int) (string, error) {
	return pageTokenCodec.Encode(offset)
}

// decodePageOffset unwraps a v1 opaque page token into its numeric list
// offset. Empty means "from the start".
func decodePageOffset(token string) (int, error) {
	return pageTokenCodec.Decode(token)
}

// InvokeTool is the v1 name for synchronous tool invocation; ExecuteTool is
// kept as a deprecated alias with identical behavior.
func (s *GRPCServer) InvokeTool(ctx context.Context, req *proto.ExecuteToolRequest) (*proto.ExecuteToolResponse, error) {
	return s.ExecuteTool(ctx, req)
}

// GetTool resolves a tool by ID or by name: the ID wins when both reference
// fields are set.
func (s *GRPCServer) GetTool(ctx context.Context, req *proto.GetToolRequest) (*proto.GetToolResponse, error) {
	if strings.TrimSpace(req.ToolId) != "" {
		return s.GetToolById(ctx, &proto.GetToolByIdRequest{SessionId: req.SessionId, ToolId: req.ToolId})
	}
	if strings.TrimSpace(req.ToolName) != "" {
		return s.GetToolByName(ctx, &proto.GetToolByNameRequest{SessionId: req.SessionId, ToolName: req.ToolName})
	}
	return nil, status.Error(codes.InvalidArgument, "failed to get tool: provide either tool_id or tool_name")
}
