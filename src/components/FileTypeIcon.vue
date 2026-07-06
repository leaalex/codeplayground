<script setup>
import { computed } from 'vue'
import { Icon } from '@iconify/vue'
import goIcon from '@iconify-icons/vscode-icons/file-type-go'
import pythonIcon from '@iconify-icons/vscode-icons/file-type-python'
import markdownIcon from '@iconify-icons/vscode-icons/file-type-markdown'
import { detectLanguage } from '../utils/language'

const props = defineProps({
  filename: {
    type: String,
    default: '',
  },
  size: {
    type: String,
    default: 'md',
  },
})

const lang = computed(() => detectLanguage(props.filename))

const icon = computed(() => {
  switch (lang.value) {
    case 'python':
      return pythonIcon
    case 'markdown':
      return markdownIcon
    default:
      return goIcon
  }
})

const iconSize = computed(() => (props.size === 'sm' ? 14 : 16))

const title = computed(() => {
  switch (lang.value) {
    case 'python':
      return 'Python'
    case 'markdown':
      return 'Markdown'
    default:
      return 'Go'
  }
})
</script>

<template>
  <Icon
    :icon="icon"
    :width="iconSize"
    :height="iconSize"
    :title="title"
    class="inline-flex shrink-0"
    aria-hidden="true"
  />
</template>
