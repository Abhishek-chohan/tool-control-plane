"""Module-level tool registry for file-based tool modules.

A tools file decorated with the bare ``@tool`` imports nothing from
Toolplane and binds no session at import time — metadata is collected
here and the ``toolplane-provider`` CLI binds it to a session at serve.
Importing such a module performs no I/O.

    from toolplane.provider_registry import tool

    @tool(name="add", description="Add two numbers")
    def add(a: int, b: int) -> int:
        return a + b
"""

import threading
from dataclasses import dataclass
from typing import Callable, List, Optional

_lock = threading.Lock()
_registry: List["RegistryTool"] = []


@dataclass
class RegistryTool:
    """A collected tool definition, pre-binding."""

    name: str
    func: Callable
    description: Optional[str] = None
    stream: bool = False
    tags: Optional[List[str]] = None
    schema: Optional[dict] = None


def tool(
    name: Optional[str] = None,
    description: Optional[str] = None,
    stream: bool = False,
    tags: Optional[List[str]] = None,
    schema: Optional[dict] = None,
):
    """Collect a tool definition into the module registry. Import-time
    only: no session, no server, no I/O. Works bare (@tool) or with
    arguments (@tool(name=..., description=...))."""

    def _register(func: Callable) -> Callable:
        with _lock:
            _registry.append(
                RegistryTool(
                    name=name or func.__name__,
                    func=func,
                    description=description,
                    stream=stream,
                    tags=list(tags) if tags else None,
                    schema=schema,
                )
            )
        return func

    # Bare usage: @tool directly above a function.
    if callable(name):
        func, name = name, None
        return _register(func)

    return _register


def collect() -> List[RegistryTool]:
    """Snapshot the collected tools."""
    with _lock:
        return list(_registry)


def clear() -> None:
    """Reset the registry (test isolation)."""
    with _lock:
        _registry.clear()
