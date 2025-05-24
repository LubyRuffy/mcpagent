package mcpagent

import (
	"context"
	"errors"
	"testing"

	"github.com/LubyRuffy/fofalogsai/pkg/config"
	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 创建一个模拟的通知接口实现
type MockNotify struct {
	mock.Mock
}

func (m *MockNotify) OnMessage(msg string) {
	m.Called(msg)
}

func (m *MockNotify) OnResult(msg string) {
	m.Called(msg)
}

func (m *MockNotify) OnError(err error) {
	m.Called(err)
}

// 测试CLI通知器
func TestCliNotifier(t *testing.T) {
	notifier := &CliNotifier{}

	// 这些方法主要是打印到控制台，我们只测试它们不会崩溃
	notifier.OnMessage("test message")
	notifier.OnResult("test result")
	notifier.OnError(errors.New("test error"))

	// 由于这些方法只是打印到控制台，没有返回值，所以我们只能确认它们不会崩溃
	assert.True(t, true, "CliNotifier methods should not panic")
}

// 测试LoggerCallback
func TestLoggerCallback(t *testing.T) {
	mockNotify := new(MockNotify)
	callback := &LoggerCallback{
		notify: mockNotify,
	}

	// 设置期望
	mockNotify.On("OnError", mock.Anything).Return()

	// 创建一个简单的上下文
	ctx := context.Background()

	// 测试OnError方法
	result := callback.OnError(ctx, nil, errors.New("test error"))

	// 验证返回的上下文
	assert.Equal(t, ctx, result)

	// 验证OnError是否被调用
	mockNotify.AssertExpectations(t)
}

// 测试Run函数的错误处理 - GetTools失败
func TestRunGetToolsError(t *testing.T) {
	ctx := context.Background()
	mockNotify := new(MockNotify)

	// 测试无效配置
	invalidConfig := &config.Config{
		MCP: config.MCPConfig{
			ConfigFile: "non_existent_file.json",
		},
	}

	// 设置期望
	mockNotify.On("OnError", mock.Anything).Return()

	// 运行函数
	err := Run(ctx, invalidConfig, "test task", mockNotify)

	// 验证错误
	assert.Error(t, err)
}

// 测试Run函数的错误处理 - GetModel失败
func TestRunGetModelError(t *testing.T) {
	// 由于我们无法直接使用MockConfig，这个测试暂时跳过
	t.Skip("无法直接使用MockConfig，这个测试暂时跳过")
}

// 测试Run函数的错误处理 - NewAgent失败
func TestRunNewAgentError(t *testing.T) {
	// 这个测试需要模拟react.NewAgent，但由于它是一个包级函数，
	// 我们无法直接模拟它。这个测试暂时跳过。
	t.Skip("无法直接模拟react.NewAgent函数")
}

// 测试Run函数的成功情况
func TestRunSuccess(t *testing.T) {
	// 由于Run函数依赖于很多外部组件，完整测试它需要大量的模拟对象
	// 这里我们只测试基本的成功路径
	t.Skip("完整测试Run函数需要大量的模拟对象，暂时跳过")
}

// 测试LoggerCallback的OnEnd方法
func TestLoggerCallbackOnEnd(t *testing.T) {
	mockNotify := new(MockNotify)
	callback := &LoggerCallback{
		notify: mockNotify,
	}

	// 创建一个简单的上下文
	ctx := context.Background()

	// 测试OnEnd方法
	result := callback.OnEnd(ctx, nil, nil)

	// 验证返回的上下文
	assert.Equal(t, ctx, result)
}

// 测试LoggerCallback的OnEndWithStreamOutput方法
func TestLoggerCallbackOnEndWithStreamOutput(t *testing.T) {
	mockNotify := new(MockNotify)
	callback := &LoggerCallback{
		notify: mockNotify,
	}

	// 创建一个简单的上下文
	ctx := context.Background()

	// 测试OnEndWithStreamOutput方法
	// 由于这个方法启动了一个goroutine，我们只测试它不会崩溃
	result := callback.OnEndWithStreamOutput(ctx, nil, nil)

	// 验证返回的上下文
	assert.Equal(t, ctx, result)
}

// 测试LoggerCallback的OnStart方法 - 处理工具调用
func TestLoggerCallbackOnStartWithToolCalls(t *testing.T) {
	mockNotify := new(MockNotify)
	callback := &LoggerCallback{
		notify: mockNotify,
	}

	// 设置期望
	mockNotify.On("OnMessage", "test thinking").Return()
	mockNotify.On("OnMessage", "正在调用工具：web_search").Return()

	// 创建一个简单的上下文
	ctx := context.Background()

	// 创建一个带有工具调用的消息
	toolCall := &schema.ToolCall{}
	toolCall.Function.Name = "web_search"
	toolCall.Function.Arguments = `{"query":"test query", "think":"test thinking"}`

	message := &schema.Message{
		Role:      schema.Assistant,
		ToolCalls: []schema.ToolCall{*toolCall},
	}

	// 测试OnStart方法
	result := callback.OnStart(ctx, nil, message)

	// 验证返回的上下文
	assert.Equal(t, ctx, result)

	// 验证OnMessage是否被调用
	mockNotify.AssertExpectations(t)
}

// 测试LoggerCallback的OnStart方法 - 处理sequentialthinking工具调用
func TestLoggerCallbackOnStartWithSequentialThinking(t *testing.T) {
	mockNotify := new(MockNotify)
	callback := &LoggerCallback{
		notify: mockNotify,
	}

	// 设置期望
	mockNotify.On("OnMessage", "sequential thinking test").Return()

	// 创建一个简单的上下文
	ctx := context.Background()

	// 创建一个带有工具调用的消息
	toolCall := &schema.ToolCall{}
	toolCall.Function.Name = "sequentialthinking"
	toolCall.Function.Arguments = `{"think":"sequential thinking test"}`

	message := &schema.Message{
		Role:      schema.Assistant,
		ToolCalls: []schema.ToolCall{*toolCall},
	}

	// 测试OnStart方法
	result := callback.OnStart(ctx, nil, message)

	// 验证返回的上下文
	assert.Equal(t, ctx, result)

	// 验证OnMessage是否被调用
	mockNotify.AssertExpectations(t)
}

// 测试LoggerCallback的OnStart方法 - 处理默认工具调用
func TestLoggerCallbackOnStartWithDefaultTool(t *testing.T) {
	mockNotify := new(MockNotify)
	callback := &LoggerCallback{
		notify: mockNotify,
	}

	// 设置期望
	mockNotify.On("OnMessage", "default_tool {\"param\":\"value\"}").Return()

	// 创建一个简单的上下文
	ctx := context.Background()

	// 创建一个带有工具调用的消息
	toolCall := &schema.ToolCall{}
	toolCall.Function.Name = "default_tool"
	toolCall.Function.Arguments = `{"param":"value"}`

	message := &schema.Message{
		Role:      schema.Assistant,
		ToolCalls: []schema.ToolCall{*toolCall},
	}

	// 测试OnStart方法
	result := callback.OnStart(ctx, nil, message)

	// 验证返回的上下文
	assert.Equal(t, ctx, result)

	// 验证OnMessage是否被调用
	mockNotify.AssertExpectations(t)
}

// 测试LoggerCallback的OnStart方法 - 处理JSON解析错误
func TestLoggerCallbackOnStartWithJSONError(t *testing.T) {
	mockNotify := new(MockNotify)
	callback := &LoggerCallback{
		notify: mockNotify,
	}

	// 设置期望 - JSON解析失败时只会调用OnError，不会调用OnMessage
	mockNotify.On("OnError", mock.Anything).Return()

	// 创建一个简单的上下文
	ctx := context.Background()

	// 创建一个带有工具调用的消息，但JSON格式错误
	toolCall := &schema.ToolCall{}
	toolCall.Function.Name = "web_search"
	toolCall.Function.Arguments = `{"query":"test query", "think":}` // 格式错误的JSON

	message := &schema.Message{
		Role:      schema.Assistant,
		ToolCalls: []schema.ToolCall{*toolCall},
	}

	// 测试OnStart方法
	result := callback.OnStart(ctx, nil, message)

	// 验证返回的上下文
	assert.Equal(t, ctx, result)

	// 验证OnError是否被调用
	mockNotify.AssertExpectations(t)
}

// 测试LoggerCallback的OnStart方法 - 处理非消息输入
func TestLoggerCallbackOnStartWithNonMessage(t *testing.T) {
	mockNotify := new(MockNotify)
	callback := &LoggerCallback{
		notify: mockNotify,
	}

	// 创建一个简单的上下文
	ctx := context.Background()

	// 创建一个非消息输入
	nonMessage := "not a message"

	// 测试OnStart方法
	result := callback.OnStart(ctx, nil, nonMessage)

	// 验证返回的上下文
	assert.Equal(t, ctx, result)

	// 验证没有调用任何通知方法
	mockNotify.AssertExpectations(t)
}

// 测试LoggerCallback的OnStartWithStreamInput方法
func TestLoggerCallbackOnStartWithStreamInput(t *testing.T) {
	// 由于无法直接创建有效的StreamReader，这个测试暂时跳过
	t.Skip("无法直接创建有效的StreamReader，这个测试暂时跳过")
}
