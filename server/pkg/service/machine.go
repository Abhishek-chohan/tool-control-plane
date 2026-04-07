package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
	"toolplane/pkg/trace"
)

type machineDrainRequestTracker interface {
	ActiveRequestsForMachine(sessionID, machineID string) int
}

type machineDrainState struct {
	done chan struct{}
	once sync.Once
}

func newMachineDrainState() *machineDrainState {
	return &machineDrainState{done: make(chan struct{})}
}

func (s *machineDrainState) complete() {
	if s == nil {
		return
	}
	s.once.Do(func() {
		close(s.done)
	})
}

// MachinesService handles machine registration and management
type MachinesService struct {
	// In-memory storage for machines
	machines      map[string]map[string]*model.Machine // map[sessionID]map[machineID]Machine
	machinesMutex sync.RWMutex
	drainMutex    sync.RWMutex

	machineCapacity  map[string]map[string]int
	machineInFlight  map[string]map[string]int
	capacityMutex    sync.RWMutex
	drainingMachines map[string]map[string]*machineDrainState
	requestTracker   machineDrainRequestTracker

	// Tool service reference for registering tools from machines
	toolService *ToolService

	tracer trace.SessionTracer
	store  *storage.Store
}

// NewMachinesService creates a new machines service
func NewMachinesService(toolService *ToolService, tracer trace.SessionTracer, store *storage.Store) *MachinesService {
	if tracer == nil {
		tracer = trace.NopTracer()
	}

	service := &MachinesService{
		machines:         make(map[string]map[string]*model.Machine),
		machineCapacity:  make(map[string]map[string]int),
		machineInFlight:  make(map[string]map[string]int),
		drainingMachines: make(map[string]map[string]*machineDrainState),
		toolService:      toolService,
		tracer:           tracer,
		store:            store,
	}

	if store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if machines, err := store.AllMachines(ctx); err != nil {
			log.Printf("machine persistence load failed: %v", err)
		} else {
			for _, machine := range machines {
				if _, ok := service.machines[machine.SessionID]; !ok {
					service.machines[machine.SessionID] = make(map[string]*model.Machine)
				}
				service.machines[machine.SessionID][machine.ID] = machine
				service.setMachineCapacity(machine.SessionID, machine.ID, maxMachineConcurrentRequests)
			}
		}
	}

	// Start cleanup goroutine for inactive machines
	go service.cleanupInactiveMachines()

	return service
}

// RegisterMachine registers a new machine or updates an existing one
func (s *MachinesService) RegisterMachine(
	sessionID, machineID, sdkVersion, sdkLanguage, ip string,
	tools []*model.Tool,
) (*model.Machine, error) {
	if machineID != "" && s.IsMachineDraining(sessionID, machineID) {
		return nil, fmt.Errorf("machine %s is draining", machineID)
	}

	s.machinesMutex.Lock()
	defer s.machinesMutex.Unlock()

	// Initialize machines map for this session if not exists
	if _, ok := s.machines[sessionID]; !ok {
		s.machines[sessionID] = make(map[string]*model.Machine)
	}

	// Create new machine or update existing one
	var machine *model.Machine
	id := machineID
	if id == "" {
		// Generate new ID if not provided
		machine = model.NewMachine(sessionID, "", sdkVersion, sdkLanguage, ip)
		id = machine.ID
	} else if existing, ok := s.machines[sessionID][id]; ok {
		// Update existing machine
		existing.SDKVersion = sdkVersion
		existing.SDKLanguage = sdkLanguage
		existing.IP = ip
		existing.UpdatePing()
		machine = existing
	} else {
		// Create new machine with provided ID
		machine = model.NewMachine(sessionID, id, sdkVersion, sdkLanguage, ip)
	}

	// Store machine
	s.machines[sessionID][id] = machine
	s.setMachineCapacity(sessionID, id, maxMachineConcurrentRequests)

	s.tracer.Record(trace.SessionEvent{
		SessionID: sessionID,
		MachineID: id,
		Event:     trace.EventMachineRegistered,
		Timestamp: machine.CreatedAt,
		Metadata: map[string]any{
			"sdkVersion":  sdkVersion,
			"sdkLanguage": sdkLanguage,
			"ip":          ip,
			"tools":       len(tools),
		},
	})

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if err := s.store.SaveMachine(ctx, machine); err != nil {
			log.Printf("persist machine save failed: %v", err)
		}
	}

	// Track active machine for tool ownership checks
	s.toolService.TrackMachine(sessionID, id)

	// Register tools provided by this machine
	if len(tools) > 0 {
		s.cleanupExistingTools(sessionID, id)
		for _, tool := range tools {
			if tool == nil {
				continue
			}
			registeredTool, err := s.toolService.RegisterTool(sessionID, id, tool.Name, tool.Description, tool.Schema, tool.Config, tool.Tags)
			if err != nil {
				if errors.Is(err, storage.ErrToolOwnershipConflict) {
					log.Printf("machine %s failed to claim tool %s due to active owner", id, tool.Name)
				} else {
					log.Printf("machine %s failed to register tool %s: %v", id, tool.Name, err)
				}
				continue
			}
			if registeredTool != nil {
				log.Printf("machine %s registered tool %s (%s)", id, registeredTool.Name, registeredTool.ID)
			}
		}
	}

	return machine, nil
}

// SetRequestTracker lets the machine service observe in-flight requests for drain handling.
func (s *MachinesService) SetRequestTracker(tracker machineDrainRequestTracker) {
	s.drainMutex.Lock()
	s.requestTracker = tracker
	s.drainMutex.Unlock()
}

// GetMachineByID gets a machine by ID
func (s *MachinesService) GetMachineByID(sessionID, machineID string) (*model.Machine, error) {
	s.machinesMutex.RLock()
	defer s.machinesMutex.RUnlock()

	// Check if session exists
	if _, ok := s.machines[sessionID]; !ok {
		return nil, fmt.Errorf("no machines found for session %s", sessionID)
	}

	// Get machine
	machine, ok := s.machines[sessionID][machineID]
	if !ok {
		return nil, fmt.Errorf("machine %s not found in session %s", machineID, sessionID)
	}

	return machine, nil
}

// DrainMachine blocks new work for a machine, waits for in-flight requests to finish, then unregisters it.
func (s *MachinesService) DrainMachine(ctx context.Context, sessionID, machineID string) error {
	state, started, exists := s.getOrCreateDrainState(sessionID, machineID)
	if !exists {
		return nil
	}

	if started {
		activeRequests := 0
		if s.requestTracker != nil {
			activeRequests = s.requestTracker.ActiveRequestsForMachine(sessionID, machineID)
		} else {
			activeRequests, _ = s.MachineLoadInfo(sessionID, machineID)
		}
		s.tracer.Record(trace.SessionEvent{
			SessionID: sessionID,
			MachineID: machineID,
			Event:     trace.EventMachineDrainStarted,
			Timestamp: time.Now(),
			Metadata: map[string]any{
				"activeRequests": activeRequests,
			},
		})
		s.cleanupExistingTools(sessionID, machineID)
		s.toolService.UntrackMachine(sessionID, machineID)
		go s.waitForDrainAndUnregister(sessionID, machineID, state)
	}

	select {
	case <-state.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// IsMachineDraining reports whether the machine is in graceful-drain mode.
func (s *MachinesService) IsMachineDraining(sessionID, machineID string) bool {
	s.drainMutex.RLock()
	defer s.drainMutex.RUnlock()
	if sessionDrains, ok := s.drainingMachines[sessionID]; ok {
		_, ok = sessionDrains[machineID]
		return ok
	}
	return false
}

// ListMachines lists all machines in a session
func (s *MachinesService) ListMachines(sessionID string) ([]*model.Machine, error) {
	s.machinesMutex.RLock()
	defer s.machinesMutex.RUnlock()

	if _, ok := s.machines[sessionID]; !ok {
		return []*model.Machine{}, nil
	}

	machines := make([]*model.Machine, 0, len(s.machines[sessionID]))
	for _, machine := range s.machines[sessionID] {
		machines = append(machines, machine)
	}

	return machines, nil
}

// UpdateMachinePing updates a machine's last ping time
func (s *MachinesService) UpdateMachinePing(sessionID, machineID string) (*model.Machine, error) {
	s.machinesMutex.Lock()
	defer s.machinesMutex.Unlock()

	if _, ok := s.machines[sessionID]; !ok {
		return nil, fmt.Errorf("no machines found for session %s", sessionID)
	}

	machine, ok := s.machines[sessionID][machineID]
	if !ok {
		return nil, fmt.Errorf("machine %s not found in session %s", machineID, sessionID)
	}

	machine.UpdatePing()
	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if err := s.store.SaveMachine(ctx, machine); err != nil {
			log.Printf("persist machine ping failed: %v", err)
		}
	}

	s.tracer.Record(trace.SessionEvent{
		SessionID: sessionID,
		MachineID: machineID,
		Event:     trace.EventMachinePingUpdate,
		Timestamp: machine.LastPingAt,
	})

	return machine, nil
}

// UnregisterMachine unregisters a machine
func (s *MachinesService) UnregisterMachine(sessionID, machineID string) error {
	s.machinesMutex.Lock()
	defer s.machinesMutex.Unlock()

	machinesForSession, ok := s.machines[sessionID]
	if !ok {
		s.finishMachineDrain(sessionID, machineID, nil)
		return nil
	}

	machine, ok := machinesForSession[machineID]
	if !ok {
		s.finishMachineDrain(sessionID, machineID, nil)
		return nil
	}

	s.toolService.DeleteToolsByMachine(sessionID, machineID)
	delete(machinesForSession, machineID)
	if len(machinesForSession) == 0 {
		delete(s.machines, sessionID)
	}
	s.clearMachineCapacity(sessionID, machineID)

	s.toolService.UntrackMachine(sessionID, machineID)
	s.tracer.Record(trace.SessionEvent{
		SessionID: sessionID,
		MachineID: machineID,
		Event:     trace.EventMachineUnregistered,
		Timestamp: time.Now(),
		Metadata: map[string]any{
			"sdkVersion":  machine.SDKVersion,
			"sdkLanguage": machine.SDKLanguage,
		},
	})

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		defer cancel()
		if err := s.store.DeleteMachine(ctx, machineID); err != nil {
			log.Printf("persist machine delete failed: %v", err)
		}
	}

	s.finishMachineDrain(sessionID, machineID, nil)

	return nil
}

// FindMachinesWithTool finds machines that have a specific tool
func (s *MachinesService) FindMachinesWithTool(sessionID, toolName string) ([]*model.Machine, error) {
	s.machinesMutex.RLock()
	defer s.machinesMutex.RUnlock()

	tool, err := s.toolService.GetToolByName(sessionID, toolName)
	if err != nil {
		return nil, fmt.Errorf("tool not found: %v", err)
	}

	if tool.MachineID == "" {
		return nil, fmt.Errorf("tool %s is not associated with any machine", toolName)
	}

	sessionMachines, ok := s.machines[sessionID]
	if !ok {
		return nil, fmt.Errorf("no machines found for session %s", sessionID)
	}

	machine, ok := sessionMachines[tool.MachineID]
	if !ok {
		return nil, fmt.Errorf("machine %s for tool %s not registered", tool.MachineID, toolName)
	}

	return []*model.Machine{machine}, nil
}

// cleanupInactiveMachines periodically cleans up inactive machines
func (s *MachinesService) cleanupInactiveMachines() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanupMachines()
	}
}

// cleanupMachines removes machines that haven't pinged in a while
func (s *MachinesService) cleanupMachines() {
	cutoff := time.Now().Add(-machineHeartbeatTTL)
	staleCandidates := make(map[string]string)

	s.machinesMutex.RLock()
	for sessionID, machines := range s.machines {
		for machineID, machine := range machines {
			if machine.LastPingAt.Before(cutoff) {
				staleCandidates[machineID] = sessionID
			}
		}
	}
	s.machinesMutex.RUnlock()

	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		staleFromStore, err := s.store.ListStaleMachines(ctx, cutoff, 256)
		cancel()
		if err != nil {
			log.Printf("persist machine stale scan failed: %v", err)
		} else {
			for _, machine := range staleFromStore {
				staleCandidates[machine.ID] = machine.SessionID
			}
		}
	}

	if len(staleCandidates) == 0 {
		return
	}

	machinesRemoved := 0
	toolsDetached := 0
	for machineID, sessionID := range staleCandidates {
		removed, detached := s.reclaimMachine(sessionID, machineID, cutoff)
		if removed {
			machinesRemoved++
			toolsDetached += detached
		}
	}

	if machinesRemoved > 0 {
		fmt.Printf("Cleanup completed: %d machines reclaimed and %d tools detached due to inactivity\n", machinesRemoved, toolsDetached)
	}
}

func (s *MachinesService) cleanupExistingTools(sessionID, machineID string) {
	s.toolService.DeleteToolsByMachine(sessionID, machineID)
}

func (s *MachinesService) reclaimMachine(sessionID, machineID string, cutoff time.Time) (bool, int) {
	var machineSnapshot *model.Machine

	s.machinesMutex.RLock()
	if sessionID != "" {
		if machines, ok := s.machines[sessionID]; ok {
			machineSnapshot = machines[machineID]
		}
	}
	if machineSnapshot == nil {
		for sess, machines := range s.machines {
			if candidate, ok := machines[machineID]; ok {
				sessionID = sess
				machineSnapshot = candidate
				break
			}
		}
	}
	s.machinesMutex.RUnlock()

	detached := 0
	if s.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPersistenceTimeout)
		persistedSessionID, updates, removed, err := s.store.ReclaimMachine(ctx, machineID, cutoff)
		cancel()
		if err != nil {
			log.Printf("persist machine reclaim failed: %v", err)
			return false, 0
		}
		if !removed {
			return false, 0
		}
		if persistedSessionID != "" {
			sessionID = persistedSessionID
		}
		detached = len(updates)
		s.toolService.DetachToolsForMachine(updates, machineID)
	} else {
		if sessionID == "" {
			return false, 0
		}
		s.toolService.DeleteToolsByMachine(sessionID, machineID)
	}

	s.machinesMutex.Lock()
	if sessionID != "" {
		if machines, ok := s.machines[sessionID]; ok {
			delete(machines, machineID)
			if len(machines) == 0 {
				delete(s.machines, sessionID)
			}
		}
	}
	s.machinesMutex.Unlock()
	s.clearMachineCapacity(sessionID, machineID)

	s.toolService.UntrackMachine(sessionID, machineID)
	s.finishMachineDrain(sessionID, machineID, nil)

	if machineSnapshot != nil {
		s.tracer.Record(trace.SessionEvent{
			SessionID: sessionID,
			MachineID: machineID,
			Event:     trace.EventMachineInactivePruned,
			Timestamp: time.Now(),
			Metadata: map[string]any{
				"lastPingAt": machineSnapshot.LastPingAt,
			},
		})
	}

	fmt.Printf("Cleaned up inactive machine %s from session %s\n", machineID, sessionID)
	return true, detached
}

func (s *MachinesService) getOrCreateDrainState(sessionID, machineID string) (*machineDrainState, bool, bool) {
	if !s.machineExists(sessionID, machineID) {
		return nil, false, false
	}

	s.drainMutex.Lock()
	defer s.drainMutex.Unlock()

	if sessionDrains, ok := s.drainingMachines[sessionID]; ok {
		if state, ok := sessionDrains[machineID]; ok {
			return state, false, true
		}
	} else {
		s.drainingMachines[sessionID] = make(map[string]*machineDrainState)
	}

	state := newMachineDrainState()
	s.drainingMachines[sessionID][machineID] = state
	return state, true, true
}

func (s *MachinesService) waitForDrainAndUnregister(sessionID, machineID string, state *machineDrainState) {
	ticker := time.NewTicker(machineDrainPollInterval)
	defer ticker.Stop()

	for {
		if !s.machineExists(sessionID, machineID) {
			s.finishMachineDrain(sessionID, machineID, state)
			return
		}

		if !s.hasActiveRequestsForMachine(sessionID, machineID) {
			if err := s.UnregisterMachine(sessionID, machineID); err == nil {
				return
			} else {
				log.Printf("machine drain unregister failed for %s/%s: %v", sessionID, machineID, err)
			}
		}

		<-ticker.C
	}
}

func (s *MachinesService) hasActiveRequestsForMachine(sessionID, machineID string) bool {
	s.drainMutex.RLock()
	tracker := s.requestTracker
	s.drainMutex.RUnlock()
	if tracker != nil {
		return tracker.ActiveRequestsForMachine(sessionID, machineID) > 0
	}
	load, _ := s.MachineLoadInfo(sessionID, machineID)
	return load > 0
}

func (s *MachinesService) finishMachineDrain(sessionID, machineID string, state *machineDrainState) {
	s.drainMutex.Lock()
	if state == nil {
		if sessionDrains, ok := s.drainingMachines[sessionID]; ok {
			state = sessionDrains[machineID]
		}
	}
	if sessionDrains, ok := s.drainingMachines[sessionID]; ok {
		delete(sessionDrains, machineID)
		if len(sessionDrains) == 0 {
			delete(s.drainingMachines, sessionID)
		}
	}
	s.drainMutex.Unlock()
	if state != nil {
		s.tracer.Record(trace.SessionEvent{
			SessionID: sessionID,
			MachineID: machineID,
			Event:     trace.EventMachineDrainCompleted,
			Timestamp: time.Now(),
		})
		state.complete()
	}
}

func (s *MachinesService) machineExists(sessionID, machineID string) bool {
	s.machinesMutex.RLock()
	defer s.machinesMutex.RUnlock()
	if machines, ok := s.machines[sessionID]; ok {
		_, ok = machines[machineID]
		return ok
	}
	return false
}

// ReserveMachineSlot increments the active work counter when capacity remains.
func (s *MachinesService) ReserveMachineSlot(sessionID, machineID string) bool {
	s.machinesMutex.RLock()
	sessionMachines, ok := s.machines[sessionID]
	if !ok {
		s.machinesMutex.RUnlock()
		return false
	}
	if _, ok := sessionMachines[machineID]; !ok {
		s.machinesMutex.RUnlock()
		return false
	}
	s.machinesMutex.RUnlock()

	s.capacityMutex.Lock()
	s.ensureCapacityLocked(sessionID, machineID)
	cap := s.machineCapacity[sessionID][machineID]
	if cap <= 0 {
		cap = maxMachineConcurrentRequests
		s.machineCapacity[sessionID][machineID] = cap
	}
	current := s.machineInFlight[sessionID][machineID]
	if current >= cap {
		s.capacityMutex.Unlock()
		return false
	}
	s.machineInFlight[sessionID][machineID] = current + 1
	s.capacityMutex.Unlock()
	return true
}

// ReleaseMachineSlot decrements the active work counter.
func (s *MachinesService) ReleaseMachineSlot(sessionID, machineID string) {
	s.capacityMutex.Lock()
	if loads, ok := s.machineInFlight[sessionID]; ok {
		if current, ok := loads[machineID]; ok && current > 0 {
			loads[machineID] = current - 1
		}
		if len(loads) == 0 {
			delete(s.machineInFlight, sessionID)
		}
	}
	s.capacityMutex.Unlock()
}

func (s *MachinesService) ensureCapacityLocked(sessionID, machineID string) {
	if _, ok := s.machineCapacity[sessionID]; !ok {
		s.machineCapacity[sessionID] = make(map[string]int)
	}
	if _, ok := s.machineInFlight[sessionID]; !ok {
		s.machineInFlight[sessionID] = make(map[string]int)
	}
	if _, ok := s.machineCapacity[sessionID][machineID]; !ok {
		s.machineCapacity[sessionID][machineID] = maxMachineConcurrentRequests
	}
	if _, ok := s.machineInFlight[sessionID][machineID]; !ok {
		s.machineInFlight[sessionID][machineID] = 0
	}
}

func (s *MachinesService) setMachineCapacity(sessionID, machineID string, capacity int) {
	if sessionID == "" || machineID == "" {
		return
	}
	if capacity <= 0 {
		capacity = maxMachineConcurrentRequests
	}
	s.capacityMutex.Lock()
	s.ensureCapacityLocked(sessionID, machineID)
	s.machineCapacity[sessionID][machineID] = capacity
	s.capacityMutex.Unlock()
}

func (s *MachinesService) clearMachineCapacity(sessionID, machineID string) {
	s.capacityMutex.Lock()
	if caps, ok := s.machineCapacity[sessionID]; ok {
		delete(caps, machineID)
		if len(caps) == 0 {
			delete(s.machineCapacity, sessionID)
		}
	}
	if loads, ok := s.machineInFlight[sessionID]; ok {
		delete(loads, machineID)
		if len(loads) == 0 {
			delete(s.machineInFlight, sessionID)
		}
	}
	s.capacityMutex.Unlock()
}

// OrderMachinesByLoad sorts machines by current load to support scheduling.
func (s *MachinesService) OrderMachinesByLoad(sessionID string, machines []*model.Machine) []*model.Machine {
	if len(machines) <= 1 {
		return machines
	}
	s.capacityMutex.Lock()
	type metric struct {
		machine *model.Machine
		load    int
		cap     int
	}
	metrics := make([]metric, len(machines))
	for i, m := range machines {
		s.ensureCapacityLocked(sessionID, m.ID)
		metrics[i] = metric{
			machine: m,
			load:    s.machineInFlight[sessionID][m.ID],
			cap:     s.machineCapacity[sessionID][m.ID],
		}
		if metrics[i].cap <= 0 {
			metrics[i].cap = maxMachineConcurrentRequests
			s.machineCapacity[sessionID][m.ID] = metrics[i].cap
		}
	}
	s.capacityMutex.Unlock()

	sort.Slice(metrics, func(i, j int) bool {
		left := metrics[i].load * metrics[j].cap
		right := metrics[j].load * metrics[i].cap
		if left == right {
			return metrics[i].machine.CreatedAt.Before(metrics[j].machine.CreatedAt)
		}
		return left < right
	})

	ordered := make([]*model.Machine, len(metrics))
	for i, m := range metrics {
		ordered[i] = m.machine
	}
	return ordered
}

// MachineLoadInfo exposes current in-flight count and capacity.
func (s *MachinesService) MachineLoadInfo(sessionID, machineID string) (int, int) {
	s.capacityMutex.RLock()
	defer s.capacityMutex.RUnlock()
	load := 0
	cap := maxMachineConcurrentRequests
	if loads, ok := s.machineInFlight[sessionID]; ok {
		if current, ok := loads[machineID]; ok {
			load = current
		}
	}
	if caps, ok := s.machineCapacity[sessionID]; ok {
		if current, ok := caps[machineID]; ok && current > 0 {
			cap = current
		}
	}
	return load, cap
}

// MachineMetricsSnapshot returns current machine counts for observability scrapes.
func (s *MachinesService) MachineMetricsSnapshot() (active, draining, inflight int) {
	s.machinesMutex.RLock()
	for _, machines := range s.machines {
		active += len(machines)
	}
	s.machinesMutex.RUnlock()

	s.drainMutex.RLock()
	for _, machineDrains := range s.drainingMachines {
		draining += len(machineDrains)
	}
	s.drainMutex.RUnlock()

	s.capacityMutex.RLock()
	for _, loads := range s.machineInFlight {
		for _, load := range loads {
			inflight += load
		}
	}
	s.capacityMutex.RUnlock()

	return active, draining, inflight
}
