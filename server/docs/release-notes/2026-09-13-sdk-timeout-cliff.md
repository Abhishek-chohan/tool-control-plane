# 2026-09-13 — Python SDK timeout cliff closed

## Summary

`invoke` waited a fixed 60 seconds for completion, never sent a timeout
to the server, returned the full status envelope instead of the tool's
result (contradicting every documented example), and flattened typed
server errors into a generic `ToolplaneError`. A tool that legitimately
ran longer than a minute was abandoned by its caller while the request
kept executing and was eventually requeued — duplicate work with no
way for the caller to know.

## Changes

- `timeout_seconds` threads through `invoke`/`stream`/`ainvoke`/
  `astream` → `execute_tool`/`stream_tool` → `ExecuteToolRequest` and
  the HTTP `InvokeTool` payload. The wire field already existed; the
  server already enforced it. `0` keeps the server default.
- `invoke` unwraps and returns the tool's result value. The full status
  dict remains available through `get_request_status`.
- Typed server errors (`ToolplaneAPIError` and subclasses) propagate
  with their identity instead of being rewrapped as a generic failure.
- A lapsed local wait raises `ToolplaneTimeoutError` carrying the
  request ID and the last observed status. The default wait budget is
  60s, or `timeout_seconds + 15` when a timeout is supplied;
  `wait_timeout` overrides explicitly.
- Both facades (gRPC `Toolplane` and `ToolplaneHTTP`) and both session
  contexts share the behavior; the conformance adapters' unwrap
  workaround is removed since the SDK now does it.

## Migration

- Callers that unpacked the status envelope themselves (reading
  `result["result"]`) should read the returned value directly. Error
  handling gains precision: catch `ToolplaneNotFoundError` etc. before
  `ToolplaneError`, and `ToolplaneTimeoutError` for wait lapses.
