package mcpagent

import (
	"context"

	"github.com/LubyRuffy/mcpagent/pkg/config"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/mock"
)

// MockConfig 是一个模拟的 config.Config 对象
type MockConfig struct {
	mock.Mock
	config.Config
}

// GetTools 模拟 GetTools 方法
func (m *MockConfig) GetTools(ctx context.Context) ([]tool.BaseTool, func(), error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]tool.BaseTool), args.Get(1).(func()), args.Error(2)
}

// GetModel 模拟 GetModel 方法
func (m *MockConfig) GetModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(model.ToolCallingChatModel), args.Error(1)
}

// MockToolCallingChatModel 是一个模拟的 model.ToolCallingChatModel 对象
type MockToolCallingChatModel struct {
	mock.Mock
}

// Chat 模拟 Chat 方法
func (m *MockToolCallingChatModel) Chat(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	args := m.Called(ctx, messages)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.Message), args.Error(1)
}

// ChatStream 模拟 ChatStream 方法
func (m *MockToolCallingChatModel) ChatStream(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	args := m.Called(ctx, messages)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.StreamReader[*schema.Message]), args.Error(1)
}

// Generate 模拟 Generate 方法
func (m *MockToolCallingChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	args := m.Called(ctx, messages, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.Message), args.Error(1)
}

// GenerateStream 模拟 GenerateStream 方法
func (m *MockToolCallingChatModel) GenerateStream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	args := m.Called(ctx, messages, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.StreamReader[*schema.Message]), args.Error(1)
}

// Stream 模拟 Stream 方法
func (m *MockToolCallingChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	args := m.Called(ctx, messages, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.StreamReader[*schema.Message]), args.Error(1)
}

// WithTools 模拟 WithTools 方法
func (m *MockToolCallingChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	args := m.Called(tools)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(model.ToolCallingChatModel), args.Error(1)
}

// Info 模拟 Info 方法
func (m *MockToolCallingChatModel) Info() interface{} {
	args := m.Called()
	return args.Get(0)
}

// MockBaseTool 是一个模拟的 tool.BaseTool 对象
type MockBaseTool struct {
	mock.Mock
}

// Info 模拟 Info 方法
func (m *MockBaseTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.ToolInfo), args.Error(1)
}

// Run 模拟 Run 方法
func (m *MockBaseTool) Run(ctx context.Context, params map[string]interface{}) (string, error) {
	args := m.Called(ctx, params)
	return args.String(0), args.Error(1)
}
