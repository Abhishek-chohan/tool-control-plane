package storage

import (
	"context"
	"database/sql"
	"fmt"

	"toolplane/pkg/model"
)

func (s *Store) AllTasks(ctx context.Context) ([]*model.Task, error) {
	if s == nil {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, session_id, tool_name, status, input, result, result_type, error, attempts, max_attempts, backoff_seconds, next_attempt_at, timeout_seconds, dead_letter, last_error, created_at, updated_at, completed_at FROM tasks`)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		t := &model.Task{}
		var result, resultType, errorText sql.NullString
		var completedAt sql.NullTime
		var nextAttempt sql.NullTime
		var lastError sql.NullString
		if err := rows.Scan(&t.ID, &t.SessionID, &t.ToolName, &t.Status, &t.Input, &result, &resultType, &errorText, &t.Attempts, &t.MaxAttempts, &t.BackoffSeconds, &nextAttempt, &t.TimeoutSeconds, &t.DeadLetter, &lastError, &t.CreatedAt, &t.UpdatedAt, &completedAt); err != nil {
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
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (s *Store) SaveTask(ctx context.Context, task *model.Task) error {
	if s == nil || task == nil {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
        INSERT INTO tasks (id, session_id, tool_name, status, input, result, result_type, error, attempts, max_attempts, backoff_seconds, next_attempt_at, timeout_seconds, dead_letter, last_error, created_at, updated_at, completed_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
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
            created_at = EXCLUDED.created_at,
            updated_at = EXCLUDED.updated_at,
            completed_at = EXCLUDED.completed_at
    `, task.ID, task.SessionID, task.ToolName, string(task.Status), task.Input, nullString(task.Result), nullString(task.ResultType), nullString(task.Error), task.Attempts, task.MaxAttempts, task.BackoffSeconds, nullableTime(task.NextAttemptAt), task.TimeoutSeconds, task.DeadLetter, nullString(task.LastError), task.CreatedAt, task.UpdatedAt, nullableTime(task.CompletedAt))
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
