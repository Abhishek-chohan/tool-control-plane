"""HTTP session context implementation."""

import asyncio
import json
import logging
import time
import uuid
from typing import Any, Callable, Dict, List, Optional

from toolplane.utils.schema import generate_schema_from_function

from ..core.errors import (
    ToolplaneError,
    ToolplaneInvalidArgumentError,
)
from .http_connection import HTTPConnectionManager
from .http_machine import HTTPMachineManager
from .http_request import HTTPRequestManager
from .http_session import HTTPSessionManager
from .http_tool import HTTPToolManager

logger = logging.getLogger(__name__)


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
            logger.debug(
                "Registered machine %s for session %s", self.machine_id, self.session_id
            )
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
            logger.debug(
                "Registered tool %s%s for session %s",
                name,
                " (streaming)" if stream else "",
                self.session_id,
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

    async def ainvoke(self, tool_name: str, **params) -> str:
        """Submit a tool invocation without blocking the caller.

        Returns the request ID once the gateway accepts the work; poll
        get_request_status (or await astream) for the outcome. The blocking
        submission runs in a worker thread, so this is safe to await from a
        running event loop.
        """
        try:
            return await asyncio.to_thread(
                self.request_manager.create_request,
                self.session_id,
                tool_name,
                json.dumps(params),
            )
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
                    chunk_data = chunk.get("result", {})
                    is_final = chunk_data.get("isFinal", False)
                    chunk_content = chunk_data.get("chunk", "")

                    if chunk_data.get("requestId"):
                        request_id = chunk_data.get("requestId")
                    if chunk_data.get("seq"):
                        last_seq = max(last_seq, int(chunk_data.get("seq", 0)))

                    callback(chunk_content, is_final)
                    all_chunks.append(chunk_content)

                    stream_error = chunk_data.get("error") or chunk.get("error")
                    if stream_error:
                        raise ToolplaneError(f"Streaming error: {stream_error}")

                    if is_final:
                        break

                return all_chunks

            except ToolplaneError:
                # The tool itself failed: re-executing it would duplicate the
                # side effect, so surface the failure.
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
            self._stream_via_polling(
                tool_name,
                callback,
                params,
                idempotency_key,
                request_id=request_id,
                skip=len(all_chunks),
                accumulate=all_chunks,
            )
        return all_chunks

    def _stream_via_polling(
        self,
        tool_name: str,
        callback: Callable,
        params: Dict,
        idempotency_key: str = "",
        request_id: Optional[str] = None,
        skip: int = 0,
        accumulate: Optional[List] = None,
    ):
        """Stream via polling fallback.

        When request_id is given, polls that request; otherwise invokes the
        tool (under idempotency_key when provided) and polls the result.
        skip suppresses the first skip chunks, which the caller already
        delivered. When accumulate is given, polled chunks append to it so
        callers keep the chunks they already delivered.
        """
        if request_id is None:
            request_id = self.tool_manager.execute_tool(
                self.session_id, tool_name, params, idempotency_key
            )

        all_chunks = accumulate if accumulate is not None else []
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

            time.sleep(0.5)

        return all_chunks

    async def astream(
        self, tool_name: str, callback: Callable[[Any, bool], None], **params
    ):
        """Awaitable stream: runs the blocking stream loop in a worker
        thread and resolves with the collected chunks."""
        return await asyncio.to_thread(self.stream, tool_name, callback, **params)

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
        return self.session_manager.bulk_delete_sessions(user_id, session_ids, filter)

    def get_session_stats(self, user_id: str) -> Dict[str, int]:
        """Get session statistics for a user."""
        return self.session_manager.get_session_stats(user_id)

    def invalidate_session(self, reason: str = "") -> bool:
        """Invalidate a session."""
        return self.session_manager.invalidate_session(self.session_id, reason)
