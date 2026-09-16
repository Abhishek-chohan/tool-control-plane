package service

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
	"toolplane/pkg/trace"
)

// ToolService implements the tool service logic
type ToolService struct {
	// In-memory storage for tools, in a real implementation this would use a database
	tools          map[string]map[string]*model.Tool // map[sessionID]map[toolID]Tool
	activeMachines map[string]map[string]struct{}
	toolsMutex     sync.RWMutex
	tracer         trace.SessionTracer
	store          storage.Storer
}

// NewToolService creates a new tool service
func NewToolService(tracer trace.SessionTracer, store storage.Storer) *ToolService {
	if tracer == nil {
		tracer = trace.NopTracer()
	}

	svc := &ToolService{
		tools:          make(map[string]map[string]*model.Tool),
		activeMachines: make(map[string]map[string]struct{}),
		tracer:         tracer,
		store:          store,
	}
	if store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), startupLoadTimeout)
		defer cancel()
		if tools, err := store.AllTools(ctx); err != nil {
			log.Printf("tool persistence load failed: %v", err)
		} else {
			for _, tool := range tools {
				if _, ok := svc.tools[tool.SessionID]; !ok {
					svc.tools[tool.SessionID] = make(map[string]*model.Tool)
				}
				// Ensure Config/Tags maps exist to avoid nil map usage later on.
				if tool.Config == nil {
					tool.Config = make(map[string]interface{})
				}
				if tool.Tags == nil {
					tool.Tags = []string{}
				}
				svc.tools[tool.SessionID][tool.ID] = tool
			}
		}
	}
	return svc
}

// RegisterTool registers a new tool
func (s *ToolService) RegisterTool(sessionID, machineID, name, description, schema string, config map[string]interface{}, tags []string) (*model.Tool, error) {
	log.Printf("ToolService.RegisterTool called with sessionID: %s, name: %s, tags: %v", sessionID, name, tags)
	s.toolsMutex.Lock()
	defer s.toolsMutex.Unlock()

	if _, ok := s.tools[sessionID]; !ok {
		s.tools[sessionID] = make(map[string]*model.Tool)
	}

	if s.store == nil {
		return s.registerToolInMemoryLocked(sessionID, machineID, name, description, schema, config, tags)
	}

	candidate := model.NewTool(sessionID, machineID, name, description, schema, cloneConfig(config), cloneTags(tags))
	ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
	defer cancel()

	updated, replacedMachineID, err := s.store.ClaimToolOwnership(ctx, candidate, time.Now().Add(-machineHeartbeatTTL))
	if err != nil {
		if errors.Is(err, storage.ErrToolOwnershipConflict) {
			existing := s.findToolByName(sessionID, name)
			toolID := ""
			ownerMachine := ""
			if existing != nil {
				toolID = existing.ID
				ownerMachine = existing.MachineID
			}
			s.tracer.Record(trace.SessionEvent{
				SessionID: sessionID,
				MachineID: machineID,
				ToolID:    toolID,
				Event:     trace.EventToolRegistrationRejected,
				Timestamp: time.Now(),
				Metadata: map[string]any{
					"name":           name,
					"ownerMachineId": ownerMachine,
				},
			})
			// One sentinel for both storage modes, carrying the owning
			// machine: the handler funnels this to FAILED_PRECONDITION
			// through statusFromDomainError.
			if ownerMachine == "" {
				return existing, wrapf(storage.ErrToolOwnershipConflict,
					"tool %s is owned by an active machine", name)
			}
			return existing, wrapf(storage.ErrToolOwnershipConflict,
				"tool %s is owned by active machine %s", name, ownerMachine)
		}
		return nil, err
	}

	sessionTools := s.tools[sessionID]
	_, existed := sessionTools[updated.ID]
	sessionTools[updated.ID] = updated

	// Remove stale duplicates by name if any linger in memory
	for id, tool := range sessionTools {
		if id != updated.ID && tool != nil && tool.Name == updated.Name {
			delete(sessionTools, id)
		}
	}

	metadata := map[string]any{
		"name": updated.Name,
	}
	if replacedMachineID != "" && replacedMachineID != machineID {
		metadata["reclaimedFrom"] = replacedMachineID
	}

	event := trace.EventToolRegistered
	timestamp := updated.CreatedAt
	if existed {
		event = trace.EventToolRefreshed
		timestamp = time.Now()
	}

	s.tracer.Record(trace.SessionEvent{
		SessionID: sessionID,
		MachineID: machineID,
		ToolID:    updated.ID,
		Event:     event,
		Timestamp: timestamp,
		Metadata:  metadata,
	})

	return updated, nil
}

func (s *ToolService) registerToolInMemoryLocked(sessionID, machineID, name, description, schema string, config map[string]interface{}, tags []string) (*model.Tool, error) {
	existingTool := s.findToolByName(sessionID, name)
	if existingTool != nil {
		if existingTool.MachineID != "" && existingTool.MachineID != machineID {
			if s.machineActiveLocked(sessionID, existingTool.MachineID) {
				s.tracer.Record(trace.SessionEvent{
					SessionID: sessionID,
					MachineID: machineID,
					ToolID:    existingTool.ID,
					Event:     trace.EventToolRegistrationRejected,
					Timestamp: time.Now(),
					Metadata: map[string]any{
						"name":           name,
						"ownerMachineId": existingTool.MachineID,
					},
				})
				return existingTool, wrapf(storage.ErrToolOwnershipConflict,
					"tool %s is owned by active machine %s", name, existingTool.MachineID)
			}
		}

		existingTool.MachineID = machineID
		existingTool.Description = description
		existingTool.Schema = schema
		cfg := cloneConfig(config)
		if cfg == nil {
			cfg = make(map[string]interface{})
		}
		existingTool.Config = cfg
		tg := cloneTags(tags)
		if tg == nil {
			tg = []string{}
		}
		existingTool.Tags = tg
		existingTool.UpdatePing()

		s.tracer.Record(trace.SessionEvent{
			SessionID: sessionID,
			MachineID: machineID,
			ToolID:    existingTool.ID,
			Event:     trace.EventToolRefreshed,
			Timestamp: time.Now(),
			Metadata: map[string]any{
				"name": name,
			},
		})
		return existingTool, nil
	}

	tool := model.NewTool(sessionID, machineID, name, description, schema, cloneConfig(config), cloneTags(tags))
	s.tools[sessionID][tool.ID] = tool

	s.tracer.Record(trace.SessionEvent{
		SessionID: sessionID,
		MachineID: machineID,
		ToolID:    tool.ID,
		Event:     trace.EventToolRegistered,
		Timestamp: tool.CreatedAt,
		Metadata: map[string]any{
			"name": tool.Name,
		},
	})

	return tool, nil
}

func (s *ToolService) DeleteToolsByMachine(sessionID, machineID string) {
	if machineID == "" {
		return
	}

	s.toolsMutex.Lock()
	defer s.toolsMutex.Unlock()

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		deleted, err := s.store.DeleteToolsByMachine(ctx, machineID)
		if err != nil {
			log.Printf("persist tool bulk delete failed: %v", err)
		} else {
			for _, upd := range deleted {
				if upd.SessionID == "" || upd.ToolID == "" {
					continue
				}
				if sessionTools, ok := s.tools[upd.SessionID]; ok {
					delete(sessionTools, upd.ToolID)
					if len(sessionTools) == 0 {
						delete(s.tools, upd.SessionID)
					}
				}
			}
		}
	}

	if sessionID == "" {
		return
	}
	if sessionTools, ok := s.tools[sessionID]; ok {
		for id, tool := range sessionTools {
			if tool != nil && tool.MachineID == machineID {
				delete(sessionTools, id)
			}
		}
		if len(sessionTools) == 0 {
			delete(s.tools, sessionID)
		}
	}
}

func (s *ToolService) DetachToolsForMachine(updates []storage.ToolOwnershipUpdate, machineID string) {
	if len(updates) == 0 {
		return
	}

	s.toolsMutex.Lock()
	defer s.toolsMutex.Unlock()

	for _, upd := range updates {
		if upd.SessionID == "" || upd.ToolID == "" {
			continue
		}
		sessionTools, ok := s.tools[upd.SessionID]
		if !ok {
			continue
		}
		tool, ok := sessionTools[upd.ToolID]
		if !ok || tool == nil {
			continue
		}
		if machineID == "" || tool.MachineID == machineID {
			tool.MachineID = ""
			tool.UpdatePing()
		}
	}
}

// TrackMachine marks a machine as active for ownership checks.
func (s *ToolService) TrackMachine(sessionID, machineID string) {
	if sessionID == "" || machineID == "" {
		return
	}

	s.toolsMutex.Lock()
	defer s.toolsMutex.Unlock()

	if _, ok := s.activeMachines[sessionID]; !ok {
		s.activeMachines[sessionID] = make(map[string]struct{})
	}
	s.activeMachines[sessionID][machineID] = struct{}{}
}

// UntrackMachine marks a machine as inactive for ownership checks.
func (s *ToolService) UntrackMachine(sessionID, machineID string) {
	if sessionID == "" || machineID == "" {
		return
	}

	s.toolsMutex.Lock()
	defer s.toolsMutex.Unlock()

	if sessionMachines, ok := s.activeMachines[sessionID]; ok {
		delete(sessionMachines, machineID)
		if len(sessionMachines) == 0 {
			delete(s.activeMachines, sessionID)
		}
	}
}

func (s *ToolService) machineActiveLocked(sessionID, machineID string) bool {
	if machineID == "" {
		return false
	}

	sessionMachines, ok := s.activeMachines[sessionID]
	if !ok {
		return false
	}

	_, ok = sessionMachines[machineID]
	return ok
}

// GetToolByID gets a tool by ID
func (s *ToolService) GetToolByID(sessionID, toolID string) (*model.Tool, error) {
	s.toolsMutex.RLock()
	defer s.toolsMutex.RUnlock()

	// Check if session exists
	if _, ok := s.tools[sessionID]; !ok {
		return nil, wrapf(ErrNotFound, "session %s not found", sessionID)
	}

	// Find tool by ID
	tool, ok := s.tools[sessionID][toolID]
	if !ok {
		return nil, wrapf(ErrNotFound, "tool %s not found in session %s", toolID, sessionID)
	}

	return tool, nil
}

// GetToolByName gets a tool by name in a specific session
// GetToolByName gets a tool by name in a specific session. Store-first when
// a store is configured: a registration made on another replica is resolved
// here instead of missing, and the local registry mirrors the row so
// claim-time ownership sees it too. Store errors degrade to the cache.
func (s *ToolService) GetToolByName(sessionID, name string) (*model.Tool, error) {
	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		tool, err := s.store.GetToolByName(ctx, sessionID, name)
		cancel()
		if err != nil {
			// Only a store error degrades to the cache. A successful miss is
			// authoritative: the row is gone (cross-replica delete), so the
			// stale local mirror is dropped and NOT_FOUND is returned rather
			// than silently serving the deleted tool.
			log.Printf("tool read-through failed: %v", err)
		} else {
			s.toolsMutex.Lock()
			if sessionTools, ok := s.tools[sessionID]; ok {
				for id, t := range sessionTools {
					if t != nil && t.Name == name {
						delete(sessionTools, id)
					}
				}
			}
			s.toolsMutex.Unlock()
			if tool == nil {
				return nil, wrapf(ErrNotFound, "tool %s not found in session %s", name, sessionID)
			}
			s.cacheTool(sessionID, tool)
			return tool, nil
		}
	}

	s.toolsMutex.RLock()
	defer s.toolsMutex.RUnlock()

	// Find tool by name
	tool := s.findToolByName(sessionID, name)
	if tool == nil {
		return nil, wrapf(ErrNotFound, "tool %s not found in session %s", name, sessionID)
	}

	return tool, nil
}

// cacheTool mirrors a store-read tool into the local registry so subsequent
// lookups (claim-time ownership, listings) resolve locally.
func (s *ToolService) cacheTool(sessionID string, tool *model.Tool) {
	s.toolsMutex.Lock()
	defer s.toolsMutex.Unlock()
	if _, ok := s.tools[sessionID]; !ok {
		s.tools[sessionID] = make(map[string]*model.Tool)
	}
	s.tools[sessionID][tool.ID] = tool
}

// findToolByName is an internal helper to find a tool by name (assumes lock is held)
func (s *ToolService) findToolByName(sessionID, name string) *model.Tool {
	// Check if session exists
	if _, ok := s.tools[sessionID]; !ok {
		return nil
	}

	// Find tool by name
	for _, tool := range s.tools[sessionID] {
		if tool.Name == name {
			return tool
		}
	}

	return nil
}

// ToolNamesForMachine returns the names of the session's tools the machine
// registered. Claim-time ownership is derived from this registry, never from
// caller-supplied filters alone. An empty machine identity yields nothing.
func (s *ToolService) ToolNamesForMachine(sessionID, machineID string) []string {
	if machineID == "" {
		return nil
	}
	s.toolsMutex.RLock()
	defer s.toolsMutex.RUnlock()

	var names []string
	for _, tool := range s.tools[sessionID] {
		if tool != nil && tool.MachineID == machineID {
			names = append(names, tool.Name)
		}
	}
	return names
}

// MachineProvidesTool reports whether the machine registered the named tool
// in the session. An empty machine identity never matches: a detached or
// malformed tool row with an empty MachineID must not become claimable by an
// unidentified caller.
func (s *ToolService) MachineProvidesTool(sessionID, machineID, toolName string) bool {
	if machineID == "" {
		return false
	}
	s.toolsMutex.RLock()
	defer s.toolsMutex.RUnlock()

	tool := s.findToolByName(sessionID, toolName)
	return tool != nil && tool.MachineID == machineID
}

// ListTools lists all tools in a session. Store-first when a store is
// configured so listings include registrations made on other replicas; every
// store row is mirrored into the local registry. Store errors degrade to the
// cache.
func (s *ToolService) ListTools(sessionID string) ([]*model.Tool, error) {
	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		tools, err := s.store.ListToolsBySession(ctx, sessionID)
		cancel()
		if err != nil {
			log.Printf("tool list read-through failed: %v", err)
		} else {
			s.toolsMutex.Lock()
			if _, ok := s.tools[sessionID]; !ok {
				s.tools[sessionID] = make(map[string]*model.Tool)
			}
			for _, tool := range tools {
				s.tools[sessionID][tool.ID] = tool
			}
			s.toolsMutex.Unlock()
			return tools, nil
		}
	}

	s.toolsMutex.RLock()
	defer s.toolsMutex.RUnlock()

	// Check if session exists
	if _, ok := s.tools[sessionID]; !ok {
		return []*model.Tool{}, nil
	}

	// Get all tools for this session
	tools := make([]*model.Tool, 0, len(s.tools[sessionID]))
	for _, tool := range s.tools[sessionID] {
		tools = append(tools, tool)
	}

	return tools, nil
}

// UpdateToolPing updates a tool's last ping time
func (s *ToolService) UpdateToolPing(sessionID, toolID string) (*model.Tool, error) {
	s.toolsMutex.Lock()
	defer s.toolsMutex.Unlock()

	// Check if session exists
	if _, ok := s.tools[sessionID]; !ok {
		return nil, wrapf(ErrNotFound, "session %s not found", sessionID)
	}

	// Find tool by ID
	tool, ok := s.tools[sessionID][toolID]
	if !ok {
		return nil, wrapf(ErrNotFound, "tool %s not found in session %s", toolID, sessionID)
	}

	// Update ping time
	tool.UpdatePing()

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if err := s.store.SaveTool(ctx, tool); err != nil {
			log.Printf("persist tool ping failed: %v", err)
		}
	}

	return tool, nil
}

// DeleteTool deletes a tool
func (s *ToolService) DeleteTool(sessionID, toolID string) error {
	s.toolsMutex.Lock()
	defer s.toolsMutex.Unlock()

	// Check if session exists
	if _, ok := s.tools[sessionID]; !ok {
		return wrapf(ErrNotFound, "session %s not found", sessionID)
	}

	// Check if tool exists
	tool, ok := s.tools[sessionID][toolID]
	if !ok {
		return wrapf(ErrNotFound, "tool %s not found in session %s", toolID, sessionID)
	}

	// Delete the tool
	delete(s.tools[sessionID], toolID)
	s.tracer.Record(trace.SessionEvent{
		SessionID: sessionID,
		MachineID: tool.MachineID,
		ToolID:    toolID,
		Event:     trace.EventToolDeleted,
		Timestamp: time.Now(),
		Metadata: map[string]any{
			"name": tool.Name,
		},
	})

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if err := s.store.DeleteTool(ctx, toolID); err != nil {
			log.Printf("persist tool delete failed: %v", err)
		}
	}

	return nil
}

// GetAllTools gets all tools across all sessions
func (s *ToolService) GetAllTools() ([]*model.Tool, error) {
	s.toolsMutex.RLock()
	defer s.toolsMutex.RUnlock()

	allTools := []*model.Tool{}

	// Iterate through all sessions
	for _, sessionTools := range s.tools {
		// Get all tools for this session
		for _, tool := range sessionTools {
			allTools = append(allTools, tool)
		}
	}

	return allTools, nil
}

func cloneConfig(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return nil
	}
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneTags(src []string) []string {
	if src == nil {
		return nil
	}
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

// GetToolsSummary gets a summary of all available tools with their schemas
func (s *ToolService) GetToolsSummary() ([]map[string]interface{}, error) {
	allTools, err := s.GetAllTools()
	if err != nil {
		return nil, err
	}

	// Format the tools summary
	summary := make([]map[string]interface{}, 0, len(allTools))
	for _, tool := range allTools {
		// Extract schemas
		inputSchema := tool.Schema
		outputSchema := ""
		if tool.Config != nil {
			if outputSchemaStr, ok := tool.Config["outputSchema"]; ok {
				// Add type assertion to convert interface{} to string
				if strValue, ok := outputSchemaStr.(string); ok {
					outputSchema = strValue
				}
				// If not a string, outputSchema remains empty
			}
		}

		toolSummary := map[string]interface{}{
			"id":           tool.ID,
			"name":         tool.Name,
			"description":  tool.Description,
			"sessionId":    tool.SessionID,
			"inputSchema":  inputSchema,
			"outputSchema": outputSchema,
			"createdAt":    tool.CreatedAt,
		}

		summary = append(summary, toolSummary)
	}

	return summary, nil
}

// GetOpenAITools gets tools in OpenAI compatible format
func (s *ToolService) GetOpenAITools(sessionID string) ([]map[string]interface{}, error) {
	tools, err := s.ListTools(sessionID)
	if err != nil {
		return nil, err
	}

	openAITools := make([]map[string]interface{}, 0, len(tools))
	for _, tool := range tools {
		openAITool := map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Schema,
			},
		}

		openAITools = append(openAITools, openAITool)
	}

	return openAITools, nil
}
