// GENERATED CODE -- DO NOT EDIT!

// Original file comments:
// SYNCED COPY — DO NOT EDIT.
// Source: server/proto/service.proto
// Regenerate: cd server && make gen-proto-all
//
'use strict';
var grpc = require('@grpc/grpc-js');
var proto_service_pb = require('../proto/service_pb.js');
var google_api_annotations_pb = require('../google/api/annotations_pb.js');

function serialize_api_ApiKey(arg) {
  if (!(arg instanceof proto_service_pb.ApiKey)) {
    throw new Error('Expected argument of type api.ApiKey');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ApiKey(buffer_arg) {
  return proto_service_pb.ApiKey.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_AppendRequestChunksRequest(arg) {
  if (!(arg instanceof proto_service_pb.AppendRequestChunksRequest)) {
    throw new Error('Expected argument of type api.AppendRequestChunksRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_AppendRequestChunksRequest(buffer_arg) {
  return proto_service_pb.AppendRequestChunksRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_AppendRequestChunksResponse(arg) {
  if (!(arg instanceof proto_service_pb.AppendRequestChunksResponse)) {
    throw new Error('Expected argument of type api.AppendRequestChunksResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_AppendRequestChunksResponse(buffer_arg) {
  return proto_service_pb.AppendRequestChunksResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_BulkDeleteSessionsRequest(arg) {
  if (!(arg instanceof proto_service_pb.BulkDeleteSessionsRequest)) {
    throw new Error('Expected argument of type api.BulkDeleteSessionsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_BulkDeleteSessionsRequest(buffer_arg) {
  return proto_service_pb.BulkDeleteSessionsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_BulkDeleteSessionsResponse(arg) {
  if (!(arg instanceof proto_service_pb.BulkDeleteSessionsResponse)) {
    throw new Error('Expected argument of type api.BulkDeleteSessionsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_BulkDeleteSessionsResponse(buffer_arg) {
  return proto_service_pb.BulkDeleteSessionsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_CancelRequestRequest(arg) {
  if (!(arg instanceof proto_service_pb.CancelRequestRequest)) {
    throw new Error('Expected argument of type api.CancelRequestRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_CancelRequestRequest(buffer_arg) {
  return proto_service_pb.CancelRequestRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_CancelRequestResponse(arg) {
  if (!(arg instanceof proto_service_pb.CancelRequestResponse)) {
    throw new Error('Expected argument of type api.CancelRequestResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_CancelRequestResponse(buffer_arg) {
  return proto_service_pb.CancelRequestResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_CancelTaskRequest(arg) {
  if (!(arg instanceof proto_service_pb.CancelTaskRequest)) {
    throw new Error('Expected argument of type api.CancelTaskRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_CancelTaskRequest(buffer_arg) {
  return proto_service_pb.CancelTaskRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_CancelTaskResponse(arg) {
  if (!(arg instanceof proto_service_pb.CancelTaskResponse)) {
    throw new Error('Expected argument of type api.CancelTaskResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_CancelTaskResponse(buffer_arg) {
  return proto_service_pb.CancelTaskResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_ClaimRequestRequest(arg) {
  if (!(arg instanceof proto_service_pb.ClaimRequestRequest)) {
    throw new Error('Expected argument of type api.ClaimRequestRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ClaimRequestRequest(buffer_arg) {
  return proto_service_pb.ClaimRequestRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_CreateApiKeyRequest(arg) {
  if (!(arg instanceof proto_service_pb.CreateApiKeyRequest)) {
    throw new Error('Expected argument of type api.CreateApiKeyRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_CreateApiKeyRequest(buffer_arg) {
  return proto_service_pb.CreateApiKeyRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_CreateRequestRequest(arg) {
  if (!(arg instanceof proto_service_pb.CreateRequestRequest)) {
    throw new Error('Expected argument of type api.CreateRequestRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_CreateRequestRequest(buffer_arg) {
  return proto_service_pb.CreateRequestRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_CreateSessionRequest(arg) {
  if (!(arg instanceof proto_service_pb.CreateSessionRequest)) {
    throw new Error('Expected argument of type api.CreateSessionRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_CreateSessionRequest(buffer_arg) {
  return proto_service_pb.CreateSessionRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_CreateSessionResponse(arg) {
  if (!(arg instanceof proto_service_pb.CreateSessionResponse)) {
    throw new Error('Expected argument of type api.CreateSessionResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_CreateSessionResponse(buffer_arg) {
  return proto_service_pb.CreateSessionResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_CreateTaskRequest(arg) {
  if (!(arg instanceof proto_service_pb.CreateTaskRequest)) {
    throw new Error('Expected argument of type api.CreateTaskRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_CreateTaskRequest(buffer_arg) {
  return proto_service_pb.CreateTaskRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_DeleteSessionRequest(arg) {
  if (!(arg instanceof proto_service_pb.DeleteSessionRequest)) {
    throw new Error('Expected argument of type api.DeleteSessionRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_DeleteSessionRequest(buffer_arg) {
  return proto_service_pb.DeleteSessionRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_DeleteSessionResponse(arg) {
  if (!(arg instanceof proto_service_pb.DeleteSessionResponse)) {
    throw new Error('Expected argument of type api.DeleteSessionResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_DeleteSessionResponse(buffer_arg) {
  return proto_service_pb.DeleteSessionResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_DeleteToolRequest(arg) {
  if (!(arg instanceof proto_service_pb.DeleteToolRequest)) {
    throw new Error('Expected argument of type api.DeleteToolRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_DeleteToolRequest(buffer_arg) {
  return proto_service_pb.DeleteToolRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_DeleteToolResponse(arg) {
  if (!(arg instanceof proto_service_pb.DeleteToolResponse)) {
    throw new Error('Expected argument of type api.DeleteToolResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_DeleteToolResponse(buffer_arg) {
  return proto_service_pb.DeleteToolResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_DrainMachineRequest(arg) {
  if (!(arg instanceof proto_service_pb.DrainMachineRequest)) {
    throw new Error('Expected argument of type api.DrainMachineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_DrainMachineRequest(buffer_arg) {
  return proto_service_pb.DrainMachineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_DrainMachineResponse(arg) {
  if (!(arg instanceof proto_service_pb.DrainMachineResponse)) {
    throw new Error('Expected argument of type api.DrainMachineResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_DrainMachineResponse(buffer_arg) {
  return proto_service_pb.DrainMachineResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_ExecuteToolChunk(arg) {
  if (!(arg instanceof proto_service_pb.ExecuteToolChunk)) {
    throw new Error('Expected argument of type api.ExecuteToolChunk');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ExecuteToolChunk(buffer_arg) {
  return proto_service_pb.ExecuteToolChunk.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_ExecuteToolRequest(arg) {
  if (!(arg instanceof proto_service_pb.ExecuteToolRequest)) {
    throw new Error('Expected argument of type api.ExecuteToolRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ExecuteToolRequest(buffer_arg) {
  return proto_service_pb.ExecuteToolRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_ExecuteToolResponse(arg) {
  if (!(arg instanceof proto_service_pb.ExecuteToolResponse)) {
    throw new Error('Expected argument of type api.ExecuteToolResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ExecuteToolResponse(buffer_arg) {
  return proto_service_pb.ExecuteToolResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_GetMachineRequest(arg) {
  if (!(arg instanceof proto_service_pb.GetMachineRequest)) {
    throw new Error('Expected argument of type api.GetMachineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_GetMachineRequest(buffer_arg) {
  return proto_service_pb.GetMachineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_GetRequestChunksRequest(arg) {
  if (!(arg instanceof proto_service_pb.GetRequestChunksRequest)) {
    throw new Error('Expected argument of type api.GetRequestChunksRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_GetRequestChunksRequest(buffer_arg) {
  return proto_service_pb.GetRequestChunksRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_GetRequestChunksResponse(arg) {
  if (!(arg instanceof proto_service_pb.GetRequestChunksResponse)) {
    throw new Error('Expected argument of type api.GetRequestChunksResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_GetRequestChunksResponse(buffer_arg) {
  return proto_service_pb.GetRequestChunksResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_GetRequestRequest(arg) {
  if (!(arg instanceof proto_service_pb.GetRequestRequest)) {
    throw new Error('Expected argument of type api.GetRequestRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_GetRequestRequest(buffer_arg) {
  return proto_service_pb.GetRequestRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_GetSessionRequest(arg) {
  if (!(arg instanceof proto_service_pb.GetSessionRequest)) {
    throw new Error('Expected argument of type api.GetSessionRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_GetSessionRequest(buffer_arg) {
  return proto_service_pb.GetSessionRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_GetSessionStatsRequest(arg) {
  if (!(arg instanceof proto_service_pb.GetSessionStatsRequest)) {
    throw new Error('Expected argument of type api.GetSessionStatsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_GetSessionStatsRequest(buffer_arg) {
  return proto_service_pb.GetSessionStatsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_GetSessionStatsResponse(arg) {
  if (!(arg instanceof proto_service_pb.GetSessionStatsResponse)) {
    throw new Error('Expected argument of type api.GetSessionStatsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_GetSessionStatsResponse(buffer_arg) {
  return proto_service_pb.GetSessionStatsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_GetTaskRequest(arg) {
  if (!(arg instanceof proto_service_pb.GetTaskRequest)) {
    throw new Error('Expected argument of type api.GetTaskRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_GetTaskRequest(buffer_arg) {
  return proto_service_pb.GetTaskRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_GetToolByIdRequest(arg) {
  if (!(arg instanceof proto_service_pb.GetToolByIdRequest)) {
    throw new Error('Expected argument of type api.GetToolByIdRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_GetToolByIdRequest(buffer_arg) {
  return proto_service_pb.GetToolByIdRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_GetToolByNameRequest(arg) {
  if (!(arg instanceof proto_service_pb.GetToolByNameRequest)) {
    throw new Error('Expected argument of type api.GetToolByNameRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_GetToolByNameRequest(buffer_arg) {
  return proto_service_pb.GetToolByNameRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_GetToolResponse(arg) {
  if (!(arg instanceof proto_service_pb.GetToolResponse)) {
    throw new Error('Expected argument of type api.GetToolResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_GetToolResponse(buffer_arg) {
  return proto_service_pb.GetToolResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_HealthCheckRequest(arg) {
  if (!(arg instanceof proto_service_pb.HealthCheckRequest)) {
    throw new Error('Expected argument of type api.HealthCheckRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_HealthCheckRequest(buffer_arg) {
  return proto_service_pb.HealthCheckRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_HealthCheckResponse(arg) {
  if (!(arg instanceof proto_service_pb.HealthCheckResponse)) {
    throw new Error('Expected argument of type api.HealthCheckResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_HealthCheckResponse(buffer_arg) {
  return proto_service_pb.HealthCheckResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_InvalidateSessionRequest(arg) {
  if (!(arg instanceof proto_service_pb.InvalidateSessionRequest)) {
    throw new Error('Expected argument of type api.InvalidateSessionRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_InvalidateSessionRequest(buffer_arg) {
  return proto_service_pb.InvalidateSessionRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_InvalidateSessionResponse(arg) {
  if (!(arg instanceof proto_service_pb.InvalidateSessionResponse)) {
    throw new Error('Expected argument of type api.InvalidateSessionResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_InvalidateSessionResponse(buffer_arg) {
  return proto_service_pb.InvalidateSessionResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_ListApiKeysRequest(arg) {
  if (!(arg instanceof proto_service_pb.ListApiKeysRequest)) {
    throw new Error('Expected argument of type api.ListApiKeysRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ListApiKeysRequest(buffer_arg) {
  return proto_service_pb.ListApiKeysRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_ListApiKeysResponse(arg) {
  if (!(arg instanceof proto_service_pb.ListApiKeysResponse)) {
    throw new Error('Expected argument of type api.ListApiKeysResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ListApiKeysResponse(buffer_arg) {
  return proto_service_pb.ListApiKeysResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_ListMachinesRequest(arg) {
  if (!(arg instanceof proto_service_pb.ListMachinesRequest)) {
    throw new Error('Expected argument of type api.ListMachinesRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ListMachinesRequest(buffer_arg) {
  return proto_service_pb.ListMachinesRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_ListMachinesResponse(arg) {
  if (!(arg instanceof proto_service_pb.ListMachinesResponse)) {
    throw new Error('Expected argument of type api.ListMachinesResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ListMachinesResponse(buffer_arg) {
  return proto_service_pb.ListMachinesResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_ListRequestsRequest(arg) {
  if (!(arg instanceof proto_service_pb.ListRequestsRequest)) {
    throw new Error('Expected argument of type api.ListRequestsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ListRequestsRequest(buffer_arg) {
  return proto_service_pb.ListRequestsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_ListRequestsResponse(arg) {
  if (!(arg instanceof proto_service_pb.ListRequestsResponse)) {
    throw new Error('Expected argument of type api.ListRequestsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ListRequestsResponse(buffer_arg) {
  return proto_service_pb.ListRequestsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_ListSessionsRequest(arg) {
  if (!(arg instanceof proto_service_pb.ListSessionsRequest)) {
    throw new Error('Expected argument of type api.ListSessionsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ListSessionsRequest(buffer_arg) {
  return proto_service_pb.ListSessionsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_ListSessionsResponse(arg) {
  if (!(arg instanceof proto_service_pb.ListSessionsResponse)) {
    throw new Error('Expected argument of type api.ListSessionsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ListSessionsResponse(buffer_arg) {
  return proto_service_pb.ListSessionsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_ListTasksRequest(arg) {
  if (!(arg instanceof proto_service_pb.ListTasksRequest)) {
    throw new Error('Expected argument of type api.ListTasksRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ListTasksRequest(buffer_arg) {
  return proto_service_pb.ListTasksRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_ListTasksResponse(arg) {
  if (!(arg instanceof proto_service_pb.ListTasksResponse)) {
    throw new Error('Expected argument of type api.ListTasksResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ListTasksResponse(buffer_arg) {
  return proto_service_pb.ListTasksResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_ListToolsRequest(arg) {
  if (!(arg instanceof proto_service_pb.ListToolsRequest)) {
    throw new Error('Expected argument of type api.ListToolsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ListToolsRequest(buffer_arg) {
  return proto_service_pb.ListToolsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_ListToolsResponse(arg) {
  if (!(arg instanceof proto_service_pb.ListToolsResponse)) {
    throw new Error('Expected argument of type api.ListToolsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ListToolsResponse(buffer_arg) {
  return proto_service_pb.ListToolsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_ListUserSessionsRequest(arg) {
  if (!(arg instanceof proto_service_pb.ListUserSessionsRequest)) {
    throw new Error('Expected argument of type api.ListUserSessionsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ListUserSessionsRequest(buffer_arg) {
  return proto_service_pb.ListUserSessionsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_ListUserSessionsResponse(arg) {
  if (!(arg instanceof proto_service_pb.ListUserSessionsResponse)) {
    throw new Error('Expected argument of type api.ListUserSessionsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ListUserSessionsResponse(buffer_arg) {
  return proto_service_pb.ListUserSessionsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_Machine(arg) {
  if (!(arg instanceof proto_service_pb.Machine)) {
    throw new Error('Expected argument of type api.Machine');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_Machine(buffer_arg) {
  return proto_service_pb.Machine.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_RefreshSessionTokenRequest(arg) {
  if (!(arg instanceof proto_service_pb.RefreshSessionTokenRequest)) {
    throw new Error('Expected argument of type api.RefreshSessionTokenRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_RefreshSessionTokenRequest(buffer_arg) {
  return proto_service_pb.RefreshSessionTokenRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_RefreshSessionTokenResponse(arg) {
  if (!(arg instanceof proto_service_pb.RefreshSessionTokenResponse)) {
    throw new Error('Expected argument of type api.RefreshSessionTokenResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_RefreshSessionTokenResponse(buffer_arg) {
  return proto_service_pb.RefreshSessionTokenResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_RegisterMachineRequest(arg) {
  if (!(arg instanceof proto_service_pb.RegisterMachineRequest)) {
    throw new Error('Expected argument of type api.RegisterMachineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_RegisterMachineRequest(buffer_arg) {
  return proto_service_pb.RegisterMachineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_RegisterToolRequest(arg) {
  if (!(arg instanceof proto_service_pb.RegisterToolRequest)) {
    throw new Error('Expected argument of type api.RegisterToolRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_RegisterToolRequest(buffer_arg) {
  return proto_service_pb.RegisterToolRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_RegisterToolResponse(arg) {
  if (!(arg instanceof proto_service_pb.RegisterToolResponse)) {
    throw new Error('Expected argument of type api.RegisterToolResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_RegisterToolResponse(buffer_arg) {
  return proto_service_pb.RegisterToolResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_Request(arg) {
  if (!(arg instanceof proto_service_pb.Request)) {
    throw new Error('Expected argument of type api.Request');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_Request(buffer_arg) {
  return proto_service_pb.Request.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_ResumeStreamRequest(arg) {
  if (!(arg instanceof proto_service_pb.ResumeStreamRequest)) {
    throw new Error('Expected argument of type api.ResumeStreamRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_ResumeStreamRequest(buffer_arg) {
  return proto_service_pb.ResumeStreamRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_RevokeApiKeyRequest(arg) {
  if (!(arg instanceof proto_service_pb.RevokeApiKeyRequest)) {
    throw new Error('Expected argument of type api.RevokeApiKeyRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_RevokeApiKeyRequest(buffer_arg) {
  return proto_service_pb.RevokeApiKeyRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_RevokeApiKeyResponse(arg) {
  if (!(arg instanceof proto_service_pb.RevokeApiKeyResponse)) {
    throw new Error('Expected argument of type api.RevokeApiKeyResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_RevokeApiKeyResponse(buffer_arg) {
  return proto_service_pb.RevokeApiKeyResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_Session(arg) {
  if (!(arg instanceof proto_service_pb.Session)) {
    throw new Error('Expected argument of type api.Session');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_Session(buffer_arg) {
  return proto_service_pb.Session.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_SubmitRequestResultRequest(arg) {
  if (!(arg instanceof proto_service_pb.SubmitRequestResultRequest)) {
    throw new Error('Expected argument of type api.SubmitRequestResultRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_SubmitRequestResultRequest(buffer_arg) {
  return proto_service_pb.SubmitRequestResultRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_SubmitRequestResultResponse(arg) {
  if (!(arg instanceof proto_service_pb.SubmitRequestResultResponse)) {
    throw new Error('Expected argument of type api.SubmitRequestResultResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_SubmitRequestResultResponse(buffer_arg) {
  return proto_service_pb.SubmitRequestResultResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_Task(arg) {
  if (!(arg instanceof proto_service_pb.Task)) {
    throw new Error('Expected argument of type api.Task');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_Task(buffer_arg) {
  return proto_service_pb.Task.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_Tool(arg) {
  if (!(arg instanceof proto_service_pb.Tool)) {
    throw new Error('Expected argument of type api.Tool');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_Tool(buffer_arg) {
  return proto_service_pb.Tool.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_UnregisterMachineRequest(arg) {
  if (!(arg instanceof proto_service_pb.UnregisterMachineRequest)) {
    throw new Error('Expected argument of type api.UnregisterMachineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_UnregisterMachineRequest(buffer_arg) {
  return proto_service_pb.UnregisterMachineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_UnregisterMachineResponse(arg) {
  if (!(arg instanceof proto_service_pb.UnregisterMachineResponse)) {
    throw new Error('Expected argument of type api.UnregisterMachineResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_UnregisterMachineResponse(buffer_arg) {
  return proto_service_pb.UnregisterMachineResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_UpdateMachinePingRequest(arg) {
  if (!(arg instanceof proto_service_pb.UpdateMachinePingRequest)) {
    throw new Error('Expected argument of type api.UpdateMachinePingRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_UpdateMachinePingRequest(buffer_arg) {
  return proto_service_pb.UpdateMachinePingRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_UpdateRequestRequest(arg) {
  if (!(arg instanceof proto_service_pb.UpdateRequestRequest)) {
    throw new Error('Expected argument of type api.UpdateRequestRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_UpdateRequestRequest(buffer_arg) {
  return proto_service_pb.UpdateRequestRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_UpdateSessionRequest(arg) {
  if (!(arg instanceof proto_service_pb.UpdateSessionRequest)) {
    throw new Error('Expected argument of type api.UpdateSessionRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_UpdateSessionRequest(buffer_arg) {
  return proto_service_pb.UpdateSessionRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_UpdateToolPingRequest(arg) {
  if (!(arg instanceof proto_service_pb.UpdateToolPingRequest)) {
    throw new Error('Expected argument of type api.UpdateToolPingRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_UpdateToolPingRequest(buffer_arg) {
  return proto_service_pb.UpdateToolPingRequest.deserializeBinary(new Uint8Array(buffer_arg));
}


// ======================
// Tool Service
// ======================
var ToolServiceService = exports.ToolServiceService = {
  // Tool management
registerTool: {
    path: '/api.ToolService/RegisterTool',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.RegisterToolRequest,
    responseType: proto_service_pb.RegisterToolResponse,
    requestSerialize: serialize_api_RegisterToolRequest,
    requestDeserialize: deserialize_api_RegisterToolRequest,
    responseSerialize: serialize_api_RegisterToolResponse,
    responseDeserialize: deserialize_api_RegisterToolResponse,
  },
  listTools: {
    path: '/api.ToolService/ListTools',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.ListToolsRequest,
    responseType: proto_service_pb.ListToolsResponse,
    requestSerialize: serialize_api_ListToolsRequest,
    requestDeserialize: deserialize_api_ListToolsRequest,
    responseSerialize: serialize_api_ListToolsResponse,
    responseDeserialize: deserialize_api_ListToolsResponse,
  },
  getToolById: {
    path: '/api.ToolService/GetToolById',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.GetToolByIdRequest,
    responseType: proto_service_pb.GetToolResponse,
    requestSerialize: serialize_api_GetToolByIdRequest,
    requestDeserialize: deserialize_api_GetToolByIdRequest,
    responseSerialize: serialize_api_GetToolResponse,
    responseDeserialize: deserialize_api_GetToolResponse,
  },
  getToolByName: {
    path: '/api.ToolService/GetToolByName',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.GetToolByNameRequest,
    responseType: proto_service_pb.GetToolResponse,
    requestSerialize: serialize_api_GetToolByNameRequest,
    requestDeserialize: deserialize_api_GetToolByNameRequest,
    responseSerialize: serialize_api_GetToolResponse,
    responseDeserialize: deserialize_api_GetToolResponse,
  },
  deleteTool: {
    path: '/api.ToolService/DeleteTool',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.DeleteToolRequest,
    responseType: proto_service_pb.DeleteToolResponse,
    requestSerialize: serialize_api_DeleteToolRequest,
    requestDeserialize: deserialize_api_DeleteToolRequest,
    responseSerialize: serialize_api_DeleteToolResponse,
    responseDeserialize: deserialize_api_DeleteToolResponse,
  },
  updateToolPing: {
    path: '/api.ToolService/UpdateToolPing',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.UpdateToolPingRequest,
    responseType: proto_service_pb.Tool,
    requestSerialize: serialize_api_UpdateToolPingRequest,
    requestDeserialize: deserialize_api_UpdateToolPingRequest,
    responseSerialize: serialize_api_Tool,
    responseDeserialize: deserialize_api_Tool,
  },
  // Consumer execution entrypoints create request-backed work on the control plane.
// Request state, claim, and result lifecycle ownership stays in RequestsService.
streamExecuteTool: {
    path: '/api.ToolService/StreamExecuteTool',
    requestStream: false,
    responseStream: true,
    requestType: proto_service_pb.ExecuteToolRequest,
    responseType: proto_service_pb.ExecuteToolChunk,
    requestSerialize: serialize_api_ExecuteToolRequest,
    requestDeserialize: deserialize_api_ExecuteToolRequest,
    responseSerialize: serialize_api_ExecuteToolChunk,
    responseDeserialize: deserialize_api_ExecuteToolChunk,
  },
  // Resume a broken stream from the last acknowledged sequence number.
// The server retains a bounded chunk window; if the requested sequence has
// fallen out of that window, this RPC fails with OUT_OF_RANGE.
resumeStream: {
    path: '/api.ToolService/ResumeStream',
    requestStream: false,
    responseStream: true,
    requestType: proto_service_pb.ResumeStreamRequest,
    responseType: proto_service_pb.ExecuteToolChunk,
    requestSerialize: serialize_api_ResumeStreamRequest,
    requestDeserialize: deserialize_api_ResumeStreamRequest,
    responseSerialize: serialize_api_ExecuteToolChunk,
    responseDeserialize: deserialize_api_ExecuteToolChunk,
  },
  executeTool: {
    path: '/api.ToolService/ExecuteTool',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.ExecuteToolRequest,
    responseType: proto_service_pb.ExecuteToolResponse,
    requestSerialize: serialize_api_ExecuteToolRequest,
    requestDeserialize: deserialize_api_ExecuteToolRequest,
    responseSerialize: serialize_api_ExecuteToolResponse,
    responseDeserialize: deserialize_api_ExecuteToolResponse,
  },
  // Health check (can be in any service or its own)
healthCheck: {
    path: '/api.ToolService/HealthCheck',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.HealthCheckRequest,
    responseType: proto_service_pb.HealthCheckResponse,
    requestSerialize: serialize_api_HealthCheckRequest,
    requestDeserialize: deserialize_api_HealthCheckRequest,
    responseSerialize: serialize_api_HealthCheckResponse,
    responseDeserialize: deserialize_api_HealthCheckResponse,
  },
};

exports.ToolServiceClient = grpc.makeGenericClientConstructor(ToolServiceService, 'ToolService');
// ======================
// Session Service
// ======================
var SessionsServiceService = exports.SessionsServiceService = {
  // Session management
createSession: {
    path: '/api.SessionsService/CreateSession',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.CreateSessionRequest,
    responseType: proto_service_pb.CreateSessionResponse,
    requestSerialize: serialize_api_CreateSessionRequest,
    requestDeserialize: deserialize_api_CreateSessionRequest,
    responseSerialize: serialize_api_CreateSessionResponse,
    responseDeserialize: deserialize_api_CreateSessionResponse,
  },
  getSession: {
    path: '/api.SessionsService/GetSession',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.GetSessionRequest,
    responseType: proto_service_pb.Session,
    requestSerialize: serialize_api_GetSessionRequest,
    requestDeserialize: deserialize_api_GetSessionRequest,
    responseSerialize: serialize_api_Session,
    responseDeserialize: deserialize_api_Session,
  },
  listSessions: {
    path: '/api.SessionsService/ListSessions',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.ListSessionsRequest,
    responseType: proto_service_pb.ListSessionsResponse,
    requestSerialize: serialize_api_ListSessionsRequest,
    requestDeserialize: deserialize_api_ListSessionsRequest,
    responseSerialize: serialize_api_ListSessionsResponse,
    responseDeserialize: deserialize_api_ListSessionsResponse,
  },
  updateSession: {
    path: '/api.SessionsService/UpdateSession',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.UpdateSessionRequest,
    responseType: proto_service_pb.Session,
    requestSerialize: serialize_api_UpdateSessionRequest,
    requestDeserialize: deserialize_api_UpdateSessionRequest,
    responseSerialize: serialize_api_Session,
    responseDeserialize: deserialize_api_Session,
  },
  deleteSession: {
    path: '/api.SessionsService/DeleteSession',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.DeleteSessionRequest,
    responseType: proto_service_pb.DeleteSessionResponse,
    requestSerialize: serialize_api_DeleteSessionRequest,
    requestDeserialize: deserialize_api_DeleteSessionRequest,
    responseSerialize: serialize_api_DeleteSessionResponse,
    responseDeserialize: deserialize_api_DeleteSessionResponse,
  },
  // Enhanced session management
listUserSessions: {
    path: '/api.SessionsService/ListUserSessions',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.ListUserSessionsRequest,
    responseType: proto_service_pb.ListUserSessionsResponse,
    requestSerialize: serialize_api_ListUserSessionsRequest,
    requestDeserialize: deserialize_api_ListUserSessionsRequest,
    responseSerialize: serialize_api_ListUserSessionsResponse,
    responseDeserialize: deserialize_api_ListUserSessionsResponse,
  },
  bulkDeleteSessions: {
    path: '/api.SessionsService/BulkDeleteSessions',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.BulkDeleteSessionsRequest,
    responseType: proto_service_pb.BulkDeleteSessionsResponse,
    requestSerialize: serialize_api_BulkDeleteSessionsRequest,
    requestDeserialize: deserialize_api_BulkDeleteSessionsRequest,
    responseSerialize: serialize_api_BulkDeleteSessionsResponse,
    responseDeserialize: deserialize_api_BulkDeleteSessionsResponse,
  },
  getSessionStats: {
    path: '/api.SessionsService/GetSessionStats',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.GetSessionStatsRequest,
    responseType: proto_service_pb.GetSessionStatsResponse,
    requestSerialize: serialize_api_GetSessionStatsRequest,
    requestDeserialize: deserialize_api_GetSessionStatsRequest,
    responseSerialize: serialize_api_GetSessionStatsResponse,
    responseDeserialize: deserialize_api_GetSessionStatsResponse,
  },
  refreshSessionToken: {
    path: '/api.SessionsService/RefreshSessionToken',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.RefreshSessionTokenRequest,
    responseType: proto_service_pb.RefreshSessionTokenResponse,
    requestSerialize: serialize_api_RefreshSessionTokenRequest,
    requestDeserialize: deserialize_api_RefreshSessionTokenRequest,
    responseSerialize: serialize_api_RefreshSessionTokenResponse,
    responseDeserialize: deserialize_api_RefreshSessionTokenResponse,
  },
  invalidateSession: {
    path: '/api.SessionsService/InvalidateSession',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.InvalidateSessionRequest,
    responseType: proto_service_pb.InvalidateSessionResponse,
    requestSerialize: serialize_api_InvalidateSessionRequest,
    requestDeserialize: deserialize_api_InvalidateSessionRequest,
    responseSerialize: serialize_api_InvalidateSessionResponse,
    responseDeserialize: deserialize_api_InvalidateSessionResponse,
  },
  // API key management
createApiKey: {
    path: '/api.SessionsService/CreateApiKey',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.CreateApiKeyRequest,
    responseType: proto_service_pb.ApiKey,
    requestSerialize: serialize_api_CreateApiKeyRequest,
    requestDeserialize: deserialize_api_CreateApiKeyRequest,
    responseSerialize: serialize_api_ApiKey,
    responseDeserialize: deserialize_api_ApiKey,
  },
  listApiKeys: {
    path: '/api.SessionsService/ListApiKeys',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.ListApiKeysRequest,
    responseType: proto_service_pb.ListApiKeysResponse,
    requestSerialize: serialize_api_ListApiKeysRequest,
    requestDeserialize: deserialize_api_ListApiKeysRequest,
    responseSerialize: serialize_api_ListApiKeysResponse,
    responseDeserialize: deserialize_api_ListApiKeysResponse,
  },
  revokeApiKey: {
    path: '/api.SessionsService/RevokeApiKey',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.RevokeApiKeyRequest,
    responseType: proto_service_pb.RevokeApiKeyResponse,
    requestSerialize: serialize_api_RevokeApiKeyRequest,
    requestDeserialize: deserialize_api_RevokeApiKeyRequest,
    responseSerialize: serialize_api_RevokeApiKeyResponse,
    responseDeserialize: deserialize_api_RevokeApiKeyResponse,
  },
};

exports.SessionsServiceClient = grpc.makeGenericClientConstructor(SessionsServiceService, 'SessionsService');
// ======================
// Machine Service
// ======================
var MachinesServiceService = exports.MachinesServiceService = {
  // Machine management
registerMachine: {
    path: '/api.MachinesService/RegisterMachine',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.RegisterMachineRequest,
    responseType: proto_service_pb.Machine,
    requestSerialize: serialize_api_RegisterMachineRequest,
    requestDeserialize: deserialize_api_RegisterMachineRequest,
    responseSerialize: serialize_api_Machine,
    responseDeserialize: deserialize_api_Machine,
  },
  listMachines: {
    path: '/api.MachinesService/ListMachines',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.ListMachinesRequest,
    responseType: proto_service_pb.ListMachinesResponse,
    requestSerialize: serialize_api_ListMachinesRequest,
    requestDeserialize: deserialize_api_ListMachinesRequest,
    responseSerialize: serialize_api_ListMachinesResponse,
    responseDeserialize: deserialize_api_ListMachinesResponse,
  },
  getMachine: {
    path: '/api.MachinesService/GetMachine',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.GetMachineRequest,
    responseType: proto_service_pb.Machine,
    requestSerialize: serialize_api_GetMachineRequest,
    requestDeserialize: deserialize_api_GetMachineRequest,
    responseSerialize: serialize_api_Machine,
    responseDeserialize: deserialize_api_Machine,
  },
  updateMachinePing: {
    path: '/api.MachinesService/UpdateMachinePing',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.UpdateMachinePingRequest,
    responseType: proto_service_pb.Machine,
    requestSerialize: serialize_api_UpdateMachinePingRequest,
    requestDeserialize: deserialize_api_UpdateMachinePingRequest,
    responseSerialize: serialize_api_Machine,
    responseDeserialize: deserialize_api_Machine,
  },
  unregisterMachine: {
    path: '/api.MachinesService/UnregisterMachine',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.UnregisterMachineRequest,
    responseType: proto_service_pb.UnregisterMachineResponse,
    requestSerialize: serialize_api_UnregisterMachineRequest,
    requestDeserialize: deserialize_api_UnregisterMachineRequest,
    responseSerialize: serialize_api_UnregisterMachineResponse,
    responseDeserialize: deserialize_api_UnregisterMachineResponse,
  },
  // Drain a machine: finish in-flight work then unregister
drainMachine: {
    path: '/api.MachinesService/DrainMachine',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.DrainMachineRequest,
    responseType: proto_service_pb.DrainMachineResponse,
    requestSerialize: serialize_api_DrainMachineRequest,
    requestDeserialize: deserialize_api_DrainMachineRequest,
    responseSerialize: serialize_api_DrainMachineResponse,
    responseDeserialize: deserialize_api_DrainMachineResponse,
  },
};

exports.MachinesServiceClient = grpc.makeGenericClientConstructor(MachinesServiceService, 'MachinesService');
// ======================
// Request Service
// ======================
var RequestsServiceService = exports.RequestsServiceService = {
  // Request management
createRequest: {
    path: '/api.RequestsService/CreateRequest',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.CreateRequestRequest,
    responseType: proto_service_pb.Request,
    requestSerialize: serialize_api_CreateRequestRequest,
    requestDeserialize: deserialize_api_CreateRequestRequest,
    responseSerialize: serialize_api_Request,
    responseDeserialize: deserialize_api_Request,
  },
  getRequest: {
    path: '/api.RequestsService/GetRequest',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.GetRequestRequest,
    responseType: proto_service_pb.Request,
    requestSerialize: serialize_api_GetRequestRequest,
    requestDeserialize: deserialize_api_GetRequestRequest,
    responseSerialize: serialize_api_Request,
    responseDeserialize: deserialize_api_Request,
  },
  listRequests: {
    path: '/api.RequestsService/ListRequests',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.ListRequestsRequest,
    responseType: proto_service_pb.ListRequestsResponse,
    requestSerialize: serialize_api_ListRequestsRequest,
    requestDeserialize: deserialize_api_ListRequestsRequest,
    responseSerialize: serialize_api_ListRequestsResponse,
    responseDeserialize: deserialize_api_ListRequestsResponse,
  },
  updateRequest: {
    path: '/api.RequestsService/UpdateRequest',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.UpdateRequestRequest,
    responseType: proto_service_pb.Request,
    requestSerialize: serialize_api_UpdateRequestRequest,
    requestDeserialize: deserialize_api_UpdateRequestRequest,
    responseSerialize: serialize_api_Request,
    responseDeserialize: deserialize_api_Request,
  },
  claimRequest: {
    path: '/api.RequestsService/ClaimRequest',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.ClaimRequestRequest,
    responseType: proto_service_pb.Request,
    requestSerialize: serialize_api_ClaimRequestRequest,
    requestDeserialize: deserialize_api_ClaimRequestRequest,
    responseSerialize: serialize_api_Request,
    responseDeserialize: deserialize_api_Request,
  },
  cancelRequest: {
    path: '/api.RequestsService/CancelRequest',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.CancelRequestRequest,
    responseType: proto_service_pb.CancelRequestResponse,
    requestSerialize: serialize_api_CancelRequestRequest,
    requestDeserialize: deserialize_api_CancelRequestRequest,
    responseSerialize: serialize_api_CancelRequestResponse,
    responseDeserialize: deserialize_api_CancelRequestResponse,
  },
  submitRequestResult: {
    path: '/api.RequestsService/SubmitRequestResult',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.SubmitRequestResultRequest,
    responseType: proto_service_pb.SubmitRequestResultResponse,
    requestSerialize: serialize_api_SubmitRequestResultRequest,
    requestDeserialize: deserialize_api_SubmitRequestResultRequest,
    responseSerialize: serialize_api_SubmitRequestResultResponse,
    responseDeserialize: deserialize_api_SubmitRequestResultResponse,
  },
  appendRequestChunks: {
    path: '/api.RequestsService/AppendRequestChunks',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.AppendRequestChunksRequest,
    responseType: proto_service_pb.AppendRequestChunksResponse,
    requestSerialize: serialize_api_AppendRequestChunksRequest,
    requestDeserialize: deserialize_api_AppendRequestChunksRequest,
    responseSerialize: serialize_api_AppendRequestChunksResponse,
    responseDeserialize: deserialize_api_AppendRequestChunksResponse,
  },
  getRequestChunks: {
    path: '/api.RequestsService/GetRequestChunks',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.GetRequestChunksRequest,
    responseType: proto_service_pb.GetRequestChunksResponse,
    requestSerialize: serialize_api_GetRequestChunksRequest,
    requestDeserialize: deserialize_api_GetRequestChunksRequest,
    responseSerialize: serialize_api_GetRequestChunksResponse,
    responseDeserialize: deserialize_api_GetRequestChunksResponse,
  },
};

exports.RequestsServiceClient = grpc.makeGenericClientConstructor(RequestsServiceService, 'RequestsService');
// ======================
// Task Service (Tasks)
// ======================
var TasksServiceService = exports.TasksServiceService = {
  // Task management
createTask: {
    path: '/api.TasksService/CreateTask',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.CreateTaskRequest,
    responseType: proto_service_pb.Task,
    requestSerialize: serialize_api_CreateTaskRequest,
    requestDeserialize: deserialize_api_CreateTaskRequest,
    responseSerialize: serialize_api_Task,
    responseDeserialize: deserialize_api_Task,
  },
  getTask: {
    path: '/api.TasksService/GetTask',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.GetTaskRequest,
    responseType: proto_service_pb.Task,
    requestSerialize: serialize_api_GetTaskRequest,
    requestDeserialize: deserialize_api_GetTaskRequest,
    responseSerialize: serialize_api_Task,
    responseDeserialize: deserialize_api_Task,
  },
  listTasks: {
    path: '/api.TasksService/ListTasks',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.ListTasksRequest,
    responseType: proto_service_pb.ListTasksResponse,
    requestSerialize: serialize_api_ListTasksRequest,
    requestDeserialize: deserialize_api_ListTasksRequest,
    responseSerialize: serialize_api_ListTasksResponse,
    responseDeserialize: deserialize_api_ListTasksResponse,
  },
  cancelTask: {
    path: '/api.TasksService/CancelTask',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.CancelTaskRequest,
    responseType: proto_service_pb.CancelTaskResponse,
    requestSerialize: serialize_api_CancelTaskRequest,
    requestDeserialize: deserialize_api_CancelTaskRequest,
    responseSerialize: serialize_api_CancelTaskResponse,
    responseDeserialize: deserialize_api_CancelTaskResponse,
  },
};

exports.TasksServiceClient = grpc.makeGenericClientConstructor(TasksServiceService, 'TasksService');
