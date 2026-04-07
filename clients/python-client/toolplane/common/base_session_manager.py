"""Base session manager class shared between gRPC and HTTP clients."""

import threading
from abc import ABC, abstractmethod
from typing import Any, Dict, List, Optional, Set

try:
    from ..core.errors import SessionError
except ImportError:
    # Fallback for standalone testing
    class SessionError(Exception):
        """Session error for standalone testing."""

        pass


from .utils import (
    generate_session_id,
    sanitize_input,
    validate_description,
    validate_session_name,
)


class BaseSessionManager(ABC):
    """Base class for session management."""

    def __init__(self, connection_manager):
        """Initialize base session manager."""
        self.connection_manager = connection_manager
        self.sessions: Dict[str, Any] = {}
        self.sessions_lock = threading.RLock()
        self._owned_sessions: Set[str] = set()

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
        # Validate inputs
        if name and not validate_session_name(name):
            raise SessionError(f"Invalid session name: {name}")

        if description and not validate_description(description):
            raise SessionError(
                f"Session description too long: {len(description)} characters"
            )

        # Sanitize inputs
        if name:
            name = sanitize_input(name, max_length=100)
        if description:
            description = sanitize_input(description, max_length=500)
        if namespace:
            namespace = sanitize_input(namespace, max_length=100)

        # Generate session ID if not provided
        if not session_id:
            session_id = generate_session_id()

        try:
            self.connection_manager.ensure_connected()
            created_session_id = self._create_session_on_server(
                session_id=session_id,
                user_id=user_id,
                name=name,
                description=description,
                namespace=namespace,
                api_key=api_key,
            )

            # Store session info locally
            with self.sessions_lock:
                self.sessions[created_session_id] = {
                    "id": created_session_id,
                    "user_id": user_id,
                    "name": name,
                    "description": description,
                    "namespace": namespace,
                    "api_key": "",
                    "created_at": threading.current_thread().ident,  # Basic tracking
                }
                self._owned_sessions.add(created_session_id)

            return created_session_id

        except Exception as e:
            raise SessionError(f"Failed to create session: {e}")

    @abstractmethod
    def _create_session_on_server(
        self,
        session_id: str,
        user_id: Optional[str],
        name: Optional[str],
        description: Optional[str],
        namespace: Optional[str],
        api_key: Optional[str],
    ) -> str:
        """Create session on server (protocol-specific)."""
        pass

    def get_session(self, session_id: str) -> Optional[Dict[str, Any]]:
        """Get session information."""
        try:
            self.connection_manager.ensure_connected()

            # Try to get from local cache first
            with self.sessions_lock:
                local_session = self.sessions.get(session_id)
                if local_session:
                    return local_session

            # Get from server
            server_session = self._get_session_from_server(session_id)

            # Cache locally
            if server_session:
                with self.sessions_lock:
                    self.sessions[session_id] = server_session

            return server_session

        except Exception as e:
            raise SessionError(f"Failed to get session {session_id}: {e}")

    @abstractmethod
    def _get_session_from_server(self, session_id: str) -> Optional[Dict[str, Any]]:
        """Get session from server (protocol-specific)."""
        pass

    def delete_session(self, session_id: str):
        """Delete a session."""
        try:
            self.connection_manager.ensure_connected()

            # Delete from server
            self._delete_session_on_server(session_id)

            # Remove from local cache
            with self.sessions_lock:
                self.sessions.pop(session_id, None)
                self._owned_sessions.discard(session_id)

        except Exception as e:
            raise SessionError(f"Failed to delete session {session_id}: {e}")

    @abstractmethod
    def _delete_session_on_server(self, session_id: str):
        """Delete session on server (protocol-specific)."""
        pass

    def list_sessions(self) -> List[Dict[str, Any]]:
        """List all sessions."""
        try:
            self.connection_manager.ensure_connected()
            sessions = self._list_sessions_on_server()

            # Update local cache
            with self.sessions_lock:
                for session in sessions:
                    session_id = session.get("id")
                    if session_id:
                        self.sessions[session_id] = session

            return sessions

        except Exception as e:
            raise SessionError(f"Failed to list sessions: {e}")

    @abstractmethod
    def _list_sessions_on_server(self) -> List[Dict[str, Any]]:
        """List sessions on server (protocol-specific)."""
        pass

    def session_exists(self, session_id: str) -> bool:
        """Check if a session exists."""
        try:
            session = self.get_session(session_id)
            return session is not None
        except Exception:
            return False

    def get_session_info(self, session_id: str) -> Dict[str, Any]:
        """Get detailed session information."""
        session = self.get_session(session_id)
        if not session:
            raise SessionError(f"Session {session_id} not found")

        return {
            "id": session.get("id"),
            "name": session.get("name"),
            "description": session.get("description"),
            "namespace": session.get("namespace"),
            "user_id": session.get("user_id"),
            "created_at": session.get("created_at"),
            "status": session.get("status", "active"),
        }

    def update_session(
        self,
        session_id: str,
        name: Optional[str] = None,
        description: Optional[str] = None,
        namespace: Optional[str] = None,
    ):
        """Update session information."""
        # Validate inputs
        if name and not validate_session_name(name):
            raise SessionError(f"Invalid session name: {name}")

        if description and not validate_description(description):
            raise SessionError(
                f"Session description too long: {len(description)} characters"
            )

        # Sanitize inputs
        if name:
            name = sanitize_input(name, max_length=100)
        if description:
            description = sanitize_input(description, max_length=500)
        if namespace:
            namespace = sanitize_input(namespace, max_length=100)

        try:
            self.connection_manager.ensure_connected()

            # Update on server
            updated_session = self._update_session_on_server(
                session_id, name, description, namespace
            )

            # Update local cache
            with self.sessions_lock:
                if updated_session:
                    self.sessions[session_id] = updated_session
                elif session_id in self.sessions:
                    if name:
                        self.sessions[session_id]["name"] = name
                    if description:
                        self.sessions[session_id]["description"] = description
                    if namespace:
                        self.sessions[session_id]["namespace"] = namespace

            return updated_session

        except Exception as e:
            raise SessionError(f"Failed to update session {session_id}: {e}")

    @abstractmethod
    def _update_session_on_server(
        self,
        session_id: str,
        name: Optional[str],
        description: Optional[str],
        namespace: Optional[str],
    ) -> Dict[str, Any]:
        """Update session on server (protocol-specific)."""
        pass

    def cleanup_session(self, session_id: str):
        """Cleanup a session and its resources."""
        try:
            # Delete session only if this client created it
            owned = False
            with self.sessions_lock:
                owned = session_id in self._owned_sessions

            if owned:
                self.delete_session(session_id)

            # Remove from local cache
            with self.sessions_lock:
                self.sessions.pop(session_id, None)
                self._owned_sessions.discard(session_id)

        except Exception:
            # Log error but don't raise for cleanup
            pass

    def cleanup_all(self):
        """Cleanup all sessions."""
        with self.sessions_lock:
            session_ids = list(self.sessions.keys())
            for session_id in session_ids:
                self.cleanup_session(session_id)

    def get_session_count(self) -> int:
        """Get the number of active sessions."""
        with self.sessions_lock:
            return len(self.sessions)

    def get_session_stats(self) -> Dict[str, Any]:
        """Get session statistics."""
        with self.sessions_lock:
            sessions = list(self.sessions.values())

            stats = {
                "total_sessions": len(sessions),
                "session_ids": list(self.sessions.keys()),
                "sessions_by_namespace": {},
                "sessions_by_user": {},
            }

            for session in sessions:
                namespace = session.get("namespace", "default")
                user_id = session.get("user_id", "unknown")

                stats["sessions_by_namespace"][namespace] = (
                    stats["sessions_by_namespace"].get(namespace, 0) + 1
                )
                stats["sessions_by_user"][user_id] = (
                    stats["sessions_by_user"].get(user_id, 0) + 1
                )

            return stats
