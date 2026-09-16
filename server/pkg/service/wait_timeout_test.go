package service

import (
	"testing"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestValidateWaitTimeoutSeconds pins the wait ceiling on the synchronous
// execution entrypoints: caller-supplied durations cap at maxWaitTimeout
// (the timeout_seconds ceiling), rejected above it with the same reason
// code — OUT_OF_RANGE / TIMEOUT_ABOVE_MAX — before anything is created.
func TestValidateWaitTimeoutSeconds(t *testing.T) {
	cases := []struct {
		name    string
		seconds int32
		wantErr bool
	}{
		{"fire and forget", 0, false},
		{"negative treated as fire and forget", -1, false},
		{"short wait", 30, false},
		{"at max accepted", int32(maxWaitTimeout.Seconds()), false},
		{"one past max rejected", int32(maxWaitTimeout.Seconds()) + 1, true},
		{"a day rejected", 86400, true},
		{"int32 max rejected", 2147483647, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateWaitTimeoutSeconds(tc.seconds)
			if tc.wantErr && err == nil {
				t.Fatalf("wait %ds accepted, want rejection", tc.seconds)
			}
			if !tc.wantErr {
				if err != nil {
					t.Fatalf("wait %d rejected: %v", tc.seconds, err)
				}
				return
			}
			mapped := statusFromDomainError("execute tool", err)
			if code := status.Code(mapped); code != codes.OutOfRange {
				t.Fatalf("over-max wait maps to %v, want OutOfRange", code)
			}
			st, ok := status.FromError(mapped)
			if !ok {
				t.Fatal("expected a gRPC status")
			}
			reason := ""
			for _, detail := range st.Details() {
				if info, ok := detail.(*errdetails.ErrorInfo); ok {
					reason = info.Reason
				}
			}
			if reason != ReasonTimeoutAboveMax {
				t.Fatalf("over-max wait reason = %q, want %q", reason, ReasonTimeoutAboveMax)
			}
		})
	}
}
