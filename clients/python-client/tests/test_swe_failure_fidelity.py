"""Failure-fidelity contract for the SWE toolkit.

Tool failures must raise — never return error text — so the provider submits
a rejection, the durable request records FAILED, and MCP clients see
isError: true. Models, eval scoring, and retry heuristics rely on reading the
failure from the status instead of parsing output strings.
"""

import pytest

from toolplane.toolkits.swe.swe_toolkit import BashTool, ToolExecutionError


def test_bash_nonzero_exit_raises_with_captured_output():
    tool = BashTool()
    with pytest.raises(ToolExecutionError) as excinfo:
        tool._run(command="echo stdout-marker && echo stderr-marker 1>&2 && exit 3")
    message = str(excinfo.value)
    assert "stdout-marker" in message
    assert "stderr-marker" in message


def test_bash_start_failure_raises():
    tool = BashTool()

    def _broken_run_command(_command):
        raise OSError("shell exploded")

    original = tool._run.__globals__["run_command"]
    try:
        import toolplane.toolkits.swe.swe_toolkit as swe_module

        swe_module.run_command = _broken_run_command
        with pytest.raises(ToolExecutionError, match="shell exploded"):
            tool._run(command="echo hi")
    finally:
        swe_module.run_command = original


def test_bash_success_still_returns_output():
    tool = BashTool()
    output = tool._run(command="echo success-marker")
    assert "success-marker" in output
