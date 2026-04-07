# Observability Contract

This document is the maintained operator-facing observability contract for Toolplane.

It answers three questions directly:

- which signals are supported today
- which identifiers and redaction rules are stable enough to depend on
- where operators should look first for the most common runtime failures

## Supported Signals

### Server Lifecycle Trace Events

Source of truth:

- `server/pkg/trace/tracer.go`
- emitters under `server/pkg/service/`
- auth emitters under `server/cmd/server/auth/authorizer.go`

When `toolplane-server --trace-sessions` is enabled, the server emits structured JSON events prefixed with `session_trace`.

The stable top-level identifiers are:

- `sessionId`
- `machineId`
- `toolId`
- `requestId`
- `taskId`
- `event`
- `timestamp`

Request and task events still carry richer status or retry details in `metadata`, but correlation should start from the top-level identifiers above rather than from ad hoc metadata parsing.

The supported event families are:

- auth: `auth_validated`, `auth_rejected`, `auth_policy_denied`
- session and api key: `session_created`, `session_updated`, `session_deleted`, `api_key_created`, `api_key_revoked`
- tool lifecycle: `tool_registered`, `tool_refreshed`, `tool_deleted`, `tool_registration_rejected`
- machine lifecycle: `machine_registered`, `machine_ping_update`, `machine_drain_started`, `machine_drain_completed`, `machine_unregistered`, `machine_inactive_pruned`
- request lifecycle: `request_created`, `request_claimed`, `request_execution_started`, `request_execution_completed`, `request_execution_failed`, `request_cancelled`, `request_chunks_appended`, `request_lease_expired`, `request_requeued`, `request_dead_lettered`
- task lifecycle: `task_created`, `task_execution_started`, `task_retry_scheduled`, `task_execution_completed`, `task_execution_failed`, `task_cancelled`, `task_dead_lettered`

### Server Runtime Metrics

Source of truth:

- `server/pkg/observability/runtime_metrics.go`
- runtime state readers under `server/pkg/service/`

The server now exposes a Prometheus-style text endpoint at `/metrics` on the listener configured by `toolplane-server --metrics-listen`.

The default listener is `127.0.0.1:0`, which binds a free local port and logs the chosen address. Operators should set an explicit address such as `:9102` in environments where the endpoint will be scraped.

The supported metrics are:

- `toolplane_request_queue_depth`
- `toolplane_request_inflight`
- `toolplane_request_dead_letter_current`
- `toolplane_request_requeues_total`
- `toolplane_request_dead_letters_total`
- `toolplane_machine_active`
- `toolplane_machine_draining`
- `toolplane_machine_inflight_load`
- `toolplane_task_pending`
- `toolplane_task_running`
- `toolplane_task_dead_letter_current`
- `toolplane_task_retries_total`
- `toolplane_task_dead_letters_total`

These metrics are intentionally small and map directly to the first operator questions this repo already supports: how much work is queued, whether work is stuck in flight, whether retries or dead letters are rising, and whether machine drain is making progress.

### Proxy Health Payload

Source of truth:

- `server/cmd/proxy/main.go`

`toolplane-gateway` exposes `/health` with a maintained JSON payload containing:

- overall `status`
- circuit-breaker `state`, `inflight`, `accepted`, `rejected`, and breaker counts
- aggregate `rateLimitRejects`
- per-reason throttle counters in `throttle`
- a UTC `timestamp`

The supported throttle reasons are:

- `api_rate_limit`
- `ip_rate_limit`
- `concurrency_limit`
- `circuit_open`
- `circuit_probe_limit`
- `throttle_unknown`

### Proxy Throttle Logs

Source of truth:

- `server/cmd/proxy/throttle_metrics.go`

Each throttled request emits a structured log line with:

- `reason`
- `retry_after`
- `api_key_present`
- `client_fingerprint`
- `detail`

The fingerprint is a short SHA-256-derived identifier of the client IP. It is intended for correlation of repeated throttling from the same client without logging the raw address.

## Correlation And Redaction Rules

These are part of the maintained contract, not incidental implementation detail.

### Stable Correlation Keys

- server traces use top-level `sessionId`, `machineId`, `toolId`, `requestId`, and `taskId`
- request and task retries or dead letters should be correlated through `requestId` and `taskId` first, then through `sessionId`
- machine drain and request expiry should be correlated through `machineId` plus `requestId`
- proxy throttling should be correlated through the stable throttle `reason`, `Retry-After`, and `client_fingerprint`

### Redaction Rules

- raw API-key secrets are never emitted in the supported observability surface
- auth traces may include `tokenPreview` or `keyId`, but not the secret token itself
- proxy throttle logs expose only `api_key_present`, not the API key value
- proxy throttle logs do not emit raw client IPs; they emit `client_fingerprint` instead
- metrics and `/health` expose aggregate counts only and do not include request payloads, user identifiers, or client addresses

## Operator Playbook

| Symptom | First check | Likely owner | Next confirming check |
| --- | --- | --- | --- |
| Startup rejected in production mode | process startup error plus env validation output | server bootstrap | confirm `TOOLPLANE_ENV_MODE`, auth mode, storage mode, and proxy origin or backend settings |
| Request queue growing or requests not starting | `toolplane_request_queue_depth` and `toolplane_machine_active` | request runtime or provider availability | inspect `request_created`, `request_claimed`, and `machine_registered` traces |
| Request stuck in flight | `toolplane_request_inflight` and `toolplane_machine_inflight_load` | provider runtime or drain | inspect `request_execution_started`, `request_lease_expired`, and `machine_drain_started` traces |
| Request repeatedly retries | `toolplane_request_requeues_total` | request runtime | inspect `request_lease_expired` and `request_requeued` events for the same `requestId` |
| Requests are dead-lettering | `toolplane_request_dead_letters_total` and `toolplane_request_dead_letter_current` | request runtime or provider failure | inspect `request_dead_lettered` traces and the preceding retry history |
| Drain appears stuck | `toolplane_machine_draining` | machine runtime | inspect `machine_drain_started` without a matching `machine_drain_completed`, then inspect active `requestId` values for that machine |
| Task retries keep increasing | `toolplane_task_retries_total` | task runtime | inspect `task_retry_scheduled` and the underlying request failure for the linked `requestId` |
| Tasks are dead-lettering | `toolplane_task_dead_letters_total` and `toolplane_task_dead_letter_current` | task runtime | inspect `task_dead_lettered` plus the terminal request error |
| Proxy is throttling clients | `/health` throttle counters and proxy throttle logs | gateway | inspect `reason`, `Retry-After`, and `client_fingerprint`; confirm configured API-key or IP rate limits and breaker state |
| Circuit breaker is open | `/health` `circuit.state` and HTTP 503 from `/health` | gateway | inspect breaker `Counts` and recent `concurrency_limit` or `circuit_open` throttle logs |

## Validation Coverage

The maintained validation coverage for this contract is:

- request and task trace identifiers: focused assertions in `server/pkg/service/request_runtime_test.go` and `server/pkg/service/task_test.go`
- runtime metrics rendering: `server/pkg/observability/runtime_metrics_test.go`
- proxy health payload, throttle accounting, and `Retry-After`: `server/cmd/proxy/observability_test.go`
- broader request, drain, retry, and dead-letter lifecycle behavior: the focused service tests referenced throughout `server/DOCUMENTATION.md`

The release gate documentation in `server/docs/release-gate.md` calls out which observability claims are backed by direct tests versus documented operator guidance.
