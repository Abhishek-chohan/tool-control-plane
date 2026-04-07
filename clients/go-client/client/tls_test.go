package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTransportCredentialsUseInsecureWhenTLSDisabled(t *testing.T) {
	client := &ToolplaneClient{}

	creds, err := client.transportCredentials()
	if err != nil {
		t.Fatalf("transportCredentials returned unexpected error: %v", err)
	}
	if got := creds.Info().SecurityProtocol; got != "insecure" {
		t.Fatalf("transportCredentials security protocol = %q, want insecure", got)
	}
}

func TestTransportCredentialsLoadCustomCAForTLS(t *testing.T) {
	caPath := writeTestCACertificate(t)
	client := &ToolplaneClient{tlsConfig: GRPCTLSConfig{Enabled: true, CACertPath: caPath, ServerName: "localhost"}}

	creds, err := client.transportCredentials()
	if err != nil {
		t.Fatalf("transportCredentials returned unexpected error: %v", err)
	}
	if got := creds.Info().SecurityProtocol; got != "tls" {
		t.Fatalf("transportCredentials security protocol = %q, want tls", got)
	}
}

func TestTransportCredentialsRejectMissingCAFile(t *testing.T) {
	client := &ToolplaneClient{tlsConfig: GRPCTLSConfig{Enabled: true, CACertPath: filepath.Join(t.TempDir(), "missing-ca.crt")}}

	_, err := client.transportCredentials()
	if err == nil {
		t.Fatal("transportCredentials succeeded for missing CA file")
	}
}

func writeTestCACertificate(t *testing.T) string {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate test CA private key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Toolplane Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	certificateDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("failed to create test CA certificate: %v", err)
	}

	certificatePEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateDER})
	certificatePath := filepath.Join(t.TempDir(), "ca.crt")
	if err := os.WriteFile(certificatePath, certificatePEM, 0o600); err != nil {
		t.Fatalf("failed to write test CA certificate: %v", err)
	}

	return certificatePath
}
