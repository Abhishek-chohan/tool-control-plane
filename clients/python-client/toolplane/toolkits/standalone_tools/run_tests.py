#!/usr/bin/env python3
"""
Test runner script for standalone tools.

This script provides an easy way to run the test suite with various options.
"""

import argparse
import subprocess
import sys
from pathlib import Path


def run_tests(test_pattern=None, verbose=False, coverage=False, markers=None):
    """Run the test suite with specified options."""
    cmd = [sys.executable, "-m", "pytest"]

    if verbose:
        cmd.append("-v")

    if coverage:
        cmd.extend(["--cov=.", "--cov-report=html", "--cov-report=term"])

    if markers:
        cmd.extend(["-m", markers])

    if test_pattern:
        cmd.extend(["-k", test_pattern])

    # Add the test file
    cmd.append("test_standalone_toolkit.py")

    print(f"Running: {' '.join(cmd)}")
    return subprocess.run(cmd).returncode


def main():
    parser = argparse.ArgumentParser(description="Run standalone tools test suite")
    parser.add_argument("-v", "--verbose", action="store_true", help="Verbose output")
    parser.add_argument(
        "-c", "--coverage", action="store_true", help="Run with coverage analysis"
    )
    parser.add_argument("-k", "--pattern", help="Run tests matching pattern")
    parser.add_argument(
        "-m", "--markers", help="Run tests with specific markers (e.g., 'not slow')"
    )
    parser.add_argument(
        "--install-deps", action="store_true", help="Install test dependencies first"
    )

    args = parser.parse_args()

    if args.install_deps:
        print("Installing test dependencies...")
        subprocess.run([sys.executable, "-m", "pip", "install", "pytest", "pytest-cov"])

    return run_tests(
        test_pattern=args.pattern,
        verbose=args.verbose,
        coverage=args.coverage,
        markers=args.markers,
    )


if __name__ == "__main__":
    sys.exit(main())
