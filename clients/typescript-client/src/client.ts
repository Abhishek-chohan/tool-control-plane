#!/usr/bin/env ts-node

import { ToolplaneClient } from './core/toolplane_client';

function readTlsConfigFromEnv() {
  const useTLS = (process.env.TOOLPLANE_USE_TLS || '').toLowerCase() === 'true';
  if (!useTLS) {
    return undefined;
  }

  return {
    enabled: true,
    caCertPath: process.env.TOOLPLANE_TLS_CA_CERT_PATH || undefined,
    serverName: process.env.TOOLPLANE_TLS_SERVER_NAME || undefined,
  };
}

async function main() {
  console.log('=== TypeScript Toolplane Client ===');
  console.log('Maintained path: gRPC control-plane helpers');

  // Get server configuration from environment or use defaults
  const serverHost = process.env.TOOLPLANE_SERVER_HOST || 'localhost';
  const grpcPort = Number(process.env.TOOLPLANE_SERVER_PORT || 9001);
  const sessionId = process.env.TOOLPLANE_SESSION_ID || 'typescript-client-session';
  const userId = process.env.TOOLPLANE_USER_ID || 'typescript-client-user';
  const apiKey = process.env.TOOLPLANE_API_KEY || 'toolplane-conformance-fixture-key';

  const serverPort = grpcPort;
  console.log(`Using maintained gRPC control-plane protocol on port ${serverPort}`);

  const client = ToolplaneClient.createGRPCClient(serverHost, serverPort, sessionId, userId, apiKey, readTlsConfigFromEnv());

  try {
    // Connect to server
    console.log(`Connecting to ${serverHost}:${serverPort}...`);
    await client.connect();
    console.log('✅ Connected successfully!');

    await runBasicOperations(client);
    await runGRPCFeatures(client);

  } catch (error) {
    console.error('❌ Error:', error);
    process.exit(1);
  } finally {
    await client.disconnect();
  }
}

async function runBasicOperations(client: ToolplaneClient): Promise<void> {
  console.log('\n=== Connectivity ===');

  try {
    const pingResult = await client.ping();
    console.log(`🏓 Ping: ${pingResult}`);

    console.log('ℹ️  Live gRPC tool execution requires a machine-backed provider loop. This maintained demo focuses on session and machine lifecycle first.');

  } catch (error) {
    console.error('Error in basic operations:', error);
  }
}

async function runGRPCFeatures(client: ToolplaneClient): Promise<void> {
  console.log('\n=== Maintained gRPC Control-Plane Features ===');

  try {
    const session = await client.createSession(
      'TypeScript Control Plane Session',
      'Provider-backed session created by the TypeScript gRPC client demo',
      'examples'
    );
    console.log(`📁 Created session: ${session.name} (ID: ${session.id})`);

    const toolDefinitions = [
      {
        sessionId: session.id,
        name: 'session_status',
        description: 'Report session context for a connected operator or automation client',
        schema: JSON.stringify({
          type: 'object',
          properties: {
            requester: { type: 'string', description: 'The caller requesting session state' }
          },
          required: ['requester']
        }),
        config: {
          version: '1.0',
          language: 'typescript',
          created_by: 'typescript-client'
        },
        tags: ['session', 'status', 'operator']
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
        config: {
          version: '1.0',
          language: 'typescript',
          created_by: 'typescript-client'
        },
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
        config: {
          version: '1.0',
          language: 'typescript',
          created_by: 'typescript-client'
        },
        tags: ['incident', 'summary', 'operations']
      }
    ];

    const machine = await client.registerMachine('', '1.0.0', toolDefinitions);
    console.log(`🖥️  Registered machine: ${machine.id}`);

    const allTools = await client.listTools();
    console.log(`📋 Session contains ${allTools.length} tools:`);
    allTools.forEach((tool, index) => {
      console.log(`   ${index + 1}. ${tool.name} - ${tool.description}`);
    });

    console.log('ℹ️  Live tool execution now has a maintained TypeScript provider runtime. See src/examples/provider_runtime_example.ts for the end-to-end provider loop.');

    const sessionInfo = await client.getSession();
    console.log('📄 Session details:');
    console.log(`   Name: ${sessionInfo.name}`);
    console.log(`   ID: ${sessionInfo.id}`);
    console.log(`   Namespace: ${sessionInfo.namespace}`);
    console.log(`   Created: ${sessionInfo.createdAt}`);

  } catch (error) {
    console.error('Error in gRPC features:', error);
  }
}

// Run the main function
if (require.main === module) {
  main().catch(console.error);
}