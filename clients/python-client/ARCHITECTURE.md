# Python Client Architecture

The Python client is the most complete SDK in the repository. It exposes both gRPC and HTTP facades, supports session management, remote invocation, streaming, and local tool-provider registration.

## Module Graph

```text
toolplane/__init__.py
  -> toolplane/toolplane_client.py
  -> toolplane/provider_runtime.py
    -> toolplane/core/*
    -> toolplane/common/*
    -> toolplane/factories/*
    -> toolplane/session/*
    -> toolplane/interfaces/*
  -> toolplane/toolplane_http_client.py
    -> toolplane/http_core/*
    -> toolplane/common/*
    -> toolplane/session/*
```

## Public Entry Points

- `toolplane/__init__.py`: public export surface for both transport facades and shared types.
- `toolplane/toolplane_client.py`: gRPC-first `Toolplane` facade.
- `toolplane/toolplane_http_client.py`: HTTP-first `ToolplaneHTTP` facade.
- `toolplane/provider_runtime.py`: explicit provider lifecycle surface shared by both transports.

## Top-Level API Surface

The two facades intentionally mirror each other closely.

### `Toolplane`

- Connection lifecycle: `connect()`, `disconnect()`, context-manager helpers.
- Session lifecycle (consumer scope — portable across maintained SDKs): `create_session()`, `get_session()`, `list_sessions()`.
- Session admin (admin scope — Python-only): `list_user_sessions()`, `bulk_delete_sessions()`, `get_session_stats()`, `refresh_session_token()`, `invalidate_session()`.
- Tool invocation (consumer scope): `invoke()`, `ainvoke()`, `stream()`, `astream()`, `get_available_tools()`, `get_request_status()`.
- Provider runtime access: `provider_runtime()`.
- Backward-compatible provider aliases: `tool()`, `start()`, `stop()`.

### `ToolplaneHTTP`

- Mirrors most of the gRPC facade and adds explicit `health()` for HTTP health probing.

### `ProviderRuntime`

- Provider lifecycle: `create_session()`, `attach_session()`, `managed_session_ids()`.
- Tool registration: `register_tool()`, `tool()`.
- Runtime control: `poll_once()`, `start_in_background()`, `run_forever()`, `stop()`.

## Data Flow

### Remote invocation

1. User code creates an `Toolplane` or `ToolplaneHTTP` instance.
2. The client establishes a connection and initializes session context.
3. `invoke()` or `stream()` forwards the tool request through the relevant transport layer.
4. Session and request helpers normalize the response into Python-friendly objects.

### Provider mode

1. User creates or reuses an explicit `ProviderRuntime` from an `Toolplane` or `ToolplaneHTTP` client.
2. `ProviderRuntime.create_session()` or `ProviderRuntime.attach_session()` establishes machine ownership for the target session.
3. User decorates local callables with `ProviderRuntime.tool()`.
4. `ProviderRuntime.start_in_background()` or `ProviderRuntime.run_forever()` starts heartbeats and the provider poll loop.
5. Registered Python callables execute locally and results are sent back to the server.

## Key Folders

| Path | Responsibility |
| --- | --- |
| `toolplane/core/` | gRPC-side connection, error, machine, request, session, and tool primitives |
| `toolplane/http_core/` | HTTP-side connection and session implementations |
| `toolplane/common/` | Shared configs, base managers, retries, validation, and cache helpers |
| `toolplane/factories/` | Factory helpers used to compose transport-specific components |
| `toolplane/interfaces/` | Interface and protocol contracts for client modules |
| `toolplane/session/` | Session context models used by both transports |
| `toolplane/toolkits/` | Toolkit-oriented helpers layered on top of the SDK |
| `toolplane/utils/` | General utility helpers |
| `toolplane/proto/` | Python protobuf outputs used by the gRPC facade |

## Notes For Agents

- When comparing SDK capability, use this client as the primary reference before checking Go or TypeScript.
- Public behavior is concentrated in `toolplane_client.py` and `toolplane_http_client.py`; lower-level folders mostly exist to support those facades.
- Provider execution is now explicit in `provider_runtime.py`; do not assume that `connect()` or consumer-side `create_session()` implies machine registration or background polling.
- If a feature appears in one Python transport facade, check the other facade before assuming parity is missing.
