"""Typed error translation tests (no server required)."""

import grpc

from toolplane.core.errors import (
    RETRYABLE_GRPC_CODES,
    ToolplaneFailedPreconditionError,
    ToolplaneInvalidArgumentError,
    ToolplaneNotFoundError,
    ToolplaneResourceExhaustedError,
    ToolplaneUnauthenticatedError,
    ToolplaneUnavailableError,
    api_error_from_http_response,
    api_error_from_rpc_error,
)


class _FakeRpcError(grpc.RpcError):
    def __init__(self, code, details=""):
        super().__init__()
        self._code = code
        self._details = details

    def code(self):
        return self._code

    def details(self):
        return self._details


def test_unauthenticated_is_not_retryable():
    err = api_error_from_rpc_error(
        _FakeRpcError(grpc.StatusCode.UNAUTHENTICATED, "api key not found"),
        context="invoke echo",
    )
    assert isinstance(err, ToolplaneUnauthenticatedError)
    assert err.code == "UNAUTHENTICATED"
    assert err.retryable is False
    assert "invoke echo" in str(err)
    assert "api key not found" in str(err)


def test_unavailable_is_retryable():
    err = api_error_from_rpc_error(
        _FakeRpcError(grpc.StatusCode.UNAVAILABLE, "connection refused")
    )
    assert isinstance(err, ToolplaneUnavailableError)
    assert err.retryable is True


def test_resource_exhausted_is_retryable():
    err = api_error_from_rpc_error(
        _FakeRpcError(grpc.StatusCode.RESOURCE_EXHAUSTED, "machine at capacity")
    )
    assert isinstance(err, ToolplaneResourceExhaustedError)
    assert err.retryable is True


def test_failed_precondition_covers_claim_and_lease_conflicts():
    err = api_error_from_rpc_error(
        _FakeRpcError(grpc.StatusCode.FAILED_PRECONDITION, "request not claimable"),
        request_id="req_1",
        status="pending",
    )
    assert isinstance(err, ToolplaneFailedPreconditionError)
    assert err.retryable is False
    assert err.request_id == "req_1"
    assert err.status == "pending"


def test_not_found_and_invalid_argument():
    missing = api_error_from_rpc_error(
        _FakeRpcError(grpc.StatusCode.NOT_FOUND, "request req_x not found")
    )
    assert isinstance(missing, ToolplaneNotFoundError)

    bad = api_error_from_rpc_error(
        _FakeRpcError(grpc.StatusCode.INVALID_ARGUMENT, "capabilities required")
    )
    assert isinstance(bad, ToolplaneInvalidArgumentError)


def test_gateway_body_code_is_authoritative():
    # grpc-gateway maps FAILED_PRECONDITION to HTTP 400 but carries the
    # numeric gRPC code (9) in the body; the body wins over the status.
    err = api_error_from_http_response(
        400, '{"code": 9, "message": "lease conflict"}'
    )
    assert isinstance(err, ToolplaneFailedPreconditionError)
    assert err.code == "FAILED_PRECONDITION"
    assert err.retryable is False
    assert "lease conflict" in str(err)


def test_streaming_error_frame_envelope_is_unwrapped():
    # Streaming endpoints render errors as a nested {"error": {...}} frame
    # carrying the status, still with the mapped HTTP status code on the
    # wire (OUT_OF_RANGE -> 400).
    err = api_error_from_http_response(
        400, '{"error": {"code": 11, "message": "replay window expired"}}'
    )
    assert err.code == "OUT_OF_RANGE"
    assert err.retryable is False
    assert "replay window expired" in str(err)


def test_http_status_fallback_without_gateway_body():
    err = api_error_from_http_response(503, "")
    assert isinstance(err, ToolplaneUnavailableError)
    assert err.retryable is True

    missing = api_error_from_http_response(404, "no json")
    assert isinstance(missing, ToolplaneNotFoundError)


def test_retryable_set_is_transport_and_capacity_only():
    assert RETRYABLE_GRPC_CODES == frozenset(
        {grpc.StatusCode.UNAVAILABLE, grpc.StatusCode.RESOURCE_EXHAUSTED}
    )
