package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	insecurecreds "google.golang.org/grpc/credentials/insecure"
	proto "toolplane/proto"
)

// Reliability drills under load: the failure semantics the idle drill
// matrix proves (lease expiry requeue, drain completing in-flight work,
// no double-dispatch across replicas) exercised while the control plane
// is saturated. Each drill runs one shape of load, injects its failure
// mid-run, and asserts the invariants held — the report carries the
// assertions and the exit code reflects them.

const codeNameFailedPrecondition = "FailedPrecondition"
const codeNameNotFound = "NotFound"

type drillAssertion struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
}

type drillResult struct {
	Name       string           `json:"name"`
	Assertions []drillAssertion `json:"assertions"`
}

func (d *drillResult) failed() bool {
	for _, a := range d.Assertions {
		if !a.OK {
			return true
		}
	}
	return false
}

func (d *drillResult) add(name string, ok bool, format string, args ...interface{}) {
	d.Assertions = append(d.Assertions, drillAssertion{Name: name, OK: ok, Detail: fmt.Sprintf(format, args...)})
}

// finishDrillMetrics freezes the counter deltas after the drill's
// providers have drained, so assertions read settled numbers.
func (h *harness) finishDrillMetrics() {
	h.drillMetrics = metricsDelta(h.startMetrics, scrapeMetrics(h.metricsURL))
}

func counterDelta(metrics map[string]float64, name string) float64 {
	for key, value := range metrics {
		if key == name || strings.HasPrefix(key, name+"{") {
			return value
		}
	}
	return 0
}

// runDrill drives the configured drill; each drill manages its own
// workers and provider shutdown.
func runDrill(ctx context.Context, h *harness) *drillResult {
	result := &drillResult{Name: h.cfg.drill}

	switch h.cfg.drill {
	case "provider-kill":
		drillProviderKill(ctx, h, result)
	case "drain-under-backlog":
		drillDrainUnderBacklog(ctx, h, result)
	case "multi-instance-contention":
		drillMultiInstanceContention(ctx, h, result)
	default:
		fmt.Fprintf(os.Stderr, "loadgen: unknown drill %q\n", h.cfg.drill)
		os.Exit(2)
	}
	return result
}

// waitForRequestState polls a request until cond holds or the tries run
// out; drills use it to observe transitions the load itself does not
// wait on.
func (h *harness) waitForRequestState(requestID, sessionID string, cond func(*proto.Request) bool, tries int, interval time.Duration) bool {
	for i := 0; i < tries; i++ {
		getCtx, getCancel := context.WithTimeout(withAPIKey(context.Background(), h.cfg.apiKey), 2*time.Second)
		current, err := h.reqs.GetRequest(getCtx, &proto.GetRequestRequest{
			SessionId: sessionID, RequestId: requestID,
		})
		getCancel()
		if err != nil {
			return false
		}
		if cond(current) {
			return true
		}
		time.Sleep(interval)
	}
	return false
}

// drillProviderKill saturates with agent turns, then makes one provider
// die while HOLDING a claimed request: the orphaned claim must requeue
// after lease expiry and finish on the replacement machine. The
// replacement's registration retries until the dead machine's row is
// gone, because tool ownership is exclusive while the old owner is
// still listed. Needs a window of ~45s: the 30s lease must expire
// before the requeue can happen.
func drillProviderKill(ctx context.Context, h *harness, result *drillResult) {
	victim := h.machines[0]
	h.agentTurnWorkers(ctx, deadlineSignal(h))

	time.Sleep(3 * time.Second)

	// Hold one request on the victim: create without a wait, let the
	// victim claim it, then kill the goroutine mid-execution. The
	// streaming tool is the held one — its 30s execution makes the
	// claimed state observable before the kill.
	createCtx, createCancel := context.WithTimeout(withAPIKey(ctx, h.cfg.apiKey), 10*time.Second)
	held, holdErr := h.reqs.CreateRequest(createCtx, &proto.CreateRequestRequest{
		SessionId: victim.sessionID, ToolName: streamTool, Input: `{}`,
	})
	createCancel()
	result.add("held request created", holdErr == nil, "create: %v", holdErr)
	if holdErr == nil {
		claimed := h.waitForRequestState(held.GetId(), victim.sessionID,
			func(r *proto.Request) bool { return r.GetLeasedBy() == victim.machineID }, 100, 50*time.Millisecond)
		result.add("victim held the request", claimed, "claim observed before the kill")
	}

	// Process death: stop the loop, then retire the machine row so the
	// replacement can take the tool names. Unregister blocks until the
	// orphaned claim is requeued away — the drain-like semantics under
	// test — so it runs concurrently with the rest of the window.
	if stop, ok := h.providerStops[victim.machineID]; ok {
		close(stop)
	}
	unregisterDone := make(chan error, 1)
	go func() {
		unregisterCtx, unregisterCancel := context.WithTimeout(withAPIKey(context.Background(), h.cfg.apiKey), 2*time.Minute)
		_, err := h.mach.UnregisterMachine(unregisterCtx, &proto.UnregisterMachineRequest{
			SessionId: victim.sessionID, MachineId: victim.machineID,
		})
		unregisterCancel()
		unregisterDone <- err
	}()

	replacementID := victim.machineID + "-replacement"
	registered := false
	for attempt := 0; attempt < 240 && !registered; attempt++ {
		time.Sleep(500 * time.Millisecond)
		registerCtx, registerCancel := context.WithTimeout(withAPIKey(ctx, h.cfg.apiKey), 5*time.Second)
		_, registerErr := h.mach.RegisterMachine(registerCtx, &proto.RegisterMachineRequest{
			SessionId:   victim.sessionID,
			MachineId:   replacementID,
			SdkVersion:  "loadgen",
			SdkLanguage: "go",
			Tools: []*proto.RegisterToolRequest{
				{Name: echoTool, Description: "resolves immediately", Schema: `{"type":"object"}`},
				{Name: streamTool, Description: "streams chunks then resolves", Schema: `{"type":"object"}`},
			},
		})
		registerCancel()
		if registerErr == nil {
			registered = true
		}
		if ctx.Err() != nil {
			break
		}
	}
	result.add("replacement registered", registered, "after the dead owner released the tool names")
	if registered {
		h.wg.Add(1)
		machineStop := make(chan struct{})
		h.providerStops[replacementID] = machineStop
		go h.runFakeProvider(victim.sessionID, replacementID, machineStop)
	}

	h.workerWG.Wait()
	h.shutdown()
	unregisterErr := <-unregisterDone
	h.finishDrillMetrics()

	result.add("dead machine unregistered", unregisterErr == nil, "unregister: %v (returns after the orphan requeues)", unregisterErr)

	heldDone := false
	if holdErr == nil {
		heldDone = h.waitForRequestState(held.GetId(), victim.sessionID,
			func(r *proto.Request) bool { return r.GetStatus() == proto.RequestStatus_REQUEST_STATUS_DONE }, 400, 200*time.Millisecond)
	}
	result.add("held request finished on replacement", heldDone, "terminal DONE observed")

	// The requeue counter only tells the whole story on the store-backed
	// reaper path; the memory store reclaims expired leases lazily at
	// claim time, where the outcome above is the proof.
	requeues := counterDelta(h.drillMetrics, "toolplane_request_requeues_total")
	deadLetters := counterDelta(h.drillMetrics, "toolplane_request_dead_letters_total")
	if strings.TrimSpace(os.Getenv("TOOLPLANE_DATABASE_URL")) != "" {
		result.add("orphaned work requeued", requeues >= 1, "requeues=%.0f (want >=1: the store-backed reaper must requeue the expired claim)", requeues)
	} else {
		result.add("orphaned work recovered", heldDone, "memory store reclaims lazily at claim time; recovery observed via the held request")
	}
	result.add("no dead letters", deadLetters == 0, "dead_letters=%.0f (want 0: the replacement provider recovers everything)", deadLetters)
}

// drillDrainUnderBacklog saturates with agent turns and drains one
// provider mid-run. The drain RPC returns only after that machine's
// in-flight work completed; invokes refused afterwards (no provider)
// are the designed behavior and are counted as refusals, not faults.
func drillDrainUnderBacklog(ctx context.Context, h *harness, result *drillResult) {
	target := h.machines[0]
	h.refusalExpected = true
	h.agentTurnWorkers(ctx, deadlineSignal(h))

	time.Sleep(5 * time.Second)
	drainCtx, drainCancel := context.WithTimeout(withAPIKey(ctx, h.cfg.apiKey), 60*time.Second)
	drainStart := time.Now()
	drainResp, drainErr := h.mach.DrainMachine(drainCtx, &proto.DrainMachineRequest{
		SessionId: target.sessionID, MachineId: target.machineID,
	})
	drainCancel()
	result.add("drain rpc succeeded", drainErr == nil, "drain: %v", drainErr)
	if drainErr == nil {
		result.add("drain reported drained", drainResp.GetDrained(), "drained=%v", drainResp.GetDrained())
		result.add("drain bounded", time.Since(drainStart) < 45*time.Second, "drain took %s", time.Since(drainStart).Round(time.Millisecond))
	}
	if stop, ok := h.providerStops[target.machineID]; ok {
		close(stop) // the drained machine's loop would only idle now
	}

	// The claim-blocking semantics are the idle drill matrix's ground
	// (TestMachinesServiceDrainMachineWaitsForInflightRequestAndBlocksNewWork
	// and friends); under load this drill proves the RPC completes
	// while saturated, the in-flight work it waited for is intact, and
	// nothing dead-lettered.
	h.workerWG.Wait()
	h.shutdown()
	h.finishDrillMetrics()

	deadLetters := counterDelta(h.drillMetrics, "toolplane_request_dead_letters_total")
	result.add("no dead letters", deadLetters == 0, "dead_letters=%.0f", deadLetters)
}

// drillMultiInstanceContention runs consumers against a second server
// replica sharing the same store while providers serve from the first —
// the active-active shape. Every invoke must complete with no
// serialization exhaustion. Postgres only: two replicas need shared
// durable state.
func drillMultiInstanceContention(ctx context.Context, h *harness, result *drillResult) {
	if strings.TrimSpace(os.Getenv("TOOLPLANE_DATABASE_URL")) == "" {
		result.add("skipped without shared storage", true, "multi-instance contention requires TOOLPLANE_DATABASE_URL (the two replicas share it); skipped")
		return
	}

	addrB, _, shutdownB := bootEmbeddedServer()
	defer shutdownB()
	connB, err := grpc.NewClient(addrB, grpc.WithTransportCredentials(insecurecreds.NewCredentials()))
	if err != nil {
		result.add("dial replica B", false, "dial: %v", err)
		return
	}
	defer connB.Close()
	h.altTools = proto.NewToolServiceClient(connB)

	h.agentTurnWorkers(ctx, deadlineSignal(h))

	h.workerWG.Wait()
	h.shutdown()
	h.finishDrillMetrics()

	exhausted := counterDelta(h.drillMetrics, "toolplane_storage_serialization_exhausted_total")
	result.add("no serialization exhaustion", exhausted == 0, "exhausted=%.0f (want 0 under cross-replica contention)", exhausted)
}
