# AWS Credentials Provider

The post2post library includes an AWS credentials provider that enables secure credential retrieval through Tailscale mesh networking. This provider communicates with an AWS Lambda function to assume IAM roles and return temporary credentials.

## Overview

```
AWS SDK Client → Post2Post Provider → Tailscale Network → Lambda Function → STS AssumeRole → Return Credentials
```

The AWS credentials provider:
1. **Implements** the `aws.CredentialsProvider` interface
2. **Uses** post2post's `RoundTripPost` for synchronous communication
3. **Secures** communication through Tailscale mesh networking
4. **Caches** credentials until near expiration
5. **Supports** credential invalidation and refresh

## Features

- **Secure Communication**: All credential requests go through Tailscale encrypted tunnels
- **Automatic Caching**: Credentials are cached until 5 minutes before expiration
- **Synchronous Operation**: Uses round-trip posting for immediate credential retrieval
- **Error Handling**: Comprehensive error handling with detailed error messages
- **Configurable Duration**: Support for custom credential duration (default 1 hour)
- **Session Management**: Configurable session names for audit trails

## Prerequisites

1. **AWS Lambda Deployment**: Deploy the `examples/aws-lambda` function
2. **Tailscale Network**: Both client and Lambda must be on the same tailnet
3. **IAM Role Configuration**: Target roles must be in `/remote/` path
4. **Network Access**: Lambda must be accessible via Tailscale domain

## Usage

### Basic Setup

```go
package main

import (
    "context"
    "log"
    
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/pgdad/post2post"
)

func main() {
    // Create the credentials provider
    provider, err := post2post.NewAWSCredentialsProvider(post2post.AWSCredentialsProviderConfig{
        LambdaURL:   "https://your-lambda.lambda-url.us-east-1.on.aws/",
        RoleARN:     "arn:aws:iam::123456789012:role/remote/MyRole",
        TailnetKey:  "tskey-auth-your-key-here",
        SessionName: "my-application",
        Duration:    1 * time.Hour,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Close()

    // Use with AWS SDK
    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithCredentialsProvider(provider),
        config.WithRegion("us-east-1"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Use AWS services normally
    s3Client := s3.NewFromConfig(cfg)
    buckets, err := s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
    // ... handle response
}
```

### Environment Variable Configuration

```bash
export AWS_LAMBDA_URL="https://your-lambda.lambda-url.us-east-1.on.aws/"
export AWS_ROLE_ARN="arn:aws:iam::123456789012:role/remote/MyRole"
export TAILSCALE_AUTH_KEY="tskey-auth-your-key-here"

go run examples/aws_credentials_example.go
```

### Advanced Configuration

```go
provider, err := post2post.NewAWSCredentialsProvider(post2post.AWSCredentialsProviderConfig{
    LambdaURL:   "https://your-lambda.lambda-url.us-east-1.on.aws/",
    RoleARN:     "arn:aws:iam::123456789012:role/remote/CrossAccountRole",
    TailnetKey:  "tskey-auth-your-key-here",
    SessionName: "cross-account-access-session",
    Duration:    4 * time.Hour, // Longer duration for batch jobs
})
```

## Configuration Options

### AWSCredentialsProviderConfig

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `LambdaURL` | `string` | Yes | Lambda Function URL endpoint |
| `RoleARN` | `string` | Yes | IAM Role ARN to assume (must be in `/remote/` path) |
| `TailnetKey` | `string` | Yes | Tailscale auth key for secure communication |
| `SessionName` | `string` | No | Session name for audit (default: "post2post-credentials-provider") |
| `Duration` | `time.Duration` | No | Credential lifetime (default: 1 hour, max: 12 hours) |

### Required IAM Role Configuration

The target IAM role must:
1. **Have path `/remote/`**: e.g., `arn:aws:iam::ACCOUNT:role/remote/MyRole`
2. **Trust the Lambda execution role** in its assume role policy
3. **Have necessary permissions** for your application's AWS operations

Example role creation:
```bash
aws iam create-role \
  --role-name MyApplicationRole \
  --path /remote/ \
  --assume-role-policy-document '{
    "Version": "2012-10-17",
    "Statement": [{
      "Effect": "Allow",
      "Principal": {"AWS": "arn:aws:iam::ACCOUNT:role/post2post-receiver-role"},
      "Action": "sts:AssumeRole"
    }]
  }'

aws iam attach-role-policy \
  --role-name MyApplicationRole \
  --policy-arn arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess
```

## API Reference

### Methods

#### `NewAWSCredentialsProvider(config AWSCredentialsProviderConfig) (*AWSCredentialsProvider, error)`
Creates a new AWS credentials provider instance.

#### `Retrieve(ctx context.Context) (aws.Credentials, error)`
Retrieves AWS credentials. Uses cached credentials if available and not expired.

#### `Close() error`
Stops the internal post2post server and cleans up resources.

#### `InvalidateCache()`
Forces the provider to fetch fresh credentials on the next `Retrieve()` call.

#### `GetRoleARN() string`
Returns the configured IAM role ARN.

#### `GetLambdaURL() string`
Returns the configured Lambda function URL.

#### `GetSessionName() string`
Returns the configured session name.

### Credential Caching

- **Cache Duration**: Until 5 minutes before credential expiration
- **Cache Key**: Based on role ARN and session name
- **Thread Safety**: All operations are thread-safe with read-write mutexes
- **Invalidation**: Manual via `InvalidateCache()` or automatic on expiration

## Security Considerations

### Network Security
- **Tailscale Encryption**: All communication is end-to-end encrypted
- **Mesh Networking**: No exposure to public internet required
- **Domain Validation**: Lambda validates callback URLs against configured tailnet domain

### Credential Security
- **Temporary Credentials**: All credentials are temporary with configurable expiration
- **No Persistent Storage**: Credentials are only cached in memory
- **Least Privilege**: IAM roles should follow principle of least privilege
- **Audit Trail**: Session names provide audit trails in CloudTrail

### Access Control
- **Path Restrictions**: Only roles in `/remote/` path can be assumed
- **Account Isolation**: Lambda can only assume roles in the same AWS account
- **Network Isolation**: Communication restricted to Tailscale network

## Error Handling

Common errors and solutions:

### `lambda URL is required`
- Ensure `LambdaURL` is provided in configuration

### `role ARN is required`
- Ensure `RoleARN` is provided and follows format: `arn:aws:iam::ACCOUNT:role/remote/ROLE_NAME`

### `tailnet key is required for secure communication`
- Ensure `TailnetKey` is provided with valid Tailscale auth key

### `failed to retrieve credentials from Lambda`
- Check Lambda function is running and accessible
- Verify Tailscale connectivity between client and Lambda
- Check Lambda logs for processing errors

### `Lambda returned error status`
- Check IAM role exists and is assumable by Lambda execution role
- Verify role has `/remote/` path
- Check Lambda CloudWatch logs for detailed error information

## Performance Considerations

### Credential Caching
- **First Call**: ~2-3 seconds (includes Tailscale connection + Lambda cold start)
- **Cached Calls**: ~1ms (memory lookup)
- **Refresh**: ~1-2 seconds (Lambda warm start)

### Optimization Tips
1. **Reuse Provider Instance**: Create once and reuse across requests
2. **Configure Longer Duration**: For batch jobs, use longer credential duration
3. **Monitor Expiration**: Use credentials well before expiration
4. **Handle Errors**: Implement retry logic for transient network errors

## Integration Examples

### With AWS SDK v2
```go
cfg, err := config.LoadDefaultConfig(context.TODO(),
    config.WithCredentialsProvider(provider),
)
```

### With Multiple Services
```go
s3Client := s3.NewFromConfig(cfg)
ec2Client := ec2.NewFromConfig(cfg)
rdsClient := rds.NewFromConfig(cfg)
```

### With Custom Context
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

credentials, err := provider.Retrieve(ctx)
```

This AWS credentials provider enables secure, scalable credential management for applications running in distributed environments with Tailscale networking.