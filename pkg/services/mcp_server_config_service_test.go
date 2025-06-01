package services

import (
	"path/filepath"
	"testing"

	"github.com/LubyRuffy/mcpagent/pkg/database"
	"github.com/LubyRuffy/mcpagent/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMCPTestDB(t *testing.T) {
	// 创建临时数据库文件
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_mcp.db")

	// 初始化数据库
	err := database.InitDatabase(dbPath)
	require.NoError(t, err)

	// 清理数据库中的测试数据
	database.GetDB().Exec("DELETE FROM mcp_server_configs")
}

func teardownMCPTestDB(t *testing.T) {
	if database.GetDB() != nil {
		database.CloseDatabase()
	}
}

func TestMCPServerConfigService_CreateConfig(t *testing.T) {
	setupMCPTestDB(t)
	defer teardownMCPTestDB(t)

	service := NewMCPServerConfigService()

	config := &models.MCPServerConfigModel{
		Name:        "test-server",
		Description: "Test server description",
		Command:     "uvx",
		IsActive:    true,
	}

	err := config.SetArgs([]string{"duckduckgo-mcp-server"})
	require.NoError(t, err)

	err = service.CreateConfig(config)
	assert.NoError(t, err)
	assert.NotZero(t, config.ID)
}

func TestMCPServerConfigService_CreateConfig_DuplicateName(t *testing.T) {
	setupMCPTestDB(t)
	defer teardownMCPTestDB(t)

	service := NewMCPServerConfigService()

	// 创建第一个配置
	config1 := &models.MCPServerConfigModel{
		Name:        "test-server",
		Description: "Test server description",
		Command:     "uvx",
		IsActive:    true,
	}

	err := service.CreateConfig(config1)
	assert.NoError(t, err)

	// 尝试创建同名配置
	config2 := &models.MCPServerConfigModel{
		Name:        "test-server", // 同名
		Description: "Another description",
		Command:     "npx",
		IsActive:    true,
	}

	err = service.CreateConfig(config2)
	assert.Error(t, err)
	assert.Equal(t, models.ErrMCPServerConfigNameExists, err)
}

func TestMCPServerConfigService_GetConfig(t *testing.T) {
	setupMCPTestDB(t)
	defer teardownMCPTestDB(t)

	service := NewMCPServerConfigService()

	// 创建配置
	config := &models.MCPServerConfigModel{
		Name:        "test-server",
		Description: "Test server description",
		Command:     "uvx",
		IsActive:    true,
	}

	err := service.CreateConfig(config)
	require.NoError(t, err)

	// 获取配置
	retrieved, err := service.GetConfig(config.ID)
	assert.NoError(t, err)
	assert.Equal(t, config.Name, retrieved.Name)
	assert.Equal(t, config.Command, retrieved.Command)
}

func TestMCPServerConfigService_GetConfigByName(t *testing.T) {
	setupMCPTestDB(t)
	defer teardownMCPTestDB(t)

	service := NewMCPServerConfigService()

	// 创建配置
	config := &models.MCPServerConfigModel{
		Name:        "test-server",
		Description: "Test server description",
		Command:     "uvx",
		IsActive:    true,
	}

	err := service.CreateConfig(config)
	require.NoError(t, err)

	// 按名称获取配置
	retrieved, err := service.GetConfigByName("test-server")
	assert.NoError(t, err)
	assert.Equal(t, config.Name, retrieved.Name)
	assert.Equal(t, config.Command, retrieved.Command)
}

func TestMCPServerConfigService_UpdateConfig(t *testing.T) {
	setupMCPTestDB(t)
	defer teardownMCPTestDB(t)

	service := NewMCPServerConfigService()

	// 创建配置
	config := &models.MCPServerConfigModel{
		Name:        "test-server",
		Description: "Test server description",
		Command:     "uvx",
		IsActive:    true,
	}

	err := service.CreateConfig(config)
	require.NoError(t, err)

	// 更新配置
	updates := &models.MCPServerConfigModel{
		Name:        "test-server",
		Description: "Updated description",
		Command:     "npx",
		IsActive:    true,
	}

	err = service.UpdateConfig(config.ID, updates)
	assert.NoError(t, err)

	// 验证更新
	retrieved, err := service.GetConfig(config.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated description", retrieved.Description)
	assert.Equal(t, "npx", retrieved.Command)
}

func TestMCPServerConfigService_DeleteConfig(t *testing.T) {
	setupMCPTestDB(t)
	defer teardownMCPTestDB(t)

	service := NewMCPServerConfigService()

	// 创建配置
	config := &models.MCPServerConfigModel{
		Name:        "test-server",
		Description: "Test server description",
		Command:     "uvx",
		IsActive:    true,
	}

	err := service.CreateConfig(config)
	require.NoError(t, err)

	// 删除配置
	err = service.DeleteConfig(config.ID)
	assert.NoError(t, err)

	// 验证删除
	_, err = service.GetConfig(config.ID)
	assert.Error(t, err)
	assert.Equal(t, models.ErrMCPServerConfigNotFound, err)
}

func TestMCPServerConfigService_ListConfigs(t *testing.T) {
	setupMCPTestDB(t)
	defer teardownMCPTestDB(t)

	service := NewMCPServerConfigService()

	// 创建多个配置
	configs := []*models.MCPServerConfigModel{
		{
			Name:        "server1",
			Description: "Server 1",
			Command:     "uvx",
			IsActive:    true,
		},
		{
			Name:        "server2",
			Description: "Server 2",
			Command:     "npx",
			IsActive:    true,
		},
	}

	for _, config := range configs {
		err := service.CreateConfig(config)
		require.NoError(t, err)
	}

	// 获取配置列表
	list, err := service.ListConfigs()
	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestMCPServerConfigService_GetAllActiveConfigs(t *testing.T) {
	setupMCPTestDB(t)
	defer teardownMCPTestDB(t)

	service := NewMCPServerConfigService()

	// 创建活跃和非活跃配置
	activeConfig := &models.MCPServerConfigModel{
		Name:        "active-server",
		Description: "Active server",
		Command:     "uvx",
		Disabled:    false,
		IsActive:    true,
	}

	disabledConfig := &models.MCPServerConfigModel{
		Name:        "disabled-server",
		Description: "Disabled server",
		Command:     "npx",
		Disabled:    true,
		IsActive:    true,
	}

	err := service.CreateConfig(activeConfig)
	require.NoError(t, err)

	err = service.CreateConfig(disabledConfig)
	require.NoError(t, err)

	// 获取活跃配置
	configMap, err := service.GetAllActiveConfigs()
	assert.NoError(t, err)
	assert.Len(t, configMap, 1)
	assert.Contains(t, configMap, "active-server")
	assert.NotContains(t, configMap, "disabled-server")
}
