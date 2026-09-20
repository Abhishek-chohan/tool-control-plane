package server

import (
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func validateGRPCTLSSettings(environment, certFile, keyFile string, trustedTransport bool) error {
	environment = strings.TrimSpace(strings.ToLower(environment))
	certFile = strings.TrimSpace(certFile)
	keyFile = strings.TrimSpace(keyFile)

	switch {
	case certFile == "" && keyFile == "":
		if environment == "production" && !trustedTransport {
			return fmt.Errorf("gRPC TLS certificate and key files are required when TOOLPLANE_ENV_MODE=production; pass --tls-cert-file/--tls-key-file (or TOOLPLANE_SERVER_TLS_CERT_FILE/TOOLPLANE_SERVER_TLS_KEY_FILE), or declare an upstream TLS terminator with TOOLPLANE_SERVER_TRUSTED_TRANSPORT=1")
		}
		return nil
	case certFile == "" || keyFile == "":
		return fmt.Errorf("gRPC TLS requires both certificate and key files")
	default:
		return nil
	}
}

// trustedTransportDeclared reports the production-legal plaintext escape:
// an explicit declaration that TLS terminates upstream of this process
// (service mesh, sidecar, or terminating proxy) and the gRPC hop inside
// that boundary is intentionally plaintext.
func trustedTransportDeclared() bool {
	return boolEnv("TOOLPLANE_SERVER_TRUSTED_TRANSPORT", false)
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
