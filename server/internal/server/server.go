package server

import (
	"context"
	"errors"
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

	"toolplane/internal/auth"
)

// Options carries the server's operator-facing settings. Entry points
// (cmd/server, cmd/toolplane serve) parse their own flag surfaces into
// this struct; environment-based configuration (TOOLPLANE_ENV_MODE, auth
// mode, storage mode) is resolved inside Run against the same env contract
// as before, so the entry points cannot drift apart.
type Options struct {
	Port          int
	EnableTrace   bool
	MetricsListen string
	MigrateOnly   bool
	TLSCertFile   string
	TLSKeyFile    string
	// Listener, when set, is served instead of binding :Port — tests and
	// embedded callers hand a pre-bound listener to avoid pick-port races.
	Listener net.Listener
}

// DefaultOptions returns the development defaults the standalone binary
// has always shipped: explicit, non-secret, non-TLS.
func DefaultOptions() Options {
	return Options{
		Port:          9001,
		MetricsListen: "127.0.0.1:0",
	}
}

// storeClose closes the underlying store if the implementation exposes Close.
// Both the Postgres store (*storage.Store) and the in-memory store
// (*memory.Store) implement Close().
func storeClose(store storage.Storer) error {
	if c, ok := store.(interface{ Close() error }); ok {
		return c.Close()
	}
	return nil
}

// Run owns the process-level lifecycle: it wires SIGINT/SIGTERM into
// context cancellation and hands off to runContext. Every path returns
// normally so deferred cleanup runs; os.Exit belongs to the entry point.
func Run(opts Options) int {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	return RunContext(ctx, opts)
}

// runContext drives the server lifecycle until the context is cancelled
// (shutdown signal or test harness) or a serve error forces teardown. It
// returns the process exit code.
func RunContext(ctx context.Context, opts Options) int {

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
	if !opts.MigrateOnly {
		if err := validateGRPCTLSSettings(cfg.environment, opts.TLSCertFile, opts.TLSKeyFile, trustedTransportDeclared()); err != nil {
			slog.Error("invalid gRPC TLS configuration", slog.Any("err", err))
			return 1
		}
	}

	// Derive from the caller's context so an internal failure can trigger
	// the shutdown drain without owning signal handling (Run wires that).
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	metricsCollector := observability.NewRuntimeMetricsCollector()
	var tracers []trace.SessionTracer
	tracers = append(tracers, metricsCollector)
	var tracer trace.SessionTracer = metricsCollector

	pgStore, err := storage.OpenFromEnv(ctx, log.Default())
	var store storage.Storer
	storageState := ""
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrExplicitInMemoryMode):
			store = memory.New()
			storageState = "memory (explicit, non-durable)"
			service.StorageSummary = "memory"
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
		storageState = "postgres (durable)"
		service.StorageSummary = "postgres"
		// Serialization-retry telemetry: only the Postgres store has a
		// retry loop to observe (memory-mode stores serialize under a lock).
		pgStore.SetSerializationObserver(metricsCollector)
	}

	// Loud posture banner for non-production configurations: enumerate
	// exactly what is insecure so nobody discovers it from an incident.
	// Runs after storage resolution so the banner reports the resolved mode.
	if cfg.environment != "production" {
		tlsState := "disabled (plaintext)"
		if opts.TLSCertFile != "" {
			tlsState = "enabled"
		}
		authDetail := map[string]string{
			"disabled": "every caller receives an anonymous all-capabilities principal",
			"fixed":    "one shared API key for all local callers",
			"postgres": "per-session API keys",
		}[cfg.authMode]
		slog.Warn("INSECURE DEVELOPMENT CONFIGURATION — do not expose this server\n" +
			"  - auth: " + cfg.authMode + " (" + authDetail + ")\n" +
			"  - gRPC transport: " + tlsState + "\n" +
			"  - storage: " + storageState + "\n" +
			"  Accept the disabled-auth risk only by setting TOOLPLANE_ALLOW_INSECURE_DEV=1.")
	}
	defer func() {
		if cerr := storeClose(store); cerr != nil {
			slog.Error("error closing storage", slog.Any("err", cerr))
		}
	}()
	if opts.MigrateOnly {
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
	if opts.EnableTrace {
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

	// A pre-bound listener (tests, socket injection) wins over the port;
	// otherwise bind :Port. Binding here — not in the entry point — keeps
	// the listen and the health flip atomic with the serve loop.
	lis := opts.Listener
	if lis == nil {
		l, listenErr := net.Listen("tcp", fmt.Sprintf(":%d", opts.Port))
		if listenErr != nil {
			slog.Error("failed to listen", slog.Any("err", listenErr))
			return 1
		}
		lis = l
	}

	serverOptions, transportSummary, err := grpcServerTransport(opts.TLSCertFile, opts.TLSKeyFile)
	if err != nil {
		slog.Error("failed to configure gRPC transport", slog.Any("err", err))
		return 1
	}
	metricsCollector.SetTransportTLS(transportSummary == "tls")
	// The gate above only lets production boot plaintext when an upstream
	// terminator is declared; say so in the listening line instead of
	// leaving the plaintext label to explain itself.
	if transportSummary == "plaintext" && cfg.environment == "production" {
		transportSummary = "plaintext (upstream TLS terminator declared via TOOLPLANE_SERVER_TRUSTED_TRANSPORT)"
	}
	// Explicit message bounds and keepalive enforcement: the defaults leave
	// message size implicit (4MiB) and idle connections unpoliced. 16MiB
	// accommodates a full chunk batch (32 x 512KiB) per RPC. MinTime must
	// stay below the proxy/gateway ping interval (10s): a client pinging at
	// or under MinTime gets GOAWAY and its connection killed.
	serverOptions = append(serverOptions,
		grpc.MaxRecvMsgSize(model.MaxChunkBatchBytes),
		grpc.MaxSendMsgSize(model.MaxChunkBatchBytes),
		// Bound concurrent streams: the default is unlimited, and the
		// concurrency contract is enforced per machine at the request layer.
		grpc.MaxConcurrentStreams(2048),
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

	metricsDone := startMetricsServer(ctx, opts.MetricsListen, metricsCollector)

	// The shutdown goroutine owns the drain and nothing else: it never
	// consumes the serve/metrics channels, so the main select below is
	// their single reader (a shared channel consumed from two selects can
	// deadlock when one receiver blocks forever on a result the other
	// already took). Shutdown fires when the context is cancelled — by a
	// SIGINT/SIGTERM (wired in Run) or by a test harness.
	stopDone := make(chan struct{})
	go func() {
		defer close(stopDone)
		<-ctx.Done()
		slog.Info("shutdown requested, graceful stop")
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
