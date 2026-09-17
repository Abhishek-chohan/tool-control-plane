package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	"toolplane/pkg/model"
	gw "toolplane/proto"
	// Registers google.rpc.ErrorInfo so the gateway can marshal status
	// details attached by the server.
	_ "google.golang.org/genproto/googleapis/rpc/errdetails"
)

// corsMiddleware applies explicit development or production CORS behavior.
func corsMiddleware(cfg proxyConfig, next http.Handler) http.Handler {
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

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Grpc-Metadata-api_key, X-API-Key")
		w.Header().Set("Access-Control-Expose-Headers", "Retry-After")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// authHeaderMatcher forwards Authorization headers to gRPC metadata
func authHeaderMatcher(key string) (string, bool) {
	// Forward Authorization header to gRPC metadata
	if strings.EqualFold(key, "Authorization") {
		return "authorization", true
	}
	// Forward API key header for backward compatibility
	if strings.EqualFold(key, "Grpc-Metadata-api_key") {
		return "api_key", true
	}
	if strings.EqualFold(key, "X-API-Key") {
		return "api_key", true
	}
	// Per-machine credential for provide-scoped RPCs.
	if strings.EqualFold(key, "X-Toolplane-Machine-Token") {
		return "x-toolplane-machine-token", true
	}
	// W3C trace correlation: forwarded so the server can attach the
	// caller's trace context to request logs. The gateway already forwards
	// its (validated) traceparent the same way.
	if strings.EqualFold(key, "Traceparent") {
		return "traceparent", true
	}
	return runtime.DefaultHeaderMatcher(key)
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

type healthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	// Breaker and throttle statistics are deliberately omitted: /health is
	// unauthenticated, and the counters disclose caller behavior to anyone
	// who can reach the endpoint. Operators scrape the server's
	// Prometheus /metrics instead.
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{ResponseWriter: w, status: http.StatusOK}
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Flush() {
	flusher, ok := r.ResponseWriter.(http.Flusher)
	if !ok {
		log.Printf("ERROR: Flush not supported in %T", r.ResponseWriter)
		return
	}
	flusher.Flush()
}

// proxyControlMiddleware coordinates rate limiting and circuit breaker checks before dispatching to the mux.
func proxyControlMiddleware(cfg proxyConfig, breaker *CircuitBreakerManager, rlm *RateLimiterManager, tracker *ThrottleTracker, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Rate limiting and throttle tracking key on a hash of the credential,
		// never the raw secret.
		apiKey := hashClientSecret(extractAPIKey(r))
		clientIP := extractClientIP(r, cfg.trustedProxy)

		if rlm != nil {
			allowed, wait, rlReason, rlMessage := rlm.Allow(apiKey, clientIP)
			if !allowed {
				applied := applyRetryAfterHeader(w, wait)
				if tracker != nil {
					tracker.Record(rlReason, applied, apiKey, clientIP, rlMessage)
				}
				http.Error(w, rlMessage, http.StatusTooManyRequests)
				return
			}
		}

		done, release, wait, err := breaker.Begin()
		if err != nil {
			applied := applyRetryAfterHeader(w, wait)
			var reason ThrottleReason
			var message string
			switch err {
			case errTooManyConcurrent:
				reason = ThrottleReasonConcurrency
				message = err.Error()
			case gobreaker.ErrOpenState:
				reason = ThrottleReasonCircuitOpen
				message = "circuit breaker open"
			case gobreaker.ErrTooManyRequests:
				reason = ThrottleReasonCircuitProbe
				message = "circuit breaker probe limit reached"
			default:
				reason = ThrottleReasonUnknown
				message = err.Error()
			}

			if tracker != nil {
				tracker.Record(reason, applied, apiKey, clientIP, message)
			}
			http.Error(w, message, http.StatusTooManyRequests)
			return
		}
		defer release()

		recorder := newResponseRecorder(w)
		defer func() {
			if rec := recover(); rec != nil {
				if done != nil {
					done(false)
				}
				panic(rec)
			}
			// Only gateway-visible backend failures trip the breaker:
			// 502/503/504. Client-induced 4xx/5xx (invalid payloads mapped to
			// INTERNAL) and app-level 500s say nothing about backend health —
			// counting them would open the gateway-wide breaker on healthy
			// backends.
			if done != nil {
				status := recorder.status
				// done(success): only 502/503/504 are gateway-visible backend
				// failures; a 500 or client 4xx is a success for the breaker.
				success := status != http.StatusBadGateway &&
					status != http.StatusServiceUnavailable &&
					status != http.StatusGatewayTimeout
				done(success)
			}
		}()

		next.ServeHTTP(recorder, r)
	})
}

// extractAPIKey reads the caller credential from headers only. Query-string
// credentials (?api_key=) are deliberately not accepted: URLs land in access
// logs, browser history, Referer headers, and APM traces.
func extractAPIKey(r *http.Request) string {
	if key := r.Header.Get("Grpc-Metadata-api_key"); key != "" {
		return key
	}
	if key := r.Header.Get("X-API-Key"); key != "" {
		return key
	}
	// The documented auth path: Authorization: Bearer <key>. Without this
	// fallback, bearer-key callers bypass per-key rate limiting entirely.
	// RFC 7235: the auth-scheme token is case-insensitive.
	if auth := r.Header.Get("Authorization"); auth != "" {
		if scheme, rest, found := strings.Cut(auth, " "); found && strings.EqualFold(scheme, "bearer") {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

// hashClientSecret derives a stable non-secret identity from an API key for
// rate-limiter buckets and throttle observability, so raw credentials never
// live in limiter maps or tracker snapshots.
func hashClientSecret(secret string) string {
	if secret == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:8])
}

// extractClientIP resolves the client identity. X-Forwarded-For is only
// trusted when the operator explicitly declared a terminating reverse proxy
// (TOOLPLANE_TRUSTED_PROXY=1); otherwise the header is client-controlled and
// RemoteAddr is the only trustworthy source.
func extractClientIP(r *http.Request, trustForwardedFor bool) string {
	if trustForwardedFor {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			if len(parts) > 0 {
				return strings.TrimSpace(parts[0])
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

// deadlinePolicyMiddleware owns the gateway's deadline policy. The
// grpc-gateway runtime honors a client-supplied Grpc-Timeout header — it
// becomes the context deadline before the backend dial interceptor runs,
// so any HTTP caller could raise or shrink the effective bound in either
// direction (including killing a streaming replay mid-flight). The header
// is stripped here instead: HTTP callers bound themselves with their own
// client timeouts, and the per-method dial policy (see unaryDeadline)
// decides everything else.
func deadlinePolicyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("Grpc-Timeout")
		next.ServeHTTP(w, r)
	})
}

// proxyUnaryDeadline is the default unary bound: fast management calls
// (create/get/list) complete in milliseconds, so 30 seconds is generously
// above their p99 while still failing fast on a stuck backend.
const proxyUnaryDeadline = 30 * time.Second

// waitMethodBackstop bounds the two wait-capable entrypoints. The server
// itself rejects waits above its ceiling (see
// ExecuteToolRequest.wait_timeout_seconds, 3600s); the backstop sits above
// it so an honest max-length wait completes, while a server bug still
// cannot hang the gateway connection forever.
const waitMethodBackstop = time.Hour + 30*time.Second

// waitMethods are the full method names whose unary bound is the wait
// backstop instead of the default. Both are ToolService methods;
// InvokeTool delegates to ExecuteTool server-side.
var waitMethods = map[string]bool{
	"/api.v1.ToolService/InvokeTool":  true,
	"/api.v1.ToolService/ExecuteTool": true,
}

// unaryDeadline picks the dial deadline for one unary method: the wait
// backstop for the long-poll entrypoints, the fast-call default for
// everything else.
func unaryDeadline(method string) time.Duration {
	if waitMethods[method] {
		return waitMethodBackstop
	}
	return proxyUnaryDeadline
}

// unaryDeadlineInterceptor applies the per-method deadline policy to the
// backend dial — but only as a default. An existing deadline (an HTTP
// client's own timeout, propagated through the request context) always
// wins; the client-supplied Grpc-Timeout override is stripped at the edge
// (see deadlinePolicyMiddleware), so the policy here is gateway-owned.
func unaryDeadlineInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, unaryDeadline(method))
		defer cancel()
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

// newProxyRootHandler assembles the HTTP chain: CORS, then rate limiting
// and circuit breaking, then the deadline policy, then the gateway mux.
func newProxyRootHandler(
	cfg proxyConfig,
	breaker *CircuitBreakerManager,
	rateLimiter *RateLimiterManager,
	throttleTracker *ThrottleTracker,
	apiHandler http.Handler,
) http.Handler {
	root := http.NewServeMux()
	root.Handle("/", corsMiddleware(cfg, proxyControlMiddleware(cfg, breaker, rateLimiter, throttleTracker, deadlinePolicyMiddleware(apiHandler))))
	root.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeHealthResponse(w, breaker, rateLimiter, throttleTracker, time.Now().UTC())
	})
	return root
}

func buildHealthResponse(
	breaker *CircuitBreakerManager,
	rateLimiter *RateLimiterManager,
	throttleTracker *ThrottleTracker,
	now time.Time,
) (healthResponse, int) {
	// Rate-limit and throttle counters are deliberately not echoed here:
	// /health is unauthenticated, and the counters disclose caller behavior
	// to anyone who can reach the endpoint. Operators scrape the server's
	// Prometheus /metrics instead.
	_ = rateLimiter
	_ = throttleTracker

	response := healthResponse{
		Status:    "ok",
		Timestamp: now,
	}

	statusCode := http.StatusOK
	if breaker != nil && breaker.IsOpen() {
		response.Status = "degraded"
		statusCode = http.StatusServiceUnavailable
	}

	return response, statusCode
}

func writeHealthResponse(
	w http.ResponseWriter,
	breaker *CircuitBreakerManager,
	rateLimiter *RateLimiterManager,
	throttleTracker *ThrottleTracker,
	now time.Time,
) {
	response, statusCode := buildHealthResponse(breaker, rateLimiter, throttleTracker, now)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode health response: %v", err)
	}
}

// Options carries the gateway's operator-facing settings; entry points
// (cmd/toolplane gateway serve) parse their own flag surfaces into it.
type Options struct {
	Listen                string
	Backend               string
	MaxMsgSize            int
	MaxConcurrentRequests int64
	APIRate               float64
	APIBurst              int
	IPRate                float64
	IPBurst               int
	TLSCertFile           string
	TLSKeyFile            string
}

// DefaultOptions returns the standalone binary's defaults.
func DefaultOptions() Options {
	return Options{
		Listen:                ":8080",
		Backend:               "localhost:9001",
		MaxMsgSize:            model.MaxChunkBatchBytes + 1<<20,
		MaxConcurrentRequests: 1000,
	}
}

// Run owns process-level lifecycle: signals become context cancellation.
func Run(opts Options) int {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	return RunContext(ctx, opts)
}

// RunContext drives the gateway until the context is cancelled or a serve
// error forces teardown; it returns the process exit code.
func RunContext(ctx context.Context, opts Options) int {
	cfg, err := loadProxyConfig()
	if err != nil {
		log.Printf("invalid proxy configuration: %v", err)
		return 1
	}

	// Production must not serve client-facing plaintext: either terminate TLS
	// here, or the operator explicitly declares a TLS-terminating reverse
	// proxy via TOOLPLANE_TRUSTED_PROXY=1.
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

	// Create circuit breaker/backpressure manager
	breaker := NewCircuitBreakerManager(opts.MaxConcurrentRequests)
	rateLimiter := NewRateLimiterManager(ctx, rate.Limit(opts.APIRate), opts.APIBurst, rate.Limit(opts.IPRate), opts.IPBurst)
	throttleTracker := NewThrottleTracker()

	apiDisabled := opts.APIRate <= 0 || opts.APIBurst <= 0
	ipDisabled := opts.IPRate <= 0 || opts.IPBurst <= 0
	switch {
	case apiDisabled && ipDisabled:
		log.Println("WARNING: rate limiting is disabled for every dimension (rate or burst set to 0); callers share the backend without throttling")
	case apiDisabled:
		log.Println("WARNING: per-API-key rate limiting is disabled (api-rate or api-burst is 0)")
	case ipDisabled:
		log.Println("WARNING: per-IP rate limiting is disabled (ip-rate or ip-burst is 0)")
	}

	transportCredentials, err := backendTransportCredentials(cfg)
	if err != nil {
		log.Printf("invalid backend TLS configuration: %v", err)
		return 1
	}

	// Setup gRPC connection options with backpressure controls
	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCredentials),

		// Limit message sizes to prevent memory exhaustion
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(opts.MaxMsgSize),
			grpc.MaxCallSendMsgSize(opts.MaxMsgSize),
		),
		// Configure keepalive settings
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // ping server every 10 seconds
			Timeout:             3 * time.Second,  // timeout for pings
			PermitWithoutStream: true,             // allow pings even without active streams
		}),
		// Add context propagation
		grpc.WithUnaryInterceptor(unaryDeadlineInterceptor),
		// Set stream buffer sizes for flow control
		grpc.WithReadBufferSize(1024 * 64),  // 64KB read buffer
		grpc.WithWriteBufferSize(1024 * 64), // 64KB write buffer
	}

	// Register gRPC-Gateway mux with custom options
	muxOpts := []runtime.ServeMuxOption{
		// Forward auth headers to gRPC metadata
		runtime.WithIncomingHeaderMatcher(authHeaderMatcher),
		// Configure context propagation
		runtime.WithMetadata(func(ctx context.Context, r *http.Request) metadata.MD {
			md := metadata.New(nil)
			// Add X-Forwarded headers
			if clientIP := r.Header.Get("X-Forwarded-For"); clientIP != "" {
				md.Set("x-forwarded-for", clientIP)
			}
			if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
				md.Set("x-forwarded-proto", proto)
			}
			if apiKey := extractAPIKey(r); apiKey != "" {
				md.Set("api_key", apiKey)
			}
			if directIP := extractClientIP(r, cfg.trustedProxy); directIP != "" {
				md.Set("client-ip", directIP)
			}
			return md
		}),
	}

	mux := runtime.NewServeMux(muxOpts...)

	// Register each API service handler
	if err := gw.RegisterToolServiceHandlerFromEndpoint(ctx, mux, opts.Backend, dialOptions); err != nil {
		log.Printf("failed to register ToolService handler: %v", err)
		return 1
	}
	if err := gw.RegisterSessionsServiceHandlerFromEndpoint(ctx, mux, opts.Backend, dialOptions); err != nil {
		log.Printf("failed to register SessionsService handler: %v", err)
		return 1
	}
	if err := gw.RegisterMachinesServiceHandlerFromEndpoint(ctx, mux, opts.Backend, dialOptions); err != nil {
		log.Printf("failed to register MachinesService handler: %v", err)
		return 1
	}
	if err := gw.RegisterRequestsServiceHandlerFromEndpoint(ctx, mux, opts.Backend, dialOptions); err != nil {
		log.Printf("failed to register RequestsService handler: %v", err)
		return 1
	}
	if err := gw.RegisterTasksServiceHandlerFromEndpoint(ctx, mux, opts.Backend, dialOptions); err != nil {
		log.Printf("failed to register TasksService handler: %v", err)
		return 1
	}

	log.Printf("JSON gateway listening on %s → gRPC %s (env=%s cors=%s backend=%s)", opts.Listen, opts.Backend, cfg.environment, cfg.corsSummary(), cfg.backendSecuritySummary())

	root := newProxyRootHandler(cfg, breaker, rateLimiter, throttleTracker, mux)

	// Start HTTP server. WriteTimeout stays unset on purpose: the gateway
	// proxies server-streaming RPCs (ResumeStream/StreamExecuteTool) whose
	// responses can legitimately outlive any fixed write deadline.
	httpServer := &http.Server{
		Addr:              opts.Listen,
		Handler:           root,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		// IdleTimeout polices parked keep-alive connections; WriteTimeout
		// stays unset because streaming responses outlive any fixed deadline.
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
			log.Printf("proxy serve error: %v", err)
			return 1
		}
		return 0
	case <-ctx.Done():
		log.Println("shutdown requested, draining proxy")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("proxy shutdown error: %v", err)
		}
		return 0
	}
}
