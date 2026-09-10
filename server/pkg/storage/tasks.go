package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"toolplane/pkg/model"
)

// taskColumns is the single column list for task SELECTs; scanTaskRow is the
// matching scanner. Keep them in sync with SaveTask's INSERT columns.
const taskColumns = `id, session_id, tool_name, status, input, result, result_type, error, attempts, max_attempts, backoff_seconds, next_attempt_at, timeout_seconds, dead_letter, last_error, current_request_id, idempotency_key, created_at, updated_at, completed_at`

func (s *Store) AllTasks(ctx context.Context) ([]*model.Task, error) {
	if s == nil {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+taskColumns+` FROM tasks`)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		t, err := scanTaskRow(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (s *Store) SaveTask(ctx context.Context, task *model.Task) error {
	if s == nil || task == nil {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
        INSERT INTO tasks (id, session_id, tool_name, status, input, result, result_type, error, attempts, max_attempts, backoff_seconds, next_attempt_at, timeout_seconds, dead_letter, last_error, current_request_id, idempotency_key, created_at, updated_at, completed_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
        ON CONFLICT (id) DO UPDATE SET
            session_id = EXCLUDED.session_id,
            tool_name = EXCLUDED.tool_name,
            status = EXCLUDED.status,
            input = EXCLUDED.input,
            result = EXCLUDED.result,
            result_type = EXCLUDED.result_type,
            error = EXCLUDED.error,
            attempts = EXCLUDED.attempts,
            max_attempts = EXCLUDED.max_attempts,
            backoff_seconds = EXCLUDED.backoff_seconds,
            next_attempt_at = EXCLUDED.next_attempt_at,
            timeout_seconds = EXCLUDED.timeout_seconds,
            dead_letter = EXCLUDED.dead_letter,
            last_error = EXCLUDED.last_error,
            current_request_id = EXCLUDED.current_request_id,
            idempotency_key = EXCLUDED.idempotency_key,
            created_at = EXCLUDED.created_at,
            updated_at = EXCLUDED.updated_at,
            completed_at = EXCLUDED.completed_at
    `, task.ID, task.SessionID, task.ToolName, string(task.Status), task.Input, nullString(task.Result), nullString(task.ResultType), nullString(task.Error), task.Attempts, task.MaxAttempts, task.BackoffSeconds, nullableTime(task.NextAttemptAt), task.TimeoutSeconds, task.DeadLetter, nullString(task.LastError), nullString(task.CurrentRequestID), nullString(task.IdempotencyKey), task.CreatedAt, task.UpdatedAt, nullableTime(task.CompletedAt))
	if err != nil {
		return fmt.Errorf("upsert task: %w", err)
	}
	return nil
}

func (s *Store) DeleteTask(ctx context.Context, taskID string) error {
	if s == nil {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM tasks WHERE id=$1`, taskID); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}

// FindNonTerminalTasks returns tasks that are not in a terminal state
// (completed/failed/cancelled). Used on startup to re-adopt in-flight work so
// a task interrupted by an instance restart resumes instead of stalling.
// FindAdoptableTasks returns up to limit due tasks whose adoption lease is
// free: unowned, or owned but untouched for leaseTTL. Oldest due first, so
// starved work wins ties.
func (s *Store) FindAdoptableTasks(ctx context.Context, now time.Time, leaseTTL time.Duration, limit int) ([]*model.Task, error) {
	if s == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 32
	}
	leaseCutoff := now.Add(-leaseTTL)
	rows, err := s.db.QueryContext(ctx, `SELECT `+taskColumns+` FROM tasks
		WHERE status NOT IN ($1,$2,$3) AND dead_letter=false
		AND (next_attempt_at IS NULL OR next_attempt_at <= $4)
		AND (adopted_by IS NULL OR adopted_by = '' OR updated_at <= $5)
		ORDER BY COALESCE(next_attempt_at, created_at) ASC
		LIMIT $6`,
		string(model.StatusCompleted), string(model.StatusFailed), string(model.StatusCancelled), now, leaseCutoff, limit)
	if err != nil {
		return nil, fmt.Errorf("find adoptable tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		task, err := scanTaskRow(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

// GetTaskByIdempotencyKey fetches the task a session created under the given
// dedup key (nil when absent). CreateTask uses it to make retries return the
// original task without re-executing it.
func (s *Store) GetTaskByIdempotencyKey(ctx context.Context, sessionID, idempotencyKey string) (*model.Task, error) {
	if s == nil || idempotencyKey == "" {
		return nil, nil
	}
	row := s.db.QueryRowContext(ctx, `SELECT `+taskColumns+` FROM tasks WHERE session_id=$1 AND idempotency_key=$2`, sessionID, idempotencyKey)
	task, err := scanTaskRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get task by idempotency key: %w", err)
	}
	return task, nil
}

func (s *Store) FindNonTerminalTasks(ctx context.Context) ([]*model.Task, error) {
	if s == nil {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+taskColumns+` FROM tasks WHERE status NOT IN ($1,$2,$3) AND dead_letter=false`, string(model.StatusCompleted), string(model.StatusFailed), string(model.StatusCancelled))
	if err != nil {
		return nil, fmt.Errorf("find non-terminal tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		t, err := scanTaskRow(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// ClaimTaskForAdoption atomically acquires execution ownership of a task:
// it succeeds when the task is unowned or the previous owner's lease
// (leaseTTL since its last touch) has expired, and records instanceID as the
// owner. Only the first caller wins; it is the fencing gate for task
// execution across replicas.
func (s *Store) ClaimTaskForAdoption(ctx context.Context, taskID, instanceID string, leaseTTL time.Duration) (*model.Task, bool, error) {
	if s == nil {
		return nil, false, nil
	}
	var claimed *model.Task
	err := s.withSerializableTx(ctx, func(tx *sql.Tx) error {
		var adoptedBy sql.NullString
		var updatedAt time.Time
		if err := tx.QueryRowContext(ctx, `SELECT adopted_by, updated_at FROM tasks WHERE id=$1 FOR UPDATE`, taskID).Scan(&adoptedBy, &updatedAt); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return sql.ErrNoRows
			}
			return fmt.Errorf("claim task adoption: load: %w", err)
		}
		now := time.Now()
		// Acquire when unowned, or when the previous owner's lease (its
		// last touch + TTL) has expired: a live owner keeps the task.
		if adoptedBy.Valid && adoptedBy.String != "" && updatedAt.After(now.Add(-leaseTTL)) {
			return nil // owned and leased; don't steal
		}
		if _, err := tx.ExecContext(ctx, `UPDATE tasks SET adopted_by=$1, updated_at=$2 WHERE id=$3`, instanceID, now, taskID); err != nil {
			return fmt.Errorf("claim task adoption: update: %w", err)
		}
		row := tx.QueryRowContext(ctx, `SELECT `+taskColumns+` FROM tasks WHERE id=$1`, taskID)
		t, err := scanTaskRow(row)
		if err != nil {
			return err
		}
		claimed = t
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

// RecordAuditEvent persists one audit row. Details are encoded best-effort:
// values JSON cannot represent are dropped rather than failing the write.
func (s *Store) RecordAuditEvent(ctx context.Context, event *model.AuditEvent) error {
	if s == nil || event == nil {
		return nil
	}
	// map[string]any is not a bindable driver value: marshal to JSON like
	// every other JSONB write. Unrepresentable values drop the details
	// rather than failing the row.
	var details any
	if len(event.Details) > 0 {
		if encoded, err := json.Marshal(event.Details); err == nil {
			details = encoded
		}
	}
	if _, err := s.db.ExecContext(ctx, `
        INSERT INTO audit_events (created_at, event, session_id, machine_id, request_id, task_id, details)
        VALUES ($1,$2,$3,$4,$5,$6,$7)
    `, event.CreatedAt, event.Event, nullString(event.SessionID), nullString(event.MachineID), nullString(event.RequestID), nullString(event.TaskID), details); err != nil {
		return fmt.Errorf("record audit event: %w", err)
	}
	return nil
}

// RenewTaskAdoption refreshes the owning instance's lease. It returns false
// when ownership was lost (the lease expired and another instance claimed
// the task), at which point the previous owner must stop executing.
func (s *Store) RenewTaskAdoption(ctx context.Context, taskID, instanceID string) (bool, error) {
	if s == nil {
		return false, nil
	}
	result, err := s.db.ExecContext(ctx, `UPDATE tasks SET updated_at=$1 WHERE id=$2 AND adopted_by=$3`, time.Now(), taskID, instanceID)
	if err != nil {
		return false, fmt.Errorf("renew task adoption: %w", err)
	}
	renewed, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("renew task adoption: rows affected: %w", err)
	}
	return renewed > 0, nil
}

// ReleaseTaskAdoption clears ownership so any instance can pick the task up
// (retry scheduling, terminal states). Only the recorded owner may release.
func (s *Store) ReleaseTaskAdoption(ctx context.Context, taskID, instanceID string) error {
	if s == nil {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE tasks SET adopted_by='', updated_at=$1 WHERE id=$2 AND adopted_by=$3`, time.Now(), taskID, instanceID); err != nil {
		return fmt.Errorf("release task adoption: %w", err)
	}
	return nil
}

// scanTaskRow scans a single task row from a QueryRow or rows.Next result.
func scanTaskRow(row interface {
	Scan(dest ...interface{}) error
}) (*model.Task, error) {
	t := &model.Task{}
	var result, resultType, errorText sql.NullString
	var completedAt sql.NullTime
	var nextAttempt sql.NullTime
	var lastError sql.NullString
	var currentRequestID sql.NullString
	var idempotencyKey sql.NullString
	if err := row.Scan(&t.ID, &t.SessionID, &t.ToolName, &t.Status, &t.Input, &result, &resultType, &errorText, &t.Attempts, &t.MaxAttempts, &t.BackoffSeconds, &nextAttempt, &t.TimeoutSeconds, &t.DeadLetter, &lastError, &currentRequestID, &idempotencyKey, &t.CreatedAt, &t.UpdatedAt, &completedAt); err != nil {
		return nil, fmt.Errorf("scan task: %w", err)
	}
	if result.Valid {
		t.Result = result.String
	}
	if resultType.Valid {
		t.ResultType = resultType.String
	}
	if errorText.Valid {
		t.Error = errorText.String
	}
	if completedAt.Valid {
		ct := completedAt.Time
		t.CompletedAt = &ct
	}
	if nextAttempt.Valid {
		nt := nextAttempt.Time
		t.NextAttemptAt = &nt
	}
	if lastError.Valid {
		t.LastError = lastError.String
	}
	if idempotencyKey.Valid {
		t.IdempotencyKey = idempotencyKey.String
	}
	if currentRequestID.Valid {
		t.CurrentRequestID = currentRequestID.String
	}
	return t, nil
}
