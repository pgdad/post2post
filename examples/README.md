# Post2Post Examples

This directory contains example programs demonstrating how to use the post2post library for round trip posting and configurable payload processing.

## Programs

### 1. `receiver.go` - Configurable Receiving Web Server

A standalone Go program that demonstrates configurable payload processing. This server:

- Uses the post2post library with configurable processors
- Supports multiple processor types via command line arguments
- Listens on a random port and provides `/webhook` endpoint
- Processes incoming payloads according to the selected processor
- Posts processed responses back to callback URLs

**Available Processors:**
- `hello` - Hello World Processor (always returns "Hello World")
- `echo` - Echo Processor (returns original payload with metadata) *default*
- `timestamp` - Timestamp Processor (adds detailed timestamp info)
- `counter` - Counter Processor (maintains request counter)
- `advanced` - Advanced Context Processor (includes processing context)
- `transform` - Transform Processor (converts strings to uppercase)
- `validator` - Validator Processor (validates required fields: name, email)
- `chain` - Chain Processor (combines timestamp → transform → echo)

**Usage:**
```bash
go run receiver.go [processor_type]
```

**Examples:**
```bash
go run receiver.go           # Uses echo processor (default)
go run receiver.go hello     # Uses hello world processor
go run receiver.go transform # Uses transform processor
go run receiver.go validator # Uses validator processor
```

### 2. `client.go` - Post2Post Client

A demonstration client that uses the post2post library to:

- Start a local web server for receiving responses
- Send various types of payloads to the receiver
- Wait for responses with configurable timeouts
- Demonstrate different usage patterns

**Examples included:**
- Simple map payload round trip
- Struct payload round trip
- Custom timeout configuration
- Concurrent round trip requests
- Fire-and-forget JSON posting
- Tailscale integration demonstration
- Payload processor testing (Hello World, Transform, Validator)

## Running the Examples

### Prerequisites

1. Make sure you're in the post2post project directory
2. Initialize Go modules if not already done:
   ```bash
   go mod tidy
   ```

### Step 1: Start the Receiver Server

In one terminal, choose a processor type and start the receiver:

```bash
cd examples
go run receiver.go echo        # or any other processor type
```

You should see:
```
Using Echo Processor
Receiving server started at: http://127.0.0.1:xxxxx
Send POST requests to /webhook endpoint
Available endpoints:
  - http://127.0.0.1:xxxxx/webhook (for payload processing)
  - http://127.0.0.1:xxxxx/roundtrip (for round-trip responses)  
  - http://127.0.0.1:xxxxx/ (for server info)
```

To test different processors, try:
```bash
go run receiver.go hello      # Always returns "Hello World"
go run receiver.go transform  # Converts strings to uppercase
go run receiver.go validator  # Validates required fields
go run receiver.go counter    # Counts requests
```

### Step 2: Run the Client

First, update the client to connect to your receiver's actual port. In the receiver output, note the port number (e.g., `http://127.0.0.1:54321`), then edit `client.go`:

```go
// Update this line in client.go with the receiver's actual port
WithPostURL("http://127.0.0.1:RECEIVER_PORT/webhook").
```

Then in another terminal:

```bash
cd examples  
go run client.go
```

The client will demonstrate various round trip scenarios and display the results based on the processor type you selected.

## Example Flow

1. **Client starts** its own web server (random port)
2. **Client posts** JSON data to receiver at `/webhook` endpoint
3. **Receiver processes** the data using the configured processor
4. **Receiver posts back** processed data to client's `/roundtrip` endpoint
5. **Client receives** the response and displays results

The exact processing depends on which processor type you selected:
- **Hello World**: Ignores input, always returns "Hello World"
- **Echo**: Returns original payload with metadata
- **Transform**: Converts strings to uppercase
- **Validator**: Checks for required fields and reports validation status
- **Counter**: Adds incrementing counter to responses
- **Advanced**: Includes detailed context and processing information

## JSON Structure

### Request from Client to Receiver:
```json
{
  "url": "http://localhost:3000/roundtrip",
  "payload": {
    "message": "Hello from client",
    "data": {"key": "value"}
  },
  "request_id": "req_1234567890",
  "tailnet_key": "tskey-auth-xyz123..."
}
```

### Response from Receiver back to Client:
```json
{
  "request_id": "req_1234567890",
  "payload": {
    "original_data": {
      "message": "Hello from client", 
      "data": {"key": "value"}
    },
    "timestamp": "2023-12-01 14:30:25 UTC",
    "processed_by": "post2post-receiver",
    "status": "processed"
  }
}
```

## Tailscale Integration

Both examples include support for optional Tailscale networking:

### Receiver Features
- Detects `tailnet_key` in incoming requests
- Logs Tailscale integration when present
- Framework for routing responses through Tailscale networks

### Client Features  
- `PostJSONWithTailnet()` method for Tailscale-enabled requests
- Demonstrates Tailscale framework integration
- Shows how to include auth keys in requests

### Enabling Full Tailscale Support

To enable complete Tailscale functionality:

1. Install the Tailscale tsnet package: `go get tailscale.com/tsnet`
2. Configure the `createTailscaleClient()` method in post2post.go
3. Provide valid Tailscale auth keys in your applications

## Customization

### Receiver Server

You can modify `receiver.go` to:
- Change the listening port
- Add custom processing logic
- Modify the response payload structure
- Add authentication or validation
- Implement different response delays
- Enable full Tailscale tsnet integration

### Client

You can modify `client.go` to:
- Test with different payload types
- Adjust timeout values
- Add error handling scenarios
- Test with multiple receivers
- Implement custom response processing
- Test Tailscale networking with real auth keys

## Troubleshooting

### Common Issues

1. **Port already in use**: Change the port in receiver.go if 8081 is occupied
2. **Connection refused**: Make sure the receiver is running before starting the client
3. **Timeout errors**: Increase timeout values if processing takes longer
4. **JSON parsing errors**: Ensure payload structures are JSON-serializable

### Debugging

Both programs include extensive logging. Check the console output for:
- Request/response details
- Error messages
- Processing timestamps
- Network operation status

## Advanced Usage

For production use, consider:
- Adding proper error handling and retries
- Implementing authentication mechanisms
- Using configuration files for endpoints
- Adding metrics and monitoring
- Implementing graceful shutdown
- Using structured logging