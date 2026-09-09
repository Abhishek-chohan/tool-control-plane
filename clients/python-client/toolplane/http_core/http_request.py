"""HTTP request management for Toolplane client."""

import json
import logging
import threading
import time
from concurrent.futures import ThreadPoolExecutor
from typing import Any, Callable, Dict, List, Optional

from ..common.constants import (
    DEFAULT_MAX_WORKERS,
    DEFAULT_POLL_INTERVAL,
)
from ..core.errors import (
    RequestError,
    ToolplaneFailedPreconditionError,
    api_error_from_http_response,
)
from .http_connection import HTTPConnectionManager

logger = logging.getLogger(__name__)

# Mirrors toolplane.core.request: renew at one third of the server's default
# 30s lease TTL so a healthy executor stays ahead of the reaper.
LEASE_RENEWAL_INTERVAL_SECONDS = 10.0


class HTTPRequestManager:
    """Manages request processing and polling for HTTP client."""

    def __init__(
        self,
        connection_manager: HTTPConnectionManager,
        max_workers: int = DEFAULT_MAX_WORKERS,
    ):
        """Initialize HTTP request manager."""
        self.connection_manager = connection_manager
        self.executor = ThreadPoolExecutor(max_workers=max_workers)
        self._running = False
        self._poll_thread: Optional[threading.Thread] = None
        self._poll_interval = DEFAULT_POLL_INTERVAL
        # In-flight lease registry: request_id -> lease grant metadata used by
        # the renewal loop and by fenced provider writes.
        self._active_leases: Dict[str, Dict[str, Any]] = {}
        self._leases_lock = threading.Lock()
        self._renewal_running = False
        self._renewal_thread: Optional[threading.Thread] = None
        self._renewal_interval = LEASE_RENEWAL_INTERVAL_SECONDS

    def _normalize_request(self, response: Dict[str, Any]) -> Dict[str, Any]:
        normalized = {
            "id": response.get("id"),
            "sessionId": response.get("sessionId", response.get("session_id")),
            "toolName": response.get("toolName", response.get("tool_name")),
            "status": response.get("status"),
            "input": response.get("input"),
            "createdAt": response.get("createdAt", response.get("created_at")),
            "updatedAt": response.get("updatedAt", response.get("updated_at")),
            "executingMachineId": response.get(
                "executingMachineId", response.get("executing_machine_id")
            ),
            "leasedBy": response.get("leasedBy", response.get("leased_by", "")),
            "leaseEpoch": int(
                response.get("leaseEpoch", response.get("lease_epoch", 0)) or 0
            ),
            "leaseExpiresAt": response.get(
                "leaseExpiresAt", response.get("lease_expires_at", "")
            ),
            "timeoutSeconds": int(
                response.get("timeoutSeconds", response.get("timeout_seconds", 0)) or 0
            ),
        }

        if response.get("result") not in (None, ""):
            try:
                normalized["result"] = json.loads(response.get("result"))
            except Exception:
                normalized["result"] = response.get("result")

        result_type = response.get("resultType", response.get("result_type"))
        if result_type:
            normalized["resultType"] = result_type

        if response.get("error"):
            normalized["error"] = response.get("error")

        stream_results = response.get("streamResults", response.get("stream_results"))
        if isinstance(stream_results, list) and stream_results:
            normalized["streamResults"] = stream_results

        return normalized

    def start_polling(self, poll_interval: float = DEFAULT_POLL_INTERVAL):
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
                except ToolplaneFailedPreconditionError:
                    # The lease was reclaimed or expired; stop renewing and
                    # let the fenced writes surface the loss.
                    logger.warning(
                        "Lease for request %s was lost; stopping renewal.",
                        request_id,
                    )
                    self.release_active_lease(request_id)
                except Exception as e:
                    logger.warning(
                        "Lease renewal failed for request %s: %s", request_id, e
                    )

    def renew_request_lease(
        self, session_id: str, request_id: str, machine_id: str, lease_epoch: int
    ) -> Dict[str, Any]:
        """Renew the execution lease for a claimed/running request."""
        response = self.connection_manager.renew_request_lease(
            session_id, request_id, machine_id, lease_epoch
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
            payload = {"sessionId": session_id, "status": "pending", "limit": limit}

            response = self.connection_manager.list_requests(payload)
            requests = response.get("requests", [])

            # Process each request
            for req in requests:
                request_id = req.get("id")

                try:
                    # Claim the request; the response carries the lease grant.
                    claimed = self.connection_manager.claim_request(
                        session_id, request_id, machine_id
                    )

                    # Track the lease grant so the renewal loop keeps it alive.
                    self.register_active_lease(
                        session_id,
                        request_id,
                        claimed.get("leasedBy", claimed.get("leased_by", machine_id))
                        or machine_id,
                        int(
                            claimed.get("leaseEpoch", claimed.get("lease_epoch", 0))
                            or 0
                        ),
                    )

                    # Execute in thread pool
                    self.executor.submit(
                        self._execute_request, claimed, tools, streaming_tools
                    )

                except Exception:
                    # Ignore claim errors (request might be claimed by another machine)
                    pass

        except Exception as e:
            raise RequestError(f"Failed to poll requests for session {session_id}: {e}")

    def _execute_request(
        self, request, tools: Dict[str, Callable], streaming_tools: set
    ):
        """Execute a claimed request."""
        tool_name = request.get("toolName", request.get("tool_name"))
        request_id = request.get("id")
        session_id = request.get("sessionId", request.get("session_id"))
        machine_id = request.get("leasedBy", request.get("leased_by", "")) or ""
        lease_epoch = int(request.get("leaseEpoch", request.get("lease_epoch", 0)) or 0)

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
                params = json.loads(request.get("input", "{}"))
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
        payload = {
            "sessionId": session_id,
            "requestId": request_id,
            "status": status,
            "machineId": machine_id,
            "leaseEpoch": lease_epoch,
        }

        self.connection_manager.update_request(payload)

    def _update_request(
        self,
        session_id: str,
        request_id: str,
        result_type: str,
        machine_id: str = "",
        lease_epoch: int = 0,
    ):
        """Update request with result type (fenced provider write)."""
        payload = {
            "sessionId": session_id,
            "requestId": request_id,
            "resultType": result_type,
            "machineId": machine_id,
            "leaseEpoch": lease_epoch,
        }

        self.connection_manager.update_request(payload)

    def _append_request_chunk(
        self,
        session_id: str,
        request_id: str,
        chunk: str,
        machine_id: str = "",
        lease_epoch: int = 0,
    ):
        """Append chunk to request (fenced provider write)."""
        payload = {
            "sessionId": session_id,
            "requestId": request_id,
            "chunks": [chunk],
            "resultType": "streaming",
            "machineId": machine_id,
            "leaseEpoch": lease_epoch,
        }

        self.connection_manager.append_request_chunks(payload)

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
        payload = {
            "sessionId": session_id,
            "requestId": request_id,
            "result": result,
            "resultType": result_type,
            "meta": {},
            "machineId": machine_id,
            "leaseEpoch": lease_epoch,
        }

        self.connection_manager.submit_request_result(payload)

    def _submit_error_result(
        self,
        session_id: str,
        request_id: str,
        error: str,
        machine_id: str = "",
        lease_epoch: int = 0,
    ):
        """Submit error result.

        A rejection that itself fails fencing (the lease was already lost,
        FAILED_PRECONDITION) is logged rather than propagated.
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
        except Exception as e:
            if not isinstance(e, ToolplaneFailedPreconditionError):
                raise
            logger.warning(
                "Could not submit rejection for request %s: lease was lost",
                request_id,
            )

    # ---------------- Consumer API ----------------

    def get_request_status(self, session_id: str, request_id: str) -> Dict[str, Any]:
        """Get request status."""
        try:
            self.connection_manager.ensure_connected()

            response = self.connection_manager.get_request(session_id, request_id)

            result = self._normalize_request(response)

            try:
                chunk_response = self.connection_manager.get_request_chunks(
                    session_id, request_id
                )
                if chunk_response.get("chunks"):
                    result["streamResults"] = chunk_response.get("chunks")
            except Exception:
                pass

            return result

        except Exception as e:
            raise RequestError(f"Failed to get request status: {e}")

    def resume_stream(self, session_id: str, request_id: str, last_seq: int = 0):
        """Resume a request chunk stream after the given absolute sequence.

        Yields chunk dicts (seq, request_id, chunk, is_final, error) covering
        everything the server still retains after last_seq. Raises
        ToolplaneInvalidArgumentError (OUT_OF_RANGE) when the retained window
        has moved past last_seq and replay is no longer possible.
        """
        response = None
        try:
            self.connection_manager.ensure_connected()
            response = self.connection_manager.stream_post(
                "api/ResumeStream",
                {"sessionId": session_id, "requestId": request_id, "lastSeq": last_seq},
            )

            for line in response.iter_lines(decode_unicode=True):
                if not line:
                    continue
                try:
                    chunk = json.loads(line)
                except ValueError:
                    continue
                if isinstance(chunk, dict) and isinstance(chunk.get("result"), dict):
                    chunk = chunk["result"]

                error_text = chunk.get("error", "") or ""
                if error_text:
                    raise api_error_from_http_response(
                        400,
                        json.dumps({"error": {"code": 11, "message": error_text}}),
                        context=f"Failed to resume stream for request {request_id}",
                    )

                value = chunk.get("chunk")
                if value not in (None, ""):
                    yield {
                        "seq": int(chunk.get("seq", 0) or 0),
                        "request_id": chunk.get("requestId", request_id),
                        "chunk": value,
                        "is_final": bool(chunk.get("isFinal", False)),
                        "error": "",
                    }
                if chunk.get("isFinal"):
                    return
        finally:
            if response is not None:
                response.close()

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

            payload = {
                "sessionId": session_id,
                "status": status,
                "toolName": tool_name,
                "limit": limit,
                "offset": offset,
            }
            response = self.connection_manager.list_requests(payload)
            requests = response.get("requests", [])
            return [self._normalize_request(entry) for entry in requests]

        except Exception as e:
            raise RequestError(f"Failed to list requests: {e}")

    def create_request(
        self,
        session_id: str,
        tool_name: str,
        input_data: str,
        timeout_seconds: int = 0,
        idempotency_key: str = "",
    ) -> str:
        """Create a new request.

        timeout_seconds optionally overrides the absolute execution timeout;
        zero keeps the server default. idempotency_key, when set, dedups
        creates within the session: retrying with the same key returns the
        original request instead of enqueueing duplicate work.
        """
        try:
            self.connection_manager.ensure_connected()

            payload = {
                "sessionId": session_id,
                "toolName": tool_name,
                "input": input_data,
            }
            if idempotency_key:
                payload["idempotencyKey"] = idempotency_key
            if timeout_seconds > 0:
                payload["timeoutSeconds"] = timeout_seconds

            response = self.connection_manager.create_request(payload)

            if response.get("error"):
                raise RequestError(f"Failed to create request: {response.get('error')}")

            return response.get("id")

        except Exception as e:
            raise RequestError(f"Failed to create request: {e}")

    def cancel_request(self, session_id: str, request_id: str) -> bool:
        """Cancel a request."""
        try:
            self.connection_manager.ensure_connected()

            response = self.connection_manager.cancel_request(session_id, request_id)
            if response.get("error"):
                raise RequestError(f"Failed to cancel request: {response.get('error')}")

            return bool(response.get("success", False))

        except Exception as e:
            raise RequestError(f"Failed to cancel request: {e}")

    def shutdown(self):
        """Shutdown request manager."""
        self.stop_polling()
        self.stop_lease_renewal()
        self.executor.shutdown(wait=False)
