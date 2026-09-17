# HealthCheck reports the real build version

Date: 2026-09-17

`HealthCheck`'s `version` field — already part of `api.v1` — used to
return a hardcoded `"1.0.0"` placeholder that was wrong for every build
except the fictional one. It now carries the binaries' build identity:

- `make build` injects it via `-ldflags` from
  `git describe --tags --always --dirty` (overridable with `VERSION=...`
  or CI tag metadata), so release artifacts report their tag and source
  builds honestly report `dev`.
- `toolplane --version` prints the same identity the server's
  `HealthCheck` reports, making client/server skew visible with two
  commands and no parsing: compare `toolplane --version` against the
  `version` in the health response.

Notes: no wire change — the field already existed on
`HealthCheckResponse`; only its value became meaningful. The standard
`grpc.health.v1` service (used by orchestrator probes) is unchanged and
carries no version, per its spec.
