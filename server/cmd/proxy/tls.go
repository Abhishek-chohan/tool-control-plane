package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func backendTransportCredentials(cfg proxyConfig) (credentials.TransportCredentials, error) {
	if cfg.allowInsecureBackend {
		return insecure.NewCredentials(), nil
	}

	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if cfg.backendTLSServerName != "" {
		tlsConfig.ServerName = cfg.backendTLSServerName
	}

	if caFile := strings.TrimSpace(cfg.backendTLSCAFile); caFile != "" {
		pemBytes, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("read proxy backend TLS CA file: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pemBytes) {
			return nil, fmt.Errorf("parse proxy backend TLS CA file: no certificates found")
		}
		tlsConfig.RootCAs = pool
	}

	return credentials.NewTLS(tlsConfig), nil
}
