package trace

import (
	"encoding/json"
	"log"
	"time"
)

// SessionEventType identifies the category of an emitted event.
type SessionEventType string

const (
	EventSessionCreated            SessionEventType = "session_created"
	EventSessionUpdated            SessionEventType = "session_updated"
	EventSessionDeleted            SessionEventType = "session_deleted"
	EventAPIKeyCreated             SessionEventType = "api_key_created"
	EventAPIKeyRevoked             SessionEventType = "api_key_revoked"
	EventToolRegistered            SessionEventType = "tool_registered"
	EventToolRefreshed             SessionEventType = "tool_refreshed"
	EventToolDeleted               SessionEventType = "tool_deleted"
	EventToolRegistrationRejected  SessionEventType = "tool_registration_rejected"
	EventMachineRegistered         SessionEventType = "machine_registered"
	EventMachineDrainStarted       SessionEventType = "machine_drain_started"
	EventMachineDrainCompleted     SessionEventType = "machine_drain_completed"
	EventMachineUnregistered       SessionEventType = "machine_unregistered"
	EventMachinePingUpdate         SessionEventType = "machine_ping_update"
	EventMachineInactivePruned     SessionEventType = "machine_inactive_pruned"
	EventRequestCreated            SessionEventType = "request_created"
	EventRequestClaimed            SessionEventType = "request_claimed"
	EventRequestExecutionStarted   SessionEventType = "request_execution_started"
	EventRequestExecutionCompleted SessionEventType = "request_execution_completed"
	EventRequestExecutionFailed    SessionEventType = "request_execution_failed"
	EventRequestCancelled          SessionEventType = "request_cancelled"
	EventRequestChunksAppended     SessionEventType = "request_chunks_appended"
	EventRequestLeaseExpired       SessionEventType = "request_lease_expired"
	EventRequestRequeued           SessionEventType = "request_requeued"
	EventRequestDeadLettered       SessionEventType = "request_dead_lettered"
	EventTaskCreated               SessionEventType = "task_created"
	EventTaskExecutionStarted      SessionEventType = "task_execution_started"
	EventTaskRetryScheduled        SessionEventType = "task_retry_scheduled"
	EventTaskExecutionCompleted    SessionEventType = "task_execution_completed"
	EventTaskExecutionFailed       SessionEventType = "task_execution_failed"
	EventTaskCancelled             SessionEventType = "task_cancelled"
	EventTaskDeadLettered          SessionEventType = "task_dead_lettered"
)

// SessionEvent captures a notable lifecycle change for a session.
type SessionEvent struct {
	SessionID string           `json:"sessionId"`
	MachineID string           `json:"machineId"`
	ToolID    string           `json:"toolId"`
	Event     SessionEventType `json:"event"`
	Timestamp time.Time        `json:"timestamp"`
	Metadata  map[string]any   `json:"metadata,omitempty"`
}

// SessionTracer receives lifecycle events for optional diagnostics.
type SessionTracer interface {
	Record(event SessionEvent)
}

type noopTracer struct{}

// Record implements SessionTracer.
func (noopTracer) Record(SessionEvent) {}

// NopTracer returns a tracer that ignores all events.
func NopTracer() SessionTracer {
	return noopTracer{}
}

// LoggingTracer emits events to the provided logger as JSON.
type LoggingTracer struct {
	logger *log.Logger
}

// NewLoggingTracer creates a tracer that logs every event.
func NewLoggingTracer(logger *log.Logger) SessionTracer {
	if logger == nil {
		logger = log.Default()
	}
	return &LoggingTracer{logger: logger}
}

// Record implements SessionTracer.
func (t *LoggingTracer) Record(event SessionEvent) {
	if t == nil || t.logger == nil {
		return
	}
	payload, err := json.Marshal(event)
	if err != nil {
		t.logger.Printf("session_trace marshal error: %v", err)
		return
	}
	t.logger.Printf("session_trace %s", payload)
}
