import { ToolplaneClient } from '../core/toolplane_client';

async function providerRuntimeExample() {
  const serverHost = process.env.TOOLPLANE_SERVER_HOST || 'localhost';
  const apiKey = process.env.TOOLPLANE_API_KEY || 'toolplane-conformance-fixture-key';
  const userId = process.env.TOOLPLANE_USER_ID || 'provider-example-user';

  const client = ToolplaneClient.createGRPCClient(
    serverHost,
    9001,
    '',
    userId,
    apiKey,
  );

  await client.connect();

  const provider = client.providerRuntime({
    pollIntervalMs: 250,
    heartbeatIntervalMs: 30_000,
    sdkVersion: '1.0.0-example',
  });

  const session = await provider.createSession({
    name: 'TypeScript Provider Runtime Example',
    description: 'Maintained provider runtime example for the TypeScript SDK',
    namespace: 'examples',
  });

  await provider.tool(
    {
      sessionId: session.id,
      name: 'echo_tool',
      description: 'Echo the caller message back from the maintained TypeScript provider runtime',
      schema: {
        type: 'object',
        properties: {
          message: { type: 'string' },
        },
        required: ['message'],
      },
      tags: ['provider', 'example', 'typescript'],
    },
    async (input) => ({
      echo: String(input.message ?? ''),
      handled_by: 'typescript-provider-runtime',
    }),
  );

  console.log('=== TypeScript Provider Runtime Example ===');
  console.log(`Session ID: ${session.id}`);
  console.log('Registered tool: echo_tool');
  console.log('Provider runtime is polling for work. Press Ctrl+C to drain and exit.');

  const shutdown = async () => {
    console.log('\nDraining provider runtime...');
    await provider.drain();
    await client.disconnect();
    process.exit(0);
  };

  process.once('SIGINT', () => {
    void shutdown();
  });
  process.once('SIGTERM', () => {
    void shutdown();
  });

  await provider.runForever();
}

if (require.main === module) {
  providerRuntimeExample().catch(async (error) => {
    console.error('Provider runtime example failed:', error);
    process.exit(1);
  });
}