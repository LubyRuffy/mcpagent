// Package examples provides usage examples for the mcpagent package.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/LubyRuffy/mcpagent/pkg/config"
	"github.com/LubyRuffy/mcpagent/pkg/mcpagent"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run examples/stream_example.go \"your task description\"")
		os.Exit(1)
	}

	// Get the task from command line argument
	task := os.Args[1]

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Load configuration
	cfg, err := config.LoadConfig("default_config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Example 1: Run with streaming via StreamCliNotifier
	fmt.Println("======= Example 1: Using StreamCliNotifier =======")
	fmt.Println("This will stream responses directly to the console:")

	// Create a streaming CLI notifier
	streamNotify := mcpagent.NewStreamCliNotifier()

	// Execute task with streaming
	if err := mcpagent.Run(ctx, cfg, task, streamNotify); err != nil {
		log.Fatalf("Failed to execute task: %v", err)
	}

	// fmt.Println("\n\n======= Example 2: Using RunStream API =======")
	// fmt.Println("This will provide programmatic access to the stream:")

	// // Create another streaming CLI notifier
	// streamNotify2 := mcpagent.NewStreamCliNotifier()

	// // Get a stream reader
	// stream, err := mcpagent.RunStream(ctx, cfg, task, streamNotify2)
	// if err != nil {
	// 	log.Fatalf("Failed to start streaming task: %v", err)
	// }
	// defer stream.Close()

	// // Process the stream manually
	// fmt.Println("Manual stream processing:")
	// for {
	// 	chunk, err := stream.Recv()
	// 	if errors.Is(err, io.EOF) {
	// 		fmt.Println("\nStream completed")
	// 		break
	// 	}
	// 	if err != nil {
	// 		log.Fatalf("Stream error: %v", err)
	// 	}

	// 	// Process each message chunk
	// 	if chunk.Content != "" {
	// 		fmt.Print(chunk.Content)
	// 	}

	// 	// If the message has tool calls, report them
	// 	if len(chunk.ToolCalls) > 0 {
	// 		fmt.Printf("\n[Using %d tool(s)]\n", len(chunk.ToolCalls))
	// 	}
	// }

	fmt.Println("\nExamples completed successfully")
}
