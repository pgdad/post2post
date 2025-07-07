package org.pgdad.post2post;

import okhttp3.mockwebserver.MockResponse;
import okhttp3.mockwebserver.MockWebServer;
import okhttp3.mockwebserver.RecordedRequest;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.time.Duration;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.TimeUnit;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Unit tests for the Post2PostClient.
 */
public class Post2PostClientTest {
    
    private MockWebServer mockServer;
    private Post2PostClient client;

    @BeforeEach
    public void setUp() throws IOException {
        mockServer = new MockWebServer();
        mockServer.start();
        
        String callbackUrl = mockServer.url("/callback").toString();
        
        Post2PostClient.Config config = new Post2PostClient.Config()
            .callbackUrl(callbackUrl)
            .timeout(Duration.ofSeconds(5));
            
        client = new Post2PostClient(config);
    }

    @AfterEach
    public void tearDown() throws IOException {
        if (client != null) {
            client.close();
        }
        if (mockServer != null) {
            mockServer.shutdown();
        }
    }

    @Test
    public void testPostSuccess() throws Exception {
        // Mock successful response
        mockServer.enqueue(new MockResponse()
            .setResponseCode(200)
            .setBody("{\"status\": \"success\"}")
            .addHeader("Content-Type", "application/json"));

        // Test POST request
        String targetUrl = mockServer.url("/test").toString();
        CompletableFuture<Boolean> future = client.post(targetUrl, "test payload");
        
        Boolean result = future.get(5, TimeUnit.SECONDS);
        assertTrue(result, "POST request should succeed");

        // Verify request
        RecordedRequest request = mockServer.takeRequest();
        assertEquals("POST", request.getMethod());
        assertEquals("/test", request.getPath());
        assertTrue(request.getHeader("Content-Type").startsWith("application/json"));
        assertTrue(request.getHeader("User-Agent").contains("post2post-java-client"));
    }

    @Test
    public void testPostFailure() throws Exception {
        // Mock error response
        mockServer.enqueue(new MockResponse()
            .setResponseCode(500)
            .setBody("Internal Server Error"));

        // Test POST request
        String targetUrl = mockServer.url("/test").toString();
        CompletableFuture<Boolean> future = client.post(targetUrl, "test payload");
        
        Boolean result = future.get(5, TimeUnit.SECONDS);
        assertFalse(result, "POST request should fail with 500 status");
    }

    @Test
    public void testRoundTripPostSuccess() throws Exception {
        // Mock successful acknowledgment
        mockServer.enqueue(new MockResponse()
            .setResponseCode(200)
            .setBody("{\"status\": \"accepted\", \"message\": \"Processing request\"}")
            .addHeader("Content-Type", "application/json"));

        // Test round-trip POST request
        String targetUrl = mockServer.url("/lambda").toString();
        CompletableFuture<Post2PostClient.RoundTripResponse<String>> future = 
            client.roundTripPost(targetUrl, "test payload", "arn:aws:iam::123:role/remote/TestRole", 
                              "tskey-auth-test", String.class);
        
        Post2PostClient.RoundTripResponse<String> result = future.get(5, TimeUnit.SECONDS);
        assertTrue(result.isSuccess(), "Round-trip POST should succeed");
        assertFalse(result.isTimeout(), "Round-trip POST should not timeout");
        assertNotNull(result.getRequestId(), "Request ID should be generated");

        // Verify request structure
        RecordedRequest request = mockServer.takeRequest();
        assertEquals("POST", request.getMethod());
        assertEquals("/lambda", request.getPath());
        
        // Verify JSON structure contains required fields
        String body = request.getBody().readUtf8();
        assertTrue(body.contains("\"url\""), "Request should contain callback URL");
        assertTrue(body.contains("\"payload\""), "Request should contain payload");
        assertTrue(body.contains("requestId") || body.contains("request_id"), "Request should contain request ID");
        assertTrue(body.contains("roleArn") || body.contains("role_arn"), "Request should contain role ARN");
        assertTrue(body.contains("tailnetKey") || body.contains("tailnet_key"), "Request should contain tailnet key");
    }

    @Test
    public void testRoundTripPostWithoutTailscale() throws Exception {
        // Mock successful acknowledgment
        mockServer.enqueue(new MockResponse()
            .setResponseCode(200)
            .setBody("{\"status\": \"accepted\"}")
            .addHeader("Content-Type", "application/json"));

        // Test round-trip POST without Tailscale key
        String targetUrl = mockServer.url("/lambda").toString();
        CompletableFuture<Post2PostClient.RoundTripResponse<String>> future = 
            client.roundTripPost(targetUrl, "test payload", "arn:aws:iam::123:role/remote/TestRole", 
                              null, String.class);
        
        Post2PostClient.RoundTripResponse<String> result = future.get(5, TimeUnit.SECONDS);
        assertTrue(result.isSuccess(), "Round-trip POST should succeed without Tailscale");

        // Verify request doesn't contain tailnet_key when not provided
        RecordedRequest request = mockServer.takeRequest();
        String body = request.getBody().readUtf8();
        assertFalse(body.contains("tailnetKey\":\"\")") && !body.contains("tailnet_key\":\"\""), "Request should not contain empty tailnet key");
    }

    @Test
    public void testRoundTripPostTimeout() throws Exception {
        // Don't enqueue any response to simulate timeout
        // The client should timeout waiting for response

        String targetUrl = mockServer.url("/lambda").toString();
        CompletableFuture<Post2PostClient.RoundTripResponse<String>> future = 
            client.roundTripPost(targetUrl, "test payload", "arn:aws:iam::123:role/remote/TestRole", 
                              "tskey-auth-test", String.class, Duration.ofMillis(100));
        
        Post2PostClient.RoundTripResponse<String> result = future.get(5, TimeUnit.SECONDS);
        assertFalse(result.isSuccess(), "Round-trip POST should fail on timeout");
        assertTrue(result.isTimeout(), "Round-trip POST should indicate timeout");
        assertTrue(result.getError().contains("timed out"), "Error should mention timeout");
    }

    @Test
    public void testInvalidUrl() throws Exception {
        // Test with invalid URL
        CompletableFuture<Boolean> future = client.post("not-a-valid-url", "test payload");
        
        Boolean result = future.get(5, TimeUnit.SECONDS);
        assertFalse(result, "POST to invalid URL should fail");
    }
}