#!/usr/bin/env python3
"""Regenerate the generated API-surface sections of the SDK READMEs.

Each SDK README owns one block between the BEGIN/END api-surface marker
comments. This script rewrites only what is inside those markers, from the
current client sources, so the reference tables cannot drift from the code:

  clients/python-client/README.md          classes parsed with ast
  clients/go-client/README.md              exported ToolplaneClient methods
  clients/typescript-client/README.md      ToolplaneClient + ProviderRuntime
  clients/typescript-mcp-adapter/README.md environment variables

Usage:
  python tools/gen_sdk_readmes.py            # rewrite the marked blocks
  python tools/gen_sdk_readmes.py --check    # exit 1 if a block is stale

The script is stdlib-only so CI can run it without installing anything.
It requires Python 3.9+ (it uses ast.unparse); CI runs 3.12.
"""

from __future__ import annotations

import argparse
import ast
import re
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Callable

REPO_ROOT = Path(__file__).resolve().parents[1]


def md(value: str) -> str:
    """Escape a source-derived value for use inside a Markdown table cell."""
    return value.replace("|", "\\|")

BEGIN = (
    "<!-- BEGIN GENERATED: api-surface -- tooling: tools/gen_sdk_readmes.py;"
    " edits inside this block are overwritten -->"
)
END = "<!-- END GENERATED: api-surface -->"


@dataclass
class Target:
    readme: Path
    anchor: str  # section heading used to place the block when markers are absent
    render: Callable[[], str]


# --------------------------------------------------------------------------
# Python client: parse the public classes with ast.
# --------------------------------------------------------------------------

def _decorator_names(item: ast.FunctionDef | ast.AsyncFunctionDef) -> set[str]:
    names = set()
    for dec in item.decorator_list:
        target = dec.func if isinstance(dec, ast.Call) else dec
        if isinstance(target, ast.Name):
            names.add(target.id)
        elif isinstance(target, ast.Attribute):
            names.add(target.attr)
    return names


def python_class_rows(path: Path, class_name: str) -> list[tuple[str, str]]:
    tree = ast.parse(path.read_text(encoding="utf-8"))
    for node in tree.body:
        if not (isinstance(node, ast.ClassDef) and node.name == class_name):
            continue
        rows: list[tuple[str, str]] = []
        for item in node.body:
            if not isinstance(item, (ast.FunctionDef, ast.AsyncFunctionDef)):
                continue
            if item.name.startswith("_"):
                continue
            decorators = _decorator_names(item)
            is_property = "property" in decorators
            # Render properties without call parens: usage is provider.running,
            # not provider.running().
            bound = "cls" if "classmethod" in decorators else "self"
            args = ast.unparse(item.args)
            if args == bound:
                args = ""
            elif args.startswith(bound + ","):
                args = args[len(bound) + 1:].strip()
            prefix = "async " if isinstance(item, ast.AsyncFunctionDef) else ""
            if is_property:
                signature = item.name
            else:
                signature = f"{prefix}{item.name}({args})"
            if item.returns is not None:
                signature += f" -> {ast.unparse(item.returns)}"
            doc = (ast.get_docstring(item) or "").strip().splitlines()
            first = doc[0] if doc else ""
            if is_property:
                first = f"Property. {first}" if first else "Property."
            rows.append((signature, first))
        return rows
    raise SystemExit(f"class {class_name} not found in {path}")


def python_table(title: str, source: str, rows: list[tuple[str, str]]) -> str:
    out = [f"#### `{title}`", "", f"Public methods parsed from `{source}`:", ""]
    out += ["| Method | Description |", "| --- | --- |"]
    for signature, description in rows:
        out.append(f"| `{md(signature)}` | {md(description)} |")
    out.append("")
    return "\n".join(out)


def render_python() -> str:
    pkg = REPO_ROOT / "clients/python-client/toolplane"
    blocks = [
        python_table(
            "Toolplane",
            "toolplane/toolplane_client.py",
            python_class_rows(pkg / "toolplane_client.py", "Toolplane"),
        ),
        python_table(
            "ToolplaneHTTP",
            "toolplane/toolplane_http_client.py",
            python_class_rows(pkg / "toolplane_http_client.py", "ToolplaneHTTP"),
        ),
        python_table(
            "ProviderRuntime",
            "toolplane/provider_runtime.py",
            python_class_rows(pkg / "provider_runtime.py", "ProviderRuntime"),
        ),
    ]
    return "\n".join(blocks)


# --------------------------------------------------------------------------
# Go client: exported methods of ToolplaneClient, in source order.
# --------------------------------------------------------------------------

def go_client_rows(path: Path) -> list[tuple[str, str]]:
    text = path.read_text(encoding="utf-8")
    rows: list[tuple[str, str]] = []
    for m in re.finditer(r"^func \(c \*ToolplaneClient\) ([A-Z]\w*)\(", text, re.M):
        start = m.end()  # just past the opening paren
        depth, i = 1, start
        while depth:
            if text[i] == "(":
                depth += 1
            elif text[i] == ")":
                depth -= 1
            i += 1
        params = " ".join(text[start:i - 1].split()).replace(", )", ")").rstrip(",")
        brace = text.index("{", i)
        returns = " ".join(text[i:brace].split())
        rows.append((f"{m.group(1)}({params})", returns))
    if not rows:
        raise SystemExit(f"no exported ToolplaneClient methods found in {path}")
    return rows


def render_go() -> str:
    path = REPO_ROOT / "clients/go-client/client/toolplane_client.go"
    rows = go_client_rows(path)
    out = [
        "### Method reference",
        "",
        f"All {len(rows)} exported `ToolplaneClient` methods, parsed from `client/toolplane_client.go`:",
        "",
        "| Method | Returns |",
        "| --- | --- |",
    ]
    for signature, returns in rows:
        out.append(f"| `{md(signature)}` | `{md(returns)}` |")
    out.append("")
    return "\n".join(out)


# --------------------------------------------------------------------------
# TypeScript client: class members of ToolplaneClient and ProviderRuntime.
# --------------------------------------------------------------------------

def ts_class_rows(path: Path, class_name: str) -> list[tuple[str, str]]:
    text = path.read_text(encoding="utf-8")
    start = re.search(rf"^export (?:abstract )?class {class_name}\b", text, re.M)
    if not start:
        raise SystemExit(f"class {class_name} not found in {path}")
    end = text.index("\n}", start.end())  # closing brace of the class at column 0
    body = text[start.end():end]

    rows: list[tuple[str, str]] = []
    for m in re.finditer(
        r"^  ((?:static |async |get |set )*)([A-Za-z_]\w*)\(", body, re.M
    ):
        modifiers, name = m.group(1), m.group(2)
        if name.startswith("_") or "private" in m.group(1) or name == "constructor":
            continue
        i = m.end()  # just past the opening paren
        depth = 1
        while depth:
            if body[i] == "(":
                depth += 1
            elif body[i] == ")":
                depth -= 1
            i += 1
        params = " ".join(body[m.end():i - 1].split()).replace(", )", ")").rstrip(",")
        # Return annotation: scan from ':' to the balanced start of the body so
        # object-literal types like Promise<{ chunks: unknown[] }> survive.
        ret = ""
        j = i
        while j < len(body) and body[j].isspace():
            j += 1
        if j < len(body) and body[j] == ":":
            j += 1
            depth, k = 0, j
            while k < len(body):
                ch = body[k]
                # A whitespace-prefixed '{' at depth 0 opens the method body;
                # one attached to type characters opens an object-literal type.
                if ch == "{" and depth == 0 and body[k - 1] in " \t\r\n":
                    break
                if ch in "([{":
                    depth += 1
                elif ch in ")]}":
                    if depth == 0:
                        break
                    depth -= 1
                k += 1
            ret = " ".join(body[j:k].split())
        rows.append((f"{modifiers.strip()} {name}({params})".strip(), ret))
    if not rows:
        raise SystemExit(f"no public methods found for {class_name} in {path}")
    return rows


def ts_table(title: str, source: str, rows: list[tuple[str, str]]) -> str:
    out = [f"#### `{title}`", "", f"Public methods parsed from `{source}`:", ""]
    out += ["| Method | Returns |", "| --- | --- |"]
    for signature, returns in rows:
        out.append(f"| `{md(signature)}` | `{md(returns)}` |")
    out.append("")
    return "\n".join(out)


def render_typescript() -> str:
    src = REPO_ROOT / "clients/typescript-client/src"
    blocks = [
        ts_table(
            "ToolplaneClient",
            "src/core/toolplane_client.ts",
            ts_class_rows(src / "core/toolplane_client.ts", "ToolplaneClient"),
        ),
        ts_table(
            "ProviderRuntime",
            "src/provider_runtime.ts",
            ts_class_rows(src / "provider_runtime.ts", "ProviderRuntime"),
        ),
    ]
    return "\n".join(blocks)


# --------------------------------------------------------------------------
# MCP adapter: environment variables read in src/config.ts, in source order.
# --------------------------------------------------------------------------

def render_adapter_env() -> str:
    path = REPO_ROOT / "clients/typescript-mcp-adapter/src/config.ts"
    names = list(dict.fromkeys(
        re.findall(r"env\.(TOOLPLANE_[A-Z0-9_]+)", path.read_text(encoding="utf-8"))
    ))
    if not names:
        raise SystemExit(f"no environment variables found in {path}")
    out = ["Generated from `src/config.ts`:", ""]
    out += [f"- `{name}`" for name in names]
    out.append("")
    return "\n".join(out)


TARGETS = [
    Target(REPO_ROOT / "clients/python-client/README.md", "## API Reference", render_python),
    Target(REPO_ROOT / "clients/go-client/README.md", "## Public API Snapshot", render_go),
    Target(REPO_ROOT / "clients/typescript-client/README.md", "## Public API Snapshot", render_typescript),
    Target(REPO_ROOT / "clients/typescript-mcp-adapter/README.md", "## Environment", render_adapter_env),
]


def place_block(text: str, target: Target, block: str) -> str:
    """Replace the marked block, or insert one at the end of the anchor section."""
    begin_at = text.find(BEGIN)
    if begin_at != -1:
        end_at = text.index(END, begin_at) + len(END)
        if text[end_at:end_at + 1] == "\n":
            end_at += 1
        return text[:begin_at] + block + text[end_at:]

    anchor_at = text.index(target.anchor)
    heading = re.compile(r"^## ", re.M)
    nxt = heading.search(text, anchor_at + len(target.anchor))
    insert_at = nxt.start() if nxt else len(text)
    return text[:insert_at] + block + "\n" + text[insert_at:]


def marked_block(block: str) -> str:
    return f"{BEGIN}\n{block}{END}\n"


def main() -> int:
    if sys.version_info < (3, 9):
        raise SystemExit("tools/gen_sdk_readmes.py requires Python 3.9+ (it uses ast.unparse)")
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="fail instead of writing when stale")
    args = parser.parse_args()

    stale: list[str] = []
    for target in TARGETS:
        text = target.readme.read_text(encoding="utf-8")
        updated = place_block(text, target, marked_block(target.render()))
        if updated == text:
            print(f"up to date: {target.readme.relative_to(REPO_ROOT)}")
            continue
        if args.check and BEGIN in text:
            stale.append(str(target.readme.relative_to(REPO_ROOT)))
            continue
        if args.check:
            print(f"missing api-surface markers: {target.readme.relative_to(REPO_ROOT)}")
            stale.append(str(target.readme.relative_to(REPO_ROOT)))
            continue
        target.readme.write_text(updated, encoding="utf-8")
        print(f"updated:   {target.readme.relative_to(REPO_ROOT)}")

    if stale:
        print("\nStale generated README blocks; run: python tools/gen_sdk_readmes.py")
        for name in stale:
            print(f"  {name}")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
