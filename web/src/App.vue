<template>
  <div id="app">
    <MainLayout />

    <!-- 全局错误提示 -->
    <el-dialog
      v-model="showErrorDialog"
      :title="$t('common.error')"
      width="400px"
      :before-close="clearError"
    >
      <p>{{ appStore.error }}</p>
      <template #footer>
        <el-button @click="clearError">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, computed, watch, ref, provide } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useConfigStore } from '@/stores/config'
import { useChatStore } from '@/stores/chat'
import MainLayout from '@/components/layout/MainLayout.vue'

const { locale } = useI18n()
const appStore = useAppStore()
const configStore = useConfigStore()
const chatStore = useChatStore()

// 计算属性
const showErrorDialog = computed(() => !!appStore.error)

// 方法
const clearError = () => {
  appStore.clearError()
}

// 监听语言变化
watch(
  () => appStore.language,
  (newLanguage) => {
    locale.value = newLanguage
  },
  { immediate: true }
)

// 全局加载状态
const isAppInitialized = ref(false)
// 提供给子组件使用的应用初始化状态
provide('appInitialized', isAppInitialized)

// 初始化应用
onMounted(async () => {
  try {
    console.log('【App】onMounted开始执行', new Date().toISOString())
    
    if (isAppInitialized.value) {
      console.log('【App】应用已初始化，跳过', new Date().toISOString())
      return
    }
    
    // 初始化应用状态
    console.log('【App】初始化应用状态', new Date().toISOString())
    appStore.init()

    // 加载配置
    console.log('【App】开始加载配置', new Date().toISOString())
    try {
      await configStore.loadConfiguration()
      console.log('【App】配置加载完成', new Date().toISOString())
    } catch (configError) {
      console.error('【App】配置加载失败，但将继续初始化应用:', configError)
      // 不阻止应用初始化，使用默认配置
    }

    // 初始化SSE连接
    console.log('【App】初始化SSE连接', new Date().toISOString())
    chatStore.initSSE()

    isAppInitialized.value = true
    console.log('【App】应用初始化完成', new Date().toISOString())
    
  } catch (error) {
    console.error('应用初始化失败:', error)
    appStore.setError(error instanceof Error ? error.message : '应用初始化失败')
    isAppInitialized.value = false // 允许重试
  }
})

// 页面卸载时清理资源
window.addEventListener('beforeunload', () => {
  chatStore.disconnect()
})
</script>

<style>
/* 全局样式已在main.css中定义 */
</style>
