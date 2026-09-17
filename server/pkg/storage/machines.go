package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"toolplane/pkg/model"
)

var errMachineFresh = errors.New("storage: machine heartbeat still valid")

func (s *Store) AllMachines(ctx context.Context) ([]*model.Machine, error) {
	if s == nil {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, session_id, sdk_version, sdk_language, ip, created_at, last_ping_at, token_hash, draining FROM machines`)
	if err != nil {
		return nil, fmt.Errorf("query machines: %w", err)
	}
	defer rows.Close()

	var machines []*model.Machine
	for rows.Next() {
		m := &model.Machine{}
		var (
			sdkVersion, sdkLanguage, ip, tokenHash sql.NullString
			draining                               sql.NullBool
		)
		if err := rows.Scan(&m.ID, &m.SessionID, &sdkVersion, &sdkLanguage, &ip, &m.CreatedAt, &m.LastPingAt, &tokenHash, &draining); err != nil {
			return nil, fmt.Errorf("scan machine: %w", err)
		}
		m.Draining = draining.Valid && draining.Bool
		if sdkVersion.Valid {
			m.SDKVersion = sdkVersion.String
		}
		if sdkLanguage.Valid {
			m.SDKLanguage = sdkLanguage.String
		}
		if ip.Valid {
			m.IP = ip.String
		}
		if tokenHash.Valid {
			m.TokenHash = tokenHash.String
		}
		machines = append(machines, m)
	}
	return machines, rows.Err()
}

// ListMachinesBySession returns the machines registered in a session, so
// serving-time listings are not partition-dependent.
func (s *Store) ListMachinesBySession(ctx context.Context, sessionID string) ([]*model.Machine, error) {
	if s == nil {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, session_id, sdk_version, sdk_language, ip, created_at, last_ping_at, token_hash, draining FROM machines WHERE session_id=$1`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list machines by session: %w", err)
	}
	defer rows.Close()

	var machines []*model.Machine
	for rows.Next() {
		m := &model.Machine{}
		var (
			sdkVersion, sdkLanguage, ip, tokenHash sql.NullString
			draining                               sql.NullBool
		)
		if err := rows.Scan(&m.ID, &m.SessionID, &sdkVersion, &sdkLanguage, &ip, &m.CreatedAt, &m.LastPingAt, &tokenHash, &draining); err != nil {
			return nil, fmt.Errorf("scan machine: %w", err)
		}
		m.Draining = draining.Valid && draining.Bool
		if sdkVersion.Valid {
			m.SDKVersion = sdkVersion.String
		}
		if sdkLanguage.Valid {
			m.SDKLanguage = sdkLanguage.String
		}
		if ip.Valid {
			m.IP = ip.String
		}
		if tokenHash.Valid {
			m.TokenHash = tokenHash.String
		}
		machines = append(machines, m)
	}
	return machines, rows.Err()
}

// GetMachine fetches a single machine by ID (nil when absent). The returned
// copy never carries a machine token plaintext.
func (s *Store) GetMachine(ctx context.Context, machineID string) (*model.Machine, error) {
	if s == nil {
		return nil, nil
	}
	row := s.db.QueryRowContext(ctx, `SELECT id, session_id, sdk_version, sdk_language, ip, created_at, last_ping_at, token_hash, draining FROM machines WHERE id=$1`, machineID)
	m := &model.Machine{}
	var (
		sdkVersion, sdkLanguage, ip, tokenHash sql.NullString
		draining                               sql.NullBool
	)
	if err := row.Scan(&m.ID, &m.SessionID, &sdkVersion, &sdkLanguage, &ip, &m.CreatedAt, &m.LastPingAt, &tokenHash, &draining); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get machine: %w", err)
	}
	m.Draining = draining.Valid && draining.Bool
	if sdkVersion.Valid {
		m.SDKVersion = sdkVersion.String
	}
	if sdkLanguage.Valid {
		m.SDKLanguage = sdkLanguage.String
	}
	if ip.Valid {
		m.IP = ip.String
	}
	if tokenHash.Valid {
		m.TokenHash = tokenHash.String
	}
	return m, nil
}

func (s *Store) SaveMachine(ctx context.Context, machine *model.Machine) error {
	if s == nil || machine == nil {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
        INSERT INTO machines (id, session_id, sdk_version, sdk_language, ip, created_at, last_ping_at, token_hash)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
        ON CONFLICT (id) DO UPDATE SET
            session_id = EXCLUDED.session_id,
            sdk_version = EXCLUDED.sdk_version,
            sdk_language = EXCLUDED.sdk_language,
            ip = EXCLUDED.ip,
            -- Identity fields are first-registration-wins: a re-register on
            -- any replica (including one whose cold cache mints a fresh
            -- credential) must not silently rotate the machine's token or
            -- reset its age. The only writer of token_hash after first
            -- registration is BindMachineToken's compare-and-set.
            created_at = machines.created_at,
            last_ping_at = EXCLUDED.last_ping_at,
            token_hash = CASE
                WHEN machines.token_hash IS NULL OR machines.token_hash = ''
                THEN EXCLUDED.token_hash
                ELSE machines.token_hash
            END,
            draining = machines.draining -- drain state is written only by Set/ClearMachineDraining
    `, machine.ID, machine.SessionID, nullString(machine.SDKVersion), nullString(machine.SDKLanguage), nullString(machine.IP), machine.CreatedAt, machine.LastPingAt, nullString(machine.TokenHash))
	if err != nil {
		return fmt.Errorf("upsert machine: %w", err)
	}
	return nil
}

func (s *Store) DeleteMachine(ctx context.Context, machineID string) error {
	if s == nil {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM machines WHERE id=$1`, machineID); err != nil {
		return fmt.Errorf("delete machine: %w", err)
	}
	return nil
}

// TouchMachineLastPing advances only the heartbeat timestamp. A heartbeat is
// a continuous background write and must never rewrite the rest of the row:
// drain state belongs to SetMachineDraining / ClearMachineDraining, and
// credentials to BindMachineToken.
func (s *Store) TouchMachineLastPing(ctx context.Context, sessionID, machineID string, at time.Time) error {
	if s == nil {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE machines SET last_ping_at=$3 WHERE id=$1 AND session_id=$2`, machineID, sessionID, at); err != nil {
		return fmt.Errorf("touch machine last ping: %w", err)
	}
	return nil
}

// BindMachineToken claims the machine's credential slot: the write lands only
// while no hash is bound, and the return reports whether this call won. A
// plain overwrite would let two replicas bind different first credentials and
// each keep accepting its own.
func (s *Store) BindMachineToken(ctx context.Context, machineID, tokenHash string) (bool, error) {
	if s == nil {
		return false, nil
	}
	res, err := s.db.ExecContext(ctx, `UPDATE machines SET token_hash=$2 WHERE id=$1 AND (token_hash IS NULL OR token_hash='')`, machineID, nullString(tokenHash))
	if err != nil {
		return false, fmt.Errorf("bind machine token: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("bind machine token: %w", err)
	}
	return n > 0, nil
}

func (s *Store) ListStaleMachines(ctx context.Context, cutoff time.Time, limit int) ([]*model.Machine, error) {
	if s == nil {
		return nil, nil
	}
	query := `SELECT id, session_id, sdk_version, sdk_language, ip, created_at, last_ping_at, token_hash, draining FROM machines WHERE last_ping_at < $1 ORDER BY last_ping_at ASC`
	var rows *sql.Rows
	var err error
	if limit > 0 {
		query += ` LIMIT $2`
		rows, err = s.db.QueryContext(ctx, query, cutoff, limit)
	} else {
		rows, err = s.db.QueryContext(ctx, query, cutoff)
	}
	if err != nil {
		return nil, fmt.Errorf("list stale machines: %w", err)
	}
	defer rows.Close()

	var machines []*model.Machine
	for rows.Next() {
		m := &model.Machine{}
		var (
			sdkVersion, sdkLanguage, ip, tokenHash sql.NullString
			draining                               sql.NullBool
		)
		if err := rows.Scan(&m.ID, &m.SessionID, &sdkVersion, &sdkLanguage, &ip, &m.CreatedAt, &m.LastPingAt, &tokenHash, &draining); err != nil {
			return nil, fmt.Errorf("scan stale machine: %w", err)
		}
		m.Draining = draining.Valid && draining.Bool
		if sdkVersion.Valid {
			m.SDKVersion = sdkVersion.String
		}
		if sdkLanguage.Valid {
			m.SDKLanguage = sdkLanguage.String
		}
		if ip.Valid {
			m.IP = ip.String
		}
		if tokenHash.Valid {
			m.TokenHash = tokenHash.String
		}
		machines = append(machines, m)
	}
	return machines, rows.Err()
}

func (s *Store) ReclaimMachine(ctx context.Context, machineID string, cutoff time.Time) (string, []ToolOwnershipUpdate, bool, error) {
	if s == nil {
		return "", nil, false, nil
	}
	var (
		updates            []ToolOwnershipUpdate
		reclaimedSessionID string
	)
	err := s.withSerializableTx(ctx, func(tx *sql.Tx) error {
		var (
			lastPing  time.Time
			sessionID string
		)
		if err := tx.QueryRowContext(ctx, `SELECT session_id, last_ping_at FROM machines WHERE id=$1 FOR UPDATE`, machineID).Scan(&sessionID, &lastPing); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return sql.ErrNoRows
			}
			return fmt.Errorf("reclaim machine: load machine: %w", err)
		}

		if !lastPing.Before(cutoff) {
			return errMachineFresh
		}

		rows, err := tx.QueryContext(ctx, `UPDATE tools SET machine_id = NULL WHERE machine_id=$1 RETURNING id, session_id`, machineID)
		if err != nil {
			return fmt.Errorf("reclaim machine: detach tools: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var upd ToolOwnershipUpdate
			if err := rows.Scan(&upd.ToolID, &upd.SessionID); err != nil {
				return fmt.Errorf("reclaim machine: scan detached tool: %w", err)
			}
			updates = append(updates, upd)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `DELETE FROM machines WHERE id=$1`, machineID); err != nil {
			return fmt.Errorf("reclaim machine: delete machine: %w", err)
		}

		reclaimedSessionID = sessionID
		return nil
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil, false, nil
		}
		if errors.Is(err, errMachineFresh) {
			return "", nil, false, nil
		}
		return "", nil, false, err
	}

	return reclaimedSessionID, updates, true, nil
}

// SetMachineDraining marks the machine as draining so drain state is coherent
// across instances. It is idempotent.
func (s *Store) SetMachineDraining(ctx context.Context, sessionID, machineID string) error {
	if s == nil {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE machines SET draining=true WHERE id=$1 AND session_id=$2`, machineID, sessionID); err != nil {
		return fmt.Errorf("set machine draining: %w", err)
	}
	return nil
}

// ClearMachineDraining clears the drain flag, e.g. after the machine has been
// fully unregistered or the drain was aborted. It is idempotent.
func (s *Store) ClearMachineDraining(ctx context.Context, machineID string) error {
	if s == nil {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE machines SET draining=false WHERE id=$1`, machineID); err != nil {
		return fmt.Errorf("clear machine draining: %w", err)
	}
	return nil
}

// IsMachineDraining reports whether the machine is currently draining, reading
// the persisted flag so the answer is consistent across instances.
func (s *Store) IsMachineDraining(ctx context.Context, machineID string) (bool, error) {
	if s == nil {
		return false, nil
	}
	var draining bool
	err := s.db.QueryRowContext(ctx, `SELECT draining FROM machines WHERE id=$1`, machineID).Scan(&draining)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("is machine draining: %w", err)
	}
	return draining, nil
}
