package client

import (
	"errors"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFromGRPCPreservesCodeAndContext(t *testing.T) {
	cause := status.Error(codes.NotFound, "request req_1 not found")
	err := FromGRPC("get request", "req_1", cause)

	typed, ok := err.(*Error)
	if !ok {
		t.Fatalf("FromGRPC returned %T, want *Error", err)
	}
	if typed.Code != codes.NotFound {
		t.Fatalf("code = %v, want NotFound", typed.Code)
	}
	if typed.RequestID != "req_1" {
		t.Fatalf("request id = %q, want req_1", typed.RequestID)
	}
	if typed.Retryable {
		t.Fatalf("not found marked retryable")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("errors.Is(err, ErrNotFound) = false")
	}
	// status.Code must work through the wrapper.
	if status.Code(err) != codes.NotFound {
		t.Fatalf("status.Code through wrapper = %v, want NotFound", status.Code(err))
	}
	// The original error stays reachable.
	if !errors.Is(err, cause) {
		t.Fatalf("original error not unwrappable")
	}
}

func TestRetryableCodesAreTransportAndCapacityOnly(t *testing.T) {
	retryable := map[codes.Code]bool{
		codes.Unavailable:       IsRetryableCode(codes.Unavailable),
		codes.ResourceExhausted: IsRetryableCode(codes.ResourceExhausted),
	}
	if !retryable[codes.Unavailable] || !retryable[codes.ResourceExhausted] {
		t.Fatalf("unavailable/resource-exhausted must be retryable: %v", retryable)
	}

	for _, code := range []codes.Code{
		codes.Unauthenticated,
		codes.PermissionDenied,
		codes.InvalidArgument,
		codes.FailedPrecondition,
		codes.NotFound,
		codes.AlreadyExists,
		codes.DeadlineExceeded,
	} {
		err := FromGRPC("unit", "req_1", status.Error(code, "x"))
		if err.(*Error).Retryable {
			t.Fatalf("code %v must not be retryable", code)
		}
	}
}

func TestSentinelMatchingByCode(t *testing.T) {
	cases := []struct {
		code     codes.Code
		sentinel error
	}{
		{codes.NotFound, ErrNotFound},
		{codes.InvalidArgument, ErrInvalidArgument},
		{codes.OutOfRange, ErrInvalidArgument},
		{codes.FailedPrecondition, ErrFailedPrecondition},
		{codes.Aborted, ErrFailedPrecondition},
		{codes.ResourceExhausted, ErrResourceExhausted},
		{codes.Unauthenticated, ErrUnauthenticated},
		{codes.PermissionDenied, ErrPermissionDenied},
		{codes.AlreadyExists, ErrAlreadyExists},
		{codes.DeadlineExceeded, ErrDeadlineExceeded},
		{codes.Unavailable, ErrUnavailable},
	}
	for _, tc := range cases {
		err := FromGRPC("unit", "", status.Error(tc.code, "boom"))
		if !errors.Is(err, tc.sentinel) {
			t.Fatalf("errors.Is(%v, %v) = false", tc.code, tc.sentinel)
		}
	}
}

func TestFromGRPCNilPassesThrough(t *testing.T) {
	if err := FromGRPC("unit", "req_1", nil); err != nil {
		t.Fatalf("nil wrapped to %v", err)
	}
}

func TestFromGRPCNonStatusErrorIsUnknown(t *testing.T) {
	err := FromGRPC("unit", "", fmt.Errorf("plain failure"))
	if status.Code(err) != codes.Unknown {
		t.Fatalf("plain error code = %v, want Unknown", status.Code(err))
	}
}
