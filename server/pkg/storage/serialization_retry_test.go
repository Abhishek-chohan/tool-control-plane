package storage

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// syntheticSerializationFailure looks exactly like what Postgres returns on
// a SERIALIZABLE abort, without needing real row contention.
func syntheticSerializationFailure() error {
	return &pgconn.PgError{Code: "40001", Message: "could not serialize access due to read/write dependencies among transactions"}
}

type countingObserver struct {
	retries   int
	exhausted int
}

func (o *countingObserver) SerializationRetry()     { o.retries++ }
func (o *countingObserver) SerializationExhausted() { o.exhausted++ }

// openRetryTestStore opens the store the retry-loop tests drive; the
// transaction body never touches the tx handle, so any database works.
func openRetryTestStore(t *testing.T) *Store {
	t.Helper()
	databaseURL := strings.TrimSpace(os.Getenv("TOOLPLANE_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("TOOLPLANE_DATABASE_URL not set")
	}
	t.Setenv("TOOLPLANE_DATABASE_URL", databaseURL)
	store, err := OpenFromEnv(context.Background(), nil)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

// TestWithSerializableTxExhaustionSurfacesSentinel drives the retry loop
// against a real connection whose transaction body always aborts with
// SQLSTATE 40001: the attempt budget exhausts, the outcome is
// ErrSerializationConflict (carrying the last error), and the observer saw
// every retry plus the exhaustion.
func TestWithSerializableTxExhaustionSurfacesSentinel(t *testing.T) {
	store := openRetryTestStore(t)
	observer := &countingObserver{}
	store.SetSerializationObserver(observer)

	err := store.withSerializableTx(context.Background(), func(tx *sql.Tx) error {
		return syntheticSerializationFailure()
	})
	if err == nil {
		t.Fatal("exhausting loop must fail")
	}
	if !errors.Is(err, ErrSerializationConflict) {
		t.Fatalf("exhaustion error = %v, want ErrSerializationConflict in the chain", err)
	}
	if !strings.Contains(err.Error(), "40001") {
		t.Fatalf("exhaustion error must carry the last SQL error: %v", err)
	}
	if observer.exhausted != 1 {
		t.Fatalf("exhausted events = %d, want 1", observer.exhausted)
	}
	if observer.retries == 0 {
		t.Fatal("retry events = 0, want the intermediate retries counted")
	}
}

// TestWithSerializableTxRetriesThenSucceeds: one serialization abort
// followed by success returns nil and counts exactly one retry.
func TestWithSerializableTxRetriesThenSucceeds(t *testing.T) {
	store := openRetryTestStore(t)
	observer := &countingObserver{}
	store.SetSerializationObserver(observer)

	attempts := 0
	err := store.withSerializableTx(context.Background(), func(tx *sql.Tx) error {
		attempts++
		if attempts == 1 {
			return syntheticSerializationFailure()
		}
		return nil
	})
	if err != nil {
		t.Fatalf("retry-then-succeed: %v", err)
	}
	if observer.retries != 1 || observer.exhausted != 0 {
		t.Fatalf("observer = %+v, want one retry and no exhaustion", observer)
	}
}

// TestSerializationBackoffJitters pins the non-constant delay: repeated
// draws over the same base must produce more than one distinct value, so
// contending replicas de-synchronize instead of re-aborting in lockstep.
func TestSerializationBackoffJitters(t *testing.T) {
	seen := make(map[time.Duration]bool)
	for i := 0; i < 64; i++ {
		seen[serializationBackoff(2*time.Millisecond)] = true
		if len(seen) > 1 {
			return
		}
	}
	t.Fatalf("backoff draws were constant across 64 samples: %v", seen)
}

// TestSerializationBackoffBounds: full jitter draws stay within [0, base)
// and a non-positive base means no sleep.
func TestSerializationBackoffBounds(t *testing.T) {
	if got := serializationBackoff(0); got != 0 {
		t.Fatalf("zero base must mean no sleep, got %v", got)
	}
	for i := 0; i < 64; i++ {
		if got := serializationBackoff(8 * time.Millisecond); got < 0 || got >= 8*time.Millisecond {
			t.Fatalf("draw %v out of [0, 8ms)", got)
		}
	}
}
