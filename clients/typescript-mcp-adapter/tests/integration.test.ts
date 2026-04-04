import assert from 'node:assert/strict';
import type { Readable } from 'node:stream';
import path from 'node:path';
import { after, afterEach, before, beforeEach, describe, it } from 'node:test';

import { Client } from '@modelcontextprotocol/sdk/client';
import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js';

import type { ConformanceEnvironment } from '../../typescript-client/tests/conformance/environment';
import { startConformanceEnvironment } from '../../typescript-client/tests/conformance/environment';
import { GrpcConformanceAdapter } from '../../typescript-client/tests/conformance/adapters/grpc_adapter';

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

class PendingRequestWorker {
  private stopped = false;

  private loopPromise?: Promise<void>;

  private readonly inFlight = new Set<string>();

  constructor(
    private readonly provider: GrpcConformanceAdapter,
    private readonly sessionId: string,
  ) {}

  start(): void {
    if (!this.loopPromise) {
      this.loopPromise = this.loop();
    }
  }

  async stop(): Promise<void> {
    this.stopped = true;

    if (this.loopPromise) {
      await this.loopPromise;
    }

    while (this.inFlight.size > 0) {
      await sleep(25);
    }
  }

  private async loop(): Promise<void> {
    while (!this.stopped) {
      const pendingRequests = await this.provider.listRequests(this.sessionId, {
        list_status: 'pending',
        limit: 20,
      });

      for (const request of pendingRequests) {
        const requestId = String(request.id ?? '');
        if (!requestId || this.inFlight.has(requestId)) {
          continue;
        }

        this.inFlight.add(requestId);
        void this.provider
          .startRequestProcessing(this.sessionId, requestId)
          .finally(() => {
            this.inFlight.delete(requestId);
          });
      }

      await sleep(50);
    }
  }
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