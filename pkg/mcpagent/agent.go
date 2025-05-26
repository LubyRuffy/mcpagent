// Package mcpagent provides MCP (Model Context Protocol) agent functionality
// for executing tasks with tool calling capabilities and notification support.
//
// The package implements a ReAct (Reasoning and Acting) agent that can:
//   - Execute complex tasks using available MCP tools
//   - Provide real-time notifications during execution
//   - Handle streaming responses and tool calls
//   - Support multiple LLM providers (OpenAI, Ollama)
//   - Manage tool execution lifecycle with proper cleanup
//
// Example usage:
//
//	notify := &mcpagent.CliNotifier{}
//	err := mcpagent.Run(ctx, cfg, "分析这个网站的安全性", notify)
//	if err != nil {
//		log.Fatal(err)
//	}
package mcpagent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/LubyRuffy/mcpagent/pkg/config"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

// Tool name constants for special handling of specific tools
const (
	// ToolSequentialThinking represents the sequential thinking tool name
	ToolSequentialThinking = "sequentialthinking"
	// ToolWebSearch represents the web search tool name
	ToolWebSearch = "web_search"
	// ToolURLMarkdown represents the URL markdown tool name
	ToolURLMarkdown = "url_markdown"
)

// Field name constants for tool arguments parsing
var (
	// ThinkFieldName represents the think field name in tool arguments
	ThinkFieldName = []string{"think", "thought"}
)

// Error message constants provide consistent error reporting
const (
	errMsgConfigNil         = "配置不能为空"
	errMsgTaskEmpty         = "任务不能为空"
	errMsgNotifyNil         = "通知处理器不能为空"
	errMsgGetToolsFailed    = "获取工具失败: %w"
	errMsgGetModelFailed    = "获取模型失败: %w"
	errMsgCreateAgentFailed = "创建agent失败: %w"
	errMsgExecuteTaskFailed = "执行任务失败: %w"
	errMsgParseArgsFailed   = "解析工具参数失败: %w"
	errMsgHandleToolFailed  = "处理工具调用失败: %w"
	errMsgFormatMsgFailed   = "格式化消息失败: %w"
	errMsgGenerateOutFailed = "生成输出失败: %w"
	errMsgSerializeFrame    = "序列化流帧失败: %w"
)

// Notify defines the interface for handling various types of notifications
// during agent execution. Implementations should handle these notifications
// appropriately for their context (CLI, web UI, etc.).
//
// The interface provides three types of notifications:
//   - OnMessage: For progress updates and informational messages
//   - OnThinking: For progress updates and informational messages
//   - OnToolCall: For tool calls
//   - OnResult: For final results when the agent completes successfully
//   - OnError: For error notifications when something goes wrong
type Notify interface {
	// OnMessage sends a message notification during execution
	// This is typically used for progress updates and informational messages
	OnMessage(msg string)

	// OnThinking sends a thinking notification during execution
	OnThinking(msg string)

	// OnToolCall sends a tool call notification during execution
	OnToolCall(toolName string, params any)

	// OnResult sends a result notification when the agent completes successfully
	// This contains the final output from the agent
	OnResult(msg string)

	// OnError sends an error notification when something goes wrong
	// This should be used for all error conditions during execution
	OnError(err error)
}

// LoggerCallback implements the callback interface for logging and notification
// during agent execution. It provides hooks for different stages of the agent's
// lifecycle including start, end, error, and streaming operations.
//
// The callback processes tool calls and extracts relevant information for
// user notification, particularly handling special tools like sequential thinking
// and web-related operations.
//
// This callback is designed to be thread-safe and can handle concurrent
// operations from the agent framework.
type LoggerCallback struct {
	notify                   Notify // Notification handler for user feedback
	callbacks.HandlerBuilder        // Embedded handler builder for callback implementation
}

// OnStart is called when a callback operation starts. It processes tool calls
// and sends appropriate notifications based on the tool type.
// This method is particularly important for providing real-time feedback
// during tool execution.
//
// The method filters messages to only process assistant messages with tool calls,
// ensuring that only relevant tool execution events trigger notifications.
//
// Parameters:
//   - ctx: Context for the operation
//   - info: Runtime information about the callback
//   - input: Input data for the callback (expected to be a schema.Message)
//
// Returns:
//   - context.Context: The same context (no modifications)
func (cb *LoggerCallback) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	message, ok := input.(*schema.Message)
	if !ok {
		return ctx
	}

	if message.Role != schema.Assistant || len(message.ToolCalls) == 0 {
		return ctx
	}

	cb.processToolCalls(message.ToolCalls)
	return ctx
}

// processToolCalls processes the tool calls and sends appropriate notifications.
// It handles each tool call individually and logs any errors that occur during processing.
// This method ensures that errors in processing one tool call don't affect others.
//
// Parameters:
//   - toolCalls: List of tool calls to process
func (cb *LoggerCallback) processToolCalls(toolCalls []schema.ToolCall) {
	for _, toolCall := range toolCalls {
		if err := cb.handleSingleToolCall(toolCall); err != nil {
			log.Printf("处理工具调用失败: %v", err)
			cb.notify.OnError(fmt.Errorf(errMsgHandleToolFailed, err))
		}
	}
}

// handleSingleToolCall processes a single tool call and extracts relevant information.
// It parses the tool arguments and delegates to specific handlers based on tool type.
// This method provides the main logic for interpreting different types of tool calls.
//
// Parameters:
//   - toolCall: The tool call to process
//
// Returns:
//   - error: Error if processing fails
func (cb *LoggerCallback) handleSingleToolCall(toolCall schema.ToolCall) error {
	arguments, err := cb.parseToolArguments(toolCall.Function.Arguments)
	if err != nil {
		return fmt.Errorf(errMsgParseArgsFailed, err)
	}

	cb.handleGenericTool(toolCall.Function.Name, arguments)

	return nil
}

// parseToolArguments parses the JSON arguments of a tool call.
// It converts the JSON string to a map for easier processing and handles
// empty or malformed JSON gracefully.
//
// Parameters:
//   - arguments: JSON string containing tool arguments
//
// Returns:
//   - map[string]any: Parsed arguments as a map
//   - error: Error if JSON parsing fails
func (cb *LoggerCallback) parseToolArguments(arguments string) (map[string]any, error) {
	argStr := strings.TrimSpace(arguments)
	if argStr == "" {
		return make(map[string]interface{}), nil
	}

	var parsedArgs map[string]interface{}
	if err := json.Unmarshal([]byte(argStr), &parsedArgs); err != nil {
		return nil, err
	}
	return parsedArgs, nil
}

// handleThinkingTool handles the sequential thinking tool.
// It extracts the thinking content and sends it as a notification.
// This tool is special because it represents the agent's reasoning process.
//
// Parameters:
//   - arguments: Parsed tool arguments containing thinking content
func (cb *LoggerCallback) handleThinkingTool(arguments map[string]any) {
	for _, fieldName := range ThinkFieldName {
		if thinkValue, exists := arguments[fieldName]; exists {
			if thinkStr, ok := thinkValue.(string); ok && strings.TrimSpace(thinkStr) != "" {
				cb.notify.OnThinking(thinkStr)
				return
			}
		}
	}
}

// handleGenericTool handles all other tools with a generic approach.
// It provides basic notification about tool execution with arguments.
//
// Parameters:
//   - toolName: Name of the tool being executed
//   - arguments: Raw JSON arguments string
func (cb *LoggerCallback) handleGenericTool(toolName string, arguments map[string]any) {
	// First, handle any thinking content
	cb.handleThinkingTool(arguments)

	// Then notify about the tool execution
	cb.notify.OnToolCall(toolName, arguments)
}

// OnEnd is called when a callback operation ends successfully.
// Currently, no specific handling is needed for successful completion.
//
// Parameters:
//   - ctx: Context for the operation
//   - info: Runtime information about the callback
//   - output: Output data from the callback
//
// Returns:
//   - context.Context: The same context (no modifications)
func (cb *LoggerCallback) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	// Currently no specific handling needed for OnEnd
	return ctx
}

// OnError is called when a callback operation encounters an error.
// It forwards the error to the notification handler for user feedback.
//
// Parameters:
//   - ctx: Context for the operation
//   - info: Runtime information about the callback
//   - err: The error that occurred
//
// Returns:
//   - context.Context: The same context (no modifications)
func (cb *LoggerCallback) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	cb.notify.OnError(err)
	return ctx
}

// OnEndWithStreamOutput handles the end of streaming output operations.
// It processes streaming output in a separate goroutine to avoid blocking
// the main execution flow.
//
// Parameters:
//   - ctx: Context for the operation
//   - info: Runtime information about the callback
//   - output: Stream reader for callback output
//
// Returns:
//   - context.Context: The same context (no modifications)
func (cb *LoggerCallback) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo,
	output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {

	go cb.handleStreamOutput(info, output)
	return ctx
}

// handleStreamOutput processes streaming output in a separate goroutine.
// It reads from the stream until EOF and processes each frame.
// The method includes panic recovery to ensure stability.
//
// Parameters:
//   - info: Runtime information about the callback
//   - output: Stream reader for callback output
func (cb *LoggerCallback) handleStreamOutput(info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("[StreamOutput] 恢复从panic: %v", err)
		}
	}()

	defer output.Close()

	for {
		frame, err := output.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Printf("流输出内部错误: %v", err)
			return
		}

		if err := cb.processStreamFrame(info, frame); err != nil {
			log.Printf("处理流帧错误: %v", err)
		}
	}
}

// processStreamFrame processes a single frame from the stream output.
// It serializes the frame and logs it if it's from the main graph.
// This helps with debugging and monitoring agent execution.
//
// Parameters:
//   - info: Runtime information about the callback
//   - frame: The stream frame to process
//
// Returns:
//   - error: Error if frame processing fails
func (cb *LoggerCallback) processStreamFrame(info *callbacks.RunInfo, frame callbacks.CallbackOutput) error {
	frameData, err := json.Marshal(frame)
	if err != nil {
		return fmt.Errorf(errMsgSerializeFrame, err)
	}

	// 仅打印 graph 的输出, 否则每个 stream 节点的输出都会打印一遍
	if info.Name == react.GraphName {
		fmt.Printf("%s: %s\n", info.Name, string(frameData))
	}

	return nil
}

// OnStartWithStreamInput handles the start of streaming input operations.
// It ensures proper cleanup of the input stream.
//
// Parameters:
//   - ctx: Context for the operation
//   - info: Runtime information about the callback
//   - input: Stream reader for callback input
//
// Returns:
//   - context.Context: The same context (no modifications)
func (cb *LoggerCallback) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo,
	input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	defer input.Close()
	return ctx
}

// Run executes an MCP Agent task with the provided configuration and notification handler.
//
// This function is the main entry point for executing agent tasks. It orchestrates
// the entire process including:
//  1. Input parameter validation
//  2. Tool and model initialization
//  3. Agent creation and configuration
//  4. Task execution with proper error handling
//
// The function ensures proper resource cleanup and provides comprehensive error
// reporting through the notification interface.
//
// Parameters:
//   - ctx: Context for controlling execution flow and cancellation
//   - cfg: Configuration containing model, tool, and system settings
//   - task: Task description to execute (must not be empty)
//   - notify: Notification handler for progress updates and results
//
// Returns:
//   - error: Error if execution fails at any stage
//
// Example:
//
//	notify := &mcpagent.CliNotifier{}
//	err := mcpagent.Run(ctx, cfg, "分析这个网站的安全性", notify)
//	if err != nil {
//		log.Printf("任务执行失败: %v", err)
//	}
func Run(ctx context.Context, cfg *config.Config, task string, notify Notify) error {
	// 输入参数验证
	if err := validateRunParameters(cfg, task, notify); err != nil {
		return err
	}

	// 获取工具
	einoTools, cleanup, err := cfg.GetTools(ctx)
	if err != nil {
		return fmt.Errorf(errMsgGetToolsFailed, err)
	}
	defer cleanup()

	// 获取模型
	toolableChatModel, err := cfg.GetModel(ctx)
	if err != nil {
		return fmt.Errorf(errMsgGetModelFailed, err)
	}

	// 创建agent
	ragent, err := createReActAgent(ctx, cfg, einoTools, toolableChatModel)
	if err != nil {
		return fmt.Errorf(errMsgCreateAgentFailed, err)
	}

	// 执行任务
	return executeAgentTask(ctx, cfg, ragent, task, notify)
}

// validateRunParameters validates the input parameters for the Run function.
// It ensures all required parameters are provided and not nil/empty.
//
// Parameters:
//   - cfg: Configuration to validate
//   - task: Task string to validate
//   - notify: Notification handler to validate
//
// Returns:
//   - error: Validation error if any parameter is invalid
func validateRunParameters(cfg *config.Config, task string, notify Notify) error {
	if cfg == nil {
		return errors.New(errMsgConfigNil)
	}
	if strings.TrimSpace(task) == "" {
		return errors.New(errMsgTaskEmpty)
	}
	if notify == nil {
		return errors.New(errMsgNotifyNil)
	}
	return nil
}

// createReActAgent creates and configures a ReAct agent with the provided tools and model.
// It sets up the agent with appropriate configuration including maximum steps
// and tool integration.
//
// The agent is configured with a step multiplier to allow for complex reasoning
// that may require multiple iterations.
//
// Parameters:
//   - ctx: Context for the operation
//   - cfg: Configuration containing agent settings
//   - einoTools: List of tools available to the agent
//   - chatModel: Chat model for the agent to use
//
// Returns:
//   - *react.Agent: Configured ReAct agent ready for task execution
//   - error: Error if agent creation fails
func createReActAgent(ctx context.Context, cfg *config.Config, einoTools []tool.BaseTool, chatModel model.ToolCallingChatModel) (*react.Agent, error) {
	tools := compose.ToolsNodeConfig{
		Tools: einoTools,
	}

	agentConfig := &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig:      tools,
		MaxStep:          cfg.MaxStep * 5, // Allow more steps for complex reasoning
	}

	return react.NewAgent(ctx, agentConfig)
}

// executeAgentTask executes the specific agent task with proper message formatting
// and callback handling. It creates the chat template, formats the system prompt,
// and manages the agent execution lifecycle.
//
// The function handles template variable substitution including current date
// and configuration parameters.
//
// Parameters:
//   - ctx: Context for the operation
//   - cfg: Configuration containing system prompt and other settings
//   - ragent: Configured ReAct agent to execute the task
//   - task: Task description to execute
//   - notify: Notification handler for results
//
// Returns:
//   - error: Error if task execution fails
func executeAgentTask(ctx context.Context, cfg *config.Config, ragent *react.Agent, task string, notify Notify) error {
	// 创建聊天模板
	chatTemplate := prompt.FromMessages(schema.FString,
		&schema.Message{
			Role:    schema.System,
			Content: cfg.SystemPrompt,
		},
		&schema.Message{
			Role:    schema.User,
			Content: task,
		})

	// 格式化消息
	placeHolders := map[string]any{
		"date": time.Now().Format("2006-01-02"),
	}
	for k, v := range cfg.PlaceHolders {
		placeHolders[k] = v
	}

	msg, err := chatTemplate.Format(ctx, placeHolders)
	if err != nil {
		return fmt.Errorf(errMsgFormatMsgFailed, err)
	}

	// 生成输出
	output, err := ragent.Generate(ctx, msg, agent.WithComposeOptions(
		compose.WithCallbacks(&LoggerCallback{
			notify: notify,
		})))
	if err != nil {
		return fmt.Errorf(errMsgGenerateOutFailed, err)
	}

	notify.OnResult(output.Content)
	return nil
}
