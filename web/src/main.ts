import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'

import App from './App.vue'
import zhCN from './locales/zh-CN'
import enUS from './locales/en-US'
import './styles/main.css'

// 创建i18n实例
const i18n = createI18n({
  legacy: false,
  locale: localStorage.getItem('mcpagent-language') || 'zh-CN',
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS
  }
})

// 创建Pinia实例
const pinia = createPinia()

// 创建Vue应用
const app = createApp(App)

// 使用插件
app.use(pinia)
app.use(i18n)
app.use(ElementPlus)

// 挂载应用
app.mount('#app')
