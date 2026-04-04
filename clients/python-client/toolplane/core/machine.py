"""Machine management for Toolplane client."""

import threading
import time
from typing import Any, Dict, List, Optional

import grpc

from toolplane.interfaces.event_interface import Event, EventType
from toolplane.proto.service_pb2 import (
    DrainMachineRequest,
    GetMachineRequest,
    ListMachinesRequest,
    RegisterMachineRequest,
    UnregisterMachineRequest,
    UpdateMachinePingRequest,
)

from .connection import ConnectionManager
from .errors import MachineError


class MachineManager:
    """Manages machine registration and heartbeats."""

    def __init__(
        self,
        connection_manager: ConnectionManager,
        event_emitter: Optional[Any] = None,
        tool_manager=None,
        session_manager=None,
    ):
        """Initialize machine manager."""
        self.connection_manager = connection_manager
        self.machines: Dict[str, str] = {}  # session_id -> machine_id
        self.machines_lock = threading.RLock()
        self._last_heartbeat: Dict[str, float] = {}
        self._heartbeat_thread: Optional[threading.Thread] = None
        self._running = False
        self._tool_manager = tool_manager
        self._session_manager = session_manager
        self._last_recovery_attempt: Dict[str, float] = {}
        self._event_emitter = event_emitter

    def _emit_event(self, event_type: EventType, data: Dict[str, Any]) -> None:
        if not self._event_emitter:
            return
        try:
            self._event_emitter.emit(
                Event(type=event_type, source="machine_manager", data=data)
            )
        except Exception:
            # Event emission should not break core flows
            pass

    def attach_tool_manager(self, tool_manager):
        """Attach tool manager to support re-registration after recovery."""
        self._tool_manager = tool_manager

    def attach_session_manager(self, session_manager):
        """Attach session manager to update session context on recovery."""
        self._session_manager = session_manager

    def _normalize_machine(self, machine: Any) -> Dict[str, Any]:
        return {
            "id": machine.id,
            "session_id": machine.session_id,
            "sdk_version": machine.sdk_version,
            "sdk_language": machine.sdk_language,
            "ip": machine.ip,
            "created_at": machine.created_at,
            "last_ping_at": getattr(machine, "last_ping_at", ""),
        }

    def _clear_local_machine(self, session_id: str, machine_id: str) -> None:
        with self.machines_lock:
            if self.machines.get(session_id) == machine_id:
                self.machines.pop(session_id, None)
                self._last_heartbeat.pop(session_id, None)

    def list_machines(self, session_id: str) -> List[Dict[str, Any]]:
        """List machines for a session."""
        self.connection_manager.ensure_connected()

        try:
            request = ListMachinesRequest(session_id=session_id)
            response = self.connection_manager.machine_stub.ListMachines(
                request, metadata=self.connection_manager.get_metadata()
            )
            return [self._normalize_machine(machine) for machine in response.machines]
        except grpc.RpcError as rpc_error:
            self._handle_rpc_error(rpc_error)
            raise MachineError(
                f"Failed to list machines for session {session_id}: {rpc_error}"
            )
        except Exception as e:
            raise MachineError(f"Failed to list machines for session {session_id}: {e}")

    def get_machine(self, session_id: str, machine_id: str) -> Dict[str, Any]:
        """Get a specific machine by ID."""
        self.connection_manager.ensure_connected()

        try:
            request = GetMachineRequest(session_id=session_id, machine_id=machine_id)
            response = self.connection_manager.machine_stub.GetMachine(
                request, metadata=self.connection_manager.get_metadata()
            )
            return self._normalize_machine(response)
        except grpc.RpcError as rpc_error:
            self._handle_rpc_error(rpc_error)
            raise MachineError(
                f"Failed to get machine {machine_id} for session {session_id}: {rpc_error}"
            )
        except Exception as e:
            raise MachineError(
                f"Failed to get machine {machine_id} for session {session_id}: {e}"
            )

    def register_machine(self, session_id: str) -> str:
        """Register a machine for a session."""
        self.connection_manager.ensure_connected()

        self._emit_event(
            EventType.MACHINE_REGISTERED,
            {
                "session_id": session_id,
                "phase": "start",
            },
        )

        try:
            request = RegisterMachineRequest(
                session_id=session_id,
                machine_id="",  # Server generates ID
                sdk_version="1.0.0",
                sdk_language="python",
            )

            response = self.connection_manager.machine_stub.RegisterMachine(
                request, metadata=self.connection_manager.get_metadata()
            )

            machine_id = response.id

            with self.machines_lock:
                self.machines[session_id] = machine_id
                self._last_heartbeat[session_id] = time.time()

            self._emit_event(
                EventType.MACHINE_REGISTERED,
                {
                    "session_id": session_id,
                    "machine_id": machine_id,
                    "phase": "success",
                },
            )

            return machine_id

        except grpc.RpcError as rpc_error:
            self._emit_event(
                EventType.MACHINE_REGISTERED,
                {
                    "session_id": session_id,
                    "phase": "error",
                    "error": str(rpc_error),
                },
            )
            self._handle_rpc_error(rpc_error)
            raise MachineError(
                f"Failed to register machine for session {session_id}: {rpc_error}"
            )
        except Exception as e:
            self._emit_event(
                EventType.MACHINE_REGISTERED,
                {
                    "session_id": session_id,
                    "phase": "error",
                    "error": str(e),
                },
            )
            raise MachineError(
                f"Failed to register machine for session {session_id}: {e}"
            )

    def unregister_machine(
        self,
        session_id: str,
        machine_id: Optional[str] = None,
        reason: str = "explicit",
    ) -> bool:
        """Unregister a machine for a session."""
        machine_id = machine_id or self.get_machine_id(session_id)
        if not machine_id:
            return True

        try:
            self._emit_event(
                EventType.MACHINE_UNREGISTERED,
                {
                    "session_id": session_id,
                    "machine_id": machine_id,
                    "phase": "start",
                    "reason": reason,
                },
            )

            # Drain machine first
            self.drain_machine(session_id, machine_id)

            # Unregister machine
            request = UnregisterMachineRequest(
                session_id=session_id,
                machine_id=machine_id,
            )

            self.connection_manager.machine_stub.UnregisterMachine(
                request, metadata=self.connection_manager.get_metadata()
            )

            self._clear_local_machine(session_id, machine_id)

            self._emit_event(
                EventType.MACHINE_UNREGISTERED,
                {
                    "session_id": session_id,
                    "machine_id": machine_id,
                    "phase": "success",
                    "reason": reason,
                },
            )

            return True

        except grpc.RpcError as rpc_error:
            self._emit_event(
                EventType.MACHINE_UNREGISTERED,
                {
                    "session_id": session_id,
                    "machine_id": machine_id,
                    "phase": "error",
                    "reason": reason,
                    "error": str(rpc_error),
                },
            )
            self._handle_rpc_error(rpc_error)
            raise MachineError(
                f"Failed to unregister machine for session {session_id}: {rpc_error}"
            )
        except Exception as e:
            self._emit_event(
                EventType.MACHINE_UNREGISTERED,
                {
                    "session_id": session_id,
                    "machine_id": machine_id,
                    "phase": "error",
                    "reason": reason,
                    "error": str(e),
                },
            )
            raise MachineError(
                f"Failed to unregister machine for session {session_id}: {e}"
            )

    def drain_machine(self, session_id: str, machine_id: Optional[str] = None) -> bool:
        """Drain a machine for a session."""
        machine_id = machine_id or self.get_machine_id(session_id)
        if not machine_id:
            return True

        try:
            request = DrainMachineRequest(
                session_id=session_id,
                machine_id=machine_id,
            )

            self.connection_manager.machine_stub.DrainMachine(
                request, metadata=self.connection_manager.get_metadata()
            )

            self._clear_local_machine(session_id, machine_id)

            return True

        except grpc.RpcError as rpc_error:
            self._handle_rpc_error(rpc_error)
            raise MachineError(
                f"Failed to drain machine for session {session_id}: {rpc_error}"
            )
        except Exception as e:
            raise MachineError(f"Failed to drain machine for session {session_id}: {e}")

    def get_machine_id(self, session_id: str) -> Optional[str]:
        """Get machine ID for a session."""
        with self.machines_lock:
            return self.machines.get(session_id)

    def start_heartbeat(self, heartbeat_interval: int = 60):
        """Start heartbeat thread."""
        if self._running:
            return

        self._running = True
        self._heartbeat_thread = threading.Thread(
            target=self._heartbeat_loop, args=(heartbeat_interval,), daemon=True
        )
        self._heartbeat_thread.start()

    def stop_heartbeat(self):
        """Stop heartbeat thread."""
        self._running = False
        if self._heartbeat_thread:
            self._heartbeat_thread.join(timeout=1)

    def _heartbeat_loop(self, interval: int):
        """Heartbeat loop for all machines."""
        while self._running:
            sessions_to_recover: List[str] = []

            try:
                self.connection_manager.ensure_connected()
            except Exception:
                time.sleep(min(interval, 5))
                continue

            now = time.time()

            with self.machines_lock:
                due_machines = [
                    (session_id, machine_id)
                    for session_id, machine_id in self.machines.items()
                    if now - self._last_heartbeat.get(session_id, 0) >= interval
                ]

            for session_id, machine_id in due_machines:
                try:
                    request = UpdateMachinePingRequest(
                        session_id=session_id, machine_id=machine_id
                    )

                    self.connection_manager.machine_stub.UpdateMachinePing(
                        request, metadata=self.connection_manager.get_metadata()
                    )

                    with self.machines_lock:
                        self._last_heartbeat[session_id] = now

                except grpc.RpcError as rpc_error:
                    code = rpc_error.code()
                    if code in (
                        grpc.StatusCode.NOT_FOUND,
                        grpc.StatusCode.FAILED_PRECONDITION,
                    ):
                        self._emit_event(
                            EventType.MACHINE_HEARTBEAT_FAILED,
                            {
                                "session_id": session_id,
                                "machine_id": machine_id,
                                "status_code": code.name,
                                "error": str(rpc_error),
                            },
                        )
                        last_attempt = self._last_recovery_attempt.get(session_id, 0)
                        if now - last_attempt >= interval:
                            self._last_recovery_attempt[session_id] = now
                            sessions_to_recover.append(session_id)
                        with self.machines_lock:
                            self._last_heartbeat[session_id] = now
                    elif code in (
                        grpc.StatusCode.UNAVAILABLE,
                        grpc.StatusCode.DEADLINE_EXCEEDED,
                        grpc.StatusCode.INTERNAL,
                        grpc.StatusCode.UNAUTHENTICATED,
                    ):
                        self.connection_manager.mark_unhealthy()
                        with self.machines_lock:
                            self._last_heartbeat[session_id] = now
                    else:
                        print(f"Heartbeat error for session {session_id}: {rpc_error}")
                        with self.machines_lock:
                            self._last_heartbeat[session_id] = now

                except Exception as exc:
                    print(f"Unexpected heartbeat error for session {session_id}: {exc}")
                    with self.machines_lock:
                        self._last_heartbeat[session_id] = now

            for session_id in sessions_to_recover:
                self._recover_session(session_id)

            time.sleep(5)  # Check frequently, but send only per interval

    def cleanup_all(self):
        """Cleanup all machines."""
        self.stop_heartbeat()

        with self.machines_lock:
            self._last_recovery_attempt.clear()
            for session_id in list(self.machines.keys()):
                try:
                    self.unregister_machine(session_id, reason="cleanup_all")
                except Exception:
                    # Ignore cleanup errors
                    pass

    def _handle_rpc_error(self, rpc_error: grpc.RpcError):
        """Handle recoverable RPC errors by resetting the connection."""
        if rpc_error.code() in (
            grpc.StatusCode.UNAVAILABLE,
            grpc.StatusCode.DEADLINE_EXCEEDED,
            grpc.StatusCode.INTERNAL,
            grpc.StatusCode.UNAUTHENTICATED,
        ):
            self.connection_manager.mark_unhealthy()

    def _recover_session(self, session_id: str):
        """Attempt to recover machine and tool registrations for a session."""
        try:
            previous_machine_id = self.get_machine_id(session_id)
            self._emit_event(
                EventType.MACHINE_RECOVERY_STARTED,
                {
                    "session_id": session_id,
                    "previous_machine_id": previous_machine_id,
                },
            )

            new_machine_id = self.register_machine(session_id)

            if self._tool_manager:
                self._tool_manager.reregister_session_tools(session_id, new_machine_id)

            if self._session_manager:
                context = self._session_manager.get_session_context(session_id)
                if context:
                    context.machine_id = new_machine_id

            with self.machines_lock:
                self._last_recovery_attempt.pop(session_id, None)

            print(
                f"Recovered machine for session {session_id} with new machine ID {new_machine_id}"
            )

            self._emit_event(
                EventType.MACHINE_RECOVERY_SUCCEEDED,
                {
                    "session_id": session_id,
                    "machine_id": new_machine_id,
                },
            )

        except Exception as exc:
            self._emit_event(
                EventType.MACHINE_RECOVERY_FAILED,
                {
                    "session_id": session_id,
                    "error": str(exc),
                },
            )
            print(f"Failed to recover machine for session {session_id}: {exc}")
