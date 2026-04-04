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
    from .create_directory import create_directory
    from .create_file import create_file
    from .file_search import file_search
    from .grep_search import grep_search
    from .list_dir import list_dir
    from .read_file import read_file
    from .replace_string_in_file import replace_string_in_file
    from .semantic_search import semantic_search
    from .test_failure_analysis import analyze_test_failures
except ImportError:
    # Fallback for when running as standalone
    from standalone_tools.create_directory import create_directory
    from standalone_tools.create_file import create_file
    from standalone_tools.file_search import file_search
    from standalone_tools.grep_search import grep_search
    from standalone_tools.list_dir import list_dir
    from standalone_tools.read_file import read_file
    from standalone_tools.replace_string_in_file import replace_string_in_file
    from standalone_tools.semantic_search import semantic_search
    from standalone_tools.test_failure_analysis import analyze_test_failures


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


class CreateFileInput(BaseModel):
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
    path: str = Field(description="The absolute path to the directory to list")
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


class CreateFileTool(BaseTool):
    """📄 Create files with specified content and automatic directory creation."""

    name: str = "create_file"
    description: str = """Create a new file with the given content and automatically creates
    parent directories if they don't exist. It includes validation and error handling.
    
    Best for:
    - Creating new source files
    - Generating configuration files
    - Setting up project templates
    - Batch file creation"""

    args_schema: Type[BaseModel] = CreateFileInput

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
        path: str,
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
            results = list_dir(
                path=path,
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


class TestFailureAnalysisTool(BaseTool):
    """🔧 Test failure analysis tool for examining and debugging test failures."""

    name: str = "test_failure_analysis"
    description: str = """Analyze test failures, provide detailed error analysis, and suggest
    potential fixes based on common failure patterns.
    
    Best for:
    - Debugging test failures
    - Test maintenance
    - Error pattern recognition
    - Development workflow optimization"""

    args_schema: Type[BaseModel] = TestFailureAnalysisInput

    def _run(
        self,
        test_output: Optional[str] = None,
        test_framework: Optional[str] = None,
        verbose: Optional[bool] = False,
        suggest_fixes: Optional[bool] = True,
        group_by_type: Optional[bool] = True,
        run_manager: Optional[CallbackManagerForToolRun] = None,
    ) -> str:
        try:
            analysis = analyze_test_failures(
                test_output=test_output,
                test_framework=test_framework,
                verbose=verbose or False,
                suggest_fixes=suggest_fixes if suggest_fixes is not None else True,
                group_by_type=group_by_type if group_by_type is not None else True,
            )

            # Format analysis results
            output = []

            # Header
            output.append("TEST FAILURE ANALYSIS")
            output.append("=" * 50)

            # Summary
            if "summary" in analysis:
                summary = analysis["summary"]
                output.append(f"Total failures: {summary['total_failures']}")
                output.append(f"Unique tests: {summary['unique_tests']}")
                output.append(f"Unique files: {summary['unique_files']}")
                output.append(f"Most common error: {summary['most_common_error']}")
                output.append("")

            # Error groups
            if "error_groups" in analysis:
                output.append("ERROR GROUPS:")
                output.append("-" * 30)
                for error_type, group_failures in analysis["error_groups"].items():
                    output.append(f"{error_type}: {len(group_failures)} failures")
                    for failure in group_failures[:3]:  # Show first 3
                        output.append(f"  - {failure['test_name']}")
                    if len(group_failures) > 3:
                        output.append(f"  ... and {len(group_failures) - 3} more")
                output.append("")

            # Suggestions
            if analysis.get("suggestions"):
                output.append("SUGGESTIONS:")
                output.append("-" * 30)
                for suggestion in analysis["suggestions"]:
                    output.append(
                        f"[{suggestion['priority'].upper()}] {suggestion['message']}"
                    )
                    for action in suggestion["actions"]:
                        output.append(f"  • {action}")
                    output.append("")

            # Detailed failures
            if analysis.get("failures"):
                output.append("DETAILED FAILURES:")
                output.append("-" * 30)
                for i, failure in enumerate(
                    analysis["failures"][:5], 1
                ):  # Show first 5
                    output.append(f"{i}. {failure['test_name']}")
                    if failure.get("file_path"):
                        output.append(
                            f"   File: {failure['file_path']}:{failure.get('line_number', 'N/A')}"
                        )
                    output.append(f"   Error: {failure['error_type']}")
                    output.append(f"   Message: {failure['error_message'][:100]}...")
                    output.append("")

                if len(analysis["failures"]) > 5:
                    output.append(
                        f"... and {len(analysis['failures']) - 5} more failures"
                    )

            return "\n".join(output)

        except Exception as e:
            return f"Error analyzing test failures: {str(e)}"


def get_standalone_toolkit() -> List[BaseTool]:
    """Get all standalone tools with proper initialization.

    Returns:
        List of initialized standalone tools ready for LangChain integration.
    """
    return [
        CreateDirectoryTool(),
        CreateFileTool(),
        FileSearchTool(),
        GrepSearchTool(),
        ListDirTool(),
        ReadFileTool(),
        ReplaceStringTool(),
        SemanticSearchTool(),
        TestFailureAnalysisTool(),
    ]


def get_file_tools() -> List[BaseTool]:
    """Get only file-related tools."""
    return [
        CreateDirectoryTool(),
        CreateFileTool(),
        ListDirTool(),
        ReadFileTool(),
        ReplaceStringTool(),
    ]


def get_search_tools() -> List[BaseTool]:
    """Get only search-related tools."""
    return [
        FileSearchTool(),
        GrepSearchTool(),
        SemanticSearchTool(),
    ]


def get_analysis_tools() -> List[BaseTool]:
    """Get only analysis-related tools."""
    return [
        TestFailureAnalysisTool(),
        SemanticSearchTool(),
    ]


# Tool selection guide for LLMs
STANDALONE_TOOL_SELECTION_GUIDE = """
🛠️ STANDALONE TOOLKIT SELECTION GUIDE FOR AI AGENTS:

📁 FILE OPERATIONS:
1. **create_directory** - Create directory structures with permissions
2. **create_file** - Create files with content and auto-directory creation
3. **list_dir** - List directory contents with filtering and sorting
4. **read_file** - Read file contents with line ranges and encoding detection
5. **replace_string** - Replace strings in files with backup and validation

🔍 SEARCH OPERATIONS:
6. **file_search** - Find files using glob patterns
7. **grep_search** - Fast text search in files with regex support
8. **semantic_search** - Semantic code search using text similarity

🌐 WEB OPERATIONS:
9. **fetch_webpage** - Fetch and search web page content
10. **web_search** - Search the web for information

🔧 ANALYSIS OPERATIONS:
11. **test_failure_analysis** - Analyze test failures and suggest fixes

⚠️ WORKFLOW RECOMMENDATIONS:
1. Use create_directory/create_file for project setup
2. Use file_search to locate files by pattern
3. Use grep_search for content-based searches
4. Use semantic_search for natural language code queries
5. Use read_file to examine specific file contents
6. Use replace_string for safe file modifications
7. Use web_search for external information
8. Use test_failure_analysis for debugging tests

💡 BEST PRACTICES:
- Always use dry_run=True for replace_string before actual changes
- Use appropriate file_types filters for semantic_search
- Combine tools for comprehensive analysis workflows
- Use context_lines in grep_search for better understanding
- Enable show_details in file operations for thorough analysis

🔧 EXAMPLE USAGE PATTERNS:
# Find Python files and search for functions
file_search(query="*.py", show_details=True)
grep_search(query="def ", is_regexp=True, include_pattern="*.py")

# Safe file modification
replace_string(file_path="config.py", old_string="DEBUG = False", 
              new_string="DEBUG = True", dry_run=True)

# Semantic code search
semantic_search(query="authentication function", file_types=[".py"])

# Web research
web_search(query="Python async programming best practices")
"""

# Export all tools and utilities
__all__ = [
    # Tool classes
    "CreateDirectoryTool",
    "CreateFileTool",
    "FileSearchTool",
    "GrepSearchTool",
    "ListDirTool",
    "ReadFileTool",
    "ReplaceStringTool",
    "SemanticSearchTool",
    "TestFailureAnalysisTool",
    # Tool getters
    "get_standalone_toolkit",
    "get_file_tools",
    "get_search_tools",
    "get_analysis_tools",
    "get_web_tools",
    # Guide
    "STANDALONE_TOOL_SELECTION_GUIDE",
]
