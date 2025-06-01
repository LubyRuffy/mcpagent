// 通知事件类型定义，对应Go后端的Notify接口

export type NotifyEventType = 'message' | 'thinking' | 'tool_call' | 'result' | 'error'

export interface BaseNotifyEvent {
  type: NotifyEventType
  timestamp: number
  id: string
}

export interface MessageEvent extends BaseNotifyEvent {
  type: 'message'
  content: string
}

export interface ThinkingEvent extends BaseNotifyEvent {
  type: 'thinking'
  content: string
}

export interface ToolCallEvent extends BaseNotifyEvent {
  type: 'tool_call'
  tool_name: string
  parameters: any
  status?: 'calling' | 'success' | 'error'
  result?: any
  error?: string
}

export interface ResultEvent extends BaseNotifyEvent {
  type: 'result'
  content: string
}

export interface ErrorEvent extends BaseNotifyEvent {
  type: 'error'
  error: string
  details?: any
}

export type NotifyEvent = MessageEvent | ThinkingEvent | ToolCallEvent | ResultEvent | ErrorEvent

// SSE消息类型
export interface SSEMessage {
  type: 'notify' | 'config' | 'status' | 'ping'
  data: any
}

// WebSocket消息类型（保留向后兼容）
export interface WebSocketMessage {
  type: 'notify' | 'config' | 'status'
  data: any
}

// 聊天消息类型
export interface ChatMessage {
  id: string
  type: 'user' | 'assistant' | 'system' | 'error'
  content: string
  timestamp: number
  events?: NotifyEvent[]
}

// 任务执行状态
export interface TaskStatus {
  id: string
  status: 'pending' | 'running' | 'completed' | 'error'
  progress?: number
  current_step?: string
  total_steps?: number
}

// 用户输入
export interface UserInput {
  content: string
  files?: File[]
}
