# Tailscale CLI and OAuth Client Setup

## Short Answer: No Direct CLI Command

**Tailscale does NOT provide a CLI command to create OAuth clients.** OAuth clients must be created through the **Tailscale Admin Console web interface**.

However, once created, OAuth clients can be used extensively with CLI tools and automation.

## Why No CLI for OAuth Client Creation?

OAuth client creation is an **administrative action** that requires:
- Owner, Admin, Network admin, or IT admin permissions
- Careful scope and permission configuration
- Secure credential handling
- Audit trail for security

Tailscale intentionally requires this to be done through the web interface for security and accountability.

## What the CLI CAN Do with OAuth

### 1. Use OAuth Client Secrets as Auth Keys

```bash
# Use OAuth client secret directly as auth key
tailscale up --auth-key="tskey-client-your-oauth-secret"

# With additional parameters
tailscale up --auth-key="tskey-client-your-oauth-secret" \
  --advertise-tags="tag:ephemeral-device"
```

### 2. Environment Variable Configuration

```bash
# Set OAuth credentials for API usage
export TS_API_CLIENT_ID="your-client-id"
export TS_API_CLIENT_SECRET="your-client-secret"

# Use in containers/Docker
export TS_AUTHKEY="tskey-client-your-oauth-secret"
export TS_EXTRA_ARGS="--advertise-tags=tag:container"
```

### 3. API Integration with CLI Tools

```bash
# Get OAuth token via curl
TOKEN=$(curl -s -X POST https://api.tailscale.com/api/v2/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -u "$TS_API_CLIENT_ID:$TS_API_CLIENT_SECRET" \
  -d "grant_type=client_credentials&scope=devices" | \
  jq -r '.access_token')

# Use token for API calls
curl -H "Authorization: Bearer $TOKEN" \
  https://api.tailscale.com/api/v2/tailnet/example.com/devices
```

## Alternative: Automated OAuth Client Setup Script

Since there's no CLI command, here's a script that guides you through the process:

```bash
#!/bin/bash
# oauth_client_setup.sh - Guide for OAuth client creation

echo "Tailscale OAuth Client Setup Guide"
echo "=================================="
echo
echo "Since Tailscale doesn't provide CLI commands for OAuth client creation,"
echo "you need to use the web interface. This script will guide you through it."
echo

read -p "Press Enter to open Tailscale Admin Console OAuth page..."

# Try to open the browser (works on most systems)
if command -v open >/dev/null 2>&1; then
    open "https://login.tailscale.com/admin/settings/oauth"
elif command -v xdg-open >/dev/null 2>&1; then
    xdg-open "https://login.tailscale.com/admin/settings/oauth"
else
    echo "Please open: https://login.tailscale.com/admin/settings/oauth"
fi

echo
echo "Follow these steps in the web interface:"
echo "1. Click 'Generate OAuth client'"
echo "2. Set Name: 'Auth Key Generator'"
echo "3. ✅ CHECK 'devices' scope"
echo "4. Add tags: tag:ephemeral-device"
echo "5. Click 'Generate client'"
echo "6. Copy both Client ID and Secret"
echo

read -p "Have you created the OAuth client and copied the credentials? (y/n): " created

if [ "$created" = "y" ]; then
    echo
    echo "Great! Now let's set up your environment variables:"
    echo
    read -p "Enter OAuth Client ID: " client_id
    read -s -p "Enter OAuth Client Secret: " client_secret
    echo
    read -p "Enter Tailnet (e.g., example.com): " tailnet
    
    echo
    echo "Add these to your ~/.bashrc or ~/.zshrc:"
    echo
    echo "export TAILSCALE_OAUTH_CLIENT_ID=\"$client_id\""
    echo "export TAILSCALE_OAUTH_CLIENT_SECRET=\"$client_secret\""
    echo "export TAILSCALE_TAILNET=\"$tailnet\""
    echo
    echo "Then run: source ~/.bashrc"
    echo "And test with: go run auth_setup.go"
else
    echo "Please complete the OAuth client creation first, then run this script again."
fi
```

## Comparison: What Other Tools Provide

### What Tailscale CLI Has:
- ✅ Device management (`tailscale status`, `tailscale logout`)
- ✅ Network operations (`tailscale ping`, `tailscale netcheck`)
- ✅ File sharing (`tailscale file`)
- ✅ Basic auth (`tailscale login`, `tailscale up --auth-key`)

### What Tailscale CLI Does NOT Have:
- ❌ OAuth client creation
- ❌ ACL management
- ❌ Admin user management
- ❌ Billing/subscription management

## Third-Party Tools

Some community tools attempt to fill this gap:

### 1. Terraform Provider
```hcl
# Using Tailscale Terraform provider
resource "tailscale_tailnet_key" "example_key" {
  reusable      = true
  ephemeral     = true
  preauthorized = false
  tags          = ["tag:example"]
}
```

### 2. Go SDK
```go
// Using unofficial Go SDK
client := tailscale.NewClient("api-key")
key, err := client.CreateAuthKey(tailscale.CreateAuthKeyRequest{
    Capabilities: tailscale.KeyCapabilities{
        Devices: tailscale.KeyCapabilityDevices{
            Create: tailscale.KeyCapabilityDevicesCreate{
                Reusable:      true,
                Ephemeral:     true,
                Preauthorized: false,
                Tags:          []string{"tag:ephemeral"},
            },
        },
    },
})
```

## Workarounds and Solutions

### 1. Use Our `auth_setup.go` Program
This is essentially a CLI tool for OAuth-based auth key creation:

```bash
# One-time OAuth client setup (web interface)
# Then use our program for automated key generation
export TAILSCALE_OAUTH_CLIENT_ID="your-id"
export TAILSCALE_OAUTH_CLIENT_SECRET="your-secret"
export TAILSCALE_TAILNET="example.com"

go run auth_setup.go  # Creates ephemeral auth keys
```

### 2. Shell Functions for Convenience
```bash
# Add to ~/.bashrc
tailscale-create-key() {
    if [ -z "$TAILSCALE_OAUTH_CLIENT_ID" ]; then
        echo "Error: Set TAILSCALE_OAUTH_CLIENT_ID first"
        return 1
    fi
    
    cd /path/to/post2post/examples
    go run auth_setup.go
}

tailscale-setup-oauth() {
    echo "Opening OAuth client creation page..."
    if command -v open >/dev/null; then
        open "https://login.tailscale.com/admin/settings/oauth"
    else
        echo "Go to: https://login.tailscale.com/admin/settings/oauth"
    fi
    echo "Remember to:"
    echo "1. Check 'devices' scope"
    echo "2. Add required tags"
    echo "3. Copy credentials to environment variables"
}
```

### 3. Container/CI Integration
```yaml
# GitHub Actions example
- name: Setup Tailscale OAuth
  env:
    TAILSCALE_OAUTH_CLIENT_ID: ${{ secrets.TAILSCALE_CLIENT_ID }}
    TAILSCALE_OAUTH_CLIENT_SECRET: ${{ secrets.TAILSCALE_CLIENT_SECRET }}
    TAILSCALE_TAILNET: ${{ secrets.TAILSCALE_TAILNET }}
  run: |
    cd examples
    go run auth_setup.go > auth_key.txt
    export TAILSCALE_AUTH_KEY=$(grep "tskey-auth" auth_key.txt)
```

## Feature Request Status

There are GitHub issues requesting CLI OAuth client creation:
- [Issue #7982](https://github.com/tailscale/tailscale/issues/7982): "let tailscale up accept oauth creds"

But as of now, OAuth client creation remains web-interface only.

## Summary

**Current State:**
- ❌ No CLI command for OAuth client creation
- ✅ OAuth clients can be used with CLI tools
- ✅ Our `auth_setup.go` provides automated auth key generation
- ✅ Various workarounds exist for automation

**Best Practice:**
1. Create OAuth client once via web interface
2. Use `auth_setup.go` or similar tools for automated key generation
3. Store OAuth credentials securely in environment variables
4. Integrate with CI/CD pipelines for automated deployments