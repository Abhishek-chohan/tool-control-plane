package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"toolplane/cmd/server/auth"
)

type serverConfig struct {
	environment        string
	authMode           string
	authDebug          bool
	fixedAPIKey        string
	supabaseURL        string
	supabaseServiceKey string
	storageMode        string
	databaseURL        string
}

func loadServerConfig() (serverConfig, error) {
	cfg := serverConfig{
		environment:        normalizedEnvValue(os.Getenv("TOOLPLANE_ENV_MODE"), "development"),
		authMode:           normalizedEnvValue(os.Getenv("TOOLPLANE_AUTH_MODE"), "disabled"),
		authDebug:          boolEnv("TOOLPLANE_AUTH_DEBUG", false),
		fixedAPIKey:        strings.TrimSpace(os.Getenv("TOOLPLANE_AUTH_FIXED_API_KEY")),
		supabaseURL:        strings.TrimSpace(os.Getenv("TOOLPLANE_SUPABASE_URL")),
		supabaseServiceKey: strings.TrimSpace(os.Getenv("TOOLPLANE_SUPABASE_SERVICE_KEY")),
		storageMode:        normalizedEnvValue(os.Getenv("TOOLPLANE_STORAGE_MODE"), ""),
		databaseURL:        strings.TrimSpace(os.Getenv("TOOLPLANE_DATABASE_URL")),
	}

	switch cfg.environment {
	case "development", "test", "production":
	default:
		return serverConfig{}, fmt.Errorf("TOOLPLANE_ENV_MODE must be one of development, test, or production")
	}

	switch cfg.authMode {
	case "disabled":
		if cfg.environment == "production" {
			return serverConfig{}, fmt.Errorf("TOOLPLANE_AUTH_MODE=disabled is not allowed when TOOLPLANE_ENV_MODE=production")
		}
	case "fixed":
		if cfg.fixedAPIKey == "" {
			return serverConfig{}, fmt.Errorf("TOOLPLANE_AUTH_MODE=fixed requires TOOLPLANE_AUTH_FIXED_API_KEY")
		}
		if cfg.environment == "production" {
			return serverConfig{}, fmt.Errorf("TOOLPLANE_AUTH_MODE=fixed is for development or test only")
		}
	case "supabase":
		if cfg.supabaseURL == "" || cfg.supabaseServiceKey == "" {
			return serverConfig{}, fmt.Errorf("TOOLPLANE_AUTH_MODE=supabase requires TOOLPLANE_SUPABASE_URL and TOOLPLANE_SUPABASE_SERVICE_KEY")
		}
	default:
		return serverConfig{}, fmt.Errorf("TOOLPLANE_AUTH_MODE must be one of disabled, fixed, or supabase")
	}

	effectiveStorageMode := cfg.storageMode
	if effectiveStorageMode == "" && cfg.databaseURL != "" {
		effectiveStorageMode = "postgres"
	}

	switch effectiveStorageMode {
	case "", "memory", "in-memory", "inmemory", "postgres":
	default:
		return serverConfig{}, fmt.Errorf("TOOLPLANE_STORAGE_MODE must be one of memory or postgres")
	}

	if effectiveStorageMode == "postgres" && cfg.databaseURL == "" {
		return serverConfig{}, fmt.Errorf("TOOLPLANE_STORAGE_MODE=postgres requires TOOLPLANE_DATABASE_URL")
	}
	if cfg.environment == "production" && effectiveStorageMode != "postgres" {
		return serverConfig{}, fmt.Errorf("TOOLPLANE_ENV_MODE=production requires TOOLPLANE_STORAGE_MODE=postgres or TOOLPLANE_DATABASE_URL")
	}

	return cfg, nil
}

func (cfg serverConfig) buildValidator() (func(context.Context, string) bool, string, error) {
	switch cfg.authMode {
	case "disabled":
		return nil, "disabled (development or legacy mode)", nil
	case "fixed":
		validator := auth.NewFixedAPIKeyValidator(cfg.fixedAPIKey, cfg.authDebug)
		return validator.Validate, "fixed fixture token", nil
	case "supabase":
		validator := auth.NewSupabaseValidator(cfg.supabaseURL, cfg.supabaseServiceKey, cfg.authDebug)
		return validator.Validate, "supabase", nil
	default:
		return nil, "", fmt.Errorf("unsupported auth mode %q", cfg.authMode)
	}
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
