# Post2Post Examples

This directory contains example programs demonstrating how to use the post2post library for round trip posting.

## Programs

### 1. `receiver.go` - Receiving Web Server

A standalone Go program that acts as the receiving web server. This server:

- Listens on port 8081 at `/webhook` endpoint
- Receives POST requests with JSON data
- Extracts the callback URL from the request
- Processes the data and adds a timestamp field
- Posts the enhanced response back to the callback URL

**Features:**
- Adds human-readable timestamp to responses
- Includes processing metadata (processed_by, status)
- Preserves original payload data
- Provides logging of all operations
- Handles concurrent requests

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

## Running the Examples

### Prerequisites

1. Make sure you're in the post2post project directory
2. Initialize Go modules if not already done:
   ```bash
   go mod tidy
   ```

### Step 1: Start the Receiver Server

In one terminal:

```bash
cd examples
go run receiver.go
```

You should see:
```
Receiving server starting on port :8081
Send POST requests to http://localhost:8081/webhook
```

### Step 2: Run the Client

In another terminal:

```bash
cd examples
go run client.go
```

The client will demonstrate various round trip scenarios and display the results.

## Example Flow

1. **Client starts** its own web server (random port)
2. **Client posts** JSON data to receiver at `http://localhost:8081/webhook`
3. **Receiver processes** the data and adds timestamp
4. **Receiver posts back** enhanced data to client's `/roundtrip` endpoint
5. **Client receives** the response and displays results

## JSON Structure

### Request from Client to Receiver:
```json
{
  "url": "http://localhost:3000/roundtrip",
  "payload": {
    "message": "Hello from client",
    "data": {"key": "value"}
  },
  "request_id": "req_1234567890"
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

## Customization

### Receiver Server

You can modify `receiver.go` to:
- Change the listening port
- Add custom processing logic
- Modify the response payload structure
- Add authentication or validation
- Implement different response delays

### Client

You can modify `client.go` to:
- Test with different payload types
- Adjust timeout values
- Add error handling scenarios
- Test with multiple receivers
- Implement custom response processing

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