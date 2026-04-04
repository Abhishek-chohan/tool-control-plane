from typing import Any, Callable, Dict, List

from toolplane.utils import generate_schema_from_function

__all__ = ["SessionContext"]


class SessionContext:
    """
    Represents a session context with its own tools and machine registration.
    Encapsulates all session-specific state and operations.
    """

    def __init__(self, client: Any, session_id: str, machine_id: str = None):
        """
        Initialize a session context.

        Args:
            client: Reference to the parent Toolplane client
            session_id: The session ID for this context
            machine_id: The machine ID registered for this session
        """
        self.client = client
        self.session_id = session_id
        self.machine_id = machine_id

        # Session-specific tool storage
        self.tools = {}
        self.tool_schemas = {}
        self.streaming_tools = set()

    def register_tool(
        self,
        name: str,
        func: Callable,
        schema: Dict,
        stream: bool = False,
        tags: List[str] = None,
    ):
        """Register a tool for this specific session."""
        if tags is None:
            tags = []

        # Add tags to schema
        schema["tags"] = tags

        self.tools[name] = func
        self.tool_schemas[name] = schema

        # Mark as streaming tool if applicable
        if stream:
            self.streaming_tools.add(name)

        self.client.register_tool(
            name=name,
            func=func,
            schema=schema,
            stream=stream,
            session_id=self.session_id,
            tags=tags,
        )

    def invoke(self, tool_name: str, **params):
        """Invoke a tool in this session context."""
        return self.client.invoke(tool_name, session_id=self.session_id, **params)

    def ainvoke(self, tool_name: str, **params):
        """Asynchronously invoke a tool in this session context."""
        return self.client.ainvoke(tool_name, session_id=self.session_id, **params)

    def get_available_tools(self):
        """Get tools available in this session."""
        return self.client.get_available_tools(session_id=self.session_id)

    def get_request_status(self, request_id: str):
        """Get request status for this session."""
        return self.client.get_request_status(request_id, session_id=self.session_id)

    def stream(self, tool_name: str, callback: Callable[[Any, bool], None], **params):
        """Stream results from a tool execution in this session."""
        return self.client.stream(
            tool_name, callback, session_id=self.session_id, **params
        )

    def astream(self, tool_name: str, callback: Callable[[Any, bool], None], **params):
        """Alias for stream method."""
        return self.stream(tool_name, callback, **params)

    def tool(self, name=None, description=None, stream=False, tags=None):
        """Session-specific tool decorator."""
        if tags is None:
            tags = []

        def decorator(func):
            tool_name = name or func.__name__
            tool_schema = generate_schema_from_function(func)

            if description:
                tool_schema["description"] = description

            self.register_tool(tool_name, func, tool_schema, stream, tags)
            return func

        return decorator

    def start(self):
        """Start monitoring this session for requests."""
        if not self.machine_id:
            print(f"No machine registered for session {self.session_id}. Cannot start.")
            return False

        print(
            f"Started monitoring session {self.session_id} with machine {self.machine_id}"
        )
        print(f"Registered tools: {list(self.tools.keys())}")
        return True

    def stop(self):
        """Stop monitoring this session and cleanup."""
        self.client._cleanup_session(self)
        print(f"Stopped and cleaned up session {self.session_id}")
