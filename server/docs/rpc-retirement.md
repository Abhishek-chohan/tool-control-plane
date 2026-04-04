# `/rpc` Retirement And Migration Plan

## Status

`/rpc` is a server-side compatibility endpoint on the documented removal path to `v2.0.0`. It is outside the maintained product boundary and receives no new feature work.

Maintained client code paths no longer use `/rpc`. The remaining removal work is now server-side only.

## Product Boundary

The maintained product is:

"A remote tool-execution control plane with a protobuf/gRPC contract, an HTTP gateway compatibility layer, and optional ecosystem adapters."

That means:

- The protobuf/gRPC contract in `server/proto/service.proto` is the source of truth.
- The HTTP gateway generated from protobuf annotations is the maintained non-gRPC compatibility surface.
- `/rpc` is migration-only server behavior and is not a normal choice for new integrations.

## Target Removal Release

- Target removal release: `v2.0.0`
- No new SDK examples or required workflows should depend on `/rpc` between now and removal.
- Required CI and the release gate remain focused on the maintained control-plane story and do not grow `/rpc` coverage.

## Migration Paths

### Go And TypeScript SDK Consumers

- Use the maintained gRPC client surfaces for session, machine, tool, request, and task lifecycle work.
- The maintained Go and TypeScript SDKs no longer include in-tree JSON-RPC helpers.
- Move any remaining external `/rpc` integrations to the maintained gRPC client surfaces or the HTTP gateway.

### Direct HTTP Consumers

- Migrate from JSON-RPC payloads on `/rpc` to the maintained protobuf HTTP gateway endpoints.
- Treat the HTTP gateway as a compatibility layer over the same contract, not as a separate product surface.

### JavaScript Consumers

- Use `clients/typescript-client/` for the maintained JavaScript-family path.
- There is no in-tree Node JSON-RPC client.

## Compatibility Window Rules

- `/rpc` may remain on the server only where it is still needed to unblock migration.
- No maintained SDK, example, or required workflow may depend on `/rpc`.
- Historical notes about `/rpc` should point to this document rather than recreating compatibility helpers in client folders.

## Current Repository Effects

- Root and SDK documentation now describe Go and TypeScript as gRPC-only maintained SDKs.
- The shared parity story now lives across Python, Go, and TypeScript only.
- There is no in-tree Node helper for `/rpc`.
