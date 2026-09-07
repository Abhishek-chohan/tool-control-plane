package service

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
)

// Domain error sentinels.
//
// Service methods wrap these with operational detail via
// fmt.Errorf("%w: ...", Err...) and the gRPC layer translates every sentinel
// to its canonical status code in statusFromDomainError — the single place
// that decides how a domain failure surfaces to clients. A new condition
// gets a sentinel (or a wrap of an existing one) here, not another
// handler-local codes.Xxx call.
var (
	// ErrNotFound reports a lookup miss: unknown session, request, machine,
	// tool, API key, or task.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists reports a create that collided with an existing entity.
	ErrAlreadyExists = errors.New("already exists")

	// ErrInvalidArgument reports caller input that failed validation before
	// any state was touched.
	ErrInvalidArgument = errors.New("invalid argument")

	// ErrMachineDraining reports that the target machine is draining and
	// refuses new work.
	ErrMachineDraining = errors.New("machine is draining")

	// ErrRequestNotClaimable reports a claim against a request that is not in
	// a claimable state (already claimed by another machine, running, or
	// terminal).
	ErrRequestNotClaimable = errors.New("request not claimable")

	// ErrRequestNotCancellable reports a cancel against a request that
	// already reached a terminal state.
	ErrRequestNotCancellable = errors.New("request not cancellable")

	// ErrTaskNotCancellable reports a cancel against a task that already
	// reached a terminal state.
	ErrTaskNotCancellable = errors.New("task not cancellable")

	// ErrMachineAtCapacity reports that the machine's concurrent-request
	// limit is reached; the work was requeued with backoff.
	ErrMachineAtCapacity = errors.New("machine at capacity")

	// ErrNoProviderAvailable reports a tool with no registered machine able
	// to run it right now.
	ErrNoProviderAvailable = errors.New("no provider available")

	// ErrRequestTimeoutOutOfRange reports a timeout_seconds above the
	// configured maximum.
	ErrRequestTimeoutOutOfRange = errors.New("request timeout out of range")

	// ErrMachineCredentialRejected reports a failed per-machine credential
	// check (wrong token on re-registration, or an unknown token).
	ErrMachineCredentialRejected = errors.New("machine credential rejected")
)

// statusFromDomainError translates a service-layer error into its gRPC
// status. RPC handlers funnel through here so one domain condition always
// surfaces as the same code regardless of which RPC produced it:
//
//	missing entity                       -> NOT_FOUND
//	create collision                     -> ALREADY_EXISTS
//	bad input                            -> INVALID_ARGUMENT
//	wrong or missing machine credential  -> PERMISSION_DENIED
//	draining, not claimable, terminal,
//	stale lease, no provider             -> FAILED_PRECONDITION
//	machine concurrency limit reached    -> RESOURCE_EXHAUSTED
//	timeout above the maximum,
//	expired stream replay window         -> OUT_OF_RANGE
//	anything unmapped                    -> INTERNAL (fail closed)
func statusFromDomainError(action string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrNotFound),
		errors.Is(err, storage.ErrNotFound):
		return status.Errorf(codes.NotFound, "failed to %s: %v", action, err)
	case errors.Is(err, ErrAlreadyExists):
		return status.Errorf(codes.AlreadyExists, "failed to %s: %v", action, err)
	case errors.Is(err, ErrInvalidArgument),
		errors.Is(err, model.ErrUnsupportedAPIKeyCapability),
		errors.Is(err, model.ErrAPIKeyCapabilitiesRequired):
		return status.Errorf(codes.InvalidArgument, "failed to %s: %v", action, err)
	case errors.Is(err, ErrMachineCredentialRejected):
		return status.Errorf(codes.PermissionDenied, "failed to %s: %v", action, err)
	case errors.Is(err, ErrMachineDraining),
		errors.Is(err, ErrRequestNotClaimable),
		errors.Is(err, ErrRequestNotCancellable),
		errors.Is(err, ErrTaskNotCancellable),
		errors.Is(err, ErrNoProviderAvailable),
		errors.Is(err, storage.ErrLeaseConflict),
		errors.Is(err, storage.ErrRequestTerminal):
		return status.Errorf(codes.FailedPrecondition, "failed to %s: %v", action, err)
	case errors.Is(err, ErrMachineAtCapacity):
		return status.Errorf(codes.ResourceExhausted, "failed to %s: %v", action, err)
	case errors.Is(err, ErrRequestTimeoutOutOfRange):
		return status.Errorf(codes.OutOfRange, "failed to %s: %v", action, err)
	default:
		var expired *RequestStreamExpiredError
		if errors.As(err, &expired) {
			return status.Errorf(codes.OutOfRange, "failed to %s: %v", action, err)
		}
		return status.Errorf(codes.Internal, "failed to %s: %v", action, err)
	}
}

// wrapf attaches a domain sentinel to an operational message:
// wrapf(ErrNotFound, "request %s not found", id) fails errors.Is against
// ErrNotFound while keeping the detail text.
func wrapf(err error, format string, args ...interface{}) error {
	return fmt.Errorf("%w: "+format, append([]interface{}{err}, args...)...)
}
