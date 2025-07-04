# post2post

A simple Go library for starting and managing local web servers with configurable network options.

## Installation

```bash
go get github.com/pgdad/post2post
```

## Usage

```go
package main

import (
    "fmt"
    "log"
    "github.com/pgdad/post2post"
)

func main() {
    // Create a server with default settings (TCP4, random port, all interfaces)
    server := post2post.NewServer()
    
    // Start the server
    err := server.Start()
    if err != nil {
        log.Fatal(err)
    }
    defer server.Stop()
    
    // Get server information
    fmt.Printf("Server listening at: %s\n", server.GetURL())
    
    // Create a server with custom settings and POST URL
    customServer := post2post.NewServer().
        WithNetwork("tcp6").
        WithInterface("127.0.0.1").
        WithPostURL("https://webhook.site/your-unique-url")
    
    err = customServer.Start()
    if err != nil {
        log.Fatal(err)
    }
    defer customServer.Stop()
    
    fmt.Printf("Custom server listening at: %s\n", customServer.GetURL())
    
    // Post JSON data with server URL and custom payload
    payload := map[string]interface{}{
        "message": "Hello from server",
        "timestamp": "2023-12-01T10:00:00Z",
        "data": map[string]string{
            "key1": "value1",
            "key2": "value2",
        },
    }
    
    err = customServer.PostJSON(payload)
    if err != nil {
        log.Printf("Failed to post JSON: %v", err)
    } else {
        fmt.Println("JSON posted successfully!")
    }
    
    // Round trip post: send data and wait for response
    roundTripPayload := map[string]interface{}{
        "action": "process_data",
        "data":   []int{1, 2, 3, 4, 5},
    }
    
    response, err := server.RoundTripPost(roundTripPayload)
    if err != nil {
        log.Printf("Round trip failed: %v", err)
    } else if response.Success {
        fmt.Printf("Round trip successful! Response: %+v\n", response.Payload)
    } else if response.Timeout {
        fmt.Println("Round trip timed out")
    } else {
        fmt.Printf("Round trip failed: %s\n", response.Error)
    }
}
```

## API

### Server Creation and Configuration

#### `NewServer() *Server`
Creates a new server instance with default settings (TCP4, random port, all interfaces).

#### `(*Server) WithNetwork(network string) *Server`
Sets the network type ("tcp4" or "tcp6"). Default is "tcp4".

#### `(*Server) WithInterface(iface string) *Server`
Sets the interface to listen on. Default is "" (all interfaces).

#### `(*Server) WithPostURL(url string) *Server`
Sets the URL for posting JSON data with server information and payload.

#### `(*Server) WithTimeout(timeout time.Duration) *Server`
Sets the default timeout for round trip posts. Default is 30 seconds.

### Server Lifecycle

#### `(*Server) Start() error`
Starts the server on the configured network and interface.

#### `(*Server) Stop() error`
Stops the server.

#### `(*Server) IsRunning() bool`
Returns whether the server is currently running.

### Server Information

#### `(*Server) GetPort() int`
Returns the port the server is listening on.

#### `(*Server) GetInterface() string`
Returns the interface the server is listening on ("localhost" if not specified).

#### `(*Server) GetNetwork() string`
Returns the network type (tcp4 or tcp6).

#### `(*Server) GetURL() string`
Returns the full HTTP URL for the server (e.g., "http://localhost:8080").

#### `(*Server) GetPostURL() string`
Returns the configured POST URL for JSON data.

### JSON Posting

#### `(*Server) PostJSON(payload interface{}) error`
Posts JSON data to the configured URL. The posted data includes:
- `url`: The server's full URL
- `payload`: The provided generic payload (can be any JSON-marshallable type)

The payload can be any Go type that can be marshaled to JSON:
- `map[string]interface{}`
- Custom structs with JSON tags
- Slices, arrays, primitive types, etc.

Example posted JSON structure:
```json
{
  "url": "http://localhost:8080/roundtrip",
  "payload": {
    "message": "Hello World",
    "count": 42,
    "active": true
  },
  "request_id": "req_1234567890"
}
```

### Round Trip Posting

#### `(*Server) RoundTripPost(payload interface{}) (*RoundTripResponse, error)`
Posts JSON data to the configured URL and waits for a response back to the server within the default timeout (30 seconds).

#### `(*Server) RoundTripPostWithTimeout(payload interface{}, timeout time.Duration) (*RoundTripResponse, error)`
Posts JSON data and waits for a response with a custom timeout.

**Round Trip Process:**
1. Posts data with unique request ID to configured URL
2. Waits for external service to process and post response back to server's `/roundtrip` endpoint
3. Returns response payload or timeout/error information

**Response Structure:**
- `Success`: boolean indicating if round trip completed successfully
- `Payload`: the response data from the external service (if successful)
- `Error`: error message (if failed)
- `Timeout`: boolean indicating if the operation timed out
- `RequestID`: unique identifier for the request

Example external service response format:
```json
{
  "request_id": "req_1234567890",
  "payload": {
    "status": "processed",
    "result": "success"
  }
}
```

## Testing

Run tests with:

```bash
go test
```

## License

MIT