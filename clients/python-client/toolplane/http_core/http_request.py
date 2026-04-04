"""HTTP request management for Toolplane client."""

import json
import threading
import time
from concurrent.futures import ThreadPoolExecutor
from typing import Any, Callable, Dict, List, Optional

from ..common.constants import (
    DEFAULT_MAX_WORKERS,
    DEFAULT_POLL_INTERVAL,
)
from ..core.errors import RequestError
from .http_connection import HTTPConnectionManager


class HTTPRequestManager:
    """Manages request processing and polling for HTTP client."""

    def __init__(
        self,
        connection_manager: HTTPConnectionManager,
        max_workers: int = DEFAULT_MAX_WORKERS,
    ):
        """Initialize HTTP request manager."""
        self.connection_manager = connection_manager
        self.executor = ThreadPoolExecutor(max_workers=max_workers)
        self._running = False
        self._poll_thread: Optional[threading.Thread] = None
        self._poll_interval = DEFAULT_POLL_INTERVAL

    def _normalize_request(self, response: Dict[str, Any]) -> Dict[str, Any]:
        normalized = {
            "id": response.get("id"),
            "sessionId": response.get("sessionId", response.get("session_id")),
            "toolName": response.get("toolName", response.get("tool_name")),
            "status": response.get("status"),
            "input": response.get("input"),
            "createdAt": response.get("createdAt", response.get("created_at")),
            "updatedAt": response.get("updatedAt", response.get("updated_at")),
            "executingMachineId": response.get(
                "executingMachineId", response.get("executing_machine_id")
            ),
        }

        if response.get("result") not in (None, ""):
            try:
                normalized["result"] = json.loads(response.get("result"))
            except Exception:
                normalized["result"] = response.get("result")

        result_type = response.get("resultType", response.get("result_type"))
        if result_type:
            normalized["resultType"] = result_type

        if response.get("error"):
            normalized["error"] = response.get("error")

        stream_results = response.get("streamResults", response.get("stream_results"))
        if isinstance(stream_results, list) and stream_results:
            normalized["streamResults"] = stream_results

        return normalized

    def start_polling(self, poll_interval: float = DEFAULT_POLL_INTERVAL):
        """Start request polling."""
        if self._running:
            return

        self._running = True
        self._poll_interval = poll_interval
        self._poll_thread = threading.Thread(target=self._poll_loop, daemon=True)
        self._poll_thread.start()

    def stop_polling(self):
        """Stop request polling."""
        self._running = False
        if self._poll_thread:
            self._poll_thread.join(timeout=1)

    def _poll_loop(self):
        """Main polling loop."""
        while self._running:
            try:
                # This will be called by the main client with session info
                time.sleep(self._poll_interval)
            except Exception:
                # Ignore polling errors
                pass

    def poll_session_requests(
        self,
        session_id: str,
        machine_id: str,
        tools: Dict[str, Callable],
        streaming_tools: set,
        limit: int = 5,
    ):
        """Poll for requests in a specific session."""
        try:
            self.connection_manager.ensure_connected()

            # Get pending requests
            payload = {"sessionId": session_id, "status": "pending", "limit": limit}

            response = self.connection_manager.list_requests(payload)
            requests = response.get("requests", [])

            # Process each request
            for req in requests:
                request_id = req.get("id")

                try:
                    # Claim the request
                    self.connection_manager.claim_request(
                        session_id, request_id, machine_id
                    )

                    # Execute in thread pool
                    self.executor.submit(
                        self._execute_request, req, tools, streaming_tools
                    )

                except Exception:
                    # Ignore claim errors (request might be claimed by another machine)
                    pass

        except Exception as e:
            raise RequestError(f"Failed to poll requests for session {session_id}: {e}")

    def _execute_request(
        self, request, tools: Dict[str, Callable], streaming_tools: set
    ):
        """Execute a claimed request."""
        tool_name = request.get("toolName")
        request_id = request.get("id")
        session_id = request.get("sessionId")

        if tool_name not in tools:
            self._submit_error_result(
                session_id, request_id, f"Tool '{tool_name}' not found"
            )
            return

        try:
            # Parse input parameters
            try:
                params = json.loads(request.get("input", "{}"))
            except json.JSONDecodeError:
                self._submit_error_result(session_id, request_id, "Invalid JSON input")
                return

            # Execute the tool
            tool_func = tools[tool_name]
            is_streaming = tool_name in streaming_tools

            self._handle_tool_execution(
                session_id, request_id, tool_func, params, is_streaming
            )

        except Exception as e:
            self._submit_error_result(session_id, request_id, str(e))

    def _handle_tool_execution(
        self,
        session_id: str,
        request_id: str,
        tool_func: Callable,
        params: Dict,
        is_streaming: bool,
    ):
        """Handle tool execution (streaming or non-streaming)."""
        try:
            # Mark as running
            self._update_request_status(session_id, request_id, "running")

            if is_streaming:
                self._handle_streaming_execution(
                    session_id, request_id, tool_func, params
                )
            else:
                self._handle_normal_execution(session_id, request_id, tool_func, params)

        except Exception as e:
            self._submit_error_result(session_id, request_id, str(e))

    def _handle_streaming_execution(
        self, session_id: str, request_id: str, tool_func: Callable, params: Dict
    ):
        """Handle streaming tool execution."""
        # Set streaming mode
        self._update_request(session_id, request_id, result_type="streaming")

        chunks = []
        for chunk in tool_func(**params):
            data = chunk if isinstance(chunk, str) else json.dumps(chunk)

            # Append chunk
            self._append_request_chunk(session_id, request_id, data)
            chunks.append(data)

        # Submit final result
        self._submit_result(session_id, request_id, json.dumps(chunks), "resolution")

    def _handle_normal_execution(
        self, session_id: str, request_id: str, tool_func: Callable, params: Dict
    ):
        """Handle normal tool execution."""
        result = tool_func(**params)
        self._submit_result(session_id, request_id, json.dumps(result), "resolution")

    def _update_request_status(self, session_id: str, request_id: str, status: str):
        """Update request status."""
        payload = {"sessionId": session_id, "requestId": request_id, "status": status}

        self.connection_manager.update_request(payload)

    def _update_request(self, session_id: str, request_id: str, result_type: str):
        """Update request with result type."""
        payload = {
            "sessionId": session_id,
            "requestId": request_id,
            "resultType": result_type,
        }

        self.connection_manager.update_request(payload)

    def _append_request_chunk(self, session_id: str, request_id: str, chunk: str):
        """Append chunk to request."""
        payload = {
            "sessionId": session_id,
            "requestId": request_id,
            "chunks": [chunk],
            "resultType": "streaming",
        }

        self.connection_manager.append_request_chunks(payload)

    def _submit_result(
        self, session_id: str, request_id: str, result: str, result_type: str
    ):
        """Submit request result."""
        payload = {
            "sessionId": session_id,
            "requestId": request_id,
            "result": result,
            "resultType": result_type,
            "meta": {},
        }

        self.connection_manager.submit_request_result(payload)

    def _submit_error_result(self, session_id: str, request_id: str, error: str):
        """Submit error result."""
        self._submit_result(
            session_id, request_id, json.dumps({"error": error}), "rejection"
        )

    def get_request_status(self, session_id: str, request_id: str) -> Dict[str, Any]:
        """Get request status."""
        try:
            self.connection_manager.ensure_connected()

            response = self.connection_manager.get_request(session_id, request_id)

            result = self._normalize_request(response)

            try:
                chunk_response = self.connection_manager.get_request_chunks(
                    session_id, request_id
                )
                if chunk_response.get("chunks"):
                    result["streamResults"] = chunk_response.get("chunks")
            except Exception:
                pass

            return result

        except Exception as e:
            raise RequestError(f"Failed to get request status: {e}")

    def list_requests(
        self,
        session_id: str,
        status: str = "",
        tool_name: str = "",
        limit: int = 10,
        offset: int = 0,
    ) -> List[Dict[str, Any]]:
        """List requests in a session."""
        try:
            self.connection_manager.ensure_connected()

            payload = {
                "sessionId": session_id,
                "status": status,
                "toolName": tool_name,
                "limit": limit,
                "offset": offset,
            }
            response = self.connection_manager.list_requests(payload)
            requests = response.get("requests", [])
            return [self._normalize_request(entry) for entry in requests]

        except Exception as e:
            raise RequestError(f"Failed to list requests: {e}")

    def create_request(self, session_id: str, tool_name: str, input_data: str) -> str:
        """Create a new request."""
        try:
            self.connection_manager.ensure_connected()

            payload = {
                "sessionId": session_id,
                "toolName": tool_name,
                "input": input_data,
            }

            response = self.connection_manager.create_request(payload)

            if response.get("error"):
                raise RequestError(f"Failed to create request: {response.get('error')}")

            return response.get("id")

        except Exception as e:
            raise RequestError(f"Failed to create request: {e}")

    def cancel_request(self, session_id: str, request_id: str) -> bool:
        """Cancel a request."""
        try:
            self.connection_manager.ensure_connected()

            response = self.connection_manager.cancel_request(session_id, request_id)
            if response.get("error"):
                raise RequestError(f"Failed to cancel request: {response.get('error')}")

            return bool(response.get("success", False))

        except Exception as e:
            raise RequestError(f"Failed to cancel request: {e}")

    def shutdown(self):
        """Shutdown request manager."""
        self.stop_polling()
        self.executor.shutdown(wait=False)
