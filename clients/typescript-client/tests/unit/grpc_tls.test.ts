import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';

import * as grpc from '@grpc/grpc-js';

import { ToolplaneClient } from '../../src/core/toolplane_client';
import { ClientConfig, ClientProtocol } from '../../src/interfaces';

type ToolplaneClientWithInternals = ToolplaneClient & {
  config: ClientConfig;
  createChannelCredentials(): grpc.ChannelCredentials;
  channelOptions(): grpc.ChannelOptions;
};

function createBareClient(config: ClientConfig): ToolplaneClientWithInternals {
  const client = Object.create(ToolplaneClient.prototype) as ToolplaneClientWithInternals;
  client.config = config;
  return client;
}

test('createChannelCredentials uses insecure credentials when TLS is disabled', () => {
  const originalCreateInsecure = grpc.credentials.createInsecure;
  const insecureCredentials = { kind: 'insecure' } as unknown as grpc.ChannelCredentials;
  let createInsecureCalls = 0;

  grpc.credentials.createInsecure = (() => {
    createInsecureCalls += 1;
    return insecureCredentials;
  }) as typeof grpc.credentials.createInsecure;

  try {
    const client = createBareClient({
      protocol: ClientProtocol.GRPC,
      serverHost: 'localhost',
      serverPort: 9001,
      sessionId: 'session-1',
      userId: 'user-1',
    });

    const credentials = client.createChannelCredentials();
    assert.equal(credentials, insecureCredentials);
    assert.equal(createInsecureCalls, 1);
  } finally {
    grpc.credentials.createInsecure = originalCreateInsecure;
  }
});

test('createChannelCredentials loads the configured CA bundle when TLS is enabled', () => {
  const originalCreateSsl = grpc.credentials.createSsl;
  const secureCredentials = { kind: 'secure' } as unknown as grpc.ChannelCredentials;
  const caCertificate = Buffer.from('test-ca-certificate');
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'toolplane-ts-tls-'));
  const caPath = path.join(tempDir, 'ca.crt');
  fs.writeFileSync(caPath, caCertificate);

  let createSslArgument: Buffer | undefined;
  grpc.credentials.createSsl = ((rootCerts?: Buffer | null) => {
    createSslArgument = rootCerts ?? undefined;
    return secureCredentials;
  }) as typeof grpc.credentials.createSsl;

  try {
    const client = createBareClient({
      protocol: ClientProtocol.GRPC,
      serverHost: 'localhost',
      serverPort: 9443,
      sessionId: 'session-1',
      userId: 'user-1',
      tls: {
        enabled: true,
        caCertPath: caPath,
        serverName: 'toolplane-server',
      },
    });

    const credentials = client.createChannelCredentials();
    const channelOptions = client.channelOptions();

    assert.equal(credentials, secureCredentials);
    assert.deepEqual(createSslArgument, caCertificate);
    assert.equal(channelOptions['grpc.ssl_target_name_override'], 'toolplane-server');
    assert.equal(channelOptions['grpc.default_authority'], 'toolplane-server');
  } finally {
    grpc.credentials.createSsl = originalCreateSsl;
    fs.rmSync(tempDir, { recursive: true, force: true });
  }
});

test('createChannelCredentials throws when the configured CA bundle is missing', () => {
  const client = createBareClient({
    protocol: ClientProtocol.GRPC,
    serverHost: 'localhost',
    serverPort: 9443,
    sessionId: 'session-1',
    userId: 'user-1',
    tls: {
      enabled: true,
      caCertPath: path.join(os.tmpdir(), `missing-${Date.now()}.crt`),
    },
  });

  assert.throws(() => client.createChannelCredentials(), /ENOENT/);
});