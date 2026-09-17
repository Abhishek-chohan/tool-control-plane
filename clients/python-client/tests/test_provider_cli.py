"""The toolplane-provider surface: module registry, execution-free
validation, local calls, generated schemas, and the serve-side late
session binding."""

import json

import pytest

from toolplane.provider_cli import (
    _decorated_functions,
    _registry_tools,
    cmd_call,
    cmd_schema,
    cmd_validate,
)
from toolplane.provider_registry import clear


def write_tools_file(tmp_path, body):
    path = tmp_path / "tools.py"
    path.write_text(body, encoding="utf-8")
    return str(path)


TOOLS_FILE = """
from toolplane.provider_registry import tool


@tool(name="add", description="Add two numbers")
def add(a: int, b: int) -> int:
    return a + b


@tool
def echo(message: str = "hi") -> str:
    return message


@tool(name="bad", schema="{not json")
def bad(a: int) -> int:
    return a
"""

VALID_TOOLS_FILE = """
from toolplane.provider_registry import tool


@tool(name="add", description="Add two numbers")
def add(a: int, b: int) -> int:
    return a + b


@tool
def echo(message: str = "hi") -> str:
    return message
"""


@pytest.fixture(autouse=True)
def _clean_registry():
    clear()
    yield
    clear()


class TestRegistry:
    def test_collects_name_and_metadata(self, tmp_path):
        path = write_tools_file(tmp_path, TOOLS_FILE)
        tools = _registry_tools(path)
        names = [t.name for t in tools]
        assert names == ["add", "echo", "bad"]
        assert tools[0].description == "Add two numbers"
        assert callable(tools[0].func)
        assert tools[0].func(a=2, b=3) == 5

    def test_collect_is_a_snapshot(self, tmp_path):
        path = write_tools_file(tmp_path, TOOLS_FILE)
        first = _registry_tools(path)
        second = _registry_tools(path)
        assert len(first) == len(second) == 3
        assert first is not second


class TestValidate:
    def test_valid_file_passes_and_names_tools(self, tmp_path, capsys):
        path = write_tools_file(tmp_path, VALID_TOOLS_FILE)
        assert cmd_validate(type("A", (), {"file": path})()) == 0
        output = capsys.readouterr().out
        assert "add" in output and "echo" in output

    def test_malformed_schema_fails(self, tmp_path, capsys):
        path = write_tools_file(tmp_path, TOOLS_FILE)
        # The file carries one deliberately invalid schema literal.
        assert cmd_validate(type("A", (), {"file": path})()) == 2
        captured = capsys.readouterr()
        assert "invalid schema(s)" in captured.err
        assert "not valid JSON" in captured.out

    def test_syntax_error_fails(self, tmp_path, capsys):
        path = write_tools_file(tmp_path, "def broken(:\n")
        assert cmd_validate(type("A", (), {"file": path})()) == 2
        assert "not valid Python" in capsys.readouterr().err

    def test_no_tools_fails(self, tmp_path, capsys):
        path = write_tools_file(tmp_path, "x = 1\n")
        assert cmd_validate(type("A", (), {"file": path})()) == 2
        assert "no @tool functions" in capsys.readouterr().err

    def test_decorated_functions_found_without_execution(self):
        pairs = _decorated_functions(TOOLS_FILE)
        # AST walk only: the module was parsed, never executed (the file
        # imports toolplane.provider_registry — collection happens on
        # execution; AST discovery is static).
        assert [name for name, _ in pairs] == ["add", "echo", "bad"]


class TestLocalCallAndSchema:
    def test_call_executes_without_server(self, tmp_path, capsys):
        path = write_tools_file(tmp_path, TOOLS_FILE)
        args = type("A", (), {"file": path, "tool": "add", "input": '{"a":2,"b":40}'})()
        assert cmd_call(args) == 0
        assert "42" in capsys.readouterr().out

    def test_call_unknown_tool_exits_3(self, tmp_path):
        path = write_tools_file(tmp_path, TOOLS_FILE)
        args = type("A", (), {"file": path, "tool": "nope", "input": "{}"})()
        assert cmd_call(args) == 3

    def test_schema_prints_generated_json(self, tmp_path, capsys):
        path = write_tools_file(tmp_path, TOOLS_FILE)
        args = type("A", (), {"file": path, "tool": "add"})()
        assert cmd_schema(args) == 0
        schema = json.loads(capsys.readouterr().out)
        assert schema.get("type") == "object"


class TestLateSessionBinding:
    def test_resolver_picks_single_managed_session(self):
        """A deferred tool with session_id=None resolves to the runtime's
        single managed session."""
        from toolplane.provider_runtime import ProviderRuntime

        config = type("Config", (), {"poll_interval": 1.0, "heartbeat_interval": 60})()
        client = type("C", (), {"config": config})()
        runtime = ProviderRuntime(client=client)
        runtime._session_ids.add("sess-only")
        assert runtime._resolve_default_session() == "sess-only"

    def test_resolver_is_ambiguous_without_exactly_one_session(self):
        from toolplane.provider_runtime import ProviderRuntime

        config = type("Config", (), {"poll_interval": 1.0, "heartbeat_interval": 60})()
        client = type("C", (), {"config": config})()
        runtime = ProviderRuntime(client=client)
        with pytest.raises(Exception) as empty:
            runtime._resolve_default_session()
        assert "no session_id" in str(empty.value)

        runtime._session_ids.update(["sess-a", "sess-b"])
        with pytest.raises(Exception) as ambiguous:
            runtime._resolve_default_session()
        assert "exactly one" in str(ambiguous.value)
