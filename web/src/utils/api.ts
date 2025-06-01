import type { LLMConfig, AppConfig, LLMConfigModel, CreateLLMConfigForm, MCPServerConfigModel, CreateMCPServerConfigForm, SystemPromptModel, CreateSystemPromptForm } from '@/types/config'

// API基础URL
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

// API响应类型
export interface ApiResponse<T = any> {
  success?: boolean
  message?: string
  data?: T
  error?: string
}

// 扩展API响应类型，添加工具专用的响应类型
export interface ToolsApiResponse extends ApiResponse {
  tools?: Array<{ name: string; description: string; server: string }>;
}

// HTTP请求工具函数
async function request<T = any>(
  url: string,
  options: RequestInit = {}
): Promise<ApiResponse<T>> {
  try {
    const response = await fetch(`${API_BASE_URL}${url}`, {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || `HTTP ${response.status}`)
    }

    const data = await response.json()
    return data
  } catch (error) {
    console.error('API请求失败:', error)
    throw error
  }
}

// 配置相关API
export const configApi = {
  // 获取配置
  async getConfig(): Promise<AppConfig> {
    const response = await request<AppConfig>('/config')
    return response.data || response as any
  },

  // 更新配置
  async updateConfig(config: AppConfig): Promise<ApiResponse> {
    return request('/config', {
      method: 'POST',
      body: JSON.stringify(config),
    })
  },
}

// LLM相关API
export const llmApi = {
  // 测试LLM连接
  async testConnection(llmConfig: LLMConfig): Promise<ApiResponse> {
    return request('/llm/test', {
      method: 'POST',
      body: JSON.stringify(llmConfig),
    })
  },

  // 获取LLM配置列表
  async getConfigs(): Promise<ApiResponse<LLMConfigModel[]>> {
    console.log('【API】调用llmApi.getConfigs方法', new Date().toISOString(), '调用栈:', new Error().stack)
    const result = await request('/llm/configs')
    console.log('【API】llmApi.getConfigs返回结果', new Date().toISOString())
    return result
  },

  // 创建LLM配置
  async createConfig(config: CreateLLMConfigForm): Promise<ApiResponse<LLMConfigModel>> {
    return request('/llm/configs', {
      method: 'POST',
      body: JSON.stringify(config),
    })
  },

  // 获取单个LLM配置
  async getConfig(id: number): Promise<ApiResponse<LLMConfigModel>> {
    return request(`/llm/configs/${id}`)
  },

  // 更新LLM配置
  async updateConfig(id: number, config: Partial<CreateLLMConfigForm>): Promise<ApiResponse> {
    return request(`/llm/configs/${id}`, {
      method: 'PUT',
      body: JSON.stringify(config),
    })
  },

  // 删除LLM配置
  async deleteConfig(id: number): Promise<ApiResponse> {
    return request(`/llm/configs/${id}`, {
      method: 'DELETE',
    })
  },

  // 设置默认LLM配置
  async setDefaultConfig(id: number): Promise<ApiResponse> {
    return request(`/llm/configs/${id}/default`, {
      method: 'POST',
    })
  },
}

// 任务相关API
export const taskApi = {
  // 执行任务
  async executeTask(task: string, config?: AppConfig): Promise<ApiResponse> {
    const requestBody: any = { task }
    if (config) {
      console.log('【API】executeTask调用，任务:', task)
      console.log('【API】工具配置:', config.mcp.tools)
      requestBody.config = config
    }

    return request('/task', {
      method: 'POST',
      body: JSON.stringify(requestBody),
    })
  },
}

// MCP相关API
export const mcpApi = {
  // 获取MCP服务器工具列表
  async getTools(mcpServers: Record<string, any>): Promise<ToolsApiResponse> {
    return request('/mcp/tools', {
      method: 'POST',
      body: JSON.stringify({ mcp_servers: mcpServers }),
    })
  },

  // 获取已配置的MCP服务器工具列表
  async getToolsFromDB(): Promise<ToolsApiResponse> {
    return request('/mcp/tools/configured')
  },

  // 获取MCP服务器配置列表
  async getServerConfigs(): Promise<ApiResponse<MCPServerConfigModel[]>> {
    return request('/mcp/servers')
  },

  // 创建MCP服务器配置
  async createServerConfig(config: CreateMCPServerConfigForm): Promise<ApiResponse<MCPServerConfigModel>> {
    return request('/mcp/servers', {
      method: 'POST',
      body: JSON.stringify(config),
    })
  },

  // 获取单个MCP服务器配置
  async getServerConfig(id: number): Promise<ApiResponse<MCPServerConfigModel>> {
    return request(`/mcp/servers/${id}`)
  },

  // 更新MCP服务器配置
  async updateServerConfig(id: number, config: Partial<CreateMCPServerConfigForm>): Promise<ApiResponse> {
    return request(`/mcp/servers/${id}`, {
      method: 'PUT',
      body: JSON.stringify(config),
    })
  },

  // 删除MCP服务器配置
  async deleteServerConfig(id: number): Promise<ApiResponse> {
    return request(`/mcp/servers/${id}`, {
      method: 'DELETE',
    })
  },
}

// SystemPrompt相关API
export const systemPromptApi = {
  // 获取SystemPrompt配置列表
  async getPrompts(): Promise<ApiResponse<SystemPromptModel[]>> {
    return request('/system-prompts')
  },

  // 创建SystemPrompt配置
  async createPrompt(prompt: CreateSystemPromptForm): Promise<ApiResponse<SystemPromptModel>> {
    return request('/system-prompts', {
      method: 'POST',
      body: JSON.stringify(prompt),
    })
  },

  // 获取单个SystemPrompt配置
  async getPrompt(id: number): Promise<ApiResponse<SystemPromptModel>> {
    return request(`/system-prompts/${id}`)
  },

  // 更新SystemPrompt配置
  async updatePrompt(id: number, prompt: Partial<CreateSystemPromptForm>): Promise<ApiResponse> {
    return request(`/system-prompts/${id}`, {
      method: 'PUT',
      body: JSON.stringify(prompt),
    })
  },

  // 删除SystemPrompt配置
  async deletePrompt(id: number): Promise<ApiResponse> {
    return request(`/system-prompts/${id}`, {
      method: 'DELETE',
    })
  },

  // 设置默认SystemPrompt配置
  async setDefaultPrompt(id: number): Promise<ApiResponse> {
    return request(`/system-prompts/${id}/default`, {
      method: 'POST',
    })
  },
}

// 导出所有API
export default {
  config: configApi,
  llm: llmApi,
  task: taskApi,
  mcp: mcpApi,
  systemPrompt: systemPromptApi,
}
