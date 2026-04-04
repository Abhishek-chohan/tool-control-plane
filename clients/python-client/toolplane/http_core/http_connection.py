"""HTTP connection management for Toolplane client."""

import time
from typing import Any, Dict, Optional

import requests

from ..common.constants import (
    DEFAULT_MAX_RETRIES,
    DEFAULT_RETRY_BACKOFF_MS,
    ERROR_CONNECTION_FAILED,
)
from ..common.utils import format_error_message, with_retry
from ..core.errors import ConnectionError
from .http_config import HTTPClientConfig


class HTTPConnectionManager:
    """Manages HTTP connections and request handling."""

    def __init__(self, config: HTTPClientConfig):
        """Initialize HTTP connection manager with configuration."""
        self.config = config
        self.session = requests.Session()
        self.connected = False
        self.current_buffer_size = 0

    def connect(self) -> bool:
        """Test connection to HTTP server."""
        try:
            self._post("api/HealthCheck")
            self.connected = True
            return True
        except Exception as e:
            error_msg = format_error_message(e, ERROR_CONNECTION_FAILED)
            raise ConnectionError(error_msg)

    def disconnect(self):
        """Close HTTP session."""
        if self.session:
            self.session.close()
        self.connected = False

    def ensure_connected(self):
        """Ensure connection is established."""
        if not self.connected:
            if not self.connect():
                raise ConnectionError("Failed to establish HTTP connection")

    @with_retry(max_retries=DEFAULT_MAX_RETRIES, backoff_ms=DEFAULT_RETRY_BACKOFF_MS)
    def _post(self, path: str, payload: Optional[Dict] = None) -> Any:
        """Make a POST request with retry logic and backpressure handling."""
        url = self.config.server_url.rstrip("/") + "/" + path
        headers = self.config.get_headers()

        try:
            response = self.session.post(
                url,
                json=payload or {},
                headers=headers,
                timeout=self.config.request_timeout,
            )

            # Handle backpressure response codes
            if response.status_code in (429, 503):
                # Get retry-after header or use exponential backoff
                retry_after = response.headers.get("Retry-After")
                if retry_after:
                    time.sleep(float(retry_after))
                raise ConnectionError(f"Server busy (HTTP {response.status_code})")

            # Handle successful response
            if response.ok:
                return response.json()
            else:
                raise ConnectionError(f"HTTP {response.status_code} {response.text}")

        except requests.exceptions.Timeout:
            raise ConnectionError(
                f"Request timed out after {self.config.request_timeout}s"
            )

        except requests.exceptions.RequestException as e:
            raise ConnectionError(f"HTTP request failed: {str(e)}")

    @with_retry(max_retries=DEFAULT_MAX_RETRIES, backoff_ms=DEFAULT_RETRY_BACKOFF_MS)
    def stream_post(self, path: str, payload: Optional[Dict] = None):
        """Make a streaming POST request."""
        url = self.config.server_url.rstrip("/") + "/" + path
        headers = self.config.get_headers()

        try:
            response = self.session.post(
                url,
                json=payload or {},
                headers=headers,
                stream=True,
                timeout=self.config.request_timeout,
            )

            # Handle backpressure responses
            if response.status_code in (429, 503):
                retry_after = response.headers.get("Retry-After")
                if retry_after:
                    time.sleep(float(retry_after))
                raise ConnectionError(f"Server busy (HTTP {response.status_code})")

            if response.status_code != 200:
                raise ConnectionError(f"HTTP {response.status_code} {response.text}")

            return response

        except requests.exceptions.Timeout:
            raise ConnectionError(
                f"Stream request timed out after {self.config.request_timeout}s"
            )

        except requests.exceptions.RequestException as e:
            raise ConnectionError(f"Stream request failed: {str(e)}")

    # Health check
    def health_check(self):
        """Check server health."""
        return self._post("api/HealthCheck")

    # Session endpoints
    def create_session(self, payload: Dict):
        """Create a new session."""
        return self._post("api/CreateSession", payload)

    def get_session(self, session_id: str):
        """Get session by ID."""
        return self._post("api/GetSession", {"sessionId": session_id})

    def list_sessions(self, user_id: str):
        """List sessions for user."""
        return self._post("api/ListSessions", {"userId": user_id})

    def update_session(self, payload: Dict):
        """Update session."""
        return self._post("api/UpdateSession", payload)

    def delete_session(self, session_id: str):
        """Delete session."""
        return self._post("api/DeleteSession", {"sessionId": session_id})

    # New endpoints for user session management
    def list_user_sessions(self, payload: Dict):
        """List user sessions with pagination and filtering."""
        return self._post("api/ListUserSessions", payload)

    def bulk_delete_sessions(self, payload: Dict):
        """Bulk delete sessions."""
        return self._post("api/BulkDeleteSessions", payload)

    def get_session_stats(self, payload: Dict):
        """Get session statistics."""
        return self._post("api/GetSessionStats", payload)

    def refresh_session_token(self, payload: Dict):
        """Refresh session token."""
        return self._post("api/RefreshSessionToken", payload)

    def invalidate_session(self, payload: Dict):
        """Invalidate session."""
        return self._post("api/InvalidateSession", payload)

    def create_api_key(self, payload: Dict):
        """Create an API key for a session."""
        return self._post("api/CreateApiKey", payload)

    def list_api_keys(self, payload: Dict):
        """List API keys for a session."""
        return self._post("api/ListApiKeys", payload)

    def revoke_api_key(self, payload: Dict):
        """Revoke an API key for a session."""
        return self._post("api/RevokeApiKey", payload)

    # Tool endpoints
    def register_tool(self, payload: Dict):
        """Register a tool."""
        return self._post("api/RegisterTool", payload)

    def list_tools(self, session_id: str):
        """List tools for session."""
        return self._post("api/ListTools", {"sessionId": session_id})

    def get_tool_by_id(self, session_id: str, tool_id: str):
        """Get tool by ID."""
        return self._post(
            "api/GetToolById", {"sessionId": session_id, "toolId": tool_id}
        )

    def get_tool_by_name(self, session_id: str, tool_name: str):
        """Get tool by name."""
        return self._post(
            "api/GetToolByName", {"sessionId": session_id, "toolName": tool_name}
        )

    def delete_tool(self, session_id: str, tool_id: str):
        """Delete tool."""
        return self._post(
            "api/DeleteTool", {"sessionId": session_id, "toolId": tool_id}
        )

    # Machine endpoints
    def register_machine(self, payload: Dict):
        """Register a machine."""
        return self._post("api/RegisterMachine", payload)

    def update_machine_ping(self, session_id: str, machine_id: str):
        """Update machine ping."""
        return self._post(
            "api/UpdateMachinePing", {"sessionId": session_id, "machineId": machine_id}
        )

    def list_machines(self, session_id: str):
        """List machines for a session."""
        return self._post("api/ListMachines", {"sessionId": session_id})

    def get_machine(self, session_id: str, machine_id: str):
        """Get a machine by ID."""
        return self._post(
            "api/GetMachine", {"sessionId": session_id, "machineId": machine_id}
        )

    def unregister_machine(self, session_id: str, machine_id: str):
        """Unregister machine."""
        return self._post(
            "api/UnregisterMachine", {"sessionId": session_id, "machineId": machine_id}
        )

    def drain_machine(self, session_id: str, machine_id: str):
        """Drain machine."""
        return self._post(
            "api/DrainMachine", {"sessionId": session_id, "machineId": machine_id}
        )

    # Request endpoints
    def create_request(self, payload: Dict):
        """Create a request."""
        return self._post("api/CreateRequest", payload)

    def get_request(self, session_id: str, request_id: str):
        """Get request by ID."""
        return self._post(
            "api/GetRequest", {"sessionId": session_id, "requestId": request_id}
        )

    def list_requests(self, payload: Dict):
        """List requests."""
        return self._post("api/ListRequests", payload)

    def update_request(self, payload: Dict):
        """Update request."""
        return self._post("api/UpdateRequest", payload)

    def claim_request(self, session_id: str, request_id: str, machine_id: str):
        """Claim request."""
        return self._post(
            "api/ClaimRequest",
            {"sessionId": session_id, "requestId": request_id, "machineId": machine_id},
        )

    def cancel_request(self, session_id: str, request_id: str):
        """Cancel request."""
        return self._post(
            "api/CancelRequest",
            {"sessionId": session_id, "requestId": request_id},
        )

    def submit_request_result(self, payload: Dict):
        """Submit request result."""
        return self._post("api/SubmitRequestResult", payload)

    def append_request_chunks(self, payload: Dict):
        """Append request chunks."""
        return self._post("api/AppendRequestChunks", payload)

    def get_request_chunks(self, session_id: str, request_id: str):
        """Get request chunks."""
        return self._post(
            "api/GetRequestChunks", {"sessionId": session_id, "requestId": request_id}
        )

    # Task endpoints
    def create_task(self, payload: Dict):
        """Create a task."""
        return self._post("api/CreateTask", payload)

    def get_task(self, session_id: str, task_id: str):
        """Get task by ID."""
        return self._post("api/GetTask", {"sessionId": session_id, "taskId": task_id})

    def list_tasks(self, session_id: str):
        """List tasks for a session."""
        return self._post("api/ListTasks", {"sessionId": session_id})

    def cancel_task(self, session_id: str, task_id: str):
        """Cancel task by ID."""
        return self._post(
            "api/CancelTask", {"sessionId": session_id, "taskId": task_id}
        )

    # Execution endpoints
    def execute_tool(self, session_id: str, tool_name: str, input_data: str):
        """Execute tool."""
        return self._post(
            "api/ExecuteTool",
            {"sessionId": session_id, "toolName": tool_name, "input": input_data},
        )

    def stream_execute_tool(self, session_id: str, tool_name: str, input_data: str):
        """Stream execute tool."""
        return self.stream_post(
            "api/StreamExecuteTool",
            {"sessionId": session_id, "toolName": tool_name, "input": input_data},
        )
