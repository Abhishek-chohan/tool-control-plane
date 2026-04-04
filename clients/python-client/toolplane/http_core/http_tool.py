"""HTTP tool management for Toolplane client."""

import json
import time
from typing import Any, Callable, Dict, List, Optional

from toolplane.utils.schema import generate_schema_from_function

from ..common.base_tool_manager import BaseToolManager
from ..common.utils import validate_tool_name
from ..core.errors import ToolError
from .http_connection import HTTPConnectionManager


class HTTPToolManager(BaseToolManager):
    """Manages tool registration and execution for HTTP client."""

    def __init__(self, connection_manager: HTTPConnectionManager):
        """Initialize HTTP tool manager."""
        super().__init__(connection_manager)

    def _normalize_tool(self, tool: Dict[str, Any]) -> Dict[str, Any]:
        schema = tool.get("schema", {})
        if isinstance(schema, str):
            try:
                schema = json.loads(schema)
            except Exception:
                schema = {}

        if not isinstance(schema, dict):
            schema = {}

        return {
            "id": tool.get("id", ""),
            "name": tool.get("name", ""),
            "description": tool.get("description", ""),
            "schema": schema,
            "config": tool.get("config", {}),
            "created_at": tool.get("createdAt", tool.get("created_at", "")),
            "last_ping_at": tool.get("lastPingAt", tool.get("last_ping_at", "")),
            "session_id": tool.get("sessionId", tool.get("session_id", "")),
            "tags": tool.get("tags", []),
        }

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
        if tags is None:
            tags = []

        # Validate tool name
        if not validate_tool_name(name):
            raise ToolError(f"Invalid tool name: {name}")

        # Generate schema if not provided
        if schema is None:
            schema = generate_schema_from_function(func)

        # Add description and tags to schema
        if description:
            schema["description"] = description
        schema["tags"] = tags

        # Store tool locally
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

    def _register_tool_with_server(
        self, session_id: str, machine_id: str, name: str, schema: Dict
    ):
        """Register tool with server."""
        try:
            self.connection_manager.ensure_connected()

            payload = {
                "sessionId": session_id,
                "machineId": machine_id,
                "name": name,
                "description": schema.get("description", ""),
                "schema": json.dumps(schema.get("schema", {})),
                "tags": schema.get("tags", []),
                "config": {},  # Default config
            }

            self.connection_manager.register_tool(payload)

        except Exception as e:
            raise ToolError(f"Failed to register tool {name} with server: {e}")

    def unregister_tool(self, session_id: str, name: str):
        """Unregister a tool from a session."""
        try:
            # Remove from local storage
            if session_id in self.tools:
                self.tools[session_id].pop(name, None)
                self.tool_schemas[session_id].pop(name, None)
                self.streaming_tools[session_id].discard(name)

            # Remove from server
            self._unregister_tool_from_server(session_id, name)

        except Exception as e:
            raise ToolError(f"Failed to unregister tool {name}: {e}")

    def _unregister_tool_from_server(self, session_id: str, name: str):
        """Unregister tool from server."""
        try:
            self.connection_manager.ensure_connected()

            # Get tool by name first
            tool_response = self.connection_manager.get_tool_by_name(session_id, name)

            if "tool" in tool_response:
                tool_id = tool_response["tool"].get("id")
                if tool_id:
                    self.connection_manager.delete_tool(session_id, tool_id)

        except Exception as e:
            raise ToolError(f"Failed to unregister tool {name} from server: {e}")

    def get_tool(self, session_id: str, name: str) -> Optional[Callable]:
        """Get a tool function by name."""
        return self.tools.get(session_id, {}).get(name)

    def is_streaming_tool(self, session_id: str, name: str) -> bool:
        """Check if a tool is a streaming tool."""
        return name in self.streaming_tools.get(session_id, set())

    def get_session_tools(self, session_id: str) -> Dict[str, Callable]:
        """Get all tools for a session."""
        return self.tools.get(session_id, {})

    def _get_available_tools_from_server(self, session_id: str) -> Dict[str, Any]:
        """Get available tools from server."""
        # Check cache first
        now = time.time()
        cached = self._tool_cache.get(session_id)
        if cached and now - cached[0] < 30:  # 30 second cache
            return {"tools": cached[1]}

        try:
            self.connection_manager.ensure_connected()

            response = self.connection_manager.list_tools(session_id)

            # Handle case where response might be a string or dict
            if isinstance(response, str):
                try:
                    response = json.loads(response)
                except (TypeError, ValueError, json.JSONDecodeError):
                    return {"tools": []}

            if not isinstance(response, dict):
                return {"tools": []}

            tools = []
            for tool in response.get("tools", []):
                try:
                    # Handle case where tool might be a string
                    if isinstance(tool, str):
                        try:
                            tool = json.loads(tool)
                        except (TypeError, ValueError, json.JSONDecodeError):
                            continue

                    if not isinstance(tool, dict):
                        continue

                    tools.append(self._normalize_tool(tool))
                except Exception:
                    continue

            # Cache result
            self._tool_cache[session_id] = (now, tools)
            return {"tools": tools}

        except Exception as e:
            raise ToolError(f"Failed to get available tools: {e}")

    def list_tools(self, session_id: str) -> List[Dict[str, Any]]:
        """List tools for a session."""
        return self.get_available_tools(session_id).get("tools", [])

    def get_tool_by_id(self, session_id: str, tool_id: str) -> Dict[str, Any]:
        """Get a tool by ID."""
        try:
            self.connection_manager.ensure_connected()
            response = self.connection_manager.get_tool_by_id(session_id, tool_id)
            payload = response.get("tool", response)
            if not isinstance(payload, dict):
                raise ToolError(f"Unexpected tool lookup payload for {tool_id}")
            return self._normalize_tool(payload)
        except Exception as e:
            raise ToolError(f"Failed to get tool {tool_id}: {e}")

    def get_tool_by_name(self, session_id: str, tool_name: str) -> Dict[str, Any]:
        """Get a tool by name."""
        try:
            self.connection_manager.ensure_connected()
            response = self.connection_manager.get_tool_by_name(session_id, tool_name)
            payload = response.get("tool", response)
            if not isinstance(payload, dict):
                raise ToolError(f"Unexpected tool lookup payload for {tool_name}")
            return self._normalize_tool(payload)
        except Exception as e:
            raise ToolError(f"Failed to get tool {tool_name}: {e}")

    def delete_tool(self, session_id: str, tool_id: str) -> bool:
        """Delete a tool by ID."""
        try:
            self.connection_manager.ensure_connected()

            tool_name = None
            try:
                tool = self.get_tool_by_id(session_id, tool_id)
                tool_name = tool.get("name")
            except Exception:
                tool_name = None

            response = self.connection_manager.delete_tool(session_id, tool_id)
            success = (
                bool(response.get("success", False))
                if isinstance(response, dict)
                else bool(response)
            )

            if success:
                self.get_available_tools.cache_clear()
                self._tool_cache.pop(session_id, None)
                with self._lock:
                    if tool_name and session_id in self.tools:
                        self.tools[session_id].pop(tool_name, None)
                        self.tool_schemas[session_id].pop(tool_name, None)
                        self.streaming_tools[session_id].discard(tool_name)

            return success
        except Exception as e:
            raise ToolError(f"Failed to delete tool {tool_id}: {e}")

    def _execute_tool_on_server(
        self, session_id: str, tool_name: str, params: Dict
    ) -> str:
        """Execute a tool and return request ID."""
        try:
            self.connection_manager.ensure_connected()

            response = self.connection_manager.execute_tool(
                session_id, tool_name, json.dumps(params)
            )

            if response.get("error"):
                raise ToolError(f"Tool execution failed: {response.get('error')}")

            return response.get("requestId")

        except Exception as e:
            raise ToolError(f"Failed to execute tool {tool_name}: {e}")

    def _stream_tool_on_server(self, session_id: str, tool_name: str, params: Dict):
        """Stream tool execution."""
        try:
            self.connection_manager.ensure_connected()

            response = self.connection_manager.stream_execute_tool(
                session_id, tool_name, json.dumps(params)
            )

            # Track buffer for backpressure
            self.connection_manager.current_buffer_size = 0

            # Parse newline-delimited JSON (NDJSON)
            for line in response.iter_lines(decode_unicode=True):
                if not line:
                    continue

                # Check buffer size and apply backpressure if needed
                line_size = len(line.encode("utf-8"))
                self.connection_manager.current_buffer_size += line_size

                if (
                    self.connection_manager.current_buffer_size
                    > self.connection_manager.config.max_buffer_size
                ):
                    # Apply backpressure by pausing before processing
                    time.sleep(0.05)

                try:
                    chunk = json.loads(line)
                    yield chunk

                    # Update buffer tracking as we process
                    self.connection_manager.current_buffer_size -= line_size

                    if chunk.get("isFinal"):
                        break

                except ValueError:
                    continue

        except Exception as e:
            raise ToolError(f"Failed to stream tool {tool_name}: {e}")

    def cleanup_session_tools(self, session_id: str):
        """Cleanup all tools for a session."""
        if session_id in self.tools:
            for tool_name in list(self.tools[session_id].keys()):
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
        for session_id in list(self.tools.keys()):
            self.cleanup_session_tools(session_id)
