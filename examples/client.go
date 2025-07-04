package main

import (
	"fmt"
	"log"
	"time"

	"github.com/pgdad/post2post"
)

func main() {
	fmt.Println("Post2Post Client Example")
	fmt.Println("========================")

	// Create server with configuration
	server := post2post.NewServer().
		WithInterface("127.0.0.1").                         // Listen on localhost
		WithPostURL("http://127.0.0.1:8081/webhook").       // Receiver server endpoint  
		WithTimeout(10 * time.Second)                       // 10 second timeout

	// Start the server
	fmt.Println("Starting post2post server...")
	err := server.Start()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	fmt.Printf("Server started at: %s\n", server.GetURL())
	fmt.Printf("Will post to receiver at: %s\n", server.GetPostURL())
	fmt.Println()

	// Example 1: Simple round trip with map payload
	fmt.Println("Example 1: Simple round trip with map payload")
	fmt.Println("----------------------------------------------")
	
	payload1 := map[string]interface{}{
		"message": "Hello from post2post client!",
		"number":  42,
		"active":  true,
		"data": map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	response1, err := server.RoundTripPost(payload1)
	if err != nil {
		log.Printf("Round trip failed: %v", err)
	} else {
		printResponse("Example 1", response1)
	}

	fmt.Println()

	// Example 2: Round trip with struct payload
	fmt.Println("Example 2: Round trip with struct payload")
	fmt.Println("-----------------------------------------")

	type UserData struct {
		Name    string   `json:"name"`
		Age     int      `json:"age"`
		Email   string   `json:"email"`
		Hobbies []string `json:"hobbies"`
	}

	payload2 := UserData{
		Name:    "Alice Johnson",
		Age:     28,
		Email:   "alice@example.com",
		Hobbies: []string{"reading", "hiking", "programming"},
	}

	response2, err := server.RoundTripPost(payload2)
	if err != nil {
		log.Printf("Round trip failed: %v", err)
	} else {
		printResponse("Example 2", response2)
	}

	fmt.Println()

	// Example 3: Round trip with custom timeout
	fmt.Println("Example 3: Round trip with custom timeout (5 seconds)")
	fmt.Println("-----------------------------------------------------")

	payload3 := map[string]interface{}{
		"action":    "process_data",
		"data":      []int{1, 2, 3, 4, 5},
		"priority":  "high",
		"timestamp": time.Now().Unix(),
	}

	response3, err := server.RoundTripPostWithTimeout(payload3, 5*time.Second)
	if err != nil {
		log.Printf("Round trip failed: %v", err)
	} else {
		printResponse("Example 3", response3)
	}

	fmt.Println()

	// Example 4: Multiple concurrent round trips
	fmt.Println("Example 4: Multiple concurrent round trips")
	fmt.Println("------------------------------------------")

	const numRequests = 3
	results := make(chan result, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(id int) {
			payload := map[string]interface{}{
				"request_id": fmt.Sprintf("concurrent_%d", id),
				"message":    fmt.Sprintf("Concurrent request #%d", id),
				"timestamp":  time.Now().Unix(),
			}

			response, err := server.RoundTripPost(payload)
			results <- result{ID: id, Response: response, Error: err}
		}(i)
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		res := <-results
		if res.Error != nil {
			log.Printf("Concurrent request %d failed: %v", res.ID, res.Error)
		} else {
			printResponse(fmt.Sprintf("Concurrent #%d", res.ID), res.Response)
		}
	}

	fmt.Println()

	// Example 5: Test regular JSON posting (fire and forget)
	fmt.Println("Example 5: Regular JSON posting (fire and forget)")
	fmt.Println("-------------------------------------------------")

	payload5 := map[string]interface{}{
		"type":    "notification",
		"message": "This is a fire-and-forget message",
		"sent_at": time.Now().Format("2006-01-02 15:04:05"),
	}

	err = server.PostJSON(payload5)
	if err != nil {
		log.Printf("JSON post failed: %v", err)
	} else {
		fmt.Println("JSON posted successfully (no response expected)")
	}

	fmt.Println()

	// Example 6: Demonstrate Tailscale integration (framework)
	fmt.Println("Example 6: Tailscale integration demonstration")
	fmt.Println("----------------------------------------------")

	payload6 := map[string]interface{}{
		"action":     "secure_processing",
		"message":    "This would use Tailscale networking",
		"timestamp":  time.Now().Unix(),
		"sensitive":  true,
	}

	// Demonstrate PostJSONWithTailnet (will show framework message)
	err = server.PostJSONWithTailnet(payload6, "tskey-auth-example123")
	if err != nil {
		fmt.Printf("Tailscale integration: %v\n", err)
		fmt.Println("Note: This demonstrates the Tailscale framework. To enable full functionality,")
		fmt.Println("      configure tsnet in the createTailscaleClient() method.")
	} else {
		fmt.Println("Tailscale request posted successfully!")
	}

	fmt.Println()
	
	// Example 7: Test different processors (these would work with different receiver configurations)
	fmt.Println("Example 7: Testing different payload types for various processors")
	fmt.Println("------------------------------------------------------------")
	
	// Test payload for Hello World processor
	fmt.Println("Testing Hello World processor (ignores payload):")
	response7a, err := server.RoundTripPost("Any payload")
	if err != nil {
		log.Printf("Hello World test failed: %v", err)
	} else {
		printResponse("Hello World", response7a)
	}
	
	// Test payload for Transform processor
	fmt.Println("Testing Transform processor (uppercase strings):")
	transformPayload := map[string]interface{}{
		"message": "hello world",
		"greeting": "good morning",
		"number": 42,
	}
	response7b, err := server.RoundTripPost(transformPayload)
	if err != nil {
		log.Printf("Transform test failed: %v", err)
	} else {
		printResponse("Transform", response7b)
	}
	
	// Test payload for Validator processor
	fmt.Println("Testing Validator processor (with missing fields):")
	validatorPayload := map[string]interface{}{
		"name": "John Doe",
		// Missing "email" field - validator should catch this
		"age": 30,
	}
	response7c, err := server.RoundTripPost(validatorPayload)
	if err != nil {
		log.Printf("Validator test failed: %v", err)
	} else {
		printResponse("Validator (invalid)", response7c)
	}
	
	// Test payload for Validator processor with all required fields
	fmt.Println("Testing Validator processor (with all required fields):")
	validPayload := map[string]interface{}{
		"name": "Jane Smith",
		"email": "jane@example.com",
		"age": 25,
	}
	response7d, err := server.RoundTripPost(validPayload)
	if err != nil {
		log.Printf("Valid payload test failed: %v", err)
	} else {
		printResponse("Validator (valid)", response7d)
	}
	
	fmt.Println()
	fmt.Println("All examples completed!")
	fmt.Println()
	fmt.Println("To test different processors, restart the receiver with:")
	fmt.Println("  go run receiver.go hello      # Hello World processor")
	fmt.Println("  go run receiver.go echo       # Echo processor (default)")
	fmt.Println("  go run receiver.go timestamp  # Timestamp processor")
	fmt.Println("  go run receiver.go counter    # Counter processor")
	fmt.Println("  go run receiver.go advanced   # Advanced Context processor")
	fmt.Println("  go run receiver.go transform  # Transform processor")
	fmt.Println("  go run receiver.go validator  # Validator processor")
	fmt.Println("  go run receiver.go chain      # Chain processor")
}

// result holds the result of a concurrent request
type result struct {
	ID       int
	Response *post2post.RoundTripResponse
	Error    error
}

// printResponse prints the response in a formatted way
func printResponse(label string, response *post2post.RoundTripResponse) {
	fmt.Printf("%s Result:\n", label)
	fmt.Printf("  Success: %v\n", response.Success)
	fmt.Printf("  Timeout: %v\n", response.Timeout)
	fmt.Printf("  Request ID: %s\n", response.RequestID)
	
	if response.Error != "" {
		fmt.Printf("  Error: %s\n", response.Error)
	}
	
	if response.Success && response.Payload != nil {
		fmt.Printf("  Response Payload:\n")
		if payloadMap, ok := response.Payload.(map[string]interface{}); ok {
			for key, value := range payloadMap {
				fmt.Printf("    %s: %v\n", key, value)
			}
		} else {
			fmt.Printf("    %+v\n", response.Payload)
		}
	}
	fmt.Println()
}