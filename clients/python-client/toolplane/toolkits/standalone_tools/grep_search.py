#!/usr/bin/env python3
"""
Description: Fast text search in files using grep-like functionality with advanced features.

This tool performs fast text search in files using exact strings or regex patterns.
It provides context lines, file filtering, and other advanced search options.

Parameters:
  query (string, required): The pattern to search for in files.
  is_regexp (boolean, required): Whether the pattern is a regex.
  include_pattern (string, optional): Search files matching this glob pattern.
  max_results (integer, optional): Maximum number of results to return.
  context_lines (integer, optional): Number of context lines to show around matches.
  ignore_case (boolean, optional): Perform case-insensitive search.
  whole_word (boolean, optional): Match whole words only.
  invert_match (boolean, optional): Show lines that don't match the pattern.
"""

import argparse
import glob
import os
import re
import sys
from pathlib import Path
from typing import Any, Dict, Iterator, List


def grep_search(
    query: str,
    is_regexp: bool,
    include_pattern: str = None,
    max_results: int = None,
    context_lines: int = 0,
    ignore_case: bool = False,
    whole_word: bool = False,
    invert_match: bool = False,
    base_path: str = None,
) -> List[Dict[str, Any]]:
    """
    Fast text search in files using grep-like functionality.

    Args:
        query: The pattern to search for
        is_regexp: Whether the pattern is a regex
        include_pattern: Search files matching this glob pattern
        max_results: Maximum number of results to return
        context_lines: Number of context lines to show around matches
        ignore_case: Perform case-insensitive search
        whole_word: Match whole words only
        invert_match: Show lines that don't match the pattern
        base_path: Base directory to search from

    Returns:
        List of dictionaries containing search results
    """
    if base_path is None:
        base_path = os.getcwd()

    # Prepare the search pattern
    if is_regexp:
        pattern = query
    else:
        pattern = re.escape(query)

    if whole_word:
        pattern = r"\b" + pattern + r"\b"

    flags = re.IGNORECASE if ignore_case else 0
    try:
        compiled_pattern = re.compile(pattern, flags)
    except re.error as e:
        print(f"Invalid regex pattern: {e}", file=sys.stderr)
        return []

    # Get list of files to search
    if include_pattern:
        files_to_search = get_files_by_pattern(include_pattern, base_path)
    else:
        files_to_search = get_all_text_files(base_path)

    results = []
    total_matches = 0

    for file_path in files_to_search:
        if max_results and total_matches >= max_results:
            break

        try:
            file_results = search_file(
                file_path, compiled_pattern, context_lines, invert_match
            )

            if file_results["matches"]:
                results.append(file_results)
                total_matches += len(file_results["matches"])

        except Exception as e:
            # Skip files that can't be read
            continue

    return results


def get_files_by_pattern(pattern: str, base_path: str) -> List[str]:
    """Get list of files matching the glob pattern."""
    original_cwd = os.getcwd()
    os.chdir(base_path)

    try:
        matches = glob.glob(pattern, recursive=True)
        files = []

        for match in matches:
            path = Path(match)
            if path.is_file():
                files.append(str(path.resolve()))

        return files

    finally:
        os.chdir(original_cwd)


def get_all_text_files(base_path: str) -> List[str]:
    """Get all text files in the directory tree."""
    text_extensions = {
        ".txt",
        ".py",
        ".js",
        ".ts",
        ".jsx",
        ".tsx",
        ".java",
        ".cpp",
        ".c",
        ".h",
        ".cs",
        ".php",
        ".rb",
        ".go",
        ".rs",
        ".swift",
        ".kt",
        ".scala",
        ".r",
        ".sql",
        ".html",
        ".htm",
        ".css",
        ".scss",
        ".sass",
        ".less",
        ".xml",
        ".json",
        ".yaml",
        ".yml",
        ".toml",
        ".ini",
        ".cfg",
        ".conf",
        ".md",
        ".rst",
        ".tex",
        ".sh",
        ".bash",
        ".zsh",
        ".fish",
        ".ps1",
        ".bat",
        ".dockerfile",
        ".makefile",
        ".cmake",
        ".gradle",
        ".properties",
    }

    files = []
    for root, dirs, file_names in os.walk(base_path):
        # Skip hidden directories
        dirs[:] = [d for d in dirs if not d.startswith(".")]

        for filename in file_names:
            if filename.startswith("."):
                continue

            path = Path(root) / filename

            # Check extension
            if path.suffix.lower() in text_extensions:
                files.append(str(path.resolve()))
            # Check if it's a text file by reading first few bytes
            elif not path.suffix and is_text_file(path):
                files.append(str(path.resolve()))

    return files


def is_text_file(file_path: Path) -> bool:
    """Check if a file is a text file by examining its content."""
    try:
        with open(file_path, "rb") as f:
            chunk = f.read(1024)

        # Check for null bytes (binary files usually contain them)
        if b"\x00" in chunk:
            return False

        # Try to decode as UTF-8
        try:
            chunk.decode("utf-8")
            return True
        except UnicodeDecodeError:
            return False

    except (IOError, OSError):
        return False


def search_file(
    file_path: str, pattern: re.Pattern, context_lines: int, invert_match: bool
) -> Dict[str, Any]:
    """Search for pattern in a single file."""
    try:
        with open(file_path, "r", encoding="utf-8", errors="ignore") as f:
            lines = f.readlines()
    except Exception as e:
        return {"file": file_path, "error": str(e), "matches": []}

    matches = []

    for line_num, line in enumerate(lines, 1):
        line = line.rstrip("\n\r")

        if invert_match:
            match = not pattern.search(line)
        else:
            match = pattern.search(line)

        if match:
            # Get context lines
            start_line = max(0, line_num - context_lines - 1)
            end_line = min(len(lines), line_num + context_lines)

            context = []
            for i in range(start_line, end_line):
                context_line = lines[i].rstrip("\n\r")
                context.append(
                    {
                        "line_number": i + 1,
                        "content": context_line,
                        "is_match": i + 1 == line_num,
                    }
                )

            matches.append(
                {
                    "line_number": line_num,
                    "content": line,
                    "context": context if context_lines > 0 else None,
                }
            )

    return {"file": file_path, "matches": matches, "total_matches": len(matches)}


def main():
    parser = argparse.ArgumentParser(
        description="Fast text search in files using grep-like functionality with advanced features."
    )
    parser.add_argument("query", type=str, help="The pattern to search for in files")
    parser.add_argument(
        "--is_regexp",
        action="store_true",
        default=False,
        help="Whether the pattern is a regex",
    )
    parser.add_argument(
        "--include_pattern", type=str, help="Search files matching this glob pattern"
    )
    parser.add_argument(
        "--max_results", type=int, help="Maximum number of results to return"
    )
    parser.add_argument(
        "--context_lines",
        type=int,
        default=0,
        help="Number of context lines to show around matches",
    )
    parser.add_argument(
        "--ignore_case",
        action="store_true",
        default=False,
        help="Perform case-insensitive search",
    )
    parser.add_argument(
        "--whole_word",
        action="store_true",
        default=False,
        help="Match whole words only",
    )
    parser.add_argument(
        "--invert_match",
        action="store_true",
        default=False,
        help="Show lines that don't match the pattern",
    )
    parser.add_argument(
        "--base_path",
        type=str,
        help="Base directory to search from (default: current directory)",
    )
    parser.add_argument(
        "--output_format",
        choices=["default", "json", "count"],
        default="default",
        help="Output format (default: default)",
    )

    args = parser.parse_args()

    results = grep_search(
        args.query,
        args.is_regexp,
        include_pattern=args.include_pattern,
        max_results=args.max_results,
        context_lines=args.context_lines,
        ignore_case=args.ignore_case,
        whole_word=args.whole_word,
        invert_match=args.invert_match,
        base_path=args.base_path,
    )

    if not results:
        print(f"No matches found for pattern: {args.query}")
        sys.exit(0)

    if args.output_format == "json":
        import json

        print(json.dumps(results, indent=2))

    elif args.output_format == "count":
        total_matches = sum(r["total_matches"] for r in results)
        print(f"Total matches: {total_matches}")
        print(f"Files with matches: {len(results)}")

        for result in results:
            print(f"{result['file']}: {result['total_matches']} matches")

    else:  # default format
        total_matches = sum(r["total_matches"] for r in results)
        print(f"Found {total_matches} matches in {len(results)} files")
        print("=" * 60)

        for result in results:
            if "error" in result:
                print(f"Error in {result['file']}: {result['error']}")
                continue

            print(f"\nFile: {result['file']}")
            print(f"Matches: {result['total_matches']}")
            print("-" * 40)

            for match in result["matches"]:
                if args.context_lines > 0 and match["context"]:
                    for ctx in match["context"]:
                        prefix = ">" if ctx["is_match"] else " "
                        print(f"{prefix}{ctx['line_number']:4d}: {ctx['content']}")
                else:
                    print(f"{match['line_number']:4d}: {match['content']}")


if __name__ == "__main__":
    main()
