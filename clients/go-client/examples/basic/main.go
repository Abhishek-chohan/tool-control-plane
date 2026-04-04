package main

import (
	"fmt"
	"log"
	"os"

	"toolplane-go-client/client"
	clientpb "toolplane-go-client/proto"
)

func main() {
	fmt.Println("=== Go Toolplane Maintained Example ===")
	fmt.Println("Maintained path: gRPC control-plane helpers (session, machine, tool, request lifecycle)")

	fmt.Println("\n1. Maintained gRPC Client:")
	testGRPCClient()
}

func testGRPCClient() {
	serverHost := getEnv("TOOLPLANE_SERVER_HOST", "localhost")
	grpcClient, err := client.NewToolplaneClient(
		client.ProtocolGRPC,
		serverHost,
		9001,
		"test-session",
		"test-user",
		getEnv("TOOLPLANE_API_KEY", "toolplane-conformance-fixture-key"),
	)
	if err != nil {
		log.Printf("Failed to create gRPC client: %v", err)
		return
	}

	if err := grpcClient.Connect(); err != nil {
		log.Printf("   gRPC server not available (expected if not running): %v", err)
		return
	}
	defer grpcClient.Disconnect()

	fmt.Printf("   Connected to gRPC server\n")
	testConnectivity(grpcClient)
	testGRPCSpecificFeatures(grpcClient)
}

func testConnectivity(toolplaneClient *client.ToolplaneClient) {
	if result, err := toolplaneClient.Ping(); err == nil {
		fmt.Printf("   Ping: %s\n", result)
	} else {
		fmt.Printf("   Ping failed: %v\n", err)
	}

	fmt.Printf("   Live gRPC tool execution now requires a machine-backed provider loop. See client/toolplane_client_integration_test.go for a runnable example.\n")
}

func testGRPCSpecificFeatures(toolplaneClient *client.ToolplaneClient) {
	session, err := toolplaneClient.CreateSession("Go Test Session", "Session created by Go client", "test-namespace")
	if err != nil {
		log.Printf("   Session creation failed: %v", err)
		return
	}
	fmt.Printf("   Created session: %s (ID: %s)\n", session.Name, session.Id)

	toolDefinitions := []*clientpb.RegisterToolRequest{
		{
			Name:        "session_status",
			Description: "Report session context for a connected operator or automation client",
			Schema:      `{"type":"object","properties":{"requester":{"type":"string"}},"required":["requester"]}`,
			Config:      map[string]string{"version": "1.0"},
			Tags:        []string{"session", "status", "go"},
		},
	}

	machine, err := toolplaneClient.RegisterMachine("", "1.0.0", toolDefinitions)
	if err != nil {
		log.Printf("   Machine registration failed: %v", err)
		return
	}
	fmt.Printf("   Registered machine: %s\n", machine.Id)

	tools, err := toolplaneClient.ListTools()
	if err != nil {
		log.Printf("   Tool listing failed: %v", err)
		return
	}
	fmt.Printf("   Found %d tools in session\n", len(tools))
	fmt.Printf("   Live execution requires a provider loop to claim requests. See client/toolplane_client_integration_test.go for a runnable example.\n")

	sessionInfo, err := toolplaneClient.GetSession()
	if err != nil {
		log.Printf("   Session retrieval failed: %v", err)
		return
	}
	fmt.Printf("   Session info: %s (Created: %s)\n", sessionInfo.Name, sessionInfo.CreatedAt)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
