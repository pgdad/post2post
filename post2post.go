package post2post

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2/clientcredentials"
	"tailscale.com/client/tailscale"
)

// Server represents a configurable web server
type Server struct {
	network         string
	iface           string
	port            int
	listener        net.Listener
	server          *http.Server
	mu              sync.RWMutex
	running         bool
	postURL         string
	client          *http.Client
	roundTripChans  map[string]chan *RoundTripResponse
	defaultTimeout  time.Duration
	processor       PayloadProcessor
}

// PostData represents the JSON payload structure
type PostData struct {
	URL        string      `json:"url"`
	Payload    interface{} `json:"payload"`
	RequestID  string      `json:"request_id,omitempty"`
	TailnetKey string      `json:"tailnet_key,omitempty"`
}

// RoundTripResponse represents the response from a round trip post
type RoundTripResponse struct {
	Payload   interface{} `json:"payload"`
	Success   bool        `json:"success"`
	Error     string      `json:"error,omitempty"`
	Timeout   bool        `json:"timeout"`
	RequestID string      `json:"request_id,omitempty"`
}

// PayloadProcessor defines the interface for processing incoming payloads
type PayloadProcessor interface {
	Process(payload interface{}, requestID string) (interface{}, error)
}

// ProcessorContext provides context information for payload processing
type ProcessorContext struct {
	RequestID   string
	URL         string
	TailnetKey  string
	ReceivedAt  time.Time
}

// AdvancedPayloadProcessor defines an interface for processors that need access to context
type AdvancedPayloadProcessor interface {
	ProcessWithContext(payload interface{}, context ProcessorContext) (interface{}, error)
}

// NewServer creates a new server instance with default settings
func NewServer() *Server {
	return &Server{
		network:        "tcp4",
		iface:          "",
		port:           0,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		roundTripChans: make(map[string]chan *RoundTripResponse),
		defaultTimeout: 30 * time.Second,
	}
}

// WithNetwork sets the network type (tcp4 or tcp6)
func (s *Server) WithNetwork(network string) *Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if network == "tcp4" || network == "tcp6" {
		s.network = network
	}
	return s
}

// WithInterface sets the interface to listen on
func (s *Server) WithInterface(iface string) *Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.iface = iface
	return s
}

// WithPostURL sets the URL for posting JSON data
func (s *Server) WithPostURL(url string) *Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.postURL = url
	return s
}

// WithTimeout sets the default timeout for round trip posts
func (s *Server) WithTimeout(timeout time.Duration) *Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.defaultTimeout = timeout
	return s
}

// WithProcessor sets a custom payload processor
func (s *Server) WithProcessor(processor PayloadProcessor) *Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.processor = processor
	return s
}

// Start starts the server
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.running {
		return fmt.Errorf("server is already running")
	}
	
	addr := fmt.Sprintf("%s:%d", s.iface, s.port)
	
	listener, err := net.Listen(s.network, addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	
	s.listener = listener
	
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.defaultHandler)
	mux.HandleFunc("/roundtrip", s.roundTripHandler)
	mux.HandleFunc("/webhook", s.webhookHandler)
	
	s.server = &http.Server{
		Handler: mux,
	}
	
	// Extract the actual port from the listener
	if tcpAddr, ok := listener.Addr().(*net.TCPAddr); ok {
		s.port = tcpAddr.Port
	}
	
	log.Printf("Server starting on %s network, interface: %s, port: %d", s.network, s.iface, s.port)
	log.Printf("Server listening on: %s", listener.Addr().String())
	log.Printf("Server available routes: /, /roundtrip, /webhook")
	
	s.running = true
	
	go func() {
		log.Printf("HTTP server goroutine starting...")
		if err := s.server.Serve(listener); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
		log.Printf("HTTP server goroutine finished")
	}()
	
	return nil
}

// Stop stops the server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.running {
		return fmt.Errorf("server is not running")
	}
	
	s.running = false
	
	if s.server != nil {
		s.server.Close()
	}
	
	if s.listener != nil {
		s.listener.Close()
	}
	
	return nil
}

// GetPort returns the port the server is listening on
func (s *Server) GetPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.port
}

// GetInterface returns the interface the server is listening on
func (s *Server) GetInterface() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.iface == "" {
		return "localhost"
	}
	return s.iface
}

// IsRunning returns whether the server is currently running
func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.running
}

// GetNetwork returns the network type (tcp4 or tcp6)
func (s *Server) GetNetwork() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.network
}

// GetURL returns the full URL for the server
func (s *Server) GetURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	scheme := "http"
	host := s.GetInterface()
	if host == "localhost" && s.iface == "" {
		host = "localhost"
	}
	return fmt.Sprintf("%s://%s:%d", scheme, host, s.port)
}

// GetPostURL returns the configured post URL
func (s *Server) GetPostURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.postURL
}

// GetTailscaleURL returns the full URL for the server using Tailscale hostname
func (s *Server) GetTailscaleURL() (string, error) {
	s.mu.RLock()
	port := s.port
	s.mu.RUnlock()
	
	// Get Tailscale status to find our hostname
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	client := &tailscale.LocalClient{}
	status, err := client.Status(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get Tailscale status: %w", err)
	}
	
	if status.Self == nil {
		return "", fmt.Errorf("Tailscale not connected or no self node found")
	}
	
	// Use the Tailscale hostname (machine name + tailnet domain)
	hostname := status.Self.DNSName
	if hostname == "" {
		return "", fmt.Errorf("no Tailscale hostname available")
	}
	
	// Remove trailing dot if present
	hostname = strings.TrimSuffix(hostname, ".")
	
	return fmt.Sprintf("http://%s:%d", hostname, port), nil
}

// GetTailscaleIP returns the Tailscale IP address for binding interfaces
func (s *Server) GetTailscaleIP() (string, error) {
	// Get Tailscale status to find our IP address
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	client := &tailscale.LocalClient{}
	status, err := client.Status(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get Tailscale status: %w", err)
	}
	
	if status.Self == nil {
		return "", fmt.Errorf("Tailscale not connected or no self node found")
	}
	
	// Get the first Tailscale IP address
	if len(status.Self.TailscaleIPs) == 0 {
		return "", fmt.Errorf("no Tailscale IP addresses available")
	}
	
	// Use the first IP address (usually IPv4)
	tailscaleIP := status.Self.TailscaleIPs[0].String()
	return tailscaleIP, nil
}

// PostJSON posts JSON data to the configured URL with server URL and payload
func (s *Server) PostJSON(payload interface{}) error {
	return s.PostJSONWithTailnet(payload, "")
}

// PostJSONWithTailnet posts JSON data using an optional Tailscale connection
func (s *Server) PostJSONWithTailnet(payload interface{}, tailnetKey string) error {
	s.mu.RLock()
	postURL := s.postURL
	serverURL := s.GetURL()
	client := s.client
	s.mu.RUnlock()
	
	if postURL == "" {
		return fmt.Errorf("post URL not configured")
	}
	
	if !s.IsRunning() {
		return fmt.Errorf("server is not running")
	}
	
	data := PostData{
		URL:        serverURL,
		Payload:    payload,
		TailnetKey: tailnetKey,
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post JSON: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("post request failed with status: %d", resp.StatusCode)
	}
	
	return nil
}

// RoundTripPost posts JSON data and waits for a response back to the server
func (s *Server) RoundTripPost(payload interface{}, tailnetKey string) (*RoundTripResponse, error) {
	return s.RoundTripPostWithTimeout(payload, tailnetKey, s.defaultTimeout)
}

// RoundTripPostWithTimeout posts JSON data and waits for a response with custom timeout
func (s *Server) RoundTripPostWithTimeout(payload interface{}, tailnetKey string, timeout time.Duration) (*RoundTripResponse, error) {
	s.mu.RLock()
	postURL := s.postURL
	serverURL := s.GetURL()
	client := s.client
	s.mu.RUnlock()
	
	if postURL == "" {
		return nil, fmt.Errorf("post URL not configured")
	}
	
	if !s.IsRunning() {
		return nil, fmt.Errorf("server is not running")
	}
	
	// Extract or generate request ID from payload
	var requestID string
	
	// Try to extract RequestID from payload using reflection
	v := reflect.ValueOf(payload)
	if v.Kind() == reflect.Struct {
		if field := v.FieldByName("RequestID"); field.IsValid() && field.Kind() == reflect.String && field.String() != "" {
			requestID = field.String()
			log.Printf("RoundTripPostWithTimeout: Using payload RequestID: %s", requestID)
		} else {
			// Generate unique request ID if not found in payload
			requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
			log.Printf("RoundTripPostWithTimeout: Generated new RequestID (no RequestID field): %s", requestID)
		}
	} else {
		// Generate unique request ID if payload is not a struct
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
		log.Printf("RoundTripPostWithTimeout: Generated new RequestID (not struct): %s", requestID)
	}
	
	// Create response channel
	responseChan := make(chan *RoundTripResponse, 1)
	s.mu.Lock()
	s.roundTripChans[requestID] = responseChan
	log.Printf("RoundTripPostWithTimeout: Created channel for RequestID: %s, total channels: %d", requestID, len(s.roundTripChans))
	s.mu.Unlock()
	
	// Cleanup function
	defer func() {
		s.mu.Lock()
		delete(s.roundTripChans, requestID)
		close(responseChan)
		log.Printf("RoundTripPostWithTimeout: Cleaned up channel for RequestID: %s, remaining channels: %d", requestID, len(s.roundTripChans))
		s.mu.Unlock()
	}()
	
	// Prepare the data with request ID
	data := PostData{
		URL:       fmt.Sprintf("%s/roundtrip", serverURL),
		Payload:   payload,
		RequestID: requestID,
		TailnetKey: tailnetKey,
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return &RoundTripResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to marshal JSON: %v", err),
			Timeout: false,
		}, nil
	}
	
	log.Printf("RoundTripPostWithTimeout: Sending request to %s with RequestID: %s", postURL, requestID)
	log.Printf("RoundTripPostWithTimeout: JSON DATA: %s", string(jsonData))
	
	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return &RoundTripResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to create request: %v", err),
			Timeout: false,
		}, nil
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// Send the request
	log.Printf("RoundTripPostWithTimeout: Making HTTP request for RequestID: %s", requestID)
	resp, err := client.Do(req)
	if err != nil {
		return &RoundTripResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to post JSON: %v", err),
			Timeout: false,
		}, nil
	}
	resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		log.Printf("RoundTripPostWithTimeout: HTTP request failed with status %d for RequestID: %s", resp.StatusCode, requestID)
		return &RoundTripResponse{
			Success: false,
			Error:   fmt.Sprintf("post request failed with status: %d", resp.StatusCode),
			Timeout: false,
		}, nil
	}
	
	log.Printf("RoundTripPostWithTimeout: HTTP request successful (%d), waiting for response on channel for RequestID: %s", resp.StatusCode, requestID)
	
	// Wait for response or timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	select {
	case response := <-responseChan:
		log.Printf("RoundTripPostWithTimeout: Received response from channel for RequestID: %s", requestID)
		
		// Log the response content for debugging
		if response != nil {
			responseJSON, err := json.Marshal(response)
			if err != nil {
				log.Printf("RoundTripPostWithTimeout: Failed to marshal response for logging: %v", err)
			} else {
				log.Printf("RoundTripPostWithTimeout: Response content: %s", string(responseJSON))
			}
			
			// Also log the payload specifically if it exists
			if response.Payload != nil {
				payloadJSON, err := json.Marshal(response.Payload)
				if err != nil {
					log.Printf("RoundTripPostWithTimeout: Failed to marshal payload for logging: %v", err)
				} else {
					log.Printf("RoundTripPostWithTimeout: Response payload: %s", string(payloadJSON))
				}
			}
		}
		
		return response, nil
	case <-ctx.Done():
		log.Printf("RoundTripPostWithTimeout: Timeout waiting for response for RequestID: %s", requestID)
		return &RoundTripResponse{
			Success:   false,
			Error:     "timeout waiting for response",
			Timeout:   true,
			RequestID: requestID,
		}, nil
	}
}

func (s *Server) GenerateTailnetKeyFromOAuth(reusable bool, ephemeral bool, preauth bool, tags string) (string, error) {
	// Acknowledge the unstable API at package level
	tailscale.I_Acknowledge_This_API_Is_Unstable = true
	
	clientID := os.Getenv("TS_API_CLIENT_ID")
	clientSecret := os.Getenv("TS_API_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("TS_API_CLIENT_ID and TS_API_CLIENT_SECRET must be set")
	}

	if tags == "" {
		return "", fmt.Errorf("at least one tag must be specified")
	}

	baseURL := "https://api.tailscale.com"

	credentials := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     baseURL + "/api/v2/oauth/token",
	}

	ctx := context.Background()
	tsClient := tailscale.NewClient("-", nil)
	tsClient.HTTPClient = credentials.Client(ctx)
	tsClient.BaseURL = baseURL

	caps := tailscale.KeyCapabilities{
		Devices: tailscale.KeyDeviceCapabilities{
			Create: tailscale.KeyDeviceCreateCapabilities{
				Reusable:      reusable,
				Ephemeral:     ephemeral,
				Preauthorized: preauth,
				Tags:          strings.Split(tags, ","),
			},
		},
	}

	authkey, _, err := tsClient.CreateKey(ctx, caps)
	if err != nil {
		return "", fmt.Errorf("failed to create Tailscale auth key: %w", err)
	}

	log.Printf("Generated Tailscale auth key: %s...", authkey[:min(10, len(authkey))])
	return authkey, nil
}


// createTailscaleClient creates an HTTP client that routes through Tailscale
func (s *Server) createTailscaleClient(tailnetKey string) (*http.Client, error) {
	// Framework for Tailscale integration using tsnet
	// 
	// To implement full Tailscale integration, uncomment and modify the following:
	//
	// import "tailscale.com/tsnet"
	//
	// srv := &tsnet.Server{
	//     Hostname: "post2post-server",
	//     AuthKey:  tailnetKey,
	// }
	// 
	// // Start the tsnet server
	// if err := srv.Start(); err != nil {
	//     return nil, fmt.Errorf("failed to start tsnet server: %w", err)
	// }
	//
	// // Create HTTP client that routes through Tailscale
	// client := srv.HTTPClient()
	// return client, nil
	
	// For now, return an informative error with the key for development
	return nil, fmt.Errorf("Tailscale integration is available but requires tsnet configuration with auth key: %s", tailnetKey)
}

// postWithOptionalTailscale makes an HTTP POST request, optionally using Tailscale
func (s *Server) postWithOptionalTailscale(url string, data []byte, tailnetKey string) (*http.Response, error) {
	var client *http.Client
	var err error
	
	if tailnetKey != "" {
		// Use Tailscale client if tailnet_key is provided
		client, err = s.createTailscaleClient(tailnetKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create Tailscale client: %w", err)
		}
	} else {
		// Use regular HTTP client
		s.mu.RLock()
		client = s.client
		s.mu.RUnlock()
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	return client.Do(req)
}

// roundTripHandler handles incoming responses for round trip requests
func (s *Server) roundTripHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("roundTripHandler: Received %s request from %s to %s", r.Method, r.RemoteAddr, r.URL.Path)
	log.Printf("roundTripHandler: Request headers: %+v", r.Header)
	
	if r.Method != "POST" {
		log.Printf("roundTripHandler: Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("roundTripHandler: Failed to read request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	log.Printf("roundTripHandler: Request body: %s", string(body))
	
	var responseData struct {
		RequestID  string      `json:"request_id"`
		Payload    interface{} `json:"payload"`
		TailnetKey string      `json:"tailnet_key,omitempty"`
	}
	
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		log.Printf("roundTripHandler: Failed to unmarshal JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	log.Printf("roundTripHandler: Parsed request - RequestID: %s, TailnetKey: %s", responseData.RequestID, responseData.TailnetKey)
	
	// Find the waiting channel
	s.mu.RLock()
	responseChan, exists := s.roundTripChans[responseData.RequestID]
	
	// Log all current channels for debugging
	log.Printf("roundTripHandler: Looking for RequestID '%s'", responseData.RequestID)
	log.Printf("roundTripHandler: Current channels (%d total):", len(s.roundTripChans))
	for id := range s.roundTripChans {
		log.Printf("roundTripHandler: - Channel exists for RequestID: '%s'", id)
	}
	log.Printf("roundTripHandler: Channel found for RequestID '%s': %v", responseData.RequestID, exists)
	
	s.mu.RUnlock()
	
	if !exists {
		log.Printf("roundTripHandler: No waiting channel found for RequestID: %s", responseData.RequestID)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	
	// Send response to waiting goroutine
	response := &RoundTripResponse{
		Payload:   responseData.Payload,
		Success:   true,
		RequestID: responseData.RequestID,
	}
	
	select {
	case responseChan <- response:
		log.Printf("roundTripHandler: Successfully sent response to waiting channel for RequestID: %s", responseData.RequestID)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Response received"))
	default:
		// Channel might be closed or full
		log.Printf("roundTripHandler: Failed to send response - channel closed or full for RequestID: %s", responseData.RequestID)
		w.WriteHeader(http.StatusGone)
	}
}

// webhookHandler handles incoming webhook requests with configurable processing
func (s *Server) webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	var requestData PostData
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	// Process the payload using the configured processor
	var processedPayload interface{}
	s.mu.RLock()
	processor := s.processor
	s.mu.RUnlock()
	
	if processor != nil {
		// Check if processor supports advanced context
		if advancedProcessor, ok := processor.(AdvancedPayloadProcessor); ok {
			context := ProcessorContext{
				RequestID:  requestData.RequestID,
				URL:        requestData.URL,
				TailnetKey: requestData.TailnetKey,
				ReceivedAt: time.Now(),
			}
			processedPayload, err = advancedProcessor.ProcessWithContext(requestData.Payload, context)
		} else {
			processedPayload, err = processor.Process(requestData.Payload, requestData.RequestID)
		}
		
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Processing error: %v", err)))
			return
		}
	} else {
		// Default processing - just echo back the payload
		processedPayload = requestData.Payload
	}
	
	// Acknowledge the request
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "received", "message": "Processing request"}`))
	
	// Post back the processed response if callback URL is provided
	if requestData.URL != "" {
		go s.postProcessedResponse(requestData.URL, requestData.RequestID, processedPayload, requestData.TailnetKey)
	}
}

// postProcessedResponse posts the processed response back to the callback URL
func (s *Server) postProcessedResponse(callbackURL, requestID string, payload interface{}, tailnetKey string) {
	// Add a small delay to simulate processing time
	time.Sleep(100 * time.Millisecond)
	
	responseData := map[string]interface{}{
		"request_id": requestID,
		"payload":    payload,
	}
	
	// Include tailnet_key if it was provided
	if tailnetKey != "" {
		responseData["tailnet_key"] = tailnetKey
	}
	
	responseJSON, err := json.Marshal(responseData)
	if err != nil {
		return
	}
	
	// Use appropriate HTTP client based on tailnet_key
	if tailnetKey != "" {
		s.postWithOptionalTailscale(callbackURL, responseJSON, tailnetKey)
	} else {
		s.mu.RLock()
		client := s.client
		s.mu.RUnlock()
		
		resp, err := client.Post(callbackURL, "application/json", bytes.NewBuffer(responseJSON))
		if err == nil {
			resp.Body.Close()
		}
	}
}

// defaultHandler is a simple HTTP handler that returns server information
func (s *Server) defaultHandler(w http.ResponseWriter, r *http.Request) {
	response := fmt.Sprintf("post2post server\nListening on: %s:%d\nNetwork: %s\nPath: %s\n", 
		s.GetInterface(), s.GetPort(), s.GetNetwork(), r.URL.Path)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}
