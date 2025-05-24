package config

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
)

//go:generate mockgen -destination=mcphub_mock.go -package=config github.com/LubyRuffy/mcpagent/pkg/config MCPHubInterface

// MCPHubInterface 定义了 MCPHub 的接口
type MCPHubInterface interface {
	GetEinoTools(ctx context.Context, toolNameList []string) ([]tool.BaseTool, error)
	CloseServers() error
}
