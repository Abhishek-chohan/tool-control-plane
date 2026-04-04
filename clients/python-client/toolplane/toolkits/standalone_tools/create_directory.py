#!/usr/bin/env python3
"""
Description: Create a new directory structure with enhanced features.

This tool creates directories recursively (like mkdir -p) and provides
additional features like permission handling and validation.

Parameters:
  dirPath (string, required): The absolute path to the directory to create.
  mode (integer, optional): The permission mode for the directory (default: 0o755).
  parents (boolean, optional): Create parent directories if they don't exist (default: True).
  exist_ok (boolean, optional): Don't raise an error if the directory already exists (default: True).
"""

import argparse
import os
import sys
from pathlib import Path


def create_directory(
    dir_path: str, mode: int = 0o755, parents: bool = True, exist_ok: bool = True
) -> bool:
    """
    Create a directory with the specified parameters.

    Args:
        dir_path: The absolute path to the directory to create
        mode: Permission mode for the directory
        parents: Create parent directories if they don't exist
        exist_ok: Don't raise error if directory already exists

    Returns:
        bool: True if directory was created or already exists, False otherwise
    """
    try:
        path = Path(dir_path)
        path.mkdir(mode=mode, parents=parents, exist_ok=exist_ok)

        if path.exists():
            print(f"Directory created successfully: {dir_path}")
            return True
        else:
            print(f"Failed to create directory: {dir_path}")
            return False

    except PermissionError:
        print(f"Permission denied: Cannot create directory {dir_path}")
        return False
    except FileExistsError:
        print(f"Directory already exists: {dir_path}")
        return exist_ok
    except Exception as e:
        print(f"Error creating directory {dir_path}: {e}")
        return False


def main():
    parser = argparse.ArgumentParser(
        description="Create a new directory structure with enhanced features."
    )
    parser.add_argument(
        "dirPath", type=str, help="The absolute path to the directory to create."
    )
    parser.add_argument(
        "--mode",
        type=lambda x: int(x, 8),  # Parse as octal
        default=0o755,
        help="Permission mode for the directory (octal, default: 755)",
    )
    parser.add_argument(
        "--parents",
        action="store_true",
        default=True,
        help="Create parent directories if they don't exist (default: True)",
    )
    parser.add_argument(
        "--exist_ok",
        action="store_true",
        default=True,
        help="Don't raise error if directory already exists (default: True)",
    )

    args = parser.parse_args()

    success = create_directory(
        args.dirPath, mode=args.mode, parents=args.parents, exist_ok=args.exist_ok
    )

    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
