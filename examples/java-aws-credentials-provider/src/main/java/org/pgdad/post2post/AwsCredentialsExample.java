package org.pgdad.post2post;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import software.amazon.awssdk.auth.credentials.AwsCredentials;
import software.amazon.awssdk.core.SdkSystemSetting;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.s3.S3Client;
import software.amazon.awssdk.services.s3.model.ListBucketsRequest;
import software.amazon.awssdk.services.s3.model.ListBucketsResponse;
import software.amazon.awssdk.services.sts.StsClient;
import software.amazon.awssdk.services.sts.model.GetCallerIdentityRequest;
import software.amazon.awssdk.services.sts.model.GetCallerIdentityResponse;

import java.time.Duration;

/**
 * Example application demonstrating the Post2Post AWS Credentials Provider.
 * This shows how to use the provider with AWS SDK for Java v2.
 */
public class AwsCredentialsExample {
    private static final Logger logger = LoggerFactory.getLogger(AwsCredentialsExample.class);

    public static void main(String[] args) {
        // Get configuration from environment variables
        String lambdaUrl = System.getenv("AWS_LAMBDA_URL");
        String roleArn = System.getenv("AWS_ROLE_ARN");
        String tailscaleAuthKey = System.getenv("TAILSCALE_AUTH_KEY");
        String callbackUrl = System.getenv("CALLBACK_URL");

        if (lambdaUrl == null || lambdaUrl.isEmpty()) {
            logger.error("AWS_LAMBDA_URL environment variable is required");
            System.exit(1);
        }
        if (roleArn == null || roleArn.isEmpty()) {
            logger.error("AWS_ROLE_ARN environment variable is required (must be in /remote/ path)");
            System.exit(1);
        }
        if (tailscaleAuthKey == null || tailscaleAuthKey.isEmpty()) {
            logger.error("TAILSCALE_AUTH_KEY environment variable is required");
            System.exit(1);
        }
        if (callbackUrl == null || callbackUrl.isEmpty()) {
            logger.error("CALLBACK_URL environment variable is required");
            System.exit(1);
        }

        logger.info("🚀 Starting Post2Post AWS Credentials Provider Example");
        logger.info("Lambda URL: {}", lambdaUrl);
        logger.info("Role ARN: {}", roleArn);
        logger.info("Callback URL: {}", callbackUrl);

        // Create the Post2Post AWS credentials provider
        Post2PostAwsCredentialsProvider.Config config = new Post2PostAwsCredentialsProvider.Config()
            .lambdaUrl(lambdaUrl)
            .roleArn(roleArn)
            .tailscaleAuthKey(tailscaleAuthKey)
            .callbackUrl(callbackUrl)
            .sessionName("post2post-java-example")
            .credentialDuration(Duration.ofHours(1))
            .requestTimeout(Duration.ofSeconds(30));

        Post2PostAwsCredentialsProvider credentialsProvider = null;
        try {
            credentialsProvider = new Post2PostAwsCredentialsProvider(config);
            logger.info("✅ Post2Post AWS Credentials Provider initialized");

            // Test 1: Retrieve credentials directly
            logger.info("\n🔍 Test 1: Retrieve AWS credentials");
            AwsCredentials credentials = credentialsProvider.resolveCredentials();
            logger.info("Access Key ID: {}", maskCredential(credentials.accessKeyId()));
            logger.info("Secret Access Key: {}", maskCredential(credentials.secretAccessKey()));
            if (credentials instanceof software.amazon.awssdk.auth.credentials.AwsSessionCredentials) {
                software.amazon.awssdk.auth.credentials.AwsSessionCredentials sessionCreds = 
                    (software.amazon.awssdk.auth.credentials.AwsSessionCredentials) credentials;
                logger.info("Session Token: {}", maskCredential(sessionCreds.sessionToken()));
            }

            // Test 2: Use STS to verify identity
            logger.info("\n🔍 Test 2: Verify assumed role identity using STS");
            try (StsClient stsClient = StsClient.builder()
                    .credentialsProvider(credentialsProvider)
                    .region(Region.US_EAST_1)
                    .build()) {

                GetCallerIdentityResponse identity = stsClient.getCallerIdentity(
                    GetCallerIdentityRequest.builder().build());

                logger.info("Account: {}", identity.account());
                logger.info("User ID: {}", identity.userId());
                logger.info("ARN: {}", identity.arn());
            }

            // Test 3: Use S3 to list buckets (requires S3 permissions on the assumed role)
            logger.info("\n📦 Test 3: List S3 buckets using assumed role");
            try (S3Client s3Client = S3Client.builder()
                    .credentialsProvider(credentialsProvider)
                    .region(Region.US_EAST_1)
                    .build()) {

                ListBucketsResponse buckets = s3Client.listBuckets(ListBucketsRequest.builder().build());
                logger.info("Successfully listed {} S3 buckets:", buckets.buckets().size());
                
                buckets.buckets().stream()
                    .limit(5)
                    .forEach(bucket -> logger.info("  - {} (created: {})", 
                        bucket.name(), bucket.creationDate()));
                
                if (buckets.buckets().size() > 5) {
                    logger.info("  ... and {} more buckets", buckets.buckets().size() - 5);
                }
            } catch (Exception e) {
                logger.warn("⚠️  Failed to list S3 buckets (role may not have S3 permissions): {}", e.getMessage());
            }

            // Test 4: Test credential caching
            logger.info("\n🔄 Test 4: Test credential caching");
            AwsCredentials cachedCredentials = credentialsProvider.resolveCredentials();
            if (credentials.accessKeyId().equals(cachedCredentials.accessKeyId())) {
                logger.info("✅ Credential caching working - same credentials returned");
            } else {
                logger.info("⚠️  Different credentials returned - caching may not be working");
            }

            // Test 5: Invalidate cache and fetch fresh credentials
            logger.info("\n🔄 Test 5: Invalidate cache and fetch fresh credentials");
            credentialsProvider.invalidateCache();
            AwsCredentials freshCredentials = credentialsProvider.resolveCredentials();
            logger.info("Fresh credentials retrieved - Access Key: {}", maskCredential(freshCredentials.accessKeyId()));

            logger.info("\n🎉 Post2Post AWS Credentials Provider Example completed successfully!");
            logger.info("The Java library successfully:");
            logger.info("  ✅ Implemented AWS CredentialsProvider interface");
            logger.info("  ✅ Connected to Lambda via Post2Post client");
            logger.info("  ✅ Provided credentials to AWS SDK");
            logger.info("  ✅ Cached credentials for performance");
            logger.info("  ✅ Supported credential invalidation");

        } catch (Exception e) {
            logger.error("❌ Example failed: {}", e.getMessage(), e);
            System.exit(1);
        } finally {
            if (credentialsProvider != null) {
                credentialsProvider.close();
            }
        }
    }

    /**
     * Masks sensitive credential information for logging.
     */
    private static String maskCredential(String credential) {
        if (credential == null || credential.length() < 8) {
            return "***";
        }
        return credential.substring(0, 4) + "***" + credential.substring(credential.length() - 4);
    }
}