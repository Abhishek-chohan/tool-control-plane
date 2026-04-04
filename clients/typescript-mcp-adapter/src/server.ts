import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolRequestSchema,
  ListResourcesRequestSchema,
  ListToolsRequestSchema,
  ReadResourceRequestSchema,
} from '@modelcontextprotocol/sdk/types.js';

import type { AdapterOptions } from './config';
import { createAdapterOptionsFromEnv } from './config';
import { debugLog } from './debug';
import { ToolplaneMcpBridge } from './bridge';
import { ADAPTER_INSTRUCTIONS, ADAPTER_NAME, ADAPTER_VERSION } from './resources';

export class ToolplaneMcpAdapterServer {
  private readonly bridge: ToolplaneMcpBridge;

  private readonly server: Server;

  private transport?: StdioServerTransport;

  constructor(options: AdapterOptions) {
    this.bridge = new ToolplaneMcpBridge(options);
    this.server = new Server(
      {
        name: ADAPTER_NAME,
        version: ADAPTER_VERSION,
      },
      {
        capabilities: {
          resources: {
            listChanged: true,
          },
          tools: {},
        },
        instructions: ADAPTER_INSTRUCTIONS,
      },
    );

    this.registerHandlers();
  }

  async start(transport: StdioServerTransport = new StdioServerTransport()): Promise<void> {
    debugLog('starting adapter bridge bootstrap');
    await this.bridge.connect();
    debugLog('bridge bootstrap complete; attaching stdio transport');
    this.transport = transport;
    await this.server.connect(transport);
    debugLog('stdio transport attached');
  }

  async close(): Promise<void> {
    await Promise.allSettled([
      this.transport?.close(),
      this.bridge.close(),
    ]);
  }

  private registerHandlers(): void {
    this.server.setRequestHandler(ListToolsRequestSchema, async () => ({
      tools: await this.bridge.listTools(),
    }));

    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      const result = await this.bridge.callTool(request.params.name, request.params.arguments ?? {});
      void this.server.sendResourceListChanged().catch(() => undefined);
      return result;
    });

    this.server.setRequestHandler(ListResourcesRequestSchema, async () => ({
      resources: await this.bridge.listResources(),
    }));

    this.server.setRequestHandler(ReadResourceRequestSchema, async (request) => {
      return this.bridge.readResource(request.params.uri);
    });
  }
}

export function createToolplaneMcpAdapterServerFromEnv(env: NodeJS.ProcessEnv = process.env): ToolplaneMcpAdapterServer {
  return new ToolplaneMcpAdapterServer(createAdapterOptionsFromEnv(env));
}