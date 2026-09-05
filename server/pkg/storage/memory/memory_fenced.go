package memory

import (
	"context"
	"fmt"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
)

// Fenced request writes for the in-memory store. They mirror the Postgres
// guarded-transaction semantics: look up the authoritative row, verify the
// caller still holds the current lease grant via storage.CheckLeaseFence,
// apply the mutation, and return a clone. All of it happens under the store
// lock, so check and mutation are atomic here.

func (s *Store) lookupFenced(sessionID, requestID string) (*model.Request, error) {
	r, ok := s.requests[requestID]
	if !ok || r.SessionID != sessionID {
		return nil, fmt.Errorf("request %s not found in session %s", requestID, sessionID)
	}
	return r, nil
}

func (s *Store) RenewRequestLease(ctx context.Context, sessionID, requestID, machineID string, leaseEpoch int64, leaseDuration time.Duration) (*model.Request, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	req, err := s.lookupFenced(sessionID, requestID)
	if err != nil {
		return nil, err
	}
	if req.Status != model.RequestStatusClaimed && req.Status != model.RequestStatusRunning {
		return nil, fmt.Errorf("%w: request %s is in state %s and has no renewable lease", storage.ErrLeaseConflict, requestID, req.Status)
	}
	if err := storage.CheckLeaseFence(req, machineID, leaseEpoch); err != nil {
		return nil, err
	}
	now := time.Now()
	if !req.VisibleAt.After(now) {
		return nil, fmt.Errorf("%w: lease deadline already passed for request %s", storage.ErrLeaseConflict, requestID)
	}
	newVisible := now.Add(leaseDuration)
	if req.LeasedAt != nil {
		absoluteDeadline := req.LeasedAt.Add(time.Duration(req.TimeoutSeconds) * time.Second)
		if newVisible.After(absoluteDeadline) {
			newVisible = absoluteDeadline
		}
	}
	if !newVisible.After(now) {
		return nil, fmt.Errorf("%w: renewal cannot extend request %s past its absolute timeout", storage.ErrLeaseConflict, requestID)
	}
	req.VisibleAt = newVisible
	req.UpdatedAt = now
	return cloneRequest(req), nil
}

func (s *Store) SubmitRequestResultFenced(ctx context.Context, sessionID, requestID, machineID string, leaseEpoch int64, result interface{}, resultType model.ResultType, meta map[string]string) (*model.Request, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	req, err := s.lookupFenced(sessionID, requestID)
	if err != nil {
		return nil, err
	}
	if err := storage.CheckLeaseFence(req, machineID, leaseEpoch); err != nil {
		return nil, err
	}

	if req.Status == model.RequestStatusDone || req.Status == model.RequestStatusFailed {
		if resultType != model.ResultTypeStreaming {
			return nil, fmt.Errorf("%w: request %s is already in state %s", storage.ErrRequestTerminal, requestID, req.Status)
		}
		if resultStr, ok := result.(string); ok {
			req.AddStreamChunk(resultStr)
		}
		return cloneRequest(req), nil
	}

	if meta != nil {
		if req.Meta == nil {
			req.Meta = make(map[string]string)
		}
		for k, v := range meta {
			req.Meta[k] = v
		}
	}

	errorMsg := ""
	if resultType == model.ResultTypeRejection {
		if errObj, ok := result.(map[string]interface{}); ok {
			if msg, ok := errObj["error"]; ok {
				if errMsg, ok := msg.(string); ok {
					errorMsg = errMsg
				}
			}
		}
	}

	req.SetResult(result, resultType, errorMsg)
	if resultType == model.ResultTypeRejection {
		req.LastError = errorMsg
	} else {
		req.LastError = ""
	}
	req.DeadLetter = false
	return cloneRequest(req), nil
}

func (s *Store) AppendRequestChunksFenced(ctx context.Context, sessionID, requestID, machineID string, leaseEpoch int64, chunks []string) (*model.Request, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	req, err := s.lookupFenced(sessionID, requestID)
	if err != nil {
		return nil, err
	}
	if err := storage.CheckLeaseFence(req, machineID, leaseEpoch); err != nil {
		return nil, err
	}
	for _, chunk := range chunks {
		req.AddStreamChunk(chunk)
	}
	return cloneRequest(req), nil
}

func (s *Store) UpdateRequestFenced(ctx context.Context, sessionID, requestID, machineID string, leaseEpoch int64, status model.RequestStatus, result interface{}, resultType model.ResultType, leaseDuration time.Duration) (*model.Request, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	req, err := s.lookupFenced(sessionID, requestID)
	if err != nil {
		return nil, err
	}
	if err := storage.CheckLeaseFence(req, machineID, leaseEpoch); err != nil {
		return nil, err
	}
	if req.Status == model.RequestStatusDone || req.Status == model.RequestStatusFailed {
		// A terminal request cannot be driven back into execution.
		return nil, storage.ErrRequestTerminal
	}

	now := time.Now()
	if status != "" && status != req.Status {
		switch status {
		case model.RequestStatusRunning:
			req.Status = status
			req.ExecutingMachineID = machineID
			req.LeasedBy = machineID
			req.LeasedAt = &now
			req.VisibleAt = now.Add(leaseDuration)
			req.NextAttemptAt = nil
			req.Error = ""
			req.LastError = ""
		case model.RequestStatusDone, model.RequestStatusFailed:
			req.Status = status
			if status == model.RequestStatusFailed && req.Error != "" {
				req.LastError = req.Error
			}
			req.LeasedBy = ""
			req.LeasedAt = nil
			req.VisibleAt = now
			req.NextAttemptAt = nil
		default:
			req.Status = status
		}
	}

	if result != nil {
		req.Result = result
	}
	if resultType != "" {
		req.ResultType = resultType
	}
	req.UpdatedAt = now
	return cloneRequest(req), nil
}

func (s *Store) RequeueRequestFenced(ctx context.Context, sessionID, requestID, machineID string, leaseEpoch int64, reason string, backoff time.Duration) (*model.Request, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	req, err := s.lookupFenced(sessionID, requestID)
	if err != nil {
		return nil, err
	}
	if err := storage.CheckLeaseFence(req, machineID, leaseEpoch); err != nil {
		return nil, err
	}
	now := time.Now()
	retryAt := now.Add(backoff)
	req.Status = model.RequestStatusPending
	req.ExecutingMachineID = ""
	req.LeasedBy = ""
	req.LeasedAt = nil
	req.VisibleAt = retryAt
	req.NextAttemptAt = &retryAt
	req.Error = reason
	req.LastError = reason
	req.UpdatedAt = now
	return cloneRequest(req), nil
}
