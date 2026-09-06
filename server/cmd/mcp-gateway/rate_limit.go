package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Gateway rate limiting mirrors cmd/proxy's controls: per-API-key and
// per-client-IP token buckets, disabled when the rates are zero. Identities
// are hashed before use so raw credentials never live in limiter state.

type gatewayLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type gatewayKeyedLimiter struct {
	mu      sync.Mutex
	limit   rate.Limit
	burst   int
	entries map[string]*gatewayLimiterEntry
}

func newGatewayKeyedLimiter(limit rate.Limit, burst int) *gatewayKeyedLimiter {
	return &gatewayKeyedLimiter{
		limit:   limit,
		burst:   burst,
		entries: make(map[string]*gatewayLimiterEntry),
	}
}

func (kl *gatewayKeyedLimiter) allow(key string, now time.Time) bool {
	if key == "" {
		return true
	}
	kl.mu.Lock()
	defer kl.mu.Unlock()
	// Opportunistic eviction keeps the map bounded without a background loop.
	if len(kl.entries) > 4096 {
		cutoff := now.Add(-10 * time.Minute)
		for entryKey, entry := range kl.entries {
			if entry.lastSeen.Before(cutoff) {
				delete(kl.entries, entryKey)
			}
		}
	}
	entry, ok := kl.entries[key]
	if !ok {
		entry = &gatewayLimiterEntry{limiter: rate.NewLimiter(kl.limit, kl.burst)}
		kl.entries[key] = entry
	}
	entry.lastSeen = now
	return entry.limiter.AllowN(now, 1)
}

type gatewayRateLimiter struct {
	apiLimiter *gatewayKeyedLimiter
	ipLimiter  *gatewayKeyedLimiter
}

func newGatewayRateLimiter(apiRate rate.Limit, apiBurst int, ipRate rate.Limit, ipBurst int) *gatewayRateLimiter {
	limiter := &gatewayRateLimiter{}
	if apiRate > 0 && apiBurst > 0 {
		limiter.apiLimiter = newGatewayKeyedLimiter(apiRate, apiBurst)
	}
	if ipRate > 0 && ipBurst > 0 {
		limiter.ipLimiter = newGatewayKeyedLimiter(ipRate, ipBurst)
	}
	if limiter.apiLimiter == nil && limiter.ipLimiter == nil {
		return nil
	}
	return limiter
}

func hashGatewaySecret(secret string) string {
	if secret == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:8])
}

func gatewayAPIKey(r *http.Request) string {
	if key := r.Header.Get("X-API-Key"); key != "" {
		return key
	}
	authorization := r.Header.Get("Authorization")
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(authorization), "Bearer "))
}

func gatewayClientIP(r *http.Request, trustForwardedFor bool) string {
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

func gatewayRateLimitMiddleware(limiter *gatewayRateLimiter, trustForwardedFor bool, next http.Handler) http.Handler {
	if limiter == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		if limiter.apiLimiter != nil {
			if !limiter.apiLimiter.allow(hashGatewaySecret(gatewayAPIKey(r)), now) {
				w.Header().Set("Retry-After", "1")
				http.Error(w, "API key rate limit exceeded", http.StatusTooManyRequests)
				return
			}
		}
		if limiter.ipLimiter != nil {
			if !limiter.ipLimiter.allow(gatewayClientIP(r, trustForwardedFor), now) {
				w.Header().Set("Retry-After", "1")
				http.Error(w, "IP rate limit exceeded", http.StatusTooManyRequests)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
