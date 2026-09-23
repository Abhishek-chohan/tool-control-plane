# AGENTS.md — Toolplane instructions for coding agents

Instructions for coding agents working in this repository. Contains
rules and pointers, never inventories — the maintained sources hold the
detail. The Lint workflow verifies every referenced path; a stale
reference fails CI, so structural changes must update this file in the
same PR.

## What this repository is

Toolplane is a gRPC control plane for durable remote tool execution:
requests outlive callers, providers own machines with leases and
heartbeats, stream output has bounded replay, and rollouts are
drain-safe. Quick orientation and the use-it-or-not decision rule live
in `README.md`.

## Sources of truth — read these, never restate them

| Question | Answer lives in |
| --- | --- |
| The wire contract | `server/proto/service.proto` |
| Runtime semantics (lifecycle, leases, replay, drain) | `server/DOCUMENTATION.md` |
| Which SDK exposes which RPC | `SDK_MAP.md` |
| Build, test, and verification commands | `server/Makefile` (run `make help` inside `server/`) |
| What changed recently and why | `CHANGELOG.md` and `server/docs/release-notes/` |
| Portable cross-SDK behavior | `conformance/cases/` and `conformance/schema/test_case.schema.json` |
| Operator signals and audit contract | `server/docs/observability.md` |
| Production topology and rollout order | `server/docs/reference-deployment.md` |
| Failure semantics proven by runnable drills | `server/docs/reliability-drills.md` |
| Measured performance and how to reproduce it | `server/docs/capacity.md` |
| Contribution etiquette and security policy | `CONTRIBUTING.md` and `SECURITY.md` |

## How to trace a change (order is the rule)

1. Start at `server/proto/service.proto` — it is the contract.
2. Find the transport adapter that implements the RPC in
   `server/pkg/service/server.go`.
3. Move to the owning domain file in `server/pkg/service/` (tool, session,
   machine, request, or task behavior — pick by what the RPC does, not by
   name similarity).
4. Check `SDK_MAP.md` for which SDKs expose the RPC.
5. Decide whether `conformance/cases/` needs a new or updated fixture.

## Stable engineering rules

These are doctrine; they change rarely and only deliberately:

- **Error taxonomy is the contract.** New failure conditions get a
  sentinel in `server/pkg/service/errors.go` and an arm in
  `statusFromDomainError` — never a handler-local status code call.
  Unmapped conditions fail closed to INTERNAL.
- **Both storage backends.** Storage-path changes are tested against
  memory under `-race` and against Postgres in the release gate; the
  store is authoritative over per-replica caches, and anything mirrored
  into a cache is cloned.
- **Proto changes regenerate everything.** Edit
  `server/proto/service.proto`, then run the full regeneration target
  from `server/Makefile`; drift is CI-checked.
- **Behavior changes carry release notes.** User-visible changes add a
  `CHANGELOG.md` entry and a file under `server/docs/release-notes/` in
  the same change.
- **Caller-supplied durations and message sizes are bounded**, and the
  bounds are documented on the proto fields that accept them.
- **Performance evidence is advisory.** The benchmark, load, and soak
  targets in `server/Makefile` are never CI gates; sustained runs belong
  to the nightly load workflow, never per-PR CI.
- **Releases are tag-derived.** One `v*` tag publishes every artifact
  (`.github/workflows/release.yml`); package versions come from the tag
  — never hand-set a version in a package manifest.
- **New Python test files join the unit sweep.** Add them to the
  `python-unit` target's list in `server/Makefile` — a file outside it
  never runs there.

## How to discover current structure

Do not trust memory or this file for layout — look:

- Binaries and commands: `server/cmd/`
- Server-internal packages: `server/internal/` and `server/pkg/`
- SDKs: `clients/`

## Rules for this file

- Reference paths only inside backticks, and only real repo paths — the
  Lint workflow resolves every backticked token containing a slash.
- Write RPC names as `ServiceName.MethodName` (no slashes) so they are
  never mistaken for paths.
- When a referenced path moves or dies, fix this file in the same
  change that moved it.
- Session and process state (local toolchains, CI credentials, working
  conventions per contributor) does not belong here — this file is
  repository contract only.
