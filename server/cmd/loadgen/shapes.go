package main

import (
	"context"
	"io"
	"time"

	"google.golang.org/grpc/status"
	proto "toolplane/proto"
)

// runShape drives the consumer side of the selected workload. It spawns
// the shape's workers and returns only when every worker has drained —
// workers track in their own wait group precisely so the providers stay
// alive until all driven work finishes; stopping them earlier would turn
// every in-flight wait into a timeout.
func runShape(ctx context.Context, h *harness) {
	deadline := deadlineSignal(h)

	switch h.cfg.shape {
	case "agent-turn":
		h.agentTurnWorkers(ctx, deadline)
	case "token-stream":
		h.tokenStreamWorkers(ctx, deadline)
	case "heartbeat-fleet":
		// Providers already heartbeat every 30s; the fleet shape is
		// simply more machines doing nothing else. Session setup
		// registered them; idle here until the deadline.
		<-deadline
	case "discovery-churn":
		h.discoveryChurnWorkers(ctx, deadline)
	case "mixed":
		go h.agentTurnWorkers(ctx, deadline)
		go h.tokenStreamWorkers(ctx, deadline)
		h.discoveryChurnWorkers(ctx, deadline)
	}
	h.workerWG.Wait()
}

// partialOrError keeps window-edge cancellations out of the failure
// count: an op cut short by the deadline is a partial, not a fault.
func (h *harness) partialOrError(name string, durationMS float64, code string, ctx context.Context) {
	if code != "" && h.pastDeadline() && ctx.Err() != nil {
		h.collector.recordOp(name+"-partial", durationMS, "")
		return
	}
	h.collector.recordOp(name, durationMS, code)
}

// agentTurnWorkers is the canonical agent loop per session: discover
// tools, fire a burst of synchronous invokes, repeat.
func (h *harness) agentTurnWorkers(ctx context.Context, deadline <-chan struct{}) {
	for _, sessionID := range h.sessions {
		h.workerWG.Add(1)
		go func(sessionID string) {
			defer h.workerWG.Done()
			for {
				select {
				case <-deadline:
					return
				case <-ctx.Done():
					return
				default:
				}
				h.listTools(ctx, sessionID)
				for i := 0; i < h.cfg.burst; i++ {
					select {
					case <-deadline:
						return
					case <-ctx.Done():
						return
					default:
					}
					h.invokeEcho(ctx, sessionID)
				}
			}
		}(sessionID)
	}
}

// tokenStreamWorkers follows long streams end to end — the model-leg
// shape: a minute-scale request whose chunks arrive continuously.
func (h *harness) tokenStreamWorkers(ctx context.Context, deadline <-chan struct{}) {
	for i := 0; i < h.cfg.streamCount; i++ {
		sessionID := h.sessions[i%len(h.sessions)]
		h.workerWG.Add(1)
		go func(sessionID string) {
			defer h.workerWG.Done()
			for {
				select {
				case <-deadline:
					return
				case <-ctx.Done():
					return
				default:
				}
				h.followStream(ctx, sessionID)
			}
		}(sessionID)
	}
}

// discoveryChurnWorkers keeps ListTools warm across sessions.
func (h *harness) discoveryChurnWorkers(ctx context.Context, deadline <-chan struct{}) {
	if len(h.sessions) == 0 {
		return
	}
	h.workerWG.Add(1)
	go func() {
		defer h.workerWG.Done()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		index := 0
		for {
			select {
			case <-deadline:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				h.listTools(ctx, h.sessions[index%len(h.sessions)])
				index++
			}
		}
	}()
}

// consumerTools is the client consumer ops ride — replica B in the
// multi-instance drill, the primary connection otherwise.
func (h *harness) consumerTools() proto.ToolServiceClient {
	if h.altTools != nil {
		return h.altTools
	}
	return h.tools
}

func (h *harness) listTools(ctx context.Context, sessionID string) {
	callCtx, cancel := context.WithTimeout(withAPIKey(ctx, h.cfg.apiKey), 10*time.Second)
	start := time.Now()
	_, err := h.consumerTools().ListTools(callCtx, &proto.ListToolsRequest{SessionId: sessionID})
	cancel()
	h.collector.recordOp("list-tools", milliseconds(start), errorCode(err))
}

func (h *harness) invokeEcho(ctx context.Context, sessionID string) {
	callCtx, cancel := context.WithTimeout(withAPIKey(ctx, h.cfg.apiKey), 60*time.Second)
	start := time.Now()
	resp, err := h.consumerTools().InvokeTool(callCtx, &proto.ExecuteToolRequest{
		SessionId:          sessionID,
		ToolName:           echoTool,
		Input:              `{}`,
		WaitTimeoutSeconds: 30,
	})
	cancel()
	if err != nil && h.refusalExpected &&
		(status.Code(err).String() == codeNameFailedPrecondition || status.Code(err).String() == codeNameNotFound) {
		// The designed refusal once a session's provider is drained
		// away: no-provider FAILED_PRECONDITION while the machine is
		// draining, NOT_FOUND once the drained machine took its tools
		// with it. Counted separately, never a fault.
		h.collector.recordOp("invoke-refused", milliseconds(start), "")
		return
	}
	h.partialOrError("invoke", milliseconds(start), errorCode(err), ctx)
	if err == nil && resp.GetStatus() != proto.RequestStatus_REQUEST_STATUS_DONE {
		// The wait expired before the request finished — a dispatch-speed
		// signal for the report, not an RPC failure.
		h.collector.recordOp("invoke-nondone", milliseconds(start), "")
	}
}

// followStream subscribes to StreamExecuteTool and consumes chunks until
// the terminal marker, recording stream duration and chunk count.
func (h *harness) followStream(ctx context.Context, sessionID string) {
	callCtx, cancel := context.WithTimeout(withAPIKey(ctx, h.cfg.apiKey), 90*time.Second)
	defer cancel()
	start := time.Now()
	stream, err := h.tools.StreamExecuteTool(callCtx, &proto.ExecuteToolRequest{
		SessionId: sessionID,
		ToolName:  streamTool,
		Input:     `{}`,
	})
	if err != nil {
		h.partialOrError("stream", milliseconds(start), status.Code(err).String(), ctx)
		return
	}
	chunks := 0
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			h.collector.recordOp("stream", milliseconds(start), "")
			h.collector.recordChunkCount(chunks)
			return
		}
		if err != nil {
			h.partialOrError("stream", milliseconds(start), status.Code(err).String(), ctx)
			h.collector.recordChunkCount(chunks)
			return
		}
		chunks++
	}
}

// deadlineSignal returns a channel that closes once at the load window's
// end — closed, not fired once, because every worker goroutine selects
// on it and a single-value channel would wake only the first.
func deadlineSignal(h *harness) <-chan struct{} {
	done := make(chan struct{})
	time.AfterFunc(time.Until(h.deadlineAt), func() { close(done) })
	return done
}
