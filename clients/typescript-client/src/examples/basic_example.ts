import { ToolplaneClient } from '../core/toolplane_client';

/**
 * Basic example demonstrating the maintained gRPC control-plane path.
 * This is the recommended first-touch example for the TypeScript SDK.
 */
async function basicExample() {
  const serverHost = process.env.TOOLPLANE_SERVER_HOST || 'localhost';
  const apiKey = process.env.TOOLPLANE_API_KEY || 'toolplane-conformance-fixture-key';
  const userId = process.env.TOOLPLANE_USER_ID || 'example-user';
  console.log('=== TypeScript Toolplane gRPC Example ===\n');

  const grpcClient = ToolplaneClient.createGRPCClient(
    serverHost,
    9001,
    process.env.TOOLPLANE_SESSION_ID || 'grpc-example-session',
    userId,
    apiKey,
  );

  try {
    console.log('🔗 Connecting to gRPC server...');
    await grpcClient.connect();
    console.log('✅ Connected to gRPC server!\n');

    console.log('📁 Session Management:');
    const session = await grpcClient.createSession(
      'TypeScript Example Session',
      'Example session for demonstrating the maintained gRPC path',
      'examples'
    );
    console.log(`   Created session: ${session.name}`);
    console.log(`   Session ID: ${session.id}\n`);

    console.log('🖥️  Machine Registration:');
    const machine = await grpcClient.registerMachine('', '2.0.0', [
      {
        sessionId: session.id,
        name: 'session_status',
        description: 'Report session context for a connected operator or automation client',
        schema: JSON.stringify({
          type: 'object',
          properties: {
            requester: { type: 'string' }
          },
          required: ['requester']
        }),
        config: { version: '1.0', author: 'typescript-example' },
        tags: ['session', 'status', 'typescript']
      },
      {
        sessionId: session.id,
        name: 'incident_brief',
        description: 'Generate an operator-facing summary for an incident or change event',
        schema: JSON.stringify({
          type: 'object',
          properties: {
            service: { type: 'string' },
            severity: { type: 'string' }
          },
          required: ['service', 'severity']
        }),
        config: { version: '1.0', author: 'typescript-example' },
        tags: ['incident', 'summary', 'typescript']
      }
    ]);
    console.log(`   Registered machine: ${machine.id}`);
    console.log(`   SDK Language: ${machine.sdkLanguage}\n`);

    const tools = await grpcClient.listTools();
    console.log('📋 Available Tools:');
    tools.forEach((tool, index) => {
      console.log(`   ${index + 1}. ${tool.name}: ${tool.description}`);
    });
    console.log();

    const sessionDetails = await grpcClient.getSession();
    console.log('📄 Session Details:');
    console.log(`   Name: ${sessionDetails.name}`);
    console.log(`   Namespace: ${sessionDetails.namespace}`);
    console.log(`   Created: ${sessionDetails.createdAt}\n`);

    console.log('Live execution still requires a provider loop to claim requests. See the conformance and integration flows for a runnable end-to-end example.');

  } catch (error) {
    console.error('❌ Error:', error);
  } finally {
    await grpcClient.disconnect();
    console.log('🔌 Disconnected from server');
  }
}

async function main() {
  await basicExample();
}

if (require.main === module) {
  main().catch(console.error);
}