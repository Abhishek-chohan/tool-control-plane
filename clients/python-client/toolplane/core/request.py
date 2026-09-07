"""Request management for Toolplane client."""

import json
import logging
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
    RenewRequestLeaseRequest,
    SubmitRequestResultRequest,
    UpdateRequestRequest,
)

from .connection import ConnectionManager
from .errors import RequestError, api_error_from_rpc_error

logger = logging.getLogger(__name__)

# The server grants a 30s lease TTL by default; renewing at one third of that
# keeps a healthy executor comfortably ahead of the reaper.
LEASE_RENEWAL_INTERVAL_SECONDS = 10.0


class RequestManager:
    """Manages request processing and polling."""

    def __init__(self, connection_manager: ConnectionManager, max_workers: int = 10):
        """Initialize request manager."""
        self.connection_manager = connection_manager
        self.executor = ThreadPoolExecutor(max_workers=max_workers)
        self._running = False
        self._poll_thread: Optional[threading.Thread] = None
        self._poll_interval = 1.0
        # In-flight lease registry: request_id -> lease grant metadata used by
        # the renewal loop and by fenced provider writes.
        self._active_leases: Dict[str, Dict[str, Any]] = {}
        self._leases_lock = threading.Lock()
        self._renewal_running = False
        self._renewal_thread: Optional[threading.Thread] = None
        self._renewal_interval = LEASE_RENEWAL_INTERVAL_SECONDS

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
            "leasedBy": request.leased_by,
            "leaseEpoch": request.lease_epoch,
            "leaseExpiresAt": request.lease_expires_at,
            "timeoutSeconds": request.timeout_seconds,
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

    # ---------------- Lease bookkeeping ----------------

    def register_active_lease(
        self, session_id: str, request_id: str, machine_id: str, lease_epoch: int
    ):
        """Track an in-flight lease so the renewal loop can keep it alive."""
        with self._leases_lock:
            self._active_leases[request_id] = {
                "session_id": session_id,
                "machine_id": machine_id,
                "lease_epoch": lease_epoch,
            }

    def release_active_lease(self, request_id: str):
        """Stop tracking a lease once its execution finished or the lease was lost."""
        with self._leases_lock:
            self._active_leases.pop(request_id, None)

    def start_lease_renewal(self, interval: float = LEASE_RENEWAL_INTERVAL_SECONDS):
        """Start the background lease renewal loop."""
        if self._renewal_running:
            return
        self._renewal_running = True
        self._renewal_interval = interval
        self._renewal_thread = threading.Thread(
            target=self._lease_renewal_loop, daemon=True
        )
        self._renewal_thread.start()

    def stop_lease_renewal(self):
        """Stop the background lease renewal loop."""
        self._renewal_running = False
        if self._renewal_thread:
            self._renewal_thread.join(timeout=1)

    def _lease_renewal_loop(self):
        """Renew every tracked lease that is due."""
        while self._renewal_running:
            time.sleep(self._renewal_interval)
            if not self._renewal_running:
                return
            with self._leases_lock:
                leases = dict(self._active_leases)
            for request_id, lease in leases.items():
                if not self._renewal_running:
                    return
                try:
                    self.renew_request_lease(
                        lease["session_id"],
                        request_id,
                        lease["machine_id"],
                        lease["lease_epoch"],
                    )
                except grpc.RpcError as rpc_error:
                    if rpc_error.code() == grpc.StatusCode.FAILED_PRECONDITION:
                        # The lease was reclaimed or expired: stop renewing and
                        # let the fenced writes surface the loss.
                        logger.warning(
                            "Lease for request %s was lost (%s); stopping renewal.",
                            request_id,
                            rpc_error.code().name,
                        )
                        self.release_active_lease(request_id)
                    else:
                        # Transient transport error: keep the lease tracked and
                        # retry on the next pass.
                        logger.warning(
                            "Lease renewal failed for request %s: %s",
                            request_id,
                            rpc_error,
                        )
                except Exception as e:
                    logger.warning(
                        "Lease renewal failed for request %s: %s", request_id, e
                    )

    def renew_request_lease(
        self, session_id: str, request_id: str, machine_id: str, lease_epoch: int
    ) -> Dict[str, Any]:
        """Renew the execution lease for a claimed/running request.

        Only the current lease holder may renew; the server rejects stale or
        mismatched grants with FAILED_PRECONDITION.
        """
        request = RenewRequestLeaseRequest(
            session_id=session_id,
            request_id=request_id,
            machine_id=machine_id,
            lease_epoch=lease_epoch,
        )

        response = self.connection_manager.requests_stub.RenewRequestLease(
            request, metadata=self.connection_manager.get_metadata()
        )
        return self._normalize_request(response)

    # ---------------- Provider poll/execution ----------------

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

                    # Track the lease grant so the renewal loop keeps it alive.
                    self.register_active_lease(
                        session_id,
                        claimed_req.id,
                        claimed_req.leased_by or machine_id,
                        claimed_req.lease_epoch,
                    )

                    # Execute in thread pool
                    self.executor.submit(
                        self._execute_request, claimed_req, tools, streaming_tools
                    )

                except grpc.RpcError as rpc_error:
                    code = rpc_error.code()
                    if code in (
                        grpc.StatusCode.FAILED_PRECONDITION,
                        grpc.StatusCode.NOT_FOUND,
                    ):
                        # Lost the claim race (or the request vanished between
                        # list and claim): expected contention, not an error.
                        logger.debug(
                            "Claim race lost for request %s in session %s: %s",
                            req.id,
                            session_id,
                            rpc_error,
                        )
                    else:
                        logger.warning(
                            "Failed to claim request %s in session %s: %s",
                            req.id,
                            session_id,
                            rpc_error,
                        )
                except Exception as e:
                    logger.warning(
                        "Failed to claim request %s in session %s: %s",
                        req.id,
                        session_id,
                        e,
                    )

        except grpc.RpcError as rpc_error:
            if rpc_error.code() == grpc.StatusCode.UNAVAILABLE:
                self.connection_manager.mark_unhealthy()
            raise api_error_from_rpc_error(
                rpc_error, context=f"Failed to poll requests for session {session_id}"
            ) from rpc_error
        except Exception as e:
            raise RequestError(f"Failed to poll requests for session {session_id}: {e}")

    def _execute_request(
        self, request, tools: Dict[str, Callable], streaming_tools: set
    ):
        """Execute a claimed request."""
        tool_name = request.tool_name
        request_id = request.id
        session_id = request.session_id
        machine_id = request.leased_by
        lease_epoch = request.lease_epoch

        if tool_name not in tools:
            self._submit_error_result(
                session_id,
                request_id,
                f"Tool '{tool_name}' not found",
                machine_id=machine_id,
                lease_epoch=lease_epoch,
            )
            self.release_active_lease(request_id)
            return

        try:
            # Parse input parameters
            try:
                params = json.loads(request.input)
            except json.JSONDecodeError:
                self._submit_error_result(
                    session_id,
                    request_id,
                    "Invalid JSON input",
                    machine_id=machine_id,
                    lease_epoch=lease_epoch,
                )
                self.release_active_lease(request_id)
                return

            # Execute the tool
            tool_func = tools[tool_name]
            is_streaming = tool_name in streaming_tools

            self._handle_tool_execution(
                session_id,
                request_id,
                tool_func,
                params,
                is_streaming,
                machine_id=machine_id,
                lease_epoch=lease_epoch,
            )

        except Exception as e:
            self._submit_error_result(
                session_id,
                request_id,
                str(e),
                machine_id=machine_id,
                lease_epoch=lease_epoch,
            )
        finally:
            self.release_active_lease(request_id)

    def _handle_tool_execution(
        self,
        session_id: str,
        request_id: str,
        tool_func: Callable,
        params: Dict,
        is_streaming: bool,
        machine_id: str = "",
        lease_epoch: int = 0,
    ):
        """Handle tool execution (streaming or non-streaming)."""
        try:
            # Mark as running
            self._update_request_status(
                session_id,
                request_id,
                "running",
                machine_id=machine_id,
                lease_epoch=lease_epoch,
            )

            if is_streaming:
                self._handle_streaming_execution(
                    session_id,
                    request_id,
                    tool_func,
                    params,
                    machine_id=machine_id,
                    lease_epoch=lease_epoch,
                )
            else:
                self._handle_normal_execution(
                    session_id,
                    request_id,
                    tool_func,
                    params,
                    machine_id=machine_id,
                    lease_epoch=lease_epoch,
                )

        except Exception as e:
            self._submit_error_result(
                session_id,
                request_id,
                str(e),
                machine_id=machine_id,
                lease_epoch=lease_epoch,
            )

    def _handle_streaming_execution(
        self,
        session_id: str,
        request_id: str,
        tool_func: Callable,
        params: Dict,
        machine_id: str = "",
        lease_epoch: int = 0,
    ):
        """Handle streaming tool execution."""
        # Set streaming mode
        self._update_request(
            session_id,
            request_id,
            result_type="streaming",
            machine_id=machine_id,
            lease_epoch=lease_epoch,
        )

        chunks = []
        for chunk in tool_func(**params):
            data = chunk if isinstance(chunk, str) else json.dumps(chunk)

            # Append chunk
            self._append_request_chunk(
                session_id,
                request_id,
                data,
                machine_id=machine_id,
                lease_epoch=lease_epoch,
            )
            chunks.append(data)

        # Submit final result
        self._submit_result(
            session_id,
            request_id,
            json.dumps(chunks),
            "resolution",
            machine_id=machine_id,
            lease_epoch=lease_epoch,
        )

    def _handle_normal_execution(
        self,
        session_id: str,
        request_id: str,
        tool_func: Callable,
        params: Dict,
        machine_id: str = "",
        lease_epoch: int = 0,
    ):
        """Handle normal tool execution."""
        result = tool_func(**params)
        self._submit_result(
            session_id,
            request_id,
            json.dumps(result),
            "resolution",
            machine_id=machine_id,
            lease_epoch=lease_epoch,
        )

    def _update_request_status(
        self,
        session_id: str,
        request_id: str,
        status: str,
        machine_id: str = "",
        lease_epoch: int = 0,
    ):
        """Update request status (fenced provider write)."""
        request = UpdateRequestRequest(
            session_id=session_id,
            request_id=request_id,
            status=status,
            machine_id=machine_id,
            lease_epoch=lease_epoch,
        )

        self.connection_manager.requests_stub.UpdateRequest(
            request, metadata=self.connection_manager.get_metadata()
        )

    def _update_request(
        self,
        session_id: str,
        request_id: str,
        result_type: str,
        machine_id: str = "",
        lease_epoch: int = 0,
    ):
        """Update request with result type (fenced provider write)."""
        request = UpdateRequestRequest(
            session_id=session_id,
            request_id=request_id,
            result_type=result_type,
            machine_id=machine_id,
            lease_epoch=lease_epoch,
        )

        self.connection_manager.requests_stub.UpdateRequest(
            request, metadata=self.connection_manager.get_metadata()
        )

    def _append_request_chunk(
        self,
        session_id: str,
        request_id: str,
        chunk: str,
        machine_id: str = "",
        lease_epoch: int = 0,
    ):
        """Append chunk to request (fenced provider write)."""
        request = AppendRequestChunksRequest(
            session_id=session_id,
            request_id=request_id,
            chunks=[chunk],
            result_type="streaming",
            machine_id=machine_id,
            lease_epoch=lease_epoch,
        )

        self.connection_manager.requests_stub.AppendRequestChunks(
            request, metadata=self.connection_manager.get_metadata()
        )

    def _submit_result(
        self,
        session_id: str,
        request_id: str,
        result: str,
        result_type: str,
        machine_id: str = "",
        lease_epoch: int = 0,
    ):
        """Submit request result (fenced provider write)."""
        request = SubmitRequestResultRequest(
            session_id=session_id,
            request_id=request_id,
            result=result,
            result_type=result_type,
            machine_id=machine_id,
            lease_epoch=lease_epoch,
        )

        self.connection_manager.requests_stub.SubmitRequestResult(
            request, metadata=self.connection_manager.get_metadata()
        )

    def _submit_error_result(
        self,
        session_id: str,
        request_id: str,
        error: str,
        machine_id: str = "",
        lease_epoch: int = 0,
    ):
        """Submit error result.

        A rejection that itself fails fencing (the lease was already lost) is
        logged rather than propagated, so error handling cannot loop.
        """
        try:
            self._submit_result(
                session_id,
                request_id,
                json.dumps({"error": error}),
                "rejection",
                machine_id=machine_id,
                lease_epoch=lease_epoch,
            )
        except grpc.RpcError as rpc_error:
            if rpc_error.code() == grpc.StatusCode.FAILED_PRECONDITION:
                logger.warning(
                    "Could not submit rejection for request %s: lease was lost (%s)",
                    request_id,
                    rpc_error.code().name,
                )
            else:
                raise

    # ---------------- Consumer API ----------------

    def create_request(
        self,
        session_id: str,
        tool_name: str,
        input_data: str,
        timeout_seconds: int = 0,
    ) -> str:
        """Create a new request.

        timeout_seconds optionally overrides the absolute execution timeout;
        zero keeps the server default.
        """
        try:
            self.connection_manager.ensure_connected()

            request = CreateRequestRequest(
                session_id=session_id,
                tool_name=tool_name,
                input=input_data,
                timeout_seconds=timeout_seconds,
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
        self.stop_lease_renewal()
        self.executor.shutdown(wait=False)
