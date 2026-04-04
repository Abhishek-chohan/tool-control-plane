package model

import (
	"log"
	"time"

	"github.com/google/uuid"
)

// Tool represents a registered tool
type Tool struct {
	ID          string                 `json:"id"`
	MachineID   string                 `json:"machineId"`
	SessionID   string                 `json:"sessionId"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Schema      string                 `json:"schema"` // JSON schema stored as string
	Config      map[string]interface{} `json:"config,omitempty"`
	Tags        []string               `json:"tags,omitempty"` // Optional tags for categorization
	CreatedAt   time.Time              `json:"createdAt"`
	LastPingAt  time.Time              `json:"lastPingAt"`
}

// NewTool creates a new tool
func NewTool(sessionID, machineID, name, description, schema string, config map[string]interface{}, tags []string) *Tool {
	log.Printf("NewTool called with name: %s, tags: %v", name, tags)
	now := time.Now()
	if config == nil {
		config = make(map[string]interface{})
	}
	if tags == nil {
		tags = []string{}
	}

	tool := &Tool{
		ID:          uuid.New().String(),
		SessionID:   sessionID,
		MachineID:   machineID,
		Name:        name,
		Description: description,
		Schema:      schema,
		Config:      config,
		Tags:        tags,
		CreatedAt:   now,
		LastPingAt:  now,
	}

	log.Printf("NewTool created tool: %+v", tool)
	return tool
}

// UpdatePing updates the last ping time
func (t *Tool) UpdatePing() {
	t.LastPingAt = time.Now()
}

// GetOpenAITool returns this tool in OpenAI format
func (t *Tool) GetOpenAITool() map[string]interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        t.Name,
			"description": t.Description,
			"parameters":  t.Schema, // This will be parsed by the caller
		},
	}
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	RequestID  string
	Result     string
	ResultType string
}

// Session represents a logical grouping of tools and machines
type Session struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Namespace   string    `json:"namespace,omitempty"` // Optional namespace for organization
	CreatedAt   time.Time `json:"createdAt"`
	CreatedBy   string    `json:"createdBy"`        // User ID of creator
	ApiKey      string    `json:"apiKey,omitempty"` // optional api key to lock session
}

// NewSession creates a new session with generated ID and timestamp
func NewSession(name, description, createdBy, apiKey, namespace string) *Session {
	return &Session{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Namespace:   namespace,
		CreatedAt:   time.Now(),
		CreatedBy:   createdBy,
		ApiKey:      apiKey,
	}
}

// ApiKey represents an API key for authentication
type ApiKey struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Key       string     `json:"key"`
	SessionID string     `json:"sessionId"`
	CreatedAt time.Time  `json:"createdAt"`
	CreatedBy string     `json:"createdBy"` // User ID of creator
	RevokedAt *time.Time `json:"revokedAt,omitempty"`
}

// NewApiKey creates a new API key with generated ID, key, and timestamp
func NewApiKey(name, sessionID, createdBy string) *ApiKey {
	return &ApiKey{
		ID:        uuid.New().String(),
		Name:      name,
		Key:       generateApiKey(sessionID),
		SessionID: sessionID,
		CreatedAt: time.Now(),
		CreatedBy: createdBy,
	}
}

// generateApiKey generates a new API key
func generateApiKey(sessionID string) string {
	randomPart := uuid.New().String()
	return "toolplane_session_" + sessionID + "_" + randomPart
}

// IsRevoked checks if an API key is revoked
func (k *ApiKey) IsRevoked() bool {
	return k.RevokedAt != nil
}

// Revoke revokes the API key
func (k *ApiKey) Revoke() {
	now := time.Now()
	k.RevokedAt = &now
}

// Machine represents a registered machine that can execute tools
type Machine struct {
	ID          string    `json:"id"`
	SessionID   string    `json:"sessionId"`
	SDKVersion  string    `json:"sdkVersion"`
	SDKLanguage string    `json:"sdkLanguage"`
	IP          string    `json:"ip"`
	CreatedAt   time.Time `json:"createdAt"`
	LastPingAt  time.Time `json:"lastPingAt"`
}

// NewMachine creates a new machine with generated ID (if not provided) and timestamps
func NewMachine(sessionID, machineID, sdkVersion, sdkLanguage, ip string) *Machine {
	id := machineID
	if id == "" {
		id = uuid.New().String()
	}

	now := time.Now()
	return &Machine{
		ID:          id,
		SessionID:   sessionID,
		SDKVersion:  sdkVersion,
		SDKLanguage: sdkLanguage,
		IP:          ip,
		CreatedAt:   now,
		LastPingAt:  now,
	}
}

// UpdatePing updates the machine's last ping time
func (m *Machine) UpdatePing() {
	m.LastPingAt = time.Now()
}
