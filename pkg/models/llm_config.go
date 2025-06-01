// Package models provides database models for the MCP Agent application.
// It defines the data structures used for persistent storage of application configuration.
package models

import (
	"time"

	"gorm.io/gorm"
)

// LLMConfigModel represents a saved LLM configuration in the database.
// It extends the basic LLM configuration with metadata for management.
type LLMConfigModel struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"uniqueIndex;not null" json:"name"`        // 配置名称，用于用户识别
	Description string         `gorm:"type:text" json:"description"`            // 配置描述
	Type        string         `gorm:"not null" json:"type"`                    // LLM类型：openai, ollama
	BaseURL     string         `gorm:"not null" json:"base_url"`                // API基础URL
	Model       string         `gorm:"not null" json:"model"`                   // 模型名称
	APIKey      string         `gorm:"not null" json:"api_key"`                 // API密钥
	Temperature *float64       `json:"temperature,omitempty"`                   // 温度参数
	MaxTokens   *int           `json:"max_tokens,omitempty"`                    // 最大token数
	IsDefault   bool           `gorm:"default:false" json:"is_default"`         // 是否为默认配置
	IsActive    bool           `gorm:"default:true" json:"is_active"`           // 是否启用
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for LLMConfigModel
func (LLMConfigModel) TableName() string {
	return "llm_configs"
}

// Validate validates the LLM configuration model.
// It ensures all required fields are present and valid.
func (l *LLMConfigModel) Validate() error {
	if l.Name == "" {
		return ErrLLMConfigNameEmpty
	}
	if l.Type == "" {
		return ErrLLMConfigTypeEmpty
	}
	if l.Type != "openai" && l.Type != "ollama" {
		return ErrLLMConfigTypeInvalid
	}
	if l.BaseURL == "" {
		return ErrLLMConfigBaseURLEmpty
	}
	if l.Model == "" {
		return ErrLLMConfigModelEmpty
	}
	if l.APIKey == "" {
		return ErrLLMConfigAPIKeyEmpty
	}
	return nil
}

// ToConfigLLM converts the database model to the config package's LLMConfig struct.
func (l *LLMConfigModel) ToConfigLLM() map[string]interface{} {
	config := map[string]interface{}{
		"type":     l.Type,
		"base_url": l.BaseURL,
		"model":    l.Model,
		"api_key":  l.APIKey,
	}
	
	if l.Temperature != nil {
		config["temperature"] = *l.Temperature
	}
	if l.MaxTokens != nil {
		config["max_tokens"] = *l.MaxTokens
	}
	
	return config
}

// FromConfigLLM creates a LLMConfigModel from basic LLM configuration data.
func (l *LLMConfigModel) FromConfigLLM(name, description string, config map[string]interface{}) {
	l.Name = name
	l.Description = description
	
	if v, ok := config["type"].(string); ok {
		l.Type = v
	}
	if v, ok := config["base_url"].(string); ok {
		l.BaseURL = v
	}
	if v, ok := config["model"].(string); ok {
		l.Model = v
	}
	if v, ok := config["api_key"].(string); ok {
		l.APIKey = v
	}
	if v, ok := config["temperature"].(float64); ok {
		l.Temperature = &v
	}
	if v, ok := config["max_tokens"].(int); ok {
		l.MaxTokens = &v
	}
}
