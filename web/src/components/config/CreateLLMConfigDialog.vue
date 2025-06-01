<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="$emit('update:visible', $event)"
    :title="isEdit ? '编辑LLM配置' : '新建LLM配置'"
    width="600px"
    :before-close="handleClose"
    :z-index="9999"
    :append-to-body="true"
    :lock-scroll="true"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
      size="default"
    >
      <!-- 配置名称 -->
      <el-form-item label="配置名称" prop="name">
        <el-input
          v-model="form.name"
          placeholder="请输入配置名称"
          maxlength="50"
          show-word-limit
        />
      </el-form-item>

      <!-- 配置描述 -->
      <el-form-item label="配置描述" prop="description">
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="2"
          placeholder="请输入配置描述（可选）"
          maxlength="200"
          show-word-limit
        />
      </el-form-item>

      <!-- LLM类型 -->
      <el-form-item label="LLM类型" prop="type">
        <el-select
          v-model="form.type"
          @change="handleTypeChange"
          style="width: 100%"
        >
          <el-option label="OpenAI" value="openai" />
          <el-option label="Ollama" value="ollama" />
        </el-select>
      </el-form-item>

      <!-- Base URL -->
      <el-form-item label="Base URL" prop="base_url">
        <el-input
          v-model="form.base_url"
          placeholder="请输入API基础URL"
        />
      </el-form-item>

      <!-- 模型名称 -->
      <el-form-item label="模型名称" prop="model">
        <el-input
          v-model="form.model"
          placeholder="请输入模型名称"
        />
      </el-form-item>

      <!-- API Key -->
      <el-form-item label="API Key" prop="api_key">
        <el-input
          v-model="form.api_key"
          type="password"
          show-password
          placeholder="请输入API Key"
        />
      </el-form-item>

      <!-- 高级参数 -->
      <el-collapse v-model="advancedOpen">
        <el-collapse-item title="高级参数" name="advanced">
          <!-- Temperature -->
          <el-form-item label="Temperature">
            <el-slider
              v-model="form.temperature"
              :min="0"
              :max="2"
              :step="0.1"
              show-input
              :input-size="'small'"
            />
          </el-form-item>

          <!-- Max Tokens -->
          <el-form-item label="Max Tokens">
            <el-input-number
              v-model="form.max_tokens"
              :min="1"
              :max="32000"
              :step="100"
              style="width: 100%"
            />
          </el-form-item>
        </el-collapse-item>
      </el-collapse>

      <!-- 设为默认 -->
      <el-form-item>
        <el-checkbox v-model="form.is_default">
          设为默认配置
        </el-checkbox>
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button
          type="primary"
          :loading="loading"
          @click="handleSubmit"
        >
          {{ isEdit ? '更新' : '创建' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import type { CreateLLMConfigForm, LLMConfigModel } from '@/types/config'

interface Props {
  visible: boolean
  editConfig?: LLMConfigModel | null
}

interface Emits {
  (e: 'update:visible', value: boolean): void
  (e: 'success', config: LLMConfigModel): void
}

const props = withDefaults(defineProps<Props>(), {
  editConfig: null
})

const emit = defineEmits<Emits>()

// 响应式状态
const formRef = ref<FormInstance>()
const loading = ref(false)
const advancedOpen = ref<string[]>([])

// 表单数据
const form = reactive<CreateLLMConfigForm>({
  name: '',
  description: '',
  type: 'ollama',
  base_url: 'http://127.0.0.1:11434',
  model: 'qwen3:14b',
  api_key: 'ollama',
  temperature: 0.7,
  max_tokens: 4000,
  is_default: false
})

// 计算属性
const isEdit = computed(() => !!props.editConfig)

// 表单验证规则
const rules: FormRules = {
  name: [
    { required: true, message: '请输入配置名称', trigger: 'blur' },
    { min: 1, max: 50, message: '配置名称长度在 1 到 50 个字符', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择LLM类型', trigger: 'change' }
  ],
  base_url: [
    { required: true, message: '请输入Base URL', trigger: 'blur' },
    { type: 'url', message: '请输入有效的URL', trigger: 'blur' }
  ],
  model: [
    { required: true, message: '请输入模型名称', trigger: 'blur' }
  ],
  api_key: [
    { required: true, message: '请输入API Key', trigger: 'blur' }
  ]
}

// 方法
const handleTypeChange = (type: string) => {
  if (type === 'ollama') {
    form.base_url = 'http://127.0.0.1:11434'
    form.model = 'qwen3:14b'
    form.api_key = 'ollama'
  } else if (type === 'openai') {
    form.base_url = 'https://api.openai.com/v1'
    form.model = 'gpt-4'
    form.api_key = ''
  }
}

const resetForm = () => {
  Object.assign(form, {
    name: '',
    description: '',
    type: 'ollama',
    base_url: 'http://127.0.0.1:11434',
    model: 'qwen3:14b',
    api_key: 'ollama',
    temperature: 0.7,
    max_tokens: 4000,
    is_default: false
  })
  formRef.value?.clearValidate()
}

const loadEditData = () => {
  if (props.editConfig) {
    Object.assign(form, {
      name: props.editConfig.name,
      description: props.editConfig.description,
      type: props.editConfig.type,
      base_url: props.editConfig.base_url,
      model: props.editConfig.model,
      api_key: props.editConfig.api_key,
      temperature: props.editConfig.temperature || 0.7,
      max_tokens: props.editConfig.max_tokens || 4000,
      is_default: props.editConfig.is_default
    })
  }
}

const handleClose = () => {
  emit('update:visible', false)
  resetForm()
}

const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    loading.value = true

    // 这里应该调用API创建或更新配置
    // 由于我们在组件中不直接调用store，所以通过事件传递数据
    emit('success', form as any)

    handleClose()
  } catch (error) {
    console.error('表单验证失败:', error)
  } finally {
    loading.value = false
  }
}

// 监听编辑配置变化
watch(() => props.editConfig, () => {
  if (props.visible && props.editConfig) {
    loadEditData()
  }
}, { immediate: true })

// 监听对话框显示状态
watch(() => props.visible, (newVal) => {
  if (newVal) {
    if (props.editConfig) {
      loadEditData()
    } else {
      resetForm()
    }
  }
})
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

:deep(.el-select) {
  width: 100%;
}

:deep(.el-radio) {
  margin-right: 12px;
  height: 28px;
  line-height: 28px;
}

:deep(.el-radio__label) {
  font-size: 13px;
}

:deep(.el-radio__input) {
  margin-right: 4px;
}

.advanced-toggle {
  margin: 6px 0;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  color: var(--el-color-primary);
}

.advanced-toggle-icon {
  margin-left: 4px;
  transition: transform 0.2s;
}

.advanced-toggle-icon.expanded {
  transform: rotate(180deg);
}

.advanced-section {
  margin-top: 6px;
  padding: 6px;
  background-color: var(--el-fill-color-lighter);
  border-radius: 4px;
  border: 1px solid var(--el-border-color-light);
}

.slider-container {
  display: flex;
  align-items: center;
}

.slider-value {
  width: 40px;
  text-align: right;
  margin-left: 6px;
  font-size: 13px;
  color: var(--text-color-secondary);
}

:deep(.el-slider) {
  flex: 1;
  margin: 0;
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
</style>
