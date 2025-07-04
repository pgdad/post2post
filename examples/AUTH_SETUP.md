# Tailscale OAuth Auth Key Generator

This program (`auth_setup.go`) generates ephemeral Tailscale auth keys using OAuth client credentials. It's designed for automated systems that need to provision ephemeral devices without manual intervention.

## Features

- ✅ **OAuth Authentication**: Uses OAuth client credentials flow
- ✅ **Ephemeral Keys**: Creates keys for temporary/container devices
- ✅ **Configurable Tags**: Supports custom device tags
- ✅ **Automated Generation**: Perfect for CI/CD and container deployments
- ✅ **Secure**: Uses environment variables for credentials
- ✅ **Detailed Output**: Provides comprehensive key information

## Prerequisites

### 1. Create OAuth Client

1. Go to [Tailscale Admin Console](https://login.tailscale.com/admin/settings/oauth)
2. Click "Generate OAuth client"
3. Configure the OAuth client:
   - **Name**: `Ephemeral Key Generator`
   - **Scopes**: Select `devices` (required for auth key creation)
   - **Tags**: Add tags that the client can create devices with (e.g., `tag:ephemeral-device`)
4. Save the client ID and secret

### 2. Required Environment Variables

```bash
# OAuth client credentials (required)
export TAILSCALE_OAUTH_CLIENT_ID="your-client-id"
export TAILSCALE_OAUTH_CLIENT_SECRET="your-client-secret"

# Tailnet identifier (required)
export TAILSCALE_TAILNET="example.com"  # or "tail123abc.ts.net"

# Optional configuration
export TAILSCALE_TAGS="tag:ephemeral-device,tag:ci"
export TAILSCALE_KEY_DESCRIPTION="Generated for automated deployment"
```

## Usage

### Basic Usage

```bash
# Set required environment variables
export TAILSCALE_OAUTH_CLIENT_ID="your-client-id"
export TAILSCALE_OAUTH_CLIENT_SECRET="your-client-secret"
export TAILSCALE_TAILNET="example.com"

# Generate ephemeral auth key
go run auth_setup.go
```

### With Custom Configuration

```bash
# Set all environment variables
export TAILSCALE_OAUTH_CLIENT_ID="your-client-id"
export TAILSCALE_OAUTH_CLIENT_SECRET="your-client-secret"
export TAILSCALE_TAILNET="example.com"
export TAILSCALE_TAGS="tag:ci,tag:ephemeral-node"
export TAILSCALE_KEY_DESCRIPTION="CI/CD deployment key"

# Generate auth key
go run auth_setup.go
```

### Build and Run

```bash
# Build the program
go build -o auth_setup auth_setup.go

# Run with environment variables
TAILSCALE_OAUTH_CLIENT_ID="your-id" \
TAILSCALE_OAUTH_CLIENT_SECRET="your-secret" \
TAILSCALE_TAILNET="example.com" \
./auth_setup
```

## Example Output

```
Tailscale OAuth Auth Key Generator
==================================

Using tags: [tag:ephemeral-device]
Tailnet: example.com
Description: Ephemeral auth key generated at 2023-12-01 14:30:25

Step 1: Obtaining OAuth access token...
✓ Successfully obtained access token (expires in 3600 seconds)

Step 2: Creating ephemeral auth key...
✓ Successfully created ephemeral auth key

Auth Key Details:
================
Key ID: kabc123def456
Description: Ephemeral auth key generated at 2023-12-01 14:30:25
Created: 2023-12-01T14:30:25Z
Expires: 2024-02-29T14:30:25Z
Ephemeral: true
Reusable: true
Preauthorized: false
Tags: [tag:ephemeral-device]

=== AUTH KEY ===
tskey-auth-kABC123DEF456GHI789JKL-MNOP
===============

Usage Instructions:
==================
1. Use this auth key to connect ephemeral devices:
   tailscale up --auth-key='tskey-auth-kABC123DEF456GHI789JKL-MNOP'

2. For automated systems, export as environment variable:
   export TAILSCALE_AUTH_KEY='tskey-auth-kABC123DEF456GHI789JKL-MNOP'

3. In your applications:
   TAILSCALE_AUTH_KEY='tskey-auth-kABC123DEF456GHI789JKL-MNOP' ./your-app

Note: This is an ephemeral auth key. Devices using it will be
automatically removed when they go offline.
```

## Integration Examples

### CI/CD Pipeline

```yaml
# GitHub Actions example
- name: Generate Tailscale Auth Key
  run: |
    export TAILSCALE_OAUTH_CLIENT_ID="${{ secrets.TAILSCALE_CLIENT_ID }}"
    export TAILSCALE_OAUTH_CLIENT_SECRET="${{ secrets.TAILSCALE_CLIENT_SECRET }}"
    export TAILSCALE_TAILNET="${{ secrets.TAILSCALE_TAILNET }}"
    export TAILSCALE_TAGS="tag:ci,tag:github-actions"
    go run auth_setup.go > auth_key.txt
    export TAILSCALE_AUTH_KEY=$(tail -n 1 auth_key.txt)
```

### Docker Container

```dockerfile
FROM golang:1.21-alpine
COPY auth_setup.go /app/
WORKDIR /app
RUN go build -o auth_setup auth_setup.go

# Usage:
# docker run -e TAILSCALE_OAUTH_CLIENT_ID=... -e TAILSCALE_OAUTH_CLIENT_SECRET=... your-image
```

### Automated Deployment Script

```bash
#!/bin/bash
# deploy.sh

# Generate fresh auth key
export TAILSCALE_OAUTH_CLIENT_ID="your-client-id"
export TAILSCALE_OAUTH_CLIENT_SECRET="your-client-secret"
export TAILSCALE_TAILNET="example.com"
export TAILSCALE_TAGS="tag:deployment,tag:ephemeral"

echo "Generating Tailscale auth key..."
AUTH_KEY=$(go run auth_setup.go | grep "tskey-auth-" | tail -n 1)

if [ -z "$AUTH_KEY" ]; then
    echo "Failed to generate auth key"
    exit 1
fi

echo "Starting deployment with ephemeral Tailscale connection..."
TAILSCALE_AUTH_KEY="$AUTH_KEY" ./deploy-with-tailscale.sh
```

## Environment Variables Reference

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `TAILSCALE_OAUTH_CLIENT_ID` | ✅ | OAuth client ID from Tailscale admin console | `k123ABC456DEF` |
| `TAILSCALE_OAUTH_CLIENT_SECRET` | ✅ | OAuth client secret from Tailscale admin console | `tskey-client-k123...` |
| `TAILSCALE_TAILNET` | ✅ | Tailnet identifier (domain or tail ID) | `example.com` |
| `TAILSCALE_TAGS` | ❌ | Comma-separated device tags | `tag:ci,tag:ephemeral` |
| `TAILSCALE_KEY_DESCRIPTION` | ❌ | Description for the generated key | `CI deployment key` |

## Security Considerations

1. **Credential Storage**: Store OAuth credentials securely (e.g., in CI/CD secrets, not in code)
2. **Key Rotation**: OAuth clients don't expire, but consider rotating them periodically
3. **Access Control**: Use specific tags to limit what ephemeral devices can access
4. **Audit Logging**: OAuth client usage is logged in Tailscale audit logs
5. **Least Privilege**: Only grant the `devices` scope to OAuth clients

## Troubleshooting

### Common Errors

1. **"401 Unauthorized"**: Check OAuth client credentials
2. **"exactly one capability scope must be populated"**: Ensure tags are properly configured
3. **"insufficient permissions"**: Verify OAuth client has `devices` scope
4. **"invalid tailnet"**: Check the tailnet identifier format

### Debug Mode

Add debug logging by modifying the program:

```go
// Add after imports
import "log"

// Add before main()
func init() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)
}
```

### Testing OAuth Setup

```bash
# Test OAuth token generation only
curl -X POST https://api.tailscale.com/api/v2/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -u "client_id:client_secret" \
  -d "grant_type=client_credentials&scope=devices"
```

## Use Cases

- **Container Orchestration**: Generate auth keys for ephemeral containers
- **CI/CD Pipelines**: Connect build agents to private networks
- **Serverless Functions**: Enable Lambda functions to access private resources
- **Automated Testing**: Create isolated test environments
- **Edge Computing**: Deploy ephemeral edge nodes

This tool provides a secure, automated way to generate Tailscale auth keys for ephemeral devices without manual intervention.