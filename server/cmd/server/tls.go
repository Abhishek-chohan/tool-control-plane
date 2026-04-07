package main

import (
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func validateGRPCTLSSettings(environment, certFile, keyFile string) error {
	environment = strings.TrimSpace(strings.ToLower(environment))
	certFile = strings.TrimSpace(certFile)
	keyFile = strings.TrimSpace(keyFile)

	switch {
	case certFile == "" && keyFile == "":
		if environment == "production" {
			return fmt.Errorf("gRPC TLS certificate and key files are required when TOOLPLANE_ENV_MODE=production")
		}
		return nil
	case certFile == "" || keyFile == "":
		return fmt.Errorf("gRPC TLS requires both certificate and key files")
	default:
		return nil
	}
}

func grpcServerTransport(certFile, keyFile string) ([]grpc.ServerOption, string, error) {
	certFile = strings.TrimSpace(certFile)
	keyFile = strings.TrimSpace(keyFile)
	if certFile == "" && keyFile == "" {
		return nil, "plaintext", nil
	}

	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		return nil, "", fmt.Errorf("load gRPC TLS credentials: %w", err)
	}
	return []grpc.ServerOption{grpc.Creds(creds)}, "tls", nil
}
