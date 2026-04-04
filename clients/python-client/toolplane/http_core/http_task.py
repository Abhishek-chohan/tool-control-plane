"""HTTP task management for Toolplane client."""

from typing import Any, Dict, List

from ..common.utils import format_error_message
from ..core.errors import TaskError
from .http_connection import HTTPConnectionManager


class HTTPTaskManager:
    """Manages task lifecycle for the HTTP client."""

    def __init__(self, connection_manager: HTTPConnectionManager):
        """Initialize HTTP task manager."""
        self.connection_manager = connection_manager

    def _normalize_task(self, task: Dict[str, Any]) -> Dict[str, Any]:
        return {
            "id": task.get("id", ""),
            "session_id": task.get("sessionId", task.get("session_id", "")),
            "tool_name": task.get("toolName", task.get("tool_name", "")),
            "status": task.get("status", ""),
            "input": task.get("input", ""),
            "result": task.get("result", ""),
            "result_type": task.get("resultType", task.get("result_type", "")),
            "error": task.get("error", ""),
            "created_at": task.get("createdAt", task.get("created_at", "")),
            "updated_at": task.get("updatedAt", task.get("updated_at", "")),
            "completed_at": task.get("completedAt", task.get("completed_at", "")),
        }

    def create_task(
        self,
        session_id: str,
        tool_name: str,
        input_data: str,
    ) -> Dict[str, Any]:
        """Create a task for a session."""
        try:
            self.connection_manager.ensure_connected()
            response = self.connection_manager.create_task(
                {
                    "sessionId": session_id,
                    "toolName": tool_name,
                    "input": input_data,
                }
            )
            task = response.get("task", response)
            if not isinstance(task, dict):
                raise TaskError(
                    f"Unexpected task payload for session {session_id}: {task}"
                )
            return self._normalize_task(task)
        except Exception as exc:
            raise TaskError(format_error_message(exc, "Failed to create task"))

    def get_task(self, session_id: str, task_id: str) -> Dict[str, Any]:
        """Get a task for a session."""
        try:
            self.connection_manager.ensure_connected()
            response = self.connection_manager.get_task(session_id, task_id)
            task = response.get("task", response)
            if not isinstance(task, dict):
                raise TaskError(
                    f"Unexpected task payload for session {session_id}: {task}"
                )
            return self._normalize_task(task)
        except Exception as exc:
            raise TaskError(format_error_message(exc, f"Failed to get task {task_id}"))

    def list_tasks(self, session_id: str) -> List[Dict[str, Any]]:
        """List tasks for a session."""
        try:
            self.connection_manager.ensure_connected()
            response = self.connection_manager.list_tasks(session_id)
            tasks = response.get("tasks", [])
            return [self._normalize_task(task) for task in tasks]
        except Exception as exc:
            raise TaskError(
                format_error_message(
                    exc, f"Failed to list tasks for session {session_id}"
                )
            )

    def cancel_task(self, session_id: str, task_id: str) -> bool:
        """Cancel a task for a session."""
        try:
            self.connection_manager.ensure_connected()
            response = self.connection_manager.cancel_task(session_id, task_id)
            return response.get("success", False)
        except Exception as exc:
            raise TaskError(
                format_error_message(exc, f"Failed to cancel task {task_id}")
            )
