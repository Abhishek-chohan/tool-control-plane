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

from toolplane.core.request import RequestManager
from toolplane.http_core.http_request import HTTPRequestManager
from toolplane.toolplane_client import Toolplane
from toolplane.toolplane_http_client import ToolplaneHTTP


class FakeGrpcRequestsStub:
    def __init__(self):
        self.cancel_calls = []

    def CancelRequest(self, request, metadata=None):
        self.cancel_calls.append((request, metadata))
        return type("CancelResponse", (), {"success": True})()


class FakeGrpcConnectionManager:
    def __init__(self):
        self.connected = True
        self.requests_stub = FakeGrpcRequestsStub()

    def ensure_connected(self):
        return None

    def get_metadata(self):
        return [("api_key", "test-key")]


class FakeHTTPConnectionManager:
    def __init__(self):
        self.connected = True
        self.cancel_calls = []

    def ensure_connected(self):
        return None

    def cancel_request(self, session_id, request_id):
        self.cancel_calls.append((session_id, request_id))
        return {"success": True}


class FakeRequestManager:
    def __init__(self):
        self.cancel_calls = []

    def cancel_request(self, session_id, request_id):
        self.cancel_calls.append((session_id, request_id))
        return True


def test_request_manager_cancel_request_calls_grpc_stub():
    connection_manager = FakeGrpcConnectionManager()
    manager = RequestManager(connection_manager)

    success = manager.cancel_request("session-1", "request-1")

    assert success is True
    request, metadata = connection_manager.requests_stub.cancel_calls[0]
    assert request.session_id == "session-1"
    assert request.request_id == "request-1"
    assert metadata == [("api_key", "test-key")]
    manager.shutdown()


def test_http_request_manager_cancel_request_calls_transport():
    connection_manager = FakeHTTPConnectionManager()
    manager = HTTPRequestManager(connection_manager)

    success = manager.cancel_request("session-2", "request-2")

    assert success is True
    assert connection_manager.cancel_calls == [("session-2", "request-2")]
    manager.shutdown()


def test_toolplane_cancel_request_delegates_to_request_manager():
    client = Toolplane.__new__(Toolplane)
    client.connection_manager = type("ConnectionManager", (), {"connected": True})()
    client.request_manager = FakeRequestManager()

    success = client.cancel_request("session-3", "request-3")

    assert success is True
    assert client.request_manager.cancel_calls == [("session-3", "request-3")]


def test_toolplane_http_cancel_request_delegates_to_request_manager():
    client = ToolplaneHTTP.__new__(ToolplaneHTTP)
    client.connection_manager = type("ConnectionManager", (), {"connected": True})()
    client.request_manager = FakeRequestManager()

    success = client.cancel_request("session-4", "request-4")

    assert success is True
    assert client.request_manager.cancel_calls == [("session-4", "request-4")]