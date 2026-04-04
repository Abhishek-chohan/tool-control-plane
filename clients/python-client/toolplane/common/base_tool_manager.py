"""Base tool manager class shared between gRPC and HTTP clients."""

from abc import ABC, abstractmethod
from threading import RLock
from typing import Any, Callable, Dict, List, Optional, Set

try:
    from toolplane.utils.schema import generate_schema_from_function
except ImportError:
    # Fallback for standalone testing
    def generate_schema_from_function(func):
        """Fallback schema generator for standalone testing."""
        return {"type": "object", "properties": {}}


try:
    from ..core.errors import ToolError
except ImportError:
    # Fallback for standalone testing
    class ToolError(Exception):
        """Tool error for standalone testing."""

        pass


from .constants import CACHE_TTL_SECONDS
from .utils import (
    cache_with_ttl,
    sanitize_input,
    validate_tool_name,
)


class BaseToolManager(ABC):
    """Base class for tool management."""

    def __init__(self, connection_manager):
        """Initialize base tool manager."""
        self.connection_manager = connection_manager
        self.tools: Dict[str, Dict[str, Callable]] = (
            {}
        )  # session_id -> {tool_name: func}
        self.tool_schemas: Dict[str, Dict[str, Dict]] = (
            {}
        )  # session_id -> {tool_name: schema}
        self.streaming_tools: Dict[str, Set[str]] = {}  # session_id -> {tool_names}
        self._tool_cache: Dict[str, tuple] = {}  # session_id -> (timestamp, tools)
        self._lock = RLock()

    def register_tool(
        self,
        session_id: str,
        machine_id: str,
        name: str,
        func: Callable,
        schema: Optional[Dict] = None,
        description: Optional[str] = None,
        stream: bool = False,
        tags: Optional[List[str]] = None,
    ):
        """Register a tool for a session."""
        # Validate inputs
        if not validate_tool_name(name):
            raise ToolError(f"Invalid tool name: {name}")

        if tags is None:
            tags = []

        # Sanitize description
        if description:
            description = sanitize_input(description, max_length=500)

        # Generate schema if not provided
        if schema is None:
            try:
                schema = generate_schema_from_function(func)
            except Exception as e:
                raise ToolError(f"Failed to generate schema for tool {name}: {e}")

        # Add description and tags to schema
        if description:
            schema["description"] = description
        schema["tags"] = tags

        # Store tool locally
        with self._lock:
            if session_id not in self.tools:
                self.tools[session_id] = {}
                self.tool_schemas[session_id] = {}
                self.streaming_tools[session_id] = set()

            self.tools[session_id][name] = func
            self.tool_schemas[session_id][name] = schema

            if stream:
                self.streaming_tools[session_id].add(name)

        # Register with server
        self._register_tool_with_server(session_id, machine_id, name, schema)

    @abstractmethod
    def _register_tool_with_server(
        self, session_id: str, machine_id: str, name: str, schema: Dict
    ):
        """Register tool with server (protocol-specific)."""
        pass

    def unregister_tool(self, session_id: str, name: str):
        """Unregister a tool from a session."""
        try:
            # Remove from local storage
            with self._lock:
                if session_id in self.tools:
                    self.tools[session_id].pop(name, None)
                    self.tool_schemas[session_id].pop(name, None)
                    self.streaming_tools[session_id].discard(name)

            # Remove from server
            self._unregister_tool_from_server(session_id, name)

        except Exception as e:
            raise ToolError(f"Failed to unregister tool {name}: {e}")

    @abstractmethod
    def _unregister_tool_from_server(self, session_id: str, name: str):
        """Unregister tool from server (protocol-specific)."""
        pass

    def get_tool(self, session_id: str, name: str) -> Optional[Callable]:
        """Get a tool function by name."""
        with self._lock:
            return self.tools.get(session_id, {}).get(name)

    def is_streaming_tool(self, session_id: str, name: str) -> bool:
        """Check if a tool is a streaming tool."""
        with self._lock:
            return name in self.streaming_tools.get(session_id, set())

    def get_session_tools(self, session_id: str) -> Dict[str, Callable]:
        """Get all tools for a session."""
        with self._lock:
            return self.tools.get(session_id, {}).copy()

    def get_tool_schema(self, session_id: str, name: str) -> Optional[Dict]:
        """Get tool schema by name."""
        with self._lock:
            return self.tool_schemas.get(session_id, {}).get(name)

    def list_session_tools(self, session_id: str) -> List[str]:
        """List all tool names for a session."""
        with self._lock:
            return list(self.tools.get(session_id, {}).keys())

    @cache_with_ttl(CACHE_TTL_SECONDS)
    def get_available_tools(self, session_id: str) -> Dict[str, Any]:
        """Get available tools from server (cached)."""
        try:
            self.connection_manager.ensure_connected()
            return self._get_available_tools_from_server(session_id)

        except Exception as e:
            raise ToolError(f"Failed to get available tools: {e}")

    @abstractmethod
    def _get_available_tools_from_server(self, session_id: str) -> Dict[str, Any]:
        """Get available tools from server (protocol-specific)."""
        pass

    def execute_tool(self, session_id: str, tool_name: str, params: Dict) -> str:
        """Execute a tool and return request ID."""
        try:
            # Validate tool name
            if not validate_tool_name(tool_name):
                raise ToolError(f"Invalid tool name: {tool_name}")

            self.connection_manager.ensure_connected()
            return self._execute_tool_on_server(session_id, tool_name, params)

        except Exception as e:
            raise ToolError(f"Failed to execute tool {tool_name}: {e}")

    @abstractmethod
    def _execute_tool_on_server(
        self, session_id: str, tool_name: str, params: Dict
    ) -> str:
        """Execute tool on server (protocol-specific)."""
        pass

    def stream_tool(self, session_id: str, tool_name: str, params: Dict):
        """Stream tool execution."""
        try:
            # Validate tool name
            if not validate_tool_name(tool_name):
                raise ToolError(f"Invalid tool name: {tool_name}")

            self.connection_manager.ensure_connected()
            yield from self._stream_tool_on_server(session_id, tool_name, params)

        except Exception as e:
            raise ToolError(f"Failed to stream tool {tool_name}: {e}")

    @abstractmethod
    def _stream_tool_on_server(self, session_id: str, tool_name: str, params: Dict):
        """Stream tool execution on server (protocol-specific)."""
        pass

    def cleanup_session_tools(self, session_id: str):
        """Cleanup all tools for a session."""
        with self._lock:
            if session_id in self.tools:
                tool_names = list(self.tools[session_id].keys())
                for tool_name in tool_names:
                    try:
                        self.unregister_tool(session_id, tool_name)
                    except Exception:
                        # Ignore cleanup errors
                        pass

                # Clear local storage
                self.tools.pop(session_id, None)
                self.tool_schemas.pop(session_id, None)
                self.streaming_tools.pop(session_id, None)

    def cleanup_all(self):
        """Cleanup all tools."""
        with self._lock:
            session_ids = list(self.tools.keys())
            for session_id in session_ids:
                self.cleanup_session_tools(session_id)

    def get_tool_stats(self, session_id: str) -> Dict[str, Any]:
        """Get tool statistics for a session."""
        with self._lock:
            session_tools = self.tools.get(session_id, {})
            streaming_tools = self.streaming_tools.get(session_id, set())

            return {
                "total_tools": len(session_tools),
                "streaming_tools": len(streaming_tools),
                "regular_tools": len(session_tools) - len(streaming_tools),
                "tool_names": list(session_tools.keys()),
                "streaming_tool_names": list(streaming_tools),
            }

    def reregister_session_tools(self, session_id: str, machine_id: Optional[str]):
        """Re-register all tools for a session after reconnecting."""
        if not machine_id:
            return

        with self._lock:
            schemas = list(self.tool_schemas.get(session_id, {}).items())

        for tool_name, schema in schemas:
            try:
                self._register_tool_with_server(
                    session_id, machine_id, tool_name, schema
                )
            except Exception as exc:
                if "ALREADY_EXISTS" in str(exc).upper():
                    continue
                print(
                    f"Warning: Failed to re-register tool {tool_name} for session {session_id}: {exc}"
                )

    def validate_tool_params(
        self, session_id: str, tool_name: str, params: Dict
    ) -> bool:
        """Validate tool parameters against schema."""
        schema = self.get_tool_schema(session_id, tool_name)
        if not schema:
            return True  # No schema to validate against

        try:
            # Basic validation - can be extended with jsonschema
            tool_schema = schema.get("schema", {})
            if "properties" in tool_schema:
                required_fields = tool_schema.get("required", [])
                for field in required_fields:
                    if field not in params:
                        return False

            return True

        except Exception:
            return False
