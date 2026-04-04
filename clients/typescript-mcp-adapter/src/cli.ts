import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';

import { debugLog } from './debug';
import { ADAPTER_NAME } from './resources';
import { createToolplaneMcpAdapterServerFromEnv } from './server';

async function main(): Promise<void> {
  debugLog('creating adapter server from environment');
  const server = createToolplaneMcpAdapterServerFromEnv(process.env);
  const transport = new StdioServerTransport();
  let shuttingDown = false;

  const shutdown = async (signal: string): Promise<void> => {
    if (shuttingDown) {
      return;
    }

    shuttingDown = true;
    try {
      await server.close();
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      console.error(`[${ADAPTER_NAME}] failed to shut down after ${signal}: ${message}`);
    }
  };

  process.once('SIGINT', () => {
    void shutdown('SIGINT').finally(() => process.exit(0));
  });

  process.once('SIGTERM', () => {
    void shutdown('SIGTERM').finally(() => process.exit(0));
  });

  debugLog('starting stdio server');
  await server.start(transport);
  debugLog('adapter startup complete');
}

void main().catch((error) => {
  const message = error instanceof Error ? error.stack ?? error.message : String(error);
  console.error(`[${ADAPTER_NAME}] ${message}`);
  process.exitCode = 1;
});