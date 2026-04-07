"""Session management for Toolplane gRPC client."""

import threading
from typing import Any, Dict, List, Optional

from toolplane.proto.service_pb2 import (
    BulkDeleteSessionsRequest,
    CreateApiKeyRequest,
    CreateSessionRequest,
    DeleteSessionRequest,
    GetSessionRequest,
    GetSessionStatsRequest,
    InvalidateSessionRequest,
    ListApiKeysRequest,
    ListSessionsRequest,
    ListUserSessionsRequest,
    RefreshSessionTokenRequest,
    RevokeApiKeyRequest,
    UpdateSessionRequest,
)

from ..common.base_session_manager import BaseSessionManager
from .connection import ConnectionManager
from .errors import SessionError


class SessionManager(BaseSessionManager):
    """Manages session lifecycle for gRPC client."""

    def __init__(self, connection_manager: ConnectionManager):
        """Initialize session manager."""
        super().__init__(connection_manager)
        self.session_contexts: Dict[str, Any] = {}
        self.session_contexts_lock = threading.RLock()

    def _normalize_session(self, session: Any) -> Dict[str, Any]:
        created_by = getattr(session, "created_by", "")
        return {
            "id": session.id,
            "name": session.name,
            "description": session.description,
            "namespace": session.namespace,
            "created_at": session.created_at,
            "created_by": created_by,
            "user_id": created_by,
            "api_key": "",
            "status": "active",
        }

    def _normalize_api_key(self, api_key: Any) -> Dict[str, Any]:
        return {
            "id": api_key.id,
            "name": api_key.name,
            "key": api_key.key,
            "key_preview": getattr(api_key, "key_preview", ""),
            "capabilities": list(getattr(api_key, "capabilities", [])),
            "session_id": api_key.session_id,
            "created_at": api_key.created_at,
            "created_by": api_key.created_by,
            "revoked_at": getattr(api_key, "revoked_at", ""),
        }

    def _create_session_on_server(
        self,
        session_id: str,
        user_id: Optional[str],
        name: Optional[str],
        description: Optional[str],
        namespace: Optional[str],
        api_key: Optional[str],
    ) -> str:
        """Create session on gRPC server."""
        request = CreateSessionRequest(
            user_id=user_id or "",
            name=name or "",
            description=description or "",
            session_id=session_id or "",
            namespace=namespace or "",
        )

        response = self.connection_manager.session_stub.CreateSession(
            request, metadata=self.connection_manager.get_metadata()
        )

        return response.session.id

    def _get_session_from_server(self, session_id: str) -> Optional[Dict[str, Any]]:
        """Get session from gRPC server."""
        request = GetSessionRequest(session_id=session_id)

        try:
            response = self.connection_manager.session_stub.GetSession(
                request, metadata=self.connection_manager.get_metadata()
            )

            return self._normalize_session(response)

        except Exception:
            return None

    def _delete_session_on_server(self, session_id: str):
        """Delete session on gRPC server."""
        request = DeleteSessionRequest(session_id=session_id)
        self.connection_manager.session_stub.DeleteSession(
            request, metadata=self.connection_manager.get_metadata()
        )

    def _list_sessions_on_server(self) -> List[Dict[str, Any]]:
        """List sessions on gRPC server."""
        user_id = getattr(self.connection_manager.config, "user_id", "") or ""
        request = ListSessionsRequest(user_id=user_id)

        response = self.connection_manager.session_stub.ListSessions(
            request, metadata=self.connection_manager.get_metadata()
        )

        sessions = []
        for session in response.sessions:
            sessions.append(self._normalize_session(session))

        return sessions

    def _update_session_on_server(
        self,
        session_id: str,
        name: Optional[str],
        description: Optional[str],
        namespace: Optional[str],
    ) -> Dict[str, Any]:
        """Update session on gRPC server."""
        request = UpdateSessionRequest(
            session_id=session_id,
            name=name or "",
            description=description or "",
            namespace=namespace or "",
        )

        response = self.connection_manager.session_stub.UpdateSession(
            request, metadata=self.connection_manager.get_metadata()
        )

        return self._normalize_session(response)

    def create_api_key(
        self,
        session_id: str,
        name: str,
        capabilities: Optional[List[str]] = None,
    ) -> Dict[str, Any]:
        """Create a new API key for a session."""
        try:
            self.connection_manager.ensure_connected()

            request = CreateApiKeyRequest(
                session_id=session_id,
                name=name,
                capabilities=capabilities or [],
            )
            response = self.connection_manager.session_stub.CreateApiKey(
                request, metadata=self.connection_manager.get_metadata()
            )
            return self._normalize_api_key(response)

        except Exception as e:
            raise SessionError(
                f"Failed to create API key for session {session_id}: {e}"
            )

    def list_api_keys(self, session_id: str) -> List[Dict[str, Any]]:
        """List active API keys for a session."""
        try:
            self.connection_manager.ensure_connected()

            request = ListApiKeysRequest(session_id=session_id)
            response = self.connection_manager.session_stub.ListApiKeys(
                request, metadata=self.connection_manager.get_metadata()
            )
            return [self._normalize_api_key(api_key) for api_key in response.api_keys]

        except Exception as e:
            raise SessionError(f"Failed to list API keys for session {session_id}: {e}")

    def revoke_api_key(self, session_id: str, key_id: str) -> bool:
        """Revoke an API key for a session."""
        try:
            self.connection_manager.ensure_connected()

            request = RevokeApiKeyRequest(session_id=session_id, key_id=key_id)
            response = self.connection_manager.session_stub.RevokeApiKey(
                request, metadata=self.connection_manager.get_metadata()
            )
            return response.success

        except Exception as e:
            raise SessionError(
                f"Failed to revoke API key {key_id} for session {session_id}: {e}"
            )

    # New methods for user session management
    def list_user_sessions(
        self, user_id: str, page_size: int = 10, page_token: int = 0, filter: str = ""
    ) -> Dict[str, Any]:
        """List user sessions with pagination and filtering."""
        try:
            self.connection_manager.ensure_connected()

            request = ListUserSessionsRequest(
                user_id=user_id,
                page_size=page_size,
                page_token=page_token,
                filter=filter,
            )

            response = self.connection_manager.session_stub.ListUserSessions(
                request, metadata=self.connection_manager.get_metadata()
            )

            sessions = []
            for session in response.sessions:
                sessions.append(self._normalize_session(session))

            return {
                "sessions": sessions,
                "total_count": response.total_count,
                "next_page_token": response.next_page_token,
            }

        except Exception as e:
            raise SessionError(f"Failed to list user sessions: {e}")

    def bulk_delete_sessions(
        self, user_id: str, session_ids: Optional[List[str]] = None, filter: str = ""
    ) -> Dict[str, Any]:
        """Bulk delete sessions for a user."""
        try:
            self.connection_manager.ensure_connected()

            request = BulkDeleteSessionsRequest(
                user_id=user_id, session_ids=session_ids or [], filter=filter
            )

            response = self.connection_manager.session_stub.BulkDeleteSessions(
                request, metadata=self.connection_manager.get_metadata()
            )

            return {
                "deleted_count": response.deleted_count,
                "failed_deletions": response.failed_deletions,
            }

        except Exception as e:
            raise SessionError(f"Failed to bulk delete sessions: {e}")

    def get_session_stats(self, user_id: str) -> Dict[str, int]:
        """Get session statistics for a user."""
        try:
            self.connection_manager.ensure_connected()

            request = GetSessionStatsRequest(user_id=user_id)

            response = self.connection_manager.session_stub.GetSessionStats(
                request, metadata=self.connection_manager.get_metadata()
            )

            return {
                "total_sessions": response.total_sessions,
                "active_sessions": response.active_sessions,
                "expired_sessions": response.expired_sessions,
            }

        except Exception as e:
            raise SessionError(f"Failed to get session stats: {e}")

    def refresh_session_token(self, session_id: str) -> Dict[str, str]:
        """Refresh session token."""
        try:
            self.connection_manager.ensure_connected()

            request = RefreshSessionTokenRequest(session_id=session_id)

            response = self.connection_manager.session_stub.RefreshSessionToken(
                request, metadata=self.connection_manager.get_metadata()
            )

            return {"new_token": response.new_token, "expires_at": response.expires_at}

        except Exception as e:
            raise SessionError(f"Failed to refresh session token: {e}")

    def invalidate_session(self, session_id: str, reason: str = "") -> bool:
        """Invalidate a session."""
        try:
            self.connection_manager.ensure_connected()

            request = InvalidateSessionRequest(session_id=session_id, reason=reason)

            response = self.connection_manager.session_stub.InvalidateSession(
                request, metadata=self.connection_manager.get_metadata()
            )

            return response.success

        except Exception as e:
            raise SessionError(f"Failed to invalidate session: {e}")

    def register_session_context(self, session_id: str, context: Any):
        """Register a session context."""
        with self.session_contexts_lock:
            self.session_contexts[session_id] = context

    def get_session_context(self, session_id: str) -> Optional[Any]:
        """Get session context."""
        with self.session_contexts_lock:
            return self.session_contexts.get(session_id)

    def list_session_contexts(self) -> List[Any]:
        """List all session contexts."""
        with self.session_contexts_lock:
            return list(self.session_contexts.values())

    def cleanup_session_context(self, session_id: str):
        """Cleanup session context."""
        with self.session_contexts_lock:
            context = self.session_contexts.pop(session_id, None)
            if context and hasattr(context, "cleanup"):
                context.cleanup()

    def cleanup_all(self):
        """Cleanup all sessions and contexts."""
        with self.session_contexts_lock:
            session_ids = list(self.session_contexts.keys())
            for session_id in session_ids:
                self.cleanup_session_context(session_id)

        super().cleanup_all()
