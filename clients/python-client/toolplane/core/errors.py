"""Error handling for Toolplane client."""

import json
from typing import Any, Optional

from grpc import StatusCode


class ToolplaneError(Exception):
    """Base exception for Toolplane client errors."""

    pass


class ConnectionError(ToolplaneError):
    """Error related to gRPC connection."""

    pass


class ToolError(ToolplaneError):
    """Error related to tool operations."""

    pass


class SessionError(ToolplaneError):
    """Error related to session operations."""

    pass


class MachineError(ToolplaneError):
    """Error related to machine operations."""

    pass


class RequestError(ToolplaneError):
    """Error related to request operations."""

    pass


class TaskError(ToolplaneError):
    """Error related to task operations."""

    pass


# gRPC codes where re-issuing the same call can plausibly succeed (the server
# or a proxy was momentarily unavailable, or a capacity limit will clear).
# Everything else — bad credentials, bad arguments, state conflicts,
# missing entities — is deterministic: retrying re-runs the same failure.
RETRYABLE_GRPC_CODES = frozenset(
    {StatusCode.UNAVAILABLE, StatusCode.RESOURCE_EXHAUSTED}
)

# The JSON gateway translates gRPC codes to HTTP statuses and carries the
# numeric gRPC code in the error body; this table covers responses whose body
# is not that envelope (bare proxies, load balancers).
_HTTP_STATUS_FALLBACK = {
    400: StatusCode.INVALID_ARGUMENT,
    401: StatusCode.UNAUTHENTICATED,
    403: StatusCode.PERMISSION_DENIED,
    404: StatusCode.NOT_FOUND,
    409: StatusCode.ALREADY_EXISTS,
    422: StatusCode.INVALID_ARGUMENT,
    429: StatusCode.RESOURCE_EXHAUSTED,
    500: StatusCode.INTERNAL,
    502: StatusCode.UNAVAILABLE,
    503: StatusCode.UNAVAILABLE,
    504: StatusCode.DEADLINE_EXCEEDED,
}

# grpc.StatusCode members are keyed by their wire strings ("failed
# precondition"), not by the protocol's numeric codes, so the numbers from
# the gateway error envelope are mapped explicitly.
_GRPC_NUMERIC_CODE = {
    0: StatusCode.OK,
    1: StatusCode.CANCELLED,
    2: StatusCode.UNKNOWN,
    3: StatusCode.INVALID_ARGUMENT,
    4: StatusCode.DEADLINE_EXCEEDED,
    5: StatusCode.NOT_FOUND,
    6: StatusCode.ALREADY_EXISTS,
    7: StatusCode.PERMISSION_DENIED,
    8: StatusCode.RESOURCE_EXHAUSTED,
    9: StatusCode.FAILED_PRECONDITION,
    10: StatusCode.ABORTED,
    11: StatusCode.OUT_OF_RANGE,
    12: StatusCode.UNIMPLEMENTED,
    13: StatusCode.INTERNAL,
    14: StatusCode.UNAVAILABLE,
    15: StatusCode.DATA_LOSS,
    16: StatusCode.UNAUTHENTICATED,
}


class ToolplaneAPIError(ToolplaneError):
    """A server-reported failure carrying a machine-readable status code.

    Attributes:
        code: gRPC status name, e.g. "NOT_FOUND" or "FAILED_PRECONDITION".
        retryable: True when re-issuing the call can plausibly succeed
            (UNAVAILABLE or RESOURCE_EXHAUSTED).
        request_id: The request the failing call operated on, when known.
        status: Last-known request status for poll-style calls, when known.
        details: Raw transport error (gRPC trailing status or HTTP body).
    """

    def __init__(
        self,
        message: str,
        *,
        code: str,
        retryable: bool,
        request_id: Optional[str] = None,
        status: Optional[str] = None,
        details: Any = None,
    ):
        super().__init__(message)
        self.code = code
        self.retryable = retryable
        self.request_id = request_id
        self.status = status
        self.details = details


class ToolplaneNotFoundError(ToolplaneAPIError):
    """The targeted session, request, machine, tool, key, or task is missing."""


class ToolplaneInvalidArgumentError(ToolplaneAPIError):
    """The server rejected the request payload (INVALID_ARGUMENT)."""


class ToolplaneFailedPreconditionError(ToolplaneAPIError):
    """The operation conflicts with server state: a lost claim race, a
    stale lease, a draining machine, a terminal request, or no provider
    registered for the tool (FAILED_PRECONDITION)."""


class ToolplaneResourceExhaustedError(ToolplaneAPIError):
    """A capacity limit was hit (RESOURCE_EXHAUSTED); retryable with backoff."""


class ToolplaneUnauthenticatedError(ToolplaneAPIError):
    """The presented API key is missing, unknown, or revoked
    (UNAUTHENTICATED). Retrying cannot succeed."""


class ToolplanePermissionDeniedError(ToolplaneAPIError):
    """The caller lacks the required capability, or the per-machine
    credential check failed (PERMISSION_DENIED)."""


class ToolplaneAlreadyExistsError(ToolplaneAPIError):
    """The created entity already exists (ALREADY_EXISTS)."""


class ToolplaneTimeoutError(ToolplaneAPIError):
    """The call exceeded its deadline (DEADLINE_EXCEEDED)."""


class ToolplaneUnavailableError(ToolplaneAPIError):
    """The server or a proxy was momentarily unreachable (UNAVAILABLE);
    retryable with backoff."""


class ToolplaneCancelledError(ToolplaneAPIError):
    """The call was cancelled before completing (CANCELLED)."""


class ToolplaneInternalError(ToolplaneAPIError):
    """An unexpected server-side failure, or an unrecognized status code."""


_GRPC_CODE_TO_ERROR = {
    StatusCode.NOT_FOUND: ToolplaneNotFoundError,
    StatusCode.INVALID_ARGUMENT: ToolplaneInvalidArgumentError,
    StatusCode.FAILED_PRECONDITION: ToolplaneFailedPreconditionError,
    StatusCode.OUT_OF_RANGE: ToolplaneInvalidArgumentError,
    StatusCode.RESOURCE_EXHAUSTED: ToolplaneResourceExhaustedError,
    StatusCode.UNAUTHENTICATED: ToolplaneUnauthenticatedError,
    StatusCode.PERMISSION_DENIED: ToolplanePermissionDeniedError,
    StatusCode.ALREADY_EXISTS: ToolplaneAlreadyExistsError,
    StatusCode.DEADLINE_EXCEEDED: ToolplaneTimeoutError,
    StatusCode.UNAVAILABLE: ToolplaneUnavailableError,
    StatusCode.CANCELLED: ToolplaneCancelledError,
    StatusCode.INTERNAL: ToolplaneInternalError,
    StatusCode.UNKNOWN: ToolplaneInternalError,
    StatusCode.UNIMPLEMENTED: ToolplaneInternalError,
    StatusCode.ABORTED: ToolplaneFailedPreconditionError,
}


def api_error_from_rpc_error(
    rpc_error: Any,
    context: str = "",
    request_id: Optional[str] = None,
    status: Optional[str] = None,
) -> ToolplaneAPIError:
    """Translate a grpc.RpcError into a typed ToolplaneAPIError subclass."""
    code = rpc_error.code()
    message = rpc_error.details() or str(rpc_error)
    if context:
        message = f"{context}: {message}"
    error_class = _GRPC_CODE_TO_ERROR.get(code, ToolplaneInternalError)
    return error_class(
        message,
        code=code.name,
        retryable=code in RETRYABLE_GRPC_CODES,
        request_id=request_id,
        status=status,
        details=rpc_error,
    )


def api_error_from_http_response(
    http_status: int,
    body: str,
    context: str = "",
    request_id: Optional[str] = None,
) -> ToolplaneAPIError:
    """Translate an HTTP error response into a typed ToolplaneAPIError.

    The JSON gateway emits ``{"code": <grpc code number>, "message": ...}``
    for unary endpoints and wraps the status in an ``{"error": {...}}``
    frame for streaming endpoints; the numeric code is authoritative when
    present in either shape. Otherwise the HTTP status is mapped through
    _HTTP_STATUS_FALLBACK.
    """
    message = ""
    code: Optional[StatusCode] = None
    try:
        parsed = json.loads(body) if body else None
        if isinstance(parsed, dict):
            envelope = parsed
            nested = parsed.get("error")
            if isinstance(nested, dict):
                envelope = nested
            message = str(envelope.get("message") or "")
            raw_code = envelope.get("code")
            if raw_code is None:
                raw_code = parsed.get("code")
            if isinstance(raw_code, int):
                code = _GRPC_NUMERIC_CODE.get(raw_code)
    except (ValueError, TypeError):
        pass

    if code is None:
        code = _HTTP_STATUS_FALLBACK.get(http_status, StatusCode.UNKNOWN)
    if not message:
        message = f"HTTP {http_status} {body or ''}".strip()
    if context:
        message = f"{context}: {message}"

    error_class = _GRPC_CODE_TO_ERROR.get(code, ToolplaneInternalError)
    return error_class(
        message,
        code=code.name,
        retryable=code in RETRYABLE_GRPC_CODES,
        request_id=request_id,
        details=body,
    )


def status_for_wire(name):
    """Map a friendly status ("done", "pending") onto the v1 enum name
    ("REQUEST_STATUS_DONE") used on requests with status filters. An empty
    value maps to 0 (UNSPECIFIED), which gRPC accepts and means "no filter".
    """
    if not isinstance(name, str) or not name:
        return 0
    return "REQUEST_STATUS_" + name.upper()


def normalize_status_name(status):
    """Map a v1 enum status name (e.g. "REQUEST_STATUS_DONE") onto the
    friendly lowercase form ("done") clients have always seen. Passes
    through anything else unchanged."""
    if not isinstance(status, str):
        return status
    for prefix in ("REQUEST_STATUS_", "TASK_STATUS_"):
        if status.startswith(prefix):
            return status[len(prefix) :].lower()
    return status
