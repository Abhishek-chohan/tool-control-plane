# Go Client Architecture

The Go client is a narrower maintained projection of the Toolplane control plane: a gRPC-only library plus demo entries, with no provider-runtime harness. The contract source of truth is `server/proto/service.proto`; per-RPC support labels live in `SDK_MAP.md` — do not infer parity from this folder's shape.

## Module Graph

```text
client/toolplane_client.go    library: connection, typed errors, request-lifecycle helpers
client.go                     runnable demo entry
examples/basic, examples/advanced   usage references
proto/                        generated stubs (read-only; regenerated from server/proto)
```

## Notes For Agents

- Keep library behavior in `client/toolplane_client.go`; `client.go` and `examples/` demonstrate, they don't own logic.
- Errors are typed (`client.Error` with code and retryable; `errors.Is` against `client.Err…` sentinels) — never string-match.
- If a feature is missing here but exists in Python, decide explicitly whether Go gains support or stays intentionally narrower; record it in `SDK_MAP.md`.
