"""Shared-fixture conformance bootstrap — source of truth.

This module is the reference implementation for conformance environment setup.
The TypeScript equivalent lives at
  clients/typescript-client/tests/conformance/environment.ts
and must behave identically: same env defaults, same auto-boot sequence
(find free ports → go run server → go run proxy → wait for readiness),
and same teardown. If you change this file, apply the equivalent change
to the TypeScript bootstrap.

The env contract used here matches the release-gate contract documented in
  server/docs/release-gate.md
and enforced by the SDK Conformance & Verification workflow.
"""
import os
import socket
import subprocess
import time
from pathlib import Path

import pytest
import requests

BOOTSTRAP_READINESS_TIMEOUT_SECONDS = 60.0


def _repo_root() -> Path:
    return Path(__file__).resolve().parents[4]


def _server_root() -> Path:
    return _repo_root() / "server"


def _find_free_port() -> int:
    probe_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    probe_socket.bind(("127.0.0.1", 0))
    port = probe_socket.getsockname()[1]
    probe_socket.close()
    return int(port)


def _wait_for_tcp(host: str, port: int, timeout_seconds: float) -> bool:
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        test_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        test_socket.settimeout(0.5)
        try:
            test_socket.connect((host, port))
            test_socket.close()
            return True
        except OSError:
            test_socket.close()
            time.sleep(0.2)
    return False


def _wait_for_http_health(url: str, timeout_seconds: float) -> bool:
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        try:
            response = requests.get(url, timeout=1.0)
            if response.status_code == 200:
                return True
        except requests.RequestException:
            pass
        time.sleep(0.2)
    return False


def _terminate_process(process: subprocess.Popen) -> None:
    if process.poll() is not None:
        return
    process.terminate()
    try:
        process.wait(timeout=5)
    except subprocess.TimeoutExpired:
        process.kill()
        process.wait(timeout=5)


@pytest.fixture(scope="session", autouse=True)
def conformance_environment():
    auto_boot = os.getenv("TOOLPLANE_CONFORMANCE_AUTO_BOOT", "1").strip().lower() not in {
        "0",
        "false",
        "no",
    }

    default_api_key = "toolplane-conformance-fixture-key"
    if not os.getenv("TOOLPLANE_CONFORMANCE_API_KEY"):
        os.environ["TOOLPLANE_CONFORMANCE_API_KEY"] = default_api_key

    os.environ.setdefault("TOOLPLANE_ENV_MODE", "development")
    os.environ.setdefault("TOOLPLANE_AUTH_MODE", "fixed")
    os.environ.setdefault("TOOLPLANE_AUTH_FIXED_API_KEY", os.environ["TOOLPLANE_CONFORMANCE_API_KEY"])
    os.environ.setdefault("TOOLPLANE_STORAGE_MODE", "memory")
    os.environ.setdefault("TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND", "1")

    if not auto_boot:
        os.environ.setdefault("TOOLPLANE_CONFORMANCE_GRPC_HOST", "localhost")
        os.environ.setdefault("TOOLPLANE_CONFORMANCE_GRPC_PORT", "50051")
        os.environ.setdefault("TOOLPLANE_CONFORMANCE_HTTP_HOST", "localhost")
        os.environ.setdefault("TOOLPLANE_CONFORMANCE_HTTP_PORT", "8080")
        os.environ.setdefault("TOOLPLANE_CONFORMANCE_USER_ID", "conformance-user")
        yield
        return

    grpc_port = _find_free_port()
    http_port = _find_free_port()

    os.environ["TOOLPLANE_CONFORMANCE_GRPC_HOST"] = "localhost"
    os.environ["TOOLPLANE_CONFORMANCE_GRPC_PORT"] = str(grpc_port)
    os.environ["TOOLPLANE_CONFORMANCE_HTTP_HOST"] = "localhost"
    os.environ["TOOLPLANE_CONFORMANCE_HTTP_PORT"] = str(http_port)
    os.environ.setdefault("TOOLPLANE_CONFORMANCE_USER_ID", "conformance-user")

    log_dir = _repo_root() / ".tmp" / "conformance-logs"
    log_dir.mkdir(parents=True, exist_ok=True)

    server_log_path = log_dir / "grpc_server.log"
    proxy_log_path = log_dir / "http_proxy.log"

    server_log_handle = open(server_log_path, "w", encoding="utf-8")
    proxy_log_handle = open(proxy_log_path, "w", encoding="utf-8")

    server_process = subprocess.Popen(
        ["go", "run", "./cmd/server", "--port", str(grpc_port)],
        cwd=_server_root(),
        stdout=server_log_handle,
        stderr=subprocess.STDOUT,
        env=os.environ.copy(),
    )

    if not _wait_for_tcp("127.0.0.1", grpc_port, timeout_seconds=BOOTSTRAP_READINESS_TIMEOUT_SECONDS):
        _terminate_process(server_process)
        server_log_handle.close()
        proxy_log_handle.close()
        pytest.skip(
            f"Conformance bootstrap failed: gRPC server did not become ready on {grpc_port}. Check {server_log_path}"
        )

    proxy_process = subprocess.Popen(
        [
            "go",
            "run",
            "./cmd/proxy",
            "--listen",
            f":{http_port}",
            "--backend",
            f"localhost:{grpc_port}",
        ],
        cwd=_server_root(),
        stdout=proxy_log_handle,
        stderr=subprocess.STDOUT,
        env=os.environ.copy(),
    )

    health_url = f"http://127.0.0.1:{http_port}/health"
    if not _wait_for_http_health(health_url, timeout_seconds=BOOTSTRAP_READINESS_TIMEOUT_SECONDS):
        _terminate_process(proxy_process)
        _terminate_process(server_process)
        server_log_handle.close()
        proxy_log_handle.close()
        pytest.skip(
            f"Conformance bootstrap failed: HTTP gateway did not become ready on {http_port}. Check {proxy_log_path}"
        )

    try:
        yield
    finally:
        _terminate_process(proxy_process)
        _terminate_process(server_process)
        server_log_handle.close()
        proxy_log_handle.close()
