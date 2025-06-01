package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLLMConfigModel_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  LLMConfigModel
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid ollama config",
			config: LLMConfigModel{
				Name:    "Test Ollama",
				Type:    "ollama",
				BaseURL: "http://localhost:11434",
				Model:   "qwen3:14b",
				APIKey:  "ollama",
			},
			wantErr: false,
		},
		{
			name: "valid openai config",
			config: LLMConfigModel{
				Name:    "Test OpenAI",
				Type:    "openai",
				BaseURL: "https://api.openai.com/v1",
				Model:   "gpt-4",
				APIKey:  "sk-test",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			config: LLMConfigModel{
				Type:    "ollama",
				BaseURL: "http://localhost:11434",
				Model:   "qwen3:14b",
				APIKey:  "ollama",
			},
			wantErr: true,
			errMsg:  "LLM配置名称不能为空",
		},
		{
			name: "empty type",
			config: LLMConfigModel{
				Name:    "Test Config",
				BaseURL: "http://localhost:11434",
				Model:   "qwen3:14b",
				APIKey:  "ollama",
			},
			wantErr: true,
			errMsg:  "LLM类型不能为空",
		},
		{
			name: "invalid type",
			config: LLMConfigModel{
				Name:    "Test Config",
				Type:    "invalid",
				BaseURL: "http://localhost:11434",
				Model:   "qwen3:14b",
				APIKey:  "ollama",
			},
			wantErr: true,
			errMsg:  "LLM类型无效，仅支持 openai 和 ollama",
		},
		{
			name: "empty base url",
			config: LLMConfigModel{
				Name:   "Test Config",
				Type:   "ollama",
				Model:  "qwen3:14b",
				APIKey: "ollama",
			},
			wantErr: true,
			errMsg:  "LLM Base URL不能为空",
		},
		{
			name: "empty model",
			config: LLMConfigModel{
				Name:    "Test Config",
				Type:    "ollama",
				BaseURL: "http://localhost:11434",
				APIKey:  "ollama",
			},
			wantErr: true,
			errMsg:  "LLM模型名称不能为空",
		},
		{
			name: "empty api key",
			config: LLMConfigModel{
				Name:    "Test Config",
				Type:    "ollama",
				BaseURL: "http://localhost:11434",
				Model:   "qwen3:14b",
			},
			wantErr: true,
			errMsg:  "LLM API Key不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLLMConfigModel_ToConfigLLM(t *testing.T) {
	temperature := 0.7
	maxTokens := 4000

	config := LLMConfigModel{
		Name:        "Test Config",
		Type:        "openai",
		BaseURL:     "https://api.openai.com/v1",
		Model:       "gpt-4",
		APIKey:      "sk-test",
		Temperature: &temperature,
		MaxTokens:   &maxTokens,
	}

	result := config.ToConfigLLM()

	expected := map[string]interface{}{
		"type":        "openai",
		"base_url":    "https://api.openai.com/v1",
		"model":       "gpt-4",
		"api_key":     "sk-test",
		"temperature": 0.7,
		"max_tokens":  4000,
	}

	assert.Equal(t, expected, result)
}

func TestLLMConfigModel_ToConfigLLM_WithoutOptionalFields(t *testing.T) {
	config := LLMConfigModel{
		Name:    "Test Config",
		Type:    "ollama",
		BaseURL: "http://localhost:11434",
		Model:   "qwen3:14b",
		APIKey:  "ollama",
	}

	result := config.ToConfigLLM()

	expected := map[string]interface{}{
		"type":     "ollama",
		"base_url": "http://localhost:11434",
		"model":    "qwen3:14b",
		"api_key":  "ollama",
	}

	assert.Equal(t, expected, result)
}

func TestLLMConfigModel_FromConfigLLM(t *testing.T) {
	configData := map[string]interface{}{
		"type":        "openai",
		"base_url":    "https://api.openai.com/v1",
		"model":       "gpt-4",
		"api_key":     "sk-test",
		"temperature": 0.8,
		"max_tokens":  8000,
	}

	var config LLMConfigModel
	config.FromConfigLLM("Test OpenAI", "Test description", configData)

	assert.Equal(t, "Test OpenAI", config.Name)
	assert.Equal(t, "Test description", config.Description)
	assert.Equal(t, "openai", config.Type)
	assert.Equal(t, "https://api.openai.com/v1", config.BaseURL)
	assert.Equal(t, "gpt-4", config.Model)
	assert.Equal(t, "sk-test", config.APIKey)
	assert.NotNil(t, config.Temperature)
	assert.Equal(t, 0.8, *config.Temperature)
	assert.NotNil(t, config.MaxTokens)
	assert.Equal(t, 8000, *config.MaxTokens)
}

func TestLLMConfigModel_TableName(t *testing.T) {
	config := LLMConfigModel{}
	assert.Equal(t, "llm_configs", config.TableName())
}
