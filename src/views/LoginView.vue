<script setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuth } from '../composables/useAuth'
import { api } from '../composables/useApi'
import AppLogo from '../components/AppLogo.vue'

const route = useRoute()
const router = useRouter()
const { setToken } = useAuth()

const mode = ref('login')
const fullname = ref('')
const email = ref('')
const password = ref('')
const error = ref('')

function redirectAfterAuth() {
  const raw = route.query.redirect
  if (typeof raw === 'string' && raw.startsWith('/') && !raw.startsWith('//') && raw !== '/login') {
    return raw
  }
  return '/files'
}

async function submit() {
  error.value = ''
  try {
    const path = mode.value === 'login' ? '/auth/login' : '/auth/register'
    const body = mode.value === 'login'
      ? { email: email.value, password: password.value }
      : { fullname: fullname.value, email: email.value, password: password.value }
    const data = await api(path, {
      method: 'POST',
      body: JSON.stringify(body),
    })
    setToken(data.accessToken, data)
    router.push(redirectAfterAuth())
  } catch (e) {
    error.value = e.message
  }
}
</script>

<template>
  <div class="flex min-h-screen items-center justify-center bg-slate-50 p-4">
    <div class="w-full max-w-sm rounded-lg border border-slate-200 bg-white p-6 shadow-lg text-center">
      <AppLogo size="md" centered />
      <p class="mt-0.5 text-xs text-slate-500">Sign in to save your files</p>

      <form class="mt-4 flex flex-col gap-3" @submit.prevent="submit">
        <input
          v-if="mode === 'register'"
          v-model="fullname"
          type="text"
          placeholder="Full name"
          required
          class="w-full rounded border border-slate-300 px-3 py-2 text-sm text-slate-800 placeholder-slate-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
        <input
          v-model="email"
          type="email"
          placeholder="Email"
          required
          class="w-full rounded border border-slate-300 px-3 py-2 text-sm text-slate-800 placeholder-slate-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
        <input
          v-model="password"
          type="password"
          :placeholder="mode === 'register' ? 'Password (min 6)' : 'Password'"
          required
          :minlength="mode === 'register' ? 6 : undefined"
          class="w-full rounded border border-slate-300 px-3 py-2 text-sm text-slate-800 placeholder-slate-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
        <p v-if="error" class="text-xs text-red-600">{{ error }}</p>
        <button
          type="submit"
          class="w-full rounded bg-blue-600 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          {{ mode === 'login' ? 'Log in' : 'Register' }}
        </button>
      </form>

      <button
        type="button"
        class="mt-3 w-full text-center text-xs text-slate-500 underline hover:text-slate-700"
        @click="mode = mode === 'login' ? 'register' : 'login'"
      >
        {{ mode === 'login' ? "Don't have an account? Register" : 'Already have an account? Log in' }}
      </button>
    </div>
  </div>
</template>
