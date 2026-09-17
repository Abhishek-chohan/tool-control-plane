package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"toolplane/internal/cli"
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
	var configPath string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the Toolplane gRPC control plane",
		Long: `Run the Toolplane gRPC control plane.

Development defaults are explicit and non-secret: in-memory storage,
fixed API key, no TLS. The startup banner prints exactly what is
insecure. Production mode refuses to boot until storage, auth, and
transport are real (TOOLPLANE_ENV_MODE=production requires Postgres
storage, Postgres auth, and gRPC TLS certificates).

A YAML config file (--config) supplies the base layer: flags override
environment variables, which override the file, which overrides
defaults. Unknown keys are a boot error. Values expand ${VAR} references,
and storage.database_url_file reads a secret from a mounted file.
--dry-run prints the resolved configuration with per-key provenance and
exits without serving.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var fileSettings *server.FileSettings
			if configPath != "" {
				settings, err := server.ApplyConfigFile(configPath)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return exitError{code: cli.ExitInvalidArgument}
				}
				fileSettings = settings
				// File values for env-backed flags were applied as
				// environment defaults after flag construction: re-read
				// the unchanged TLS flags so file-provided paths win over
				// their empty defaults, exactly as an env var would have.
				if !cmd.Flags().Changed("tls-cert-file") && opts.TLSCertFile == "" {
					opts.TLSCertFile = strings.TrimSpace(os.Getenv("TOOLPLANE_SERVER_TLS_CERT_FILE"))
				}
				if !cmd.Flags().Changed("tls-key-file") && opts.TLSKeyFile == "" {
					opts.TLSKeyFile = strings.TrimSpace(os.Getenv("TOOLPLANE_SERVER_TLS_KEY_FILE"))
				}
				if fileSettings.Port != 0 && !cmd.Flags().Changed("port") {
					opts.Port = fileSettings.Port
				}
				if fileSettings.MetricsListen != "" && !cmd.Flags().Changed("metrics-listen") {
					opts.MetricsListen = fileSettings.MetricsListen
				}
			}

			if dryRun {
				printResolvedConfig(cmd, opts, configPath, fileSettings)
				return exitError{code: cli.ExitOK}
			}
			return exitError{code: server.Run(opts)}
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to a YAML config file supplying the base layer (flags and environment override it)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "resolve configuration and exit without serving")

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
