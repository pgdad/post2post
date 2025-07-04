# Tailscale ACL and OAuth Setup Guide

This guide walks you through setting up ACLs and OAuth clients for the `auth_setup.go` program.

## Problem: "requested tags are invalid or not permitted"

This error occurs when:
1. Tags are not defined in your ACL configuration
2. OAuth client doesn't have permission to use the tags
3. Tags are not properly formatted

## Step-by-Step Setup

### 1. Define Tags in ACL Configuration

Go to [Tailscale Admin Console > Access Controls](https://login.tailscale.com/admin/acls) and update your ACL:

```json
{
  "tagOwners": {
    "tag:ephemeral-device": [],
    "tag:ci": [],
    "tag:automation": [],
    "tag:post2post": []
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
    },
    {
      "action": "accept",
      "src": ["tag:ci"],
      "dst": ["*:22", "*:80", "*:443", "*:8080", "*:8082"]
    },
    {
      "action": "accept",
      "src": ["tag:automation"],
      "dst": ["*:*"]
    },
    {
      "action": "accept",
      "src": ["tag:post2post"],
      "dst": ["*:8080", "*:8082", "*:443", "*:80"]
    },
    {
      "action": "accept",
      "src": ["*"],
      "dst": ["tag:ephemeral-device:*", "tag:ci:*", "tag:automation:*", "tag:post2post:*"]
    }
  ]
}
```

### 2. Key Points for ACL Configuration

#### `tagOwners` Section
- **Purpose**: Defines which tags exist and who can assign them
- **Empty array `[]`**: Allows OAuth clients to use these tags
- **With users**: Only specified users can assign these tags

```json
"tagOwners": {
  "tag:ephemeral-device": [],        // OAuth clients can use this
  "tag:ci": [],                      // OAuth clients can use this
  "tag:admin-only": ["admin@example.com"]  // Only admin can assign
}
```

#### `acls` Section
- **Purpose**: Defines network access rules
- **`src`**: Source (who can connect)
- **`dst`**: Destination (what can be connected to)

### 3. Create OAuth Client

1. Go to [Tailscale Admin Console > Settings > OAuth](https://login.tailscale.com/admin/settings/oauth)
2. Click "Generate OAuth client"
3. Configure:
   - **Name**: `Auth Key Generator`
   - **Scopes**: Select `devices`
   - **Tags**: Add the tags you defined (e.g., `tag:ephemeral-device`, `tag:ci`)
4. Copy the client ID and secret

### 4. Test Your Setup

Update your environment variables:

```bash
export TAILSCALE_OAUTH_CLIENT_ID="your-client-id"
export TAILSCALE_OAUTH_CLIENT_SECRET="your-client-secret"
export TAILSCALE_TAILNET="example.com"
export TAILSCALE_TAGS="tag:ephemeral-device"  # Use defined tags
```

Run the auth setup:
```bash
go run auth_setup.go
```

## Common Issues and Solutions

### Issue 1: "src=tag not found: tag:ephemeral-device"

**Problem**: Tag is referenced in ACL rules but not defined in `tagOwners`

**Solution**: Add the tag to `tagOwners`:
```json
"tagOwners": {
  "tag:ephemeral-device": []
}
```

### Issue 2: "requested tags are invalid or not permitted"

**Problem**: OAuth client doesn't have permission to use the tags

**Solution**: 
1. Ensure tags are in `tagOwners` with empty array `[]`
2. Verify OAuth client has the tags listed in its configuration
3. Check that tags use the `tag:` prefix

### Issue 3: "insufficient permissions for scope"

**Problem**: OAuth client doesn't have the `devices` scope

**Solution**: Edit OAuth client and ensure `devices` scope is selected

### Issue 4: Tags not working in ACL rules

**Problem**: Tags not properly formatted or not accessible

**Solution**: 
- Use exact format: `tag:name` (not `name` or `tags:name`)
- Ensure tags are defined in `tagOwners`
- Check that source can access destination

## Minimal Working ACL Example

For a basic setup with just ephemeral devices:

```json
{
  "tagOwners": {
    "tag:ephemeral-device": []
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
    },
    {
      "action": "accept",
      "src": ["*"],
      "dst": ["tag:ephemeral-device:*"]
    }
  ]
}
```

## Complete Setup Checklist

- [ ] Define tags in ACL `tagOwners` section
- [ ] Set `tagOwners` to empty array `[]` for OAuth client usage
- [ ] Add ACL rules for tag-based access
- [ ] Save and test ACL configuration
- [ ] Create OAuth client with `devices` scope
- [ ] Add tags to OAuth client configuration
- [ ] Copy client ID and secret
- [ ] Set environment variables
- [ ] Test with `auth_setup.go`

## Verification Steps

1. **Test ACL**: Save the ACL configuration and check for errors
2. **Test OAuth**: Try generating a token manually:
   ```bash
   curl -X POST https://api.tailscale.com/api/v2/oauth/token \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -u "client_id:client_secret" \
     -d "grant_type=client_credentials&scope=devices"
   ```
3. **Test Auth Key**: Run `auth_setup.go` with proper environment variables

## Example Environment Variables

```bash
# Required
export TAILSCALE_OAUTH_CLIENT_ID="k123ABC456DEF"
export TAILSCALE_OAUTH_CLIENT_SECRET="tskey-client-k123ABC456DEF-GHI789JKL"
export TAILSCALE_TAILNET="example.com"

# Optional - use tags defined in your ACL
export TAILSCALE_TAGS="tag:ephemeral-device,tag:ci"
export TAILSCALE_KEY_DESCRIPTION="Automated ephemeral device key"
```

## Advanced ACL Features

### Environment-Specific Tags
```json
"tagOwners": {
  "tag:prod": [],
  "tag:staging": [],
  "tag:dev": []
}
```

### Service-Specific Tags
```json
"tagOwners": {
  "tag:web-server": [],
  "tag:database": [],
  "tag:api-gateway": []
}
```

### Restricted Tags
```json
"tagOwners": {
  "tag:admin": ["admin@example.com"],
  "tag:sensitive": ["security@example.com"]
}
```

This configuration ensures that your OAuth client can create ephemeral auth keys with the proper tags and network access permissions.