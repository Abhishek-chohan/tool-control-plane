package service

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
	"toolplane/pkg/trace"
)

// SessionsService handles session-related operations
type SessionsService struct {
	// In-memory storage for sessions
	sessions      map[string]*model.Session // map[sessionID]Session
	sessionsMutex sync.RWMutex

	// In-memory storage for API keys
	apiKeys      map[string]map[string]*model.ApiKey // map[sessionID]map[keyID]ApiKey
	apiKeysMutex sync.RWMutex

	// keyIndex maps SHA-256(key secret) -> key for O(1) authentication.
	// Guarded by apiKeysMutex; revoked/deleted keys are removed so the index
	// only ever contains live grants.
	keyIndex map[string]*model.ApiKey

	// keyVerifiedAt records when keyIndex entries were last confirmed
	// against the store. Guarded by apiKeysMutex. Entries older than
	// apiKeyRevalidationInterval are re-read so revocations and session
	// deletions made on other replicas propagate within the interval.
	keyVerifiedAt map[string]time.Time

	// In-memory mapping of users to sessions
	userSessions      map[string][]string // map[userID][]sessionID
	userSessionsMutex sync.RWMutex

	userLocks sync.Map // map[userID]*sync.Mutex to serialize per-user changes

	tracer trace.SessionTracer
	store  storage.Storer
}

// NewSessionsService creates a new sessions service
func NewSessionsService(tracer trace.SessionTracer, store storage.Storer) *SessionsService {
	if tracer == nil {
		tracer = trace.NopTracer()
	}

	svc := &SessionsService{
		sessions:      make(map[string]*model.Session),
		apiKeys:       make(map[string]map[string]*model.ApiKey),
		keyIndex:      make(map[string]*model.ApiKey),
		keyVerifiedAt: make(map[string]time.Time),
		userSessions:  make(map[string][]string),
		tracer:        tracer,
		store:         store,
	}
	if store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()

		if sessions, err := store.AllSessions(ctx); err != nil {
			log.Printf("sessions persistence load failed: %v", err)
		} else {
			for _, session := range sessions {
				svc.sessions[session.ID] = session
			}
		}

		if keys, err := store.AllApiKeys(ctx); err != nil {
			log.Printf("api key persistence load failed: %v", err)
		} else {
			for _, key := range keys {
				if key == nil {
					continue
				}
				if key.EnsureSecurityMetadata() {
					persistCtx, persistCancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
					if err := store.SaveApiKey(persistCtx, key); err != nil {
						log.Printf("persist api key hardening migration failed: %v", err)
					}
					persistCancel()
				}
				if _, ok := svc.apiKeys[key.SessionID]; !ok {
					svc.apiKeys[key.SessionID] = make(map[string]*model.ApiKey)
				}
				svc.apiKeys[key.SessionID][key.ID] = key
				if !key.IsRevoked() {
					svc.keyIndex[key.KeyHash] = key
					svc.keyVerifiedAt[key.KeyHash] = time.Now()
				}
			}
		}

		if mapping, err := store.AllUserSessions(ctx); err != nil {
			log.Printf("user session persistence load failed: %v", err)
		} else {
			for userID, sessionIDs := range mapping {
				copied := make([]string, len(sessionIDs))
				copy(copied, sessionIDs)
				svc.userSessions[userID] = copied
			}
		}

		for _, session := range svc.sessions {
			if session == nil {
				continue
			}
			if session.ApiKey != "" {
				session.ApiKey = ""
				persistCtx, persistCancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
				if err := store.SaveSession(persistCtx, session); err != nil {
					log.Printf("persist legacy session api key retirement failed: %v", err)
				}
				persistCancel()
			}
			list, ok := svc.userSessions[session.CreatedBy]
			if !ok {
				svc.userSessions[session.CreatedBy] = []string{session.ID}
				continue
			}
			found := false
			for _, id := range list {
				if id == session.ID {
					found = true
					break
				}
			}
			if !found {
				svc.userSessions[session.CreatedBy] = append(list, session.ID)
			}
		}
	}
	return svc
}

// CreateSession creates a new session with an optional api key and an initial API key
func (s *SessionsService) CreateSession(userID, name, description, apiKey, requestedID, namespace string) (*model.Session, error) {
	_ = apiKey
	userLock := s.userLock(userID)
	userLock.Lock()
	defer userLock.Unlock()

	// If client provided a session ID, ensure it does not already exist
	s.sessionsMutex.RLock()
	if requestedID != "" {
		if _, exists := s.sessions[requestedID]; exists {
			s.sessionsMutex.RUnlock()
			return nil, wrapf(ErrAlreadyExists, "session %s already exists", requestedID)
		}
	}
	s.sessionsMutex.RUnlock()

	// Create new session (random ID) then override if requested
	session := model.NewSession(name, description, userID, "", namespace)
	if requestedID != "" {
		session.ID = requestedID
	}

	// Store session
	s.sessionsMutex.Lock()
	s.sessions[session.ID] = session
	s.sessionsMutex.Unlock()

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		// Insert-if-absent dedups across replicas: two instances racing to
		// create the same client-supplied session ID agree on one winner and
		// the loser reports AlreadyExists instead of overwriting the row.
		inserted, err := s.store.InsertSessionIfAbsent(ctx, session)
		if err != nil {
			log.Printf("persist session create failed: %v", err)
		} else if !inserted {
			s.sessionsMutex.Lock()
			delete(s.sessions, session.ID)
			s.sessionsMutex.Unlock()
			return nil, wrapf(ErrAlreadyExists, "session %s already exists", session.ID)
		}
	}

	// Associate session with user
	s.userSessionsMutex.Lock()
	if _, ok := s.userSessions[userID]; !ok {
		s.userSessions[userID] = make([]string, 0)
	}
	s.userSessions[userID] = append(s.userSessions[userID], session.ID)
	s.userSessionsMutex.Unlock()

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if err := s.store.AddUserSession(ctx, userID, session.ID); err != nil {
			log.Printf("persist user session mapping failed: %v", err)
		}
	}

	s.recordSessionEvent(session.ID, "", trace.EventSessionCreated, session.CreatedAt, map[string]any{
		"createdBy": userID,
		"name":      session.Name,
		"namespace": session.Namespace,
	})

	return session, nil
}

// GetSessionByID gets a session by ID
func (s *SessionsService) GetSessionByID(sessionID string) (*model.Session, error) {
	s.sessionsMutex.RLock()
	session, ok := s.sessions[sessionID]
	s.sessionsMutex.RUnlock()
	if ok {
		return session, nil
	}

	// Cache miss: fall back to the store so a session created on another
	// instance is visible here (multi-instance read-through). Populate the
	// local cache on a successful lookup so subsequent reads stay fast.
	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		found, err := s.store.GetSession(ctx, sessionID)
		if err != nil {
			return nil, fmt.Errorf("session %s lookup failed: %w", sessionID, err)
		}
		if found == nil {
			return nil, wrapf(ErrNotFound, "session %s not found", sessionID)
		}
		s.sessionsMutex.Lock()
		if _, exists := s.sessions[sessionID]; !exists {
			s.sessions[sessionID] = found
		}
		s.sessionsMutex.Unlock()
		return found, nil
	}

	return nil, wrapf(ErrNotFound, "session %s not found", sessionID)
}

// ListSessions lists all sessions for a user
func (s *SessionsService) ListSessions(userID string) ([]*model.Session, error) {
	s.userSessionsMutex.RLock()
	sessionIDs, ok := s.userSessions[userID]
	s.userSessionsMutex.RUnlock()

	if !ok {
		return []*model.Session{}, nil
	}

	s.sessionsMutex.RLock()
	defer s.sessionsMutex.RUnlock()

	sessions := make([]*model.Session, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		if session, ok := s.sessions[sessionID]; ok {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

// UpdateSession updates a session
func (s *SessionsService) UpdateSession(sessionID, name, description, namespace string) (*model.Session, error) {
	s.sessionsMutex.Lock()
	defer s.sessionsMutex.Unlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, wrapf(ErrNotFound, "session %s not found", sessionID)
	}

	// Update fields if provided
	if name != "" {
		session.Name = name
	}

	if description != "" {
		session.Description = description
	}

	if namespace != "" {
		session.Namespace = namespace
	}

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if err := s.store.SaveSession(ctx, session); err != nil {
			log.Printf("persist session update failed: %v", err)
		}
	}

	s.recordSessionEvent(session.ID, "", trace.EventSessionUpdated, time.Now(), map[string]any{
		"name":        session.Name,
		"description": session.Description,
		"namespace":   session.Namespace,
	})

	return session, nil
}

// DeleteSession deletes a session and all its associated API keys
func (s *SessionsService) DeleteSession(sessionID string) error {
	s.sessionsMutex.RLock()
	session, ok := s.sessions[sessionID]
	s.sessionsMutex.RUnlock()
	if !ok {
		return wrapf(ErrNotFound, "session %s not found", sessionID)
	}

	userLock := s.userLock(session.CreatedBy)
	userLock.Lock()
	defer userLock.Unlock()

	s.sessionsMutex.Lock()
	// re-check under write lock in case it was removed concurrently
	session, ok = s.sessions[sessionID]
	if !ok {
		s.sessionsMutex.Unlock()
		return wrapf(ErrNotFound, "session %s not found", sessionID)
	}
	delete(s.sessions, sessionID)
	s.sessionsMutex.Unlock()

	userID := session.CreatedBy

	s.userSessionsMutex.Lock()
	if userSessions, ok := s.userSessions[userID]; ok {
		filtered := userSessions[:0]
		for _, id := range userSessions {
			if id != sessionID {
				filtered = append(filtered, id)
			}
		}
		if len(filtered) == 0 {
			delete(s.userSessions, userID)
		} else {
			s.userSessions[userID] = append([]string(nil), filtered...)
		}
	}
	s.userSessionsMutex.Unlock()

	s.apiKeysMutex.Lock()
	for _, apiKey := range s.apiKeys[sessionID] {
		if apiKey != nil {
			delete(s.keyIndex, apiKey.KeyHash)
			delete(s.keyVerifiedAt, apiKey.KeyHash)
		}
	}
	delete(s.apiKeys, sessionID)
	s.apiKeysMutex.Unlock()

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if err := s.store.RemoveUserSession(ctx, userID, sessionID); err != nil {
			log.Printf("persist user session delete failed: %v", err)
		}
		if err := s.store.DeleteSession(ctx, sessionID); err != nil {
			log.Printf("persist session delete failed: %v", err)
		}
	}

	s.recordSessionEvent(sessionID, "", trace.EventSessionDeleted, time.Now(), map[string]any{
		"createdBy": userID,
		"name":      session.Name,
	})

	return nil
}

// CreateApiKey creates a new API key for a session
func (s *SessionsService) CreateApiKey(sessionID, name, createdBy string, capabilityValues []string) (*model.ApiKey, error) {
	// Check if session exists
	s.sessionsMutex.RLock()
	_, ok := s.sessions[sessionID]
	s.sessionsMutex.RUnlock()

	if !ok {
		return nil, wrapf(ErrNotFound, "session %s not found", sessionID)
	}

	capabilities, err := model.ParseAPIKeyCapabilitiesStrict(capabilityValues)
	if err != nil {
		return nil, err
	}

	// Create new API key
	apiKey := model.NewApiKey(name, sessionID, createdBy, capabilities)

	s.apiKeysMutex.Lock()
	defer s.apiKeysMutex.Unlock()

	// Initialize API keys map for this session if not exists
	if _, ok := s.apiKeys[sessionID]; !ok {
		s.apiKeys[sessionID] = make(map[string]*model.ApiKey)
	}

	// Store API key
	s.apiKeys[sessionID][apiKey.ID] = apiKey
	s.keyIndex[apiKey.KeyHash] = apiKey
	s.keyVerifiedAt[apiKey.KeyHash] = time.Now()

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if err := s.store.SaveApiKey(ctx, apiKey); err != nil {
			log.Printf("persist api key create failed: %v", err)
		}
	}

	s.recordSessionEvent(sessionID, "", trace.EventAPIKeyCreated, apiKey.CreatedAt, map[string]any{
		"apiKeyID":      apiKey.ID,
		"name":          apiKey.Name,
		"createdBy":     createdBy,
		"capabilities":  model.CapabilityStrings(apiKey.Capabilities),
		"keyPreview":    apiKey.KeyPreview,
		"storageBacked": s.store != nil,
	})

	return apiKey, nil
}

// ListApiKeys lists all API keys for a session
func (s *SessionsService) ListApiKeys(sessionID string) ([]*model.ApiKey, error) {
	s.apiKeysMutex.RLock()
	defer s.apiKeysMutex.RUnlock()

	if _, ok := s.apiKeys[sessionID]; !ok {
		return []*model.ApiKey{}, nil
	}

	// Get all API keys for this session
	apiKeys := make([]*model.ApiKey, 0, len(s.apiKeys[sessionID]))
	for _, apiKey := range s.apiKeys[sessionID] {
		// Skip revoked keys
		if !apiKey.IsRevoked() {
			apiKeys = append(apiKeys, apiKey.CloneWithoutSecret())
		}
	}

	return apiKeys, nil
}

// RevokeApiKey revokes an API key
func (s *SessionsService) RevokeApiKey(sessionID, keyID string) error {
	s.apiKeysMutex.Lock()
	defer s.apiKeysMutex.Unlock()

	if _, ok := s.apiKeys[sessionID]; !ok {
		return wrapf(ErrNotFound, "no API keys found for session %s", sessionID)
	}

	apiKey, ok := s.apiKeys[sessionID][keyID]
	if !ok {
		return wrapf(ErrNotFound, "API key %s not found for session %s", keyID, sessionID)
	}

	// Revoke the API key
	apiKey.Revoke()
	delete(s.keyIndex, apiKey.KeyHash)
	delete(s.keyVerifiedAt, apiKey.KeyHash)

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if err := s.store.SaveApiKey(ctx, apiKey); err != nil {
			log.Printf("persist api key revoke failed: %v", err)
		}
	}

	revokedAt := apiKey.CreatedAt
	if apiKey.RevokedAt != nil {
		revokedAt = *apiKey.RevokedAt
	}
	s.recordSessionEvent(sessionID, "", trace.EventAPIKeyRevoked, revokedAt, map[string]any{
		"apiKeyID":     apiKey.ID,
		"name":         apiKey.Name,
		"capabilities": model.CapabilityStrings(apiKey.Capabilities),
	})

	return nil
}

func (s *SessionsService) AuthenticateAPIKey(key string) (*model.AuthPrincipal, error) {
	normalizedKey := strings.TrimSpace(key)
	if strings.HasPrefix(strings.ToLower(normalizedKey), "bearer ") {
		normalizedKey = strings.TrimSpace(normalizedKey[7:])
	}
	if normalizedKey == "" {
		return nil, fmt.Errorf("api key is empty")
	}

	// O(1) lookup by SHA-256 hash instead of scanning every stored key. The
	// index only contains live (non-revoked) grants; comparing hashes is safe
	// because recovering a secret from its SHA-256 is infeasible.
	keyHash := model.HashAPIKeySecret(normalizedKey)
	s.apiKeysMutex.RLock()
	matchedKey := s.keyIndex[keyHash]
	verifiedAt := s.keyVerifiedAt[keyHash]
	s.apiKeysMutex.RUnlock()

	if matchedKey == nil {
		return nil, fmt.Errorf("api key not found")
	}
	if matchedKey.IsRevoked() {
		return nil, fmt.Errorf("api key is revoked")
	}

	// Revalidate against the store when the cached grant is stale so a
	// revocation or session deletion performed on another replica stops
	// authenticating within apiKeyRevalidationInterval. A store error keeps
	// the cached grant (availability over revocation latency) and retries on
	// the next authentication.
	if s.store != nil && time.Since(verifiedAt) > apiKeyRevalidationInterval {
		matchedKey = s.revalidateAPIKeyLocked(keyHash, matchedKey)
		if matchedKey == nil {
			return nil, fmt.Errorf("api key not found")
		}
		if matchedKey.IsRevoked() {
			return nil, fmt.Errorf("api key is revoked")
		}
	}
	matchedSessionID := matchedKey.SessionID

	userID := ""
	s.sessionsMutex.RLock()
	if session, ok := s.sessions[matchedSessionID]; ok && session != nil {
		userID = session.CreatedBy
	}
	s.sessionsMutex.RUnlock()

	return &model.AuthPrincipal{
		Mode:         model.AuthModeSessionKey,
		SessionID:    matchedSessionID,
		UserID:       userID,
		KeyID:        matchedKey.ID,
		Capabilities: append([]model.APIKeyCapability(nil), matchedKey.Capabilities...),
		TokenPreview: matchedKey.KeyPreview,
	}, nil
}

// apiKeyRevalidationInterval bounds how long a cached API-key grant stays
// trusted without a store confirmation. Revocations propagate to every
// replica within roughly this interval. It is a variable so tests can
// shorten it instead of sleeping.
var apiKeyRevalidationInterval = 5 * time.Second

// revalidateAPIKeyLocked re-reads the key by hash from the store and refreshes
// the cache. It returns the (possibly replaced) live key, or nil when the key
// is gone or revoked — the caller treats both as authentication failure and
// the cache entries are dropped so the index only holds live grants.
func (s *SessionsService) revalidateAPIKeyLocked(keyHash string, cached *model.ApiKey) *model.ApiKey {
	ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
	defer cancel()
	fresh, err := s.store.GetAPIKeyByHash(ctx, keyHash)
	if err != nil {
		log.Printf("api key revalidation lookup failed: %v", err)
		return cached
	}
	if fresh == nil || fresh.IsRevoked() {
		s.apiKeysMutex.Lock()
		if current, ok := s.keyIndex[keyHash]; ok && current == cached {
			delete(s.keyIndex, keyHash)
			delete(s.keyVerifiedAt, keyHash)
		}
		s.apiKeysMutex.Unlock()
		return nil
	}
	s.apiKeysMutex.Lock()
	s.keyIndex[keyHash] = fresh
	s.keyVerifiedAt[keyHash] = time.Now()
	s.apiKeysMutex.Unlock()
	return fresh
}

// ValidateApiKey validates an API key and returns the session ID if valid
func (s *SessionsService) ValidateApiKey(key string) (string, error) {
	principal, err := s.AuthenticateAPIKey(key)
	if err != nil {
		return "", err
	}
	return principal.SessionID, nil
}

// ListUserSessions lists sessions for a user with pagination and filtering
func (s *SessionsService) ListUserSessions(userID string, pageSize, pageToken int, filter string) ([]*model.Session, int, error) {
	// Get all sessions for the user
	s.userSessionsMutex.RLock()
	sessionIDs, ok := s.userSessions[userID]
	s.userSessionsMutex.RUnlock()

	if !ok || len(sessionIDs) == 0 {
		return []*model.Session{}, 0, nil
	}

	// Apply filtering if specified
	var filteredSessions []*model.Session
	s.sessionsMutex.RLock()
	for _, sessionID := range sessionIDs {
		if session, exists := s.sessions[sessionID]; exists {
			// Apply filter if specified (simplified implementation)
			if filter == "" ||
				(filter == "active" && s.sessionHasActiveAPIKeys(sessionID)) ||
				(filter == "inactive" && !s.sessionHasActiveAPIKeys(sessionID)) {
				filteredSessions = append(filteredSessions, session)
			}
		}
	}
	s.sessionsMutex.RUnlock()

	sort.SliceStable(filteredSessions, func(i, j int) bool {
		left := filteredSessions[i]
		right := filteredSessions[j]
		if left.CreatedAt.Equal(right.CreatedAt) {
			return left.ID > right.ID
		}
		return left.CreatedAt.After(right.CreatedAt)
	})

	// Calculate total count
	totalCount := len(filteredSessions)

	// Apply pagination
	if pageSize <= 0 {
		pageSize = 10 // Default page size
	}

	startIdx := pageToken * pageSize
	if startIdx >= totalCount {
		return []*model.Session{}, totalCount, nil
	}

	endIdx := startIdx + pageSize
	if endIdx > totalCount {
		endIdx = totalCount
	}

	return filteredSessions[startIdx:endIdx], totalCount, nil
}

// BulkDeleteSessions deletes multiple sessions for a user
func (s *SessionsService) BulkDeleteSessions(userID string, sessionIDs []string, filter string) (int, []string, error) {
	s.userSessionsMutex.RLock()
	userSessionIDs, ok := s.userSessions[userID]
	s.userSessionsMutex.RUnlock()

	if !ok {
		return 0, []string{}, nil
	}

	// Create a set of user's session IDs for quick lookup
	userSessionSet := make(map[string]bool)
	for _, id := range userSessionIDs {
		userSessionSet[id] = true
	}

	// If filter is specified, get sessions based on filter
	var sessionsToDelete []string
	if filter != "" {
		s.sessionsMutex.RLock()
		for _, sessionID := range userSessionIDs {
			if _, exists := s.sessions[sessionID]; exists {
				// Apply filter (simplified implementation)
				if (filter == "active" && s.sessionHasActiveAPIKeys(sessionID)) ||
					(filter == "inactive" && !s.sessionHasActiveAPIKeys(sessionID)) {
					sessionsToDelete = append(sessionsToDelete, sessionID)
				}
			}
		}
		s.sessionsMutex.RUnlock()
	} else {
		// Validate that all specified session IDs belong to the user
		for _, sessionID := range sessionIDs {
			if userSessionSet[sessionID] {
				sessionsToDelete = append(sessionsToDelete, sessionID)
			}
		}
	}

	// Delete the sessions
	failedDeletions := []string{}
	deletedCount := 0

	for _, sessionID := range sessionsToDelete {
		if err := s.DeleteSession(sessionID); err != nil {
			failedDeletions = append(failedDeletions, sessionID)
		} else {
			deletedCount++
		}
	}

	return deletedCount, failedDeletions, nil
}

// GetSessionStats retrieves statistics about user sessions
func (s *SessionsService) GetSessionStats(userID string) (int, int, int, error) {
	s.userSessionsMutex.RLock()
	sessionIDs, ok := s.userSessions[userID]
	s.userSessionsMutex.RUnlock()

	if !ok {
		return 0, 0, 0, nil
	}

	totalSessions := len(sessionIDs)
	activeSessions := 0
	inactiveSessions := 0

	s.sessionsMutex.RLock()
	for _, sessionID := range sessionIDs {
		if _, exists := s.sessions[sessionID]; exists {
			if s.sessionHasActiveAPIKeys(sessionID) {
				activeSessions++
			} else {
				inactiveSessions++
			}
		}
	}
	s.sessionsMutex.RUnlock()

	return totalSessions, activeSessions, inactiveSessions, nil
}

// InvalidateSession is the session-wide kill switch: it revokes every live
// API key of the session so no credential for that session authenticates
// again. The session record itself is kept (DeleteSession removes it); this
// is the incident-response path for suspected key compromise. It returns the
// number of keys revoked.
func (s *SessionsService) InvalidateSession(sessionID, reason string) (int, error) {
	s.apiKeysMutex.Lock()
	keys := make([]*model.ApiKey, 0, len(s.apiKeys[sessionID]))
	for _, apiKey := range s.apiKeys[sessionID] {
		if apiKey == nil || apiKey.IsRevoked() {
			continue
		}
		apiKey.Revoke()
		delete(s.keyIndex, apiKey.KeyHash)
		delete(s.keyVerifiedAt, apiKey.KeyHash)
		keys = append(keys, apiKey)
	}
	s.apiKeysMutex.Unlock()

	if s.store != nil {
		for _, apiKey := range keys {
			ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
			if err := s.store.SaveApiKey(ctx, apiKey); err != nil {
				log.Printf("persist session invalidation revoke failed for key %s: %v", apiKey.ID, err)
			}
			cancel()
		}
	}

	s.recordSessionEvent(sessionID, "", trace.EventAPIKeyRevoked, time.Now(), map[string]any{
		"reason":       "session_invalidated:" + reason,
		"revokedCount": len(keys),
	})

	return len(keys), nil
}

func (s *SessionsService) userLock(userID string) *sync.Mutex {
	key := userID
	if key == "" {
		key = "__anon__"
	}
	lock, _ := s.userLocks.LoadOrStore(key, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func (s *SessionsService) recordSessionEvent(sessionID, machineID string, event trace.SessionEventType, timestamp time.Time, metadata map[string]any) {
	if s == nil || s.tracer == nil {
		return
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	s.tracer.Record(trace.SessionEvent{
		SessionID: sessionID,
		MachineID: machineID,
		Event:     event,
		Timestamp: timestamp,
		Metadata:  metadata,
	})
}

func (s *SessionsService) sessionHasActiveAPIKeys(sessionID string) bool {
	s.apiKeysMutex.RLock()
	defer s.apiKeysMutex.RUnlock()

	keys, ok := s.apiKeys[sessionID]
	if !ok {
		return false
	}
	for _, apiKey := range keys {
		if apiKey != nil && !apiKey.IsRevoked() {
			return true
		}
	}
	return false
}
