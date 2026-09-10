package storage_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
)

// TestAppendRequestChunksFenced_TableBackedWindow proves the chunk contract
// end to end against both stores: appends assign ascending sequences, the
// window read returns exactly the retained chunks for the advertised
// bookkeeping, and a stale lease grant is rejected before anything is
// written.
func TestAppendRequestChunksFenced_TableBackedWindow(t *testing.T) {
	runAgainstBoth(t, "chunk table window", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		sess := "sess-" + uid(t)
		mach := "mach-" + uid(t)
		reqID := "req-" + uid(t)
		tool := "tool-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, mach, time.Now())
		seedPendingRequest(t, s, sess, reqID, tool)

		// Claim so the append's lease fence accepts the write.
		claimed, err := s.LeasePendingRequest(ctx, sess, mach, []string{tool}, 30*time.Second)
		if err != nil || claimed == nil {
			t.Fatalf("claim for append: v=%v err=%v", claimed, err)
		}

		updated, err := s.AppendRequestChunksFenced(ctx, sess, reqID, mach, claimed.LeaseEpoch, []string{"one", "two"})
		if err != nil {
			t.Fatalf("append: %v", err)
		}
		if updated.NextStreamSeq != 3 {
			t.Fatalf("next seq = %d, want 3", updated.NextStreamSeq)
		}

		window, err := s.GetRequestChunksByRequest(ctx, reqID, updated.StreamStartSeq, updated.NextStreamSeq)
		if err != nil {
			t.Fatalf("window read: %v", err)
		}
		if len(window.Chunks) != 2 || window.Chunks[0] != "one" || window.Chunks[1] != "two" {
			t.Fatalf("window = %v, want [one two]", window.Chunks)
		}

		// A second append continues the sequence.
		if _, err := s.AppendRequestChunksFenced(ctx, sess, reqID, mach, claimed.LeaseEpoch, []string{"three"}); err != nil {
			t.Fatalf("second append: %v", err)
		}
		window, err = s.GetRequestChunksByRequest(ctx, reqID, 1, 4)
		if err != nil {
			t.Fatalf("window read 2: %v", err)
		}
		if len(window.Chunks) != 3 || window.Chunks[2] != "three" {
			t.Fatalf("window = %v, want [one two three]", window.Chunks)
		}

		// A stale lease grant is rejected.
		if _, err := s.AppendRequestChunksFenced(ctx, sess, reqID, mach, claimed.LeaseEpoch-1, []string{"forged"}); err == nil {
			t.Fatal("forged epoch accepted")
		}
	})
}

// TestAddStreamChunkByteWindowTrim exercises the model-level byte bound.
func TestAddStreamChunkByteWindowTrim(t *testing.T) {
	large := strings.Repeat("x", 5<<20) // 5 MiB

	req := model.NewRequest("sess-chunk-trim", "echo", `{}`)
	req.AddStreamChunk(large)
	req.AddStreamChunk(large)

	if req.NextStreamSeq != 3 {
		t.Fatalf("next seq = %d, want 3", req.NextStreamSeq)
	}
	window := req.StreamChunkWindow()
	if len(window.Chunks) != 1 {
		t.Fatalf("retained %d chunks, want 1 (byte trim)", len(window.Chunks))
	}
	if window.StartSeq != 2 {
		t.Fatalf("start seq = %d, want 2 (oldest trimmed)", window.StartSeq)
	}
	if window.Chunks[0] != large {
		t.Fatal("wrong chunk retained")
	}
}
