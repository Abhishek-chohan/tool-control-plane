"""Explicit provider runtime for machine-backed execution."""

import logging
import threading
import time
from dataclasses import dataclass
from typing import Any, Callable, Iterable, List, Optional, Set

logger = logging.getLogger(__name__)

try:
    from .core.errors import ConnectionError, ToolplaneError
except ImportError:
    from core.errors import ConnectionError, ToolplaneError


@dataclass
class _PendingTool:
    """A tool registration deferred until the runtime starts."""

    session_id: str
    name: str
    func: Callable
    description: Optional[str] = None
    stream: bool = False
    tags: List[str] = None


class ProviderRuntime:
    """Owns the provider lifecycle for an Toolplane client instance."""

    def __init__(
        self,
        client: Any,
        session_ids: Optional[Iterable[str]] = None,
        poll_interval: Optional[float] = None,
        heartbeat_interval: Optional[int] = None,
    ):
        self.client = client
        self._registration_lock = threading.Lock()
        self._pending_registrations: List[_PendingTool] = []
        self._session_ids: Set[str] = set(session_ids or [])
        self._poll_interval = (
            poll_interval
            if poll_interval is not None
            else getattr(client.config, "poll_interval", 1.0)
        )
        self._heartbeat_interval = (
            heartbeat_interval
            if heartbeat_interval is not None
            else getattr(client.config, "heartbeat_interval", 60)
        )
        self._running = False
        self._main_thread: Optional[threading.Thread] = None
        self._lock = threading.RLock()

    @property
    def running(self) -> bool:
        return self._running

    def add_sessions(self, session_ids: Optional[Iterable[str]]) -> None:
        if not session_ids:
            return
        for session_id in session_ids:
            if session_id:
                self._session_ids.add(session_id)

    def managed_session_ids(self) -> List[str]:
        session_ids: List[str] = []
        seen: Set[str] = set()
        for session_id in list(self._session_ids) + list(
            getattr(self.client, "session_ids", [])
        ):
            if not session_id or session_id in seen:
                continue
            session_ids.append(session_id)
            seen.add(session_id)
        return session_ids

    def attach_session(self, session_id: str, register_machine: bool = True):
        if not session_id:
            raise ToolplaneError(
                "ProviderRuntime.attach_session requires a non-empty session_id"
            )

        if not self.client.connection_manager.connected:
            if not self.client.connect():
                raise ConnectionError("Failed to connect to server")

        self._session_ids.add(session_id)
        return self.client.ensure_session_context(
            session_id,
            create_if_missing=False,
            register_machine=register_machine,
        )

    def create_session(
        self,
        session_id: Optional[str] = None,
        user_id: Optional[str] = None,
        name: Optional[str] = None,
        description: Optional[str] = None,
        namespace: Optional[str] = None,
        register_machine: bool = True,
    ):
        context = self.client.create_session(
            session_id=session_id,
            user_id=user_id,
            name=name,
            description=description,
            namespace=namespace,
            register_machine=False,
        )
        self._session_ids.add(context.session_id)
        if register_machine and not getattr(context, "machine_id", None):
            context = self.attach_session(context.session_id, register_machine=True)
        return context

    def register_tool(
        self,
        session_id: str,
        name: str,
        func: Callable,
        schema: Optional[dict] = None,
        description: Optional[str] = None,
        stream: bool = False,
        tags: Optional[List[str]] = None,
    ) -> Callable:
        context = self.attach_session(session_id, register_machine=True)
        context.register_tool(
            name=name,
            func=func,
            schema=schema,
            description=description,
            stream=stream,
            tags=tags,
        )
        return func

    def tool(
        self,
        session_id: Optional[str] = None,
        name: Optional[str] = None,
        description: Optional[str] = None,
        stream: bool = False,
        tags: Optional[List[str]] = None,
    ):
        """Queue a tool for registration at runtime start.

        Decorating defers the (network) registration until
        start_in_background/run_forever/poll_once, so importing a module of
        decorated tools performs no I/O. session_id may be omitted when the
        runtime manages exactly one session: the tool binds to it at
        start (file-based tool modules use this to stay session-free).
        """

        def decorator(func: Callable) -> Callable:
            with self._registration_lock:
                self._pending_registrations.append(
                    _PendingTool(
                        session_id=session_id,
                        name=name or func.__name__,
                        func=func,
                        description=description,
                        stream=stream,
                        tags=tags or [],
                    )
                )
            return func

        return decorator

    def _resolve_default_session(self) -> str:
        """Resolve the session an unbound (session_id=None) deferred tool
        registers into: the runtime's single managed session. Ambiguous or
        empty setups fail loudly instead of guessing."""
        managed = self.managed_session_ids()
        if len(managed) == 1:
            return managed[0]
        raise ToolplaneError(
            "tool has no session_id and the runtime does not manage exactly "
            f"one session (managed: {managed or 'none'}); bind it explicitly "
            "or manage a single session"
        )

    def _apply_pending_registrations(self) -> None:
        """Attach sessions and register every deferred tool.

        A failed registration puts the item back at the head of the queue so
        a later start attempt retries it; tools are never silently lost.
        """
        while True:
            with self._registration_lock:
                if not self._pending_registrations:
                    return
                item = self._pending_registrations.pop(0)

            try:
                self.register_tool(
                    session_id=item.session_id or self._resolve_default_session(),
                    name=item.name,
                    func=item.func,
                    description=item.description,
                    stream=item.stream,
                    tags=item.tags,
                )
            except Exception:
                with self._registration_lock:
                    self._pending_registrations.insert(0, item)
                raise

    def poll_once(self) -> None:
        self._apply_pending_registrations()
        for session_id in self.managed_session_ids():
            context = self.client.ensure_session_context(
                session_id,
                create_if_missing=False,
                register_machine=False,
            )
            if getattr(context, "machine_id", None):
                context.poll_requests()

    def start_in_background(
        self, session_ids: Optional[Iterable[str]] = None
    ) -> "ProviderRuntime":
        # Queued @tool registrations may introduce the very sessions this
        # call is about to check for; apply them before validating.
        self._apply_pending_registrations()
        with self._lock:
            self.add_sessions(session_ids)
            if self._running:
                return self

            if not self.client.connection_manager.connected:
                if not self.client.connect():
                    raise ConnectionError("Failed to connect to server")
            else:
                self.client._initialize_sessions(register_machine=False)

            managed_session_ids = self.managed_session_ids()
            if not managed_session_ids:
                raise ToolplaneError(
                    "ProviderRuntime requires at least one attached or configured session"
                )

            self._apply_pending_registrations()

            for session_id in managed_session_ids:
                self.attach_session(session_id, register_machine=True)

            self.client.machine_manager.start_heartbeat(self._heartbeat_interval)
            # Keep claimed executions alive: the renewal loop extends each
            # in-flight lease so long-running tools are not reclaimed mid-run.
            self.client.request_manager.start_lease_renewal()
            self._running = True
            if hasattr(self.client, "running"):
                self.client.running = True
            self._main_thread = threading.Thread(
                target=self._main_loop,
                name="toolplane-provider-runtime",
                daemon=True,
            )
            self._main_thread.start()
            return self

    def run_forever(self, session_ids: Optional[Iterable[str]] = None) -> None:
        self.start_in_background(session_ids=session_ids)
        try:
            while self._running:
                time.sleep(1)
        except KeyboardInterrupt:
            self.stop()

    def stop(self) -> None:
        with self._lock:
            self._running = False
            if hasattr(self.client, "running"):
                self.client.running = False

        if self._main_thread:
            self._main_thread.join(timeout=1)
            self._main_thread = None

        self.client.request_manager.stop_lease_renewal()
        self.client.machine_manager.stop_heartbeat()

    def close(self) -> None:
        self.stop()

    def _main_loop(self) -> None:
        while self._running:
            try:
                self.poll_once()
            except Exception as exc:
                logger.warning("Error in provider runtime loop: %s", exc)
            time.sleep(self._poll_interval)

    def __enter__(self) -> "ProviderRuntime":
        return self.start_in_background()

    def __exit__(self, exc_type, exc_val, exc_tb) -> None:
        self.stop()
