# Release Note — SDK Surface Cleanup (2026-09-11)

Branch: `chore/sdk-surface-cleanup`

## What changed

### Dead Python layers removed

- `toolplane/factories/` (client/component/strategy factories) deleted:
  imported by nothing — a parallel construction API that never became the
  supported path and drifted out of every maintained flow.
- `toolplane/session/` (config, legacy context) deleted: superseded by
  `core/session_context.py`; also imported by nothing.
- `MODULAR_ARCHITECTURE.md` deleted: described the factories/session
  architecture that never shipped as the supported surface. The live module
  graph is documented in `ARCHITECTURE.md` (updated).

### Math helpers removed

The demo `Add`/`Subtract`/`Multiply`/`Divide` methods (Go) and
`add`/`subtract`/`multiply`/`divide` (TypeScript) are gone, along with the
numeric-result parsers only they used. They hardcoded a fictional
server-side calculator into the client surface: any real deployment
registers its own tools, and the helpers encouraged shipping a client that
advertised capabilities the server does not have. Go/TypeScript tool
execution is `ExecuteTool`/`executeTool` with your own registered tools.
The Python client never exposed them.

### Real async in Python

`ainvoke` and `astream` were misnamed: `ainvoke` submitted work
synchronously and returned the request ID, and `astream` was a plain alias
for the blocking `stream`. Both are now real coroutines (`async def`) that
run the blocking transport calls in a worker thread via `asyncio.to_thread`,
so they can be awaited from a running event loop:

```python
request_id = await client.ainvoke("echo", session_id=..., text="hi")
chunks = await client.astream("echo", callback, session_id=...)
```

`ainvoke` keeps its fire-and-poll-later semantics (returns the request ID);
`astream` resolves with the collected chunks after the final marker.
Consumers that only want submission without a coroutine use the honestly
named `create_request`.

### get_request_status parameter order fixed

The Python facades took `get_request_status(request_id, session_id)` —
inverted relative to every other facade method (`session_id` first) and to
the manager methods underneath. Now `get_request_status(session_id,
request_id)`. Internal callers updated.

### example.py converter fix

The LangChain converter called `args_schema.model_json_schema()`
unconditionally — an `AttributeError` for any tool whose `args_schema` is
an already-material dict (or the `{}` default). It now handles pydantic v2
models, v1 models, plain dicts, and empty values; the unused `ToolException`
import is gone.

### Toolkit dedup

Seven standalone tools (`read_file`, `create_file`, `create_directory`,
`file_search`, `grep_search`, `semantic_search`,
`replace_string_in_file`) existed as byte-identical copies in
`toolkits/standalone_tools/` and `toolkits/swe/`. The `swe/` copies are now
thin re-export shims of the `standalone_tools/` originals — the single
maintained copy — so a fix lands once. The SWE toolkit's import surface is
unchanged.

### Quickstart demo

`make demo` (in `server/`) runs the README "First Offload Path" as one
command via `scripts/demo_quickstart.sh`: builds the server, boots it
in-memory, starts `example_client.py` (provider), waits for the session ID
it prints, runs `example_user.py` (consumer) against that session, and
tears everything down. Verified end to end. The README's async sample was
updated for the coroutine `ainvoke`.

## Compatibility (breaking changes in the Python surface)

- `ainvoke`/`astream` are now coroutines: `await` them. Callers that used
  the old synchronous submission should call `create_request` instead (same
  result: the request ID).
- `get_request_status` argument order changed on the facades.
- The Go client's `Add`/`Subtract`/`Multiply`/`Divide` and the TypeScript
  `add`/`subtract`/`multiply`/`divide` methods are removed.
- `toolplane.factories` and `toolplane.session` are removed from the wheel.
- `example.py`'s converter no longer crashes on dict-typed `args_schema`.

## Verification

- Python: black/isort/flake8 clean; 16 unit tests; conformance 38 passed
  (grpc + http, excluding the env-gated multi-instance/mcp cases).
- Go client: gofmt/vet/tests clean; golangci-lint clean.
- TypeScript: build, eslint, 37 unit tests, conformance 44/44 — including
  the chunk-window and stream cases through the restored `resumeStream`.
- `make demo` verified end to end against a live in-memory server.
