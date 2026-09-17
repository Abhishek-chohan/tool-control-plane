"""toolplane-provider: serve, validate, test, and inspect tool files.

The consumer of a tools file — a Python module using the bare ``@tool``
decorator. ``serve`` binds the collected tools to a session and runs the
provider loop (claim, execute, heartbeat, lease renewal); ``validate``
checks the file without executing it; ``call`` runs one tool locally;
``schema`` prints a tool's generated JSON schema.
"""

import argparse
import ast
import json
import logging
import sys
from typing import List, Optional, Tuple

from toolplane import Toolplane
from toolplane.provider_registry import RegistryTool, clear, collect
from toolplane.utils.schema import generate_schema_from_function

logger = logging.getLogger("toolplane.provider_cli")


def _load_module(path: str):
    """Import a tools file as a module (executes it — the standard Python
    import contract; use ``validate`` for an execution-free check)."""
    import importlib.abc

    spec = importlib.util.spec_from_file_location("_toolplane_tools", path)
    if spec is None or spec.loader is None:
        raise ValueError(f"cannot import {path}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def _registry_tools(path: str) -> List[RegistryTool]:
    clear()
    _load_module(path)
    return collect()


def _decorated_functions(source: str) -> List[Tuple[str, Optional[str]]]:
    """Extract (function name, schema literal) pairs for functions
    decorated with @tool, without executing the module."""
    parsed = ast.parse(source)
    pairs: List[Tuple[str, Optional[str]]] = []
    for node in ast.walk(parsed):
        if not isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
            continue
        for decorator in node.decorator_list:
            name = None
            if isinstance(decorator, ast.Name) and decorator.id == "tool":
                name = node.name
            elif isinstance(decorator, ast.Call):
                func = decorator.func
                if isinstance(func, ast.Name) and func.id == "tool":
                    name = node.name
            if name:
                pairs.append((node.name, _schema_kwarg_literal(decorator)))
                break
    return pairs


def _schema_kwarg_literal(decorator: ast.AST) -> Optional[str]:
    if not isinstance(decorator, ast.Call):
        return None
    for keyword in decorator.keywords:
        if keyword.arg == "schema" and isinstance(keyword.value, ast.Constant):
            value = keyword.value.value
            return value if isinstance(value, str) else None
    return None


def cmd_validate(args: argparse.Namespace) -> int:
    """Execution-free check of a tools file: parseable, every @tool
    function collectable, schemas well-formed (JSON object root — the
    server's registration bar)."""
    path = args.file
    try:
        with open(path, "r", encoding="utf-8") as handle:
            source = handle.read()
    except OSError as exc:
        print(f"cannot read {path}: {exc}", file=sys.stderr)
        return 2

    try:
        pairs = _decorated_functions(source)
    except SyntaxError as exc:
        print(f"{path} is not valid Python: {exc}", file=sys.stderr)
        return 2

    if not pairs:
        print(f"{path}: no @tool functions found — nothing to serve", file=sys.stderr)
        return 2

    failures = 0
    for name, schema_literal in pairs:
        if schema_literal is None:
            continue
        try:
            decoded = json.loads(schema_literal)
        except json.JSONDecodeError as exc:
            print(f"  {name}: schema is not valid JSON: {exc}")
            failures += 1
            continue
        if not isinstance(decoded, dict):
            print(f"  {name}: schema root must be a JSON object")
            failures += 1

    if failures:
        print(f"{path}: {failures} invalid schema(s)", file=sys.stderr)
        return 2

    print(f"{path}: OK — {len(pairs)} tool(s): {', '.join(name for name, _ in pairs)}")
    return 0


def cmd_serve(args: argparse.Namespace) -> int:
    tools = _registry_tools(args.file)
    if not tools:
        print(
            f"{args.file}: no @tool functions found — nothing to serve", file=sys.stderr
        )
        return 2

    client = Toolplane(
        server_host=args.host,
        server_port=args.port,
        api_key=args.api_key or "",
    )
    runtime = client.provider_runtime(poll_interval=args.poll_interval)
    session = runtime.create_session(session_id=args.session or None)

    for entry in tools:
        schema = entry.schema or generate_schema_from_function(entry.func)
        runtime.tool(
            session_id=session.session_id,
            name=entry.name,
            description=entry.description,
            stream=entry.stream,
            tags=entry.tags,
        )
        session.register_tool(
            name=entry.name,
            func=entry.func,
            schema=schema,
            description=entry.description,
            stream=entry.stream,
            tags=entry.tags or [],
        )

    print(f"serving {len(tools)} tool(s) in session {session.session_id}")
    try:
        runtime.run_forever()
    except KeyboardInterrupt:
        pass
    finally:
        runtime.stop()
    return 0


def cmd_call(args: argparse.Namespace) -> int:
    tools = _registry_tools(args.file)
    entry = next((t for t in tools if t.name == args.tool), None)
    if entry is None:
        print(
            f"tool {args.tool} not found in {args.file}; available: "
            f"{', '.join(t.name for t in tools) or '(none)'}",
            file=sys.stderr,
        )
        return 3
    try:
        input_data = json.loads(args.input) if args.input else {}
    except json.JSONDecodeError as exc:
        print(f"--input must be a JSON object: {exc}", file=sys.stderr)
        return 2
    result = entry.func(**input_data) if isinstance(input_data, dict) else entry.func()
    print(json.dumps(result, default=str) if not isinstance(result, str) else result)
    return 0


def cmd_schema(args: argparse.Namespace) -> int:
    tools = _registry_tools(args.file)
    entry = next((t for t in tools if t.name == args.tool), None)
    if entry is None:
        print(f"tool {args.tool} not found in {args.file}", file=sys.stderr)
        return 3
    generated = entry.schema or generate_schema_from_function(entry.func)
    # generate_schema_from_function wraps the schema with name/description;
    # the operator wants the schema itself.
    body = (
        generated.get("schema", generated) if isinstance(generated, dict) else generated
    )
    print(json.dumps(body, indent=2))
    return 0


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="toolplane-provider",
        description="Serve and manage Toolplane tool files (bare @tool modules).",
    )
    sub = parser.add_subparsers(dest="command", required=True)

    serve = sub.add_parser("serve", help="bind a tools file to a session and serve it")
    serve.add_argument(
        "file", nargs="?", default="tools.py", help="tools file (default: ./tools.py)"
    )
    serve.add_argument(
        "--session",
        default=None,
        help="session id (created when missing; generated when omitted)",
    )
    serve.add_argument("--host", default="localhost")
    serve.add_argument("--port", type=int, default=9001)
    serve.add_argument("--api-key", default="")
    serve.add_argument("--poll-interval", type=float, default=1.0)
    serve.set_defaults(func=cmd_serve)

    validate = sub.add_parser(
        "validate", help="check a tools file without executing it"
    )
    validate.add_argument("file", help="tools file")
    validate.set_defaults(func=cmd_validate)

    call = sub.add_parser("call", help="run one tool locally (no server)")
    call.add_argument("file", help="tools file")
    call.add_argument("tool", help="tool name")
    call.add_argument(
        "--input", default="{}", help="tool input as a JSON object string"
    )
    call.set_defaults(func=cmd_call)

    schema = sub.add_parser("schema", help="print a tool's generated JSON schema")
    schema.add_argument("file", help="tools file")
    schema.add_argument("tool", help="tool name")
    schema.set_defaults(func=cmd_schema)

    return parser


def main(argv: Optional[List[str]] = None) -> int:
    logging.basicConfig(
        level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s"
    )
    parser = build_parser()
    args = parser.parse_args(argv)
    return args.func(args)


if __name__ == "__main__":
    sys.exit(main())
