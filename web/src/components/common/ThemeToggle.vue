<template>
  <el-dropdown @command="handleThemeChange" trigger="click">
    <el-button circle :icon="themeIcon" :title="$t('theme.toggle')" />
    
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item 
          command="light"
          :class="{ active: appStore.theme === 'light' }"
        >
          <el-icon><Sunny /></el-icon>
          {{ $t('theme.light') }}
        </el-dropdown-item>
        
        <el-dropdown-item 
          command="dark"
          :class="{ active: appStore.theme === 'dark' }"
        >
          <el-icon><Moon /></el-icon>
          {{ $t('theme.dark') }}
        </el-dropdown-item>
        
        <el-dropdown-item 
          command="auto"
          :class="{ active: appStore.theme === 'auto' }"
        >
          <el-icon><Monitor /></el-icon>
          {{ $t('theme.auto') }}
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Sunny, Moon, Monitor } from '@element-plus/icons-vue'
import { useAppStore } from '@/stores/app'
import type { Theme } from '@/stores/app'

const appStore = useAppStore()

// 计算当前主题图标
const themeIcon = computed(() => {
  switch (appStore.theme) {
    case 'light':
      return Sunny
    case 'dark':
      return Moon
    case 'auto':
      return Monitor
    default:
      return Monitor
  }
})

// 处理主题切换
const handleThemeChange = (theme: Theme) => {
  appStore.setTheme(theme)
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
