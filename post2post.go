package post2post

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
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
	
	s.server = &http.Server{
		Handler: mux,
	}
	
	// Extract the actual port from the listener
	if tcpAddr, ok := listener.Addr().(*net.TCPAddr); ok {
		s.port = tcpAddr.Port
	}
	
	s.running = true
	
	go func() {
		s.server.Serve(listener)
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
func (s *Server) RoundTripPost(payload interface{}) (*RoundTripResponse, error) {
	return s.RoundTripPostWithTimeout(payload, s.defaultTimeout)
}

// RoundTripPostWithTimeout posts JSON data and waits for a response with custom timeout
func (s *Server) RoundTripPostWithTimeout(payload interface{}, timeout time.Duration) (*RoundTripResponse, error) {
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
	
	// Generate unique request ID
	requestID := fmt.Sprintf("req_%d", time.Now().UnixNano())
	
	// Create response channel
	responseChan := make(chan *RoundTripResponse, 1)
	s.mu.Lock()
	s.roundTripChans[requestID] = responseChan
	s.mu.Unlock()
	
	// Cleanup function
	defer func() {
		s.mu.Lock()
		delete(s.roundTripChans, requestID)
		close(responseChan)
		s.mu.Unlock()
	}()
	
	// Prepare the data with request ID
	data := PostData{
		URL:       fmt.Sprintf("%s/roundtrip", serverURL),
		Payload:   payload,
		RequestID: requestID,
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return &RoundTripResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to marshal JSON: %v", err),
			Timeout: false,
		}, nil
	}
	
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
		return &RoundTripResponse{
			Success: false,
			Error:   fmt.Sprintf("post request failed with status: %d", resp.StatusCode),
			Timeout: false,
		}, nil
	}
	
	// Wait for response or timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	select {
	case response := <-responseChan:
		return response, nil
	case <-ctx.Done():
		return &RoundTripResponse{
			Success:   false,
			Error:     "timeout waiting for response",
			Timeout:   true,
			RequestID: requestID,
		}, nil
	}
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
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	var responseData struct {
		RequestID  string      `json:"request_id"`
		Payload    interface{} `json:"payload"`
		TailnetKey string      `json:"tailnet_key,omitempty"`
	}
	
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	// Find the waiting channel
	s.mu.RLock()
	responseChan, exists := s.roundTripChans[responseData.RequestID]
	s.mu.RUnlock()
	
	if !exists {
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
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Response received"))
	default:
		// Channel might be closed or full
		w.WriteHeader(http.StatusGone)
	}
}

// defaultHandler is a simple HTTP handler that returns server information
func (s *Server) defaultHandler(w http.ResponseWriter, r *http.Request) {
	response := fmt.Sprintf("post2post server\nListening on: %s:%d\nNetwork: %s\nPath: %s\n", 
		s.GetInterface(), s.GetPort(), s.GetNetwork(), r.URL.Path)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}