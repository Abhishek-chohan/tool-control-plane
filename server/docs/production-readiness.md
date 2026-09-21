# Production Readiness

## Purpose

One page answering the evaluator's real question — "what happens when
this breaks, and can I run it in production?" — with links into the
maintained sources rather than duplicated claims. Every section names
where the proof lives.

## Secure by default

- Production mode refuses to boot without the real thing: Postgres
  storage, Postgres-backed auth, and TLS. The core gRPC server requires
  certificate files, or an explicit upstream-terminator declaration
  (`TOOLPLANE_SERVER_TRUSTED_TRANSPORT=1`); both gateways enforce TLS
  or a declared trusted proxy. See
  [reference-deployment.md](reference-deployment.md) and the
  [trusted-transport release note](release-notes/2026-09-20-server-trusted-transport.md).
- A served-plaintext gauge (`toolplane_server_tls_enabled`) makes
  transport posture alertable instead of boot-log-only.
- Per-key tool allowlists are enforced on every invocation; capabilities
  gate every RPC. Contract: [server/DOCUMENTATION.md](../DOCUMENTATION.md).

## Failure semantics you can drill

The [reliability drill matrix](reliability-drills.md) names eight
failure scenarios — provider crash with lease expiry, caller disconnect
with bounded replay, drain with in-flight work, restart recovery,
multi-instance no-double-dispatch, lease renewal and fencing — each
with its runnable proof and its explicit limits. The load driver
re-runs the load-sensitive drills under saturation
([running drills under load](reliability-drills.md#running-drills-under-load)).

## Measured, not vibes

The [capacity envelope](capacity.md) records what the hot paths cost on
reference hardware (the two SERIALIZABLE write transactions dominate;
reads are sub-millisecond), the reproduction commands for every figure,
and what the document does not claim. Micro-benchmarks (`make bench`),
agent-shaped load (`make load`), and a soak with leak assertions
(`make soak`) are all runnable from the tree; a nightly workflow keeps
the soak evidence current.

## One contract, three transports

Transport-neutral conformance fixtures (`conformance/cases/` with the
shared schema) run against gRPC, HTTP, and MCP adapters on every
change, plus MCP-specific lifecycle coverage and a multi-instance
end-to-end case. The [release gate](release-gate.md) is the
authoritative signal; CI runs it on every change to `main`.

## Observable by contract

The [observability contract](observability.md) fixes the metric,
trace, and audit vocabulary — queue depth, in-flight, dead letters,
serialization-retry contention signals, TLS posture — with stable
identifiers and an append-only audit trail. The
[operator runbook](operator-runbook.md) maps symptoms to signals.

## Honest versioning

Pre-1.0: the [compatibility policy](compatibility-policy.md) governs
wire changes, error-taxonomy changes are centralized and release-noted,
and package versions are derived from tags — nothing claims a release
that does not exist.
