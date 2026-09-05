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
import re
import socket
import subprocess
import time
from collections import defaultdict
from pathlib import Path

import pytest
import requests

BOOTSTRAP_READINESS_TIMEOUT_SECONDS = 60.0
BOOTSTRAP_LOG_TAIL_BYTES = 8192


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


def _wait_for_tcp(
    host: str, port: int, timeout_seconds: float, process: subprocess.Popen = None
) -> bool:
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        if process is not None and process.poll() is not None:
            # The bootstrapped process exited; waiting longer cannot help.
            return False
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


def _wait_for_http_health(
    url: str, timeout_seconds: float, process: subprocess.Popen = None
) -> bool:
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        if process is not None and process.poll() is not None:
            return False
        try:
            response = requests.get(url, timeout=1.0)
            if response.status_code == 200:
                return True
        except requests.RequestException:
            pass
        time.sleep(0.2)
    return False


def _read_log_tail(log_path: Path) -> str:
    try:
        with open(log_path, "rb") as handle:
            handle.seek(0, os.SEEK_END)
            size = handle.tell()
            handle.seek(max(0, size - BOOTSTRAP_LOG_TAIL_BYTES))
            return handle.read().decode("utf-8", errors="replace")
    except OSError:
        return "<log unavailable>"


def _fail_bootstrap(message: str, log_path: Path) -> None:
    """Bootstrap failures are hard failures, never skips.

    A skipped conformance suite exits 0 and turns the release gate green while
    proving nothing; a server that fails to compile or bind must fail the run
    loudly, with its log tail attached.
    """
    tail = _read_log_tail(log_path)
    pytest.fail(f"{message}\n--- {log_path} (tail) ---\n{tail}", pytrace=False)


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
    auto_boot = os.getenv(
        "TOOLPLANE_CONFORMANCE_AUTO_BOOT", "1"
    ).strip().lower() not in {
        "0",
        "false",
        "no",
    }
    multi_instance = os.getenv(
        "TOOLPLANE_CONFORMANCE_MULTI_INSTANCE", "0"
    ).strip().lower() in {
        "1",
        "true",
        "yes",
    }

    default_api_key = "toolplane-conformance-fixture-key"
    if not os.getenv("TOOLPLANE_CONFORMANCE_API_KEY"):
        os.environ["TOOLPLANE_CONFORMANCE_API_KEY"] = default_api_key

    os.environ.setdefault("TOOLPLANE_ENV_MODE", "development")
    os.environ.setdefault("TOOLPLANE_AUTH_MODE", "fixed")
    os.environ.setdefault(
        "TOOLPLANE_AUTH_FIXED_API_KEY", os.environ["TOOLPLANE_CONFORMANCE_API_KEY"]
    )
    os.environ.setdefault("TOOLPLANE_STORAGE_MODE", "memory")
    os.environ.setdefault("TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND", "1")

    # Multi-instance mode requires a shared durable store: two server processes
    # must see the same request/machine/tool state. In-memory mode gives each
    # process an isolated store, so force Postgres when multi-instance is on.
    # The release-gate workflow provisions Postgres and sets the DSN.
    if multi_instance:
        if not os.getenv("TOOLPLANE_DATABASE_URL"):
            pytest.fail(
                "TOOLPLANE_CONFORMANCE_MULTI_INSTANCE=1 requires TOOLPLANE_DATABASE_URL "
                "(two server replicas must share one Postgres store)",
                pytrace=False,
            )
        os.environ["TOOLPLANE_STORAGE_MODE"] = "postgres"

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

    if not _wait_for_tcp(
        "127.0.0.1",
        grpc_port,
        timeout_seconds=BOOTSTRAP_READINESS_TIMEOUT_SECONDS,
        process=server_process,
    ):
        exit_code = server_process.poll()
        _terminate_process(server_process)
        server_log_handle.close()
        proxy_log_handle.close()
        detail = (
            f" (process exited with code {exit_code})" if exit_code is not None else ""
        )
        _fail_bootstrap(
            f"Conformance bootstrap failed: gRPC server did not become ready on {grpc_port}{detail}.",
            server_log_path,
        )

    # Optional second server instance for multi-instance (active-active)
    # conformance. Both replicas share the same Postgres store so request claim,
    # drain, and requeue behave coherently across them. The second instance gets
    # its own gRPC port and its own metrics port; the adapter layer targets it
    # via TOOLPLANE_CONFORMANCE_GRPC_PORT_B.
    server_b_process = None
    server_b_log_handle = None
    if multi_instance:
        grpc_port_b = _find_free_port()
        metrics_port_b = _find_free_port()
        os.environ["TOOLPLANE_CONFORMANCE_GRPC_PORT_B"] = str(grpc_port_b)

        server_b_log_path = log_dir / "grpc_server_b.log"
        server_b_log_handle = open(server_b_log_path, "w", encoding="utf-8")

        server_b_process = subprocess.Popen(
            [
                "go",
                "run",
                "./cmd/server",
                "--port",
                str(grpc_port_b),
                "--metrics-listen",
                f"127.0.0.1:{metrics_port_b}",
            ],
            cwd=_server_root(),
            stdout=server_b_log_handle,
            stderr=subprocess.STDOUT,
            env=os.environ.copy(),
        )

        if not _wait_for_tcp(
            "127.0.0.1",
            grpc_port_b,
            timeout_seconds=BOOTSTRAP_READINESS_TIMEOUT_SECONDS,
            process=server_b_process,
        ):
            exit_code = server_b_process.poll()
            _terminate_process(server_b_process)
            _terminate_process(server_process)
            server_log_handle.close()
            proxy_log_handle.close()
            if server_b_log_handle:
                server_b_log_handle.close()
            detail = (
                f" (process exited with code {exit_code})"
                if exit_code is not None
                else ""
            )
            _fail_bootstrap(
                f"Conformance bootstrap failed: second gRPC server did not become ready on {grpc_port_b}{detail}.",
                server_b_log_path,
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
    if not _wait_for_http_health(
        health_url,
        timeout_seconds=BOOTSTRAP_READINESS_TIMEOUT_SECONDS,
        process=proxy_process,
    ):
        exit_code = proxy_process.poll()
        _terminate_process(proxy_process)
        _terminate_process(server_process)
        if server_b_process is not None:
            _terminate_process(server_b_process)
        server_log_handle.close()
        proxy_log_handle.close()
        if server_b_log_handle:
            server_b_log_handle.close()
        detail = (
            f" (process exited with code {exit_code})" if exit_code is not None else ""
        )
        _fail_bootstrap(
            f"Conformance bootstrap failed: HTTP gateway did not become ready on {http_port}{detail}.",
            proxy_log_path,
        )

    # Optional MCP facade (toolplane-mcp-gateway) for the mcp transport. It is
    # a thin stateless JSON-RPC layer over the same gRPC backend, so it boots
    # against the primary server instance. Enabled behind
    # TOOLPLANE_CONFORMANCE_MCP=1; when off the mcp transport is skipped.
    mcp_enabled = os.getenv("TOOLPLANE_CONFORMANCE_MCP", "0").strip().lower() in {
        "1",
        "true",
        "yes",
    }
    mcp_process = None
    mcp_log_handle = None
    if mcp_enabled:
        mcp_port = _find_free_port()
        os.environ["TOOLPLANE_CONFORMANCE_MCP_HOST"] = "localhost"
        os.environ["TOOLPLANE_CONFORMANCE_MCP_PORT"] = str(mcp_port)
        os.environ.setdefault("TOOLPLANE_MCP_ALLOW_INSECURE_BACKEND", "1")

        mcp_log_path = log_dir / "mcp_gateway.log"
        mcp_log_handle = open(mcp_log_path, "w", encoding="utf-8")

        mcp_process = subprocess.Popen(
            [
                "go",
                "run",
                "./cmd/mcp-gateway",
                "--listen",
                f":{mcp_port}",
                "--backend",
                f"localhost:{grpc_port}",
            ],
            cwd=_server_root(),
            stdout=mcp_log_handle,
            stderr=subprocess.STDOUT,
            env=os.environ.copy(),
        )

        mcp_health_url = f"http://127.0.0.1:{mcp_port}/health"
        if not _wait_for_http_health(
            mcp_health_url,
            timeout_seconds=BOOTSTRAP_READINESS_TIMEOUT_SECONDS,
            process=mcp_process,
        ):
            exit_code = mcp_process.poll()
            _terminate_process(mcp_process)
            _terminate_process(proxy_process)
            _terminate_process(server_process)
            if server_b_process is not None:
                _terminate_process(server_b_process)
            server_log_handle.close()
            proxy_log_handle.close()
            if server_b_log_handle:
                server_b_log_handle.close()
            mcp_log_handle.close()
            detail = (
                f" (process exited with code {exit_code})"
                if exit_code is not None
                else ""
            )
            _fail_bootstrap(
                f"Conformance bootstrap failed: MCP gateway did not become ready on {mcp_port}{detail}.",
                mcp_log_path,
            )

    try:
        yield
    finally:
        if mcp_process is not None:
            _terminate_process(mcp_process)
        _terminate_process(proxy_process)
        _terminate_process(server_process)
        if server_b_process is not None:
            _terminate_process(server_b_process)
        server_log_handle.close()
        proxy_log_handle.close()
        if server_b_log_handle:
            server_b_log_handle.close()
        if mcp_log_handle is not None:
            mcp_log_handle.close()


# ---------------------------------------------------------------------------
# Session integrity guard.
#
# pytest exits 0 when every test skips, which historically let a broken
# bootstrap (or a mass exception-to-skip conversion) turn the conformance
# signal green while proving nothing. When the suite boots its own environment
# (TOOLPLANE_CONFORMANCE_AUTO_BOOT enabled, the default), a session that
# executed conformance tests but recorded zero passes fails the run.
#
# The rule is global rather than per-transport on purpose: legitimate runs use
# -k filtering (e.g. the release gate's `-k mcp` leg), under which a transport
# can execute only environment-gated skips while another transport carries all
# the passes. Deselected tests produce no reports, and a bootstrap failure
# surfaces as session-fixture errors (failures, not skips), so "ran something,
# passed nothing" is exactly the empty-green signal this guard must catch.
# ---------------------------------------------------------------------------

_TRANSPORT_TOKENS = ("grpc", "http", "mcp")

_outcomes_by_transport = defaultdict(lambda: {"passed": 0, "failed": 0, "skipped": 0})


def _transport_from_nodeid(nodeid: str) -> str:
    match = re.search(r"\[([^\]]+)\]", nodeid)
    if not match:
        return ""
    segments = match.group(1).split("-")
    for segment in segments:
        if segment in _TRANSPORT_TOKENS:
            return segment
    return ""


def pytest_runtest_logreport(report):
    transport = _transport_from_nodeid(report.nodeid)
    if not transport:
        return
    bucket = _outcomes_by_transport[transport]
    if report.when == "call":
        if report.passed:
            bucket["passed"] += 1
        elif report.failed:
            bucket["failed"] += 1
        elif report.skipped:
            bucket["skipped"] += 1
    elif report.when == "setup":
        if report.failed:
            bucket["failed"] += 1
        elif report.skipped:
            bucket["skipped"] += 1


def pytest_sessionfinish(session, exitstatus):
    # Only upgrade a would-be-green session: when the run already failed,
    # errored, or was interrupted, keep the original exit code and failure
    # output instead of overlaying the guard banner on top of it.
    if exitstatus != 0:
        return

    auto_boot = os.getenv(
        "TOOLPLANE_CONFORMANCE_AUTO_BOOT", "1"
    ).strip().lower() not in {
        "0",
        "false",
        "no",
    }
    if not auto_boot:
        return

    totals = {"passed": 0, "failed": 0, "skipped": 0}
    for bucket in _outcomes_by_transport.values():
        for key in totals:
            totals[key] += bucket[key]

    ran = totals["passed"] + totals["failed"] + totals["skipped"]
    if ran == 0 or totals["passed"] > 0:
        return

    reporter = session.config.pluginmanager.get_plugin("terminalreporter")
    if reporter is not None:
        reporter.write_sep("=", "CONFORMANCE INTEGRITY GUARD", red=True)
        for transport in sorted(_outcomes_by_transport):
            bucket = _outcomes_by_transport[transport]
            reporter.write_line(
                f"transport {transport}: {bucket['passed']} passed, "
                f"{bucket['failed']} failed, {bucket['skipped']} skipped"
            )
        reporter.write_line(
            "Refusing to pass a conformance run that executed tests but "
            "recorded zero passing cases. Set "
            "TOOLPLANE_CONFORMANCE_AUTO_BOOT=0 only when targeting an "
            "externally managed server."
        )
    session.exitstatus = 1
