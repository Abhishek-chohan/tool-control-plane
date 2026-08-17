package mcp

import (
	"net/http"
	"strings"
)

// W3C Trace Context (SEP-414) propagation.
//
// The MCP 2026-07-28 revision carries distributed-tracing context in the
// standard W3C `traceparent` / `tracestate` HTTP headers. The gateway does not
// run an OpenTelemetry SDK; it validates the incoming headers and forwards them
// to the backend as gRPC metadata so the trace context flows end to end and any
// downstream tracer can continue the span. Invalid headers are dropped rather
// than propagated.

const (
	w3cTraceparentHeader = "Traceparent"
	w3cTracestateHeader  = "Tracestate"

	// gRPC metadata keys used to carry the W3C headers to the backend. They are
	// lowercase (gRPC metadata keys must be) and keep the W3C names so a
	// downstream extractor recognizes them.
	traceparentMetadataKey = "traceparent"
	tracestateMetadataKey  = "tracestate"
)

// traceparentFieldCount is version "-" trace-id "-" parent-id "-" flags.
const traceparentFieldCount = 4

// parseTraceparent validates a W3C `traceparent` header value and returns it
// normalized (trimmed) when valid, or "" when it must not be propagated.
//
// Valid form: version-traceid-parentid-flags where version is two hex digits
// (version "ff" is invalid), traceid is 32 lowercase-hex digits and not all
// zero, parentid is 16 lowercase-hex digits and not all zero, and flags is two
// hex digits. Unknown trailing fields (future versions) are tolerated.
func parseTraceparent(value string) string {
	header := strings.TrimSpace(value)
	if header == "" {
		return ""
	}
	fields := strings.Split(header, "-")
	if len(fields) < traceparentFieldCount {
		return ""
	}
	version, traceID, parentID, flags := fields[0], fields[1], fields[2], fields[3]

	// The version field is exactly two hex digits; "ff" (in any case) is
	// reserved/invalid per W3C Trace Context.
	if len(version) != 2 || !isHex(version) || strings.EqualFold(version, "ff") {
		return ""
	}
	// Version 00 must have exactly four fields; later versions may extend.
	if version == "00" && len(fields) != traceparentFieldCount {
		return ""
	}
	if len(traceID) != 32 || !isHex(traceID) || allZero(traceID) {
		return ""
	}
	if len(parentID) != 16 || !isHex(parentID) || allZero(parentID) {
		return ""
	}
	if len(flags) != 2 || !isHex(flags) {
		return ""
	}
	return header
}

// parseTracestate validates a W3C `tracestate` header value and returns it
// trimmed when plausible, or "" when it should not be propagated. The header is
// a comma-separated list of `key=value` list-members; we enforce the gross
// shape (non-empty members, an '=' in each) without parsing every vendor's
// value grammar.
func parseTracestate(value string) string {
	header := strings.TrimSpace(value)
	if header == "" {
		return ""
	}
	members := strings.Split(header, ",")
	if len(members) > 32 { // W3C caps the list at 32 members
		return ""
	}
	for _, member := range members {
		trimmed := strings.TrimSpace(member)
		if trimmed == "" {
			return ""
		}
		if !strings.Contains(trimmed, "=") {
			return ""
		}
	}
	return header
}

// traceMetadata extracts valid W3C trace-context headers from the request as
// gRPC metadata. Keys absent or invalid contribute nothing.
func traceMetadata(r *http.Request) map[string]string {
	out := map[string]string{}
	if tp := parseTraceparent(r.Header.Get(w3cTraceparentHeader)); tp != "" {
		out[traceparentMetadataKey] = tp
	}
	if ts := parseTracestate(r.Header.Get(w3cTracestateHeader)); ts != "" {
		// tracestate is only meaningful alongside a valid traceparent.
		if _, ok := out[traceparentMetadataKey]; ok {
			out[tracestateMetadataKey] = ts
		}
	}
	return out
}

func isHex(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}

func allZero(s string) bool {
	for _, c := range s {
		if c != '0' {
			return false
		}
	}
	return true
}
