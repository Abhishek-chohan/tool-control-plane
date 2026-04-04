# Conformance Architecture

The conformance directory defines transport-neutral behavior fixtures that client runners can use to verify shared semantics across SDKs.

## Module Graph

```text
cases/*.json
  -> schema/test_case.schema.json
  -> consumed by language-specific conformance runners
```

## Structure

- `cases/`: individual test fixtures organized by feature family: session lifecycle, request lifecycle, request recovery, tool invocation, API-key lifecycle, and machine lifecycle.
- `schema/test_case.schema.json`: JSON schema used to validate fixture shape. The `feature` enum is the authoritative list of supported case families.
- `README.md`: high-level semantics for statuses, streaming assertions, recovery-window behavior, and mutation rules.

## Current Fixture Coverage

| Case | Feature family | Purpose |
| --- | --- | --- |
| `session_create.json` | `session_create` | Validate session creation behavior |
| `session_list.json` | `session_list` | Validate session listing behavior |
| `session_update.json` | `session_update` | Validate persisted session mutation behavior |
| `request_create.json` | `request_create` | Validate request creation, pending status, and list membership |
| `request_recovery_chunk_window.json` | `request_recovery` | Validate retained chunk-window baseline for stream recovery |
| `request_recovery_resume.json` | `request_recovery` | Validate ordered replay after last acknowledged sequence and terminal marker |
| `request_recovery_expired_window.json` | `request_recovery` | Validate canonical `out_of_range` failure for expired window |
| `invoke_unary.json` | `invoke_unary` | Validate unary invocation result matching |
| `invoke_stream.json` | `invoke_stream` | Validate ordered streaming chunks and terminal markers |
| `tool_discovery.json` | `tool_discovery` | Validate tool listing, lookup, and delete semantics |
| `api_key_lifecycle.json` | `api_key_lifecycle` | Validate API-key creation, list visibility, and revoke semantics |
| `machine_lifecycle.json` | `machine_lifecycle` | Validate machine registration, lookup, listing, and drain semantics |
| `machine_lifecycle_drain_under_load.json` | `machine_lifecycle` | Validate graceful drain behavior while in-flight work is active |

## Coverage Model

### Shared conformance (Python and TypeScript)

Python and TypeScript are the shared-fixture conformance SDKs. Both runners execute every case above across `grpc` and `http` transports, producing identical test matrices. Shared conformance is the primary public guarantee: if a behavior is claimed in `SDK_MAP.md` and backed by a case here, it is enforced across both SDKs on every CI run.

### Focused integration coverage (Go)

The Go client does not consume the shared JSON fixtures. Instead, `clients/go-client/client/toolplane_client_integration_test.go` provides opt-in live provider-backed gRPC coverage for unary and streaming execution. This test is focused supporting evidence, not full shared-fixture parity. Go's conformance role should not be confused with the shared-fixture contract.

## Notes For Agents

- These fixtures are the fastest way to confirm intended cross-client behavior before changing SDK code.
- Python and TypeScript both ship shared-fixture conformance runners against this directory; treat the fixture contract as cross-SDK input, not as a single-runner detail.
- If you change the shared contract or a client behavior that should be portable, check whether a new or updated conformance case is required.
- Do not promote Go's focused integration test to shared-fixture parity without adding actual case consumption.
