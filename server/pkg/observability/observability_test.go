package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"toolplane/pkg/model"
	"toolplane/pkg/storage/memory"
	"toolplane/pkg/trace"
)

func TestGRPCInterceptorRecordsMethodCodeAndDuration(t *testing.T) {
	collector := NewRuntimeMetricsCollector()
	unary := collector.UnaryServerInterceptor()

	info := &grpc.UnaryServerInfo{FullMethod: "/api.ToolService/ExecuteTool"}
	if _, err := unary(context.Background(), nil, info, func(context.Context, interface{}) (interface{}, error) {
		return nil, nil
	}); err != nil {
		t.Fatalf("ok handler propagated error: %v", err)
	}
	if _, err := unary(context.Background(), nil, info, func(context.Context, interface{}) (interface{}, error) {
		return nil, status.Error(codes.NotFound, "missing")
	}); err == nil {
		t.Fatal("error handler returned nil error")
	}

	body := scrapeMetrics(t, collector)
	for _, fragment := range []string{
		`toolplane_grpc_requests_total{code="OK",method="/api.ToolService/ExecuteTool"} 1`,
		`toolplane_grpc_requests_total{code="NotFound",method="/api.ToolService/ExecuteTool"} 1`,
		`toolplane_grpc_request_duration_seconds_bucket{method="/api.ToolService/ExecuteTool",le="`,
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("metrics output missing %q", fragment)
		}
	}
}

func TestGRPCInterceptorSurfacesTraceparentOnFailureLog(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	defer slog.SetDefault(prev)

	collector := NewRuntimeMetricsCollector()
	unary := collector.UnaryServerInterceptor()
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"))
	info := &grpc.UnaryServerInfo{FullMethod: "/api.SessionsService/GetSession"}

	_, err := unary(ctx, nil, info, func(context.Context, interface{}) (interface{}, error) {
		return nil, status.Error(codes.PermissionDenied, "denied")
	})
	if err == nil {
		t.Fatal("expected handler error")
	}

	line := buf.String()
	if !strings.Contains(line, `"traceparent":"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"`) {
		t.Fatalf("request log missing traceparent: %s", line)
	}
	if !strings.Contains(line, `"code":"PermissionDenied"`) {
		t.Fatalf("request log missing status code: %s", line)
	}
}

func TestAuditRecorderPersistsFilteredEvents(t *testing.T) {
	store := memory.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	recorder := NewAuditRecorder(ctx, store)

	recorder.Record(trace.SessionEvent{Event: trace.EventSessionCreated, SessionID: "s1"})
	recorder.Record(trace.SessionEvent{Event: trace.EventAPIKeyRevoked, SessionID: "s1"})
	recorder.Record(trace.SessionEvent{Event: trace.EventRequestClaimed, SessionID: "s1"}) // high-frequency: excluded
	recorder.Record(trace.SessionEvent{Event: trace.EventTaskDeadLettered, SessionID: "s1", TaskID: "t1"})

	// Writes are asynchronous (the recorder must never block its caller);
	// wait for the three durable rows to land.
	var events []*model.AuditEvent
	deadline := time.Now().Add(2 * time.Second)
	for {
		events = store.AuditEvents()
		if len(events) == 3 || time.Now().After(deadline) {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if len(events) != 3 {
		t.Fatalf("retained %d audit events, want 3 (high-frequency excluded)", len(events))
	}
	seen := map[string]bool{}
	for _, ev := range events {
		seen[ev.Event] = true
	}
	if !seen["session_created"] || !seen["api_key_revoked"] || !seen["task_dead_lettered"] {
		t.Fatalf("unexpected audit set: %+v", seen)
	}
	if seen["request_claimed"] {
		t.Fatal("high-frequency execution event was audited")
	}
}

func TestAuditRecorderNilStoreIsNoop(t *testing.T) {
	recorder := NewAuditRecorder(context.Background(), nil)
	recorder.Record(trace.SessionEvent{Event: trace.EventSessionCreated}) // must not panic
}

func TestNewLoggerJSONEmitsStructuredRecords(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("json", &buf)
	logger.Info("hello", "key", "value")

	var record map[string]any
	if err := json.Unmarshal(buf.Bytes(), &record); err != nil {
		t.Fatalf("JSON log line not parseable: %v (%q)", err, buf.String())
	}
	if record["msg"] != "hello" || record["key"] != "value" {
		t.Fatalf("JSON record missing fields: %+v", record)
	}
}

func TestRouteStdLogBridgesThroughSlog(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("json", &buf)
	slog.SetDefault(logger)
	defer slog.SetDefault(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))

	RouteStdLog(logger)
	prevFlags := log.Flags()
	defer log.SetFlags(prevFlags)

	log.Printf("bridged message")

	var record map[string]any
	if err := json.Unmarshal(buf.Bytes(), &record); err != nil {
		t.Fatalf("bridged line not JSON: %v (%q)", err, buf.String())
	}
	if record["msg"] != "bridged message" {
		t.Fatalf("bridged record lost the message: %+v", record)
	}
}

func TestStdLogWriterDropsBlankLines(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("json", &buf)
	w := &stdLogWriter{logger: logger}
	if _, err := w.Write([]byte("\n")); err != nil {
		t.Fatalf("blank write errored: %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("blank line produced output: %q", buf.String())
	}
}

func scrapeMetrics(t *testing.T, collector *RuntimeMetricsCollector) string {
	t.Helper()
	recorder := httptest.NewRecorder()
	collector.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("metrics status = %d", recorder.Code)
	}
	return recorder.Body.String()
}
