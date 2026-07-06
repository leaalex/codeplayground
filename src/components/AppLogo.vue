<script setup>
defineProps({
  size: {
    type: String,
    default: 'md',
    validator: (value) => ['sm', 'md', 'lg'].includes(value),
  },
  centered: {
    type: Boolean,
    default: false,
  },
})

const sizeClasses = {
  sm: 'h-8 max-w-[180px]',
  md: 'h-10 max-w-[220px]',
  lg: 'h-12 max-w-[260px]',
}
</script>

<template>
  <img
    src="/logo.svg"
    alt="Code Playground"
    :class="[
      sizeClasses[size],
      'w-auto shrink-0 object-contain object-left',
      centered ? 'mx-auto' : '',
    ]"
  />
</template>
