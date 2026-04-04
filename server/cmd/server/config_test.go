package main

import "testing"

func TestLoadServerConfigRejectsProductionInMemoryStorage(t *testing.T) {
	t.Setenv("TOOLPLANE_ENV_MODE", "production")
	t.Setenv("TOOLPLANE_AUTH_MODE", "supabase")
	t.Setenv("TOOLPLANE_SUPABASE_URL", "https://example.supabase.co")
	t.Setenv("TOOLPLANE_SUPABASE_SERVICE_KEY", "service-key")
	t.Setenv("TOOLPLANE_STORAGE_MODE", "memory")
	t.Setenv("TOOLPLANE_DATABASE_URL", "")

	if _, err := loadServerConfig(); err == nil {
		t.Fatal("expected production config with in-memory storage to fail")
	}
}

func TestLoadServerConfigAllowsProductionPostgresStorage(t *testing.T) {
	t.Setenv("TOOLPLANE_ENV_MODE", "production")
	t.Setenv("TOOLPLANE_AUTH_MODE", "supabase")
	t.Setenv("TOOLPLANE_SUPABASE_URL", "https://example.supabase.co")
	t.Setenv("TOOLPLANE_SUPABASE_SERVICE_KEY", "service-key")
	t.Setenv("TOOLPLANE_STORAGE_MODE", "postgres")
	t.Setenv("TOOLPLANE_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/toolplane?sslmode=disable")

	if _, err := loadServerConfig(); err != nil {
		t.Fatalf("expected production config with Postgres storage to pass: %v", err)
	}
}
