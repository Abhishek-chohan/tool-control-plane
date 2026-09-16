"""Reconnect re-registration must not swallow registration failures.

Registration is an upsert: a same-machine re-register never conflicts, so
every failure surfaced during ``reregister_session_tools`` is real —
including a tool name owned by another live machine (FAILED_PRECONDITION
on the wire) — and must be logged as a warning, not silently dropped the
way the old ALREADY_EXISTS string-match swallow did.
"""

import logging
from typing import List

from toolplane.common.base_tool_manager import BaseToolManager


class _StubToolManager(BaseToolManager):
    """Records registration attempts; "owned_by_rival" always fails."""

    def __init__(self):
        super().__init__(connection_manager=None)
        self.registered: List[str] = []

    def _register_tool_with_server(self, session_id, machine_id, tool_name, schema):
        self.registered.append(tool_name)
        if tool_name == "owned_by_rival":
            raise RuntimeError(
                "failed to register tool: rpc error: code = FailedPrecondition"
            )

    def _execute_tool_on_server(self, *args, **kwargs):
        raise NotImplementedError

    def _stream_tool_on_server(self, *args, **kwargs):
        raise NotImplementedError

    def _unregister_tool_from_server(self, *args, **kwargs):
        raise NotImplementedError

    def _get_available_tools_from_server(self, *args, **kwargs):
        return []


def test_reregister_surfaces_conflict_as_warning(caplog):
    manager = _StubToolManager()
    manager.tool_schemas["session-reconnect"] = {
        "owned_by_rival": {"type": "object"},
        "healthy_tool": {"type": "object"},
    }

    with caplog.at_level(
        logging.WARNING, logger="toolplane.common.base_tool_manager"
    ):
        manager.reregister_session_tools("session-reconnect", "machine-1")

    # Both tools were attempted: a failing registration does not abort the
    # reconnect walk.
    assert manager.registered == ["owned_by_rival", "healthy_tool"]

    warnings = [r for r in caplog.records if r.levelno == logging.WARNING]
    assert any("owned_by_rival" in record.getMessage() for record in warnings), (
        "registration failures must surface as warnings, not be swallowed"
    )
