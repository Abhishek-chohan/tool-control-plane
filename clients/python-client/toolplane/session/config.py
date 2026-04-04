from typing import Annotated, Dict, List, Optional, Union

from pydantic import BaseModel

__all__ = ["SessionConfig", "parse_session_config"]


class SessionConfig(BaseModel):
    name: Annotated[str, "Session Name"]
    description: Annotated[str, "Session Description"]
    namespace: Annotated[str, "Session Namespace"]


def parse_session_config(
    sessions: Optional[
        Union[
            SessionConfig,
            tuple[str, SessionConfig],
            List[SessionConfig],
            List[tuple[str, SessionConfig]],
            Dict[str, SessionConfig],
        ]
    ],
) -> Dict[Optional[str], SessionConfig]:
    """
    Parse various session configuration formats into a unified dict.

    Returns:
        Dict[Optional[str], SessionConfig]: Mapping of session_id (or None for auto-gen) to SessionConfig
    """
    if sessions is None:
        return {}

    result = {}

    if isinstance(sessions, SessionConfig):
        # Single session with auto-generated ID
        result[None] = sessions

    elif isinstance(sessions, tuple) and len(sessions) == 2:
        # Single session with specific ID
        session_id, config = sessions
        if isinstance(config, SessionConfig):
            result[session_id] = config

    elif isinstance(sessions, list):
        # List of sessions
        for item in sessions:
            if isinstance(item, SessionConfig):
                # Auto-generated ID
                result[None] = item
            elif isinstance(item, tuple) and len(item) == 2:
                # Specific ID
                session_id, config = item
                if isinstance(config, SessionConfig):
                    result[session_id] = config

    elif isinstance(sessions, dict):
        # Dict mapping session_id to SessionConfig
        for session_id, config in sessions.items():
            if isinstance(config, SessionConfig):
                result[session_id] = config

    return result
