# Tailscale Integration Examples

This directory contains specialized examples demonstrating Tailscale integration with the post2post library for secure peer-to-peer networking.

## Files

### `client_tailnet.go` - Tailscale-Enabled Client
A client that demonstrates secure webhook communication using Tailscale networking:
- Reads Tailscale auth key from environment variable `TAILSCALE_AUTH_KEY`
- Generates random strings for payload testing
- Includes tailnet key in payload for receiver to use
- Supports round-trip and fire-and-forget messaging
- Handles concurrent Tailscale requests

### `receiver_tailnet.go` - Tailscale-Enabled Receiver  
A receiver that processes webhooks and responds via Tailscale networking:
- Extracts tailnet key from incoming POST requests
- Uses Tailscale tsnet for secure response posting
- Includes human-readable timestamps in responses
- Provides fallback to regular HTTP if Tailscale fails
- Echoes original payload data including random strings

## Features

### 🔐 **Secure Networking**
- End-to-end encrypted communication via Tailscale
- Zero-config VPN networking between client and receiver
- Works across NATs and firewalls

### 🎲 **Random Data Testing**
- Client generates random hex strings for each request
- Receiver echoes back all original data
- Useful for testing data integrity and processing

### 🔑 **Environment-Based Configuration**
- Tailscale auth keys from environment variables
- Configurable receiver URLs
- Secure credential management

### ⚡ **Performance Features**
- Async response processing
- Concurrent request handling
- HTTP fallback for reliability

## Prerequisites

1. **Tailscale Account**: Sign up at [tailscale.com](https://tailscale.com)
2. **Tailscale Installed**: Install on all participating systems
3. **Auth Key**: Generate from Tailscale admin console
4. **Go 1.21+**: For building the examples

## Setup Instructions

### 1. Install Tailscale

**Ubuntu/Debian:**
```bash
curl -fsSL https://tailscale.com/install.sh | sh
```

**macOS:**
```bash
brew install tailscale
```

**Windows:** Download from [tailscale.com/download](https://tailscale.com/download)

### 2. Set Up Tailscale

```bash
# Start Tailscale
sudo tailscale up

# Check status
tailscale status
```

### 3. Generate Auth Key

1. Go to [Tailscale Admin Console](https://login.tailscale.com/admin/settings/authkeys)
2. Click "Generate auth key"
3. Choose options:
   - **Reusable**: Yes (for testing)
   - **Ephemeral**: Yes (for temporary devices)
   - **Preauthorized**: Yes (to skip approval)
4. Copy the generated key (starts with `tskey-auth-`)

### 4. Configure Environment

```bash
# Set Tailscale auth key
export TAILSCALE_AUTH_KEY='tskey-auth-your-key-here'

# Set receiver URL (optional, defaults to localhost:8082)
export RECEIVER_URL='http://receiver-hostname:8082/webhook'
```

### 5. Enable Full Tailscale Integration

For production use, enable full tsnet integration:

#### In `receiver_tailnet.go`:
```go
// Uncomment the tsnet import at the top
import "tailscale.com/tsnet"

// In createTailscaleClient(), uncomment and configure:
srv := &tsnet.Server{
    Hostname: "post2post-receiver",
    AuthKey:  tailnetKey,
    Ephemeral: true,
}

err := srv.Start()
if err != nil {
    return nil, fmt.Errorf("failed to start tsnet server: %w", err)
}

client := srv.HTTPClient()
return client, nil
```

#### Update go.mod:
```bash
cd examples
go mod edit -require tailscale.com@v1.76.1
go mod tidy
```

## Usage Examples

### Basic Usage

**Terminal 1 - Start Receiver:**
```bash
cd examples
export TAILSCALE_AUTH_KEY='tskey-auth-your-key-here'
go run receiver_tailnet.go
```

**Terminal 2 - Run Client:**
```bash
cd examples  
export TAILSCALE_AUTH_KEY='tskey-auth-your-key-here'
export RECEIVER_URL='http://127.0.0.1:8082/webhook'
go run client_tailnet.go
```

### Cross-Network Usage

**On Receiver Machine:**
```bash
# Get Tailscale IP
tailscale ip

# Start receiver (accessible via Tailscale network)
export TAILSCALE_AUTH_KEY='tskey-auth-your-key-here'
go run receiver_tailnet.go
```

**On Client Machine:**
```bash
# Use receiver's Tailscale IP
export TAILSCALE_AUTH_KEY='tskey-auth-your-key-here'
export RECEIVER_URL='http://100.x.x.x:8082/webhook'  # Tailscale IP
go run client_tailnet.go
```

### Docker Usage

**Dockerfile for Receiver:**
```dockerfile
FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build receiver_tailnet.go
EXPOSE 8082
CMD ["./receiver_tailnet"]
```

**Run with Tailscale:**
```bash
docker run -e TAILSCALE_AUTH_KEY='tskey-auth-your-key' \
  -p 8082:8082 your-receiver-image
```

## Example Output

### Client Output:
```
Post2Post Tailscale Client Example
===================================
Tailscale key: tskey-auth...
Receiver URL: http://127.0.0.1:8082/webhook

Local server started at: http://127.0.0.1:54321
Ready to receive Tailscale responses at: http://127.0.0.1:54321/roundtrip

Test 1: Tailscale round-trip request
-----------------------------------
Sending payload with random data: a1b2c3d4e5f6g7h8

Test 1 Result:
  Success: true
  Timeout: false
  Request ID: req_1701234567890
  Tailscale Response:
    Tailnet Key: tske...key
    Server Timestamp: 2023-12-01 14:30:25 MST
    Processed Via: tailscale-receiver
    Random Data Echoed: a1b2c3d4e5f6g7h8
```

### Receiver Output:
```
Tailscale-enabled receiving server starting on port :8082
Send POST requests to http://localhost:8082/webhook

Received Tailscale request from http://127.0.0.1:54321/roundtrip with ID: req_1701234567890
Tailscale integration enabled with key: tskey-auth...
Original payload: map[random_data:a1b2c3d4e5f6g7h8 message:Hello from Tailscale client!]
Posting Tailscale response back to: http://127.0.0.1:54321/roundtrip
Successfully posted response via Tailscale for request ID: req_1701234567890
```

## Request/Response Format

### Client → Receiver Request:
```json
{
  "url": "http://127.0.0.1:54321/roundtrip",
  "payload": {
    "message": "Hello from Tailscale client!",
    "random_data": "a1b2c3d4e5f6g7h8",
    "client_info": {
      "timestamp": 1701234567,
      "version": "1.0",
      "secure": true
    },
    "tailnet_key": "tskey-auth-your-key-here"
  },
  "request_id": "req_1701234567890"
}
```

### Receiver → Client Response:
```json
{
  "request_id": "req_1701234567890",
  "payload": {
    "original_data": {
      "message": "Hello from Tailscale client!",
      "random_data": "a1b2c3d4e5f6g7h8",
      "client_info": {
        "timestamp": 1701234567,
        "version": "1.0", 
        "secure": true
      }
    },
    "tailnet_key": "tskey-auth-your-key-here",
    "timestamp": "2023-12-01 14:30:25 MST",
    "processed_via": "tailscale-receiver",
    "status": "processed",
    "server_info": {
      "hostname": "post2post-tailscale-receiver",
      "processed_at": "2023-12-01 14:30:25.123 MST",
      "tailscale_mode": "framework-ready",
      "network_security": "tailscale-secured"
    }
  }
}
```

## Troubleshooting

### Common Issues

1. **"TAILSCALE_AUTH_KEY environment variable is required"**
   - Set the environment variable with your auth key
   - Verify key starts with `tskey-auth-`

2. **"failed to create Tailscale client"**
   - Uncomment tsnet integration code
   - Add tailscale.com dependency to go.mod
   - Verify auth key is valid and not expired

3. **Connection timeouts**
   - Check Tailscale status: `tailscale status`
   - Verify both machines are on same tailnet
   - Test connectivity: `tailscale ping <peer-ip>`

4. **"Falling back to regular HTTP"**
   - This is normal when tsnet isn't fully configured
   - Enable full integration for production use

### Debugging

**Check Tailscale Status:**
```bash
tailscale status
tailscale netcheck
tailscale debug daemon-goroutines
```

**Test Connectivity:**
```bash
# Get peer IPs
tailscale ip

# Test ping
tailscale ping 100.x.x.x

# Test HTTP
curl http://100.x.x.x:8082/
```

**Enable Debug Logging:**
```bash
# In receiver_tailnet.go, add:
log.SetLevel(log.DebugLevel)

# Or use verbose flags:
go run receiver_tailnet.go -v
```

## Security Considerations

1. **Auth Key Management**
   - Use ephemeral keys for testing
   - Rotate keys regularly in production
   - Store keys securely (not in code)

2. **Network Security**
   - Tailscale provides automatic encryption
   - Consider additional application-layer security
   - Monitor Tailscale access logs

3. **Access Control**
   - Use Tailscale ACLs to restrict access
   - Implement application-level authorization
   - Monitor for unusual traffic patterns

## Production Considerations

1. **Key Rotation**
   - Implement automatic key rotation
   - Monitor key expiration
   - Have backup communication methods

2. **Monitoring**
   - Track Tailscale connection health
   - Monitor response times and success rates
   - Alert on fallback usage

3. **Scaling**
   - Consider tsnet performance characteristics
   - Test under load
   - Plan for network partitions

4. **Backup Connectivity**
   - Always implement HTTP fallback
   - Consider multiple networking options
   - Test failure scenarios

## Advanced Usage

### Multiple Tailnets
```go
// Support multiple tailnets
type TailnetConfig struct {
    AuthKey string
    Hostname string
}

configs := map[string]TailnetConfig{
    "prod": {AuthKey: "tskey-auth-prod-...", Hostname: "prod-receiver"},
    "dev":  {AuthKey: "tskey-auth-dev-...", Hostname: "dev-receiver"},
}
```

### Custom tsnet Configuration
```go
srv := &tsnet.Server{
    Hostname:  "post2post-receiver",
    AuthKey:   authKey,
    Ephemeral: false,  // Persistent device
    Logf:      log.Printf,
    Store:     tsnet.MemoryStore{}, // Custom state storage
}
```

### Health Monitoring
```go
// Add health check endpoint
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    status := map[string]interface{}{
        "tailscale_ready": isTailscaleReady(),
        "timestamp": time.Now(),
        "hostname": os.Hostname(),
    }
    json.NewEncoder(w).Encode(status)
})
```

This completes the Tailscale integration examples with comprehensive documentation for secure peer-to-peer webhook communication.