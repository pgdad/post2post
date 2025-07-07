# Post2Post AWS Credentials Provider - Java

A Java library that implements AWS Credentials Provider using Post2Post for secure credential retrieval through Tailscale mesh networks. This library provides seamless integration with AWS SDK for Java v2.

## Overview

This library enables Java applications to securely retrieve AWS credentials from a remote Lambda function through Tailscale encrypted channels. It implements the AWS SDK v2 `AwsCredentialsProvider` interface for drop-in compatibility with existing AWS applications.

```
Java Application → Post2Post Client → Tailscale Network → Lambda Function → STS AssumeRole → Return Credentials
```

## Features

- **AWS SDK v2 Compatible**: Implements `AwsCredentialsProvider` interface
- **Secure Communication**: All communication through Tailscale mesh networking
- **Automatic Caching**: Credentials cached until 5 minutes before expiration
- **Thread-Safe**: All operations are thread-safe with proper synchronization
- **Configurable**: Flexible configuration for timeouts, session names, and durations
- **Comprehensive Logging**: SLF4J-based logging for debugging and monitoring

## Prerequisites

- **Java 11+**: Minimum Java version requirement
- **Maven 3.6+**: For building the library
- **AWS Lambda**: Deployed post2post Lambda function
- **Tailscale Network**: Both client and Lambda on same tailnet
- **IAM Roles**: Target roles must be in `/remote/` path

## Maven Dependency

Add to your `pom.xml`:

```xml
<dependency>
    <groupId>org.pgdad.post2post</groupId>
    <artifactId>aws-credentials-provider</artifactId>
    <version>1.0.0</version>
</dependency>
```

## Quick Start

### Basic Usage

```java
import org.pgdad.post2post.Post2PostAwsCredentialsProvider;
import software.amazon.awssdk.auth.credentials.AwsCredentialsProvider;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.s3.S3Client;

public class Example {
    public static void main(String[] args) {
        // Configure the credentials provider
        Post2PostAwsCredentialsProvider.Config config = 
            new Post2PostAwsCredentialsProvider.Config()
                .lambdaUrl("https://your-lambda.lambda-url.us-east-1.on.aws/")
                .roleArn("arn:aws:iam::123456789012:role/remote/MyRole")
                .tailscaleAuthKey("tskey-auth-your-key-here")
                .callbackUrl("http://your-callback-server.tailnet.ts.net:8080/callback")
                .sessionName("my-java-application");

        // Create the provider
        AwsCredentialsProvider credentialsProvider = 
            new Post2PostAwsCredentialsProvider(config);

        // Use with AWS services
        S3Client s3Client = S3Client.builder()
            .credentialsProvider(credentialsProvider)
            .region(Region.US_EAST_1)
            .build();

        // Use S3 client normally
        var buckets = s3Client.listBuckets();
        buckets.buckets().forEach(bucket -> 
            System.out.println("Bucket: " + bucket.name()));
    }
}
```

### Environment Variable Configuration

```java
public class EnvironmentExample {
    public static void main(String[] args) {
        Post2PostAwsCredentialsProvider.Config config = 
            new Post2PostAwsCredentialsProvider.Config()
                .lambdaUrl(System.getenv("AWS_LAMBDA_URL"))
                .roleArn(System.getenv("AWS_ROLE_ARN"))
                .tailscaleAuthKey(System.getenv("TAILSCALE_AUTH_KEY"))
                .callbackUrl(System.getenv("CALLBACK_URL"));

        AwsCredentialsProvider provider = new Post2PostAwsCredentialsProvider(config);
        
        // Use with any AWS service
        // ...
    }
}
```

### Advanced Configuration

```java
Post2PostAwsCredentialsProvider.Config config = 
    new Post2PostAwsCredentialsProvider.Config()
        .lambdaUrl("https://lambda-url.amazonaws.com/")
        .roleArn("arn:aws:iam::123456789012:role/remote/CrossAccountRole")
        .tailscaleAuthKey("tskey-auth-your-key")
        .callbackUrl("http://callback.tailnet.ts.net:8080/webhook")
        .sessionName("batch-processing-job")
        .credentialDuration(Duration.ofHours(4))  // Longer duration for batch jobs
        .requestTimeout(Duration.ofSeconds(60));  // Longer timeout
```

## Configuration Options

### Post2PostAwsCredentialsProvider.Config

| Method | Type | Required | Description |
|--------|------|----------|-------------|
| `lambdaUrl(String)` | String | Yes | Lambda Function URL endpoint |
| `roleArn(String)` | String | Yes | IAM Role ARN to assume (must be in `/remote/` path) |
| `tailscaleAuthKey(String)` | String | Yes | Tailscale auth key for secure communication |
| `callbackUrl(String)` | String | Yes | Callback URL for receiving responses |
| `sessionName(String)` | String | No | Session name for audit (default: "post2post-java-session") |
| `credentialDuration(Duration)` | Duration | No | Credential lifetime (default: 1 hour, max: 12 hours) |
| `requestTimeout(Duration)` | Duration | No | Request timeout (default: 30 seconds) |

### Required IAM Role Configuration

The target IAM role must:
1. **Have path `/remote/`**: e.g., `arn:aws:iam::ACCOUNT:role/remote/MyRole`
2. **Trust the Lambda execution role** in its assume role policy
3. **Have necessary permissions** for your application's AWS operations

## Building from Source

```bash
# Clone the repository
git clone https://github.com/pgdad/post2post.git
cd post2post/examples/java-aws-credentials-provider

# Build the library
mvn clean compile

# Run tests
mvn test

# Package the JAR
mvn package

# Install to local repository
mvn install
```

### Build Requirements

- Java 11 or higher
- Maven 3.6 or higher
- Internet connection for dependency download

## API Reference

### Post2PostAwsCredentialsProvider

The main credentials provider class implementing `AwsCredentialsProvider`.

#### Methods

- `AwsCredentials resolveCredentials()` - Retrieves AWS credentials (cached or fresh)
- `void invalidateCache()` - Forces fresh credential fetch on next call
- `String getRoleArn()` - Returns configured IAM role ARN
- `String getSessionName()` - Returns configured session name
- `String getLambdaUrl()` - Returns configured Lambda URL
- `void close()` - Closes provider and releases resources

### Post2PostClient

Core HTTP client for Post2Post communication.

#### Methods

- `CompletableFuture<RoundTripResponse<T>> roundTripPost(...)` - Round-trip POST with response
- `CompletableFuture<Boolean> post(String, Object)` - Simple POST request
- `void close()` - Closes client and releases resources

## Security Considerations

### Network Security
- **End-to-End Encryption**: All communication encrypted via Tailscale
- **Mesh Networking**: No exposure to public internet required
- **Domain Validation**: Lambda validates callback URLs against tailnet domain

### Credential Security
- **Temporary Credentials**: All credentials are temporary with configurable expiration
- **Memory-Only Caching**: Credentials cached in memory only, never persisted
- **Automatic Expiry**: Credentials automatically refreshed before expiration
- **Thread-Safe Access**: Concurrent access properly synchronized

### Access Control
- **Path Restrictions**: Only roles in `/remote/` path can be assumed
- **Account Isolation**: Lambda restricted to same AWS account roles
- **Audit Trails**: Session names provide CloudTrail audit visibility

## Error Handling

### Common Exceptions

| Exception | Cause | Solution |
|-----------|-------|----------|
| `IllegalArgumentException` | Missing required configuration | Provide all required config values |
| `RuntimeException` | Lambda communication failure | Check Lambda status and network connectivity |
| `RuntimeException` | Credential parsing failure | Verify Lambda response format |
| `RuntimeException` | Authentication failure | Check Tailscale auth key and role permissions |

### Logging

The library uses SLF4J for logging. Configure your logging framework to see detailed output:

```xml
<!-- logback.xml example -->
<configuration>
    <logger name="org.pgdad.post2post" level="INFO"/>
    <root level="WARN">
        <appender-ref ref="STDOUT"/>
    </root>
</configuration>
```

## Testing

### Unit Tests

```bash
# Run all tests
mvn test

# Run specific test class
mvn test -Dtest=Post2PostClientTest

# Run with debug logging
mvn test -Dorg.slf4j.simpleLogger.defaultLogLevel=DEBUG
```

### Integration Testing

For integration testing with actual Lambda functions:

```bash
export AWS_LAMBDA_URL="https://your-lambda-url.amazonaws.com/"
export AWS_ROLE_ARN="arn:aws:iam::123456789012:role/remote/TestRole"
export TAILSCALE_AUTH_KEY="tskey-auth-your-key"
export CALLBACK_URL="http://your-callback.tailnet.ts.net:8080/callback"

mvn exec:java -Dexec.mainClass="org.pgdad.post2post.AwsCredentialsExample"
```

## Performance Considerations

### Credential Caching
- **First Call**: ~2-3 seconds (includes Lambda cold start)
- **Cached Calls**: ~1ms (memory lookup)
- **Refresh**: ~1-2 seconds (Lambda warm start)

### Optimization Tips
1. **Reuse Provider**: Create once and reuse across requests
2. **Configure Duration**: Use longer durations for batch jobs
3. **Monitor Expiry**: Refresh credentials before expiration
4. **Handle Timeouts**: Implement retry logic for network issues

## Dependencies

### Runtime Dependencies
- **AWS SDK for Java v2**: `software.amazon.awssdk:auth`, `software.amazon.awssdk:sts`
- **OkHttp**: HTTP client for network communication
- **Jackson**: JSON processing for request/response serialization
- **SLF4J**: Logging abstraction

### Test Dependencies
- **JUnit 5**: Unit testing framework
- **Mockito**: Mocking framework for unit tests
- **MockWebServer**: HTTP server mocking for integration tests

## Examples

See the `src/main/java/org/pgdad/post2post/AwsCredentialsExample.java` file for a complete example demonstrating:

- Credential provider configuration
- AWS service integration (STS, S3)
- Error handling
- Credential caching
- Cache invalidation

## License

This library is part of the post2post project. See the main repository for license information.

## Contributing

Contributions are welcome! Please see the main post2post repository for contribution guidelines.

## Support

For issues and questions:
1. Check the [main post2post documentation](../../README.md)
2. Review the [AWS Lambda setup guide](../aws-lambda/README.md)
3. Open an issue in the main repository

This Java library provides a robust, secure way to manage AWS credentials in distributed applications using Tailscale mesh networking.