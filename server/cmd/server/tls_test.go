package main

import "testing"

func TestValidateGRPCTLSSettingsRejectsPartialTLSConfiguration(t *testing.T) {
	if err := validateGRPCTLSSettings("development", "/tmp/server.crt", ""); err == nil {
		t.Fatal("expected partial TLS configuration to fail")
	}
}

func TestValidateGRPCTLSSettingsAllowsPlaintextWhenBothTLSFieldsAreEmpty(t *testing.T) {
	if err := validateGRPCTLSSettings("production", "", ""); err != nil {
		t.Fatalf("expected empty TLS configuration to allow plaintext: %v", err)
	}
}
