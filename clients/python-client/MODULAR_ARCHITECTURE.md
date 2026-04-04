# Toolplane Modular Architecture Documentation

> **Historical design reference.** This document describes the modular architecture patterns (factories, DI, event bus) explored during Python client development. For the current authoritative architecture overview, see `clients/python-client/ARCHITECTURE.md`. This file is preserved as design-level reference only.

## Overview

The Toolplane client has been enhanced with a sophisticated modular architecture that follows industry best practices for building maintainable, scalable, and extensible software systems. This architecture emphasizes loose coupling, high cohesion, and separation of concerns.

## Architecture Principles

### 1. Interface-Based Design
All major components implement well-defined interfaces, enabling:
- Protocol-agnostic implementations
- Easy component swapping and testing
- Clear contracts between modules
- Support for multiple implementation strategies

### 2. Factory Pattern
Multiple factory implementations provide:
- Centralized object creation logic
- Dependency injection capabilities
- Configuration validation
- Protocol-specific client creation

### 3. Strategy Pattern
Pluggable strategies for:
- Connection management (Direct, Pooled, Load Balanced)
- Request execution (Sync, Async, Parallel, Streaming)
- Retry mechanisms (Fixed, Exponential, Linear backoff)
- Load balancing (Round Robin, Least Connections, etc.)

### 4. Event-Driven Architecture
Decoupled communication through:
- Central EventBus with weak references
- Comprehensive event logging
- Performance metrics collection
- Custom event handlers

### 5. Dependency Injection
Automatic component resolution with:
- Interface-based registration
- Singleton and transient lifecycles
- Constructor parameter injection
- Circular dependency detection

## Module Structure

```
clients/python-client/toolplane/
├── interfaces/              # Protocol interfaces
│   ├── client_interface.py     # IToolplaneClient, IClientFactory
│   ├── connection_interface.py # IConnectionManager, strategies
│   ├── session_interface.py    # ISessionManager, ISessionContext
│   ├── tool_interface.py       # IToolManager, validation
│   ├── request_interface.py    # IRequestManager, queuing
│   └── event_interface.py      # IEventEmitter, event system
├── factories/               # Component factories
│   ├── client_factory.py       # Protocol-specific client creation
│   ├── component_factory.py    # DI container and assembly
│   └── strategy_factory.py     # Strategy implementations
├── common/                  # Shared utilities
│   ├── base_config.py          # Abstract configuration
│   ├── constants.py            # System constants
│   └── utils.py                # Helper functions
├── core/                    # Core gRPC implementation
├── http_core/              # HTTP-specific implementation
└── examples/               # Usage examples
    └── modular_usage_example.py
```

## Key Components

### Interface Layer

#### IToolplaneClient
```python
@runtime_checkable
class IToolplaneClient(Protocol):
    def connect(self) -> bool: ...
    def disconnect(self) -> None: ...
    def create_session(...) -> ISessionContext: ...
    def get_session(session_id: str) -> Optional[ISessionContext]: ...
    def start(self) -> None: ...
    def stop(self) -> None: ...
```

#### IConnectionManager
```python
@runtime_checkable
class IConnectionManager(Protocol):
    @property
    def connected(self) -> bool: ...
    @property  
    def state(self) -> ConnectionState: ...
    def connect(self) -> bool: ...
    def disconnect(self) -> None: ...
    def health_check(self) -> bool: ...
```

### Factory Layer

#### ClientFactory
Creates protocol-specific clients with validation:
```python
# Create HTTP client
http_client = create_client(
    ClientProtocol.HTTP,
    server_host="localhost",
    server_port=8080
)

# Create gRPC client  
grpc_client = create_client(
    ClientProtocol.GRPC,
    server_host="localhost", 
    server_port=50051
)
```

#### ComponentFactory
Assembles components with dependency injection:
```python
factory = get_component_factory()

# Register custom implementations
factory.register_component(
    IConnectionStrategy,
    CustomConnectionStrategy
)

# Components automatically resolve dependencies
connection_manager = factory.create_connection_manager(
    protocol="http",
    config={"server_host": "localhost", "server_port": 8080}
)
```

### Strategy Layer

#### Connection Strategies
```python
# Direct connection
direct = ConnectionStrategyFactory.create_strategy(
    ConnectionStrategy.DIRECT
)

# Connection pooling
pooled = ConnectionStrategyFactory.create_strategy(
    ConnectionStrategy.POOLED,
    pool_size=5
)
```

#### Execution Strategies  
```python
# Synchronous execution
sync_strategy = StrategyFactory().create_execution_strategy(
    ExecutionStrategy.SYNCHRONOUS,
    timeout=30
)

# Parallel execution
parallel_strategy = StrategyFactory().create_execution_strategy(
    ExecutionStrategy.PARALLEL,
    max_workers=5
)
```

### Event System

#### Event Bus
```python
event_bus = get_global_event_bus()

# Subscribe to events
sub_id = event_bus.subscribe(
    EventType.TOOL_EXECUTED,
    my_event_handler,
    filter_func=lambda e: e.data.get('tool_name') == 'critical_tool'
)

# Emit events
event_bus.emit(Event(
    type=EventType.TOOL_EXECUTED,
    source="my_tool",
    data={"result": "success", "execution_time": 0.5}
))
```

#### Built-in Handlers
```python
# Event logging
logger = EventLogger(lambda msg: print(f"[LOG] {msg}"))

# Metrics collection
metrics = EventMetrics()
event_bus.subscribe(EventType.TOOL_EXECUTED, metrics)

# Get metrics
performance_data = metrics.get_metrics()
```

### Request Management

#### Priority Queuing
```python
# Create priority queue
queue = RequestQueueFactory.create_queue("priority")

# Create requests with different priorities
urgent_request = RequestInfo(
    request_id="urgent_1",
    session_id="session_1", 
    tool_name="critical_operation",
    parameters={},
    status=RequestStatus.PENDING,
    priority=RequestPriority.URGENT
)

# Higher priority requests processed first
queue.enqueue(urgent_request)
next_request = queue.dequeue()  # Returns urgent_request
```

### Tool Validation

#### Schema Validation
```python
# Tool with automatic schema generation
def add_numbers(a: int, b: int) -> int:
    """Add two numbers together."""
    return a + b

tool_def = ToolDefinition(
    name="add_numbers",
    func=add_numbers,
    description="Adds two numbers",
    tags=["math"]
)

# Validation happens automatically
validator = DefaultToolValidator()
errors = validator.validate_tool_definition(tool_def)

# Parameter validation
sanitized_params = validator.sanitize_parameters(
    tool_def, 
    {"a": "5", "b": "3"}  # Strings converted to integers
)
```

## Usage Examples

### Basic Client Creation
```python
from toolplane.factories import create_client
from toolplane.interfaces import ClientProtocol

# Create and use HTTP client
client = create_client(
    ClientProtocol.HTTP,
    server_host="localhost",
    server_port=8080,
    session_ids=["my_session"]
)

with client:
    session = client.get_session("my_session")
    
    # Register tool
    @client.tool("my_session", description="Add two numbers")
    def add(a: int, b: int) -> int:
        return a + b
    
    # Invoke tool
    result = session.invoke("add", a=5, b=3)
    print(f"Result: {result}")
```

### Advanced Configuration
```python
from toolplane.factories import get_component_factory
from toolplane.interfaces.connection_interface import ConnectionStrategy
from toolplane.interfaces.request_interface import RequestPriority

# Get component factory
factory = get_component_factory()

# Create connection manager with pooling
connection_manager = factory.create_connection_manager(
    protocol="http",
    config={
        "server_host": "localhost",
        "server_port": 8080,
        "strategy": ConnectionStrategy.POOLED,
        "pool_size": 10
    }
)

# Create request manager with priority queue
request_manager = factory.create_request_manager(
    connection_manager,
    max_workers=5,
    queue_type="priority"
)
```

### Event-Driven Monitoring
```python
from toolplane.interfaces.event_interface import get_global_event_bus, EventType

# Set up monitoring
event_bus = get_global_event_bus()

class PerformanceMonitor(IEventHandler):
    def handle_event(self, event: Event):
        if event.type == EventType.TOOL_EXECUTED:
            execution_time = event.data.get('execution_time', 0)
            if execution_time > 1.0:  # Slow operation
                print(f"Slow operation detected: {event.data}")

monitor = PerformanceMonitor()
event_bus.subscribe(EventType.TOOL_EXECUTED, monitor)
```

## Benefits

### Development Benefits
- **Testability**: Easy mocking and unit testing of components
- **Maintainability**: Clear separation of concerns and responsibilities  
- **Extensibility**: Add new protocols and strategies without changing existing code
- **Code Reuse**: Self-contained modules usable across different projects

### Runtime Benefits
- **Performance**: Connection pooling, request queuing, parallel execution
- **Reliability**: Retry mechanisms, health monitoring, circuit breakers
- **Observability**: Comprehensive logging, metrics, and event tracking
- **Scalability**: Load balancing, horizontal scaling support

### Operational Benefits
- **Configuration**: Pluggable strategies for different deployment scenarios
- **Monitoring**: Built-in metrics collection and performance tracking
- **Debugging**: Event logging and comprehensive error information
- **Deployment**: Support for multiple protocols and connection strategies

## Migration Path

Existing code using the basic Toolplane client can be gradually migrated:

1. **Phase 1**: Use factory functions instead of direct instantiation
2. **Phase 2**: Implement custom event handlers for monitoring
3. **Phase 3**: Add custom strategies for specific requirements  
4. **Phase 4**: Use dependency injection for component assembly

The modular architecture is backward compatible while providing a clear path to advanced features.

## Testing

The modular architecture enables comprehensive testing:

```python
# Mock interfaces for unit testing
class MockConnectionManager(IConnectionManager):
    def __init__(self):
        self.connected = True
        
    def connect(self) -> bool:
        return True
        
    # ... implement other methods

# Inject mock for testing
factory = ComponentFactory()
factory.register_instance(IConnectionManager, MockConnectionManager())

# Test component in isolation
tool_manager = factory.create_tool_manager(
    connection_manager=MockConnectionManager()
)
```

This architecture transforms the Toolplane client from a simple RPC library into a comprehensive, enterprise-ready framework for building distributed applications.