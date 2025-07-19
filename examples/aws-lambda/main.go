package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"tailscale.com/tsnet"
)

// LambdaRequest represents the incoming request payload
type LambdaRequest struct {
	URL        string      `json:"url"`
	Payload    interface{} `json:"payload"`
	RequestID  string      `json:"request_id"`
	TailnetKey string      `json:"tailnet_key,omitempty"`
	RoleARN    string      `json:"role_arn"`
}

// AssumeRoleResponse represents the response from AWS STS AssumeRole
type AssumeRoleResponse struct {
	Credentials    *types.Credentials `json:"credentials"`
	AssumedRoleUser *types.AssumedRoleUser `json:"assumed_role_user"`
	PackedPolicySize *int32 `json:"packed_policy_size,omitempty"`
	SourceIdentity   *string `json:"source_identity,omitempty"`
}

// ProcessedResponse represents the final response payload
type ProcessedResponse struct {
	OriginalPayload  interface{}        `json:"original_payload"`
	AssumeRoleResult AssumeRoleResponse `json:"assume_role_result"`
	ProcessedAt      string            `json:"processed_at"`
	ProcessedBy      string            `json:"processed_by"`
	LambdaRequestID  string            `json:"lambda_request_id"`
	Status           string            `json:"status"`
}

// LambdaResponse represents the response sent back to the callback URL
type LambdaResponse struct {
	RequestID string      `json:"request_id"`
	Payload   interface{} `json:"payload"`
	TailnetKey string     `json:"tailnet_key,omitempty"`
}

// Global AWS configuration
var awsConfig aws.Config
var stsClient *sts.Client
var allowedTailnetDomain string

func init() {
	// Initialize AWS configuration
	var err error
	awsConfig, err = config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	
	stsClient = sts.NewFromConfig(awsConfig)
	
	// Get required Tailscale domain configuration
	allowedTailnetDomain = os.Getenv("TAILNET_DOMAIN")
	if allowedTailnetDomain == "" {
		log.Fatalf("TAILNET_DOMAIN environment variable is required but not set")
	}
	
	log.Printf("AWS Lambda post2post receiver initialized with Tailnet domain: %s", allowedTailnetDomain)
}

// handleRequest processes the Lambda URL request
func handleRequest(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	log.Printf("Received request: %s %s", request.RequestContext.HTTP.Method, request.RawPath)
	
	// Only handle POST requests
	if request.RequestContext.HTTP.Method != "POST" {
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusMethodNotAllowed,
			Body:       `{"error": "Method not allowed"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}
	
	// Parse the incoming request
	var lambdaReq LambdaRequest
	if err := json.Unmarshal([]byte(request.Body), &lambdaReq); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusBadRequest,
			Body:       `{"error": "Invalid JSON payload"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}
	
	log.Printf("Processing request ID: %s", lambdaReq.RequestID)
	log.Printf("Role ARN to assume: %s", lambdaReq.RoleARN)
	if lambdaReq.TailnetKey != "" {
		log.Printf("Tailscale integration enabled with key: %s...", lambdaReq.TailnetKey[:min(len(lambdaReq.TailnetKey), 10)])
	}
	
	// Validate required fields
	if lambdaReq.RoleARN == "" {
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusBadRequest,
			Body:       `{"error": "role_arn is required"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}
	
	if lambdaReq.URL == "" {
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusBadRequest,
			Body:       `{"error": "callback url is required"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}
	
	// Validate callback URL domain against configured Tailnet domain
	if err := validateCallbackURL(lambdaReq.URL); err != nil {
		log.Printf("Invalid callback URL %s: %v", lambdaReq.URL, err)
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusForbidden,
			Body:       fmt.Sprintf(`{"error": "Invalid callback URL: %s"}`, err.Error()),
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}
	
	// Process the request asynchronously
	go func() {
		processRequest(ctx, lambdaReq, request.RequestContext.RequestID)
	}()
	
	// Return immediate acknowledgment
	return events.LambdaFunctionURLResponse{
		StatusCode: http.StatusOK,
		Body:       `{"status": "accepted", "message": "Processing request"}`,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

// processRequest handles the actual processing and response posting
func processRequest(ctx context.Context, req LambdaRequest, lambdaRequestID string) {
	// Add processing delay to simulate work
	time.Sleep(100 * time.Millisecond)
	
	log.Printf("Starting role assumption for request: %s", req.RequestID)
	
	// Assume the specified IAM role
	assumeRoleResult, err := assumeRole(ctx, req.RoleARN, req.RequestID)
	if err != nil {
		log.Printf("Failed to assume role %s: %v", req.RoleARN, err)
		postErrorResponse(req, fmt.Sprintf("Failed to assume role: %v", err), lambdaRequestID)
		return
	}
	
	log.Printf("Successfully assumed role: %s", req.RoleARN)
	
	// Create the processed response
	processedResponse := ProcessedResponse{
		OriginalPayload:  req.Payload,
		AssumeRoleResult: *assumeRoleResult,
		ProcessedAt:      time.Now().Format("2006-01-02 15:04:05 MST"),
		ProcessedBy:      "aws-lambda-post2post-receiver",
		LambdaRequestID:  lambdaRequestID,
		Status:           "success",
	}
	
	// Create the response to send back
	response := LambdaResponse{
		RequestID:  req.RequestID,
		Payload:    processedResponse,
		TailnetKey: req.TailnetKey,
	}
	
	// Post the response back using Tailscale if specified
	if err := postResponse(req.URL, response, req.TailnetKey); err != nil {
		log.Printf("Failed to post response back to %s: %v", req.URL, err)
	} else {
		log.Printf("Successfully posted response back to %s", req.URL)
	}
}

// assumeRole performs AWS STS AssumeRole operation
func assumeRole(ctx context.Context, roleARN, sessionName string) (*AssumeRoleResponse, error) {
	// Create a unique session name
	fullSessionName := fmt.Sprintf("post2post-%s-%d", sessionName, time.Now().Unix())
	
	// Prepare the AssumeRole request
	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleARN),
		RoleSessionName: aws.String(fullSessionName),
		DurationSeconds: aws.Int32(3600), // 1 hour
	}
	
	// Execute the AssumeRole call
	result, err := stsClient.AssumeRole(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("STS AssumeRole failed: %w", err)
	}
	
	// Return the structured response
	return &AssumeRoleResponse{
		Credentials:      result.Credentials,
		AssumedRoleUser:  result.AssumedRoleUser,
		PackedPolicySize: result.PackedPolicySize,
		SourceIdentity:   result.SourceIdentity,
	}, nil
}

// postResponse posts the response back to the callback URL, optionally using Tailscale
func postResponse(callbackURL string, response LambdaResponse, tailnetKey string) error {
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	
	var client *http.Client
	
	if tailnetKey != "" {
		// Use Tailscale client for secure networking
		tailscaleClient, err := createTailscaleClient(tailnetKey)
		if err != nil {
			log.Printf("Failed to create Tailscale client, falling back to regular HTTP: %v", err)
			client = &http.Client{Timeout: 30 * time.Second}
		} else {
			client = tailscaleClient
			log.Println("Using Tailscale networking for response")
		}
	} else {
		// Use regular HTTP client
		client = &http.Client{Timeout: 30 * time.Second}
	}
	
	// Create and send the POST request
	req, err := http.NewRequest("POST", callbackURL, bytes.NewBuffer(responseJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "aws-lambda-post2post/1.0")
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post response: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("callback returned error status: %d", resp.StatusCode)
	}
	
	return nil
}

// createTailscaleClient creates an HTTP client that routes through Tailscale
func createTailscaleClient(tailnetKey string) (*http.Client, error) {
	// Create tsnet server for Tailscale integration
	srv := &tsnet.Server{
		Hostname:  "lambda-post2post-receiver",
		AuthKey:   tailnetKey,
		Ephemeral: true, // Lambda instances are ephemeral
	}
	
	// Start the tsnet server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Start(); err != nil {
		return nil, fmt.Errorf("failed to start tsnet server: %w", err)
	}
	
	// Wait for the server to be ready
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for Tailscale to start")
		default:
			status, err := srv.Up(ctx)
			if err != nil {
				log.Printf("Error checking Tailscale status: %v", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}
			if status != nil {
				log.Println("Tailscale tsnet server is ready")
				goto ready
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	
ready:
	// Create HTTP client that routes through Tailscale
	client := srv.HTTPClient()
	return client, nil
}

// validateCallbackURL validates that the callback URL domain matches the configured Tailnet domain
func validateCallbackURL(callbackURL string) error {
	parsedURL, err := url.Parse(callbackURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}
	
	// Extract the hostname from the URL
	hostname := parsedURL.Hostname()
	if hostname == "" {
		return fmt.Errorf("no hostname found in URL")
	}
	
	// Check if hostname ends with the allowed Tailnet domain
	if !strings.HasSuffix(hostname, allowedTailnetDomain) {
		return fmt.Errorf("hostname %s does not match allowed Tailnet domain %s", hostname, allowedTailnetDomain)
	}
	
	return nil
}

// postErrorResponse posts an error response back to the callback URL
func postErrorResponse(req LambdaRequest, errorMsg, lambdaRequestID string) {
	errorResponse := LambdaResponse{
		RequestID: req.RequestID,
		Payload: map[string]interface{}{
			"error":             errorMsg,
			"processed_at":      time.Now().Format("2006-01-02 15:04:05 MST"),
			"processed_by":      "aws-lambda-post2post-receiver",
			"lambda_request_id": lambdaRequestID,
			"status":            "error",
		},
		TailnetKey: req.TailnetKey,
	}
	
	if err := postResponse(req.URL, errorResponse, req.TailnetKey); err != nil {
		log.Printf("Failed to post error response: %v", err)
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	// Check if we're running in Lambda environment
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		log.Println("Starting AWS Lambda handler")
		lambda.Start(handleRequest)
	} else {
		log.Println("Not running in Lambda environment. Use 'go run main.go' for local testing.")
		log.Println("For Lambda deployment, build with: GOOS=linux GOARCH=amd64 go build -o bootstrap main.go")
	}
}