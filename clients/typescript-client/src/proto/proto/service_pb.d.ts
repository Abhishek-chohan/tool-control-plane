// package: api
// file: proto/service.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";

export class ResumeStreamRequest extends jspb.Message { 
    getRequestId(): string;
    setRequestId(value: string): ResumeStreamRequest;
    getLastSeq(): number;
    setLastSeq(value: number): ResumeStreamRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ResumeStreamRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ResumeStreamRequest): ResumeStreamRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ResumeStreamRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ResumeStreamRequest;
    static deserializeBinaryFromReader(message: ResumeStreamRequest, reader: jspb.BinaryReader): ResumeStreamRequest;
}

export namespace ResumeStreamRequest {
    export type AsObject = {
        requestId: string,
        lastSeq: number,
    }
}

export class DrainMachineRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): DrainMachineRequest;
    getMachineId(): string;
    setMachineId(value: string): DrainMachineRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DrainMachineRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DrainMachineRequest): DrainMachineRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DrainMachineRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DrainMachineRequest;
    static deserializeBinaryFromReader(message: DrainMachineRequest, reader: jspb.BinaryReader): DrainMachineRequest;
}

export namespace DrainMachineRequest {
    export type AsObject = {
        sessionId: string,
        machineId: string,
    }
}

export class DrainMachineResponse extends jspb.Message { 
    getDrained(): boolean;
    setDrained(value: boolean): DrainMachineResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DrainMachineResponse.AsObject;
    static toObject(includeInstance: boolean, msg: DrainMachineResponse): DrainMachineResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DrainMachineResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DrainMachineResponse;
    static deserializeBinaryFromReader(message: DrainMachineResponse, reader: jspb.BinaryReader): DrainMachineResponse;
}

export namespace DrainMachineResponse {
    export type AsObject = {
        drained: boolean,
    }
}

export class Tool extends jspb.Message { 
    getId(): string;
    setId(value: string): Tool;
    getName(): string;
    setName(value: string): Tool;
    getDescription(): string;
    setDescription(value: string): Tool;
    getSchema(): string;
    setSchema(value: string): Tool;

    getConfigMap(): jspb.Map<string, string>;
    clearConfigMap(): void;
    getCreatedAt(): string;
    setCreatedAt(value: string): Tool;
    getLastPingAt(): string;
    setLastPingAt(value: string): Tool;
    getSessionId(): string;
    setSessionId(value: string): Tool;
    getMachineId(): string;
    setMachineId(value: string): Tool;
    clearTagsList(): void;
    getTagsList(): Array<string>;
    setTagsList(value: Array<string>): Tool;
    addTags(value: string, index?: number): string;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Tool.AsObject;
    static toObject(includeInstance: boolean, msg: Tool): Tool.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Tool, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Tool;
    static deserializeBinaryFromReader(message: Tool, reader: jspb.BinaryReader): Tool;
}

export namespace Tool {
    export type AsObject = {
        id: string,
        name: string,
        description: string,
        schema: string,

        configMap: Array<[string, string]>,
        createdAt: string,
        lastPingAt: string,
        sessionId: string,
        machineId: string,
        tagsList: Array<string>,
    }
}

export class Session extends jspb.Message { 
    getId(): string;
    setId(value: string): Session;
    getName(): string;
    setName(value: string): Session;
    getDescription(): string;
    setDescription(value: string): Session;
    getCreatedAt(): string;
    setCreatedAt(value: string): Session;
    getCreatedBy(): string;
    setCreatedBy(value: string): Session;
    getApiKey(): string;
    setApiKey(value: string): Session;
    getNamespace(): string;
    setNamespace(value: string): Session;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Session.AsObject;
    static toObject(includeInstance: boolean, msg: Session): Session.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Session, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Session;
    static deserializeBinaryFromReader(message: Session, reader: jspb.BinaryReader): Session;
}

export namespace Session {
    export type AsObject = {
        id: string,
        name: string,
        description: string,
        createdAt: string,
        createdBy: string,
        apiKey: string,
        namespace: string,
    }
}

export class ApiKey extends jspb.Message { 
    getId(): string;
    setId(value: string): ApiKey;
    getName(): string;
    setName(value: string): ApiKey;
    getKey(): string;
    setKey(value: string): ApiKey;
    getSessionId(): string;
    setSessionId(value: string): ApiKey;
    getCreatedAt(): string;
    setCreatedAt(value: string): ApiKey;
    getCreatedBy(): string;
    setCreatedBy(value: string): ApiKey;
    getRevokedAt(): string;
    setRevokedAt(value: string): ApiKey;
    clearCapabilitiesList(): void;
    getCapabilitiesList(): Array<string>;
    setCapabilitiesList(value: Array<string>): ApiKey;
    addCapabilities(value: string, index?: number): string;
    getKeyPreview(): string;
    setKeyPreview(value: string): ApiKey;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ApiKey.AsObject;
    static toObject(includeInstance: boolean, msg: ApiKey): ApiKey.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ApiKey, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ApiKey;
    static deserializeBinaryFromReader(message: ApiKey, reader: jspb.BinaryReader): ApiKey;
}

export namespace ApiKey {
    export type AsObject = {
        id: string,
        name: string,
        key: string,
        sessionId: string,
        createdAt: string,
        createdBy: string,
        revokedAt: string,
        capabilitiesList: Array<string>,
        keyPreview: string,
    }
}

export class Machine extends jspb.Message { 
    getId(): string;
    setId(value: string): Machine;
    getSessionId(): string;
    setSessionId(value: string): Machine;
    getSdkVersion(): string;
    setSdkVersion(value: string): Machine;
    getSdkLanguage(): string;
    setSdkLanguage(value: string): Machine;
    getIp(): string;
    setIp(value: string): Machine;
    getCreatedAt(): string;
    setCreatedAt(value: string): Machine;
    getLastPingAt(): string;
    setLastPingAt(value: string): Machine;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Machine.AsObject;
    static toObject(includeInstance: boolean, msg: Machine): Machine.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Machine, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Machine;
    static deserializeBinaryFromReader(message: Machine, reader: jspb.BinaryReader): Machine;
}

export namespace Machine {
    export type AsObject = {
        id: string,
        sessionId: string,
        sdkVersion: string,
        sdkLanguage: string,
        ip: string,
        createdAt: string,
        lastPingAt: string,
    }
}

export class Request extends jspb.Message { 
    getId(): string;
    setId(value: string): Request;
    getSessionId(): string;
    setSessionId(value: string): Request;
    getToolName(): string;
    setToolName(value: string): Request;
    getStatus(): string;
    setStatus(value: string): Request;
    getInput(): string;
    setInput(value: string): Request;
    getResult(): string;
    setResult(value: string): Request;
    getResultType(): string;
    setResultType(value: string): Request;
    getError(): string;
    setError(value: string): Request;
    getCreatedAt(): string;
    setCreatedAt(value: string): Request;
    getUpdatedAt(): string;
    setUpdatedAt(value: string): Request;
    getExecutingMachineId(): string;
    setExecutingMachineId(value: string): Request;
    clearStreamResultsList(): void;
    getStreamResultsList(): Array<string>;
    setStreamResultsList(value: Array<string>): Request;
    addStreamResults(value: string, index?: number): string;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Request.AsObject;
    static toObject(includeInstance: boolean, msg: Request): Request.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Request, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Request;
    static deserializeBinaryFromReader(message: Request, reader: jspb.BinaryReader): Request;
}

export namespace Request {
    export type AsObject = {
        id: string,
        sessionId: string,
        toolName: string,
        status: string,
        input: string,
        result: string,
        resultType: string,
        error: string,
        createdAt: string,
        updatedAt: string,
        executingMachineId: string,
        streamResultsList: Array<string>,
    }
}

export class RegisterToolRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): RegisterToolRequest;
    getMachineId(): string;
    setMachineId(value: string): RegisterToolRequest;
    getName(): string;
    setName(value: string): RegisterToolRequest;
    getDescription(): string;
    setDescription(value: string): RegisterToolRequest;
    getSchema(): string;
    setSchema(value: string): RegisterToolRequest;

    getConfigMap(): jspb.Map<string, string>;
    clearConfigMap(): void;
    clearTagsList(): void;
    getTagsList(): Array<string>;
    setTagsList(value: Array<string>): RegisterToolRequest;
    addTags(value: string, index?: number): string;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RegisterToolRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RegisterToolRequest): RegisterToolRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RegisterToolRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RegisterToolRequest;
    static deserializeBinaryFromReader(message: RegisterToolRequest, reader: jspb.BinaryReader): RegisterToolRequest;
}

export namespace RegisterToolRequest {
    export type AsObject = {
        sessionId: string,
        machineId: string,
        name: string,
        description: string,
        schema: string,

        configMap: Array<[string, string]>,
        tagsList: Array<string>,
    }
}

export class RegisterToolResponse extends jspb.Message { 

    hasTool(): boolean;
    clearTool(): void;
    getTool(): Tool | undefined;
    setTool(value?: Tool): RegisterToolResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RegisterToolResponse.AsObject;
    static toObject(includeInstance: boolean, msg: RegisterToolResponse): RegisterToolResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RegisterToolResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RegisterToolResponse;
    static deserializeBinaryFromReader(message: RegisterToolResponse, reader: jspb.BinaryReader): RegisterToolResponse;
}

export namespace RegisterToolResponse {
    export type AsObject = {
        tool?: Tool.AsObject,
    }
}

export class ListToolsRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): ListToolsRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListToolsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListToolsRequest): ListToolsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListToolsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListToolsRequest;
    static deserializeBinaryFromReader(message: ListToolsRequest, reader: jspb.BinaryReader): ListToolsRequest;
}

export namespace ListToolsRequest {
    export type AsObject = {
        sessionId: string,
    }
}

export class ListToolsResponse extends jspb.Message { 
    clearToolsList(): void;
    getToolsList(): Array<Tool>;
    setToolsList(value: Array<Tool>): ListToolsResponse;
    addTools(value?: Tool, index?: number): Tool;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListToolsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ListToolsResponse): ListToolsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListToolsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListToolsResponse;
    static deserializeBinaryFromReader(message: ListToolsResponse, reader: jspb.BinaryReader): ListToolsResponse;
}

export namespace ListToolsResponse {
    export type AsObject = {
        toolsList: Array<Tool.AsObject>,
    }
}

export class GetToolByIdRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): GetToolByIdRequest;
    getToolId(): string;
    setToolId(value: string): GetToolByIdRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetToolByIdRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetToolByIdRequest): GetToolByIdRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetToolByIdRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetToolByIdRequest;
    static deserializeBinaryFromReader(message: GetToolByIdRequest, reader: jspb.BinaryReader): GetToolByIdRequest;
}

export namespace GetToolByIdRequest {
    export type AsObject = {
        sessionId: string,
        toolId: string,
    }
}

export class GetToolByNameRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): GetToolByNameRequest;
    getToolName(): string;
    setToolName(value: string): GetToolByNameRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetToolByNameRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetToolByNameRequest): GetToolByNameRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetToolByNameRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetToolByNameRequest;
    static deserializeBinaryFromReader(message: GetToolByNameRequest, reader: jspb.BinaryReader): GetToolByNameRequest;
}

export namespace GetToolByNameRequest {
    export type AsObject = {
        sessionId: string,
        toolName: string,
    }
}

export class GetToolResponse extends jspb.Message { 

    hasTool(): boolean;
    clearTool(): void;
    getTool(): Tool | undefined;
    setTool(value?: Tool): GetToolResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetToolResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetToolResponse): GetToolResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetToolResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetToolResponse;
    static deserializeBinaryFromReader(message: GetToolResponse, reader: jspb.BinaryReader): GetToolResponse;
}

export namespace GetToolResponse {
    export type AsObject = {
        tool?: Tool.AsObject,
    }
}

export class DeleteToolRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): DeleteToolRequest;
    getToolId(): string;
    setToolId(value: string): DeleteToolRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteToolRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteToolRequest): DeleteToolRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteToolRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteToolRequest;
    static deserializeBinaryFromReader(message: DeleteToolRequest, reader: jspb.BinaryReader): DeleteToolRequest;
}

export namespace DeleteToolRequest {
    export type AsObject = {
        sessionId: string,
        toolId: string,
    }
}

export class DeleteToolResponse extends jspb.Message { 
    getSuccess(): boolean;
    setSuccess(value: boolean): DeleteToolResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteToolResponse.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteToolResponse): DeleteToolResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteToolResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteToolResponse;
    static deserializeBinaryFromReader(message: DeleteToolResponse, reader: jspb.BinaryReader): DeleteToolResponse;
}

export namespace DeleteToolResponse {
    export type AsObject = {
        success: boolean,
    }
}

export class UpdateToolPingRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): UpdateToolPingRequest;
    getToolId(): string;
    setToolId(value: string): UpdateToolPingRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): UpdateToolPingRequest.AsObject;
    static toObject(includeInstance: boolean, msg: UpdateToolPingRequest): UpdateToolPingRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: UpdateToolPingRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): UpdateToolPingRequest;
    static deserializeBinaryFromReader(message: UpdateToolPingRequest, reader: jspb.BinaryReader): UpdateToolPingRequest;
}

export namespace UpdateToolPingRequest {
    export type AsObject = {
        sessionId: string,
        toolId: string,
    }
}

export class CreateSessionRequest extends jspb.Message { 
    getUserId(): string;
    setUserId(value: string): CreateSessionRequest;
    getName(): string;
    setName(value: string): CreateSessionRequest;
    getDescription(): string;
    setDescription(value: string): CreateSessionRequest;
    getApiKey(): string;
    setApiKey(value: string): CreateSessionRequest;
    getSessionId(): string;
    setSessionId(value: string): CreateSessionRequest;
    getNamespace(): string;
    setNamespace(value: string): CreateSessionRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreateSessionRequest.AsObject;
    static toObject(includeInstance: boolean, msg: CreateSessionRequest): CreateSessionRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreateSessionRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreateSessionRequest;
    static deserializeBinaryFromReader(message: CreateSessionRequest, reader: jspb.BinaryReader): CreateSessionRequest;
}

export namespace CreateSessionRequest {
    export type AsObject = {
        userId: string,
        name: string,
        description: string,
        apiKey: string,
        sessionId: string,
        namespace: string,
    }
}

export class CreateSessionResponse extends jspb.Message { 

    hasSession(): boolean;
    clearSession(): void;
    getSession(): Session | undefined;
    setSession(value?: Session): CreateSessionResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreateSessionResponse.AsObject;
    static toObject(includeInstance: boolean, msg: CreateSessionResponse): CreateSessionResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreateSessionResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreateSessionResponse;
    static deserializeBinaryFromReader(message: CreateSessionResponse, reader: jspb.BinaryReader): CreateSessionResponse;
}

export namespace CreateSessionResponse {
    export type AsObject = {
        session?: Session.AsObject,
    }
}

export class GetSessionRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): GetSessionRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetSessionRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetSessionRequest): GetSessionRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetSessionRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetSessionRequest;
    static deserializeBinaryFromReader(message: GetSessionRequest, reader: jspb.BinaryReader): GetSessionRequest;
}

export namespace GetSessionRequest {
    export type AsObject = {
        sessionId: string,
    }
}

export class ListSessionsRequest extends jspb.Message { 
    getUserId(): string;
    setUserId(value: string): ListSessionsRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListSessionsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListSessionsRequest): ListSessionsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListSessionsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListSessionsRequest;
    static deserializeBinaryFromReader(message: ListSessionsRequest, reader: jspb.BinaryReader): ListSessionsRequest;
}

export namespace ListSessionsRequest {
    export type AsObject = {
        userId: string,
    }
}

export class ListSessionsResponse extends jspb.Message { 
    clearSessionsList(): void;
    getSessionsList(): Array<Session>;
    setSessionsList(value: Array<Session>): ListSessionsResponse;
    addSessions(value?: Session, index?: number): Session;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListSessionsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ListSessionsResponse): ListSessionsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListSessionsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListSessionsResponse;
    static deserializeBinaryFromReader(message: ListSessionsResponse, reader: jspb.BinaryReader): ListSessionsResponse;
}

export namespace ListSessionsResponse {
    export type AsObject = {
        sessionsList: Array<Session.AsObject>,
    }
}

export class ListUserSessionsRequest extends jspb.Message { 
    getUserId(): string;
    setUserId(value: string): ListUserSessionsRequest;
    getPageSize(): number;
    setPageSize(value: number): ListUserSessionsRequest;
    getPageToken(): number;
    setPageToken(value: number): ListUserSessionsRequest;
    getFilter(): string;
    setFilter(value: string): ListUserSessionsRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListUserSessionsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListUserSessionsRequest): ListUserSessionsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListUserSessionsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListUserSessionsRequest;
    static deserializeBinaryFromReader(message: ListUserSessionsRequest, reader: jspb.BinaryReader): ListUserSessionsRequest;
}

export namespace ListUserSessionsRequest {
    export type AsObject = {
        userId: string,
        pageSize: number,
        pageToken: number,
        filter: string,
    }
}

export class ListUserSessionsResponse extends jspb.Message { 
    clearSessionsList(): void;
    getSessionsList(): Array<Session>;
    setSessionsList(value: Array<Session>): ListUserSessionsResponse;
    addSessions(value?: Session, index?: number): Session;
    getNextPageToken(): number;
    setNextPageToken(value: number): ListUserSessionsResponse;
    getTotalCount(): number;
    setTotalCount(value: number): ListUserSessionsResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListUserSessionsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ListUserSessionsResponse): ListUserSessionsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListUserSessionsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListUserSessionsResponse;
    static deserializeBinaryFromReader(message: ListUserSessionsResponse, reader: jspb.BinaryReader): ListUserSessionsResponse;
}

export namespace ListUserSessionsResponse {
    export type AsObject = {
        sessionsList: Array<Session.AsObject>,
        nextPageToken: number,
        totalCount: number,
    }
}

export class BulkDeleteSessionsRequest extends jspb.Message { 
    getUserId(): string;
    setUserId(value: string): BulkDeleteSessionsRequest;
    clearSessionIdsList(): void;
    getSessionIdsList(): Array<string>;
    setSessionIdsList(value: Array<string>): BulkDeleteSessionsRequest;
    addSessionIds(value: string, index?: number): string;
    getFilter(): string;
    setFilter(value: string): BulkDeleteSessionsRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BulkDeleteSessionsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: BulkDeleteSessionsRequest): BulkDeleteSessionsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BulkDeleteSessionsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BulkDeleteSessionsRequest;
    static deserializeBinaryFromReader(message: BulkDeleteSessionsRequest, reader: jspb.BinaryReader): BulkDeleteSessionsRequest;
}

export namespace BulkDeleteSessionsRequest {
    export type AsObject = {
        userId: string,
        sessionIdsList: Array<string>,
        filter: string,
    }
}

export class BulkDeleteSessionsResponse extends jspb.Message { 
    getDeletedCount(): number;
    setDeletedCount(value: number): BulkDeleteSessionsResponse;
    clearFailedDeletionsList(): void;
    getFailedDeletionsList(): Array<string>;
    setFailedDeletionsList(value: Array<string>): BulkDeleteSessionsResponse;
    addFailedDeletions(value: string, index?: number): string;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BulkDeleteSessionsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: BulkDeleteSessionsResponse): BulkDeleteSessionsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BulkDeleteSessionsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BulkDeleteSessionsResponse;
    static deserializeBinaryFromReader(message: BulkDeleteSessionsResponse, reader: jspb.BinaryReader): BulkDeleteSessionsResponse;
}

export namespace BulkDeleteSessionsResponse {
    export type AsObject = {
        deletedCount: number,
        failedDeletionsList: Array<string>,
    }
}

export class GetSessionStatsRequest extends jspb.Message { 
    getUserId(): string;
    setUserId(value: string): GetSessionStatsRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetSessionStatsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetSessionStatsRequest): GetSessionStatsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetSessionStatsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetSessionStatsRequest;
    static deserializeBinaryFromReader(message: GetSessionStatsRequest, reader: jspb.BinaryReader): GetSessionStatsRequest;
}

export namespace GetSessionStatsRequest {
    export type AsObject = {
        userId: string,
    }
}

export class GetSessionStatsResponse extends jspb.Message { 
    getTotalSessions(): number;
    setTotalSessions(value: number): GetSessionStatsResponse;
    getActiveSessions(): number;
    setActiveSessions(value: number): GetSessionStatsResponse;
    getExpiredSessions(): number;
    setExpiredSessions(value: number): GetSessionStatsResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetSessionStatsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetSessionStatsResponse): GetSessionStatsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetSessionStatsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetSessionStatsResponse;
    static deserializeBinaryFromReader(message: GetSessionStatsResponse, reader: jspb.BinaryReader): GetSessionStatsResponse;
}

export namespace GetSessionStatsResponse {
    export type AsObject = {
        totalSessions: number,
        activeSessions: number,
        expiredSessions: number,
    }
}

export class RefreshSessionTokenRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): RefreshSessionTokenRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RefreshSessionTokenRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RefreshSessionTokenRequest): RefreshSessionTokenRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RefreshSessionTokenRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RefreshSessionTokenRequest;
    static deserializeBinaryFromReader(message: RefreshSessionTokenRequest, reader: jspb.BinaryReader): RefreshSessionTokenRequest;
}

export namespace RefreshSessionTokenRequest {
    export type AsObject = {
        sessionId: string,
    }
}

export class RefreshSessionTokenResponse extends jspb.Message { 
    getNewToken(): string;
    setNewToken(value: string): RefreshSessionTokenResponse;
    getExpiresAt(): string;
    setExpiresAt(value: string): RefreshSessionTokenResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RefreshSessionTokenResponse.AsObject;
    static toObject(includeInstance: boolean, msg: RefreshSessionTokenResponse): RefreshSessionTokenResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RefreshSessionTokenResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RefreshSessionTokenResponse;
    static deserializeBinaryFromReader(message: RefreshSessionTokenResponse, reader: jspb.BinaryReader): RefreshSessionTokenResponse;
}

export namespace RefreshSessionTokenResponse {
    export type AsObject = {
        newToken: string,
        expiresAt: string,
    }
}

export class InvalidateSessionRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): InvalidateSessionRequest;
    getReason(): string;
    setReason(value: string): InvalidateSessionRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): InvalidateSessionRequest.AsObject;
    static toObject(includeInstance: boolean, msg: InvalidateSessionRequest): InvalidateSessionRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: InvalidateSessionRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): InvalidateSessionRequest;
    static deserializeBinaryFromReader(message: InvalidateSessionRequest, reader: jspb.BinaryReader): InvalidateSessionRequest;
}

export namespace InvalidateSessionRequest {
    export type AsObject = {
        sessionId: string,
        reason: string,
    }
}

export class InvalidateSessionResponse extends jspb.Message { 
    getSuccess(): boolean;
    setSuccess(value: boolean): InvalidateSessionResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): InvalidateSessionResponse.AsObject;
    static toObject(includeInstance: boolean, msg: InvalidateSessionResponse): InvalidateSessionResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: InvalidateSessionResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): InvalidateSessionResponse;
    static deserializeBinaryFromReader(message: InvalidateSessionResponse, reader: jspb.BinaryReader): InvalidateSessionResponse;
}

export namespace InvalidateSessionResponse {
    export type AsObject = {
        success: boolean,
    }
}

export class UpdateSessionRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): UpdateSessionRequest;
    getName(): string;
    setName(value: string): UpdateSessionRequest;
    getDescription(): string;
    setDescription(value: string): UpdateSessionRequest;
    getNamespace(): string;
    setNamespace(value: string): UpdateSessionRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): UpdateSessionRequest.AsObject;
    static toObject(includeInstance: boolean, msg: UpdateSessionRequest): UpdateSessionRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: UpdateSessionRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): UpdateSessionRequest;
    static deserializeBinaryFromReader(message: UpdateSessionRequest, reader: jspb.BinaryReader): UpdateSessionRequest;
}

export namespace UpdateSessionRequest {
    export type AsObject = {
        sessionId: string,
        name: string,
        description: string,
        namespace: string,
    }
}

export class DeleteSessionRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): DeleteSessionRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteSessionRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteSessionRequest): DeleteSessionRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteSessionRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteSessionRequest;
    static deserializeBinaryFromReader(message: DeleteSessionRequest, reader: jspb.BinaryReader): DeleteSessionRequest;
}

export namespace DeleteSessionRequest {
    export type AsObject = {
        sessionId: string,
    }
}

export class DeleteSessionResponse extends jspb.Message { 
    getSuccess(): boolean;
    setSuccess(value: boolean): DeleteSessionResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteSessionResponse.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteSessionResponse): DeleteSessionResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteSessionResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteSessionResponse;
    static deserializeBinaryFromReader(message: DeleteSessionResponse, reader: jspb.BinaryReader): DeleteSessionResponse;
}

export namespace DeleteSessionResponse {
    export type AsObject = {
        success: boolean,
    }
}

export class CreateApiKeyRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): CreateApiKeyRequest;
    getName(): string;
    setName(value: string): CreateApiKeyRequest;
    clearCapabilitiesList(): void;
    getCapabilitiesList(): Array<string>;
    setCapabilitiesList(value: Array<string>): CreateApiKeyRequest;
    addCapabilities(value: string, index?: number): string;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreateApiKeyRequest.AsObject;
    static toObject(includeInstance: boolean, msg: CreateApiKeyRequest): CreateApiKeyRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreateApiKeyRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreateApiKeyRequest;
    static deserializeBinaryFromReader(message: CreateApiKeyRequest, reader: jspb.BinaryReader): CreateApiKeyRequest;
}

export namespace CreateApiKeyRequest {
    export type AsObject = {
        sessionId: string,
        name: string,
        capabilitiesList: Array<string>,
    }
}

export class ListApiKeysRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): ListApiKeysRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListApiKeysRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListApiKeysRequest): ListApiKeysRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListApiKeysRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListApiKeysRequest;
    static deserializeBinaryFromReader(message: ListApiKeysRequest, reader: jspb.BinaryReader): ListApiKeysRequest;
}

export namespace ListApiKeysRequest {
    export type AsObject = {
        sessionId: string,
    }
}

export class ListApiKeysResponse extends jspb.Message { 
    clearApiKeysList(): void;
    getApiKeysList(): Array<ApiKey>;
    setApiKeysList(value: Array<ApiKey>): ListApiKeysResponse;
    addApiKeys(value?: ApiKey, index?: number): ApiKey;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListApiKeysResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ListApiKeysResponse): ListApiKeysResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListApiKeysResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListApiKeysResponse;
    static deserializeBinaryFromReader(message: ListApiKeysResponse, reader: jspb.BinaryReader): ListApiKeysResponse;
}

export namespace ListApiKeysResponse {
    export type AsObject = {
        apiKeysList: Array<ApiKey.AsObject>,
    }
}

export class RevokeApiKeyRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): RevokeApiKeyRequest;
    getKeyId(): string;
    setKeyId(value: string): RevokeApiKeyRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RevokeApiKeyRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RevokeApiKeyRequest): RevokeApiKeyRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RevokeApiKeyRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RevokeApiKeyRequest;
    static deserializeBinaryFromReader(message: RevokeApiKeyRequest, reader: jspb.BinaryReader): RevokeApiKeyRequest;
}

export namespace RevokeApiKeyRequest {
    export type AsObject = {
        sessionId: string,
        keyId: string,
    }
}

export class RevokeApiKeyResponse extends jspb.Message { 
    getSuccess(): boolean;
    setSuccess(value: boolean): RevokeApiKeyResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RevokeApiKeyResponse.AsObject;
    static toObject(includeInstance: boolean, msg: RevokeApiKeyResponse): RevokeApiKeyResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RevokeApiKeyResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RevokeApiKeyResponse;
    static deserializeBinaryFromReader(message: RevokeApiKeyResponse, reader: jspb.BinaryReader): RevokeApiKeyResponse;
}

export namespace RevokeApiKeyResponse {
    export type AsObject = {
        success: boolean,
    }
}

export class RegisterMachineRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): RegisterMachineRequest;
    getMachineId(): string;
    setMachineId(value: string): RegisterMachineRequest;
    getSdkVersion(): string;
    setSdkVersion(value: string): RegisterMachineRequest;
    getSdkLanguage(): string;
    setSdkLanguage(value: string): RegisterMachineRequest;
    clearToolsList(): void;
    getToolsList(): Array<RegisterToolRequest>;
    setToolsList(value: Array<RegisterToolRequest>): RegisterMachineRequest;
    addTools(value?: RegisterToolRequest, index?: number): RegisterToolRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RegisterMachineRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RegisterMachineRequest): RegisterMachineRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RegisterMachineRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RegisterMachineRequest;
    static deserializeBinaryFromReader(message: RegisterMachineRequest, reader: jspb.BinaryReader): RegisterMachineRequest;
}

export namespace RegisterMachineRequest {
    export type AsObject = {
        sessionId: string,
        machineId: string,
        sdkVersion: string,
        sdkLanguage: string,
        toolsList: Array<RegisterToolRequest.AsObject>,
    }
}

export class ListMachinesRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): ListMachinesRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListMachinesRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListMachinesRequest): ListMachinesRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListMachinesRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListMachinesRequest;
    static deserializeBinaryFromReader(message: ListMachinesRequest, reader: jspb.BinaryReader): ListMachinesRequest;
}

export namespace ListMachinesRequest {
    export type AsObject = {
        sessionId: string,
    }
}

export class ListMachinesResponse extends jspb.Message { 
    clearMachinesList(): void;
    getMachinesList(): Array<Machine>;
    setMachinesList(value: Array<Machine>): ListMachinesResponse;
    addMachines(value?: Machine, index?: number): Machine;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListMachinesResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ListMachinesResponse): ListMachinesResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListMachinesResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListMachinesResponse;
    static deserializeBinaryFromReader(message: ListMachinesResponse, reader: jspb.BinaryReader): ListMachinesResponse;
}

export namespace ListMachinesResponse {
    export type AsObject = {
        machinesList: Array<Machine.AsObject>,
    }
}

export class GetMachineRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): GetMachineRequest;
    getMachineId(): string;
    setMachineId(value: string): GetMachineRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetMachineRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetMachineRequest): GetMachineRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetMachineRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetMachineRequest;
    static deserializeBinaryFromReader(message: GetMachineRequest, reader: jspb.BinaryReader): GetMachineRequest;
}

export namespace GetMachineRequest {
    export type AsObject = {
        sessionId: string,
        machineId: string,
    }
}

export class UpdateMachinePingRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): UpdateMachinePingRequest;
    getMachineId(): string;
    setMachineId(value: string): UpdateMachinePingRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): UpdateMachinePingRequest.AsObject;
    static toObject(includeInstance: boolean, msg: UpdateMachinePingRequest): UpdateMachinePingRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: UpdateMachinePingRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): UpdateMachinePingRequest;
    static deserializeBinaryFromReader(message: UpdateMachinePingRequest, reader: jspb.BinaryReader): UpdateMachinePingRequest;
}

export namespace UpdateMachinePingRequest {
    export type AsObject = {
        sessionId: string,
        machineId: string,
    }
}

export class UnregisterMachineRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): UnregisterMachineRequest;
    getMachineId(): string;
    setMachineId(value: string): UnregisterMachineRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): UnregisterMachineRequest.AsObject;
    static toObject(includeInstance: boolean, msg: UnregisterMachineRequest): UnregisterMachineRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: UnregisterMachineRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): UnregisterMachineRequest;
    static deserializeBinaryFromReader(message: UnregisterMachineRequest, reader: jspb.BinaryReader): UnregisterMachineRequest;
}

export namespace UnregisterMachineRequest {
    export type AsObject = {
        sessionId: string,
        machineId: string,
    }
}

export class UnregisterMachineResponse extends jspb.Message { 
    getSuccess(): boolean;
    setSuccess(value: boolean): UnregisterMachineResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): UnregisterMachineResponse.AsObject;
    static toObject(includeInstance: boolean, msg: UnregisterMachineResponse): UnregisterMachineResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: UnregisterMachineResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): UnregisterMachineResponse;
    static deserializeBinaryFromReader(message: UnregisterMachineResponse, reader: jspb.BinaryReader): UnregisterMachineResponse;
}

export namespace UnregisterMachineResponse {
    export type AsObject = {
        success: boolean,
    }
}

export class ExecuteToolRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): ExecuteToolRequest;
    getToolName(): string;
    setToolName(value: string): ExecuteToolRequest;
    getInput(): string;
    setInput(value: string): ExecuteToolRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ExecuteToolRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ExecuteToolRequest): ExecuteToolRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ExecuteToolRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ExecuteToolRequest;
    static deserializeBinaryFromReader(message: ExecuteToolRequest, reader: jspb.BinaryReader): ExecuteToolRequest;
}

export namespace ExecuteToolRequest {
    export type AsObject = {
        sessionId: string,
        toolName: string,
        input: string,
    }
}

export class ExecuteToolResponse extends jspb.Message { 
    getRequestId(): string;
    setRequestId(value: string): ExecuteToolResponse;
    getStatus(): string;
    setStatus(value: string): ExecuteToolResponse;
    getResult(): string;
    setResult(value: string): ExecuteToolResponse;
    getResultType(): string;
    setResultType(value: string): ExecuteToolResponse;
    getError(): string;
    setError(value: string): ExecuteToolResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ExecuteToolResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ExecuteToolResponse): ExecuteToolResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ExecuteToolResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ExecuteToolResponse;
    static deserializeBinaryFromReader(message: ExecuteToolResponse, reader: jspb.BinaryReader): ExecuteToolResponse;
}

export namespace ExecuteToolResponse {
    export type AsObject = {
        requestId: string,
        status: string,
        result: string,
        resultType: string,
        error: string,
    }
}

export class ExecuteToolChunk extends jspb.Message { 
    getSeq(): number;
    setSeq(value: number): ExecuteToolChunk;
    getRequestId(): string;
    setRequestId(value: string): ExecuteToolChunk;
    getChunk(): string;
    setChunk(value: string): ExecuteToolChunk;
    getIsFinal(): boolean;
    setIsFinal(value: boolean): ExecuteToolChunk;
    getError(): string;
    setError(value: string): ExecuteToolChunk;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ExecuteToolChunk.AsObject;
    static toObject(includeInstance: boolean, msg: ExecuteToolChunk): ExecuteToolChunk.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ExecuteToolChunk, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ExecuteToolChunk;
    static deserializeBinaryFromReader(message: ExecuteToolChunk, reader: jspb.BinaryReader): ExecuteToolChunk;
}

export namespace ExecuteToolChunk {
    export type AsObject = {
        seq: number,
        requestId: string,
        chunk: string,
        isFinal: boolean,
        error: string,
    }
}

export class CreateRequestRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): CreateRequestRequest;
    getToolName(): string;
    setToolName(value: string): CreateRequestRequest;
    getInput(): string;
    setInput(value: string): CreateRequestRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreateRequestRequest.AsObject;
    static toObject(includeInstance: boolean, msg: CreateRequestRequest): CreateRequestRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreateRequestRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreateRequestRequest;
    static deserializeBinaryFromReader(message: CreateRequestRequest, reader: jspb.BinaryReader): CreateRequestRequest;
}

export namespace CreateRequestRequest {
    export type AsObject = {
        sessionId: string,
        toolName: string,
        input: string,
    }
}

export class GetRequestRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): GetRequestRequest;
    getRequestId(): string;
    setRequestId(value: string): GetRequestRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetRequestRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetRequestRequest): GetRequestRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetRequestRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetRequestRequest;
    static deserializeBinaryFromReader(message: GetRequestRequest, reader: jspb.BinaryReader): GetRequestRequest;
}

export namespace GetRequestRequest {
    export type AsObject = {
        sessionId: string,
        requestId: string,
    }
}

export class ListRequestsRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): ListRequestsRequest;
    getStatus(): string;
    setStatus(value: string): ListRequestsRequest;
    getToolName(): string;
    setToolName(value: string): ListRequestsRequest;
    getLimit(): number;
    setLimit(value: number): ListRequestsRequest;
    getOffset(): number;
    setOffset(value: number): ListRequestsRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListRequestsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListRequestsRequest): ListRequestsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListRequestsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListRequestsRequest;
    static deserializeBinaryFromReader(message: ListRequestsRequest, reader: jspb.BinaryReader): ListRequestsRequest;
}

export namespace ListRequestsRequest {
    export type AsObject = {
        sessionId: string,
        status: string,
        toolName: string,
        limit: number,
        offset: number,
    }
}

export class ListRequestsResponse extends jspb.Message { 
    clearRequestsList(): void;
    getRequestsList(): Array<Request>;
    setRequestsList(value: Array<Request>): ListRequestsResponse;
    addRequests(value?: Request, index?: number): Request;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListRequestsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ListRequestsResponse): ListRequestsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListRequestsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListRequestsResponse;
    static deserializeBinaryFromReader(message: ListRequestsResponse, reader: jspb.BinaryReader): ListRequestsResponse;
}

export namespace ListRequestsResponse {
    export type AsObject = {
        requestsList: Array<Request.AsObject>,
    }
}

export class UpdateRequestRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): UpdateRequestRequest;
    getRequestId(): string;
    setRequestId(value: string): UpdateRequestRequest;
    getStatus(): string;
    setStatus(value: string): UpdateRequestRequest;
    getResult(): string;
    setResult(value: string): UpdateRequestRequest;
    getResultType(): string;
    setResultType(value: string): UpdateRequestRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): UpdateRequestRequest.AsObject;
    static toObject(includeInstance: boolean, msg: UpdateRequestRequest): UpdateRequestRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: UpdateRequestRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): UpdateRequestRequest;
    static deserializeBinaryFromReader(message: UpdateRequestRequest, reader: jspb.BinaryReader): UpdateRequestRequest;
}

export namespace UpdateRequestRequest {
    export type AsObject = {
        sessionId: string,
        requestId: string,
        status: string,
        result: string,
        resultType: string,
    }
}

export class ClaimRequestRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): ClaimRequestRequest;
    getRequestId(): string;
    setRequestId(value: string): ClaimRequestRequest;
    getMachineId(): string;
    setMachineId(value: string): ClaimRequestRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ClaimRequestRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ClaimRequestRequest): ClaimRequestRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ClaimRequestRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ClaimRequestRequest;
    static deserializeBinaryFromReader(message: ClaimRequestRequest, reader: jspb.BinaryReader): ClaimRequestRequest;
}

export namespace ClaimRequestRequest {
    export type AsObject = {
        sessionId: string,
        requestId: string,
        machineId: string,
    }
}

export class CancelRequestRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): CancelRequestRequest;
    getRequestId(): string;
    setRequestId(value: string): CancelRequestRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CancelRequestRequest.AsObject;
    static toObject(includeInstance: boolean, msg: CancelRequestRequest): CancelRequestRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CancelRequestRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CancelRequestRequest;
    static deserializeBinaryFromReader(message: CancelRequestRequest, reader: jspb.BinaryReader): CancelRequestRequest;
}

export namespace CancelRequestRequest {
    export type AsObject = {
        sessionId: string,
        requestId: string,
    }
}

export class CancelRequestResponse extends jspb.Message { 
    getSuccess(): boolean;
    setSuccess(value: boolean): CancelRequestResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CancelRequestResponse.AsObject;
    static toObject(includeInstance: boolean, msg: CancelRequestResponse): CancelRequestResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CancelRequestResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CancelRequestResponse;
    static deserializeBinaryFromReader(message: CancelRequestResponse, reader: jspb.BinaryReader): CancelRequestResponse;
}

export namespace CancelRequestResponse {
    export type AsObject = {
        success: boolean,
    }
}

export class SubmitRequestResultRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): SubmitRequestResultRequest;
    getRequestId(): string;
    setRequestId(value: string): SubmitRequestResultRequest;
    getResult(): string;
    setResult(value: string): SubmitRequestResultRequest;
    getResultType(): string;
    setResultType(value: string): SubmitRequestResultRequest;

    getMetaMap(): jspb.Map<string, string>;
    clearMetaMap(): void;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SubmitRequestResultRequest.AsObject;
    static toObject(includeInstance: boolean, msg: SubmitRequestResultRequest): SubmitRequestResultRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SubmitRequestResultRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SubmitRequestResultRequest;
    static deserializeBinaryFromReader(message: SubmitRequestResultRequest, reader: jspb.BinaryReader): SubmitRequestResultRequest;
}

export namespace SubmitRequestResultRequest {
    export type AsObject = {
        sessionId: string,
        requestId: string,
        result: string,
        resultType: string,

        metaMap: Array<[string, string]>,
    }
}

export class SubmitRequestResultResponse extends jspb.Message { 
    getSuccess(): boolean;
    setSuccess(value: boolean): SubmitRequestResultResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SubmitRequestResultResponse.AsObject;
    static toObject(includeInstance: boolean, msg: SubmitRequestResultResponse): SubmitRequestResultResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SubmitRequestResultResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SubmitRequestResultResponse;
    static deserializeBinaryFromReader(message: SubmitRequestResultResponse, reader: jspb.BinaryReader): SubmitRequestResultResponse;
}

export namespace SubmitRequestResultResponse {
    export type AsObject = {
        success: boolean,
    }
}

export class AppendRequestChunksRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): AppendRequestChunksRequest;
    getRequestId(): string;
    setRequestId(value: string): AppendRequestChunksRequest;
    clearChunksList(): void;
    getChunksList(): Array<string>;
    setChunksList(value: Array<string>): AppendRequestChunksRequest;
    addChunks(value: string, index?: number): string;
    getResultType(): string;
    setResultType(value: string): AppendRequestChunksRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AppendRequestChunksRequest.AsObject;
    static toObject(includeInstance: boolean, msg: AppendRequestChunksRequest): AppendRequestChunksRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AppendRequestChunksRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AppendRequestChunksRequest;
    static deserializeBinaryFromReader(message: AppendRequestChunksRequest, reader: jspb.BinaryReader): AppendRequestChunksRequest;
}

export namespace AppendRequestChunksRequest {
    export type AsObject = {
        sessionId: string,
        requestId: string,
        chunksList: Array<string>,
        resultType: string,
    }
}

export class AppendRequestChunksResponse extends jspb.Message { 
    getSuccess(): boolean;
    setSuccess(value: boolean): AppendRequestChunksResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AppendRequestChunksResponse.AsObject;
    static toObject(includeInstance: boolean, msg: AppendRequestChunksResponse): AppendRequestChunksResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AppendRequestChunksResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AppendRequestChunksResponse;
    static deserializeBinaryFromReader(message: AppendRequestChunksResponse, reader: jspb.BinaryReader): AppendRequestChunksResponse;
}

export namespace AppendRequestChunksResponse {
    export type AsObject = {
        success: boolean,
    }
}

export class GetRequestChunksRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): GetRequestChunksRequest;
    getRequestId(): string;
    setRequestId(value: string): GetRequestChunksRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetRequestChunksRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetRequestChunksRequest): GetRequestChunksRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetRequestChunksRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetRequestChunksRequest;
    static deserializeBinaryFromReader(message: GetRequestChunksRequest, reader: jspb.BinaryReader): GetRequestChunksRequest;
}

export namespace GetRequestChunksRequest {
    export type AsObject = {
        sessionId: string,
        requestId: string,
    }
}

export class GetRequestChunksResponse extends jspb.Message { 
    clearChunksList(): void;
    getChunksList(): Array<string>;
    setChunksList(value: Array<string>): GetRequestChunksResponse;
    addChunks(value: string, index?: number): string;
    getStartSeq(): number;
    setStartSeq(value: number): GetRequestChunksResponse;
    getNextSeq(): number;
    setNextSeq(value: number): GetRequestChunksResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetRequestChunksResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetRequestChunksResponse): GetRequestChunksResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetRequestChunksResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetRequestChunksResponse;
    static deserializeBinaryFromReader(message: GetRequestChunksResponse, reader: jspb.BinaryReader): GetRequestChunksResponse;
}

export namespace GetRequestChunksResponse {
    export type AsObject = {
        chunksList: Array<string>,
        startSeq: number,
        nextSeq: number,
    }
}

export class HealthCheckRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): HealthCheckRequest.AsObject;
    static toObject(includeInstance: boolean, msg: HealthCheckRequest): HealthCheckRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: HealthCheckRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): HealthCheckRequest;
    static deserializeBinaryFromReader(message: HealthCheckRequest, reader: jspb.BinaryReader): HealthCheckRequest;
}

export namespace HealthCheckRequest {
    export type AsObject = {
    }
}

export class HealthCheckResponse extends jspb.Message { 
    getStatus(): string;
    setStatus(value: string): HealthCheckResponse;
    getVersion(): string;
    setVersion(value: string): HealthCheckResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): HealthCheckResponse.AsObject;
    static toObject(includeInstance: boolean, msg: HealthCheckResponse): HealthCheckResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: HealthCheckResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): HealthCheckResponse;
    static deserializeBinaryFromReader(message: HealthCheckResponse, reader: jspb.BinaryReader): HealthCheckResponse;
}

export namespace HealthCheckResponse {
    export type AsObject = {
        status: string,
        version: string,
    }
}

export class Task extends jspb.Message { 
    getId(): string;
    setId(value: string): Task;
    getSessionId(): string;
    setSessionId(value: string): Task;
    getToolName(): string;
    setToolName(value: string): Task;
    getStatus(): string;
    setStatus(value: string): Task;
    getInput(): string;
    setInput(value: string): Task;
    getResult(): string;
    setResult(value: string): Task;
    getResultType(): string;
    setResultType(value: string): Task;
    getError(): string;
    setError(value: string): Task;
    getCreatedAt(): string;
    setCreatedAt(value: string): Task;
    getUpdatedAt(): string;
    setUpdatedAt(value: string): Task;
    getCompletedAt(): string;
    setCompletedAt(value: string): Task;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Task.AsObject;
    static toObject(includeInstance: boolean, msg: Task): Task.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Task, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Task;
    static deserializeBinaryFromReader(message: Task, reader: jspb.BinaryReader): Task;
}

export namespace Task {
    export type AsObject = {
        id: string,
        sessionId: string,
        toolName: string,
        status: string,
        input: string,
        result: string,
        resultType: string,
        error: string,
        createdAt: string,
        updatedAt: string,
        completedAt: string,
    }
}

export class CreateTaskRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): CreateTaskRequest;
    getToolName(): string;
    setToolName(value: string): CreateTaskRequest;
    getInput(): string;
    setInput(value: string): CreateTaskRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreateTaskRequest.AsObject;
    static toObject(includeInstance: boolean, msg: CreateTaskRequest): CreateTaskRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreateTaskRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreateTaskRequest;
    static deserializeBinaryFromReader(message: CreateTaskRequest, reader: jspb.BinaryReader): CreateTaskRequest;
}

export namespace CreateTaskRequest {
    export type AsObject = {
        sessionId: string,
        toolName: string,
        input: string,
    }
}

export class GetTaskRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): GetTaskRequest;
    getTaskId(): string;
    setTaskId(value: string): GetTaskRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetTaskRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetTaskRequest): GetTaskRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetTaskRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetTaskRequest;
    static deserializeBinaryFromReader(message: GetTaskRequest, reader: jspb.BinaryReader): GetTaskRequest;
}

export namespace GetTaskRequest {
    export type AsObject = {
        sessionId: string,
        taskId: string,
    }
}

export class ListTasksRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): ListTasksRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListTasksRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListTasksRequest): ListTasksRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListTasksRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListTasksRequest;
    static deserializeBinaryFromReader(message: ListTasksRequest, reader: jspb.BinaryReader): ListTasksRequest;
}

export namespace ListTasksRequest {
    export type AsObject = {
        sessionId: string,
    }
}

export class ListTasksResponse extends jspb.Message { 
    clearTasksList(): void;
    getTasksList(): Array<Task>;
    setTasksList(value: Array<Task>): ListTasksResponse;
    addTasks(value?: Task, index?: number): Task;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListTasksResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ListTasksResponse): ListTasksResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListTasksResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListTasksResponse;
    static deserializeBinaryFromReader(message: ListTasksResponse, reader: jspb.BinaryReader): ListTasksResponse;
}

export namespace ListTasksResponse {
    export type AsObject = {
        tasksList: Array<Task.AsObject>,
    }
}

export class CancelTaskRequest extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): CancelTaskRequest;
    getTaskId(): string;
    setTaskId(value: string): CancelTaskRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CancelTaskRequest.AsObject;
    static toObject(includeInstance: boolean, msg: CancelTaskRequest): CancelTaskRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CancelTaskRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CancelTaskRequest;
    static deserializeBinaryFromReader(message: CancelTaskRequest, reader: jspb.BinaryReader): CancelTaskRequest;
}

export namespace CancelTaskRequest {
    export type AsObject = {
        sessionId: string,
        taskId: string,
    }
}

export class CancelTaskResponse extends jspb.Message { 
    getSuccess(): boolean;
    setSuccess(value: boolean): CancelTaskResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CancelTaskResponse.AsObject;
    static toObject(includeInstance: boolean, msg: CancelTaskResponse): CancelTaskResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CancelTaskResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CancelTaskResponse;
    static deserializeBinaryFromReader(message: CancelTaskResponse, reader: jspb.BinaryReader): CancelTaskResponse;
}

export namespace CancelTaskResponse {
    export type AsObject = {
        success: boolean,
    }
}
