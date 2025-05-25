// Package mcphost provides MCP (Model Context Protocol) server management functionality.
// It implements a hub pattern for managing multiple MCP server connections and provides
// unified tool access through the Eino framework.
//
// The package supports both SSE (Server-Sent Events) and stdio transport mechanisms,
// allowing flexible integration with different types of MCP servers. It handles
// connection lifecycle, tool discovery, and provides thread-safe access to tools.
//
// Example usage:
//
//	// Create hub from configuration file
//	hub, err := NewMCPHub(ctx, "mcpservers.json")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer hub.CloseServers()
//
//	// Get available tools
//	tools, err := hub.GetEinoTools(ctx, []string{"tool1", "tool2"})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Invoke a tool
//	result, err := hub.InvokeTool(ctx, "tool1", map[string]interface{}{"param": "value"})
//	if err != nil {
//		log.Fatal(err)
//	}
package mcphost

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// Error message constants
const (
	errMsgUnknownError = "unknown error"
	errMsgNoContent    = "no content"
)

// MCPTools represents a collection of tools from an MCP server.
// It includes the server name, available tools, and any errors encountered during discovery.
type MCPTools struct {
	Name  string     // Server name
	Tools []mcp.Tool // Available tools
	Err   error      // Error if any occurred during tool discovery
}

// MCPHub manages connections to multiple MCP servers and provides unified tool access.
// It maintains a pool of connections and provides thread-safe access to tools.
// The hub automatically discovers tools from connected servers and makes them available
// through the Eino framework.
type MCPHub struct {
	mu          sync.RWMutex                  // Protects concurrent access to connections and tools
	connections map[string]*Connection        // Active server connections indexed by server name
	tools       map[string]tool.InvokableTool // Available tools from all servers indexed by tool key
	config      *MCPSettings                  // Configuration settings for all servers
}

// Connection represents a connection to a single MCP server.
// It encapsulates the client instance and its configuration.
type Connection struct {
	Client client.MCPClient // MCP client instance for communication
	Config ServerConfig     // Server configuration used to establish the connection
}

// newMCPHub creates a new MCPHub instance with the given settings.
// It initializes all enabled servers and discovers their tools.
// This is an internal constructor used by the public factory functions.
func newMCPHub(ctx context.Context, settings *MCPSettings) (*MCPHub, error) {
	h := &MCPHub{
		connections: make(map[string]*Connection),
		tools:       make(map[string]tool.InvokableTool),
		config:      settings,
	}

	if err := h.initializeServers(ctx); err != nil {
		return nil, fmt.Errorf("初始化服务器失败: %w", err)
	}

	return h, nil
}

// NewMCPHubFromString creates a new MCPHub from a JSON configuration string.
// This is useful for programmatic configuration or testing scenarios.
//
// Parameters:
//   - ctx: Context for the operation
//   - config: JSON configuration string containing MCP server settings
//
// Returns:
//   - *MCPHub: Initialized MCPHub instance
//   - error: Error if initialization fails
func NewMCPHubFromString(ctx context.Context, config string) (*MCPHub, error) {
	settings, err := LoadSettingsFromString(config)
	if err != nil {
		return nil, fmt.Errorf("加载配置字符串失败: %w", err)
	}
	return newMCPHub(ctx, settings)
}

// NewMCPHub creates a new MCPHub from a configuration file.
// This is the primary way to create an MCPHub instance in production.
//
// Parameters:
//   - ctx: Context for the operation
//   - configPath: Path to the configuration file
//
// Returns:
//   - *MCPHub: Initialized MCPHub instance
//   - error: Error if initialization fails
func NewMCPHub(ctx context.Context, configPath string) (*MCPHub, error) {
	settings, err := LoadSettings(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置文件失败: %w", err)
	}
	return newMCPHub(ctx, settings)
}

// NewMCPHubFromSettings creates a new MCPHub directly from MCPSettings.
// This is useful when settings are already available in memory.
//
// Parameters:
//   - ctx: Context for the operation
//   - settings: MCPSettings containing server configurations
//
// Returns:
//   - *MCPHub: Initialized MCPHub instance
//   - error: Error if initialization fails
func NewMCPHubFromSettings(ctx context.Context, settings *MCPSettings) (*MCPHub, error) {
	return newMCPHub(ctx, settings)
}

// initializeServers initializes all enabled MCP servers.
// It iterates through the configuration and establishes connections to each enabled server.
func (h *MCPHub) initializeServers(ctx context.Context) error {
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
//   - serverName: Name of the server
//
// Returns:
//   - client.MCPClient: MCP client for the server
//   - error: Error if server not found or disabled
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
// It returns a function that can be used by the Eino framework to invoke the tool.
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
// It converts MCP tool definitions to Eino tool format and registers them in the hub.
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

// convertToolSchema converts MCP tool input schema to OpenAPI v3 schema.
// This conversion is necessary for integrating MCP tools with the Eino framework,
// which expects OpenAPI v3 schema format for tool parameters.
//
// Parameters:
//   - mcpTool: MCP tool definition containing the input schema to convert
//
// Returns:
//   - *openapi3.Schema: Converted OpenAPI v3 schema ready for Eino integration
//   - error: Error if schema conversion fails
func (h *MCPHub) convertToolSchema(mcpTool mcp.Tool) (*openapi3.Schema, error) {
	// fetch mcp 的bug：https://github.com/modelcontextprotocol/servers/issues/1817
	// 标准中exclusiveMaximum和exclusiveMinimum应该是bool，实际上设置成了integer，导致报错，需要处理
	for _, v := range mcpTool.InputSchema.Properties {
		switch values := v.(type) {
		case map[string]any:
			if _, ok := values["exclusiveMaximum"]; ok {
				delete(values, "exclusiveMaximum")
			}
			if _, ok := values["exclusiveMinimum"]; ok {
				delete(values, "exclusiveMinimum")
			}
		}
	}

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
// It handles different transport types (SSE and stdio) and manages the complete
// connection lifecycle including initialization, tool discovery, and error handling.
//
// The function performs the following steps:
//  1. Closes any existing connection for the server
//  2. Creates a new MCP client based on transport configuration
//  3. Sets up logging for server stderr output
//  4. Initializes the MCP protocol handshake
//  5. Discovers and registers available tools
//
// Parameters:
//   - ctx: Context for the operation
//   - serverName: Unique name identifier for the server
//   - config: Server configuration including transport and connection details
//
// Returns:
//   - error: Error if connection establishment fails at any step
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

// closeExistingConnection closes an existing connection if it exists.
// This function ensures clean connection management by properly closing
// and removing existing connections before establishing new ones.
//
// Parameters:
//   - serverName: Name of the server whose connection should be closed
//
// Returns:
//   - error: Error if connection closure fails
func (h *MCPHub) closeExistingConnection(serverName string) error {
	if existing, exists := h.connections[serverName]; exists {
		if err := existing.Client.Close(); err != nil {
			return fmt.Errorf("关闭现有连接失败: %w", err)
		}
		delete(h.connections, serverName)
	}
	return nil
}

// createMCPClient creates an MCP client based on the transport configuration.
// It supports both SSE (Server-Sent Events) and stdio transport mechanisms,
// with stdio being the default when no transport type is specified.
//
// Transport types:
//   - SSE: Creates a client that communicates via HTTP Server-Sent Events
//   - Stdio: Creates a client that communicates via standard input/output
//
// Parameters:
//   - config: Server configuration containing transport type and connection details
//
// Returns:
//   - *client.Client: Configured MCP client ready for initialization
//   - error: Error if client creation fails or transport type is unsupported
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

// buildEnvironment builds environment variables for stdio transport.
// It converts a map of environment variables to the slice format expected
// by the stdio MCP client, with each entry in "KEY=VALUE" format.
//
// Parameters:
//   - envMap: Map of environment variable names to values
//
// Returns:
//   - []string: Environment variables in "KEY=VALUE" format
func (h *MCPHub) buildEnvironment(envMap map[string]string) []string {
	var env []string
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

// setupServerLogging sets up logging for server stderr output.
// It creates a goroutine to continuously read from the server's stderr
// and logs the output with the server name prefix for debugging purposes.
// This is particularly useful for stdio-based MCP servers that may output
// diagnostic information to stderr.
//
// Parameters:
//   - mcpClient: MCP client instance to get stderr from
//   - serverName: Server name used as prefix in log messages
func (h *MCPHub) setupServerLogging(mcpClient *client.Client, serverName string) {
	stderr, _ := client.GetStderr(mcpClient)

	if stderr != nil {
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				log.Printf("[%s] %s", serverName, scanner.Text())
			}
			if err := scanner.Err(); err != nil && errors.Is(err, io.EOF) {
				log.Printf("读取服务器 %s 的stderr时出错: %v", serverName, err)
			}
		}()
	}
}

// CloseServers closes all server connections gracefully.
// It should be called when the MCPHub is no longer needed to ensure
// proper cleanup of resources and connections. The function attempts
// to close all connections and collects any errors that occur.
//
// After closing connections, it clears the internal connection and tool
// maps to ensure the hub is in a clean state.
//
// Returns:
//   - error: Aggregated error if any connection fails to close, nil if all succeed
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
// If toolNameList is empty, it returns all available tools from all connected servers.
// This method provides thread-safe access to the tool registry and is the primary
// way to retrieve tools for use with the Eino framework.
//
// The returned tools are ready for use with Eino's agent system and include
// all necessary metadata and invocation functions.
//
// Parameters:
//   - ctx: Context for the operation (currently unused but kept for future extensibility)
//   - toolNameList: List of specific tool names to retrieve (empty for all tools)
//
// Returns:
//   - []tool.BaseTool: List of available tools ready for Eino integration
//   - error: Error if any requested tool is not found
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
// This method provides direct tool invocation without going through the Eino framework,
// making it useful for simple tool calls or testing scenarios.
//
// The function performs thread-safe tool lookup and handles argument serialization
// automatically. It returns the raw tool output as a string.
//
// Parameters:
//   - ctx: Context for the operation
//   - toolName: Name of the tool to invoke (must exist in the tool registry)
//   - arguments: Arguments to pass to the tool as key-value pairs
//
// Returns:
//   - string: Tool execution result as returned by the MCP server
//   - error: Error if tool is not found, argument serialization fails, or invocation fails
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
