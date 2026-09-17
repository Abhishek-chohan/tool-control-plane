package main

import (
	"flag"
	"os"

	"toolplane/internal/mcpgateway"
)

func main() {
	opts := mcpgateway.DefaultOptions()
	flag.StringVar(&opts.Listen, "listen", opts.Listen, "HTTP listen address for the MCP facade")
	flag.StringVar(&opts.Backend, "backend", opts.Backend, "Toolplane gRPC server endpoint")
	flag.DurationVar(&opts.SyncTimeout, "sync-timeout", opts.SyncTimeout, "maximum time a tools/call blocks for clients without Tasks support")
	flag.DurationVar(&opts.PollInterval, "poll-interval", opts.PollInterval, "backend task poll cadence")
	flag.StringVar(&opts.DefaultUserID, "default-user-id", opts.DefaultUserID, "user ID for auto-provisioned sessions")
	// Sized from the server's chunk ladder: one AppendRequestChunks batch
	// (16 MiB payload) plus envelope headroom.
	flag.IntVar(&opts.MaxMsgSize, "max-msg-size", opts.MaxMsgSize, "Maximum gRPC message size in bytes")
	flag.Float64Var(&opts.APIRate, "api-rate", opts.APIRate, "Maximum requests per second per API key (0 disables)")
	flag.IntVar(&opts.APIBurst, "api-burst", opts.APIBurst, "Burst size per API key when rate limiting is enabled")
	flag.Float64Var(&opts.IPRate, "ip-rate", opts.IPRate, "Maximum requests per second per client IP (0 disables)")
	flag.IntVar(&opts.IPBurst, "ip-burst", opts.IPBurst, "Burst size per client IP when rate limiting is enabled")
	flag.StringVar(&opts.TLSCertFile, "tls-cert-file", opts.TLSCertFile, "TLS certificate for the client-facing listener (enables HTTPS)")
	flag.StringVar(&opts.TLSKeyFile, "tls-key-file", opts.TLSKeyFile, "TLS private key for the client-facing listener")
	flag.Parse()

	os.Exit(mcpgateway.Run(opts))
}
