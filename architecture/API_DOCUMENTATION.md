# API Documentation - Post2Post Library

## 1. Core API Reference

### Server Configuration API

#### `NewServer() *Server`
Creates a new post2post server instance with default configuration.

**Returns**: Pointer to configured `Server` instance

**Example**:
```go
server := post2post.NewServer()
```

#### `WithNetwork(network string) *Server`
Configures the network type for the server.

**Parameters**:
- `network` (string): Network type ("tcp4" or "tcp6")

**Returns**: Server instance for method chaining

**Example**:
```go
server := post2post.NewServer().WithNetwork("tcp4")
```

#### `WithInterface(iface string) *Server`
Sets the network interface to listen on.

**Parameters**:
- `iface` (string): Interface address (e.g., "0.0.0.0", "127.0.0.1")

**Returns**: Server instance for method chaining

**Example**:
```go
server := post2post.NewServer().WithInterface("0.0.0.0")
```

#### `WithPostURL(url string) *Server`
Configures the target URL for posting JSON data.

**Parameters**:
- `url` (string): Target webhook URL

**Returns**: Server instance for method chaining

**Example**:
```go
server := post2post.NewServer().WithPostURL("http://receiver:8082/webhook")
```

#### `WithTimeout(timeout time.Duration) *Server`
Sets the default timeout for round-trip operations.

**Parameters**:
- `timeout` (time.Duration): Timeout duration

**Returns**: Server instance for method chaining

**Example**:
```go
server := post2post.NewServer().WithTimeout(30 * time.Second)
```

#### `WithProcessor(processor PayloadProcessor) *Server`
Configures a custom payload processor.

**Parameters**:
- `processor` (PayloadProcessor): Implementation of payload processing interface

**Returns**: Server instance for method chaining

**Example**:
```go
processor := &MyCustomProcessor{}
server := post2post.NewServer().WithProcessor(processor)
```

## 2. Server Control API

#### `Start() error`
Starts the HTTP server and begins listening for requests.

**Returns**: Error if startup fails

**Example**:
```go
err := server.Start()
if err != nil {
    log.Fatal("Failed to start server:", err)
}
```

#### `Stop() error`
Stops the HTTP server and closes all connections.

**Returns**: Error if shutdown fails

**Example**:
```go
err := server.Stop()
if err != nil {
    log.Printf("Error stopping server: %v", err)
}
```

#### `IsRunning() bool`
Checks if the server is currently running.

**Returns**: True if server is running, false otherwise

**Example**:
```go
if server.IsRunning() {
    fmt.Println("Server is active")
}
```

## 3. Server Information API

#### `GetPort() int`
Returns the port number the server is listening on.

**Returns**: Port number (int)

**Example**:
```go
port := server.GetPort()
fmt.Printf("Server listening on port: %d\n", port)
```

#### `GetInterface() string`
Returns the network interface the server is bound to.

**Returns**: Interface address (string)

**Example**:
```go
iface := server.GetInterface()
fmt.Printf("Server interface: %s\n", iface)
```

#### `GetURL() string`
Returns the complete URL for the server.

**Returns**: Full server URL (string)

**Example**:
```go
url := server.GetURL()
fmt.Printf("Server URL: %s\n", url)
```

#### `GetNetwork() string`
Returns the network type (tcp4/tcp6) the server is using.

**Returns**: Network type (string)

**Example**:
```go
network := server.GetNetwork()
fmt.Printf("Network type: %s\n", network)
```

## 4. Communication API

#### `PostJSON(payload interface{}) error`
Posts JSON data to the configured URL using regular HTTP.

**Parameters**:
- `payload` (interface{}): Data to be JSON-encoded and sent

**Returns**: Error if posting fails

**Example**:
```go
payload := map[string]interface{}{
    "message": "Hello World",
    "timestamp": time.Now(),
}

err := server.PostJSON(payload)
if err != nil {
    log.Printf("Failed to post: %v", err)
}
```

#### `PostJSONWithTailnet(payload interface{}, tailnetKey string) error`
Posts JSON data using optional Tailscale networking.

**Parameters**:
- `payload` (interface{}): Data to be JSON-encoded and sent
- `tailnetKey` (string): Tailscale auth key for secure networking

**Returns**: Error if posting fails

**Example**:
```go
payload := map[string]interface{}{
    "message": "Secure message",
    "data": sensitiveData,
}

err := server.PostJSONWithTailnet(payload, "tskey-auth-abc123...")
if err != nil {
    log.Printf("Failed to post via Tailscale: %v", err)
}
```

#### `RoundTripPost(payload interface{}, tailnetKey string) (*RoundTripResponse, error)`
Sends a round-trip request and waits for a response.

**Parameters**:
- `payload` (interface{}): Data to send
- `tailnetKey` (string): Optional Tailscale auth key

**Returns**: 
- `*RoundTripResponse`: Response data structure
- `error`: Error if operation fails

**Example**:
```go
payload := map[string]string{"command": "process"}

response, err := server.RoundTripPost(payload, tailnetKey)
if err != nil {
    log.Printf("Round-trip failed: %v", err)
    return
}

if response.Success {
    fmt.Printf("Response: %+v\n", response.Payload)
}
```

#### `RoundTripPostWithTimeout(payload interface{}, tailnetKey string, timeout time.Duration) (*RoundTripResponse, error)`
Sends a round-trip request with custom timeout.

**Parameters**:
- `payload` (interface{}): Data to send
- `tailnetKey` (string): Optional Tailscale auth key
- `timeout` (time.Duration): Custom timeout duration

**Returns**: 
- `*RoundTripResponse`: Response data structure
- `error`: Error if operation fails

**Example**:
```go
response, err := server.RoundTripPostWithTimeout(
    payload, 
    tailnetKey, 
    60*time.Second,
)
```

## 5. OAuth Integration API

#### `GenerateTailnetKeyFromOAuth(reusable bool, ephemeral bool, preauth bool, tags string) (string, error)`
Generates Tailscale auth keys using OAuth credentials.

**Parameters**:
- `reusable` (bool): Whether the key can be used multiple times
- `ephemeral` (bool): Whether devices are automatically removed when offline
- `preauth` (bool): Whether devices are pre-authorized
- `tags` (string): Comma-separated list of device tags

**Returns**: 
- `string`: Generated auth key (tskey-auth-...)
- `error`: Error if generation fails

**Environment Variables Required**:
- `TS_API_CLIENT_ID`: OAuth client ID
- `TS_API_CLIENT_SECRET`: OAuth client secret

**Example**:
```go
authKey, err := server.GenerateTailnetKeyFromOAuth(
    true,  // reusable
    true,  // ephemeral
    false, // not preauthorized
    "tag:ephemeral-device,tag:ci",
)

if err != nil {
    log.Fatal("Failed to generate auth key:", err)
}

fmt.Printf("Generated key: %s\n", authKey)
```

## 6. Data Structures

### `PostData`
Structure for HTTP request payloads.

```go
type PostData struct {
    URL        string      `json:"url"`         // Callback URL
    Payload    interface{} `json:"payload"`     // User data
    RequestID  string      `json:"request_id"`  // Unique request identifier
    TailnetKey string      `json:"tailnet_key"` // Optional Tailscale key
}
```

### `RoundTripResponse`
Structure for round-trip response data.

```go
type RoundTripResponse struct {
    Payload   interface{} `json:"payload"`    // Response data
    Success   bool        `json:"success"`    // Success indicator
    Error     string      `json:"error"`      // Error message if failed
    Timeout   bool        `json:"timeout"`    // Timeout indicator
    RequestID string      `json:"request_id"` // Matching request ID
}
```

### `ProcessorContext`
Context data for advanced payload processors.

```go
type ProcessorContext struct {
    RequestID  string    `json:"request_id"`  // Request identifier
    URL        string    `json:"url"`         // Source URL
    TailnetKey string    `json:"tailnet_key"` // Tailscale key
    ReceivedAt time.Time `json:"received_at"` // Timestamp
}
```

## 7. Payload Processor Interfaces

### `PayloadProcessor`
Basic payload processing interface.

```go
type PayloadProcessor interface {
    Process(payload interface{}, requestID string) (interface{}, error)
}
```

**Implementation Example**:
```go
type CustomProcessor struct{}

func (p *CustomProcessor) Process(payload interface{}, requestID string) (interface{}, error) {
    // Transform payload
    result := map[string]interface{}{
        "original": payload,
        "processed_at": time.Now(),
        "request_id": requestID,
    }
    return result, nil
}
```

### `AdvancedPayloadProcessor`
Context-aware payload processing interface.

```go
type AdvancedPayloadProcessor interface {
    ProcessWithContext(payload interface{}, context ProcessorContext) (interface{}, error)
}
```

**Implementation Example**:
```go
type AdvancedProcessor struct{}

func (p *AdvancedProcessor) ProcessWithContext(payload interface{}, ctx ProcessorContext) (interface{}, error) {
    result := map[string]interface{}{
        "payload": payload,
        "context": map[string]interface{}{
            "request_id": ctx.RequestID,
            "received_at": ctx.ReceivedAt,
            "has_tailscale": ctx.TailnetKey != "",
        },
    }
    return result, nil
}
```

## 8. Built-in Processors

### `HelloWorldProcessor`
Returns a simple "Hello World" message.

```go
processor := &HelloWorldProcessor{}
server := post2post.NewServer().WithProcessor(processor)
```

### `EchoProcessor`
Returns the original payload unchanged.

```go
processor := &EchoProcessor{}
server := post2post.NewServer().WithProcessor(processor)
```

### `TimestampProcessor`
Adds timestamp information to payloads.

```go
processor := &TimestampProcessor{}
server := post2post.NewServer().WithProcessor(processor)
```

### `CounterProcessor`
Maintains a request counter and adds count to responses.

```go
processor := &CounterProcessor{}
server := post2post.NewServer().WithProcessor(processor)

// Get current count
count := processor.GetCount()

// Reset counter
processor.Reset()
```

### `TransformProcessor`
Applies custom transformation functions to payloads.

```go
transformFunc := func(payload interface{}) interface{} {
    return map[string]interface{}{
        "transformed": payload,
        "timestamp": time.Now(),
    }
}

processor := NewTransformProcessor(transformFunc)
server := post2post.NewServer().WithProcessor(processor)
```

### `ValidatorProcessor`
Validates payloads using custom validation functions.

```go
validateFunc := func(payload interface{}) error {
    payloadMap, ok := payload.(map[string]interface{})
    if !ok {
        return fmt.Errorf("payload must be an object")
    }
    
    if _, exists := payloadMap["required_field"]; !exists {
        return fmt.Errorf("missing required field")
    }
    
    return nil
}

processor := NewValidatorProcessor(validateFunc)
server := post2post.NewServer().WithProcessor(processor)
```

### `ChainProcessor`
Chains multiple processors for sequential processing.

```go
processors := []PayloadProcessor{
    &ValidatorProcessor{},
    &TransformProcessor{},
    &CounterProcessor{},
}

chainProcessor := NewChainProcessor(processors...)
server := post2post.NewServer().WithProcessor(chainProcessor)
```

## 9. HTTP Endpoints

### `GET /`
Returns basic server information.

**Response**: Plain text server status

### `POST /webhook`
Processes incoming webhook requests with configurable payload processing.

**Request Body**:
```json
{
    "url": "http://callback-url/endpoint",
    "payload": { /* user data */ },
    "request_id": "optional-request-id",
    "tailnet_key": "optional-tailscale-key"
}
```

**Response**:
```json
{
    "status": "received",
    "message": "Processing request"
}
```

### `POST /roundtrip`
Receives responses for round-trip communications.

**Request Body**:
```json
{
    "request_id": "matching-request-id",
    "payload": { /* response data */ },
    "tailnet_key": "optional-tailscale-key"
}
```

**Response**: 200 OK with confirmation message

## 10. Error Handling

### Common Error Types

1. **Configuration Errors**
   ```go
   // Missing required configuration
   err := server.Start() // Returns error if no interface configured
   ```

2. **Network Errors**
   ```go
   // Network connectivity issues
   err := server.PostJSON(payload) // Returns network error
   ```

3. **Timeout Errors**
   ```go
   response, err := server.RoundTripPost(payload, "")
   if response.Timeout {
       // Handle timeout condition
   }
   ```

4. **Processing Errors**
   ```go
   // Payload processor errors
   processor := &ValidatorProcessor{...}
   // Returns error if validation fails
   ```

### Error Response Format

```go
type ErrorResponse struct {
    Error   string `json:"error"`
    Code    int    `json:"code"`
    Details string `json:"details,omitempty"`
}
```

This API documentation provides comprehensive coverage of all public interfaces and usage patterns for the post2post library.