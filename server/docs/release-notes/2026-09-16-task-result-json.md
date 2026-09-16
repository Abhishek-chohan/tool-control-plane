# Task results as durable JSON

Date: 2026-09-16

Completed tasks now record — and serve — their result as the JSON encoding
of the tool's return value, byte-identical to what the synchronous
InvokeTool long-poll serves. Previously the task waiter rendered the
decoded result with Go's `%v`, so structured results were durably
corrupted in `tasks.result`: a tool returning `{"items": ["a", "b"],
"count": 2}` landed as `map[items:[a b] count:2]`, which no client (and no
JSON parser) can read back.

- **Structured results** round-trip: what the provider submitted as JSON is
  what `GetTask` / `tasks/get` return, and what survives a restart in the
  store. The task API and the ExecuteTool long-poll now emit identical
  bytes for the same request.
- **String results** now carry their JSON encoding (`"done"` instead of
  the bare `done`) — the same bytes the provider SDKs submit via
  `json.dumps` and the long-poll already served. Clients that parse the
  result as JSON (the documented contract for the field) are unaffected;
  clients comparing the raw bytes of a text result must now expect quotes.
- The dead self-claiming execution path (`RequestsService.ExecuteTool` /
  `ExecuteRequest`, superseded by the provider-claim model) was removed;
  it had no callers and carried the same `%v` corruption.

Notes: there is no migration — terminal task rows written by older
servers keep whatever bytes they hold. Pre-1.0, per the compatibility
policy, the field's documented contract ("JSON result as string") is the
boundary and this change makes the implementation match it.
