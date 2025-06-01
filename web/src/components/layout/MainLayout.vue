<template>
  <div class="main-layout">
    <!-- 左侧配置面板 -->
    <div
      class="sidebar"
      :class="{
        collapsed: appStore.sidebarCollapsed,
        'mobile-open': mobileMenuOpen
      }"
    >
      <ConfigPanel />
    </div>

    <!-- 移动端遮罩层 -->
    <div
      v-if="isMobile && mobileMenuOpen"
      class="mobile-overlay"
      @click="closeMobileMenu"
    />

    <!-- 右侧主内容区域 -->
    <div class="main-content">
      <!-- 顶部工具栏 -->
      <div class="header">
        <div class="header-left">
          <el-button
            :icon="appStore.sidebarCollapsed ? Expand : Fold"
            circle
            @click="toggleSidebar"
          />
          <h1 class="app-title">{{ $t('app.title') }}</h1>
        </div>

        <div class="header-right">
          <!-- 主题切换 -->
          <ThemeToggle />

          <!-- 语言切换 -->
          <LanguageToggle />
        </div>
      </div>

      <!-- 聊天区域 -->
      <div class="chat-container">
        <ChatArea />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Expand, Fold } from '@element-plus/icons-vue'
import { useAppStore } from '@/stores/app'
import { useChatStore } from '@/stores/chat'
import ConfigPanel from './ConfigPanel.vue'
import ChatArea from './ChatArea.vue'
import ThemeToggle from '@/components/common/ThemeToggle.vue'
import LanguageToggle from '@/components/common/LanguageToggle.vue'

const appStore = useAppStore()
const chatStore = useChatStore()

// 响应式状态
const mobileMenuOpen = ref(false)
const windowWidth = ref(window.innerWidth)

// 计算属性
const isMobile = computed(() => windowWidth.value <= 768)

// 方法
const toggleSidebar = () => {
  if (isMobile.value) {
    mobileMenuOpen.value = !mobileMenuOpen.value
  } else {
    appStore.toggleSidebar()
  }
}

const closeMobileMenu = () => {
  mobileMenuOpen.value = false
}

const handleResize = () => {
  windowWidth.value = window.innerWidth
  if (!isMobile.value) {
    mobileMenuOpen.value = false
  }
}

// 生命周期
onMounted(() => {
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped>
.main-layout {
  display: flex;
  height: 100vh;
  background-color: var(--bg-color);
}

.sidebar {
  width: var(--sidebar-width);
  background-color: var(--bg-color-secondary);
  border-right: 1px solid var(--border-color);
  transition: width var(--transition-duration) ease, transform var(--transition-duration) ease; /* 确保宽度和变换都有过渡效果 */
  overflow: hidden;
  z-index: 100;
}

.sidebar.collapsed {
  width: var(--sidebar-collapsed-width);
  transform: translateX(-100%); /* 完全移出视图 */
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  height: 100%; /* 确保 main-content 高度为 100% */
}

.header {
  height: var(--header-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 12px;
  background-color: var(--bg-color);
  border-bottom: 1px solid var(--border-color);
  box-shadow: var(--shadow-light);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.app-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.chat-container {
  flex: 1; /* 让 chat-container 填充剩余空间 */
  display: flex;
  flex-direction: column;
  overflow: hidden;
  height: calc(100% - var(--header-height)); /* 减去 header 的高度 */
}

/* 移动端样式 */
.mobile-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 99;
}

@media (max-width: 768px) {
  .sidebar {
    position: fixed;
    left: 0;
    top: 0;
    height: 100vh;
    z-index: 100;
    transform: translateX(-100%);
    transition: transform var(--transition-duration) ease; /* 确保变换有过渡效果 */
  }

  .sidebar.mobile-open {
    transform: translateX(0);
  }

  .app-title {
    font-size: 16px;
  }

  .header {
    padding: 0 12px;
  }

  .header-right {
    gap: 8px;
  }
}
</style>
