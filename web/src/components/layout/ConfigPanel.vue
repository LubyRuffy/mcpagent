<template>
  <div class="config-panel">
    <!-- 配置面板标题 -->
    <div class="panel-header">
      <h2 class="panel-title">{{ $t('config.title') }}</h2>

      <!-- 配置操作按钮 -->
      <div class="panel-actions">
        <el-dropdown @command="handleConfigAction" trigger="click">
          <el-button circle size="small">
            <el-icon><MoreFilled /></el-icon>
          </el-button>

          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="save">
                <el-icon><DocumentAdd /></el-icon>
                {{ $t('config.actions.save') }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <!-- 配置内容区域 -->
    <div class="panel-content">
      <el-scrollbar>
        <div class="config-sections">
          <!-- LLM配置 -->
          <LLMConfig />

          <!-- MCP服务器配置 -->
          <MCPServerConfig />

          <!-- 系统提示词配置 -->
          <SystemPromptConfig />

          <!-- 其他配置 -->
          <OtherConfig />
        </div>
      </el-scrollbar>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  MoreFilled,
  DocumentAdd
} from '@element-plus/icons-vue'
import { useConfigStore } from '@/stores/config'
import LLMConfig from '@/components/config/LLMConfig.vue'
import MCPServerConfig from '@/components/config/MCPServerConfig.vue'
import SystemPromptConfig from '@/components/config/SystemPromptConfig.vue'
import OtherConfig from '@/components/config/OtherConfig.vue'
import { configApi } from '@/utils/api'

const { t } = useI18n()
const configStore = useConfigStore()

// 保存配置
const saveConfig = async () => {
  try {
    await configStore.saveConfiguration()
    ElMessage.success(t('success.configSaved'))
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('error.saveConfigFailed'))
  }
}

// 处理配置操作
const handleConfigAction = async (command: string) => {
  switch (command) {
    case 'save':
      await saveConfig()
      break
  }
}
</script>

<style scoped>
.config-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: var(--bg-color-secondary);
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 10px;
  border-bottom: 1px solid var(--border-color);
  background-color: var(--bg-color);
}

.panel-title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.panel-content {
  flex: 1;
  overflow: hidden;
}

.config-sections {
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.panel-footer {
  padding: 8px 10px;
  border-top: 1px solid var(--border-color);
  background-color: var(--bg-color);
  position: sticky;
  bottom: 0;
  z-index: 10;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
}

.save-button {
  width: 100%;
  max-width: 100%;
  font-weight: 500;
  height: 32px;
  font-size: 13px;
}

:deep(.el-dropdown-menu__item) {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}

:deep(.el-card) {
  --el-card-padding: 8px;
  margin-bottom: 0;
}

:deep(.el-card__header) {
  padding: 8px 10px;
  min-height: auto;
}

:deep(.el-card__body) {
  padding: 8px 10px;
}

:deep(.el-form-item) {
  margin-bottom: 8px;
}

:deep(.el-form-item__label) {
  font-size: 13px;
  padding-bottom: 2px;
}

:deep(.el-input__wrapper) {
  padding: 0 8px;
}

:deep(.el-input__inner) {
  height: 28px;
  font-size: 13px;
}

:deep(.el-button) {
  padding: 6px 12px;
  font-size: 13px;
}

:deep(.el-button--small) {
  padding: 4px 10px;
  font-size: 12px;
}

:deep(.el-select) {
  width: 100%;
}

:deep(.el-descriptions__label) {
  padding: 6px 8px;
  font-size: 12px;
}

:deep(.el-descriptions__content) {
  padding: 6px 8px;
  font-size: 12px;
}

:deep(.el-collapse-item__header) {
  padding: 6px 8px;
  font-size: 13px;
  height: auto;
}

:deep(.el-collapse-item__content) {
  padding: 6px 8px;
}

:deep(.el-divider__text) {
  font-size: 13px;
}

/* 确保小屏幕上按钮仍然可见 */
@media (max-width: 500px) {
  .panel-footer {
    padding: 6px 8px;
  }
  
  .save-button {
    font-size: 12px;
    height: 28px;
  }
  
  .config-sections {
    padding-bottom: 40px; /* 确保内容不会被底部按钮遮挡 */
  }
}
</style>
