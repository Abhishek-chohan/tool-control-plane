package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"toolplane/internal/cli"
)

// portOfAddress extracts the numeric port from a host:port address.
func portOfAddress(address string) string {
	parts := strings.Split(address, ":")
	return parts[len(parts)-1]
}

// errorsAs unwraps a command's exitError.
func errorsAs(err error) (exitError, bool) {
	var coded exitError
	if errors.As(err, &coded) {
		return coded, true
	}
	return exitError{}, false
}

// TestInitScaffoldsWorkingDirectory pins the newcomer scaffold: three
// files written once, refuse-to-overwrite on a second run, and a tools
// file whose content mirrors the documented bare-@tool module.
func TestInitScaffoldsWorkingDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "proj")
	var out bytes.Buffer
	root := newRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"init", "--dir", dir})
	if err := root.Execute(); err != nil {
		coded, ok := errorsAs(err)
		if !ok || coded.code != cli.ExitOK {
			t.Fatalf("init: %v", err)
		}
	}

	for _, name := range []string{"tools.py", "toolplane.yaml", ".env"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("scaffold missing %s: %v", name, err)
		}
	}
	toolsPy, _ := os.ReadFile(filepath.Join(dir, "tools.py"))
	for _, fragment := range []string{"from toolplane.provider_registry import tool", "@tool(name=\"add\""} {
		if !bytes.Contains(toolsPy, []byte(fragment)) {
			t.Fatalf("tools.py missing %q:\n%s", fragment, toolsPy)
		}
	}
	env, _ := os.ReadFile(filepath.Join(dir, ".env"))
	if !bytes.Contains(env, []byte("TOOLPLANE_AUTH_FIXED_API_KEY=dev-key")) {
		t.Fatalf(".env missing the dev key: %s", env)
	}

	// Second run refuses to overwrite: exit invalid-argument, files intact.
	var errOut bytes.Buffer
	root2 := newRootCommand("test")
	root2.SetOut(&errOut)
	root2.SetErr(&errOut)
	root2.SetArgs([]string{"init", "--dir", dir})
	err := root2.Execute()
	coded, ok := errorsAs(err)
	if !ok || coded.code != cli.ExitInvalidArgument {
		t.Fatalf("second init err=%v, want exitError(invalid-argument)", err)
	}
	if !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("refusal message missing: %s", err.Error())
	}
	// The scaffold survives untouched (the add template still reads the same).
	toolsPy2, _ := os.ReadFile(filepath.Join(dir, "tools.py"))
	if !bytes.Equal(toolsPy, toolsPy2) {
		t.Fatal("second init modified tools.py")
	}
}

// TestWaitReadyAgainstLiveServer: the happy path — wait exits 0 once the
// control plane answers, well inside the deadline.
func TestWaitReadyAgainstLiveServer(t *testing.T) {
	backend := bootTestServer(t)
	port := portOfAddress(backend)
	conn := cli.NewConnection("localhost:"+port, "dev-key", 5*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := cli.WaitReady(ctx, conn, 100*time.Millisecond); err != nil {
		t.Fatalf("WaitReady against a live server: %v", err)
	}
}

// TestWaitReadyTimesOutWithTeaching: against a dead port, WaitReady
// surfaces the last probe error so the verb can print the
// toolplane-serve next step.
func TestWaitReadyTimesOutWithTeaching(t *testing.T) {
	port, err := pickFreePort()
	if err != nil {
		t.Fatalf("pick port: %v", err)
	}
	conn := cli.NewConnection(fmt.Sprintf("localhost:%d", port), "", time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
	defer cancel()
	err = cli.WaitReady(ctx, conn, 100*time.Millisecond)
	if err == nil {
		t.Fatal("WaitReady against a dead port returned nil")
	}
	taught := cli.TeachError(err, conn)
	if !strings.Contains(taught, "toolplane serve") {
		t.Fatalf("teaching message missing the serve hint: %s", taught)
	}
}

// TestInvokeUnavailableTeachesServe pins the invoke-side error-UX
// convention: an unreachable server prints the toolplane-serve next step.
func TestInvokeUnavailableTeachesServe(t *testing.T) {
	port, err := pickFreePort()
	if err != nil {
		t.Fatalf("pick port: %v", err)
	}
	conn := cli.NewConnection(fmt.Sprintf("localhost:%d", port), "dev-key", time.Second)
	var out bytes.Buffer
	code, invokeErr := runInvoke(context.Background(), conn, &out, invokeOptions{
		Tool:    "add",
		Session: "sess-x",
		Input:   `{"a":1}`,
		Wait:    0,
		Format:  cli.FormatTable,
	})
	if code != cli.ExitUnavailable {
		t.Fatalf("invoke exit code = %d, want ExitUnavailable", code)
	}
	taught := cli.TeachError(invokeErr, conn)
	if !strings.Contains(taught, "toolplane serve") {
		t.Fatalf("invoke error must teach the serve command: %s", taught)
	}
}
