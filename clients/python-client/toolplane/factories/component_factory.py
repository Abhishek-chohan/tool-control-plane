"""Factory for creating Toolplane components with dependency injection."""

import inspect
from abc import ABC, abstractmethod
from typing import Any, Dict, Optional, Type, TypeVar, get_type_hints

from ..interfaces.connection_interface import IConnectionStrategy
from ..interfaces.event_interface import IEventEmitter
from ..interfaces.request_interface import IRequestQueue
from ..interfaces.tool_interface import IToolValidator

T = TypeVar("T")


class IDependencyContainer(ABC):
    """Abstract interface for dependency injection container."""

    @abstractmethod
    def register(
        self, interface: Type[T], implementation: Type[T], singleton: bool = True
    ) -> None:
        """Register an implementation for an interface."""
        pass

    @abstractmethod
    def register_instance(self, interface: Type[T], instance: T) -> None:
        """Register a specific instance for an interface."""
        pass

    @abstractmethod
    def resolve(self, interface: Type[T]) -> T:
        """Resolve an implementation for an interface."""
        pass

    @abstractmethod
    def can_resolve(self, interface: Type[T]) -> bool:
        """Check if interface can be resolved."""
        pass


class DependencyContainer(IDependencyContainer):
    """Simple dependency injection container."""

    def __init__(self):
        self._registrations: Dict[Type, Type] = {}
        self._instances: Dict[Type, Any] = {}
        self._singletons: Dict[Type, bool] = {}

    def register(
        self, interface: Type[T], implementation: Type[T], singleton: bool = True
    ) -> None:
        """Register an implementation for an interface."""
        self._registrations[interface] = implementation
        self._singletons[interface] = singleton

    def register_instance(self, interface: Type[T], instance: T) -> None:
        """Register a specific instance for an interface."""
        self._instances[interface] = instance
        self._singletons[interface] = True

    def resolve(self, interface: Type[T]) -> T:
        """Resolve an implementation for an interface."""
        # Check if we have a pre-created instance
        if interface in self._instances:
            return self._instances[interface]

        # Check if we have a registration
        if interface not in self._registrations:
            raise ValueError(f"No registration found for {interface}")

        implementation = self._registrations[interface]
        instance = self._create_instance(implementation)

        # Store instance if singleton
        if self._singletons.get(interface, True):
            self._instances[interface] = instance

        return instance

    def can_resolve(self, interface: Type[T]) -> bool:
        """Check if interface can be resolved."""
        return interface in self._registrations or interface in self._instances

    def _create_instance(self, cls: Type[T]) -> T:
        """Create instance with dependency injection."""
        # Get constructor signature
        sig = inspect.signature(cls.__init__)

        # Get type hints for parameters
        type_hints = get_type_hints(cls.__init__)

        # Resolve dependencies
        kwargs = {}
        for param_name, param in sig.parameters.items():
            if param_name == "self":
                continue

            # Try to resolve parameter type
            if param_name in type_hints:
                param_type = type_hints[param_name]
                if self.can_resolve(param_type):
                    kwargs[param_name] = self.resolve(param_type)
                elif param.default != param.empty:
                    # Use default value if available
                    kwargs[param_name] = param.default
                else:
                    raise ValueError(
                        f"Cannot resolve dependency {param_type} for {cls}"
                    )

        return cls(**kwargs)


class ComponentFactory:
    """Factory for creating Toolplane components with dependency injection."""

    def __init__(self, container: Optional[IDependencyContainer] = None):
        self.container = container or DependencyContainer()
        self._register_default_components()

    def _register_default_components(self) -> None:
        """Register default component implementations."""
        from ..interfaces.connection_interface import (
            DirectConnectionStrategy,
        )
        from ..interfaces.event_interface import EventBus
        from ..interfaces.request_interface import PriorityRequestQueue
        from ..interfaces.tool_interface import DefaultToolValidator

        # Register default implementations
        self.container.register(IConnectionStrategy, DirectConnectionStrategy)
        self.container.register(IRequestQueue, PriorityRequestQueue)
        self.container.register(IToolValidator, DefaultToolValidator)
        self.container.register(IEventEmitter, EventBus)

    def create_connection_manager(
        self,
        protocol: str,
        config: Dict[str, Any],
        strategy: Optional[IConnectionStrategy] = None,
    ):
        """Create a connection manager."""
        if strategy is None:
            strategy = self.container.resolve(IConnectionStrategy)

        if protocol == "http":
            from ..http_core.http_config import HTTPClientConfig
            from ..http_core.http_connection import HTTPConnectionManager

            client_config = HTTPClientConfig(**config)
            return HTTPConnectionManager(client_config, strategy)
        elif protocol == "grpc":
            from ..core.config import ClientConfig
            from ..core.connection import ConnectionManager

            client_config = ClientConfig(**config)
            return ConnectionManager(client_config, strategy)
        else:
            raise ValueError(f"Unsupported protocol: {protocol}")

    def create_tool_manager(self, connection_manager, validator: Optional = None):
        """Create a tool manager."""
        if validator is None:
            validator = self.container.resolve(IToolValidator)

        # Determine protocol from connection manager type
        if (
            hasattr(connection_manager, "__class__")
            and "HTTP" in connection_manager.__class__.__name__
        ):
            from ..http_core.http_tool import HTTPToolManager

            return HTTPToolManager(connection_manager, validator)
        else:
            from ..core.tool import ToolManager

            return ToolManager(connection_manager, validator)

    def create_request_manager(
        self, connection_manager, max_workers: int = 10, queue: Optional = None
    ):
        """Create a request manager."""
        if queue is None:
            queue = self.container.resolve(IRequestQueue)

        # Determine protocol from connection manager type
        if (
            hasattr(connection_manager, "__class__")
            and "HTTP" in connection_manager.__class__.__name__
        ):
            from ..http_core.http_request import HTTPRequestManager

            return HTTPRequestManager(connection_manager, max_workers, queue)
        else:
            from ..core.request import RequestManager

            return RequestManager(connection_manager, max_workers, queue)

    def create_session_manager(
        self, connection_manager, lifecycle_handler: Optional = None
    ):
        """Create a session manager."""
        # Determine protocol from connection manager type
        if (
            hasattr(connection_manager, "__class__")
            and "HTTP" in connection_manager.__class__.__name__
        ):
            from ..http_core.http_session import HTTPSessionManager

            return HTTPSessionManager(connection_manager, lifecycle_handler)
        else:
            from ..core.session import SessionManager

            return SessionManager(connection_manager, lifecycle_handler)

    def create_machine_manager(
        self, connection_manager, event_emitter: Optional = None
    ):
        """Create a machine manager."""
        if event_emitter is None:
            event_emitter = self.container.resolve(IEventEmitter)

        # Determine protocol from connection manager type
        if (
            hasattr(connection_manager, "__class__")
            and "HTTP" in connection_manager.__class__.__name__
        ):
            from ..http_core.http_machine import HTTPMachineManager

            return HTTPMachineManager(connection_manager, event_emitter)
        else:
            from ..core.machine import MachineManager

            return MachineManager(connection_manager, event_emitter)

    def register_component(
        self, interface: Type[T], implementation: Type[T], singleton: bool = True
    ) -> None:
        """Register a component implementation."""
        self.container.register(interface, implementation, singleton)

    def register_instance(self, interface: Type[T], instance: T) -> None:
        """Register a component instance."""
        self.container.register_instance(interface, instance)

    def resolve(self, interface: Type[T]) -> T:
        """Resolve a component."""
        return self.container.resolve(interface)


# Global component factory instance
_global_component_factory: Optional[ComponentFactory] = None


def get_component_factory() -> ComponentFactory:
    """Get global component factory instance."""
    global _global_component_factory
    if _global_component_factory is None:
        _global_component_factory = ComponentFactory()
    return _global_component_factory


def register_component(
    interface: Type[T], implementation: Type[T], singleton: bool = True
) -> None:
    """Register a component with the global factory."""
    factory = get_component_factory()
    factory.register_component(interface, implementation, singleton)


def resolve_component(interface: Type[T]) -> T:
    """Resolve a component from the global factory."""
    factory = get_component_factory()
    return factory.resolve(interface)
