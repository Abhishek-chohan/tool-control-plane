"""Session management interface definitions."""

from abc import ABC, abstractmethod
from enum import Enum
from typing import Any, Callable, Dict, List, Optional, Protocol, runtime_checkable


class SessionState(Enum):
    """Session states."""

    CREATED = "created"
    ACTIVE = "active"
    SUSPENDED = "suspended"
    TERMINATED = "terminated"
    ERROR = "error"


@runtime_checkable
class ISessionContext(Protocol):
    """Protocol interface for session contexts."""

    @property
    def session_id(self) -> str:
        """Get session ID."""
        ...

    @property
    def machine_id(self) -> Optional[str]:
        """Get machine ID for this session."""
        ...

    @property
    def state(self) -> SessionState:
        """Get session state."""
        ...

    def register_machine(self) -> bool:
        """Register a machine for this session."""
        ...

    def register_tool(
        self,
        name: str,
        func: Callable,
        schema: Optional[Dict] = None,
        description: Optional[str] = None,
        stream: bool = False,
        tags: Optional[List[str]] = None,
    ) -> None:
        """Register a tool for this session."""
        ...

    def invoke(self, tool_name: str, **params) -> Any:
        """Invoke a tool in this session."""
        ...

    def ainvoke(self, tool_name: str, **params) -> str:
        """Invoke a tool asynchronously."""
        ...

    def stream(
        self, tool_name: str, callback: Callable[[Any, bool], None], **params
    ) -> List[Any]:
        """Stream tool execution."""
        ...

    def get_available_tools(self) -> Dict[str, Any]:
        """Get available tools for this session."""
        ...

    def get_request_status(self, request_id: str) -> Dict[str, Any]:
        """Get request status."""
        ...


@runtime_checkable
class ISessionManager(Protocol):
    """Protocol interface for session management."""

    def create_session(
        self,
        session_id: Optional[str] = None,
        user_id: Optional[str] = None,
        name: Optional[str] = None,
        description: Optional[str] = None,
        namespace: Optional[str] = None,
        api_key: Optional[str] = None,
    ) -> str:
        """Create a new session."""
        ...

    def get_session(self, session_id: str) -> Optional[Dict[str, Any]]:
        """Get session information."""
        ...

    def list_sessions(self) -> List[Dict[str, Any]]:
        """List all sessions."""
        ...

    def delete_session(self, session_id: str) -> bool:
        """Delete a session."""
        ...

    def register_session_context(
        self, session_id: str, context: ISessionContext
    ) -> None:
        """Register a session context."""
        ...

    def get_session_context(self, session_id: str) -> Optional[ISessionContext]:
        """Get session context by ID."""
        ...

    def list_session_contexts(self) -> List[ISessionContext]:
        """List all session contexts."""
        ...


class ISessionLifecycleHandler(ABC):
    """Abstract interface for session lifecycle handling."""

    @abstractmethod
    def on_session_created(self, session_id: str, context: Dict[str, Any]) -> None:
        """Handle session creation."""
        pass

    @abstractmethod
    def on_session_activated(self, session_id: str) -> None:
        """Handle session activation."""
        pass

    @abstractmethod
    def on_session_suspended(self, session_id: str, reason: str) -> None:
        """Handle session suspension."""
        pass

    @abstractmethod
    def on_session_terminated(self, session_id: str, reason: str) -> None:
        """Handle session termination."""
        pass

    @abstractmethod
    def on_session_error(self, session_id: str, error: Exception) -> None:
        """Handle session error."""
        pass


class DefaultSessionLifecycleHandler(ISessionLifecycleHandler):
    """Default implementation of session lifecycle handler."""

    def __init__(self, log_func: Optional[Callable[[str], None]] = None):
        self.log_func = log_func or print

    def on_session_created(self, session_id: str, context: Dict[str, Any]) -> None:
        """Handle session creation."""
        self.log_func(f"Session created: {session_id}")

    def on_session_activated(self, session_id: str) -> None:
        """Handle session activation."""
        self.log_func(f"Session activated: {session_id}")

    def on_session_suspended(self, session_id: str, reason: str) -> None:
        """Handle session suspension."""
        self.log_func(f"Session suspended: {session_id}, reason: {reason}")

    def on_session_terminated(self, session_id: str, reason: str) -> None:
        """Handle session termination."""
        self.log_func(f"Session terminated: {session_id}, reason: {reason}")

    def on_session_error(self, session_id: str, error: Exception) -> None:
        """Handle session error."""
        self.log_func(f"Session error: {session_id}, error: {error}")


class SessionRegistry:
    """Registry for managing session contexts and lifecycle handlers."""

    def __init__(self):
        self._contexts: Dict[str, ISessionContext] = {}
        self._handlers: List[ISessionLifecycleHandler] = []
        self._session_metadata: Dict[str, Dict[str, Any]] = {}

    def register_context(self, session_id: str, context: ISessionContext) -> None:
        """Register a session context."""
        self._contexts[session_id] = context
        self._session_metadata[session_id] = {
            "created_at": __import__("datetime").datetime.now(),
            "state": SessionState.CREATED,
        }

        # Notify handlers
        for handler in self._handlers:
            try:
                handler.on_session_created(
                    session_id, self._session_metadata[session_id]
                )
            except Exception as e:
                print(f"Error in session lifecycle handler: {e}")

    def unregister_context(self, session_id: str, reason: str = "manual") -> bool:
        """Unregister a session context."""
        if session_id not in self._contexts:
            return False

        # Notify handlers
        for handler in self._handlers:
            try:
                handler.on_session_terminated(session_id, reason)
            except Exception as e:
                print(f"Error in session lifecycle handler: {e}")

        del self._contexts[session_id]
        if session_id in self._session_metadata:
            del self._session_metadata[session_id]

        return True

    def get_context(self, session_id: str) -> Optional[ISessionContext]:
        """Get session context by ID."""
        return self._contexts.get(session_id)

    def list_contexts(self) -> List[ISessionContext]:
        """List all session contexts."""
        return list(self._contexts.values())

    def add_lifecycle_handler(self, handler: ISessionLifecycleHandler) -> None:
        """Add a session lifecycle handler."""
        self._handlers.append(handler)

    def remove_lifecycle_handler(self, handler: ISessionLifecycleHandler) -> bool:
        """Remove a session lifecycle handler."""
        if handler in self._handlers:
            self._handlers.remove(handler)
            return True
        return False

    def update_session_state(self, session_id: str, state: SessionState) -> None:
        """Update session state and notify handlers."""
        if session_id not in self._session_metadata:
            return

        old_state = self._session_metadata[session_id].get("state")
        self._session_metadata[session_id]["state"] = state

        # Notify handlers based on state change
        for handler in self._handlers:
            try:
                if state == SessionState.ACTIVE and old_state != SessionState.ACTIVE:
                    handler.on_session_activated(session_id)
                elif state == SessionState.SUSPENDED:
                    handler.on_session_suspended(session_id, "state_change")
                elif state == SessionState.ERROR:
                    handler.on_session_error(
                        session_id, Exception("Session state changed to ERROR")
                    )
            except Exception as e:
                print(f"Error in session lifecycle handler: {e}")

    def get_session_metadata(self, session_id: str) -> Optional[Dict[str, Any]]:
        """Get session metadata."""
        return self._session_metadata.get(session_id)

    def cleanup_expired_sessions(self, max_age_hours: int = 24) -> List[str]:
        """Clean up expired sessions."""
        import datetime

        expired_sessions = []
        cutoff_time = datetime.datetime.now() - datetime.timedelta(hours=max_age_hours)

        for session_id, metadata in self._session_metadata.items():
            created_at = metadata.get("created_at")
            if created_at and created_at < cutoff_time:
                expired_sessions.append(session_id)

        # Remove expired sessions
        for session_id in expired_sessions:
            self.unregister_context(session_id, "expired")

        return expired_sessions
