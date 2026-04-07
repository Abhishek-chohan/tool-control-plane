package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"toolplane-go-client/client"
	clientpb "toolplane-go-client/proto"
)

func main() {
	fmt.Println("=== Advanced Go Toolplane Client Demo ===")
	fmt.Println("Maintained path: gRPC control-plane helpers")

	fmt.Println("\n1. Testing Maintained Session Management:")
	testSessionManagement()

	fmt.Println("\n2. Testing Boundary And Connection Errors:")
	testBoundaryErrors()

	fmt.Println("\n3. Testing Concurrent Session Reads:")
	testConcurrentSessionReads()
}

func testConcurrentSessionReads() {
	serverHost := getEnv("TOOLPLANE_SERVER_HOST", "localhost")
	serverPort := getEnvInt("TOOLPLANE_SERVER_PORT", 9001)
	toolplaneClient, err := client.NewToolplaneClient(
		client.ProtocolGRPC,
		serverHost,
		serverPort,
		"",
		"concurrency-test-user",
		getEnv("TOOLPLANE_API_KEY", "toolplane-conformance-fixture-key"),
		buildClientOptions()...,
	)
	if err != nil {
		log.Printf("Failed to create gRPC client: %v", err)
		return
	}

	if err := toolplaneClient.Connect(); err != nil {
		log.Printf("   gRPC server not available: %v", err)
		return
	}
	defer toolplaneClient.Disconnect()

	session, err := toolplaneClient.CreateSession("Concurrent Reads", "Concurrent session inspection example", "advanced-examples")
	if err != nil {
		log.Printf("   Failed to create session for concurrent reads: %v", err)
		return
	}
	fmt.Printf("   Created session for concurrent reads: %s (ID: %s)\n", session.Name, session.Id)

	start := time.Now()
	readers := 4
	var wg sync.WaitGroup
	results := make(chan string, readers)

	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			sessionInfo, err := toolplaneClient.GetSession()
			if err != nil {
				results <- fmt.Sprintf("Reader %d: Session read failed - %v", readerID, err)
				return
			}

			results <- fmt.Sprintf("Reader %d: Session %s (%s)", readerID, sessionInfo.Name, sessionInfo.Id)
		}(i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		fmt.Printf("   %s\n", result)
	}
	fmt.Printf("   Completed %d concurrent reads in %v\n", readers, time.Since(start))
}

func testSessionManagement() {
	serverHost := getEnv("TOOLPLANE_SERVER_HOST", "localhost")
	serverPort := getEnvInt("TOOLPLANE_SERVER_PORT", 9001)
	toolplaneClient, err := client.NewToolplaneClient(
		client.ProtocolGRPC,
		serverHost,
		serverPort,
		"",
		"session-test-user",
		getEnv("TOOLPLANE_API_KEY", "toolplane-conformance-fixture-key"),
		buildClientOptions()...,
	)
	if err != nil {
		log.Printf("Failed to create gRPC client: %v", err)
		return
	}

	if err := toolplaneClient.Connect(); err != nil {
		log.Printf("   gRPC server not available: %v", err)
		return
	}
	defer toolplaneClient.Disconnect()

	sessions := []struct {
		name        string
		description string
		namespace   string
	}{
		{"Development", "Development environment session", "dev"},
		{"Testing", "Testing environment session", "test"},
		{"Production", "Production environment session", "prod"},
	}

	for _, sessionSpec := range sessions {
		session, err := toolplaneClient.CreateSession(sessionSpec.name, sessionSpec.description, sessionSpec.namespace)
		if err != nil {
			log.Printf("   Failed to create session %s: %v", sessionSpec.name, err)
			continue
		}
		fmt.Printf("   Created session: %s (ID: %s, Namespace: %s)\n", session.Name, session.Id, session.Namespace)

		toolDefinitions := make([]*clientpb.RegisterToolRequest, 0, 3)
		for i := 0; i < 3; i++ {
			toolName := fmt.Sprintf("tool_%s_%d", sessionSpec.namespace, i)
			toolDefinitions = append(toolDefinitions, &clientpb.RegisterToolRequest{
				Name:        toolName,
				Description: fmt.Sprintf("Tool %d for %s environment", i, sessionSpec.name),
				Schema: fmt.Sprintf(`{
					"type": "object",
					"properties": {
						"input": {"type": "string", "description": "Input for %s"}
					},
					"required": ["input"]
				}`, toolName),
				Config: map[string]string{
					"environment": sessionSpec.namespace,
					"version":     "1.0",
				},
				Tags: []string{sessionSpec.namespace, "auto-generated"},
			})
		}

		machine, err := toolplaneClient.RegisterMachine("", "1.0.0", toolDefinitions)
		if err != nil {
			log.Printf("   Failed to register machine for session %s: %v", sessionSpec.name, err)
			continue
		}
		fmt.Printf("   Registered machine: %s\n", machine.Id)

		tools, err := toolplaneClient.ListTools()
		if err != nil {
			log.Printf("   Failed to list tools: %v", err)
			continue
		}
		fmt.Printf("   Session %s has %d tools\n", session.Name, len(tools))
	}
}

func testBoundaryErrors() {
	invalidClient, err := client.NewToolplaneClient(
		client.ProtocolGRPC,
		"invalid-host",
		9999,
		"",
		"error-test-user",
		getEnv("TOOLPLANE_API_KEY", "toolplane-conformance-fixture-key"),
		buildClientOptions()...,
	)
	if err != nil {
		log.Printf("Failed to create invalid-host client: %v", err)
		return
	}

	fmt.Printf("   Testing invalid gRPC server: ")
	if err := invalidClient.Connect(); err != nil {
		fmt.Printf("Error handled correctly: %v\n", err)
	} else {
		defer invalidClient.Disconnect()
		fmt.Printf("Expected error but got success\n")
	}

	fmt.Println("   Demo-only arithmetic aliases remain thin wrappers over the maintained gRPC request flow.")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}

func buildClientOptions() []client.ClientOption {
	if getEnv("TOOLPLANE_USE_TLS", "") == "" {
		return nil
	}
	if getEnv("TOOLPLANE_USE_TLS", "false") != "true" {
		return nil
	}

	return []client.ClientOption{
		client.WithGRPCTLS(
			getEnv("TOOLPLANE_TLS_CA_CERT_PATH", ""),
			getEnv("TOOLPLANE_TLS_SERVER_NAME", ""),
		),
	}
}
