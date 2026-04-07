package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "toolplane-go-client/proto"
)

const (
	defaultGRPCConnectTimeout   = 10 * time.Second
	defaultGRPCCallTimeout      = 5 * time.Second
	defaultGRPCExecutionTimeout = 30 * time.Second
	defaultRequestPollInterval  = 100 * time.Millisecond
)

// ClientProtocol defines the protocol type
type ClientProtocol string

const (
	ProtocolGRPC ClientProtocol = "grpc"
)

// ToolplaneClient represents the main client interface
type ToolplaneClient struct {
	protocol   ClientProtocol
	serverHost string
	serverPort int
	sessionID  string
	machineID  string
	userID     string
	apiKey     string

	// gRPC client fields
	grpcConn       *grpc.ClientConn
	toolClient     pb.ToolServiceClient
	sessionClient  pb.SessionsServiceClient
	machineClient  pb.MachinesServiceClient
	requestsClient pb.RequestsServiceClient
	tasksClient    pb.TasksServiceClient
}

// NewToolplaneClient creates a new Toolplane client
func NewToolplaneClient(protocol ClientProtocol, serverHost string, serverPort int, sessionID, userID, apiKey string) (*ToolplaneClient, error) {
	client := &ToolplaneClient{
		protocol:   protocol,
		serverHost: serverHost,
		serverPort: serverPort,
		sessionID:  sessionID,
		userID:     userID,
		apiKey:     apiKey,
	}

	if protocol != ProtocolGRPC {
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}

	return client, nil
}

// Connect establishes connection to the server
func (c *ToolplaneClient) Connect() error {
	if c.protocol != ProtocolGRPC {
		return fmt.Errorf("unsupported protocol: %s", c.protocol)
	}

	return c.connectGRPC()
}

// Disconnect closes the connection
func (c *ToolplaneClient) Disconnect() error {
	if c.protocol != ProtocolGRPC {
		return nil
	}

	if c.grpcConn == nil {
		return nil
	}

	var disconnectErr error
	if c.machineClient != nil && c.sessionID != "" && c.machineID != "" {
		ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
		defer cancel()

		_, err := c.machineClient.UnregisterMachine(ctx, &pb.UnregisterMachineRequest{
			SessionId: c.sessionID,
			MachineId: c.machineID,
		})
		if err != nil {
			disconnectErr = fmt.Errorf("failed to unregister machine %s: %w", c.machineID, err)
		}
	}

	if err := c.grpcConn.Close(); err != nil && disconnectErr == nil {
		disconnectErr = fmt.Errorf("failed to close gRPC connection: %w", err)
	}

	c.grpcConn = nil
	c.toolClient = nil
	c.sessionClient = nil
	c.machineClient = nil
	c.requestsClient = nil
	c.tasksClient = nil
	c.machineID = ""

	return disconnectErr
}

// connectGRPC establishes a real gRPC connection.
func (c *ToolplaneClient) connectGRPC() error {
	address := fmt.Sprintf("%s:%d", c.serverHost, c.serverPort)

	ctx, cancel := context.WithTimeout(context.Background(), defaultGRPCConnectTimeout)
	defer cancel()

	conn, err := dialGRPC(ctx, address)
	if err != nil {
		return fmt.Errorf("failed to dial gRPC server: %w", err)
	}

	c.grpcConn = conn
	c.toolClient = pb.NewToolServiceClient(conn)
	c.sessionClient = pb.NewSessionsServiceClient(conn)
	c.machineClient = pb.NewMachinesServiceClient(conn)
	c.requestsClient = pb.NewRequestsServiceClient(conn)
	c.tasksClient = pb.NewTasksServiceClient(conn)

	healthCtx, healthCancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer healthCancel()

	response, err := c.toolClient.HealthCheck(healthCtx, &pb.HealthCheckRequest{})
	if err != nil {
		_ = conn.Close()
		c.grpcConn = nil
		c.toolClient = nil
		c.sessionClient = nil
		c.machineClient = nil
		c.requestsClient = nil
		c.tasksClient = nil
		return fmt.Errorf("gRPC health check failed: %w", err)
	}

	log.Printf("gRPC connection established to %s (status=%s)", address, response.Status)
	return nil
}

func dialGRPC(ctx context.Context, address string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	conn.Connect()
	for {
		state := conn.GetState()
		switch state {
		case connectivity.Ready:
			return conn, nil
		case connectivity.Idle:
			conn.Connect()
		case connectivity.Shutdown:
			_ = conn.Close()
			return nil, fmt.Errorf("gRPC connection shut down before becoming ready")
		}

		if !conn.WaitForStateChange(ctx, state) {
			_ = conn.Close()
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("timed out waiting for gRPC connection to become ready")
		}
	}
}

func (c *ToolplaneClient) ensureGRPCConnected() error {
	if c.protocol != ProtocolGRPC {
		return fmt.Errorf("operation only supported with gRPC protocol")
	}
	if c.grpcConn == nil || c.toolClient == nil || c.sessionClient == nil || c.machineClient == nil || c.requestsClient == nil {
		return fmt.Errorf("gRPC client is not connected")
	}
	return nil
}

func (c *ToolplaneClient) ensureTasksClientConnected() error {
	if c.protocol != ProtocolGRPC {
		return fmt.Errorf("operation only supported with gRPC protocol")
	}
	if c.grpcConn == nil || c.tasksClient == nil {
		return fmt.Errorf("gRPC tasks client is not connected")
	}
	return nil
}

func (c *ToolplaneClient) requireSessionID(operation string) (string, error) {
	if c.sessionID == "" {
		return "", fmt.Errorf("%s requires a session ID", operation)
	}
	return c.sessionID, nil
}

func (c *ToolplaneClient) resolveMachineID(operation, machineID string) (string, error) {
	if machineID != "" {
		return machineID, nil
	}
	if c.machineID != "" {
		return c.machineID, nil
	}
	return "", fmt.Errorf("%s requires a machine ID or a registered machine", operation)
}

func (c *ToolplaneClient) grpcContext(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}

	ctx := parent
	cancel := func() {}
	if _, hasDeadline := parent.Deadline(); !hasDeadline && timeout > 0 {
		ctx, cancel = context.WithTimeout(parent, timeout)
	}

	if c.apiKey != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "api_key", c.apiKey)
	}

	return ctx, cancel
}

func (c *ToolplaneClient) waitForRequestCompletion(ctx context.Context, requestID string) (*pb.Request, error) {
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	execCtx, cancel := c.grpcContext(ctx, defaultGRPCExecutionTimeout)
	defer cancel()

	ticker := time.NewTicker(defaultRequestPollInterval)
	defer ticker.Stop()

	for {
		requestCtx, requestCancel := c.grpcContext(execCtx, defaultGRPCCallTimeout)
		request, err := c.requestsClient.GetRequest(requestCtx, &pb.GetRequestRequest{
			SessionId: c.sessionID,
			RequestId: requestID,
		})
		requestCancel()
		if err != nil {
			return nil, fmt.Errorf("failed to get request %s: %w", requestID, err)
		}

		switch request.Status {
		case "done":
			return request, nil
		case "failure":
			errMsg := request.Error
			if errMsg == "" {
				errMsg = "tool execution failed"
			}
			return nil, fmt.Errorf("request %s failed: %s", requestID, errMsg)
		}

		select {
		case <-execCtx.Done():
			return nil, fmt.Errorf("timed out waiting for request %s: %w", requestID, execCtx.Err())
		case <-ticker.C:
		}
	}
}

func parseNumericResult(resultJSON string) (float64, error) {
	if resultJSON == "" {
		return 0, fmt.Errorf("empty result payload")
	}

	var numeric float64
	if err := json.Unmarshal([]byte(resultJSON), &numeric); err == nil {
		return numeric, nil
	}

	var numericString string
	if err := json.Unmarshal([]byte(resultJSON), &numericString); err == nil {
		parsed, parseErr := strconv.ParseFloat(numericString, 64)
		if parseErr != nil {
			return 0, fmt.Errorf("failed to parse numeric string result %q: %w", numericString, parseErr)
		}
		return parsed, nil
	}

	return 0, fmt.Errorf("failed to parse numeric result payload: %s", resultJSON)
}

// Ping tests server connectivity
func (c *ToolplaneClient) Ping() (string, error) {
	if err := c.ensureGRPCConnected(); err != nil {
		return "", err
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	resp, err := c.toolClient.HealthCheck(ctx, &pb.HealthCheckRequest{})
	if err != nil {
		return "", err
	}
	return resp.Status, nil
}

// Add performs addition
func (c *ToolplaneClient) Add(a, b float64) (float64, error) {
	request, err := c.executeToolGRPC(context.Background(), "add", map[string]interface{}{"a": a, "b": b})
	if err != nil {
		return 0, err
	}
	return parseNumericResult(request.GetResult())
}

// Subtract performs subtraction
func (c *ToolplaneClient) Subtract(a, b float64) (float64, error) {
	request, err := c.executeToolGRPC(context.Background(), "subtract", map[string]interface{}{"a": a, "b": b})
	if err != nil {
		return 0, err
	}
	return parseNumericResult(request.GetResult())
}

// Multiply performs multiplication
func (c *ToolplaneClient) Multiply(a, b float64) (float64, error) {
	request, err := c.executeToolGRPC(context.Background(), "multiply", map[string]interface{}{"a": a, "b": b})
	if err != nil {
		return 0, err
	}
	return parseNumericResult(request.GetResult())
}

// Divide performs division
func (c *ToolplaneClient) Divide(a, b float64) (float64, error) {
	request, err := c.executeToolGRPC(context.Background(), "divide", map[string]interface{}{"a": a, "b": b})
	if err != nil {
		return 0, err
	}
	return parseNumericResult(request.GetResult())
}

// ExecuteTool executes a tool via gRPC and waits for the final request result.
func (c *ToolplaneClient) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (*pb.Request, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("tool execution only supported with gRPC protocol")
	}

	return c.executeToolGRPC(ctx, toolName, params)
}

// executeToolGRPC executes a tool via gRPC and waits for the final request result.
func (c *ToolplaneClient) executeToolGRPC(ctx context.Context, toolName string, params map[string]interface{}) (*pb.Request, error) {
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	request := &pb.ExecuteToolRequest{
		SessionId: c.sessionID,
		ToolName:  toolName,
		Input:     string(paramsJSON),
	}

	execCtx, cancel := c.grpcContext(ctx, defaultGRPCCallTimeout)
	defer cancel()

	response, err := c.toolClient.ExecuteTool(execCtx, request)
	if err != nil {
		return nil, err
	}

	if response.Error != "" {
		return nil, fmt.Errorf("tool execution error: %s", response.Error)
	}

	if response.RequestId == "" {
		return nil, fmt.Errorf("tool execution did not return a request ID")
	}

	return c.waitForRequestCompletion(ctx, response.RequestId)
}

// StreamExecuteTool consumes a live gRPC execution stream in-order.
func (c *ToolplaneClient) StreamExecuteTool(
	ctx context.Context,
	toolName string,
	params map[string]interface{},
	onChunk func(*pb.ExecuteToolChunk) error,
) ([]*pb.ExecuteToolChunk, error) {
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	streamCtx, cancel := c.grpcContext(ctx, defaultGRPCExecutionTimeout)
	defer cancel()

	stream, err := c.toolClient.StreamExecuteTool(streamCtx, &pb.ExecuteToolRequest{
		SessionId: c.sessionID,
		ToolName:  toolName,
		Input:     string(paramsJSON),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start execution stream: %w", err)
	}

	chunks := make([]*pb.ExecuteToolChunk, 0, 8)
	for {
		chunk, recvErr := stream.Recv()
		if recvErr != nil {
			return chunks, fmt.Errorf("failed to receive stream chunk: %w", recvErr)
		}

		chunks = append(chunks, chunk)
		if onChunk != nil {
			if callbackErr := onChunk(chunk); callbackErr != nil {
				return chunks, callbackErr
			}
		}

		if chunk.IsFinal {
			if chunk.Error != "" {
				return chunks, fmt.Errorf("stream execution failed: %s", chunk.Error)
			}
			return chunks, nil
		}
	}
}

// RegisterTool registers a new tool (gRPC only)
func (c *ToolplaneClient) RegisterTool(name, description, schema string, config map[string]string, tags []string) (*pb.Tool, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("tool registration only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	request := &pb.RegisterToolRequest{
		SessionId:   c.sessionID,
		MachineId:   c.machineID,
		Name:        name,
		Description: description,
		Schema:      schema,
		Config:      config,
		Tags:        tags,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	response, err := c.toolClient.RegisterTool(ctx, request)
	if err != nil {
		return nil, err
	}

	return response.Tool, nil
}

// ListTools lists all tools in the session (gRPC only)
func (c *ToolplaneClient) ListTools() ([]*pb.Tool, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("tool listing only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	request := &pb.ListToolsRequest{
		SessionId: c.sessionID,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	response, err := c.toolClient.ListTools(ctx, request)
	if err != nil {
		return nil, err
	}

	return response.Tools, nil
}

// GetToolByID retrieves a tool by ID for the current session (gRPC only).
func (c *ToolplaneClient) GetToolByID(toolID string) (*pb.Tool, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("tool lookup only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	sessionID, err := c.requireSessionID("tool lookup")
	if err != nil {
		return nil, err
	}

	request := &pb.GetToolByIdRequest{
		SessionId: sessionID,
		ToolId:    toolID,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	response, err := c.toolClient.GetToolById(ctx, request)
	if err != nil {
		return nil, err
	}

	return response.Tool, nil
}

// GetToolByName retrieves a tool by name for the current session (gRPC only).
func (c *ToolplaneClient) GetToolByName(toolName string) (*pb.Tool, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("tool lookup only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	sessionID, err := c.requireSessionID("tool lookup")
	if err != nil {
		return nil, err
	}

	request := &pb.GetToolByNameRequest{
		SessionId: sessionID,
		ToolName:  toolName,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	response, err := c.toolClient.GetToolByName(ctx, request)
	if err != nil {
		return nil, err
	}

	return response.Tool, nil
}

// DeleteTool deletes a tool by ID for the current session (gRPC only).
func (c *ToolplaneClient) DeleteTool(toolID string) (bool, error) {
	if c.protocol != ProtocolGRPC {
		return false, fmt.Errorf("tool deletion only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return false, err
	}

	sessionID, err := c.requireSessionID("tool deletion")
	if err != nil {
		return false, err
	}

	request := &pb.DeleteToolRequest{
		SessionId: sessionID,
		ToolId:    toolID,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	response, err := c.toolClient.DeleteTool(ctx, request)
	if err != nil {
		return false, err
	}

	return response.Success, nil
}

// CreateSession creates a new session (gRPC only)
func (c *ToolplaneClient) CreateSession(name, description, namespace string) (*pb.Session, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("session creation only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	sessionID := c.sessionID
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	request := &pb.CreateSessionRequest{
		UserId:      c.userID,
		Name:        name,
		Description: description,
		SessionId:   sessionID,
		Namespace:   namespace,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	response, err := c.sessionClient.CreateSession(ctx, request)
	if err != nil {
		return nil, err
	}

	// Update client session ID
	c.sessionID = response.Session.Id

	return response.Session, nil
}

// GetSession retrieves session information (gRPC only)
func (c *ToolplaneClient) GetSession() (*pb.Session, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("session retrieval only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	request := &pb.GetSessionRequest{
		SessionId: c.sessionID,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	return c.sessionClient.GetSession(ctx, request)
}

// ListSessions lists sessions for the configured user (gRPC only).
func (c *ToolplaneClient) ListSessions() ([]*pb.Session, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("session listing only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	request := &pb.ListSessionsRequest{UserId: c.userID}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	response, err := c.sessionClient.ListSessions(ctx, request)
	if err != nil {
		return nil, err
	}

	return response.Sessions, nil
}

// UpdateSession updates the current session metadata (gRPC only).
func (c *ToolplaneClient) UpdateSession(name, description, namespace string) (*pb.Session, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("session update only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	sessionID, err := c.requireSessionID("session update")
	if err != nil {
		return nil, err
	}

	request := &pb.UpdateSessionRequest{
		SessionId:   sessionID,
		Name:        name,
		Description: description,
		Namespace:   namespace,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	return c.sessionClient.UpdateSession(ctx, request)
}

// CreateAPIKey creates an API key for the current session (gRPC only).
// The returned secret is only populated on creation; ListAPIKeys returns metadata.
func (c *ToolplaneClient) CreateAPIKey(name string, capabilities ...string) (*pb.ApiKey, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("api key creation only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	sessionID, err := c.requireSessionID("api key creation")
	if err != nil {
		return nil, err
	}

	request := &pb.CreateApiKeyRequest{
		SessionId:    sessionID,
		Name:         name,
		Capabilities: append([]string(nil), capabilities...),
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	return c.sessionClient.CreateApiKey(ctx, request)
}

// ListAPIKeys lists API keys for the current session (gRPC only).
func (c *ToolplaneClient) ListAPIKeys() ([]*pb.ApiKey, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("api key listing only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	sessionID, err := c.requireSessionID("api key listing")
	if err != nil {
		return nil, err
	}

	request := &pb.ListApiKeysRequest{SessionId: sessionID}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	response, err := c.sessionClient.ListApiKeys(ctx, request)
	if err != nil {
		return nil, err
	}

	return response.ApiKeys, nil
}

// RevokeAPIKey revokes an API key for the current session (gRPC only).
func (c *ToolplaneClient) RevokeAPIKey(keyID string) (bool, error) {
	if c.protocol != ProtocolGRPC {
		return false, fmt.Errorf("api key revocation only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return false, err
	}

	sessionID, err := c.requireSessionID("api key revocation")
	if err != nil {
		return false, err
	}

	request := &pb.RevokeApiKeyRequest{
		SessionId: sessionID,
		KeyId:     keyID,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	response, err := c.sessionClient.RevokeApiKey(ctx, request)
	if err != nil {
		return false, err
	}

	return response.Success, nil
}

// RegisterMachine registers this client as a machine (gRPC only)
func (c *ToolplaneClient) RegisterMachine(machineID, sdkVersion string, tools []*pb.RegisterToolRequest) (*pb.Machine, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("machine registration only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	if machineID == "" {
		machineID = uuid.New().String()
	}

	request := &pb.RegisterMachineRequest{
		SessionId:   c.sessionID,
		MachineId:   machineID,
		SdkVersion:  sdkVersion,
		SdkLanguage: "go",
		Tools:       tools,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	machine, err := c.machineClient.RegisterMachine(ctx, request)
	if err != nil {
		return nil, err
	}

	c.machineID = machine.Id
	return machine, nil
}

// ListMachines lists machines for the current session (gRPC only)
func (c *ToolplaneClient) ListMachines() ([]*pb.Machine, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("machine listing only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	sessionID, err := c.requireSessionID("machine listing")
	if err != nil {
		return nil, err
	}

	request := &pb.ListMachinesRequest{SessionId: sessionID}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	response, err := c.machineClient.ListMachines(ctx, request)
	if err != nil {
		return nil, err
	}

	return response.Machines, nil
}

// GetMachine retrieves a machine by ID for the current session (gRPC only)
func (c *ToolplaneClient) GetMachine(machineID string) (*pb.Machine, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("machine retrieval only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	sessionID, err := c.requireSessionID("machine retrieval")
	if err != nil {
		return nil, err
	}

	request := &pb.GetMachineRequest{
		SessionId: sessionID,
		MachineId: machineID,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	return c.machineClient.GetMachine(ctx, request)
}

// UnregisterMachine unregisters a machine for the current session (gRPC only)
func (c *ToolplaneClient) UnregisterMachine(machineID string) (bool, error) {
	if c.protocol != ProtocolGRPC {
		return false, fmt.Errorf("machine unregistration only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return false, err
	}

	sessionID, err := c.requireSessionID("machine unregistration")
	if err != nil {
		return false, err
	}

	resolvedMachineID, err := c.resolveMachineID("machine unregistration", machineID)
	if err != nil {
		return false, err
	}

	request := &pb.UnregisterMachineRequest{
		SessionId: sessionID,
		MachineId: resolvedMachineID,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	response, err := c.machineClient.UnregisterMachine(ctx, request)
	if err != nil {
		return false, err
	}

	if response.Success && resolvedMachineID == c.machineID {
		c.machineID = ""
	}

	return response.Success, nil
}

// DrainMachine drains a machine for the current session (gRPC only)
func (c *ToolplaneClient) DrainMachine(machineID string) (bool, error) {
	if c.protocol != ProtocolGRPC {
		return false, fmt.Errorf("machine drain only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return false, err
	}

	sessionID, err := c.requireSessionID("machine drain")
	if err != nil {
		return false, err
	}

	resolvedMachineID, err := c.resolveMachineID("machine drain", machineID)
	if err != nil {
		return false, err
	}

	request := &pb.DrainMachineRequest{
		SessionId: sessionID,
		MachineId: resolvedMachineID,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	response, err := c.machineClient.DrainMachine(ctx, request)
	if err != nil {
		return false, err
	}

	if response.Drained && resolvedMachineID == c.machineID {
		c.machineID = ""
	}

	return response.Drained, nil
}

// CreateRequest creates a request for the current session (gRPC only).
func (c *ToolplaneClient) CreateRequest(toolName, input string) (*pb.Request, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("request creation only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	sessionID, err := c.requireSessionID("request creation")
	if err != nil {
		return nil, err
	}

	request := &pb.CreateRequestRequest{
		SessionId: sessionID,
		ToolName:  toolName,
		Input:     input,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	return c.requestsClient.CreateRequest(ctx, request)
}

// GetRequest retrieves a request by ID for the current session (gRPC only).
func (c *ToolplaneClient) GetRequest(requestID string) (*pb.Request, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("request retrieval only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	sessionID, err := c.requireSessionID("request retrieval")
	if err != nil {
		return nil, err
	}

	request := &pb.GetRequestRequest{
		SessionId: sessionID,
		RequestId: requestID,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	return c.requestsClient.GetRequest(ctx, request)
}

// ListRequests lists requests for the current session (gRPC only).
func (c *ToolplaneClient) ListRequests(status, toolName string, limit, offset int32) ([]*pb.Request, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("request listing only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	sessionID, err := c.requireSessionID("request listing")
	if err != nil {
		return nil, err
	}

	request := &pb.ListRequestsRequest{
		SessionId: sessionID,
		Status:    status,
		ToolName:  toolName,
		Limit:     limit,
		Offset:    offset,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	response, err := c.requestsClient.ListRequests(ctx, request)
	if err != nil {
		return nil, err
	}

	return response.Requests, nil
}

// CancelRequest cancels a request for the current session (gRPC only).
func (c *ToolplaneClient) CancelRequest(requestID string) (bool, error) {
	if c.protocol != ProtocolGRPC {
		return false, fmt.Errorf("request cancellation only supported with gRPC protocol")
	}
	if err := c.ensureGRPCConnected(); err != nil {
		return false, err
	}

	sessionID, err := c.requireSessionID("request cancellation")
	if err != nil {
		return false, err
	}

	request := &pb.CancelRequestRequest{
		SessionId: sessionID,
		RequestId: requestID,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	response, err := c.requestsClient.CancelRequest(ctx, request)
	if err != nil {
		return false, err
	}

	return response.Success, nil
}

// CreateTask creates a task for the current session (gRPC only).
func (c *ToolplaneClient) CreateTask(toolName, input string) (*pb.Task, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("task creation only supported with gRPC protocol")
	}
	if err := c.ensureTasksClientConnected(); err != nil {
		return nil, err
	}

	sessionID, err := c.requireSessionID("task creation")
	if err != nil {
		return nil, err
	}

	request := &pb.CreateTaskRequest{
		SessionId: sessionID,
		ToolName:  toolName,
		Input:     input,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	return c.tasksClient.CreateTask(ctx, request)
}

// GetTask retrieves a task by ID for the current session (gRPC only).
func (c *ToolplaneClient) GetTask(taskID string) (*pb.Task, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("task retrieval only supported with gRPC protocol")
	}
	if err := c.ensureTasksClientConnected(); err != nil {
		return nil, err
	}

	sessionID, err := c.requireSessionID("task retrieval")
	if err != nil {
		return nil, err
	}

	request := &pb.GetTaskRequest{
		SessionId: sessionID,
		TaskId:    taskID,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	return c.tasksClient.GetTask(ctx, request)
}

// ListTasks lists tasks for the current session (gRPC only).
func (c *ToolplaneClient) ListTasks() ([]*pb.Task, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("task listing only supported with gRPC protocol")
	}
	if err := c.ensureTasksClientConnected(); err != nil {
		return nil, err
	}

	sessionID, err := c.requireSessionID("task listing")
	if err != nil {
		return nil, err
	}

	request := &pb.ListTasksRequest{SessionId: sessionID}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	response, err := c.tasksClient.ListTasks(ctx, request)
	if err != nil {
		return nil, err
	}

	return response.Tasks, nil
}

// CancelTask cancels a task for the current session (gRPC only).
func (c *ToolplaneClient) CancelTask(taskID string) (bool, error) {
	if c.protocol != ProtocolGRPC {
		return false, fmt.Errorf("task cancellation only supported with gRPC protocol")
	}
	if err := c.ensureTasksClientConnected(); err != nil {
		return false, err
	}

	sessionID, err := c.requireSessionID("task cancellation")
	if err != nil {
		return false, err
	}

	request := &pb.CancelTaskRequest{
		SessionId: sessionID,
		TaskId:    taskID,
	}

	ctx, cancel := c.grpcContext(context.Background(), defaultGRPCCallTimeout)
	defer cancel()

	response, err := c.tasksClient.CancelTask(ctx, request)
	if err != nil {
		return false, err
	}

	return response.Success, nil
}
