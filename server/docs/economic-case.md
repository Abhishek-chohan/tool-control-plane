# Economic Case

## Purpose

This document makes Toolplane's economic case explicit without turning it into a speculative ROI or price-savings story.

The maintained claim is narrower: for the remote-tool workloads that fit the wedge, Toolplane reduces repeated engineering work and lowers incident risk by centralizing request lifecycle, replay, drain, observability, rollout procedure, and compatibility expectations in one maintained control plane.

## Narrow Claim

Toolplane adds a control-plane layer. That layer is only justified when the work already forces product teams to rebuild operational behavior around remote execution.

Use this case only when at least one of the following is already true:

- the tool can outlive the original caller
- streaming output needs retained replay or later inspection
- provider ownership, heartbeats, or drain matter operationally
- deploys or restarts would otherwise risk dropping in-flight work
- teams are already rebuilding retries, cancellation, request inspection, or rollout logic around one remote tool family

Direct tool calling remains simpler for short-lived same-lifecycle work.

## Simplification Thesis

For the workloads above, Toolplane centralizes recurring engineering work in one maintained place instead of letting every team reimplement it around each remote tool family.

| Concern | Maintained artifact | Simplification claim | Current limit |
| --- | --- | --- | --- |
| Request lifecycle, retries, cancellation, and terminal-state inspection | `server/DOCUMENTATION.md`, `server/docs/release-gate.md` | Teams do not each need to invent remote request IDs, claim state, retry bookkeeping, or inspection semantics for the same class of tool | The value is for remote tools; local same-process calls are still simpler without the control plane |
| Streaming recovery and replay-window behavior | `server/DOCUMENTATION.md`, `server/docs/reliability-drills.md`, `conformance/README.md` | Bounded replay behavior and its failure modes stay one maintained contract instead of per-tool glue | Replay is limited to the retained 100-chunk window; there is no durable full-history replay guarantee |
| Provider ownership and drain | `server/docs/agent-runtime-integration-seam.md`, `server/DOCUMENTATION.md`, `server/docs/operator-runbook.md`, `server/docs/reference-deployment.md` | Provider lifecycle, heartbeats, and drain-safe rollout become platform behavior rather than ad hoc worker conventions | Pending requests are not migrated automatically during drain |
| Operator diagnosis | `server/docs/observability.md`, `server/docs/operator-runbook.md`, `server/docs/release-gate.md` | Operators can start from one stable vocabulary for queue depth, inflight work, dead letters, drain, and gateway throttling | The signal set is intentionally small; it does not replace workload-specific business metrics |
| Cross-SDK behavior and compatibility expectations | `conformance/README.md`, `server/docs/compatibility-policy.md`, `SDK_MAP.md` | Shared fixtures, focused live coverage, explicit support labels, and a written compatibility policy reduce semantic drift and hidden wrapper differences | This is not a claim of full parity beyond what `SDK_MAP.md` and shared conformance prove |
| Rollout and recovery procedure | `server/docs/reference-deployment.md`, `server/docs/release-gate.md`, `server/docs/reliability-drills.md` | Upgrade order, drain handoff, rollback checks, and recovery expectations are maintained once instead of rediscovered by each team | The reference path does not prove every topology or live Postgres-backed auth configuration |

## Why The Seam Matters

The economic case is not tied to one vendor runtime or one edge adapter.

- `server/docs/agent-runtime-integration-seam.md` keeps lifecycle semantics in the control plane and treats adapters as thin, replaceable translation layers.
- `server/docs/incremental-adoption.md` keeps the first migration narrow: move one painful remote tool first and leave the rest of the caller stack in place.

That matters economically because the durable value sits below replaceable adapters. Teams should not have to restart their runtime design every time an upstream model or agent API changes.

## Before And After: One First Offloaded Tool

Use one sandboxed code-execution worker or one environment-bound browser worker as the reference scenario.

### Before Toolplane

- the caller or workflow runtime owns remote request IDs, polling, retries, cancellation, and terminal-state inspection itself
- stream disconnect behavior is custom to that tool or SDK integration
- worker ownership and drain depend on local shutdown hooks or best-effort queue handling
- rollout reasoning starts from per-service conventions instead of one maintained handoff contract
- on-call diagnosis starts from scattered logs, custom identifiers, and tool-specific debugging steps

### After Toolplane For The Same Tool

- the surrounding orchestrator can remain the caller of record while Toolplane owns request lifecycle for that one session-scoped tool family
- retained replay, inspection, lease expiry, and drain semantics come from one maintained runtime contract
- operators start from maintained request, task, and machine identifiers plus a small published metric and health vocabulary
- deployment and rollback reasoning reuse the written reference path and release gate instead of rediscovering rollout order and drain checks
- SDK support labels and shared fixtures keep the maintained behavior more explicit across projections

The point is not that complexity disappears. The point is that recurring complexity moves into one maintained platform boundary.

## Credible Claims

The current repo can credibly claim the following for wedge workloads only:

- fewer dropped long-running jobs on the maintained durable path because disconnect and restart do not automatically destroy request state
- safer rollout reasoning because drain order, handoff checks, and rollback checks are explicit and documented
- clearer debugging because request, task, and machine identifiers plus runtime metrics and `/health` are maintained operator surfaces
- less repeated retry, replay, and drain logic in product teams because those semantics are already modeled centrally
- lower cross-SDK ambiguity because shared fixtures, focused live coverage, explicit support labels, and compatibility rules make maintained behavior easier to audit

## Claims The Repo Should Not Make

- guaranteed dollar savings or universal ROI
- universal benefit for short-lived same-process tool calls
- full elimination of queueing, runtime, or operational complexity
- parity beyond what `SDK_MAP.md`, shared conformance, and focused integration coverage actually prove

## Decision Rubric

Use the smallest thing that fits.

| Option | Prefer it when | Move past it when |
| --- | --- | --- |
| Direct tool calling | The work is quick, local, and shares the caller lifecycle | The caller would need to add persistence, replay, provider ownership, or drain behavior around the tool |
| Basic queue or job system | Deferral is enough and operators do not need request inspection, retained replay, or explicit provider ownership | The workload also needs bounded replay, safe drain, or machine-scoped ownership |
| Toolplane | One remote tool already needs request lifecycle control, retained replay, provider ownership, deploy-safe drain, and operator-visible inspection | The work is still simple enough that the control-plane layer would be pure overhead |

Three practical questions should decide the first evaluation:

1. Is one remote tool already forcing the team to build bespoke retry, recovery, cancellation, or drain logic?
2. Would a disconnect, provider restart, or deploy currently create dropped work or opaque debugging?
3. Can the team keep the rest of its orchestration stack unchanged and move only that one hard tool first?

If the answer to those questions is mostly no, direct tool calling or a simpler queue remains the better choice.

## Validation Trail

- `README.md` defines the durable-remote-tool wedge and the first offload boundary.
- `server/docs/agent-runtime-integration-seam.md` defines why the durable value sits below replaceable adapters.
- `server/docs/reliability-drills.md` and `server/docs/release-gate.md` define the failure proof and validation path behind lower-incident-risk claims.
- `server/docs/operator-runbook.md` and `server/docs/observability.md` define the published operator vocabulary.
- `server/docs/incremental-adoption.md` defines the one-tool migration path that makes simplification observable in practice.
- `server/docs/reference-deployment.md` defines the maintained rollout, drain, and rollback procedure.
- `conformance/README.md`, `SDK_MAP.md`, and `server/docs/compatibility-policy.md` define how maintained SDK behavior stays explicit and auditable.

If a future economic claim cannot be traced to those maintained artifacts, it should not be treated as part of the supported Toolplane story.
