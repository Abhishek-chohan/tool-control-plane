package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
	"toolplane/pkg/trace"
)

// Service-layer benchmarks over the full stack (validation, cache,
// store) — the paths a request takes in production. Memory mode always
// runs; Postgres mode runs when TOOLPLANE_DATABASE_URL is set
// (make bench / make bench-postgres). Numbers are hardware- and
// store-relative: run them on the hardware that will serve the load
// before reading anything into them.

func benchStack(b *testing.B, store storage.Storer) (*SessionsService, *MachinesService, *RequestsService, *ToolService) {
	b.Helper()
	ctx := context.Background()
	tracer := trace.NopTracer()
	sessionService := NewSessionsService(tracer, store)
	toolService := NewToolService(tracer, store)
	machineService := NewMachinesService(ctx, toolService, tracer, store)
	requestService := NewRequestsService(ctx, toolService, machineService, tracer, store)
	return sessionService, machineService, requestService, toolService
}

// benchSeedSession creates the session row Postgres-mode machines and
// requests foreign-key onto, then registers a machine with one echo tool.
func benchSeedSession(b *testing.B, sessionService *SessionsService, machineService *MachinesService, sessionID string) string {
	b.Helper()
	if _, err := sessionService.CreateSession("bench-user", sessionID, "bench", sessionID, ""); err != nil {
		b.Fatalf("create session: %v", err)
	}
	machineID := "machine-" + sessionID
	if _, err := machineService.RegisterMachine(sessionID, machineID, "bench", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo", `{}`, nil, nil),
	}, ""); err != nil {
		b.Fatalf("register machine: %v", err)
	}
	return machineID
}

func runBenchAgainstBothStores(b *testing.B, fn func(b *testing.B, store storage.Storer)) {
	b.Run("memory", func(b *testing.B) { fn(b, nil) })
	b.Run("postgres", func(b *testing.B) {
		if strings.TrimSpace(os.Getenv("TOOLPLANE_DATABASE_URL")) == "" {
			b.Skip("TOOLPLANE_DATABASE_URL not set; skipping Postgres benchmark")
		}
		store, err := storage.OpenFromEnv(context.Background(), log.New(io.Discard, "", 0))
		if err != nil {
			b.Fatalf("open postgres store: %v", err)
		}
		b.Cleanup(func() { _ = store.Close() })
		fn(b, store)
	})
}

var benchSessionSeq atomic.Uint64

// benchSessionName yields a unique session id per invocation: benchmarks
// re-run with growing b.N and the Postgres path shares one database.
func benchSessionName(kind string) string {
	return fmt.Sprintf("bench-%s-%d-%d", kind, time.Now().UnixNano()%1_000_000_000, benchSessionSeq.Add(1))
}

// BenchmarkRequestLifecycle measures the sequential agent loop — create,
// claim, resolve — which keeps the per-session backlog at one pending
// request, so the number isolates the fixed per-request cost rather than
// backlog behavior.
func BenchmarkRequestLifecycle(b *testing.B) {
	runBenchAgainstBothStores(b, func(b *testing.B, store storage.Storer) {
		sessionService, machineService, requestService, _ := benchStack(b, store)
		sessionID := benchSessionName("lifecycle")
		machineID := benchSeedSession(b, sessionService, machineService, sessionID)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			request, err := requestService.CreateRequest(sessionID, "echo", `{}`, 0, "")
			if err != nil {
				b.Fatalf("create: %v", err)
			}
			claimed, err := requestService.ClaimRequest(sessionID, request.ID, machineID)
			if err != nil {
				b.Fatalf("claim: %v", err)
			}
			if err := requestService.SubmitRequestResult(sessionID, request.ID, claimed.LeasedBy, claimed.LeaseEpoch, map[string]string{"i": fmt.Sprintf("%d", i)}, model.ResultTypeResolution, nil); err != nil {
				b.Fatalf("submit: %v", err)
			}
		}
	})
}

// BenchmarkGetRequestByID measures the store-first status read — the
// work every poll performs (the 250ms waiter loop and every client
// GetRequest).
func BenchmarkGetRequestByID(b *testing.B) {
	runBenchAgainstBothStores(b, func(b *testing.B, store storage.Storer) {
		sessionService, machineService, requestService, _ := benchStack(b, store)
		sessionID := benchSessionName("get")
		machineID := benchSeedSession(b, sessionService, machineService, sessionID)

		request, err := requestService.CreateRequest(sessionID, "echo", `{}`, 0, "")
		if err != nil {
			b.Fatalf("create: %v", err)
		}
		claimed, err := requestService.ClaimRequest(sessionID, request.ID, machineID)
		if err != nil {
			b.Fatalf("claim: %v", err)
		}
		if err := requestService.SubmitRequestResult(sessionID, request.ID, claimed.LeasedBy, claimed.LeaseEpoch, map[string]string{"done": "true"}, model.ResultTypeResolution, nil); err != nil {
			b.Fatalf("submit: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := requestService.GetRequestByID(sessionID, request.ID); err != nil {
				b.Fatalf("get: %v", err)
			}
		}
	})
}

// BenchmarkListTools measures session-scoped discovery with 100
// registered tools — the store-first read plus cache mirror that every
// agent turn's discovery call performs.
func BenchmarkListTools(b *testing.B) {
	runBenchAgainstBothStores(b, func(b *testing.B, store storage.Storer) {
		sessionService, machineService, _, toolService := benchStack(b, store)
		sessionID := benchSessionName("list")
		machineID := benchSeedSession(b, sessionService, machineService, sessionID)
		for i := 0; i < 100; i++ {
			if _, err := toolService.RegisterTool(sessionID, machineID, fmt.Sprintf("tool-%03d", i), "bench tool", "{}", nil, nil); err != nil {
				b.Fatalf("register tool %d: %v", i, err)
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := toolService.ListTools(sessionID); err != nil {
				b.Fatalf("list tools: %v", err)
			}
		}
	})
}
