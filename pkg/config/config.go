// Package config provides comprehensive configuration management for the FOFA Logs AI application.
// It handles loading, parsing, validating, and managing application configuration including
// MCP servers, LLM models, proxy settings, and various runtime parameters.
//
// The package supports multiple configuration sources:
// - YAML configuration files
// - Environment variables (with MCPHOST prefix)
// - Command-line arguments (via external packages)
//
// Configuration validation ensures all required fields are present and valid
// before the application starts.
package config

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/LubyRuffy/fofalogsai/pkg/mcphost"
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/spf13/viper"
)

// LLM provider type constants
const (
	// LLMProviderOpenAI represents OpenAI-compatible LLM provider
	LLMProviderOpenAI = "openai"
	// LLMProviderOllama represents Ollama LLM provider
	LLMProviderOllama = "ollama"
)

// Default configuration values
const (
	defaultConfigName    = "config"
	defaultConfigType    = "yaml"
	defaultMCPConfigFile = "mcpservers.json"
	defaultOllamaBaseURL = "http://127.0.0.1:11434"
	defaultOllamaModel   = "qwen3:4b"
	defaultOllamaAPIKey  = "ollama"
	defaultMaxStep       = 20
	defaultSystemPrompt  = `你是精通互联网的信息收集专家，需要帮助用户进行信息收集，当前时间是：{date}。`
)

// Environment variable configuration
const (
	envPrefix = "MCPHOST"
)

// Error messages
const (
	errMsgMCPConfigFileEmpty = "MCP配置文件路径不能为空"
	errMsgLLMTypeEmpty       = "LLM类型不能为空"
	errMsgLLMTypeUnsupported = "不支持的LLM类型: %s"
	errMsgLLMBaseURLEmpty    = "LLM BaseURL不能为空"
	errMsgLLMModelEmpty      = "LLM模型名称不能为空"
	errMsgMaxStepInvalid     = "最大步骤数必须大于0"
	errMsgConfigFileEmpty    = "配置文件路径不能为空"
)

// MCPHubInterface defines the interface for MCP hub operations.
// This interface allows for dependency injection during testing.
type MCPHubInterface interface {
	GetEinoTools(ctx context.Context, toolNameList []string) ([]tool.BaseTool, error)
	CloseServers() error
}

// mcpHubFactory is a factory function for creating MCPHub instances.
// This variable allows for dependency injection during testing.
var mcpHubFactory = func(ctx context.Context, configFile string) (MCPHubInterface, error) {
	return mcphost.NewMCPHub(ctx, configFile)
}

// MCPConfig represents MCP (Model Context Protocol) server configuration settings.
// It contains the path to MCP server configuration file and the list of tools to use.
type MCPConfig struct {
	ConfigFile string   `mapstructure:"config_file" json:"config_file" yaml:"config_file"` // MCP服务器配置文件路径
	Tools      []string `mapstructure:"tools" json:"tools" yaml:"tools"`                   // 工具列表
}

// Validate validates the MCP configuration.
// It ensures that the configuration file path is not empty.
// Note: File existence is not checked here as it's more appropriate to check at runtime.
func (m *MCPConfig) Validate() error {
	if strings.TrimSpace(m.ConfigFile) == "" {
		return errors.New(errMsgMCPConfigFileEmpty)
	}
	return nil
}

// LLMConfig represents Large Language Model configuration settings.
// It supports both OpenAI-compatible and Ollama providers with their respective settings.
type LLMConfig struct {
	Type    string `mapstructure:"type" json:"type" yaml:"type"`             // 大模型类型，openai 或 ollama
	BaseURL string `mapstructure:"base_url" json:"base_url" yaml:"base_url"` // 大模型API基础URL
	Model   string `mapstructure:"model" json:"model" yaml:"model"`          // 大模型名称
	APIKey  string `mapstructure:"api_key" json:"api_key" yaml:"api_key"`    // 大模型API密钥
}

// Validate validates the LLM configuration.
// It ensures all required fields are present and the LLM type is supported.
func (l *LLMConfig) Validate() error {
	if strings.TrimSpace(l.Type) == "" {
		return errors.New(errMsgLLMTypeEmpty)
	}
	if l.Type != LLMProviderOpenAI && l.Type != LLMProviderOllama {
		return fmt.Errorf(errMsgLLMTypeUnsupported, l.Type)
	}
	if strings.TrimSpace(l.BaseURL) == "" {
		return errors.New(errMsgLLMBaseURLEmpty)
	}
	if strings.TrimSpace(l.Model) == "" {
		return errors.New(errMsgLLMModelEmpty)
	}
	return nil
}

// Config represents the main application configuration structure.
// It contains all settings needed to run the FOFA Logs AI application including proxy settings,
// MCP server configuration, LLM settings, and runtime parameters.
type Config struct {
	Proxy        string    `mapstructure:"proxy" json:"proxy" yaml:"proxy"`                         // 代理配置，用于调试查看大模型的请求和响应
	MCP          MCPConfig `mapstructure:"mcp" json:"mcp" yaml:"mcp"`                               // MCP服务器配置
	LLM          LLMConfig `mapstructure:"llm" json:"llm" yaml:"llm"`                               // 大模型配置
	SystemPrompt string    `mapstructure:"system_prompt" json:"system_prompt" yaml:"system_prompt"` // 系统提示词
	MaxStep      int       `mapstructure:"max_step" json:"max_step" yaml:"max_step"`                // 最大思考步骤数
}

// Validate validates the entire configuration.
// It performs comprehensive validation of all configuration sections.
func (c *Config) Validate() error {
	if err := c.MCP.Validate(); err != nil {
		return fmt.Errorf("MCP配置验证失败: %w", err)
	}
	if err := c.LLM.Validate(); err != nil {
		return fmt.Errorf("LLM配置验证失败: %w", err)
	}
	if c.MaxStep <= 0 {
		return errors.New(errMsgMaxStepInvalid)
	}
	return nil
}

// GetModel creates and returns a configured LLM model instance.
// It supports both OpenAI-compatible and Ollama providers.
// The method configures HTTP client with proxy if specified.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - model.ToolCallingChatModel: Configured model instance
//   - error: Error if model creation fails
func (c *Config) GetModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	httpClient, err := c.createHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("创建HTTP客户端失败: %w", err)
	}

	switch c.LLM.Type {
	case LLMProviderOpenAI:
		return c.createOpenAIModel(ctx, httpClient)
	case LLMProviderOllama:
		return c.createOllamaModel(ctx, httpClient)
	default:
		return nil, fmt.Errorf(errMsgLLMTypeUnsupported, c.LLM.Type)
	}
}

// createHTTPClient creates an HTTP client with optional proxy configuration.
// If proxy is configured, it creates a client with proxy transport.
func (c *Config) createHTTPClient() (*http.Client, error) {
	if strings.TrimSpace(c.Proxy) == "" {
		return http.DefaultClient, nil
	}

	proxyURL, err := url.Parse(c.Proxy)
	if err != nil {
		return nil, fmt.Errorf("解析代理URL错误: %w", err)
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}, nil
}

// createOpenAIModel creates an OpenAI-compatible model instance.
func (c *Config) createOpenAIModel(ctx context.Context, httpClient *http.Client) (model.ToolCallingChatModel, error) {
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:    c.LLM.BaseURL,
		Model:      c.LLM.Model,
		APIKey:     c.LLM.APIKey,
		HTTPClient: httpClient,
	})
}

// createOllamaModel creates an Ollama model instance.
func (c *Config) createOllamaModel(ctx context.Context, httpClient *http.Client) (model.ToolCallingChatModel, error) {
	return ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL:    c.LLM.BaseURL,
		Model:      c.LLM.Model,
		HTTPClient: httpClient,
	})
}

// GetTools connects to MCP servers and retrieves the configured tools.
// It returns the tools along with a cleanup function that should be called
// when the tools are no longer needed.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - []tool.BaseTool: List of available tools
//   - func(): Cleanup function to close MCP connections
//   - error: Error if tool retrieval fails
func (c *Config) GetTools(ctx context.Context) ([]tool.BaseTool, func(), error) {
	// 连接mcp服务器
	mcpHub, err := mcpHubFactory(ctx, c.MCP.ConfigFile)
	if err != nil {
		return nil, nil, fmt.Errorf("连接MCP服务器失败: %w", err)
	}

	cleanupFunc := func() {
		if closeErr := mcpHub.CloseServers(); closeErr != nil {
			log.Printf("关闭MCP服务器失败: %v", closeErr)
		}
	}

	// Initialize the required tools
	einoTools, err := mcpHub.GetEinoTools(ctx, c.MCP.Tools)
	if err != nil {
		cleanupFunc()
		return nil, nil, fmt.Errorf("获取MCP工具失败: %w", err)
	}

	return einoTools, cleanupFunc, nil
}

// SaveConfig saves the current configuration to the specified file.
// It uses viper to serialize the configuration to YAML format.
//
// Parameters:
//   - configFile: Path to the configuration file to save
//
// Returns:
//   - error: Error if saving fails
func (c *Config) SaveConfig(configFile string) error {
	if strings.TrimSpace(configFile) == "" {
		return errors.New(errMsgConfigFileEmpty)
	}

	// 将配置结构体的值设置到viper中
	c.setViperValues()

	if err := viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	return nil
}

// setViperValues sets configuration values in viper for serialization.
func (c *Config) setViperValues() {
	viper.Set("proxy", c.Proxy)
	viper.Set("mcp.config_file", c.MCP.ConfigFile)
	viper.Set("mcp.tools", c.MCP.Tools)
	viper.Set("llm.type", c.LLM.Type)
	viper.Set("llm.base_url", c.LLM.BaseURL)
	viper.Set("llm.model", c.LLM.Model)
	viper.Set("llm.api_key", c.LLM.APIKey)
	viper.Set("system_prompt", c.SystemPrompt)
	viper.Set("max_step", c.MaxStep)
}

// NewDefaultConfig returns a default configuration with sensible defaults.
// This function creates a configuration that can be used as a starting point
// or fallback when no configuration file is available.
func NewDefaultConfig() *Config {
	return &Config{
		Proxy: "",
		MCP: MCPConfig{
			ConfigFile: defaultMCPConfigFile,
			Tools:      []string{},
		},
		LLM: LLMConfig{
			Type:    LLMProviderOllama,
			BaseURL: defaultOllamaBaseURL,
			Model:   defaultOllamaModel,
			APIKey:  defaultOllamaAPIKey,
		},
		SystemPrompt: defaultSystemPrompt,
		MaxStep:      defaultMaxStep,
	}
}

// LoadConfig loads configuration from file or creates default configuration.
// It supports both specified config file path and automatic discovery of config files.
// The function follows this priority order:
// 1. Specified config file
// 2. Environment variables
// 3. Default configuration
//
// Parameters:
//   - configFile: Path to configuration file. If empty, will search for config file automatically
//
// Returns:
//   - *Config: Loaded configuration
//   - error: Error if loading fails
func LoadConfig(configFile string) (*Config, error) {
	config := NewDefaultConfig()

	if err := setupViper(configFile); err != nil {
		return nil, fmt.Errorf("设置viper失败: %w", err)
	}

	if err := readConfigFile(); err != nil {
		log.Println("未找到配置文件，使用默认配置")
	}

	// 将配置文件内容解析到结构体
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("解析配置文件错误: %w", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return config, nil
}

// setupViper configures viper for config file reading.
// It sets up file paths, environment variable handling, and other viper settings.
func setupViper(configFile string) error {
	if strings.TrimSpace(configFile) != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName(defaultConfigName)
		viper.SetConfigType(defaultConfigType)

		// 设置查找配置文件的路径
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("$HOME/.mcphost")
	}

	// 读取环境变量
	viper.AutomaticEnv()
	viper.SetEnvPrefix(envPrefix)

	return nil
}

// readConfigFile attempts to read the configuration file.
// It handles common file reading errors gracefully.
func readConfigFile() error {
	if err := viper.ReadInConfig(); err != nil {
		// 如果找不到配置文件，返回错误
		var configFileNotFoundError viper.ConfigFileNotFoundError
		var pathError *fs.PathError
		if !errors.As(err, &configFileNotFoundError) && !errors.As(err, &pathError) {
			return fmt.Errorf("读取配置文件错误: %w", err)
		}
		return err
	}
	return nil
}
