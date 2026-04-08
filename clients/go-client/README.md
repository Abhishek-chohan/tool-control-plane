# Go Toolplane Client

This Go package is a supported secondary SDK for Toolplane's durable remote tool-execution control plane. Use it when Go code needs the maintained gRPC control-plane flow for a Toolplane-backed remote tool, but not the full provider worker loop. The canonical contract is `server/proto/service.proto`.

## Decision Rule

Reach for Toolplane when one remote tool may outlive the caller, needs inspection after disconnect, needs explicit provider ownership or drain behavior, or is queue-backed enough that request lifecycle control matters. If the work is quick and same-lifecycle, direct tool calling is simpler.

For a concrete first-offload workload, think of one sandboxed code-execution worker. Go is a good fit for the consumer-side and operator-side access around that worker. It is not the maintained SDK for building the provider loop itself.

## When To Start Here

- Your Go code needs consumer-side or operator-side access around a Toolplane-backed remote tool.
- You want session, machine, tool, request, and task helpers in Go while keeping the server-owned runtime semantics intact.
- You do not need a maintained claim-and-submit provider worker loop in Go; for that path, use Python or TypeScript.

## Support Status

- Supported secondary SDK with a focused gRPC surface.
- The maintained path is the gRPC control-plane flow: session lifecycle, machine registration, tool ownership, request execution, and task lifecycle.
- This SDK does not ship a maintained provider runtime harness that claims requests, renews heartbeats, and submits results. Use Python's explicit `ProviderRuntime` for the maintained provider-mode story.

## Features

- gRPC connection and health-check helpers through `Connect()`, `Disconnect()`, and `Ping()`.
- Public wrappers for session, API-key, tool, machine, request, and task lifecycle RPCs.
- Generic `ExecuteTool()` and `StreamExecuteTool()` helpers that follow the real request lifecycle.
- Numeric convenience helpers `Add()`, `Subtract()`, `Multiply()`, and `Divide()` that call `ExecuteTool()` against named tools.
- Tool discovery helpers `ListTools()`, `GetToolByID()`, `GetToolByName()`, and `DeleteTool()`.
- Opt-in live integration coverage in `client/toolplane_client_integration_test.go`.

## Installation

```bash
cd clients/go-client
go mod tidy
```

## Canonical Flow

The canonical end-to-end path for Toolplane is to offload one painful remote tool first: register a provider, create a session, execute a request, stream or recover results, and drain the machine. The Python and TypeScript SDKs ship the maintained provider runtimes for that flow. The Go examples below are the narrower consumer-side and operator-side projection you use alongside that flow, not in place of the maintained provider runtime.

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "toolplane-go-client/client"
    pb "toolplane-go-client/proto"
)

func main() {
    grpcClient, err := client.NewToolplaneClient(
        client.ProtocolGRPC,
        "localhost",
        9001,
        "my-session",
        "my-user",
        "toolplane-conformance-fixture-key",
        client.WithGRPCTLS("../../server/deploy/reference/certs/ca.crt", "localhost"),
    )
    if err != nil {
        log.Fatal(err)
    }

    if err := grpcClient.Connect(); err != nil {
        log.Fatal(err)
    }
    defer grpcClient.Disconnect()

    session, err := grpcClient.CreateSession(
        "My Session",
        "Session created from the Go client",
        "development",
    )
    if err != nil {
        log.Fatal(err)
    }

    machine, err := grpcClient.RegisterMachine("", "1.0.0", []*pb.RegisterToolRequest{
        {
            Name:        "session_status",
            Description: "Report session context for a connected operator or automation client",
            Schema:      `{"type":"object","properties":{"requester":{"type":"string"}},"required":["requester"]}`,
            Config:      map[string]string{"version": "1.0"},
            Tags:        []string{"session", "status"},
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    tools, err := grpcClient.ListTools()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Created session %s on machine %s with %d tools\n", session.Id, machine.Id, len(tools))
}
```

## Public API Snapshot

### Client Creation

```go
NewToolplaneClient(protocol ClientProtocol, serverHost string, serverPort int, sessionID, userID, apiKey string, opts ...ClientOption) (*ToolplaneClient, error)
```

- `protocol` must be `client.ProtocolGRPC`; any other value returns an error.
- `serverHost` and `serverPort` identify the gRPC endpoint.
- `sessionID`, `userID`, and `apiKey` seed the client context used by outgoing RPCs.
- `opts` lets callers enable TLS without breaking existing plaintext callers. Use `client.WithGRPCTLS(caCertPath, serverName)` for the reference deployment's TLS endpoint.

For the reference deployment, set `TOOLPLANE_SERVER_PORT=9001`, `TOOLPLANE_USE_TLS=true`, `TOOLPLANE_TLS_CA_CERT_PATH=/path/to/ca.crt`, and `TOOLPLANE_TLS_SERVER_NAME=localhost` before running the maintained examples.

### Connectivity

```go
Connect() error
Disconnect() error
Ping() (string, error)
```

### Tool Execution

```go
ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (*pb.Request, error)
StreamExecuteTool(ctx context.Context, toolName string, params map[string]interface{}, onChunk func(*pb.ExecuteToolChunk) error) ([]*pb.ExecuteToolChunk, error)
Add(a, b float64) (float64, error)
Subtract(a, b float64) (float64, error)
Multiply(a, b float64) (float64, error)
Divide(a, b float64) (float64, error)
```

These methods use the live request lifecycle on the Go server. They require a registered machine-backed provider to claim requests and submit results.

### Lifecycle Wrappers

The client exposes public wrappers for:

- Sessions: `CreateSession`, `GetSession`, `ListSessions`, `UpdateSession`
- API keys: `CreateAPIKey(name string, capabilities ...string)`, `ListAPIKeys`, `RevokeAPIKey`
- Tools: `RegisterTool`, `ListTools`, `GetToolByID`, `GetToolByName`, `DeleteTool`
- Machines: `RegisterMachine`, `ListMachines`, `GetMachine`, `UnregisterMachine`, `DrainMachine`
- Requests: `CreateRequest`, `GetRequest`, `ListRequests`, `CancelRequest`
- Tasks: `CreateTask`, `GetTask`, `ListTasks`, `CancelTask`

`RegisterTool` binds tools to the current machine ID when one is registered. For brand-new sessions, either register a machine first or embed tool definitions directly in `RegisterMachine(...)`.

`CreateSession` uses the configured transport auth metadata and does not forward the legacy `CreateSession.api_key` request field. `CreateAPIKey` is the only maintained helper that returns secret material. `ListAPIKeys` returns redacted metadata with `Key == ""`, plus `KeyPreview` and `Capabilities`.

## Examples

```bash
go run ./examples/basic
go run ./examples/advanced
go run client.go
```

## Testing

```bash
go test ./...
```

The default Go test command runs the fast unit suite. The live provider-backed gRPC test stays opt-in:

```bash
TOOLPLANE_RUN_INTEGRATION_TESTS=1 go test ./client -run TestGRPCLiveExecuteAndStreamExecuteTool -count=1
```

## Notes

- gRPC requests require outgoing auth metadata. Set `TOOLPLANE_API_KEY` when using the Go server.
- Unary and streaming execution now follow the real request queue. See `client/toolplane_client_integration_test.go` for a complete provider-backed example.

## Compatibility

- Go version: 1.23+
- Maintained server path: the Go server gRPC control-plane surface
- Public protocol: gRPC with Protocol Buffers
