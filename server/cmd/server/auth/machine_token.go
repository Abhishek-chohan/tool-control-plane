package auth

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"toolplane/pkg/model"
)

// MachineTokenMetadataKey is the gRPC metadata key carrying the per-machine
// credential on provide-scoped RPCs. HTTP edges forward it as the canonical
// header X-Toolplane-Machine-Token.
const MachineTokenMetadataKey = "x-toolplane-machine-token" // #nosec G101 -- well-known header name, not a secret

// MachineTokenFromContext extracts the presented per-machine credential, if
// any. Exported so handlers that need the credential for their own checks
// (RegisterMachine's takeover guard) read the same source as the interceptor.
func MachineTokenFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get(MachineTokenMetadataKey)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// machineScopedRequest is implemented by every provide-scoped request message
// (they all carry session_id and machine_id).
type machineScopedRequest interface {
	GetSessionId() string
	GetMachineId() string
}

// MachineTokenAuthorizer validates a presented per-machine credential against
// the machine registry (implemented by MachinesService.AuthorizeMachineToken).
type MachineTokenAuthorizer func(sessionID, machineID, token string) error

// authorizeMachineToken enforces the per-machine credential on provide-scoped
// RPCs when session-key auth is active. Fixed mode (single trusted dev key)
// and auth-disabled mode are exempt: production refuses both, so the gate
// covers exactly the multi-credential deployments where hijacking matters.
func (a *APIKeyAuthorizer) authorizeMachineToken(ctx context.Context, principal *model.AuthPrincipal, req interface{}, policy MethodPolicy) error {
	if a.machineTokenAuth == nil {
		return nil
	}
	if policy.Capability != model.APIKeyCapabilityProvide {
		return nil
	}
	if principal == nil || principal.Mode != model.AuthModeSessionKey {
		return nil
	}
	machineRequest, ok := req.(machineScopedRequest)
	if !ok {
		return nil
	}
	machineID := machineRequest.GetMachineId()
	if machineID == "" {
		return nil
	}
	sessionID := machineRequest.GetSessionId()
	if err := a.machineTokenAuth(sessionID, machineID, MachineTokenFromContext(ctx)); err != nil {
		return status.Errorf(codes.PermissionDenied, "machine credential rejected: %v", err)
	}
	return nil
}

// WithMachineTokenAuth wires the per-machine credential validator into the
// authorizer. Provide-scoped RPCs in session-key mode must present the
// credential minted at machine registration.
func WithMachineTokenAuth(authorizer MachineTokenAuthorizer) func(*APIKeyAuthorizer) {
	return func(a *APIKeyAuthorizer) {
		a.machineTokenAuth = authorizer
	}
}
