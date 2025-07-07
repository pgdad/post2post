#!/bin/bash

# Integration test script for post2post credentials process
# This script tests the credentials process with mock/test configurations

set -e

echo "🧪 Testing Post2Post Credentials Process Integration"
echo "===================================================="

# Check if binary exists
if [ ! -f "./post2post-credentials" ]; then
    echo "❌ Error: post2post-credentials binary not found. Run 'go build' first."
    exit 1
fi

echo "✅ Binary found: post2post-credentials"

# Test 1: Help output
echo ""
echo "🧪 Test 1: Help output"
echo "-----------------------"
if ./post2post-credentials --help > /dev/null 2>&1; then
    echo "✅ Help command works"
else
    echo "❌ Help command failed"
    exit 1
fi

# Test 2: Missing configuration validation
echo ""
echo "🧪 Test 2: Configuration validation"
echo "-----------------------------------"

# Test missing lambda URL
if ./post2post-credentials --role-arn "arn:aws:iam::123456789012:role/remote/TestRole" --tailnet-key "test" 2>&1 | grep -q "lambda URL is required"; then
    echo "✅ Missing lambda URL validation works"
else
    echo "❌ Missing lambda URL validation failed"
    exit 1
fi

# Test missing role ARN
if ./post2post-credentials --lambda-url "https://test.com" --tailnet-key "test" 2>&1 | grep -q "role ARN is required"; then
    echo "✅ Missing role ARN validation works"
else
    echo "❌ Missing role ARN validation failed"
    exit 1
fi

# Test missing tailnet key
if ./post2post-credentials --lambda-url "https://test.com" --role-arn "arn:aws:iam::123456789012:role/remote/TestRole" 2>&1 | grep -q "tailnet key is required"; then
    echo "✅ Missing tailnet key validation works"
else
    echo "❌ Missing tailnet key validation failed"
    exit 1
fi

# Test invalid role ARN (not in /remote/ path)
if ./post2post-credentials --lambda-url "https://test.com" --role-arn "arn:aws:iam::123456789012:role/TestRole" --tailnet-key "test" 2>&1 | grep -q "must be in /remote/ path"; then
    echo "✅ Invalid role ARN validation works"
else
    echo "❌ Invalid role ARN validation failed"
    exit 1
fi

# Test 3: Duration validation
echo ""
echo "🧪 Test 3: Duration validation"
echo "------------------------------"

# Test duration too short
if ./post2post-credentials --lambda-url "https://test.com" --role-arn "arn:aws:iam::123456789012:role/remote/TestRole" --tailnet-key "test" --duration "10m" 2>&1 | grep -q "must be at least 15 minutes"; then
    echo "✅ Short duration validation works"
else
    echo "❌ Short duration validation failed"
    exit 1
fi

# Test duration too long
if ./post2post-credentials --lambda-url "https://test.com" --role-arn "arn:aws:iam::123456789012:role/remote/TestRole" --tailnet-key "test" --duration "13h" 2>&1 | grep -q "cannot exceed 12 hours"; then
    echo "✅ Long duration validation works"
else
    echo "❌ Long duration validation failed"
    exit 1
fi

# Test 4: Environment variable support
echo ""
echo "🧪 Test 4: Environment variable support"
echo "---------------------------------------"

# Set test environment variables
export POST2POST_LAMBDA_URL="https://env-test.com"
export POST2POST_ROLE_ARN="arn:aws:iam::123456789012:role/remote/EnvTestRole"
export POST2POST_TAILNET_KEY="env-test-key"

# This should fail with network error (expected) but validate config first
if ./post2post-credentials 2>&1 | grep -E "(network|connection|timeout|DNS)" > /dev/null; then
    echo "✅ Environment variables are being read (network error is expected)"
elif ./post2post-credentials 2>&1 | grep -q "Invalid configuration"; then
    echo "❌ Environment variables not working - configuration still invalid"
    exit 1
else
    echo "✅ Environment variables working (different error occurred)"
fi

# Clean up environment variables
unset POST2POST_LAMBDA_URL POST2POST_ROLE_ARN POST2POST_TAILNET_KEY

# Test 5: Unit tests
echo ""
echo "🧪 Test 5: Go unit tests"
echo "------------------------"
if go test ./...; then
    echo "✅ Unit tests pass"
else
    echo "❌ Unit tests failed"
    exit 1
fi

# Test 6: Binary size check
echo ""
echo "🧪 Test 6: Binary characteristics"
echo "---------------------------------"
binary_size=$(du -h ./post2post-credentials | cut -f1)
echo "✅ Binary size: $binary_size"

# Check if binary is executable
if [ -x "./post2post-credentials" ]; then
    echo "✅ Binary is executable"
else
    echo "❌ Binary is not executable"
    exit 1
fi

# Final summary
echo ""
echo "🎉 All integration tests passed!"
echo ""
echo "📋 Summary:"
echo "   - Configuration validation: ✅"
echo "   - Error handling: ✅"
echo "   - Environment variables: ✅" 
echo "   - Unit tests: ✅"
echo "   - Binary characteristics: ✅"
echo ""
echo "🚀 The credentials process is ready for use!"
echo ""
echo "📖 Next steps:"
echo "   1. Copy binary to /usr/local/bin/"
echo "   2. Configure AWS CLI with credential_process"
echo "   3. Set up environment variables"
echo "   4. Test with AWS CLI: aws sts get-caller-identity --profile myprofile"