package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestLoadgenSmokeAgentTurn drives the full embedded path — server boot,
// session setup, fake providers, agent-turn load, report — for a few
// seconds and asserts the run completed clean with real work recorded.
func TestLoadgenSmokeAgentTurn(t *testing.T) {
	if testing.Short() {
		t.Skip("load smoke skipped under -short")
	}

	reportPath := filepath.Join(t.TempDir(), "report.json")
	cfg := config{
		mode:       "embedded",
		shape:      "agent-turn",
		sessions:   2,
		duration:   4 * time.Second,
		reportPath: reportPath,
		apiKey:     embeddedAPIKey,
		burst:      3,
	}

	if code := runLoad(context.Background(), cfg); code != 0 {
		t.Fatalf("loadgen exited %d, want 0", code)
	}

	raw, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	var report loadReport
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatalf("parse report: %v", err)
	}

	ops := map[string]opReport{}
	for _, op := range report.Ops {
		ops[op.Name] = op
		if len(op.Errors) > 0 {
			t.Fatalf("op %q recorded errors: %v", op.Name, op.Errors)
		}
	}
	if invoke := ops["invoke"]; invoke.Count < cfg.sessions {
		t.Fatalf("invoke count = %d, want at least one per session (%d)", invoke.Count, cfg.sessions)
	}
	if list := ops["list-tools"]; list.Count < cfg.sessions {
		t.Fatalf("list-tools count = %d, want at least one per session (%d)", list.Count, cfg.sessions)
	}
	if resolve := ops["resolve"]; resolve.Count < invokeFloor(ops["invoke"].Count) {
		t.Fatalf("resolve count = %d, want to track completed invokes", resolve.Count)
	}
	// The claim loop is the provider heartbeat of the run: no claim polls
	// means the providers never ran.
	if claim := ops["claim-poll"]; claim.Count == 0 {
		t.Fatal("claim-poll count = 0, want the provider loop to have polled")
	}
	// Every invoke must finish inside its wait window on a healthy
	// in-memory run — nondones mean dispatch fell behind the wait budget.
	if nondone := ops["invoke-nondone"]; nondone.Count > 0 {
		t.Fatalf("invoke-nondone count = %d, want 0", nondone.Count)
	}
	// Embedded run scrapes metrics; the gRPC counter deltas must show the
	// driven traffic.
	if report.MetricsDelta != nil {
		found := false
		for key, delta := range report.MetricsDelta {
			if len(key) > len("toolplane_grpc_requests_total") &&
				key[:len("toolplane_grpc_requests_total")] == "toolplane_grpc_requests_total" && delta > 0 {
				found = true
			}
		}
		if !found {
			t.Fatal("metrics delta shows no gRPC request counters; scrape or plumbing broken")
		}
	}
}

// TestLoadgenSmokeTokenStream exercises the streaming shape: streams
// killed by the deadline window are partials, not failures.
func TestLoadgenSmokeTokenStream(t *testing.T) {
	if testing.Short() {
		t.Skip("load smoke skipped under -short")
	}

	reportPath := filepath.Join(t.TempDir(), "report.json")
	cfg := config{
		mode:        "embedded",
		shape:       "token-stream",
		sessions:    2,
		duration:    5 * time.Second,
		reportPath:  reportPath,
		apiKey:      embeddedAPIKey,
		streamCount: 2,
	}

	if code := runLoad(context.Background(), cfg); code != 0 {
		t.Fatalf("loadgen exited %d, want 0 (window-edge cancellations must be partials)", code)
	}

	raw, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	var report loadReport
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatalf("parse report: %v", err)
	}
	for _, op := range report.Ops {
		if len(op.Errors) > 0 {
			t.Fatalf("op %q recorded errors: %v", op.Name, op.Errors)
		}
	}
	if report.ChunksFollowed == 0 {
		t.Fatal("no chunks followed; the stream shape did not run")
	}
}

// invokeFloor keeps the resolve assertion readable: the provider may
// still be resolving the last invokes when the window closes.
func invokeFloor(invokes int) int {
	if invokes < 2 {
		return invokes
	}
	return invokes - 2
}
