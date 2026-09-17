package server

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileSettings carries the config-file keys that have no environment
// variable and therefore cannot flow through the env-contract apply step;
// the entry point applies them to its Options itself (flags still win via
// their Changed() check).
type FileSettings struct {
	Port          int
	MetricsListen string
	// Provenance records, per applied key, where the winning value came
	// from ("file" or "env") for the --dry-run echo.
	Provenance map[string]string
}

// configFile mirrors the supported YAML keys. Unknown keys are rejected:
// a renamed key must fail loudly at boot, not silently ignore a value the
// operator believes is live.
type configFile struct {
	Env  string `yaml:"env"`
	Auth struct {
		Mode        string `yaml:"mode"`
		FixedAPIKey string `yaml:"fixed_api_key"`
		Debug       bool   `yaml:"debug"`
	} `yaml:"auth"`
	Storage struct {
		Mode            string `yaml:"mode"`
		DatabaseURL     string `yaml:"database_url"`
		DatabaseURLFile string `yaml:"database_url_file"`
	} `yaml:"storage"`
	Server struct {
		Port          int    `yaml:"port"`
		MetricsListen string `yaml:"metrics_listen"`
		TLSCertFile   string `yaml:"cert_file"`
		TLSKeyFile    string `yaml:"key_file"`
	} `yaml:"server"`
}

// envContract maps a YAML path to the environment variable that carries
// the same setting. File values become environment defaults: applied only
// when the variable is unset, so the existing precedence (flag > env >
// file > default) falls out of the ordinary resolution order with no
// second config system.
var envContract = []struct {
	yamlPath string
	envVar   string
	value    func(c configFile) string
}{
	{"env", "TOOLPLANE_ENV_MODE", func(c configFile) string { return c.Env }},
	{"auth.mode", "TOOLPLANE_AUTH_MODE", func(c configFile) string { return c.Auth.Mode }},
	{"auth.fixed_api_key", "TOOLPLANE_AUTH_FIXED_API_KEY", func(c configFile) string { return c.Auth.FixedAPIKey }},
	{"auth.debug", "TOOLPLANE_AUTH_DEBUG", func(c configFile) string { return fmt.Sprintf("%t", c.Auth.Debug) }},
	{"storage.mode", "TOOLPLANE_STORAGE_MODE", func(c configFile) string { return c.Storage.Mode }},
	{"storage.database_url", "TOOLPLANE_DATABASE_URL", func(c configFile) string { return c.Storage.DatabaseURL }},
	{"server.cert_file", "TOOLPLANE_SERVER_TLS_CERT_FILE", func(c configFile) string { return c.Server.TLSCertFile }},
	{"server.key_file", "TOOLPLANE_SERVER_TLS_KEY_FILE", func(c configFile) string { return c.Server.TLSKeyFile }},
}

// ApplyConfigFile loads a YAML config file and applies it as the base
// configuration layer: every mapped environment variable that is still
// unset receives the file's value, so flags and pre-set environment
// variables keep precedence over the file. It is safe to call before any
// other configuration resolution.
//
// Behavior:
//   - unknown keys are a boot failure naming the key;
//   - ${VAR} references in string values expand from the environment;
//   - storage.database_url_file reads the secret from a file instead of
//     accepting it as a literal;
//   - a group/world-readable config file logs a warning (it can carry
//     secrets).
//
// The returned FileSettings exposes the non-environment keys and the
// provenance of every applied value for dry-run echoes.
func ApplyConfigFile(path string) (*FileSettings, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}
	if info, statErr := os.Stat(path); statErr == nil && info.Mode().Perm()&0o077 != 0 {
		slog.Warn("config file is readable by group or others; it may carry secrets",
			slog.String("file", path), slog.String("mode", info.Mode().Perm().String()))
	}

	var cfg configFile
	decoder := yaml.NewDecoder(strings.NewReader(string(raw)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil && !errorsIsEOF(err) {
		return nil, fmt.Errorf("config file %s: %w", path, err)
	}

	settings := &FileSettings{Provenance: map[string]string{}}
	apply := func(envVar, value, source string) {
		if value == "" {
			return
		}
		if os.Getenv(envVar) != "" {
			settings.Provenance[envVar] = "env"
			return
		}
		if err := os.Setenv(envVar, value); err != nil {
			slog.Error("config file value could not be applied", slog.String("key", envVar), slog.Any("err", err))
			return
		}
		settings.Provenance[envVar] = source
	}

	expand := func(value string) string {
		if value == "" {
			return ""
		}
		return os.ExpandEnv(value)
	}

	for _, mapping := range envContract {
		apply(mapping.envVar, expand(mapping.value(cfg)), "file")
	}

	// Secret indirection: the URL lives in its own file (a mounted secret),
	// never as a literal in the config.
	if cfg.Storage.DatabaseURLFile != "" && os.Getenv("TOOLPLANE_DATABASE_URL") == "" {
		secret, err := readSecretFile(expand(cfg.Storage.DatabaseURLFile))
		if err != nil {
			return nil, err
		}
		apply("TOOLPLANE_DATABASE_URL", secret, "file:"+cfg.Storage.DatabaseURLFile)
	}

	if cfg.Server.Port != 0 {
		settings.Port = cfg.Server.Port
	}
	if cfg.Server.MetricsListen != "" {
		settings.MetricsListen = expand(cfg.Server.MetricsListen)
	}
	return settings, nil
}

func readSecretFile(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read secret file %s: %w", path, err)
	}
	value := strings.TrimSpace(string(raw))
	if value == "" {
		return "", fmt.Errorf("secret file %s is empty", path)
	}
	return value, nil
}

func errorsIsEOF(err error) bool {
	return err == io.EOF
}
