"""Base connection manager class shared between gRPC and HTTP clients."""

import threading
import time
from abc import ABC, abstractmethod
from typing import Any, Dict

try:
    from ..core.errors import ConnectionError
except ImportError:
    # Fallback for standalone testing
    class ConnectionError(Exception):
        """Connection error for standalone testing."""

        pass


from .constants import (
    DEFAULT_MAX_RETRIES,
    DEFAULT_RETRY_BACKOFF_MS,
)
from .utils import with_retry


class BaseConnectionManager(ABC):
    """Base class for connection management."""

    def __init__(self, config):
        """Initialize base connection manager."""
        self.config = config
        self.connected = False
        self.connection_lock = threading.RLock()
        self.last_heartbeat = 0
        self.connection_stats = {
            "connect_time": 0,
            "last_activity": 0,
            "total_requests": 0,
            "failed_requests": 0,
            "reconnect_count": 0,
        }

    @abstractmethod
    def connect(self):
        """Establish connection (protocol-specific)."""
        pass

    @abstractmethod
    def disconnect(self):
        """Close connection (protocol-specific)."""
        pass

    @abstractmethod
    def is_connected(self) -> bool:
        """Check if connection is active (protocol-specific)."""
        pass

    def ensure_connected(self):
        """Ensure connection is established."""
        if not self.is_connected():
            self.connect()

    @with_retry(max_retries=DEFAULT_MAX_RETRIES, backoff_ms=DEFAULT_RETRY_BACKOFF_MS)
    def reconnect(self):
        """Reconnect with retry logic."""
        try:
            self.disconnect()
            time.sleep(0.1)  # Brief pause before reconnecting
            self.connect()

            with self.connection_lock:
                self.connection_stats["reconnect_count"] += 1

        except Exception as e:
            raise ConnectionError(f"Failed to reconnect: {e}")

    def health_check(self) -> bool:
        """Perform health check."""
        try:
            return self._perform_health_check()
        except Exception:
            return False

    @abstractmethod
    def _perform_health_check(self) -> bool:
        """Perform protocol-specific health check."""
        pass

    def heartbeat(self):
        """Send heartbeat to server."""
        try:
            if self.is_connected():
                current_time = time.time()
                if current_time - self.last_heartbeat >= self.config.heartbeat_interval:
                    self._send_heartbeat()
                    self.last_heartbeat = current_time
        except Exception:
            # Heartbeat failures are not critical
            pass

    @abstractmethod
    def _send_heartbeat(self):
        """Send protocol-specific heartbeat."""
        pass

    def get_connection_info(self) -> Dict[str, Any]:
        """Get connection information."""
        with self.connection_lock:
            return {
                "connected": self.is_connected(),
                "config": self.config.to_dict(),
                "stats": self.connection_stats.copy(),
                "last_heartbeat": self.last_heartbeat,
            }

    def reset_stats(self):
        """Reset connection statistics."""
        with self.connection_lock:
            self.connection_stats = {
                "connect_time": time.time() if self.is_connected() else 0,
                "last_activity": time.time(),
                "total_requests": 0,
                "failed_requests": 0,
                "reconnect_count": 0,
            }

    def _record_request(self, success: bool = True):
        """Record request statistics."""
        with self.connection_lock:
            self.connection_stats["total_requests"] += 1
            self.connection_stats["last_activity"] = time.time()
            if not success:
                self.connection_stats["failed_requests"] += 1

    def get_request_stats(self) -> Dict[str, Any]:
        """Get request statistics."""
        with self.connection_lock:
            stats = self.connection_stats.copy()

            # Calculate success rate
            total = stats["total_requests"]
            failed = stats["failed_requests"]
            success_rate = ((total - failed) / total * 100) if total > 0 else 0

            return {
                "total_requests": total,
                "failed_requests": failed,
                "success_rate": success_rate,
                "reconnect_count": stats["reconnect_count"],
                "uptime": (
                    time.time() - stats["connect_time"]
                    if stats["connect_time"] > 0
                    else 0
                ),
                "last_activity": stats["last_activity"],
            }

    def __enter__(self):
        """Context manager entry."""
        self.connect()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.disconnect()

    def __del__(self):
        """Destructor to ensure connection is closed."""
        try:
            self.disconnect()
        except Exception:
            pass
