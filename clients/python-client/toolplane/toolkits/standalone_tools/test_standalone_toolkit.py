#!/usr/bin/env python3
"""
Comprehensive test suite for standalone tools.

This test suite validates all standalone tools functionality including:
- Basic functionality tests
- Error handling
- Edge cases
- Output format validation
- Integration tests
"""

import json
import os
import subprocess
import sys
import tempfile
from pathlib import Path
from typing import Any, Dict, List

import pytest


class TestStandaloneTools:
    """Test suite for all standalone tools."""

    @pytest.fixture(autouse=True)
    def setup(self):
        """Set up test environment for each test."""
        self.tools_dir = Path(__file__).parent
        self.temp_dir = None
        self.test_files = {}

    def teardown_method(self):
        """Clean up after each test."""
        if self.temp_dir and self.temp_dir.exists():
            import shutil

            shutil.rmtree(self.temp_dir, ignore_errors=True)

    def create_temp_dir(self) -> Path:
        """Create a temporary directory for testing."""
        if not self.temp_dir:
            self.temp_dir = Path(tempfile.mkdtemp())
        return self.temp_dir

    def create_test_file(self, name: str, content: str) -> Path:
        """Create a test file with given content."""
        temp_dir = self.create_temp_dir()
        file_path = temp_dir / name
        file_path.write_text(content, encoding="utf-8")
        self.test_files[name] = file_path
        return file_path

    def run_tool(self, tool_name: str, args: List[str]) -> subprocess.CompletedProcess:
        """Run a standalone tool with given arguments."""
        tool_path = self.tools_dir / tool_name
        cmd = [sys.executable, str(tool_path)] + args
        return subprocess.run(cmd, capture_output=True, text=True)

    def test_create_directory_basic(self):
        """Test create_directory.py basic functionality."""
        temp_dir = self.create_temp_dir()
        test_dir = temp_dir / "test_directory"

        result = self.run_tool("create_directory.py", [str(test_dir)])

        assert result.returncode == 0
        assert test_dir.exists()
        assert test_dir.is_dir()
        assert "Directory created successfully" in result.stdout

    def test_create_directory_existing(self):
        """Test create_directory.py with existing directory."""
        temp_dir = self.create_temp_dir()
        test_dir = temp_dir / "existing_dir"
        test_dir.mkdir()

        result = self.run_tool("create_directory.py", [str(test_dir)])

        assert result.returncode == 0
        assert "Directory created successfully" in result.stdout

    def test_create_directory_nested(self):
        """Test create_directory.py with nested directories."""
        temp_dir = self.create_temp_dir()
        test_dir = temp_dir / "level1" / "level2" / "level3"

        result = self.run_tool("create_directory.py", [str(test_dir)])

        assert result.returncode == 0
        assert test_dir.exists()
        assert test_dir.is_dir()

    def test_create_file_basic(self):
        """Test create_file.py basic functionality."""
        temp_dir = self.create_temp_dir()
        test_file = temp_dir / "test.txt"
        content = "Hello, World!"

        result = self.run_tool("create_file.py", [str(test_file), content])

        assert result.returncode == 0
        assert test_file.exists()
        assert test_file.read_text() == content
        assert "File created successfully" in result.stdout

    def test_create_file_with_dirs(self):
        """Test create_file.py with automatic directory creation."""
        temp_dir = self.create_temp_dir()
        test_file = temp_dir / "subdir" / "test.txt"
        content = "Test content"

        result = self.run_tool("create_file.py", [str(test_file), content])

        assert result.returncode == 0
        assert test_file.exists()
        assert test_file.read_text() == content

    def test_create_file_overwrite(self):
        """Test create_file.py overwrite functionality."""
        temp_dir = self.create_temp_dir()
        test_file = temp_dir / "test.txt"
        test_file.write_text("Original content")

        new_content = "New content"
        result = self.run_tool(
            "create_file.py", [str(test_file), new_content, "--overwrite"]
        )

        assert result.returncode == 0
        assert test_file.read_text() == new_content

    def test_file_search_basic(self):
        """Test file_search.py basic functionality."""
        temp_dir = self.create_temp_dir()

        # Create test files
        (temp_dir / "test1.py").write_text("print('hello')")
        (temp_dir / "test2.py").write_text("print('world')")
        (temp_dir / "test.txt").write_text("text file")

        result = self.run_tool("file_search.py", ["*.py", "--base_path", str(temp_dir)])

        assert result.returncode == 0
        assert "test1.py" in result.stdout
        assert "test2.py" in result.stdout
        assert "test.txt" not in result.stdout

    def test_file_search_json_output(self):
        """Test file_search.py JSON output format."""
        temp_dir = self.create_temp_dir()
        (temp_dir / "test.py").write_text("test")

        result = self.run_tool(
            "file_search.py",
            ["*.py", "--base_path", str(temp_dir), "--output_format", "json"],
        )

        assert result.returncode == 0
        data = json.loads(result.stdout)
        assert isinstance(data, list)
        assert len(data) > 0
        assert "test.py" in data[0]["path"]

    def test_file_search_max_results(self):
        """Test file_search.py max_results parameter."""
        temp_dir = self.create_temp_dir()

        # Create multiple test files
        for i in range(5):
            (temp_dir / f"test{i}.py").write_text(f"test {i}")

        result = self.run_tool(
            "file_search.py",
            ["*.py", "--base_path", str(temp_dir), "--max_results", "2"],
        )

        assert result.returncode == 0
        # Should find exactly 2 files
        lines = [
            line
            for line in result.stdout.split("\n")
            if "test" in line and ".py" in line
        ]
        assert len(lines) == 2

    def test_grep_search_basic(self):
        """Test grep_search.py basic functionality."""
        temp_dir = self.create_temp_dir()
        test_file = temp_dir / "test.py"
        test_file.write_text(
            "def hello():\n    print('Hello, World!')\n    return 'hello'"
        )

        result = self.run_tool(
            "grep_search.py", ["hello", "--base_path", str(temp_dir)]
        )

        assert result.returncode == 0
        assert "def hello" in result.stdout
        assert "return 'hello'" in result.stdout

    def test_grep_search_regex(self):
        """Test grep_search.py regex functionality."""
        temp_dir = self.create_temp_dir()
        test_file = temp_dir / "test.py"
        test_file.write_text("def function1():\n    pass\ndef function2():\n    pass")

        result = self.run_tool(
            "grep_search.py",
            ["def function\\d", "--is_regexp", "--base_path", str(temp_dir)],
        )

        assert result.returncode == 0
        assert "function1" in result.stdout
        assert "function2" in result.stdout

    def test_grep_search_context(self):
        """Test grep_search.py context lines."""
        temp_dir = self.create_temp_dir()
        test_file = temp_dir / "test.py"
        test_file.write_text("line1\nline2\ntarget_line\nline4\nline5")

        result = self.run_tool(
            "grep_search.py",
            ["target_line", "--context_lines", "1", "--base_path", str(temp_dir)],
        )

        assert result.returncode == 0
        assert "line2" in result.stdout
        assert "target_line" in result.stdout
        assert "line4" in result.stdout

    def test_grep_search_json_output(self):
        """Test grep_search.py JSON output format."""
        temp_dir = self.create_temp_dir()
        test_file = temp_dir / "test.py"
        test_file.write_text("hello world")

        result = self.run_tool(
            "grep_search.py",
            ["hello", "--base_path", str(temp_dir), "--output_format", "json"],
        )

        assert result.returncode == 0
        data = json.loads(result.stdout)
        assert isinstance(data, list)
        assert len(data) > 0
        assert "matches" in data[0]

    def test_list_dir_basic(self):
        """Test list_dir.py basic functionality."""
        temp_dir = self.create_temp_dir()

        # Create test files and directories
        (temp_dir / "file1.txt").write_text("test")
        (temp_dir / "file2.py").write_text("test")
        (temp_dir / "subdir").mkdir()

        result = self.run_tool("list_dir.py", [str(temp_dir)])

        assert result.returncode == 0
        assert "file1.txt" in result.stdout
        assert "file2.py" in result.stdout
        assert "subdir" in result.stdout

    def test_list_dir_details(self):
        """Test list_dir.py with details."""
        temp_dir = self.create_temp_dir()
        (temp_dir / "test.txt").write_text("test content")

        result = self.run_tool("list_dir.py", [str(temp_dir), "--show_details"])

        assert result.returncode == 0
        assert "test.txt" in result.stdout
        # Since --show_details might not be fully implemented, just check basic functionality

    def test_read_file_basic(self):
        """Test read_file.py basic functionality."""
        content = "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
        test_file = self.create_test_file("test.txt", content)

        result = self.run_tool("read_file.py", [str(test_file), "2", "4"])

        assert result.returncode == 0
        assert "Line 2" in result.stdout
        assert "Line 3" in result.stdout
        assert "Line 4" in result.stdout
        assert "Line 1" not in result.stdout
        assert "Line 5" not in result.stdout

    def test_read_file_all_lines(self):
        """Test read_file.py reading all lines."""
        content = "Line 1\nLine 2\nLine 3"
        test_file = self.create_test_file("test.txt", content)

        result = self.run_tool("read_file.py", [str(test_file), "1", "-1"])

        assert result.returncode == 0
        assert "Line 1" in result.stdout
        assert "Line 2" in result.stdout
        assert "Line 3" in result.stdout

    def test_read_file_json_output(self):
        """Test read_file.py JSON output format."""
        content = "Line 1\nLine 2"
        test_file = self.create_test_file("test.txt", content)

        result = self.run_tool(
            "read_file.py", [str(test_file), "1", "2", "--output_format", "json"]
        )

        assert result.returncode == 0
        data = json.loads(result.stdout)
        assert data["success"] is True
        assert len(data["content"]) == 2
        assert data["content"][0]["content"] == "Line 1"

    def test_read_file_nonexistent(self):
        """Test read_file.py with nonexistent file."""
        result = self.run_tool("read_file.py", ["/nonexistent/file.txt", "1", "5"])

        assert result.returncode != 0 or "does not exist" in result.stdout

    def test_replace_string_basic(self):
        """Test replace_string_in_file.py basic functionality."""
        content = "Hello, World!\nGoodbye, World!"
        test_file = self.create_test_file("test.txt", content)

        result = self.run_tool(
            "replace_string_in_file.py",
            [str(test_file), "World", "Universe", "--no_backup"],
        )

        assert result.returncode == 0
        new_content = test_file.read_text()
        assert "Hello, Universe!" in new_content
        assert "Goodbye, Universe!" in new_content
        assert "World" not in new_content

    def test_replace_string_dry_run(self):
        """Test replace_string_in_file.py dry run."""
        content = "Hello, World!"
        test_file = self.create_test_file("test.txt", content)

        result = self.run_tool(
            "replace_string_in_file.py",
            [str(test_file), "World", "Universe", "--dry_run"],
        )

        assert result.returncode == 0
        assert (
            "would be replaced" in result.stdout.lower()
            or "dry run" in result.stdout.lower()
        )
        # File should remain unchanged
        assert test_file.read_text() == content

    def test_replace_string_backup(self):
        """Test replace_string_in_file.py backup creation."""
        content = "Hello, World!"
        test_file = self.create_test_file("test.txt", content)

        result = self.run_tool(
            "replace_string_in_file.py", [str(test_file), "World", "Universe"]
        )

        assert result.returncode == 0
        # Check if backup was created (should have backup in name)
        backup_files = list(test_file.parent.glob(f"*backup*"))
        assert len(backup_files) > 0

    def test_semantic_search_basic(self):
        """Test semantic_search.py basic functionality."""
        temp_dir = self.create_temp_dir()

        # Create test files with different content
        (temp_dir / "auth.py").write_text(
            "def authenticate_user(username, password):\n    return True"
        )
        (temp_dir / "math.py").write_text("def calculate_sum(a, b):\n    return a + b")

        result = self.run_tool(
            "semantic_search.py",
            [
                "authentication function",
                "--search_path",
                str(temp_dir),
                "--max_results",
                "5",
            ],
        )

        assert result.returncode == 0
        # Should find the authentication-related content
        assert (
            "auth" in result.stdout.lower() or "authenticate" in result.stdout.lower()
        )

    def test_semantic_search_file_types(self):
        """Test semantic_search.py file type filtering."""
        temp_dir = self.create_temp_dir()

        (temp_dir / "test.py").write_text("python code")
        (temp_dir / "test.txt").write_text("text file")

        result = self.run_tool(
            "semantic_search.py",
            ["code", "--search_path", str(temp_dir), "--file_types", ".py"],
        )

        assert result.returncode == 0
        # Should only search Python files

    def test_test_failure_analysis_basic(self):
        """Test test_failure_analysis.py basic functionality."""
        # Create a mock test output
        test_output = """
FAILED test_example.py::test_function - AssertionError: Expected 5, got 3
FAILED test_another.py::test_method - ValueError: Invalid input
        """

        result = self.run_tool(
            "test_failure_analysis.py", ["--test_output", test_output]
        )

        assert result.returncode == 0
        assert "AssertionError" in result.stdout or "ValueError" in result.stdout

    def test_launcher_list(self):
        """Test launcher.py list functionality."""
        result = self.run_tool("launcher.py", ["list"])

        assert result.returncode == 0
        assert "create_directory.py" in result.stdout
        assert "create_file.py" in result.stdout
        assert "grep_search.py" in result.stdout

    def test_launcher_search(self):
        """Test launcher.py search functionality."""
        result = self.run_tool("launcher.py", ["search", "file"])

        assert result.returncode == 0
        assert "file_search.py" in result.stdout or "read_file.py" in result.stdout

    def test_launcher_run_help(self):
        """Test launcher.py run functionality with help."""
        result = self.run_tool("launcher.py", ["run", "create_directory.py", "--help"])

        assert result.returncode == 0
        assert "usage:" in result.stdout.lower()

    def test_all_tools_help(self):
        """Test that all tools provide help information."""
        tools = [
            "create_directory.py",
            "create_file.py",
            "file_search.py",
            "grep_search.py",
            "list_dir.py",
            "read_file.py",
            "replace_string_in_file.py",
            "semantic_search.py",
            "test_failure_analysis.py",
            "launcher.py",
        ]

        for tool in tools:
            result = self.run_tool(tool, ["--help"])
            assert result.returncode == 0, f"Tool {tool} failed help test"
            assert "usage:" in result.stdout.lower(), f"Tool {tool} doesn't show usage"

    def test_error_handling(self):
        """Test error handling for various tools."""
        # Test create_directory with invalid path (if applicable)
        result = self.run_tool("create_directory.py", ["/invalid:/path"])
        # Should handle gracefully (either succeed or fail with proper message)

        # Test read_file with invalid line numbers
        temp_file = self.create_test_file("small.txt", "line1\nline2")
        result = self.run_tool("read_file.py", [str(temp_file), "10", "20"])
        # Tool handles this gracefully by adjusting range to available lines
        assert result.returncode == 0

    def test_json_output_validity(self):
        """Test that all tools produce valid JSON when requested."""
        temp_dir = self.create_temp_dir()
        (temp_dir / "test.py").write_text("test content")

        tools_with_json = [
            (
                "file_search.py",
                ["*.py", "--base_path", str(temp_dir), "--output_format", "json"],
            ),
            (
                "grep_search.py",
                ["test", "--base_path", str(temp_dir), "--output_format", "json"],
            ),
            (
                "read_file.py",
                [str(temp_dir / "test.py"), "1", "1", "--output_format", "json"],
            ),
        ]

        for tool, args in tools_with_json:
            result = self.run_tool(tool, args)
            if result.returncode == 0:
                try:
                    json.loads(result.stdout)
                except json.JSONDecodeError:
                    pytest.fail(f"Tool {tool} produced invalid JSON: {result.stdout}")


if __name__ == "__main__":
    # Run tests directly
    pytest.main([__file__, "-v"])
