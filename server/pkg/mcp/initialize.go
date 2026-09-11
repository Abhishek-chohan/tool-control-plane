package mcp

import (
	"encoding/json"
	"strings"
)

// initializeParams is the legacy initialize handshake's params object. The
// requested protocol version sits at the top level of params (unlike the
// 2026 revision, which carries it in _meta on every request).
type initializeParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ClientInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo"`
}

// handleInitialize serves the legacy initialize handshake. The gateway is
// stateless and holds no per-connection session, so there is nothing to
// negotiate beyond the version: it echoes the client's requested revision
// when supported, otherwise answers with DefaultLegacyProtocolVersion and
// lets the client decide (the legacy spec's server-side behavior).
func handleInitialize(params json.RawMessage) (any, *Error) {
	var req initializeParams
	if len(params) > 0 {
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, errInvalidParams("initialize params must be a JSON object: " + err.Error())
		}
	}

	requested := strings.TrimSpace(req.ProtocolVersion)
	chosen := DefaultLegacyProtocolVersion
	for _, version := range LegacyProtocolVersions {
		if requested == version {
			chosen = requested
			break
		}
	}

	return map[string]any{
		"protocolVersion": chosen,
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    ServerName,
			"version": ServerVersion,
		},
		"instructions": serverInstructions,
	}, nil
}
