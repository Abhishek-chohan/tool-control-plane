import { randomUUID } from 'node:crypto';

import axios, { type AxiosInstance } from 'axios';

import type { ConformanceAdapter } from '../types';

interface ProviderTool {
  stream: boolean;
}

interface ProviderState {
  machineId: string;
  tools: Map<string, ProviderTool>;
}

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

function normalizeGatewayErrorCode(payload: unknown): string {
  if (typeof payload === 'string') {
    try {
      return normalizeGatewayErrorCode(JSON.parse(payload));
    } catch {
      const lowered = payload.toLowerCase();
      if (
        lowered.includes('out_of_range') ||
        lowered.includes('out of range') ||
        lowered.includes('outofrange')
      ) {
        return 'out_of_range';
      }
      return '';
    }
  }

  if (typeof payload === 'object' && payload !== null) {
    const payloadObject = payload as Record<string, unknown>;
    if (payloadObject.error !== undefined) {
      const nested = normalizeGatewayErrorCode(payloadObject.error);
      if (nested) {
        return nested;
      }
    }

    const code = payloadObject.code;
    if (code === 11) {
      return 'out_of_range';
    }
    if (typeof code === 'string') {
      const normalized = code.trim().toLowerCase().replace(/[-\s]+/g, '_');
      if (normalized === 'outofrange') {
        return 'out_of_range';
      }
      return normalized;
    }
  }

  return '';
}

function readStreamText(stream: NodeJS.ReadableStream, timeoutMs = 2_000): Promise<string> {
  return new Promise((resolve, reject) => {
    let buffer = '';

    const cleanup = () => {
      clearTimeout(timer);
      stream.removeListener('data', onData);
      stream.removeListener('end', onEnd);
      stream.removeListener('error', onError);
      if ('destroy' in stream && typeof stream.destroy === 'function' && !stream.destroyed) {
        stream.destroy();
      }
    };

    const onData = (chunk: unknown) => {
      buffer += String(chunk);
    };

    const onEnd = () => {
      cleanup();
      resolve(buffer);
    };

    const onError = (error: unknown) => {
      cleanup();
      reject(error);
    };

    const timer = setTimeout(() => {
      cleanup();
      resolve(buffer);
    }, timeoutMs);

    stream.on('data', onData);
    stream.on('end', onEnd);
    stream.on('error', onError);
  });
}

export class HttpConformanceAdapter implements ConformanceAdapter {
  private readonly client: AxiosInstance;

  private readonly providerStates = new Map<string, ProviderState>();

  private readonly requestProcessors = new Map<string, Promise<void>>();

  private readonly requestProcessorErrors = new Map<string, Error>();

  private readonly requestSessions = new Map<string, string>();

  constructor(host: string, port: number, private readonly userId: string, private readonly apiKey: string = '') {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'User-Agent': 'Toolplane-TypeScript-Conformance/1.0.0',
    };

    if (apiKey) {
      headers.api_key = apiKey;
      headers.authorization = `Bearer ${apiKey}`;
    }

    this.client = axios.create({
      baseURL: `http://${host}:${port}`,
      timeout: 30_000,
      headers,
    });
  }

  async connect(): Promise<void> {
    await this.post('api/HealthCheck');
  }

  async close(): Promise<void> {
    const processors = Array.from(this.requestProcessors.values());
    this.requestProcessors.clear();
    this.requestSessions.clear();
    await Promise.allSettled(processors);

    const providers = Array.from(this.providerStates.entries());
    this.providerStates.clear();

    await Promise.allSettled(
      providers.map(async ([sessionId, state]) => {
        await this.post('api/UnregisterMachine', {
          sessionId,
          machineId: state.machineId,
        });
      }),
    );
  }

  async createSession(request: Record<string, unknown>): Promise<string> {
    const response = await this.post<Record<string, unknown>>('api/CreateSession', {
      userId: String(request.user_id ?? this.userId),
      name: String(request.name ?? ''),
      description: String(request.description ?? ''),
      namespace: String(request.namespace ?? ''),
      apiKey: this.apiKey,
    });

    const session = this.unwrapObject(response.session ?? response);
    const sessionId = session.id;
    if (typeof sessionId !== 'string' || sessionId.trim().length === 0) {
      throw new Error('No session ID returned from CreateSession');
    }
    return sessionId;
  }

  async getSessionContext(sessionId: string): Promise<Record<string, unknown> | null> {
    const response = await this.post<Record<string, unknown>>('api/GetSession', { sessionId });
    const session = this.unwrapObject(response.session ?? response);
    return Object.keys(session).length > 0 ? this.normalizeSession(session) : null;
  }

  async updateSession(sessionId: string, request: Record<string, unknown>): Promise<Record<string, unknown>> {
    const response = await this.post<Record<string, unknown>>('api/UpdateSession', {
      sessionId,
      name: String(request.updated_name ?? ''),
      description: String(request.updated_description ?? ''),
      namespace: String(request.updated_namespace ?? ''),
    });

    return this.normalizeSession(this.unwrapObject(response.session ?? response));
  }

  async listUserSessions(request: Record<string, unknown>): Promise<Record<string, unknown>> {
    const response = await this.post<Record<string, unknown>>('api/ListUserSessions', {
      userId: String(request.user_id ?? this.userId),
      pageSize: numberValue(request.page_size, 10),
      pageToken: numberValue(request.page_token, 0),
      filter: String(request.filter ?? ''),
    });

    const sessions = Array.isArray(response.sessions) ? response.sessions : [];
    return {
      ...response,
      sessions: sessions
        .filter((entry): entry is Record<string, unknown> => typeof entry === 'object' && entry !== null)
        .map((entry) => this.normalizeSession(entry)),
    };
  }

  async registerUnaryEchoTool(sessionId: string, toolName: string, description: string): Promise<void> {
    const state = await this.ensureMachine(sessionId);
    if (state.tools.has(toolName)) {
      return;
    }

    await this.post('api/RegisterTool', {
      sessionId,
      machineId: state.machineId,
      name: toolName,
      description,
      schema: JSON.stringify({
        type: 'object',
        properties: {
          message: { type: 'string' },
        },
        required: ['message'],
      }),
      config: {},
      tags: ['conformance'],
    });

    state.tools.set(toolName, { stream: false });
  }

  async registerStreamTool(sessionId: string, toolName: string, description: string): Promise<void> {
    const state = await this.ensureMachine(sessionId);
    if (state.tools.has(toolName)) {
      return;
    }

    await this.post('api/RegisterTool', {
      sessionId,
      machineId: state.machineId,
      name: toolName,
      description,
      schema: JSON.stringify({
        type: 'object',
        properties: {
          prefix: { type: 'string' },
          count: { type: 'integer' },
        },
        required: ['prefix', 'count'],
      }),
      config: {},
      tags: ['conformance', 'stream'],
    });

    state.tools.set(toolName, { stream: true });
  }

  async listTools(sessionId: string): Promise<Record<string, unknown>[]> {
    const response = await this.post<Record<string, unknown>>('api/ListTools', { sessionId });
    const tools = Array.isArray(response.tools) ? response.tools : [];
    return tools
      .filter((entry): entry is Record<string, unknown> => typeof entry === 'object' && entry !== null)
      .map((entry) => this.normalizeTool(entry));
  }

  async getToolById(sessionId: string, toolId: string): Promise<Record<string, unknown>> {
    const response = await this.post<Record<string, unknown>>('api/GetToolById', { sessionId, toolId });
    return this.normalizeTool(this.unwrapObject(response.tool ?? response));
  }

  async getToolByName(sessionId: string, toolName: string): Promise<Record<string, unknown>> {
    const response = await this.post<Record<string, unknown>>('api/GetToolByName', { sessionId, toolName });
    return this.normalizeTool(this.unwrapObject(response.tool ?? response));
  }

  async deleteTool(sessionId: string, toolId: string): Promise<boolean> {
    const response = await this.post<Record<string, unknown>>('api/DeleteTool', { sessionId, toolId });
    return response.success === true;
  }

  async createRequest(sessionId: string, toolName: string, params: Record<string, unknown>): Promise<string> {
    const response = await this.post<Record<string, unknown>>('api/CreateRequest', {
      sessionId,
      toolName,
      input: JSON.stringify(params),
    });

    const request = this.unwrapObject(response.request ?? response);
    const requestId = request.id;
    if (typeof requestId !== 'string' || requestId.trim().length === 0) {
      throw new Error('No request ID returned from CreateRequest');
    }
    this.requestSessions.set(requestId, sessionId);
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
    const response = await this.post<Record<string, unknown>>('api/GetRequest', {
      sessionId,
      requestId,
    });

    const normalized = this.normalizeRequest(this.unwrapObject(response.request ?? response));

    try {
      const chunkResponse = await this.post<Record<string, unknown>>('api/GetRequestChunks', {
        sessionId,
        requestId,
      });
      const chunks = Array.isArray(chunkResponse.chunks) ? chunkResponse.chunks : [];
      if (chunks.length > 0) {
        normalized.streamResults = chunks;
      }
    } catch {
      // Ignore chunk fetch failures for non-streaming requests.
    }

    return normalized;
  }

  async getRequestChunksWindow(sessionId: string, requestId: string): Promise<Record<string, unknown>> {
    const response = await this.post<Record<string, unknown>>('api/GetRequestChunks', {
      sessionId,
      requestId,
    });

    return {
      chunks: Array.isArray(response.chunks) ? response.chunks.map((chunk) => parseMaybeJSON(chunk)) : [],
      startSeq: numberValue(response.startSeq ?? response.start_seq, 0),
      nextSeq: numberValue(response.nextSeq ?? response.next_seq, 0),
    };
  }

  async resumeStream(requestId: string, lastSeq: number): Promise<Record<string, unknown>> {
    const response = await this.client.post<NodeJS.ReadableStream>(
      '/api/ResumeStream',
      { requestId, lastSeq },
      {
        responseType: 'stream',
        timeout: 60_000,
        validateStatus: () => true,
      },
    );

    const stream = response.data as NodeJS.ReadableStream;
    if (response.status !== 200) {
      const body = await readStreamText(stream);
      if ('destroy' in stream && typeof stream.destroy === 'function') {
        stream.destroy();
      }
      return {
        chunks: [],
        sawFinal: false,
        finalSeq: 0,
        errorCode: normalizeGatewayErrorCode(body),
        errorMessage: body || `HTTP ${response.status}`,
      };
    }

    return new Promise<Record<string, unknown>>((resolve, reject) => {
      let buffer = '';
      let settled = false;
      let finalSeq = 0;
      const chunks: unknown[] = [];

      const cleanup = () => {
        stream.removeAllListeners('data');
        stream.removeAllListeners('end');
        stream.removeAllListeners('error');
        if ('destroy' in stream && typeof stream.destroy === 'function' && !stream.destroyed) {
          stream.destroy();
        }
      };

      const finish = (result: Record<string, unknown>) => {
        if (settled) {
          return;
        }
        settled = true;
        cleanup();
        resolve(result);
      };

      const fail = (error: Error) => {
        if (settled) {
          return;
        }
        settled = true;
        cleanup();
        reject(error);
      };

      const handleLine = (line: string) => {
        if (!line.trim()) {
          return;
        }

        let chunk: Record<string, unknown>;
        try {
          chunk = JSON.parse(line) as Record<string, unknown>;
        } catch {
          return;
        }

        if (typeof chunk.result === 'object' && chunk.result !== null) {
          chunk = chunk.result as Record<string, unknown>;
        }

        finalSeq = numberValue(chunk.seq, 0);
        if (chunk.isFinal === true) {
          const errorMessage = String(chunk.error ?? '');
          finish({
            chunks,
            sawFinal: true,
            finalSeq,
            errorCode: normalizeGatewayErrorCode(errorMessage),
            errorMessage,
          });
          return;
        }

        const value = parseMaybeJSON(chunk.chunk);
        if (value !== null && value !== '') {
          chunks.push(value);
        }
      };

      stream.on('data', (data) => {
        buffer += data.toString();
        let newlineIndex = buffer.indexOf('\n');
        while (newlineIndex >= 0) {
          const line = buffer.slice(0, newlineIndex);
          buffer = buffer.slice(newlineIndex + 1);
          handleLine(line);
          newlineIndex = buffer.indexOf('\n');
        }
      });

      stream.on('end', () => {
        if (buffer.trim()) {
          handleLine(buffer.trim());
        }
        if (!settled) {
          void this.resumeFromChunkWindow(requestId, lastSeq)
            .then((fallback) => {
              if (fallback) {
                finish(fallback);
                return;
              }

              fail(new Error('HTTP resume stream ended before final marker'));
            })
            .catch((error) => {
              fail(error instanceof Error ? error : new Error(String(error)));
            });
        }
      });

      stream.on('error', (error) => {
        fail(error instanceof Error ? error : new Error(String(error)));
      });
    });
  }

  async listRequests(sessionId: string, request: Record<string, unknown>): Promise<Record<string, unknown>[]> {
    const response = await this.post<Record<string, unknown>>('api/ListRequests', {
      sessionId,
      status: String(request.list_status ?? ''),
      toolName: String(request.tool_name_filter ?? ''),
      limit: numberValue(request.limit, 10),
      offset: numberValue(request.offset, 0),
    });

    const requests = Array.isArray(response.requests) ? response.requests : [];
    return requests
      .filter((entry): entry is Record<string, unknown> => typeof entry === 'object' && entry !== null)
      .map((entry) => this.normalizeRequest(entry));
  }

  async createApiKey(sessionId: string, name: string): Promise<Record<string, unknown>> {
    const response = await this.post<Record<string, unknown>>('api/CreateApiKey', {
      sessionId,
      name,
    });
    return this.normalizeApiKey(this.unwrapObject(response.apiKey ?? response));
  }

  async listApiKeys(sessionId: string): Promise<Record<string, unknown>[]> {
    const response = await this.post<Record<string, unknown>>('api/ListApiKeys', { sessionId });
    const apiKeys = Array.isArray(response.apiKeys)
      ? response.apiKeys
      : Array.isArray(response.api_keys)
        ? response.api_keys
        : [];

    return apiKeys
      .filter((entry): entry is Record<string, unknown> => typeof entry === 'object' && entry !== null)
      .map((entry) => this.normalizeApiKey(entry));
  }

  async revokeApiKey(sessionId: string, keyId: string): Promise<boolean> {
    const response = await this.post<Record<string, unknown>>('api/RevokeApiKey', {
      sessionId,
      keyId,
    });
    return response.success === true;
  }

  async registerMachine(sessionId: string, request: Record<string, unknown>): Promise<Record<string, unknown>> {
    const response = await this.post<Record<string, unknown>>('api/RegisterMachine', {
      sessionId,
      machineId: String(request.machine_id ?? randomUUID()),
      sdkVersion: String(request.sdk_version ?? '1.0.0-conformance'),
      sdkLanguage: String(request.sdk_language ?? 'conformance'),
      tools: [],
    });

    return this.normalizeMachine(this.unwrapObject(response.machine ?? response));
  }

  async listMachines(sessionId: string): Promise<Record<string, unknown>[]> {
    const response = await this.post<Record<string, unknown>>('api/ListMachines', { sessionId });
    const machines = Array.isArray(response.machines) ? response.machines : [];

    return machines
      .filter((entry): entry is Record<string, unknown> => typeof entry === 'object' && entry !== null)
      .map((entry) => this.normalizeMachine(entry));
  }

  async getMachine(sessionId: string, machineId: string): Promise<Record<string, unknown>> {
    const response = await this.post<Record<string, unknown>>('api/GetMachine', {
      sessionId,
      machineId,
    });

    return this.normalizeMachine(this.unwrapObject(response.machine ?? response));
  }

  async unregisterMachine(sessionId: string, machineId: string): Promise<boolean> {
    const response = await this.post<Record<string, unknown>>('api/UnregisterMachine', {
      sessionId,
      machineId,
    });
    return response.success === true;
  }

  async drainMachine(sessionId: string, machineId: string): Promise<boolean> {
    const response = await this.post<Record<string, unknown>>('api/DrainMachine', {
      sessionId,
      machineId,
    });
    return response.drained === true;
  }

  async invoke(sessionId: string, toolName: string, params: Record<string, unknown>): Promise<unknown> {
    const response = await this.post<Record<string, unknown>>('api/ExecuteTool', {
      sessionId,
      toolName,
      input: JSON.stringify(params),
    });

    const requestId = this.extractRequestId(response);
    await this.processPendingRequests(sessionId, requestId);
    const status = await this.waitForRequestCompletion(sessionId, requestId);
    return status.result;
  }

  async stream(sessionId: string, toolName: string, params: Record<string, unknown>): Promise<[unknown[], boolean]> {
    const response = await this.post<Record<string, unknown>>('api/ExecuteTool', {
      sessionId,
      toolName,
      input: JSON.stringify(params),
    });

    const requestId = this.extractRequestId(response);
    const processor = this.processPendingRequests(sessionId, requestId);
    const collected: unknown[] = [];
    let lastCount = 0;
    const deadline = Date.now() + 60_000;

    while (Date.now() < deadline) {
      const status = await this.getRequestStatus(sessionId, requestId);
      const streamResults = Array.isArray(status.streamResults) ? status.streamResults : [];
      if (streamResults.length > lastCount) {
        for (const value of streamResults.slice(lastCount)) {
          if (value !== null && value !== '') {
            collected.push(value);
          }
        }
        lastCount = streamResults.length;
      }

      if (status.status === 'done') {
        await processor;
        if (Array.isArray(status.result) && status.result.length > collected.length) {
          return [status.result, true];
        }
        return [collected, true];
      }

      if (status.status === 'failure') {
        await processor.catch(() => undefined);
        throw new Error(`HTTP stream request failed: ${String(status.error ?? 'unknown error')}`);
      }

      await sleep(100);
    }

    await processor.catch(() => undefined);
    throw new Error('HTTP stream request timed out');
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

  private async post<T>(path: string, payload: Record<string, unknown> = {}): Promise<T> {
    const response = await this.client.post<T>(`/${path}`, payload);
    return response.data;
  }

  private startRequestProcessor(sessionId: string, requestId: string): Promise<void> {
    const existing = this.requestProcessors.get(requestId);
    if (existing) {
      return existing;
    }

    const processor = this.processPendingRequests(sessionId, requestId).finally(() => {
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

  private async resumeFromChunkWindow(requestId: string, lastSeq: number): Promise<Record<string, unknown> | null> {
    const sessionId = this.requestSessions.get(requestId);
    if (!sessionId) {
      return null;
    }

    const fallback = await this.getRequestChunksWindow(sessionId, requestId);
    const startSeq = numberValue(fallback.startSeq, 0);
    const nextSeq = numberValue(fallback.nextSeq, 0);
    const chunks = Array.isArray(fallback.chunks) ? fallback.chunks : [];
    const startIndex = Math.max(lastSeq - startSeq + 1, 0);

    return {
      chunks: chunks.slice(startIndex),
      sawFinal: true,
      finalSeq: nextSeq,
      errorMessage: '',
    };
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

    await this.post('api/RegisterMachine', {
      sessionId,
      machineId: state.machineId,
      sdkVersion: '1.0.0-conformance',
      sdkLanguage: 'typescript',
      tools: [],
    });

    this.providerStates.set(sessionId, state);
    return state;
  }

  private async processPendingRequests(sessionId: string, targetRequestId: string): Promise<void> {
    const state = this.providerStates.get(sessionId);
    if (!state) {
      throw new Error(`No provider state registered for session ${sessionId}`);
    }

    const deadline = Date.now() + 5_000;
    while (Date.now() < deadline) {
      const response = await this.post<Record<string, unknown>>('api/ListRequests', {
        sessionId,
        status: 'pending',
        limit: 20,
      });

      const requests = Array.isArray(response.requests) ? response.requests : [];
      const target = requests
        .filter((entry): entry is Record<string, unknown> => typeof entry === 'object' && entry !== null)
        .find((entry) => entry.id === targetRequestId);

      if (!target) {
        await sleep(100);
        continue;
      }

      await this.post('api/ClaimRequest', {
        sessionId,
        requestId: targetRequestId,
        machineId: state.machineId,
      });

      await this.post('api/UpdateRequest', {
        sessionId,
        requestId: targetRequestId,
        status: 'running',
      });

      const normalized = this.normalizeRequest(target);
      const tool = state.tools.get(String(normalized.toolName ?? ''));
      const params = this.parseRequestParams(normalized.input);

      if (!tool) {
        await this.submitResult(sessionId, targetRequestId, JSON.stringify({ error: 'tool not found' }), 'rejection');
        return;
      }

      if (tool.stream) {
        await this.fulfillStreamingRequest(sessionId, targetRequestId, params);
      } else {
        await this.fulfillUnaryRequest(sessionId, targetRequestId, params);
      }
      return;
    }

    throw new Error(`Timed out waiting to claim request ${targetRequestId}`);
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

    for (let index = 0; index < count; index += 1) {
      const value = `${prefix}-${index + 1}`;
      chunks.push(value);

      await this.post('api/AppendRequestChunks', {
        sessionId,
        requestId,
        chunks: [value],
        resultType: 'streaming',
      });

      await sleep(250);
    }

    await this.submitResult(sessionId, requestId, JSON.stringify(chunks), 'resolution');
  }

  private async submitResult(sessionId: string, requestId: string, result: string, resultType: string): Promise<void> {
    await this.post('api/SubmitRequestResult', {
      sessionId,
      requestId,
      result,
      resultType,
      meta: {
        handled_by: 'typescript-conformance',
      },
    });
  }

  private extractRequestId(response: Record<string, unknown>): string {
    const requestId = response.requestId ?? response.request_id;
    if (typeof requestId !== 'string' || requestId.trim().length === 0) {
      throw new Error('No request ID returned from execution response');
    }
    return requestId;
  }

  private parseRequestParams(input: unknown): Record<string, unknown> {
    const parsed = parseMaybeJSON(input);
    return typeof parsed === 'object' && parsed !== null ? { ...parsed } : {};
  }

  private normalizeSession(session: Record<string, unknown>): Record<string, unknown> {
    const createdBy = session.createdBy ?? session.created_by ?? '';
    return {
      id: session.id ?? '',
      name: session.name ?? '',
      description: session.description ?? '',
      namespace: session.namespace ?? '',
      created_at: session.createdAt ?? session.created_at ?? '',
      created_by: createdBy,
      user_id: createdBy,
      api_key: session.apiKey ?? session.api_key ?? '',
      status: 'active',
    };
  }

  private normalizeApiKey(apiKey: Record<string, unknown>): Record<string, unknown> {
    return {
      id: apiKey.id ?? '',
      name: apiKey.name ?? '',
      key: apiKey.key ?? '',
      session_id: apiKey.sessionId ?? apiKey.session_id ?? '',
      created_at: apiKey.createdAt ?? apiKey.created_at ?? '',
      created_by: apiKey.createdBy ?? apiKey.created_by ?? '',
      revoked_at: apiKey.revokedAt ?? apiKey.revoked_at ?? '',
    };
  }

  private normalizeMachine(machine: Record<string, unknown>): Record<string, unknown> {
    return {
      id: machine.id ?? '',
      session_id: machine.sessionId ?? machine.session_id ?? '',
      sdk_version: machine.sdkVersion ?? machine.sdk_version ?? '',
      sdk_language: machine.sdkLanguage ?? machine.sdk_language ?? '',
      ip: machine.ip ?? '',
      created_at: machine.createdAt ?? machine.created_at ?? '',
      last_ping_at: machine.lastPingAt ?? machine.last_ping_at ?? '',
    };
  }

  private normalizeTool(tool: Record<string, unknown>): Record<string, unknown> {
    return {
      id: tool.id ?? '',
      name: tool.name ?? '',
      description: tool.description ?? '',
      schema: parseMaybeJSON(tool.schema ?? ''),
      config: typeof tool.config === 'object' && tool.config !== null ? tool.config : {},
      created_at: tool.createdAt ?? tool.created_at ?? '',
      last_ping_at: tool.lastPingAt ?? tool.last_ping_at ?? '',
      session_id: tool.sessionId ?? tool.session_id ?? '',
      tags: Array.isArray(tool.tags) ? tool.tags : [],
    };
  }

  private normalizeRequest(request: Record<string, unknown>): Record<string, unknown> {
    const normalized: Record<string, unknown> = {
      id: request.id ?? '',
      sessionId: request.sessionId ?? request.session_id ?? '',
      toolName: request.toolName ?? request.tool_name ?? '',
      status: request.status ?? '',
      input: request.input ?? '',
      createdAt: request.createdAt ?? request.created_at ?? '',
      updatedAt: request.updatedAt ?? request.updated_at ?? '',
      executingMachineId: request.executingMachineId ?? request.executing_machine_id ?? '',
    };

    if (request.result !== undefined && request.result !== '') {
      normalized.result = parseMaybeJSON(request.result);
    }

    const resultType = request.resultType ?? request.result_type;
    if (resultType) {
      normalized.resultType = resultType;
    }

    if (request.error) {
      normalized.error = request.error;
    }

    const streamResults = request.streamResults ?? request.stream_results;
    if (Array.isArray(streamResults) && streamResults.length > 0) {
      normalized.streamResults = streamResults;
    }

    return normalized;
  }

  private unwrapObject(value: unknown): Record<string, unknown> {
    return typeof value === 'object' && value !== null ? { ...value } : {};
  }
}