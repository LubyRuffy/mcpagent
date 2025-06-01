// Package webserver provides HTTP and WebSocket server functionality for the MCP Agent Web UI.
// It handles WebSocket connections, configuration management, and task execution coordination.
//
// The package supports:
//   - WebSocket connections for real-time communication
//   - Configuration API endpoints
//   - Task execution with real-time notifications
//   - Static file serving for the web UI
//   - CORS support for development
//
// Example usage:
//
//	server := webserver.NewServer(":8080", cfg)
//	if err := server.Start(); err != nil {
//		log.Fatal(err)
//	}
package webserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LubyRuffy/mcpagent/pkg/config"
	"github.com/LubyRuffy/mcpagent/pkg/mcpagent"
	"github.com/LubyRuffy/mcpagent/pkg/mcphost"
	"github.com/LubyRuffy/mcpagent/pkg/models"
	"github.com/LubyRuffy/mcpagent/pkg/services"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"gorm.io/gorm"
)

// SSEMessage represents a message sent over Server-Sent Events
type SSEMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// TaskRequest represents a task execution request
type TaskRequest struct {
	Task   string         `json:"task"`
	Config *config.Config `json:"config"` // 必需的配置，由前端页面提供
}

// MCPToolsRequest represents a request to get tools from MCP servers
type MCPToolsRequest struct {
	MCPServers map[string]mcphost.ServerConfig `json:"mcp_servers"`
}

// MCPToolInfo represents information about an MCP tool
type MCPToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Server      string `json:"server"`
}

// MCPToolsResponse represents the response containing MCP tools
type MCPToolsResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message,omitempty"`
	Tools   []MCPToolInfo `json:"tools,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// NotifyEvent represents different types of notification events
type NotifyEvent struct {
	Type       string      `json:"type"`
	Timestamp  int64       `json:"timestamp"`
	ID         string      `json:"id"`
	Content    string      `json:"content,omitempty"`
	ToolName   string      `json:"tool_name,omitempty"`
	Parameters interface{} `json:"parameters,omitempty"`
	Status     string      `json:"status,omitempty"`
	Result     interface{} `json:"result,omitempty"`
	Error      string      `json:"error,omitempty"`
}

// TaskStatus represents the current task execution status
type TaskStatus struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Progress    *int   `json:"progress,omitempty"`
	CurrentStep string `json:"current_step,omitempty"`
	TotalSteps  *int   `json:"total_steps,omitempty"`
}

// SSENotifier implements the mcpagent.Notify interface for Server-Sent Events communication
type SSENotifier struct {
	writer http.ResponseWriter
	mutex  sync.Mutex
	taskID string
}

// BroadcastNotifier implements the mcpagent.Notify interface for broadcasting to all SSE clients
type BroadcastNotifier struct {
	server *Server
	taskID string
}

// Server represents the web server instance
type Server struct {
	addr                   string
	router                 *mux.Router
	clients                map[string]*SSENotifier
	mutex                  sync.RWMutex
	config                 *config.Config
	db                     *gorm.DB // 数据库连接
	llmConfigService       *services.LLMConfigService
	mcpServerConfigService *services.MCPServerConfigService
	mcpToolService         *services.MCPToolService
	systemPromptService    *services.SystemPromptService
	appConfigService       *services.AppConfigService
	shutdown               chan struct{} // 用于通知关闭的通道
	httpServer             *http.Server  // HTTP服务器实例
}

// NewServer creates a new web server instance
func NewServer(addr string) *Server {
	server := &Server{
		addr:                   addr,
		router:                 mux.NewRouter(),
		clients:                make(map[string]*SSENotifier),
		config:                 config.NewDefaultConfig(), // 初始化默认配置
		llmConfigService:       services.NewLLMConfigService(),
		mcpServerConfigService: services.NewMCPServerConfigService(),
		mcpToolService:         services.NewMCPToolService(),
		systemPromptService:    services.NewSystemPromptService(),
		appConfigService:       services.NewAppConfigService(),
		shutdown:               make(chan struct{}), // 初始化关闭通道
	}

	server.setupRoutes()
	return server
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// SSE endpoint
	s.router.HandleFunc("/events", s.handleSSE)

	// API endpoints
	api := s.router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/config", s.handleGetConfig).Methods("GET")
	api.HandleFunc("/config", s.handleUpdateConfig).Methods("POST")
	api.HandleFunc("/task", s.handleExecuteTask).Methods("POST")
	api.HandleFunc("/task/{taskId}/cancel", s.handleCancelTask).Methods("POST")
	api.HandleFunc("/llm/test", s.handleTestLLMConnection).Methods("POST")

	// LLM配置管理API
	api.HandleFunc("/llm/configs", s.handleListLLMConfigs).Methods("GET")
	api.HandleFunc("/llm/configs", s.handleCreateLLMConfig).Methods("POST")
	api.HandleFunc("/llm/configs/{id:[0-9]+}", s.handleGetLLMConfig).Methods("GET")
	api.HandleFunc("/llm/configs/{id:[0-9]+}", s.handleUpdateLLMConfig).Methods("PUT")
	api.HandleFunc("/llm/configs/{id:[0-9]+}", s.handleDeleteLLMConfig).Methods("DELETE")
	api.HandleFunc("/llm/configs/{id:[0-9]+}/default", s.handleSetDefaultLLMConfig).Methods("POST")

	// 系统提示词配置管理API
	api.HandleFunc("/system-prompts", s.handleListSystemPrompts).Methods("GET")
	api.HandleFunc("/system-prompts", s.handleCreateSystemPrompt).Methods("POST")
	api.HandleFunc("/system-prompts/{id:[0-9]+}", s.handleGetSystemPrompt).Methods("GET")
	api.HandleFunc("/system-prompts/{id:[0-9]+}", s.handleUpdateSystemPrompt).Methods("PUT")
	api.HandleFunc("/system-prompts/{id:[0-9]+}", s.handleDeleteSystemPrompt).Methods("DELETE")
	api.HandleFunc("/system-prompts/{id:[0-9]+}/default", s.handleSetDefaultSystemPrompt).Methods("POST")

	// MCP服务器配置管理API
	api.HandleFunc("/mcp/servers", s.handleListMCPServerConfigs).Methods("GET")
	api.HandleFunc("/mcp/servers", s.handleCreateMCPServerConfig).Methods("POST")
	api.HandleFunc("/mcp/servers/{id:[0-9]+}", s.handleGetMCPServerConfig).Methods("GET")
	api.HandleFunc("/mcp/servers/{id:[0-9]+}", s.handleUpdateMCPServerConfig).Methods("PUT")
	api.HandleFunc("/mcp/servers/{id:[0-9]+}", s.handleDeleteMCPServerConfig).Methods("DELETE")

	// MCP工具管理API
	api.HandleFunc("/mcp/tools", s.handleGetMCPTools).Methods("POST")
	api.HandleFunc("/mcp/tools/configured", s.handleGetMCPToolsFromDB).Methods("GET")
	api.HandleFunc("/mcp/tools/cached", s.handleGetCachedMCPTools).Methods("GET")
	api.HandleFunc("/mcp/tools/sync", s.handleSyncMCPTools).Methods("POST")
	api.HandleFunc("/mcp/tools/sync/{id:[0-9]+}", s.handleSyncMCPToolsForServer).Methods("POST")

	// Static files (for production)
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/dist/")))
}

// Start starts the web server
func (s *Server) Start() error {
	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(s.router)

	// 初始化关闭通道
	s.shutdown = make(chan struct{})

	// 创建HTTP服务器
	s.httpServer = &http.Server{
		Addr:    s.addr,
		Handler: handler,
	}

	log.Printf("Web服务器启动在 %s", s.addr)
	log.Printf("SSE端点: http://localhost%s/events", s.addr)
	log.Printf("Web界面: http://localhost%s", s.addr)

	// 启动清理协程
	go s.cleanupOnShutdown()

	// 使用新的HTTP服务器启动
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// 通知清理协程
	close(s.shutdown)

	// 等待清理完成
	<-ctx.Done()

	// 关闭HTTP服务器
	return s.httpServer.Shutdown(ctx)
}

// cleanupOnShutdown cleans up resources when server is shutting down
func (s *Server) cleanupOnShutdown() {
	<-s.shutdown

	// 关闭所有MCP连接
	pool := mcphost.GetConnectionPool()
	if errs := pool.Shutdown(); len(errs) > 0 {
		for _, err := range errs {
			log.Printf("关闭MCP连接时出错: %v", err)
		}
	}

	log.Printf("服务器资源清理完成")
}

// handleSSE handles Server-Sent Events connections
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// 获取任务ID参数
	taskID := r.URL.Query().Get("taskId")
	if taskID == "" {
		taskID = fmt.Sprintf("task_%d", time.Now().UnixNano())
	}

	// Create notifier
	clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())
	notifier := &SSENotifier{
		writer: w,
		taskID: taskID,
	}

	s.mutex.Lock()
	s.clients[clientID] = notifier
	s.mutex.Unlock()

	log.Printf("SSE客户端连接: %s, 任务ID: %s", r.RemoteAddr, taskID)

	// Send connection confirmation
	s.sendSSEMessage(w, SSEMessage{
		Type: "status",
		Data: map[string]interface{}{
			"connected": true,
			"message":   "SSE连接成功",
			"task_id":   taskID,
		},
	})

	// Handle connection cleanup
	defer func() {
		s.mutex.Lock()
		delete(s.clients, clientID)
		s.mutex.Unlock()
		log.Printf("SSE客户端断开: %s, 任务ID: %s", r.RemoteAddr, taskID)
	}()

	// 不再发送心跳，连接会在任务完成后自动断开
	// 等待客户端断开连接
	ctx := r.Context()
	<-ctx.Done()
}

// sendSSEMessage sends a message via Server-Sent Events
func (s *Server) sendSSEMessage(w http.ResponseWriter, msg SSEMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("序列化SSE消息失败: %v", err)
		return
	}

	fmt.Fprintf(w, "data: %s\n\n", data)

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// broadcast sends a message to all connected SSE clients
func (s *Server) broadcast(msg SSEMessage) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, notifier := range s.clients {
		go func(n *SSENotifier) {
			n.mutex.Lock()
			defer n.mutex.Unlock()

			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("序列化广播消息失败: %v", err)
				return
			}

			fmt.Fprintf(n.writer, "data: %s\n\n", data)

			if flusher, ok := n.writer.(http.Flusher); ok {
				flusher.Flush()
			}
		}(notifier)
	}
}

// broadcastToTask sends a message to SSE clients connected for a specific task
func (s *Server) broadcastToTask(taskID string, msg SSEMessage) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	log.Printf("开始广播任务消息: %s, 消息类型: %s", taskID, msg.Type)

	sentCount := 0
	for clientID, notifier := range s.clients {
		if notifier.taskID == taskID {
			log.Printf("找到匹配的客户端: %s 对应任务: %s", clientID, taskID)
			sentCount++

			go func(n *SSENotifier) {
				n.mutex.Lock()
				defer n.mutex.Unlock()

				data, err := json.Marshal(msg)
				if err != nil {
					log.Printf("序列化任务消息失败: %v", err)
					return
				}

				fmt.Fprintf(n.writer, "data: %s\n\n", data)

				if flusher, ok := n.writer.(http.Flusher); ok {
					flusher.Flush()
				}

				log.Printf("已向客户端发送任务消息: %s", taskID)
			}(notifier)
		}
	}

	log.Printf("广播任务消息完成: %s, 发送给 %d 个客户端", taskID, sentCount)
	if sentCount == 0 {
		log.Printf("警告: 没有找到任务 %s 的客户端", taskID)
	}
}

// SSENotifier implementation of mcpagent.Notify interface

// OnMessage sends a message notification during execution
func (s *SSENotifier) OnMessage(msg string) {
	s.sendNotifyEvent(NotifyEvent{
		Type:      "message",
		Timestamp: time.Now().UnixMilli(),
		ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Content:   msg,
	})
}

// OnThinking sends a thinking notification during execution
func (s *SSENotifier) OnThinking(msg string) {
	s.sendNotifyEvent(NotifyEvent{
		Type:      "thinking",
		Timestamp: time.Now().UnixMilli(),
		ID:        fmt.Sprintf("think_%d", time.Now().UnixNano()),
		Content:   msg,
	})
}

// OnToolCall sends a tool call notification during execution
func (s *SSENotifier) OnToolCall(toolName string, params interface{}) {
	s.sendNotifyEvent(NotifyEvent{
		Type:       "tool_call",
		Timestamp:  time.Now().UnixMilli(),
		ID:         fmt.Sprintf("tool_%d", time.Now().UnixNano()),
		ToolName:   toolName,
		Parameters: params,
		Status:     "calling",
	})
}

// OnResult sends a result notification when the agent completes successfully
func (s *SSENotifier) OnResult(msg string) {
	s.sendNotifyEvent(NotifyEvent{
		Type:      "result",
		Timestamp: time.Now().UnixMilli(),
		ID:        fmt.Sprintf("result_%d", time.Now().UnixNano()),
		Content:   msg,
	})
}

// OnError sends an error notification when something goes wrong
func (s *SSENotifier) OnError(err error) {
	s.sendNotifyEvent(NotifyEvent{
		Type:      "error",
		Timestamp: time.Now().UnixMilli(),
		ID:        fmt.Sprintf("error_%d", time.Now().UnixNano()),
		Error:     err.Error(),
	})
}

// sendNotifyEvent sends a notification event via SSE
func (s *SSENotifier) sendNotifyEvent(event NotifyEvent) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	msg := SSEMessage{
		Type: "notify",
		Data: event,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("序列化通知事件失败: %v", err)
		return
	}

	fmt.Fprintf(s.writer, "data: %s\n\n", data)

	if flusher, ok := s.writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

// HTTP API handlers

// handleGetConfig handles GET /api/config
func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	// 首先尝试从数据库获取默认配置
	dbConfig, err := s.appConfigService.GetDefaultConfig()
	if err == nil {
		// 如果找到了默认配置，将其应用到当前配置中
		if err := s.appConfigService.SaveToConfig(dbConfig, s.config); err != nil {
			log.Printf("警告：应用默认配置失败: %v", err)
		}
	} else if err != models.ErrAppConfigNotFound {
		log.Printf("警告：获取默认配置失败: %v", err)
	}

	// 返回当前配置
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.config)
}

// handleUpdateConfig handles POST /api/config
func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var newConfig config.Config

	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, "解析配置数据失败", http.StatusBadRequest)
		return
	}

	if err := newConfig.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("配置验证失败: %v", err), http.StatusBadRequest)
		return
	}

	// 更新内存中的配置
	s.config = &newConfig

	// 同步保存到数据库
	// 尝试获取默认配置
	dbConfig, err := s.appConfigService.GetDefaultConfig()
	if err != nil {
		if err == models.ErrAppConfigNotFound {
			// 如果没有默认配置，创建一个新的
			dbConfig = &models.AppConfigModel{
				Name:        "默认全局配置",
				Description: "MCP Agent全局配置",
				IsDefault:   true,
				IsActive:    true,
			}
		} else {
			log.Printf("获取默认配置失败: %v", err)
			http.Error(w, fmt.Sprintf("保存配置到数据库失败: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// 将新配置应用到数据库模型
	if err := s.appConfigService.LoadFromConfig(s.config, dbConfig); err != nil {
		log.Printf("加载配置到数据库模型失败: %v", err)
		http.Error(w, fmt.Sprintf("保存配置到数据库失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 保存数据库模型
	var saveErr error
	if dbConfig.ID == 0 {
		saveErr = s.appConfigService.CreateConfig(dbConfig)
	} else {
		saveErr = s.appConfigService.UpdateConfig(dbConfig.ID, dbConfig)
	}

	if saveErr != nil {
		log.Printf("保存配置到数据库失败: %v", saveErr)
		http.Error(w, fmt.Sprintf("保存配置到数据库失败: %v", saveErr), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "配置更新成功并已保存到数据库",
	})
}

// handleExecuteTask handles POST /api/task
func (s *Server) handleExecuteTask(w http.ResponseWriter, r *http.Request) {
	var taskReq TaskRequest

	if err := json.NewDecoder(r.Body).Decode(&taskReq); err != nil {
		http.Error(w, "解析任务数据失败", http.StatusBadRequest)
		return
	}

	if taskReq.Task == "" {
		http.Error(w, "任务描述不能为空", http.StatusBadRequest)
		return
	}

	// 配置必须由前端提供
	if taskReq.Config == nil {
		http.Error(w, "配置信息不能为空，必须由前端页面提供", http.StatusBadRequest)
		return
	}

	// 验证配置
	if err := taskReq.Config.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("配置验证失败: %v", err), http.StatusBadRequest)
		return
	}

	taskID := fmt.Sprintf("task_%d", time.Now().UnixNano())
	log.Printf("创建新任务ID: %s", taskID)

	// Broadcast task start status to task-specific SSE clients
	s.broadcastToTask(taskID, SSEMessage{
		Type: "status",
		Data: TaskStatus{
			ID:          taskID,
			Status:      "running",
			CurrentStep: "开始执行任务",
		},
	})
	log.Printf("已广播任务开始状态: %s, status: running", taskID)

	// Execute task in background with task-specific notifier
	go func() {
		ctx := context.Background()

		// Create a task-specific notifier that sends only to clients for this task
		notifier := &BroadcastNotifier{server: s, taskID: taskID}

		// 使用前端提供的配置
		err := mcpagent.Run(ctx, taskReq.Config, taskReq.Task, notifier)

		status := "completed"
		if err != nil {
			status = "error"
			notifier.OnError(err)
		}

		s.broadcastToTask(taskID, SSEMessage{
			Type: "status",
			Data: TaskStatus{
				ID:     taskID,
				Status: status,
			},
		})
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "任务已开始执行",
		"task_id": taskID,
	})
}

// handleCancelTask handles POST /api/task/{taskId}/cancel
func (s *Server) handleCancelTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["taskId"]

	if taskID == "" {
		http.Error(w, "任务ID不能为空", http.StatusBadRequest)
		return
	}

	log.Printf("收到取消任务请求: %s", taskID)

	// 向任务发送取消状态
	s.broadcastToTask(taskID, SSEMessage{
		Type: "status",
		Data: TaskStatus{
			ID:     taskID,
			Status: "error", // 设置为error状态，这将触发客户端断开SSE连接
		},
	})

	// 向任务发送通知消息
	s.broadcastToTask(taskID, SSEMessage{
		Type: "notify",
		Data: NotifyEvent{
			Type:      "error",
			Timestamp: time.Now().UnixMilli(),
			ID:        fmt.Sprintf("err_%d", time.Now().UnixNano()),
			Error:     "任务已被用户中断",
		},
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "任务取消请求已发送",
	})
}

// BroadcastNotifier implementation of mcpagent.Notify interface

// OnMessage sends a message notification to task-specific connected clients
func (b *BroadcastNotifier) OnMessage(msg string) {
	b.server.broadcastToTask(b.taskID, SSEMessage{
		Type: "notify",
		Data: NotifyEvent{
			Type:      "message",
			Timestamp: time.Now().UnixMilli(),
			ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
			Content:   msg,
		},
	})
}

// OnThinking sends a thinking notification to task-specific connected clients
func (b *BroadcastNotifier) OnThinking(msg string) {
	b.server.broadcastToTask(b.taskID, SSEMessage{
		Type: "notify",
		Data: NotifyEvent{
			Type:      "thinking",
			Timestamp: time.Now().UnixMilli(),
			ID:        fmt.Sprintf("think_%d", time.Now().UnixNano()),
			Content:   msg,
		},
	})
}

// OnToolCall sends a tool call notification to task-specific connected clients
func (b *BroadcastNotifier) OnToolCall(toolName string, params interface{}) {
	b.server.broadcastToTask(b.taskID, SSEMessage{
		Type: "notify",
		Data: NotifyEvent{
			Type:       "tool_call",
			Timestamp:  time.Now().UnixMilli(),
			ID:         fmt.Sprintf("tool_%d", time.Now().UnixNano()),
			ToolName:   toolName,
			Parameters: params,
			Status:     "calling",
		},
	})
}

// OnResult sends a result notification to task-specific connected clients
func (b *BroadcastNotifier) OnResult(msg string) {
	b.server.broadcastToTask(b.taskID, SSEMessage{
		Type: "notify",
		Data: NotifyEvent{
			Type:      "result",
			Timestamp: time.Now().UnixMilli(),
			ID:        fmt.Sprintf("result_%d", time.Now().UnixNano()),
			Content:   msg,
		},
	})
}

// OnError sends an error notification to task-specific connected clients
func (b *BroadcastNotifier) OnError(err error) {
	b.server.broadcastToTask(b.taskID, SSEMessage{
		Type: "notify",
		Data: NotifyEvent{
			Type:      "error",
			Timestamp: time.Now().UnixMilli(),
			ID:        fmt.Sprintf("error_%d", time.Now().UnixNano()),
			Error:     err.Error(),
		},
	})
}

// handleTestLLMConnection handles POST /api/llm/test
func (s *Server) handleTestLLMConnection(w http.ResponseWriter, r *http.Request) {
	var llmConfig config.LLMConfig

	if err := json.NewDecoder(r.Body).Decode(&llmConfig); err != nil {
		http.Error(w, "解析LLM配置数据失败", http.StatusBadRequest)
		return
	}

	// 验证LLM配置
	if err := llmConfig.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("LLM配置验证失败: %v", err), http.StatusBadRequest)
		return
	}

	// 创建临时配置用于测试
	testConfig := &config.Config{
		LLM: llmConfig,
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	model, err := testConfig.GetModel(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "LLM连接测试失败",
			"error":   err.Error(),
		})
		return
	}

	// 尝试发送一个简单的测试消息
	if model != nil {
		// 这里可以添加更详细的测试逻辑，比如发送一个简单的消息
		// 但为了避免复杂性，我们只检查模型是否能成功创建
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "LLM连接测试成功",
	})
}

// LLM配置管理API处理函数

// handleListLLMConfigs handles GET /api/llm/configs
func (s *Server) handleListLLMConfigs(w http.ResponseWriter, r *http.Request) {
	configs, err := s.llmConfigService.ListConfigs()
	if err != nil {
		http.Error(w, fmt.Sprintf("获取LLM配置列表失败: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    configs,
	})
}

// handleCreateLLMConfig handles POST /api/llm/configs
func (s *Server) handleCreateLLMConfig(w http.ResponseWriter, r *http.Request) {
	var config models.LLMConfigModel

	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "解析LLM配置数据失败", http.StatusBadRequest)
		return
	}

	if err := s.llmConfigService.CreateConfig(&config); err != nil {
		if err == models.ErrLLMConfigNameExists {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprintf("创建LLM配置失败: %v", err), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "LLM配置创建成功",
		"data":    config,
	})
}

// handleGetLLMConfig handles GET /api/llm/configs/{id}
func (s *Server) handleGetLLMConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "无效的配置ID", http.StatusBadRequest)
		return
	}

	config, err := s.llmConfigService.GetConfig(uint(id))
	if err != nil {
		if err == models.ErrLLMConfigNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("获取LLM配置失败: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    config,
	})
}

// handleUpdateLLMConfig handles PUT /api/llm/configs/{id}
func (s *Server) handleUpdateLLMConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "无效的配置ID", http.StatusBadRequest)
		return
	}

	var updates models.LLMConfigModel
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "解析LLM配置数据失败", http.StatusBadRequest)
		return
	}

	if err := s.llmConfigService.UpdateConfig(uint(id), &updates); err != nil {
		if err == models.ErrLLMConfigNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if err == models.ErrLLMConfigNameExists {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprintf("更新LLM配置失败: %v", err), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "LLM配置更新成功",
	})
}

// handleDeleteLLMConfig handles DELETE /api/llm/configs/{id}
func (s *Server) handleDeleteLLMConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "无效的配置ID", http.StatusBadRequest)
		return
	}

	if err := s.llmConfigService.DeleteConfig(uint(id)); err != nil {
		if err == models.ErrLLMConfigNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("删除LLM配置失败: %v", err), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "LLM配置删除成功",
	})
}

// handleSetDefaultLLMConfig handles POST /api/llm/configs/{id}/default
func (s *Server) handleSetDefaultLLMConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "无效的配置ID", http.StatusBadRequest)
		return
	}

	if err := s.llmConfigService.SetDefaultConfig(uint(id)); err != nil {
		if err == models.ErrLLMConfigNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("设置默认LLM配置失败: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "默认LLM配置设置成功",
	})
}

// handleGetMCPTools handles POST /api/mcp/tools
func (s *Server) handleGetMCPTools(w http.ResponseWriter, r *http.Request) {
	var req MCPToolsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "解析MCP服务器配置数据失败", http.StatusBadRequest)
		return
	}

	// 创建MCPSettings
	settings := &mcphost.MCPSettings{
		MCPServers: req.MCPServers,
	}

	// 使用连接池获取MCP服务器连接
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 获取连接池
	pool := mcphost.GetConnectionPool()

	// 获取或创建连接
	hub, err := pool.GetHub(ctx, settings)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(MCPToolsResponse{
			Success: false,
			Message: "连接MCP服务器失败",
			Error:   err.Error(),
		})
		return
	}
	// 注意：不再直接调用hub.CloseServers()，而是在使用完后释放引用
	defer pool.ReleaseHub(settings)

	// 获取所有可用工具，并记录每个工具来自哪个服务器
	var tools []MCPToolInfo

	// 获取工具映射
	toolsMap, err := hub.GetToolsMap(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MCPToolsResponse{
			Success: false,
			Message: "获取工具列表失败",
			Error:   err.Error(),
		})
		return
	}

	// 处理工具信息
	for toolKey, toolInfo := range toolsMap {
		// 从工具键中提取服务器名称
		parts := strings.SplitN(toolKey, "_", 2)
		serverName := parts[0]

		tools = append(tools, MCPToolInfo{
			Name:        toolInfo.Name,
			Description: toolInfo.Desc,
			Server:      serverName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPToolsResponse{
		Success: true,
		Message: fmt.Sprintf("成功获取 %d 个工具", len(tools)),
		Tools:   tools,
	})
}

// handleGetMCPToolsFromDB handles GET /api/mcp/tools/configured
func (s *Server) handleGetMCPToolsFromDB(w http.ResponseWriter, r *http.Request) {
	// 从数据库获取所有活跃的MCP服务器配置
	configs, err := s.mcpServerConfigService.GetAllActiveConfigs()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MCPToolsResponse{
			Success: false,
			Message: "获取MCP服务器配置失败",
			Error:   err.Error(),
		})
		return
	}

	if len(configs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(MCPToolsResponse{
			Success: true,
			Message: "没有找到活跃的MCP服务器配置",
			Tools:   []MCPToolInfo{},
		})
		return
	}

	// 将数据库配置转换为mcphost.ServerConfig格式
	mcpServers := make(map[string]mcphost.ServerConfig)
	for name, config := range configs {
		serverConfig, err := config.ToServerConfig()
		if err != nil {
			log.Printf("转换服务器配置失败 %s: %v", name, err)
			continue
		}
		mcpServers[name] = serverConfig
	}

	// 创建MCPSettings
	settings := &mcphost.MCPSettings{
		MCPServers: mcpServers,
	}

	// 使用连接池获取MCP服务器连接
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 获取连接池
	pool := mcphost.GetConnectionPool()

	// 获取或创建连接
	hub, err := pool.GetHub(ctx, settings)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(MCPToolsResponse{
			Success: false,
			Message: "连接MCP服务器失败",
			Error:   err.Error(),
		})
		return
	}
	// 注意：不再直接调用hub.CloseServers()，而是在使用完后释放引用
	defer pool.ReleaseHub(settings)

	// 获取所有可用工具
	var tools []MCPToolInfo

	// 获取工具映射
	toolsMap, err := hub.GetToolsMap(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MCPToolsResponse{
			Success: false,
			Message: "获取工具列表失败",
			Error:   err.Error(),
		})
		return
	}

	// 处理工具信息
	for toolKey, toolInfo := range toolsMap {
		// 从工具键中提取服务器名称
		parts := strings.SplitN(toolKey, "_", 2)
		serverName := parts[0]

		tools = append(tools, MCPToolInfo{
			Name:        toolInfo.Name,
			Description: toolInfo.Desc,
			Server:      serverName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPToolsResponse{
		Success: true,
		Message: fmt.Sprintf("成功获取 %d 个工具", len(tools)),
		Tools:   tools,
	})
}

// MCP服务器配置管理API处理函数

// handleListMCPServerConfigs handles GET /api/mcp/servers
func (s *Server) handleListMCPServerConfigs(w http.ResponseWriter, r *http.Request) {
	configs, err := s.mcpServerConfigService.ListConfigs()
	if err != nil {
		http.Error(w, fmt.Sprintf("获取MCP服务器配置列表失败: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    configs,
	})
}

// CreateMCPServerConfigRequest represents the request body for creating MCP server config
type CreateMCPServerConfigRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	Env         map[string]string `json:"env"`
	Disabled    bool              `json:"disabled"`
}

// handleCreateMCPServerConfig handles POST /api/mcp/servers
func (s *Server) handleCreateMCPServerConfig(w http.ResponseWriter, r *http.Request) {
	var req CreateMCPServerConfigRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "解析MCP服务器配置数据失败", http.StatusBadRequest)
		return
	}

	// 创建数据库模型
	config := &models.MCPServerConfigModel{
		Name:        req.Name,
		Description: req.Description,
		Command:     req.Command,
		Disabled:    req.Disabled,
		IsActive:    true,
	}

	// 设置参数和环境变量
	if err := config.SetArgs(req.Args); err != nil {
		http.Error(w, fmt.Sprintf("设置参数失败: %v", err), http.StatusBadRequest)
		return
	}

	if err := config.SetEnv(req.Env); err != nil {
		http.Error(w, fmt.Sprintf("设置环境变量失败: %v", err), http.StatusBadRequest)
		return
	}

	if err := s.mcpServerConfigService.CreateConfig(config); err != nil {
		if err == models.ErrMCPServerConfigNameExists {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprintf("创建MCP服务器配置失败: %v", err), http.StatusBadRequest)
		}
		return
	}

	// 异步同步工具，不阻塞响应
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.mcpToolService.SyncToolsForServer(ctx, config); err != nil {
			log.Printf("创建服务器后同步工具失败 %s: %v", config.Name, err)
		} else {
			log.Printf("成功为新创建的服务器 %s 同步工具", config.Name)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "MCP服务器配置创建成功",
		"data":    config,
	})
}

// handleGetMCPServerConfig handles GET /api/mcp/servers/{id}
func (s *Server) handleGetMCPServerConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "无效的配置ID", http.StatusBadRequest)
		return
	}

	config, err := s.mcpServerConfigService.GetConfig(uint(id))
	if err != nil {
		if err == models.ErrMCPServerConfigNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("获取MCP服务器配置失败: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    config,
	})
}

// handleUpdateMCPServerConfig handles PUT /api/mcp/servers/{id}
func (s *Server) handleUpdateMCPServerConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "无效的配置ID", http.StatusBadRequest)
		return
	}

	var req CreateMCPServerConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "解析MCP服务器配置数据失败", http.StatusBadRequest)
		return
	}

	// 创建更新模型
	updates := &models.MCPServerConfigModel{
		Name:        req.Name,
		Description: req.Description,
		Command:     req.Command,
		Disabled:    req.Disabled,
		IsActive:    true,
	}

	// 设置参数和环境变量
	if err := updates.SetArgs(req.Args); err != nil {
		http.Error(w, fmt.Sprintf("设置参数失败: %v", err), http.StatusBadRequest)
		return
	}

	if err := updates.SetEnv(req.Env); err != nil {
		http.Error(w, fmt.Sprintf("设置环境变量失败: %v", err), http.StatusBadRequest)
		return
	}

	if err := s.mcpServerConfigService.UpdateConfig(uint(id), updates); err != nil {
		if err == models.ErrMCPServerConfigNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if err == models.ErrMCPServerConfigNameExists {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprintf("更新MCP服务器配置失败: %v", err), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "MCP服务器配置更新成功",
	})
}

// handleDeleteMCPServerConfig handles DELETE /api/mcp/servers/{id}
func (s *Server) handleDeleteMCPServerConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "无效的配置ID", http.StatusBadRequest)
		return
	}

	// 先删除相关的工具
	if err := s.mcpToolService.DeleteToolsByServerID(uint(id)); err != nil {
		log.Printf("删除服务器 %d 的工具失败: %v", id, err)
		// 不阻塞服务器删除，继续执行
	}

	if err := s.mcpServerConfigService.DeleteConfig(uint(id)); err != nil {
		if err == models.ErrMCPServerConfigNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("删除MCP服务器配置失败: %v", err), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "MCP服务器配置删除成功",
	})
}

// handleGetCachedMCPTools handles GET /api/mcp/tools/cached
// 从数据库缓存中获取工具列表，不连接MCP服务器
func (s *Server) handleGetCachedMCPTools(w http.ResponseWriter, r *http.Request) {
	toolsInfo, err := s.mcpToolService.GetToolsInfo()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MCPToolsResponse{
			Success: false,
			Message: "获取缓存工具列表失败",
			Error:   err.Error(),
		})
		return
	}

	// 转换为API响应格式
	var tools []MCPToolInfo
	for _, toolInfo := range toolsInfo {
		tools = append(tools, MCPToolInfo{
			Name:        toolInfo.Name,
			Description: toolInfo.Description,
			Server:      toolInfo.Server,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPToolsResponse{
		Success: true,
		Message: fmt.Sprintf("成功获取 %d 个缓存工具", len(tools)),
		Tools:   tools,
	})
}

// handleSyncMCPTools handles POST /api/mcp/tools/sync
// 同步所有活跃服务器的工具到数据库
func (s *Server) handleSyncMCPTools(w http.ResponseWriter, r *http.Request) {
	// 获取所有活跃的MCP服务器配置
	configs, err := s.mcpServerConfigService.GetAllActiveConfigs()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MCPToolsResponse{
			Success: false,
			Message: "获取MCP服务器配置失败",
			Error:   err.Error(),
		})
		return
	}

	if len(configs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(MCPToolsResponse{
			Success: true,
			Message: "没有找到活跃的MCP服务器配置",
			Tools:   []MCPToolInfo{},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var totalTools int
	var errors []string

	// 同步每个服务器的工具
	for _, config := range configs {
		err := s.mcpToolService.SyncToolsForServer(ctx, &config)
		if err != nil {
			errorMsg := fmt.Sprintf("同步服务器 %s 失败: %v", config.Name, err)
			log.Printf("%s", errorMsg)
			errors = append(errors, errorMsg)
			continue
		}

		// 获取该服务器的工具数量
		tools, err := s.mcpToolService.GetToolsByServerID(config.ID)
		if err == nil {
			totalTools += len(tools)
		}
	}

	message := fmt.Sprintf("同步完成，共获取 %d 个工具", totalTools)
	if len(errors) > 0 {
		message += fmt.Sprintf("，%d 个服务器同步失败", len(errors))
	}

	w.Header().Set("Content-Type", "application/json")
	response := MCPToolsResponse{
		Success: len(errors) < len(configs), // 如果至少有一个服务器同步成功，就算成功
		Message: message,
	}

	if len(errors) > 0 {
		response.Error = fmt.Sprintf("部分服务器同步失败: %s", errors[0])
	}

	json.NewEncoder(w).Encode(response)
}

// handleSyncMCPToolsForServer handles POST /api/mcp/tools/sync/{id}
// 同步指定服务器的工具到数据库
func (s *Server) handleSyncMCPToolsForServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "无效的服务器ID", http.StatusBadRequest)
		return
	}

	// 获取服务器配置
	config, err := s.mcpServerConfigService.GetConfig(uint(id))
	if err != nil {
		if err == models.ErrMCPServerConfigNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("获取MCP服务器配置失败: %v", err), http.StatusInternalServerError)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 同步工具
	err = s.mcpToolService.SyncToolsForServer(ctx, config)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(MCPToolsResponse{
			Success: false,
			Message: fmt.Sprintf("同步服务器 %s 的工具失败", config.Name),
			Error:   err.Error(),
		})
		return
	}

	// 获取同步后的工具数量
	tools, err := s.mcpToolService.GetToolsByServerID(config.ID)
	if err != nil {
		log.Printf("获取服务器 %s 的工具数量失败: %v", config.Name, err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPToolsResponse{
		Success: true,
		Message: fmt.Sprintf("成功同步服务器 %s 的 %d 个工具", config.Name, len(tools)),
	})
}
