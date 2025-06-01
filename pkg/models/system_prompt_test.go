package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSystemPromptModel_Validate(t *testing.T) {
	tests := []struct {
		name        string
		prompt      SystemPromptModel
		expectedErr error
	}{
		{
			name: "有效的系统提示词",
			prompt: SystemPromptModel{
				Name:    "测试提示词",
				Content: "这是一个测试提示词内容",
			},
			expectedErr: nil,
		},
		{
			name: "名称为空",
			prompt: SystemPromptModel{
				Content: "这是一个测试提示词内容",
			},
			expectedErr: ErrSystemPromptNameEmpty,
		},
		{
			name: "内容为空",
			prompt: SystemPromptModel{
				Name: "测试提示词",
			},
			expectedErr: ErrSystemPromptContentEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.prompt.Validate()
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestSystemPromptModel_PlaceholdersHandling(t *testing.T) {
	tests := []struct {
		name         string
		placeholders []string
	}{
		{
			name:         "空占位符",
			placeholders: []string{},
		},
		{
			name:         "单个占位符",
			placeholders: []string{"name"},
		},
		{
			name:         "多个占位符",
			placeholders: []string{"name", "date", "field"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := SystemPromptModel{}

			// 设置占位符
			err := prompt.SetPlaceholdersFromStringSlice(tt.placeholders)
			assert.NoError(t, err)

			// 获取占位符
			result, err := prompt.GetPlaceholdersAsStringSlice()
			assert.NoError(t, err)
			assert.Equal(t, tt.placeholders, result)
		})
	}
}

func TestSystemPromptModel_ToSystemPromptConfig(t *testing.T) {
	// 创建测试模型
	prompt := SystemPromptModel{
		ID:          1,
		Name:        "测试提示词",
		Description: "这是一个测试提示词",
		Content:     "你是一个测试助手，用户名是{name}，今天是{date}。",
		IsDefault:   true,
		IsActive:    true,
	}

	// 设置占位符
	err := prompt.SetPlaceholdersFromStringSlice([]string{"name", "date"})
	assert.NoError(t, err)

	// 测试转换
	config := prompt.ToSystemPromptConfig()

	// 验证结果
	assert.Equal(t, uint(1), config["id"])
	assert.Equal(t, "测试提示词", config["name"])
	assert.Equal(t, "这是一个测试提示词", config["description"])
	assert.Equal(t, "你是一个测试助手，用户名是{name}，今天是{date}。", config["content"])
	assert.Equal(t, true, config["is_default"])
	assert.Equal(t, true, config["is_active"])

	// 验证占位符
	placeholders, ok := config["placeholders"].([]string)
	assert.True(t, ok, "placeholders应该是[]string类型")
	assert.Equal(t, []string{"name", "date"}, placeholders)
}

func TestSystemPromptModel_FromSystemPromptConfig(t *testing.T) {
	// 创建配置数据
	config := map[string]interface{}{
		"name":         "测试提示词",
		"description":  "这是一个测试提示词",
		"content":      "你是一个测试助手，用户名是{name}，今天是{date}。",
		"placeholders": []string{"name", "date"},
		"is_default":   true,
	}

	// 创建模型并测试
	prompt := SystemPromptModel{}
	err := prompt.FromSystemPromptConfig(config)
	assert.NoError(t, err)

	// 验证结果
	assert.Equal(t, "测试提示词", prompt.Name)
	assert.Equal(t, "这是一个测试提示词", prompt.Description)
	assert.Equal(t, "你是一个测试助手，用户名是{name}，今天是{date}。", prompt.Content)
	assert.Equal(t, true, prompt.IsDefault)

	// 验证占位符
	placeholders, err := prompt.GetPlaceholdersAsStringSlice()
	assert.NoError(t, err)
	assert.Equal(t, []string{"name", "date"}, placeholders)
}
