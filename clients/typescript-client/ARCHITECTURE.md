# TypeScript Client Architecture

The TypeScript client is the maintained JavaScript-family projection of the Toolplane control plane: a grpc-js public surface plus an explicit provider runtime. The contract source of truth is `server/proto/service.proto`; per-RPC support labels live in `SDK_MAP.md`.

## Module Graph

```text
src/index.ts                  public exports (keep aligned with API changes)
src/core/toolplane_client.ts  main client implementation, connection, request lifecycle
src/provider_runtime.ts       explicit provider lifecycle (machine, claim, chunks, drain)
src/errors/                   typed error classes per gRPC code
src/proto/                    generated stubs (read-only)
tests/conformance/            shared-fixture runner, grpc + http transports (repo-internal)
```

## Notes For Agents

- The public surface is narrower than Python; verify missing wrappers against the proto and Python before adding.
- The HTTP adapter under `tests/conformance/` is internal harness plumbing, not a public HTTP SDK.
- `toolplane-provider` (the Python entrypoint) has no TypeScript counterpart; provider mode here is the in-process `ProviderRuntime`.
