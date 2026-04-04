#!/usr/bin/env python3
"""
Description: Read file contents with enhanced features and line range support.

This tool reads the contents of a file with support for line ranges,
encoding detection, and various output formats.

Parameters:
  file_path (string, required): The absolute path of the file to read.
  start_line (integer, required): The line number to start reading from (1-based).
  end_line (integer, required): The inclusive line number to end reading at (1-based).
  encoding (string, optional): The encoding to use for reading the file (default: auto-detect).
  show_line_numbers (boolean, optional): Show line numbers in output (default: True).
  highlight_syntax (boolean, optional): Attempt to highlight syntax if possible (default: False).
  max_line_length (integer, optional): Maximum line length before truncation (default: 1000).
"""

import argparse
import os
import sys
from pathlib import Path
from typing import Any, Dict, List, Optional


def read_file(
    file_path: str,
    start_line: int,
    end_line: int,
    encoding: str = None,
    show_line_numbers: bool = True,
    highlight_syntax: bool = False,
    max_line_length: int = 1000,
) -> Dict[str, Any]:
    """
    Read file contents with enhanced features and line range support.

    Args:
        file_path: The absolute path of the file to read
        start_line: Line number to start reading from (1-based)
        end_line: Line number to end reading at (1-based)
        encoding: Encoding to use for reading the file
        show_line_numbers: Show line numbers in output
        highlight_syntax: Attempt to highlight syntax
        max_line_length: Maximum line length before truncation

    Returns:
        Dictionary containing file content and metadata
    """
    try:
        path = Path(file_path)

        if not path.exists():
            return {
                "success": False,
                "error": f"File does not exist: {file_path}",
                "content": None,
            }

        if not path.is_file():
            return {
                "success": False,
                "error": f"Path is not a file: {file_path}",
                "content": None,
            }

        # Detect encoding if not specified
        if encoding is None:
            encoding = detect_encoding(path)

        # Read the file
        try:
            with open(path, "r", encoding=encoding, errors="replace") as f:
                all_lines = f.readlines()
        except UnicodeDecodeError:
            # Fallback to UTF-8 with error handling
            with open(path, "r", encoding="utf-8", errors="replace") as f:
                all_lines = f.readlines()

        total_lines = len(all_lines)

        # Validate line numbers
        if start_line < 1:
            start_line = 1
        if end_line < 1:
            end_line = total_lines
        if end_line == -1:
            end_line = total_lines

        start_line = min(start_line, total_lines)
        end_line = min(end_line, total_lines)

        if start_line > end_line:
            return {
                "success": False,
                "error": f"Start line ({start_line}) is greater than end line ({end_line})",
                "content": None,
            }

        # Extract the requested lines
        selected_lines = all_lines[start_line - 1 : end_line]

        # Process lines
        processed_lines = []
        for i, line in enumerate(selected_lines, start=start_line):
            # Remove trailing newline
            line = line.rstrip("\n\r")

            # Truncate long lines
            if len(line) > max_line_length:
                line = line[:max_line_length] + "... [line truncated]"

            processed_lines.append(
                {
                    "line_number": i,
                    "content": line,
                    "original_length": len(all_lines[i - 1].rstrip("\n\r")),
                }
            )

        # Get file metadata
        stat_info = path.stat()

        result = {
            "success": True,
            "file_path": str(path.resolve()),
            "encoding": encoding,
            "total_lines": total_lines,
            "start_line": start_line,
            "end_line": end_line,
            "lines_returned": len(processed_lines),
            "file_size": stat_info.st_size,
            "content": processed_lines,
            "show_line_numbers": show_line_numbers,
            "highlight_syntax": highlight_syntax
            and get_file_language(path) is not None,
            "file_language": get_file_language(path),
        }

        return result

    except PermissionError:
        return {
            "success": False,
            "error": f"Permission denied: {file_path}",
            "content": None,
        }
    except Exception as e:
        return {"success": False, "error": f"Error reading file: {e}", "content": None}


def detect_encoding(file_path: Path) -> str:
    """Detect file encoding using simple heuristics."""
    try:
        # Try common encodings in order
        encodings = ["utf-8", "utf-16", "latin-1", "cp1252"]

        for encoding in encodings:
            try:
                with open(file_path, "r", encoding=encoding) as f:
                    f.read(1024)  # Try to read first 1KB
                return encoding
            except UnicodeDecodeError:
                continue

        # If all fail, return utf-8 as default
        return "utf-8"

    except Exception:
        return "utf-8"


def get_file_language(file_path: Path) -> Optional[str]:
    """Determine the programming language based on file extension."""
    extension_map = {
        ".py": "python",
        ".js": "javascript",
        ".ts": "typescript",
        ".jsx": "javascript",
        ".tsx": "typescript",
        ".java": "java",
        ".cpp": "cpp",
        ".c": "c",
        ".h": "c",
        ".hpp": "cpp",
        ".cs": "csharp",
        ".php": "php",
        ".rb": "ruby",
        ".go": "go",
        ".rs": "rust",
        ".swift": "swift",
        ".kt": "kotlin",
        ".scala": "scala",
        ".r": "r",
        ".sql": "sql",
        ".html": "html",
        ".htm": "html",
        ".css": "css",
        ".scss": "scss",
        ".sass": "sass",
        ".less": "less",
        ".xml": "xml",
        ".json": "json",
        ".yaml": "yaml",
        ".yml": "yaml",
        ".toml": "toml",
        ".ini": "ini",
        ".cfg": "ini",
        ".conf": "ini",
        ".md": "markdown",
        ".rst": "rst",
        ".tex": "latex",
        ".sh": "bash",
        ".bash": "bash",
        ".zsh": "zsh",
        ".fish": "fish",
        ".ps1": "powershell",
        ".bat": "batch",
        ".dockerfile": "dockerfile",
        ".makefile": "makefile",
        ".cmake": "cmake",
        ".gradle": "gradle",
    }

    return extension_map.get(file_path.suffix.lower())


def format_output(result: Dict[str, Any], output_format: str) -> str:
    """Format the output according to the specified format."""
    if not result["success"]:
        return f"Error: {result['error']}"

    if output_format == "json":
        import json

        return json.dumps(result, indent=2)

    elif output_format == "raw":
        return "\n".join(line["content"] for line in result["content"])

    else:  # default format
        lines = []

        # Header
        lines.append(f"File: {result['file_path']}")
        lines.append(
            f"Lines: {result['start_line']}-{result['end_line']} (of {result['total_lines']})"
        )
        lines.append(f"Encoding: {result['encoding']}")
        if result["file_language"]:
            lines.append(f"Language: {result['file_language']}")
        lines.append("-" * 60)

        # Content
        for line_info in result["content"]:
            if result["show_line_numbers"]:
                line_num = f"{line_info['line_number']:4d}"
                content = line_info["content"]

                # Show truncation indicator
                if line_info["original_length"] > len(content):
                    truncated = " [truncated]"
                else:
                    truncated = ""

                lines.append(f"{line_num}: {content}{truncated}")
            else:
                lines.append(line_info["content"])

        return "\n".join(lines)


def main():
    parser = argparse.ArgumentParser(
        description="Read file contents with enhanced features and line range support."
    )
    parser.add_argument(
        "file_path", type=str, help="The absolute path of the file to read"
    )
    parser.add_argument(
        "start_line", type=int, help="The line number to start reading from (1-based)"
    )
    parser.add_argument(
        "end_line",
        type=int,
        help="The inclusive line number to end reading at (1-based, use -1 for end of file)",
    )
    parser.add_argument(
        "--encoding",
        type=str,
        help="The encoding to use for reading the file (default: auto-detect)",
    )
    parser.add_argument(
        "--show_line_numbers",
        action="store_true",
        default=True,
        help="Show line numbers in output (default: True)",
    )
    parser.add_argument(
        "--no_line_numbers",
        action="store_true",
        default=False,
        help="Hide line numbers in output",
    )
    parser.add_argument(
        "--highlight_syntax",
        action="store_true",
        default=False,
        help="Attempt to highlight syntax if possible (default: False)",
    )
    parser.add_argument(
        "--max_line_length",
        type=int,
        default=1000,
        help="Maximum line length before truncation (default: 1000)",
    )
    parser.add_argument(
        "--output_format",
        choices=["default", "json", "raw"],
        default="default",
        help="Output format (default: default)",
    )

    args = parser.parse_args()

    # Handle line numbers flag
    show_line_numbers = args.show_line_numbers and not args.no_line_numbers

    result = read_file(
        args.file_path,
        args.start_line,
        args.end_line,
        encoding=args.encoding,
        show_line_numbers=show_line_numbers,
        highlight_syntax=args.highlight_syntax,
        max_line_length=args.max_line_length,
    )

    output = format_output(result, args.output_format)
    print(output)

    # Exit with error code if reading failed
    sys.exit(0 if result["success"] else 1)


if __name__ == "__main__":
    main()
