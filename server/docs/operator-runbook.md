# Operator Runbook

## Purpose

This document is the symptom-first day-2 runbook for Toolplane.

It is grounded in maintained signals only:

- `server/docs/observability.md` for stable traces, metrics, `/health`, throttle logs, and redaction rules
- `server/DOCUMENTATION.md` for request lifecycle, drain semantics, and request-inspection boundaries
- `server/docs/reference-deployment.md` for the rollout and drain order
- `server/docs/reliability-drills.md` for the named drill vocabulary used in failure diagnosis

## Supported Surfaces

### Observability

Use observability first when the question is about runtime state across many requests or machines.

- `session_trace` events with stable top-level `sessionId`, `machineId`, `toolId`, `requestId`, and `taskId`
- `/metrics` on the server for queue depth, in-flight work, retries, dead letters, active machines, draining machines, and machine load
- `/health` on the gateway for circuit and throttle state
- proxy throttle logs with `reason`, `retry_after`, `api_key_present`, and redacted `client_fingerprint`

### Request Inspection

Use request inspection when the question is about one request, one machine owner, or one retained replay window.

- `GetRequest` for `status`, `error`, `resultType`, `executingMachineId`, and retained `streamResults`
- `ListRequests` for session-scoped request sets filtered by `status` or `toolName`
- `GetRequestChunks` for `startSeq`, `nextSeq`, and the current retained chunk window
- `ListMachines` and `GetMachine` for session-scoped machine ownership checks

Replay-window state belongs here, not in the maintained observability contract.

### Control Actions

- `CancelRequest` stops one request without changing provider ownership globally.
- `DrainMachine` stops new routing to one provider while preserving in-flight completion behavior.

## Stable Correlation Model

- Start with `requestId` or `taskId` for a specific execution problem.
- Use `machineId` to connect request ownership, drain behavior, and lease-expiry churn.
- Use `sessionId` to follow related runtime activity across tools, requests, and machines.
- Use gateway throttle `reason`, `Retry-After`, and `client_fingerprint` for proxy-side incidents.

## Redaction Rules

- Do not expect raw API-key secrets in traces, metrics, `/health`, or logs.
- Do not expect raw client IPs in throttle logs; use `client_fingerprint` instead.
- Use `key_preview`, `tokenPreview`, or `keyId` only for correlation, never as secret material.

## First Response Checklist

1. Decide whether the incident is request-scoped, machine-scoped, or gateway-scoped.
2. Capture any known `requestId`, `taskId`, `machineId`, and `sessionId`.
3. Use observability to answer the first queue, retry, drain, throttle, or auth question.
4. Move to request inspection only when you need one request's owner, terminal state, or retained replay window.

## Common Inspection Calls

These examples use the maintained HTTP gateway surface and the same `ADMIN_API_KEY` pattern used in the reference deployment guide.

### Get One Request

```bash
curl -sS \
  "http://127.0.0.1:${TOOLPLANE_HTTP_PUBLISHED_PORT:-8080}/api/GetRequest" \
  -H "Authorization: Bearer ${ADMIN_API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{"sessionId":"your-session-id","requestId":"your-request-id"}' | python3 -m json.tool
```

Use `executingMachineId`, `status`, `error`, and `resultType` first.

### List Pending Or Running Requests

```bash
curl -sS \
  "http://127.0.0.1:${TOOLPLANE_HTTP_PUBLISHED_PORT:-8080}/api/ListRequests" \
  -H "Authorization: Bearer ${ADMIN_API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{"sessionId":"your-session-id","status":"running","limit":100,"offset":0}' | python3 -m json.tool
```

If you need all active requests for a draining machine, list `claimed` and `running` requests and then filter client-side by `executingMachineId`.

### Inspect The Retained Replay Window

```bash
curl -sS \
  "http://127.0.0.1:${TOOLPLANE_HTTP_PUBLISHED_PORT:-8080}/api/GetRequestChunks" \
  -H "Authorization: Bearer ${ADMIN_API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{"sessionId":"your-session-id","requestId":"your-request-id"}' | python3 -m json.tool
```

Use `startSeq` and `nextSeq` to decide whether the caller's last acknowledged chunk is still inside the retained window.

### List Machines For A Session

```bash
curl -sS \
  "http://127.0.0.1:${TOOLPLANE_HTTP_PUBLISHED_PORT:-8080}/api/ListMachines" \
  -H "Authorization: Bearer ${ADMIN_API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{"sessionId":"your-session-id"}' | python3 -m json.tool
```

Use this to confirm whether replacement providers are online or whether a drained machine has disappeared.

## Workflow 1: Request Is Not Starting

**Start with:**

- `toolplane_request_queue_depth`
- `toolplane_machine_active`

**Confirm with traces:**

- `request_created`
- `request_claimed`
- `machine_registered`

**Use request inspection when:**

- You need to confirm that requests are still `pending` in `ListRequests`.
- You need to confirm whether the session has any live machines in `ListMachines`.

**Decide:**

- Queue depth rising with `toolplane_machine_active == 0` usually means no provider or ownership path exists.
- `request_created` without a later `request_claimed` usually means provider availability, drain, or ownership mismatch.
- A non-zero `toolplane_machine_draining` during the same window usually points to a handoff or rollout issue rather than a caller-side problem.

**Safe action:**

- Bring a provider online before cancelling queued work.
- If the issue is a rollout, bring replacement providers online before draining the old ones.

## Workflow 2: Request Is Stuck In Flight

**Start with:**

- `toolplane_request_inflight`
- `toolplane_machine_inflight_load`

**Confirm with traces:**

- `request_execution_started`
- `request_lease_expired`
- `machine_drain_started`

**Use request inspection when:**

- You need `GetRequest` to confirm `executingMachineId`, `status`, `error`, or `resultType`.
- You need `GetRequestChunks` because the question is about retained replay, not general observability.

**Decide:**

- `request_execution_started` with no later lease-expiry or drain events usually means slow provider execution.
- Repeated `request_lease_expired` events on the same `requestId` point to provider crash or progress loss.
- `machine_drain_started` plus a still-running request points to expected drain waiting behavior until in-flight work resolves.
- If `GetRequestChunks.startSeq` has advanced past the caller's last acknowledged chunk, this is a replay-window issue covered by drill D3, not an observability gap.

**Safe action:**

- Use `CancelRequest` if one request must stop.
- Do not drain a provider just to stop one bad request.

## Workflow 3: Drain Looks Stuck

**Start with:**

- `toolplane_machine_draining`

**Confirm with traces:**

- `machine_drain_started`
- absence of `machine_drain_completed`

**Use request inspection when:**

- You need active `claimed` or `running` requests from `ListRequests`.
- You need to match those active requests to the draining `machineId` via `executingMachineId`.
- You need to confirm the machine still exists in `ListMachines`.

**Decide:**

- If active requests are still tied to the draining machine, the drain is waiting for in-flight work or lease-expiry resolution.
- If no active requests remain but `toolplane_machine_draining` is still non-zero, escalate as a runtime inconsistency rather than force-killing the provider blindly.

**Safe action:**

- Keep replacement providers online during the handoff.
- Do not terminate the old provider until both `toolplane_machine_draining == 0` and the machine disappears from `ListMachines`.

## Workflow 4: Requests Are Requeueing Or Dead-Lettering

**Start with:**

- `toolplane_request_requeues_total`
- `toolplane_request_dead_letters_total`
- `toolplane_request_dead_letter_current`

**Confirm with traces:**

- `request_lease_expired`
- `request_requeued`
- `request_dead_lettered`

**Use request inspection when:**

- You need `GetRequest` for `error`, `resultType`, and `executingMachineId`.
- You need to compare repeated failures for one request or one machine owner.

**Decide:**

- Lease-expiry followed by requeue usually points to provider failure or loss of progress.
- Requeues concentrated on one machine owner often point to provider instability rather than a global queue problem.
- Dead letters with request-specific validation or tool errors can be request-shape problems instead of provider availability failures.

**Safe action:**

- Restore provider capacity or replace the failing machine before draining healthy providers.
- Use `CancelRequest` only when you intentionally want a request to terminate instead of retrying.

## Workflow 5: Gateway Is Throttling Or Degraded

**Start with:**

- `/health` `status`
- `/health` `circuit.state`
- `/health` `rateLimitRejects`
- `/health` `throttle.total`, `throttle.apiRate`, `throttle.ipRate`, `throttle.concurrency`, `throttle.circuitOpen`, `throttle.circuitProbe`, and `throttle.unknown`

**Confirm with logs:**

- throttle `reason`
- `Retry-After`
- `client_fingerprint`
- `api_key_present`

**Decide:**

- `reason=api_rate_limit` points to API-key throttling.
- `reason=ip_rate_limit` points to per-client throttling.
- `reason=concurrency_limit` points to too many simultaneous requests at the gateway.
- `circuit.state=open` or `reason=circuit_open` means the gateway is protecting itself from backend instability.
- `reason=circuit_probe_limit` means half-open probe capacity is exhausted.

**Safe action:**

- Respect `Retry-After` in callers and operator scripts.
- Use `client_fingerprint` for repeated-client correlation; do not expect raw client IPs.
- Treat `status=degraded` as a gateway-side incident even when server traces show otherwise-normal request flow.

## Workflow 6: Policy Or Auth Denial

**Start with:**

- `auth_rejected`
- `auth_policy_denied`

**Use session inspection when:**

- You need `GetSession` to confirm the targeted session.
- You need `ListApiKeys` to confirm the session-scoped `capabilities` and `keyPreview` for the affected key.

**Decide:**

- `auth_rejected` usually means the credential is missing, malformed, revoked, or otherwise invalid.
- `auth_policy_denied` usually means the credential is valid but lacks the required capability or is being used against the wrong session boundary.
- A wrong-session issue often looks like a valid key aimed at a session it does not own.

**Safe action:**

- Reissue or rotate the right session-scoped key instead of broadly widening capabilities.
- Keep the diagnosis grounded in traces and session metadata; secrets should never enter the expected workflow.

## Safe Control Actions

### Cancel One Request

Use this when the goal is to stop one execution path without collapsing provider ownership globally.

```bash
curl -sS \
  "http://127.0.0.1:${TOOLPLANE_HTTP_PUBLISHED_PORT:-8080}/api/CancelRequest" \
  -H "Authorization: Bearer ${ADMIN_API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{"sessionId":"your-session-id","requestId":"your-request-id"}' | python3 -m json.tool
```

After cancellation, confirm the request reaches a terminal rejected or dead-lettered state through `GetRequest` and the matching `request_cancelled` trace.

### Drain One Machine

Use this when the goal is rollout, provider retirement, or safe ownership handoff.

```bash
curl -sS \
  "http://127.0.0.1:${TOOLPLANE_HTTP_PUBLISHED_PORT:-8080}/api/DrainMachine" \
  -H "Authorization: Bearer ${ADMIN_API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{"sessionId":"your-session-id","machineId":"your-machine-id"}' | python3 -m json.tool
```

Follow the maintained completion checks from `server/docs/reference-deployment.md`:

- `toolplane_machine_draining == 0`
- the machine no longer appears in `ListMachines`
- `machine_drain_completed` appears in the session trace stream

Bring replacement providers online before drain if pending requests would otherwise have no owner path.

## Validation Trail

- `server/docs/observability.md` defines the stable signals, correlation keys, and redaction rules.
- `server/docs/reference-deployment.md` defines the rollout order and drain completion checks.
- `server/docs/reliability-drills.md` defines the drill vocabulary reused here for failure diagnosis.
- `server/pkg/observability/runtime_metrics_test.go` proves the maintained metric names.
- `server/cmd/proxy/observability_test.go` proves `/health`, throttle counters, `Retry-After`, and redaction behavior.

If a workflow step cannot be traced to those maintained docs or tests, it should not be treated as part of the supported operator path.
