package main

import "testing"

func TestLoadProxyConfigAllowsProductionCustomCA(t *testing.T) {
	t.Setenv("TOOLPLANE_ENV_MODE", "production")
	t.Setenv("TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND", "0")
	t.Setenv("TOOLPLANE_PROXY_ALLOWED_ORIGINS", "https://app.example.com")
	t.Setenv("TOOLPLANE_PROXY_BACKEND_TLS_SERVER_NAME", "server")
	t.Setenv("TOOLPLANE_PROXY_BACKEND_TLS_CA_FILE", "/certs/ca.crt")

	cfg, err := loadProxyConfig()
	if err != nil {
		t.Fatalf("expected production proxy config with custom CA to pass: %v", err)
	}
	if cfg.backendTLSCAFile != "/certs/ca.crt" {
		t.Fatalf("backend TLS CA file = %q, want /certs/ca.crt", cfg.backendTLSCAFile)
	}
}

func TestBackendTransportCredentialsRejectsMissingCAFile(t *testing.T) {
	_, err := backendTransportCredentials(proxyConfig{
		environment:          "production",
		allowInsecureBackend: false,
		backendTLSServerName: "server",
		backendTLSCAFile:     "/tmp/does-not-exist-ca.crt",
	})
	if err == nil {
		t.Fatal("expected backend TLS credential setup to fail for a missing CA file")
	}
}
