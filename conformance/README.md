# Conformance Fixtures

This directory contains language-agnostic behavior fixtures used by SDK conformance runners.

## Fixture contract

- Each case is a JSON document validated against `schema/test_case.schema.json`.
- Each case is transport-neutral and must run over both `grpc` and `http` adapters.
- Case IDs are globally unique and stable.

## Coverage model

- **Shared conformance**: Python and TypeScript execute every case across both transports. These runners are the primary public guarantee for the maintained SDK surface.
- **Focused integration**: Go provides opt-in live provider-backed gRPC coverage via `clients/go-client/client/toolplane_client_integration_test.go`. This is supporting evidence, not full shared-fixture parity.
- Tool-discovery RPCs (`ListTools`, `GetToolById`, `GetToolByName`, `DeleteTool`) now have shared fixture coverage through `tool_discovery.json`, so maintained support labels can match the public wrappers.

## Why Shared Fixtures Matter

Shared fixtures are part of the repo's simplification case, not just its test layout.

- Python and TypeScript shared fixtures keep portable request, replay, drain, and provider-runtime semantics from drifting silently by SDK.
- Focused Go integration coverage plus explicit support labels keep narrower surfaces visible instead of implying full parity.
- Together with `SDK_MAP.md` and `server/docs/compatibility-policy.md`, this reduces the amount of semantic guesswork adopters would otherwise do across maintained client projections.

See `../server/docs/economic-case.md` for how this conformance model fits into the broader operational argument.

## Features in this milestone

- `session_create`
- `session_list`
- `invoke_unary`
- `invoke_stream`
- `tool_discovery`
- `session_update`
- `request_create`
- `request_recovery`
- `api_key_lifecycle`
- `machine_lifecycle`
- `provider_runtime`

## Status semantics

- Expected terminal statuses: `done` or `failure`.
- Runners may validate richer intermediate states when available.
- Request creation fixtures may assert the initial `pending` state before any machine consumes the request.

## Streaming semantics

- Stream cases define expected ordered chunks.
- Runners assert:
  - chunk order,
  - chunk count,
  - terminal callback marker (`is_final=True`) was observed.

## Request recovery semantics

- Recovery cases assert the retained chunk window returned by `GetRequestChunks()`, including ordered chunks plus the `start_seq` and `next_seq` metadata that define the current replay window.
- Resume cases assert ordered replay after the last acknowledged sequence number and the terminal callback marker for the resumed stream, including replay from inside a trimmed retained window after long-running streams overflow the cap.
- Expired-window cases assert that resuming before the retained window fails with the canonical `out_of_range` error instead of silently skipping lost chunks.

## Mutation semantics

- Session update cases assert persisted mutable fields on the updated session payload.
- API key lifecycle cases assert non-empty identifiers and key material on create, list visibility while active, and absence from active listings after revoke.
- Machine lifecycle cases assert machine registration, list visibility, direct lookup by ID, and disappearance from machine listings after drain.
- Provider runtime cases assert that the explicit Python `ProviderRuntime` surface claims work, submits final results, appends streaming chunks, and survives drain-under-load through the shared transport adapters.

## Drain-under-load semantics

- The `machine_lifecycle_drain_under_load` case asserts that `DrainMachine` stops new work routing immediately while still allowing in-flight requests to complete before unregistering the machine.
- This is closer to the repo's control-plane differentiator than basic CRUD-style machine fixtures and should remain in the maintained release story.

## Provider Runtime Semantics

- The `provider_runtime` cases exercise the explicit Python provider runtime surface rather than ad hoc polling threads hidden in the conformance adapters.
- Unary provider-runtime cases cover request claim and final result submission through the maintained request lifecycle.
- Streaming provider-runtime cases cover chunk append behavior plus terminal result submission.
- Drain-under-load provider-runtime cases assert that the explicit provider runtime keeps in-flight work alive while the server drains and unregisters the machine.

## Reliability Drill Coverage

The named drill matrix lives in `server/docs/reliability-drills.md`. Shared conformance covers the transport-neutral slice of that matrix.

- D2. Caller disconnect during streaming: `request_recovery_chunk_window.json` and `request_recovery_resume.json` prove retained-window inspection and replay from the last acknowledged sequence.
- D3. Replay behind the retained window: `request_recovery_expired_window.json` proves the explicit expired-window failure path, and `request_recovery_resume_trimmed_window.json` proves replay from inside a trimmed retained window.
- D4. Deploy drain with in-flight work: `provider_runtime_drain_under_load.json` and `machine_lifecycle_drain_under_load.json` prove drain stopping new routing while in-flight work still resolves.
- Provider-runtime unary and streaming fixtures establish the maintained claim, chunk-append, and submit lifecycle that the failure drills build on, but provider crash or lease expiry and durable restart recovery remain primarily focused runtime proof today.
- D6. Claim-state and capacity safety is only partially represented in shared conformance today; drain-blocked routing is exercised indirectly through the drain-under-load cases, while standalone machine-capacity rejection is not yet a named shared fixture.
