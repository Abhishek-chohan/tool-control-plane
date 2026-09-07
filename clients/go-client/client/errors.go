package client

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Error is a typed server error. It preserves the gRPC status of the failed
// call (status.Code and status.FromError work through the wrapper), adds the
// operation and request it concerned, and states whether re-issuing the call
// can plausibly succeed.
type Error struct {
	// Op names the operation, e.g. "execute tool" or "get request".
	Op string
	// RequestID is the request the call operated on, when known.
	RequestID string
	// Code is the gRPC status code of the failure.
	Code codes.Code
	// Message is the human-readable failure detail.
	Message string
	// Retryable reports whether the code is transport-level (Unavailable) or
	// capacity (ResourceExhausted) — the only codes where a retry can help.
	Retryable bool

	cause error
}

func (e *Error) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("%s %s: %s", e.Op, e.RequestID, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Op, e.Message)
}

// Unwrap exposes the original gRPC error.
func (e *Error) Unwrap() error { return e.cause }

// GRPCStatus keeps status.Code / status.FromError working through wraps.
func (e *Error) GRPCStatus() *status.Status {
	return status.New(e.Code, e.Message)
}

// Domain sentinels. Match a wrapped *Error by status code, so callers write
// errors.Is(err, client.ErrNotFound) regardless of which RPC failed.
var (
	ErrNotFound           = errors.New("toolplane: not found")
	ErrInvalidArgument    = errors.New("toolplane: invalid argument")
	ErrFailedPrecondition = errors.New("toolplane: failed precondition")
	ErrResourceExhausted  = errors.New("toolplane: resource exhausted")
	ErrUnauthenticated    = errors.New("toolplane: unauthenticated")
	ErrPermissionDenied   = errors.New("toolplane: permission denied")
	ErrAlreadyExists      = errors.New("toolplane: already exists")
	ErrDeadlineExceeded   = errors.New("toolplane: deadline exceeded")
	ErrUnavailable        = errors.New("toolplane: unavailable")
)

// Is reports whether the error's status code matches a domain sentinel.
func (e *Error) Is(target error) bool {
	switch target {
	case ErrNotFound:
		return e.Code == codes.NotFound
	case ErrInvalidArgument:
		return e.Code == codes.InvalidArgument || e.Code == codes.OutOfRange
	case ErrFailedPrecondition:
		return e.Code == codes.FailedPrecondition || e.Code == codes.Aborted
	case ErrResourceExhausted:
		return e.Code == codes.ResourceExhausted
	case ErrUnauthenticated:
		return e.Code == codes.Unauthenticated
	case ErrPermissionDenied:
		return e.Code == codes.PermissionDenied
	case ErrAlreadyExists:
		return e.Code == codes.AlreadyExists
	case ErrDeadlineExceeded:
		return e.Code == codes.DeadlineExceeded
	case ErrUnavailable:
		return e.Code == codes.Unavailable
	}
	return false
}

// IsRetryableCode reports whether re-issuing a call that failed with this
// code can plausibly succeed. Only connection loss (Unavailable) and
// capacity (ResourceExhausted) qualify; credentials, arguments, and state
// conflicts are deterministic.
func IsRetryableCode(code codes.Code) bool {
	return code == codes.Unavailable || code == codes.ResourceExhausted
}

// FromGRPC wraps a stub-call error with its operation and request context.
// nil passes through; errors without a gRPC status wrap as Code=Unknown.
func FromGRPC(op, requestID string, err error) error {
	if err == nil {
		return nil
	}
	st, _ := status.FromError(err)
	return &Error{
		Op:        op,
		RequestID: requestID,
		Code:      st.Code(),
		Message:   st.Message(),
		Retryable: IsRetryableCode(st.Code()),
		cause:     err,
	}
}
