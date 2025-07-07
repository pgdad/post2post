#!/bin/bash

# Build script for Post2Post AWS Credentials Process
# This script builds the Go credentials process binary for multiple platforms

set -e

echo "🚀 Building Post2Post AWS Credentials Process"
echo "=============================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed or not in PATH"
    echo "Please install Go 1.24+ to build this program"
    exit 1
fi

# Check Go version
echo "📋 Checking Go version..."
go_version=$(go version)
echo "Using: $go_version"

# Ensure we're in the right directory
if [ ! -f "main.go" ]; then
    echo "❌ Error: main.go not found. Please run this script from the go-credentials-process directory"
    exit 1
fi

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -f post2post-credentials*

# Get build information
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION="v1.0.0"

# Build flags
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

# Initialize/update Go modules
echo "📦 Updating Go modules..."
go mod tidy

# Build for current platform
echo "🔨 Building for current platform..."
go build -ldflags="${LDFLAGS}" -o post2post-credentials main.go

# Make executable
chmod +x post2post-credentials

# Build for common platforms
echo "🔨 Building for multiple platforms..."

# Linux AMD64
echo "  - Building for Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o post2post-credentials-linux-amd64 main.go

# Linux ARM64
echo "  - Building for Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o post2post-credentials-linux-arm64 main.go

# macOS AMD64
echo "  - Building for macOS AMD64..."
GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o post2post-credentials-darwin-amd64 main.go

# macOS ARM64 (Apple Silicon)
echo "  - Building for macOS ARM64..."
GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o post2post-credentials-darwin-arm64 main.go

# Windows AMD64
echo "  - Building for Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o post2post-credentials-windows-amd64.exe main.go

# Display build results
echo ""
echo "✅ Build completed successfully!"
echo ""
echo "📊 Build artifacts created:"
ls -la post2post-credentials* | while read line; do
    echo "   $line"
done
echo ""

# Show binary sizes
echo "📈 Binary sizes:"
for binary in post2post-credentials*; do
    if [ -f "$binary" ]; then
        size=$(du -h "$binary" | cut -f1)
        echo "   $binary: $size"
    fi
done
echo ""

# Test the main binary
echo "🧪 Testing main binary..."
if ./post2post-credentials --help > /dev/null 2>&1; then
    echo "   ✅ Binary executes successfully"
else
    echo "   ❌ Binary test failed"
    exit 1
fi

# Installation instructions
echo "🔧 Installation instructions:"
echo ""
echo "Local installation:"
echo "   cp post2post-credentials ~/.local/bin/"
echo "   # or"
echo "   sudo cp post2post-credentials /usr/local/bin/"
echo ""
echo "Create AWS config:"
echo "   [profile myprofile]"
echo "   credential_process = /usr/local/bin/post2post-credentials"
echo "   region = us-east-1"
echo ""
echo "Set environment variables:"
echo "   export POST2POST_LAMBDA_URL=\"https://your-lambda-url.amazonaws.com/\""
echo "   export POST2POST_ROLE_ARN=\"arn:aws:iam::123456789012:role/remote/MyRole\""
echo "   export POST2POST_TAILNET_KEY=\"tskey-auth-your-key\""
echo ""
echo "Test the installation:"
echo "   aws sts get-caller-identity --profile myprofile"
echo ""

echo "🎉 Go credentials process build complete!"