package auth

import (
	"context"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"toolplane/pkg/model"
	"toolplane/proto"
)

func TestNormalizeAPIKeyCapabilitiesSplitsLegacyExecute(t *testing.T) {
	capabilities, err := model.NormalizeAPIKeyCapabilities([]string{"read", "execute", "admin"})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	want := []model.APIKeyCapability{
		model.APIKeyCapabilityRead,
		model.APIKeyCapabilityInvoke,
		model.APIKeyCapabilityProvide,
		model.APIKeyCapabilityAdmin,
	}
	if len(capabilities) != len(want) {
		t.Fatalf("normalized = %v, want %v", capabilities, want)
	}
	for index := range want {
		if capabilities[index] != want[index] {
			t.Fatalf("normalized = %v, want %v", capabilities, want)
		}
	}
}

func TestMethodPoliciesPartitionByRole(t *testing.T) {
	provide := []string{
		"/api.v1.ToolService/RegisterTool",
		"/api.v1.ToolService/DeleteTool",
		"/api.v1.ToolService/UpdateToolPing",
		"/api.v1.MachinesService/RegisterMachine",
		"/api.v1.MachinesService/UpdateMachinePing",
		"/api.v1.MachinesService/UnregisterMachine",
		"/api.v1.MachinesService/DrainMachine",
		"/api.v1.RequestsService/UpdateRequest",
		"/api.v1.RequestsService/ClaimRequest",
		"/api.v1.RequestsService/SubmitRequestResult",
		"/api.v1.RequestsService/AppendRequestChunks",
		"/api.v1.RequestsService/RenewRequestLease",
	}
	invoke := []string{
		"/api.v1.ToolService/ExecuteTool",
		"/api.v1.ToolService/StreamExecuteTool",
		"/api.v1.ToolService/ResumeStream",
		"/api.v1.RequestsService/CreateRequest",
		"/api.v1.RequestsService/CancelRequest",
		"/api.v1.TasksService/CreateTask",
		"/api.v1.TasksService/CancelTask",
	}

	for _, method := range provide {
		policy, ok := methodPolicies[method]
		if !ok {
			t.Fatalf("%s has no policy", method)
		}
		if policy.Capability != model.APIKeyCapabilityProvide {
			t.Fatalf("%s capability = %s, want provide", method, policy.Capability)
		}
	}
	for _, method := range invoke {
		policy, ok := methodPolicies[method]
		if !ok {
			t.Fatalf("%s has no policy", method)
		}
		if policy.Capability != model.APIKeyCapabilityInvoke {
			t.Fatalf("%s capability = %s, want invoke", method, policy.Capability)
		}
	}
}

func TestLegacyExecuteKeyPassesInvokeAndProvide(t *testing.T) {
	principal := &model.AuthPrincipal{
		Mode:         model.AuthModeSessionKey,
		SessionID:    "session-1",
		Capabilities: []model.APIKeyCapability{model.APIKeyCapabilityRead, model.APIKeyCapabilityInvoke, model.APIKeyCapabilityProvide},
	}
	if !principal.HasCapability(model.APIKeyCapabilityInvoke) {
		t.Fatal("legacy execute key should satisfy invoke")
	}
	if !principal.HasCapability(model.APIKeyCapabilityProvide) {
		t.Fatal("legacy execute key should satisfy provide")
	}
}

// machineAuthHarness drives the interceptor with a session-key principal and
// a captured validator, mirroring the production unary path.
type machineAuthHarness struct {
	authorizer *APIKeyAuthorizer
	validator  func(sessionID, machineID, token string) error
}

func newMachineAuthHarness(t *testing.T, validator func(sessionID, machineID, token string) error) *machineAuthHarness {
	t.Helper()
	authorizer := NewAPIKeyAuthorizer(
		func(_ context.Context, _ string) (*model.AuthPrincipal, error) {
			return &model.AuthPrincipal{
				Mode:         model.AuthModeSessionKey,
				SessionID:    "session-1",
				Capabilities: []model.APIKeyCapability{model.APIKeyCapabilityProvide},
			}, nil
		},
		nil,
		WithMachineTokenAuth(func(sessionID, machineID, token string) error {
			return validator(sessionID, machineID, token)
		}),
	)
	return &machineAuthHarness{authorizer: authorizer, validator: validator}
}

func (h *machineAuthHarness) call(t *testing.T, ctx context.Context, req interface{}) error {
	t.Helper()
	_, err := h.authorizer.UnaryInterceptor()(
		ctx,
		req,
		&grpc.UnaryServerInfo{FullMethod: "/api.v1.MachinesService/UpdateMachinePing"},
		func(ctx context.Context, _ interface{}) (interface{}, error) { return nil, nil },
	)
	return err
}

func TestMachineTokenGate(t *testing.T) {
	var seenToken string
	harness := newMachineAuthHarness(t, func(_, _, token string) error {
		seenToken = token
		if token != "secret-machine-token" {
			return errMachineTokenRejected
		}
		return nil
	})

	req := &proto.UpdateMachinePingRequest{SessionId: "session-1", MachineId: "machine-9"}

	// The auth layer requires incoming metadata; the machine credential rides
	// alongside the API key.
	authCtx := func(extra ...string) context.Context {
		pairs := append([]string{"api_key", "session-key"}, extra...)
		return metadata.NewIncomingContext(context.Background(), metadata.Pairs(pairs...))
	}

	// No credential presented.
	if err := harness.call(t, authCtx(), req); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("no token: err = %v, want permission denied", err)
	}

	// Wrong credential.
	ctxWrong := authCtx("x-toolplane-machine-token", "wrong")
	if err := harness.call(t, ctxWrong, req); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("wrong token: err = %v, want permission denied", err)
	}

	// Correct credential.
	ctxRight := authCtx("x-toolplane-machine-token", "secret-machine-token")
	if err := harness.call(t, ctxRight, req); err != nil {
		t.Fatalf("right token: %v", err)
	}
	if seenToken != "secret-machine-token" {
		t.Fatalf("validator saw token %q", seenToken)
	}
}

func TestMachineTokenGateSkipsFixedModeAndNonProvideRPCs(t *testing.T) {
	called := false
	authorizer := NewAPIKeyAuthorizer(
		func(_ context.Context, _ string) (*model.AuthPrincipal, error) {
			return &model.AuthPrincipal{Mode: model.AuthModeFixed}, nil
		},
		nil,
		WithMachineTokenAuth(func(string, string, string) error {
			called = true
			return nil
		}),
	)
	req := &proto.UpdateMachinePingRequest{SessionId: "session-1", MachineId: "machine-9"}
	fixedCtx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("api_key", "fixed-dev-key"))
	_, err := authorizer.UnaryInterceptor()(
		fixedCtx,
		req,
		&grpc.UnaryServerInfo{FullMethod: "/api.v1.MachinesService/UpdateMachinePing"},
		func(ctx context.Context, _ interface{}) (interface{}, error) { return nil, nil },
	)
	if err != nil {
		t.Fatalf("fixed mode provide call: %v", err)
	}
	if called {
		t.Fatal("machine token validator must not run in fixed mode")
	}

	// Session-key mode on an invoke-scoped RPC: no machine token required.
	sessionAuthorizer := NewAPIKeyAuthorizer(
		func(_ context.Context, _ string) (*model.AuthPrincipal, error) {
			return &model.AuthPrincipal{
				Mode:         model.AuthModeSessionKey,
				SessionID:    "session-1",
				Capabilities: []model.APIKeyCapability{model.APIKeyCapabilityInvoke},
			}, nil
		},
		nil,
		WithMachineTokenAuth(func(string, string, string) error {
			called = true
			return nil
		}),
	)
	_, err = sessionAuthorizer.UnaryInterceptor()(
		metadata.NewIncomingContext(context.Background(), metadata.Pairs("api_key", "session-key")),
		&proto.CreateRequestRequest{SessionId: "session-1", ToolName: "echo"},
		&grpc.UnaryServerInfo{FullMethod: "/api.v1.RequestsService/CreateRequest"},
		func(ctx context.Context, _ interface{}) (interface{}, error) { return nil, nil },
	)
	if err != nil {
		t.Fatalf("invoke call without machine token: %v", err)
	}
}

var errMachineTokenRejected = status.Error(codes.PermissionDenied, "rejected")

func TestMachineTokenMetadataKeyLowercase(t *testing.T) {
	if strings.ToLower(MachineTokenMetadataKey) != MachineTokenMetadataKey {
		t.Fatalf("metadata key must be lowercase, got %q", MachineTokenMetadataKey)
	}
}
