// Package models provides database models for the MCP Agent application.
// It defines the data structures used for persistent storage of MCP server configuration.
package models

import (
	"encoding/json"
	"time"

	"github.com/LubyRuffy/einomcphost"
	"gorm.io/gorm"
)

// MCPServerConfigModel represents a saved MCP server configuration in the database.
// It stores MCP server settings with metadata for management.
// Supports both STDIO and SSE transport types.
type MCPServerConfigModel struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	Name          string         `gorm:"uniqueIndex;not null" json:"name"`               // 服务器名称，用于用户识别
	Description   string         `gorm:"type:text" json:"description"`                   // 服务器描述
	TransportType string         `gorm:"not null;default:'stdio'" json:"transport_type"` // 传输类型：stdio 或 sse
	Command       string         `json:"command"`                                        // 启动命令（stdio类型必需）
	Args          string         `gorm:"type:text" json:"args"`                          // 参数列表（JSON格式存储，stdio类型使用）
	Env           string         `gorm:"type:text" json:"env"`                           // 环境变量（JSON格式存储，stdio类型使用）
	URL           string         `json:"url"`                                            // SSE服务器URL（sse类型必需）
	Headers       string         `gorm:"type:text" json:"headers"`                       // HTTP头部（JSON格式存储，sse类型使用）
	Disabled      bool           `gorm:"default:false" json:"disabled"`                  // 是否禁用
	IsActive      bool           `gorm:"default:true" json:"is_active"`                  // 是否启用
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for MCPServerConfigModel
func (MCPServerConfigModel) TableName() string {
	return "mcp_server_configs"
}

// Validate validates the MCP server configuration
func (m *MCPServerConfigModel) Validate() error {
	if m.Name == "" {
		return ErrMCPServerConfigNameEmpty
	}

	// 验证传输类型
	if m.TransportType == "" {
		m.TransportType = "stdio" // 默认为stdio以保持向后兼容
	}

	switch m.TransportType {
	case "stdio":
		if m.Command == "" {
			return ErrMCPServerConfigCommandEmpty
		}
	case "sse", "http":
		if m.URL == "" {
			return ErrMCPServerConfigURLEmpty
		}
	default:
		return ErrMCPServerConfigInvalidTransportType
	}

	return nil
}

// ToServerConfig converts the database model to mcphost.ServerConfig
func (m *MCPServerConfigModel) ToServerConfig() (einomcphost.ServerConfig, error) {
	// 设置默认传输类型以保持向后兼容
	transportType := m.TransportType
	if transportType == "" {
		transportType = "stdio"
	}

	config := einomcphost.ServerConfig{
		TransportType: transportType,
		Disabled:      m.Disabled,
	}

	switch transportType {
	case "stdio":
		config.Command = m.Command

		// 解析参数列表
		if m.Args != "" {
			var args []string
			if err := json.Unmarshal([]byte(m.Args), &args); err != nil {
				return config, err
			}
			config.Args = args
		}

		// 解析环境变量
		if m.Env != "" {
			var env map[string]string
			if err := json.Unmarshal([]byte(m.Env), &env); err != nil {
				return config, err
			}
			config.Env = env
		}

	case "sse", "http":
		config.URL = m.URL

		// 注意: einomcphost.ServerConfig 不支持Headers字段
		// Headers信息目前存储在数据库中但不会传递给einomcphost
		// 如果未来einomcphost支持Headers，可以在这里添加解析逻辑
	}

	return config, nil
}

// FromServerConfig populates the model from mcphost.ServerConfig
func (m *MCPServerConfigModel) FromServerConfig(name, description string, config einomcphost.ServerConfig) error {
	m.Name = name
	m.Description = description
	m.Command = config.Command
	m.Disabled = config.Disabled

	// 序列化参数列表
	if len(config.Args) > 0 {
		argsJSON, err := json.Marshal(config.Args)
		if err != nil {
			return err
		}
		m.Args = string(argsJSON)
	}

	// 序列化环境变量
	if len(config.Env) > 0 {
		envJSON, err := json.Marshal(config.Env)
		if err != nil {
			return err
		}
		m.Env = string(envJSON)
	}

	return nil
}

// GetArgsSlice returns the args as a slice
func (m *MCPServerConfigModel) GetArgsSlice() ([]string, error) {
	if m.Args == "" {
		return []string{}, nil
	}

	var args []string
	if err := json.Unmarshal([]byte(m.Args), &args); err != nil {
		return nil, err
	}
	return args, nil
}

// GetEnvMap returns the env as a map
func (m *MCPServerConfigModel) GetEnvMap() (map[string]string, error) {
	if m.Env == "" {
		return map[string]string{}, nil
	}

	var env map[string]string
	if err := json.Unmarshal([]byte(m.Env), &env); err != nil {
		return nil, err
	}
	return env, nil
}

// SetArgs sets the args from a slice
func (m *MCPServerConfigModel) SetArgs(args []string) error {
	if len(args) == 0 {
		m.Args = ""
		return nil
	}

	argsJSON, err := json.Marshal(args)
	if err != nil {
		return err
	}
	m.Args = string(argsJSON)
	return nil
}

// SetEnv sets the env from a map
func (m *MCPServerConfigModel) SetEnv(env map[string]string) error {
	if len(env) == 0 {
		m.Env = ""
		return nil
	}

	envJSON, err := json.Marshal(env)
	if err != nil {
		return err
	}
	m.Env = string(envJSON)
	return nil
}

// GetHeadersSlice returns the headers as a slice
func (m *MCPServerConfigModel) GetHeadersSlice() ([]string, error) {
	if m.Headers == "" {
		return []string{}, nil
	}

	var headers []string
	if err := json.Unmarshal([]byte(m.Headers), &headers); err != nil {
		return nil, err
	}
	return headers, nil
}

// SetHeaders sets the headers from a slice
func (m *MCPServerConfigModel) SetHeaders(headers []string) error {
	if len(headers) == 0 {
		m.Headers = ""
		return nil
	}

	headersJSON, err := json.Marshal(headers)
	if err != nil {
		return err
	}
	m.Headers = string(headersJSON)
	return nil
}
