"""HTTP session management for Toolplane client."""

import threading
from typing import Any, Dict, List, Optional, Set

from ..common.utils import format_error_message, validate_session_name
from ..core.errors import SessionError
from .http_connection import HTTPConnectionManager


class HTTPSessionManager:
    """Manages session lifecycle for HTTP client."""

    def __init__(self, connection_manager: HTTPConnectionManager):
        """Initialize HTTP session manager."""
        self.connection_manager = connection_manager
        self.sessions: Dict[str, "HTTPSessionContext"] = {}
        self.sessions_lock = threading.RLock()
        self._owned_sessions: Set[str] = set()

    def _normalize_session(self, session: Dict[str, Any]) -> Dict[str, Any]:
        created_by = session.get("createdBy", session.get("created_by", ""))
        return {
            "id": session.get("id", ""),
            "name": session.get("name", ""),
            "description": session.get("description", ""),
            "namespace": session.get("namespace", ""),
            "created_at": session.get("createdAt", session.get("created_at", "")),
            "created_by": created_by,
            "user_id": created_by,
            "api_key": session.get("apiKey", session.get("api_key", "")),
            "status": "active",
        }

    def _normalize_api_key(self, api_key: Dict[str, Any]) -> Dict[str, Any]:
        return {
            "id": api_key.get("id", ""),
            "name": api_key.get("name", ""),
            "key": api_key.get("key", ""),
            "session_id": api_key.get("sessionId", api_key.get("session_id", "")),
            "created_at": api_key.get("createdAt", api_key.get("created_at", "")),
            "created_by": api_key.get("createdBy", api_key.get("created_by", "")),
            "revoked_at": api_key.get("revokedAt", api_key.get("revoked_at", "")),
        }

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
        try:
            # Validate session name if provided
            if name and not validate_session_name(name):
                raise SessionError(f"Invalid session name: {name}")

            self.connection_manager.ensure_connected()

            payload = {
                "userId": user_id or "",
                "name": name or "",
                "description": description or "",
                "apiKey": api_key or "",
            }

            if namespace:
                payload["namespace"] = namespace

            if session_id:
                payload["sessionId"] = session_id

            response = self.connection_manager.create_session(payload)

            # Extract session ID from response
            created_session_id = response.get("session", {}).get("id") or response.get(
                "id"
            )
            if not created_session_id:
                raise SessionError("No session ID returned from server")

            with self.sessions_lock:
                self._owned_sessions.add(created_session_id)

            return created_session_id

        except Exception as e:
            error_msg = format_error_message(e, "Failed to create session")
            raise SessionError(error_msg)

    def get_session(self, session_id: str) -> Optional[Dict[str, Any]]:
        """Get session information from server."""
        try:
            self.connection_manager.ensure_connected()

            response = self.connection_manager.get_session(session_id)
            session = response.get("session", response)
            if not isinstance(session, dict):
                return None

            return self._normalize_session(session)

        except Exception:
            return None

    def update_session(
        self,
        session_id: str,
        name: Optional[str] = None,
        description: Optional[str] = None,
        namespace: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Update session information."""
        try:
            self.connection_manager.ensure_connected()

            payload = {
                "sessionId": session_id,
                "name": name or "",
                "description": description or "",
            }

            if namespace:
                payload["namespace"] = namespace

            response = self.connection_manager.update_session(payload)
            session = response.get("session", response)
            if isinstance(session, dict):
                return self._normalize_session(session)
            return self.get_session(session_id) or {}

        except Exception as e:
            error_msg = format_error_message(e, "Failed to update session")
            raise SessionError(error_msg)

    def delete_session(self, session_id: str) -> bool:
        """Delete a session."""
        try:
            self.connection_manager.ensure_connected()

            self.connection_manager.delete_session(session_id)
            with self.sessions_lock:
                self.sessions.pop(session_id, None)
                self._owned_sessions.discard(session_id)
            return True

        except Exception as e:
            raise SessionError(f"Failed to delete session: {e}")

    def list_sessions(self, user_id: Optional[str] = None) -> List[Dict]:
        """List sessions for a user."""
        try:
            self.connection_manager.ensure_connected()

            resolved_user_id = user_id or getattr(
                self.connection_manager.config, "user_id", ""
            )
            if not resolved_user_id:
                raise SessionError(
                    "user_id is required to list sessions. Pass user_id or set it in client config."
                )

            response = self.connection_manager.list_sessions(resolved_user_id)
            sessions = response.get("sessions", [])
            normalized_sessions = []
            for session in sessions:
                normalized_sessions.append(self._normalize_session(session))
            return normalized_sessions

        except Exception as e:
            raise SessionError(f"Failed to list sessions: {e}")

    # New methods for user session management
    def list_user_sessions(
        self, user_id: str, page_size: int = 10, page_token: int = 0, filter: str = ""
    ) -> Dict[str, Any]:
        """List user sessions with pagination and filtering."""
        try:
            self.connection_manager.ensure_connected()

            payload = {
                "userId": user_id,
                "pageSize": page_size,
                "pageToken": page_token,
                "filter": filter,
            }

            response = self.connection_manager.list_user_sessions(payload)
            return response

        except Exception as e:
            raise SessionError(f"Failed to list user sessions: {e}")

    def bulk_delete_sessions(
        self, user_id: str, session_ids: Optional[List[str]] = None, filter: str = ""
    ) -> Dict[str, Any]:
        """Bulk delete sessions for a user."""
        try:
            self.connection_manager.ensure_connected()

            payload = {
                "userId": user_id,
                "sessionIds": session_ids or [],
                "filter": filter,
            }

            response = self.connection_manager.bulk_delete_sessions(payload)
            return response

        except Exception as e:
            raise SessionError(f"Failed to bulk delete sessions: {e}")

    def get_session_stats(self, user_id: str) -> Dict[str, int]:
        """Get session statistics for a user."""
        try:
            self.connection_manager.ensure_connected()

            payload = {"userId": user_id}
            response = self.connection_manager.get_session_stats(payload)
            return response

        except Exception as e:
            raise SessionError(f"Failed to get session stats: {e}")

    def refresh_session_token(self, session_id: str) -> Dict[str, str]:
        """Refresh session token."""
        try:
            self.connection_manager.ensure_connected()

            payload = {"sessionId": session_id}
            response = self.connection_manager.refresh_session_token(payload)
            return response

        except Exception as e:
            raise SessionError(f"Failed to refresh session token: {e}")

    def invalidate_session(self, session_id: str, reason: str = "") -> bool:
        """Invalidate a session."""
        try:
            self.connection_manager.ensure_connected()

            payload = {"sessionId": session_id, "reason": reason}
            response = self.connection_manager.invalidate_session(payload)
            return response.get("success", False)

        except Exception as e:
            raise SessionError(f"Failed to invalidate session: {e}")

    def create_api_key(self, session_id: str, name: str) -> Dict[str, Any]:
        """Create a new API key for a session."""
        try:
            self.connection_manager.ensure_connected()

            payload = {"sessionId": session_id, "name": name}
            response = self.connection_manager.create_api_key(payload)
            return self._normalize_api_key(response)

        except Exception as e:
            raise SessionError(
                f"Failed to create API key for session {session_id}: {e}"
            )

    def list_api_keys(self, session_id: str) -> List[Dict[str, Any]]:
        """List active API keys for a session."""
        try:
            self.connection_manager.ensure_connected()

            response = self.connection_manager.list_api_keys({"sessionId": session_id})
            api_keys = response.get("apiKeys", response.get("api_keys", []))
            return [self._normalize_api_key(api_key) for api_key in api_keys]

        except Exception as e:
            raise SessionError(f"Failed to list API keys for session {session_id}: {e}")

    def revoke_api_key(self, session_id: str, key_id: str) -> bool:
        """Revoke an API key for a session."""
        try:
            self.connection_manager.ensure_connected()

            response = self.connection_manager.revoke_api_key(
                {"sessionId": session_id, "keyId": key_id}
            )
            return bool(response.get("success", False))

        except Exception as e:
            raise SessionError(
                f"Failed to revoke API key {key_id} for session {session_id}: {e}"
            )

    def register_session_context(self, session_id: str, context: "HTTPSessionContext"):
        """Register a session context."""
        with self.sessions_lock:
            self.sessions[session_id] = context

    def get_session_context(self, session_id: str) -> Optional["HTTPSessionContext"]:
        """Get session context."""
        with self.sessions_lock:
            return self.sessions.get(session_id)

    def list_session_contexts(self) -> List["HTTPSessionContext"]:
        """List all session contexts."""
        with self.sessions_lock:
            return list(self.sessions.values())

    def remove_session_context(self, session_id: str):
        """Remove session context."""
        with self.sessions_lock:
            self.sessions.pop(session_id, None)
            self._owned_sessions.discard(session_id)

    def cleanup_all(self):
        """Cleanup all sessions."""
        with self.sessions_lock:
            self.sessions.clear()
            self._owned_sessions.clear()


# Forward declaration for type hinting
class HTTPSessionContext:
    """HTTP session context placeholder for type hinting."""

    pass
