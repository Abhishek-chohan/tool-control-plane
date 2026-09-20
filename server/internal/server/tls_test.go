package server

import (
	"strings"
	"testing"
)

func TestValidateGRPCTLSSettingsRejectsPartialTLSConfiguration(t *testing.T) {
	if err := validateGRPCTLSSettings("development", "/tmp/server.crt", "", false); err == nil {
		t.Fatal("expected partial TLS configuration to fail")
	}
}

func TestValidateGRPCTLSSettingsAllowsPlaintextOutsideProduction(t *testing.T) {
	if err := validateGRPCTLSSettings("development", "", "", false); err != nil {
		t.Fatalf("expected empty TLS configuration to allow plaintext outside production: %v", err)
	}
}

func TestValidateGRPCTLSSettingsRejectsPlaintextInProduction(t *testing.T) {
	err := validateGRPCTLSSettings("production", "", "", false)
	if err == nil {
		t.Fatal("expected empty TLS configuration to fail in production")
	}
	if !strings.Contains(err.Error(), "TOOLPLANE_SERVER_TRUSTED_TRANSPORT") {
		t.Fatalf("expected the error to name the trusted-transport declaration, got: %v", err)
	}
}

func TestValidateGRPCTLSSettingsAllowsPlaintextInProductionWithTrustedTransport(t *testing.T) {
	if err := validateGRPCTLSSettings("production", "", "", true); err != nil {
		t.Fatalf("expected declared upstream TLS terminator to allow plaintext in production: %v", err)
	}
}

func TestTrustedTransportDeclarationFollowsEnvironment(t *testing.T) {
	t.Setenv("TOOLPLANE_SERVER_TRUSTED_TRANSPORT", "1")
	if !trustedTransportDeclared() {
		t.Fatal("expected TOOLPLANE_SERVER_TRUSTED_TRANSPORT=1 to declare")
	}
	t.Setenv("TOOLPLANE_SERVER_TRUSTED_TRANSPORT", "")
	if trustedTransportDeclared() {
		t.Fatal("expected an unset declaration to be off")
	}
}
