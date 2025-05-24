// Package config provides configuration management functionality for MCP Agent.
// It handles loading, parsing, and managing application configuration including
// MCP servers, LLM models, and various runtime settings.
package config

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"

	"github.com/LubyRuffy/mcpagent/pkg/mcphost"
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/spf13/viper"
)

// 常量定义
const (
	// LLMTypeOpenAI represents OpenAI-compatible LLM provider
	LLMTypeOpenAI = "openai"
	// LLMTypeOllama represents Ollama LLM provider
	LLMTypeOllama = "ollama"

	// 默认配置值
	defaultConfigName    = "config"
	defaultConfigType    = "yaml"
	defaultMCPConfigFile = "mcpservers.json"
	defaultOllamaBaseURL = "http://127.0.0.1:11434"
	defaultOllamaModel   = "qwen3:4b"
	defaultOllamaAPIKey  = "ollama"
	defaultMaxStep       = 20
	defaultSystemPrompt  = `你是精通互联网的信息收集专家，需要帮助用户进行信息收集，当前时间是：{date}。`

	// 环境变量前缀
	envPrefix = "MCPHOST"
)

// 定义一个变量来存储 mcphost.NewMCPHub 函数，便于测试时进行模拟
var mcphostNewMCPHub = func(ctx context.Context, configFile string) (MCPHubInterface, error) {
	return mcphost.NewMCPHub(ctx, configFile)
}

// MCPConfig represents MCP server configuration settings.
// It contains the path to MCP server configuration file and the list of tools to use.
type MCPConfig struct {
	ConfigFile string   `mapstructure:"config_file" json:"config_file"` // MCP服务器配置文件路径
	Tools      []string `mapstructure:"tools" json:"tools"`             // 工具列表
}

// Validate validates the MCP configuration
func (m *MCPConfig) Validate() error {
	if m.ConfigFile == "" {
		return errors.New("MCP配置文件路径不能为空")
	}
	// 注意：我们不检查文件是否实际存在，因为这在运行时检查更合适
	return nil
}

// LLMConfig represents Large Language Model configuration settings.
// It supports both OpenAI-compatible and Ollama providers.
type LLMConfig struct {
	Type    string `mapstructure:"type" json:"type"`         // 大模型类型，openai 或 ollama
	BaseURL string `mapstructure:"base_url" json:"base_url"` // 大模型API基础URL
	Model   string `mapstructure:"model" json:"model"`       // 大模型名称
	APIKey  string `mapstructure:"api_key" json:"api_key"`   // 大模型API密钥
}

// Validate validates the LLM configuration
func (l *LLMConfig) Validate() error {
	if l.Type == "" {
		return errors.New("LLM类型不能为空")
	}
	if l.Type != LLMTypeOpenAI && l.Type != LLMTypeOllama {
		return fmt.Errorf("不支持的LLM类型: %s", l.Type)
	}
	if l.BaseURL == "" {
		return errors.New("LLM BaseURL不能为空")
	}
	if l.Model == "" {
		return errors.New("LLM模型名称不能为空")
	}
	return nil
}

// Config represents the main application configuration structure.
// It contains all settings needed to run the MCP Agent including proxy settings,
// MCP server configuration, LLM settings, and runtime parameters.
type Config struct {
	Proxy        string    `mapstructure:"proxy" json:"proxy"`                 // 代理配置，用于调试查看大模型的请求和响应
	MCP          MCPConfig `mapstructure:"mcp" json:"mcp"`                     // MCP服务器配置
	LLM          LLMConfig `mapstructure:"llm" json:"llm"`                     // 大模型配置
	SystemPrompt string    `mapstructure:"system_prompt" json:"system_prompt"` // 系统提示词
	MaxStep      int       `mapstructure:"max_step" json:"max_step"`           // 最大思考步骤数
}

// Validate validates the entire configuration
func (c *Config) Validate() error {
	if err := c.MCP.Validate(); err != nil {
		return fmt.Errorf("MCP配置验证失败: %w", err)
	}
	if err := c.LLM.Validate(); err != nil {
		return fmt.Errorf("LLM配置验证失败: %w", err)
	}
	if c.MaxStep <= 0 {
		return errors.New("最大步骤数必须大于0")
	}
	return nil
}

// GetModel creates and returns a configured LLM model instance.
// It supports both OpenAI-compatible and Ollama providers.
// The method configures HTTP client with proxy if specified.
//
// Parameters:
// - ctx: Context for the operation
//
// Returns:
// - model.ToolCallingChatModel: Configured model instance
// - error: Error if model creation fails
func (c *Config) GetModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	httpClient, err := c.createHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("创建HTTP客户端失败: %w", err)
	}

	switch c.LLM.Type {
	case LLMTypeOpenAI:
		return c.createOpenAIModel(ctx, httpClient)
	case LLMTypeOllama:
		return c.createOllamaModel(ctx, httpClient)
	default:
		return nil, fmt.Errorf("不支持的LLM类型: %s", c.LLM.Type)
	}
}

// createHTTPClient creates an HTTP client with optional proxy configuration
func (c *Config) createHTTPClient() (*http.Client, error) {
	httpClient := http.DefaultClient

	if c.Proxy != "" {
		proxyURL, err := url.Parse(c.Proxy)
		if err != nil {
			return nil, fmt.Errorf("解析代理URL错误: %w", err)
		}

		httpClient = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}
	}

	return httpClient, nil
}

// createOpenAIModel creates an OpenAI-compatible model instance
func (c *Config) createOpenAIModel(ctx context.Context, httpClient *http.Client) (model.ToolCallingChatModel, error) {
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:    c.LLM.BaseURL,
		Model:      c.LLM.Model,
		APIKey:     c.LLM.APIKey,
		HTTPClient: httpClient,
	})
}

// createOllamaModel creates an Ollama model instance
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
// - ctx: Context for the operation
//
// Returns:
// - []tool.BaseTool: List of available tools
// - func(): Cleanup function to close MCP connections
// - error: Error if tool retrieval fails
func (c *Config) GetTools(ctx context.Context) ([]tool.BaseTool, func(), error) {
	// 连接mcp服务器
	mcpHub, err := mcphostNewMCPHub(ctx, c.MCP.ConfigFile)
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
// - configFile: Path to the configuration file to save
//
// Returns:
// - error: Error if saving fails
func (c *Config) SaveConfig(configFile string) error {
	if configFile == "" {
		return errors.New("配置文件路径不能为空")
	}

	// 将配置结构体的值设置到viper中
	c.setViperValues()

	if err := viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	return nil
}

// setViperValues sets configuration values in viper
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

// getDefaultConfig returns a default configuration with sensible defaults
func getDefaultConfig() Config {
	return Config{
		Proxy: "",
		MCP: MCPConfig{
			ConfigFile: defaultMCPConfigFile,
			Tools:      []string{},
		},
		LLM: LLMConfig{
			Type:    LLMTypeOllama,
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
//
// Parameters:
// - configFile: Path to configuration file. If empty, will search for config file automatically
//
// Returns:
// - *Config: Loaded configuration
// - error: Error if loading fails
func LoadConfig(configFile string) (*Config, error) {
	config := getDefaultConfig()

	if err := setupViper(configFile); err != nil {
		return nil, fmt.Errorf("设置viper失败: %w", err)
	}

	if err := readConfigFile(); err != nil {
		log.Println("未找到配置文件，使用默认配置")
	}

	// 将配置文件内容解析到结构体
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件错误: %w", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// setupViper configures viper for config file reading
func setupViper(configFile string) error {
	if configFile != "" {
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

// readConfigFile attempts to read the configuration file
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
