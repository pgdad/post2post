package main

import (
	"fmt"
	"log"
	"time"
	
	"github.com/pgdad/post2post"
)

func main() {
	fmt.Println("AWS Lambda Post2Post Test Client")
	fmt.Println("=================================")
	
	// Replace with your actual Lambda Function URL
	lambdaURL := "https://your-lambda-url.lambda-url.us-east-1.on.aws/"
	
	// Create post2post server for receiving responses
	server := post2post.NewServer().
		WithInterface("127.0.0.1").
		WithPostURL(lambdaURL).
		WithTimeout(60 * time.Second) // Longer timeout for Lambda cold starts
	
	// Start the local server to receive responses
	err := server.Start()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()
	
	fmt.Printf("Local server started at: %s\n", server.GetURL())
	fmt.Printf("Will post to Lambda at: %s\n", lambdaURL)
	fmt.Println()
	
	// Test 1: Basic IAM role assumption without Tailscale
	fmt.Println("Test 1: Basic IAM role assumption")
	fmt.Println("---------------------------------")
	
	payload1 := map[string]interface{}{
		"message":     "Hello from post2post client!",
		"client_info": map[string]interface{}{
			"version": "1.0",
			"type":    "test",
		},
		"timestamp": time.Now().Unix(),
		// IMPORTANT: Replace with an actual IAM role ARN that your Lambda can assume
		"role_arn":  "arn:aws:iam::123456789012:role/ExampleRole",
	}
	
	response1, err := server.RoundTripPost(payload1)
	if err != nil {
		log.Printf("Test 1 failed: %v", err)
	} else {
		printLambdaResponse("Test 1", response1)
	}
	
	fmt.Println()
	
	// Test 2: IAM role assumption with Tailscale integration
	fmt.Println("Test 2: IAM role assumption with Tailscale")
	fmt.Println("------------------------------------------")
	
	payload2 := map[string]interface{}{
		"secure_data": "This should use Tailscale networking",
		"environment": "production",
		"role_arn":    "arn:aws:iam::123456789012:role/SecureRole",
	}
	
	// Note: Replace with your actual Tailscale auth key
	// For demo purposes, we'll use the PostJSONWithTailnet method
	err = server.PostJSONWithTailnet(payload2, "tskey-auth-example-key")
	if err != nil {
		fmt.Printf("Test 2 Tailscale demo: %v\n", err)
		fmt.Println("Note: This demonstrates the Tailscale framework integration.")
		fmt.Println("To enable full functionality, configure tsnet in both client and Lambda.")
	} else {
		fmt.Println("Tailscale integration request sent successfully!")
	}
	
	fmt.Println()
	
	// Test 3: Error case - invalid role ARN
	fmt.Println("Test 3: Error handling (invalid role ARN)")
	fmt.Println("----------------------------------------")
	
	payload3 := map[string]interface{}{
		"test_case":   "error_handling",
		"role_arn":    "arn:aws:iam::123456789012:role/NonExistentRole",
	}
	
	response3, err := server.RoundTripPost(payload3)
	if err != nil {
		log.Printf("Test 3 failed with client error: %v", err)
	} else {
		printLambdaResponse("Test 3 (Error Case)", response3)
	}
	
	fmt.Println()
	
	// Test 4: Missing role ARN
	fmt.Println("Test 4: Validation error (missing role_arn)")
	fmt.Println("-------------------------------------------")
	
	payload4 := map[string]interface{}{
		"test_case": "validation_error",
		"message":   "This should fail validation",
		// Missing role_arn field
	}
	
	response4, err := server.RoundTripPost(payload4)
	if err != nil {
		log.Printf("Test 4 failed with client error: %v", err)
	} else {
		printLambdaResponse("Test 4 (Validation Error)", response4)
	}
	
	fmt.Println()
	fmt.Println("All tests completed!")
	fmt.Println()
	printUsageInstructions()
}

func printLambdaResponse(testName string, response *post2post.RoundTripResponse) {
	fmt.Printf("%s Result:\n", testName)
	fmt.Printf("  Success: %v\n", response.Success)
	fmt.Printf("  Timeout: %v\n", response.Timeout)
	fmt.Printf("  Request ID: %s\n", response.RequestID)
	
	if response.Error != "" {
		fmt.Printf("  Error: %s\n", response.Error)
	}
	
	if response.Success && response.Payload != nil {
		fmt.Printf("  Lambda Response:\n")
		if payloadMap, ok := response.Payload.(map[string]interface{}); ok {
			// Print key information from the Lambda response
			if status, exists := payloadMap["status"]; exists {
				fmt.Printf("    Status: %v\n", status)
			}
			if processedBy, exists := payloadMap["processed_by"]; exists {
				fmt.Printf("    Processed by: %v\n", processedBy)
			}
			if processedAt, exists := payloadMap["processed_at"]; exists {
				fmt.Printf("    Processed at: %v\n", processedAt)
			}
			
			// Show AssumeRole result if present
			if assumeResult, exists := payloadMap["assume_role_result"]; exists {
				fmt.Printf("    AssumeRole Success: ✓\n")
				if result, ok := assumeResult.(map[string]interface{}); ok {
					if creds, exists := result["credentials"]; exists {
						if credMap, ok := creds.(map[string]interface{}); ok {
							if accessKey, exists := credMap["access_key_id"]; exists {
								accessKeyStr := accessKey.(string)
								fmt.Printf("    Access Key: %s...%s\n", 
									accessKeyStr[:min(4, len(accessKeyStr))],
									accessKeyStr[max(0, len(accessKeyStr)-4):])
							}
							if expiration, exists := credMap["expiration"]; exists {
								fmt.Printf("    Expires: %v\n", expiration)
							}
						}
					}
					if assumedRole, exists := result["assumed_role_user"]; exists {
						if roleMap, ok := assumedRole.(map[string]interface{}); ok {
							if arn, exists := roleMap["arn"]; exists {
								fmt.Printf("    Assumed Role ARN: %v\n", arn)
							}
						}
					}
				}
			}
		} else {
			fmt.Printf("    Raw payload: %+v\n", response.Payload)
		}
	}
	fmt.Println()
}

func printUsageInstructions() {
	fmt.Println("Usage Instructions:")
	fmt.Println("==================")
	fmt.Println()
	fmt.Println("1. Deploy the Lambda function:")
	fmt.Println("   cd examples/aws-lambda")
	fmt.Println("   ./build.sh")
	fmt.Println("   cd terraform && terraform apply")
	fmt.Println()
	fmt.Println("2. Update the Lambda URL in this test client:")
	fmt.Println("   Replace 'your-lambda-url' with your actual Lambda Function URL")
	fmt.Println()
	fmt.Println("3. Update IAM role ARNs:")
	fmt.Println("   Replace the example role ARNs with roles your Lambda can assume")
	fmt.Println()
	fmt.Println("4. For Tailscale integration:")
	fmt.Println("   - Get a Tailscale auth key from your admin console")
	fmt.Println("   - Update the Lambda environment variable TAILSCALE_AUTH_KEY")
	fmt.Println("   - Enable tsnet integration in both client and Lambda code")
	fmt.Println()
	fmt.Println("5. Run the test:")
	fmt.Println("   go run test-client.go")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}