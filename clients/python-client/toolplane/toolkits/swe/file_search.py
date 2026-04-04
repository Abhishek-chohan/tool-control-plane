#!/usr/bin/env python3
"""
Description: Search for files in the workspace using glob patterns with advanced features.

This tool searches for files using glob patterns and provides additional
filtering and sorting options. It returns file paths matching the pattern.

Parameters:
  query (string, required): Glob pattern to search for files.
  max_results (integer, optional): Maximum number of results to return.
  include_hidden (boolean, optional): Include hidden files in results (default: False).
  sort_by (string, optional): Sort results by 'name', 'size', 'modified' (default: 'name').
  reverse_sort (boolean, optional): Reverse the sort order (default: False).
  show_details (boolean, optional): Show file details like size and modification time (default: False).
"""

import argparse
import glob
import os
import sys
import time
from pathlib import Path
from typing import Any, Dict, List


def file_search(
    query: str,
    max_results: int = None,
    include_hidden: bool = False,
    sort_by: str = "name",
    reverse_sort: bool = False,
    show_details: bool = False,
    base_path: str = None,
) -> List[Dict[str, Any]]:
    """
    Search for files using glob patterns with advanced features.

    Args:
        query: Glob pattern to search for files
        max_results: Maximum number of results to return
        include_hidden: Include hidden files in results
        sort_by: Sort results by 'name', 'size', 'modified'
        reverse_sort: Reverse the sort order
        show_details: Show file details like size and modification time
        base_path: Base directory to search from (default: current directory)

    Returns:
        List of dictionaries containing file information
    """
    if base_path is None:
        base_path = os.getcwd()

    # Change to base directory for glob search
    original_cwd = os.getcwd()
    os.chdir(base_path)

    try:
        # Use glob to find matching files
        matches = glob.glob(query, recursive=True)

        results = []
        for match in matches:
            try:
                path = Path(match)
                absolute_path = path.resolve()

                # Skip hidden files if not requested
                if not include_hidden and any(
                    part.startswith(".") for part in path.parts
                ):
                    continue

                # Get file stats
                stat = absolute_path.stat()

                result = {
                    "path": str(absolute_path),
                    "relative_path": str(path),
                    "name": path.name,
                    "is_file": absolute_path.is_file(),
                    "is_dir": absolute_path.is_dir(),
                    "size": stat.st_size if absolute_path.is_file() else 0,
                    "modified": stat.st_mtime,
                    "modified_readable": time.strftime(
                        "%Y-%m-%d %H:%M:%S", time.localtime(stat.st_mtime)
                    ),
                }

                results.append(result)

            except (OSError, PermissionError) as e:
                # Skip files that can't be accessed
                continue

        # Sort results
        if sort_by == "name":
            results.sort(key=lambda x: x["name"].lower(), reverse=reverse_sort)
        elif sort_by == "size":
            results.sort(key=lambda x: x["size"], reverse=reverse_sort)
        elif sort_by == "modified":
            results.sort(key=lambda x: x["modified"], reverse=reverse_sort)

        # Limit results if specified
        if max_results and len(results) > max_results:
            results = results[:max_results]

        return results

    finally:
        # Restore original working directory
        os.chdir(original_cwd)


def format_size(size_bytes: int) -> str:
    """Format file size in human-readable format."""
    if size_bytes == 0:
        return "0 B"

    units = ["B", "KB", "MB", "GB", "TB"]
    i = 0
    while size_bytes >= 1024 and i < len(units) - 1:
        size_bytes /= 1024
        i += 1

    return f"{size_bytes:.1f} {units[i]}"


def main():
    parser = argparse.ArgumentParser(
        description="Search for files in the workspace using glob patterns with advanced features."
    )
    parser.add_argument(
        "query",
        type=str,
        help="Glob pattern to search for files (e.g., '*.py', '**/*.txt', 'src/**/*.js')",
    )
    parser.add_argument(
        "--max_results", type=int, help="Maximum number of results to return"
    )
    parser.add_argument(
        "--include_hidden",
        action="store_true",
        default=False,
        help="Include hidden files in results (default: False)",
    )
    parser.add_argument(
        "--sort_by",
        choices=["name", "size", "modified"],
        default="name",
        help="Sort results by 'name', 'size', or 'modified' (default: name)",
    )
    parser.add_argument(
        "--reverse_sort",
        action="store_true",
        default=False,
        help="Reverse the sort order (default: False)",
    )
    parser.add_argument(
        "--show_details",
        action="store_true",
        default=False,
        help="Show file details like size and modification time (default: False)",
    )
    parser.add_argument(
        "--base_path",
        type=str,
        help="Base directory to search from (default: current directory)",
    )
    parser.add_argument(
        "--output_format",
        choices=["list", "table", "json"],
        default="list",
        help="Output format (default: list)",
    )

    args = parser.parse_args()

    results = file_search(
        args.query,
        max_results=args.max_results,
        include_hidden=args.include_hidden,
        sort_by=args.sort_by,
        reverse_sort=args.reverse_sort,
        show_details=args.show_details,
        base_path=args.base_path,
    )

    if not results:
        print(f"No files found matching pattern: {args.query}")
        sys.exit(0)

    if args.output_format == "json":
        import json

        print(json.dumps(results, indent=2))

    elif args.output_format == "table" or args.show_details:
        print(f"Found {len(results)} files matching pattern: {args.query}")
        print("-" * 80)

        if args.show_details:
            print(f"{'Name':<40} {'Size':<12} {'Modified':<20} {'Type':<8}")
            print("-" * 80)

            for result in results:
                file_type = "DIR" if result["is_dir"] else "FILE"
                size_str = format_size(result["size"]) if result["is_file"] else "-"

                print(
                    f"{result['name']:<40} {size_str:<12} {result['modified_readable']:<20} {file_type:<8}"
                )
        else:
            for result in results:
                print(result["path"])

    else:  # list format
        print(f"Found {len(results)} files matching pattern: {args.query}")
        for result in results:
            print(result["path"])

    # Show truncation warning if results were limited
    if args.max_results and len(results) == args.max_results:
        print(
            f"\nResults truncated to {args.max_results} items. Use --max_results to see more."
        )


if __name__ == "__main__":
    main()
