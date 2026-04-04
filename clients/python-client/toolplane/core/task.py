"""Task management for Toolplane gRPC client."""

from typing import Any, Dict, List

import grpc

from toolplane.proto.service_pb2 import (
    CancelTaskRequest,
    CreateTaskRequest,
    GetTaskRequest,
    ListTasksRequest,
)

from .connection import ConnectionManager
from .errors import TaskError


class TaskManager:
    """Manages task lifecycle for the gRPC client."""

    def __init__(self, connection_manager: ConnectionManager):
        """Initialize task manager."""
        self.connection_manager = connection_manager

    def _normalize_task(self, task: Any) -> Dict[str, Any]:
        return {
            "id": task.id,
            "session_id": task.session_id,
            "tool_name": task.tool_name,
            "status": task.status,
            "input": task.input,
            "result": task.result,
            "result_type": task.result_type,
            "error": task.error,
            "created_at": task.created_at,
            "updated_at": task.updated_at,
            "completed_at": getattr(task, "completed_at", ""),
        }

    def create_task(
        self,
        session_id: str,
        tool_name: str,
        input_data: str,
    ) -> Dict[str, Any]:
        """Create a task for a session."""
        self.connection_manager.ensure_connected()

        try:
            request = CreateTaskRequest(
                session_id=session_id,
                tool_name=tool_name,
                input=input_data,
            )
            response = self.connection_manager.tasks_stub.CreateTask(
                request, metadata=self.connection_manager.get_metadata()
            )
            return self._normalize_task(response)
        except grpc.RpcError as rpc_error:
            raise TaskError(
                f"Failed to create task for session {session_id}: {rpc_error}"
            )
        except Exception as exc:
            raise TaskError(f"Failed to create task for session {session_id}: {exc}")

    def get_task(self, session_id: str, task_id: str) -> Dict[str, Any]:
        """Get a task for a session."""
        self.connection_manager.ensure_connected()

        try:
            request = GetTaskRequest(session_id=session_id, task_id=task_id)
            response = self.connection_manager.tasks_stub.GetTask(
                request, metadata=self.connection_manager.get_metadata()
            )
            return self._normalize_task(response)
        except grpc.RpcError as rpc_error:
            raise TaskError(
                f"Failed to get task {task_id} for session {session_id}: {rpc_error}"
            )
        except Exception as exc:
            raise TaskError(
                f"Failed to get task {task_id} for session {session_id}: {exc}"
            )

    def list_tasks(self, session_id: str) -> List[Dict[str, Any]]:
        """List tasks for a session."""
        self.connection_manager.ensure_connected()

        try:
            request = ListTasksRequest(session_id=session_id)
            response = self.connection_manager.tasks_stub.ListTasks(
                request, metadata=self.connection_manager.get_metadata()
            )
            return [self._normalize_task(task) for task in response.tasks]
        except grpc.RpcError as rpc_error:
            raise TaskError(
                f"Failed to list tasks for session {session_id}: {rpc_error}"
            )
        except Exception as exc:
            raise TaskError(f"Failed to list tasks for session {session_id}: {exc}")

    def cancel_task(self, session_id: str, task_id: str) -> bool:
        """Cancel a task for a session."""
        self.connection_manager.ensure_connected()

        try:
            request = CancelTaskRequest(session_id=session_id, task_id=task_id)
            response = self.connection_manager.tasks_stub.CancelTask(
                request, metadata=self.connection_manager.get_metadata()
            )
            return response.success
        except grpc.RpcError as rpc_error:
            raise TaskError(
                f"Failed to cancel task {task_id} for session {session_id}: {rpc_error}"
            )
        except Exception as exc:
            raise TaskError(
                f"Failed to cancel task {task_id} for session {session_id}: {exc}"
            )
