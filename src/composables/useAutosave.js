import { ref, watch, onMounted, onBeforeUnmount, unref } from 'vue'
import { api } from './useApi'

const DEBOUNCE_MS = 1500

export function useAutosave({ fileId, content, enabled, loading, active }) {
  const saveStatus = ref('saved')
  const lastSavedContent = ref('')

  let debounceTimer = null
  let saving = false

  function isDirty() {
    return content.value !== lastSavedContent.value
  }

  function syncBaseline(value) {
    lastSavedContent.value = value ?? content.value
    saveStatus.value = 'saved'
  }

  async function persist() {
    const id = unref(fileId)
    if (!id || unref(loading) || !unref(active)) return
    if (!isDirty()) {
      saveStatus.value = 'saved'
      return
    }

    saving = true
    saveStatus.value = 'saving'
    try {
      await api(`/files/${id}`, {
        method: 'PUT',
        body: JSON.stringify({ content: content.value }),
      })
      lastSavedContent.value = content.value
      saveStatus.value = 'saved'
    } catch {
      saveStatus.value = 'error'
    } finally {
      saving = false
    }
  }

  function scheduleSave() {
    if (debounceTimer) clearTimeout(debounceTimer)
    if (!unref(active) || unref(loading)) return
    if (!unref(enabled)) {
      saveStatus.value = isDirty() ? 'unsaved' : 'saved'
      return
    }
    if (!isDirty()) {
      saveStatus.value = 'saved'
      return
    }
    saveStatus.value = 'unsaved'
    debounceTimer = setTimeout(() => {
      debounceTimer = null
      if (unref(enabled) && unref(active)) persist()
    }, DEBOUNCE_MS)
  }

  async function saveNow() {
    if (debounceTimer) {
      clearTimeout(debounceTimer)
      debounceTimer = null
    }
    if (!unref(active) || unref(loading)) return
    if (!isDirty()) {
      saveStatus.value = 'saved'
      return
    }
    await persist()
  }

  function onBeforeUnload(e) {
    if (!unref(active) || unref(loading)) return
    if (isDirty() && (saveStatus.value === 'unsaved' || saveStatus.value === 'error')) {
      e.preventDefault()
      e.returnValue = ''
    }
  }

  watch(content, scheduleSave)

  watch(enabled, () => {
    if (!unref(enabled) && isDirty()) {
      saveStatus.value = 'unsaved'
    } else if (unref(enabled) && isDirty()) {
      scheduleSave()
    }
  })

  onMounted(() => {
    window.addEventListener('beforeunload', onBeforeUnload)
  })

  onBeforeUnmount(() => {
    if (debounceTimer) clearTimeout(debounceTimer)
    window.removeEventListener('beforeunload', onBeforeUnload)
  })

  return {
    saveStatus,
    syncBaseline,
    saveNow,
    isDirty,
  }
}
