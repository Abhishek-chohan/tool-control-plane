package model

import (
	"errors"
	"fmt"
	"strings"
)

type APIKeyCapability string

const (
	APIKeyCapabilityRead APIKeyCapability = "read"
	// APIKeyCapabilityInvoke authorizes consumer operations: creating
	// requests, invoking tools, cancelling work.
	APIKeyCapabilityInvoke APIKeyCapability = "invoke"
	// APIKeyCapabilityProvide authorizes provider operations: registering
	// machines/tools, claiming requests, submitting results and chunks,
	// renewing leases, draining.
	APIKeyCapabilityProvide APIKeyCapability = "provide"
	APIKeyCapabilityAdmin   APIKeyCapability = "admin"

	// APIKeyCapabilityExecute is the legacy pre-split value that combined
	// invoke and provide. It is still accepted everywhere capabilities are
	// normalized (existing keys and database rows keep working) and expands
	// to invoke+provide. Do not mint new keys with it.
	APIKeyCapabilityExecute APIKeyCapability = "execute"
)

var apiKeyCapabilityOrder = []APIKeyCapability{
	APIKeyCapabilityRead,
	APIKeyCapabilityInvoke,
	APIKeyCapabilityProvide,
	APIKeyCapabilityAdmin,
}

type AuthMode string

var ErrUnsupportedAPIKeyCapability = errors.New("unsupported api key capability")

// ErrAPIKeyCapabilitiesRequired is returned when a caller asks to mint an API
// key without stating any capabilities. Keys must be created least-privilege
// and explicit; there is no implicit full-access default.
var ErrAPIKeyCapabilitiesRequired = errors.New("api key capabilities are required")

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

// ParseAPIKeyCapabilitiesStrict validates an explicit capability list for
// minting new keys. Unlike NormalizeAPIKeyCapabilities it never falls back to
// the full-capability default: an empty (or entirely blank) list is
// ErrAPIKeyCapabilitiesRequired, so provisioning a key always states intent.
func ParseAPIKeyCapabilitiesStrict(values []string) ([]APIKeyCapability, error) {
	if len(values) == 0 {
		return nil, ErrAPIKeyCapabilitiesRequired
	}
	hasValue := false
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			hasValue = true
			break
		}
	}
	if !hasValue {
		return nil, ErrAPIKeyCapabilitiesRequired
	}
	return NormalizeAPIKeyCapabilities(values)
}

func NormalizeAPIKeyCapabilities(values []string) ([]APIKeyCapability, error) {
	if len(values) == 0 {
		return DefaultAPIKeyCapabilities(), nil
	}

	allowed := map[string][]APIKeyCapability{
		string(APIKeyCapabilityRead):    {APIKeyCapabilityRead},
		string(APIKeyCapabilityInvoke):  {APIKeyCapabilityInvoke},
		string(APIKeyCapabilityProvide): {APIKeyCapabilityProvide},
		string(APIKeyCapabilityAdmin):   {APIKeyCapabilityAdmin},
		// Legacy alias from before the invoke/provide split.
		string(APIKeyCapabilityExecute): {APIKeyCapabilityInvoke, APIKeyCapabilityProvide},
	}
	seen := make(map[APIKeyCapability]struct{})

	for _, value := range values {
		normalized := strings.TrimSpace(strings.ToLower(value))
		if normalized == "" {
			continue
		}
		expanded, ok := allowed[normalized]
		if !ok {
			return nil, fmt.Errorf("%w %q", ErrUnsupportedAPIKeyCapability, value)
		}
		for _, capability := range expanded {
			if _, ok := seen[capability]; ok {
				continue
			}
			seen[capability] = struct{}{}
		}
	}

	if len(seen) == 0 {
		return DefaultAPIKeyCapabilities(), nil
	}

	ordered := make([]APIKeyCapability, 0, len(seen))
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
