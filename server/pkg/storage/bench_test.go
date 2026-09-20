package storage_test

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
	"toolplane/pkg/storage/memory"
)

// Benchmarks for the store's hot paths. They run against the in-memory
// store unconditionally and against Postgres when TOOLPLANE_DATABASE_URL
// is set (make bench / make bench-postgres). Numbers are hardware- and
// store-relative: run them on the serving hardware before reading
// anything into them.

var benchSeq atomic.Uint64

func benchUID(b *testing.B, kind string) string {
	b.Helper()
	n := benchSeq.Add(1)
	return fmt.Sprintf("bench_%s_%d_%d", kind, time.Now().UnixNano()%1_000_000, n)
}

func benchSeedSession(b *testing.B, s storage.Storer) string {
	b.Helper()
	sess := model.NewSession("bench", "bench", "bench-user", "")
	if err := s.SaveSession(context.Background(), sess); err != nil {
		b.Fatalf("seed session: %v", err)
	}
	return sess.ID
}

func benchSeedClaimedStream(b *testing.B, s storage.Storer, windowChunks int) (sessionID, requestID, machineID string, leaseEpoch int64) {
	b.Helper()
	ctx := context.Background()

	sessionID = benchUID(b, "sess")
	sess := model.NewSession("bench", "bench", "bench-user", "")
	sess.ID = sessionID
	if err := s.SaveSession(ctx, sess); err != nil {
		b.Fatalf("seed session: %v", err)
	}

	machineID = "mach-" + benchUID(b, "mach")
	m := model.NewMachine(sessionID, machineID, "bench", "go", "127.0.0.1")
	m.LastPingAt = time.Now()
	m.CreatedAt = time.Now()
	if err := s.SaveMachine(ctx, m); err != nil {
		b.Fatalf("seed machine: %v", err)
	}

	tool := model.NewTool(sessionID, machineID, "bench-stream", "bench", "{}", nil, nil)
	if err := s.SaveTool(ctx, tool); err != nil {
		b.Fatalf("seed tool: %v", err)
	}

	req := model.NewRequest(sessionID, "bench-stream", `{}`)
	req.VisibleAt = time.Now()
	if err := s.SaveRequest(ctx, req); err != nil {
		b.Fatalf("seed request: %v", err)
	}
	requestID = req.ID

	claimed, ok, err := s.ClaimRequest(ctx, sessionID, requestID, machineID, time.Minute)
	if err != nil || !ok {
		b.Fatalf("claim: ok=%v err=%v", ok, err)
	}

	// Fill the retained window so append benches measure steady-state cost
	// (row lock + inserts + eager trim + window re-read) rather than the
	// cheap first-chunks case.
	chunk := strings.Repeat("x", 1024)
	appended := 0
	for appended < windowChunks {
		take := windowChunks - appended
		if take > 32 {
			take = 32
		}
		batch := make([]string, take)
		for i := range batch {
			batch[i] = chunk
		}
		claimed, err = s.AppendRequestChunksFenced(ctx, sessionID, requestID, machineID, claimed.LeaseEpoch, batch)
		if err != nil {
			b.Fatalf("prefill chunks: %v", err)
		}
		appended += take
	}
	return sessionID, requestID, machineID, claimed.LeaseEpoch
}

// BenchmarkInsertRequest measures the durable create: the SERIALIZABLE
// insert transaction with the per-session pending-cap count inside it.
// Sessions rotate every 256 inserts so the cap never rejects; the
// rotation cost is amortized (1 SaveSession per 256 inserts).
func BenchmarkInsertRequest(b *testing.B) {
	b.Run("memory", func(b *testing.B) {
		s := memory.New()
		defer func() { _ = s.Close() }()
		runInsertRequestBench(b, s)
	})
	b.Run("postgres", func(b *testing.B) {
		s := benchPostgresStore(b)
		runInsertRequestBench(b, s)
	})
}

func runInsertRequestBench(b *testing.B, s storage.Storer) {
	ctx := context.Background()
	const perSession = 256
	sessionID := ""
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%perSession == 0 {
			b.StopTimer()
			sessionID = benchSeedSession(b, s)
			b.StartTimer()
		}
		if err := s.InsertRequest(ctx, model.NewRequest(sessionID, "bench-tool", `{}`)); err != nil {
			b.Fatalf("insert request: %v", err)
		}
	}
}

// BenchmarkAppendRequestChunksFenced measures a provider streaming batch
// (8 x 1KiB) into a full retained window — the worst-case steady-state
// append: row lock, batch inserts, eager window trim, and the retained
// window re-read, all inside one SERIALIZABLE transaction per call.
func BenchmarkAppendRequestChunksFenced(b *testing.B) {
	b.Run("memory", func(b *testing.B) {
		s := memory.New()
		defer func() { _ = s.Close() }()
		runAppendChunksBench(b, s)
	})
	b.Run("postgres", func(b *testing.B) {
		s := benchPostgresStore(b)
		runAppendChunksBench(b, s)
	})
}

func runAppendChunksBench(b *testing.B, s storage.Storer) {
	ctx := context.Background()
	sessionID, requestID, machineID, epoch := benchSeedClaimedStream(b, s, 100)
	batch := make([]string, 8)
	for i := range batch {
		batch[i] = strings.Repeat("x", 1024)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.AppendRequestChunksFenced(ctx, sessionID, requestID, machineID, epoch, batch); err != nil {
			b.Fatalf("append chunks: %v", err)
		}
	}
}

// BenchmarkGetRequestChunksByRequest measures the replay read: the full
// retained window (100 x 1KiB) fetched and validated for gap-free
// contiguity — the work behind every ResumeStream catch-up.
func BenchmarkGetRequestChunksByRequest(b *testing.B) {
	b.Run("memory", func(b *testing.B) {
		s := memory.New()
		defer func() { _ = s.Close() }()
		runReplayBench(b, s)
	})
	b.Run("postgres", func(b *testing.B) {
		s := benchPostgresStore(b)
		runReplayBench(b, s)
	})
}

func runReplayBench(b *testing.B, s storage.Storer) {
	ctx := context.Background()
	_, requestID, _, _ := benchSeedClaimedStream(b, s, 100)
	stored, err := s.GetRequest(ctx, requestID)
	if err != nil {
		b.Fatalf("load request bounds: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.GetRequestChunksByRequest(ctx, requestID, stored.StreamStartSeq, stored.NextStreamSeq); err != nil {
			b.Fatalf("read chunk window: %v", err)
		}
	}
}

func benchPostgresStore(b *testing.B) *storage.Store {
	b.Helper()
	if strings.TrimSpace(os.Getenv("TOOLPLANE_DATABASE_URL")) == "" {
		b.Skip("TOOLPLANE_DATABASE_URL not set; skipping Postgres benchmark")
	}
	store, err := storage.OpenFromEnv(context.Background(), log.New(io.Discard, "", 0))
	if err != nil {
		b.Fatalf("open postgres store: %v", err)
	}
	b.Cleanup(func() { _ = store.Close() })
	return store
}
