// Package models provides database models for the MCP Agent application.
// It defines the data structures used for persistent storage of MCP tools.
package models

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

// MCPToolModel represents a MCP tool stored in the database.
// It stores tool information with metadata for management and caching.
type MCPToolModel struct {
	ID          uint                 `gorm:"primarykey" json:"id"`
	Name        string               `gorm:"not null;index" json:"name"`           // 工具名称
	Description string               `gorm:"type:text" json:"description"`         // 工具描述
	ServerID    uint                 `gorm:"not null;index" json:"server_id"`      // 关联的MCP服务器ID
	Server      MCPServerConfigModel `gorm:"foreignKey:ServerID" json:"server"`    // 关联的MCP服务器
	InputSchema string               `gorm:"type:text" json:"input_schema"`        // 输入模式（JSON格式存储）
	ToolKey     string               `gorm:"uniqueIndex;not null" json:"tool_key"` // 工具唯一标识（server_name + "_" + tool_name）
	IsActive    bool                 `json:"is_active"`                            // 是否启用
	LastSyncAt  *time.Time           `json:"last_sync_at"`                         // 最后同步时间
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	DeletedAt   gorm.DeletedAt       `gorm:"index" json:"-"`
}

// TableName returns the table name for MCPToolModel
func (MCPToolModel) TableName() string {
	return "mcp_tools"
}

// Validate validates the MCP tool
func (m *MCPToolModel) Validate() error {
	if m.Name == "" {
		return ErrMCPToolNameEmpty
	}
	if m.ServerID == 0 {
		return ErrMCPToolServerIDEmpty
	}
	if m.ToolKey == "" {
		return ErrMCPToolKeyEmpty
	}
	return nil
}

// SetInputSchema sets the input schema from a map
func (m *MCPToolModel) SetInputSchema(schema map[string]interface{}) error {
	if schema == nil {
		m.InputSchema = ""
		return nil
	}

	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return err
	}
	m.InputSchema = string(schemaJSON)
	return nil
}

// GetInputSchema returns the input schema as a map
func (m *MCPToolModel) GetInputSchema() (map[string]interface{}, error) {
	if m.InputSchema == "" {
		return nil, nil
	}

	var schema map[string]interface{}
	err := json.Unmarshal([]byte(m.InputSchema), &schema)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

// GenerateToolKey generates the tool key from server name and tool name
func GenerateToolKey(serverName, toolName string) string {
	// 检查并处理特殊字符
	if strings.Contains(toolName, "@") {
		// 如果工具名称中已包含@符号，可能是已经格式化为 "name@server" 的格式
		parts := strings.SplitN(toolName, "@", 2)
		if len(parts) == 2 && (serverName == "" || serverName == parts[1]) {
			toolName = parts[0]
			if serverName == "" {
				serverName = parts[1]
			}
		}
		log.Printf("【工具键生成】工具名包含@符号，分解为工具名=%s，服务器=%s", toolName, serverName)
	}

	// 确保服务器名称不为空，默认为"inner"
	if serverName == "" {
		serverName = "inner"
		log.Printf("【工具键生成】服务器名为空，使用默认值 inner")
	}

	// 生成并返回工具键
	toolKey := serverName + "_" + toolName
	log.Printf("【工具键生成】最终生成的工具键: %s", toolKey)
	return toolKey
}

// MCPToolInfo represents tool information for API responses
type MCPToolInfo struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Server      string     `json:"server"`
	ToolKey     string     `json:"tool_key"`
	IsActive    bool       `json:"is_active"`
	LastSyncAt  *time.Time `json:"last_sync_at,omitempty"`
}

// ToMCPToolInfo converts MCPToolModel to MCPToolInfo
func (m *MCPToolModel) ToMCPToolInfo() MCPToolInfo {
	serverName := ""
	if m.Server.ID != 0 {
		serverName = m.Server.Name
	} else {
		// 从 ToolKey 中提取服务器名称（格式为 serverName_toolName）
		parts := strings.SplitN(m.ToolKey, "_", 2)
		if len(parts) > 0 {
			serverName = parts[0]
		}
	}

	return MCPToolInfo{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Server:      serverName,
		ToolKey:     m.ToolKey,
		IsActive:    m.IsActive,
		LastSyncAt:  m.LastSyncAt,
	}
}
