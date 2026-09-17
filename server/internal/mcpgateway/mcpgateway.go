package mcpgateway

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/keepalive"

	"toolplane/pkg/mcp"
	"toolplane/pkg/model"
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

// gatewayUnaryDeadline is the default unary bound for the facade's
// backend calls; the application loops (sync tools/call, task polling)
// issue short RPCs well under it. Mirrors cmd/proxy.
const gatewayUnaryDeadline = 30 * time.Second

// waitMethodBackstop bounds the wait-capable entrypoints the facade may
// call; it sits above the server's wait ceiling (3600s, see
// ExecuteToolRequest.wait_timeout_seconds) so an honest max-length wait
// completes while a server bug cannot hang the call. Mirrors cmd/proxy.
const waitMethodBackstop = time.Hour + 30*time.Second

var waitMethods = map[string]bool{
	"/api.v1.ToolService/InvokeTool":  true,
	"/api.v1.ToolService/ExecuteTool": true,
}

func unaryDeadline(method string) time.Duration {
	if waitMethods[method] {
		return waitMethodBackstop
	}
	return gatewayUnaryDeadline
}

// unaryDeadlineInterceptor applies the per-method deadline policy to the
// backend dial, defaulting only: an existing deadline (the facade's own
// sync-timeout budget) always wins. Mirrors cmd/proxy.
func unaryDeadlineInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, unaryDeadline(method))
		defer cancel()
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

// Options carries the MCP gateway's operator-facing settings; entry
// points (cmd/toolplane mcp serve) parse their own flag surfaces into it.
type Options struct {
	Listen        string
	Backend       string
	SyncTimeout   time.Duration
	PollInterval  time.Duration
	DefaultUserID string
	MaxMsgSize    int
	APIRate       float64
	APIBurst      int
	IPRate        float64
	IPBurst       int
	TLSCertFile   string
	TLSKeyFile    string
}

// DefaultOptions returns the standalone binary's defaults.
func DefaultOptions() Options {
	return Options{
		Listen:        ":8081",
		Backend:       "localhost:9001",
		SyncTimeout:   60 * time.Second,
		PollInterval:  250 * time.Millisecond,
		DefaultUserID: "mcp-gateway",
		MaxMsgSize:    model.MaxChunkBatchBytes + 1<<20,
	}
}

// Run owns process-level lifecycle: signals become context cancellation.
func Run(opts Options) int {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	return RunContext(ctx, opts)
}

// RunContext drives the MCP gateway until the context is cancelled or a
// serve error forces teardown; it returns the process exit code.
func RunContext(ctx context.Context, opts Options) int {
	cfg, err := loadGatewayConfig()
	if err != nil {
		log.Printf("invalid mcp-gateway configuration: %v", err)
		return 1
	}

	serveTLS := opts.TLSCertFile != "" || opts.TLSKeyFile != ""
	if serveTLS && (opts.TLSCertFile == "" || opts.TLSKeyFile == "") {
		log.Printf("--tls-cert-file and --tls-key-file must be set together")
		return 1
	}
	if cfg.environment == "production" && !serveTLS && !cfg.trustedProxy {
		log.Printf("production requires client-facing TLS (--tls-cert-file/--tls-key-file) or an explicit TOOLPLANE_TRUSTED_PROXY=1 declaration")
		return 1
	}

	// Derive from the caller's context: an internal failure triggers the
	// drain without owning signal handling.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rateLimiter := newGatewayRateLimiter(rate.Limit(opts.APIRate), opts.APIBurst, rate.Limit(opts.IPRate), opts.IPBurst)

	transportCredentials, err := backendTransportCredentials(cfg)
	if err != nil {
		log.Printf("invalid backend TLS configuration: %v", err)
		return 1
	}

	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCredentials),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(opts.MaxMsgSize),
			grpc.MaxCallSendMsgSize(opts.MaxMsgSize),
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithUnaryInterceptor(unaryDeadlineInterceptor),
	}

	conn, err := grpc.NewClient(opts.Backend, dialOptions...)
	if err != nil {
		log.Printf("failed to create backend connection: %v", err)
		return 1
	}
	defer conn.Close()

	// Auto-provisioning is a development convenience: production Postgres
	// auth refuses the create-session call, so gate it to non-production.
	// Production clients bind sessions explicitly via _meta.
	facade := mcp.NewServer(conn,
		mcp.WithSyncTimeout(opts.SyncTimeout),
		mcp.WithPollInterval(opts.PollInterval),
		mcp.WithDefaultUserID(opts.DefaultUserID),
		mcp.WithSessionAutoProvision(cfg.environment != "production"),
	)

	root := http.NewServeMux()
	root.Handle("/", corsMiddleware(cfg, gatewayRateLimitMiddleware(rateLimiter, cfg.trustedProxy, facade.Handler())))
	// Health probes transport-level connectivity only: backend RPCs require API
	// keys, which an unauthenticated health check does not carry. The probe
	// waits up to its deadline for the (lazily established) gRPC connection to
	// reach READY so an idle-but-healthy backend is not reported as degraded.
	root.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		healthCtx, healthCancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer healthCancel()
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
			if !conn.WaitForStateChange(healthCtx, state) {
				break // context expired before the connection became READY
			}
		}
		if conn.GetState() == connectivity.Ready {
			writeHealth(w, "ok", opts.Backend, http.StatusOK)
			return
		}
		writeHealth(w, "degraded", opts.Backend, http.StatusServiceUnavailable)
	})

	log.Printf("MCP gateway listening on %s → gRPC %s (env=%s cors=%s backend=%s sync-timeout=%s)",
		opts.Listen, opts.Backend, cfg.environment, cfg.corsSummary(), cfg.backendSecuritySummary(), opts.SyncTimeout)
	// ReadHeaderTimeout bounds slow-loris header reads; WriteTimeout stays
	// unset because JSON-RPC responses ride long-lived keep-alive connections.
	httpServer := &http.Server{
		Addr:              opts.Listen,
		Handler:           root,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		// IdleTimeout polices parked keep-alive connections; WriteTimeout
		// stays unset because JSON-RPC responses ride long-lived streams.
		IdleTimeout: 120 * time.Second,
	}
	serveErr := make(chan error, 1)
	go func() {
		if serveTLS {
			log.Printf("client-facing TLS enabled (cert=%s)", opts.TLSCertFile)
			serveErr <- httpServer.ListenAndServeTLS(opts.TLSCertFile, opts.TLSKeyFile)
			return
		}
		serveErr <- httpServer.ListenAndServe()
	}()

	// Shutdown drains in-flight requests up to the deadline, mirroring the
	// signal path the standalone binary has always taken.
	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("gateway serve error: %v", err)
			return 1
		}
		return 0
	case <-ctx.Done():
		log.Println("shutdown requested, draining gateway")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("gateway shutdown error: %v", err)
		}
		return 0
	}
}
