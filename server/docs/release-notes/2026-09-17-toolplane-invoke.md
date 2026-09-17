# toolplane invoke

Date: 2026-09-17

The consumer side of the unified command: one line from "I have a running
server and a registered tool" to a result.

```bash
toolplane invoke add --session demo --input '{"a":2,"b":3}' --wait 30s
```

- **Fire-and-forget by default** (`--wait 0`): the request ID prints with
  the follow-up polling command, and execution continues server-side.
- **`--wait 30s`** blocks server-side until the tool reaches a terminal
  state or the wait elapses (returning the in-flight status — the
  documented wait contract, bounded by the server's 1h ceiling with the
  documented `OUT_OF_RANGE` rejection for larger values).
- **`--stream`** attaches to the tool's chunk stream and prints chunks as
  they land.
- **`--idempotency-key`** dedups retries: the same key returns the
  original request instead of double-executing a side-effecting tool.
- **Exit codes carry the contract**: 0 success, 2 invalid input
  (malformed `--input`, over-ceiling waits via `OUT_OF_RANGE`), 3 unknown
  tool/session, 5 precondition failures, 7 unreachable server — with a
  teaching message pointing at `toolplane serve`.
- **`--format json`** renders `{request_id, status, result, error}` for
  scripts; table mode prints the bare result for humans.

Notes: `invoke` speaks `api.v1` directly (InvokeTool /
StreamExecuteTool). Argument validation — JSON well-formedness, session
presence — runs before any network activity with messages that name the
fix.
