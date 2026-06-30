<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../composables/useApi'

const router = useRouter()
const name = ref('')
const path = ref('')
const error = ref('')
const creating = ref(false)

async function create() {
  const fileName = name.value.trim() || 'untitled.go'
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
        name: fileName.endsWith('.go') ? fileName : fileName + '.go',
        path: path.value.trim(),
        content: 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println("Hello!")\n}\n',
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
  <div class="min-h-screen bg-slate-50">
    <header class="border-b border-slate-200 bg-white px-4 py-2 shadow-sm">
      <div class="mx-auto flex max-w-6xl items-center justify-between">
        <router-link to="/files" class="text-sm text-blue-600 hover:underline">← Back to files</router-link>
        <h1 class="text-base font-semibold text-slate-800">Create file</h1>
        <div class="w-20" />
      </div>
    </header>

    <main class="mx-auto max-w-md px-4 py-4">
      <form
        class="rounded-lg border border-slate-200 bg-white p-4 shadow-sm"
        @submit.prevent="create"
      >
        <div class="space-y-3">
          <div>
            <label for="name" class="mb-0.5 block text-xs font-medium text-slate-700">Name</label>
            <input
              id="name"
              v-model="name"
              type="text"
              placeholder="main.go"
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
            {{ creating ? 'Creating...' : 'Create' }}
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
  </div>
</template>
