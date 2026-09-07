package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	request, err := requestService.CreateRequest("sess-taxonomy", "echo", `{}`, 0)
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
	other, err := requestService.CreateRequest("sess-taxonomy", "echo", `{}`, 0)
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

	request, err := requestService.CreateRequest("sess-taxonomy", "echo", `{}`, 0)
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
