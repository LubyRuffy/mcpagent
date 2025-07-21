// Package database provides migration scripts for database schema updates.
package database

import (
	"fmt"
	"log"
	"strings"

	"github.com/LubyRuffy/mcpagent/pkg/models"
)

// 执行所有迁移脚本
func RunMigrations() error {
	if DB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	// 检查是否需要迁移app_configs表添加mcp_settings字段
	if err := migrateAppConfigAddMCPSettings(); err != nil {
		return fmt.Errorf("迁移app_configs表添加mcp_settings字段失败: %w", err)
	}

	// 检查是否需要迁移mcp_server_configs表添加SSE支持字段
	if err := migrateMCPServerConfigAddSSESupport(); err != nil {
		return fmt.Errorf("迁移mcp_server_configs表添加SSE支持字段失败: %w", err)
	}

	return nil
}

// 迁移mcp_server_configs表添加SSE支持字段
func migrateMCPServerConfigAddSSESupport() error {
	// 检查mcp_server_configs表是否存在
	if !DB.Migrator().HasTable(&models.MCPServerConfigModel{}) {
		log.Println("mcp_server_configs表不存在，跳过迁移")
		return nil
	}

	// 检查是否需要添加transport_type字段
	if !DB.Migrator().HasColumn(&models.MCPServerConfigModel{}, "transport_type") {
		log.Println("mcp_server_configs表添加transport_type字段")

		if err := DB.Migrator().AddColumn(&models.MCPServerConfigModel{}, "transport_type"); err != nil {
			return fmt.Errorf("添加transport_type字段失败: %w", err)
		}

		// 设置默认值为stdio以保持向后兼容
		if err := DB.Exec("UPDATE mcp_server_configs SET transport_type = 'stdio' WHERE transport_type IS NULL OR TRIM(transport_type) = ''").Error; err != nil {
			return fmt.Errorf("设置transport_type默认值失败: %w", err)
		}

		log.Println("已添加transport_type字段并设置默认值")
	}

	// 检查是否需要添加url字段
	if !DB.Migrator().HasColumn(&models.MCPServerConfigModel{}, "url") {
		log.Println("mcp_server_configs表添加url字段")

		if err := DB.Migrator().AddColumn(&models.MCPServerConfigModel{}, "url"); err != nil {
			return fmt.Errorf("添加url字段失败: %w", err)
		}

		log.Println("已添加url字段")
	}

	// 检查是否需要添加headers字段
	if !DB.Migrator().HasColumn(&models.MCPServerConfigModel{}, "headers") {
		log.Println("mcp_server_configs表添加headers字段")

		if err := DB.Migrator().AddColumn(&models.MCPServerConfigModel{}, "headers"); err != nil {
			return fmt.Errorf("添加headers字段失败: %w", err)
		}

		log.Println("已添加headers字段")
	}

	// 修改command字段为非必需（移除NOT NULL约束）
	// 注意：SQLite不支持直接修改列约束，但GORM会在AutoMigrate时处理这个问题
	log.Println("SSE支持字段迁移完成")

	return nil
}

// 迁移app_configs表添加mcp_settings字段
func migrateAppConfigAddMCPSettings() error {
	// 检查app_configs表是否存在
	if !DB.Migrator().HasTable(&models.AppConfigModel{}) {
		log.Println("app_configs表不存在，跳过迁移")
		return nil
	}

	// 检查app_configs表是否有mcp_settings字段
	if !DB.Migrator().HasColumn(&models.AppConfigModel{}, "mcp_settings") {
		log.Println("app_configs表添加mcp_settings字段")

		// 添加字段
		if err := DB.Migrator().AddColumn(&models.AppConfigModel{}, "mcp_settings"); err != nil {
			return fmt.Errorf("添加mcp_settings字段失败: %w", err)
		}

		// 设置默认值
		if err := DB.Exec("UPDATE app_configs SET mcp_settings = '{}' WHERE mcp_settings IS NULL OR TRIM(mcp_settings) = ''").Error; err != nil {
			return fmt.Errorf("设置mcp_settings默认值失败: %w", err)
		}

		log.Println("已添加mcp_settings字段并设置默认值")
	}

	// 查找现有的默认配置
	var defaultConfigs []models.AppConfigModel
	if err := DB.Where("is_default = ? AND is_active = ?", true, true).Find(&defaultConfigs).Error; err != nil {
		return fmt.Errorf("查询默认配置失败: %w", err)
	}

	// 如果有默认配置，但未初始化MCP设置，则设置默认值
	for _, config := range defaultConfigs {
		if strings.TrimSpace(config.MCPSettings) == "" || config.MCPSettings == "{}" {
			// 设置默认MCP配置
			defaultMCPConfig := &models.MCPConfig{
				ConfigFile: "",
				Tools:      []models.MCPToolConfig{},
			}

			if err := config.SetMCPConfig(defaultMCPConfig); err != nil {
				log.Printf("警告：为配置 %d 设置默认MCP配置失败: %v", config.ID, err)
				continue
			}

			if err := DB.Save(&config).Error; err != nil {
				log.Printf("警告：保存配置 %d 的MCP配置失败: %v", config.ID, err)
				continue
			}

			log.Printf("已为配置 %d 设置默认MCP配置", config.ID)
		}
	}

	return nil
}
