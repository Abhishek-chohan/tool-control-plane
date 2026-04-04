#!/usr/bin/env python3
"""
Description: Search for a term in either a directory or a single file.

Behavior:
* If `--path` points to a directory (default is `.`), we recursively search all non-hidden files and directories.
* If `--path` points to a file, we run `grep -n` on that file to find line numbers containing the search term.
* If more than 100 files match (directory search scenario), the tool will stop listing and inform you to narrow your search.
* If no files are found that match your search term, the tool will inform you of that as well.

**Parameters:**
  1. **search_term** (`string`, required): The term to search for in files.
  2. **path** (`string`, optional): The file or directory in which to search. If not provided, defaults to the current directory (i.e., `.`).
"""

import argparse
import os
import subprocess
import sys


def search_in_directory(
    search_term: str, directory: str = ".", python_only: bool = False
):
    """
    Searches for `search_term` in all non-hidden files under `directory`
    (or only in .py files if `python_only=True`), excluding hidden directories.
    Prints how many matches were found per file.
    """
    directory = os.path.realpath(directory)

    if not os.path.isdir(directory):
        print(f"Directory '{directory}' not found or not a directory.")
        sys.exit(1)

    matches = {}
    num_files_matched = 0

    for root, dirs, files in os.walk(directory):
        # Exclude hidden directories
        dirs[:] = [d for d in dirs if not d.startswith(".")]
        for file in files:
            # Skip hidden files
            if file.startswith("."):
                continue

            # If --python_only is set, only search .py files
            if python_only and not file.endswith(".py"):
                continue

            filepath = os.path.join(root, file)
            try:
                with open(filepath, "r", errors="ignore") as f:
                    file_matches = 0
                    for line_num, line in enumerate(f, 1):
                        if search_term in line:
                            file_matches += 1
                    if file_matches > 0:
                        matches[filepath] = file_matches
                        num_files_matched += 1
            except (UnicodeDecodeError, PermissionError):
                # Skip files that can't be read
                continue

    if not matches:
        print(f'No matches found for "{search_term}" in {directory}')
        sys.exit(0)

    # Summarize
    num_matches = sum(matches.values())
    if num_files_matched > 100:
        print(
            f'More than {num_files_matched} files matched for "{search_term}" in {directory}. '
            "Please narrow your search."
        )
        sys.exit(0)

    print(f'Found {num_matches} matches for "{search_term}" in {directory}:')

    # Print matched files
    for filepath, count in matches.items():
        relative_path = os.path.relpath(filepath, start=os.getcwd())
        if not relative_path.startswith("./"):
            relative_path = "./" + relative_path
        print(f"{relative_path} ({count} matches)")

    print(f'End of matches for "{search_term}" in {directory}')


def search_in_directory_old(search_term: str, directory: str = ".", python_only=False):
    """
    Searches for `search_term` in all non-hidden files under `directory`,
    excluding hidden directories. Prints how many matches were found per file.
    """
    directory = os.path.realpath(directory)

    if not os.path.isdir(directory):
        print(f"Directory '{directory}' not found or not a directory.")
        sys.exit(1)

    matches = {}
    num_files_matched = 0

    for root, dirs, files in os.walk(directory):
        # Exclude hidden directories
        dirs[:] = [d for d in dirs if not d.startswith(".")]
        for file in files:
            # Skip hidden files
            if file.startswith("."):
                continue
            filepath = os.path.join(root, file)
            try:
                with open(filepath, "r", errors="ignore") as f:
                    file_matches = 0
                    for line_num, line in enumerate(f, 1):
                        if search_term in line:
                            file_matches += 1
                    if file_matches > 0:
                        matches[filepath] = file_matches
                        num_files_matched += 1
            except (UnicodeDecodeError, PermissionError):
                # Skip files that can't be read
                continue

    if not matches:
        print(f'No matches found for "{search_term}" in {directory}')
        sys.exit(0)

    # Summarize
    num_matches = sum(matches.values())
    if num_files_matched > 100:
        print(
            f'More than {num_files_matched} files matched for "{search_term}" in {directory}. '
            "Please narrow your search."
        )
        sys.exit(0)

    print(f'Found {num_matches} matches for "{search_term}" in {directory}:')

    # Print matched files
    for filepath, count in matches.items():
        # Convert absolute path to relative path
        relative_path = os.path.relpath(filepath, start=os.getcwd())
        if not relative_path.startswith("./"):
            relative_path = "./" + relative_path
        print(f"{relative_path} ({count} matches)")

    print(f'End of matches for "{search_term}" in {directory}')


def search_in_file(search_term: str, filepath: str):
    """
    Searches for `search_term` in a single file and prints matching lines with line numbers.
    Uses grep on Unix systems, falls back to pure Python search on Windows.
    """
    filepath = os.path.realpath(filepath)

    if not os.path.isfile(filepath):
        print(f"File '{filepath}' not found or is not a file.")
        sys.exit(1)

    # Try using grep first (Unix systems)
    try:
        try:
            # Try modern parameters if Python 3.7+ (capture_output, text)
            result = subprocess.run(
                ["grep", "-n", search_term, filepath], capture_output=True, text=True
            )
        except TypeError:
            # Fallback for Python 3.5/3.6
            result = subprocess.run(
                ["grep", "-n", search_term, filepath],
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                universal_newlines=True,
            )

        if result.returncode == 0:
            # grep found matches
            print(f'Matches for "{search_term}" in {filepath}:')
            print(result.stdout.strip())
            return
        elif result.returncode == 1:
            # grep found no matches
            print(f'No matches found for "{search_term}" in {filepath}')
            return
        else:
            # grep had an error, fall back to Python search
            raise subprocess.CalledProcessError(result.returncode, "grep")

    except (FileNotFoundError, subprocess.CalledProcessError):
        # grep not available or failed, use Python fallback
        pass

    # Fallback to pure Python search (cross-platform)
    try:
        matches_found = False
        with open(filepath, "r", errors="ignore") as f:
            for line_num, line in enumerate(f, 1):
                if search_term in line:
                    if not matches_found:
                        print(f'Matches for "{search_term}" in {filepath}:')
                        matches_found = True
                    # Format similar to grep output
                    print(f"{line_num}:{line.rstrip()}")

        if not matches_found:
            print(f'No matches found for "{search_term}" in {filepath}')

    except (UnicodeDecodeError, PermissionError) as e:
        print(f"Error reading file '{filepath}': {e}")
        sys.exit(1)
    # try:
    #     # Run grep -n <search_term> <filepath>
    #     result = subprocess.run(
    #         ["grep", "-n", search_term, filepath], capture_output=True, text=True
    #     )
    #     if result.returncode != 0:
    #         # grep exit code = 1 means no matches
    #         print(f'No matches found for "{search_term}" in {filepath}')
    #         sys.exit(0)
    #     # Print grep output directly
    #     print(f'Matches for "{search_term}" in {filepath}:')
    #     print(result.stdout.strip())
    # except FileNotFoundError:
    #     print(
    #         "`grep` is not available on this system. Please install or use another method."
    #     )
    #     sys.exit(1)


def main():
    parser = argparse.ArgumentParser(
        description="search tool: run subcommands such as `search` for files or directories."
    )
    parser.add_argument(
        "--search_term", help="Term to search for in files.", required=True
    )
    parser.add_argument(
        "--path",
        help="File or directory to search in (defaults to current dir).",
        default=".",
    )
    # NEW ARGUMENT:
    parser.add_argument(
        "--python_only",
        default=True,
        help="If set, only search for matches in .py files when searching a directory.",
    )

    args = parser.parse_args()
    # Check if path is a file or a directory
    if os.path.isfile(args.path):
        search_in_file(args.search_term, args.path)
    else:
        search_in_directory(args.search_term, args.path, python_only=args.python_only)


if __name__ == "__main__":
    main()
