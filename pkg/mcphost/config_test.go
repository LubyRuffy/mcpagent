package mcphost

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 测试ServerConfig的GetTimeoutDuration方法
func TestServerConfig_GetTimeoutDuration(t *testing.T) {
	// 测试默认超时
	config := ServerConfig{}
	assert.Equal(t, time.Duration(DefaultMCPTimeoutSeconds)*time.Second, config.GetTimeoutDuration())

	// 测试自定义超时
	customTimeout := 60 * time.Second
	config.Timeout = customTimeout
	assert.Equal(t, customTimeout, config.GetTimeoutDuration())

	// 测试最小超时
	minTimeout := time.Duration(MinMCPTimeoutSeconds) * time.Second
	config.Timeout = minTimeout
	assert.Equal(t, minTimeout, config.GetTimeoutDuration())
}

// 测试validateSettings函数的各种情况
func TestValidateSettings_AdditionalCases(t *testing.T) {
	// 测试空的MCPServers
	emptySettings := &MCPSettings{
		MCPServers: map[string]ServerConfig{},
	}
	err := validateSettings(emptySettings)
	assert.NoError(t, err)

	// 测试默认传输类型（空字符串）
	defaultTransportSettings := &MCPSettings{
		MCPServers: map[string]ServerConfig{
			"default_transport": {
				TransportType: "", // 默认为stdio
				Command:       "echo",
			},
		},
	}
	err = validateSettings(defaultTransportSettings)
	assert.NoError(t, err)

	// 测试刚好达到最小超时的配置
	minTimeoutSettings := &MCPSettings{
		MCPServers: map[string]ServerConfig{
			"min_timeout": {
				TransportType: "stdio",
				Command:       "echo",
				Timeout:       time.Duration(MinMCPTimeoutSeconds) * time.Second,
			},
		},
	}
	err = validateSettings(minTimeoutSettings)
	assert.NoError(t, err)
}

// 测试LoadSettingsFromString函数的边界情况
func TestLoadSettingsFromString_EdgeCases(t *testing.T) {
	// 测试空配置字符串
	_, err := LoadSettingsFromString("")
	assert.Error(t, err)

	// 测试空的JSON对象
	emptyJSON := `{}`
	settings, err := LoadSettingsFromString(emptyJSON)
	assert.NoError(t, err)
	assert.NotNil(t, settings)
	assert.Empty(t, settings.MCPServers)

	// 测试只有mcpServers字段但为空的情况
	emptyServersJSON := `{"mcpServers":{}}`
	settings, err = LoadSettingsFromString(emptyServersJSON)
	assert.NoError(t, err)
	assert.NotNil(t, settings)
	assert.Empty(t, settings.MCPServers)

	// 测试无效的配置（验证失败）
	invalidConfigJSON := `{
		"mcpServers": {
			"invalid_server": {
				"transportType": "invalid"
			}
		}
	}`
	_, err = LoadSettingsFromString(invalidConfigJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid settings")
}
