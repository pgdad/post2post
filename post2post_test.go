package post2post

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	server := NewServer()
	
	if server.GetNetwork() != "tcp4" {
		t.Errorf("NewServer() network = %v, want tcp4", server.GetNetwork())
	}
	
	if server.GetInterface() != "localhost" {
		t.Errorf("NewServer() interface = %v, want localhost", server.GetInterface())
	}
	
	if server.GetPort() != 0 {
		t.Errorf("NewServer() port = %v, want 0", server.GetPort())
	}
	
	if server.IsRunning() {
		t.Error("NewServer() should not be running initially")
	}
}

func TestServerConfiguration(t *testing.T) {
	server := NewServer().
		WithNetwork("tcp6").
		WithInterface("127.0.0.1")
	
	if server.GetNetwork() != "tcp6" {
		t.Errorf("WithNetwork() = %v, want tcp6", server.GetNetwork())
	}
	
	if server.GetInterface() != "127.0.0.1" {
		t.Errorf("WithInterface() = %v, want 127.0.0.1", server.GetInterface())
	}
}

func TestServerStartStop(t *testing.T) {
	server := NewServer()
	
	// Test start
	err := server.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	
	if !server.IsRunning() {
		t.Error("Server should be running after Start()")
	}
	
	if server.GetPort() == 0 {
		t.Error("Server port should be assigned after Start()")
	}
	
	// Test that we can't start again
	err = server.Start()
	if err == nil {
		t.Error("Start() should fail when server is already running")
	}
	
	// Test stop
	err = server.Stop()
	if err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}
	
	if server.IsRunning() {
		t.Error("Server should not be running after Stop()")
	}
	
	// Test that we can't stop again
	err = server.Stop()
	if err == nil {
		t.Error("Stop() should fail when server is not running")
	}
}

func TestServerHTTPResponse(t *testing.T) {
	server := NewServer()
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer server.Stop()
	
	// Give the server a moment to start
	time.Sleep(10 * time.Millisecond)
	
	// Test HTTP request
	url := fmt.Sprintf("http://%s:%d/test", server.GetInterface(), server.GetPort())
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("HTTP GET failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("HTTP response status = %v, want %v", resp.StatusCode, http.StatusOK)
	}
}

func TestServerWithCustomInterface(t *testing.T) {
	server := NewServer().WithInterface("127.0.0.1")
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer server.Stop()
	
	if server.GetInterface() != "127.0.0.1" {
		t.Errorf("Custom interface = %v, want 127.0.0.1", server.GetInterface())
	}
}

func TestServerWithTCP6(t *testing.T) {
	server := NewServer().WithNetwork("tcp6")
	
	err := server.Start()
	if err != nil {
		// Skip test if IPv6 is not available
		t.Skipf("IPv6 not available: %v", err)
	}
	defer server.Stop()
	
	if server.GetNetwork() != "tcp6" {
		t.Errorf("Network type = %v, want tcp6", server.GetNetwork())
	}
}

func TestServerInvalidNetwork(t *testing.T) {
	server := NewServer().WithNetwork("invalid")
	
	// Should ignore invalid network and keep default
	if server.GetNetwork() != "tcp4" {
		t.Errorf("Invalid network should be ignored, got %v, want tcp4", server.GetNetwork())
	}
}

func TestConcurrentServerOperations(t *testing.T) {
	server := NewServer()
	
	// Start server
	err := server.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer server.Stop()
	
	// Test concurrent access to server information
	done := make(chan bool, 3)
	
	go func() {
		for i := 0; i < 100; i++ {
			_ = server.GetPort()
			_ = server.GetInterface()
			_ = server.GetNetwork()
			_ = server.IsRunning()
		}
		done <- true
	}()
	
	go func() {
		for i := 0; i < 100; i++ {
			_ = server.GetPort()
			_ = server.GetInterface()
			_ = server.GetNetwork()
			_ = server.IsRunning()
		}
		done <- true
	}()
	
	go func() {
		for i := 0; i < 100; i++ {
			_ = server.GetPort()
			_ = server.GetInterface()
			_ = server.GetNetwork()
			_ = server.IsRunning()
		}
		done <- true
	}()
	
	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}
}

func TestServerGetURL(t *testing.T) {
	server := NewServer()
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer server.Stop()
	
	url := server.GetURL()
	expectedPrefix := "http://localhost:"
	if !strings.HasPrefix(url, expectedPrefix) {
		t.Errorf("GetURL() = %v, want prefix %v", url, expectedPrefix)
	}
	
	// Test with custom interface
	customServer := NewServer().WithInterface("127.0.0.1")
	err = customServer.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer customServer.Stop()
	
	customURL := customServer.GetURL()
	expectedCustomPrefix := "http://127.0.0.1:"
	if !strings.HasPrefix(customURL, expectedCustomPrefix) {
		t.Errorf("GetURL() with custom interface = %v, want prefix %v", customURL, expectedCustomPrefix)
	}
}

func TestServerWithPostURL(t *testing.T) {
	server := NewServer().WithPostURL("http://example.com/webhook")
	
	if server.GetPostURL() != "http://example.com/webhook" {
		t.Errorf("WithPostURL() = %v, want http://example.com/webhook", server.GetPostURL())
	}
}

func TestServerPostJSON(t *testing.T) {
	// Create a test server to receive the POST request
	var receivedData PostData
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}
		
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
			return
		}
		
		err = json.Unmarshal(body, &receivedData)
		if err != nil {
			t.Errorf("Failed to unmarshal JSON: %v", err)
			return
		}
		
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()
	
	// Create our server
	server := NewServer().WithPostURL(testServer.URL)
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer server.Stop()
	
	// Test posting JSON with a map payload
	payload := map[string]interface{}{
		"message": "hello world",
		"count":   42,
		"active":  true,
	}
	
	err = server.PostJSON(payload)
	if err != nil {
		t.Fatalf("PostJSON() failed: %v", err)
	}
	
	// Verify the received data
	if receivedData.URL != server.GetURL() {
		t.Errorf("Received URL = %v, want %v", receivedData.URL, server.GetURL())
	}
	
	payloadMap, ok := receivedData.Payload.(map[string]interface{})
	if !ok {
		t.Errorf("Payload is not a map: %T", receivedData.Payload)
	} else {
		if payloadMap["message"] != "hello world" {
			t.Errorf("Payload message = %v, want hello world", payloadMap["message"])
		}
		if payloadMap["count"] != float64(42) { // JSON numbers are float64
			t.Errorf("Payload count = %v, want 42", payloadMap["count"])
		}
		if payloadMap["active"] != true {
			t.Errorf("Payload active = %v, want true", payloadMap["active"])
		}
	}
}

func TestServerPostJSONWithStruct(t *testing.T) {
	// Define a custom struct for the payload
	type TestPayload struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		Active bool   `json:"active"`
	}
	
	// Create a test server to receive the POST request
	var receivedData PostData
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
			return
		}
		
		err = json.Unmarshal(body, &receivedData)
		if err != nil {
			t.Errorf("Failed to unmarshal JSON: %v", err)
			return
		}
		
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()
	
	// Create our server
	server := NewServer().WithPostURL(testServer.URL)
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer server.Stop()
	
	// Test posting JSON with a struct payload
	payload := TestPayload{
		Name:   "Alice",
		Age:    30,
		Active: true,
	}
	
	err = server.PostJSON(payload)
	if err != nil {
		t.Fatalf("PostJSON() failed: %v", err)
	}
	
	// Verify the received data
	if receivedData.URL != server.GetURL() {
		t.Errorf("Received URL = %v, want %v", receivedData.URL, server.GetURL())
	}
	
	payloadMap, ok := receivedData.Payload.(map[string]interface{})
	if !ok {
		t.Errorf("Payload is not a map: %T", receivedData.Payload)
	} else {
		if payloadMap["name"] != "Alice" {
			t.Errorf("Payload name = %v, want Alice", payloadMap["name"])
		}
		if payloadMap["age"] != float64(30) { // JSON numbers are float64
			t.Errorf("Payload age = %v, want 30", payloadMap["age"])
		}
		if payloadMap["active"] != true {
			t.Errorf("Payload active = %v, want true", payloadMap["active"])
		}
	}
}

func TestServerPostJSONErrors(t *testing.T) {
	server := NewServer()
	
	// Test posting without configuring post URL
	err := server.PostJSON(map[string]string{"test": "data"})
	if err == nil || !strings.Contains(err.Error(), "post URL not configured") {
		t.Errorf("Expected 'post URL not configured' error, got: %v", err)
	}
	
	// Test posting without starting server
	server.WithPostURL("http://example.com/webhook")
	err = server.PostJSON(map[string]string{"test": "data"})
	if err == nil || !strings.Contains(err.Error(), "server is not running") {
		t.Errorf("Expected 'server is not running' error, got: %v", err)
	}
	
	// Test posting to invalid URL
	server.WithPostURL("invalid-url")
	server.Start()
	defer server.Stop()
	
	err = server.PostJSON(map[string]string{"test": "data"})
	if err == nil {
		t.Error("Expected error when posting to invalid URL")
	}
}

func TestServerPostJSONHTTPError(t *testing.T) {
	// Create a test server that returns an error
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer testServer.Close()
	
	server := NewServer().WithPostURL(testServer.URL)
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer server.Stop()
	
	err = server.PostJSON(map[string]string{"test": "data"})
	if err == nil || !strings.Contains(err.Error(), "post request failed with status: 500") {
		t.Errorf("Expected HTTP 500 error, got: %v", err)
	}
}

func TestServerWithTimeout(t *testing.T) {
	timeout := 10 * time.Second
	server := NewServer().WithTimeout(timeout)
	
	// We can't directly access defaultTimeout, but we can test via round trip timeout
	if server.defaultTimeout != timeout {
		t.Errorf("WithTimeout() did not set timeout correctly")
	}
}

func TestRoundTripPostSuccess(t *testing.T) {
	// Create a test server that will respond back to our server
	var receivedData PostData
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
			return
		}
		
		err = json.Unmarshal(body, &receivedData)
		if err != nil {
			t.Errorf("Failed to unmarshal JSON: %v", err)
			return
		}
		
		// Simulate responding back to the server
		responsePayload := map[string]interface{}{
			"status":  "processed",
			"message": "Round trip successful",
			"data":    receivedData.Payload,
		}
		
		responseData := map[string]interface{}{
			"request_id": receivedData.RequestID,
			"payload":    responsePayload,
		}
		
		responseJSON, _ := json.Marshal(responseData)
		
		// Post back to the server's /roundtrip endpoint
		go func() {
			time.Sleep(100 * time.Millisecond) // Small delay to simulate processing
			http.Post(receivedData.URL, "application/json", bytes.NewBuffer(responseJSON))
		}()
		
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()
	
	// Create our server
	server := NewServer().WithPostURL(testServer.URL)
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer server.Stop()
	
	// Test round trip post
	payload := map[string]interface{}{
		"test":   "round trip",
		"number": 123,
	}
	
	response, err := server.RoundTripPost(payload)
	if err != nil {
		t.Fatalf("RoundTripPost() failed: %v", err)
	}
	
	if !response.Success {
		t.Errorf("RoundTripPost() success = false, want true")
	}
	
	if response.Timeout {
		t.Errorf("RoundTripPost() timeout = true, want false")
	}
	
	if response.Error != "" {
		t.Errorf("RoundTripPost() error = %v, want empty", response.Error)
	}
	
	// Verify the response payload
	payloadMap, ok := response.Payload.(map[string]interface{})
	if !ok {
		t.Errorf("Response payload is not a map: %T", response.Payload)
	} else {
		if payloadMap["status"] != "processed" {
			t.Errorf("Response status = %v, want processed", payloadMap["status"])
		}
		if payloadMap["message"] != "Round trip successful" {
			t.Errorf("Response message = %v, want 'Round trip successful'", payloadMap["message"])
		}
	}
}

func TestRoundTripPostTimeout(t *testing.T) {
	// Create a test server that doesn't respond back
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Just acknowledge the request but don't respond back
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()
	
	// Create our server with short timeout
	server := NewServer().
		WithPostURL(testServer.URL).
		WithTimeout(200 * time.Millisecond)
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer server.Stop()
	
	// Test round trip post that should timeout
	payload := map[string]string{"test": "timeout"}
	
	response, err := server.RoundTripPost(payload)
	if err != nil {
		t.Fatalf("RoundTripPost() failed: %v", err)
	}
	
	if response.Success {
		t.Errorf("RoundTripPost() success = true, want false")
	}
	
	if !response.Timeout {
		t.Errorf("RoundTripPost() timeout = false, want true")
	}
	
	if !strings.Contains(response.Error, "timeout") {
		t.Errorf("RoundTripPost() error = %v, want timeout error", response.Error)
	}
}

func TestRoundTripPostWithCustomTimeout(t *testing.T) {
	// Create a test server that doesn't respond back
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()
	
	// Create our server
	server := NewServer().WithPostURL(testServer.URL)
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer server.Stop()
	
	// Test round trip post with custom short timeout
	payload := map[string]string{"test": "custom timeout"}
	customTimeout := 100 * time.Millisecond
	
	start := time.Now()
	response, err := server.RoundTripPostWithTimeout(payload, customTimeout)
	elapsed := time.Since(start)
	
	if err != nil {
		t.Fatalf("RoundTripPostWithTimeout() failed: %v", err)
	}
	
	if response.Success {
		t.Errorf("RoundTripPostWithTimeout() success = true, want false")
	}
	
	if !response.Timeout {
		t.Errorf("RoundTripPostWithTimeout() timeout = false, want true")
	}
	
	// Check that it actually timed out around the expected time
	if elapsed < customTimeout || elapsed > customTimeout+100*time.Millisecond {
		t.Errorf("RoundTripPostWithTimeout() elapsed = %v, expected around %v", elapsed, customTimeout)
	}
}

func TestRoundTripPostErrors(t *testing.T) {
	server := NewServer()
	
	// Test without configuring post URL
	response, err := server.RoundTripPost(map[string]string{"test": "data"})
	if err == nil || !strings.Contains(err.Error(), "post URL not configured") {
		t.Errorf("Expected 'post URL not configured' error, got: %v", err)
	}
	
	// Test without starting server
	server.WithPostURL("http://example.com/webhook")
	response, err = server.RoundTripPost(map[string]string{"test": "data"})
	if err == nil || !strings.Contains(err.Error(), "server is not running") {
		t.Errorf("Expected 'server is not running' error, got: %v", err)
	}
	
	// Test with invalid URL
	server.WithPostURL("invalid-url")
	server.Start()
	defer server.Stop()
	
	response, err = server.RoundTripPost(map[string]string{"test": "data"})
	if err != nil {
		t.Errorf("Expected response with error, got error: %v", err)
	}
	if response == nil || response.Success {
		t.Error("Expected failed response due to invalid URL")
	}
}

func TestRoundTripHandlerInvalidMethods(t *testing.T) {
	server := NewServer()
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer server.Stop()
	
	// Test GET request to roundtrip endpoint
	url := fmt.Sprintf("http://%s:%d/roundtrip", server.GetInterface(), server.GetPort())
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("HTTP GET failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("GET /roundtrip status = %v, want %v", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestRoundTripHandlerInvalidJSON(t *testing.T) {
	server := NewServer()
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer server.Stop()
	
	// Test POST with invalid JSON
	url := fmt.Sprintf("http://%s:%d/roundtrip", server.GetInterface(), server.GetPort())
	resp, err := http.Post(url, "application/json", strings.NewReader("invalid json"))
	if err != nil {
		t.Fatalf("HTTP POST failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("POST /roundtrip with invalid JSON status = %v, want %v", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestConcurrentRoundTripPosts(t *testing.T) {
	// Create a test server that responds back after different delays
	var mu sync.Mutex
	responses := make(map[string]bool)
	
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var receivedData PostData
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
			return
		}
		
		err = json.Unmarshal(body, &receivedData)
		if err != nil {
			t.Errorf("Failed to unmarshal JSON: %v", err)
			return
		}
		
		mu.Lock()
		responses[receivedData.RequestID] = true
		mu.Unlock()
		
		// Respond back after a small delay
		go func() {
			time.Sleep(50 * time.Millisecond)
			
			responseData := map[string]interface{}{
				"request_id": receivedData.RequestID,
				"payload":    map[string]interface{}{"response": "ok", "id": receivedData.RequestID},
			}
			
			responseJSON, _ := json.Marshal(responseData)
			http.Post(receivedData.URL, "application/json", bytes.NewBuffer(responseJSON))
		}()
		
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()
	
	// Create our server
	server := NewServer().WithPostURL(testServer.URL)
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer server.Stop()
	
	// Start multiple concurrent round trip posts
	const numRequests = 5
	results := make(chan *RoundTripResponse, numRequests)
	errors := make(chan error, numRequests)
	
	for i := 0; i < numRequests; i++ {
		go func(id int) {
			payload := map[string]interface{}{
				"request": id,
				"test":    "concurrent",
			}
			
			response, err := server.RoundTripPost(payload)
			if err != nil {
				errors <- err
				return
			}
			results <- response
		}(i)
	}
	
	// Collect results
	successCount := 0
	for i := 0; i < numRequests; i++ {
		select {
		case response := <-results:
			if response.Success {
				successCount++
			}
		case err := <-errors:
			t.Errorf("Concurrent round trip failed: %v", err)
		case <-time.After(5 * time.Second):
			t.Error("Timeout waiting for concurrent round trip responses")
		}
	}
	
	if successCount != numRequests {
		t.Errorf("Expected %d successful responses, got %d", numRequests, successCount)
	}
}

func TestPostJSONWithTailnet(t *testing.T) {
	// Create a test server to receive the POST request
	var receivedData PostData
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
			return
		}
		
		err = json.Unmarshal(body, &receivedData)
		if err != nil {
			t.Errorf("Failed to unmarshal JSON: %v", err)
			return
		}
		
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()
	
	// Create our server
	server := NewServer().WithPostURL(testServer.URL)
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer server.Stop()
	
	// Test posting JSON with tailnet key
	payload := map[string]interface{}{
		"message": "test with tailnet",
		"data":    "some data",
	}
	
	err = server.PostJSONWithTailnet(payload, "test-auth-key")
	if err != nil {
		t.Fatalf("PostJSONWithTailnet() failed: %v", err)
	}
	
	// Verify the received data includes tailnet_key
	if receivedData.TailnetKey != "test-auth-key" {
		t.Errorf("TailnetKey = %v, want test-auth-key", receivedData.TailnetKey)
	}
	
	if receivedData.URL != server.GetURL() {
		t.Errorf("URL = %v, want %v", receivedData.URL, server.GetURL())
	}
}

func TestTailscaleClientCreation(t *testing.T) {
	server := NewServer()
	
	// Test that Tailscale client creation returns expected error
	_, err := server.createTailscaleClient("test-key")
	if err == nil {
		t.Error("Expected error from createTailscaleClient, got nil")
	}
	
	if !strings.Contains(err.Error(), "test-key") {
		t.Errorf("Error should contain the auth key, got: %v", err)
	}
	
	if !strings.Contains(err.Error(), "tsnet configuration") {
		t.Errorf("Error should mention tsnet configuration, got: %v", err)
	}
}

func TestPostWithOptionalTailscale(t *testing.T) {
	server := NewServer()
	
	// Test with empty tailnet key (should use regular client but will fail due to invalid URL)
	_, err := server.postWithOptionalTailscale("invalid-url", []byte("test"), "")
	if err == nil {
		t.Error("Expected error with invalid URL")
	}
	
	// Test with tailnet key (should fail with Tailscale setup error)
	_, err = server.postWithOptionalTailscale("http://example.com", []byte("test"), "auth-key")
	if err == nil {
		t.Error("Expected error from Tailscale client creation")
	}
	
	if !strings.Contains(err.Error(), "failed to create Tailscale client") {
		t.Errorf("Error should mention Tailscale client creation, got: %v", err)
	}
}

func TestServerWithProcessor(t *testing.T) {
	processor := &HelloWorldProcessor{}
	server := NewServer().WithProcessor(processor)
	
	// Access the processor field to verify it was set
	server.mu.RLock()
	setProcessor := server.processor
	server.mu.RUnlock()
	
	if setProcessor != processor {
		t.Error("WithProcessor() did not set the processor correctly")
	}
}

func TestWebhookHandlerWithoutProcessor(t *testing.T) {
	server := NewServer()
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer server.Stop()
	
	// Test POST to webhook endpoint without processor (should echo)
	testPayload := map[string]interface{}{
		"message": "test webhook",
		"data":    "some data",
	}
	
	postData := PostData{
		URL:       fmt.Sprintf("%s/roundtrip", server.GetURL()),
		Payload:   testPayload,
		RequestID: "test_req_123",
	}
	
	jsonData, _ := json.Marshal(postData)
	
	url := fmt.Sprintf("http://%s:%d/webhook", server.GetInterface(), server.GetPort())
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Webhook POST failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Webhook response status = %v, want %v", resp.StatusCode, http.StatusOK)
	}
}

func TestWebhookHandlerWithHelloWorldProcessor(t *testing.T) {
	processor := &HelloWorldProcessor{}
	server := NewServer().WithProcessor(processor)
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer server.Stop()
	
	// Create a test server to receive the processed response
	var receivedResponse map[string]interface{}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedResponse)
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()
	
	// Test POST to webhook endpoint with Hello World processor
	testPayload := map[string]interface{}{
		"message": "original message",
		"data":    42,
	}
	
	postData := PostData{
		URL:       testServer.URL,
		Payload:   testPayload,
		RequestID: "test_hello_123",
	}
	
	jsonData, _ := json.Marshal(postData)
	
	url := fmt.Sprintf("http://%s:%d/webhook", server.GetInterface(), server.GetPort())
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Webhook POST failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Webhook response status = %v, want %v", resp.StatusCode, http.StatusOK)
	}
	
	// Wait a moment for the async response
	time.Sleep(200 * time.Millisecond)
	
	// Verify the processed response
	if receivedResponse["request_id"] != "test_hello_123" {
		t.Errorf("Response request_id = %v, want test_hello_123", receivedResponse["request_id"])
	}
	
	if payload, ok := receivedResponse["payload"].(map[string]interface{}); ok {
		if payload["message"] != "Hello World" {
			t.Errorf("Processed message = %v, want Hello World", payload["message"])
		}
	} else {
		t.Error("Response payload is not a map")
	}
}

func TestWebhookHandlerInvalidMethods(t *testing.T) {
	server := NewServer()
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer server.Stop()
	
	// Test GET request to webhook endpoint
	url := fmt.Sprintf("http://%s:%d/webhook", server.GetInterface(), server.GetPort())
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("HTTP GET failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("GET /webhook status = %v, want %v", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestHelloWorldProcessor(t *testing.T) {
	processor := &HelloWorldProcessor{}
	
	result, err := processor.Process("any payload", "test_123")
	if err != nil {
		t.Fatalf("Process() failed: %v", err)
	}
	
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not a map: %T", result)
	}
	
	if resultMap["message"] != "Hello World" {
		t.Errorf("Message = %v, want Hello World", resultMap["message"])
	}
	
	if resultMap["request_id"] != "test_123" {
		t.Errorf("Request ID = %v, want test_123", resultMap["request_id"])
	}
}

func TestEchoProcessor(t *testing.T) {
	processor := &EchoProcessor{}
	
	testPayload := map[string]interface{}{
		"test": "data",
		"num":  42,
	}
	
	result, err := processor.Process(testPayload, "echo_test")
	if err != nil {
		t.Fatalf("Process() failed: %v", err)
	}
	
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not a map: %T", result)
	}
	
	if resultMap["processor"] != "echo" {
		t.Errorf("Processor = %v, want echo", resultMap["processor"])
	}
	
	originalPayload := resultMap["original_payload"].(map[string]interface{})
	if originalPayload["test"] != "data" {
		t.Errorf("Original payload test = %v, want data", originalPayload["test"])
	}
}

func TestCounterProcessor(t *testing.T) {
	processor := NewCounterProcessor()
	
	// Test multiple calls to verify counter increments
	for i := 1; i <= 3; i++ {
		result, err := processor.Process("test", fmt.Sprintf("req_%d", i))
		if err != nil {
			t.Fatalf("Process() call %d failed: %v", i, err)
		}
		
		resultMap := result.(map[string]interface{})
		count := int(resultMap["count"].(int))
		if count != i {
			t.Errorf("Call %d: count = %v, want %d", i, count, i)
		}
	}
}

func TestAdvancedContextProcessor(t *testing.T) {
	processor := NewAdvancedContextProcessor("test-service")
	
	context := ProcessorContext{
		RequestID:  "ctx_test_123",
		URL:        "http://test.example.com/callback",
		TailnetKey: "test-tailnet-key",
		ReceivedAt: time.Now(),
	}
	
	result, err := processor.ProcessWithContext("test payload", context)
	if err != nil {
		t.Fatalf("ProcessWithContext() failed: %v", err)
	}
	
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not a map: %T", result)
	}
	
	if resultMap["service_name"] != "test-service" {
		t.Errorf("Service name = %v, want test-service", resultMap["service_name"])
	}
	
	contextMap := resultMap["context"].(map[string]interface{})
	if contextMap["request_id"] != "ctx_test_123" {
		t.Errorf("Context request_id = %v, want ctx_test_123", contextMap["request_id"])
	}
	
	// Verify Tailscale info is present
	tailscaleMap := resultMap["tailscale"].(map[string]interface{})
	if tailscaleMap["enabled"] != true {
		t.Errorf("Tailscale enabled = %v, want true", tailscaleMap["enabled"])
	}
}

func TestTransformProcessor(t *testing.T) {
	processor := &TransformProcessor{}
	
	// Test string transformation
	result1, err := processor.Process("hello world", "transform_test")
	if err != nil {
		t.Fatalf("Process() with string failed: %v", err)
	}
	
	resultMap1 := result1.(map[string]interface{})
	if resultMap1["transformed"] != "HELLO WORLD" {
		t.Errorf("Transformed string = %v, want HELLO WORLD", resultMap1["transformed"])
	}
	
	// Test map transformation
	testMap := map[string]interface{}{
		"message": "hello",
		"greeting": "good morning",
		"number": 42,
	}
	
	result2, err := processor.Process(testMap, "transform_test")
	if err != nil {
		t.Fatalf("Process() with map failed: %v", err)
	}
	
	resultMap2 := result2.(map[string]interface{})
	transformedMap := resultMap2["transformed"].(map[string]interface{})
	if transformedMap["message"] != "HELLO" {
		t.Errorf("Transformed message = %v, want HELLO", transformedMap["message"])
	}
	if transformedMap["number"] != 42 {
		t.Errorf("Transformed number = %v, want 42", transformedMap["number"])
	}
}

func TestValidatorProcessor(t *testing.T) {
	processor := NewValidatorProcessor([]string{"name", "email"})
	
	// Test valid payload
	validPayload := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}
	
	result1, err := processor.Process(validPayload, "valid_test")
	if err != nil {
		t.Fatalf("Process() with valid payload failed: %v", err)
	}
	
	resultMap1 := result1.(map[string]interface{})
	validation1 := resultMap1["validation"].(map[string]interface{})
	if validation1["valid"] != true {
		t.Errorf("Valid payload validation = %v, want true", validation1["valid"])
	}
	
	// Test invalid payload
	invalidPayload := map[string]interface{}{
		"name": "Jane Doe",
		// Missing email
		"age": 25,
	}
	
	result2, err := processor.Process(invalidPayload, "invalid_test")
	if err != nil {
		t.Fatalf("Process() with invalid payload failed: %v", err)
	}
	
	resultMap2 := result2.(map[string]interface{})
	validation2 := resultMap2["validation"].(map[string]interface{})
	if validation2["valid"] != false {
		t.Errorf("Invalid payload validation = %v, want false", validation2["valid"])
	}
}

func TestChainProcessor(t *testing.T) {
	// Create a chain of processors
	processor := NewChainProcessor(
		&TimestampProcessor{},
		&EchoProcessor{},
	)
	
	result, err := processor.Process("test chain", "chain_test")
	if err != nil {
		t.Fatalf("Process() chain failed: %v", err)
	}
	
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Chain result is not a map: %T", result)
	}
	
	if resultMap["processor"] != "chain" {
		t.Errorf("Chain processor = %v, want chain", resultMap["processor"])
	}
	
	if resultMap["chain_length"] != 2 {
		t.Errorf("Chain length = %v, want 2", resultMap["chain_length"])
	}
}