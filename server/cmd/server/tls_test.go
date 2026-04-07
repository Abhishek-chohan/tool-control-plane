package main

import "testing"

func TestValidateGRPCTLSSettingsRejectsPartialTLSConfiguration(t *testing.T) {
	if err := validateGRPCTLSSettings("development", "/tmp/server.crt", ""); err == nil {
		t.Fatal("expected partial TLS configuration to fail")
	}
}

func TestValidateGRPCTLSSettingsAllowsPlaintextOutsideProduction(t *testing.T) {
	if err := validateGRPCTLSSettings("development", "", ""); err != nil {
		t.Fatalf("expected empty TLS configuration to allow plaintext outside production: %v", err)
	}
}

func TestValidateGRPCTLSSettingsRejectsPlaintextInProduction(t *testing.T) {
	if err := validateGRPCTLSSettings("production", "", ""); err == nil {
		t.Fatal("expected empty TLS configuration to fail in production")
	}
}
