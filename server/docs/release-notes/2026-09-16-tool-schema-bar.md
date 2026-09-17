# Tool schemas clear a validity bar at registration

Date: 2026-09-16

Tool schemas were opaque strings end to end — nothing validated them, so a
provider could register a "schema" that was never JSON (one repo test
itself registered `{\"type\":\"object\"}` — literal backslashes, invalid
JSON — and nothing complained). Consumers sanitized defensively
downstream, which made the garbage someone else's problem.

- **The bar**: a non-empty schema must be well-formed JSON whose root is
  an object — the minimum the MCP facade's fallback already assumes.
  Anything else is `INVALID_ARGUMENT` before any state is touched, on
  both `RegisterTool` and machine registration (a bad schema in a
  machine's tool list fails the whole registration instead of being
  silently skipped by the reconcile). Empty schemas stay allowed:
  consumers supply their own fallback.
- **Scope**: semantic JSON-Schema validation (required fields, type
  correctness) is deliberately out of scope and documented as such on
  the proto field — the schema remains advisory to consumers; this
  change guarantees only that it is parseable.
- **TypeScript provider runtime** applies the same check client-side with
  the caller's schema in the error message, instead of forwarding
  non-JSON strings to the server's rejection.
- **Dead code removed**: `GetOpenAITools` (double-encoded the schema into
  OpenAI's `parameters` as a string) and `GetToolsSummary` — zero
  callers between them.
- **Conformance**: a new `tool_schema_validation` fixture pins the
  `INVALID_ARGUMENT` rejection on every transport (the TypeScript
  conformance gateway-error mapping also learned numeric code 3).

Notes: the escaped-schema test fixture this uncovered was corrected —
the garbage registration it performed silently for its whole life is
exactly what the bar now rejects.
