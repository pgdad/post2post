package org.pgdad.post2post;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import okhttp3.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.time.Duration;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicLong;

/**
 * Java implementation of the post2post client for HTTP-to-HTTP communication.
 * This client enables round-trip communication patterns and supports Tailscale networking.
 */
public class Post2PostClient {
    private static final Logger logger = LoggerFactory.getLogger(Post2PostClient.class);
    private static final MediaType JSON = MediaType.get("application/json; charset=utf-8");
    private static final AtomicLong requestCounter = new AtomicLong(0);
    
    private final OkHttpClient httpClient;
    private final ObjectMapper objectMapper;
    private final String callbackUrl;
    private final Duration defaultTimeout;

    /**
     * Configuration for the Post2Post client.
     */
    public static class Config {
        private Duration timeout = Duration.ofSeconds(30);
        private String callbackUrl;
        private boolean enableTailscale = false;
        private String tailscaleAuthKey;

        public Config timeout(Duration timeout) {
            this.timeout = timeout;
            return this;
        }

        public Config callbackUrl(String callbackUrl) {
            this.callbackUrl = callbackUrl;
            return this;
        }

        public Config enableTailscale(String authKey) {
            this.enableTailscale = true;
            this.tailscaleAuthKey = authKey;
            return this;
        }

        public Duration getTimeout() { return timeout; }
        public String getCallbackUrl() { return callbackUrl; }
        public boolean isTailscaleEnabled() { return enableTailscale; }
        public String getTailscaleAuthKey() { return tailscaleAuthKey; }
    }

    /**
     * Request payload for Lambda function calls.
     */
    public static class LambdaRequest {
        private String url;
        private Object payload;
        private String requestId;
        private String tailnetKey;
        private String roleArn;

        public LambdaRequest() {}

        public LambdaRequest(String url, Object payload, String requestId, String roleArn) {
            this.url = url;
            this.payload = payload;
            this.requestId = requestId;
            this.roleArn = roleArn;
        }

        // Getters and setters
        public String getUrl() { return url; }
        public void setUrl(String url) { this.url = url; }
        
        public Object getPayload() { return payload; }
        public void setPayload(Object payload) { this.payload = payload; }
        
        public String getRequestId() { return requestId; }
        public void setRequestId(String requestId) { this.requestId = requestId; }
        
        public String getTailnetKey() { return tailnetKey; }
        public void setTailnetKey(String tailnetKey) { this.tailnetKey = tailnetKey; }
        
        public String getRoleArn() { return roleArn; }
        public void setRoleArn(String roleArn) { this.roleArn = roleArn; }
    }

    /**
     * Response from Lambda function.
     */
    public static class LambdaResponse<T> {
        private String requestId;
        private T payload;
        private String tailnetKey;

        public LambdaResponse() {}

        // Getters and setters
        public String getRequestId() { return requestId; }
        public void setRequestId(String requestId) { this.requestId = requestId; }
        
        public T getPayload() { return payload; }
        public void setPayload(T payload) { this.payload = payload; }
        
        public String getTailnetKey() { return tailnetKey; }
        public void setTailnetKey(String tailnetKey) { this.tailnetKey = tailnetKey; }
    }

    /**
     * Round-trip response wrapper.
     */
    public static class RoundTripResponse<T> {
        private final T payload;
        private final boolean success;
        private final String error;
        private final boolean timeout;
        private final String requestId;

        public RoundTripResponse(T payload, boolean success, String error, boolean timeout, String requestId) {
            this.payload = payload;
            this.success = success;
            this.error = error;
            this.timeout = timeout;
            this.requestId = requestId;
        }

        public T getPayload() { return payload; }
        public boolean isSuccess() { return success; }
        public String getError() { return error; }
        public boolean isTimeout() { return timeout; }
        public String getRequestId() { return requestId; }
    }

    /**
     * Creates a new Post2Post client with the specified configuration.
     */
    public Post2PostClient(Config config) {
        this.callbackUrl = config.getCallbackUrl();
        this.defaultTimeout = config.getTimeout();
        
        // Configure HTTP client
        OkHttpClient.Builder clientBuilder = new OkHttpClient.Builder()
            .connectTimeout(config.getTimeout())
            .writeTimeout(config.getTimeout())
            .readTimeout(config.getTimeout());

        // TODO: Add Tailscale support when available in Java
        if (config.isTailscaleEnabled()) {
            logger.warn("Tailscale support not yet implemented in Java client");
        }

        this.httpClient = clientBuilder.build();
        
        // Configure JSON mapper
        this.objectMapper = new ObjectMapper();
        this.objectMapper.registerModule(new JavaTimeModule());
        
        logger.info("Post2Post client initialized with callback URL: {}", callbackUrl);
    }

    /**
     * Performs a round-trip POST request with synchronous response.
     */
    public <T> CompletableFuture<RoundTripResponse<T>> roundTripPost(
            String targetUrl, 
            Object payload, 
            String roleArn,
            String tailscaleAuthKey,
            Class<T> responseType) {
        return roundTripPost(targetUrl, payload, roleArn, tailscaleAuthKey, responseType, defaultTimeout);
    }

    /**
     * Performs a round-trip POST request with custom timeout.
     */
    public <T> CompletableFuture<RoundTripResponse<T>> roundTripPost(
            String targetUrl, 
            Object payload, 
            String roleArn,
            String tailscaleAuthKey,
            Class<T> responseType,
            Duration timeout) {
        
        String requestId = generateRequestId();
        logger.info("Starting round-trip POST to {} with request ID: {}", targetUrl, requestId);

        // Create the request
        LambdaRequest request = new LambdaRequest(callbackUrl, payload, requestId, roleArn);
        if (tailscaleAuthKey != null && !tailscaleAuthKey.isEmpty()) {
            request.setTailnetKey(tailscaleAuthKey);
        }

        CompletableFuture<RoundTripResponse<T>> future = new CompletableFuture<>();

        try {
            // Serialize request
            String jsonBody = objectMapper.writeValueAsString(request);
            RequestBody body = RequestBody.create(jsonBody, JSON);
            
            // Build HTTP request
            Request httpRequest = new Request.Builder()
                .url(targetUrl)
                .post(body)
                .addHeader("Content-Type", "application/json")
                .addHeader("User-Agent", "post2post-java-client/1.0")
                .build();

            // Execute request
            httpClient.newCall(httpRequest).enqueue(new Callback() {
                @Override
                public void onFailure(Call call, IOException e) {
                    logger.error("Round-trip POST failed for request {}: {}", requestId, e.getMessage());
                    future.complete(new RoundTripResponse<>(
                        null, false, "HTTP request failed: " + e.getMessage(), false, requestId));
                }

                @Override
                public void onResponse(Call call, Response response) throws IOException {
                    try (ResponseBody responseBody = response.body()) {
                        if (!response.isSuccessful()) {
                            String error = String.format("HTTP %d: %s", response.code(), response.message());
                            logger.error("Round-trip POST failed for request {}: {}", requestId, error);
                            future.complete(new RoundTripResponse<>(
                                null, false, error, false, requestId));
                            return;
                        }

                        // For now, we'll parse the immediate acknowledgment
                        // In a full implementation, this would set up a callback server
                        // to receive the async response
                        logger.info("Round-trip POST acknowledged for request {}", requestId);
                        
                        // TODO: Implement callback server for receiving async responses
                        // For now, return success with acknowledgment
                        future.complete(new RoundTripResponse<>(
                            null, true, null, false, requestId));
                    } catch (Exception e) {
                        logger.error("Error processing response for request {}: {}", requestId, e.getMessage());
                        future.complete(new RoundTripResponse<>(
                            null, false, "Response processing failed: " + e.getMessage(), false, requestId));
                    }
                }
            });

            // Set up timeout
            CompletableFuture.delayedExecutor(timeout.toMillis(), TimeUnit.MILLISECONDS)
                .execute(() -> {
                    if (!future.isDone()) {
                        logger.warn("Round-trip POST timed out for request {}", requestId);
                        future.complete(new RoundTripResponse<>(
                            null, false, "Request timed out", true, requestId));
                    }
                });

        } catch (Exception e) {
            logger.error("Failed to initiate round-trip POST for request {}: {}", requestId, e.getMessage());
            future.complete(new RoundTripResponse<>(
                null, false, "Failed to initiate request: " + e.getMessage(), false, requestId));
        }

        return future;
    }

    /**
     * Performs a simple POST request without expecting a response.
     */
    public CompletableFuture<Boolean> post(String targetUrl, Object payload) {
        CompletableFuture<Boolean> future = new CompletableFuture<>();

        try {
            String jsonBody = objectMapper.writeValueAsString(payload);
            RequestBody body = RequestBody.create(jsonBody, JSON);
            
            Request request = new Request.Builder()
                .url(targetUrl)
                .post(body)
                .addHeader("Content-Type", "application/json")
                .addHeader("User-Agent", "post2post-java-client/1.0")
                .build();

            httpClient.newCall(request).enqueue(new Callback() {
                @Override
                public void onFailure(Call call, IOException e) {
                    logger.error("POST request failed to {}: {}", targetUrl, e.getMessage());
                    future.complete(false);
                }

                @Override
                public void onResponse(Call call, Response response) throws IOException {
                    try (ResponseBody responseBody = response.body()) {
                        boolean success = response.isSuccessful();
                        if (success) {
                            logger.info("POST request successful to {}", targetUrl);
                        } else {
                            logger.error("POST request failed to {} with status: {} {}", 
                                targetUrl, response.code(), response.message());
                        }
                        future.complete(success);
                    }
                }
            });

        } catch (Exception e) {
            logger.error("Failed to initiate POST to {}: {}", targetUrl, e.getMessage());
            future.complete(false);
        }

        return future;
    }

    /**
     * Generates a unique request ID.
     */
    private String generateRequestId() {
        return "req_" + System.currentTimeMillis() + "_" + requestCounter.incrementAndGet();
    }

    /**
     * Closes the HTTP client and releases resources.
     */
    public void close() {
        httpClient.dispatcher().executorService().shutdown();
        httpClient.connectionPool().evictAll();
        logger.info("Post2Post client closed");
    }
}