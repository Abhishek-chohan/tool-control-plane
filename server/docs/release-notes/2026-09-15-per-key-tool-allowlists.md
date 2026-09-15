# Per-key tool allowlists

Date: 2026-09-15

`CreateApiKey` accepts an optional `allowed_tools` list. When present, the
key may invoke only those tools: `InvokeTool`, `ExecuteTool`,
`StreamExecuteTool`, and `CreateRequest` are denied with `PermissionDenied`
at the authorization interceptor for any other tool name. Keys created
without the list (all existing keys) remain unrestricted within their
session, so adoption requires no migration.

Rules:

- Entries are trimmed and de-duplicated. A list whose entries are all blank
  is rejected with `INVALID_ARGUMENT` — omit the field to say unrestricted.
- The allowlist is enforced in the authorization interceptor from the
  request payload, so it covers the gRPC, HTTP, and MCP edges alike. The
  MCP gateway's backend credential is subject to it the same way.
- The allowlist is stored per key (`api_keys.allowed_tools`, additive
  migration), returned on `ApiKey` payloads, and survives re-validation
  across replicas.

The allowlist composes with claim-time machine-tool ownership: the
allowlist decides which invocations a key may create; ownership decides
which provider machines may execute them.
