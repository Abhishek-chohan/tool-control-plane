package service

import (
	"context"
	"testing"

	"toolplane/pkg/model"
	"toolplane/pkg/storage"
	"toolplane/pkg/storage/memory"
)

// runListDeterminism pins page determinism for same-tick creations: the
// CreatedAt+ID tiebreak must order identical timestamps identically on
// every page and on both storage modes — the unstable sort previously
// reshuffled same-tick rows between pages (map iteration order decided).
func runListDeterminism(t *testing.T, store storage.Storer) {
	t.Helper()
	tracer := &recordingTracer{}
	toolService := NewToolService(tracer, store)
	machineService := NewMachinesService(context.Background(), toolService, tracer, store)
	requestService := NewRequestsService(context.Background(), toolService, machineService, tracer, store)

	const sessionID = "session-g7-determinism"
	const machineID = "machine-g7"
	if _, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine: %v", err)
	}

	// Create a batch of requests that share CreatedAt as closely as the
	// clock allows — the tiebreak has to make order independent of that.
	const total = 12
	for i := 0; i < total; i++ {
		if _, err := requestService.CreateRequest(sessionID, "echo", `{}`, 0, ""); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}

	// Walk two pages and require a deterministic global order: sorted by
	// (CreatedAt, ID), page boundaries included.
	pageSize := 5
	var walked []string
	for offset := 0; offset < total; offset += pageSize {
		page, _, err := requestService.ListRequests(sessionID, "", "", pageSize, offset)
		if err != nil {
			t.Fatalf("list offset %d: %v", offset, err)
		}
		for _, req := range page {
			walked = append(walked, req.ID)
		}
	}
	if len(walked) != total {
		t.Fatalf("walked %d requests, want %d", len(walked), total)
	}

	// The order must equal the sort the service documents: CreatedAt, then
	// ID ascending. Re-derive it from a full listing and compare — and do
	// it twice to catch map-iteration nondeterminism.
	full, _, err := requestService.ListRequests(sessionID, "", "", total, 0)
	if err != nil {
		t.Fatalf("full list: %v", err)
	}
	for i := range full {
		if full[i].ID != walked[i] {
			t.Fatalf("page walk diverged from full listing at %d: %s vs %s (same-tick order must be stable)", i, walked[i], full[i].ID)
		}
	}
	again, _, err := requestService.ListRequests(sessionID, "", "", total, 0)
	if err != nil {
		t.Fatalf("second full list: %v", err)
	}
	for i := range again {
		if again[i].ID != full[i].ID {
			t.Fatalf("listing order changed between identical calls at %d: %s vs %s", i, full[i].ID, again[i].ID)
		}
	}
}

func TestListRequestsDeterministicCacheOnly(t *testing.T) {
	runListDeterminism(t, nil)
}

func TestListRequestsDeterministicMemoryStore(t *testing.T) {
	runListDeterminism(t, memory.New())
}
