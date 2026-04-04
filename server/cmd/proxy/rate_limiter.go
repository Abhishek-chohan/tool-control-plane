package main

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type keyedLimiter struct {
	mu      sync.Mutex
	limit   rate.Limit
	burst   int
	ttl     time.Duration
	entries map[string]*clientLimiter
}

func newKeyedLimiter(limit rate.Limit, burst int, ttl time.Duration) *keyedLimiter {
	return &keyedLimiter{
		limit:   limit,
		burst:   burst,
		ttl:     ttl,
		entries: make(map[string]*clientLimiter),
	}
}

func (kl *keyedLimiter) allow(key string, now time.Time) (bool, time.Duration) {
	if key == "" {
		return true, 0
	}
	kl.mu.Lock()
	entry, ok := kl.entries[key]
	if !ok {
		entry = &clientLimiter{limiter: rate.NewLimiter(kl.limit, kl.burst)}
		kl.entries[key] = entry
	}
	entry.lastSeen = now
	if entry.limiter.AllowN(now, 1) {
		kl.mu.Unlock()
		return true, 0
	}
	reservation := entry.limiter.ReserveN(now, 1)
	if !reservation.OK() {
		kl.mu.Unlock()
		return false, kl.ttl
	}
	delay := reservation.DelayFrom(now)
	reservation.CancelAt(now)
	kl.mu.Unlock()
	if delay < time.Second {
		delay = time.Second
	}
	return false, delay
}

func (kl *keyedLimiter) cleanup(now time.Time) {
	cutoff := now.Add(-kl.ttl)
	kl.mu.Lock()
	for key, entry := range kl.entries {
		if entry.lastSeen.Before(cutoff) {
			delete(kl.entries, key)
		}
	}
	kl.mu.Unlock()
}

// RateLimiterManager handles rate limiting for API keys and client IPs.
type RateLimiterManager struct {
	apiLimiter *keyedLimiter
	ipLimiter  *keyedLimiter

	totalRejected atomic.Int64
}

func NewRateLimiterManager(ctx context.Context, apiRate rate.Limit, apiBurst int, ipRate rate.Limit, ipBurst int) *RateLimiterManager {
	if (apiRate <= 0 || apiBurst <= 0) && (ipRate <= 0 || ipBurst <= 0) {
		return nil
	}

	manager := &RateLimiterManager{}
	ttl := 10 * time.Minute
	if apiRate > 0 && apiBurst > 0 {
		manager.apiLimiter = newKeyedLimiter(apiRate, apiBurst, ttl)
	}
	if ipRate > 0 && ipBurst > 0 {
		manager.ipLimiter = newKeyedLimiter(ipRate, ipBurst, ttl)
	}
	if manager.apiLimiter != nil || manager.ipLimiter != nil {
		go func() {
			ticker := time.NewTicker(ttl)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case now := <-ticker.C:
					if manager.apiLimiter != nil {
						manager.apiLimiter.cleanup(now)
					}
					if manager.ipLimiter != nil {
						manager.ipLimiter.cleanup(now)
					}
				}
			}
		}()
	}
	return manager
}

// Allow evaluates rate limits for both API key and IP address.
func (m *RateLimiterManager) Allow(apiKey, ip string) (bool, time.Duration, ThrottleReason, string) {
	if m == nil {
		return true, 0, ThrottleReasonNone, ""
	}

	now := time.Now()
	var wait time.Duration
	if m.apiLimiter != nil && apiKey != "" {
		allowed, delay := m.apiLimiter.allow(apiKey, now)
		if !allowed {
			m.totalRejected.Add(1)
			return false, delay, ThrottleReasonAPIRate, "API key rate limit exceeded"
		}
		wait = delay
	}
	if m.ipLimiter != nil && ip != "" {
		allowed, delay := m.ipLimiter.allow(ip, now)
		if !allowed {
			m.totalRejected.Add(1)
			if delay > wait {
				wait = delay
			}
			return false, wait, ThrottleReasonIPRate, "IP rate limit exceeded"
		}
	}
	return true, wait, ThrottleReasonNone, ""
}

func (m *RateLimiterManager) Stats() int64 {
	if m == nil {
		return 0
	}
	return m.totalRejected.Load()
}
