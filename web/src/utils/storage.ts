import type { AppConfig } from '@/types/config'

const STORAGE_KEYS = {
  CONFIG: 'mcpagent-config',
  CHAT_HISTORY: 'mcpagent-chat-history',
  USER_PREFERENCES: 'mcpagent-preferences'
}

// 配置存储
export const saveConfig = async (config: AppConfig): Promise<void> => {
  try {
    const configStr = JSON.stringify(config, null, 2)
    localStorage.setItem(STORAGE_KEYS.CONFIG, configStr)
  } catch (error) {
    throw new Error('保存配置失败: ' + (error instanceof Error ? error.message : '未知错误'))
  }
}

export const loadConfig = async (): Promise<AppConfig | null> => {
  try {
    const configStr = localStorage.getItem(STORAGE_KEYS.CONFIG)
    if (!configStr) return null

    return JSON.parse(configStr) as AppConfig
  } catch (error) {
    console.error('加载配置失败:', error)
    return null
  }
}

export const clearConfig = (): void => {
  localStorage.removeItem(STORAGE_KEYS.CONFIG)
}

// 配置验证
export interface ValidationResult {
  valid: boolean
  errors: string[]
}

export const validateConfig = (config: AppConfig): ValidationResult => {
  const errors: string[] = []

  // LLM配置验证
  if (!config.llm.type) {
    errors.push('LLM类型不能为空')
  }

  if (!config.llm.base_url) {
    errors.push('LLM Base URL不能为空')
  } else {
    try {
      new URL(config.llm.base_url)
    } catch {
      errors.push('LLM Base URL格式不正确')
    }
  }

  if (!config.llm.model) {
    errors.push('LLM模型不能为空')
  }

  if (!config.llm.api_key) {
    errors.push('LLM API Key不能为空')
  }

  // MCP配置验证 - 现在只检查mcp_servers是否存在
  if (!config.mcp.mcp_servers) {
    errors.push('MCP服务器配置不能为空')
  }

  // 代理配置验证
  if (config.proxy) {
    try {
      new URL(config.proxy)
    } catch {
      errors.push('代理URL格式不正确')
    }
  }

  // 最大步数验证
  if (config.max_step <= 0) {
    errors.push('最大步数必须大于0')
  }

  return {
    valid: errors.length === 0,
    errors
  }
}

// 聊天历史存储
export const saveChatHistory = (messages: any[]): void => {
  try {
    const historyStr = JSON.stringify(messages)
    localStorage.setItem(STORAGE_KEYS.CHAT_HISTORY, historyStr)
  } catch (error) {
    console.error('保存聊天历史失败:', error)
  }
}

export const loadChatHistory = (): any[] => {
  try {
    const historyStr = localStorage.getItem(STORAGE_KEYS.CHAT_HISTORY)
    if (!historyStr) return []

    return JSON.parse(historyStr)
  } catch (error) {
    console.error('加载聊天历史失败:', error)
    return []
  }
}

export const clearChatHistory = (): void => {
  localStorage.removeItem(STORAGE_KEYS.CHAT_HISTORY)
}

// 用户偏好设置
export interface UserPreferences {
  theme: 'light' | 'dark' | 'auto'
  language: 'zh-CN' | 'en-US'
  sidebarCollapsed: boolean
  chatConfig: {
    autoScroll: boolean
    showTimestamp: boolean
    showToolDetails: boolean
  }
}

export const saveUserPreferences = (preferences: UserPreferences): void => {
  try {
    const preferencesStr = JSON.stringify(preferences)
    localStorage.setItem(STORAGE_KEYS.USER_PREFERENCES, preferencesStr)
  } catch (error) {
    console.error('保存用户偏好失败:', error)
  }
}

export const loadUserPreferences = (): UserPreferences | null => {
  try {
    const preferencesStr = localStorage.getItem(STORAGE_KEYS.USER_PREFERENCES)
    if (!preferencesStr) return null

    return JSON.parse(preferencesStr) as UserPreferences
  } catch (error) {
    console.error('加载用户偏好失败:', error)
    return null
  }
}

// 导入/导出配置
export const exportConfig = (config: AppConfig): void => {
  const configStr = JSON.stringify(config, null, 2)
  const blob = new Blob([configStr], { type: 'application/json' })
  const url = URL.createObjectURL(blob)

  const a = document.createElement('a')
  a.href = url
  a.download = `mcpagent-config-${new Date().toISOString().split('T')[0]}.json`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)

  URL.revokeObjectURL(url)
}

export const importConfig = (): Promise<AppConfig> => {
  return new Promise((resolve, reject) => {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = '.json'

    input.onchange = (event) => {
      const file = (event.target as HTMLInputElement).files?.[0]
      if (!file) {
        reject(new Error('未选择文件'))
        return
      }

      const reader = new FileReader()
      reader.onload = (e) => {
        try {
          const config = JSON.parse(e.target?.result as string) as AppConfig
          const validation = validateConfig(config)

          if (!validation.valid) {
            reject(new Error('配置文件无效: ' + validation.errors.join(', ')))
            return
          }

          resolve(config)
        } catch (error) {
          reject(new Error('解析配置文件失败'))
        }
      }

      reader.readAsText(file)
    }

    input.click()
  })
}
