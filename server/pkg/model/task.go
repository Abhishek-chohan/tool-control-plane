package model

import (
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
	StatusCancelled TaskStatus = "cancelled"
)

// Task represents a scheduled or running task
type Task struct {
	ID             string     `json:"id"`
	SessionID      string     `json:"sessionId"`
	ToolName       string     `json:"toolName"`
	Input          string     `json:"input"`
	Status         TaskStatus `json:"status"`
	Result         string     `json:"result,omitempty"`
	ResultType     string     `json:"resultType,omitempty"`
	Error          string     `json:"error,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	Attempts       int        `json:"attempts"`
	MaxAttempts    int        `json:"maxAttempts"`
	TimeoutSeconds int        `json:"timeoutSeconds"`
	BackoffSeconds int        `json:"backoffSeconds"`
	NextAttemptAt  *time.Time `json:"nextAttemptAt,omitempty"`
	DeadLetter     bool       `json:"deadLetter"`
	LastError      string     `json:"lastError,omitempty"`
}

// NewTask creates a new task
func NewTask(sessionID, toolName, input string) *Task {
	now := time.Now()
	return &Task{
		ID:             uuid.New().String(),
		SessionID:      sessionID,
		ToolName:       toolName,
		Input:          input,
		Status:         StatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
		Attempts:       0,
		MaxAttempts:    defaultTaskMaxAttempts,
		TimeoutSeconds: defaultTaskTimeoutSeconds,
		BackoffSeconds: defaultTaskBackoffSeconds,
	}
}

const (
	defaultTaskMaxAttempts    = 3
	defaultTaskTimeoutSeconds = 60
	defaultTaskBackoffSeconds = 5
)

// ScheduleRetry updates the task with the next backoff attempt.
func (t *Task) ScheduleRetry() time.Time {
	t.Attempts++
	backoff := time.Duration(t.BackoffSeconds*t.Attempts) * time.Second
	next := time.Now().Add(backoff)
	t.NextAttemptAt = &next
	t.UpdatedAt = time.Now()
	return next
}

// MarkDeadLetter ends the task after exhausting retries.
func (t *Task) MarkDeadLetter(errMsg string) {
	t.Status = StatusFailed
	t.DeadLetter = true
	t.LastError = errMsg
	t.UpdatedAt = time.Now()
}
