// Package main provides the web interface for the MCP Agent application.
// It starts an HTTP server with Server-Sent Events (SSE) support for real-time communication
// with the Vue.js frontend.
//
// The application supports:
// - Server-Sent Events (SSE) connections for real-time task execution
// - HTTP API endpoints for task execution (config provided by frontend)
// - Static file serving for the web UI
// - CORS support for development
// - Graceful shutdown on interrupt signals
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LubyRuffy/mcpagent/pkg/database"
	"github.com/LubyRuffy/mcpagent/pkg/mcphost"
	"github.com/LubyRuffy/mcpagent/pkg/services"
	"github.com/LubyRuffy/mcpagent/pkg/webserver"
)

// Exit code constants
const (
	// ExitCodeSuccess represents successful execution
	ExitCodeSuccess = 0
	// ExitCodeError represents error during execution
	ExitCodeError = 1
)

// Error message constants
const (
	errMsgServerStartFailed = "启动Web服务器失败: %w"
)

// CommandLineArgs holds all command line arguments for the web server
type CommandLineArgs struct {
	Port   *string // Server port
	Host   *string // Server host
	DBPath *string // Database file path
}

// parseCommandLineArgs parses and returns command line arguments
func parseCommandLineArgs() *CommandLineArgs {
	args := &CommandLineArgs{
		Port:   flag.String("port", "8081", "服务器端口"),
		Host:   flag.String("host", "", "服务器主机地址"),
		DBPath: flag.String("db", "./data/mcpagent.db", "数据库文件路径"),
	}

	flag.Parse()
	return args
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

// startWebServer starts the web server
func startWebServer(ctx context.Context, addr string) error {
	server := webserver.NewServer(addr)

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := server.Start(); err != nil {
			serverErr <- fmt.Errorf(errMsgServerStartFailed, err)
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		log.Println("正在关闭Web服务器...")
		// 创建一个有超时的上下文，用于优雅关闭
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		// 优雅关闭服务器
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("关闭Web服务器时出错: %v", err)
		}

		return nil
	case err := <-serverErr:
		return err
	}
}

// printStartupInfo prints startup information
func printStartupInfo(addr string) {
	log.Println("=== MCP Agent Web UI ===")
	log.Printf("配置: 由前端页面提供")
	log.Printf("服务器地址: http://localhost%s", addr)
	log.Printf("SSE端点: http://localhost%s/events", addr)

	log.Println("========================")
}

// main is the entry point of the web application
func main() {
	// Parse command line arguments
	args := parseCommandLineArgs()

	// Construct server address
	addr := fmt.Sprintf("%s:%s", *args.Host, *args.Port)
	if *args.Host == "" {
		addr = ":" + *args.Port
	}

	// Initialize database
	if err := database.InitDatabase(*args.DBPath); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	log.Printf("数据库初始化成功: %s", *args.DBPath)

	// 同步内置工具到数据库
	log.Println("开始同步内置工具到数据库...")
	if err := services.SyncInternalToolsWithDatabase(context.Background()); err != nil {
		log.Printf("警告: 同步内置工具失败: %v", err)
	} else {
		log.Println("内置工具同步成功")
	}

	// Print startup information
	printStartupInfo(addr)

	log.Println("Web服务器启动成功，配置将由前端页面提供")

	// Setup context and signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupSignalHandling(cancel)

	// Start web server
	if err := startWebServer(ctx, addr); err != nil {
		log.Fatalf("Web服务器错误: %v", err)
	}

	// Clean up all resources
	log.Println("正在清理资源...")

	// 清理MCP连接池
	pool := mcphost.GetConnectionPool()
	if errs := pool.Shutdown(); len(errs) > 0 {
		for _, err := range errs {
			log.Printf("关闭MCP连接池时出错: %v", err)
		}
	} else {
		log.Println("MCP连接池已清理")
	}

	// Close database connection
	if err := database.CloseDatabase(); err != nil {
		log.Printf("关闭数据库连接失败: %v", err)
	} else {
		log.Println("数据库连接已关闭")
	}

	log.Println("Web服务器已关闭")
}
