package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCTLSConfig controls how the Go client dials TLS-enabled gRPC endpoints.
type GRPCTLSConfig struct {
	Enabled    bool
	CACertPath string
	ServerName string
}

// ClientOption mutates client construction settings without breaking existing call sites.
type ClientOption func(*ToolplaneClient)

// WithGRPCTLS enables TLS for the direct gRPC client and optionally supplies a custom CA bundle and server name.
func WithGRPCTLS(caCertPath, serverName string) ClientOption {
	return WithGRPCTLSConfig(GRPCTLSConfig{
		Enabled:    true,
		CACertPath: caCertPath,
		ServerName: serverName,
	})
}

// WithGRPCTLSConfig sets the full TLS configuration for the direct gRPC client.
func WithGRPCTLSConfig(cfg GRPCTLSConfig) ClientOption {
	normalized := GRPCTLSConfig{
		Enabled:    cfg.Enabled || strings.TrimSpace(cfg.CACertPath) != "" || strings.TrimSpace(cfg.ServerName) != "",
		CACertPath: strings.TrimSpace(cfg.CACertPath),
		ServerName: strings.TrimSpace(cfg.ServerName),
	}

	return func(client *ToolplaneClient) {
		client.tlsConfig = normalized
	}
}

func (c *ToolplaneClient) transportCredentials() (credentials.TransportCredentials, error) {
	if !c.tlsConfig.Enabled {
		return insecure.NewCredentials(), nil
	}

	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if c.tlsConfig.ServerName != "" {
		tlsConfig.ServerName = c.tlsConfig.ServerName
	}

	if c.tlsConfig.CACertPath != "" {
		pemBytes, err := os.ReadFile(c.tlsConfig.CACertPath)
		if err != nil {
			return nil, fmt.Errorf("read gRPC TLS CA certificate: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pemBytes) {
			return nil, fmt.Errorf("parse gRPC TLS CA certificate: no certificates found")
		}
		tlsConfig.RootCAs = pool
	}

	return credentials.NewTLS(tlsConfig), nil
}
