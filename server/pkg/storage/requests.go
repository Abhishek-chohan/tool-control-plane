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
	rows, err := s.db.QueryContext(ctx, `SELECT id, session_id, tool_name, status, input, result, result_type, error, executing_machine_id, meta, stream_results, stream_start_seq, next_stream_seq, attempts, max_attempts, backoff_seconds, visible_at, next_attempt_at, leased_by, leased_at, lease_epoch, timeout_seconds, dead_letter, last_error, created_at, updated_at FROM requests`)
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
	INSERT INTO requests (id, session_id, tool_name, status, input, result, result_type, error, executing_machine_id, meta, stream_results, stream_start_seq, next_stream_seq, attempts, max_attempts, backoff_seconds, visible_at, next_attempt_at, leased_by, leased_at, lease_epoch, timeout_seconds, dead_letter, last_error, created_at, updated_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26)
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
            lease_epoch = EXCLUDED.lease_epoch,
            timeout_seconds = EXCLUDED.timeout_seconds,
            dead_letter = EXCLUDED.dead_letter,
            last_error = EXCLUDED.last_error,
            created_at = EXCLUDED.created_at,
            updated_at = EXCLUDED.updated_at
	`, req.ID, req.SessionID, req.ToolName, string(req.Status), req.Input, resultBytes, nullString(string(req.ResultType)), nullString(req.Error), nullString(req.ExecutingMachineID), metaBytes, streamBytes, req.StreamStartSeq, req.NextStreamSeq, req.Attempts, req.MaxAttempts, req.BackoffSeconds, req.VisibleAt, nextAttempt, nullString(req.LeasedBy), leasedAtVal, req.LeaseEpoch, req.TimeoutSeconds, req.DeadLetter, nullString(req.LastError), req.CreatedAt, req.UpdatedAt)
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
		queryBuilder.WriteString(`SELECT id, session_id, tool_name, status, input, result, result_type, error, executing_machine_id, meta, stream_results, stream_start_seq, next_stream_seq, attempts, max_attempts, backoff_seconds, visible_at, next_attempt_at, leased_by, leased_at, lease_epoch, timeout_seconds, dead_letter, last_error, created_at, updated_at FROM requests WHERE session_id=$1 AND status=$2 AND dead_letter=false AND visible_at <= NOW()`)

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
		// Each successful claim grants a fresh lease epoch: the fencing token
		// that fenced writes (submit/append/update/renew) must echo.
		epoch := req.LeaseEpoch + 1
		if _, err := tx.ExecContext(ctx, `
            UPDATE requests
            SET status=$1,
                executing_machine_id=$2,
                leased_by=$2,
                leased_at=$3,
                attempts=$4,
                visible_at=$5,
                lease_epoch=$6,
                next_attempt_at=NULL,
                last_error=NULL,
                updated_at=$3
            WHERE id=$7
        `, string(model.RequestStatusClaimed), machineID, now, attempts, visible, epoch, req.ID); err != nil {
			return fmt.Errorf("lease request: update: %w", err)
		}

		req.Status = model.RequestStatusClaimed
		req.ExecutingMachineID = machineID
		req.Attempts = attempts
		req.VisibleAt = visible
		req.LeasedBy = machineID
		req.LeasedAt = &now
		req.LeaseEpoch = epoch
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
	// A leased request is reclaimable once either its lease deadline
	// (visible_at, extended by renewals) or its absolute per-attempt deadline
	// (leased_at + timeout_seconds) has passed. Pushing both predicates down
	// and ordering by visible_at keeps the reaper from head-of-line blocking:
	// long-running-but-renewed leases no longer sit at the head of the scan
	// and starve newer expired rows behind the LIMIT window.
	query := `SELECT id, session_id, tool_name, status, input, result, result_type, error, executing_machine_id, meta, stream_results, stream_start_seq, next_stream_seq, attempts, max_attempts, backoff_seconds, visible_at, next_attempt_at, leased_by, leased_at, lease_epoch, timeout_seconds, dead_letter, last_error, created_at, updated_at FROM requests WHERE dead_letter=false AND leased_at IS NOT NULL AND (visible_at <= NOW() OR leased_at + timeout_seconds * INTERVAL '1 second' <= NOW()) ORDER BY visible_at ASC`
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
		if req.LeaseExpired(now) || req.HasTimedOut(now) {
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
	if err := row.Scan(&r.ID, &r.SessionID, &r.ToolName, &r.Status, &r.Input, &resultBytes, &resultType, &errorStr, &execMachineID, &metaBytes, &streamBytes, &r.StreamStartSeq, &r.NextStreamSeq, &r.Attempts, &r.MaxAttempts, &r.BackoffSeconds, &r.VisibleAt, &nextAttempt, &leasedBy, &leasedAt, &r.LeaseEpoch, &r.TimeoutSeconds, &r.DeadLetter, &lastError, &r.CreatedAt, &r.UpdatedAt); err != nil {
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

// requestColumns is the canonical column list used by guarded single-row
// request reads/updates. It must stay in sync with scanRequestRow.
const requestColumns = "id, session_id, tool_name, status, input, result, result_type, error, executing_machine_id, meta, stream_results, stream_start_seq, next_stream_seq, attempts, max_attempts, backoff_seconds, visible_at, next_attempt_at, leased_by, leased_at, lease_epoch, timeout_seconds, dead_letter, last_error, created_at, updated_at"

// GetRequest fetches a single request by ID regardless of session. It supports
// store-backed reads when a request is not present in the local cache, which is
// required for multi-instance visibility.
// ListRequestsBySession returns every request in a session, oldest first,
// regardless of the caller's local cache. The service uses it as the
// read-through when another replica created requests this one has not seen.
func (s *Store) ListRequestsBySession(ctx context.Context, sessionID string) ([]*model.Request, error) {
	if s == nil {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`SELECT %s FROM requests WHERE session_id=$1 ORDER BY created_at ASC`, requestColumns), sessionID)
	if err != nil {
		return nil, fmt.Errorf("query requests by session: %w", err)
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

func (s *Store) GetRequest(ctx context.Context, requestID string) (*model.Request, error) {
	if s == nil {
		return nil, nil
	}
	row := s.db.QueryRowContext(ctx, fmt.Sprintf(`SELECT %s FROM requests WHERE id=$1`, requestColumns), requestID)
	req, err := scanRequestRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get request: %w", err)
	}
	return req, nil
}

// ClaimRequest atomically transitions a pending request to claimed for machineID
// inside a serializable transaction. It is the explicit-claim (by request ID)
// counterpart to LeasePendingRequest (which selects any matching pending row).
// Returns (req, true, nil) on a successful claim, or (nil, false, nil) when the
// request is missing, not pending, or already claimed/draining.
func (s *Store) ClaimRequest(ctx context.Context, sessionID, requestID, machineID string, leaseDuration time.Duration) (*model.Request, bool, error) {
	if s == nil {
		return nil, false, nil
	}
	var claimed *model.Request
	err := s.withSerializableTx(ctx, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx, fmt.Sprintf(`SELECT %s FROM requests WHERE id=$1 FOR UPDATE`, requestColumns), requestID)
		req, err := scanRequestRow(row)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return sql.ErrNoRows
			}
			return err
		}
		if req.SessionID != sessionID || req.Status != model.RequestStatusPending || req.DeadLetter {
			// Not claimable by this caller; report as not-claimed without error.
			return nil
		}
		now := time.Now()
		visible := now.Add(leaseDuration)
		attempts := req.Attempts + 1
		// Each successful claim grants a fresh lease epoch: the fencing token
		// that fenced writes (submit/append/update/renew) must echo.
		epoch := req.LeaseEpoch + 1
		if _, err := tx.ExecContext(ctx, `
            UPDATE requests
            SET status=$1,
                executing_machine_id=$2,
                leased_by=$2,
                leased_at=$3,
                attempts=$4,
                visible_at=$5,
                lease_epoch=$6,
                next_attempt_at=NULL,
                last_error=NULL,
                updated_at=$3
            WHERE id=$7
        `, string(model.RequestStatusClaimed), machineID, now, attempts, visible, epoch, req.ID); err != nil {
			return fmt.Errorf("claim request: update: %w", err)
		}
		req.Status = model.RequestStatusClaimed
		req.ExecutingMachineID = machineID
		req.Attempts = attempts
		req.VisibleAt = visible
		req.LeasedBy = machineID
		req.LeasedAt = &now
		req.LeaseEpoch = epoch
		req.NextAttemptAt = nil
		req.LastError = ""
		req.UpdatedAt = now
		claimed = req
		return nil
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if claimed == nil {
		return nil, false, nil
	}
	return claimed, true, nil
}

// MachineInFlightCount returns the number of requests currently claimed or
// running on the given machine. It is the multi-instance-safe source of truth
// for per-machine capacity enforcement.
func (s *Store) MachineInFlightCount(ctx context.Context, sessionID, machineID string) (int, error) {
	if s == nil {
		return 0, nil
	}
	var count int
	err := s.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM requests
        WHERE executing_machine_id=$1 AND session_id=$2 AND status IN ($3,$4)
    `, machineID, sessionID, string(model.RequestStatusClaimed), string(model.RequestStatusRunning)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("machine in-flight count: %w", err)
	}
	return count, nil
}

// ReclaimExpiredRequest reclaims a single request whose lease has expired,
// inside a guarded transaction. It either marks the request dead-lettered (when
// attempts are exhausted) or requeues it to pending with a retry backoff. It
// returns the updated request and reclaimed=true when this caller won the row,
// or reclaimed=false (nil error) when the request was not expired, already
// terminal, or handled by another caller.
func (s *Store) ReclaimExpiredRequest(ctx context.Context, requestID string, now time.Time, leaseDuration time.Duration, maxAttempts int, backoff time.Duration) (*model.Request, bool, error) {
	if s == nil {
		return nil, false, nil
	}
	var (
		result    *model.Request
		reclaimed bool
	)
	err := s.withSerializableTx(ctx, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx, fmt.Sprintf(`SELECT %s FROM requests WHERE id=$1 FOR UPDATE`, requestColumns), requestID)
		req, err := scanRequestRow(row)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return sql.ErrNoRows
			}
			return err
		}
		// Only claimed/running requests with an active lease are reclaimable.
		// Reclaim is allowed once the lease deadline passed without a renewal,
		// or once the absolute per-attempt timeout was exceeded (renewal moves
		// the lease deadline but never the absolute timeout).
		if req.Status != model.RequestStatusClaimed && req.Status != model.RequestStatusRunning {
			return nil
		}
		if !req.LeaseExpired(now) && !req.HasTimedOut(now) {
			return nil
		}

		updated := *req // shallow copy; we mutate fields below
		updated.ExecutingMachineID = ""
		updated.LeasedBy = ""
		updated.LeasedAt = nil
		updated.NextAttemptAt = nil
		updated.UpdatedAt = now

		if req.Attempts >= maxAttempts {
			// Exhausted: dead-letter and mark terminal-failed.
			updated.DeadLetter = true
			updated.Status = model.RequestStatusFailed
			updated.LastError = "request timed out: max attempts reached"
			updated.Error = updated.LastError
			updated.VisibleAt = now
		} else {
			// Requeue to pending with linear backoff.
			updated.Status = model.RequestStatusPending
			updated.Attempts = req.Attempts + 1
			updated.LastError = "request lease expired"
			retryAt := now.Add(backoff)
			updated.VisibleAt = retryAt
			updated.NextAttemptAt = &retryAt
		}

		if err := persistRequestInTx(ctx, tx, &updated); err != nil {
			return err
		}
		result = &updated
		reclaimed = true
		return nil
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return result, reclaimed, nil
}

// persistRequestInTx writes the full request row within an existing
// transaction. It mirrors SaveRequest but accepts a *sql.Tx so guarded methods
// can persist atomically.
func persistRequestInTx(ctx context.Context, tx *sql.Tx, req *model.Request) error {
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
	if _, err := tx.ExecContext(ctx, `
	INSERT INTO requests (id, session_id, tool_name, status, input, result, result_type, error, executing_machine_id, meta, stream_results, stream_start_seq, next_stream_seq, attempts, max_attempts, backoff_seconds, visible_at, next_attempt_at, leased_by, leased_at, lease_epoch, timeout_seconds, dead_letter, last_error, created_at, updated_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26)
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
            lease_epoch = EXCLUDED.lease_epoch,
            timeout_seconds = EXCLUDED.timeout_seconds,
            dead_letter = EXCLUDED.dead_letter,
            last_error = EXCLUDED.last_error,
            created_at = EXCLUDED.created_at,
            updated_at = EXCLUDED.updated_at
	`, req.ID, req.SessionID, req.ToolName, string(req.Status), req.Input, resultBytes, nullString(string(req.ResultType)), nullString(req.Error), nullString(req.ExecutingMachineID), metaBytes, streamBytes, req.StreamStartSeq, req.NextStreamSeq, req.Attempts, req.MaxAttempts, req.BackoffSeconds, req.VisibleAt, nextAttempt, nullString(req.LeasedBy), leasedAtVal, req.LeaseEpoch, req.TimeoutSeconds, req.DeadLetter, nullString(req.LastError), req.CreatedAt, req.UpdatedAt); err != nil {
		return fmt.Errorf("persist request in tx: %w", err)
	}
	return nil
}
