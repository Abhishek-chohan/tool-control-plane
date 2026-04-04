"""Connection management interface definitions."""

from abc import ABC, abstractmethod
from enum import Enum
from typing import Any, Dict, Protocol, runtime_checkable


class ConnectionState(Enum):
    """Connection states."""

    DISCONNECTED = "disconnected"
    CONNECTING = "connecting"
    CONNECTED = "connected"
    RECONNECTING = "reconnecting"
    ERROR = "error"


class ConnectionStrategy(Enum):
    """Connection strategies."""

    DIRECT = "direct"
    POOLED = "pooled"
    LOAD_BALANCED = "load_balanced"
    CIRCUIT_BREAKER = "circuit_breaker"


@runtime_checkable
class IConnectionManager(Protocol):
    """Protocol interface for connection management."""

    @property
    def connected(self) -> bool:
        """Check if connected to server."""
        ...

    @property
    def state(self) -> ConnectionState:
        """Get current connection state."""
        ...

    def connect(self) -> bool:
        """Connect to server."""
        ...

    def disconnect(self) -> None:
        """Disconnect from server."""
        ...

    def health_check(self) -> bool:
        """Perform health check."""
        ...

    def get_connection_info(self) -> Dict[str, Any]:
        """Get connection information."""
        ...


class IConnectionStrategy(ABC):
    """Abstract strategy for connection management."""

    @abstractmethod
    def connect(self, config: Dict[str, Any]) -> bool:
        """Execute connection strategy."""
        pass

    @abstractmethod
    def disconnect(self) -> None:
        """Execute disconnection strategy."""
        pass

    @abstractmethod
    def health_check(self) -> bool:
        """Execute health check strategy."""
        pass

    @abstractmethod
    def get_metrics(self) -> Dict[str, Any]:
        """Get connection metrics."""
        pass


class DirectConnectionStrategy(IConnectionStrategy):
    """Direct connection strategy implementation."""

    def __init__(self):
        self._connected = False
        self._connection_info = {}

    def connect(self, config: Dict[str, Any]) -> bool:
        """Establish direct connection."""
        try:
            # Implementation would depend on protocol
            # This is a template for the strategy pattern
            self._connection_info = {
                "host": config.get("server_host", "localhost"),
                "port": config.get("server_port", 8080),
                "protocol": config.get("protocol", "http"),
                "connected_at": __import__("time").time(),
            }
            self._connected = True
            return True
        except Exception:
            self._connected = False
            return False

    def disconnect(self) -> None:
        """Close direct connection."""
        self._connected = False
        self._connection_info.clear()

    def health_check(self) -> bool:
        """Check direct connection health."""
        return self._connected

    def get_metrics(self) -> Dict[str, Any]:
        """Get direct connection metrics."""
        return {
            "connected": self._connected,
            "connection_info": self._connection_info.copy(),
            "strategy": "direct",
        }


class PooledConnectionStrategy(IConnectionStrategy):
    """Pooled connection strategy implementation."""

    def __init__(self, pool_size: int = 5):
        self.pool_size = pool_size
        self._pool = []
        self._active_connections = 0

    def connect(self, config: Dict[str, Any]) -> bool:
        """Establish pooled connections."""
        try:
            # Initialize connection pool
            for _ in range(self.pool_size):
                conn_info = {
                    "id": len(self._pool),
                    "host": config.get("server_host", "localhost"),
                    "port": config.get("server_port", 8080),
                    "active": False,
                    "created_at": __import__("time").time(),
                }
                self._pool.append(conn_info)

            self._active_connections = self.pool_size
            return True
        except Exception:
            return False

    def disconnect(self) -> None:
        """Close all pooled connections."""
        self._pool.clear()
        self._active_connections = 0

    def health_check(self) -> bool:
        """Check pool health."""
        return self._active_connections > 0

    def get_metrics(self) -> Dict[str, Any]:
        """Get pooled connection metrics."""
        return {
            "pool_size": len(self._pool),
            "active_connections": self._active_connections,
            "strategy": "pooled",
        }


class ConnectionStrategyFactory:
    """Factory for creating connection strategies."""

    _strategies = {
        ConnectionStrategy.DIRECT: DirectConnectionStrategy,
        ConnectionStrategy.POOLED: PooledConnectionStrategy,
    }

    @classmethod
    def create_strategy(
        self, strategy_type: ConnectionStrategy, **kwargs
    ) -> IConnectionStrategy:
        """Create a connection strategy instance."""
        if strategy_type not in self._strategies:
            raise ValueError(f"Unsupported connection strategy: {strategy_type}")

        strategy_class = self._strategies[strategy_type]
        return strategy_class(**kwargs)

    @classmethod
    def register_strategy(
        cls, strategy_type: ConnectionStrategy, strategy_class: type
    ) -> None:
        """Register a new connection strategy."""
        cls._strategies[strategy_type] = strategy_class
