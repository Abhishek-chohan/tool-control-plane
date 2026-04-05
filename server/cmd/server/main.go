package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"toolplane/pkg/service"
	"toolplane/pkg/storage"
	"toolplane/pkg/trace"
	proto "toolplane/proto"

	"toolplane/cmd/server/auth"
)

func main() {
	port := flag.Int("port", 9001, "Port for gRPC server")
	enableTrace := flag.Bool("trace-sessions", false, "Log session lifecycle tracing events")
	flag.Parse()

	cfg, err := loadServerConfig()
	if err != nil {
		log.Fatalf("invalid server configuration: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var tracer trace.SessionTracer
	if *enableTrace {
		tracer = trace.NewLoggingTracer(log.Default())
	} else {
		tracer = trace.NopTracer()
	}

	store, err := storage.OpenFromEnv(ctx, log.Default())
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrExplicitInMemoryMode):
			log.Printf("storage mode: explicit in-memory")
		case errors.Is(err, storage.ErrConfigMissing):
			log.Fatalf("storage configuration error: %v", err)
		default:
			log.Fatalf("failed to initialize storage: %v", err)
		}
	}
	if store != nil {
		defer func() {
			if cerr := store.Close(); cerr != nil {
				log.Printf("error closing storage: %v", cerr)
			}
		}()
	}

	// initialize your services
	sessionSvc := service.NewSessionsService(tracer, store)
	toolSvc := service.NewToolService(tracer, store)
	machineSvc := service.NewMachinesService(toolSvc, tracer, store)
	requestSvc := service.NewRequestsService(toolSvc, machineSvc, tracer, store)
	tasksSvc := service.NewTasksService(ctx, toolSvc, machineSvc, requestSvc, tracer, store)

	postgresValidator := func(_ context.Context, token string) bool {
		_, err := sessionSvc.ValidateApiKey(token)
		return err == nil
	}

	validateAPIKey, authSummary, err := cfg.buildValidator(postgresValidator)
	if err != nil {
		log.Fatalf("failed to configure auth: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	serverOptions := make([]grpc.ServerOption, 0, 2)
	if validateAPIKey != nil {
		serverOptions = append(serverOptions,
			grpc.UnaryInterceptor(auth.UnaryAPIKeyInterceptor(validateAPIKey)),
			grpc.StreamInterceptor(auth.StreamAPIKeyInterceptor(validateAPIKey)),
		)
	}
	server := grpc.NewServer(serverOptions...)

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

	log.Printf("gRPC server listening at %v (env=%s auth=%s)", lis.Addr(), cfg.environment, authSummary)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("gRPC serve error: %v", err)
	}
}
