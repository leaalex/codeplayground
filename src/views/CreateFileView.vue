<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import AppHeader from '../components/AppHeader.vue'
import AppFooter from '../components/AppFooter.vue'
import { api } from '../composables/useApi'
import {
  defaultTemplate,
  defaultFileName,
  ensureExtension,
  languageLabel,
} from '../utils/language'

const router = useRouter()
const name = ref('')
const path = ref('')
const language = ref('go')
const error = ref('')
const creating = ref(false)

const namePlaceholder = computed(() =>
  language.value === 'python' ? 'main.py' : 'main.go'
)

async function create() {
  const fileName = ensureExtension(name.value.trim() || defaultFileName(language.value), language.value)
  if (!fileName) {
    error.value = 'Name is required'
    return
  }
  error.value = ''
  creating.value = true
  try {
    const file = await api('/files', {
      method: 'POST',
      body: JSON.stringify({
        name: fileName,
        path: path.value.trim(),
        content: defaultTemplate(language.value),
      }),
    })
    router.push({ name: 'playground', params: { id: file.id } })
  } catch (e) {
    error.value = e.message
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <div class="flex min-h-screen flex-col bg-slate-50">
    <AppHeader>
      <template #left>
        <router-link to="/files" class="text-sm text-blue-600 hover:underline">← Back to files</router-link>
        <h1 class="text-sm font-medium text-slate-800">Create file</h1>
      </template>
    </AppHeader>

    <main class="mx-auto w-full max-w-md flex-1 px-4 py-4">
      <form
        class="rounded-lg border border-slate-200 bg-white p-4 shadow-sm"
        @submit.prevent="create"
      >
        <div class="space-y-3">
          <div>
            <label class="mb-0.5 block text-xs font-medium text-slate-700">Language</label>
            <div class="flex gap-2">
              <label class="flex cursor-pointer items-center gap-1.5 rounded border border-slate-300 px-3 py-1.5 text-sm has-[:checked]:border-blue-500 has-[:checked]:bg-blue-50">
                <input v-model="language" type="radio" value="go" class="text-blue-600" />
                Go
              </label>
              <label class="flex cursor-pointer items-center gap-1.5 rounded border border-slate-300 px-3 py-1.5 text-sm has-[:checked]:border-blue-500 has-[:checked]:bg-blue-50">
                <input v-model="language" type="radio" value="python" class="text-blue-600" />
                Python
              </label>
            </div>
          </div>
          <div>
            <label for="name" class="mb-0.5 block text-xs font-medium text-slate-700">Name</label>
            <input
              id="name"
              v-model="name"
              type="text"
              :placeholder="namePlaceholder"
              class="w-full rounded border border-slate-300 px-3 py-1.5 text-sm text-slate-800 placeholder-slate-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>
          <div>
            <label for="path" class="mb-0.5 block text-xs font-medium text-slate-700">Path (optional)</label>
            <input
              id="path"
              v-model="path"
              type="text"
              placeholder="/"
              class="w-full rounded border border-slate-300 px-3 py-1.5 text-sm text-slate-800 placeholder-slate-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>
          <p v-if="error" class="text-xs text-red-600">{{ error }}</p>
        </div>
        <div class="mt-4 flex gap-2">
          <button
            type="submit"
            :disabled="creating"
            class="rounded bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {{ creating ? 'Creating...' : `Create ${languageLabel(language)} file` }}
          </button>
          <router-link
            to="/files"
            class="rounded border border-slate-300 px-3 py-1.5 text-xs font-medium text-slate-700 hover:bg-slate-50"
          >
            Cancel
          </router-link>
        </div>
      </form>
    </main>

    <AppFooter />
  </div>
</template>
