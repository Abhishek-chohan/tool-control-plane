import { randomUUID } from 'node:crypto';

import * as grpc from '@grpc/grpc-js';

import {
  AppendRequestChunksRequest,
  ClaimRequestRequest,
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
  RegisterToolRequest,
  RevokeApiKeyRequest,
  UnregisterMachineRequest,
  Session as ProtoSession,
  ApiKey as ProtoApiKey,
  Request as ProtoRequest,
  ResumeStreamRequest,
  UnregisterMachineResponse,
  UpdateRequestRequest,
  UpdateSessionRequest,
  SubmitRequestResultRequest,
  Tool as ProtoTool,
} from '../../../src/proto/proto/service_pb';

import {
  MachinesServiceClient,
  RequestsServiceClient,
  SessionsServiceClient,
  ToolServiceClient,
} from '../../../src/proto/proto/service_grpc_pb';

import type { ConformanceAdapter } from '../types';

interface ProviderTool {
  stream: boolean;
}

interface ProviderState {
  machineId: string;
  tools: Map<string, ProviderTool>;
}

interface PendingRequestTarget {
  requestId?: string;
  toolName?: string;
  serializedInput?: string;
}

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

  private readonly providerStates = new Map<string, ProviderState>();

  private readonly requestProcessors = new Map<string, Promise<void>>();

  private readonly requestProcessorErrors = new Map<string, Error>();

  constructor(host: string, port: number, private readonly userId: string, private readonly apiKey: string = '') {
    this.target = `${host}:${port}`;
    const credentials = grpc.credentials.createInsecure();
    this.toolClient = new ToolServiceClient(this.target, credentials);
    this.sessionClient = new SessionsServiceClient(this.target, credentials);
    this.machineClient = new MachinesServiceClient(this.target, credentials);
    this.requestsClient = new RequestsServiceClient(this.target, credentials);
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
    const processors = Array.from(this.requestProcessors.values());
    this.requestProcessors.clear();
    await Promise.allSettled(processors);

    const providers = Array.from(this.providerStates.entries());
    this.providerStates.clear();

    await Promise.allSettled(
      providers.map(async ([sessionId, state]) => {
        const request = new UnregisterMachineRequest();
        request.setSessionId(sessionId);
        request.setMachineId(state.machineId);

        await this.callUnary(
          (metadata, options, callback) => this.machineClient.unregisterMachine(request, metadata, options, callback),
          `failed to unregister machine ${state.machineId}`,
        );
      }),
    );

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
    const state = await this.ensureMachine(sessionId);
    if (state.tools.has(toolName)) {
      return;
    }

    const request = new RegisterToolRequest();
    request.setSessionId(sessionId);
    request.setMachineId(state.machineId);
    request.setName(toolName);
    request.setDescription(description);
    request.setSchema(
      JSON.stringify({
        type: 'object',
        properties: {
          message: { type: 'string' },
        },
        required: ['message'],
      }),
    );
    request.setTagsList(['conformance']);

    await this.callUnary(
      (metadata, options, callback) => this.toolClient.registerTool(request, metadata, options, callback),
      `failed to register unary tool ${toolName}`,
    );

    state.tools.set(toolName, { stream: false });
  }

  async registerStreamTool(sessionId: string, toolName: string, description: string): Promise<void> {
    const state = await this.ensureMachine(sessionId);
    if (state.tools.has(toolName)) {
      return;
    }

    const request = new RegisterToolRequest();
    request.setSessionId(sessionId);
    request.setMachineId(state.machineId);
    request.setName(toolName);
    request.setDescription(description);
    request.setSchema(
      JSON.stringify({
        type: 'object',
        properties: {
          prefix: { type: 'string' },
          count: { type: 'integer' },
        },
        required: ['prefix', 'count'],
      }),
    );
    request.setTagsList(['conformance', 'stream']);

    await this.callUnary(
      (metadata, options, callback) => this.toolClient.registerTool(request, metadata, options, callback),
      `failed to register stream tool ${toolName}`,
    );

    state.tools.set(toolName, { stream: true });
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
    await this.ensureMachine(sessionId);
  }

  async startRequestProcessing(sessionId: string, requestId: string): Promise<void> {
    void this.startRequestProcessor(sessionId, requestId);
  }

  async startStreamingRequest(sessionId: string, toolName: string, params: Record<string, unknown>): Promise<string> {
    const requestId = await this.createRequest(sessionId, toolName, params);
    this.startRequestProcessor(sessionId, requestId);
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

    await this.processPendingRequests(sessionId, { requestId });
    const status = await this.waitForRequestCompletion(sessionId, requestId);
    return status.result;
  }

  async stream(sessionId: string, toolName: string, params: Record<string, unknown>): Promise<[unknown[], boolean]> {
    const request = new ExecuteToolRequest();
    const serializedInput = JSON.stringify(params);
    request.setSessionId(sessionId);
    request.setToolName(toolName);
    request.setInput(serializedInput);

    const processor = this.processPendingRequests(sessionId, {
      toolName,
      serializedInput,
    });

    const stream = this.toolClient.streamExecuteTool(request, this.createMetadata(), this.callOptions(60_000));
    const collected: unknown[] = [];

    return new Promise<[unknown[], boolean]>((resolve, reject) => {
      let settled = false;

      const rejectWithProcessor = async (error: unknown) => {
        try {
          await processor.catch(() => undefined);
        } finally {
          reject(error);
        }
      };

      stream.on('data', (chunk) => {
        if (settled) {
          return;
        }

        if (chunk.getIsFinal()) {
          settled = true;
          const finalPayload = parseMaybeJSON(chunk.getChunk());
          void processor
            .then(() => {
              if (chunk.getError()) {
                reject(new Error(`gRPC stream request failed: ${chunk.getError()}`));
                return;
              }

              if (Array.isArray(finalPayload) && finalPayload.length >= collected.length) {
                resolve([finalPayload, true]);
                return;
              }

              resolve([collected, true]);
            })
            .catch(reject);
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
        void rejectWithProcessor(error);
      });

      stream.on('end', () => {
        if (settled) {
          return;
        }
        settled = true;
        void rejectWithProcessor(new Error('gRPC stream ended before final marker'));
      });
    });
  }

  async waitForRequestCompletion(sessionId: string, requestId: string): Promise<Record<string, unknown>> {
    const deadline = Date.now() + 60_000;
    while (Date.now() < deadline) {
      const processorError = this.requestProcessorErrors.get(requestId);
      if (processorError) {
        this.requestProcessorErrors.delete(requestId);
        throw processorError;
      }

      const status = await this.getRequestStatus(sessionId, requestId);
      if (status.status === 'done') {
        await this.awaitRequestProcessor(requestId);
        return status;
      }
      if (status.status === 'failure') {
        await this.awaitRequestProcessor(requestId).catch(() => undefined);
        throw new Error(`Request ${requestId} failed: ${String(status.error ?? 'unknown error')}`);
      }
      await sleep(100);
    }

    throw new Error(`Timed out waiting for request ${requestId}`);
  }

  private startRequestProcessor(sessionId: string, requestId: string): Promise<void> {
    const existing = this.requestProcessors.get(requestId);
    if (existing) {
      return existing;
    }

    const processor = this.processPendingRequests(sessionId, { requestId }).finally(() => {
      this.requestProcessors.delete(requestId);
    });

    const trackedProcessor = processor.catch((error) => {
      this.requestProcessorErrors.set(
        requestId,
        error instanceof Error ? error : new Error(String(error)),
      );
    });

    this.requestProcessors.set(requestId, trackedProcessor);
    return trackedProcessor;
  }

  private async awaitRequestProcessor(requestId: string): Promise<void> {
    const processor = this.requestProcessors.get(requestId);
    if (processor) {
      await processor;
    }
  }

  private async ensureMachine(sessionId: string): Promise<ProviderState> {
    const existing = this.providerStates.get(sessionId);
    if (existing) {
      return existing;
    }

    const state: ProviderState = {
      machineId: randomUUID(),
      tools: new Map(),
    };

    const request = new RegisterMachineRequest();
    request.setSessionId(sessionId);
    request.setMachineId(state.machineId);
    request.setSdkVersion('1.0.0-conformance');
    request.setSdkLanguage('typescript');

    await this.callUnary(
      (metadata, options, callback) => this.machineClient.registerMachine(request, metadata, options, callback),
      `failed to register machine for session ${sessionId}`,
    );

    this.providerStates.set(sessionId, state);
    return state;
  }

  private async processPendingRequests(sessionId: string, target: PendingRequestTarget): Promise<void> {
    const state = this.providerStates.get(sessionId);
    if (!state) {
      throw new Error(`No provider state registered for session ${sessionId}`);
    }

    const deadline = Date.now() + 5_000;
    while (Date.now() < deadline) {
      const listRequest = new ListRequestsRequest();
      listRequest.setSessionId(sessionId);
      listRequest.setStatus('pending');
      listRequest.setLimit(20);

      const response = await this.callUnary(
        (metadata, options, callback) => this.requestsClient.listRequests(listRequest, metadata, options, callback),
        `failed to list pending requests for session ${sessionId}`,
      );

      const request = response.getRequestsList().find((candidate) => {
        if (target.requestId && candidate.getId() !== target.requestId) {
          return false;
        }
        if (target.toolName && candidate.getToolName() !== target.toolName) {
          return false;
        }
        if (target.serializedInput && candidate.getInput() !== target.serializedInput) {
          return false;
        }
        return true;
      });

      if (!request) {
        await sleep(100);
        continue;
      }

      const claimRequest = new ClaimRequestRequest();
      claimRequest.setSessionId(sessionId);
      claimRequest.setRequestId(request.getId());
      claimRequest.setMachineId(state.machineId);

      const claimedRequest = await this.callUnary(
        (metadata, options, callback) => this.requestsClient.claimRequest(claimRequest, metadata, options, callback),
        `failed to claim request ${request.getId()}`,
      );

      const runningRequest = new UpdateRequestRequest();
      runningRequest.setSessionId(sessionId);
      runningRequest.setRequestId(claimedRequest.getId());
      runningRequest.setStatus('running');
      await this.callUnary(
        (metadata, options, callback) => this.requestsClient.updateRequest(runningRequest, metadata, options, callback),
        `failed to mark request ${claimedRequest.getId()} as running`,
      );

      const tool = state.tools.get(claimedRequest.getToolName());
      const params = this.parseRequestParams(claimedRequest.getInput());

      if (!tool) {
        await this.submitResult(sessionId, claimedRequest.getId(), JSON.stringify({ error: 'tool not found' }), 'rejection');
        return;
      }

      if (tool.stream) {
        await this.fulfillStreamingRequest(sessionId, claimedRequest.getId(), params);
      } else {
        await this.fulfillUnaryRequest(sessionId, claimedRequest.getId(), params);
      }
      return;
    }

    throw new Error(`Timed out waiting to claim request for session ${sessionId}`);
  }

  private async fulfillUnaryRequest(sessionId: string, requestId: string, params: Record<string, unknown>): Promise<void> {
    const message = typeof params.message === 'string' ? params.message : String(params.message ?? '');
    const delayMs = numberValue(params.delay_ms, 0);
    if (delayMs > 0) {
      await sleep(delayMs);
    }
    await this.submitResult(sessionId, requestId, JSON.stringify({ echo: message }), 'resolution');
  }

  private async fulfillStreamingRequest(sessionId: string, requestId: string, params: Record<string, unknown>): Promise<void> {
    const prefix = typeof params.prefix === 'string' && params.prefix.length > 0 ? params.prefix : 'chunk';
    const count = numberValue(params.count, 5);
    const chunks: string[] = [];

    const updateRequest = new UpdateRequestRequest();
    updateRequest.setSessionId(sessionId);
    updateRequest.setRequestId(requestId);
    updateRequest.setResultType('streaming');
    await this.callUnary(
      (metadata, options, callback) => this.requestsClient.updateRequest(updateRequest, metadata, options, callback),
      `failed to set streaming mode for request ${requestId}`,
    );

    for (let index = 0; index < count; index += 1) {
      const value = `${prefix}-${index + 1}`;
      chunks.push(value);

      const appendRequest = new AppendRequestChunksRequest();
      appendRequest.setSessionId(sessionId);
      appendRequest.setRequestId(requestId);
      appendRequest.setChunksList([value]);
      appendRequest.setResultType('streaming');
      await this.callUnary(
        (metadata, options, callback) =>
          this.requestsClient.appendRequestChunks(appendRequest, metadata, options, callback),
        `failed to append chunk for request ${requestId}`,
      );

      await sleep(250);
    }

    await this.submitResult(sessionId, requestId, JSON.stringify(chunks), 'resolution');
  }

  private async submitResult(
    sessionId: string,
    requestId: string,
    result: string,
    resultType: string,
  ): Promise<void> {
    const request = new SubmitRequestResultRequest();
    request.setSessionId(sessionId);
    request.setRequestId(requestId);
    request.setResult(result);
    request.setResultType(resultType);
    request.getMetaMap().set('handled_by', 'typescript-conformance');

    await this.callUnary(
      (metadata, options, callback) => this.requestsClient.submitRequestResult(request, metadata, options, callback),
      `failed to submit result for request ${requestId}`,
    );
  }

  private parseRequestParams(input: string): Record<string, unknown> {
    const parsed = parseMaybeJSON(input);
    return typeof parsed === 'object' && parsed !== null ? { ...parsed } : {};
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