package services

import (
	"testing"

	"github.com/LubyRuffy/mcpagent/pkg/database"
	"github.com/LubyRuffy/mcpagent/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupMCPToolTestDB(t *testing.T) {
	// 使用内存数据库进行测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 设置全局数据库实例
	database.DB = db

	// 执行迁移
	err = db.AutoMigrate(
		&models.LLMConfigModel{},
		&models.MCPServerConfigModel{},
		&models.MCPToolModel{},
	)
	require.NoError(t, err)
}

func teardownMCPToolTestDB(t *testing.T) {
	if database.DB != nil {
		sqlDB, err := database.DB.DB()
		require.NoError(t, err)
		sqlDB.Close()
	}
}

func createTestMCPServer(t *testing.T) *models.MCPServerConfigModel {
	server := &models.MCPServerConfigModel{
		Name:        "test-server",
		Description: "Test server description",
		Command:     "uvx",
		IsActive:    true,
	}

	err := server.SetArgs([]string{"test-mcp-server"})
	require.NoError(t, err)

	err = database.GetDB().Create(server).Error
	require.NoError(t, err)

	return server
}

func TestMCPToolService_CreateTool(t *testing.T) {
	setupMCPToolTestDB(t)
	defer teardownMCPToolTestDB(t)

	service := NewMCPToolService()
	server := createTestMCPServer(t)

	tool := &models.MCPToolModel{
		Name:        "test_tool",
		Description: "Test tool description",
		ServerID:    server.ID,
		ToolKey:     models.GenerateToolKey(server.Name, "test_tool"),
		IsActive:    true,
	}

	err := service.CreateTool(tool)
	assert.NoError(t, err)
	assert.NotZero(t, tool.ID)
	assert.NotNil(t, tool.LastSyncAt)
}

func TestMCPToolService_CreateTool_DuplicateKey(t *testing.T) {
	setupMCPToolTestDB(t)
	defer teardownMCPToolTestDB(t)

	service := NewMCPToolService()
	server := createTestMCPServer(t)

	tool1 := &models.MCPToolModel{
		Name:        "test_tool",
		Description: "Test tool description",
		ServerID:    server.ID,
		ToolKey:     models.GenerateToolKey(server.Name, "test_tool"),
		IsActive:    true,
	}

	err := service.CreateTool(tool1)
	require.NoError(t, err)

	// 尝试创建相同工具键的工具
	tool2 := &models.MCPToolModel{
		Name:        "test_tool",
		Description: "Another test tool description",
		ServerID:    server.ID,
		ToolKey:     models.GenerateToolKey(server.Name, "test_tool"),
		IsActive:    true,
	}

	err = service.CreateTool(tool2)
	assert.Equal(t, models.ErrMCPToolKeyExists, err)
}

func TestMCPToolService_GetAllActiveTools(t *testing.T) {
	setupMCPToolTestDB(t)
	defer teardownMCPToolTestDB(t)

	service := NewMCPToolService()
	server := createTestMCPServer(t)

	// 创建几个工具
	tools := []*models.MCPToolModel{
		{
			Name:        "tool1",
			Description: "Tool 1 description",
			ServerID:    server.ID,
			ToolKey:     models.GenerateToolKey(server.Name, "tool1"),
			IsActive:    true,
		},
		{
			Name:        "tool2",
			Description: "Tool 2 description",
			ServerID:    server.ID,
			ToolKey:     models.GenerateToolKey(server.Name, "tool2"),
			IsActive:    true,
		},
		{
			Name:        "tool3",
			Description: "Tool 3 description",
			ServerID:    server.ID,
			ToolKey:     models.GenerateToolKey(server.Name, "tool3"),
			IsActive:    false, // 非活跃工具
		},
	}

	for _, tool := range tools {
		err := service.CreateTool(tool)
		require.NoError(t, err)
	}

	// 获取所有活跃工具
	activeTools, err := service.GetAllActiveTools()
	assert.NoError(t, err)
	assert.Len(t, activeTools, 2) // 只有2个活跃工具

	// 验证预加载的服务器信息
	for _, tool := range activeTools {
		assert.Equal(t, server.Name, tool.Server.Name)
	}
}

func TestMCPToolService_GetToolsByServerID(t *testing.T) {
	setupMCPToolTestDB(t)
	defer teardownMCPToolTestDB(t)

	service := NewMCPToolService()
	server1 := createTestMCPServer(t)

	// 创建第二个服务器
	server2 := &models.MCPServerConfigModel{
		Name:        "test-server-2",
		Description: "Test server 2 description",
		Command:     "uvx",
		IsActive:    true,
	}
	err := server2.SetArgs([]string{"test-mcp-server-2"})
	require.NoError(t, err)
	err = database.GetDB().Create(server2).Error
	require.NoError(t, err)

	// 为每个服务器创建工具
	tool1 := &models.MCPToolModel{
		Name:        "tool1",
		Description: "Tool 1 description",
		ServerID:    server1.ID,
		ToolKey:     models.GenerateToolKey(server1.Name, "tool1"),
		IsActive:    true,
	}
	err = service.CreateTool(tool1)
	require.NoError(t, err)

	tool2 := &models.MCPToolModel{
		Name:        "tool2",
		Description: "Tool 2 description",
		ServerID:    server2.ID,
		ToolKey:     models.GenerateToolKey(server2.Name, "tool2"),
		IsActive:    true,
	}
	err = service.CreateTool(tool2)
	require.NoError(t, err)

	// 获取server1的工具
	server1Tools, err := service.GetToolsByServerID(server1.ID)
	assert.NoError(t, err)
	assert.Len(t, server1Tools, 1)
	assert.Equal(t, "tool1", server1Tools[0].Name)

	// 获取server2的工具
	server2Tools, err := service.GetToolsByServerID(server2.ID)
	assert.NoError(t, err)
	assert.Len(t, server2Tools, 1)
	assert.Equal(t, "tool2", server2Tools[0].Name)
}

func TestMCPToolService_GetToolByKey(t *testing.T) {
	setupMCPToolTestDB(t)
	defer teardownMCPToolTestDB(t)

	service := NewMCPToolService()
	server := createTestMCPServer(t)

	tool := &models.MCPToolModel{
		Name:        "test_tool",
		Description: "Test tool description",
		ServerID:    server.ID,
		ToolKey:     models.GenerateToolKey(server.Name, "test_tool"),
		IsActive:    true,
	}

	err := service.CreateTool(tool)
	require.NoError(t, err)

	// 通过工具键获取工具
	foundTool, err := service.GetToolByKey(tool.ToolKey)
	assert.NoError(t, err)
	assert.Equal(t, tool.Name, foundTool.Name)
	assert.Equal(t, tool.ToolKey, foundTool.ToolKey)
	assert.Equal(t, server.Name, foundTool.Server.Name)

	// 尝试获取不存在的工具
	_, err = service.GetToolByKey("nonexistent_key")
	assert.Equal(t, models.ErrMCPToolNotFound, err)
}

func TestMCPToolService_UpdateTool(t *testing.T) {
	setupMCPToolTestDB(t)
	defer teardownMCPToolTestDB(t)

	service := NewMCPToolService()
	server := createTestMCPServer(t)

	tool := &models.MCPToolModel{
		Name:        "test_tool",
		Description: "Test tool description",
		ServerID:    server.ID,
		ToolKey:     models.GenerateToolKey(server.Name, "test_tool"),
		IsActive:    true,
	}

	err := service.CreateTool(tool)
	require.NoError(t, err)

	// 更新工具
	updates := &models.MCPToolModel{
		Name:        "updated_tool",
		Description: "Updated tool description",
		ServerID:    server.ID,
		ToolKey:     models.GenerateToolKey(server.Name, "updated_tool"),
		IsActive:    true,
	}

	err = service.UpdateTool(tool.ID, updates)
	assert.NoError(t, err)

	// 验证更新
	updatedTool, err := service.GetToolByKey(updates.ToolKey)
	assert.NoError(t, err)
	assert.Equal(t, "updated_tool", updatedTool.Name)
	assert.Equal(t, "Updated tool description", updatedTool.Description)
}

func TestMCPToolService_DeleteTool(t *testing.T) {
	setupMCPToolTestDB(t)
	defer teardownMCPToolTestDB(t)

	service := NewMCPToolService()
	server := createTestMCPServer(t)

	tool := &models.MCPToolModel{
		Name:        "test_tool",
		Description: "Test tool description",
		ServerID:    server.ID,
		ToolKey:     models.GenerateToolKey(server.Name, "test_tool"),
		IsActive:    true,
	}

	err := service.CreateTool(tool)
	require.NoError(t, err)

	// 删除工具
	err = service.DeleteTool(tool.ID)
	assert.NoError(t, err)

	// 验证工具已被软删除
	_, err = service.GetToolByKey(tool.ToolKey)
	assert.Equal(t, models.ErrMCPToolNotFound, err)
}

func TestMCPToolService_DeleteToolsByServerID(t *testing.T) {
	setupMCPToolTestDB(t)
	defer teardownMCPToolTestDB(t)

	service := NewMCPToolService()
	server := createTestMCPServer(t)

	// 创建多个工具
	tools := []*models.MCPToolModel{
		{
			Name:        "tool1",
			Description: "Tool 1 description",
			ServerID:    server.ID,
			ToolKey:     models.GenerateToolKey(server.Name, "tool1"),
			IsActive:    true,
		},
		{
			Name:        "tool2",
			Description: "Tool 2 description",
			ServerID:    server.ID,
			ToolKey:     models.GenerateToolKey(server.Name, "tool2"),
			IsActive:    true,
		},
	}

	for _, tool := range tools {
		err := service.CreateTool(tool)
		require.NoError(t, err)
	}

	// 删除服务器的所有工具
	err := service.DeleteToolsByServerID(server.ID)
	assert.NoError(t, err)

	// 验证所有工具都被软删除
	serverTools, err := service.GetToolsByServerID(server.ID)
	assert.NoError(t, err)
	assert.Len(t, serverTools, 0)
}

func TestMCPToolModel_SetGetInputSchema(t *testing.T) {
	tool := &models.MCPToolModel{}

	// 测试设置和获取输入模式
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query",
			},
		},
		"required": []string{"query"},
	}

	err := tool.SetInputSchema(schema)
	assert.NoError(t, err)

	retrievedSchema, err := tool.GetInputSchema()
	assert.NoError(t, err)
	assert.Equal(t, "object", retrievedSchema["type"])

	// 测试空模式
	err = tool.SetInputSchema(nil)
	assert.NoError(t, err)
	assert.Equal(t, "", tool.InputSchema)

	retrievedSchema, err = tool.GetInputSchema()
	assert.NoError(t, err)
	assert.Nil(t, retrievedSchema)
}

func TestGenerateToolKey(t *testing.T) {
	serverName := "test-server"
	toolName := "test-tool"
	expected := "test-server_test-tool"

	result := models.GenerateToolKey(serverName, toolName)
	assert.Equal(t, expected, result)
}
