#!/usr/bin/env python3
"""
Description: List directory contents with enhanced features.

This tool lists the contents of a directory with various display options,
sorting capabilities, and filtering features.

Parameters:
  path (string, required): The absolute path to the directory to list.
  show_hidden (boolean, optional): Show hidden files and directories (default: False).
  show_details (boolean, optional): Show detailed information like size and permissions (default: False).
  sort_by (string, optional): Sort by 'name', 'size', 'modified', 'type' (default: 'name').
  reverse_sort (boolean, optional): Reverse the sort order (default: False).
  recursive (boolean, optional): List contents recursively (default: False).
  max_depth (integer, optional): Maximum depth for recursive listing (default: 3).
  file_filter (string, optional): Filter files by extension (e.g., '.py', '.txt').
"""

import argparse
import os
import stat
import sys
import time
from pathlib import Path
from typing import Any, Dict, List, Optional


def list_dir(
    path: str,
    show_hidden: bool = False,
    show_details: bool = False,
    sort_by: str = "name",
    reverse_sort: bool = False,
    recursive: bool = False,
    max_depth: int = 3,
    file_filter: Optional[str] = None,
) -> List[Dict[str, Any]]:
    """
    List directory contents with enhanced features.

    Args:
        path: The directory path to list
        show_hidden: Show hidden files and directories
        show_details: Show detailed information
        sort_by: Sort by 'name', 'size', 'modified', 'type'
        reverse_sort: Reverse the sort order
        recursive: List contents recursively
        max_depth: Maximum depth for recursive listing
        file_filter: Filter files by extension

    Returns:
        List of dictionaries containing file/directory information
    """
    try:
        dir_path = Path(path)

        if not dir_path.exists():
            print(f"Error: Directory '{path}' does not exist.", file=sys.stderr)
            return []

        if not dir_path.is_dir():
            print(f"Error: '{path}' is not a directory.", file=sys.stderr)
            return []

        results = []

        base_dir = dir_path.resolve()
        if recursive:
            results = _list_recursive(
                dir_path, base_dir, show_hidden, max_depth, file_filter, current_depth=0
            )
        else:
            results = _list_single_dir(dir_path, base_dir, show_hidden, file_filter)

        # Add detailed information if requested
        if show_details:
            for result in results:
                _add_detailed_info(result)

        # Sort results
        results = _sort_results(results, sort_by, reverse_sort)

        return results

    except PermissionError:
        print(f"Error: Permission denied accessing '{path}'.", file=sys.stderr)
        return []
    except Exception as e:
        print(f"Error listing directory '{path}': {e}", file=sys.stderr)
        return []


def _list_single_dir(
    dir_path: Path, base_dir: Path, show_hidden: bool, file_filter: Optional[str]
) -> List[Dict[str, Any]]:
    """List contents of a single directory."""
    results = []

    try:
        for item in dir_path.iterdir():
            # Skip hidden files if not requested
            if not show_hidden and item.name.startswith("."):
                continue

            # Apply file filter
            if file_filter and item.is_file() and not item.name.endswith(file_filter):
                continue

            try:
                stat_info = item.stat()

                result = {
                    "name": item.name,
                    "path": str(item.resolve()),
                    "relative_path": str(item.resolve().relative_to(base_dir)),
                    "is_file": item.is_file(),
                    "is_dir": item.is_dir(),
                    "is_symlink": item.is_symlink(),
                    "size": stat_info.st_size if item.is_file() else 0,
                    "modified": stat_info.st_mtime,
                    "modified_readable": time.strftime(
                        "%Y-%m-%d %H:%M:%S", time.localtime(stat_info.st_mtime)
                    ),
                    "permissions": stat.filemode(stat_info.st_mode),
                    "depth": 0,
                }

                results.append(result)

            except (OSError, PermissionError):
                # Skip items that can't be accessed
                continue

    except PermissionError:
        pass

    return results


def _list_recursive(
    dir_path: Path,
    base_dir: Path,
    show_hidden: bool,
    max_depth: int,
    file_filter: Optional[str],
    current_depth: int = 0,
) -> List[Dict[str, Any]]:
    """List directory contents recursively.

    Includes items at depth == max_depth but does not recurse deeper.
    """
    results = []

    try:
        for item in dir_path.iterdir():
            # Skip hidden files if not requested
            if not show_hidden and item.name.startswith("."):
                continue

            # Apply file filter
            if file_filter and item.is_file() and not item.name.endswith(file_filter):
                continue

            try:
                stat_info = item.stat()

                result = {
                    "name": item.name,
                    "path": str(item.resolve()),
                    "relative_path": str(item.resolve().relative_to(base_dir)),
                    "is_file": item.is_file(),
                    "is_dir": item.is_dir(),
                    "is_symlink": item.is_symlink(),
                    "size": stat_info.st_size if item.is_file() else 0,
                    "modified": stat_info.st_mtime,
                    "modified_readable": time.strftime(
                        "%Y-%m-%d %H:%M:%S", time.localtime(stat_info.st_mtime)
                    ),
                    "permissions": stat.filemode(stat_info.st_mode),
                    "depth": current_depth,
                }

                results.append(result)

                # Recurse into subdirectories if we haven't reached max_depth
                if (
                    item.is_dir()
                    and not item.is_symlink()
                    and current_depth < max_depth
                ):
                    sub_results = _list_recursive(
                        item,
                        base_dir,
                        show_hidden,
                        max_depth,
                        file_filter,
                        current_depth + 1,
                    )
                    results.extend(sub_results)

            except (OSError, PermissionError):
                # Skip items that can't be accessed
                continue

    except PermissionError:
        pass

    return results


def _add_detailed_info(result: Dict[str, Any]) -> None:
    """Add detailed information to a result entry."""
    try:
        path = Path(result["path"])
        stat_info = path.stat()

        result.update(
            {
                "owner_readable": bool(stat_info.st_mode & stat.S_IRUSR),
                "owner_writable": bool(stat_info.st_mode & stat.S_IWUSR),
                "owner_executable": bool(stat_info.st_mode & stat.S_IXUSR),
                "group_readable": bool(stat_info.st_mode & stat.S_IRGRP),
                "group_writable": bool(stat_info.st_mode & stat.S_IWGRP),
                "group_executable": bool(stat_info.st_mode & stat.S_IXGRP),
                "other_readable": bool(stat_info.st_mode & stat.S_IROTH),
                "other_writable": bool(stat_info.st_mode & stat.S_IWOTH),
                "other_executable": bool(stat_info.st_mode & stat.S_IXOTH),
                "inode": stat_info.st_ino,
                "device": stat_info.st_dev,
                "nlink": stat_info.st_nlink,
                "uid": stat_info.st_uid,
                "gid": stat_info.st_gid,
                "accessed": stat_info.st_atime,
                "accessed_readable": time.strftime(
                    "%Y-%m-%d %H:%M:%S", time.localtime(stat_info.st_atime)
                ),
                "created": getattr(stat_info, "st_birthtime", stat_info.st_ctime),
                "created_readable": time.strftime(
                    "%Y-%m-%d %H:%M:%S",
                    time.localtime(
                        getattr(stat_info, "st_birthtime", stat_info.st_ctime)
                    ),
                ),
            }
        )
    except (OSError, PermissionError):
        pass


def _sort_results(
    results: List[Dict[str, Any]], sort_by: str, reverse_sort: bool
) -> List[Dict[str, Any]]:
    """Sort results by the specified criteria."""
    if sort_by == "name":
        results.sort(key=lambda x: x["name"].lower(), reverse=reverse_sort)
    elif sort_by == "size":
        results.sort(key=lambda x: x["size"], reverse=reverse_sort)
    elif sort_by == "modified":
        results.sort(key=lambda x: x["modified"], reverse=reverse_sort)
    elif sort_by == "type":
        # Sort by type (directories first, then files)
        results.sort(
            key=lambda x: (not x["is_dir"], x["name"].lower()), reverse=reverse_sort
        )

    return results


def format_size(size_bytes: int) -> str:
    """Format file size in human-readable format."""
    if size_bytes == 0:
        return "0 B"

    units = ["B", "KB", "MB", "GB", "TB"]
    i = 0
    size = float(size_bytes)
    while size >= 1024.0 and i < len(units) - 1:
        size /= 1024.0
        i += 1

    return f"{size:.1f} {units[i]}"


def main():
    parser = argparse.ArgumentParser(
        description="List directory contents with enhanced features."
    )
    parser.add_argument(
        "path", type=str, help="The absolute path to the directory to list"
    )
    parser.add_argument(
        "--show_hidden",
        action="store_true",
        default=False,
        help="Show hidden files and directories (default: False)",
    )
    parser.add_argument(
        "--show_details",
        action="store_true",
        default=False,
        help="Show detailed information like size and permissions (default: False)",
    )
    parser.add_argument(
        "--sort_by",
        choices=["name", "size", "modified", "type"],
        default="name",
        help="Sort by 'name', 'size', 'modified', or 'type' (default: name)",
    )
    parser.add_argument(
        "--reverse_sort",
        action="store_true",
        default=False,
        help="Reverse the sort order (default: False)",
    )
    parser.add_argument(
        "--recursive",
        action="store_true",
        default=False,
        help="List contents recursively (default: False)",
    )
    parser.add_argument(
        "--max_depth",
        type=int,
        default=3,
        help="Maximum depth for recursive listing (default: 3)",
    )
    parser.add_argument(
        "--file_filter",
        type=str,
        help="Filter files by extension (e.g., '.py', '.txt')",
    )
    parser.add_argument(
        "--output_format",
        choices=["list", "table", "json", "tree"],
        default="list",
        help="Output format (default: list)",
    )

    args = parser.parse_args()

    results = list_dir(
        args.path,
        show_hidden=args.show_hidden,
        show_details=args.show_details,
        sort_by=args.sort_by,
        reverse_sort=args.reverse_sort,
        recursive=args.recursive,
        max_depth=args.max_depth,
        file_filter=args.file_filter,
    )

    if not results:
        sys.exit(1)

    if args.output_format == "json":
        import json

        print(json.dumps(results, indent=2))

    elif args.output_format == "table":
        if args.show_details:
            print(
                f"{'Name':<30} {'Type':<5} {'Size':<10} {'Permissions':<12} {'Modified':<20}"
            )
            print("-" * 90)

            for result in results:
                file_type = "DIR" if result["is_dir"] else "FILE"
                if result["is_symlink"]:
                    file_type = "LINK"

                size_str = format_size(result["size"]) if result["is_file"] else "-"
                indent = "  " * result.get("depth", 0)

                print(
                    f"{indent}{result['name']:<30} {file_type:<5} {size_str:<10} {result['permissions']:<12} {result['modified_readable']:<20}"
                )
        else:
            print(f"{'Name':<40} {'Type':<5} {'Size':<10}")
            print("-" * 60)

            for result in results:
                file_type = "DIR" if result["is_dir"] else "FILE"
                if result["is_symlink"]:
                    file_type = "LINK"

                size_str = format_size(result["size"]) if result["is_file"] else "-"
                indent = "  " * result.get("depth", 0)

                print(f"{indent}{result['name']:<40} {file_type:<5} {size_str:<10}")

    elif args.output_format == "tree":
        for result in results:
            indent = "  " * result.get("depth", 0)
            marker = "├── " if result.get("depth", 0) > 0 else ""
            suffix = "/" if result["is_dir"] else ""

            print(f"{indent}{marker}{result['name']}{suffix}")

    else:  # list format
        for result in results:
            suffix = "/" if result["is_dir"] else ""
            print(f"{result['name']}{suffix}")


if __name__ == "__main__":
    main()
