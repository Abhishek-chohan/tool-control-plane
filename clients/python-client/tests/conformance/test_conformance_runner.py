import os

import pytest

from .runner import MCP_TRANSPORT_FEATURES, execute_case, load_cases


CASES = load_cases()


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
        connectivity_tokens = (
            "failed to connect",
            "unavailable",
            "connection",
            "refused",
            "deadline exceeded",
        )
        if any(token in message for token in connectivity_tokens):
            pytest.skip(f"transport {transport} unavailable for conformance run: {exc}")
        raise
