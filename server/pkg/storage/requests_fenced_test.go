package storage_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
)

// These tests pin the lease-fencing contract shared by the Postgres and
// in-memory stores: every claim grants a fresh lease epoch, fenced writes are
// rejected for anyone who does not hold the current grant, and renewal extends
// the lease deadline without ever crossing the absolute per-attempt timeout.

func TestClaimAssignsLeaseEpoch(t *testing.T) {
	runAgainstBoth(t, "claim epoch", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		sess := "sess-" + uid(t)
		mach := "mach-" + uid(t)
		reqID := "req-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, mach, time.Now())
		seedOwnedTool(t, s, sess, mach, "tool")
		seedPendingRequest(t, s, sess, reqID, "tool")

		claimed, ok, err := s.ClaimRequest(ctx, sess, reqID, mach, 30*time.Second)
		if err != nil || !ok {
			t.Fatalf("claim: ok=%v err=%v", ok, err)
		}
		if claimed.LeaseEpoch != 1 {
			t.Fatalf("first claim lease epoch: got %d want 1", claimed.LeaseEpoch)
		}
		if claimed.LeasedBy != mach {
			t.Fatalf("claim leasedBy: got %q want %q", claimed.LeasedBy, mach)
		}
	})
}

func TestRenewRequestLeaseExtendsDeadline(t *testing.T) {
	runAgainstBoth(t, "renew extends", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		sess := "sess-" + uid(t)
		mach := "mach-" + uid(t)
		reqID := "req-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, mach, time.Now())
		seedOwnedTool(t, s, sess, mach, "tool")
		seedPendingRequest(t, s, sess, reqID, "tool")

		claimed, ok, err := s.ClaimRequest(ctx, sess, reqID, mach, 5*time.Second)
		if err != nil || !ok {
			t.Fatalf("claim: ok=%v err=%v", ok, err)
		}

		// Ensure the clock advances past the claim's tick so the renewal's
		// now+ttl deadline is strictly later.
		time.Sleep(2 * time.Millisecond)
		renewed, err := s.RenewRequestLease(ctx, sess, reqID, mach, claimed.LeaseEpoch, 30*time.Second)
		if err != nil {
			t.Fatalf("renew: %v", err)
		}
		if !renewed.VisibleAt.After(claimed.VisibleAt) {
			t.Fatalf("renewal did not extend the lease deadline: before=%v after=%v", claimed.VisibleAt, renewed.VisibleAt)
		}
		if renewed.LeaseEpoch != claimed.LeaseEpoch {
			t.Fatalf("renewal changed the lease epoch: before=%d after=%d", claimed.LeaseEpoch, renewed.LeaseEpoch)
		}
		if renewed.LeasedBy != mach {
			t.Fatalf("renewal changed the lease holder: got %q want %q", renewed.LeasedBy, mach)
		}
	})
}

func TestRenewRequestLeaseCapsAtAbsoluteTimeout(t *testing.T) {
	runAgainstBoth(t, "renew capped", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		sess := "sess-" + uid(t)
		mach := "mach-" + uid(t)
		reqID := "req-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, mach, time.Now())
		seedOwnedTool(t, s, sess, mach, "tool")
		req := seedPendingRequest(t, s, sess, reqID, "tool")
		// Absolute per-attempt timeout of 5 seconds: renewal must never push
		// the lease deadline past leased_at + 5s no matter the requested TTL.
		req.TimeoutSeconds = 5
		if err := s.SaveRequest(ctx, req); err != nil {
			t.Fatalf("set timeout: %v", err)
		}

		claimed, ok, err := s.ClaimRequest(ctx, sess, reqID, mach, 3*time.Second)
		if err != nil || !ok {
			t.Fatalf("claim: ok=%v err=%v", ok, err)
		}

		renewed, err := s.RenewRequestLease(ctx, sess, reqID, mach, claimed.LeaseEpoch, time.Hour)
		if err != nil {
			t.Fatalf("renew: %v", err)
		}
		absoluteDeadline := claimed.LeasedAt.Add(5 * time.Second)
		if renewed.VisibleAt.After(absoluteDeadline.Add(time.Second)) {
			t.Fatalf("renewal crossed the absolute timeout: visibleAt=%v absoluteDeadline=%v", renewed.VisibleAt, absoluteDeadline)
		}

		// Once the absolute deadline passes, renewal is refused: the lease
		// cannot be resurrected.
		expired := *claimed
		pastLease := expired.LeasedAt.Add(-10 * time.Minute)
		expired.LeasedAt = &pastLease
		expired.VisibleAt = pastLease.Add(time.Second)
		if err := s.SaveRequest(ctx, &expired); err != nil {
			t.Fatalf("force expiry: %v", err)
		}
		if _, err := s.RenewRequestLease(ctx, sess, reqID, mach, claimed.LeaseEpoch, time.Hour); !errors.Is(err, storage.ErrLeaseConflict) {
			t.Fatalf("renewal past the absolute timeout should fail with ErrLeaseConflict, got: %v", err)
		}
	})
}

func TestRenewRequestLeaseRejectsStaleEpochAndWrongMachine(t *testing.T) {
	runAgainstBoth(t, "renew fence", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		sess := "sess-" + uid(t)
		machA := "machA-" + uid(t)
		machB := "machB-" + uid(t)
		reqID := "req-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, machA, time.Now())
		seedMachine(t, s, sess, machB, time.Now())
		seedOwnedTool(t, s, sess, machA, "tool")
		seedPendingRequest(t, s, sess, reqID, "tool")

		claimed, ok, err := s.ClaimRequest(ctx, sess, reqID, machA, 30*time.Second)
		if err != nil || !ok {
			t.Fatalf("claim: ok=%v err=%v", ok, err)
		}

		if _, err := s.RenewRequestLease(ctx, sess, reqID, machA, claimed.LeaseEpoch+1, 30*time.Second); !errors.Is(err, storage.ErrLeaseConflict) {
			t.Fatalf("renewal with stale epoch should fail with ErrLeaseConflict, got: %v", err)
		}
		if _, err := s.RenewRequestLease(ctx, sess, reqID, machB, claimed.LeaseEpoch, 30*time.Second); !errors.Is(err, storage.ErrLeaseConflict) {
			t.Fatalf("renewal by non-holder should fail with ErrLeaseConflict, got: %v", err)
		}
	})
}

// TestFencedWritesRejectStaleHolderAfterReclaim is the split-brain proof:
// after a lease is reclaimed and re-claimed by a second machine, the stale
// first executor can no longer submit results or append chunks, while the
// current holder can.
func TestFencedWritesRejectStaleHolderAfterReclaim(t *testing.T) {
	runAgainstBoth(t, "stale holder fenced", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		sess := "sess-" + uid(t)
		machA := "machA-" + uid(t)
		machB := "machB-" + uid(t)
		reqID := "req-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, machA, time.Now())
		seedMachine(t, s, sess, machB, time.Now())
		tool := model.NewTool(sess, machA, "tool", "test tool", `{}`, nil, nil)
		if err := s.SaveTool(ctx, tool); err != nil {
			t.Fatalf("seed tool: %v", err)
		}
		seedPendingRequest(t, s, sess, reqID, "tool")

		firstClaim, ok, err := s.ClaimRequest(ctx, sess, reqID, machA, 30*time.Second)
		if err != nil || !ok {
			t.Fatalf("claim A: ok=%v err=%v", ok, err)
		}
		staleEpoch := firstClaim.LeaseEpoch

		// Force the lease to look expired, then reclaim through the guarded
		// primitive (the reaper's path).
		stored, err := s.GetRequest(ctx, reqID)
		if err != nil || stored == nil {
			t.Fatalf("get request: %v", err)
		}
		past := time.Now().Add(-time.Hour)
		stored.LeasedAt = &past
		stored.VisibleAt = past.Add(time.Second)
		if err := s.SaveRequest(ctx, stored); err != nil {
			t.Fatalf("force expiry: %v", err)
		}
		requeued, reclaimed, err := s.ReclaimExpiredRequest(ctx, reqID, time.Now(), 30*time.Second, 3, time.Millisecond)
		if err != nil || !reclaimed {
			t.Fatalf("reclaim: reclaimed=%v err=%v", reclaimed, err)
		}
		if requeued.Status != model.RequestStatusPending {
			t.Fatalf("reclaim status: got %s want pending", requeued.Status)
		}

		// The requeue scheduled a retry backoff; let it elapse before the
		// second claim (claims now honor a not-yet-elapsed backoff).
		time.Sleep(5 * time.Millisecond)

		// Machine B took over the tool provision: the same registry row
		// moves to B, so the claim ownership check accepts it.
		tool.MachineID = machB
		if err := s.SaveTool(ctx, tool); err != nil {
			t.Fatalf("transfer tool to machB: %v", err)
		}

		secondClaim, ok, err := s.ClaimRequest(ctx, sess, reqID, machB, 30*time.Second)
		if err != nil || !ok {
			t.Fatalf("claim B: ok=%v err=%v", ok, err)
		}
		if secondClaim.LeaseEpoch <= staleEpoch {
			t.Fatalf("second claim epoch %d must exceed stale epoch %d", secondClaim.LeaseEpoch, staleEpoch)
		}

		// Stale executor tries to write: both writes must be rejected.
		if _, err := s.AppendRequestChunksFenced(ctx, sess, reqID, machA, staleEpoch, []string{"stale-chunk"}); !errors.Is(err, storage.ErrLeaseConflict) {
			t.Fatalf("stale append should fail with ErrLeaseConflict, got: %v", err)
		}
		if _, err := s.SubmitRequestResultFenced(ctx, sess, reqID, machA, staleEpoch, `{"stale":true}`, model.ResultTypeResolution, nil); !errors.Is(err, storage.ErrLeaseConflict) {
			t.Fatalf("stale submit should fail with ErrLeaseConflict, got: %v", err)
		}

		// Current holder writes succeed and land coherently.
		if _, err := s.AppendRequestChunksFenced(ctx, sess, reqID, machB, secondClaim.LeaseEpoch, []string{"chunk-1"}); err != nil {
			t.Fatalf("holder append: %v", err)
		}
		submitted, err := s.SubmitRequestResultFenced(ctx, sess, reqID, machB, secondClaim.LeaseEpoch, `{"ok":true}`, model.ResultTypeResolution, nil)
		if err != nil {
			t.Fatalf("holder submit: %v", err)
		}
		if submitted.Status != model.RequestStatusDone {
			t.Fatalf("submit status: got %s want done", submitted.Status)
		}
		// Chunk payloads live in the append-only table; assert coherence
		// through the authoritative window read.
		window, err := s.GetRequestChunksByRequest(ctx, reqID, submitted.StreamStartSeq, submitted.NextStreamSeq)
		if err != nil {
			t.Fatalf("chunk window read: %v", err)
		}
		if len(window.Chunks) != 1 || window.Chunks[0] != "chunk-1" {
			t.Fatalf("chunk window incoherent after fenced writes: %v", window.Chunks)
		}
	})
}

func TestFencedAppendRejectsNonHolder(t *testing.T) {
	runAgainstBoth(t, "append fence", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		sess := "sess-" + uid(t)
		machA := "machA-" + uid(t)
		machB := "machB-" + uid(t)
		reqID := "req-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, machA, time.Now())
		seedMachine(t, s, sess, machB, time.Now())
		seedOwnedTool(t, s, sess, machA, "tool")
		seedPendingRequest(t, s, sess, reqID, "tool")

		claimed, ok, err := s.ClaimRequest(ctx, sess, reqID, machA, 30*time.Second)
		if err != nil || !ok {
			t.Fatalf("claim: ok=%v err=%v", ok, err)
		}

		// Wrong machine with the right epoch.
		if _, err := s.AppendRequestChunksFenced(ctx, sess, reqID, machB, claimed.LeaseEpoch, []string{"forged"}); !errors.Is(err, storage.ErrLeaseConflict) {
			t.Fatalf("append by non-holder should fail with ErrLeaseConflict, got: %v", err)
		}
		// Right machine with a wrong epoch.
		if _, err := s.AppendRequestChunksFenced(ctx, sess, reqID, machA, claimed.LeaseEpoch+7, []string{"forged"}); !errors.Is(err, storage.ErrLeaseConflict) {
			t.Fatalf("append with stale epoch should fail with ErrLeaseConflict, got: %v", err)
		}
		// Holder succeeds.
		updated, err := s.AppendRequestChunksFenced(ctx, sess, reqID, machA, claimed.LeaseEpoch, []string{"real-1", "real-2"})
		if err != nil {
			t.Fatalf("holder append: %v", err)
		}
		if len(updated.StreamResults) != 2 {
			t.Fatalf("append window: got %v", updated.StreamResults)
		}
	})
}

func TestUpdateRequestFencedRunningTransition(t *testing.T) {
	runAgainstBoth(t, "update fence", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		sess := "sess-" + uid(t)
		machA := "machA-" + uid(t)
		machB := "machB-" + uid(t)
		reqID := "req-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, machA, time.Now())
		seedMachine(t, s, sess, machB, time.Now())
		seedOwnedTool(t, s, sess, machA, "tool")
		req := seedPendingRequest(t, s, sess, reqID, "tool")
		req.TimeoutSeconds = 77
		if err := s.SaveRequest(ctx, req); err != nil {
			t.Fatalf("set timeout: %v", err)
		}

		claimed, ok, err := s.ClaimRequest(ctx, sess, reqID, machA, 30*time.Second)
		if err != nil || !ok {
			t.Fatalf("claim: ok=%v err=%v", ok, err)
		}

		// Non-holder cannot flip the request to running.
		if _, err := s.UpdateRequestFenced(ctx, sess, reqID, machB, claimed.LeaseEpoch, model.RequestStatusRunning, nil, "", 30*time.Second); !errors.Is(err, storage.ErrLeaseConflict) {
			t.Fatalf("running transition by non-holder should fail with ErrLeaseConflict, got: %v", err)
		}

		running, err := s.UpdateRequestFenced(ctx, sess, reqID, machA, claimed.LeaseEpoch, model.RequestStatusRunning, nil, "", 30*time.Second)
		if err != nil {
			t.Fatalf("running transition: %v", err)
		}
		if running.Status != model.RequestStatusRunning || running.LeasedBy != machA {
			t.Fatalf("running state: status=%s leasedBy=%s", running.Status, running.LeasedBy)
		}
		// The request's own timeout survives the running transition (the
		// transition must not clobber a caller-supplied timeout override).
		if running.TimeoutSeconds != 77 {
			t.Fatalf("running transition clobbered timeout: got %d want 77", running.TimeoutSeconds)
		}
	})
}

func TestSubmitRequestResultFencedTerminalSemantics(t *testing.T) {
	runAgainstBoth(t, "terminal semantics", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		sess := "sess-" + uid(t)
		mach := "mach-" + uid(t)
		reqID := "req-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, mach, time.Now())
		seedOwnedTool(t, s, sess, mach, "tool")
		seedPendingRequest(t, s, sess, reqID, "tool")

		claimed, ok, err := s.ClaimRequest(ctx, sess, reqID, mach, 30*time.Second)
		if err != nil || !ok {
			t.Fatalf("claim: ok=%v err=%v", ok, err)
		}

		done, err := s.SubmitRequestResultFenced(ctx, sess, reqID, mach, claimed.LeaseEpoch, `{"ok":true}`, model.ResultTypeResolution, nil)
		if err != nil {
			t.Fatalf("submit: %v", err)
		}
		if done.Status != model.RequestStatusDone {
			t.Fatalf("submit status: got %s want done", done.Status)
		}

		// A second non-streaming submission reports the typed terminal error
		// (not a fencing error).
		if _, err := s.SubmitRequestResultFenced(ctx, sess, reqID, mach, claimed.LeaseEpoch, `{"again":true}`, model.ResultTypeResolution, nil); !errors.Is(err, storage.ErrRequestTerminal) {
			t.Fatalf("second submit should fail with ErrRequestTerminal, got: %v", err)
		}

		// The final lease holder may still append a trailing streaming chunk
		// using its epoch.
		appended, err := s.SubmitRequestResultFenced(ctx, sess, reqID, mach, claimed.LeaseEpoch, "trailing-chunk", model.ResultTypeStreaming, nil)
		if err != nil {
			t.Fatalf("trailing streaming submit: %v", err)
		}
		if len(appended.StreamResults) != 1 || appended.StreamResults[0] != "trailing-chunk" {
			t.Fatalf("trailing chunk window: %v", appended.StreamResults)
		}

		// A stranger with a fabricated epoch gets a fencing rejection.
		if _, err := s.SubmitRequestResultFenced(ctx, sess, reqID, mach, claimed.LeaseEpoch+1, "forged", model.ResultTypeStreaming, nil); !errors.Is(err, storage.ErrLeaseConflict) {
			t.Fatalf("forged trailing submit should fail with ErrLeaseConflict, got: %v", err)
		}
	})
}

func TestRequeueRequestFenced(t *testing.T) {
	runAgainstBoth(t, "requeue fence", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		sess := "sess-" + uid(t)
		machA := "machA-" + uid(t)
		machB := "machB-" + uid(t)
		reqID := "req-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, machA, time.Now())
		seedMachine(t, s, sess, machB, time.Now())
		seedOwnedTool(t, s, sess, machA, "tool")
		seedPendingRequest(t, s, sess, reqID, "tool")

		claimed, ok, err := s.ClaimRequest(ctx, sess, reqID, machA, 30*time.Second)
		if err != nil || !ok {
			t.Fatalf("claim: ok=%v err=%v", ok, err)
		}

		if _, err := s.RequeueRequestFenced(ctx, sess, reqID, machB, claimed.LeaseEpoch, "machine at capacity", time.Second); !errors.Is(err, storage.ErrLeaseConflict) {
			t.Fatalf("requeue by non-holder should fail with ErrLeaseConflict, got: %v", err)
		}

		requeued, err := s.RequeueRequestFenced(ctx, sess, reqID, machA, claimed.LeaseEpoch, "machine at capacity", time.Second)
		if err != nil {
			t.Fatalf("requeue: %v", err)
		}
		if requeued.Status != model.RequestStatusPending || requeued.LeasedBy != "" {
			t.Fatalf("requeue state: status=%s leasedBy=%q", requeued.Status, requeued.LeasedBy)
		}
		if !strings.Contains(requeued.LastError, "capacity") {
			t.Fatalf("requeue lastError: %q", requeued.LastError)
		}
	})
}

func TestCancelRequestFencedCancelsNonTerminal(t *testing.T) {
	runAgainstBoth(t, "cancel pending", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		sess := "sess-" + uid(t)
		reqID := "req-" + uid(t)
		seedSession(t, s, sess)
		seedPendingRequest(t, s, sess, reqID, "tool")

		cancelled, ok, err := s.CancelRequestFenced(ctx, sess, reqID, "Request was cancelled")
		if err != nil || !ok {
			t.Fatalf("cancel: ok=%v err=%v", ok, err)
		}
		if cancelled.Status != model.RequestStatusFailed {
			t.Fatalf("cancel status: got %s want failed", cancelled.Status)
		}
		if !cancelled.DeadLetter {
			t.Fatal("cancel should mark dead_letter")
		}
		if cancelled.LastError != "Request was cancelled" {
			t.Fatalf("cancel lastError: %q", cancelled.LastError)
		}
		if cancelled.LeasedBy != "" || cancelled.LeasedAt != nil {
			t.Fatalf("cancel should release the lease: leasedBy=%q leasedAt=%v", cancelled.LeasedBy, cancelled.LeasedAt)
		}

		// The cancel is durable: a re-read sees the terminal state.
		stored, err := s.GetRequest(ctx, reqID)
		if err != nil {
			t.Fatalf("get after cancel: %v", err)
		}
		if stored.Status != model.RequestStatusFailed || !stored.DeadLetter {
			t.Fatalf("stored after cancel: status=%s deadLetter=%v", stored.Status, stored.DeadLetter)
		}
	})
}

func TestCancelRequestFencedRefusesTerminal(t *testing.T) {
	runAgainstBoth(t, "cancel terminal", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		sess := "sess-" + uid(t)
		reqID := "req-" + uid(t)
		seedSession(t, s, sess)
		req := seedPendingRequest(t, s, sess, reqID, "tool")

		// Complete the request first: the terminal outcome must win.
		req.SetResult(map[string]string{"answer": "42"}, model.ResultTypeResolution, "")
		if err := s.SaveRequest(ctx, req); err != nil {
			t.Fatalf("persist completed request: %v", err)
		}

		terminal, ok, err := s.CancelRequestFenced(ctx, sess, reqID, "Request was cancelled")
		if err != nil {
			t.Fatalf("cancel terminal: %v", err)
		}
		if ok {
			t.Fatal("cancel of a terminal request must return cancelled=false")
		}
		if terminal.Status != model.RequestStatusDone {
			t.Fatalf("terminal status: got %s want done", terminal.Status)
		}

		// The result survived untouched.
		stored, err := s.GetRequest(ctx, reqID)
		if err != nil {
			t.Fatalf("get after refused cancel: %v", err)
		}
		if stored.Status != model.RequestStatusDone {
			t.Fatalf("stored status after refused cancel: %s want done", stored.Status)
		}
		if stored.Result == nil {
			t.Fatal("terminal result was clobbered by the refused cancel")
		}
	})
}

func TestCancelRequestFencedMissingRequest(t *testing.T) {
	runAgainstBoth(t, "cancel missing", func(t *testing.T, s storage.Storer) {
		sess := "sess-" + uid(t)
		seedSession(t, s, sess)
		_, _, err := s.CancelRequestFenced(context.Background(), sess, "req-does-not-exist", "Request was cancelled")
		if !errors.Is(err, storage.ErrNotFound) {
			t.Fatalf("cancel missing request: want ErrNotFound, got %v", err)
		}
	})
}

func TestInsertRequestRejectsDuplicate(t *testing.T) {
	runAgainstBoth(t, "insert duplicate", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		sess := "sess-" + uid(t)
		reqID := "req-" + uid(t)
		seedSession(t, s, sess)
		req := seedPendingRequest(t, s, sess, reqID, "tool")

		// InsertRequest on an already-persisted id must fail instead of
		// silently overwriting the row (insert-only create contract).
		if err := s.InsertRequest(ctx, req); err == nil {
			t.Fatal("InsertRequest on an existing id should fail")
		}
	})
}

func TestClaimRequestRejectsFutureVisibleAt(t *testing.T) {
	runAgainstBoth(t, "claim future visible_at", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		sess := "sess-" + uid(t)
		reqID := "req-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, "mach-1", time.Now())
		seedOwnedTool(t, s, sess, "mach-1", "tool")
		req := seedPendingRequest(t, s, sess, reqID, "tool")

		// A queued retry whose backoff has not elapsed is not claimable.
		req.VisibleAt = time.Now().Add(time.Hour)
		if err := s.SaveRequest(ctx, req); err != nil {
			t.Fatalf("persist future visible_at: %v", err)
		}
		if _, ok, err := s.ClaimRequest(ctx, sess, reqID, "mach-1", 30*time.Second); err != nil || ok {
			t.Fatalf("claim with future visible_at: ok=%v err=%v, want ok=false", ok, err)
		}

		// Once the backoff elapses the same request claims normally.
		req.VisibleAt = time.Now().Add(-time.Second)
		if err := s.SaveRequest(ctx, req); err != nil {
			t.Fatalf("persist past visible_at: %v", err)
		}
		claimed, ok, err := s.ClaimRequest(ctx, sess, reqID, "mach-1", 30*time.Second)
		if err != nil || !ok {
			t.Fatalf("claim after backoff: ok=%v err=%v", ok, err)
		}
		if claimed.Attempts != 1 {
			t.Fatalf("first claim attempts=%d want 1", claimed.Attempts)
		}
	})
}

func TestClaimRequestRejectsExhaustedAttempts(t *testing.T) {
	runAgainstBoth(t, "claim exhausted", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		sess := "sess-" + uid(t)
		reqID := "req-" + uid(t)
		seedSession(t, s, sess)
		req := seedPendingRequest(t, s, sess, reqID, "tool")
		req.MaxAttempts = 2
		req.Attempts = 2
		if err := s.SaveRequest(ctx, req); err != nil {
			t.Fatalf("persist exhausted request: %v", err)
		}

		if _, ok, err := s.ClaimRequest(ctx, sess, reqID, "mach-1", 30*time.Second); err != nil || ok {
			t.Fatalf("claim of an exhausted request: ok=%v err=%v, want ok=false", ok, err)
		}
	})
}

func TestRequeueDoesNotDoubleCountAttempts(t *testing.T) {
	runAgainstBoth(t, "requeue attempts", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		sess := "sess-" + uid(t)
		reqID := "req-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, "mach-1", time.Now())
		seedOwnedTool(t, s, sess, "mach-1", "tool")
		seedPendingRequest(t, s, sess, reqID, "tool")

		// First execution: the claim counts the attempt.
		claimed, ok, err := s.ClaimRequest(ctx, sess, reqID, "mach-1", 30*time.Second)
		if err != nil || !ok {
			t.Fatalf("claim: ok=%v err=%v", ok, err)
		}
		if claimed.Attempts != 1 {
			t.Fatalf("claimed attempts=%d want 1", claimed.Attempts)
		}

		// The lease expires and the reaper requeues. The attempt was already
		// counted by the claim: the requeue must not bump it to 2, or the
		// request dead-letters one execution early.
		stored, err := s.GetRequest(ctx, reqID)
		if err != nil {
			t.Fatalf("get claimed request: %v", err)
		}
		past := time.Now().Add(-time.Hour)
		stored.LeasedAt = &past
		stored.VisibleAt = past.Add(30 * time.Second)
		if err := s.SaveRequest(ctx, stored); err != nil {
			t.Fatalf("forge expired lease: %v", err)
		}
		requeued, ok, err := s.ReclaimExpiredRequest(ctx, reqID, time.Now(), 30*time.Second, 3, 5*time.Second)
		if err != nil || !ok {
			t.Fatalf("reclaim: ok=%v err=%v", ok, err)
		}
		if requeued.Status != model.RequestStatusPending {
			t.Fatalf("requeued status=%s want pending", requeued.Status)
		}
		if requeued.Attempts != 1 {
			t.Fatalf("requeued attempts=%d want 1 (double-increment = bug)", requeued.Attempts)
		}
		if !requeued.VisibleAt.After(time.Now()) {
			t.Fatalf("requeued visible_at %v must honor the backoff", requeued.VisibleAt)
		}

		// While the backoff holds, the requeued request is not claimable.
		if _, ok, err := s.ClaimRequest(ctx, sess, reqID, "mach-2", 30*time.Second); err != nil || ok {
			t.Fatalf("claim during backoff: ok=%v err=%v, want ok=false", ok, err)
		}
	})
}

func TestDeadLetterAfterExactlyMaxAttemptsExecutions(t *testing.T) {
	runAgainstBoth(t, "dead-letter attempts", func(t *testing.T, s storage.Storer) {
		ctx := context.Background()
		sess := "sess-" + uid(t)
		reqID := "req-" + uid(t)
		seedSession(t, s, sess)
		seedMachine(t, s, sess, "mach-1", time.Now())
		seedMachine(t, s, sess, "mach-2", time.Now())
		tool := model.NewTool(sess, "mach-1", "tool", "test tool", `{}`, nil, nil)
		if err := s.SaveTool(ctx, tool); err != nil {
			t.Fatalf("seed tool: %v", err)
		}
		req := seedPendingRequest(t, s, sess, reqID, "tool")
		req.MaxAttempts = 2
		if err := s.SaveRequest(ctx, req); err != nil {
			t.Fatalf("persist request: %v", err)
		}

		// Execution 1: claim, expire, requeue (still within budget).
		if _, ok, err := s.ClaimRequest(ctx, sess, reqID, "mach-1", 30*time.Second); err != nil || !ok {
			t.Fatalf("claim 1: ok=%v err=%v", ok, err)
		}
		stored, err := s.GetRequest(ctx, reqID)
		if err != nil {
			t.Fatalf("get claimed request: %v", err)
		}
		past := time.Now().Add(-time.Hour)
		stored.LeasedAt = &past
		stored.VisibleAt = past.Add(30 * time.Second)
		if err := s.SaveRequest(ctx, stored); err != nil {
			t.Fatalf("forge expired lease: %v", err)
		}
		requeued, ok, err := s.ReclaimExpiredRequest(ctx, reqID, time.Now(), 30*time.Second, 2, 0)
		if err != nil || !ok || requeued.Status != model.RequestStatusPending {
			t.Fatalf("reclaim 1: ok=%v status=%s err=%v", ok, requeued.Status, err)
		}

		// Execution 2: the second claim exhausts the budget of 2. Mach-2
		// took over the tool provision: the same registry row moves to it.
		tool.MachineID = "mach-2"
		if err := s.SaveTool(ctx, tool); err != nil {
			t.Fatalf("transfer tool to mach-2: %v", err)
		}
		if _, ok, err := s.ClaimRequest(ctx, sess, reqID, "mach-2", 30*time.Second); err != nil || !ok {
			t.Fatalf("claim 2: ok=%v err=%v", ok, err)
		}
		stored2, err := s.GetRequest(ctx, reqID)
		if err != nil {
			t.Fatalf("get claimed request 2: %v", err)
		}
		past2 := time.Now().Add(-time.Hour)
		stored2.LeasedAt = &past2
		stored2.VisibleAt = past2.Add(30 * time.Second)
		if err := s.SaveRequest(ctx, stored2); err != nil {
			t.Fatalf("forge expired lease 2: %v", err)
		}
		dead, ok, err := s.ReclaimExpiredRequest(ctx, reqID, time.Now(), 30*time.Second, 2, 0)
		if err != nil || !ok {
			t.Fatalf("reclaim 2: ok=%v err=%v", ok, err)
		}
		if dead.Status != model.RequestStatusFailed || !dead.DeadLetter {
			t.Fatalf("final status=%s deadLetter=%v, want failed+dead-letter", dead.Status, dead.DeadLetter)
		}
		if dead.Attempts != 2 {
			t.Fatalf("final attempts=%d want 2 (exactly MaxAttempts executions)", dead.Attempts)
		}
	})
}
