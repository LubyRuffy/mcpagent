package models

import (
	"testing"

	"github.com/LubyRuffy/einomcphost"
	"github.com/stretchr/testify/assert"
)

func TestMCPServerConfigModel_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  MCPServerConfigModel
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid stdio config",
			config: MCPServerConfigModel{
				Name:          "test-server",
				TransportType: "stdio",
				Command:       "uvx",
			},
			wantErr: false,
		},
		{
			name: "valid sse config",
			config: MCPServerConfigModel{
				Name:          "sse-server",
				TransportType: "sse",
				URL:           "http://localhost:8000/sse",
			},
			wantErr: false,
		},
		{
			name: "valid http config",
			config: MCPServerConfigModel{
				Name:          "http-server",
				TransportType: "http",
				URL:           "http://localhost:8000/api",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			config: MCPServerConfigModel{
				TransportType: "stdio",
				Command:       "uvx",
			},
			wantErr: true,
			errMsg:  "MCP服务器配置名称不能为空",
		},
		{
			name: "stdio without command",
			config: MCPServerConfigModel{
				Name:          "test-server",
				TransportType: "stdio",
			},
			wantErr: true,
			errMsg:  "MCP服务器启动命令不能为空",
		},
		{
			name: "sse without url",
			config: MCPServerConfigModel{
				Name:          "sse-server",
				TransportType: "sse",
			},
			wantErr: true,
			errMsg:  "MCP服务器URL不能为空",
		},
		{
			name: "invalid transport type",
			config: MCPServerConfigModel{
				Name:          "test-server",
				TransportType: "invalid",
			},
			wantErr: true,
			errMsg:  "MCP服务器传输类型无效，仅支持 stdio、sse 和 http",
		},
		{
			name: "backward compatibility - empty transport type defaults to stdio",
			config: MCPServerConfigModel{
				Name:    "legacy-server",
				Command: "uvx",
			},
			wantErr: false,
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

func TestMCPServerConfigModel_ToServerConfig(t *testing.T) {
	config := MCPServerConfigModel{
		Name:          "test-server",
		TransportType: "stdio",
		Command:       "uvx",
		Args:          `["duckduckgo-mcp-server"]`,
		Env:           `{"TEST_VAR":"test_value"}`,
		Disabled:      false,
	}

	serverConfig, err := config.ToServerConfig()
	assert.NoError(t, err)

	expected := einomcphost.ServerConfig{
		TransportType: "stdio",
		Command:       "uvx",
		Args:          []string{"duckduckgo-mcp-server"},
		Env:           map[string]string{"TEST_VAR": "test_value"},
		Disabled:      false,
	}

	assert.Equal(t, expected, serverConfig)
}

func TestMCPServerConfigModel_ToServerConfig_SSE(t *testing.T) {
	config := MCPServerConfigModel{
		Name:          "sse-server",
		TransportType: "sse",
		URL:           "http://localhost:8000/sse",
		Headers:       `["Authorization: Bearer token"]`,
		Disabled:      false,
	}

	serverConfig, err := config.ToServerConfig()
	assert.NoError(t, err)

	expected := einomcphost.ServerConfig{
		TransportType: "sse",
		URL:           "http://localhost:8000/sse",
		Disabled:      false,
	}

	assert.Equal(t, expected, serverConfig)
}

func TestMCPServerConfigModel_ToServerConfig_DefaultTransportType(t *testing.T) {
	// Test backward compatibility - empty transport type should default to stdio
	config := MCPServerConfigModel{
		Name:     "legacy-server",
		Command:  "uvx",
		Args:     `["test-server"]`,
		Disabled: false,
	}

	serverConfig, err := config.ToServerConfig()
	assert.NoError(t, err)

	expected := einomcphost.ServerConfig{
		TransportType: "stdio",
		Command:       "uvx",
		Args:          []string{"test-server"},
		Disabled:      false,
	}

	assert.Equal(t, expected, serverConfig)
}

func TestMCPServerConfigModel_FromServerConfig(t *testing.T) {
	serverConfig := einomcphost.ServerConfig{
		Command:  "uvx",
		Args:     []string{"duckduckgo-mcp-server"},
		Env:      map[string]string{"TEST_VAR": "test_value"},
		Disabled: false,
	}

	var config MCPServerConfigModel
	err := config.FromServerConfig("test-server", "Test description", serverConfig)
	assert.NoError(t, err)

	assert.Equal(t, "test-server", config.Name)
	assert.Equal(t, "Test description", config.Description)
	assert.Equal(t, "uvx", config.Command)
	assert.Equal(t, `["duckduckgo-mcp-server"]`, config.Args)
	assert.Equal(t, `{"TEST_VAR":"test_value"}`, config.Env)
	assert.Equal(t, false, config.Disabled)
}

func TestMCPServerConfigModel_GetArgsSlice(t *testing.T) {
	config := MCPServerConfigModel{
		Args: `["arg1","arg2","arg3"]`,
	}

	args, err := config.GetArgsSlice()
	assert.NoError(t, err)
	assert.Equal(t, []string{"arg1", "arg2", "arg3"}, args)

	// Test empty args
	config.Args = ""
	args, err = config.GetArgsSlice()
	assert.NoError(t, err)
	assert.Equal(t, []string{}, args)
}

func TestMCPServerConfigModel_GetEnvMap(t *testing.T) {
	config := MCPServerConfigModel{
		Env: `{"VAR1":"value1","VAR2":"value2"}`,
	}

	env, err := config.GetEnvMap()
	assert.NoError(t, err)
	expected := map[string]string{
		"VAR1": "value1",
		"VAR2": "value2",
	}
	assert.Equal(t, expected, env)

	// Test empty env
	config.Env = ""
	env, err = config.GetEnvMap()
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{}, env)
}

func TestMCPServerConfigModel_SetArgs(t *testing.T) {
	var config MCPServerConfigModel

	err := config.SetArgs([]string{"arg1", "arg2"})
	assert.NoError(t, err)
	assert.Equal(t, `["arg1","arg2"]`, config.Args)

	// Test empty args
	err = config.SetArgs([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "", config.Args)
}

func TestMCPServerConfigModel_SetEnv(t *testing.T) {
	var config MCPServerConfigModel

	env := map[string]string{
		"VAR1": "value1",
		"VAR2": "value2",
	}

	err := config.SetEnv(env)
	assert.NoError(t, err)
	// JSON marshaling order is not guaranteed, so we parse it back to compare
	parsedEnv, err := config.GetEnvMap()
	assert.NoError(t, err)
	assert.Equal(t, env, parsedEnv)

	// Test empty env
	err = config.SetEnv(map[string]string{})
	assert.NoError(t, err)
	assert.Equal(t, "", config.Env)
}

func TestMCPServerConfigModel_TableName(t *testing.T) {
	config := MCPServerConfigModel{}
	assert.Equal(t, "mcp_server_configs", config.TableName())
}

func TestMCPServerConfigModel_GetHeadersSlice(t *testing.T) {
	config := MCPServerConfigModel{
		Headers: `["Authorization: Bearer token", "Content-Type: application/json"]`,
	}

	headers, err := config.GetHeadersSlice()
	assert.NoError(t, err)
	assert.Equal(t, []string{"Authorization: Bearer token", "Content-Type: application/json"}, headers)

	// Test empty headers
	config.Headers = ""
	headers, err = config.GetHeadersSlice()
	assert.NoError(t, err)
	assert.Equal(t, []string{}, headers)
}

func TestMCPServerConfigModel_SetHeaders(t *testing.T) {
	var config MCPServerConfigModel
	headers := []string{"Authorization: Bearer token", "Content-Type: application/json"}

	err := config.SetHeaders(headers)
	assert.NoError(t, err)
	assert.Equal(t, `["Authorization: Bearer token","Content-Type: application/json"]`, config.Headers)

	// Test empty headers
	err = config.SetHeaders([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "", config.Headers)
}
