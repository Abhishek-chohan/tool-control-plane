package model

import "time"

// AuditEvent is a durable record of a security- or lifecycle-relevant
// transition: session and API-key lifecycle, machine registration, and
// terminal dead-letter outcomes. Events land in the audit_events table and
// outlive the entities they describe.
type AuditEvent struct {
	ID        int64          `json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	Event     string         `json:"event"`
	SessionID string         `json:"sessionId,omitempty"`
	MachineID string         `json:"machineId,omitempty"`
	RequestID string         `json:"requestId,omitempty"`
	TaskID    string         `json:"taskId,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
}
