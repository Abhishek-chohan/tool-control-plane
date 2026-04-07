from typing import Any, Dict, List


def _assert_non_empty_string(
    value: Any, label: str, case_id: str, transport: str
) -> None:
    if not isinstance(value, str) or not value.strip():
        raise AssertionError(f"[{transport}] {case_id}: {label} must be non-empty")


def _assert_field_equals(
    payload: Dict[str, Any],
    field: str,
    expected: Any,
    label: str,
    case_id: str,
    transport: str,
) -> None:
    actual = payload.get(field)
    if actual != expected:
        raise AssertionError(
            f"[{transport}] {case_id}: {label}.{field} mismatch. expected={expected}, actual={actual}"
        )


def assert_session_id_non_empty(session_id: str, case_id: str, transport: str) -> None:
    _assert_non_empty_string(session_id, "session_id", case_id, transport)


def assert_session_context_present(context: Any, case_id: str, transport: str) -> None:
    if context is None:
        raise AssertionError(f"[{transport}] {case_id}: session context should be available")


def assert_sessions_array(response: Dict[str, Any], case_id: str, transport: str) -> None:
    sessions = response.get("sessions")
    if not isinstance(sessions, list):
        raise AssertionError(f"[{transport}] {case_id}: response.sessions must be a list")


def assert_contains_session(
    response: Dict[str, Any], session_id: str, case_id: str, transport: str
) -> None:
    sessions = response.get("sessions", [])
    session_ids = {entry.get("id") for entry in sessions if isinstance(entry, dict)}
    if session_id not in session_ids:
        raise AssertionError(
            f"[{transport}] {case_id}: expected session_id '{session_id}' in listed sessions"
        )


def assert_unary_result(
    result: Any, expected_result: Dict[str, Any], case_id: str, transport: str
) -> None:
    if result != expected_result:
        raise AssertionError(
            f"[{transport}] {case_id}: unary result mismatch. expected={expected_result}, actual={result}"
        )


def assert_stream_chunks(
    chunks: List[Any], expected_chunks: List[Any], case_id: str, transport: str
) -> None:
    if chunks != expected_chunks:
        raise AssertionError(
            f"[{transport}] {case_id}: stream chunks mismatch. expected={expected_chunks}, actual={chunks}"
        )


def assert_final_marker(saw_final: bool, case_id: str, transport: str) -> None:
    if not saw_final:
        raise AssertionError(f"[{transport}] {case_id}: expected final stream marker")


def assert_session_field_equals(
    session: Dict[str, Any],
    field: str,
    expected: Any,
    case_id: str,
    transport: str,
) -> None:
    _assert_field_equals(session, field, expected, "session", case_id, transport)


def assert_tool_id_non_empty(tool_id: str, case_id: str, transport: str) -> None:
    _assert_non_empty_string(tool_id, "tool.id", case_id, transport)


def assert_tool_field_equals(
    tool: Dict[str, Any],
    field: str,
    expected: Any,
    case_id: str,
    transport: str,
) -> None:
    _assert_field_equals(tool, field, expected, "tool", case_id, transport)


def assert_tool_list_contains(
    tools: List[Dict[str, Any]], tool_id: str, case_id: str, transport: str
) -> None:
    tool_ids = {entry.get("id") for entry in tools if isinstance(entry, dict)}
    if tool_id not in tool_ids:
        raise AssertionError(
            f"[{transport}] {case_id}: expected tool '{tool_id}' in listed tools"
        )


def assert_tool_list_excludes(
    tools: List[Dict[str, Any]], tool_id: str, case_id: str, transport: str
) -> None:
    tool_ids = {entry.get("id") for entry in tools if isinstance(entry, dict)}
    if tool_id in tool_ids:
        raise AssertionError(
            f"[{transport}] {case_id}: did not expect tool '{tool_id}' in listed tools"
        )


def assert_request_id_non_empty(request_id: str, case_id: str, transport: str) -> None:
    _assert_non_empty_string(request_id, "request_id", case_id, transport)


def assert_request_status(
    request: Dict[str, Any], expected_status: str, case_id: str, transport: str
) -> None:
    _assert_field_equals(request, "status", expected_status, "request", case_id, transport)


def assert_request_field_equals(
    request: Dict[str, Any],
    field: str,
    expected: Any,
    case_id: str,
    transport: str,
) -> None:
    _assert_field_equals(request, field, expected, "request", case_id, transport)


def assert_request_field_non_empty(
    request: Dict[str, Any], field: str, case_id: str, transport: str
) -> None:
    _assert_non_empty_string(request.get(field), f"request.{field}", case_id, transport)


def assert_request_list_contains(
    requests: List[Dict[str, Any]], request_id: str, case_id: str, transport: str
) -> None:
    request_ids = {entry.get("id") for entry in requests if isinstance(entry, dict)}
    if request_id not in request_ids:
        raise AssertionError(
            f"[{transport}] {case_id}: expected request_id '{request_id}' in listed requests"
        )


def assert_chunk_window_field_equals(
    window: Dict[str, Any],
    field: str,
    expected: Any,
    case_id: str,
    transport: str,
) -> None:
    _assert_field_equals(window, field, expected, "chunk_window", case_id, transport)


def assert_chunk_window_length(
    window: Dict[str, Any], expected_length: int, case_id: str, transport: str
) -> None:
    chunks = window.get("chunks")
    if not isinstance(chunks, list) or len(chunks) != expected_length:
        actual_length = len(chunks) if isinstance(chunks, list) else None
        raise AssertionError(
            f"[{transport}] {case_id}: chunk_window length mismatch. expected={expected_length}, actual={actual_length}"
        )


def assert_chunk_window_edge(
    window: Dict[str, Any],
    edge: str,
    expected: Any,
    case_id: str,
    transport: str,
) -> None:
    chunks = window.get("chunks")
    if not isinstance(chunks, list) or not chunks:
        raise AssertionError(
            f"[{transport}] {case_id}: chunk_window.chunks must be a non-empty list"
        )

    actual = chunks[0] if edge == "first" else chunks[-1]
    if actual != expected:
        raise AssertionError(
            f"[{transport}] {case_id}: {edge} retained chunk mismatch. expected={expected}, actual={actual}"
        )


def assert_error_code_equals(
    payload: Dict[str, Any], expected: str, case_id: str, transport: str
) -> None:
    _assert_field_equals(payload, "errorCode", expected, "result", case_id, transport)


def assert_machine_id_non_empty(
    machine: Dict[str, Any], case_id: str, transport: str
) -> None:
    _assert_non_empty_string(machine.get("id"), "machine.id", case_id, transport)


def assert_machine_field_equals(
    machine: Dict[str, Any],
    field: str,
    expected: Any,
    case_id: str,
    transport: str,
) -> None:
    _assert_field_equals(machine, field, expected, "machine", case_id, transport)


def assert_machine_list_contains(
    machines: List[Dict[str, Any]], machine_id: str, case_id: str, transport: str
) -> None:
    machine_ids = {entry.get("id") for entry in machines if isinstance(entry, dict)}
    if machine_id not in machine_ids:
        raise AssertionError(
            f"[{transport}] {case_id}: expected machine '{machine_id}' in listed machines"
        )


def assert_machine_list_excludes(
    machines: List[Dict[str, Any]], machine_id: str, case_id: str, transport: str
) -> None:
    machine_ids = {entry.get("id") for entry in machines if isinstance(entry, dict)}
    if machine_id in machine_ids:
        raise AssertionError(
            f"[{transport}] {case_id}: did not expect machine '{machine_id}' in listed machines"
        )


def assert_api_key_id_non_empty(
    api_key: Dict[str, Any], case_id: str, transport: str
) -> None:
    _assert_non_empty_string(api_key.get("id"), "api_key.id", case_id, transport)


def assert_api_key_value_non_empty(
    api_key: Dict[str, Any], case_id: str, transport: str
) -> None:
    _assert_non_empty_string(api_key.get("key"), "api_key.key", case_id, transport)


def assert_api_key_value_empty(
    api_key: Dict[str, Any], case_id: str, transport: str
) -> None:
    if api_key.get("key", "") != "":
        raise AssertionError(
            f"[{transport}] {case_id}: api_key.key must be empty in listed api keys"
        )


def assert_api_key_preview_non_empty(
    api_key: Dict[str, Any], case_id: str, transport: str
) -> None:
    _assert_non_empty_string(
        api_key.get("key_preview"), "api_key.key_preview", case_id, transport
    )


def assert_api_key_capabilities_equal(
    api_key: Dict[str, Any], expected: List[str], case_id: str, transport: str
) -> None:
    actual = api_key.get("capabilities")
    if actual != expected:
        raise AssertionError(
            f"[{transport}] {case_id}: api_key.capabilities mismatch. expected={expected}, actual={actual}"
        )


def assert_api_key_field_equals(
    api_key: Dict[str, Any],
    field: str,
    expected: Any,
    case_id: str,
    transport: str,
) -> None:
    _assert_field_equals(api_key, field, expected, "api_key", case_id, transport)


def assert_api_key_list_contains(
    api_keys: List[Dict[str, Any]], key_id: str, case_id: str, transport: str
) -> None:
    key_ids = {entry.get("id") for entry in api_keys if isinstance(entry, dict)}
    if key_id not in key_ids:
        raise AssertionError(
            f"[{transport}] {case_id}: expected api key '{key_id}' in listed api keys"
        )


def assert_api_key_list_excludes(
    api_keys: List[Dict[str, Any]], key_id: str, case_id: str, transport: str
) -> None:
    key_ids = {entry.get("id") for entry in api_keys if isinstance(entry, dict)}
    if key_id in key_ids:
        raise AssertionError(
            f"[{transport}] {case_id}: did not expect api key '{key_id}' in listed api keys"
        )


def assert_success_true(success: bool, label: str, case_id: str, transport: str) -> None:
    if success is not True:
        raise AssertionError(f"[{transport}] {case_id}: expected {label} to be true")
