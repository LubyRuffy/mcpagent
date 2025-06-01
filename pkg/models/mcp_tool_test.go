package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMCPToolModel_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tool    MCPToolModel
		wantErr error
	}{
		{
			name: "valid tool",
			tool: MCPToolModel{
				Name:     "test_tool",
				ServerID: 1,
				ToolKey:  "server_test_tool",
			},
			wantErr: nil,
		},
		{
			name: "empty name",
			tool: MCPToolModel{
				Name:     "",
				ServerID: 1,
				ToolKey:  "server_test_tool",
			},
			wantErr: ErrMCPToolNameEmpty,
		},
		{
			name: "empty server ID",
			tool: MCPToolModel{
				Name:     "test_tool",
				ServerID: 0,
				ToolKey:  "server_test_tool",
			},
			wantErr: ErrMCPToolServerIDEmpty,
		},
		{
			name: "empty tool key",
			tool: MCPToolModel{
				Name:     "test_tool",
				ServerID: 1,
				ToolKey:  "",
			},
			wantErr: ErrMCPToolKeyEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tool.Validate()
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMCPToolModel_SetInputSchema(t *testing.T) {
	tool := &MCPToolModel{}

	tests := []struct {
		name    string
		schema  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid schema",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
				},
				"required": []string{"query"},
			},
			wantErr: false,
		},
		{
			name:    "nil schema",
			schema:  nil,
			wantErr: false,
		},
		{
			name:    "empty schema",
			schema:  map[string]interface{}{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tool.SetInputSchema(tt.schema)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.schema == nil {
					assert.Equal(t, "", tool.InputSchema)
				} else {
					assert.NotEqual(t, "", tool.InputSchema)
				}
			}
		})
	}
}

func TestMCPToolModel_GetInputSchema(t *testing.T) {
	tool := &MCPToolModel{}

	// 测试空模式
	schema, err := tool.GetInputSchema()
	assert.NoError(t, err)
	assert.Nil(t, schema)

	// 测试有效模式
	inputSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query",
			},
		},
		"required": []string{"query"},
	}

	err = tool.SetInputSchema(inputSchema)
	assert.NoError(t, err)

	retrievedSchema, err := tool.GetInputSchema()
	assert.NoError(t, err)
	assert.Equal(t, "object", retrievedSchema["type"])
	
	properties, ok := retrievedSchema["properties"].(map[string]interface{})
	assert.True(t, ok)
	
	query, ok := properties["query"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "string", query["type"])
	assert.Equal(t, "Search query", query["description"])

	// 测试无效JSON
	tool.InputSchema = "invalid json"
	_, err = tool.GetInputSchema()
	assert.Error(t, err)
}

func TestGenerateToolKey(t *testing.T) {
	tests := []struct {
		name       string
		serverName string
		toolName   string
		expected   string
	}{
		{
			name:       "normal case",
			serverName: "test-server",
			toolName:   "test-tool",
			expected:   "test-server_test-tool",
		},
		{
			name:       "empty server name",
			serverName: "",
			toolName:   "test-tool",
			expected:   "_test-tool",
		},
		{
			name:       "empty tool name",
			serverName: "test-server",
			toolName:   "",
			expected:   "test-server_",
		},
		{
			name:       "both empty",
			serverName: "",
			toolName:   "",
			expected:   "_",
		},
		{
			name:       "with special characters",
			serverName: "test-server-123",
			toolName:   "test_tool.v1",
			expected:   "test-server-123_test_tool.v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateToolKey(tt.serverName, tt.toolName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMCPToolModel_ToMCPToolInfo(t *testing.T) {
	now := time.Now()
	tool := MCPToolModel{
		ID:          1,
		Name:        "test_tool",
		Description: "Test tool description",
		ServerID:    1,
		Server: MCPServerConfigModel{
			Name: "test-server",
		},
		ToolKey:    "test-server_test_tool",
		IsActive:   true,
		LastSyncAt: &now,
	}

	info := tool.ToMCPToolInfo()

	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "test_tool", info.Name)
	assert.Equal(t, "Test tool description", info.Description)
	assert.Equal(t, "test-server", info.Server)
	assert.Equal(t, "test-server_test_tool", info.ToolKey)
	assert.True(t, info.IsActive)
	assert.Equal(t, &now, info.LastSyncAt)
}

func TestMCPToolModel_TableName(t *testing.T) {
	tool := MCPToolModel{}
	assert.Equal(t, "mcp_tools", tool.TableName())
}
