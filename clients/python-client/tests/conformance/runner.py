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
from .assertions import (
    assert_api_key_field_equals,
    assert_api_key_id_non_empty,
    assert_api_key_list_contains,
    assert_api_key_list_excludes,
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
    raise ValueError(f"Unsupported transport: {transport}")


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
                session_id, request.get("api_key_name", "conformance-key")
            )
            if expected.get("api_key_id_non_empty", False):
                assert_api_key_id_non_empty(api_key, case_id, transport)
            if expected.get("api_key_value_non_empty", False):
                assert_api_key_value_non_empty(api_key, case_id, transport)
            if expected.get("session_id_matches_created", False):
                assert_api_key_field_equals(
                    api_key, "session_id", session_id, case_id, transport
                )
            if "name_equals" in expected:
                assert_api_key_field_equals(
                    api_key, "name", expected["name_equals"], case_id, transport
                )

            listed_api_keys = adapter.list_api_keys(session_id)
            if expected.get("listed_after_create", False):
                assert_api_key_list_contains(
                    listed_api_keys, api_key["id"], case_id, transport
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

        raise ValueError(f"[{transport}] {case_id}: unsupported feature {feature}")

    finally:
        adapter.close()
