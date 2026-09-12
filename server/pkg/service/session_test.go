package service

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/trace"
)

func TestSessionsServiceRecordsAuditEvents(t *testing.T) {
	tracer := &recordingTracer{}
	svc := NewSessionsService(tracer, nil)

	session, err := svc.CreateSession("user-audit", "Audit Session", "session for trace coverage", "", "tenant-a")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if _, err := svc.UpdateSession(session.ID, "Audit Session Updated", "updated", "tenant-b"); err != nil {
		t.Fatalf("update session: %v", err)
	}

	apiKey, err := svc.CreateApiKey(session.ID, "primary", "user-audit", []string{"read", "execute", "admin"})
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

	session, err := svc.CreateSession("user-audit", "Auth Session", "session auth coverage", "", "tenant-a")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	apiKey, err := svc.CreateApiKey(session.ID, "primary", "user-audit", []string{"read", "execute", "admin"})
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

	session, err := svc.CreateSession("user-audit", "Auth Session", "session auth coverage", "", "tenant-a")
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
	// The key was minted with ["read", "execute"]; execute is the legacy
	// pre-split alias and normalizes to invoke+provide on storage/normalization.
	if !reflect.DeepEqual(listed[0].Capabilities, []model.APIKeyCapability{model.APIKeyCapabilityRead, model.APIKeyCapabilityInvoke, model.APIKeyCapabilityProvide}) {
		t.Fatalf("listed api key capabilities = %v, want [read invoke provide]", listed[0].Capabilities)
	}
	if listed[0].ID != apiKey.ID {
		t.Fatalf("listed api key id = %q, want %q", listed[0].ID, apiKey.ID)
	}
}

func TestSessionsServiceAuthenticateAPIKeyReturnsPrincipal(t *testing.T) {
	svc := NewSessionsService(trace.NopTracer(), nil)

	session, err := svc.CreateSession("user-audit", "Auth Session", "session auth coverage", "", "tenant-a")
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
	// The key's legacy execute capability expands to invoke+provide.
	if !principal.HasCapability(model.APIKeyCapabilityInvoke) || !principal.HasCapability(model.APIKeyCapabilityProvide) {
		t.Fatalf("principal capabilities = %v, want invoke+provide (legacy execute)", principal.Capabilities)
	}
}

func TestSessionsServiceListUserSessionsReturnsNewestSessionsFirst(t *testing.T) {
	svc := NewSessionsService(trace.NopTracer(), nil)

	oldest, err := svc.CreateSession("user-order", "oldest", "", "", "tenant-a")
	if err != nil {
		t.Fatalf("create oldest session: %v", err)
	}
	middle, err := svc.CreateSession("user-order", "middle", "", "", "tenant-a")
	if err != nil {
		t.Fatalf("create middle session: %v", err)
	}
	newest, err := svc.CreateSession("user-order", "newest", "", "", "tenant-a")
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

	// The third argument is the item offset the opaque v1 cursor decodes
	// to: page two of a size-2 listing starts at item 2.
	pageOne, totalCount, err := svc.ListUserSessions("user-order", 2, 2, "")
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

func TestSessionsServiceCreateApiKeyRequiresExplicitCapabilities(t *testing.T) {
	svc := NewSessionsService(trace.NopTracer(), nil)
	session, err := svc.CreateSession("user-caps", "Caps Session", "", "", "")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if _, err := svc.CreateApiKey(session.ID, "no-caps", "user-caps", nil); err == nil {
		t.Fatal("CreateApiKey with no capabilities should fail")
	}
	if _, err := svc.CreateApiKey(session.ID, "blank-caps", "user-caps", []string{"", "  "}); err == nil {
		t.Fatal("CreateApiKey with only blank capabilities should fail")
	}
	if _, err := svc.CreateApiKey(session.ID, "reader", "user-caps", []string{"read"}); err != nil {
		t.Fatalf("CreateApiKey with explicit capabilities: %v", err)
	}
}

func TestSessionsServiceApiKeySecretDoesNotEmbedSessionID(t *testing.T) {
	svc := NewSessionsService(trace.NopTracer(), nil)
	session, err := svc.CreateSession("user-fmt", "Format Session", "", "", "")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	apiKey, err := svc.CreateApiKey(session.ID, "fmt", "user-fmt", []string{"read"})
	if err != nil {
		t.Fatalf("create api key: %v", err)
	}
	if strings.Contains(apiKey.Key, session.ID) {
		t.Fatalf("api key secret embeds the session ID: %q", apiKey.KeyPreview)
	}
	if !strings.HasPrefix(apiKey.Key, "toolplane_key_") {
		t.Fatalf("api key secret = %q, want toolplane_key_ prefix", apiKey.Key)
	}
}

func TestSessionsServiceInvalidateSessionRevokesEveryLiveKey(t *testing.T) {
	svc := NewSessionsService(trace.NopTracer(), nil)
	session, err := svc.CreateSession("user-inval", "Invalidate Session", "", "", "")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	first, err := svc.CreateApiKey(session.ID, "first", "user-inval", []string{"read", "execute"})
	if err != nil {
		t.Fatalf("create first key: %v", err)
	}
	second, err := svc.CreateApiKey(session.ID, "second", "user-inval", []string{"admin"})
	if err != nil {
		t.Fatalf("create second key: %v", err)
	}

	// Sanity: both keys authenticate before invalidation.
	if _, err := svc.AuthenticateAPIKey(first.Key); err != nil {
		t.Fatalf("authenticate first before invalidation: %v", err)
	}
	if _, err := svc.AuthenticateAPIKey(second.Key); err != nil {
		t.Fatalf("authenticate second before invalidation: %v", err)
	}

	revoked, err := svc.InvalidateSession(session.ID, "suspected compromise")
	if err != nil {
		t.Fatalf("invalidate session: %v", err)
	}
	if revoked != 2 {
		t.Fatalf("revoked count = %d, want 2", revoked)
	}

	if _, err := svc.AuthenticateAPIKey(first.Key); err == nil {
		t.Fatal("first key still authenticates after session invalidation")
	}
	if _, err := svc.AuthenticateAPIKey(second.Key); err == nil {
		t.Fatal("second key still authenticates after session invalidation")
	}

	// Invalidating again is a no-op, not an error.
	revokedAgain, err := svc.InvalidateSession(session.ID, "already done")
	if err != nil {
		t.Fatalf("second invalidation: %v", err)
	}
	if revokedAgain != 0 {
		t.Fatalf("second invalidation revoked = %d, want 0", revokedAgain)
	}
}
