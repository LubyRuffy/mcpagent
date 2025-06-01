<template>
  <div class="tool-call-display">
    <div class="tool-call-header" @click="toggleExpanded">
      <div class="tool-info">
        <el-icon class="tool-icon">
          <Tools />
        </el-icon>
        
        <div class="tool-details">
          <span class="tool-name">{{ toolCall.tool_name }}</span>
          <el-tag 
            :type="getStatusType(toolCall.status)"
            size="small"
            class="tool-status"
          >
            {{ getStatusText(toolCall.status) }}
          </el-tag>
        </div>
      </div>
      
      <div class="tool-actions">
        <el-button 
          size="small"
          circle
          @click.stop="copyToolCall"
        >
          <el-icon><DocumentCopy /></el-icon>
        </el-button>
        
        <el-button 
          size="small"
          circle
          @click.stop="toggleExpanded"
        >
          <el-icon>
            <ArrowDown v-if="!expanded" />
            <ArrowUp v-else />
          </el-icon>
        </el-button>
      </div>
    </div>
    
    <!-- 展开的详细信息 -->
    <transition name="tool-expand">
      <div v-if="expanded || showDetails" class="tool-call-body">
        <!-- 参数 -->
        <div v-if="toolCall.parameters" class="tool-section">
          <div class="section-title">{{ $t('tool.parameters') }}</div>
          <div class="parameters-content">
            <pre class="json-content">{{ formatJson(toolCall.parameters) }}</pre>
          </div>
        </div>
        
        <!-- 结果 -->
        <div v-if="toolCall.result" class="tool-section">
          <div class="section-title">{{ $t('tool.result') }}</div>
          <div class="result-content">
            <pre v-if="isJsonResult" class="json-content">{{ formatJson(toolCall.result) }}</pre>
            <div v-else class="text-result">{{ toolCall.result }}</div>
          </div>
        </div>
        
        <!-- 错误信息 -->
        <div v-if="toolCall.error" class="tool-section error-section">
          <div class="section-title">{{ $t('common.error') }}</div>
          <div class="error-content">
            {{ toolCall.error }}
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  Tools, 
  DocumentCopy, 
  ArrowDown, 
  ArrowUp 
} from '@element-plus/icons-vue'
import type { ToolCallEvent } from '@/types/notify'

interface Props {
  toolCall: ToolCallEvent
  showDetails?: boolean
}

const props = defineProps<Props>()

// 响应式状态
const expanded = ref(false)

// 计算属性
const isJsonResult = computed(() => {
  if (!props.toolCall.result) return false
  
  try {
    if (typeof props.toolCall.result === 'object') return true
    JSON.parse(props.toolCall.result)
    return true
  } catch {
    return false
  }
})

// 方法
const getStatusType = (status?: string) => {
  switch (status) {
    case 'success': return 'success'
    case 'error': return 'danger'
    case 'calling': return 'warning'
    default: return 'info'
  }
}

const getStatusText = (status?: string) => {
  switch (status) {
    case 'success': return '成功'
    case 'error': return '失败'
    case 'calling': return '调用中'
    default: return '未知'
  }
}

const formatJson = (data: any) => {
  try {
    if (typeof data === 'string') {
      return JSON.stringify(JSON.parse(data), null, 2)
    }
    return JSON.stringify(data, null, 2)
  } catch {
    return String(data)
  }
}

const toggleExpanded = () => {
  expanded.value = !expanded.value
}

const copyToolCall = async () => {
  const toolCallData = {
    tool_name: props.toolCall.tool_name,
    parameters: props.toolCall.parameters,
    result: props.toolCall.result,
    status: props.toolCall.status,
    error: props.toolCall.error
  }
  
  try {
    await navigator.clipboard.writeText(JSON.stringify(toolCallData, null, 2))
    ElMessage.success('工具调用信息已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}
</script>

<style scoped>
.tool-call-display {
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-base);
  background-color: var(--bg-color);
  overflow: hidden;
}

.tool-call-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background-color: var(--bg-color-secondary);
  cursor: pointer;
  transition: background-color var(--transition-duration) ease;
}

.tool-call-header:hover {
  background-color: var(--border-color);
}

.tool-info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.tool-icon {
  font-size: 18px;
  color: var(--primary-color);
}

.tool-details {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.tool-name {
  font-weight: 600;
  color: var(--text-color-primary);
  font-size: 14px;
}

.tool-status {
  align-self: flex-start;
}

.tool-actions {
  display: flex;
  gap: 8px;
}

.tool-call-body {
  padding: 16px;
  border-top: 1px solid var(--border-color);
}

.tool-section {
  margin-bottom: 16px;
}

.tool-section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-weight: 600;
  color: var(--text-color-primary);
  margin-bottom: 8px;
  font-size: 14px;
}

.parameters-content,
.result-content {
  background-color: var(--bg-color-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-base);
  padding: 12px;
  overflow-x: auto;
}

.json-content {
  margin: 0;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.5;
  color: var(--text-color-primary);
  white-space: pre;
}

.text-result {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.5;
  color: var(--text-color-primary);
  white-space: pre-wrap;
  word-break: break-word;
}

.error-section .section-title {
  color: var(--danger-color);
}

.error-content {
  background-color: #fef0f0;
  border: 1px solid #fbc4c4;
  border-radius: var(--border-radius-base);
  padding: 12px;
  color: var(--danger-color);
  font-size: 14px;
}

/* 展开动画 */
.tool-expand-enter-active,
.tool-expand-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}

.tool-expand-enter-from,
.tool-expand-leave-to {
  max-height: 0;
  opacity: 0;
  padding-top: 0;
  padding-bottom: 0;
}

.tool-expand-enter-to,
.tool-expand-leave-from {
  max-height: 1000px;
  opacity: 1;
}
</style>
