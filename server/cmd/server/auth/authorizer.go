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
}

type sessionScopedRequest interface {
	GetSessionId() string
}

type userScopedRequest interface {
	GetUserId() string
}

type APIKeyAuthorizer struct {
	authenticate AuthenticateFunc
	tracer       trace.SessionTracer
}

func NewAPIKeyAuthorizer(authenticate AuthenticateFunc, tracer trace.SessionTracer) *APIKeyAuthorizer {
	if tracer == nil {
		tracer = trace.NopTracer()
	}
	return &APIKeyAuthorizer{authenticate: authenticate, tracer: tracer}
}

func PrincipalFromContext(ctx context.Context) (*model.AuthPrincipal, bool) {
	principal, ok := ctx.Value(authPrincipalContextKey).(*model.AuthPrincipal)
	return principal, ok
}

func RequireSessionCapability(ctx context.Context, sessionID string, capability model.APIKeyCapability) error {
	principal, ok := PrincipalFromContext(ctx)
	if !ok || principal == nil {
		return nil
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
		return nil
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
		principal, authErr := a.authenticateRequest(ctx)
		if authErr != nil {
			a.recordRejected("", info.FullMethod, "authentication_failed", authErr.Error(), redactTokenFromContext(ctx))
			return nil, authErr
		}

		if authzErr := a.authorizeUnary(principal, info.FullMethod, req); authzErr != nil {
			a.recordDenied(principal, info.FullMethod, authzErr.Error())
			return nil, authzErr
		}

		ctx = context.WithValue(ctx, authPrincipalContextKey, principal)
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
			ctx:          context.WithValue(ss.Context(), authPrincipalContextKey, principal),
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

var methodPolicies = map[string]MethodPolicy{
	"/api.ToolService/RegisterTool":            {Capability: model.APIKeyCapabilityExecute, BindSession: true},
	"/api.ToolService/ListTools":               {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.ToolService/GetToolById":             {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.ToolService/GetToolByName":           {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.ToolService/DeleteTool":              {Capability: model.APIKeyCapabilityExecute, BindSession: true},
	"/api.ToolService/UpdateToolPing":          {Capability: model.APIKeyCapabilityExecute, BindSession: true},
	"/api.ToolService/ExecuteTool":             {Capability: model.APIKeyCapabilityExecute, BindSession: true},
	"/api.ToolService/StreamExecuteTool":       {Capability: model.APIKeyCapabilityExecute},
	"/api.ToolService/ResumeStream":            {Capability: model.APIKeyCapabilityExecute},
	"/api.ToolService/HealthCheck":             {Capability: model.APIKeyCapabilityRead},
	"/api.SessionsService/CreateSession":       {Capability: model.APIKeyCapabilityAdmin, BindUser: true},
	"/api.SessionsService/GetSession":          {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.SessionsService/ListSessions":        {Capability: model.APIKeyCapabilityAdmin, BindUser: true},
	"/api.SessionsService/UpdateSession":       {Capability: model.APIKeyCapabilityAdmin, BindSession: true},
	"/api.SessionsService/DeleteSession":       {Capability: model.APIKeyCapabilityAdmin, BindSession: true},
	"/api.SessionsService/ListUserSessions":    {Capability: model.APIKeyCapabilityAdmin, BindUser: true},
	"/api.SessionsService/BulkDeleteSessions":  {Capability: model.APIKeyCapabilityAdmin, BindUser: true},
	"/api.SessionsService/GetSessionStats":     {Capability: model.APIKeyCapabilityAdmin, BindUser: true},
	"/api.SessionsService/RefreshSessionToken": {Capability: model.APIKeyCapabilityAdmin, BindSession: true},
	"/api.SessionsService/InvalidateSession":   {Capability: model.APIKeyCapabilityAdmin, BindSession: true},
	"/api.SessionsService/CreateApiKey":        {Capability: model.APIKeyCapabilityAdmin, BindSession: true},
	"/api.SessionsService/ListApiKeys":         {Capability: model.APIKeyCapabilityAdmin, BindSession: true},
	"/api.SessionsService/RevokeApiKey":        {Capability: model.APIKeyCapabilityAdmin, BindSession: true},
	"/api.MachinesService/RegisterMachine":     {Capability: model.APIKeyCapabilityExecute, BindSession: true},
	"/api.MachinesService/ListMachines":        {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.MachinesService/GetMachine":          {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.MachinesService/UpdateMachinePing":   {Capability: model.APIKeyCapabilityExecute, BindSession: true},
	"/api.MachinesService/UnregisterMachine":   {Capability: model.APIKeyCapabilityExecute, BindSession: true},
	"/api.MachinesService/DrainMachine":        {Capability: model.APIKeyCapabilityExecute, BindSession: true},
	"/api.RequestsService/CreateRequest":       {Capability: model.APIKeyCapabilityExecute, BindSession: true},
	"/api.RequestsService/GetRequest":          {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.RequestsService/ListRequests":        {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.RequestsService/UpdateRequest":       {Capability: model.APIKeyCapabilityExecute, BindSession: true},
	"/api.RequestsService/ClaimRequest":        {Capability: model.APIKeyCapabilityExecute, BindSession: true},
	"/api.RequestsService/CancelRequest":       {Capability: model.APIKeyCapabilityExecute, BindSession: true},
	"/api.RequestsService/SubmitRequestResult": {Capability: model.APIKeyCapabilityExecute, BindSession: true},
	"/api.RequestsService/AppendRequestChunks": {Capability: model.APIKeyCapabilityExecute, BindSession: true},
	"/api.RequestsService/GetRequestChunks":    {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.TasksService/CreateTask":             {Capability: model.APIKeyCapabilityExecute, BindSession: true},
	"/api.TasksService/GetTask":                {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.TasksService/ListTasks":              {Capability: model.APIKeyCapabilityRead, BindSession: true},
	"/api.TasksService/CancelTask":             {Capability: model.APIKeyCapabilityExecute, BindSession: true},
}

func DebugPrincipal(principal *model.AuthPrincipal) string {
	if principal == nil {
		return "<nil>"
	}
	return fmt.Sprintf("mode=%s session=%s user=%s key=%s caps=%v", principal.Mode, principal.SessionID, principal.UserID, principal.KeyID, model.CapabilityStrings(principal.Capabilities))
}
