package mcp_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"toolplane/pkg/mcp"
	"toolplane/pkg/model"
	"toolplane/pkg/service"
	"toolplane/pkg/storage/memory"
	"toolplane/pkg/trace"
)

func TestInitializeNegotiatesLegacyVersion(t *testing.T) {
	conn, _, _ := startBackend(t)
	handler := newFacade(t, conn)

	for _, requested := range mcp.LegacyProtocolVersions {
		status, envelope := postRPC(t, handler, "initialize", map[string]any{
			"protocolVersion": requested,
			"capabilities":    map[string]any{},
			"clientInfo":      map[string]any{"name": "legacy-client", "version": "1.0.0"},
		})
		if status != http.StatusOK {
			t.Fatalf("initialize %s: status = %d", requested, status)
		}
		result := rpcResult(t, envelope)
		if result["protocolVersion"] != requested {
			t.Fatalf("initialize %s negotiated %v", requested, result["protocolVersion"])
		}
		serverInfo, ok := result["serverInfo"].(map[string]any)
		if !ok || serverInfo["name"] != mcp.ServerName {
			t.Fatalf("initialize %s serverInfo = %v", requested, result["serverInfo"])
		}
		if _, ok := result["capabilities"].(map[string]any); !ok {
			t.Fatalf("initialize %s missing capabilities: %v", requested, result)
		}
	}
}

func TestInitializeUnsupportedVersionGetsServerDefault(t *testing.T) {
	conn, _, _ := startBackend(t)
	handler := newFacade(t, conn)

	// Per the legacy handshake, an unsupported requested version gets the
	// server's supported revision back (not an error); the client decides
	// whether to continue.
	status, envelope := postRPC(t, handler, "initialize", map[string]any{
		"protocolVersion": "1999-01-01",
	})
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	result := rpcResult(t, envelope)
	if result["protocolVersion"] != mcp.DefaultLegacyProtocolVersion {
		t.Fatalf("negotiated %v, want default %s", result["protocolVersion"], mcp.DefaultLegacyProtocolVersion)
	}
}

func TestPingServedWithoutMeta(t *testing.T) {
	conn, _, _ := startBackend(t)
	handler := newFacade(t, conn)

	status, envelope := postRPC(t, handler, "ping", map[string]any{})
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	result := rpcResult(t, envelope)
	if result == nil {
		t.Fatalf("ping missing result: %v", envelope)
	}
}

// TestLegacyToolsListAndCallEndToEnd proves the compatibility path serves
// initialize-handshake clients: tools/list returns the plain legacy shape,
// and tools/call executes a real provider round trip without any _meta on
// either the request or the result.
func TestLegacyToolsListAndCallEndToEnd(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := memory.New()
	tracer := trace.NopTracer()
	toolService := service.NewToolService(tracer, store)
	sessionService := service.NewSessionsService(tracer, store)
	machineService := service.NewMachinesService(ctx, toolService, tracer, store)
	requestService := service.NewRequestsService(ctx, toolService, machineService, tracer, store)
	tasksService := service.NewTasksService(ctx, toolService, machineService, requestService, tracer, store)

	conn := serveBackend(t, toolService, sessionService, machineService, requestService, tasksService)
	handler := mcp.NewServer(conn, mcp.WithPollInterval(20*time.Millisecond), mcp.WithSyncTimeout(10*time.Second)).Handler()

	const machineID = "machine-legacy-e2e"

	// First contact: a meta-less tools/list makes the gateway provision its
	// per-API-key session. Find that session so the provider can register
	// into it.
	status, envelope := postRPC(t, handler, "tools/list", map[string]any{})
	if status != http.StatusOK {
		t.Fatalf("provisioning tools/list status = %d", status)
	}
	result := rpcResult(t, envelope)
	if _, ok := result["tools"].([]any); !ok {
		t.Fatalf("legacy tools/list missing tools array: %v", result)
	}

	sessions, err := sessionService.ListSessions(mcp.DefaultUserID)
	if err != nil {
		t.Fatalf("list gateway sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("gateway provisioned %d sessions, want 1", len(sessions))
	}
	sessionID := sessions[0].ID

	if _, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	// The catalog is now visible through the legacy list.
	status, envelope = postRPC(t, handler, "tools/list", map[string]any{})
	if status != http.StatusOK {
		t.Fatalf("tools/list status = %d", status)
	}
	result = rpcResult(t, envelope)
	if _, hasResultType := result["resultType"]; hasResultType {
		t.Fatalf("legacy tools/list carries 2026 resultType: %v", result)
	}
	tools, ok := result["tools"].([]any)
	if !ok || len(tools) != 1 {
		t.Fatalf("legacy tools/list tools = %v", result["tools"])
	}

	// Drive the provider side concurrently, as a real provider would.
	go func() {
		var requestID string
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			requests, err := requestService.ListRequests(sessionID, model.RequestStatusPending, "", 10, 0)
			if err == nil && len(requests) > 0 {
				requestID = requests[0].ID
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if requestID == "" {
			t.Errorf("provider never saw a pending request")
			return
		}
		claimed, err := requestService.ClaimRequest(sessionID, requestID, machineID)
		if err != nil {
			t.Errorf("claim: %v", err)
			return
		}
		if err := requestService.SubmitRequestResult(sessionID, requestID, machineID, claimed.LeaseEpoch, map[string]string{"echo": "legacy-hello"}, model.ResultTypeResolution, nil); err != nil {
			t.Errorf("submit result: %v", err)
		}
	}()

	status, envelope = postRPC(t, handler, "tools/call", map[string]any{
		"name":      "echo",
		"arguments": map[string]any{"message": "legacy-hello"},
	})
	if status != http.StatusOK {
		t.Fatalf("tools/call status = %d (body %v)", status, envelope)
	}
	result = rpcResult(t, envelope)
	if isTruthyErrorResult(result) {
		t.Fatalf("legacy tools/call returned an error result: %v", result)
	}
	if _, hasMeta := result["_meta"]; hasMeta {
		t.Fatalf("legacy result carries 2026 _meta: %v", result["_meta"])
	}
	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("legacy result missing content: %v", result)
	}
	if !isTruthyErrorResult(map[string]any{}) && !strings.Contains(strings.ToLower(serializeContent(content)), "legacy-hello") {
		t.Fatalf("legacy result content missing echoed text: %v", content)
	}
}

func isTruthyErrorResult(result map[string]any) bool {
	isError, ok := result["isError"].(bool)
	return ok && isError
}

// serializeContent flattens content blocks into one lowercase string for
// substring assertions.
func serializeContent(content []any) string {
	var b strings.Builder
	for _, block := range content {
		if textBlock, ok := block.(map[string]any); ok {
			if text, ok := textBlock["text"].(string); ok {
				b.WriteString(text)
				b.WriteString("\n")
			}
		}
	}
	return strings.ToLower(b.String())
}
