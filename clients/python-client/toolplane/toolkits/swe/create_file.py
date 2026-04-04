#!/usr/bin/env python3
"""
Description: Create a new file with specified content and enhanced features.

This tool creates a new file with the given content and automatically creates
parent directories if they don't exist. It includes validation and error handling.

Parameters:
  filePath (string, required): The absolute path to the file to create.
  content (string, required): The content to write to the file.
  encoding (string, optional): The encoding to use for the file (default: utf-8).
  mode (string, optional): The file creation mode (default: 'w').
  auto_create_dirs (boolean, optional): Create parent directories if they don't exist (default: True).
  overwrite (boolean, optional): Overwrite the file if it already exists (default: False).
"""

import argparse
import os
import sys
from pathlib import Path


def create_file(
    file_path: str,
    content: str,
    encoding: str = "utf-8",
    mode: str = "w",
    auto_create_dirs: bool = True,
    overwrite: bool = False,
) -> bool:
    """
    Create a new file with the specified content.

    Args:
        file_path: The absolute path to the file to create
        content: The content to write to the file
        encoding: The encoding to use for the file
        mode: The file creation mode
        auto_create_dirs: Create parent directories if they don't exist
        overwrite: Overwrite the file if it already exists

    Returns:
        bool: True if file was created successfully, False otherwise
    """
    try:
        path = Path(file_path)

        # Check if file already exists
        if path.exists() and not overwrite:
            print(f"File already exists: {file_path}")
            print("Use --overwrite flag to overwrite existing files.")
            return False

        # Create parent directories if needed
        if auto_create_dirs and not path.parent.exists():
            path.parent.mkdir(parents=True, exist_ok=True)
            print(f"Created parent directories for: {file_path}")

        # Write the file
        with open(path, mode, encoding=encoding) as f:
            f.write(content)

        print(f"File created successfully: {file_path}")
        print(f"Content length: {len(content)} characters")
        return True

    except PermissionError:
        print(f"Permission denied: Cannot create file {file_path}")
        return False
    except FileNotFoundError:
        print(f"Directory does not exist: {path.parent}")
        print("Use --auto_create_dirs flag to create parent directories.")
        return False
    except Exception as e:
        print(f"Error creating file {file_path}: {e}")
        return False


def main():
    parser = argparse.ArgumentParser(
        description="Create a new file with specified content and enhanced features."
    )
    parser.add_argument(
        "filePath", type=str, help="The absolute path to the file to create."
    )
    parser.add_argument("content", type=str, help="The content to write to the file.")
    parser.add_argument(
        "--encoding",
        type=str,
        default="utf-8",
        help="The encoding to use for the file (default: utf-8)",
    )
    parser.add_argument(
        "--mode", type=str, default="w", help="The file creation mode (default: 'w')"
    )
    parser.add_argument(
        "--auto_create_dirs",
        action="store_true",
        default=True,
        help="Create parent directories if they don't exist (default: True)",
    )
    parser.add_argument(
        "--overwrite",
        action="store_true",
        default=False,
        help="Overwrite the file if it already exists (default: False)",
    )

    args = parser.parse_args()

    success = create_file(
        args.filePath,
        args.content,
        encoding=args.encoding,
        mode=args.mode,
        auto_create_dirs=args.auto_create_dirs,
        overwrite=args.overwrite,
    )

    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
