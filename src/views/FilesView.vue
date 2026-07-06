<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { PencilSquareIcon, CheckIcon, EyeIcon, XMarkIcon, SignalIcon, ArrowDownTrayIcon, TrashIcon } from '@heroicons/vue/24/outline'
import AppHeader from '../components/AppHeader.vue'
import AppFooter from '../components/AppFooter.vue'
import FileTypeIcon from '../components/FileTypeIcon.vue'
import { useAuth } from '../composables/useAuth'
import { useFilesFilters } from '../composables/useFilesFilters'
import { api } from '../composables/useApi'
import {
  detectLanguage,
  defaultFileName,
  defaultTemplate,
  preserveExtension,
  exportMimeType,
  isCodeFile,
  isMarkdownFile,
} from '../utils/language'

const LIST_POLL_MS = 10000

const route = useRoute()
const router = useRouter()
const { isAdmin, user } = useAuth()

const {
  searchQuery,
  verifiedFilter,
  userFilter,
  sortBy,
  sortAsc,
  groupBy,
  typeFilter,
  initFromQuery,
} = useFilesFilters(route, router, isAdmin)

const files = ref([])
const loading = ref(false)
const importInputRef = ref(null)
const editingFileId = ref(null)
const editingName = ref('')

const selectedFileIds = ref(new Set())
const previewFile = ref(null)
const allUsers = ref([])
const batchTransferUserId = ref('')
const presenceMap = ref({})
const refreshing = ref(false)

let listPollTimer = null

const availableUsers = computed(() => {
  if (!isAdmin.value) return []
  const currentUserId = user.value?.id
  const seen = new Set()
  const out = []
  for (const f of files.value) {
    const u = f.user
    const email = u?.email || 'unknown'
    if (seen.has(email)) continue
    seen.add(email)
    const label = (currentUserId && f.user_id === currentUserId) ? 'Your files' : (u?.fullname || email)
    out.push({ email, fullname: label })
  }
  return out.sort((a, b) => {
    if (a.fullname === 'Your files') return -1
    if (b.fullname === 'Your files') return 1
    return (a.fullname || a.email).localeCompare(b.fullname || b.email)
  })
})

const filteredFiles = computed(() => {
  let list = [...files.value]

  if (isAdmin.value && searchQuery.value.trim()) {
    const q = searchQuery.value.trim().toLowerCase()
    list = list.filter((f) => {
      const nameMatch = (f.name || '').toLowerCase().includes(q)
      const userMatch =
        (f.user?.fullname || '').toLowerCase().includes(q) ||
        (f.user?.email || '').toLowerCase().includes(q)
      return nameMatch || userMatch
    })
  }

  if (isAdmin.value && verifiedFilter.value === 'verified') {
    list = list.filter((f) => f.verified)
  } else if (isAdmin.value && verifiedFilter.value === 'unverified') {
    list = list.filter((f) => !f.verified)
  }

  if (isAdmin.value && userFilter.value) {
    list = list.filter((f) => (f.user?.email || 'unknown') === userFilter.value)
  }

  if (isAdmin.value && typeFilter.value === 'code') {
    list = list.filter((f) => isCodeFile(f.name))
  } else if (isAdmin.value && typeFilter.value === 'instructions') {
    list = list.filter((f) => isMarkdownFile(f.name))
  }

  list.sort((a, b) => {
    let cmp = 0
    if (sortBy.value === 'name') {
      cmp = (a.name || '').localeCompare(b.name || '')
    } else if (sortBy.value === 'verified') {
      cmp = (a.verified ? 1 : 0) - (b.verified ? 1 : 0)
    } else {
      cmp = new Date(a.updated_at || 0) - new Date(b.updated_at || 0)
    }
    return sortAsc.value ? cmp : -cmp
  })

  return list
})

const groupedFiles = computed(() => {
  const list = filteredFiles.value
  if (!isAdmin.value) {
    return [{ label: null, isOwn: false, files: list }]
  }

  const mode = groupBy.value

  if (mode === 'none') {
    return [{ label: null, isOwn: false, files: list }]
  }

  if (mode === 'verified') {
    const verified = list.filter((f) => f.verified)
    const unverified = list.filter((f) => !f.verified)
    return [
      ...(verified.length > 0 ? [{ label: `Verified (${verified.length})`, isOwn: false, files: verified }] : []),
      ...(unverified.length > 0 ? [{ label: `Not verified (${unverified.length})`, isOwn: false, files: unverified }] : []),
    ]
  }

  const currentUserId = user.value?.id
  const ownFiles = []
  const byUser = new Map()
  for (const f of list) {
    if (currentUserId && f.user_id === currentUserId) {
      ownFiles.push(f)
      continue
    }
    const key = f.user ? `${f.user.email}` : 'unknown'
    if (!byUser.has(key)) {
      const u = f.user || { fullname: 'Unknown', email: 'unknown' }
      byUser.set(key, {
        label: `${u.fullname || u.email} (${u.email})`,
        isOwn: false,
        files: [],
        user: u,
      })
    }
    byUser.get(key).files.push(f)
  }
  const result = []
  if (ownFiles.length > 0) {
    result.push({ label: `Your files -- ${ownFiles.length} files, ${ownFiles.filter((f) => f.verified).length} verified`, isOwn: true, files: ownFiles })
  }
  for (const g of byUser.values()) {
    g.label = `${g.label} -- ${g.files.length} files, ${g.files.filter((f) => f.verified).length} verified`
    result.push(g)
  }
  return result
})

const markdownFiles = computed(() => files.value.filter((f) => isMarkdownFile(f.name)))

const importAccept = computed(() => (isAdmin.value ? '.go,.py,.md' : '.go,.py'))

function setSort(field) {
  if (sortBy.value === field) {
    sortAsc.value = !sortAsc.value
  } else {
    sortBy.value = field
    sortAsc.value = field === 'name'
  }
}

function toggleSelect(file) {
  const next = new Set(selectedFileIds.value)
  if (next.has(file.id)) next.delete(file.id)
  else next.add(file.id)
  selectedFileIds.value = next
}

function toggleSelectAllInGroup(group) {
  const ids = group.files.map((f) => f.id)
  const allSelected = ids.every((id) => selectedFileIds.value.has(id))
  const next = new Set(selectedFileIds.value)
  for (const id of ids) {
    if (allSelected) next.delete(id)
    else next.add(id)
  }
  selectedFileIds.value = next
}

function isGroupAllSelected(group) {
  return group.files.length > 0 && group.files.every((f) => selectedFileIds.value.has(f.id))
}

async function linkInstructions(file, e) {
  if (!isAdmin.value || !isCodeFile(file.name)) return
  const val = e.target.value
  try {
    const body = val === ''
      ? { clear_instructions: true }
      : { instructions_file_id: Number(val) }
    const updated = await api(`/files/${file.id}`, {
      method: 'PUT',
      body: JSON.stringify(body),
    })
    file.instructions_file_id = updated.instructions_file_id ?? null
  } catch (err) {
    alert(err.message)
  }
}

function instructionName(fileId) {
  const md = markdownFiles.value.find((f) => f.id === fileId)
  return md?.name || ''
}

async function batchVerify(verified) {
  const ids = [...selectedFileIds.value]
  if (!ids.length) return
  try {
    await Promise.all(
      ids.map((id) =>
        api(`/files/${id}`, { method: 'PUT', body: JSON.stringify({ verified }) })
      )
    )
    for (const id of ids) {
      const f = files.value.find((x) => x.id === id)
      if (f) f.verified = verified
    }
  } catch (e) {
    alert(e.message)
  }
  selectedFileIds.value = new Set()
}

async function batchDelete() {
  const ids = [...selectedFileIds.value]
  if (!ids.length) return
  const count = ids.length
  if (!confirm(`Delete ${count} selected file${count === 1 ? '' : 's'}?`)) return
  try {
    await Promise.all(ids.map((id) => api(`/files/${id}`, { method: 'DELETE' })))
    const idSet = new Set(ids)
    files.value = files.value.filter((f) => !idSet.has(f.id))
    if (previewFile.value && idSet.has(previewFile.value.id)) {
      previewFile.value = null
    }
  } catch (e) {
    alert(e.message)
  }
  selectedFileIds.value = new Set()
}

function userOptionLabel(u) {
  return (u?.fullname || u?.email || 'Unknown').trim() || 'Unknown'
}

function applyFileUpdate(updated) {
  const f = files.value.find((x) => x.id === updated.id)
  if (f) {
    f.user_id = updated.user_id
    f.user = updated.user
    f.name = updated.name
    f.path = updated.path
    f.content = updated.content
    f.verified = updated.verified
    f.updated_at = updated.updated_at
  }
  if (previewFile.value?.id === updated.id) {
    previewFile.value = { ...previewFile.value, ...updated }
  }
}

async function transferOwner(file, userId) {
  if (!isAdmin.value || userId === file.user_id) return
  try {
    const updated = await api(`/files/${file.id}`, {
      method: 'PUT',
      body: JSON.stringify({ user_id: userId }),
    })
    applyFileUpdate(updated)
  } catch (e) {
    alert(e.message)
  }
}

async function batchTransfer() {
  const userId = Number(batchTransferUserId.value)
  if (!userId) return
  const ids = [...selectedFileIds.value]
  if (!ids.length) return
  try {
    const results = await Promise.all(
      ids.map((id) =>
        api(`/files/${id}`, { method: 'PUT', body: JSON.stringify({ user_id: userId }) })
      )
    )
    for (const updated of results) {
      applyFileUpdate(updated)
    }
  } catch (e) {
    alert(e.message)
  }
  selectedFileIds.value = new Set()
  batchTransferUserId.value = ''
}

async function verifyFromPreview() {
  if (!previewFile.value || !isAdmin.value) return
  try {
    const updated = await api(`/files/${previewFile.value.id}`, {
      method: 'PUT',
      body: JSON.stringify({ verified: !previewFile.value.verified }),
    })
    previewFile.value.verified = updated.verified
    const f = files.value.find((x) => x.id === previewFile.value.id)
    if (f) f.verified = updated.verified
  } catch (e) {
    alert(e.message)
  }
}

async function load({ silent = false } = {}) {
  if (!silent) loading.value = true
  else refreshing.value = true
  try {
    const requests = [api('/files')]
    if (isAdmin.value) {
      requests.push(api('/users'), api('/files/presence'))
    }
    const results = await Promise.all(requests)
    const filesData = results[0]
    files.value = filesData
    if (isAdmin.value) {
      allUsers.value = results[1] || []
      presenceMap.value = results[2] || {}
    }
    if (previewFile.value) {
      const updated = filesData.find((f) => f.id === previewFile.value.id)
      if (updated) previewFile.value = updated
    }
  } catch (e) {
    console.error(e)
  } finally {
    if (!silent) loading.value = false
    else refreshing.value = false
  }
}

function presenceLabel(fileId) {
  const session = presenceMap.value[String(fileId)]
  if (!session) return ''
  return (session.fullname || session.email || 'Owner').trim()
}

function isFileOpened(fileId) {
  return Boolean(presenceMap.value[String(fileId)])
}

function startListPolling() {
  stopListPolling()
  listPollTimer = setInterval(() => {
    if (route.name === 'files') {
      load({ silent: true })
    }
  }, LIST_POLL_MS)
}

function stopListPolling() {
  if (listPollTimer) {
    clearInterval(listPollTimer)
    listPollTimer = null
  }
}

function onVisibilityChange() {
  if (document.visibilityState === 'visible' && route.name === 'files') {
    load({ silent: true })
  }
}

function formatDate(dateStr) {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleString()
}

function exportFile(file) {
  const lang = detectLanguage(file.name)
  const blob = new Blob([file.content || ''], { type: exportMimeType(lang) })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = file.name || defaultFileName(lang)
  a.click()
  URL.revokeObjectURL(url)
}

async function deleteFile(file, e) {
  e?.preventDefault()
  if (!confirm('Delete this file?')) return
  try {
    await api(`/files/${file.id}`, { method: 'DELETE' })
    files.value = files.value.filter((f) => f.id !== file.id)
  } catch (e) {
    alert(e.message)
  }
}

function openFile(file) {
  router.push({ name: 'playground', params: { id: file.id } })
}

function watchFile(file) {
  router.push({ name: 'playground', params: { id: file.id }, query: { watch: '1' } })
}

async function toggleVerified(file) {
  if (!isAdmin.value) return
  try {
    const updated = await api(`/files/${file.id}`, {
      method: 'PUT',
      body: JSON.stringify({ verified: !file.verified }),
    })
    file.verified = updated.verified
  } catch (e) {
    alert(e.message)
  }
}

function startRename(file) {
  editingFileId.value = file.id
  editingName.value = file.name || ''
}

function cancelRename() {
  editingFileId.value = null
  editingName.value = ''
}

async function saveRename() {
  const id = editingFileId.value
  if (!id) return
  const file = files.value.find((x) => x.id === id)
  const name = preserveExtension(editingName.value, file?.name || defaultFileName('go'))
  try {
    const updated = await api(`/files/${id}`, {
      method: 'PUT',
      body: JSON.stringify({ name }),
    })
    const f = files.value.find((x) => x.id === id)
    if (f) f.name = updated.name
  } catch (e) {
    alert(e.message)
  }
  cancelRename()
}

function triggerImport() {
  importInputRef.value?.click()
}

async function onImport(e) {
  const input = e.target
  const file = input.files?.[0]
  if (!file) return
  input.value = ''
  try {
    const content = await file.text()
    const name = file.name || defaultFileName(detectLanguage(file.name))
    const created = await api('/files', {
      method: 'POST',
      body: JSON.stringify({
        name,
        path: '',
        content: content || defaultTemplate(detectLanguage(name)),
      }),
    })
    files.value = [created, ...files.value]
    router.push({ name: 'playground', params: { id: created.id } })
  } catch (err) {
    alert(err.message)
  }
}

onMounted(() => {
  initFromQuery()
  load()
  startListPolling()
  document.addEventListener('visibilitychange', onVisibilityChange)
})

onBeforeUnmount(() => {
  stopListPolling()
  document.removeEventListener('visibilitychange', onVisibilityChange)
})

watch(
  () => route.name,
  (name) => {
    if (name === 'files') {
      load()
    }
  }
)
</script>

<template>
  <div class="flex min-h-screen flex-col bg-slate-50">
    <AppHeader>
      <router-link
        v-if="isAdmin"
        to="/users"
        class="rounded border border-slate-300 bg-white px-3 py-1.5 text-xs font-medium text-slate-700 hover:bg-slate-50"
      >
        Users
      </router-link>
      <router-link
        v-if="isAdmin"
        to="/files/new?lang=markdown"
        class="rounded border border-violet-300 bg-violet-50 px-3 py-1.5 text-xs font-medium text-violet-800 hover:bg-violet-100"
      >
        New instruction
      </router-link>
      <router-link
        to="/files/new"
        class="rounded bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700"
      >
        New file
      </router-link>
      <button
        type="button"
        class="rounded border border-slate-300 bg-white px-3 py-1.5 text-xs font-medium text-slate-700 hover:bg-slate-50 disabled:opacity-50"
        :disabled="loading || refreshing"
        @click="load()"
      >
        {{ refreshing ? 'Refreshing...' : 'Refresh' }}
      </button>
      <button
        type="button"
        class="rounded border border-slate-300 bg-white px-3 py-1.5 text-xs font-medium text-slate-700 hover:bg-slate-50"
        @click="triggerImport"
      >
        Import
      </button>
      <input
        ref="importInputRef"
        type="file"
        :accept="importAccept"
        class="hidden"
        @change="onImport"
      />
    </AppHeader>

    <main class="flex-1 px-2 py-4">
      <div class="rounded-lg border border-slate-200 bg-white shadow-sm">
        <div class="border-b border-slate-200 px-4 py-2">
          <h2 class="text-sm font-medium text-slate-800">Your files</h2>
        </div>

        <div
          v-if="isAdmin && !loading"
          class="flex flex-wrap items-center gap-2 border-b border-slate-200 bg-slate-50 px-4 py-2"
        >
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search by name or student..."
            class="min-w-[160px] rounded border border-slate-300 px-2 py-1 text-xs"
          />
          <div class="flex gap-0.5">
            <button
              type="button"
              class="rounded px-2 py-0.5 text-xs"
              :class="typeFilter === 'all' ? 'bg-slate-200 text-slate-800' : 'text-slate-600 hover:bg-slate-100'"
              @click="typeFilter = 'all'"
            >
              All types
            </button>
            <button
              type="button"
              class="rounded px-2 py-0.5 text-xs"
              :class="typeFilter === 'code' ? 'bg-slate-200 text-slate-800' : 'text-slate-600 hover:bg-slate-100'"
              @click="typeFilter = 'code'"
            >
              Code
            </button>
            <button
              type="button"
              class="rounded px-2 py-0.5 text-xs"
              :class="typeFilter === 'instructions' ? 'bg-slate-200 text-slate-800' : 'text-slate-600 hover:bg-slate-100'"
              @click="typeFilter = 'instructions'"
            >
              Instructions
            </button>
          </div>
          <div class="flex gap-0.5">
            <button
              type="button"
              class="rounded px-2 py-0.5 text-xs"
              :class="verifiedFilter === 'all' ? 'bg-slate-200 text-slate-800' : 'text-slate-600 hover:bg-slate-100'"
              @click="verifiedFilter = 'all'"
            >
              All
            </button>
            <button
              type="button"
              class="rounded px-2 py-0.5 text-xs"
              :class="verifiedFilter === 'verified' ? 'bg-slate-200 text-slate-800' : 'text-slate-600 hover:bg-slate-100'"
              @click="verifiedFilter = 'verified'"
            >
              Verified
            </button>
            <button
              type="button"
              class="rounded px-2 py-0.5 text-xs"
              :class="verifiedFilter === 'unverified' ? 'bg-slate-200 text-slate-800' : 'text-slate-600 hover:bg-slate-100'"
              @click="verifiedFilter = 'unverified'"
            >
              Not verified
            </button>
          </div>
          <select
            v-model="userFilter"
            class="rounded border border-slate-300 px-2 py-1 text-xs"
          >
            <option value="">All users</option>
            <option
              v-for="u in availableUsers"
              :key="u.email"
              :value="u.email"
            >
              {{ u.fullname || u.email }}
            </option>
          </select>
          <select
            v-model="groupBy"
            class="rounded border border-slate-300 px-2 py-1 text-xs"
          >
            <option value="author">Group by author</option>
            <option value="verified">Group by verified</option>
            <option value="none">No grouping</option>
          </select>
        </div>

        <div v-if="selectedFileIds.size > 0 && isAdmin" class="flex items-center gap-2 border-b border-amber-200 bg-amber-50 px-4 py-1.5">
          <span class="text-xs font-medium text-amber-800">Selected: {{ selectedFileIds.size }} files</span>
          <button
            type="button"
            class="rounded bg-green-600 px-2 py-0.5 text-xs text-white hover:bg-green-700"
            @click="batchVerify(true)"
          >
            Verify selected
          </button>
          <button
            type="button"
            class="rounded border border-slate-400 px-2 py-0.5 text-xs text-slate-700 hover:bg-slate-100"
            @click="batchVerify(false)"
          >
            Unverify selected
          </button>
          <select
            v-model="batchTransferUserId"
            class="rounded border border-slate-300 px-2 py-0.5 text-xs text-slate-700"
          >
            <option value="">Transfer to...</option>
            <option v-for="u in allUsers" :key="u.id" :value="u.id">
              {{ userOptionLabel(u) }}
            </option>
          </select>
          <button
            type="button"
            class="rounded bg-blue-600 px-2 py-0.5 text-xs text-white hover:bg-blue-700 disabled:opacity-50"
            :disabled="!batchTransferUserId"
            @click="batchTransfer"
          >
            Transfer selected
          </button>
          <button
            type="button"
            class="rounded bg-red-600 px-2 py-0.5 text-xs text-white hover:bg-red-700"
            @click="batchDelete"
          >
            Delete selected
          </button>
          <button
            type="button"
            class="rounded px-2 py-0.5 text-xs text-slate-600 hover:bg-slate-100"
            @click="selectedFileIds = new Set()"
          >
            Clear
          </button>
        </div>

        <div v-if="loading" class="px-4 py-8 text-center text-sm text-slate-500">
          Loading...
        </div>

        <template v-else>
          <template v-for="(group, idx) in groupedFiles" :key="group.label || `group-${idx}`">
            <div v-if="isAdmin && group.label" class="border-b border-slate-100 bg-slate-50 px-4 py-1">
              <span class="text-xs font-medium" :class="group.isOwn ? 'text-slate-700' : 'text-slate-600'">
                {{ group.label }}
              </span>
            </div>
            <div class="overflow-x-auto">
            <table class="w-full table-fixed min-w-[720px]">
              <colgroup>
                <col v-if="isAdmin" style="width: 28px" />
                <col style="width: 22%" />
                <col v-if="isAdmin" style="width: 14%" />
                <col v-if="isAdmin" style="width: 12%" />
                <col style="width: 36px" />
                <col style="width: 18%" />
                <col style="width: 260px" />
              </colgroup>
              <thead>
                <tr class="border-b border-slate-200 text-left text-xs text-slate-500">
                  <th v-if="isAdmin" class="w-7 py-2">
                    <input
                      v-if="group.files.length"
                      type="checkbox"
                      :checked="isGroupAllSelected(group)"
                      class="mx-auto block h-3.5 w-3.5 rounded"
                      :title="isGroupAllSelected(group) ? 'Deselect all' : 'Select all'"
                      @change="toggleSelectAllInGroup(group)"
                    />
                  </th>
                  <th class="px-4 py-2 font-medium">
                    <template v-if="isAdmin">
                      <button
                        type="button"
                        class="flex items-center gap-0.5 hover:text-slate-700"
                        @click="setSort('name')"
                      >
                        Name
                        <span v-if="sortBy === 'name'" class="text-slate-400">{{ sortAsc ? '↑' : '↓' }}</span>
                      </button>
                    </template>
                    <span v-else>Name</span>
                  </th>
                  <th v-if="isAdmin" class="px-4 py-2 font-medium">Author</th>
                  <th v-if="isAdmin" class="px-4 py-2 font-medium">Instruction</th>
                  <th class="px-1 py-2 text-center font-medium">
                    <template v-if="isAdmin">
                      <button
                        type="button"
                        class="mx-auto flex items-center justify-center gap-0.5 hover:text-slate-700"
                        @click="setSort('verified')"
                      >
                        ✓
                        <span v-if="sortBy === 'verified'" class="text-slate-400">{{ sortAsc ? '↑' : '↓' }}</span>
                      </button>
                    </template>
                    <span v-else>✓</span>
                  </th>
                  <th class="px-4 py-2 font-medium">
                    <template v-if="isAdmin">
                      <button
                        type="button"
                        class="flex items-center gap-0.5 hover:text-slate-700"
                        @click="setSort('updated_at')"
                      >
                        Updated
                        <span v-if="sortBy === 'updated_at'" class="text-slate-400">{{ sortAsc ? '↑' : '↓' }}</span>
                      </button>
                    </template>
                    <span v-else>Updated</span>
                  </th>
                  <th class="px-4 py-2 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="file in group.files"
                  :key="file.id"
                  class="border-b border-slate-100 hover:bg-slate-50"
                >
                  <td v-if="isAdmin" class="w-7 py-2">
                    <input
                      type="checkbox"
                      :checked="selectedFileIds.has(file.id)"
                      class="mx-auto block h-3.5 w-3.5 rounded"
                      @change="toggleSelect(file)"
                    />
                  </td>
                  <td class="px-4 py-2">
                    <div class="flex min-w-0 items-center gap-1">
                      <template v-if="editingFileId === file.id">
                        <input
                          v-model="editingName"
                          type="text"
                          class="min-w-0 flex-1 rounded border border-slate-300 px-1.5 py-0.5 text-xs"
                          @keydown.enter="saveRename"
                          @keydown.esc="cancelRename"
                          @blur="saveRename"
                        />
                      </template>
                      <template v-else>
                        <button
                          type="button"
                          class="flex min-w-0 items-center gap-1 truncate text-left text-xs font-medium text-blue-600 hover:underline"
                          @click="openFile(file)"
                        >
                          <FileTypeIcon :filename="file.name" />
                          <span class="truncate">{{ file.name }}</span>
                        </button>
                        <span
                          v-if="isAdmin && isCodeFile(file.name) && file.instructions_file_id"
                          class="shrink-0 rounded bg-violet-100 px-1 py-0.5 text-[10px] font-medium text-violet-700"
                          :title="instructionName(file.instructions_file_id)"
                        >
                          Instructions
                        </span>
                        <span
                          v-if="isAdmin && file.autosave_enabled === false"
                          class="shrink-0 rounded bg-amber-100 px-1 py-0.5 text-[10px] font-medium text-amber-700"
                        >
                          Autosave off
                        </span>
                        <span
                          v-if="isAdmin && isFileOpened(file.id)"
                          class="shrink-0 rounded bg-green-100 px-1 py-0.5 text-[10px] font-medium text-green-700"
                          :title="`Opened by ${presenceLabel(file.id)}`"
                        >
                          Opened
                        </span>
                        <button
                          type="button"
                          class="shrink-0 rounded p-0.5 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
                          title="Rename"
                          @click.stop="startRename(file)"
                        >
                          <PencilSquareIcon class="h-3.5 w-3.5" />
                        </button>
                      </template>
                    </div>
                  </td>
                  <td v-if="isAdmin" class="truncate px-4 py-2 text-xs text-slate-600">
                    <select
                      :value="file.user_id"
                      class="max-w-full truncate rounded border border-slate-300 px-1 py-0.5 text-xs"
                      @change="transferOwner(file, Number($event.target.value))"
                    >
                      <option v-for="u in allUsers" :key="u.id" :value="u.id">
                        {{ userOptionLabel(u) }}
                      </option>
                    </select>
                  </td>
                  <td v-if="isAdmin" class="px-4 py-2">
                    <select
                      v-if="isCodeFile(file.name)"
                      :value="file.instructions_file_id ?? ''"
                      class="max-w-full truncate rounded border border-slate-300 px-1 py-0.5 text-xs"
                      @change="linkInstructions(file, $event)"
                    >
                      <option value="">None</option>
                      <option v-for="md in markdownFiles" :key="md.id" :value="md.id">
                        {{ md.name }}
                      </option>
                    </select>
                    <span v-else class="text-xs text-slate-400">—</span>
                  </td>
                  <td class="px-1 py-2 text-center align-middle">
                    <template v-if="isAdmin">
                      <button
                        type="button"
                        class="inline-flex h-4 w-4 items-center justify-center rounded border transition-colors"
                        :class="file.verified ? 'border-green-500 bg-green-50 text-green-600 hover:bg-green-100' : 'border-slate-300 hover:border-slate-400 hover:bg-slate-50'"
                        :title="file.verified ? 'Unverify' : 'Verify'"
                        @click="toggleVerified(file)"
                      >
                        <CheckIcon v-if="file.verified" class="h-3 w-3" stroke-width="3" />
                      </button>
                    </template>
                    <CheckIcon v-else-if="file.verified" class="mx-auto h-4 w-4 text-green-600" stroke-width="2.5" />
                  </td>
                  <td class="truncate px-4 py-2 text-xs text-slate-500">{{ formatDate(file.updated_at) }}</td>
                  <td class="whitespace-nowrap px-2 py-2">
                    <div class="flex flex-nowrap items-center gap-1">
                      <button
                        v-if="isAdmin"
                        type="button"
                        class="inline-flex shrink-0 items-center gap-0.5 rounded px-1.5 py-0.5 text-xs text-slate-600 hover:bg-slate-100"
                        @click="previewFile = file"
                      >
                        <EyeIcon class="h-3.5 w-3.5" />
                        Preview
                      </button>
                      <button
                        v-if="isAdmin"
                        type="button"
                        class="inline-flex shrink-0 items-center gap-0.5 rounded px-1.5 py-0.5 text-xs text-violet-700 hover:bg-violet-50"
                        @click="watchFile(file)"
                      >
                        <SignalIcon class="h-3.5 w-3.5" />
                        Watch
                      </button>
                      <button
                        type="button"
                        class="inline-flex shrink-0 items-center gap-0.5 rounded px-1.5 py-0.5 text-xs text-slate-600 hover:bg-slate-100"
                        @click="exportFile(file)"
                      >
                        <ArrowDownTrayIcon class="h-3.5 w-3.5" />
                        Export
                      </button>
                      <button
                        type="button"
                        class="inline-flex shrink-0 items-center gap-0.5 rounded px-1.5 py-0.5 text-xs text-red-600 hover:bg-red-50"
                        @click="deleteFile(file, $event)"
                      >
                        <TrashIcon class="h-3.5 w-3.5" />
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
            </div>
          </template>

          <div v-if="files.length === 0 && !loading" class="px-4 py-8 text-center text-sm text-slate-500">
            No files yet. Create one or import from your computer.
          </div>
        </template>
      </div>
    </main>

    <AppFooter />

    <Teleport to="body">
      <div
        v-if="previewFile"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
        @click.self="previewFile = null"
      >
        <div class="flex max-h-[90vh] w-full max-w-3xl flex-col rounded-lg border border-slate-200 bg-white shadow-xl">
          <div class="flex items-center justify-between border-b border-slate-200 px-4 py-2">
            <div class="flex items-center gap-2 text-slate-800">
              <FileTypeIcon :filename="previewFile.name" />
              <span class="text-sm font-medium">{{ previewFile.name }}</span>
              <span
                v-if="previewFile.verified"
                class="rounded bg-green-100 px-1.5 py-0.5 text-[10px] font-medium text-green-700"
              >
                Verified
              </span>
            </div>
            <button
              type="button"
              class="rounded p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
              @click="previewFile = null"
            >
              <XMarkIcon class="h-5 w-5" />
            </button>
          </div>
          <div class="flex-1 min-h-0 overflow-auto p-4">
            <pre class="font-mono text-xs leading-relaxed text-slate-800 whitespace-pre-wrap break-words">{{ previewFile.content || '// Empty' }}</pre>
          </div>
          <div class="flex gap-2 border-t border-slate-200 px-4 py-2">
            <button
              type="button"
              class="rounded bg-blue-600 px-3 py-1 text-xs font-medium text-white hover:bg-blue-700"
              @click="openFile(previewFile); previewFile = null"
            >
              Open in editor
            </button>
            <button
              v-if="isAdmin"
              type="button"
              class="rounded border border-slate-300 px-3 py-1 text-xs font-medium text-slate-700 hover:bg-slate-50"
              @click="verifyFromPreview"
            >
              {{ previewFile.verified ? 'Unverify' : 'Verify' }}
            </button>
            <button
              type="button"
              class="rounded border border-slate-300 px-3 py-1 text-xs font-medium text-slate-700 hover:bg-slate-50"
              @click="previewFile = null"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
