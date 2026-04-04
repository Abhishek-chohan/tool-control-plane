#!/usr/bin/env python3
"""
Primary maintained Python SDK example.

This example demonstrates the control-plane-first path:
- connect to the gRPC server
- create a session
- register machine-backed tools
- keep the provider loop running so requests can be claimed and completed
"""

import os
import sys
import time

from toolplane import Toolplane

# Configuration
USER_ID = os.getenv("TOOLPLANE_USER_ID", "example-provider")
API_KEY = os.getenv("TOOLPLANE_API_KEY", "toolplane-conformance-fixture-key")
SERVER_HOST = os.getenv("TOOLPLANE_SERVER_HOST", "localhost")
SERVER_PORT = int(os.getenv("TOOLPLANE_SERVER_PORT", "9001"))
SESSION_NAME = os.getenv("TOOLPLANE_SESSION_NAME", "Python Control Plane Example")
SESSION_DESCRIPTION = os.getenv(
    "TOOLPLANE_SESSION_DESCRIPTION",
    "Provider-backed execution example for the primary Python SDK",
)
SESSION_NAMESPACE = os.getenv("TOOLPLANE_SESSION_NAMESPACE", "examples")


def main():
    print("=" * 60)
    print("Toolplane Python Provider Example")
    print("=" * 60)

    print(f"\n[1] Initializing maintained gRPC client...")
    print(f"    Server: {SERVER_HOST}:{SERVER_PORT}")
    print(f"    User ID: {USER_ID}")

    client = Toolplane(
        server_host=SERVER_HOST,
        server_port=SERVER_PORT,
        api_key=API_KEY,
        user_id=USER_ID,
        session_name=SESSION_NAME,
        session_description=SESSION_DESCRIPTION,
        session_namespace=SESSION_NAMESPACE,
        heartbeat_interval=60,
        max_workers=5,
        request_timeout=30,
    )
    provider = client.provider_runtime()

    # Connect to server
    print("\n[2] Preparing explicit provider runtime...")
    try:
        client.connect()
        print("    ✓ Connected successfully")
        print("    ✓ Explicit provider runtime is ready")
    except Exception as e:
        print(f"    ✗ Connection failed: {e}")
        sys.exit(1)

    print("\n[3] Creating session...")
    try:
        session = provider.create_session(
            user_id=USER_ID,
            name=SESSION_NAME,
            description=SESSION_DESCRIPTION,
            namespace=SESSION_NAMESPACE,
        )
        session_id = session.session_id
        print(f"    ✓ Session created: {session_id}")
        print(f"    Export TOOLPLANE_SESSION_ID={session_id} before running example_user.py")
    except Exception as e:
        print(f"    ✗ Session creation failed: {e}")
        sys.exit(1)

    # Register tools
    print("\n[4] Registering machine-backed tools...")

    @provider.tool(
        session_id=session_id,
        name="session_status",
        description="Summarize provider and session context for an operator or automation client",
        tags=["session", "status", "provider"],
    )
    def session_status(requester: str, objective: str = "status") -> dict:
        return {
            "requester": requester,
            "objective": objective,
            "provider_user_id": USER_ID,
            "session_id": session_id,
            "session_namespace": SESSION_NAMESPACE,
            "server": f"{SERVER_HOST}:{SERVER_PORT}",
        }

    print("    ✓ Registered tool: session_status")

    @provider.tool(
        session_id=session_id,
        name="text_transform",
        description="Transform rollout or operator text into a requested presentation style",
        tags=["text", "transform", "operations"],
    )
    def text_transform(text: str, style: str = "title") -> dict:
        operations = {
            "uppercase": lambda value: value.upper(),
            "lowercase": lambda value: value.lower(),
            "title": lambda value: value.title(),
            "reverse": lambda value: value[::-1],
        }

        if style not in operations:
            return {"error": f"Unknown style: {style}", "available_styles": list(operations)}

        return {
            "style": style,
            "original": text,
            "transformed": operations[style](text),
        }

    print("    ✓ Registered tool: text_transform")

    @provider.tool(
        session_id=session_id,
        name="incident_brief",
        description="Create a concise provider-side summary for an incident or change event",
        tags=["operations", "summary", "incident"],
    )
    def incident_brief(service: str, state: str, owner: str = "platform-team") -> dict:
        return {
            "service": service,
            "state": state,
            "owner": owner,
            "next_step": f"Coordinate follow-up for {service} with {owner}",
        }

    print("    ✓ Registered tool: incident_brief")

    print("\n[5] Provider tool catalog...")
    print("    session_status, text_transform, incident_brief")

    print("\n[6] Starting provider loop...")
    print("    Press Ctrl+C to stop")
    print()

    try:
        provider.start_in_background()

        print("    Provider is running and ready to claim requests...")
        print("    Run example_user.py with the printed TOOLPLANE_SESSION_ID to invoke work")
        print()

        while True:
            time.sleep(1)

    except KeyboardInterrupt:
        print("\n\n[7] Shutting down...")
        provider.stop()
        client.disconnect()
        print("    ✓ Client stopped successfully")
        print("\n" + "=" * 60)


if __name__ == "__main__":
    main()
