// Package services provides business logic services for the MCP Agent application.
// It implements the service layer that handles business operations and data validation.
package services

import (
	"fmt"

	"github.com/LubyRuffy/mcpagent/pkg/database"
	"github.com/LubyRuffy/mcpagent/pkg/models"
	"gorm.io/gorm"
)

// SystemPromptService provides business logic for system prompt configuration management
type SystemPromptService struct {
	db *gorm.DB
}

// NewSystemPromptService creates a new system prompt service instance
func NewSystemPromptService() *SystemPromptService {
	return &SystemPromptService{
		db: database.GetDB(),
	}
}

// ListPrompts returns all active system prompt configurations
func (s *SystemPromptService) ListPrompts() ([]models.SystemPromptModel, error) {
	var prompts []models.SystemPromptModel
	err := s.db.Where("is_active = ?", true).Order("is_default DESC, created_at ASC").Find(&prompts).Error
	return prompts, err
}

// GetPrompt returns a specific system prompt configuration by ID
func (s *SystemPromptService) GetPrompt(id uint) (*models.SystemPromptModel, error) {
	var prompt models.SystemPromptModel
	err := s.db.Where("id = ? AND is_active = ?", id, true).First(&prompt).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrSystemPromptNotFound
		}
		return nil, err
	}
	return &prompt, nil
}

// GetDefaultPrompt returns the default system prompt configuration
func (s *SystemPromptService) GetDefaultPrompt() (*models.SystemPromptModel, error) {
	var prompt models.SystemPromptModel
	err := s.db.Where("is_default = ? AND is_active = ?", true, true).First(&prompt).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrSystemPromptNotFound
		}
		return nil, err
	}
	return &prompt, nil
}

// CreatePrompt creates a new system prompt configuration
func (s *SystemPromptService) CreatePrompt(prompt *models.SystemPromptModel) error {
	// 验证配置
	if err := prompt.Validate(); err != nil {
		return err
	}

	// 检查名称是否已存在
	var count int64
	err := s.db.Model(&models.SystemPromptModel{}).Where("name = ? AND is_active = ?", prompt.Name, true).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return models.ErrSystemPromptNameExists
	}

	// 如果设置为默认配置，需要先取消其他默认配置
	if prompt.IsDefault {
		if err := s.clearDefaultPrompts(); err != nil {
			return err
		}
	}

	// 创建配置
	return s.db.Create(prompt).Error
}

// UpdatePrompt updates an existing system prompt configuration
func (s *SystemPromptService) UpdatePrompt(id uint, updates *models.SystemPromptModel) error {
	// 验证更新数据
	if err := updates.Validate(); err != nil {
		return err
	}

	// 检查配置是否存在
	var existingPrompt models.SystemPromptModel
	err := s.db.Where("id = ? AND is_active = ?", id, true).First(&existingPrompt).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.ErrSystemPromptNotFound
		}
		return err
	}

	// 检查名称是否与其他配置冲突
	if updates.Name != existingPrompt.Name {
		var count int64
		err := s.db.Model(&models.SystemPromptModel{}).Where("name = ? AND id != ? AND is_active = ?", updates.Name, id, true).Count(&count).Error
		if err != nil {
			return err
		}
		if count > 0 {
			return models.ErrSystemPromptNameExists
		}
	}

	// 如果设置为默认配置，需要先取消其他默认配置
	if updates.IsDefault && !existingPrompt.IsDefault {
		if err := s.clearDefaultPrompts(); err != nil {
			return err
		}
	}

	// 更新配置
	updates.ID = id
	return s.db.Model(&existingPrompt).Updates(updates).Error
}

// DeletePrompt soft deletes a system prompt configuration
func (s *SystemPromptService) DeletePrompt(id uint) error {
	// 检查配置是否存在
	var prompt models.SystemPromptModel
	err := s.db.Where("id = ? AND is_active = ?", id, true).First(&prompt).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.ErrSystemPromptNotFound
		}
		return err
	}

	// 检查是否为默认配置
	if prompt.IsDefault {
		return fmt.Errorf("不能删除默认配置")
	}

	// 软删除配置
	return s.db.Model(&prompt).Update("is_active", false).Error
}

// SetDefaultPrompt sets a configuration as the default one
func (s *SystemPromptService) SetDefaultPrompt(id uint) error {
	// 检查配置是否存在
	var prompt models.SystemPromptModel
	err := s.db.Where("id = ? AND is_active = ?", id, true).First(&prompt).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.ErrSystemPromptNotFound
		}
		return err
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 取消所有默认配置
	if err := tx.Model(&models.SystemPromptModel{}).Where("is_active = ?", true).Update("is_default", false).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 设置新的默认配置
	if err := tx.Model(&prompt).Update("is_default", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// clearDefaultPrompts removes default flag from all configurations
func (s *SystemPromptService) clearDefaultPrompts() error {
	return s.db.Model(&models.SystemPromptModel{}).Where("is_active = ?", true).Update("is_default", false).Error
}
