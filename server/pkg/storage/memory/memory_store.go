// Package memory provides an in-memory storage.Storer implementation used for
// development, local bootstrap, and CI fixtures where a Postgres instance is
// not provisioned. It mirrors the Postgres store's semantics so the service
// layer is storage-mode-agnostic and multi-instance ready: authoritative state
// lives behind the storage.Storer interface, not in service-layer maps.
//
// This implementation is safe for concurrent use (all access is guarded by a
// single sync.RWMutex) but, being process-local, it does NOT provide
// cross-process visibility. It is intended for the development/CI memory
// storage mode only and is gated out of production by the server config (see
// cmd/server/config.go, which forbids in-memory storage when
// TOOLPLANE_ENV_MODE=production).
package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
)

// Store is an in-memory storage.Storer.
type Store struct {
	mu sync.RWMutex

	requests     map[string]*model.Request      // by ID
	machines     map[string]*model.Machine      // by ID
	tasks        map[string]*model.Task         // by ID
	tools        map[string]*model.Tool         // by ID
	sessions     map[string]*model.Session      // by ID
	apiKeys      map[string]*model.ApiKey       // by ID
	userSess     map[string]map[string]struct{} // userID -> set(sessionID)
	machineTools map[string]map[string]struct{} // machineID -> set(toolID)
	draining     map[string]struct{}            // set(machineID)
}

// Compile-time check: the in-memory implementation satisfies storage.Storer.
var _ storage.Storer = (*Store)(nil)

// New returns an empty in-memory store.
func New() *Store {
	return &Store{
		requests:     make(map[string]*model.Request),
		machines:     make(map[string]*model.Machine),
		tasks:        make(map[string]*model.Task),
		tools:        make(map[string]*model.Tool),
		sessions:     make(map[string]*model.Session),
		apiKeys:      make(map[string]*model.ApiKey),
		userSess:     make(map[string]map[string]struct{}),
		machineTools: make(map[string]map[string]struct{}),
		draining:     make(map[string]struct{}),
	}
}

// Close is a no-op; present to mirror the Postgres store lifecycle.
func (s *Store) Close() error { return nil }

// ---------------- Requests ----------------

func (s *Store) AllRequests(ctx context.Context) ([]*model.Request, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.Request, 0, len(s.requests))
	for _, r := range s.requests {
		out = append(out, cloneRequest(r))
	}
	return out, nil
}

func (s *Store) GetRequestByIdempotencyKey(ctx context.Context, sessionID, idempotencyKey string) (*model.Request, error) {
	if idempotencyKey == "" {
		return nil, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.requests {
		if r.SessionID == sessionID && r.IdempotencyKey == idempotencyKey {
			return cloneRequest(r), nil
		}
	}
	return nil, nil
}

func (s *Store) SaveRequest(ctx context.Context, req *model.Request) error {
	if req == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requests[req.ID] = cloneRequest(req)
	return nil
}

func (s *Store) LeasePendingRequest(ctx context.Context, sessionID, machineID string, toolNames []string, leaseDuration time.Duration) (*model.Request, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	// Find oldest pending, non-dead-letter, visible request matching toolNames.
	var (
		pick   *model.Request
		pickAt time.Time
	)
	for _, r := range s.requests {
		if r.SessionID != sessionID || r.Status != model.RequestStatusPending || r.DeadLetter {
			continue
		}
		if !r.VisibleAt.IsZero() && r.VisibleAt.After(now) {
			continue
		}
		if !toolMatches(r.ToolName, toolNames) {
			continue
		}
		if pick == nil || r.CreatedAt.Before(pickAt) {
			pick = r
			pickAt = r.CreatedAt
		}
	}
	if pick == nil {
		return nil, nil
	}
	visible := now.Add(leaseDuration)
	pick.Status = model.RequestStatusClaimed
	pick.ExecutingMachineID = machineID
	pick.Attempts++
	pick.LeaseEpoch++
	pick.VisibleAt = visible
	pick.LeasedBy = machineID
	pick.LeasedAt = &now
	pick.NextAttemptAt = nil
	pick.LastError = ""
	pick.UpdatedAt = now
	return cloneRequest(pick), nil
}

func (s *Store) FindExpiredRequests(ctx context.Context, limit int) ([]*model.Request, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now()
	var out []*model.Request
	for _, r := range s.requests {
		if r.DeadLetter || r.LeasedAt == nil {
			continue
		}
		if r.LeaseExpired(now) || r.HasTimedOut(now) {
			out = append(out, cloneRequest(r))
		}
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (s *Store) GetRequest(ctx context.Context, requestID string) (*model.Request, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.requests[requestID]
	if !ok {
		return nil, nil
	}
	return cloneRequest(r), nil
}

func (s *Store) ListRequestsBySession(ctx context.Context, sessionID string) ([]*model.Request, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*model.Request
	for _, r := range s.requests {
		if r.SessionID == sessionID {
			out = append(out, cloneRequest(r))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (s *Store) ClaimRequest(ctx context.Context, sessionID, requestID, machineID string, leaseDuration time.Duration) (*model.Request, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.requests[requestID]
	if !ok {
		return nil, false, nil
	}
	if r.SessionID != sessionID || r.Status != model.RequestStatusPending || r.DeadLetter {
		return nil, false, nil
	}
	now := time.Now()
	visible := now.Add(leaseDuration)
	r.Status = model.RequestStatusClaimed
	r.ExecutingMachineID = machineID
	r.Attempts++
	r.LeaseEpoch++
	r.VisibleAt = visible
	r.LeasedBy = machineID
	r.LeasedAt = &now
	r.NextAttemptAt = nil
	r.LastError = ""
	r.UpdatedAt = now
	return cloneRequest(r), true, nil
}

func (s *Store) ReclaimExpiredRequest(ctx context.Context, requestID string, now time.Time, leaseDuration time.Duration, maxAttempts int, backoff time.Duration) (*model.Request, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.requests[requestID]
	if !ok {
		return nil, false, nil
	}
	if r.Status != model.RequestStatusClaimed && r.Status != model.RequestStatusRunning {
		return nil, false, nil
	}
	if !r.LeaseExpired(now) && !r.HasTimedOut(now) {
		return nil, false, nil
	}
	r.ExecutingMachineID = ""
	r.LeasedBy = ""
	r.LeasedAt = nil
	r.NextAttemptAt = nil
	r.UpdatedAt = now
	if r.Attempts >= maxAttempts {
		r.DeadLetter = true
		r.Status = model.RequestStatusFailed
		r.LastError = "request timed out: max attempts reached"
		r.Error = r.LastError
		r.VisibleAt = now
	} else {
		r.Status = model.RequestStatusPending
		r.Attempts++
		r.LastError = "request lease expired"
		retryAt := now.Add(backoff)
		r.VisibleAt = retryAt
		r.NextAttemptAt = &retryAt
	}
	return cloneRequest(r), true, nil
}

func (s *Store) MachineInFlightCount(ctx context.Context, sessionID, machineID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, r := range s.requests {
		if r.ExecutingMachineID == machineID && r.SessionID == sessionID && (r.Status == model.RequestStatusClaimed || r.Status == model.RequestStatusRunning) {
			count++
		}
	}
	return count, nil
}

// ---------------- Machines ----------------

func (s *Store) AllMachines(ctx context.Context) ([]*model.Machine, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.Machine, 0, len(s.machines))
	for _, m := range s.machines {
		out = append(out, cloneMachine(m))
	}
	return out, nil
}

func (s *Store) GetMachine(ctx context.Context, machineID string) (*model.Machine, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.machines[machineID]
	if !ok || m == nil {
		return nil, nil
	}
	return cloneMachine(m), nil
}

func (s *Store) SaveMachine(ctx context.Context, machine *model.Machine) error {
	if machine == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.machines[machine.ID] = cloneMachine(machine)
	return nil
}

func (s *Store) DeleteMachine(ctx context.Context, machineID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.machines, machineID)
	delete(s.machineTools, machineID)
	// Match the Postgres store, where the drain flag is a column on the
	// deleted row: removing the machine ends any drain state with it.
	delete(s.draining, machineID)
	return nil
}

func (s *Store) ListStaleMachines(ctx context.Context, cutoff time.Time, limit int) ([]*model.Machine, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*model.Machine
	for _, m := range s.machines {
		if m.LastPingAt.Before(cutoff) {
			out = append(out, cloneMachine(m))
		}
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (s *Store) ReclaimMachine(ctx context.Context, machineID string, cutoff time.Time) (string, []storage.ToolOwnershipUpdate, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.machines[machineID]
	if !ok {
		return "", nil, false, nil
	}
	if !m.LastPingAt.Before(cutoff) {
		return "", nil, false, nil
	}
	sessionID := m.SessionID
	var updates []storage.ToolOwnershipUpdate
	for _, t := range s.tools {
		if t.MachineID == machineID {
			t.MachineID = ""
			updates = append(updates, storage.ToolOwnershipUpdate{ToolID: t.ID, SessionID: t.SessionID})
		}
	}
	delete(s.machines, machineID)
	delete(s.machineTools, machineID)
	return sessionID, updates, true, nil
}

func (s *Store) SetMachineDraining(ctx context.Context, sessionID, machineID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Mirror the Postgres store's UPDATE ... WHERE id=$1 AND session_id=$2: only
	// mark draining if the machine exists in the given session, otherwise it's a
	// no-op. This prevents IsMachineDraining from blocking work for a
	// non-existent or wrong-session machine.
	if m, ok := s.machines[machineID]; ok && m.SessionID == sessionID {
		s.draining[machineID] = struct{}{}
	}
	return nil
}

func (s *Store) ClearMachineDraining(ctx context.Context, machineID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.draining, machineID)
	return nil
}

func (s *Store) IsMachineDraining(ctx context.Context, machineID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.draining[machineID]
	return ok, nil
}

// ---------------- Tasks ----------------

func (s *Store) AllTasks(ctx context.Context) ([]*model.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		out = append(out, cloneTask(t))
	}
	return out, nil
}

func (s *Store) GetTaskByIdempotencyKey(ctx context.Context, sessionID, idempotencyKey string) (*model.Task, error) {
	if idempotencyKey == "" {
		return nil, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.tasks {
		if t.SessionID == sessionID && t.IdempotencyKey == idempotencyKey {
			return cloneTask(t), nil
		}
	}
	return nil, nil
}

func (s *Store) SaveTask(ctx context.Context, task *model.Task) error {
	if task == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = cloneTask(task)
	return nil
}

func (s *Store) DeleteTask(ctx context.Context, taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, taskID)
	return nil
}

func (s *Store) FindNonTerminalTasks(ctx context.Context) ([]*model.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*model.Task
	for _, t := range s.tasks {
		if t.DeadLetter {
			continue
		}
		switch t.Status {
		case model.StatusCompleted, model.StatusFailed, model.StatusCancelled:
			continue
		}
		out = append(out, cloneTask(t))
	}
	return out, nil
}

func (s *Store) ClaimTaskForAdoption(ctx context.Context, taskID string, minAge time.Duration) (*model.Task, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[taskID]
	if !ok {
		return nil, false, nil
	}
	cutoff := time.Now().Add(-minAge)
	if t.UpdatedAt.After(cutoff) {
		return nil, false, nil // recently touched; don't claim
	}
	t.UpdatedAt = time.Now()
	return cloneTask(t), true, nil
}

// ---------------- Sessions ----------------

func (s *Store) AllSessions(ctx context.Context) ([]*model.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.Session, 0, len(s.sessions))
	for _, sess := range s.sessions {
		out = append(out, cloneSession(sess))
	}
	return out, nil
}

func (s *Store) GetSession(ctx context.Context, sessionID string) (*model.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[sessionID]
	if !ok {
		return nil, nil
	}
	return cloneSession(sess), nil
}

func (s *Store) InsertSessionIfAbsent(ctx context.Context, session *model.Session) (bool, error) {
	if session == nil {
		return false, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.sessions[session.ID]; exists {
		return false, nil
	}
	s.sessions[session.ID] = cloneSession(session)
	return true, nil
}

func (s *Store) SaveSession(ctx context.Context, session *model.Session) error {
	if session == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = cloneSession(session)
	return nil
}

func (s *Store) DeleteSession(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
	for _, set := range s.userSess {
		delete(set, sessionID)
	}
	for id, k := range s.apiKeys {
		if k.SessionID == sessionID {
			delete(s.apiKeys, id)
		}
	}
	return nil
}

func (s *Store) AllUserSessions(ctx context.Context) (map[string][]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string][]string, len(s.userSess))
	for u, set := range s.userSess {
		ids := make([]string, 0, len(set))
		for id := range set {
			ids = append(ids, id)
		}
		out[u] = ids
	}
	return out, nil
}

func (s *Store) AddUserSession(ctx context.Context, userID, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	set, ok := s.userSess[userID]
	if !ok {
		set = make(map[string]struct{})
		s.userSess[userID] = set
	}
	set[sessionID] = struct{}{}
	return nil
}

func (s *Store) RemoveUserSession(ctx context.Context, userID, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if set, ok := s.userSess[userID]; ok {
		delete(set, sessionID)
	}
	return nil
}

func (s *Store) AllApiKeys(ctx context.Context) ([]*model.ApiKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.ApiKey, 0, len(s.apiKeys))
	for _, k := range s.apiKeys {
		out = append(out, cloneApiKey(k))
	}
	return out, nil
}

func (s *Store) GetAPIKeyByHash(ctx context.Context, keyHash string) (*model.ApiKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, k := range s.apiKeys {
		if k.KeyHash == keyHash {
			return cloneApiKey(k), nil
		}
	}
	return nil, nil
}

func (s *Store) SaveApiKey(ctx context.Context, key *model.ApiKey) error {
	if key == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.apiKeys[key.ID] = cloneApiKey(key)
	return nil
}

// ---------------- Tools ----------------

func (s *Store) AllTools(ctx context.Context) ([]*model.Tool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.Tool, 0, len(s.tools))
	for _, t := range s.tools {
		out = append(out, cloneTool(t))
	}
	return out, nil
}

func (s *Store) SaveTool(ctx context.Context, tool *model.Tool) error {
	if tool == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[tool.ID] = cloneTool(tool)
	if tool.MachineID != "" {
		set, ok := s.machineTools[tool.MachineID]
		if !ok {
			set = make(map[string]struct{})
			s.machineTools[tool.MachineID] = set
		}
		set[tool.ID] = struct{}{}
	}
	return nil
}

func (s *Store) DeleteTool(ctx context.Context, toolID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.tools[toolID]; ok {
		if set, ok := s.machineTools[t.MachineID]; ok {
			delete(set, toolID)
		}
	}
	delete(s.tools, toolID)
	return nil
}

func (s *Store) DeleteToolsByMachine(ctx context.Context, machineID string) ([]storage.ToolOwnershipUpdate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var updates []storage.ToolOwnershipUpdate
	for _, t := range s.tools {
		if t.MachineID == machineID {
			updates = append(updates, storage.ToolOwnershipUpdate{ToolID: t.ID, SessionID: t.SessionID})
			delete(s.tools, t.ID)
		}
	}
	delete(s.machineTools, machineID)
	return updates, nil
}

func (s *Store) ClaimToolOwnership(ctx context.Context, tool *model.Tool, staleCutoff time.Time) (*model.Tool, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Find existing tool with the same name in the same session.
	var existing *model.Tool
	for _, t := range s.tools {
		if t.SessionID == tool.SessionID && t.Name == tool.Name {
			existing = t
			break
		}
	}
	if existing != nil {
		// Check if the owning machine is stale; if so, transfer ownership.
		if m, ok := s.machines[existing.MachineID]; ok && m.LastPingAt.Before(staleCutoff) {
			replaced := existing.MachineID
			existing.MachineID = tool.MachineID
			existing.Description = tool.Description
			existing.Schema = tool.Schema
			existing.Config = cloneConfig(tool.Config)
			existing.Tags = cloneTags(tool.Tags)
			existing.LastPingAt = tool.LastPingAt
			return cloneTool(existing), replaced, nil
		}
		return cloneTool(existing), "", storage.ErrToolOwnershipConflict
	}
	if tool.Config == nil {
		tool.Config = make(map[string]interface{})
	}
	if tool.Tags == nil {
		tool.Tags = []string{}
	}
	s.tools[tool.ID] = cloneTool(tool)
	if tool.MachineID != "" {
		set, ok := s.machineTools[tool.MachineID]
		if !ok {
			set = make(map[string]struct{})
			s.machineTools[tool.MachineID] = set
		}
		set[tool.ID] = struct{}{}
	}
	return cloneTool(tool), "", nil
}

// ---------------- helpers ----------------

func toolMatches(toolName string, toolNames []string) bool {
	if len(toolNames) == 0 {
		return true
	}
	for _, n := range toolNames {
		if n == toolName {
			return true
		}
	}
	return false
}

func cloneRequest(r *model.Request) *model.Request {
	if r == nil {
		return nil
	}
	c := *r
	if r.Meta != nil {
		c.Meta = make(map[string]string, len(r.Meta))
		for k, v := range r.Meta {
			c.Meta[k] = v
		}
	}
	if r.StreamResults != nil {
		c.StreamResults = append([]string(nil), r.StreamResults...)
	}
	if r.LeasedAt != nil {
		t := *r.LeasedAt
		c.LeasedAt = &t
	}
	if r.NextAttemptAt != nil {
		t := *r.NextAttemptAt
		c.NextAttemptAt = &t
	}
	return &c
}

func cloneMachine(m *model.Machine) *model.Machine {
	if m == nil {
		return nil
	}
	c := *m
	// Machine tokens are return-once credentials; stored/returned copies
	// never carry the plaintext.
	c.Token = ""
	return &c
}

func cloneTask(t *model.Task) *model.Task {
	if t == nil {
		return nil
	}
	c := *t
	if t.CompletedAt != nil {
		tt := *t.CompletedAt
		c.CompletedAt = &tt
	}
	if t.NextAttemptAt != nil {
		tt := *t.NextAttemptAt
		c.NextAttemptAt = &tt
	}
	return &c
}

func cloneSession(sess *model.Session) *model.Session {
	if sess == nil {
		return nil
	}
	c := *sess
	return &c
}

func cloneApiKey(k *model.ApiKey) *model.ApiKey {
	if k == nil {
		return nil
	}
	c := *k
	if k.Capabilities != nil {
		c.Capabilities = append([]model.APIKeyCapability(nil), k.Capabilities...)
	}
	if k.RevokedAt != nil {
		t := *k.RevokedAt
		c.RevokedAt = &t
	}
	return &c
}

func cloneTool(t *model.Tool) *model.Tool {
	if t == nil {
		return nil
	}
	c := *t
	c.Config = cloneConfig(t.Config)
	c.Tags = cloneTags(t.Tags)
	return &c
}

func cloneConfig(in map[string]interface{}) map[string]interface{} {
	if in == nil {
		return make(map[string]interface{})
	}
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneTags(in []string) []string {
	if in == nil {
		return []string{}
	}
	return append([]string(nil), in...)
}
