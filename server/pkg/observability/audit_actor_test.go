package observability

import (
	"context"
	"testing"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/trace"
)

// capturingAuditStore records every persisted event for assertions.
type capturingAuditStore struct {
	seen chan *model.AuditEvent
}

func (s *capturingAuditStore) RecordAuditEvent(ctx context.Context, event *model.AuditEvent) error {
	s.seen <- event
	return nil
}

// TestAuditRecorderGraduatesActorToColumn: the acting API key rides the
// trace metadata from the record site and lands in the dedicated
// ActorKeyID column — not duplicated inside details. Events without an
// actor (system-driven) keep the empty value.
func TestAuditRecorderGraduatesActorToColumn(t *testing.T) {
	store := &capturingAuditStore{seen: make(chan *model.AuditEvent, 8)}
	ctx, cancel := context.WithCancel(context.Background())
	recorder := NewAuditRecorder(ctx, store)
	defer func() {
		cancel()
		time.Sleep(50 * time.Millisecond) // let the writer drain
	}()

	recorder.Record(trace.SessionEvent{
		Event:     trace.EventAPIKeyRevoked,
		SessionID: "sess-audit-actor",
		Metadata: map[string]any{
			"actorKeyId": "key-acting-1",
			"reason":     "rotated",
		},
	})
	recorder.Record(trace.SessionEvent{
		Event:     trace.EventRequestDeadLettered,
		SessionID: "sess-audit-actor",
		Metadata:  map[string]any{"attempts": 3},
	})

	select {
	case event := <-store.seen:
		if event.Event != string(trace.EventAPIKeyRevoked) {
			// Ordering between the two queued events is FIFO; tolerate the
			// dead-letter arriving first by consuming both.
			if event.ActorKeyID != "" {
				t.Fatalf("dead-letter event carried actor %q; system events must not", event.ActorKeyID)
			}
			second := <-store.seen
			if second.ActorKeyID != "key-acting-1" {
				t.Fatalf("revocation actor = %q, want key-acting-1", second.ActorKeyID)
			}
			if _, duplicated := second.Details["actorKeyId"]; duplicated {
				t.Fatal("actor must not be duplicated inside details")
			}
			if second.Details["reason"] != "rotated" {
				t.Fatalf("unrelated details lost: %+v", second.Details)
			}
			return
		}
		if event.ActorKeyID != "key-acting-1" {
			t.Fatalf("revocation actor = %q, want key-acting-1", event.ActorKeyID)
		}
		if _, duplicated := event.Details["actorKeyId"]; duplicated {
			t.Fatal("actor must not be duplicated inside details")
		}
		if event.Details["reason"] != "rotated" {
			t.Fatalf("unrelated details lost: %+v", event.Details)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("audit event was not persisted")
	}
}
