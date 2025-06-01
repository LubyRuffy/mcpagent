package services

import (
	"os"
	"testing"

	"github.com/LubyRuffy/mcpagent/pkg/database"
	"github.com/LubyRuffy/mcpagent/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) {
	// 使用内存数据库进行测试
	err := database.InitDatabase(":memory:")
	require.NoError(t, err)
}

func teardownTestDB(t *testing.T) {
	err := database.CloseDatabase()
	if err != nil {
		t.Logf("关闭测试数据库失败: %v", err)
	}
}

func TestLLMConfigService_CreateConfig(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	service := NewLLMConfigService()

	config := &models.LLMConfigModel{
		Name:        "Test Config",
		Description: "Test description",
		Type:        "ollama",
		BaseURL:     "http://localhost:11434",
		Model:       "qwen3:14b",
		APIKey:      "ollama",
		IsDefault:   false,
		IsActive:    true,
	}

	err := service.CreateConfig(config)
	assert.NoError(t, err)
	assert.NotZero(t, config.ID)
}

func TestLLMConfigService_CreateConfig_DuplicateName(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	service := NewLLMConfigService()

	// 创建第一个配置
	config1 := &models.LLMConfigModel{
		Name:        "Test Config",
		Description: "Test description",
		Type:        "ollama",
		BaseURL:     "http://localhost:11434",
		Model:       "qwen3:14b",
		APIKey:      "ollama",
		IsDefault:   false,
		IsActive:    true,
	}

	err := service.CreateConfig(config1)
	assert.NoError(t, err)

	// 尝试创建同名配置
	config2 := &models.LLMConfigModel{
		Name:        "Test Config", // 同名
		Description: "Another description",
		Type:        "openai",
		BaseURL:     "https://api.openai.com/v1",
		Model:       "gpt-4",
		APIKey:      "sk-test",
		IsDefault:   false,
		IsActive:    true,
	}

	err = service.CreateConfig(config2)
	assert.Error(t, err)
	assert.Equal(t, models.ErrLLMConfigNameExists, err)
}

func TestLLMConfigService_GetConfig(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	service := NewLLMConfigService()

	// 创建配置
	config := &models.LLMConfigModel{
		Name:        "Test Config",
		Description: "Test description",
		Type:        "ollama",
		BaseURL:     "http://localhost:11434",
		Model:       "qwen3:14b",
		APIKey:      "ollama",
		IsDefault:   false,
		IsActive:    true,
	}

	err := service.CreateConfig(config)
	require.NoError(t, err)

	// 获取配置
	retrieved, err := service.GetConfig(config.ID)
	assert.NoError(t, err)
	assert.Equal(t, config.Name, retrieved.Name)
	assert.Equal(t, config.Type, retrieved.Type)
}

func TestLLMConfigService_GetConfig_NotFound(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	service := NewLLMConfigService()

	_, err := service.GetConfig(999)
	assert.Error(t, err)
	assert.Equal(t, models.ErrLLMConfigNotFound, err)
}

func TestLLMConfigService_ListConfigs(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	service := NewLLMConfigService()

	// 创建多个配置
	configs := []*models.LLMConfigModel{
		{
			Name:        "Config 1",
			Description: "Description 1",
			Type:        "ollama",
			BaseURL:     "http://localhost:11434",
			Model:       "qwen3:14b",
			APIKey:      "ollama",
			IsDefault:   true,
			IsActive:    true,
		},
		{
			Name:        "Config 2",
			Description: "Description 2",
			Type:        "openai",
			BaseURL:     "https://api.openai.com/v1",
			Model:       "gpt-4",
			APIKey:      "sk-test",
			IsDefault:   false,
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
	assert.Len(t, list, 3) // 包括默认创建的配置
}

func TestLLMConfigService_UpdateConfig(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	service := NewLLMConfigService()

	// 创建配置
	config := &models.LLMConfigModel{
		Name:        "Test Config",
		Description: "Test description",
		Type:        "ollama",
		BaseURL:     "http://localhost:11434",
		Model:       "qwen3:14b",
		APIKey:      "ollama",
		IsDefault:   false,
		IsActive:    true,
	}

	err := service.CreateConfig(config)
	require.NoError(t, err)

	// 更新配置
	updates := &models.LLMConfigModel{
		Name:        "Updated Config",
		Description: "Updated description",
		Type:        "openai",
		BaseURL:     "https://api.openai.com/v1",
		Model:       "gpt-4",
		APIKey:      "sk-updated",
		IsDefault:   false,
		IsActive:    true,
	}

	err = service.UpdateConfig(config.ID, updates)
	assert.NoError(t, err)

	// 验证更新
	retrieved, err := service.GetConfig(config.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Config", retrieved.Name)
	assert.Equal(t, "openai", retrieved.Type)
}

func TestLLMConfigService_DeleteConfig(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	service := NewLLMConfigService()

	// 创建配置
	config := &models.LLMConfigModel{
		Name:        "Test Config",
		Description: "Test description",
		Type:        "ollama",
		BaseURL:     "http://localhost:11434",
		Model:       "qwen3:14b",
		APIKey:      "ollama",
		IsDefault:   false,
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
	assert.Equal(t, models.ErrLLMConfigNotFound, err)
}

func TestLLMConfigService_SetDefaultConfig(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	service := NewLLMConfigService()

	// 创建配置
	config := &models.LLMConfigModel{
		Name:        "Test Config",
		Description: "Test description",
		Type:        "ollama",
		BaseURL:     "http://localhost:11434",
		Model:       "qwen3:14b",
		APIKey:      "ollama",
		IsDefault:   false,
		IsActive:    true,
	}

	err := service.CreateConfig(config)
	require.NoError(t, err)

	// 设置为默认配置
	err = service.SetDefaultConfig(config.ID)
	assert.NoError(t, err)

	// 验证设置
	retrieved, err := service.GetConfig(config.ID)
	assert.NoError(t, err)
	assert.True(t, retrieved.IsDefault)

	// 验证其他配置不再是默认配置
	defaultConfig, err := service.GetDefaultConfig()
	assert.NoError(t, err)
	assert.Equal(t, config.ID, defaultConfig.ID)
}

func TestMain(m *testing.M) {
	// 运行测试
	code := m.Run()
	os.Exit(code)
}
