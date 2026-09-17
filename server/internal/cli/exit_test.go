package cli

import (
	"errors"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestExitCodeFor pins the exit-code contract: every command exits with
// the code matching the failure's gRPC status, so scripts branch on
// semantics without parsing messages.
func TestExitCodeFor(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, ExitOK},
		{"invalid argument", status.Error(codes.InvalidArgument, "bad input"), ExitInvalidArgument},
		{"not found", status.Error(codes.NotFound, "missing"), ExitNotFound},
		{"permission denied", status.Error(codes.PermissionDenied, "denied"), ExitPermissionDenied},
		{"failed precondition", status.Error(codes.FailedPrecondition, "draining"), ExitFailedPrecondition},
		{"resource exhausted", status.Error(codes.ResourceExhausted, "backlog full"), ExitResourceExhausted},
		{"unavailable", status.Error(codes.Unavailable, "backend down"), ExitUnavailable},
		{"deadline exceeded", status.Error(codes.DeadlineExceeded, "too slow"), ExitDeadlineExceeded},
		{"cancelled", status.Error(codes.Canceled, "walked away"), ExitCancelled},
		{"internal fail-closed", status.Error(codes.Internal, "boom"), ExitError},
		{"unknown code", status.Error(codes.DataLoss, "lost"), ExitError},
		{"non-grpc error", errors.New("connection refused"), ExitError},
		{"wrapped grpc error", fmt.Errorf("list sessions: %w", status.Error(codes.NotFound, "nope")), ExitNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ExitCodeFor(tc.err); got != tc.want {
				t.Fatalf("ExitCodeFor = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestParseFormat(t *testing.T) {
	for value, want := range map[string]Format{
		"":       FormatTable,
		"table":  FormatTable,
		"json":   FormatJSON,
		"JSON":   FormatJSON,
		" json ": FormatJSON,
	} {
		if got, err := ParseFormat(value); err != nil || got != want {
			t.Fatalf("ParseFormat(%q) = %v, %v; want %v", value, got, err, want)
		}
	}
	if _, err := ParseFormat("yaml"); err == nil {
		t.Fatal("ParseFormat(yaml) accepted; want rejection")
	}
}

func TestConnectionEnvFallback(t *testing.T) {
	t.Setenv(APIKeyEnvVar, "env-key")
	conn := NewConnection("", "", 0)
	if conn.Address != DefaultAddress || conn.APIKey != "env-key" || conn.Timeout != DefaultTimeout {
		t.Fatalf("connection = %+v, want defaults with env key", conn)
	}
	conn = NewConnection("host:1", "flag-key", 5)
	if conn.APIKey != "flag-key" || conn.Address != "host:1" || conn.Timeout != 5 {
		t.Fatalf("flag values must win over env: %+v", conn)
	}
}
