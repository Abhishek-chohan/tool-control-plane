package main

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"toolplane/internal/server"
)

// newServeCommand wraps the control-plane lifecycle behind the same flag
// surface as the standalone toolplane-server binary: every flag maps 1:1
// onto that binary's flags, with the same environment fallbacks, so
// either entry point can be swapped without re-learning configuration.
// Configuration beyond these flags (env mode, auth mode, storage mode)
// comes from the TOOLPLANE_* environment contract both entry points share.
func newServeCommand() *cobra.Command {
	opts := server.DefaultOptions()

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the Toolplane gRPC control plane",
		Long: `Run the Toolplane gRPC control plane.

Development defaults are explicit and non-secret: in-memory storage,
fixed API key, no TLS. The startup banner prints exactly what is
insecure. Production mode refuses to boot until storage, auth, and
transport are real (TOOLPLANE_ENV_MODE=production requires Postgres
storage, Postgres auth, and gRPC TLS certificates).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return exitError{code: server.Run(opts)}
		},
	}

	// Environment fallbacks mirror cmd/server: an unset TLS flag falls
	// back to its env var so `flag > env > default` holds on both entry
	// points.
	cmd.Flags().IntVar(&opts.Port, "port", opts.Port, "port for the gRPC server")
	cmd.Flags().BoolVar(&opts.EnableTrace, "trace-sessions", opts.EnableTrace, "log session lifecycle tracing events")
	cmd.Flags().StringVar(&opts.MetricsListen, "metrics-listen", opts.MetricsListen, "HTTP listen address for Prometheus metrics; empty disables the endpoint")
	cmd.Flags().BoolVar(&opts.MigrateOnly, "migrate-only", opts.MigrateOnly, "validate config, initialize storage, run migrations, then exit")
	cmd.Flags().StringVar(&opts.TLSCertFile, "tls-cert-file", strings.TrimSpace(os.Getenv("TOOLPLANE_SERVER_TLS_CERT_FILE")), "path to a PEM-encoded gRPC TLS certificate; empty disables TLS outside production")
	cmd.Flags().StringVar(&opts.TLSKeyFile, "tls-key-file", strings.TrimSpace(os.Getenv("TOOLPLANE_SERVER_TLS_KEY_FILE")), "path to a PEM-encoded gRPC TLS private key; empty disables TLS outside production")

	return cmd
}
