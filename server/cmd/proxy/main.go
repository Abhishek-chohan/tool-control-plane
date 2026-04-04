package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	gw "toolplane/proto"
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
	return runtime.DefaultHeaderMatcher(key)
}

type responseRecorder struct {
	http.ResponseWriter
	status int
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
func proxyControlMiddleware(breaker *CircuitBreakerManager, rlm *RateLimiterManager, tracker *ThrottleTracker, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		apiKey := extractAPIKey(r)
		clientIP := extractClientIP(r)

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
			if done != nil {
				done(recorder.status < 500)
			}
		}()

		next.ServeHTTP(recorder, r)
	})
}

func extractAPIKey(r *http.Request) string {
	if key := r.Header.Get("Grpc-Metadata-api_key"); key != "" {
		return key
	}
	if key := r.Header.Get("X-API-Key"); key != "" {
		return key
	}
	if key := r.URL.Query().Get("api_key"); key != "" {
		return key
	}
	return ""
}

func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func main() {
	cfg, err := loadProxyConfig()
	if err != nil {
		log.Fatalf("invalid proxy configuration: %v", err)
	}

	// HTTP server listen address for JSON API
	httpListen := flag.String("listen", ":8080", "HTTP listen address for JSON gateway")
	// gRPC backend endpoint
	grpcEndpoint := flag.String("backend", "localhost:9001", "gRPC server endpoint")
	// Max message size settings (for backpressure management)
	maxMsgSize := flag.Int("max-msg-size", 4*1024*1024, "Maximum message size in bytes")
	// Max concurrent requests
	maxConcurrentRequests := flag.Int64("max-concurrent", 1000, "Maximum concurrent requests")
	// Rate limiting controls
	apiRate := flag.Float64("api-rate", 0, "Maximum requests per second per API key (0 disables)")
	apiBurst := flag.Int("api-burst", 0, "Burst size per API key when rate limiting is enabled")
	ipRate := flag.Float64("ip-rate", 0, "Maximum requests per second per client IP (0 disables)")
	ipBurst := flag.Int("ip-burst", 0, "Burst size per client IP when rate limiting is enabled")
	flag.Parse()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create circuit breaker/backpressure manager
	breaker := NewCircuitBreakerManager(*maxConcurrentRequests)
	rateLimiter := NewRateLimiterManager(ctx, rate.Limit(*apiRate), *apiBurst, rate.Limit(*ipRate), *ipBurst)
	throttleTracker := NewThrottleTracker()

	transportCredentials := credentials.TransportCredentials(insecure.NewCredentials())
	if !cfg.allowInsecureBackend {
		transportCredentials = credentials.NewTLS(&tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: cfg.backendTLSServerName,
		})
	}

	// Setup gRPC connection options with backpressure controls
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCredentials),

		// Limit message sizes to prevent memory exhaustion
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(*maxMsgSize),
			grpc.MaxCallSendMsgSize(*maxMsgSize),
		),
		// Configure keepalive settings
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // ping server every 10 seconds
			Timeout:             3 * time.Second,  // timeout for pings
			PermitWithoutStream: true,             // allow pings even without active streams
		}),
		// Add context propagation
		grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			// Apply deadline if not already set
			if _, ok := ctx.Deadline(); !ok {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
				defer cancel()
			}
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
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
			if directIP := extractClientIP(r); directIP != "" {
				md.Set("client-ip", directIP)
			}
			return md
		}),
	}

	mux := runtime.NewServeMux(muxOpts...)

	// Register each API service handler
	if err := gw.RegisterToolServiceHandlerFromEndpoint(ctx, mux, *grpcEndpoint, opts); err != nil {
		log.Fatalf("Failed to register ToolService handler: %v", err)
	}
	if err := gw.RegisterSessionsServiceHandlerFromEndpoint(ctx, mux, *grpcEndpoint, opts); err != nil {
		log.Fatalf("Failed to register SessionsService handler: %v", err)
	}
	if err := gw.RegisterMachinesServiceHandlerFromEndpoint(ctx, mux, *grpcEndpoint, opts); err != nil {
		log.Fatalf("Failed to register MachinesService handler: %v", err)
	}
	if err := gw.RegisterRequestsServiceHandlerFromEndpoint(ctx, mux, *grpcEndpoint, opts); err != nil {
		log.Fatalf("Failed to register RequestsService handler: %v", err)
	}
	if err := gw.RegisterTasksServiceHandlerFromEndpoint(ctx, mux, *grpcEndpoint, opts); err != nil {
		log.Fatalf("Failed to register TasksService handler: %v", err)
	}

	log.Printf("JSON gateway listening on %s → gRPC %s (env=%s cors=%s backend=%s)", *httpListen, *grpcEndpoint, cfg.environment, cfg.corsSummary(), cfg.backendSecuritySummary())

	// Create root handler
	root := http.NewServeMux()

	// Add CORS and proxy control middleware to API
	root.Handle("/", corsMiddleware(cfg, proxyControlMiddleware(breaker, rateLimiter, throttleTracker, mux)))

	// Add health check endpoint with diagnostic payload
	root.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		stats := breaker.Stats()
		throttleStats := ThrottleSnapshot{}
		if throttleTracker != nil {
			throttleStats = throttleTracker.Snapshot()
		}
		response := struct {
			Status           string              `json:"status"`
			Circuit          CircuitBreakerStats `json:"circuit"`
			RateLimitRejects int64               `json:"rateLimitRejects"`
			Throttle         ThrottleSnapshot    `json:"throttle"`
			Timestamp        time.Time           `json:"timestamp"`
		}{
			Status:           "ok",
			Circuit:          stats,
			RateLimitRejects: 0,
			Throttle:         throttleStats,
			Timestamp:        time.Now().UTC(),
		}

		statusCode := http.StatusOK
		if breaker.IsOpen() {
			response.Status = "degraded"
			statusCode = http.StatusServiceUnavailable
		}
		if rateLimiter != nil {
			response.RateLimitRejects = rateLimiter.Stats()
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("failed to encode health response: %v", err)
		}
	})

	// Start HTTP server
	log.Fatal(http.ListenAndServe(*httpListen, root))
}
