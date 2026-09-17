package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"toolplane/internal/gateway"
	"toolplane/internal/mcpgateway"
	"toolplane/internal/server"
)

// bootTestServer starts the control plane on a free port with in-memory
// dev defaults and returns the address; the caller cancels the context to
// drain it.
func bootTestServer(t *testing.T) string {
	t.Helper()
	t.Setenv("TOOLPLANE_ENV_MODE", "development")
	t.Setenv("TOOLPLANE_AUTH_MODE", "fixed")
	t.Setenv("TOOLPLANE_AUTH_FIXED_API_KEY", "dev-key")
	t.Setenv("TOOLPLANE_STORAGE_MODE", "memory")

	port, err := pickFreePort()
	if err != nil {
		t.Fatalf("pick port: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	opts := server.DefaultOptions()
	opts.Port = port
	done := make(chan int, 1)
	go func() { done <- server.RunContext(ctx, opts) }()
	// One cleanup doing cancel-then-join: Cleanup runs LIFO, so separate
	// cancellations and joins would wait before cancelling — a deadlock.
	t.Cleanup(func() {
		cancel()
		<-done
	})
	return fmt.Sprintf("localhost:%d", port)
}

// waitForHTTP polls url until it answers 200 or the deadline passes.
func waitForHTTP(t *testing.T, url string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	client := &http.Client{Timeout: 500 * time.Millisecond}
	for {
		resp, err := client.Get(url)
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("%s never answered 200 within the deadline", url)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// TestGatewayServeCommandProxiesHealth proves `toolplane gateway serve`
// embeds the real edge: booted against a live control plane it answers
// /health 200, and drains on cancellation.
func TestGatewayServeCommandProxiesHealth(t *testing.T) {
	backend := bootTestServer(t)

	port, err := pickFreePort()
	if err != nil {
		t.Fatalf("pick port: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan int, 1)
	go func() {
		opts := gateway.DefaultOptions()
		opts.Listen = fmt.Sprintf("127.0.0.1:%d", port)
		opts.Backend = backend
		done <- gateway.RunContext(ctx, opts)
	}()
	t.Cleanup(func() {
		cancel()
		<-done
	})

	waitForHTTP(t, fmt.Sprintf("http://127.0.0.1:%d/health", port))
}

// TestMCPServeCommandProxiesHealth proves `toolplane mcp serve` embeds
// the real facade: /health reports ok against a live control plane.
func TestMCPServeCommandProxiesHealth(t *testing.T) {
	backend := bootTestServer(t)

	port, err := pickFreePort()
	if err != nil {
		t.Fatalf("pick port: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan int, 1)
	go func() {
		opts := mcpgateway.DefaultOptions()
		opts.Listen = fmt.Sprintf("127.0.0.1:%d", port)
		opts.Backend = backend
		done <- mcpgateway.RunContext(ctx, opts)
	}()
	t.Cleanup(func() {
		cancel()
		<-done
	})

	waitForHTTP(t, fmt.Sprintf("http://127.0.0.1:%d/health", port))
}

// TestRootCommandExposesEdges pins the command tree: the edges are
// reachable as `toolplane gateway serve` and `toolplane mcp serve`.
func TestRootCommandExposesEdges(t *testing.T) {
	root := newRootCommand("test")
	for _, path := range [][]string{{"gateway", "serve"}, {"mcp", "serve"}, {"serve"}} {
		found, _, err := root.Find(path)
		if err != nil || found == nil || found.Name() != path[len(path)-1] {
			t.Fatalf("command path %v not found in the tree (err=%v)", path, err)
		}
	}
}
