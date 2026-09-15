# Contributing to Toolplane

Toolplane is a gRPC control plane that gives agents structured access to tools — sessions, machines, requests, and keys — with gRPC, HTTP, and MCP gateways and Python, TypeScript, and Go SDKs. Bug reports, questions, and design discussion (as issues) and focused pull requests are all welcome.

## Ground rules

- Participation follows the [Code of Conduct](CODE_OF_CONDUCT.md).
- Security issues never go through public issues or pull requests. Report them privately per [SECURITY.md](SECURITY.md).

## Getting oriented

- [README.md](README.md) — what Toolplane is, quickstart, architecture, and repository layout.
- [server/docs/local-development.md](server/docs/local-development.md) — the supported local bootstrap: server and gateways with development defaults (memory storage, fixed dev key). Postgres belongs to the [reference deployment](server/docs/reference-deployment.md), not the local path.
- [server/DOCUMENTATION.md](server/DOCUMENTATION.md) — runtime semantics and the request lifecycle.
- [server/docs/compatibility-policy.md](server/docs/compatibility-policy.md) — protobuf, gateway, and SDK compatibility rules; `api.v1` is the version boundary.

## Checks

Run the checks that touch your change.

Makefile targets run from `server/`:

| Command | What it does |
| --- | --- |
| `make test-race` | Full Go server suite under `-race` |
| `make python-unit` | Python SDK unit tests |
| `make conformance-python` | Shared-fixture conformance over gRPC + HTTP (auto-boots a server) |
| `make conformance-python-mcp` | Adds the MCP transport (boots the MCP gateway) |
| `make release-gate` | Conformance, observability, and runtime slices; storage legs run on memory unless `TOOLPLANE_DATABASE_URL` is set |
| `make check-proto-drift` | Regenerated stubs must match what's committed |

The minimum for any pull request is `make test-race`, `make conformance-python`, and `make check-proto-drift`. CI's Release Gate runs the same target against a Postgres service and adds separate multi-instance and MCP steps, so it — not a local memory-backed run — is the authoritative result.

Other checks run from their own working directories:

- **Python lint** (`clients/python-client`): install with `pip install -e ".[dev]"`, then `black --check toolplane`, `isort --check-only toolplane`, and `flake8 toolplane` (configured in `clients/python-client/.flake8`). mypy runs advisories-only in CI.
- **Go lint** (`server`, `clients/go-client`): golangci-lint.
- **TypeScript lint** (`clients/typescript-client`, `clients/typescript-mcp-adapter`): `npm ci && npm run lint`.
- **SDK README drift** (repository root): `python tools/gen_sdk_readmes.py --check` (run it without `--check` to regenerate).

## Making changes

### Proto changes

`server/proto/service.proto` is the wire contract. On the current `api.v1` line, changes stay additive; wire-breaking changes require an explicit `v2` boundary rather than an in-place edit. The full rules are in [server/docs/compatibility-policy.md](server/docs/compatibility-policy.md). Regenerate stubs following [server/docs/proto_regeneration.md](server/docs/proto_regeneration.md) (it pins the tool versions) and keep `make check-proto-drift` green. Generated client code ships in-repo, so it is part of the same PR.

### SDK surface changes

If an SDK's public surface changes, update [SDK_MAP.md](SDK_MAP.md) and `CHANGELOG.md` in the same PR and regenerate the README API sections with `python tools/gen_sdk_readmes.py` — CI enforces this with the drift check.

### Conformance fixtures

Shared, transport-neutral behavior fixtures live in `conformance/cases/` with a JSON schema. Python and TypeScript run every fixture over gRPC and HTTP, so a fixture added there covers both SDKs at once. Prefer a fixture over SDK-specific tests for wire-level behavior.

### Server changes

Keep memory and Postgres storage parity; backend differences get caught by Release Gate's Postgres-backed scenario and multi-instance slices. When behavior that operators rely on changes, update the operational docs under `server/docs/` (runbook, reference deployment, reliability drills) in the same PR.

### Changelog

`CHANGELOG.md` is Keep-a-Changelog; detailed per-change notes go in a dated file under `server/docs/release-notes/`.

## Pull request expectations

- Use the pull request template; its sections mirror how changes in this repo are described. Write "None" where a section doesn't apply.
- One logical change per PR; split unrelated fixes.
- Commit messages follow Conventional Commits with a scope, e.g. `fix(server): ...`, `feat(python-client): ...`, `docs(conformance): ...`.
- All workflows must be green before merge: Lint, SDK Conformance & Verification, and Release Gate (Proto Drift runs when proto inputs change).

## Filing issues

- **Bug reports**: include what you ran (commands or requests), expected vs actual, the component and version or commit, and the deployment mode (`TOOLPLANE_STORAGE_MODE` / `TOOLPLANE_AUTH_MODE`). The bug template collects this.
- **Questions and design discussion**: state the goal before the mechanism, and link related issues and docs. Blank issues stay enabled for exactly this — no form needed.
- Search existing issues first.
