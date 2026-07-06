<script setup>
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { CheckIcon } from '@heroicons/vue/24/outline'
import { Splitpanes, Pane } from 'splitpanes'
import 'splitpanes/dist/splitpanes.css'
import AppHeader from '../components/AppHeader.vue'
import AppFooter from '../components/AppFooter.vue'
import CodeEditor from '../components/CodeEditor.vue'
import ConsoleOutput from '../components/ConsoleOutput.vue'
import { useAuth } from '../composables/useAuth'
import { useAutosave } from '../composables/useAutosave'
import { useFilePresence } from '../composables/useFilePresence'
import { api } from '../composables/useApi'
import {
  detectLanguage,
  defaultTemplate,
  preserveExtension,
} from '../utils/language'

const WATCH_POLL_MS = 2000

const route = useRoute()
const router = useRouter()
const { isAdmin, user } = useAuth()

const fileName = ref('')
const fileUserId = ref(null)
const fileUser = ref(null)
const verified = ref(false)
const autosaveEnabled = ref(true)
const code = ref('// Loading...\n')
const logs = ref([])
const running = ref(false)
const loading = ref(true)
const horizontal = ref(false)
const editingName = ref(false)
const renameValue = ref('')
const lastRemoteUpdatedAt = ref(null)

let watchPollTimer = null

const isWatchMode = computed(() => route.query.watch === '1' && isAdmin.value)
const autosaveActive = computed(() => !isWatchMode.value)

const { saveStatus, syncBaseline, saveNow } = useAutosave({
  fileId: computed(() => route.params.id),
  content: code,
  enabled: autosaveEnabled,
  loading,
  active: autosaveActive,
})

useFilePresence({
  fileId: computed(() => route.params.id),
  fileUserId,
  user,
  isWatchMode,
  loading,
})

const saveStatusLabel = computed(() => {
  if (isWatchMode.value) return 'Watching'
  switch (saveStatus.value) {
    case 'saving':
      return 'Saving...'
    case 'unsaved':
      return 'Unsaved'
    case 'error':
      return 'Save failed'
    default:
      return autosaveEnabled.value ? 'Saved' : 'Manual save'
  }
})

const saveStatusClass = computed(() => {
  if (isWatchMode.value) return 'text-violet-600'
  switch (saveStatus.value) {
    case 'unsaved':
      return 'text-amber-600'
    case 'error':
      return 'text-red-600'
    case 'saving':
      return 'text-slate-500'
    default:
      return 'text-green-600'
  }
})

function startWatchPolling() {
  stopWatchPolling()
  if (!isWatchMode.value) return
  watchPollTimer = setInterval(pollWatchFile, WATCH_POLL_MS)
}

function stopWatchPolling() {
  if (watchPollTimer) {
    clearInterval(watchPollTimer)
    watchPollTimer = null
  }
}

async function pollWatchFile() {
  if (!isWatchMode.value || !route.params.id) return
  try {
    const file = await api(`/files/${route.params.id}`)
    if (file.updated_at !== lastRemoteUpdatedAt.value) {
      lastRemoteUpdatedAt.value = file.updated_at
      const next = file.content ?? defaultTemplate(detectLanguage(file.name))
      if (next !== code.value) {
        code.value = next
      }
    }
  } catch (e) {
    console.error(e)
  }
}

async function loadFile() {
  const id = route.params.id
  if (!id) {
    router.push('/files')
    return
  }
  loading.value = true
  stopWatchPolling()
  try {
    const file = await api(`/files/${id}`)
    fileName.value = file.name
    verified.value = file.verified || false
    autosaveEnabled.value = file.autosave_enabled !== false
    code.value = file.content || defaultTemplate(detectLanguage(file.name))
    fileUserId.value = file.user_id
    fileUser.value = file.user
    lastRemoteUpdatedAt.value = file.updated_at
    syncBaseline(code.value)
    if (isWatchMode.value) {
      startWatchPolling()
    }
  } catch (e) {
    alert(e.message)
    router.push('/files')
  } finally {
    loading.value = false
  }
}

async function runCode(source) {
  running.value = true
  logs.value = []
  try {
    const res = await api('/run', {
      method: 'POST',
      body: JSON.stringify({ code: source, language: language.value }),
    })
    if (res.error) {
      logs.value = [{ type: 'error', args: res.error }]
    } else {
      const lines = (res.output || '').split('\n').filter(Boolean)
      logs.value = lines.map((line) => ({ type: 'log', args: line }))
    }
  } catch (e) {
    logs.value = [{ type: 'error', args: e.message }]
  } finally {
    running.value = false
  }
}

function handleRun() {
  if (running.value || isWatchMode.value) return
  runCode(code.value)
}

async function save() {
  if (isWatchMode.value) return
  await saveNow()
}

function startRename() {
  if (isWatchMode.value) return
  editingName.value = true
  renameValue.value = fileName.value || ''
}

function cancelRename() {
  editingName.value = false
  renameValue.value = ''
}

async function saveRename() {
  const finalName = preserveExtension(renameValue.value, fileName.value)
  if (!route.params.id) return
  try {
    await api(`/files/${route.params.id}`, {
      method: 'PUT',
      body: JSON.stringify({ name: finalName }),
    })
    fileName.value = finalName
  } catch (e) {
    alert(e.message)
  }
  cancelRename()
}

async function toggleVerified() {
  if (!isAdmin.value || !route.params.id || isWatchMode.value) return
  try {
    const updated = await api(`/files/${route.params.id}`, {
      method: 'PUT',
      body: JSON.stringify({ verified: !verified.value }),
    })
    verified.value = updated.verified
  } catch (e) {
    alert(e.message)
  }
}

async function toggleAutosave() {
  if (!isAdmin.value || !route.params.id || isWatchMode.value) return
  const next = !autosaveEnabled.value
  try {
    const updated = await api(`/files/${route.params.id}`, {
      method: 'PUT',
      body: JSON.stringify({ autosave_enabled: next }),
    })
    autosaveEnabled.value = updated.autosave_enabled !== false
    if (autosaveEnabled.value && saveStatus.value === 'unsaved') {
      await saveNow()
    }
  } catch (e) {
    alert(e.message)
  }
}

const breadcrumbLabel = computed(() => {
  if (fileUserId.value == null) return 'Your files'
  const isOwn = user.value && fileUserId.value === user.value.id
  if (isOwn) return 'Your files'
  const u = fileUser.value
  return (u?.fullname || u?.email || 'Unknown').trim() || 'Unknown'
})

const language = computed(() => detectLanguage(fileName.value))

watch(() => route.params.id, loadFile, { immediate: true })

watch(isWatchMode, (watching) => {
  if (watching) {
    startWatchPolling()
  } else {
    stopWatchPolling()
  }
})

onBeforeUnmount(() => {
  stopWatchPolling()
})
</script>

<template>
  <div class="flex h-screen flex-col overflow-hidden bg-slate-50">
    <div
      v-if="isWatchMode"
      class="border-b border-violet-200 bg-violet-50 px-4 py-1 text-center text-xs font-medium text-violet-800"
    >
      Watch mode — read only. Updates appear after the owner saves.
    </div>
    <AppHeader>
      <template #left>
        <router-link
          to="/files"
          class="rounded border border-slate-300 bg-white px-2 py-1 text-xs font-medium text-slate-700 hover:bg-slate-50"
        >
          ← Back
        </router-link>
        <div v-if="breadcrumbLabel" class="flex items-center gap-1 text-sm text-slate-500">
          <span>{{ breadcrumbLabel }}</span>
          <span class="text-slate-400">/</span>
        </div>
        <div class="flex items-center gap-1.5">
          <input
            v-if="editingName"
            v-model="renameValue"
            type="text"
            class="max-w-[240px] rounded border border-slate-300 px-1.5 py-0.5 text-sm"
            @keydown.enter="saveRename"
            @keydown.esc="cancelRename"
            @blur="saveRename"
          />
          <h1
            v-else
            class="text-sm font-medium text-slate-800"
            :class="isWatchMode ? '' : 'cursor-pointer hover:text-blue-600'"
            :title="isWatchMode ? undefined : 'Click to rename'"
            @click="startRename"
          >
            {{ fileName || 'Loading...' }}
          </h1>
          <template v-if="!editingName && !isWatchMode">
            <button
              v-if="isAdmin"
              type="button"
              class="inline-flex h-4 w-4 shrink-0 items-center justify-center rounded border transition-colors"
              :class="verified ? 'border-green-500 bg-green-50 text-green-600 hover:bg-green-100' : 'border-slate-300 hover:border-slate-400 hover:bg-slate-50'"
              :title="verified ? 'Unverify' : 'Verify'"
              @click="toggleVerified"
            >
              <CheckIcon v-if="verified" class="h-3 w-3" stroke-width="3" />
            </button>
            <span
              v-else-if="verified"
              class="rounded bg-green-100 px-1 py-0.5 text-[10px] font-medium text-green-700"
            >
              Verified
            </span>
          </template>
        </div>
      </template>
      <span class="text-xs font-medium" :class="saveStatusClass">{{ saveStatusLabel }}</span>
      <label
        v-if="isAdmin && !isWatchMode"
        class="flex cursor-pointer items-center gap-1.5 text-xs text-slate-600"
        title="Only admins can change autosave for this file"
      >
        <input
          type="checkbox"
          :checked="autosaveEnabled"
          class="rounded"
          @change="toggleAutosave"
        />
        Autosave
      </label>
      <button
        type="button"
        class="rounded border border-slate-300 bg-white px-2 py-0.5 text-xs text-slate-600 hover:bg-slate-50"
        :title="horizontal ? 'Code left, console right' : 'Code top, console bottom'"
        @click="horizontal = !horizontal"
      >
        {{ horizontal ? '⊟ Vertical' : '⊞ Horizontal' }}
      </button>
      <button
        v-if="!isWatchMode"
        :disabled="saveStatus === 'saving'"
        class="rounded border border-slate-300 bg-white px-3 py-1 text-xs font-medium text-slate-700 hover:bg-slate-50 disabled:opacity-50"
        title="Save (Ctrl+S)"
        @click="save"
      >
        {{ saveStatus === 'saving' ? 'Saving...' : 'Save (Ctrl+S)' }}
      </button>
      <button
        v-if="!isWatchMode"
        :disabled="running"
        class="rounded bg-blue-600 px-3 py-1 text-xs font-medium text-white hover:bg-blue-700 disabled:opacity-50"
        @click="handleRun"
      >
        {{ running ? 'Running...' : 'Run (Ctrl+Enter)' }}
      </button>
    </AppHeader>

    <div class="flex-1 min-h-0 overflow-hidden">
      <Splitpanes :horizontal="horizontal" class="h-full">
        <Pane :min-size="35" :size="70">
          <div class="h-full">
            <CodeEditor
              v-model="code"
              :language="language"
              :read-only="isWatchMode"
              @run="handleRun"
              @save="save"
            />
          </div>
        </Pane>
        <Pane :min-size="10" :size="30">
          <div class="flex h-full flex-col border-t border-slate-200 bg-white">
            <div class="flex-1 min-h-0 overflow-auto">
              <ConsoleOutput :logs="logs" :running="running" />
            </div>
          </div>
        </Pane>
      </Splitpanes>
    </div>

    <AppFooter />
  </div>
</template>
