<script setup>
import { ref, watch, onBeforeUnmount } from 'vue'

const props = defineProps({
  logs: {
    type: Array,
    default: () => [],
  },
  running: {
    type: Boolean,
    default: false,
  },
})

const stages = [
  'Building your project...',
  'Launching your program...',
  'Gathering results...',
]

const stageIndex = ref(0)
let stageTimer = null

watch(
  () => props.running,
  (isRunning) => {
    clearInterval(stageTimer)
    if (isRunning) {
      stageIndex.value = 0
      stageTimer = setInterval(() => {
        if (stageIndex.value < stages.length - 1) {
          stageIndex.value++
        }
      }, 4000)
    }
  }
)

onBeforeUnmount(() => clearInterval(stageTimer))
</script>

<template>
  <div class="flex h-full min-h-0 flex-1 flex-col overflow-y-auto p-2 font-mono text-sm leading-tight">
    <div
      v-if="running"
      class="flex flex-1 flex-col items-center justify-center gap-3 py-8 text-slate-500"
    >
      <div class="h-8 w-8 animate-spin rounded-full border-2 border-slate-200 border-t-blue-600" />
      <p class="text-sm font-medium">{{ stages[stageIndex] }}</p>
    </div>
    <template v-else>
      <div
        v-for="(log, i) in logs"
        :key="i"
        class="break-all leading-tight"
        :class="{
          'text-slate-800': log.type === 'log',
          'text-red-600': log.type === 'error',
          'text-amber-600': log.type === 'warn',
        }"
      >
        {{ log.args }}
      </div>
      <div v-if="logs.length === 0" class="italic text-slate-400">
        Output will appear here after you run your code.
      </div>
    </template>
  </div>
</template>
