"""Connection management for Toolplane client."""

from pathlib import Path
import random
import threading
import time
from typing import Optional

import grpc

from toolplane.proto.service_pb2_grpc import (
    MachinesServiceStub,
    RequestsServiceStub,
    SessionsServiceStub,
    TasksServiceStub,
    ToolServiceStub,
)

from .config import ClientConfig
from .errors import ConnectionError


class ConnectionManager:
    """Manages gRPC connections and stubs."""

    def __init__(self, config: ClientConfig):
        """Initialize connection manager with configuration."""
        self.config = config
        self.channel: Optional[grpc.Channel] = None
        self.tool_stub: Optional[ToolServiceStub] = None
        self.session_stub: Optional[SessionsServiceStub] = None
        self.machine_stub: Optional[MachinesServiceStub] = None
        self.requests_stub: Optional[RequestsServiceStub] = None
        self.tasks_stub: Optional[TasksServiceStub] = None
        self.connected = False
        self._connect_lock = threading.RLock()

        # Retry configuration
        self.max_retries = getattr(config, "max_retries", 3)
        self.retry_base_delay = getattr(config, "retry_base_delay", 1.0)
        self.retry_max_delay = getattr(config, "retry_max_delay", 60.0)
        self.retry_backoff_factor = getattr(config, "retry_backoff_factor", 2.0)
        self.connection_state = (
            "disconnected"  # disconnected, connecting, connected, failed
        )
        self.consecutive_failures = 0

    def _cleanup_failed_connection(self):
        """Reset stubs and close any existing channel after a failed attempt."""
        if self.channel is not None:
            try:
                self.channel.close()
            except Exception as e:
                print(f"Error closing channel: {e}")
        self._reset_stubs()
        self.connected = False
        self.connection_state = "failed"

    def _calculate_delay(self, attempt: int) -> float:
        """Calculate delay with exponential backoff and jitter."""
        delay = min(
            self.retry_base_delay * (self.retry_backoff_factor**attempt),
            self.retry_max_delay,
        )
        # Add jitter to prevent thundering herd
        jitter = random.uniform(0, 0.1 * delay)
        return delay + jitter

    def _should_retry(self, attempt: int, exception: Exception) -> bool:
        """Determine if we should retry based on exception type and attempt count."""
        if attempt >= self.max_retries:
            return False

        # Retry on specific gRPC status codes
        if isinstance(exception, grpc.RpcError):
            retryable_codes = [
                grpc.StatusCode.UNAVAILABLE,
                grpc.StatusCode.DEADLINE_EXCEEDED,
                grpc.StatusCode.RESOURCE_EXHAUSTED,
            ]
            return exception.code() in retryable_codes

        # Retry on network-related exceptions
        if isinstance(exception, (ConnectionError, TimeoutError)):
            return True

        return False

    def _load_tls_bytes(self, path_value: Optional[str], label: str) -> Optional[bytes]:
        """Load TLS material from disk when configured."""
        if not path_value:
            return None

        try:
            return Path(path_value).expanduser().read_bytes()
        except OSError as exc:
            raise ConnectionError(f"Failed to read {label}: {exc}") from exc

    def _channel_options(self) -> list[tuple[str, str]]:
        """Build gRPC channel options for TLS overrides."""
        server_name = getattr(self.config, "tls_server_name", None)
        if not server_name:
            return []

        return [
            ("grpc.ssl_target_name_override", server_name),
            ("grpc.default_authority", server_name),
        ]

    def _create_channel(self, target: str) -> grpc.Channel:
        """Create a secure or insecure gRPC channel from client configuration."""
        use_tls = bool(
            getattr(self.config, "use_tls", False)
            or getattr(self.config, "tls_ca_cert_path", None)
            or getattr(self.config, "tls_server_name", None)
            or getattr(self.config, "tls_cert_path", None)
            or getattr(self.config, "tls_key_path", None)
        )
        options = self._channel_options()
        if not use_tls:
            return grpc.insecure_channel(target, options=options)

        certificate_chain = self._load_tls_bytes(
            getattr(self.config, "tls_cert_path", None),
            "TLS client certificate",
        )
        private_key = self._load_tls_bytes(
            getattr(self.config, "tls_key_path", None),
            "TLS client key",
        )
        if bool(certificate_chain) != bool(private_key):
            raise ConnectionError(
                "TLS client authentication requires both certificate and key files"
            )

        root_certificates = self._load_tls_bytes(
            getattr(self.config, "tls_ca_cert_path", None),
            "TLS CA certificate",
        )
        credentials = grpc.ssl_channel_credentials(
            root_certificates=root_certificates,
            private_key=private_key,
            certificate_chain=certificate_chain,
        )
        return grpc.secure_channel(target, credentials, options=options)

    def connect(self) -> bool:
        """Establish connection to gRPC server with retry logic."""
        with self._connect_lock:
            if self.connected and self.channel is not None:
                return True

            last_exception = None
            ready_timeout = max(1, min(self.config.request_timeout, 10))

            for attempt in range(self.max_retries + 1):
                try:
                    if attempt > 0:
                        delay = self._calculate_delay(attempt - 1)
                        time.sleep(delay)
                        print(
                            f"Retry attempt {attempt}/{self.max_retries} after {delay:.2f}s delay"
                        )

                    self.connection_state = "connecting"
                    target = self.config.get_server_address()
                    self.channel = self._create_channel(target)

                    # Wait until the channel reports ready or timeout occurs
                    grpc.channel_ready_future(self.channel).result(
                        timeout=ready_timeout
                    )

                    self._initialize_stubs()
                    self.connected = True
                    self.connection_state = "connected"
                    self.consecutive_failures = 0
                    print(f"Successfully connected to {target}")
                    return True

                except Exception as e:
                    last_exception = e
                    self.consecutive_failures += 1
                    print(f"Connection attempt {attempt + 1} failed: {e}")
                    self._cleanup_failed_connection()

                    if attempt < self.max_retries and self._should_retry(attempt, e):
                        continue
                    break

            raise ConnectionError(
                f"Failed to connect after {self.max_retries + 1} attempts. Last error: {last_exception}"
            )

    def disconnect(self):
        """Close connection to gRPC server."""
        with self._connect_lock:
            if self.channel is not None:
                try:
                    self.channel.close()
                except Exception as e:
                    print(f"Error closing channel: {e}")
            self._reset_stubs()
            self.connected = False
            self.connection_state = "disconnected"

    def _initialize_stubs(self):
        """Initialize all gRPC service stubs."""
        if self.channel is None:
            raise ConnectionError("Channel is not initialized")

        self.tool_stub = ToolServiceStub(self.channel)
        self.session_stub = SessionsServiceStub(self.channel)
        self.machine_stub = MachinesServiceStub(self.channel)
        self.requests_stub = RequestsServiceStub(self.channel)
        self.tasks_stub = TasksServiceStub(self.channel)

    def _reset_stubs(self):
        """Reset all service stubs."""
        self.tool_stub = None
        self.session_stub = None
        self.machine_stub = None
        self.requests_stub = None
        self.tasks_stub = None
        self.channel = None

    def ensure_connected(self):
        """Ensure connection is established."""
        if not self.connected or self.channel is None:
            if not self.connect():
                raise ConnectionError("Failed to establish connection")

    def get_metadata(self):
        """Get metadata for gRPC calls."""
        return self.config.get_metadata()

    def mark_unhealthy(self):
        """Mark the current channel as unhealthy so the next call reconnects."""
        with self._connect_lock:
            if self.channel is not None:
                try:
                    self.channel.close()
                except Exception as e:
                    print(f"Error closing unhealthy channel: {e}")
            self._reset_stubs()
            self.connected = False
            self.connection_state = "disconnected"
