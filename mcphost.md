# mcp host

一个支持读取mcpServers json文件的模块。

## 需求
- 支持加载mcpServers.json文件
- 支持指定工具的启用
- 支持工具的调用
- 支持获取工具列表，给 cloudwego/eino-ext 的model使用

## 问题
- 两个server有相同名称的tool，应该如何处理？server_toolName作为一个唯一主键
- ai大模型会如何调用？不应该是一个tools列表吗？是的，返回一个