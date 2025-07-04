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