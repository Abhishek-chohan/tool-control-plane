/**
 * Gateway initialize-handshake conformance.
 *
 * Proves the mcp-gateway serves the installed base of initialize-based MCP
 * clients (the 2025-06-18 revision negotiated here) by connecting the
 * OFFICIAL @modelcontextprotocol/sdk Client over StreamableHTTP and running
 * a real provider round trip through it: initialize -> tools/list ->
 * tools/call -> completion.
 *
 * The gateway boots via `go run ./cmd/mcp-gateway` pointed at the same
 * auto-booted gRPC server the adapter tests use.
 */
import assert from 'node:assert/strict';
import net from 'node:net';
import {spawn} from 'node:child_process';
import test from 'node:test';

import {Client} from '@modelcontextprotocol/sdk/client/index.js';
import {StreamableHTTPClientTransport} from '@modelcontextprotocol/sdk/client/streamableHttp.js';

import {GrpcConformanceAdapter} from '../../typescript-client/tests/conformance/adapters/grpc_adapter';
import {startConformanceEnvironment, type ConformanceEnvironment} from '../../typescript-client/tests/conformance/environment';
import {PendingRequestWorker} from './helpers/pending_worker';

function findFreePort(): Promise<number> {
  return new Promise((resolve, reject) => {
    const server = net.createServer();
    server.listen(0, '127.0.0.1', () => {
      const address = server.address();
      if (address && typeof address === 'object') {
        const port = address.port;
        server.close(() => resolve(port));
      } else {
        server.close(() => reject(new Error('no address')));
      }
    });
    server.on('error', reject);
  });
}

function waitForHTTP(url: string, timeoutMs: number): Promise<void> {
  return new Promise((resolve, reject) => {
    const deadline = Date.now() + timeoutMs;
    const attempt = () => {
      fetch(url, {method: 'POST', headers: {'content-type': 'application/json'}, body: '{}'})
        .then(() => resolve())
        .catch(() => {
          if (Date.now() > deadline) {
            reject(new Error(`gateway did not become ready on ${url}`));
            return;
          }
          setTimeout(attempt, 300);
        });
    };
    attempt();
  });
}

test('official SDK client connects through initialize and executes a tool', {timeout: 180_000}, async (t) => {
  const environment: ConformanceEnvironment = await startConformanceEnvironment();
  t.after(async () => {
    await environment.cleanup();
  });

  const grpcPort = Number.parseInt(process.env.TOOLPLANE_CONFORMANCE_GRPC_PORT ?? '', 10);
  assert.ok(Number.isFinite(grpcPort), 'conformance environment did not expose a gRPC port');

  // Boot the gateway in front of the gRPC server.
  const gatewayPort = await findFreePort();
  const gateway = spawn(
    'go',
    [
      'run',
      './cmd/mcp-gateway',
      '--backend',
      `localhost:${grpcPort}`,
      '--listen',
      `:${gatewayPort}`,
    ],
    {cwd: process.env.TOOLPLANE_REPO_SERVER_DIR ?? '../../server', stdio: 'ignore'},
  );
  t.after(() => {
    gateway.kill('SIGTERM');
  });

  const gatewayURL = `http://127.0.0.1:${gatewayPort}/mcp`;
  await waitForHTTP(gatewayURL, 60_000);

  // The official SDK client negotiates initialize over StreamableHTTP.
  const apiKey = process.env.TOOLPLANE_CONFORMANCE_API_KEY ?? 'toolplane-conformance-fixture-key';
  const transport = new StreamableHTTPClientTransport(new URL(gatewayURL), {
    requestInit: {
      headers: {
        // The backend runs fixed-token auth in conformance mode; real
        // deployments pass their session key the same way.
        authorization: `Bearer ${apiKey}`,
      },
    },
  });
  const client = new Client({name: 'gateway-initialize-test', version: '1.0.0'}, {capabilities: {}});
  t.after(async () => {
    await client.close();
  });
  await client.connect(transport);

  // tools/list through the compatibility path: plain legacy shape.
  const toolsResult = await client.listTools();
  assert.ok(Array.isArray(toolsResult.tools), 'legacy tools/list missing tools array');

  // Provider side: register an echo tool and serve requests.
  const adapter = new GrpcConformanceAdapter(
    process.env.TOOLPLANE_CONFORMANCE_GRPC_HOST ?? 'localhost',
    grpcPort,
    process.env.TOOLPLANE_CONFORMANCE_USER_ID ?? 'conformance-user',
    process.env.TOOLPLANE_CONFORMANCE_API_KEY ?? 'toolplane-conformance-fixture-key',
  );
  await adapter.connect();
  t.after(async () => {
    await adapter.close();
  });

  // The gateway provisions one session per API key under its own user
  // ("mcp-gateway"); a legacy client has no way to name a session, so the
  // provider registers into the provisioned one. Find it via the first
  // tools/list (which triggers provisioning).
  const provisioned = await adapter.listUserSessions({user_id: 'mcp-gateway'});
  const sessions = (provisioned.sessions as Array<Record<string, unknown>>) ?? [];
  assert.ok(sessions.length >= 1, 'gateway did not provision a session');
  const sessionId = String(sessions[sessions.length - 1].id);

  await adapter.registerUnaryEchoTool(
    sessionId,
    'gateway_initialize_tool',
    'gateway initialize conformance tool',
  );
  adapter.startProviderRuntime(sessionId);

  const toolsAfter = await client.listTools();
  const namesAfter = toolsAfter.tools.map((tool) => tool.name);
  assert.ok(
    namesAfter.includes('gateway_initialize_tool'),
    `tools/list did not surface the registered tool; got ${JSON.stringify(namesAfter)}`,
  );

  // tools/call end to end: the gateway creates the request, the provider
  // claims and completes it, the sync poll aggregates the result.
  const worker = new PendingRequestWorker(adapter, sessionId);
  worker.start();
  t.after(async () => {
    await worker.stop();
  });

  const callResult = await client.callTool({
    name: 'gateway_initialize_tool',
    arguments: {message: 'official-sdk'},
  });
  assert.notEqual(callResult.isError, true, `tools/call reported isError: ${JSON.stringify(callResult)}`);

  const textBlocks = (callResult.content ?? []).filter(
    (block): block is {type: 'text'; text: string} =>
      typeof block === 'object' && block !== null && (block as {type?: unknown}).type === 'text',
  );
  const joined = textBlocks.map((block) => block.text).join('\n');
  assert.ok(
    joined.includes('official-sdk'),
    `tools/call content missing echoed payload: ${joined}`,
  );
});
