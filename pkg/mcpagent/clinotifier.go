// Package mcpagent provides notification implementations for the MCP agent.
package mcpagent

import "fmt"

// CliNotifier implements the Notify interface for command-line interface output.
// It provides simple console-based notifications suitable for CLI applications.
// All notifications are printed directly to stdout/stderr.
type CliNotifier struct{}

// NewCliNotifier creates a new CLI notifier instance.
// This is the preferred way to create a CliNotifier.
func NewCliNotifier() *CliNotifier {
	return &CliNotifier{}
}

// OnMessage prints a message notification to stdout.
// This is typically used for progress updates and informational messages.
func (n *CliNotifier) OnMessage(msg string) {
	fmt.Println(msg)
}

// OnResult prints a result notification to stdout.
// This is used when the agent completes successfully with a final result.
func (n *CliNotifier) OnResult(msg string) {
	fmt.Println(msg)
}

// OnError prints an error notification to stdout.
// In a production CLI application, this might be better suited for stderr.
func (n *CliNotifier) OnError(err error) {
	fmt.Printf("错误: %v\n", err)
}
