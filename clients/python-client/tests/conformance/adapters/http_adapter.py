import time
import json
import re
from typing import Any, Dict, List, Tuple

from toolplane import ToolplaneHTTP


def _parse_maybe_json(value: Any) -> Any:
    if not isinstance(value, str):
        return value

    try:
        return json.loads(value)
    except Exception:
        return value


def _normalize_gateway_error_code(code: Any) -> str:
    if isinstance(code, int):
        if code == 11:
            return "out_of_range"
        return str(code)

    if isinstance(code, dict):
        nested = code.get("error")
        if nested is not None:
            nested_code = _normalize_gateway_error_code(nested)
            if nested_code:
                return nested_code

        return _normalize_gateway_error_code(code.get("code"))

    if isinstance(code, str):
        normalized = code.strip().lower().replace("-", "_").replace(" ", "_")
        if normalized in {"out_of_range", "outofrange"}:
            return "out_of_range"
        lowered = code.lower()
        if (
            "out_of_range" in lowered
            or "out of range" in lowered
            or "outofrange" in lowered
        ):
            return "out_of_range"
        return normalized

    return ""


class HttpConformanceAdapter:
    def __init__(self, host: str, port: int, user_id: str, api_key: str = ""):
        self.client = ToolplaneHTTP(
            server_host=host,
            server_port=port,
            user_id=user_id,
            api_key=api_key or None,
        )
        self._request_sessions: Dict[str, str] = {}
        self._provider_runtime = self.client.provider_runtime()

    def connect(self) -> None:
        self.client.connect()

    def close(self) -> None:
        self._provider_runtime.stop()
        self._request_sessions.clear()
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
        request_id = self.client.create_request(session_id, tool_name, json.dumps(params))
        self._request_sessions[request_id] = session_id
        return request_id

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
        response = self.client.connection_manager.get_request_chunks(session_id, request_id)
        return {
            "chunks": [_parse_maybe_json(chunk) for chunk in response.get("chunks", [])],
            "startSeq": response.get("startSeq", response.get("start_seq", 0)),
            "nextSeq": response.get("nextSeq", response.get("next_seq", 0)),
        }

    def resume_stream(self, request_id: str, last_seq: int) -> Dict[str, Any]:
        response = None
        try:
            self.client.connection_manager.ensure_connected()
            response = self.client.connection_manager.stream_post(
                "api/ResumeStream",
                {"requestId": request_id, "lastSeq": last_seq},
            )

            chunks: List[Any] = []
            final_seq = 0
            for line in response.iter_lines(decode_unicode=True):
                if not line:
                    continue
                try:
                    chunk = json.loads(line)
                except ValueError:
                    continue

                if isinstance(chunk, dict) and isinstance(chunk.get("result"), dict):
                    chunk = chunk["result"]

                final_seq = int(chunk.get("seq", 0) or 0)
                if chunk.get("isFinal"):
                    error_message = chunk.get("error", "")
                    return {
                        "chunks": chunks,
                        "sawFinal": True,
                        "finalSeq": final_seq,
                        "errorCode": _normalize_gateway_error_code(error_message),
                        "errorMessage": error_message,
                    }

                value = _parse_maybe_json(chunk.get("chunk"))
                if value not in (None, ""):
                    chunks.append(value)

            fallback = self._resume_from_chunk_window(request_id, last_seq)
            if fallback is not None:
                return fallback

            raise RuntimeError("HTTP resume stream ended before final marker")
        except Exception as exc:
            error_code = self._normalize_resume_error_code(exc)
            if error_code:
                return {
                    "chunks": [],
                    "sawFinal": False,
                    "finalSeq": 0,
                    "errorCode": error_code,
                    "errorMessage": str(exc),
                }
            raise
        finally:
            if response is not None:
                response.close()

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

        payload = {
            "sessionId": session_id,
            "machineId": str(request.get("machine_id", "")),
            "sdkVersion": str(request.get("sdk_version", "1.0.0-conformance")),
            "sdkLanguage": str(request.get("sdk_language", "conformance")),
            "tools": [],
        }

        response = self.client.connection_manager.register_machine(payload)
        machine_id = response.get("id", payload["machineId"])

        context.machine_id = machine_id
        with self.client.machine_manager.machines_lock:
            self.client.machine_manager.machines[session_id] = machine_id
            self.client.machine_manager._last_heartbeat[session_id] = time.time()

        return {
            "id": machine_id,
            "session_id": response.get("sessionId", response.get("session_id", session_id)),
            "sdk_version": response.get("sdkVersion", response.get("sdk_version", payload["sdkVersion"])),
            "sdk_language": response.get("sdkLanguage", response.get("sdk_language", payload["sdkLanguage"])),
            "ip": response.get("ip", ""),
            "created_at": response.get("createdAt", response.get("created_at", "")),
            "last_ping_at": response.get("lastPingAt", response.get("last_ping_at", "")),
        }

    def list_machines(self, session_id: str) -> List[Dict[str, Any]]:
        return self.client.list_machines(session_id)

    def get_machine(self, session_id: str, machine_id: str) -> Dict[str, Any]:
        return self.client.get_machine(session_id, machine_id)

    def unregister_machine(self, session_id: str, machine_id: str) -> bool:
        return self.client.unregister_machine(session_id, machine_id)

    def drain_machine(self, session_id: str, machine_id: str) -> bool:
        return self.client.drain_machine(session_id, machine_id)

    def create_api_key(
        self,
        session_id: str,
        name: str,
        capabilities: List[str] | None = None,
    ) -> Dict[str, Any]:
        return self.client.create_api_key(session_id, name, capabilities)

    def list_api_keys(self, session_id: str) -> List[Dict[str, Any]]:
        return self.client.list_api_keys(session_id)

    def revoke_api_key(self, session_id: str, key_id: str) -> bool:
        return self.client.revoke_api_key(session_id, key_id)

    def invoke(self, session_id: str, tool_name: str, params: Dict[str, Any]) -> Any:
        self.start_provider_runtime(session_id)
        return self.client.invoke(tool_name=tool_name, session_id=session_id, **params)

    def stream(self, session_id: str, tool_name: str, params: Dict[str, Any]) -> Tuple[List[Any], bool]:
        self.start_provider_runtime(session_id)
        try:
            request_id = self.client.ainvoke(
                tool_name=tool_name,
                session_id=session_id,
                **params,
            )

            collected: List[Any] = []
            last_count = 0
            deadline = time.time() + 60

            while time.time() < deadline:
                status = self.client.get_request_status(request_id, session_id)

                stream_results = status.get("streamResults", [])
                if isinstance(stream_results, list) and len(stream_results) > last_count:
                    for value in stream_results[last_count:]:
                        if value not in (None, ""):
                            collected.append(value)
                    last_count = len(stream_results)

                if status.get("status") == "done":
                    final_value = status.get("result")
                    if isinstance(final_value, str):
                        try:
                            parsed = json.loads(final_value)
                            if isinstance(parsed, list) and len(parsed) > len(collected):
                                collected = parsed
                        except Exception:
                            pass
                    return collected, True

                if status.get("status") == "failure":
                    raise RuntimeError(
                        f"HTTP stream request failed: {status.get('error', 'unknown error')}"
                    )

                time.sleep(0.1)

            raise RuntimeError("HTTP stream request timed out")

        except Exception:
            raise

    def _resume_from_chunk_window(
        self, request_id: str, last_seq: int
    ) -> Dict[str, Any] | None:
        session_id = self._request_sessions.get(request_id)
        if not session_id:
            return None

        chunk_window = self.get_request_chunks_window(session_id, request_id)
        chunks = chunk_window.get("chunks")
        if not isinstance(chunks, list):
            return None

        start_seq = int(chunk_window.get("startSeq", 0) or 0)
        next_seq = int(chunk_window.get("nextSeq", 0) or 0)
        start_index = max(last_seq - start_seq + 1, 0)
        replay_chunks = chunks[start_index:]
        return {
            "chunks": replay_chunks,
            "sawFinal": True,
            "finalSeq": next_seq,
            "errorMessage": "",
        }

    def _normalize_resume_error_code(self, exc: Exception) -> str:
        message = str(exc)
        match = re.search(r"(\{.*\})", message)
        if match:
            try:
                payload = json.loads(match.group(1))
                return _normalize_gateway_error_code(payload)
            except Exception:
                pass

        lowered = message.lower()
        if "out_of_range" in lowered or "out of range" in lowered:
            return "out_of_range"

        return ""
