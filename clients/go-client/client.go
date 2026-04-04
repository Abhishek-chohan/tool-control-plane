package main

import (
	"fmt"
	"log"
	"os"

	"toolplane-go-client/client"
	clientpb "toolplane-go-client/proto"
)

func main() {
	fmt.Println("=== Go Toolplane Client ===")
	fmt.Println("Maintained path: gRPC control-plane helpers")

	serverHost := getEnv("TOOLPLANE_SERVER_HOST", "localhost")
	serverPort := 9001
	sessionID := getEnv("TOOLPLANE_SESSION_ID", "go-client-session")
	userID := getEnv("TOOLPLANE_USER_ID", "go-client-user")
	apiKey := getEnv("TOOLPLANE_API_KEY", "toolplane-conformance-fixture-key")

	toolplaneClient, err := client.NewToolplaneClient(
		client.ProtocolGRPC,
		serverHost,
		serverPort,
		sessionID,
		userID,
		apiKey,
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Printf("Connecting to %s:%d...\n", serverHost, serverPort)
	if err := toolplaneClient.Connect(); err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer func() {
		if err := toolplaneClient.Disconnect(); err != nil {
			log.Printf("Disconnect warning: %v", err)
		}
	}()

	fmt.Println("Connected successfully")
	runBasicOperations(toolplaneClient)
	runGRPCFeatures(toolplaneClient)
}

func runBasicOperations(c *client.ToolplaneClient) {
	fmt.Println("\n=== Connectivity ===")

	if result, err := c.Ping(); err == nil {
		fmt.Printf("Ping: %s\n", result)
	} else {
		fmt.Printf("Ping failed: %v\n", err)
	}

	fmt.Println("Live gRPC tool execution requires a machine-backed provider loop. This maintained demo focuses on session and machine lifecycle first.")
}

func runGRPCFeatures(c *client.ToolplaneClient) {
	fmt.Println("\n=== Maintained gRPC Control-Plane Features ===")

	session, err := c.CreateSession(
		"Go Control Plane Session",
		"Provider-backed session created by the Go gRPC client demo",
		"examples",
	)
	if err != nil {
		fmt.Printf("Session creation failed: %v\n", err)
		return
	}
	fmt.Printf("Created session: %s (ID: %s)\n", session.Name, session.Id)

	machineTools := []*clientpb.RegisterToolRequest{
		{
			Name:        "session_status",
			Description: "Reports session context for a connected operator or automation client",
			Schema:      `{"type":"object","properties":{"requester":{"type":"string","description":"The caller requesting session state"}},"required":["requester"]}`,
			Config: map[string]string{
				"version":    "1.0",
				"language":   "go",
				"created_by": "go-client",
			},
			Tags: []string{"session", "status", "operator"},
		},
		{
			Name:        "change_summary",
			Description: "Builds a concise summary for a rollout or change event",
			Schema:      `{"type":"object","properties":{"service":{"type":"string"},"state":{"type":"string"}},"required":["service","state"]}`,
			Config: map[string]string{
				"version":    "1.0",
				"language":   "go",
				"created_by": "go-client",
			},
			Tags: []string{"change", "summary", "operations"},
		},
		{
			Name:        "incident_brief",
			Description: "Generates a brief operator-facing incident summary",
			Schema:      `{"type":"object","properties":{"service":{"type":"string"},"severity":{"type":"string"}},"required":["service","severity"]}`,
			Config: map[string]string{
				"version":    "1.0",
				"language":   "go",
				"created_by": "go-client",
			},
			Tags: []string{"incident", "summary", "operations"},
		},
	}

	machine, err := c.RegisterMachine("", "1.0.0", machineTools)
	if err != nil {
		fmt.Printf("Machine registration failed: %v\n", err)
		return
	}
	fmt.Printf("Registered machine: %s\n", machine.Id)

	allTools, err := c.ListTools()
	if err != nil {
		fmt.Printf("Failed to list tools: %v\n", err)
		return
	}
	fmt.Printf("Session contains %d tools:\n", len(allTools))
	for i, tool := range allTools {
		fmt.Printf("  %d. %s - %s\n", i+1, tool.Name, tool.Description)
	}

	fmt.Println("Tool execution requires a provider loop to claim requests. See client/toolplane_client_integration_test.go for a live unary and streaming example.")

	sessionInfo, err := c.GetSession()
	if err != nil {
		fmt.Printf("Session retrieval failed: %v\n", err)
		return
	}
	fmt.Printf("Session details:\n")
	fmt.Printf("  Name: %s\n", sessionInfo.Name)
	fmt.Printf("  ID: %s\n", sessionInfo.Id)
	fmt.Printf("  Namespace: %s\n", sessionInfo.Namespace)
	fmt.Printf("  Created: %s\n", sessionInfo.CreatedAt)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
