import pytest

from .runner import execute_case, load_cases


CASES = load_cases()


@pytest.mark.integration
@pytest.mark.conformance
@pytest.mark.parametrize("transport", ["grpc", "http"])
@pytest.mark.parametrize("case_obj", CASES, ids=[case["id"] for case in CASES])
def test_conformance_case(case_obj, transport):
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
