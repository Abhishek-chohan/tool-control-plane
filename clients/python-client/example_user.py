#!/usr/bin/env python3
"""
Primary maintained consumer-side example for the Python SDK.

This example demonstrates:
- connecting to an existing provider-backed session
- listing the registered control-plane tools
- invoking synchronous and asynchronous work through the server
"""

import json
import os
import sys
import time

from toolplane import Toolplane

# Configuration
USER_ID = os.getenv("TOOLPLANE_USER_ID", "example-operator")
API_KEY = os.getenv("TOOLPLANE_API_KEY", "toolplane-conformance-fixture-key")
SERVER_HOST = os.getenv("TOOLPLANE_SERVER_HOST", "localhost")
SERVER_PORT = int(os.getenv("TOOLPLANE_SERVER_PORT", "9001"))
SESSION_ID = os.getenv("TOOLPLANE_SESSION_ID", "")


def resolve_session_id():
    if len(sys.argv) > 1 and sys.argv[1].strip():
        return sys.argv[1].strip()
    if SESSION_ID:
        return SESSION_ID
    raise SystemExit(
        "Set TOOLPLANE_SESSION_ID or pass the session ID as the first argument. Run example_client.py first."
    )


def list_tools(client, session_id):
    """List all tools in a session."""
    registry = client.get_available_tools(session_id)
    tools = registry.get("tools", [])
    response = []
    for tool in tools:
        name = tool.get("name", "Unknown")
        description = tool.get("description", "No description")
        schema = tool.get("schema", {})
        response.append(
            {
                "name": name,
                "description": description,
                "schema": schema,
            }
        )
    return response


def list_and_print_tool_schemas(client, session_id):
    """List tools in a session and print the schema for each tool."""
    try:
        tools = list_tools(client, session_id)

        if not tools:
            print("    No tools found in the session.")
            return

        print(f"    Total tools: {len(tools)}")
        for tool in tools:
            schema = tool.get("schema", {})
            print(f"\n    Tool: {tool['name']}")
            print(f"    Schema: {json.dumps(schema, indent=4)}")
    except Exception as e:
        print(f"    ✗ Failed to list and print tool schemas: {e}")


def wait_for_request_completion(client, session_id, request_id, timeout_seconds=20):
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        request_state = client.get_request_status(request_id, session_id)
        status = request_state.get("status", "unknown")
        if status in {"done", "failure", "stalled"}:
            return request_state
        time.sleep(0.5)
    raise TimeoutError(f"Timed out waiting for request {request_id}")


def main():
    print("=" * 60)
    print("Toolplane Python Consumer Example")
    print("=" * 60)
    session_id = resolve_session_id()

    client = Toolplane(
        server_host=SERVER_HOST,
        server_port=SERVER_PORT,
        api_key=API_KEY,
        user_id=USER_ID,
        session_ids=[session_id],
        request_timeout=120,
    )

    print(f"\n[1] Configuration")
    print(f"    Server: {SERVER_HOST}:{SERVER_PORT}")
    print(f"    User ID: {USER_ID}")
    print(f"    Session ID: {session_id}")

    print(f"\n[2] Connecting to Toolplane server...")
    try:
        if not client.connect():
            raise Exception("Failed to connect")
        print("    ✓ Connected")
    except Exception as e:
        print(f"    ✗ Failed to connect: {e}")
        return

    print(f"\n[3] Inspecting the provider tool catalog...")
    try:
        tools = list_tools(client, session_id)
        tool_names = [tool["name"] for tool in tools]
        print(f"    Available tools: {', '.join(tool_names)}")
        print(f"    Total: {len(tools)} tools")

        list_and_print_tool_schemas(client, session_id)
    except Exception as e:
        print(f"    ✗ Failed to get tools: {e}")
        client.disconnect()
        return

    print("\n[4] Invoking provider-backed tools...")
    print("-" * 60)

    print("\n[Test 1] Invoking 'session_status'")
    try:
        result = client.invoke(
            "session_status",
            session_id,
            requester=USER_ID,
            objective="verify provider availability",
        )
        print(f"    Result: {json.dumps(result, indent=2)}")
        print("    ✓ Success")
    except Exception as e:
        print(f"    ✗ Failed: {e}")

    time.sleep(0.5)

    print("\n[Test 2] Invoking 'text_transform'")
    try:
        result = client.invoke(
            "text_transform",
            session_id,
            text="prepare gateway rollout note",
            style="uppercase",
        )
        print(f"    Result: {json.dumps(result, indent=2)}")
        print("    ✓ Success")
    except Exception as e:
        print(f"    ✗ Failed: {e}")

    time.sleep(0.5)

    print("\n[Test 3] Creating async request for 'incident_brief'")
    try:
        request_id = client.ainvoke(
            "incident_brief",
            session_id,
            service="http-gateway",
            state="degraded",
            owner="platform-team",
        )
        print(f"    Request ID: {request_id}")
        request_state = wait_for_request_completion(client, session_id, request_id)
        print(f"    Final request state: {json.dumps(request_state, indent=2)}")
        print("    ✓ Success")
    except Exception as e:
        print(f"    ✗ Failed: {e}")

    client.disconnect()

    print("\n" + "=" * 60)
    print("Provider-backed request execution example completed!")
    print("=" * 60)
    print()


if __name__ == "__main__":
    main()
