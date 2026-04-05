package main

import "testing"

func TestLoadServerConfigRejectsProductionInMemoryStorage(t *testing.T) {
	t.Setenv("TOOLPLANE_ENV_MODE", "production")
	t.Setenv("TOOLPLANE_AUTH_MODE", "postgres")
	t.Setenv("TOOLPLANE_STORAGE_MODE", "memory")
	t.Setenv("TOOLPLANE_DATABASE_URL", "")

	if _, err := loadServerConfig(); err == nil {
		t.Fatal("expected production config with in-memory storage to fail")
	}
}

func TestLoadServerConfigAllowsProductionPostgresStorage(t *testing.T) {
	t.Setenv("TOOLPLANE_ENV_MODE", "production")
	t.Setenv("TOOLPLANE_AUTH_MODE", "postgres")
	t.Setenv("TOOLPLANE_STORAGE_MODE", "postgres")
	t.Setenv("TOOLPLANE_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/toolplane?sslmode=disable")

	if _, err := loadServerConfig(); err != nil {
		t.Fatalf("expected production config with Postgres storage to pass: %v", err)
	}
}

func TestLoadServerConfigAllowsProductionPostgresAuth(t *testing.T) {
	t.Setenv("TOOLPLANE_ENV_MODE", "production")
	t.Setenv("TOOLPLANE_AUTH_MODE", "postgres")
	t.Setenv("TOOLPLANE_STORAGE_MODE", "")
	t.Setenv("TOOLPLANE_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/toolplane?sslmode=disable")

	if _, err := loadServerConfig(); err != nil {
		t.Fatalf("expected production config with Postgres auth and storage to pass: %v", err)
	}
}

func TestLoadServerConfigRejectsPostgresAuthWithoutPostgresStorage(t *testing.T) {
	t.Setenv("TOOLPLANE_ENV_MODE", "development")
	t.Setenv("TOOLPLANE_AUTH_MODE", "postgres")
	t.Setenv("TOOLPLANE_STORAGE_MODE", "memory")
	t.Setenv("TOOLPLANE_DATABASE_URL", "")

	if _, err := loadServerConfig(); err == nil {
		t.Fatal("expected postgres auth without postgres storage to fail")
	}
}
