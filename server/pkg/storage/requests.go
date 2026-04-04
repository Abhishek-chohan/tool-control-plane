package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"toolplane/pkg/model"
)

func (s *Store) AllRequests(ctx context.Context) ([]*model.Request, error) {
	if s == nil {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, session_id, tool_name, status, input, result, result_type, error, executing_machine_id, meta, stream_results, stream_start_seq, next_stream_seq, attempts, max_attempts, backoff_seconds, visible_at, next_attempt_at, leased_by, leased_at, timeout_seconds, dead_letter, last_error, created_at, updated_at FROM requests`)
	if err != nil {
		return nil, fmt.Errorf("query requests: %w", err)
	}
	defer rows.Close()

	var requests []*model.Request
	for rows.Next() {
		req, err := scanRequestRow(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, rows.Err()
}

func (s *Store) SaveRequest(ctx context.Context, req *model.Request) error {
	if s == nil || req == nil {
		return nil
	}
	metaBytes := toJSON(req.Meta)
	if req.Meta == nil {
		metaBytes = []byte("{}")
	}
	streamBytes := toJSON(req.StreamResults)
	if req.StreamResults == nil {
		streamBytes = []byte("[]")
	}
	resultBytes := []byte("null")
	if req.Result != nil {
		resultBytes = toJSON(req.Result)
	}
	leasedAtVal := nullableTime(req.LeasedAt)
	nextAttempt := nullableTime(req.NextAttemptAt)
	_, err := s.db.ExecContext(ctx, `
	INSERT INTO requests (id, session_id, tool_name, status, input, result, result_type, error, executing_machine_id, meta, stream_results, stream_start_seq, next_stream_seq, attempts, max_attempts, backoff_seconds, visible_at, next_attempt_at, leased_by, leased_at, timeout_seconds, dead_letter, last_error, created_at, updated_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25)
        ON CONFLICT (id) DO UPDATE SET
            session_id = EXCLUDED.session_id,
            tool_name = EXCLUDED.tool_name,
            status = EXCLUDED.status,
            input = EXCLUDED.input,
            result = EXCLUDED.result,
            result_type = EXCLUDED.result_type,
            error = EXCLUDED.error,
            executing_machine_id = EXCLUDED.executing_machine_id,
            meta = EXCLUDED.meta,
            stream_results = EXCLUDED.stream_results,
			stream_start_seq = EXCLUDED.stream_start_seq,
			next_stream_seq = EXCLUDED.next_stream_seq,
            attempts = EXCLUDED.attempts,
            max_attempts = EXCLUDED.max_attempts,
            backoff_seconds = EXCLUDED.backoff_seconds,
            visible_at = EXCLUDED.visible_at,
            next_attempt_at = EXCLUDED.next_attempt_at,
            leased_by = EXCLUDED.leased_by,
            leased_at = EXCLUDED.leased_at,
            timeout_seconds = EXCLUDED.timeout_seconds,
            dead_letter = EXCLUDED.dead_letter,
            last_error = EXCLUDED.last_error,
            created_at = EXCLUDED.created_at,
            updated_at = EXCLUDED.updated_at
	`, req.ID, req.SessionID, req.ToolName, string(req.Status), req.Input, resultBytes, nullString(string(req.ResultType)), nullString(req.Error), nullString(req.ExecutingMachineID), metaBytes, streamBytes, req.StreamStartSeq, req.NextStreamSeq, req.Attempts, req.MaxAttempts, req.BackoffSeconds, req.VisibleAt, nextAttempt, nullString(req.LeasedBy), leasedAtVal, req.TimeoutSeconds, req.DeadLetter, nullString(req.LastError), req.CreatedAt, req.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert request: %w", err)
	}
	return nil
}

func (s *Store) LeasePendingRequest(ctx context.Context, sessionID, machineID string, toolNames []string, leaseDuration time.Duration) (*model.Request, error) {
	if s == nil {
		return nil, nil
	}

	var selected *model.Request
	err := s.withSerializableTx(ctx, func(tx *sql.Tx) error {
		queryBuilder := strings.Builder{}
		args := []interface{}{sessionID, string(model.RequestStatusPending)}
		queryBuilder.WriteString(`SELECT id, session_id, tool_name, status, input, result, result_type, error, executing_machine_id, meta, stream_results, stream_start_seq, next_stream_seq, attempts, max_attempts, backoff_seconds, visible_at, next_attempt_at, leased_by, leased_at, timeout_seconds, dead_letter, last_error, created_at, updated_at FROM requests WHERE session_id=$1 AND status=$2 AND dead_letter=false AND visible_at <= NOW()`)

		if len(toolNames) > 0 {
			queryBuilder.WriteString(" AND tool_name IN (")
			for i, name := range toolNames {
				placeholder := fmt.Sprintf("$%d", len(args)+1)
				queryBuilder.WriteString(placeholder)
				if i < len(toolNames)-1 {
					queryBuilder.WriteString(",")
				}
				args = append(args, name)
			}
			queryBuilder.WriteString(")")
		}

		queryBuilder.WriteString(" ORDER BY created_at ASC FOR UPDATE SKIP LOCKED LIMIT 1")

		row := tx.QueryRowContext(ctx, queryBuilder.String(), args...)
		req, err := scanRequestRow(row)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return sql.ErrNoRows
			}
			return err
		}

		now := time.Now()
		visible := now.Add(leaseDuration)
		attempts := req.Attempts + 1
		if _, err := tx.ExecContext(ctx, `
            UPDATE requests
            SET status=$1,
                executing_machine_id=$2,
                leased_by=$2,
                leased_at=$3,
                attempts=$4,
                visible_at=$5,
                next_attempt_at=NULL,
                last_error=NULL,
                updated_at=$3
            WHERE id=$6
        `, string(model.RequestStatusClaimed), machineID, now, attempts, visible, req.ID); err != nil {
			return fmt.Errorf("lease request: update: %w", err)
		}

		req.Status = model.RequestStatusClaimed
		req.ExecutingMachineID = machineID
		req.Attempts = attempts
		req.VisibleAt = visible
		req.LeasedBy = machineID
		req.LeasedAt = &now
		req.NextAttemptAt = nil
		req.LastError = ""
		req.UpdatedAt = now
		selected = req
		return nil
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return selected, nil
}

func (s *Store) MarkRequestRunning(ctx context.Context, requestID, machineID string, timeout time.Duration) error {
	if s == nil {
		return nil
	}
	now := time.Now()
	deadline := now.Add(timeout)
	if _, err := s.db.ExecContext(ctx, `
        UPDATE requests
        SET status=$1,
            executing_machine_id=$2,
            leased_by=$2,
            leased_at=$3,
            visible_at=$4,
            updated_at=$3,
            next_attempt_at=NULL
        WHERE id=$5
    `, string(model.RequestStatusRunning), machineID, now, deadline, requestID); err != nil {
		return fmt.Errorf("mark request running: %w", err)
	}
	return nil
}

func (s *Store) FindExpiredRequests(ctx context.Context, limit int) ([]*model.Request, error) {
	if s == nil {
		return nil, nil
	}
	query := `SELECT id, session_id, tool_name, status, input, result, result_type, error, executing_machine_id, meta, stream_results, stream_start_seq, next_stream_seq, attempts, max_attempts, backoff_seconds, visible_at, next_attempt_at, leased_by, leased_at, timeout_seconds, dead_letter, last_error, created_at, updated_at FROM requests WHERE dead_letter=false AND leased_at IS NOT NULL ORDER BY leased_at ASC`
	args := []interface{}{}
	if limit > 0 {
		query += " LIMIT $1"
		args = append(args, limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("find expired requests: %w", err)
	}
	defer rows.Close()

	now := time.Now()
	var expired []*model.Request
	for rows.Next() {
		req, err := scanRequestRow(rows)
		if err != nil {
			return nil, err
		}
		if req.HasTimedOut(now) {
			expired = append(expired, req)
		}
	}
	return expired, rows.Err()
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanRequestRow(row rowScanner) (*model.Request, error) {
	r := &model.Request{}
	var resultBytes, metaBytes, streamBytes []byte
	var resultType, errorStr, execMachineID sql.NullString
	var leasedBy sql.NullString
	var leasedAt sql.NullTime
	var nextAttempt sql.NullTime
	var lastError sql.NullString
	if err := row.Scan(&r.ID, &r.SessionID, &r.ToolName, &r.Status, &r.Input, &resultBytes, &resultType, &errorStr, &execMachineID, &metaBytes, &streamBytes, &r.StreamStartSeq, &r.NextStreamSeq, &r.Attempts, &r.MaxAttempts, &r.BackoffSeconds, &r.VisibleAt, &nextAttempt, &leasedBy, &leasedAt, &r.TimeoutSeconds, &r.DeadLetter, &lastError, &r.CreatedAt, &r.UpdatedAt); err != nil {
		return nil, fmt.Errorf("scan request: %w", err)
	}
	if len(resultBytes) > 0 {
		var raw interface{}
		if err := json.Unmarshal(resultBytes, &raw); err == nil {
			r.Result = raw
		}
	}
	if resultType.Valid {
		r.ResultType = model.ResultType(resultType.String)
	}
	if errorStr.Valid {
		r.Error = errorStr.String
	}
	if execMachineID.Valid {
		r.ExecutingMachineID = execMachineID.String
	}
	if len(metaBytes) > 0 {
		if err := json.Unmarshal(metaBytes, &r.Meta); err != nil {
			return nil, fmt.Errorf("decode request meta: %w", err)
		}
	}
	if r.Meta == nil {
		r.Meta = make(map[string]string)
	}
	if len(streamBytes) > 0 {
		if err := json.Unmarshal(streamBytes, &r.StreamResults); err != nil {
			return nil, fmt.Errorf("decode request stream: %w", err)
		}
	}
	if leasedBy.Valid {
		r.LeasedBy = leasedBy.String
	}
	if leasedAt.Valid {
		t := leasedAt.Time
		r.LeasedAt = &t
	}
	if nextAttempt.Valid {
		t := nextAttempt.Time
		r.NextAttemptAt = &t
	}
	if lastError.Valid {
		r.LastError = lastError.String
	}
	return r, nil
}
