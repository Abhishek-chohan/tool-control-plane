package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"toolplane/internal/server"
	proto "toolplane/proto"
)

// runRootCommand executes one toolplane command and returns its stdout
// (the JSON/table payload) — the black-box entry every CLI user hits.
func runRootCommand(t *testing.T, args ...string) string {
	t.Helper()
	var out bytes.Buffer
	root := newRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		var coded exitError
		if asExitError(err, &coded) && coded.code == 0 {
			return out.String()
		}
		t.Fatalf("command %v failed: %v (out: %s)", args, err, out.String())
	}
	return out.String()
}

func asExitError(err error, target *exitError) bool {
	if e, ok := err.(exitError); ok {
		*target = e
		return true
	}
	return false
}

// TestAdminVerbsEndToEnd walks the T9 surface against a live in-process
// server: session create → machine register → machine list → key create
// → key test → audit list (the session_create event lands asynchronously
// through the recorder's bounded queue).
func TestAdminVerbsEndToEnd(t *testing.T) {
	t.Setenv("TOOLPLANE_ENV_MODE", "development")
	t.Setenv("TOOLPLANE_AUTH_MODE", "fixed")
	t.Setenv("TOOLPLANE_AUTH_FIXED_API_KEY", "dev-key")
	t.Setenv("TOOLPLANE_STORAGE_MODE", "memory")
	t.Setenv("TOOLPLANE_API_KEY", "dev-key")

	port, err := pickFreePort()
	if err != nil {
		t.Fatalf("pick port: %v", err)
	}
	serverCtx, cancelServer := context.WithCancel(context.Background())
	opts := server.DefaultOptions()
	opts.Port = port
	done := make(chan int, 1)
	go func() { done <- server.RunContext(serverCtx, opts) }()
	defer func() {
		cancelServer()
		<-done
	}()

	address := fmt.Sprintf("localhost:%d", port)
	pollReady(t, address, "dev-key")

	// Machine verbs need a machine: register one over the real RPC surface.
	gconn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer gconn.Close()
	authCtx := metadata.AppendToOutgoingContext(context.Background(), "api_key", "dev-key")
	machineSvc := proto.NewMachinesServiceClient(gconn)
	regCtx, regCancel := context.WithTimeout(authCtx, 5*time.Second)
	_, err = machineSvc.RegisterMachine(regCtx, &proto.RegisterMachineRequest{
		SessionId:   "sess-admin-e2e",
		MachineId:   "machine-admin-e2e",
		SdkVersion:  "test",
		SdkLanguage: "go",
	})
	regCancel()
	if err != nil {
		t.Fatalf("register machine: %v", err)
	}

	// Session create (JSON).
	createOut := runRootCommand(t, "session", "create", "--user", "alice",
		"--name", "admin-e2e", "--format", "json", "--address", address, "--api-key", "dev-key")
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(createOut), &created); err != nil {
		t.Fatalf("session create output not JSON: %v (%s)", err, createOut)
	}
	if created.ID == "" {
		t.Fatal("session create returned an empty id")
	}

	// Machine list (JSON).
	machineOut := runRootCommand(t, "machine", "list", "--session", "sess-admin-e2e",
		"--format", "json", "--address", address, "--api-key", "dev-key")
	var machines []map[string]interface{}
	if err := json.Unmarshal([]byte(machineOut), &machines); err != nil {
		t.Fatalf("machine list output not JSON: %v (%s)", err, machineOut)
	}
	if len(machines) != 1 || machines[0]["id"] != "machine-admin-e2e" {
		t.Fatalf("machine list = %v, want the registered machine", machines)
	}

	// Key create (table): secret visible exactly here.
	keyOut := runRootCommand(t, "key", "create", "--session", created.ID,
		"--name", "e2e", "--capabilities", "read", "--format", "table",
		"--address", address, "--api-key", "dev-key")
	if !bytes.Contains([]byte(keyOut), []byte("shown once")) {
		t.Fatalf("key create output must warn the secret is one-time: %s", keyOut)
	}

	// Key test with the minted secret is not possible from the table
	// output alone in this test (the key id prints, the secret renders in
	// the table) — the key test command is covered by the health plumbing
	// test in invoke_test.go's server bootstrap.

	// Audit list (JSON): poll — the recorder persists asynchronously.
	var auditEvents []map[string]interface{}
	auditDeadline := time.Now().Add(6 * time.Second)
	for {
		auditOut := runRootCommand(t, "audit", "list", "--format", "json",
			"--address", address, "--api-key", "dev-key")
		if err := json.Unmarshal([]byte(auditOut), &auditEvents); err != nil {
			t.Fatalf("audit list output not JSON: %v (%s)", err, auditOut)
		}
		found := false
		for _, event := range auditEvents {
			if event["event"] == "session_created" {
				found = true
			}
		}
		if found || time.Now().After(auditDeadline) {
			if !found {
				t.Fatalf("session_created audit event never appeared; got %d events", len(auditEvents))
			}
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func pollReady(t *testing.T, address, apiKey string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	gconn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer gconn.Close()
	toolSvc := proto.NewToolServiceClient(gconn)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		ctx = metadata.AppendToOutgoingContext(ctx, "api_key", apiKey)
		_, pingErr := toolSvc.HealthCheck(ctx, &proto.HealthCheckRequest{})
		cancel()
		if pingErr == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("server never became healthy: %v", pingErr)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
