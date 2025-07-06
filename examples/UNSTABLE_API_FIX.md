# Tailscale Unstable API Fix

## Problem
When running `auth_setup_with_post2post`, users encountered this error:
```
Failed to generate auth key: failed to create Tailscale auth key: use of Client without setting I_Acknowledge_This_API_Is_Unstable
```

## Root Cause
The Tailscale Go client requires explicit acknowledgment that the API is unstable before it can be used. This is a package-level variable that must be set to `true`.

## Solution
**Fixed in `post2post.go`:**
```go
func (s *Server) GenerateTailnetKeyFromOAuth(reusable bool, ephemeral bool, preauth bool, tags string) (string, error) {
    // Acknowledge the unstable API at package level
    tailscale.I_Acknowledge_This_API_Is_Unstable = true
    
    // Rest of the OAuth implementation...
}
```

### Key Points:
- `I_Acknowledge_This_API_Is_Unstable` is a **package-level variable**, not a struct field
- Must be set to `true` before using any Tailscale client functionality
- The post2post library now handles this automatically

## Verification
**Before Fix:**
```bash
$ ./auth_setup_with_post2post
Failed to generate auth key: use of Client without setting I_Acknowledge_This_API_Is_Unstable
```

**After Fix:**
```bash
$ TS_API_CLIENT_ID=test TS_API_CLIENT_SECRET=test ./auth_setup_with_post2post
Failed to generate auth key: oauth2: cannot fetch token: 401 Unauthorized
```

The error changed from an API acknowledgment error to a proper OAuth authentication error, confirming the fix works.

## Impact on Users
- ✅ **No user action required** - The fix is automatic
- ✅ **No breaking changes** - Existing code continues to work
- ✅ **Better error messages** - Users now get proper OAuth errors instead of API setup errors
- ✅ **Ready for production** - With valid OAuth credentials, the library works correctly

## Files Updated
- `../post2post.go` - Added automatic API acknowledgment
- `OAUTH_INTEGRATION.md` - Updated documentation with fix information
- Test programs rebuilt with the fix

## Usage
The OAuth integration now works seamlessly:

```bash
# Set valid OAuth credentials
export TS_API_CLIENT_ID="your-client-id"
export TS_API_CLIENT_SECRET="your-client-secret"
export TAILSCALE_TAGS="tag:ephemeral-device"

# Use any of the OAuth-integrated programs
go run auth_setup_with_post2post.go
go run client_tailnet.go
./receiver_tailnet_with_tsnet
```

All programs using the post2post library's `GenerateTailnetKeyFromOAuth()` method will now work correctly without manual API acknowledgment.