import {
  Machine,
  ProviderRuntimeClient,
  ProviderRuntimeOptions,
  ProviderRuntimeSessionClient,
  ProviderSessionAttachOptions,
  ProviderSessionCreateOptions,
  ProviderToolContext,
  ProviderToolHandler,
  ProviderToolRegistration,
  Request,
  RegisterToolRequest,
  Session,
} from './interfaces';

import {
  ConnectionError,
  ToolplaneError,
} from './errors';

const DEFAULT_POLL_INTERVAL_MS = 1_000;
const DEFAULT_HEARTBEAT_INTERVAL_MS = 60_000;
const DEFAULT_SDK_VERSION = '1.0.0';

interface RegisteredProviderTool {
  name: string;
  description: string;
  schema: string;
  config: Record<string, string>;
  tags: string[];
  stream: boolean;
  handler: ProviderToolHandler;
}

interface ManagedSessionState {
  sessionId: string;
  client: ProviderRuntimeSessionClient;
  machineId: string;
  sdkVersion: string;
  tools: Map<string, RegisteredProviderTool>;
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

function isAsyncIterable(value: unknown): value is AsyncIterable<unknown> {
  return value !== null && typeof value === 'object' && Symbol.asyncIterator in value;
}

function isIterable(value: unknown): value is Iterable<unknown> {
  return value !== null && typeof value === 'object' && Symbol.iterator in value;
}

function normalizeSchema(schema: string | Record<string, unknown> | undefined): string {
  if (typeof schema === 'string' && schema.trim()) {
    return schema;
  }

  if (schema && typeof schema === 'object') {
    return JSON.stringify(schema);
  }

  return JSON.stringify({
    type: 'object',
    additionalProperties: true,
  });
}

function normalizeErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }

  return String(error);
}

async function* toChunkStream(result: unknown): AsyncIterable<unknown> {
  if (result === undefined) {
    return;
  }

  if (typeof result === 'string') {
    yield result;
    return;
  }

  if (isAsyncIterable(result)) {
    for await (const chunk of result) {
      yield chunk;
    }
    return;
  }

  if (isIterable(result)) {
    for (const chunk of result) {
      yield chunk;
    }
    return;
  }

  yield result;
}

export class ProviderRuntime {
  private readonly sessions = new Map<string, ManagedSessionState>();

  private readonly activeRequests = new Map<string, Promise<void>>();

  private readonly pollIntervalMs: number;

  private readonly heartbeatIntervalMs: number;

  private readonly sdkVersion: string;

  private _running = false;

  private backgroundLoop?: Promise<void>;

  private polling = false;

  constructor(
    private readonly client: ProviderRuntimeClient,
    options: ProviderRuntimeOptions = {},
  ) {
    this.pollIntervalMs = options.pollIntervalMs ?? DEFAULT_POLL_INTERVAL_MS;
    this.heartbeatIntervalMs = options.heartbeatIntervalMs ?? DEFAULT_HEARTBEAT_INTERVAL_MS;
    this.sdkVersion = options.sdkVersion ?? DEFAULT_SDK_VERSION;
  }

  get running(): boolean {
    return this._running;
  }

  managedSessionIds(): string[] {
    return Array.from(this.sessions.keys());
  }

  async attachSession(
    sessionId: string,
    options: ProviderSessionAttachOptions = {},
  ): Promise<Session> {
    if (!sessionId.trim()) {
      throw new ToolplaneError('ProviderRuntime.attachSession requires a non-empty sessionId');
    }

    const state = this.getOrCreateSessionState(sessionId, options.sdkVersion);
    await this.ensureConnected(state.client);

    const session = await state.client.getSession();
    if (options.registerMachine !== false) {
      await this.ensureMachine(state, options.machineId);
    }

    return session;
  }

  async createSession(options: ProviderSessionCreateOptions): Promise<Session> {
    const sessionClient = this.client.forkSession(options.sessionId ?? '');
    await this.ensureConnected(sessionClient);

    const session = await sessionClient.createSession(
      options.name,
      options.description,
      options.namespace ?? 'default',
      options.sessionId,
    );

    const state = this.getOrCreateSessionState(session.id, options.sdkVersion);
    state.client = sessionClient;

    if (options.registerMachine !== false) {
      await this.ensureMachine(state, options.machineId);
    }

    return session;
  }

  async registerTool(definition: ProviderToolRegistration): Promise<ProviderToolHandler> {
    if (!definition.sessionId.trim()) {
      throw new ToolplaneError('ProviderRuntime.registerTool requires a non-empty sessionId');
    }
    if (!definition.name.trim()) {
      throw new ToolplaneError('ProviderRuntime.registerTool requires a non-empty tool name');
    }

    await this.attachSession(definition.sessionId, { registerMachine: false });
    const state = this.getOrCreateSessionState(definition.sessionId);

    const registeredTool: RegisteredProviderTool = {
      name: definition.name,
      description: definition.description,
      schema: normalizeSchema(definition.schema),
      config: { ...(definition.config ?? {}) },
      tags: [...(definition.tags ?? [])],
      stream: definition.stream ?? false,
      handler: definition.handler,
    };

    const hadMachine = Boolean(state.machineId);
    state.tools.set(definition.name, registeredTool);

    if (!hadMachine) {
      await this.ensureMachine(state);
      return definition.handler;
    }

    await state.client.registerTool(
      registeredTool.name,
      registeredTool.description,
      registeredTool.schema,
      registeredTool.config,
      registeredTool.tags,
    );

    return definition.handler;
  }

  tool(definition: Omit<ProviderToolRegistration, 'handler'>, handler: ProviderToolHandler): Promise<ProviderToolHandler>;
  tool(
    definition: Omit<ProviderToolRegistration, 'handler'>,
  ): (handler: ProviderToolHandler) => Promise<ProviderToolHandler>;
  tool(
    definition: Omit<ProviderToolRegistration, 'handler'>,
    handler?: ProviderToolHandler,
  ): Promise<ProviderToolHandler> | ((handler: ProviderToolHandler) => Promise<ProviderToolHandler>) {
    if (handler) {
      return this.registerTool({ ...definition, handler });
    }

    return async (registeredHandler: ProviderToolHandler) => {
      await this.registerTool({ ...definition, handler: registeredHandler });
      return registeredHandler;
    };
  }

  async pollOnce(): Promise<void> {
    if (this.polling) {
      return;
    }

    this.polling = true;
    try {
      for (const sessionId of this.managedSessionIds()) {
        await this.pollSession(sessionId);
      }
    } finally {
      this.polling = false;
    }
  }

  async startInBackground(sessionIds?: Iterable<string>): Promise<ProviderRuntime> {
    if (sessionIds) {
      for (const sessionId of sessionIds) {
        this.getOrCreateSessionState(sessionId);
      }
    }

    if (this.sessions.size === 0) {
      throw new ToolplaneError('ProviderRuntime requires at least one attached or created session');
    }

    for (const state of this.sessions.values()) {
      await this.attachSession(state.sessionId, { registerMachine: true, sdkVersion: state.sdkVersion });
    }

    if (this._running) {
      return this;
    }

    this._running = true;
    this.backgroundLoop = this.runLoop();
    return this;
  }

  async runForever(sessionIds?: Iterable<string>): Promise<void> {
    await this.startInBackground(sessionIds);
    while (this._running) {
      await sleep(1_000);
    }
  }

  async stop(): Promise<void> {
    this._running = false;
    if (this.backgroundLoop) {
      await this.backgroundLoop;
      this.backgroundLoop = undefined;
    }

    await Promise.allSettled(Array.from(this.activeRequests.values()));
  }

  async drain(): Promise<void> {
    await this.stop();

    await Promise.allSettled(
      Array.from(this.sessions.values()).map(async (state) => {
        try {
          if (state.machineId) {
            await state.client.drainMachine();
          }
        } finally {
          try {
            await state.client.disconnect();
          } finally {
            state.client = this.client.forkSession(state.sessionId);
            state.machineId = '';
          }
        }
      }),
    );
  }

  async close(): Promise<void> {
    await this.drain();
  }

  private getOrCreateSessionState(sessionId: string, sdkVersion?: string): ManagedSessionState {
    const existing = this.sessions.get(sessionId);
    if (existing) {
      if (sdkVersion) {
        existing.sdkVersion = sdkVersion;
      }
      return existing;
    }

    const state: ManagedSessionState = {
      sessionId,
      client: this.client.forkSession(sessionId),
      machineId: '',
      sdkVersion: sdkVersion ?? this.sdkVersion,
      tools: new Map(),
    };

    this.sessions.set(sessionId, state);
    return state;
  }

  private async ensureConnected(client: ProviderRuntimeSessionClient): Promise<void> {
    if (client.isConnected()) {
      return;
    }

    try {
      await client.connect();
    } catch (error) {
      throw new ConnectionError(`Failed to connect provider runtime client: ${normalizeErrorMessage(error)}`);
    }
  }

  private async ensureMachine(state: ManagedSessionState, requestedMachineId?: string): Promise<Machine> {
    await this.ensureConnected(state.client);

    if (state.machineId) {
      return state.client.updateMachinePing(state.machineId);
    }

    const tools: RegisterToolRequest[] = Array.from(state.tools.values()).map((tool) => ({
      sessionId: state.sessionId,
      name: tool.name,
      description: tool.description,
      schema: tool.schema,
      config: tool.config,
      tags: tool.tags,
    }));

    const machine = await state.client.registerMachine(requestedMachineId ?? '', state.sdkVersion, tools);
    state.machineId = machine.id;
    return machine;
  }

  private async runLoop(): Promise<void> {
    let lastHeartbeatAt = 0;

    while (this._running) {
      try {
        const now = Date.now();
        if (now - lastHeartbeatAt >= this.heartbeatIntervalMs) {
          await this.sendHeartbeats();
          lastHeartbeatAt = Date.now();
        }

        await this.pollOnce();
      } catch {
        // Keep the provider loop alive; request-level failures are reflected in request state.
      }

      if (!this._running) {
        break;
      }

      await sleep(this.pollIntervalMs);
    }
  }

  private async sendHeartbeats(): Promise<void> {
    await Promise.allSettled(
      Array.from(this.sessions.values())
        .filter((state) => Boolean(state.machineId))
        .map((state) => state.client.updateMachinePing(state.machineId)),
    );
  }

  private async pollSession(sessionId: string): Promise<void> {
    const state = this.sessions.get(sessionId);
    if (!state) {
      return;
    }

    await this.ensureMachine(state);

    const pendingRequests = await state.client.listRequests({
      status: 'pending',
      limit: 20,
    });

    const nextRequest = pendingRequests.find((request) => (
      request.id &&
      state.tools.has(request.toolName) &&
      !this.activeRequests.has(request.id)
    ));

    if (!nextRequest) {
      return;
    }

    const activeRequest = this.handleRequest(state, nextRequest).finally(() => {
      this.activeRequests.delete(nextRequest.id);
    });

    this.activeRequests.set(nextRequest.id, activeRequest);
    await activeRequest;
  }

  private async handleRequest(state: ManagedSessionState, request: Request): Promise<void> {
    const claimedRequest = await state.client.claimRequest(request.id, state.machineId);
    await state.client.updateRequest(claimedRequest.id, { status: 'running' });

    const tool = state.tools.get(claimedRequest.toolName);
    if (!tool) {
      await state.client.submitRequestResult(
        claimedRequest.id,
        { error: 'tool not found' },
        'rejection',
        { handled_by: 'typescript-provider-runtime' },
      );
      return;
    }

    const input = this.parseInput(claimedRequest.input);
    const streamedChunks: unknown[] = [];
    const context: ProviderToolContext = {
      sessionId: state.sessionId,
      requestId: claimedRequest.id,
      toolName: claimedRequest.toolName,
      machineId: state.machineId,
      input,
      appendChunk: async (chunk: unknown) => {
        await state.client.appendRequestChunks(claimedRequest.id, [chunk], 'streaming');
        streamedChunks.push(chunk);
      },
      heartbeat: async () => state.client.updateMachinePing(state.machineId),
    };

    try {
      if (tool.stream) {
        await state.client.updateRequest(claimedRequest.id, { resultType: 'streaming' });
        const result = await tool.handler(input, context);

        for await (const chunk of toChunkStream(result)) {
          await context.appendChunk(chunk);
        }

        await state.client.submitRequestResult(
          claimedRequest.id,
          streamedChunks,
          'resolution',
          { handled_by: 'typescript-provider-runtime' },
        );
        return;
      }

      const result = await tool.handler(input, context);
      await state.client.submitRequestResult(
        claimedRequest.id,
        result,
        'resolution',
        { handled_by: 'typescript-provider-runtime' },
      );
    } catch (error) {
      await state.client.submitRequestResult(
        claimedRequest.id,
        { error: normalizeErrorMessage(error) },
        'rejection',
        { handled_by: 'typescript-provider-runtime' },
      );
    }
  }

  private parseInput(input: string): Record<string, unknown> {
    if (!input) {
      return {};
    }

    try {
      const parsed = JSON.parse(input);
      return typeof parsed === 'object' && parsed !== null ? { ...parsed } : {};
    } catch {
      return {};
    }
  }
}