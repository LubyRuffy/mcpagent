<template>
  <div class="message-list">
    <!-- 欢迎消息 -->
    <div v-if="!chatStore.hasMessages" class="welcome-message">
      <div class="welcome-content">
        <el-icon class="welcome-icon"><ChatDotRound /></el-icon>
        <h3>{{ $t('welcome.title') }}</h3>
        <p>{{ $t('welcome.description') }}</p>

        <div class="example-tasks">
          <h4>{{ $t('welcome.exampleTitle') }}:</h4>
          <div class="task-examples">
            <el-tag
              v-for="example in taskExamples"
              :key="example"
              class="example-tag"
              @click="selectExample(example)"
            >
              {{ $t(example) }}
            </el-tag>
          </div>
        </div>
      </div>
    </div>

    <!-- 消息列表 -->
    <transition-group
      name="message"
      tag="div"
      class="messages"
    >
      <MessageItem
        v-for="message in chatStore.messages"
        :key="message.id"
        :message="message"
        :show-timestamp="chatStore.config.showTimestamp"
        :show-tool-details="chatStore.config.showToolDetails"
      />
    </transition-group>

    <!-- 正在输入指示器 -->
    <div v-if="chatStore.isTyping" class="typing-indicator">
      <div class="typing-content">
        <el-avatar :size="32" class="typing-avatar">
          AI
        </el-avatar>

        <div class="typing-text">
          <div class="typing-dots">
            <span></span>
            <span></span>
            <span></span>
          </div>
          <span class="typing-label">{{ $t('chat.message.thinking') }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ChatDotRound } from '@element-plus/icons-vue'
import { useChatStore } from '@/stores/chat'
import MessageItem from './MessageItem.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const chatStore = useChatStore()

// 示例任务
const taskExamples = [
  'welcome.example1',
  'welcome.example2',
  'welcome.example3',
  'welcome.example4'
]

// 方法
const selectExample = (example: string) => {
  // 触发父组件的输入事件
  const event = new CustomEvent('select-example', {
    detail: { example: t(example) }
  })
  document.dispatchEvent(event)
}
</script>

<style scoped>
.message-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 100%;
}

.welcome-message {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
  padding: 40px 20px;
}

.welcome-content {
  text-align: center;
  max-width: 500px;
}

.welcome-icon {
  font-size: 48px;
  color: var(--primary-color);
  margin-bottom: 16px;
}

.welcome-content h3 {
  margin: 0 0 12px 0;
  font-size: 24px;
  color: var(--text-color-primary);
}

.welcome-content p {
  margin: 0 0 24px 0;
  color: var(--text-color-secondary);
  line-height: 1.6;
}

.example-tasks {
  text-align: left;
}

.example-tasks h4 {
  margin: 0 0 12px 0;
  font-size: 16px;
  color: var(--text-color-primary);
}

.task-examples {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.example-tag {
  cursor: pointer;
  transition: all var(--transition-duration) ease;
  padding: 8px 12px;
  border-radius: var(--border-radius-base);
  text-align: left;
}

.example-tag:hover {
  background-color: var(--primary-color);
  color: white;
  transform: translateY(-1px);
}

.messages {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.typing-indicator {
  display: flex;
  justify-content: flex-start;
  margin-top: 8px;
}

.typing-content {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  max-width: 80%;
}

.typing-avatar {
  background-color: var(--primary-color);
  color: white;
  flex-shrink: 0;
}

.typing-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px 16px;
  background-color: var(--bg-color-secondary);
  border-radius: var(--border-radius-base);
  border: 1px solid var(--border-color);
}

.typing-dots {
  display: flex;
  gap: 4px;
}

.typing-dots span {
  width: 6px;
  height: 6px;
  background-color: var(--primary-color);
  border-radius: 50%;
  animation: typing 1.4s infinite ease-in-out;
}

.typing-dots span:nth-child(1) {
  animation-delay: -0.32s;
}

.typing-dots span:nth-child(2) {
  animation-delay: -0.16s;
}

.typing-label {
  font-size: 12px;
  color: var(--text-color-secondary);
}

@keyframes typing {
  0%, 80%, 100% {
    transform: scale(0.8);
    opacity: 0.5;
  }
  40% {
    transform: scale(1);
    opacity: 1;
  }
}

/* 消息动画 */
.message-enter-active {
  transition: all 0.3s ease;
}

.message-enter-from {
  opacity: 0;
  transform: translateY(20px);
}

.message-leave-active {
  transition: all 0.3s ease;
}

.message-leave-to {
  opacity: 0;
  transform: translateY(-20px);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .welcome-message {
    padding: 20px 16px;
  }

  .welcome-content {
    max-width: 100%;
  }

  .welcome-content h3 {
    font-size: 20px;
  }

  .task-examples {
    gap: 6px;
  }

  .example-tag {
    padding: 6px 10px;
    font-size: 14px;
  }
}
</style>
