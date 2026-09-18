# TypeScript MCP Adapter Architecture

An optional edge adapter that binds one Toolplane session behind an MCP stdio server. It is translation only — request lifecycle, replay, and drain stay native to the control plane (`server/proto/service.proto`).

## Module Graph

```text
src/cli.ts        stdio entry point
src/server.ts     MCP request dispatcher, adapter lifecycle
src/bridge.ts     Toolplane client orchestration and semantic translation
src/protocol.ts   MCP 2026-07-28 stateless wire contract (+ legacy initialize path)
src/resources.ts  static concept-map and resource URIs
tests/            live integration tests against a booted server + provider
```

## Translation Model

- One adapter process = one bound Toolplane session; `tools/list` maps to session-scoped discovery, `tools/call` to request-backed execution (durable task handles for Tasks-capable 2026 clients, synchronous service otherwise).
- Task ids are Toolplane request ids; `tasks/get` reads the same records (chunk window honors the `dev.toolplane/last_seq` cursor), `tasks/cancel` maps to native cancellation.
- Read-only resources expose session context and recent request records — no parallel state model.

## Notes For Agents

- Keep lifecycle semantics in the control plane; only translation belongs here.
- Validate changes with the live integration tests, not README examples alone.
