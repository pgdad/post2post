package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pgdad/post2post"
)

func main() {
	fmt.Println("Post2Post Tailscale Client Example")
	fmt.Println("===================================")

	// Get Tailscale auth key from environment variable
	tailnetKey := os.Getenv("TAILSCALE_AUTH_KEY")
	if tailnetKey == "" {
		log.Fatal("TAILSCALE_AUTH_KEY environment variable is required")
	}

	// Get receiver URL from environment variable or use default
	receiverURL := os.Getenv("RECEIVER_URL")
	if receiverURL == "" {
		receiverURL = "http://127.0.0.1:8082/webhook"
		fmt.Printf("Using default receiver URL: %s\n", receiverURL)
		fmt.Println("Set RECEIVER_URL environment variable to override")
	}

	// Get network interface address from environment variable or use default
	interfaceAddr := os.Getenv("LISTEN_INTERFACE")
	if interfaceAddr == "" {
		interfaceAddr = "127.0.0.1"
		fmt.Printf("Using default interface: %s\n", interfaceAddr)
		fmt.Println("Set LISTEN_INTERFACE environment variable to override")
	}

	fmt.Printf("Tailscale key: %s...\n", tailnetKey[:min(10, len(tailnetKey))])
	fmt.Printf("Receiver URL: %s\n", receiverURL)
	fmt.Printf("Listen interface: %s\n", interfaceAddr)
	fmt.Println()

	// Create post2post server for receiving responses
	server := post2post.NewServer().
		WithInterface(interfaceAddr).
		WithPostURL(receiverURL).
		WithTimeout(30 * time.Second)

	// Start the local server to receive responses
	err := server.Start()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	fmt.Printf("Local server started at: %s\n", server.GetURL())
	fmt.Printf("Ready to receive Tailscale responses at: %s/roundtrip\n", server.GetURL())
	fmt.Println()

	// Test 1: Send request with Tailscale integration
	fmt.Println("Test 1: Tailscale round-trip request")
	fmt.Println("-----------------------------------")

	// Generate random string for payload
	randomData := generateRandomString(16)
	
	payload1 := map[string]interface{}{
		"message":     "Hello from Tailscale client!",
		"random_data": randomData,
		"client_info": map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"version":   "1.0",
			"secure":    true,
		},
		"test_id": "tailscale-test-001",
	}

	fmt.Printf("Sending payload with random data: %s\n", randomData)

	// For round-trip with Tailscale, we need to include the tailnet_key in the payload
	// The receiver will use it for the response posting
	payload1["tailnet_key"] = tailnetKey
	
	response1, err := server.RoundTripPost(payload1)
	if err != nil {
		log.Printf("Tailscale round-trip failed: %v", err)
		fmt.Println("\nNote: Make sure the receiver is built and running with full Tailscale integration.")
		fmt.Println("Build receiver with: ./build_receiver_tailnet.sh")
		fmt.Println("Run receiver with: ./receiver_tailnet_with_tsnet")
	} else {
		printTailscaleResponse("Test 1", response1)
	}

	fmt.Println()

	// Test 2: Multiple concurrent Tailscale requests
	fmt.Println("Test 2: Concurrent Tailscale requests")
	fmt.Println("------------------------------------")

	const numRequests = 3
	results := make(chan result, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(id int) {
			randomData := generateRandomString(12)
			payload := map[string]interface{}{
				"request_id":  fmt.Sprintf("concurrent_%d", id),
				"message":     fmt.Sprintf("Concurrent Tailscale request #%d", id),
				"random_data": randomData,
				"timestamp":   time.Now().Unix(),
				"secure":      true,
				"tailnet_key": tailnetKey,
			}

			response, err := server.RoundTripPost(payload)
			results <- result{ID: id, Response: response, Error: err}
		}(i)
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		res := <-results
		if res.Error != nil {
			log.Printf("Concurrent Tailscale request %d failed: %v", res.ID, res.Error)
		} else {
			printTailscaleResponse(fmt.Sprintf("Concurrent #%d", res.ID), res.Response)
		}
	}

	fmt.Println()

	// Test 3: Fire-and-forget Tailscale post
	fmt.Println("Test 3: Fire-and-forget Tailscale post")
	fmt.Println("-------------------------------------")

	payload3 := map[string]interface{}{
		"type":        "notification",
		"message":     "Fire-and-forget Tailscale message",
		"random_data": generateRandomString(8),
		"sent_at":     time.Now().Format("2006-01-02 15:04:05"),
		"priority":    "high",
	}

	err = server.PostJSONWithTailnet(payload3, tailnetKey)
	if err != nil {
		log.Printf("Fire-and-forget Tailscale post failed: %v", err)
	} else {
		fmt.Println("Fire-and-forget Tailscale message sent successfully!")
	}

	fmt.Println()
	fmt.Println("All Tailscale tests completed!")
	fmt.Println()
	printTailscaleInstructions()
}

// result holds the result of a concurrent request
type result struct {
	ID       int
	Response *post2post.RoundTripResponse
	Error    error
}

// generateRandomString creates a random hex string of specified length
func generateRandomString(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		log.Printf("Failed to generate random data: %v", err)
		return fmt.Sprintf("fallback_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// printTailscaleResponse prints the response in a formatted way with Tailscale-specific info
func printTailscaleResponse(label string, response *post2post.RoundTripResponse) {
	fmt.Printf("%s Result:\n", label)
	fmt.Printf("  Success: %v\n", response.Success)
	fmt.Printf("  Timeout: %v\n", response.Timeout)
	fmt.Printf("  Request ID: %s\n", response.RequestID)

	if response.Error != "" {
		fmt.Printf("  Error: %s\n", response.Error)
	}

	if response.Success && response.Payload != nil {
		fmt.Printf("  Tailscale Response:\n")
		if payloadMap, ok := response.Payload.(map[string]interface{}); ok {
			// Print Tailscale-specific information
			if tailnetKey, exists := payloadMap["tailnet_key"]; exists {
				tailnetStr := tailnetKey.(string)
				fmt.Printf("    Tailnet Key: %s...%s\n", 
					tailnetStr[:min(4, len(tailnetStr))],
					tailnetStr[max(0, len(tailnetStr)-4):])
			}
			
			if timestamp, exists := payloadMap["timestamp"]; exists {
				fmt.Printf("    Server Timestamp: %v\n", timestamp)
			}
			
			if processed, exists := payloadMap["processed_via"]; exists {
				fmt.Printf("    Processed Via: %v\n", processed)
			}
			
			if originalData, exists := payloadMap["original_data"]; exists {
				fmt.Printf("    Original Data Received: ✓\n")
				if dataMap, ok := originalData.(map[string]interface{}); ok {
					if randomData, exists := dataMap["random_data"]; exists {
						fmt.Printf("    Random Data Echoed: %v\n", randomData)
					}
				}
			}
			
			// Show any additional fields
			for key, value := range payloadMap {
				if key != "tailnet_key" && key != "timestamp" && key != "processed_via" && key != "original_data" {
					fmt.Printf("    %s: %v\n", key, value)
				}
			}
		} else {
			fmt.Printf("    Raw payload: %+v\n", response.Payload)
		}
	}
	fmt.Println()
}

// printTailscaleInstructions prints usage instructions
func printTailscaleInstructions() {
	fmt.Println("Tailscale Integration Instructions:")
	fmt.Println("==================================")
	fmt.Println()
	fmt.Println("1. Setup Tailscale:")
	fmt.Println("   - Install Tailscale on your system")
	fmt.Println("   - Create a Tailscale account and set up your tailnet")
	fmt.Println("   - Generate an auth key from the Tailscale admin console")
	fmt.Println()
	fmt.Println("2. Set environment variables:")
	fmt.Println("   export TAILSCALE_AUTH_KEY='tskey-auth-your-key-here'")
	fmt.Println("   export RECEIVER_URL='http://receiver-host:port/webhook'")
	fmt.Println("   export LISTEN_INTERFACE='0.0.0.0'  # Optional, defaults to 127.0.0.1")
	fmt.Println()
	fmt.Println("3. Build the Tailscale-enabled receiver:")
	fmt.Println("   ./build_receiver_tailnet.sh")
	fmt.Println("   (This creates receiver_tailnet_with_tsnet with full tsnet integration)")
	fmt.Println()
	fmt.Println("4. Start the receiver:")
	fmt.Println("   ./receiver_tailnet_with_tsnet")
	fmt.Println()
	fmt.Println("5. Run this client:")
	fmt.Println("   go run client_tailnet.go")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Printf("  TAILSCALE_AUTH_KEY: %s\n", getEnvStatus("TAILSCALE_AUTH_KEY"))
	fmt.Printf("  RECEIVER_URL: %s\n", getEnvStatus("RECEIVER_URL"))
	fmt.Printf("  LISTEN_INTERFACE: %s\n", getEnvStatus("LISTEN_INTERFACE"))
}

// getEnvStatus returns the status of an environment variable
func getEnvStatus(name string) string {
	value := os.Getenv(name)
	if value == "" {
		return "Not set"
	}
	if len(value) > 20 {
		return fmt.Sprintf("Set (%s...)", value[:10])
	}
	return fmt.Sprintf("Set (%s)", value)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}