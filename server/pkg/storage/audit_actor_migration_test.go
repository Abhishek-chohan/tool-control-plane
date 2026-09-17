package storage

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"toolplane/pkg/model"
)

// TestAuditActorColumnAcceptsWrites runs against the migrated Postgres
// schema: audit rows carry actor_key_id. A missing column (migration not
// applied) fails the insert loudly, so this doubles as the migration check.
func TestAuditActorColumnAcceptsWrites(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("TOOLPLANE_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("TOOLPLANE_DATABASE_URL not set")
	}
	t.Setenv("TOOLPLANE_DATABASE_URL", databaseURL)
	store, err := OpenFromEnv(context.Background(), nil)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	err = store.RecordAuditEvent(context.Background(), &model.AuditEvent{
		CreatedAt:  time.Now(),
		Event:      "test_actor_column",
		SessionID:  "sess-audit-migration",
		ActorKeyID: "key-actor-migration",
		Details:    map[string]any{"probe": true},
	})
	if err != nil {
		t.Fatalf("audit insert with actor: %v (is the actor_key_id migration applied?)", err)
	}
}
