package post2post

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
)

// AWSCredentialsProvider implements aws.CredentialsProvider using post2post
type AWSCredentialsProvider struct {
	server      *Server
	lambdaURL   string
	roleARN     string
	tailnetKey  string
	sessionName string
	duration    time.Duration
	
	// Cached credentials
	mu          sync.RWMutex
	credentials *aws.Credentials
	expiry      time.Time
}

// AWSCredentialsProviderConfig holds configuration for the AWS credentials provider
type AWSCredentialsProviderConfig struct {
	LambdaURL   string        // Lambda Function URL endpoint
	RoleARN     string        // IAM Role ARN to assume (must be in /remote/ path)
	TailnetKey  string        // Tailscale auth key for secure communication
	SessionName string        // Session name for the assumed role (optional)
	Duration    time.Duration // Credential duration (optional, default 1 hour)
}

// LambdaAssumeRoleRequest represents the request sent to the Lambda function
type LambdaAssumeRoleRequest struct {
	URL        string `json:"url"`
	Payload    string `json:"payload"`
	RequestID  string `json:"request_id"`
	TailnetKey string `json:"tailnet_key,omitempty"`
	RoleARN    string `json:"role_arn"`
}

// LambdaAssumeRoleResponse represents the response from the Lambda function
type LambdaAssumeRoleResponse struct {
	RequestID string                 `json:"request_id"`
	Payload   LambdaProcessedPayload `json:"payload"`
	TailnetKey string                `json:"tailnet_key,omitempty"`
}

// LambdaProcessedPayload represents the processed payload from the Lambda
type LambdaProcessedPayload struct {
	OriginalPayload  string                   `json:"original_payload"`
	AssumeRoleResult LambdaAssumeRoleResult   `json:"assume_role_result"`
	ProcessedAt      string                   `json:"processed_at"`
	ProcessedBy      string                   `json:"processed_by"`
	LambdaRequestID  string                   `json:"lambda_request_id"`
	Status           string                   `json:"status"`
}

// LambdaAssumeRoleResult represents the STS AssumeRole result from Lambda
type LambdaAssumeRoleResult struct {
	Credentials      *types.Credentials      `json:"credentials"`
	AssumedRoleUser  *types.AssumedRoleUser  `json:"assumed_role_user"`
	PackedPolicySize *int32                  `json:"packed_policy_size,omitempty"`
	SourceIdentity   *string                 `json:"source_identity,omitempty"`
}

// NewAWSCredentialsProvider creates a new AWS credentials provider using post2post
func NewAWSCredentialsProvider(config AWSCredentialsProviderConfig) (*AWSCredentialsProvider, error) {
	if config.LambdaURL == "" {
		return nil, fmt.Errorf("lambda URL is required")
	}
	if config.RoleARN == "" {
		return nil, fmt.Errorf("role ARN is required")
	}
	if config.TailnetKey == "" {
		return nil, fmt.Errorf("tailnet key is required for secure communication")
	}

	// Set defaults
	if config.SessionName == "" {
		config.SessionName = "post2post-credentials-provider"
	}
	if config.Duration == 0 {
		config.Duration = 1 * time.Hour
	}

	// Create a post2post server for handling responses
	server := NewServer().WithPostURL(config.LambdaURL)
	
	// Start the server on an available port
	if err := server.Start(); err != nil {
		return nil, fmt.Errorf("failed to start post2post server: %w", err)
	}

	provider := &AWSCredentialsProvider{
		server:      server,
		lambdaURL:   config.LambdaURL,
		roleARN:     config.RoleARN,
		tailnetKey:  config.TailnetKey,
		sessionName: config.SessionName,
		duration:    config.Duration,
	}

	log.Printf("AWS Credentials Provider initialized with Lambda URL: %s", config.LambdaURL)
	log.Printf("Will assume role: %s", config.RoleARN)

	return provider, nil
}

// Retrieve implements aws.CredentialsProvider.Retrieve
func (p *AWSCredentialsProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	p.mu.RLock()
	// Check if we have valid cached credentials
	if p.credentials != nil && time.Now().Before(p.expiry) {
		creds := *p.credentials
		p.mu.RUnlock()
		log.Printf("Using cached AWS credentials (expires: %s)", p.expiry.Format(time.RFC3339))
		return creds, nil
	}
	p.mu.RUnlock()

	// Need to fetch new credentials
	log.Printf("Fetching new AWS credentials from Lambda: %s", p.lambdaURL)
	
	// Generate a unique request ID
	requestID := fmt.Sprintf("creds-%d", time.Now().UnixNano())
	
	// Get the appropriate URL for the callback
	var callbackURL string
	if p.tailnetKey != "" {
		// Use Tailscale hostname when Tailnet key is available
		tailscaleURL, err := p.server.GetTailscaleURL()
		if err != nil {
			log.Printf("Failed to get Tailscale URL, falling back to localhost: %v", err)
			callbackURL = p.server.GetURL() + "/roundtrip"
		} else {
			callbackURL = tailscaleURL + "/roundtrip"
			log.Printf("Using Tailscale callback URL: %s", callbackURL)
		}
	} else {
		callbackURL = p.server.GetURL() + "/roundtrip"
	}

	// Prepare the request payload
	request := LambdaAssumeRoleRequest{
		URL:        callbackURL,
		Payload:    fmt.Sprintf("assume-role-request-%s", requestID),
		RequestID:  requestID,
		TailnetKey: p.tailnetKey,
		RoleARN:    p.roleARN,
	}

	// Use RoundTripPost to get the response synchronously
	response, err := p.server.RoundTripPostWithTimeout(request, p.tailnetKey, 30*time.Second)
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("failed to retrieve credentials from Lambda: %w", err)
	}

	// Parse the response
	var lambdaResponse LambdaAssumeRoleResponse
	responseBytes, err := json.Marshal(response.Payload)
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("failed to marshal response payload: %w", err)
	}

	if err := json.Unmarshal(responseBytes, &lambdaResponse); err != nil {
		return aws.Credentials{}, fmt.Errorf("failed to parse Lambda response: %w", err)
	}

	// Check if the request was successful
	if lambdaResponse.Payload.Status != "success" {
		return aws.Credentials{}, fmt.Errorf("Lambda returned error status: %s", lambdaResponse.Payload.Status)
	}

	// Extract credentials from the response
	stsCredentials := lambdaResponse.Payload.AssumeRoleResult.Credentials
	if stsCredentials == nil {
		return aws.Credentials{}, fmt.Errorf("no credentials returned in Lambda response")
	}

	// Convert to aws.Credentials
	credentials := aws.Credentials{
		AccessKeyID:     *stsCredentials.AccessKeyId,
		SecretAccessKey: *stsCredentials.SecretAccessKey,
		SessionToken:    *stsCredentials.SessionToken,
		Source:          "Post2PostAWSCredentialsProvider",
		CanExpire:       true,
		Expires:         *stsCredentials.Expiration,
	}

	// Cache the credentials with a buffer before expiry
	expiryBuffer := 5 * time.Minute
	p.mu.Lock()
	p.credentials = &credentials
	p.expiry = credentials.Expires.Add(-expiryBuffer)
	p.mu.Unlock()

	log.Printf("Successfully retrieved AWS credentials (expires: %s)", credentials.Expires.Format(time.RFC3339))
	log.Printf("Assumed role user: %s", *lambdaResponse.Payload.AssumeRoleResult.AssumedRoleUser.Arn)

	return credentials, nil
}

// Close stops the internal post2post server
func (p *AWSCredentialsProvider) Close() error {
	if p.server != nil {
		return p.server.Stop()
	}
	return nil
}

// GetRoleARN returns the configured role ARN
func (p *AWSCredentialsProvider) GetRoleARN() string {
	return p.roleARN
}

// GetSessionName returns the configured session name
func (p *AWSCredentialsProvider) GetSessionName() string {
	return p.sessionName
}

// GetLambdaURL returns the configured Lambda URL
func (p *AWSCredentialsProvider) GetLambdaURL() string {
	return p.lambdaURL
}

// InvalidateCache forces the provider to fetch new credentials on the next Retrieve call
func (p *AWSCredentialsProvider) InvalidateCache() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.credentials = nil
	p.expiry = time.Time{}
	log.Printf("AWS credentials cache invalidated")
}