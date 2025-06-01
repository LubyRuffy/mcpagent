import type { NotifyEvent, ChatMessage, TaskStatus } from './notify'

// 聊天相关状态
export interface ChatState {
  messages: ChatMessage[]
  currentTask: TaskStatus | null
  isConnected: boolean
  isTyping: boolean
  inputHistory: string[]
  historyIndex: number
}

// 用户输入
export interface UserInput {
  content: string
  files?: File[]
}

// 聊天配置
export interface ChatConfig {
  autoScroll: boolean
  showTimestamp: boolean
  showToolDetails: boolean
  maxMessages: number
}

// 消息渲染选项
export interface MessageRenderOptions {
  showMarkdown: boolean
  showCodeHighlight: boolean
  showToolCallDetails: boolean
  maxContentLength?: number
}
