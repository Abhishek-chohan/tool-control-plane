import assert from 'node:assert/strict';
import test from 'node:test';

import { ProviderRuntime } from '../../src/provider_runtime';
import {
  Machine,
  ProviderRuntimeClient,
  ProviderRuntimeSessionClient,
  Request,
  RequestListOptions,
  RequestUpdate,
  RegisterToolOptions,
  RegisterToolRequest,
  Session,
  Tool,
} from '../../src/interfaces';

function createRequest(overrides: Partial<Request> = {}): Request {
  return {
    id: overrides.id ?? 'request-1',
    sessionId: overrides.sessionId ?? 'session-1',
    toolName: overrides.toolName ?? 'echo_tool',
    status: overrides.status ?? 'pending',
    input: overrides.input ?? '{"message":"hello"}',
    createdAt: overrides.createdAt ?? '2025-01-01T00:00:00Z',
    updatedAt: overrides.updatedAt ?? '2025-01-01T00:00:00Z',
    executingMachineId: overrides.executingMachineId ?? '',
    result: overrides.result,
    resultType: overrides.resultType,
    error: overrides.error,
    streamResults: overrides.streamResults,
  };
}

class FakeSessionClient implements ProviderRuntimeSessionClient {
  connected = false;

  machineId = '';

  heartbeatCount = 0;

  drainCount = 0;

  unregisterCount = 0;

  registeredTools: Array<{ name: string; description: string; schema: string; tags: string[] }> = [];

  registeredMachineTools: RegisterToolRequest[] = [];

  pendingRequests: Request[] = [];

  updatedRequests: Array<{ requestId: string; update: RequestUpdate }> = [];

  appendedChunks: Array<{ requestId: string; chunks: unknown[]; resultType: string }> = [];

  submittedResults: Array<{ requestId: string; result: unknown; resultType: string; meta: Record<string, string> }> = [];

  constructor(readonly sessionId: string) {}

  async connect(): Promise<void> {
    this.connected = true;
  }

  async disconnect(): Promise<void> {
    this.connected = false;
  }

  isConnected(): boolean {
    return this.connected;
  }

  async createSession(name: string, description: string, namespace: string = 'default', requestedSessionId: string = ''): Promise<Session> {
    return {
      id: requestedSessionId || this.sessionId,
      name,
      description,
      createdAt: '2025-01-01T00:00:00Z',
      createdBy: 'provider-test',
      apiKey: '',
      namespace,
    };
  }

  async getSession(): Promise<Session> {
    return {
      id: this.sessionId,
      name: `Session ${this.sessionId}`,
      description: 'Provider runtime test session',
      createdAt: '2025-01-01T00:00:00Z',
      createdBy: 'provider-test',
      apiKey: '',
      namespace: 'tests',
    };
  }

  async registerMachine(machineId: string = '', sdkVersion: string = '1.0.0', tools: RegisterToolRequest[] = []): Promise<Machine> {
    this.machineId = machineId || `machine-${this.sessionId}`;
    this.registeredMachineTools = tools;
    return {
      id: this.machineId,
      sessionId: this.sessionId,
      sdkVersion,
      sdkLanguage: 'typescript',
      ip: '127.0.0.1',
      createdAt: '2025-01-01T00:00:00Z',
      lastPingAt: '2025-01-01T00:00:00Z',
    };
  }

  async registerTool(
    name: string,
    description: string,
    schema: string,
    _config: Record<string, string> = {},
    tags: string[] = [],
    _options: RegisterToolOptions = {},
  ): Promise<Tool> {
    this.registeredTools.push({ name, description, schema, tags });
    return {
      id: `${name}-id`,
      name,
      description,
      schema,
      config: {},
      createdAt: '2025-01-01T00:00:00Z',
      sessionId: this.sessionId,
      tags,
    };
  }

  async listRequests(options: RequestListOptions = {}): Promise<Request[]> {
    return this.pendingRequests.filter((request) => {
      if (options.status && request.status !== options.status) {
        return false;
      }
      if (options.toolName && request.toolName !== options.toolName) {
        return false;
      }
      return true;
    });
  }

  async claimRequest(requestId: string, machineId: string = ''): Promise<Request> {
    const request = this.pendingRequests.find((candidate) => candidate.id === requestId);
    if (!request) {
      throw new Error(`Unknown request ${requestId}`);
    }

    request.status = 'claimed';
    request.executingMachineId = machineId || this.machineId;
    return { ...request };
  }

  async updateRequest(requestId: string, update: RequestUpdate): Promise<Request> {
    const request = this.pendingRequests.find((candidate) => candidate.id === requestId);
    if (!request) {
      throw new Error(`Unknown request ${requestId}`);
    }

    this.updatedRequests.push({ requestId, update });
    request.status = update.status ?? request.status;
    request.resultType = update.resultType ?? request.resultType;
    request.result = update.result ?? request.result;
    return { ...request };
  }

  async appendRequestChunks(requestId: string, chunks: unknown[], resultType: string = 'streaming'): Promise<boolean> {
    this.appendedChunks.push({ requestId, chunks, resultType });
    const request = this.pendingRequests.find((candidate) => candidate.id === requestId);
    if (request) {
      request.streamResults = [...(request.streamResults ?? []), ...chunks];
      request.resultType = resultType;
    }
    return true;
  }

  async submitRequestResult(
    requestId: string,
    result: unknown,
    resultType: string = 'resolution',
    meta: Record<string, string> = {},
  ): Promise<boolean> {
    this.submittedResults.push({ requestId, result, resultType, meta });
    const request = this.pendingRequests.find((candidate) => candidate.id === requestId);
    if (request) {
      request.status = resultType === 'rejection' ? 'failure' : 'done';
      request.result = result;
      request.resultType = resultType;
      if (resultType === 'rejection') {
        request.error = typeof result === 'object' && result !== null ? String((result as { error?: unknown }).error ?? '') : String(result);
      }
    }
    return true;
  }

  async updateMachinePing(machineId: string = ''): Promise<Machine> {
    this.heartbeatCount += 1;
    return {
      id: machineId || this.machineId,
      sessionId: this.sessionId,
      sdkVersion: '1.0.0-test',
      sdkLanguage: 'typescript',
      ip: '127.0.0.1',
      createdAt: '2025-01-01T00:00:00Z',
      lastPingAt: '2025-01-01T00:00:01Z',
    };
  }

  async unregisterMachine(): Promise<boolean> {
    this.unregisterCount += 1;
    this.machineId = '';
    return true;
  }

  async drainMachine(): Promise<boolean> {
    this.drainCount += 1;
    this.machineId = '';
    return true;
  }
}

class FakeProviderClient extends FakeSessionClient implements ProviderRuntimeClient {
  private readonly sessionClients = new Map<string, FakeSessionClient>();

  constructor() {
    super('');
  }

  forkSession(sessionId: string): ProviderRuntimeSessionClient {
    const existing = this.sessionClients.get(sessionId);
    if (existing) {
      return existing;
    }

    const client = new FakeSessionClient(sessionId);
    this.sessionClients.set(sessionId, client);
    return client;
  }

  lookup(sessionId: string): FakeSessionClient {
    return this.forkSession(sessionId) as FakeSessionClient;
  }
}

test('ProviderRuntime processes a unary request end to end', async () => {
  const client = new FakeProviderClient();
  const runtime = new ProviderRuntime(client, { sdkVersion: '9.9.9-test' });

  const session = await runtime.createSession({
    sessionId: 'session-unary',
    name: 'Unary Session',
    description: 'Unary provider test',
  });

  await runtime.registerTool({
    sessionId: session.id,
    name: 'echo_tool',
    description: 'Echo tool',
    handler: async (input) => ({ echo: input.message }),
  });

  const sessionClient = client.lookup(session.id);
  sessionClient.pendingRequests.push(createRequest({
    id: 'request-unary',
    sessionId: session.id,
    toolName: 'echo_tool',
    input: '{"message":"hello"}',
  }));

  await runtime.pollOnce();

  assert.equal(sessionClient.updatedRequests[0]?.update.status, 'running');
  assert.deepEqual(sessionClient.submittedResults[0], {
    requestId: 'request-unary',
    result: { echo: 'hello' },
    resultType: 'resolution',
    meta: { handled_by: 'typescript-provider-runtime' },
  });
});

test('ProviderRuntime streams chunks and submits the retained final payload', async () => {
  const client = new FakeProviderClient();
  const runtime = new ProviderRuntime(client, { sdkVersion: '9.9.9-test' });

  const session = await runtime.createSession({
    sessionId: 'session-stream',
    name: 'Stream Session',
    description: 'Streaming provider test',
  });

  await runtime.registerTool({
    sessionId: session.id,
    name: 'stream_tool',
    description: 'Stream tool',
    stream: true,
    handler: async function* (input) {
      const prefix = String(input.prefix ?? 'chunk');
      yield `${prefix}-1`;
      yield `${prefix}-2`;
    },
  });

  const sessionClient = client.lookup(session.id);
  sessionClient.pendingRequests.push(createRequest({
    id: 'request-stream',
    sessionId: session.id,
    toolName: 'stream_tool',
    input: '{"prefix":"piece"}',
  }));

  await runtime.pollOnce();

  assert.equal(sessionClient.updatedRequests.at(-1)?.update.resultType, 'streaming');
  assert.deepEqual(sessionClient.appendedChunks, [
    { requestId: 'request-stream', chunks: ['piece-1'], resultType: 'streaming' },
    { requestId: 'request-stream', chunks: ['piece-2'], resultType: 'streaming' },
  ]);
  assert.deepEqual(sessionClient.submittedResults[0], {
    requestId: 'request-stream',
    result: ['piece-1', 'piece-2'],
    resultType: 'resolution',
    meta: { handled_by: 'typescript-provider-runtime' },
  });
});

test('ProviderRuntime start and drain manage heartbeat and graceful machine cleanup', async () => {
  const client = new FakeProviderClient();
  const runtime = new ProviderRuntime(client, {
    pollIntervalMs: 10,
    heartbeatIntervalMs: 5,
    sdkVersion: '9.9.9-test',
  });

  await runtime.createSession({
    sessionId: 'session-drain',
    name: 'Drain Session',
    description: 'Drain provider test',
  });

  await runtime.startInBackground();
  await new Promise((resolve) => setTimeout(resolve, 20));
  await runtime.drain();

  const sessionClient = client.lookup('session-drain');
  assert.equal(sessionClient.heartbeatCount > 0, true);
  assert.equal(sessionClient.drainCount, 1);
  assert.equal(runtime.running, false);
});