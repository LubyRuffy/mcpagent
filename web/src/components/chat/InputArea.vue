<template>
  <div class="input-area">
    <!-- 输入框 -->
    <div class="input-container">
      <el-input
        ref="inputRef"
        v-model="inputText"
        type="textarea"
        :rows="inputRows"
        :placeholder="$t('chat.input.placeholder')"
        :disabled="isLocalTaskRunning"
        resize="none"
        @keydown="handleKeydown as any"
        @input="handleInput"
        class="message-input"
      />

      <!-- 输入工具栏 -->
      <div class="input-toolbar">
        <div class="toolbar-left">
          <!-- 文件上传 -->
          <el-upload
            ref="uploadRef"
            :show-file-list="false"
            :before-upload="handleFileUpload"
            accept=".txt,.md,.json,.csv"
            :disabled="isLocalTaskRunning"
          >
            <el-button
              size="small"
              :disabled="isLocalTaskRunning"
            >
              <el-icon><Paperclip /></el-icon>
            </el-button>
          </el-upload>

          <!-- 历史记录 -->
          <el-dropdown
            @command="selectHistory"
            :disabled="inputHistory.length === 0"
          >
            <el-button
              size="small"
              :disabled="inputHistory.length === 0"
            >
              <el-icon><Clock /></el-icon>
            </el-button>

            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item
                  v-for="(item, index) in recentHistory"
                  :key="index"
                  :command="item"
                  class="history-item"
                >
                  {{ truncateText(item, 50) }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>

          <!-- 清空对话 -->
          <el-button
            size="small"
            type="warning"
            @click="clearChat"
            :disabled="!chatStore.hasMessages"
          >
            <el-icon><Delete /></el-icon>
          </el-button>
        </div>

        <div class="toolbar-right">
          <!-- 字数统计 -->
          <span class="char-count">{{ inputText.length }}</span>

          <!-- 发送按钮 -->
          <el-button
            :type="isLocalTaskRunning ? 'danger' : 'primary'"
            :loading="false"
            :disabled="isLocalTaskRunning ? false : !canSend"
            @click="isLocalTaskRunning ? stopTask() : sendMessage()"
            class="send-button"
          >
            <el-icon v-if="isLocalTaskRunning">
              <Delete />
            </el-icon>
            <el-icon v-else>
              <Promotion />
            </el-icon>
            {{ isLocalTaskRunning ? $t('chat.input.stop') : $t('chat.input.send') }}
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  Paperclip,
  Clock,
  Delete,
  Promotion
} from '@element-plus/icons-vue'
import { useChatStore } from '@/stores/chat'

const { t } = useI18n()

const chatStore = useChatStore()

// 响应式状态
const inputText = ref('')
const inputRows = ref(3)
const inputRef = ref()
const uploadRef = ref()
const isLocalTaskRunning = ref(false) // 本地任务状态

// 监听任务状态变化
watch(() => chatStore.currentTask?.status, (newStatus) => {
  console.log('【InputArea】监听到任务状态变化:', newStatus)
  // 如果任务完成或出错，重置本地任务状态
  if (newStatus === 'completed' || newStatus === 'error') {
    console.log('【InputArea】任务已完成或出错，重置本地任务状态')
    isLocalTaskRunning.value = false
  } else if (newStatus === 'running') {
    // 确保与后端状态同步
    console.log('【InputArea】任务正在运行，设置本地任务状态')
    isLocalTaskRunning.value = true
  }
})

// 计算属性
const canSend = computed(() => {
  return inputText.value.trim().length > 0 &&
         !isLocalTaskRunning.value
})

const inputHistory = computed(() => chatStore.inputHistory)

const recentHistory = computed(() => {
  return inputHistory.value.slice(-10).reverse()
})

// 方法
const handleInput = () => {
  // 自动调整输入框高度
  const lines = inputText.value.split('\n').length
  inputRows.value = Math.min(Math.max(lines, 3), 8)
}

const handleKeydown = (event: KeyboardEvent) => {
  // Ctrl/Cmd + Enter 发送消息
  if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
    event.preventDefault()
    sendMessage()
    return
  }

  // 上下箭头浏览历史
  if (event.key === 'ArrowUp' && event.ctrlKey) {
    event.preventDefault()
    const prev = chatStore.getPreviousHistory()
    if (prev !== null) {
      inputText.value = prev
    }
    return
  }

  if (event.key === 'ArrowDown' && event.ctrlKey) {
    event.preventDefault()
    const next = chatStore.getNextHistory()
    if (next !== null) {
      inputText.value = next
    }
    return
  }

  // Escape 清空输入
  if (event.key === 'Escape') {
    inputText.value = ''
  }
}

const sendMessage = async () => {
  if (!canSend.value) return

  const message = inputText.value.trim()

  try {
    isLocalTaskRunning.value = true // 设置本地任务状态为运行中
    await chatStore.sendMessage({ content: message })
    inputText.value = ''
    inputRows.value = 3

    // 聚焦输入框
    nextTick(() => {
      inputRef.value?.focus()
    })

  } catch (error) {
    isLocalTaskRunning.value = false // 发生错误时重置本地任务状态
    ElMessage.error(error instanceof Error ? error.message : '发送消息失败')
  }
}

const selectHistory = (historyItem: string) => {
  inputText.value = historyItem

  nextTick(() => {
    inputRef.value?.focus()
  })
}

const clearChat = async () => {
  try {
    await ElMessageBox.confirm(
      t('confirm.clearChat'),
      t('common.confirmAction'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )

    chatStore.clearMessages()
    ElMessage.success(t('success.chatCleared'))
  } catch {
    // 用户取消操作
  }
}

const handleFileUpload = (file: File) => {
  // 检查文件大小（限制为5MB）
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.error(t('error.fileTooLarge', { size: '5MB' }))
    return false
  }

  // 读取文件内容
  const reader = new FileReader()
  reader.onload = (e) => {
    const content = e.target?.result as string
    if (content) {
      const filePrompt = t('chat.input.filePrompt', { 
        filename: file.name, 
        content: content 
      })
      inputText.value = filePrompt
    }
  }

  reader.onerror = () => {
    ElMessage.error('文件读取失败')
  }

  reader.readAsText(file)
  return false // 阻止自动上传
}

const truncateText = (text: string, maxLength: number) => {
  if (text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}

const stopTask = async () => {
  try {
    await chatStore.stopTask()
    ElMessage.info(t('task.stopRequest'))
    isLocalTaskRunning.value = false // 手动停止后重置本地任务状态
  } catch (error) {
    ElMessage.error(t('error.stopTaskFailed', { 
      message: error instanceof Error ? error.message : t('error.unknown') 
    }))
  }
}

// 监听示例选择事件
onMounted(() => {
  // 初始化本地任务状态
  isLocalTaskRunning.value = chatStore.isTaskRunning
  console.log('【InputArea】初始化本地任务状态:', isLocalTaskRunning.value)
  
  document.addEventListener('select-example', (event: any) => {
    const example = event.detail.example
    if (example) {
      inputText.value = example
      nextTick(() => {
        inputRef.value?.focus()
      })
    }
  })
})
</script>

<style scoped>
.input-area {
  padding: 16px;
  background-color: var(--bg-color);
  border-top: 1px solid var(--border-color);
}

.input-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.message-input :deep(.el-textarea__inner) {
  border-radius: var(--border-radius-base);
  border-color: var(--border-color);
  background-color: var(--bg-color);
  color: var(--text-color-primary);
  font-size: 14px;
  line-height: 1.5;
  resize: none;
  transition: all var(--transition-duration) ease;
}

.message-input :deep(.el-textarea__inner):focus {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.input-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.char-count {
  font-size: 12px;
  color: var(--text-color-secondary);
  min-width: 30px;
  text-align: right;
}

.history-item {
  max-width: 300px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .input-area {
    padding: 12px;
  }

  .input-toolbar {
    flex-direction: column;
    gap: 8px;
    align-items: stretch;
  }

  .toolbar-left,
  .toolbar-right {
    justify-content: space-between;
  }
}
</style>
