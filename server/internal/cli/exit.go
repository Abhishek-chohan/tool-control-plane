// Package cli holds the surfaces every toolplane command shares: the
// exit-code contract over gRPC statuses, global connection/output flags,
// and the table/JSON renderer. Admin and consumer verbs are thin gRPC
// clients; this package is what makes them scriptable without parsing
// error strings.
package cli

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Exit codes are a stable contract for scripts: every toolplane command
// exits with the code matching the failure's gRPC status, so automation
// branches on semantics without parsing messages. Zero is success; 1 is
// the fail-closed bucket for anything without a specific code.
const (
	ExitOK                 = 0
	ExitError              = 1
	ExitInvalidArgument    = 2
	ExitNotFound           = 3
	ExitPermissionDenied   = 4
	ExitFailedPrecondition = 5
	ExitResourceExhausted  = 6
	ExitUnavailable        = 7
	ExitDeadlineExceeded   = 8
	ExitCancelled          = 9
)

// ExitCodeFor maps an error from a toolplane RPC onto the exit-code
// contract. Non-gRPC errors (connection setup, flag validation) land in
// the generic bucket; connection failures surface as codes.Unavailable
// from the transport and map to the retryable exit code.
func ExitCodeFor(err error) int {
	if err == nil {
		return ExitOK
	}
	switch status.Code(err) {
	case codes.InvalidArgument:
		return ExitInvalidArgument
	case codes.NotFound:
		return ExitNotFound
	case codes.PermissionDenied:
		return ExitPermissionDenied
	case codes.FailedPrecondition:
		return ExitFailedPrecondition
	case codes.ResourceExhausted:
		return ExitResourceExhausted
	case codes.Unavailable:
		return ExitUnavailable
	case codes.DeadlineExceeded:
		return ExitDeadlineExceeded
	case codes.Canceled:
		return ExitCancelled
	default:
		return ExitError
	}
}

// IsUnavailable reports whether err is a gRPC UNAVAILABLE — the
// "nothing listening / backend down" shape used by the CLI's teaching
// error messages.
func IsUnavailable(err error) bool {
	return err != nil && status.Code(err) == codes.Unavailable
}

// ErrNoConfig is returned when a command needs an API key and neither the
// flag nor TOOLPLANE_API_KEY provided one.
var ErrNoConfig = errors.New("no API key: pass --api-key or set TOOLPLANE_API_KEY")
