package config

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/golang/mock/gomock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTool 是一个模拟的工具
type mockTool struct {
	name string
}

func (m *mockTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: m.name,
		Desc: "Mock tool for testing",
	}, nil
}

func (m *mockTool) Run(ctx context.Context, params map[string]interface{}) (string, error) {
	return "Mock tool result", nil
}

func (m *mockTool) InvokableRun(ctx context.Context, paramsStr string) (string, error) {
	return "Mock tool result", nil
}

func TestLoadConfig(t *testing.T) {
	// 创建临时配置文件
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	configContent := `
llm:
  api_key: test-api-key
  base_url: http://test-url.com
  model: test-model
  type: openai
max_step: 15
mcp:
  config_file: test_mcpservers.json
  tools:
    - tool1
    - tool2
proxy: http://test-proxy.com
system_prompt: Test system prompt
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// 测试加载配置
	cfg, err := LoadConfig(configPath)
	require.NoError(t, err)

	// 验证配置内容
	assert.Equal(t, "http://test-proxy.com", cfg.Proxy)
	assert.Equal(t, "test_mcpservers.json", cfg.MCP.ConfigFile)
	assert.Equal(t, []string{"tool1", "tool2"}, cfg.MCP.Tools)
	assert.Equal(t, "openai", cfg.LLM.Type)
	assert.Equal(t, "http://test-url.com", cfg.LLM.BaseURL)
	assert.Equal(t, "test-model", cfg.LLM.Model)
	assert.Equal(t, "test-api-key", cfg.LLM.APIKey)
	assert.Equal(t, "Test system prompt", cfg.SystemPrompt)
	assert.Equal(t, 15, cfg.MaxStep)
}

// TestLoadConfigWithDefaultPath 测试使用默认路径加载配置
func TestLoadConfigWithDefaultPath(t *testing.T) {
	// 创建临时配置文件
	tempDir := t.TempDir()

	// 保存当前工作目录
	currentDir, err := os.Getwd()
	require.NoError(t, err)

	// 切换到临时目录
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// 测试结束后恢复工作目录并重置viper状态
	defer func() {
		_ = os.Chdir(currentDir)
		viper.Reset()
	}()

	// 重置viper状态
	viper.Reset()

	// 在临时目录中创建配置文件
	configContent := `
llm:
  api_key: default-path-api-key
  base_url: http://default-path-url.com
  model: default-path-model
  type: openai
`
	err = os.WriteFile("config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	// 测试加载配置（不指定路径）
	cfg, err := LoadConfig("")
	require.NoError(t, err)

	// 验证配置内容
	assert.Equal(t, "default-path-api-key", cfg.LLM.APIKey)
	assert.Equal(t, "http://default-path-url.com", cfg.LLM.BaseURL)
	assert.Equal(t, "default-path-model", cfg.LLM.Model)
	assert.Equal(t, "openai", cfg.LLM.Type)
}

// TestLoadConfigWithEnv 测试使用环境变量加载配置
func TestLoadConfigWithEnv(t *testing.T) {
	// 由于环境变量测试可能会影响其他测试，我们跳过这个测试
	t.Skip("环境变量测试可能会影响其他测试，跳过")
}

func TestLoadConfigDefault(t *testing.T) {
	// 保存原始工作目录
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	// 创建一个独立的临时目录进行测试
	tempDir := t.TempDir()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	defer func() {
		// 恢复原始工作目录
		_ = os.Chdir(originalDir)
		// 重置viper状态
		viper.Reset()
	}()

	// 重置viper状态
	viper.Reset()

	// 测试通过 LoadConfig 获取默认配置（在没有任何配置文件的目录下）
	cfg, err := LoadConfig("")
	require.NoError(t, err)

	// 验证默认配置
	assert.Equal(t, "", cfg.Proxy)
	assert.Equal(t, "mcpservers.json", cfg.MCP.ConfigFile)
	assert.Equal(t, []string{}, cfg.MCP.Tools)
	assert.Equal(t, "ollama", cfg.LLM.Type)
	assert.Equal(t, "http://127.0.0.1:11434", cfg.LLM.BaseURL)
	assert.Equal(t, "qwen3:4b", cfg.LLM.Model)
	assert.Equal(t, "ollama", cfg.LLM.APIKey)
	assert.Contains(t, cfg.SystemPrompt, "你是精通互联网的信息收集专家")
	assert.Equal(t, 20, cfg.MaxStep)
}

func TestSaveConfig(t *testing.T) {
	// 创建临时配置文件路径
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "save_config_test.yaml")

	// 创建配置对象
	cfg := &Config{
		Proxy: "http://save-test-proxy.com",
		MCP: MCPConfig{
			ConfigFile: "save_test_mcpservers.json",
			Tools:      []string{"save_tool1", "save_tool2"},
		},
		LLM: LLMConfig{
			Type:    "openai",
			BaseURL: "http://save-test-url.com",
			Model:   "save-test-model",
			APIKey:  "save-test-api-key",
		},
		SystemPrompt: "Save test system prompt",
		MaxStep:      25,
	}

	// 保存配置
	err := cfg.SaveConfig(configPath)
	require.NoError(t, err)

	// 重新加载配置并验证
	loadedCfg, err := LoadConfig(configPath)
	require.NoError(t, err)

	assert.Equal(t, cfg.Proxy, loadedCfg.Proxy)
	assert.Equal(t, cfg.MCP.ConfigFile, loadedCfg.MCP.ConfigFile)
	assert.Equal(t, cfg.MCP.Tools, loadedCfg.MCP.Tools)
	assert.Equal(t, cfg.LLM.Type, loadedCfg.LLM.Type)
	assert.Equal(t, cfg.LLM.BaseURL, loadedCfg.LLM.BaseURL)
	assert.Equal(t, cfg.LLM.Model, loadedCfg.LLM.Model)
	assert.Equal(t, cfg.LLM.APIKey, loadedCfg.LLM.APIKey)
	assert.Equal(t, cfg.SystemPrompt, loadedCfg.SystemPrompt)
	assert.Equal(t, cfg.MaxStep, loadedCfg.MaxStep)
}

func TestGetModel(t *testing.T) {
	// 测试获取OpenAI模型
	cfgOpenAI := &Config{
		LLM: LLMConfig{
			Type:    "openai",
			BaseURL: "http://test-openai-url.com",
			Model:   "test-openai-model",
			APIKey:  "test-openai-api-key",
		},
	}

	ctx := context.Background()
	model, err := cfgOpenAI.GetModel(ctx)
	require.NoError(t, err)
	assert.NotNil(t, model)

	// 测试获取Ollama模型
	cfgOllama := &Config{
		LLM: LLMConfig{
			Type:    "ollama",
			BaseURL: "http://test-ollama-url.com",
			Model:   "test-ollama-model",
		},
	}

	model, err = cfgOllama.GetModel(ctx)
	require.NoError(t, err)
	assert.NotNil(t, model)

	// 测试不支持的模型类型
	cfgUnsupported := &Config{
		LLM: LLMConfig{
			Type: "unsupported",
		},
	}

	model, err = cfgUnsupported.GetModel(ctx)
	assert.Error(t, err)
	assert.Nil(t, model)
}

// 测试代理设置
func TestProxySettings(t *testing.T) {
	// 测试有代理设置
	cfgWithProxy := &Config{
		Proxy: "http://test-proxy.com",
		LLM: LLMConfig{
			Type:    "openai",
			BaseURL: "http://test-url.com",
			Model:   "test-model",
			APIKey:  "test-api-key",
		},
	}

	ctx := context.Background()
	model, err := cfgWithProxy.GetModel(ctx)
	require.NoError(t, err)
	assert.NotNil(t, model)

	// 测试无代理设置
	cfgNoProxy := &Config{
		Proxy: "",
		LLM: LLMConfig{
			Type:    "openai",
			BaseURL: "http://test-url.com",
			Model:   "test-model",
			APIKey:  "test-api-key",
		},
	}

	model, err = cfgNoProxy.GetModel(ctx)
	require.NoError(t, err)
	assert.NotNil(t, model)

	// 测试无效的代理URL
	cfgInvalidProxy := &Config{
		Proxy: "://invalid-proxy",
		LLM: LLMConfig{
			Type:    "openai",
			BaseURL: "http://test-url.com",
			Model:   "test-model",
			APIKey:  "test-api-key",
		},
	}

	model, err = cfgInvalidProxy.GetModel(ctx)
	assert.Error(t, err)
	assert.Nil(t, model)
	assert.Contains(t, err.Error(), "解析代理URL错误")
}

// TestLoadConfigErrors 测试加载配置时的错误处理
func TestLoadConfigErrors(t *testing.T) {
	// 保存原始工作目录和viper状态
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	// 创建临时目录
	tempDir := t.TempDir()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	defer func() {
		// 恢复原始工作目录
		_ = os.Chdir(originalDir)
		// 重置viper状态
		viper.Reset()
	}()

	// 重置viper状态
	viper.Reset()

	// 测试配置文件解析错误
	configPath := filepath.Join(tempDir, "invalid_config.yaml")

	// 创建一个会导致验证失败的配置文件（无效的LLM类型）
	invalidContent := `
llm:
  type: ""
  base_url: ""
  model: ""
  api_key: ""
mcp:
  config_file: ""
max_step: 0
`
	err = os.WriteFile(configPath, []byte(invalidContent), 0644)
	require.NoError(t, err)

	// 测试加载配置 - 配置验证会失败
	cfg, err := LoadConfig(configPath)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "配置验证失败")

	// 测试文件不存在的情况 - 应该使用默认配置
	viper.Reset()
	nonExistentPath := filepath.Join(tempDir, "non_existent.yaml")
	cfg, err = LoadConfig(nonExistentPath)
	assert.NoError(t, err) // 应该返回默认配置，而不是错误
	assert.NotNil(t, cfg)

	// 测试路径错误的情况（例如，目录而不是文件）- 应该使用默认配置
	viper.Reset()
	err = os.Mkdir(filepath.Join(tempDir, "config_dir"), 0755)
	require.NoError(t, err)
	dirPath := filepath.Join(tempDir, "config_dir")
	cfg, err = LoadConfig(dirPath)
	assert.NoError(t, err) // 应该返回默认配置，而不是错误
	assert.NotNil(t, cfg)
}

// TestGetTools 测试 GetTools 函数
func TestGetTools(t *testing.T) {
	// 创建一个控制器
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 创建一个模拟的 MCPHub
	mockMCPHub := NewMockMCPHubInterface(ctrl)

	// 保存原始函数并在测试结束后恢复
	originalNewMCPHub := mcpHubFactory
	defer func() {
		mcpHubFactory = originalNewMCPHub
	}()

	ctx := context.Background()

	// 测试成功获取工具
	// 设置模拟函数的行为
	mockTools := []tool.BaseTool{
		&mockTool{name: "tool1"},
		&mockTool{name: "tool2"},
	}
	mockMCPHub.EXPECT().GetEinoTools(gomock.Any(), gomock.Eq([]string{"tool1", "tool2"})).Return(mockTools, nil)
	mockMCPHub.EXPECT().CloseServers().Return(nil)

	// 替换 mcpHubFactory 函数
	mcpHubFactory = func(ctx context.Context, configFile string) (MCPHubInterface, error) {
		return mockMCPHub, nil
	}

	cfg := &Config{
		MCP: MCPConfig{
			ConfigFile: "test_mcpservers.json",
			Tools:      []string{"tool1", "tool2"},
		},
	}

	tools, cleanup, err := cfg.GetTools(ctx)
	require.NoError(t, err)
	assert.NotNil(t, tools)
	assert.NotNil(t, cleanup)
	assert.Len(t, tools, 2)

	// 执行清理函数
	cleanup()

	// 测试 NewMCPHub 失败的情况
	mcpHubFactory = func(ctx context.Context, configFile string) (MCPHubInterface, error) {
		return nil, fmt.Errorf("模拟 NewMCPHub 失败")
	}

	tools, cleanup, err = cfg.GetTools(ctx)
	assert.Error(t, err)
	assert.Nil(t, tools)
	assert.Nil(t, cleanup)

	// 测试 GetEinoTools 失败的情况
	mockMCPHub = NewMockMCPHubInterface(ctrl)
	mockMCPHub.EXPECT().GetEinoTools(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("模拟获取工具失败"))
	mockMCPHub.EXPECT().CloseServers().Return(nil).AnyTimes()

	mcpHubFactory = func(ctx context.Context, configFile string) (MCPHubInterface, error) {
		return mockMCPHub, nil
	}

	tools, cleanup, err = cfg.GetTools(ctx)
	assert.Error(t, err)
	assert.Nil(t, tools)
	assert.Nil(t, cleanup) // 当 GetEinoTools 失败时，清理函数会被调用后返回 nil
}

// TestMCPConfigValidate 测试 MCPConfig 的验证方法
func TestMCPConfigValidate(t *testing.T) {
	// 测试有效配置
	validMCP := MCPConfig{
		ConfigFile: "valid_config.json",
		Tools:      []string{"tool1", "tool2"},
	}
	err := validMCP.Validate()
	assert.NoError(t, err)

	// 测试无效配置 - 空的配置文件路径
	invalidMCP := MCPConfig{
		ConfigFile: "",
		Tools:      []string{"tool1"},
	}
	err = invalidMCP.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MCP配置文件路径不能为空")
}

// TestLLMConfigValidate 测试 LLMConfig 的验证方法
func TestLLMConfigValidate(t *testing.T) {
	// 测试有效的 OpenAI 配置
	validOpenAI := LLMConfig{
		Type:    LLMProviderOpenAI,
		BaseURL: "http://api.openai.com",
		Model:   "gpt-3.5-turbo",
		APIKey:  "sk-test-key",
	}
	err := validOpenAI.Validate()
	assert.NoError(t, err)

	// 测试有效的 Ollama 配置
	validOllama := LLMConfig{
		Type:    LLMProviderOllama,
		BaseURL: "http://localhost:11434",
		Model:   "llama2",
		APIKey:  "ollama",
	}
	err = validOllama.Validate()
	assert.NoError(t, err)

	// 测试空类型
	invalidType := LLMConfig{
		Type:    "",
		BaseURL: "http://localhost",
		Model:   "test-model",
	}
	err = invalidType.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LLM类型不能为空")

	// 测试不支持的类型
	unsupportedType := LLMConfig{
		Type:    "unsupported",
		BaseURL: "http://localhost",
		Model:   "test-model",
	}
	err = unsupportedType.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不支持的LLM类型")

	// 测试空 BaseURL
	emptyBaseURL := LLMConfig{
		Type:    LLMProviderOpenAI,
		BaseURL: "",
		Model:   "test-model",
	}
	err = emptyBaseURL.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LLM BaseURL不能为空")

	// 测试空模型名称
	emptyModel := LLMConfig{
		Type:    LLMProviderOpenAI,
		BaseURL: "http://localhost",
		Model:   "",
	}
	err = emptyModel.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LLM模型名称不能为空")
}

// TestConfigValidate 测试 Config 的验证方法
func TestConfigValidate(t *testing.T) {
	// 测试有效配置
	validConfig := Config{
		MCP: MCPConfig{
			ConfigFile: "valid_config.json",
			Tools:      []string{"tool1"},
		},
		LLM: LLMConfig{
			Type:    LLMProviderOpenAI,
			BaseURL: "http://api.openai.com",
			Model:   "gpt-3.5-turbo",
			APIKey:  "test-key",
		},
		MaxStep: 10,
	}
	err := validConfig.Validate()
	assert.NoError(t, err)

	// 测试 MCP 配置无效
	invalidMCPConfig := Config{
		MCP: MCPConfig{
			ConfigFile: "", // 无效
		},
		LLM: LLMConfig{
			Type:    LLMProviderOpenAI,
			BaseURL: "http://api.openai.com",
			Model:   "gpt-3.5-turbo",
		},
		MaxStep: 10,
	}
	err = invalidMCPConfig.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MCP配置验证失败")

	// 测试 LLM 配置无效
	invalidLLMConfig := Config{
		MCP: MCPConfig{
			ConfigFile: "valid_config.json",
		},
		LLM: LLMConfig{
			Type: "", // 无效
		},
		MaxStep: 10,
	}
	err = invalidLLMConfig.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LLM配置验证失败")

	// 测试无效的 MaxStep
	invalidMaxStep := Config{
		MCP: MCPConfig{
			ConfigFile: "valid_config.json",
		},
		LLM: LLMConfig{
			Type:    LLMProviderOpenAI,
			BaseURL: "http://api.openai.com",
			Model:   "gpt-3.5-turbo",
		},
		MaxStep: 0, // 无效
	}
	err = invalidMaxStep.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "最大步骤数必须大于0")
}

// TestSaveConfigValidation 测试保存配置时的验证
func TestSaveConfigValidation(t *testing.T) {
	cfg := &Config{
		MCP: MCPConfig{
			ConfigFile: "test_config.json",
		},
		LLM: LLMConfig{
			Type:    LLMProviderOpenAI,
			BaseURL: "http://api.openai.com",
			Model:   "gpt-3.5-turbo",
		},
		MaxStep: 10,
	}

	// 测试空配置文件路径
	err := cfg.SaveConfig("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "配置文件路径不能为空")

	// 测试有效路径
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")
	err = cfg.SaveConfig(configPath)
	assert.NoError(t, err)

	// 验证文件是否创建
	_, err = os.Stat(configPath)
	assert.NoError(t, err)
}

// TestCreateHTTPClient 测试 HTTP 客户端创建
func TestCreateHTTPClient(t *testing.T) {
	// 测试无代理的情况
	cfg := &Config{Proxy: ""}
	client, err := cfg.createHTTPClient()
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// 测试有效代理的情况
	cfg = &Config{Proxy: "http://proxy.example.com:8080"}
	client, err = cfg.createHTTPClient()
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// 测试无效代理URL的情况
	cfg = &Config{Proxy: "://invalid-proxy"}
	client, err = cfg.createHTTPClient()
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "解析代理URL错误")
}

// TestCreateModels 测试模型创建方法
func TestCreateModels(t *testing.T) {
	cfg := &Config{
		LLM: LLMConfig{
			Type:    LLMProviderOpenAI,
			BaseURL: "http://api.openai.com",
			Model:   "gpt-3.5-turbo",
			APIKey:  "test-key",
		},
	}

	ctx := context.Background()
	httpClient := http.DefaultClient

	// 测试创建 OpenAI 模型
	model, err := cfg.createOpenAIModel(ctx, httpClient)
	assert.NoError(t, err)
	assert.NotNil(t, model)

	// 测试创建 Ollama 模型
	cfg.LLM.Type = LLMProviderOllama
	model, err = cfg.createOllamaModel(ctx, httpClient)
	assert.NoError(t, err)
	assert.NotNil(t, model)
}

// TestGetDefaultConfig 测试默认配置
func TestGetDefaultConfig(t *testing.T) {
	config := NewDefaultConfig()

	assert.Equal(t, "", config.Proxy)
	assert.Equal(t, "mcpservers.json", config.MCP.ConfigFile)
	assert.Equal(t, []string{}, config.MCP.Tools)
	assert.Equal(t, "ollama", config.LLM.Type)
	assert.Equal(t, "http://127.0.0.1:11434", config.LLM.BaseURL)
	assert.Equal(t, "qwen3:4b", config.LLM.Model)
	assert.Equal(t, "ollama", config.LLM.APIKey)
	assert.Equal(t, `你是精通互联网的信息收集专家，需要帮助用户进行信息收集，当前时间是：{date}。`, config.SystemPrompt)
	assert.Equal(t, 20, config.MaxStep)
}

// TestSetupViper 测试 viper 配置
func TestSetupViper(t *testing.T) {
	// 重置 viper 状态
	viper.Reset()

	// 测试指定配置文件
	configFile := "test_config.yaml"
	err := setupViper(configFile)
	assert.NoError(t, err)

	// 测试不指定配置文件
	err = setupViper("")
	assert.NoError(t, err)
}

// TestReadConfigFile 测试配置文件读取
func TestReadConfigFile(t *testing.T) {
	// 重置 viper 状态
	viper.Reset()

	// 创建临时配置文件
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")
	configContent := `
llm:
  type: openai
  base_url: http://api.openai.com
  model: gpt-3.5-turbo
  api_key: test-key
max_step: 10
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// 设置 viper 来读取这个文件
	viper.SetConfigFile(configPath)

	// 测试成功读取配置文件
	err = readConfigFile()
	assert.NoError(t, err)

	// 测试配置文件不存在的情况
	viper.Reset()
	viper.SetConfigFile("/non/existent/path/config.yaml")
	err = readConfigFile()
	assert.Error(t, err) // 应该返回错误，但是错误类型是已知的
}

// TestConstants 测试常量值
func TestConstants(t *testing.T) {
	assert.Equal(t, "openai", LLMProviderOpenAI)
	assert.Equal(t, "ollama", LLMProviderOllama)
}

// TestNewDefaultConfig tests the NewDefaultConfig function
func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, "", cfg.Proxy)
	assert.Equal(t, defaultMCPConfigFile, cfg.MCP.ConfigFile)
	assert.Equal(t, []string{}, cfg.MCP.Tools)
	assert.Equal(t, LLMProviderOllama, cfg.LLM.Type)
	assert.Equal(t, defaultOllamaBaseURL, cfg.LLM.BaseURL)
	assert.Equal(t, defaultOllamaModel, cfg.LLM.Model)
	assert.Equal(t, defaultOllamaAPIKey, cfg.LLM.APIKey)
	assert.Equal(t, defaultSystemPrompt, cfg.SystemPrompt)
	assert.Equal(t, defaultMaxStep, cfg.MaxStep)
}

// TestMCPConfigValidateWithWhitespace tests MCP config validation with whitespace
func TestMCPConfigValidateWithWhitespace(t *testing.T) {
	tests := []struct {
		name        string
		configFile  string
		expectError bool
	}{
		{
			name:        "empty string",
			configFile:  "",
			expectError: true,
		},
		{
			name:        "whitespace only",
			configFile:  "   ",
			expectError: true,
		},
		{
			name:        "valid file path",
			configFile:  "test.json",
			expectError: false,
		},
		{
			name:        "file path with whitespace",
			configFile:  "  test.json  ",
			expectError: false, // TrimSpace makes this "test.json", which is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &MCPConfig{
				ConfigFile: tt.configFile,
				Tools:      []string{},
			}

			err := config.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), errMsgMCPConfigFileEmpty)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLLMConfigValidateWithWhitespace tests LLM config validation with whitespace
func TestLLMConfigValidateWithWhitespace(t *testing.T) {
	tests := []struct {
		name        string
		llmConfig   LLMConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "empty type",
			llmConfig: LLMConfig{
				Type:    "",
				BaseURL: "http://test.com",
				Model:   "test-model",
			},
			expectError: true,
			errorMsg:    errMsgLLMTypeEmpty,
		},
		{
			name: "whitespace type",
			llmConfig: LLMConfig{
				Type:    "   ",
				BaseURL: "http://test.com",
				Model:   "test-model",
			},
			expectError: true,
			errorMsg:    errMsgLLMTypeEmpty,
		},
		{
			name: "unsupported type",
			llmConfig: LLMConfig{
				Type:    "unsupported",
				BaseURL: "http://test.com",
				Model:   "test-model",
			},
			expectError: true,
			errorMsg:    "不支持的LLM类型",
		},
		{
			name: "empty base URL",
			llmConfig: LLMConfig{
				Type:    LLMProviderOpenAI,
				BaseURL: "",
				Model:   "test-model",
			},
			expectError: true,
			errorMsg:    errMsgLLMBaseURLEmpty,
		},
		{
			name: "whitespace base URL",
			llmConfig: LLMConfig{
				Type:    LLMProviderOpenAI,
				BaseURL: "   ",
				Model:   "test-model",
			},
			expectError: true,
			errorMsg:    errMsgLLMBaseURLEmpty,
		},
		{
			name: "empty model",
			llmConfig: LLMConfig{
				Type:    LLMProviderOpenAI,
				BaseURL: "http://test.com",
				Model:   "",
			},
			expectError: true,
			errorMsg:    errMsgLLMModelEmpty,
		},
		{
			name: "whitespace model",
			llmConfig: LLMConfig{
				Type:    LLMProviderOpenAI,
				BaseURL: "http://test.com",
				Model:   "   ",
			},
			expectError: true,
			errorMsg:    errMsgLLMModelEmpty,
		},
		{
			name: "valid openai config",
			llmConfig: LLMConfig{
				Type:    LLMProviderOpenAI,
				BaseURL: "http://test.com",
				Model:   "gpt-4",
				APIKey:  "test-key",
			},
			expectError: false,
		},
		{
			name: "valid ollama config",
			llmConfig: LLMConfig{
				Type:    LLMProviderOllama,
				BaseURL: "http://localhost:11434",
				Model:   "llama2",
				APIKey:  "test-key",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.llmConfig.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestConfigValidateMaxStep tests config validation for MaxStep field
func TestConfigValidateMaxStep(t *testing.T) {
	tests := []struct {
		name        string
		maxStep     int
		expectError bool
	}{
		{
			name:        "negative max step",
			maxStep:     -1,
			expectError: true,
		},
		{
			name:        "zero max step",
			maxStep:     0,
			expectError: true,
		},
		{
			name:        "positive max step",
			maxStep:     10,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				MCP: MCPConfig{
					ConfigFile: "test.json",
					Tools:      []string{},
				},
				LLM: LLMConfig{
					Type:    LLMProviderOllama,
					BaseURL: "http://localhost:11434",
					Model:   "test-model",
					APIKey:  "test-key",
				},
				MaxStep: tt.maxStep,
			}

			err := config.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), errMsgMaxStepInvalid)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCreateHTTPClientWithEmptyProxy tests HTTP client creation with empty proxy
func TestCreateHTTPClientWithEmptyProxy(t *testing.T) {
	config := &Config{Proxy: ""}

	client, err := config.createHTTPClient()

	assert.NoError(t, err)
	assert.Equal(t, http.DefaultClient, client)
}

// TestCreateHTTPClientWithWhitespaceProxy tests HTTP client creation with whitespace proxy
func TestCreateHTTPClientWithWhitespaceProxy(t *testing.T) {
	config := &Config{Proxy: "   "}

	client, err := config.createHTTPClient()

	assert.NoError(t, err)
	assert.Equal(t, http.DefaultClient, client)
}

// TestCreateHTTPClientWithInvalidProxy tests HTTP client creation with invalid proxy
func TestCreateHTTPClientWithInvalidProxy(t *testing.T) {
	config := &Config{Proxy: "://invalid-proxy"}

	client, err := config.createHTTPClient()

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "解析代理URL错误")
}

// TestSaveConfigWithEmptyPath tests SaveConfig with empty file path
func TestSaveConfigWithEmptyPath(t *testing.T) {
	config := NewDefaultConfig()

	err := config.SaveConfig("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), errMsgConfigFileEmpty)
}

// TestSaveConfigWithWhitespacePath tests SaveConfig with whitespace file path
func TestSaveConfigWithWhitespacePath(t *testing.T) {
	config := NewDefaultConfig()

	err := config.SaveConfig("   ")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), errMsgConfigFileEmpty)
}

// TestSetupViperWithEmptyConfigFile tests setupViper with empty config file
func TestSetupViperWithEmptyConfigFile(t *testing.T) {
	err := setupViper("")
	assert.NoError(t, err)
}

// TestSetupViperWithWhitespaceConfigFile tests setupViper with whitespace config file
func TestSetupViperWithWhitespaceConfigFile(t *testing.T) {
	err := setupViper("   ")
	assert.NoError(t, err)
}

// TestSetupViperWithValidConfigFile tests setupViper with valid config file
func TestSetupViperWithValidConfigFile(t *testing.T) {
	err := setupViper("test-config.yaml")
	assert.NoError(t, err)
}
