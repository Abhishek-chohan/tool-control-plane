package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
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

// ErrLeaseConflict reports that a fenced write (update/submit/append/renew)
// was rejected because the caller does not hold the request's current lease
// grant: the machine identity or lease epoch did not match, the request was
// reclaimed, or it already reached a terminal state.
var ErrLeaseConflict = errors.New("storage: lease conflict: caller does not hold the current lease grant")

// ErrRequestTerminal reports that a fenced write targeted a request that is
// already in a terminal state (done/failure).
var ErrRequestTerminal = errors.New("storage: request is already in a terminal state")

// ErrMachineNotToolOwner reports that a claim was rejected because the
// claiming machine never registered the request's tool: a provider may only
// lease and answer requests for tools it provides, and the ownership filter
// is derived from the tools registry, never from caller-supplied names alone.
var ErrMachineNotToolOwner = errors.New("storage: machine does not provide the requested tool")

// ErrNotFound reports that the targeted row does not exist. Fenced
// primitives return it (wrapped with detail) so callers can distinguish a
// missing entity from a lease rejection without parsing messages.
var ErrNotFound = errors.New("storage: not found")

// ErrRequestExists reports that a non-upsert insert targeted a request id
// that is already persisted. Request creation is insert-only, so this error
// surfaces instead of silently overwriting the existing row.
var ErrRequestExists = errors.New("storage: request already exists")

// ErrTooManyPendingRequests reports that a request insert was rejected
// because the session already holds the maximum outstanding pending
// backlog. The store owns this verdict: per-replica caches count only
// their own creates, so the durable count is the only enforcement point
// that holds across replicas.
var ErrTooManyPendingRequests = errors.New("storage: too many pending requests in session")

// MaxPendingRequestsPerSession caps the outstanding pending backlog a
// single session may accumulate. Providers are expected to drain it; a
// session whose claims stall (no provider, capacity) hits this ceiling
// instead of growing the queue without bound. Enforced inside the insert
// transaction so the verdict is authoritative across replicas.
const MaxPendingRequestsPerSession = 512

// ErrChunkWindowGap reports that the retained chunk table does not cover the
// row's advertised window contiguously — the window bookkeeping raced the
// trim (cross-replica append/trim interleaving). Stream readers map this to
// OUT_OF_RANGE: the caller asked for history the window no longer holds.
var ErrChunkWindowGap = errors.New("storage: chunk window gap")

// Store provides persistence for core server models.
type Store struct {
	db     *sql.DB
	logger *log.Logger

	// serializationObserver, when attached, counts retry-loop lifecycle
	// events for operational metrics.
	serializationObserver SerializationObserver
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
	// GetRequestByIdempotencyKey fetches the request a session created under
	// the given dedup key (nil when absent); CreateRequest uses it to make
	// retries return the original request instead of duplicate work.
	GetRequestByIdempotencyKey(ctx context.Context, sessionID, idempotencyKey string) (*model.Request, error)
	// GetRequestChunksByRequest returns the retained chunk window for a
	// request from the append-only chunk table: chunks with startSeq <= seq <
	// nextSeq. Store-backed window reads are served from here; the request
	// row no longer carries chunk payloads.
	GetRequestChunksByRequest(ctx context.Context, requestID string, startSeq, nextSeq int32) (model.RequestChunkWindow, error)
	// ListRequestsBySession returns every request in a session, oldest first.
	// It is the read-through for ListRequests when other replicas created
	// requests this instance has not seen.
	ListRequestsBySession(ctx context.Context, sessionID string) ([]*model.Request, error)

	// Fenced request writes. Every method loads the authoritative row, verifies
	// that (machineID, leaseEpoch) identify the request's current lease grant,
	// applies its mutation, and persists atomically. Mismatched or stale lease
	// grants fail with ErrLeaseConflict, so a reclaimed executor (or any
	// non-holder) can no longer write results or chunks.
	RenewRequestLease(ctx context.Context, sessionID, requestID, machineID string, leaseEpoch int64, leaseDuration time.Duration) (*model.Request, error)
	SubmitRequestResultFenced(ctx context.Context, sessionID, requestID, machineID string, leaseEpoch int64, result interface{}, resultType model.ResultType, meta map[string]string) (*model.Request, error)
	AppendRequestChunksFenced(ctx context.Context, sessionID, requestID, machineID string, leaseEpoch int64, chunks []string) (*model.Request, error)
	UpdateRequestFenced(ctx context.Context, sessionID, requestID, machineID string, leaseEpoch int64, status model.RequestStatus, result interface{}, resultType model.ResultType, leaseDuration time.Duration) (*model.Request, error)
	RequeueRequestFenced(ctx context.Context, sessionID, requestID, machineID string, leaseEpoch int64, reason string, backoff time.Duration) (*model.Request, error)
	// CancelRequestFenced cancels a request under the row lock. It returns
	// (request, false, nil) when the request is already terminal — carrying
	// the authoritative terminal row so the caller reports reality instead of
	// overwriting it — and otherwise applies the cancellation (rejection
	// result, dead_letter, lease release) and persists it in the same
	// transaction. A cancel racing a result submission therefore serializes:
	// exactly one of the two writes the terminal state.
	CancelRequestFenced(ctx context.Context, sessionID, requestID, message string) (*model.Request, bool, error)
	// InsertRequest writes a brand-new request row without an upsert clause:
	// a colliding insert (same id, or the idempotency-key partial unique
	// index) fails instead of overwriting the persisted row. It also enforces
	// the per-session pending backlog cap (MaxPendingRequestsPerSession)
	// atomically with the insert, rejecting with ErrTooManyPendingRequests.
	InsertRequest(ctx context.Context, req *model.Request) error
	// CountPendingRequests returns the total number of pending,
	// non-dead-lettered requests across all sessions. It is the
	// store-sourced reading behind the queue-depth gauge.
	CountPendingRequests(ctx context.Context) (int, error)

	// Machines
	AllMachines(ctx context.Context) ([]*model.Machine, error)
	// GetMachine fetches a single machine by ID regardless of session;
	// returns nil when absent. Used by the machine-token gate's cross-replica
	// read-through.
	GetMachine(ctx context.Context, machineID string) (*model.Machine, error)
	// ListMachinesBySession returns the machines registered in a session. It
	// backs serving-time machine reads so answers are not partition-dependent.
	ListMachinesBySession(ctx context.Context, sessionID string) ([]*model.Machine, error)
	SaveMachine(ctx context.Context, machine *model.Machine) error
	// TouchMachineLastPing advances only the heartbeat timestamp. Heartbeats
	// arrive continuously while a machine may be mid-drain, so this write is
	// deliberately column-scoped: drain state is owned exclusively by
	// SetMachineDraining / ClearMachineDraining (and registration).
	TouchMachineLastPing(ctx context.Context, sessionID, machineID string, at time.Time) error
	// BindMachineToken persists a machine's credential hash only when none is
	// bound yet, and reports whether this call won the bind. Two replicas
	// racing to bind the first credential cannot both win; the loser
	// re-validates the presented credential against the winning hash.
	BindMachineToken(ctx context.Context, machineID, tokenHash string) (bool, error)
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
	// GetTaskByID fetches a task row by ID (nil when absent). Serving-time
	// task reads use it so a task created on another replica resolves, and
	// CancelTask falls back to its durable current_request_id.
	GetTaskByID(ctx context.Context, taskID string) (*model.Task, error)
	// ListTasksBySession returns the session's tasks.
	ListTasksBySession(ctx context.Context, sessionID string) ([]*model.Task, error)
	// Retention. DeleteTerminalRequestsBefore removes done/failed requests
	// whose last transition precedes the cutoff (non-terminal rows are never
	// eligible); DeleteAuditEventsBefore removes audit events created before
	// the cutoff. Both return the number of rows removed.
	DeleteTerminalRequestsBefore(ctx context.Context, cutoff time.Time) (int64, error)
	DeleteAuditEventsBefore(ctx context.Context, cutoff time.Time) (int64, error)
	// GetTaskByIdempotencyKey fetches the task a session created under the
	// given dedup key (nil when absent); CreateTask uses it to make retries
	// return the original task without re-executing it.
	GetTaskByIdempotencyKey(ctx context.Context, sessionID, idempotencyKey string) (*model.Task, error)
	// FindNonTerminalTasks returns tasks that are not in a terminal state
	// (done/failed/cancelled). Used on startup to re-adopt in-flight work.
	FindNonTerminalTasks(ctx context.Context) ([]*model.Task, error)
	// FindAdoptableTasks returns up to limit non-terminal tasks whose next
	// attempt is due (next_attempt_at null or past) and whose adoption lease
	// is free: unowned, or owned but untouched for leaseTTL. The adoption
	// sweep uses it so each tick touches only claimable work, not every
	// non-terminal row.
	FindAdoptableTasks(ctx context.Context, now time.Time, leaseTTL time.Duration, limit int) ([]*model.Task, error)
	// ClaimTaskForAdoption atomically acquires execution ownership of a task:
	// it succeeds when the task is unowned or the previous owner's lease
	// (leaseTTL since its last touch) has expired, and records instanceID as the
	// owner. Only the first caller wins; it is the fencing gate for task
	// execution across replicas.
	ClaimTaskForAdoption(ctx context.Context, taskID, instanceID string, leaseTTL time.Duration) (*model.Task, bool, error)
	// RenewTaskAdoption refreshes the owning instance's lease. It returns
	// false when ownership was lost (expired and re-claimed elsewhere), at
	// which point the caller must stop executing the task.
	RenewTaskAdoption(ctx context.Context, taskID, instanceID string) (bool, error)
	// ReleaseTaskAdoption clears ownership when the executing instance
	// reaches a terminal state or schedules a retry for any instance to pick
	// up. Only the recorded owner may release.
	ReleaseTaskAdoption(ctx context.Context, taskID, instanceID string) error

	// Sessions
	AllSessions(ctx context.Context) ([]*model.Session, error)
	GetSession(ctx context.Context, sessionID string) (*model.Session, error)
	SaveSession(ctx context.Context, session *model.Session) error
	// InsertSessionIfAbsent inserts only when the ID is free (cross-replica
	// CreateSession dedup); returns inserted=false when the row exists.
	InsertSessionIfAbsent(ctx context.Context, session *model.Session) (bool, error)
	DeleteSession(ctx context.Context, sessionID string) error
	AllUserSessions(ctx context.Context) (map[string][]string, error)
	AddUserSession(ctx context.Context, userID, sessionID string) error
	RemoveUserSession(ctx context.Context, userID, sessionID string) error
	AllApiKeys(ctx context.Context) ([]*model.ApiKey, error)
	SaveApiKey(ctx context.Context, key *model.ApiKey) error
	// GetAPIKeyByHash fetches one key by its SHA-256 hash (nil when absent);
	// the auth revalidation path uses it to notice revocations and deletions
	// made on other replicas.
	GetAPIKeyByHash(ctx context.Context, keyHash string) (*model.ApiKey, error)

	// Audit events: durable rows for security- and lifecycle-relevant
	// transitions, written best-effort (failures are logged, never surfaced
	// to the operation that produced the event).
	RecordAuditEvent(ctx context.Context, event *model.AuditEvent) error
	// ListAuditEvents returns the durable audit trail newest-first
	// (created_at, id) with the filter applied, plus the total matching
	// count for pagination. Audit is append-only and pruned on the
	// retention schedule; the listing reflects whatever the window holds.
	ListAuditEvents(ctx context.Context, filter model.AuditEventFilter) ([]*model.AuditEvent, int, error)

	// Tools
	AllTools(ctx context.Context) ([]*model.Tool, error)
	// GetToolByName fetches a tool row by session and name (nil when
	// absent). It backs serving-time tool reads so CreateRequest on a
	// replica that never saw the registration still resolves.
	GetToolByName(ctx context.Context, sessionID, name string) (*model.Tool, error)
	// ListToolsBySession returns the session's tools. It backs serving-time
	// listing so answers are not partition-dependent.
	ListToolsBySession(ctx context.Context, sessionID string) ([]*model.Tool, error)
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

// ErrSerializationConflict reports that a SERIALIZABLE transaction was
// retried to exhaustion on serialization failures (SQLSTATE 40001/40P01):
// the datastore kept aborting the transaction under contention. The work
// was never applied. Callers surface it as UNAVAILABLE — retryable at the
// RPC layer — instead of failing closed to INTERNAL with raw SQLSTATE text.
var ErrSerializationConflict = errors.New("storage: serialization conflict retries exhausted")

// SerializationObserver receives serialization-retry lifecycle events for
// operational metrics. Implementations must be safe for concurrent use.
type SerializationObserver interface {
	SerializationRetry()
	SerializationExhausted()
}

// SetSerializationObserver attaches retry telemetry to the store. The
// observer is optional; without one the retry loop is unchanged.
func (s *Store) SetSerializationObserver(observer SerializationObserver) {
	if s == nil {
		return
	}
	s.serializationObserver = observer
}

// serializationBackoff returns the full-jitter sleep for one retry: a
// uniform draw over [0, base). Jitter de-synchronizes replicas retrying
// against each other — deterministic backoff lets contending instances
// re-abort in lockstep.
func serializationBackoff(base time.Duration) time.Duration {
	if base <= 0 {
		return 0
	}
	// #nosec G404 -- retry jitter only de-synchronizes contending
	// replicas; predictability is harmless, crypto randomness is not
	// worth the cost on this hot path.
	return rand.N(base)
}

// withSerializableTx runs fn inside a SERIALIZABLE transaction, retrying
// serialization failures (SQLSTATE 40001). Concurrent replicas contending on
// the same rows are routine under serializable isolation: Postgres aborts
// one of the transactions, and the correct response is to restart it, not to
// surface an error to the caller. Retries are jittered and counted; when
// the attempt budget is exhausted on serialization failures the outcome is
// ErrSerializationConflict (wrapping the last error) so callers can map it
// to a retryable status instead of failing closed.
func (s *Store) withSerializableTx(ctx context.Context, fn func(*sql.Tx) error) error {
	if s == nil {
		return errors.New("storage: store is nil")
	}

	const maxAttempts = 4
	backoff := 2 * time.Millisecond
	for attempt := 0; ; attempt++ {
		err := s.attemptSerializableTx(ctx, fn)
		if err == nil {
			return nil
		}
		if !isSerializationFailure(err) {
			return err
		}
		if attempt+1 >= maxAttempts {
			if s.serializationObserver != nil {
				s.serializationObserver.SerializationExhausted()
			}
			s.logger.Printf("serializable transaction abandoned after %d attempts: %v", maxAttempts, err)
			return fmt.Errorf("%w: %v", ErrSerializationConflict, err)
		}
		if s.serializationObserver != nil {
			s.serializationObserver.SerializationRetry()
		}
		select {
		case <-time.After(serializationBackoff(backoff)):
		case <-ctx.Done():
			// Cancellation is the caller's outcome, not a retryable
			// serialization failure; surface it instead of the 40001.
			return ctx.Err()
		}
		backoff *= 2
	}
}

func (s *Store) attemptSerializableTx(ctx context.Context, fn func(*sql.Tx) error) error {
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

// IsUniqueViolation reports whether err carries Postgres SQLSTATE 23505
// (unique_violation). Callers use it to detect insert races that a re-read
// resolves.
func IsUniqueViolation(err error) bool {
	var sqlStater interface{ SQLState() string }
	if errors.As(err, &sqlStater) {
		return sqlStater.SQLState() == "23505"
	}
	return false
}

// isSerializationFailure reports whether err carries Postgres SQLSTATE 40001

// (serialization_failure). pgx errors expose the SQLState method.
func isSerializationFailure(err error) bool {
	// 40001 (serialization_failure) and 40P01 (deadlock_detected) are both
	// "the system aborted your transaction, retry it" outcomes: Postgres
	// picks a victim in a write-write conflict and the caller is expected
	// to run the transaction again.
	var sqlStater interface{ SQLState() string }
	if errors.As(err, &sqlStater) {
		switch sqlStater.SQLState() {
		case "40001", "40P01":
			return true
		}
	}
	return false
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
