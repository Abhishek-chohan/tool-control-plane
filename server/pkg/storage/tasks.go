package storage

import (
	"context"
	"database/sql"
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

// ClaimTaskForAdoption atomically claims a non-terminal task for re-adoption by
// this instance, preventing duplicate execution across replicas. It works by
// bumping updated_at inside a guarded transaction: only the first instance to
// call this for a given task wins (subsequent callers see updated_at is too
// recent and return claimed=false). The minAge threshold ensures two instances
// starting simultaneously don't both adopt — the first one's updated_at bump
// excludes the second.
func (s *Store) ClaimTaskForAdoption(ctx context.Context, taskID string, minAge time.Duration) (*model.Task, bool, error) {
	if s == nil {
		return nil, false, nil
	}
	var claimed *model.Task
	err := s.withSerializableTx(ctx, func(tx *sql.Tx) error {
		var updatedAt time.Time
		if err := tx.QueryRowContext(ctx, `SELECT updated_at FROM tasks WHERE id=$1 FOR UPDATE`, taskID).Scan(&updatedAt); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return sql.ErrNoRows
			}
			return fmt.Errorf("claim task adoption: load: %w", err)
		}
		// Only adopt if the task hasn't been touched recently (another instance
		// may have just adopted it).
		cutoff := time.Now().Add(-minAge)
		if updatedAt.After(cutoff) {
			return nil // recently touched; don't claim
		}
		now := time.Now()
		if _, err := tx.ExecContext(ctx, `UPDATE tasks SET updated_at=$1 WHERE id=$2`, now, taskID); err != nil {
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
