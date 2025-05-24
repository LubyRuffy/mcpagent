// Package mcphost provides MCP (Model Context Protocol) server management functionality.
// It handles connections to multiple MCP servers, tool discovery, and tool invocation
// through various transport mechanisms including stdio and SSE.
package mcphost

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
)

// 常量定义
const (
	// TransportTypeSSE represents Server-Sent Events transport
	TransportTypeSSE = "sse"
	// TransportTypeStdio represents standard input/output transport
	TransportTypeStdio = "stdio"

	// 错误消息
	errMsgUnknownError = "unknown error"
	errMsgNoContent    = "no content"
)

// MCPTools represents a collection of tools from an MCP server
type MCPTools struct {
	Name  string     // Server name
	Tools []mcp.Tool // Available tools
	Err   error      // Error if any occurred during tool discovery
}

// MCPHub manages connections to multiple MCP servers and provides unified tool access.
// It maintains a pool of connections and provides thread-safe access to tools.
type MCPHub struct {
	mu          sync.RWMutex                  // Protects concurrent access to connections and tools
	connections map[string]*Connection        // Active server connections
	tools       map[string]tool.InvokableTool // Available tools from all servers
	config      *MCPSettings                  // Configuration settings
}

// Connection represents a connection to a single MCP server
type Connection struct {
	Client client.MCPClient // MCP client instance
	Config ServerConfig     // Server configuration
}

// newMCPHub creates a new MCPHub instance with the given settings.
// It initializes all enabled servers and discovers their tools.
func newMCPHub(ctx context.Context, settings *MCPSettings) (*MCPHub, error) {
	h := &MCPHub{
		connections: make(map[string]*Connection),
		tools:       make(map[string]tool.InvokableTool),
		config:      settings,
	}

	if err := h.initServers(ctx); err != nil {
		return nil, fmt.Errorf("初始化服务器失败: %w", err)
	}

	return h, nil
}

// NewMCPHubFromString creates a new MCPHub from a JSON configuration string.
//
// Parameters:
// - ctx: Context for the operation
// - config: JSON configuration string
//
// Returns:
// - *MCPHub: Initialized MCPHub instance
// - error: Error if initialization fails
func NewMCPHubFromString(ctx context.Context, config string) (*MCPHub, error) {
	settings, err := LoadSettingsFromString(config)
	if err != nil {
		return nil, fmt.Errorf("加载配置字符串失败: %w", err)
	}
	return newMCPHub(ctx, settings)
}

// NewMCPHub creates a new MCPHub from a configuration file.
//
// Parameters:
// - ctx: Context for the operation
// - configPath: Path to the configuration file
//
// Returns:
// - *MCPHub: Initialized MCPHub instance
// - error: Error if initialization fails
func NewMCPHub(ctx context.Context, configPath string) (*MCPHub, error) {
	settings, err := LoadSettings(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置文件失败: %w", err)
	}
	return newMCPHub(ctx, settings)
}

// initServers initializes all enabled MCP servers
func (h *MCPHub) initServers(ctx context.Context) error {
	for name, config := range h.config.MCPServers {
		if config.Disabled {
			log.Printf("跳过已禁用的服务器: %s", name)
			continue
		}

		if err := h.connectToServer(ctx, name, config); err != nil {
			return fmt.Errorf("连接服务器 %s 失败: %w", name, err)
		}
	}

	return nil
}

// GetClient returns the client for the specified server.
// It provides thread-safe access to server connections.
//
// Parameters:
// - serverName: Name of the server
//
// Returns:
// - client.MCPClient: MCP client for the server
// - error: Error if server not found or disabled
func (h *MCPHub) GetClient(serverName string) (client.MCPClient, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conn, exists := h.connections[serverName]
	if !exists {
		return nil, fmt.Errorf("未找到服务器连接: %s", serverName)
	}

	if conn.Config.Disabled {
		return nil, fmt.Errorf("服务器已禁用: %s", serverName)
	}

	return conn.Client, nil
}

// createToolInvoker creates a tool invocation function for a specific server and tool.
// This function encapsulates the logic for calling MCP tools and handling responses.
func createToolInvoker(serverName, toolName string, cli *client.Client) func(ctx context.Context, params map[string]interface{}) (string, error) {
	return func(ctx context.Context, params map[string]interface{}) (string, error) {
		req := mcp.CallToolRequest{}
		req.Params.Name = toolName
		req.Params.Arguments = params

		callToolResult, err := cli.CallTool(ctx, req)
		if err != nil {
			return "", errors.Wrapf(err, "调用工具 %s/%s 失败", serverName, toolName)
		}

		if callToolResult.IsError {
			errMsg := errMsgUnknownError
			if len(callToolResult.Content) > 0 {
				errMsg = fmt.Sprintf("%v", callToolResult.Content[0])
			}
			return "", fmt.Errorf("MCP: 工具调用错误: %s", errMsg)
		}

		if len(callToolResult.Content) == 0 {
			return "", fmt.Errorf("MCP: 工具调用 %s 返回空内容", toolName)
		}

		textContent, ok := callToolResult.Content[0].(mcp.TextContent)
		if !ok {
			return "", fmt.Errorf("MCP: 工具调用 %s 返回不支持的内容类型: %T", toolName, callToolResult.Content[0])
		}

		return textContent.Text, nil
	}
}

// discoverTools discovers and registers tools from a specific MCP server.
// It converts MCP tool definitions to Eino tool format.
func (h *MCPHub) discoverTools(ctx context.Context, serverName string, cli *client.Client) error {
	listResults, err := cli.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("列出MCP工具失败: %w", err)
	}

	for _, mcpTool := range listResults.Tools {
		if err := h.registerTool(serverName, mcpTool, cli); err != nil {
			return fmt.Errorf("注册工具 %s 失败: %w", mcpTool.Name, err)
		}
	}

	return nil
}

// registerTool registers a single MCP tool as an Eino tool
func (h *MCPHub) registerTool(serverName string, mcpTool mcp.Tool, cli *client.Client) error {
	// Convert MCP tool schema to OpenAPI schema
	inputSchema, err := h.convertToolSchema(mcpTool)
	if err != nil {
		return fmt.Errorf("转换工具模式失败: %w", err)
	}

	// Create tool key with server prefix
	toolKey := serverName + "_" + mcpTool.Name

	// Register the tool
	h.tools[toolKey] = utils.NewTool(
		&schema.ToolInfo{
			Name:        mcpTool.Name,
			Desc:        mcpTool.Description,
			ParamsOneOf: schema.NewParamsOneOfByOpenAPIV3(inputSchema),
		},
		createToolInvoker(serverName, mcpTool.Name, cli),
	)

	return nil
}

// convertToolSchema converts MCP tool input schema to OpenAPI v3 schema
func (h *MCPHub) convertToolSchema(mcpTool mcp.Tool) (*openapi3.Schema, error) {
	marshaledInputSchema, err := sonic.Marshal(mcpTool.InputSchema)
	if err != nil {
		return nil, fmt.Errorf("序列化工具输入模式失败: %w", err)
	}

	inputSchema := &openapi3.Schema{}
	if err := sonic.Unmarshal(marshaledInputSchema, inputSchema); err != nil {
		return nil, fmt.Errorf("反序列化工具输入模式失败: %w", err)
	}

	return inputSchema, nil
}

// connectToServer establishes connection to a single MCP server.
// It handles different transport types and manages connection lifecycle.
func (h *MCPHub) connectToServer(ctx context.Context, serverName string, config ServerConfig) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	log.Printf("正在连接到MCP服务器: %s", serverName)

	// Close existing connection if any
	if err := h.closeExistingConnection(serverName); err != nil {
		return fmt.Errorf("关闭现有连接失败: %w", err)
	}

	// Create new client based on transport type
	mcpClient, err := h.createMCPClient(config)
	if err != nil {
		return fmt.Errorf("创建MCP客户端失败: %w", err)
	}

	// Setup logging for server stderr
	h.setupServerLogging(mcpClient, serverName)

	// Initialize the client
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "mcphost",
		Version: "0.1.0",
	}

	if _, err := mcpClient.Initialize(ctx, initRequest); err != nil {
		mcpClient.Close()
		return fmt.Errorf("初始化MCP客户端失败: %w", err)
	}

	// Store the connection
	h.connections[serverName] = &Connection{
		Client: mcpClient,
		Config: config,
	}

	// Discover and register tools
	if err := h.discoverTools(ctx, serverName, mcpClient); err != nil {
		return fmt.Errorf("发现工具失败: %w", err)
	}

	log.Printf("成功连接到MCP服务器: %s", serverName)
	return nil
}

// closeExistingConnection closes an existing connection if it exists
func (h *MCPHub) closeExistingConnection(serverName string) error {
	if existing, exists := h.connections[serverName]; exists {
		if err := existing.Client.Close(); err != nil {
			return fmt.Errorf("关闭现有连接失败: %w", err)
		}
		delete(h.connections, serverName)
	}
	return nil
}

// createMCPClient creates an MCP client based on the transport configuration
func (h *MCPHub) createMCPClient(config ServerConfig) (*client.Client, error) {
	switch config.TransportType {
	case TransportTypeSSE:
		return client.NewSSEMCPClient(config.URL)
	case TransportTypeStdio, "": // default to stdio
		env := h.buildEnvironment(config.Env)
		return client.NewStdioMCPClient(config.Command, env, config.Args...)
	default:
		return nil, fmt.Errorf("不支持的传输类型: %s", config.TransportType)
	}
}

// buildEnvironment builds environment variables for stdio transport
func (h *MCPHub) buildEnvironment(envMap map[string]string) []string {
	var env []string
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

// setupServerLogging sets up logging for server stderr output
func (h *MCPHub) setupServerLogging(mcpClient *client.Client, serverName string) {
	stderr, _ := client.GetStderr(mcpClient)

	if stderr != nil {
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				log.Printf("[%s] %s", serverName, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				log.Printf("读取服务器 %s 的stderr时出错: %v", serverName, err)
			}
		}()
	}
}

// CloseServers closes all server connections gracefully.
// It should be called when the MCPHub is no longer needed.
//
// Returns:
// - error: Error if any connection fails to close
func (h *MCPHub) CloseServers() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var errors []error
	for name, conn := range h.connections {
		if err := conn.Client.Close(); err != nil {
			errors = append(errors, fmt.Errorf("关闭服务器 %s 失败: %w", name, err))
		}
	}

	// Clear connections and tools
	h.connections = make(map[string]*Connection)
	h.tools = make(map[string]tool.InvokableTool)

	if len(errors) > 0 {
		return fmt.Errorf("关闭服务器时发生错误: %v", errors)
	}

	return nil
}

// GetEinoTools returns a list of Eino tools based on the provided tool names.
// If toolNameList is empty, it returns all available tools.
//
// Parameters:
// - ctx: Context for the operation
// - toolNameList: List of specific tool names to retrieve (empty for all tools)
//
// Returns:
// - []tool.BaseTool: List of available tools
// - error: Error if tool retrieval fails
func (h *MCPHub) GetEinoTools(ctx context.Context, toolNameList []string) ([]tool.BaseTool, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []tool.BaseTool

	if len(toolNameList) == 0 {
		// Return all tools if no specific tools requested
		for _, t := range h.tools {
			result = append(result, t)
		}
		return result, nil
	}

	// Return specific tools
	for _, toolName := range toolNameList {
		if t, exists := h.tools[toolName]; exists {
			result = append(result, t)
		} else {
			return nil, fmt.Errorf("工具不存在: %s", toolName)
		}
	}

	return result, nil
}

// InvokeTool invokes a specific tool with the given arguments.
// This method provides direct tool invocation without going through the Eino framework.
//
// Parameters:
// - ctx: Context for the operation
// - toolName: Name of the tool to invoke
// - arguments: Arguments to pass to the tool
//
// Returns:
// - string: Tool execution result
// - error: Error if tool invocation fails
func (h *MCPHub) InvokeTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	t, exists := h.tools[toolName]
	if !exists {
		return "", fmt.Errorf("工具不存在: %s", toolName)
	}

	// Convert arguments to JSON string for InvokableRun
	argsJSON, err := json.Marshal(arguments)
	if err != nil {
		return "", fmt.Errorf("序列化参数失败: %w", err)
	}

	return t.InvokableRun(ctx, string(argsJSON))
}
