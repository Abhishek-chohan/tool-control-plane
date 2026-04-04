package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
	"toolplane/pkg/trace"
)

// RequestsService handles request-related operations
type RequestsService struct {
	// In-memory storage for requests
	requests      map[string]map[string]*model.Request // map[sessionID]map[requestID]Request
	requestsMutex sync.RWMutex
	signals       sync.Map // map[requestID]*requestSignal

	// Service dependencies
	toolService    *ToolService
	machineService *MachinesService
	store          *storage.Store
	tracer         trace.SessionTracer

	leaseDuration    time.Duration
	dispatchInterval time.Duration
	retryBackoff     time.Duration
}

// RequestStreamSnapshot is a stable view of the retained chunk window for a request.
type RequestStreamSnapshot struct {
	RequestID string
	SessionID string
	Status    model.RequestStatus
	Result    interface{}
	Error     string
	Window    model.RequestChunkWindow
}

func (s RequestStreamSnapshot) IsTerminal() bool {
	return s.Status == model.RequestStatusDone || s.Status == model.RequestStatusFailed
}

func (s RequestStreamSnapshot) FinalSeq() int32 {
	return s.Window.NextSeq
}

// RequestStreamExpiredError reports that the caller asked to resume before the retained chunk window.
type RequestStreamExpiredError struct {
	RequestID string
	LastSeq   int32
	StartSeq  int32
	NextSeq   int32
}

func (e *RequestStreamExpiredError) Error() string {
	return fmt.Sprintf("request %s can only resume from seq %d; last acknowledged seq %d is outside retained window ending at seq %d", e.RequestID, e.StartSeq-1, e.LastSeq, e.NextSeq-1)
}

// NewRequestsService creates a new requests service
func NewRequestsService(toolService *ToolService, machineService *MachinesService, tracer trace.SessionTracer, store *storage.Store) *RequestsService {
	if tracer == nil {
		tracer = trace.NopTracer()
	}

	service := &RequestsService{
		requests:         make(map[string]map[string]*model.Request),
		toolService:      toolService,
		machineService:   machineService,
		store:            store,
		tracer:           tracer,
		leaseDuration:    requestLeaseDuration,
		dispatchInterval: requestDispatchInterval,
		retryBackoff:     requestBackoff,
	}
	if machineService != nil {
		machineService.SetRequestTracker(service)
	}

	if store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if requests, err := store.AllRequests(ctx); err != nil {
			log.Printf("request persistence load failed: %v", err)
		} else {
			for _, req := range requests {
				if _, ok := service.requests[req.SessionID]; !ok {
					service.requests[req.SessionID] = make(map[string]*model.Request)
				}
				service.ensureRequestDefaults(req)
				service.requests[req.SessionID][req.ID] = req
			}
		}
	}

	// Start cleanup goroutine for stalled requests
	go service.cleanupStalledRequests()

	return service
}

// CreateRequest creates a new tool execution request
func (s *RequestsService) CreateRequest(sessionID, toolName, input string) (*model.Request, error) {
	// Check if tool exists
	_, err := s.toolService.GetToolByName(sessionID, toolName)
	if err != nil {
		return nil, fmt.Errorf("tool %s not found: %v", toolName, err)
	}

	// Find machines that can handle this tool
	machines, err := s.machineService.FindMachinesWithTool(sessionID, toolName)
	if err != nil {
		return nil, fmt.Errorf("error finding machines: %v", err)
	}

	if len(machines) == 0 {
		return nil, fmt.Errorf("no machines available for tool %s", toolName)
	}

	// Create a new request
	request := model.NewRequest(sessionID, toolName, input)
	s.ensureRequestDefaults(request)
	request.TimeoutSeconds = int(requestTimeout.Seconds())
	request.BackoffSeconds = int(s.retryBackoff.Seconds())
	request.VisibleAt = time.Now()

	s.requestsMutex.Lock()
	defer s.requestsMutex.Unlock()

	// Initialize requests map for this session if not exists
	if _, ok := s.requests[sessionID]; !ok {
		s.requests[sessionID] = make(map[string]*model.Request)
	}

	// Store request
	s.requests[sessionID][request.ID] = request

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if err := s.store.SaveRequest(ctx, request); err != nil {
			log.Printf("persist request create failed: %v", err)
		}
	}

	s.recordRequestEvent(request, trace.EventRequestCreated, "", map[string]any{
		"toolName":       toolName,
		"timeoutSeconds": request.TimeoutSeconds,
		"maxAttempts":    request.MaxAttempts,
	})

	s.notifyRequestUpdate(request.ID)

	return request, nil
}

// GetRequestByID gets a request by ID
func (s *RequestsService) GetRequestByID(sessionID, requestID string) (*model.Request, error) {
	s.requestsMutex.RLock()
	sessionRequests, ok := s.requests[sessionID]
	if !ok {
		s.requestsMutex.RUnlock()
		return nil, fmt.Errorf("no requests found for session %s", sessionID)
	}
	request, ok := sessionRequests[requestID]
	s.requestsMutex.RUnlock()
	if !ok {
		return nil, fmt.Errorf("request %s not found in session %s", requestID, sessionID)
	}
	s.ensureRequestDefaults(request)

	return request, nil
}

// GetRequestByIDAnySession gets a request by ID without requiring a known session ID.
func (s *RequestsService) GetRequestByIDAnySession(requestID string) (*model.Request, error) {
	s.requestsMutex.RLock()
	defer s.requestsMutex.RUnlock()

	for _, sessionRequests := range s.requests {
		if request, ok := sessionRequests[requestID]; ok {
			s.ensureRequestDefaults(request)
			return request, nil
		}
	}

	return nil, fmt.Errorf("request %s not found", requestID)
}

// ListRequests lists all requests in a session with optional filtering
func (s *RequestsService) ListRequests(
	sessionID string,
	status model.RequestStatus,
	toolName string,
	limit int,
	offset int,
) ([]*model.Request, error) {
	s.requestsMutex.RLock()
	defer s.requestsMutex.RUnlock()

	// Check if session exists
	if _, ok := s.requests[sessionID]; !ok {
		return []*model.Request{}, nil
	}

	// Apply filters
	var filtered []*model.Request
	for _, req := range s.requests[sessionID] {
		include := true

		// Filter by status if provided
		if status != "" && req.Status != status {
			include = false
		}

		// Filter by tool name if provided
		if toolName != "" && req.ToolName != toolName {
			include = false
		}

		if include {
			filtered = append(filtered, req)
		}
	}

	// Apply pagination
	if limit <= 0 {
		limit = 10 // Default limit
	}

	if offset < 0 {
		offset = 0
	}

	// Calculate end index
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	// Check bounds
	if offset >= len(filtered) {
		return []*model.Request{}, nil
	}

	return filtered[offset:end], nil
}

// UpdateRequest updates a request's status, result, and result type
func (s *RequestsService) UpdateRequest(
	sessionID, requestID string,
	status model.RequestStatus,
	result interface{},
	resultType model.ResultType,
) (*model.Request, error) {
	s.requestsMutex.Lock()
	defer s.requestsMutex.Unlock()

	// Check if session exists
	if _, ok := s.requests[sessionID]; !ok {
		return nil, fmt.Errorf("no requests found for session %s", sessionID)
	}

	// Get request
	request, ok := s.requests[sessionID][requestID]
	if !ok {
		return nil, fmt.Errorf("request %s not found in session %s", requestID, sessionID)
	}

	prevStatus := request.Status
	reservedSlot := false
	now := time.Now()
	if status != "" && status != request.Status {
		s.ensureRequestDefaults(request)
		switch status {
		case model.RequestStatusRunning:
			machineID := request.ExecutingMachineID
			if machineID != "" {
				if !s.machineService.ReserveMachineSlot(sessionID, machineID) {
					retryAt := now.Add(s.retryBackoff)
					request.Status = model.RequestStatusPending
					request.ExecutingMachineID = ""
					request.LeasedBy = ""
					request.LeasedAt = nil
					request.VisibleAt = retryAt
					request.NextAttemptAt = &retryAt
					request.Error = "machine at capacity"
					request.LastError = request.Error
					request.UpdatedAt = now
					if s.store != nil {
						ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
						_ = s.store.SaveRequest(ctx, request)
						cancel()
					}
					s.recordRequestEvent(request, trace.EventRequestRequeued, machineID, map[string]any{
						"reason": "machine_at_capacity",
					})
					s.notifyRequestUpdate(request.ID)
					return nil, fmt.Errorf("machine %s at capacity", machineID)
				}
				reservedSlot = true
			}
			request.Status = status
			request.LeasedBy = machineID
			leaseTime := now
			request.LeasedAt = &leaseTime
			request.TimeoutSeconds = int(requestTimeout.Seconds())
			request.VisibleAt = leaseTime.Add(requestTimeout)
			request.NextAttemptAt = nil
			request.Error = ""
			request.LastError = ""
		case model.RequestStatusDone, model.RequestStatusFailed:
			request.Status = status
			if status == model.RequestStatusFailed && request.Error != "" {
				request.LastError = request.Error
			}
			request.LeasedBy = ""
			request.LeasedAt = nil
			request.VisibleAt = now
			request.NextAttemptAt = nil
		default:
			request.Status = status
		}
	}

	if result != nil {
		request.Result = result
	}

	if resultType != "" {
		request.ResultType = resultType
	}

	request.UpdatedAt = now

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if err := s.store.SaveRequest(ctx, request); err != nil {
			if reservedSlot && request.ExecutingMachineID != "" {
				s.machineService.ReleaseMachineSlot(sessionID, request.ExecutingMachineID)
			}
			return nil, fmt.Errorf("persist request update failed: %w", err)
		}
	}

	if status != "" {
		switch status {
		case model.RequestStatusRunning:
			s.recordRequestEvent(request, trace.EventRequestExecutionStarted, request.ExecutingMachineID, nil)
		case model.RequestStatusDone:
			s.recordRequestEvent(request, trace.EventRequestExecutionCompleted, request.ExecutingMachineID, map[string]any{"source": "update_request"})
		case model.RequestStatusFailed:
			s.recordRequestEvent(request, trace.EventRequestExecutionFailed, request.ExecutingMachineID, map[string]any{"source": "update_request"})
		}

		if (status == model.RequestStatusDone || status == model.RequestStatusFailed) && request.ExecutingMachineID != "" {
			s.machineService.ReleaseMachineSlot(sessionID, request.ExecutingMachineID)
		} else if status == model.RequestStatusRunning && !reservedSlot && prevStatus == model.RequestStatusRunning && request.ExecutingMachineID != "" {
			// Refresh counters for resumed execution when lock was already held
			s.machineService.ReserveMachineSlot(sessionID, request.ExecutingMachineID)
		}
	}

	s.notifyRequestUpdate(request.ID)

	return request, nil
}

// ClaimRequest marks a request as claimed by a machine
func (s *RequestsService) ClaimRequest(sessionID, requestID, machineID string) (*model.Request, error) {
	s.requestsMutex.Lock()
	defer s.requestsMutex.Unlock()
	if s.machineService.IsMachineDraining(sessionID, machineID) {
		return nil, fmt.Errorf("machine %s is draining", machineID)
	}

	// Check if session exists
	if _, ok := s.requests[sessionID]; !ok {
		return nil, fmt.Errorf("no requests found for session %s", sessionID)
	}

	// Get request
	request, ok := s.requests[sessionID][requestID]
	if !ok {
		return nil, fmt.Errorf("request %s not found in session %s", requestID, sessionID)
	}

	// Check if request is in a claimable state
	if request.Status != model.RequestStatusPending {
		return nil, fmt.Errorf("request %s is already in state %s", requestID, request.Status)
	}

	// Mark request as claimed
	request.SetClaimedBy(machineID)
	s.ensureRequestDefaults(request)
	now := time.Now()
	if request.Attempts < request.MaxAttempts {
		request.Attempts++
	}
	request.LeasedBy = machineID
	request.LeasedAt = &now
	request.VisibleAt = now.Add(s.leaseDuration)
	request.NextAttemptAt = nil
	request.LastError = ""

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if err := s.store.SaveRequest(ctx, request); err != nil {
			log.Printf("persist request claim failed: %v", err)
		}
	}

	s.recordRequestEvent(request, trace.EventRequestClaimed, machineID, map[string]any{
		"attempts": request.Attempts,
	})

	s.notifyRequestUpdate(request.ID)

	return request, nil
}

// ClaimPendingRequest finds and claims a pending request for a machine
func (s *RequestsService) ClaimPendingRequest(sessionID, machineID string, toolNames []string) (*model.Request, error) {
	if s.machineService.IsMachineDraining(sessionID, machineID) {
		return nil, fmt.Errorf("machine %s is draining", machineID)
	}

	if machineID != "" {
		load, capacity := s.machineService.MachineLoadInfo(sessionID, machineID)
		if load >= capacity {
			return nil, fmt.Errorf("machine %s at capacity", machineID)
		}
	}

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		req, err := s.store.LeasePendingRequest(ctx, sessionID, machineID, toolNames, s.leaseDuration)
		cancel()
		if err != nil {
			return nil, fmt.Errorf("persist request lease failed: %w", err)
		}
		if req == nil {
			return nil, fmt.Errorf("no pending requests found for the specified tools")
		}
		s.ensureRequestDefaults(req)
		s.requestsMutex.Lock()
		if _, ok := s.requests[req.SessionID]; !ok {
			s.requests[req.SessionID] = make(map[string]*model.Request)
		}
		s.requests[req.SessionID][req.ID] = req
		s.requestsMutex.Unlock()
		s.notifyRequestUpdate(req.ID)
		return req, nil
	}

	s.requestsMutex.Lock()
	defer s.requestsMutex.Unlock()

	sessionRequests, ok := s.requests[sessionID]
	if !ok {
		return nil, fmt.Errorf("no requests found for session %s", sessionID)
	}

	var oldestRequest *model.Request
	var oldestTime time.Time
	for _, req := range sessionRequests {
		if req.Status != model.RequestStatusPending {
			continue
		}
		for _, name := range toolNames {
			if req.ToolName != name {
				continue
			}
			if oldestRequest == nil || req.CreatedAt.Before(oldestTime) {
				oldestRequest = req
				oldestTime = req.CreatedAt
			}
			break
		}
	}

	if oldestRequest == nil {
		return nil, fmt.Errorf("no pending requests found for the specified tools")
	}

	oldestRequest.SetClaimedBy(machineID)
	s.ensureRequestDefaults(oldestRequest)
	now := time.Now()
	if oldestRequest.Attempts < oldestRequest.MaxAttempts {
		oldestRequest.Attempts++
	}
	oldestRequest.LeasedBy = machineID
	oldestRequest.LeasedAt = &now
	oldestRequest.VisibleAt = now.Add(s.leaseDuration)
	oldestRequest.NextAttemptAt = nil
	oldestRequest.LastError = ""

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		if err := s.store.SaveRequest(ctx, oldestRequest); err != nil {
			log.Printf("persist request auto-claim failed: %v", err)
		}
		cancel()
	}

	s.recordRequestEvent(oldestRequest, trace.EventRequestClaimed, machineID, map[string]any{
		"attempts": oldestRequest.Attempts,
		"source":   "claim_pending",
	})

	s.notifyRequestUpdate(oldestRequest.ID)
	return oldestRequest, nil
}

// SubmitRequestResult submits a result for a request
func (s *RequestsService) SubmitRequestResult(
	sessionID, requestID string,
	result interface{},
	resultType model.ResultType,
	meta map[string]string,
) error {
	s.requestsMutex.Lock()
	defer s.requestsMutex.Unlock()

	// Check if session exists
	if _, ok := s.requests[sessionID]; !ok {
		return fmt.Errorf("no requests found for session %s", sessionID)
	}

	// Get request
	request, ok := s.requests[sessionID][requestID]
	if !ok {
		return fmt.Errorf("request %s not found in session %s", requestID, sessionID)
	}

	// Special handling for streaming updates to allow continued updates
	if resultType == model.ResultTypeStreaming {
		// For streaming updates, handle differently
		if request.Status == model.RequestStatusDone || request.Status == model.RequestStatusFailed {
			// For completed requests, add to stream results
			if resultStr, ok := result.(string); ok {
				request.AddStreamChunk(resultStr)
				if s.store != nil {
					ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
					defer cancel()
					if err := s.store.SaveRequest(ctx, request); err != nil {
						log.Printf("persist request streaming append failed: %v", err)
					}
				}
				s.notifyRequestUpdate(request.ID)
			}
			return nil
		}
	}

	// Don't update already completed requests unless streaming
	if (request.Status == model.RequestStatusDone || request.Status == model.RequestStatusFailed) &&
		resultType != model.ResultTypeStreaming {
		return fmt.Errorf("request %s is already in state %s", requestID, request.Status)
	}

	// Update metadata if provided
	if meta != nil {
		if request.Meta == nil {
			request.Meta = make(map[string]string)
		}
		for k, v := range meta {
			request.Meta[k] = v
		}
	}

	// Update request with result
	errorMsg := ""
	if resultType == model.ResultTypeRejection {
		// Extract error message if available
		if errObj, ok := result.(map[string]interface{}); ok {
			if msg, ok := errObj["error"]; ok {
				if errMsg, ok := msg.(string); ok {
					errorMsg = errMsg
				}
			}
		}
	}

	request.SetResult(result, resultType, errorMsg)
	if resultType == model.ResultTypeRejection {
		request.LastError = errorMsg
	} else {
		request.LastError = ""
	}
	request.DeadLetter = false

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if err := s.store.SaveRequest(ctx, request); err != nil {
			log.Printf("persist request result failed: %v", err)
		}
	}

	resultEvent := trace.EventRequestExecutionCompleted
	if resultType == model.ResultTypeRejection {
		resultEvent = trace.EventRequestExecutionFailed
	}
	s.recordRequestEvent(request, resultEvent, request.ExecutingMachineID, map[string]any{
		"resultType": resultType,
	})

	if request.ExecutingMachineID != "" {
		s.machineService.ReleaseMachineSlot(sessionID, request.ExecutingMachineID)
	}

	s.notifyRequestUpdate(request.ID)

	return nil
}

// CancelRequest cancels a request
func (s *RequestsService) CancelRequest(sessionID, requestID string) error {
	s.requestsMutex.Lock()
	defer s.requestsMutex.Unlock()

	// Check if session exists
	if _, ok := s.requests[sessionID]; !ok {
		return fmt.Errorf("no requests found for session %s", sessionID)
	}

	// Get request
	request, ok := s.requests[sessionID][requestID]
	if !ok {
		return fmt.Errorf("request %s not found in session %s", requestID, sessionID)
	}

	// Check if request can be cancelled
	if request.Status == model.RequestStatusDone || request.Status == model.RequestStatusFailed {
		return fmt.Errorf("request %s is already in state %s", requestID, request.Status)
	}

	// Cancel request
	request.SetResult(map[string]string{"message": "Request was cancelled"}, model.ResultTypeRejection, "Request was cancelled")
	request.LastError = "Request was cancelled"
	request.DeadLetter = true

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if err := s.store.SaveRequest(ctx, request); err != nil {
			log.Printf("persist request cancel failed: %v", err)
		}
	}

	s.recordRequestEvent(request, trace.EventRequestCancelled, request.ExecutingMachineID, nil)

	if request.ExecutingMachineID != "" {
		s.machineService.ReleaseMachineSlot(sessionID, request.ExecutingMachineID)
	}

	s.notifyRequestUpdate(request.ID)

	return nil
}

// AppendRequestChunks appends chunks to a request's streaming results
func (s *RequestsService) AppendRequestChunks(
	sessionID, requestID string,
	chunks []string,
	resultType model.ResultType,
) error {
	s.requestsMutex.Lock()
	defer s.requestsMutex.Unlock()

	// Check if session exists
	if _, ok := s.requests[sessionID]; !ok {
		return fmt.Errorf("no requests found for session %s", sessionID)
	}

	// Get request
	request, ok := s.requests[sessionID][requestID]
	if !ok {
		return fmt.Errorf("request %s not found in session %s", requestID, sessionID)
	}

	// Add chunks to request
	for _, chunk := range chunks {
		request.AddStreamChunk(chunk)
	}

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if err := s.store.SaveRequest(ctx, request); err != nil {
			log.Printf("persist request append chunks failed: %v", err)
		}
	}

	s.recordRequestEvent(request, trace.EventRequestChunksAppended, request.ExecutingMachineID, map[string]any{
		"chunkCount": len(chunks),
		"nextSeq":    request.NextStreamSeq,
	})

	s.notifyRequestUpdate(request.ID)

	return nil
}

// GetRequestChunks gets the current retained chunk window for a request.
func (s *RequestsService) GetRequestChunks(sessionID, requestID string) (model.RequestChunkWindow, error) {
	snapshot, err := s.GetRequestStream(sessionID, requestID)
	if err != nil {
		return model.RequestChunkWindow{}, err
	}
	return snapshot.Window, nil
}

// GetRequestStream returns a stable snapshot of the current retained stream window.
func (s *RequestsService) GetRequestStream(sessionID, requestID string) (*RequestStreamSnapshot, error) {
	s.requestsMutex.RLock()
	defer s.requestsMutex.RUnlock()

	request, err := s.getRequestLocked(sessionID, requestID)
	if err != nil {
		return nil, err
	}

	return snapshotRequestStreamLocked(request), nil
}

// GetRequestStreamAnySession returns a retained stream snapshot without requiring a session ID.
func (s *RequestsService) GetRequestStreamAnySession(requestID string) (*RequestStreamSnapshot, error) {
	s.requestsMutex.RLock()
	defer s.requestsMutex.RUnlock()

	request, err := s.getRequestAnySessionLocked(requestID)
	if err != nil {
		return nil, err
	}

	return snapshotRequestStreamLocked(request), nil
}

// GetRequestReplayStream returns only the retained chunks that follow lastSeq.
func (s *RequestsService) GetRequestReplayStream(sessionID, requestID string, lastSeq int32) (*RequestStreamSnapshot, error) {
	s.requestsMutex.RLock()
	defer s.requestsMutex.RUnlock()

	request, err := s.getRequestLocked(sessionID, requestID)
	if err != nil {
		return nil, err
	}

	return snapshotRequestReplayLocked(request, lastSeq)
}

// GetRequestReplayStreamAnySession returns only the retained chunks that follow lastSeq.
func (s *RequestsService) GetRequestReplayStreamAnySession(requestID string, lastSeq int32) (*RequestStreamSnapshot, error) {
	s.requestsMutex.RLock()
	defer s.requestsMutex.RUnlock()

	request, err := s.getRequestAnySessionLocked(requestID)
	if err != nil {
		return nil, err
	}

	return snapshotRequestReplayLocked(request, lastSeq)
}

// ActiveRequestsForMachine returns the number of claimed or running requests still assigned to a machine.
func (s *RequestsService) ActiveRequestsForMachine(sessionID, machineID string) int {
	s.requestsMutex.RLock()
	defer s.requestsMutex.RUnlock()

	requests, ok := s.requests[sessionID]
	if !ok {
		return 0
	}

	active := 0
	for _, request := range requests {
		if request == nil || request.ExecutingMachineID != machineID {
			continue
		}
		switch request.Status {
		case model.RequestStatusClaimed, model.RequestStatusRunning:
			active++
		}
	}

	return active
}

// cleanupStalledRequests periodically cleans up stalled requests
func (s *RequestsService) cleanupStalledRequests() {
	ticker := time.NewTicker(s.dispatchInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.markStalledRequests()
	}
}

// markStalledRequests marks stalled requests
func (s *RequestsService) markStalledRequests() {
	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		expired, err := s.store.FindExpiredRequests(ctx, 256)
		cancel()
		if err != nil {
			log.Printf("persist request expired scan failed: %v", err)
			return
		}
		if len(expired) == 0 {
			return
		}
		for _, req := range expired {
			s.handleExpiredRequest(req)
		}
		return
	}

	now := time.Now()
	expired := make([]*model.Request, 0)
	s.requestsMutex.RLock()
	for _, requests := range s.requests {
		for _, request := range requests {
			s.ensureRequestDefaults(request)
			if (request.Status == model.RequestStatusClaimed || request.Status == model.RequestStatusRunning) &&
				!request.VisibleAt.IsZero() && !request.VisibleAt.After(now) {
				expired = append(expired, request)
			}
		}
	}
	s.requestsMutex.RUnlock()

	for _, request := range expired {
		s.handleExpiredRequest(request)
	}
}

// ExecuteTool executes a tool on a specific machine
func (s *RequestsService) ExecuteTool(sessionID, machineID, toolName, input string) (*model.ToolResult, error) {
	// Create a new request
	request, err := s.CreateRequest(sessionID, toolName, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	return s.ExecuteRequest(context.Background(), sessionID, machineID, request.ID)
}

// ExecuteRequest claims and runs an existing request until it reaches a terminal state or the context is cancelled.
func (s *RequestsService) ExecuteRequest(ctx context.Context, sessionID, machineID, requestID string) (*model.ToolResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		timeoutCtx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()
		ctx = timeoutCtx
	}

	if err := ctx.Err(); err != nil {
		_ = s.CancelRequest(sessionID, requestID)
		return nil, err
	}

	// Claim the request for the specified machine
	_, err := s.ClaimRequest(sessionID, requestID, machineID)
	if err != nil {
		return nil, fmt.Errorf("failed to claim request: %v", err)
	}

	// Update request status to running
	_, err = s.UpdateRequest(sessionID, requestID, model.RequestStatusRunning, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to update request status: %v", err)
	}

	watch := s.subscribeRequest(requestID)

	for {
		req, err := s.GetRequestByID(sessionID, requestID)
		if err != nil {
			return nil, fmt.Errorf("failed to get request: %v", err)
		}

		switch req.Status {
		case model.RequestStatusDone:
			return &model.ToolResult{
				RequestID:  req.ID,
				Result:     fmt.Sprintf("%v", req.Result),
				ResultType: string(req.ResultType),
			}, nil
		case model.RequestStatusFailed:
			if req.Error != "" {
				return nil, fmt.Errorf("tool execution failed: %s", req.Error)
			}
			return nil, fmt.Errorf("tool execution failed")
		case model.RequestStatusStalled:
			return nil, fmt.Errorf("tool execution stalled")
		}

		select {
		case <-watch:
			watch = s.subscribeRequest(requestID)
		case <-ctx.Done():
			_ = s.CancelRequest(sessionID, requestID)
			return nil, ctx.Err()
		}
	}
}

type requestSignal struct {
	mu sync.Mutex
	ch chan struct{}
}

func newRequestSignal() *requestSignal {
	return &requestSignal{ch: make(chan struct{})}
}

func (s *requestSignal) subscribe() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ch
}

func (s *requestSignal) broadcast() {
	s.mu.Lock()
	close(s.ch)
	s.ch = make(chan struct{})
	s.mu.Unlock()
}

func (s *RequestsService) ensureRequestDefaults(req *model.Request) {
	if req == nil {
		return
	}
	if req.Meta == nil {
		req.Meta = make(map[string]string)
	}
	if req.StreamResults == nil {
		req.StreamResults = []string{}
	}
	req.EnsureStreamSequenceDefaults()
	if req.TimeoutSeconds <= 0 {
		req.TimeoutSeconds = int(requestTimeout.Seconds())
	}
	if req.BackoffSeconds <= 0 {
		req.BackoffSeconds = int(s.retryBackoff.Seconds())
	}
	if req.MaxAttempts <= 0 {
		req.MaxAttempts = 3
	}
	if req.VisibleAt.IsZero() {
		req.VisibleAt = time.Now()
	}
}

func (s *RequestsService) recordRequestEvent(
	req *model.Request,
	event trace.SessionEventType,
	machineID string,
	metadata map[string]any,
) {
	if req == nil || s.tracer == nil {
		return
	}

	eventMachineID := machineID
	if eventMachineID == "" {
		eventMachineID = req.ExecutingMachineID
	}

	timestamp := time.Now()
	if !req.UpdatedAt.IsZero() {
		timestamp = req.UpdatedAt
	}

	eventMetadata := map[string]any{
		"requestID": req.ID,
		"toolName":  req.ToolName,
		"status":    req.Status,
		"attempts":  req.Attempts,
	}
	for key, value := range metadata {
		eventMetadata[key] = value
	}

	s.tracer.Record(trace.SessionEvent{
		SessionID: req.SessionID,
		MachineID: eventMachineID,
		Event:     event,
		Timestamp: timestamp,
		Metadata:  eventMetadata,
	})
}

func (s *RequestsService) getRequestLocked(sessionID, requestID string) (*model.Request, error) {
	sessionRequests, ok := s.requests[sessionID]
	if !ok {
		return nil, fmt.Errorf("no requests found for session %s", sessionID)
	}

	request, ok := sessionRequests[requestID]
	if !ok {
		return nil, fmt.Errorf("request %s not found in session %s", requestID, sessionID)
	}

	s.ensureRequestDefaults(request)
	return request, nil
}

func (s *RequestsService) getRequestAnySessionLocked(requestID string) (*model.Request, error) {
	for _, sessionRequests := range s.requests {
		if request, ok := sessionRequests[requestID]; ok {
			s.ensureRequestDefaults(request)
			return request, nil
		}
	}

	return nil, fmt.Errorf("request %s not found", requestID)
}

func snapshotRequestStreamLocked(request *model.Request) *RequestStreamSnapshot {
	request.EnsureStreamSequenceDefaults()
	return &RequestStreamSnapshot{
		RequestID: request.ID,
		SessionID: request.SessionID,
		Status:    request.Status,
		Result:    request.Result,
		Error:     request.Error,
		Window:    request.StreamChunkWindow(),
	}
}

func snapshotRequestReplayLocked(request *model.Request, lastSeq int32) (*RequestStreamSnapshot, error) {
	snapshot := snapshotRequestStreamLocked(request)
	if lastSeq < 0 {
		lastSeq = 0
	}
	requestedSeq := lastSeq + 1
	if requestedSeq < snapshot.Window.StartSeq && requestedSeq < snapshot.Window.NextSeq {
		return nil, &RequestStreamExpiredError{
			RequestID: snapshot.RequestID,
			LastSeq:   lastSeq,
			StartSeq:  snapshot.Window.StartSeq,
			NextSeq:   snapshot.Window.NextSeq,
		}
	}
	snapshot.Window = request.StreamChunkWindowAfter(lastSeq)
	return snapshot, nil
}

func (s *RequestsService) notifyRequestUpdate(requestID string) {
	if requestID == "" {
		return
	}
	val, _ := s.signals.LoadOrStore(requestID, newRequestSignal())
	sig := val.(*requestSignal)
	sig.broadcast()
}

func (s *RequestsService) subscribeRequest(requestID string) <-chan struct{} {
	if requestID == "" {
		closed := make(chan struct{})
		close(closed)
		return closed
	}
	val, _ := s.signals.LoadOrStore(requestID, newRequestSignal())
	sig := val.(*requestSignal)
	return sig.subscribe()
}

func (s *RequestsService) handleExpiredRequest(req *model.Request) {
	if req == nil {
		return
	}
	s.ensureRequestDefaults(req)
	sessionID := req.SessionID
	machineID := req.ExecutingMachineID
	if machineID != "" {
		s.machineService.ReleaseMachineSlot(sessionID, machineID)
	}
	message := "request lease expired"
	now := time.Now()
	s.recordRequestEvent(req, trace.EventRequestLeaseExpired, machineID, map[string]any{
		"visibleAt": req.VisibleAt,
	})
	if req.Attempts >= req.MaxAttempts {
		req.MarkDeadLetter(message)
		req.Error = message
		req.ResultType = model.ResultTypeRejection
		req.Result = map[string]string{"error": message}
		req.VisibleAt = now
		s.recordRequestEvent(req, trace.EventRequestDeadLettered, machineID, map[string]any{
			"reason": message,
		})
	} else {
		req.Status = model.RequestStatusPending
		req.ExecutingMachineID = ""
		req.LeasedBy = ""
		req.LeasedAt = nil
		req.Error = message
		req.LastError = message
		req.Result = nil
		req.ResultType = ""
		next := req.ScheduleRetry()
		if next.Before(now) {
			req.VisibleAt = now.Add(s.retryBackoff)
		}
		s.recordRequestEvent(req, trace.EventRequestRequeued, machineID, map[string]any{
			"reason":        message,
			"nextAttemptAt": req.NextAttemptAt,
		})
	}
	req.UpdatedAt = now

	s.requestsMutex.Lock()
	if _, ok := s.requests[sessionID]; !ok {
		s.requests[sessionID] = make(map[string]*model.Request)
	}
	s.requests[sessionID][req.ID] = req
	s.requestsMutex.Unlock()

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		if err := s.store.SaveRequest(ctx, req); err != nil {
			log.Printf("persist request requeue failed: %v", err)
		}
		cancel()
	}

	s.notifyRequestUpdate(req.ID)
}
