package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"toolplane/internal/server"
)

// printResolvedConfig echoes the effective configuration with per-key
// provenance so a misconfigured boot is diagnosable from one screen:
// every line says which layer supplied the winning value.
func printResolvedConfig(cmd *cobra.Command, opts server.Options, configPath string, fileSettings *server.FileSettings) {
	w := cmd.OutOrStdout()
	source := func(envVar, flagName string) string {
		if flagName != "" && cmd.Flags().Changed(flagName) {
			return "(flag)"
		}
		if fileSettings != nil {
			if origin, ok := fileSettings.Provenance[envVar]; ok && origin != "env" {
				if strings.HasPrefix(origin, "file:") {
					return "(file secret)"
				}
				return "(file)"
			}
		}
		if os.Getenv(envVar) != "" {
			return "(env)"
		}
		return "(default)"
	}
	secret := func(value string) string {
		if value == "" {
			return ""
		}
		return "***set***"
	}

	fmt.Fprintln(w, "resolved configuration:")
	fmt.Fprintf(w, "  port:               %-10d %s\n", opts.Port, layerFlag(fileSettings != nil && fileSettings.Port != 0, cmd, "port"))
	fmt.Fprintf(w, "  metrics_listen:     %-10s %s\n", opts.MetricsListen, layerFlag(fileSettings != nil && fileSettings.MetricsListen != "", cmd, "metrics-listen"))
	fmt.Fprintf(w, "  env:                %-10s %s\n", orDefault(os.Getenv("TOOLPLANE_ENV_MODE"), "development"), source("TOOLPLANE_ENV_MODE", ""))
	fmt.Fprintf(w, "  auth.mode:          %-10s %s\n", orDefault(os.Getenv("TOOLPLANE_AUTH_MODE"), "disabled"), source("TOOLPLANE_AUTH_MODE", ""))
	fmt.Fprintf(w, "  auth.fixed_api_key: %-10s %s\n", secret(os.Getenv("TOOLPLANE_AUTH_FIXED_API_KEY")), source("TOOLPLANE_AUTH_FIXED_API_KEY", ""))
	fmt.Fprintf(w, "  storage.mode:       %-10s %s\n", orDefault(os.Getenv("TOOLPLANE_STORAGE_MODE"), "(resolved at boot)"), source("TOOLPLANE_STORAGE_MODE", ""))
	fmt.Fprintf(w, "  database_url:       %-10s %s\n", secret(os.Getenv("TOOLPLANE_DATABASE_URL")), source("TOOLPLANE_DATABASE_URL", ""))
	fmt.Fprintf(w, "  tls.cert_file:      %-10s %s\n", orDefault(opts.TLSCertFile, "(disabled)"), source("TOOLPLANE_SERVER_TLS_CERT_FILE", "tls-cert-file"))
	fmt.Fprintf(w, "  tls.key_file:       %-10s %s\n", secret(opts.TLSKeyFile), source("TOOLPLANE_SERVER_TLS_KEY_FILE", "tls-key-file"))
	if configPath != "" {
		fmt.Fprintf(w, "  config_file:        %s\n", configPath)
	}
	fmt.Fprintln(w, "\nproduction gates run at boot; see server/docs/local-development.md for the environment contract.")
}

func orDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func layerFlag(fileSupplies bool, cmd *cobra.Command, flagName string) string {
	if cmd.Flags().Changed(flagName) {
		return "(flag)"
	}
	if fileSupplies {
		return "(file)"
	}
	return "(default)"
}
