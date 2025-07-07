# Post2Post AWS Credentials Process - Go

A Go implementation of AWS `credential_process` that uses Post2Post for secure credential retrieval through Tailscale mesh networks. This program integrates with AWS CLI and SDKs to provide seamless credential management.

## Overview

This program implements the AWS CLI `credential_process` interface as documented at [AWS CLI External Credential Process](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sourcing-external.html). It securely retrieves AWS credentials from a remote Lambda function through Tailscale encrypted channels.

```
AWS CLI/SDK → credentials_process → Post2Post Client → Tailscale Network → Lambda Function → STS AssumeRole → Return Credentials
```

## Features

- **AWS CLI Compatible**: Implements the standard `credential_process` interface
- **Secure Communication**: All communication through Tailscale mesh networking
- **Automatic Refresh**: AWS CLI handles credential caching and refresh
- **Flexible Configuration**: Command-line flags and environment variables
- **Comprehensive Validation**: Role ARN and configuration validation
- **Detailed Logging**: Structured logging to stderr for debugging

## Prerequisites

- **Go 1.24+**: For building the application
- **AWS Lambda**: Deployed post2post Lambda function
- **Tailscale Network**: Both client and Lambda on same tailnet
- **IAM Roles**: Target roles must be in `/remote/` path
- **AWS CLI/SDK**: For using the credentials process

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/pgdad/post2post.git
cd post2post/examples/go-credentials-process

# Build the binary
go build -o post2post-credentials main.go

# Install to system location (optional)
sudo cp post2post-credentials /usr/local/bin/
```

### Build Script

```bash
#!/bin/bash
# build.sh - Build script for post2post credentials process

set -e

echo "Building post2post credentials process..."

# Build for current platform
go build -ldflags="-s -w" -o post2post-credentials main.go

# Build for common platforms (optional)
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o post2post-credentials-linux-amd64 main.go
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o post2post-credentials-darwin-amd64 main.go
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o post2post-credentials-windows-amd64.exe main.go

echo "Build complete!"
ls -la post2post-credentials*
```

## Configuration

### Command Line Options

```bash
Usage: post2post-credentials [options]

Options:
  -lambda-url string
        AWS Lambda Function URL endpoint
  -role-arn string
        IAM Role ARN to assume (must be in /remote/ path)
  -tailnet-key string
        Tailscale auth key for secure communication
  -session-name string
        Session name for the assumed role (default "post2post-credentials-process")
  -duration duration
        Credential duration (e.g., 1h, 30m) (default 1h0m0s)
  -timeout duration
        Request timeout (e.g., 30s, 1m) (default 30s)
```

### Environment Variables

Environment variables take precedence over command-line flags:

| Variable | Description | Example |
|----------|-------------|---------|
| `POST2POST_LAMBDA_URL` | Lambda Function URL endpoint | `https://lambda-url.amazonaws.com/` |
| `POST2POST_ROLE_ARN` | IAM Role ARN to assume | `arn:aws:iam::123456789012:role/remote/MyRole` |
| `POST2POST_TAILNET_KEY` | Tailscale auth key | `tskey-auth-your-key-here` |
| `POST2POST_SESSION_NAME` | Session name for assumed role | `my-application` |
| `POST2POST_DURATION` | Credential duration | `2h`, `30m`, `45m` |
| `POST2POST_TIMEOUT` | Request timeout | `30s`, `1m`, `2m` |

## AWS CLI Configuration

### Basic Configuration

Add to your AWS config file (`~/.aws/config`):

```ini
[profile myprofile]
credential_process = /usr/local/bin/post2post-credentials --lambda-url https://your-lambda-url.amazonaws.com/ --role-arn arn:aws:iam::123456789012:role/remote/MyRole --tailnet-key tskey-auth-xyz
region = us-east-1
```

### Environment Variable Configuration

```ini
[profile myprofile]
credential_process = /usr/local/bin/post2post-credentials
region = us-east-1
```

Set environment variables:
```bash
export POST2POST_LAMBDA_URL="https://your-lambda-url.amazonaws.com/"
export POST2POST_ROLE_ARN="arn:aws:iam::123456789012:role/remote/MyRole"
export POST2POST_TAILNET_KEY="tskey-auth-your-key-here"
```

### Multiple Profiles

```ini
[profile dev]
credential_process = /usr/local/bin/post2post-credentials --role-arn arn:aws:iam::123456789012:role/remote/DevRole --session-name dev-session
region = us-east-1

[profile prod]
credential_process = /usr/local/bin/post2post-credentials --role-arn arn:aws:iam::987654321098:role/remote/ProdRole --session-name prod-session
region = us-west-2
```

## Usage Examples

### AWS CLI

```bash
# Use specific profile
aws s3 ls --profile myprofile

# List EC2 instances
aws ec2 describe-instances --profile myprofile

# Default profile (if configured)
aws sts get-caller-identity
```

### AWS SDK for Go

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
    // Load AWS configuration (will use credential_process)
    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithSharedConfigProfile("myprofile"))
    if err != nil {
        panic(err)
    }
    
    // Create S3 client
    client := s3.NewFromConfig(cfg)
    
    // Use the client
    result, err := client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
    if err != nil {
        panic(err)
    }
    
    for _, bucket := range result.Buckets {
        fmt.Printf("Bucket: %s\n", *bucket.Name)
    }
}
```

### Direct Testing

Test the credentials process directly:

```bash
# Test with environment variables
export POST2POST_LAMBDA_URL="https://your-lambda-url.amazonaws.com/"
export POST2POST_ROLE_ARN="arn:aws:iam::123456789012:role/remote/TestRole"
export POST2POST_TAILNET_KEY="tskey-auth-test-key"

./post2post-credentials

# Test with command line arguments
./post2post-credentials \
  --lambda-url "https://your-lambda-url.amazonaws.com/" \
  --role-arn "arn:aws:iam::123456789012:role/remote/TestRole" \
  --tailnet-key "tskey-auth-test-key" \
  --session-name "test-session"
```

## JSON Output Format

The program outputs AWS credentials in the standard `credential_process` JSON format:

```json
{
  "Version": 1,
  "AccessKeyId": "AKIA...",
  "SecretAccessKey": "...",
  "SessionToken": "...",
  "Expiration": "2023-12-25T12:00:00Z"
}
```

## Security Considerations

### Network Security
- **End-to-End Encryption**: All communication encrypted via Tailscale
- **Mesh Networking**: No exposure to public internet required
- **Domain Validation**: Lambda validates callback URLs against tailnet domain

### Credential Security
- **Temporary Credentials**: All credentials are temporary with configurable expiration
- **No Persistence**: Credentials output to stdout only, never stored
- **Automatic Refresh**: AWS CLI handles refresh before expiration
- **Role Path Restriction**: Only roles in `/remote/` path can be assumed

### Access Control
- **Path Restrictions**: Program validates `/remote/` path requirement
- **Account Isolation**: Lambda restricted to same AWS account roles
- **Audit Trails**: Session names provide CloudTrail audit visibility

## Error Handling

### Exit Codes
- **0**: Success - credentials output to stdout
- **1**: Error - details logged to stderr

### Common Issues

| Error | Cause | Solution |
|-------|-------|----------|
| "lambda URL is required" | Missing configuration | Set `--lambda-url` or `POST2POST_LAMBDA_URL` |
| "role ARN must be in /remote/ path" | Invalid role ARN | Use role ARN with `/remote/` path |
| "failed to retrieve credentials" | Network or Lambda error | Check Lambda status and network connectivity |
| "invalid duration format" | Malformed duration | Use format like `1h`, `30m`, `45m` |

### Debugging

Enable detailed logging by checking stderr output:

```bash
# Redirect stderr to see detailed logs
./post2post-credentials 2>debug.log

# Or view logs in real-time
./post2post-credentials 2>&1 | grep "post2post-credentials:"
```

## Building and Development

### Development Setup

```bash
# Clone repository
git clone https://github.com/pgdad/post2post.git
cd post2post/examples/go-credentials-process

# Initialize module
go mod tidy

# Run tests
go test ./...

# Build for development
go build -o post2post-credentials main.go
```

### Cross-Platform Building

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o post2post-credentials-linux main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o post2post-credentials-macos main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o post2post-credentials.exe main.go
```

## Dependencies

### Runtime Dependencies
- **AWS SDK for Go v2**: AWS credential types and configuration
- **Post2Post Library**: Core post2post functionality

### No External Dependencies
The binary is self-contained and has no runtime dependencies beyond the post2post library.

## Performance

### Credential Retrieval Time
- **First Call**: ~2-3 seconds (includes Lambda cold start)
- **Subsequent Calls**: Handled by AWS CLI caching
- **Timeout**: Configurable (default 30 seconds)

### Resource Usage
- **Memory**: Minimal (~10MB)
- **CPU**: Low (mostly network I/O)
- **Network**: Single HTTPS request per credential retrieval

## Integration Examples

### Docker

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o post2post-credentials main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/post2post-credentials .
ENTRYPOINT ["./post2post-credentials"]
```

### CI/CD Pipeline

```yaml
# GitHub Actions example
name: AWS Deploy
on: [push]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Configure AWS credentials
        run: |
          mkdir -p ~/.aws
          cat > ~/.aws/config << EOF
          [default]
          credential_process = ./post2post-credentials
          region = us-east-1
          EOF
        env:
          POST2POST_LAMBDA_URL: ${{ secrets.LAMBDA_URL }}
          POST2POST_ROLE_ARN: ${{ secrets.ROLE_ARN }}
          POST2POST_TAILNET_KEY: ${{ secrets.TAILNET_KEY }}
      - name: Deploy to AWS
        run: aws s3 sync . s3://my-bucket/
```

## Troubleshooting

### Common Problems

1. **Permission Denied**
   ```bash
   chmod +x post2post-credentials
   ```

2. **Role Not Found**
   - Verify role ARN is correct
   - Check role is in `/remote/` path
   - Confirm Lambda has permission to assume role

3. **Network Timeout**
   - Check Tailscale connectivity
   - Verify Lambda URL is accessible
   - Increase timeout with `--timeout`

4. **Invalid JSON Output**
   - Check stderr for error messages
   - Verify all required configuration is provided

### Support

For issues and questions:
1. Check the [main post2post documentation](../../README.md)
2. Review the [AWS Lambda setup guide](../aws-lambda/README.md)
3. Open an issue in the main repository

This Go credentials process provides a secure, efficient way to integrate post2post credential retrieval with AWS CLI and SDKs.