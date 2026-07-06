<script setup>
import { computed } from 'vue'
import DOMPurify from 'dompurify'
import { markdown } from '../utils/markdown'
import 'highlight.js/styles/github.min.css'

const props = defineProps({
  content: {
    type: String,
    default: '',
  },
})

const html = computed(() => {
  const raw = markdown.parse(props.content || '', { async: false })
  return DOMPurify.sanitize(raw, {
    ADD_ATTR: ['class'],
  })
})
</script>

<template>
  <div class="markdown-preview prose prose-sm max-w-none px-4 py-3 text-slate-800" v-html="html" />
</template>

<style scoped>
.markdown-preview :deep(h1) {
  font-size: 1.25rem;
  font-weight: 600;
  margin: 0 0 0.75rem;
}
.markdown-preview :deep(h2) {
  font-size: 1.1rem;
  font-weight: 600;
  margin: 1rem 0 0.5rem;
}
.markdown-preview :deep(h3) {
  font-size: 1rem;
  font-weight: 600;
  margin: 0.75rem 0 0.5rem;
}
.markdown-preview :deep(p) {
  margin: 0 0 0.75rem;
  line-height: 1.6;
}
.markdown-preview :deep(ul),
.markdown-preview :deep(ol) {
  margin: 0 0 0.75rem;
  padding-left: 1.25rem;
}
.markdown-preview :deep(li) {
  margin: 0.25rem 0;
}
.markdown-preview :deep(code:not(.hljs)) {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 0.85em;
  background: #f1f5f9;
  padding: 0.1em 0.35em;
  border-radius: 0.25rem;
}
.markdown-preview :deep(pre) {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  padding: 0.75rem;
  overflow-x: auto;
  margin: 0 0 0.75rem;
}
.markdown-preview :deep(pre code.hljs) {
  display: block;
  padding: 0;
  background: none;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 0.8125rem;
  line-height: 1.5;
}
.markdown-preview :deep(blockquote) {
  border-left: 3px solid #cbd5e1;
  padding-left: 0.75rem;
  color: #64748b;
  margin: 0 0 0.75rem;
}
.markdown-preview :deep(a) {
  color: #2563eb;
  text-decoration: underline;
}
</style>
