package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// RequestData represents the incoming request structure
type RequestData struct {
	URL       string      `json:"url"`
	Payload   interface{} `json:"payload"`
	RequestID string      `json:"request_id"`
}

// ResponseData represents the response structure we send back
type ResponseData struct {
	RequestID string      `json:"request_id"`
	Payload   interface{} `json:"payload"`
}

// EnhancedPayload represents the enhanced payload with timestamp
type EnhancedPayload struct {
	OriginalData interface{} `json:"original_data"`
	Timestamp    string      `json:"timestamp"`
	ProcessedBy  string      `json:"processed_by"`
	Status       string      `json:"status"`
}

func main() {
	http.HandleFunc("/webhook", webhookHandler)
	http.HandleFunc("/", rootHandler)

	port := ":8080"
	fmt.Printf("Receiving server starting on port %s\n", port)
	fmt.Println("Send POST requests to http://localhost:8080/webhook")
	log.Fatal(http.ListenAndServe(port, nil))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	response := `Receiving Server

This server listens for POST requests at /webhook and responds back to the sender.

Usage:
1. Send POST request to /webhook with JSON containing:
   - url: The callback URL to post response back to
   - payload: Your data
   - request_id: Unique identifier for the request

2. Server will process the data and POST back to the provided URL with:
   - Enhanced payload including timestamp and processing info
   - Same request_id for matching

Example request:
{
  "url": "http://localhost:3000/roundtrip",
  "payload": {"message": "hello world"},
  "request_id": "req_123456"
}
`
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Parse the incoming JSON
	var requestData RequestData
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		log.Printf("Error parsing JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Received request from %s with ID: %s", requestData.URL, requestData.RequestID)
	log.Printf("Original payload: %+v", requestData.Payload)

	// Create enhanced payload with timestamp
	enhancedPayload := EnhancedPayload{
		OriginalData: requestData.Payload,
		Timestamp:    time.Now().Format("2006-01-02 15:04:05 MST"),
		ProcessedBy:  "post2post-receiver",
		Status:       "processed",
	}

	// Create response data
	responseData := ResponseData{
		RequestID: requestData.RequestID,
		Payload:   enhancedPayload,
	}

	// Marshal response to JSON
	responseJSON, err := json.Marshal(responseData)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	// Acknowledge the original request
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "received", "message": "Processing request"}`))

	// Post back to the caller in a separate goroutine
	go func() {
		// Add a small delay to simulate processing time
		time.Sleep(100 * time.Millisecond)

		log.Printf("Posting response back to: %s", requestData.URL)
		resp, err := http.Post(requestData.URL, "application/json", bytes.NewBuffer(responseJSON))
		if err != nil {
			log.Printf("Error posting response back: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			log.Printf("Successfully posted response back for request ID: %s", requestData.RequestID)
		} else {
			log.Printf("Failed to post response back, status: %d for request ID: %s", resp.StatusCode, requestData.RequestID)
		}
	}()
}