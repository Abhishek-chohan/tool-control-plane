"""Request management for Toolplane client."""

import json
import threading
import time
from concurrent.futures import ThreadPoolExecutor
from typing import Any, Callable, Dict, List, Optional

import grpc

from toolplane.proto.service_pb2 import (
    AppendRequestChunksRequest,
    CancelRequestRequest,
    ClaimRequestRequest,
    CreateRequestRequest,
    GetRequestRequest,
    ListRequestsRequest,
    SubmitRequestResultRequest,
    UpdateRequestRequest,
)

from .connection import ConnectionManager
from .errors import RequestError


class RequestManager:
    """Manages request processing and polling."""

    def __init__(self, connection_manager: ConnectionManager, max_workers: int = 10):
        """Initialize request manager."""
        self.connection_manager = connection_manager
        self.executor = ThreadPoolExecutor(max_workers=max_workers)
        self._running = False
        self._poll_thread: Optional[threading.Thread] = None
        self._poll_interval = 1.0

    def _normalize_request(self, request: Any) -> Dict[str, Any]:
        normalized = {
            "id": request.id,
            "sessionId": request.session_id,
            "toolName": request.tool_name,
            "status": request.status,
            "input": request.input,
            "createdAt": request.created_at,
            "updatedAt": request.updated_at,
            "executingMachineId": request.executing_machine_id,
        }

        if request.result:
            try:
                normalized["result"] = json.loads(request.result)
            except Exception:
                normalized["result"] = request.result

        if request.result_type:
            normalized["resultType"] = request.result_type

        if request.error:
            normalized["error"] = request.error

        if request.stream_results:
            normalized["streamResults"] = list(request.stream_results)

        return normalized

    def start_polling(self, poll_interval: float = 1.0):
        """Start request polling."""
        if self._running:
            return

        self._running = True
        self._poll_interval = poll_interval
        self._poll_thread = threading.Thread(target=self._poll_loop, daemon=True)
        self._poll_thread.start()

    def stop_polling(self):
        """Stop request polling."""
        self._running = False
        if self._poll_thread:
            self._poll_thread.join(timeout=1)

    def _poll_loop(self):
        """Main polling loop."""
        while self._running:
            try:
                # This will be called by the main client with session info
                time.sleep(self._poll_interval)
            except Exception:
                # Ignore polling errors
                pass

    def poll_session_requests(
        self,
        session_id: str,
        machine_id: str,
        tools: Dict[str, Callable],
        streaming_tools: set,
        limit: int = 5,
    ):
        """Poll for requests in a specific session."""
        try:
            self.connection_manager.ensure_connected()

            # Get pending requests
            request = ListRequestsRequest(
                session_id=session_id, status="pending", limit=limit
            )

            response = self.connection_manager.requests_stub.ListRequests(
                request, metadata=self.connection_manager.get_metadata()
            )

            # Process each request
            for req in response.requests:
                try:
                    # Claim the request
                    claim_request = ClaimRequestRequest(
                        session_id=session_id,
                        request_id=req.id,
                        machine_id=machine_id,
                    )

                    claimed_req = self.connection_manager.requests_stub.ClaimRequest(
                        claim_request, metadata=self.connection_manager.get_metadata()
                    )

                    # Execute in thread pool
                    self.executor.submit(
                        self._execute_request, claimed_req, tools, streaming_tools
                    )

                except Exception:
                    # Ignore claim errors (request might be claimed by another machine)
                    pass

        except grpc.RpcError as rpc_error:
            if rpc_error.code() in (
                grpc.StatusCode.UNAVAILABLE,
                grpc.StatusCode.DEADLINE_EXCEEDED,
                grpc.StatusCode.INTERNAL,
                grpc.StatusCode.UNAUTHENTICATED,
            ):
                self.connection_manager.mark_unhealthy()
            raise RequestError(
                f"Failed to poll requests for session {session_id}: {rpc_error}"
            )
        except Exception as e:
            raise RequestError(f"Failed to poll requests for session {session_id}: {e}")

    def _execute_request(
        self, request, tools: Dict[str, Callable], streaming_tools: set
    ):
        """Execute a claimed request."""
        tool_name = request.tool_name
        request_id = request.id
        session_id = request.session_id

        if tool_name not in tools:
            self._submit_error_result(
                session_id, request_id, f"Tool '{tool_name}' not found"
            )
            return

        try:
            # Parse input parameters
            try:
                params = json.loads(request.input)
            except json.JSONDecodeError:
                self._submit_error_result(session_id, request_id, "Invalid JSON input")
                return

            # Execute the tool
            tool_func = tools[tool_name]
            is_streaming = tool_name in streaming_tools

            self._handle_tool_execution(
                session_id, request_id, tool_func, params, is_streaming
            )

        except Exception as e:
            self._submit_error_result(session_id, request_id, str(e))

    def _handle_tool_execution(
        self,
        session_id: str,
        request_id: str,
        tool_func: Callable,
        params: Dict,
        is_streaming: bool,
    ):
        """Handle tool execution (streaming or non-streaming)."""
        try:
            # Mark as running
            self._update_request_status(session_id, request_id, "running")

            if is_streaming:
                self._handle_streaming_execution(
                    session_id, request_id, tool_func, params
                )
            else:
                self._handle_normal_execution(session_id, request_id, tool_func, params)

        except Exception as e:
            self._submit_error_result(session_id, request_id, str(e))

    def _handle_streaming_execution(
        self, session_id: str, request_id: str, tool_func: Callable, params: Dict
    ):
        """Handle streaming tool execution."""
        # Set streaming mode
        self._update_request(session_id, request_id, result_type="streaming")

        chunks = []
        for chunk in tool_func(**params):
            data = chunk if isinstance(chunk, str) else json.dumps(chunk)

            # Append chunk
            self._append_request_chunk(session_id, request_id, data)
            chunks.append(data)

        # Submit final result
        self._submit_result(session_id, request_id, json.dumps(chunks), "resolution")

    def _handle_normal_execution(
        self, session_id: str, request_id: str, tool_func: Callable, params: Dict
    ):
        """Handle normal tool execution."""
        result = tool_func(**params)
        self._submit_result(session_id, request_id, json.dumps(result), "resolution")

    def _update_request_status(self, session_id: str, request_id: str, status: str):
        """Update request status."""
        request = UpdateRequestRequest(
            session_id=session_id, request_id=request_id, status=status
        )

        self.connection_manager.requests_stub.UpdateRequest(
            request, metadata=self.connection_manager.get_metadata()
        )

    def _update_request(self, session_id: str, request_id: str, result_type: str):
        """Update request with result type."""
        request = UpdateRequestRequest(
            session_id=session_id, request_id=request_id, result_type=result_type
        )

        self.connection_manager.requests_stub.UpdateRequest(
            request, metadata=self.connection_manager.get_metadata()
        )

    def _append_request_chunk(self, session_id: str, request_id: str, chunk: str):
        """Append chunk to request."""
        request = AppendRequestChunksRequest(
            session_id=session_id,
            request_id=request_id,
            chunks=[chunk],
            result_type="streaming",
        )

        self.connection_manager.requests_stub.AppendRequestChunks(
            request, metadata=self.connection_manager.get_metadata()
        )

    def _submit_result(
        self, session_id: str, request_id: str, result: str, result_type: str
    ):
        """Submit request result."""
        request = SubmitRequestResultRequest(
            session_id=session_id,
            request_id=request_id,
            result=result,
            result_type=result_type,
        )

        self.connection_manager.requests_stub.SubmitRequestResult(
            request, metadata=self.connection_manager.get_metadata()
        )

    def _submit_error_result(self, session_id: str, request_id: str, error: str):
        """Submit error result."""
        self._submit_result(
            session_id, request_id, json.dumps({"error": error}), "rejection"
        )

    def create_request(self, session_id: str, tool_name: str, input_data: str) -> str:
        """Create a new request."""
        try:
            self.connection_manager.ensure_connected()

            request = CreateRequestRequest(
                session_id=session_id,
                tool_name=tool_name,
                input=input_data,
            )

            response = self.connection_manager.requests_stub.CreateRequest(
                request, metadata=self.connection_manager.get_metadata()
            )
            return response.id

        except Exception as e:
            raise RequestError(f"Failed to create request: {e}")

    def list_requests(
        self,
        session_id: str,
        status: str = "",
        tool_name: str = "",
        limit: int = 10,
        offset: int = 0,
    ) -> List[Dict[str, Any]]:
        """List requests in a session."""
        try:
            self.connection_manager.ensure_connected()

            request = ListRequestsRequest(
                session_id=session_id,
                status=status,
                tool_name=tool_name,
                limit=limit,
                offset=offset,
            )

            response = self.connection_manager.requests_stub.ListRequests(
                request, metadata=self.connection_manager.get_metadata()
            )
            return [self._normalize_request(entry) for entry in response.requests]

        except Exception as e:
            raise RequestError(f"Failed to list requests: {e}")

    def get_request_status(self, session_id: str, request_id: str) -> Dict[str, Any]:
        """Get request status."""
        try:
            self.connection_manager.ensure_connected()

            request = GetRequestRequest(session_id=session_id, request_id=request_id)

            response = self.connection_manager.requests_stub.GetRequest(
                request, metadata=self.connection_manager.get_metadata()
            )

            return self._normalize_request(response)

        except Exception as e:
            raise RequestError(f"Failed to get request status: {e}")

    def cancel_request(self, session_id: str, request_id: str) -> bool:
        """Cancel a request."""
        try:
            self.connection_manager.ensure_connected()

            request = CancelRequestRequest(
                session_id=session_id,
                request_id=request_id,
            )

            response = self.connection_manager.requests_stub.CancelRequest(
                request, metadata=self.connection_manager.get_metadata()
            )

            return response.success

        except Exception as e:
            raise RequestError(f"Failed to cancel request: {e}")

    def shutdown(self):
        """Shutdown request manager."""
        self.stop_polling()
        self.executor.shutdown(wait=False)
