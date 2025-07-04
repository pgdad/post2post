#!/bin/bash

# Build script for receiver_tailnet.go with Tailscale integration
set -e

echo "Building Tailscale-enabled receiver..."
echo "======================================"

# Check if we're in the examples directory
if [ ! -f "receiver_tailnet.go" ]; then
    echo "Error: receiver_tailnet.go not found. Run this script from the examples directory."
    exit 1
fi

# Save original directory
ORIGINAL_DIR="$PWD"

# Create temporary directory for building
BUILD_DIR=$(mktemp -d)
echo "Using temporary build directory: $BUILD_DIR"

# Copy the receiver source
cp receiver_tailnet.go "$BUILD_DIR/"

# Create the go.mod with Tailscale dependency
cat > "$BUILD_DIR/go.mod" << 'EOF'
module receiver-tailnet

go 1.21

require tailscale.com v1.76.1
EOF

# Change to build directory
cd "$BUILD_DIR"

echo "Downloading Tailscale dependencies..."
go mod tidy

echo "Building receiver_tailnet..."
go build -o receiver_tailnet receiver_tailnet.go

# Copy back to original directory
cp receiver_tailnet "$ORIGINAL_DIR/receiver_tailnet_with_tsnet"

# Cleanup
rm -rf "$BUILD_DIR"

echo ""
echo "Build complete!"
echo "Executable created: receiver_tailnet_with_tsnet"
echo ""
echo "To run:"
echo "  ./receiver_tailnet_with_tsnet"
echo ""
echo "Note: This receiver has full Tailscale integration enabled."
echo "Make sure to provide valid Tailscale auth keys in requests."
echo ""
echo "Environment setup:"
echo "  export TAILSCALE_AUTH_KEY='tskey-auth-your-key-here'"
echo "  ./receiver_tailnet_with_tsnet"