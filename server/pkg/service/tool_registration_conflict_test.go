package service

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
	"toolplane/pkg/storage/memory"
)

// runRegisterToolOwnershipContract drives the registration ownership
// contract against one backend: same-machine re-register and stale-owner
// takeover are upserts that preserve the tool ID and return OK; a name
// held by another live machine is a state conflict that funnels to
// FAILED_PRECONDITION — never AlreadyExists, never Internal.
func runRegisterToolOwnershipContract(t *testing.T, store storage.Storer, sessionID string) {
	t.Helper()

	tracer := &recordingTracer{}
	toolService := NewToolService(tracer, store)
	machineService := NewMachinesService(context.Background(), toolService, tracer, store)

	machineA := sessionID + "-machine-a"
	machineB := sessionID + "-machine-b"

	if _, err := machineService.RegisterMachine(sessionID, machineA, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineA, "shared", "shared tool", `{}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine A: %v", err)
	}
	original, err := toolService.RegisterTool(sessionID, machineA, "shared", "shared tool", `{}`, nil, nil)
	if err != nil {
		t.Fatalf("initial registration: %v", err)
	}
	if original.ID == "" {
		t.Fatal("initial registration returned no tool ID")
	}

	// Same-machine re-register: an upsert that keeps the tool ID.
	reregistered, err := toolService.RegisterTool(sessionID, machineA, "shared", "shared tool v2", `{}`, nil, nil)
	if err != nil {
		t.Fatalf("same-machine re-register: %v", err)
	}
	if reregistered.ID != original.ID {
		t.Fatalf("same-machine re-register changed tool ID: %s -> %s", original.ID, reregistered.ID)
	}

	// A second live machine claiming the name: the one real conflict.
	if _, err := machineService.RegisterMachine(sessionID, machineB, "1.0.0", "go", "127.0.0.1", nil, ""); err != nil {
		t.Fatalf("register machine B: %v", err)
	}
	_, err = toolService.RegisterTool(sessionID, machineB, "shared", "rival", `{}`, nil, nil)
	if !errors.Is(err, storage.ErrToolOwnershipConflict) {
		t.Fatalf("fresh-owner conflict: err=%v, want storage.ErrToolOwnershipConflict", err)
	}
	mapped := statusFromDomainError("register tool", err)
	if code := status.Code(mapped); code != codes.FailedPrecondition {
		t.Fatalf("fresh-owner conflict maps to %v, want FailedPrecondition", code)
	}
	if !strings.Contains(mapped.Error(), machineA) {
		t.Fatalf("conflict message must name the owning machine %s: %v", machineA, mapped)
	}

	// Stale-owner takeover: once A's heartbeat is past the TTL, B's
	// registration is an ownership transfer, not a conflict.
	if store != nil {
		stored, err := store.GetMachine(context.Background(), machineA)
		if err != nil || stored == nil {
			t.Fatalf("load machine A: %v (stored=%v)", err, stored)
		}
		stored.LastPingAt = time.Now().Add(-2 * machineHeartbeatTTL)
		if err := store.SaveMachine(context.Background(), stored); err != nil {
			t.Fatalf("age machine A heartbeat: %v", err)
		}
		taken, err := toolService.RegisterTool(sessionID, machineB, "shared", "taken over", `{}`, nil, nil)
		if err != nil {
			t.Fatalf("stale-owner takeover: %v", err)
		}
		if taken.ID != original.ID {
			t.Fatalf("takeover changed tool ID: %s -> %s", original.ID, taken.ID)
		}
		if taken.MachineID != machineB {
			t.Fatalf("takeover left owner as %s, want %s", taken.MachineID, machineB)
		}
	}
}

// TestRegisterToolOwnershipContractCacheOnly runs the ownership contract
// through the no-store in-memory registry.
func TestRegisterToolOwnershipContractCacheOnly(t *testing.T) {
	runRegisterToolOwnershipContract(t, nil, "session-g3-cache")
}

// TestRegisterToolOwnershipContractMemoryStore runs the ownership contract
// through the memory Storer — the production memory storage mode.
func TestRegisterToolOwnershipContractMemoryStore(t *testing.T) {
	runRegisterToolOwnershipContract(t, memory.New(), "session-g3-memstore")
}

// TestRegisterToolOwnershipContractPostgres runs the ownership contract
// against the durable claim transaction.
func TestRegisterToolOwnershipContractPostgres(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("TOOLPLANE_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("TOOLPLANE_DATABASE_URL not set")
	}
	t.Setenv("TOOLPLANE_STORAGE_MODE", "postgres")
	t.Setenv("TOOLPLANE_DATABASE_URL", databaseURL)

	store := openPersistentStoreForTest(t)
	defer func() {
		if err := store.Close(); err != nil {
			t.Fatalf("close store: %v", err)
		}
	}()

	tracer := &recordingTracer{}
	sessionService := NewSessionsService(tracer, store)
	session, err := sessionService.CreateSession("g3-user", "G3 Ownership", "", "", "")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	t.Cleanup(func() {
		_ = sessionService.DeleteSession(session.ID)
	})

	runRegisterToolOwnershipContract(t, store, session.ID)
}

// claimFailingStore delegates every method to a memory store except the
// ownership claim, which fails with the injected error — the shape of a
// Postgres outage or persistence deadline during registration.
type claimFailingStore struct {
	*memory.Store
	claimErr error
}

func (f *claimFailingStore) ClaimToolOwnership(ctx context.Context, tool *model.Tool, staleCutoff time.Time) (*model.Tool, string, error) {
	return nil, "", f.claimErr
}

// TestRegisterToolStoreFailureKeepsRealCode pins the end of the catch-all:
// infrastructure failures during registration keep their real code instead
// of masquerading as AlreadyExists.
func TestRegisterToolStoreFailureKeepsRealCode(t *testing.T) {
	tracer := &recordingTracer{}

	outage := NewToolService(tracer, &claimFailingStore{Store: memory.New(), claimErr: errors.New("connection refused")})
	_, err := outage.RegisterTool("session-g3-outage", "machine-g3", "tool", "d", `{}`, nil, nil)
	if err == nil {
		t.Fatal("registration against a failing store must fail")
	}
	if code := status.Code(statusFromDomainError("register tool", err)); code != codes.Internal {
		t.Fatalf("store outage maps to %v, want Internal (fail-closed)", code)
	}

	deadline := NewToolService(tracer, &claimFailingStore{Store: memory.New(), claimErr: context.DeadlineExceeded})
	_, err = deadline.RegisterTool("session-g3-deadline", "machine-g3", "tool", "d", `{}`, nil, nil)
	if err == nil {
		t.Fatal("registration against a deadline store must fail")
	}
	if code := status.Code(statusFromDomainError("register tool", err)); code != codes.DeadlineExceeded {
		t.Fatalf("persistence deadline maps to %v, want DeadlineExceeded", code)
	}
}
