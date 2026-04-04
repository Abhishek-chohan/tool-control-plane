#!/usr/bin/env python3

from pathlib import Path
import sys


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: fix_python_proto_imports.py <service_pb2_grpc.py>", file=sys.stderr)
        return 1

    target_path = Path(sys.argv[1])
    source = target_path.read_text(encoding="utf-8")
    old = "from proto import service_pb2 as proto_dot_service__pb2\n"
    new = "from . import service_pb2 as proto_dot_service__pb2\n"

    if old not in source:
        if new in source:
            return 0
        print(
            f"expected generated import not found in {target_path}",
            file=sys.stderr,
        )
        return 1

    target_path.write_text(source.replace(old, new), encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())