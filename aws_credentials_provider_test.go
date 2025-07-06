package post2post

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func TestAWSCredentialsProvider_NewProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      AWSCredentialsProviderConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: AWSCredentialsProviderConfig{
				LambdaURL:  "https://lambda.example.com",
				RoleARN:    "arn:aws:iam::123456789012:role/remote/TestRole",
				TailnetKey: "tskey-auth-test123",
			},
			expectError: false,
		},
		{
			name: "missing lambda URL",
			config: AWSCredentialsProviderConfig{
				RoleARN:    "arn:aws:iam::123456789012:role/remote/TestRole",
				TailnetKey: "tskey-auth-test123",
			},
			expectError: true,
		},
		{
			name: "missing role ARN",
			config: AWSCredentialsProviderConfig{
				LambdaURL:  "https://lambda.example.com",
				TailnetKey: "tskey-auth-test123",
			},
			expectError: true,
		},
		{
			name: "missing tailnet key",
			config: AWSCredentialsProviderConfig{
				LambdaURL: "https://lambda.example.com",
				RoleARN:   "arn:aws:iam::123456789012:role/remote/TestRole",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewAWSCredentialsProvider(tt.config)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if provider == nil {
				t.Errorf("expected provider but got nil")
				return
			}
			
			// Verify configuration
			if provider.GetLambdaURL() != tt.config.LambdaURL {
				t.Errorf("expected Lambda URL %s, got %s", tt.config.LambdaURL, provider.GetLambdaURL())
			}
			
			if provider.GetRoleARN() != tt.config.RoleARN {
				t.Errorf("expected Role ARN %s, got %s", tt.config.RoleARN, provider.GetRoleARN())
			}
			
			// Clean up
			provider.Close()
		})
	}
}

func TestAWSCredentialsProvider_Retrieve(t *testing.T) {
	t.Skip("Skipping integration test - requires full Lambda setup")
	// This test would require a complete mock of the Lambda response format
	// For now, we'll focus on unit tests for the configuration and caching logic
}

func TestAWSCredentialsProvider_InvalidateCache(t *testing.T) {
	config := AWSCredentialsProviderConfig{
		LambdaURL:  "https://lambda.example.com",
		RoleARN:    "arn:aws:iam::123456789012:role/remote/TestRole",
		TailnetKey: "tskey-auth-test123",
	}

	provider, err := NewAWSCredentialsProvider(config)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	defer provider.Close()

	// Set some mock cached credentials
	provider.mu.Lock()
	provider.credentials = &aws.Credentials{
		AccessKeyID:     "AKIATEST123456789",
		SecretAccessKey: "secretkey123456789",
		SessionToken:    "sessiontoken123456789",
	}
	provider.expiry = time.Now().Add(1 * time.Hour)
	provider.mu.Unlock()

	// Invalidate cache
	provider.InvalidateCache()

	// Verify cache is cleared
	provider.mu.RLock()
	if provider.credentials != nil {
		t.Errorf("expected credentials to be nil after invalidation")
	}
	if !provider.expiry.IsZero() {
		t.Errorf("expected expiry to be zero after invalidation")
	}
	provider.mu.RUnlock()
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}