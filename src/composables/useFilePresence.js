import { watch, onBeforeUnmount, unref } from 'vue'
import { api } from './useApi'

const HEARTBEAT_MS = 20000

export function useFilePresence({ fileId, fileUserId, user, isWatchMode, loading, isCodeFile }) {
  let heartbeatTimer = null
  let activeFileId = null

  async function touch(id) {
    if (!id) return
    await api(`/files/${id}/presence`, { method: 'PUT' })
  }

  async function leave(id) {
    if (!id) return
    try {
      await api(`/files/${id}/presence`, { method: 'DELETE' })
    } catch {
      // ignore errors on unload
    }
  }

  function isOwnerActive() {
    if (unref(isCodeFile) === false) return false
    if (unref(isWatchMode) || unref(loading)) return false
    const id = unref(fileId)
    const ownerId = unref(fileUserId)
    const currentUser = unref(user)
    if (!id || ownerId == null || !currentUser) return false
    return currentUser.id === ownerId
  }

  function stopPresence() {
    if (heartbeatTimer) {
      clearInterval(heartbeatTimer)
      heartbeatTimer = null
    }
    if (activeFileId) {
      leave(activeFileId)
      activeFileId = null
    }
  }

  function startPresence(id) {
    stopPresence()
    if (!id || !isOwnerActive()) return
    activeFileId = id
    touch(id).catch(() => {})
    heartbeatTimer = setInterval(() => {
      touch(id).catch(() => {})
    }, HEARTBEAT_MS)
  }

  function syncPresence() {
    const id = unref(fileId)
    if (isOwnerActive()) {
      startPresence(id)
    } else {
      stopPresence()
    }
  }

  function onBeforeUnload() {
    if (activeFileId) {
      leave(activeFileId)
    }
  }

  watch([fileId, fileUserId, user, isWatchMode, loading, isCodeFile], () => {
    syncPresence()
  })

  onBeforeUnmount(() => {
    window.removeEventListener('beforeunload', onBeforeUnload)
    stopPresence()
  })

  window.addEventListener('beforeunload', onBeforeUnload)

  return {
    syncPresence,
    stopPresence,
  }
}
