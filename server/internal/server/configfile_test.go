package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "toolplane.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func unsetContractEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"TOOLPLANE_ENV_MODE", "TOOLPLANE_AUTH_MODE", "TOOLPLANE_AUTH_FIXED_API_KEY",
		"TOOLPLANE_AUTH_DEBUG", "TOOLPLANE_STORAGE_MODE", "TOOLPLANE_DATABASE_URL",
		"TOOLPLANE_SERVER_TLS_CERT_FILE", "TOOLPLANE_SERVER_TLS_KEY_FILE",
	} {
		t.Setenv(key, "") // t.Setenv restores the prior value afterwards
		os.Unsetenv(key)
	}
}

// TestApplyConfigFileBecomesEnvDefault pins the precedence shape: a file
// value applies only when the environment variable is unset, so
// environment keeps precedence over the file without a second resolution
// system.
func TestApplyConfigFileBecomesEnvDefault(t *testing.T) {
	unsetContractEnv(t)

	path := writeConfigFile(t, `
env: test
auth:
  mode: fixed
  fixed_api_key: file-key
`)
	settings, err := ApplyConfigFile(path)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if os.Getenv("TOOLPLANE_ENV_MODE") != "test" || os.Getenv("TOOLPLANE_AUTH_FIXED_API_KEY") != "file-key" {
		t.Fatalf("file values not applied: env=%q key=%q", os.Getenv("TOOLPLANE_ENV_MODE"), os.Getenv("TOOLPLANE_AUTH_FIXED_API_KEY"))
	}
	if settings.Provenance["TOOLPLANE_AUTH_MODE"] != "file" {
		t.Fatalf("provenance = %v, want file for auth.mode", settings.Provenance)
	}

	// Now the environment wins: a pre-set variable keeps its value.
	t.Setenv("TOOLPLANE_AUTH_FIXED_API_KEY", "env-key")
	path = writeConfigFile(t, `
auth:
  mode: fixed
  fixed_api_key: file-key
`)
	if _, err := ApplyConfigFile(path); err != nil {
		t.Fatalf("apply 2: %v", err)
	}
	if os.Getenv("TOOLPLANE_AUTH_FIXED_API_KEY") != "env-key" {
		t.Fatalf("env value overwritten by file: %q", os.Getenv("TOOLPLANE_AUTH_FIXED_API_KEY"))
	}
}

// TestApplyConfigFileRejectsUnknownKeys: a renamed key must fail loudly
// naming the key, not silently ignore a value the operator believes is
// live.
func TestApplyConfigFileRejectsUnknownKeys(t *testing.T) {
	path := writeConfigFile(t, `
auth:
  modeX: fixed
`)
	_, err := ApplyConfigFile(path)
	if err == nil {
		t.Fatal("unknown key accepted")
	}
	if !strings.Contains(err.Error(), "modeX") {
		t.Fatalf("error must name the offending key: %v", err)
	}
}

// TestApplyConfigFileExpandsVariables: ${VAR} references in string values
// expand from the environment at apply time.
func TestApplyConfigFileExpandsVariables(t *testing.T) {
	unsetContractEnv(t)
	t.Setenv("MY_DB_URL", "postgres://expanded@example:5432/db")

	path := writeConfigFile(t, `
storage:
  database_url: ${MY_DB_URL}
`)
	if _, err := ApplyConfigFile(path); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if got := os.Getenv("TOOLPLANE_DATABASE_URL"); got != "postgres://expanded@example:5432/db" {
		t.Fatalf("expansion failed: %q", got)
	}
}

// TestApplyConfigFileSecretIndirection: database_url_file reads the URL
// from a mounted secret file instead of accepting a literal.
func TestApplyConfigFileSecretIndirection(t *testing.T) {
	unsetContractEnv(t)
	secretPath := filepath.Join(t.TempDir(), "db_url")
	if err := os.WriteFile(secretPath, []byte("postgres://secret@db/toolplane\n"), 0o600); err != nil {
		t.Fatalf("write secret: %v", err)
	}

	path := writeConfigFile(t, `
storage:
  database_url_file: `+secretPath+`
`)
	settings, err := ApplyConfigFile(path)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if got := os.Getenv("TOOLPLANE_DATABASE_URL"); got != "postgres://secret@db/toolplane" {
		t.Fatalf("secret indirection failed: %q", got)
	}
	if origin, ok := settings.Provenance["TOOLPLANE_DATABASE_URL"]; !ok || !strings.HasPrefix(origin, "file:") {
		t.Fatalf("provenance = %q, want the secret file origin", origin)
	}
}

// TestApplyConfigFileSurfacesNonEnvKeys: port and metrics_listen have no
// environment variable; the entry point applies them to Options itself.
func TestApplyConfigFileSurfacesNonEnvKeys(t *testing.T) {
	unsetContractEnv(t)
	path := writeConfigFile(t, `
server:
  port: 9101
  metrics_listen: ":9102"
`)
	settings, err := ApplyConfigFile(path)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if settings.Port != 9101 || settings.MetricsListen != ":9102" {
		t.Fatalf("settings = %+v, want port 9101 and :9102", settings)
	}
}
