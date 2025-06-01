// Package models provides database models for the MCP Agent application.
// It defines the data structures used for persistent storage of application configuration.
package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// SystemPromptModel represents a saved system prompt configuration in the database.
// It stores predefined system prompts that can be used in agent tasks.
type SystemPromptModel struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	Name         string         `gorm:"uniqueIndex;not null" json:"name"`           // 配置名称，用于用户识别
	Description  string         `gorm:"type:text" json:"description"`               // 配置描述
	Content      string         `gorm:"type:text;not null" json:"content"`          // 提示词内容
	Placeholders string         `gorm:"type:json;default:'[]'" json:"placeholders"` // 提示词中的占位符列表，JSON格式存储
	IsDefault    bool           `gorm:"default:false" json:"is_default"`            // 是否为默认配置
	IsActive     bool           `gorm:"default:true" json:"is_active"`              // 是否启用
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for SystemPromptModel
func (SystemPromptModel) TableName() string {
	return "system_prompts"
}

// Validate validates the system prompt configuration model.
// It ensures all required fields are present and valid.
func (s *SystemPromptModel) Validate() error {
	if s.Name == "" {
		return ErrSystemPromptNameEmpty
	}
	if s.Content == "" {
		return ErrSystemPromptContentEmpty
	}
	return nil
}

// GetPlaceholdersAsStringSlice returns the placeholders as a string slice
func (s *SystemPromptModel) GetPlaceholdersAsStringSlice() ([]string, error) {
	var placeholders []string
	if s.Placeholders == "" {
		return []string{}, nil
	}

	// 从JSON字符串解析为字符串数组
	if err := json.Unmarshal([]byte(s.Placeholders), &placeholders); err != nil {
		return nil, err
	}

	return placeholders, nil
}

// SetPlaceholdersFromStringSlice sets the placeholders from a string slice
func (s *SystemPromptModel) SetPlaceholdersFromStringSlice(placeholders []string) error {
	jsonData, err := json.Marshal(placeholders)
	if err != nil {
		return err
	}
	s.Placeholders = string(jsonData)
	return nil
}

// ToSystemPromptConfig converts the database model to a map for use in the config.
func (s *SystemPromptModel) ToSystemPromptConfig() map[string]interface{} {
	placeholders, _ := s.GetPlaceholdersAsStringSlice()

	return map[string]interface{}{
		"id":           s.ID,
		"name":         s.Name,
		"description":  s.Description,
		"content":      s.Content,
		"placeholders": placeholders,
		"is_default":   s.IsDefault,
		"is_active":    s.IsActive,
	}
}

// FromSystemPromptConfig populates the model from configuration data.
func (s *SystemPromptModel) FromSystemPromptConfig(config map[string]interface{}) error {
	if v, ok := config["name"].(string); ok {
		s.Name = v
	}

	if v, ok := config["description"].(string); ok {
		s.Description = v
	}

	if v, ok := config["content"].(string); ok {
		s.Content = v
	}

	if v, ok := config["is_default"].(bool); ok {
		s.IsDefault = v
	}

	// 处理占位符数组
	if placeholders, ok := config["placeholders"].([]string); ok {
		return s.SetPlaceholdersFromStringSlice(placeholders)
	} else if placeholders, ok := config["placeholders"].([]interface{}); ok {
		strPlaceholders := make([]string, 0, len(placeholders))
		for _, p := range placeholders {
			if str, ok := p.(string); ok {
				strPlaceholders = append(strPlaceholders, str)
			}
		}
		return s.SetPlaceholdersFromStringSlice(strPlaceholders)
	}

	return nil
}
