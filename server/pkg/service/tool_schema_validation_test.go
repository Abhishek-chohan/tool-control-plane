package service

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toolplane/pkg/model"
)

// TestRegisterToolSchemaValidation pins the registration bar: a non-empty
// schema must be well-formed JSON with an object root; garbage is
// INVALID_ARGUMENT before any state is touched. Empty schemas stay
// allowed (consumers supply their own fallback), and semantically wrong
// but structurally valid schemas are accepted — validation is validity,
// not a JSON-Schema engine.
func TestRegisterToolSchemaValidation(t *testing.T) {
	tracer := &recordingTracer{}
	toolService := NewToolService(tracer, nil)

	const sessionID = "session-g8-schema"
	const machineID = "machine-g8-schema"

	cases := []struct {
		name    string
		schema  string
		wantErr bool
	}{
		{"empty allowed", "", false},
		{"object allowed", `{"type":"object"}`, false},
		{"deeply wrong but valid JSON object allowed", `{"type":"unicorn","required":42}`, false},
		{"malformed JSON rejected", `{"type":"object"`, true},
		{"non-JSON text rejected", `not json at all`, true},
		{"array root rejected", `["type"]`, true},
		{"string root rejected", `"object"`, true},
		{"number root rejected", `42`, true},
		{"null root rejected", `null`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := toolService.RegisterTool(sessionID, machineID, "tool-"+tc.name, "d", tc.schema, nil, nil)
			if tc.wantErr {
				if !errors.Is(err, ErrInvalidArgument) {
					t.Fatalf("schema %q: err=%v, want ErrInvalidArgument", tc.schema, err)
				}
				if code := status.Code(statusFromDomainError("register tool", err)); code != codes.InvalidArgument {
					t.Fatalf("schema %q maps to %v, want InvalidArgument", tc.schema, code)
				}
				return
			}
			if err != nil {
				t.Fatalf("schema %q rejected: %v", tc.schema, err)
			}
		})
	}
}

// TestRegisterMachineRejectsBadToolSchema: machine registration fails
// fast on an invalid tool schema instead of silently skipping the tool in
// the reconcile (caller input errors are not best-effort).
func TestRegisterMachineRejectsBadToolSchema(t *testing.T) {
	tracer := &recordingTracer{}
	toolService := NewToolService(tracer, nil)
	machineService := NewMachinesService(context.Background(), toolService, tracer, nil)

	_, err := machineService.RegisterMachine("session-g8-machine", "machine-g8", "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool("session-g8-machine", "machine-g8", "good", "d", `{"type":"object"}`, nil, nil),
		model.NewTool("session-g8-machine", "machine-g8", "bad", "d", `not json`, nil, nil),
	}, "")
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("register machine with bad schema: err=%v, want ErrInvalidArgument", err)
	}
	if _, err := toolService.GetToolByName("session-g8-machine", "good"); err == nil {
		t.Fatal("the valid tool must not be registered when the batch is rejected")
	}
}
