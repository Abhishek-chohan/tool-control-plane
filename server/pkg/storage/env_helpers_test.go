package storage

import (
	"testing"
	"time"
)

func TestIntEnv(t *testing.T) {
	cases := []struct {
		name     string
		value    string
		set      bool
		fallback int
		want     int
	}{
		{"unset", "", false, 25, 25},
		{"empty", "", true, 25, 25},
		{"whitespace only", "   ", true, 25, 25},
		{"non-numeric", "abc", true, 25, 25},
		{"zero", "0", true, 25, 25},
		{"negative", "-5", true, 25, 25},
		{"valid", "50", true, 25, 50},
		{"valid with whitespace", "  50  ", true, 25, 50},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := "TOOPLANE_TEST_INT_ENV"
			if tc.set {
				t.Setenv(key, tc.value)
			}
			got := intEnv(key, tc.fallback)
			if got != tc.want {
				t.Fatalf("intEnv(%q, %d) with value %q = %d, want %d", key, tc.fallback, tc.value, got, tc.want)
			}
		})
	}
}

func TestDurationEnv(t *testing.T) {
	cases := []struct {
		name     string
		value    string
		set      bool
		fallback time.Duration
		want     time.Duration
	}{
		{"unset", "", false, 5 * time.Minute, 5 * time.Minute},
		{"empty", "", true, 5 * time.Minute, 5 * time.Minute},
		{"whitespace only", "   ", true, 5 * time.Minute, 5 * time.Minute},
		{"invalid", "not-a-duration", true, 5 * time.Minute, 5 * time.Minute},
		{"zero", "0s", true, 5 * time.Minute, 5 * time.Minute},
		{"negative", "-5m", true, 5 * time.Minute, 5 * time.Minute},
		{"valid", "30s", true, 5 * time.Minute, 30 * time.Second},
		{"valid with whitespace", "  30s  ", true, 5 * time.Minute, 30 * time.Second},
		{"valid minutes", "10m", true, 5 * time.Minute, 10 * time.Minute},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := "TOOPLANE_TEST_DURATION_ENV"
			if tc.set {
				t.Setenv(key, tc.value)
			}
			got := durationEnv(key, tc.fallback)
			if got != tc.want {
				t.Fatalf("durationEnv(%q, %v) with value %q = %v, want %v", key, tc.fallback, tc.value, got, tc.want)
			}
		})
	}
}
