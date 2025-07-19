package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	post2post "github.com/pgdad/post2post"
)

// CredentialsProcessOutput represents the JSON output format required by AWS CLI
// as documented at https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sourcing-external.html
type CredentialsProcessOutput struct {
	Version         int    `json:"Version"`
	AccessKeyId     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken,omitempty"`
	Expiration      string `json:"Expiration,omitempty"`
}

// CachedCredentials represents credentials stored in the cache file
type CachedCredentials struct {
	Credentials CredentialsProcessOutput `json:"credentials"`
	CachedAt    time.Time                `json:"cached_at"`
	ExpiresAt   time.Time                `json:"expires_at"`
	RoleARN     string                   `json:"role_arn"`
	LambdaURL   string                   `json:"lambda_url"`
}

// Config holds the configuration for the credentials process
type Config struct {
	LambdaURL   string
	RoleARN     string
	TailnetKey  string
	SessionName string
	Duration    time.Duration
	Timeout     time.Duration
}

func main() {
	// Configure logging to stderr (AWS credentials_process requirement)
	log.SetOutput(os.Stderr)
	log.SetPrefix("post2post-credentials: ")

	// Parse command line arguments
	config, err := parseFlags()
	if err != nil {
		log.Printf("Configuration error: %v", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		log.Printf("Invalid configuration: %v", err)
		os.Exit(1)
	}

	// Try to load cached credentials first
	var output *CredentialsProcessOutput
	cachedOutput, err := loadCachedCredentials(config)
	if err != nil {
		log.Printf("Warning: failed to load cached credentials: %v", err)
	}
	
	if cachedOutput != nil {
		// Use cached credentials
		output = cachedOutput
	} else {
		// Retrieve fresh credentials
		log.Printf("Retrieving fresh credentials from Lambda")
		credentials, err := retrieveCredentials(config)
		if err != nil {
			log.Printf("Failed to retrieve credentials: %v", err)
			os.Exit(1)
		}

		// Convert to output format
		output = &CredentialsProcessOutput{
			Version:         1,
			AccessKeyId:     credentials.AccessKeyID,
			SecretAccessKey: credentials.SecretAccessKey,
			SessionToken:    credentials.SessionToken,
		}

		// Add expiration if available
		if !credentials.Expires.IsZero() {
			output.Expiration = credentials.Expires.Format(time.RFC3339)
		}
		
		// Save to cache
		if err := saveCachedCredentials(config, output); err != nil {
			log.Printf("Warning: failed to save credentials to cache: %v", err)
		}
	}

	// Marshal and output JSON to stdout
	jsonOutput, err := json.Marshal(output)
	if err != nil {
		log.Printf("Failed to marshal credentials to JSON: %v", err)
		os.Exit(1)
	}

	// Output credentials to stdout (required for AWS CLI)
	fmt.Println(string(jsonOutput))
}

// parseFlags parses command line arguments and environment variables
func parseFlags() (*Config, error) {
	config := &Config{}

	// Command line flags
	flag.StringVar(&config.LambdaURL, "lambda-url", "", "AWS Lambda Function URL endpoint")
	flag.StringVar(&config.RoleARN, "role-arn", "", "IAM Role ARN to assume (must be in /remote/ path)")
	flag.StringVar(&config.TailnetKey, "tailnet-key", "", "Tailscale auth key for secure communication")
	flag.StringVar(&config.SessionName, "session-name", "post2post-credentials-process", "Session name for the assumed role")
	flag.DurationVar(&config.Duration, "duration", 1*time.Hour, "Credential duration (e.g., 1h, 30m)")
	flag.DurationVar(&config.Timeout, "timeout", 30*time.Second, "Request timeout (e.g., 30s, 1m)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "AWS credentials_process implementation using post2post for secure credential retrieval.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment variables (take precedence over flags):\n")
		fmt.Fprintf(os.Stderr, "  POST2POST_LAMBDA_URL     Lambda Function URL endpoint\n")
		fmt.Fprintf(os.Stderr, "  POST2POST_ROLE_ARN       IAM Role ARN to assume\n")
		fmt.Fprintf(os.Stderr, "  POST2POST_TAILNET_KEY    Tailscale auth key\n")
		fmt.Fprintf(os.Stderr, "  POST2POST_SESSION_NAME   Session name for assumed role\n")
		fmt.Fprintf(os.Stderr, "  POST2POST_DURATION       Credential duration (e.g., 1h, 30m)\n")
		fmt.Fprintf(os.Stderr, "  POST2POST_TIMEOUT        Request timeout (e.g., 30s, 1m)\n")
		fmt.Fprintf(os.Stderr, "\nExample usage in AWS config:\n")
		fmt.Fprintf(os.Stderr, "  [profile myprofile]\n")
		fmt.Fprintf(os.Stderr, "  credential_process = /usr/local/bin/post2post-credentials --lambda-url https://lambda-url.amazonaws.com/ --role-arn arn:aws:iam::123456789012:role/remote/MyRole --tailnet-key tskey-auth-xyz\n")
	}

	flag.Parse()

	// Override with environment variables if set
	if envLambdaURL := os.Getenv("POST2POST_LAMBDA_URL"); envLambdaURL != "" {
		config.LambdaURL = envLambdaURL
	}
	if envRoleARN := os.Getenv("POST2POST_ROLE_ARN"); envRoleARN != "" {
		config.RoleARN = envRoleARN
	}
	if envTailnetKey := os.Getenv("POST2POST_TAILNET_KEY"); envTailnetKey != "" {
		config.TailnetKey = envTailnetKey
	}
	if envSessionName := os.Getenv("POST2POST_SESSION_NAME"); envSessionName != "" {
		config.SessionName = envSessionName
	}
	if envDuration := os.Getenv("POST2POST_DURATION"); envDuration != "" {
		if duration, err := time.ParseDuration(envDuration); err == nil {
			config.Duration = duration
		} else {
			return nil, fmt.Errorf("invalid duration format in POST2POST_DURATION: %v", err)
		}
	}
	if envTimeout := os.Getenv("POST2POST_TIMEOUT"); envTimeout != "" {
		if timeout, err := time.ParseDuration(envTimeout); err == nil {
			config.Timeout = timeout
		} else {
			return nil, fmt.Errorf("invalid timeout format in POST2POST_TIMEOUT: %v", err)
		}
	}

	return config, nil
}

// validateConfig validates the configuration parameters
func validateConfig(config *Config) error {
	if config.LambdaURL == "" {
		return fmt.Errorf("lambda URL is required (use --lambda-url or POST2POST_LAMBDA_URL)")
	}
	if config.RoleARN == "" {
		return fmt.Errorf("role ARN is required (use --role-arn or POST2POST_ROLE_ARN)")
	}
	if config.TailnetKey == "" {
		return fmt.Errorf("tailnet key is required (use --tailnet-key or POST2POST_TAILNET_KEY)")
	}

	// Validate role ARN format (must be in /remote/ path for security)
	if !isValidRemoteRoleARN(config.RoleARN) {
		return fmt.Errorf("role ARN must be in /remote/ path for security (e.g., arn:aws:iam::123456789012:role/remote/MyRole)")
	}

	// Validate duration limits (AWS STS limits)
	if config.Duration < 15*time.Minute {
		return fmt.Errorf("credential duration must be at least 15 minutes")
	}
	if config.Duration > 12*time.Hour {
		return fmt.Errorf("credential duration cannot exceed 12 hours")
	}

	return nil
}

// isValidRemoteRoleARN checks if the role ARN is in the required /remote/ path
func isValidRemoteRoleARN(roleARN string) bool {
	// Expected format: arn:aws:iam::123456789012:role/remote/RoleName
	// Check basic format and that it contains "/remote/" in the role path
	if len(roleARN) < 40 {
		return false
	}
	
	// Check prefix
	if !strings.HasPrefix(roleARN, "arn:aws:iam::") {
		return false
	}
	
	// Check that it contains the required /remote/ path
	return strings.Contains(roleARN, ":role/remote/")
}

// retrieveCredentials uses the post2post AWS credentials provider to get credentials
func retrieveCredentials(config *Config) (aws.Credentials, error) {
	log.Printf("Initializing post2post credentials provider")
	log.Printf("Lambda URL: %s", config.LambdaURL)
	log.Printf("Role ARN: %s", config.RoleARN)
	log.Printf("Session Name: %s", config.SessionName)
	log.Printf("Duration: %s", config.Duration)

	// Create AWS credentials provider configuration
	providerConfig := post2post.AWSCredentialsProviderConfig{
		LambdaURL:   config.LambdaURL,
		RoleARN:     config.RoleARN,
		TailnetKey:  config.TailnetKey,
		SessionName: config.SessionName,
		Duration:    config.Duration,
	}

	// Create the credentials provider
	provider, err := post2post.NewAWSCredentialsProvider(providerConfig)
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("failed to create credentials provider: %w", err)
	}
	defer func() {
		if closeErr := provider.Close(); closeErr != nil {
			log.Printf("Warning: failed to close credentials provider: %v", closeErr)
		}
	}()

	// Retrieve credentials with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	log.Printf("Retrieving AWS credentials via post2post...")
	credentials, err := provider.Retrieve(ctx)
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("failed to retrieve credentials: %w", err)
	}

	log.Printf("Successfully retrieved credentials (expires: %s)", credentials.Expires.Format(time.RFC3339))
	
	return credentials, nil
}

// getCacheFilePath returns the path to the cache file based on session name
func getCacheFilePath(sessionName string) (string, error) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	
	// Create cache directory path
	cacheDir := filepath.Join(homeDir, ".cache")
	
	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}
	
	// Create cache file path using session name
	cacheFile := filepath.Join(cacheDir, sessionName)
	return cacheFile, nil
}

// loadCachedCredentials attempts to load valid cached credentials
func loadCachedCredentials(config *Config) (*CredentialsProcessOutput, error) {
	cacheFile, err := getCacheFilePath(config.SessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache file path: %w", err)
	}
	
	log.Printf("Checking for cached credentials in: %s", cacheFile)
	
	// Check if cache file exists
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("No cached credentials found")
			return nil, nil // No cache file exists
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}
	
	// Parse cached credentials
	var cached CachedCredentials
	if err := json.Unmarshal(data, &cached); err != nil {
		log.Printf("Invalid cache file format, ignoring: %v", err)
		return nil, nil // Invalid cache, ignore it
	}
	
	// Validate that cache matches current configuration
	if cached.RoleARN != config.RoleARN || cached.LambdaURL != config.LambdaURL {
		log.Printf("Cache configuration mismatch (RoleARN: %s vs %s, LambdaURL: %s vs %s), ignoring cache", 
			cached.RoleARN, config.RoleARN, cached.LambdaURL, config.LambdaURL)
		return nil, nil
	}
	
	// Check if credentials are still valid (not within 10 minutes of expiration)
	now := time.Now()
	expirationBuffer := 10 * time.Minute
	expiresWithBuffer := cached.ExpiresAt.Add(-expirationBuffer)
	
	if now.After(expiresWithBuffer) {
		log.Printf("Cached credentials expire soon (at %s, buffer until %s), refreshing", 
			cached.ExpiresAt.Format(time.RFC3339), expiresWithBuffer.Format(time.RFC3339))
		return nil, nil // Need to refresh
	}
	
	log.Printf("Using valid cached credentials (expires: %s)", cached.ExpiresAt.Format(time.RFC3339))
	return &cached.Credentials, nil
}

// saveCachedCredentials saves credentials to the cache file
func saveCachedCredentials(config *Config, credentials *CredentialsProcessOutput) error {
	cacheFile, err := getCacheFilePath(config.SessionName)
	if err != nil {
		return fmt.Errorf("failed to get cache file path: %w", err)
	}
	
	// Parse expiration time from credentials
	var expiresAt time.Time
	if credentials.Expiration != "" {
		if parsed, err := time.Parse(time.RFC3339, credentials.Expiration); err == nil {
			expiresAt = parsed
		} else {
			log.Printf("Warning: failed to parse expiration time, using 1 hour from now: %v", err)
			expiresAt = time.Now().Add(1 * time.Hour)
		}
	} else {
		// Default to 1 hour if no expiration provided
		expiresAt = time.Now().Add(1 * time.Hour)
	}
	
	// Create cached credentials structure
	cached := CachedCredentials{
		Credentials: *credentials,
		CachedAt:    time.Now(),
		ExpiresAt:   expiresAt,
		RoleARN:     config.RoleARN,
		LambdaURL:   config.LambdaURL,
	}
	
	// Marshal to JSON
	data, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cached credentials: %w", err)
	}
	
	// Write to cache file with restricted permissions
	if err := os.WriteFile(cacheFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}
	
	log.Printf("Cached credentials saved to: %s (expires: %s)", cacheFile, expiresAt.Format(time.RFC3339))
	return nil
}