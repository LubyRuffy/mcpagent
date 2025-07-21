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
type MCPServerConfigModel struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"uniqueIndex;not null" json:"name"` // 服务器名称，用于用户识别
	Description string         `gorm:"type:text" json:"description"`     // 服务器描述
	Command     string         `gorm:"not null" json:"command"`          // 启动命令
	Args        string         `gorm:"type:text" json:"args"`            // 参数列表（JSON格式存储）
	Env         string         `gorm:"type:text" json:"env"`             // 环境变量（JSON格式存储）
	Disabled    bool           `gorm:"default:false" json:"disabled"`    // 是否禁用
	IsActive    bool           `gorm:"default:true" json:"is_active"`    // 是否启用
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
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
	if m.Command == "" {
		return ErrMCPServerConfigCommandEmpty
	}
	return nil
}

// ToServerConfig converts the database model to mcphost.ServerConfig
func (m *MCPServerConfigModel) ToServerConfig() (einomcphost.ServerConfig, error) {
	config := einomcphost.ServerConfig{
		Command:  m.Command,
		Disabled: m.Disabled,
	}

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
