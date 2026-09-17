package cli

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TeachError upgrades a failure into next-step guidance: every common
// failure names the command that fixes it, so the CLI teaches the
// three-process model implicitly instead of failing with a bare code.
func TeachError(err error, conn Connection) string {
	if err == nil {
		return ""
	}
	switch {
	case IsUnavailable(err):
		return fmt.Sprintf("%v\nnothing is answering on %s — start a server with:\n  toolplane serve", err, conn.Address)
	case errors.Is(err, context.DeadlineExceeded),
		status.Code(err) == codes.DeadlineExceeded:
		// A lapsed deadline while waiting on the server is the same story:
		// nothing answered in time. Point at the serve command. gRPC wraps
		// probe timeouts as status errors, so match the code, not just the
		// sentinel.
		return fmt.Sprintf("%v\nnothing answered on %s in time — start a server with:\n  toolplane serve", err, conn.Address)
	}
	return err.Error()
}
