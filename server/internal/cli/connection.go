package cli

import (
	"os"
	"strings"
	"time"
)

// Defaults shared by every server-facing verb.
const (
	DefaultAddress = "localhost:9001"
	DefaultTimeout = 30 * time.Second
	// APIKeyEnvVar is the documented primary source for the caller's key;
	// the --api-key flag is the override.
	APIKeyEnvVar = "TOOLPLANE_API_KEY" // #nosec G101 -- env var name, not a credential
)

// Connection is the resolved global connection surface: --address,
// --api-key / TOOLPLANE_API_KEY, --timeout. The environment is the
// documented primary for the key (command-line keys land in shell
// history); the flag is the override.
type Connection struct {
	Address string
	APIKey  string
	Timeout time.Duration
}

// NewConnection resolves the connection surface from flag values, where
// an empty flag value means "fall back to the environment / default".
func NewConnection(address, apiKey string, timeout time.Duration) Connection {
	if strings.TrimSpace(address) == "" {
		address = DefaultAddress
	}
	if strings.TrimSpace(apiKey) == "" {
		apiKey = strings.TrimSpace(os.Getenv(APIKeyEnvVar))
	}
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return Connection{Address: address, APIKey: apiKey, Timeout: timeout}
}
