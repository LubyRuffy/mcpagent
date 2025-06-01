<template>
  <div
    class="message-item"
    :class="{
      'user-message': message.type === 'user',
      'assistant-message': message.type === 'assistant',
      'system-message': message.type === 'system'
    }"
  >
    <!-- 用户消息 -->
    <div v-if="message.type === 'user'" class="message-content user-content">
      <div class="message-header">
        <el-avatar :size="32" class="user-avatar">
          <el-icon><User /></el-icon>
        </el-avatar>

        <div class="message-meta">
          <span class="message-sender">{{ $t('chat.message.user') }}</span>
          <span v-if="showTimestamp" class="message-time">
            {{ formatTime(message.timestamp) }}
          </span>
        </div>

        <div class="message-actions">
          <el-button
            size="small"
            circle
            @click="copyMessage(message.content)"
          >
            <el-icon><DocumentCopy /></el-icon>
          </el-button>
        </div>
      </div>

      <div class="message-body">
        <div class="message-text">{{ message.content }}</div>
      </div>
    </div>

    <!-- 助手消息 -->
    <div v-else-if="message.type === 'assistant'" class="message-content assistant-content">
      <div class="message-header">
        <el-avatar :size="32" class="assistant-avatar">
          AI
        </el-avatar>

        <div class="message-meta">
          <span class="message-sender">{{ $t('chat.message.assistant') }}</span>
          <span v-if="showTimestamp" class="message-time">
            {{ formatTime(message.timestamp) }}
          </span>
        </div>

        <div class="message-actions">
          <el-button
            size="small"
            circle
            @click="copyMessage(message.content)"
          >
            <el-icon><DocumentCopy /></el-icon>
          </el-button>
        </div>
      </div>

      <div class="message-body">
        <!-- 消息内容 -->
        <div v-if="message.content" class="message-text">
          <MarkdownRenderer :content="message.content" />
        </div>

        <!-- 事件列表 -->
        <div v-if="message.events && message.events.length > 0" class="message-events">
          <div
            v-for="event in message.events"
            :key="event.id"
            class="event-item"
          >
            <!-- 思考事件 -->
            <div v-if="event.type === 'thinking'" class="thinking-event">
              <el-icon class="event-icon"><InfoFilled /></el-icon>
              <span class="event-text">{{ event.content }}</span>
            </div>

            <!-- 工具调用事件 -->
            <ToolCallDisplay
              v-else-if="event.type === 'tool_call'"
              :tool-call="event"
              :show-details="showToolDetails"
            />

            <!-- 错误事件 -->
            <div v-else-if="event.type === 'error'" class="error-event">
              <span class="error-icon">⚠️</span>
              <span class="event-text">{{ event.error }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 系统消息 -->
    <div v-else class="message-content system-content">
      <div class="system-message-body">
        <el-icon class="system-icon"><InfoFilled /></el-icon>
        <span class="system-text">{{ message.content }}</span>
        <span v-if="showTimestamp" class="system-time">
          {{ formatTime(message.timestamp) }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import {
  User,
  DocumentCopy,
  InfoFilled
} from '@element-plus/icons-vue'
import type { ChatMessage } from '@/types/notify'
import MarkdownRenderer from './MarkdownRenderer.vue'
import ToolCallDisplay from './ToolCallDisplay.vue'

interface Props {
  message: ChatMessage
  showTimestamp?: boolean
  showToolDetails?: boolean
}

defineProps<Props>()

// 方法
const formatTime = (timestamp: number) => {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  // 如果是今天，只显示时间
  if (diff < 24 * 60 * 60 * 1000 && date.getDate() === now.getDate()) {
    return date.toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  // 否则显示日期和时间
  return date.toLocaleString('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const copyMessage = async (content: string) => {
  try {
    await navigator.clipboard.writeText(content)
    ElMessage.success('消息已复制到剪贴板')
  } catch {
    // 降级方案
    const textArea = document.createElement('textarea')
    textArea.value = content
    document.body.appendChild(textArea)
    textArea.select()
    document.execCommand('copy')
    document.body.removeChild(textArea)
    ElMessage.success('消息已复制到剪贴板')
  }
}
</script>

<style scoped>
.message-item {
  display: flex;
  margin-bottom: 16px;
}

.user-message {
  justify-content: flex-end;
}

.assistant-message,
.system-message {
  justify-content: flex-start;
}

.message-content {
  max-width: 80%;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.user-content {
  align-items: flex-end;
}

.assistant-content {
  align-items: flex-start;
}

.message-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-avatar {
  background-color: var(--primary-color);
  color: white;
}

.assistant-avatar {
  background-color: var(--success-color);
  color: white;
}

.message-meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.message-sender {
  font-weight: 600;
  font-size: 14px;
  color: var(--text-color-primary);
}

.message-time {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.message-actions {
  margin-left: auto;
}

.message-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.message-text {
  padding: 12px 16px;
  border-radius: var(--border-radius-base);
  line-height: 1.6;
}

.user-content .message-text {
  background-color: var(--primary-color);
  color: white;
  border-bottom-right-radius: 4px;
}

.assistant-content .message-text {
  background-color: var(--bg-color-secondary);
  border: 1px solid var(--border-color);
  color: var(--text-color-primary);
  border-bottom-left-radius: 4px;
}

.message-events {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-left: 40px;
}

.event-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.thinking-event {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background-color: var(--bg-color-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-base);
  font-size: 14px;
  color: var(--text-color-secondary);
}

.error-event {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background-color: #fef0f0;
  border: 1px solid #fbc4c4;
  border-radius: var(--border-radius-base);
  font-size: 14px;
  color: var(--danger-color);
}

.event-icon {
  font-size: 16px;
  flex-shrink: 0;
}

.error-icon {
  color: var(--danger-color);
}

.event-text {
  flex: 1;
  word-break: break-word;
}

.system-content {
  width: 100%;
  max-width: 100%;
  align-items: center;
}

.system-message-body {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: var(--info-color);
  color: white;
  border-radius: var(--border-radius-base);
  font-size: 14px;
}

.system-icon {
  font-size: 16px;
  flex-shrink: 0;
}

.system-text {
  flex: 1;
}

.system-time {
  font-size: 12px;
  opacity: 0.8;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .message-content {
    max-width: 90%;
  }

  .message-events {
    margin-left: 20px;
  }

  .message-header {
    gap: 6px;
  }

  .message-text {
    padding: 10px 12px;
  }
}
</style>
