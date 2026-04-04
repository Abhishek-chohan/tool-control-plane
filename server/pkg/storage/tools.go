package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"toolplane/pkg/model"
)

type ToolOwnershipUpdate struct {
	ToolID    string
	SessionID string
}

func (s *Store) AllTools(ctx context.Context) ([]*model.Tool, error) {
	if s == nil {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, session_id, machine_id, name, description, schema, config, tags, created_at, last_ping_at FROM tools`)
	if err != nil {
		return nil, fmt.Errorf("query tools: %w", err)
	}
	defer rows.Close()

	var tools []*model.Tool
	for rows.Next() {
		t := &model.Tool{}
		var machineID, description sql.NullString
		var configBytes, tagsBytes []byte
		if err := rows.Scan(&t.ID, &t.SessionID, &machineID, &t.Name, &description, &t.Schema, &configBytes, &tagsBytes, &t.CreatedAt, &t.LastPingAt); err != nil {
			return nil, fmt.Errorf("scan tool: %w", err)
		}
		if machineID.Valid {
			t.MachineID = machineID.String
		}
		if description.Valid {
			t.Description = description.String
		}
		if len(configBytes) > 0 {
			if err := fromJSON(configBytes, &t.Config); err != nil {
				return nil, fmt.Errorf("decode tool config: %w", err)
			}
		} else {
			t.Config = make(map[string]interface{})
		}
		if len(tagsBytes) > 0 {
			if err := fromJSON(tagsBytes, &t.Tags); err != nil {
				return nil, fmt.Errorf("decode tool tags: %w", err)
			}
		}
		tools = append(tools, t)
	}
	return tools, rows.Err()
}

func (s *Store) SaveTool(ctx context.Context, tool *model.Tool) error {
	if s == nil || tool == nil {
		return nil
	}
	configBytes := toJSON(tool.Config)
	if tool.Config == nil {
		configBytes = []byte("{}")
	}
	tagsBytes := toJSON(tool.Tags)
	if tool.Tags == nil {
		tagsBytes = []byte("[]")
	}
	_, err := s.db.ExecContext(ctx, `
        INSERT INTO tools (id, session_id, machine_id, name, description, schema, config, tags, created_at, last_ping_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
        ON CONFLICT (id) DO UPDATE SET
            session_id = EXCLUDED.session_id,
            machine_id = EXCLUDED.machine_id,
            name = EXCLUDED.name,
            description = EXCLUDED.description,
            schema = EXCLUDED.schema,
            config = EXCLUDED.config,
            tags = EXCLUDED.tags,
            created_at = EXCLUDED.created_at,
            last_ping_at = EXCLUDED.last_ping_at
    `, tool.ID, tool.SessionID, nullString(tool.MachineID), tool.Name, nullString(tool.Description), tool.Schema, configBytes, tagsBytes, tool.CreatedAt, tool.LastPingAt)
	if err != nil {
		return fmt.Errorf("upsert tool: %w", err)
	}
	return nil
}

func (s *Store) DeleteTool(ctx context.Context, toolID string) error {
	if s == nil {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM tools WHERE id=$1`, toolID); err != nil {
		return fmt.Errorf("delete tool: %w", err)
	}
	return nil
}

func (s *Store) DeleteToolsByMachine(ctx context.Context, machineID string) ([]ToolOwnershipUpdate, error) {
	if s == nil {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `DELETE FROM tools WHERE machine_id=$1 RETURNING id, session_id`, machineID)
	if err != nil {
		return nil, fmt.Errorf("delete tools by machine: %w", err)
	}
	defer rows.Close()

	var ids []ToolOwnershipUpdate
	for rows.Next() {
		var upd ToolOwnershipUpdate
		if err := rows.Scan(&upd.ToolID, &upd.SessionID); err != nil {
			return nil, fmt.Errorf("scan deleted tool id: %w", err)
		}
		ids = append(ids, upd)
	}
	return ids, rows.Err()
}

func (s *Store) ClaimToolOwnership(ctx context.Context, tool *model.Tool, staleCutoff time.Time) (*model.Tool, string, error) {
	if s == nil {
		return nil, "", errors.New("storage: store is nil")
	}
	if tool == nil {
		return nil, "", errors.New("storage: tool is nil")
	}

	var updated *model.Tool
	var replaced string

	err := s.withSerializableTx(ctx, func(tx *sql.Tx) error {
		var machineLastPing time.Time
		if err := tx.QueryRowContext(ctx, `SELECT last_ping_at FROM machines WHERE id=$1 AND session_id=$2 FOR UPDATE`, tool.MachineID, tool.SessionID).Scan(&machineLastPing); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("claim tool: machine %s not registered", tool.MachineID)
			}
			return fmt.Errorf("claim tool: load machine: %w", err)
		}
		if machineLastPing.Before(staleCutoff) {
			return fmt.Errorf("claim tool: machine %s heartbeat expired", tool.MachineID)
		}

		// Keep timestamps stable for insert/update operations
		createdAt := tool.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now()
		}
		lastPingAt := tool.LastPingAt
		if lastPingAt.IsZero() {
			lastPingAt = createdAt
		}

		configBytes := toJSON(tool.Config)
		if tool.Config == nil {
			configBytes = []byte("{}")
		}
		tagsBytes := toJSON(tool.Tags)
		if tool.Tags == nil {
			tagsBytes = []byte("[]")
		}

		row := tx.QueryRowContext(ctx, `SELECT id, machine_id, created_at FROM tools WHERE session_id=$1 AND name=$2 FOR UPDATE`, tool.SessionID, tool.Name)

		var (
			existingID        string
			existingMachineID sql.NullString
			existingCreatedAt time.Time
		)

		switch err := row.Scan(&existingID, &existingMachineID, &existingCreatedAt); err {
		case nil:
			if existingMachineID.Valid && existingMachineID.String != tool.MachineID {
				// ensure prior owner is stale before taking ownership
				var priorLastPing time.Time
				err := tx.QueryRowContext(ctx, `SELECT last_ping_at FROM machines WHERE id=$1 FOR UPDATE`, existingMachineID.String).Scan(&priorLastPing)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					return fmt.Errorf("claim tool: load prior machine: %w", err)
				}
				if err == nil && !priorLastPing.Before(staleCutoff) {
					return ErrToolOwnershipConflict
				}
				replaced = existingMachineID.String
			}

			if _, err := tx.ExecContext(ctx, `
                UPDATE tools
                SET machine_id = $1,
                    description = $2,
                    schema = $3,
                    config = $4,
                    tags = $5,
                    last_ping_at = $6
                WHERE id = $7
            `, tool.MachineID, nullString(tool.Description), tool.Schema, configBytes, tagsBytes, lastPingAt, existingID); err != nil {
				return fmt.Errorf("claim tool: update existing: %w", err)
			}

			updatedTool := &model.Tool{
				ID:          existingID,
				SessionID:   tool.SessionID,
				MachineID:   tool.MachineID,
				Name:        tool.Name,
				Description: tool.Description,
				Schema:      tool.Schema,
				Config:      cloneConfig(tool.Config),
				Tags:        cloneTags(tool.Tags),
				CreatedAt:   existingCreatedAt,
				LastPingAt:  lastPingAt,
			}
			updated = updatedTool
			return nil

		case sql.ErrNoRows:
			if _, err := tx.ExecContext(ctx, `
                INSERT INTO tools (id, session_id, machine_id, name, description, schema, config, tags, created_at, last_ping_at)
                VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
            `, tool.ID, tool.SessionID, nullString(tool.MachineID), tool.Name, nullString(tool.Description), tool.Schema, configBytes, tagsBytes, createdAt, lastPingAt); err != nil {
				return fmt.Errorf("claim tool: insert new: %w", err)
			}
			inserted := *tool
			inserted.CreatedAt = createdAt
			inserted.LastPingAt = lastPingAt
			inserted.Config = cloneConfig(tool.Config)
			inserted.Tags = cloneTags(tool.Tags)
			updated = &inserted
			return nil

		default:
			return fmt.Errorf("claim tool: select existing: %w", err)
		}
	})

	if err != nil {
		return nil, replaced, err
	}

	if updated == nil {
		return nil, replaced, fmt.Errorf("claim tool: no tool returned")
	}

	return updated, replaced, nil
}

func cloneConfig(src map[string]interface{}) map[string]interface{} {
	if len(src) == 0 {
		return make(map[string]interface{})
	}
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneTags(src []string) []string {
	if len(src) == 0 {
		return []string{}
	}
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}
