"""Configuration management for Toolplane gRPC client."""

from dataclasses import dataclass
from typing import Any, Dict, List, Optional

from ..common.base_config import BaseConfig
from ..common.constants import GRPC_DEFAULT_PORT


@dataclass
class ClientConfig(BaseConfig):
    """Configuration for Toolplane gRPC client."""

    server_host: str = "localhost"
    server_port: int = GRPC_DEFAULT_PORT
    use_tls: bool = False
    tls_cert_path: Optional[str] = None
    tls_key_path: Optional[str] = None
    tls_ca_cert_path: Optional[str] = None
    tls_server_name: Optional[str] = None
    # Retry configuration parameters
    max_retries: int = 3
    retry_base_delay: float = 1.0
    retry_max_delay: float = 60.0
    retry_backoff_factor: float = 2.0

    def normalize(self):
        """Normalize gRPC-specific configuration values."""
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

        # Ensure retry values are reasonable
        if self.max_retries < 0:
            self.max_retries = 3
        if self.retry_base_delay <= 0:
            self.retry_base_delay = 1.0
        if self.retry_max_delay <= 0:
            self.retry_max_delay = 60.0
        if self.retry_backoff_factor <= 1.0:
            self.retry_backoff_factor = 2.0

    def get_auth_info(self) -> Dict[str, Any]:
        """Get gRPC authentication information."""
        return {
            "metadata": self.get_metadata(),
            "use_tls": self.use_tls,
            "tls_cert_path": self.tls_cert_path,
            "tls_key_path": self.tls_key_path,
            "tls_ca_cert_path": self.tls_ca_cert_path,
            "tls_server_name": self.tls_server_name,
        }

    def get_metadata(self) -> List[tuple]:
        """Get gRPC metadata headers."""
        metadata = []
        if self.api_key:
            metadata.append(("authorization", f"Bearer {self.api_key}"))
            metadata.append(("api_key", self.api_key))
        return metadata

    def get_server_address(self) -> str:
        """Get formatted server address."""
        return f"{self.server_host}:{self.server_port}"

    def copy(self) -> "ClientConfig":
        """Create a copy of the configuration."""
        return ClientConfig(**self.to_dict())

    def merge(self, other: "ClientConfig") -> "ClientConfig":
        """Merge this configuration with another."""
        merged_dict = self.to_dict()
        merged_dict.update(other.to_dict())
        return ClientConfig(**merged_dict)

    def to_dict(self) -> Dict[str, Any]:
        """Convert configuration to dictionary."""
        result = super().to_dict()
        result.update(
            {
                "server_host": self.server_host,
                "server_port": self.server_port,
                "use_tls": self.use_tls,
                "tls_cert_path": self.tls_cert_path,
                "tls_key_path": self.tls_key_path,
                "tls_ca_cert_path": self.tls_ca_cert_path,
                "tls_server_name": self.tls_server_name,
                "max_retries": self.max_retries,
                "retry_base_delay": self.retry_base_delay,
                "retry_max_delay": self.retry_max_delay,
                "retry_backoff_factor": self.retry_backoff_factor,
            }
        )
        return result
