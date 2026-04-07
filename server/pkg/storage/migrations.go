package storage

import (
	"context"
	"fmt"
)

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
            last_ping_at TIMESTAMPTZ NOT NULL
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
		`ALTER TABLE requests ADD COLUMN IF NOT EXISTS timeout_seconds INTEGER NOT NULL DEFAULT 45`,
		`ALTER TABLE requests ADD COLUMN IF NOT EXISTS dead_letter BOOLEAN NOT NULL DEFAULT FALSE`,
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
		`ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS key_hash TEXT`,
		`ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS key_preview TEXT`,
		`ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS capabilities JSONB NOT NULL DEFAULT '["read","execute","admin"]'::jsonb`,
		`ALTER TABLE api_keys ALTER COLUMN key DROP NOT NULL`,
	}

	tx, err := s.db.BeginTx(ctx, nil)
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
