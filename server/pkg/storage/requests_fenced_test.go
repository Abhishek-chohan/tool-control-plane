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
		if len(submitted.StreamResults) != 1 || submitted.StreamResults[0] != "chunk-1" {
			t.Fatalf("chunk window incoherent after fenced writes: %v", submitted.StreamResults)
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

		// A second non-streaming submission reports the historical
		// already-terminal error (not a fencing error).
		if _, err := s.SubmitRequestResultFenced(ctx, sess, reqID, mach, claimed.LeaseEpoch, `{"again":true}`, model.ResultTypeResolution, nil); err == nil || errors.Is(err, storage.ErrLeaseConflict) {
			t.Fatalf("second submit should fail with already-in-state error, got: %v", err)
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
