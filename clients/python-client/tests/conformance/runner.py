import json
import os
import sys
import threading
import time
from pathlib import Path
from typing import Any, Dict, List


def _ensure_python_client_on_path() -> None:
    current = Path(__file__).resolve()
    python_client_root = current.parents[2]
    if str(python_client_root) not in sys.path:
        sys.path.insert(0, str(python_client_root))


_ensure_python_client_on_path()

from .adapters.grpc_adapter import GrpcConformanceAdapter
from .adapters.http_adapter import HttpConformanceAdapter
from .adapters.mcp_adapter import McpConformanceAdapter
from .assertions import (
    assert_api_key_capabilities_equal,
    assert_api_key_field_equals,
    assert_api_key_id_non_empty,
    assert_api_key_list_contains,
    assert_api_key_list_excludes,
    assert_api_key_preview_non_empty,
    assert_api_key_value_empty,
    assert_api_key_value_non_empty,
    assert_chunk_window_edge,
    assert_chunk_window_field_equals,
    assert_chunk_window_length,
    assert_contains_session,
    assert_error_code_equals,
    assert_final_marker,
    assert_machine_field_equals,
    assert_machine_id_non_empty,
    assert_machine_list_contains,
    assert_machine_list_excludes,
    assert_request_field_equals,
    assert_request_field_non_empty,
    assert_request_id_non_empty,
    assert_request_list_contains,
    assert_request_status,
    assert_session_context_present,
    assert_session_field_equals,
    assert_session_id_non_empty,
    assert_sessions_array,
    assert_stream_chunks,
    assert_success_true,
    assert_tool_field_equals,
    assert_tool_id_non_empty,
    assert_tool_list_contains,
    assert_tool_list_excludes,
    assert_unary_result,
)


SUPPORTED_FEATURES = {
    "session_create",
    "session_list",
    "invoke_unary",
    "invoke_stream",
    "tool_discovery",
    "session_update",
    "request_create",
    "request_recovery",
    "api_key_lifecycle",
    "machine_lifecycle",
    "provider_runtime",
    "multi_instance",
    "mcp_tasks",
}

# Features the MCP transport can exercise: tool-plane operations only. The
# management-plane fixtures (sessions, machines, API keys, request recovery)
# stay on the grpc/http transports.
MCP_TRANSPORT_FEATURES = {
    "invoke_unary",
    "invoke_stream",
    "mcp_tasks",
}


def _number_value(value: Any, fallback: int) -> int:
    if isinstance(value, bool):
        return int(value)
    if isinstance(value, int):
        return value
    if isinstance(value, str):
        try:
            return int(value)
        except ValueError:
            return fallback
    return fallback


def _wait_for_running_request(
    adapter, session_id: str, tool_name: str, case_id: str, transport: str
) -> None:
    deadline = time.time() + 5.0
    request_filter = {
        "list_status": "running",
        "tool_name_filter": tool_name,
        "limit": 20,
    }

    while time.time() < deadline:
        requests = adapter.list_requests(session_id, request_filter)
        if requests:
            return
        time.sleep(0.1)

    raise AssertionError(
        f"[{transport}] {case_id}: timed out waiting for running request for tool '{tool_name}'"
    )


def _wait_for_chunk_window_progress(
    adapter,
    session_id: str,
    request_id: str,
    minimum_next_seq: int,
    case_id: str,
    transport: str,
) -> None:
    deadline = time.time() + 10.0
    while time.time() < deadline:
        chunk_window = adapter.get_request_chunks_window(session_id, request_id)
        if _number_value(chunk_window.get("nextSeq"), 0) >= minimum_next_seq:
            return

        request_status = adapter.get_request_status(session_id, request_id)
        if request_status.get("status") == "failure":
            raise AssertionError(
                f"[{transport}] {case_id}: request {request_id} failed while waiting for chunk window progress"
            )

        time.sleep(0.1)

    raise AssertionError(
        f"[{transport}] {case_id}: timed out waiting for retained chunk window to reach next_seq {minimum_next_seq}"
    )


def _execute_request_recovery_case(
    adapter,
    session_id: str,
    request: Dict[str, Any],
    expected: Dict[str, Any],
    case_id: str,
    transport: str,
) -> None:
    tool_name = request["tool_name"]
    adapter.register_stream_tool(
        session_id=session_id,
        tool_name=tool_name,
        description=request.get("tool_description", "conformance request recovery tool"),
    )

    request_id = adapter.start_streaming_request(
        session_id, tool_name, request.get("params", {})
    )
    assert_request_id_non_empty(request_id, case_id, transport)

    has_resume = "resume_from_seq" in request
    has_mid_stream_gate = "wait_for_next_seq_at_least" in request
    resume_from_seq = _number_value(request.get("resume_from_seq"), 0)

    if has_mid_stream_gate:
        _wait_for_chunk_window_progress(
            adapter,
            session_id,
            request_id,
            _number_value(request.get("wait_for_next_seq_at_least"), 0),
            case_id,
            transport,
        )

    request_status = adapter.wait_for_request_completion(session_id, request_id)
    if "final_status_equals" in expected:
        assert_request_status(
            request_status, expected["final_status_equals"], case_id, transport
        )

    chunk_window = adapter.get_request_chunks_window(session_id, request_id)
    if "ordered_chunks" in expected:
        assert_stream_chunks(
            chunk_window.get("chunks", []), expected["ordered_chunks"], case_id, transport
        )
    if "chunk_count_equals" in expected:
        assert_chunk_window_length(
            chunk_window, _number_value(expected["chunk_count_equals"], 0), case_id, transport
        )
    if "start_seq_equals" in expected:
        assert_chunk_window_field_equals(
            chunk_window, "startSeq", expected["start_seq_equals"], case_id, transport
        )
    if "next_seq_equals" in expected:
        assert_chunk_window_field_equals(
            chunk_window, "nextSeq", expected["next_seq_equals"], case_id, transport
        )
    if "first_chunk_equals" in expected:
        assert_chunk_window_edge(
            chunk_window,
            "first",
            expected["first_chunk_equals"],
            case_id,
            transport,
        )
    if "last_chunk_equals" in expected:
        assert_chunk_window_edge(
            chunk_window,
            "last",
            expected["last_chunk_equals"],
            case_id,
            transport,
        )

    resume_result = None
    if has_resume:
        resume_result = adapter.resume_stream(request_id, resume_from_seq)

    if resume_result is not None:
        if "resume_ordered_chunks" in expected:
            assert_stream_chunks(
                resume_result.get("chunks", []),
                expected["resume_ordered_chunks"],
                case_id,
                transport,
            )
        if expected.get("final_marker", False):
            assert_final_marker(resume_result.get("sawFinal", False), case_id, transport)
        if "final_seq_equals" in expected:
            assert_chunk_window_field_equals(
                {"finalSeq": resume_result.get("finalSeq")},
                "finalSeq",
                expected["final_seq_equals"],
                case_id,
                transport,
            )
        if "resume_error_code_equals" in expected:
            assert_error_code_equals(
                resume_result, expected["resume_error_code_equals"], case_id, transport
            )


def _execute_multi_instance_case(
    case_id: str,
    request: Dict[str, Any],
    expected: Dict[str, Any],
    transport: str,
    user_id: str,
) -> None:
    """End-to-end two-process multi-instance proof.

    Creates a session, registers a tool, and creates a request through server
    instance A; then, through a SEPARATE adapter connected to server instance B
    (a different process sharing the same Postgres store), proves the request is
    visible and inspectable. This is the process-level counterpart to the Go-side
    TestActiveActive_* proofs: it exercises two real `toolplane-server` processes
    against one Postgres store.

    Requires the conformance bootstrap to have booted the second instance
    (TOOLPLANE_CONFORMANCE_MULTI_INSTANCE=1 + TOOLPLANE_DATABASE_URL).
    """
    if not os.getenv("TOOLPLANE_CONFORMANCE_GRPC_PORT_B"):
        raise AssertionError(
            f"[{transport}] {case_id}: multi-instance case requires "
            "TOOLPLANE_CONFORMANCE_GRPC_PORT_B (set TOOLPLANE_CONFORMANCE_MULTI_INSTANCE=1)"
        )

    tool_name = request["tool_name"]
    session_request = {
        "user_id": user_id,
        "name": request.get("name", "conformance-multi-instance"),
        "description": request.get("description", "multi-instance conformance session"),
        "namespace": request.get("namespace", "conformance"),
    }

    adapter_a = _adapter_for_instance(transport, user_id, "a")
    adapter_b = _adapter_for_instance(transport, user_id, "b")
    try:
        adapter_a.connect()
        adapter_b.connect()

        # Instance A: create session, register tool, create request.
        session_id = adapter_a.create_session(session_request)
        assert_session_id_non_empty(session_id, case_id, transport)

        adapter_a.register_unary_echo_tool(
            session_id=session_id,
            tool_name=tool_name,
            description=request.get("tool_description", "conformance multi-instance tool"),
        )
        request_id = adapter_a.create_request(
            session_id, tool_name, request.get("params", {})
        )
        assert_request_id_non_empty(request_id, case_id, transport)

        # Instance B attaches to the session created on A. This resolves the
        # session via the server's GetSession RPC, which (with the store
        # read-through) works across replicas. All subsequent B operations
        # (request read, second request create) go through this attached context.
        adapter_b.attach_session(session_id)

        # Instance B: prove the request created on A is visible through B's
        # store-backed read path. This is the core multi-instance guarantee: a
        # request created on one replica is inspectable on another.
        if expected.get("request_visible_on_instance_b", False):
            b_status = adapter_b.get_request_status(session_id, request_id)
            if not b_status or not b_status.get("id"):
                raise AssertionError(
                    f"[{transport}] {case_id}: request {request_id} created on "
                    "instance A was not visible on instance B (cross-instance "
                    "visibility failed)"
                )
            if b_status.get("id") != request_id:
                raise AssertionError(
                    f"[{transport}] {case_id}: instance B returned a different "
                    f"request id {b_status.get('id')!r}, expected {request_id!r}"
                )

        # Instance B: a second request created for the same tool must get a
        # distinct id and be independently visible (no cross-instance collision).
        if expected.get("second_request_distinct", False):
            second_id = adapter_b.create_request(
                session_id, tool_name, request.get("params", {})
            )
            if not second_id or second_id == request_id:
                raise AssertionError(
                    f"[{transport}] {case_id}: second request id {second_id!r} "
                    f"collided with the first {request_id!r}"
                )

        # Instance A: both requests are visible from A's read path.
        if expected.get("both_requests_listed_on_a", False):
            listed = adapter_a.list_requests(session_id, {"limit": 20})
            ids = {entry.get("id") for entry in listed if isinstance(entry, dict)}
            if request_id not in ids:
                raise AssertionError(
                    f"[{transport}] {case_id}: request {request_id} missing "
                    f"from instance A listing {ids}"
                )
    finally:
        adapter_a.close()
        adapter_b.close()


def get_repo_root() -> Path:
    return Path(__file__).resolve().parents[4]


def load_cases() -> List[Dict[str, Any]]:
    case_dir = get_repo_root() / "conformance" / "cases"
    cases: List[Dict[str, Any]] = []
    for file_path in sorted(case_dir.glob("*.json")):
        with open(file_path, "r", encoding="utf-8") as handle:
            case_obj = json.load(handle)
            _validate_case_shape(case_obj, str(file_path))
            cases.append(case_obj)
    if not cases:
        raise RuntimeError(f"No conformance cases found in {case_dir}")
    return cases


def _validate_case_shape(case_obj: Dict[str, Any], source: str) -> None:
    required_fields = ["id", "feature", "description", "request", "expected"]
    for field in required_fields:
        if field not in case_obj:
            raise ValueError(f"Case {source} missing required field '{field}'")
    if case_obj["feature"] not in SUPPORTED_FEATURES:
        raise ValueError(
            f"Case {source} has unsupported feature '{case_obj['feature']}'"
        )


def _adapter_for_transport(transport: str, user_id: str):
    api_key = os.getenv("TOOLPLANE_CONFORMANCE_API_KEY", "")
    if transport == "grpc":
        host = os.getenv("TOOLPLANE_CONFORMANCE_GRPC_HOST", "localhost")
        port = int(os.getenv("TOOLPLANE_CONFORMANCE_GRPC_PORT", "50051"))
        return GrpcConformanceAdapter(host=host, port=port, user_id=user_id, api_key=api_key)
    if transport == "http":
        host = os.getenv("TOOLPLANE_CONFORMANCE_HTTP_HOST", "localhost")
        port = int(os.getenv("TOOLPLANE_CONFORMANCE_HTTP_PORT", "8080"))
        return HttpConformanceAdapter(host=host, port=port, user_id=user_id, api_key=api_key)
    if transport == "mcp":
        if not os.getenv("TOOLPLANE_CONFORMANCE_MCP_PORT"):
            raise RuntimeError(
                "TOOLPLANE_CONFORMANCE_MCP_PORT not set; mcp conformance requires "
                "TOOLPLANE_CONFORMANCE_MCP=1 so the bootstrap boots cmd/mcp-gateway"
            )
        mcp_host = os.getenv("TOOLPLANE_CONFORMANCE_MCP_HOST", "localhost")
        mcp_port = int(os.getenv("TOOLPLANE_CONFORMANCE_MCP_PORT", "8081"))
        proxy_host = os.getenv("TOOLPLANE_CONFORMANCE_HTTP_HOST", "localhost")
        proxy_port = int(os.getenv("TOOLPLANE_CONFORMANCE_HTTP_PORT", "8080"))
        return McpConformanceAdapter(
            mcp_host=mcp_host,
            mcp_port=mcp_port,
            proxy_host=proxy_host,
            proxy_port=proxy_port,
            user_id=user_id,
            api_key=api_key,
        )
    raise ValueError(f"Unsupported transport: {transport}")


def _adapter_for_instance(transport: str, user_id: str, instance: str):
    """Build an adapter targeting a specific server instance.

    instance="a" targets the primary server (TOOLPLANE_CONFORMANCE_GRPC_PORT).
    instance="b" targets the optional second server
    (TOOLPLANE_CONFORMANCE_GRPC_PORT_B), used by the multi-instance feature.
    The second instance is only available when the conformance bootstrap booted
    it behind TOOLPLANE_CONFORMANCE_MULTI_INSTANCE=1.
    """
    api_key = os.getenv("TOOLPLANE_CONFORMANCE_API_KEY", "")
    if transport == "grpc":
        host = os.getenv("TOOLPLANE_CONFORMANCE_GRPC_HOST", "localhost")
        if instance == "b":
            port_b = os.getenv("TOOLPLANE_CONFORMANCE_GRPC_PORT_B")
            if not port_b:
                raise RuntimeError(
                    "TOOLPLANE_CONFORMANCE_GRPC_PORT_B not set; multi-instance "
                    "conformance requires TOOLPLANE_CONFORMANCE_MULTI_INSTANCE=1"
                )
            return GrpcConformanceAdapter(
                host=host, port=int(port_b), user_id=user_id, api_key=api_key
            )
        port = int(os.getenv("TOOLPLANE_CONFORMANCE_GRPC_PORT", "50051"))
        return GrpcConformanceAdapter(host=host, port=port, user_id=user_id, api_key=api_key)
    raise ValueError(
        f"Multi-instance conformance targets gRPC server instances directly; "
        f"transport {transport!r} is not supported for instance targeting"
    )


def _execute_mcp_tasks_case(
    adapter,
    session_id: str,
    request: Dict[str, Any],
    expected: Dict[str, Any],
    case_id: str,
    transport: str,
) -> None:
    """Exercise the MCP Tasks extension end to end through the facade.

    Flow: register a slow streaming tool, start the provider runtime, then
    1) tools/call with the Tasks capability advertised must return an
       immediate task handle (resultType "task", status working);
    2) tasks/get must surface the running task and, while chunks are being
       appended, the dev.toolplane/chunks cursor window; replaying with
       last_seq must not redeliver already-seen chunks;
    3) the task must reach the expected terminal status with the tool result
       embedded as a CallToolResult;
    4) a second in-flight task cancelled via tasks/cancel must reach the
       cancelled state.
    """
    from .adapters.mcp_adapter import CHUNKS_META_KEY

    tool_name = str(request["tool_name"])
    params = request.get("params", {})

    adapter.register_tasks_tool(
        session_id,
        tool_name,
        request.get("tool_description", "MCP tasks conformance tool"),
    )
    adapter.start_provider_runtime(session_id)

    handle = adapter.mcp_call_tool(session_id, tool_name, params, with_tasks=True)
    if handle.get("resultType") != "task":
        raise AssertionError(
            f"[{transport}] {case_id}: expected resultType 'task', got {handle.get('resultType')!r}"
        )
    task_id = str(handle.get("taskId", ""))
    if not task_id:
        raise AssertionError(f"[{transport}] {case_id}: task handle missing taskId")
    if handle.get("status") != expected.get("handle_status", "working"):
        raise AssertionError(
            f"[{transport}] {case_id}: handle status {handle.get('status')!r} "
            f"!= expected {expected.get('handle_status')!r}"
        )

    # Poll while the task is in flight, watching for the chunk cursor window.
    saw_chunk_window = False
    deadline = time.time() + 30.0
    snapshot: Dict[str, Any] = {}
    while time.time() < deadline:
        snapshot = adapter.mcp_tasks_get(session_id, task_id)
        status = snapshot.get("status")
        chunks_meta = (snapshot.get("_meta") or {}).get(CHUNKS_META_KEY) or {}
        window_chunks = list(chunks_meta.get("chunks") or [])
        if window_chunks:
            saw_chunk_window = True
            next_seq = int(chunks_meta.get("nextSeq", 0))
            replay = adapter.mcp_tasks_get(session_id, task_id, last_seq=next_seq)
            replay_chunks = list(
                ((replay.get("_meta") or {}).get(CHUNKS_META_KEY) or {}).get("chunks") or []
            )
            redelivered = [chunk for chunk in replay_chunks if chunk in window_chunks]
            if redelivered:
                raise AssertionError(
                    f"[{transport}] {case_id}: replay from last_seq={next_seq} "
                    f"redelivered chunks {redelivered}"
                )
        if status in {"completed", "failed", "cancelled"}:
            break
        time.sleep(0.2)

    final = adapter.tasks_poll_until_terminal(session_id, task_id)
    terminal_status = final.get("status")
    if terminal_status != expected.get("terminal_status", "completed"):
        raise AssertionError(
            f"[{transport}] {case_id}: terminal status {terminal_status!r} "
            f"!= expected {expected.get('terminal_status')!r}"
        )
    if expected.get("chunk_cursor_observed", False) and not saw_chunk_window:
        raise AssertionError(
            f"[{transport}] {case_id}: never observed a chunk window on tasks/get"
        )

    if terminal_status == "completed":
        result = final.get("result") or {}
        structured = result.get("structuredContent")
        expected_result = expected.get("result_equals")
        if expected_result is not None and structured != expected_result:
            raise AssertionError(
                f"[{transport}] {case_id}: task result {structured!r} != expected {expected_result!r}"
            )

    # Cancellation path: start a second task and cancel it mid-flight.
    cancel_handle = adapter.mcp_call_tool(session_id, tool_name, params, with_tasks=True)
    if cancel_handle.get("resultType") != "task":
        raise AssertionError(
            f"[{transport}] {case_id}: second tools/call did not return a task handle "
            f"(resultType {cancel_handle.get('resultType')!r}); cannot exercise tasks/cancel"
        )
    cancel_task_id = str(cancel_handle.get("taskId", ""))
    if not cancel_task_id:
        raise AssertionError(
            f"[{transport}] {case_id}: second tools/call returned a task handle with an "
            f"empty taskId; cannot exercise tasks/cancel"
        )
    ack = adapter.mcp_tasks_cancel(session_id, cancel_task_id)
    if ack.get("resultType") != "complete":
        raise AssertionError(
            f"[{transport}] {case_id}: tasks/cancel ack resultType {ack.get('resultType')!r}"
        )
    cancelled = adapter.tasks_poll_until_terminal(session_id, cancel_task_id)
    if cancelled.get("status") != expected.get("cancel_status", "cancelled"):
        raise AssertionError(
            f"[{transport}] {case_id}: cancelled task status {cancelled.get('status')!r} "
            f"!= expected {expected.get('cancel_status')!r}"
        )


def execute_case(case_obj: Dict[str, Any], transport: str) -> None:
    case_id = case_obj["id"]
    feature = case_obj["feature"]
    request = case_obj["request"]
    expected = case_obj["expected"]
    user_id = request.get("user_id", os.getenv("TOOLPLANE_CONFORMANCE_USER_ID", "conformance-user"))

    adapter = _adapter_for_transport(transport, user_id)
    session_id = ""
    try:
        adapter.connect()

        if feature == "session_create":
            session_id = adapter.create_session(request)
            assert_session_id_non_empty(session_id, case_id, transport)
            if expected.get("session_context_available", False):
                context = adapter.get_session_context(session_id)
                assert_session_context_present(context, case_id, transport)
            return

        session_request = {
            "user_id": user_id,
            "name": request.get("name", "conformance-session"),
            "description": request.get("description", "conformance session"),
            "namespace": request.get("namespace", "conformance"),
        }
        session_id = adapter.create_session(session_request)
        assert_session_id_non_empty(session_id, case_id, transport)

        if feature == "session_list":
            list_response = adapter.list_user_sessions(request)
            if expected.get("sessions_array_present", False):
                assert_sessions_array(list_response, case_id, transport)
            if expected.get("contains_created_session", False):
                assert_contains_session(list_response, session_id, case_id, transport)
            return

        if feature == "session_update":
            updated_session = adapter.update_session(session_id, request)
            if expected.get("session_id_matches_created", False):
                assert_session_field_equals(
                    updated_session, "id", session_id, case_id, transport
                )
            if "name_equals" in expected:
                assert_session_field_equals(
                    updated_session,
                    "name",
                    expected["name_equals"],
                    case_id,
                    transport,
                )
            if "description_equals" in expected:
                assert_session_field_equals(
                    updated_session,
                    "description",
                    expected["description_equals"],
                    case_id,
                    transport,
                )
            if "namespace_equals" in expected:
                assert_session_field_equals(
                    updated_session,
                    "namespace",
                    expected["namespace_equals"],
                    case_id,
                    transport,
                )
            return

        if feature == "tool_discovery":
            tool_name = str(request.get("tool_name", ""))
            adapter.register_unary_echo_tool(
                session_id=session_id,
                tool_name=tool_name,
                description=request.get(
                    "tool_description", "conformance tool discovery tool"
                ),
            )

            tools = adapter.list_tools(session_id)
            listed_tool = next(
                (tool for tool in tools if tool.get("name") == tool_name), None
            )

            if expected.get("listed_after_register", False):
                if listed_tool is None:
                    raise AssertionError(
                        f"[{transport}] {case_id}: expected tool '{tool_name}' in listed tools"
                    )

            tool_id = str((listed_tool or {}).get("id", ""))
            if expected.get("tool_id_non_empty", False):
                assert_tool_id_non_empty(tool_id, case_id, transport)
            if listed_tool and expected.get("session_id_matches_created", False):
                assert_tool_field_equals(
                    listed_tool, "session_id", session_id, case_id, transport
                )
            if listed_tool and "name_equals" in expected:
                assert_tool_field_equals(
                    listed_tool, "name", expected["name_equals"], case_id, transport
                )
            if listed_tool and "description_equals" in expected:
                assert_tool_field_equals(
                    listed_tool,
                    "description",
                    expected["description_equals"],
                    case_id,
                    transport,
                )

            if expected.get("lookup_by_id", False):
                tool_by_id = adapter.get_tool_by_id(session_id, tool_id)
                assert_tool_field_equals(tool_by_id, "id", tool_id, case_id, transport)
                if "name_equals" in expected:
                    assert_tool_field_equals(
                        tool_by_id,
                        "name",
                        expected["name_equals"],
                        case_id,
                        transport,
                    )

            if expected.get("lookup_by_name", False):
                tool_by_name = adapter.get_tool_by_name(session_id, tool_name)
                assert_tool_field_equals(
                    tool_by_name, "id", tool_id, case_id, transport
                )
                if "description_equals" in expected:
                    assert_tool_field_equals(
                        tool_by_name,
                        "description",
                        expected["description_equals"],
                        case_id,
                        transport,
                    )

            deleted = adapter.delete_tool(session_id, tool_id)
            if expected.get("delete_success", False):
                assert_success_true(deleted, "tool delete result", case_id, transport)

            tools_after_delete = adapter.list_tools(session_id)
            if expected.get("absent_after_delete", False):
                assert_tool_list_excludes(
                    tools_after_delete, tool_id, case_id, transport
                )
            else:
                assert_tool_list_contains(
                    tools_after_delete, tool_id, case_id, transport
                )
            return

        if feature == "request_create":
            tool_name = request["tool_name"]
            adapter.register_unary_echo_tool(
                session_id=session_id,
                tool_name=tool_name,
                description=request.get("tool_description", "conformance request tool"),
            )
            request_id = adapter.create_request(
                session_id, tool_name, request.get("params", {})
            )
            assert_request_id_non_empty(request_id, case_id, transport)

            request_status = adapter.get_request_status(session_id, request_id)
            if "status_equals" in expected:
                assert_request_status(
                    request_status, expected["status_equals"], case_id, transport
                )
            if "tool_name_equals" in expected:
                assert_request_field_equals(
                    request_status,
                    "toolName",
                    expected["tool_name_equals"],
                    case_id,
                    transport,
                )
            if expected.get("request_id_matches_created", False):
                assert_request_field_equals(
                    request_status, "id", request_id, case_id, transport
                )

            listed_requests = adapter.list_requests(session_id, request)
            if expected.get("listed_request_present", False):
                assert_request_list_contains(
                    listed_requests, request_id, case_id, transport
                )
            return

        if feature == "request_recovery":
            _execute_request_recovery_case(
                adapter, session_id, request, expected, case_id, transport
            )
            return

        if feature == "api_key_lifecycle":
            api_key = adapter.create_api_key(
                session_id,
                request.get("api_key_name", "conformance-key"),
                request.get("api_key_capabilities"),
            )
            if expected.get("api_key_id_non_empty", False):
                assert_api_key_id_non_empty(api_key, case_id, transport)
            if expected.get("api_key_value_non_empty", False):
                assert_api_key_value_non_empty(api_key, case_id, transport)
            if expected.get("key_preview_non_empty", False):
                assert_api_key_preview_non_empty(api_key, case_id, transport)
            if expected.get("session_id_matches_created", False):
                assert_api_key_field_equals(
                    api_key, "session_id", session_id, case_id, transport
                )
            if "name_equals" in expected:
                assert_api_key_field_equals(
                    api_key, "name", expected["name_equals"], case_id, transport
                )
            if isinstance(expected.get("capabilities_equal"), list):
                assert_api_key_capabilities_equal(
                    api_key,
                    [str(value) for value in expected["capabilities_equal"]],
                    case_id,
                    transport,
                )

            listed_api_keys = adapter.list_api_keys(session_id)
            if expected.get("listed_after_create", False):
                assert_api_key_list_contains(
                    listed_api_keys, api_key["id"], case_id, transport
                )
            listed_api_key = next(
                (
                    entry
                    for entry in listed_api_keys
                    if isinstance(entry, dict) and entry.get("id") == api_key.get("id")
                ),
                None,
            )
            if expected.get("listed_key_value_empty", False) and listed_api_key:
                assert_api_key_value_empty(listed_api_key, case_id, transport)
            if expected.get("listed_key_preview_non_empty", False) and listed_api_key:
                assert_api_key_preview_non_empty(listed_api_key, case_id, transport)
            if isinstance(expected.get("listed_capabilities_equal"), list) and listed_api_key:
                assert_api_key_capabilities_equal(
                    listed_api_key,
                    [str(value) for value in expected["listed_capabilities_equal"]],
                    case_id,
                    transport,
                )

            revoke_success = adapter.revoke_api_key(session_id, api_key["id"])
            if expected.get("revoke_success", False):
                assert_success_true(
                    revoke_success, "api key revoke result", case_id, transport
                )

            listed_after_revoke = adapter.list_api_keys(session_id)
            if expected.get("absent_after_revoke", False):
                assert_api_key_list_excludes(
                    listed_after_revoke, api_key["id"], case_id, transport
                )
            return

        if feature == "machine_lifecycle":
            machine = adapter.register_machine(session_id, request)
            machine_id = machine.get("id", "")

            if expected.get("machine_id_non_empty", False):
                assert_machine_id_non_empty(machine, case_id, transport)
            if expected.get("session_id_matches_created", False):
                assert_machine_field_equals(
                    machine, "session_id", session_id, case_id, transport
                )
            if "sdk_version_equals" in expected:
                assert_machine_field_equals(
                    machine,
                    "sdk_version",
                    expected["sdk_version_equals"],
                    case_id,
                    transport,
                )
            if "sdk_language_equals" in expected:
                assert_machine_field_equals(
                    machine,
                    "sdk_language",
                    expected["sdk_language_equals"],
                    case_id,
                    transport,
                )

            listed_machines = adapter.list_machines(session_id)
            if expected.get("listed_after_register", False):
                assert_machine_list_contains(
                    listed_machines, machine_id, case_id, transport
                )

            fetched_machine = adapter.get_machine(session_id, machine_id)
            if expected.get("retrieved_by_id", False):
                assert_machine_field_equals(
                    fetched_machine, "id", machine_id, case_id, transport
                )

            if "inflight_result_equals" in expected:
                tool_name = request["tool_name"]
                adapter.register_unary_echo_tool(
                    session_id=session_id,
                    tool_name=tool_name,
                    description=request.get(
                        "tool_description", "conformance busy drain tool"
                    ),
                )

                invoke_result: List[Any] = []
                invoke_errors: List[BaseException] = []

                def _invoke() -> None:
                    try:
                        invoke_result.append(
                            adapter.invoke(
                                session_id,
                                tool_name,
                                request.get("invoke_params", {}),
                            )
                        )
                    except BaseException as exc:  # pragma: no cover - surfaced below
                        invoke_errors.append(exc)

                invoke_thread = threading.Thread(target=_invoke, daemon=True)
                invoke_thread.start()

                _wait_for_running_request(
                    adapter, session_id, tool_name, case_id, transport
                )

                drained = adapter.drain_machine(session_id, machine_id)
                invoke_thread.join(timeout=float(request.get("invoke_timeout_seconds", 10)))

                if invoke_thread.is_alive():
                    raise AssertionError(
                        f"[{transport}] {case_id}: delayed invoke did not complete before timeout"
                    )
                if invoke_errors:
                    raise AssertionError(
                        f"[{transport}] {case_id}: delayed invoke failed: {invoke_errors[0]}"
                    )
                if not invoke_result:
                    raise AssertionError(
                        f"[{transport}] {case_id}: delayed invoke completed without a result payload"
                    )

                assert_unary_result(
                    invoke_result[0],
                    expected.get("inflight_result_equals", {}),
                    case_id,
                    transport,
                )

                if expected.get("drain_success", False):
                    assert_success_true(
                        drained, "machine drain result", case_id, transport
                    )

                listed_after_drain = adapter.list_machines(session_id)
                if expected.get("absent_after_drain", False):
                    assert_machine_list_excludes(
                        listed_after_drain, machine_id, case_id, transport
                    )
                return

            drained = adapter.drain_machine(session_id, machine_id)
            if expected.get("drain_success", False):
                assert_success_true(drained, "machine drain result", case_id, transport)

            listed_after_drain = adapter.list_machines(session_id)
            if expected.get("absent_after_drain", False):
                assert_machine_list_excludes(
                    listed_after_drain, machine_id, case_id, transport
                )
            return

        if feature == "provider_runtime":
            mode = request.get("mode", "unary")
            tool_name = request["tool_name"]

            if mode == "unary":
                adapter.register_unary_echo_tool(
                    session_id=session_id,
                    tool_name=tool_name,
                    description=request.get(
                        "tool_description", "conformance provider unary tool"
                    ),
                )
                adapter.start_provider_runtime(session_id)
                request_id = adapter.create_request(
                    session_id, tool_name, request.get("params", {})
                )
                if expected.get("request_id_non_empty", False):
                    assert_request_id_non_empty(request_id, case_id, transport)

                request_status = adapter.wait_for_request_completion(
                    session_id,
                    request_id,
                    timeout_seconds=float(request.get("timeout_seconds", 30)),
                )
                if "status_equals" in expected:
                    assert_request_status(
                        request_status, expected["status_equals"], case_id, transport
                    )
                if expected.get("executing_machine_present", False):
                    assert_request_field_non_empty(
                        request_status, "executingMachineId", case_id, transport
                    )
                if "result_equals" in expected:
                    assert_unary_result(
                        request_status.get("result"),
                        expected["result_equals"],
                        case_id,
                        transport,
                    )
                return

            if mode == "fenced":
                # Lease fencing proof: claim manually (no runtime polling),
                # reject forged writes, renew the lease, then submit with the
                # real grant.
                adapter.register_unary_echo_tool(
                    session_id=session_id,
                    tool_name=tool_name,
                    description=request.get(
                        "tool_description", "conformance fenced provider tool"
                    ),
                )
                machine_id = adapter.get_provider_machine_id(session_id)
                request_id = adapter.create_request(
                    session_id, tool_name, request.get("params", {})
                )
                if expected.get("request_id_non_empty", False):
                    assert_request_id_non_empty(request_id, case_id, transport)

                claimed = adapter.claim_request(session_id, request_id, machine_id)
                if claimed.get("errorCode"):
                    raise AssertionError(
                        f"[{case_id}][{transport}] claim failed: "
                        f"{claimed.get('errorCode')} {claimed.get('errorMessage')}"
                    )
                lease_epoch = int(claimed.get("leaseEpoch", 0))
                if lease_epoch <= 0:
                    raise AssertionError(
                        f"[{case_id}][{transport}] claim did not grant a lease epoch"
                    )

                forged_submit = adapter.submit_fenced_result(
                    session_id, request_id, machine_id, lease_epoch + 41, {"forged": True}
                )
                if "forged_submit_error_code" in expected:
                    assert_error_code_equals(
                        forged_submit,
                        expected["forged_submit_error_code"],
                        case_id,
                        transport,
                    )

                forged_renew = adapter.renew_request_lease(
                    session_id, request_id, machine_id, lease_epoch + 41
                )
                if "forged_renew_error_code" in expected:
                    assert_error_code_equals(
                        forged_renew,
                        expected["forged_renew_error_code"],
                        case_id,
                        transport,
                    )

                renewed = adapter.renew_request_lease(
                    session_id, request_id, machine_id, lease_epoch
                )
                if renewed.get("errorCode"):
                    raise AssertionError(
                        f"[{case_id}][{transport}] holder renewal failed: "
                        f"{renewed.get('errorCode')} {renewed.get('errorMessage')}"
                    )
                if expected.get("renewed_lease_expires_non_empty", False):
                    assert_request_field_non_empty(
                        renewed, "leaseExpiresAt", case_id, transport
                    )
                assert_request_field_equals(
                    renewed, "leaseEpoch", lease_epoch, case_id, transport
                )

                holder_submit = adapter.submit_fenced_result(
                    session_id,
                    request_id,
                    machine_id,
                    lease_epoch,
                    expected.get("submit_result", {"echo": "fenced"}),
                )
                if holder_submit.get("errorCode"):
                    raise AssertionError(
                        f"[{case_id}][{transport}] holder submit failed: "
                        f"{holder_submit.get('errorCode')} "
                        f"{holder_submit.get('errorMessage')}"
                    )

                request_status = adapter.get_request_status(session_id, request_id)
                if "status_equals" in expected:
                    assert_request_status(
                        request_status, expected["status_equals"], case_id, transport
                    )
                if "result_equals" in expected:
                    assert_unary_result(
                        request_status.get("result"),
                        expected["result_equals"],
                        case_id,
                        transport,
                    )
                return

            if mode == "stream":
                adapter.register_stream_tool(
                    session_id=session_id,
                    tool_name=tool_name,
                    description=request.get(
                        "tool_description", "conformance provider stream tool"
                    ),
                )
                adapter.start_provider_runtime(session_id)
                request_id = adapter.create_request(
                    session_id, tool_name, request.get("params", {})
                )
                if expected.get("request_id_non_empty", False):
                    assert_request_id_non_empty(request_id, case_id, transport)

                request_status = adapter.wait_for_request_completion(
                    session_id,
                    request_id,
                    timeout_seconds=float(request.get("timeout_seconds", 30)),
                )
                if "status_equals" in expected:
                    assert_request_status(
                        request_status, expected["status_equals"], case_id, transport
                    )
                if expected.get("executing_machine_present", False):
                    assert_request_field_non_empty(
                        request_status, "executingMachineId", case_id, transport
                    )
                assert_stream_chunks(
                    request_status.get("streamResults", []),
                    expected.get("ordered_chunks", []),
                    case_id,
                    transport,
                )
                return

            if mode == "drain":
                adapter.register_unary_echo_tool(
                    session_id=session_id,
                    tool_name=tool_name,
                    description=request.get(
                        "tool_description", "conformance provider drain tool"
                    ),
                )
                adapter.start_provider_runtime(session_id)
                context = adapter.get_session_context(session_id)
                machine_id = getattr(context, "machine_id", "")
                if not machine_id:
                    raise AssertionError(
                        f"[{transport}] {case_id}: provider runtime did not attach a machine"
                    )

                request_id = adapter.create_request(
                    session_id, tool_name, request.get("params", {})
                )
                if expected.get("request_id_non_empty", False):
                    assert_request_id_non_empty(request_id, case_id, transport)

                _wait_for_running_request(
                    adapter, session_id, tool_name, case_id, transport
                )
                drained = adapter.drain_machine(session_id, machine_id)
                request_status = adapter.wait_for_request_completion(
                    session_id,
                    request_id,
                    timeout_seconds=float(request.get("timeout_seconds", 30)),
                )
                if "status_equals" in expected:
                    assert_request_status(
                        request_status, expected["status_equals"], case_id, transport
                    )
                if expected.get("executing_machine_present", False):
                    assert_request_field_non_empty(
                        request_status, "executingMachineId", case_id, transport
                    )
                if "result_equals" in expected:
                    assert_unary_result(
                        request_status.get("result"),
                        expected["result_equals"],
                        case_id,
                        transport,
                    )
                if expected.get("drain_success", False):
                    assert_success_true(drained, "provider drain result", case_id, transport)

                listed_after_drain = adapter.list_machines(session_id)
                if expected.get("absent_after_drain", False):
                    assert_machine_list_excludes(
                        listed_after_drain, machine_id, case_id, transport
                    )
                return

            raise ValueError(f"[{transport}] {case_id}: unsupported provider runtime mode {mode}")

        if feature == "invoke_unary":
            tool_name = request["tool_name"]
            adapter.register_unary_echo_tool(
                session_id=session_id,
                tool_name=tool_name,
                description=request.get("tool_description", "conformance unary tool"),
            )
            result = adapter.invoke(session_id, tool_name, request.get("params", {}))
            assert_unary_result(result, expected.get("result_equals", {}), case_id, transport)
            return

        if feature == "invoke_stream":
            tool_name = request["tool_name"]
            adapter.register_stream_tool(
                session_id=session_id,
                tool_name=tool_name,
                description=request.get("tool_description", "conformance stream tool"),
            )
            chunks, saw_final = adapter.stream(session_id, tool_name, request.get("params", {}))
            assert_stream_chunks(chunks, expected.get("ordered_chunks", []), case_id, transport)
            if expected.get("final_marker", False):
                assert_final_marker(saw_final, case_id, transport)
            return

        if feature == "multi_instance":
            _execute_multi_instance_case(
                case_id, request, expected, transport, user_id
            )
            return

        if feature == "mcp_tasks":
            _execute_mcp_tasks_case(adapter, session_id, request, expected, case_id, transport)
            return

        raise ValueError(f"[{transport}] {case_id}: unsupported feature {feature}")

    finally:
        adapter.close()
