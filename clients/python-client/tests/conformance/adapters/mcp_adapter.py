"""Hybrid MCP conformance adapter.

Management-plane operations (sessions, machines, tools registration, provider
runtime, request inspection) delegate to the HTTP SDK adapter, while the
tool-plane operations under test (discovery, invocation, tasks) go through the
Go MCP gateway facade as MCP 2026-07-28 JSON-RPC. This split lets the suite
exercise the facade exactly as an external MCP client would, without losing
the conformance harness's fixture machinery.
"""
import json
import time
import urllib.error
import urllib.request
from typing import Any, Dict, List, Optional, Tuple

from .http_adapter import HttpConformanceAdapter


PROTOCOL_VERSION = "2026-07-28"
TASKS_EXTENSION_ID = "io.modelcontextprotocol/tasks"
SESSION_META_KEY = "dev.toolplane/session_id"
LAST_SEQ_META_KEY = "dev.toolplane/last_seq"
CHUNKS_META_KEY = "dev.toolplane/chunks"
TASK_ID_DATA_KEY = "toolplaneTaskId"


class McpProtocolError(RuntimeError):
    def __init__(self, code: int, message: str, data: Any = None):
        super().__init__(f"MCP error {code}: {message}")
        self.code = code
        self.message = message
        self.data = data


class McpConformanceAdapter:
    def __init__(
        self,
        mcp_host: str,
        mcp_port: int,
        proxy_host: str,
        proxy_port: int,
        user_id: str,
        api_key: str = "",
    ):
        self._management = HttpConformanceAdapter(
            host=proxy_host, port=proxy_port, user_id=user_id, api_key=api_key
        )
        self._mcp_url = f"http://{mcp_host}:{mcp_port}/mcp"
        self._api_key = api_key
        self._next_request_id = 0

    # ------------------------------------------------------------------
    # Management plane: delegated to the HTTP SDK adapter.
    # ------------------------------------------------------------------
    def connect(self) -> None:
        self._management.connect()

    def close(self) -> None:
        self._management.close()

    def create_session(self, request: Dict[str, Any]) -> str:
        return self._management.create_session(request)

    def get_session_context(self, session_id: str):
        return self._management.get_session_context(session_id)

    def update_session(self, session_id: str, request: Dict[str, Any]) -> Dict[str, Any]:
        return self._management.update_session(session_id, request)

    def list_user_sessions(self, request: Dict[str, Any]) -> Dict[str, Any]:
        return self._management.list_user_sessions(request)

    def start_provider_runtime(self, session_id: str) -> None:
        self._management.start_provider_runtime(session_id)

    def register_unary_echo_tool(self, session_id: str, tool_name: str, description: str):
        self._management.register_unary_echo_tool(session_id, tool_name, description)

    def register_stream_tool(self, session_id: str, tool_name: str, description: str):
        self._management.register_stream_tool(session_id, tool_name, description)

    def get_tool_by_id(self, session_id: str, tool_id: str) -> Dict[str, Any]:
        return self._management.get_tool_by_id(session_id, tool_id)

    def get_tool_by_name(self, session_id: str, tool_name: str) -> Dict[str, Any]:
        return self._management.get_tool_by_name(session_id, tool_name)

    def delete_tool(self, session_id: str, tool_id: str) -> bool:
        return self._management.delete_tool(session_id, tool_id)

    def create_request(self, session_id: str, tool_name: str, params: Dict[str, Any]) -> str:
        return self._management.create_request(session_id, tool_name, params)

    def start_streaming_request(self, session_id: str, tool_name: str, params: Dict[str, Any]) -> str:
        return self._management.start_streaming_request(session_id, tool_name, params)

    def get_request_status(self, session_id: str, request_id: str) -> Dict[str, Any]:
        return self._management.get_request_status(session_id, request_id)

    def get_request_chunks_window(self, session_id: str, request_id: str) -> Dict[str, Any]:
        return self._management.get_request_chunks_window(session_id, request_id)

    def resume_stream(self, request_id: str, last_seq: int) -> Dict[str, Any]:
        return self._management.resume_stream(request_id, last_seq)

    def wait_for_request_completion(self, session_id: str, request_id: str) -> Dict[str, Any]:
        return self._management.wait_for_request_completion(session_id, request_id)

    def list_requests(self, session_id: str, request: Dict[str, Any]) -> List[Dict[str, Any]]:
        return self._management.list_requests(session_id, request)

    def register_machine(self, session_id: str, request: Dict[str, Any]) -> Dict[str, Any]:
        return self._management.register_machine(session_id, request)

    def list_machines(self, session_id: str) -> List[Dict[str, Any]]:
        return self._management.list_machines(session_id)

    def get_machine(self, session_id: str, machine_id: str) -> Dict[str, Any]:
        return self._management.get_machine(session_id, machine_id)

    def unregister_machine(self, session_id: str, machine_id: str) -> bool:
        return self._management.unregister_machine(session_id, machine_id)

    def drain_machine(self, session_id: str, machine_id: str) -> bool:
        return self._management.drain_machine(session_id, machine_id)

    def create_api_key(self, session_id: str, name: str, capabilities: Optional[List[str]] = None) -> Dict[str, Any]:
        return self._management.create_api_key(session_id, name, capabilities)

    def list_api_keys(self, session_id: str) -> List[Dict[str, Any]]:
        return self._management.list_api_keys(session_id)

    def revoke_api_key(self, session_id: str, key_id: str) -> bool:
        return self._management.revoke_api_key(session_id, key_id)

    # ------------------------------------------------------------------
    # MCP JSON-RPC transport.
    # ------------------------------------------------------------------
    def _build_meta(
        self,
        session_id: str,
        with_tasks: bool = False,
        last_seq: Optional[int] = None,
    ) -> Dict[str, Any]:
        capabilities: Dict[str, Any] = {}
        if with_tasks:
            capabilities["extensions"] = {TASKS_EXTENSION_ID: {}}
        meta: Dict[str, Any] = {
            "io.modelcontextprotocol/protocolVersion": PROTOCOL_VERSION,
            "io.modelcontextprotocol/clientCapabilities": capabilities,
        }
        if session_id:
            meta[SESSION_META_KEY] = session_id
        if last_seq is not None:
            meta[LAST_SEQ_META_KEY] = last_seq
        return meta

    def _rpc(self, method: str, params: Dict[str, Any]) -> Dict[str, Any]:
        self._next_request_id += 1
        envelope = {"jsonrpc": "2.0", "id": self._next_request_id, "method": method, "params": params}
        request = urllib.request.Request(
            self._mcp_url,
            data=json.dumps(envelope).encode("utf-8"),
            headers={
                "Content-Type": "application/json",
                "X-API-Key": self._api_key,
                "MCP-Protocol-Version": PROTOCOL_VERSION,
            },
            method="POST",
        )
        try:
            with urllib.request.urlopen(request, timeout=90) as response:
                payload = json.loads(response.read().decode("utf-8"))
        except urllib.error.HTTPError as exc:
            body = exc.read().decode("utf-8", errors="replace")
            try:
                payload = json.loads(body)
            except json.JSONDecodeError:
                raise McpProtocolError(exc.code, f"HTTP {exc.code}: {body}") from exc
        if "error" in payload and payload["error"] is not None:
            error = payload["error"]
            raise McpProtocolError(
                int(error.get("code", 0)), str(error.get("message", "")), error.get("data")
            )
        result = payload.get("result")
        if not isinstance(result, dict):
            raise McpProtocolError(-1, f"expected object result, got {result!r}")
        return result

    def mcp_discover(self) -> Dict[str, Any]:
        return self._rpc("server/discover", {"_meta": self._build_meta("")})

    def mcp_list_tools(self, session_id: str) -> Dict[str, Any]:
        return self._rpc("tools/list", {"_meta": self._build_meta(session_id)})

    def mcp_call_tool(self, session_id: str, tool_name: str, params: Dict[str, Any], with_tasks: bool) -> Dict[str, Any]:
        return self._rpc(
            "tools/call",
            {
                "_meta": self._build_meta(session_id, with_tasks=with_tasks),
                "name": tool_name,
                "arguments": params,
            },
        )

    def mcp_tasks_get(self, session_id: str, task_id: str, last_seq: Optional[int] = None) -> Dict[str, Any]:
        return self._rpc(
            "tasks/get",
            {"_meta": self._build_meta(session_id, last_seq=last_seq), "taskId": task_id},
        )

    def mcp_tasks_cancel(self, session_id: str, task_id: str) -> Dict[str, Any]:
        return self._rpc(
            "tasks/cancel",
            {"_meta": self._build_meta(session_id), "taskId": task_id},
        )

    # ------------------------------------------------------------------
    # Runner-facing tool-plane surface (MCP-backed).
    # ------------------------------------------------------------------
    def list_tools(self, session_id: str) -> List[Dict[str, Any]]:
        result = self.mcp_list_tools(session_id)
        return list(result.get("tools", []))

    def invoke(self, session_id: str, tool_name: str, params: Dict[str, Any]) -> Any:
        self.start_provider_runtime(session_id)
        result = self.mcp_call_tool(session_id, tool_name, params, with_tasks=False)
        if result.get("resultType") == "task":
            raise AssertionError("sync tools/call unexpectedly returned a task handle")
        if result.get("isError"):
            text = _first_text(result)
            raise AssertionError(f"MCP tool call reported isError: {text}")
        if "structuredContent" in result:
            return result["structuredContent"]
        return _first_text(result)

    def stream(
        self, session_id: str, tool_name: str, params: Dict[str, Any]
    ) -> Tuple[List[Any], bool]:
        self.start_provider_runtime(session_id)
        result = self.mcp_call_tool(session_id, tool_name, params, with_tasks=False)
        if result.get("resultType") == "task":
            raise AssertionError("sync tools/call unexpectedly returned a task handle")
        meta = result.get("_meta") or {}
        chunks_meta = meta.get(CHUNKS_META_KEY) or {}
        chunks = list(chunks_meta.get("chunks") or [])
        saw_final = not result.get("isError", False)
        return chunks, saw_final

    # ------------------------------------------------------------------
    # MCP Tasks conformance helpers.
    # ------------------------------------------------------------------
    def register_tasks_tool(self, session_id: str, tool_name: str, description: str) -> None:
        """Register a slow streaming tool whose execution stays in flight long
        enough to observe the working state and chunk cursor."""
        context = self._management._ensure_context_machine(session_id)

        def _tasks_tool(message: str = "chunk", count: int = 4, delay_ms: int = 0, **_: Any):
            total = max(int(delay_ms or 0), 0)
            steps = max(int(count or 1), 1)
            per_step = total / 1000.0 / steps
            for index in range(steps):
                if per_step:
                    time.sleep(per_step)
                yield f"{message}-{index + 1}"

        context.register_tool(
            name=tool_name,
            func=_tasks_tool,
            description=description,
            stream=True,
            tags=["conformance", "mcp", "tasks"],
        )

    def tasks_poll_until_terminal(
        self, session_id: str, task_id: str, timeout_seconds: float = 30.0
    ) -> Dict[str, Any]:
        deadline = time.time() + timeout_seconds
        last: Dict[str, Any] = {}
        while time.time() < deadline:
            last = self.mcp_tasks_get(session_id, task_id)
            status = last.get("status")
            if status in {"completed", "failed", "cancelled"}:
                return last
            time.sleep(0.2)
        raise AssertionError(
            f"task {task_id} did not reach a terminal status within {timeout_seconds}s: {last}"
        )


def _first_text(result: Dict[str, Any]) -> Any:
    content = result.get("content") or []
    for block in content:
        if isinstance(block, dict) and block.get("type") == "text":
            text = block.get("text", "")
            try:
                return json.loads(text)
            except (json.JSONDecodeError, TypeError):
                return text
    return ""
