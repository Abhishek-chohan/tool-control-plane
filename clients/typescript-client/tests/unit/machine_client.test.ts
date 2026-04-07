import assert from 'node:assert/strict';
import test from 'node:test';

import { ToolplaneClient } from '../../src/core/toolplane_client';
import { ClientProtocol } from '../../src/interfaces';
import {
  DrainMachineResponse,
  ListMachinesResponse,
  Machine as ProtoMachine,
  UnregisterMachineResponse,
} from '../../src/proto/proto/service_pb';

type MutableClientState = {
  connected: boolean;
  toolClient: unknown;
  sessionClient: unknown;
  requestsClient: unknown;
  tasksClient: unknown;
  machineClient: Record<string, unknown>;
  machineId: string;
};

function createMachine(overrides: Partial<{
  id: string;
  sessionId: string;
  sdkVersion: string;
  sdkLanguage: string;
  ip: string;
  createdAt: string;
  lastPingAt: string;
}> = {}): ProtoMachine {
  const machine = new ProtoMachine();
  machine.setId(overrides.id ?? 'machine-1');
  machine.setSessionId(overrides.sessionId ?? 'session-1');
  machine.setSdkVersion(overrides.sdkVersion ?? '1.0.0-test');
  machine.setSdkLanguage(overrides.sdkLanguage ?? 'typescript');
  machine.setIp(overrides.ip ?? '127.0.0.1');
  machine.setCreatedAt(overrides.createdAt ?? '2025-01-01T00:00:00Z');
  machine.setLastPingAt(overrides.lastPingAt ?? '2025-01-01T00:01:00Z');
  return machine;
}

function unaryResponse<T>(responseOrFactory: T | ((request: unknown) => T)) {
  return (request: unknown, _metadata: unknown, _options: unknown, callback: (error: null, response: T) => void) => {
    const response = typeof responseOrFactory === 'function'
      ? (responseOrFactory as (input: unknown) => T)(request)
      : responseOrFactory;
    callback(null, response);
    return {};
  };
}

function createConnectedClient(machineClient: Record<string, unknown>, machineId: string = ''): ToolplaneClient {
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
  state.tasksClient = {};
  state.machineClient = machineClient;
  state.machineId = machineId;

  return client;
}

test('listMachines normalizes machine responses', async () => {
  const response = new ListMachinesResponse();
  response.setMachinesList([createMachine()]);

  const client = createConnectedClient({
    listMachines: unaryResponse(response),
  });

  const machines = await client.listMachines();

  assert.deepEqual(machines, [
    {
      id: 'machine-1',
      sessionId: 'session-1',
      sdkVersion: '1.0.0-test',
      sdkLanguage: 'typescript',
      ip: '127.0.0.1',
      createdAt: '2025-01-01T00:00:00Z',
      lastPingAt: '2025-01-01T00:01:00Z',
    },
  ]);
});

test('getMachine returns a normalized machine payload', async () => {
  const machine = createMachine({ id: 'machine-42', sdkVersion: '2.0.0' });
  const client = createConnectedClient({
    getMachine: unaryResponse(machine),
  });

  const result = await client.getMachine('machine-42');

  assert.equal(result.id, 'machine-42');
  assert.equal(result.sdkVersion, '2.0.0');
  assert.equal(result.sessionId, 'session-1');
});

test('updateMachinePing uses the registered machine by default and returns a normalized machine payload', async () => {
  const client = createConnectedClient(
    {
      updateMachinePing: unaryResponse((request: { getMachineId(): string }) => {
        assert.equal(request.getMachineId(), 'machine-1');
        return createMachine({ id: 'machine-1', lastPingAt: '2025-01-01T00:02:00Z' });
      }),
    },
    'machine-1',
  );

  const result = await client.updateMachinePing();

  assert.equal(result.id, 'machine-1');
  assert.equal(result.lastPingAt, '2025-01-01T00:02:00Z');
});

test('drainMachine uses the registered machine by default and clears cached machine state', async () => {
  const response = new DrainMachineResponse();
  response.setDrained(true);

  const client = createConnectedClient(
    {
      drainMachine: unaryResponse(response),
    },
    'machine-1',
  );

  const drained = await client.drainMachine();

  assert.equal(drained, true);
  assert.equal((client as unknown as MutableClientState).machineId, '');
});

test('unregisterMachine uses the registered machine by default and clears cached machine state', async () => {
  const response = new UnregisterMachineResponse();
  response.setSuccess(true);

  const client = createConnectedClient(
    {
      unregisterMachine: unaryResponse(response),
    },
    'machine-1',
  );

  const success = await client.unregisterMachine();

  assert.equal(success, true);
  assert.equal((client as unknown as MutableClientState).machineId, '');
});