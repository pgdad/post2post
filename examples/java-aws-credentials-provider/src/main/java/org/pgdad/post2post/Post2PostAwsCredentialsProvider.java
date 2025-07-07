package org.pgdad.post2post;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import software.amazon.awssdk.auth.credentials.AwsCredentials;
import software.amazon.awssdk.auth.credentials.AwsCredentialsProvider;
import software.amazon.awssdk.auth.credentials.AwsSessionCredentials;

import java.time.Duration;
import java.time.Instant;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.locks.ReadWriteLock;
import java.util.concurrent.locks.ReentrantReadWriteLock;

/**
 * AWS Credentials Provider that uses Post2Post to retrieve credentials from a Lambda function.
 * This provider communicates with an AWS Lambda through Tailscale mesh networking to assume
 * IAM roles and return temporary credentials.
 */
public class Post2PostAwsCredentialsProvider implements AwsCredentialsProvider {
    private static final Logger logger = LoggerFactory.getLogger(Post2PostAwsCredentialsProvider.class);
    private static final Duration EXPIRY_BUFFER = Duration.ofMinutes(5);
    
    private final Post2PostClient client;
    private final String lambdaUrl;
    private final String roleArn;
    private final String tailscaleAuthKey;
    private final String sessionName;
    private final Duration credentialDuration;
    private final ObjectMapper objectMapper;
    
    // Credential caching
    private final ReadWriteLock cacheLock = new ReentrantReadWriteLock();
    private volatile AwsCredentials cachedCredentials;
    private volatile Instant cacheExpiry;

    /**
     * Configuration for the Post2Post AWS Credentials Provider.
     */
    public static class Config {
        private String lambdaUrl;
        private String roleArn;
        private String tailscaleAuthKey;
        private String sessionName = "post2post-java-session";
        private Duration credentialDuration = Duration.ofHours(1);
        private Duration requestTimeout = Duration.ofSeconds(30);
        private String callbackUrl;

        public Config lambdaUrl(String lambdaUrl) {
            this.lambdaUrl = lambdaUrl;
            return this;
        }

        public Config roleArn(String roleArn) {
            this.roleArn = roleArn;
            return this;
        }

        public Config tailscaleAuthKey(String tailscaleAuthKey) {
            this.tailscaleAuthKey = tailscaleAuthKey;
            return this;
        }

        public Config sessionName(String sessionName) {
            this.sessionName = sessionName;
            return this;
        }

        public Config credentialDuration(Duration credentialDuration) {
            this.credentialDuration = credentialDuration;
            return this;
        }

        public Config requestTimeout(Duration requestTimeout) {
            this.requestTimeout = requestTimeout;
            return this;
        }

        public Config callbackUrl(String callbackUrl) {
            this.callbackUrl = callbackUrl;
            return this;
        }

        // Getters
        public String getLambdaUrl() { return lambdaUrl; }
        public String getRoleArn() { return roleArn; }
        public String getTailscaleAuthKey() { return tailscaleAuthKey; }
        public String getSessionName() { return sessionName; }
        public Duration getCredentialDuration() { return credentialDuration; }
        public Duration getRequestTimeout() { return requestTimeout; }
        public String getCallbackUrl() { return callbackUrl; }
    }

    /**
     * Lambda response payload structure.
     */
    public static class LambdaProcessedPayload {
        @JsonProperty("original_payload")
        private String originalPayload;
        
        @JsonProperty("assume_role_result")
        private AssumeRoleResult assumeRoleResult;
        
        @JsonProperty("processed_at")
        private String processedAt;
        
        @JsonProperty("processed_by")
        private String processedBy;
        
        @JsonProperty("lambda_request_id")
        private String lambdaRequestId;
        
        @JsonProperty("status")
        private String status;

        // Getters and setters
        public String getOriginalPayload() { return originalPayload; }
        public void setOriginalPayload(String originalPayload) { this.originalPayload = originalPayload; }
        
        public AssumeRoleResult getAssumeRoleResult() { return assumeRoleResult; }
        public void setAssumeRoleResult(AssumeRoleResult assumeRoleResult) { this.assumeRoleResult = assumeRoleResult; }
        
        public String getProcessedAt() { return processedAt; }
        public void setProcessedAt(String processedAt) { this.processedAt = processedAt; }
        
        public String getProcessedBy() { return processedBy; }
        public void setProcessedBy(String processedBy) { this.processedBy = processedBy; }
        
        public String getLambdaRequestId() { return lambdaRequestId; }
        public void setLambdaRequestId(String lambdaRequestId) { this.lambdaRequestId = lambdaRequestId; }
        
        public String getStatus() { return status; }
        public void setStatus(String status) { this.status = status; }
    }

    /**
     * STS AssumeRole result structure.
     */
    public static class AssumeRoleResult {
        @JsonProperty("credentials")
        private StsCredentials credentials;
        
        @JsonProperty("assumed_role_user")
        private AssumedRoleUser assumedRoleUser;

        // Getters and setters
        public StsCredentials getCredentials() { return credentials; }
        public void setCredentials(StsCredentials credentials) { this.credentials = credentials; }
        
        public AssumedRoleUser getAssumedRoleUser() { return assumedRoleUser; }
        public void setAssumedRoleUser(AssumedRoleUser assumedRoleUser) { this.assumedRoleUser = assumedRoleUser; }
    }

    /**
     * STS credentials structure.
     */
    public static class StsCredentials {
        @JsonProperty("AccessKeyId")
        private String accessKeyId;
        
        @JsonProperty("SecretAccessKey")
        private String secretAccessKey;
        
        @JsonProperty("SessionToken")
        private String sessionToken;
        
        @JsonProperty("Expiration")
        private String expiration;

        // Getters and setters
        public String getAccessKeyId() { return accessKeyId; }
        public void setAccessKeyId(String accessKeyId) { this.accessKeyId = accessKeyId; }
        
        public String getSecretAccessKey() { return secretAccessKey; }
        public void setSecretAccessKey(String secretAccessKey) { this.secretAccessKey = secretAccessKey; }
        
        public String getSessionToken() { return sessionToken; }
        public void setSessionToken(String sessionToken) { this.sessionToken = sessionToken; }
        
        public String getExpiration() { return expiration; }
        public void setExpiration(String expiration) { this.expiration = expiration; }
    }

    /**
     * Assumed role user structure.
     */
    public static class AssumedRoleUser {
        @JsonProperty("Arn")
        private String arn;
        
        @JsonProperty("AssumedRoleId")
        private String assumedRoleId;

        // Getters and setters
        public String getArn() { return arn; }
        public void setArn(String arn) { this.arn = arn; }
        
        public String getAssumedRoleId() { return assumedRoleId; }
        public void setAssumedRoleId(String assumedRoleId) { this.assumedRoleId = assumedRoleId; }
    }

    /**
     * Creates a new Post2Post AWS Credentials Provider.
     */
    public Post2PostAwsCredentialsProvider(Config config) {
        if (config.getLambdaUrl() == null || config.getLambdaUrl().isEmpty()) {
            throw new IllegalArgumentException("Lambda URL is required");
        }
        if (config.getRoleArn() == null || config.getRoleArn().isEmpty()) {
            throw new IllegalArgumentException("Role ARN is required");
        }
        if (config.getTailscaleAuthKey() == null || config.getTailscaleAuthKey().isEmpty()) {
            throw new IllegalArgumentException("Tailscale auth key is required for secure communication");
        }
        if (config.getCallbackUrl() == null || config.getCallbackUrl().isEmpty()) {
            throw new IllegalArgumentException("Callback URL is required");
        }

        this.lambdaUrl = config.getLambdaUrl();
        this.roleArn = config.getRoleArn();
        this.tailscaleAuthKey = config.getTailscaleAuthKey();
        this.sessionName = config.getSessionName();
        this.credentialDuration = config.getCredentialDuration();

        // Initialize Post2Post client
        Post2PostClient.Config clientConfig = new Post2PostClient.Config()
            .callbackUrl(config.getCallbackUrl())
            .timeout(config.getRequestTimeout())
            .enableTailscale(config.getTailscaleAuthKey());

        this.client = new Post2PostClient(clientConfig);
        
        // Initialize JSON mapper
        this.objectMapper = new ObjectMapper();
        this.objectMapper.registerModule(new JavaTimeModule());

        logger.info("Post2Post AWS Credentials Provider initialized");
        logger.info("Lambda URL: {}", lambdaUrl);
        logger.info("Role ARN: {}", roleArn);
        logger.info("Session name: {}", sessionName);
    }

    @Override
    public AwsCredentials resolveCredentials() {
        // Check cached credentials first
        cacheLock.readLock().lock();
        try {
            if (cachedCredentials != null && Instant.now().isBefore(cacheExpiry)) {
                logger.debug("Using cached AWS credentials (expires: {})", cacheExpiry);
                return cachedCredentials;
            }
        } finally {
            cacheLock.readLock().unlock();
        }

        // Need to fetch new credentials
        logger.info("Fetching new AWS credentials from Lambda: {}", lambdaUrl);
        
        try {
            // Create payload for the Lambda request
            String payload = String.format("assume-role-request-%s-%d", 
                sessionName, System.currentTimeMillis());

            // Make round-trip request to Lambda
            CompletableFuture<Post2PostClient.RoundTripResponse<LambdaProcessedPayload>> future = 
                client.roundTripPost(lambdaUrl, payload, roleArn, tailscaleAuthKey, LambdaProcessedPayload.class);

            // Wait for response (this is a blocking operation)
            Post2PostClient.RoundTripResponse<LambdaProcessedPayload> response = 
                future.get(credentialDuration.toSeconds(), java.util.concurrent.TimeUnit.SECONDS);

            if (!response.isSuccess()) {
                String error = response.getError() != null ? response.getError() : "Unknown error";
                throw new RuntimeException("Failed to retrieve credentials from Lambda: " + error);
            }

            // For now, since we don't have full round-trip implementation,
            // we'll simulate a successful response for demonstration
            // TODO: Parse actual Lambda response when round-trip is fully implemented
            
            // This is a placeholder implementation that would work with actual responses
            logger.warn("Using placeholder credentials - full round-trip implementation needed");
            
            // Create placeholder credentials for demonstration
            AwsCredentials credentials = AwsSessionCredentials.create(
                "PLACEHOLDER_ACCESS_KEY",
                "PLACEHOLDER_SECRET_KEY", 
                "PLACEHOLDER_SESSION_TOKEN"
            );

            // Cache the credentials
            cacheLock.writeLock().lock();
            try {
                this.cachedCredentials = credentials;
                this.cacheExpiry = Instant.now().plus(credentialDuration).minus(EXPIRY_BUFFER);
                logger.info("AWS credentials cached until: {}", cacheExpiry);
            } finally {
                cacheLock.writeLock().unlock();
            }

            return credentials;

        } catch (Exception e) {
            logger.error("Failed to retrieve AWS credentials", e);
            throw new RuntimeException("Failed to retrieve AWS credentials: " + e.getMessage(), e);
        }
    }

    /**
     * Invalidates the credential cache, forcing a fresh fetch on the next call.
     */
    public void invalidateCache() {
        cacheLock.writeLock().lock();
        try {
            this.cachedCredentials = null;
            this.cacheExpiry = null;
            logger.info("AWS credentials cache invalidated");
        } finally {
            cacheLock.writeLock().unlock();
        }
    }

    /**
     * Gets the configured role ARN.
     */
    public String getRoleArn() {
        return roleArn;
    }

    /**
     * Gets the configured session name.
     */
    public String getSessionName() {
        return sessionName;
    }

    /**
     * Gets the configured Lambda URL.
     */
    public String getLambdaUrl() {
        return lambdaUrl;
    }

    /**
     * Closes the provider and releases resources.
     */
    public void close() {
        if (client != null) {
            client.close();
        }
        logger.info("Post2Post AWS Credentials Provider closed");
    }
}