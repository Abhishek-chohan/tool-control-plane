#!/usr/bin/env python3
"""
Description: Test failure analysis tool for examining and debugging test failures.

This tool analyzes test failures, provides detailed error analysis, and suggests
potential fixes based on common failure patterns.

Parameters:
  test_output (string, optional): Path to test output file or direct test output.
  test_framework (string, optional): Test framework used (pytest, unittest, jest, etc.).
  verbose (boolean, optional): Show detailed analysis (default: False).
  suggest_fixes (boolean, optional): Suggest potential fixes (default: True).
  group_by_type (boolean, optional): Group failures by error type (default: True).
"""

import argparse
import json
import os
import re
import sys
from collections import defaultdict
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple


def analyze_test_failures(
    test_output: str = None,
    test_framework: str = None,
    verbose: bool = False,
    suggest_fixes: bool = True,
    group_by_type: bool = True,
) -> Dict[str, Any]:
    """
    Analyze test failures and provide detailed analysis.

    Args:
        test_output: Path to test output file or direct output
        test_framework: Test framework used
        verbose: Show detailed analysis
        suggest_fixes: Suggest potential fixes
        group_by_type: Group failures by error type

    Returns:
        Dictionary containing analysis results
    """
    # Get test output content
    if test_output and os.path.exists(test_output):
        with open(test_output, "r", encoding="utf-8", errors="ignore") as f:
            output_content = f.read()
    elif test_output:
        output_content = test_output
    else:
        # Read from stdin
        output_content = sys.stdin.read()

    # Auto-detect test framework if not specified
    if not test_framework:
        test_framework = detect_test_framework(output_content)

    # Parse failures based on framework
    failures = parse_failures(output_content, test_framework)

    # Analyze failures
    analysis = analyze_failures(failures, suggest_fixes, group_by_type)

    # Add metadata
    analysis["metadata"] = {
        "test_framework": test_framework,
        "total_failures": len(failures),
        "unique_error_types": len(set(f["error_type"] for f in failures)),
        "verbose": verbose,
        "suggest_fixes": suggest_fixes,
    }

    return analysis


def detect_test_framework(output: str) -> str:
    """Auto-detect the test framework from output."""
    output_lower = output.lower()

    if "pytest" in output_lower or "::" in output:
        return "pytest"
    elif "unittest" in output_lower or "test_" in output:
        return "unittest"
    elif "jest" in output_lower or "describe(" in output:
        return "jest"
    elif "rspec" in output_lower or "describe " in output:
        return "rspec"
    elif "mocha" in output_lower:
        return "mocha"
    elif "phpunit" in output_lower:
        return "phpunit"
    elif "cargo test" in output_lower:
        return "cargo"
    elif "go test" in output_lower:
        return "go"
    else:
        return "unknown"


def parse_failures(output: str, framework: str) -> List[Dict[str, Any]]:
    """Parse test failures based on the framework."""
    if framework == "pytest":
        return parse_pytest_failures(output)
    elif framework == "unittest":
        return parse_unittest_failures(output)
    elif framework == "jest":
        return parse_jest_failures(output)
    else:
        return parse_generic_failures(output)


def parse_pytest_failures(output: str) -> List[Dict[str, Any]]:
    """Parse pytest failure output."""
    failures = []

    # First, try to find formal FAILURES section
    failure_section_match = re.search(
        r"=+ FAILURES =+.*?(?=^=+|\Z)", output, re.MULTILINE | re.DOTALL
    )

    if failure_section_match:
        failure_section = failure_section_match.group(0)

        # Split individual failures
        individual_failures = re.split(
            r"^_+ (.+?) _+$", failure_section, flags=re.MULTILINE
        )

        for i in range(1, len(individual_failures), 2):
            if i + 1 >= len(individual_failures):
                break

            test_name = individual_failures[i].strip()
            failure_content = individual_failures[i + 1].strip()

            # Extract file and line info
            file_match = re.search(r"(\S+\.py):(\d+):", failure_content)
            file_path = file_match.group(1) if file_match else None
            line_number = int(file_match.group(2)) if file_match else None

            # Extract error type and message
            error_match = re.search(
                r"(AssertionError|TypeError|ValueError|AttributeError|KeyError|IndexError|ImportError|NameError|RuntimeError|Exception)[:>]?\s*(.*)",
                failure_content,
                re.DOTALL,
            )
            error_type = error_match.group(1) if error_match else "Unknown"
            error_message = error_match.group(2).strip() if error_match else ""

            # Extract assertion details
            assertion_match = re.search(r"assert (.+)", failure_content)
            assertion = assertion_match.group(1) if assertion_match else None

            # Extract traceback
            traceback_lines = []
            for line in failure_content.split("\n"):
                if re.match(r'^\s+File "', line) or re.match(r"^\s+", line):
                    traceback_lines.append(line.strip())

            failure = {
                "test_name": test_name,
                "file_path": file_path,
                "line_number": line_number,
                "error_type": error_type,
                "error_message": error_message,
                "assertion": assertion,
                "traceback": traceback_lines,
                "full_output": failure_content,
            }

            failures.append(failure)

    # Also look for simple FAILED lines (common in pytest short output)
    failed_lines = re.findall(r"FAILED (.+?) - (.+)", output)
    for test_name, error_info in failed_lines:
        # Extract error type from error_info
        error_match = re.search(
            r"(AssertionError|TypeError|ValueError|AttributeError|KeyError|IndexError|ImportError|NameError|RuntimeError|Exception)[:>]?\s*(.*)",
            error_info,
        )
        error_type = error_match.group(1) if error_match else "Unknown"
        error_message = error_match.group(2).strip() if error_match else error_info

        failure = {
            "test_name": test_name,
            "file_path": None,
            "line_number": None,
            "error_type": error_type,
            "error_message": error_message,
            "assertion": None,
            "traceback": [],
            "full_output": f"FAILED {test_name} - {error_info}",
        }

        failures.append(failure)

    return failures


def parse_unittest_failures(output: str) -> List[Dict[str, Any]]:
    """Parse unittest failure output."""
    failures = []

    # Find FAIL or ERROR sections
    fail_pattern = r"(FAIL|ERROR): (\S+) \((\S+)\)\n-+\n(.*?)(?=\n\n|\Z)"
    matches = re.findall(fail_pattern, output, re.DOTALL)

    for match in matches:
        failure_type, method_name, class_name, content = match

        # Extract file and line info
        file_match = re.search(r'File "([^"]+)", line (\d+)', content)
        file_path = file_match.group(1) if file_match else None
        line_number = int(file_match.group(2)) if file_match else None

        # Extract error type and message
        error_match = re.search(
            r"(AssertionError|TypeError|ValueError|AttributeError|KeyError|IndexError|ImportError|NameError|RuntimeError|Exception)[:>]?\s*(.*)",
            content,
            re.DOTALL,
        )
        error_type = error_match.group(1) if error_match else failure_type
        error_message = error_match.group(2).strip() if error_match else ""

        failure = {
            "test_name": f"{class_name}.{method_name}",
            "file_path": file_path,
            "line_number": line_number,
            "error_type": error_type,
            "error_message": error_message,
            "full_output": content,
        }

        failures.append(failure)

    return failures


def parse_jest_failures(output: str) -> List[Dict[str, Any]]:
    """Parse Jest failure output."""
    failures = []

    # Find test failures
    fail_pattern = r"● (.+?)\n\n(.*?)(?=\n  ●|\n\nTest Suites|\Z)"
    matches = re.findall(fail_pattern, output, re.DOTALL)

    for match in matches:
        test_name, content = match

        # Extract file and line info
        file_match = re.search(r"at (.+?):(\d+):(\d+)", content)
        file_path = file_match.group(1) if file_match else None
        line_number = int(file_match.group(2)) if file_match else None

        # Extract error type and message
        error_match = re.search(
            r"(Error|TypeError|ReferenceError|SyntaxError)[:>]?\s*(.*)",
            content,
            re.DOTALL,
        )
        error_type = error_match.group(1) if error_match else "Unknown"
        error_message = error_match.group(2).strip() if error_match else ""

        failure = {
            "test_name": test_name,
            "file_path": file_path,
            "line_number": line_number,
            "error_type": error_type,
            "error_message": error_message,
            "full_output": content,
        }

        failures.append(failure)

    return failures


def parse_generic_failures(output: str) -> List[Dict[str, Any]]:
    """Parse generic test failure output."""
    failures = []

    # Look for common error patterns
    error_patterns = [
        r"(AssertionError|TypeError|ValueError|AttributeError|KeyError|IndexError|ImportError|NameError|RuntimeError|Exception)[:>]?\s*(.*)",
        r"FAILED (.+?) - (.+)",
        r"ERROR (.+?) - (.+)",
        r"✕ (.+)",
    ]

    for pattern in error_patterns:
        matches = re.findall(pattern, output, re.MULTILINE)
        for match in matches:
            if len(match) >= 2:
                error_type = match[0]
                error_message = match[1]

                failure = {
                    "test_name": "Unknown",
                    "file_path": None,
                    "line_number": None,
                    "error_type": error_type,
                    "error_message": error_message,
                    "full_output": str(match),
                }

                failures.append(failure)

    return failures


def analyze_failures(
    failures: List[Dict[str, Any]], suggest_fixes: bool, group_by_type: bool
) -> Dict[str, Any]:
    """Analyze parsed failures and provide insights."""
    analysis = {"failures": failures, "summary": {}, "patterns": {}, "suggestions": []}

    if not failures:
        return analysis

    # Group failures by error type
    if group_by_type:
        error_groups = defaultdict(list)
        for failure in failures:
            error_groups[failure["error_type"]].append(failure)

        analysis["error_groups"] = dict(error_groups)

    # Generate summary statistics
    analysis["summary"] = {
        "total_failures": len(failures),
        "unique_tests": len(set(f["test_name"] for f in failures)),
        "unique_files": len(set(f["file_path"] for f in failures if f["file_path"])),
        "most_common_error": (
            max(
                set(f["error_type"] for f in failures),
                key=lambda x: sum(1 for f in failures if f["error_type"] == x),
            )
            if failures
            else "None"
        ),
    }

    # Identify patterns
    analysis["patterns"] = identify_patterns(failures)

    # Generate suggestions
    if suggest_fixes:
        analysis["suggestions"] = generate_suggestions(failures, analysis["patterns"])

    return analysis


def identify_patterns(failures: List[Dict[str, Any]]) -> Dict[str, Any]:
    """Identify common patterns in test failures."""
    patterns = {
        "common_errors": {},
        "file_hotspots": {},
        "assertion_patterns": [],
        "import_errors": [],
    }

    # Count error types
    error_counts = defaultdict(int)
    for failure in failures:
        error_counts[failure["error_type"]] += 1

    patterns["common_errors"] = dict(error_counts)

    # Find file hotspots
    file_counts = defaultdict(int)
    for failure in failures:
        if failure["file_path"]:
            file_counts[failure["file_path"]] += 1

    patterns["file_hotspots"] = dict(file_counts)

    # Find assertion patterns
    assertion_patterns = []
    for failure in failures:
        if failure.get("assertion"):
            assertion_patterns.append(failure["assertion"])

    patterns["assertion_patterns"] = assertion_patterns

    # Find import errors
    import_errors = []
    for failure in failures:
        if failure["error_type"] in ["ImportError", "ModuleNotFoundError"]:
            import_errors.append(failure["error_message"])

    patterns["import_errors"] = import_errors

    return patterns


def generate_suggestions(
    failures: List[Dict[str, Any]], patterns: Dict[str, Any]
) -> List[Dict[str, Any]]:
    """Generate suggestions for fixing test failures."""
    suggestions = []

    # Suggestions based on common error types
    for error_type, count in patterns["common_errors"].items():
        if error_type == "AssertionError":
            suggestions.append(
                {
                    "type": "fix_suggestion",
                    "priority": "high",
                    "message": f"You have {count} assertion failures. Review test expectations and actual values.",
                    "actions": [
                        "Check if test data has changed",
                        "Verify expected vs actual values",
                        "Update assertions if requirements changed",
                    ],
                }
            )
        elif error_type == "ImportError":
            suggestions.append(
                {
                    "type": "fix_suggestion",
                    "priority": "high",
                    "message": f"You have {count} import errors. Check dependencies and module paths.",
                    "actions": [
                        "Install missing dependencies",
                        "Check PYTHONPATH or module paths",
                        "Verify module names and locations",
                    ],
                }
            )
        elif error_type == "AttributeError":
            suggestions.append(
                {
                    "type": "fix_suggestion",
                    "priority": "medium",
                    "message": f"You have {count} attribute errors. Check object interfaces and method names.",
                    "actions": [
                        "Verify object has expected attributes/methods",
                        "Check for typos in attribute names",
                        "Ensure objects are properly initialized",
                    ],
                }
            )
        elif error_type == "TypeError":
            suggestions.append(
                {
                    "type": "fix_suggestion",
                    "priority": "medium",
                    "message": f"You have {count} type errors. Check function arguments and return types.",
                    "actions": [
                        "Verify function signatures",
                        "Check argument types being passed",
                        "Ensure proper type conversions",
                    ],
                }
            )

    # Suggestions based on file hotspots
    for file_path, count in patterns["file_hotspots"].items():
        if count > 1:
            suggestions.append(
                {
                    "type": "code_review",
                    "priority": "medium",
                    "message": f"File {file_path} has {count} failing tests. Consider reviewing this file.",
                    "actions": [
                        "Review recent changes to this file",
                        "Check for systematic issues",
                        "Consider refactoring if needed",
                    ],
                }
            )

    # Suggestions based on import errors
    if patterns["import_errors"]:
        unique_imports = set(patterns["import_errors"])
        suggestions.append(
            {
                "type": "environment",
                "priority": "high",
                "message": f'Missing imports detected: {", ".join(list(unique_imports)[:3])}...',
                "actions": [
                    "Check requirements.txt or package.json",
                    "Run pip install or npm install",
                    "Verify virtual environment activation",
                ],
            }
        )

    return suggestions


def format_analysis(analysis: Dict[str, Any], output_format: str) -> str:
    """Format analysis results for display."""
    if output_format == "json":
        return json.dumps(analysis, indent=2)

    lines = []

    # Header
    lines.append("TEST FAILURE ANALYSIS")
    lines.append("=" * 50)

    # Summary
    summary = analysis["summary"]
    lines.append(f"Total failures: {summary['total_failures']}")
    lines.append(f"Unique tests: {summary['unique_tests']}")
    lines.append(f"Unique files: {summary['unique_files']}")
    lines.append(f"Most common error: {summary['most_common_error']}")
    lines.append("")

    # Error groups
    if "error_groups" in analysis:
        lines.append("ERROR GROUPS:")
        lines.append("-" * 30)
        for error_type, group_failures in analysis["error_groups"].items():
            lines.append(f"{error_type}: {len(group_failures)} failures")
            for failure in group_failures[:3]:  # Show first 3
                lines.append(f"  - {failure['test_name']}")
            if len(group_failures) > 3:
                lines.append(f"  ... and {len(group_failures) - 3} more")
        lines.append("")

    # Suggestions
    if analysis["suggestions"]:
        lines.append("SUGGESTIONS:")
        lines.append("-" * 30)
        for suggestion in analysis["suggestions"]:
            lines.append(f"[{suggestion['priority'].upper()}] {suggestion['message']}")
            for action in suggestion["actions"]:
                lines.append(f"  • {action}")
            lines.append("")

    # Detailed failures
    lines.append("DETAILED FAILURES:")
    lines.append("-" * 30)
    for i, failure in enumerate(analysis["failures"][:5], 1):  # Show first 5
        lines.append(f"{i}. {failure['test_name']}")
        if failure["file_path"]:
            lines.append(f"   File: {failure['file_path']}:{failure['line_number']}")
        lines.append(f"   Error: {failure['error_type']}")
        lines.append(f"   Message: {failure['error_message'][:100]}...")
        lines.append("")

    if len(analysis["failures"]) > 5:
        lines.append(f"... and {len(analysis['failures']) - 5} more failures")

    return "\n".join(lines)


def main():
    parser = argparse.ArgumentParser(
        description="Test failure analysis tool for examining and debugging test failures."
    )
    parser.add_argument(
        "--test_output", type=str, help="Path to test output file or direct test output"
    )
    parser.add_argument(
        "--test_framework",
        choices=[
            "pytest",
            "unittest",
            "jest",
            "rspec",
            "mocha",
            "phpunit",
            "cargo",
            "go",
        ],
        help="Test framework used",
    )
    parser.add_argument(
        "--verbose",
        action="store_true",
        default=False,
        help="Show detailed analysis (default: False)",
    )
    parser.add_argument(
        "--suggest_fixes",
        action="store_true",
        default=True,
        help="Suggest potential fixes (default: True)",
    )
    parser.add_argument(
        "--group_by_type",
        action="store_true",
        default=True,
        help="Group failures by error type (default: True)",
    )
    parser.add_argument(
        "--output_format",
        choices=["default", "json"],
        default="default",
        help="Output format (default: default)",
    )

    args = parser.parse_args()

    try:
        analysis = analyze_test_failures(
            test_output=args.test_output,
            test_framework=args.test_framework,
            verbose=args.verbose,
            suggest_fixes=args.suggest_fixes,
            group_by_type=args.group_by_type,
        )

        output = format_analysis(analysis, args.output_format)
        print(output)

    except Exception as e:
        print(f"Error analyzing test failures: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
