import { randomUUID } from 'node:crypto';

import * as grpc from '@grpc/grpc-js';

import {
  CreateApiKeyRequest,
  CreateRequestRequest,
  CreateSessionRequest,
  DeleteToolRequest,
  DeleteToolResponse,
  DrainMachineRequest,
  DrainMachineResponse,
  ExecuteToolRequest,
  GetMachineRequest,
  GetRequestChunksRequest,
  GetRequestRequest,
  GetSessionRequest,
  GetToolByIdRequest,
  GetToolByNameRequest,
  GetToolResponse,
  HealthCheckRequest,
  ListApiKeysRequest,
  ListMachinesRequest,
  ListMachinesResponse,
  ListRequestsRequest,
  ListToolsRequest,
  ListToolsResponse,
  ListUserSessionsRequest,
  Machine as ProtoMachine,
  RegisterMachineRequest,
  RevokeApiKeyRequest,
  UnregisterMachineRequest,
  Session as ProtoSession,
  ApiKey as ProtoApiKey,
  Request as ProtoRequest,
  ResumeStreamRequest,
  UnregisterMachineResponse,
  UpdateSessionRequest,
  Tool as ProtoTool,
} from '../../../src/proto/proto/service_pb';

import {
  ProviderRuntime,
  ToolplaneClient,
} from '../../../src';

import {
  MachinesServiceClient,
  RequestsServiceClient,
  SessionsServiceClient,
  ToolServiceClient,
} from '../../../src/proto/proto/service_grpc_pb';

import type { ConformanceAdapter } from '../types';

const DEFAULT_TIMEOUT_MS = 30_000;

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

function parseMaybeJSON(value: unknown): unknown {
  if (typeof value !== 'string') {
    return value;
  }

  try {
    return JSON.parse(value);
  } catch {
    return value;
  }
}

function numberValue(value: unknown, fallback: number): number {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === 'string') {
    const parsed = Number.parseInt(value, 10);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }
  return fallback;
}

function normalizeGrpcErrorCode(code: grpc.status): string {
  if (code === grpc.status.OUT_OF_RANGE) {
    return 'out_of_range';
  }

  return String(code).toLowerCase();
}

export class GrpcConformanceAdapter implements ConformanceAdapter {
  private readonly target: string;

  private readonly toolClient: ToolServiceClient;

  private readonly sessionClient: SessionsServiceClient;

  private readonly machineClient: MachinesServiceClient;

  private readonly requestsClient: RequestsServiceClient;

  private readonly providerClient: ToolplaneClient;

  private readonly runtimes = new Map<string, ProviderRuntime>();

  constructor(host: string, port: number, private readonly userId: string, private readonly apiKey: string = '') {
    this.target = `${host}:${port}`;
    const credentials = grpc.credentials.createInsecure();
    this.toolClient = new ToolServiceClient(this.target, credentials);
    this.sessionClient = new SessionsServiceClient(this.target, credentials);
    this.machineClient = new MachinesServiceClient(this.target, credentials);
    this.requestsClient = new RequestsServiceClient(this.target, credentials);
    this.providerClient = ToolplaneClient.createGRPCClient(host, port, '', userId, apiKey);
  }

  async connect(): Promise<void> {
    await Promise.all([
      this.waitForReady(this.toolClient),
      this.waitForReady(this.sessionClient),
      this.waitForReady(this.machineClient),
      this.waitForReady(this.requestsClient),
    ]);

    const request = new HealthCheckRequest();
    await this.callUnary(
      (metadata, options, callback) => this.toolClient.healthCheck(request, metadata, options, callback),
      'gRPC health check failed',
    );
  }

  async close(): Promise<void> {
    await Promise.allSettled(Array.from(this.runtimes.values()).map((runtime) => runtime.close()));
    this.runtimes.clear();

    this.toolClient.close();
    this.sessionClient.close();
    this.machineClient.close();
    this.requestsClient.close();
  }

  async createSession(request: Record<string, unknown>): Promise<string> {
    const message = new CreateSessionRequest();
    message.setUserId(String(request.user_id ?? this.userId));
    message.setName(String(request.name ?? ''));
    message.setDescription(String(request.description ?? ''));
    message.setNamespace(String(request.namespace ?? ''));
    message.setApiKey(this.apiKey);

    const response = await this.callUnary(
      (metadata, options, callback) => this.sessionClient.createSession(message, metadata, options, callback),
      'failed to create session',
    );

    const session = response.getSession();
    const sessionId = session?.getId() ?? '';
    if (!sessionId.trim()) {
      throw new Error('No session ID returned from CreateSession');
    }

    return sessionId;
  }

  async getSessionContext(sessionId: string): Promise<Record<string, unknown> | null> {
    const request = new GetSessionRequest();
    request.setSessionId(sessionId);

    const session = await this.callUnary(
      (metadata, options, callback) => this.sessionClient.getSession(request, metadata, options, callback),
      `failed to fetch session ${sessionId}`,
    );

    return session ? this.normalizeSession(session) : null;
  }

  async updateSession(sessionId: string, request: Record<string, unknown>): Promise<Record<string, unknown>> {
    const message = new UpdateSessionRequest();
    message.setSessionId(sessionId);
    message.setName(String(request.updated_name ?? ''));
    message.setDescription(String(request.updated_description ?? ''));
    message.setNamespace(String(request.updated_namespace ?? ''));

    const session = await this.callUnary(
      (metadata, options, callback) => this.sessionClient.updateSession(message, metadata, options, callback),
      `failed to update session ${sessionId}`,
    );

    return this.normalizeSession(session);
  }

  async listUserSessions(request: Record<string, unknown>): Promise<Record<string, unknown>> {
    const message = new ListUserSessionsRequest();
    message.setUserId(String(request.user_id ?? this.userId));
    message.setPageSize(numberValue(request.page_size, 10));
    message.setPageToken(numberValue(request.page_token, 0));
    message.setFilter(String(request.filter ?? ''));

    const response = await this.callUnary(
      (metadata, options, callback) => this.sessionClient.listUserSessions(message, metadata, options, callback),
      'failed to list user sessions',
    );

    return {
      sessions: response.getSessionsList().map((session) => this.normalizeSession(session)),
      nextPageToken: response.getNextPageToken(),
      totalCount: response.getTotalCount(),
    };
  }

  async registerUnaryEchoTool(sessionId: string, toolName: string, description: string): Promise<void> {
    await this.getRuntime(sessionId).registerTool({
      sessionId,
      name: toolName,
      description,
      schema: {
        type: 'object',
        properties: {
          message: { type: 'string' },
          delay_ms: { type: 'integer' },
        },
        required: ['message'],
      },
      tags: ['conformance'],
      handler: async (input) => {
        const message = typeof input.message === 'string' ? input.message : String(input.message ?? '');
        const delayMs = numberValue(input.delay_ms, 0);
        if (delayMs > 0) {
          await sleep(delayMs);
        }
        return { echo: message };
      },
    });
  }

  async registerStreamTool(sessionId: string, toolName: string, description: string): Promise<void> {
    await this.getRuntime(sessionId).registerTool({
      sessionId,
      name: toolName,
      description,
      schema: {
        type: 'object',
        properties: {
          prefix: { type: 'string' },
          count: { type: 'integer' },
        },
        required: ['prefix', 'count'],
      },
      tags: ['conformance', 'stream'],
      stream: true,
      handler: async function* (input) {
        const prefix = typeof input.prefix === 'string' && input.prefix.length > 0 ? input.prefix : 'chunk';
        const count = numberValue(input.count, 5);

        for (let index = 0; index < count; index += 1) {
          await sleep(250);
          yield `${prefix}-${index + 1}`;
        }
      },
    });
  }

  async listTools(sessionId: string): Promise<Record<string, unknown>[]> {
    const request = new ListToolsRequest();
    request.setSessionId(sessionId);

    const response = await this.callUnary<ListToolsResponse>(
      (metadata, options, callback) => this.toolClient.listTools(request, metadata, options, callback),
      `failed to list tools for session ${sessionId}`,
    );

    return response.getToolsList().map((tool) => this.normalizeTool(tool));
  }

  async getToolById(sessionId: string, toolId: string): Promise<Record<string, unknown>> {
    const request = new GetToolByIdRequest();
    request.setSessionId(sessionId);
    request.setToolId(toolId);

    const response = await this.callUnary<GetToolResponse>(
      (metadata, options, callback) => this.toolClient.getToolById(request, metadata, options, callback),
      `failed to fetch tool ${toolId}`,
    );

    const tool = response.getTool();
    if (!tool) {
      throw new Error(`No tool returned for ${toolId}`);
    }
    return this.normalizeTool(tool);
  }

  async getToolByName(sessionId: string, toolName: string): Promise<Record<string, unknown>> {
    const request = new GetToolByNameRequest();
    request.setSessionId(sessionId);
    request.setToolName(toolName);

    const response = await this.callUnary<GetToolResponse>(
      (metadata, options, callback) => this.toolClient.getToolByName(request, metadata, options, callback),
      `failed to fetch tool ${toolName}`,
    );

    const tool = response.getTool();
    if (!tool) {
      throw new Error(`No tool returned for ${toolName}`);
    }
    return this.normalizeTool(tool);
  }

  async deleteTool(sessionId: string, toolId: string): Promise<boolean> {
    const request = new DeleteToolRequest();
    request.setSessionId(sessionId);
    request.setToolId(toolId);

    const response = await this.callUnary<DeleteToolResponse>(
      (metadata, options, callback) => this.toolClient.deleteTool(request, metadata, options, callback),
      `failed to delete tool ${toolId}`,
    );

    return response.getSuccess();
  }

  async createRequest(sessionId: string, toolName: string, params: Record<string, unknown>): Promise<string> {
    const request = new CreateRequestRequest();
    request.setSessionId(sessionId);
    request.setToolName(toolName);
    request.setInput(JSON.stringify(params));

    const response = await this.callUnary(
      (metadata, options, callback) => this.requestsClient.createRequest(request, metadata, options, callback),
      `failed to create request for ${toolName}`,
    );

    const requestId = response.getId();
    if (!requestId.trim()) {
      throw new Error('No request ID returned from CreateRequest');
    }

    return requestId;
  }

  async startProviderRuntime(sessionId: string): Promise<void> {
    await this.getRuntime(sessionId).attachSession(sessionId);
    await this.getRuntime(sessionId).startInBackground();
  }

  async startRequestProcessing(sessionId: string, requestId: string): Promise<void> {
    void requestId;
    void this.getRuntime(sessionId).pollOnce();
  }

  async startStreamingRequest(sessionId: string, toolName: string, params: Record<string, unknown>): Promise<string> {
    const requestId = await this.createRequest(sessionId, toolName, params);
    await this.startProviderRuntime(sessionId);
    void this.getRuntime(sessionId).pollOnce();
    return requestId;
  }

  async getRequestStatus(sessionId: string, requestId: string): Promise<Record<string, unknown>> {
    const request = new GetRequestRequest();
    request.setSessionId(sessionId);
    request.setRequestId(requestId);

    const response = await this.callUnary(
      (metadata, options, callback) => this.requestsClient.getRequest(request, metadata, options, callback),
      `failed to fetch request ${requestId}`,
    );

    const normalized = this.normalizeRequest(response);

    try {
      const chunksRequest = new GetRequestChunksRequest();
      chunksRequest.setSessionId(sessionId);
      chunksRequest.setRequestId(requestId);
      const chunksResponse = await this.callUnary(
        (metadata, options, callback) =>
          this.requestsClient.getRequestChunks(chunksRequest, metadata, options, callback),
        `failed to fetch request chunks for ${requestId}`,
      );
      const chunks = chunksResponse.getChunksList().map((chunk) => parseMaybeJSON(chunk));
      if (chunks.length > 0) {
        normalized.streamResults = chunks;
      }
    } catch {
      // Ignore chunk lookup failures for non-streaming requests.
    }

    return normalized;
  }

  async getRequestChunksWindow(sessionId: string, requestId: string): Promise<Record<string, unknown>> {
    const request = new GetRequestChunksRequest();
    request.setSessionId(sessionId);
    request.setRequestId(requestId);

    const response = await this.callUnary(
      (metadata, options, callback) => this.requestsClient.getRequestChunks(request, metadata, options, callback),
      `failed to fetch request chunks for ${requestId}`,
    );

    return {
      chunks: response.getChunksList().map((chunk) => parseMaybeJSON(chunk)),
      startSeq: response.getStartSeq(),
      nextSeq: response.getNextSeq(),
    };
  }

  async resumeStream(requestId: string, lastSeq: number): Promise<Record<string, unknown>> {
    const request = new ResumeStreamRequest();
    request.setRequestId(requestId);
    request.setLastSeq(lastSeq);

    const stream = this.toolClient.resumeStream(request, this.createMetadata(), this.callOptions(60_000));
    const chunks: unknown[] = [];

    return new Promise<Record<string, unknown>>((resolve, reject) => {
      let settled = false;
      let finalSeq = 0;

      stream.on('data', (chunk) => {
        if (settled) {
          return;
        }

        finalSeq = chunk.getSeq();
        if (chunk.getIsFinal()) {
          settled = true;
          resolve({
            chunks,
            sawFinal: true,
            finalSeq,
            errorMessage: chunk.getError(),
          });
          return;
        }

        const value = parseMaybeJSON(chunk.getChunk());
        if (value !== null && value !== '') {
          chunks.push(value);
        }
      });

      stream.on('error', (error: grpc.ServiceError) => {
        if (settled) {
          return;
        }

        settled = true;
        resolve({
          chunks: [],
          sawFinal: false,
          finalSeq,
          errorCode: normalizeGrpcErrorCode(error.code),
          errorMessage: error.details ?? error.message,
        });
      });

      stream.on('end', () => {
        if (settled) {
          return;
        }

        reject(new Error('gRPC resume stream ended before final marker'));
      });
    });
  }

  async listRequests(sessionId: string, request: Record<string, unknown>): Promise<Record<string, unknown>[]> {
    const message = new ListRequestsRequest();
    message.setSessionId(sessionId);
    message.setStatus(String(request.list_status ?? ''));
    message.setToolName(String(request.tool_name_filter ?? ''));
    message.setLimit(numberValue(request.limit, 10));
    message.setOffset(numberValue(request.offset, 0));

    const response = await this.callUnary(
      (metadata, options, callback) => this.requestsClient.listRequests(message, metadata, options, callback),
      `failed to list requests for session ${sessionId}`,
    );

    return response.getRequestsList().map((item) => this.normalizeRequest(item));
  }

  async createApiKey(sessionId: string, name: string): Promise<Record<string, unknown>> {
    const request = new CreateApiKeyRequest();
    request.setSessionId(sessionId);
    request.setName(name);

    const response = await this.callUnary(
      (metadata, options, callback) => this.sessionClient.createApiKey(request, metadata, options, callback),
      `failed to create API key for session ${sessionId}`,
    );

    return this.normalizeApiKey(response);
  }

  async listApiKeys(sessionId: string): Promise<Record<string, unknown>[]> {
    const request = new ListApiKeysRequest();
    request.setSessionId(sessionId);

    const response = await this.callUnary(
      (metadata, options, callback) => this.sessionClient.listApiKeys(request, metadata, options, callback),
      `failed to list API keys for session ${sessionId}`,
    );

    return response.getApiKeysList().map((item) => this.normalizeApiKey(item));
  }

  async revokeApiKey(sessionId: string, keyId: string): Promise<boolean> {
    const request = new RevokeApiKeyRequest();
    request.setSessionId(sessionId);
    request.setKeyId(keyId);

    const response = await this.callUnary(
      (metadata, options, callback) => this.sessionClient.revokeApiKey(request, metadata, options, callback),
      `failed to revoke API key ${keyId}`,
    );

    return response.getSuccess();
  }

  async registerMachine(sessionId: string, request: Record<string, unknown>): Promise<Record<string, unknown>> {
    const message = new RegisterMachineRequest();
    message.setSessionId(sessionId);
    message.setMachineId(String(request.machine_id ?? randomUUID()));
    message.setSdkVersion(String(request.sdk_version ?? '1.0.0-conformance'));
    message.setSdkLanguage(String(request.sdk_language ?? 'conformance'));

    const response = await this.callUnary<ProtoMachine>(
      (metadata, options, callback) => this.machineClient.registerMachine(message, metadata, options, callback),
      `failed to register machine for session ${sessionId}`,
    );

    return this.normalizeMachine(response);
  }

  async listMachines(sessionId: string): Promise<Record<string, unknown>[]> {
    const request = new ListMachinesRequest();
    request.setSessionId(sessionId);

    const response = await this.callUnary<ListMachinesResponse>(
      (metadata, options, callback) => this.machineClient.listMachines(request, metadata, options, callback),
      `failed to list machines for session ${sessionId}`,
    );

    return response.getMachinesList().map((item) => this.normalizeMachine(item));
  }

  async getMachine(sessionId: string, machineId: string): Promise<Record<string, unknown>> {
    const request = new GetMachineRequest();
    request.setSessionId(sessionId);
    request.setMachineId(machineId);

    const response = await this.callUnary<ProtoMachine>(
      (metadata, options, callback) => this.machineClient.getMachine(request, metadata, options, callback),
      `failed to fetch machine ${machineId}`,
    );

    return this.normalizeMachine(response);
  }

  async unregisterMachine(sessionId: string, machineId: string): Promise<boolean> {
    const request = new UnregisterMachineRequest();
    request.setSessionId(sessionId);
    request.setMachineId(machineId);

    const response = await this.callUnary<UnregisterMachineResponse>(
      (metadata, options, callback) => this.machineClient.unregisterMachine(request, metadata, options, callback),
      `failed to unregister machine ${machineId}`,
    );

    return response.getSuccess();
  }

  async drainMachine(sessionId: string, machineId: string): Promise<boolean> {
    const request = new DrainMachineRequest();
    request.setSessionId(sessionId);
    request.setMachineId(machineId);

    const response = await this.callUnary<DrainMachineResponse>(
      (metadata, options, callback) => this.machineClient.drainMachine(request, metadata, options, callback),
      `failed to drain machine ${machineId}`,
    );

    return response.getDrained();
  }

  async invoke(sessionId: string, toolName: string, params: Record<string, unknown>): Promise<unknown> {
    await this.startProviderRuntime(sessionId);

    const request = new ExecuteToolRequest();
    request.setSessionId(sessionId);
    request.setToolName(toolName);
    request.setInput(JSON.stringify(params));

    const response = await this.callUnary(
      (metadata, options, callback) => this.toolClient.executeTool(request, metadata, options, callback),
      `failed to execute tool ${toolName}`,
    );

    const requestId = response.getRequestId();
    if (!requestId.trim()) {
      throw new Error('No request ID returned from ExecuteTool');
    }

    void this.getRuntime(sessionId).pollOnce();
    const status = await this.waitForRequestCompletion(sessionId, requestId);
    return status.result;
  }

  async stream(sessionId: string, toolName: string, params: Record<string, unknown>): Promise<[unknown[], boolean]> {
    await this.startProviderRuntime(sessionId);

    const request = new ExecuteToolRequest();
    request.setSessionId(sessionId);
    request.setToolName(toolName);
    request.setInput(JSON.stringify(params));

    const stream = this.toolClient.streamExecuteTool(request, this.createMetadata(), this.callOptions(60_000));
    const collected: unknown[] = [];

    return new Promise<[unknown[], boolean]>((resolve, reject) => {
      let settled = false;

      void this.getRuntime(sessionId).pollOnce();

      stream.on('data', (chunk) => {
        if (settled) {
          return;
        }

        if (chunk.getIsFinal()) {
          settled = true;
          const finalPayload = parseMaybeJSON(chunk.getChunk());
          if (chunk.getError()) {
            reject(new Error(`gRPC stream request failed: ${chunk.getError()}`));
            return;
          }

          if (Array.isArray(finalPayload) && finalPayload.length >= collected.length) {
            resolve([finalPayload, true]);
            return;
          }

          resolve([collected, true]);
          return;
        }

        const value = parseMaybeJSON(chunk.getChunk());
        if (value !== null && value !== '') {
          collected.push(value);
        }
      });

      stream.on('error', (error) => {
        if (settled) {
          return;
        }
        settled = true;
        reject(error);
      });

      stream.on('end', () => {
        if (settled) {
          return;
        }
        settled = true;
        reject(new Error('gRPC stream ended before final marker'));
      });
    });
  }

  async waitForRequestCompletion(sessionId: string, requestId: string): Promise<Record<string, unknown>> {
    const deadline = Date.now() + 60_000;
    while (Date.now() < deadline) {
      const status = await this.getRequestStatus(sessionId, requestId);
      if (status.status === 'done') {
        return status;
      }
      if (status.status === 'failure') {
        throw new Error(`Request ${requestId} failed: ${String(status.error ?? 'unknown error')}`);
      }
      await sleep(100);
    }

    throw new Error(`Timed out waiting for request ${requestId}`);
  }

  private getRuntime(sessionId: string): ProviderRuntime {
    const existing = this.runtimes.get(sessionId);
    if (existing) {
      return existing;
    }

    const runtime = this.providerClient.providerRuntime({
      pollIntervalMs: 50,
      heartbeatIntervalMs: 5_000,
      sdkVersion: '1.0.0-conformance',
    });

    this.runtimes.set(sessionId, runtime);
    return runtime;
  }

  private normalizeSession(session: ProtoSession): Record<string, unknown> {
    const createdBy = session.getCreatedBy();
    return {
      id: session.getId(),
      name: session.getName(),
      description: session.getDescription(),
      namespace: session.getNamespace(),
      created_at: session.getCreatedAt(),
      created_by: createdBy,
      user_id: createdBy,
      api_key: session.getApiKey(),
      status: 'active',
    };
  }

  private normalizeApiKey(apiKey: ProtoApiKey): Record<string, unknown> {
    return {
      id: apiKey.getId(),
      name: apiKey.getName(),
      key: apiKey.getKey(),
      session_id: apiKey.getSessionId(),
      created_at: apiKey.getCreatedAt(),
      created_by: apiKey.getCreatedBy(),
      revoked_at: apiKey.getRevokedAt(),
    };
  }

  private normalizeMachine(machine: ProtoMachine): Record<string, unknown> {
    return {
      id: machine.getId(),
      session_id: machine.getSessionId(),
      sdk_version: machine.getSdkVersion(),
      sdk_language: machine.getSdkLanguage(),
      ip: machine.getIp(),
      created_at: machine.getCreatedAt(),
      last_ping_at: machine.getLastPingAt(),
    };
  }

  private normalizeTool(tool: ProtoTool): Record<string, unknown> {
    const config: Record<string, string> = {};
    tool.getConfigMap().forEach((value: string, key: string) => {
      config[key] = value;
    });

    return {
      id: tool.getId(),
      name: tool.getName(),
      description: tool.getDescription(),
      schema: parseMaybeJSON(tool.getSchema()),
      config,
      created_at: tool.getCreatedAt(),
      last_ping_at: tool.getLastPingAt(),
      session_id: tool.getSessionId(),
      tags: tool.getTagsList(),
    };
  }

  private normalizeRequest(request: ProtoRequest): Record<string, unknown> {
    const normalized: Record<string, unknown> = {
      id: request.getId(),
      sessionId: request.getSessionId(),
      toolName: request.getToolName(),
      status: request.getStatus(),
      input: request.getInput(),
      createdAt: request.getCreatedAt(),
      updatedAt: request.getUpdatedAt(),
      executingMachineId: request.getExecutingMachineId(),
    };

    if (request.getResult()) {
      normalized.result = parseMaybeJSON(request.getResult());
    }

    if (request.getResultType()) {
      normalized.resultType = request.getResultType();
    }

    if (request.getError()) {
      normalized.error = request.getError();
    }

    const streamResults = request.getStreamResultsList().map((value) => parseMaybeJSON(value));
    if (streamResults.length > 0) {
      normalized.streamResults = streamResults;
    }

    return normalized;
  }

  private waitForReady(client: grpc.Client): Promise<void> {
    return new Promise((resolve, reject) => {
      client.waitForReady(Date.now() + DEFAULT_TIMEOUT_MS, (error) => {
        if (error) {
          reject(new Error(`failed to connect to ${this.target}: ${error.message}`));
          return;
        }
        resolve();
      });
    });
  }

  private createMetadata(): grpc.Metadata {
    const metadata = new grpc.Metadata();
    if (this.apiKey) {
      metadata.set('api_key', this.apiKey);
      metadata.set('authorization', `Bearer ${this.apiKey}`);
    }
    return metadata;
  }

  private callOptions(timeoutMs: number = DEFAULT_TIMEOUT_MS): Partial<grpc.CallOptions> {
    return {
      deadline: Date.now() + timeoutMs,
    };
  }

  private callUnary<T>(
    callFactory: (
      metadata: grpc.Metadata,
      options: Partial<grpc.CallOptions>,
      callback: (error: grpc.ServiceError | null, response: T) => void,
    ) => grpc.ClientUnaryCall,
    context: string,
  ): Promise<T> {
    return new Promise((resolve, reject) => {
      callFactory(this.createMetadata(), this.callOptions(), (error, response) => {
        if (error) {
          reject(new Error(`${context}: ${error.details || error.message}`));
          return;
        }
        resolve(response);
      });
    });
  }
}