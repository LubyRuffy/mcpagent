package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

// MockMCPHub 是一个模拟的 MCPHub 结构体
type MockMCPHub struct {
	ConfigFile string
	Tools      []string
	ShouldFail bool
}

// CloseServers 模拟关闭服务器
func (m *MockMCPHub) CloseServers() error {
	// 模拟关闭服务器的行为
	return nil
}

// GetEinoTools 模拟获取工具
func (m *MockMCPHub) GetEinoTools(ctx context.Context, toolNameList []string) ([]tool.BaseTool, error) {
	if m.ShouldFail {
		return nil, fmt.Errorf("模拟获取工具失败")
	}

	var tools []tool.BaseTool
	for _, toolName := range toolNameList {
		// 创建一个模拟的工具
		mockTool := utils.NewTool(
			&schema.ToolInfo{
				Name: toolName,
				Desc: "Mock tool for testing",
			},
			func(ctx context.Context, params map[string]interface{}) (string, error) {
				return "Mock tool result", nil
			},
		)
		tools = append(tools, mockTool)
	}
	return tools, nil
}

// 替换 mcphost.NewMCPHub 函数，用于测试
var originalNewMCPHub = mcphostNewMCPHub

// 定义一个接口，包含我们需要模拟的方法
type mcpHubInterface interface {
	GetEinoTools(ctx context.Context, toolNameList []string) ([]tool.BaseTool, error)
	CloseServers() error
}

// 创建临时配置文件用于测试
func createTempConfigFile(t TestingT, content string) string {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")
	err := os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("创建临时配置文件失败: %v", err)
	}
	return configPath
}

// TestingT 是一个接口，用于测试
type TestingT interface {
	Fatalf(format string, args ...interface{})
	TempDir() string
}

// 重置 mcphostNewMCPHub 为原始函数
func resetMCPHostNewMCPHub() {
	mcphostNewMCPHub = originalNewMCPHub
}

// 设置 mcphostNewMCPHub 为模拟函数
func setMockMCPHostNewMCPHub(shouldFail bool) {
	// 这个函数已经不再使用，保留为空实现
}
