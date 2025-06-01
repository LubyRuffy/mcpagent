<template>
  <el-dialog
    v-model="dialogVisible"
    :title="isEditing ? '编辑系统提示词' : '创建系统提示词'"
    width="600px"
    @closed="resetForm"
    :append-to-body="true"
    :z-index="9999"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="120px"
      label-position="left"
    >
      <el-form-item label="名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入系统提示词名称" />
      </el-form-item>

      <el-form-item label="描述" prop="description">
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="2"
          placeholder="请输入系统提示词描述（可选）"
          class="description-textarea"
        />
      </el-form-item>

      <el-form-item label="提示词内容" prop="content">
        <el-input
          v-model="form.content"
          type="textarea"
          :rows="8"
          placeholder="请输入系统提示词内容"
          class="code-textarea"
        />
      </el-form-item>

      <el-form-item label="占位符">
        <div class="placeholders-container">
          <div class="placeholders-tags">
            <el-tag
              v-for="placeholder in detectedPlaceholders"
              :key="placeholder"
              class="placeholder-tag"
              type="info"
              size="small"
            >
              {{ placeholder }}
            </el-tag>
            <el-empty
              v-if="detectedPlaceholders.length === 0"
              description="暂无检测到占位符"
              :image-size="50"
            />
          </div>
          <div class="placeholder-help">
            <small>系统会自动检测提示词中使用的占位符，格式为 {placeholder}</small>
          </div>
        </div>
      </el-form-item>

      <el-form-item label="设为默认">
        <el-switch v-model="form.is_default" />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" @click="submitForm" :loading="submitting">
        {{ isEditing ? '保存' : '创建' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useConfigStore } from '@/stores/config'
import type { SystemPromptModel, CreateSystemPromptForm } from '@/types/config'

const props = defineProps<{
  visible: boolean
  editPrompt: SystemPromptModel | null
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  'success': []
}>()

const configStore = useConfigStore()
const formRef = ref()
const submitting = ref(false)

// 表单数据
const form = ref<CreateSystemPromptForm>({
  name: '',
  description: '',
  content: '',
  placeholders: [],
  is_default: false
})

// 计算属性
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value)
})

const isEditing = computed(() => !!props.editPrompt)

// 检测提示词中的占位符
const detectedPlaceholders = computed(() => {
  const content = form.value.content
  if (!content) return []

  const matches = content.match(/{([^}]+)}/g) || []
  return matches.map(match => match.slice(1, -1))
})

// 方法
const resetForm = () => {
  // 确保彻底重置表单
  form.value = {
    name: '',
    description: '',
    content: '',
    placeholders: [],
    is_default: false
  }
  
  if (formRef.value) {
    formRef.value.resetFields()
  }
  
  console.log('表单已重置')
}

const submitForm = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    
    submitting.value = true
    
    if (isEditing.value && props.editPrompt) {
      // 更新现有配置
      await configStore.updateSystemPromptById(props.editPrompt.id, form.value)
      ElMessage.success('系统提示词更新成功')
    } else {
      // 创建新配置
      await configStore.createSystemPrompt(form.value)
      ElMessage.success('系统提示词创建成功')
    }
    
    emit('success')
    dialogVisible.value = false
  } catch (error) {
    console.error('表单提交失败:', error)
    ElMessage.error('操作失败，请检查输入并重试')
  } finally {
    submitting.value = false
  }
}

// 监听占位符变化，自动更新表单
watch(detectedPlaceholders, (newPlaceholders) => {
  form.value.placeholders = newPlaceholders
})

// 监听编辑状态变化
watch(() => props.editPrompt, (newPrompt) => {
  if (newPrompt) {
    // 确保编辑时表单内容不为空
    form.value = {
      name: newPrompt.name || '',
      description: newPrompt.description || '',
      content: newPrompt.content || '',
      placeholders: Array.isArray(newPrompt.placeholders) ? [...newPrompt.placeholders] : [],
      is_default: newPrompt.is_default || false
    }
    console.log('编辑模式：已加载提示词配置', newPrompt.name, form.value)
  } else {
    resetForm()
  }
}, { immediate: true, deep: true })

// 监听对话框可见状态变化
watch(() => dialogVisible.value, (visible) => {
  if (visible && props.editPrompt) {
    // 确保对话框打开时正确加载编辑配置
    form.value = {
      name: props.editPrompt.name || '',
      description: props.editPrompt.description || '',
      content: props.editPrompt.content || '',
      placeholders: Array.isArray(props.editPrompt.placeholders) ? [...props.editPrompt.placeholders] : [],
      is_default: props.editPrompt.is_default || false
    }
    console.log('对话框打开：已加载提示词配置', props.editPrompt.name)
  }
})

// 表单验证规则
const rules = {
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  content: [
    { required: true, message: '请输入系统提示词内容', trigger: 'blur' }
  ]
}
</script>

<style scoped>
.dialog-form {
  margin-top: 6px;
}

:deep(.el-form-item) {
  margin-bottom: 12px;
}

:deep(.el-form-item__label) {
  font-size: 13px;
  padding-bottom: 4px;
}

:deep(.el-input__wrapper) {
  padding: 0 8px;
}

:deep(.el-input__inner) {
  height: 32px;
  font-size: 13px;
}

/* 强制设置描述文本域高度 */
:deep(.description-textarea .el-textarea__inner) {
  min-height: 32px !important;
  resize: none !important;
}

:deep(.el-textarea__inner) {
  font-size: 13px;
  padding: 6px 8px;
}

:deep(.el-form-item[prop="content"] .el-textarea__inner) {
  min-height: 120px !important;
}

.placeholder-section {
  margin-top: 6px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
}

.section-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-color-primary);
}

.placeholders-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 6px;
}

.placeholder-tag {
  display: flex;
  align-items: center;
  padding: 2px 6px;
  background-color: var(--el-color-primary-light-9);
  border-radius: 4px;
  font-size: 12px;
  color: var(--el-color-primary);
}

.placeholder-tag .close-icon {
  margin-left: 4px;
  cursor: pointer;
  font-size: 10px;
}

.add-placeholder {
  display: flex;
  align-items: center;
  margin-top: 6px;
}

.add-placeholder-input {
  flex: 1;
  margin-right: 6px;
}

:deep(.el-dialog__body) {
  padding: 10px 15px;
}

:deep(.el-dialog__header) {
  padding: 10px 15px;
  margin-right: 0;
}

:deep(.el-dialog__footer) {
  padding: 10px 15px;
}

@media (max-width: 500px) {
  :deep(.el-form-item__label) {
    font-size: 12px;
  }
  
  :deep(.el-input__inner) {
    height: 30px;
    font-size: 12px;
  }
  
  :deep(.el-dialog__body) {
    padding: 8px 12px;
  }
  
  :deep(.el-dialog__header) {
    padding: 8px 12px;
  }
  
  :deep(.el-dialog__footer) {
    padding: 8px 12px;
  }
}

.placeholders-container {
  display: flex;
  flex-direction: column;
  width: 100%;
}

.placeholders-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
  min-height: 30px;
}

.placeholder-tag {
  margin: 0 !important;
  padding: 4px 8px !important;
  font-size: 12px !important;
  height: auto !important;
  line-height: 1.2 !important;
}

.placeholder-help {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.4;
  padding: 4px 0;
  border-top: 1px dashed var(--el-border-color-lighter);
  margin-top: 4px;
}

:deep(.el-empty) {
  margin: 0 !important;
  padding: 6px 0 !important;
}

:deep(.el-empty__image) {
  width: 40px !important;
  height: 40px !important;
}

:deep(.el-empty__description) {
  margin-top: 4px !important;
  font-size: 12px !important;
}
</style> 