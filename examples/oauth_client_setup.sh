#!/bin/bash
# oauth_client_setup.sh - Interactive guide for OAuth client creation

set -e

echo "🔑 Tailscale OAuth Client Setup Guide"
echo "===================================="
echo
echo "Since Tailscale doesn't provide CLI commands for OAuth client creation,"
echo "you need to use the web interface. This script will guide you through it."
echo

# Check if we're in the right directory
if [ ! -f "auth_setup.go" ]; then
    echo "❌ Error: auth_setup.go not found."
    echo "Please run this script from the examples directory."
    exit 1
fi

echo "📋 What this script will help you do:"
echo "1. Open the Tailscale OAuth creation page"
echo "2. Guide you through the configuration"
echo "3. Set up environment variables"
echo "4. Test the OAuth client with auth_setup.go"
echo

read -p "Press Enter to start the OAuth client setup process..."
echo

# Step 1: Open the browser
echo "🌐 Step 1: Opening Tailscale Admin Console"
echo "----------------------------------------"

OAUTH_URL="https://login.tailscale.com/admin/settings/oauth"

if command -v open >/dev/null 2>&1; then
    echo "Opening browser..."
    open "$OAUTH_URL"
elif command -v xdg-open >/dev/null 2>&1; then
    echo "Opening browser..."
    xdg-open "$OAUTH_URL"
elif command -v wslview >/dev/null 2>&1; then
    echo "Opening browser (WSL)..."
    wslview "$OAUTH_URL"
else
    echo "Please manually open this URL in your browser:"
    echo "$OAUTH_URL"
fi

echo
echo "📝 Step 2: Configure OAuth Client"
echo "--------------------------------"
echo "In the web interface, follow these steps:"
echo
echo "1. Click 'Generate OAuth client'"
echo "2. Name: 'Auth Key Generator'"
echo "3. Scopes: ✅ CHECK 'devices' (this is critical!)"
echo "4. Tags: Add these tags (or customize as needed):"
echo "   - tag:ephemeral-device"
echo "   - tag:ci"
echo "   - tag:automation"
echo "5. Click 'Generate client'"
echo "6. 🚨 IMPORTANT: Copy both Client ID and Secret immediately!"
echo "   (You won't be able to see the secret again)"
echo

read -p "Have you created the OAuth client and copied the credentials? (y/n): " created

if [ "$created" != "y" ]; then
    echo "❌ Please complete the OAuth client creation first, then run this script again."
    echo "   Remember: You MUST check the 'devices' scope!"
    exit 1
fi

# Step 3: Collect credentials
echo
echo "🔐 Step 3: Configure Environment Variables"
echo "-----------------------------------------"

read -p "Enter OAuth Client ID: " client_id
echo
echo "Enter OAuth Client Secret (input hidden):"
read -s client_secret
echo
read -p "Enter Tailnet (e.g., example.com or tail123abc.ts.net): " tailnet

# Validate inputs
if [ -z "$client_id" ] || [ -z "$client_secret" ] || [ -z "$tailnet" ]; then
    echo "❌ Error: All fields are required"
    exit 1
fi

# Validate client secret format
if [[ ! "$client_secret" == tskey-client-* ]]; then
    echo "⚠️  Warning: Client secret should start with 'tskey-client-'"
    read -p "Continue anyway? (y/n): " continue_anyway
    if [ "$continue_anyway" != "y" ]; then
        exit 1
    fi
fi

echo
echo "✅ Credentials collected successfully!"

# Step 4: Set environment variables
echo
echo "🌍 Step 4: Setting Environment Variables"
echo "---------------------------------------"

export TAILSCALE_OAUTH_CLIENT_ID="$client_id"
export TAILSCALE_OAUTH_CLIENT_SECRET="$client_secret"
export TAILSCALE_TAILNET="$tailnet"

echo "Environment variables set for this session."

# Step 5: Test the setup
echo
echo "🧪 Step 5: Testing OAuth Client"
echo "------------------------------"

echo "Testing OAuth client with auth_setup.go..."
echo

if go run auth_setup.go; then
    echo
    echo "🎉 SUCCESS! OAuth client is working correctly!"
    echo
    echo "💾 To make these environment variables permanent, add these lines to your shell config:"
    echo
    echo "# Add to ~/.bashrc, ~/.zshrc, or ~/.profile"
    echo "export TAILSCALE_OAUTH_CLIENT_ID=\"$client_id\""
    echo "export TAILSCALE_OAUTH_CLIENT_SECRET=\"$client_secret\""
    echo "export TAILSCALE_TAILNET=\"$tailnet\""
    echo
    echo "Then run: source ~/.bashrc (or restart your terminal)"
    echo
    echo "🚀 Usage:"
    echo "  go run auth_setup.go                    # Generate new auth key"
    echo "  go run client_tailnet.go                # Test with Tailscale client"
    echo "  ./build_receiver_tailnet.sh             # Build receiver with Tailscale"
    echo
else
    echo
    echo "❌ OAuth client test failed."
    echo
    echo "🔍 Common issues:"
    echo "1. Make sure you checked the 'devices' scope when creating the OAuth client"
    echo "2. Verify the Client ID and Secret are correct"
    echo "3. Check that the tailnet identifier is correct"
    echo "4. Ensure tags are defined in your ACL (see TAILSCALE_SETUP_GUIDE.md)"
    echo
    echo "📖 For detailed troubleshooting, see:"
    echo "   - OAUTH_TROUBLESHOOTING.md"
    echo "   - TAILSCALE_SETUP_GUIDE.md"
    echo
    exit 1
fi

echo
echo "🔗 Useful Links:"
echo "  Tailscale Admin: https://login.tailscale.com/admin"
echo "  OAuth Clients: https://login.tailscale.com/admin/settings/oauth"
echo "  ACL Config: https://login.tailscale.com/admin/acls"
echo
echo "📚 Documentation files in this directory:"
echo "  - OAUTH_TROUBLESHOOTING.md  # OAuth setup issues"
echo "  - TAILSCALE_SETUP_GUIDE.md  # ACL configuration"
echo "  - AUTH_SETUP.md             # auth_setup.go usage guide"
echo "  - CLI_OAUTH_SETUP.md        # This guide's documentation"