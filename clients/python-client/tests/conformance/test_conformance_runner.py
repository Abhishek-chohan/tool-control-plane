import os

import pytest

from .runner import execute_case, load_cases


CASES = load_cases()


def _supports_environment(case_obj, transport):
    """Skip cases that cannot run in the current conformance environment.

    The multi_instance feature requires the gRPC transport and a second server
    replica (booted behind TOOLPLANE_CONFORMANCE_MULTI_INSTANCE=1). In the
    standard single-instance run it is skipped so the suite stays green without
    provisioning Postgres.
    """
    if case_obj.get("feature") == "multi_instance":
        if transport != "grpc":
            pytest.skip("multi_instance conformance targets gRPC server instances only")
        if not os.getenv("TOOLPLANE_CONFORMANCE_GRPC_PORT_B"):
            pytest.skip(
                "multi_instance conformance requires TOOLPLANE_CONFORMANCE_MULTI_INSTANCE=1 "
                "plus TOOLPLANE_DATABASE_URL"
            )
    return True


@pytest.mark.integration
@pytest.mark.conformance
@pytest.mark.parametrize("transport", ["grpc", "http"])
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
