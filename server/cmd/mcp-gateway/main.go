package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/keepalive"

	"toolplane/pkg/mcp"
)

// corsMiddleware applies explicit development or production CORS behavior,
// mirroring cmd/proxy so browser-based MCP clients (e.g. MCP Inspector) work.
func corsMiddleware(cfg gatewayConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin != "" {
			allowedOrigin, allowed := cfg.matchOrigin(origin)
			if !allowed {
				http.Error(w, "origin not allowed", http.StatusForbidden)
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Add("Vary", "Origin")
		} else if cfg.allowAnyOrigin {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, MCP-Protocol-Version")
		w.Header().Set("Access-Control-Expose-Headers", "Retry-After")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type healthResponse struct {
	Status    string    `json:"status"`
	Backend   string    `json:"backend"`
	Timestamp time.Time `json:"timestamp"`
}

func writeHealth(w http.ResponseWriter, status string, backend string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(healthResponse{
		Status:    status,
		Backend:   backend,
		Timestamp: time.Now().UTC(),
	}); err != nil {
		log.Printf("failed to encode health response: %v", err)
	}
}

func main() {
	cfg, err := loadGatewayConfig()
	if err != nil {
		log.Fatalf("invalid mcp-gateway configuration: %v", err)
	}

	httpListen := flag.String("listen", ":8081", "HTTP listen address for the MCP facade")
	grpcEndpoint := flag.String("backend", "localhost:9001", "Toolplane gRPC server endpoint")
	syncTimeout := flag.Duration("sync-timeout", 60*time.Second, "maximum time a tools/call blocks for clients without Tasks support")
	pollInterval := flag.Duration("poll-interval", 250*time.Millisecond, "backend task poll cadence")
	defaultUserID := flag.String("default-user-id", "mcp-gateway", "user ID for auto-provisioned sessions")
	maxMsgSize := flag.Int("max-msg-size", 4*1024*1024, "Maximum gRPC message size in bytes")
	// Rate limiting controls, mirroring cmd/proxy (0 disables).
	apiRate := flag.Float64("api-rate", 0, "Maximum requests per second per API key (0 disables)")
	apiBurst := flag.Int("api-burst", 0, "Burst size per API key when rate limiting is enabled")
	ipRate := flag.Float64("ip-rate", 0, "Maximum requests per second per client IP (0 disables)")
	ipBurst := flag.Int("ip-burst", 0, "Burst size per client IP when rate limiting is enabled")
	// Client-facing TLS. When both files are set the gateway serves HTTPS
	// directly instead of relying on an upstream terminator.
	tlsCertFile := flag.String("tls-cert-file", "", "TLS certificate for the client-facing listener (enables HTTPS)")
	tlsKeyFile := flag.String("tls-key-file", "", "TLS private key for the client-facing listener")
	flag.Parse()

	serveTLS := *tlsCertFile != "" || *tlsKeyFile != ""
	if serveTLS && (*tlsCertFile == "" || *tlsKeyFile == "") {
		log.Fatalf("--tls-cert-file and --tls-key-file must be set together")
	}
	if cfg.environment == "production" && !serveTLS && !cfg.trustedProxy {
		log.Fatalf("production requires client-facing TLS (--tls-cert-file/--tls-key-file) or an explicit TOOLPLANE_TRUSTED_PROXY=1 declaration")
	}

	rateLimiter := newGatewayRateLimiter(rate.Limit(*apiRate), *apiBurst, rate.Limit(*ipRate), *ipBurst)

	transportCredentials, err := backendTransportCredentials(cfg)
	if err != nil {
		log.Fatalf("invalid backend TLS configuration: %v", err)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCredentials),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(*maxMsgSize),
			grpc.MaxCallSendMsgSize(*maxMsgSize),
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, callOpts ...grpc.CallOption) error {
			if _, ok := ctx.Deadline(); !ok {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
				defer cancel()
			}
			return invoker(ctx, method, req, reply, cc, callOpts...)
		}),
	}

	conn, err := grpc.NewClient(*grpcEndpoint, opts...)
	if err != nil {
		log.Fatalf("failed to create backend connection: %v", err)
	}
	defer conn.Close()

	facade := mcp.NewServer(conn,
		mcp.WithSyncTimeout(*syncTimeout),
		mcp.WithPollInterval(*pollInterval),
		mcp.WithDefaultUserID(*defaultUserID),
	)

	root := http.NewServeMux()
	root.Handle("/", corsMiddleware(cfg, gatewayRateLimitMiddleware(rateLimiter, cfg.trustedProxy, facade.Handler())))
	// Health probes transport-level connectivity only: backend RPCs require API
	// keys, which an unauthenticated health check does not carry. The probe
	// waits up to its deadline for the (lazily established) gRPC connection to
	// reach READY so an idle-but-healthy backend is not reported as degraded.
	root.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		conn.Connect()
		for {
			// Capture the state once per iteration so WaitForStateChange waits
			// for a change *from* that observed state. Re-reading GetState()
			// between the check and the wait could otherwise pass READY into
			// WaitForStateChange and block until the connection leaves READY,
			// needlessly delaying an already-ready backend.
			state := conn.GetState()
			if state == connectivity.Ready {
				break
			}
			if !conn.WaitForStateChange(ctx, state) {
				break // context expired before the connection became READY
			}
		}
		if conn.GetState() == connectivity.Ready {
			writeHealth(w, "ok", *grpcEndpoint, http.StatusOK)
			return
		}
		writeHealth(w, "degraded", *grpcEndpoint, http.StatusServiceUnavailable)
	})

	log.Printf("MCP gateway listening on %s → gRPC %s (env=%s cors=%s backend=%s sync-timeout=%s)",
		*httpListen, *grpcEndpoint, cfg.environment, cfg.corsSummary(), cfg.backendSecuritySummary(), *syncTimeout)
	// ReadHeaderTimeout bounds slow-loris header reads; WriteTimeout stays
	// unset because JSON-RPC responses ride long-lived keep-alive connections.
	httpServer := &http.Server{
		Addr:              *httpListen,
		Handler:           root,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
	}
	if serveTLS {
		log.Printf("client-facing TLS enabled (cert=%s)", *tlsCertFile)
		log.Fatal(httpServer.ListenAndServeTLS(*tlsCertFile, *tlsKeyFile))
	}
	log.Fatal(httpServer.ListenAndServe())
}
