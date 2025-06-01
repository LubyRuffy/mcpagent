import type { NotifyEvent, TaskStatus, SSEMessage } from '@/types/notify'

export class SSEManager {
  private eventSource: EventSource | null = null
  private isManualClose = false
  private currentTaskId: string | null = null

  // 事件回调
  private onConnectCallback?: () => void
  private onDisconnectCallback?: () => void
  private onNotifyCallback?: (event: NotifyEvent) => void
  private onTaskStatusCallback?: (status: TaskStatus) => void
  private onErrorCallback?: (error: Error) => void

  constructor(private url: string = '/events') {}

  // 为特定任务建立SSE连接
  connectForTask(taskId: string): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        // 如果已有连接，先断开
        if (this.eventSource) {
          this.disconnect()
        }

        this.isManualClose = false
        this.currentTaskId = taskId

        // 建立新的SSE连接，包含任务ID
        const urlWithTaskId = `${this.url}?taskId=${taskId}`
        this.eventSource = new EventSource(urlWithTaskId)

        this.eventSource.onopen = () => {
          console.log(`SSE连接已建立，任务ID: ${taskId}`)
          this.onConnectCallback?.()
          resolve()
        }

        this.eventSource.onmessage = (event) => {
          this.handleMessage(event.data)
        }

        this.eventSource.onerror = (error) => {
          console.error('SSE连接错误:', error)
          this.onDisconnectCallback?.()

          if (!this.isManualClose) {
            const sseError = new Error('SSE连接错误')
            this.onErrorCallback?.(sseError)
            reject(sseError)
          }
        }

      } catch (error) {
        const connectError = new Error('无法创建SSE连接')
        this.onErrorCallback?.(connectError)
        reject(connectError)
      }
    })
  }

  // 保留原有的connect方法以兼容现有代码
  connect(): Promise<void> {
    return this.connectForTask('default')
  }

  disconnect(): void {
    this.isManualClose = true

    if (this.eventSource) {
      this.eventSource.close()
      this.eventSource = null
    }
  }

  // 发送任务并建立SSE连接
  async sendTask(task: string, config?: any): Promise<string> {
    try {
      const requestBody: any = { task }
      if (config) {
        console.log('【SSE】准备发送任务:', task)
        console.log('【SSE】任务配置:', JSON.stringify(config))
        
        // 特别记录工具配置和服务器信息
        if (config.mcp) {
          if (config.mcp.tools) {
            console.log('【SSE】工具列表详情:')
            config.mcp.tools.forEach((tool: string, index: number) => {
              console.log(`【SSE】工具[${index}]: ${tool}`)
            })
          }
          
          if (config.mcp.mcp_servers) {
            console.log('【SSE】服务器列表详情:')
            Object.entries(config.mcp.mcp_servers).forEach(([name, server]: [string, any]) => {
              console.log(`【SSE】服务器[${name}]:`, JSON.stringify(server))
            })
          }
        }
        
        requestBody.config = config
      }
      
      console.log('【SSE】发送请求体:', JSON.stringify(requestBody))

      const response = await fetch('/api/task', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestBody),
      })

      if (!response.ok) {
        const errorText = await response.text()
        console.error('【SSE】任务请求失败:', response.status, errorText)
        throw new Error(`发送任务失败: ${errorText}`)
      }

      const result = await response.json()
      console.log('【SSE】任务提交结果:', result)
      
      if (result.success && result.task_id) {
        // 任务提交成功后，立即建立SSE连接
        console.log('【SSE】准备建立连接，任务ID:', result.task_id)
        await this.connectForTask(result.task_id)
        return result.task_id
      } else {
        throw new Error(result.message || '任务提交失败')
      }
    } catch (error) {
      console.error('【SSE】任务发送失败:', error)
      throw new Error('发送任务失败: ' + (error instanceof Error ? error.message : '未知错误'))
    }
  }

  async sendConfig(config: any): Promise<void> {
    try {
      const response = await fetch('/api/config', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(config),
      })

      if (!response.ok) {
        throw new Error('发送配置失败')
      }
    } catch (error) {
      throw new Error('发送配置失败: ' + (error instanceof Error ? error.message : '未知错误'))
    }
  }

  private handleMessage(data: string): void {
    try {
      console.log('【SSE】收到消息:', data)
      const message: SSEMessage = JSON.parse(data)
      console.log('【SSE】解析消息类型:', message.type, '数据:', message.data)

      switch (message.type) {
        case 'notify':
          this.onNotifyCallback?.(message.data as NotifyEvent)
          break
        case 'status':
          const status = message.data as TaskStatus
          console.log('【SSE】任务状态更新:', status)
          this.onTaskStatusCallback?.(status)

          // 如果任务完成或出错，自动断开连接
          if (status.status === 'completed' || status.status === 'error') {
            console.log(`【SSE】任务 ${status.id} 已完成或出错，状态: ${status.status}，准备断开连接`)
            setTimeout(() => this.disconnect(), 1000) // 延迟1秒断开，确保最后的消息都收到
          }
          break
        case 'config':
          // 处理配置更新
          break
        case 'ping':
          // 心跳消息，忽略
          break
        default:
          console.warn('未知的消息类型:', message.type)
      }
    } catch (error) {
      console.error('解析SSE消息失败:', error)
    }
  }

  // 获取当前任务ID
  getCurrentTaskId(): string | null {
    return this.currentTaskId
  }

  // 停止当前正在执行的任务
  async stopTask(): Promise<boolean> {
    if (!this.currentTaskId) {
      console.warn('【SSE】没有正在执行的任务可以停止')
      return false
    }

    try {
      console.log(`【SSE】准备停止任务: ${this.currentTaskId}`)
      
      const response = await fetch(`/api/task/${this.currentTaskId}/cancel`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        }
      })

      if (!response.ok) {
        const errorText = await response.text()
        console.error('【SSE】停止任务请求失败:', response.status, errorText)
        throw new Error(`停止任务失败: ${errorText}`)
      }

      const result = await response.json()
      console.log('【SSE】停止任务结果:', result)
      
      if (result.success) {
        console.log(`【SSE】任务 ${this.currentTaskId} 已成功停止`)
        // 注意：服务器会发送任务状态更新，这将触发连接断开
        return true
      } else {
        throw new Error(result.message || '停止任务失败')
      }
    } catch (error) {
      console.error('【SSE】停止任务失败:', error)
      throw new Error('停止任务失败: ' + (error instanceof Error ? error.message : '未知错误'))
    }
  }

  // 事件监听器
  onConnect(callback: () => void): void {
    this.onConnectCallback = callback
  }

  onDisconnect(callback: () => void): void {
    this.onDisconnectCallback = callback
  }

  onNotify(callback: (event: NotifyEvent) => void): void {
    this.onNotifyCallback = callback
  }

  onTaskStatus(callback: (status: TaskStatus) => void): void {
    this.onTaskStatusCallback = callback
  }

  onError(callback: (error: Error) => void): void {
    this.onErrorCallback = callback
  }

  // 获取连接状态
  get isConnected(): boolean {
    return this.eventSource?.readyState === EventSource.OPEN
  }

  get readyState(): number {
    return this.eventSource?.readyState ?? EventSource.CLOSED
  }
}
