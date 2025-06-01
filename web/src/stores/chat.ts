import { defineStore } from 'pinia'
import { ref, computed, nextTick } from 'vue'
import type { ChatMessage, NotifyEvent, TaskStatus, UserInput } from '@/types/notify'
import type { ChatState, ChatConfig } from '@/types/chat'
import { SSEManager } from '@/utils/sse'
import { useConfigStore } from './config'

export const useChatStore = defineStore('chat', () => {
  // 状态
  const messages = ref<ChatMessage[]>([])
  const currentTask = ref<TaskStatus | null>(null)
  const isConnected = ref(false)
  const isTyping = ref(false)
  const inputHistory = ref<string[]>([])
  const historyIndex = ref(-1)
  const sseManager = ref<SSEManager | null>(null)

  // 配置
  const config = ref<ChatConfig>({
    autoScroll: true,
    showTimestamp: true,
    showToolDetails: false,
    maxMessages: 1000
  })

  // 计算属性
  const lastMessage = computed(() => {
    return messages.value[messages.value.length - 1]
  })

  const hasMessages = computed(() => {
    return messages.value.length > 0
  })

  const isTaskRunning = computed(() => {
    const status = currentTask.value?.status
    console.log('【检查任务状态】currentTask:', currentTask.value, '状态:', status)
    
    // 检查任务是否正在运行
    const running = status === 'running'
    console.log('【检查任务状态】是否运行中:', running)
    
    return running
  })

  // 方法
  const initSSE = () => {
    sseManager.value = new SSEManager()

    sseManager.value.onConnect(() => {
      console.log('【聊天】SSE连接已建立')
      isConnected.value = true
    })

    sseManager.value.onDisconnect(() => {
      console.log('【聊天】SSE连接已断开')
      isConnected.value = false
    })

    sseManager.value.onNotify((event: NotifyEvent) => {
      console.log('【聊天】收到通知事件:', event.type)
      handleNotifyEvent(event)
    })

    sseManager.value.onTaskStatus((status: TaskStatus) => {
      console.log('【聊天】任务状态更新:', status, '当前状态为:', currentTask.value?.status)
      currentTask.value = status
      console.log('【聊天】更新后的任务运行状态:', isTaskRunning.value, '状态:', status.status)
      
      // 如果任务已完成或出错，确保重置typing状态
      if (status.status === 'completed' || status.status === 'error') {
        isTyping.value = false
      }
    })

    sseManager.value.onError((error: Error) => {
      console.error('【聊天】SSE错误:', error)
      isConnected.value = false
    })

    // 不再自动连接，只在发送任务时连接
  }

  const sendMessage = async (input: UserInput) => {
    if (!sseManager.value) {
      throw new Error('SSE管理器未初始化')
    }

    // 获取当前配置
    const configStore = useConfigStore()

    // 添加用户消息
    const userMessage: ChatMessage = {
      id: generateMessageId(),
      type: 'user',
      content: input.content,
      timestamp: Date.now(),
      events: []
    }

    addMessage(userMessage)

    // 添加到历史记录
    addToHistory(input.content)

    try {
      console.log('【chat】准备发送任务，当前所选工具:', configStore.config.mcp.tools)
      
      // 强制加载MCP服务器配置，确保有最新数据
      console.log('【chat】强制加载MCP服务器配置开始')
      await configStore.loadMCPServerConfigs(true)
      console.log('【chat】MCP服务器配置加载完成，共', configStore.mcpServerConfigs.length, '个服务器配置')
      
      // 确保工具列表非空时，强制构建MCP配置
      if (configStore.config.mcp.tools && configStore.config.mcp.tools.length > 0) {
        console.log('【chat】检测到工具列表非空，强制构建MCP配置')
        configStore.buildMCPConfigFromDatabase()
        console.log('【chat】MCP配置构建完成，当前mcp_servers有', 
                   Object.keys(configStore.config.mcp.mcp_servers || {}).length, '个服务器')
      }
      
      // 获取后端兼容的配置
      console.log('【chat】获取后端兼容配置开始')
      const backendConfig = configStore.getConfigForBackend()
      
      // 详细检查配置
      console.log('【chat】后端兼容配置获取完成:')
      console.log('- 工具数量:', backendConfig.mcp?.tools?.length || 0)
      console.log('- 服务器数量:', Object.keys(backendConfig.mcp?.mcp_servers || {}).length)
      
      // 配置校验
      if (!backendConfig.mcp) {
        console.error('【chat】后端配置中缺少mcp字段')
        throw new Error('配置错误：mcp字段缺失')
      }
      
      if (!backendConfig.mcp.tools || backendConfig.mcp.tools.length === 0) {
        console.warn('【chat】后端配置中工具列表为空')
        // 仍然继续，不抛出错误，因为可能是用户故意没选工具
      }
      
      if (!backendConfig.mcp.mcp_servers || Object.keys(backendConfig.mcp.mcp_servers).length === 0) {
        console.error('【chat】后端配置中mcp_servers为空，尝试修复')
        
        // 应急修复：添加默认服务器
        backendConfig.mcp.mcp_servers = {
          'ddg': {
            command: 'ddg',
            args: [],
            env: {},
            status: 'disconnected'
          }
        }
        
        console.log('【chat】添加应急服务器后，mcp_servers:', JSON.stringify(backendConfig.mcp.mcp_servers))
      }
      
      // 详细记录最终发送的配置
      console.log('【chat】最终发送的配置:', JSON.stringify({
        tools: backendConfig.mcp.tools,
        mcp_servers_count: Object.keys(backendConfig.mcp.mcp_servers || {}).length,
        mcp_servers_keys: Object.keys(backendConfig.mcp.mcp_servers || {})
      }))
      
      // 发送任务并建立SSE连接
      const taskId = await sseManager.value.sendTask(input.content, backendConfig)
      console.log(`任务已提交，任务ID: ${taskId}`)

      // 创建助手消息占位符
      const assistantMessage: ChatMessage = {
        id: generateMessageId(),
        type: 'assistant',
        content: '',
        timestamp: Date.now(),
        events: []
      }

      addMessage(assistantMessage)
      isTyping.value = true
    } catch (error) {
      // 添加错误消息
      const errorMessage: any = {
        id: generateMessageId(),
        type: 'error',
        content: error instanceof Error ? error.message : '发送消息失败',
        timestamp: Date.now(),
        events: []
      }
      addMessage(errorMessage)
      throw error
    }
  }

  const handleNotifyEvent = (event: NotifyEvent) => {
    const lastMsg = messages.value[messages.value.length - 1]

    if (!lastMsg || lastMsg.type !== 'assistant') {
      return
    }

    // 添加事件到最后一条助手消息
    lastMsg.events = lastMsg.events || []
    lastMsg.events.push(event)

    switch (event.type) {
      case 'message':
        lastMsg.content += event.content
        break
      case 'thinking':
        // 思考过程可以显示在单独的区域
        break
      case 'tool_call':
        // 工具调用事件
        break
      case 'result':
        lastMsg.content = event.content
        isTyping.value = false
        break
      case 'error':
        lastMsg.content = `错误: ${event.error}`
        isTyping.value = false
        break
    }

    // 自动滚动
    if (config.value.autoScroll) {
      nextTick(() => {
        scrollToBottom()
      })
    }
  }

  const addMessage = (message: ChatMessage) => {
    messages.value.push(message)

    // 限制消息数量
    if (messages.value.length > config.value.maxMessages) {
      messages.value = messages.value.slice(-config.value.maxMessages)
    }
  }

  const clearMessages = () => {
    messages.value = []
    currentTask.value = null
    isTyping.value = false
  }

  const addToHistory = (content: string) => {
    if (content.trim() && !inputHistory.value.includes(content)) {
      inputHistory.value.push(content)
      if (inputHistory.value.length > 50) {
        inputHistory.value = inputHistory.value.slice(-50)
      }
    }
    historyIndex.value = -1
  }

  const getPreviousHistory = (): string | null => {
    if (inputHistory.value.length === 0) return null

    if (historyIndex.value === -1) {
      historyIndex.value = inputHistory.value.length - 1
    } else if (historyIndex.value > 0) {
      historyIndex.value--
    }

    return inputHistory.value[historyIndex.value]
  }

  const getNextHistory = (): string | null => {
    if (inputHistory.value.length === 0 || historyIndex.value === -1) return null

    if (historyIndex.value < inputHistory.value.length - 1) {
      historyIndex.value++
      return inputHistory.value[historyIndex.value]
    } else {
      historyIndex.value = -1
      return ''
    }
  }

  const scrollToBottom = () => {
    const chatContainer = document.querySelector('.chat-messages')
    if (chatContainer) {
      chatContainer.scrollTop = chatContainer.scrollHeight
    }
  }

  const generateMessageId = (): string => {
    return `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  const updateConfig = (newConfig: Partial<ChatConfig>) => {
    config.value = { ...config.value, ...newConfig }
  }

  const disconnect = () => {
    if (sseManager.value) {
      sseManager.value.disconnect()
      sseManager.value = null
    }
    isConnected.value = false
  }

  // 停止当前任务
  const stopTask = async () => {
    if (!sseManager.value || !isTaskRunning.value) {
      console.warn('没有正在运行的任务可以停止')
      return
    }

    try {
      const result = await sseManager.value.stopTask()
      if (result) {
        // 停止成功，等待服务器发送任务状态更新
        console.log('任务停止请求已发送，等待服务器响应')
      }
    } catch (error) {
      console.error('停止任务失败:', error)
      
      // 即使后端请求失败，也尝试更新本地状态
      if (currentTask.value) {
        currentTask.value.status = 'error'
        isTyping.value = false
        
        // 更新最后一条消息，指示任务被用户中断
        if (lastMessage.value && lastMessage.value.type === 'assistant') {
          lastMessage.value.content += '\n\n[任务已被用户中断]'
        }
      }
    }
  }

  return {
    // 状态
    messages,
    currentTask,
    isTyping,
    inputHistory,
    historyIndex,
    config,
    sseManager,

    // 计算属性
    lastMessage,
    hasMessages,
    isTaskRunning,

    // 方法
    initSSE,
    sendMessage,
    handleNotifyEvent,
    addMessage,
    clearMessages,
    addToHistory,
    getPreviousHistory,
    getNextHistory,
    scrollToBottom,
    updateConfig,
    disconnect,
    stopTask
  }
})
