// Package services provides business logic services for the MCP Agent application.
// It implements the service layer that handles business operations and data validation.
package services

import (
	"fmt"

	"github.com/LubyRuffy/mcpagent/pkg/database"
	"github.com/LubyRuffy/mcpagent/pkg/models"
	"gorm.io/gorm"
)

// LLMConfigService provides business logic for LLM configuration management
type LLMConfigService struct {
	db *gorm.DB
}

// NewLLMConfigService creates a new LLM configuration service instance
func NewLLMConfigService() *LLMConfigService {
	return &LLMConfigService{
		db: database.GetDB(),
	}
}

// ListConfigs returns all active LLM configurations
func (s *LLMConfigService) ListConfigs() ([]models.LLMConfigModel, error) {
	var configs []models.LLMConfigModel
	err := s.db.Where("is_active = ?", true).Order("is_default DESC, created_at ASC").Find(&configs).Error
	return configs, err
}

// GetConfig returns a specific LLM configuration by ID
func (s *LLMConfigService) GetConfig(id uint) (*models.LLMConfigModel, error) {
	var config models.LLMConfigModel
	err := s.db.Where("id = ? AND is_active = ?", id, true).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrLLMConfigNotFound
		}
		return nil, err
	}
	return &config, nil
}

// GetDefaultConfig returns the default LLM configuration
func (s *LLMConfigService) GetDefaultConfig() (*models.LLMConfigModel, error) {
	var config models.LLMConfigModel
	err := s.db.Where("is_default = ? AND is_active = ?", true, true).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrLLMConfigNotFound
		}
		return nil, err
	}
	return &config, nil
}

// CreateConfig creates a new LLM configuration
func (s *LLMConfigService) CreateConfig(config *models.LLMConfigModel) error {
	// 验证配置
	if err := config.Validate(); err != nil {
		return err
	}

	// 检查名称是否已存在
	var count int64
	err := s.db.Model(&models.LLMConfigModel{}).Where("name = ? AND is_active = ?", config.Name, true).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return models.ErrLLMConfigNameExists
	}

	// 如果设置为默认配置，需要先取消其他默认配置
	if config.IsDefault {
		if err := s.clearDefaultConfigs(); err != nil {
			return err
		}
	}

	// 创建配置
	return s.db.Create(config).Error
}

// UpdateConfig updates an existing LLM configuration
func (s *LLMConfigService) UpdateConfig(id uint, updates *models.LLMConfigModel) error {
	// 验证更新数据
	if err := updates.Validate(); err != nil {
		return err
	}

	// 检查配置是否存在
	var existingConfig models.LLMConfigModel
	err := s.db.Where("id = ? AND is_active = ?", id, true).First(&existingConfig).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.ErrLLMConfigNotFound
		}
		return err
	}

	// 检查名称是否与其他配置冲突
	if updates.Name != existingConfig.Name {
		var count int64
		err := s.db.Model(&models.LLMConfigModel{}).Where("name = ? AND id != ? AND is_active = ?", updates.Name, id, true).Count(&count).Error
		if err != nil {
			return err
		}
		if count > 0 {
			return models.ErrLLMConfigNameExists
		}
	}

	// 如果设置为默认配置，需要先取消其他默认配置
	if updates.IsDefault && !existingConfig.IsDefault {
		if err := s.clearDefaultConfigs(); err != nil {
			return err
		}
	}

	// 更新配置
	updates.ID = id
	return s.db.Model(&existingConfig).Updates(updates).Error
}

// DeleteConfig soft deletes an LLM configuration
func (s *LLMConfigService) DeleteConfig(id uint) error {
	// 检查配置是否存在
	var config models.LLMConfigModel
	err := s.db.Where("id = ? AND is_active = ?", id, true).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.ErrLLMConfigNotFound
		}
		return err
	}

	// 检查是否为默认配置
	if config.IsDefault {
		return fmt.Errorf("不能删除默认配置")
	}

	// 软删除配置
	return s.db.Model(&config).Update("is_active", false).Error
}

// SetDefaultConfig sets a configuration as the default one
func (s *LLMConfigService) SetDefaultConfig(id uint) error {
	// 检查配置是否存在
	var config models.LLMConfigModel
	err := s.db.Where("id = ? AND is_active = ?", id, true).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.ErrLLMConfigNotFound
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
	if err := tx.Model(&models.LLMConfigModel{}).Where("is_active = ?", true).Update("is_default", false).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 设置新的默认配置
	if err := tx.Model(&config).Update("is_default", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// clearDefaultConfigs removes default flag from all configurations
func (s *LLMConfigService) clearDefaultConfigs() error {
	return s.db.Model(&models.LLMConfigModel{}).Where("is_active = ?", true).Update("is_default", false).Error
}
