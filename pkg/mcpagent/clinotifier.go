// Package mcpagent provides notification implementations for the MCP agent.
// This file contains CLI-specific notification handlers for command-line applications.
package mcpagent

import (
	"fmt"
	"os"
)

// CliNotifier implements the Notify interface for command-line interface output.
// It provides simple console-based notifications suitable for CLI applications.
// All notifications are printed directly to stdout/stderr for immediate user feedback.
//
// This implementation is thread-safe and can be used concurrently from multiple
// goroutines without additional synchronization.
//
// Example usage:
//
//	notifier := mcpagent.NewCliNotifier()
//	err := mcpagent.Run(ctx, cfg, "task description", notifier)
type CliNotifier struct{}

// NewCliNotifier creates a new CLI notifier instance.
// This is the preferred way to create a CliNotifier and provides
// a consistent interface for future extensibility.
//
// Returns:
//   - *CliNotifier: A new CLI notifier ready for use
//
// Example:
//
//	notifier := mcpagent.NewCliNotifier()
//	notifier.OnMessage("Processing started...")
func NewCliNotifier() *CliNotifier {
	return &CliNotifier{}
}

// OnMessage prints a message notification to stdout.
// This is typically used for progress updates and informational messages
// during agent execution. Messages are printed with a newline for readability.
//
// The method is thread-safe and can be called concurrently.
//
// Parameters:
//   - msg: The message to display to the user
//
// Example:
//
//	notifier.OnMessage("正在分析网站结构...")
//	notifier.OnMessage("发现3个潜在安全问题")
func (n *CliNotifier) OnMessage(msg string) {
	fmt.Println("消息:", msg)
}

// OnResult prints a result notification to stdout.
// This is used when the agent completes successfully with a final result.
// The result is printed to stdout to allow for easy redirection and processing.
//
// Parameters:
//   - msg: The final result message from the agent
//
// Example:
//
//	notifier.OnResult("分析完成：网站安全评分为85分")
func (n *CliNotifier) OnResult(msg string) {
	fmt.Println("结果:", msg)
}

// OnError prints an error notification to stderr.
// This method outputs errors to stderr to distinguish them from normal output
// and allow for proper error handling in shell scripts and pipelines.
//
// The error is formatted with a clear "错误:" prefix for easy identification.
//
// Parameters:
//   - err: The error that occurred during agent execution
//
// Example:
//
//	notifier.OnError(fmt.Errorf("无法连接到目标服务器"))
//	// Output: 错误: 无法连接到目标服务器
func (n *CliNotifier) OnError(err error) {
	fmt.Fprintf(os.Stderr, "错误: %v\n", err)
}

// OnThinking prints a thinking notification to stdout.
// This is used when the agent is thinking about something.
//
// Parameters:
//   - msg: The thinking message from the agent
//
// Example:
//
//	notifier.OnThinking("正在思考中...")
//	// Output: 思考中: 正在思考中...
func (n *CliNotifier) OnThinking(msg string) {
	fmt.Println("思考中:", msg)
}

// OnToolCall prints a tool call notification to stdout.
// This is used when the agent is calling a tool.
//
// Parameters:
//   - toolName: The name of the tool being called
//   - params: The parameters for the tool call
//
// Example:
//
//	notifier.OnToolCall("web_search", map[string]interface{}{"query": "test query"})
//	// Output: 正在调用工具: web_search, 参数: map[query:test query]
func (n *CliNotifier) OnToolCall(toolName string, params any) {
	fmt.Printf("正在调用工具: %s, 参数: %v\n", toolName, params)
}
