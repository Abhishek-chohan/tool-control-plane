package service

import "errors"

// ErrRequestTimeoutOutOfRange reports that a caller-supplied per-request
// timeout_seconds exceeded maxRequestTimeout.
var ErrRequestTimeoutOutOfRange = errors.New("service: request timeout out of range")
