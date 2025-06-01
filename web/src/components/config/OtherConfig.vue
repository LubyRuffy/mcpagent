<template>
  <el-card class="config-card">
    <template #header>
      <span>{{ $t('config.other.title') }}</span>
    </template>

    <!-- 网络代理 -->
    <div class="config-row proxy-container">
      <div class="proxy-header">
        <div class="config-label">{{ $t('config.other.proxy') }}</div>
        <el-switch
          v-model="proxyEnabled"
          @change="handleProxyToggle as any"
        />
      </div>
      <div v-if="proxyEnabled" class="proxy-input-container">
        <el-input
          v-model="proxyUrl"
          placeholder="http://127.0.0.1:8080"
          @blur="validateProxyUrl"
        />
      </div>
    </div>

    <!-- 最大步数 -->
    <div class="config-row">
      <div class="config-label">{{ $t('config.other.maxStep') }}</div>
      <div class="config-value">
        <el-input-number
          v-model="otherConfig.max_step"
          :min="1"
          :max="100"
          :step="1"
          controls-position="right"
        />
      </div>
    </div>

    <!-- 日志级别 -->
    <div class="config-row">
      <div class="config-label">{{ $t('config.other.logLevel') }}</div>
      <div class="config-value">
        <el-select v-model="logLevel">
          <el-option label="Debug" value="debug" />
          <el-option label="Info" value="info" />
          <el-option label="Warning" value="warning" />
          <el-option label="Error" value="error" />
        </el-select>
      </div>
    </div>

    <!-- 高级设置 -->
    <el-collapse v-model="advancedOpen">
      <el-collapse-item :title="$t('config.other.advanced.title')" name="advanced">
        <!-- 请求超时 -->
        <div class="config-row">
          <div class="config-label">{{ $t('config.other.advanced.requestTimeout') }}</div>
          <div class="config-value">
            <div class="timeout-wrapper">
              <el-input-number
                v-model="requestTimeout"
                :min="5"
                :max="300"
                :step="5"
                controls-position="right"
              />
              <span class="unit-text">{{ $t('config.other.advanced.seconds') }}</span>
            </div>
          </div>
        </div>

        <!-- 重试次数 -->
        <div class="config-row">
          <div class="config-label">{{ $t('config.other.advanced.retryCount') }}</div>
          <div class="config-value">
            <el-input-number
              v-model="retryCount"
              :min="0"
              :max="10"
              :step="1"
              controls-position="right"
            />
          </div>
        </div>

        <!-- 并发限制 -->
        <div class="config-row">
          <div class="config-label">{{ $t('config.other.advanced.concurrencyLimit') }}</div>
          <div class="config-value">
            <el-input-number
              v-model="concurrencyLimit"
              :min="1"
              :max="20"
              :step="1"
              controls-position="right"
            />
          </div>
        </div>

        <!-- 调试模式 -->
        <div class="config-row">
          <div class="config-label">{{ $t('config.other.advanced.debugMode') }}</div>
          <div class="config-value">
            <el-switch
              v-model="debugMode"
              :active-text="$t('config.other.advanced.enable')"
              :inactive-text="$t('config.other.advanced.disable')"
            />
          </div>
        </div>
        
        <!-- 聊天记录 -->
        <div class="config-row">
          <div class="config-label">{{ $t('config.other.advanced.chatHistory') }}</div>
          <div class="config-value">
            <div class="button-group">
              <el-button size="small" type="warning" @click="clearChatHistory">{{ $t('config.other.advanced.clearChatHistory') }}</el-button>
            </div>
          </div>
        </div>
      </el-collapse-item>
    </el-collapse>
  </el-card>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useConfigStore } from '@/stores/config'
import { clearChatHistory as clearLocalChatHistory } from '@/utils/storage'
import { useChatStore } from '@/stores/chat'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const configStore = useConfigStore()
const chatStore = useChatStore()

// 响应式状态
const advancedOpen = ref<string[]>([])
const logLevel = ref('info')
const requestTimeout = ref(30)
const retryCount = ref(3)
const concurrencyLimit = ref(5)
const debugMode = ref(false)

// 计算属性
const otherConfig = computed(() => configStore.config)

const proxyEnabled = computed({
  get: () => !!configStore.config.proxy,
  set: (value: boolean) => {
    if (!value) {
      configStore.config.proxy = ''
    }
  }
})

const proxyUrl = computed({
  get: () => configStore.config.proxy,
  set: (value: string) => {
    configStore.config.proxy = value
  }
})

// 方法
const handleProxyToggle = (enabled: boolean) => {
  if (!enabled) {
    proxyUrl.value = ''
  } else if (!proxyUrl.value) {
    proxyUrl.value = 'http://127.0.0.1:8080'
  }
}

const validateProxyUrl = () => {
  if (!proxyUrl.value) return true

  try {
    new URL(proxyUrl.value)
    return true
  } catch {
    ElMessage.warning(t('validation.invalidUrl'))
    return false
  }
}

// 使用本地存储清空聊天记录
const clearChatHistory = () => {
  try {
    clearLocalChatHistory()
    chatStore.clearMessages()
    ElMessage.success(t('success.chatHistoryCleared'))
  } catch (error) {
    ElMessage.error(t('error.clearChatHistoryFailed'))
    console.error(t('error.clearChatHistoryFailed'), error)
  }
}
</script>

<style scoped>
.config-card {
  background-color: var(--bg-color);
  border-radius: 6px;
  overflow: hidden;
}

.config-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid var(--border-color-light);
}

.config-row:last-child {
  border-bottom: none;
}

.config-label {
  font-size: 13px;
  color: var(--text-color-primary);
}

.config-value {
  display: flex;
  align-items: center;
}

.proxy-container {
  flex-direction: column;
  align-items: flex-start;
}

.proxy-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  margin-bottom: 6px;
}

.proxy-input-container {
  width: 100%;
  margin-top: 6px;
}

.timeout-wrapper {
  display: flex;
  align-items: center;
}

.unit-text {
  margin-left: 6px;
  font-size: 12px;
}

.button-group {
  display: flex;
  gap: 6px;
}

:deep(.el-divider) {
  margin: 12px 0;
}

:deep(.el-divider__text) {
  font-size: 13px;
  padding: 0 10px;
}

:deep(.el-collapse-item__header) {
  padding: 6px 0;
  font-size: 13px;
  height: auto;
}

:deep(.el-collapse-item__content) {
  padding: 8px 0;
}

:deep(.el-input-number) {
  width: 100px;
}

:deep(.el-input-number__decrease), 
:deep(.el-input-number__increase) {
  width: 22px;
}

:deep(.el-select) {
  width: 100px;
}

@media (max-width: 500px) {
  .config-label {
    font-size: 12px;
  }
  
  .unit-text {
    font-size: 11px;
  }
  
  :deep(.el-divider__text) {
    font-size: 12px;
    padding: 0 8px;
  }
  
  :deep(.el-collapse-item__header) {
    font-size: 12px;
  }
  
  :deep(.el-input-number) {
    width: 90px;
  }
  
  :deep(.el-select) {
    width: 90px;
  }
}
</style>
