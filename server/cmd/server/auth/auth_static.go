package auth

import (
	"context"
	"crypto/subtle"
	"log"
)

// FixedAPIKeyValidator validates against one explicit non-secret fixture or local-development token.
type FixedAPIKeyValidator struct {
	expected string
	debug    bool
}

func NewFixedAPIKeyValidator(expected string, debug bool) *FixedAPIKeyValidator {
	return &FixedAPIKeyValidator{expected: normalizeToken(expected), debug: debug}
}

func (v *FixedAPIKeyValidator) Validate(_ context.Context, token string) bool {
	provided := normalizeToken(token)
	valid := subtle.ConstantTimeCompare([]byte(provided), []byte(v.expected)) == 1
	if v.debug {
		log.Printf("fixed API key validation result=%v token=%s", valid, redactToken(provided))
	}
	return valid
}
