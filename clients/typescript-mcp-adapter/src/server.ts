import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  McpError,
  type JSONRPCRequest,
} from '@modelcontextprotocol/sdk/types.js';

import { NotFoundError, ProtocolError } from 'toolplane-typescript-client';

import type { AdapterOptions } from './config';
import { createAdapterOptionsFromEnv } from './config';
import { debugLog } from './debug';
import { ToolplaneMcpBridge } from './bridge';
import {
  ERROR_INTERNAL,
  ERROR_INVALID_PARAMS,
  ERROR_METHOD_NOT_FOUND,
  RESULT_TYPE_COMPLETE,
  clientSupportsTasks,
  has2026Meta,
  isRecord,
  parseRequestMeta,
} from './protocol';
import { ADAPTER_INSTRUCTIONS, ADAPTER_NAME, ADAPTER_VERSION } from './resources';

/** gRPC status code NOT_FOUND. The client surfaces it as NotFoundError; the
 * ProtocolError path covers the pre-typed-errors shape. */
const GRPC_STATUS_NOT_FOUND = 5;

function isNotFoundError(error: unknown): boolean {
  if (error instanceof NotFoundError) {
    return true;
  }

  if (!(error instanceof ProtocolError)) {
    return false;
  }

  const data = error.data as { code?: unknown } | undefined;
  return isRecord(data) && data.code === GRPC_STATUS_NOT_FOUND;
}

function describeError(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

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

    // SDK 1.29 exposes the fallback handlers as instance properties rather
    // than constructor options. Every method the SDK does not handle itself
    // (everything but initialize) is routed through the stateless dispatcher.
    this.server.fallbackRequestHandler = (request) => this.dispatchRequest(request);
    this.server.fallbackNotificationHandler = async (notification) => {
      debugLog(`ignoring notification ${notification.method}`);
    };
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

  /**
   * Single dispatch point for every method the SDK does not handle itself
   * (initialize is handled by the SDK Server). Requests carrying a 2026
   * protocol version in _meta are served with the stateless surface
   * (resultType discriminators, task handles, tasks/*); requests without it
   * are served with legacy semantics negotiated at initialize, so pre-2026
   * clients keep working against the same adapter process.
   */
  private async dispatchRequest(request: JSONRPCRequest): Promise<Record<string, unknown>> {
    const params = request.params;

    switch (request.method) {
      case 'server/discover': {
        parseRequestMeta(params);
        return this.bridge.discoverPayload();
      }

      case 'tools/list': {
        if (has2026Meta(params)) {
          parseRequestMeta(params);
          return this.bridge.listTools2026();
        }

        return { tools: await this.bridge.listTools() };
      }

      case 'tools/call': {
        return this.handleToolsCall(params);
      }

      case 'tasks/get': {
        const meta = parseRequestMeta(params);
        const taskId = this.requireTaskId(params, 'tasks/get');

        try {
          return await this.bridge.getTask(taskId, meta.lastSeq);
        } catch (error) {
          if (isNotFoundError(error)) {
            throw new McpError(ERROR_INVALID_PARAMS, `task not found: ${taskId}`);
          }
          throw new McpError(ERROR_INTERNAL, `get task ${taskId} failed: ${describeError(error)}`);
        }
      }

      case 'tasks/cancel': {
        parseRequestMeta(params);
        const taskId = this.requireTaskId(params, 'tasks/cancel');

        try {
          await this.bridge.cancelTask(taskId);
        } catch (error) {
          if (isNotFoundError(error)) {
            throw new McpError(ERROR_INVALID_PARAMS, `task not found: ${taskId}`);
          }
          throw new McpError(ERROR_INTERNAL, `cancel task ${taskId} failed: ${describeError(error)}`);
        }

        return { resultType: RESULT_TYPE_COMPLETE };
      }

      case 'tasks/update': {
        throw new McpError(
          ERROR_METHOD_NOT_FOUND,
          'tasks/update is not supported: Toolplane tasks never request client input',
        );
      }

      case 'resources/list': {
        return { resources: await this.bridge.listResources() };
      }

      case 'resources/read': {
        const uri = isRecord(params) && typeof params.uri === 'string' ? params.uri : '';
        if (!uri) {
          throw new McpError(ERROR_INVALID_PARAMS, 'resources/read requires a uri');
        }
        return this.bridge.readResource(uri);
      }

      default: {
        throw new McpError(ERROR_METHOD_NOT_FOUND, `unknown method: ${request.method}`);
      }
    }
  }

  private async handleToolsCall(params: JSONRPCRequest['params']): Promise<Record<string, unknown>> {
    const callParams = isRecord(params) ? params : {};
    const name = typeof callParams.name === 'string' ? callParams.name.trim() : '';

    if (has2026Meta(params)) {
      const meta = parseRequestMeta(params);
      if (!name) {
        throw new McpError(ERROR_INVALID_PARAMS, 'tools/call requires a non-empty tool name');
      }

      const args = isRecord(callParams.arguments) ? callParams.arguments : {};

      let result: Record<string, unknown>;
      if (clientSupportsTasks(meta.capabilities)) {
        result = await this.bridge.createTaskHandle(name, args);
      } else {
        result = await this.bridge.callToolSync(name, args);
      }

      void this.server.sendResourceListChanged().catch(() => undefined);
      return result;
    }

    const result = await this.bridge.callTool(
      name,
      isRecord(callParams.arguments) ? callParams.arguments : {},
    );
    void this.server.sendResourceListChanged().catch(() => undefined);
    return { ...result };
  }

  private requireTaskId(params: JSONRPCRequest['params'], method: string): string {
    const taskId = isRecord(params) && typeof params.taskId === 'string' ? params.taskId.trim() : '';
    if (!taskId) {
      throw new McpError(ERROR_INVALID_PARAMS, `${method} requires taskId`);
    }
    return taskId;
  }
}

export function createToolplaneMcpAdapterServerFromEnv(env: NodeJS.ProcessEnv = process.env): ToolplaneMcpAdapterServer {
  return new ToolplaneMcpAdapterServer(createAdapterOptionsFromEnv(env));
}
