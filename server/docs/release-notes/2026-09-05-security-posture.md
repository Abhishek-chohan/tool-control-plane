# Release Note — Security Posture Hardening (2026-09-05)

Branch: `fix/security-posture`

## What changed

### Trust model

- **`CreateApiKey` requires explicit capabilities.** Omitting capabilities no
  longer mints a full read+execute+admin key; an empty (or all-blank) list is
  rejected with `INVALID_ARGUMENT`. All SDK `create_api_key`/`createApiKey`/
  `CreateAPIKey` wrappers take a required, non-empty capability list and fail
  fast client-side.
- **`InvalidateSession` is now a real session-wide kill switch**: it revokes
  every live API key of the session (persisted, propagated to the auth index)
  and returns the count in `InvalidateSessionResponse.revoked_api_keys`.
- **`RefreshSessionToken` is removed from the contract** (RPC, messages,
  handler, authz entry, SDK wrappers, docs). It fabricated a token that
  authenticated nothing over a no-op; see `server/docs/rpc-retirement.md`.

### Credential hygiene

- New API keys use the format `toolplane_key_<uuid>` — the session ID is no
  longer embedded in the secret. Existing keys keep working (authentication
  is hash-based, format-agnostic).
- API-key authentication is an O(1) SHA-256 hash-index lookup instead of a
  scan over every stored key; revoked and deleted keys are removed from the
  index immediately.
- Key previews show only a masked tail (`***<last8>`); the old first-8-chars
  preview was nearly constant given the known key prefix.

### Authorization

- `RequireSessionCapability`/`RequireUserCapability` now fail closed when no
  authenticated principal is present (previously they returned nil). The
  auth-disabled dev mode installs anonymous fixed-mode interceptors so local
  development behaves identically; production still refuses to boot disabled.
- `ResumeStream` no longer leaks cross-session request existence: a
  session-bound caller asking for another session's request gets the same
  `NOT_FOUND` as a missing request.

### HTTP edges

- The JSON proxy no longer accepts `?api_key=` query-string credentials
  (headers only) — URLs land in access logs, history, and Referer headers.
- `X-Forwarded-For` is trusted only behind an explicit
  `TOOLPLANE_TRUSTED_PROXY=1` declaration; otherwise client identity comes
  from `RemoteAddr`.
- Rate-limiter buckets and throttle observability key on a SHA-256 hash of
  the credential, never the raw secret.
- Both HTTP edges (proxy, MCP gateway) gained `--tls-cert-file` /
  `--tls-key-file`, and production refuses to boot without TLS or an explicit
  `TOOLPLANE_TRUSTED_PROXY=1`.
- The MCP gateway gained rate limiting (`--api-rate/--api-burst/--ip-rate/
  --ip-burst`, mirroring the proxy), a hashed, TTL-bounded, size-capped
  session cache (raw secrets no longer live in gateway memory), and generic
  client-facing error messages with correlation IDs — backend detail is
  logged server-side only.

### Toolkit quarantine

The SWE/standalone toolkits are explicitly quarantined rather than
half-guarded:

- The cosmetic first-token bash blocklist is removed (trivially bypassed;
  provided a false sense of guardrail) and replaced with a hard command
  timeout (`TOOLPLANE_BASH_TIMEOUT_SECONDS`, default 120) and an optional
  workspace root (`TOOLPLANE_WORKSPACE_ROOT`: bash cwd + editor path jail,
  soft boundary only).
- Tool descriptions loudly state the tools are UNSANDBOXED and must only run
  on providers isolated in a container/VM.
- The editor state file (full pre-edit copies of every edited file) moved
  from the shared world-writable temp dir to a per-user cache path
  (`~/.cache/toolplane/editor_state.json`, overridable via
  `TOOLPLANE_EDITOR_STATE_FILE` / legacy `EDITOR_STATE_FILE`).
- `grep_search` no longer mutates the process working directory (race under
  provider concurrency).
- `example.py` registers a safe echo demo by default; the SWE toolkit is
  opt-in via `--toolkit swe` / `TOOLPLANE_EXAMPLE_TOOLKIT=swe` with a printed
  warning, and langchain is no longer imported unless the toolkit is used.

## Migration

- **SDK users minting keys**: pass an explicit capability list — the server
  rejects empty lists. `create_api_key()`/`createApiKey()` signatures now
  require the argument.
- **Users of `refresh_session_token()`**: the method is gone; nothing needed
  it (nothing about sessions expired). Rotate keys with create+revoke; kill a
  session with `invalidate_session()`.
- **Operators relying on `?api_key=` or XFF trust**: move credentials to
  headers and set `TOOLPLANE_TRUSTED_PROXY=1` only when a TLS terminator is
  actually in front.
- Existing database rows are untouched: legacy keys authenticate as before,
  and revoked keys stop working on every replica path that consults the
  shared auth index.

## Known limits

- Cross-replica revocation propagation for `InvalidateSession` follows the
  same single-instance cache story as today (addressed with the store-backed
  read model work).
- `TOOLPLANE_WORKSPACE_ROOT` is a soft jail (cwd + path prefix checks), not
  a sandbox; real isolation must come from the provider's container/VM.
