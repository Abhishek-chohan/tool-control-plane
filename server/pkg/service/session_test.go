package service

import (
	"testing"

	"toolplane/pkg/trace"
)

func TestSessionsServiceRecordsAuditEvents(t *testing.T) {
	tracer := &recordingTracer{}
	svc := NewSessionsService(tracer, nil)

	session, err := svc.CreateSession("user-audit", "Audit Session", "session for trace coverage", "", "", "tenant-a")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if _, err := svc.UpdateSession(session.ID, "Audit Session Updated", "updated", "tenant-b"); err != nil {
		t.Fatalf("update session: %v", err)
	}

	apiKey, err := svc.CreateApiKey(session.ID, "primary", "user-audit")
	if err != nil {
		t.Fatalf("create api key: %v", err)
	}

	if err := svc.RevokeApiKey(session.ID, apiKey.ID); err != nil {
		t.Fatalf("revoke api key: %v", err)
	}

	if err := svc.DeleteSession(session.ID); err != nil {
		t.Fatalf("delete session: %v", err)
	}

	for _, eventType := range []trace.SessionEventType{
		trace.EventSessionCreated,
		trace.EventSessionUpdated,
		trace.EventAPIKeyCreated,
		trace.EventAPIKeyRevoked,
		trace.EventSessionDeleted,
	} {
		if !tracer.hasEvent(eventType) {
			t.Fatalf("expected trace event %q to be recorded", eventType)
		}
	}
}
