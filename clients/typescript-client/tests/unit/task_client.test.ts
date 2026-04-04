import assert from 'node:assert/strict';
import test from 'node:test';

import { ToolplaneClient } from '../../src/core/toolplane_client';
import { ClientProtocol } from '../../src/interfaces';
import {
  CancelTaskResponse,
  ListTasksResponse,
  Task as ProtoTask,
} from '../../src/proto/proto/service_pb';

type MutableClientState = {
  connected: boolean;
  toolClient: unknown;
  sessionClient: unknown;
  requestsClient: unknown;
  tasksClient: Record<string, unknown>;
  machineClient: unknown;
};

function createTask(overrides: Partial<{
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
  completedAt: string;
}> = {}): ProtoTask {
  const task = new ProtoTask();
  task.setId(overrides.id ?? 'task-1');
  task.setSessionId(overrides.sessionId ?? 'session-1');
  task.setToolName(overrides.toolName ?? 'demo_tool');
  task.setStatus(overrides.status ?? 'pending');
  task.setInput(overrides.input ?? '{"message":"hello"}');
  task.setResult(overrides.result ?? '');
  task.setResultType(overrides.resultType ?? '');
  task.setError(overrides.error ?? '');
  task.setCreatedAt(overrides.createdAt ?? '2025-01-01T00:00:00Z');
  task.setUpdatedAt(overrides.updatedAt ?? '2025-01-01T00:01:00Z');
  task.setCompletedAt(overrides.completedAt ?? '');
  return task;
}

function unaryResponse<T>(response: T) {
  return (_request: unknown, _metadata: unknown, _options: unknown, callback: (error: null, response: T) => void) => {
    callback(null, response);
    return {};
  };
}

function createConnectedClient(taskClient: Record<string, unknown>): ToolplaneClient {
  const client = new ToolplaneClient({
    protocol: ClientProtocol.GRPC,
    serverHost: 'localhost',
    serverPort: 9001,
    sessionId: 'session-1',
    userId: 'unit-user',
  });

  const state = client as unknown as MutableClientState;
  state.connected = true;
  state.toolClient = {};
  state.sessionClient = {};
  state.requestsClient = {};
  state.machineClient = {};
  state.tasksClient = taskClient;

  return client;
}

test('createTask returns a normalized task payload', async () => {
  const task = createTask({ status: 'running' });
  const client = createConnectedClient({
    createTask: unaryResponse(task),
  });

  const result = await client.createTask('demo_tool', '{"message":"hello"}');

  assert.equal(result.id, 'task-1');
  assert.equal(result.toolName, 'demo_tool');
  assert.equal(result.status, 'running');
});

test('getTask returns a normalized task payload', async () => {
  const task = createTask({ id: 'task-42', completedAt: '2025-01-01T00:02:00Z' });
  const client = createConnectedClient({
    getTask: unaryResponse(task),
  });

  const result = await client.getTask('task-42');

  assert.equal(result.id, 'task-42');
  assert.equal(result.completedAt, '2025-01-01T00:02:00Z');
});

test('listTasks normalizes task responses', async () => {
  const response = new ListTasksResponse();
  response.setTasksList([
    createTask({ id: 'task-1', status: 'pending' }),
    createTask({ id: 'task-2', status: 'done', result: '{"ok":true}', resultType: 'json' }),
  ]);

  const client = createConnectedClient({
    listTasks: unaryResponse(response),
  });

  const tasks = await client.listTasks();

  assert.deepEqual(tasks.map((task) => task.id), ['task-1', 'task-2']);
  assert.equal(tasks[1].status, 'done');
  assert.equal(tasks[1].resultType, 'json');
});

test('cancelTask returns the server success flag', async () => {
  const response = new CancelTaskResponse();
  response.setSuccess(true);

  const client = createConnectedClient({
    cancelTask: unaryResponse(response),
  });

  const success = await client.cancelTask('task-1');

  assert.equal(success, true);
});