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
from ..core.errors import (
    ConnectionError,
    ToolplaneAPIError,
    ToolplaneResourceExhaustedError,
    ToolplaneUnavailableError,
    api_error_from_http_response,
)
from .http_config import HTTPClientConfig


class HTTPConnectionManager:
    """Manages HTTP connections and request handling."""

    def __init__(self, config: HTTPClientConfig):
        """Initialize HTTP connection manager with configuration."""
        self.config = config
        self.session = requests.Session()
        self.connected = False
        self.current_buffer_size = 0
        # Per-machine credential for provide-scoped RPCs; set by the machine
        # registration that minted it.
        self.machine_token: Optional[str] = None

    def set_machine_token(self, token: str):
        """Store the per-machine credential minted at registration."""
        self.machine_token = token or None

    def _request_headers(self) -> Dict:
        headers = self.config.get_headers()
        if self.machine_token:
            headers["X-Toolplane-Machine-Token"] = self.machine_token
        return headers

    def connect(self) -> bool:
        """Test connection to HTTP server."""
        try:
            self._post("api.v1/HealthCheck")
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

    def _error_from_response(self, response) -> ToolplaneAPIError:
        """Translate a non-2xx response into a typed error.

        Honors Retry-After on 429/503 (the backpressure statuses) before
        raising; the resulting error is retryable, so the with_retry wrapper
        sleeps the server-directed interval and tries again.
        """
        if response.status_code in (429, 503):
            retry_after = response.headers.get("Retry-After")
            if retry_after:
                time.sleep(float(retry_after))
        return api_error_from_http_response(
            response.status_code, response.text, context=f"POST {response.url}"
        )

    @with_retry(
        max_retries=DEFAULT_MAX_RETRIES,
        backoff_ms=DEFAULT_RETRY_BACKOFF_MS,
        exceptions=(
            ConnectionError,
            ToolplaneUnavailableError,
            ToolplaneResourceExhaustedError,
        ),
    )
    def _post(self, path: str, payload: Optional[Dict] = None) -> Any:
        """Make a POST request, retrying only transport and capacity errors.

        Deterministic failures (bad credentials, invalid input, state
        conflicts, missing entities) surface as typed ToolplaneAPIError
        subclasses and are not retried: the same request would fail again.
        """
        url = self.config.server_url.rstrip("/") + "/" + path
        headers = self._request_headers()

        try:
            response = self.session.post(
                url,
                json=payload or {},
                headers=headers,
                timeout=self.config.request_timeout,
            )

            if not response.ok:
                raise self._error_from_response(response)

            return response.json()

        except requests.exceptions.Timeout:
            raise ConnectionError(
                f"Request timed out after {self.config.request_timeout}s"
            )

        except requests.exceptions.RequestException as e:
            raise ConnectionError(f"HTTP request failed: {str(e)}")

    @with_retry(
        max_retries=DEFAULT_MAX_RETRIES,
        backoff_ms=DEFAULT_RETRY_BACKOFF_MS,
        exceptions=(
            ConnectionError,
            ToolplaneUnavailableError,
            ToolplaneResourceExhaustedError,
        ),
    )
    def stream_post(self, path: str, payload: Optional[Dict] = None):
        """Make a streaming POST request.

        Retries the same failure classes as _post: transport errors and
        capacity/backpressure (retryable typed errors)."""
        url = self.config.server_url.rstrip("/") + "/" + path
        headers = self._request_headers()

        try:
            response = self.session.post(
                url,
                json=payload or {},
                headers=headers,
                stream=True,
                timeout=self.config.request_timeout,
            )

            if response.status_code != 200:
                raise self._error_from_response(response)

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
        return self._post("api.v1/HealthCheck")

    # Session endpoints
    def create_session(self, payload: Dict):
        """Create a new session."""
        return self._post("api.v1/CreateSession", payload)

    def get_session(self, session_id: str):
        """Get session by ID."""
        return self._post(f"api.v1/sessions/{session_id}", {"sessionId": session_id})

    def list_sessions(self, user_id: str):
        """List sessions for user."""
        return self._post("api.v1/sessions", {"userId": user_id})

    def update_session(self, payload: Dict):
        """Update session."""
        return self._post("api.v1/UpdateSession", payload)

    def delete_session(self, session_id: str):
        """Delete session."""
        return self._post("api.v1/DeleteSession", {"sessionId": session_id})

    # New endpoints for user session management
    def list_user_sessions(self, payload: Dict):
        """List user sessions with pagination and filtering."""
        user_id = payload.get("userId", "")
        return self._post(f"api.v1/users/{user_id}/sessions", payload)

    def bulk_delete_sessions(self, payload: Dict):
        """Bulk delete sessions."""
        return self._post("api.v1/BulkDeleteSessions", payload)

    def get_session_stats(self, payload: Dict):
        """Get session statistics."""
        user_id = payload.get("userId", "")
        return self._post(f"api.v1/users/{user_id}/session-stats", payload)

    def invalidate_session(self, payload: Dict):
        """Invalidate session."""
        return self._post("api.v1/InvalidateSession", payload)

    def create_api_key(self, payload: Dict):
        """Create an API key for a session."""
        return self._post("api.v1/CreateApiKey", payload)

    def list_api_keys(self, payload: Dict):
        """List API keys for a session."""
        session_id = payload.get("sessionId", "")
        return self._post(f"api.v1/sessions/{session_id}/api-keys", payload)

    def revoke_api_key(self, payload: Dict):
        """Revoke an API key for a session."""
        return self._post("api.v1/RevokeApiKey", payload)

    # Tool endpoints
    def register_tool(self, payload: Dict):
        """Register a tool."""
        return self._post("api.v1/RegisterTool", payload)

    def list_tools(self, session_id: str):
        """List tools for session."""
        return self._post("api.v1/ListTools", {"sessionId": session_id})

    def get_tool_by_id(self, session_id: str, tool_id: str):
        """Get tool by ID."""
        return self._post(
            "api.v1/GetTool", {"sessionId": session_id, "toolId": tool_id}
        )

    def get_tool_by_name(self, session_id: str, tool_name: str):
        """Get tool by name."""
        return self._post(
            "api.v1/GetTool", {"sessionId": session_id, "toolName": tool_name}
        )

    def delete_tool(self, session_id: str, tool_id: str):
        """Delete tool."""
        return self._post(
            "api.v1/DeleteTool", {"sessionId": session_id, "toolId": tool_id}
        )

    # Machine endpoints
    def register_machine(self, payload: Dict):
        """Register a machine."""
        return self._post("api.v1/RegisterMachine", payload)

    def update_machine_ping(self, session_id: str, machine_id: str):
        """Update machine ping."""
        return self._post(
            "api.v1/UpdateMachinePing",
            {"sessionId": session_id, "machineId": machine_id},
        )

    def list_machines(self, session_id: str):
        """List machines for a session."""
        return self._post(
            f"api.v1/sessions/{session_id}/machines", {"sessionId": session_id}
        )

    def get_machine(self, session_id: str, machine_id: str):
        """Get a machine by ID."""
        return self._post(
            f"api.v1/sessions/{session_id}/machines/{machine_id}",
            {"sessionId": session_id, "machineId": machine_id},
        )

    def unregister_machine(self, session_id: str, machine_id: str):
        """Unregister machine."""
        return self._post(
            "api.v1/UnregisterMachine",
            {"sessionId": session_id, "machineId": machine_id},
        )

    def drain_machine(self, session_id: str, machine_id: str):
        """Drain machine."""
        return self._post(
            "api.v1/DrainMachine", {"sessionId": session_id, "machineId": machine_id}
        )

    # Request endpoints
    def create_request(self, payload: Dict):
        """Create a request."""
        return self._post("api.v1/CreateRequest", payload)

    def get_request(self, session_id: str, request_id: str):
        """Get request by ID."""
        return self._post(
            f"api.v1/sessions/{session_id}/requests/{request_id}",
            {"sessionId": session_id, "requestId": request_id},
        )

    def list_requests(self, payload: Dict):
        """List requests."""
        session_id = payload.get("sessionId", "")
        return self._post(f"api.v1/sessions/{session_id}/requests", payload)

    def update_request(self, payload: Dict):
        """Update request."""
        return self._post("api.v1/UpdateRequest", payload)

    def claim_request(self, session_id: str, request_id: str, machine_id: str):
        """Claim request."""
        return self._post(
            "api.v1/ClaimRequest",
            {"sessionId": session_id, "requestId": request_id, "machineId": machine_id},
        )

    def cancel_request(self, session_id: str, request_id: str):
        """Cancel request."""
        return self._post(
            "api.v1/CancelRequest",
            {"sessionId": session_id, "requestId": request_id},
        )

    def submit_request_result(self, payload: Dict):
        """Submit request result."""
        return self._post("api.v1/SubmitRequestResult", payload)

    def append_request_chunks(self, payload: Dict):
        """Append request chunks."""
        return self._post("api.v1/AppendRequestChunks", payload)

    def get_request_chunks(self, session_id: str, request_id: str):
        """Get request chunks."""
        return self._post(
            f"api.v1/sessions/{session_id}/requests/{request_id}/chunks",
            {"sessionId": session_id, "requestId": request_id},
        )

    def renew_request_lease(
        self, session_id: str, request_id: str, machine_id: str, lease_epoch: int
    ):
        """Renew the execution lease for a claimed/running request."""
        return self._post(
            "api.v1/RenewRequestLease",
            {
                "sessionId": session_id,
                "requestId": request_id,
                "machineId": machine_id,
                "leaseEpoch": lease_epoch,
            },
        )

    # Task endpoints
    def create_task(self, payload: Dict):
        """Create a task."""
        return self._post("api.v1/CreateTask", payload)

    def get_task(self, session_id: str, task_id: str):
        """Get task by ID."""
        return self._post(
            f"api.v1/sessions/{session_id}/tasks/{task_id}",
            {"sessionId": session_id, "taskId": task_id},
        )

    def list_tasks(self, session_id: str):
        """List tasks for a session."""
        return self._post(
            f"api.v1/sessions/{session_id}/tasks", {"sessionId": session_id}
        )

    def cancel_task(self, session_id: str, task_id: str):
        """Cancel task by ID."""
        return self._post(
            "api.v1/CancelTask", {"sessionId": session_id, "taskId": task_id}
        )

    # Execution endpoints
    def execute_tool(
        self,
        session_id: str,
        tool_name: str,
        input_data: str,
        idempotency_key: str = "",
    ):
        """Execute tool."""
        payload = {"sessionId": session_id, "toolName": tool_name, "input": input_data}
        if idempotency_key:
            payload["idempotencyKey"] = idempotency_key
        return self._post("api.v1/InvokeTool", payload)

    def stream_execute_tool(
        self,
        session_id: str,
        tool_name: str,
        input_data: str,
        idempotency_key: str = "",
    ):
        """Stream execute tool."""
        payload = {"sessionId": session_id, "toolName": tool_name, "input": input_data}
        if idempotency_key:
            payload["idempotencyKey"] = idempotency_key
        return self.stream_post("api.v1/StreamExecuteTool", payload)
