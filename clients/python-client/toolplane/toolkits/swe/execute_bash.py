#!/usr/bin/env python3
"""
Description: Execute a bash command in the terminal.

WARNING: this tool provides NO sandboxing. It runs model-supplied shell
commands with the full privileges of the provider process. Only register it
on a provider that is itself isolated (container or VM). The former
first-token command blocklist was deliberately removed: such filters are
trivially bypassed (absolute paths, env assignments, `sh -c`, ...) and
provided a false sense of guardrail while blocking nothing.

Environment:
  TOOLPLANE_BASH_TIMEOUT_SECONDS  maximum runtime per command (default 120).
  TOOLPLANE_WORKSPACE_ROOT        when set, commands run with this working
                                  directory. This is a convenience and a soft
                                  boundary only — the shell itself is not
                                  confined; real isolation must come from the
                                  provider's container/VM.

Parameters:
  --command (string, optional): The bash command to execute. For example: --command 'python my_script.py'. If not provided, will show help.
"""

import argparse
import os
import subprocess
import sys

DEFAULT_COMMAND_TIMEOUT_SECONDS = 120.0


def command_timeout_seconds():
    raw = os.environ.get("TOOLPLANE_BASH_TIMEOUT_SECONDS", "").strip()
    if not raw:
        return DEFAULT_COMMAND_TIMEOUT_SECONDS
    try:
        value = float(raw)
    except ValueError:
        return DEFAULT_COMMAND_TIMEOUT_SECONDS
    return value if value > 0 else DEFAULT_COMMAND_TIMEOUT_SECONDS


def workspace_root():
    root = os.environ.get("TOOLPLANE_WORKSPACE_ROOT", "").strip()
    return root or None


def run_command(cmd):
    """Run a shell command with a hard timeout and the configured workspace cwd."""
    return subprocess.run(
        cmd,
        shell=True,
        capture_output=True,
        text=True,
        timeout=command_timeout_seconds(),
        cwd=workspace_root(),
    )


def main():
    parser = argparse.ArgumentParser(description="Execute a bash command.")
    parser.add_argument(
        "command",
        type=str,
        help="The command (and optional arguments) to execute. For example: 'python my_script.py'",
    )
    args = parser.parse_args()

    try:
        result = run_command(args.command)
    except subprocess.TimeoutExpired:
        print(
            f"Command timed out after {command_timeout_seconds():.0f}s "
            "(configure via TOOLPLANE_BASH_TIMEOUT_SECONDS)."
        )
        sys.exit(124)

    if result.returncode != 0:
        print("Error executing command:\n")
        print("[STDOUT]\n")
        print(result.stdout.strip(), "\n")
        print("[STDERR]\n")
        print(result.stderr.strip())
        sys.exit(result.returncode)

    print("[STDOUT]\n")
    print(result.stdout.strip(), "\n")
    print("[STDERR]\n")
    print(result.stderr.strip())


if __name__ == "__main__":
    main()
