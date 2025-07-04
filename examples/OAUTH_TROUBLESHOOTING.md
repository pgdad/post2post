# OAuth Client Troubleshooting Guide

## Error: "OAuth client cannot grant scopes 'devices'"

This error means your OAuth client doesn't have the required permissions to create auth keys.

## Root Cause

The `devices` scope is required for creating auth keys, but your OAuth client was created without this permission.

## Solution Steps

### 1. Check Current OAuth Client
1. Go to [Tailscale Admin Console > Settings > OAuth](https://login.tailscale.com/admin/settings/oauth)
2. Look at your existing OAuth client
3. Check if `devices` scope is listed

### 2. Create New OAuth Client (Recommended)

**Step A: Start Creation**
1. Click "Generate OAuth client"
2. Give it a descriptive name: `Auth Key Generator`

**Step B: Select Scopes**
The most important step - select the `devices` scope:
```
Scopes: (Select the scopes this client needs)
☐ all          - Full access to the tailnet
☐ acl          - Read and write ACL configuration  
☐ dns          - Read and write DNS configuration
☑ devices      - Read, write, and delete devices  ← **MUST BE CHECKED**
☐ routes       - Read and write subnet routes
☐ users        - Read user information
```

**Step C: Configure Tags**
Add the tags you want this client to be able to assign:
- `tag:ephemeral-device`
- `tag:ci`
- `tag:automation`
- `tag:post2post`

**Step D: Generate and Save**
1. Click "Generate Client"
2. **Important**: Copy both the Client ID and Client Secret immediately
3. Store them securely (you won't be able to see the secret again)

### 3. Update Environment Variables

Replace your old credentials with the new ones:

```bash
# Old credentials (delete these)
unset TAILSCALE_OAUTH_CLIENT_ID
unset TAILSCALE_OAUTH_CLIENT_SECRET

# New credentials with devices scope
export TAILSCALE_OAUTH_CLIENT_ID="your-new-client-id"
export TAILSCALE_OAUTH_CLIENT_SECRET="your-new-client-secret"
export TAILSCALE_TAILNET="example.com"
```

### 4. Test the Fix

Run the auth setup program:
```bash
go run auth_setup.go
```

You should now see:
```
Step 1: Obtaining OAuth access token...
✓ Successfully obtained access token (expires in 3600 seconds)

Step 2: Creating ephemeral auth key...
✓ Successfully created ephemeral auth key
```

## Alternative: Edit Existing OAuth Client

**Note**: Tailscale may not allow editing existing OAuth clients' scopes. If you can't edit the existing client, create a new one.

If editing is possible:
1. Go to your OAuth client in the admin console
2. Edit the configuration
3. Ensure `devices` scope is selected
4. Save changes

## Verification Steps

### Test OAuth Token Generation
```bash
curl -X POST https://api.tailscale.com/api/v2/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -u "your-client-id:your-client-secret" \
  -d "grant_type=client_credentials&scope=devices"
```

**Expected Response:**
```json
{
  "access_token": "tskey-api-...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "devices"
}
```

### Test Auth Key Creation
```bash
# Set environment variables
export TAILSCALE_OAUTH_CLIENT_ID="your-new-client-id"
export TAILSCALE_OAUTH_CLIENT_SECRET="your-new-client-secret"
export TAILSCALE_TAILNET="example.com"

# Run the program
go run auth_setup.go
```

## Common Scope Combinations

### For Auth Key Generation Only
```
Scopes: devices
```

### For Full Device Management
```
Scopes: devices, routes
```

### For Administrative Access
```
Scopes: devices, acl, dns, routes
```

### For Complete Access (Not Recommended)
```
Scopes: all
```

## Security Best Practices

1. **Principle of Least Privilege**: Only grant the `devices` scope if you only need auth key creation
2. **Scope Limitation**: Don't use `all` scope unless absolutely necessary
3. **Tag Restrictions**: Only allow tags that the client actually needs
4. **Credential Storage**: Store OAuth credentials securely (environment variables, secrets management)
5. **Regular Rotation**: Consider rotating OAuth clients periodically

## Troubleshooting Other OAuth Issues

### "Invalid client credentials"
- Check that Client ID and Secret are correct
- Ensure no extra spaces or characters
- Verify the client hasn't been deleted

### "Token request failed"
- Check internet connectivity
- Verify Tailscale API is accessible
- Ensure proper URL encoding in requests

### "Insufficient scope"
- Verify the `devices` scope is included
- Check that the token response includes `"scope": "devices"`

### "Access denied"
- Ensure your Tailscale account has admin privileges
- Check that the tailnet identifier is correct
- Verify ACL allows OAuth client operations

## Example Working Configuration

**ACL Configuration:**
```json
{
  "tagOwners": {
    "tag:ephemeral-device": [],
    "tag:ci": []
  },
  "groups": {
    "group:admin": ["your-email@example.com"]
  },
  "acls": [
    {
      "action": "accept",
      "src": ["group:admin"],
      "dst": ["*:*"]
    },
    {
      "action": "accept",
      "src": ["tag:ephemeral-device"],
      "dst": ["*:22", "*:80", "*:443", "*:8080", "*:8082"]
    }
  ]
}
```

**OAuth Client Configuration:**
- **Name**: Auth Key Generator
- **Scopes**: `devices` ✓
- **Tags**: `tag:ephemeral-device`, `tag:ci`

**Environment Variables:**
```bash
export TAILSCALE_OAUTH_CLIENT_ID="k123ABC456DEF"
export TAILSCALE_OAUTH_CLIENT_SECRET="tskey-client-k123ABC456DEF789GHI"
export TAILSCALE_TAILNET="example.com"
export TAILSCALE_TAGS="tag:ephemeral-device"
```

This configuration will allow successful auth key generation with ephemeral device capabilities.