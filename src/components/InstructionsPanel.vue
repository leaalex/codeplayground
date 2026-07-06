<script setup>
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import MarkdownPreview from './MarkdownPreview.vue'

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

const emit = defineEmits(['link'])

const content = computed(() => props.instructions?.content || '')
const title = computed(() => props.instructions?.name || 'Assignment')

function onLinkChange(e) {
  const val = e.target.value
  if (val === '') {
    emit('link', { clear: true })
  } else {
    emit('link', { id: Number(val) })
  }
}
</script>

<template>
  <div class="flex h-full flex-col border-r border-slate-200 bg-white">
    <div class="flex shrink-0 items-center justify-between border-b border-slate-200 px-3 py-2">
      <h2 class="truncate text-xs font-semibold text-slate-700">{{ title }}</h2>
      <RouterLink
        v-if="isAdmin && instructions?.id && !readOnly"
        :to="{ name: 'playground', params: { id: instructions.id } }"
        class="shrink-0 text-[10px] font-medium text-blue-600 hover:underline"
      >
        Edit
      </RouterLink>
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
    <div class="min-h-0 flex-1 overflow-auto">
      <MarkdownPreview v-if="content" :content="content" />
      <p v-else class="px-4 py-3 text-xs text-slate-400">No assignment linked.</p>
    </div>
  </div>
</template>
