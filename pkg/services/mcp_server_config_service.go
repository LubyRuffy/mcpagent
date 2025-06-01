// Package services provides business logic services for the MCP Agent application.
// It implements the service layer that handles business operations and data validation.
package services

import (
	"github.com/LubyRuffy/mcpagent/pkg/database"
	"github.com/LubyRuffy/mcpagent/pkg/models"
	"gorm.io/gorm"
)

// MCPServerConfigService provides business logic for MCP server configuration management
type MCPServerConfigService struct {
	db *gorm.DB
}

// NewMCPServerConfigService creates a new MCP server configuration service instance
func NewMCPServerConfigService() *MCPServerConfigService {
	return &MCPServerConfigService{
		db: database.GetDB(),
	}
}

// ListConfigs returns all active MCP server configurations
func (s *MCPServerConfigService) ListConfigs() ([]models.MCPServerConfigModel, error) {
	var configs []models.MCPServerConfigModel
	err := s.db.Where("is_active = ?", true).Order("created_at ASC").Find(&configs).Error
	return configs, err
}

// GetConfig returns a specific MCP server configuration by ID
func (s *MCPServerConfigService) GetConfig(id uint) (*models.MCPServerConfigModel, error) {
	var config models.MCPServerConfigModel
	err := s.db.Where("id = ? AND is_active = ?", id, true).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrMCPServerConfigNotFound
		}
		return nil, err
	}
	return &config, nil
}

// GetConfigByName returns a specific MCP server configuration by name
func (s *MCPServerConfigService) GetConfigByName(name string) (*models.MCPServerConfigModel, error) {
	var config models.MCPServerConfigModel
	err := s.db.Where("name = ? AND is_active = ?", name, true).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrMCPServerConfigNotFound
		}
		return nil, err
	}
	return &config, nil
}

// CreateConfig creates a new MCP server configuration
func (s *MCPServerConfigService) CreateConfig(config *models.MCPServerConfigModel) error {
	// 验证配置
	if err := config.Validate(); err != nil {
		return err
	}

	// 检查名称是否已存在
	var count int64
	err := s.db.Model(&models.MCPServerConfigModel{}).Where("name = ? AND is_active = ?", config.Name, true).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return models.ErrMCPServerConfigNameExists
	}

	// 创建配置
	return s.db.Create(config).Error
}

// UpdateConfig updates an existing MCP server configuration
func (s *MCPServerConfigService) UpdateConfig(id uint, updates *models.MCPServerConfigModel) error {
	// 验证更新数据
	if err := updates.Validate(); err != nil {
		return err
	}

	// 检查配置是否存在
	var existingConfig models.MCPServerConfigModel
	err := s.db.Where("id = ? AND is_active = ?", id, true).First(&existingConfig).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.ErrMCPServerConfigNotFound
		}
		return err
	}

	// 检查名称是否与其他配置冲突
	if updates.Name != existingConfig.Name {
		var count int64
		err := s.db.Model(&models.MCPServerConfigModel{}).Where("name = ? AND id != ? AND is_active = ?", updates.Name, id, true).Count(&count).Error
		if err != nil {
			return err
		}
		if count > 0 {
			return models.ErrMCPServerConfigNameExists
		}
	}

	// 更新配置
	updates.ID = id
	return s.db.Model(&existingConfig).Updates(updates).Error
}

// DeleteConfig soft deletes an MCP server configuration
func (s *MCPServerConfigService) DeleteConfig(id uint) error {
	// 检查配置是否存在
	var config models.MCPServerConfigModel
	err := s.db.Where("id = ? AND is_active = ?", id, true).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.ErrMCPServerConfigNotFound
		}
		return err
	}

	// 软删除配置
	return s.db.Model(&config).Update("is_active", false).Error
}

// GetAllActiveConfigs returns all active MCP server configurations as a map
func (s *MCPServerConfigService) GetAllActiveConfigs() (map[string]models.MCPServerConfigModel, error) {
	configs, err := s.ListConfigs()
	if err != nil {
		return nil, err
	}

	configMap := make(map[string]models.MCPServerConfigModel)
	for _, config := range configs {
		if !config.Disabled {
			configMap[config.Name] = config
		}
	}

	return configMap, nil
}
