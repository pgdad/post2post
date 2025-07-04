package post2post

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// Server represents a configurable web server
type Server struct {
	network   string
	iface     string
	port      int
	listener  net.Listener
	server    *http.Server
	mu        sync.RWMutex
	running   bool
	postURL   string
	client    *http.Client
}

// PostData represents the JSON payload structure
type PostData struct {
	URL     string      `json:"url"`
	Payload interface{} `json:"payload"`
}

// NewServer creates a new server instance with default settings
func NewServer() *Server {
	return &Server{
		network: "tcp4",
		iface:   "",
		port:    0,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
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
	s.server = &http.Server{
		Handler: http.HandlerFunc(s.defaultHandler),
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
		URL:     serverURL,
		Payload: payload,
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

// defaultHandler is a simple HTTP handler that returns server information
func (s *Server) defaultHandler(w http.ResponseWriter, r *http.Request) {
	response := fmt.Sprintf("post2post server\nListening on: %s:%d\nNetwork: %s\nPath: %s\n", 
		s.GetInterface(), s.GetPort(), s.GetNetwork(), r.URL.Path)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}