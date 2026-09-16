package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"toolplane/pkg/model"
)

// TestAppendRequestChunksBatchTotalBound pins the domain-level batch guard:
// the summed chunk payload of one AppendRequestChunks RPC must stay under
// MaxChunkBatchPayloadBytes (the transport ceiling minus envelope headroom),
// so an oversized batch gets a clear INVALID_ARGUMENT instead of a
// transport ResourceExhausted on the wire.
func TestAppendRequestChunksBatchTotalBound(t *testing.T) {
	tracer := &recordingTracer{}
	toolService := NewToolService(tracer, nil)
	machineService := NewMachinesService(context.Background(), toolService, tracer, nil)
	requestService := NewRequestsService(context.Background(), toolService, machineService, tracer, nil)

	chunksPerBatch := int(model.MaxChunkBatchBytes / model.MaxRequestChunkBytes)
	fullBatch := make([]string, chunksPerBatch)
	for i := range fullBatch {
		fullBatch[i] = strings.Repeat("x", int(model.MaxRequestChunkBytes))
	}

	err := requestService.AppendRequestChunks("session-g6-batch", "request-g6", "machine-g6", 0, fullBatch, model.ResultTypeResolution)
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("full-batch append: err=%v, want ErrInvalidArgument (an exactly-full batch is over the transport ceiling once the envelope rides on top)", err)
	}

	// A batch under the payload bound passes the guard: the per-chunk sizes
	// are legal and the total fits, so the call proceeds to the normal
	// not-found rejection for the bogus request — not the payload guard.
	underBatch := fullBatch[:chunksPerBatch-1]
	err = requestService.AppendRequestChunks("session-g6-batch", "request-g6", "machine-g6", 0, underBatch, model.ResultTypeResolution)
	if errors.Is(err, ErrInvalidArgument) {
		t.Fatal("under-bound batch rejected by the payload guard: only totals above MaxChunkBatchPayloadBytes may be rejected there")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("under-bound batch with unknown request: err=%v, want ErrNotFound", err)
	}
}
