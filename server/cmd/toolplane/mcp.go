package main

import (
	"github.com/spf13/cobra"

	"toolplane/internal/mcpgateway"
)

// newMCPCommand groups the MCP edge verbs; `mcp serve` runs the same
// lifecycle as the standalone toolplane-mcp-gateway binary.
func newMCPCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Model Context Protocol edge (JSON-RPC facade)",
	}
	cmd.AddCommand(newMCPServeCommand())
	return cmd
}

func newMCPServeCommand() *cobra.Command {
	opts := mcpgateway.DefaultOptions()

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the MCP gateway",
		Long: `Run the MCP gateway: a stateless JSON-RPC facade exposing Toolplane
tools over MCP at POST /mcp (health at GET /health). Sessions bind via
_meta or auto-provision per API key outside production.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return exitError{code: mcpgateway.Run(opts)}
		},
	}

	cmd.Flags().StringVar(&opts.Listen, "listen", opts.Listen, "HTTP listen address for the MCP facade")
	cmd.Flags().StringVar(&opts.Backend, "backend", opts.Backend, "Toolplane gRPC server endpoint")
	cmd.Flags().DurationVar(&opts.SyncTimeout, "sync-timeout", opts.SyncTimeout, "maximum time a tools/call blocks for clients without Tasks support")
	cmd.Flags().DurationVar(&opts.PollInterval, "poll-interval", opts.PollInterval, "backend task poll cadence")
	cmd.Flags().StringVar(&opts.DefaultUserID, "default-user-id", opts.DefaultUserID, "user ID for auto-provisioned sessions")
	cmd.Flags().IntVar(&opts.MaxMsgSize, "max-msg-size", opts.MaxMsgSize, "maximum gRPC message size in bytes")
	cmd.Flags().Float64Var(&opts.APIRate, "api-rate", opts.APIRate, "maximum requests per second per API key (0 disables)")
	cmd.Flags().IntVar(&opts.APIBurst, "api-burst", opts.APIBurst, "burst size per API key when rate limiting is enabled")
	cmd.Flags().Float64Var(&opts.IPRate, "ip-rate", opts.IPRate, "maximum requests per second per client IP (0 disables)")
	cmd.Flags().IntVar(&opts.IPBurst, "ip-burst", opts.IPBurst, "burst size per client IP when rate limiting is enabled")
	cmd.Flags().StringVar(&opts.TLSCertFile, "tls-cert-file", opts.TLSCertFile, "TLS certificate for the client-facing listener (enables HTTPS)")
	cmd.Flags().StringVar(&opts.TLSKeyFile, "tls-key-file", opts.TLSKeyFile, "TLS private key for the client-facing listener")

	return cmd
}
