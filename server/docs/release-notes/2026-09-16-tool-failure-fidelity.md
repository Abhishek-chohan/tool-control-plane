# Tool failure fidelity end to end

Date: 2026-09-16

SWE toolkit wrappers now raise on failure instead of returning error text.
Every failure — including a nonzero command exit from `execute_bash` —
propagates as an exception, the provider submits a rejection, the durable
request records FAILED, and MCP clients see `isError: true` with the
failure text in the content.

Behavior change for models and callers: a failing tool call no longer looks
like a successful DONE result whose text happens to contain an error.
Status-level consumers (retry heuristics, eval scoring, task orchestration)
read the failure from the status; the failure detail still travels in the
message.

Covered by:

- `tests/test_swe_failure_fidelity.py` (unit: raise semantics per failure
  mode, success paths unchanged)
- the conformance failure-fidelity case (gRPC + HTTP): a raising provider
  tool records FAILED and the error text surfaces
- `TestSyncCallFailureCarriesIsError` (MCP facade): rejection →
  `isError: true` with the failure message in the content
