package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"toolplane/pkg/model"
)

func (s *Store) AllSessions(ctx context.Context) ([]*model.Session, error) {
	if s == nil {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, description, namespace, created_at, created_by, api_key FROM sessions`)
	if err != nil {
		return nil, fmt.Errorf("query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*model.Session
	for rows.Next() {
		var rec model.Session
		var namespace sql.NullString
		var apiKey sql.NullString
		if err := rows.Scan(&rec.ID, &rec.Name, &rec.Description, &namespace, &rec.CreatedAt, &rec.CreatedBy, &apiKey); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		if namespace.Valid {
			rec.Namespace = namespace.String
		}
		if apiKey.Valid {
			rec.ApiKey = apiKey.String
		}
		sessions = append(sessions, &rec)
	}

	return sessions, rows.Err()
}

func (s *Store) SaveSession(ctx context.Context, session *model.Session) error {
	if s == nil || session == nil {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
        INSERT INTO sessions (id, name, description, namespace, created_at, created_by, api_key)
        VALUES ($1,$2,$3,$4,$5,$6,$7)
        ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            description = EXCLUDED.description,
            namespace = EXCLUDED.namespace,
            created_at = EXCLUDED.created_at,
            created_by = EXCLUDED.created_by,
            api_key = EXCLUDED.api_key
    `, session.ID, session.Name, session.Description, nullString(session.Namespace), session.CreatedAt, session.CreatedBy, nullString(session.ApiKey))
	if err != nil {
		return fmt.Errorf("upsert session: %w", err)
	}
	return nil
}

func (s *Store) DeleteSession(ctx context.Context, sessionID string) error {
	if s == nil {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE id=$1`, sessionID); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (s *Store) AllUserSessions(ctx context.Context) (map[string][]string, error) {
	if s == nil {
		return map[string][]string{}, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT user_id, session_id FROM user_sessions`)
	if err != nil {
		return nil, fmt.Errorf("query user_sessions: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]string)
	for rows.Next() {
		var userID, sessionID string
		if err := rows.Scan(&userID, &sessionID); err != nil {
			return nil, fmt.Errorf("scan user session: %w", err)
		}
		result[userID] = append(result[userID], sessionID)
	}
	return result, rows.Err()
}

func (s *Store) AddUserSession(ctx context.Context, userID, sessionID string) error {
	if s == nil {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
        INSERT INTO user_sessions (user_id, session_id)
        VALUES ($1, $2)
        ON CONFLICT (user_id, session_id) DO NOTHING
    `, userID, sessionID)
	if err != nil {
		return fmt.Errorf("upsert user session: %w", err)
	}
	return nil
}

func (s *Store) RemoveUserSession(ctx context.Context, userID, sessionID string) error {
	if s == nil {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE user_id=$1 AND session_id=$2`, userID, sessionID); err != nil {
		return fmt.Errorf("delete user session: %w", err)
	}
	return nil
}

func (s *Store) AllApiKeys(ctx context.Context) ([]*model.ApiKey, error) {
	if s == nil {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, session_id, name, key, key_hash, key_preview, capabilities, created_at, created_by, revoked_at FROM api_keys`)
	if err != nil {
		return nil, fmt.Errorf("query api keys: %w", err)
	}
	defer rows.Close()

	var keys []*model.ApiKey
	for rows.Next() {
		var rec model.ApiKey
		var key sql.NullString
		var keyHash sql.NullString
		var keyPreview sql.NullString
		var capabilitiesPayload []byte
		var revokedAt sql.NullTime
		if err := rows.Scan(&rec.ID, &rec.SessionID, &rec.Name, &key, &keyHash, &keyPreview, &capabilitiesPayload, &rec.CreatedAt, &rec.CreatedBy, &revokedAt); err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		if key.Valid {
			rec.Key = key.String
			rec.PlaintextPersisted = key.String != ""
		}
		if keyHash.Valid {
			rec.KeyHash = keyHash.String
		}
		if keyPreview.Valid {
			rec.KeyPreview = keyPreview.String
		}
		if len(capabilitiesPayload) > 0 {
			var capabilityValues []string
			if err := json.Unmarshal(capabilitiesPayload, &capabilityValues); err != nil {
				return nil, fmt.Errorf("unmarshal api key capabilities: %w", err)
			}
			capabilities, err := model.NormalizeAPIKeyCapabilities(capabilityValues)
			if err != nil {
				return nil, fmt.Errorf("normalize api key capabilities: %w", err)
			}
			rec.Capabilities = capabilities
		}
		if revokedAt.Valid {
			t := revokedAt.Time
			rec.RevokedAt = &t
		}
		rec.EnsureSecurityMetadata()
		keys = append(keys, &rec)
	}

	return keys, rows.Err()
}

func (s *Store) SaveApiKey(ctx context.Context, key *model.ApiKey) error {
	if s == nil || key == nil {
		return nil
	}
	key.EnsureSecurityMetadata()
	var revokedAt interface{}
	if key.RevokedAt != nil {
		revokedAt = *key.RevokedAt
	}
	capabilitiesPayload, err := json.Marshal(model.CapabilityStrings(key.Capabilities))
	if err != nil {
		return fmt.Errorf("marshal api key capabilities: %w", err)
	}
	var persistedKey interface{}
	if key.PlaintextPersisted && key.Key != "" {
		persistedKey = key.Key
	}
	_, err = s.db.ExecContext(ctx, `
	        INSERT INTO api_keys (id, session_id, name, key, key_hash, key_preview, capabilities, created_at, created_by, revoked_at)
	        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
        ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            key = EXCLUDED.key,
	            key_hash = EXCLUDED.key_hash,
	            key_preview = EXCLUDED.key_preview,
	            capabilities = EXCLUDED.capabilities,
            created_at = EXCLUDED.created_at,
            created_by = EXCLUDED.created_by,
            revoked_at = EXCLUDED.revoked_at
	    `, key.ID, key.SessionID, key.Name, persistedKey, nullString(key.KeyHash), nullString(key.KeyPreview), capabilitiesPayload, key.CreatedAt, key.CreatedBy, revokedAt)
	if err != nil {
		return fmt.Errorf("upsert api key: %w", err)
	}
	return nil
}

func (s *Store) DeleteApiKey(ctx context.Context, keyID string) error {
	if s == nil {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM api_keys WHERE id=$1`, keyID); err != nil {
		return fmt.Errorf("delete api key: %w", err)
	}
	return nil
}

func nullString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func nullableTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}
