package main

import (
	"flag"
	"os"
	"strings"

	"toolplane/internal/server"
)

func main() {
	opts := server.DefaultOptions()

	// Flag defaults resolve environment fallbacks so `flag > env` holds on
	// this surface too: an unset flag falls back to its env var when one
	// exists, otherwise the development default above.
	flag.IntVar(&opts.Port, "port", opts.Port, "Port for gRPC server")
	flag.BoolVar(&opts.EnableTrace, "trace-sessions", opts.EnableTrace, "Log session lifecycle tracing events")
	flag.StringVar(&opts.MetricsListen, "metrics-listen", opts.MetricsListen, "HTTP listen address for Prometheus metrics; empty disables the endpoint")
	flag.BoolVar(&opts.MigrateOnly, "migrate-only", opts.MigrateOnly, "Validate config, initialize storage, run migrations, then exit")
	flag.StringVar(&opts.TLSCertFile, "tls-cert-file", strings.TrimSpace(os.Getenv("TOOLPLANE_SERVER_TLS_CERT_FILE")), "Path to a PEM-encoded gRPC TLS certificate; empty disables TLS outside production")
	flag.StringVar(&opts.TLSKeyFile, "tls-key-file", strings.TrimSpace(os.Getenv("TOOLPLANE_SERVER_TLS_KEY_FILE")), "Path to a PEM-encoded gRPC TLS private key; empty disables TLS outside production")
	flag.Parse()

	os.Exit(server.Run(opts))
}
