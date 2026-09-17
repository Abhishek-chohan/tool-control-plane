package model

import "time"

// AuditEvent is a durable record of a security- or lifecycle-relevant
// transition: session and API-key lifecycle, machine registration, and
// terminal dead-letter outcomes. Events land in the audit_events table and
// outlive the entities they describe.
type AuditEvent struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Event     string    `json:"event"`
	SessionID string    `json:"sessionId,omitempty"`
	MachineID string    `json:"machineId,omitempty"`
	RequestID string    `json:"requestId,omitempty"`
	TaskID    string    `json:"taskId,omitempty"`
	// ActorKeyID is the API key that performed the action, when the event
	// was triggered by an authenticated caller. System-driven events
	// (retention sweeps, dead-lettering after retry exhaustion) and events
	// predating attribution carry the empty value.
	ActorKeyID string         `json:"actorKeyId,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
}

// AuditEventFilter narrows a ListAuditEvents query. Empty fields match
// everything; Limit <= 0 uses the store's default page size.
type AuditEventFilter struct {
	SessionID  string
	ActorKeyID string
	Event      string
	Limit      int
	Offset     int
}
