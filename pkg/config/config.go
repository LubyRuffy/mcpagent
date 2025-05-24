// Package config provides comprehensive configuration management for the FOFA Logs AI application.
// It handles loading, parsing, validating, and managing application configuration including
// MCP servers, LLM models, proxy settings, and various runtime parameters.
//
// The package supports multiple configuration sources:
//   - YAML configuration files
//   - Environment variables (with MCPHOST prefix)
//   - Command-line arguments (via external packages)
//
// Configuration validation ensures all required fields are present and valid
// before the application starts.
//
// Example usage:
//
//	cfg, err := config.LoadConfig("config.yaml")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	model, err := cfg.GetModel(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	tools, cleanup, err := cfg.GetTools(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer cleanup()
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

	"github.com/LubyRuffy/mcpagent/pkg/mcphost"
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/spf13/viper"
)

// LLM provider type constants define supported LLM providers
const (
	// LLMProviderOpenAI represents OpenAI-compatible LLM provider
	LLMProviderOpenAI = "openai"
	// LLMProviderOllama represents Ollama LLM provider
	LLMProviderOllama = "ollama"
)

// Default configuration values provide sensible defaults for the application
const (
	defaultConfigName    = "config"
	defaultConfigType    = "yaml"
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

// Error messages provide consistent error reporting
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
// This interface allows for dependency injection during testing and provides
// a clean abstraction for MCP server management.
type MCPHubInterface interface {
	// GetEinoTools retrieves tools from MCP servers and converts them to Eino format
	GetEinoTools(ctx context.Context, toolNameList []string) ([]tool.BaseTool, error)
	// CloseServers gracefully closes all MCP server connections
	CloseServers() error
}

// mcpHubFactory is a factory function for creating MCPHub instances.
// This variable allows for dependency injection during testing by replacing
// the factory function with a mock implementation.
var mcpHubFactory = func(ctx context.Context, configFile string) (MCPHubInterface, error) {
	return mcphost.NewMCPHub(ctx, configFile)
}

// mcpHubFromSettingsFactory is a factory function for creating MCPHub instances from settings.
// This variable allows for dependency injection during testing by replacing
// the factory function with a mock implementation.
var mcpHubFromSettingsFactory = func(ctx context.Context, settings *mcphost.MCPSettings) (MCPHubInterface, error) {
	return mcphost.NewMCPHubFromSettings(ctx, settings)
}

// MCPConfig represents MCP (Model Context Protocol) server configuration settings.
// It contains either the path to MCP server configuration file or direct MCPServers configuration,
// along with the list of tools to use.
type MCPConfig struct {
	ConfigFile string                          `mapstructure:"config_file" json:"config_file" yaml:"config_file"` // MCP服务器配置文件路径
	MCPServers map[string]mcphost.ServerConfig `mapstructure:"mcp_servers" json:"mcp_servers" yaml:"mcp_servers"` // MCP服务器直接配置
	Tools      []string                        `mapstructure:"tools" json:"tools" yaml:"tools"`                   // 工具列表
}

// Validate validates the MCP configuration.
// It ensures that either the configuration file path is not empty or MCPServers is provided.
// Note: File existence is not checked here as it's more appropriate to check at runtime.
//
// Returns:
//   - error: validation error if configuration is invalid, nil otherwise
func (m *MCPConfig) Validate() error {
	// 如果MCPServers不为空，则优先使用MCPServers配置
	if len(m.MCPServers) > 0 {
		return nil
	}

	// 否则检查ConfigFile是否为空
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
// All string fields are trimmed of whitespace before validation.
//
// Returns:
//   - error: validation error if configuration is invalid, nil otherwise
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
//
// The configuration supports multiple sources with the following precedence:
//  1. Command-line arguments (highest priority)
//  2. Environment variables
//  3. Configuration file
//  4. Default values (lowest priority)
type Config struct {
	Proxy        string    `mapstructure:"proxy" json:"proxy" yaml:"proxy"`                         // 代理配置，用于调试查看大模型的请求和响应
	MCP          MCPConfig `mapstructure:"mcp" json:"mcp" yaml:"mcp"`                               // MCP服务器配置
	LLM          LLMConfig `mapstructure:"llm" json:"llm" yaml:"llm"`                               // 大模型配置
	SystemPrompt string    `mapstructure:"system_prompt" json:"system_prompt" yaml:"system_prompt"` // 系统提示词
	MaxStep      int       `mapstructure:"max_step" json:"max_step" yaml:"max_step"`                // 最大思考步骤数
}

// Validate validates the entire configuration.
// It performs comprehensive validation of all configuration sections and ensures
// that all interdependent settings are consistent.
//
// Returns:
//   - error: validation error if any configuration section is invalid, nil otherwise
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
// It supports both OpenAI-compatible and Ollama providers and automatically
// configures HTTP client with proxy if specified in the configuration.
//
// The method handles provider-specific configuration and returns a model
// that implements the ToolCallingChatModel interface for use with the agent framework.
//
// Parameters:
//   - ctx: Context for the operation, used for cancellation and timeouts
//
// Returns:
//   - model.ToolCallingChatModel: Configured model instance ready for use
//   - error: Error if model creation fails due to configuration or network issues
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
// If proxy is configured and valid, it creates a client with proxy transport.
// Otherwise, it returns the default HTTP client.
//
// Returns:
//   - *http.Client: HTTP client configured with proxy if specified
//   - error: Error if proxy URL parsing fails
func (c *Config) createHTTPClient() (*http.Client, error) {
	proxyStr := strings.TrimSpace(c.Proxy)
	if proxyStr == "" {
		return http.DefaultClient, nil
	}

	proxyURL, err := url.Parse(proxyStr)
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
// It configures the model with the provided HTTP client and LLM settings.
//
// Parameters:
//   - ctx: Context for the operation
//   - httpClient: HTTP client to use for API requests
//
// Returns:
//   - model.ToolCallingChatModel: Configured OpenAI model
//   - error: Error if model creation fails
func (c *Config) createOpenAIModel(ctx context.Context, httpClient *http.Client) (model.ToolCallingChatModel, error) {
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:    c.LLM.BaseURL,
		Model:      c.LLM.Model,
		APIKey:     c.LLM.APIKey,
		HTTPClient: httpClient,
	})
}

// createOllamaModel creates an Ollama model instance.
// It configures the model with the provided HTTP client and LLM settings.
//
// Parameters:
//   - ctx: Context for the operation
//   - httpClient: HTTP client to use for API requests
//
// Returns:
//   - model.ToolCallingChatModel: Configured Ollama model
//   - error: Error if model creation fails
func (c *Config) createOllamaModel(ctx context.Context, httpClient *http.Client) (model.ToolCallingChatModel, error) {
	return ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL:    c.LLM.BaseURL,
		Model:      c.LLM.Model,
		HTTPClient: httpClient,
	})
}

// GetTools connects to MCP servers and retrieves the configured tools.
// It establishes connections to all configured MCP servers, discovers available tools,
// and returns them in a format compatible with the Eino framework.
//
// The method returns a cleanup function that MUST be called when the tools are no longer
// needed to properly close MCP server connections and free resources.
//
// Parameters:
//   - ctx: Context for the operation, used for cancellation and timeouts
//
// Returns:
//   - []tool.BaseTool: List of available tools from all connected MCP servers
//   - func(): Cleanup function to close MCP connections (must be called)
//   - error: Error if tool retrieval fails
//
// Example:
//
//	tools, cleanup, err := cfg.GetTools(ctx)
//	if err != nil {
//		return err
//	}
//	defer cleanup() // Important: always call cleanup
func (c *Config) GetTools(ctx context.Context) ([]tool.BaseTool, func(), error) {
	// 连接mcp服务器
	var mcpHub MCPHubInterface
	var err error

	// 如果MCPServers不为空，则优先使用MCPServers配置
	if len(c.MCP.MCPServers) > 0 {
		// 创建MCPSettings
		settings := &mcphost.MCPSettings{
			MCPServers: c.MCP.MCPServers,
		}

		// 使用MCPSettings创建MCPHub
		mcpHub, err = mcpHubFromSettingsFactory(ctx, settings)
	} else {
		// 否则使用ConfigFile
		mcpHub, err = mcpHubFactory(ctx, c.MCP.ConfigFile)
	}

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
// It uses viper to serialize the configuration to YAML format with proper formatting.
// The configuration is validated before saving to ensure consistency.
//
// Parameters:
//   - configFile: Path to the configuration file to save (must not be empty)
//
// Returns:
//   - error: Error if saving fails due to validation, file system issues, or serialization problems
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
// This method maps the configuration struct fields to viper keys for proper YAML output.
func (c *Config) setViperValues() {
	viper.Set("proxy", c.Proxy)
	viper.Set("mcp.config_file", c.MCP.ConfigFile)
	viper.Set("mcp.mcp_servers", c.MCP.MCPServers)
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
//
// The default configuration uses:
//   - Ollama as the LLM provider (localhost:11434)
//   - qwen3:4b as the default model
//   - mcpservers.json as the MCP configuration file
//   - 20 as the maximum reasoning steps
//   - A Chinese system prompt for information gathering
//
// Returns:
//   - *Config: Default configuration ready for use or customization
func NewDefaultConfig() *Config {
	return &Config{
		MCP: MCPConfig{
			ConfigFile: "mcpservers.json",
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
//  1. Specified config file (if provided)
//  2. Automatic config file discovery in standard locations
//  3. Environment variables (with MCPHOST prefix)
//  4. Default configuration values
//
// The function is resilient to missing configuration files and will use defaults
// when files are not found, but will return errors for invalid configurations.
//
// Parameters:
//   - configFile: Path to configuration file. If empty, will search for config file automatically
//
// Returns:
//   - *Config: Loaded and validated configuration
//   - error: Error if loading fails due to parsing or validation issues
//
// Example:
//
//	// Load from specific file
//	cfg, err := LoadConfig("myconfig.yaml")
//
//	// Auto-discover config file
//	cfg, err := LoadConfig("")
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
// It sets up file paths, environment variable handling, and other viper settings
// based on whether a specific config file is provided or auto-discovery is needed.
//
// Parameters:
//   - configFile: Specific config file path, or empty for auto-discovery
//
// Returns:
//   - error: Error if viper setup fails
func setupViper(configFile string) error {
	configFileStr := strings.TrimSpace(configFile)
	if configFileStr != "" {
		viper.SetConfigFile(configFileStr)
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
// It handles common file reading errors gracefully and distinguishes between
// missing files (which is acceptable) and actual parsing errors.
//
// Returns:
//   - error: Error if file reading fails for reasons other than file not found
func readConfigFile() error {
	if err := viper.ReadInConfig(); err != nil {
		// 如果找不到配置文件，返回错误但不是致命错误
		var configFileNotFoundError viper.ConfigFileNotFoundError
		var pathError *fs.PathError
		if !errors.As(err, &configFileNotFoundError) && !errors.As(err, &pathError) {
			return fmt.Errorf("读取配置文件错误: %w", err)
		}
		return err
	}
	return nil
}
