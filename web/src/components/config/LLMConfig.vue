<template>
  <el-card class="config-card">
    <template #header>
      <div class="card-header">
        <span>{{ $t('config.llm.title') }}</span>
      </div>
    </template>

    <!-- LLM配置选择器 -->
    <ConfigSelector
      v-model:modelValue="selectedConfigId"
      label=""
      placeholder="选择LLM配置"
      :configs="configStore.llmConfigs"
      :default-tag-text="'默认'"
      :edit-btn-text="'编辑'"
      :delete-btn-text="'删除'"
      :show-create-option="true"
      :create-option-text="'新建配置'"
      @change="handleConfigSelect"
      @edit="handleEditConfigById"
      @delete="handleDeleteConfig"
      @create="handleCreateNew"
      :key="refreshKey"
      class="config-select"
    >
      <template #actions>
        <el-tooltip :content="showDetails ? '收起详情' : '展开详情'" placement="top">
          <el-button
            v-if="currentSelectedConfig"
            size="small"
            circle
            @click="showDetails = !showDetails"
            class="icon-btn"
          >
            {{ showDetails ? '▲' : '▼' }}
          </el-button>
        </el-tooltip>
      </template>
    </ConfigSelector>

    <!-- 可折叠的配置详情 -->
    <div v-if="currentSelectedConfig && showDetails" class="config-details-container">
      <div class="config-details">
        <el-descriptions :column="1" size="small" border>
          <el-descriptions-item label="配置名称">
            {{ currentSelectedConfig.name }}
          </el-descriptions-item>
          <el-descriptions-item label="类型">
            <el-tag :type="currentSelectedConfig.type === 'openai' ? 'success' : 'primary'">
              {{ currentSelectedConfig.type === 'openai' ? 'OpenAI' : 'Ollama' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="模型">
            {{ currentSelectedConfig.model }}
          </el-descriptions-item>
          <el-descriptions-item label="Base URL">
            {{ currentSelectedConfig.base_url }}
          </el-descriptions-item>
        </el-descriptions>

        <!-- 测试连接按钮 -->
        <el-button
          type="primary"
          :loading="testing"
          @click="testConnection"
          class="test-button"
        >
          🔗 测试连接
        </el-button>

        <el-badge
            :value="connectionStatus"
            :type="connectionStatusType"
            class="status-badge"
        />
      </div>
    </div>

    <!-- 无配置时的提示 -->
    <div v-if="!currentSelectedConfig" class="no-config-tip">
      <el-empty description="请选择或创建LLM配置" :image-size="60">
        <el-button type="primary" size="small" @click="showCreateDialog = true">
          + 创建配置
        </el-button>
      </el-empty>
    </div>
  </el-card>

  <!-- 新建/编辑配置对话框 -->
  <CreateLLMConfigDialog
    v-model:visible="showCreateDialog"
    :edit-config="editingConfig"
    @success="handleConfigSuccess"
  />
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, inject } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import { useConfigStore } from '@/stores/config'
import { llmApi } from '@/utils/api'
import type { LLMConfig, LLMConfigModel, CreateLLMConfigForm } from '@/types/config'
import CreateLLMConfigDialog from './CreateLLMConfigDialog.vue'
import ConfigSelector from '@/components/common/ConfigSelector.vue'

const configStore = useConfigStore()
// 注入应用初始化状态
const appInitialized = inject('appInitialized', ref(false))

// 响应式状态
const testing = ref(false)
const advancedOpen = ref<string[]>([])
const connectionStatus = ref('未连接')
const connectionStatusType = ref<'info' | 'success' | 'warning' | 'danger'>('info')
const selectedConfigId = ref<number | string>('')
const showCreateDialog = ref(false)
const editingConfig = ref<LLMConfigModel | null>(null)
const showDetails = ref(false)
const refreshKey = ref(0)

// 计算属性
const llmConfig = computed({
  get: () => configStore.config.llm,
  set: (value: LLMConfig) => configStore.updateLLMConfig(value)
})

const currentSelectedConfig = computed(() => {
  const id = typeof selectedConfigId.value === 'string' ? parseInt(selectedConfigId.value) : selectedConfigId.value
  if (!id || isNaN(id)) return null
  return configStore.llmConfigs.find(config => config.id === id) || null
})

const availableModels = computed(() => {
  const models = configStore.availableModels
  if (llmConfig.value.type === 'ollama') {
    return ['qwen3:4b', 'qwen3:14b', 'llama3:8b', 'llama3:70b', ...models]
  } else {
    return ['gpt-4', 'gpt-4-turbo', 'gpt-3.5-turbo', ...models]
  }
})

// 方法
const handleTypeChange = (type: string) => {
  // 根据类型设置默认值
  if (type === 'ollama') {
    llmConfig.value = {
      ...llmConfig.value,
      base_url: 'http://127.0.0.1:11434',
      model: 'qwen3:14b',
      api_key: 'ollama'
    }
  } else if (type === 'openai') {
    llmConfig.value = {
      ...llmConfig.value,
      base_url: 'https://api.openai.com/v1',
      model: 'gpt-4',
      api_key: ''
    }
  }
}

const validateUrl = () => {
  try {
    new URL(llmConfig.value.base_url)
    return true
  } catch {
    ElMessage.warning('URL格式不正确')
    return false
  }
}

const testConnection = async () => {
  if (!currentSelectedConfig.value) {
    ElMessage.warning('请先选择一个LLM配置')
    return
  }

  testing.value = true
  connectionStatus.value = '连接中...'
  connectionStatusType.value = 'warning'

  try {
    // 构建测试配置
    const testConfig: LLMConfig = {
      type: currentSelectedConfig.value.type,
      base_url: currentSelectedConfig.value.base_url,
      model: currentSelectedConfig.value.model,
      api_key: currentSelectedConfig.value.api_key,
      temperature: currentSelectedConfig.value.temperature,
      max_tokens: currentSelectedConfig.value.max_tokens
    }

    // 调用后端API测试连接
    const response = await llmApi.testConnection(testConfig)

    if (response.success) {
      connectionStatus.value = '连接成功'
      connectionStatusType.value = 'success'
      ElMessage.success(response.message || 'LLM连接测试成功')
    } else {
      connectionStatus.value = '连接失败'
      connectionStatusType.value = 'danger'
      ElMessage.error(response.message || 'LLM连接测试失败')
    }

  } catch (error) {
    connectionStatus.value = '连接失败'
    connectionStatusType.value = 'danger'
    const errorMessage = error instanceof Error ? error.message : 'LLM连接测试失败'
    ElMessage.error(errorMessage)
  } finally {
    testing.value = false
  }
}

// 配置管理方法
const handleConfigSelect = (configId: number | string) => {
  // 如果选择的是"新建配置"选项，不做任何处理
  if (configId === 'create-new') {
    return
  }

  const id = typeof configId === 'string' ? parseInt(configId) : configId
  if (!isNaN(id)) {
    configStore.selectLLMConfig(id)
    connectionStatus.value = '未连接'
    connectionStatusType.value = 'info'
  }
}

const handleEditConfig = () => {
  const id = typeof selectedConfigId.value === 'string' ? parseInt(selectedConfigId.value) : selectedConfigId.value
  if (id && !isNaN(id)) {
    const config = configStore.llmConfigs.find(c => c.id === id)
    if (config) {
      editingConfig.value = config
      showCreateDialog.value = true
    }
  }
}

const handleEditConfigById = (configId: number) => {
  const config = configStore.llmConfigs.find(c => c.id === configId)
  if (config) {
    // 关闭任何可能打开的下拉框
    const selectEl = document.querySelector('.el-select');
    if (selectEl) {
      // @ts-ignore
      const selectInstance = selectEl.__vue__?.exposed;
      if (selectInstance && typeof selectInstance.blur === 'function') {
        selectInstance.blur();
      }
    }
    
    editingConfig.value = config
    // 延迟一帧打开对话框，确保下拉框已关闭
    setTimeout(() => {
      showCreateDialog.value = true
    }, 0)
  }
}

const handleCreateNew = () => {
  // 关闭任何可能打开的下拉框
  const selectEl = document.querySelector('.el-select');
  if (selectEl) {
    // @ts-ignore
    const selectInstance = selectEl.__vue__?.exposed;
    if (selectInstance && typeof selectInstance.blur === 'function') {
      selectInstance.blur();
    }
  }
  
  editingConfig.value = null
  // 延迟一帧打开对话框，确保下拉框已关闭
  setTimeout(() => {
    showCreateDialog.value = true
  }, 0)
}

const handleDeleteConfig = async (configId: number, configName: string) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除配置"${configName}"吗？此操作不可撤销。`,
      '删除确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning',
        confirmButtonClass: 'el-button--danger'
      }
    )

    // 执行删除操作
    await configStore.deleteLLMConfig(configId)

    // 如果删除的是当前选中的配置，清空选择
    if (selectedConfigId.value === configId) {
      selectedConfigId.value = ''
      connectionStatus.value = '未连接'
      connectionStatusType.value = 'info'
    }

    ElMessage.success('配置删除成功')
  } catch (error) {
    if (error !== 'cancel') {
      const errorMessage = error instanceof Error ? error.message : '删除配置失败'
      ElMessage.error(errorMessage)
    }
  }
}

const handleConfigSuccess = async (configData: CreateLLMConfigForm) => {
  try {
    if (editingConfig.value) {
      // 更新配置
      await configStore.updateLLMConfigById(editingConfig.value.id, configData)
      // 强制刷新下拉框
      refreshKey.value++
      
      // 保存当前选中的ID
      const currentId = selectedConfigId.value;
      
      // 重新加载配置列表以确保拿到最新数据
      await configStore.loadLLMConfigs(true);
      
      // 恢复选中状态
      if (currentId) {
        selectedConfigId.value = currentId;
      }
      
      ElMessage.success('配置更新成功')
    } else {
      // 创建配置
      const newConfig = await configStore.createLLMConfig(configData)
      if (newConfig && newConfig.id) {
        selectedConfigId.value = newConfig.id
        configStore.selectLLMConfig(newConfig.id)
      }
      // 强制刷新下拉框
      refreshKey.value++
      ElMessage.success('配置创建成功')
    }

    // 重置编辑状态
    editingConfig.value = null
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : '操作失败'
    ElMessage.error(errorMessage)
  }
}

// 监听llmConfigs变化强制更新
watch(() => configStore.llmConfigs, () => {
  console.log('【LLMConfig组件】检测到llmConfigs变化，强制刷新下拉框')
  refreshKey.value++
}, { deep: true })

// 初始化
onMounted(async () => {
  console.log('【LLMConfig组件】onMounted开始执行', new Date().toISOString())
  console.log('【LLMConfig组件】当前应用初始化状态:', appInitialized.value)
  console.log('【LLMConfig组件】当前配置加载状态:', configStore.configLoaded)
  console.log('【LLMConfig组件】当前llmConfigs长度:', configStore.llmConfigs.length)
  
  // 如果配置已加载，直接使用已加载的配置
  if (configStore.configLoaded) {
    initializeSelectedConfig()
  } else {
    // 监听配置加载状态变化
    console.log('【LLMConfig组件】配置尚未加载完成，等待配置加载')
    watch(() => configStore.configLoaded, (loaded) => {
      if (loaded) {
        console.log('【LLMConfig组件】检测到配置已加载完成，设置当前选中的配置')
        initializeSelectedConfig()
      }
    }, { immediate: true })
  }
  
  console.log('【LLMConfig组件】onMounted执行完成', new Date().toISOString())
})

// 初始化选中的配置
const initializeSelectedConfig = () => {
  // 设置当前选中的配置
  if (configStore.currentLLMConfigId) {
    selectedConfigId.value = configStore.currentLLMConfigId
    console.log('【LLMConfig组件】使用currentLLMConfigId:', configStore.currentLLMConfigId)
  } else if (configStore.defaultLLMConfig) {
    selectedConfigId.value = configStore.defaultLLMConfig.id
    configStore.selectLLMConfig(configStore.defaultLLMConfig.id)
    console.log('【LLMConfig组件】使用defaultLLMConfig:', configStore.defaultLLMConfig.id)
  } else if (configStore.llmConfigs.length > 0) {
    // 如果没有默认配置，选择第一个配置
    const firstConfig = configStore.llmConfigs[0]
    selectedConfigId.value = firstConfig.id
    configStore.selectLLMConfig(firstConfig.id)
    console.log('【LLMConfig组件】使用第一个配置:', firstConfig.id)
  }
}

// 监听配置变化
watch(
  () => llmConfig.value,
  () => {
    connectionStatus.value = '未连接'
    connectionStatusType.value = 'info'
  },
  { deep: true }
)

// 初始化默认值
if (!llmConfig.value.temperature) {
  llmConfig.value.temperature = 0.7
}
if (!llmConfig.value.max_tokens) {
  llmConfig.value.max_tokens = 4000
}
</script>

<style scoped>
.config-card {
  border: 1px solid var(--border-color);
  background-color: var(--bg-color);
  width: 100%; /* 适应容器宽度 */
  min-width: 200px; /* 侧边栏最小宽度 */
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.status-badge :deep(.el-badge__content) {
  font-size: 10px;
  padding: 2px 6px;
  min-width: auto;
  height: auto;
  line-height: 1;
}

:deep(.el-form-item) {
  margin-bottom: 12px;
}

:deep(.el-collapse-item__header) {
  font-size: 14px;
  padding-left: 0;
}

:deep(.el-collapse-item__content) {
  padding-bottom: 0;
}

:deep(.el-slider) {
  margin-right: 12px;
}

.config-selector {
  margin-bottom: 6px;
}

/* 确保表单项的内容区域有足够的宽度 */
:deep(.el-form-item__content) {
  width: 100%;
  overflow: visible;
}

:deep(.el-form-item) {
  margin-bottom: 6px;
}

/* 确保下拉框有正确的宽度 */
.config-select :deep(.el-input) {
  width: 100%;
}

.config-select :deep(.el-input__wrapper) {
  width: 100%;
}

.selector-row {
  display: flex;
  align-items: center;
  gap: 4px;
}

.config-select {
  flex: 1;
  min-width: 120px;
}

.action-buttons {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;
  min-width: 48px; /* 确保按钮区域有最小宽度 */
}

.icon-btn {
  width: 24px;
  height: 24px;
  padding: 0;
  font-size: 11px;
  flex-shrink: 0;
}

.config-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  min-width: 150px;
  padding: 4px 0;
}

.config-info {
  display: flex;
  align-items: center;
  flex: 1;
  gap: 8px;
}

.config-info span {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.edit-btn-in-option {
  width: 20px;
  height: 20px;
  padding: 0;
  font-size: 10px;
  margin-left: 8px;
  opacity: 0.7;
  transition: opacity 0.2s;
}

.edit-btn-in-option:hover {
  opacity: 1;
}

/* 新建配置选项样式 */
.create-new-option {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 8px 0;
  cursor: pointer;
  color: var(--el-color-primary);
  font-weight: 500;
}

.create-text {
  font-size: 13px;
}

/* 下拉框选项悬停效果 */
:deep(.el-select-dropdown__item.create-option) {
  border-top: 1px solid var(--el-border-color-light);
  background-color: var(--el-fill-color-lighter);
}

:deep(.el-select-dropdown__item.create-option:hover) {
  background-color: var(--el-color-primary-light-9);
}

/* 下拉框样式优化 */
:deep(.el-select-dropdown) {
  min-width: 200px !important;
}

:deep(.el-select-dropdown__item) {
  padding: 8px 12px;
  min-height: auto;
}

/* 响应式设计 - 侧边栏优化 */
@media (max-width: 300px) {
  .config-card {
    min-width: 180px;
  }

  .icon-btn {
    width: 20px;
    height: 20px;
    font-size: 10px;
  }

  .edit-btn-in-option {
    width: 18px;
    height: 18px;
    font-size: 9px;
  }

  .selector-row {
    gap: 4px;
  }
}

/* 配置详情容器样式 */
.config-details-container {
  margin-top: 6px;
  padding: 6px;
  background-color: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-light);
  border-radius: 4px;
  transition: all 0.3s ease;
  max-height: 200px;
  overflow-y: auto;
}

.config-details {
  padding: 0;
}

.config-details .el-descriptions {
  margin-bottom: 6px;
}

.config-details :deep(.el-descriptions__label) {
  font-size: 11px;
  width: 60px;
  padding: 4px 6px;
}

.config-details :deep(.el-descriptions__content) {
  font-size: 11px;
  padding: 4px 6px;
}

.test-button {
  width: 100%;
  margin-top: 6px;
  height: 24px;
  font-size: 11px;
  padding: 0 8px;
}

.no-config-tip {
  margin-top: 6px;
  text-align: center;
}

.no-config-tip .el-empty {
  padding: 8px 0;
}

.no-config-tip :deep(.el-empty__image) {
  width: 40px !important;
  height: 40px !important;
}

.no-config-tip :deep(.el-empty__description) {
  font-size: 12px;
  margin-top: 4px;
}

.no-config-tip .el-button {
  height: 24px;
  font-size: 11px;
  width: 100%;
}
</style>
