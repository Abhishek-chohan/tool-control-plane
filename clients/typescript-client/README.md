# Toolplane TypeScript Client

This package is the maintained JavaScript-family SDK for Toolplane's durable remote tool-execution control plane. Use it when TypeScript or Node code needs explicit provider runtime ownership, request lifecycle control, and drain-safe machine management over the canonical gRPC contract. Its public surface is gRPC-only and centers on session, tool, machine, request, and task lifecycle helpers.

## When To Start Here

- Your JavaScript-family code needs the maintained gRPC control-plane surface plus an explicit `ProviderRuntime`.
- You want a maintained provider or consumer path in TypeScript without treating the internal HTTP conformance adapters as public SDK surface.
- You want the same provider, request, recovery, and drain model as the broader platform story, but on the TypeScript-maintained path.

## Support Status

- Supported secondary SDK and the maintained JavaScript-family path.
- The maintained control-plane story is the gRPC surface for session, tool, machine, request, and task lifecycle helpers.
- This SDK now ships a maintained explicit `ProviderRuntime` for the gRPC provider lifecycle: session create or attach, machine registration, request claim, heartbeat, chunk append, result submission, and drain-aware shutdown.
- Repository-internal HTTP adapters still exist under `tests/conformance/` so the shared fixture runner can exercise the maintained HTTP gateway. They are not part of the public SDK surface.

## Features

- Live grpc-js wrappers for the maintained public helpers.
- Explicit `ProviderRuntime` lifecycle management for machine-backed providers.
- Public `executeTool()` plus numeric convenience helpers `add()`, `subtract()`, `multiply()`, and `divide()`.
- Tool discovery helpers `listTools()`, `getToolById()`, `getToolByName()`, and `deleteTool()`.
- Session, API-key, machine, request, and task lifecycle helpers, including provider-owned request claim and result submission.
- Promise-based connection management with `connect()`, `disconnect()`, `isConnected()`, and `getConnectionStatus()`.
- Shared-fixture conformance coverage in `tests/conformance/`.

## Installation

```bash
npm install
npm run build
```

## Canonical Flow

The canonical end-to-end path for Toolplane is: register a provider, create or attach a session, execute a request, stream or recover results, and drain the machine. The TypeScript SDK ships that provider loop directly through the explicit `ProviderRuntime` over the maintained gRPC path.

## Quick Start

```typescript
import { ToolplaneClient } from './src';

const client = ToolplaneClient.createGRPCClient(
  'localhost',
  9001,
  'grpc-session',
  'grpc-user',
  process.env.TOOLPLANE_API_KEY || 'toolplane-conformance-fixture-key',
  {
    enabled: true,
    caCertPath: '../../server/deploy/reference/certs/ca.crt',
    serverName: 'localhost',
  },
);

await client.connect();

const session = await client.createSession(
  'My Session',
  'Session created from the TypeScript client',
  'examples',
);

const machine = await client.registerMachine('', '1.0.0', [
  {
    sessionId: session.id,
    name: 'session_status',
    description: 'Report session context for a connected operator or automation client',
    schema: JSON.stringify({
      type: 'object',
      properties: {
        requester: { type: 'string' },
      },
      required: ['requester'],
    }),
    config: { version: '1.0' },
    tags: ['session', 'status'],
  },
]);

const tools = await client.listTools();
console.log(`Session ${session.id} has ${tools.length} tools on machine ${machine.id}`);

await client.disconnect();
```

Tool registration is machine-aware, so register a machine first or embed tool definitions in `registerMachine(...)`.

### Explicit Provider Runtime

```typescript
import { ToolplaneClient } from './src';

const client = ToolplaneClient.createGRPCClient(
  'localhost',
  9001,
  '',
  'provider-user',
  process.env.TOOLPLANE_API_KEY || 'toolplane-conformance-fixture-key',
  {
    enabled: true,
    caCertPath: '../../server/deploy/reference/certs/ca.crt',
    serverName: 'localhost',
  },
);

await client.connect();

const provider = client.providerRuntime({
  pollIntervalMs: 250,
  heartbeatIntervalMs: 30_000,
});

const session = await provider.createSession({
  name: 'TypeScript Provider Session',
  description: 'Maintained provider runtime example',
  namespace: 'examples',
});

await provider.tool(
  {
    sessionId: session.id,
    name: 'echo_tool',
    description: 'Echo the caller message',
    schema: {
      type: 'object',
      properties: {
        message: { type: 'string' },
      },
      required: ['message'],
    },
  },
  async (input) => ({ echo: input.message }),
);

await provider.startInBackground();
```

See `src/examples/provider_runtime_example.ts` for a maintained end-to-end provider loop with graceful drain.

## Public API Snapshot

### Lifecycle

```typescript
connect(): Promise<void>
disconnect(): Promise<void>
isConnected(): boolean
getConnectionStatus(): ConnectionStatus
providerRuntime(options?: ProviderRuntimeOptions): ProviderRuntime
```

### Execution

```typescript
executeTool(toolName: string, params?: Record<string, unknown>): Promise<Request>
add(a: number, b: number): Promise<number>
subtract(a: number, b: number): Promise<number>
multiply(a: number, b: number): Promise<number>
divide(a: number, b: number): Promise<number>
ping(): Promise<string>
```

These helpers use the live request lifecycle on the Go server. They require a registered machine-backed provider to claim requests and submit results.

### Lifecycle Wrappers

The client exposes public wrappers for:

- Sessions: `createSession`, `getSession`, `listSessions`, `updateSession`
- API keys: `createApiKey`, `listApiKeys`, `revokeApiKey`
- Tools: `registerTool`, `listTools`, `getToolById`, `getToolByName`, `deleteTool`
- Machines: `registerMachine`, `listMachines`, `getMachine`, `updateMachinePing`, `unregisterMachine`, `drainMachine`
- Requests: `createRequest`, `getRequest`, `listRequests`, `updateRequest`, `claimRequest`, `appendRequestChunks`, `submitRequestResult`, `cancelRequest`
- Tasks: `createTask`, `getTask`, `listTasks`, `cancelTask`

### Provider Runtime

The explicit `ProviderRuntime` exposes:

- Session ownership: `createSession()`, `attachSession()`, `managedSessionIds()`
- Tool registration: `registerTool()`, `tool()`
- Runtime control: `pollOnce()`, `startInBackground()`, `runForever()`, `stop()`, `drain()`, `close()`

For the reference deployment, set `TOOLPLANE_SERVER_PORT=9001`, `TOOLPLANE_USE_TLS=true`, `TOOLPLANE_TLS_CA_CERT_PATH=/path/to/ca.crt`, and `TOOLPLANE_TLS_SERVER_NAME=localhost` before running the maintained examples.

## Error Handling

The client exposes a small public error hierarchy:

```typescript
import {
  ToolplaneError,
  ConnectionError,
  TimeoutError,
  ProtocolError,
} from './src';

try {
  await client.listTools();
} catch (error) {
  if (error instanceof ConnectionError) {
    console.log(`Connection error: ${error.message}`);
  } else if (error instanceof ToolplaneError) {
    console.log(`Toolplane error: ${error.message}`);
  }
}
```

## Scripts

```bash
npm start
npm run dev
npm run example
npm run provider
npm run advanced
npm test
npm run test:conformance
npm run test:all
```

## Conformance

- The shared fixture runner lives under `tests/conformance/` and loads repository-wide cases from `../../conformance/cases/`.
- The public SDK remains gRPC-only.
- The repository-internal conformance runner uses both gRPC and HTTP adapters so fixture behavior is validated across the maintained server transports.
- The gRPC provider-runtime cases are now backed by the public `ProviderRuntime` surface rather than a test-only request-processing loop.

## Compatibility

- Node.js: 18+
- TypeScript: 5.0+
- Maintained server path: the Go server gRPC control-plane surface
- Public protocol: gRPC with Protocol Buffers
