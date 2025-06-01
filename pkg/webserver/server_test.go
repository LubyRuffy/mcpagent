package webserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LubyRuffy/mcpagent/pkg/config"
	"github.com/LubyRuffy/mcpagent/pkg/mcphost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewServer tests server creation
func TestNewServer(t *testing.T) {
	server := NewServer(":8080")
	assert.NotNil(t, server)
	assert.Equal(t, ":8080", server.addr)
	assert.NotNil(t, server.router)
	assert.NotNil(t, server.clients)
}

// TestSSEEndpoint tests the SSE endpoint headers
func TestSSEEndpoint(t *testing.T) {
	server := NewServer(":8080")

	// Create a test request with a context that will be cancelled
	req := httptest.NewRequest("GET", "/events", nil)
	w := httptest.NewRecorder()

	// Start the handler in a goroutine
	done := make(chan bool)
	go func() {
		server.handleSSE(w, req)
		done <- true
	}()

	// Give it a moment to set headers and start
	time.Sleep(100 * time.Millisecond)

	// Check response headers
	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
	assert.Equal(t, "keep-alive", w.Header().Get("Connection"))
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))

	// Check that initial connection message was sent
	body := w.Body.String()
	assert.Contains(t, body, "data:")
	assert.Contains(t, body, "connected")
}

// TestTaskEndpointMissingConfig tests the task execution endpoint without config
func TestTaskEndpointMissingConfig(t *testing.T) {
	server := NewServer(":8080")

	// Test task request without config (should fail)
	taskReq := TaskRequest{
		Task: "测试任务",
	}

	body, err := json.Marshal(taskReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/task", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleExecuteTask(w, req)

	// Should return 400 for missing config
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "配置信息不能为空")
}

// TestTaskEndpointInvalidJSON tests task endpoint with invalid JSON
func TestTaskEndpointInvalidJSON(t *testing.T) {
	server := NewServer(":8080")

	req := httptest.NewRequest("POST", "/api/task", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleExecuteTask(w, req)

	// Should return 400 for invalid JSON
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestTaskEndpointEmptyTask tests task endpoint with empty task
func TestTaskEndpointEmptyTask(t *testing.T) {
	server := NewServer(":8080")

	taskReq := TaskRequest{
		Task: "",
		Config: &config.Config{
			LLM: config.LLMConfig{
				Type: "openai",
			},
		},
	}

	body, err := json.Marshal(taskReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/task", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleExecuteTask(w, req)

	// Should return 400 for empty task
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestTaskEndpointWithConfig tests the task execution endpoint with custom config
func TestTaskEndpointWithConfig(t *testing.T) {
	server := NewServer(":8080")

	// Test task request with custom config
	customConfig := &config.Config{
		LLM: config.LLMConfig{
			Type:    "ollama",
			BaseURL: "http://localhost:11434",
			Model:   "qwen3:4b",
			APIKey:  "ollama",
		},
		MCP: config.MCPConfig{
			ConfigFile: "mcpservers.json", // 提供配置文件路径
			Tools:      []string{"test_tool"},
		},
		SystemPrompt: "自定义系统提示词",
		MaxStep:      10,
		PlaceHolders: map[string]any{},
	}

	taskReq := TaskRequest{
		Task:   "测试任务",
		Config: customConfig,
	}

	body, err := json.Marshal(taskReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/task", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleExecuteTask(w, req)

	// Should return 200 for valid request with config
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestTaskEndpointInvalidConfig tests the task execution endpoint with invalid config
func TestTaskEndpointInvalidConfig(t *testing.T) {
	server := NewServer(":8080")

	// Test task request with invalid config (missing MCP config file and servers)
	invalidConfig := &config.Config{
		LLM: config.LLMConfig{
			Type:    "ollama",
			BaseURL: "http://localhost:11434",
			Model:   "qwen3:4b",
			APIKey:  "ollama",
		},
		MCP: config.MCPConfig{
			// 既没有ConfigFile也没有MCPServers，应该验证失败
			Tools: []string{"test_tool"},
		},
		SystemPrompt: "自定义系统提示词",
		MaxStep:      10,
		PlaceHolders: map[string]any{},
	}

	taskReq := TaskRequest{
		Task:   "测试任务",
		Config: invalidConfig,
	}

	body, err := json.Marshal(taskReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/task", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleExecuteTask(w, req)

	// Should return 400 for invalid config
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "配置验证失败")
}

// TestSSENotifier tests the SSE notifier implementation
func TestSSENotifier(t *testing.T) {
	w := httptest.NewRecorder()
	notifier := &SSENotifier{
		writer: w,
		taskID: "test-task",
	}

	// Test OnMessage
	notifier.OnMessage("测试消息")

	// Check that data was written
	assert.Contains(t, w.Body.String(), "测试消息")
	assert.Contains(t, w.Body.String(), "data:")

	// Reset recorder
	w.Body.Reset()

	// Test OnThinking
	notifier.OnThinking("思考中...")
	assert.Contains(t, w.Body.String(), "思考中...")
	assert.Contains(t, w.Body.String(), "thinking")

	// Reset recorder
	w.Body.Reset()

	// Test OnToolCall
	notifier.OnToolCall("test-tool", map[string]interface{}{"param": "value"})
	assert.Contains(t, w.Body.String(), "test-tool")
	assert.Contains(t, w.Body.String(), "tool_call")

	// Reset recorder
	w.Body.Reset()

	// Test OnResult
	notifier.OnResult("任务完成")
	assert.Contains(t, w.Body.String(), "任务完成")
	assert.Contains(t, w.Body.String(), "result")

	// Reset recorder
	w.Body.Reset()

	// Test OnError
	notifier.OnError(assert.AnError)
	assert.Contains(t, w.Body.String(), "error")
}

// TestBroadcastNotifier tests the broadcast notifier implementation
func TestBroadcastNotifier(t *testing.T) {
	server := NewServer(":8080")
	notifier := &BroadcastNotifier{
		server: server,
	}

	// Test methods don't panic
	assert.NotPanics(t, func() {
		notifier.OnMessage("广播消息")
		notifier.OnThinking("广播思考")
		notifier.OnToolCall("broadcast-tool", nil)
		notifier.OnResult("广播结果")
		notifier.OnError(assert.AnError)
	})
}

// TestSSEMessage tests SSE message creation and serialization
func TestSSEMessage(t *testing.T) {
	msg := SSEMessage{
		Type: "notify",
		Data: NotifyEvent{
			Type:      "message",
			Timestamp: time.Now().UnixMilli(),
			ID:        "test-id",
			Content:   "测试内容",
		},
	}

	// Test serialization
	data, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.Contains(t, string(data), "notify")
	assert.Contains(t, string(data), "测试内容")

	// Test deserialization
	var decoded SSEMessage
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "notify", decoded.Type)
}

// TestHandleTestLLMConnection tests the LLM connection test endpoint
func TestHandleTestLLMConnection(t *testing.T) {
	server := NewServer(":8080")

	tests := []struct {
		name           string
		llmConfig      config.LLMConfig
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "valid ollama config",
			llmConfig: config.LLMConfig{
				Type:    "ollama",
				BaseURL: "http://localhost:11434",
				Model:   "qwen3:4b",
				APIKey:  "ollama",
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "success",
		},
		{
			name: "valid openai config",
			llmConfig: config.LLMConfig{
				Type:    "openai",
				BaseURL: "https://api.openai.com/v1",
				Model:   "gpt-4",
				APIKey:  "test-key",
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "success",
		},
		{
			name: "invalid config - empty type",
			llmConfig: config.LLMConfig{
				BaseURL: "http://localhost:11434",
				Model:   "qwen3:4b",
				APIKey:  "ollama",
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "配置验证失败",
		},
		{
			name: "invalid config - empty base_url",
			llmConfig: config.LLMConfig{
				Type:   "ollama",
				Model:  "qwen3:4b",
				APIKey: "ollama",
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "配置验证失败",
		},
		{
			name: "invalid config - empty model",
			llmConfig: config.LLMConfig{
				Type:    "ollama",
				BaseURL: "http://localhost:11434",
				APIKey:  "ollama",
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "配置验证失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.llmConfig)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/llm/test", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.handleTestLLMConnection(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedMsg)
		})
	}
}

// TestHandleTestLLMConnectionInvalidJSON tests LLM test endpoint with invalid JSON
func TestHandleTestLLMConnectionInvalidJSON(t *testing.T) {
	server := NewServer(":8080")

	req := httptest.NewRequest("POST", "/api/llm/test", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleTestLLMConnection(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "解析LLM配置数据失败")
}

// TestHandleGetConfig tests the GET /api/config endpoint
func TestHandleGetConfig(t *testing.T) {
	server := NewServer(":8080")

	// Set a test config
	testConfig := &config.Config{
		LLM: config.LLMConfig{
			Type:    "ollama",
			BaseURL: "http://localhost:11434",
			Model:   "qwen3:4b",
			APIKey:  "ollama",
		},
		SystemPrompt: "测试系统提示词",
		MaxStep:      20,
	}
	server.config = testConfig

	req := httptest.NewRequest("GET", "/api/config", nil)
	w := httptest.NewRecorder()

	server.handleGetConfig(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var responseConfig config.Config
	err := json.Unmarshal(w.Body.Bytes(), &responseConfig)
	require.NoError(t, err)
	assert.Equal(t, "ollama", responseConfig.LLM.Type)
	assert.Equal(t, "测试系统提示词", responseConfig.SystemPrompt)
}

// TestHandleUpdateConfig tests the POST /api/config endpoint
func TestHandleUpdateConfig(t *testing.T) {
	server := NewServer(":8080")

	newConfig := config.Config{
		LLM: config.LLMConfig{
			Type:    "openai",
			BaseURL: "https://api.openai.com/v1",
			Model:   "gpt-4",
			APIKey:  "test-key",
		},
		MCP: config.MCPConfig{
			ConfigFile: "mcpservers.json",
			Tools:      []string{"test_tool"},
		},
		SystemPrompt: "新的系统提示词",
		MaxStep:      30,
		PlaceHolders: map[string]any{},
	}

	body, err := json.Marshal(newConfig)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/config", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleUpdateConfig(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "配置更新成功", response["message"])

	// Verify config was actually updated
	assert.Equal(t, "openai", server.config.LLM.Type)
	assert.Equal(t, "新的系统提示词", server.config.SystemPrompt)
}

// TestHandleUpdateConfigInvalidJSON tests POST /api/config with invalid JSON
func TestHandleUpdateConfigInvalidJSON(t *testing.T) {
	server := NewServer(":8080")

	req := httptest.NewRequest("POST", "/api/config", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleUpdateConfig(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "解析配置数据失败")
}

// TestHandleUpdateConfigInvalidConfig tests POST /api/config with invalid config
func TestHandleUpdateConfigInvalidConfig(t *testing.T) {
	server := NewServer(":8080")

	invalidConfig := config.Config{
		LLM: config.LLMConfig{
			// Missing required fields
		},
	}

	body, err := json.Marshal(invalidConfig)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/config", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleUpdateConfig(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "配置验证失败")
}

func TestHandleGetMCPTools(t *testing.T) {
	// 创建测试服务器
	server := &Server{}

	tests := []struct {
		name           string
		requestBody    MCPToolsRequest
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name: "空的MCP服务器配置",
			requestBody: MCPToolsRequest{
				MCPServers: map[string]mcphost.ServerConfig{},
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name: "无效的MCP服务器配置",
			requestBody: MCPToolsRequest{
				MCPServers: map[string]mcphost.ServerConfig{
					"invalid_server": {
						Command: "invalid_command_that_does_not_exist",
						Args:    []string{"--invalid"},
						Env:     map[string]string{},
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 准备请求体
			requestBody, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			// 创建HTTP请求
			req, err := http.NewRequest("POST", "/api/mcp/tools", bytes.NewBuffer(requestBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// 创建响应记录器
			rr := httptest.NewRecorder()

			// 调用处理函数
			server.handleGetMCPTools(rr, req)

			// 检查状态码
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// 解析响应
			var response MCPToolsResponse
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err)

			// 检查响应结构
			assert.Equal(t, tt.expectSuccess, response.Success)

			if tt.expectSuccess {
				// 对于成功的情况，Tools可以是nil或空数组
				assert.Empty(t, response.Error)
			} else {
				assert.NotEmpty(t, response.Error)
			}
		})
	}
}

func TestMCPToolsRequestValidation(t *testing.T) {
	server := &Server{}

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "无效的JSON格式",
			requestBody:    `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "空的请求体",
			requestBody:    "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "有效的JSON格式",
			requestBody:    `{"mcp_servers": {}}`,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/api/mcp/tools", bytes.NewBufferString(tt.requestBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			server.handleGetMCPTools(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestMCPToolsResponseStructure(t *testing.T) {
	// 测试响应结构的正确性
	tools := []MCPToolInfo{
		{
			Name:        "test_tool",
			Description: "A test tool",
			Server:      "test_server",
		},
	}

	response := MCPToolsResponse{
		Success: true,
		Message: "测试成功",
		Tools:   tools,
		Error:   "",
	}

	// 序列化和反序列化测试
	data, err := json.Marshal(response)
	assert.NoError(t, err)

	var decoded MCPToolsResponse
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, response.Success, decoded.Success)
	assert.Equal(t, response.Message, decoded.Message)
	assert.Equal(t, len(response.Tools), len(decoded.Tools))
	if len(response.Tools) > 0 && len(decoded.Tools) > 0 {
		assert.Equal(t, response.Tools[0].Name, decoded.Tools[0].Name)
		assert.Equal(t, response.Tools[0].Description, decoded.Tools[0].Description)
		assert.Equal(t, response.Tools[0].Server, decoded.Tools[0].Server)
	}
}
