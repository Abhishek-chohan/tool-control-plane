package service

import (
	"reflect"
	"testing"
	"time"

	"toolplane/pkg/model"
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

	apiKey, err := svc.CreateApiKey(session.ID, "primary", "user-audit", nil)
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

func TestSessionsServiceValidateApiKeyAcceptsActiveAndRejectsRevoked(t *testing.T) {
	svc := NewSessionsService(trace.NopTracer(), nil)

	session, err := svc.CreateSession("user-audit", "Auth Session", "session auth coverage", "", "", "tenant-a")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	apiKey, err := svc.CreateApiKey(session.ID, "primary", "user-audit", nil)
	if err != nil {
		t.Fatalf("create api key: %v", err)
	}

	resolvedSessionID, err := svc.ValidateApiKey(apiKey.Key)
	if err != nil {
		t.Fatalf("validate active api key: %v", err)
	}
	if resolvedSessionID != session.ID {
		t.Fatalf("validate active api key returned session %q, want %q", resolvedSessionID, session.ID)
	}

	resolvedSessionID, err = svc.ValidateApiKey("Bearer " + apiKey.Key)
	if err != nil {
		t.Fatalf("validate bearer api key: %v", err)
	}
	if resolvedSessionID != session.ID {
		t.Fatalf("validate bearer api key returned session %q, want %q", resolvedSessionID, session.ID)
	}

	if err := svc.RevokeApiKey(session.ID, apiKey.ID); err != nil {
		t.Fatalf("revoke api key: %v", err)
	}

	if _, err := svc.ValidateApiKey(apiKey.Key); err == nil {
		t.Fatal("expected revoked api key to fail validation")
	}

	if _, err := svc.ValidateApiKey("missing-key"); err == nil {
		t.Fatal("expected unknown api key to fail validation")
	}
}

func TestSessionsServiceListAPIKeysRedactsSecretsAndPreservesCapabilities(t *testing.T) {
	svc := NewSessionsService(trace.NopTracer(), nil)

	session, err := svc.CreateSession("user-audit", "Auth Session", "session auth coverage", "", "", "tenant-a")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	apiKey, err := svc.CreateApiKey(session.ID, "reader", "user-audit", []string{"read", "execute"})
	if err != nil {
		t.Fatalf("create api key: %v", err)
	}
	if apiKey.Key == "" || apiKey.KeyPreview == "" {
		t.Fatalf("created api key = %#v, want returned secret and preview", apiKey)
	}

	listed, err := svc.ListApiKeys(session.ID)
	if err != nil {
		t.Fatalf("list api keys: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("list api keys returned %d results, want 1", len(listed))
	}
	if listed[0].Key != "" {
		t.Fatalf("listed api key secret = %q, want redacted empty string", listed[0].Key)
	}
	if listed[0].KeyPreview == "" {
		t.Fatalf("listed api key preview = %q, want non-empty preview", listed[0].KeyPreview)
	}
	if !reflect.DeepEqual(listed[0].Capabilities, []model.APIKeyCapability{model.APIKeyCapabilityRead, model.APIKeyCapabilityExecute}) {
		t.Fatalf("listed api key capabilities = %v, want [read execute]", listed[0].Capabilities)
	}
	if listed[0].ID != apiKey.ID {
		t.Fatalf("listed api key id = %q, want %q", listed[0].ID, apiKey.ID)
	}
}

func TestSessionsServiceAuthenticateAPIKeyReturnsPrincipal(t *testing.T) {
	svc := NewSessionsService(trace.NopTracer(), nil)

	session, err := svc.CreateSession("user-audit", "Auth Session", "session auth coverage", "", "", "tenant-a")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	apiKey, err := svc.CreateApiKey(session.ID, "runner", "user-audit", []string{"execute"})
	if err != nil {
		t.Fatalf("create api key: %v", err)
	}

	principal, err := svc.AuthenticateAPIKey("Bearer " + apiKey.Key)
	if err != nil {
		t.Fatalf("authenticate api key: %v", err)
	}
	if principal.SessionID != session.ID || principal.UserID != "user-audit" || principal.KeyID != apiKey.ID {
		t.Fatalf("principal = %#v, want session=%s user=user-audit key=%s", principal, session.ID, apiKey.ID)
	}
	if !principal.HasCapability(model.APIKeyCapabilityExecute) {
		t.Fatalf("principal capabilities = %v, want execute capability", principal.Capabilities)
	}
}

func TestSessionsServiceListUserSessionsReturnsNewestSessionsFirst(t *testing.T) {
	svc := NewSessionsService(trace.NopTracer(), nil)

	oldest, err := svc.CreateSession("user-order", "oldest", "", "", "", "tenant-a")
	if err != nil {
		t.Fatalf("create oldest session: %v", err)
	}
	middle, err := svc.CreateSession("user-order", "middle", "", "", "", "tenant-a")
	if err != nil {
		t.Fatalf("create middle session: %v", err)
	}
	newest, err := svc.CreateSession("user-order", "newest", "", "", "", "tenant-a")
	if err != nil {
		t.Fatalf("create newest session: %v", err)
	}

	base := time.Date(2026, time.April, 6, 12, 0, 0, 0, time.UTC)
	oldest.CreatedAt = base
	middle.CreatedAt = base.Add(1 * time.Minute)
	newest.CreatedAt = base.Add(2 * time.Minute)

	pageZero, totalCount, err := svc.ListUserSessions("user-order", 2, 0, "")
	if err != nil {
		t.Fatalf("list user sessions page 0: %v", err)
	}
	if totalCount != 3 {
		t.Fatalf("page 0 totalCount = %d, want 3", totalCount)
	}
	if got := []string{pageZero[0].ID, pageZero[1].ID}; !reflect.DeepEqual(got, []string{newest.ID, middle.ID}) {
		t.Fatalf("page 0 ids = %v, want [%s %s]", got, newest.ID, middle.ID)
	}

	pageOne, totalCount, err := svc.ListUserSessions("user-order", 2, 1, "")
	if err != nil {
		t.Fatalf("list user sessions page 1: %v", err)
	}
	if totalCount != 3 {
		t.Fatalf("page 1 totalCount = %d, want 3", totalCount)
	}
	if len(pageOne) != 1 || pageOne[0].ID != oldest.ID {
		t.Fatalf("page 1 ids = %v, want [%s]", sessionIDs(pageOne), oldest.ID)
	}
}

func sessionIDs(sessions []*model.Session) []string {
	ids := make([]string, 0, len(sessions))
	for _, session := range sessions {
		ids = append(ids, session.ID)
	}
	return ids
}
