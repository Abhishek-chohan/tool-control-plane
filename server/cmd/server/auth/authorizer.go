package auth

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"toolplane/pkg/model"
	"toolplane/pkg/trace"
)

type AuthenticateFunc func(context.Context, string) (*model.AuthPrincipal, error)

type contextKey string

const authPrincipalContextKey contextKey = "toolplane.auth.principal"

type MethodPolicy struct {
	Capability  model.APIKeyCapability
	BindSession bool
	BindUser    bool
	// Public marks unauthenticated surface: standard infrastructure
	// probes that must answer before any credential exists (grpc.health.v1).
	Public bool
}

type sessionScopedRequest interface {
	GetSessionId() string
}

type userScopedRequest interface {
	GetUserId() string
}

type APIKeyAuthorizer struct {
	authenticate     AuthenticateFunc
	tracer           trace.SessionTracer
	machineTokenAuth MachineTokenAuthorizer
}

type authorizerOption func(*APIKeyAuthorizer)

// NewAPIKeyAuthorizer builds the authorizer; options (e.g.
// WithMachineTokenAuth) extend it.
func NewAPIKeyAuthorizer(authenticate AuthenticateFunc, tracer trace.SessionTracer, opts ...authorizerOption) *APIKeyAuthorizer {
	if tracer == nil {
		tracer = trace.NopTracer()
	}
	authorizer := &APIKeyAuthorizer{authenticate: authenticate, tracer: tracer}
	for _, opt := range opts {
		opt(authorizer)
	}
	return authorizer
}

func PrincipalFromContext(ctx context.Context) (*model.AuthPrincipal, bool) {
	principal, ok := ctx.Value(authPrincipalContextKey).(*model.AuthPrincipal)
	return principal, ok
}

// NewContext returns a context carrying the authenticated principal. The
// interceptors use it after successful authentication; tests that drive
// handlers directly can use it to inject a principal.
func NewContext(ctx context.Context, principal *model.AuthPrincipal) context.Context {
	return context.WithValue(ctx, authPrincipalContextKey, principal)
}

func RequireSessionCapability(ctx context.Context, sessionID string, capability model.APIKeyCapability) error {
	principal, ok := PrincipalFromContext(ctx)
	if !ok || principal == nil {
		// Fail closed: a handler reached without an authenticated principal is
		// a wiring bug, not an anonymous-access grant.
		return status.Error(codes.PermissionDenied, "missing authenticated principal")
	}
	if principal.Mode == model.AuthModeFixed {
		return nil
	}
	if capability != "" && !principal.HasCapability(capability) {
		return status.Errorf(codes.PermissionDenied, "api key does not have %s capability", capability)
	}
	if sessionID != "" && principal.SessionID != sessionID {
		return status.Errorf(codes.PermissionDenied, "api key is not authorized for session %s", sessionID)
	}
	return nil
}

func RequireUserCapability(ctx context.Context, userID string, capability model.APIKeyCapability) error {
	principal, ok := PrincipalFromContext(ctx)
	if !ok || principal == nil {
		return status.Error(codes.PermissionDenied, "missing authenticated principal")
	}
	if principal.Mode == model.AuthModeFixed {
		return nil
	}
	if capability != "" && !principal.HasCapability(capability) {
		return status.Errorf(codes.PermissionDenied, "api key does not have %s capability", capability)
	}
	if userID != "" {
		if principal.UserID == "" {
			return status.Errorf(codes.PermissionDenied, "api key is missing user scope for user %s", userID)
		}
		if principal.UserID != userID {
			return status.Errorf(codes.PermissionDenied, "api key is not authorized for user %s", userID)
		}
	}
	return nil
}

func (a *APIKeyAuthorizer) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if policy, ok := methodPolicyFor(info.FullMethod); ok && policy.Public {
			// Infrastructure probes carry no credentials by design.
			ctx = NewContext(ctx, anonymousPrincipal())
			return handler(ctx, req)
		}

		principal, authErr := a.authenticateRequest(ctx)
		if authErr != nil {
			a.recordRejected("", info.FullMethod, "authentication_failed", authErr.Error(), redactTokenFromContext(ctx))
			return nil, authErr
		}

		if authzErr := a.authorizeUnary(principal, info.FullMethod, req); authzErr != nil {
			a.recordDenied(principal, info.FullMethod, authzErr.Error())
			return nil, authzErr
		}

		if policy, ok := methodPolicyFor(info.FullMethod); ok {
			if machineErr := a.authorizeMachineToken(ctx, principal, req, policy); machineErr != nil {
				a.recordDenied(principal, info.FullMethod, machineErr.Error())
				return nil, machineErr
			}
		}

		ctx = NewContext(ctx, principal)
		a.recordValidated(principal, info.FullMethod, "unary")
		return handler(ctx, req)
	}
}

func (a *APIKeyAuthorizer) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if policy, ok := methodPolicyFor(info.FullMethod); ok && policy.Public {
			return handler(srv, &wrappedServerStream{ServerStream: ss, ctx: NewContext(ss.Context(), anonymousPrincipal())})
		}

		principal, authErr := a.authenticateRequest(ss.Context())
		if authErr != nil {
			a.recordRejected("", info.FullMethod, "authentication_failed", authErr.Error(), redactTokenFromContext(ss.Context()))
			return authErr
		}

		policy, ok := methodPolicyFor(info.FullMethod)
		if !ok {
			denyErr := status.Errorf(codes.PermissionDenied, "no auth policy configured for %s", info.FullMethod)
			a.recordDenied(principal, info.FullMethod, denyErr.Error())
			return denyErr
		}
		if policy.Capability != "" && !principal.HasCapability(policy.Capability) {
			denyErr := status.Errorf(codes.PermissionDenied, "api key does not have %s capability", policy.Capability)
			a.recordDenied(principal, info.FullMethod, denyErr.Error())
			return denyErr
		}

		wrapped := &principalServerStream{
			ServerStream: ss,
			ctx:          NewContext(ss.Context(), principal),
		}
		a.recordValidated(principal, info.FullMethod, "stream")
		return handler(srv, wrapped)
	}
}

type principalServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *principalServerStream) Context() context.Context {
	return s.ctx
}

// AnonymousUnaryInterceptor / AnonymousStreamInterceptor attach a fixed-mode
// principal to every call without authenticating. They are used only when
// TOOLPLANE_AUTH_MODE=disabled (a development convenience that production
// refuses to boot with), so handler-level principal checks — which fail
// closed when no principal is present — keep working in that mode.
func AnonymousUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(NewContext(ctx, anonymousPrincipal()), req)
	}
}

func AnonymousStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapped := &principalServerStream{
			ServerStream: ss,
			ctx:          NewContext(ss.Context(), anonymousPrincipal()),
		}
		return handler(srv, wrapped)
	}
}

func anonymousPrincipal() *model.AuthPrincipal {
	return &model.AuthPrincipal{
		Mode:         model.AuthModeFixed,
		Capabilities: model.DefaultAPIKeyCapabilities(),
		TokenPreview: "<auth-disabled>",
	}
}

func (a *APIKeyAuthorizer) authenticateRequest(ctx context.Context) (*model.AuthPrincipal, error) {
	if a == nil || a.authenticate == nil {
		return nil, status.Error(codes.Unauthenticated, "api key authentication is not configured")
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}
	token, err := tokenFromMetadata(md)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	principal, authErr := a.authenticate(ctx, token)
	if authErr != nil {
		return nil, status.Error(codes.Unauthenticated, authErr.Error())
	}
	if principal == nil {
		return nil, status.Error(codes.Unauthenticated, "api key authentication returned no principal")
	}
	return principal, nil
}

func (a *APIKeyAuthorizer) authorizeUnary(principal *model.AuthPrincipal, fullMethod string, req interface{}) error {
	policy, ok := methodPolicyFor(fullMethod)
	if !ok {
		return status.Errorf(codes.PermissionDenied, "no auth policy configured for %s", fullMethod)
	}
	if principal.Mode == model.AuthModeFixed {
		return nil
	}
	if policy.Capability != "" && !principal.HasCapability(policy.Capability) {
		return status.Errorf(codes.PermissionDenied, "api key does not have %s capability", policy.Capability)
	}
	if policy.BindSession {
		sessionRequest, ok := req.(sessionScopedRequest)
		if !ok {
			return status.Errorf(codes.PermissionDenied, "%s requires session-scoped authorization", fullMethod)
		}
		targetSessionID := sessionRequest.GetSessionId()
		if targetSessionID != "" && targetSessionID != principal.SessionID {
			return status.Errorf(codes.PermissionDenied, "api key is not authorized for session %s", targetSessionID)
		}
	}
	if policy.BindUser {
		userRequest, ok := req.(userScopedRequest)
		if !ok {
			return status.Errorf(codes.PermissionDenied, "%s requires user-scoped authorization", fullMethod)
		}
		targetUserID := userRequest.GetUserId()
		if targetUserID != "" {
			if principal.UserID == "" {
				return status.Errorf(codes.PermissionDenied, "api key is missing user scope for user %s", targetUserID)
			}
			if targetUserID != principal.UserID {
				return status.Errorf(codes.PermissionDenied, "api key is not authorized for user %s", targetUserID)
			}
		}
	}
	return nil
}

func redactTokenFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "<missing>"
	}
	token, err := tokenFromMetadata(md)
	if err != nil {
		return "<missing>"
	}
	return redactToken(normalizeToken(token))
}

func (a *APIKeyAuthorizer) recordValidated(principal *model.AuthPrincipal, fullMethod, phase string) {
	if a == nil || a.tracer == nil || principal == nil {
		return
	}
	a.tracer.Record(trace.SessionEvent{
		SessionID: principal.SessionID,
		Event:     trace.EventAuthValidated,
		Timestamp: time.Now(),
		Metadata: map[string]any{
			"method":       fullMethod,
			"phase":        phase,
			"keyId":        principal.KeyID,
			"tokenPreview": principal.TokenPreview,
			"capabilities": model.CapabilityStrings(principal.Capabilities),
		},
	})
}

func (a *APIKeyAuthorizer) recordRejected(sessionID, fullMethod, reason, detail, tokenPreview string) {
	if a == nil || a.tracer == nil {
		return
	}
	a.tracer.Record(trace.SessionEvent{
		SessionID: sessionID,
		Event:     trace.EventAuthRejected,
		Timestamp: time.Now(),
		Metadata: map[string]any{
			"method":       fullMethod,
			"reason":       reason,
			"detail":       detail,
			"tokenPreview": tokenPreview,
		},
	})
}

func (a *APIKeyAuthorizer) recordDenied(principal *model.AuthPrincipal, fullMethod, detail string) {
	if a == nil || a.tracer == nil {
		return
	}
	sessionID := ""
	tokenPreview := ""
	keyID := ""
	capabilities := []string{}
	if principal != nil {
		sessionID = principal.SessionID
		tokenPreview = principal.TokenPreview
		keyID = principal.KeyID
		capabilities = model.CapabilityStrings(principal.Capabilities)
	}
	a.tracer.Record(trace.SessionEvent{
		SessionID: sessionID,
		Event:     trace.EventAuthPolicyDenied,
		Timestamp: time.Now(),
		Metadata: map[string]any{
			"method":       fullMethod,
			"detail":       detail,
			"keyId":        keyID,
			"tokenPreview": tokenPreview,
			"capabilities": capabilities,
		},
	})
}

func methodPolicyFor(fullMethod string) (MethodPolicy, bool) {
	policy, ok := methodPolicies[fullMethod]
	return policy, ok
}

// methodPolicies partitions the contract by role:
//   - read:    inspection RPCs (lists, gets, chunk windows, health)
//   - invoke:  consumer RPCs (create requests, execute tools, cancel work)
//   - provide: provider RPCs (register machines/tools, claim, submit, append,
//     renew lease, drain, unregister) — these also require the per-machine
//     token when the session-key auth mode is active
//   - admin:   session/key administration
//
// The legacy "execute" capability satisfies both invoke and provide (it
// normalizes to invoke+provide), so pre-split keys keep working.
var methodPolicies = map[string]MethodPolicy{
	"/grpc.health.v1.Health/Check":                {Public: true},
	"/grpc.health.v1.Health/Watch":                {Public: true},
	"/api.v1.ToolService/RegisterTool":            {Capability: model.APIKeyCapabilityProvide, BindSession: true},
	"/api.v1.ToolService/ListTools":               {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.v1.ToolService/GetTool":                 {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.v1.ToolService/GetToolById":             {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.v1.ToolService/GetToolByName":           {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.v1.ToolService/DeleteTool":              {Capability: model.APIKeyCapabilityProvide, BindSession: true},
	"/api.v1.ToolService/UpdateToolPing":          {Capability: model.APIKeyCapabilityProvide, BindSession: true},
	"/api.v1.ToolService/InvokeTool":              {Capability: model.APIKeyCapabilityInvoke, BindSession: true},
	"/api.v1.ToolService/ExecuteTool":             {Capability: model.APIKeyCapabilityInvoke, BindSession: true},
	"/api.v1.ToolService/StreamExecuteTool":       {Capability: model.APIKeyCapabilityInvoke},
	"/api.v1.ToolService/ResumeStream":            {Capability: model.APIKeyCapabilityInvoke},
	"/api.v1.ToolService/HealthCheck":             {Capability: model.APIKeyCapabilityRead},
	"/api.v1.SessionsService/CreateSession":       {Capability: model.APIKeyCapabilityAdmin, BindUser: true},
	"/api.v1.SessionsService/GetSession":          {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.v1.SessionsService/ListSessions":        {Capability: model.APIKeyCapabilityAdmin, BindUser: true},
	"/api.v1.SessionsService/UpdateSession":       {Capability: model.APIKeyCapabilityAdmin, BindSession: true},
	"/api.v1.SessionsService/DeleteSession":       {Capability: model.APIKeyCapabilityAdmin, BindSession: true},
	"/api.v1.SessionsService/ListUserSessions":    {Capability: model.APIKeyCapabilityAdmin, BindUser: true},
	"/api.v1.SessionsService/BulkDeleteSessions":  {Capability: model.APIKeyCapabilityAdmin, BindUser: true},
	"/api.v1.SessionsService/GetSessionStats":     {Capability: model.APIKeyCapabilityAdmin, BindUser: true},
	"/api.v1.SessionsService/InvalidateSession":   {Capability: model.APIKeyCapabilityAdmin, BindSession: true},
	"/api.v1.SessionsService/CreateApiKey":        {Capability: model.APIKeyCapabilityAdmin, BindSession: true},
	"/api.v1.SessionsService/ListApiKeys":         {Capability: model.APIKeyCapabilityAdmin, BindSession: true},
	"/api.v1.SessionsService/RevokeApiKey":        {Capability: model.APIKeyCapabilityAdmin, BindSession: true},
	"/api.v1.MachinesService/RegisterMachine":     {Capability: model.APIKeyCapabilityProvide, BindSession: true},
	"/api.v1.MachinesService/ListMachines":        {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.v1.MachinesService/GetMachine":          {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.v1.MachinesService/UpdateMachinePing":   {Capability: model.APIKeyCapabilityProvide, BindSession: true},
	"/api.v1.MachinesService/UnregisterMachine":   {Capability: model.APIKeyCapabilityProvide, BindSession: true},
	"/api.v1.MachinesService/DrainMachine":        {Capability: model.APIKeyCapabilityProvide, BindSession: true},
	"/api.v1.RequestsService/CreateRequest":       {Capability: model.APIKeyCapabilityInvoke, BindSession: true},
	"/api.v1.RequestsService/GetRequest":          {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.v1.RequestsService/ListRequests":        {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.v1.RequestsService/UpdateRequest":       {Capability: model.APIKeyCapabilityProvide, BindSession: true},
	"/api.v1.RequestsService/ClaimRequest":        {Capability: model.APIKeyCapabilityProvide, BindSession: true},
	"/api.v1.RequestsService/CancelRequest":       {Capability: model.APIKeyCapabilityInvoke, BindSession: true},
	"/api.v1.RequestsService/SubmitRequestResult": {Capability: model.APIKeyCapabilityProvide, BindSession: true},
	"/api.v1.RequestsService/AppendRequestChunks": {Capability: model.APIKeyCapabilityProvide, BindSession: true},
	"/api.v1.RequestsService/GetRequestChunks":    {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.v1.RequestsService/RenewRequestLease":   {Capability: model.APIKeyCapabilityProvide, BindSession: true},
	"/api.v1.TasksService/CreateTask":             {Capability: model.APIKeyCapabilityInvoke, BindSession: true},
	"/api.v1.TasksService/GetTask":                {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.v1.TasksService/ListTasks":              {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.v1.TasksService/CancelTask":             {Capability: model.APIKeyCapabilityInvoke, BindSession: true},
}

func DebugPrincipal(principal *model.AuthPrincipal) string {
	if principal == nil {
		return "<nil>"
	}
	return fmt.Sprintf("mode=%s session=%s user=%s key=%s caps=%v", principal.Mode, principal.SessionID, principal.UserID, principal.KeyID, model.CapabilityStrings(principal.Capabilities))
}

// wrappedServerStream overrides the context of a server stream so public
// methods can run with an anonymous principal without re-authenticating.
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context { return w.ctx }
