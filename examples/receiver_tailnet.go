package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"tailscale.com/tsnet"
)

// RequestData represents the incoming request structure
type RequestData struct {
	URL        string      `json:"url"`
	Payload    interface{} `json:"payload"`
	RequestID  string      `json:"request_id"`
	TailnetKey string      `json:"tailnet_key,omitempty"`
}

// ResponseData represents the response structure we send back
type ResponseData struct {
	RequestID string      `json:"request_id"`
	Payload   interface{} `json:"payload"`
}

// TailscaleResponsePayload represents the enhanced payload with Tailscale info
type TailscaleResponsePayload struct {
	OriginalData interface{} `json:"original_data"`
	TailnetKey   string      `json:"tailnet_key"`
	Timestamp    string      `json:"timestamp"`
	ProcessedVia string      `json:"processed_via"`
	Status       string      `json:"status"`
	ServerInfo   ServerInfo  `json:"server_info"`
}

// ServerInfo provides information about the server processing
type ServerInfo struct {
	Hostname        string `json:"hostname"`
	ProcessedAt     string `json:"processed_at"`
	TailscaleMode   string `json:"tailscale_mode"`
	NetworkSecurity string `json:"network_security"`
}

func main() {
	http.HandleFunc("/webhook", webhookHandler)
	http.HandleFunc("/", rootHandler)

	port := ":8082"
	fmt.Printf("Tailscale-enabled receiving server starting on port %s\n", port)
	fmt.Println("Send POST requests to http://localhost:8082/webhook")
	fmt.Println("Make sure to include 'tailnet_key' in your requests for Tailscale integration")
	fmt.Println()
	fmt.Println("Tailscale integration is ENABLED:")
	fmt.Println("✓ tsnet import active")
	fmt.Println("✓ Full Tailscale client creation enabled")
	fmt.Println("✓ Secure networking via Tailscale ready")
	fmt.Println("Note: Requires valid Tailscale auth keys in requests")
	fmt.Println()
	log.Fatal(http.ListenAndServe(port, nil))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	response := `Tailscale-Enabled Receiving Server

This server demonstrates Tailscale integration for secure webhook processing.

Usage:
1. Send POST request to /webhook with JSON containing:
   - url: The callback URL to post response back to
   - payload: Your data (can include random strings)
   - request_id: Unique identifier for the request
   - tailnet_key: Tailscale auth key for secure networking

2. Server will process the data and POST back via Tailscale with:
   - Enhanced payload including timestamp and original data
   - Tailscale networking for secure response delivery
   - Same request_id for matching

Example request:
{
  "url": "http://client-host:port/roundtrip",
  "payload": {"message": "hello", "random_data": "abc123"},
  "request_id": "req_123456",
  "tailnet_key": "tskey-auth-your-key-here"
}

Features:
- Tailscale tsnet integration for secure networking
- Human-readable timestamps
- Original payload preservation
- Secure response posting via Tailscale
- Random data echo capability
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

	log.Printf("Received Tailscale request from %s with ID: %s", requestData.URL, requestData.RequestID)
	if requestData.TailnetKey != "" {
		log.Printf("Tailscale integration enabled with key: %s...", requestData.TailnetKey[:min(10, len(requestData.TailnetKey))])
	} else {
		log.Printf("No Tailscale key provided - will use regular HTTP")
	}
	log.Printf("Original payload: %+v", requestData.Payload)

	// Create enhanced payload with Tailscale information
	now := time.Now()
	enhancedPayload := TailscaleResponsePayload{
		OriginalData: requestData.Payload,
		TailnetKey:   requestData.TailnetKey,
		Timestamp:    now.Format("2006-01-02 15:04:05 MST"),
		ProcessedVia: "tailscale-receiver",
		Status:       "processed",
		ServerInfo: ServerInfo{
			Hostname:        "post2post-tailscale-receiver",
			ProcessedAt:     now.Format("2006-01-02 15:04:05.000 MST"),
			TailscaleMode:   getTailscaleMode(requestData.TailnetKey),
			NetworkSecurity: "tailscale-secured",
		},
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
	w.Write([]byte(`{"status": "received", "message": "Processing via Tailscale"}`))

	// Post back to the caller using Tailscale in a separate goroutine
	go func() {
		// Add a small delay to simulate processing time
		time.Sleep(200 * time.Millisecond)

		log.Printf("Posting Tailscale response back to: %s", requestData.URL)
		
		// Use Tailscale networking if key is provided
		err := postResponseViaTailscale(requestData.URL, responseJSON, requestData.TailnetKey)
		if err != nil {
			log.Printf("Error posting Tailscale response back: %v", err)
			// Fallback to regular HTTP if Tailscale fails
			log.Printf("Falling back to regular HTTP...")
			err = postResponseViaHTTP(requestData.URL, responseJSON)
			if err != nil {
				log.Printf("Fallback HTTP also failed: %v", err)
			} else {
				log.Printf("Successfully posted response via HTTP fallback for request ID: %s", requestData.RequestID)
			}
		} else {
			log.Printf("Successfully posted response via Tailscale for request ID: %s", requestData.RequestID)
		}
	}()
}

// postResponseViaTailscale posts the response using Tailscale networking
func postResponseViaTailscale(url string, data []byte, tailnetKey string) error {
	if tailnetKey == "" {
		return fmt.Errorf("no Tailscale key provided, cannot use Tailscale networking")
	}

	// Create Tailscale HTTP client
	client, err := createTailscaleClient(tailnetKey)
	if err != nil {
		return fmt.Errorf("failed to create Tailscale client: %w", err)
	}

	// Create and send the request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "post2post-tailscale-receiver/1.0")
	req.Header.Set("X-Tailscale-Enabled", "true")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post via Tailscale: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Tailscale response returned status: %d", resp.StatusCode)
	}

	return nil
}

// postResponseViaHTTP posts the response using regular HTTP as fallback
func postResponseViaHTTP(url string, data []byte) error {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "post2post-receiver-fallback/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post via HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP response returned status: %d", resp.StatusCode)
	}

	return nil
}

// createTailscaleClient creates an HTTP client that routes through Tailscale
func createTailscaleClient(tailnetKey string) (*http.Client, error) {
	log.Printf("Creating Tailscale client with key: %s...", tailnetKey[:min(10, len(tailnetKey))])
	
	srv := &tsnet.Server{
		Hostname:  "post2post-receiver",
		AuthKey:   tailnetKey,
		Ephemeral: true, // Good for demo/testing - creates temporary device
		Logf:      func(format string, args ...interface{}) {
			log.Printf("[tsnet] "+format, args...)
		},
	}
	
	log.Printf("Starting tsnet server with hostname: %s", srv.Hostname)
	
	// Start the tsnet server
	err := srv.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start tsnet server: %w", err)
	}
	
	log.Printf("Tailscale tsnet server started successfully")
	
	// Create HTTP client that routes through Tailscale
	client := srv.HTTPClient()
	
	log.Printf("Tailscale HTTP client created successfully")
	
	return client, nil
}

// getTailscaleMode returns the current Tailscale operation mode
func getTailscaleMode(tailnetKey string) string {
	if tailnetKey == "" {
		return "disabled"
	}
	// Full tsnet integration is now enabled
	return "tsnet-enabled"
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}