# Toolplane Python Client

The Python package is the primary maintained SDK for Toolplane's durable remote tool-execution control plane. Python is the richest current client surface and the baseline for the repo's end-to-end provider, consumer, and admin capability.

## Decision Rule

Use Toolplane when one remote tool may outlive the caller, needs inspection or bounded replay after disconnect, needs explicit provider ownership or drain behavior, or is queue-backed enough that request lifecycle control matters. If the work is quick, in-process, same-lifecycle, and not operationally sensitive, direct tool calling is simpler.

A concrete first offload candidate is one sandboxed code-execution worker. The Python provider-consumer path is the maintained starting point for that shape because it covers explicit provider ownership, request lifecycle control, and the maintained HTTP gateway compatibility layer on the richest SDK surface.

## Support Status

- Primary maintained SDK and current completeness baseline.
- gRPC is the primary transport for the repo's control-plane story.
- `ToolplaneHTTP` is a maintained compatibility surface over the same session, tool, machine, request, and task flows.
- Python exposes a broader public surface than the current Go and TypeScript SDKs; confirm cross-SDK portability in [SDK_MAP.md](../../SDK_MAP.md) before assuming parity.

## First Offload Path

Start with the maintained provider-consumer pair before exploring the full API. The sample tools are intentionally simple, but this is the same lifecycle you would reuse to offload one sandboxed code-execution tool or another environment-bound worker:

1. Run `example_client.py` — connects via gRPC, creates a session through the explicit `ProviderRuntime`, registers machine-backed tools, and starts the provider loop.
2. Copy the printed `TOOLPLANE_SESSION_ID`.
3. Run `example_user.py` with that session ID — lists tools, invokes provider-backed work, and polls request state.

See [README_EXAMPLES.md](README_EXAMPLES.md) for environment defaults and the full example flow. For runtime semantics — request lifecycle, streaming, recovery windows, machine drain — see [server/DOCUMENTATION.md](../../server/DOCUMENTATION.md).

### Scope Categories

The Python public surface spans three scope categories (see [SDK_MAP.md](../../SDK_MAP.md) for the full matrix):

- **Consumer scope** — session/machine/task lifecycle, tool discovery, and remote invocation (`create_session`, `invoke`, `stream`, etc.). These methods are portable across maintained SDKs.
- **Provider scope** — tool registration, request claiming, heartbeat, and result submission through the explicit `ProviderRuntime` surface. The client also exposes convenience aliases `tool()`, `start()`, and `stop()` over that same runtime surface.
- **Admin scope** — session administration helpers (`list_user_sessions`, `bulk_delete_sessions`, `get_session_stats`, `invalidate_session`). These are currently exposed only in the Python SDK.

## Installation

The package is not published to PyPI yet; install it from the repository.

```bash
# Clone the repository
git clone https://github.com/Abhishek-chohan/tool-control-plane.git
cd tool-control-plane/clients/python-client

# Install in development mode
pip install -e .

# Install with all development dependencies
pip install -e ".[dev]"
```

Requirements: Python 3.8+ with grpcio, protobuf, requests, googleapis-common-protos, and pydantic. The optional toolkits under `toolplane/toolkits/` have heavier dependencies; install their requirements only if you use them. `grpcio-tools` is needed only for proto regeneration (`cd server && make gen-proto-python`).

## Quick Start

### Consumer Usage with gRPC Client

```python
from toolplane import Toolplane

# Initialize a consumer-oriented client
client = Toolplane(
    server_host="localhost",
    server_port=9001,
    api_key="your-api-key-here",
)

client.connect()

# Attach to an existing provider-backed session
session_id = "provider-session-123"
tools = client.get_available_tools(session_id)
print(tools)

# Execute a tool synchronously
result = client.invoke("echo-tool", session_id, text="Hello World")
print(result)  # Output: Echo: Hello World

# Submit a tool invocation without blocking; awaits the request ID
import asyncio
request_id = asyncio.run(
    client.ainvoke("echo-tool", session_id, text="Async Test")
)
print(f"Request ID: {request_id}")
# Poll client.get_request_status(session_id, request_id) for the outcome.

# Stream tool execution
def stream_callback(chunk, is_final):
    if is_final:
        print("Stream completed")
    else:
        print(f"Received chunk: {chunk}")

client.stream("echo-tool", stream_callback, session_id, text="Streaming Test")
```

### Explicit Provider Runtime with gRPC Client

```python
from toolplane import Toolplane

client = Toolplane(
    server_host="localhost",
    server_port=9001,
    api_key="your-api-key-here",
    user_id="provider-user",
)

provider = client.provider_runtime()
session = provider.create_session(
    name="My Provider Session",
    description="Session backed by the explicit provider runtime",
    namespace="development",
)

@provider.tool(
    session_id=session.session_id,
    name="echo-tool",
    description="Echo back the input text",
    tags=["utility", "testing"],
)
def echo_tool(text: str) -> str:
    return f"Echo: {text}"

provider.run_forever()
```

### Explicit Provider Runtime with HTTP Client

```python
from toolplane import ToolplaneHTTP

# Initialize HTTP client against the gateway (default port 8080)
client = ToolplaneHTTP(
    server_host="localhost",
    server_port=8080,
    api_key="your-api-key-here",
)

# Create an explicit provider runtime
provider = client.provider_runtime()
session = provider.create_session(
    user_id="user123",
    name="My HTTP Provider Session",
)

# Register a tool
@provider.tool(
    session_id=session.session_id,
    name="http-tool",
    description="HTTP-based tool execution",
)
def http_tool(text: str) -> str:
    return f"HTTP Processed: {text}"

provider.run_forever()
```

Tools are session-scoped: a consumer can only invoke tools registered into the same session, so separate provider and consumer processes must share a session ID (this is what the examples arrange through `TOOLPLANE_SESSION_ID`).

## Session Management

```python
# Consumer-side session creation does not imply machine registration
session = client.create_session(
    session_id="my-unique-session-id",
    user_id="user123",
    name="My Application Session",
    description="Session for my application",
    namespace="production",
    register_machine=False,
)

# Provider-side session creation uses the explicit runtime and attaches a machine
provider = client.provider_runtime()
provider_session = provider.create_session(
    user_id="user123",
    name="Provider Session",
    namespace="production",
)

# Get an existing session
session_context = client.get_session("my-unique-session-id")

# List user sessions with pagination (page_token is the opaque cursor
# returned by the previous page; it is a string, not an offset)
user_sessions = client.list_user_sessions(
    user_id="user123",
    page_size=10,
    page_token="",
)

# Get session statistics
stats = client.get_session_stats(user_id="user123")
print(f"Total sessions: {stats['total_sessions']}")

# Bulk delete sessions
bulk_delete_result = client.bulk_delete_sessions(
    user_id="user123",
    session_ids=["session-1", "session-2"],
)

# Invalidate a session (revokes every live API key of the session)
success = client.invalidate_session("session-123", reason="Session expired")
print(f"Session invalidated: {success}")
```

## Tool Management

```python
# Registration uses the explicit provider runtime
provider = client.provider_runtime(["session-123"])
provider.attach_session("session-123", register_machine=True)

# Register a tool with automatic schema generation
@provider.tool(
    session_id="session-123",
    name="process-data",
    description="Process and transform data",
    tags=["data-processing", "utility"],
)
def process_data(input: str, format: str = "json") -> dict:
    return {"input": input, "format": format, "processed": True}

# Register a streaming tool
@provider.tool(
    session_id="session-123",
    name="stream-data",
    description="Stream data processing",
    stream=True,
)
def stream_data(input: str):
    for char in input:
        yield f"Processing: {char}"

# Discover tools in the session
available_tools = client.get_available_tools("session-123")
print(f"Available tools: {available_tools}")
```

## Machine Management

```python
# Consumer sessions are not machine-backed by default; attach through
# the provider runtime to register a machine
provider = client.provider_runtime(["session-123"])
session_context = provider.attach_session("session-123", register_machine=True)
machine_id = session_context.machine_id
print(f"Registered machine ID: {machine_id}")

# Public machine-management helpers stay on the client
machine_status = client.get_machine("session-123", machine_id)
print(f"Machine status: {machine_status}")

# Drain the machine gracefully before shutdown, then unregister
client.drain_machine("session-123", machine_id)
client.unregister_machine("session-123", machine_id)
```

Heartbeats are owned by the explicit provider runtime (`heartbeat_interval` on the client constructor configures the cadence); you never send them yourself.

## Request Processing

```python
# Create a request directly (the same path invoke() uses)
request_id = client.create_request(
    session_id="session-123",
    tool_name="process-data",
    input_data='{"input": "test", "format": "json"}',
)

# Poll request status: arguments are (session_id, request_id)
status = client.get_request_status("session-123", request_id)
print(f"Request status: {status['status']}")

# List requests in a session; page_token is the cursor from the previous page
requests = client.list_requests("session-123", status="done", limit=10, page_token="")

# Cancel a request
client.cancel_request("session-123", request_id)
```

## Configuration Parameters

#### `Toolplane` (gRPC facade)

| Parameter | Type | Default | Description |
| --- | --- | --- | --- |
| `server_host` | str | "localhost" | Server hostname |
| `server_port` | int | 9001 | Server port |
| `use_tls` | bool | False | Enable TLS encryption |
| `tls_cert_path` | str | None | Optional client TLS certificate for mutual TLS |
| `tls_key_path` | str | None | Optional client TLS private key for mutual TLS |
| `tls_ca_cert_path` | str | None | Optional CA bundle for direct gRPC TLS |
| `tls_server_name` | str | None | Optional TLS server name override |
| `api_key` | str | None | Authentication API key |
| `user_id` | str | None | User identifier |
| `session_name` | str | None | Session name |
| `session_description` | str | None | Session description |
| `session_namespace` | str | None | Session namespace |
| `heartbeat_interval` | int | 60 | Heartbeat interval in seconds |
| `max_workers` | int | 10 | Maximum worker threads |
| `request_timeout` | int | 30 | Request timeout in seconds |
| `max_retries` | int | 3 | Maximum retry attempts |
| `retry_base_delay` | float | 1.0 | Base retry delay in seconds |
| `retry_max_delay` | float | 60.0 | Maximum retry delay in seconds |
| `retry_backoff_factor` | float | 2.0 | Exponential backoff factor |
| `debug` | bool | False | Enable debug logging |
| `log_level` | str | "INFO" | Logging level |

#### `ToolplaneHTTP` (HTTP gateway facade)

| Parameter | Type | Default | Description |
| --- | --- | --- | --- |
| `server_host` | str | "localhost" | Gateway hostname |
| `server_port` | int | 8080 | Gateway port |
| `api_key` | str | None | Authentication API key |
| `user_id` | str | None | User identifier |
| `session_name` | str | None | Session name |
| `session_description` | str | None | Session description |
| `session_namespace` | str | None | Session namespace |
| `max_buffer_size` | int | 4194304 | Buffer size (4MB) |
| `max_retries` | int | 3 | Maximum retry attempts |
| `request_timeout` | int | 30 | Request timeout in seconds |
| `retry_backoff_ms` | int | 250 | Retry backoff in milliseconds |
| `heartbeat_interval` | int | 60 | Heartbeat interval in seconds |
| `max_workers` | int | 10 | Maximum worker threads |

Configuration commonly comes from the environment; the examples read `TOOLPLANE_SERVER_HOST`, `TOOLPLANE_SERVER_PORT`, `TOOLPLANE_API_KEY`, `TOOLPLANE_USER_ID`, `TOOLPLANE_USE_TLS`, `TOOLPLANE_TLS_CA_CERT_PATH`, and `TOOLPLANE_TLS_SERVER_NAME`, defaulting the port to 9001 and the API key to `toolplane-conformance-fixture-key`.

## Error Handling

The client raises a typed hierarchy rooted at `ToolplaneError`:

```python
from toolplane import ToolplaneError, ToolplaneNotFoundError

try:
    client.invoke("missing-tool", session_id)
except ToolplaneNotFoundError as e:
    print(f"Tool not found: {e}")
except ToolplaneError as e:
    print(f"Toolplane error: {e}")
```

| Exception | Raised for |
| --- | --- |
| `ToolplaneError` | Base class for every client error |
| `ConnectionError` | The server is unreachable or the channel dropped |
| `ToolError` / `SessionError` / `MachineError` / `RequestError` / `TaskError` | Domain-level failures in the matching resource |
| `ToolplaneAPIError` | Base for gRPC status-mapped server errors |
| `ToolplaneNotFoundError`, `ToolplaneInvalidArgumentError`, `ToolplaneFailedPreconditionError`, `ToolplaneResourceExhaustedError`, `ToolplaneUnauthenticatedError`, `ToolplanePermissionDeniedError`, `ToolplaneAlreadyExistsError`, `ToolplaneTimeoutError`, `ToolplaneUnavailableError`, `ToolplaneCancelledError`, `ToolplaneInternalError` | The matching gRPC status returned by the server |

## API Reference

<!-- BEGIN GENERATED: api-surface -- tooling: tools/gen_sdk_readmes.py; edits inside this block are overwritten -->
#### `Toolplane`

Public methods parsed from `toolplane/toolplane_client.py`:

| Method | Description |
| --- | --- |
| `connect() -> bool` | Connect to Toolplane server. |
| `disconnect() -> None` | Disconnect from Toolplane server. |
| `ensure_session_context(session_id: str, create_if_missing: bool=False, register_machine: bool=False) -> SessionContext` | Ensure a local session context exists without implying provider startup. |
| `create_session(session_id: Optional[str]=None, user_id: Optional[str]=None, name: Optional[str]=None, description: Optional[str]=None, namespace: Optional[str]=None, register_machine: bool=False) -> SessionContext` | Create a new session. |
| `get_session(session_id: str) -> Optional[SessionContext]` | Get session context by ID. |
| `list_sessions() -> List[SessionContext]` | List all session contexts. |
| `list_user_sessions(user_id: str, page_size: int=10, page_token: str='', filter: str='') -> Dict[str, Any]` | List user sessions with pagination and filtering. |
| `bulk_delete_sessions(user_id: str, session_ids: Optional[List[str]]=None, filter: str='') -> Dict[str, Any]` | Bulk delete sessions for a user. |
| `get_session_stats(user_id: str) -> Dict[str, int]` | Get session statistics for a user. |
| `invalidate_session(session_id: str, reason: str='') -> bool` | Invalidate a session. |
| `update_session(session_id: str, name: Optional[str]=None, description: Optional[str]=None, namespace: Optional[str]=None) -> Dict[str, Any]` | Update session metadata. |
| `create_request(session_id: str, tool_name: str, input_data: str, idempotency_key: str='') -> str` | Create a new request in a session. |
| `list_requests(session_id: str, status: str='', tool_name: str='', limit: int=10, page_token: str='') -> List[Dict[str, Any]]` | List requests in a session. |
| `list_requests_page(session_id: str, status: str='', tool_name: str='', limit: int=10, page_token: str='') -> Dict[str, Any]` | List one page of requests, with the continuation cursor. |
| `cancel_request(session_id: str, request_id: str) -> bool` | Cancel a request in a session. |
| `create_task(session_id: str, tool_name: str, input_data: str) -> Dict[str, Any]` | Create a new task in a session. |
| `get_task(session_id: str, task_id: str) -> Dict[str, Any]` | Get a task by ID in a session. |
| `list_tasks(session_id: str) -> List[Dict[str, Any]]` | List tasks in a session. |
| `cancel_task(session_id: str, task_id: str) -> bool` | Cancel a task in a session. |
| `list_machines(session_id: str) -> List[Dict[str, Any]]` | List machines in a session. |
| `get_machine(session_id: str, machine_id: str) -> Dict[str, Any]` | Get a machine by ID. |
| `unregister_machine(session_id: str, machine_id: Optional[str]=None) -> bool` | Unregister a machine from a session. |
| `drain_machine(session_id: str, machine_id: Optional[str]=None) -> bool` | Drain a machine from a session. |
| `create_api_key(session_id: str, name: str, capabilities: List[str]) -> Dict[str, Any]` | Create a new API key for a session. |
| `list_api_keys(session_id: str) -> List[Dict[str, Any]]` | List active API keys for a session. |
| `revoke_api_key(session_id: str, key_id: str) -> bool` | Revoke an API key for a session. |
| `tool(session_id: str, name: Optional[str]=None, description: Optional[str]=None, stream: bool=False, tags: Optional[List[str]]=None) -> Callable[[Callable], Callable]` | Decorator to register a tool for a session. |
| `invoke(tool_name: str, session_id: str, **params) -> Any` | Invoke a tool in a session. |
| `async ainvoke(tool_name: str, session_id: str, **params) -> str` | Submit a tool invocation without blocking; awaits the request ID. |
| `stream(tool_name: str, callback: Callable[[Any, bool], None], session_id: str, **params) -> List[Any]` | Stream tool execution. |
| `async astream(tool_name: str, callback: Callable[[Any, bool], None], session_id: str, **params) -> List[Any]` | Awaitable stream: resolves with the collected chunks. |
| `get_available_tools(session_id: str) -> Dict[str, Any]` | Get available tools for a session. |
| `list_tools(session_id: str) -> List[Dict[str, Any]]` | List tools for a session. |
| `get_tool_by_id(session_id: str, tool_id: str) -> Dict[str, Any]` | Get a tool by ID. |
| `get_tool_by_name(session_id: str, tool_name: str) -> Dict[str, Any]` | Get a tool by name. |
| `delete_tool(session_id: str, tool_id: str) -> bool` | Delete a tool by ID. |
| `get_request_status(session_id: str, request_id: str) -> Dict[str, Any]` | Get request status (session-scoped, like every other facade method). |
| `provider_runtime(session_ids: Optional[List[str]]=None)` | Return the explicit provider runtime for this client. |
| `start() -> None` | Backward-compatible alias for the explicit provider runtime. |
| `stop() -> None` | Stop the explicit provider runtime if it is active. |

#### `ToolplaneHTTP`

Public methods parsed from `toolplane/toolplane_http_client.py`:

| Method | Description |
| --- | --- |
| `connect() -> bool` | Connect to HTTP server. |
| `disconnect()` | Disconnect from HTTP server. |
| `ensure_session_context(session_id: str, create_if_missing: bool=False, register_machine: bool=False) -> HTTPSessionContext` | Ensure a local session context exists without implying provider startup. |
| `create_session(session_id: Optional[str]=None, user_id: Optional[str]=None, name: Optional[str]=None, description: Optional[str]=None, namespace: Optional[str]=None, register_machine: bool=False) -> HTTPSessionContext` | Create a new session. |
| `get_session(session_id: str) -> Optional[HTTPSessionContext]` | Get session context by ID. |
| `list_sessions() -> List[HTTPSessionContext]` | List all session contexts. |
| `get_primary_session_context() -> Optional[HTTPSessionContext]` | Get primary session context. |
| `tool(session_id: str, name: Optional[str]=None, description: Optional[str]=None, stream: bool=False, tags: Optional[List[str]]=None)` | Decorator to register a tool for a session. |
| `invoke(tool_name: str, session_id: str, **params) -> Any` | Invoke a tool in a session. |
| `async ainvoke(tool_name: str, session_id: str, **params) -> str` | Submit a tool invocation without blocking; awaits the request ID. |
| `stream(tool_name: str, callback: Callable[[Any, bool], None], session_id: str, **params) -> List[Any]` | Stream tool execution. |
| `async astream(tool_name: str, callback: Callable[[Any, bool], None], session_id: str, **params) -> List[Any]` | Awaitable stream: resolves with the collected chunks. |
| `get_available_tools(session_id: str) -> Dict[str, Any]` | Get available tools for a session. |
| `list_tools(session_id: str) -> List[Dict[str, Any]]` | List tools for a session. |
| `get_tool_by_id(session_id: str, tool_id: str) -> Dict[str, Any]` | Get a tool by ID. |
| `get_tool_by_name(session_id: str, tool_name: str) -> Dict[str, Any]` | Get a tool by name. |
| `delete_tool(session_id: str, tool_id: str) -> bool` | Delete a tool by ID. |
| `get_request_status(session_id: str, request_id: str) -> Dict[str, Any]` | Get request status (session-scoped, like every other facade method). |
| `list_user_sessions(user_id: str, page_size: int=10, page_token: str='', filter: str='') -> Dict[str, Any]` | List user sessions with pagination and filtering. |
| `bulk_delete_sessions(user_id: str, session_ids: List[str]=None, filter: str='') -> Dict[str, Any]` | Bulk delete sessions for a user. |
| `get_session_stats(user_id: str) -> Dict[str, int]` | Get session statistics for a user. |
| `invalidate_session(session_id: str, reason: str='') -> bool` | Invalidate a session. |
| `update_session(session_id: str, name: Optional[str]=None, description: Optional[str]=None, namespace: Optional[str]=None) -> Dict[str, Any]` | Update session metadata. |
| `create_request(session_id: str, tool_name: str, input_data: str, idempotency_key: str='') -> str` | Create a new request in a session. |
| `list_requests(session_id: str, status: str='', tool_name: str='', limit: int=10, page_token: str='') -> List[Dict[str, Any]]` | List requests in a session. |
| `list_requests_page(session_id: str, status: str='', tool_name: str='', limit: int=10, page_token: str='') -> Dict[str, Any]` | List one page of requests, with the continuation cursor. |
| `cancel_request(session_id: str, request_id: str) -> bool` | Cancel a request in a session. |
| `create_task(session_id: str, tool_name: str, input_data: str) -> Dict[str, Any]` | Create a new task in a session. |
| `get_task(session_id: str, task_id: str) -> Dict[str, Any]` | Get a task by ID in a session. |
| `list_tasks(session_id: str) -> List[Dict[str, Any]]` | List tasks in a session. |
| `cancel_task(session_id: str, task_id: str) -> bool` | Cancel a task in a session. |
| `list_machines(session_id: str) -> List[Dict[str, Any]]` | List machines in a session. |
| `get_machine(session_id: str, machine_id: str) -> Dict[str, Any]` | Get a machine by ID. |
| `unregister_machine(session_id: str, machine_id: Optional[str]=None) -> bool` | Unregister a machine from a session. |
| `drain_machine(session_id: str, machine_id: Optional[str]=None) -> bool` | Drain a machine from a session. |
| `create_api_key(session_id: str, name: str, capabilities: List[str]) -> Dict[str, Any]` | Create a new API key for a session. |
| `list_api_keys(session_id: str) -> List[Dict[str, Any]]` | List active API keys for a session. |
| `revoke_api_key(session_id: str, key_id: str) -> bool` | Revoke an API key for a session. |
| `provider_runtime(session_ids: Optional[List[str]]=None)` | Return the explicit provider runtime for this client. |
| `start()` | Backward-compatible alias for the explicit provider runtime. |
| `stop()` | Stop the explicit provider runtime if it is active. |
| `health()` | Check server health. |

#### `ProviderRuntime`

Public methods parsed from `toolplane/provider_runtime.py`:

| Method | Description |
| --- | --- |
| `running() -> bool` |  |
| `add_sessions(session_ids: Optional[Iterable[str]]) -> None` |  |
| `managed_session_ids() -> List[str]` |  |
| `attach_session(session_id: str, register_machine: bool=True)` |  |
| `create_session(session_id: Optional[str]=None, user_id: Optional[str]=None, name: Optional[str]=None, description: Optional[str]=None, namespace: Optional[str]=None, register_machine: bool=True)` |  |
| `register_tool(session_id: str, name: str, func: Callable, schema: Optional[dict]=None, description: Optional[str]=None, stream: bool=False, tags: Optional[List[str]]=None) -> Callable` |  |
| `tool(session_id: str, name: Optional[str]=None, description: Optional[str]=None, stream: bool=False, tags: Optional[List[str]]=None)` |  |
| `poll_once() -> None` |  |
| `start_in_background(session_ids: Optional[Iterable[str]]=None) -> 'ProviderRuntime'` |  |
| `run_forever(session_ids: Optional[Iterable[str]]=None) -> None` |  |
| `stop() -> None` |  |
| `close() -> None` |  |
<!-- END GENERATED: api-surface -->

## Contributing

Contributions are welcome — see the [repository-level README](../../README.md) Contributing section for the required checks. Run the Python suites locally with:

```bash
cd server
make python-unit            # unit tests
make conformance-python     # shared-fixture conformance
```

## Support

- Advanced API reference: [MANUAL.md](MANUAL.md)
- Example walkthrough: [README_EXAMPLES.md](README_EXAMPLES.md)
- Package architecture: [ARCHITECTURE.md](ARCHITECTURE.md)
- Issues: [open one on the GitHub repository](https://github.com/Abhishek-chohan/tool-control-plane/issues)
