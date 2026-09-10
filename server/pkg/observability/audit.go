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

const (
	// auditQueueCapacity bounds queued audit writes. The producing
	// operation never blocks: when the queue is full the event is dropped
	// with a log line — an audit trail must not stall user requests.
	auditQueueCapacity = 256
	// auditWriteTimeout bounds one persistence attempt in the writer
	// goroutine (invisible to producers).
	auditWriteTimeout = 2 * time.Second
)

// AuditRecorder implements trace.SessionTracer, persisting the audited
// event subset through a bounded asynchronous queue. Record never blocks
// the caller; a full queue drops the event with a log line, and
// persistence failures are logged with full event identity.
type AuditRecorder struct {
	store AuditStore
	queue chan *model.AuditEvent
}

// NewAuditRecorder builds the recorder and starts its single writer
// goroutine. Shutdown is driven by ctx: the writer drains queued events
// best-effort and exits.
func NewAuditRecorder(ctx context.Context, store AuditStore) *AuditRecorder {
	r := &AuditRecorder{
		store: store,
		queue: make(chan *model.AuditEvent, auditQueueCapacity),
	}
	go r.writeLoop(ctx)
	return r
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

	select {
	case r.queue <- audit:
	default:
		slog.Warn("audit event dropped: queue full",
			slog.String("event", audit.Event),
			slog.String("sessionId", audit.SessionID),
		)
	}
}

// writeLoop is the single writer: producers only enqueue. The shutdown
// drain is best-effort and write deadlines keep it bounded.
func (r *AuditRecorder) writeLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			for {
				select {
				case event := <-r.queue:
					drainCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), auditWriteTimeout)
					r.persist(drainCtx, event)
					cancel()
				default:
					return
				}
			}
		case event := <-r.queue:
			writeCtx, cancel := context.WithTimeout(ctx, auditWriteTimeout)
			r.persist(writeCtx, event)
			cancel()
		}
	}
}

func (r *AuditRecorder) persist(ctx context.Context, event *model.AuditEvent) {
	if err := r.store.RecordAuditEvent(ctx, event); err != nil {
		slog.ErrorContext(ctx, "audit event not durable",
			slog.String("event", event.Event),
			slog.String("sessionId", event.SessionID),
			slog.String("machineId", event.MachineID),
			slog.String("requestId", event.RequestID),
			slog.String("taskId", event.TaskID),
			slog.Any("err", err),
		)
	}
}
