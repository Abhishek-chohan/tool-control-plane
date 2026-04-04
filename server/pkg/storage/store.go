package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
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

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(60 * time.Minute)

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
