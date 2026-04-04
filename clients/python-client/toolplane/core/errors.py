"""Error handling for Toolplane client."""


class ToolplaneError(Exception):
    """Base exception for Toolplane client errors."""

    pass


class ConnectionError(ToolplaneError):
    """Error related to gRPC connection."""

    pass


class ToolError(ToolplaneError):
    """Error related to tool operations."""

    pass


class SessionError(ToolplaneError):
    """Error related to session operations."""

    pass


class MachineError(ToolplaneError):
    """Error related to machine operations."""

    pass


class RequestError(ToolplaneError):
    """Error related to request operations."""

    pass


class TaskError(ToolplaneError):
    """Error related to task operations."""

    pass
