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
		WithPostURL("http://localhost:8081/webhook").       // Receiver server endpoint
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
	fmt.Println("All examples completed!")
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