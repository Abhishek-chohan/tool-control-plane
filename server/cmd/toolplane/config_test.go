package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"toolplane/internal/cli"
)

// TestServeDryRunPrintsProvenanceAndExits pins the --dry-run contract:
// the resolved configuration prints with per-key provenance and the
// command exits 0 without serving. Secrets render as set, never in
// cleartext.
func TestServeDryRunPrintsProvenanceAndExits(t *testing.T) {
	t.Setenv("TOOLPLANE_ENV_MODE", "test")
	t.Setenv("TOOLPLANE_AUTH_MODE", "fixed")
	t.Setenv("TOOLPLANE_AUTH_FIXED_API_KEY", "dev-key")
	t.Setenv("TOOLPLANE_STORAGE_MODE", "memory")

	configPath := filepath.Join(t.TempDir(), "toolplane.yaml")
	if err := os.WriteFile(configPath, []byte("env: test\nserver:\n  port: 9101\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var out bytes.Buffer
	root := newRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"serve", "--config", configPath, "--dry-run"})
	err := root.Execute()
	var coded exitError
	if !errors.As(err, &coded) || coded.code != cli.ExitOK {
		t.Fatalf("serve --dry-run: err=%v, want exitError(0)", err)
	}
	output := out.String()
	for _, fragment := range []string{
		"resolved configuration:",
		"9101",
		"(file)", // port came from the config file
		"(env)",  // env/auth came from the environment
		"auth.fixed_api_key:",
	} {
		if !bytes.Contains([]byte(output), []byte(fragment)) {
			t.Fatalf("dry-run output missing %q:\n%s", fragment, output)
		}
	}
	if bytes.Contains([]byte(output), []byte("dev-key")) {
		t.Fatal("dry-run leaked the fixed API key; secrets render as set-hidden")
	}
}

// TestServeConfigFileFeedsBoot: the config file alone (environment
// unset) configures a valid boot — the dry-run resolves auth and storage
// from the file with file provenance.
func TestServeConfigFileFeedsBoot(t *testing.T) {
	for _, key := range []string{
		"TOOLPLANE_ENV_MODE", "TOOLPLANE_AUTH_MODE", "TOOLPLANE_AUTH_FIXED_API_KEY",
		"TOOLPLANE_STORAGE_MODE", "TOOLPLANE_DATABASE_URL",
	} {
		t.Setenv(key, "")
		os.Unsetenv(key)
	}

	configPath := filepath.Join(t.TempDir(), "toolplane.yaml")
	if err := os.WriteFile(configPath, []byte(`env: test
auth:
  mode: fixed
  fixed_api_key: file-key
storage:
  mode: memory
`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var out bytes.Buffer
	root := newRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"serve", "--config", configPath, "--dry-run"})
	err := root.Execute()
	var coded exitError
	if !errors.As(err, &coded) || coded.code != cli.ExitOK {
		t.Fatalf("serve --config --dry-run: err=%v, want exitError(0)", err)
	}
	for _, fragment := range []string{"auth.mode:", "storage.mode:", "(file)"} {
		if !bytes.Contains(out.Bytes(), []byte(fragment)) {
			t.Fatalf("missing %q in:\n%s", fragment, out.String())
		}
	}
}
