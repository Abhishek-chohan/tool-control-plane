package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"toolplane/pkg/model"
)

const defaultConnectTimeout = 5 * time.Second

var ErrConfigMissing = errors.New("storage: TOOLPLANE_DATABASE_URL not set")
var ErrExplicitInMemoryMode = errors.New("storage: explicit in-memory mode")
var ErrToolOwnershipConflict = errors.New("storage: tool ownership conflict with active machine")

// Store provides persistence for core server models.
type Store struct {
	db     *sql.DB
	logger *log.Logger
}

// Store is the persistence boundary used by the service layer. The concrete
// Postgres *Store implements it, and an in-memory implementation backs the
// development and CI memory storage mode. Going through this interface is what
// makes the service layer storage-mode-agnostic and multi-instance ready: the
// authoritative state lives behind the interface, not in service-layer maps.
//
// Method semantics mirror the existing Postgres behavior. Store-first
// primitives added for multi-instance safety (guarded claim, guarded reclaim,
// capacity count, drain flag, non-terminal task adoption) are declared here so
// both implementations stay contract-equivalent; see Phase 2.
type Storer interface {
	// Requests
	AllRequests(ctx context.Context) ([]*model.Request, error)
	SaveRequest(ctx context.Context, req *model.Request) error
	LeasePendingRequest(ctx context.Context, sessionID, machineID string, toolNames []string, leaseDuration time.Duration) (*model.Request, error)
	FindExpiredRequests(ctx context.Context, limit int) ([]*model.Request, error)
	// ClaimRequest atomically transitions a pending request to claimed for the
	// given machine inside a serializable transaction. It returns the updated
	// request and claimed=true on success, or claimed=false (with a nil error)
	// when the request is no longer claimable (missing, not pending, or already
	// claimed). This is the explicit-claim counterpart to LeasePendingRequest.
	ClaimRequest(ctx context.Context, sessionID, requestID, machineID string, leaseDuration time.Duration) (*model.Request, bool, error)
	// ReclaimExpiredRequest reclaims a single request whose lease has expired,
	// inside a guarded transaction. It either marks the request dead-lettered
	// (when attempts are exhausted) or requeues it to pending with a retry
	// backoff. It returns the updated request and reclaimed=true when the row
	// was claimed by this caller, or reclaimed=false (nil error) when another
	// caller already handled it. now/leaseDuration/maxAttempts/backoff mirror
	// the model.Request constants.
	ReclaimExpiredRequest(ctx context.Context, requestID string, now time.Time, leaseDuration time.Duration, maxAttempts int, backoff time.Duration) (*model.Request, bool, error)
	// MachineInFlightCount returns the number of requests currently claimed or
	// running on the given machine. It is the multi-instance-safe source of
	// truth for per-machine capacity enforcement.
	MachineInFlightCount(ctx context.Context, sessionID, machineID string) (int, error)
	// GetRequest fetches a single request by ID, falling back across sessions.
	// It supports store-backed reads when a request is not present in the local
	// cache (multi-instance visibility).
	GetRequest(ctx context.Context, requestID string) (*model.Request, error)

	// Machines
	AllMachines(ctx context.Context) ([]*model.Machine, error)
	SaveMachine(ctx context.Context, machine *model.Machine) error
	DeleteMachine(ctx context.Context, machineID string) error
	ListStaleMachines(ctx context.Context, cutoff time.Time, limit int) ([]*model.Machine, error)
	ReclaimMachine(ctx context.Context, machineID string, cutoff time.Time) (string, []ToolOwnershipUpdate, bool, error)
	// SetMachineDraining / ClearMachineDraining / IsMachineDraining persist the
	// drain flag so drain state is coherent across instances.
	SetMachineDraining(ctx context.Context, sessionID, machineID string) error
	ClearMachineDraining(ctx context.Context, machineID string) error
	IsMachineDraining(ctx context.Context, machineID string) (bool, error)

	// Tasks
	AllTasks(ctx context.Context) ([]*model.Task, error)
	SaveTask(ctx context.Context, task *model.Task) error
	DeleteTask(ctx context.Context, taskID string) error
	// FindNonTerminalTasks returns tasks that are not in a terminal state
	// (done/failed/cancelled). Used on startup to re-adopt in-flight work.
	FindNonTerminalTasks(ctx context.Context) ([]*model.Task, error)
	// ClaimTaskForAdoption atomically claims a non-terminal task for re-adoption
	// by this instance, preventing duplicate execution across replicas. Only the
	// first caller wins; subsequent callers see the task was recently touched and
	// return claimed=false.
	ClaimTaskForAdoption(ctx context.Context, taskID string, minAge time.Duration) (*model.Task, bool, error)

	// Sessions
	AllSessions(ctx context.Context) ([]*model.Session, error)
	GetSession(ctx context.Context, sessionID string) (*model.Session, error)
	SaveSession(ctx context.Context, session *model.Session) error
	DeleteSession(ctx context.Context, sessionID string) error
	AllUserSessions(ctx context.Context) (map[string][]string, error)
	AddUserSession(ctx context.Context, userID, sessionID string) error
	RemoveUserSession(ctx context.Context, userID, sessionID string) error
	AllApiKeys(ctx context.Context) ([]*model.ApiKey, error)
	SaveApiKey(ctx context.Context, key *model.ApiKey) error

	// Tools
	AllTools(ctx context.Context) ([]*model.Tool, error)
	SaveTool(ctx context.Context, tool *model.Tool) error
	DeleteTool(ctx context.Context, toolID string) error
	DeleteToolsByMachine(ctx context.Context, machineID string) ([]ToolOwnershipUpdate, error)
	ClaimToolOwnership(ctx context.Context, tool *model.Tool, staleCutoff time.Time) (*model.Tool, string, error)
}

// Compile-time check: the Postgres implementation satisfies Storer.
var _ Storer = (*Store)(nil)

// OpenFromEnv initializes a Store using TOOLPLANE_DATABASE_URL. Returns nil when the
// environment variable is not set, allowing the server to operate in legacy
// in-memory mode.
func OpenFromEnv(parentCtx context.Context, logger *log.Logger) (*Store, error) {
	mode := strings.TrimSpace(strings.ToLower(os.Getenv("TOOLPLANE_STORAGE_MODE")))
	dsn := strings.TrimSpace(os.Getenv("TOOLPLANE_DATABASE_URL"))

	if mode == "" {
		if dsn == "" {
			return nil, fmt.Errorf("%w; set TOOLPLANE_STORAGE_MODE=memory or configure TOOLPLANE_DATABASE_URL", ErrConfigMissing)
		}
		mode = "postgres"
	}

	switch mode {
	case "memory", "in-memory", "inmemory":
		return nil, ErrExplicitInMemoryMode
	case "postgres":
		if dsn == "" {
			return nil, fmt.Errorf("storage: TOOLPLANE_STORAGE_MODE=postgres requires TOOLPLANE_DATABASE_URL")
		}
	default:
		return nil, fmt.Errorf("storage: unsupported TOOLPLANE_STORAGE_MODE %q", mode)
	}

	if logger == nil {
		logger = log.Default()
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	db.SetMaxOpenConns(intEnv("TOOLPLANE_DB_MAX_OPEN_CONNS", 25))
	db.SetMaxIdleConns(intEnv("TOOLPLANE_DB_MAX_IDLE_CONNS", 25))
	db.SetConnMaxIdleTime(durationEnv("TOOLPLANE_DB_CONN_MAX_IDLE_TIME", 5*time.Minute))
	db.SetConnMaxLifetime(durationEnv("TOOLPLANE_DB_CONN_MAX_LIFETIME", 60*time.Minute))

	ctx, cancel := context.WithTimeout(parentCtx, defaultConnectTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	store := &Store{db: db, logger: logger}
	if err := store.migrate(parentCtx); err != nil {
		db.Close()
		return nil, err
	}

	logger.Println("postgres storage enabled")
	return store, nil
}

// Close releases underlying database resources.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) withSerializableTx(ctx context.Context, fn func(*sql.Tx) error) error {
	if s == nil {
		return errors.New("storage: store is nil")
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback tx: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// intEnv reads a positive integer from the named environment variable, falling
// back to fallback when the variable is unset, empty, not parseable as an int,
// or non-positive.
func intEnv(name string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

// durationEnv reads a Go duration string (e.g. "5m", "30s") from the named
// environment variable, falling back to fallback when the variable is unset,
// empty, not parseable as a duration, or non-positive.
func durationEnv(name string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}
