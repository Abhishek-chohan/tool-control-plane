# Python Client Examples

These examples show the maintained first-offload path for the Python SDK, not a generic SDK tour. Start here when you want to keep the rest of your caller or orchestration stack in place and move one painful remote tool behind the Toolplane control plane: create or join a session, register a machine-backed provider, invoke work through the server, and inspect the results.

## Migration Claim

The intended adoption shape is narrow.

- Keep existing tool selection or orchestration logic outside Toolplane.
- Bind one painful remote tool family to one explicit Toolplane session.
- Let Toolplane own request lifecycle, inspection, and drain for that tool only.
- Move a second tool only after the first one proves its operational value.

## Reference Workload

A concrete workload to keep in mind is one sandboxed code-execution worker. It can run longer than the original caller interaction and may eventually need streaming logs, cancellation, inspection, or drain-safe rollout. The bundled sample tools are intentionally simple, but the provider-consumer shape below is the maintained starting point for that first offload.

## Stepwise First-Tool Migration

1. Start the provider in `example_client.py` so one explicit session owns the provider-backed tool catalog.
2. Copy the printed `TOOLPLANE_SESSION_ID` value.
3. Run `example_user.py` with that session ID so a separate caller can discover tools and submit work through the control plane.
4. Replace one sample tool with the first real long-running or restart-sensitive candidate instead of trying to move every tool at once.
5. Rehearse inspection, cancellation, and drain on that same session before you expand the scope.

## First Offload Path

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
- The minimal provider shape you would reuse when replacing the sample tools with your first offloaded remote tool.

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
- The consumer-side half of a first-tool offload without replacing the rest of the caller stack.

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

## Validation Beyond The Happy Path

- Inspection: `example_user.py` already polls async request state through the control plane. Keep that in the first evaluation instead of treating the session as a fire-and-forget queue.
- Cancellation: once you swap one sample tool for the real long-running candidate, create one async request and rehearse `cancel_request(session_id, request_id)` on the same session.
- Drain: after a replacement provider is online for the same session, use `list_machines(session_id)` and `drain_machine(session_id, machine_id)` to verify safe handoff. Use `server/docs/operator-runbook.md` and `server/docs/reference-deployment.md` for the maintained completion checks.
- Coexistence: keep direct local tools and the surrounding orchestrator outside Toolplane until the first offloaded tool proves its value.

## Notes

- The provider must be running for request execution to complete.
- The maintained provider path is the explicit Python `ProviderRuntime`, not an implicit background loop owned by consumer startup.
- The example API key is a non-secret local-development fixture, not a production credential.
- The HTTP gateway remains available through `ToolplaneHTTP`, but these examples intentionally lead with the maintained gRPC path.
