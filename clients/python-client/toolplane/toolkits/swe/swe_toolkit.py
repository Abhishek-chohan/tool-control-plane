#!/usr/bin/env python3

import asyncio
import io
import os
import sys
from pathlib import Path
from typing import Any, Dict, List, Optional, Type, Union

from langchain.tools import BaseTool
from pydantic import BaseModel, Field

# Import descriptions from descriptions.py
try:
    from toolplane.toolkits.swe.descriptions import (
        _BASH_DESCRIPTION,
        _FILE_EDITOR_DESCRIPTION,
        _FINISH_DESCRIPTION,
        _SEARCH_DESCRIPTION,
        _SUBMIT_DESCRIPTION,
    )
except ImportError:
    # Fallback for when running as standalone
    from descriptions import (
        _BASH_DESCRIPTION,
        _FILE_EDITOR_DESCRIPTION,
        _FINISH_DESCRIPTION,
        _SEARCH_DESCRIPTION,
        _SUBMIT_DESCRIPTION,
    )

# Import functions from individual tool files
try:
    from toolplane.toolkits.swe.execute_bash import BLOCKED_BASH_COMMANDS, run_command
    from toolplane.toolkits.swe.finish import submit as finish_submit
    from toolplane.toolkits.swe.read_file import read_file
    from toolplane.toolkits.swe.search import search_in_directory, search_in_file
    from toolplane.toolkits.swe.str_replace_editor import (
        StrReplaceEditor,
        load_history,
        save_history,
    )
    from toolplane.toolkits.swe.submit import submit as simple_submit
except ImportError:
    # Fallback for when running as standalone
    from execute_bash import BLOCKED_BASH_COMMANDS, run_command
    from finish import submit as finish_submit
    from search import search_in_directory, search_in_file
    from str_replace_editor import StrReplaceEditor, load_history, save_history
    from submit import submit as simple_submit


#!/usr/bin/env python3
"""
Standalone Toolkit for LangChain Integration

This module provides a comprehensive set of standalone development tools
wrapped as LangChain tools for AI agent integration.
"""

import os
import sys
from pathlib import Path
from typing import Any, Dict, List, Optional, Type, Union

from langchain.tools import BaseTool
from langchain_core.callbacks.manager import (
    AsyncCallbackManagerForToolRun,
    CallbackManagerForToolRun,
)
from pydantic import BaseModel, Field

# Import functions from standalone tools
try:
    from toolplane.toolkits.swe.create_directory import create_directory
    from toolplane.toolkits.swe.create_file import create_file
    from toolplane.toolkits.swe.file_search import file_search
    from toolplane.toolkits.swe.grep_search import grep_search
    from toolplane.toolkits.swe.list_dir import list_dir
    from toolplane.toolkits.swe.replace_string_in_file import replace_string_in_file
    from toolplane.toolkits.swe.semantic_search import semantic_search
except ImportError:
    # Fallback for when running as standalone
    from create_directory import create_directory
    from create_file import create_file
    from file_search import file_search
    from grep_search import grep_search
    from list_dir import list_dir
    from replace_string_in_file import replace_string_in_file
    from semantic_search import semantic_search


# Input Models for each tool
class CreateDirectoryInput(BaseModel):
    dir_path: str = Field(description="The absolute path to the directory to create")
    mode: Optional[int] = Field(
        default=0o755, description="Permission mode for the directory"
    )
    parents: Optional[bool] = Field(
        default=True, description="Create parent directories if they don't exist"
    )
    exist_ok: Optional[bool] = Field(
        default=True, description="Don't raise error if directory already exists"
    )


class WriteFileInput(BaseModel):
    file_path: str = Field(description="The absolute path to the file to create")
    content: str = Field(description="The content to write to the file")
    encoding: Optional[str] = Field(
        default="utf-8", description="The encoding to use for the file"
    )
    mode: Optional[str] = Field(default="w", description="The file creation mode")
    auto_create_dirs: Optional[bool] = Field(
        default=True, description="Create parent directories if they don't exist"
    )
    overwrite: Optional[bool] = Field(
        default=False, description="Overwrite the file if it already exists"
    )


class FetchWebpageInput(BaseModel):
    urls: List[str] = Field(description="List of URLs to fetch content from")
    query: str = Field(description="The query to search for in the web page's content")
    timeout: Optional[int] = Field(default=30, description="Request timeout in seconds")
    max_content_length: Optional[int] = Field(
        default=50000, description="Maximum content length to process"
    )
    user_agent: Optional[str] = Field(
        default=None, description="Custom user agent string"
    )


class FileSearchInput(BaseModel):
    query: str = Field(description="Glob pattern to search for files")
    max_results: Optional[int] = Field(
        default=None, description="Maximum number of results to return"
    )
    include_hidden: Optional[bool] = Field(
        default=False, description="Include hidden files in results"
    )
    sort_by: Optional[str] = Field(
        default="name", description="Sort results by 'name', 'size', or 'modified'"
    )
    reverse_sort: Optional[bool] = Field(
        default=False, description="Reverse the sort order"
    )
    show_details: Optional[bool] = Field(
        default=False, description="Show file details like size and modification time"
    )
    base_path: Optional[str] = Field(
        default=None, description="Base directory to search from"
    )


class GrepSearchInput(BaseModel):
    query: str = Field(description="The pattern to search for in files")
    is_regexp: bool = Field(description="Whether the pattern is a regex")
    include_pattern: Optional[str] = Field(
        default=None, description="Search files matching this glob pattern"
    )
    max_results: Optional[int] = Field(
        default=None, description="Maximum number of results to return"
    )
    context_lines: Optional[int] = Field(
        default=0, description="Number of context lines to show around matches"
    )
    ignore_case: Optional[bool] = Field(
        default=False, description="Perform case-insensitive search"
    )
    whole_word: Optional[bool] = Field(
        default=False, description="Match whole words only"
    )
    invert_match: Optional[bool] = Field(
        default=False, description="Show lines that don't match the pattern"
    )
    base_path: Optional[str] = Field(
        default=None, description="Base directory to search from"
    )


class ListDirInput(BaseModel):
    path: Optional[str] = Field(
        default=".",
        description="Directory to list. Use '.' for current working directory. Accepts absolute or relative paths.",
    )
    show_hidden: Optional[bool] = Field(
        default=False, description="Show hidden files and directories"
    )
    show_details: Optional[bool] = Field(
        default=False, description="Show detailed information like size and permissions"
    )
    sort_by: Optional[str] = Field(
        default="name", description="Sort by 'name', 'size', 'modified', or 'type'"
    )
    reverse_sort: Optional[bool] = Field(
        default=False, description="Reverse the sort order"
    )
    recursive: Optional[bool] = Field(
        default=False, description="List contents recursively"
    )
    max_depth: Optional[int] = Field(
        default=3, description="Maximum depth for recursive listing"
    )
    file_filter: Optional[str] = Field(
        default=None, description="Filter files by extension"
    )


class ReadFileInput(BaseModel):
    file_path: str = Field(description="The absolute path of the file to read")
    start_line: int = Field(
        description="The line number to start reading from (1-based)"
    )
    end_line: int = Field(
        description="The inclusive line number to end reading at (1-based, -1 for end)"
    )
    encoding: Optional[str] = Field(
        default=None, description="The encoding to use for reading the file"
    )
    show_line_numbers: Optional[bool] = Field(
        default=True, description="Show line numbers in output"
    )
    highlight_syntax: Optional[bool] = Field(
        default=False, description="Attempt to highlight syntax"
    )
    max_line_length: Optional[int] = Field(
        default=1000, description="Maximum line length before truncation"
    )


class ReplaceStringInput(BaseModel):
    file_path: str = Field(description="The absolute path to the file to edit")
    old_string: str = Field(description="The string to be replaced")
    new_string: str = Field(description="The replacement string")
    create_backup: Optional[bool] = Field(
        default=True, description="Create a backup before editing"
    )
    dry_run: Optional[bool] = Field(
        default=False, description="Show what would be changed without making changes"
    )
    whole_word: Optional[bool] = Field(
        default=False, description="Only replace whole words"
    )
    ignore_case: Optional[bool] = Field(
        default=False, description="Perform case-insensitive matching"
    )
    max_replacements: Optional[int] = Field(
        default=None, description="Maximum number of replacements to make"
    )


class SemanticSearchInput(BaseModel):
    query: str = Field(description="The search query in natural language")
    max_results: Optional[int] = Field(
        default=10, description="Maximum number of results to return"
    )
    file_types: Optional[List[str]] = Field(
        default=None, description="File types to search in"
    )
    similarity_threshold: Optional[float] = Field(
        default=0.1, description="Minimum similarity score to include"
    )
    context_size: Optional[int] = Field(
        default=3, description="Number of lines of context around matches"
    )
    search_path: Optional[str] = Field(
        default=None, description="Directory to search in"
    )


class WebSearchInput(BaseModel):
    query: str = Field(description="The search query")
    max_results: Optional[int] = Field(
        default=10, description="Maximum number of results to return"
    )
    search_engine: Optional[str] = Field(
        default="duckduckgo", description="Search engine to use"
    )
    include_content: Optional[bool] = Field(
        default=False, description="Include page content in results"
    )
    content_length: Optional[int] = Field(
        default=1000, description="Maximum content length to extract"
    )
    filter_domain: Optional[str] = Field(
        default=None, description="Only include results from this domain"
    )


class TestFailureAnalysisInput(BaseModel):
    test_output: Optional[str] = Field(
        default=None, description="Path to test output file or direct test output"
    )
    test_framework: Optional[str] = Field(
        default=None, description="Test framework used"
    )
    verbose: Optional[bool] = Field(default=False, description="Show detailed analysis")
    suggest_fixes: Optional[bool] = Field(
        default=True, description="Suggest potential fixes"
    )
    group_by_type: Optional[bool] = Field(
        default=True, description="Group failures by error type"
    )


# Tool Classes
class CreateDirectoryTool(BaseTool):
    """📁 Create directories with enhanced features and permission handling."""

    name: str = "create_directory"
    description: str = """Create a directory structure with enhanced features.
    
    This tool creates directories recursively (like mkdir -p) and provides
    additional features like permission handling and validation.
    
    Best for:
    - Creating project directory structures
    - Setting up development environments
    - Organizing file systems
    - Batch directory creation"""

    args_schema: Type[BaseModel] = CreateDirectoryInput

    def _run(
        self,
        dir_path: str,
        mode: Optional[int] = 0o755,
        parents: Optional[bool] = True,
        exist_ok: Optional[bool] = True,
        run_manager: Optional[CallbackManagerForToolRun] = None,
    ) -> str:
        try:
            success = create_directory(
                dir_path=dir_path,
                mode=mode or 0o755,
                parents=parents if parents is not None else True,
                exist_ok=exist_ok if exist_ok is not None else True,
            )

            if success:
                return f"Directory created successfully: {dir_path}"
            else:
                return f"Failed to create directory: {dir_path}"

        except Exception as e:
            return f"Error creating directory: {str(e)}"


class WriteFileTool(BaseTool):
    """📄 Create files with specified content and automatic directory creation."""

    name: str = "write_file"
    description: str = """Write to a file or create a new file with the given content and automatically creates
    parent directories if they don't exist. It includes validation and error handling.
    
    Best for:
    - Creating new source files
    - Generating configuration files
    - Setting up project templates
    - Batch file creation"""

    args_schema: Type[BaseModel] = WriteFileInput

    def _run(
        self,
        file_path: str,
        content: str,
        encoding: Optional[str] = "utf-8",
        mode: Optional[str] = "w",
        auto_create_dirs: Optional[bool] = True,
        overwrite: Optional[bool] = False,
        run_manager: Optional[CallbackManagerForToolRun] = None,
    ) -> str:
        try:
            success = create_file(
                file_path=file_path,
                content=content,
                encoding=encoding or "utf-8",
                mode=mode or "w",
                auto_create_dirs=(
                    auto_create_dirs if auto_create_dirs is not None else True
                ),
                overwrite=overwrite if overwrite is not None else False,
            )

            if success:
                return f"File created successfully: {file_path}"
            else:
                return f"Failed to create file: {file_path}"

        except Exception as e:
            return f"Error creating file: {str(e)}"


class FileSearchTool(BaseTool):
    """🔍 Search for files using glob patterns with advanced features."""

    name: str = "file_search"
    description: str = """Search for files using glob patterns and provides additional
    filtering and sorting options. It returns file paths matching the pattern.
    
    Best for:
    - Finding files by pattern
    - Project file discovery
    - Build system file location
    - Code organization analysis"""

    args_schema: Type[BaseModel] = FileSearchInput

    def _run(
        self,
        query: str,
        max_results: Optional[int] = None,
        include_hidden: Optional[bool] = False,
        sort_by: Optional[str] = "name",
        reverse_sort: Optional[bool] = False,
        show_details: Optional[bool] = False,
        base_path: Optional[str] = None,
        run_manager: Optional[CallbackManagerForToolRun] = None,
    ) -> str:
        try:
            results = file_search(
                query=query,
                max_results=max_results,
                include_hidden=include_hidden or False,
                sort_by=sort_by or "name",
                reverse_sort=reverse_sort or False,
                show_details=show_details or False,
                base_path=base_path,
            )

            if not results:
                return f"No files found matching pattern: {query}"

            # Format results for display
            output = [f"Found {len(results)} files matching pattern: {query}"]

            if show_details:
                output.append("-" * 80)
                for result in results:
                    file_type = "DIR" if result["is_dir"] else "FILE"
                    size_str = f"{result['size']} bytes" if result["is_file"] else "-"
                    output.append(f"{result['name']:<40} {file_type:<5} {size_str}")
            else:
                for result in results:
                    suffix = "/" if result["is_dir"] else ""
                    output.append(f"{result['path']}{suffix}")

            return "\n".join(output)

        except Exception as e:
            return f"Error searching files: {str(e)}"


class GrepSearchTool(BaseTool):
    """🔎 Fast text search in files using grep-like functionality."""

    name: str = "grep_search"
    description: str = """Perform fast text search in files using exact strings or regex patterns.
    It provides context lines, file filtering, and other advanced search options.
    
    Best for:
    - Code pattern searching
    - Log file analysis
    - Documentation searching
    - Multi-file text analysis"""

    args_schema: Type[BaseModel] = GrepSearchInput

    def _run(
        self,
        query: str,
        is_regexp: bool,
        include_pattern: Optional[str] = None,
        max_results: Optional[int] = None,
        context_lines: Optional[int] = 0,
        ignore_case: Optional[bool] = False,
        whole_word: Optional[bool] = False,
        invert_match: Optional[bool] = False,
        base_path: Optional[str] = None,
        run_manager: Optional[CallbackManagerForToolRun] = None,
    ) -> str:
        try:
            results = grep_search(
                query=query,
                is_regexp=is_regexp,
                include_pattern=include_pattern,
                max_results=max_results,
                context_lines=context_lines or 0,
                ignore_case=ignore_case or False,
                whole_word=whole_word or False,
                invert_match=invert_match or False,
                base_path=base_path,
            )

            if not results:
                return f"No matches found for pattern: {query}"

            # Format results for display
            total_matches = sum(r["total_matches"] for r in results)
            output = [f"Found {total_matches} matches in {len(results)} files"]
            output.append("=" * 60)

            for result in results:
                if "error" in result:
                    output.append(f"Error in {result['file']}: {result['error']}")
                    continue

                output.append(f"\nFile: {result['file']}")
                output.append(f"Matches: {result['total_matches']}")
                output.append("-" * 40)

                for match in result["matches"][:5]:  # Show first 5 matches
                    if context_lines and match.get("context"):
                        for ctx in match["context"]:
                            prefix = ">" if ctx["is_match"] else " "
                            output.append(
                                f"{prefix}{ctx['line_number']:4d}: {ctx['content']}"
                            )
                    else:
                        output.append(f"{match['line_number']:4d}: {match['content']}")

            return "\n".join(output)

        except Exception as e:
            return f"Error searching with grep: {str(e)}"


class ListDirTool(BaseTool):
    """📂 List directory contents with enhanced features."""

    name: str = "list_dir"
    description: str = """List the contents of a directory with various display options,
    sorting capabilities, and filtering features.
    
    Best for:
    - Directory exploration
    - Project structure analysis
    - File system navigation
    - Content organization"""

    args_schema: Type[BaseModel] = ListDirInput

    def _run(
        self,
        path: Optional[str] = ".",
        show_hidden: Optional[bool] = False,
        show_details: Optional[bool] = False,
        sort_by: Optional[str] = "name",
        reverse_sort: Optional[bool] = False,
        recursive: Optional[bool] = False,
        max_depth: Optional[int] = 3,
        file_filter: Optional[str] = None,
        run_manager: Optional[CallbackManagerForToolRun] = None,
    ) -> str:
        try:
            path = path or "."
            normalized_path = os.path.abspath(path)
            results = list_dir(
                path=normalized_path,
                show_hidden=show_hidden or False,
                show_details=show_details or False,
                sort_by=sort_by or "name",
                reverse_sort=reverse_sort or False,
                recursive=recursive or False,
                max_depth=max_depth or 3,
                file_filter=file_filter,
            )

            if not results:
                return f"No contents found in directory: {path}"

            # Format results for display
            output = [f"Contents of {path}:"]

            if show_details:
                output.append(f"{'Name':<40} {'Type':<5} {'Size':<10}")
                output.append("-" * 60)

                for result in results:
                    file_type = "DIR" if result["is_dir"] else "FILE"
                    if result["is_symlink"]:
                        file_type = "LINK"

                    size_str = f"{result['size']} bytes" if result["is_file"] else "-"
                    indent = "  " * result.get("depth", 0)

                    output.append(
                        f"{indent}{result['name']:<40} {file_type:<5} {size_str}"
                    )
            else:
                for result in results:
                    suffix = "/" if result["is_dir"] else ""
                    indent = "  " * result.get("depth", 0)
                    output.append(f"{indent}{result['name']}{suffix}")

            return "\n".join(output)

        except Exception as e:
            return f"Error listing directory: {str(e)}"


class ReadFileTool(BaseTool):
    """📖 Read file contents with line range support and encoding detection."""

    name: str = "read_file"
    description: str = """Read the contents of a file with support for line ranges,
    encoding detection, and various output formats.
    
    Best for:
    - Code review and analysis
    - Configuration file inspection
    - Log file examination
    - Content verification"""

    args_schema: Type[BaseModel] = ReadFileInput

    def _run(
        self,
        file_path: str,
        start_line: int,
        end_line: int,
        encoding: Optional[str] = None,
        show_line_numbers: Optional[bool] = True,
        highlight_syntax: Optional[bool] = False,
        max_line_length: Optional[int] = 1000,
        run_manager: Optional[CallbackManagerForToolRun] = None,
    ) -> str:
        try:
            result = read_file(
                file_path=file_path,
                start_line=start_line,
                end_line=end_line,
                encoding=encoding,
                show_line_numbers=(
                    show_line_numbers if show_line_numbers is not None else True
                ),
                highlight_syntax=highlight_syntax or False,
                max_line_length=max_line_length or 1000,
            )

            if not result["success"]:
                return f"Error: {result['error']}"

            # Format output
            output = []
            output.append(f"File: {result['file_path']}")
            output.append(
                f"Lines: {result['start_line']}-{result['end_line']} (of {result['total_lines']})"
            )
            output.append(f"Encoding: {result['encoding']}")
            if result["file_language"]:
                output.append(f"Language: {result['file_language']}")
            output.append("-" * 60)

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

                    output.append(f"{line_num}: {content}{truncated}")
                else:
                    output.append(line_info["content"])

            return "\n".join(output)

        except Exception as e:
            return f"Error reading file: {str(e)}"


class ReplaceStringTool(BaseTool):
    """🔄 Replace strings in files with safety checks and backup options."""

    name: str = "replace_string"
    description: str = """Replace strings in files with validation, backup options,
    and various safety features to prevent accidental data loss.
    
    Best for:
    - Code refactoring
    - Configuration updates
    - Batch text replacements
    - Safe file modifications"""

    args_schema: Type[BaseModel] = ReplaceStringInput

    def _run(
        self,
        file_path: str,
        old_string: str,
        new_string: str,
        create_backup: Optional[bool] = True,
        dry_run: Optional[bool] = False,
        whole_word: Optional[bool] = False,
        ignore_case: Optional[bool] = False,
        max_replacements: Optional[int] = None,
        run_manager: Optional[CallbackManagerForToolRun] = None,
    ) -> str:
        try:
            result = replace_string_in_file(
                file_path=file_path,
                old_string=old_string,
                new_string=new_string,
                create_backup=create_backup if create_backup is not None else True,
                dry_run=dry_run or False,
                whole_word=whole_word or False,
                ignore_case=ignore_case or False,
                max_replacements=max_replacements,
            )

            if not result["success"]:
                return f"Error: {result['error']}"

            output = [result["message"]]

            if result.get("dry_run") and "changes_preview" in result:
                output.append("\nChanges preview:")
                for change in result["changes_preview"]:
                    output.append(
                        f"\nMatch {change['match_number']} at line {change['line_number']}:"
                    )
                    output.append(f"Before: {change['before']}")
                    output.append(f"After:  {change['after']}")

            if result.get("backup_created"):
                output.append(f"Backup created: {result['backup_path']}")

            return "\n".join(output)

        except Exception as e:
            return f"Error replacing string: {str(e)}"


class SemanticSearchTool(BaseTool):
    """🧠 Semantic search for relevant code using text similarity."""

    name: str = "semantic_search"
    description: str = """Perform semantic search across code and documentation files using
    text similarity algorithms to find relevant content based on natural language queries.
    
    Best for:
    - Finding relevant code snippets
    - Documentation discovery
    - Code understanding and navigation
    - Knowledge base search"""

    args_schema: Type[BaseModel] = SemanticSearchInput

    def _run(
        self,
        query: str,
        max_results: Optional[int] = 10,
        file_types: Optional[List[str]] = None,
        similarity_threshold: Optional[float] = 0.1,
        context_size: Optional[int] = 3,
        search_path: Optional[str] = None,
        run_manager: Optional[CallbackManagerForToolRun] = None,
    ) -> str:
        try:
            results = semantic_search(
                query=query,
                max_results=max_results or 10,
                file_types=file_types,
                similarity_threshold=similarity_threshold or 0.1,
                context_size=context_size or 3,
                search_path=search_path,
            )

            if not results:
                return f"No results found for query: {query}"

            # Format results for display
            output = [f"Semantic search results for: {query}"]
            output.append(f"Found {len(results)} results")
            output.append("=" * 80)

            for i, result in enumerate(results, 1):
                output.append(
                    f"\nResult {i}: {result['file_path']}:{result['line_number']}"
                )
                output.append(f"Similarity Score: {result['similarity_score']:.3f}")
                output.append(f"File Type: {result['file_type']}")

                if result.get("match_type") == "function_definition":
                    output.append(f"Function: {result.get('function_name', 'unknown')}")

                output.append("Context:")
                for context_line in result["context"]:
                    marker = ">>>" if context_line["is_match"] else "   "
                    output.append(
                        f"{marker} {context_line['line_number']:4d}: {context_line['content']}"
                    )

                output.append("-" * 80)

            return "\n".join(output)

        except Exception as e:
            return f"Error performing semantic search: {str(e)}"


def search_directory_for_term(search_term: str, directory: str = "."):
    """Wrapper function for search_dir functionality"""
    import os

    if not os.path.isdir(directory):
        return f"Directory {directory} not found"

    directory = os.path.realpath(directory)
    matches = {}
    num_files_matched = 0

    for root, dirs, files in os.walk(directory):
        # Exclude hidden directories
        dirs[:] = [d for d in dirs if not d.startswith(".")]
        for file in files:
            if file.startswith("."):
                continue  # Skip hidden files
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
                continue  # Skip files that can't be read

    if not matches:
        return f'No matches found for "{search_term}" in {directory}'

    num_matches = sum(matches.values())

    if num_files_matched > 100:
        return f'More than {num_files_matched} files matched for "{search_term}" in {directory}. Please narrow your search.'

    result = f'Found {num_matches} matches for "{search_term}" in {directory}:\n'

    for filepath, count in matches.items():
        # Replace leading path with './' for consistency
        relative_path = os.path.relpath(filepath, start=os.getcwd())
        if not relative_path.startswith("./"):
            relative_path = "./" + relative_path
        result += f"{relative_path} ({count} matches)\n"

    result += f'End of matches for "{search_term}" in {directory}'
    return result


class FileEditorInput(BaseModel):
    command: str = Field(
        description="The command to run. Allowed options are: `view`, `create`, `str_replace`, `insert`, `undo_edit`."
    )
    path: str = Field(
        description="Absolute path to file or directory, e.g. `/testbed/file.py` or `/testbed`."
    )
    file_text: Optional[str] = Field(
        default=None,
        description="Required for the `create` command, contains the content of the file to be created.",
    )
    old_str: Optional[str] = Field(
        default=None,
        description="Required for the `str_replace` command, specifies the string in `path` to replace.",
    )
    new_str: Optional[str] = Field(
        default=None,
        description="Optional for the `str_replace` command to specify the replacement string. Required for the `insert` command to specify the string to insert.",
    )
    insert_line: Optional[int] = Field(
        default=None,
        description="Required for the `insert` command. The `new_str` will be inserted AFTER the line specified.",
    )
    view_range: Optional[List[int]] = Field(
        default=None,
        description="Optional for the `view` command when `path` points to a file. Specifies the line range to view. E.g., [11, 12] shows lines 11 and 12. Indexing starts at 1. Use [start_line, -1] to show all lines from `start_line` to the end.",
    )
    enable_linting: bool = Field(
        default=False, description="Enable Python linting checks before saving changes"
    )


class SearchInput(BaseModel):
    search_term: str = Field(description="The term to search for in files.")
    path: Optional[str] = Field(
        default=".",
        description="The file or directory to search in. Defaults to `.` if not specified.",
    )
    python_only: bool = Field(default=False, description="Only search in Python files")


class SearchDirInput(BaseModel):
    # Keep existing descriptions as there's no direct match in descriptions.py
    search_term: str = Field(description="The term to search for")
    directory: Optional[str] = Field(
        default=".", description="The directory to search in"
    )


class FinishInput(BaseModel):
    result: str = Field(
        default="",
        description="Optional. The result text to submit. Defaults to an empty string if not provided.",
    )


class FileEditorTool(BaseTool):
    """📝 Advanced file editor with view, create, edit, and undo capabilities."""

    name: str = "file_editor"
    description: str = _FILE_EDITOR_DESCRIPTION

    args_schema: Type[BaseModel] = FileEditorInput

    def __init__(self, **kwargs):
        super().__init__(**kwargs)

    def _run(
        self,
        command: str,
        path: str,
        file_text: Optional[str] = None,
        old_str: Optional[str] = None,
        new_str: Optional[str] = None,
        insert_line: Optional[int] = None,
        view_range: Optional[List[int]] = None,
        enable_linting: bool = False,
        run_manager: Optional[CallbackManagerForToolRun] = None,
    ) -> str:
        try:
            # Load edit history and create editor instance
            history = load_history()
            editor = StrReplaceEditor(history, enable_linting)

            # Capture stdout to get the result
            old_stdout = sys.stdout
            sys.stdout = captured_output = io.StringIO()

            try:
                # Run the editor command
                result = editor.run(
                    command=command,
                    path_str=path,
                    file_text=file_text,
                    view_range=view_range,
                    old_str=old_str,
                    new_str=new_str,
                    insert_line=insert_line,
                )

                # Save updated history
                save_history(editor.file_history)

                # Return the result
                return str(result)

            finally:
                sys.stdout = old_stdout

        except Exception as e:
            return f"Error running file editor: {str(e)}"


class SearchTool(BaseTool):
    """🔍 Search for text within files and directories."""

    name: str = "search"
    description: str = _SEARCH_DESCRIPTION

    args_schema: Type[BaseModel] = SearchInput

    def __init__(self, **kwargs):
        super().__init__(**kwargs)

    def _run(
        self,
        search_term: str,
        path: str = ".",
        python_only: bool = False,
        run_manager: Optional[CallbackManagerForToolRun] = None,
    ) -> str:
        try:
            # Capture stdout to get the result
            old_stdout = sys.stdout
            sys.stdout = captured_output = io.StringIO()

            try:
                # Check if path is a file or directory
                if os.path.isfile(path):
                    search_in_file(search_term, path)
                else:
                    search_in_directory(search_term, path, python_only)

                # Get the captured output
                result = captured_output.getvalue()
                return result

            finally:
                sys.stdout = old_stdout

        except Exception as e:
            return f"Error running search: {str(e)}"


class SearchDirTool(BaseTool):
    """📁 Search for text within all files in a directory."""

    name: str = "search_dir"
    description: str = """Recursively search for text patterns in all files within a directory.
    
    Features:
    - Recursive directory traversal
    - Match counting per file
    - Excludes hidden files and directories
    - Performance optimized for large directories
    
    Best for:
    - Codebase-wide searches
    - Finding all occurrences of a pattern
    - Project-wide refactoring preparation
    - Code analysis and auditing"""

    args_schema: Type[BaseModel] = SearchDirInput

    def __init__(self, **kwargs):
        super().__init__(**kwargs)

    def _run(
        self,
        search_term: str,
        directory: str = ".",
        run_manager: Optional[CallbackManagerForToolRun] = None,
    ) -> str:
        try:
            return search_directory_for_term(search_term, directory)
        except Exception as e:
            return f"Error running search_dir: {str(e)}"


class BashInput(BaseModel):
    command: str = Field(
        description="The command (and optional arguments) to execute. For example: 'python my_script.py'"
    )


class BashTool(BaseTool):
    """⚡ Execute bash commands with security restrictions."""

    name: str = "execute_bash"
    description: str = _BASH_DESCRIPTION.format(PWD=os.getcwd())
    args_schema: Type[BaseModel] = BashInput

    def __init__(self, **kwargs):
        super().__init__(**kwargs)

    def _run(
        self,
        command: str,
        run_manager: Optional[CallbackManagerForToolRun] = None,
    ) -> str:
        try:
            # Check if the command is blocked
            first_token = command.strip().split()[0]
            if first_token in BLOCKED_BASH_COMMANDS:
                return f"Bash command '{first_token}' is not allowed. Please use a different command or tool."

            # Run the command using the imported function
            result = run_command(command)

            if result.returncode != 0:
                output = "Error executing command:\n"
                output += "[STDOUT]\n"
                output += result.stdout.strip() + "\n"
                output += "[STDERR]\n"
                output += result.stderr.strip()
                return output

            output = "[STDOUT]\n"
            output += result.stdout.strip() + "\n"
            output += "[STDERR]\n"
            output += result.stderr.strip()
            return output

        except Exception as e:
            return f"Error running bash command: {str(e)}"


class FinishTool(BaseTool):
    """✅ Submit results and finish tasks."""

    name: str = "finish"
    description: str = _FINISH_DESCRIPTION

    args_schema: Type[BaseModel] = FinishInput

    def __init__(self, **kwargs):
        super().__init__(**kwargs)

    def _run(
        self,
        result: str = "",
        run_manager: Optional[CallbackManagerForToolRun] = None,
    ) -> str:
        try:
            # Capture stdout to get the result
            old_stdout = sys.stdout
            sys.stdout = captured_output = io.StringIO()

            try:
                # Call the finish submit function
                finish_submit(result)

                # Get the captured output
                output = captured_output.getvalue()
                return output

            finally:
                sys.stdout = old_stdout

        except Exception as e:
            return f"Error running finish: {str(e)}"


class SubmitTool(BaseTool):
    """🎯 Simple task submission tool."""

    name: str = "submit"
    description: str = _SUBMIT_DESCRIPTION

    args_schema: Type[BaseModel] = BaseModel

    def __init__(self, **kwargs):
        super().__init__(**kwargs)

    def _run(
        self,
        run_manager: Optional[CallbackManagerForToolRun] = None,
    ) -> str:
        try:
            # Capture stdout to get the result
            old_stdout = sys.stdout
            sys.stdout = captured_output = io.StringIO()

            try:
                # Call the simple submit function
                simple_submit()

                # Get the captured output
                output = captured_output.getvalue()
                return output

            finally:
                sys.stdout = old_stdout

        except Exception as e:
            return f"Error running submit: {str(e)}"


def get_swe_toolkit() -> List[BaseTool]:
    """Get all SWE (Software Engineering) tools with proper initialization.

    Args:
        tools_dir: Path to the tools directory. If None, uses the current file's directory.
                  Note: This parameter is maintained for compatibility but is no longer used
                  since tools now import functions directly.

    Returns:
        List of initialized SWE tools ready for LangChain integration.
    """

    return [
        # FileEditorTool(),
        # SearchTool(),
        # SearchDirTool(),
        BashTool(),
        CreateDirectoryTool(),
        WriteFileTool(),
        FileSearchTool(),
        GrepSearchTool(),
        # ListDirTool(),
        ReadFileTool(),
        ReplaceStringTool(),
        SemanticSearchTool(),
    ]


def get_file_editor_only(tools_dir: Optional[str] = None) -> FileEditorTool:
    """Get only the file editor tool for focused file operations."""
    if tools_dir is None:
        tools_dir = os.path.dirname(os.path.abspath(__file__))
    return FileEditorTool(tools_dir=tools_dir)


def get_search_tools_only(tools_dir: Optional[str] = None) -> List[BaseTool]:
    """Get only the search-related tools."""
    if tools_dir is None:
        tools_dir = os.path.dirname(os.path.abspath(__file__))
    return [
        SearchTool(tools_dir=tools_dir),
        SearchDirTool(tools_dir=tools_dir),
    ]


# Tool selection guide for LLMs
SWE_TOOL_SELECTION_GUIDE = """
🛠️ SWE TOOLKIT SELECTION GUIDE FOR AI AGENTS:

1. **file_editor** 📝
   - Default choice for all file operations
   - View, create, edit, and undo file changes
   - Supports Python linting and syntax checking
   - Persistent edit history

2. **search** 🔍
   - Search within specific files or directories
   - Line number reporting for matches
   - Cross-platform grep/Python fallback
   - Use when you need to find specific patterns

3. **search_dir** 📁
   - Recursive directory-wide searches
   - Match counting per file
   - Use for codebase-wide pattern finding
   - Great for refactoring preparation

4. **execute_bash** ⚡
   - Run system commands and scripts
   - Security restrictions for safety
   - Cross-platform compatibility
   - NOT for interactive or long-running commands

5. **finish** ✅
   - Submit results and complete tasks
   - Optional result text submission
   - Use when task objectives are met

6. **submit** 🎯
   - Simple task completion signal
   - No parameters needed
   - Quick completion indicator

⚠️ IMPORTANT WORKFLOW:
1. Start with file_editor to view project structure
2. Use search/search_dir to find relevant code
3. Use file_editor to make changes
4. Use execute_bash to test changes
5. Use finish/submit to complete tasks

💡 BEST PRACTICES:
- Always view files before editing
- Use search to understand codebase first
- Test changes with execute_bash
- Use specific, unique strings for str_replace
- Include context in old_str for uniqueness
"""

# Export all tools and utilities
__all__ = [
    "FileEditorInput",
    "SearchInput",
    "SearchDirInput",
    "BashInput",
    "FinishInput",
    "FileEditorTool",
    "SearchTool",
    "SearchDirTool",
    "BashTool",
    "FinishTool",
    "SubmitTool",
    "get_swe_toolkit",
    "get_file_editor_only",
    "get_search_tools_only",
    "SWE_TOOL_SELECTION_GUIDE",
]
