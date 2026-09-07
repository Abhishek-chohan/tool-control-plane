# Release Note — Capability Split and Per-Machine Identity (2026-09-06)

Branch: `feat/capability-split`

## What changed

### Capabilities: read / invoke / provide / admin

The `execute` capability previously meant both "invoke tools" and "be a
provider," so any session key could register a machine, claim queued work,
and submit results — the structural hole behind the review's result-forgery
finding (fencing narrowed it; this closes it).

- New capabilities: `invoke` (consumer operations: create requests, execute
  tools, cancel work) and `provide` (provider operations: register
  machines/tools, claim, submit results and chunks, renew leases, drain,
  unregister). `read` and `admin` are unchanged.
- **`execute` is now a legacy alias**: it is still accepted anywhere
  capabilities are parsed and expands to `invoke+provide`. Existing keys and
  database rows keep working unchanged. Do not mint new keys with it.
- The authz policy table is repartitioned accordingly: 12 provide-scoped
  RPCs, 7 invoke-scoped, plus the existing read/admin entries.

### Per-machine identity

- `RegisterMachine` now **mints a per-machine credential** (`toolplane_key`
  style UUID secret), stores only its SHA-256, and returns the plaintext
  exactly once in the new `Machine.machine_token` response field.
- **Provide-scoped RPCs must present the credential** via the
  `x-toolplane-machine-token` metadata / `X-Toolplane-Machine-Token` header
  when the server runs session-key auth. The proxy and MCP gateway forward
  the header.
- **Machine-ID takeover is closed**: re-registering an existing machine ID
  requires its credential. Previously any execute-capable key silently
  re-pointed an existing machine's SDK version, IP, and tools to itself.
- Machines registered before this change have no stored hash; the first
  credential presented for them binds (one-time migration).
- Fixed mode (single dev key) and auth-disabled mode are exempt from the
  machine-token gate — production refuses both, so the gate covers exactly
  the multi-credential deployments where hijacking matters.

### SDKs

- Python (gRPC and HTTP) and TypeScript capture `machine_token` from the
  registration response automatically and present it on every call; nothing
  changes for provider authors. The Go client has no provider surface.

## Migration

- **Keys minted with `execute`**: keep working (expand to invoke+provide).
  Rotate to explicit `invoke`/`provide` lists at your leisure.
- **Custom provider clients** (not using the maintained SDK runtimes) must
  capture `machine_token` from `RegisterMachine` and send it on provide-scoped
  RPCs, or those calls fail with `PERMISSION_DENIED` in session-key mode.
- **Operators**: no schema action needed; the `machines.token_hash` column
  migrates idempotently at startup. Existing machine rows bind a credential
  on the first provide call that presents one.

## Known limits

- Machine credentials are bearer secrets held in provider process memory;
  there is no per-machine expiry/rotation RPC yet (a machine recovers by
  re-registering with a new ID, which mints a fresh credential).
- Tool-name ownership remains single-owner per session; multi-worker capacity
  is unchanged (addressed separately).
