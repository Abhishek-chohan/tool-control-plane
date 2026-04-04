package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func tokenFromMetadata(md metadata.MD) (string, error) {
	authHeaders := md.Get("authorization")
	if len(authHeaders) > 0 && strings.TrimSpace(authHeaders[0]) != "" {
		return authHeaders[0], nil
	}

	apiKeyHeaders := md.Get("api_key")
	if len(apiKeyHeaders) > 0 && strings.TrimSpace(apiKeyHeaders[0]) != "" {
		return apiKeyHeaders[0], nil
	}

	return "", errors.New("authorization token not supplied")
}

func normalizeToken(token string) string {
	trimmed := strings.TrimSpace(token)
	if strings.HasPrefix(strings.ToLower(trimmed), "bearer ") {
		return strings.TrimSpace(trimmed[7:])
	}
	return trimmed
}

func redactToken(token string) string {
	if token == "" {
		return "<empty>"
	}
	if len(token) <= 8 {
		return "<redacted>"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

// Simple cache entry with expiration
type cacheEntry struct {
	valid     bool
	expiresAt time.Time
}

// SupabaseValidator validates API keys against Supabase.
type SupabaseValidator struct {
	baseURL    string
	serviceKey string
	httpClient *http.Client
	debug      bool

	// Simple in-memory cache
	cache    map[string]cacheEntry
	cacheMu  sync.RWMutex
	cacheTTL time.Duration
}

// NewSupabaseValidator creates a validator that checks API keys in the Supabase api_keys table.
func NewSupabaseValidator(baseURL, serviceKey string, debug bool) *SupabaseValidator {
	baseURL = strings.TrimRight(baseURL, "/")

	return &SupabaseValidator{
		baseURL:    baseURL,
		serviceKey: serviceKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		debug:    debug,
		cache:    make(map[string]cacheEntry),
		cacheTTL: 5 * time.Minute,
	}
}

// Validate returns true if the given API key exists in the api_keys table.
func (v *SupabaseValidator) Validate(ctx context.Context, token string) bool {
	apiKey := normalizeToken(token)
	if apiKey == "" {
		return false
	}

	v.cacheMu.RLock()
	entry, found := v.cache[apiKey]
	v.cacheMu.RUnlock()

	now := time.Now()
	if found && now.Before(entry.expiresAt) {
		if v.debug {
			log.Printf("supabase cache hit token=%s valid=%v", redactToken(apiKey), entry.valid)
		}
		return entry.valid
	}

	if v.debug {
		log.Printf("supabase validation cache miss token=%s", redactToken(apiKey))
	}

	encodedKey := url.QueryEscape(apiKey)
	reqURL := fmt.Sprintf("%s/rest/v1/api_keys?key=eq.%s&select=key&limit=1",
		v.baseURL, encodedKey)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return false
	}

	req.Header.Add("apikey", v.serviceKey)
	req.Header.Add("Authorization", "Bearer "+v.serviceKey)
	req.Header.Add("Accept", "application/json")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		log.Printf("HTTP request failed: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("Supabase API error: status=%d token=%s", resp.StatusCode, redactToken(apiKey))
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return false
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(body, &results); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		return false
	}

	isValid := len(results) > 0

	v.cacheMu.Lock()
	v.cache[apiKey] = cacheEntry{
		valid:     isValid,
		expiresAt: now.Add(v.cacheTTL),
	}
	v.cacheMu.Unlock()

	if v.debug {
		log.Printf("supabase validation result=%v token=%s cache_ttl=%v result_count=%d", isValid, redactToken(apiKey), v.cacheTTL, len(results))
	}

	return isValid
}

// UnaryAPIKeyInterceptor adds API key validation to all unary RPCs.
func UnaryAPIKeyInterceptor(validate func(context.Context, string) bool) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("missing metadata")
		}

		token, err := tokenFromMetadata(md)
		if err != nil {
			return nil, err
		}

		if !validate(ctx, token) {
			return nil, errors.New("invalid API key")
		}

		return handler(ctx, req)
	}
}

// StreamAPIKeyInterceptor adds API key validation to all streaming RPCs.
func StreamAPIKeyInterceptor(validate func(context.Context, string) bool) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return errors.New("missing metadata")
		}

		token, err := tokenFromMetadata(md)
		if err != nil {
			return err
		}

		if !validate(ss.Context(), token) {
			return errors.New("invalid API key")
		}

		return handler(srv, ss)
	}
}
