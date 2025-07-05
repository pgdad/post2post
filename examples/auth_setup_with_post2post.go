package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pgdad/post2post"
)

func main() {
	fmt.Println("Post2Post Tailscale OAuth Auth Key Generator")
	fmt.Println("============================================")
	fmt.Println()

	// Get OAuth client credentials from environment
	clientID := os.Getenv("TS_API_CLIENT_ID")
	clientSecret := os.Getenv("TS_API_CLIENT_SECRET")

	if clientID == "" {
		log.Fatal("TS_API_CLIENT_ID environment variable is required")
	}
	if clientSecret == "" {
		log.Fatal("TS_API_CLIENT_SECRET environment variable is required")
	}

	// Optional configuration from environment
	tags := os.Getenv("TAILSCALE_TAGS")
	if tags == "" {
		tags = "tag:ephemeral-device"
		fmt.Printf("Using default tags: %s\n", tags)
		fmt.Println("Set TAILSCALE_TAGS environment variable to override (comma-separated)")
	} else {
		fmt.Printf("Using tags: %s\n", tags)
	}

	// Validate tag format
	tagList := strings.Split(tags, ",")
	for _, tag := range tagList {
		tag = strings.TrimSpace(tag)
		if !strings.HasPrefix(tag, "tag:") {
			fmt.Printf("⚠️  Warning: Tag '%s' doesn't start with 'tag:' prefix\n", tag)
		}
	}
	
	fmt.Println()
	fmt.Println("📋 Configuration:")
	fmt.Printf("   Client ID: %s...\n", clientID[:min(10, len(clientID))])
	fmt.Printf("   Tags: %s\n", tags)
	fmt.Printf("   Ephemeral: true\n")
	fmt.Printf("   Reusable: true\n")
	fmt.Printf("   Preauthorized: false\n")
	fmt.Println()

	// Create a post2post server instance to use the OAuth method
	server := post2post.NewServer()

	// Generate ephemeral auth key using post2post's OAuth integration
	fmt.Println("🔑 Generating Tailscale auth key using post2post OAuth integration...")
	authKey, err := server.GenerateTailnetKeyFromOAuth(
		true,  // reusable
		true,  // ephemeral
		false, // preauthorized
		tags,  // tags
	)
	
	if err != nil {
		log.Fatalf("Failed to generate auth key: %v", err)
	}

	fmt.Printf("✅ Successfully generated ephemeral auth key!\n")
	fmt.Println()

	// Output the auth key and details
	fmt.Println("Generated Auth Key:")
	fmt.Println("==================")
	fmt.Printf("Key: %s\n", authKey)
	fmt.Printf("Type: Ephemeral, Reusable\n")
	fmt.Printf("Tags: %s\n", tags)
	fmt.Printf("Generated at: %s\n", time.Now().Format("2006-01-02 15:04:05 MST"))
	fmt.Println()

	// Output the actual auth key
	fmt.Println("=== AUTH KEY ===")
	fmt.Println(authKey)
	fmt.Println("===============")
	fmt.Println()

	// Usage instructions
	fmt.Println("Usage Instructions:")
	fmt.Println("==================")
	fmt.Println("1. Use this auth key to connect ephemeral devices:")
	fmt.Printf("   tailscale up --auth-key='%s'\n", authKey)
	fmt.Println()
	fmt.Println("2. For automated systems, export as environment variable:")
	fmt.Printf("   export TAILSCALE_AUTH_KEY='%s'\n", authKey)
	fmt.Println()
	fmt.Println("3. In your post2post applications:")
	fmt.Printf("   server.PostJSONWithTailnet(payload, \"%s\")\n", authKey)
	fmt.Println()
	fmt.Println("4. For client_tailnet.go:")
	fmt.Printf("   export TAILSCALE_AUTH_KEY='%s'\n", authKey)
	fmt.Printf("   go run client_tailnet.go\n")
	fmt.Println()
	fmt.Println("Note: This is an ephemeral auth key. Devices using it will be")
	fmt.Println("automatically removed when they go offline.")
	fmt.Println()
	fmt.Println("🔗 Integration with post2post:")
	fmt.Println("   - Use the generated key with PostJSONWithTailnet()")
	fmt.Println("   - Use with RoundTripPost() for secure round-trip messaging")
	fmt.Println("   - Compatible with all post2post Tailscale examples")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}