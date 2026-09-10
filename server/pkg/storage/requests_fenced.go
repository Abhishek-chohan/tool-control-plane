package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"toolplane/pkg/model"
)

// Fenced request writes enforce lease ownership at the storage layer: every
// mutation loads the authoritative row under FOR UPDATE inside a serializable
// transaction, verifies that the caller still holds the current lease grant
// (machine identity + lease epoch), applies the mutation, and persists the row
// in the same transaction. This closes the split-brain window where a stale
// executor (or any non-holder) could write results or chunks after the lease
// was reclaimed, and it removes the blind whole-row upsert race on these paths.

// CheckLeaseFence verifies that machineID/leaseEpoch identify the request's
// current lease grant. Terminal requests have no active lease holder (leased_by
// is cleared when the result lands), so for them the epoch alone identifies the
// last legitimate holder. Shared by the Postgres and in-memory stores.
func CheckLeaseFence(req *model.Request, machineID string, leaseEpoch int64) error {
	if req == nil {
		return sql.ErrNoRows
	}
	if req.LeaseEpoch != leaseEpoch {
		return fmt.Errorf("%w: epoch %d does not match current lease epoch %d", ErrLeaseConflict, leaseEpoch, req.LeaseEpoch)
	}
	if req.Status == model.RequestStatusDone || req.Status == model.RequestStatusFailed {
		// Terminal: the lease was released with the result; the epoch match
		// above proves the caller was the final lease holder.
		return nil
	}
	if req.LeasedBy != machineID {
		return fmt.Errorf("%w: machine %s does not hold the lease (current holder %q)", ErrLeaseConflict, machineID, req.LeasedBy)
	}
	return nil
}

// selectRequestForUpdate loads a single request row locked for update inside
// the transaction and validates that it belongs to the session.
func selectRequestForUpdate(ctx context.Context, tx *sql.Tx, sessionID, requestID string) (*model.Request, error) {
	row := tx.QueryRowContext(ctx, fmt.Sprintf(`SELECT %s FROM requests WHERE id=$1 FOR UPDATE`, requestColumns), requestID)
	req, err := scanRequestRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: request %s not found in session %s", ErrNotFound, requestID, sessionID)
		}
		return nil, err
	}
	if req.SessionID != sessionID {
		return nil, fmt.Errorf("%w: request %s not found in session %s", ErrNotFound, requestID, sessionID)
	}
	return req, nil
}

// selectChunkWindow reads the retained chunk payloads [startSeq, nextSeq)
// from the chunk table inside the transaction. Fenced writes use it to
// rebuild the returned model's window: the request row no longer carries
// payloads, so its loaded model starts with an empty StreamResults and
// appending to it directly would leave StartSeq + len(StreamResults) behind
// NextStreamSeq — an inconsistent window that panics slice-based readers.
func selectChunkWindow(ctx context.Context, tx *sql.Tx, requestID string, startSeq, nextSeq int32) (model.RequestChunkWindow, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT seq, chunk FROM request_chunks WHERE request_id=$1 AND seq >= $2 AND seq < $3 ORDER BY seq ASC`,
		requestID, int(startSeq), int(nextSeq))
	if err != nil {
		return model.RequestChunkWindow{}, fmt.Errorf("select request chunks: %w", err)
	}
	defer rows.Close()

	window := model.RequestChunkWindow{StartSeq: startSeq, NextSeq: nextSeq}
	for rows.Next() {
		var seq int32
		var chunk string
		if err := rows.Scan(&seq, &chunk); err != nil {
			return model.RequestChunkWindow{}, fmt.Errorf("scan request chunk: %w", err)
		}
		window.Chunks = append(window.Chunks, chunk)
	}
	return window, rows.Err()
}

// enforceChunkWindow trims the request's chunk table to the retained window
// bounds — the newest MaxRequestStreamWindowChunks rows whose cumulative
// payload fits MaxRequestStreamWindowBytes — and returns the new window
// start (the smallest surviving seq, or nextSeq when nothing survives).
// Enforcement lives here rather than in the caller's model because each
// fenced call loads the row with empty payloads: only the table knows the
// full history.
func enforceChunkWindow(ctx context.Context, tx *sql.Tx, requestID string, nextSeq int32) (int32, error) {
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM request_chunks
		WHERE request_id=$1 AND seq <= COALESCE((
			SELECT seq FROM (
				SELECT seq,
				       ROW_NUMBER() OVER (ORDER BY seq DESC) AS position,
				       SUM(length(chunk)) OVER (ORDER BY seq DESC) AS running_bytes
				FROM request_chunks WHERE request_id=$1
			) ranked
			WHERE position > $2 OR running_bytes > $3
			ORDER BY seq ASC
			LIMIT 1
		), -1)`, requestID, model.MaxRequestStreamWindowChunks, int64(model.MaxRequestStreamWindowBytes)); err != nil {
		return 0, fmt.Errorf("trim request chunks: %w", err)
	}

	var startSeq sql.NullInt32
	if err := tx.QueryRowContext(ctx,
		`SELECT MIN(seq) FROM request_chunks WHERE request_id=$1`, requestID).Scan(&startSeq); err != nil {
		return 0, fmt.Errorf("read window start: %w", err)
	}
	if !startSeq.Valid {
		return nextSeq, nil
	}
	return startSeq.Int32, nil
}

// RenewRequestLease extends the lease deadline of a claimed/running request.
// Only the current lease holder may renew (machineID + leaseEpoch must match).
// The new lease deadline is min(now + leaseDuration, leased_at +
// timeout_seconds): renewal keeps a healthy executor's lease alive but can
// never push it past the request's absolute per-attempt timeout. Renewing an
// already-expired lease fails with ErrLeaseConflict — once the deadline has
// passed the request may already be reclaimed.
func (s *Store) RenewRequestLease(ctx context.Context, sessionID, requestID, machineID string, leaseEpoch int64, leaseDuration time.Duration) (*model.Request, error) {
	if s == nil {
		return nil, fmt.Errorf("%w: request %s not found in session %s", ErrNotFound, requestID, sessionID)
	}
	var renewed *model.Request
	err := s.withSerializableTx(ctx, func(tx *sql.Tx) error {
		req, err := selectRequestForUpdate(ctx, tx, sessionID, requestID)
		if err != nil {
			return err
		}
		if req.Status != model.RequestStatusClaimed && req.Status != model.RequestStatusRunning {
			return fmt.Errorf("%w: request %s is in state %s and has no renewable lease", ErrLeaseConflict, requestID, req.Status)
		}
		if err := CheckLeaseFence(req, machineID, leaseEpoch); err != nil {
			return err
		}
		now := time.Now()
		if !req.VisibleAt.After(now) {
			return fmt.Errorf("%w: lease deadline already passed for request %s", ErrLeaseConflict, requestID)
		}
		newVisible := now.Add(leaseDuration)
		if req.LeasedAt != nil {
			absoluteDeadline := req.LeasedAt.Add(time.Duration(req.TimeoutSeconds) * time.Second)
			if newVisible.After(absoluteDeadline) {
				newVisible = absoluteDeadline
			}
		}
		if !newVisible.After(now) {
			return fmt.Errorf("%w: renewal cannot extend request %s past its absolute timeout", ErrLeaseConflict, requestID)
		}
		if _, err := tx.ExecContext(ctx, `
            UPDATE requests
            SET visible_at=$1,
                updated_at=$2
            WHERE id=$3
        `, newVisible, now, req.ID); err != nil {
			return fmt.Errorf("renew request lease: update: %w", err)
		}
		req.VisibleAt = newVisible
		req.UpdatedAt = now
		renewed = req
		return nil
	})
	if err != nil {
		return nil, err
	}
	return renewed, nil
}

// SubmitRequestResultFenced writes the terminal result as the current lease
// holder. The streaming special case is preserved: a streaming submission to a
// terminal request appends one trailing chunk when the epoch identifies the
// final lease holder. A non-streaming submission to a terminal request fails
// with ErrRequestTerminal (keeping the historical "already in state" message).
func (s *Store) SubmitRequestResultFenced(ctx context.Context, sessionID, requestID, machineID string, leaseEpoch int64, result interface{}, resultType model.ResultType, meta map[string]string) (*model.Request, error) {
	if s == nil {
		return nil, fmt.Errorf("%w: request %s not found in session %s", ErrNotFound, requestID, sessionID)
	}
	var submitted *model.Request
	err := s.withSerializableTx(ctx, func(tx *sql.Tx) error {
		req, err := selectRequestForUpdate(ctx, tx, sessionID, requestID)
		if err != nil {
			return err
		}
		if err := CheckLeaseFence(req, machineID, leaseEpoch); err != nil {
			return err
		}

		if req.Status == model.RequestStatusDone || req.Status == model.RequestStatusFailed {
			if resultType != model.ResultTypeStreaming {
				return fmt.Errorf("%w: request %s is already in state %s", ErrRequestTerminal, requestID, req.Status)
			}
			// Trailing streaming chunk from the final lease holder.
			if resultStr, ok := result.(string); ok {
				seq := req.AddStreamChunk(resultStr)
				if _, err := tx.ExecContext(ctx,
					`INSERT INTO request_chunks (request_id, seq, chunk, created_at) VALUES ($1,$2,$3,$4)
				 ON CONFLICT (request_id, seq) DO NOTHING`,
					req.ID, int(seq), resultStr, time.Now()); err != nil {
					return fmt.Errorf("insert trailing request chunk: %w", err)
				}
				startSeq, err := enforceChunkWindow(ctx, tx, req.ID, req.NextStreamSeq)
				if err != nil {
					return err
				}
				req.StreamStartSeq = startSeq
				if _, err := tx.ExecContext(ctx,
					`UPDATE requests SET stream_start_seq=$1, next_stream_seq=$2, updated_at=$3 WHERE id=$4`,
					req.StreamStartSeq, req.NextStreamSeq, req.UpdatedAt, req.ID); err != nil {
					return fmt.Errorf("update request stream bookkeeping: %w", err)
				}
				// Rebuild from the table: the row model carries no payloads, so
				// the appended copy alone would desync the window.
				window, err := selectChunkWindow(ctx, tx, req.ID, req.StreamStartSeq, req.NextStreamSeq)
				if err != nil {
					return err
				}
				req.StreamStartSeq = window.StartSeq
				req.StreamResults = window.Chunks
			}
			submitted = req
			return nil
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

		if err := persistRequestInTx(ctx, tx, req); err != nil {
			return err
		}
		submitted = req
		return nil
	})
	if err != nil {
		return nil, err
	}
	return submitted, nil
}

// AppendRequestChunksFenced appends stream chunks as the current lease holder.
// Chunk payloads go to the append-only request_chunks table; the request row
// only takes the sequence bookkeeping (stream_start_seq / next_stream_seq), so
// an append no longer rewrites a whole-row JSONB array that grows with the
// stream. Trimmed window entries are deleted in the same transaction.
func (s *Store) AppendRequestChunksFenced(ctx context.Context, sessionID, requestID, machineID string, leaseEpoch int64, chunks []string) (*model.Request, error) {
	if s == nil {
		return nil, fmt.Errorf("%w: request %s not found in session %s", ErrNotFound, requestID, sessionID)
	}
	var updated *model.Request
	err := s.withSerializableTx(ctx, func(tx *sql.Tx) error {
		req, err := selectRequestForUpdate(ctx, tx, sessionID, requestID)
		if err != nil {
			return err
		}
		if err := CheckLeaseFence(req, machineID, leaseEpoch); err != nil {
			return err
		}

		firstSeq := req.NextStreamSeq
		for _, chunk := range chunks {
			req.AddStreamChunk(chunk)
		}

		// Chunk payloads are append-only rows keyed (request_id, seq).
		for i, chunk := range chunks {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO request_chunks (request_id, seq, chunk, created_at) VALUES ($1,$2,$3,$4)
				 ON CONFLICT (request_id, seq) DO NOTHING`,
				req.ID, int(firstSeq)+i, chunk, time.Now()); err != nil {
				return fmt.Errorf("insert request chunk: %w", err)
			}
		}
		// Enforce the retained window against the table BEFORE writing row
		// bookkeeping: the row model starts each call with empty payloads,
		// so count/byte trimming on it cannot see prior chunks. The enforced
		// start is what the row must advertise.
		startSeq, err := enforceChunkWindow(ctx, tx, req.ID, req.NextStreamSeq)
		if err != nil {
			return err
		}
		req.StreamStartSeq = startSeq

		// Bookkeeping-only row update: no JSONB rewrite.
		if _, err := tx.ExecContext(ctx,
			`UPDATE requests SET stream_start_seq=$1, next_stream_seq=$2, updated_at=$3 WHERE id=$4`,
			req.StreamStartSeq, req.NextStreamSeq, req.UpdatedAt, req.ID); err != nil {
			return fmt.Errorf("update request stream bookkeeping: %w", err)
		}

		// Rebuild the returned model's window from the chunk table. The row
		// no longer carries payloads, so appending to the row-loaded model
		// would desync its window (StartSeq + len(StreamResults) < NextSeq)
		// and panic window readers. The table is authoritative.
		window, err := selectChunkWindow(ctx, tx, req.ID, req.StreamStartSeq, req.NextStreamSeq)
		if err != nil {
			return err
		}
		req.StreamStartSeq = window.StartSeq
		req.StreamResults = window.Chunks
		updated = req
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// UpdateRequestFenced applies a provider status/result transition as the
// current lease holder. The running transition refreshes the lease grant
// (leased_at/visible_at) without clobbering the request's own timeout, so a
// caller-supplied timeout_seconds survives into execution. leaseDuration sets
// the refreshed lease deadline for the running transition.
func (s *Store) UpdateRequestFenced(ctx context.Context, sessionID, requestID, machineID string, leaseEpoch int64, status model.RequestStatus, result interface{}, resultType model.ResultType, leaseDuration time.Duration) (*model.Request, error) {
	if s == nil {
		return nil, fmt.Errorf("%w: request %s not found in session %s", ErrNotFound, requestID, sessionID)
	}
	var updated *model.Request
	err := s.withSerializableTx(ctx, func(tx *sql.Tx) error {
		req, err := selectRequestForUpdate(ctx, tx, sessionID, requestID)
		if err != nil {
			return err
		}
		if err := CheckLeaseFence(req, machineID, leaseEpoch); err != nil {
			return err
		}
		if req.Status == model.RequestStatusDone || req.Status == model.RequestStatusFailed {
			// A terminal request cannot be driven back into execution.
			return ErrRequestTerminal
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

		if err := persistRequestInTx(ctx, tx, req); err != nil {
			return err
		}
		updated = req
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// RequeueRequestFenced returns a claimed/running request to pending when the
// owning machine cannot take the capacity slot. It is a fenced write: only the
// current lease holder's in-flight transition may requeue. The request becomes
// visible again after the retry backoff.
func (s *Store) RequeueRequestFenced(ctx context.Context, sessionID, requestID, machineID string, leaseEpoch int64, reason string, backoff time.Duration) (*model.Request, error) {
	if s == nil {
		return nil, fmt.Errorf("%w: request %s not found in session %s", ErrNotFound, requestID, sessionID)
	}
	var requeued *model.Request
	err := s.withSerializableTx(ctx, func(tx *sql.Tx) error {
		req, err := selectRequestForUpdate(ctx, tx, sessionID, requestID)
		if err != nil {
			return err
		}
		if err := CheckLeaseFence(req, machineID, leaseEpoch); err != nil {
			return err
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
		if err := persistRequestInTx(ctx, tx, req); err != nil {
			return err
		}
		requeued = req
		return nil
	})
	if err != nil {
		return nil, err
	}
	return requeued, nil
}
