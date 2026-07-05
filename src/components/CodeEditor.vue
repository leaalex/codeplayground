<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { monacoLanguage } from '../utils/language'

const props = defineProps({
  modelValue: {
    type: String,
    default: '',
  },
  language: {
    type: String,
    default: 'go',
  },
})

const emit = defineEmits(['update:modelValue', 'run', 'save'])

const editorRef = ref(null)
let editor = null
let monaco = null

onMounted(async () => {
  monaco = await import('monaco-editor')
  editor = monaco.editor.create(editorRef.value, {
    value: props.modelValue,
    language: monacoLanguage(props.language),
    theme: 'vs',
    automaticLayout: true,
    minimap: { enabled: false },
    lineNumbers: 'on',
    scrollBeyondLastLine: false,
    fontSize: 14,
    fontFamily: "'IBM Plex Mono', monospace",
  })

  editor.onDidChangeModelContent(() => {
    emit('update:modelValue', editor.getValue())
  })

  editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => {
    emit('run')
  })
  editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
    emit('save')
  })
})

onBeforeUnmount(() => {
  if (editor) {
    editor.dispose()
  }
})

watch(
  () => props.modelValue,
  (newVal) => {
    if (editor && editor.getValue() !== newVal) {
      editor.setValue(newVal)
    }
  }
)

watch(
  () => props.language,
  (lang) => {
    if (editor && monaco) {
      monaco.editor.setModelLanguage(editor.getModel(), monacoLanguage(lang))
    }
  }
)
</script>

<template>
  <div ref="editorRef" class="code-editor"></div>
</template>

<style scoped>
.code-editor {
  width: 100%;
  height: 100%;
}
</style>
