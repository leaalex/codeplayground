import { ref, watch, unref, onBeforeUnmount } from 'vue'

const SEARCH_DEBOUNCE_MS = 400

const VERIFIED_VALUES = new Set(['all', 'verified', 'unverified'])
const SORT_VALUES = new Set(['name', 'verified', 'updated_at'])
const GROUP_VALUES = new Set(['author', 'verified', 'none'])
const TYPE_VALUES = new Set(['all', 'code', 'instructions'])

function parseVerified(raw) {
  const v = raw?.toString() || 'all'
  return VERIFIED_VALUES.has(v) ? v : 'all'
}

function parseSort(raw) {
  const v = raw?.toString() || 'updated_at'
  return SORT_VALUES.has(v) ? v : 'updated_at'
}

function parseGroup(raw) {
  const v = raw?.toString() || 'author'
  return GROUP_VALUES.has(v) ? v : 'author'
}

function parseType(raw) {
  const v = raw?.toString() || 'all'
  return TYPE_VALUES.has(v) ? v : 'all'
}

function normalizeQuery(query) {
  const out = {}
  for (const [k, v] of Object.entries(query)) {
    out[k] = Array.isArray(v) ? v[0] : v
  }
  return out
}

function queriesEqual(a, b) {
  const keys = new Set(Object.keys(a).concat(Object.keys(b)))
  for (const k of keys) {
    const av = a[k] != null ? a[k] : ''
    const bv = b[k] != null ? b[k] : ''
    if (String(av) !== String(bv)) return false
  }
  return true
}

export function useFilesFilters(route, router, isAdmin) {
  const searchQuery = ref('')
  const verifiedFilter = ref('all')
  const userFilter = ref('')
  const sortBy = ref('updated_at')
  const sortAsc = ref(false)
  const groupBy = ref('author')
  const typeFilter = ref('all')

  let syncing = false
  let searchDebounceTimer = null

  function initFromQuery() {
    syncing = true
    const q = route.query

    if (unref(isAdmin)) {
      searchQuery.value = q.q?.toString() || ''
      verifiedFilter.value = parseVerified(q.verified)
      userFilter.value = q.user?.toString() || ''
      groupBy.value = parseGroup(q.group)
      typeFilter.value = parseType(q.type)
    } else {
      searchQuery.value = ''
      verifiedFilter.value = 'all'
      userFilter.value = ''
      groupBy.value = 'author'
      typeFilter.value = 'all'
    }

    sortBy.value = parseSort(q.sort)
    sortAsc.value = q.asc === '1'

    syncing = false
  }

  function buildQuery() {
    const query = {}

    if (unref(isAdmin)) {
      const q = searchQuery.value.trim()
      if (q) query.q = q
      if (verifiedFilter.value !== 'all') query.verified = verifiedFilter.value
      if (userFilter.value) query.user = userFilter.value
      if (groupBy.value !== 'author') query.group = groupBy.value
      if (typeFilter.value !== 'all') query.type = typeFilter.value
    }

    if (sortBy.value !== 'updated_at') query.sort = sortBy.value
    if (sortAsc.value) query.asc = '1'

    return query
  }

  function syncToQuery() {
    if (syncing) return
    const next = buildQuery()
    const current = normalizeQuery(route.query)
    if (queriesEqual(next, current)) return
    router.replace({ query: next })
  }

  function scheduleSearchSync() {
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
    searchDebounceTimer = setTimeout(() => {
      searchDebounceTimer = null
      syncToQuery()
    }, SEARCH_DEBOUNCE_MS)
  }

  watch(
    () => route.query,
    () => {
      initFromQuery()
    }
  )

  watch([verifiedFilter, userFilter, sortBy, sortAsc, groupBy, typeFilter], () => {
    syncToQuery()
  })

  watch(searchQuery, () => {
    scheduleSearchSync()
  })

  onBeforeUnmount(() => {
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
  })

  return {
    searchQuery,
    verifiedFilter,
    userFilter,
    sortBy,
    sortAsc,
    groupBy,
    typeFilter,
    initFromQuery,
  }
}
