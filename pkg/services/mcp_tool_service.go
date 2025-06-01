// Package services provides business logic services for the MCP Agent application.
// It implements the service layer that handles business operations and data validation.
package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/LubyRuffy/mcpagent/pkg/database"
	"github.com/LubyRuffy/mcpagent/pkg/mcphost"
	"github.com/LubyRuffy/mcpagent/pkg/models"
	"gorm.io/gorm"
)

// MCPToolService provides business logic for MCP tool management
type MCPToolService struct {
	db *gorm.DB
}

// NewMCPToolService creates a new MCP tool service instance
func NewMCPToolService() *MCPToolService {
	return &MCPToolService{
		db: database.GetDB(),
	}
}

// GetAllActiveTools returns all active tools from the database
func (s *MCPToolService) GetAllActiveTools() ([]models.MCPToolModel, error) {
	var tools []models.MCPToolModel
	// 加载所有活跃的工具，并预加载Server关联
	err := s.db.Preload("Server").Where("is_active = ?", true).Find(&tools).Error

	// 确保工具记录的服务器关联正确
	for i, tool := range tools {
		if tool.Server.ID == 0 {
			// 如果服务器未加载，尝试单独加载
			var server models.MCPServerConfigModel
			if err := s.db.Where("id = ?", tool.ServerID).First(&server).Error; err == nil {
				tools[i].Server = server
			}
		}
	}

	return tools, err
}

// GetToolsByServerID returns all active tools for a specific server
func (s *MCPToolService) GetToolsByServerID(serverID uint) ([]models.MCPToolModel, error) {
	var tools []models.MCPToolModel
	err := s.db.Preload("Server").Where("server_id = ? AND is_active = ?", serverID, true).Find(&tools).Error
	return tools, err
}

// GetToolByKey returns a tool by its unique key
func (s *MCPToolService) GetToolByKey(toolKey string) (*models.MCPToolModel, error) {
	var tool models.MCPToolModel
	err := s.db.Preload("Server").Where("tool_key = ? AND is_active = ?", toolKey, true).First(&tool).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrMCPToolNotFound
		}
		return nil, err
	}
	return &tool, nil
}

// CreateTool creates a new tool
func (s *MCPToolService) CreateTool(tool *models.MCPToolModel) error {
	// 验证工具
	if err := tool.Validate(); err != nil {
		return err
	}

	// 检查工具键是否已存在
	var count int64
	err := s.db.Model(&models.MCPToolModel{}).Where("tool_key = ? AND is_active = ?", tool.ToolKey, true).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return models.ErrMCPToolKeyExists
	}

	// 设置同步时间
	now := time.Now()
	tool.LastSyncAt = &now

	// 创建工具
	return s.db.Create(tool).Error
}

// UpdateTool updates an existing tool
func (s *MCPToolService) UpdateTool(id uint, updates *models.MCPToolModel) error {
	// 验证更新数据
	if err := updates.Validate(); err != nil {
		return err
	}

	// 检查工具是否存在
	var existingTool models.MCPToolModel
	err := s.db.Where("id = ? AND is_active = ?", id, true).First(&existingTool).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.ErrMCPToolNotFound
		}
		return err
	}

	// 检查工具键是否与其他工具冲突
	if updates.ToolKey != existingTool.ToolKey {
		var count int64
		err := s.db.Model(&models.MCPToolModel{}).Where("tool_key = ? AND id != ? AND is_active = ?", updates.ToolKey, id, true).Count(&count).Error
		if err != nil {
			return err
		}
		if count > 0 {
			return models.ErrMCPToolKeyExists
		}
	}

	// 设置同步时间
	now := time.Now()
	updates.LastSyncAt = &now

	// 更新工具
	updates.ID = id
	return s.db.Model(&existingTool).Updates(updates).Error
}

// DeleteTool soft deletes a tool
func (s *MCPToolService) DeleteTool(id uint) error {
	// 检查工具是否存在
	var tool models.MCPToolModel
	err := s.db.Where("id = ? AND is_active = ?", id, true).First(&tool).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.ErrMCPToolNotFound
		}
		return err
	}

	// 软删除工具
	return s.db.Model(&tool).Update("is_active", false).Error
}

// DeleteToolsByServerID soft deletes all tools for a specific server
func (s *MCPToolService) DeleteToolsByServerID(serverID uint) error {
	return s.db.Model(&models.MCPToolModel{}).Where("server_id = ? AND is_active = ?", serverID, true).Update("is_active", false).Error
}

// SyncToolsForServer synchronizes tools for a specific server by connecting to it
func (s *MCPToolService) SyncToolsForServer(ctx context.Context, serverConfig *models.MCPServerConfigModel) error {
	// 将数据库配置转换为mcphost.ServerConfig格式
	mcpServerConfig, err := serverConfig.ToServerConfig()
	if err != nil {
		return fmt.Errorf("转换服务器配置失败: %w", err)
	}

	// 创建MCPSettings
	settings := &mcphost.MCPSettings{
		MCPServers: map[string]mcphost.ServerConfig{
			serverConfig.Name: mcpServerConfig,
		},
	}

	// 使用连接池获取MCP服务器连接
	pool := mcphost.GetConnectionPool()

	// 获取或创建连接
	hub, err := pool.GetHub(ctx, settings)
	if err != nil {
		return fmt.Errorf("连接MCP服务器失败: %w", err)
	}
	// 注意：不再直接调用hub.CloseServers()，而是在使用完后释放引用
	defer pool.ReleaseHub(settings)

	// 获取工具列表
	toolsMap, err := hub.GetToolsMap(ctx)
	if err != nil {
		return fmt.Errorf("获取工具列表失败: %w", err)
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 先删除该服务器的所有现有工具
	if err := tx.Model(&models.MCPToolModel{}).Where("server_id = ?", serverConfig.ID).Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除现有工具失败: %w", err)
	}

	// 添加新工具
	now := time.Now()
	for toolKey, toolInfo := range toolsMap {
		// 解析工具键获取工具名称
		toolName := toolInfo.Name
		if toolName == "" {
			// 如果工具名称为空，从工具键中提取
			if len(toolKey) > len(serverConfig.Name)+1 {
				toolName = toolKey[len(serverConfig.Name)+1:]
			} else {
				toolName = toolKey
			}
		}

		tool := &models.MCPToolModel{
			Name:        toolName,
			Description: toolInfo.Desc,
			ServerID:    serverConfig.ID,
			ToolKey:     toolKey,
			IsActive:    true,
			LastSyncAt:  &now,
		}

		// 设置输入模式（如果有的话）
		if toolInfo.ParamsOneOf != nil {
			// 尝试将参数模式转换为OpenAPI v3格式
			if openAPISchema, err := toolInfo.ParamsOneOf.ToOpenAPIV3(); err == nil && openAPISchema != nil {
				// 将OpenAPI模式转换为简单的map格式存储
				schemaMap := make(map[string]interface{})
				schemaMap["type"] = openAPISchema.Type
				if openAPISchema.Properties != nil {
					schemaMap["properties"] = openAPISchema.Properties
				}
				if openAPISchema.Required != nil {
					schemaMap["required"] = openAPISchema.Required
				}
				if err := tool.SetInputSchema(schemaMap); err != nil {
					log.Printf("设置工具 %s 的输入模式失败: %v", toolKey, err)
				}
			} else {
				// 如果转换失败，设置一个基本的对象模式
				schemaMap := map[string]interface{}{
					"type": "object",
				}
				if err := tool.SetInputSchema(schemaMap); err != nil {
					log.Printf("设置工具 %s 的基本输入模式失败: %v", toolKey, err)
				}
			}
		}

		if err := tx.Create(tool).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("创建工具 %s 失败: %w", toolKey, err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	log.Printf("成功同步服务器 %s 的 %d 个工具", serverConfig.Name, len(toolsMap))
	return nil
}

// GetToolsInfo returns tool information for API responses
func (s *MCPToolService) GetToolsInfo() ([]models.MCPToolInfo, error) {
	tools, err := s.GetAllActiveTools()
	if err != nil {
		return nil, err
	}

	var toolsInfo []models.MCPToolInfo
	for _, tool := range tools {
		toolsInfo = append(toolsInfo, tool.ToMCPToolInfo())
	}

	return toolsInfo, nil
}
