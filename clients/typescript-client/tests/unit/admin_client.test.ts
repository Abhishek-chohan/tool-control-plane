import assert from 'node:assert/strict';
import test from 'node:test';

import { ToolplaneClient } from '../../src/core/toolplane_client';
import { ClientProtocol } from '../../src/interfaces';
import {
  ApiKey as ProtoApiKey,
  DeleteToolResponse,
  GetToolResponse,
  ListApiKeysResponse,
  ListSessionsResponse,
  RevokeApiKeyResponse,
  Session as ProtoSession,
  Tool as ProtoTool,
} from '../../src/proto/proto/service_pb';

type MutableClientState = {
  connected: boolean;
  toolClient: Record<string, unknown>;
  sessionClient: Record<string, unknown>;
  requestsClient: unknown;
  tasksClient: unknown;
  machineClient: unknown;
  machineId: string;
};

function createTool(overrides: Partial<{
  id: string;
  name: string;
  description: string;
  schema: string;
  config: Record<string, string>;
  createdAt: string;
  lastPingAt: string;
  sessionId: string;
  tags: string[];
}> = {}): ProtoTool {
  const tool = new ProtoTool();
  tool.setId(overrides.id ?? 'tool-1');
  tool.setName(overrides.name ?? 'echo');
  tool.setDescription(overrides.description ?? 'echo tool');
  tool.setSchema(overrides.schema ?? '{"type":"object"}');
  for (const [key, value] of Object.entries(overrides.config ?? { region: 'test' })) {
    tool.getConfigMap().set(key, value);
  }
  tool.setCreatedAt(overrides.createdAt ?? '2025-01-01T00:00:00Z');
  tool.setLastPingAt(overrides.lastPingAt ?? '2025-01-01T00:01:00Z');
  tool.setSessionId(overrides.sessionId ?? 'session-1');
  tool.setTagsList(overrides.tags ?? ['core']);
  return tool;
}

function createSession(overrides: Partial<{
  id: string;
  name: string;
  description: string;
  createdAt: string;
  createdBy: string;
  apiKey: string;
  namespace: string;
}> = {}): ProtoSession {
  const session = new ProtoSession();
  session.setId(overrides.id ?? 'session-1');
  session.setName(overrides.name ?? 'primary');
  session.setDescription(overrides.description ?? 'demo session');
  session.setCreatedAt(overrides.createdAt ?? '2025-01-01T00:00:00Z');
  session.setCreatedBy(overrides.createdBy ?? 'unit-user');
  session.setApiKey(overrides.apiKey ?? 'api-key');
  session.setNamespace(overrides.namespace ?? 'default');
  return session;
}

function createApiKey(overrides: Partial<{
  id: string;
  name: string;
  key: string;
  sessionId: string;
  createdAt: string;
  createdBy: string;
  revokedAt: string;
}> = {}): ProtoApiKey {
  const apiKey = new ProtoApiKey();
  apiKey.setId(overrides.id ?? 'key-1');
  apiKey.setName(overrides.name ?? 'cli');
  apiKey.setKey(overrides.key ?? 'secret');
  apiKey.setSessionId(overrides.sessionId ?? 'session-1');
  apiKey.setCreatedAt(overrides.createdAt ?? '2025-01-01T00:00:00Z');
  apiKey.setCreatedBy(overrides.createdBy ?? 'unit-user');
  apiKey.setRevokedAt(overrides.revokedAt ?? '');
  return apiKey;
}

function unaryResponse<T>(responseOrFactory: T | ((request: unknown) => T)) {
  return (request: unknown, _metadata: unknown, _options: unknown, callback: (error: null, response: T) => void) => {
    const response = typeof responseOrFactory === 'function'
      ? (responseOrFactory as (request: unknown) => T)(request)
      : responseOrFactory;
    callback(null, response);
    return {};
  };
}

function createConnectedClient(overrides: {
  toolClient?: Record<string, unknown>;
  sessionClient?: Record<string, unknown>;
} = {}): ToolplaneClient {
  const client = new ToolplaneClient({
    protocol: ClientProtocol.GRPC,
    serverHost: 'localhost',
    serverPort: 9001,
    sessionId: 'session-1',
    userId: 'unit-user',
  });

  const state = client as unknown as MutableClientState;
  state.connected = true;
  state.toolClient = overrides.toolClient ?? {};
  state.sessionClient = overrides.sessionClient ?? {};
  state.requestsClient = {};
  state.tasksClient = {};
  state.machineClient = {};
  state.machineId = '';

  return client;
}

test('getToolById returns a normalized tool payload', async () => {
  const response = new GetToolResponse();
  response.setTool(createTool({ id: 'tool-42' }));

  const client = createConnectedClient({
    toolClient: {
      getToolById: unaryResponse(response),
    },
  });

  const tool = await client.getToolById('tool-42');

  assert.equal(tool.id, 'tool-42');
  assert.equal(tool.name, 'echo');
  assert.deepEqual(tool.config, { region: 'test' });
});

test('getToolByName returns a normalized tool payload', async () => {
  const response = new GetToolResponse();
  response.setTool(createTool({ name: 'adder', tags: ['math'] }));

  const client = createConnectedClient({
    toolClient: {
      getToolByName: unaryResponse(response),
    },
  });

  const tool = await client.getToolByName('adder');

  assert.equal(tool.name, 'adder');
  assert.deepEqual(tool.tags, ['math']);
});

test('deleteTool returns the server success flag', async () => {
  const response = new DeleteToolResponse();
  response.setSuccess(true);

  const client = createConnectedClient({
    toolClient: {
      deleteTool: unaryResponse(response),
    },
  });

  const success = await client.deleteTool('tool-1');

  assert.equal(success, true);
});

test('listSessions uses the configured user id and normalizes responses', async () => {
  const response = new ListSessionsResponse();
  response.setSessionsList([
    createSession({ id: 'session-1' }),
    createSession({ id: 'session-2', name: 'secondary' }),
  ]);

  const client = createConnectedClient({
    sessionClient: {
      listSessions: unaryResponse((request) => {
        assert.equal((request as { getUserId(): string }).getUserId(), 'unit-user');
        return response;
      }),
    },
  });

  const sessions = await client.listSessions();

  assert.deepEqual(sessions.map((session) => session.id), ['session-1', 'session-2']);
  assert.equal(sessions[1].name, 'secondary');
});

test('updateSession sends updated fields and normalizes the response', async () => {
  const client = createConnectedClient({
    sessionClient: {
      updateSession: unaryResponse((request) => {
        assert.equal((request as { getSessionId(): string }).getSessionId(), 'session-1');
        assert.equal((request as { getName(): string }).getName(), 'updated');
        assert.equal((request as { getDescription(): string }).getDescription(), 'refreshed');
        assert.equal((request as { getNamespace(): string }).getNamespace(), 'ops');
        return createSession({ name: 'updated', description: 'refreshed', namespace: 'ops' });
      }),
    },
  });

  const session = await client.updateSession('updated', 'refreshed', 'ops');

  assert.equal(session.name, 'updated');
  assert.equal(session.namespace, 'ops');
});

test('createApiKey returns a normalized api key payload', async () => {
  const client = createConnectedClient({
    sessionClient: {
      createApiKey: unaryResponse((request) => {
        assert.equal((request as { getSessionId(): string }).getSessionId(), 'session-1');
        assert.equal((request as { getName(): string }).getName(), 'cli');
        return createApiKey({ name: 'cli' });
      }),
    },
  });

  const apiKey = await client.createApiKey('cli');

  assert.equal(apiKey.id, 'key-1');
  assert.equal(apiKey.name, 'cli');
  assert.equal(apiKey.key, 'secret');
});

test('listApiKeys normalizes api key responses', async () => {
  const response = new ListApiKeysResponse();
  response.setApiKeysList([
    createApiKey({ id: 'key-1' }),
    createApiKey({ id: 'key-2', revokedAt: '2025-01-01T01:00:00Z' }),
  ]);

  const client = createConnectedClient({
    sessionClient: {
      listApiKeys: unaryResponse(response),
    },
  });

  const apiKeys = await client.listApiKeys();

  assert.deepEqual(apiKeys.map((apiKey) => apiKey.id), ['key-1', 'key-2']);
  assert.equal(apiKeys[1].revokedAt, '2025-01-01T01:00:00Z');
});

test('revokeApiKey returns the server success flag', async () => {
  const response = new RevokeApiKeyResponse();
  response.setSuccess(true);

  const client = createConnectedClient({
    sessionClient: {
      revokeApiKey: unaryResponse(response),
    },
  });

  const success = await client.revokeApiKey('key-1');

  assert.equal(success, true);
});