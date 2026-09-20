package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	proto "toolplane/proto"
)

// writeSelfSignedCert mints a throwaway certificate for 127.0.0.1/localhost
// and returns the cert/key file paths plus the cert to trust as root.
func writeSelfSignedCert(t *testing.T) (certFile, keyFile string, cert *x509.Certificate) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		t.Fatalf("serial: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "toolplane-test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:         true,
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}

	dir := t.TempDir()
	certFile = filepath.Join(dir, "server.crt")
	keyFile = filepath.Join(dir, "server.key")
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	if err := os.WriteFile(certFile, certPEM, 0o600); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}
	return certFile, keyFile, parsed
}

func bootTLSServer(t *testing.T, certFile, keyFile string) (addr string, stop func()) {
	t.Helper()

	t.Setenv("TOOLPLANE_ENV_MODE", "development")
	t.Setenv("TOOLPLANE_AUTH_MODE", "fixed")
	t.Setenv("TOOLPLANE_AUTH_FIXED_API_KEY", "dev-key")
	t.Setenv("TOOLPLANE_STORAGE_MODE", "memory")

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("bind listener: %v", err)
	}
	port := lis.Addr().(*net.TCPAddr).Port
	opts := DefaultOptions()
	opts.Listener = lis
	opts.Port = port
	opts.TLSCertFile = certFile
	opts.TLSKeyFile = keyFile
	opts.MetricsListen = ""

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan int, 1)
	go func() { done <- RunContext(ctx, opts) }()
	return fmt.Sprintf("localhost:%d", port), func() {
		cancel()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatalf("server did not stop after cancel")
		}
	}
}

func waitHealthyOverTLS(t *testing.T, svc proto.ToolServiceClient) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		ctx, cancel := context.WithTimeout(metadata.AppendToOutgoingContext(context.Background(), "api_key", "dev-key"), 500*time.Millisecond)
		_, pingErr := svc.HealthCheck(ctx, &proto.HealthCheckRequest{})
		cancel()
		if pingErr == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("server never became healthy over TLS: %v", pingErr)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestServerTLSHandshakeServesAPIOverLoopback(t *testing.T) {
	certFile, keyFile, cert := writeSelfSignedCert(t)
	addr, stop := bootTLSServer(t, certFile, keyFile)
	defer stop()

	pool := x509.NewCertPool()
	pool.AddCert(cert)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		ServerName: "localhost",
		RootCAs:    pool,
		MinVersion: tls.VersionTLS12,
	})))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	waitHealthyOverTLS(t, proto.NewToolServiceClient(conn))
}

func TestServerTLSRefusesPlaintextClients(t *testing.T) {
	certFile, keyFile, _ := writeSelfSignedCert(t)
	addr, stop := bootTLSServer(t, certFile, keyFile)
	defer stop()

	// Wait until the listener is actually serving TLS before asserting the
	// plaintext handshake is refused — otherwise the failure is a dial race.
	pool := x509.NewCertPool()
	pool.AddCert(mustCert(t, certFile))
	tlsConn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		ServerName: "localhost",
		RootCAs:    pool,
		MinVersion: tls.VersionTLS12,
	})))
	if err != nil {
		t.Fatalf("dial tls: %v", err)
	}
	waitHealthyOverTLS(t, proto.NewToolServiceClient(tlsConn))
	tlsConn.Close()

	plainConn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial plaintext: %v", err)
	}
	defer plainConn.Close()

	plainSvc := proto.NewToolServiceClient(plainConn)
	ctx, cancel := context.WithTimeout(metadata.AppendToOutgoingContext(context.Background(), "api_key", "dev-key"), 2*time.Second)
	defer cancel()
	if _, err := plainSvc.HealthCheck(ctx, &proto.HealthCheckRequest{}); err == nil {
		t.Fatal("plaintext client unexpectedly served by a TLS listener")
	}
}

func mustCert(t *testing.T, certFile string) *x509.Certificate {
	t.Helper()
	raw, err := os.ReadFile(certFile)
	if err != nil {
		t.Fatalf("read cert: %v", err)
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		t.Fatal("no PEM block in cert file")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}
	return cert
}
