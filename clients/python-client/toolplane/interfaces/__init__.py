"""Interface definitions for Toolplane client components."""

from .client_interface import ClientProtocol, IClientFactory, IToolplaneClient
from .connection_interface import IConnectionManager, IConnectionStrategy
from .event_interface import IEventEmitter, IEventHandler
from .request_interface import IRequestManager, IRequestProcessor
from .session_interface import ISessionContext, ISessionManager
from .tool_interface import IToolExecutor, IToolManager

__all__ = [
    # Client interfaces
    "IToolplaneClient",
    "IClientFactory",
    "ClientProtocol",
    # Component interfaces
    "IConnectionManager",
    "IConnectionStrategy",
    "ISessionManager",
    "ISessionContext",
    "IToolManager",
    "IToolExecutor",
    "IRequestManager",
    "IRequestProcessor",
    # Event interfaces
    "IEventEmitter",
    "IEventHandler",
]
