# Coding Conventions

> Conventions are inferred from the maintained source, lint configs, and the
> per-area `.github/instructions/*.instructions.md` files. Generated code
> (`.pb.go`, `_pb2.py`, `_pb.js`, `gen/`) is excluded.

## Core Sections (Required)

### 1) File Naming

| Surface | Convention | Example | Evidence |
|---------|-----------|---------|----------|
| Go | `snake_case.go`, one domain per file | `requests.go`, `machine.go` | `server/pkg/service/` listing |
| Go test files | `<name>_test.go` alongside source | `requests.go` / `request_runtime_test.go` | `server/pkg/service/` |
| Python | `snake_case.py`; layered sub-packages | `toolplane_client.py`, `provider_runtime.py`, `http_core/http_request.py` | `clients/python-client/toolplane/` |
| TypeScript | `camelCase.ts` for source, `kebab-case` dirs | `toolplane_client.ts`, `provider_runtime.ts` | `clients/typescript-client/src/` |
| Conformance fixtures | `<feature>.json` (snake_case) | `request_recovery_resume.json`, `machine_lifecycle_drain_under_load.json` | `conformance/cases/` |
| Docs | `kebab-case.md` | `operator-runbook.md`, `release-gate.md` | `server/docs/` |

### 2) Function / Variable / Type Naming

- **Go**: exported types `PascalCase` (`RequestsService`, `GRPCServer`,
  `RequestStreamExpiredError`), unexported `camelCase` (`cleanupStalledRequests`,
  `ensureRequestDefaults`). Constants `PascalCase`
  (`machineHeartbeatTTL`, `requestLeaseDuration`). Interfaces suffixed with the
  method verb when narrow (`sessionScopedRequest`, `requestMetricsSource`).
- **Python**: `PascalCase` classes (`Toolplane`, `ProviderRuntime`), `snake_case`
  functions/methods, private members prefixed `_` (e.g. `_delete_session_on_server`,
  per `SDK_MAP.md:105`).
- **TypeScript**: `PascalCase` classes/exports (`ToolplaneClient`,
  `ProviderRuntime`), `camelCase` methods (`createSession`, `appendRequestChunks`).

### 3) Project / Layer Conventions (from AI-navigation instructions)

Codified in `.github/instructions/`:
- `server.go` is the **transport adapter only** — proto conversion, status-code
  mapping, validation. Business logic belongs in the owning service file
  (`server-services.instructions.md`).
- `service.proto` is the source of truth; for contract changes the order is:
  edit proto → regenerate → update `server.go` → update owning service → update
  SDK wrappers → check conformance (`proto-tracing.instructions.md`).
- Python is the **parity baseline**; Go/TS are intentionally narrower — verify
  against Python or the proto before assuming parity (`sdk-navigation.instructions.md`).
- Keep SDK-specific fixes inside the owning client folder; do not reintroduce
  `/rpc` JSON-RPC helpers (`go-client.instructions.md`).

### 4) Linting & Formatting

| Surface | Linter | Key config | Evidence |
|---------|--------|-----------|----------|
| Go (server + go-client) | golangci-lint v1.64.8 | enabled: `errcheck, gofmt, govet, ineffassign, unused`; `timeout: 5m`; `modules-download-mode: readonly`; excludes generated strictly | `.golangci.yml` |
| Python | Black (line 88), isort (black profile), Flake8 (+docstrings/builtins/annotations), mypy (strict: `disallow_untyped_defs`, `strict_optional`) | `pyproject.toml:81-116`; `.github/workflows/lint.yml:31-41` | |
| TypeScript | ESLint `--max-warnings=0` with `@typescript-eslint` | `clients/typescript-client/package.json:9`; `lint.yml:72-94` | |

Generated proto dirs are excluded from formatting: Black `extend-exclude`
includes `toolplane/proto`; isort `skip_glob = ["toolplane/proto/**"]`
(`pyproject.toml:85-107`). golangci `exclude-generated: strict`.

### 5) Error Handling

- **Go server → gRPC**: domain errors are plain `fmt.Errorf`; the transport
  adapter wraps them with `status.Errorf(codes.X, ...)` (e.g.
  `server.go:67` uses `codes.AlreadyExists`, `codes.Internal`). Typed sentinel
  errors exist in storage (`ErrConfigMissing`, `ErrExplicitInMemoryMode`,
  `ErrToolOwnershipConflict`) and requests (`RequestStreamExpiredError`) for
  `errors.Is` matching.
- **Auth**: structured gRPC codes — `Unauthenticated` (no/bad token),
  `PermissionDenied` (missing capability or session/user binding mismatch)
  (`authorizer.go:160-217`). Tokens are **redacted** in trace events
  (`redactToken`, `redactTokenFromContext`).
- **Python**: exception hierarchy under `toolplane/core/errors.py`
  (`ToolplaneError`, `ConnectionError`, `RequestError`, `SessionError`,
  `ToolError`, `MachineError`, `TaskError`), re-exported via `__init__.py`.
- **TypeScript**: dedicated error classes under `src/errors/`
  (`ToolplaneError`, `ConnectionError`, `TimeoutError`, `ProtocolError`,
  `ValidationError`); instructions forbid throwing plain strings
  (`typescript-client.instructions.md`).

### 6) Logging / Tracing

- Go server uses stdlib `log` (`log.Printf`, `log.Fatalf`). No third-party
  structured logger in the server runtime.
- A `trace.SessionTracer` (`pkg/trace/tracer.go`) records structured
  `SessionEvent`s (request created/requeued/dead-lettered, task retry/dead-letter,
  auth validated/rejected/denied). The metrics collector implements `Record()`
  to increment counters (`runtime_metrics.go:61-75`); a logging tracer can be
  layered on with `--trace-sessions` (`main.go:51-55`).
- The proxy logs component-specific errors (e.g. `"ERROR: Flush not supported"`).

### 7) Import Organization

- **Go**: stdlib first, then external (`google.golang.org/grpc`), then internal
  (`toolplane/...`) — grouped with blank lines (see `main.go:3-26`).
- **Python**: Black + isort (`profile="black"`, `multi_line_output=3`); layered
  package imports (`from .core import ...`).
- **TypeScript**: barrel exports from `src/index.ts` (`export * from './interfaces'`);
  examples import from the public surface.

### 8) Concurrency

- Go: `sync.RWMutex` around in-memory state maps
  (`RequestsService.requestsMutex`, `requests.go:19`). `sync.Map` for per-request
  signals (`signals`). `sync/atomic` counters in the metrics collector.
- Storage writes use serializable transactions (`withSerializableTx`,
  `store.go:92-112`) for claim/capacity safety.

### 9) Evidence

- `.golangci.yml`, `clients/python-client/pyproject.toml:81-116`
- `.github/instructions/{server-services,proto-tracing,sdk-navigation,python-client,go-client,typescript-client,conformance}.instructions.md`
- `server/pkg/service/server.go` (status codes), `server/cmd/server/auth/authorizer.go` (codes + redaction)
- `clients/python-client/toolplane/core/errors.py`, `clients/typescript-client/src/errors/`
- `server/pkg/service/requests.go:16-31` (mutex/concurrency), `server/pkg/storage/store.go:92-112` (serializable tx)
