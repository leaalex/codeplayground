<script setup>
import { computed, ref, watch } from 'vue'
import MarkdownPreview from './MarkdownPreview.vue'
import TipTapEditor from './TipTapEditor.vue'
import { api } from '../composables/useApi'

const props = defineProps({
  instructions: {
    type: Object,
    default: null,
  },
  isAdmin: {
    type: Boolean,
    default: false,
  },
  markdownFiles: {
    type: Array,
    default: () => [],
  },
  instructionsFileId: {
    type: Number,
    default: null,
  },
  readOnly: {
    type: Boolean,
    default: true,
  },
})

const emit = defineEmits(['link', 'updated'])

const draft = ref('')
const savedContent = ref('')
const saveStatus = ref('saved')

const canEdit = computed(
  () => props.isAdmin && !props.readOnly && Boolean(props.instructions?.id)
)
const content = computed(() => props.instructions?.content || '')
const title = computed(() => props.instructions?.name || 'Assignment')
const isDirty = computed(() => draft.value !== savedContent.value)

const saveStatusLabel = computed(() => {
  switch (saveStatus.value) {
    case 'saving':
      return 'Saving...'
    case 'unsaved':
      return 'Unsaved'
    case 'error':
      return 'Save failed'
    default:
      return 'Saved'
  }
})

const saveStatusClass = computed(() => {
  switch (saveStatus.value) {
    case 'unsaved':
      return 'text-amber-600'
    case 'error':
      return 'text-red-600'
    case 'saving':
      return 'text-slate-500'
    default:
      return 'text-slate-400'
  }
})

watch(
  () => [props.instructions?.id, props.instructions?.content],
  () => {
    const next = props.instructions?.content || ''
    draft.value = next
    savedContent.value = next
    saveStatus.value = 'saved'
  },
  { immediate: true }
)

watch(draft, (val) => {
  if (!canEdit.value) return
  if (saveStatus.value === 'saving') return
  saveStatus.value = val !== savedContent.value ? 'unsaved' : 'saved'
})

function onLinkChange(e) {
  const val = e.target.value
  if (val === '') {
    emit('link', { clear: true })
  } else {
    emit('link', { id: Number(val) })
  }
}

async function saveInstructions() {
  if (!canEdit.value || !props.instructions?.id) return
  if (!isDirty.value && saveStatus.value !== 'error') return
  saveStatus.value = 'saving'
  try {
    const updated = await api(`/files/${props.instructions.id}`, {
      method: 'PUT',
      body: JSON.stringify({ content: draft.value }),
    })
    savedContent.value = draft.value
    saveStatus.value = 'saved'
    emit('updated', {
      id: updated.id,
      name: updated.name,
      content: updated.content,
    })
  } catch (e) {
    saveStatus.value = 'error'
    alert(e.message)
  }
}
</script>

<template>
  <div class="flex h-full flex-col border-r border-slate-200 bg-white">
    <div class="flex shrink-0 items-center justify-between gap-2 border-b border-slate-200 px-3 py-2">
      <h2 class="min-w-0 truncate text-xs font-semibold text-slate-700">{{ title }}</h2>
      <div v-if="canEdit" class="flex shrink-0 items-center gap-2">
        <span class="text-[10px]" :class="saveStatusClass">{{ saveStatusLabel }}</span>
        <button
          type="button"
          class="rounded border border-slate-300 bg-white px-2 py-0.5 text-[10px] font-medium text-slate-700 hover:bg-slate-50 disabled:opacity-50"
          :disabled="saveStatus === 'saving' || (!isDirty && saveStatus !== 'error')"
          title="Save (Ctrl+S)"
          @click="saveInstructions"
        >
          Save
        </button>
      </div>
    </div>
    <div v-if="isAdmin && !readOnly && markdownFiles.length" class="shrink-0 border-b border-slate-200 px-3 py-2">
      <label class="mb-0.5 block text-[10px] font-medium text-slate-500">Link instruction</label>
      <select
        :value="instructionsFileId ?? ''"
        class="w-full rounded border border-slate-300 px-2 py-1 text-xs"
        @change="onLinkChange"
      >
        <option value="">None</option>
        <option v-for="md in markdownFiles" :key="md.id" :value="md.id">
          {{ md.name }}
        </option>
      </select>
    </div>
    <div class="min-h-0 flex-1 overflow-hidden">
      <TipTapEditor
        v-if="canEdit"
        v-model="draft"
        @save="saveInstructions"
      />
      <div v-else class="h-full overflow-auto">
        <MarkdownPreview v-if="content" :content="content" />
        <p v-else class="px-4 py-3 text-xs text-slate-400">No assignment linked.</p>
      </div>
    </div>
  </div>
</template>
