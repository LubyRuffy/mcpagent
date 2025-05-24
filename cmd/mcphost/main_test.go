package main

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/LubyRuffy/mcpagent/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseToolsList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "空字符串",
			input:    "",
			expected: []string{},
		},
		{
			name:     "单个工具",
			input:    "tool1",
			expected: []string{"tool1"},
		},
		{
			name:     "多个工具",
			input:    "tool1,tool2,tool3",
			expected: []string{"tool1", "tool2", "tool3"},
		},
		{
			name:     "带空格的工具",
			input:    " tool1 , tool2 , tool3 ",
			expected: []string{"tool1", "tool2", "tool3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseToolsList(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateTask(t *testing.T) {
	tests := []struct {
		name      string
		task      string
		expectErr bool
	}{
		{
			name:      "有效任务",
			task:      "测试任务",
			expectErr: false,
		},
		{
			name:      "空任务",
			task:      "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTask(tt.task)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "请使用 -task 参数指定要执行的任务")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMergeCommandLineArgs(t *testing.T) {
	// 创建基础配置
	cfg := &config.Config{
		Proxy: "original-proxy",
		MCP: config.MCPConfig{
			ConfigFile: "original-config.json",
			Tools:      []string{"original-tool"},
		},
		LLM: config.LLMConfig{
			Type:    "original-type",
			BaseURL: "original-url",
			Model:   "original-model",
			APIKey:  "original-key",
		},
		SystemPrompt: "original-prompt",
		MaxStep:      10,
	}

	// 创建命令行参数
	proxy := "new-proxy"
	mcpConfigFile := "new-config.json"
	tools := "new-tool1,new-tool2"
	llmType := "new-type"
	llmBaseURL := "new-url"
	llmModel := "new-model"
	llmAPIKey := "new-key"
	systemPrompt := "new-prompt"
	maxStep := 20

	args := &CommandLineArgs{
		Proxy:         &proxy,
		MCPConfigFile: &mcpConfigFile,
		MCPTools:      &tools,
		LLMType:       &llmType,
		LLMBaseURL:    &llmBaseURL,
		LLMModel:      &llmModel,
		LLMAPIKey:     &llmAPIKey,
		SystemPrompt:  &systemPrompt,
		MaxStep:       &maxStep,
	}

	// 合并配置
	mergeCommandLineArgs(cfg, args)

	// 验证结果
	assert.Equal(t, "new-proxy", cfg.Proxy)
	assert.Equal(t, "new-config.json", cfg.MCP.ConfigFile)
	assert.Equal(t, []string{"new-tool1", "new-tool2"}, cfg.MCP.Tools)
	assert.Equal(t, "new-type", cfg.LLM.Type)
	assert.Equal(t, "new-url", cfg.LLM.BaseURL)
	assert.Equal(t, "new-model", cfg.LLM.Model)
	assert.Equal(t, "new-key", cfg.LLM.APIKey)
	assert.Equal(t, "new-prompt", cfg.SystemPrompt)
	assert.Equal(t, 20, cfg.MaxStep)
}

func TestMergeCommandLineArgsWithEmptyValues(t *testing.T) {
	// 创建基础配置
	cfg := &config.Config{
		Proxy: "original-proxy",
		MCP: config.MCPConfig{
			ConfigFile: "original-config.json",
			Tools:      []string{"original-tool"},
		},
		LLM: config.LLMConfig{
			Type:    "original-type",
			BaseURL: "original-url",
			Model:   "original-model",
			APIKey:  "original-key",
		},
		SystemPrompt: "original-prompt",
		MaxStep:      10,
	}

	// 创建空的命令行参数
	proxy := ""
	mcpConfigFile := ""
	tools := ""
	llmType := ""
	llmBaseURL := ""
	llmModel := ""
	llmAPIKey := ""
	systemPrompt := ""
	maxStep := 0

	args := &CommandLineArgs{
		Proxy:         &proxy,
		MCPConfigFile: &mcpConfigFile,
		MCPTools:      &tools,
		LLMType:       &llmType,
		LLMBaseURL:    &llmBaseURL,
		LLMModel:      &llmModel,
		LLMAPIKey:     &llmAPIKey,
		SystemPrompt:  &systemPrompt,
		MaxStep:       &maxStep,
	}

	// 合并配置
	mergeCommandLineArgs(cfg, args)

	// 验证原始值保持不变
	assert.Equal(t, "original-proxy", cfg.Proxy)
	assert.Equal(t, "original-config.json", cfg.MCP.ConfigFile)
	assert.Equal(t, []string{"original-tool"}, cfg.MCP.Tools)
	assert.Equal(t, "original-type", cfg.LLM.Type)
	assert.Equal(t, "original-url", cfg.LLM.BaseURL)
	assert.Equal(t, "original-model", cfg.LLM.Model)
	assert.Equal(t, "original-key", cfg.LLM.APIKey)
	assert.Equal(t, "original-prompt", cfg.SystemPrompt)
	assert.Equal(t, 10, cfg.MaxStep)
}

func TestSaveConfigIfNeeded(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	configFile := tempDir + "/test_config.yaml"

	// 创建测试配置
	cfg := &config.Config{
		Proxy: "test-proxy",
		MCP: config.MCPConfig{
			ConfigFile: "test-mcp.json",
			Tools:      []string{"test-tool"},
		},
		LLM: config.LLMConfig{
			Type:    "ollama",
			BaseURL: "http://localhost:11434",
			Model:   "test-model",
			APIKey:  "test-key",
		},
		SystemPrompt: "test prompt",
		MaxStep:      15,
	}

	// 测试保存配置
	err := saveConfigIfNeeded(cfg, configFile)
	assert.NoError(t, err)

	// 验证文件是否存在
	_, err = os.Stat(configFile)
	assert.NoError(t, err)

	// 测试空配置文件路径
	err = saveConfigIfNeeded(cfg, "")
	assert.NoError(t, err)
}

func TestLoadAndMergeConfig(t *testing.T) {
	// 创建临时配置文件
	tempDir := t.TempDir()
	configFile := tempDir + "/test_config.yaml"

	// 创建有效的配置内容
	configContent := `
proxy: ""
mcp:
  config_file: "mcpservers.json"
  tools: []
llm:
  type: "ollama"
  base_url: "http://127.0.0.1:11434"
  model: "qwen3:4b"
  api_key: "ollama"
system_prompt: "你是精通互联网的信息收集专家，需要帮助用户进行信息收集，当前时间是：{date}。"
max_step: 20
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// 创建命令行参数
	proxy := "test-proxy"
	args := &CommandLineArgs{
		ConfigFile:    &configFile,
		Proxy:         &proxy,
		MCPConfigFile: new(string),
		MCPTools:      new(string),
		LLMType:       new(string),
		LLMBaseURL:    new(string),
		LLMModel:      new(string),
		LLMAPIKey:     new(string),
		SystemPrompt:  new(string),
		MaxStep:       new(int),
	}

	// 测试加载和合并配置
	cfg, err := loadAndMergeConfig(args)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "test-proxy", cfg.Proxy) // 命令行参数应该覆盖配置文件
}

func TestLoadAndMergeConfigWithInvalidFile(t *testing.T) {
	// 测试不存在的配置文件
	nonExistentFile := "/non/existent/config.yaml"
	args := &CommandLineArgs{
		ConfigFile:    &nonExistentFile,
		Proxy:         new(string),
		MCPConfigFile: new(string),
		MCPTools:      new(string),
		LLMType:       new(string),
		LLMBaseURL:    new(string),
		LLMModel:      new(string),
		LLMAPIKey:     new(string),
		SystemPrompt:  new(string),
		MaxStep:       new(int),
	}

	// 应该使用默认配置，不会出错
	cfg, err := loadAndMergeConfig(args)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestSetupSignalHandling(t *testing.T) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 设置信号处理
	setupSignalHandling(cancel)

	// 验证上下文仍然有效
	select {
	case <-ctx.Done():
		t.Fatal("上下文不应该被取消")
	default:
		// 正常情况
	}

	// 注意：实际的信号测试比较复杂，这里只测试函数不会panic
}

func TestConstants(t *testing.T) {
	assert.Equal(t, 0, ExitCodeSuccess)
	assert.Equal(t, 1, ExitCodeError)
	assert.Equal(t, ",", defaultToolsSeparator)
}

// 测试命令行参数解析的辅助函数
func TestParseCommandLineArgsStructure(t *testing.T) {
	// 重置flag包的状态
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// 模拟命令行参数
	os.Args = []string{"mcphost", "-task", "test task", "-proxy", "http://proxy.com"}

	args := parseCommandLineArgs()

	// 验证结构体字段不为nil
	assert.NotNil(t, args.ConfigFile)
	assert.NotNil(t, args.Proxy)
	assert.NotNil(t, args.MCPConfigFile)
	assert.NotNil(t, args.LLMType)
	assert.NotNil(t, args.LLMBaseURL)
	assert.NotNil(t, args.LLMModel)
	assert.NotNil(t, args.LLMAPIKey)
	assert.NotNil(t, args.SystemPrompt)
	assert.NotNil(t, args.MCPTools)
	assert.NotNil(t, args.MaxStep)
	assert.NotNil(t, args.Task)

	// 验证解析的值
	assert.Equal(t, "test task", *args.Task)
	assert.Equal(t, "http://proxy.com", *args.Proxy)
}
