# Toolplane Python Client

The Python package is the primary maintained SDK for Toolplane's durable remote tool-execution control plane. Toolplane is built for remote tools that need explicit provider ownership, request lifecycle control, retained-window recovery, and deploy-safe drain over the canonical contract in `server/proto/service.proto`. Python is the richest current client surface and the baseline for end-to-end platform capability.

## When To Start Here

- You want the clearest first-touch path for Toolplane's full provider-consumer lifecycle.
- You need the richest maintained provider runtime plus the maintained HTTP gateway compatibility surface.
- You want the repo's baseline surface before comparing narrower Go or TypeScript projections in `SDK_MAP.md`.

## Support Status

- Primary maintained SDK and current completeness baseline.
- gRPC is the primary transport for the repo's control-plane story.
- `ToolplaneHTTP` is a maintained compatibility surface over the same session, tool, machine, request, and task flows.
- Python exposes a broader public surface than the current Go and TypeScript SDKs; confirm cross-SDK portability in `SDK_MAP.md` before assuming parity.

## Canonical First-Touch Path

Start with the maintained provider-consumer pair before exploring the full API:

1. Run `example_client.py` — connects via gRPC, creates a session through the explicit `ProviderRuntime`, registers machine-backed tools, and starts the provider loop.
2. Copy the printed `TOOLPLANE_SESSION_ID`.
3. Run `example_user.py` with that session ID — lists tools, invokes provider-backed work, and polls request state.

See `README_EXAMPLES.md` for environment defaults and the full example flow. For runtime semantics — request lifecycle, streaming, recovery windows, machine drain — see `server/DOCUMENTATION.md`.

### Scope Categories

The Python public surface spans three scope categories (see `SDK_MAP.md` for the full matrix):

- **Consumer scope** — session/machine/task lifecycle, tool discovery, and remote invocation (`create_session`, `invoke`, `stream`, etc.). These methods are portable across maintained SDKs.
- **Provider scope** — tool registration, request claiming, heartbeat, and result submission through the explicit `ProviderRuntime` surface (`provider_runtime()`, `ProviderRuntime.create_session()`, `ProviderRuntime.attach_session()`, `ProviderRuntime.tool()`, `ProviderRuntime.start_in_background()`, `ProviderRuntime.run_forever()`, `ProviderRuntime.stop()`). The client also exposes convenience aliases `tool()`, `start()`, and `stop()` over that same runtime surface.
- **Admin scope** — session administration helpers (`list_user_sessions`, `bulk_delete_sessions`, `get_session_stats`, `refresh_session_token`, `invalidate_session`). These are currently exposed only in the Python SDK.

## Overview

Use the Python client when you need the richest maintained Toolplane surface:

- Provider registration and machine lifecycle management
- Explicit provider-runtime ownership for claim, heartbeat, result submission, and drain flows
- Session and API-key lifecycle flows
- Unary execution, streaming execution, and request lifecycle helpers
- Task lifecycle helpers
- Both gRPC and HTTP gateway access to the maintained control-plane flows

## Installation

### From PyPI (Recommended)

```bash
# Install the latest stable version
pip install toolplane-python-client

# Install a specific version
pip install toolplane-python-client==1.0.0
```

### From Source

```bash
# Clone the repository
git clone https://github.com/your-org/tool-control-plane.git
cd toolplane/clients/python

# Install in development mode
pip install -e .

# Install with all development dependencies
pip install -e ".[dev]"
```

### Requirements

The toolplane Python client requires Python 3.8 or higher and the following dependencies:

```bash
# Core dependencies
pip install grpcio grpcio-tools requests

# Optional dependencies for enhanced features
pip install pydantic typing-extensions
```

## Quick Start

### Consumer Usage with gRPC Client

```python
from toolplane import Toolplane

# Initialize a consumer-oriented client
client = Toolplane(
    server_host="localhost",
    server_port=50051,
    api_key="your-api-key-here"
)

client.connect()

# Attach to an existing provider-backed session
session_id = "provider-session-123"
tools = client.get_available_tools(session_id)
print(tools)

# Execute a tool synchronously
result = client.invoke("echo-tool", session_id=session_id, text="Hello World")
print(result)  # Output: Echo: Hello World

# Execute a tool asynchronously
request_id = client.ainvoke("echo-tool", session_id=session_id, text="Async Test")
print(f"Request ID: {request_id}")

# Stream tool execution
def stream_callback(chunk, is_final):
    if is_final:
        print("Stream completed")
    else:
        print(f"Received chunk: {chunk}")

client.stream("echo-tool", stream_callback, session_id=session_id, text="Streaming Test")
```

### Explicit Provider Runtime with gRPC Client

```python
from toolplane import Toolplane

client = Toolplane(
    server_host="localhost",
    server_port=50051,
    api_key="your-api-key-here",
    user_id="provider-user"
)

provider = client.provider_runtime()
session = provider.create_session(
    name="My Provider Session",
    description="Session backed by the explicit provider runtime",
    namespace="development"
)

@provider.tool(
    session_id=session.session_id,
    name="echo-tool",
    description="Echo back the input text",
    tags=["utility", "testing"]
)
def echo_tool(text: str) -> str:
    return f"Echo: {text}"

provider.run_forever()
```

### Explicit Provider Runtime with HTTP Client

```python
from toolplane import ToolplaneHTTP

# Initialize HTTP client
client = ToolplaneHTTP(
    server_host="localhost",
    server_port=8080,
    api_key="your-api-key-here"
)

# Create an explicit provider runtime
provider = client.provider_runtime()
session = provider.create_session(
    user_id="user123",
    name="My HTTP Provider Session"
)

# Register a tool
@provider.tool(
    session_id=session.session_id,
    name="http-tool",
    description="HTTP-based tool execution"
)
def http_tool(text: str) -> str:
    """Process text via HTTP."""
    return f"HTTP Processed: {text}"

provider.run_forever()
```

## Core Components

### Toolplane Client (gRPC)

The main `Toolplane` class provides a comprehensive interface for gRPC-based interactions with the Toolplane server:

```python
from toolplane import Toolplane

# Initialize with comprehensive configuration
client = Toolplane(
    # Server connection settings
    server_host="localhost",
    server_port=9001,
    use_tls=True,
    tls_ca_cert_path="../../server/deploy/reference/certs/ca.crt",
    tls_server_name="localhost",
    
    # Authentication
    api_key="your-api-key",
    user_id="user123",
    
    # Session configuration
    session_name="My Application Session",
    session_description="Session for my application",
    session_namespace="production",
    
    # Performance settings
    heartbeat_interval=60,
    max_workers=10,
    request_timeout=30,
    
    # Retry configuration
    max_retries=3,
    retry_base_delay=1.0,
    retry_max_delay=60.0,
    retry_backoff_factor=2.0
)
```

### ToolplaneHTTP Client (HTTP)

The `ToolplaneHTTP` class provides HTTP-based interaction with the Toolplane server:

```python
from toolplane import ToolplaneHTTP

# Initialize with HTTP-specific configuration
client = ToolplaneHTTP(
    server_host="localhost",
    server_port=8080,
    api_key="your-api-key",
    user_id="user123",
    session_name="My HTTP Session",
    
    # HTTP-specific settings
    max_buffer_size=4*1024*1024,  # 4MB default
    max_retries=3,
    request_timeout=30,
    retry_backoff_ms=250,
    heartbeat_interval=60
)
```

## Session Management

### Creating and Managing Sessions

```python
# Consumer-side session creation does not imply machine registration
session = client.create_session(
    session_id="my-unique-session-id",
    user_id="user123",
    name="My Application Session",
    description="Session for my application",
    namespace="production",
    register_machine=False
)

# Provider-side session creation uses the explicit runtime and attaches a machine
provider = client.provider_runtime()
provider_session = provider.create_session(
    user_id="user123",
    name="Provider Session",
    description="Machine-backed provider session",
    namespace="production"
)

# Create session without specifying ID (auto-generated)
session = client.create_session(
    user_id="user123",
    name="Auto-generated Session"
)

# Get existing session
session_context = client.get_session("my-session-123")

# List user sessions with pagination
user_sessions = client.list_user_sessions(
    user_id="user123",
    page_size=10,
    page_token=0,
    filter="active"  # Filter by status
)

# Get session statistics
stats = client.get_session_stats(user_id="user123")
print(f"Total sessions: {stats['total_sessions']}")
print(f"Active sessions: {stats['active_sessions']}")
print(f"Expired sessions: {stats['expired_sessions']}")
```

### Session Operations

```python
# Bulk delete sessions
bulk_delete_result = client.bulk_delete_sessions(
    user_id="user123",
    session_ids=["session-1", "session-2"],
    filter="inactive"  # Delete inactive sessions
)

# Refresh session token
refresh_result = client.refresh_session_token(session_id="session-123")
print(f"New token: {refresh_result['new_token']}")

# Invalidate session
success = client.invalidate_session(session_id="session-123", reason="Session expired")
print(f"Session invalidated: {success}")
```

## Tool Management

### Registering Tools with Schema Validation

```python
# Provider registration uses the explicit runtime
provider = client.provider_runtime(["session-123"])
provider.attach_session("session-123", register_machine=True)

# Register a tool with automatic schema generation
@provider.tool(
    session_id="session-123", 
    name="process-data",
    description="Process and transform data",
    stream=False,
    tags=["data-processing", "utility"]
)
def process_data(input: str, format: str = "json") -> dict:
    """
    Process input data with specified format.
    
    Args:
        input: Raw input data
        format: Output format (json, xml, csv)
        
    Returns:
        Processed data structure
    """
    processed = {"input": input, "format": format, "processed": True}
    return processed

# Register a streaming tool
@provider.tool(
    session_id="session-123",
    name="stream-data",
    description="Stream data processing",
    stream=True,
    tags=["streaming", "data-processing"]
)
def stream_data(input: str):
    """Stream data processing."""
    for char in input:
        yield f"Processing: {char}"

# Register tool with custom schema
@provider.tool(
    session_id="session-123",
    name="custom-tool",
    description="Custom tool with explicit schema"
)
def custom_tool(data: str, number: int = 1) -> dict:
    """
    A custom tool with explicit schema.
    
    Args:
        data: Input data
        number: Number parameter
        
    Returns:
        Result dictionary
    """
    return {"result": f"{data}-{number}"}

# Get available tools
available_tools = client.get_available_tools("session-123")
print(f"Available tools: {available_tools}")
```

### Tool Execution Patterns

```python
# Synchronous execution
result = client.invoke("process-data", session_id="session-123", input="test data", format="json")

# Asynchronous execution (returns request ID)
request_id = client.ainvoke("process-data", session_id="session-123", input="test data")

# Streaming execution
def stream_callback(chunk, is_final):
    if is_final:
        print("Stream completed")
    else:
        print(f"Received chunk: {chunk}")

client.stream("stream-data", stream_callback, session_id="session-123", input="stream data")

# Get request status
status = client.get_request_status("request-123", session_id="session-123")
print(f"Request status: {status['status']}")
```

## Machine Management

### Automatic Machine Registration

```python
# Consumer sessions are not machine-backed by default
provider = client.provider_runtime(["session-123"])
session_context = provider.attach_session("session-123", register_machine=True)
machine_id = session_context.machine_id
print(f"Registered machine ID: {machine_id}")

# Public machine-management helpers stay on the client
machine_status = client.get_machine("session-123", machine_id)
print(f"Machine status: {machine_status}")

# Drain machine gracefully before shutdown
client.drain_machine("session-123", machine_id)

# Unregister machine
client.unregister_machine("session-123", machine_id)
```

### Heartbeat Management

```python
# Heartbeat is owned by the explicit provider runtime
client = Toolplane(
    server_host="localhost",
    heartbeat_interval=30  # Send heartbeat every 30 seconds
)

provider = client.provider_runtime(["session-123"])
provider.start_in_background()

# Stop heartbeat when needed
provider.stop()
```

## Request Processing

### Request Management

```python
# Create and process requests
request_id = client.create_request(
    session_id="session-123",
    tool_name="process-data",
    input_data='{"text": "test"}'
)

# Get request status
status = client.get_request_status(request_id, "session-123")
print(f"Request status: {status['status']}")

# Poll for completed requests
def poll_requests(client, session_id, max_attempts=10):
    """Poll for request completion."""
    for attempt in range(max_attempts):
        try:
            # Simulate polling
            time.sleep(1)
            status = client.get_request_status("request-123", session_id)
            if status['status'] in ['done', 'failure']:
                print(f"Request completed with status: {status['status']}")
                return status
        except Exception as e:
            print(f"Polling error: {e}")
            break
    return None
```

### Streaming Request Processing

```python
# Streaming with callbacks
def streaming_process_callback(chunk, is_final):
    """Process streaming chunks."""
    if is_final:
        print("Stream completed successfully")
    else:
        try:
            # Parse JSON chunk if needed
            data = json.loads(chunk) if isinstance(chunk, str) else chunk
            print(f"Received chunk: {data}")
        except json.JSONDecodeError:
            print(f"Received raw chunk: {chunk}")

# Execute streaming request
stream_result = client.stream(
    tool_name="stream-data",
    callback=streaming_process_callback,
    session_id="session-123",
    input="streaming test data"
)
```

## Configuration

### Client Configuration Parameters

```python
# gRPC Client Configuration
grpc_client = Toolplane(
    # Server Connection Settings
    server_host="localhost",           # Server hostname
    server_port=9001,                  # Server port
    use_tls=False,                     # Enable TLS encryption
    tls_ca_cert_path=None,             # Optional CA bundle for direct gRPC TLS
    tls_server_name=None,              # Optional server name override (use localhost for the reference stack)
    
    # Authentication Settings
    api_key="your-api-key",            # Authentication key
    user_id="user123",                 # User identifier
    
    # Session Configuration
    session_name="My Session",         # Session name
    session_description="Session desc", # Session description
    session_namespace="production",    # Namespace for session
    
    # Performance Settings
    heartbeat_interval=60,             # Heartbeat interval in seconds
    max_workers=10,                    # Maximum worker threads
    request_timeout=30,                # Request timeout in seconds
    
    # Retry Configuration
    max_retries=3,                     # Maximum retry attempts
    retry_base_delay=1.0,              # Base retry delay in seconds
    retry_max_delay=60.0,              # Maximum retry delay in seconds
    retry_backoff_factor=2.0,          # Exponential backoff factor
    
    # Advanced Settings
    debug=False,                       # Enable debug logging
    log_level="INFO"                   # Logging level
)

# HTTP Client Configuration
http_client = ToolplaneHTTP(
    # Server Connection Settings
    server_host="localhost",           # Server hostname
    server_port=8080,                  # Server port
    use_tls=False,                     # Enable TLS encryption
    
    # Authentication Settings
    api_key="your-api-key",            # Authentication key
    user_id="user123",                 # User identifier
    
    # Session Configuration
    session_name="My HTTP Session",    # Session name
    session_description="Session desc", # Session description
    session_namespace="production",    # Namespace for session
    
    # HTTP-Specific Settings
    max_buffer_size=4*1024*1024,       # Buffer size (4MB default)
    max_retries=3,                     # Maximum retry attempts
    request_timeout=30,                # Request timeout in seconds
    retry_backoff_ms=250,              # Retry backoff in milliseconds
    
    # Performance Settings
    heartbeat_interval=60,             # Heartbeat interval in seconds
    max_workers=10,                    # Maximum worker threads
    
    # Advanced Settings
    debug=False,                       # Enable debug logging
    log_level="INFO"                   # Logging level
)
```

## Error Handling

### Comprehensive Error Handling

```python
from toolplane.core.errors import (
    ToolplaneError,
    ConnectionError,
    ToolError,
    SessionError,
    MachineError,
    RequestError
)

# Global error handling strategy
def handle_toolplane_errors(error, operation):
    """Handle Toolplane client errors comprehensively."""
    print(f"Error occurred during {operation}: {error}")
    
    if isinstance(error, ConnectionError):
        print("Connection error - check server availability")
        return "connection-error"
    elif isinstance(error, ToolError):
        print("Tool error - check tool parameters and implementation")
        return "tool-error"
    elif isinstance(error, SessionError):
        print("Session error - check session validity and permissions")
        return "session-error"
    elif isinstance(error, MachineError):
        print("Machine error - check machine registration and status")
        return "machine-error"
    elif isinstance(error, RequestError):
        print("Request error - check request payload and server status")
        return "request-error"
    elif isinstance(error, ToolplaneError):
        print("General Toolplane error - check logs for details")
        return "general-error"
    else:
        print("Unknown error type")
        return "unknown-error"

# Error handling in operations
def safe_operation(client, operation_name, operation_func, *args, **kwargs):
    """Execute operation with comprehensive error handling."""
    try:
        result = operation_func(*args, **kwargs)
        print(f"Operation {operation_name} completed successfully")
        return result
    except Exception as e:
        error_type = handle_toolplane_errors(e, operation_name)
        # Log error details
        print(f"Error type: {error_type}")
        print(f"Error details: {str(e)}")
        # Re-raise for upper-level handling if needed
        raise

# Usage examples
try:
    # Safe tool invocation
    result = safe_operation(
        client, 
        "tool invocation", 
        client.invoke, 
        "echo-tool", 
        "session-123", 
        text="test"
    )
    
    # Safe session creation
    session = safe_operation(
        client, 
        "session creation", 
        client.create_session, 
        "session-123", 
        "user123", 
        "Test Session"
    )
    
except Exception as e:
    print(f"Critical error in operation: {e}")
```

## Advanced Features

### Async Operations

```python
import asyncio
import concurrent.futures

# Async client usage
async def async_tool_operations():
    """Demonstrate async tool operations."""
    # Create client
    client = ToolplaneHTTP()
    provider = client.provider_runtime(["session-123"])
    provider.attach_session("session-123", register_machine=True)
    
    # Register async tool
    @provider.tool("session-123", "async-tool")
    def async_tool(data: str) -> str:
        return f"Async processed: {data}"
    
    # Execute async operation
    request_id = client.ainvoke("async-tool", "session-123", data="async data")
    print(f"Async request submitted: {request_id}")
    
    # Wait for completion
    result = await asyncio.get_event_loop().run_in_executor(
        None, 
        lambda: client.get_request_status(request_id, "session-123")
    )
    print(f"Async result: {result}")
    
    return result

# Run async operations
# asyncio.run(async_tool_operations())
```

### Batch Operations

```python
def batch_tool_registration(client, session_id, tools_config):
    """Register multiple tools in batch."""
    provider = client.provider_runtime([session_id])
    provider.attach_session(session_id, register_machine=True)
    results = []
    
    for tool_config in tools_config:
        try:
            # Register tool
            tool_func = tool_config['function']
            tool_name = tool_config['name']
            
            @provider.tool(session_id, tool_name, **tool_config.get('options', {}))
            def wrapped_tool(*args, **kwargs):
                return tool_func(*args, **kwargs)
            
            results.append({
                'name': tool_name,
                'status': 'success',
                'error': None
            })
        except Exception as e:
            results.append({
                'name': tool_name,
                'status': 'failed',
                'error': str(e)
            })
    
    return results

# Usage
tools_config = [
    {
        'name': 'tool-1',
        'function': lambda x: f"Processed {x}",
        'options': {'description': 'First tool'}
    },
    {
        'name': 'tool-2', 
        'function': lambda x: f"Transformed {x}",
        'options': {'description': 'Second tool'}
    }
]

# Register batch of tools
batch_results = batch_tool_registration(client, "session-123", tools_config)
print(f"Batch registration results: {batch_results}")
```

## Security and Authentication

### API Key Management

```python
import os
from cryptography.fernet import Fernet

class SecureAPIClient:
    """Secure API client with key management."""
    
    def __init__(self, config):
        self.config = config
        self._key = None
        self._cipher_suite = None
        
    def setup_encryption(self, encryption_key=None):
        """Setup encryption for sensitive data."""
        if encryption_key:
            self._key = encryption_key
        else:
            self._key = Fernet.generate_key()
            
        self._cipher_suite = Fernet(self._key)
        
    def encrypt_api_key(self, api_key):
        """Encrypt API key."""
        if not self._cipher_suite:
            raise ValueError("Encryption not setup")
        return self._cipher_suite.encrypt(api_key.encode())
        
    def decrypt_api_key(self, encrypted_key):
        """Decrypt API key."""
        if not self._cipher_suite:
            raise ValueError("Encryption not setup")
        return self._cipher_suite.decrypt(encrypted_key).decode()

# Secure API key handling
secure_client = SecureAPIClient({
    'server_host': 'localhost',
    'server_port': 50051
})

# Handle API keys securely
api_key = "your-sensitive-api-key"
encrypted_key = secure_client.encrypt_api_key(api_key)
print(f"Encrypted key: {encrypted_key}")

# Use in client
client = Toolplane(
    server_host="localhost",
    api_key=api_key,  # Use decrypted key
    user_id="user123"
)
```

### Authentication with Environment Variables

```python
import os
from typing import Optional

def get_secure_config() -> dict:
    """Get secure configuration from environment variables."""
    config = {
        'server_host': os.getenv('TOOLPLANE_SERVER_HOST', 'localhost'),
        'server_port': int(os.getenv('TOOLPLANE_SERVER_PORT', '50051')),
        'api_key': os.getenv('TOOLPLANE_API_KEY'),
        'user_id': os.getenv('TOOLPLANE_USER_ID', ''),
        'session_name': os.getenv('TOOLPLANE_SESSION_NAME', 'Default Session')
    }
    
    # Validate required settings
    if not config['api_key']:
        raise ValueError("TOOLPLANE_API_KEY environment variable is required")
    
    return config

# Usage with environment variables
try:
    secure_config = get_secure_config()
    client = Toolplane(**secure_config)
    print("Client configured with secure settings")
except ValueError as e:
    print(f"Configuration error: {e}")
```

## Monitoring and Debugging

### Client Statistics and Metrics

```python
import time
from collections import defaultdict

class ClientMonitor:
    """Client monitoring and statistics collector."""
    
    def __init__(self):
        self.metrics = defaultdict(int)
        self.start_time = time.time()
        
    def record_operation(self, operation_name, duration=None):
        """Record operation metrics."""
        self.metrics[f"operations_{operation_name}"] += 1
        if duration:
            self.metrics[f"duration_{operation_name}"] += duration
            
    def get_metrics(self):
        """Get current metrics."""
        return dict(self.metrics)
        
    def reset_metrics(self):
        """Reset all metrics."""
        self.metrics.clear()
        self.start_time = time.time()

# Monitor client operations
monitor = ClientMonitor()

# Wrap client operations with monitoring
def monitored_invoke(client, tool_name, session_id, **params):
    """Monitored tool invocation."""
    start_time = time.time()
    try:
        result = client.invoke(tool_name, session_id, **params)
        duration = time.time() - start_time
        monitor.record_operation(f"invoke_{tool_name}", duration)
        return result
    except Exception as e:
        monitor.record_operation(f"invoke_{tool_name}_error")
        raise

# Usage
result = monitored_invoke(client, "echo-tool", "session-123", text="test")
print(f"Result: {result}")
print(f"Metrics: {monitor.get_metrics()}")
```

## Production Usage

### Production Configuration

```python
import os
import sys
from typing import Dict, Any

class ProductionClientConfig:
    """Production-ready client configuration."""
    
    @staticmethod
    def get_production_config() -> Dict[str, Any]:
        """Get production configuration."""
        return {
            'server_host': os.getenv('TOOLPLANE_SERVER_HOST', 'localhost'),
            'server_port': int(os.getenv('TOOLPLANE_SERVER_PORT', '50051')),
            'api_key': os.getenv('TOOLPLANE_API_KEY', ''),
            'user_id': os.getenv('TOOLPLANE_USER_ID', 'production-user'),
            'session_name': os.getenv('TOOLPLANE_SESSION_NAME', 'Production Session'),
            'session_namespace': os.getenv('TOOLPLANE_NAMESPACE', 'production'),
            
            # Performance tuning
            'heartbeat_interval': int(os.getenv('TOOLPLANE_HEARTBEAT_INTERVAL', '30')),
            'max_workers': int(os.getenv('TOOLPLANE_MAX_WORKERS', '20')),
            'request_timeout': int(os.getenv('TOOLPLANE_REQUEST_TIMEOUT', '60')),
            
            # Retry configuration
            'max_retries': int(os.getenv('TOOLPLANE_MAX_RETRIES', '3')),
            'retry_base_delay': float(os.getenv('TOOLPLANE_RETRY_BASE_DELAY', '1.0')),
            'retry_max_delay': float(os.getenv('TOOLPLANE_RETRY_MAX_DELAY', '60.0')),
            
            # Security settings
            'use_tls': os.getenv('TOOLPLANE_USE_TLS', 'false').lower() == 'true',
            
            # Debug settings
            'debug': os.getenv('TOOLPLANE_DEBUG', 'false').lower() == 'true'
        }

# Production usage example
def create_production_client():
    """Create a production-ready client."""
    config = ProductionClientConfig.get_production_config()
    
    # Validate configuration
    if not config['api_key']:
        raise ValueError("API key is required for production")
    
    # Create client
    client = Toolplane(**config)
    
    # Validate connection
    try:
        health = client.health()
        print(f"Server health check: {health}")
    except Exception as e:
        print(f"Health check failed: {e}")
        raise
    
    return client

# Usage in production environment
try:
    prod_client = create_production_client()
    print("Production client created successfully")
except Exception as e:
    print(f"Failed to create production client: {e}")
    sys.exit(1)
```

## Docker Integration

### Dockerfile Configuration

```dockerfile
# Dockerfile for toolplane client
FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Copy requirements
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY . .

# Create non-root user
RUN useradd --create-home --shell /bin/bash appuser && \
    chown -R appuser:appuser /app
USER appuser

# Expose port if needed
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import toolplane; print('Client ready')" || exit 1

# Default command
CMD ["python", "app.py"]
```

### Docker Compose Configuration

```yaml
# docker-compose.yml
version: '3.8'

services:
  toolplane-client:
    build: .
    environment:
      - TOOLPLANE_SERVER_HOST=toolplane-server
      - TOOLPLANE_SERVER_PORT=50051
      - TOOLPLANE_API_KEY=${TOOLPLANE_API_KEY}
      - TOOLPLANE_USER_ID=${TOOLPLANE_USER_ID:-client-user}
    depends_on:
      - toolplane-server
    restart: unless-stopped
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
    healthcheck:
      test: ["CMD", "python", "-c", "import toolplane; print('Client healthy')"]
      interval: 30s
      timeout: 10s
      retries: 3

  toolplane-server:
    image: toolplane-server:latest
    ports:
      - "50051:50051"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

## Troubleshooting

### Common Issues and Solutions

```python
import logging
import traceback
from typing import Dict, Any

class ClientTroubleshooter:
    """Troubleshooting utilities for Toolplane client."""
    
    def __init__(self):
        self.logger = logging.getLogger("toolplane.troubleshooter")
        
    def diagnose_connection(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """Diagnose connection issues."""
        diagnosis = {
            'connection_status': 'unknown',
            'errors': [],
            'recommendations': []
        }
        
        try:
            # Test basic connectivity
            client = Toolplane(**config)
            health = client.health()
            diagnosis['connection_status'] = 'successful'
            diagnosis['health'] = health
            
        except Exception as e:
            diagnosis['connection_status'] = 'failed'
            diagnosis['errors'].append(str(e))
            diagnosis['recommendations'].append("Check server availability and network connectivity")
            
            # More specific diagnostics
            if "Connection refused" in str(e):
                diagnosis['recommendations'].append("Verify server is running on port " + str(config.get('server_port', 50051)))
            elif "Permission denied" in str(e):
                diagnosis['recommendations'].append("Check API key permissions")
            elif "timeout" in str(e).lower():
                diagnosis['recommendations'].append("Increase timeout values or check network latency")
                
        return diagnosis
        
    def log_detailed_error(self, error: Exception, context: str = ""):
        """Log detailed error information."""
        self.logger.error(f"Error in {context}: {error}")
        self.logger.error(f"Error type: {type(error).__name__}")
        self.logger.error(f"Traceback:\n{traceback.format_exc()}")

# Usage
troubleshooter = ClientTroubleshooter()

# Diagnose connection
config = {
    'server_host': 'localhost',
    'server_port': 50051,
    'api_key': 'test-key'
}

diagnosis = troubleshooter.diagnose_connection(config)
print(f"Diagnosis: {diagnosis}")

# Log detailed error
try:
    client = Toolplane(**config)
    client.invoke("nonexistent-tool", "session-123")
except Exception as e:
    troubleshooter.log_detailed_error(e, "tool invocation")
```

### Error Code Reference

| Error Code | Description | Resolution |
| ------------ | ------------- | ------------ |
| **ConnectionError** | Network connection failure | Check server availability, network connectivity, and firewall rules |
| **ToolError** | Tool execution failure | Validate tool parameters, check tool implementation, and verify tool existence |
| **SessionError** | Session management issues | Verify session ID validity, check user permissions, and validate session state |
| **MachineError** | Machine registration issues | Ensure machine credentials are correct, check network connectivity, and verify server status |
| **RequestError** | Request processing issues | Check request payload format, validate server status, and review error logs |
| **ToolplaneError** | General Toolplane errors | Review error logs, check configuration settings, and consult documentation |
| **AuthenticationError** | Authentication failure | Verify API key validity, check permissions, and validate credential format |
| **TimeoutError** | Operation timeout | Increase timeout values, check network latency, and optimize server performance |

## Contributing

We welcome contributions to the toolplane Python client! Follow these guidelines to contribute effectively:

### Development Setup

1. **Clone the repository:**

```bash
git clone https://github.com/your-org/tool-control-plane.git
cd toolplane/clients/python
```

1. **Install development dependencies:**

```bash
pip install -e ".[dev]"
```

### Code Style Guidelines

1. **PEP 8 Compliance**: Follow Python's official style guide
2. **Docstring Format**: Use Google-style docstrings
3. **Type Hints**: Provide comprehensive type hints
4. **Error Handling**: Implement proper error handling with descriptive messages
5. **Logging**: Use structured logging for debugging and monitoring

### Testing Procedures

1. **Unit Tests**: All new features must include unit tests
2. **Integration Tests**: Test with actual server connections when possible
3. **Test Coverage**: Maintain at least 85% test coverage
4. **Continuous Integration**: All tests must pass before merging

### Pull Request Template

```markdown
## Description

Brief description of the changes introduced by this PR.

## Related Issues

- Fixes #issue-number
- Related to #feature-request

## Changes Made

- [ ] Feature addition
- [ ] Bug fix
- [ ] Documentation update
- [ ] Test enhancement

## Testing

- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing performed
- [ ] Edge cases covered

## Checklist

- [ ] Code follows PEP 8 style guide
- [ ] All tests pass
- [ ] Documentation updated
- [ ] Type hints provided
- [ ] Error handling implemented
```

## Performance Tuning

### Configuration Optimization

```python
from typing import Dict, Any

class PerformanceOptimizer:
    """Performance optimization utilities."""
    
    @staticmethod
    def optimize_config_for_load(config: Dict[str, Any], load_profile: str = "normal") -> Dict[str, Any]:
        """Optimize configuration based on load profile."""
        optimized = config.copy()
        
        # Load profiles
        if load_profile == "low":
            optimized.update({
                'heartbeat_interval': 60,
                'max_workers': 5,
                'request_timeout': 30
            })
        elif load_profile == "medium":
            optimized.update({
                'heartbeat_interval': 30,
                'max_workers': 10,
                'request_timeout': 60
            })
        elif load_profile == "high":
            optimized.update({
                'heartbeat_interval': 15,
                'max_workers': 20,
                'request_timeout': 120
            })
        else:
            # Default to normal
            optimized.update({
                'heartbeat_interval': 30,
                'max_workers': 10,
                'request_timeout': 60
            })
            
        return optimized
    
    @staticmethod
    def get_optimal_worker_count() -> int:
        """Get optimal worker count based on system resources."""
        import multiprocessing
        cpu_count = multiprocessing.cpu_count()
        # Use 2x CPU count but cap at reasonable levels
        optimal = min(cpu_count * 2, 50)
        return max(optimal, 5)  # Minimum 5 workers

# Usage
config = {
    'server_host': 'localhost',
    'server_port': 50051,
    'api_key': 'your-api-key'
}

# Optimize for high load
optimized_config = PerformanceOptimizer.optimize_config_for_load(config, "high")
print(f"Optimized config: {optimized_config}")

# Get optimal worker count
worker_count = PerformanceOptimizer.get_optimal_worker_count()
print(f"Optimal worker count: {worker_count}")
```

## API Reference

### Core Classes and Methods

#### Toolplane Client Class

**Methods:**

- `__init__(**kwargs)` - Initialize client with configuration
- `connect()` - Establish connection to server
- `disconnect()` - Close connection to server
- `create_session(**kwargs)` - Create new session
- `get_session(session_id)` - Get session context
- `list_sessions()` - List all session contexts
- `provider_runtime(session_ids=None)` - Create or extend the explicit provider runtime for this client
- `tool(session_id, name, description, stream, tags)` - Backward-compatible provider decorator alias over `ProviderRuntime.tool(...)`
- `invoke(tool_name, session_id, **params)` - Synchronous tool execution
- `ainvoke(tool_name, session_id, **params)` - Asynchronous tool execution
- `stream(tool_name, callback, session_id, **params)` - Streaming tool execution
- `get_available_tools(session_id)` - Get available tools for session
- `get_request_status(request_id, session_id)` - Get request status
- `start()` / `stop()` - Backward-compatible aliases over the explicit provider runtime

**Properties:**

- `running` - Client running status
- `session_ids` - List of managed session IDs
- `config` - Client configuration object

#### ToolplaneHTTP Client Class

**Methods:**

- `__init__(**kwargs)` - Initialize HTTP client with configuration
- `connect()` - Establish HTTP connection to server
- `disconnect()` - Close HTTP connection to server
- `create_session(**kwargs)` - Create new session via HTTP
- `get_session(session_id)` - Get session context via HTTP
- `list_sessions()` - List all session contexts via HTTP
- `provider_runtime(session_ids=None)` - Create or extend the explicit provider runtime for this client
- `tool(session_id, name, description, stream, tags)` - Backward-compatible provider decorator alias over `ProviderRuntime.tool(...)`
- `invoke(tool_name, session_id, **params)` - Synchronous tool execution via HTTP
- `ainvoke(tool_name, session_id, **params)` - Asynchronous tool execution via HTTP
- `stream(tool_name, callback, session_id, **params)` - Streaming tool execution via HTTP
- `get_available_tools(session_id)` - Get available tools for session via HTTP
- `get_request_status(request_id, session_id)` - Get request status via HTTP
- `health()` - Check HTTP server health

#### ProviderRuntime Class

**Methods:**

- `create_session(**kwargs)` - Create a provider-owned session and attach a machine when needed
- `attach_session(session_id, register_machine=True)` - Attach the provider runtime to an existing session
- `tool(session_id, name, description, stream, tags)` - Register a machine-backed tool on the provider runtime
- `start_in_background(session_ids=None)` - Start heartbeats and polling without blocking the current thread
- `run_forever(session_ids=None)` - Start the provider runtime and block until stopped
- `stop()` - Stop heartbeats and polling for the explicit provider runtime

#### Session Management Methods (Admin Scope — Python-Only)

**Session Operations:**

- `list_user_sessions(user_id, page_size, page_token, filter)` - List user sessions with pagination
- `bulk_delete_sessions(user_id, session_ids, filter)` - Bulk delete user sessions
- `get_session_stats(user_id)` - Get session statistics for user
- `refresh_session_token(session_id)` - Refresh session token
- `invalidate_session(session_id, reason)` - Invalidate session

#### Tool Management Methods

**Tool Operations:**

- `get_available_tools(session_id)` - Get available tools for session
- `get_tool_stats(session_id)` - Get tool statistics for session
- `validate_tool_params(session_id, tool_name, params)` - Validate tool parameters

### Configuration Parameters

#### Toolplane Configuration

| Parameter | Type | Default | Description |
| ----------- | ------ | --------- | ------------- |
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

#### ToolplaneHTTP Configuration

| Parameter | Type | Default | Description |
| ----------- | ------ | --------- | ------------- |
| `server_host` | str | "localhost" | Server hostname |
| `server_port` | int | 8080 | Server port |
| `use_tls` | bool | False | Enable TLS encryption |
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
| `debug` | bool | False | Enable debug logging |
| `log_level` | str | "INFO" | Logging level |

## Support

For issues and feature requests, please open an issue on the GitHub repository. We aim to respond within 24 hours.

*Built for modern distributed computing needs - scalable, secure, and production-ready.*

**Version:** 1.0.0  
**Last Updated:** 2024  
**Compatibility:** Python 3.8+  
**Dependencies:** grpcio, requests, cryptography  

---

*This documentation provides comprehensive guidance for using the toolplane Python client library. For the most current information, please refer to the official repository and documentation.*
