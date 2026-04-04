#!/usr/bin/env python3
"""
Standalone Tools Launcher

This script helps you discover and run the available standalone tools.
"""

import argparse
import os
import subprocess
import sys
from pathlib import Path


def get_available_tools():
    """Get list of available tools in the current directory."""
    tools = []
    current_dir = Path(__file__).parent

    for file in current_dir.glob("*.py"):
        if file.name not in ["launcher.py", "__init__.py"]:
            tools.append(file.name)

    return sorted(tools)


def get_tool_description(tool_name):
    """Extract description from tool's docstring."""
    try:
        tool_path = Path(__file__).parent / tool_name
        with open(tool_path, "r", encoding="utf-8") as f:
            content = f.read()

        # Extract description from docstring
        if '"""' in content:
            docstring = content.split('"""')[1]
            if "Description:" in docstring:
                desc_line = docstring.split("Description:")[1].split("\n")[0]
                return desc_line.strip()

        return "No description available"
    except:
        return "No description available"


def list_tools():
    """List all available tools with their descriptions."""
    print("Available Standalone Tools:")
    print("=" * 50)

    tools = get_available_tools()
    for tool in tools:
        description = get_tool_description(tool)
        print(f"{tool:<30} - {description}")

    print(f"\nTotal: {len(tools)} tools available")
    print(
        "\nUse 'python launcher.py run <tool_name> --help' to see tool-specific options"
    )


def run_tool(tool_name, args):
    """Run a specific tool with the given arguments."""
    tools = get_available_tools()

    if tool_name not in tools:
        print(f"Error: Tool '{tool_name}' not found.")
        print("Available tools:")
        for tool in tools:
            print(f"  - {tool}")
        return 1

    tool_path = Path(__file__).parent / tool_name

    # Execute the tool
    try:
        cmd = [sys.executable, str(tool_path)] + args
        result = subprocess.run(cmd, check=False)
        return result.returncode
    except Exception as e:
        print(f"Error running tool '{tool_name}': {e}")
        return 1


def search_tools(keyword):
    """Search for tools by keyword in name or description."""
    print(f"Searching for tools containing '{keyword}':")
    print("=" * 50)

    tools = get_available_tools()
    found = []

    for tool in tools:
        description = get_tool_description(tool)
        if keyword.lower() in tool.lower() or keyword.lower() in description.lower():
            found.append((tool, description))

    if found:
        for tool, description in found:
            print(f"{tool:<30} - {description}")
        print(f"\nFound {len(found)} matching tools")
    else:
        print("No tools found matching the keyword.")


def main():
    parser = argparse.ArgumentParser(
        description="Launcher for standalone development tools",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  python launcher.py list                    # List all available tools
  python launcher.py search grep             # Search for tools containing 'grep'
  python launcher.py run file_search.py "*.py"  # Run file_search.py with arguments
  python launcher.py run read_file.py --help    # Show help for read_file.py
        """,
    )

    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # List command
    list_parser = subparsers.add_parser("list", help="List all available tools")

    # Search command
    search_parser = subparsers.add_parser("search", help="Search for tools by keyword")
    search_parser.add_argument("keyword", help="Keyword to search for")

    # Run command
    run_parser = subparsers.add_parser("run", help="Run a specific tool")
    run_parser.add_argument("tool", help="Name of the tool to run")
    run_parser.add_argument("args", nargs="*", help="Arguments to pass to the tool")

    args = parser.parse_args()

    if args.command == "list" or args.command is None:
        list_tools()
    elif args.command == "search":
        search_tools(args.keyword)
    elif args.command == "run":
        return run_tool(args.tool, args.args)

    return 0


if __name__ == "__main__":
    sys.exit(main())
