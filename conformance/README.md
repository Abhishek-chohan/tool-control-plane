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
- Resume cases assert ordered replay after the last acknowledged sequence number and the terminal callback marker for the resumed stream.
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
