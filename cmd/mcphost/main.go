// Package main provides the command-line interface for the MCP Agent.
// It handles configuration parsing, command-line argument processing,
// and orchestrates the execution of MCP agent tasks.
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

// 常量定义
const (
	// ExitCodeSuccess represents successful execution
	ExitCodeSuccess = 0
	// ExitCodeError represents error during execution
	ExitCodeError = 1

	// 默认值
	defaultToolsSeparator = ","
)

// CommandLineArgs holds all command line arguments
type CommandLineArgs struct {
	ConfigFile    *string
	Proxy         *string
	MCPConfigFile *string
	LLMType       *string
	LLMBaseURL    *string
	LLMModel      *string
	LLMAPIKey     *string
	SystemPrompt  *string
	Tools         *string
	MaxStep       *int
	Task          *string
}

// fatalError handles fatal errors by logging and exiting
func fatalError(err error) {
	if err != nil {
		log.Fatalf("致命错误: %v", err)
	}
}

// parseCommandLineArgs parses and returns command line arguments
func parseCommandLineArgs() *CommandLineArgs {
	args := &CommandLineArgs{
		ConfigFile:    flag.String("config", "", "配置文件路径"),
		Proxy:         flag.String("proxy", "", "代理服务器地址"),
		MCPConfigFile: flag.String("mcp-config", "", "MCP服务器配置文件路径"),
		LLMType:       flag.String("llm-type", "", "LLM类型 (openai 或 ollama)"),
		LLMBaseURL:    flag.String("llm-base-url", "", "LLM API基础URL"),
		LLMModel:      flag.String("llm-model", "", "LLM模型名称"),
		LLMAPIKey:     flag.String("llm-api-key", "", "LLM API密钥"),
		SystemPrompt:  flag.String("system-prompt", "", "系统提示词"),
		Tools:         flag.String("tools", "", "工具列表（用逗号分隔）"),
		MaxStep:       flag.Int("max-step", 0, "最大步骤数"),
		Task:          flag.String("task", "", "要执行的任务"),
	}

	flag.Parse()
	return args
}

// loadAndMergeConfig loads configuration from file and merges with command line arguments
func loadAndMergeConfig(args *CommandLineArgs) (*config.Config, error) {
	// 加载配置文件
	cfg, err := config.LoadConfig(*args.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 命令行参数优先级高于配置文件
	mergeCommandLineArgs(cfg, args)

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return cfg, nil
}

// mergeCommandLineArgs merges command line arguments into configuration
func mergeCommandLineArgs(cfg *config.Config, args *CommandLineArgs) {
	if *args.Proxy != "" {
		cfg.Proxy = *args.Proxy
	}
	if *args.MCPConfigFile != "" {
		cfg.MCP.ConfigFile = *args.MCPConfigFile
	}
	if *args.Tools != "" {
		cfg.MCP.Tools = parseToolsList(*args.Tools)
	}
	if *args.LLMType != "" {
		cfg.LLM.Type = *args.LLMType
	}
	if *args.LLMBaseURL != "" {
		cfg.LLM.BaseURL = *args.LLMBaseURL
	}
	if *args.LLMModel != "" {
		cfg.LLM.Model = *args.LLMModel
	}
	if *args.LLMAPIKey != "" {
		cfg.LLM.APIKey = *args.LLMAPIKey
	}
	if *args.SystemPrompt != "" {
		cfg.SystemPrompt = *args.SystemPrompt
	}
	if *args.MaxStep != 0 {
		cfg.MaxStep = *args.MaxStep
	}
}

// parseToolsList parses comma-separated tools list
func parseToolsList(toolsStr string) []string {
	if toolsStr == "" {
		return []string{}
	}

	tools := strings.Split(toolsStr, defaultToolsSeparator)
	// 去除空白字符
	for i, tool := range tools {
		tools[i] = strings.TrimSpace(tool)
	}

	return tools
}

// saveConfigIfNeeded saves configuration to file if a config file path is provided
func saveConfigIfNeeded(cfg *config.Config, configFile string) error {
	if configFile == "" {
		return nil
	}

	if err := cfg.SaveConfig(configFile); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
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
		return fmt.Errorf("请使用 -task 参数指定要执行的任务")
	}
	return nil
}

// runAgent executes the MCP agent with the given configuration and task
func runAgent(ctx context.Context, cfg *config.Config, task string) error {
	notify := &mcpagent.CliNotifier{}

	log.Printf("开始执行任务: %s", task)

	if err := mcpagent.Run(ctx, cfg, task, notify); err != nil {
		return fmt.Errorf("执行任务失败: %w", err)
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
	if err := saveConfigIfNeeded(cfg, *args.ConfigFile); err != nil {
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
