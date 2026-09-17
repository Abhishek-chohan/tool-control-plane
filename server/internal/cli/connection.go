package cli

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
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

// TimeoutFlag is a cobra-friendly duration flag for the --timeout global.
type TimeoutFlag time.Duration

func (t *TimeoutFlag) String() string { return time.Duration(*t).String() }
func (t *TimeoutFlag) Set(v string) error {
	d, err := time.ParseDuration(v)
	if err != nil {
		return err
	}
	*t = TimeoutFlag(d)
	return nil
}

// Globals carries the global connection/output flags shared by verbs.
// Each command binds its own copy so defaults resolve per command.
type Globals struct {
	Address string
	APIKey  string
	Timeout TimeoutFlag
	Format  string
}

// Bind registers the global flags on a command.
func (g *Globals) Bind(cmd *cobra.Command) {
	cmd.Flags().StringVar(&g.Address, "address", DefaultAddress, "control plane address")
	cmd.Flags().StringVar(&g.APIKey, "api-key", "", "API key (defaults to $"+APIKeyEnvVar+")")
	cmd.Flags().DurationVar((*time.Duration)(&g.Timeout), "timeout", DefaultTimeout, "per-RPC deadline")
	cmd.Flags().StringVar(&g.Format, "format", string(FormatTable), "output format: table or json")
}

// Connection resolves the flags into a Connection (flag > env > default).
func (g Globals) Connection() Connection {
	return NewConnection(g.Address, g.APIKey, time.Duration(g.Timeout))
}

// FormatParsed validates and returns the selected output format.
func (g Globals) FormatParsed() (Format, error) {
	return ParseFormat(g.Format)
}

// ValidateFormat reports whether the selected --format value is known.
func (g Globals) ValidateFormat() error {
	_, err := ParseFormat(g.Format)
	return err
}

// FormatParsedOrDefault returns the parsed format, defaulting to table
// when unset — call ValidateFormat first to reject bad values loudly.
func (g Globals) FormatParsedOrDefault() Format {
	f, _ := ParseFormat(g.Format)
	return f
}
