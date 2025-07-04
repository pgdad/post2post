package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// OAuthTokenResponse represents the response from Tailscale OAuth token endpoint
type OAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

// AuthKeyRequest represents the request to create an auth key
type AuthKeyRequest struct {
	Capabilities struct {
		Devices struct {
			Create struct {
				Reusable      bool     `json:"reusable"`
				Ephemeral     bool     `json:"ephemeral"`
				Preauthorized bool     `json:"preauthorized"`
				Tags          []string `json:"tags"`
			} `json:"create"`
		} `json:"devices"`
	} `json:"capabilities"`
	ExpirySeconds int `json:"expirySeconds"`
	Description   string `json:"description"`
}

// AuthKeyResponse represents the response from creating an auth key
type AuthKeyResponse struct {
	ID          string   `json:"id"`
	Key         string   `json:"key"`
	Description string   `json:"description"`
	Created     string   `json:"created"`
	Expires     string   `json:"expires"`
	Capabilities struct {
		Devices struct {
			Create struct {
				Reusable      bool     `json:"reusable"`
				Ephemeral     bool     `json:"ephemeral"`
				Preauthorized bool     `json:"preauthorized"`
				Tags          []string `json:"tags"`
			} `json:"create"`
		} `json:"devices"`
	} `json:"capabilities"`
}

func main() {
	fmt.Println("Tailscale OAuth Auth Key Generator")
	fmt.Println("==================================")
	fmt.Println()

	// Get OAuth client credentials from environment
	clientID := os.Getenv("TAILSCALE_OAUTH_CLIENT_ID")
	clientSecret := os.Getenv("TAILSCALE_OAUTH_CLIENT_SECRET")
	tailnet := os.Getenv("TAILSCALE_TAILNET")

	if clientID == "" {
		log.Fatal("TAILSCALE_OAUTH_CLIENT_ID environment variable is required")
	}
	if clientSecret == "" {
		log.Fatal("TAILSCALE_OAUTH_CLIENT_SECRET environment variable is required")
	}
	if tailnet == "" {
		log.Fatal("TAILSCALE_TAILNET environment variable is required (e.g., 'example.com' or 'tail123abc.ts.net')")
	}

	// Optional configuration from environment
	tags := strings.Split(os.Getenv("TAILSCALE_TAGS"), ",")
	if len(tags) == 1 && tags[0] == "" {
		tags = []string{"tag:ephemeral-device"}
		fmt.Printf("Using default tags: %v\n", tags)
		fmt.Println("Set TAILSCALE_TAGS environment variable to override (comma-separated)")
	} else {
		fmt.Printf("Using tags: %v\n", tags)
	}

	// Validate tag format
	for _, tag := range tags {
		if !strings.HasPrefix(tag, "tag:") {
			fmt.Printf("⚠️  Warning: Tag '%s' doesn't start with 'tag:' prefix\n", tag)
		}
	}
	
	fmt.Println()
	fmt.Println("📋 Tag Configuration Requirements:")
	fmt.Println("   1. Tags must be defined in ACL 'tagOwners' section")
	fmt.Println("   2. OAuth client must have permission to use these tags")
	fmt.Println("   3. Tags must use 'tag:' prefix")
	fmt.Println("   4. Set tagOwners to empty array [] for OAuth client access")
	fmt.Println("   📖 See TAILSCALE_SETUP_GUIDE.md for detailed instructions")

	description := os.Getenv("TAILSCALE_KEY_DESCRIPTION")
	if description == "" {
		description = fmt.Sprintf("Ephemeral auth key generated at %s", time.Now().Format("2006-01-02 15:04:05"))
	}

	fmt.Printf("Tailnet: %s\n", tailnet)
	fmt.Printf("Description: %s\n", description)
	fmt.Println()

	// Step 1: Get OAuth access token
	fmt.Println("Step 1: Obtaining OAuth access token...")
	accessToken, err := getOAuthToken(clientID, clientSecret)
	if err != nil {
		log.Fatalf("Failed to get OAuth access token: %v", err)
	}
	fmt.Printf("✓ Successfully obtained access token (expires in %d seconds)\n", accessToken.ExpiresIn)
	fmt.Println()

	// Step 2: Create ephemeral auth key
	fmt.Println("Step 2: Creating ephemeral auth key...")
	authKey, err := createAuthKey(accessToken.AccessToken, tailnet, tags, description)
	if err != nil {
		log.Fatalf("Failed to create auth key: %v", err)
	}

	fmt.Printf("✓ Successfully created ephemeral auth key\n")
	fmt.Println()

	// Output the auth key and details
	fmt.Println("Auth Key Details:")
	fmt.Println("================")
	fmt.Printf("Key ID: %s\n", authKey.ID)
	fmt.Printf("Description: %s\n", authKey.Description)
	fmt.Printf("Created: %s\n", authKey.Created)
	fmt.Printf("Expires: %s\n", authKey.Expires)
	fmt.Printf("Ephemeral: %t\n", authKey.Capabilities.Devices.Create.Ephemeral)
	fmt.Printf("Reusable: %t\n", authKey.Capabilities.Devices.Create.Reusable)
	fmt.Printf("Preauthorized: %t\n", authKey.Capabilities.Devices.Create.Preauthorized)
	fmt.Printf("Tags: %v\n", authKey.Capabilities.Devices.Create.Tags)
	fmt.Println()

	// Output the actual auth key
	fmt.Println("=== AUTH KEY ===")
	fmt.Println(authKey.Key)
	fmt.Println("===============")
	fmt.Println()

	// Usage instructions
	fmt.Println("Usage Instructions:")
	fmt.Println("==================")
	fmt.Println("1. Use this auth key to connect ephemeral devices:")
	fmt.Printf("   tailscale up --auth-key='%s'\n", authKey.Key)
	fmt.Println()
	fmt.Println("2. For automated systems, export as environment variable:")
	fmt.Printf("   export TAILSCALE_AUTH_KEY='%s'\n", authKey.Key)
	fmt.Println()
	fmt.Println("3. In your applications:")
	fmt.Printf("   TAILSCALE_AUTH_KEY='%s' ./your-app\n", authKey.Key)
	fmt.Println()
	fmt.Println("Note: This is an ephemeral auth key. Devices using it will be")
	fmt.Println("automatically removed when they go offline.")
}

// getOAuthToken obtains an OAuth access token using client credentials flow
func getOAuthToken(clientID, clientSecret string) (*OAuthTokenResponse, error) {
	// Prepare OAuth token request
	tokenURL := "https://api.tailscale.com/api/v2/oauth/token"
	
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("scope", "devices")

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make OAuth token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read OAuth token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OAuth token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResponse OAuthTokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OAuth token response: %w", err)
	}

	return &tokenResponse, nil
}

// createAuthKey creates an ephemeral auth key using the Tailscale API
func createAuthKey(accessToken, tailnet string, tags []string, description string) (*AuthKeyResponse, error) {
	// Prepare auth key creation request
	authKeyURL := fmt.Sprintf("https://api.tailscale.com/api/v2/tailnet/%s/keys", tailnet)
	
	authKeyReq := AuthKeyRequest{
		ExpirySeconds: 7776000, // 90 days (maximum allowed)
		Description:   description,
	}

	// Configure capabilities for ephemeral device creation
	authKeyReq.Capabilities.Devices.Create.Reusable = true
	authKeyReq.Capabilities.Devices.Create.Ephemeral = true
	authKeyReq.Capabilities.Devices.Create.Preauthorized = false
	authKeyReq.Capabilities.Devices.Create.Tags = tags

	reqBody, err := json.Marshal(authKeyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal auth key request: %w", err)
	}

	req, err := http.NewRequest("POST", authKeyURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create auth key request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make auth key request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read auth key response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Provide specific guidance for common tag-related errors
		if resp.StatusCode == 422 || resp.StatusCode == 400 {
			if strings.Contains(string(body), "invalid") && strings.Contains(string(body), "tag") {
				return nil, fmt.Errorf(`auth key creation failed: %s

COMMON ISSUE: Tags not properly configured in ACL
To fix this issue:

1. Go to Tailscale Admin Console > Access Controls
2. Add your tags to the "tagOwners" section:
   {
     "tagOwners": {
       "tag:ephemeral-device": [],
       "tag:ci": [],
       "tag:automation": []
     }
   }

3. Ensure your OAuth client has permission to use these tags
4. Tags must use the "tag:" prefix

For detailed setup instructions, see TAILSCALE_SETUP_GUIDE.md`, resp.StatusCode, string(body))
			}
		}
		return nil, fmt.Errorf("auth key creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authKeyResponse AuthKeyResponse
	err = json.Unmarshal(body, &authKeyResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse auth key response: %w", err)
	}

	return &authKeyResponse, nil
}