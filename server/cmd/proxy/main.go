package main

import (
	"flag"
	"os"

	"toolplane/internal/gateway"
)

func main() {
	opts := gateway.DefaultOptions()
	flag.StringVar(&opts.Listen, "listen", opts.Listen, "HTTP listen address for JSON gateway")
	flag.StringVar(&opts.Backend, "backend", opts.Backend, "gRPC server endpoint")
	// Sized from the server's chunk ladder: one AppendRequestChunks batch
	// (16 MiB payload) plus envelope headroom.
	flag.IntVar(&opts.MaxMsgSize, "max-msg-size", opts.MaxMsgSize, "Maximum message size in bytes")
	flag.Int64Var(&opts.MaxConcurrentRequests, "max-concurrent", opts.MaxConcurrentRequests, "Maximum concurrent requests")
	flag.Float64Var(&opts.APIRate, "api-rate", opts.APIRate, "Maximum requests per second per API key (0 disables)")
	flag.IntVar(&opts.APIBurst, "api-burst", opts.APIBurst, "Burst size per API key when rate limiting is enabled")
	flag.Float64Var(&opts.IPRate, "ip-rate", opts.IPRate, "Maximum requests per second per client IP (0 disables)")
	flag.IntVar(&opts.IPBurst, "ip-burst", opts.IPBurst, "Burst size per client IP when rate limiting is enabled")
	flag.StringVar(&opts.TLSCertFile, "tls-cert-file", opts.TLSCertFile, "TLS certificate for the client-facing listener (enables HTTPS)")
	flag.StringVar(&opts.TLSKeyFile, "tls-key-file", opts.TLSKeyFile, "TLS private key for the client-facing listener")
	flag.Parse()

	os.Exit(gateway.Run(opts))
}
