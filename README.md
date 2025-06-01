# MCPAgent

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

MCPAgent 是一个基于 Model Context Protocol (MCP) 的智能代理框架，支持多种大语言模型和工具集成，专为测试单体agent的效果而设计。

## 🔑 背景
目前的Agent设计基本上依赖大模型本身的能力来进行思考和规划，对于一个没有特定业务经验的人来说，能够解决从0分到80分的问题，
但是对于特定行业场景来说，一个经验丰富的专家通过控制工具的调用能够做到95分，这时候，经验流和特定场景的MCP工具就比大模型本身能力更重要。

事实上我们也通过一些方式证明了，在网络安全领域，靠大模型本身的能力去做规划和执行，任何一个所谓的最顶级的大模型的输出都是一个灾难，有惊喜有惊吓，
很难满足生产环境的需求。而从另一个方面我们也证明了，即便是14b的小模型（甚至是qwen3 4b）在提供领域的工具和限定最佳实践的流程情况下，
输出远比通用大模型效果更好。

```
通用大模型 + 自规划流程 + 通用工具 + 公开数据 < 小模型 + 最佳实践的限定流程 + 垂直领域的数据 + 垂直领域的工具
```
无论从效果还是成本考虑，未来都会变成不同垂直领域的经验流的固化和落地。

市面上没有这样的工具来进行基于MCP的Agent效果测试，像是类似cherry studio这样的综合型工具存在两个问题：
一个是内嵌固化了很长的通用提示词反而干扰了效果，
另一个是他们还是基于以前CoT方式的用文本做工具调用的流程，而不是直接用到tools的API接口来适配最佳实践。

基于上面的考虑，我们想做一个专为“测试单体Agent效果”而设计的GUI工具。

## ✨ 功能特性

- 🤖 **多模型支持**: 支持 OpenAI 和 Ollama 等多种 LLM 提供商
- 🔧 **工具集成**: 基于 MCP 协议的丰富工具生态系统
- 🎯 **ReAct 架构**: 采用推理-行动循环的智能代理模式
- ⚙️ **灵活配置**: 支持 YAML 配置文件和命令行参数
- 🔄 **实时通知**: 提供任务执行过程的实时反馈
- 🌐 **Web界面**: 基于Vue3的现代化Web UI，支持实时交互
- 📱 **响应式设计**: 适配桌面端和移动端设备
- 🌍 **多语言支持**: 中文和英文界面切换
- 🎨 **主题切换**: 明暗主题自由切换

## 🚀 快速开始

### 安装

- 安装主程序

```bash
# 克隆项目
git clone https://github.com/LubyRuffy/mcpagent.git
cd mcpagent

# 安装依赖
go mod download

# 构建项目
go build -o mcpagent ./cmd/mcpagent
```

- 安装mcp依赖

```shell
# uvx
curl -LsSf https://astral.sh/uv/install.sh | sh
# pip install uv

# npx
brew install node
npm install -g npx
```

- 安装ollama(可选)

```shell
brew install ollama
```

默认使用qwen3，你也配置配置任何一个兼容openai的api接口。

### 基本使用

#### 命令行模式

```bash
# 使用默认配置执行任务
./mcpagent -task "分析网络安全领域的最新研究趋势"

# 使用自定义配置文件
./mcpagent -config news_config.yaml -task "分析特朗普的一些新政策对中美关系的影响"

# 使用数据库的mcp server可以直接做数据查询
./mcpagent -config dbagent_config.yaml -task "最近的用户查询最多的产品是什么"
```

#### Web界面模式

```bash
# 构建Web应用
./scripts/build-web.sh

# 启动Web服务器
cd build
./start.sh

# 或者直接运行开发模式
./scripts/dev.sh
```

访问 http://localhost:8080 使用Web界面。

**Web界面特性：**
- 🎛️ 可视化配置管理（LLM、MCP服务器、工具选择）
- 💬 实时聊天交互，支持流式响应
- 🔧 工具调用过程可视化展示
- 📝 系统提示词模板和自定义编辑
- 🌐 多语言界面（中文/英文）
- 🎨 明暗主题切换
- 📱 响应式设计，支持移动端

> ✅ **Web UI已完全实现并可正常使用！** 详细使用指南请查看 [WEB_UI_USAGE_GUIDE.md](WEB_UI_USAGE_GUIDE.md)

## ⚙️ 配置说明

### 主配置文件 (config.yaml)

```yaml
# LLM 配置
llm:
  type: ollama              # 支持 openai, ollama
  base_url: http://127.0.0.1:11434
  model: qwen3:14b
  api_key: ollama

# MCP 服务器配置
mcp:
  config_file: mcp_servers.json
  #mcp_servers: 也可以直接配置mcp_servers
  tools:
    - fetch_fetch
    - ddg-search_search
    - sequential-thinking_sequentialthinking

# 代理配置
proxy: ""                  # HTTP代理地址（比如burp），用于调试查看大模型的请求和响应
max_step: 20               # 最大推理步数

# 系统提示词
field: "网络安全领域" # 用于system_prompt的{field}占位符
system_prompt: |
  你是一位经验丰富的学术研究员...
```

### MCP 服务器配置 (mcp_servers.json)

参考 [官方文档](https://modelcontextprotocol.io/quickstart/user)

```json
{
  "mcpServers": {
    "ddg-search": {
      "command": "uvx",
      "args": ["duckduckgo-mcp-server"]
    },
    "sequential-thinking": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"]
    },
    "fetch": {
      "command": "uvx",
      "args": ["mcp-server-fetch"]
    }
  }
}
```

也可以直接在yaml配置文件中进行配置。

## 📖 使用示例

### 学术论文撰写

```bash
./mcphost -task "撰写一篇关于'大语言模型在网络安全中的应用'的学术论文，包含文献综述、方法论和案例分析，字数不少于2000字"
```

### 研究趋势分析

```bash
./mcphost -task "分析2024年人工智能安全领域的最新研究趋势，总结主要技术发展和挑战"
```

## ❓ 常见问题

### Q: 如何添加新的 MCP 服务器？

A: 在 `mcp_servers.json` 中添加新的服务器配置，然后在 `config.yaml` 的 `mcp.tools` 中启用相应工具。

### Q: 支持哪些大语言模型？

A: 目前支持 OpenAI API 兼容的模型和 Ollama 本地模型。可以通过配置文件切换。

### Q: 如何自定义系统提示词？

A: 修改配置文件中的 `system_prompt` 字段，或使用 `-system-prompt` 命令行参数。

## 📄 许可证

本项目采用 MIT 许可证。详情请查看 [LICENSE](LICENSE) 文件。

## 🤝 致谢

- [Cloudwego Eino](https://github.com/cloudwego/eino) - AI 应用开发框架
- [MCP Protocol](https://modelcontextprotocol.io/) - 模型上下文协议
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - Go MCP 实现
- [Testify](https://github.com/stretchr/testify) - Go 测试框架
- [Viper](https://github.com/spf13/viper) - Go 配置管理

## 📞 支持与反馈

- **Issues**: [GitHub Issues](https://github.com/LubyRuffy/mcpagent/issues)
- **Discussions**: [GitHub Discussions](https://github.com/LubyRuffy/mcpagent/discussions)
- **Email**: 项目相关问题请通过 GitHub Issues 提交

