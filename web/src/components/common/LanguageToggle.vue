<template>
  <el-dropdown @command="handleLanguageChange" trigger="click">
    <el-button circle :title="$t('language.toggle')">
      🌐
    </el-button>

    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item
          command="zh-CN"
          :class="{ active: appStore.language === 'zh-CN' }"
        >
          🇨🇳 {{ $t('language.zh-CN') }}
        </el-dropdown-item>

        <el-dropdown-item
          command="en-US"
          :class="{ active: appStore.language === 'en-US' }"
        >
          🇺🇸 {{ $t('language.en-US') }}
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import { useAppStore } from '@/stores/app'
import type { Language } from '@/stores/app'

const appStore = useAppStore()

// 处理语言切换
const handleLanguageChange = (language: Language) => {
  appStore.setLanguage(language)
}
</script>

<style scoped>
.active {
  background-color: var(--primary-color);
  color: white;
}

:deep(.el-dropdown-menu__item) {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
