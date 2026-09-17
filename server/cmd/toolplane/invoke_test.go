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

	"net"
	"toolplane/internal/cli"
	"toolplane/internal/server"
	proto "toolplane/proto"
)

// invokeE2E boots the in-memory control plane, registers a machine with
// an echo tool over the real RPC surface, and returns a connected client
// plus the session it registered in.
func invokeE2E(t *testing.T) (proto.ToolServiceClient, proto.RequestsServiceClient, string, cli.Connection, func()) {
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
	serverCtx, cancelServer := context.WithCancel(context.Background())
	opts := server.DefaultOptions()
	opts.Listener = lis
	opts.Port = port
	done := make(chan int, 1)
	go func() { done <- server.RunContext(serverCtx, opts) }()

	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		cancelServer()
		t.Fatalf("dial: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	metadataCtx := metadata.AppendToOutgoingContext(ctx, "api_key", "dev-key")
	toolSvc := proto.NewToolServiceClient(conn)
	machineSvc := proto.NewMachinesServiceClient(conn)
	requestSvc := proto.NewRequestsServiceClient(conn)

	// Wait for readiness through the api.v1 health probe.
	healthDeadline := time.Now().Add(10 * time.Second)
	for {
		healthCtx, healthCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		_, pingErr := toolSvc.HealthCheck(metadata.AppendToOutgoingContext(healthCtx, "api_key", "dev-key"),
			&proto.HealthCheckRequest{})
		healthCancel()
		if pingErr == nil {
			break
		}
		if time.Now().After(healthDeadline) {
			cancel()
			cancelServer()
			t.Fatalf("server never became healthy: %v", pingErr)
		}
		time.Sleep(50 * time.Millisecond)
	}

	const (
		sessionID = "sess-invoke-e2e"
		machineID = "machine-invoke-e2e"
	)
	registerCtx, registerCancel := context.WithTimeout(metadataCtx, 5*time.Second)
	_, err = machineSvc.RegisterMachine(registerCtx, &proto.RegisterMachineRequest{
		SessionId:   sessionID,
		MachineId:   machineID,
		SdkVersion:  "test",
		SdkLanguage: "go",
		Tools: []*proto.RegisterToolRequest{{
			Name:        "add",
			Description: "adds two numbers",
			Schema:      `{"type":"object"}`,
		}},
	})
	registerCancel()
	if err != nil {
		cancel()
		cancelServer()
		t.Fatalf("register machine: %v", err)
	}

	connection := cli.NewConnection(addr, "dev-key", 30*time.Second)
	cleanup := func() {
		cancel()
		conn.Close()
		cancelServer()
		<-done
	}
	return toolSvc, requestSvc, sessionID, connection, cleanup
}

// startEchoProvider claims and completes any pending add-request on the
// machine that registered the tool, like the SDK provider loop does.
func startEchoProvider(t *testing.T, requestSvc proto.RequestsServiceClient, ctx context.Context, sessionID, machineID string, metadataCtx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			claimCtx, cancel := context.WithTimeout(metadataCtx, time.Second)
			claimed, err := requestSvc.ClaimNextRequest(claimCtx, &proto.ClaimNextRequestRequest{
				SessionId: sessionID,
				MachineId: machineID,
				ToolNames: []string{"add"},
			})
			cancel()
			if err != nil || claimed == nil || !claimed.Claimed {
				time.Sleep(25 * time.Millisecond)
				continue
			}
			runCtx, runCancel := context.WithTimeout(metadataCtx, 5*time.Second)
			_, _ = requestSvc.UpdateRequest(runCtx, &proto.UpdateRequestRequest{
				SessionId:  sessionID,
				RequestId:  claimed.Request.Id,
				Status:     proto.RequestStatus_REQUEST_STATUS_RUNNING,
				MachineId:  machineID,
				LeaseEpoch: claimed.Request.LeaseEpoch,
			})
			var input struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}
			_ = json.Unmarshal([]byte(claimed.Request.Input), &input)
			sum := input.A + input.B
			_, _ = requestSvc.SubmitRequestResult(runCtx, &proto.SubmitRequestResultRequest{
				SessionId:  sessionID,
				RequestId:  claimed.Request.Id,
				Result:     fmt.Sprintf(`{"sum":%v}`, sum),
				ResultType: "resolution",
				MachineId:  machineID,
				LeaseEpoch: claimed.Request.LeaseEpoch,
			})
			runCancel()
		}
	}()
}

// TestInvokeRunCompletesAgainstServer pins the happy path end to end:
// invoke --wait against a live server with a serving provider returns
// exit 0 and the tool's JSON result.
func TestInvokeRunCompletesAgainstServer(t *testing.T) {
	toolSvc, requestSvc, sessionID, connection, cleanup := invokeE2E(t)
	defer cleanup()

	metadataCtx := metadata.AppendToOutgoingContext(context.Background(), "api_key", "dev-key")
	startEchoProvider(t, requestSvc, metadataCtx, sessionID, "machine-invoke-e2e", metadataCtx)

	conn := connection
	var out bytes.Buffer
	code, err := runInvoke(context.Background(), conn, &out, invokeOptions{
		Tool:    "add",
		Session: sessionID,
		Input:   `{"a":2,"b":3}`,
		Wait:    10 * time.Second,
		Format:  cli.FormatTable,
	})
	if err != nil {
		t.Fatalf("invoke: %v (output: %s)", err, out.String())
	}
	if code != cli.ExitOK {
		t.Fatalf("invoke exit code = %d, want 0 (output: %s)", code, out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte(`"sum":5`)) {
		t.Fatalf("invoke output missing the result (output: %s)", out.String())
	}
	_ = toolSvc
}

// TestInvokeValidatesArguments pins argument validation before any
// network activity: missing session and malformed input are
// invalid-argument exits with teaching messages.
func TestInvokeValidatesArguments(t *testing.T) {
	conn := cli.NewConnection("", "", 0)
	var out bytes.Buffer

	code, err := runInvoke(context.Background(), conn, &out, invokeOptions{
		Tool:   "add",
		Input:  `{}`,
		Format: cli.FormatTable,
	})
	if code != cli.ExitInvalidArgument || err == nil {
		t.Fatalf("missing session: code=%d err=%v, want invalid-argument", code, err)
	}

	code, err = runInvoke(context.Background(), conn, &out, invokeOptions{
		Tool:    "add",
		Session: "sess-x",
		Input:   `not json`,
		Format:  cli.FormatTable,
	})
	if code != cli.ExitInvalidArgument || err == nil {
		t.Fatalf("malformed input: code=%d err=%v, want invalid-argument", code, err)
	}
	if !bytes.Contains([]byte(err.Error()), []byte("--input must be a JSON object")) {
		t.Fatalf("validation error must teach: %v", err)
	}
}

// TestInvokeFireAndForgetPinsRequestID: with --wait 0 the call returns
// immediately with the request ID and a pending status, exit 0.
func TestInvokeFireAndForgetPinsRequestID(t *testing.T) {
	toolSvc, requestSvc, sessionID, connection, cleanup := invokeE2E(t)
	defer cleanup()

	metadataCtx := metadata.AppendToOutgoingContext(context.Background(), "api_key", "dev-key")
	startEchoProvider(t, requestSvc, metadataCtx, sessionID, "machine-invoke-e2e", metadataCtx)

	conn := connection
	var out bytes.Buffer
	code, err := runInvoke(context.Background(), conn, &out, invokeOptions{
		Tool:    "add",
		Session: sessionID,
		Input:   `{"a":1,"b":1}`,
		Wait:    0,
		Format:  cli.FormatTable,
	})
	if err != nil {
		t.Fatalf("fire-and-forget invoke: %v", err)
	}
	if code != cli.ExitOK {
		t.Fatalf("fire-and-forget exit code = %d, want 0", code)
	}
	if !bytes.Contains(out.Bytes(), []byte("request ")) || !bytes.Contains(out.Bytes(), []byte("toolplane request status")) {
		t.Fatalf("fire-and-forget output must carry the request ID and the follow-up command: %s", out.String())
	}
	_ = toolSvc
}
