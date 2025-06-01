package services

import (
	"testing"

	"github.com/LubyRuffy/mcpagent/pkg/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSystemPromptTestDB(t *testing.T) *gorm.DB {
	// 使用内存数据库进行测试
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// 自动迁移表结构
	err = db.AutoMigrate(&models.SystemPromptModel{})
	assert.NoError(t, err)

	return db
}

func TestSystemPromptService_CreatePrompt(t *testing.T) {
	// 设置测试数据库
	db := setupSystemPromptTestDB(t)

	// 创建服务实例，使用测试数据库
	service := &SystemPromptService{
		db: db,
	}

	// 测试数据
	prompt := &models.SystemPromptModel{
		Name:        "测试提示词",
		Description: "这是一个测试提示词",
		Content:     "你是一个测试助手，请帮助我测试{system}。",
		IsActive:    true,
	}

	// 设置占位符
	err := prompt.SetPlaceholdersFromStringSlice([]string{"system"})
	assert.NoError(t, err)

	// 执行测试
	err = service.CreatePrompt(prompt)
	assert.NoError(t, err)
	assert.NotZero(t, prompt.ID, "ID应该被自动赋值")

	// 验证结果
	var savedPrompt models.SystemPromptModel
	err = db.First(&savedPrompt, prompt.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "测试提示词", savedPrompt.Name)
	assert.Equal(t, "这是一个测试提示词", savedPrompt.Description)
	assert.Equal(t, "你是一个测试助手，请帮助我测试{system}。", savedPrompt.Content)
	assert.Equal(t, true, savedPrompt.IsActive)

	// 测试占位符
	placeholders, err := savedPrompt.GetPlaceholdersAsStringSlice()
	assert.NoError(t, err)
	assert.Equal(t, []string{"system"}, placeholders)
}

func TestSystemPromptService_GetPrompt(t *testing.T) {
	// 设置测试数据库
	db := setupSystemPromptTestDB(t)

	// 创建服务实例
	service := &SystemPromptService{
		db: db,
	}

	// 插入测试数据
	prompt := &models.SystemPromptModel{
		Name:        "唯一测试提示词名称", // 使用唯一的名称避免冲突
		Description: "这是一个测试提示词",
		Content:     "你是一个测试助手。",
		IsActive:    true,
	}
	err := prompt.SetPlaceholdersFromStringSlice([]string{})
	assert.NoError(t, err)

	err = db.Create(prompt).Error
	assert.NoError(t, err)
	assert.NotZero(t, prompt.ID, "提示词ID应该被自动分配")

	// 执行测试
	result, err := service.GetPrompt(prompt.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, prompt.ID, result.ID)
	assert.Equal(t, prompt.Name, result.Name)

	// 测试获取不存在的提示词
	nonExistentID := prompt.ID + 100
	result, err = service.GetPrompt(nonExistentID)
	assert.Error(t, err)
	assert.Equal(t, models.ErrSystemPromptNotFound, err)
	assert.Nil(t, result)
}

func TestSystemPromptService_UpdatePrompt(t *testing.T) {
	// 设置测试数据库
	db := setupSystemPromptTestDB(t)

	// 创建服务实例
	service := &SystemPromptService{
		db: db,
	}

	// 插入测试数据
	prompt := &models.SystemPromptModel{
		Name:        "唯一测试提示词-更新测试",
		Description: "这是一个测试提示词",
		Content:     "你是一个测试助手。",
		IsActive:    true,
	}
	err := prompt.SetPlaceholdersFromStringSlice([]string{})
	assert.NoError(t, err)

	err = db.Create(prompt).Error
	assert.NoError(t, err)
	assert.NotZero(t, prompt.ID, "提示词ID应该被自动分配")

	// 创建更新数据
	updates := &models.SystemPromptModel{
		Name:        "更新后的提示词",
		Description: "这是更新后的提示词",
		Content:     "你是一个更新后的测试助手，今天是{date}。",
	}
	err = updates.SetPlaceholdersFromStringSlice([]string{"date"})
	assert.NoError(t, err)

	// 执行测试
	err = service.UpdatePrompt(prompt.ID, updates)
	assert.NoError(t, err)

	// 验证结果
	var updatedPrompt models.SystemPromptModel
	err = db.First(&updatedPrompt, prompt.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "更新后的提示词", updatedPrompt.Name)
	assert.Equal(t, "这是更新后的提示词", updatedPrompt.Description)
	assert.Equal(t, "你是一个更新后的测试助手，今天是{date}。", updatedPrompt.Content)

	// 测试占位符
	placeholders, err := updatedPrompt.GetPlaceholdersAsStringSlice()
	assert.NoError(t, err)
	assert.Equal(t, []string{"date"}, placeholders)
}

func TestSystemPromptService_DeletePrompt(t *testing.T) {
	// 设置测试数据库
	db := setupSystemPromptTestDB(t)

	// 创建服务实例
	service := &SystemPromptService{
		db: db,
	}

	// 插入测试数据
	prompt := &models.SystemPromptModel{
		Name:        "唯一测试提示词-删除测试",
		Description: "这是一个测试提示词",
		Content:     "你是一个测试助手。",
		IsActive:    true,
	}
	err := prompt.SetPlaceholdersFromStringSlice([]string{})
	assert.NoError(t, err)

	err = db.Create(prompt).Error
	assert.NoError(t, err)
	assert.NotZero(t, prompt.ID, "提示词ID应该被自动分配")

	// 执行测试
	err = service.DeletePrompt(prompt.ID)
	assert.NoError(t, err)

	// 验证结果
	var count int64
	db.Model(&models.SystemPromptModel{}).Where("id = ? AND is_active = ?", prompt.ID, true).Count(&count)
	assert.Equal(t, int64(0), count, "记录应该被标记为非活动状态")
}

func TestSystemPromptService_SetDefaultPrompt(t *testing.T) {
	// 设置测试数据库
	db := setupSystemPromptTestDB(t)

	// 创建服务实例
	service := &SystemPromptService{
		db: db,
	}

	// 插入多个测试数据
	prompt1 := &models.SystemPromptModel{
		Name:        "唯一提示词1-默认测试",
		Description: "这是提示词1",
		Content:     "你是助手1。",
		IsDefault:   true,
		IsActive:    true,
	}
	err := prompt1.SetPlaceholdersFromStringSlice([]string{})
	assert.NoError(t, err)

	prompt2 := &models.SystemPromptModel{
		Name:        "唯一提示词2-默认测试",
		Description: "这是提示词2",
		Content:     "你是助手2。",
		IsDefault:   false,
		IsActive:    true,
	}
	err = prompt2.SetPlaceholdersFromStringSlice([]string{})
	assert.NoError(t, err)

	err = db.Create(prompt1).Error
	assert.NoError(t, err)
	assert.NotZero(t, prompt1.ID, "提示词ID应该被自动分配")

	err = db.Create(prompt2).Error
	assert.NoError(t, err)
	assert.NotZero(t, prompt2.ID, "提示词ID应该被自动分配")

	// 执行测试，设置prompt2为默认
	err = service.SetDefaultPrompt(prompt2.ID)
	assert.NoError(t, err)

	// 验证结果
	var updatedPrompt1 models.SystemPromptModel
	err = db.First(&updatedPrompt1, prompt1.ID).Error
	assert.NoError(t, err)
	assert.False(t, updatedPrompt1.IsDefault, "prompt1应该不再是默认配置")

	var updatedPrompt2 models.SystemPromptModel
	err = db.First(&updatedPrompt2, prompt2.ID).Error
	assert.NoError(t, err)
	assert.True(t, updatedPrompt2.IsDefault, "prompt2应该成为默认配置")
}
