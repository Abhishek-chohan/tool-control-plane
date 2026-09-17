package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"toolplane/internal/server"
)

// TestServeHelpDocumentsLifecycle pins the help surface a newcomer sees:
// the dev-defaults posture and the production gates must be visible from
// `toolplane serve --help` alone.
func TestServeHelpDocumentsLifecycle(t *testing.T) {
	var out bytes.Buffer
	root := newRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"serve", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("serve --help: %v", err)
	}
	help := out.String()
	for _, fragment := range []string{
		"in-memory storage",
		"refuses to boot",
		"--port",
		"--metrics-listen",
		"--migrate-only",
		"--tls-cert-file",
	} {
		if !bytes.Contains([]byte(help), []byte(fragment)) {
			t.Fatalf("serve help missing %q:\n%s", fragment, help)
		}
	}
}

// TestVersionReportsBuild pins `toolplane --version`.
func TestVersionReportsBuild(t *testing.T) {
	var out bytes.Buffer
	root := newRootCommand("v9.9.9-test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("--version: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("v9.9.9-test")) {
		t.Fatalf("--version output missing the build version:\n%s", out.String())
	}
}

// TestServeCommandBootsAndServesHealth is the embedded-lifecycle proof:
// `toolplane serve` against in-memory dev defaults listens on gRPC,
// reports SERVING on standard grpc.health.v1, and drains cleanly on
// context cancellation — the same path SIGINT takes in production.
func TestServeCommandBootsAndServesHealth(t *testing.T) {
	// TOOLPLANE_AUTH_MODE defaults to fixed with a dev key when unset in
	// dev env; pin the fixed key so the boot matches the documented
	// development posture.
	t.Setenv("TOOLPLANE_ENV_MODE", "development")
	t.Setenv("TOOLPLANE_AUTH_MODE", "fixed")
	t.Setenv("TOOLPLANE_AUTH_FIXED_API_KEY", "dev-key")
	t.Setenv("TOOLPLANE_STORAGE_MODE", "memory")

	opts := server.DefaultOptions()
	opts.Port = 0 // port 0: kernel-assigned, parallel-test safe
	// The listener must print its address for the ping; runContext logs
	// it, but the test needs it programmatically — pick a port first.
	listen, err := pickFreePort()
	if err != nil {
		t.Fatalf("pick port: %v", err)
	}
	opts.Port = listen

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan int, 1)
	go func() { done <- server.RunContext(ctx, opts) }()

	// Poll the health service until the server reports SERVING (it flips
	// only after registration completes). NewClient connects lazily, so
	// the poll doubles as the dial-wait.
	addr := fmt.Sprintf("localhost:%d", listen)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		cancel()
		t.Fatalf("dial %s: %v", addr, err)
	}
	defer conn.Close()
	healthClient := healthpb.NewHealthClient(conn)
	deadline := time.Now().Add(10 * time.Second)
	for {
		checkCtx, checkCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		_, pingErr := healthClient.Check(checkCtx, &healthpb.HealthCheckRequest{})
		checkCancel()
		if pingErr == nil {
			break
		}
		if time.Now().After(deadline) {
			cancel()
			t.Fatalf("server never reported healthy on %s: %v", addr, pingErr)
		}
		time.Sleep(100 * time.Millisecond)
	}

	checkCtx, checkCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer checkCancel()
	resp, err := healthClient.Check(checkCtx, &healthpb.HealthCheckRequest{})
	if err != nil {
		cancel()
		t.Fatalf("health check: %v", err)
	}
	if resp.Status != healthpb.HealthCheckResponse_SERVING {
		cancel()
		t.Fatalf("health status = %v, want SERVING", resp.Status)
	}

	cancel()
	select {
	case code := <-done:
		if code != 0 {
			t.Fatalf("serve exit code = %d, want 0 on clean shutdown", code)
		}
	case <-time.After(25 * time.Second):
		t.Fatal("server did not shut down within 25s of cancellation")
	}
}

func pickFreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
