"""HTTP session context implementation."""

import json
import time
from typing import Any, Callable, Dict, List, Optional

from toolplane.utils.schema import generate_schema_from_function

from ..core.errors import ToolplaneError
from .http_connection import HTTPConnectionManager
from .http_machine import HTTPMachineManager
from .http_request import HTTPRequestManager
from .http_session import HTTPSessionManager
from .http_tool import HTTPToolManager


class HTTPSessionContext:
    """
    Represents an HTTP session context with its own tools and machine registration.
    Encapsulates all session-specific state and operations.
    """

    def __init__(
        self,
        session_id: str,
        connection_manager: HTTPConnectionManager,
        machine_manager: HTTPMachineManager,
        tool_manager: HTTPToolManager,
        request_manager: HTTPRequestManager,
        session_manager: HTTPSessionManager,
    ):
        """Initialize HTTP session context."""
        self.session_id = session_id
        self.connection_manager = connection_manager
        self.machine_manager = machine_manager
        self.tool_manager = tool_manager
        self.request_manager = request_manager
        self.session_manager = session_manager

        self.machine_id: Optional[str] = None

    def register_machine(self) -> bool:
        """Register a machine for this session."""
        try:
            self.machine_id = self.machine_manager.register_machine(self.session_id)
            print(f"Registered machine {self.machine_id} for session {self.session_id}")
            return True
        except Exception as e:
            print(f"Error registering machine for session {self.session_id}: {e}")
            return False

    def register_tool(
        self,
        name: str,
        func: Callable,
        schema: Optional[Dict] = None,
        description: Optional[str] = None,
        stream: bool = False,
        tags: Optional[List[str]] = None,
    ):
        """Register a tool for this session."""
        if not self.machine_id:
            raise ToolplaneError(
                f"Session {self.session_id} has no machine registration. Use ProviderRuntime.attach_session(...) or register_machine() before registering tools."
            )

        try:
            self.tool_manager.register_tool(
                self.session_id,
                self.machine_id,
                name,
                func,
                schema,
                description,
                stream,
                tags,
            )
            print(
                f"Registered tool '{name}' for session {self.session_id}"
                + (" (streaming)" if stream else "")
            )
        except Exception as e:
            raise ToolplaneError(
                f"Failed to register tool {name} for session {self.session_id}: {e}"
            )

    def invoke(self, tool_name: str, **params) -> Any:
        """Invoke a tool in this session."""
        try:
            request_id = self.tool_manager.execute_tool(
                self.session_id, tool_name, params
            )

            # Poll for completion
            return self._wait_for_completion(request_id)

        except Exception as e:
            raise ToolplaneError(f"Failed to invoke tool {tool_name}: {e}")

    def ainvoke(self, tool_name: str, **params) -> str:
        """Invoke a tool asynchronously."""
        try:
            return self.request_manager.create_request(
                self.session_id, tool_name, json.dumps(params)
            )
        except Exception as e:
            raise ToolplaneError(f"Failed to async invoke tool {tool_name}: {e}")

    def stream(self, tool_name: str, callback: Callable[[Any, bool], None], **params):
        """Stream tool execution."""
        try:
            all_chunks = []

            # Try direct streaming first
            try:
                for chunk in self.tool_manager.stream_tool(
                    self.session_id, tool_name, params
                ):
                    chunk_data = chunk.get("result", {})
                    is_final = chunk_data.get("isFinal", False)
                    chunk_content = chunk_data.get("chunk", "")

                    callback(chunk_content, is_final)
                    all_chunks.append(chunk_content)

                    stream_error = chunk_data.get("error") or chunk.get("error")
                    if stream_error:
                        raise ToolplaneError(f"Streaming error: {stream_error}")

                    if is_final:
                        break

                return all_chunks

            except Exception:
                # Fall back to polling
                return self._stream_via_polling(tool_name, callback, params)

        except Exception as e:
            raise ToolplaneError(f"Failed to stream tool {tool_name}: {e}")

    def _stream_via_polling(self, tool_name: str, callback: Callable, params: Dict):
        """Stream via polling fallback."""
        request_id = self.ainvoke(tool_name, **params)

        all_chunks = []
        last_chunk_count = 0

        while True:
            status = self.get_request_status(request_id)

            if "streamResults" in status:
                chunks = status["streamResults"]

                # Process new chunks
                for i in range(last_chunk_count, len(chunks)):
                    callback(chunks[i], False)
                    all_chunks.append(chunks[i])

                last_chunk_count = len(chunks)

            if status["status"] == "done":
                callback("", True)
                break

            if status["status"] == "failure":
                raise ToolplaneError(
                    f"Streaming failed: {status.get('error', 'Unknown error')}"
                )

            time.sleep(0.5)

        return all_chunks

    def astream(self, tool_name: str, callback: Callable[[Any, bool], None], **params):
        """Stream results from a tool execution (alias)."""
        return self.stream(tool_name, callback, **params)

    def get_request_status(self, request_id: str) -> Dict[str, Any]:
        """Get request status."""
        return self.request_manager.get_request_status(self.session_id, request_id)

    def get_available_tools(self) -> Dict[str, Any]:
        """Get available tools for this session."""
        return self.tool_manager.get_available_tools(self.session_id)

    def list_tools(self) -> List[Dict[str, Any]]:
        """List tools for this session."""
        return self.tool_manager.list_tools(self.session_id)

    def get_tool_by_id(self, tool_id: str) -> Dict[str, Any]:
        """Get a tool by ID for this session."""
        return self.tool_manager.get_tool_by_id(self.session_id, tool_id)

    def get_tool_by_name(self, tool_name: str) -> Dict[str, Any]:
        """Get a tool by name for this session."""
        return self.tool_manager.get_tool_by_name(self.session_id, tool_name)

    def delete_tool(self, tool_id: str) -> bool:
        """Delete a tool by ID for this session."""
        return self.tool_manager.delete_tool(self.session_id, tool_id)

    def tool(self, name=None, description=None, stream=False, tags=None):
        """Decorator for registering tools."""
        if tags is None:
            tags = []

        def decorator(func):
            tool_name = name or func.__name__
            tool_schema = generate_schema_from_function(func)

            if description:
                tool_schema["description"] = description

            self.register_tool(tool_name, func, tool_schema, description, stream, tags)
            return func

        return decorator

    def _wait_for_completion(self, request_id: str, timeout: int = 60) -> Any:
        """Wait for request completion."""
        start_time = time.time()

        while time.time() - start_time < timeout:
            status = self.get_request_status(request_id)

            if status["status"] == "done":
                try:
                    return json.loads(status["result"])
                except (TypeError, ValueError, json.JSONDecodeError):
                    return status["result"]

            if status["status"] == "failure":
                raise ToolplaneError(
                    f"Tool execution failed: {status.get('error', 'Unknown error')}"
                )

            time.sleep(0.5)

        raise ToolplaneError("Tool execution timed out")

    def cleanup(self):
        """Cleanup this session."""
        try:
            # Cleanup tools
            self.tool_manager.cleanup_session_tools(self.session_id)

            # Unregister machine
            if self.machine_id:
                self.machine_manager.unregister_machine(
                    self.session_id, reason="session_context_cleanup"
                )

            # Remove from session manager
            self.session_manager.remove_session_context(self.session_id)

        except Exception as e:
            print(f"Error cleaning up session {self.session_id}: {e}")

    def poll_requests(self):
        """Poll for requests in this session."""
        if not self.machine_id:
            return

        try:
            tools = self.tool_manager.get_session_tools(self.session_id)
            streaming_tools = self.tool_manager.streaming_tools.get(
                self.session_id, set()
            )

            self.request_manager.poll_session_requests(
                self.session_id, self.machine_id, tools, streaming_tools
            )
        except Exception as e:
            print(f"Error polling requests for session {self.session_id}: {e}")

    # New methods for user session management
    def list_user_sessions(
        self, user_id: str, page_size: int = 10, page_token: int = 0, filter: str = ""
    ) -> Dict[str, Any]:
        """List user sessions with pagination and filtering."""
        return self.session_manager.list_user_sessions(
            user_id, page_size, page_token, filter
        )

    def bulk_delete_sessions(
        self, user_id: str, session_ids: Optional[List[str]] = None, filter: str = ""
    ) -> Dict[str, Any]:
        """Bulk delete sessions for a user."""
        return self.session_manager.bulk_delete_sessions(user_id, session_ids, filter)

    def get_session_stats(self, user_id: str) -> Dict[str, int]:
        """Get session statistics for a user."""
        return self.session_manager.get_session_stats(user_id)

    def refresh_session_token(self) -> Dict[str, str]:
        """Refresh session token."""
        return self.session_manager.refresh_session_token(self.session_id)

    def invalidate_session(self, reason: str = "") -> bool:
        """Invalidate a session."""
        return self.session_manager.invalidate_session(self.session_id, reason)
