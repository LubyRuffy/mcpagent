// Package models provides database models for the MCP Agent application.
// It defines the data structures used for persistent storage of application configuration.
package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// MCPConfig 存储MCP的配置信息
type MCPConfig struct {
	ConfigFile string   `json:"config_file"`
	Tools      []string `json:"tools"`
}

// AppConfigModel represents a saved application configuration in the database.
// It stores the global application settings for MCP Agent.
type AppConfigModel struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	Name         string         `gorm:"uniqueIndex;not null" json:"name"`           // 配置名称，如 "default"
	Description  string         `gorm:"type:text" json:"description"`               // 配置描述
	Proxy        string         `json:"proxy"`                                      // 代理配置
	SystemPrompt string         `gorm:"type:text" json:"system_prompt"`             // 系统提示词
	MaxStep      int            `gorm:"default:20" json:"max_step"`                 // 最大步数
	PlaceHolders string         `gorm:"type:json;default:'{}'" json:"placeholders"` // 占位符，JSON格式存储
	MCPSettings  string         `gorm:"type:json;default:'{}'" json:"mcp_settings"` // MCP配置，JSON格式存储
	IsDefault    bool           `gorm:"default:false" json:"is_default"`            // 是否为默认配置
	IsActive     bool           `gorm:"default:true" json:"is_active"`              // 是否启用
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for AppConfigModel
func (AppConfigModel) TableName() string {
	return "app_configs"
}

// Validate validates the application configuration model.
// It ensures all required fields are present and valid.
func (a *AppConfigModel) Validate() error {
	if a.Name == "" {
		return ErrAppConfigNameEmpty
	}
	if a.MaxStep <= 0 {
		return ErrAppConfigMaxStepInvalid
	}
	return nil
}

// GetPlaceHoldersAsMap returns the placeholders as a map
func (a *AppConfigModel) GetPlaceHoldersAsMap() (map[string]interface{}, error) {
	if a.PlaceHolders == "" {
		return map[string]interface{}{}, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(a.PlaceHolders), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SetPlaceHoldersFromMap sets the placeholders from a map
func (a *AppConfigModel) SetPlaceHoldersFromMap(placeholders map[string]interface{}) error {
	if placeholders == nil {
		a.PlaceHolders = "{}"
		return nil
	}

	data, err := json.Marshal(placeholders)
	if err != nil {
		return err
	}
	a.PlaceHolders = string(data)
	return nil
}

// GetMCPConfig returns the MCP configuration
func (a *AppConfigModel) GetMCPConfig() (*MCPConfig, error) {
	if a.MCPSettings == "" {
		return &MCPConfig{
			Tools: []string{},
		}, nil
	}

	var result MCPConfig
	if err := json.Unmarshal([]byte(a.MCPSettings), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetMCPConfig sets the MCP configuration
func (a *AppConfigModel) SetMCPConfig(mcpConfig *MCPConfig) error {
	if mcpConfig == nil {
		a.MCPSettings = "{}"
		return nil
	}

	data, err := json.Marshal(mcpConfig)
	if err != nil {
		return err
	}
	a.MCPSettings = string(data)
	return nil
}
