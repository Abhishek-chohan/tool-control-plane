"""HTTP configuration management for Toolplane client."""

from dataclasses import dataclass
from typing import Any, Dict, Optional

from ..common.base_config import BaseConfig
from ..common.constants import (
    DEFAULT_MAX_RETRIES,
    DEFAULT_REQUEST_TIMEOUT,
    DEFAULT_RETRY_BACKOFF_MS,
    HTTP_DEFAULT_PORT,
)


@dataclass
class HTTPClientConfig(BaseConfig):
    """Configuration for HTTP Toolplane client."""

    server_host: str = "localhost"
    server_port: int = HTTP_DEFAULT_PORT
    use_tls: bool = False
    tls_cert_path: Optional[str] = None
    tls_key_path: Optional[str] = None
    tls_ca_cert_path: Optional[str] = None

    def normalize(self):
        """Normalize HTTP-specific configuration values."""
        # Normalize server host
        if self.server_host.startswith("http://"):
            self.server_host = self.server_host[7:]
        elif self.server_host.startswith("https://"):
            self.server_host = self.server_host[8:]
            self.use_tls = True

        # Extract port from host if present
        if ":" in self.server_host:
            host_parts = self.server_host.split(":")
            if len(host_parts) == 2:
                try:
                    self.server_port = int(host_parts[1])
                    self.server_host = host_parts[0]
                except ValueError:
                    pass

        # Ensure timeout and retry values are reasonable
        if self.request_timeout <= 0:
            self.request_timeout = DEFAULT_REQUEST_TIMEOUT
        if self.max_retries < 0:
            self.max_retries = DEFAULT_MAX_RETRIES
        if self.retry_backoff_ms < 0:
            self.retry_backoff_ms = DEFAULT_RETRY_BACKOFF_MS

    def get_auth_info(self) -> Dict[str, Any]:
        """Get HTTP authentication information."""
        return {
            "headers": self.get_headers(),
            "server_url": self.server_url,
        }

    def get_headers(self) -> Dict[str, str]:
        """Get HTTP headers for requests."""
        headers = {"Content-Type": "application/json"}
        if self.api_key:
            headers["Authorization"] = f"Bearer {self.api_key}"
            # Keep legacy header for backward compatibility
            headers["Grpc-Metadata-api_key"] = self.api_key
        return headers

    def get_server_address(self) -> str:
        """Get formatted server address."""
        return f"{self.server_host}:{self.server_port}"

    @property
    def server_url(self) -> str:
        """Get complete server URL with protocol."""
        protocol = "https" if self.use_tls else "http"
        return f"{protocol}://{self.server_host}:{self.server_port}"

    def copy(self) -> "HTTPClientConfig":
        """Create a copy of the configuration."""
        return HTTPClientConfig(**self.to_dict())

    def merge(self, other: "HTTPClientConfig") -> "HTTPClientConfig":
        """Merge this configuration with another."""
        merged_dict = self.to_dict()
        merged_dict.update(other.to_dict())
        return HTTPClientConfig(**merged_dict)

    def to_dict(self) -> Dict[str, Any]:
        """Convert configuration to dictionary."""
        result = super().to_dict()
        result.update(
            {
                "server_url": self.server_url,
            }
        )
        return result
