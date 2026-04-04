"""Factory for creating Toolplane clients with different protocols."""

from typing import Any, Dict, Optional, Type

from ..interfaces import ClientProtocol, IToolplaneClient
from ..interfaces.client_interface import BaseClientFactory


class ClientFactory(BaseClientFactory):
    """Concrete implementation of client factory."""

    def __init__(self):
        super().__init__()
        self._register_default_clients()

    def _register_default_clients(self) -> None:
        """Register default client implementations."""
        # Register HTTP client
        try:
            from ..toolplane_http_client import ToolplaneHTTP

            self.register_client(
                ClientProtocol.HTTP, ToolplaneHTTP, self._validate_http_config
            )
        except ImportError:
            pass

        # Register gRPC client if available
        try:
            from ..toolplane_client import Toolplane

            self.register_client(
                ClientProtocol.GRPC, Toolplane, self._validate_grpc_config
            )
        except ImportError:
            pass

    def _validate_http_config(self, config: Dict[str, Any]) -> bool:
        """Validate HTTP client configuration."""
        required_fields = ["server_host", "server_port"]
        for field in required_fields:
            if field not in config:
                return False

        # Validate types
        if not isinstance(config["server_host"], str):
            return False
        if not isinstance(config["server_port"], int):
            return False
        if config["server_port"] <= 0 or config["server_port"] > 65535:
            return False

        return True

    def _validate_grpc_config(self, config: Dict[str, Any]) -> bool:
        """Validate gRPC client configuration."""
        required_fields = ["server_host", "server_port"]
        for field in required_fields:
            if field not in config:
                return False

        # Validate types
        if not isinstance(config["server_host"], str):
            return False
        if not isinstance(config["server_port"], int):
            return False
        if config["server_port"] <= 0 or config["server_port"] > 65535:
            return False

        return True


# Global factory instance
_global_factory: Optional[ClientFactory] = None


def get_client_factory() -> ClientFactory:
    """Get global client factory instance."""
    global _global_factory
    if _global_factory is None:
        _global_factory = ClientFactory()
    return _global_factory


def create_client(protocol: ClientProtocol, **config) -> IToolplaneClient:
    """Convenience function to create a client."""
    factory = get_client_factory()
    return factory.create_client(protocol, config)


def register_client_type(
    protocol: ClientProtocol, client_class: Type, validator: Optional[callable] = None
) -> None:
    """Register a new client type with the global factory."""
    factory = get_client_factory()
    factory.register_client(protocol, client_class, validator)


# Convenience functions for specific protocols
def create_http_client(**config) -> IToolplaneClient:
    """Create an HTTP Toolplane client."""
    return create_client(ClientProtocol.HTTP, **config)


def create_grpc_client(**config) -> IToolplaneClient:
    """Create a gRPC Toolplane client."""
    return create_client(ClientProtocol.GRPC, **config)


def create_websocket_client(**config) -> IToolplaneClient:
    """Create a WebSocket Toolplane client."""
    return create_client(ClientProtocol.WEBSOCKET, **config)
