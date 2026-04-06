# Toolplane Python Client Documentation

> **This is the advanced API reference.** If you are new to Toolplane, start with the canonical first-touch path in `README.md` and the example walkthrough in `README_EXAMPLES.md`. Come back here for detailed constructor parameters, method signatures, and configuration options.

This manual documents the Python SDK as the primary maintained client for the Toolplane control plane. The maintained story is gRPC control-plane execution first, with `ToolplaneHTTP` available as a compatibility gateway surface over the same session, machine, request, and task flows.

## Overview

This documentation covers the Modular Remote Procedure Call (Toolplane) Python client implementation. The client allows you to interact with Toolplane servers to register tools, manage sessions, and execute distributed functions.

## Architecture

The Toolplane Python client is organized into several key components:

- **Core Components**: Connection management, machine management, tool management, and request handling
- **Session Management**: Handles multiple session contexts
- **HTTP Client Alternative**: Provides HTTP-based communication as an alternative to gRPC

## Installation

```bash
pip install -r requirements.txt
```

## Quick Start Example

```python
from toolplane.toolplane_client import Toolplane

# Initialize client
client = Toolplane(
    server_host="localhost",
    server_port=80,
    session_ids=["session-123"],
    user_id="user-456",
    api_key="your-api-key"
)

# Connect to server
client.connect()

# Register a tool
@client.tool("session-123", name="echo_tool")
def echo_tool(input_str: str) -> str:
    return f"Echo: {input_str}"

# Start the client
client.start()
```

## Core Classes

### Toolplane (Main Client)

The main Toolplane client class that manages connections, sessions, and tools.

#### Constructor Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `server_host` | str | Hostname of the Toolplane server |
| `server_port` | int | Port of the Toolplane server |
| `session_ids` | List[str] | List of session IDs to manage |
| `api_key` | Optional[str] | API key for authentication |
| `user_id` | Optional[str] | User identifier |
| `session_name` | Optional[str] | Name for the session |
| `session_description` | Optional[str] | Description of the session |
| `session_namespace` | Optional[str] | Namespace for the session |
| `heartbeat_interval` | int | Interval for heartbeat messages (seconds) |
| `max_workers` | int | Maximum number of worker threads |
| `request_timeout` | int | Request timeout in seconds |
| `poll_interval` | float | Interval for polling requests (seconds) |
| `max_retries` | int | Maximum number of retry attempts for connection failures |
| `retry_base_delay` | float | Initial delay between retries (seconds) |
| `retry_max_delay` | float | Maximum delay between retries (seconds) |
| `retry_backoff_factor` | float | Exponential backoff factor |

#### Methods

##### `connect() -> bool`
Establishes connection to the Toolplane server.

##### `disconnect() -> None`
Closes connection to the Toolplane server.

##### `create_session(session_id: Optional[str] = None, user_id: Optional[str] = None, name: Optional[str] = None, description: Optional[str] = None, namespace: Optional[str] = None, register_machine: bool = True) -> SessionContext`
Creates a new session on the server.

##### `get_session(session_id: str) -> Optional[SessionContext]`
Retrieves a session context by ID.

##### `list_sessions() -> List[SessionContext]`
Lists all active session contexts.

##### `tool(session_id: str, name: Optional[str] = None, description: Optional[str] = None, stream: bool = False, tags: Optional[List[str]] = None) -> Callable[[Callable], Callable]`
Decorator to register a tool for a session.

##### `invoke(tool_name: str, session_id: str, **params) -> Any`
Invokes a tool in a session synchronously.

##### `ainvoke(tool_name: str, session_id: str, **params) -> str`
Invokes a tool asynchronously.

##### `stream(tool_name: str, callback: Callable[[Any, bool], None], session_id: str, **params) -> List[Any]`
Streams tool execution results.

##### `astream(tool_name: str, callback: Callable[[Any, bool], None], session_id: str, **params) -> List[Any]`
Alias for stream method.

##### `get_available_tools(session_id: str) -> Dict[str, Any]`
Gets available tools for a session.

##### `get_request_status(request_id: str, session_id: str) -> Dict[str, Any]`
Gets status of a request.

##### `start() -> None`
Starts the client and begins polling for requests.

##### `stop() -> None`
Stops the client and cleans up resources.

### ToolplaneHTTP (HTTP Client Alternative)

An HTTP-based alternative to the gRPC client.

#### Constructor Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `server_host` | str | Hostname of the Toolplane server |
| `server_port` | int | Port of the Toolplane server |
| `session_ids` | List[str] | List of session IDs to manage |
| `api_key` | Optional[str] | API key for authentication |
| `user_id` | Optional[str] | User identifier |
| `session_name` | Optional[str] | Name for the session |
| `session_description` | Optional[str] | Description of the session |
| `session_namespace` | Optional[str] | Namespace for the session |
| `max_buffer_size` | int | Maximum buffer size for HTTP requests |
| `max_retries` | int | Maximum number of retry attempts |
| `request_timeout` | int | Request timeout in seconds |
| `retry_backoff_ms` | int | Initial backoff time for retries (milliseconds) |
| `heartbeat_interval` | int | Interval for heartbeat messages (seconds) |
| `max_workers` | int | Maximum number of worker threads |
| `poll_interval` | float | Interval for polling requests (seconds) |

#### Methods

Same as Toolplane client with the same signatures and functionality.

## Session Management

### SessionContext

Represents a session context with its own tools and machine registration.

#### Methods

##### `register_machine() -> bool`
Registers a machine for this session.

##### `register_tool(name: str, func: Callable, schema: Optional[Dict] = None, description: Optional[str] = None, stream: bool = False, tags: Optional[List[str]] = None) -> None`
Registers a tool for this session.

##### `invoke(tool_name: str, **params) -> Any`
Invokes a tool in this session synchronously.

##### `ainvoke(tool_name: str, **params) -> str`
Invokes a tool asynchronously.

##### `stream(tool_name: str, callback: Callable[[Any, bool], None], **params)`
Streams tool execution.

##### `get_available_tools() -> Dict[str, Any]`
Gets available tools for this session.

##### `get_request_status(request_id: str) -> Dict[str, Any]`
Gets status of a request.

## Tool Registration

Tools are registered using the `@client.tool` decorator:

```python
@client.tool("session-123", name="my_tool", description="A sample tool")
def my_tool(param1: str, param2: int) -> str:
    return f"Processed {param1} with {param2}"
```

### Tool Registration Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `session_id` | str | ID of the session to register the tool for |
| `name` | Optional[str] | Tool name (defaults to function name) |
| `description` | Optional[str] | Tool description |
| `stream` | bool | Whether tool supports streaming |
| `tags` | Optional[List[str]] | Tags for categorizing the tool |

## Error Handling

The client raises `ToolplaneError` for most operation failures and `ConnectionError` for connection-related issues.

## Type Safety

All public APIs are fully typed with proper type annotations to enable:
- Better IDE support (autocomplete, error detection)
- Static type checking with tools like `mypy`
- Self-documenting code

## Retry Logic

Both Toolplane clients now include comprehensive retry logic to handle transient failures:

### gRPC Client Retry Configuration

```python
# Example with custom retry configuration
client = Toolplane(
    server_host="localhost",
    server_port=80,
    max_retries=5,
    retry_base_delay=2.0,
    retry_max_delay=120.0,
    retry_backoff_factor=3.0
)
```

### HTTP Client Retry Configuration

```python
# Example with custom retry configuration
client = ToolplaneHTTP(
    server_host="localhost",
    server_port=8080,
    max_retries=5,
    retry_backoff_ms=500
)
```

### Retry Features

1. **Exponential Backoff**: Delays increase exponentially with each retry attempt
2. **Jitter**: Random variation to prevent thundering herds
3. **Configurable Limits**: Users can tune retry behavior to their needs
4. **Smart Error Detection**: Retries only on retryable errors (UNAVAILABLE, DEADLINE_EXCEEDED, etc.)
5. **Connection State Tracking**: Monitor connection health and recovery attempts

### Retry Parameters

| Parameter | Default Value | Description |
|-----------|---------------|-------------|
| `max_retries` | 3 | Maximum retry attempts |
| `retry_base_delay` | 1.0 seconds | Initial delay between retries |
| `retry_max_delay` | 60.0 seconds | Maximum delay between retries |
| `retry_backoff_factor` | 2.0 | Exponential backoff factor |
| `retry_backoff_ms` | 250 milliseconds | HTTP client initial backoff |

## Usage Examples

### Basic Usage

```python
from toolplane.toolplane_client import Toolplane

# Initialize client
client = Toolplane(
    server_host="localhost",
    server_port=80,
    session_ids=["session-123"],
    user_id="user-456",
    api_key="your-api-key"
)

# Connect to server
client.connect()

# Register tools
@client.tool("session-123", name="echo")
def echo(input_str: str) -> str:
    return f"Echo: {input_str}"

# Start client
client.start()
```

### Asynchronous Tool Invocation

```python
# Invoke tool asynchronously
request_id = client.ainvoke("echo", session_id="session-123", input_str="Hello World")
print(f"Request ID: {request_id}")
```

### Streaming Tool Execution

```python
def callback(chunk, is_final):
    print(f"Chunk: {chunk}")

# Stream tool execution
chunks = client.stream("echo", callback, session_id="session-123", input_str="Hello World")
print(f"Final result: {chunks}")
```

## Configuration

### Environment Variables

The client supports configuration through environment variables:
- `TOOLPLANE_SERVER_HOST` - Server hostname
- `TOOLPLANE_SERVER_PORT` - Server port
- `TOOLPLANE_API_KEY` - API key for authentication
- `TOOLPLANE_USER_ID` - User identifier

### Configuration Priority

1. Explicit constructor parameters
2. Environment variables
3. Default values

## Performance Considerations

- Uses thread pools for concurrent request handling
- Configurable worker count for heavy workloads
- Efficient connection reuse
- Heartbeat mechanism for connection liveness

## Troubleshooting

### Connection Issues

If you encounter connection problems:
1. Verify server address and port
2. Check network connectivity
3. Ensure the server is running
4. Confirm credentials are valid

### Timeout Issues

If operations timeout:
1. Increase `request_timeout` parameter
2. Check server performance
3. Monitor network latency

## Development Guidelines

### Code Style

- Follow PEP 8 guidelines
- Use descriptive variable names
- Include comprehensive docstrings
- Maintain type safety with proper annotations

### Testing

Write unit tests for all public APIs:
- Test connection establishment
- Test tool registration
- Test invocation scenarios
- Test error conditions

## Future Enhancements

### Planned Features

1. **Automatic Reconnection**: Built-in reconnection logic with exponential backoff
2. **Enhanced Logging**: More detailed logging for debugging
3. **Improved Retry Logic**: More sophisticated retry mechanisms
4. **Better Session Management**: Enhanced session lifecycle handling
5. **Async/Await Support**: Native async/await patterns for Python 3.7+

## Contributing

1. Fork the repository
2. Create feature branch
3. Write tests for new functionality
4. Submit pull request with clear description