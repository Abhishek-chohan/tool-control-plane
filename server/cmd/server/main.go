package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"toolplane/pkg/model"
	"toolplane/pkg/observability"
	"toolplane/pkg/service"
	"toolplane/pkg/storage"
	"toolplane/pkg/storage/memory"
	"toolplane/pkg/trace"
	proto "toolplane/proto"

	"toolplane/cmd/server/auth"
)

// storeClose closes the underlying store if the implementation exposes Close.
// Both the Postgres store (*storage.Store) and the in-memory store
// (*memory.Store) implement Close().
func storeClose(store storage.Storer) error {
	if c, ok := store.(interface{ Close() error }); ok {
		return c.Close()
	}
	return nil
}

func main() {
	os.Exit(run())
}

// run owns the process lifecycle and returns the exit code; every path
// returns normally so the deferred storage close runs — os.Exit lives only
// in main, above.
func run() int {
	port := flag.Int("port", 9001, "Port for gRPC server")
	enableTrace := flag.Bool("trace-sessions", false, "Log session lifecycle tracing events")
	metricsListen := flag.String("metrics-listen", "127.0.0.1:0", "HTTP listen address for Prometheus metrics; empty disables the endpoint")
	migrateOnly := flag.Bool("migrate-only", false, "Validate config, initialize storage, run migrations, then exit")
	tlsCertFile := flag.String("tls-cert-file", strings.TrimSpace(os.Getenv("TOOLPLANE_SERVER_TLS_CERT_FILE")), "Path to a PEM-encoded gRPC TLS certificate; empty disables TLS outside production")
	tlsKeyFile := flag.String("tls-key-file", strings.TrimSpace(os.Getenv("TOOLPLANE_SERVER_TLS_KEY_FILE")), "Path to a PEM-encoded gRPC TLS private key; empty disables TLS outside production")
	flag.Parse()

	// Structured logging for the whole process; the std log bridge below
	// routes the services' existing log.Printf output through the same
	// handler, so every line is one format.
	logger := observability.NewLogger(observability.LogFormatFromEnv(), os.Stdout)
	slog.SetDefault(logger)
	observability.RouteStdLog(logger)

	cfg, err := loadServerConfig()
	if err != nil {
		slog.Error("invalid server configuration", slog.Any("err", err))
		return 1
	}
	// migrate-only does not start the gRPC server, so transport settings are not required.
	if !*migrateOnly {
		if err := validateGRPCTLSSettings(cfg.environment, *tlsCertFile, *tlsKeyFile); err != nil {
			slog.Error("invalid gRPC TLS configuration", slog.Any("err", err))
			return 1
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metricsCollector := observability.NewRuntimeMetricsCollector()
	var tracers []trace.SessionTracer
	tracers = append(tracers, metricsCollector)
	var tracer trace.SessionTracer = metricsCollector

	pgStore, err := storage.OpenFromEnv(ctx, log.Default())
	var store storage.Storer
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrExplicitInMemoryMode):
			store = memory.New()
			slog.Info("storage mode: explicit in-memory")
		case errors.Is(err, storage.ErrConfigMissing):
			slog.Error("storage configuration error", slog.Any("err", err))
			return 1
		default:
			slog.Error("failed to initialize storage", slog.Any("err", err))
			return 1
		}
	} else {
		store = pgStore
	}
	defer func() {
		if cerr := storeClose(store); cerr != nil {
			slog.Error("error closing storage", slog.Any("err", cerr))
		}
	}()
	if *migrateOnly {
		if pgStore == nil {
			slog.Error("migrate-only requires Postgres-backed storage")
			return 1
		}
		slog.Info("database schema ready", "env", cfg.environment, "storage", "postgres")
		return 0
	}

	// The durable audit trail piggybacks on lifecycle events; failures
	// are logged by the recorder and never block the operation.
	var auditRecorder *observability.AuditRecorder
	if store != nil {
		auditRecorder = observability.NewAuditRecorder(ctx, store)
		tracers = append(tracers, auditRecorder)
	}
	if *enableTrace {
		tracers = append(tracers, trace.NewLoggingTracer(log.Default()))
	}
	if len(tracers) > 1 {
		tracer = trace.NewMultiTracer(tracers...)
	}

	// initialize your services
	sessionSvc := service.NewSessionsService(tracer, store)
	toolSvc := service.NewToolService(tracer, store)
	machineSvc := service.NewMachinesService(ctx, toolSvc, tracer, store)
	requestSvc := service.NewRequestsService(ctx, toolSvc, machineSvc, tracer, store)
	tasksSvc := service.NewTasksService(ctx, toolSvc, machineSvc, requestSvc, tracer, store)
	metricsCollector.Bind(requestSvc, machineSvc, tasksSvc)

	postgresAuthenticator := func(_ context.Context, token string) (*model.AuthPrincipal, error) {
		return sessionSvc.AuthenticateAPIKey(token)
	}

	authenticateAPIKey, authSummary, err := cfg.buildAuthenticator(postgresAuthenticator)
	if err != nil {
		slog.Error("failed to configure auth", slog.Any("err", err))
		return 1
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		slog.Error("failed to listen", slog.Any("err", err))
		return 1
	}

	serverOptions, transportSummary, err := grpcServerTransport(*tlsCertFile, *tlsKeyFile)
	if err != nil {
		slog.Error("failed to configure gRPC transport", slog.Any("err", err))
		return 1
	}
	// Explicit message bounds and keepalive enforcement: the defaults leave
	// message size implicit (4MiB) and idle connections unpoliced. 16MiB
	// accommodates a full chunk batch (32 x 512KiB) per RPC. MinTime must
	// stay below the proxy/gateway ping interval (10s): a client pinging at
	// or under MinTime gets GOAWAY and its connection killed.
	serverOptions = append(serverOptions,
		grpc.MaxRecvMsgSize(model.MaxChunkBatchBytes),
		grpc.MaxSendMsgSize(model.MaxChunkBatchBytes),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	// One interceptor chain, assembled explicitly. Metrics outermost so
	// auth rejections and transport failures are counted like any other
	// request; auth (or its anonymous dev-mode stand-in) sits inside it.
	var unaryInterceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor
	unaryInterceptors = append(unaryInterceptors, metricsCollector.UnaryServerInterceptor())
	streamInterceptors = append(streamInterceptors, metricsCollector.StreamServerInterceptor())
	if authenticateAPIKey != nil {
		authorizer := auth.NewAPIKeyAuthorizer(
			authenticateAPIKey,
			tracer,
			auth.WithMachineTokenAuth(machineSvc.AuthorizeMachineToken),
		)
		unaryInterceptors = append(unaryInterceptors, authorizer.UnaryInterceptor())
		streamInterceptors = append(streamInterceptors, authorizer.StreamInterceptor())
	} else {
		// Auth-disabled dev mode: still attach an anonymous fixed-mode
		// principal so handler-level fail-closed checks behave consistently.
		unaryInterceptors = append(unaryInterceptors, auth.AnonymousUnaryInterceptor())
		streamInterceptors = append(streamInterceptors, auth.AnonymousStreamInterceptor())
	}
	serverOptions = append(serverOptions,
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)
	server := grpc.NewServer(serverOptions...)

	// Standard grpc.health.v1: load balancers and orchestrators probe this
	// instead of inferring liveness from connection acceptance.
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(server, healthSrv)

	metricsDone := startMetricsServer(ctx, *metricsListen, metricsCollector)

	// The shutdown goroutine owns the drain and nothing else: it never
	// consumes the serve/metrics channels, so the main select below is
	// their single reader (a shared channel consumed from two selects can
	// deadlock when one receiver blocks forever on a result the other
	// already took).
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	stopDone := make(chan struct{})
	go func() {
		defer close(stopDone)
		select {
		case <-sigCh:
		case <-ctx.Done():
		}
		slog.Info("shutdown signal received, graceful stop")
		cancel()
		healthSrv.Shutdown() // report NOT_SERVING before draining
		stopped := make(chan struct{})
		go func() {
			server.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
		case <-time.After(20 * time.Second):
			slog.Warn("graceful stop deadline exceeded; forcing stop")
			server.Stop()
		}
	}()

	adapter := service.NewGRPCServer(
		toolSvc, sessionSvc, machineSvc, requestSvc, tasksSvc,
	)

	// register services
	proto.RegisterToolServiceServer(server, adapter)
	proto.RegisterSessionsServiceServer(server, adapter)
	proto.RegisterMachinesServiceServer(server, adapter)
	proto.RegisterRequestsServiceServer(server, adapter)
	proto.RegisterTasksServiceServer(server, adapter)

	slog.Info("gRPC server listening",
		"addr", lis.Addr().String(),
		"env", cfg.environment,
		"auth", authSummary,
		"transport", transportSummary)
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(lis) }()

	// Single exit path. A signal drains gracefully; an unsolicited serve
	// failure tears everything down; a nil serve result means GracefulStop
	// completed. Every path joins the drain (stopDone) before returning so
	// the deferred storage close runs with both servers fully stopped.
	exitCode := 0
	for {
		select {
		case err := <-serveErr:
			if err != nil {
				slog.Error("gRPC serve error", slog.Any("err", err))
				exitCode = 1
				cancel()
				healthSrv.Shutdown()
				server.Stop()
			}
			<-stopDone
			return exitCode
		case err := <-metricsDone:
			if err != nil {
				slog.Error("metrics serve error", slog.Any("err", err))
				exitCode = 1
				cancel()
				healthSrv.Shutdown()
				server.Stop()
				<-stopDone
				return exitCode
			}
			// nil: the metrics server drained as part of shutdown — keep
			// waiting for the gRPC result or the drain completion.
			metricsDone = nil
		case <-stopDone:
			return exitCode
		}
	}
}

// startMetricsServer serves /metrics and reports its terminal outcome on
// the returned channel; an empty listen address disables it (nil channel —
// a receive blocks forever, which the caller's select tolerates).
func startMetricsServer(ctx context.Context, listenAddr string, collector *observability.RuntimeMetricsCollector) <-chan error {
	if collector == nil || strings.TrimSpace(listenAddr) == "" {
		return nil
	}

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		slog.Error("failed to listen for metrics", slog.Any("err", err))
		out := make(chan error, 1)
		out <- err
		return out
	}

	server := &http.Server{
		Handler: collector.Handler(),
		// Bound header reads so a slow-loris client cannot pin connections.
		ReadHeaderTimeout: 10 * time.Second,
	}
	out := make(chan error, 1)
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("metrics server shutdown error", slog.Any("err", err))
		}
	}()

	go func() {
		slog.Info("metrics server listening", "addr", listener.Addr().String())
		err := server.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			out <- err
			return
		}
		out <- nil
	}()
	return out
}
