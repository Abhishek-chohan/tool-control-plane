package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestDemoFailurePathReapsProvider pins the demo's teardown contract on
// the failure path: when the provider never registers its tool, the demo
// reports the teaching error AND the provider process is reaped — an
// orphaned provider outlives the command and holds the inherited stdout
// pipe open, so a piped `toolplane demo | tail` would never finish.
//
// The fake provider sleeps instead of registering, so the demo's
// registration wait fails after its bounded deadline (~20s).
func TestDemoFailurePathReapsProvider(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script fake provider")
	}

	dir := t.TempDir()
	pidFile := filepath.Join(dir, "provider.pid")
	script := filepath.Join(dir, "fake-provider")
	body := fmt.Sprintf("#!/bin/sh\necho $$ > %s\nexec sleep 300\n", pidFile)
	// #nosec G306 -- the file is an executable fake provider; the
	// subprocess exec requires the execute bit.
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatalf("write fake provider: %v", err)
	}

	var out bytes.Buffer
	root := newRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"demo", "--provider", script})
	started := time.Now()
	_ = root.Execute()
	elapsed := time.Since(started)

	output := out.String()
	if !strings.Contains(output, "==> provider serving tools.py") {
		t.Fatalf("demo did not reach the provider phase, output:\n%s", output)
	}
	if elapsed > 40*time.Second {
		t.Fatalf("demo failure path took %v; the registration wait must be bounded", elapsed)
	}

	raw, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("fake provider never started (no pid file): %v", err)
	}
	var pid int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(raw)), "%d", &pid); err != nil {
		t.Fatalf("parse fake provider pid %q: %v", raw, err)
	}

	// The process must be gone shortly after the demo returns; Kill(0)
	// keeps succeeding for a zombie until reaped, and our teardown waits,
	// so a short poll covers scheduling.
	deadline := time.Now().Add(5 * time.Second)
	for {
		if syscall.Kill(pid, 0) == syscall.ESRCH {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("provider pid %d still alive after the demo exited — the demo must not leak its provider", pid)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
