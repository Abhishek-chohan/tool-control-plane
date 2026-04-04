"""Client interface definitions for protocol-agnostic Toolplane clients."""

from abc import ABC, abstractmethod
from enum import Enum
from typing import Any, Callable, Dict, List, Optional, Protocol, runtime_checkable

from .session_interface import ISessionContext


class ClientProtocol(Enum):
    """Supported client protocols."""

    HTTP = "http"
    GRPC = "grpc"
    WEBSOCKET = "websocket"


@runtime_checkable
class IToolplaneClient(Protocol):
    """Protocol interface for Toolplane clients."""

    def connect(self) -> bool:
        """Connect to Toolplane server."""
        ...

    def disconnect(self) -> None:
        """Disconnect from Toolplane server."""
        ...

    def create_session(
        self,
        session_id: Optional[str] = None,
        user_id: Optional[str] = None,
        name: Optional[str] = None,
        description: Optional[str] = None,
        namespace: Optional[str] = None,
        register_machine: bool = False,
    ) -> ISessionContext:
        """Create a new session."""
        ...

    def get_session(self, session_id: str) -> Optional[ISessionContext]:
        """Get session context by ID."""
        ...

    def list_sessions(self) -> List[ISessionContext]:
        """List all session contexts."""
        ...

    def start(self) -> None:
        """Start the client."""
        ...

    def stop(self) -> None:
        """Stop the client."""
        ...


class IClientFactory(ABC):
    """Abstract factory for creating protocol-specific Toolplane clients."""

    @abstractmethod
    def create_client(
        self, protocol: ClientProtocol, config: Dict[str, Any]
    ) -> IToolplaneClient:
        """Create a client for the specified protocol."""
        pass

    @abstractmethod
    def get_supported_protocols(self) -> List[ClientProtocol]:
        """Get list of supported protocols."""
        pass

    @abstractmethod
    def validate_config(self, protocol: ClientProtocol, config: Dict[str, Any]) -> bool:
        """Validate configuration for the specified protocol."""
        pass


class BaseClientFactory(IClientFactory):
    """Base implementation of client factory with registry pattern."""

    def __init__(self):
        self._client_classes: Dict[ClientProtocol, type] = {}
        self._config_validators: Dict[ClientProtocol, Callable] = {}

    def register_client(
        self,
        protocol: ClientProtocol,
        client_class: type,
        config_validator: Optional[Callable] = None,
    ) -> None:
        """Register a client implementation for a protocol."""
        self._client_classes[protocol] = client_class
        if config_validator:
            self._config_validators[protocol] = config_validator

    def create_client(
        self, protocol: ClientProtocol, config: Dict[str, Any]
    ) -> IToolplaneClient:
        """Create a client for the specified protocol."""
        if protocol not in self._client_classes:
            raise ValueError(f"Unsupported protocol: {protocol}")

        if not self.validate_config(protocol, config):
            raise ValueError(f"Invalid configuration for protocol: {protocol}")

        client_class = self._client_classes[protocol]
        return client_class(**config)

    def get_supported_protocols(self) -> List[ClientProtocol]:
        """Get list of supported protocols."""
        return list(self._client_classes.keys())

    def validate_config(self, protocol: ClientProtocol, config: Dict[str, Any]) -> bool:
        """Validate configuration for the specified protocol."""
        if protocol in self._config_validators:
            try:
                return self._config_validators[protocol](config)
            except Exception:
                return False
        return True  # No validator means valid by default
