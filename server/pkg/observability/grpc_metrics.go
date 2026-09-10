package observability

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// traceparentMetadataKey is the W3C trace correlation header carried in gRPC
// metadata. The gateway already validates and forwards it; the proxy adds it
// to its forwarded-header allowlist. The server consumes it for request-log
// correlation only — no span is created.
const traceparentMetadataKey = "traceparent"

// UnaryServerInterceptor records per-RPC metrics (count by method and status
// code, latency histogram by method) and emits a debug-level request log
// carrying the caller's traceparent when one was provided.
func (c *RuntimeMetricsCollector) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		c.observeGRPC(info.FullMethod, status.Code(err).String(), time.Since(start).Seconds())
		logRPC(ctx, info.FullMethod, err, time.Since(start))
		return resp, err
	}
}

// StreamServerInterceptor is the streaming counterpart of the unary
// interceptor; the observation covers the whole stream lifetime.
func (c *RuntimeMetricsCollector) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		err := handler(srv, ss)
		c.observeGRPC(info.FullMethod, status.Code(err).String(), time.Since(start).Seconds())
		logRPC(ss.Context(), info.FullMethod, err, time.Since(start))
		return err
	}
}

func logRPC(ctx context.Context, method string, err error, elapsed time.Duration) {
	logger := slog.Default()
	if logger == nil {
		return
	}
	failed := err != nil && status.Code(err) != codes.OK
	// Failures surface at Info even when the handler's floor is Info;
	// successes log only when Debug is enabled.
	if !failed && !logger.Enabled(ctx, slog.LevelDebug) {
		return
	}
	attrs := []slog.Attr{
		slog.String("rpc", method),
		slog.String("code", status.Code(err).String()),
		slog.Duration("duration", elapsed),
	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if values := md.Get(traceparentMetadataKey); len(values) > 0 && values[0] != "" {
			attrs = append(attrs, slog.String("traceparent", values[0]))
		}
	}
	level := slog.LevelDebug
	if failed {
		level = slog.LevelInfo
	}
	logger.LogAttrs(ctx, level, "rpc", attrs...)
}
