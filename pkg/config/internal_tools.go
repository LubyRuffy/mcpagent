package config

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/tool/sequentialthinking"
	"github.com/cloudwego/eino/components/tool"
)

// 定义内置工具的常量
const (
	// InnerServerName 是内置工具服务器的名称
	InnerServerName = "inner"

	// SequentialThinkingToolName 是顺序思考工具的名称
	SequentialThinkingToolName = "sequentialthinking"
)

// GetInternalTools 返回所有内置工具列表
func GetInternalTools(ctx context.Context) ([]tool.BaseTool, error) {
	var tools []tool.BaseTool

	// 创建顺序思考工具并进行自定义
	seqThinking, err := sequentialthinking.NewTool()
	if err != nil {
		return nil, fmt.Errorf("创建顺序思考工具失败: %w", err)
	}

	// 由于我们不能直接修改 eino 工具的名称，我们需要在生成 toolKey 时使用自定义名称
	// 在 SyncInternalTools 中使用工具信息时会自动使用我们定义的常量

	tools = append(tools, seqThinking)

	// 这里可以继续添加其他内置工具...

	return tools, nil
}

// GetInternalToolMap 返回内置工具的映射表，键为工具名称，值为工具实例
func GetInternalToolMap(ctx context.Context) (map[string]tool.BaseTool, error) {
	tools, err := GetInternalTools(ctx)
	if err != nil {
		return nil, err
	}

	toolMap := make(map[string]tool.BaseTool)
	for _, t := range tools {
		toolInfo, err := t.Info(ctx)
		if err != nil {
			return nil, fmt.Errorf("获取工具 %T 信息失败: %w", t, err)
		}
		toolMap[toolInfo.Name] = t
	}

	return toolMap, nil
}
