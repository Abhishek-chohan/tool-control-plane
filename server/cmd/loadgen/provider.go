package main

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	proto "toolplane/proto"
)

// runFakeProvider is the provider side of the load: it speaks the same
// RPC loop a real provider process speaks — claim, stream, resolve,
// heartbeat — so the load includes the dispatch path, not just the
// ingress path. machineStop lets a drill kill exactly this provider
// (h.stop kills the whole fleet).
func (h *harness) runFakeProvider(sessionID, machineID string, machineStop chan struct{}) {
	defer h.wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		select {
		case <-h.stop:
			cancel()
		case <-machineStop:
			cancel()
		case <-ctx.Done():
		}
	}()

	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			pingCtx, pingCancel := context.WithTimeout(withAPIKey(ctx, h.cfg.apiKey), 5*time.Second)
			start := time.Now()
			_, err := h.mach.UpdateMachinePing(pingCtx, &proto.UpdateMachinePingRequest{
				SessionId: sessionID, MachineId: machineID,
			})
			pingCancel()
			code := errorCode(err)
			if ctx.Err() != nil {
				code = ""
			}
			h.collector.recordOp("heartbeat", milliseconds(start), code)
			continue
		default:
		}

		start := time.Now()
		var resp *proto.ClaimNextRequestResponse
		err := h.withRetry(ctx, func() error {
			callCtx, callCancel := context.WithTimeout(withAPIKey(ctx, h.cfg.apiKey), 60*time.Second)
			defer callCancel()
			var callErr error
			resp, callErr = h.reqs.ClaimNextRequest(callCtx, &proto.ClaimNextRequestRequest{
				SessionId: sessionID,
				MachineId: machineID,
				ToolNames: []string{echoTool, streamTool},
			})
			return callErr
		})
		if err != nil {
			// FAILED_PRECONDITION on the poll is the draining refusal —
			// this machine is retired; stop claiming like a real
			// provider would. Anything else is a driver-visible fault,
			// unless the run window closed under it.
			if status.Code(err) == codes.FailedPrecondition {
				return
			}
			code := status.Code(err).String()
			if ctx.Err() != nil {
				code = ""
			}
			h.collector.recordOp("claim-poll", milliseconds(start), code)
			select {
			case <-ctx.Done():
				return
			case <-time.After(idlePollInterval):
			}
			continue
		}
		h.collector.recordOp("claim-poll", milliseconds(start), "")

		request := resp.GetRequest()
		if !resp.GetClaimed() || request == nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(idlePollInterval):
			}
			continue
		}
		switch request.GetToolName() {
		case streamTool:
			h.executeStreamingTool(ctx, sessionID, machineID, request)
		default:
			h.resolveRequest(ctx, sessionID, machineID, request, `{"loadgen":"done"}`, "resolution")
		}
	}
}

// withRetry retries UNAVAILABLE calls the way real provider SDKs do —
// UNAVAILABLE is the retryable code (serialization-retry exhaustion
// among others), and without the retry the fake provider would count
// transient contention as faults the real fleet absorbs. Three attempts
// with linear backoff; anything still failing returns to the caller.
func (h *harness) withRetry(ctx context.Context, call func() error) error {
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		err = call()
		if err == nil || status.Code(err) != codes.Unavailable || ctx.Err() != nil {
			return err
		}
		time.Sleep(time.Duration(50*(attempt+1)) * time.Millisecond)
	}
	return err
}

// executeStreamingTool emits the token-stream shape on the provider
// side: 1KiB chunks at 10/s for streamSeconds, then a resolution. The
// consumer-side follower is the stream shape in shapes.go.
func (h *harness) executeStreamingTool(ctx context.Context, sessionID, machineID string, request *proto.Request) {
	const streamSeconds = 30
	chunk := strings.Repeat("t", 1024)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.After(streamSeconds * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-deadline:
			h.resolveRequest(ctx, sessionID, machineID, request, `{"loadgen":"streamed"}`, "resolution")
			return
		case <-ticker.C:
			start := time.Now()
			err := h.withRetry(ctx, func() error {
				appendCtx, cancel := context.WithTimeout(withAPIKey(ctx, h.cfg.apiKey), 10*time.Second)
				defer cancel()
				_, appendErr := h.reqs.AppendRequestChunks(appendCtx, &proto.AppendRequestChunksRequest{
					SessionId:  sessionID,
					RequestId:  request.GetId(),
					Chunks:     []string{chunk},
					ResultType: "streaming",
					MachineId:  machineID,
					LeaseEpoch: request.GetLeaseEpoch(),
				})
				return appendErr
			})
			code := errorCode(err)
			if ctx.Err() != nil {
				code = ""
			}
			h.collector.recordOp("append-chunk", milliseconds(start), code)
		}
	}
}

func (h *harness) resolveRequest(ctx context.Context, sessionID, machineID string, request *proto.Request, result, resultType string) {
	start := time.Now()
	err := h.withRetry(ctx, func() error {
		resolveCtx, cancel := context.WithTimeout(withAPIKey(ctx, h.cfg.apiKey), 10*time.Second)
		defer cancel()
		_, resolveErr := h.reqs.SubmitRequestResult(resolveCtx, &proto.SubmitRequestResultRequest{
			SessionId:  sessionID,
			RequestId:  request.GetId(),
			Result:     result,
			ResultType: resultType,
			MachineId:  machineID,
			LeaseEpoch: request.GetLeaseEpoch(),
		})
		return resolveErr
	})
	code := errorCode(err)
	if ctx.Err() != nil {
		code = ""
	}
	h.collector.recordOp("resolve", milliseconds(start), code)
}

func withAPIKey(ctx context.Context, key string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "api_key", key)
}

func errorCode(err error) string {
	if err == nil {
		return ""
	}
	return status.Code(err).String()
}

func milliseconds(start time.Time) float64 {
	return float64(time.Since(start).Microseconds()) / 1000.0
}
