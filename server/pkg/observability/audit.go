package observability

import (
	"context"
	"log/slog"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/trace"
)

// AuditStore is the persistence sink for audit events (satisfied by both
// storage implementations).
type AuditStore interface {
	RecordAuditEvent(ctx context.Context, event *model.AuditEvent) error
}

// auditedEvents is the filtered set worth a durable row: credential and
// session lifecycle (who created/removed what), machine registration (who
// can serve work), and terminal dead-letter outcomes (work lost after
// retries). High-frequency execution events are deliberately excluded —
// they belong to metrics, not an audit log.
var auditedEvents = map[trace.SessionEventType]bool{
	trace.EventSessionCreated:      true,
	trace.EventSessionDeleted:      true,
	trace.EventAPIKeyCreated:       true,
	trace.EventAPIKeyRevoked:       true,
	trace.EventMachineRegistered:   true,
	trace.EventMachineUnregistered: true,
	trace.EventRequestDeadLettered: true,
	trace.EventTaskDeadLettered:    true,
}

// AuditRecorder implements trace.SessionTracer, persisting the audited
// event subset. Persistence failures are logged with full event identity
// and dropped: the audit trail must never block or fail the user operation
// that produced the event.
type AuditRecorder struct {
	store AuditStore
	ctx   context.Context
}

// NewAuditRecorder builds the recorder; ctx bounds persistence calls (the
// server's lifecycle context, so shutdown stops in-flight writes).
func NewAuditRecorder(ctx context.Context, store AuditStore) *AuditRecorder {
	return &AuditRecorder{store: store, ctx: ctx}
}

// Record implements trace.SessionTracer.
func (r *AuditRecorder) Record(event trace.SessionEvent) {
	if r == nil || r.store == nil || !auditedEvents[event.Event] {
		return
	}

	audit := &model.AuditEvent{
		CreatedAt: time.Now(),
		Event:     string(event.Event),
		SessionID: event.SessionID,
		MachineID: event.MachineID,
		RequestID: event.RequestID,
		TaskID:    event.TaskID,
		Details:   event.Metadata,
	}

	ctx, cancel := context.WithTimeout(r.ctx, 2*time.Second)
	defer cancel()
	if err := r.store.RecordAuditEvent(ctx, audit); err != nil {
		slog.ErrorContext(r.ctx, "audit event not durable",
			slog.String("event", audit.Event),
			slog.String("sessionId", audit.SessionID),
			slog.String("machineId", audit.MachineID),
			slog.String("requestId", audit.RequestID),
			slog.String("taskId", audit.TaskID),
			slog.Any("err", err),
		)
	}
}
