# MCP Agent Web UI

基于Vue3 + TypeScript的MCP Agent Web界面，提供直观的图形化界面来管理和使用MCP Agent。

## ✨ 功能特性

### 🎛️ 配置管理
- **LLM配置**: 支持OpenAI和Ollama等多种模型提供商
- **MCP服务器管理**: 可视化添加、编辑、删除MCP服务器
- **工具选择**: 直观的工具选择界面，支持多选和描述预览
- **系统提示词**: 内置模板和自定义编辑器，支持占位符配置
- **其他设置**: 代理配置、日志级别、最大步数等

### 💬 聊天交互
- **实时对话**: Server-Sent Events (SSE)连接，支持流式响应
- **消息类型**: 支持用户消息、助手回复、系统通知
- **工具调用展示**: 可展开查看工具调用参数和结果
- **Markdown渲染**: 支持代码高亮和表格渲染
- **历史记录**: 输入历史和对话记录管理

### 🌐 多语言支持
- 中文（简体）
- English
- 支持动态切换

### 🎨 主题系统
- 浅色主题
- 深色主题
- 跟随系统设置

### 📱 响应式设计
- 适配桌面端和移动端
- 左右分栏布局（30%/70%）
- 移动端侧边栏折叠

## 🚀 快速开始

### 环境要求

- Node.js 18+
- Go 1.24+
- 现代浏览器（支持Server-Sent Events）

### 安装依赖

```bash
# 进入web目录
cd web

# 安装前端依赖
npm install
# 或使用yarn
yarn install
# 或使用pnpm
pnpm install
```

### 开发模式

```bash
# 启动前端开发服务器
npm run dev

# 在另一个终端启动后端服务器
cd ..
go run cmd/mcpagent-web/main.go -config config_public.yaml -dev
```

前端开发服务器将在 http://localhost:3000 启动
后端API服务器将在 http://localhost:8080 启动

### 生产构建

```bash
# 构建前端
npm run build

# 构建后端
go build -o mcpagent-web cmd/mcpagent-web/main.go

# 启动生产服务器
./mcpagent-web -config config_public.yaml
```

## 📁 项目结构

```
web/
├── public/                 # 静态资源
├── src/
│   ├── components/         # Vue组件
│   │   ├── layout/        # 布局组件
│   │   ├── config/        # 配置组件
│   │   ├── chat/          # 聊天组件
│   │   └── common/        # 通用组件
│   ├── stores/            # Pinia状态管理
│   ├── types/             # TypeScript类型定义
│   ├── utils/             # 工具函数
│   ├── locales/           # 国际化文件
│   └── styles/            # 样式文件
├── package.json
├── vite.config.ts
└── tsconfig.json
```

## 🔧 配置说明

### 后端配置

后端使用YAML配置文件，示例：

```yaml
llm:
  type: openai
  base_url: https://api.openai.com/v1
  model: gpt-4
  api_key: your-api-key

mcp:
  config_file: mcp_servers.json
  tools:
    - fetch_fetch
    - ddg-search_search

system_prompt: |
  你是一个智能助手...

max_step: 20
```

### 前端配置

前端配置通过环境变量或构建时配置：

```bash
# 开发环境
VITE_API_BASE_URL=http://localhost:8080
VITE_SSE_URL=http://localhost:8080/events

# 生产环境
VITE_API_BASE_URL=/api
VITE_SSE_URL=/events
```

## 🌐 API接口

### Server-Sent Events接口

- **连接**: `http://localhost:8080/events`
- **消息格式**:
  ```json
  {
    "type": "notify|status|ping",
    "data": {...}
  }
  ```

### HTTP API接口

- `GET /api/config` - 获取配置
- `POST /api/config` - 更新配置
- `POST /api/task` - 执行任务

## 🎯 使用指南

### 1. 配置LLM

1. 在左侧配置面板选择LLM类型
2. 填入相应的API地址和密钥
3. 选择模型名称
4. 点击"测试连接"验证配置

### 2. 配置MCP服务器

1. 点击"添加服务器"
2. 填入服务器名称和启动命令
3. 配置参数和环境变量
4. 保存配置

### 3. 选择工具

1. 在工具选择区域勾选需要的工具
2. 查看工具描述了解功能
3. 保存配置

### 4. 开始对话

1. 在右侧输入框输入任务描述
2. 点击发送或使用Ctrl+Enter快捷键
3. 查看实时响应和工具调用过程

## 🔍 故障排除

### 常见问题

1. **SSE连接失败**
   - 检查后端服务是否启动
   - 确认端口号是否正确
   - 检查防火墙设置

2. **配置保存失败**
   - 检查配置格式是否正确
   - 确认必填字段已填写
   - 查看浏览器控制台错误信息

3. **工具调用失败**
   - 检查MCP服务器是否正常运行
   - 确认工具配置是否正确
   - 查看后端日志

### 调试模式

启用开发模式获取更多调试信息：

```bash
# 前端调试
npm run dev

# 后端调试
go run cmd/mcpagent-web/main.go -dev -config config.yaml
```

## 🤝 贡献指南

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建Pull Request

## 📄 许可证

本项目采用MIT许可证 - 查看[LICENSE](../LICENSE)文件了解详情。
