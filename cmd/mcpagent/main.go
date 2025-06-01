// Package main provides the command-line interface for the FOFA Logs AI application.
// It handles configuration parsing, command-line argument processing,
// and orchestrates the execution of MCP agent tasks.
//
// The application supports:
// - Configuration loading from files and environment variables
// - Command-line argument override of configuration values
// - Graceful shutdown on interrupt signals
// - Comprehensive error handling and logging
// - Multiple LLM providers and MCP server configurations
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/LubyRuffy/mcpagent/pkg/config"
	"github.com/LubyRuffy/mcpagent/pkg/mcpagent"
)

// Exit code constants
const (
	// ExitCodeSuccess represents successful execution
	ExitCodeSuccess = 0
	// ExitCodeError represents error during execution
	ExitCodeError = 1
)

// Default values for command-line parsing
const (
	defaultToolsSeparator = ","
)

// Error message constants
const (
	errMsgTaskRequired     = "请使用 -task 参数指定要执行的任务"
	errMsgLoadConfigFailed = "加载配置失败: %w"
	errMsgConfigValidation = "配置验证失败: %w"
	errMsgSaveConfigFailed = "保存配置失败: %w"
	errMsgExecutionFailed  = "执行任务失败: %w"
)

// CommandLineArgs holds all command line arguments in a structured format.
// This struct provides type safety and clear documentation for all available options.
type CommandLineArgs struct {
	ConfigFile    *string // Path to configuration file
	Proxy         *string // Proxy server address for HTTP requests
	MCPConfigFile *string // Path to MCP server configuration file
	MCPTools      *string // Comma-separated list of tools to use
	LLMType       *string // LLM provider type (openai or ollama)
	LLMBaseURL    *string // Base URL for LLM API
	LLMModel      *string // LLM model name
	LLMAPIKey     *string // API key for LLM provider
	SystemPrompt  *string // System prompt for the agent
	MaxStep       *int    // Maximum number of reasoning steps
	Task          *string // Task description to execute
}

// fatalError handles fatal errors by logging and exiting with error code.
// This provides a consistent way to handle unrecoverable errors.
func fatalError(err error) {
	if err != nil {
		log.Fatalf("致命错误: %v", err)
	}
}

// parseCommandLineArgs parses and returns command line arguments.
// It sets up all available flags with appropriate descriptions and default values.
func parseCommandLineArgs() *CommandLineArgs {
	args := &CommandLineArgs{
		ConfigFile:    flag.String("config", "default_config.yaml", "配置文件路径"),
		Proxy:         flag.String("proxy", "", "代理服务器地址"),
		MCPConfigFile: flag.String("mcp-config", "", "MCP服务器配置文件路径"),
		MCPTools:      flag.String("mcp-tools", "", "工具列表（用逗号分隔）"),
		LLMType:       flag.String("llm-type", "", "LLM类型 (openai 或 ollama)"),
		LLMBaseURL:    flag.String("llm-base-url", "", "LLM API基础URL"),
		LLMModel:      flag.String("llm-model", "", "LLM模型名称"),
		LLMAPIKey:     flag.String("llm-api-key", "", "LLM API密钥"),
		SystemPrompt:  flag.String("system-prompt", "", "系统提示词"),
		MaxStep:       flag.Int("max-step", 0, "最大步骤数"),
		Task:          flag.String("task", "", "要执行的任务"),
	}

	flag.Parse()
	return args
}

// loadAndMergeConfig loads configuration from file and merges with command line arguments.
// Command line arguments take precedence over configuration file values.
// This allows for flexible configuration management with override capabilities.
func loadAndMergeConfig(args *CommandLineArgs) (*config.Config, error) {
	// Load base configuration from file
	cfg, err := config.LoadConfig(*args.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf(errMsgLoadConfigFailed, err)
	}

	// Merge command line arguments (they take precedence)
	mergeCommandLineArgs(cfg, args)

	// Validate the final configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf(errMsgConfigValidation, err)
	}

	return cfg, nil
}

// mergeCommandLineArgs merges command line arguments into configuration.
// Only non-empty command line values override configuration file values.
// This preserves the configuration file defaults when command line args are not provided.
func mergeCommandLineArgs(cfg *config.Config, args *CommandLineArgs) {
	if strings.TrimSpace(*args.Proxy) != "" {
		cfg.Proxy = *args.Proxy
	}
	if strings.TrimSpace(*args.MCPConfigFile) != "" {
		cfg.MCP.ConfigFile = *args.MCPConfigFile
	}
	if strings.TrimSpace(*args.MCPTools) != "" {
		cfg.MCP.Tools = parseToolsList(*args.MCPTools)
	}
	if strings.TrimSpace(*args.LLMType) != "" {
		cfg.LLM.Type = *args.LLMType
	}
	if strings.TrimSpace(*args.LLMBaseURL) != "" {
		cfg.LLM.BaseURL = *args.LLMBaseURL
	}
	if strings.TrimSpace(*args.LLMModel) != "" {
		cfg.LLM.Model = *args.LLMModel
	}
	if strings.TrimSpace(*args.LLMAPIKey) != "" {
		cfg.LLM.APIKey = *args.LLMAPIKey
	}
	if strings.TrimSpace(*args.SystemPrompt) != "" {
		cfg.SystemPrompt = *args.SystemPrompt
	}
	if *args.MaxStep != 0 {
		cfg.MaxStep = *args.MaxStep
	}
}

// parseToolsList parses comma-separated tools list into a slice of MCPToolConfig.
// It handles whitespace trimming and empty string filtering.
// 工具格式为 "server:tool_name"，如果不包含冒号，则默认使用内置服务器
func parseToolsList(toolsStr string) []config.MCPToolConfig {
	if strings.TrimSpace(toolsStr) == "" {
		return []config.MCPToolConfig{}
	}

	toolSpecs := strings.Split(toolsStr, defaultToolsSeparator)

	// 构建工具配置列表
	var toolConfigs []config.MCPToolConfig
	for _, toolSpec := range toolSpecs {
		toolSpec = strings.TrimSpace(toolSpec)
		if toolSpec == "" {
			continue
		}

		// 解析服务器和工具名称
		parts := strings.SplitN(toolSpec, ":", 2)
		if len(parts) == 2 {
			// 格式为 "server:tool_name"
			server := strings.TrimSpace(parts[0])
			toolName := strings.TrimSpace(parts[1])
			if server != "" && toolName != "" {
				toolConfigs = append(toolConfigs, config.MCPToolConfig{
					Server: server,
					Name:   toolName,
				})
			}
		} else {
			// 仅提供工具名称，使用内置服务器
			toolConfigs = append(toolConfigs, config.MCPToolConfig{
				Server: config.InnerServerName,
				Name:   toolSpec,
			})
		}
	}

	return toolConfigs
}

// saveConfigIfNeeded saves configuration to file if a config file path is provided
func saveConfigIfNeeded(cfg *config.Config, configFile string) error {
	if configFile == "" {
		return nil
	}

	if err := cfg.SaveConfig(configFile); err != nil {
		return fmt.Errorf(errMsgSaveConfigFailed, err)
	}

	log.Printf("配置已保存到: %s", configFile)
	return nil
}

// setupSignalHandling sets up graceful shutdown on interrupt signals
func setupSignalHandling(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("收到信号 %v，正在优雅关闭...", sig)
		cancel()
	}()
}

// validateTask validates that a task is provided
func validateTask(task string) error {
	if task == "" {
		return fmt.Errorf(errMsgTaskRequired)
	}
	return nil
}

// runAgent executes the MCP agent with the given configuration and task
func runAgent(ctx context.Context, cfg *config.Config, task string) error {
	notify := &mcpagent.CliNotifier{}

	log.Printf("开始执行任务: %s", task)

	if err := mcpagent.Run(ctx, cfg, task, notify); err != nil {
		return fmt.Errorf(errMsgExecutionFailed, err)
	}

	log.Println("任务执行完成")
	return nil
}

// main is the entry point of the application
func main() {
	// 解析命令行参数
	args := parseCommandLineArgs()

	// 验证任务参数
	if err := validateTask(*args.Task); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		flag.Usage()
		os.Exit(ExitCodeError)
	}

	// 加载和合并配置
	cfg, err := loadAndMergeConfig(args)
	if err != nil {
		log.Fatalf("配置错误: %v", err)
	}

	// 保存配置（如果需要）
	if err := saveConfigIfNeeded(cfg, "config.yaml"); err != nil {
		log.Printf("警告: %v", err)
	}

	// 设置上下文和信号处理
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupSignalHandling(cancel)

	// 执行任务
	if err := runAgent(ctx, cfg, *args.Task); err != nil {
		log.Fatalf("执行失败: %v", err)
	}
}
