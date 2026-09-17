package auth

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func tokenFromMetadata(md metadata.MD) (string, error) {
	authHeaders := md.Get("authorization")
	if len(authHeaders) > 0 && strings.TrimSpace(authHeaders[0]) != "" {
		return authHeaders[0], nil
	}

	apiKeyHeaders := md.Get("api_key")
	if len(apiKeyHeaders) > 0 && strings.TrimSpace(apiKeyHeaders[0]) != "" {
		return apiKeyHeaders[0], nil
	}

	return "", errors.New("authorization token not supplied")
}

func normalizeToken(token string) string {
	trimmed := strings.TrimSpace(token)
	if strings.HasPrefix(strings.ToLower(trimmed), "bearer ") {
		return strings.TrimSpace(trimmed[7:])
	}
	return trimmed
}

func redactToken(token string) string {
	if token == "" {
		return "<empty>"
	}
	if len(token) <= 8 {
		return "<redacted>"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

// UnaryAPIKeyInterceptor adds API key validation to all unary RPCs.
func UnaryAPIKeyInterceptor(validate func(context.Context, string) bool) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("missing metadata")
		}

		token, err := tokenFromMetadata(md)
		if err != nil {
			return nil, err
		}

		if !validate(ctx, token) {
			return nil, errors.New("invalid API key")
		}

		return handler(ctx, req)
	}
}

// StreamAPIKeyInterceptor adds API key validation to all streaming RPCs.
func StreamAPIKeyInterceptor(validate func(context.Context, string) bool) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return errors.New("missing metadata")
		}

		token, err := tokenFromMetadata(md)
		if err != nil {
			return err
		}

		if !validate(ss.Context(), token) {
			return errors.New("invalid API key")
		}

		return handler(srv, ss)
	}
}
