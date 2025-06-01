import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export type Theme = 'light' | 'dark' | 'auto'
export type Language = 'zh-CN' | 'en-US'

export const useAppStore = defineStore('app', () => {
  // 状态
  const theme = ref<Theme>('auto')
  const language = ref<Language>('zh-CN')
  const sidebarCollapsed = ref(false)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 计算属性
  const isDark = computed(() => {
    if (theme.value === 'auto') {
      return window.matchMedia('(prefers-color-scheme: dark)').matches
    }
    return theme.value === 'dark'
  })

  // 方法
  const setTheme = (newTheme: Theme) => {
    theme.value = newTheme
    localStorage.setItem('mcpagent-theme', newTheme)
    updateThemeClass()
  }

  const setLanguage = (newLanguage: Language) => {
    language.value = newLanguage
    localStorage.setItem('mcpagent-language', newLanguage)
  }

  const toggleSidebar = () => {
    sidebarCollapsed.value = !sidebarCollapsed.value
    localStorage.setItem('mcpagent-sidebar-collapsed', String(sidebarCollapsed.value))
  }

  const setLoading = (isLoading: boolean) => {
    loading.value = isLoading
  }

  const setError = (errorMessage: string | null) => {
    error.value = errorMessage
  }

  const clearError = () => {
    error.value = null
  }

  const updateThemeClass = () => {
    const html = document.documentElement
    if (isDark.value) {
      html.classList.add('dark')
      html.setAttribute('data-theme', 'dark')
    } else {
      html.classList.remove('dark')
      html.setAttribute('data-theme', 'light')
    }
  }

  // 初始化
  const init = () => {
    // 从localStorage恢复设置
    const savedTheme = localStorage.getItem('mcpagent-theme') as Theme
    if (savedTheme) {
      theme.value = savedTheme
    }

    const savedLanguage = localStorage.getItem('mcpagent-language') as Language
    if (savedLanguage) {
      language.value = savedLanguage
    }

    const savedSidebarState = localStorage.getItem('mcpagent-sidebar-collapsed')
    if (savedSidebarState) {
      sidebarCollapsed.value = savedSidebarState === 'true'
    }

    updateThemeClass()

    // 监听系统主题变化
    if (theme.value === 'auto') {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
      mediaQuery.addEventListener('change', updateThemeClass)
    }
  }

  return {
    // 状态
    theme,
    language,
    sidebarCollapsed,
    loading,
    error,

    // 计算属性
    isDark,

    // 方法
    setTheme,
    setLanguage,
    toggleSidebar,
    setLoading,
    setError,
    clearError,
    init
  }
})
