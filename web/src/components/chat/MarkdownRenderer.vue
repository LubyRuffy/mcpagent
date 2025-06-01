<template>
  <div class="markdown-renderer" v-html="renderedContent"></div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { marked } from 'marked'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.css'

interface Props {
  content: string
}

const props = defineProps<Props>()

// 配置marked
marked.setOptions({
  breaks: true,
  gfm: true
} as any)

// 渲染Markdown内容
const renderedContent = computed(() => {
  if (!props.content) return ''

  try {
    return marked(props.content)
  } catch (error) {
    console.error('Markdown渲染失败:', error)
    return props.content
  }
})
</script>

<style scoped>
.markdown-renderer {
  line-height: 1.6;
  color: var(--text-color-primary);
}

/* Markdown样式 */
.markdown-renderer :deep(h1),
.markdown-renderer :deep(h2),
.markdown-renderer :deep(h3),
.markdown-renderer :deep(h4),
.markdown-renderer :deep(h5),
.markdown-renderer :deep(h6) {
  margin: 16px 0 8px 0;
  font-weight: 600;
  color: var(--text-color-primary);
}

.markdown-renderer :deep(h1) {
  font-size: 24px;
  border-bottom: 2px solid var(--border-color);
  padding-bottom: 8px;
}

.markdown-renderer :deep(h2) {
  font-size: 20px;
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 4px;
}

.markdown-renderer :deep(h3) {
  font-size: 18px;
}

.markdown-renderer :deep(h4) {
  font-size: 16px;
}

.markdown-renderer :deep(h5),
.markdown-renderer :deep(h6) {
  font-size: 14px;
}

.markdown-renderer :deep(p) {
  margin: 8px 0;
}

.markdown-renderer :deep(ul),
.markdown-renderer :deep(ol) {
  margin: 8px 0;
  padding-left: 24px;
}

.markdown-renderer :deep(li) {
  margin: 4px 0;
}

.markdown-renderer :deep(blockquote) {
  margin: 16px 0;
  padding: 8px 16px;
  border-left: 4px solid var(--primary-color);
  background-color: var(--bg-color-secondary);
  color: var(--text-color-secondary);
}

.markdown-renderer :deep(code) {
  padding: 2px 4px;
  background-color: var(--bg-color-secondary);
  border: 1px solid var(--border-color);
  border-radius: 3px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
}

.markdown-renderer :deep(pre) {
  margin: 16px 0;
  padding: 16px;
  background-color: var(--bg-color-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-base);
  overflow-x: auto;
}

.markdown-renderer :deep(pre code) {
  padding: 0;
  background: none;
  border: none;
  font-size: 14px;
  line-height: 1.5;
}

.markdown-renderer :deep(table) {
  width: 100%;
  margin: 16px 0;
  border-collapse: collapse;
  border: 1px solid var(--border-color);
}

.markdown-renderer :deep(th),
.markdown-renderer :deep(td) {
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  text-align: left;
}

.markdown-renderer :deep(th) {
  background-color: var(--bg-color-secondary);
  font-weight: 600;
}

.markdown-renderer :deep(tr:nth-child(even)) {
  background-color: var(--bg-color-secondary);
}

.markdown-renderer :deep(a) {
  color: var(--primary-color);
  text-decoration: none;
}

.markdown-renderer :deep(a:hover) {
  text-decoration: underline;
}

.markdown-renderer :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: var(--border-radius-base);
  margin: 8px 0;
}

.markdown-renderer :deep(hr) {
  margin: 24px 0;
  border: none;
  border-top: 1px solid var(--border-color);
}

/* 代码高亮主题适配 */
.dark .markdown-renderer :deep(.hljs) {
  background: #2d3748 !important;
  color: #e2e8f0 !important;
}

.dark .markdown-renderer :deep(.hljs-keyword) {
  color: #81c784 !important;
}

.dark .markdown-renderer :deep(.hljs-string) {
  color: #ffb74d !important;
}

.dark .markdown-renderer :deep(.hljs-comment) {
  color: #90a4ae !important;
}

.dark .markdown-renderer :deep(.hljs-number) {
  color: #f48fb1 !important;
}

.dark .markdown-renderer :deep(.hljs-function) {
  color: #64b5f6 !important;
}
</style>
