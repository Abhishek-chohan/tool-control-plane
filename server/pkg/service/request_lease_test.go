package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"toolplane/pkg/model"
	"toolplane/pkg/storage"
	"toolplane/pkg/trace"
	proto "toolplane/proto"
)

// These tests pin the lease-renewal and fencing guarantees: a renewed
// long-running execution is not reclaimed mid-flight, the absolute
// per-attempt timeout still bounds it, and once a lease is reclaimed the
// stale executor can no longer write results or chunks.

func newLeaseTestStack() (*RequestsService, *MachinesService, string, string) {
	tracer := trace.NopTracer()
	toolService := NewToolService(tracer, nil)
	machineService := NewMachinesService(context.Background(), toolService, tracer, nil)
	requestService := NewRequestsService(context.Background(), toolService, machineService, tracer, nil)

	const sessionID = "session-lease"
	const machineID = "machine-lease"
	_, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, "")
	if err != nil {
		panic(err)
	}
	return requestService, machineService, sessionID, machineID
}

func TestCreateRequestTimeoutOverride(t *testing.T) {
	requestService, _, sessionID, _ := newLeaseTestStack()

	request, err := requestService.CreateRequest(sessionID, "echo", `{"x":1}`, 120, "")
	if err != nil {
		t.Fatalf("create with override: %v", err)
	}
	if request.TimeoutSeconds != 120 {
		t.Fatalf("timeout = %d, want 120", request.TimeoutSeconds)
	}

	if _, err := requestService.CreateRequest(sessionID, "echo", `{"x":1}`, 0, ""); err != nil {
		t.Fatalf("create with default: %v", err)
	}

	if _, err := requestService.CreateRequest(sessionID, "echo", `{"x":1}`, int(maxRequestTimeout.Seconds())+1, ""); !errors.Is(err, ErrRequestTimeoutOutOfRange) {
		t.Fatalf("over-maximum timeout should fail with ErrRequestTimeoutOutOfRange, got: %v", err)
	}
}

func TestRenewRequestLeasePreventsReclaimAndAbsoluteCapStillBounds(t *testing.T) {
	requestService, _, sessionID, machineID := newLeaseTestStack()

	request, err := requestService.CreateRequest(sessionID, "echo", `{"x":1}`, 0, "")
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	claimed, err := requestService.ClaimRequest(sessionID, request.ID, machineID)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	// The in-memory service path returns the cached request pointer, so copy
	// the claim deadline before renewing mutates the same object.
	claimDeadline := claimed.VisibleAt
	claimEpoch := claimed.LeaseEpoch

	// Ensure the clock advances past the claim's tick so the renewal's
	// now+leaseDuration deadline is strictly later.
	time.Sleep(2 * time.Millisecond)
	renewed, err := requestService.RenewRequestLease(sessionID, request.ID, machineID, claimEpoch)
	if err != nil {
		t.Fatalf("renew: %v", err)
	}
	if !renewed.VisibleAt.After(claimDeadline) {
		t.Fatalf("renewal did not extend the lease: before=%v after=%v", claimDeadline, renewed.VisibleAt)
	}

	// Forge 35 seconds of elapsed time (past the 30s lease TTL) while staying
	// inside the 45s absolute timeout: without the renewal above the request
	// would be reclaimable now; with it, the reaper must leave it alone.
	mutateCachedRequestForTest(requestService, request.ID, func(r *model.Request) {
		r.LeasedAt = ptrTime(time.Now().Add(-35 * time.Second))
	})
	requestService.markStalledRequests()
	afterReap, err := requestService.GetRequestByID(sessionID, request.ID)
	if err != nil {
		t.Fatalf("get request: %v", err)
	}
	if afterReap.Status != model.RequestStatusRunning && afterReap.Status != model.RequestStatusClaimed {
		t.Fatalf("renewed lease was reclaimed: status=%s (want claimed/running)", afterReap.Status)
	}

	// Once the lease deadline passes without a further renewal, the reaper
	// reclaims as usual.
	mutateCachedRequestForTest(requestService, request.ID, func(r *model.Request) {
		r.VisibleAt = time.Now().Add(-time.Second)
	})
	requestService.markStalledRequests()
	requeued, err := requestService.GetRequestByID(sessionID, request.ID)
	if err != nil {
		t.Fatalf("get request after expiry: %v", err)
	}
	if requeued.Status != model.RequestStatusPending {
		t.Fatalf("expired renewed lease not reclaimed: status=%s want pending", requeued.Status)
	}
}

func TestRenewRequestLeaseCannotCrossAbsoluteTimeout(t *testing.T) {
	requestService, _, sessionID, machineID := newLeaseTestStack()

	// Absolute timeout of 2 seconds: renewal must be capped at leased_at+2s
	// and refused entirely once that deadline has passed.
	request, err := requestService.CreateRequest(sessionID, "echo", `{"x":1}`, 2, "")
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	claimed, err := requestService.ClaimRequest(sessionID, request.ID, machineID)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}

	renewed, err := requestService.RenewRequestLease(sessionID, request.ID, machineID, claimed.LeaseEpoch)
	if err != nil {
		t.Fatalf("renew: %v", err)
	}
	absoluteDeadline := claimed.LeasedAt.Add(2 * time.Second)
	if renewed.VisibleAt.After(absoluteDeadline.Add(100 * time.Millisecond)) {
		t.Fatalf("renewal crossed the absolute timeout: visibleAt=%v deadline=%v", renewed.VisibleAt, absoluteDeadline)
	}

	// Forge the absolute deadline into the past: renewal is now refused.
	mutateCachedRequestForTest(requestService, request.ID, func(r *model.Request) {
		r.LeasedAt = ptrTime(time.Now().Add(-time.Hour))
		r.VisibleAt = time.Now().Add(-time.Second)
	})
	if _, err := requestService.RenewRequestLease(sessionID, request.ID, machineID, claimed.LeaseEpoch); !errors.Is(err, storage.ErrLeaseConflict) {
		t.Fatalf("renewal past the absolute timeout should fail with ErrLeaseConflict, got: %v", err)
	}
}

// TestFencedWritesAfterReclaim is the split-brain proof at the service layer:
// after a lease is reclaimed and re-claimed by a second machine, the stale
// first executor's submissions are rejected while the current holder's writes
// succeed. This is the in-memory dev-mode path.
func TestFencedWritesAfterReclaim(t *testing.T) {
	requestService, machineService, sessionID, machineA := newLeaseTestStack()
	const machineB = "machine-lease-b"
	if _, err := machineService.RegisterMachine(sessionID, machineB, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineB, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine B: %v", err)
	}

	request, err := requestService.CreateRequest(sessionID, "echo", `{"x":1}`, 0, "")
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	firstClaim, err := requestService.ClaimRequest(sessionID, request.ID, machineA)
	if err != nil {
		t.Fatalf("claim A: %v", err)
	}
	staleEpoch := firstClaim.LeaseEpoch

	// Force lease expiry and let the reaper requeue.
	mutateCachedRequestForTest(requestService, request.ID, func(r *model.Request) {
		r.VisibleAt = time.Now().Add(-time.Second)
	})
	requestService.markStalledRequests()
	requeued, err := requestService.GetRequestByID(sessionID, request.ID)
	if err != nil || requeued.Status != model.RequestStatusPending {
		t.Fatalf("reclaim: status=%v err=%v", requeued, err)
	}

	secondClaim, err := requestService.ClaimRequest(sessionID, request.ID, machineB)
	if err != nil {
		t.Fatalf("claim B: %v", err)
	}
	if secondClaim.LeaseEpoch <= staleEpoch {
		t.Fatalf("second claim epoch %d must exceed stale epoch %d", secondClaim.LeaseEpoch, staleEpoch)
	}

	// Stale executor writes are rejected.
	if err := requestService.AppendRequestChunks(sessionID, request.ID, machineA, staleEpoch, []string{"stale"}, model.ResultTypeStreaming); !errors.Is(err, storage.ErrLeaseConflict) {
		t.Fatalf("stale append should fail with ErrLeaseConflict, got: %v", err)
	}
	if err := requestService.SubmitRequestResult(sessionID, request.ID, machineA, staleEpoch, `{"stale":true}`, model.ResultTypeResolution, nil); !errors.Is(err, storage.ErrLeaseConflict) {
		t.Fatalf("stale submit should fail with ErrLeaseConflict, got: %v", err)
	}
	if _, err := requestService.UpdateRequest(sessionID, request.ID, machineA, staleEpoch, model.RequestStatusRunning, nil, ""); !errors.Is(err, storage.ErrLeaseConflict) {
		t.Fatalf("stale update should fail with ErrLeaseConflict, got: %v", err)
	}
	if _, err := requestService.RenewRequestLease(sessionID, request.ID, machineA, staleEpoch); !errors.Is(err, storage.ErrLeaseConflict) {
		t.Fatalf("stale renew should fail with ErrLeaseConflict, got: %v", err)
	}

	// Current holder writes succeed and land coherently.
	if _, err := requestService.UpdateRequest(sessionID, request.ID, machineB, secondClaim.LeaseEpoch, model.RequestStatusRunning, nil, ""); err != nil {
		t.Fatalf("holder update: %v", err)
	}
	if err := requestService.AppendRequestChunks(sessionID, request.ID, machineB, secondClaim.LeaseEpoch, []string{"chunk-1"}, model.ResultTypeStreaming); err != nil {
		t.Fatalf("holder append: %v", err)
	}
	if err := requestService.SubmitRequestResult(sessionID, request.ID, machineB, secondClaim.LeaseEpoch, `{"ok":true}`, model.ResultTypeResolution, nil); err != nil {
		t.Fatalf("holder submit: %v", err)
	}
	final, err := requestService.GetRequestByID(sessionID, request.ID)
	if err != nil {
		t.Fatalf("get final: %v", err)
	}
	if final.Status != model.RequestStatusDone {
		t.Fatalf("final status = %s, want done", final.Status)
	}
	if len(final.StreamResults) != 1 || final.StreamResults[0] != "chunk-1" {
		t.Fatalf("chunk window incoherent after fenced re-execution: %v", final.StreamResults)
	}
}

// TestActiveActiveFencedWritesAcrossInstances proves fencing holds across two
// service stacks sharing one store: a stale write presented to the "wrong"
// replica is still rejected because the check happens against the store row.
func TestActiveActiveFencedWritesAcrossInstances(t *testing.T) {
	store, svcA, machA, svcB, _ := newActiveActiveStacks(t)
	seedSessionForAA(t, store)

	echoTool := model.NewTool("sess-aa", "machine-a", "echo", "d", `{}`, nil, nil)
	if _, err := machA.RegisterMachine("sess-aa", "machine-a", "1.0", "go", "127.0.0.1", []*model.Tool{echoTool}, ""); err != nil {
		t.Fatalf("RegisterMachine on A: %v", err)
	}
	machineB := model.NewMachine("sess-aa", "machine-b", "1.0", "go", "127.0.0.1")
	_ = store.SaveMachine(context.Background(), machineB)

	req, err := svcA.CreateRequest("sess-aa", "echo", `{"x":1}`, 0, "")
	if err != nil {
		t.Fatalf("CreateRequest on A: %v", err)
	}
	claimedA, err := svcA.ClaimRequest("sess-aa", req.ID, "machine-a")
	if err != nil {
		t.Fatalf("ClaimRequest on A: %v", err)
	}

	// Cross-instance renewal with the correct grant succeeds (store-backed).
	if _, err := svcB.RenewRequestLease("sess-aa", req.ID, "machine-a", claimedA.LeaseEpoch); err != nil {
		t.Fatalf("cross-instance renew: %v", err)
	}

	// Force expiry + reclaim, then re-claim on the other instance's machine.
	stored, _ := store.GetRequest(context.Background(), req.ID)
	stored.LeasedAt = ptrTime(time.Now().Add(-time.Hour))
	stored.VisibleAt = time.Now().Add(-time.Second)
	if err := store.SaveRequest(context.Background(), stored); err != nil {
		t.Fatalf("force expiry: %v", err)
	}
	svcA.markStalledRequests()
	requeued, err := store.GetRequest(context.Background(), req.ID)
	if err != nil || requeued.Status != model.RequestStatusPending {
		t.Fatalf("reclaim: status=%v err=%v", requeued, err)
	}
	claimedB, err := svcB.ClaimRequest("sess-aa", req.ID, "machine-b")
	if err != nil {
		t.Fatalf("ClaimRequest on B: %v", err)
	}

	// The stale grant presented to instance A must be rejected even though the
	// current holder is known only to instance B's cache.
	if err := svcA.SubmitRequestResult("sess-aa", req.ID, "machine-a", claimedA.LeaseEpoch, `{"stale":true}`, model.ResultTypeResolution, nil); !errors.Is(err, storage.ErrLeaseConflict) {
		t.Fatalf("stale cross-instance submit should fail with ErrLeaseConflict, got: %v", err)
	}
	// The current holder's fenced write succeeds via instance B.
	if err := svcB.SubmitRequestResult("sess-aa", req.ID, "machine-b", claimedB.LeaseEpoch, `{"ok":true}`, model.ResultTypeResolution, nil); err != nil {
		t.Fatalf("holder submit on B: %v", err)
	}
	final, _ := store.GetRequest(context.Background(), req.ID)
	if final.Status != model.RequestStatusDone {
		t.Fatalf("final status = %s, want done", final.Status)
	}
}

// TestGRPCServerFencedWritesMapLeaseConflictToFailedPrecondition pins the wire
// contract: fencing rejections surface as FAILED_PRECONDITION so SDKs can
// distinguish "lost the lease" from "unknown request" (NOT_FOUND).
func TestGRPCServerFencedWritesMapLeaseConflictToFailedPrecondition(t *testing.T) {
	requestService, _, sessionID, machineID := newLeaseTestStack()
	toolService := NewToolService(trace.NopTracer(), nil)
	machineService := NewMachinesService(context.Background(), toolService, trace.NopTracer(), nil)
	server := NewGRPCServer(toolService, nil, machineService, requestService, nil)

	request, err := requestService.CreateRequest(sessionID, "echo", `{"x":1}`, 0, "")
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	claimed, err := requestService.ClaimRequest(sessionID, request.ID, machineID)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}

	_, err = server.SubmitRequestResult(context.Background(), &proto.SubmitRequestResultRequest{
		SessionId:  sessionID,
		RequestId:  request.ID,
		MachineId:  machineID,
		LeaseEpoch: claimed.LeaseEpoch + 41,
		Result:     `{"forged":true}`,
		ResultType: string(model.ResultTypeResolution),
	})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("forged submit status = %v, want %v (err=%v)", status.Code(err), codes.FailedPrecondition, err)
	}

	_, err = server.RenewRequestLease(context.Background(), &proto.RenewRequestLeaseRequest{
		SessionId:  sessionID,
		RequestId:  request.ID,
		MachineId:  machineID,
		LeaseEpoch: claimed.LeaseEpoch + 41,
	})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("forged renew status = %v, want %v (err=%v)", status.Code(err), codes.FailedPrecondition, err)
	}

	// Unknown requests keep the NOT_FOUND mapping.
	_, err = server.RenewRequestLease(context.Background(), &proto.RenewRequestLeaseRequest{
		SessionId:  sessionID,
		RequestId:  "no-such-request",
		MachineId:  machineID,
		LeaseEpoch: 1,
	})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("unknown request renew status = %v, want %v (err=%v)", status.Code(err), codes.NotFound, err)
	}

	// A correct grant renews over the wire and exposes lease metadata.
	renewed, err := server.RenewRequestLease(context.Background(), &proto.RenewRequestLeaseRequest{
		SessionId:  sessionID,
		RequestId:  request.ID,
		MachineId:  machineID,
		LeaseEpoch: claimed.LeaseEpoch,
	})
	if err != nil {
		t.Fatalf("renew over the wire: %v", err)
	}
	if renewed.GetLeaseEpoch() != claimed.LeaseEpoch {
		t.Fatalf("renewed epoch = %d, want %d", renewed.GetLeaseEpoch(), claimed.LeaseEpoch)
	}
	if renewed.GetLeasedBy() != machineID {
		t.Fatalf("renewed leasedBy = %q, want %q", renewed.GetLeasedBy(), machineID)
	}
	if renewed.GetLeaseExpiresAt() == "" {
		t.Fatal("renewed lease_expires_at should be populated for an active lease")
	}

	// A valid submit completes the request; a second submit is a write against
	// a terminal request and must surface as FAILED_PRECONDITION, not
	// NOT_FOUND.
	if _, err := server.SubmitRequestResult(context.Background(), &proto.SubmitRequestResultRequest{
		SessionId:  sessionID,
		RequestId:  request.ID,
		MachineId:  machineID,
		LeaseEpoch: claimed.LeaseEpoch,
		Result:     `{"ok":true}`,
		ResultType: string(model.ResultTypeResolution),
	}); err != nil {
		t.Fatalf("holder submit: %v", err)
	}
	_, err = server.SubmitRequestResult(context.Background(), &proto.SubmitRequestResultRequest{
		SessionId:  sessionID,
		RequestId:  request.ID,
		MachineId:  machineID,
		LeaseEpoch: claimed.LeaseEpoch,
		Result:     `{"again":true}`,
		ResultType: string(model.ResultTypeResolution),
	})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("terminal submit status = %v, want %v (err=%v)", status.Code(err), codes.FailedPrecondition, err)
	}
}

func ptrTime(t time.Time) *time.Time { return &t }
