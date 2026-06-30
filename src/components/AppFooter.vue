<script setup>
// __APP_VERSION__ is injected by Vite at build time from package.json
const version = __APP_VERSION__
</script>

<template>
  <footer class="border-t border-slate-200 bg-white px-4 py-2">
    <p class="text-right text-xs text-slate-400">
      (c) 2026 Go Playground · v{{ version }}
    </p>
  </footer>
</template>
