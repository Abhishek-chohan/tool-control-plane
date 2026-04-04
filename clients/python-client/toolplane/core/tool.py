"""Tool management for Toolplane gRPC client."""

import json
from typing import Any, Dict, List

import grpc

from toolplane.proto.service_pb2 import (
    DeleteToolRequest,
    ExecuteToolRequest,
    GetToolByIdRequest,
    GetToolByNameRequest,
    ListToolsRequest,
    RegisterToolRequest,
)

from ..common.base_tool_manager import BaseToolManager
from ..common.utils import parse_json_safe
from .connection import ConnectionManager
from .errors import ToolError


class ToolManager(BaseToolManager):
    """Manages tool registration and execution for gRPC client."""

    def __init__(self, connection_manager: ConnectionManager):
        """Initialize gRPC tool manager."""
        super().__init__(connection_manager)

    def _normalize_tool(self, tool: Any) -> Dict[str, Any]:
        try:
            schema = parse_json_safe(tool.schema)
        except Exception:
            schema = {}

        if not isinstance(schema, dict):
            schema = {}

        return {
            "id": tool.id,
            "name": tool.name,
            "description": tool.description,
            "schema": schema,
            "config": dict(tool.config),
            "created_at": tool.created_at,
            "last_ping_at": tool.last_ping_at,
            "session_id": tool.session_id,
            "tags": list(tool.tags),
        }

    def _register_tool_with_server(
        self, session_id: str, machine_id: str, name: str, schema: Dict
    ):
        """Register tool with server."""
        try:
            self.connection_manager.ensure_connected()

            request = RegisterToolRequest(
                session_id=session_id,
                machine_id=machine_id,
                name=name,
                description=schema.get("description", ""),
                schema=json.dumps(schema.get("schema", {})),
                tags=schema.get("tags", []),
            )

            self.connection_manager.tool_stub.RegisterTool(
                request, metadata=self.connection_manager.get_metadata()
            )

        except grpc.RpcError as rpc_error:
            self._handle_rpc_error(rpc_error)
            raise ToolError(f"Failed to register tool {name} with server: {rpc_error}")
        except Exception as e:
            raise ToolError(f"Failed to register tool {name} with server: {e}")

    def _unregister_tool_from_server(self, session_id: str, name: str):
        """Unregister tool from server."""
        try:
            self.connection_manager.ensure_connected()

            # Get tool ID first
            tool_request = GetToolByNameRequest(session_id=session_id, tool_name=name)

            tool_response = self.connection_manager.tool_stub.GetToolByName(
                tool_request, metadata=self.connection_manager.get_metadata()
            )

            # Delete the tool
            delete_request = DeleteToolRequest(
                session_id=session_id, tool_id=tool_response.tool.id
            )

            self.connection_manager.tool_stub.DeleteTool(
                delete_request, metadata=self.connection_manager.get_metadata()
            )

        except grpc.RpcError as rpc_error:
            self._handle_rpc_error(rpc_error)
            raise ToolError(
                f"Failed to unregister tool {name} from server: {rpc_error}"
            )
        except Exception as e:
            raise ToolError(f"Failed to unregister tool {name} from server: {e}")

    def _get_available_tools_from_server(self, session_id: str) -> Dict[str, Any]:
        """Get available tools from server."""
        try:
            self.connection_manager.ensure_connected()

            request = ListToolsRequest(session_id=session_id)
            response = self.connection_manager.tool_stub.ListTools(
                request, metadata=self.connection_manager.get_metadata()
            )

            return {"tools": [self._normalize_tool(tool) for tool in response.tools]}

        except grpc.RpcError as rpc_error:
            self._handle_rpc_error(rpc_error)
            raise ToolError(f"Failed to get available tools: {rpc_error}")
        except Exception as e:
            raise ToolError(f"Failed to get available tools: {e}")

    def list_tools(self, session_id: str) -> List[Dict[str, Any]]:
        """List tools for a session."""
        return self.get_available_tools(session_id).get("tools", [])

    def get_tool_by_id(self, session_id: str, tool_id: str) -> Dict[str, Any]:
        """Get a tool by ID."""
        try:
            self.connection_manager.ensure_connected()

            request = GetToolByIdRequest(session_id=session_id, tool_id=tool_id)
            response = self.connection_manager.tool_stub.GetToolById(
                request, metadata=self.connection_manager.get_metadata()
            )
            return self._normalize_tool(response.tool)

        except grpc.RpcError as rpc_error:
            self._handle_rpc_error(rpc_error)
            raise ToolError(f"Failed to get tool {tool_id}: {rpc_error}")
        except Exception as e:
            raise ToolError(f"Failed to get tool {tool_id}: {e}")

    def get_tool_by_name(self, session_id: str, tool_name: str) -> Dict[str, Any]:
        """Get a tool by name."""
        try:
            self.connection_manager.ensure_connected()

            request = GetToolByNameRequest(session_id=session_id, tool_name=tool_name)
            response = self.connection_manager.tool_stub.GetToolByName(
                request, metadata=self.connection_manager.get_metadata()
            )
            return self._normalize_tool(response.tool)

        except grpc.RpcError as rpc_error:
            self._handle_rpc_error(rpc_error)
            raise ToolError(f"Failed to get tool {tool_name}: {rpc_error}")
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

            request = DeleteToolRequest(session_id=session_id, tool_id=tool_id)
            response = self.connection_manager.tool_stub.DeleteTool(
                request, metadata=self.connection_manager.get_metadata()
            )

            if response.success:
                self.get_available_tools.cache_clear()
                with self._lock:
                    if tool_name and session_id in self.tools:
                        self.tools[session_id].pop(tool_name, None)
                        self.tool_schemas[session_id].pop(tool_name, None)
                        self.streaming_tools[session_id].discard(tool_name)

            return response.success

        except grpc.RpcError as rpc_error:
            self._handle_rpc_error(rpc_error)
            raise ToolError(f"Failed to delete tool {tool_id}: {rpc_error}")
        except Exception as e:
            raise ToolError(f"Failed to delete tool {tool_id}: {e}")

    def _execute_tool_on_server(
        self, session_id: str, tool_name: str, params: Dict
    ) -> str:
        """Execute tool on server."""
        try:
            self.connection_manager.ensure_connected()

            request = ExecuteToolRequest(
                session_id=session_id, tool_name=tool_name, input=json.dumps(params)
            )

            response = self.connection_manager.tool_stub.ExecuteTool(
                request, metadata=self.connection_manager.get_metadata()
            )

            if response.error:
                raise ToolError(f"Tool execution failed: {response.error}")

            return response.request_id

        except grpc.RpcError as rpc_error:
            self._handle_rpc_error(rpc_error)
            raise ToolError(f"Failed to execute tool {tool_name}: {rpc_error}")
        except Exception as e:
            raise ToolError(f"Failed to execute tool {tool_name}: {e}")

    def _stream_tool_on_server(self, session_id: str, tool_name: str, params: Dict):
        """Stream tool execution on server."""
        try:
            self.connection_manager.ensure_connected()

            request = ExecuteToolRequest(
                session_id=session_id, tool_name=tool_name, input=json.dumps(params)
            )

            # Use streaming endpoint
            for chunk in self.connection_manager.tool_stub.StreamExecuteTool(
                request, metadata=self.connection_manager.get_metadata()
            ):
                yield chunk

        except grpc.RpcError as rpc_error:
            self._handle_rpc_error(rpc_error)
            raise ToolError(f"Failed to stream tool {tool_name}: {rpc_error}")
        except Exception as e:
            raise ToolError(f"Failed to stream tool {tool_name}: {e}")

    def _handle_rpc_error(self, rpc_error: grpc.RpcError):
        """Reset connection on recoverable gRPC errors."""
        if rpc_error.code() in (
            grpc.StatusCode.UNAVAILABLE,
            grpc.StatusCode.DEADLINE_EXCEEDED,
            grpc.StatusCode.INTERNAL,
            grpc.StatusCode.UNAUTHENTICATED,
        ):
            self.connection_manager.mark_unhealthy()
