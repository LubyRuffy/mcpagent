package config

import (
	"context"
	"testing"

	"github.com/LubyRuffy/mcpagent/pkg/mcphost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMCPConfigIntegration 测试MCP配置的集成功能
func TestMCPConfigIntegration(t *testing.T) {
	t.Run("使用mcp_servers配置", func(t *testing.T) {
		// 创建包含mcp_servers的配置
		cfg := &Config{
			MCP: MCPConfig{
				MCPServers: map[string]mcphost.ServerConfig{
					"test-server": {
						TransportType: "stdio",
						Command:       "echo",
						Args:          []string{"hello", "world"},
						Env: map[string]string{
							"TEST_VAR": "test_value",
						},
					},
				},
				Tools: []string{"test-tool"},
			},
			LLM: LLMConfig{
				Type:    LLMProviderOllama,
				BaseURL: "http://localhost:11434",
				Model:   "test-model",
				APIKey:  "test-key",
			},
			MaxStep: 20,
		}

		// 验证配置
		err := cfg.Validate()
		assert.NoError(t, err, "配置验证应该成功")

		// 验证MCP配置结构
		assert.NotNil(t, cfg.MCP.MCPServers, "MCPServers不应该为nil")
		assert.Equal(t, 1, len(cfg.MCP.MCPServers), "应该有1个服务器")
		assert.Contains(t, cfg.MCP.MCPServers, "test-server", "应该包含test-server")

		// 验证服务器配置
		server := cfg.MCP.MCPServers["test-server"]
		assert.Equal(t, "stdio", server.TransportType)
		assert.Equal(t, "echo", server.Command)
		assert.Equal(t, []string{"hello", "world"}, server.Args)
		assert.Equal(t, "test_value", server.Env["TEST_VAR"])
	})

	t.Run("空的mcp_servers配置", func(t *testing.T) {
		// 创建空的mcp_servers配置
		cfg := &Config{
			MCP: MCPConfig{
				MCPServers: make(map[string]mcphost.ServerConfig),
				Tools:      []string{},
			},
			LLM: LLMConfig{
				Type:    LLMProviderOllama,
				BaseURL: "http://localhost:11434",
				Model:   "test-model",
				APIKey:  "test-key",
			},
			MaxStep: 20,
		}

		// 验证配置
		err := cfg.Validate()
		assert.NoError(t, err, "空的mcp_servers配置验证应该成功")

		// 验证MCP配置结构
		assert.NotNil(t, cfg.MCP.MCPServers, "MCPServers不应该为nil")
		assert.Equal(t, 0, len(cfg.MCP.MCPServers), "应该有0个服务器")
	})

	t.Run("nil的mcp_servers配置应该失败", func(t *testing.T) {
		// 创建nil的mcp_servers配置
		cfg := &Config{
			MCP: MCPConfig{
				MCPServers: nil,
				ConfigFile: "", // 空的ConfigFile
				Tools:      []string{},
			},
			LLM: LLMConfig{
				Type:    LLMProviderOllama,
				BaseURL: "http://localhost:11434",
				Model:   "test-model",
				APIKey:  "test-key",
			},
			MaxStep: 20,
		}

		// 验证配置应该失败
		err := cfg.Validate()
		assert.Error(t, err, "nil的mcp_servers配置验证应该失败")
		assert.Contains(t, err.Error(), "MCP配置文件路径不能为空")
	})

	t.Run("默认配置应该有效", func(t *testing.T) {
		// 获取默认配置
		cfg := NewDefaultConfig()

		// 验证配置
		err := cfg.Validate()
		assert.NoError(t, err, "默认配置验证应该成功")

		// 验证MCP配置结构
		assert.NotNil(t, cfg.MCP.MCPServers, "默认配置的MCPServers不应该为nil")
		assert.Equal(t, 0, len(cfg.MCP.MCPServers), "默认配置应该有0个服务器")
		assert.Equal(t, "", cfg.MCP.ConfigFile, "默认配置的ConfigFile应该为空")
	})

	t.Run("GetTools方法应该使用mcp_servers", func(t *testing.T) {
		// 保存原始函数
		originalFactory := mcpHubFromSettingsFactory
		defer func() {
			mcpHubFromSettingsFactory = originalFactory
		}()

		// 创建模拟函数
		var capturedSettings *mcphost.MCPSettings
		mcpHubFromSettingsFactory = func(ctx context.Context, settings *mcphost.MCPSettings) (MCPHubInterface, error) {
			capturedSettings = settings
			// 返回错误而不是nil，避免空指针
			return nil, assert.AnError
		}

		cfg := &Config{
			MCP: MCPConfig{
				MCPServers: map[string]mcphost.ServerConfig{
					"test-server": {
						TransportType: "stdio",
						Command:       "echo",
					},
				},
				Tools: []string{"test-tool"},
			},
		}

		ctx := context.Background()
		_, _, err := cfg.GetTools(ctx)

		// 我们期望这里会有错误（因为我们返回了错误）
		assert.Error(t, err, "应该有错误因为我们返回了错误")

		// 验证传递给mcpHubFromSettingsFactory的参数
		require.NotNil(t, capturedSettings, "应该调用了mcpHubFromSettingsFactory")
		assert.Equal(t, 1, len(capturedSettings.MCPServers), "应该传递了1个服务器")
		assert.Contains(t, capturedSettings.MCPServers, "test-server", "应该包含test-server")
	})
}

// TestMCPConfigBackwardCompatibility 测试向后兼容性
func TestMCPConfigBackwardCompatibility(t *testing.T) {
	t.Run("仍然支持config_file配置", func(t *testing.T) {
		// 创建使用config_file的配置
		cfg := &Config{
			MCP: MCPConfig{
				ConfigFile: "test_mcpservers.json",
				MCPServers: nil, // nil表示不使用mcp_servers
				Tools:      []string{"test-tool"},
			},
			LLM: LLMConfig{
				Type:    LLMProviderOllama,
				BaseURL: "http://localhost:11434",
				Model:   "test-model",
				APIKey:  "test-key",
			},
			MaxStep: 20,
		}

		// 验证配置
		err := cfg.Validate()
		assert.NoError(t, err, "使用config_file的配置验证应该成功")
	})

	t.Run("mcp_servers优先于config_file", func(t *testing.T) {
		// 保存原始函数
		originalSettingsFactory := mcpHubFromSettingsFactory
		originalFileFactory := mcpHubFactory
		defer func() {
			mcpHubFromSettingsFactory = originalSettingsFactory
			mcpHubFactory = originalFileFactory
		}()

		// 创建模拟函数来验证调用
		settingsFactoryCalled := false
		fileFactoryCalled := false

		mcpHubFromSettingsFactory = func(ctx context.Context, settings *mcphost.MCPSettings) (MCPHubInterface, error) {
			settingsFactoryCalled = true
			return nil, assert.AnError
		}

		mcpHubFactory = func(ctx context.Context, configFile string) (MCPHubInterface, error) {
			fileFactoryCalled = true
			return nil, assert.AnError
		}

		cfg := &Config{
			MCP: MCPConfig{
				ConfigFile: "test_mcpservers.json", // 有config_file
				MCPServers: map[string]mcphost.ServerConfig{ // 也有mcp_servers
					"test-server": {
						TransportType: "stdio",
						Command:       "echo",
					},
				},
				Tools: []string{"test-tool"},
			},
		}

		ctx := context.Background()
		_, _, err := cfg.GetTools(ctx)

		// 验证调用了正确的工厂函数
		assert.Error(t, err, "应该有错误因为我们返回了nil")
		assert.True(t, settingsFactoryCalled, "应该调用了mcpHubFromSettingsFactory")
		assert.False(t, fileFactoryCalled, "不应该调用mcpHubFactory")
	})
}
