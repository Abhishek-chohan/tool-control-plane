#!/usr/bin/env python3
"""Experimental toolkit-integration sample for the Python SDK.

This file is not the first-touch getting-started path. Start with
`example_client.py` and `example_user.py` for the maintained control-plane flow.
"""

import os
import sys
import time

from langchain_core.tools import ToolException

from toolplane import Toolplane
from toolplane.toolkits.swe.swe_toolkit import get_swe_toolkit

# Get configuration
config = {
    "server_host": os.getenv("TOOLPLANE_SERVER_HOST", "localhost"),
    "server_port": int(os.getenv("TOOLPLANE_SERVER_PORT", "9001")),
    "user_id": os.getenv("TOOLPLANE_USER_ID", "toolkit-integration-user"),
    "api_key": os.getenv("TOOLPLANE_API_KEY", "toolplane-conformance-fixture-key"),
    "session_id": os.getenv("TOOLPLANE_SESSION_ID", ""),
    "session_name": "swe-toolkit-session",
    "session_description": "Software Engineering Toolkit via Toolplane",
    "namespace": "swe-tools",
    "enable_streaming": True,
    "enable_batch_operations": True,
    "max_file_size": 10 * 1024 * 1024,  # 10MB
    "max_search_results": 100,
    "default_chunk_size": 1024,
    "default_stream_delay": 0.1,
}

if len(sys.argv) >= 2 and sys.argv[1].strip():
    config["session_id"] = sys.argv[1].strip()

if not config["session_id"]:
    print("❌ Error: set TOOLPLANE_SESSION_ID or pass the session ID as the first argument.")
    sys.exit(1)

client = Toolplane(
    server_host=config["server_host"],
    server_port=config["server_port"],
    user_id=config["user_id"],
    api_key=config["api_key"],
    session_ids=[config["session_id"]],
)
client.connect()
provider = client.provider_runtime([config["session_id"]])
provider.attach_session(config["session_id"], register_machine=True)


def convert_langchain_tool_to_toolplane_tool(
    langchain_tool,
) -> tuple[str, callable, dict, bool]:
    """
    Convert a LangChain StructuredTool into Toolplane tool components:
      (name, func, schema_dict, stream_flag).
    """
    name = langchain_tool.name
    description = getattr(langchain_tool, "description", "") or ""
    args_schema = getattr(langchain_tool, "args_schema", {}) or {}
    schema = {}
    args_schema = args_schema.model_json_schema()
    schema["name"] = name
    schema["description"] = description
    schema["schema"] = args_schema

    def func(**kwargs):
        # Try sync run first
        if hasattr(langchain_tool, "run"):

            return langchain_tool.run(kwargs)
        # Fallback to async
        elif hasattr(langchain_tool, "arun"):
            import asyncio

            return asyncio.get_event_loop().run_until_complete(
                langchain_tool.arun(kwargs)
            )

        else:
            raise ToolException(f"Tool '{name}' has no run()/arun()")

    # By default we treat it as non‐streaming
    stream_flag = False
    return name, func, schema, description, stream_flag


def register_langchain_tool(provider, session_id: str, langchain_tool, stream: bool = False):
    """
    Register a LangChain StructuredTool as an Toolplane tool.
    """
    tool_name, func, tool_schema, description, detected_stream = (
        convert_langchain_tool_to_toolplane_tool(langchain_tool)
    )
    # Use user‐provided stream override or the detected flag
    provider.register_tool(
        session_id=session_id,
        name=tool_name,
        func=func,
        schema=tool_schema,
        description=description,
        stream=detected_stream,
        tags=[tool_name],
    )


tools = get_swe_toolkit()
# tools = get_standalone_toolkit()
for tool in tools:
    register_langchain_tool(
        provider,
        config["session_id"],
        tool,
        stream=False,
    )

if __name__ == "__main__":
    try:
        print("\n🚀 Starting explicit provider runtime...")
        provider.start_in_background()

        print("✅ SWE Toolkit Toolplane server is running!")
        print("Press Ctrl+C to stop the server.")

        # Keep the client running
        while True:
            time.sleep(1)

    except KeyboardInterrupt:
        print("\n🛑 Shutting down SWE Toolkit Toolplane server...")
        provider.stop()
        print("✅ Server stopped successfully.")
        sys.exit(0)

    except Exception as e:
        print(f"❌ Error: {e}")
        provider.stop()
        sys.exit(1)

    finally:
        client.disconnect()
