# Python Client Examples

These examples show the Python package as the primary maintained SDK for the Toolplane control plane. Start here if you want a provider-backed flow that matches the current repo story: create or join a session, register a machine-backed provider, invoke work through the server, and inspect the results.

## First-Touch Path

1. Start the provider in `example_client.py`.
2. Copy the printed `TOOLPLANE_SESSION_ID` value.
3. Run `example_user.py` with that session ID to invoke provider-backed work.

The examples default to the same local fixture contract used by CI and local development:

- `TOOLPLANE_SERVER_HOST=localhost`
- `TOOLPLANE_SERVER_PORT=9001`
- `TOOLPLANE_API_KEY=toolplane-conformance-fixture-key`

## Files

### `example_client.py` - Provider And Session Bootstrap

What it demonstrates:

- Connect to the maintained gRPC control-plane surface.
- Construct the explicit Python `ProviderRuntime` surface.
- Create a machine-backed session for an operator or automation user.
- Register machine-backed tools for remote execution.
- Start the provider loop so the server can route requests to this machine.

Usage:

```bash
python example_client.py
```

The script prints the created session ID so you can export it before running `example_user.py`.

### `example_user.py` - Request Execution Against An Existing Session

What it demonstrates:

- Connect to an existing session created by `example_client.py`.
- Inspect the registered tool catalog.
- Execute synchronous work and one asynchronous request.
- Poll request state through the control plane.

Usage:

```bash
TOOLPLANE_SESSION_ID=<session-id-from-provider> python example_user.py
```

You can also pass the session ID as the first command-line argument.

### `example.py` - Experimental Toolkit Integration

This file is an experimental LangChain toolkit integration sample. It is not the primary getting-started path. Use `example_client.py` and `example_user.py` first if you want the maintained control-plane flow.

## Environment

All examples accept environment overrides instead of shipping fixed IDs or secrets:

```bash
export TOOLPLANE_SERVER_HOST=localhost
export TOOLPLANE_SERVER_PORT=9001
export TOOLPLANE_API_KEY=toolplane-conformance-fixture-key
export TOOLPLANE_USER_ID=example-operator
```

If the provider prints a session ID, export it before running the user example:

```bash
export TOOLPLANE_SESSION_ID=<session-id-from-example-client>
python example_user.py
```

## Example Flow

```text
example_client.py
  -> connects to gRPC server
  -> creates explicit provider runtime
  -> creates machine-backed session
  -> registers machine-backed tools
  -> starts provider loop

example_user.py
  -> connects to same session
  -> lists tools
  -> invokes provider-backed work
  -> polls request status for async work
```

## Notes

- The provider must be running for request execution to complete.
- The maintained provider path is the explicit Python `ProviderRuntime`, not an implicit background loop owned by consumer startup.
- The example API key is a non-secret local-development fixture, not a production credential.
- The HTTP gateway remains available through `ToolplaneHTTP`, but these examples intentionally lead with the maintained gRPC path.
