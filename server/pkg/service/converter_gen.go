// Hand-maintained model<->proto converters. Keep in sync with
// server/proto/service.proto manually; there is no generator for this file
// (see server/docs/proto_regeneration.md). The compiler enforces field
// existence, but new proto fields must be added here by hand.

package service

import (
	"encoding/json"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"toolplane/pkg/model"
	proto "toolplane/proto"
)

// timestampProto converts a model time into its proto representation; the
// zero time maps to nil (field absent on the wire).
// timestampProtoFromPtr converts an optional model time; nil maps to nil.
func timestampProtoFromPtr(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestampProto(*t)
}

func timestampProto(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

// protoRequestStatus maps the model's string status onto the v1 enum.
func protoRequestStatus(s model.RequestStatus) proto.RequestStatus {
	switch s {
	case model.RequestStatusPending:
		return proto.RequestStatus_REQUEST_STATUS_PENDING
	case model.RequestStatusClaimed:
		return proto.RequestStatus_REQUEST_STATUS_CLAIMED
	case model.RequestStatusRunning:
		return proto.RequestStatus_REQUEST_STATUS_RUNNING
	case model.RequestStatusDone:
		return proto.RequestStatus_REQUEST_STATUS_DONE
	case model.RequestStatusFailed:
		return proto.RequestStatus_REQUEST_STATUS_FAILED
	default:
		return proto.RequestStatus_REQUEST_STATUS_UNSPECIFIED
	}
}

// protoTaskStatus maps the model's string status onto the v1 enum.
func protoTaskStatus(s model.TaskStatus) proto.TaskStatus {
	switch s {
	case model.StatusPending:
		return proto.TaskStatus_TASK_STATUS_PENDING
	case model.StatusRunning:
		return proto.TaskStatus_TASK_STATUS_RUNNING
	case model.StatusCompleted:
		return proto.TaskStatus_TASK_STATUS_COMPLETED
	case model.StatusFailed:
		return proto.TaskStatus_TASK_STATUS_FAILED
	case model.StatusCancelled:
		return proto.TaskStatus_TASK_STATUS_CANCELLED
	default:
		return proto.TaskStatus_TASK_STATUS_UNSPECIFIED
	}
}

func convertModelToolToProto(in *model.Tool) *proto.Tool {
	if in == nil {
		return nil
	}
	// Convert config
	configMap := make(map[string]string)
	for k, v := range in.Config {
		switch val := v.(type) {
		case string:
			configMap[k] = val
		default:
			if bytes, err := json.Marshal(val); err == nil {
				configMap[k] = string(bytes)
			}
		}
	}
	out := &proto.Tool{
		Id:          in.ID,
		Name:        in.Name,
		Description: in.Description,
		Schema:      in.Schema,
		Config:      configMap,
		CreatedAt:   timestampProto(in.CreatedAt),
		LastPingAt:  timestampProto(in.LastPingAt),
		SessionId:   in.SessionID,
		Tags:        in.Tags,
	}
	return out
}

func convertModelSessionToProto(in *model.Session) *proto.Session {
	if in == nil {
		return nil
	}
	out := &proto.Session{
		Id:          in.ID,
		Name:        in.Name,
		Description: in.Description,
		CreatedAt:   timestampProto(in.CreatedAt),
		CreatedBy:   in.CreatedBy,
		Namespace:   in.Namespace,
	}
	return out
}

func convertModelApiKeyToProto(in *model.ApiKey) *proto.ApiKey {
	if in == nil {
		return nil
	}
	revokedAt := timestampProtoFromPtr(in.RevokedAt)
	out := &proto.ApiKey{
		Id:        in.ID,
		Name:      in.Name,
		Key:       in.Key,
		SessionId: in.SessionID,
		CreatedAt: timestampProto(in.CreatedAt),
		CreatedBy: in.CreatedBy,
		RevokedAt: revokedAt,
	}
	return out
}

func convertModelMachineToProto(in *model.Machine) *proto.Machine {
	if in == nil {
		return nil
	}
	out := &proto.Machine{
		Id:          in.ID,
		SessionId:   in.SessionID,
		SdkVersion:  in.SDKVersion,
		SdkLanguage: in.SDKLanguage,
		Ip:          in.IP,
		CreatedAt:   timestampProto(in.CreatedAt),
		LastPingAt:  timestampProto(in.LastPingAt),
	}
	return out
}

func convertModelRequestToProto(in *model.Request) *proto.Request {
	if in == nil {
		return nil
	}
	// Convert result
	resultStr := ""
	if in.Result != nil {
		if bytes, err := json.Marshal(in.Result); err == nil {
			resultStr = string(bytes)
		}
	}
	var leaseExpiresAt *timestamppb.Timestamp
	if !in.VisibleAt.IsZero() && (in.Status == model.RequestStatusClaimed || in.Status == model.RequestStatusRunning) {
		leaseExpiresAt = timestampProto(in.VisibleAt)
	}
	out := &proto.Request{
		Id:                 in.ID,
		SessionId:          in.SessionID,
		ToolName:           in.ToolName,
		Status:             protoRequestStatus(in.Status),
		Input:              in.Input,
		Result:             resultStr,
		ResultType:         string(in.ResultType),
		Error:              in.Error,
		CreatedAt:          timestampProto(in.CreatedAt),
		UpdatedAt:          timestampProto(in.UpdatedAt),
		ExecutingMachineId: in.ExecutingMachineID,
		LeasedBy:           in.LeasedBy,
		LeaseEpoch:         in.LeaseEpoch,
		LeaseExpiresAt:     leaseExpiresAt,
		TimeoutSeconds:     int32(in.TimeoutSeconds),
	}
	return out
}

func convertModelTaskToProto(in *model.Task) *proto.Task {
	if in == nil {
		return nil
	}
	completedAt := timestampProtoFromPtr(in.CompletedAt)
	out := &proto.Task{
		Id:               in.ID,
		SessionId:        in.SessionID,
		ToolName:         in.ToolName,
		Status:           protoTaskStatus(in.Status),
		Input:            in.Input,
		Result:           in.Result,
		ResultType:       in.ResultType,
		Error:            in.Error,
		CreatedAt:        timestampProto(in.CreatedAt),
		UpdatedAt:        timestampProto(in.UpdatedAt),
		CompletedAt:      completedAt,
		CurrentRequestId: in.CurrentRequestID,
	}
	return out
}

// requestStatusFromProto maps the v1 enum onto the model's string status.
func requestStatusFromProto(s proto.RequestStatus) model.RequestStatus {
	switch s {
	case proto.RequestStatus_REQUEST_STATUS_PENDING:
		return model.RequestStatusPending
	case proto.RequestStatus_REQUEST_STATUS_CLAIMED:
		return model.RequestStatusClaimed
	case proto.RequestStatus_REQUEST_STATUS_RUNNING:
		return model.RequestStatusRunning
	case proto.RequestStatus_REQUEST_STATUS_DONE:
		return model.RequestStatusDone
	case proto.RequestStatus_REQUEST_STATUS_FAILED:
		return model.RequestStatusFailed
	default:
		return ""
	}
}
