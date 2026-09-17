package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"toolplane/pkg/service"
)

// TestStatusCommandAgainstLiveServer: status reports reachability,
// version, and the resolved storage mode against a live control plane.
func TestStatusCommandAgainstLiveServer(t *testing.T) {
	address, cleanup := bootAdminE2EServer(t)
	defer cleanup()

	var out bytes.Buffer
	root := newRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"status", "--address", address, "--api-key", "dev-key"})
	if err := root.Execute(); err != nil {
		t.Fatalf("status: %v", err)
	}
	output := out.String()
	for _, fragment := range []string{"reachable at " + address, "version:", "storage:"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("status output missing %q:\n%s", fragment, output)
		}
	}
}

// TestDoctorPassesOnHealthySetup: with the dev environment contract
// satisfied, doctor reports config, connectivity, and version checks as
// passing.
func TestDoctorPassesOnHealthySetup(t *testing.T) {
	version = "test"
	address, cleanup := bootAdminE2EServer(t)
	defer cleanup()
	// Client and server share the process-global build identity, so the
	// skew check passes on a healthy setup.
	service.BuildVersion = "test"

	var out bytes.Buffer
	root := newRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doctor", "--address", address, "--api-key", "dev-key"})
	if err := root.Execute(); err != nil {
		t.Fatalf("doctor failed on a healthy setup: %v\nOUTPUT:\n%s", err, out.String())
	}
	output := out.String()
	for _, fragment := range []string{"[ok  ] config", "[ok  ] connectivity", "[ok  ] version"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("doctor output missing %q:\n%s", fragment, output)
		}
	}
}

// TestDoctorNamesMissingConfig: a doctor run without the environment
// contract fails its config check and names the missing variable.
func TestDoctorNamesMissingConfig(t *testing.T) {
	for _, key := range []string{"TOOLPLANE_AUTH_MODE", "TOOLPLANE_AUTH_FIXED_API_KEY", "TOOLPLANE_STORAGE_MODE"} {
		t.Setenv(key, "")
		os.Unsetenv(key)
	}

	var out bytes.Buffer
	root := newRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doctor", "--address", "localhost:1", "--timeout", "1s"})
	err := root.Execute()
	if err == nil {
		t.Fatal("doctor with missing config passed; it must fail loudly")
	}
	if !strings.Contains(err.Error(), "TOOLPLANE_AUTH_MODE") && !strings.Contains(out.String(), "TOOLPLANE_AUTH_MODE") {
		t.Fatalf("doctor must name the missing variable (got: %s | %s)", err.Error(), out.String())
	}
}
