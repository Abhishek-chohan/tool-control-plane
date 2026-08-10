package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"google.golang.org/grpc"
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
	port := flag.Int("port", 9001, "Port for gRPC server")
	enableTrace := flag.Bool("trace-sessions", false, "Log session lifecycle tracing events")
	metricsListen := flag.String("metrics-listen", "127.0.0.1:0", "HTTP listen address for Prometheus metrics; empty disables the endpoint")
	migrateOnly := flag.Bool("migrate-only", false, "Validate config, initialize storage, run migrations, then exit")
	tlsCertFile := flag.String("tls-cert-file", strings.TrimSpace(os.Getenv("TOOLPLANE_SERVER_TLS_CERT_FILE")), "Path to a PEM-encoded gRPC TLS certificate; empty disables TLS outside production")
	tlsKeyFile := flag.String("tls-key-file", strings.TrimSpace(os.Getenv("TOOLPLANE_SERVER_TLS_KEY_FILE")), "Path to a PEM-encoded gRPC TLS private key; empty disables TLS outside production")
	flag.Parse()

	cfg, err := loadServerConfig()
	if err != nil {
		log.Fatalf("invalid server configuration: %v", err)
	}
	// migrate-only does not start the gRPC server, so transport settings are not required.
	if !*migrateOnly {
		if err := validateGRPCTLSSettings(cfg.environment, *tlsCertFile, *tlsKeyFile); err != nil {
			log.Fatalf("invalid gRPC TLS configuration: %v", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metricsCollector := observability.NewRuntimeMetricsCollector()
	tracer := trace.SessionTracer(metricsCollector)
	if *enableTrace {
		tracer = trace.NewMultiTracer(metricsCollector, trace.NewLoggingTracer(log.Default()))
	}

	pgStore, err := storage.OpenFromEnv(ctx, log.Default())
	var store storage.Storer
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrExplicitInMemoryMode):
			store = memory.New()
			log.Printf("storage mode: explicit in-memory")
		case errors.Is(err, storage.ErrConfigMissing):
			log.Fatalf("storage configuration error: %v", err)
		default:
			log.Fatalf("failed to initialize storage: %v", err)
		}
	} else {
		store = pgStore
	}
	defer func() {
		if cerr := storeClose(store); cerr != nil {
			log.Printf("error closing storage: %v", cerr)
		}
	}()
	if *migrateOnly {
		if pgStore == nil {
			log.Fatalf("migrate-only requires Postgres-backed storage")
		}
		log.Printf("database schema ready (env=%s storage=postgres)", cfg.environment)
		return
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
		log.Fatalf("failed to configure auth: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	serverOptions, transportSummary, err := grpcServerTransport(*tlsCertFile, *tlsKeyFile)
	if err != nil {
		log.Fatalf("failed to configure gRPC transport: %v", err)
	}
	if authenticateAPIKey != nil {
		authorizer := auth.NewAPIKeyAuthorizer(authenticateAPIKey, tracer)
		serverOptions = append(serverOptions,
			grpc.UnaryInterceptor(authorizer.UnaryInterceptor()),
			grpc.StreamInterceptor(authorizer.StreamInterceptor()),
		)
	}
	server := grpc.NewServer(serverOptions...)
	startMetricsServer(ctx, *metricsListen, metricsCollector)

	// graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("shutdown signal received, graceful stop")
		cancel()
		server.GracefulStop()
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

	log.Printf("gRPC server listening at %v (env=%s auth=%s transport=%s)", lis.Addr(), cfg.environment, authSummary, transportSummary)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("gRPC serve error: %v", err)
	}
}

func startMetricsServer(ctx context.Context, listenAddr string, collector *observability.RuntimeMetricsCollector) {
	if collector == nil || strings.TrimSpace(listenAddr) == "" {
		return
	}

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen for metrics: %v", err)
	}

	server := &http.Server{Handler: collector.Handler()}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("metrics server shutdown error: %v", err)
		}
	}()

	go func() {
		log.Printf("metrics server listening at %v", listener.Addr())
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("metrics serve error: %v", err)
		}
	}()
}
