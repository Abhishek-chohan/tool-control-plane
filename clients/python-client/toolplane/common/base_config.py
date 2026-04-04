"""Base configuration class shared between gRPC and HTTP clients."""

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import Any, Dict, Optional

from .constants import (
    DEFAULT_BUFFER_SIZE,
    DEFAULT_HEARTBEAT_INTERVAL,
    DEFAULT_MAX_RETRIES,
    DEFAULT_MAX_WORKERS,
    DEFAULT_POLL_INTERVAL,
    DEFAULT_REQUEST_TIMEOUT,
    DEFAULT_RETRY_BACKOFF_MS,
)
from .utils import validate_description, validate_session_name


@dataclass
class BaseConfig(ABC):
    """Base configuration for Toolplane clients."""

    # Common configuration fields
    api_key: Optional[str] = None
    user_id: Optional[str] = None
    session_name: Optional[str] = None
    session_description: Optional[str] = None
    session_namespace: Optional[str] = None

    # Performance settings
    heartbeat_interval: int = DEFAULT_HEARTBEAT_INTERVAL
    max_workers: int = DEFAULT_MAX_WORKERS
    poll_interval: float = DEFAULT_POLL_INTERVAL
    request_timeout: int = DEFAULT_REQUEST_TIMEOUT
    max_retries: int = DEFAULT_MAX_RETRIES
    max_buffer_size: int = DEFAULT_BUFFER_SIZE
    retry_backoff_ms: int = DEFAULT_RETRY_BACKOFF_MS

    # Additional configuration
    extra_config: Dict[str, Any] = field(default_factory=dict)

    def __post_init__(self):
        """Validate configuration after initialization."""
        self.validate()
        self.normalize()

    def validate(self):
        """Validate configuration parameters."""
        if self.session_name and not validate_session_name(self.session_name):
            raise ValueError(f"Invalid session name: {self.session_name}")

        if self.session_description and not validate_description(
            self.session_description
        ):
            raise ValueError(
                f"Session description too long: {len(self.session_description)} characters"
            )

        if self.heartbeat_interval <= 0:
            raise ValueError("Heartbeat interval must be positive")

        if self.max_workers <= 0:
            raise ValueError("Max workers must be positive")

        if self.poll_interval <= 0:
            raise ValueError("Poll interval must be positive")

        if self.request_timeout <= 0:
            raise ValueError("Request timeout must be positive")

        if self.max_retries < 0:
            raise ValueError("Max retries must be non-negative")

        if self.max_buffer_size <= 0:
            raise ValueError("Max buffer size must be positive")

        if self.retry_backoff_ms < 0:
            raise ValueError("Retry backoff must be non-negative")

    @abstractmethod
    def normalize(self):
        """Normalize configuration values (protocol-specific)."""
        pass

    @abstractmethod
    def get_auth_info(self) -> Dict[str, Any]:
        """Get authentication information (protocol-specific)."""
        pass

    def get_common_fields(self) -> Dict[str, Any]:
        """Get common configuration fields."""
        return {
            "api_key": self.api_key,
            "user_id": self.user_id,
            "session_name": self.session_name,
            "session_description": self.session_description,
            "session_namespace": self.session_namespace,
            "heartbeat_interval": self.heartbeat_interval,
            "max_workers": self.max_workers,
            "poll_interval": self.poll_interval,
            "request_timeout": self.request_timeout,
            "max_retries": self.max_retries,
            "max_buffer_size": self.max_buffer_size,
            "retry_backoff_ms": self.retry_backoff_ms,
        }

    def update_from_dict(self, config_dict: Dict[str, Any]):
        """Update configuration from dictionary."""
        for key, value in config_dict.items():
            if hasattr(self, key):
                setattr(self, key, value)
            else:
                self.extra_config[key] = value

    def to_dict(self) -> Dict[str, Any]:
        """Convert configuration to dictionary."""
        result = self.get_common_fields()
        result.update(self.extra_config)
        return result

    def copy(self) -> "BaseConfig":
        """Create a copy of the configuration."""
        # This will be implemented by subclasses
        raise NotImplementedError("Subclasses must implement copy method")

    def merge(self, other: "BaseConfig") -> "BaseConfig":
        """Merge this configuration with another."""
        # This will be implemented by subclasses
        raise NotImplementedError("Subclasses must implement merge method")
