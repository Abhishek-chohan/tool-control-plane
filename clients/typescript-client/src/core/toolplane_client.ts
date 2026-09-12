import * as grpc from '@grpc/grpc-js';
import * as fs from 'node:fs';
import { v4 as uuidv4 } from 'uuid';

import {
  ApiKey,
  ClientConfig,
  ClientProtocol,
  ConnectionStatus,
  CreateSessionRequest,
  GRPCTLSConfig,
  LeaseContext,
  Machine,
  ProviderRuntimeOptions,
  RegisterToolOptions,
  Request as RequestModel,
  RequestUpdate,
  Task,
  RegisterToolRequest,
  Session,
  Tool,
} from '../interfaces';

import { ProviderRuntime } from '../provider_runtime';

import {
  AlreadyExistsError,
  ConnectionError,
  FailedPreconditionError,
  InvalidArgumentError,
  NotFoundError,
  PermissionDeniedError,
  ProtocolError,
  ResourceExhaustedError,
  TimeoutError,
  ToolplaneError,
  UnauthenticatedError,
  UnavailableError,
} from '../errors';

import {
  ApiKey as ProtoApiKey,
  AppendRequestChunksRequest as AppendRequestChunksMessage,
  AppendRequestChunksResponse as AppendRequestChunksResponseMessage,
  CancelRequestRequest as CancelRequestMessage,
  CancelRequestResponse as CancelRequestResponseMessage,
  ClaimRequestRequest as ClaimRequestMessage,
  CancelTaskRequest as CancelTaskMessage,
  CancelTaskResponse as CancelTaskResponseMessage,
  CreateApiKeyRequest as CreateApiKeyMessage,
  CreateRequestRequest as CreateRequestMessage,
  CreateSessionResponse as CreateSessionResponseMessage,
  CreateSessionRequest as CreateSessionMessage,
  CreateTaskRequest as CreateTaskMessage,
  DeleteToolRequest as DeleteToolMessage,
  DeleteToolResponse as DeleteToolResponseMessage,
  DrainMachineRequest as DrainMachineMessage,
  DrainMachineResponse as DrainMachineResponseMessage,
  ExecuteToolChunk as ExecuteToolChunkMessage,
  ExecuteToolRequest as ExecuteToolMessage,
  ExecuteToolResponse as ExecuteToolResponseMessage,
  GetMachineRequest as GetMachineMessage,
  GetToolRequest as GetToolMessage,
  GetToolResponse as GetToolResponseMessage,
  GetTaskRequest as GetTaskMessage,
  GetRequestChunksRequest as GetRequestChunksMessage,
  GetRequestChunksResponse as GetRequestChunksResponseMessage,
  GetRequestRequest,
  GetSessionRequest,
  HealthCheckRequest,
  HealthCheckResponse as HealthCheckResponseMessage,
  ListApiKeysRequest as ListApiKeysMessage,
  ListApiKeysResponse as ListApiKeysResponseMessage,
  ListMachinesRequest as ListMachinesMessage,
  ListMachinesResponse as ListMachinesResponseMessage,
  ListRequestsRequest as ListRequestsMessage,
  ListRequestsResponse as ListRequestsResponseMessage,
  ListSessionsRequest as ListSessionsMessage,
  ListSessionsResponse as ListSessionsResponseMessage,
  ListTasksRequest as ListTasksMessage,
  ListTasksResponse as ListTasksResponseMessage,
  ListToolsRequest,
  ListToolsResponse as ListToolsResponseMessage,
  Machine as ProtoMachine,
  RenewRequestLeaseRequest as RenewRequestLeaseMessage,
  ResumeStreamRequest as ResumeStreamMessage,
  RevokeApiKeyRequest as RevokeApiKeyMessage,
  RevokeApiKeyResponse as RevokeApiKeyResponseMessage,
  RegisterMachineRequest as RegisterMachineMessage,
  RegisterToolRequest as RegisterToolMessage,
  RegisterToolResponse as RegisterToolResponseMessage,
  Request as ProtoRequest,
  RequestStatus,
  Session as ProtoSession,
  SubmitRequestResultRequest as SubmitRequestResultMessage,
  SubmitRequestResultResponse as SubmitRequestResultResponseMessage,
  Task as ProtoTask,
  TaskStatus,
  Tool as ProtoTool,
  UpdateMachinePingRequest as UpdateMachinePingMessage,
  UpdateRequestRequest as UpdateRequestMessage,
  UpdateSessionRequest as UpdateSessionMessage,
  UnregisterMachineRequest as UnregisterMachineMessage,
  UnregisterMachineResponse as UnregisterMachineResponseMessage,
} from '../proto/proto/service_pb';

import {
  MachinesServiceClient,
  RequestsServiceClient,
  SessionsServiceClient,
  TasksServiceClient,
  ToolServiceClient,
} from '../proto/proto/service_grpc_pb';

/** A normalized execution chunk: the wire fields of ExecuteToolChunk as
 * plain values, safe to hand to callers without the proto machinery. */
export interface ExecuteToolChunkModel {
  seq: number;
  requestId: string;
  chunk: string;
  isFinal: boolean;
  error: string;
}

const DEFAULT_TIMEOUT_MS = 30_000;
const REQUEST_POLL_INTERVAL_MS = 100;
const HEALTHY_STATUSES = new Set(['healthy', 'ok']);

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

// v1 wire statuses are enums; the SDK keeps exposing the friendly lowercase
// names callers have always seen.
function normalizeRequestStatus(status: number): string {
  switch (status) {
    case RequestStatus.REQUEST_STATUS_DONE:
      return 'done';
    case RequestStatus.REQUEST_STATUS_FAILED:
      return 'failure';
    case RequestStatus.REQUEST_STATUS_PENDING:
      return 'pending';
    case RequestStatus.REQUEST_STATUS_CLAIMED:
      return 'claimed';
    case RequestStatus.REQUEST_STATUS_RUNNING:
      return 'running';
    default:
      return '';
  }
}

function requestStatusForWire(status?: string): number {
  switch ((status ?? '').trim().toLowerCase()) {
    case 'done':
      return RequestStatus.REQUEST_STATUS_DONE;
    case 'failure':
    case 'failed':
      return RequestStatus.REQUEST_STATUS_FAILED;
    case 'pending':
      return RequestStatus.REQUEST_STATUS_PENDING;
    case 'claimed':
      return RequestStatus.REQUEST_STATUS_CLAIMED;
    case 'running':
      return RequestStatus.REQUEST_STATUS_RUNNING;
    default:
      return RequestStatus.REQUEST_STATUS_UNSPECIFIED;
  }
}

function normalizeTaskStatus(status: number): string {
  switch (status) {
    case TaskStatus.TASK_STATUS_COMPLETED:
      return 'completed';
    case TaskStatus.TASK_STATUS_FAILED:
      return 'failed';
    case TaskStatus.TASK_STATUS_CANCELLED:
      return 'cancelled';
    case TaskStatus.TASK_STATUS_DEAD_LETTER:
      return 'dead_letter';
    case TaskStatus.TASK_STATUS_PENDING:
      return 'pending';
    case TaskStatus.TASK_STATUS_RUNNING:
      return 'running';
    default:
      return '';
  }
}

// v1 carries times as google.protobuf.Timestamp; the SDK surfaces ISO strings.
function timestampToIso(value: { toDate(): Date } | undefined): string {
  return value ? value.toDate().toISOString() : '';
}

export class ToolplaneClient {
  private config: ClientConfig;

  private connected = false;

  private grpcAddress = '';

  private toolClient?: ToolServiceClient;

  private sessionClient?: SessionsServiceClient;

  private machineClient?: MachinesServiceClient;

  private requestsClient?: RequestsServiceClient;

  private tasksClient?: TasksServiceClient;

  private machineId = '';

  // Per-machine credential minted at machine registration; presented on
  // provide-scoped RPCs via the x-toolplane-machine-token metadata.
  private machineToken = '';

  constructor(config: ClientConfig) {
    this.config = {
      timeout: DEFAULT_TIMEOUT_MS,
      retryAttempts: 3,
      retryDelay: 1000,
      ...config,
    };

    this.initializeClients();
  }

  private get timeoutMs(): number {
    return this.config.timeout ?? DEFAULT_TIMEOUT_MS;
  }

  private initializeClients(): void {
    if (this.config.protocol !== ClientProtocol.GRPC) {
      throw new ProtocolError(`Unsupported protocol: ${String(this.config.protocol)}`);
    }

    this.grpcAddress = `${this.config.serverHost}:${this.config.serverPort}`;
    const credentials = this.createChannelCredentials();
    const channelOptions = this.channelOptions();
    this.toolClient = new ToolServiceClient(this.grpcAddress, credentials, channelOptions);
    this.sessionClient = new SessionsServiceClient(this.grpcAddress, credentials, channelOptions);
    this.machineClient = new MachinesServiceClient(this.grpcAddress, credentials, channelOptions);
    this.requestsClient = new RequestsServiceClient(this.grpcAddress, credentials, channelOptions);
    this.tasksClient = new TasksServiceClient(this.grpcAddress, credentials, channelOptions);
  }

  private isTLSEnabled(): boolean {
    const tlsConfig = this.config.tls;
    if (!tlsConfig) {
      return false;
    }

    return Boolean(
      tlsConfig.enabled
      || tlsConfig.caCertPath?.trim()
      || tlsConfig.serverName?.trim(),
    );
  }

  private createChannelCredentials(): grpc.ChannelCredentials {
    if (!this.isTLSEnabled()) {
      return grpc.credentials.createInsecure();
    }

    const caCertPath = this.config.tls?.caCertPath?.trim();
    const caCert = caCertPath ? fs.readFileSync(caCertPath) : undefined;
    return grpc.credentials.createSsl(caCert);
  }

  private channelOptions(): grpc.ChannelOptions {
    if (!this.isTLSEnabled()) {
      return {};
    }

    const serverName = this.config.tls?.serverName?.trim();
    if (!serverName) {
      return {};
    }

    return {
      'grpc.ssl_target_name_override': serverName,
      'grpc.default_authority': serverName,
    };
  }

  async connect(): Promise<void> {
    try {
      await this.connectGRPC();
      this.connected = true;
    } catch (error) {
      this.connected = false;
      throw new ConnectionError(`Failed to connect via ${this.config.protocol}: ${String(error)}`);
    }
  }

  private async connectGRPC(): Promise<void> {
    this.ensureGRPCClientsInitialized();

    await Promise.all([
      this.waitForClientReady(this.toolClient!),
      this.waitForClientReady(this.sessionClient!),
      this.waitForClientReady(this.machineClient!),
      this.waitForClientReady(this.requestsClient!),
      this.waitForClientReady(this.tasksClient!),
    ]);

    const request = new HealthCheckRequest();
    const response = await this.invokeGRPCUnary<HealthCheckResponseMessage>(
      (metadata, options, callback) => this.toolClient!.healthCheck(request, metadata, options, callback),
      'gRPC health check failed',
    );

    if (!HEALTHY_STATUSES.has(response.getStatus().toLowerCase())) {
      throw new ConnectionError(`gRPC health check returned unexpected status: ${response.getStatus()}`);
    }
  }

  async disconnect(): Promise<void> {
    try {
      const sessionId = this.config.sessionId.trim();
      if (this.connected && sessionId && this.machineId && this.machineClient) {
        const request = new UnregisterMachineMessage();
        request.setSessionId(sessionId);
        request.setMachineId(this.machineId);
        try {
          await this.invokeGRPCUnary(
            (metadata, options, callback) => this.machineClient!.unregisterMachine(request, metadata, options, callback),
            `failed to unregister machine ${this.machineId}`,
          );
        } catch {
          // Best-effort cleanup during disconnect.
        }
      }

      this.toolClient?.close();
      this.sessionClient?.close();
      this.machineClient?.close();
      this.requestsClient?.close();
      this.tasksClient?.close();
      this.machineId = '';
    } finally {
      this.connected = false;
    }
  }

  isConnected(): boolean {
    return this.connected;
  }

  getConnectionStatus(): ConnectionStatus {
    return {
      connected: this.connected,
      protocol: this.config.protocol,
      serverUrl: this.grpcAddress || `${this.config.serverHost}:${this.config.serverPort}`,
      lastPing: new Date(),
    };
  }

  async executeTool(
    toolName: string,
    params: Record<string, unknown> = {},
    options: { idempotencyKey?: string } = {},
  ): Promise<RequestModel> {
    const request = await this.executeToolGRPC(toolName, params, options.idempotencyKey ?? '');
    return this.normalizeRequest(request);
  }

  async ping(): Promise<string> {
    return this.pingGRPC();
  }

  async registerTool(
    name: string,
    description: string,
    schema: string,
    config: Record<string, string> = {},
    tags: string[] = [],
    options: RegisterToolOptions = {},
  ): Promise<Tool> {
    this.ensureGRPCConnected('tool registration');

    const sessionId = options.sessionId?.trim() || this.getRequiredSessionId('tool registration');
    const machineId = options.machineId?.trim() || this.getRequiredMachineId('tool registration');
    const request = new RegisterToolMessage();
    request.setSessionId(sessionId);
    request.setMachineId(machineId);
    request.setName(name);
    request.setDescription(description);
    request.setSchema(schema);
    for (const [key, value] of Object.entries(config)) {
      request.getConfigMap().set(key, value);
    }
    request.setTagsList(tags);

    const response = await this.invokeGRPCUnary<RegisterToolResponseMessage>(
      (metadata, options, callback) => this.toolClient!.registerTool(request, metadata, options, callback),
      `failed to register tool ${name}`,
    );

    return this.normalizeTool(response.getTool());
  }

  async listTools(): Promise<Tool[]> {
    this.ensureGRPCConnected('tool listing');

    const request = new ListToolsRequest();
    request.setSessionId(this.getRequiredSessionId('tool listing'));

    const response = await this.invokeGRPCUnary<ListToolsResponseMessage>(
      (metadata, options, callback) => this.toolClient!.listTools(request, metadata, options, callback),
      'failed to list tools',
    );

    return response.getToolsList().map((tool) => this.normalizeTool(tool));
  }

  async getToolById(toolId: string): Promise<Tool> {
    this.ensureGRPCConnected('tool lookup');

    const request = new GetToolMessage();
    request.setSessionId(this.getRequiredSessionId('tool lookup'));
    request.setToolId(toolId);

    const response = await this.invokeGRPCUnary<GetToolResponseMessage>(
      (metadata, options, callback) => this.toolClient!.getTool(request, metadata, options, callback),
      `failed to retrieve tool ${toolId}`,
    );

    return this.normalizeTool(response.getTool());
  }

  async getToolByName(toolName: string): Promise<Tool> {
    this.ensureGRPCConnected('tool lookup');

    const request = new GetToolMessage();
    request.setSessionId(this.getRequiredSessionId('tool lookup'));
    request.setToolName(toolName);

    const response = await this.invokeGRPCUnary<GetToolResponseMessage>(
      (metadata, options, callback) => this.toolClient!.getTool(request, metadata, options, callback),
      `failed to retrieve tool ${toolName}`,
    );

    return this.normalizeTool(response.getTool());
  }

  async deleteTool(toolId: string): Promise<boolean> {
    this.ensureGRPCConnected('tool deletion');

    const request = new DeleteToolMessage();
    request.setSessionId(this.getRequiredSessionId('tool deletion'));
    request.setToolId(toolId);

    const response = await this.invokeGRPCUnary<DeleteToolResponseMessage>(
      (metadata, options, callback) => this.toolClient!.deleteTool(request, metadata, options, callback),
      `failed to delete tool ${toolId}`,
    );

    return response.getSuccess();
  }

  async createSession(
    name: string,
    description: string,
    namespace: string = 'default',
    requestedSessionId: string = '',
  ): Promise<Session> {
    this.ensureGRPCConnected('session creation');

    const sessionId = requestedSessionId || this.config.sessionId || uuidv4();
    const request: CreateSessionRequest = {
      userId: this.config.userId,
      name,
      description,
      sessionId,
      namespace,
    };

    const message = new CreateSessionMessage();
    message.setUserId(request.userId);
    message.setName(request.name);
    message.setDescription(request.description);
    message.setSessionId(request.sessionId);
    message.setNamespace(request.namespace);

    const response = await this.invokeGRPCUnary<CreateSessionResponseMessage>(
      (metadata, options, callback) => this.sessionClient!.createSession(message, metadata, options, callback),
      'failed to create session',
    );

    const session = this.normalizeSession(response.getSession());
    this.config.sessionId = session.id;
    this.machineId = '';
    return session;
  }

  async getSession(): Promise<Session> {
    this.ensureGRPCConnected('session retrieval');

    const request = new GetSessionRequest();
    request.setSessionId(this.getRequiredSessionId('session retrieval'));

    const response = await this.invokeGRPCUnary<ProtoSession>(
      (metadata, options, callback) => this.sessionClient!.getSession(request, metadata, options, callback),
      'failed to retrieve session',
    );

    return this.normalizeSession(response);
  }

  async listSessions(): Promise<Session[]> {
    this.ensureGRPCConnected('session listing');

    const request = new ListSessionsMessage();
    request.setUserId(this.config.userId);

    const response = await this.invokeGRPCUnary<ListSessionsResponseMessage>(
      (metadata, options, callback) => this.sessionClient!.listSessions(request, metadata, options, callback),
      'failed to list sessions',
    );

    return response.getSessionsList().map((session) => this.normalizeSession(session));
  }

  async updateSession(name: string = '', description: string = '', namespace: string = ''): Promise<Session> {
    this.ensureGRPCConnected('session update');

    const request = new UpdateSessionMessage();
    request.setSessionId(this.getRequiredSessionId('session update'));
    request.setName(name);
    request.setDescription(description);
    request.setNamespace(namespace);

    const response = await this.invokeGRPCUnary<ProtoSession>(
      (metadata, options, callback) => this.sessionClient!.updateSession(request, metadata, options, callback),
      'failed to update session',
    );

    return this.normalizeSession(response);
  }

  async createApiKey(name: string, capabilities: string[]): Promise<ApiKey> {
    // Keys are minted least-privilege and explicit: fail fast instead of
    // sending an empty capability list the server rejects.
    if (!Array.isArray(capabilities) || capabilities.filter((capability) => capability.trim()).length === 0) {
      throw new ToolplaneError('createApiKey requires an explicit non-empty capabilities list');
    }
    this.ensureGRPCConnected('api key creation');

    const request = new CreateApiKeyMessage();
    request.setSessionId(this.getRequiredSessionId('api key creation'));
    request.setName(name);
		request.setCapabilitiesList(capabilities);

    const response = await this.invokeGRPCUnary<ProtoApiKey>(
      (metadata, options, callback) => this.sessionClient!.createApiKey(request, metadata, options, callback),
      `failed to create api key ${name}`,
    );

    return this.normalizeApiKey(response);
  }

  async listApiKeys(): Promise<ApiKey[]> {
    this.ensureGRPCConnected('api key listing');

    const request = new ListApiKeysMessage();
    request.setSessionId(this.getRequiredSessionId('api key listing'));

    const response = await this.invokeGRPCUnary<ListApiKeysResponseMessage>(
      (metadata, options, callback) => this.sessionClient!.listApiKeys(request, metadata, options, callback),
      'failed to list api keys',
    );

    return response.getApiKeysList().map((apiKey) => this.normalizeApiKey(apiKey));
  }

  async revokeApiKey(keyId: string): Promise<boolean> {
    this.ensureGRPCConnected('api key revocation');

    const request = new RevokeApiKeyMessage();
    request.setSessionId(this.getRequiredSessionId('api key revocation'));
    request.setKeyId(keyId);

    const response = await this.invokeGRPCUnary<RevokeApiKeyResponseMessage>(
      (metadata, options, callback) => this.sessionClient!.revokeApiKey(request, metadata, options, callback),
      `failed to revoke api key ${keyId}`,
    );

    return response.getSuccess();
  }

  async registerMachine(
    machineId: string = '',
    sdkVersion: string = '1.0.0',
    tools: RegisterToolRequest[] = [],
  ): Promise<Machine> {
    this.ensureGRPCConnected('machine registration');

    const sessionId = this.getRequiredSessionId('machine registration');
    const resolvedMachineId = machineId || uuidv4();
    const request = new RegisterMachineMessage();
    request.setSessionId(sessionId);
    request.setMachineId(resolvedMachineId);
    request.setSdkVersion(sdkVersion);
    request.setSdkLanguage('typescript');
    request.setToolsList(
      tools.map((tool) => {
        const embeddedTool = new RegisterToolMessage();
        embeddedTool.setSessionId(tool.sessionId || sessionId);
        embeddedTool.setMachineId(resolvedMachineId);
        embeddedTool.setName(tool.name);
        embeddedTool.setDescription(tool.description);
        embeddedTool.setSchema(tool.schema);
        for (const [key, value] of Object.entries(tool.config || {})) {
          embeddedTool.getConfigMap().set(key, value);
        }
        embeddedTool.setTagsList(tool.tags || []);
        return embeddedTool;
      }),
    );

    const response = await this.invokeGRPCUnary<ProtoMachine>(
      (metadata, options, callback) => this.machineClient!.registerMachine(request, metadata, options, callback),
      'failed to register machine',
    );

    const machine = this.normalizeMachine(response);
    this.machineId = machine.id || resolvedMachineId;
    // The registration response carries the per-machine credential exactly
    // once; keep it so provide-scoped calls present it (createMetadata).
    const machineToken = response.getMachineToken();
    if (machineToken) {
      this.machineToken = machineToken;
    }
    return machine;
  }

  async listMachines(): Promise<Machine[]> {
    this.ensureGRPCConnected('machine listing');

    const request = new ListMachinesMessage();
    request.setSessionId(this.getRequiredSessionId('machine listing'));

    const response = await this.invokeGRPCUnary<ListMachinesResponseMessage>(
      (metadata, options, callback) => this.machineClient!.listMachines(request, metadata, options, callback),
      'failed to list machines',
    );

    return response.getMachinesList().map((machine) => this.normalizeMachine(machine));
  }

  async getMachine(machineId: string): Promise<Machine> {
    this.ensureGRPCConnected('machine retrieval');

    const request = new GetMachineMessage();
    request.setSessionId(this.getRequiredSessionId('machine retrieval'));
    request.setMachineId(machineId);

    const response = await this.invokeGRPCUnary<ProtoMachine>(
      (metadata, options, callback) => this.machineClient!.getMachine(request, metadata, options, callback),
      `failed to retrieve machine ${machineId}`,
    );

    return this.normalizeMachine(response);
  }

  async updateMachinePing(machineId: string = ''): Promise<Machine> {
    this.ensureGRPCConnected('machine heartbeat');

    const resolvedMachineId = machineId || this.getRequiredMachineId('machine heartbeat');
    const request = new UpdateMachinePingMessage();
    request.setSessionId(this.getRequiredSessionId('machine heartbeat'));
    request.setMachineId(resolvedMachineId);

    const response = await this.invokeGRPCUnary<ProtoMachine>(
      (metadata, options, callback) => this.machineClient!.updateMachinePing(request, metadata, options, callback),
      `failed to update heartbeat for machine ${resolvedMachineId}`,
    );

    return this.normalizeMachine(response);
  }

  async unregisterMachine(machineId: string = ''): Promise<boolean> {
    this.ensureGRPCConnected('machine unregistration');

    const resolvedMachineId = machineId || this.getRequiredMachineId('machine unregistration');
    const request = new UnregisterMachineMessage();
    request.setSessionId(this.getRequiredSessionId('machine unregistration'));
    request.setMachineId(resolvedMachineId);

    const response = await this.invokeGRPCUnary<UnregisterMachineResponseMessage>(
      (metadata, options, callback) => this.machineClient!.unregisterMachine(request, metadata, options, callback),
      `failed to unregister machine ${resolvedMachineId}`,
    );

    if (response.getSuccess() && resolvedMachineId === this.machineId) {
      this.machineId = '';
    }

    return response.getSuccess();
  }

  async drainMachine(machineId: string = ''): Promise<boolean> {
    this.ensureGRPCConnected('machine drain');

    const resolvedMachineId = machineId || this.getRequiredMachineId('machine drain');
    const request = new DrainMachineMessage();
    request.setSessionId(this.getRequiredSessionId('machine drain'));
    request.setMachineId(resolvedMachineId);

    const response = await this.invokeGRPCUnary<DrainMachineResponseMessage>(
      (metadata, options, callback) => this.machineClient!.drainMachine(request, metadata, options, callback),
      `failed to drain machine ${resolvedMachineId}`,
    );

    if (response.getDrained() && resolvedMachineId === this.machineId) {
      this.machineId = '';
    }

    return response.getDrained();
  }

  async createTask(toolName: string, input: string, idempotencyKey: string = ''): Promise<Task> {
    this.ensureGRPCConnected('task creation');

    const request = new CreateTaskMessage();
    request.setSessionId(this.getRequiredSessionId('task creation'));
    request.setToolName(toolName);
    request.setInput(input);
    request.setIdempotencyKey(idempotencyKey);

    const response = await this.invokeGRPCUnary<ProtoTask>(
      (metadata, options, callback) => this.tasksClient!.createTask(request, metadata, options, callback),
      `failed to create task for ${toolName}`,
    );

    return this.normalizeTask(response);
  }

  async getTask(taskId: string): Promise<Task> {
    this.ensureGRPCConnected('task retrieval');

    const request = new GetTaskMessage();
    request.setSessionId(this.getRequiredSessionId('task retrieval'));
    request.setTaskId(taskId);

    const response = await this.invokeGRPCUnary<ProtoTask>(
      (metadata, options, callback) => this.tasksClient!.getTask(request, metadata, options, callback),
      `failed to retrieve task ${taskId}`,
    );

    return this.normalizeTask(response);
  }

  async listTasks(): Promise<Task[]> {
    this.ensureGRPCConnected('task listing');

    const request = new ListTasksMessage();
    request.setSessionId(this.getRequiredSessionId('task listing'));

    const response = await this.invokeGRPCUnary<ListTasksResponseMessage>(
      (metadata, options, callback) => this.tasksClient!.listTasks(request, metadata, options, callback),
      'failed to list tasks',
    );

    return response.getTasksList().map((task) => this.normalizeTask(task));
  }

  async cancelTask(taskId: string): Promise<boolean> {
    this.ensureGRPCConnected('task cancellation');

    const request = new CancelTaskMessage();
    request.setSessionId(this.getRequiredSessionId('task cancellation'));
    request.setTaskId(taskId);

    const response = await this.invokeGRPCUnary<CancelTaskResponseMessage>(
      (metadata, options, callback) => this.tasksClient!.cancelTask(request, metadata, options, callback),
      `failed to cancel task ${taskId}`,
    );

    return response.getSuccess();
  }

  async createRequest(toolName: string, input: string, idempotencyKey: string = ''): Promise<RequestModel> {
    this.ensureGRPCConnected('request creation');

    const request = new CreateRequestMessage();
    request.setSessionId(this.getRequiredSessionId('request creation'));
    request.setToolName(toolName);
    request.setInput(input);
    request.setIdempotencyKey(idempotencyKey);

    const response = await this.invokeGRPCUnary<ProtoRequest>(
      (metadata, options, callback) => this.requestsClient!.createRequest(request, metadata, options, callback),
      `failed to create request for ${toolName}`,
    );

    return this.normalizeRequest(response);
  }

  async getRequest(requestId: string): Promise<RequestModel> {
    this.ensureGRPCConnected('request retrieval');

    const request = new GetRequestRequest();
    request.setSessionId(this.getRequiredSessionId('request retrieval'));
    request.setRequestId(requestId);

    const response = await this.invokeGRPCUnary<ProtoRequest>(
      (metadata, options, callback) => this.requestsClient!.getRequest(request, metadata, options, callback),
      `failed to retrieve request ${requestId}`,
      requestId,
    );

    return this.normalizeRequest(response);
  }

  async getRequestChunksWindow(requestId: string): Promise<{
    chunks: unknown[];
    startSeq: number;
    nextSeq: number;
  }> {
    this.ensureGRPCConnected('request chunks retrieval');

    const request = new GetRequestChunksMessage();
    request.setSessionId(this.getRequiredSessionId('request chunks retrieval'));
    request.setRequestId(requestId);

    const response = await this.invokeGRPCUnary<GetRequestChunksResponseMessage>(
      (metadata, options, callback) => this.requestsClient!.getRequestChunks(request, metadata, options, callback),
      `failed to fetch chunks for request ${requestId}`,
      requestId,
    );

    return {
      chunks: response.getChunksList().map((chunk) => this.parseResultPayload(chunk)),
      startSeq: response.getStartSeq(),
      nextSeq: response.getNextSeq(),
    };
  }

  async listRequests(options: {
    status?: string;
    toolName?: string;
    limit?: number;
    /** Opaque cursor from a previous page; omit to start from the first page. */
    pageToken?: string;
  } = {}): Promise<RequestModel[]> {
    this.ensureGRPCConnected('request listing');

    const request = new ListRequestsMessage();
    request.setSessionId(this.getRequiredSessionId('request listing'));
    request.setStatus(requestStatusForWire(options.status));
    request.setToolName(options.toolName ?? '');
    request.setPageSize(options.limit ?? 0);
    request.setPageToken(options.pageToken ?? '');

    const response = await this.invokeGRPCUnary<ListRequestsResponseMessage>(
      (metadata, optionsArg, callback) => this.requestsClient!.listRequests(request, metadata, optionsArg, callback),
      'failed to list requests',
    );

    return response.getRequestsList().map((item) => this.normalizeRequest(item));
  }

  /**
   * Fenced provider writes must present the lease grant returned by
   * claimRequest. Fail fast client-side with an actionable error instead of
   * sending an empty grant that the server always rejects with
   * FAILED_PRECONDITION. A zero leaseEpoch is accepted: rows claimed before
   * lease fencing existed legitimately carry epoch 0.
   */
  private assertLeaseGrant(machineId: string | undefined, operation: string): void {
    if (!machineId || !machineId.trim()) {
      throw new ToolplaneError(
        `${operation} requires the lease grant from claimRequest (machineId + leaseEpoch); ` +
        'pass the claimed request\'s lease context',
      );
    }
  }

  async updateRequest(requestId: string, update: RequestUpdate): Promise<RequestModel> {
    this.ensureGRPCConnected('request update');
    this.assertLeaseGrant(update.machineId, 'updateRequest');

    const request = new UpdateRequestMessage();
    request.setSessionId(this.getRequiredSessionId('request update'));
    request.setRequestId(requestId);
    request.setStatus(requestStatusForWire(update.status));
    request.setResult(update.result ?? '');
    request.setResultType(update.resultType ?? '');
    // Fencing: present the lease grant for provider writes.
    request.setMachineId(update.machineId ?? '');
    request.setLeaseEpoch(update.leaseEpoch ?? 0);

    const response = await this.invokeGRPCUnary<ProtoRequest>(
      (metadata, options, callback) => this.requestsClient!.updateRequest(request, metadata, options, callback),
      `failed to update request ${requestId}`,
    );

    return this.normalizeRequest(response);
  }

  async claimRequest(requestId: string, machineId: string = ''): Promise<RequestModel> {
    this.ensureGRPCConnected('request claim');

    const resolvedMachineId = machineId || this.getRequiredMachineId('request claim');
    const request = new ClaimRequestMessage();
    request.setSessionId(this.getRequiredSessionId('request claim'));
    request.setRequestId(requestId);
    request.setMachineId(resolvedMachineId);

    const response = await this.invokeGRPCUnary<ProtoRequest>(
      (metadata, options, callback) => this.requestsClient!.claimRequest(request, metadata, options, callback),
      `failed to claim request ${requestId}`,
      requestId,
    );

    return this.normalizeRequest(response);
  }

  async appendRequestChunks(
    requestId: string,
    chunks: unknown[],
    resultType: string = 'streaming',
    lease?: LeaseContext,
  ): Promise<boolean> {
    this.ensureGRPCConnected('request chunk append');
    this.assertLeaseGrant(lease?.machineId, 'appendRequestChunks');

    const request = new AppendRequestChunksMessage();
    request.setSessionId(this.getRequiredSessionId('request chunk append'));
    request.setRequestId(requestId);
    request.setChunksList(chunks.map((chunk) => this.serializePayload(chunk)));
    request.setResultType(resultType);
    // Fencing: present the lease grant for provider writes.
    request.setMachineId(lease?.machineId ?? '');
    request.setLeaseEpoch(lease?.leaseEpoch ?? 0);

    const response = await this.invokeGRPCUnary<AppendRequestChunksResponseMessage>(
      (metadata, options, callback) => this.requestsClient!.appendRequestChunks(request, metadata, options, callback),
      `failed to append chunks for request ${requestId}`,
      requestId,
    );

    return response.getSuccess();
  }

  async submitRequestResult(
    requestId: string,
    result: unknown,
    resultType: string = 'resolution',
    meta: Record<string, string> = {},
    lease?: LeaseContext,
  ): Promise<boolean> {
    this.ensureGRPCConnected('request result submission');
    this.assertLeaseGrant(lease?.machineId, 'submitRequestResult');

    const request = new SubmitRequestResultMessage();
    request.setSessionId(this.getRequiredSessionId('request result submission'));
    request.setRequestId(requestId);
    request.setResult(this.serializePayload(result));
    request.setResultType(resultType);
    for (const [key, value] of Object.entries(meta)) {
      request.getMetaMap().set(key, value);
    }
    // Fencing: present the lease grant for provider writes.
    request.setMachineId(lease?.machineId ?? '');
    request.setLeaseEpoch(lease?.leaseEpoch ?? 0);

    const response = await this.invokeGRPCUnary<SubmitRequestResultResponseMessage>(
      (metadata, options, callback) => this.requestsClient!.submitRequestResult(request, metadata, options, callback),
      `failed to submit result for request ${requestId}`,
      requestId,
    );

    return response.getSuccess();
  }

  async renewRequestLease(
    requestId: string,
    machineId: string,
    leaseEpoch: number,
  ): Promise<RequestModel> {
    this.ensureGRPCConnected('request lease renewal');
    this.assertLeaseGrant(machineId, 'renewRequestLease');

    const request = new RenewRequestLeaseMessage();
    request.setSessionId(this.getRequiredSessionId('request lease renewal'));
    request.setRequestId(requestId);
    request.setMachineId(machineId);
    request.setLeaseEpoch(leaseEpoch);

    const response = await this.invokeGRPCUnary<ProtoRequest>(
      (metadata, options, callback) => this.requestsClient!.renewRequestLease(request, metadata, options, callback),
      `failed to renew lease for request ${requestId}`,
      requestId,
    );

    return this.normalizeRequest(response);
  }

  async cancelRequest(requestId: string): Promise<boolean> {
    this.ensureGRPCConnected('request cancellation');

    const request = new CancelRequestMessage();
    request.setSessionId(this.getRequiredSessionId('request cancellation'));
    request.setRequestId(requestId);

    const response = await this.invokeGRPCUnary<CancelRequestResponseMessage>(
      (metadata, options, callback) => this.requestsClient!.cancelRequest(request, metadata, options, callback),
      `failed to cancel request ${requestId}`,
      requestId,
    );

    return response.getSuccess();
  }

  providerRuntime(options: ProviderRuntimeOptions = {}): ProviderRuntime {
    return new ProviderRuntime(this, options);
  }

  forkSession(sessionId: string): ToolplaneClient {
    return new ToolplaneClient({
      ...this.config,
      sessionId,
    });
  }

  static createGRPCClient(
    serverHost: string,
    serverPort: number,
    sessionId: string,
    userId: string,
    apiKey?: string,
    tls?: GRPCTLSConfig,
  ): ToolplaneClient {
    return new ToolplaneClient({
      protocol: ClientProtocol.GRPC,
      serverHost,
      serverPort,
      sessionId,
      userId,
      apiKey,
      tls,
    });
  }

  private async pingGRPC(): Promise<string> {
    this.ensureGRPCConnected('ping');

    const request = new HealthCheckRequest();
    const response = await this.invokeGRPCUnary<HealthCheckResponseMessage>(
      (metadata, options, callback) => this.toolClient!.healthCheck(request, metadata, options, callback),
      'gRPC ping failed',
    );

    return response.getStatus();
  }

  /**
   * Replays the request's retained chunks that follow lastSeq and streams
   * live chunks until the final marker. Resolving rejects with an
   * OutOfRange-mapped error when the retained window has moved past lastSeq.
   */
  async resumeStream(
    requestId: string,
    lastSeq: number = 0,
    onChunk?: (chunk: ExecuteToolChunkModel) => void,
  ): Promise<ExecuteToolChunkModel[]> {
    this.ensureGRPCConnected('stream resume');

    const request = new ResumeStreamMessage();
    request.setRequestId(requestId);
    request.setLastSeq(lastSeq);

    const stream = this.toolClient!.resumeStream(request, this.createMetadata());
    const chunks: ExecuteToolChunkModel[] = [];

    return new Promise((resolve, reject) => {
      stream.on('data', (chunk: ExecuteToolChunkMessage) => {
        const normalized = {
          seq: chunk.getSeq(),
          requestId: chunk.getRequestId(),
          chunk: chunk.getChunk(),
          isFinal: chunk.getIsFinal(),
          error: chunk.getError(),
        };
        chunks.push(normalized);
        onChunk?.(normalized);
      });
      stream.on('error', (error: grpc.ServiceError) => {
        reject(this.translateGRPCError(error, `failed to resume stream for request ${requestId}`, requestId));
      });
      stream.on('end', () => resolve(chunks));
    });
  }

  private async executeToolGRPC(
    toolName: string,
    params: Record<string, unknown>,
    idempotencyKey: string = '',
  ): Promise<ProtoRequest> {
    this.ensureGRPCConnected('tool execution');

    const request = new ExecuteToolMessage();
    request.setSessionId(this.getRequiredSessionId('tool execution'));
    request.setToolName(toolName);
    request.setInput(JSON.stringify(params));
    request.setIdempotencyKey(idempotencyKey);

    const response = await this.invokeGRPCUnary<ExecuteToolResponseMessage>(
      (metadata, options, callback) => this.toolClient!.invokeTool(request, metadata, options, callback),
      `failed to execute tool ${toolName}`,
    );

    if (response.getError()) {
      throw new ToolplaneError(`Tool execution failed: ${response.getError()}`);
    }

    const requestId = response.getRequestId();
    if (!requestId) {
      throw new ProtocolError('Tool execution did not return a request ID');
    }

    return this.waitForRequestCompletion(this.getRequiredSessionId('tool execution'), requestId);
  }

  private async waitForRequestCompletion(sessionId: string, requestId: string): Promise<ProtoRequest> {
    this.ensureGRPCConnected('request polling');

    const deadline = Date.now() + this.timeoutMs;
    while (Date.now() < deadline) {
      const request = new GetRequestRequest();
      request.setSessionId(sessionId);
      request.setRequestId(requestId);

      const response = await this.invokeGRPCUnary<ProtoRequest>(
        (metadata, options, callback) => this.requestsClient!.getRequest(request, metadata, options, callback),
        `failed to fetch request ${requestId}`,
        requestId,
      );

      switch (response.getStatus()) {
        case RequestStatus.REQUEST_STATUS_DONE:
          return response;
        case RequestStatus.REQUEST_STATUS_FAILED: {
          const errorMessage = response.getError() || `request ${requestId} failed`;
          throw new FailedPreconditionError(errorMessage, { requestId, status: 'failure' });
        }
        default:
          await sleep(REQUEST_POLL_INTERVAL_MS);
      }
    }

    throw new TimeoutError(`Timed out waiting for request ${requestId}`);
  }

  private ensureGRPCClientsInitialized(): void {
    if (!this.toolClient || !this.sessionClient || !this.machineClient || !this.requestsClient || !this.tasksClient) {
      throw new ConnectionError('gRPC clients not initialized');
    }
  }

  private ensureGRPCConnected(operation: string): void {
    if (this.config.protocol !== ClientProtocol.GRPC) {
      throw new ProtocolError(`${operation} only supported with gRPC protocol`);
    }

    this.ensureGRPCClientsInitialized();

    if (!this.connected) {
      throw new ConnectionError('gRPC client not connected to server');
    }
  }

  private getRequiredSessionId(operation: string): string {
    const sessionId = this.config.sessionId.trim();
    if (!sessionId) {
      throw new ProtocolError(`${operation} requires a session ID`);
    }
    return sessionId;
  }

  private getRequiredMachineId(operation: string): string {
    if (!this.machineId) {
      throw new ProtocolError(`${operation} requires a registered machine; call registerMachine() first`);
    }
    return this.machineId;
  }

  private createMetadata(): grpc.Metadata {
    const metadata = new grpc.Metadata();
    if (this.config.apiKey) {
      metadata.set('api_key', this.config.apiKey);
      metadata.set('authorization', `Bearer ${this.config.apiKey}`);
    }
    // Per-machine credential for provide-scoped RPCs; sent on every call and
    // only consulted by the server on provider operations.
    if (this.machineToken) {
      metadata.set('x-toolplane-machine-token', this.machineToken);
    }
    return metadata;
  }

  private grpcCallOptions(timeoutMs: number = this.timeoutMs): Partial<grpc.CallOptions> {
    return {
      deadline: Date.now() + timeoutMs,
    };
  }

  private waitForClientReady(client: grpc.Client): Promise<void> {
    return new Promise((resolve, reject) => {
      client.waitForReady(Date.now() + this.timeoutMs, (error) => {
        if (error) {
          reject(new ConnectionError(`failed to connect to gRPC server: ${error.message}`, error));
          return;
        }
        resolve();
      });
    });
  }

  private invokeGRPCUnary<T>(
    callFactory: (
      metadata: grpc.Metadata,
      options: Partial<grpc.CallOptions>,
      callback: (error: grpc.ServiceError | null, response: T) => void,
    ) => grpc.ClientUnaryCall,
    context: string,
    requestId?: string,
  ): Promise<T> {
    return new Promise((resolve, reject) => {
      callFactory(this.createMetadata(), this.grpcCallOptions(), (error, response) => {
        if (error) {
          reject(this.translateGRPCError(error, context, requestId));
          return;
        }

        resolve(response);
      });
    });
  }

  // Maps a gRPC failure to a typed error: one class per status so callers can
  // branch with instanceof instead of parsing messages. Deterministic codes
  // (auth, arguments, state conflicts, missing entities) are never retryable.
  private translateGRPCError(error: grpc.ServiceError, context: string, requestId?: string): ToolplaneError {
    const message = error.details || error.message || context;
    const ctx = requestId ? { requestId, data: error } : { data: error };
    const prefixed = `${context}: ${message}`;

    switch (error.code) {
      case grpc.status.NOT_FOUND:
        return new NotFoundError(prefixed, ctx);
      case grpc.status.INVALID_ARGUMENT:
      case grpc.status.OUT_OF_RANGE:
        return new InvalidArgumentError(prefixed, ctx);
      case grpc.status.FAILED_PRECONDITION:
      case grpc.status.ABORTED:
        return new FailedPreconditionError(prefixed, ctx);
      case grpc.status.RESOURCE_EXHAUSTED:
        return new ResourceExhaustedError(prefixed, ctx);
      case grpc.status.UNAUTHENTICATED:
        return new UnauthenticatedError(prefixed, ctx);
      case grpc.status.PERMISSION_DENIED:
        return new PermissionDeniedError(prefixed, ctx);
      case grpc.status.ALREADY_EXISTS:
        return new AlreadyExistsError(prefixed, ctx);
      case grpc.status.UNAVAILABLE:
        return new UnavailableError(prefixed, ctx);
      case grpc.status.DEADLINE_EXCEEDED:
        return new TimeoutError(prefixed, error);
      case grpc.status.CANCELLED:
        return new ConnectionError(prefixed, error);
      default:
        return new ToolplaneError(prefixed, error.code, error);
    }
  }

  private parseResultPayload(value: string): unknown {
    if (!value) {
      return undefined;
    }

    try {
      return JSON.parse(value);
    } catch {
      return value;
    }
  }

  private serializePayload(value: unknown): string {
    if (typeof value === 'string') {
      return value;
    }

    return JSON.stringify(value ?? null);
  }

  private normalizeTool(tool: ProtoTool | undefined): Tool {
    if (!tool) {
      throw new ProtocolError('Tool response was empty');
    }

    const config: Record<string, string> = {};
    tool.getConfigMap().forEach((value: string, key: string) => {
      config[key] = value;
    });

    return {
      id: tool.getId(),
      name: tool.getName(),
      description: tool.getDescription(),
      schema: tool.getSchema(),
      config,
      createdAt: timestampToIso(tool.getCreatedAt()),
      lastPingAt: timestampToIso(tool.getLastPingAt()) || undefined,
      sessionId: tool.getSessionId(),
      tags: tool.getTagsList(),
    };
  }

  private normalizeSession(session: ProtoSession | undefined): Session {
    if (!session) {
      throw new ProtocolError('Session response was empty');
    }

    return {
      id: session.getId(),
      name: session.getName(),
      description: session.getDescription(),
      createdAt: timestampToIso(session.getCreatedAt()),
      createdBy: session.getCreatedBy(),
      namespace: session.getNamespace(),
    };
  }

  private normalizeApiKey(apiKey: ProtoApiKey | undefined): ApiKey {
    if (!apiKey) {
      throw new ProtocolError('API key response was empty');
    }

    return {
      id: apiKey.getId(),
      name: apiKey.getName(),
      key: apiKey.getKey(),
      keyPreview: apiKey.getKeyPreview() || undefined,
      sessionId: apiKey.getSessionId(),
      createdAt: timestampToIso(apiKey.getCreatedAt()),
      createdBy: apiKey.getCreatedBy(),
      capabilities: apiKey.getCapabilitiesList(),
      revokedAt: timestampToIso(apiKey.getRevokedAt()) || undefined,
    };
  }

  private normalizeMachine(machine: ProtoMachine | undefined): Machine {
    if (!machine) {
      throw new ProtocolError('Machine response was empty');
    }

    return {
      id: machine.getId(),
      sessionId: machine.getSessionId(),
      sdkVersion: machine.getSdkVersion(),
      sdkLanguage: machine.getSdkLanguage(),
      ip: machine.getIp(),
      createdAt: timestampToIso(machine.getCreatedAt()),
      lastPingAt: timestampToIso(machine.getLastPingAt()) || undefined,
    };
  }

  private normalizeTask(task: ProtoTask | undefined): Task {
    if (!task) {
      throw new ProtocolError('Task response was empty');
    }

    return {
      id: task.getId(),
      sessionId: task.getSessionId(),
      toolName: task.getToolName(),
      status: normalizeTaskStatus(task.getStatus()),
      input: task.getInput(),
      result: task.getResult(),
      resultType: task.getResultType(),
      error: task.getError(),
      createdAt: timestampToIso(task.getCreatedAt()),
      updatedAt: timestampToIso(task.getUpdatedAt()),
      completedAt: timestampToIso(task.getCompletedAt()) || undefined,
    };
  }

  private normalizeRequest(request: ProtoRequest | undefined): RequestModel {
    if (!request) {
      throw new ProtocolError('Request response was empty');
    }

    const normalized: RequestModel = {
      id: request.getId(),
      sessionId: request.getSessionId(),
      toolName: request.getToolName(),
      status: normalizeRequestStatus(request.getStatus()),
      input: request.getInput(),
      createdAt: timestampToIso(request.getCreatedAt()),
      updatedAt: timestampToIso(request.getUpdatedAt()),
      executingMachineId: request.getExecutingMachineId(),
      leasedBy: request.getLeasedBy(),
      leaseEpoch: request.getLeaseEpoch(),
      leaseExpiresAt: timestampToIso(request.getLeaseExpiresAt()) || undefined,
      timeoutSeconds: request.getTimeoutSeconds(),
    };

    const parsedResult = this.parseResultPayload(request.getResult());
    if (parsedResult !== undefined) {
      normalized.result = parsedResult;
    }

    if (request.getResultType()) {
      normalized.resultType = request.getResultType();
    }

    if (request.getError()) {
      normalized.error = request.getError();
    }

    return normalized;
  }
}