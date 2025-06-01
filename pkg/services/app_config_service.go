// Package services provides business logic services for the MCP Agent application.
// It handles interactions between controllers and database models.
package services

import (
	"fmt"

	"github.com/LubyRuffy/mcpagent/pkg/config"
	"github.com/LubyRuffy/mcpagent/pkg/database"
	"github.com/LubyRuffy/mcpagent/pkg/models"
	"gorm.io/gorm"
)

// AppConfigService provides services for managing application configurations
type AppConfigService struct {
	db *gorm.DB
}

// NewAppConfigService creates a new AppConfigService instance
func NewAppConfigService() *AppConfigService {
	return &AppConfigService{db: database.GetDB()}
}

// ListConfigs returns all active application configurations
func (s *AppConfigService) ListConfigs() ([]models.AppConfigModel, error) {
	var configs []models.AppConfigModel
	err := s.db.Where("is_active = ?", true).Find(&configs).Error
	return configs, err
}

// GetConfig returns a specific application configuration by ID
func (s *AppConfigService) GetConfig(id uint) (*models.AppConfigModel, error) {
	var config models.AppConfigModel
	err := s.db.Where("id = ? AND is_active = ?", id, true).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrAppConfigNotFound
		}
		return nil, err
	}
	return &config, nil
}

// GetDefaultConfig returns the default application configuration
func (s *AppConfigService) GetDefaultConfig() (*models.AppConfigModel, error) {
	var config models.AppConfigModel
	err := s.db.Where("is_default = ? AND is_active = ?", true, true).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrAppConfigNotFound
		}
		return nil, err
	}
	return &config, nil
}

// CreateConfig creates a new application configuration
func (s *AppConfigService) CreateConfig(config *models.AppConfigModel) error {
	// 验证配置
	if err := config.Validate(); err != nil {
		return err
	}

	// 检查名称是否已存在
	var count int64
	err := s.db.Model(&models.AppConfigModel{}).Where("name = ? AND is_active = ?", config.Name, true).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return models.ErrAppConfigNameExists
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

// UpdateConfig updates an existing application configuration
func (s *AppConfigService) UpdateConfig(id uint, updates *models.AppConfigModel) error {
	// 验证更新数据
	if err := updates.Validate(); err != nil {
		return err
	}

	// 检查配置是否存在
	var existingConfig models.AppConfigModel
	err := s.db.Where("id = ? AND is_active = ?", id, true).First(&existingConfig).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.ErrAppConfigNotFound
		}
		return err
	}

	// 检查名称是否与其他配置冲突
	if updates.Name != existingConfig.Name {
		var count int64
		err := s.db.Model(&models.AppConfigModel{}).Where("name = ? AND id != ? AND is_active = ?", updates.Name, id, true).Count(&count).Error
		if err != nil {
			return err
		}
		if count > 0 {
			return models.ErrAppConfigNameExists
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

// DeleteConfig soft deletes an application configuration
func (s *AppConfigService) DeleteConfig(id uint) error {
	// 检查配置是否存在
	var config models.AppConfigModel
	err := s.db.Where("id = ? AND is_active = ?", id, true).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.ErrAppConfigNotFound
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
func (s *AppConfigService) SetDefaultConfig(id uint) error {
	// 检查配置是否存在
	var config models.AppConfigModel
	err := s.db.Where("id = ? AND is_active = ?", id, true).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.ErrAppConfigNotFound
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
	if err := tx.Model(&models.AppConfigModel{}).Where("is_active = ?", true).Update("is_default", false).Error; err != nil {
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

// SaveToConfig converts the database model to a config.Config
func (s *AppConfigService) SaveToConfig(appConfig *models.AppConfigModel, targetConfig *config.Config) error {
	if appConfig == nil {
		return fmt.Errorf("配置为空")
	}

	// 设置基本配置
	targetConfig.Proxy = appConfig.Proxy
	targetConfig.SystemPrompt = appConfig.SystemPrompt
	targetConfig.MaxStep = appConfig.MaxStep

	// 获取并设置占位符
	placeholders, err := appConfig.GetPlaceHoldersAsMap()
	if err != nil {
		return err
	}
	targetConfig.PlaceHolders = placeholders

	// 获取MCP配置
	mcpConfig, err := appConfig.GetMCPConfig()
	if err == nil && mcpConfig != nil {
		// 设置MCP配置
		targetConfig.MCP.ConfigFile = mcpConfig.ConfigFile
		targetConfig.MCP.Tools = mcpConfig.Tools

		// 注意: MCP服务器信息需要从MCPServerConfigService获取，这里不会覆盖
		// 但会保留tools的选择
	}

	return nil
}

// LoadFromConfig converts a config.Config to a database model
func (s *AppConfigService) LoadFromConfig(sourceConfig *config.Config, appConfig *models.AppConfigModel) error {
	if sourceConfig == nil {
		return fmt.Errorf("配置为空")
	}

	// 设置基本配置
	appConfig.Proxy = sourceConfig.Proxy
	appConfig.SystemPrompt = sourceConfig.SystemPrompt
	appConfig.MaxStep = sourceConfig.MaxStep

	// 设置占位符
	if err := appConfig.SetPlaceHoldersFromMap(sourceConfig.PlaceHolders); err != nil {
		return err
	}

	// 设置MCP配置
	mcpConfig := &models.MCPConfig{
		ConfigFile: sourceConfig.MCP.ConfigFile,
		Tools:      sourceConfig.MCP.Tools,
	}
	if err := appConfig.SetMCPConfig(mcpConfig); err != nil {
		return err
	}

	return nil
}

// clearDefaultConfigs removes default flag from all configurations
func (s *AppConfigService) clearDefaultConfigs() error {
	return s.db.Model(&models.AppConfigModel{}).Where("is_active = ?", true).Update("is_default", false).Error
}
