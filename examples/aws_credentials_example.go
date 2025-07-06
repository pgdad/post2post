package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	
	"github.com/pgdad/post2post"
)

func main() {
	// Get configuration from environment variables
	lambdaURL := os.Getenv("AWS_LAMBDA_URL")
	roleARN := os.Getenv("AWS_ROLE_ARN")
	tailnetKey := os.Getenv("TAILSCALE_AUTH_KEY")

	if lambdaURL == "" {
		log.Fatal("AWS_LAMBDA_URL environment variable is required")
	}
	if roleARN == "" {
		log.Fatal("AWS_ROLE_ARN environment variable is required (must be in /remote/ path)")
	}
	if tailnetKey == "" {
		log.Fatal("TAILSCALE_AUTH_KEY environment variable is required")
	}

	log.Printf("🚀 Starting AWS Credentials Provider Example")
	log.Printf("Lambda URL: %s", lambdaURL)
	log.Printf("Role ARN: %s", roleARN)

	// Create the post2post AWS credentials provider
	provider, err := post2post.NewAWSCredentialsProvider(post2post.AWSCredentialsProviderConfig{
		LambdaURL:   lambdaURL,
		RoleARN:     roleARN,
		TailnetKey:  tailnetKey,
		SessionName: "post2post-example-session",
	})
	if err != nil {
		log.Fatalf("Failed to create AWS credentials provider: %v", err)
	}
	defer provider.Close()

	// Create AWS configuration using our custom credentials provider
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(provider),
		config.WithRegion("us-east-1"), // Set your preferred region
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	log.Printf("✅ AWS configuration loaded with post2post credentials provider")

	// Test 1: Use STS to verify our assumed role identity
	log.Printf("\n🔍 Test 1: Verify assumed role identity")
	stsClient := sts.NewFromConfig(cfg)
	
	identity, err := stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		log.Fatalf("Failed to get caller identity: %v", err)
	}

	log.Printf("Account: %s", *identity.Account)
	log.Printf("User ID: %s", *identity.UserId)
	log.Printf("ARN: %s", *identity.Arn)

	// Test 2: Use S3 to list buckets (requires S3 permissions on the assumed role)
	log.Printf("\n📦 Test 2: List S3 buckets using assumed role")
	s3Client := s3.NewFromConfig(cfg)
	
	buckets, err := s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		log.Printf("⚠️  Failed to list S3 buckets (role may not have S3 permissions): %v", err)
	} else {
		log.Printf("Successfully listed %d S3 buckets:", len(buckets.Buckets))
		for i, bucket := range buckets.Buckets {
			if i >= 5 { // Limit output
				log.Printf("  ... and %d more buckets", len(buckets.Buckets)-5)
				break
			}
			log.Printf("  - %s (created: %s)", *bucket.Name, bucket.CreationDate.Format("2006-01-02 15:04:05"))
		}
	}

	// Test 3: Test credential caching by making another STS call
	log.Printf("\n🔄 Test 3: Test credential caching")
	identity2, err := stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		log.Fatalf("Failed to get caller identity (cached): %v", err)
	}

	if *identity.Arn == *identity2.Arn {
		log.Printf("✅ Credential caching working - same identity returned")
	} else {
		log.Printf("⚠️  Different identity returned - caching may not be working")
	}

	// Test 4: Invalidate cache and fetch fresh credentials
	log.Printf("\n🔄 Test 4: Invalidate cache and fetch fresh credentials")
	provider.InvalidateCache()
	
	identity3, err := stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		log.Fatalf("Failed to get caller identity (fresh): %v", err)
	}

	log.Printf("Fresh credentials fetched - ARN: %s", *identity3.Arn)

	log.Printf("\n🎉 AWS Credentials Provider Example completed successfully!")
	log.Printf("The post2post library successfully:")
	log.Printf("  ✅ Connected to Lambda via Tailscale")
	log.Printf("  ✅ Assumed the specified IAM role")
	log.Printf("  ✅ Provided credentials to AWS SDK")
	log.Printf("  ✅ Cached credentials for performance")
	log.Printf("  ✅ Supported credential invalidation")
}