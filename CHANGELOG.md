# Changelog

Notable changes to this repository. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); detailed
release notes live in `server/docs/release-notes/`.

## [Unreleased]

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
