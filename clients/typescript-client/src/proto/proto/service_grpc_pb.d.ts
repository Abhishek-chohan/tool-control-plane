// package: api
// file: proto/service.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import * as proto_service_pb from "../proto/service_pb";

interface IToolServiceService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    registerTool: IToolServiceService_IRegisterTool;
    listTools: IToolServiceService_IListTools;
    getToolById: IToolServiceService_IGetToolById;
    getToolByName: IToolServiceService_IGetToolByName;
    deleteTool: IToolServiceService_IDeleteTool;
    updateToolPing: IToolServiceService_IUpdateToolPing;
    streamExecuteTool: IToolServiceService_IStreamExecuteTool;
    resumeStream: IToolServiceService_IResumeStream;
    executeTool: IToolServiceService_IExecuteTool;
    healthCheck: IToolServiceService_IHealthCheck;
}

interface IToolServiceService_IRegisterTool extends grpc.MethodDefinition<proto_service_pb.RegisterToolRequest, proto_service_pb.RegisterToolResponse> {
    path: "/api.ToolService/RegisterTool";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.RegisterToolRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.RegisterToolRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.RegisterToolResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.RegisterToolResponse>;
}
interface IToolServiceService_IListTools extends grpc.MethodDefinition<proto_service_pb.ListToolsRequest, proto_service_pb.ListToolsResponse> {
    path: "/api.ToolService/ListTools";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.ListToolsRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.ListToolsRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.ListToolsResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.ListToolsResponse>;
}
interface IToolServiceService_IGetToolById extends grpc.MethodDefinition<proto_service_pb.GetToolByIdRequest, proto_service_pb.GetToolResponse> {
    path: "/api.ToolService/GetToolById";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.GetToolByIdRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.GetToolByIdRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.GetToolResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.GetToolResponse>;
}
interface IToolServiceService_IGetToolByName extends grpc.MethodDefinition<proto_service_pb.GetToolByNameRequest, proto_service_pb.GetToolResponse> {
    path: "/api.ToolService/GetToolByName";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.GetToolByNameRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.GetToolByNameRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.GetToolResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.GetToolResponse>;
}
interface IToolServiceService_IDeleteTool extends grpc.MethodDefinition<proto_service_pb.DeleteToolRequest, proto_service_pb.DeleteToolResponse> {
    path: "/api.ToolService/DeleteTool";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.DeleteToolRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.DeleteToolRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.DeleteToolResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.DeleteToolResponse>;
}
interface IToolServiceService_IUpdateToolPing extends grpc.MethodDefinition<proto_service_pb.UpdateToolPingRequest, proto_service_pb.Tool> {
    path: "/api.ToolService/UpdateToolPing";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.UpdateToolPingRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.UpdateToolPingRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.Tool>;
    responseDeserialize: grpc.deserialize<proto_service_pb.Tool>;
}
interface IToolServiceService_IStreamExecuteTool extends grpc.MethodDefinition<proto_service_pb.ExecuteToolRequest, proto_service_pb.ExecuteToolChunk> {
    path: "/api.ToolService/StreamExecuteTool";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<proto_service_pb.ExecuteToolRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.ExecuteToolRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.ExecuteToolChunk>;
    responseDeserialize: grpc.deserialize<proto_service_pb.ExecuteToolChunk>;
}
interface IToolServiceService_IResumeStream extends grpc.MethodDefinition<proto_service_pb.ResumeStreamRequest, proto_service_pb.ExecuteToolChunk> {
    path: "/api.ToolService/ResumeStream";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<proto_service_pb.ResumeStreamRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.ResumeStreamRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.ExecuteToolChunk>;
    responseDeserialize: grpc.deserialize<proto_service_pb.ExecuteToolChunk>;
}
interface IToolServiceService_IExecuteTool extends grpc.MethodDefinition<proto_service_pb.ExecuteToolRequest, proto_service_pb.ExecuteToolResponse> {
    path: "/api.ToolService/ExecuteTool";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.ExecuteToolRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.ExecuteToolRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.ExecuteToolResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.ExecuteToolResponse>;
}
interface IToolServiceService_IHealthCheck extends grpc.MethodDefinition<proto_service_pb.HealthCheckRequest, proto_service_pb.HealthCheckResponse> {
    path: "/api.ToolService/HealthCheck";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.HealthCheckRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.HealthCheckRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.HealthCheckResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.HealthCheckResponse>;
}

export const ToolServiceService: IToolServiceService;

export interface IToolServiceServer extends grpc.UntypedServiceImplementation {
    registerTool: grpc.handleUnaryCall<proto_service_pb.RegisterToolRequest, proto_service_pb.RegisterToolResponse>;
    listTools: grpc.handleUnaryCall<proto_service_pb.ListToolsRequest, proto_service_pb.ListToolsResponse>;
    getToolById: grpc.handleUnaryCall<proto_service_pb.GetToolByIdRequest, proto_service_pb.GetToolResponse>;
    getToolByName: grpc.handleUnaryCall<proto_service_pb.GetToolByNameRequest, proto_service_pb.GetToolResponse>;
    deleteTool: grpc.handleUnaryCall<proto_service_pb.DeleteToolRequest, proto_service_pb.DeleteToolResponse>;
    updateToolPing: grpc.handleUnaryCall<proto_service_pb.UpdateToolPingRequest, proto_service_pb.Tool>;
    streamExecuteTool: grpc.handleServerStreamingCall<proto_service_pb.ExecuteToolRequest, proto_service_pb.ExecuteToolChunk>;
    resumeStream: grpc.handleServerStreamingCall<proto_service_pb.ResumeStreamRequest, proto_service_pb.ExecuteToolChunk>;
    executeTool: grpc.handleUnaryCall<proto_service_pb.ExecuteToolRequest, proto_service_pb.ExecuteToolResponse>;
    healthCheck: grpc.handleUnaryCall<proto_service_pb.HealthCheckRequest, proto_service_pb.HealthCheckResponse>;
}

export interface IToolServiceClient {
    registerTool(request: proto_service_pb.RegisterToolRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.RegisterToolResponse) => void): grpc.ClientUnaryCall;
    registerTool(request: proto_service_pb.RegisterToolRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.RegisterToolResponse) => void): grpc.ClientUnaryCall;
    registerTool(request: proto_service_pb.RegisterToolRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.RegisterToolResponse) => void): grpc.ClientUnaryCall;
    listTools(request: proto_service_pb.ListToolsRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListToolsResponse) => void): grpc.ClientUnaryCall;
    listTools(request: proto_service_pb.ListToolsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListToolsResponse) => void): grpc.ClientUnaryCall;
    listTools(request: proto_service_pb.ListToolsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListToolsResponse) => void): grpc.ClientUnaryCall;
    getToolById(request: proto_service_pb.GetToolByIdRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetToolResponse) => void): grpc.ClientUnaryCall;
    getToolById(request: proto_service_pb.GetToolByIdRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetToolResponse) => void): grpc.ClientUnaryCall;
    getToolById(request: proto_service_pb.GetToolByIdRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetToolResponse) => void): grpc.ClientUnaryCall;
    getToolByName(request: proto_service_pb.GetToolByNameRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetToolResponse) => void): grpc.ClientUnaryCall;
    getToolByName(request: proto_service_pb.GetToolByNameRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetToolResponse) => void): grpc.ClientUnaryCall;
    getToolByName(request: proto_service_pb.GetToolByNameRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetToolResponse) => void): grpc.ClientUnaryCall;
    deleteTool(request: proto_service_pb.DeleteToolRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.DeleteToolResponse) => void): grpc.ClientUnaryCall;
    deleteTool(request: proto_service_pb.DeleteToolRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.DeleteToolResponse) => void): grpc.ClientUnaryCall;
    deleteTool(request: proto_service_pb.DeleteToolRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.DeleteToolResponse) => void): grpc.ClientUnaryCall;
    updateToolPing(request: proto_service_pb.UpdateToolPingRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Tool) => void): grpc.ClientUnaryCall;
    updateToolPing(request: proto_service_pb.UpdateToolPingRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Tool) => void): grpc.ClientUnaryCall;
    updateToolPing(request: proto_service_pb.UpdateToolPingRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Tool) => void): grpc.ClientUnaryCall;
    streamExecuteTool(request: proto_service_pb.ExecuteToolRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<proto_service_pb.ExecuteToolChunk>;
    streamExecuteTool(request: proto_service_pb.ExecuteToolRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<proto_service_pb.ExecuteToolChunk>;
    resumeStream(request: proto_service_pb.ResumeStreamRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<proto_service_pb.ExecuteToolChunk>;
    resumeStream(request: proto_service_pb.ResumeStreamRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<proto_service_pb.ExecuteToolChunk>;
    executeTool(request: proto_service_pb.ExecuteToolRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ExecuteToolResponse) => void): grpc.ClientUnaryCall;
    executeTool(request: proto_service_pb.ExecuteToolRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ExecuteToolResponse) => void): grpc.ClientUnaryCall;
    executeTool(request: proto_service_pb.ExecuteToolRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ExecuteToolResponse) => void): grpc.ClientUnaryCall;
    healthCheck(request: proto_service_pb.HealthCheckRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.HealthCheckResponse) => void): grpc.ClientUnaryCall;
    healthCheck(request: proto_service_pb.HealthCheckRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.HealthCheckResponse) => void): grpc.ClientUnaryCall;
    healthCheck(request: proto_service_pb.HealthCheckRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.HealthCheckResponse) => void): grpc.ClientUnaryCall;
}

export class ToolServiceClient extends grpc.Client implements IToolServiceClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public registerTool(request: proto_service_pb.RegisterToolRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.RegisterToolResponse) => void): grpc.ClientUnaryCall;
    public registerTool(request: proto_service_pb.RegisterToolRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.RegisterToolResponse) => void): grpc.ClientUnaryCall;
    public registerTool(request: proto_service_pb.RegisterToolRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.RegisterToolResponse) => void): grpc.ClientUnaryCall;
    public listTools(request: proto_service_pb.ListToolsRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListToolsResponse) => void): grpc.ClientUnaryCall;
    public listTools(request: proto_service_pb.ListToolsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListToolsResponse) => void): grpc.ClientUnaryCall;
    public listTools(request: proto_service_pb.ListToolsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListToolsResponse) => void): grpc.ClientUnaryCall;
    public getToolById(request: proto_service_pb.GetToolByIdRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetToolResponse) => void): grpc.ClientUnaryCall;
    public getToolById(request: proto_service_pb.GetToolByIdRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetToolResponse) => void): grpc.ClientUnaryCall;
    public getToolById(request: proto_service_pb.GetToolByIdRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetToolResponse) => void): grpc.ClientUnaryCall;
    public getToolByName(request: proto_service_pb.GetToolByNameRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetToolResponse) => void): grpc.ClientUnaryCall;
    public getToolByName(request: proto_service_pb.GetToolByNameRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetToolResponse) => void): grpc.ClientUnaryCall;
    public getToolByName(request: proto_service_pb.GetToolByNameRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetToolResponse) => void): grpc.ClientUnaryCall;
    public deleteTool(request: proto_service_pb.DeleteToolRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.DeleteToolResponse) => void): grpc.ClientUnaryCall;
    public deleteTool(request: proto_service_pb.DeleteToolRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.DeleteToolResponse) => void): grpc.ClientUnaryCall;
    public deleteTool(request: proto_service_pb.DeleteToolRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.DeleteToolResponse) => void): grpc.ClientUnaryCall;
    public updateToolPing(request: proto_service_pb.UpdateToolPingRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Tool) => void): grpc.ClientUnaryCall;
    public updateToolPing(request: proto_service_pb.UpdateToolPingRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Tool) => void): grpc.ClientUnaryCall;
    public updateToolPing(request: proto_service_pb.UpdateToolPingRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Tool) => void): grpc.ClientUnaryCall;
    public streamExecuteTool(request: proto_service_pb.ExecuteToolRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<proto_service_pb.ExecuteToolChunk>;
    public streamExecuteTool(request: proto_service_pb.ExecuteToolRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<proto_service_pb.ExecuteToolChunk>;
    public resumeStream(request: proto_service_pb.ResumeStreamRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<proto_service_pb.ExecuteToolChunk>;
    public resumeStream(request: proto_service_pb.ResumeStreamRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<proto_service_pb.ExecuteToolChunk>;
    public executeTool(request: proto_service_pb.ExecuteToolRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ExecuteToolResponse) => void): grpc.ClientUnaryCall;
    public executeTool(request: proto_service_pb.ExecuteToolRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ExecuteToolResponse) => void): grpc.ClientUnaryCall;
    public executeTool(request: proto_service_pb.ExecuteToolRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ExecuteToolResponse) => void): grpc.ClientUnaryCall;
    public healthCheck(request: proto_service_pb.HealthCheckRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.HealthCheckResponse) => void): grpc.ClientUnaryCall;
    public healthCheck(request: proto_service_pb.HealthCheckRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.HealthCheckResponse) => void): grpc.ClientUnaryCall;
    public healthCheck(request: proto_service_pb.HealthCheckRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.HealthCheckResponse) => void): grpc.ClientUnaryCall;
}

interface ISessionsServiceService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    createSession: ISessionsServiceService_ICreateSession;
    getSession: ISessionsServiceService_IGetSession;
    listSessions: ISessionsServiceService_IListSessions;
    updateSession: ISessionsServiceService_IUpdateSession;
    deleteSession: ISessionsServiceService_IDeleteSession;
    listUserSessions: ISessionsServiceService_IListUserSessions;
    bulkDeleteSessions: ISessionsServiceService_IBulkDeleteSessions;
    getSessionStats: ISessionsServiceService_IGetSessionStats;
    refreshSessionToken: ISessionsServiceService_IRefreshSessionToken;
    invalidateSession: ISessionsServiceService_IInvalidateSession;
    createApiKey: ISessionsServiceService_ICreateApiKey;
    listApiKeys: ISessionsServiceService_IListApiKeys;
    revokeApiKey: ISessionsServiceService_IRevokeApiKey;
}

interface ISessionsServiceService_ICreateSession extends grpc.MethodDefinition<proto_service_pb.CreateSessionRequest, proto_service_pb.CreateSessionResponse> {
    path: "/api.SessionsService/CreateSession";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.CreateSessionRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.CreateSessionRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.CreateSessionResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.CreateSessionResponse>;
}
interface ISessionsServiceService_IGetSession extends grpc.MethodDefinition<proto_service_pb.GetSessionRequest, proto_service_pb.Session> {
    path: "/api.SessionsService/GetSession";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.GetSessionRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.GetSessionRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.Session>;
    responseDeserialize: grpc.deserialize<proto_service_pb.Session>;
}
interface ISessionsServiceService_IListSessions extends grpc.MethodDefinition<proto_service_pb.ListSessionsRequest, proto_service_pb.ListSessionsResponse> {
    path: "/api.SessionsService/ListSessions";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.ListSessionsRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.ListSessionsRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.ListSessionsResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.ListSessionsResponse>;
}
interface ISessionsServiceService_IUpdateSession extends grpc.MethodDefinition<proto_service_pb.UpdateSessionRequest, proto_service_pb.Session> {
    path: "/api.SessionsService/UpdateSession";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.UpdateSessionRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.UpdateSessionRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.Session>;
    responseDeserialize: grpc.deserialize<proto_service_pb.Session>;
}
interface ISessionsServiceService_IDeleteSession extends grpc.MethodDefinition<proto_service_pb.DeleteSessionRequest, proto_service_pb.DeleteSessionResponse> {
    path: "/api.SessionsService/DeleteSession";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.DeleteSessionRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.DeleteSessionRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.DeleteSessionResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.DeleteSessionResponse>;
}
interface ISessionsServiceService_IListUserSessions extends grpc.MethodDefinition<proto_service_pb.ListUserSessionsRequest, proto_service_pb.ListUserSessionsResponse> {
    path: "/api.SessionsService/ListUserSessions";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.ListUserSessionsRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.ListUserSessionsRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.ListUserSessionsResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.ListUserSessionsResponse>;
}
interface ISessionsServiceService_IBulkDeleteSessions extends grpc.MethodDefinition<proto_service_pb.BulkDeleteSessionsRequest, proto_service_pb.BulkDeleteSessionsResponse> {
    path: "/api.SessionsService/BulkDeleteSessions";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.BulkDeleteSessionsRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.BulkDeleteSessionsRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.BulkDeleteSessionsResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.BulkDeleteSessionsResponse>;
}
interface ISessionsServiceService_IGetSessionStats extends grpc.MethodDefinition<proto_service_pb.GetSessionStatsRequest, proto_service_pb.GetSessionStatsResponse> {
    path: "/api.SessionsService/GetSessionStats";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.GetSessionStatsRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.GetSessionStatsRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.GetSessionStatsResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.GetSessionStatsResponse>;
}
interface ISessionsServiceService_IRefreshSessionToken extends grpc.MethodDefinition<proto_service_pb.RefreshSessionTokenRequest, proto_service_pb.RefreshSessionTokenResponse> {
    path: "/api.SessionsService/RefreshSessionToken";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.RefreshSessionTokenRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.RefreshSessionTokenRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.RefreshSessionTokenResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.RefreshSessionTokenResponse>;
}
interface ISessionsServiceService_IInvalidateSession extends grpc.MethodDefinition<proto_service_pb.InvalidateSessionRequest, proto_service_pb.InvalidateSessionResponse> {
    path: "/api.SessionsService/InvalidateSession";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.InvalidateSessionRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.InvalidateSessionRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.InvalidateSessionResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.InvalidateSessionResponse>;
}
interface ISessionsServiceService_ICreateApiKey extends grpc.MethodDefinition<proto_service_pb.CreateApiKeyRequest, proto_service_pb.ApiKey> {
    path: "/api.SessionsService/CreateApiKey";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.CreateApiKeyRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.CreateApiKeyRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.ApiKey>;
    responseDeserialize: grpc.deserialize<proto_service_pb.ApiKey>;
}
interface ISessionsServiceService_IListApiKeys extends grpc.MethodDefinition<proto_service_pb.ListApiKeysRequest, proto_service_pb.ListApiKeysResponse> {
    path: "/api.SessionsService/ListApiKeys";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.ListApiKeysRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.ListApiKeysRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.ListApiKeysResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.ListApiKeysResponse>;
}
interface ISessionsServiceService_IRevokeApiKey extends grpc.MethodDefinition<proto_service_pb.RevokeApiKeyRequest, proto_service_pb.RevokeApiKeyResponse> {
    path: "/api.SessionsService/RevokeApiKey";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.RevokeApiKeyRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.RevokeApiKeyRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.RevokeApiKeyResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.RevokeApiKeyResponse>;
}

export const SessionsServiceService: ISessionsServiceService;

export interface ISessionsServiceServer extends grpc.UntypedServiceImplementation {
    createSession: grpc.handleUnaryCall<proto_service_pb.CreateSessionRequest, proto_service_pb.CreateSessionResponse>;
    getSession: grpc.handleUnaryCall<proto_service_pb.GetSessionRequest, proto_service_pb.Session>;
    listSessions: grpc.handleUnaryCall<proto_service_pb.ListSessionsRequest, proto_service_pb.ListSessionsResponse>;
    updateSession: grpc.handleUnaryCall<proto_service_pb.UpdateSessionRequest, proto_service_pb.Session>;
    deleteSession: grpc.handleUnaryCall<proto_service_pb.DeleteSessionRequest, proto_service_pb.DeleteSessionResponse>;
    listUserSessions: grpc.handleUnaryCall<proto_service_pb.ListUserSessionsRequest, proto_service_pb.ListUserSessionsResponse>;
    bulkDeleteSessions: grpc.handleUnaryCall<proto_service_pb.BulkDeleteSessionsRequest, proto_service_pb.BulkDeleteSessionsResponse>;
    getSessionStats: grpc.handleUnaryCall<proto_service_pb.GetSessionStatsRequest, proto_service_pb.GetSessionStatsResponse>;
    refreshSessionToken: grpc.handleUnaryCall<proto_service_pb.RefreshSessionTokenRequest, proto_service_pb.RefreshSessionTokenResponse>;
    invalidateSession: grpc.handleUnaryCall<proto_service_pb.InvalidateSessionRequest, proto_service_pb.InvalidateSessionResponse>;
    createApiKey: grpc.handleUnaryCall<proto_service_pb.CreateApiKeyRequest, proto_service_pb.ApiKey>;
    listApiKeys: grpc.handleUnaryCall<proto_service_pb.ListApiKeysRequest, proto_service_pb.ListApiKeysResponse>;
    revokeApiKey: grpc.handleUnaryCall<proto_service_pb.RevokeApiKeyRequest, proto_service_pb.RevokeApiKeyResponse>;
}

export interface ISessionsServiceClient {
    createSession(request: proto_service_pb.CreateSessionRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.CreateSessionResponse) => void): grpc.ClientUnaryCall;
    createSession(request: proto_service_pb.CreateSessionRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.CreateSessionResponse) => void): grpc.ClientUnaryCall;
    createSession(request: proto_service_pb.CreateSessionRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.CreateSessionResponse) => void): grpc.ClientUnaryCall;
    getSession(request: proto_service_pb.GetSessionRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Session) => void): grpc.ClientUnaryCall;
    getSession(request: proto_service_pb.GetSessionRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Session) => void): grpc.ClientUnaryCall;
    getSession(request: proto_service_pb.GetSessionRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Session) => void): grpc.ClientUnaryCall;
    listSessions(request: proto_service_pb.ListSessionsRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListSessionsResponse) => void): grpc.ClientUnaryCall;
    listSessions(request: proto_service_pb.ListSessionsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListSessionsResponse) => void): grpc.ClientUnaryCall;
    listSessions(request: proto_service_pb.ListSessionsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListSessionsResponse) => void): grpc.ClientUnaryCall;
    updateSession(request: proto_service_pb.UpdateSessionRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Session) => void): grpc.ClientUnaryCall;
    updateSession(request: proto_service_pb.UpdateSessionRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Session) => void): grpc.ClientUnaryCall;
    updateSession(request: proto_service_pb.UpdateSessionRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Session) => void): grpc.ClientUnaryCall;
    deleteSession(request: proto_service_pb.DeleteSessionRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.DeleteSessionResponse) => void): grpc.ClientUnaryCall;
    deleteSession(request: proto_service_pb.DeleteSessionRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.DeleteSessionResponse) => void): grpc.ClientUnaryCall;
    deleteSession(request: proto_service_pb.DeleteSessionRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.DeleteSessionResponse) => void): grpc.ClientUnaryCall;
    listUserSessions(request: proto_service_pb.ListUserSessionsRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListUserSessionsResponse) => void): grpc.ClientUnaryCall;
    listUserSessions(request: proto_service_pb.ListUserSessionsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListUserSessionsResponse) => void): grpc.ClientUnaryCall;
    listUserSessions(request: proto_service_pb.ListUserSessionsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListUserSessionsResponse) => void): grpc.ClientUnaryCall;
    bulkDeleteSessions(request: proto_service_pb.BulkDeleteSessionsRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.BulkDeleteSessionsResponse) => void): grpc.ClientUnaryCall;
    bulkDeleteSessions(request: proto_service_pb.BulkDeleteSessionsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.BulkDeleteSessionsResponse) => void): grpc.ClientUnaryCall;
    bulkDeleteSessions(request: proto_service_pb.BulkDeleteSessionsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.BulkDeleteSessionsResponse) => void): grpc.ClientUnaryCall;
    getSessionStats(request: proto_service_pb.GetSessionStatsRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetSessionStatsResponse) => void): grpc.ClientUnaryCall;
    getSessionStats(request: proto_service_pb.GetSessionStatsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetSessionStatsResponse) => void): grpc.ClientUnaryCall;
    getSessionStats(request: proto_service_pb.GetSessionStatsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetSessionStatsResponse) => void): grpc.ClientUnaryCall;
    refreshSessionToken(request: proto_service_pb.RefreshSessionTokenRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.RefreshSessionTokenResponse) => void): grpc.ClientUnaryCall;
    refreshSessionToken(request: proto_service_pb.RefreshSessionTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.RefreshSessionTokenResponse) => void): grpc.ClientUnaryCall;
    refreshSessionToken(request: proto_service_pb.RefreshSessionTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.RefreshSessionTokenResponse) => void): grpc.ClientUnaryCall;
    invalidateSession(request: proto_service_pb.InvalidateSessionRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.InvalidateSessionResponse) => void): grpc.ClientUnaryCall;
    invalidateSession(request: proto_service_pb.InvalidateSessionRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.InvalidateSessionResponse) => void): grpc.ClientUnaryCall;
    invalidateSession(request: proto_service_pb.InvalidateSessionRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.InvalidateSessionResponse) => void): grpc.ClientUnaryCall;
    createApiKey(request: proto_service_pb.CreateApiKeyRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ApiKey) => void): grpc.ClientUnaryCall;
    createApiKey(request: proto_service_pb.CreateApiKeyRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ApiKey) => void): grpc.ClientUnaryCall;
    createApiKey(request: proto_service_pb.CreateApiKeyRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ApiKey) => void): grpc.ClientUnaryCall;
    listApiKeys(request: proto_service_pb.ListApiKeysRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListApiKeysResponse) => void): grpc.ClientUnaryCall;
    listApiKeys(request: proto_service_pb.ListApiKeysRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListApiKeysResponse) => void): grpc.ClientUnaryCall;
    listApiKeys(request: proto_service_pb.ListApiKeysRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListApiKeysResponse) => void): grpc.ClientUnaryCall;
    revokeApiKey(request: proto_service_pb.RevokeApiKeyRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.RevokeApiKeyResponse) => void): grpc.ClientUnaryCall;
    revokeApiKey(request: proto_service_pb.RevokeApiKeyRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.RevokeApiKeyResponse) => void): grpc.ClientUnaryCall;
    revokeApiKey(request: proto_service_pb.RevokeApiKeyRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.RevokeApiKeyResponse) => void): grpc.ClientUnaryCall;
}

export class SessionsServiceClient extends grpc.Client implements ISessionsServiceClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public createSession(request: proto_service_pb.CreateSessionRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.CreateSessionResponse) => void): grpc.ClientUnaryCall;
    public createSession(request: proto_service_pb.CreateSessionRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.CreateSessionResponse) => void): grpc.ClientUnaryCall;
    public createSession(request: proto_service_pb.CreateSessionRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.CreateSessionResponse) => void): grpc.ClientUnaryCall;
    public getSession(request: proto_service_pb.GetSessionRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Session) => void): grpc.ClientUnaryCall;
    public getSession(request: proto_service_pb.GetSessionRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Session) => void): grpc.ClientUnaryCall;
    public getSession(request: proto_service_pb.GetSessionRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Session) => void): grpc.ClientUnaryCall;
    public listSessions(request: proto_service_pb.ListSessionsRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListSessionsResponse) => void): grpc.ClientUnaryCall;
    public listSessions(request: proto_service_pb.ListSessionsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListSessionsResponse) => void): grpc.ClientUnaryCall;
    public listSessions(request: proto_service_pb.ListSessionsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListSessionsResponse) => void): grpc.ClientUnaryCall;
    public updateSession(request: proto_service_pb.UpdateSessionRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Session) => void): grpc.ClientUnaryCall;
    public updateSession(request: proto_service_pb.UpdateSessionRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Session) => void): grpc.ClientUnaryCall;
    public updateSession(request: proto_service_pb.UpdateSessionRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Session) => void): grpc.ClientUnaryCall;
    public deleteSession(request: proto_service_pb.DeleteSessionRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.DeleteSessionResponse) => void): grpc.ClientUnaryCall;
    public deleteSession(request: proto_service_pb.DeleteSessionRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.DeleteSessionResponse) => void): grpc.ClientUnaryCall;
    public deleteSession(request: proto_service_pb.DeleteSessionRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.DeleteSessionResponse) => void): grpc.ClientUnaryCall;
    public listUserSessions(request: proto_service_pb.ListUserSessionsRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListUserSessionsResponse) => void): grpc.ClientUnaryCall;
    public listUserSessions(request: proto_service_pb.ListUserSessionsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListUserSessionsResponse) => void): grpc.ClientUnaryCall;
    public listUserSessions(request: proto_service_pb.ListUserSessionsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListUserSessionsResponse) => void): grpc.ClientUnaryCall;
    public bulkDeleteSessions(request: proto_service_pb.BulkDeleteSessionsRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.BulkDeleteSessionsResponse) => void): grpc.ClientUnaryCall;
    public bulkDeleteSessions(request: proto_service_pb.BulkDeleteSessionsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.BulkDeleteSessionsResponse) => void): grpc.ClientUnaryCall;
    public bulkDeleteSessions(request: proto_service_pb.BulkDeleteSessionsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.BulkDeleteSessionsResponse) => void): grpc.ClientUnaryCall;
    public getSessionStats(request: proto_service_pb.GetSessionStatsRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetSessionStatsResponse) => void): grpc.ClientUnaryCall;
    public getSessionStats(request: proto_service_pb.GetSessionStatsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetSessionStatsResponse) => void): grpc.ClientUnaryCall;
    public getSessionStats(request: proto_service_pb.GetSessionStatsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetSessionStatsResponse) => void): grpc.ClientUnaryCall;
    public refreshSessionToken(request: proto_service_pb.RefreshSessionTokenRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.RefreshSessionTokenResponse) => void): grpc.ClientUnaryCall;
    public refreshSessionToken(request: proto_service_pb.RefreshSessionTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.RefreshSessionTokenResponse) => void): grpc.ClientUnaryCall;
    public refreshSessionToken(request: proto_service_pb.RefreshSessionTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.RefreshSessionTokenResponse) => void): grpc.ClientUnaryCall;
    public invalidateSession(request: proto_service_pb.InvalidateSessionRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.InvalidateSessionResponse) => void): grpc.ClientUnaryCall;
    public invalidateSession(request: proto_service_pb.InvalidateSessionRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.InvalidateSessionResponse) => void): grpc.ClientUnaryCall;
    public invalidateSession(request: proto_service_pb.InvalidateSessionRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.InvalidateSessionResponse) => void): grpc.ClientUnaryCall;
    public createApiKey(request: proto_service_pb.CreateApiKeyRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ApiKey) => void): grpc.ClientUnaryCall;
    public createApiKey(request: proto_service_pb.CreateApiKeyRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ApiKey) => void): grpc.ClientUnaryCall;
    public createApiKey(request: proto_service_pb.CreateApiKeyRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ApiKey) => void): grpc.ClientUnaryCall;
    public listApiKeys(request: proto_service_pb.ListApiKeysRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListApiKeysResponse) => void): grpc.ClientUnaryCall;
    public listApiKeys(request: proto_service_pb.ListApiKeysRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListApiKeysResponse) => void): grpc.ClientUnaryCall;
    public listApiKeys(request: proto_service_pb.ListApiKeysRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListApiKeysResponse) => void): grpc.ClientUnaryCall;
    public revokeApiKey(request: proto_service_pb.RevokeApiKeyRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.RevokeApiKeyResponse) => void): grpc.ClientUnaryCall;
    public revokeApiKey(request: proto_service_pb.RevokeApiKeyRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.RevokeApiKeyResponse) => void): grpc.ClientUnaryCall;
    public revokeApiKey(request: proto_service_pb.RevokeApiKeyRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.RevokeApiKeyResponse) => void): grpc.ClientUnaryCall;
}

interface IMachinesServiceService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    registerMachine: IMachinesServiceService_IRegisterMachine;
    listMachines: IMachinesServiceService_IListMachines;
    getMachine: IMachinesServiceService_IGetMachine;
    updateMachinePing: IMachinesServiceService_IUpdateMachinePing;
    unregisterMachine: IMachinesServiceService_IUnregisterMachine;
    drainMachine: IMachinesServiceService_IDrainMachine;
}

interface IMachinesServiceService_IRegisterMachine extends grpc.MethodDefinition<proto_service_pb.RegisterMachineRequest, proto_service_pb.Machine> {
    path: "/api.MachinesService/RegisterMachine";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.RegisterMachineRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.RegisterMachineRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.Machine>;
    responseDeserialize: grpc.deserialize<proto_service_pb.Machine>;
}
interface IMachinesServiceService_IListMachines extends grpc.MethodDefinition<proto_service_pb.ListMachinesRequest, proto_service_pb.ListMachinesResponse> {
    path: "/api.MachinesService/ListMachines";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.ListMachinesRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.ListMachinesRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.ListMachinesResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.ListMachinesResponse>;
}
interface IMachinesServiceService_IGetMachine extends grpc.MethodDefinition<proto_service_pb.GetMachineRequest, proto_service_pb.Machine> {
    path: "/api.MachinesService/GetMachine";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.GetMachineRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.GetMachineRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.Machine>;
    responseDeserialize: grpc.deserialize<proto_service_pb.Machine>;
}
interface IMachinesServiceService_IUpdateMachinePing extends grpc.MethodDefinition<proto_service_pb.UpdateMachinePingRequest, proto_service_pb.Machine> {
    path: "/api.MachinesService/UpdateMachinePing";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.UpdateMachinePingRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.UpdateMachinePingRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.Machine>;
    responseDeserialize: grpc.deserialize<proto_service_pb.Machine>;
}
interface IMachinesServiceService_IUnregisterMachine extends grpc.MethodDefinition<proto_service_pb.UnregisterMachineRequest, proto_service_pb.UnregisterMachineResponse> {
    path: "/api.MachinesService/UnregisterMachine";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.UnregisterMachineRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.UnregisterMachineRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.UnregisterMachineResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.UnregisterMachineResponse>;
}
interface IMachinesServiceService_IDrainMachine extends grpc.MethodDefinition<proto_service_pb.DrainMachineRequest, proto_service_pb.DrainMachineResponse> {
    path: "/api.MachinesService/DrainMachine";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.DrainMachineRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.DrainMachineRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.DrainMachineResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.DrainMachineResponse>;
}

export const MachinesServiceService: IMachinesServiceService;

export interface IMachinesServiceServer extends grpc.UntypedServiceImplementation {
    registerMachine: grpc.handleUnaryCall<proto_service_pb.RegisterMachineRequest, proto_service_pb.Machine>;
    listMachines: grpc.handleUnaryCall<proto_service_pb.ListMachinesRequest, proto_service_pb.ListMachinesResponse>;
    getMachine: grpc.handleUnaryCall<proto_service_pb.GetMachineRequest, proto_service_pb.Machine>;
    updateMachinePing: grpc.handleUnaryCall<proto_service_pb.UpdateMachinePingRequest, proto_service_pb.Machine>;
    unregisterMachine: grpc.handleUnaryCall<proto_service_pb.UnregisterMachineRequest, proto_service_pb.UnregisterMachineResponse>;
    drainMachine: grpc.handleUnaryCall<proto_service_pb.DrainMachineRequest, proto_service_pb.DrainMachineResponse>;
}

export interface IMachinesServiceClient {
    registerMachine(request: proto_service_pb.RegisterMachineRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Machine) => void): grpc.ClientUnaryCall;
    registerMachine(request: proto_service_pb.RegisterMachineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Machine) => void): grpc.ClientUnaryCall;
    registerMachine(request: proto_service_pb.RegisterMachineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Machine) => void): grpc.ClientUnaryCall;
    listMachines(request: proto_service_pb.ListMachinesRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListMachinesResponse) => void): grpc.ClientUnaryCall;
    listMachines(request: proto_service_pb.ListMachinesRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListMachinesResponse) => void): grpc.ClientUnaryCall;
    listMachines(request: proto_service_pb.ListMachinesRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListMachinesResponse) => void): grpc.ClientUnaryCall;
    getMachine(request: proto_service_pb.GetMachineRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Machine) => void): grpc.ClientUnaryCall;
    getMachine(request: proto_service_pb.GetMachineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Machine) => void): grpc.ClientUnaryCall;
    getMachine(request: proto_service_pb.GetMachineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Machine) => void): grpc.ClientUnaryCall;
    updateMachinePing(request: proto_service_pb.UpdateMachinePingRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Machine) => void): grpc.ClientUnaryCall;
    updateMachinePing(request: proto_service_pb.UpdateMachinePingRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Machine) => void): grpc.ClientUnaryCall;
    updateMachinePing(request: proto_service_pb.UpdateMachinePingRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Machine) => void): grpc.ClientUnaryCall;
    unregisterMachine(request: proto_service_pb.UnregisterMachineRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.UnregisterMachineResponse) => void): grpc.ClientUnaryCall;
    unregisterMachine(request: proto_service_pb.UnregisterMachineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.UnregisterMachineResponse) => void): grpc.ClientUnaryCall;
    unregisterMachine(request: proto_service_pb.UnregisterMachineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.UnregisterMachineResponse) => void): grpc.ClientUnaryCall;
    drainMachine(request: proto_service_pb.DrainMachineRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.DrainMachineResponse) => void): grpc.ClientUnaryCall;
    drainMachine(request: proto_service_pb.DrainMachineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.DrainMachineResponse) => void): grpc.ClientUnaryCall;
    drainMachine(request: proto_service_pb.DrainMachineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.DrainMachineResponse) => void): grpc.ClientUnaryCall;
}

export class MachinesServiceClient extends grpc.Client implements IMachinesServiceClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public registerMachine(request: proto_service_pb.RegisterMachineRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Machine) => void): grpc.ClientUnaryCall;
    public registerMachine(request: proto_service_pb.RegisterMachineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Machine) => void): grpc.ClientUnaryCall;
    public registerMachine(request: proto_service_pb.RegisterMachineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Machine) => void): grpc.ClientUnaryCall;
    public listMachines(request: proto_service_pb.ListMachinesRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListMachinesResponse) => void): grpc.ClientUnaryCall;
    public listMachines(request: proto_service_pb.ListMachinesRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListMachinesResponse) => void): grpc.ClientUnaryCall;
    public listMachines(request: proto_service_pb.ListMachinesRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListMachinesResponse) => void): grpc.ClientUnaryCall;
    public getMachine(request: proto_service_pb.GetMachineRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Machine) => void): grpc.ClientUnaryCall;
    public getMachine(request: proto_service_pb.GetMachineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Machine) => void): grpc.ClientUnaryCall;
    public getMachine(request: proto_service_pb.GetMachineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Machine) => void): grpc.ClientUnaryCall;
    public updateMachinePing(request: proto_service_pb.UpdateMachinePingRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Machine) => void): grpc.ClientUnaryCall;
    public updateMachinePing(request: proto_service_pb.UpdateMachinePingRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Machine) => void): grpc.ClientUnaryCall;
    public updateMachinePing(request: proto_service_pb.UpdateMachinePingRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Machine) => void): grpc.ClientUnaryCall;
    public unregisterMachine(request: proto_service_pb.UnregisterMachineRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.UnregisterMachineResponse) => void): grpc.ClientUnaryCall;
    public unregisterMachine(request: proto_service_pb.UnregisterMachineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.UnregisterMachineResponse) => void): grpc.ClientUnaryCall;
    public unregisterMachine(request: proto_service_pb.UnregisterMachineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.UnregisterMachineResponse) => void): grpc.ClientUnaryCall;
    public drainMachine(request: proto_service_pb.DrainMachineRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.DrainMachineResponse) => void): grpc.ClientUnaryCall;
    public drainMachine(request: proto_service_pb.DrainMachineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.DrainMachineResponse) => void): grpc.ClientUnaryCall;
    public drainMachine(request: proto_service_pb.DrainMachineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.DrainMachineResponse) => void): grpc.ClientUnaryCall;
}

interface IRequestsServiceService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    createRequest: IRequestsServiceService_ICreateRequest;
    getRequest: IRequestsServiceService_IGetRequest;
    listRequests: IRequestsServiceService_IListRequests;
    updateRequest: IRequestsServiceService_IUpdateRequest;
    claimRequest: IRequestsServiceService_IClaimRequest;
    cancelRequest: IRequestsServiceService_ICancelRequest;
    submitRequestResult: IRequestsServiceService_ISubmitRequestResult;
    appendRequestChunks: IRequestsServiceService_IAppendRequestChunks;
    getRequestChunks: IRequestsServiceService_IGetRequestChunks;
}

interface IRequestsServiceService_ICreateRequest extends grpc.MethodDefinition<proto_service_pb.CreateRequestRequest, proto_service_pb.Request> {
    path: "/api.RequestsService/CreateRequest";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.CreateRequestRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.CreateRequestRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.Request>;
    responseDeserialize: grpc.deserialize<proto_service_pb.Request>;
}
interface IRequestsServiceService_IGetRequest extends grpc.MethodDefinition<proto_service_pb.GetRequestRequest, proto_service_pb.Request> {
    path: "/api.RequestsService/GetRequest";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.GetRequestRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.GetRequestRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.Request>;
    responseDeserialize: grpc.deserialize<proto_service_pb.Request>;
}
interface IRequestsServiceService_IListRequests extends grpc.MethodDefinition<proto_service_pb.ListRequestsRequest, proto_service_pb.ListRequestsResponse> {
    path: "/api.RequestsService/ListRequests";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.ListRequestsRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.ListRequestsRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.ListRequestsResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.ListRequestsResponse>;
}
interface IRequestsServiceService_IUpdateRequest extends grpc.MethodDefinition<proto_service_pb.UpdateRequestRequest, proto_service_pb.Request> {
    path: "/api.RequestsService/UpdateRequest";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.UpdateRequestRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.UpdateRequestRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.Request>;
    responseDeserialize: grpc.deserialize<proto_service_pb.Request>;
}
interface IRequestsServiceService_IClaimRequest extends grpc.MethodDefinition<proto_service_pb.ClaimRequestRequest, proto_service_pb.Request> {
    path: "/api.RequestsService/ClaimRequest";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.ClaimRequestRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.ClaimRequestRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.Request>;
    responseDeserialize: grpc.deserialize<proto_service_pb.Request>;
}
interface IRequestsServiceService_ICancelRequest extends grpc.MethodDefinition<proto_service_pb.CancelRequestRequest, proto_service_pb.CancelRequestResponse> {
    path: "/api.RequestsService/CancelRequest";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.CancelRequestRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.CancelRequestRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.CancelRequestResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.CancelRequestResponse>;
}
interface IRequestsServiceService_ISubmitRequestResult extends grpc.MethodDefinition<proto_service_pb.SubmitRequestResultRequest, proto_service_pb.SubmitRequestResultResponse> {
    path: "/api.RequestsService/SubmitRequestResult";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.SubmitRequestResultRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.SubmitRequestResultRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.SubmitRequestResultResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.SubmitRequestResultResponse>;
}
interface IRequestsServiceService_IAppendRequestChunks extends grpc.MethodDefinition<proto_service_pb.AppendRequestChunksRequest, proto_service_pb.AppendRequestChunksResponse> {
    path: "/api.RequestsService/AppendRequestChunks";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.AppendRequestChunksRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.AppendRequestChunksRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.AppendRequestChunksResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.AppendRequestChunksResponse>;
}
interface IRequestsServiceService_IGetRequestChunks extends grpc.MethodDefinition<proto_service_pb.GetRequestChunksRequest, proto_service_pb.GetRequestChunksResponse> {
    path: "/api.RequestsService/GetRequestChunks";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.GetRequestChunksRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.GetRequestChunksRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.GetRequestChunksResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.GetRequestChunksResponse>;
}

export const RequestsServiceService: IRequestsServiceService;

export interface IRequestsServiceServer extends grpc.UntypedServiceImplementation {
    createRequest: grpc.handleUnaryCall<proto_service_pb.CreateRequestRequest, proto_service_pb.Request>;
    getRequest: grpc.handleUnaryCall<proto_service_pb.GetRequestRequest, proto_service_pb.Request>;
    listRequests: grpc.handleUnaryCall<proto_service_pb.ListRequestsRequest, proto_service_pb.ListRequestsResponse>;
    updateRequest: grpc.handleUnaryCall<proto_service_pb.UpdateRequestRequest, proto_service_pb.Request>;
    claimRequest: grpc.handleUnaryCall<proto_service_pb.ClaimRequestRequest, proto_service_pb.Request>;
    cancelRequest: grpc.handleUnaryCall<proto_service_pb.CancelRequestRequest, proto_service_pb.CancelRequestResponse>;
    submitRequestResult: grpc.handleUnaryCall<proto_service_pb.SubmitRequestResultRequest, proto_service_pb.SubmitRequestResultResponse>;
    appendRequestChunks: grpc.handleUnaryCall<proto_service_pb.AppendRequestChunksRequest, proto_service_pb.AppendRequestChunksResponse>;
    getRequestChunks: grpc.handleUnaryCall<proto_service_pb.GetRequestChunksRequest, proto_service_pb.GetRequestChunksResponse>;
}

export interface IRequestsServiceClient {
    createRequest(request: proto_service_pb.CreateRequestRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    createRequest(request: proto_service_pb.CreateRequestRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    createRequest(request: proto_service_pb.CreateRequestRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    getRequest(request: proto_service_pb.GetRequestRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    getRequest(request: proto_service_pb.GetRequestRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    getRequest(request: proto_service_pb.GetRequestRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    listRequests(request: proto_service_pb.ListRequestsRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListRequestsResponse) => void): grpc.ClientUnaryCall;
    listRequests(request: proto_service_pb.ListRequestsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListRequestsResponse) => void): grpc.ClientUnaryCall;
    listRequests(request: proto_service_pb.ListRequestsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListRequestsResponse) => void): grpc.ClientUnaryCall;
    updateRequest(request: proto_service_pb.UpdateRequestRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    updateRequest(request: proto_service_pb.UpdateRequestRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    updateRequest(request: proto_service_pb.UpdateRequestRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    claimRequest(request: proto_service_pb.ClaimRequestRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    claimRequest(request: proto_service_pb.ClaimRequestRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    claimRequest(request: proto_service_pb.ClaimRequestRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    cancelRequest(request: proto_service_pb.CancelRequestRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.CancelRequestResponse) => void): grpc.ClientUnaryCall;
    cancelRequest(request: proto_service_pb.CancelRequestRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.CancelRequestResponse) => void): grpc.ClientUnaryCall;
    cancelRequest(request: proto_service_pb.CancelRequestRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.CancelRequestResponse) => void): grpc.ClientUnaryCall;
    submitRequestResult(request: proto_service_pb.SubmitRequestResultRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.SubmitRequestResultResponse) => void): grpc.ClientUnaryCall;
    submitRequestResult(request: proto_service_pb.SubmitRequestResultRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.SubmitRequestResultResponse) => void): grpc.ClientUnaryCall;
    submitRequestResult(request: proto_service_pb.SubmitRequestResultRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.SubmitRequestResultResponse) => void): grpc.ClientUnaryCall;
    appendRequestChunks(request: proto_service_pb.AppendRequestChunksRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.AppendRequestChunksResponse) => void): grpc.ClientUnaryCall;
    appendRequestChunks(request: proto_service_pb.AppendRequestChunksRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.AppendRequestChunksResponse) => void): grpc.ClientUnaryCall;
    appendRequestChunks(request: proto_service_pb.AppendRequestChunksRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.AppendRequestChunksResponse) => void): grpc.ClientUnaryCall;
    getRequestChunks(request: proto_service_pb.GetRequestChunksRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetRequestChunksResponse) => void): grpc.ClientUnaryCall;
    getRequestChunks(request: proto_service_pb.GetRequestChunksRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetRequestChunksResponse) => void): grpc.ClientUnaryCall;
    getRequestChunks(request: proto_service_pb.GetRequestChunksRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetRequestChunksResponse) => void): grpc.ClientUnaryCall;
}

export class RequestsServiceClient extends grpc.Client implements IRequestsServiceClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public createRequest(request: proto_service_pb.CreateRequestRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    public createRequest(request: proto_service_pb.CreateRequestRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    public createRequest(request: proto_service_pb.CreateRequestRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    public getRequest(request: proto_service_pb.GetRequestRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    public getRequest(request: proto_service_pb.GetRequestRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    public getRequest(request: proto_service_pb.GetRequestRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    public listRequests(request: proto_service_pb.ListRequestsRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListRequestsResponse) => void): grpc.ClientUnaryCall;
    public listRequests(request: proto_service_pb.ListRequestsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListRequestsResponse) => void): grpc.ClientUnaryCall;
    public listRequests(request: proto_service_pb.ListRequestsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListRequestsResponse) => void): grpc.ClientUnaryCall;
    public updateRequest(request: proto_service_pb.UpdateRequestRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    public updateRequest(request: proto_service_pb.UpdateRequestRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    public updateRequest(request: proto_service_pb.UpdateRequestRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    public claimRequest(request: proto_service_pb.ClaimRequestRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    public claimRequest(request: proto_service_pb.ClaimRequestRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    public claimRequest(request: proto_service_pb.ClaimRequestRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Request) => void): grpc.ClientUnaryCall;
    public cancelRequest(request: proto_service_pb.CancelRequestRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.CancelRequestResponse) => void): grpc.ClientUnaryCall;
    public cancelRequest(request: proto_service_pb.CancelRequestRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.CancelRequestResponse) => void): grpc.ClientUnaryCall;
    public cancelRequest(request: proto_service_pb.CancelRequestRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.CancelRequestResponse) => void): grpc.ClientUnaryCall;
    public submitRequestResult(request: proto_service_pb.SubmitRequestResultRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.SubmitRequestResultResponse) => void): grpc.ClientUnaryCall;
    public submitRequestResult(request: proto_service_pb.SubmitRequestResultRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.SubmitRequestResultResponse) => void): grpc.ClientUnaryCall;
    public submitRequestResult(request: proto_service_pb.SubmitRequestResultRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.SubmitRequestResultResponse) => void): grpc.ClientUnaryCall;
    public appendRequestChunks(request: proto_service_pb.AppendRequestChunksRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.AppendRequestChunksResponse) => void): grpc.ClientUnaryCall;
    public appendRequestChunks(request: proto_service_pb.AppendRequestChunksRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.AppendRequestChunksResponse) => void): grpc.ClientUnaryCall;
    public appendRequestChunks(request: proto_service_pb.AppendRequestChunksRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.AppendRequestChunksResponse) => void): grpc.ClientUnaryCall;
    public getRequestChunks(request: proto_service_pb.GetRequestChunksRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetRequestChunksResponse) => void): grpc.ClientUnaryCall;
    public getRequestChunks(request: proto_service_pb.GetRequestChunksRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetRequestChunksResponse) => void): grpc.ClientUnaryCall;
    public getRequestChunks(request: proto_service_pb.GetRequestChunksRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.GetRequestChunksResponse) => void): grpc.ClientUnaryCall;
}

interface ITasksServiceService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    createTask: ITasksServiceService_ICreateTask;
    getTask: ITasksServiceService_IGetTask;
    listTasks: ITasksServiceService_IListTasks;
    cancelTask: ITasksServiceService_ICancelTask;
}

interface ITasksServiceService_ICreateTask extends grpc.MethodDefinition<proto_service_pb.CreateTaskRequest, proto_service_pb.Task> {
    path: "/api.TasksService/CreateTask";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.CreateTaskRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.CreateTaskRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.Task>;
    responseDeserialize: grpc.deserialize<proto_service_pb.Task>;
}
interface ITasksServiceService_IGetTask extends grpc.MethodDefinition<proto_service_pb.GetTaskRequest, proto_service_pb.Task> {
    path: "/api.TasksService/GetTask";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.GetTaskRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.GetTaskRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.Task>;
    responseDeserialize: grpc.deserialize<proto_service_pb.Task>;
}
interface ITasksServiceService_IListTasks extends grpc.MethodDefinition<proto_service_pb.ListTasksRequest, proto_service_pb.ListTasksResponse> {
    path: "/api.TasksService/ListTasks";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.ListTasksRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.ListTasksRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.ListTasksResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.ListTasksResponse>;
}
interface ITasksServiceService_ICancelTask extends grpc.MethodDefinition<proto_service_pb.CancelTaskRequest, proto_service_pb.CancelTaskResponse> {
    path: "/api.TasksService/CancelTask";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<proto_service_pb.CancelTaskRequest>;
    requestDeserialize: grpc.deserialize<proto_service_pb.CancelTaskRequest>;
    responseSerialize: grpc.serialize<proto_service_pb.CancelTaskResponse>;
    responseDeserialize: grpc.deserialize<proto_service_pb.CancelTaskResponse>;
}

export const TasksServiceService: ITasksServiceService;

export interface ITasksServiceServer extends grpc.UntypedServiceImplementation {
    createTask: grpc.handleUnaryCall<proto_service_pb.CreateTaskRequest, proto_service_pb.Task>;
    getTask: grpc.handleUnaryCall<proto_service_pb.GetTaskRequest, proto_service_pb.Task>;
    listTasks: grpc.handleUnaryCall<proto_service_pb.ListTasksRequest, proto_service_pb.ListTasksResponse>;
    cancelTask: grpc.handleUnaryCall<proto_service_pb.CancelTaskRequest, proto_service_pb.CancelTaskResponse>;
}

export interface ITasksServiceClient {
    createTask(request: proto_service_pb.CreateTaskRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Task) => void): grpc.ClientUnaryCall;
    createTask(request: proto_service_pb.CreateTaskRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Task) => void): grpc.ClientUnaryCall;
    createTask(request: proto_service_pb.CreateTaskRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Task) => void): grpc.ClientUnaryCall;
    getTask(request: proto_service_pb.GetTaskRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Task) => void): grpc.ClientUnaryCall;
    getTask(request: proto_service_pb.GetTaskRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Task) => void): grpc.ClientUnaryCall;
    getTask(request: proto_service_pb.GetTaskRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Task) => void): grpc.ClientUnaryCall;
    listTasks(request: proto_service_pb.ListTasksRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListTasksResponse) => void): grpc.ClientUnaryCall;
    listTasks(request: proto_service_pb.ListTasksRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListTasksResponse) => void): grpc.ClientUnaryCall;
    listTasks(request: proto_service_pb.ListTasksRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListTasksResponse) => void): grpc.ClientUnaryCall;
    cancelTask(request: proto_service_pb.CancelTaskRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.CancelTaskResponse) => void): grpc.ClientUnaryCall;
    cancelTask(request: proto_service_pb.CancelTaskRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.CancelTaskResponse) => void): grpc.ClientUnaryCall;
    cancelTask(request: proto_service_pb.CancelTaskRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.CancelTaskResponse) => void): grpc.ClientUnaryCall;
}

export class TasksServiceClient extends grpc.Client implements ITasksServiceClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public createTask(request: proto_service_pb.CreateTaskRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Task) => void): grpc.ClientUnaryCall;
    public createTask(request: proto_service_pb.CreateTaskRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Task) => void): grpc.ClientUnaryCall;
    public createTask(request: proto_service_pb.CreateTaskRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Task) => void): grpc.ClientUnaryCall;
    public getTask(request: proto_service_pb.GetTaskRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Task) => void): grpc.ClientUnaryCall;
    public getTask(request: proto_service_pb.GetTaskRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Task) => void): grpc.ClientUnaryCall;
    public getTask(request: proto_service_pb.GetTaskRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.Task) => void): grpc.ClientUnaryCall;
    public listTasks(request: proto_service_pb.ListTasksRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListTasksResponse) => void): grpc.ClientUnaryCall;
    public listTasks(request: proto_service_pb.ListTasksRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListTasksResponse) => void): grpc.ClientUnaryCall;
    public listTasks(request: proto_service_pb.ListTasksRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.ListTasksResponse) => void): grpc.ClientUnaryCall;
    public cancelTask(request: proto_service_pb.CancelTaskRequest, callback: (error: grpc.ServiceError | null, response: proto_service_pb.CancelTaskResponse) => void): grpc.ClientUnaryCall;
    public cancelTask(request: proto_service_pb.CancelTaskRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: proto_service_pb.CancelTaskResponse) => void): grpc.ClientUnaryCall;
    public cancelTask(request: proto_service_pb.CancelTaskRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: proto_service_pb.CancelTaskResponse) => void): grpc.ClientUnaryCall;
}
