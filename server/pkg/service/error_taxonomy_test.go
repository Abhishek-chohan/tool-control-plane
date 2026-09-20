package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"toolplane/internal/auth"
	"toolplane/pkg/model"
	"toolplane/pkg/storage"
	"toolplane/pkg/trace"
	proto "toolplane/proto"
)

func TestStatusFromDomainErrorMapping(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want codes.Code
	}{
		{"not found", wrapf(ErrNotFound, "request r1"), codes.NotFound},
		{"storage not found", fmt.Errorf("wrap: %w", storage.ErrNotFound), codes.NotFound},
		{"already exists", wrapf(ErrAlreadyExists, "session s1"), codes.AlreadyExists},
		{"invalid argument", wrapf(ErrInvalidArgument, "bad input"), codes.InvalidArgument},
		{"capability required", model.ErrAPIKeyCapabilitiesRequired, codes.InvalidArgument},
		{"unsupported capability", model.ErrUnsupportedAPIKeyCapability, codes.InvalidArgument},
		{"machine draining", wrapf(ErrMachineDraining, "machine m1"), codes.FailedPrecondition},
		{"request not claimable", wrapf(ErrRequestNotClaimable, "request r1"), codes.FailedPrecondition},
		{"request not cancellable", wrapf(ErrRequestNotCancellable, "request r1"), codes.FailedPrecondition},
		{"task not cancellable", wrapf(ErrTaskNotCancellable, "task t1"), codes.FailedPrecondition},
		{"no provider available", wrapf(ErrNoProviderAvailable, "tool echo"), codes.FailedPrecondition},
		{"lease conflict", fmt.Errorf("wrap: %w", storage.ErrLeaseConflict), codes.FailedPrecondition},
		{"request terminal", fmt.Errorf("wrap: %w", storage.ErrRequestTerminal), codes.FailedPrecondition},
		{"machine at capacity", wrapf(ErrMachineAtCapacity, "machine m1"), codes.ResourceExhausted},
		{"timeout out of range", wrapf(ErrRequestTimeoutOutOfRange, "7200s"), codes.OutOfRange},
		{"expired replay window", &RequestStreamExpiredError{}, codes.OutOfRange},
		{"machine credential rejected", wrapf(ErrMachineCredentialRejected, "machine m1"), codes.PermissionDenied},
		{"client disconnected", wrapf(ErrClientDisconnected, "mid-stream"), codes.Canceled},
		{"stream send failed", wrapf(ErrStreamSendFailed, "%v", errors.New("broken pipe")), codes.Internal},
		{"principal required", ErrPrincipalRequired, codes.PermissionDenied},
		{"canceled context keeps its code", fmt.Errorf("drain machine: %w", context.Canceled), codes.Canceled},
		{"unmapped fails closed as internal", errors.New("persist claim failed: connection refused"), codes.Internal},
		{"nil is nil", nil, codes.OK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := statusFromDomainError("unit test", tc.err)
			if tc.err == nil {
				if got != nil {
					t.Fatalf("nil error mapped to %v, want nil", got)
				}
				return
			}
			if status.Code(got) != tc.want {
				t.Fatalf("mapping for %v = %v, want %v", tc.err, status.Code(got), tc.want)
			}
		})
	}
}

// newTaxonomyTestServer wires a full in-memory service stack.
func newTaxonomyTestServer(t *testing.T) (*GRPCServer, *SessionsService, *MachinesService, *RequestsService, *ToolService, string) {
	t.Helper()
	tracer := trace.NopTracer()
	toolService := NewToolService(tracer, nil)
	machineService := NewMachinesService(context.Background(), toolService, tracer, nil)
	requestService := NewRequestsService(context.Background(), toolService, machineService, tracer, nil)
	sessionService := NewSessionsService(tracer, nil)
	tasksService := NewTasksService(context.Background(), toolService, machineService, requestService, tracer, nil)

	server := NewGRPCServer(toolService, sessionService, machineService, requestService, tasksService)

	machine, err := machineService.RegisterMachine("sess-taxonomy", "", "test-sdk", "go", "127.0.0.1", []*model.Tool{
		{Name: "echo", Description: "echo", Schema: "{}"},
	}, "")
	if err != nil {
		t.Fatalf("register machine: %v", err)
	}
	if _, err := toolService.RegisterTool("sess-taxonomy", machine.ID, "echo", "echo tool", "{}", nil, nil); err != nil {
		t.Fatalf("register tool: %v", err)
	}
	return server, sessionService, machineService, requestService, toolService, machine.ID
}

func TestClaimRequestErrorCodeTaxonomy(t *testing.T) {
	server, _, machineService, requestService, _, machineID := newTaxonomyTestServer(t)

	request, err := requestService.CreateRequest("sess-taxonomy", "echo", `{}`, 0, "")
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	// A missing request is NOT_FOUND.
	if _, err := server.ClaimRequest(context.Background(), &proto.ClaimRequestRequest{
		SessionId: "sess-taxonomy", RequestId: "req-missing", MachineId: machineID,
	}); status.Code(err) != codes.NotFound {
		t.Fatalf("claim missing = %v, want not found", err)
	}

	// First claim succeeds and mints the lease.
	if _, err := server.ClaimRequest(context.Background(), &proto.ClaimRequestRequest{
		SessionId: "sess-taxonomy", RequestId: request.ID, MachineId: machineID,
	}); err != nil {
		t.Fatalf("first claim: %v", err)
	}

	// A second claim of the same (now claimed) request is a state conflict,
	// not a missing entity.
	if _, err := server.ClaimRequest(context.Background(), &proto.ClaimRequestRequest{
		SessionId: "sess-taxonomy", RequestId: request.ID, MachineId: "machine-rival",
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("claim contention = %v, want failed precondition", err)
	}

	// A draining machine refuses work with a precondition error, not a miss.
	// getOrCreateDrainState marks the machine draining without starting the
	// async unregister, keeping the state deterministic for the assertion.
	machineService.getOrCreateDrainState("sess-taxonomy", machineID)
	other, err := requestService.CreateRequest("sess-taxonomy", "echo", `{}`, 0, "")
	if err != nil {
		t.Fatalf("create second request: %v", err)
	}
	if _, err := server.ClaimRequest(context.Background(), &proto.ClaimRequestRequest{
		SessionId: "sess-taxonomy", RequestId: other.ID, MachineId: machineID,
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("claim while draining = %v, want failed precondition", err)
	}
}

func TestCancelRequestErrorCodeTaxonomy(t *testing.T) {
	server, _, _, requestService, _, _ := newTaxonomyTestServer(t)

	request, err := requestService.CreateRequest("sess-taxonomy", "echo", `{}`, 0, "")
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	if _, err := server.CancelRequest(context.Background(), &proto.CancelRequestRequest{
		SessionId: "sess-taxonomy", RequestId: "req-missing",
	}); status.Code(err) != codes.NotFound {
		t.Fatalf("cancel missing = %v, want not found", err)
	}

	if _, err := server.CancelRequest(context.Background(), &proto.CancelRequestRequest{
		SessionId: "sess-taxonomy", RequestId: request.ID,
	}); err != nil {
		t.Fatalf("cancel pending: %v", err)
	}

	// Cancelling an already-terminal request reports the state conflict.
	if _, err := server.CancelRequest(context.Background(), &proto.CancelRequestRequest{
		SessionId: "sess-taxonomy", RequestId: request.ID,
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("cancel terminal = %v, want failed precondition", err)
	}
}

func TestCreateRequestErrorCodeTaxonomy(t *testing.T) {
	server, _, _, _, toolService, _ := newTaxonomyTestServer(t)

	// Unknown tool: NOT_FOUND.
	if _, err := server.CreateRequest(context.Background(), &proto.CreateRequestRequest{
		SessionId: "sess-taxonomy", ToolName: "ghost-tool",
	}); status.Code(err) != codes.NotFound {
		t.Fatalf("create for unknown tool = %v, want not found", err)
	}

	// Over-maximum timeout: OUT_OF_RANGE.
	if _, err := server.CreateRequest(context.Background(), &proto.CreateRequestRequest{
		SessionId: "sess-taxonomy", ToolName: "echo", TimeoutSeconds: int32(maxRequestTimeout.Seconds()) + 1,
	}); status.Code(err) != codes.OutOfRange {
		t.Fatalf("create with over-range timeout = %v, want out of range", err)
	}

	// Tool registered but its provider machine gone: no provider available.
	if _, err := toolService.RegisterTool("sess-taxonomy", "machine-ghost", "orphan", "orphan tool", "{}", nil, nil); err != nil {
		t.Fatalf("register orphan tool: %v", err)
	}
	if _, err := server.CreateRequest(context.Background(), &proto.CreateRequestRequest{
		SessionId: "sess-taxonomy", ToolName: "orphan",
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("create without provider = %v, want failed precondition", err)
	}
}

// failingSendStream rejects every chunk, driving the send-failure path.
type failingSendStream struct {
	*collectingExecuteToolStream
}

func (f *failingSendStream) Send(*proto.ExecuteToolChunk) error {
	return errors.New("broken pipe")
}

func TestStreamHandlerErrorCodeTaxonomy(t *testing.T) {
	server, requestService, sessionID := newRequestStreamTestServer(t)

	// A stream caller that goes away mid-flight reports CANCELED; the work
	// itself keeps running and is reattachable via ResumeStream.
	liveCtx, cancelLive := context.WithCancel(fixedPrincipalContext())
	defer cancelLive()
	time.AfterFunc(50*time.Millisecond, cancelLive)
	if err := server.StreamExecuteTool(&proto.ExecuteToolRequest{
		SessionId: sessionID, ToolName: "echo", Input: `{}`,
	}, newCollectingExecuteToolStream(liveCtx)); status.Code(err) != codes.Canceled {
		t.Fatalf("client disconnect = %v, want canceled", err)
	}

	// A resumed stream reports the same condition when its caller leaves.
	request, err := requestService.CreateRequest(sessionID, "echo", `{}`, 0, "")
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	claimed := claimForStreamTest(t, requestService, sessionID, request.ID)
	if err := requestService.AppendRequestChunks(sessionID, request.ID, claimed.LeasedBy, claimed.LeaseEpoch, []string{"alpha"}, model.ResultTypeStreaming); err != nil {
		t.Fatalf("append chunks: %v", err)
	}
	resumeCtx, cancelResume := context.WithCancel(fixedPrincipalContext())
	defer cancelResume()
	time.AfterFunc(50*time.Millisecond, cancelResume)
	if err := server.ResumeStream(&proto.ResumeStreamRequest{RequestId: request.ID}, newCollectingExecuteToolStream(resumeCtx)); status.Code(err) != codes.Canceled {
		t.Fatalf("resume client disconnect = %v, want canceled", err)
	}

	// A caller whose transport cannot receive chunks reports INTERNAL.
	if err := server.ResumeStream(&proto.ResumeStreamRequest{RequestId: request.ID},
		&failingSendStream{collectingExecuteToolStream: newCollectingExecuteToolStream(fixedPrincipalContext())}); status.Code(err) != codes.Internal {
		t.Fatalf("send failure = %v, want internal", err)
	}
}

func TestResumeStreamFailsClosedWithoutPrincipal(t *testing.T) {
	server, _, _ := newRequestStreamTestServer(t)

	// No principal in the context: PERMISSION_DENIED before any lookup, so
	// request existence is not leaked to an unauthenticated caller.
	stream := newCollectingExecuteToolStream(context.Background())
	if err := server.ResumeStream(&proto.ResumeStreamRequest{RequestId: "req-any"}, stream); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("resume without principal = %v, want permission denied", err)
	}
}

func TestResumeStreamHidesCrossSessionExistence(t *testing.T) {
	server, requestService, sessionID := newRequestStreamTestServer(t)

	request, err := requestService.CreateRequest(sessionID, "echo", `{}`, 0, "")
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	// A session-bound principal must get NotFound for a missing request and
	// for a request in another session, with messages that differ only by
	// the caller-supplied id — no hint that the second one exists.
	otherSession := auth.NewContext(context.Background(), &model.AuthPrincipal{Mode: model.AuthModeSessionKey, SessionID: "session-other"})

	missing := server.ResumeStream(&proto.ResumeStreamRequest{RequestId: "req-missing"}, newCollectingExecuteToolStream(otherSession))
	mismatch := server.ResumeStream(&proto.ResumeStreamRequest{RequestId: request.ID}, newCollectingExecuteToolStream(otherSession))

	if status.Code(missing) != codes.NotFound || status.Code(mismatch) != codes.NotFound {
		t.Fatalf("missing=%v mismatch=%v, want both not found", missing, mismatch)
	}
	wantShape := "failed to resolve request: not found: "
	if msg := status.Convert(missing).Message(); !strings.HasPrefix(msg, wantShape+"req-missing") {
		t.Fatalf("missing message = %q, want shape %q<id>", msg, wantShape)
	}
	if msg := status.Convert(mismatch).Message(); !strings.HasPrefix(msg, wantShape+request.ID) {
		t.Fatalf("mismatch message = %q, want shape %q<id> with no extra detail", msg, wantShape)
	}
}

func TestListHandlerErrorCodeTaxonomy(t *testing.T) {
	server, sessionService, _, _, _, _ := newTaxonomyTestServer(t)

	// A malformed page token is invalid caller input on every list that
	// accepts one.
	for _, handler := range []struct {
		name string
		call func() error
	}{
		{"list user sessions", func() error {
			_, err := server.ListUserSessions(context.Background(), &proto.ListUserSessionsRequest{UserId: "u", PageToken: "!!not-base64!!"})
			return err
		}},
		{"list requests", func() error {
			_, err := server.ListRequests(context.Background(), &proto.ListRequestsRequest{SessionId: "sess-taxonomy", PageToken: "!!not-base64!!"})
			return err
		}},
	} {
		if err := handler.call(); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("%s with malformed page token = %v, want invalid argument", handler.name, err)
		}
	}

	// A create collision is bare AlreadyExists — no existing-session
	// payload leaks back to the caller.
	if _, err := sessionService.CreateSession("user", "dup", "", "sess-dup", ""); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	_, err := server.CreateSession(context.Background(), &proto.CreateSessionRequest{UserId: "user", Name: "dup", SessionId: "sess-dup"})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate create = %v, want already exists", err)
	}
	if msg := status.Convert(err).Message(); strings.Contains(msg, "\"id\"") || strings.Contains(msg, "created_at") {
		t.Fatalf("collision response leaks session payload: %q", msg)
	}
}

func TestLookupHandlersReturnNotFound(t *testing.T) {
	server, _, _, _, _, _ := newTaxonomyTestServer(t)

	if _, err := server.GetRequest(context.Background(), &proto.GetRequestRequest{
		SessionId: "sess-taxonomy", RequestId: "req-missing",
	}); status.Code(err) != codes.NotFound {
		t.Fatalf("get missing request = %v, want not found", err)
	}

	if _, err := server.GetMachine(context.Background(), &proto.GetMachineRequest{
		SessionId: "sess-taxonomy", MachineId: "machine-missing",
	}); status.Code(err) != codes.NotFound {
		t.Fatalf("get missing machine = %v, want not found", err)
	}

	if _, err := server.GetToolByName(context.Background(), &proto.GetToolByNameRequest{
		SessionId: "sess-taxonomy", ToolName: "ghost",
	}); status.Code(err) != codes.NotFound {
		t.Fatalf("get missing tool = %v, want not found", err)
	}

	if _, err := server.GetTask(context.Background(), &proto.GetTaskRequest{
		SessionId: "sess-taxonomy", TaskId: "task-missing",
	}); status.Code(err) != codes.NotFound {
		t.Fatalf("get missing task = %v, want not found", err)
	}
}

// TestStatusCarriesMachineReadableReason pins the typed-error contract: the
// overloaded OUT_OF_RANGE conditions carry distinct ErrorInfo reasons, and
// capacity rejections carry RetryInfo, so clients branch on semantics rather
// than parsing error strings.
func TestStatusCarriesMachineReadableReason(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantCode   codes.Code
		wantReason string
	}{
		{"timeout above max", wrapf(ErrRequestTimeoutOutOfRange, "7200s"), codes.OutOfRange, ReasonTimeoutAboveMax},
		{"replay window expired", &RequestStreamExpiredError{}, codes.OutOfRange, ReasonReplayWindowExpired},
		{"machine at capacity", wrapf(ErrMachineAtCapacity, "machine m1"), codes.ResourceExhausted, ReasonCapacityExhausted},
		{"session backlog full", wrapf(ErrTooManyPendingRequests, "session s1"), codes.ResourceExhausted, ReasonSessionBacklogFull},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st, ok := status.FromError(statusFromDomainError("unit test", tc.err))
			if !ok {
				t.Fatal("expected a gRPC status")
			}
			foundReason := false
			foundRetry := tc.wantCode != codes.ResourceExhausted
			for _, detail := range st.Details() {
				switch info := detail.(type) {
				case *errdetails.ErrorInfo:
					if info.Reason == tc.wantReason {
						foundReason = true
					}
				case *errdetails.RetryInfo:
					foundRetry = true
				}
			}
			if !foundReason {
				t.Fatalf("status details missing reason %s: %+v", tc.wantReason, st.Details())
			}
			if !foundRetry {
				t.Fatal("capacity rejection missing RetryInfo")
			}
			if status.Code(st.Err()) != tc.wantCode {
				t.Fatalf("code=%v want %v", status.Code(st.Err()), tc.wantCode)
			}
		})
	}
}
