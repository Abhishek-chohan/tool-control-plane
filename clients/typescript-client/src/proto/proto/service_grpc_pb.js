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
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');

function serialize_api_v1_ApiKey(arg) {
  if (!(arg instanceof proto_service_pb.ApiKey)) {
    throw new Error('Expected argument of type api.v1.ApiKey');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ApiKey(buffer_arg) {
  return proto_service_pb.ApiKey.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_AppendRequestChunksRequest(arg) {
  if (!(arg instanceof proto_service_pb.AppendRequestChunksRequest)) {
    throw new Error('Expected argument of type api.v1.AppendRequestChunksRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_AppendRequestChunksRequest(buffer_arg) {
  return proto_service_pb.AppendRequestChunksRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_AppendRequestChunksResponse(arg) {
  if (!(arg instanceof proto_service_pb.AppendRequestChunksResponse)) {
    throw new Error('Expected argument of type api.v1.AppendRequestChunksResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_AppendRequestChunksResponse(buffer_arg) {
  return proto_service_pb.AppendRequestChunksResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_BulkDeleteSessionsRequest(arg) {
  if (!(arg instanceof proto_service_pb.BulkDeleteSessionsRequest)) {
    throw new Error('Expected argument of type api.v1.BulkDeleteSessionsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_BulkDeleteSessionsRequest(buffer_arg) {
  return proto_service_pb.BulkDeleteSessionsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_BulkDeleteSessionsResponse(arg) {
  if (!(arg instanceof proto_service_pb.BulkDeleteSessionsResponse)) {
    throw new Error('Expected argument of type api.v1.BulkDeleteSessionsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_BulkDeleteSessionsResponse(buffer_arg) {
  return proto_service_pb.BulkDeleteSessionsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_CancelRequestRequest(arg) {
  if (!(arg instanceof proto_service_pb.CancelRequestRequest)) {
    throw new Error('Expected argument of type api.v1.CancelRequestRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_CancelRequestRequest(buffer_arg) {
  return proto_service_pb.CancelRequestRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_CancelRequestResponse(arg) {
  if (!(arg instanceof proto_service_pb.CancelRequestResponse)) {
    throw new Error('Expected argument of type api.v1.CancelRequestResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_CancelRequestResponse(buffer_arg) {
  return proto_service_pb.CancelRequestResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_CancelTaskRequest(arg) {
  if (!(arg instanceof proto_service_pb.CancelTaskRequest)) {
    throw new Error('Expected argument of type api.v1.CancelTaskRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_CancelTaskRequest(buffer_arg) {
  return proto_service_pb.CancelTaskRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_CancelTaskResponse(arg) {
  if (!(arg instanceof proto_service_pb.CancelTaskResponse)) {
    throw new Error('Expected argument of type api.v1.CancelTaskResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_CancelTaskResponse(buffer_arg) {
  return proto_service_pb.CancelTaskResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_ClaimRequestRequest(arg) {
  if (!(arg instanceof proto_service_pb.ClaimRequestRequest)) {
    throw new Error('Expected argument of type api.v1.ClaimRequestRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ClaimRequestRequest(buffer_arg) {
  return proto_service_pb.ClaimRequestRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_CreateApiKeyRequest(arg) {
  if (!(arg instanceof proto_service_pb.CreateApiKeyRequest)) {
    throw new Error('Expected argument of type api.v1.CreateApiKeyRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_CreateApiKeyRequest(buffer_arg) {
  return proto_service_pb.CreateApiKeyRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_CreateRequestRequest(arg) {
  if (!(arg instanceof proto_service_pb.CreateRequestRequest)) {
    throw new Error('Expected argument of type api.v1.CreateRequestRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_CreateRequestRequest(buffer_arg) {
  return proto_service_pb.CreateRequestRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_CreateSessionRequest(arg) {
  if (!(arg instanceof proto_service_pb.CreateSessionRequest)) {
    throw new Error('Expected argument of type api.v1.CreateSessionRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_CreateSessionRequest(buffer_arg) {
  return proto_service_pb.CreateSessionRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_CreateSessionResponse(arg) {
  if (!(arg instanceof proto_service_pb.CreateSessionResponse)) {
    throw new Error('Expected argument of type api.v1.CreateSessionResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_CreateSessionResponse(buffer_arg) {
  return proto_service_pb.CreateSessionResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_CreateTaskRequest(arg) {
  if (!(arg instanceof proto_service_pb.CreateTaskRequest)) {
    throw new Error('Expected argument of type api.v1.CreateTaskRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_CreateTaskRequest(buffer_arg) {
  return proto_service_pb.CreateTaskRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_DeleteSessionRequest(arg) {
  if (!(arg instanceof proto_service_pb.DeleteSessionRequest)) {
    throw new Error('Expected argument of type api.v1.DeleteSessionRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_DeleteSessionRequest(buffer_arg) {
  return proto_service_pb.DeleteSessionRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_DeleteSessionResponse(arg) {
  if (!(arg instanceof proto_service_pb.DeleteSessionResponse)) {
    throw new Error('Expected argument of type api.v1.DeleteSessionResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_DeleteSessionResponse(buffer_arg) {
  return proto_service_pb.DeleteSessionResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_DeleteToolRequest(arg) {
  if (!(arg instanceof proto_service_pb.DeleteToolRequest)) {
    throw new Error('Expected argument of type api.v1.DeleteToolRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_DeleteToolRequest(buffer_arg) {
  return proto_service_pb.DeleteToolRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_DeleteToolResponse(arg) {
  if (!(arg instanceof proto_service_pb.DeleteToolResponse)) {
    throw new Error('Expected argument of type api.v1.DeleteToolResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_DeleteToolResponse(buffer_arg) {
  return proto_service_pb.DeleteToolResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_DrainMachineRequest(arg) {
  if (!(arg instanceof proto_service_pb.DrainMachineRequest)) {
    throw new Error('Expected argument of type api.v1.DrainMachineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_DrainMachineRequest(buffer_arg) {
  return proto_service_pb.DrainMachineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_DrainMachineResponse(arg) {
  if (!(arg instanceof proto_service_pb.DrainMachineResponse)) {
    throw new Error('Expected argument of type api.v1.DrainMachineResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_DrainMachineResponse(buffer_arg) {
  return proto_service_pb.DrainMachineResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_ExecuteToolChunk(arg) {
  if (!(arg instanceof proto_service_pb.ExecuteToolChunk)) {
    throw new Error('Expected argument of type api.v1.ExecuteToolChunk');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ExecuteToolChunk(buffer_arg) {
  return proto_service_pb.ExecuteToolChunk.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_ExecuteToolRequest(arg) {
  if (!(arg instanceof proto_service_pb.ExecuteToolRequest)) {
    throw new Error('Expected argument of type api.v1.ExecuteToolRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ExecuteToolRequest(buffer_arg) {
  return proto_service_pb.ExecuteToolRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_ExecuteToolResponse(arg) {
  if (!(arg instanceof proto_service_pb.ExecuteToolResponse)) {
    throw new Error('Expected argument of type api.v1.ExecuteToolResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ExecuteToolResponse(buffer_arg) {
  return proto_service_pb.ExecuteToolResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_GetMachineRequest(arg) {
  if (!(arg instanceof proto_service_pb.GetMachineRequest)) {
    throw new Error('Expected argument of type api.v1.GetMachineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_GetMachineRequest(buffer_arg) {
  return proto_service_pb.GetMachineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_GetRequestChunksRequest(arg) {
  if (!(arg instanceof proto_service_pb.GetRequestChunksRequest)) {
    throw new Error('Expected argument of type api.v1.GetRequestChunksRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_GetRequestChunksRequest(buffer_arg) {
  return proto_service_pb.GetRequestChunksRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_GetRequestChunksResponse(arg) {
  if (!(arg instanceof proto_service_pb.GetRequestChunksResponse)) {
    throw new Error('Expected argument of type api.v1.GetRequestChunksResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_GetRequestChunksResponse(buffer_arg) {
  return proto_service_pb.GetRequestChunksResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_GetRequestRequest(arg) {
  if (!(arg instanceof proto_service_pb.GetRequestRequest)) {
    throw new Error('Expected argument of type api.v1.GetRequestRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_GetRequestRequest(buffer_arg) {
  return proto_service_pb.GetRequestRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_GetSessionRequest(arg) {
  if (!(arg instanceof proto_service_pb.GetSessionRequest)) {
    throw new Error('Expected argument of type api.v1.GetSessionRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_GetSessionRequest(buffer_arg) {
  return proto_service_pb.GetSessionRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_GetSessionStatsRequest(arg) {
  if (!(arg instanceof proto_service_pb.GetSessionStatsRequest)) {
    throw new Error('Expected argument of type api.v1.GetSessionStatsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_GetSessionStatsRequest(buffer_arg) {
  return proto_service_pb.GetSessionStatsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_GetSessionStatsResponse(arg) {
  if (!(arg instanceof proto_service_pb.GetSessionStatsResponse)) {
    throw new Error('Expected argument of type api.v1.GetSessionStatsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_GetSessionStatsResponse(buffer_arg) {
  return proto_service_pb.GetSessionStatsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_GetTaskRequest(arg) {
  if (!(arg instanceof proto_service_pb.GetTaskRequest)) {
    throw new Error('Expected argument of type api.v1.GetTaskRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_GetTaskRequest(buffer_arg) {
  return proto_service_pb.GetTaskRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_GetToolByIdRequest(arg) {
  if (!(arg instanceof proto_service_pb.GetToolByIdRequest)) {
    throw new Error('Expected argument of type api.v1.GetToolByIdRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_GetToolByIdRequest(buffer_arg) {
  return proto_service_pb.GetToolByIdRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_GetToolByNameRequest(arg) {
  if (!(arg instanceof proto_service_pb.GetToolByNameRequest)) {
    throw new Error('Expected argument of type api.v1.GetToolByNameRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_GetToolByNameRequest(buffer_arg) {
  return proto_service_pb.GetToolByNameRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_GetToolRequest(arg) {
  if (!(arg instanceof proto_service_pb.GetToolRequest)) {
    throw new Error('Expected argument of type api.v1.GetToolRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_GetToolRequest(buffer_arg) {
  return proto_service_pb.GetToolRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_GetToolResponse(arg) {
  if (!(arg instanceof proto_service_pb.GetToolResponse)) {
    throw new Error('Expected argument of type api.v1.GetToolResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_GetToolResponse(buffer_arg) {
  return proto_service_pb.GetToolResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_HealthCheckRequest(arg) {
  if (!(arg instanceof proto_service_pb.HealthCheckRequest)) {
    throw new Error('Expected argument of type api.v1.HealthCheckRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_HealthCheckRequest(buffer_arg) {
  return proto_service_pb.HealthCheckRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_HealthCheckResponse(arg) {
  if (!(arg instanceof proto_service_pb.HealthCheckResponse)) {
    throw new Error('Expected argument of type api.v1.HealthCheckResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_HealthCheckResponse(buffer_arg) {
  return proto_service_pb.HealthCheckResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_InvalidateSessionRequest(arg) {
  if (!(arg instanceof proto_service_pb.InvalidateSessionRequest)) {
    throw new Error('Expected argument of type api.v1.InvalidateSessionRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_InvalidateSessionRequest(buffer_arg) {
  return proto_service_pb.InvalidateSessionRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_InvalidateSessionResponse(arg) {
  if (!(arg instanceof proto_service_pb.InvalidateSessionResponse)) {
    throw new Error('Expected argument of type api.v1.InvalidateSessionResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_InvalidateSessionResponse(buffer_arg) {
  return proto_service_pb.InvalidateSessionResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_ListApiKeysRequest(arg) {
  if (!(arg instanceof proto_service_pb.ListApiKeysRequest)) {
    throw new Error('Expected argument of type api.v1.ListApiKeysRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ListApiKeysRequest(buffer_arg) {
  return proto_service_pb.ListApiKeysRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_ListApiKeysResponse(arg) {
  if (!(arg instanceof proto_service_pb.ListApiKeysResponse)) {
    throw new Error('Expected argument of type api.v1.ListApiKeysResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ListApiKeysResponse(buffer_arg) {
  return proto_service_pb.ListApiKeysResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_ListMachinesRequest(arg) {
  if (!(arg instanceof proto_service_pb.ListMachinesRequest)) {
    throw new Error('Expected argument of type api.v1.ListMachinesRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ListMachinesRequest(buffer_arg) {
  return proto_service_pb.ListMachinesRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_ListMachinesResponse(arg) {
  if (!(arg instanceof proto_service_pb.ListMachinesResponse)) {
    throw new Error('Expected argument of type api.v1.ListMachinesResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ListMachinesResponse(buffer_arg) {
  return proto_service_pb.ListMachinesResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_ListRequestsRequest(arg) {
  if (!(arg instanceof proto_service_pb.ListRequestsRequest)) {
    throw new Error('Expected argument of type api.v1.ListRequestsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ListRequestsRequest(buffer_arg) {
  return proto_service_pb.ListRequestsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_ListRequestsResponse(arg) {
  if (!(arg instanceof proto_service_pb.ListRequestsResponse)) {
    throw new Error('Expected argument of type api.v1.ListRequestsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ListRequestsResponse(buffer_arg) {
  return proto_service_pb.ListRequestsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_ListSessionsRequest(arg) {
  if (!(arg instanceof proto_service_pb.ListSessionsRequest)) {
    throw new Error('Expected argument of type api.v1.ListSessionsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ListSessionsRequest(buffer_arg) {
  return proto_service_pb.ListSessionsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_ListSessionsResponse(arg) {
  if (!(arg instanceof proto_service_pb.ListSessionsResponse)) {
    throw new Error('Expected argument of type api.v1.ListSessionsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ListSessionsResponse(buffer_arg) {
  return proto_service_pb.ListSessionsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_ListTasksRequest(arg) {
  if (!(arg instanceof proto_service_pb.ListTasksRequest)) {
    throw new Error('Expected argument of type api.v1.ListTasksRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ListTasksRequest(buffer_arg) {
  return proto_service_pb.ListTasksRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_ListTasksResponse(arg) {
  if (!(arg instanceof proto_service_pb.ListTasksResponse)) {
    throw new Error('Expected argument of type api.v1.ListTasksResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ListTasksResponse(buffer_arg) {
  return proto_service_pb.ListTasksResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_ListToolsRequest(arg) {
  if (!(arg instanceof proto_service_pb.ListToolsRequest)) {
    throw new Error('Expected argument of type api.v1.ListToolsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ListToolsRequest(buffer_arg) {
  return proto_service_pb.ListToolsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_ListToolsResponse(arg) {
  if (!(arg instanceof proto_service_pb.ListToolsResponse)) {
    throw new Error('Expected argument of type api.v1.ListToolsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ListToolsResponse(buffer_arg) {
  return proto_service_pb.ListToolsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_ListUserSessionsRequest(arg) {
  if (!(arg instanceof proto_service_pb.ListUserSessionsRequest)) {
    throw new Error('Expected argument of type api.v1.ListUserSessionsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ListUserSessionsRequest(buffer_arg) {
  return proto_service_pb.ListUserSessionsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_ListUserSessionsResponse(arg) {
  if (!(arg instanceof proto_service_pb.ListUserSessionsResponse)) {
    throw new Error('Expected argument of type api.v1.ListUserSessionsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ListUserSessionsResponse(buffer_arg) {
  return proto_service_pb.ListUserSessionsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_Machine(arg) {
  if (!(arg instanceof proto_service_pb.Machine)) {
    throw new Error('Expected argument of type api.v1.Machine');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_Machine(buffer_arg) {
  return proto_service_pb.Machine.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_RegisterMachineRequest(arg) {
  if (!(arg instanceof proto_service_pb.RegisterMachineRequest)) {
    throw new Error('Expected argument of type api.v1.RegisterMachineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_RegisterMachineRequest(buffer_arg) {
  return proto_service_pb.RegisterMachineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_RegisterToolRequest(arg) {
  if (!(arg instanceof proto_service_pb.RegisterToolRequest)) {
    throw new Error('Expected argument of type api.v1.RegisterToolRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_RegisterToolRequest(buffer_arg) {
  return proto_service_pb.RegisterToolRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_RegisterToolResponse(arg) {
  if (!(arg instanceof proto_service_pb.RegisterToolResponse)) {
    throw new Error('Expected argument of type api.v1.RegisterToolResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_RegisterToolResponse(buffer_arg) {
  return proto_service_pb.RegisterToolResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_RenewRequestLeaseRequest(arg) {
  if (!(arg instanceof proto_service_pb.RenewRequestLeaseRequest)) {
    throw new Error('Expected argument of type api.v1.RenewRequestLeaseRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_RenewRequestLeaseRequest(buffer_arg) {
  return proto_service_pb.RenewRequestLeaseRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_Request(arg) {
  if (!(arg instanceof proto_service_pb.Request)) {
    throw new Error('Expected argument of type api.v1.Request');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_Request(buffer_arg) {
  return proto_service_pb.Request.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_ResumeStreamRequest(arg) {
  if (!(arg instanceof proto_service_pb.ResumeStreamRequest)) {
    throw new Error('Expected argument of type api.v1.ResumeStreamRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_ResumeStreamRequest(buffer_arg) {
  return proto_service_pb.ResumeStreamRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_RevokeApiKeyRequest(arg) {
  if (!(arg instanceof proto_service_pb.RevokeApiKeyRequest)) {
    throw new Error('Expected argument of type api.v1.RevokeApiKeyRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_RevokeApiKeyRequest(buffer_arg) {
  return proto_service_pb.RevokeApiKeyRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_RevokeApiKeyResponse(arg) {
  if (!(arg instanceof proto_service_pb.RevokeApiKeyResponse)) {
    throw new Error('Expected argument of type api.v1.RevokeApiKeyResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_RevokeApiKeyResponse(buffer_arg) {
  return proto_service_pb.RevokeApiKeyResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_Session(arg) {
  if (!(arg instanceof proto_service_pb.Session)) {
    throw new Error('Expected argument of type api.v1.Session');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_Session(buffer_arg) {
  return proto_service_pb.Session.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_SubmitRequestResultRequest(arg) {
  if (!(arg instanceof proto_service_pb.SubmitRequestResultRequest)) {
    throw new Error('Expected argument of type api.v1.SubmitRequestResultRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_SubmitRequestResultRequest(buffer_arg) {
  return proto_service_pb.SubmitRequestResultRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_SubmitRequestResultResponse(arg) {
  if (!(arg instanceof proto_service_pb.SubmitRequestResultResponse)) {
    throw new Error('Expected argument of type api.v1.SubmitRequestResultResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_SubmitRequestResultResponse(buffer_arg) {
  return proto_service_pb.SubmitRequestResultResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_Task(arg) {
  if (!(arg instanceof proto_service_pb.Task)) {
    throw new Error('Expected argument of type api.v1.Task');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_Task(buffer_arg) {
  return proto_service_pb.Task.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_Tool(arg) {
  if (!(arg instanceof proto_service_pb.Tool)) {
    throw new Error('Expected argument of type api.v1.Tool');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_Tool(buffer_arg) {
  return proto_service_pb.Tool.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_UnregisterMachineRequest(arg) {
  if (!(arg instanceof proto_service_pb.UnregisterMachineRequest)) {
    throw new Error('Expected argument of type api.v1.UnregisterMachineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_UnregisterMachineRequest(buffer_arg) {
  return proto_service_pb.UnregisterMachineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_UnregisterMachineResponse(arg) {
  if (!(arg instanceof proto_service_pb.UnregisterMachineResponse)) {
    throw new Error('Expected argument of type api.v1.UnregisterMachineResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_UnregisterMachineResponse(buffer_arg) {
  return proto_service_pb.UnregisterMachineResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_UpdateMachinePingRequest(arg) {
  if (!(arg instanceof proto_service_pb.UpdateMachinePingRequest)) {
    throw new Error('Expected argument of type api.v1.UpdateMachinePingRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_UpdateMachinePingRequest(buffer_arg) {
  return proto_service_pb.UpdateMachinePingRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_UpdateRequestRequest(arg) {
  if (!(arg instanceof proto_service_pb.UpdateRequestRequest)) {
    throw new Error('Expected argument of type api.v1.UpdateRequestRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_UpdateRequestRequest(buffer_arg) {
  return proto_service_pb.UpdateRequestRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_UpdateSessionRequest(arg) {
  if (!(arg instanceof proto_service_pb.UpdateSessionRequest)) {
    throw new Error('Expected argument of type api.v1.UpdateSessionRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_UpdateSessionRequest(buffer_arg) {
  return proto_service_pb.UpdateSessionRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_v1_UpdateToolPingRequest(arg) {
  if (!(arg instanceof proto_service_pb.UpdateToolPingRequest)) {
    throw new Error('Expected argument of type api.v1.UpdateToolPingRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_v1_UpdateToolPingRequest(buffer_arg) {
  return proto_service_pb.UpdateToolPingRequest.deserializeBinary(new Uint8Array(buffer_arg));
}


// ======================
// Tool Service
// ======================
var ToolServiceService = exports.ToolServiceService = {
  // Tool management
registerTool: {
    path: '/api.v1.ToolService/RegisterTool',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.RegisterToolRequest,
    responseType: proto_service_pb.RegisterToolResponse,
    requestSerialize: serialize_api_v1_RegisterToolRequest,
    requestDeserialize: deserialize_api_v1_RegisterToolRequest,
    responseSerialize: serialize_api_v1_RegisterToolResponse,
    responseDeserialize: deserialize_api_v1_RegisterToolResponse,
  },
  listTools: {
    path: '/api.v1.ToolService/ListTools',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.ListToolsRequest,
    responseType: proto_service_pb.ListToolsResponse,
    requestSerialize: serialize_api_v1_ListToolsRequest,
    requestDeserialize: deserialize_api_v1_ListToolsRequest,
    responseSerialize: serialize_api_v1_ListToolsResponse,
    responseDeserialize: deserialize_api_v1_ListToolsResponse,
  },
  // GetTool resolves a tool by ID or by name: exactly one of the reference
// fields should be set; the ID wins when both are provided.
getTool: {
    path: '/api.v1.ToolService/GetTool',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.GetToolRequest,
    responseType: proto_service_pb.GetToolResponse,
    requestSerialize: serialize_api_v1_GetToolRequest,
    requestDeserialize: deserialize_api_v1_GetToolRequest,
    responseSerialize: serialize_api_v1_GetToolResponse,
    responseDeserialize: deserialize_api_v1_GetToolResponse,
  },
  // Deprecated: use GetTool. Retained as a stable alias.
getToolById: {
    path: '/api.v1.ToolService/GetToolById',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.GetToolByIdRequest,
    responseType: proto_service_pb.GetToolResponse,
    requestSerialize: serialize_api_v1_GetToolByIdRequest,
    requestDeserialize: deserialize_api_v1_GetToolByIdRequest,
    responseSerialize: serialize_api_v1_GetToolResponse,
    responseDeserialize: deserialize_api_v1_GetToolResponse,
  },
  // Deprecated: use GetTool. Retained as a stable alias.
getToolByName: {
    path: '/api.v1.ToolService/GetToolByName',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.GetToolByNameRequest,
    responseType: proto_service_pb.GetToolResponse,
    requestSerialize: serialize_api_v1_GetToolByNameRequest,
    requestDeserialize: deserialize_api_v1_GetToolByNameRequest,
    responseSerialize: serialize_api_v1_GetToolResponse,
    responseDeserialize: deserialize_api_v1_GetToolResponse,
  },
  deleteTool: {
    path: '/api.v1.ToolService/DeleteTool',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.DeleteToolRequest,
    responseType: proto_service_pb.DeleteToolResponse,
    requestSerialize: serialize_api_v1_DeleteToolRequest,
    requestDeserialize: deserialize_api_v1_DeleteToolRequest,
    responseSerialize: serialize_api_v1_DeleteToolResponse,
    responseDeserialize: deserialize_api_v1_DeleteToolResponse,
  },
  updateToolPing: {
    path: '/api.v1.ToolService/UpdateToolPing',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.UpdateToolPingRequest,
    responseType: proto_service_pb.Tool,
    requestSerialize: serialize_api_v1_UpdateToolPingRequest,
    requestDeserialize: deserialize_api_v1_UpdateToolPingRequest,
    responseSerialize: serialize_api_v1_Tool,
    responseDeserialize: deserialize_api_v1_Tool,
  },
  // Consumer execution entrypoints create request-backed work on the control plane.
// Request state, claim, and result lifecycle ownership stays in RequestsService.
streamExecuteTool: {
    path: '/api.v1.ToolService/StreamExecuteTool',
    requestStream: false,
    responseStream: true,
    requestType: proto_service_pb.ExecuteToolRequest,
    responseType: proto_service_pb.ExecuteToolChunk,
    requestSerialize: serialize_api_v1_ExecuteToolRequest,
    requestDeserialize: deserialize_api_v1_ExecuteToolRequest,
    responseSerialize: serialize_api_v1_ExecuteToolChunk,
    responseDeserialize: deserialize_api_v1_ExecuteToolChunk,
  },
  // Resume a broken stream from the last acknowledged sequence number.
// The server retains a bounded chunk window; if the requested sequence has
// fallen out of that window, this RPC fails with OUT_OF_RANGE.
resumeStream: {
    path: '/api.v1.ToolService/ResumeStream',
    requestStream: false,
    responseStream: true,
    requestType: proto_service_pb.ResumeStreamRequest,
    responseType: proto_service_pb.ExecuteToolChunk,
    requestSerialize: serialize_api_v1_ResumeStreamRequest,
    requestDeserialize: deserialize_api_v1_ResumeStreamRequest,
    responseSerialize: serialize_api_v1_ExecuteToolChunk,
    responseDeserialize: deserialize_api_v1_ExecuteToolChunk,
  },
  // InvokeTool is the v1 name for synchronous tool invocation.
invokeTool: {
    path: '/api.v1.ToolService/InvokeTool',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.ExecuteToolRequest,
    responseType: proto_service_pb.ExecuteToolResponse,
    requestSerialize: serialize_api_v1_ExecuteToolRequest,
    requestDeserialize: deserialize_api_v1_ExecuteToolRequest,
    responseSerialize: serialize_api_v1_ExecuteToolResponse,
    responseDeserialize: deserialize_api_v1_ExecuteToolResponse,
  },
  // Deprecated: use InvokeTool. Retained as a stable alias.
executeTool: {
    path: '/api.v1.ToolService/ExecuteTool',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.ExecuteToolRequest,
    responseType: proto_service_pb.ExecuteToolResponse,
    requestSerialize: serialize_api_v1_ExecuteToolRequest,
    requestDeserialize: deserialize_api_v1_ExecuteToolRequest,
    responseSerialize: serialize_api_v1_ExecuteToolResponse,
    responseDeserialize: deserialize_api_v1_ExecuteToolResponse,
  },
  // Health check (can be in any service or its own)
healthCheck: {
    path: '/api.v1.ToolService/HealthCheck',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.HealthCheckRequest,
    responseType: proto_service_pb.HealthCheckResponse,
    requestSerialize: serialize_api_v1_HealthCheckRequest,
    requestDeserialize: deserialize_api_v1_HealthCheckRequest,
    responseSerialize: serialize_api_v1_HealthCheckResponse,
    responseDeserialize: deserialize_api_v1_HealthCheckResponse,
  },
};

exports.ToolServiceClient = grpc.makeGenericClientConstructor(ToolServiceService, 'ToolService');
// ======================
// Session Service
// ======================
var SessionsServiceService = exports.SessionsServiceService = {
  // Session management
createSession: {
    path: '/api.v1.SessionsService/CreateSession',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.CreateSessionRequest,
    responseType: proto_service_pb.CreateSessionResponse,
    requestSerialize: serialize_api_v1_CreateSessionRequest,
    requestDeserialize: deserialize_api_v1_CreateSessionRequest,
    responseSerialize: serialize_api_v1_CreateSessionResponse,
    responseDeserialize: deserialize_api_v1_CreateSessionResponse,
  },
  getSession: {
    path: '/api.v1.SessionsService/GetSession',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.GetSessionRequest,
    responseType: proto_service_pb.Session,
    requestSerialize: serialize_api_v1_GetSessionRequest,
    requestDeserialize: deserialize_api_v1_GetSessionRequest,
    responseSerialize: serialize_api_v1_Session,
    responseDeserialize: deserialize_api_v1_Session,
  },
  listSessions: {
    path: '/api.v1.SessionsService/ListSessions',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.ListSessionsRequest,
    responseType: proto_service_pb.ListSessionsResponse,
    requestSerialize: serialize_api_v1_ListSessionsRequest,
    requestDeserialize: deserialize_api_v1_ListSessionsRequest,
    responseSerialize: serialize_api_v1_ListSessionsResponse,
    responseDeserialize: deserialize_api_v1_ListSessionsResponse,
  },
  updateSession: {
    path: '/api.v1.SessionsService/UpdateSession',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.UpdateSessionRequest,
    responseType: proto_service_pb.Session,
    requestSerialize: serialize_api_v1_UpdateSessionRequest,
    requestDeserialize: deserialize_api_v1_UpdateSessionRequest,
    responseSerialize: serialize_api_v1_Session,
    responseDeserialize: deserialize_api_v1_Session,
  },
  deleteSession: {
    path: '/api.v1.SessionsService/DeleteSession',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.DeleteSessionRequest,
    responseType: proto_service_pb.DeleteSessionResponse,
    requestSerialize: serialize_api_v1_DeleteSessionRequest,
    requestDeserialize: deserialize_api_v1_DeleteSessionRequest,
    responseSerialize: serialize_api_v1_DeleteSessionResponse,
    responseDeserialize: deserialize_api_v1_DeleteSessionResponse,
  },
  // Enhanced session management
listUserSessions: {
    path: '/api.v1.SessionsService/ListUserSessions',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.ListUserSessionsRequest,
    responseType: proto_service_pb.ListUserSessionsResponse,
    requestSerialize: serialize_api_v1_ListUserSessionsRequest,
    requestDeserialize: deserialize_api_v1_ListUserSessionsRequest,
    responseSerialize: serialize_api_v1_ListUserSessionsResponse,
    responseDeserialize: deserialize_api_v1_ListUserSessionsResponse,
  },
  bulkDeleteSessions: {
    path: '/api.v1.SessionsService/BulkDeleteSessions',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.BulkDeleteSessionsRequest,
    responseType: proto_service_pb.BulkDeleteSessionsResponse,
    requestSerialize: serialize_api_v1_BulkDeleteSessionsRequest,
    requestDeserialize: deserialize_api_v1_BulkDeleteSessionsRequest,
    responseSerialize: serialize_api_v1_BulkDeleteSessionsResponse,
    responseDeserialize: deserialize_api_v1_BulkDeleteSessionsResponse,
  },
  getSessionStats: {
    path: '/api.v1.SessionsService/GetSessionStats',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.GetSessionStatsRequest,
    responseType: proto_service_pb.GetSessionStatsResponse,
    requestSerialize: serialize_api_v1_GetSessionStatsRequest,
    requestDeserialize: deserialize_api_v1_GetSessionStatsRequest,
    responseSerialize: serialize_api_v1_GetSessionStatsResponse,
    responseDeserialize: deserialize_api_v1_GetSessionStatsResponse,
  },
  // InvalidateSession is the session-wide kill switch: it revokes every live
// API key for the session so no credential authenticates again. Use it for
// suspected key compromise; the session record itself is kept (use
// DeleteSession to remove it).
invalidateSession: {
    path: '/api.v1.SessionsService/InvalidateSession',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.InvalidateSessionRequest,
    responseType: proto_service_pb.InvalidateSessionResponse,
    requestSerialize: serialize_api_v1_InvalidateSessionRequest,
    requestDeserialize: deserialize_api_v1_InvalidateSessionRequest,
    responseSerialize: serialize_api_v1_InvalidateSessionResponse,
    responseDeserialize: deserialize_api_v1_InvalidateSessionResponse,
  },
  // API key management
createApiKey: {
    path: '/api.v1.SessionsService/CreateApiKey',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.CreateApiKeyRequest,
    responseType: proto_service_pb.ApiKey,
    requestSerialize: serialize_api_v1_CreateApiKeyRequest,
    requestDeserialize: deserialize_api_v1_CreateApiKeyRequest,
    responseSerialize: serialize_api_v1_ApiKey,
    responseDeserialize: deserialize_api_v1_ApiKey,
  },
  listApiKeys: {
    path: '/api.v1.SessionsService/ListApiKeys',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.ListApiKeysRequest,
    responseType: proto_service_pb.ListApiKeysResponse,
    requestSerialize: serialize_api_v1_ListApiKeysRequest,
    requestDeserialize: deserialize_api_v1_ListApiKeysRequest,
    responseSerialize: serialize_api_v1_ListApiKeysResponse,
    responseDeserialize: deserialize_api_v1_ListApiKeysResponse,
  },
  revokeApiKey: {
    path: '/api.v1.SessionsService/RevokeApiKey',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.RevokeApiKeyRequest,
    responseType: proto_service_pb.RevokeApiKeyResponse,
    requestSerialize: serialize_api_v1_RevokeApiKeyRequest,
    requestDeserialize: deserialize_api_v1_RevokeApiKeyRequest,
    responseSerialize: serialize_api_v1_RevokeApiKeyResponse,
    responseDeserialize: deserialize_api_v1_RevokeApiKeyResponse,
  },
};

exports.SessionsServiceClient = grpc.makeGenericClientConstructor(SessionsServiceService, 'SessionsService');
// ======================
// Machine Service
// ======================
var MachinesServiceService = exports.MachinesServiceService = {
  // Machine management
registerMachine: {
    path: '/api.v1.MachinesService/RegisterMachine',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.RegisterMachineRequest,
    responseType: proto_service_pb.Machine,
    requestSerialize: serialize_api_v1_RegisterMachineRequest,
    requestDeserialize: deserialize_api_v1_RegisterMachineRequest,
    responseSerialize: serialize_api_v1_Machine,
    responseDeserialize: deserialize_api_v1_Machine,
  },
  listMachines: {
    path: '/api.v1.MachinesService/ListMachines',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.ListMachinesRequest,
    responseType: proto_service_pb.ListMachinesResponse,
    requestSerialize: serialize_api_v1_ListMachinesRequest,
    requestDeserialize: deserialize_api_v1_ListMachinesRequest,
    responseSerialize: serialize_api_v1_ListMachinesResponse,
    responseDeserialize: deserialize_api_v1_ListMachinesResponse,
  },
  getMachine: {
    path: '/api.v1.MachinesService/GetMachine',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.GetMachineRequest,
    responseType: proto_service_pb.Machine,
    requestSerialize: serialize_api_v1_GetMachineRequest,
    requestDeserialize: deserialize_api_v1_GetMachineRequest,
    responseSerialize: serialize_api_v1_Machine,
    responseDeserialize: deserialize_api_v1_Machine,
  },
  updateMachinePing: {
    path: '/api.v1.MachinesService/UpdateMachinePing',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.UpdateMachinePingRequest,
    responseType: proto_service_pb.Machine,
    requestSerialize: serialize_api_v1_UpdateMachinePingRequest,
    requestDeserialize: deserialize_api_v1_UpdateMachinePingRequest,
    responseSerialize: serialize_api_v1_Machine,
    responseDeserialize: deserialize_api_v1_Machine,
  },
  unregisterMachine: {
    path: '/api.v1.MachinesService/UnregisterMachine',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.UnregisterMachineRequest,
    responseType: proto_service_pb.UnregisterMachineResponse,
    requestSerialize: serialize_api_v1_UnregisterMachineRequest,
    requestDeserialize: deserialize_api_v1_UnregisterMachineRequest,
    responseSerialize: serialize_api_v1_UnregisterMachineResponse,
    responseDeserialize: deserialize_api_v1_UnregisterMachineResponse,
  },
  // Drain a machine: finish in-flight work then unregister
drainMachine: {
    path: '/api.v1.MachinesService/DrainMachine',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.DrainMachineRequest,
    responseType: proto_service_pb.DrainMachineResponse,
    requestSerialize: serialize_api_v1_DrainMachineRequest,
    requestDeserialize: deserialize_api_v1_DrainMachineRequest,
    responseSerialize: serialize_api_v1_DrainMachineResponse,
    responseDeserialize: deserialize_api_v1_DrainMachineResponse,
  },
};

exports.MachinesServiceClient = grpc.makeGenericClientConstructor(MachinesServiceService, 'MachinesService');
// ======================
// Request Service
// ======================
var RequestsServiceService = exports.RequestsServiceService = {
  // Request management
createRequest: {
    path: '/api.v1.RequestsService/CreateRequest',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.CreateRequestRequest,
    responseType: proto_service_pb.Request,
    requestSerialize: serialize_api_v1_CreateRequestRequest,
    requestDeserialize: deserialize_api_v1_CreateRequestRequest,
    responseSerialize: serialize_api_v1_Request,
    responseDeserialize: deserialize_api_v1_Request,
  },
  getRequest: {
    path: '/api.v1.RequestsService/GetRequest',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.GetRequestRequest,
    responseType: proto_service_pb.Request,
    requestSerialize: serialize_api_v1_GetRequestRequest,
    requestDeserialize: deserialize_api_v1_GetRequestRequest,
    responseSerialize: serialize_api_v1_Request,
    responseDeserialize: deserialize_api_v1_Request,
  },
  listRequests: {
    path: '/api.v1.RequestsService/ListRequests',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.ListRequestsRequest,
    responseType: proto_service_pb.ListRequestsResponse,
    requestSerialize: serialize_api_v1_ListRequestsRequest,
    requestDeserialize: deserialize_api_v1_ListRequestsRequest,
    responseSerialize: serialize_api_v1_ListRequestsResponse,
    responseDeserialize: deserialize_api_v1_ListRequestsResponse,
  },
  updateRequest: {
    path: '/api.v1.RequestsService/UpdateRequest',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.UpdateRequestRequest,
    responseType: proto_service_pb.Request,
    requestSerialize: serialize_api_v1_UpdateRequestRequest,
    requestDeserialize: deserialize_api_v1_UpdateRequestRequest,
    responseSerialize: serialize_api_v1_Request,
    responseDeserialize: deserialize_api_v1_Request,
  },
  claimRequest: {
    path: '/api.v1.RequestsService/ClaimRequest',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.ClaimRequestRequest,
    responseType: proto_service_pb.Request,
    requestSerialize: serialize_api_v1_ClaimRequestRequest,
    requestDeserialize: deserialize_api_v1_ClaimRequestRequest,
    responseSerialize: serialize_api_v1_Request,
    responseDeserialize: deserialize_api_v1_Request,
  },
  cancelRequest: {
    path: '/api.v1.RequestsService/CancelRequest',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.CancelRequestRequest,
    responseType: proto_service_pb.CancelRequestResponse,
    requestSerialize: serialize_api_v1_CancelRequestRequest,
    requestDeserialize: deserialize_api_v1_CancelRequestRequest,
    responseSerialize: serialize_api_v1_CancelRequestResponse,
    responseDeserialize: deserialize_api_v1_CancelRequestResponse,
  },
  submitRequestResult: {
    path: '/api.v1.RequestsService/SubmitRequestResult',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.SubmitRequestResultRequest,
    responseType: proto_service_pb.SubmitRequestResultResponse,
    requestSerialize: serialize_api_v1_SubmitRequestResultRequest,
    requestDeserialize: deserialize_api_v1_SubmitRequestResultRequest,
    responseSerialize: serialize_api_v1_SubmitRequestResultResponse,
    responseDeserialize: deserialize_api_v1_SubmitRequestResultResponse,
  },
  appendRequestChunks: {
    path: '/api.v1.RequestsService/AppendRequestChunks',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.AppendRequestChunksRequest,
    responseType: proto_service_pb.AppendRequestChunksResponse,
    requestSerialize: serialize_api_v1_AppendRequestChunksRequest,
    requestDeserialize: deserialize_api_v1_AppendRequestChunksRequest,
    responseSerialize: serialize_api_v1_AppendRequestChunksResponse,
    responseDeserialize: deserialize_api_v1_AppendRequestChunksResponse,
  },
  getRequestChunks: {
    path: '/api.v1.RequestsService/GetRequestChunks',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.GetRequestChunksRequest,
    responseType: proto_service_pb.GetRequestChunksResponse,
    requestSerialize: serialize_api_v1_GetRequestChunksRequest,
    requestDeserialize: deserialize_api_v1_GetRequestChunksRequest,
    responseSerialize: serialize_api_v1_GetRequestChunksResponse,
    responseDeserialize: deserialize_api_v1_GetRequestChunksResponse,
  },
  // RenewRequestLease extends the execution lease of a claimed/running request
// so long-running tools are not reclaimed mid-flight. Only the current lease
// holder may renew: machine_id and lease_epoch must match the request's
// current lease grant. Renewal moves the lease deadline (visible_at) forward
// but never past the request's absolute timeout (leased_at + timeout_seconds).
// The server rejects renewals with FAILED_PRECONDITION when the lease is
// stale, reclaimed, or expired.
renewRequestLease: {
    path: '/api.v1.RequestsService/RenewRequestLease',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.RenewRequestLeaseRequest,
    responseType: proto_service_pb.Request,
    requestSerialize: serialize_api_v1_RenewRequestLeaseRequest,
    requestDeserialize: deserialize_api_v1_RenewRequestLeaseRequest,
    responseSerialize: serialize_api_v1_Request,
    responseDeserialize: deserialize_api_v1_Request,
  },
};

exports.RequestsServiceClient = grpc.makeGenericClientConstructor(RequestsServiceService, 'RequestsService');
// ======================
// Task Service (Tasks)
// ======================
var TasksServiceService = exports.TasksServiceService = {
  // Task management
createTask: {
    path: '/api.v1.TasksService/CreateTask',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.CreateTaskRequest,
    responseType: proto_service_pb.Task,
    requestSerialize: serialize_api_v1_CreateTaskRequest,
    requestDeserialize: deserialize_api_v1_CreateTaskRequest,
    responseSerialize: serialize_api_v1_Task,
    responseDeserialize: deserialize_api_v1_Task,
  },
  getTask: {
    path: '/api.v1.TasksService/GetTask',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.GetTaskRequest,
    responseType: proto_service_pb.Task,
    requestSerialize: serialize_api_v1_GetTaskRequest,
    requestDeserialize: deserialize_api_v1_GetTaskRequest,
    responseSerialize: serialize_api_v1_Task,
    responseDeserialize: deserialize_api_v1_Task,
  },
  listTasks: {
    path: '/api.v1.TasksService/ListTasks',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.ListTasksRequest,
    responseType: proto_service_pb.ListTasksResponse,
    requestSerialize: serialize_api_v1_ListTasksRequest,
    requestDeserialize: deserialize_api_v1_ListTasksRequest,
    responseSerialize: serialize_api_v1_ListTasksResponse,
    responseDeserialize: deserialize_api_v1_ListTasksResponse,
  },
  cancelTask: {
    path: '/api.v1.TasksService/CancelTask',
    requestStream: false,
    responseStream: false,
    requestType: proto_service_pb.CancelTaskRequest,
    responseType: proto_service_pb.CancelTaskResponse,
    requestSerialize: serialize_api_v1_CancelTaskRequest,
    requestDeserialize: deserialize_api_v1_CancelTaskRequest,
    responseSerialize: serialize_api_v1_CancelTaskResponse,
    responseDeserialize: deserialize_api_v1_CancelTaskResponse,
  },
};

exports.TasksServiceClient = grpc.makeGenericClientConstructor(TasksServiceService, 'TasksService');
