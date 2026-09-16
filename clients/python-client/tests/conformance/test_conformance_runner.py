import os

import pytest

from .runner import (
    MCP_TRANSPORT_FEATURES,
    _adapter_for_transport,
    assert_request_field_non_empty,
    assert_request_status,
    execute_case,
    load_cases,
)

CASES = load_cases()

# Connectivity failures are only converted to skips when the run explicitly
# targets an externally managed environment
# (TOOLPLANE_CONFORMANCE_ALLOW_SKIP=1). In the default auto-boot mode a
# connectivity error is a real failure: the bootstrap owns the server
# lifecycle, and "deadline exceeded" in particular is a plausible symptom of a
# dispatch/lease regression, not an environment problem.
CONNECTIVITY_SKIP_TOKENS = (
    "failed to connect",
    "connection refused",
    "connection reset",
    "unavailable",
)


def _allow_connectivity_skips() -> bool:
    return os.getenv("TOOLPLANE_CONFORMANCE_ALLOW_SKIP", "0").strip().lower() in {
        "1",
        "true",
        "yes",
    }


def _supports_environment(case_obj, transport):
    """Skip cases that cannot run in the current conformance environment.

    The multi_instance feature requires the gRPC transport and a second server
    replica (booted behind TOOLPLANE_CONFORMANCE_MULTI_INSTANCE=1). In the
    standard single-instance run it is skipped so the suite stays green without
    provisioning Postgres.

    The mcp transport exercises only the tool-plane features
    (MCP_TRANSPORT_FEATURES) through the MCP gateway facade; management-plane
    fixtures stay on grpc/http. The mcp_tasks fixture conversely runs only on
    the mcp transport. Both need TOOLPLANE_CONFORMANCE_MCP=1 so the bootstrap
    boots cmd/mcp-gateway.
    """
    feature = case_obj.get("feature")
    if feature == "multi_instance":
        if transport != "grpc":
            pytest.skip("multi_instance conformance targets gRPC server instances only")
        if not os.getenv("TOOLPLANE_CONFORMANCE_GRPC_PORT_B"):
            pytest.skip(
                "multi_instance conformance requires TOOLPLANE_CONFORMANCE_MULTI_INSTANCE=1 "
                "plus TOOLPLANE_DATABASE_URL"
            )
    if transport == "mcp":
        if not os.getenv("TOOLPLANE_CONFORMANCE_MCP_PORT"):
            pytest.skip(
                "mcp conformance requires TOOLPLANE_CONFORMANCE_MCP=1 (cmd/mcp-gateway not booted)"
            )
        if feature not in MCP_TRANSPORT_FEATURES:
            pytest.skip(f"feature {feature} is not covered by the mcp transport")
    elif feature == "mcp_tasks":
        pytest.skip("mcp_tasks runs only on the mcp transport")
    return True


@pytest.mark.integration
@pytest.mark.conformance
@pytest.mark.parametrize("transport", ["grpc", "http", "mcp"])
@pytest.mark.parametrize("case_obj", CASES, ids=[case["id"] for case in CASES])
def test_conformance_case(case_obj, transport):
    _supports_environment(case_obj, transport)
    try:
        execute_case(case_obj, transport)
    except Exception as exc:
        message = str(exc).lower()
        if _allow_connectivity_skips() and any(
            token in message for token in CONNECTIVITY_SKIP_TOKENS
        ):
            pytest.skip(f"transport {transport} unavailable for conformance run: {exc}")
        raise


@pytest.mark.conformance
@pytest.mark.parametrize("transport", ["grpc", "http"])
def test_tool_failure_records_failure_status(transport):
    """A provider tool that raises records FAILED — not DONE with error text —
    so status-level consumers (retries, eval scoring, isError rendering) see
    the failure without parsing output strings."""
    adapter = _adapter_for_transport(transport, "conformance-user")
    session_id = ""
    try:
        adapter.connect()
        session_id = adapter.create_session(
            {
                "user_id": "conformance-user",
                "name": "failure-fidelity",
                "description": "tool failure fidelity case",
                "namespace": "conformance",
            }
        )
        adapter.register_failing_tool(
            session_id=session_id,
            tool_name="always_fails",
            description="raises intentionally",
        )
        adapter.start_provider_runtime(session_id)
        request_id = adapter.create_request(
            session_id, "always_fails", {"message": "trigger"}
        )

        # Poll to terminal directly (the adapter's wait helper treats
        # failure as an exception; here failure IS the expected outcome).
        import time as _time

        deadline = _time.time() + 30
        status: dict = {}
        while _time.time() < deadline:
            status = adapter.get_request_status(session_id, request_id)
            if str(status.get("status")) in ("done", "failed", "cancelled"):
                break
            _time.sleep(0.1)

        assert_request_status(status, "failed", "tool_failure", transport)
        assert "intentional" in str(status.get("error", "")), (
            f"failure must carry the tool's error text: {status}"
        )
    finally:
        adapter.close()
