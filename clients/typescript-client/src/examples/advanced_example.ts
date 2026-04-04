import { ToolplaneClient } from '../core/toolplane_client';
import { ClientProtocol } from '../interfaces';
import { ToolplaneError, ConnectionError, ProtocolError } from '../errors';

const serverHost = process.env.TOOLPLANE_SERVER_HOST || 'localhost';
const grpcApiKey = process.env.TOOLPLANE_API_KEY || 'toolplane-conformance-fixture-key';

/**
 * Advanced example demonstrating maintained gRPC control-plane usage,
 * transport-boundary errors, and concurrent lifecycle reads.
 *
 * Execution order:
 *   1. testGRPCProtocol          — maintained control-plane path
 *   2. testBoundaryErrors        — connection and protocol-boundary demo
 *   3. testConcurrentSessionReads — maintained concurrent session reads
 */
async function advancedExample() {
  console.log('=== TypeScript Toolplane Advanced Example ===');
  console.log('Maintained path: gRPC control-plane helpers');
  console.log('gRPC-only maintained SDK surface\n');

  await testGRPCProtocol();
  console.log('\n' + '='.repeat(60) + '\n');
  await testBoundaryErrors();
  console.log('\n' + '='.repeat(60) + '\n');
  await testConcurrentSessionReads();
}

async function testBoundaryErrors() {
  console.log('⚠️  Boundary And Connection Errors');
  console.log('-'.repeat(30));

  const invalidClient = new ToolplaneClient({
    protocol: ClientProtocol.GRPC,
    serverHost: 'invalid-host',
    serverPort: 9999,
    sessionId: 'invalid-session',
    userId: 'invalid-user',
    apiKey: grpcApiKey,
    timeout: 10000,
  });

  try {
    await invalidClient.connect();
    console.log('❌ Expected invalid-host connection to fail, but it succeeded');
  } catch (error) {
    if (error instanceof ConnectionError || error instanceof ToolplaneError) {
      console.log(`   Invalid host error handled: ${error.message}`);
    } else {
      console.log(`   Invalid host produced unexpected error: ${error}`);
    }
  } finally {
    if (invalidClient.isConnected()) {
      await invalidClient.disconnect();
    }
  }

  const client = new ToolplaneClient({
    protocol: ClientProtocol.GRPC,
    serverHost,
    serverPort: 9001,
    sessionId: 'boundary-grpc-session',
    userId: 'boundary-grpc-user',
    apiKey: grpcApiKey,
  });

  try {
    await client.connect();
    console.log('✅ gRPC connection established for boundary checks');

    await client.registerTool('orphan_tool', 'tool without machine ownership', JSON.stringify({ type: 'object' }));
    console.log('❌ Expected machine-ownership boundary error, but registerTool succeeded');

  } catch (error) {
    if (error instanceof ProtocolError) {
      console.log(`   Protocol boundary handled: ${error.message}`);
    } else if (error instanceof ToolplaneError) {
      console.log(`   Boundary error handled: ${error.message}`);
    } else {
      console.log(`   Unexpected boundary error: ${error}`);
    }
  } finally {
    if (client.isConnected()) {
      await client.disconnect();
    }
    console.log('\n🔌 Boundary test client disconnected');
  }
}

async function testGRPCProtocol() {
  console.log('⚡ Maintained gRPC Control-Plane Testing');
  console.log('-'.repeat(30));

  const client = new ToolplaneClient({
    protocol: ClientProtocol.GRPC,
    serverHost,
    serverPort: 9001,
    sessionId: 'advanced-grpc-session',
    userId: 'advanced-grpc-user',
    apiKey: grpcApiKey
  });

  try {
    await client.connect();
    console.log('✅ gRPC connection established');

    const session = await client.createSession(
      'Advanced TypeScript Session',
      'Advanced example demonstrating maintained gRPC capabilities',
      'advanced-examples'
    );
    console.log(`\n📁 Session created: ${session.name}`);
    console.log(`   ID: ${session.id}`);
    console.log(`   Namespace: ${session.namespace}`);

    const tools = [
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
        tags: ['session', 'status', 'operations']
      },
      {
        sessionId: session.id,
        name: 'change_summary',
        description: 'Build a concise summary for a rollout or change event',
        schema: JSON.stringify({
          type: 'object',
          properties: {
            service: { type: 'string' },
            state: { type: 'string' }
          },
          required: ['service', 'state']
        }),
        tags: ['change', 'summary', 'operations']
      },
      {
        sessionId: session.id,
        name: 'incident_brief',
        description: 'Generate a brief operator-facing incident summary',
        schema: JSON.stringify({
          type: 'object',
          properties: {
            service: { type: 'string' },
            severity: { type: 'string' }
          },
          required: ['service', 'severity']
        }),
        tags: ['incident', 'summary', 'operations']
      }
    ];

    const machine = await client.registerMachine('advanced-typescript-machine', '2.1.0', tools.map((tool) => ({
      ...tool,
      config: {
        version: '2.0',
        language: 'typescript',
        created_by: 'advanced-example'
      }
    })));
    console.log(`\n🖥️  Machine Registration:`);
    console.log(`   ID: ${machine.id}`);
    console.log(`   Language: ${machine.sdkLanguage}`);
    console.log(`   Version: ${machine.sdkVersion}`);

    const allTools = await client.listTools();
    console.log(`\n📋 Total Tools Registered: ${allTools.length}`);
    allTools.forEach((tool, index) => {
      console.log(`   ${index + 1}. ${tool.name}`);
      console.log(`      Description: ${tool.description}`);
      console.log(`      Tags: ${tool.tags.join(', ')}`);
      if (index < allTools.length - 1) console.log();
    });

    console.log('\nℹ️  Live tool execution still requires a provider loop to claim requests. This example focuses on the maintained session and machine lifecycle.');

    const finalSession = await client.getSession();
    console.log(`\n📄 Final Session State:`);
    console.log(`   Name: ${finalSession.name}`);
    console.log(`   Created: ${finalSession.createdAt}`);
    console.log(`   Namespace: ${finalSession.namespace}`);

  } catch (error) {
    console.error('❌ gRPC Protocol Error:', error);
  } finally {
    await client.disconnect();
    console.log('\n🔌 gRPC client disconnected');
  }
}

async function testConcurrentSessionReads() {
  console.log('🔄 Concurrent Session Reads');
  console.log('-'.repeat(30));

  const client = ToolplaneClient.createGRPCClient(
    serverHost,
    9001,
    'concurrent-session',
    'concurrent-user',
    grpcApiKey,
  );

  try {
    await client.connect();
    console.log('✅ Connected for concurrent maintained testing');

    const session = await client.createSession(
      'Concurrent Session Reads',
      'Concurrent read example for the maintained gRPC path',
      'advanced-examples',
    );
    console.log(`📁 Created session: ${session.name} (${session.id})`);

    console.log('\n⚡ Running concurrent session reads...');
    const startTime = Date.now();

    const results = await Promise.all(
      Array.from({ length: 4 }, async (_, index) => {
        const sessionInfo = await client.getSession();
        return `Reader ${index + 1}: ${sessionInfo.name} (${sessionInfo.id})`;
      }),
    );
    const endTime = Date.now();

    console.log(`\n📊 Concurrent Results (${endTime - startTime}ms):`);
    results.forEach((result) => {
      console.log(`   ${result}`);
    });

  } catch (error) {
    console.error('❌ Concurrent Session Read Error:', error);
  } finally {
    await client.disconnect();
    console.log('\n🔌 Concurrent client disconnected');
  }
}

if (require.main === module) {
  advancedExample().catch(console.error);
}