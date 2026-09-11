# Release Note — MCP Initialize Compatibility (2026-09-11)

Branch: `feat/mcp-interop`

## What changed

The MCP gateway previously spoke only the 2026-07-28 stateless revision:
every request had to carry `_meta` with a protocol version and client
capabilities, and there was no `initialize` handshake. That is correct for
the newest revision, but it locked out the installed base of MCP clients —
the initialize-based 2025-03-26 / 2025-06-18 / 2025-11-25 revisions — which
could not connect at all.

### Initialize compatibility path

The gateway now answers the standard `initialize` handshake on the same
`POST /mcp` endpoint:

- The client's requested revision is echoed when supported
  (`2025-03-26`, `2025-06-18`, `2025-11-25`); an unsupported revision gets
  `2025-06-18` back (the legacy spec's server-side behavior — the client
  decides whether to continue).
- The response carries the standard InitializeResult shape: negotiated
  `protocolVersion`, `capabilities` (tools), `serverInfo`
  (`toolplane-mcp-gateway`), and `instructions`.
- After the handshake, initialize-based clients send plain `tools/list` /
  `tools/call` without `_meta`. The gateway serves them with legacy shapes:
  `tools/list` returns `{tools}` without the 2026 `resultType` wrapper, and
  `tools/call` returns a plain CallToolResult without the 2026 `_meta`
  chunk decoration. Execution is identical — durable request, provider
  claim, synchronous aggregation.
- `ping` is served for both generations. `notifications/initialized` and
  other notifications are accepted with 202 as before.
- 2026-only methods (`server/discover`, `tasks/*`) requested without
  `_meta` get a JSON-RPC method-not-found whose message names the missing
  `_meta` keys, so a 2026 client that simply forgot the envelope can
  self-correct. The 2026 surface itself is unchanged, and the stateless
  per-request validation (missing `clientCapabilities` with a declared
  version, bad cursors, unsupported 2026 versions) still applies.

A legacy client has no way to name a Toolplane session, so the gateway
provisions one session per API key under its own `mcp-gateway` user, exactly
as the 2026 path does when `dev.toolplane/session_id` is absent. Providers
register into that session to serve a legacy client.

### Conformance via the official MCP SDK

A new TypeScript adapter test connects the official
`@modelcontextprotocol/sdk` Client over StreamableHTTP to a live gateway:
initialize handshake, `tools/list` surfacing a freshly registered tool, and
`tools/call` completed end to end by a provider. This exercises the exact
client stack the installed base ships, not a hand-rolled HTTP caller.

### README claim correction

"Any MCP client" was an overclaim: initialize-based clients could not
connect. The README now states the two supported client generations and how
each connects.

## Compatibility

- The 2026-07-28 surface is byte-for-byte unchanged: same methods, same
  `_meta` requirements, same result shapes, same errors.
- The `initialize` and `ping` methods are additive. A 2026 client that
  somehow sent `initialize` with a 2026 `_meta` gets the same legacy-style
  answer; no 2026 client does this.
- `tools/list`/`tools/call` without `_meta` now succeed for
  initialize-based clients. Deployments relying on rejecting them (none
  known) would need an edge-level rule.

## Known limits

- The compatibility path serves `tools` only: `resources/*`, `prompts/*`,
  and sampling are not translated. The Tasks extension remains 2026-only.
- One session per API key on the legacy path: a legacy client cannot target
  or switch sessions (it has no `_meta` channel). Multi-session legacy use
  goes through the TypeScript adapter, which owns the session mapping.
- Legacy `tools/call` results omit the 2026 `_meta` chunk cursor; streaming
  chunk content still appears as leading text blocks.

## Verification

- Go: `pkg/mcp/initialize_test.go` — version negotiation for each supported
  revision, unsupported-version defaulting, meta-less `ping`, legacy
  `tools/list` shape (no `resultType`), and legacy `tools/call` executed
  end to end by a real provider with the 2026 `_meta` stripped. Full suite
  under `-race` clean; lint clean.
- TypeScript: the new official-SDK test passes alongside the full adapter
  suite (13/13); gateway build and eslint clean.
