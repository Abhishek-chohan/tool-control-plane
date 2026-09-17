package service

import (
	"errors"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toolplane/pkg/storage"
)

// TestSerializationConflictMapsToUnavailable: retry exhaustion surfaces as
// UNAVAILABLE — retryable at the RPC layer, which every SDK already does
// for UNAVAILABLE — instead of failing closed to INTERNAL with raw
// SQLSTATE text.
func TestSerializationConflictMapsToUnavailable(t *testing.T) {
	err := fmt.Errorf("claim request: %w: could not serialize access (SQLSTATE 40001)", storage.ErrSerializationConflict)
	mapped := statusFromDomainError("claim request", err)
	if code := status.Code(mapped); code != codes.Unavailable {
		t.Fatalf("serialization exhaustion maps to %v, want Unavailable", code)
	}
	if !errors.Is(err, storage.ErrSerializationConflict) {
		t.Fatal("sentinel must survive wrapping")
	}
}
