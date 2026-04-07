package model

import (
	"errors"
	"fmt"
	"strings"
)

type APIKeyCapability string

const (
	APIKeyCapabilityRead    APIKeyCapability = "read"
	APIKeyCapabilityExecute APIKeyCapability = "execute"
	APIKeyCapabilityAdmin   APIKeyCapability = "admin"
)

var apiKeyCapabilityOrder = []APIKeyCapability{
	APIKeyCapabilityRead,
	APIKeyCapabilityExecute,
	APIKeyCapabilityAdmin,
}

type AuthMode string

var ErrUnsupportedAPIKeyCapability = errors.New("unsupported api key capability")

const (
	AuthModeFixed      AuthMode = "fixed"
	AuthModeSessionKey AuthMode = "session_key"
)

type AuthPrincipal struct {
	Mode         AuthMode
	SessionID    string
	UserID       string
	KeyID        string
	Capabilities []APIKeyCapability
	TokenPreview string
}

func DefaultAPIKeyCapabilities() []APIKeyCapability {
	capabilities := make([]APIKeyCapability, len(apiKeyCapabilityOrder))
	copy(capabilities, apiKeyCapabilityOrder)
	return capabilities
}

func NormalizeAPIKeyCapabilities(values []string) ([]APIKeyCapability, error) {
	if len(values) == 0 {
		return DefaultAPIKeyCapabilities(), nil
	}

	allowed := map[string]APIKeyCapability{
		string(APIKeyCapabilityRead):    APIKeyCapabilityRead,
		string(APIKeyCapabilityExecute): APIKeyCapabilityExecute,
		string(APIKeyCapabilityAdmin):   APIKeyCapabilityAdmin,
	}
	seen := make(map[APIKeyCapability]struct{})
	capabilities := make([]APIKeyCapability, 0, len(values))

	for _, value := range values {
		normalized := strings.TrimSpace(strings.ToLower(value))
		if normalized == "" {
			continue
		}
		capability, ok := allowed[normalized]
		if !ok {
			return nil, fmt.Errorf("%w %q", ErrUnsupportedAPIKeyCapability, value)
		}
		if _, ok := seen[capability]; ok {
			continue
		}
		seen[capability] = struct{}{}
		capabilities = append(capabilities, capability)
	}

	if len(capabilities) == 0 {
		return DefaultAPIKeyCapabilities(), nil
	}

	ordered := make([]APIKeyCapability, 0, len(capabilities))
	for _, capability := range apiKeyCapabilityOrder {
		if _, ok := seen[capability]; ok {
			ordered = append(ordered, capability)
		}
	}

	return ordered, nil
}

func CapabilityStrings(capabilities []APIKeyCapability) []string {
	values := make([]string, 0, len(capabilities))
	for _, capability := range capabilities {
		if capability == "" {
			continue
		}
		values = append(values, string(capability))
	}
	return values
}

func (p *AuthPrincipal) HasCapability(required APIKeyCapability) bool {
	if required == "" {
		return true
	}
	if p == nil {
		return false
	}
	if p.Mode == AuthModeFixed {
		return true
	}
	for _, capability := range p.Capabilities {
		if capability == required {
			return true
		}
	}
	return false
}
