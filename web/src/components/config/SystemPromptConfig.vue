<template>
  <el-card class="config-card">
    <template #header>
      <div class="card-header">
        <span>{{ $t('config.prompt.title') }}</span>
      </div>
    </template>

    <!-- 系统提示词配置选择器 -->
    <ConfigSelector
      v-model:modelValue="selectedConfigId"
      label=""
      :placeholder="$t('config.prompt.selectConfig')"
      :configs="configStore.systemPrompts"
      :templates="[]"
      :option-groups="false"
      :saved-configs-label="$t('config.prompt.savedConfigs')"
      :templates-label="$t('config.prompt.presetTemplates')"
      :default-tag-text="$t('common.default')"
      :edit-btn-text="$t('common.edit')"
      :delete-btn-text="$t('common.delete')"
      :show-create-option="true"
      :create-option-text="$t('config.prompt.createNew')"
      @change="handleConfigSelect"
      @edit="handleEditConfigById"
      @delete="handleDeleteConfig"
      @create="handleCreateNew"
      :key="refreshKey"
      class="config-select"
    >
      <template #actions>
        <el-tooltip :content="showDetails ? $t('config.prompt.collapseDetails') : $t('config.prompt.expandDetails')" placement="top">
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
          <el-descriptions-item :label="$t('config.prompt.configName')">
            {{ currentSelectedConfig.name }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('common.description')">
            {{ currentSelectedConfig.description || $t('common.noDescription') }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('config.prompt.placeholders')">
            <el-tag 
              v-for="placeholder in currentSelectedConfig.placeholders" 
              :key="placeholder"
              size="small"
              class="placeholder-tag"
            >
              {{ placeholder }}
            </el-tag>
            <span v-if="!currentSelectedConfig.placeholders || currentSelectedConfig.placeholders.length === 0">{{ $t('config.prompt.noPlaceholders') }}</span>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </div>

    <!-- 添加当前提示词内容显示区域 -->
    <div class="prompt-content-section" v-if="currentSelectedConfig || isTemplateSelected">
      <div class="section-header">
        <span class="section-title">{{ $t('config.prompt.content') }}</span>
      </div>
      <el-input
        v-model="systemPrompt"
        type="textarea"
        :rows="5"
        :placeholder="$t('config.prompt.contentPlaceholder')"
        class="prompt-textarea"
      />
    </div>

    <!-- 占位符配置 -->
    <div class="placeholders-section" v-if="showPlaceholders">
      <div class="section-header">
        <span class="section-title">{{ $t('config.prompt.placeholders') }}</span>
      </div>

      <div class="placeholders-list">
        <div
          v-for="(value, key) in placeholders"
          :key="key"
          class="placeholder-item"
        >
          <div class="placeholder-key">
            <el-tag size="small">{{ '{' + key + '}' }}</el-tag>
          </div>

          <div class="placeholder-value">
            <el-input
              v-model="placeholders[key]"
              size="small"
              @input="updatePlaceholder(key, $event)"
            />
          </div>
        </div>

        <el-empty
          v-if="Object.keys(placeholders).length === 0"
          :description="$t('config.prompt.noPlaceholders')"
          :image-size="60"
        />
      </div>
    </div>

    <!-- 预览区域 -->
    <div class="preview-section">
      <div class="section-header">
        <span class="section-title">{{ $t('config.prompt.preview') }}</span>
        <el-button
          size="small"
          @click="showPreview = !showPreview"
        >
          {{ showPreview ? $t('config.prompt.hidePreview') : $t('config.prompt.showPreview') }}
        </el-button>
      </div>
      <div v-if="showPreview" class="preview-content">
        <pre class="preview-text">{{ previewText }}</pre>
      </div>
    </div>

    <!-- 无配置时的提示 -->
    <div v-if="!currentSelectedConfig && !isTemplateSelected" class="no-config-tip">
      <el-empty :description="$t('config.prompt.selectOrCreate')" :image-size="60">
        <el-button type="primary" size="small" @click="showCreateDialog = true">
          + {{ $t('config.prompt.createConfig') }}
        </el-button>
      </el-empty>
    </div>
  </el-card>

  <!-- 新建/编辑配置对话框 -->
  <CreateSystemPromptDialog
    v-model:visible="showCreateDialog"
    :edit-prompt="editingConfig"
    @success="handleConfigSuccess"
  />
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useConfigStore } from '@/stores/config'
import { useI18n } from 'vue-i18n'
import CreateSystemPromptDialog from './CreateSystemPromptDialog.vue'
import ConfigSelector from '@/components/common/ConfigSelector.vue'
import type { SystemPromptModel, CreateSystemPromptForm } from '@/types/config'

const configStore = useConfigStore()
const { t } = useI18n()

// 响应式状态
const selectedConfigId = ref<number | string>('')
const showCreateDialog = ref(false)
const editingConfig = ref<SystemPromptModel | null>(null)
const showDetails = ref(false)
const showPreview = ref(false)
const showPlaceholders = ref(true)
const refreshKey = ref(0)

// 计算属性
const systemPrompt = computed({
  get: () => configStore.config.system_prompt,
  set: (value: string) => configStore.updateSystemPrompt(value)
})

const placeholders = computed({
  get: () => configStore.config.placeholders || {},
  set: (value: Record<string, any>) => configStore.updatePlaceholders(value)
})

const currentSelectedConfig = computed(() => {
  if (!selectedConfigId.value || typeof selectedConfigId.value === 'string' && selectedConfigId.value.startsWith('template-')) {
    return null
  }
  
  const id = typeof selectedConfigId.value === 'string' ? parseInt(selectedConfigId.value) : selectedConfigId.value
  if (!id || isNaN(id)) return null
  
  return configStore.systemPrompts.find(config => config.id === id) || null
})

const isTemplateSelected = computed(() => {
  // 由于模板已移除，现在不再有带template前缀的选项
  return false;
})

const selectedTemplateName = computed(() => {
  return '';
})

const previewText = computed(() => {
  let text = systemPrompt.value

  // 替换占位符
  Object.entries(placeholders.value).forEach(([key, value]) => {
    const placeholder = '{' + key + '}'
    text = text.replace(new RegExp(placeholder, 'g'), String(value))
  })

  // 替换特殊占位符
  text = text.replace(/{date}/g, new Date().toLocaleDateString('zh-CN'))

  return text
})

// 初始化
onMounted(() => {
  // 如果有默认的系统提示词配置，选中它
  if (configStore.defaultSystemPrompt) {
    selectedConfigId.value = configStore.defaultSystemPrompt.id
  } else if (configStore.systemPrompts.length > 0) {
    // 否则选中第一个
    selectedConfigId.value = configStore.systemPrompts[0].id
  }
})

// 监听当前系统提示词ID变化
watch(() => configStore.currentSystemPromptId, (newId) => {
  if (newId) {
    selectedConfigId.value = newId
  }
}, { immediate: true })

// 监听选项变化
watch(() => selectedConfigId.value, (newValue) => {
  if (!newValue) return
  
  if (typeof newValue === 'string' && newValue.startsWith('template-')) {
    // 模板已移除，这部分逻辑不再需要
    return;
  }
}, { immediate: true })

// 方法
const handleConfigSelect = (configId: number | string) => {
  if (!configId) return
  
  if (typeof configId === 'string' && configId.startsWith('template-')) {
    // 模板已移除，这部分逻辑不再需要
    return;
  }
  
  const id = typeof configId === 'number' ? configId : parseInt(configId as string)
  if (!isNaN(id)) {
    configStore.selectSystemPrompt(id)
    
    // 自动显示预览
    showPreview.value = true
  }
}

const handleCreateNew = () => {
  editingConfig.value = null
  showCreateDialog.value = true
  // 保持选中的配置ID不变
}

const handleEditConfigById = (id: number) => {
  const config = configStore.systemPrompts.find(c => c.id === id)
  if (config) {
    // 深拷贝配置对象，防止引用问题
    editingConfig.value = JSON.parse(JSON.stringify(config))
    console.log('准备编辑配置:', editingConfig.value)
    showCreateDialog.value = true
  } else {
    console.error('未找到要编辑的配置:', id)
  }
}

const handleDeleteConfig = (id: number, name: string) => {
  ElMessageBox.confirm(
    t('config.prompt.deleteConfirmText', { name }),
    t('config.prompt.deleteConfirmTitle'),
    {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }
  )
    .then(async () => {
      try {
        await configStore.deleteSystemPrompt(id)
        ElMessage.success(t('config.prompt.deleteSuccess'))
        
        // 如果删除的是当前选中的配置，重置选择
        if (selectedConfigId.value === id) {
          selectedConfigId.value = ''
        }
        
        // 刷新选择器
        refreshKey.value++
      } catch (error) {
        console.error('删除失败:', error)
        ElMessage.error(t('config.prompt.deleteFailed'))
      }
    })
    .catch(() => {
      // 用户取消删除
    })
}

const handleConfigSuccess = () => {
  // 刷新选择器
  refreshKey.value++
  
  // 选中新创建/编辑的配置
  if (editingConfig.value) {
    selectedConfigId.value = editingConfig.value.id
  } else if (configStore.systemPrompts.length > 0) {
    // 选中最新创建的配置
    const latestConfig = configStore.systemPrompts[configStore.systemPrompts.length - 1]
    selectedConfigId.value = latestConfig.id
  }
}

const updatePlaceholder = (key: string, value: any) => {
  const newPlaceholders = { ...placeholders.value }
  newPlaceholders[key] = value
  placeholders.value = newPlaceholders
}

const saveCurrentAsNew = () => {
  if (!systemPrompt.value) {
    ElMessage.warning(t('config.prompt.contentRequired'))
    return
  }
  
  const newPrompt: CreateSystemPromptForm = {
    name: selectedTemplateName.value ? `${selectedTemplateName.value} (${t('config.prompt.custom')})` : t('config.prompt.newConfig'),
    description: t('config.prompt.templateCreatedDesc'),
    content: systemPrompt.value,
    placeholders: Object.keys(placeholders.value),
    is_default: false
  }
  
  // 预填充表单
  editingConfig.value = null
  showCreateDialog.value = true
  
  // 延迟执行，确保对话框已完全打开
  setTimeout(() => {
    // 这里需要找到一种方式将新的表单数据传递给CreateSystemPromptDialog组件
    ElMessage.info(t('config.prompt.fillNameAndSave'))
  }, 300)
}
</script>

<style>
/* 确保SystemPromptConfig.vue中的下拉菜单按钮正确显示 */
.config-select .el-select-dropdown__item .config-option {
  display: flex !important;
  justify-content: space-between !important;
  width: 100% !important;
  padding-right: 10px !important;
}

/* 确保下拉菜单的z-index低于弹窗 */
:deep(.el-select__popper) {
  z-index: 2000 !important;
}

.config-select .el-select-dropdown__item .config-option .option-buttons {
  display: flex !important;
  visibility: visible !important;
  opacity: 1 !important;
  gap: 4px !important;
}

.config-select .el-select-dropdown__item .option-buttons .el-button {
  visibility: visible !important;
  opacity: 1 !important;
  display: inline-flex !important;
  align-items: center !important;
  justify-content: center !important;
  width: 24px !important;
  height: 24px !important;
  padding: 0 !important;
  min-height: auto !important;
  border-radius: 50% !important;
}

.config-select .el-select-dropdown__item .option-buttons .edit-btn {
  background-color: #409EFF !important;
  border-color: #409EFF !important;
  color: white !important;
}

.config-select .el-select-dropdown__item .option-buttons .delete-btn {
  background-color: #F56C6C !important;
  border-color: #F56C6C !important;
  color: white !important;
}
</style>

<style scoped>
.config-card {
  background-color: var(--bg-color);
  border-radius: 6px;
  overflow: hidden;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.icon-btn {
  width: 24px;
  height: 24px;
  padding: 0;
  font-size: 11px;
}

.prompt-content-section {
  margin-top: 6px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}

.section-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-color-primary);
}

.prompt-textarea {
  width: 100%;
}

/* 使用类选择器确保样式一致性 */
:deep(.prompt-textarea .el-textarea__inner) {
  min-height: 120px !important;
  resize: vertical;
}

/* 占位符相关样式 */
.placeholders-section {
  margin-top: 6px;
}

.placeholders-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.placeholder-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.placeholder-key {
  width: 80px;
  flex-shrink: 0;
}

.placeholder-value {
  flex: 1;
}

.placeholder-tag {
  margin-right: 4px;
  margin-bottom: 4px;
}

/* 预览区域 */
.preview-section {
  margin-top: 6px;
}

.preview-content {
  margin-top: 4px;
  padding: 6px;
  background-color: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-light);
  border-radius: 4px;
  max-height: 150px;
  overflow-y: auto;
}

.preview-text {
  margin: 0;
  white-space: pre-wrap;
  font-size: 12px;
  color: var(--text-color-regular);
  overflow-wrap: break-word;
}

/* 配置详情容器样式 */
.config-details-container {
  margin-top: 6px;
  padding: 6px;
  background-color: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-light);
  border-radius: 4px;
  transition: all 0.3s ease;
  max-height: 150px;
  overflow-y: auto;
}

.config-details {
  padding: 0;
}

/* 无配置时的提示 */
.no-config-tip {
  margin-top: 6px;
  text-align: center;
}

.no-config-tip :deep(.el-empty) {
  padding: 6px 0;
}

.no-config-tip :deep(.el-empty__image) {
  width: 40px !important;
  height: 40px !important;
}

.no-config-tip :deep(.el-empty__description) {
  font-size: 12px;
  margin-top: 4px;
}

.no-config-tip :deep(.el-button) {
  height: 24px;
  font-size: 11px;
}

@media (max-width: 500px) {
  .section-title {
    font-size: 12px;
  }
  
  .preview-content {
    max-height: 100px;
  }
  
  .config-details-container {
    max-height: 120px;
  }
}
</style>
