package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// migrationAdvisoryLockKey serializes migration passes across processes.
// Concurrent passes (parallel test binaries, several server replicas booting
// at once) interleave per-statement ACCESS EXCLUSIVE locks with each other
// and with concurrent DML, and deadlock; one database-scoped advisory lock
// makes every pass atomic with respect to all others.
const migrationAdvisoryLockKey = 0x746F6F6C // "tool"

func (s *Store) migrate(ctx context.Context) error {
	if s == nil {
		return fmt.Errorf("store is nil")
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT NOT NULL,
            namespace TEXT,
            created_at TIMESTAMPTZ NOT NULL,
            created_by TEXT NOT NULL,
            api_key TEXT
        )`,
		`CREATE TABLE IF NOT EXISTS api_keys (
            id TEXT PRIMARY KEY,
            session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
            name TEXT NOT NULL,
            key TEXT,
		    key_hash TEXT,
		    key_preview TEXT,
		    capabilities JSONB NOT NULL DEFAULT '["read","execute","admin"]'::jsonb,
            created_at TIMESTAMPTZ NOT NULL,
            created_by TEXT NOT NULL,
            revoked_at TIMESTAMPTZ
        )`,
		`CREATE TABLE IF NOT EXISTS machines (
            id TEXT PRIMARY KEY,
            session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
            sdk_version TEXT,
            sdk_language TEXT,
            ip TEXT,
            created_at TIMESTAMPTZ NOT NULL,
            last_ping_at TIMESTAMPTZ NOT NULL,
            draining BOOLEAN NOT NULL DEFAULT FALSE
        )`,
		`CREATE TABLE IF NOT EXISTS tools (
            id TEXT PRIMARY KEY,
            session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
            machine_id TEXT,
            name TEXT NOT NULL,
            description TEXT,
            schema TEXT NOT NULL,
            config JSONB NOT NULL DEFAULT '{}'::jsonb,
            tags JSONB NOT NULL DEFAULT '[]'::jsonb,
            created_at TIMESTAMPTZ NOT NULL,
            last_ping_at TIMESTAMPTZ NOT NULL,
            UNIQUE (session_id, name)
        )`,
		`CREATE TABLE IF NOT EXISTS requests (
            id TEXT PRIMARY KEY,
            session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
            tool_name TEXT NOT NULL,
            status TEXT NOT NULL,
            input TEXT NOT NULL,
            result JSONB,
            result_type TEXT,
            error TEXT,
            executing_machine_id TEXT,
            meta JSONB NOT NULL DEFAULT '{}'::jsonb,
            stream_results JSONB NOT NULL DEFAULT '[]'::jsonb,
		    stream_start_seq INTEGER NOT NULL DEFAULT 1,
		    next_stream_seq INTEGER NOT NULL DEFAULT 1,
            attempts INTEGER NOT NULL DEFAULT 0,
            max_attempts INTEGER NOT NULL DEFAULT 3,
            backoff_seconds INTEGER NOT NULL DEFAULT 5,
            visible_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            next_attempt_at TIMESTAMPTZ,
            leased_by TEXT,
            leased_at TIMESTAMPTZ,
            lease_epoch BIGINT NOT NULL DEFAULT 0,
            timeout_seconds INTEGER NOT NULL DEFAULT 45,
            dead_letter BOOLEAN NOT NULL DEFAULT FALSE,
            last_error TEXT,
            created_at TIMESTAMPTZ NOT NULL,
            updated_at TIMESTAMPTZ NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS tasks (
            id TEXT PRIMARY KEY,
            session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
            tool_name TEXT NOT NULL,
            status TEXT NOT NULL,
            input TEXT NOT NULL,
            result TEXT,
            result_type TEXT,
            error TEXT,
            attempts INTEGER NOT NULL DEFAULT 0,
            max_attempts INTEGER NOT NULL DEFAULT 3,
            backoff_seconds INTEGER NOT NULL DEFAULT 5,
            next_attempt_at TIMESTAMPTZ,
            timeout_seconds INTEGER NOT NULL DEFAULT 60,
            dead_letter BOOLEAN NOT NULL DEFAULT FALSE,
            last_error TEXT,
            current_request_id TEXT,
            created_at TIMESTAMPTZ NOT NULL,
            updated_at TIMESTAMPTZ NOT NULL,
            completed_at TIMESTAMPTZ
        )`,
		`CREATE TABLE IF NOT EXISTS user_sessions (
            user_id TEXT NOT NULL,
            session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            PRIMARY KEY (user_id, session_id)
        )`,
		`CREATE INDEX IF NOT EXISTS idx_machines_last_ping ON machines(last_ping_at)`,
		`CREATE INDEX IF NOT EXISTS idx_machines_session_last_ping ON machines(session_id, last_ping_at)`,
		`CREATE INDEX IF NOT EXISTS idx_tools_session_name ON tools(session_id, name)`,
		`CREATE INDEX IF NOT EXISTS idx_tools_machine ON tools(machine_id)`,
		`CREATE INDEX IF NOT EXISTS idx_requests_session_status ON requests(session_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_requests_ready ON requests(session_id, status, tool_name, visible_at)`,
		`CREATE INDEX IF NOT EXISTS idx_requests_dead_letter ON requests(session_id, dead_letter)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_session_status ON tasks(session_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_retry ON tasks(session_id, dead_letter, next_attempt_at)`,
	}

	alterStatements := []string{
		`ALTER TABLE requests ADD COLUMN IF NOT EXISTS attempts INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE requests ADD COLUMN IF NOT EXISTS max_attempts INTEGER NOT NULL DEFAULT 3`,
		`ALTER TABLE requests ADD COLUMN IF NOT EXISTS backoff_seconds INTEGER NOT NULL DEFAULT 5`,
		`ALTER TABLE requests ADD COLUMN IF NOT EXISTS visible_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`ALTER TABLE requests ADD COLUMN IF NOT EXISTS next_attempt_at TIMESTAMPTZ`,
		`ALTER TABLE requests ADD COLUMN IF NOT EXISTS leased_by TEXT`,
		`ALTER TABLE requests ADD COLUMN IF NOT EXISTS leased_at TIMESTAMPTZ`,
		`ALTER TABLE requests ADD COLUMN IF NOT EXISTS lease_epoch BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE requests ADD COLUMN IF NOT EXISTS timeout_seconds INTEGER NOT NULL DEFAULT 45`,
		`ALTER TABLE requests ADD COLUMN IF NOT EXISTS dead_letter BOOLEAN NOT NULL DEFAULT FALSE`,
		// Reaper and retention indexes. These live in the alter pass because
		// they reference columns older schemas only gain above: running them
		// before the ALTERs would fail direct upgrades and roll back the
		// whole migration.
		`CREATE INDEX IF NOT EXISTS idx_requests_reaper_lease ON requests (visible_at) WHERE dead_letter = false AND leased_at IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_requests_reaper_timeout ON requests (leased_at) WHERE dead_letter = false AND leased_at IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_requests_retention ON requests (updated_at) WHERE status IN ('done', 'failed')`,
		`ALTER TABLE requests ADD COLUMN IF NOT EXISTS last_error TEXT`,
		`ALTER TABLE requests ADD COLUMN IF NOT EXISTS stream_start_seq INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE requests ADD COLUMN IF NOT EXISTS next_stream_seq INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS attempts INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS max_attempts INTEGER NOT NULL DEFAULT 3`,
		`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS backoff_seconds INTEGER NOT NULL DEFAULT 5`,
		`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS next_attempt_at TIMESTAMPTZ`,
		`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS timeout_seconds INTEGER NOT NULL DEFAULT 60`,
		`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS dead_letter BOOLEAN NOT NULL DEFAULT FALSE`,
		`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS last_error TEXT`,
		`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS current_request_id TEXT`,
		`ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS key_hash TEXT`,
		`ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS key_preview TEXT`,
		`ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS capabilities JSONB NOT NULL DEFAULT '["read","execute","admin"]'::jsonb`,
		`ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS allowed_tools JSONB`,
		`ALTER TABLE api_keys ALTER COLUMN key DROP NOT NULL`,
		`ALTER TABLE requests ADD COLUMN IF NOT EXISTS idempotency_key TEXT`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_requests_idempotency ON requests(session_id, idempotency_key) WHERE idempotency_key <> ''`,
		`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS idempotency_key TEXT`,
		`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS adopted_by TEXT`,
		`CREATE TABLE IF NOT EXISTS audit_events (
            id BIGSERIAL PRIMARY KEY,
            created_at TIMESTAMPTZ NOT NULL,
            event TEXT NOT NULL,
            session_id TEXT,
            machine_id TEXT,
            request_id TEXT,
            task_id TEXT,
            details JSONB
        )`,
		`CREATE TABLE IF NOT EXISTS request_chunks (
            request_id TEXT NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
            seq INTEGER NOT NULL,
            chunk TEXT NOT NULL,
            created_at TIMESTAMPTZ NOT NULL,
            PRIMARY KEY (request_id, seq)
        )`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_tasks_idempotency ON tasks(session_id, idempotency_key) WHERE idempotency_key <> ''`,
		// The audit-events index sits after every table creation: this table
		// itself is created above in this pass, and the index needs it.
		`CREATE INDEX IF NOT EXISTS idx_audit_events_created_at ON audit_events (created_at)`,
		`ALTER TABLE machines ADD COLUMN IF NOT EXISTS draining BOOLEAN NOT NULL DEFAULT FALSE`,
		`ALTER TABLE machines ADD COLUMN IF NOT EXISTS token_hash TEXT`,
		// Audit attribution: the acting API key behind each audited event.
		// Existing rows predate attribution and stay NULL — the trail is
		// append-only, so no backfill is possible or needed.
		`ALTER TABLE audit_events ADD COLUMN IF NOT EXISTS actor_key_id TEXT`,
		// The legacy session-lock column is dead: written (always empty) by
		// every session upsert, read by nothing. Pre-1.0 with no tagged
		// releases, the wire-only compatibility policy does not pin internal
		// storage shape.
		`ALTER TABLE sessions DROP COLUMN IF EXISTS api_key`,
	}

	conn, err := s.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("migration advisory lock conn: %w", err)
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, migrationAdvisoryLockKey); err != nil {
		return fmt.Errorf("migration advisory lock: %w", err)
	}
	defer func() {
		// Session-scoped lock: release on the same connection, even when the
		// migration tx below fails.
		_, _ = conn.ExecContext(context.WithoutCancel(ctx), `SELECT pg_advisory_unlock($1)`, migrationAdvisoryLockKey)
	}()

	// The advisory lock serializes this pass against other migrations, but a
	// migration tx still interleaves table locks with concurrent DML (provider
	// writes landing mid-boot) and can be chosen as the deadlock victim. Retry
	// those; every other error fails immediately.
	for attempt := 0; attempt < 3; attempt++ {
		err = s.runMigrationStatements(ctx, conn, stmts, alterStatements)
		if err == nil {
			return nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "40P01" {
			time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
			continue
		}
		return err
	}
	return fmt.Errorf("migration tx: %w (retried after repeated deadlocks)", err)
}

// runMigrationStatements executes the migration transaction on the SAME
// connection that holds the advisory lock: pg_advisory_lock is session-scoped,
// so taking it on conn while the tx ran on a pool connection would leave the
// pass unprotected (and could starve a one-connection pool waiting on itself).
func (s *Store) runMigrationStatements(ctx context.Context, conn *sql.Conn, stmts, alterStatements []string) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for _, stmt := range stmts {
		if _, err = tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("execute migration: %w", err)
		}
	}

	for _, stmt := range alterStatements {
		if _, err = tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("execute alter migration: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit migration tx: %w", err)
	}
	return nil
}
