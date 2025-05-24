// Package mcpagent provides MCP (Model Context Protocol) agent functionality
// for executing tasks with tool calling capabilities and notification support.
package mcpagent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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

// 常量定义
const (
	// ToolSequentialThinking represents the sequential thinking tool name
	ToolSequentialThinking = "sequentialthinking"
	// ToolWebSearch represents the web search tool name
	ToolWebSearch = "web_search"
	// ToolURLMarkdown represents the URL markdown tool name
	ToolURLMarkdown = "url_markdown"
	// ThinkFieldName represents the think field name in tool arguments
	ThinkFieldName = "think"
	// ToolCallingPrefix represents the prefix for tool calling messages
	ToolCallingPrefix = "正在调用工具："
)

// Notify defines the interface for handling various types of notifications
// during agent execution.
type Notify interface {
	// OnMessage sends a message notification
	OnMessage(msg string)
	// OnResult sends a result notification when the agent completes successfully
	OnResult(msg string)
	// OnError sends an error notification when something goes wrong
	OnError(err error)
}

// LoggerCallback implements the callback interface for logging and notification
// during agent execution. It provides hooks for different stages of the agent's
// lifecycle including start, end, error, and streaming operations.
type LoggerCallback struct {
	notify                   Notify
	callbacks.HandlerBuilder // 可以用 callbacks.HandlerBuilder 来辅助实现 callback
}

// OnStart is called when a callback operation starts. It processes tool calls
// and sends appropriate notifications based on the tool type.
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

// processToolCalls processes the tool calls and sends appropriate notifications
func (cb *LoggerCallback) processToolCalls(toolCalls []schema.ToolCall) {
	for _, toolCall := range toolCalls {
		if err := cb.handleSingleToolCall(toolCall); err != nil {
			log.Printf("处理工具调用失败: %v", err)
			cb.notify.OnError(fmt.Errorf("处理工具调用失败: %w", err))
		}
	}
}

// handleSingleToolCall processes a single tool call and extracts relevant information
func (cb *LoggerCallback) handleSingleToolCall(toolCall schema.ToolCall) error {
	arguments, err := cb.parseToolArguments(toolCall.Function.Arguments)
	if err != nil {
		return fmt.Errorf("解析工具参数失败: %w", err)
	}

	switch toolCall.Function.Name {
	case ToolSequentialThinking:
		cb.handleThinkingTool(arguments)
	case ToolWebSearch, ToolURLMarkdown:
		cb.handleWebTool(toolCall.Function.Name, arguments)
	default:
		cb.handleGenericTool(toolCall.Function.Name, toolCall.Function.Arguments)
	}

	return nil
}

// parseToolArguments parses the JSON arguments of a tool call
func (cb *LoggerCallback) parseToolArguments(arguments string) (map[string]interface{}, error) {
	var parsedArgs map[string]interface{}
	if err := json.Unmarshal([]byte(arguments), &parsedArgs); err != nil {
		return nil, err
	}
	return parsedArgs, nil
}

// handleThinkingTool handles the sequential thinking tool
func (cb *LoggerCallback) handleThinkingTool(arguments map[string]interface{}) {
	if thinkValue, exists := arguments[ThinkFieldName]; exists {
		if thinkStr, ok := thinkValue.(string); ok {
			cb.notify.OnMessage(thinkStr)
		}
	}
}

// handleWebTool handles web-related tools (search, markdown)
func (cb *LoggerCallback) handleWebTool(toolName string, arguments map[string]interface{}) {
	if thinkValue, exists := arguments[ThinkFieldName]; exists {
		if thinkStr, ok := thinkValue.(string); ok {
			cb.notify.OnMessage(thinkStr)
		}
	}
	cb.notify.OnMessage(ToolCallingPrefix + toolName)
}

// handleGenericTool handles all other tools with a generic approach
func (cb *LoggerCallback) handleGenericTool(toolName, arguments string) {
	cb.notify.OnMessage(fmt.Sprintf("%s %s", toolName, arguments))
}

// OnEnd is called when a callback operation ends successfully
func (cb *LoggerCallback) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	// Currently no specific handling needed for OnEnd
	return ctx
}

// OnError is called when a callback operation encounters an error
func (cb *LoggerCallback) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	cb.notify.OnError(err)
	return ctx
}

// OnEndWithStreamOutput handles the end of streaming output operations
func (cb *LoggerCallback) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo,
	output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {

	go cb.handleStreamOutput(info, output)
	return ctx
}

// handleStreamOutput processes streaming output in a separate goroutine
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

// processStreamFrame processes a single frame from the stream output
func (cb *LoggerCallback) processStreamFrame(info *callbacks.RunInfo, frame callbacks.CallbackOutput) error {
	frameData, err := json.Marshal(frame)
	if err != nil {
		return fmt.Errorf("序列化流帧失败: %w", err)
	}

	// 仅打印 graph 的输出, 否则每个 stream 节点的输出都会打印一遍
	if info.Name == react.GraphName {
		fmt.Printf("%s: %s\n", info.Name, string(frameData))
	}

	return nil
}

// OnStartWithStreamInput handles the start of streaming input operations
func (cb *LoggerCallback) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo,
	input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	defer input.Close()
	return ctx
}

// Run 运行MCP Agent
//
// 该函数是主要的入口点，用于设置和执行MCP Agent任务。它会：
// 1. 验证输入参数
// 2. 获取配置的工具和模型
// 3. 创建并配置ReAct Agent
// 4. 执行任务并处理结果
//
// 参数:
// - ctx: 上下文，用于控制执行流程和传递元数据
// - cfg: 配置对象，包含模型、工具和系统提示等配置信息
// - task: 要执行的任务描述
// - notify: 通知处理器，用于接收执行过程中的各种通知
//
// 返回:
// - error: 如果执行过程中出现错误则返回错误信息
func Run(ctx context.Context, cfg *config.Config, task string, notify Notify) error {
	// 输入参数验证
	if err := validateRunParameters(cfg, task, notify); err != nil {
		return err
	}

	// 获取工具
	einoTools, cleanup, err := cfg.GetTools(ctx)
	if err != nil {
		return fmt.Errorf("获取工具失败: %w", err)
	}
	defer cleanup()

	// 获取模型
	toolableChatModel, err := cfg.GetModel(ctx)
	if err != nil {
		return fmt.Errorf("获取模型失败: %w", err)
	}

	// 创建agent
	ragent, err := createReActAgent(ctx, cfg, einoTools, toolableChatModel)
	if err != nil {
		return fmt.Errorf("创建agent失败: %w", err)
	}

	// 执行任务
	return executeAgentTask(ctx, cfg, ragent, task, notify)
}

// validateRunParameters 验证Run函数的输入参数
func validateRunParameters(cfg *config.Config, task string, notify Notify) error {
	if cfg == nil {
		return errors.New("配置不能为空")
	}
	if task == "" {
		return errors.New("任务不能为空")
	}
	if notify == nil {
		return errors.New("通知处理器不能为空")
	}
	return nil
}

// createReActAgent 创建并配置ReAct agent
func createReActAgent(ctx context.Context, cfg *config.Config, einoTools []tool.BaseTool, chatModel model.ToolCallingChatModel) (*react.Agent, error) {
	tools := compose.ToolsNodeConfig{
		Tools: einoTools,
	}

	agentConfig := &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig:      tools,
		MaxStep:          cfg.MaxStep * 5,
	}

	return react.NewAgent(ctx, agentConfig)
}

// executeAgentTask 执行具体的agent任务
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
	msg, err := chatTemplate.Format(ctx, map[string]interface{}{
		"date":           time.Now().Format("2006-01-02"),
		"total_thoughts": cfg.MaxStep,
	})
	if err != nil {
		return fmt.Errorf("格式化消息失败: %w", err)
	}

	// 生成输出
	output, err := ragent.Generate(ctx, msg, agent.WithComposeOptions(
		compose.WithCallbacks(&LoggerCallback{
			notify: notify,
		})))
	if err != nil {
		return fmt.Errorf("生成输出失败: %w", err)
	}

	notify.OnResult(output.Content)
	return nil
}
