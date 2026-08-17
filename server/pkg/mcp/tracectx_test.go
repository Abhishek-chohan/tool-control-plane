package mcp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseTraceparent(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "valid sampled",
			value: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
			want:  "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		},
		{
			name:  "valid not sampled with surrounding space",
			value: "  00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00  ",
			want:  "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00",
		},
		{name: "empty", value: "", want: ""},
		{name: "too few fields", value: "00-abc-def-01", want: ""},
		{name: "invalid version ff", value: "ff-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", want: ""},
		{name: "invalid version FF uppercase", value: "FF-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", want: ""},
		{name: "version single hex digit", value: "0-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", want: ""},
		{name: "version three hex digits", value: "000-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", want: ""},
		{name: "version 00 with extra field", value: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01-extra", want: ""},
		{name: "trace id all zero", value: "00-00000000000000000000000000000000-00f067aa0ba902b7-01", want: ""},
		{name: "parent id all zero", value: "00-4bf92f3577b34da6a3ce929d0e0e4736-0000000000000000-01", want: ""},
		{name: "trace id wrong length", value: "00-4bf92f3577b34da6-00f067aa0ba902b7-01", want: ""},
		{name: "parent id wrong length", value: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa-01", want: ""},
		{name: "non-hex trace id", value: "00-zbf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", want: ""},
		{name: "flags wrong length", value: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-1", want: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseTraceparent(tc.value); got != tc.want {
				t.Fatalf("parseTraceparent(%q) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}

func TestParseTracestate(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  string
	}{
		{name: "valid single member", value: "congo=t61rcWkgMzE", want: "congo=t61rcWkgMzE"},
		{name: "valid multiple members", value: "rojo=00f067aa0ba902b7,congo=t61rcWkgMzE", want: "rojo=00f067aa0ba902b7,congo=t61rcWkgMzE"},
		{name: "empty", value: "", want: ""},
		{name: "member without equals", value: "rojo", want: ""},
		{name: "empty member", value: "rojo=1,,congo=2", want: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseTracestate(tc.value); got != tc.want {
				t.Fatalf("parseTracestate(%q) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}

func TestTraceMetadataRequiresValidTraceparent(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	req.Header.Set("Tracestate", "rojo=00f067aa0ba902b7")

	md := traceMetadata(req)
	if md[traceparentMetadataKey] != "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01" {
		t.Fatalf("traceparent not propagated: %v", md)
	}
	if md[tracestateMetadataKey] != "rojo=00f067aa0ba902b7" {
		t.Fatalf("tracestate not propagated: %v", md)
	}
}

func TestTraceMetadataDropsTracestateWithoutTraceparent(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Tracestate", "rojo=00f067aa0ba902b7")

	if md := traceMetadata(req); len(md) != 0 {
		t.Fatalf("expected no metadata without a valid traceparent, got %v", md)
	}
}

func TestTraceMetadataDropsInvalidTraceparent(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Traceparent", "not-a-valid-header")
	req.Header.Set("Tracestate", "rojo=1")

	if md := traceMetadata(req); len(md) != 0 {
		t.Fatalf("expected invalid traceparent to be dropped, got %v", md)
	}
}
