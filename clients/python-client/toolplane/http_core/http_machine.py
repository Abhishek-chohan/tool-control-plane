"""HTTP machine management for Toolplane client."""

import threading
import time
import uuid
from typing import Any, Dict, List, Optional

from toolplane.interfaces.event_interface import Event, EventType

from ..common.constants import DEFAULT_HEARTBEAT_INTERVAL
from ..common.utils import format_error_message
from ..core.errors import MachineError
from .http_connection import HTTPConnectionManager


class HTTPMachineManager:
    """Manages machine registration and heartbeats for HTTP client."""

    def __init__(
        self,
        connection_manager: HTTPConnectionManager,
        event_emitter: Optional[Any] = None,
    ):
        """Initialize HTTP machine manager."""
        self.connection_manager = connection_manager
        self.machines: Dict[str, str] = {}  # session_id -> machine_id
        self.machines_lock = threading.RLock()
        self._last_heartbeat: Dict[str, float] = {}
        self._heartbeat_thread: Optional[threading.Thread] = None
        self._running = False
        self._event_emitter = event_emitter

    def _emit_event(self, event_type: EventType, data: Dict[str, Any]) -> None:
        if not self._event_emitter:
            return
        try:
            self._event_emitter.emit(
                Event(type=event_type, source="http_machine_manager", data=data)
            )
        except Exception:
            pass

    def _normalize_machine(self, machine: Dict[str, Any]) -> Dict[str, Any]:
        return {
            "id": machine.get("id", ""),
            "session_id": machine.get("sessionId", machine.get("session_id", "")),
            "sdk_version": machine.get("sdkVersion", machine.get("sdk_version", "")),
            "sdk_language": machine.get("sdkLanguage", machine.get("sdk_language", "")),
            "ip": machine.get("ip", ""),
            "created_at": machine.get("createdAt", machine.get("created_at", "")),
            "last_ping_at": machine.get("lastPingAt", machine.get("last_ping_at", "")),
        }

    def _clear_local_machine(self, session_id: str, machine_id: str) -> None:
        with self.machines_lock:
            if self.machines.get(session_id) == machine_id:
                self.machines.pop(session_id, None)
                self._last_heartbeat.pop(session_id, None)

    def list_machines(self, session_id: str) -> List[Dict[str, Any]]:
        """List machines for a session."""
        try:
            self.connection_manager.ensure_connected()
            response = self.connection_manager.list_machines(session_id)
            machines = response.get("machines", [])
            return [self._normalize_machine(machine) for machine in machines]
        except Exception as e:
            error_msg = format_error_message(
                e, f"Failed to list machines for session {session_id}"
            )
            raise MachineError(error_msg)

    def get_machine(self, session_id: str, machine_id: str) -> Dict[str, Any]:
        """Get a machine by ID."""
        try:
            self.connection_manager.ensure_connected()
            response = self.connection_manager.get_machine(session_id, machine_id)
            machine = response.get("machine", response)
            if not isinstance(machine, dict):
                raise MachineError(
                    f"Unexpected machine payload for session {session_id}: {machine}"
                )
            return self._normalize_machine(machine)
        except Exception as e:
            error_msg = format_error_message(
                e, f"Failed to get machine {machine_id} for session {session_id}"
            )
            raise MachineError(error_msg)

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
            machine_id = str(uuid.uuid4())

            payload = {
                "sessionId": session_id,
                "machineId": machine_id,
                "sdkVersion": "1.0.0",
                "sdkLanguage": "python-http",
                "tools": [],
            }

            response = self.connection_manager.register_machine(payload)
            actual_machine_id = response.get("id", machine_id)

            with self.machines_lock:
                self.machines[session_id] = actual_machine_id
                self._last_heartbeat[session_id] = time.time()

            self._emit_event(
                EventType.MACHINE_REGISTERED,
                {
                    "session_id": session_id,
                    "machine_id": actual_machine_id,
                    "phase": "success",
                },
            )

            return actual_machine_id

        except Exception as e:
            error_msg = format_error_message(
                e, f"Failed to register machine for session {session_id}"
            )
            self._emit_event(
                EventType.MACHINE_REGISTERED,
                {
                    "session_id": session_id,
                    "phase": "error",
                    "error": error_msg,
                },
            )
            raise MachineError(error_msg)

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
            self.connection_manager.unregister_machine(session_id, machine_id)

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

        except Exception as e:
            error_msg = format_error_message(
                e, f"Failed to unregister machine for session {session_id}"
            )
            self._emit_event(
                EventType.MACHINE_UNREGISTERED,
                {
                    "session_id": session_id,
                    "machine_id": machine_id,
                    "phase": "error",
                    "reason": reason,
                    "error": error_msg,
                },
            )
            raise MachineError(error_msg)

    def drain_machine(self, session_id: str, machine_id: Optional[str] = None) -> bool:
        """Drain a machine for a session."""
        machine_id = machine_id or self.get_machine_id(session_id)
        if not machine_id:
            return True

        try:
            response = self.connection_manager.drain_machine(session_id, machine_id)
            if response.get("drained", False):
                self._clear_local_machine(session_id, machine_id)
            return response.get("drained", False)

        except Exception as e:
            error_msg = format_error_message(
                e, f"Failed to drain machine for session {session_id}"
            )
            raise MachineError(error_msg)

    def get_machine_id(self, session_id: str) -> Optional[str]:
        """Get machine ID for a session."""
        with self.machines_lock:
            return self.machines.get(session_id)

    def start_heartbeat(self, heartbeat_interval: int = DEFAULT_HEARTBEAT_INTERVAL):
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
            try:
                now = time.time()

                with self.machines_lock:
                    for session_id, machine_id in self.machines.items():
                        if now - self._last_heartbeat.get(session_id, 0) < interval:
                            continue

                        try:
                            self.connection_manager.update_machine_ping(
                                session_id, machine_id
                            )
                            self._last_heartbeat[session_id] = now

                        except Exception:
                            self._emit_event(
                                EventType.MACHINE_HEARTBEAT_FAILED,
                                {
                                    "session_id": session_id,
                                    "machine_id": machine_id,
                                    "error": "heartbeat_failed",
                                },
                            )

            except Exception:
                self._emit_event(
                    EventType.MACHINE_HEARTBEAT_FAILED,
                    {
                        "session_id": "*",
                        "error": "heartbeat_loop_error",
                    },
                )

            time.sleep(5)  # Check frequently, but send only per interval

    def cleanup_all(self):
        """Cleanup all machines."""
        self.stop_heartbeat()

        with self.machines_lock:
            for session_id in list(self.machines.keys()):
                try:
                    self.unregister_machine(session_id, reason="cleanup_all")
                except Exception:
                    # Ignore cleanup errors
                    pass
