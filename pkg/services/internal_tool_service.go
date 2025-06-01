// Package services provides business logic services for the MCP Agent application.
// It implements the service layer that handles business operations and data validation.
package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/LubyRuffy/mcpagent/pkg/config"
	"github.com/LubyRuffy/mcpagent/pkg/database"
	"github.com/LubyRuffy/mcpagent/pkg/models"
	"github.com/cloudwego/eino/components/tool"
	"gorm.io/gorm"
)

// InternalToolService provides business logic for internal tool management
type InternalToolService struct {
	db *gorm.DB
}

// NewInternalToolService creates a new internal tool service instance
func NewInternalToolService() *InternalToolService {
	return &InternalToolService{
		db: database.GetDB(),
	}
}

// SyncInternalTools synchronizes internal tools with the database
// It makes sure that all internal tools are present in the database,
// updates existing tools if needed, and removes tools that no longer exist.
// This function should be called during application startup.
func (s *InternalToolService) SyncInternalTools(ctx context.Context) error {
	// 获取内置工具
	internalTools, err := config.GetInternalTools(ctx)
	if err != nil {
		return fmt.Errorf("获取内置工具失败: %w", err)
	}

	log.Printf("获取到 %d 个内置工具", len(internalTools))
	for i, t := range internalTools {
		info, _ := t.Info(ctx)
		log.Printf("内置工具 #%d: %s (%T)", i+1, info.Name, t)
	}

	// 确保有一个内置服务器记录
	internalServer, err := s.ensureInternalServerExists()
	if err != nil {
		return fmt.Errorf("确保内置服务器记录存在失败: %w", err)
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取数据库中所有内置工具的记录
	var existingTools []models.MCPToolModel
	if err := tx.Where("server_id = ? AND is_active = ?", internalServer.ID, true).Find(&existingTools).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("查询内置工具失败: %w", err)
	}

	// 创建一个映射表，用于跟踪现有工具
	existingToolMap := make(map[string]*models.MCPToolModel)
	for i := range existingTools {
		existingToolMap[existingTools[i].Name] = &existingTools[i]
	}

	// 创建一个新工具的映射表，用于跟踪更新后的工具
	newToolMap := make(map[string]bool)

	// 处理每个内置工具
	now := time.Now()
	for i, einoTool := range internalTools {
		toolInfo, err := einoTool.Info(ctx)
		if err != nil {
			log.Printf("获取工具 %T 信息失败: %v", einoTool, err)
			continue
		}

		// 保持工具原始名称，使界面上显示与外部工具一致
		toolName := ""
		if i == 0 { // 假设第一个工具是 sequentialthinking
			toolName = config.SequentialThinkingToolName
		} else {
			// 对于其他工具，可以在这里添加对应的名称映射
			toolName = toolInfo.Name
		}

		// 标记这个工具已处理
		newToolMap[toolName] = true

		// 生成工具键 - 确保工具键在数据库中唯一
		toolKey := models.GenerateToolKey(config.InnerServerName, toolName)

		// 记录详细日志
		log.Printf("处理内置工具: 原始名称=%s, 界面显示名称=%s, 描述=%s, 工具键=%s",
			toolInfo.Name, toolName, toolInfo.Desc, toolKey)

		// 检查工具是否已存在
		existingTool, exists := existingToolMap[toolName]
		if exists {
			// 更新现有工具
			existingTool.Description = toolInfo.Desc
			existingTool.LastSyncAt = &now
			existingTool.IsActive = true

			// 尝试更新工具的输入模式
			if err := s.updateToolSchema(existingTool, einoTool); err != nil {
				log.Printf("更新工具 %s 的输入模式失败: %v", toolName, err)
			}

			// 保存更新
			if err := tx.Save(existingTool).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("更新工具 %s 失败: %w", toolName, err)
			}
		} else {
			// 创建新工具
			newTool := &models.MCPToolModel{
				Name:        toolName,
				Description: toolInfo.Desc,
				ServerID:    internalServer.ID,
				ToolKey:     toolKey,
				IsActive:    true,
				LastSyncAt:  &now,
			}

			// 设置工具的输入模式
			if err := s.updateToolSchema(newTool, einoTool); err != nil {
				log.Printf("设置工具 %s 的输入模式失败: %v", toolInfo.Name, err)
			}

			// 保存新工具
			if err := tx.Create(newTool).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("创建工具 %s 失败: %w", toolInfo.Name, err)
			}
		}
	}

	// 处理已删除的工具 - 如果数据库中有工具，但是内置工具列表中没有，则禁用它
	for toolName, existingTool := range existingToolMap {
		if !newToolMap[toolName] {
			existingTool.IsActive = false
			if err := tx.Save(existingTool).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("禁用已删除工具 %s 失败: %w", toolName, err)
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	log.Printf("成功同步 %d 个内置工具", len(internalTools))
	return nil
}

// ensureInternalServerExists makes sure that the internal server record exists in the database
func (s *InternalToolService) ensureInternalServerExists() (*models.MCPServerConfigModel, error) {
	var server models.MCPServerConfigModel
	err := s.db.Where("name = ?", config.InnerServerName).First(&server).Error
	if err == nil {
		// 服务器已存在，确保是可见且不可删除的
		updates := map[string]interface{}{
			"is_active": true,
			"disabled":  false,
		}
		if err := s.db.Model(&server).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("更新内置服务器失败: %w", err)
		}
		return &server, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("查询内置服务器失败: %w", err)
	}

	// 创建内置服务器
	server = models.MCPServerConfigModel{
		Name:        config.InnerServerName,
		Description: "内置工具服务器",
		Command:     "", // 命令为空，表示内置服务器
		IsActive:    true,
		Disabled:    false,
	}

	if err := s.db.Create(&server).Error; err != nil {
		return nil, fmt.Errorf("创建内置服务器失败: %w", err)
	}

	return &server, nil
}

// updateToolSchema tries to extract and set the input schema for a tool
func (s *InternalToolService) updateToolSchema(toolModel *models.MCPToolModel, einoTool tool.BaseTool) error {
	// 设置基本架构
	schemaMap := map[string]interface{}{
		"type": "object",
	}

	// 由于tool.BaseTool接口没有直接提供详细参数的方法，我们只设置基本结构
	// 如果需要提取参数细节，可能需要在eino框架的更新中支持，或者转换为特定类型

	return toolModel.SetInputSchema(schemaMap)
}

// CleanupInternalTools 删除所有内置工具记录，以便重新创建
func (s *InternalToolService) CleanupInternalTools() error {
	// 获取内置服务器ID
	var server models.MCPServerConfigModel
	err := s.db.Where("name = ?", config.InnerServerName).First(&server).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 内置服务器不存在，无需清理
			return nil
		}
		return fmt.Errorf("查询内置服务器失败: %w", err)
	}

	// 物理删除内置服务器的所有工具
	// 注意：这里使用 Unscoped().Delete 是物理删除而不是软删除
	result := s.db.Unscoped().Where("server_id = ?", server.ID).Delete(&models.MCPToolModel{})
	if result.Error != nil {
		return fmt.Errorf("删除内置工具失败: %w", result.Error)
	}

	log.Printf("已清理 %d 个内置工具记录", result.RowsAffected)
	return nil
}

// SyncInternalToolsWithDatabase 是一个导出函数，用于在应用程序启动时同步内置工具到数据库
// 该函数应该在应用程序主函数中被调用，以确保内置工具与数据库保持同步
//
// 参数:
//   - ctx: 上下文对象，用于控制同步过程
//
// 返回:
//   - error: 同步过程中的错误，如果成功则返回nil
func SyncInternalToolsWithDatabase(ctx context.Context) error {
	service := NewInternalToolService()

	// 删除并重建内置工具，解决命名冲突问题
	if err := service.CleanupInternalTools(); err != nil {
		log.Printf("清理内置工具失败: %v", err)
		// 继续执行，不中断流程
	}

	// 同步内置工具
	err := service.SyncInternalTools(ctx)
	if err != nil {
		log.Printf("同步内置工具失败: %v", err)
		return err
	}
	return nil
}
