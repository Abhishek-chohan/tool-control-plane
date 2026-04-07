package service

import (
	"toolplane/pkg/model"
	proto "toolplane/proto"
)

// Model-to-proto conversion lives in converter_gen.go.
// The checked-in converter table provides the private lowercase helpers
// (convertModelToolToProto, convertModelSessionToProto, etc.) used by
// the gRPC handlers in server.go.  Public exported wrappers were removed
// in Tier 2 Item 9 because they had zero call sites outside this package.

func convertPublicSessionToProto(in *model.Session) *proto.Session {
	out := convertModelSessionToProto(in)
	if out == nil {
		return nil
	}
	out.ApiKey = ""
	return out
}

func convertPublicAPIKeyToProto(in *model.ApiKey, includeSecret bool) *proto.ApiKey {
	out := convertModelApiKeyToProto(in)
	if out == nil {
		return nil
	}
	if !includeSecret {
		out.Key = ""
	}
	out.Capabilities = model.CapabilityStrings(in.Capabilities)
	out.KeyPreview = in.KeyPreview
	return out
}
