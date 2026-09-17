package service

import (
	"testing"

	"toolplane/pkg/trace"
)

// TestAuditActorAttribution pins the actor contract at the record sites:
// credential and session lifecycle events carry the acting API key in the
// trace metadata (actorKeyId), which the audit recorder graduates to the
// durable actor column.
func TestAuditActorAttribution(t *testing.T) {
	tracer := &recordingTracer{}
	svc := NewSessionsService(tracer, nil)

	session, err := svc.CreateSession("user-audit-actor", "Audit Actor", "actor attribution coverage", "", "")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	const actor = "key-acting-1"

	apiKey, err := svc.CreateApiKey(session.ID, "minted", "user-audit-actor", actor, []string{"read"}, nil)
	if err != nil {
		t.Fatalf("create api key: %v", err)
	}
	assertActorForKey(t, tracer, trace.EventAPIKeyCreated, actor)

	if err := svc.RevokeApiKey(session.ID, apiKey.ID, actor); err != nil {
		t.Fatalf("revoke api key: %v", err)
	}
	assertActorForKey(t, tracer, trace.EventAPIKeyRevoked, actor)

	if _, err := svc.InvalidateSession(session.ID, "suspected compromise", actor); err != nil {
		t.Fatalf("invalidate session: %v", err)
	}
	if actorOf(t, tracer, trace.EventAPIKeyRevoked) != actor {
		t.Fatal("kill-switch revocation must attribute the acting admin key")
	}

	if err := svc.DeleteSession(session.ID, actor); err != nil {
		t.Fatalf("delete session: %v", err)
	}
	assertActorForKey(t, tracer, trace.EventSessionDeleted, actor)
}

func actorOf(t *testing.T, tracer *recordingTracer, event trace.SessionEventType) string {
	t.Helper()
	tracer.mu.Lock()
	defer tracer.mu.Unlock()
	actor := "<missing>"
	for _, recorded := range tracer.events {
		if recorded.Event == event {
			actor, _ = recorded.Metadata["actorKeyId"].(string)
		}
	}
	return actor
}

func assertActorForKey(t *testing.T, tracer *recordingTracer, event trace.SessionEventType, want string) {
	t.Helper()
	if got := actorOf(t, tracer, event); got != want {
		t.Fatalf("%s actor = %q, want %q", event, got, want)
	}
}
