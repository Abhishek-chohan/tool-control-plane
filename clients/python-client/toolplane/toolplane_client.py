"""Modular Toolplane client implementation."""

import threading
import time
from typing import Any, Callable, Dict, List, Optional

try:
    from .common import (
        DEFAULT_HEARTBEAT_INTERVAL,
        DEFAULT_MAX_WORKERS,
        DEFAULT_POLL_INTERVAL,
        DEFAULT_REQUEST_TIMEOUT,
        generate_session_id,
        validate_session_name,
    )
    from .core import (
        ClientConfig,
        ConnectionError,
        ConnectionManager,
        MachineManager,
        RequestManager,
        SessionContext,
        SessionManager,
        TaskManager,
        ToolManager,
        ToolplaneError,
    )
except ImportError:
    # Fallback for direct execution
    from common import (
        DEFAULT_HEARTBEAT_INTERVAL,
        DEFAULT_MAX_WORKERS,
        DEFAULT_POLL_INTERVAL,
        DEFAULT_REQUEST_TIMEOUT,
        generate_session_id,
        validate_session_name,
    )
    from core import (
        ClientConfig,
        ConnectionError,
        ConnectionManager,
        MachineManager,
        RequestManager,
        SessionContext,
        SessionManager,
        TaskManager,
        ToolManager,
        ToolplaneError,
    )


class Toolplane:
    """
    Modular Toolplane client for registering and executing tools.
    Supports multi-session mode with explicit session management.
    """

    def __init__(
        self,
        server_host: str = "localhost",
        server_port: int = 80,
        session_ids: Optional[List[str]] = None,
        api_key: Optional[str] = None,
        user_id: Optional[str] = None,
        session_name: Optional[str] = None,
        session_description: Optional[str] = None,
        session_namespace: Optional[str] = None,
        heartbeat_interval: int = DEFAULT_HEARTBEAT_INTERVAL,
        max_workers: int = DEFAULT_MAX_WORKERS,
        request_timeout: int = DEFAULT_REQUEST_TIMEOUT,
        poll_interval: float = DEFAULT_POLL_INTERVAL,
        # Retry configuration
        max_retries: int = 3,
        retry_base_delay: float = 1.0,
        retry_max_delay: float = 60.0,
        retry_backoff_factor: float = 2.0,
        event_emitter: Optional[Any] = None,
    ):
        """Initialize Toolplane client."""
        # Create configuration
        self.config = ClientConfig(
            server_host=server_host,
            server_port=server_port,
            api_key=api_key,
            user_id=user_id,
            session_name=session_name,
            session_description=session_description,
            session_namespace=session_namespace,
            heartbeat_interval=heartbeat_interval,
            max_workers=max_workers,
            request_timeout=request_timeout,
            poll_interval=poll_interval,
            # Pass retry configuration to config so ConnectionManager can access it
            max_retries=max_retries,
            retry_base_delay=retry_base_delay,
            retry_max_delay=retry_max_delay,
            retry_backoff_factor=retry_backoff_factor,
        )

        # Initialize managers
        self.connection_manager = ConnectionManager(self.config)
        self.machine_manager = MachineManager(self.connection_manager, event_emitter)
        self.tool_manager = ToolManager(self.connection_manager)
        self.request_manager = RequestManager(self.connection_manager, max_workers)
        self.task_manager = TaskManager(self.connection_manager)
        self.session_manager = SessionManager(self.connection_manager)

        # Wire managers that coordinate during recovery flows
        self.machine_manager.attach_tool_manager(self.tool_manager)
        self.machine_manager.attach_session_manager(self.session_manager)

        # Client state
        self.running = False
        self.session_ids = session_ids or []
        self._main_thread: Optional[threading.Thread] = None
        self._provider_runtime = None

    def connect(self) -> bool:
        """Connect to Toolplane server."""
        try:
            success = self.connection_manager.connect()
            if success:
                self._initialize_sessions(register_machine=False)
            return success
        except Exception as e:
            raise ConnectionError(f"Failed to connect: {e}")

    def disconnect(self) -> None:
        """Disconnect from Toolplane server."""
        self.stop()
        try:
            self.request_manager.stop_polling()
        except Exception:
            pass
        try:
            self.machine_manager.stop_heartbeat()
        except Exception:
            pass
        self.machine_manager.cleanup_all()
        self.tool_manager.cleanup_all()
        self.session_manager.cleanup_all()
        self.connection_manager.disconnect()

    def _initialize_sessions(self, register_machine: bool = False) -> None:
        """Initialize specified sessions."""
        for session_id in list(self.session_ids):
            try:
                self.ensure_session_context(
                    session_id,
                    create_if_missing=True,
                    register_machine=register_machine,
                )
            except Exception as e:
                print(f"Failed to initialize session {session_id}: {e}")

    def ensure_session_context(
        self,
        session_id: str,
        create_if_missing: bool = False,
        register_machine: bool = False,
    ) -> SessionContext:
        """Ensure a local session context exists without implying provider startup."""
        if not session_id or not isinstance(session_id, str):
            raise ToolplaneError(f"Invalid session ID: {session_id}")

        existing_context = self.session_manager.get_session_context(session_id)
        if existing_context:
            if register_machine and not getattr(existing_context, "machine_id", None):
                if not existing_context.register_machine():
                    raise ToolplaneError(
                        f"Failed to register machine for session {session_id}"
                    )
            if session_id not in self.session_ids:
                self.session_ids.append(session_id)
            return existing_context

        resolved_session_id = session_id
        if not self.session_manager.get_session(session_id):
            if not create_if_missing:
                raise ToolplaneError(f"Session {session_id} does not exist")
            resolved_session_id = self.session_manager.create_session(
                session_id=session_id,
                user_id=self.config.user_id,
                name=self.config.session_name,
                description=self.config.session_description,
                namespace=self.config.session_namespace,
                api_key=self.config.api_key,
            )

        context = SessionContext(
            resolved_session_id,
            self.connection_manager,
            self.machine_manager,
            self.tool_manager,
            self.request_manager,
            self.session_manager,
        )
        self.session_manager.register_session_context(resolved_session_id, context)
        if resolved_session_id not in self.session_ids:
            self.session_ids.append(resolved_session_id)

        if register_machine and not context.register_machine():
            raise ToolplaneError(
                f"Failed to register machine for session {resolved_session_id}"
            )

        return context

    def create_session(
        self,
        session_id: Optional[str] = None,
        user_id: Optional[str] = None,
        name: Optional[str] = None,
        description: Optional[str] = None,
        namespace: Optional[str] = None,
        register_machine: bool = False,
    ) -> SessionContext:
        """Create a new session."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")

        # Validate session name if provided
        if name and not validate_session_name(name):
            raise ToolplaneError(f"Invalid session name: {name}")

        # Generate session ID if not provided
        if not session_id:
            session_id = generate_session_id()

        try:
            # Create session on server
            created_session_id = self.session_manager.create_session(
                session_id=session_id,
                user_id=user_id or self.config.user_id,
                name=name or self.config.session_name,
                description=description or self.config.session_description,
                namespace=namespace or self.config.session_namespace,
                api_key=self.config.api_key,
            )

            print(f"Created new session: {created_session_id}")

            context = self.ensure_session_context(
                created_session_id,
                create_if_missing=False,
                register_machine=register_machine,
            )
            return context

        except Exception as e:
            raise ToolplaneError(f"Failed to create session: {e}")

    def get_session(self, session_id: str) -> Optional[SessionContext]:
        """Get session context by ID."""
        return self.session_manager.get_session_context(session_id)

    def list_sessions(self) -> List[SessionContext]:
        """List all session contexts."""
        return self.session_manager.list_session_contexts()

    # Session admin helpers (admin scope — Python-only; not portable across SDKs)
    def list_user_sessions(
        self, user_id: str, page_size: int = 10, page_token: int = 0, filter: str = ""
    ) -> Dict[str, Any]:
        """List user sessions with pagination and filtering."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.session_manager.list_user_sessions(
            user_id, page_size, page_token, filter
        )

    def bulk_delete_sessions(
        self, user_id: str, session_ids: Optional[List[str]] = None, filter: str = ""
    ) -> Dict[str, Any]:
        """Bulk delete sessions for a user."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.session_manager.bulk_delete_sessions(
            user_id, session_ids or [], filter
        )

    def get_session_stats(self, user_id: str) -> Dict[str, int]:
        """Get session statistics for a user."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.session_manager.get_session_stats(user_id)

    def refresh_session_token(self, session_id: str) -> Dict[str, str]:
        """Refresh session token."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.session_manager.refresh_session_token(session_id)

    def invalidate_session(self, session_id: str, reason: str = "") -> bool:
        """Invalidate a session."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.session_manager.invalidate_session(session_id, reason)

    def update_session(
        self,
        session_id: str,
        name: Optional[str] = None,
        description: Optional[str] = None,
        namespace: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Update session metadata."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.session_manager.update_session(
            session_id=session_id,
            name=name,
            description=description,
            namespace=namespace,
        )

    def create_request(self, session_id: str, tool_name: str, input_data: str) -> str:
        """Create a new request in a session."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.request_manager.create_request(session_id, tool_name, input_data)

    def list_requests(
        self,
        session_id: str,
        status: str = "",
        tool_name: str = "",
        limit: int = 10,
        offset: int = 0,
    ) -> List[Dict[str, Any]]:
        """List requests in a session."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.request_manager.list_requests(
            session_id=session_id,
            status=status,
            tool_name=tool_name,
            limit=limit,
            offset=offset,
        )

    def cancel_request(self, session_id: str, request_id: str) -> bool:
        """Cancel a request in a session."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.request_manager.cancel_request(session_id, request_id)

    def create_task(
        self,
        session_id: str,
        tool_name: str,
        input_data: str,
    ) -> Dict[str, Any]:
        """Create a new task in a session."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.task_manager.create_task(session_id, tool_name, input_data)

    def get_task(self, session_id: str, task_id: str) -> Dict[str, Any]:
        """Get a task by ID in a session."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.task_manager.get_task(session_id, task_id)

    def list_tasks(self, session_id: str) -> List[Dict[str, Any]]:
        """List tasks in a session."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.task_manager.list_tasks(session_id)

    def cancel_task(self, session_id: str, task_id: str) -> bool:
        """Cancel a task in a session."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.task_manager.cancel_task(session_id, task_id)

    def list_machines(self, session_id: str) -> List[Dict[str, Any]]:
        """List machines in a session."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.machine_manager.list_machines(session_id)

    def get_machine(self, session_id: str, machine_id: str) -> Dict[str, Any]:
        """Get a machine by ID."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.machine_manager.get_machine(session_id, machine_id)

    def unregister_machine(
        self, session_id: str, machine_id: Optional[str] = None
    ) -> bool:
        """Unregister a machine from a session."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.machine_manager.unregister_machine(session_id, machine_id)

    def drain_machine(self, session_id: str, machine_id: Optional[str] = None) -> bool:
        """Drain a machine from a session."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.machine_manager.drain_machine(session_id, machine_id)

    def create_api_key(self, session_id: str, name: str) -> Dict[str, Any]:
        """Create a new API key for a session."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.session_manager.create_api_key(session_id, name)

    def list_api_keys(self, session_id: str) -> List[Dict[str, Any]]:
        """List active API keys for a session."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.session_manager.list_api_keys(session_id)

    def revoke_api_key(self, session_id: str, key_id: str) -> bool:
        """Revoke an API key for a session."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.session_manager.revoke_api_key(session_id, key_id)

    def tool(
        self,
        session_id: str,
        name: Optional[str] = None,
        description: Optional[str] = None,
        stream: bool = False,
        tags: Optional[List[str]] = None,
    ) -> Callable[[Callable], Callable]:
        """Decorator to register a tool for a session."""

        def decorator(func):
            context = self.get_session(session_id)
            if not context:
                raise ToolplaneError(f"Session {session_id} not found")

            context.register_tool(
                name or func.__name__,
                func,
                description=description,
                stream=stream,
                tags=tags or [],
            )
            return func

        return decorator

    def invoke(self, tool_name: str, session_id: str, **params) -> Any:
        """Invoke a tool in a session."""
        context = self.get_session(session_id)
        if not context:
            raise ToolplaneError(f"Session {session_id} not found")

        return context.invoke(tool_name, **params)

    def ainvoke(self, tool_name: str, session_id: str, **params) -> str:
        """Invoke a tool asynchronously."""
        context = self.get_session(session_id)
        if not context:
            raise ToolplaneError(f"Session {session_id} not found")

        return context.ainvoke(tool_name, **params)

    def stream(
        self,
        tool_name: str,
        callback: Callable[[Any, bool], None],
        session_id: str,
        **params,
    ) -> List[Any]:
        """Stream tool execution."""
        context = self.get_session(session_id)
        if not context:
            raise ToolplaneError(f"Session {session_id} not found")

        return context.stream(tool_name, callback, **params)

    def astream(
        self,
        tool_name: str,
        callback: Callable[[Any, bool], None],
        session_id: str,
        **params,
    ) -> List[Any]:
        """Alias for stream method."""
        return self.stream(tool_name, callback, session_id, **params)

    def get_available_tools(self, session_id: str) -> Dict[str, Any]:
        """Get available tools for a session."""
        context = self.get_session(session_id)
        if not context:
            raise ToolplaneError(f"Session {session_id} not found")

        return context.get_available_tools()

    def list_tools(self, session_id: str) -> List[Dict[str, Any]]:
        """List tools for a session."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.tool_manager.list_tools(session_id)

    def get_tool_by_id(self, session_id: str, tool_id: str) -> Dict[str, Any]:
        """Get a tool by ID."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.tool_manager.get_tool_by_id(session_id, tool_id)

    def get_tool_by_name(self, session_id: str, tool_name: str) -> Dict[str, Any]:
        """Get a tool by name."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.tool_manager.get_tool_by_name(session_id, tool_name)

    def delete_tool(self, session_id: str, tool_id: str) -> bool:
        """Delete a tool by ID."""
        if not self.connection_manager.connected:
            if not self.connect():
                raise ConnectionError("Failed to connect to server")
        return self.tool_manager.delete_tool(session_id, tool_id)

    def get_request_status(self, request_id: str, session_id: str) -> Dict[str, Any]:
        """Get request status."""
        context = self.get_session(session_id)
        if not context:
            raise ToolplaneError(f"Session {session_id} not found")

        return context.get_request_status(request_id)

    def provider_runtime(self, session_ids: Optional[List[str]] = None):
        """Return the explicit provider runtime for this client."""
        try:
            from .provider_runtime import ProviderRuntime
        except ImportError:
            from provider_runtime import ProviderRuntime

        if self._provider_runtime is None:
            self._provider_runtime = ProviderRuntime(self, session_ids=session_ids)
        elif session_ids:
            self._provider_runtime.add_sessions(session_ids)
        return self._provider_runtime

    def start(self) -> None:
        """Backward-compatible alias for the explicit provider runtime."""
        self.provider_runtime().run_forever()

    def stop(self) -> None:
        """Stop the explicit provider runtime if it is active."""
        self.running = False
        if self._provider_runtime is not None:
            self._provider_runtime.stop()

    def _main_loop(self) -> None:
        """Main polling loop for all sessions."""
        while self.running:
            try:
                for context in self.list_sessions():
                    if context.machine_id:
                        context.poll_requests()

                time.sleep(self.config.poll_interval)

            except Exception as e:
                print(f"Error in main loop: {e}")
                time.sleep(1)

    def __enter__(self):
        """Context manager entry."""
        self.connect()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.disconnect()

    # Legacy methods for backward compatibility
    def _register_machine_internal(self):
        """Legacy method for backward compatibility."""
        if not self.session_ids:
            return False
        for session_id in self.session_ids:
            ctx = self.get_session(session_id)
            if ctx is None or getattr(ctx, "machine_id", None) is None:
                return False
        return True

    def _register_tool(
        self,
        name: str,
        func: Callable,
        schema: Dict,
        stream: bool = False,
        tags: Optional[List[str]] = None,
    ):
        """Legacy method for registering tools."""
        # For backward compatibility, register on all sessions
        for session_id in self.session_ids:
            context = self.get_session(session_id)
            if context:
                context.register_tool(name, func, schema, stream=stream, tags=tags)

    def _register_tools_with_server(self):
        """Legacy method - tools are now registered immediately."""
        pass

    def _poll_pending_requests(self):
        """Legacy method - polling is now handled by main loop."""
        pass
