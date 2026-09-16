"""The derived wait must stay inside the server's wait ceiling.

The server rejects ``wait_timeout_seconds`` above 3600 (the
``timeout_seconds`` ceiling) with OUT_OF_RANGE / TIMEOUT_ABOVE_MAX. The
SDK derives its wait as ``timeout_seconds + 15`` — which would trip that
rejection for a legal timeout at the ceiling — so derived waits clamp to
the ceiling while explicit caller choices pass through untouched (the
server rejects those loudly; the SDK must not silently shrink them).
"""

from toolplane.common.constants import WAIT_TIMEOUT_MAX_SECONDS
from toolplane.core.session_context import derive_wait_for


def test_default_wait_without_timeout():
    assert derive_wait_for(0, None) == 60


def test_derived_wait_tracks_timeout_plus_margin():
    assert derive_wait_for(45, None) == 60


def test_derived_wait_clamps_at_server_ceiling():
    # A legal timeout at the ceiling plus the completion margin must not
    # trip the server's wait rejection.
    assert derive_wait_for(WAIT_TIMEOUT_MAX_SECONDS, None) == WAIT_TIMEOUT_MAX_SECONDS
    assert derive_wait_for(WAIT_TIMEOUT_MAX_SECONDS - 5, None) == WAIT_TIMEOUT_MAX_SECONDS


def test_explicit_wait_passes_through():
    # Explicit budgets are the caller's choice: over-max values surface the
    # server's loud rejection instead of a silent clamp.
    assert derive_wait_for(0, 7200) == 7200
    assert derive_wait_for(45, 5) == 5
