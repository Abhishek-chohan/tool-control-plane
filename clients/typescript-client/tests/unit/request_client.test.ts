import assert from 'node:assert/strict';
import test from 'node:test';

import { ToolplaneClient } from '../../src/core/toolplane_client';
import { ClientProtocol } from '../../src/interfaces';
import {
  CancelRequestResponse,
  ExecuteToolResponse,
  ListRequestsResponse,
  Request as ProtoRequest,
} from '../../src/proto/proto/service_pb';

type MutableClientState = {
  connected: boolean;
  toolClient: unknown;
  sessionClient: unknown;
  requestsClient: Record<string, unknown>;
  tasksClient: unknown;
  machineClient: unknown;
  machineId: string;
};

function createRequest(overrides: Partial<{
  id: string;
  sessionId: string;
  toolName: string;
  status: string;
  input: string;
  result: string;
  resultType: string;
  error: string;
  createdAt: string;
  updatedAt: string;
  executingMachineId: string;
  streamResults: string[];
}> = {}): ProtoRequest {
  const request = new ProtoRequest();
  request.setId(overrides.id ?? 'request-1');
  request.setSessionId(overrides.sessionId ?? 'session-1');
  request.setToolName(overrides.toolName ?? 'demo_tool');
  request.setStatus(overrides.status ?? 'pending');
  request.setInput(overrides.input ?? '{"message":"hello"}');
  request.setResult(overrides.result ?? '');
  request.setResultType(overrides.resultType ?? '');
  request.setError(overrides.error ?? '');
  request.setCreatedAt(overrides.createdAt ?? '2025-01-01T00:00:00Z');
  request.setUpdatedAt(overrides.updatedAt ?? '2025-01-01T00:01:00Z');
  request.setExecutingMachineId(overrides.executingMachineId ?? 'machine-1');
  request.setStreamResultsList(overrides.streamResults ?? []);
  return request;
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

function createConnectedClient(
  requestsClient: Record<string, unknown>,
  toolClient: Record<string, unknown> = {},
): ToolplaneClient {
  const client = new ToolplaneClient({
    protocol: ClientProtocol.GRPC,
    serverHost: 'localhost',
    serverPort: 9001,
    sessionId: 'session-1',
    userId: 'unit-user',
  });

  const state = client as unknown as MutableClientState;
  state.connected = true;
  state.toolClient = toolClient;
  state.sessionClient = {};
  state.requestsClient = requestsClient;
  state.tasksClient = {};
  state.machineClient = {};
  state.machineId = '';

  return client;
}

test('executeTool submits the request and normalizes the final payload', async () => {
  const response = new ExecuteToolResponse();
  response.setRequestId('request-99');

  let getRequestCalls = 0;
  const client = createConnectedClient(
    {
      getRequest: unaryResponse(() => {
        getRequestCalls += 1;
        return createRequest({
          id: 'request-99',
          status: getRequestCalls === 1 ? 'running' : 'done',
          result: '{"echo":"hello"}',
          resultType: 'resolution',
        });
      }),
    },
    {
      executeTool: unaryResponse((request) => {
        assert.equal((request as { getToolName(): string }).getToolName(), 'demo_tool');
        assert.equal((request as { getInput(): string }).getInput(), '{"message":"hello"}');
        return response;
      }),
    },
  );

  const result = await client.executeTool('demo_tool', { message: 'hello' });

  assert.equal(result.id, 'request-99');
  assert.equal(result.status, 'done');
  assert.deepEqual(result.result, { echo: 'hello' });
  assert.equal(result.resultType, 'resolution');
});

test('createRequest returns a normalized request payload', async () => {
  const client = createConnectedClient({
    createRequest: unaryResponse(createRequest({ status: 'running', result: '{"ok":true}', resultType: 'json' })),
  });

  const result = await client.createRequest('demo_tool', '{"message":"hello"}');

  assert.equal(result.id, 'request-1');
  assert.equal(result.toolName, 'demo_tool');
  assert.equal(result.status, 'running');
  assert.deepEqual(result.result, { ok: true });
  assert.equal(result.resultType, 'json');
});

test('getRequest returns a normalized request payload', async () => {
  const client = createConnectedClient({
    getRequest: unaryResponse(createRequest({ id: 'request-42', status: 'done', streamResults: ['"a"', '"b"'] })),
  });

  const result = await client.getRequest('request-42');

  assert.equal(result.id, 'request-42');
  assert.equal(result.status, 'done');
  assert.deepEqual(result.streamResults, ['a', 'b']);
});

test('listRequests passes filters and normalizes responses', async () => {
  const response = new ListRequestsResponse();
  response.setRequestsList([
    createRequest({ id: 'request-1', status: 'pending' }),
    createRequest({ id: 'request-2', status: 'done', result: '{"ok":true}' }),
  ]);

  const client = createConnectedClient({
    listRequests: unaryResponse((request) => {
      assert.equal((request as { getStatus(): string }).getStatus(), 'running');
      assert.equal((request as { getToolName(): string }).getToolName(), 'demo_tool');
      assert.equal((request as { getLimit(): number }).getLimit(), 5);
      assert.equal((request as { getOffset(): number }).getOffset(), 2);
      return response;
    }),
  });

  const requests = await client.listRequests({ status: 'running', toolName: 'demo_tool', limit: 5, offset: 2 });

  assert.deepEqual(requests.map((request) => request.id), ['request-1', 'request-2']);
  assert.deepEqual(requests[1].result, { ok: true });
});

test('cancelRequest returns the server success flag', async () => {
  const response = new CancelRequestResponse();
  response.setSuccess(true);

  const client = createConnectedClient({
    cancelRequest: unaryResponse(response),
  });

  const success = await client.cancelRequest('request-1');

  assert.equal(success, true);
});