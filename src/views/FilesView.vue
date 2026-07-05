<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { PencilSquareIcon, CheckIcon, EyeIcon, XMarkIcon } from '@heroicons/vue/24/outline'
import AppHeader from '../components/AppHeader.vue'
import AppFooter from '../components/AppFooter.vue'
import { useAuth } from '../composables/useAuth'
import { api } from '../composables/useApi'
import {
  detectLanguage,
  defaultFileName,
  defaultTemplate,
  preserveExtension,
  exportMimeType,
  languageLabel,
} from '../utils/language'

const router = useRouter()
const { isAdmin, user } = useAuth()

const files = ref([])
const loading = ref(false)
const importInputRef = ref(null)
const editingFileId = ref(null)
const editingName = ref('')

const searchQuery = ref('')
const verifiedFilter = ref('all')
const userFilter = ref('')
const sortBy = ref('updated_at')
const sortAsc = ref(false)
const groupBy = ref('author')
const selectedFileIds = ref(new Set())
const previewFile = ref(null)

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

async function load() {
  loading.value = true
  try {
    files.value = await api('/files')
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
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

onMounted(load)
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
        to="/files/new"
        class="rounded bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700"
      >
        New file
      </router-link>
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
        accept=".go,.py"
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
            <table class="w-full table-fixed">
              <colgroup>
                <col v-if="isAdmin" style="width: 28px" />
                <col style="width: 24%" />
                <col v-if="isAdmin" style="width: 16%" />
                <col style="width: 36px" />
                <col style="width: 20%" />
                <col style="width: 14%" />
                <col style="width: auto" />
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
                  <th class="px-4 py-2 font-medium">Path</th>
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
                          class="min-w-0 truncate text-left text-xs font-medium text-blue-600 hover:underline"
                          @click="openFile(file)"
                        >
                          {{ file.name }}
                        </button>
                        <span
                          class="shrink-0 rounded bg-slate-100 px-1 py-0.5 text-[10px] font-medium text-slate-600"
                        >
                          {{ languageLabel(detectLanguage(file.name)) }}
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
                  <td v-if="isAdmin" class="truncate px-4 py-2 text-xs text-slate-600" :title="file.user?.fullname || file.user?.email || '-'">
                    {{ file.user?.fullname || file.user?.email || '-' }}
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
                  <td class="truncate px-4 py-2 text-xs text-slate-600" :title="file.path || '-'">{{ file.path || '-' }}</td>
                  <td class="truncate px-4 py-2 text-xs text-slate-500">{{ formatDate(file.updated_at) }}</td>
                  <td class="px-4 py-2">
                    <div class="flex flex-wrap gap-1">
                      <button
                        type="button"
                        class="rounded px-1.5 py-0.5 text-xs text-blue-600 hover:bg-blue-50"
                        @click="openFile(file)"
                      >
                        Open
                      </button>
                      <button
                        v-if="isAdmin"
                        type="button"
                        class="inline-flex items-center gap-0.5 rounded px-1.5 py-0.5 text-xs text-slate-600 hover:bg-slate-100"
                        title="Preview"
                        @click="previewFile = file"
                      >
                        <EyeIcon class="h-3.5 w-3.5" />
                        Preview
                      </button>
                      <button
                        type="button"
                        class="rounded px-1.5 py-0.5 text-xs text-slate-600 hover:bg-slate-100"
                        @click="exportFile(file)"
                      >
                        Export
                      </button>
                      <button
                        type="button"
                        class="rounded px-1.5 py-0.5 text-xs text-red-600 hover:bg-red-50"
                        @click="deleteFile(file, $event)"
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
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
            <div class="flex items-center gap-2">
              <span class="text-sm font-medium text-slate-800">{{ previewFile.name }}</span>
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
