// source: proto/service.proto
/**
 * @fileoverview
 * @enhanceable
 * @suppress {missingRequire} reports error on implicit type usages.
 * @suppress {messageConventions} JS Compiler reports an error if a variable or
 *     field starts with 'MSG_' and isn't a translatable message.
 * @public
 */
// GENERATED CODE -- DO NOT EDIT!
/* eslint-disable */
// @ts-nocheck

var jspb = require('google-protobuf');
var goog = jspb;
var global = (function() {
  if (this) { return this; }
  if (typeof window !== 'undefined') { return window; }
  if (typeof global !== 'undefined') { return global; }
  if (typeof self !== 'undefined') { return self; }
  return Function('return this')();
}.call(null));

var google_api_annotations_pb = require('../google/api/annotations_pb.js');
goog.object.extend(proto, google_api_annotations_pb);
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');
goog.object.extend(proto, google_protobuf_timestamp_pb);
goog.exportSymbol('proto.api.v1.ApiKey', null, global);
goog.exportSymbol('proto.api.v1.AppendRequestChunksRequest', null, global);
goog.exportSymbol('proto.api.v1.AppendRequestChunksResponse', null, global);
goog.exportSymbol('proto.api.v1.BulkDeleteSessionsRequest', null, global);
goog.exportSymbol('proto.api.v1.BulkDeleteSessionsResponse', null, global);
goog.exportSymbol('proto.api.v1.CancelRequestRequest', null, global);
goog.exportSymbol('proto.api.v1.CancelRequestResponse', null, global);
goog.exportSymbol('proto.api.v1.CancelTaskRequest', null, global);
goog.exportSymbol('proto.api.v1.CancelTaskResponse', null, global);
goog.exportSymbol('proto.api.v1.ClaimRequestRequest', null, global);
goog.exportSymbol('proto.api.v1.CreateApiKeyRequest', null, global);
goog.exportSymbol('proto.api.v1.CreateRequestRequest', null, global);
goog.exportSymbol('proto.api.v1.CreateSessionRequest', null, global);
goog.exportSymbol('proto.api.v1.CreateSessionResponse', null, global);
goog.exportSymbol('proto.api.v1.CreateTaskRequest', null, global);
goog.exportSymbol('proto.api.v1.DeleteSessionRequest', null, global);
goog.exportSymbol('proto.api.v1.DeleteSessionResponse', null, global);
goog.exportSymbol('proto.api.v1.DeleteToolRequest', null, global);
goog.exportSymbol('proto.api.v1.DeleteToolResponse', null, global);
goog.exportSymbol('proto.api.v1.DrainMachineRequest', null, global);
goog.exportSymbol('proto.api.v1.DrainMachineResponse', null, global);
goog.exportSymbol('proto.api.v1.ExecuteToolChunk', null, global);
goog.exportSymbol('proto.api.v1.ExecuteToolRequest', null, global);
goog.exportSymbol('proto.api.v1.ExecuteToolResponse', null, global);
goog.exportSymbol('proto.api.v1.GetMachineRequest', null, global);
goog.exportSymbol('proto.api.v1.GetRequestChunksRequest', null, global);
goog.exportSymbol('proto.api.v1.GetRequestChunksResponse', null, global);
goog.exportSymbol('proto.api.v1.GetRequestRequest', null, global);
goog.exportSymbol('proto.api.v1.GetSessionRequest', null, global);
goog.exportSymbol('proto.api.v1.GetSessionStatsRequest', null, global);
goog.exportSymbol('proto.api.v1.GetSessionStatsResponse', null, global);
goog.exportSymbol('proto.api.v1.GetTaskRequest', null, global);
goog.exportSymbol('proto.api.v1.GetToolByIdRequest', null, global);
goog.exportSymbol('proto.api.v1.GetToolByNameRequest', null, global);
goog.exportSymbol('proto.api.v1.GetToolRequest', null, global);
goog.exportSymbol('proto.api.v1.GetToolResponse', null, global);
goog.exportSymbol('proto.api.v1.HealthCheckRequest', null, global);
goog.exportSymbol('proto.api.v1.HealthCheckResponse', null, global);
goog.exportSymbol('proto.api.v1.InvalidateSessionRequest', null, global);
goog.exportSymbol('proto.api.v1.InvalidateSessionResponse', null, global);
goog.exportSymbol('proto.api.v1.ListApiKeysRequest', null, global);
goog.exportSymbol('proto.api.v1.ListApiKeysResponse', null, global);
goog.exportSymbol('proto.api.v1.ListMachinesRequest', null, global);
goog.exportSymbol('proto.api.v1.ListMachinesResponse', null, global);
goog.exportSymbol('proto.api.v1.ListPage', null, global);
goog.exportSymbol('proto.api.v1.ListRequestsRequest', null, global);
goog.exportSymbol('proto.api.v1.ListRequestsResponse', null, global);
goog.exportSymbol('proto.api.v1.ListSessionsRequest', null, global);
goog.exportSymbol('proto.api.v1.ListSessionsResponse', null, global);
goog.exportSymbol('proto.api.v1.ListTasksRequest', null, global);
goog.exportSymbol('proto.api.v1.ListTasksResponse', null, global);
goog.exportSymbol('proto.api.v1.ListToolsRequest', null, global);
goog.exportSymbol('proto.api.v1.ListToolsResponse', null, global);
goog.exportSymbol('proto.api.v1.ListUserSessionsRequest', null, global);
goog.exportSymbol('proto.api.v1.ListUserSessionsResponse', null, global);
goog.exportSymbol('proto.api.v1.Machine', null, global);
goog.exportSymbol('proto.api.v1.RegisterMachineRequest', null, global);
goog.exportSymbol('proto.api.v1.RegisterToolRequest', null, global);
goog.exportSymbol('proto.api.v1.RegisterToolResponse', null, global);
goog.exportSymbol('proto.api.v1.RenewRequestLeaseRequest', null, global);
goog.exportSymbol('proto.api.v1.Request', null, global);
goog.exportSymbol('proto.api.v1.RequestStatus', null, global);
goog.exportSymbol('proto.api.v1.ResumeStreamRequest', null, global);
goog.exportSymbol('proto.api.v1.RevokeApiKeyRequest', null, global);
goog.exportSymbol('proto.api.v1.RevokeApiKeyResponse', null, global);
goog.exportSymbol('proto.api.v1.Session', null, global);
goog.exportSymbol('proto.api.v1.SubmitRequestResultRequest', null, global);
goog.exportSymbol('proto.api.v1.SubmitRequestResultResponse', null, global);
goog.exportSymbol('proto.api.v1.Task', null, global);
goog.exportSymbol('proto.api.v1.TaskStatus', null, global);
goog.exportSymbol('proto.api.v1.Tool', null, global);
goog.exportSymbol('proto.api.v1.UnregisterMachineRequest', null, global);
goog.exportSymbol('proto.api.v1.UnregisterMachineResponse', null, global);
goog.exportSymbol('proto.api.v1.UpdateMachinePingRequest', null, global);
goog.exportSymbol('proto.api.v1.UpdateRequestRequest', null, global);
goog.exportSymbol('proto.api.v1.UpdateSessionRequest', null, global);
goog.exportSymbol('proto.api.v1.UpdateToolPingRequest', null, global);
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ListPage = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.ListPage, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ListPage.displayName = 'proto.api.v1.ListPage';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ResumeStreamRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.ResumeStreamRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ResumeStreamRequest.displayName = 'proto.api.v1.ResumeStreamRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.DrainMachineRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.DrainMachineRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.DrainMachineRequest.displayName = 'proto.api.v1.DrainMachineRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.DrainMachineResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.DrainMachineResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.DrainMachineResponse.displayName = 'proto.api.v1.DrainMachineResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.Tool = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api.v1.Tool.repeatedFields_, null);
};
goog.inherits(proto.api.v1.Tool, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.Tool.displayName = 'proto.api.v1.Tool';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.Session = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.Session, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.Session.displayName = 'proto.api.v1.Session';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ApiKey = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api.v1.ApiKey.repeatedFields_, null);
};
goog.inherits(proto.api.v1.ApiKey, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ApiKey.displayName = 'proto.api.v1.ApiKey';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.Machine = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.Machine, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.Machine.displayName = 'proto.api.v1.Machine';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.Request = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.Request, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.Request.displayName = 'proto.api.v1.Request';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.RegisterToolRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api.v1.RegisterToolRequest.repeatedFields_, null);
};
goog.inherits(proto.api.v1.RegisterToolRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.RegisterToolRequest.displayName = 'proto.api.v1.RegisterToolRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.RegisterToolResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.RegisterToolResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.RegisterToolResponse.displayName = 'proto.api.v1.RegisterToolResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ListToolsRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.ListToolsRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ListToolsRequest.displayName = 'proto.api.v1.ListToolsRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ListToolsResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api.v1.ListToolsResponse.repeatedFields_, null);
};
goog.inherits(proto.api.v1.ListToolsResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ListToolsResponse.displayName = 'proto.api.v1.ListToolsResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.GetToolRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.GetToolRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.GetToolRequest.displayName = 'proto.api.v1.GetToolRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.GetToolByIdRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.GetToolByIdRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.GetToolByIdRequest.displayName = 'proto.api.v1.GetToolByIdRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.GetToolByNameRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.GetToolByNameRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.GetToolByNameRequest.displayName = 'proto.api.v1.GetToolByNameRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.GetToolResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.GetToolResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.GetToolResponse.displayName = 'proto.api.v1.GetToolResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.DeleteToolRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.DeleteToolRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.DeleteToolRequest.displayName = 'proto.api.v1.DeleteToolRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.DeleteToolResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.DeleteToolResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.DeleteToolResponse.displayName = 'proto.api.v1.DeleteToolResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.UpdateToolPingRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.UpdateToolPingRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.UpdateToolPingRequest.displayName = 'proto.api.v1.UpdateToolPingRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.CreateSessionRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.CreateSessionRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.CreateSessionRequest.displayName = 'proto.api.v1.CreateSessionRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.CreateSessionResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.CreateSessionResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.CreateSessionResponse.displayName = 'proto.api.v1.CreateSessionResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.GetSessionRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.GetSessionRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.GetSessionRequest.displayName = 'proto.api.v1.GetSessionRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ListSessionsRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.ListSessionsRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ListSessionsRequest.displayName = 'proto.api.v1.ListSessionsRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ListSessionsResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api.v1.ListSessionsResponse.repeatedFields_, null);
};
goog.inherits(proto.api.v1.ListSessionsResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ListSessionsResponse.displayName = 'proto.api.v1.ListSessionsResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ListUserSessionsRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.ListUserSessionsRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ListUserSessionsRequest.displayName = 'proto.api.v1.ListUserSessionsRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ListUserSessionsResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api.v1.ListUserSessionsResponse.repeatedFields_, null);
};
goog.inherits(proto.api.v1.ListUserSessionsResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ListUserSessionsResponse.displayName = 'proto.api.v1.ListUserSessionsResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.BulkDeleteSessionsRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api.v1.BulkDeleteSessionsRequest.repeatedFields_, null);
};
goog.inherits(proto.api.v1.BulkDeleteSessionsRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.BulkDeleteSessionsRequest.displayName = 'proto.api.v1.BulkDeleteSessionsRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.BulkDeleteSessionsResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api.v1.BulkDeleteSessionsResponse.repeatedFields_, null);
};
goog.inherits(proto.api.v1.BulkDeleteSessionsResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.BulkDeleteSessionsResponse.displayName = 'proto.api.v1.BulkDeleteSessionsResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.GetSessionStatsRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.GetSessionStatsRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.GetSessionStatsRequest.displayName = 'proto.api.v1.GetSessionStatsRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.GetSessionStatsResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.GetSessionStatsResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.GetSessionStatsResponse.displayName = 'proto.api.v1.GetSessionStatsResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.InvalidateSessionRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.InvalidateSessionRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.InvalidateSessionRequest.displayName = 'proto.api.v1.InvalidateSessionRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.InvalidateSessionResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.InvalidateSessionResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.InvalidateSessionResponse.displayName = 'proto.api.v1.InvalidateSessionResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.UpdateSessionRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.UpdateSessionRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.UpdateSessionRequest.displayName = 'proto.api.v1.UpdateSessionRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.DeleteSessionRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.DeleteSessionRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.DeleteSessionRequest.displayName = 'proto.api.v1.DeleteSessionRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.DeleteSessionResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.DeleteSessionResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.DeleteSessionResponse.displayName = 'proto.api.v1.DeleteSessionResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.CreateApiKeyRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api.v1.CreateApiKeyRequest.repeatedFields_, null);
};
goog.inherits(proto.api.v1.CreateApiKeyRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.CreateApiKeyRequest.displayName = 'proto.api.v1.CreateApiKeyRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ListApiKeysRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.ListApiKeysRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ListApiKeysRequest.displayName = 'proto.api.v1.ListApiKeysRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ListApiKeysResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api.v1.ListApiKeysResponse.repeatedFields_, null);
};
goog.inherits(proto.api.v1.ListApiKeysResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ListApiKeysResponse.displayName = 'proto.api.v1.ListApiKeysResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.RevokeApiKeyRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.RevokeApiKeyRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.RevokeApiKeyRequest.displayName = 'proto.api.v1.RevokeApiKeyRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.RevokeApiKeyResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.RevokeApiKeyResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.RevokeApiKeyResponse.displayName = 'proto.api.v1.RevokeApiKeyResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.RegisterMachineRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api.v1.RegisterMachineRequest.repeatedFields_, null);
};
goog.inherits(proto.api.v1.RegisterMachineRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.RegisterMachineRequest.displayName = 'proto.api.v1.RegisterMachineRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ListMachinesRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.ListMachinesRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ListMachinesRequest.displayName = 'proto.api.v1.ListMachinesRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ListMachinesResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api.v1.ListMachinesResponse.repeatedFields_, null);
};
goog.inherits(proto.api.v1.ListMachinesResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ListMachinesResponse.displayName = 'proto.api.v1.ListMachinesResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.GetMachineRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.GetMachineRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.GetMachineRequest.displayName = 'proto.api.v1.GetMachineRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.UpdateMachinePingRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.UpdateMachinePingRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.UpdateMachinePingRequest.displayName = 'proto.api.v1.UpdateMachinePingRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.UnregisterMachineRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.UnregisterMachineRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.UnregisterMachineRequest.displayName = 'proto.api.v1.UnregisterMachineRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.UnregisterMachineResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.UnregisterMachineResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.UnregisterMachineResponse.displayName = 'proto.api.v1.UnregisterMachineResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ExecuteToolRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.ExecuteToolRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ExecuteToolRequest.displayName = 'proto.api.v1.ExecuteToolRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ExecuteToolResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.ExecuteToolResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ExecuteToolResponse.displayName = 'proto.api.v1.ExecuteToolResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ExecuteToolChunk = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.ExecuteToolChunk, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ExecuteToolChunk.displayName = 'proto.api.v1.ExecuteToolChunk';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.CreateRequestRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.CreateRequestRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.CreateRequestRequest.displayName = 'proto.api.v1.CreateRequestRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.GetRequestRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.GetRequestRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.GetRequestRequest.displayName = 'proto.api.v1.GetRequestRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ListRequestsRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.ListRequestsRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ListRequestsRequest.displayName = 'proto.api.v1.ListRequestsRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ListRequestsResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api.v1.ListRequestsResponse.repeatedFields_, null);
};
goog.inherits(proto.api.v1.ListRequestsResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ListRequestsResponse.displayName = 'proto.api.v1.ListRequestsResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.UpdateRequestRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.UpdateRequestRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.UpdateRequestRequest.displayName = 'proto.api.v1.UpdateRequestRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ClaimRequestRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.ClaimRequestRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ClaimRequestRequest.displayName = 'proto.api.v1.ClaimRequestRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.CancelRequestRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.CancelRequestRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.CancelRequestRequest.displayName = 'proto.api.v1.CancelRequestRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.CancelRequestResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.CancelRequestResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.CancelRequestResponse.displayName = 'proto.api.v1.CancelRequestResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.SubmitRequestResultRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.SubmitRequestResultRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.SubmitRequestResultRequest.displayName = 'proto.api.v1.SubmitRequestResultRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.SubmitRequestResultResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.SubmitRequestResultResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.SubmitRequestResultResponse.displayName = 'proto.api.v1.SubmitRequestResultResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.AppendRequestChunksRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api.v1.AppendRequestChunksRequest.repeatedFields_, null);
};
goog.inherits(proto.api.v1.AppendRequestChunksRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.AppendRequestChunksRequest.displayName = 'proto.api.v1.AppendRequestChunksRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.AppendRequestChunksResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.AppendRequestChunksResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.AppendRequestChunksResponse.displayName = 'proto.api.v1.AppendRequestChunksResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.GetRequestChunksRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.GetRequestChunksRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.GetRequestChunksRequest.displayName = 'proto.api.v1.GetRequestChunksRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.GetRequestChunksResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api.v1.GetRequestChunksResponse.repeatedFields_, null);
};
goog.inherits(proto.api.v1.GetRequestChunksResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.GetRequestChunksResponse.displayName = 'proto.api.v1.GetRequestChunksResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.RenewRequestLeaseRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.RenewRequestLeaseRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.RenewRequestLeaseRequest.displayName = 'proto.api.v1.RenewRequestLeaseRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.HealthCheckRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.HealthCheckRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.HealthCheckRequest.displayName = 'proto.api.v1.HealthCheckRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.HealthCheckResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.HealthCheckResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.HealthCheckResponse.displayName = 'proto.api.v1.HealthCheckResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.Task = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.Task, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.Task.displayName = 'proto.api.v1.Task';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.CreateTaskRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.CreateTaskRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.CreateTaskRequest.displayName = 'proto.api.v1.CreateTaskRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.GetTaskRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.GetTaskRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.GetTaskRequest.displayName = 'proto.api.v1.GetTaskRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ListTasksRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.ListTasksRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ListTasksRequest.displayName = 'proto.api.v1.ListTasksRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.ListTasksResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api.v1.ListTasksResponse.repeatedFields_, null);
};
goog.inherits(proto.api.v1.ListTasksResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.ListTasksResponse.displayName = 'proto.api.v1.ListTasksResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.CancelTaskRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.CancelTaskRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.CancelTaskRequest.displayName = 'proto.api.v1.CancelTaskRequest';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api.v1.CancelTaskResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api.v1.CancelTaskResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api.v1.CancelTaskResponse.displayName = 'proto.api.v1.CancelTaskResponse';
}



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ListPage.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ListPage.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ListPage} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListPage.toObject = function(includeInstance, msg) {
  var f, obj = {
    nextPageToken: jspb.Message.getFieldWithDefault(msg, 1, ""),
    totalSize: jspb.Message.getFieldWithDefault(msg, 2, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ListPage}
 */
proto.api.v1.ListPage.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ListPage;
  return proto.api.v1.ListPage.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ListPage} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ListPage}
 */
proto.api.v1.ListPage.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setNextPageToken(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setTotalSize(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ListPage.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ListPage.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ListPage} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListPage.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getNextPageToken();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getTotalSize();
  if (f !== 0) {
    writer.writeInt32(
      2,
      f
    );
  }
};


/**
 * optional string next_page_token = 1;
 * @return {string}
 */
proto.api.v1.ListPage.prototype.getNextPageToken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ListPage} returns this
 */
proto.api.v1.ListPage.prototype.setNextPageToken = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional int32 total_size = 2;
 * @return {number}
 */
proto.api.v1.ListPage.prototype.getTotalSize = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.ListPage} returns this
 */
proto.api.v1.ListPage.prototype.setTotalSize = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ResumeStreamRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ResumeStreamRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ResumeStreamRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ResumeStreamRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    requestId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    lastSeq: jspb.Message.getFieldWithDefault(msg, 2, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ResumeStreamRequest}
 */
proto.api.v1.ResumeStreamRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ResumeStreamRequest;
  return proto.api.v1.ResumeStreamRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ResumeStreamRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ResumeStreamRequest}
 */
proto.api.v1.ResumeStreamRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setRequestId(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setLastSeq(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ResumeStreamRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ResumeStreamRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ResumeStreamRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ResumeStreamRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRequestId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getLastSeq();
  if (f !== 0) {
    writer.writeInt32(
      2,
      f
    );
  }
};


/**
 * optional string request_id = 1;
 * @return {string}
 */
proto.api.v1.ResumeStreamRequest.prototype.getRequestId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ResumeStreamRequest} returns this
 */
proto.api.v1.ResumeStreamRequest.prototype.setRequestId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional int32 last_seq = 2;
 * @return {number}
 */
proto.api.v1.ResumeStreamRequest.prototype.getLastSeq = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.ResumeStreamRequest} returns this
 */
proto.api.v1.ResumeStreamRequest.prototype.setLastSeq = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.DrainMachineRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.DrainMachineRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.DrainMachineRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.DrainMachineRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    machineId: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.DrainMachineRequest}
 */
proto.api.v1.DrainMachineRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.DrainMachineRequest;
  return proto.api.v1.DrainMachineRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.DrainMachineRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.DrainMachineRequest}
 */
proto.api.v1.DrainMachineRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setMachineId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.DrainMachineRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.DrainMachineRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.DrainMachineRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.DrainMachineRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getMachineId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.DrainMachineRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.DrainMachineRequest} returns this
 */
proto.api.v1.DrainMachineRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string machine_id = 2;
 * @return {string}
 */
proto.api.v1.DrainMachineRequest.prototype.getMachineId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.DrainMachineRequest} returns this
 */
proto.api.v1.DrainMachineRequest.prototype.setMachineId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.DrainMachineResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.DrainMachineResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.DrainMachineResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.DrainMachineResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    drained: jspb.Message.getBooleanFieldWithDefault(msg, 1, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.DrainMachineResponse}
 */
proto.api.v1.DrainMachineResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.DrainMachineResponse;
  return proto.api.v1.DrainMachineResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.DrainMachineResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.DrainMachineResponse}
 */
proto.api.v1.DrainMachineResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setDrained(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.DrainMachineResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.DrainMachineResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.DrainMachineResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.DrainMachineResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getDrained();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
};


/**
 * optional bool drained = 1;
 * @return {boolean}
 */
proto.api.v1.DrainMachineResponse.prototype.getDrained = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.api.v1.DrainMachineResponse} returns this
 */
proto.api.v1.DrainMachineResponse.prototype.setDrained = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api.v1.Tool.repeatedFields_ = [10];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.Tool.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.Tool.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.Tool} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.Tool.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, ""),
    name: jspb.Message.getFieldWithDefault(msg, 2, ""),
    description: jspb.Message.getFieldWithDefault(msg, 3, ""),
    schema: jspb.Message.getFieldWithDefault(msg, 4, ""),
    configMap: (f = msg.getConfigMap()) ? f.toObject(includeInstance, undefined) : [],
    createdAt: (f = msg.getCreatedAt()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    lastPingAt: (f = msg.getLastPingAt()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    sessionId: jspb.Message.getFieldWithDefault(msg, 8, ""),
    machineId: jspb.Message.getFieldWithDefault(msg, 9, ""),
    tagsList: (f = jspb.Message.getRepeatedField(msg, 10)) == null ? undefined : f
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.Tool}
 */
proto.api.v1.Tool.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.Tool;
  return proto.api.v1.Tool.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.Tool} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.Tool}
 */
proto.api.v1.Tool.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setSchema(value);
      break;
    case 5:
      var value = msg.getConfigMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString, null, "", "");
         });
      break;
    case 6:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setCreatedAt(value);
      break;
    case 7:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setLastPingAt(value);
      break;
    case 8:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 9:
      var value = /** @type {string} */ (reader.readString());
      msg.setMachineId(value);
      break;
    case 10:
      var value = /** @type {string} */ (reader.readString());
      msg.addTags(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.Tool.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.Tool.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.Tool} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.Tool.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getDescription();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getSchema();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getConfigMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(5, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
  f = message.getCreatedAt();
  if (f != null) {
    writer.writeMessage(
      6,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getLastPingAt();
  if (f != null) {
    writer.writeMessage(
      7,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      8,
      f
    );
  }
  f = message.getMachineId();
  if (f.length > 0) {
    writer.writeString(
      9,
      f
    );
  }
  f = message.getTagsList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      10,
      f
    );
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.api.v1.Tool.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Tool} returns this
 */
proto.api.v1.Tool.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string name = 2;
 * @return {string}
 */
proto.api.v1.Tool.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Tool} returns this
 */
proto.api.v1.Tool.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string description = 3;
 * @return {string}
 */
proto.api.v1.Tool.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Tool} returns this
 */
proto.api.v1.Tool.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string schema = 4;
 * @return {string}
 */
proto.api.v1.Tool.prototype.getSchema = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Tool} returns this
 */
proto.api.v1.Tool.prototype.setSchema = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * map<string, string> config = 5;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.api.v1.Tool.prototype.getConfigMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 5, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.api.v1.Tool} returns this
 */
proto.api.v1.Tool.prototype.clearConfigMap = function() {
  this.getConfigMap().clear();
  return this;};


/**
 * optional google.protobuf.Timestamp created_at = 6;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.api.v1.Tool.prototype.getCreatedAt = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 6));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.api.v1.Tool} returns this
*/
proto.api.v1.Tool.prototype.setCreatedAt = function(value) {
  return jspb.Message.setWrapperField(this, 6, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api.v1.Tool} returns this
 */
proto.api.v1.Tool.prototype.clearCreatedAt = function() {
  return this.setCreatedAt(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api.v1.Tool.prototype.hasCreatedAt = function() {
  return jspb.Message.getField(this, 6) != null;
};


/**
 * optional google.protobuf.Timestamp last_ping_at = 7;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.api.v1.Tool.prototype.getLastPingAt = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 7));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.api.v1.Tool} returns this
*/
proto.api.v1.Tool.prototype.setLastPingAt = function(value) {
  return jspb.Message.setWrapperField(this, 7, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api.v1.Tool} returns this
 */
proto.api.v1.Tool.prototype.clearLastPingAt = function() {
  return this.setLastPingAt(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api.v1.Tool.prototype.hasLastPingAt = function() {
  return jspb.Message.getField(this, 7) != null;
};


/**
 * optional string session_id = 8;
 * @return {string}
 */
proto.api.v1.Tool.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 8, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Tool} returns this
 */
proto.api.v1.Tool.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 8, value);
};


/**
 * optional string machine_id = 9;
 * @return {string}
 */
proto.api.v1.Tool.prototype.getMachineId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 9, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Tool} returns this
 */
proto.api.v1.Tool.prototype.setMachineId = function(value) {
  return jspb.Message.setProto3StringField(this, 9, value);
};


/**
 * repeated string tags = 10;
 * @return {!Array<string>}
 */
proto.api.v1.Tool.prototype.getTagsList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 10));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.api.v1.Tool} returns this
 */
proto.api.v1.Tool.prototype.setTagsList = function(value) {
  return jspb.Message.setField(this, 10, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.api.v1.Tool} returns this
 */
proto.api.v1.Tool.prototype.addTags = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 10, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api.v1.Tool} returns this
 */
proto.api.v1.Tool.prototype.clearTagsList = function() {
  return this.setTagsList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.Session.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.Session.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.Session} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.Session.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, ""),
    name: jspb.Message.getFieldWithDefault(msg, 2, ""),
    description: jspb.Message.getFieldWithDefault(msg, 3, ""),
    createdAt: (f = msg.getCreatedAt()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    createdBy: jspb.Message.getFieldWithDefault(msg, 5, ""),
    namespace: jspb.Message.getFieldWithDefault(msg, 6, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.Session}
 */
proto.api.v1.Session.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.Session;
  return proto.api.v1.Session.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.Session} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.Session}
 */
proto.api.v1.Session.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 4:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setCreatedAt(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setCreatedBy(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setNamespace(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.Session.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.Session.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.Session} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.Session.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getDescription();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getCreatedAt();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getCreatedBy();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getNamespace();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.api.v1.Session.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Session} returns this
 */
proto.api.v1.Session.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string name = 2;
 * @return {string}
 */
proto.api.v1.Session.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Session} returns this
 */
proto.api.v1.Session.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string description = 3;
 * @return {string}
 */
proto.api.v1.Session.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Session} returns this
 */
proto.api.v1.Session.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional google.protobuf.Timestamp created_at = 4;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.api.v1.Session.prototype.getCreatedAt = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 4));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.api.v1.Session} returns this
*/
proto.api.v1.Session.prototype.setCreatedAt = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api.v1.Session} returns this
 */
proto.api.v1.Session.prototype.clearCreatedAt = function() {
  return this.setCreatedAt(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api.v1.Session.prototype.hasCreatedAt = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * optional string created_by = 5;
 * @return {string}
 */
proto.api.v1.Session.prototype.getCreatedBy = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Session} returns this
 */
proto.api.v1.Session.prototype.setCreatedBy = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional string namespace = 6;
 * @return {string}
 */
proto.api.v1.Session.prototype.getNamespace = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Session} returns this
 */
proto.api.v1.Session.prototype.setNamespace = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api.v1.ApiKey.repeatedFields_ = [8];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ApiKey.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ApiKey.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ApiKey} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ApiKey.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, ""),
    name: jspb.Message.getFieldWithDefault(msg, 2, ""),
    key: jspb.Message.getFieldWithDefault(msg, 3, ""),
    sessionId: jspb.Message.getFieldWithDefault(msg, 4, ""),
    createdAt: (f = msg.getCreatedAt()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    createdBy: jspb.Message.getFieldWithDefault(msg, 6, ""),
    revokedAt: (f = msg.getRevokedAt()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    capabilitiesList: (f = jspb.Message.getRepeatedField(msg, 8)) == null ? undefined : f,
    keyPreview: jspb.Message.getFieldWithDefault(msg, 9, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ApiKey}
 */
proto.api.v1.ApiKey.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ApiKey;
  return proto.api.v1.ApiKey.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ApiKey} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ApiKey}
 */
proto.api.v1.ApiKey.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setKey(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 5:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setCreatedAt(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setCreatedBy(value);
      break;
    case 7:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setRevokedAt(value);
      break;
    case 8:
      var value = /** @type {string} */ (reader.readString());
      msg.addCapabilities(value);
      break;
    case 9:
      var value = /** @type {string} */ (reader.readString());
      msg.setKeyPreview(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ApiKey.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ApiKey.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ApiKey} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ApiKey.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getKey();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getCreatedAt();
  if (f != null) {
    writer.writeMessage(
      5,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getCreatedBy();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getRevokedAt();
  if (f != null) {
    writer.writeMessage(
      7,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getCapabilitiesList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      8,
      f
    );
  }
  f = message.getKeyPreview();
  if (f.length > 0) {
    writer.writeString(
      9,
      f
    );
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.api.v1.ApiKey.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ApiKey} returns this
 */
proto.api.v1.ApiKey.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string name = 2;
 * @return {string}
 */
proto.api.v1.ApiKey.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ApiKey} returns this
 */
proto.api.v1.ApiKey.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string key = 3;
 * @return {string}
 */
proto.api.v1.ApiKey.prototype.getKey = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ApiKey} returns this
 */
proto.api.v1.ApiKey.prototype.setKey = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string session_id = 4;
 * @return {string}
 */
proto.api.v1.ApiKey.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ApiKey} returns this
 */
proto.api.v1.ApiKey.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional google.protobuf.Timestamp created_at = 5;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.api.v1.ApiKey.prototype.getCreatedAt = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 5));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.api.v1.ApiKey} returns this
*/
proto.api.v1.ApiKey.prototype.setCreatedAt = function(value) {
  return jspb.Message.setWrapperField(this, 5, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api.v1.ApiKey} returns this
 */
proto.api.v1.ApiKey.prototype.clearCreatedAt = function() {
  return this.setCreatedAt(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api.v1.ApiKey.prototype.hasCreatedAt = function() {
  return jspb.Message.getField(this, 5) != null;
};


/**
 * optional string created_by = 6;
 * @return {string}
 */
proto.api.v1.ApiKey.prototype.getCreatedBy = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ApiKey} returns this
 */
proto.api.v1.ApiKey.prototype.setCreatedBy = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * optional google.protobuf.Timestamp revoked_at = 7;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.api.v1.ApiKey.prototype.getRevokedAt = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 7));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.api.v1.ApiKey} returns this
*/
proto.api.v1.ApiKey.prototype.setRevokedAt = function(value) {
  return jspb.Message.setWrapperField(this, 7, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api.v1.ApiKey} returns this
 */
proto.api.v1.ApiKey.prototype.clearRevokedAt = function() {
  return this.setRevokedAt(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api.v1.ApiKey.prototype.hasRevokedAt = function() {
  return jspb.Message.getField(this, 7) != null;
};


/**
 * repeated string capabilities = 8;
 * @return {!Array<string>}
 */
proto.api.v1.ApiKey.prototype.getCapabilitiesList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 8));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.api.v1.ApiKey} returns this
 */
proto.api.v1.ApiKey.prototype.setCapabilitiesList = function(value) {
  return jspb.Message.setField(this, 8, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.api.v1.ApiKey} returns this
 */
proto.api.v1.ApiKey.prototype.addCapabilities = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 8, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api.v1.ApiKey} returns this
 */
proto.api.v1.ApiKey.prototype.clearCapabilitiesList = function() {
  return this.setCapabilitiesList([]);
};


/**
 * optional string key_preview = 9;
 * @return {string}
 */
proto.api.v1.ApiKey.prototype.getKeyPreview = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 9, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ApiKey} returns this
 */
proto.api.v1.ApiKey.prototype.setKeyPreview = function(value) {
  return jspb.Message.setProto3StringField(this, 9, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.Machine.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.Machine.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.Machine} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.Machine.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, ""),
    sessionId: jspb.Message.getFieldWithDefault(msg, 2, ""),
    sdkVersion: jspb.Message.getFieldWithDefault(msg, 3, ""),
    sdkLanguage: jspb.Message.getFieldWithDefault(msg, 4, ""),
    ip: jspb.Message.getFieldWithDefault(msg, 5, ""),
    createdAt: (f = msg.getCreatedAt()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    lastPingAt: (f = msg.getLastPingAt()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    machineToken: jspb.Message.getFieldWithDefault(msg, 8, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.Machine}
 */
proto.api.v1.Machine.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.Machine;
  return proto.api.v1.Machine.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.Machine} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.Machine}
 */
proto.api.v1.Machine.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setSdkVersion(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setSdkLanguage(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setIp(value);
      break;
    case 6:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setCreatedAt(value);
      break;
    case 7:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setLastPingAt(value);
      break;
    case 8:
      var value = /** @type {string} */ (reader.readString());
      msg.setMachineToken(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.Machine.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.Machine.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.Machine} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.Machine.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getSdkVersion();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getSdkLanguage();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getIp();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getCreatedAt();
  if (f != null) {
    writer.writeMessage(
      6,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getLastPingAt();
  if (f != null) {
    writer.writeMessage(
      7,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getMachineToken();
  if (f.length > 0) {
    writer.writeString(
      8,
      f
    );
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.api.v1.Machine.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Machine} returns this
 */
proto.api.v1.Machine.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string session_id = 2;
 * @return {string}
 */
proto.api.v1.Machine.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Machine} returns this
 */
proto.api.v1.Machine.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string sdk_version = 3;
 * @return {string}
 */
proto.api.v1.Machine.prototype.getSdkVersion = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Machine} returns this
 */
proto.api.v1.Machine.prototype.setSdkVersion = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string sdk_language = 4;
 * @return {string}
 */
proto.api.v1.Machine.prototype.getSdkLanguage = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Machine} returns this
 */
proto.api.v1.Machine.prototype.setSdkLanguage = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string ip = 5;
 * @return {string}
 */
proto.api.v1.Machine.prototype.getIp = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Machine} returns this
 */
proto.api.v1.Machine.prototype.setIp = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional google.protobuf.Timestamp created_at = 6;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.api.v1.Machine.prototype.getCreatedAt = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 6));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.api.v1.Machine} returns this
*/
proto.api.v1.Machine.prototype.setCreatedAt = function(value) {
  return jspb.Message.setWrapperField(this, 6, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api.v1.Machine} returns this
 */
proto.api.v1.Machine.prototype.clearCreatedAt = function() {
  return this.setCreatedAt(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api.v1.Machine.prototype.hasCreatedAt = function() {
  return jspb.Message.getField(this, 6) != null;
};


/**
 * optional google.protobuf.Timestamp last_ping_at = 7;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.api.v1.Machine.prototype.getLastPingAt = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 7));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.api.v1.Machine} returns this
*/
proto.api.v1.Machine.prototype.setLastPingAt = function(value) {
  return jspb.Message.setWrapperField(this, 7, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api.v1.Machine} returns this
 */
proto.api.v1.Machine.prototype.clearLastPingAt = function() {
  return this.setLastPingAt(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api.v1.Machine.prototype.hasLastPingAt = function() {
  return jspb.Message.getField(this, 7) != null;
};


/**
 * optional string machine_token = 8;
 * @return {string}
 */
proto.api.v1.Machine.prototype.getMachineToken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 8, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Machine} returns this
 */
proto.api.v1.Machine.prototype.setMachineToken = function(value) {
  return jspb.Message.setProto3StringField(this, 8, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.Request.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.Request.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.Request} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.Request.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, ""),
    sessionId: jspb.Message.getFieldWithDefault(msg, 2, ""),
    toolName: jspb.Message.getFieldWithDefault(msg, 3, ""),
    status: jspb.Message.getFieldWithDefault(msg, 4, 0),
    input: jspb.Message.getFieldWithDefault(msg, 5, ""),
    result: jspb.Message.getFieldWithDefault(msg, 6, ""),
    resultType: jspb.Message.getFieldWithDefault(msg, 7, ""),
    error: jspb.Message.getFieldWithDefault(msg, 8, ""),
    createdAt: (f = msg.getCreatedAt()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    updatedAt: (f = msg.getUpdatedAt()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    executingMachineId: jspb.Message.getFieldWithDefault(msg, 11, ""),
    leasedBy: jspb.Message.getFieldWithDefault(msg, 13, ""),
    leaseEpoch: jspb.Message.getFieldWithDefault(msg, 14, 0),
    leaseExpiresAt: (f = msg.getLeaseExpiresAt()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    timeoutSeconds: jspb.Message.getFieldWithDefault(msg, 16, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.Request}
 */
proto.api.v1.Request.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.Request;
  return proto.api.v1.Request.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.Request} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.Request}
 */
proto.api.v1.Request.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setToolName(value);
      break;
    case 4:
      var value = /** @type {!proto.api.v1.RequestStatus} */ (reader.readEnum());
      msg.setStatus(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setInput(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setResult(value);
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.setResultType(value);
      break;
    case 8:
      var value = /** @type {string} */ (reader.readString());
      msg.setError(value);
      break;
    case 9:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setCreatedAt(value);
      break;
    case 10:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setUpdatedAt(value);
      break;
    case 11:
      var value = /** @type {string} */ (reader.readString());
      msg.setExecutingMachineId(value);
      break;
    case 13:
      var value = /** @type {string} */ (reader.readString());
      msg.setLeasedBy(value);
      break;
    case 14:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setLeaseEpoch(value);
      break;
    case 15:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setLeaseExpiresAt(value);
      break;
    case 16:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setTimeoutSeconds(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.Request.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.Request.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.Request} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.Request.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getToolName();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getStatus();
  if (f !== 0.0) {
    writer.writeEnum(
      4,
      f
    );
  }
  f = message.getInput();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getResult();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getResultType();
  if (f.length > 0) {
    writer.writeString(
      7,
      f
    );
  }
  f = message.getError();
  if (f.length > 0) {
    writer.writeString(
      8,
      f
    );
  }
  f = message.getCreatedAt();
  if (f != null) {
    writer.writeMessage(
      9,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getUpdatedAt();
  if (f != null) {
    writer.writeMessage(
      10,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getExecutingMachineId();
  if (f.length > 0) {
    writer.writeString(
      11,
      f
    );
  }
  f = message.getLeasedBy();
  if (f.length > 0) {
    writer.writeString(
      13,
      f
    );
  }
  f = message.getLeaseEpoch();
  if (f !== 0) {
    writer.writeInt64(
      14,
      f
    );
  }
  f = message.getLeaseExpiresAt();
  if (f != null) {
    writer.writeMessage(
      15,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getTimeoutSeconds();
  if (f !== 0) {
    writer.writeInt32(
      16,
      f
    );
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.api.v1.Request.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Request} returns this
 */
proto.api.v1.Request.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string session_id = 2;
 * @return {string}
 */
proto.api.v1.Request.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Request} returns this
 */
proto.api.v1.Request.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string tool_name = 3;
 * @return {string}
 */
proto.api.v1.Request.prototype.getToolName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Request} returns this
 */
proto.api.v1.Request.prototype.setToolName = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional RequestStatus status = 4;
 * @return {!proto.api.v1.RequestStatus}
 */
proto.api.v1.Request.prototype.getStatus = function() {
  return /** @type {!proto.api.v1.RequestStatus} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {!proto.api.v1.RequestStatus} value
 * @return {!proto.api.v1.Request} returns this
 */
proto.api.v1.Request.prototype.setStatus = function(value) {
  return jspb.Message.setProto3EnumField(this, 4, value);
};


/**
 * optional string input = 5;
 * @return {string}
 */
proto.api.v1.Request.prototype.getInput = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Request} returns this
 */
proto.api.v1.Request.prototype.setInput = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional string result = 6;
 * @return {string}
 */
proto.api.v1.Request.prototype.getResult = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Request} returns this
 */
proto.api.v1.Request.prototype.setResult = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * optional string result_type = 7;
 * @return {string}
 */
proto.api.v1.Request.prototype.getResultType = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 7, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Request} returns this
 */
proto.api.v1.Request.prototype.setResultType = function(value) {
  return jspb.Message.setProto3StringField(this, 7, value);
};


/**
 * optional string error = 8;
 * @return {string}
 */
proto.api.v1.Request.prototype.getError = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 8, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Request} returns this
 */
proto.api.v1.Request.prototype.setError = function(value) {
  return jspb.Message.setProto3StringField(this, 8, value);
};


/**
 * optional google.protobuf.Timestamp created_at = 9;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.api.v1.Request.prototype.getCreatedAt = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 9));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.api.v1.Request} returns this
*/
proto.api.v1.Request.prototype.setCreatedAt = function(value) {
  return jspb.Message.setWrapperField(this, 9, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api.v1.Request} returns this
 */
proto.api.v1.Request.prototype.clearCreatedAt = function() {
  return this.setCreatedAt(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api.v1.Request.prototype.hasCreatedAt = function() {
  return jspb.Message.getField(this, 9) != null;
};


/**
 * optional google.protobuf.Timestamp updated_at = 10;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.api.v1.Request.prototype.getUpdatedAt = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 10));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.api.v1.Request} returns this
*/
proto.api.v1.Request.prototype.setUpdatedAt = function(value) {
  return jspb.Message.setWrapperField(this, 10, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api.v1.Request} returns this
 */
proto.api.v1.Request.prototype.clearUpdatedAt = function() {
  return this.setUpdatedAt(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api.v1.Request.prototype.hasUpdatedAt = function() {
  return jspb.Message.getField(this, 10) != null;
};


/**
 * optional string executing_machine_id = 11;
 * @return {string}
 */
proto.api.v1.Request.prototype.getExecutingMachineId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 11, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Request} returns this
 */
proto.api.v1.Request.prototype.setExecutingMachineId = function(value) {
  return jspb.Message.setProto3StringField(this, 11, value);
};


/**
 * optional string leased_by = 13;
 * @return {string}
 */
proto.api.v1.Request.prototype.getLeasedBy = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 13, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Request} returns this
 */
proto.api.v1.Request.prototype.setLeasedBy = function(value) {
  return jspb.Message.setProto3StringField(this, 13, value);
};


/**
 * optional int64 lease_epoch = 14;
 * @return {number}
 */
proto.api.v1.Request.prototype.getLeaseEpoch = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 14, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.Request} returns this
 */
proto.api.v1.Request.prototype.setLeaseEpoch = function(value) {
  return jspb.Message.setProto3IntField(this, 14, value);
};


/**
 * optional google.protobuf.Timestamp lease_expires_at = 15;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.api.v1.Request.prototype.getLeaseExpiresAt = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 15));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.api.v1.Request} returns this
*/
proto.api.v1.Request.prototype.setLeaseExpiresAt = function(value) {
  return jspb.Message.setWrapperField(this, 15, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api.v1.Request} returns this
 */
proto.api.v1.Request.prototype.clearLeaseExpiresAt = function() {
  return this.setLeaseExpiresAt(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api.v1.Request.prototype.hasLeaseExpiresAt = function() {
  return jspb.Message.getField(this, 15) != null;
};


/**
 * optional int32 timeout_seconds = 16;
 * @return {number}
 */
proto.api.v1.Request.prototype.getTimeoutSeconds = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 16, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.Request} returns this
 */
proto.api.v1.Request.prototype.setTimeoutSeconds = function(value) {
  return jspb.Message.setProto3IntField(this, 16, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api.v1.RegisterToolRequest.repeatedFields_ = [7];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.RegisterToolRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.RegisterToolRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.RegisterToolRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.RegisterToolRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    machineId: jspb.Message.getFieldWithDefault(msg, 2, ""),
    name: jspb.Message.getFieldWithDefault(msg, 3, ""),
    description: jspb.Message.getFieldWithDefault(msg, 4, ""),
    schema: jspb.Message.getFieldWithDefault(msg, 5, ""),
    configMap: (f = msg.getConfigMap()) ? f.toObject(includeInstance, undefined) : [],
    tagsList: (f = jspb.Message.getRepeatedField(msg, 7)) == null ? undefined : f
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.RegisterToolRequest}
 */
proto.api.v1.RegisterToolRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.RegisterToolRequest;
  return proto.api.v1.RegisterToolRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.RegisterToolRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.RegisterToolRequest}
 */
proto.api.v1.RegisterToolRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setMachineId(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setSchema(value);
      break;
    case 6:
      var value = msg.getConfigMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString, null, "", "");
         });
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.addTags(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.RegisterToolRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.RegisterToolRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.RegisterToolRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.RegisterToolRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getMachineId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getDescription();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getSchema();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getConfigMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(6, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
  f = message.getTagsList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      7,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.RegisterToolRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.RegisterToolRequest} returns this
 */
proto.api.v1.RegisterToolRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string machine_id = 2;
 * @return {string}
 */
proto.api.v1.RegisterToolRequest.prototype.getMachineId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.RegisterToolRequest} returns this
 */
proto.api.v1.RegisterToolRequest.prototype.setMachineId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string name = 3;
 * @return {string}
 */
proto.api.v1.RegisterToolRequest.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.RegisterToolRequest} returns this
 */
proto.api.v1.RegisterToolRequest.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string description = 4;
 * @return {string}
 */
proto.api.v1.RegisterToolRequest.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.RegisterToolRequest} returns this
 */
proto.api.v1.RegisterToolRequest.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string schema = 5;
 * @return {string}
 */
proto.api.v1.RegisterToolRequest.prototype.getSchema = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.RegisterToolRequest} returns this
 */
proto.api.v1.RegisterToolRequest.prototype.setSchema = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * map<string, string> config = 6;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.api.v1.RegisterToolRequest.prototype.getConfigMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 6, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.api.v1.RegisterToolRequest} returns this
 */
proto.api.v1.RegisterToolRequest.prototype.clearConfigMap = function() {
  this.getConfigMap().clear();
  return this;};


/**
 * repeated string tags = 7;
 * @return {!Array<string>}
 */
proto.api.v1.RegisterToolRequest.prototype.getTagsList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 7));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.api.v1.RegisterToolRequest} returns this
 */
proto.api.v1.RegisterToolRequest.prototype.setTagsList = function(value) {
  return jspb.Message.setField(this, 7, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.api.v1.RegisterToolRequest} returns this
 */
proto.api.v1.RegisterToolRequest.prototype.addTags = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 7, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api.v1.RegisterToolRequest} returns this
 */
proto.api.v1.RegisterToolRequest.prototype.clearTagsList = function() {
  return this.setTagsList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.RegisterToolResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.RegisterToolResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.RegisterToolResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.RegisterToolResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    tool: (f = msg.getTool()) && proto.api.v1.Tool.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.RegisterToolResponse}
 */
proto.api.v1.RegisterToolResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.RegisterToolResponse;
  return proto.api.v1.RegisterToolResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.RegisterToolResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.RegisterToolResponse}
 */
proto.api.v1.RegisterToolResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.api.v1.Tool;
      reader.readMessage(value,proto.api.v1.Tool.deserializeBinaryFromReader);
      msg.setTool(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.RegisterToolResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.RegisterToolResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.RegisterToolResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.RegisterToolResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTool();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.api.v1.Tool.serializeBinaryToWriter
    );
  }
};


/**
 * optional Tool tool = 1;
 * @return {?proto.api.v1.Tool}
 */
proto.api.v1.RegisterToolResponse.prototype.getTool = function() {
  return /** @type{?proto.api.v1.Tool} */ (
    jspb.Message.getWrapperField(this, proto.api.v1.Tool, 1));
};


/**
 * @param {?proto.api.v1.Tool|undefined} value
 * @return {!proto.api.v1.RegisterToolResponse} returns this
*/
proto.api.v1.RegisterToolResponse.prototype.setTool = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api.v1.RegisterToolResponse} returns this
 */
proto.api.v1.RegisterToolResponse.prototype.clearTool = function() {
  return this.setTool(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api.v1.RegisterToolResponse.prototype.hasTool = function() {
  return jspb.Message.getField(this, 1) != null;
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ListToolsRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ListToolsRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ListToolsRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListToolsRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ListToolsRequest}
 */
proto.api.v1.ListToolsRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ListToolsRequest;
  return proto.api.v1.ListToolsRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ListToolsRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ListToolsRequest}
 */
proto.api.v1.ListToolsRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ListToolsRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ListToolsRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ListToolsRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListToolsRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.ListToolsRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ListToolsRequest} returns this
 */
proto.api.v1.ListToolsRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api.v1.ListToolsResponse.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ListToolsResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ListToolsResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ListToolsResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListToolsResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    toolsList: jspb.Message.toObjectList(msg.getToolsList(),
    proto.api.v1.Tool.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ListToolsResponse}
 */
proto.api.v1.ListToolsResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ListToolsResponse;
  return proto.api.v1.ListToolsResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ListToolsResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ListToolsResponse}
 */
proto.api.v1.ListToolsResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.api.v1.Tool;
      reader.readMessage(value,proto.api.v1.Tool.deserializeBinaryFromReader);
      msg.addTools(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ListToolsResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ListToolsResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ListToolsResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListToolsResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getToolsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.api.v1.Tool.serializeBinaryToWriter
    );
  }
};


/**
 * repeated Tool tools = 1;
 * @return {!Array<!proto.api.v1.Tool>}
 */
proto.api.v1.ListToolsResponse.prototype.getToolsList = function() {
  return /** @type{!Array<!proto.api.v1.Tool>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.api.v1.Tool, 1));
};


/**
 * @param {!Array<!proto.api.v1.Tool>} value
 * @return {!proto.api.v1.ListToolsResponse} returns this
*/
proto.api.v1.ListToolsResponse.prototype.setToolsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.api.v1.Tool=} opt_value
 * @param {number=} opt_index
 * @return {!proto.api.v1.Tool}
 */
proto.api.v1.ListToolsResponse.prototype.addTools = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.api.v1.Tool, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api.v1.ListToolsResponse} returns this
 */
proto.api.v1.ListToolsResponse.prototype.clearToolsList = function() {
  return this.setToolsList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.GetToolRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.GetToolRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.GetToolRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetToolRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    toolId: jspb.Message.getFieldWithDefault(msg, 2, ""),
    toolName: jspb.Message.getFieldWithDefault(msg, 3, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.GetToolRequest}
 */
proto.api.v1.GetToolRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.GetToolRequest;
  return proto.api.v1.GetToolRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.GetToolRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.GetToolRequest}
 */
proto.api.v1.GetToolRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setToolId(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setToolName(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.GetToolRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.GetToolRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.GetToolRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetToolRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getToolId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getToolName();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.GetToolRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.GetToolRequest} returns this
 */
proto.api.v1.GetToolRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string tool_id = 2;
 * @return {string}
 */
proto.api.v1.GetToolRequest.prototype.getToolId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.GetToolRequest} returns this
 */
proto.api.v1.GetToolRequest.prototype.setToolId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string tool_name = 3;
 * @return {string}
 */
proto.api.v1.GetToolRequest.prototype.getToolName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.GetToolRequest} returns this
 */
proto.api.v1.GetToolRequest.prototype.setToolName = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.GetToolByIdRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.GetToolByIdRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.GetToolByIdRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetToolByIdRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    toolId: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.GetToolByIdRequest}
 */
proto.api.v1.GetToolByIdRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.GetToolByIdRequest;
  return proto.api.v1.GetToolByIdRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.GetToolByIdRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.GetToolByIdRequest}
 */
proto.api.v1.GetToolByIdRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setToolId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.GetToolByIdRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.GetToolByIdRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.GetToolByIdRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetToolByIdRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getToolId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.GetToolByIdRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.GetToolByIdRequest} returns this
 */
proto.api.v1.GetToolByIdRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string tool_id = 2;
 * @return {string}
 */
proto.api.v1.GetToolByIdRequest.prototype.getToolId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.GetToolByIdRequest} returns this
 */
proto.api.v1.GetToolByIdRequest.prototype.setToolId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.GetToolByNameRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.GetToolByNameRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.GetToolByNameRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetToolByNameRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    toolName: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.GetToolByNameRequest}
 */
proto.api.v1.GetToolByNameRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.GetToolByNameRequest;
  return proto.api.v1.GetToolByNameRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.GetToolByNameRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.GetToolByNameRequest}
 */
proto.api.v1.GetToolByNameRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setToolName(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.GetToolByNameRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.GetToolByNameRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.GetToolByNameRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetToolByNameRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getToolName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.GetToolByNameRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.GetToolByNameRequest} returns this
 */
proto.api.v1.GetToolByNameRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string tool_name = 2;
 * @return {string}
 */
proto.api.v1.GetToolByNameRequest.prototype.getToolName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.GetToolByNameRequest} returns this
 */
proto.api.v1.GetToolByNameRequest.prototype.setToolName = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.GetToolResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.GetToolResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.GetToolResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetToolResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    tool: (f = msg.getTool()) && proto.api.v1.Tool.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.GetToolResponse}
 */
proto.api.v1.GetToolResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.GetToolResponse;
  return proto.api.v1.GetToolResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.GetToolResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.GetToolResponse}
 */
proto.api.v1.GetToolResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.api.v1.Tool;
      reader.readMessage(value,proto.api.v1.Tool.deserializeBinaryFromReader);
      msg.setTool(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.GetToolResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.GetToolResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.GetToolResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetToolResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTool();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.api.v1.Tool.serializeBinaryToWriter
    );
  }
};


/**
 * optional Tool tool = 1;
 * @return {?proto.api.v1.Tool}
 */
proto.api.v1.GetToolResponse.prototype.getTool = function() {
  return /** @type{?proto.api.v1.Tool} */ (
    jspb.Message.getWrapperField(this, proto.api.v1.Tool, 1));
};


/**
 * @param {?proto.api.v1.Tool|undefined} value
 * @return {!proto.api.v1.GetToolResponse} returns this
*/
proto.api.v1.GetToolResponse.prototype.setTool = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api.v1.GetToolResponse} returns this
 */
proto.api.v1.GetToolResponse.prototype.clearTool = function() {
  return this.setTool(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api.v1.GetToolResponse.prototype.hasTool = function() {
  return jspb.Message.getField(this, 1) != null;
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.DeleteToolRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.DeleteToolRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.DeleteToolRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.DeleteToolRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    toolId: jspb.Message.getFieldWithDefault(msg, 2, ""),
    machineId: jspb.Message.getFieldWithDefault(msg, 3, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.DeleteToolRequest}
 */
proto.api.v1.DeleteToolRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.DeleteToolRequest;
  return proto.api.v1.DeleteToolRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.DeleteToolRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.DeleteToolRequest}
 */
proto.api.v1.DeleteToolRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setToolId(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setMachineId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.DeleteToolRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.DeleteToolRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.DeleteToolRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.DeleteToolRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getToolId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getMachineId();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.DeleteToolRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.DeleteToolRequest} returns this
 */
proto.api.v1.DeleteToolRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string tool_id = 2;
 * @return {string}
 */
proto.api.v1.DeleteToolRequest.prototype.getToolId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.DeleteToolRequest} returns this
 */
proto.api.v1.DeleteToolRequest.prototype.setToolId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string machine_id = 3;
 * @return {string}
 */
proto.api.v1.DeleteToolRequest.prototype.getMachineId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.DeleteToolRequest} returns this
 */
proto.api.v1.DeleteToolRequest.prototype.setMachineId = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.DeleteToolResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.DeleteToolResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.DeleteToolResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.DeleteToolResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    success: jspb.Message.getBooleanFieldWithDefault(msg, 1, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.DeleteToolResponse}
 */
proto.api.v1.DeleteToolResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.DeleteToolResponse;
  return proto.api.v1.DeleteToolResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.DeleteToolResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.DeleteToolResponse}
 */
proto.api.v1.DeleteToolResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setSuccess(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.DeleteToolResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.DeleteToolResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.DeleteToolResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.DeleteToolResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSuccess();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
};


/**
 * optional bool success = 1;
 * @return {boolean}
 */
proto.api.v1.DeleteToolResponse.prototype.getSuccess = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.api.v1.DeleteToolResponse} returns this
 */
proto.api.v1.DeleteToolResponse.prototype.setSuccess = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.UpdateToolPingRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.UpdateToolPingRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.UpdateToolPingRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.UpdateToolPingRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    toolId: jspb.Message.getFieldWithDefault(msg, 2, ""),
    machineId: jspb.Message.getFieldWithDefault(msg, 3, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.UpdateToolPingRequest}
 */
proto.api.v1.UpdateToolPingRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.UpdateToolPingRequest;
  return proto.api.v1.UpdateToolPingRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.UpdateToolPingRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.UpdateToolPingRequest}
 */
proto.api.v1.UpdateToolPingRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setToolId(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setMachineId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.UpdateToolPingRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.UpdateToolPingRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.UpdateToolPingRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.UpdateToolPingRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getToolId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getMachineId();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.UpdateToolPingRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.UpdateToolPingRequest} returns this
 */
proto.api.v1.UpdateToolPingRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string tool_id = 2;
 * @return {string}
 */
proto.api.v1.UpdateToolPingRequest.prototype.getToolId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.UpdateToolPingRequest} returns this
 */
proto.api.v1.UpdateToolPingRequest.prototype.setToolId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string machine_id = 3;
 * @return {string}
 */
proto.api.v1.UpdateToolPingRequest.prototype.getMachineId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.UpdateToolPingRequest} returns this
 */
proto.api.v1.UpdateToolPingRequest.prototype.setMachineId = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.CreateSessionRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.CreateSessionRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.CreateSessionRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.CreateSessionRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    userId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    name: jspb.Message.getFieldWithDefault(msg, 2, ""),
    description: jspb.Message.getFieldWithDefault(msg, 3, ""),
    sessionId: jspb.Message.getFieldWithDefault(msg, 4, ""),
    namespace: jspb.Message.getFieldWithDefault(msg, 5, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.CreateSessionRequest}
 */
proto.api.v1.CreateSessionRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.CreateSessionRequest;
  return proto.api.v1.CreateSessionRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.CreateSessionRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.CreateSessionRequest}
 */
proto.api.v1.CreateSessionRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setUserId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setNamespace(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.CreateSessionRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.CreateSessionRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.CreateSessionRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.CreateSessionRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getUserId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getDescription();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getNamespace();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
};


/**
 * optional string user_id = 1;
 * @return {string}
 */
proto.api.v1.CreateSessionRequest.prototype.getUserId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.CreateSessionRequest} returns this
 */
proto.api.v1.CreateSessionRequest.prototype.setUserId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string name = 2;
 * @return {string}
 */
proto.api.v1.CreateSessionRequest.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.CreateSessionRequest} returns this
 */
proto.api.v1.CreateSessionRequest.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string description = 3;
 * @return {string}
 */
proto.api.v1.CreateSessionRequest.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.CreateSessionRequest} returns this
 */
proto.api.v1.CreateSessionRequest.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string session_id = 4;
 * @return {string}
 */
proto.api.v1.CreateSessionRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.CreateSessionRequest} returns this
 */
proto.api.v1.CreateSessionRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string namespace = 5;
 * @return {string}
 */
proto.api.v1.CreateSessionRequest.prototype.getNamespace = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.CreateSessionRequest} returns this
 */
proto.api.v1.CreateSessionRequest.prototype.setNamespace = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.CreateSessionResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.CreateSessionResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.CreateSessionResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.CreateSessionResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    session: (f = msg.getSession()) && proto.api.v1.Session.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.CreateSessionResponse}
 */
proto.api.v1.CreateSessionResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.CreateSessionResponse;
  return proto.api.v1.CreateSessionResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.CreateSessionResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.CreateSessionResponse}
 */
proto.api.v1.CreateSessionResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.api.v1.Session;
      reader.readMessage(value,proto.api.v1.Session.deserializeBinaryFromReader);
      msg.setSession(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.CreateSessionResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.CreateSessionResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.CreateSessionResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.CreateSessionResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSession();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.api.v1.Session.serializeBinaryToWriter
    );
  }
};


/**
 * optional Session session = 1;
 * @return {?proto.api.v1.Session}
 */
proto.api.v1.CreateSessionResponse.prototype.getSession = function() {
  return /** @type{?proto.api.v1.Session} */ (
    jspb.Message.getWrapperField(this, proto.api.v1.Session, 1));
};


/**
 * @param {?proto.api.v1.Session|undefined} value
 * @return {!proto.api.v1.CreateSessionResponse} returns this
*/
proto.api.v1.CreateSessionResponse.prototype.setSession = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api.v1.CreateSessionResponse} returns this
 */
proto.api.v1.CreateSessionResponse.prototype.clearSession = function() {
  return this.setSession(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api.v1.CreateSessionResponse.prototype.hasSession = function() {
  return jspb.Message.getField(this, 1) != null;
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.GetSessionRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.GetSessionRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.GetSessionRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetSessionRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.GetSessionRequest}
 */
proto.api.v1.GetSessionRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.GetSessionRequest;
  return proto.api.v1.GetSessionRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.GetSessionRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.GetSessionRequest}
 */
proto.api.v1.GetSessionRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.GetSessionRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.GetSessionRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.GetSessionRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetSessionRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.GetSessionRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.GetSessionRequest} returns this
 */
proto.api.v1.GetSessionRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ListSessionsRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ListSessionsRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ListSessionsRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListSessionsRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    userId: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ListSessionsRequest}
 */
proto.api.v1.ListSessionsRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ListSessionsRequest;
  return proto.api.v1.ListSessionsRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ListSessionsRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ListSessionsRequest}
 */
proto.api.v1.ListSessionsRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setUserId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ListSessionsRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ListSessionsRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ListSessionsRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListSessionsRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getUserId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string user_id = 1;
 * @return {string}
 */
proto.api.v1.ListSessionsRequest.prototype.getUserId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ListSessionsRequest} returns this
 */
proto.api.v1.ListSessionsRequest.prototype.setUserId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api.v1.ListSessionsResponse.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ListSessionsResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ListSessionsResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ListSessionsResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListSessionsResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionsList: jspb.Message.toObjectList(msg.getSessionsList(),
    proto.api.v1.Session.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ListSessionsResponse}
 */
proto.api.v1.ListSessionsResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ListSessionsResponse;
  return proto.api.v1.ListSessionsResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ListSessionsResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ListSessionsResponse}
 */
proto.api.v1.ListSessionsResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.api.v1.Session;
      reader.readMessage(value,proto.api.v1.Session.deserializeBinaryFromReader);
      msg.addSessions(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ListSessionsResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ListSessionsResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ListSessionsResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListSessionsResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.api.v1.Session.serializeBinaryToWriter
    );
  }
};


/**
 * repeated Session sessions = 1;
 * @return {!Array<!proto.api.v1.Session>}
 */
proto.api.v1.ListSessionsResponse.prototype.getSessionsList = function() {
  return /** @type{!Array<!proto.api.v1.Session>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.api.v1.Session, 1));
};


/**
 * @param {!Array<!proto.api.v1.Session>} value
 * @return {!proto.api.v1.ListSessionsResponse} returns this
*/
proto.api.v1.ListSessionsResponse.prototype.setSessionsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.api.v1.Session=} opt_value
 * @param {number=} opt_index
 * @return {!proto.api.v1.Session}
 */
proto.api.v1.ListSessionsResponse.prototype.addSessions = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.api.v1.Session, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api.v1.ListSessionsResponse} returns this
 */
proto.api.v1.ListSessionsResponse.prototype.clearSessionsList = function() {
  return this.setSessionsList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ListUserSessionsRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ListUserSessionsRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ListUserSessionsRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListUserSessionsRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    userId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    pageSize: jspb.Message.getFieldWithDefault(msg, 2, 0),
    pageToken: jspb.Message.getFieldWithDefault(msg, 3, ""),
    filter: jspb.Message.getFieldWithDefault(msg, 4, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ListUserSessionsRequest}
 */
proto.api.v1.ListUserSessionsRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ListUserSessionsRequest;
  return proto.api.v1.ListUserSessionsRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ListUserSessionsRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ListUserSessionsRequest}
 */
proto.api.v1.ListUserSessionsRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setUserId(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setPageSize(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setPageToken(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setFilter(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ListUserSessionsRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ListUserSessionsRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ListUserSessionsRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListUserSessionsRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getUserId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getPageSize();
  if (f !== 0) {
    writer.writeInt32(
      2,
      f
    );
  }
  f = message.getPageToken();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getFilter();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
};


/**
 * optional string user_id = 1;
 * @return {string}
 */
proto.api.v1.ListUserSessionsRequest.prototype.getUserId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ListUserSessionsRequest} returns this
 */
proto.api.v1.ListUserSessionsRequest.prototype.setUserId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional int32 page_size = 2;
 * @return {number}
 */
proto.api.v1.ListUserSessionsRequest.prototype.getPageSize = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.ListUserSessionsRequest} returns this
 */
proto.api.v1.ListUserSessionsRequest.prototype.setPageSize = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional string page_token = 3;
 * @return {string}
 */
proto.api.v1.ListUserSessionsRequest.prototype.getPageToken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ListUserSessionsRequest} returns this
 */
proto.api.v1.ListUserSessionsRequest.prototype.setPageToken = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string filter = 4;
 * @return {string}
 */
proto.api.v1.ListUserSessionsRequest.prototype.getFilter = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ListUserSessionsRequest} returns this
 */
proto.api.v1.ListUserSessionsRequest.prototype.setFilter = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api.v1.ListUserSessionsResponse.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ListUserSessionsResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ListUserSessionsResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ListUserSessionsResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListUserSessionsResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionsList: jspb.Message.toObjectList(msg.getSessionsList(),
    proto.api.v1.Session.toObject, includeInstance),
    page: (f = msg.getPage()) && proto.api.v1.ListPage.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ListUserSessionsResponse}
 */
proto.api.v1.ListUserSessionsResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ListUserSessionsResponse;
  return proto.api.v1.ListUserSessionsResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ListUserSessionsResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ListUserSessionsResponse}
 */
proto.api.v1.ListUserSessionsResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.api.v1.Session;
      reader.readMessage(value,proto.api.v1.Session.deserializeBinaryFromReader);
      msg.addSessions(value);
      break;
    case 2:
      var value = new proto.api.v1.ListPage;
      reader.readMessage(value,proto.api.v1.ListPage.deserializeBinaryFromReader);
      msg.setPage(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ListUserSessionsResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ListUserSessionsResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ListUserSessionsResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListUserSessionsResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.api.v1.Session.serializeBinaryToWriter
    );
  }
  f = message.getPage();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.api.v1.ListPage.serializeBinaryToWriter
    );
  }
};


/**
 * repeated Session sessions = 1;
 * @return {!Array<!proto.api.v1.Session>}
 */
proto.api.v1.ListUserSessionsResponse.prototype.getSessionsList = function() {
  return /** @type{!Array<!proto.api.v1.Session>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.api.v1.Session, 1));
};


/**
 * @param {!Array<!proto.api.v1.Session>} value
 * @return {!proto.api.v1.ListUserSessionsResponse} returns this
*/
proto.api.v1.ListUserSessionsResponse.prototype.setSessionsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.api.v1.Session=} opt_value
 * @param {number=} opt_index
 * @return {!proto.api.v1.Session}
 */
proto.api.v1.ListUserSessionsResponse.prototype.addSessions = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.api.v1.Session, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api.v1.ListUserSessionsResponse} returns this
 */
proto.api.v1.ListUserSessionsResponse.prototype.clearSessionsList = function() {
  return this.setSessionsList([]);
};


/**
 * optional ListPage page = 2;
 * @return {?proto.api.v1.ListPage}
 */
proto.api.v1.ListUserSessionsResponse.prototype.getPage = function() {
  return /** @type{?proto.api.v1.ListPage} */ (
    jspb.Message.getWrapperField(this, proto.api.v1.ListPage, 2));
};


/**
 * @param {?proto.api.v1.ListPage|undefined} value
 * @return {!proto.api.v1.ListUserSessionsResponse} returns this
*/
proto.api.v1.ListUserSessionsResponse.prototype.setPage = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api.v1.ListUserSessionsResponse} returns this
 */
proto.api.v1.ListUserSessionsResponse.prototype.clearPage = function() {
  return this.setPage(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api.v1.ListUserSessionsResponse.prototype.hasPage = function() {
  return jspb.Message.getField(this, 2) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api.v1.BulkDeleteSessionsRequest.repeatedFields_ = [2];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.BulkDeleteSessionsRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.BulkDeleteSessionsRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.BulkDeleteSessionsRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.BulkDeleteSessionsRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    userId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    sessionIdsList: (f = jspb.Message.getRepeatedField(msg, 2)) == null ? undefined : f,
    filter: jspb.Message.getFieldWithDefault(msg, 3, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.BulkDeleteSessionsRequest}
 */
proto.api.v1.BulkDeleteSessionsRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.BulkDeleteSessionsRequest;
  return proto.api.v1.BulkDeleteSessionsRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.BulkDeleteSessionsRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.BulkDeleteSessionsRequest}
 */
proto.api.v1.BulkDeleteSessionsRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setUserId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.addSessionIds(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setFilter(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.BulkDeleteSessionsRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.BulkDeleteSessionsRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.BulkDeleteSessionsRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.BulkDeleteSessionsRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getUserId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getSessionIdsList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      2,
      f
    );
  }
  f = message.getFilter();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
};


/**
 * optional string user_id = 1;
 * @return {string}
 */
proto.api.v1.BulkDeleteSessionsRequest.prototype.getUserId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.BulkDeleteSessionsRequest} returns this
 */
proto.api.v1.BulkDeleteSessionsRequest.prototype.setUserId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * repeated string session_ids = 2;
 * @return {!Array<string>}
 */
proto.api.v1.BulkDeleteSessionsRequest.prototype.getSessionIdsList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 2));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.api.v1.BulkDeleteSessionsRequest} returns this
 */
proto.api.v1.BulkDeleteSessionsRequest.prototype.setSessionIdsList = function(value) {
  return jspb.Message.setField(this, 2, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.api.v1.BulkDeleteSessionsRequest} returns this
 */
proto.api.v1.BulkDeleteSessionsRequest.prototype.addSessionIds = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 2, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api.v1.BulkDeleteSessionsRequest} returns this
 */
proto.api.v1.BulkDeleteSessionsRequest.prototype.clearSessionIdsList = function() {
  return this.setSessionIdsList([]);
};


/**
 * optional string filter = 3;
 * @return {string}
 */
proto.api.v1.BulkDeleteSessionsRequest.prototype.getFilter = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.BulkDeleteSessionsRequest} returns this
 */
proto.api.v1.BulkDeleteSessionsRequest.prototype.setFilter = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api.v1.BulkDeleteSessionsResponse.repeatedFields_ = [2];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.BulkDeleteSessionsResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.BulkDeleteSessionsResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.BulkDeleteSessionsResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.BulkDeleteSessionsResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    deletedCount: jspb.Message.getFieldWithDefault(msg, 1, 0),
    failedDeletionsList: (f = jspb.Message.getRepeatedField(msg, 2)) == null ? undefined : f
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.BulkDeleteSessionsResponse}
 */
proto.api.v1.BulkDeleteSessionsResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.BulkDeleteSessionsResponse;
  return proto.api.v1.BulkDeleteSessionsResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.BulkDeleteSessionsResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.BulkDeleteSessionsResponse}
 */
proto.api.v1.BulkDeleteSessionsResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setDeletedCount(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.addFailedDeletions(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.BulkDeleteSessionsResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.BulkDeleteSessionsResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.BulkDeleteSessionsResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.BulkDeleteSessionsResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getDeletedCount();
  if (f !== 0) {
    writer.writeInt32(
      1,
      f
    );
  }
  f = message.getFailedDeletionsList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      2,
      f
    );
  }
};


/**
 * optional int32 deleted_count = 1;
 * @return {number}
 */
proto.api.v1.BulkDeleteSessionsResponse.prototype.getDeletedCount = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.BulkDeleteSessionsResponse} returns this
 */
proto.api.v1.BulkDeleteSessionsResponse.prototype.setDeletedCount = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * repeated string failed_deletions = 2;
 * @return {!Array<string>}
 */
proto.api.v1.BulkDeleteSessionsResponse.prototype.getFailedDeletionsList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 2));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.api.v1.BulkDeleteSessionsResponse} returns this
 */
proto.api.v1.BulkDeleteSessionsResponse.prototype.setFailedDeletionsList = function(value) {
  return jspb.Message.setField(this, 2, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.api.v1.BulkDeleteSessionsResponse} returns this
 */
proto.api.v1.BulkDeleteSessionsResponse.prototype.addFailedDeletions = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 2, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api.v1.BulkDeleteSessionsResponse} returns this
 */
proto.api.v1.BulkDeleteSessionsResponse.prototype.clearFailedDeletionsList = function() {
  return this.setFailedDeletionsList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.GetSessionStatsRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.GetSessionStatsRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.GetSessionStatsRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetSessionStatsRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    userId: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.GetSessionStatsRequest}
 */
proto.api.v1.GetSessionStatsRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.GetSessionStatsRequest;
  return proto.api.v1.GetSessionStatsRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.GetSessionStatsRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.GetSessionStatsRequest}
 */
proto.api.v1.GetSessionStatsRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setUserId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.GetSessionStatsRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.GetSessionStatsRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.GetSessionStatsRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetSessionStatsRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getUserId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string user_id = 1;
 * @return {string}
 */
proto.api.v1.GetSessionStatsRequest.prototype.getUserId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.GetSessionStatsRequest} returns this
 */
proto.api.v1.GetSessionStatsRequest.prototype.setUserId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.GetSessionStatsResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.GetSessionStatsResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.GetSessionStatsResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetSessionStatsResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    totalSessions: jspb.Message.getFieldWithDefault(msg, 1, 0),
    activeSessions: jspb.Message.getFieldWithDefault(msg, 2, 0),
    expiredSessions: jspb.Message.getFieldWithDefault(msg, 3, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.GetSessionStatsResponse}
 */
proto.api.v1.GetSessionStatsResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.GetSessionStatsResponse;
  return proto.api.v1.GetSessionStatsResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.GetSessionStatsResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.GetSessionStatsResponse}
 */
proto.api.v1.GetSessionStatsResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setTotalSessions(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setActiveSessions(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setExpiredSessions(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.GetSessionStatsResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.GetSessionStatsResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.GetSessionStatsResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetSessionStatsResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTotalSessions();
  if (f !== 0) {
    writer.writeInt32(
      1,
      f
    );
  }
  f = message.getActiveSessions();
  if (f !== 0) {
    writer.writeInt32(
      2,
      f
    );
  }
  f = message.getExpiredSessions();
  if (f !== 0) {
    writer.writeInt32(
      3,
      f
    );
  }
};


/**
 * optional int32 total_sessions = 1;
 * @return {number}
 */
proto.api.v1.GetSessionStatsResponse.prototype.getTotalSessions = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.GetSessionStatsResponse} returns this
 */
proto.api.v1.GetSessionStatsResponse.prototype.setTotalSessions = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional int32 active_sessions = 2;
 * @return {number}
 */
proto.api.v1.GetSessionStatsResponse.prototype.getActiveSessions = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.GetSessionStatsResponse} returns this
 */
proto.api.v1.GetSessionStatsResponse.prototype.setActiveSessions = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional int32 expired_sessions = 3;
 * @return {number}
 */
proto.api.v1.GetSessionStatsResponse.prototype.getExpiredSessions = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.GetSessionStatsResponse} returns this
 */
proto.api.v1.GetSessionStatsResponse.prototype.setExpiredSessions = function(value) {
  return jspb.Message.setProto3IntField(this, 3, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.InvalidateSessionRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.InvalidateSessionRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.InvalidateSessionRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.InvalidateSessionRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    reason: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.InvalidateSessionRequest}
 */
proto.api.v1.InvalidateSessionRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.InvalidateSessionRequest;
  return proto.api.v1.InvalidateSessionRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.InvalidateSessionRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.InvalidateSessionRequest}
 */
proto.api.v1.InvalidateSessionRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setReason(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.InvalidateSessionRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.InvalidateSessionRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.InvalidateSessionRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.InvalidateSessionRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getReason();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.InvalidateSessionRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.InvalidateSessionRequest} returns this
 */
proto.api.v1.InvalidateSessionRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string reason = 2;
 * @return {string}
 */
proto.api.v1.InvalidateSessionRequest.prototype.getReason = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.InvalidateSessionRequest} returns this
 */
proto.api.v1.InvalidateSessionRequest.prototype.setReason = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.InvalidateSessionResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.InvalidateSessionResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.InvalidateSessionResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.InvalidateSessionResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    success: jspb.Message.getBooleanFieldWithDefault(msg, 1, false),
    revokedApiKeys: jspb.Message.getFieldWithDefault(msg, 2, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.InvalidateSessionResponse}
 */
proto.api.v1.InvalidateSessionResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.InvalidateSessionResponse;
  return proto.api.v1.InvalidateSessionResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.InvalidateSessionResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.InvalidateSessionResponse}
 */
proto.api.v1.InvalidateSessionResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setSuccess(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setRevokedApiKeys(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.InvalidateSessionResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.InvalidateSessionResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.InvalidateSessionResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.InvalidateSessionResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSuccess();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
  f = message.getRevokedApiKeys();
  if (f !== 0) {
    writer.writeInt32(
      2,
      f
    );
  }
};


/**
 * optional bool success = 1;
 * @return {boolean}
 */
proto.api.v1.InvalidateSessionResponse.prototype.getSuccess = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.api.v1.InvalidateSessionResponse} returns this
 */
proto.api.v1.InvalidateSessionResponse.prototype.setSuccess = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};


/**
 * optional int32 revoked_api_keys = 2;
 * @return {number}
 */
proto.api.v1.InvalidateSessionResponse.prototype.getRevokedApiKeys = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.InvalidateSessionResponse} returns this
 */
proto.api.v1.InvalidateSessionResponse.prototype.setRevokedApiKeys = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.UpdateSessionRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.UpdateSessionRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.UpdateSessionRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.UpdateSessionRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    name: jspb.Message.getFieldWithDefault(msg, 2, ""),
    description: jspb.Message.getFieldWithDefault(msg, 3, ""),
    namespace: jspb.Message.getFieldWithDefault(msg, 4, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.UpdateSessionRequest}
 */
proto.api.v1.UpdateSessionRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.UpdateSessionRequest;
  return proto.api.v1.UpdateSessionRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.UpdateSessionRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.UpdateSessionRequest}
 */
proto.api.v1.UpdateSessionRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setNamespace(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.UpdateSessionRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.UpdateSessionRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.UpdateSessionRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.UpdateSessionRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getDescription();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getNamespace();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.UpdateSessionRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.UpdateSessionRequest} returns this
 */
proto.api.v1.UpdateSessionRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string name = 2;
 * @return {string}
 */
proto.api.v1.UpdateSessionRequest.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.UpdateSessionRequest} returns this
 */
proto.api.v1.UpdateSessionRequest.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string description = 3;
 * @return {string}
 */
proto.api.v1.UpdateSessionRequest.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.UpdateSessionRequest} returns this
 */
proto.api.v1.UpdateSessionRequest.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string namespace = 4;
 * @return {string}
 */
proto.api.v1.UpdateSessionRequest.prototype.getNamespace = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.UpdateSessionRequest} returns this
 */
proto.api.v1.UpdateSessionRequest.prototype.setNamespace = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.DeleteSessionRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.DeleteSessionRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.DeleteSessionRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.DeleteSessionRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.DeleteSessionRequest}
 */
proto.api.v1.DeleteSessionRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.DeleteSessionRequest;
  return proto.api.v1.DeleteSessionRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.DeleteSessionRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.DeleteSessionRequest}
 */
proto.api.v1.DeleteSessionRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.DeleteSessionRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.DeleteSessionRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.DeleteSessionRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.DeleteSessionRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.DeleteSessionRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.DeleteSessionRequest} returns this
 */
proto.api.v1.DeleteSessionRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.DeleteSessionResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.DeleteSessionResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.DeleteSessionResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.DeleteSessionResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    success: jspb.Message.getBooleanFieldWithDefault(msg, 1, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.DeleteSessionResponse}
 */
proto.api.v1.DeleteSessionResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.DeleteSessionResponse;
  return proto.api.v1.DeleteSessionResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.DeleteSessionResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.DeleteSessionResponse}
 */
proto.api.v1.DeleteSessionResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setSuccess(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.DeleteSessionResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.DeleteSessionResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.DeleteSessionResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.DeleteSessionResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSuccess();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
};


/**
 * optional bool success = 1;
 * @return {boolean}
 */
proto.api.v1.DeleteSessionResponse.prototype.getSuccess = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.api.v1.DeleteSessionResponse} returns this
 */
proto.api.v1.DeleteSessionResponse.prototype.setSuccess = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api.v1.CreateApiKeyRequest.repeatedFields_ = [3];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.CreateApiKeyRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.CreateApiKeyRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.CreateApiKeyRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.CreateApiKeyRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    name: jspb.Message.getFieldWithDefault(msg, 2, ""),
    capabilitiesList: (f = jspb.Message.getRepeatedField(msg, 3)) == null ? undefined : f
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.CreateApiKeyRequest}
 */
proto.api.v1.CreateApiKeyRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.CreateApiKeyRequest;
  return proto.api.v1.CreateApiKeyRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.CreateApiKeyRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.CreateApiKeyRequest}
 */
proto.api.v1.CreateApiKeyRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.addCapabilities(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.CreateApiKeyRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.CreateApiKeyRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.CreateApiKeyRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.CreateApiKeyRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getCapabilitiesList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      3,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.CreateApiKeyRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.CreateApiKeyRequest} returns this
 */
proto.api.v1.CreateApiKeyRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string name = 2;
 * @return {string}
 */
proto.api.v1.CreateApiKeyRequest.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.CreateApiKeyRequest} returns this
 */
proto.api.v1.CreateApiKeyRequest.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * repeated string capabilities = 3;
 * @return {!Array<string>}
 */
proto.api.v1.CreateApiKeyRequest.prototype.getCapabilitiesList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 3));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.api.v1.CreateApiKeyRequest} returns this
 */
proto.api.v1.CreateApiKeyRequest.prototype.setCapabilitiesList = function(value) {
  return jspb.Message.setField(this, 3, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.api.v1.CreateApiKeyRequest} returns this
 */
proto.api.v1.CreateApiKeyRequest.prototype.addCapabilities = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 3, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api.v1.CreateApiKeyRequest} returns this
 */
proto.api.v1.CreateApiKeyRequest.prototype.clearCapabilitiesList = function() {
  return this.setCapabilitiesList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ListApiKeysRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ListApiKeysRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ListApiKeysRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListApiKeysRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ListApiKeysRequest}
 */
proto.api.v1.ListApiKeysRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ListApiKeysRequest;
  return proto.api.v1.ListApiKeysRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ListApiKeysRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ListApiKeysRequest}
 */
proto.api.v1.ListApiKeysRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ListApiKeysRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ListApiKeysRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ListApiKeysRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListApiKeysRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.ListApiKeysRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ListApiKeysRequest} returns this
 */
proto.api.v1.ListApiKeysRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api.v1.ListApiKeysResponse.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ListApiKeysResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ListApiKeysResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ListApiKeysResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListApiKeysResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    apiKeysList: jspb.Message.toObjectList(msg.getApiKeysList(),
    proto.api.v1.ApiKey.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ListApiKeysResponse}
 */
proto.api.v1.ListApiKeysResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ListApiKeysResponse;
  return proto.api.v1.ListApiKeysResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ListApiKeysResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ListApiKeysResponse}
 */
proto.api.v1.ListApiKeysResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.api.v1.ApiKey;
      reader.readMessage(value,proto.api.v1.ApiKey.deserializeBinaryFromReader);
      msg.addApiKeys(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ListApiKeysResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ListApiKeysResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ListApiKeysResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListApiKeysResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getApiKeysList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.api.v1.ApiKey.serializeBinaryToWriter
    );
  }
};


/**
 * repeated ApiKey api_keys = 1;
 * @return {!Array<!proto.api.v1.ApiKey>}
 */
proto.api.v1.ListApiKeysResponse.prototype.getApiKeysList = function() {
  return /** @type{!Array<!proto.api.v1.ApiKey>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.api.v1.ApiKey, 1));
};


/**
 * @param {!Array<!proto.api.v1.ApiKey>} value
 * @return {!proto.api.v1.ListApiKeysResponse} returns this
*/
proto.api.v1.ListApiKeysResponse.prototype.setApiKeysList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.api.v1.ApiKey=} opt_value
 * @param {number=} opt_index
 * @return {!proto.api.v1.ApiKey}
 */
proto.api.v1.ListApiKeysResponse.prototype.addApiKeys = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.api.v1.ApiKey, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api.v1.ListApiKeysResponse} returns this
 */
proto.api.v1.ListApiKeysResponse.prototype.clearApiKeysList = function() {
  return this.setApiKeysList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.RevokeApiKeyRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.RevokeApiKeyRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.RevokeApiKeyRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.RevokeApiKeyRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    keyId: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.RevokeApiKeyRequest}
 */
proto.api.v1.RevokeApiKeyRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.RevokeApiKeyRequest;
  return proto.api.v1.RevokeApiKeyRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.RevokeApiKeyRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.RevokeApiKeyRequest}
 */
proto.api.v1.RevokeApiKeyRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setKeyId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.RevokeApiKeyRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.RevokeApiKeyRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.RevokeApiKeyRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.RevokeApiKeyRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getKeyId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.RevokeApiKeyRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.RevokeApiKeyRequest} returns this
 */
proto.api.v1.RevokeApiKeyRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string key_id = 2;
 * @return {string}
 */
proto.api.v1.RevokeApiKeyRequest.prototype.getKeyId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.RevokeApiKeyRequest} returns this
 */
proto.api.v1.RevokeApiKeyRequest.prototype.setKeyId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.RevokeApiKeyResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.RevokeApiKeyResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.RevokeApiKeyResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.RevokeApiKeyResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    success: jspb.Message.getBooleanFieldWithDefault(msg, 1, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.RevokeApiKeyResponse}
 */
proto.api.v1.RevokeApiKeyResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.RevokeApiKeyResponse;
  return proto.api.v1.RevokeApiKeyResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.RevokeApiKeyResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.RevokeApiKeyResponse}
 */
proto.api.v1.RevokeApiKeyResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setSuccess(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.RevokeApiKeyResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.RevokeApiKeyResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.RevokeApiKeyResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.RevokeApiKeyResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSuccess();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
};


/**
 * optional bool success = 1;
 * @return {boolean}
 */
proto.api.v1.RevokeApiKeyResponse.prototype.getSuccess = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.api.v1.RevokeApiKeyResponse} returns this
 */
proto.api.v1.RevokeApiKeyResponse.prototype.setSuccess = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api.v1.RegisterMachineRequest.repeatedFields_ = [5];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.RegisterMachineRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.RegisterMachineRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.RegisterMachineRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.RegisterMachineRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    machineId: jspb.Message.getFieldWithDefault(msg, 2, ""),
    sdkVersion: jspb.Message.getFieldWithDefault(msg, 3, ""),
    sdkLanguage: jspb.Message.getFieldWithDefault(msg, 4, ""),
    toolsList: jspb.Message.toObjectList(msg.getToolsList(),
    proto.api.v1.RegisterToolRequest.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.RegisterMachineRequest}
 */
proto.api.v1.RegisterMachineRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.RegisterMachineRequest;
  return proto.api.v1.RegisterMachineRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.RegisterMachineRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.RegisterMachineRequest}
 */
proto.api.v1.RegisterMachineRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setMachineId(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setSdkVersion(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setSdkLanguage(value);
      break;
    case 5:
      var value = new proto.api.v1.RegisterToolRequest;
      reader.readMessage(value,proto.api.v1.RegisterToolRequest.deserializeBinaryFromReader);
      msg.addTools(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.RegisterMachineRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.RegisterMachineRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.RegisterMachineRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.RegisterMachineRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getMachineId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getSdkVersion();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getSdkLanguage();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getToolsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      5,
      f,
      proto.api.v1.RegisterToolRequest.serializeBinaryToWriter
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.RegisterMachineRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.RegisterMachineRequest} returns this
 */
proto.api.v1.RegisterMachineRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string machine_id = 2;
 * @return {string}
 */
proto.api.v1.RegisterMachineRequest.prototype.getMachineId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.RegisterMachineRequest} returns this
 */
proto.api.v1.RegisterMachineRequest.prototype.setMachineId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string sdk_version = 3;
 * @return {string}
 */
proto.api.v1.RegisterMachineRequest.prototype.getSdkVersion = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.RegisterMachineRequest} returns this
 */
proto.api.v1.RegisterMachineRequest.prototype.setSdkVersion = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string sdk_language = 4;
 * @return {string}
 */
proto.api.v1.RegisterMachineRequest.prototype.getSdkLanguage = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.RegisterMachineRequest} returns this
 */
proto.api.v1.RegisterMachineRequest.prototype.setSdkLanguage = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * repeated RegisterToolRequest tools = 5;
 * @return {!Array<!proto.api.v1.RegisterToolRequest>}
 */
proto.api.v1.RegisterMachineRequest.prototype.getToolsList = function() {
  return /** @type{!Array<!proto.api.v1.RegisterToolRequest>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.api.v1.RegisterToolRequest, 5));
};


/**
 * @param {!Array<!proto.api.v1.RegisterToolRequest>} value
 * @return {!proto.api.v1.RegisterMachineRequest} returns this
*/
proto.api.v1.RegisterMachineRequest.prototype.setToolsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 5, value);
};


/**
 * @param {!proto.api.v1.RegisterToolRequest=} opt_value
 * @param {number=} opt_index
 * @return {!proto.api.v1.RegisterToolRequest}
 */
proto.api.v1.RegisterMachineRequest.prototype.addTools = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 5, opt_value, proto.api.v1.RegisterToolRequest, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api.v1.RegisterMachineRequest} returns this
 */
proto.api.v1.RegisterMachineRequest.prototype.clearToolsList = function() {
  return this.setToolsList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ListMachinesRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ListMachinesRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ListMachinesRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListMachinesRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ListMachinesRequest}
 */
proto.api.v1.ListMachinesRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ListMachinesRequest;
  return proto.api.v1.ListMachinesRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ListMachinesRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ListMachinesRequest}
 */
proto.api.v1.ListMachinesRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ListMachinesRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ListMachinesRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ListMachinesRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListMachinesRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.ListMachinesRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ListMachinesRequest} returns this
 */
proto.api.v1.ListMachinesRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api.v1.ListMachinesResponse.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ListMachinesResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ListMachinesResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ListMachinesResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListMachinesResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    machinesList: jspb.Message.toObjectList(msg.getMachinesList(),
    proto.api.v1.Machine.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ListMachinesResponse}
 */
proto.api.v1.ListMachinesResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ListMachinesResponse;
  return proto.api.v1.ListMachinesResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ListMachinesResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ListMachinesResponse}
 */
proto.api.v1.ListMachinesResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.api.v1.Machine;
      reader.readMessage(value,proto.api.v1.Machine.deserializeBinaryFromReader);
      msg.addMachines(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ListMachinesResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ListMachinesResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ListMachinesResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListMachinesResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getMachinesList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.api.v1.Machine.serializeBinaryToWriter
    );
  }
};


/**
 * repeated Machine machines = 1;
 * @return {!Array<!proto.api.v1.Machine>}
 */
proto.api.v1.ListMachinesResponse.prototype.getMachinesList = function() {
  return /** @type{!Array<!proto.api.v1.Machine>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.api.v1.Machine, 1));
};


/**
 * @param {!Array<!proto.api.v1.Machine>} value
 * @return {!proto.api.v1.ListMachinesResponse} returns this
*/
proto.api.v1.ListMachinesResponse.prototype.setMachinesList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.api.v1.Machine=} opt_value
 * @param {number=} opt_index
 * @return {!proto.api.v1.Machine}
 */
proto.api.v1.ListMachinesResponse.prototype.addMachines = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.api.v1.Machine, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api.v1.ListMachinesResponse} returns this
 */
proto.api.v1.ListMachinesResponse.prototype.clearMachinesList = function() {
  return this.setMachinesList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.GetMachineRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.GetMachineRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.GetMachineRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetMachineRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    machineId: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.GetMachineRequest}
 */
proto.api.v1.GetMachineRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.GetMachineRequest;
  return proto.api.v1.GetMachineRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.GetMachineRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.GetMachineRequest}
 */
proto.api.v1.GetMachineRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setMachineId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.GetMachineRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.GetMachineRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.GetMachineRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetMachineRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getMachineId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.GetMachineRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.GetMachineRequest} returns this
 */
proto.api.v1.GetMachineRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string machine_id = 2;
 * @return {string}
 */
proto.api.v1.GetMachineRequest.prototype.getMachineId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.GetMachineRequest} returns this
 */
proto.api.v1.GetMachineRequest.prototype.setMachineId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.UpdateMachinePingRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.UpdateMachinePingRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.UpdateMachinePingRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.UpdateMachinePingRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    machineId: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.UpdateMachinePingRequest}
 */
proto.api.v1.UpdateMachinePingRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.UpdateMachinePingRequest;
  return proto.api.v1.UpdateMachinePingRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.UpdateMachinePingRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.UpdateMachinePingRequest}
 */
proto.api.v1.UpdateMachinePingRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setMachineId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.UpdateMachinePingRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.UpdateMachinePingRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.UpdateMachinePingRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.UpdateMachinePingRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getMachineId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.UpdateMachinePingRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.UpdateMachinePingRequest} returns this
 */
proto.api.v1.UpdateMachinePingRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string machine_id = 2;
 * @return {string}
 */
proto.api.v1.UpdateMachinePingRequest.prototype.getMachineId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.UpdateMachinePingRequest} returns this
 */
proto.api.v1.UpdateMachinePingRequest.prototype.setMachineId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.UnregisterMachineRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.UnregisterMachineRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.UnregisterMachineRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.UnregisterMachineRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    machineId: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.UnregisterMachineRequest}
 */
proto.api.v1.UnregisterMachineRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.UnregisterMachineRequest;
  return proto.api.v1.UnregisterMachineRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.UnregisterMachineRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.UnregisterMachineRequest}
 */
proto.api.v1.UnregisterMachineRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setMachineId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.UnregisterMachineRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.UnregisterMachineRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.UnregisterMachineRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.UnregisterMachineRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getMachineId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.UnregisterMachineRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.UnregisterMachineRequest} returns this
 */
proto.api.v1.UnregisterMachineRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string machine_id = 2;
 * @return {string}
 */
proto.api.v1.UnregisterMachineRequest.prototype.getMachineId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.UnregisterMachineRequest} returns this
 */
proto.api.v1.UnregisterMachineRequest.prototype.setMachineId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.UnregisterMachineResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.UnregisterMachineResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.UnregisterMachineResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.UnregisterMachineResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    success: jspb.Message.getBooleanFieldWithDefault(msg, 1, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.UnregisterMachineResponse}
 */
proto.api.v1.UnregisterMachineResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.UnregisterMachineResponse;
  return proto.api.v1.UnregisterMachineResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.UnregisterMachineResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.UnregisterMachineResponse}
 */
proto.api.v1.UnregisterMachineResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setSuccess(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.UnregisterMachineResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.UnregisterMachineResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.UnregisterMachineResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.UnregisterMachineResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSuccess();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
};


/**
 * optional bool success = 1;
 * @return {boolean}
 */
proto.api.v1.UnregisterMachineResponse.prototype.getSuccess = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.api.v1.UnregisterMachineResponse} returns this
 */
proto.api.v1.UnregisterMachineResponse.prototype.setSuccess = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ExecuteToolRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ExecuteToolRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ExecuteToolRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ExecuteToolRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    toolName: jspb.Message.getFieldWithDefault(msg, 2, ""),
    input: jspb.Message.getFieldWithDefault(msg, 3, ""),
    timeoutSeconds: jspb.Message.getFieldWithDefault(msg, 4, 0),
    idempotencyKey: jspb.Message.getFieldWithDefault(msg, 5, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ExecuteToolRequest}
 */
proto.api.v1.ExecuteToolRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ExecuteToolRequest;
  return proto.api.v1.ExecuteToolRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ExecuteToolRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ExecuteToolRequest}
 */
proto.api.v1.ExecuteToolRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setToolName(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setInput(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setTimeoutSeconds(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setIdempotencyKey(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ExecuteToolRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ExecuteToolRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ExecuteToolRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ExecuteToolRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getToolName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getInput();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getTimeoutSeconds();
  if (f !== 0) {
    writer.writeInt32(
      4,
      f
    );
  }
  f = message.getIdempotencyKey();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.ExecuteToolRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ExecuteToolRequest} returns this
 */
proto.api.v1.ExecuteToolRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string tool_name = 2;
 * @return {string}
 */
proto.api.v1.ExecuteToolRequest.prototype.getToolName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ExecuteToolRequest} returns this
 */
proto.api.v1.ExecuteToolRequest.prototype.setToolName = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string input = 3;
 * @return {string}
 */
proto.api.v1.ExecuteToolRequest.prototype.getInput = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ExecuteToolRequest} returns this
 */
proto.api.v1.ExecuteToolRequest.prototype.setInput = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional int32 timeout_seconds = 4;
 * @return {number}
 */
proto.api.v1.ExecuteToolRequest.prototype.getTimeoutSeconds = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.ExecuteToolRequest} returns this
 */
proto.api.v1.ExecuteToolRequest.prototype.setTimeoutSeconds = function(value) {
  return jspb.Message.setProto3IntField(this, 4, value);
};


/**
 * optional string idempotency_key = 5;
 * @return {string}
 */
proto.api.v1.ExecuteToolRequest.prototype.getIdempotencyKey = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ExecuteToolRequest} returns this
 */
proto.api.v1.ExecuteToolRequest.prototype.setIdempotencyKey = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ExecuteToolResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ExecuteToolResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ExecuteToolResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ExecuteToolResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    requestId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    status: jspb.Message.getFieldWithDefault(msg, 2, 0),
    result: jspb.Message.getFieldWithDefault(msg, 3, ""),
    resultType: jspb.Message.getFieldWithDefault(msg, 4, ""),
    error: jspb.Message.getFieldWithDefault(msg, 5, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ExecuteToolResponse}
 */
proto.api.v1.ExecuteToolResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ExecuteToolResponse;
  return proto.api.v1.ExecuteToolResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ExecuteToolResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ExecuteToolResponse}
 */
proto.api.v1.ExecuteToolResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setRequestId(value);
      break;
    case 2:
      var value = /** @type {!proto.api.v1.RequestStatus} */ (reader.readEnum());
      msg.setStatus(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setResult(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setResultType(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setError(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ExecuteToolResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ExecuteToolResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ExecuteToolResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ExecuteToolResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRequestId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getStatus();
  if (f !== 0.0) {
    writer.writeEnum(
      2,
      f
    );
  }
  f = message.getResult();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getResultType();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getError();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
};


/**
 * optional string request_id = 1;
 * @return {string}
 */
proto.api.v1.ExecuteToolResponse.prototype.getRequestId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ExecuteToolResponse} returns this
 */
proto.api.v1.ExecuteToolResponse.prototype.setRequestId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional RequestStatus status = 2;
 * @return {!proto.api.v1.RequestStatus}
 */
proto.api.v1.ExecuteToolResponse.prototype.getStatus = function() {
  return /** @type {!proto.api.v1.RequestStatus} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {!proto.api.v1.RequestStatus} value
 * @return {!proto.api.v1.ExecuteToolResponse} returns this
 */
proto.api.v1.ExecuteToolResponse.prototype.setStatus = function(value) {
  return jspb.Message.setProto3EnumField(this, 2, value);
};


/**
 * optional string result = 3;
 * @return {string}
 */
proto.api.v1.ExecuteToolResponse.prototype.getResult = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ExecuteToolResponse} returns this
 */
proto.api.v1.ExecuteToolResponse.prototype.setResult = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string result_type = 4;
 * @return {string}
 */
proto.api.v1.ExecuteToolResponse.prototype.getResultType = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ExecuteToolResponse} returns this
 */
proto.api.v1.ExecuteToolResponse.prototype.setResultType = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string error = 5;
 * @return {string}
 */
proto.api.v1.ExecuteToolResponse.prototype.getError = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ExecuteToolResponse} returns this
 */
proto.api.v1.ExecuteToolResponse.prototype.setError = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ExecuteToolChunk.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ExecuteToolChunk.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ExecuteToolChunk} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ExecuteToolChunk.toObject = function(includeInstance, msg) {
  var f, obj = {
    seq: jspb.Message.getFieldWithDefault(msg, 1, 0),
    requestId: jspb.Message.getFieldWithDefault(msg, 2, ""),
    chunk: jspb.Message.getFieldWithDefault(msg, 3, ""),
    isFinal: jspb.Message.getBooleanFieldWithDefault(msg, 4, false),
    error: jspb.Message.getFieldWithDefault(msg, 5, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ExecuteToolChunk}
 */
proto.api.v1.ExecuteToolChunk.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ExecuteToolChunk;
  return proto.api.v1.ExecuteToolChunk.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ExecuteToolChunk} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ExecuteToolChunk}
 */
proto.api.v1.ExecuteToolChunk.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setSeq(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setRequestId(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setChunk(value);
      break;
    case 4:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setIsFinal(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setError(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ExecuteToolChunk.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ExecuteToolChunk.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ExecuteToolChunk} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ExecuteToolChunk.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSeq();
  if (f !== 0) {
    writer.writeInt32(
      1,
      f
    );
  }
  f = message.getRequestId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getChunk();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getIsFinal();
  if (f) {
    writer.writeBool(
      4,
      f
    );
  }
  f = message.getError();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
};


/**
 * optional int32 seq = 1;
 * @return {number}
 */
proto.api.v1.ExecuteToolChunk.prototype.getSeq = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.ExecuteToolChunk} returns this
 */
proto.api.v1.ExecuteToolChunk.prototype.setSeq = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional string request_id = 2;
 * @return {string}
 */
proto.api.v1.ExecuteToolChunk.prototype.getRequestId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ExecuteToolChunk} returns this
 */
proto.api.v1.ExecuteToolChunk.prototype.setRequestId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string chunk = 3;
 * @return {string}
 */
proto.api.v1.ExecuteToolChunk.prototype.getChunk = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ExecuteToolChunk} returns this
 */
proto.api.v1.ExecuteToolChunk.prototype.setChunk = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional bool is_final = 4;
 * @return {boolean}
 */
proto.api.v1.ExecuteToolChunk.prototype.getIsFinal = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 4, false));
};


/**
 * @param {boolean} value
 * @return {!proto.api.v1.ExecuteToolChunk} returns this
 */
proto.api.v1.ExecuteToolChunk.prototype.setIsFinal = function(value) {
  return jspb.Message.setProto3BooleanField(this, 4, value);
};


/**
 * optional string error = 5;
 * @return {string}
 */
proto.api.v1.ExecuteToolChunk.prototype.getError = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ExecuteToolChunk} returns this
 */
proto.api.v1.ExecuteToolChunk.prototype.setError = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.CreateRequestRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.CreateRequestRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.CreateRequestRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.CreateRequestRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    toolName: jspb.Message.getFieldWithDefault(msg, 2, ""),
    input: jspb.Message.getFieldWithDefault(msg, 3, ""),
    timeoutSeconds: jspb.Message.getFieldWithDefault(msg, 4, 0),
    idempotencyKey: jspb.Message.getFieldWithDefault(msg, 5, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.CreateRequestRequest}
 */
proto.api.v1.CreateRequestRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.CreateRequestRequest;
  return proto.api.v1.CreateRequestRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.CreateRequestRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.CreateRequestRequest}
 */
proto.api.v1.CreateRequestRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setToolName(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setInput(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setTimeoutSeconds(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setIdempotencyKey(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.CreateRequestRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.CreateRequestRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.CreateRequestRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.CreateRequestRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getToolName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getInput();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getTimeoutSeconds();
  if (f !== 0) {
    writer.writeInt32(
      4,
      f
    );
  }
  f = message.getIdempotencyKey();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.CreateRequestRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.CreateRequestRequest} returns this
 */
proto.api.v1.CreateRequestRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string tool_name = 2;
 * @return {string}
 */
proto.api.v1.CreateRequestRequest.prototype.getToolName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.CreateRequestRequest} returns this
 */
proto.api.v1.CreateRequestRequest.prototype.setToolName = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string input = 3;
 * @return {string}
 */
proto.api.v1.CreateRequestRequest.prototype.getInput = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.CreateRequestRequest} returns this
 */
proto.api.v1.CreateRequestRequest.prototype.setInput = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional int32 timeout_seconds = 4;
 * @return {number}
 */
proto.api.v1.CreateRequestRequest.prototype.getTimeoutSeconds = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.CreateRequestRequest} returns this
 */
proto.api.v1.CreateRequestRequest.prototype.setTimeoutSeconds = function(value) {
  return jspb.Message.setProto3IntField(this, 4, value);
};


/**
 * optional string idempotency_key = 5;
 * @return {string}
 */
proto.api.v1.CreateRequestRequest.prototype.getIdempotencyKey = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.CreateRequestRequest} returns this
 */
proto.api.v1.CreateRequestRequest.prototype.setIdempotencyKey = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.GetRequestRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.GetRequestRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.GetRequestRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetRequestRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    requestId: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.GetRequestRequest}
 */
proto.api.v1.GetRequestRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.GetRequestRequest;
  return proto.api.v1.GetRequestRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.GetRequestRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.GetRequestRequest}
 */
proto.api.v1.GetRequestRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setRequestId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.GetRequestRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.GetRequestRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.GetRequestRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetRequestRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getRequestId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.GetRequestRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.GetRequestRequest} returns this
 */
proto.api.v1.GetRequestRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string request_id = 2;
 * @return {string}
 */
proto.api.v1.GetRequestRequest.prototype.getRequestId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.GetRequestRequest} returns this
 */
proto.api.v1.GetRequestRequest.prototype.setRequestId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ListRequestsRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ListRequestsRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ListRequestsRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListRequestsRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    status: jspb.Message.getFieldWithDefault(msg, 2, 0),
    toolName: jspb.Message.getFieldWithDefault(msg, 3, ""),
    pageSize: jspb.Message.getFieldWithDefault(msg, 4, 0),
    pageToken: jspb.Message.getFieldWithDefault(msg, 5, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ListRequestsRequest}
 */
proto.api.v1.ListRequestsRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ListRequestsRequest;
  return proto.api.v1.ListRequestsRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ListRequestsRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ListRequestsRequest}
 */
proto.api.v1.ListRequestsRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {!proto.api.v1.RequestStatus} */ (reader.readEnum());
      msg.setStatus(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setToolName(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setPageSize(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setPageToken(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ListRequestsRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ListRequestsRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ListRequestsRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListRequestsRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getStatus();
  if (f !== 0.0) {
    writer.writeEnum(
      2,
      f
    );
  }
  f = message.getToolName();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getPageSize();
  if (f !== 0) {
    writer.writeInt32(
      4,
      f
    );
  }
  f = message.getPageToken();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.ListRequestsRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ListRequestsRequest} returns this
 */
proto.api.v1.ListRequestsRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional RequestStatus status = 2;
 * @return {!proto.api.v1.RequestStatus}
 */
proto.api.v1.ListRequestsRequest.prototype.getStatus = function() {
  return /** @type {!proto.api.v1.RequestStatus} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {!proto.api.v1.RequestStatus} value
 * @return {!proto.api.v1.ListRequestsRequest} returns this
 */
proto.api.v1.ListRequestsRequest.prototype.setStatus = function(value) {
  return jspb.Message.setProto3EnumField(this, 2, value);
};


/**
 * optional string tool_name = 3;
 * @return {string}
 */
proto.api.v1.ListRequestsRequest.prototype.getToolName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ListRequestsRequest} returns this
 */
proto.api.v1.ListRequestsRequest.prototype.setToolName = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional int32 page_size = 4;
 * @return {number}
 */
proto.api.v1.ListRequestsRequest.prototype.getPageSize = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.ListRequestsRequest} returns this
 */
proto.api.v1.ListRequestsRequest.prototype.setPageSize = function(value) {
  return jspb.Message.setProto3IntField(this, 4, value);
};


/**
 * optional string page_token = 5;
 * @return {string}
 */
proto.api.v1.ListRequestsRequest.prototype.getPageToken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ListRequestsRequest} returns this
 */
proto.api.v1.ListRequestsRequest.prototype.setPageToken = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api.v1.ListRequestsResponse.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ListRequestsResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ListRequestsResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ListRequestsResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListRequestsResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    requestsList: jspb.Message.toObjectList(msg.getRequestsList(),
    proto.api.v1.Request.toObject, includeInstance),
    page: (f = msg.getPage()) && proto.api.v1.ListPage.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ListRequestsResponse}
 */
proto.api.v1.ListRequestsResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ListRequestsResponse;
  return proto.api.v1.ListRequestsResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ListRequestsResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ListRequestsResponse}
 */
proto.api.v1.ListRequestsResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.api.v1.Request;
      reader.readMessage(value,proto.api.v1.Request.deserializeBinaryFromReader);
      msg.addRequests(value);
      break;
    case 2:
      var value = new proto.api.v1.ListPage;
      reader.readMessage(value,proto.api.v1.ListPage.deserializeBinaryFromReader);
      msg.setPage(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ListRequestsResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ListRequestsResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ListRequestsResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListRequestsResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRequestsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.api.v1.Request.serializeBinaryToWriter
    );
  }
  f = message.getPage();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.api.v1.ListPage.serializeBinaryToWriter
    );
  }
};


/**
 * repeated Request requests = 1;
 * @return {!Array<!proto.api.v1.Request>}
 */
proto.api.v1.ListRequestsResponse.prototype.getRequestsList = function() {
  return /** @type{!Array<!proto.api.v1.Request>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.api.v1.Request, 1));
};


/**
 * @param {!Array<!proto.api.v1.Request>} value
 * @return {!proto.api.v1.ListRequestsResponse} returns this
*/
proto.api.v1.ListRequestsResponse.prototype.setRequestsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.api.v1.Request=} opt_value
 * @param {number=} opt_index
 * @return {!proto.api.v1.Request}
 */
proto.api.v1.ListRequestsResponse.prototype.addRequests = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.api.v1.Request, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api.v1.ListRequestsResponse} returns this
 */
proto.api.v1.ListRequestsResponse.prototype.clearRequestsList = function() {
  return this.setRequestsList([]);
};


/**
 * optional ListPage page = 2;
 * @return {?proto.api.v1.ListPage}
 */
proto.api.v1.ListRequestsResponse.prototype.getPage = function() {
  return /** @type{?proto.api.v1.ListPage} */ (
    jspb.Message.getWrapperField(this, proto.api.v1.ListPage, 2));
};


/**
 * @param {?proto.api.v1.ListPage|undefined} value
 * @return {!proto.api.v1.ListRequestsResponse} returns this
*/
proto.api.v1.ListRequestsResponse.prototype.setPage = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api.v1.ListRequestsResponse} returns this
 */
proto.api.v1.ListRequestsResponse.prototype.clearPage = function() {
  return this.setPage(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api.v1.ListRequestsResponse.prototype.hasPage = function() {
  return jspb.Message.getField(this, 2) != null;
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.UpdateRequestRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.UpdateRequestRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.UpdateRequestRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.UpdateRequestRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    requestId: jspb.Message.getFieldWithDefault(msg, 2, ""),
    status: jspb.Message.getFieldWithDefault(msg, 3, 0),
    result: jspb.Message.getFieldWithDefault(msg, 4, ""),
    resultType: jspb.Message.getFieldWithDefault(msg, 5, ""),
    machineId: jspb.Message.getFieldWithDefault(msg, 6, ""),
    leaseEpoch: jspb.Message.getFieldWithDefault(msg, 7, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.UpdateRequestRequest}
 */
proto.api.v1.UpdateRequestRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.UpdateRequestRequest;
  return proto.api.v1.UpdateRequestRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.UpdateRequestRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.UpdateRequestRequest}
 */
proto.api.v1.UpdateRequestRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setRequestId(value);
      break;
    case 3:
      var value = /** @type {!proto.api.v1.RequestStatus} */ (reader.readEnum());
      msg.setStatus(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setResult(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setResultType(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setMachineId(value);
      break;
    case 7:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setLeaseEpoch(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.UpdateRequestRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.UpdateRequestRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.UpdateRequestRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.UpdateRequestRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getRequestId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getStatus();
  if (f !== 0.0) {
    writer.writeEnum(
      3,
      f
    );
  }
  f = message.getResult();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getResultType();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getMachineId();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getLeaseEpoch();
  if (f !== 0) {
    writer.writeInt64(
      7,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.UpdateRequestRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.UpdateRequestRequest} returns this
 */
proto.api.v1.UpdateRequestRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string request_id = 2;
 * @return {string}
 */
proto.api.v1.UpdateRequestRequest.prototype.getRequestId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.UpdateRequestRequest} returns this
 */
proto.api.v1.UpdateRequestRequest.prototype.setRequestId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional RequestStatus status = 3;
 * @return {!proto.api.v1.RequestStatus}
 */
proto.api.v1.UpdateRequestRequest.prototype.getStatus = function() {
  return /** @type {!proto.api.v1.RequestStatus} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {!proto.api.v1.RequestStatus} value
 * @return {!proto.api.v1.UpdateRequestRequest} returns this
 */
proto.api.v1.UpdateRequestRequest.prototype.setStatus = function(value) {
  return jspb.Message.setProto3EnumField(this, 3, value);
};


/**
 * optional string result = 4;
 * @return {string}
 */
proto.api.v1.UpdateRequestRequest.prototype.getResult = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.UpdateRequestRequest} returns this
 */
proto.api.v1.UpdateRequestRequest.prototype.setResult = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string result_type = 5;
 * @return {string}
 */
proto.api.v1.UpdateRequestRequest.prototype.getResultType = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.UpdateRequestRequest} returns this
 */
proto.api.v1.UpdateRequestRequest.prototype.setResultType = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional string machine_id = 6;
 * @return {string}
 */
proto.api.v1.UpdateRequestRequest.prototype.getMachineId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.UpdateRequestRequest} returns this
 */
proto.api.v1.UpdateRequestRequest.prototype.setMachineId = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * optional int64 lease_epoch = 7;
 * @return {number}
 */
proto.api.v1.UpdateRequestRequest.prototype.getLeaseEpoch = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.UpdateRequestRequest} returns this
 */
proto.api.v1.UpdateRequestRequest.prototype.setLeaseEpoch = function(value) {
  return jspb.Message.setProto3IntField(this, 7, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ClaimRequestRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ClaimRequestRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ClaimRequestRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ClaimRequestRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    requestId: jspb.Message.getFieldWithDefault(msg, 2, ""),
    machineId: jspb.Message.getFieldWithDefault(msg, 3, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ClaimRequestRequest}
 */
proto.api.v1.ClaimRequestRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ClaimRequestRequest;
  return proto.api.v1.ClaimRequestRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ClaimRequestRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ClaimRequestRequest}
 */
proto.api.v1.ClaimRequestRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setRequestId(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setMachineId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ClaimRequestRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ClaimRequestRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ClaimRequestRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ClaimRequestRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getRequestId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getMachineId();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.ClaimRequestRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ClaimRequestRequest} returns this
 */
proto.api.v1.ClaimRequestRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string request_id = 2;
 * @return {string}
 */
proto.api.v1.ClaimRequestRequest.prototype.getRequestId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ClaimRequestRequest} returns this
 */
proto.api.v1.ClaimRequestRequest.prototype.setRequestId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string machine_id = 3;
 * @return {string}
 */
proto.api.v1.ClaimRequestRequest.prototype.getMachineId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ClaimRequestRequest} returns this
 */
proto.api.v1.ClaimRequestRequest.prototype.setMachineId = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.CancelRequestRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.CancelRequestRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.CancelRequestRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.CancelRequestRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    requestId: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.CancelRequestRequest}
 */
proto.api.v1.CancelRequestRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.CancelRequestRequest;
  return proto.api.v1.CancelRequestRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.CancelRequestRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.CancelRequestRequest}
 */
proto.api.v1.CancelRequestRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setRequestId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.CancelRequestRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.CancelRequestRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.CancelRequestRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.CancelRequestRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getRequestId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.CancelRequestRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.CancelRequestRequest} returns this
 */
proto.api.v1.CancelRequestRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string request_id = 2;
 * @return {string}
 */
proto.api.v1.CancelRequestRequest.prototype.getRequestId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.CancelRequestRequest} returns this
 */
proto.api.v1.CancelRequestRequest.prototype.setRequestId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.CancelRequestResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.CancelRequestResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.CancelRequestResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.CancelRequestResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    success: jspb.Message.getBooleanFieldWithDefault(msg, 1, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.CancelRequestResponse}
 */
proto.api.v1.CancelRequestResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.CancelRequestResponse;
  return proto.api.v1.CancelRequestResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.CancelRequestResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.CancelRequestResponse}
 */
proto.api.v1.CancelRequestResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setSuccess(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.CancelRequestResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.CancelRequestResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.CancelRequestResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.CancelRequestResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSuccess();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
};


/**
 * optional bool success = 1;
 * @return {boolean}
 */
proto.api.v1.CancelRequestResponse.prototype.getSuccess = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.api.v1.CancelRequestResponse} returns this
 */
proto.api.v1.CancelRequestResponse.prototype.setSuccess = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.SubmitRequestResultRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.SubmitRequestResultRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.SubmitRequestResultRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.SubmitRequestResultRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    requestId: jspb.Message.getFieldWithDefault(msg, 2, ""),
    result: jspb.Message.getFieldWithDefault(msg, 3, ""),
    resultType: jspb.Message.getFieldWithDefault(msg, 4, ""),
    metaMap: (f = msg.getMetaMap()) ? f.toObject(includeInstance, undefined) : [],
    machineId: jspb.Message.getFieldWithDefault(msg, 6, ""),
    leaseEpoch: jspb.Message.getFieldWithDefault(msg, 7, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.SubmitRequestResultRequest}
 */
proto.api.v1.SubmitRequestResultRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.SubmitRequestResultRequest;
  return proto.api.v1.SubmitRequestResultRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.SubmitRequestResultRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.SubmitRequestResultRequest}
 */
proto.api.v1.SubmitRequestResultRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setRequestId(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setResult(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setResultType(value);
      break;
    case 5:
      var value = msg.getMetaMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString, null, "", "");
         });
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setMachineId(value);
      break;
    case 7:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setLeaseEpoch(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.SubmitRequestResultRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.SubmitRequestResultRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.SubmitRequestResultRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.SubmitRequestResultRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getRequestId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getResult();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getResultType();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getMetaMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(5, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
  f = message.getMachineId();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getLeaseEpoch();
  if (f !== 0) {
    writer.writeInt64(
      7,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.SubmitRequestResultRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.SubmitRequestResultRequest} returns this
 */
proto.api.v1.SubmitRequestResultRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string request_id = 2;
 * @return {string}
 */
proto.api.v1.SubmitRequestResultRequest.prototype.getRequestId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.SubmitRequestResultRequest} returns this
 */
proto.api.v1.SubmitRequestResultRequest.prototype.setRequestId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string result = 3;
 * @return {string}
 */
proto.api.v1.SubmitRequestResultRequest.prototype.getResult = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.SubmitRequestResultRequest} returns this
 */
proto.api.v1.SubmitRequestResultRequest.prototype.setResult = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string result_type = 4;
 * @return {string}
 */
proto.api.v1.SubmitRequestResultRequest.prototype.getResultType = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.SubmitRequestResultRequest} returns this
 */
proto.api.v1.SubmitRequestResultRequest.prototype.setResultType = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * map<string, string> meta = 5;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.api.v1.SubmitRequestResultRequest.prototype.getMetaMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 5, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.api.v1.SubmitRequestResultRequest} returns this
 */
proto.api.v1.SubmitRequestResultRequest.prototype.clearMetaMap = function() {
  this.getMetaMap().clear();
  return this;};


/**
 * optional string machine_id = 6;
 * @return {string}
 */
proto.api.v1.SubmitRequestResultRequest.prototype.getMachineId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.SubmitRequestResultRequest} returns this
 */
proto.api.v1.SubmitRequestResultRequest.prototype.setMachineId = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * optional int64 lease_epoch = 7;
 * @return {number}
 */
proto.api.v1.SubmitRequestResultRequest.prototype.getLeaseEpoch = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.SubmitRequestResultRequest} returns this
 */
proto.api.v1.SubmitRequestResultRequest.prototype.setLeaseEpoch = function(value) {
  return jspb.Message.setProto3IntField(this, 7, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.SubmitRequestResultResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.SubmitRequestResultResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.SubmitRequestResultResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.SubmitRequestResultResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    success: jspb.Message.getBooleanFieldWithDefault(msg, 1, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.SubmitRequestResultResponse}
 */
proto.api.v1.SubmitRequestResultResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.SubmitRequestResultResponse;
  return proto.api.v1.SubmitRequestResultResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.SubmitRequestResultResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.SubmitRequestResultResponse}
 */
proto.api.v1.SubmitRequestResultResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setSuccess(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.SubmitRequestResultResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.SubmitRequestResultResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.SubmitRequestResultResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.SubmitRequestResultResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSuccess();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
};


/**
 * optional bool success = 1;
 * @return {boolean}
 */
proto.api.v1.SubmitRequestResultResponse.prototype.getSuccess = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.api.v1.SubmitRequestResultResponse} returns this
 */
proto.api.v1.SubmitRequestResultResponse.prototype.setSuccess = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api.v1.AppendRequestChunksRequest.repeatedFields_ = [3];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.AppendRequestChunksRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.AppendRequestChunksRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.AppendRequestChunksRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.AppendRequestChunksRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    requestId: jspb.Message.getFieldWithDefault(msg, 2, ""),
    chunksList: (f = jspb.Message.getRepeatedField(msg, 3)) == null ? undefined : f,
    resultType: jspb.Message.getFieldWithDefault(msg, 4, ""),
    machineId: jspb.Message.getFieldWithDefault(msg, 5, ""),
    leaseEpoch: jspb.Message.getFieldWithDefault(msg, 6, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.AppendRequestChunksRequest}
 */
proto.api.v1.AppendRequestChunksRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.AppendRequestChunksRequest;
  return proto.api.v1.AppendRequestChunksRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.AppendRequestChunksRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.AppendRequestChunksRequest}
 */
proto.api.v1.AppendRequestChunksRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setRequestId(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.addChunks(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setResultType(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setMachineId(value);
      break;
    case 6:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setLeaseEpoch(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.AppendRequestChunksRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.AppendRequestChunksRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.AppendRequestChunksRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.AppendRequestChunksRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getRequestId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getChunksList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      3,
      f
    );
  }
  f = message.getResultType();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getMachineId();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getLeaseEpoch();
  if (f !== 0) {
    writer.writeInt64(
      6,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.AppendRequestChunksRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.AppendRequestChunksRequest} returns this
 */
proto.api.v1.AppendRequestChunksRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string request_id = 2;
 * @return {string}
 */
proto.api.v1.AppendRequestChunksRequest.prototype.getRequestId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.AppendRequestChunksRequest} returns this
 */
proto.api.v1.AppendRequestChunksRequest.prototype.setRequestId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * repeated string chunks = 3;
 * @return {!Array<string>}
 */
proto.api.v1.AppendRequestChunksRequest.prototype.getChunksList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 3));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.api.v1.AppendRequestChunksRequest} returns this
 */
proto.api.v1.AppendRequestChunksRequest.prototype.setChunksList = function(value) {
  return jspb.Message.setField(this, 3, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.api.v1.AppendRequestChunksRequest} returns this
 */
proto.api.v1.AppendRequestChunksRequest.prototype.addChunks = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 3, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api.v1.AppendRequestChunksRequest} returns this
 */
proto.api.v1.AppendRequestChunksRequest.prototype.clearChunksList = function() {
  return this.setChunksList([]);
};


/**
 * optional string result_type = 4;
 * @return {string}
 */
proto.api.v1.AppendRequestChunksRequest.prototype.getResultType = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.AppendRequestChunksRequest} returns this
 */
proto.api.v1.AppendRequestChunksRequest.prototype.setResultType = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string machine_id = 5;
 * @return {string}
 */
proto.api.v1.AppendRequestChunksRequest.prototype.getMachineId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.AppendRequestChunksRequest} returns this
 */
proto.api.v1.AppendRequestChunksRequest.prototype.setMachineId = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional int64 lease_epoch = 6;
 * @return {number}
 */
proto.api.v1.AppendRequestChunksRequest.prototype.getLeaseEpoch = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 6, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.AppendRequestChunksRequest} returns this
 */
proto.api.v1.AppendRequestChunksRequest.prototype.setLeaseEpoch = function(value) {
  return jspb.Message.setProto3IntField(this, 6, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.AppendRequestChunksResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.AppendRequestChunksResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.AppendRequestChunksResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.AppendRequestChunksResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    success: jspb.Message.getBooleanFieldWithDefault(msg, 1, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.AppendRequestChunksResponse}
 */
proto.api.v1.AppendRequestChunksResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.AppendRequestChunksResponse;
  return proto.api.v1.AppendRequestChunksResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.AppendRequestChunksResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.AppendRequestChunksResponse}
 */
proto.api.v1.AppendRequestChunksResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setSuccess(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.AppendRequestChunksResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.AppendRequestChunksResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.AppendRequestChunksResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.AppendRequestChunksResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSuccess();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
};


/**
 * optional bool success = 1;
 * @return {boolean}
 */
proto.api.v1.AppendRequestChunksResponse.prototype.getSuccess = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.api.v1.AppendRequestChunksResponse} returns this
 */
proto.api.v1.AppendRequestChunksResponse.prototype.setSuccess = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.GetRequestChunksRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.GetRequestChunksRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.GetRequestChunksRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetRequestChunksRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    requestId: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.GetRequestChunksRequest}
 */
proto.api.v1.GetRequestChunksRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.GetRequestChunksRequest;
  return proto.api.v1.GetRequestChunksRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.GetRequestChunksRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.GetRequestChunksRequest}
 */
proto.api.v1.GetRequestChunksRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setRequestId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.GetRequestChunksRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.GetRequestChunksRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.GetRequestChunksRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetRequestChunksRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getRequestId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.GetRequestChunksRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.GetRequestChunksRequest} returns this
 */
proto.api.v1.GetRequestChunksRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string request_id = 2;
 * @return {string}
 */
proto.api.v1.GetRequestChunksRequest.prototype.getRequestId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.GetRequestChunksRequest} returns this
 */
proto.api.v1.GetRequestChunksRequest.prototype.setRequestId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api.v1.GetRequestChunksResponse.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.GetRequestChunksResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.GetRequestChunksResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.GetRequestChunksResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetRequestChunksResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    chunksList: (f = jspb.Message.getRepeatedField(msg, 1)) == null ? undefined : f,
    startSeq: jspb.Message.getFieldWithDefault(msg, 2, 0),
    nextSeq: jspb.Message.getFieldWithDefault(msg, 3, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.GetRequestChunksResponse}
 */
proto.api.v1.GetRequestChunksResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.GetRequestChunksResponse;
  return proto.api.v1.GetRequestChunksResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.GetRequestChunksResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.GetRequestChunksResponse}
 */
proto.api.v1.GetRequestChunksResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.addChunks(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setStartSeq(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setNextSeq(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.GetRequestChunksResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.GetRequestChunksResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.GetRequestChunksResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetRequestChunksResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getChunksList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      1,
      f
    );
  }
  f = message.getStartSeq();
  if (f !== 0) {
    writer.writeInt32(
      2,
      f
    );
  }
  f = message.getNextSeq();
  if (f !== 0) {
    writer.writeInt32(
      3,
      f
    );
  }
};


/**
 * repeated string chunks = 1;
 * @return {!Array<string>}
 */
proto.api.v1.GetRequestChunksResponse.prototype.getChunksList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 1));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.api.v1.GetRequestChunksResponse} returns this
 */
proto.api.v1.GetRequestChunksResponse.prototype.setChunksList = function(value) {
  return jspb.Message.setField(this, 1, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.api.v1.GetRequestChunksResponse} returns this
 */
proto.api.v1.GetRequestChunksResponse.prototype.addChunks = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 1, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api.v1.GetRequestChunksResponse} returns this
 */
proto.api.v1.GetRequestChunksResponse.prototype.clearChunksList = function() {
  return this.setChunksList([]);
};


/**
 * optional int32 start_seq = 2;
 * @return {number}
 */
proto.api.v1.GetRequestChunksResponse.prototype.getStartSeq = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.GetRequestChunksResponse} returns this
 */
proto.api.v1.GetRequestChunksResponse.prototype.setStartSeq = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional int32 next_seq = 3;
 * @return {number}
 */
proto.api.v1.GetRequestChunksResponse.prototype.getNextSeq = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.GetRequestChunksResponse} returns this
 */
proto.api.v1.GetRequestChunksResponse.prototype.setNextSeq = function(value) {
  return jspb.Message.setProto3IntField(this, 3, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.RenewRequestLeaseRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.RenewRequestLeaseRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.RenewRequestLeaseRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.RenewRequestLeaseRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    requestId: jspb.Message.getFieldWithDefault(msg, 2, ""),
    machineId: jspb.Message.getFieldWithDefault(msg, 3, ""),
    leaseEpoch: jspb.Message.getFieldWithDefault(msg, 4, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.RenewRequestLeaseRequest}
 */
proto.api.v1.RenewRequestLeaseRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.RenewRequestLeaseRequest;
  return proto.api.v1.RenewRequestLeaseRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.RenewRequestLeaseRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.RenewRequestLeaseRequest}
 */
proto.api.v1.RenewRequestLeaseRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setRequestId(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setMachineId(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setLeaseEpoch(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.RenewRequestLeaseRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.RenewRequestLeaseRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.RenewRequestLeaseRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.RenewRequestLeaseRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getRequestId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getMachineId();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getLeaseEpoch();
  if (f !== 0) {
    writer.writeInt64(
      4,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.RenewRequestLeaseRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.RenewRequestLeaseRequest} returns this
 */
proto.api.v1.RenewRequestLeaseRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string request_id = 2;
 * @return {string}
 */
proto.api.v1.RenewRequestLeaseRequest.prototype.getRequestId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.RenewRequestLeaseRequest} returns this
 */
proto.api.v1.RenewRequestLeaseRequest.prototype.setRequestId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string machine_id = 3;
 * @return {string}
 */
proto.api.v1.RenewRequestLeaseRequest.prototype.getMachineId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.RenewRequestLeaseRequest} returns this
 */
proto.api.v1.RenewRequestLeaseRequest.prototype.setMachineId = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional int64 lease_epoch = 4;
 * @return {number}
 */
proto.api.v1.RenewRequestLeaseRequest.prototype.getLeaseEpoch = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.api.v1.RenewRequestLeaseRequest} returns this
 */
proto.api.v1.RenewRequestLeaseRequest.prototype.setLeaseEpoch = function(value) {
  return jspb.Message.setProto3IntField(this, 4, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.HealthCheckRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.HealthCheckRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.HealthCheckRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.HealthCheckRequest.toObject = function(includeInstance, msg) {
  var f, obj = {

  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.HealthCheckRequest}
 */
proto.api.v1.HealthCheckRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.HealthCheckRequest;
  return proto.api.v1.HealthCheckRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.HealthCheckRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.HealthCheckRequest}
 */
proto.api.v1.HealthCheckRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.HealthCheckRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.HealthCheckRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.HealthCheckRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.HealthCheckRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.HealthCheckResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.HealthCheckResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.HealthCheckResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.HealthCheckResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    status: jspb.Message.getFieldWithDefault(msg, 1, ""),
    version: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.HealthCheckResponse}
 */
proto.api.v1.HealthCheckResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.HealthCheckResponse;
  return proto.api.v1.HealthCheckResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.HealthCheckResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.HealthCheckResponse}
 */
proto.api.v1.HealthCheckResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setStatus(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setVersion(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.HealthCheckResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.HealthCheckResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.HealthCheckResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.HealthCheckResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getStatus();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getVersion();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string status = 1;
 * @return {string}
 */
proto.api.v1.HealthCheckResponse.prototype.getStatus = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.HealthCheckResponse} returns this
 */
proto.api.v1.HealthCheckResponse.prototype.setStatus = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string version = 2;
 * @return {string}
 */
proto.api.v1.HealthCheckResponse.prototype.getVersion = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.HealthCheckResponse} returns this
 */
proto.api.v1.HealthCheckResponse.prototype.setVersion = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.Task.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.Task.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.Task} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.Task.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, ""),
    sessionId: jspb.Message.getFieldWithDefault(msg, 2, ""),
    toolName: jspb.Message.getFieldWithDefault(msg, 3, ""),
    status: jspb.Message.getFieldWithDefault(msg, 4, 0),
    input: jspb.Message.getFieldWithDefault(msg, 5, ""),
    result: jspb.Message.getFieldWithDefault(msg, 6, ""),
    resultType: jspb.Message.getFieldWithDefault(msg, 7, ""),
    error: jspb.Message.getFieldWithDefault(msg, 8, ""),
    createdAt: (f = msg.getCreatedAt()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    updatedAt: (f = msg.getUpdatedAt()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    completedAt: (f = msg.getCompletedAt()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    currentRequestId: jspb.Message.getFieldWithDefault(msg, 12, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.Task}
 */
proto.api.v1.Task.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.Task;
  return proto.api.v1.Task.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.Task} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.Task}
 */
proto.api.v1.Task.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setToolName(value);
      break;
    case 4:
      var value = /** @type {!proto.api.v1.TaskStatus} */ (reader.readEnum());
      msg.setStatus(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setInput(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setResult(value);
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.setResultType(value);
      break;
    case 8:
      var value = /** @type {string} */ (reader.readString());
      msg.setError(value);
      break;
    case 9:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setCreatedAt(value);
      break;
    case 10:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setUpdatedAt(value);
      break;
    case 11:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setCompletedAt(value);
      break;
    case 12:
      var value = /** @type {string} */ (reader.readString());
      msg.setCurrentRequestId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.Task.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.Task.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.Task} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.Task.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getToolName();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getStatus();
  if (f !== 0.0) {
    writer.writeEnum(
      4,
      f
    );
  }
  f = message.getInput();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getResult();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getResultType();
  if (f.length > 0) {
    writer.writeString(
      7,
      f
    );
  }
  f = message.getError();
  if (f.length > 0) {
    writer.writeString(
      8,
      f
    );
  }
  f = message.getCreatedAt();
  if (f != null) {
    writer.writeMessage(
      9,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getUpdatedAt();
  if (f != null) {
    writer.writeMessage(
      10,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getCompletedAt();
  if (f != null) {
    writer.writeMessage(
      11,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getCurrentRequestId();
  if (f.length > 0) {
    writer.writeString(
      12,
      f
    );
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.api.v1.Task.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Task} returns this
 */
proto.api.v1.Task.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string session_id = 2;
 * @return {string}
 */
proto.api.v1.Task.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Task} returns this
 */
proto.api.v1.Task.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string tool_name = 3;
 * @return {string}
 */
proto.api.v1.Task.prototype.getToolName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Task} returns this
 */
proto.api.v1.Task.prototype.setToolName = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional TaskStatus status = 4;
 * @return {!proto.api.v1.TaskStatus}
 */
proto.api.v1.Task.prototype.getStatus = function() {
  return /** @type {!proto.api.v1.TaskStatus} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {!proto.api.v1.TaskStatus} value
 * @return {!proto.api.v1.Task} returns this
 */
proto.api.v1.Task.prototype.setStatus = function(value) {
  return jspb.Message.setProto3EnumField(this, 4, value);
};


/**
 * optional string input = 5;
 * @return {string}
 */
proto.api.v1.Task.prototype.getInput = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Task} returns this
 */
proto.api.v1.Task.prototype.setInput = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional string result = 6;
 * @return {string}
 */
proto.api.v1.Task.prototype.getResult = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Task} returns this
 */
proto.api.v1.Task.prototype.setResult = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * optional string result_type = 7;
 * @return {string}
 */
proto.api.v1.Task.prototype.getResultType = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 7, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Task} returns this
 */
proto.api.v1.Task.prototype.setResultType = function(value) {
  return jspb.Message.setProto3StringField(this, 7, value);
};


/**
 * optional string error = 8;
 * @return {string}
 */
proto.api.v1.Task.prototype.getError = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 8, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Task} returns this
 */
proto.api.v1.Task.prototype.setError = function(value) {
  return jspb.Message.setProto3StringField(this, 8, value);
};


/**
 * optional google.protobuf.Timestamp created_at = 9;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.api.v1.Task.prototype.getCreatedAt = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 9));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.api.v1.Task} returns this
*/
proto.api.v1.Task.prototype.setCreatedAt = function(value) {
  return jspb.Message.setWrapperField(this, 9, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api.v1.Task} returns this
 */
proto.api.v1.Task.prototype.clearCreatedAt = function() {
  return this.setCreatedAt(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api.v1.Task.prototype.hasCreatedAt = function() {
  return jspb.Message.getField(this, 9) != null;
};


/**
 * optional google.protobuf.Timestamp updated_at = 10;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.api.v1.Task.prototype.getUpdatedAt = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 10));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.api.v1.Task} returns this
*/
proto.api.v1.Task.prototype.setUpdatedAt = function(value) {
  return jspb.Message.setWrapperField(this, 10, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api.v1.Task} returns this
 */
proto.api.v1.Task.prototype.clearUpdatedAt = function() {
  return this.setUpdatedAt(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api.v1.Task.prototype.hasUpdatedAt = function() {
  return jspb.Message.getField(this, 10) != null;
};


/**
 * optional google.protobuf.Timestamp completed_at = 11;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.api.v1.Task.prototype.getCompletedAt = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 11));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.api.v1.Task} returns this
*/
proto.api.v1.Task.prototype.setCompletedAt = function(value) {
  return jspb.Message.setWrapperField(this, 11, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api.v1.Task} returns this
 */
proto.api.v1.Task.prototype.clearCompletedAt = function() {
  return this.setCompletedAt(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api.v1.Task.prototype.hasCompletedAt = function() {
  return jspb.Message.getField(this, 11) != null;
};


/**
 * optional string current_request_id = 12;
 * @return {string}
 */
proto.api.v1.Task.prototype.getCurrentRequestId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 12, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.Task} returns this
 */
proto.api.v1.Task.prototype.setCurrentRequestId = function(value) {
  return jspb.Message.setProto3StringField(this, 12, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.CreateTaskRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.CreateTaskRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.CreateTaskRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.CreateTaskRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    toolName: jspb.Message.getFieldWithDefault(msg, 2, ""),
    input: jspb.Message.getFieldWithDefault(msg, 3, ""),
    idempotencyKey: jspb.Message.getFieldWithDefault(msg, 4, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.CreateTaskRequest}
 */
proto.api.v1.CreateTaskRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.CreateTaskRequest;
  return proto.api.v1.CreateTaskRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.CreateTaskRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.CreateTaskRequest}
 */
proto.api.v1.CreateTaskRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setToolName(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setInput(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setIdempotencyKey(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.CreateTaskRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.CreateTaskRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.CreateTaskRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.CreateTaskRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getToolName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getInput();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getIdempotencyKey();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.CreateTaskRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.CreateTaskRequest} returns this
 */
proto.api.v1.CreateTaskRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string tool_name = 2;
 * @return {string}
 */
proto.api.v1.CreateTaskRequest.prototype.getToolName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.CreateTaskRequest} returns this
 */
proto.api.v1.CreateTaskRequest.prototype.setToolName = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string input = 3;
 * @return {string}
 */
proto.api.v1.CreateTaskRequest.prototype.getInput = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.CreateTaskRequest} returns this
 */
proto.api.v1.CreateTaskRequest.prototype.setInput = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string idempotency_key = 4;
 * @return {string}
 */
proto.api.v1.CreateTaskRequest.prototype.getIdempotencyKey = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.CreateTaskRequest} returns this
 */
proto.api.v1.CreateTaskRequest.prototype.setIdempotencyKey = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.GetTaskRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.GetTaskRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.GetTaskRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetTaskRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    taskId: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.GetTaskRequest}
 */
proto.api.v1.GetTaskRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.GetTaskRequest;
  return proto.api.v1.GetTaskRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.GetTaskRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.GetTaskRequest}
 */
proto.api.v1.GetTaskRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setTaskId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.GetTaskRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.GetTaskRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.GetTaskRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.GetTaskRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getTaskId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.GetTaskRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.GetTaskRequest} returns this
 */
proto.api.v1.GetTaskRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string task_id = 2;
 * @return {string}
 */
proto.api.v1.GetTaskRequest.prototype.getTaskId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.GetTaskRequest} returns this
 */
proto.api.v1.GetTaskRequest.prototype.setTaskId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ListTasksRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ListTasksRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ListTasksRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListTasksRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ListTasksRequest}
 */
proto.api.v1.ListTasksRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ListTasksRequest;
  return proto.api.v1.ListTasksRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ListTasksRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ListTasksRequest}
 */
proto.api.v1.ListTasksRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ListTasksRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ListTasksRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ListTasksRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListTasksRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.ListTasksRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.ListTasksRequest} returns this
 */
proto.api.v1.ListTasksRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api.v1.ListTasksResponse.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.ListTasksResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.ListTasksResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.ListTasksResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListTasksResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    tasksList: jspb.Message.toObjectList(msg.getTasksList(),
    proto.api.v1.Task.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.ListTasksResponse}
 */
proto.api.v1.ListTasksResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.ListTasksResponse;
  return proto.api.v1.ListTasksResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.ListTasksResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.ListTasksResponse}
 */
proto.api.v1.ListTasksResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.api.v1.Task;
      reader.readMessage(value,proto.api.v1.Task.deserializeBinaryFromReader);
      msg.addTasks(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.ListTasksResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.ListTasksResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.ListTasksResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.ListTasksResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTasksList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.api.v1.Task.serializeBinaryToWriter
    );
  }
};


/**
 * repeated Task tasks = 1;
 * @return {!Array<!proto.api.v1.Task>}
 */
proto.api.v1.ListTasksResponse.prototype.getTasksList = function() {
  return /** @type{!Array<!proto.api.v1.Task>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.api.v1.Task, 1));
};


/**
 * @param {!Array<!proto.api.v1.Task>} value
 * @return {!proto.api.v1.ListTasksResponse} returns this
*/
proto.api.v1.ListTasksResponse.prototype.setTasksList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.api.v1.Task=} opt_value
 * @param {number=} opt_index
 * @return {!proto.api.v1.Task}
 */
proto.api.v1.ListTasksResponse.prototype.addTasks = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.api.v1.Task, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api.v1.ListTasksResponse} returns this
 */
proto.api.v1.ListTasksResponse.prototype.clearTasksList = function() {
  return this.setTasksList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.CancelTaskRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.CancelTaskRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.CancelTaskRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.CancelTaskRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    sessionId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    taskId: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.CancelTaskRequest}
 */
proto.api.v1.CancelTaskRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.CancelTaskRequest;
  return proto.api.v1.CancelTaskRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.CancelTaskRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.CancelTaskRequest}
 */
proto.api.v1.CancelTaskRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSessionId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setTaskId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.CancelTaskRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.CancelTaskRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.CancelTaskRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.CancelTaskRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSessionId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getTaskId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string session_id = 1;
 * @return {string}
 */
proto.api.v1.CancelTaskRequest.prototype.getSessionId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.CancelTaskRequest} returns this
 */
proto.api.v1.CancelTaskRequest.prototype.setSessionId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string task_id = 2;
 * @return {string}
 */
proto.api.v1.CancelTaskRequest.prototype.getTaskId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api.v1.CancelTaskRequest} returns this
 */
proto.api.v1.CancelTaskRequest.prototype.setTaskId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api.v1.CancelTaskResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api.v1.CancelTaskResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api.v1.CancelTaskResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.CancelTaskResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    success: jspb.Message.getBooleanFieldWithDefault(msg, 1, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api.v1.CancelTaskResponse}
 */
proto.api.v1.CancelTaskResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api.v1.CancelTaskResponse;
  return proto.api.v1.CancelTaskResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api.v1.CancelTaskResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api.v1.CancelTaskResponse}
 */
proto.api.v1.CancelTaskResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setSuccess(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api.v1.CancelTaskResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api.v1.CancelTaskResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api.v1.CancelTaskResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api.v1.CancelTaskResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSuccess();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
};


/**
 * optional bool success = 1;
 * @return {boolean}
 */
proto.api.v1.CancelTaskResponse.prototype.getSuccess = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.api.v1.CancelTaskResponse} returns this
 */
proto.api.v1.CancelTaskResponse.prototype.setSuccess = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};


/**
 * @enum {number}
 */
proto.api.v1.RequestStatus = {
  REQUEST_STATUS_UNSPECIFIED: 0,
  REQUEST_STATUS_PENDING: 1,
  REQUEST_STATUS_CLAIMED: 2,
  REQUEST_STATUS_RUNNING: 3,
  REQUEST_STATUS_DONE: 4,
  REQUEST_STATUS_FAILED: 5
};

/**
 * @enum {number}
 */
proto.api.v1.TaskStatus = {
  TASK_STATUS_UNSPECIFIED: 0,
  TASK_STATUS_PENDING: 1,
  TASK_STATUS_RUNNING: 2,
  TASK_STATUS_COMPLETED: 3,
  TASK_STATUS_FAILED: 4,
  TASK_STATUS_CANCELLED: 5,
  TASK_STATUS_DEAD_LETTER: 6
};

goog.object.extend(exports, proto.api.v1);
