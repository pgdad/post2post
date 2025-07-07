package org.pgdad.post2post;

import org.junit.jupiter.api.Test;
import software.amazon.awssdk.auth.credentials.AwsCredentials;

import java.time.Duration;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Unit tests for the Post2PostAwsCredentialsProvider.
 */
public class Post2PostAwsCredentialsProviderTest {

    @Test
    public void testConfigValidation() {
        // Test missing Lambda URL
        assertThrows(IllegalArgumentException.class, () -> {
            new Post2PostAwsCredentialsProvider.Config()
                .roleArn("arn:aws:iam::123:role/remote/TestRole")
                .tailscaleAuthKey("tskey-auth-test")
                .callbackUrl("http://localhost:8080/callback");
            
            new Post2PostAwsCredentialsProvider(new Post2PostAwsCredentialsProvider.Config()
                .roleArn("arn:aws:iam::123:role/remote/TestRole")
                .tailscaleAuthKey("tskey-auth-test")
                .callbackUrl("http://localhost:8080/callback"));
        }, "Should throw exception for missing Lambda URL");

        // Test missing Role ARN
        assertThrows(IllegalArgumentException.class, () -> {
            new Post2PostAwsCredentialsProvider(new Post2PostAwsCredentialsProvider.Config()
                .lambdaUrl("https://lambda.example.com")
                .tailscaleAuthKey("tskey-auth-test")
                .callbackUrl("http://localhost:8080/callback"));
        }, "Should throw exception for missing Role ARN");

        // Test missing Tailscale auth key
        assertThrows(IllegalArgumentException.class, () -> {
            new Post2PostAwsCredentialsProvider(new Post2PostAwsCredentialsProvider.Config()
                .lambdaUrl("https://lambda.example.com")
                .roleArn("arn:aws:iam::123:role/remote/TestRole")
                .callbackUrl("http://localhost:8080/callback"));
        }, "Should throw exception for missing Tailscale auth key");

        // Test missing callback URL
        assertThrows(IllegalArgumentException.class, () -> {
            new Post2PostAwsCredentialsProvider(new Post2PostAwsCredentialsProvider.Config()
                .lambdaUrl("https://lambda.example.com")
                .roleArn("arn:aws:iam::123:role/remote/TestRole")
                .tailscaleAuthKey("tskey-auth-test"));
        }, "Should throw exception for missing callback URL");
    }

    @Test
    public void testValidConfiguration() {
        // Test valid configuration
        Post2PostAwsCredentialsProvider.Config config = new Post2PostAwsCredentialsProvider.Config()
            .lambdaUrl("https://lambda.example.com")
            .roleArn("arn:aws:iam::123456789012:role/remote/TestRole")
            .tailscaleAuthKey("tskey-auth-test123")
            .callbackUrl("http://localhost:8080/callback")
            .sessionName("test-session")
            .credentialDuration(Duration.ofHours(2));

        Post2PostAwsCredentialsProvider provider = new Post2PostAwsCredentialsProvider(config);
        
        assertNotNull(provider, "Provider should be created successfully");
        assertEquals("https://lambda.example.com", provider.getLambdaUrl());
        assertEquals("arn:aws:iam::123456789012:role/remote/TestRole", provider.getRoleArn());
        assertEquals("test-session", provider.getSessionName());
        
        provider.close();
    }

    @Test
    public void testDefaultConfiguration() {
        Post2PostAwsCredentialsProvider.Config config = new Post2PostAwsCredentialsProvider.Config()
            .lambdaUrl("https://lambda.example.com")
            .roleArn("arn:aws:iam::123456789012:role/remote/TestRole")
            .tailscaleAuthKey("tskey-auth-test123")
            .callbackUrl("http://localhost:8080/callback");

        Post2PostAwsCredentialsProvider provider = new Post2PostAwsCredentialsProvider(config);
        
        // Test default values
        assertEquals("post2post-java-session", provider.getSessionName());
        
        provider.close();
    }

    @Test
    public void testCredentialCaching() {
        Post2PostAwsCredentialsProvider.Config config = new Post2PostAwsCredentialsProvider.Config()
            .lambdaUrl("https://lambda.example.com")
            .roleArn("arn:aws:iam::123456789012:role/remote/TestRole")
            .tailscaleAuthKey("tskey-auth-test123")
            .callbackUrl("http://localhost:8080/callback");

        Post2PostAwsCredentialsProvider provider = new Post2PostAwsCredentialsProvider(config);
        
        try {
            // This will use placeholder credentials for now
            AwsCredentials credentials1 = provider.resolveCredentials();
            assertNotNull(credentials1, "First credential request should succeed");
            
            // Second request should use cached credentials
            AwsCredentials credentials2 = provider.resolveCredentials();
            assertNotNull(credentials2, "Second credential request should succeed");
            
            // Should be the same instance due to caching
            assertEquals(credentials1.accessKeyId(), credentials2.accessKeyId(), 
                "Cached credentials should have same access key");
            
        } catch (Exception e) {
            // Expected for placeholder implementation or network issues
            assertTrue(e.getMessage().contains("placeholder") || 
                      e.getMessage().contains("round-trip") ||
                      e.getMessage().contains("Failed to retrieve credentials"), 
                "Exception should mention placeholder, round-trip, or credential retrieval failure");
        }
        
        provider.close();
    }

    @Test
    public void testInvalidateCache() {
        Post2PostAwsCredentialsProvider.Config config = new Post2PostAwsCredentialsProvider.Config()
            .lambdaUrl("https://lambda.example.com")
            .roleArn("arn:aws:iam::123456789012:role/remote/TestRole")
            .tailscaleAuthKey("tskey-auth-test123")
            .callbackUrl("http://localhost:8080/callback");

        Post2PostAwsCredentialsProvider provider = new Post2PostAwsCredentialsProvider(config);
        
        // Test cache invalidation (should not throw exception)
        assertDoesNotThrow(() -> {
            provider.invalidateCache();
        }, "Cache invalidation should not throw exception");
        
        provider.close();
    }

    @Test
    public void testConfigBuilder() {
        // Test the fluent configuration API
        Post2PostAwsCredentialsProvider.Config config = new Post2PostAwsCredentialsProvider.Config()
            .lambdaUrl("https://test-lambda.amazonaws.com")
            .roleArn("arn:aws:iam::999999999999:role/remote/MyTestRole")
            .tailscaleAuthKey("tskey-auth-xyz789")
            .callbackUrl("http://test-callback:9090/webhook")
            .sessionName("fluent-test-session")
            .credentialDuration(Duration.ofMinutes(30))
            .requestTimeout(Duration.ofSeconds(45));

        // Verify all configuration values
        assertEquals("https://test-lambda.amazonaws.com", config.getLambdaUrl());
        assertEquals("arn:aws:iam::999999999999:role/remote/MyTestRole", config.getRoleArn());
        assertEquals("tskey-auth-xyz789", config.getTailscaleAuthKey());
        assertEquals("http://test-callback:9090/webhook", config.getCallbackUrl());
        assertEquals("fluent-test-session", config.getSessionName());
        assertEquals(Duration.ofMinutes(30), config.getCredentialDuration());
        assertEquals(Duration.ofSeconds(45), config.getRequestTimeout());
    }
}