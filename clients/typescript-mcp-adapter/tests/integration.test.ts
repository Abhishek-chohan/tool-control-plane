import assert from 'node:assert/strict';
import { spawn, type ChildProcessWithoutNullStreams } from 'node:child_process';
import type { Readable } from 'node:stream';
import path from 'node:path';
import { after, afterEach, before, beforeEach, describe, it } from 'node:test';

import { Client } from '@modelcontextprotocol/sdk/client';
import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js';

import type { ConformanceEnvironment } from '../../typescript-client/tests/conformance/environment';
import { startConformanceEnvironment } from '../../typescript-client/tests/conformance/environment';
import { GrpcConformanceAdapter } from '../../typescript-client/tests/conformance/adapters/grpc_adapter';
import { PendingRequestWorker } from './helpers/pending_worker';

import {
  CHUNKS_META_KEY,
  META_CLIENT_CAPABILITIES,
  META_PROTOCOL_VERSION,
  PROTOCOL_VERSION,
  TASKS_EXTENSION_ID,
  TASK_ID_DATA_KEY,
} from '../src/protocol';
import {
  CONCEPT_MAP_RESOURCE_URI,
  SESSION_RESOURCE_URI,
} from '../src/resources';

const UNARY_TOOL_NAME = 'mcp_adapter_echo';
const STREAM_TOOL_NAME = 'mcp_adapter_stream';

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

function textContent(result: Awaited<ReturnType<Client['readResource']>>): string {
  const first = result.contents[0];
  assert.ok(first, 'expected at least one resource payload');
  assert.ok('text' in first, 'expected a text resource payload');
  return first.text;
}

function adapterCliPath(): string {
  return path.resolve(process.cwd(), 'dist/cli.js');
}

function isNamedTool(value: unknown): value is { name: string; description?: string; inputSchema: { type: string } } {
  return typeof value === 'object' && value !== null && 'name' in value;
}

function isNamedResource(value: unknown): value is { uri: string } {
  return typeof value === 'object' && value !== null && 'uri' in value;
}

function isTextBlock(value: unknown): value is { type: 'text'; text: string } {
  return typeof value === 'object' && value !== null && (value as { type?: unknown }).type === 'text';
}

function isResourceLinkBlock(value: unknown): value is { type: 'resource_link'; uri: string } {
  return typeof value === 'object' && value !== null && (value as { type?: unknown }).type === 'resource_link';
}

interface AdapterClientHarness {
  client: Client;
  transport: StdioClientTransport;
  stderr: () => string;
}

interface NormalizedCallToolResult {
  content: unknown[];
  structuredContent?: Record<string, unknown>;
  isError?: boolean;
}

async function startAdapterClient(sessionId: string): Promise<AdapterClientHarness> {
  const stderrChunks: string[] = [];
  const transport = new StdioClientTransport({
    command: process.execPath,
    args: [adapterCliPath()],
    cwd: process.cwd(),
    env: {
      ...process.env,
      TOOLPLANE_MCP_GRPC_HOST: process.env.TOOLPLANE_CONFORMANCE_GRPC_HOST ?? 'localhost',
      TOOLPLANE_MCP_GRPC_PORT: process.env.TOOLPLANE_CONFORMANCE_GRPC_PORT ?? '50051',
      TOOLPLANE_MCP_USER_ID: process.env.TOOLPLANE_CONFORMANCE_USER_ID ?? 'conformance-user',
      TOOLPLANE_MCP_API_KEY: process.env.TOOLPLANE_CONFORMANCE_API_KEY ?? 'toolplane-conformance-fixture-key',
      TOOLPLANE_MCP_SESSION_ID: sessionId,
      TOOLPLANE_MCP_REQUEST_RESOURCE_LIMIT: '10',
    },
    stderr: 'pipe',
  });

  const stderrStream = transport.stderr as Readable | null;
  if (stderrStream) {
    stderrStream.on('data', (chunk: Buffer | string) => {
      stderrChunks.push(String(chunk));
    });
  }

  const client = new Client(
    {
      name: 'toolplane-mcp-adapter-tests',
      version: '1.0.0',
    },
    {
      capabilities: {},
    },
  );

  try {
    await client.connect(transport);
  } catch (error) {
    const stderr = stderrChunks.join('');
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`failed to connect to adapter: ${message}\n${stderr}`);
  }

  return {
    client,
    transport,
    stderr: () => stderrChunks.join(''),
  };
}

describe('TypeScript MCP adapter integration', () => {
  let environment: ConformanceEnvironment;
  let provider: GrpcConformanceAdapter;
  let pendingWorker: PendingRequestWorker;
  let adapterClient: AdapterClientHarness;
  let sessionId = '';

  before(async () => {
    environment = await startConformanceEnvironment();
  });

  after(async () => {
    await environment.cleanup();
  });

  beforeEach(async () => {
    const host = process.env.TOOLPLANE_CONFORMANCE_GRPC_HOST ?? 'localhost';
    const port = Number.parseInt(process.env.TOOLPLANE_CONFORMANCE_GRPC_PORT ?? '50051', 10);
    const userId = process.env.TOOLPLANE_CONFORMANCE_USER_ID ?? 'conformance-user';
    const apiKey = process.env.TOOLPLANE_CONFORMANCE_API_KEY ?? 'toolplane-conformance-fixture-key';

    provider = new GrpcConformanceAdapter(host, port, userId, apiKey);
    await provider.connect();

    sessionId = await provider.createSession({
      name: 'MCP adapter integration session',
      description: 'session bound to the optional MCP adapter',
      namespace: 'adapter-tests',
    });

    await provider.registerUnaryEchoTool(sessionId, UNARY_TOOL_NAME, 'Echo a message through the adapter');
    await provider.registerStreamTool(sessionId, STREAM_TOOL_NAME, 'Emit streaming chunks through the adapter');
    await provider.startProviderRuntime(sessionId);

    pendingWorker = new PendingRequestWorker(provider, sessionId);
    pendingWorker.start();

    adapterClient = await startAdapterClient(sessionId);
  });

  afterEach(async () => {
    await adapterClient.transport.close().catch(() => undefined);
    await pendingWorker.stop().catch(() => undefined);
    await provider.close().catch((error) => {
      const stderr = adapterClient.stderr();
      throw new Error(`provider cleanup failed: ${String(error)}\n${stderr}`);
    });
  });

  it('lists translated tools and static adapter resources', async () => {
    let tools: unknown[] = [];
    let resources: unknown[] = [];
    let sessionText = '';
    let conceptText = '';

    try {
      const toolResult = await adapterClient.client.listTools();
      tools = toolResult.tools;

      const unaryTool = tools
        .filter(isNamedTool)
        .find((tool) => tool.name === UNARY_TOOL_NAME);
      assert.ok(unaryTool, 'expected the unary echo tool to be visible through MCP');
      assert.equal(unaryTool.description, 'Echo a message through the adapter');
      assert.equal(unaryTool.inputSchema.type, 'object');

      const resourceResult = await adapterClient.client.listResources();
      resources = resourceResult.resources;
      assert.ok(resources.some((resource: unknown) => isNamedResource(resource) && resource.uri === SESSION_RESOURCE_URI));
      assert.ok(resources.some((resource: unknown) => isNamedResource(resource) && resource.uri === CONCEPT_MAP_RESOURCE_URI));

      const sessionResource = await adapterClient.client.readResource({ uri: SESSION_RESOURCE_URI });
      sessionText = textContent(sessionResource);
      const sessionPayload = JSON.parse(sessionText) as Record<string, unknown>;
      assert.equal(sessionPayload.id, sessionId);
      assert.equal(sessionPayload.nativeProtocol, 'toolplane');
      assert.equal(sessionPayload.adaptedProtocol, 'mcp');

      const conceptMap = await adapterClient.client.readResource({ uri: CONCEPT_MAP_RESOURCE_URI });
      conceptText = textContent(conceptMap);
      assert.match(conceptText, /aggregated into one synchronous MCP tool result/i);
    } catch (error) {
      console.error('DEBUG_TOOLS', JSON.stringify(tools, null, 2));
      console.error('DEBUG_RESOURCES', JSON.stringify(resources, null, 2));
      console.error('DEBUG_SESSION', sessionText);
      console.error('DEBUG_CONCEPT', conceptText);
      throw error;
    }
  });

  it('returns unary request output and exposes a request resource', async () => {
    const result = await adapterClient.client.callTool({
      name: UNARY_TOOL_NAME,
      arguments: {
        message: 'hello from MCP',
      },
    }) as NormalizedCallToolResult;

    assert.equal(result.isError, undefined);
    const structuredContent = result.structuredContent as Record<string, unknown>;
    assert.deepEqual(structuredContent.result, { echo: 'hello from MCP' });
    assert.equal(structuredContent.toolName, UNARY_TOOL_NAME);

    const requestLink = result.content.find((item: unknown) => isResourceLinkBlock(item));
    assert.ok(requestLink, 'expected a request resource link in the tool result');

    const requestResource = await adapterClient.client.readResource({ uri: requestLink.uri });
    const payload = JSON.parse(textContent(requestResource)) as Record<string, unknown>;
    assert.equal(payload.requestId, structuredContent.requestId);
    assert.equal(payload.status, 'done');
    assert.deepEqual(payload.result, { echo: 'hello from MCP' });

    const { resources } = await adapterClient.client.listResources();
    assert.ok(resources.some((resource: unknown) => isNamedResource(resource) && resource.uri === requestLink.uri));
  });

  it('aggregates native stream chunks into one MCP tool result', async () => {
    const result = await adapterClient.client.callTool({
      name: STREAM_TOOL_NAME,
      arguments: {
        prefix: 'chunk',
        count: 3,
      },
    }) as NormalizedCallToolResult;

    assert.equal(result.isError, undefined);
    const structuredContent = result.structuredContent as Record<string, unknown>;
    assert.deepEqual(structuredContent.streamResults, ['chunk-1', 'chunk-2', 'chunk-3']);
    assert.deepEqual(structuredContent.result, ['chunk-1', 'chunk-2', 'chunk-3']);

    const textBlock = result.content.find((item: unknown) => isTextBlock(item));
    assert.ok(textBlock, 'expected a textual summary block');
    assert.match(textBlock.text, /Aggregated 3 native stream chunk\(s\) into this MCP tool response/i);
  });
});

/* ------------------------------------------------------------------ */
/* MCP 2026-07-28 stateless-core + Tasks extension surface            */
/* ------------------------------------------------------------------ */

/** Error surfaced by the raw client when the server returns a JSON-RPC error. */
class RawMcpError extends Error {
  constructor(
    public readonly code: number,
    message: string,
    public readonly data?: unknown,
  ) {
    super(message);
    this.name = 'RawMcpError';
  }
}

interface PendingRawRequest {
  resolve: (result: unknown) => void;
  reject: (error: RawMcpError) => void;
  timer: NodeJS.Timeout;
}

/**
 * Minimal MCP 2026-07-28 stateless client over stdio. It speaks raw JSON-RPC
 * with no initialize handshake, attaching the required per-request `_meta`
 * (protocolVersion + clientCapabilities) exactly as the 2026 spec demands.
 * This exercises the adapter the way a modern stateless MCP client would.
 */
class RawStatelessMcpClient {
  private readonly child: ChildProcessWithoutNullStreams;

  private stdoutBuffer = '';

  private readonly stderrChunks: string[] = [];

  private readonly pending = new Map<number, PendingRawRequest>();

  private nextId = 1;

  constructor(sessionId: string) {
    this.child = spawn(process.execPath, [adapterCliPath()], {
      cwd: process.cwd(),
      env: {
        ...process.env,
        TOOLPLANE_MCP_GRPC_HOST: process.env.TOOLPLANE_CONFORMANCE_GRPC_HOST ?? 'localhost',
        TOOLPLANE_MCP_GRPC_PORT: process.env.TOOLPLANE_CONFORMANCE_GRPC_PORT ?? '50051',
        TOOLPLANE_MCP_USER_ID: process.env.TOOLPLANE_CONFORMANCE_USER_ID ?? 'conformance-user',
        TOOLPLANE_MCP_API_KEY: process.env.TOOLPLANE_CONFORMANCE_API_KEY ?? 'toolplane-conformance-fixture-key',
        TOOLPLANE_MCP_SESSION_ID: sessionId,
        TOOLPLANE_MCP_REQUEST_RESOURCE_LIMIT: '10',
      },
      stdio: ['pipe', 'pipe', 'pipe'],
    });

    this.child.stdout.on('data', (chunk: Buffer) => this.onStdout(chunk));
    this.child.stderr.on('data', (chunk: Buffer) => {
      this.stderrChunks.push(chunk.toString());
    });
  }

  stderr(): string {
    return this.stderrChunks.join('');
  }

  /** Builds a 2026 _meta object, optionally advertising the Tasks extension. */
  static meta(withTasks: boolean): Record<string, unknown> {
    const capabilities: Record<string, unknown> = withTasks
      ? { extensions: { [TASKS_EXTENSION_ID]: {} } }
      : {};

    return {
      [META_PROTOCOL_VERSION]: PROTOCOL_VERSION,
      [META_CLIENT_CAPABILITIES]: capabilities,
    };
  }

  /** Sends a request and resolves with the result or rejects with RawMcpError. */
  request(method: string, params: Record<string, unknown> = {}): Promise<unknown> {
    const id = this.nextId;
    this.nextId += 1;

    return new Promise((resolve, reject) => {
      const timer = setTimeout(() => {
        this.pending.delete(id);
        reject(new RawMcpError(-1, `timed out waiting for response to ${method}`));
      }, 30_000);

      this.pending.set(id, { resolve, reject, timer });
      this.child.stdin.write(`${JSON.stringify({ jsonrpc: '2.0', id, method, params })}\n`);
    });
  }

  async close(): Promise<void> {
    for (const [, pendingRequest] of this.pending) {
      clearTimeout(pendingRequest.timer);
      pendingRequest.reject(new RawMcpError(-1, 'client closed'));
    }
    this.pending.clear();

    this.child.kill('SIGTERM');
    await new Promise<void>((resolve) => {
      const timer = setTimeout(() => {
        this.child.kill('SIGKILL');
        resolve();
      }, 3000);
      this.child.once('exit', () => {
        clearTimeout(timer);
        resolve();
      });
    });
  }

  private onStdout(chunk: Buffer): void {
    this.stdoutBuffer += chunk.toString();

    let newlineIndex = this.stdoutBuffer.indexOf('\n');
    while (newlineIndex >= 0) {
      const line = this.stdoutBuffer.slice(0, newlineIndex).trim();
      this.stdoutBuffer = this.stdoutBuffer.slice(newlineIndex + 1);
      newlineIndex = this.stdoutBuffer.indexOf('\n');

      if (!line) {
        continue;
      }

      let message: { id?: number; result?: unknown; error?: { code: number; message: string; data?: unknown } };
      try {
        message = JSON.parse(line);
      } catch {
        continue;
      }

      if (message.id === undefined) {
        continue; // notification or malformed; the stateless surface does not rely on them
      }

      const pendingRequest = this.pending.get(message.id);
      if (!pendingRequest) {
        continue;
      }

      this.pending.delete(message.id);
      clearTimeout(pendingRequest.timer);

      if (message.error) {
        pendingRequest.reject(new RawMcpError(message.error.code, message.error.message, message.error.data));
      } else {
        pendingRequest.resolve(message.result);
      }
    }
  }
}

function isRecordLike(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

describe('TypeScript MCP adapter 2026 stateless surface', () => {
  let environment: ConformanceEnvironment;
  let provider: GrpcConformanceAdapter;
  let pendingWorker: PendingRequestWorker;
  let rawClient: RawStatelessMcpClient;
  let sessionId = '';

  before(async () => {
    environment = await startConformanceEnvironment();
  });

  after(async () => {
    await environment.cleanup();
  });

  beforeEach(async () => {
    const host = process.env.TOOLPLANE_CONFORMANCE_GRPC_HOST ?? 'localhost';
    const port = Number.parseInt(process.env.TOOLPLANE_CONFORMANCE_GRPC_PORT ?? '50051', 10);
    const userId = process.env.TOOLPLANE_CONFORMANCE_USER_ID ?? 'conformance-user';
    const apiKey = process.env.TOOLPLANE_CONFORMANCE_API_KEY ?? 'toolplane-conformance-fixture-key';

    provider = new GrpcConformanceAdapter(host, port, userId, apiKey);
    await provider.connect();

    sessionId = await provider.createSession({
      name: 'MCP adapter 2026 surface session',
      description: 'session bound to the adapter stateless surface',
      namespace: 'adapter-2026-tests',
    });

    await provider.registerUnaryEchoTool(sessionId, UNARY_TOOL_NAME, 'Echo a message through the adapter');
    await provider.registerStreamTool(sessionId, STREAM_TOOL_NAME, 'Emit streaming chunks through the adapter');
    await provider.startProviderRuntime(sessionId);

    pendingWorker = new PendingRequestWorker(provider, sessionId);
    pendingWorker.start();

    rawClient = new RawStatelessMcpClient(sessionId);
  });

  afterEach(async () => {
    await rawClient.close();
    await pendingWorker.stop().catch(() => undefined);
    await provider.close().catch(() => undefined);
  });

  it('server/discover advertises 2026-07-28 and the Tasks extension', async () => {
    const result = (await rawClient.request('server/discover', {
      _meta: RawStatelessMcpClient.meta(false),
    })) as Record<string, unknown>;

    assert.equal(result.resultType, 'complete');
    assert.deepEqual(result.supportedVersions, [PROTOCOL_VERSION]);

    const capabilities = result.capabilities as Record<string, unknown>;
    assert.ok(isRecordLike(capabilities), 'discover must advertise capabilities');
    assert.ok(isRecordLike(capabilities.tools), 'discover must advertise the tools capability');

    const extensions = capabilities.extensions as Record<string, unknown>;
    assert.ok(isRecordLike(extensions), 'discover must advertise extensions');
    assert.ok(TASKS_EXTENSION_ID in extensions, 'discover must advertise the Tasks extension');
    assert.equal(typeof result.instructions, 'string');
  });

  it('tools/list returns resultType-bearing entries to a 2026 client', async () => {
    const result = (await rawClient.request('tools/list', {
      _meta: RawStatelessMcpClient.meta(false),
    })) as Record<string, unknown>;

    assert.equal(result.resultType, 'complete');
    const tools = result.tools as Array<Record<string, unknown>>;
    assert.ok(Array.isArray(tools), 'tools/list must return a tools array');

    const streamTool = tools.find((tool) => tool.name === STREAM_TOOL_NAME);
    assert.ok(streamTool, 'expected the stream tool in the 2026 tools/list result');
    assert.ok(isRecordLike(streamTool.inputSchema), 'tool must carry an inputSchema');
    assert.equal((streamTool.inputSchema as Record<string, unknown>).type, 'object');
  });

  it('tools/call returns a task handle to a Tasks-capable client', async () => {
    const handle = (await rawClient.request('tools/call', {
      _meta: RawStatelessMcpClient.meta(true),
      name: STREAM_TOOL_NAME,
      arguments: { prefix: 'task-chunk', count: 4 },
    })) as Record<string, unknown>;

    try {
      assert.equal(handle.resultType, 'task');
      assert.ok(typeof handle.taskId === 'string' && handle.taskId.length > 0, 'handle must carry a taskId');
      assert.equal(handle.status, 'working');
      assert.equal(handle.ttlMs, null);
      assert.equal(handle.pollIntervalMs, 500);
    } finally {
      await rawClient.request('tasks/cancel', {
        _meta: RawStatelessMcpClient.meta(true),
        taskId: handle.taskId,
      }).catch(() => undefined);
    }
  });

  it('tasks/get surfaces the chunk window, honors the cursor, and reaches completed', async () => {
    const handle = (await rawClient.request('tools/call', {
      _meta: RawStatelessMcpClient.meta(true),
      name: STREAM_TOOL_NAME,
      arguments: { prefix: 'task-chunk', count: 4 },
    })) as Record<string, unknown>;

    const taskId = String(handle.taskId);

    // Poll until the working task exposes a chunk window.
    let window: Record<string, unknown> | undefined;
    const windowDeadline = Date.now() + 15_000;
    while (!window && Date.now() < windowDeadline) {
      const poll = (await rawClient.request('tasks/get', {
        _meta: RawStatelessMcpClient.meta(true),
        taskId,
      })) as Record<string, unknown>;

      const meta = poll._meta as Record<string, unknown> | undefined;
      const chunksMeta = isRecordLike(meta) ? meta[CHUNKS_META_KEY] : undefined;
      if (isRecordLike(chunksMeta) && Array.isArray(chunksMeta.chunks) && chunksMeta.chunks.length > 0) {
        window = chunksMeta;
        break;
      }

      if (poll.status !== 'working') {
        break;
      }
      await sleep(100);
    }

    assert.ok(window, `expected a chunk window while the task was working\n${rawClient.stderr()}`);
    const seenChunks = new Set((window.chunks as unknown[]).map((chunk) => JSON.stringify(chunk)));
    const nextSeq = window.nextSeq as number;

    // Cursor replay: asking from nextSeq must not redeliver already-seen chunks.
    const cursorPoll = (await rawClient.request('tasks/get', {
      _meta: { ...RawStatelessMcpClient.meta(true), 'dev.toolplane/last_seq': nextSeq },
      taskId,
    })) as Record<string, unknown>;

    const cursorMeta = cursorPoll._meta as Record<string, unknown> | undefined;
    const cursorWindow = isRecordLike(cursorMeta) ? cursorMeta[CHUNKS_META_KEY] : undefined;
    if (isRecordLike(cursorWindow) && Array.isArray(cursorWindow.chunks)) {
      for (const chunk of cursorWindow.chunks as unknown[]) {
        assert.ok(!seenChunks.has(JSON.stringify(chunk)), `chunk ${JSON.stringify(chunk)} was redelivered after the cursor`);
      }
    }

    // Poll to the terminal state and verify the aggregated result.
    let finalTask: Record<string, unknown> | undefined;
    const terminalDeadline = Date.now() + 15_000;
    while (Date.now() < terminalDeadline) {
      const poll = (await rawClient.request('tasks/get', {
        _meta: RawStatelessMcpClient.meta(true),
        taskId,
      })) as Record<string, unknown>;

      if (poll.status === 'completed' || poll.status === 'failed' || poll.status === 'cancelled') {
        finalTask = poll;
        break;
      }
      await sleep(100);
    }

    assert.ok(finalTask, `expected the task to reach a terminal status\n${rawClient.stderr()}`);
    assert.equal(finalTask.status, 'completed');
    assert.equal(finalTask.resultType, 'complete');

    const taskResult = finalTask.result as Record<string, unknown>;
    assert.ok(isRecordLike(taskResult), 'completed task must carry a result');
    assert.equal(taskResult.resultType, 'complete');

    const structured = taskResult.structuredContent as Record<string, unknown>;
    assert.ok(isRecordLike(structured), 'completed task result must carry structuredContent');
    assert.deepEqual(structured.result, ['task-chunk-1', 'task-chunk-2', 'task-chunk-3', 'task-chunk-4']);
  });

  it('tasks/cancel cancels a working task', async () => {
    const handle = (await rawClient.request('tools/call', {
      _meta: RawStatelessMcpClient.meta(true),
      name: STREAM_TOOL_NAME,
      arguments: { prefix: 'cancel-chunk', count: 12 },
    })) as Record<string, unknown>;

    const taskId = String(handle.taskId);

    // Wait until the task is actually executing (has emitted a chunk) so the
    // cancel lands while it is working rather than while still pending.
    const runningDeadline = Date.now() + 15_000;
    let sawChunk = false;
    while (!sawChunk && Date.now() < runningDeadline) {
      const poll = (await rawClient.request('tasks/get', {
        _meta: RawStatelessMcpClient.meta(true),
        taskId,
      })) as Record<string, unknown>;

      const meta = poll._meta as Record<string, unknown> | undefined;
      const chunksMeta = isRecordLike(meta) ? meta[CHUNKS_META_KEY] : undefined;
      if (isRecordLike(chunksMeta) && Array.isArray(chunksMeta.chunks) && chunksMeta.chunks.length > 0) {
        sawChunk = true;
        break;
      }
      await sleep(100);
    }
    assert.ok(sawChunk, `expected the task to start streaming before cancelling\n${rawClient.stderr()}`);

    const ack = (await rawClient.request('tasks/cancel', {
      _meta: RawStatelessMcpClient.meta(true),
      taskId,
    })) as Record<string, unknown>;
    assert.equal(ack.resultType, 'complete');

    // Poll to the terminal state; the task must end cancelled.
    let finalStatus = '';
    const terminalDeadline = Date.now() + 15_000;
    while (Date.now() < terminalDeadline) {
      const poll = (await rawClient.request('tasks/get', {
        _meta: RawStatelessMcpClient.meta(true),
        taskId,
      })) as Record<string, unknown>;

      finalStatus = String(poll.status);
      if (finalStatus === 'completed' || finalStatus === 'failed' || finalStatus === 'cancelled') {
        break;
      }
      await sleep(100);
    }

    assert.equal(finalStatus, 'cancelled', `expected cancelled, got ${finalStatus}\n${rawClient.stderr()}`);
  });

  it('serves a 2026 client without the Tasks extension synchronously', async () => {
    const result = (await rawClient.request('tools/call', {
      _meta: RawStatelessMcpClient.meta(false),
      name: UNARY_TOOL_NAME,
      arguments: { message: 'sync-2026' },
    })) as Record<string, unknown>;

    assert.equal(result.resultType, 'complete');
    const meta = result._meta as Record<string, unknown>;
    assert.ok(isRecordLike(meta), 'sync 2026 result must carry _meta');
    assert.ok(typeof meta[TASK_ID_DATA_KEY] === 'string' && (meta[TASK_ID_DATA_KEY] as string).length > 0);

    const structured = result.structuredContent as Record<string, unknown>;
    assert.ok(isRecordLike(structured), 'sync result must carry structuredContent');
    assert.deepEqual(structured.result, { echo: 'sync-2026' });
  });

  it('rejects an unsupported protocol version with -32022', async () => {
    await assert.rejects(
      rawClient.request('tools/list', {
        _meta: {
          [META_PROTOCOL_VERSION]: '1999-01-01',
          [META_CLIENT_CAPABILITIES]: {},
        },
      }),
      (error: unknown) => {
        assert.ok(error instanceof RawMcpError, 'expected a RawMcpError');
        assert.equal(error.code, -32022);
        const data = error.data as Record<string, unknown>;
        assert.ok(isRecordLike(data), 'unsupported-version error must carry data');
        assert.deepEqual(data.supported, [PROTOCOL_VERSION]);
        assert.equal(data.requested, '1999-01-01');
        return true;
      },
    );
  });

  it('rejects 2026-only methods that omit _meta with -32600', async () => {
    for (const method of ['server/discover', 'tasks/get', 'tasks/cancel']) {
      await assert.rejects(
        rawClient.request(method, method === 'server/discover' ? {} : { taskId: 'does-not-matter' }),
        (error: unknown) => {
          assert.ok(error instanceof RawMcpError, `expected a RawMcpError for ${method}`);
          assert.equal(error.code, -32600, `expected -32600 for ${method}`);
          return true;
        },
      );
    }
  });

  it('rejects an unknown task and refuses tasks/update', async () => {
    await assert.rejects(
      rawClient.request('tasks/get', {
        _meta: RawStatelessMcpClient.meta(true),
        taskId: 'no-such-task',
      }),
      (error: unknown) => {
        assert.ok(error instanceof RawMcpError, 'expected a RawMcpError');
        assert.equal(error.code, -32602);
        return true;
      },
    );

    await assert.rejects(
      rawClient.request('tasks/update', {
        _meta: RawStatelessMcpClient.meta(true),
        taskId: 'no-such-task',
        inputResponses: {},
      }),
      (error: unknown) => {
        assert.ok(error instanceof RawMcpError, 'expected a RawMcpError');
        assert.equal(error.code, -32601);
        return true;
      },
    );
  });
});