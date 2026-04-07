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

	// In-memory mapping of users to sessions
	userSessions      map[string][]string // map[userID][]sessionID
	userSessionsMutex sync.RWMutex

	userLocks sync.Map // map[userID]*sync.Mutex to serialize per-user changes

	tracer trace.SessionTracer
	store  *storage.Store
}

// NewSessionsService creates a new sessions service
func NewSessionsService(tracer trace.SessionTracer, store *storage.Store) *SessionsService {
	if tracer == nil {
		tracer = trace.NopTracer()
	}

	svc := &SessionsService{
		sessions:     make(map[string]*model.Session),
		apiKeys:      make(map[string]map[string]*model.ApiKey),
		userSessions: make(map[string][]string),
		tracer:       tracer,
		store:        store,
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
			return nil, fmt.Errorf("session %s already exists", requestedID)
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
		if err := s.store.SaveSession(ctx, session); err != nil {
			log.Printf("persist session create failed: %v", err)
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
	defer s.sessionsMutex.RUnlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	return session, nil
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
		return nil, fmt.Errorf("session %s not found", sessionID)
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
		return fmt.Errorf("session %s not found", sessionID)
	}

	userLock := s.userLock(session.CreatedBy)
	userLock.Lock()
	defer userLock.Unlock()

	s.sessionsMutex.Lock()
	// re-check under write lock in case it was removed concurrently
	session, ok = s.sessions[sessionID]
	if !ok {
		s.sessionsMutex.Unlock()
		return fmt.Errorf("session %s not found", sessionID)
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
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	capabilities, err := model.NormalizeAPIKeyCapabilities(capabilityValues)
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
		return fmt.Errorf("no API keys found for session %s", sessionID)
	}

	apiKey, ok := s.apiKeys[sessionID][keyID]
	if !ok {
		return fmt.Errorf("API key %s not found for session %s", keyID, sessionID)
	}

	// Revoke the API key
	apiKey.Revoke()

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

	var matchedKey *model.ApiKey
	matchedSessionID := ""
	revokedMatch := false

	s.apiKeysMutex.RLock()
	for sessionID, keys := range s.apiKeys {
		for _, apiKey := range keys {
			if apiKey == nil || !apiKey.Matches(normalizedKey) {
				continue
			}
			if apiKey.IsRevoked() {
				revokedMatch = true
				break
			}
			matchedSessionID = sessionID
			matchedKey = apiKey
			break
		}
		if matchedKey != nil || revokedMatch {
			break
		}
	}
	s.apiKeysMutex.RUnlock()

	if revokedMatch {
		return nil, fmt.Errorf("api key is revoked")
	}
	if matchedKey == nil {
		return nil, fmt.Errorf("api key not found")
	}

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

// RefreshSessionToken refreshes a session token for extended validity
func (s *SessionsService) RefreshSessionToken(sessionID string) error {
	// In this implementation, we don't have expiring tokens, so this is a no-op
	// In a real implementation, this would extend the validity of a session token
	return nil
}

// InvalidateSession immediately invalidates a session
func (s *SessionsService) InvalidateSession(sessionID, reason string) error {
	// In this implementation, we don't have separate session tokens to invalidate
	// In a real implementation, this would invalidate any active tokens for the session
	return nil
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
