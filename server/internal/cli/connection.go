package cli

import (
	"context"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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

// Dial opens a gRPC connection to the resolved address. The connection is
// lazy; Call attaches the authenticated metadata per invocation.
func (c Connection) Dial() (*grpc.ClientConn, error) {
	return grpc.NewClient(c.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

// Call derives an invocation context: the caller's timeout plus the API
// key metadata when a key is configured.
func (c Connection) Call(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(parent, c.Timeout)
	if c.APIKey != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "api_key", c.APIKey)
	}
	return ctx, cancel
}
