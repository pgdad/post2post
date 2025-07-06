# AWS Lambda Post2Post Receiver

This example demonstrates how to implement the receiving side of the post2post system using AWS Lambda with a Lambda Function URL. The Lambda function receives webhook requests, assumes specified IAM roles, and responds using optional Tailscale networking.

## Features

- **AWS Lambda Integration**: Serverless webhook receiver using Lambda Function URL
- **IAM Role Assumption**: Assumes roles specified in the payload and returns STS credentials
- **Tailscale Integration**: Optional secure networking for response posting
- **Async Processing**: Returns immediate acknowledgment and processes requests asynchronously
- **Comprehensive Error Handling**: Proper error responses and logging
- **Multiple Deployment Options**: Terraform, CloudFormation, and AWS CLI support

## Architecture

```
Client (post2post) → Lambda Function URL → Lambda Function
                                              ↓
                                         STS AssumeRole
                                              ↓
                                    [Optional Tailscale] → Client Callback
```

## Request Format

The Lambda function expects POST requests with the following JSON structure:

```json
{
  "url": "http://client-callback-url/roundtrip",
  "payload": {
    "message": "Original payload data",
    "user_id": 12345
  },
  "request_id": "req_1234567890",
  "tailnet_key": "tskey-auth-xyz123...",
  "role_arn": "arn:aws:iam::123456789012:role/ExampleRole"
}
```

### Required Fields

- `url`: Callback URL where the response should be posted
- `role_arn`: AWS IAM Role ARN to assume via STS
- `request_id`: Unique identifier for request tracking

### Optional Fields

- `payload`: Any JSON-serializable data (will be included in response)
- `tailnet_key`: Tailscale auth key for secure response posting

## Response Format

The Lambda function posts back a response with the following structure:

```json
{
  "request_id": "req_1234567890",
  "payload": {
    "original_payload": {
      "message": "Original payload data",
      "user_id": 12345
    },
    "assume_role_result": {
      "credentials": {
        "access_key_id": "ASIA...",
        "secret_access_key": "...",
        "session_token": "...",
        "expiration": "2023-12-01T15:30:00Z"
      },
      "assumed_role_user": {
        "arn": "arn:aws:sts::123456789012:assumed-role/ExampleRole/post2post-req_1234567890-1701234567",
        "assumed_role_id": "AROA....:post2post-req_1234567890-1701234567"
      }
    },
    "processed_at": "2023-12-01 14:30:25 MST",
    "processed_by": "aws-lambda-post2post-receiver",
    "lambda_request_id": "12345678-1234-1234-1234-123456789012",
    "status": "success"
  },
  "tailnet_key": "tskey-auth-xyz123..."
}
```

## Prerequisites

1. **AWS Account** with appropriate permissions
2. **Go 1.21+** for building the function
3. **AWS CLI** configured with credentials
4. **Terraform** (optional, for Terraform deployment)
5. **Tailscale Account** (optional, for secure networking)

## Environment Variables

The Lambda function requires the following environment variables:

### Required

- `TAILNET_DOMAIN`: Your Tailscale tailnet domain (e.g., `example.ts.net`)
  - **Critical**: This validates callback URLs for security
  - **Failure**: Lambda execution will fail if not configured
  - **Validation**: Callback URLs must end with this domain

### Optional

- `TAILSCALE_AUTH_KEY`: Tailscale auth key for secure networking (e.g., `tskey-auth-...`)
  - Used for creating secure connections through Tailscale mesh
  - If not provided, standard HTTP is used for responses
  - **Note**: Not configured via infrastructure templates - set manually if needed

## IAM Permissions

The Lambda function needs the following IAM permissions:

### Lambda Execution Role

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "sts:AssumeRole"
      ],
      "Resource": "*"
    }
  ]
}
```

**Security Note**: In production, restrict the `sts:AssumeRole` resource to specific role ARNs rather than using `*`.

### Target Roles

The roles you want to assume must trust the Lambda execution role:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::YOUR_ACCOUNT:role/post2post-receiver-role"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

## Building and Deployment

### Step 1: Build the Function

```bash
cd examples/aws-lambda
./build.sh
```

This creates `bootstrap.zip` ready for Lambda deployment.

### Step 2: Deploy (Choose One Method)

#### Option A: Terraform Deployment

```bash
cd terraform
terraform init
terraform apply

# With custom variables
terraform apply \
  -var="function_name=my-post2post" \
  -var="tailnet_domain=example.ts.net"
```

#### Option B: CloudFormation Deployment

```bash
aws cloudformation deploy \
  --template-file cloudformation/template.yaml \
  --stack-name post2post-receiver \
  --capabilities CAPABILITY_IAM \
  --parameter-overrides \
    TailnetDomain=example.ts.net
```

#### Option C: AWS CLI Deployment

```bash
# Create IAM role first
aws iam create-role --role-name post2post-receiver-role --assume-role-policy-document file://trust-policy.json

# Attach policies
aws iam attach-role-policy --role-name post2post-receiver-role --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole

# Create function
aws lambda create-function \
  --function-name post2post-receiver \
  --runtime provided.al2023 \
  --role arn:aws:iam::YOUR_ACCOUNT:role/post2post-receiver-role \
  --handler bootstrap \
  --zip-file fileb://bootstrap.zip \
  --environment Variables='{TAILNET_DOMAIN=example.ts.net}'

# Create function URL
aws lambda create-function-url-config \
  --function-name post2post-receiver \
  --auth-type NONE \
  --cors AllowOrigins="*",AllowMethods="POST",AllowHeaders="*"
```

## Usage Examples

### Basic Usage (No Tailscale)

**Note**: The callback URL must end with your configured `TAILNET_DOMAIN`.

```bash
curl -X POST https://your-lambda-url.lambda-url.us-east-1.on.aws/ \
  -H "Content-Type: application/json" \
  -d '{
    "url": "http://your-callback-server.example.ts.net/roundtrip",
    "payload": {"message": "Hello from client"},
    "request_id": "test-123",
    "role_arn": "arn:aws:iam::123456789012:role/TestRole"
  }'
```

### With Tailscale Integration

```bash
curl -X POST https://your-lambda-url.lambda-url.us-east-1.on.aws/ \
  -H "Content-Type: application/json" \
  -d '{
    "url": "http://secure-server.example.ts.net/roundtrip",
    "payload": {"secure": "data"},
    "request_id": "secure-123",
    "role_arn": "arn:aws:iam::123456789012:role/SecureRole",
    "tailnet_key": "tskey-auth-your-key-here"
  }'
```

### Integration with Post2Post Client

Update your post2post client to use the Lambda URL:

```go
server := post2post.NewServer().
    WithInterface("127.0.0.1").
    WithPostURL("https://your-lambda-url.lambda-url.us-east-1.on.aws/").
    WithTimeout(30 * time.Second)

// Your payload must include role_arn
payload := map[string]interface{}{
    "message": "Hello from post2post",
    "role_arn": "arn:aws:iam::123456789012:role/MyRole",
}

response, err := server.RoundTripPost(payload)
```

## Tailscale Integration

### Setup

1. **Get Tailscale Auth Key**: Generate an auth key from your Tailscale admin console
2. **Enable in Code**: Uncomment the tsnet integration in `createTailscaleClient()`
3. **Add Dependency**: Uncomment the tailscale.com requirement in go.mod
4. **Deploy with Key**: Set the auth key in your deployment configuration

### Full Tailscale Integration

To enable complete Tailscale functionality:

1. Uncomment the tsnet code in `main.go`:
```go
import "tailscale.com/tsnet"

srv := &tsnet.Server{
    Hostname: "lambda-post2post-receiver",
    AuthKey:  tailnetKey,
    Ephemeral: true, // Lambda instances are ephemeral
}

client := srv.HTTPClient()
return client, nil
```

2. Update go.mod:
```go
require tailscale.com v1.76.1
```

3. Rebuild and redeploy

## Monitoring and Troubleshooting

### CloudWatch Logs

Monitor Lambda execution in CloudWatch Logs:

```bash
aws logs tail /aws/lambda/post2post-receiver --follow
```

### Common Issues

1. **Role Assumption Fails**
   - Check IAM permissions on both Lambda role and target role
   - Verify role ARN format
   - Ensure trust relationship is configured

2. **Callback Fails**
   - Verify callback URL is accessible
   - Check network security groups/firewalls
   - Test Tailscale connectivity if using

3. **Function Timeout**
   - Increase Lambda timeout (max 15 minutes)
   - Optimize role assumption calls
   - Consider async processing for long operations

### Testing

Test the function locally (without Lambda runtime):

```bash
go run main.go
# Function will log that it's not in Lambda environment
```

Test with sample payload:

```bash
echo '{
  "url": "http://httpbin.org/post",
  "payload": {"test": "data"},
  "request_id": "test-123",
  "role_arn": "arn:aws:iam::123456789012:role/TestRole"
}' | curl -X POST https://your-lambda-url/ -d @-
```

## Security Considerations

1. **Restrict Role Access**: Limit which roles can be assumed
2. **Network Security**: Use VPC configuration for private resources
3. **Auth Keys**: Rotate Tailscale auth keys regularly
4. **Input Validation**: Validate all input parameters
5. **Logging**: Monitor for suspicious role assumption patterns
6. **Function URL**: Consider authentication for production use

## Cost Optimization

- **Memory**: Start with 128MB, increase if needed
- **Timeout**: Set appropriate timeout (30s default)
- **Provisioned Concurrency**: Only if consistent low latency required
- **Architecture**: Consider ARM64 for cost savings

## Production Considerations

1. **Error Handling**: Implement retry logic and dead letter queues
2. **Monitoring**: Set up CloudWatch alarms for errors and duration
3. **Scaling**: Lambda auto-scales, but monitor concurrent executions
4. **Secrets**: Use AWS Secrets Manager for sensitive configuration
5. **VPC**: Configure VPC if accessing private resources
6. **Reserved Concurrency**: Set limits if needed for cost control