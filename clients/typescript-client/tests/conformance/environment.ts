/**
 * Shared-fixture conformance bootstrap — TypeScript mirror.
 *
 * This module mirrors the Python reference implementation at
 *   clients/python-client/tests/conformance/conftest.py
 * and must behave identically: same env defaults, same auto-boot
 * sequence (find free ports → go run server → go run proxy → wait for
 * readiness), and same teardown. If you change the Python bootstrap,
 * apply the equivalent change here.
 *
 * The env contract matches the release-gate contract documented in
 *   server/docs/release-gate.md
 */
import { spawn, type ChildProcess } from 'node:child_process';
import { mkdir } from 'node:fs/promises';
import { closeSync, openSync } from 'node:fs';
import net from 'node:net';
import path from 'node:path';

import axios from 'axios';

const DEFAULT_API_KEY = 'toolplane-conformance-fixture-key';
const BOOTSTRAP_READINESS_TIMEOUT_MS = 60_000;

interface LoggedProcess {
  child: ChildProcess;
  logFd: number;
}

export interface ConformanceEnvironment {
  cleanup(): Promise<void>;
}

function repoRoot(): string {
  return path.resolve(process.cwd(), '../..');
}

function serverRoot(): string {
  return path.resolve(process.cwd(), '../../server');
}

async function findFreePort(): Promise<number> {
  return new Promise((resolve, reject) => {
    const server = net.createServer();
    server.unref();
    server.once('error', reject);
    server.listen(0, '127.0.0.1', () => {
      const address = server.address();
      if (!address || typeof address === 'string') {
        reject(new Error('failed to allocate free port'));
        return;
      }
      server.close((error) => {
        if (error) {
          reject(error);
          return;
        }
        resolve(address.port);
      });
    });
  });
}

async function waitForTCP(host: string, port: number, timeoutMs: number): Promise<boolean> {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    const connected = await new Promise<boolean>((resolve) => {
      const socket = net.createConnection({ host, port });
      socket.setTimeout(500);
      socket.once('connect', () => {
        socket.destroy();
        resolve(true);
      });
      socket.once('timeout', () => {
        socket.destroy();
        resolve(false);
      });
      socket.once('error', () => {
        socket.destroy();
        resolve(false);
      });
    });

    if (connected) {
      return true;
    }

    await sleep(200);
  }

  return false;
}

async function waitForHTTPHealth(url: string, timeoutMs: number): Promise<boolean> {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    try {
      const response = await axios.get(url, {
        timeout: 1000,
        validateStatus: () => true,
      });
      if (response.status === 200) {
        return true;
      }
    } catch {
      // Ignore readiness probe failures until timeout.
    }

    await sleep(200);
  }

  return false;
}

function startLoggedProcess(command: string, args: string[], cwd: string, logPath: string): LoggedProcess {
  const logFd = openSync(logPath, 'w');
  const child = spawn(command, args, {
    cwd,
    env: { ...process.env },
    stdio: ['ignore', logFd, logFd],
  });

  return { child, logFd };
}

async function waitForExit(child: ChildProcess, timeoutMs: number): Promise<boolean> {
  if (child.exitCode !== null) {
    return true;
  }

  return new Promise((resolve) => {
    const timer = setTimeout(() => {
      cleanup();
      resolve(false);
    }, timeoutMs);

    const onExit = () => {
      cleanup();
      resolve(true);
    };

    const cleanup = () => {
      clearTimeout(timer);
      child.removeListener('exit', onExit);
      child.removeListener('error', onExit);
    };

    child.once('exit', onExit);
    child.once('error', onExit);
  });
}

async function terminateProcess(process: LoggedProcess): Promise<void> {
  const { child, logFd } = process;

  if (child.exitCode === null) {
    child.kill('SIGTERM');
    const exited = await waitForExit(child, 5000);
    if (!exited && child.exitCode === null) {
      child.kill('SIGKILL');
      await waitForExit(child, 5000);
    }
  }

  closeSync(logFd);
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

export async function startConformanceEnvironment(): Promise<ConformanceEnvironment> {
  const autoBoot = !['0', 'false', 'no'].includes((process.env.TOOLPLANE_CONFORMANCE_AUTO_BOOT ?? '1').trim().toLowerCase());

  if (!process.env.TOOLPLANE_CONFORMANCE_API_KEY) {
    process.env.TOOLPLANE_CONFORMANCE_API_KEY = DEFAULT_API_KEY;
  }

  process.env.TOOLPLANE_ENV_MODE ??= 'development';
  process.env.TOOLPLANE_AUTH_MODE ??= 'fixed';
  process.env.TOOLPLANE_AUTH_FIXED_API_KEY ??= process.env.TOOLPLANE_CONFORMANCE_API_KEY;
  process.env.TOOLPLANE_STORAGE_MODE ??= 'memory';
  process.env.TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND ??= '1';

  if (!autoBoot) {
    process.env.TOOLPLANE_CONFORMANCE_GRPC_HOST ??= 'localhost';
    process.env.TOOLPLANE_CONFORMANCE_GRPC_PORT ??= '50051';
    process.env.TOOLPLANE_CONFORMANCE_HTTP_HOST ??= 'localhost';
    process.env.TOOLPLANE_CONFORMANCE_HTTP_PORT ??= '8080';
    process.env.TOOLPLANE_CONFORMANCE_USER_ID ??= 'conformance-user';

    return {
      cleanup: async () => {},
    };
  }

  const grpcPort = await findFreePort();
  const httpPort = await findFreePort();

  process.env.TOOLPLANE_CONFORMANCE_GRPC_HOST = 'localhost';
  process.env.TOOLPLANE_CONFORMANCE_GRPC_PORT = String(grpcPort);
  process.env.TOOLPLANE_CONFORMANCE_HTTP_HOST = 'localhost';
  process.env.TOOLPLANE_CONFORMANCE_HTTP_PORT = String(httpPort);
  process.env.TOOLPLANE_CONFORMANCE_USER_ID ??= 'conformance-user';

  const logDir = path.join(repoRoot(), '.tmp', 'conformance-logs');
  await mkdir(logDir, { recursive: true });

  const serverLogPath = path.join(logDir, 'ts-grpc-server.log');
  const proxyLogPath = path.join(logDir, 'ts-http-proxy.log');

  const serverProcess = startLoggedProcess('go', ['run', './cmd/server', '--port', String(grpcPort)], serverRoot(), serverLogPath);

  if (!(await waitForTCP('127.0.0.1', grpcPort, BOOTSTRAP_READINESS_TIMEOUT_MS))) {
    await terminateProcess(serverProcess);
    throw new Error(`Conformance bootstrap failed: gRPC server did not become ready on ${grpcPort}. Check ${serverLogPath}`);
  }

  const proxyProcess = startLoggedProcess(
    'go',
    ['run', './cmd/proxy', '--listen', `:${httpPort}`, '--backend', `localhost:${grpcPort}`],
    serverRoot(),
    proxyLogPath,
  );

  if (!(await waitForHTTPHealth(`http://127.0.0.1:${httpPort}/health`, BOOTSTRAP_READINESS_TIMEOUT_MS))) {
    await terminateProcess(proxyProcess);
    await terminateProcess(serverProcess);
    throw new Error(`Conformance bootstrap failed: HTTP gateway did not become ready on ${httpPort}. Check ${proxyLogPath}`);
  }

  return {
    cleanup: async () => {
      await terminateProcess(proxyProcess);
      await terminateProcess(serverProcess);
    },
  };
}