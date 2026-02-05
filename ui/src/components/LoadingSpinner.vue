<script setup lang="ts">
defineProps<{
  size?: 'sm' | 'md' | 'lg'
}>()
</script>

<template>
  <div class="flex items-center justify-center">
    <div
      class="animate-spin rounded-full border-2 border-primary border-t-transparent"
      :class="{
        'w-5 h-5': size === 'sm',
        'w-8 h-8': size === 'md' || !size,
        'w-12 h-12': size === 'lg',
      }"
    ></div>
  </div>
</template>
