import time
import json
from typing import Any, Dict, List, Tuple

import grpc

from toolplane import Toolplane
from toolplane.proto.service_pb2 import (
    GetRequestChunksRequest,
    RegisterMachineRequest,
    ResumeStreamRequest,
)


def _parse_maybe_json(value: Any) -> Any:
    if not isinstance(value, str):
        return value

    try:
        return json.loads(value)
    except Exception:
        return value


def _normalize_grpc_error_code(exc: grpc.RpcError) -> str:
    status_code = exc.code()
    if status_code is None:
        return ""

    if hasattr(status_code, "name"):
        return status_code.name.lower()

    return str(status_code).split(".")[-1].lower()


class GrpcConformanceAdapter:
    def __init__(self, host: str, port: int, user_id: str, api_key: str = ""):
        self.client = Toolplane(
            server_host=host,
            server_port=port,
            user_id=user_id,
            api_key=api_key or None,
        )
        self._provider_runtime = self.client.provider_runtime()

    def connect(self) -> None:
        self.client.connect()

    def close(self) -> None:
        self._provider_runtime.stop()
        self.client.disconnect()

    def create_session(self, request: Dict[str, Any]) -> str:
        context = self.client.create_session(
            user_id=request.get("user_id"),
            name=request.get("name"),
            description=request.get("description"),
            namespace=request.get("namespace"),
            register_machine=False,
        )
        return context.session_id

    def _get_session_context(self, session_id: str):
        context = self.client.get_session(session_id)
        if not context:
            raise RuntimeError(f"Session context not found for {session_id}")
        return context

    def _ensure_context_machine(self, session_id: str):
        return self._provider_runtime.attach_session(session_id, register_machine=True)

    def start_provider_runtime(self, session_id: str) -> None:
        self._provider_runtime.start_in_background([session_id])

    def get_session_context(self, session_id: str):
        return self.client.get_session(session_id)

    def update_session(self, session_id: str, request: Dict[str, Any]) -> Dict[str, Any]:
        return self.client.update_session(
            session_id=session_id,
            name=request.get("updated_name"),
            description=request.get("updated_description"),
            namespace=request.get("updated_namespace"),
        )

    def list_user_sessions(self, request: Dict[str, Any]) -> Dict[str, Any]:
        return self.client.list_user_sessions(
            user_id=request["user_id"],
            page_size=request.get("page_size", 10),
            page_token=request.get("page_token", 0),
            filter=request.get("filter", ""),
        )

    def register_unary_echo_tool(self, session_id: str, tool_name: str, description: str):
        context = self._ensure_context_machine(session_id)

        def _echo_tool(message: str = "", **_: Any):
            delay_ms = 0
            try:
                delay_ms = max(int(_.get("delay_ms", 0)), 0)
            except (TypeError, ValueError):
                delay_ms = 0
            if delay_ms:
                time.sleep(delay_ms / 1000)
            return {"echo": message}

        context.register_tool(
            name=tool_name,
            func=_echo_tool,
            description=description,
            stream=False,
            tags=["conformance"],
        )

    def register_stream_tool(self, session_id: str, tool_name: str, description: str):
        context = self._ensure_context_machine(session_id)

        def _stream_tool(prefix: str = "chunk", count: int = 5, **_: Any):
            for index in range(int(count)):
                yield f"{prefix}-{index + 1}"

        context.register_tool(
            name=tool_name,
            func=_stream_tool,
            description=description,
            stream=True,
            tags=["conformance", "stream"],
        )

    def list_tools(self, session_id: str) -> List[Dict[str, Any]]:
        return self.client.list_tools(session_id)

    def get_tool_by_id(self, session_id: str, tool_id: str) -> Dict[str, Any]:
        return self.client.get_tool_by_id(session_id, tool_id)

    def get_tool_by_name(self, session_id: str, tool_name: str) -> Dict[str, Any]:
        return self.client.get_tool_by_name(session_id, tool_name)

    def delete_tool(self, session_id: str, tool_id: str) -> bool:
        return self.client.delete_tool(session_id, tool_id)

    def create_request(self, session_id: str, tool_name: str, params: Dict[str, Any]) -> str:
        return self.client.create_request(session_id, tool_name, json.dumps(params))

    def start_streaming_request(
        self, session_id: str, tool_name: str, params: Dict[str, Any]
    ) -> str:
        self.start_provider_runtime(session_id)
        request_id = self.create_request(session_id, tool_name, params)
        return request_id

    def get_request_status(self, session_id: str, request_id: str) -> Dict[str, Any]:
        return self.client.get_request_status(request_id, session_id)

    def get_request_chunks_window(
        self, session_id: str, request_id: str
    ) -> Dict[str, Any]:
        request = GetRequestChunksRequest(session_id=session_id, request_id=request_id)
        response = self.client.connection_manager.requests_stub.GetRequestChunks(
            request, metadata=self.client.connection_manager.get_metadata()
        )
        return {
            "chunks": [_parse_maybe_json(chunk) for chunk in response.chunks],
            "startSeq": response.start_seq,
            "nextSeq": response.next_seq,
        }

    def resume_stream(self, request_id: str, last_seq: int) -> Dict[str, Any]:
        request = ResumeStreamRequest(request_id=request_id, last_seq=last_seq)
        chunks: List[Any] = []
        final_seq = 0

        try:
            stream = self.client.connection_manager.tool_stub.ResumeStream(
                request, metadata=self.client.connection_manager.get_metadata()
            )
            for chunk in stream:
                final_seq = chunk.seq
                if chunk.is_final:
                    return {
                        "chunks": chunks,
                        "sawFinal": True,
                        "finalSeq": final_seq,
                        "errorMessage": chunk.error,
                    }

                value = _parse_maybe_json(chunk.chunk)
                if value not in (None, ""):
                    chunks.append(value)

            raise RuntimeError("gRPC resume stream ended before final marker")
        except grpc.RpcError as exc:
            return {
                "chunks": [],
                "sawFinal": False,
                "finalSeq": final_seq,
                "errorCode": _normalize_grpc_error_code(exc),
                "errorMessage": exc.details(),
            }

    def wait_for_request_completion(
        self, session_id: str, request_id: str, timeout_seconds: float = 60.0
    ) -> Dict[str, Any]:
        deadline = time.time() + timeout_seconds
        while time.time() < deadline:
            status = self.get_request_status(session_id, request_id)
            if status.get("status") == "done":
                return status
            if status.get("status") == "failure":
                raise RuntimeError(
                    f"Request {request_id} failed: {status.get('error', 'unknown error')}"
                )
            time.sleep(0.1)

        raise RuntimeError(f"Timed out waiting for request {request_id}")

    def list_requests(self, session_id: str, request: Dict[str, Any]) -> List[Dict[str, Any]]:
        return self.client.list_requests(
            session_id=session_id,
            status=request.get("list_status", ""),
            tool_name=request.get("tool_name_filter", ""),
            limit=request.get("limit", 10),
            offset=request.get("offset", 0),
        )

    def register_machine(self, session_id: str, request: Dict[str, Any]) -> Dict[str, Any]:
        context = self._get_session_context(session_id)

        machine_request = RegisterMachineRequest(
            session_id=session_id,
            machine_id=str(request.get("machine_id", "")),
            sdk_version=str(request.get("sdk_version", "1.0.0-conformance")),
            sdk_language=str(request.get("sdk_language", "conformance")),
        )

        response = self.client.connection_manager.machine_stub.RegisterMachine(
            machine_request,
            metadata=self.client.connection_manager.get_metadata(),
        )

        context.machine_id = response.id
        with self.client.machine_manager.machines_lock:
            self.client.machine_manager.machines[session_id] = response.id
            self.client.machine_manager._last_heartbeat[session_id] = time.time()

        return {
            "id": response.id,
            "session_id": response.session_id,
            "sdk_version": response.sdk_version,
            "sdk_language": response.sdk_language,
            "ip": response.ip,
            "created_at": response.created_at,
            "last_ping_at": getattr(response, "last_ping_at", ""),
        }

    def list_machines(self, session_id: str) -> List[Dict[str, Any]]:
        return self.client.list_machines(session_id)

    def get_machine(self, session_id: str, machine_id: str) -> Dict[str, Any]:
        return self.client.get_machine(session_id, machine_id)

    def unregister_machine(self, session_id: str, machine_id: str) -> bool:
        return self.client.unregister_machine(session_id, machine_id)

    def drain_machine(self, session_id: str, machine_id: str) -> bool:
        return self.client.drain_machine(session_id, machine_id)

    def create_api_key(self, session_id: str, name: str) -> Dict[str, Any]:
        return self.client.create_api_key(session_id, name)

    def list_api_keys(self, session_id: str) -> List[Dict[str, Any]]:
        return self.client.list_api_keys(session_id)

    def revoke_api_key(self, session_id: str, key_id: str) -> bool:
        return self.client.revoke_api_key(session_id, key_id)

    def invoke(self, session_id: str, tool_name: str, params: Dict[str, Any]) -> Any:
        self.start_provider_runtime(session_id)
        result = self.client.invoke(tool_name=tool_name, session_id=session_id, **params)
        if isinstance(result, dict) and "result" in result:
            return result["result"]
        return result

    def stream(self, session_id: str, tool_name: str, params: Dict[str, Any]) -> Tuple[List[Any], bool]:
        callback_chunks: List[Any] = []
        saw_final = False
        final_payload: Any = None

        def _callback(chunk: Any, is_final: bool):
            nonlocal saw_final, final_payload
            if not is_final and chunk not in (None, ""):
                callback_chunks.append(chunk)
            if is_final:
                saw_final = True
                final_payload = chunk

        self.start_provider_runtime(session_id)
        self.client.stream(
            tool_name=tool_name,
            callback=_callback,
            session_id=session_id,
            **params,
        )
        normalized_chunks = list(callback_chunks)
        if isinstance(final_payload, str):
            try:
                parsed = json.loads(final_payload)
                if isinstance(parsed, list) and len(parsed) > len(normalized_chunks):
                    normalized_chunks = parsed
            except Exception:
                pass

        return normalized_chunks, saw_final
