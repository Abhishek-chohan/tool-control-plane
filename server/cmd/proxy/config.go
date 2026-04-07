package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type proxyConfig struct {
	environment          string
	allowAnyOrigin       bool
	allowedOrigins       map[string]struct{}
	allowInsecureBackend bool
	backendTLSServerName string
	backendTLSCAFile     string
}

func loadProxyConfig() (proxyConfig, error) {
	environment := normalizedEnvValue(os.Getenv("TOOLPLANE_ENV_MODE"), "development")
	if environment != "development" && environment != "test" && environment != "production" {
		return proxyConfig{}, fmt.Errorf("TOOLPLANE_ENV_MODE must be one of development, test, or production")
	}

	allowInsecureBackend := boolEnv("TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND", environment != "production")
	if environment == "production" && allowInsecureBackend {
		return proxyConfig{}, fmt.Errorf("TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND is not allowed when TOOLPLANE_ENV_MODE=production")
	}

	rawOrigins := strings.TrimSpace(os.Getenv("TOOLPLANE_PROXY_ALLOWED_ORIGINS"))
	cfg := proxyConfig{
		environment:          environment,
		allowedOrigins:       make(map[string]struct{}),
		allowInsecureBackend: allowInsecureBackend,
		backendTLSServerName: strings.TrimSpace(os.Getenv("TOOLPLANE_PROXY_BACKEND_TLS_SERVER_NAME")),
		backendTLSCAFile:     strings.TrimSpace(os.Getenv("TOOLPLANE_PROXY_BACKEND_TLS_CA_FILE")),
	}

	if rawOrigins == "" {
		if environment == "development" || environment == "test" {
			cfg.allowAnyOrigin = true
			return cfg, nil
		}
		return proxyConfig{}, fmt.Errorf("TOOLPLANE_PROXY_ALLOWED_ORIGINS is required outside development or test mode")
	}

	for _, origin := range strings.Split(rawOrigins, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		if trimmed == "*" {
			if environment == "production" {
				return proxyConfig{}, fmt.Errorf("TOOLPLANE_PROXY_ALLOWED_ORIGINS=* is not allowed in production")
			}
			cfg.allowAnyOrigin = true
			cfg.allowedOrigins = map[string]struct{}{}
			return cfg, nil
		}
		cfg.allowedOrigins[trimmed] = struct{}{}
	}

	if len(cfg.allowedOrigins) == 0 && !cfg.allowAnyOrigin {
		return proxyConfig{}, fmt.Errorf("TOOLPLANE_PROXY_ALLOWED_ORIGINS must contain at least one origin")
	}

	return cfg, nil
}

func (cfg proxyConfig) matchOrigin(origin string) (string, bool) {
	if cfg.allowAnyOrigin {
		return "*", true
	}
	_, ok := cfg.allowedOrigins[origin]
	return origin, ok
}

func (cfg proxyConfig) corsSummary() string {
	if cfg.allowAnyOrigin {
		return "* (development or test only)"
	}
	origins := make([]string, 0, len(cfg.allowedOrigins))
	for origin := range cfg.allowedOrigins {
		origins = append(origins, origin)
	}
	sort.Strings(origins)
	return strings.Join(origins, ",")
}

func (cfg proxyConfig) backendSecuritySummary() string {
	if cfg.allowInsecureBackend {
		return "insecure gRPC backend (development or test only)"
	}
	if cfg.backendTLSServerName != "" && cfg.backendTLSCAFile != "" {
		return "tls:" + cfg.backendTLSServerName + "+custom-ca"
	}
	if cfg.backendTLSServerName != "" {
		return "tls:" + cfg.backendTLSServerName
	}
	if cfg.backendTLSCAFile != "" {
		return "tls+custom-ca"
	}
	return "tls"
}

func normalizedEnvValue(value, fallback string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func boolEnv(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	case "":
		return fallback
	default:
		return fallback
	}
}
