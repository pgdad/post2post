package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pgdad/post2post"
)

func main() {
	// Parse command line arguments to determine which processor to use
	var processor post2post.PayloadProcessor
	
	processorType := "echo" // default
	if len(os.Args) > 1 {
		processorType = strings.ToLower(os.Args[1])
	}
	
	switch processorType {
	case "hello":
		processor = &post2post.HelloWorldProcessor{}
		fmt.Println("Using Hello World Processor")
		
	case "echo":
		processor = &post2post.EchoProcessor{}
		fmt.Println("Using Echo Processor")
		
	case "timestamp":
		processor = &post2post.TimestampProcessor{}
		fmt.Println("Using Timestamp Processor")
		
	case "counter":
		processor = post2post.NewCounterProcessor()
		fmt.Println("Using Counter Processor")
		
	case "advanced":
		processor = post2post.NewAdvancedContextProcessor("demo-receiver")
		fmt.Println("Using Advanced Context Processor")
		
	case "transform":
		processor = &post2post.TransformProcessor{}
		fmt.Println("Using Transform Processor")
		
	case "validator":
		processor = post2post.NewValidatorProcessor([]string{"name", "email"})
		fmt.Println("Using Validator Processor (requires 'name' and 'email' fields)")
		
	case "chain":
		processor = post2post.NewChainProcessor(
			&post2post.TimestampProcessor{},
			&post2post.TransformProcessor{},
			&post2post.EchoProcessor{},
		)
		fmt.Println("Using Chain Processor (timestamp -> transform -> echo)")
		
	default:
		fmt.Printf("Unknown processor type: %s\n", processorType)
		fmt.Println("Available processors: hello, echo, timestamp, counter, advanced, transform, validator, chain")
		os.Exit(1)
	}
	
	// Create and configure the server
	server := post2post.NewServer().
		WithInterface("127.0.0.1").
		WithProcessor(processor)
	
	// Start the server
	err := server.Start()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()
	
	fmt.Printf("Receiving server started at: %s\n", server.GetURL())
	fmt.Println("Send POST requests to /webhook endpoint")
	fmt.Println("Available endpoints:")
	fmt.Printf("  - %s/webhook (for payload processing)\n", server.GetURL())
	fmt.Printf("  - %s/roundtrip (for round-trip responses)\n", server.GetURL())
	fmt.Printf("  - %s/ (for server info)\n", server.GetURL())
	
	// Keep the server running
	select {}
}