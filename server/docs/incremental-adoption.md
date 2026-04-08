# Incremental Adoption

## Purpose

This document packages the maintained first-tool migration path for Toolplane.

The adoption claim is intentionally narrow: a platform team should be able to offload one painful remote tool without replacing the rest of its agent runtime or orchestration stack.

## Core Claim

Toolplane is meant to own the remote execution slice first:

- request creation and request inspection
- provider ownership and machine heartbeats
- retained replay for streaming work
- cancellation, retry visibility, and safe drain

The surrounding caller can keep its own tool-selection logic, prompt flow, and direct local tools.

## What Should Move First

Choose one remote tool that already has operational pain.

- sandboxed code execution that may outlive the original model turn
- environment-bound browser or agent work running outside the caller process
- ingestion or indexing work that keeps running after the original caller disconnects
- approval-gated or multi-step work where request inspection, cancellation, or drain matter immediately

Do not start with the easiest local tool. Start with the tool that already makes the caller own too much retry, recovery, rollout, or inspection logic itself.

## Coexistence Model

The maintained coexistence story is explicit.

- The existing model runtime, workflow engine, or orchestration layer can remain the caller of record.
- Toolplane becomes the backend for selected remote tools only.
- Direct local tools can remain outside Toolplane entirely.
- Optional edge adapters can present one Toolplane session through another protocol, but request lifecycle and drain still stay native to the Toolplane control plane.

This is coexistence first, broader migration later only if the first tool proves its value.

## Stepwise Path

### 1. Pick One Painful Tool

Select one tool family where at least one of the maintained wedge conditions is already true: long-running execution, provider restarts, streaming recovery, queueing, or deploy-safe drain.

### 2. Bind It To One Explicit Session

Run the provider runtime against one dedicated session for that tool family or operator boundary. Keep the scope narrow so request inspection, cancellation, and drain stay easy to reason about.

### 3. Keep Existing Orchestration Around It

The caller should continue deciding when to use the tool. Toolplane should only own the backend execution path for that tool: request creation, claiming, result submission, inspection, and machine drain.

### 4. Validate The Happy Path

Run one provider and one consumer against the same session, confirm that tool discovery works, and complete both a synchronous and an asynchronous request.

### 5. Exercise Inspection And Cancellation Early

Do not stop at a successful call.

- Inspect the asynchronous request through the control plane with `get_request_status()` or `GetRequest`.
- If replay matters for the candidate workload, validate `GetRequestChunks()` or `ResumeStream()` while the retained window still exists.
- Once the first real long-running tool is wired in, immediately rehearse `cancel_request()` or `CancelRequest` so cancellation behavior is visible during evaluation rather than deferred until production.

### 6. Exercise Drain Before Broadening Scope

Bring a replacement provider online for the same session and tool, list the active machines, and drain the retiring machine. Treat the rollout-safe handoff as part of the first evaluation, not a later production-only concern.

### 7. Expand Only After The First Tool Proves Out

Move a second tool only after the first one makes the control-plane value visible in practice: fewer dropped runs, easier diagnosis, safer rollouts, or less bespoke retry and recovery logic in the caller.

## Maintained Reference Flow

The repository already contains the maintained reference flow for this first migration.

- `clients/python-client/example_client.py` is the provider side: connect, create a session, register machine-backed tools, and keep the explicit provider loop running.
- `clients/python-client/example_user.py` is the consumer side: connect to the same session, list tools, invoke provider-backed work, and poll request state.
- `clients/python-client/README_EXAMPLES.md` is the quickest walkthrough for the provider-consumer pair.
- `clients/typescript-mcp-adapter/README.md` is proof that an external runtime can bind one Toolplane session outward without redefining lifecycle semantics.

The Python pair is the recommended starting point because Python is the richest maintained SDK surface in the repository today.

## Reference Workload Shape

The first candidate should look like a real remote-execution problem, not a trivial function call.

Two reference shapes fit the current repo well:

- one sandboxed code-execution worker that may run longer than the original caller interaction and benefits from request inspection, cancellation, and drain-safe rollout
- one environment-bound browser or agent worker where provider restart, machine ownership, or deployment handoff are part of the real workload shape

The bundled sample tools remain intentionally simple so the lifecycle is easy to trace. They are scaffolding for the first offload, not the claim that simple local helpers need a control plane.

## Validation Path

### Baseline Flow

1. Start the server using the supported local-development or reference-deployment path.
2. Run `clients/python-client/example_client.py` and capture the printed `TOOLPLANE_SESSION_ID`.
3. Run `clients/python-client/example_user.py` against that same session.
4. Confirm that the provider and consumer remain separate processes and that the rest of the caller stack is still outside Toolplane.

### Control-Plane Value Checks

1. Happy path: confirm synchronous and asynchronous request execution on the shared session.
2. Inspection: confirm the async request state can be inspected through `get_request_status()` or `GetRequest`.
3. Cancellation: once the sample tool is replaced with the first real long-running candidate, create one async request and rehearse `cancel_request()` or `CancelRequest` on that same session.
4. Drain: bring a replacement provider online, list machines for the session, drain the retiring machine, and confirm completion using the checks in `server/docs/operator-runbook.md` and `server/docs/reference-deployment.md`.

If the first evaluation does not include inspection, cancellation, or drain, the control-plane value will remain hidden behind a happy-path demo.

## Adapter Role

The TypeScript MCP adapter is supporting evidence for coexistence, not the center of the adoption story.

- It proves that one Toolplane-backed session can be surfaced through another protocol.
- It does not replace the native session, request, replay, or drain contract.
- It should be introduced only as proof that the surrounding runtime can stay in place while selected remote tools move behind Toolplane.

## Related Docs

- `README.md` for the wedge, comparison boundary, and first-offload overview
- `server/DOCUMENTATION.md` for request lifecycle, retained replay, and provider drain semantics
- `server/docs/operator-runbook.md` for symptom-first inspection, cancellation, and drain workflows
- `server/docs/reference-deployment.md` for the blessed topology and rollout-safe handoff order
