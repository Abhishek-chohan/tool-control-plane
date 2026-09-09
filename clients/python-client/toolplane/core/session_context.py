"""Session context implementation."""

import logging
import uuid
from typing import Any, Callable, Dict, List, Optional

from toolplane.utils.schema import generate_schema_from_function

from .connection import ConnectionManager
from .errors import ToolplaneError, ToolplaneInvalidArgumentError
from .machine import MachineManager
from .request import RequestManager
from .session import SessionManager
from .tool import ToolManager

logger = logging.getLogger(__name__)


class SessionContext:
    """
    Represents a session context with its own tools and machine registration.
    Encapsulates all session-specific state and operations.
    """

    def __init__(
        self,
        session_id: str,
        connection_manager: ConnectionManager,
        machine_manager: MachineManager,
        tool_manager: ToolManager,
        request_manager: RequestManager,
        session_manager: SessionManager,
    ):
        """Initialize session context."""
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
            return True
        except Exception as e:
            logger.warning(
                "Error registering machine for session %s: %s", self.session_id, e
            )
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
            logger.debug("Registered tool '%s' for session %s", name, self.session_id)
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
            return self.tool_manager.execute_tool(self.session_id, tool_name, params)
        except Exception as e:
            raise ToolplaneError(f"Failed to async invoke tool {tool_name}: {e}")

    def stream(self, tool_name: str, callback: Callable[[Any, bool], None], **params):
        """Stream tool execution.

        The stream is resumable: if the direct stream fails mid-flight, the
        fallback resumes from the last received sequence number (or, when it
        died before the first chunk, re-invokes under the same idempotency
        key so the server returns the original request). A tool that
        completed with an error is never re-executed.
        """
        idempotency_key = uuid.uuid4().hex
        all_chunks = []
        request_id = None
        last_seq = 0

        try:
            try:
                for chunk in self.tool_manager.stream_tool(
                    self.session_id, tool_name, params, idempotency_key
                ):
                    if chunk.request_id:
                        request_id = chunk.request_id
                    if chunk.seq:
                        last_seq = max(last_seq, chunk.seq)
                    callback(chunk.chunk, chunk.is_final)
                    all_chunks.append(chunk.chunk)

                    if chunk.error:
                        raise ToolplaneError(f"Streaming error: {chunk.error}")

                    if chunk.is_final:
                        break

                return all_chunks

            except ToolplaneError:
                # The tool itself failed (or the stream completed with an
                # error marker): re-executing it would be a side-effect
                # duplication, so surface the failure.
                raise
            except Exception:
                # Transport failure mid-stream: resume from the last sequence
                # number the server acknowledges.
                if request_id:
                    return self._resume_stream(
                        request_id,
                        last_seq,
                        callback,
                        all_chunks,
                        tool_name,
                        params,
                        idempotency_key,
                    )
                # The stream died before any chunk arrived. The server may
                # already have created the request; re-invoking under the same
                # idempotency key returns that request instead of executing
                # the tool a second time.
                return self._stream_via_polling(
                    tool_name, callback, params, idempotency_key
                )

        except ToolplaneError:
            raise
        except Exception as e:
            raise ToolplaneError(f"Failed to stream tool {tool_name}: {e}")

    def _resume_stream(
        self,
        request_id: str,
        last_seq: int,
        callback: Callable,
        all_chunks: List,
        tool_name: str,
        params: Dict,
        idempotency_key: str,
    ):
        """Continue a broken stream from the last acknowledged sequence."""
        try:
            for chunk in self.request_manager.resume_stream(
                self.session_id, request_id, last_seq
            ):
                value = chunk["chunk"]
                if value not in ("", None):
                    callback(value, chunk["is_final"])
                    all_chunks.append(value)
                if chunk["error"]:
                    raise ToolplaneError(f"Streaming error: {chunk['error']}")
                if chunk["is_final"]:
                    return all_chunks
        except ToolplaneInvalidArgumentError:
            # The retained window moved past our position; the full result is
            # still fetchable by polling the original request.
            pass
        return self._stream_via_polling(
            tool_name,
            callback,
            params,
            idempotency_key,
            request_id=request_id,
            skip=len(all_chunks),
        )

    def _stream_via_polling(
        self,
        tool_name: str,
        callback: Callable,
        params: Dict,
        idempotency_key: str = "",
        request_id: Optional[str] = None,
        skip: int = 0,
    ):
        """Stream via polling fallback.

        When request_id is given, polls that request; otherwise invokes the
        tool (under idempotency_key when provided) and polls the result.
        skip suppresses the first skip chunks, which the caller already
        delivered.
        """
        if request_id is None:
            request_id = self.tool_manager.execute_tool(
                self.session_id, tool_name, params, idempotency_key
            )

        all_chunks = []
        last_chunk_count = skip

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

            import time

            time.sleep(0.5)

        return all_chunks

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

    def _wait_for_completion(
        self, request_id: str, timeout: int = 60
    ) -> Dict[str, Any]:
        """Wait for request completion and return the full response/status dict."""
        import time

        start_time = time.time()

        while time.time() - start_time < timeout:
            status = self.get_request_status(request_id)

            if status["status"] == "done":
                # Ensure result is JSON-parsed if possible (RequestManager already attempts this)
                return status

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
            if hasattr(self.session_manager, "cleanup_session_context"):
                self.session_manager.cleanup_session_context(self.session_id)

        except Exception as e:
            logger.warning("Error cleaning up session %s: %s", self.session_id, e)

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
            logger.warning(
                "Error polling requests for session %s: %s", self.session_id, e
            )

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
        return self.session_manager.bulk_delete_sessions(
            user_id, session_ids or [], filter
        )

    def get_session_stats(self, user_id: str) -> Dict[str, int]:
        """Get session statistics for a user."""
        return self.session_manager.get_session_stats(user_id)

    def invalidate_session(self, reason: str = "") -> bool:
        """Invalidate a session."""
        return self.session_manager.invalidate_session(self.session_id, reason)
