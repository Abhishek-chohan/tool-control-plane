# Release Note — Observability and Lifecycle (2026-09-11)

Branch: `chore/observability-lifecycle`

## What changed

### Prometheus-native metrics (new dependency)

The hand-rolled text renderer is replaced by
`github.com/prometheus/client_golang` (v1.22.0) on a private registry,
served by `promhttp`. Existing metric names are unchanged; what is new:

- `toolplane_grpc_requests_total{method,code}` and
  `toolplane_grpc_request_duration_seconds{method}` (histogram) recorded by
  gRPC interceptors wrapped **outermost** — auth rejections and transport
  failures count like any other request. Labels are bounded by the proto
  method surface and the gRPC code set.
- Go runtime and process collectors (`go_goroutines`,
  `process_resident_memory_bytes`, …) come with the registry.
- The release-gate observability script asserts the new series and drives
  one real RPC before scraping (labeled series only render once observed).

The interceptor composition fix this forced is the important part: the
server previously passed a singular `grpc.UnaryInterceptor` (auth); adding
metrics as `ChainUnaryInterceptor` alongside it silently kept only the
singular option. The server now builds **one explicit chain** — metrics
outermost, auth (or the anonymous dev-mode stand-in) inside — so both run.

### Structured logging (slog)

`TOOLPLANE_LOG_FORMAT=json` switches the server to `slog` JSON records;
the default stays text. A std-log bridge routes the services' existing
`log.Printf` output through the same handler, so every line shares one
format without a mechanical rewrite of every call site. The proxy and
gateway keep std log for now (they are single-purpose processes whose
output is small).

### Durable audit trail

A filtered set of lifecycle events now persists to a new `audit_events`
table (Postgres migration; capped 1024-entry ring in the in-memory store):
session create/delete, API-key create/revoke, machine register/unregister,
and request/task dead-letters. The recorder is a tracer in the existing
chain: persistence failures are logged with full event identity and
dropped — the audit trail never blocks or fails the operation that
produced the event.

### grpc.health.v1

The gRPC server registers the standard health service and reports
`SERVING` once listening; shutdown reports `NOT_SERVING` before draining.
Load balancers and orchestrators probe this instead of inferring
liveness from connection acceptance.

### Lifecycle hardening

- **Bounded GracefulStop**: a 20-second deadline backs the drain; on expiry
  the server force-stops instead of hanging forever on a stuck stream.
- **Single exit path, no `log.Fatalf` in goroutines**: the metrics server
  reports its outcome on a channel instead of `log.Fatalf` (which skips
  deferred store close), and the main loop selects over the gRPC serve
  result, the metrics result, and the shutdown completion — every exit
  runs the deferred storage close.
- **Proxy and MCP gateway**: signal handling with a bounded 10-second
  `http.Server.Shutdown` (previously `log.Fatal(ListenAndServe())` — a
  SIGTERM killed them mid-stream with no drain), plus `IdleTimeout` for
  parked keep-alive connections.

### traceparent

Decision: **consume for correlation, not propagation**. The MCP gateway
already validates and forwards W3C `traceparent` to the backend; the JSON
proxy now forwards it too (allowlisted header). The server surfaces it on
RPC log lines (debug level; failures at info) — no span is created, and no
tracing backend is assumed. Operators running OpenTelemetry collectors can
join request logs onto traces via the trace id.

## Compatibility

- Metric names are unchanged; counters now carry their canonical `_total`
  suffix in exposition (they already did in assertions). Scrapers relying
  on the previous output see the same series plus the new ones.
- The `audit_events` migration runs on startup. `Storer` gained
  `RecordAuditEvent`.
- `go.mod` gains `prometheus/client_golang` v1.22.0 with the Go 1.24
  toolchain pin intact (v1.24.1 of the library requires Go 1.25 — rejected
  deliberately).

## Known limits

- The audit trail is append-only; no query API or retention pruning yet
  (`CleanupOldTasks` is the template for an audit pruner).
- The proxy and gateway still log unstructured; they are next in line if
  their output grows.
- traceparent appears on logs but not on metrics labels — label
  cardinality would be unbounded.

## Verification

- New `pkg/observability` suite: interceptor series (method/code/duration,
  plus traceparent on the request log), audit filtering (high-frequency
  execution events excluded), JSON logger output, and the std-log bridge.
- `release_gate_observability.sh` passes end to end against a live server
  and proxy, now asserting the gRPC series and Go runtime collectors.
- Full server suite under `-race`, memory and Postgres modes: zero data
  races; golangci-lint, gofmt, vet clean; full Python and TypeScript
  conformance unchanged.
