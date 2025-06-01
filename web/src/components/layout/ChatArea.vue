<template>
  <div class="chat-area">
    <!-- 聊天消息列表 -->
    <div class="chat-messages" ref="messagesContainer">
      <el-scrollbar ref="scrollbarRef">
        <div class="messages-content">
          <MessageList />
        </div>
      </el-scrollbar>
    </div>

    <!-- 输入区域 -->
    <div class="chat-input">
      <InputArea />
    </div>



    <!-- 任务执行状态 -->
    <el-alert
      v-if="chatStore.currentTask && chatStore.isTaskRunning"
      :title="$t('task.running') + ': ' + (chatStore.currentTask.current_step || $t('task.processing'))"
      type="info"
      :closable="false"
      show-icon
      class="task-alert"
    >
      <template #default>
        <div class="task-progress">
          <el-progress
            v-if="chatStore.currentTask.progress !== undefined"
            :percentage="chatStore.currentTask.progress"
            :show-text="false"
            :stroke-width="4"
          />
          <span class="task-info">
            {{ chatStore.currentTask.current_step }}
            <span v-if="chatStore.currentTask.total_steps">
              ({{ $t('task.steps', { total: chatStore.currentTask.total_steps }) }})
            </span>
          </span>
        </div>
      </template>
    </el-alert>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, watch } from 'vue'
import { useChatStore } from '@/stores/chat'
import MessageList from '@/components/chat/MessageList.vue'
import InputArea from '@/components/chat/InputArea.vue'

const chatStore = useChatStore()

// 引用
const messagesContainer = ref<HTMLElement>()
const scrollbarRef = ref()

// 方法

const scrollToBottom = () => {
  nextTick(() => {
    if (scrollbarRef.value) {
      scrollbarRef.value.setScrollTop(scrollbarRef.value.wrapRef.scrollHeight)
    }
  })
}

// 监听消息变化，自动滚动到底部
watch(
  () => chatStore.messages.length,
  () => {
    if (chatStore.config.autoScroll) {
      scrollToBottom()
    }
  }
)

// 监听消息内容变化（流式输出）
watch(
  () => chatStore.lastMessage?.content,
  () => {
    if (chatStore.config.autoScroll) {
      scrollToBottom()
    }
  }
)
</script>

<style scoped>
.chat-area {
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: var(--bg-color);
  position: relative;
}

.chat-messages {
  flex: 1;
  overflow: hidden;
  padding: 12px;
}

.messages-content {
  min-height: 100%;
  display: flex;
  flex-direction: column;
}

.chat-input {
  border-top: 1px solid var(--border-color);
  background-color: var(--bg-color);
}

.task-alert {
  position: absolute;
  top: 12px;
  left: 12px;
  right: 12px;
  z-index: 10;
  border-radius: var(--border-radius-base);
}

.alert-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.task-progress {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
}

.task-info {
  font-size: 14px;
  color: var(--text-color-primary);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .chat-messages {
    padding: 12px;
  }

  .task-alert {
    left: 12px;
    right: 12px;
  }
}
</style>
