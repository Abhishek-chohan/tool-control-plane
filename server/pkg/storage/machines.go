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
	rows, err := s.db.QueryContext(ctx, `SELECT id, session_id, sdk_version, sdk_language, ip, created_at, last_ping_at FROM machines`)
	if err != nil {
		return nil, fmt.Errorf("query machines: %w", err)
	}
	defer rows.Close()

	var machines []*model.Machine
	for rows.Next() {
		m := &model.Machine{}
		var sdkVersion, sdkLanguage, ip sql.NullString
		if err := rows.Scan(&m.ID, &m.SessionID, &sdkVersion, &sdkLanguage, &ip, &m.CreatedAt, &m.LastPingAt); err != nil {
			return nil, fmt.Errorf("scan machine: %w", err)
		}
		if sdkVersion.Valid {
			m.SDKVersion = sdkVersion.String
		}
		if sdkLanguage.Valid {
			m.SDKLanguage = sdkLanguage.String
		}
		if ip.Valid {
			m.IP = ip.String
		}
		machines = append(machines, m)
	}
	return machines, rows.Err()
}

func (s *Store) SaveMachine(ctx context.Context, machine *model.Machine) error {
	if s == nil || machine == nil {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
        INSERT INTO machines (id, session_id, sdk_version, sdk_language, ip, created_at, last_ping_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7)
        ON CONFLICT (id) DO UPDATE SET
            session_id = EXCLUDED.session_id,
            sdk_version = EXCLUDED.sdk_version,
            sdk_language = EXCLUDED.sdk_language,
            ip = EXCLUDED.ip,
            created_at = EXCLUDED.created_at,
            last_ping_at = EXCLUDED.last_ping_at
    `, machine.ID, machine.SessionID, nullString(machine.SDKVersion), nullString(machine.SDKLanguage), nullString(machine.IP), machine.CreatedAt, machine.LastPingAt)
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

func (s *Store) ListStaleMachines(ctx context.Context, cutoff time.Time, limit int) ([]*model.Machine, error) {
	if s == nil {
		return nil, nil
	}
	query := `SELECT id, session_id, sdk_version, sdk_language, ip, created_at, last_ping_at FROM machines WHERE last_ping_at < $1 ORDER BY last_ping_at ASC`
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
		var sdkVersion, sdkLanguage, ip sql.NullString
		if err := rows.Scan(&m.ID, &m.SessionID, &sdkVersion, &sdkLanguage, &ip, &m.CreatedAt, &m.LastPingAt); err != nil {
			return nil, fmt.Errorf("scan stale machine: %w", err)
		}
		if sdkVersion.Valid {
			m.SDKVersion = sdkVersion.String
		}
		if sdkLanguage.Valid {
			m.SDKLanguage = sdkLanguage.String
		}
		if ip.Valid {
			m.IP = ip.String
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
