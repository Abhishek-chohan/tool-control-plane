package service

// Model-to-proto conversion lives in converter_gen.go.
// The checked-in converter table provides the private lowercase helpers
// (convertModelToolToProto, convertModelSessionToProto, etc.) used by
// the gRPC handlers in server.go.  Public exported wrappers were removed
// in Tier 2 Item 9 because they had zero call sites outside this package.
