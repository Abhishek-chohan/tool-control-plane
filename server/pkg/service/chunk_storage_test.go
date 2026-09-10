package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"toolplane/pkg/model"
	"toolplane/pkg/storage/memory"
	"toolplane/pkg/trace"
)

func newChunkPair(t *testing.T) (*memory.Store, *MachinesService, *RequestsService, *RequestsService) {
	t.Helper()
	store := memory.New()
	toolSvc := NewToolService(trace.NopTracer(), store)
	machineSvc := NewMachinesService(context.Background(), toolSvc, trace.NopTracer(), store)
	// Distinct RequestsService instances share the store (and the tool/
	// machine services for setup) but keep per-instance request caches —
	// exactly the replica shape the window read-through targets.
	svcA := NewRequestsService(context.Background(), toolSvc, machineSvc, trace.NopTracer(), store)
	svcB := NewRequestsService(context.Background(), toolSvc, machineSvc, trace.NopTracer(), store)
	return store, machineSvc, svcA, svcB
}

func registerEchoForChunkTests(t *testing.T, machineSvc *MachinesService, sessionID, machineID string) {
	t.Helper()
	if _, err := machineSvc.RegisterMachine(sessionID, machineID, "1.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	}, ""); err != nil {
		t.Fatalf("register machine: %v", err)
	}
}

func TestChunkWindowVisibleAcrossReplicas(t *testing.T) {
	_, machineSvc, svcA, svcB := newChunkPair(t)

	const sessionID = "sess-chunk-aa"
	const machineID = "machine-chunk-aa"
	registerEchoForChunkTests(t, machineSvc, sessionID, machineID)

	request, err := svcA.CreateRequest(sessionID, "echo", `{}`, 0, "")
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	claimed, err := svcA.ClaimRequest(sessionID, request.ID, machineID)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}

	// Instance A appends chunks (fenced, table-backed).
	if err := svcA.AppendRequestChunks(sessionID, request.ID, machineID, claimed.LeaseEpoch, []string{"alpha", "beta"}, model.ResultTypeStreaming); err != nil {
		t.Fatalf("append on A: %v", err)
	}

	// Instance B has never mirrored these chunks: its window read must come
	// from the store.
	snapshot, err := svcB.GetRequestReplayStreamAnySession(request.ID, 0)
	if err != nil {
		t.Fatalf("window read on B: %v", err)
	}
	if len(snapshot.Window.Chunks) != 2 || snapshot.Window.Chunks[0] != "alpha" {
		t.Fatalf("replica B window = %v, want [alpha beta]", snapshot.Window.Chunks)
	}
	if snapshot.Window.NextSeq != 3 {
		t.Fatalf("replica B next seq = %d, want 3", snapshot.Window.NextSeq)
	}
}

func TestAppendRequestChunksRejectsOversizedChunk(t *testing.T) {
	store := memory.New()
	toolSvc := NewToolService(trace.NopTracer(), store)
	machineSvc := NewMachinesService(context.Background(), toolSvc, trace.NopTracer(), store)
	requestsSvc := NewRequestsService(context.Background(), toolSvc, machineSvc, trace.NopTracer(), store)

	const sessionID = "sess-chunk-cap"
	const machineID = "machine-chunk-cap"
	registerEchoForChunkTests(t, machineSvc, sessionID, machineID)

	request, err := requestsSvc.CreateRequest(sessionID, "echo", `{}`, 0, "")
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	claimed, err := requestsSvc.ClaimRequest(sessionID, request.ID, machineID)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}

	oversized := strings.Repeat("x", model.MaxRequestChunkBytes+1)
	err = requestsSvc.AppendRequestChunks(sessionID, request.ID, machineID, claimed.LeaseEpoch, []string{oversized}, model.ResultTypeStreaming)
	if err == nil {
		t.Fatal("oversized chunk accepted")
	}
	requestAfter, getErr := requestsSvc.GetRequestByID(sessionID, request.ID)
	if getErr != nil {
		t.Fatalf("get request after rejection: %v", getErr)
	}
	if requestAfter.NextStreamSeq != 1 {
		t.Fatalf("rejected append mutated the window: next seq = %d", requestAfter.NextStreamSeq)
	}
}

func TestSignalMapReleasedOnTerminalUpdate(t *testing.T) {
	store := memory.New()
	toolSvc := NewToolService(trace.NopTracer(), store)
	machineSvc := NewMachinesService(context.Background(), toolSvc, trace.NopTracer(), store)
	requestsSvc := NewRequestsService(context.Background(), toolSvc, machineSvc, trace.NopTracer(), store)

	const sessionID = "sess-chunk-signal"
	const machineID = "machine-chunk-signal"
	registerEchoForChunkTests(t, machineSvc, sessionID, machineID)

	request, err := requestsSvc.CreateRequest(sessionID, "echo", `{}`, 0, "")
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	claimed, err := requestsSvc.ClaimRequest(sessionID, request.ID, machineID)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}

	// Drive the request to completion through the fenced submit.
	if err := requestsSvc.SubmitRequestResult(sessionID, request.ID, machineID, claimed.LeaseEpoch, "done", model.ResultTypeResolution, nil); err != nil {
		t.Fatalf("submit result: %v", err)
	}

	if _, loaded := requestsSvc.signals.Load(request.ID); loaded {
		t.Fatal("signal entry survived a terminal update (map leak)")
	}
}

// TestStreamExecuteSignalWakeLatency is a timing sanity check for the
// signal-driven stream path: an append broadcast must wake a waiter well
// within the old 200ms poll interval budget.
func TestStreamExecuteSignalWakeLatency(t *testing.T) {
	store := memory.New()
	toolSvc := NewToolService(trace.NopTracer(), store)
	machineSvc := NewMachinesService(context.Background(), toolSvc, trace.NopTracer(), store)
	requestsSvc := NewRequestsService(context.Background(), toolSvc, machineSvc, trace.NopTracer(), store)

	const sessionID = "sess-signal"
	const machineID = "machine-signal"
	registerEchoForChunkTests(t, machineSvc, sessionID, machineID)

	request, err := requestsSvc.CreateRequest(sessionID, "echo", `{}`, 0, "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	watch := requestsSvc.subscribeRequest(request.ID)
	go func() {
		time.Sleep(20 * time.Millisecond)
		requestsSvc.notifyRequestUpdate(request.ID)
	}()

	select {
	case <-watch:
		// woken by the broadcast
	case <-time.After(200 * time.Millisecond):
		t.Fatal("signal wake took longer than the old poll interval")
	}
}
