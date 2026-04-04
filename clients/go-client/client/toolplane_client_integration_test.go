package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/grpc/metadata"

	pb "toolplane-go-client/proto"
)

const testConformanceAPIKey = "toolplane-conformance-fixture-key"

type providerLoop struct {
	cancel context.CancelFunc
	done   chan error
}

func (p *providerLoop) Stop() error {
	p.cancel()
	return <-p.done
}

func TestGRPCLiveExecuteAndStreamExecuteTool(t *testing.T) {
	if os.Getenv("TOOLPLANE_RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("set TOOLPLANE_RUN_INTEGRATION_TESTS=1 to run live gRPC integration tests")
	}

	grpcPort := freeTCPPort(t)
	serverCmd, serverOutput := startGRPCServer(t, grpcPort)
	defer stopCommand(serverCmd)

	address := fmt.Sprintf("127.0.0.1:%d", grpcPort)
	if err := waitForTCP(address, 15*time.Second); err != nil {
		t.Fatalf("server did not start listening: %v\n%s", err, serverOutput.String())
	}

	client, err := NewToolplaneClient(ProtocolGRPC, "127.0.0.1", grpcPort, "", "go-live-test-user", testConformanceAPIKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer func() {
		if err := client.Disconnect(); err != nil {
			t.Fatalf("failed to disconnect client: %v", err)
		}
	}()

	if err := client.Connect(); err != nil {
		t.Fatalf("failed to connect client: %v\n%s", err, serverOutput.String())
	}

	session, err := client.CreateSession("Go Live gRPC Test", "Integration coverage for live gRPC execution", "integration")
	if err != nil {
		t.Fatalf("failed to create session: %v\n%s", err, serverOutput.String())
	}

	toolDefinitions := []*pb.RegisterToolRequest{
		{
			Name:        "live_echo",
			Description: "Echoes a message back in JSON",
			Schema:      `{"type":"object","properties":{"message":{"type":"string"}},"required":["message"]}`,
		},
		{
			Name:        "live_stream",
			Description: "Streams ordered chunks before returning a final array",
			Schema:      `{"type":"object","properties":{"prefix":{"type":"string"},"count":{"type":"integer"}},"required":["prefix","count"]}`,
		},
	}

	machine, err := client.RegisterMachine("", "1.0.0-test", toolDefinitions)
	if err != nil {
		t.Fatalf("failed to register machine: %v\n%s", err, serverOutput.String())
	}

	provider, err := startProviderLoop(address, session.Id, machine.Id, testConformanceAPIKey)
	if err != nil {
		t.Fatalf("failed to start provider loop: %v", err)
	}
	defer func() {
		if err := provider.Stop(); err != nil {
			t.Fatalf("provider loop failed: %v\n%s", err, serverOutput.String())
		}
	}()

	request, err := client.ExecuteTool(context.Background(), "live_echo", map[string]interface{}{"message": "hello from go"})
	if err != nil {
		t.Fatalf("failed to execute unary tool: %v\n%s", err, serverOutput.String())
	}
	if request.GetStatus() != "done" {
		t.Fatalf("expected unary request to finish with status done, got %q", request.GetStatus())
	}

	var unaryResult map[string]string
	if err := json.Unmarshal([]byte(request.GetResult()), &unaryResult); err != nil {
		t.Fatalf("failed to decode unary result %q: %v", request.GetResult(), err)
	}
	if unaryResult["echo"] != "hello from go" {
		t.Fatalf("unexpected unary result: %#v", unaryResult)
	}

	observedChunks := make([]*pb.ExecuteToolChunk, 0, 4)
	_, err = client.StreamExecuteTool(context.Background(), "live_stream", map[string]interface{}{"prefix": "part", "count": 3}, func(chunk *pb.ExecuteToolChunk) error {
		observedChunks = append(observedChunks, chunk)
		return nil
	})
	if err != nil {
		t.Fatalf("failed to execute streaming tool: %v\n%s", err, serverOutput.String())
	}

	if len(observedChunks) != 4 {
		t.Fatalf("expected 4 chunks including final marker, got %d", len(observedChunks))
	}

	for index, chunk := range observedChunks[:3] {
		wantSeq := int32(index + 1)
		wantPayload := fmt.Sprintf("part-%d", index+1)
		if chunk.GetSeq() != wantSeq {
			t.Fatalf("chunk %d seq mismatch: got %d want %d", index, chunk.GetSeq(), wantSeq)
		}
		if chunk.GetChunk() != wantPayload {
			t.Fatalf("chunk %d payload mismatch: got %q want %q", index, chunk.GetChunk(), wantPayload)
		}
		if chunk.GetIsFinal() {
			t.Fatalf("chunk %d was unexpectedly marked final", index)
		}
	}

	finalChunk := observedChunks[len(observedChunks)-1]
	if !finalChunk.GetIsFinal() {
		t.Fatalf("expected final chunk to be marked final")
	}
	if finalChunk.GetSeq() != 4 {
		t.Fatalf("final chunk seq mismatch: got %d want 4", finalChunk.GetSeq())
	}

	var finalResult []string
	if err := json.Unmarshal([]byte(finalChunk.GetChunk()), &finalResult); err != nil {
		t.Fatalf("failed to decode final stream chunk %q: %v", finalChunk.GetChunk(), err)
	}
	if len(finalResult) != 3 || finalResult[0] != "part-1" || finalResult[1] != "part-2" || finalResult[2] != "part-3" {
		t.Fatalf("unexpected final stream result: %#v", finalResult)
	}
}

func startProviderLoop(address, sessionID, machineID, apiKey string) (*providerLoop, error) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	conn, err := dialGRPC(ctx, address)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to dial provider connection: %w", err)
	}

	requestsClient := pb.NewRequestsServiceClient(conn)

	go func() {
		defer close(done)
		defer conn.Close()

		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				done <- nil
				return
			case <-ticker.C:
				if err := processPendingRequests(ctx, requestsClient, sessionID, machineID, apiKey); err != nil {
					if ctx.Err() != nil {
						done <- nil
						return
					}
					done <- err
					return
				}
			}
		}
	}()

	return &providerLoop{cancel: cancel, done: done}, nil
}

func processPendingRequests(
	ctx context.Context,
	requestsClient pb.RequestsServiceClient,
	sessionID, machineID, apiKey string,
) error {
	listCtx, listCancel := outgoingContext(ctx, apiKey, defaultGRPCCallTimeout)
	response, err := requestsClient.ListRequests(listCtx, &pb.ListRequestsRequest{
		SessionId: sessionID,
		Status:    "pending",
		Limit:     20,
	})
	listCancel()
	if err != nil {
		return fmt.Errorf("failed to list pending requests: %w", err)
	}

	for _, request := range response.Requests {
		claimCtx, claimCancel := outgoingContext(ctx, apiKey, defaultGRPCCallTimeout)
		claimed, err := requestsClient.ClaimRequest(claimCtx, &pb.ClaimRequestRequest{
			SessionId: sessionID,
			RequestId: request.Id,
			MachineId: machineID,
		})
		claimCancel()
		if err != nil {
			return fmt.Errorf("failed to claim request %s: %w", request.Id, err)
		}

		if err := fulfillRequest(ctx, requestsClient, sessionID, apiKey, claimed); err != nil {
			return err
		}
	}

	return nil
}

func fulfillRequest(
	ctx context.Context,
	requestsClient pb.RequestsServiceClient,
	sessionID, apiKey string,
	request *pb.Request,
) error {
	params := map[string]interface{}{}
	if request.GetInput() != "" {
		if err := json.Unmarshal([]byte(request.GetInput()), &params); err != nil {
			return fmt.Errorf("failed to decode request params for %s: %w", request.GetId(), err)
		}
	}

	switch request.GetToolName() {
	case "live_echo":
		message := fmt.Sprint(params["message"])
		resultJSON, err := json.Marshal(map[string]string{"echo": message})
		if err != nil {
			return fmt.Errorf("failed to encode unary result: %w", err)
		}
		return submitRequestResult(ctx, requestsClient, sessionID, apiKey, request.GetId(), string(resultJSON), "resolution")
	case "live_stream":
		prefix := fmt.Sprint(params["prefix"])
		if prefix == "" {
			prefix = "chunk"
		}
		count := toInt(params["count"], 1)
		chunks := make([]string, 0, count)
		for index := 0; index < count; index++ {
			payload := fmt.Sprintf("%s-%d", prefix, index+1)
			chunks = append(chunks, payload)

			appendCtx, appendCancel := outgoingContext(ctx, apiKey, defaultGRPCCallTimeout)
			_, err := requestsClient.AppendRequestChunks(appendCtx, &pb.AppendRequestChunksRequest{
				SessionId:  sessionID,
				RequestId:  request.GetId(),
				Chunks:     []string{payload},
				ResultType: "streaming",
			})
			appendCancel()
			if err != nil {
				return fmt.Errorf("failed to append chunk for request %s: %w", request.GetId(), err)
			}

			time.Sleep(250 * time.Millisecond)
		}

		resultJSON, err := json.Marshal(chunks)
		if err != nil {
			return fmt.Errorf("failed to encode streaming result: %w", err)
		}
		return submitRequestResult(ctx, requestsClient, sessionID, apiKey, request.GetId(), string(resultJSON), "resolution")
	default:
		return submitRequestResult(ctx, requestsClient, sessionID, apiKey, request.GetId(), `"unknown tool"`, "rejection")
	}
}

func submitRequestResult(
	ctx context.Context,
	requestsClient pb.RequestsServiceClient,
	sessionID, apiKey, requestID, result, resultType string,
) error {
	submitCtx, submitCancel := outgoingContext(ctx, apiKey, defaultGRPCCallTimeout)
	defer submitCancel()

	_, err := requestsClient.SubmitRequestResult(submitCtx, &pb.SubmitRequestResultRequest{
		SessionId:  sessionID,
		RequestId:  requestID,
		Result:     result,
		ResultType: resultType,
		Meta:       map[string]string{"handled_by": "go-integration-test"},
	})
	if err != nil {
		return fmt.Errorf("failed to submit request result for %s: %w", requestID, err)
	}

	return nil
}

func outgoingContext(parent context.Context, apiKey string, timeout time.Duration) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}

	ctx, cancel := context.WithTimeout(parent, timeout)
	if apiKey != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "api_key", apiKey)
	}
	return ctx, cancel
}

func toInt(value interface{}, fallback int) int {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	default:
		return fallback
	}
}

func startGRPCServer(t *testing.T, grpcPort int) (*exec.Cmd, *bytes.Buffer) {
	t.Helper()

	serverRoot := filepath.Clean(filepath.Join("..", "..", "..", "server"))
	cmd := exec.Command("go", "run", "./cmd/server", "-port", fmt.Sprintf("%d", grpcPort))
	cmd.Dir = serverRoot
	cmd.Env = append(
		os.Environ(),
		"TOOLPLANE_ENV_MODE=development",
		"TOOLPLANE_AUTH_MODE=fixed",
		fmt.Sprintf("TOOLPLANE_AUTH_FIXED_API_KEY=%s", testConformanceAPIKey),
		"TOOLPLANE_STORAGE_MODE=memory",
	)

	output := &bytes.Buffer{}
	cmd.Stdout = output
	cmd.Stderr = output

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start grpc server: %v", err)
	}

	return cmd, output
}

func stopCommand(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	_ = cmd.Process.Kill()
	_, _ = cmd.Process.Wait()
}

func waitForTCP(address string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, 200*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timed out waiting for %s", address)
}

func freeTCPPort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to allocate free port: %v", err)
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}
