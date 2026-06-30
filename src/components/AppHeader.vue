<script setup>
import { useAuth } from '../composables/useAuth'
import AppLogo from './AppLogo.vue'

const { logout } = useAuth()
</script>

<template>
  <header class="border-b border-slate-200 bg-white px-4 py-2 shadow-sm">
    <div class="flex w-full items-center justify-between">
      <div class="flex items-center gap-3">
        <router-link to="/files" class="shrink-0">
          <AppLogo size="sm" />
        </router-link>
        <slot name="left" />
      </div>
      <div class="flex items-center gap-2">
        <slot />
        <button
          type="button"
          class="rounded border border-slate-300 px-3 py-1.5 text-xs font-medium text-slate-700 hover:bg-slate-50"
          @click="logout"
        >
          Logout
        </button>
      </div>
    </div>
  </header>
</template>
