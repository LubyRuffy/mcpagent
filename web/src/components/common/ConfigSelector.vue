<template>
  <div class="config-selector">
    <el-form-item :label="label">
      <div class="selector-row">
        <el-select
          :model-value="modelValue"
          @update:model-value="$emit('update:modelValue', $event)"
          :placeholder="placeholder"
          @change="$emit('change', $event)"
          class="config-select"
          :key="refreshKey"
          filterable
        >
          <!-- 现有配置选项 -->
          <el-option-group v-if="optionGroups" :label="savedConfigsLabel">
            <el-option
              v-for="config in configs"
              :key="`${config.id}-${refreshKey}`"
              :label="config.name"
              :value="config.id"
            >
              <div class="config-option">
                <div class="config-info">
                  <span>{{ config.name }}</span>
                  <el-tag v-if="config.is_default" size="small" type="success">{{ defaultTagText }}</el-tag>
                </div>
                <div class="option-buttons">
                  <el-button
                    size="small"
                    circle
                    type="primary"
                    @click.stop="$emit('edit', config.id)"
                    class="edit-btn"
                  >
                    <span class="edit-icon">
                      <svg viewBox="0 0 1024 1024" width="12" height="12">
                        <path fill="currentColor" d="M832.5 191.4c-84.3-84.3-221-84.3-305.2 0L95.8 623c-4.2 4.2-7.3 9.3-9.1 14.9l-53.4 221.4c-3.9 16.1 10.2 30.2 26.3 26.3l221.4-53.4c5.6-1.3 10.7-4.5 14.9-9.1l431.5-431.5c84.2-84.3 84.2-221 0-305.2zM759 385.9L629 255.9 303.9 581l130 130L759 385.9zM150.6 875.5l35.7-147.8 112.1 112.1-147.8 35.7z"/>
                      </svg>
                    </span>
                  </el-button>
                  <el-button
                    size="small"
                    circle
                    type="danger"
                    @click.stop="$emit('delete', config.id, config.name)"
                    class="delete-btn"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
              </div>
            </el-option>
          </el-option-group>

          <!-- 常规选项（无分组） -->
          <template v-if="!optionGroups">
            <el-option
              v-for="config in configs"
              :key="`${config.id}-${refreshKey}`"
              :label="config.name"
              :value="config.id"
            >
              <div class="config-option">
                <div class="config-info">
                  <span>{{ config.name }}</span>
                  <el-tag v-if="config.is_default" size="small" type="success">{{ defaultTagText }}</el-tag>
                </div>
                <div class="option-buttons">
                  <el-button
                    size="small"
                    circle
                    type="primary"
                    @click.stop="$emit('edit', config.id)"
                    class="edit-btn"
                  >
                    <span class="edit-icon">
                      <svg viewBox="0 0 1024 1024" width="12" height="12">
                        <path fill="currentColor" d="M832.5 191.4c-84.3-84.3-221-84.3-305.2 0L95.8 623c-4.2 4.2-7.3 9.3-9.1 14.9l-53.4 221.4c-3.9 16.1 10.2 30.2 26.3 26.3l221.4-53.4c5.6-1.3 10.7-4.5 14.9-9.1l431.5-431.5c84.2-84.3 84.2-221 0-305.2zM759 385.9L629 255.9 303.9 581l130 130L759 385.9zM150.6 875.5l35.7-147.8 112.1 112.1-147.8 35.7z"/>
                      </svg>
                    </span>
                  </el-button>
                  <el-button
                    size="small"
                    circle
                    type="danger"
                    @click.stop="$emit('delete', config.id, config.name)"
                    class="delete-btn"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
              </div>
            </el-option>
          </template>

          <!-- 模板选项组 -->
          <el-option-group v-if="optionGroups && templates && templates.length > 0" :label="templatesLabel">
            <el-option
              v-for="template in templates"
              :key="template.id"
              :label="`${template.name}`"
              :value="`template-${template.id}`"
            >
              <div class="config-option">
                <div class="config-info">
                  <span>{{ template.name }}</span>
                </div>
              </div>
            </el-option>
          </el-option-group>

          <!-- 新建配置选项 -->
          <el-option
            v-if="showCreateOption"
            key="create-new"
            value="create-new"
            class="create-option"
          >
            <div class="create-new-option" @click.stop="$emit('create')">
              <span class="create-text">+ {{ createOptionText }}</span>
            </div>
          </el-option>
        </el-select>

        <!-- 详情按钮 -->
        <div class="action-buttons">
          <slot name="actions"></slot>
        </div>
      </div>
    </el-form-item>
  </div>
</template>

<script setup lang="ts">
import { ref, defineProps, defineEmits } from 'vue'
import { Delete } from '@element-plus/icons-vue'

interface ConfigItem {
  id: number | string;
  name: string;
  is_default?: boolean;
  [key: string]: any;
}

interface TemplateItem {
  id: string;
  name: string;
  [key: string]: any;
}

const props = defineProps({
  modelValue: {
    type: [Number, String],
    default: ''
  },
  label: {
    type: String,
    default: ''
  },
  placeholder: {
    type: String,
    default: '请选择配置'
  },
  configs: {
    type: Array as () => ConfigItem[],
    default: () => []
  },
  templates: {
    type: Array as () => TemplateItem[],
    default: () => []
  },
  showCreateOption: {
    type: Boolean,
    default: true
  },
  createOptionText: {
    type: String,
    default: '新建配置'
  },
  editBtnText: {
    type: String,
    default: '编辑'
  },
  deleteBtnText: {
    type: String,
    default: '删除'
  },
  defaultTagText: {
    type: String,
    default: '默认'
  },
  optionGroups: {
    type: Boolean,
    default: false
  },
  savedConfigsLabel: {
    type: String,
    default: '已保存配置'
  },
  templatesLabel: {
    type: String,
    default: '预设模板'
  }
})

defineEmits(['update:modelValue', 'change', 'edit', 'delete', 'create'])

// 刷新选择器的key
const refreshKey = ref(0)

// 刷新选择器
const refresh = () => {
  refreshKey.value++
}

// 暴露方法给父组件
defineExpose({
  refresh
})
</script>

<style scoped>
.config-selector {
  margin-bottom: 6px;
}

.selector-row {
  display: flex;
  align-items: center;
  gap: 4px;
  width: 100%;
}

.config-select {
  flex: 1;
}

.action-buttons {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;
}

.config-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 2px 0;
}

.config-info {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
}

.option-buttons {
  display: flex;
  gap: 4px;
}

.edit-btn, .delete-btn {
  padding: 0;
  min-height: 24px;
  min-width: 24px;
  height: 24px;
  width: 24px;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}

.edit-btn .edit-icon,
.delete-btn :deep(.el-icon) {
  font-size: 12px;
  line-height: 1;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 12px;
  height: 12px;
}

.edit-icon {
  font-style: normal;
}

.edit-icon svg {
  width: 12px;
  height: 12px;
}

.create-new-option {
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  padding: 4px 0;
  color: var(--el-color-primary);
}

.create-text {
  font-weight: 600;
}
</style>

<style>
/* 全局样式确保下拉菜单中的按钮可见 */
.el-select-dropdown__item .config-option {
  display: flex !important;
  justify-content: space-between !important;
  align-items: center !important;
  width: 100% !important;
}

.el-select-dropdown__item .config-option .config-info {
  flex: 1 !important;
}

.el-select-dropdown__item .config-option .option-buttons {
  display: flex !important;
  visibility: visible !important;
  opacity: 1 !important;
  z-index: 10 !important;
  gap: 8px !important;
}

.el-select-dropdown__item .config-option .option-buttons .el-button {
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

/* 编辑按钮样式 */
.el-select-dropdown__item .config-option .option-buttons .edit-btn {
  background-color: #409EFF !important;
  border-color: #409EFF !important;
  color: white !important;
}

.el-select-dropdown__item .config-option .option-buttons .edit-btn .edit-icon {
  font-size: 12px !important;
  color: white !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
}

.el-select-dropdown__item .config-option .option-buttons .edit-btn .edit-icon svg {
  width: 12px !important;
  height: 12px !important;
}

/* 删除按钮样式 */
.el-select-dropdown__item .config-option .option-buttons .delete-btn {
  background-color: #F56C6C !important;
  border-color: #F56C6C !important;
  color: white !important;
}

.el-select-dropdown__item .config-option .option-buttons .delete-btn .el-icon {
  font-size: 12px !important;
  color: white !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  width: 12px !important;
  height: 12px !important;
}

.el-select-dropdown__item .config-option .option-buttons .delete-btn svg {
  width: 12px !important;
  height: 12px !important;
}

/* 确保下拉框选项有足够空间显示按钮 */
.el-select-dropdown__item {
  padding: 8px 12px !important;
  height: auto !important;
  line-height: 1.5 !important;
}

/* 确保选项行高足够 */
.el-select-dropdown__item .config-option {
  min-height: 36px !important;
}

/* 修改表格样式 */
.el-table .option-buttons {
  display: flex !important;
  visibility: visible !important;
  opacity: 1 !important;
  z-index: 10 !important;
  gap: 8px !important;
}

.el-table .option-buttons .el-button {
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

.el-table .option-buttons .edit-btn {
  background-color: #409EFF !important;
  border-color: #409EFF !important;
  color: white !important;
}

.el-table .option-buttons .edit-btn .edit-icon {
  font-size: 12px !important;
  color: white !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
}

.el-table .option-buttons .edit-btn .edit-icon svg {
  width: 12px !important;
  height: 12px !important;
}

.el-table .option-buttons .delete-btn {
  background-color: #F56C6C !important;
  border-color: #F56C6C !important;
  color: white !important;
}

.el-table .option-buttons .delete-btn .el-icon {
  font-size: 12px !important;
  color: white !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  width: 12px !important;
  height: 12px !important;
}

.el-table .option-buttons .delete-btn svg {
  width: 12px !important;
  height: 12px !important;
}

/* 针对SystemPromptConfig.vue的特定修复 */
.config-select .el-select-dropdown__item .config-option .option-buttons {
  position: static !important;
  display: flex !important;
}

.config-select .el-select-dropdown__item .config-option .option-buttons .el-button {
  display: inline-flex !important;
}
</style> 