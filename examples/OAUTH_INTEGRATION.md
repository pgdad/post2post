# OAuth Integration in Post2Post

The post2post library now includes built-in OAuth support for automatically generating Tailscale auth keys. This eliminates the need for manual key generation and simplifies automation workflows.

## New OAuth Method

### `GenerateTailnetKeyFromOAuth()`

The post2post library now includes a method to generate Tailscale auth keys using OAuth credentials:

```go
func (s *Server) GenerateTailnetKeyFromOAuth(reusable bool, ephemeral bool, preauth bool, tags string) (string, error)
```

**Parameters:**
- `reusable`: Whether the key can be used multiple times
- `ephemeral`: Whether devices using this key are automatically removed when offline
- `preauth`: Whether devices are pre-authorized (skip manual approval)
- `tags`: Comma-separated list of tags to assign to devices

## Environment Variables

The OAuth integration uses these environment variables:

```bash
# OAuth credentials (required for OAuth key generation)
export TS_API_CLIENT_ID="your-client-id"
export TS_API_CLIENT_SECRET="your-client-secret"

# Optional configuration
export TAILSCALE_TAGS="tag:ephemeral-device,tag:ci"  # Defaults to "tag:ephemeral-device"

# Backward compatibility (still supported)
export TAILSCALE_AUTH_KEY="tskey-auth-existing-key"  # If provided, OAuth generation is skipped
```

## Updated Examples

### 1. Standalone OAuth Auth Key Generator

**File: `auth_setup_with_post2post.go`**

Uses the post2post library's OAuth integration instead of implementing OAuth logic separately:

```go
package main

import (
    "github.com/pgdad/post2post"
)

func main() {
    server := post2post.NewServer()
    
    authKey, err := server.GenerateTailnetKeyFromOAuth(
        true,  // reusable
        true,  // ephemeral
        false, // preauthorized
        "tag:ephemeral-device", // tags
    )
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Generated auth key: %s\n", authKey)
}
```

### 2. Updated Client with Automatic OAuth

**File: `client_tailnet.go`** (updated)

The client now automatically generates auth keys if none are provided:

```go
// Try to get existing auth key first
tailnetKey := os.Getenv("TAILSCALE_AUTH_KEY")

// If no auth key provided, generate one using OAuth
if tailnetKey == "" {
    if os.Getenv("TS_API_CLIENT_ID") != "" && os.Getenv("TS_API_CLIENT_SECRET") != "" {
        tempServer := post2post.NewServer()
        
        generatedKey, err := tempServer.GenerateTailnetKeyFromOAuth(
            true,  // reusable
            true,  // ephemeral
            false, // preauthorized
            os.Getenv("TAILSCALE_TAGS"), // tags
        )
        
        if err != nil {
            log.Fatal(err)
        }
        
        tailnetKey = generatedKey
    }
}
```

### 3. Updated Receiver with OAuth Fallback

**File: `receiver_tailnet.go`** (updated)

The receiver can now generate auth keys when posting responses if no key is provided:

```go
func postResponseViaTailscale(url string, data []byte, tailnetKey string) error {
    var actualTailnetKey string
    
    if tailnetKey == "" {
        // Generate key using OAuth if none provided
        tempServer := post2post.NewServer()
        
        generatedKey, err := tempServer.GenerateTailnetKeyFromOAuth(
            true, true, false, "tag:ephemeral-device",
        )
        
        if err != nil {
            return err
        }
        
        actualTailnetKey = generatedKey
    } else {
        actualTailnetKey = tailnetKey
    }
    
    // Use the key for Tailscale client creation...
}
```

## Migration Guide

### From Standalone `auth_setup.go`

**Before:**
```bash
# Required environment variables
export TAILSCALE_OAUTH_CLIENT_ID="client-id"
export TAILSCALE_OAUTH_CLIENT_SECRET="client-secret"
export TAILSCALE_TAILNET="example.com"

go run auth_setup.go
```

**After:**
```bash
# Updated environment variables (no TAILSCALE_TAILNET needed)
export TS_API_CLIENT_ID="client-id"
export TS_API_CLIENT_SECRET="client-secret"
export TAILSCALE_TAGS="tag:ephemeral-device"

go run auth_setup_with_post2post.go
```

### From Manual Key Management

**Before:**
```bash
# Manual process
1. Generate key via auth_setup.go
2. Export TAILSCALE_AUTH_KEY="generated-key"
3. Run client_tailnet.go
```

**After:**
```bash
# Automatic process
export TS_API_CLIENT_ID="client-id"
export TS_API_CLIENT_SECRET="client-secret"
go run client_tailnet.go  # Automatically generates key if needed
```

## Usage Patterns

### 1. Fully Automatic (Recommended)

Set OAuth credentials once, everything else is automatic:

```bash
export TS_API_CLIENT_ID="your-client-id"
export TS_API_CLIENT_SECRET="your-client-secret"
export TAILSCALE_TAGS="tag:ephemeral-device,tag:ci"

# All these now work automatically:
go run client_tailnet.go
go run auth_setup_with_post2post.go
./receiver_tailnet_with_tsnet
```

### 2. Hybrid (OAuth + Manual Keys)

Use OAuth for automation, manual keys for specific cases:

```bash
# For automation
export TS_API_CLIENT_ID="client-id"
export TS_API_CLIENT_SECRET="client-secret"

# For specific use case
export TAILSCALE_AUTH_KEY="tskey-auth-specific-key"
go run client_tailnet.go  # Uses manual key, skips OAuth
```

### 3. CI/CD Integration

```yaml
# GitHub Actions example
- name: Tailscale Integration Test
  env:
    TS_API_CLIENT_ID: ${{ secrets.TS_API_CLIENT_ID }}
    TS_API_CLIENT_SECRET: ${{ secrets.TS_API_CLIENT_SECRET }}
    TAILSCALE_TAGS: "tag:ci,tag:github-actions"
  run: |
    go run client_tailnet.go
    # Keys are generated automatically as needed
```

## Error Handling and Fallbacks

### Priority Order

1. **Manual Key**: If `TAILSCALE_AUTH_KEY` is set, use it directly
2. **OAuth Generation**: If OAuth credentials are available, generate ephemeral key
3. **Error**: If neither is available, fail with helpful error message

### Example Error Messages

```
Neither TAILSCALE_AUTH_KEY nor OAuth credentials (TS_API_CLIENT_ID, TS_API_CLIENT_SECRET) are available
```

```
Failed to generate OAuth auth key: OAuth client missing 'devices' scope
```

### Graceful Degradation

The receiver can fall back to regular HTTP if Tailscale key generation fails:

```go
err := postResponseViaTailscale(url, data, tailnetKey)
if err != nil {
    log.Printf("Tailscale failed: %v, falling back to HTTP", err)
    err = postResponseViaHTTP(url, data)
}
```

## Benefits of Integration

### ✅ **Simplified Setup**
- No need to run separate auth key generation
- Automatic key generation on demand
- Fewer manual steps in deployment

### ✅ **Better Security**
- Ephemeral keys generated fresh for each use
- No long-lived keys in environment variables
- OAuth credentials can be rotated independently

### ✅ **Improved Automation**
- Perfect for CI/CD pipelines
- Container-friendly (no pre-generated keys needed)
- Scales to multiple environments

### ✅ **Backward Compatibility**
- Existing `TAILSCALE_AUTH_KEY` usage still works
- No breaking changes to existing code
- Gradual migration path

## Troubleshooting

### Common Issues

1. **"TS_API_CLIENT_ID and TS_API_CLIENT_SECRET must be set"**
   - Solution: Set OAuth environment variables

2. **"at least one tag must be specified"**
   - Solution: Set `TAILSCALE_TAGS` or ensure default is used

3. **"OAuth client missing 'devices' scope"**
   - Solution: Recreate OAuth client with `devices` scope checked

4. **"requested tags are invalid or not permitted"**
   - Solution: Update ACL configuration (see `TAILSCALE_SETUP_GUIDE.md`)

### Debug Mode

Enable detailed logging:

```go
import "log"

func init() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)
}
```

### Testing OAuth Integration

```bash
# Test OAuth credentials
export TS_API_CLIENT_ID="your-id"
export TS_API_CLIENT_SECRET="your-secret"
export TAILSCALE_TAGS="tag:ephemeral-device"

go run auth_setup_with_post2post.go
```

## API Reference

### `GenerateTailnetKeyFromOAuth()`

```go
func (s *Server) GenerateTailnetKeyFromOAuth(
    reusable bool,     // Can the key be used multiple times?
    ephemeral bool,    // Are devices automatically removed when offline?
    preauth bool,      // Are devices pre-authorized?
    tags string,       // Comma-separated tags
) (string, error)
```

**Environment Variables Used:**
- `TS_API_CLIENT_ID` (required)
- `TS_API_CLIENT_SECRET` (required)

**Returns:**
- `string`: The generated auth key (starts with `tskey-auth-`)
- `error`: Any error that occurred during generation

**Example Usage:**
```go
server := post2post.NewServer()

// Generate ephemeral, reusable key
key, err := server.GenerateTailnetKeyFromOAuth(
    true,  // reusable
    true,  // ephemeral
    false, // not preauthorized
    "tag:ephemeral-device,tag:ci",
)

if err != nil {
    return err
}

// Use the key for Tailscale operations
err = server.PostJSONWithTailnet(payload, key)
```

This integration makes Tailscale auth key management seamless and automatic while maintaining full backward compatibility with existing workflows.