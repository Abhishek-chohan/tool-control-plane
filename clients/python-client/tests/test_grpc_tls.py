import sys
from pathlib import Path


def _ensure_python_client_on_path() -> None:
    current = Path(__file__).resolve()
    python_client_root = current.parents[1]
    toolplane_root = python_client_root / "toolplane"
    for path in (python_client_root, toolplane_root):
        if str(path) not in sys.path:
            sys.path.insert(0, str(path))


_ensure_python_client_on_path()

from toolplane.core.config import ClientConfig
from toolplane.core.connection import ConnectionError, ConnectionManager


class _FakeChannel:
    def close(self):
        return None


def test_connection_manager_uses_secure_channel_when_tls_is_enabled(monkeypatch, tmp_path):
    ca_path = tmp_path / "ca.crt"
    ca_path.write_text("dummy-ca", encoding="utf-8")

    captured = {}

    def fake_secure_channel(target, credentials, options=None):
        captured["target"] = target
        captured["credentials"] = credentials
        captured["options"] = options
        return _FakeChannel()

    monkeypatch.setattr("toolplane.core.connection.grpc.secure_channel", fake_secure_channel)
    monkeypatch.setattr(
        "toolplane.core.connection.grpc.ssl_channel_credentials",
        lambda **kwargs: kwargs,
    )

    manager = ConnectionManager(
        ClientConfig(
            server_host="localhost",
            server_port=9001,
            use_tls=True,
            tls_ca_cert_path=str(ca_path),
            tls_server_name="server",
        )
    )

    channel = manager._create_channel("localhost:9001")

    assert isinstance(channel, _FakeChannel)
    assert captured["target"] == "localhost:9001"
    assert captured["credentials"]["root_certificates"] == b"dummy-ca"
    assert captured["options"] == [
        ("grpc.ssl_target_name_override", "server"),
        ("grpc.default_authority", "server"),
    ]


def test_connection_manager_uses_insecure_channel_when_tls_is_disabled(monkeypatch):
    captured = {}

    def fake_insecure_channel(target, options=None):
        captured["target"] = target
        captured["options"] = options
        return _FakeChannel()

    monkeypatch.setattr("toolplane.core.connection.grpc.insecure_channel", fake_insecure_channel)

    manager = ConnectionManager(ClientConfig(server_host="localhost", server_port=9001))

    channel = manager._create_channel("localhost:9001")

    assert isinstance(channel, _FakeChannel)
    assert captured["target"] == "localhost:9001"
    assert captured["options"] == []


def test_connection_manager_rejects_partial_client_tls_identity(tmp_path):
    cert_path = tmp_path / "client.crt"
    cert_path.write_text("dummy-cert", encoding="utf-8")

    manager = ConnectionManager(
        ClientConfig(
            server_host="localhost",
            server_port=9001,
            use_tls=True,
            tls_cert_path=str(cert_path),
        )
    )

    try:
        manager._create_channel("localhost:9001")
    except ConnectionError as exc:
        assert str(exc) == "TLS client authentication requires both certificate and key files"
        return

    raise AssertionError("expected partial client TLS identity to fail")