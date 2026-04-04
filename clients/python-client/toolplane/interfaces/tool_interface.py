"""Tool management interface definitions."""

import inspect
from abc import ABC, abstractmethod
from dataclasses import dataclass
from enum import Enum
from typing import (
    Any,
    Callable,
    Dict,
    Iterator,
    List,
    Optional,
    Protocol,
    runtime_checkable,
)


class ToolExecutionStatus(Enum):
    """Tool execution status."""

    PENDING = "pending"
    RUNNING = "running"
    DONE = "done"
    FAILURE = "failure"
    COMPLETED = "done"
    FAILED = "failure"
    TIMEOUT = "timeout"
    CANCELLED = "cancelled"


@dataclass
class ToolDefinition:
    """Tool definition with metadata."""

    name: str
    func: Callable
    schema: Optional[Dict[str, Any]] = None
    description: Optional[str] = None
    stream: bool = False
    tags: List[str] = None
    timeout: Optional[int] = None

    def __post_init__(self):
        if self.tags is None:
            self.tags = []

        # Auto-generate description if not provided
        if not self.description and self.func.__doc__:
            self.description = self.func.__doc__.strip()

        # Auto-generate schema if not provided
        if not self.schema:
            self.schema = self._generate_schema()

    def _generate_schema(self) -> Dict[str, Any]:
        """Generate schema from function signature."""
        try:
            sig = inspect.signature(self.func)
            properties = {}
            required = []

            for param_name, param in sig.parameters.items():
                if param_name == "self":  # Skip self parameter
                    continue

                prop = {"type": "string"}  # Default type

                # Try to infer type from annotation
                if param.annotation != param.empty:
                    annotation = param.annotation
                    if annotation == int:
                        prop["type"] = "integer"
                    elif annotation == float:
                        prop["type"] = "number"
                    elif annotation == bool:
                        prop["type"] = "boolean"
                    elif annotation == list:
                        prop["type"] = "array"
                    elif annotation == dict:
                        prop["type"] = "object"

                # Handle optional parameters
                if param.default == param.empty:
                    required.append(param_name)
                else:
                    prop["default"] = param.default

                properties[param_name] = prop

            return {
                "type": "object",
                "properties": properties,
                "required": required,
            }
        except Exception:
            return {"type": "object", "properties": {}}


@dataclass
class ToolExecutionResult:
    """Result of tool execution."""

    tool_name: str
    session_id: str
    request_id: str
    status: ToolExecutionStatus
    result: Any = None
    error: Optional[str] = None
    execution_time: Optional[float] = None
    metadata: Dict[str, Any] = None

    def __post_init__(self):
        if self.metadata is None:
            self.metadata = {}


@runtime_checkable
class IToolExecutor(Protocol):
    """Protocol interface for tool execution."""

    def execute(
        self, tool_def: ToolDefinition, params: Dict[str, Any], context: Dict[str, Any]
    ) -> ToolExecutionResult:
        """Execute a tool synchronously."""
        ...

    def execute_async(
        self, tool_def: ToolDefinition, params: Dict[str, Any], context: Dict[str, Any]
    ) -> str:
        """Execute a tool asynchronously, return request ID."""
        ...

    def stream_execute(
        self, tool_def: ToolDefinition, params: Dict[str, Any], context: Dict[str, Any]
    ) -> Iterator[Any]:
        """Execute a tool with streaming results."""
        ...

    def cancel_execution(self, request_id: str) -> bool:
        """Cancel tool execution."""
        ...

    def get_execution_status(self, request_id: str) -> Optional[ToolExecutionResult]:
        """Get execution status."""
        ...


@runtime_checkable
class IToolManager(Protocol):
    """Protocol interface for tool management."""

    def register_tool(
        self,
        session_id: str,
        name: str,
        func: Callable,
        schema: Optional[Dict] = None,
        description: Optional[str] = None,
        stream: bool = False,
        tags: Optional[List[str]] = None,
    ) -> None:
        """Register a tool for a session."""
        ...

    def unregister_tool(self, session_id: str, name: str) -> bool:
        """Unregister a tool from a session."""
        ...

    def get_tool(self, session_id: str, name: str) -> Optional[ToolDefinition]:
        """Get tool definition."""
        ...

    def list_tools(self, session_id: str) -> List[ToolDefinition]:
        """List tools for a session."""
        ...

    def execute_tool(
        self, session_id: str, tool_name: str, params: Dict[str, Any]
    ) -> str:
        """Execute a tool."""
        ...

    def stream_tool(
        self, session_id: str, tool_name: str, params: Dict[str, Any]
    ) -> Iterator[Any]:
        """Stream tool execution."""
        ...

    def get_session_tools(self, session_id: str) -> Dict[str, ToolDefinition]:
        """Get all tools for a session."""
        ...


class IToolValidator(ABC):
    """Abstract interface for tool validation."""

    @abstractmethod
    def validate_tool_definition(self, tool_def: ToolDefinition) -> List[str]:
        """Validate tool definition, return list of errors."""
        pass

    @abstractmethod
    def validate_tool_parameters(
        self, tool_def: ToolDefinition, params: Dict[str, Any]
    ) -> List[str]:
        """Validate tool parameters, return list of errors."""
        pass

    @abstractmethod
    def sanitize_parameters(
        self, tool_def: ToolDefinition, params: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Sanitize tool parameters."""
        pass


class DefaultToolValidator(IToolValidator):
    """Default implementation of tool validator."""

    def validate_tool_definition(self, tool_def: ToolDefinition) -> List[str]:
        """Validate tool definition."""
        errors = []

        if not tool_def.name:
            errors.append("Tool name is required")
        elif not isinstance(tool_def.name, str):
            errors.append("Tool name must be a string")
        elif not tool_def.name.replace("_", "").replace("-", "").isalnum():
            errors.append("Tool name must be alphanumeric with underscores/hyphens")

        if not callable(tool_def.func):
            errors.append("Tool function must be callable")

        if tool_def.schema and not isinstance(tool_def.schema, dict):
            errors.append("Tool schema must be a dictionary")

        if tool_def.description and len(tool_def.description) > 1000:
            errors.append("Tool description is too long (max 1000 characters)")

        if tool_def.tags and not isinstance(tool_def.tags, list):
            errors.append("Tool tags must be a list")

        return errors

    def validate_tool_parameters(
        self, tool_def: ToolDefinition, params: Dict[str, Any]
    ) -> List[str]:
        """Validate tool parameters against schema."""
        errors = []

        if not tool_def.schema:
            return errors  # No schema to validate against

        schema = tool_def.schema
        if "properties" not in schema:
            return errors

        properties = schema["properties"]
        required = schema.get("required", [])

        # Check required parameters
        for req_param in required:
            if req_param not in params:
                errors.append(f"Required parameter '{req_param}' is missing")

        # Validate parameter types
        for param_name, param_value in params.items():
            if param_name not in properties:
                errors.append(f"Unknown parameter '{param_name}'")
                continue

            prop = properties[param_name]
            expected_type = prop.get("type", "string")

            if not self._validate_type(param_value, expected_type):
                errors.append(
                    f"Parameter '{param_name}' has invalid type. "
                    f"Expected {expected_type}, got {type(param_value).__name__}"
                )

        return errors

    def sanitize_parameters(
        self, tool_def: ToolDefinition, params: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Sanitize tool parameters."""
        if not tool_def.schema or "properties" not in tool_def.schema:
            return params

        sanitized = {}
        properties = tool_def.schema["properties"]

        for param_name, param_value in params.items():
            if param_name not in properties:
                continue  # Skip unknown parameters

            prop = properties[param_name]
            expected_type = prop.get("type", "string")

            # Try to convert to expected type
            try:
                sanitized[param_name] = self._convert_type(param_value, expected_type)
            except (ValueError, TypeError):
                sanitized[param_name] = param_value  # Keep original if conversion fails

        return sanitized

    def _validate_type(self, value: Any, expected_type: str) -> bool:
        """Validate if value matches expected type."""
        type_map = {
            "string": str,
            "integer": int,
            "number": (int, float),
            "boolean": bool,
            "array": list,
            "object": dict,
        }

        expected_python_type = type_map.get(expected_type)
        if not expected_python_type:
            return True  # Unknown type, assume valid

        return isinstance(value, expected_python_type)

    def _convert_type(self, value: Any, expected_type: str) -> Any:
        """Convert value to expected type."""
        if expected_type == "string":
            return str(value)
        elif expected_type == "integer":
            return int(value)
        elif expected_type == "number":
            return float(value)
        elif expected_type == "boolean":
            if isinstance(value, str):
                return value.lower() in ("true", "1", "yes", "on")
            return bool(value)
        elif expected_type == "array":
            if isinstance(value, str):
                import json

                return json.loads(value)
            return list(value)
        elif expected_type == "object":
            if isinstance(value, str):
                import json

                return json.loads(value)
            return dict(value)

        return value


class ToolRegistry:
    """Registry for managing tool definitions with validation."""

    def __init__(self, validator: Optional[IToolValidator] = None):
        self._tools: Dict[str, Dict[str, ToolDefinition]] = (
            {}
        )  # session_id -> tool_name -> tool_def
        self._validator = validator or DefaultToolValidator()

    def register_tool(self, session_id: str, tool_def: ToolDefinition) -> None:
        """Register a tool definition."""
        # Validate tool definition
        errors = self._validator.validate_tool_definition(tool_def)
        if errors:
            raise ValueError(f"Invalid tool definition: {'; '.join(errors)}")

        if session_id not in self._tools:
            self._tools[session_id] = {}

        self._tools[session_id][tool_def.name] = tool_def

    def unregister_tool(self, session_id: str, tool_name: str) -> bool:
        """Unregister a tool."""
        if session_id in self._tools and tool_name in self._tools[session_id]:
            del self._tools[session_id][tool_name]
            return True
        return False

    def get_tool(self, session_id: str, tool_name: str) -> Optional[ToolDefinition]:
        """Get tool definition."""
        return self._tools.get(session_id, {}).get(tool_name)

    def list_tools(self, session_id: str) -> List[ToolDefinition]:
        """List tools for a session."""
        return list(self._tools.get(session_id, {}).values())

    def get_session_tools(self, session_id: str) -> Dict[str, ToolDefinition]:
        """Get all tools for a session."""
        return self._tools.get(session_id, {}).copy()

    def validate_execution_params(
        self, session_id: str, tool_name: str, params: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Validate and sanitize execution parameters."""
        tool_def = self.get_tool(session_id, tool_name)
        if not tool_def:
            raise ValueError(f"Tool '{tool_name}' not found in session '{session_id}'")

        # Validate parameters
        errors = self._validator.validate_tool_parameters(tool_def, params)
        if errors:
            raise ValueError(f"Invalid parameters: {'; '.join(errors)}")

        # Sanitize parameters
        return self._validator.sanitize_parameters(tool_def, params)

    def clear_session_tools(self, session_id: str) -> None:
        """Clear all tools for a session."""
        if session_id in self._tools:
            del self._tools[session_id]

    def get_tool_stats(self) -> Dict[str, Any]:
        """Get tool registry statistics."""
        total_tools = sum(len(tools) for tools in self._tools.values())
        sessions_with_tools = len([s for s in self._tools.values() if s])

        return {
            "total_sessions": len(self._tools),
            "sessions_with_tools": sessions_with_tools,
            "total_tools": total_tools,
            "tools_per_session": {
                session_id: len(tools) for session_id, tools in self._tools.items()
            },
        }
