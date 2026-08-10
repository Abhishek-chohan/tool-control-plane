# Technology Stack

> Toolplane is a remote tool-execution control plane. It owns a protobuf/gRPC
> contract (Go server), an HTTP gateway compatibility layer, and multi-language
> client SDKs. This file documents only what is verifiable in the manifests and
> config files.

## Core Sections (Required)

### 1) Runtime Summary

Toolplane is a **multi-language polyrepo-style monorepo**: one Go server (the
source of truth) plus three maintained client SDKs (Python, Go, TypeScript) and
one optional adapter (TypeScript MCP). Each language surface has its own
manifest and runtime version.

| Area | Value | Evidence |
|------|-------|----------|
| Primary language (server) | Go 1.24.0 (toolchain go1.24.10) | `server/go.mod:3-5` |
| Primary language (richest client) | Python ≥ 3.8 | `clients/python-client/pyproject.toml:20-24,35` |
| Secondary client language | TypeScript 5.x on Node 20 | `clients/typescript-client/package.json:40`; `.github/workflows/conformance-python.yml:130` |
| Secondary client language | Go (separate module `toolplane-go-client`) | `clients/go-client/go.mod` |
| Go module path (server) | `toolplane` | `server/go.mod:1` |
| Go module path (client) | `toolplane-go-client` | `clients/go-client/go.mod` |
| Package managers | Go modules (server + go-client); `pip`/setuptools (python); npm (ts clients) | `server/go.mod`, `clients/python-client/pyproject.toml:1-4`, `clients/typescript-client/package.json` |
| Module / build system | `go build` (server); setuptools (`pyproject.toml`); `tsc` + proto codegen (ts) | `server/Makefile:48-49`, `clients/typescript-client/package.json:8` |

### 2) Production Frameworks and Dependencies

**Go server** (`server/go.mod` — these run in production):

| Dependency | Version | Role in system | Evidence |
|------------|---------|----------------|----------|
| `google.golang.org/grpc` | v1.70.0 | gRPC server + client transport (canonical API) | `server/go.mod:12` |
| `google.golang.org/protobuf` | v1.36.5 | Protobuf runtime | `server/go.mod:13` |
| `grpc-ecosystem/grpc-gateway/v2` | v2.26.3 | HTTP/JSON gateway (grpc-gateway) compiled from `.proto` annotations | `server/go.mod:9`; `server/proto/service.proto` `google.api.http` options |
| `github.com/jackc/pgx/v5` | v5.6.0 | Postgres driver (stdlib adapter) for durable storage | `server/go.mod:10`; `server/pkg/storage/store.go:13` |
| `github.com/google/uuid` | v1.6.0 | ID generation for sessions/requests/tasks/tools | `server/go.mod:8` |
| `github.com/sony/gobreaker` | v1.0.0 | Circuit breaker in the proxy/gateway | `server/go.mod:20`; `server/cmd/proxy/main.go:13` |
| `golang.org/x/time/rate` | v0.14.0 | Token-bucket rate limiting in the proxy | `server/go.mod:27`; `server/cmd/proxy/main.go:14` |
| `google.golang.org/genproto/googleapis/api` | (indirect, pinned) | `google.api.http` annotations for grpc-gateway | `server/go.mod:11` |

Note: `stretchr/testify` (v1.8.4) and `puddle/v2` are listed `// indirect` in
`server/go.mod`; testify is used by tests only.

**Python client** (`clients/python-client/pyproject.toml` — runtime deps):

| Dependency | Version | Role | Evidence |
|------------|---------|------|----------|
| `grpcio` | ≥1.80.0 | gRPC transport | `pyproject.toml:27` |
| `grpcio-tools` | ≥1.80.0 | Proto code generation | `pyproject.toml:28` |
| `protobuf` | ≥6.31.1,<7.0.0 | Protobuf runtime | `pyproject.toml:29` |
| `googleapis-common-protos` | ≥1.63.0 | google.api annotations | `pyproject.toml:30` |
| `requests` | ≥2.25.0 | HTTP gateway transport | `pyproject.toml:31` |
| `pydantic` | ≥2.0.0 | Config/schema validation | `pyproject.toml:33` |

`clients/python-client/requirements.txt` is a **pinned superset** that also
includes `langchain==0.3.19`, `langchain-core==0.3.63`, `google-api-python-client`,
and `chardet`. Verified: `langchain` is **absent from `pyproject.toml`'s
`[project.dependencies]`**, and the installed SDK surface
(`toolplane/__init__.py`, `toolplane/provider_runtime.py`) imports **neither**
`toolkits/` **nor** `langchain`. LangChain is used only by the bundled example
tools under `toolplane/toolkits/{standalone_tools,swe}/`. So the **core SDK is
langchain-free**; the extras leak in only via `pip install -r requirements.txt`
(used by example/docs workflows, not the published package). Consider moving
them to an `[project.optional-dependencies]` extra.

**TypeScript client** (`clients/typescript-client/package.json` — `dependencies`):

| Dependency | Version | Role | Evidence |
|------------|---------|------|----------|
| `@grpc/grpc-js` | ^1.14.1 | gRPC transport | `package.json:44` |
| `google-protobuf` | ^3.21.4 | Protobuf runtime | `package.json:46` |
| `axios` | ^1.6.0 | HTTP gateway transport (conformance/internal only per `SDK_MAP.md`) | `package.json:45` |
| `uuid` / `@types/uuid` | ^9.0.0 | ID generation | `package.json:47-48` |

### 3) Development Toolchain

| Tool | Purpose | Evidence |
|------|---------|----------|
| `golangci-lint` v1.64.8 | Go lint (server + go-client) | `.golangci.yml`; `.github/workflows/lint.yml:66-68` |
| `protoc` + plugins (`protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-grpc-gateway`) | Generate Go/proto stubs | `server/Makefile:69-75`; `server/Dockerfile:9-11` |
| `grpc_tools.protoc` (Python) | Generate Python proto stubs | `server/Makefile:78-85` |
| `grpc_tools_node_protoc` | Generate TS proto stubs | `clients/typescript-client/package.json:15` |
| Black / isort / Flake8 / mypy | Python lint+format | `pyproject.toml:81-115`; `.github/workflows/lint.yml:31-41` |
| ESLint + `@typescript-eslint` | TS lint (both ts packages) | `package.json:9`; `.github/workflows/lint.yml:72-94` |
| `tsc`, `ts-node`, `tsx` | TS compile + run | `clients/typescript-client/package.json:39-41` |
| pytest (≥7.0), pytest-asyncio | Python test runner | `pyproject.toml:42,48` |
| Go `testing` | Go test runner | `server/pkg/service/*_test.go` |
| Docker (multi-stage) | Container builds | `server/Dockerfile`; `clients/python-client/Dockerfile` |
| Docker Compose | Reference deployment topology | `server/deploy/reference/compose.yaml` |

### 4) Key Commands

```bash
# Go server (from server/)
make build            # builds bin/toolplane-server
make run              # runs the built server
make gen-proto-all    # regenerate all tracked proto outputs from service.proto
make check-proto-drift # fail if generated stubs diverge from source proto
make release-gate     # authoritative end-to-end gate (conformance + observability + runtime)

# Go client (from clients/go-client/)
go test ./...

# Python client (from clients/python-client/)
python -m pytest tests/conformance/test_conformance_runner.py -m conformance -q
python -m black --check toolplane && python -m isort --check-only toolplane && python -m flake8 toolplane

# TypeScript client (from clients/typescript-client/)
npm run build
npm run test:unit          # unit tests
npm run test:conformance   # shared-fixture conformance (needs live server)
npm run lint
```

Evidence: `server/Makefile`, `clients/typescript-client/package.json:7-23`.

### 5) Environment and Config

- Config sources: `server/.env.example` (server + proxy), Go flags in `server/cmd/server/main.go:29-34` and `server/cmd/proxy/main.go:262-273`, and env vars read in `server/cmd/server/config.go`.
- Server-mode enforcement: `TOOLPLANE_ENV_MODE` ∈ {development, test, production}. Production **forbids** `auth=disabled`, `auth=fixed`, and in-memory storage (`server/cmd/server/config.go:42-72`).
- Required env vars (production): `TOOLPLANE_DATABASE_URL` (Postgres DSN), `TOOLPLANE_AUTH_MODE=postgres`, `TOOLPLANE_PROXY_ALLOWED_ORIGINS`, plus gRPC TLS cert/key files (`TOOLPLANE_SERVER_TLS_CERT_FILE` / `..._KEY_FILE`). Proxy requires `TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND=0` in production.
- Development defaults (`server/.env.example`): `ENV_MODE=development`, `AUTH_MODE=fixed`, `AUTH_FIXED_API_KEY=toolplane-conformance-fixture-key`, `STORAGE_MODE=memory`. These are intentionally non-secret and exist for local work and CI fixtures only.
- Deployment/runtime constraints: gRPC on `:9001`, HTTP gateway on `:8080`, Prometheus metrics on `:9102` (see `server/Makefile:66`, `server/Dockerfile`, `server/deploy/reference/compose.yaml`).

### 6) Evidence

- `server/go.mod`, `clients/go-client/go.mod`
- `clients/python-client/pyproject.toml`, `clients/python-client/requirements.txt`
- `clients/typescript-client/package.json`, `clients/typescript-mcp-adapter/package.json`
- `server/Makefile`, `server/Dockerfile`, `server/deploy/reference/compose.yaml`
- `server/.env.example`, `server/cmd/server/config.go`, `server/cmd/proxy/config.go`
- `.github/workflows/lint.yml`, `.github/workflows/conformance-python.yml`, `.github/workflows/release-gate.yml`
- `.golangci.yml`
