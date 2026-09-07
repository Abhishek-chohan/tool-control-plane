"""Core modules for Toolplane client."""

from .config import ClientConfig
from .connection import ConnectionManager
from .errors import (
    ConnectionError,
    MachineError,
    RequestError,
    SessionError,
    TaskError,
    ToolError,
    ToolplaneAlreadyExistsError,
    ToolplaneAPIError,
    ToolplaneCancelledError,
    ToolplaneError,
    ToolplaneFailedPreconditionError,
    ToolplaneInternalError,
    ToolplaneInvalidArgumentError,
    ToolplaneNotFoundError,
    ToolplanePermissionDeniedError,
    ToolplaneResourceExhaustedError,
    ToolplaneTimeoutError,
    ToolplaneUnauthenticatedError,
    ToolplaneUnavailableError,
    api_error_from_http_response,
    api_error_from_rpc_error,
)
from .machine import MachineManager
from .request import RequestManager
from .session import SessionManager
from .session_context import SessionContext
from .task import TaskManager
from .tool import ToolManager

__all__ = [
    "ConnectionManager",
    "MachineManager",
    "ToolManager",
    "RequestManager",
    "TaskManager",
    "SessionManager",
    "SessionContext",
    "ToolplaneError",
    "ConnectionError",
    "ToolError",
    "SessionError",
    "MachineError",
    "RequestError",
    "TaskError",
    "ClientConfig",
    "ToolplaneAPIError",
    "ToolplaneNotFoundError",
    "ToolplaneInvalidArgumentError",
    "ToolplaneFailedPreconditionError",
    "ToolplaneResourceExhaustedError",
    "ToolplaneUnauthenticatedError",
    "ToolplanePermissionDeniedError",
    "ToolplaneAlreadyExistsError",
    "ToolplaneTimeoutError",
    "ToolplaneUnavailableError",
    "ToolplaneCancelledError",
    "ToolplaneInternalError",
    "api_error_from_rpc_error",
    "api_error_from_http_response",
]
