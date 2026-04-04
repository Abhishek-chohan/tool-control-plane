# Toolplane TypeScript Client

This package is the maintained JavaScript-family SDK for Toolplane's remote tool-execution control plane. Its public surface is gRPC-only and centers on session, tool, machine, request, and task lifecycle helpers.

## Support Status

- Supported secondary SDK and the maintained JavaScript-family path.
- The maintained control-plane story is the gRPC surface for session, tool, machine, request, and task lifecycle helpers.
- This SDK does not ship a maintained provider runtime harness that claims requests, renews heartbeats, and submits results. Use Python's explicit `ProviderRuntime` for the maintained provider-mode story.
- Repository-internal HTTP adapters still exist under `tests/conformance/` so the shared fixture runner can exercise the maintained HTTP gateway. They are not part of the public SDK surface.

## Features

- Live grpc-js wrappers for the maintained public helpers.
- Public `executeTool()` plus numeric convenience helpers `add()`, `subtract()`, `multiply()`, and `divide()`.
- Tool discovery helpers `listTools()`, `getToolById()`, `getToolByName()`, and `deleteTool()`.
- Session, API-key, machine, request, and task lifecycle helpers.
- Promise-based connection management with `connect()`, `disconnect()`, `isConnected()`, and `getConnectionStatus()`.
- Shared-fixture conformance coverage in `tests/conformance/`.

## Installation

```bash
npm install
npm run build
```

## Canonical Flow

The canonical end-to-end path for Toolplane is: register a provider, create a session, execute a request, stream or recover results, and drain the machine. The Python SDK has the richest working examples of this flow. The TypeScript examples below follow the same maintained gRPC path at a narrower scope.

## Quick Start

```typescript
import { ToolplaneClient } from './src';

const client = ToolplaneClient.createGRPCClient(
  'localhost',
  9001,
  'grpc-session',
  'grpc-user',
  process.env.TOOLPLANE_API_KEY || 'toolplane-conformance-fixture-key',
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

## Public API Snapshot

### Lifecycle

```typescript
connect(): Promise<void>
disconnect(): Promise<void>
isConnected(): boolean
getConnectionStatus(): ConnectionStatus
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
- Machines: `registerMachine`, `listMachines`, `getMachine`, `unregisterMachine`, `drainMachine`
- Requests: `createRequest`, `getRequest`, `listRequests`, `cancelRequest`
- Tasks: `createTask`, `getTask`, `listTasks`, `cancelTask`

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
npm run advanced
npm test
npm run test:conformance
npm run test:all
```

## Conformance

- The shared fixture runner lives under `tests/conformance/` and loads repository-wide cases from `../../conformance/cases/`.
- The public SDK remains gRPC-only.
- The repository-internal conformance runner uses both gRPC and HTTP adapters so fixture behavior is validated across the maintained server transports.

## Compatibility

- Node.js: 18+
- TypeScript: 5.0+
- Maintained server path: the Go server gRPC control-plane surface
- Public protocol: gRPC with Protocol Buffers
