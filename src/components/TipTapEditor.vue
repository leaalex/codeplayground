<script setup>
import { watch } from 'vue'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import { Markdown } from '@tiptap/markdown'
import {
  BoldIcon,
  ItalicIcon,
  H1Icon,
  H2Icon,
  H3Icon,
  ListBulletIcon,
  NumberedListIcon,
  ChatBubbleBottomCenterTextIcon,
  CodeBracketIcon,
  ArrowUturnLeftIcon,
  ArrowUturnRightIcon,
} from '@heroicons/vue/24/outline'

const props = defineProps({
  modelValue: {
    type: String,
    default: '',
  },
  readOnly: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:modelValue', 'save'])

const editor = useEditor({
  extensions: [StarterKit, Markdown],
  content: props.modelValue,
  contentType: 'markdown',
  editable: !props.readOnly,
  onUpdate: ({ editor: ed }) => {
    emit('update:modelValue', ed.getMarkdown())
  },
  editorProps: {
    handleKeyDown(_view, event) {
      if ((event.ctrlKey || event.metaKey) && event.key === 's') {
        event.preventDefault()
        emit('save')
        return true
      }
      return false
    },
  },
})

watch(
  () => props.modelValue,
  (val) => {
    const ed = editor.value
    if (!ed) return
    const current = ed.getMarkdown()
    if (val !== current) {
      ed.commands.setContent(val || '', { contentType: 'markdown', emitUpdate: false })
    }
  }
)

watch(
  () => props.readOnly,
  (readOnly) => {
    editor.value?.setEditable(!readOnly)
  }
)

function run(cmd) {
  if (props.readOnly || !editor.value) return
  cmd()
}

function btnClass(active, disabled = false) {
  if (disabled) return 'text-slate-300 cursor-not-allowed'
  return active
    ? 'bg-slate-200 text-slate-900'
    : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900'
}
</script>

<template>
  <div class="flex h-full flex-col bg-white">
    <div
      v-if="editor && !readOnly"
      class="flex shrink-0 flex-wrap items-center gap-0.5 border-b border-slate-200 px-2 py-1"
    >
      <button
        type="button"
        class="rounded p-1.5"
        :class="btnClass(editor.isActive('bold'))"
        title="Bold"
        @click="run(() => editor.chain().focus().toggleBold().run())"
      >
        <BoldIcon class="h-4 w-4" />
      </button>
      <button
        type="button"
        class="rounded p-1.5"
        :class="btnClass(editor.isActive('italic'))"
        title="Italic"
        @click="run(() => editor.chain().focus().toggleItalic().run())"
      >
        <ItalicIcon class="h-4 w-4" />
      </button>
      <span class="mx-1 h-4 w-px bg-slate-200" />
      <button
        type="button"
        class="rounded p-1.5"
        :class="btnClass(editor.isActive('heading', { level: 1 }))"
        title="Heading 1"
        @click="run(() => editor.chain().focus().toggleHeading({ level: 1 }).run())"
      >
        <H1Icon class="h-4 w-4" />
      </button>
      <button
        type="button"
        class="rounded p-1.5"
        :class="btnClass(editor.isActive('heading', { level: 2 }))"
        title="Heading 2"
        @click="run(() => editor.chain().focus().toggleHeading({ level: 2 }).run())"
      >
        <H2Icon class="h-4 w-4" />
      </button>
      <button
        type="button"
        class="rounded p-1.5"
        :class="btnClass(editor.isActive('heading', { level: 3 }))"
        title="Heading 3"
        @click="run(() => editor.chain().focus().toggleHeading({ level: 3 }).run())"
      >
        <H3Icon class="h-4 w-4" />
      </button>
      <span class="mx-1 h-4 w-px bg-slate-200" />
      <button
        type="button"
        class="rounded p-1.5"
        :class="btnClass(editor.isActive('bulletList'))"
        title="Bullet list"
        @click="run(() => editor.chain().focus().toggleBulletList().run())"
      >
        <ListBulletIcon class="h-4 w-4" />
      </button>
      <button
        type="button"
        class="rounded p-1.5"
        :class="btnClass(editor.isActive('orderedList'))"
        title="Ordered list"
        @click="run(() => editor.chain().focus().toggleOrderedList().run())"
      >
        <NumberedListIcon class="h-4 w-4" />
      </button>
      <button
        type="button"
        class="rounded p-1.5"
        :class="btnClass(editor.isActive('blockquote'))"
        title="Blockquote"
        @click="run(() => editor.chain().focus().toggleBlockquote().run())"
      >
        <ChatBubbleBottomCenterTextIcon class="h-4 w-4" />
      </button>
      <button
        type="button"
        class="rounded p-1.5"
        :class="btnClass(editor.isActive('codeBlock'))"
        title="Code block"
        @click="run(() => editor.chain().focus().toggleCodeBlock().run())"
      >
        <CodeBracketIcon class="h-4 w-4" />
      </button>
      <span class="mx-1 h-4 w-px bg-slate-200" />
      <button
        type="button"
        class="rounded p-1.5"
        :class="btnClass(false, !editor.can().undo())"
        title="Undo"
        :disabled="!editor.can().undo()"
        @click="run(() => editor.chain().focus().undo().run())"
      >
        <ArrowUturnLeftIcon class="h-4 w-4" />
      </button>
      <button
        type="button"
        class="rounded p-1.5"
        :class="btnClass(false, !editor.can().redo())"
        title="Redo"
        :disabled="!editor.can().redo()"
        @click="run(() => editor.chain().focus().redo().run())"
      >
        <ArrowUturnRightIcon class="h-4 w-4" />
      </button>
    </div>
    <div class="min-h-0 flex-1 overflow-auto">
      <EditorContent v-if="editor" :editor="editor" class="tiptap-editor h-full" />
    </div>
  </div>
</template>

<style scoped>
.tiptap-editor :deep(.tiptap) {
  min-height: 100%;
  padding: 1rem 1.25rem;
  outline: none;
  font-size: 0.875rem;
  line-height: 1.6;
  color: #1e293b;
}

.tiptap-editor :deep(.tiptap > * + *) {
  margin-top: 0.75rem;
}

.tiptap-editor :deep(.tiptap h1) {
  font-size: 1.25rem;
  font-weight: 600;
  line-height: 1.3;
}

.tiptap-editor :deep(.tiptap h2) {
  font-size: 1.1rem;
  font-weight: 600;
  line-height: 1.35;
}

.tiptap-editor :deep(.tiptap h3) {
  font-size: 1rem;
  font-weight: 600;
  line-height: 1.4;
}

.tiptap-editor :deep(.tiptap ul),
.tiptap-editor :deep(.tiptap ol) {
  padding-left: 1.25rem;
}

.tiptap-editor :deep(.tiptap blockquote) {
  border-left: 3px solid #cbd5e1;
  padding-left: 0.75rem;
  color: #64748b;
}

.tiptap-editor :deep(.tiptap pre) {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  padding: 0.75rem;
  overflow-x: auto;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 0.8125rem;
}

.tiptap-editor :deep(.tiptap code) {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 0.85em;
  background: #f1f5f9;
  padding: 0.1em 0.35em;
  border-radius: 0.25rem;
}

.tiptap-editor :deep(.tiptap pre code) {
  background: none;
  padding: 0;
}
</style>
