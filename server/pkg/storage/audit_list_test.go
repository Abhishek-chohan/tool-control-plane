package storage_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
)

// TestListAuditEvents runs the audit listing contract against both
// backends: newest-first ordering with a stable same-tick tiebreak,
// per-field filters, and pagination totals. The fixture uses unique
// session/actor identifiers so concurrent or prior runs against a shared
// database cannot interfere; absolute totals are asserted on the filtered
// set only.
func TestListAuditEvents(t *testing.T) {
	runAgainstBoth(t, "audit list", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		base := time.Now().Add(-time.Minute)
		marker := fmt.Sprintf("audit-%d", base.UnixNano())
		sessionA := "sess-a-" + marker
		sessionB := "sess-b-" + marker
		actorOne := "key-1-" + marker
		actorTwo := "key-2-" + marker

		record := func(sessionID, actor, event string, at time.Time) {
			t.Helper()
			if err := s.RecordAuditEvent(ctx, &model.AuditEvent{
				CreatedAt:  at,
				Event:      event,
				SessionID:  sessionID,
				ActorKeyID: actor,
				Details:    map[string]any{"marker": marker},
			}); err != nil {
				t.Fatalf("record audit event: %v", err)
			}
		}

		// Six events across two sessions and two actors.
		record(sessionA, actorOne, "api_key_created", base)
		record(sessionA, actorOne, "api_key_revoked", base.Add(time.Second))
		record(sessionA, actorTwo, "api_key_created", base.Add(2*time.Second))
		record(sessionB, actorTwo, "session_deleted", base.Add(3*time.Second))
		record(sessionB, actorOne, "session_deleted", base.Add(4*time.Second))
		record(sessionA, "", "request_dead_lettered", base.Add(5*time.Second)) // system-driven: no actor

		sessionFilter := model.AuditEventFilter{SessionID: sessionA, Limit: 10}
		events, total, err := s.ListAuditEvents(ctx, sessionFilter)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if total != 4 || len(events) != 4 {
			t.Fatalf("session filter: total=%d len=%d, want 4", total, len(events))
		}
		// Newest first: the last-recorded event leads.
		if events[0].Event != "request_dead_lettered" {
			t.Fatalf("newest event = %q, want the last-recorded one", events[0].Event)
		}
		for i := 1; i < len(events); i++ {
			prev, cur := events[i-1], events[i]
			if cur.CreatedAt.After(prev.CreatedAt) {
				t.Fatalf("listing not newest-first at %d: %v then %v", i, prev.CreatedAt, cur.CreatedAt)
			}
		}
		for _, e := range events {
			if e.SessionID != sessionA {
				t.Fatalf("filter leaked session %q", e.SessionID)
			}
		}

		// Actor filter: system-driven rows (empty actor) must not match an
		// actor filter.
		_, total, err = s.ListAuditEvents(ctx, model.AuditEventFilter{ActorKeyID: actorTwo, Limit: 10})
		if err != nil {
			t.Fatalf("list by actor: %v", err)
		}
		if total != 2 {
			t.Fatalf("actor filter total=%d, want 2", total)
		}

		// Event filter.
		_, total, err = s.ListAuditEvents(ctx, model.AuditEventFilter{Event: "session_deleted", SessionID: sessionB})
		if err != nil {
			t.Fatalf("list by event: %v", err)
		}
		if total != 2 {
			t.Fatalf("event filter total=%d, want 2", total)
		}

		// Pagination: limit 2 offset 2 walks the middle of the filtered
		// trail with a stable total and no overlap with page one.
		pageOne, _, err := s.ListAuditEvents(ctx, model.AuditEventFilter{SessionID: sessionA, Limit: 2})
		if err != nil {
			t.Fatalf("page one: %v", err)
		}
		pageTwo, total, err := s.ListAuditEvents(ctx, model.AuditEventFilter{SessionID: sessionA, Limit: 2, Offset: 2})
		if err != nil {
			t.Fatalf("page two: %v", err)
		}
		if len(pageTwo) != 2 || total != 4 {
			t.Fatalf("page two len=%d total=%d, want 2/4", len(pageTwo), total)
		}
		seen := map[string]bool{}
		for _, e := range append(append([]*model.AuditEvent{}, pageOne...), pageTwo...) {
			if seen[fmt.Sprintf("%d", e.ID)] {
				t.Fatalf("event %v appeared on both pages", e.ID)
			}
			seen[fmt.Sprintf("%d", e.ID)] = true
		}

		// Round-trip fidelity: details and actor survive the read.
		full, _, err := s.ListAuditEvents(ctx, model.AuditEventFilter{ActorKeyID: actorOne, Limit: 1})
		if err != nil || len(full) != 1 {
			t.Fatalf("fidelity read: %v (%d)", err, len(full))
		}
		if full[0].ActorKeyID != actorOne {
			t.Fatalf("actor lost: %+v", full[0])
		}
		if full[0].Details["marker"] != marker {
			t.Fatalf("details lost: %+v", full[0].Details)
		}
	})
}
