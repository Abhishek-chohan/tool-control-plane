# Changelog

Notable changes to this repository. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); detailed
release notes live in `server/docs/release-notes/`.

## [Unreleased]

### Security

- `CreateApiKey` requires an explicit capability list (`INVALID_ARGUMENT` on
  empty); the implicit read+execute+admin default is gone, and all SDK
  wrappers take a required non-empty list.
- `InvalidateSession` is a real session-wide kill switch: revokes every live
  API key of the session and reports the count
  (`InvalidateSessionResponse.revoked_api_keys`).
- API-key auth is an O(1) SHA-256 hash-index lookup (no all-keys scan); new
  key format `toolplane_key_<uuid>` stops embedding the session ID; previews
  show a masked tail only.
- Authorizer helpers fail closed without a principal (auth-disabled dev mode
  gets anonymous interceptors; production still refuses disabled auth).
- `ResumeStream` returns identical `NOT_FOUND` for other sessions' requests
  (cross-session existence oracle closed).
- JSON proxy: no `?api_key=` URL credentials; XFF trusted only behind
  `TOOLPLANE_TRUSTED_PROXY=1`; rate-limiter/throttle keys hashed. Both HTTP
  edges gained TLS flags and refuse production plaintext without a declared
  trusted proxy.
- MCP gateway: rate limiting, hashed TTL-bounded capped session cache, and
  generic client errors with correlation IDs (backend detail logged
  server-side only).
- Toolkits quarantined: bash timeout + optional workspace root instead of the
  cosmetic blocklist, per-user editor state file, no `os.chdir`, loud
  UNSANDBOXED warnings, `example.py` defaults to a safe echo demo (SWE behind
  `--toolkit swe`).

### Removed

- `SessionsService.RefreshSessionToken` RPC (fabricated a token that
  authenticated nothing) and its Python wrappers; rotation is create+revoke.
  See `server/docs/rpc-retirement.md`.

### Added

- Conformance integrity guarantees: bootstrap failures are hard failures
  (never skips) in auto-boot mode, connectivity-skip conversion is opt-in via
  `TOOLPLANE_CONFORMANCE_ALLOW_SKIP`, and a session guard fails runs where an
  expected transport executed tests but passed nothing.
- CI: Python SDK unit suites (`python-unit` job / `make python-unit`) and the
  Go race detector (`make test-race`; CI runs every package except
  `pkg/service` under `-race` pending the clone-at-the-boundary fix).
- `gosec` and `staticcheck` enabled in `.golangci.yml`; advisory (non-blocking)
  mypy step in the Python lint job.
- HTTP `ReadHeaderTimeout`/`ReadTimeout` on the JSON proxy, MCP gateway, and
  metrics server (Slowloris hardening).

- `RequestsService.RenewRequestLease` RPC: providers renew the execution
  lease of claimed/running requests so long-running tools are not reclaimed
  mid-flight. Renewal never extends a lease past the request's absolute
  timeout. (`lease_epoch`, `leased_by`, `lease_expires_at`, `timeout_seconds`
  are now exposed on the `Request` message.)
- Per-request `timeout_seconds` override on `CreateRequest` and `ExecuteTool`
  (server default 45s, maximum 3600s).
- Lease fencing on provider writes: `UpdateRequest`, `SubmitRequestResult`,
  and `AppendRequestChunks` require the current lease grant
  (`machine_id` + `lease_epoch` from the claim response) and reject stale or
  forged writers with `FAILED_PRECONDITION`.
- Python and TypeScript provider runtimes renew in-flight leases
  automatically and present the lease grant on every fenced write.
- Conformance case `provider_runtime_fenced_submission` (grpc + http, Python
  and TypeScript runners); Go lease/fencing suites wired into
  `release-gate-runtime`.

### Changed

- Reference deployment compose refuses to start without an explicit
  `POSTGRES_PASSWORD`, `TOOLPLANE_DATABASE_URL`, and (for the bootstrap
  profile) `TOOLPLANE_BOOTSTRAP_FIXED_API_KEY`; Postgres and metrics ports are
  bound to 127.0.0.1 by default. `.env.example` is now fill-in-the-blank with
  no working default credentials.
- Python client `requirements.txt` trimmed to actual runtime + dev
  dependencies; toolkit-only deps (langchain, chardet, google-api-python-client,
  ipython) moved out — install `toolplane/toolkits/swe/requirements.txt` when
  using the toolkits. Stub `uv.lock` removed; README install paths and
  pyproject URLs fixed.
- The lease reaper reclaims on the unrenewed lease deadline (30s TTL) or the
  absolute per-attempt timeout, and scans by `visible_at` instead of
  `leased_at`.
- Line endings pinned to LF via `.gitattributes` on all platforms.

### Deprecated

- Calling the fenced provider RPCs without `machine_id`/`lease_epoch` is
  rejected; custom clients must present the claim's lease grant.

See `server/docs/release-notes/2026-09-04-lease-renewal-and-fencing.md` for
migration details.
