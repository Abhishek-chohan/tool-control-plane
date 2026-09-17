package main

import (
	"github.com/spf13/cobra"

	"toolplane/internal/gateway"
)

// newGatewayCommand groups the HTTP/JSON edge verbs; `gateway serve` runs
// the same lifecycle as the standalone toolplane-gateway binary with the
// same flags and environment contract.
func newGatewayCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "HTTP/JSON gateway (grpc-gateway transcode edge)",
	}
	cmd.AddCommand(newGatewayServeCommand())
	return cmd
}

func newGatewayServeCommand() *cobra.Command {
	opts := gateway.DefaultOptions()

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the HTTP/JSON gateway",
		Long: `Run the HTTP/JSON gateway: grpc-gateway transcoding of the api.v1
surface with CORS, rate limiting, and a circuit breaker in front of the
gRPC backend. Development defaults are explicit; production requires
client-facing TLS or an explicit TOOLPLANE_TRUSTED_PROXY=1 declaration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return exitError{code: gateway.Run(opts)}
		},
	}

	cmd.Flags().StringVar(&opts.Listen, "listen", opts.Listen, "HTTP listen address for the JSON gateway")
	cmd.Flags().StringVar(&opts.Backend, "backend", opts.Backend, "gRPC server endpoint")
	cmd.Flags().IntVar(&opts.MaxMsgSize, "max-msg-size", opts.MaxMsgSize, "maximum message size in bytes")
	cmd.Flags().Int64Var(&opts.MaxConcurrentRequests, "max-concurrent", opts.MaxConcurrentRequests, "maximum concurrent requests")
	cmd.Flags().Float64Var(&opts.APIRate, "api-rate", opts.APIRate, "maximum requests per second per API key (0 disables)")
	cmd.Flags().IntVar(&opts.APIBurst, "api-burst", opts.APIBurst, "burst size per API key when rate limiting is enabled")
	cmd.Flags().Float64Var(&opts.IPRate, "ip-rate", opts.IPRate, "maximum requests per second per client IP (0 disables)")
	cmd.Flags().IntVar(&opts.IPBurst, "ip-burst", opts.IPBurst, "burst size per client IP when rate limiting is enabled")
	cmd.Flags().StringVar(&opts.TLSCertFile, "tls-cert-file", opts.TLSCertFile, "TLS certificate for the client-facing listener (enables HTTPS)")
	cmd.Flags().StringVar(&opts.TLSKeyFile, "tls-key-file", opts.TLSKeyFile, "TLS private key for the client-facing listener")

	return cmd
}
