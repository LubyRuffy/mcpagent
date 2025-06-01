// 配置相关类型定义

export interface LLMConfig {
  type: 'openai' | 'ollama'
  base_url: string
  model: string
  api_key: string
  temperature?: number
  max_tokens?: number
}

// 数据库中保存的LLM配置
export interface LLMConfigModel {
  id: number
  name: string
  description: string
  type: 'openai' | 'ollama'
  base_url: string
  model: string
  api_key: string
  temperature?: number
  max_tokens?: number
  is_default: boolean
  is_active: boolean
  created_at: string
  updated_at: string
}

// 创建LLM配置的表单数据
export interface CreateLLMConfigForm {
  name: string
  description: string
  type: 'openai' | 'ollama'
  base_url: string
  model: string
  api_key: string
  temperature?: number
  max_tokens?: number
  is_default?: boolean
}

// 数据库中保存的MCP服务器配置
export interface MCPServerConfigModel {
  id: number
  name: string
  description: string
  command: string
  args: string // JSON格式存储的参数列表
  env: string  // JSON格式存储的环境变量
  disabled: boolean
  is_active: boolean
  created_at: string
  updated_at: string
}

// 创建MCP服务器配置的表单数据
export interface CreateMCPServerConfigForm {
  name: string
  description: string
  command: string
  args: string[]
  env: Record<string, string>
  disabled?: boolean
}

export interface MCPServer {
  command: string
  args: string[]
  env?: Record<string, string>
  status?: 'connected' | 'disconnected' | 'connecting' | 'error'
}

export interface MCPConfig {
  config_file?: string
  mcp_servers?: Record<string, MCPServer>
  tools: string[]
}

export interface ProxyConfig {
  enabled: boolean
  host: string
  port: number
  username?: string
  password?: string
}

export interface SystemPromptTemplate {
  id: string
  name: string
  content: string
  placeholders: string[]
}

// 数据库中保存的SystemPrompt配置
export interface SystemPromptModel {
  id: number
  name: string
  description: string
  content: string
  placeholders: string[] // 预定义的占位符
  is_default: boolean
  is_active: boolean
  created_at: string
  updated_at: string
}

// 创建SystemPrompt配置的表单数据
export interface CreateSystemPromptForm {
  name: string
  description: string
  content: string
  placeholders: string[]
  is_default?: boolean
}

export interface AppConfig {
  proxy: string
  mcp: MCPConfig
  llm: LLMConfig
  system_prompt: string
  max_step: number
  placeholders: Record<string, any>
}

export interface ConfigState {
  config: AppConfig
  templates: SystemPromptTemplate[]
  availableModels: string[]
  availableTools: Array<{
    name: string
    description: string
    server: string
  }>
  llmConfigs: LLMConfigModel[]
  currentLLMConfigId: number | null
  isLoading: boolean
  error: string | null
}

// 表单验证规则
export interface ValidationRule {
  required?: boolean
  message: string
  trigger?: string
  validator?: (rule: any, value: any, callback: any) => void
}
