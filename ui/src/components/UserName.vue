<script setup lang="ts">
import { computed } from 'vue'
import type { UserCosmetics } from '@/types/users'

const props = withDefaults(defineProps<{
  name: string
  cosmetics?: UserCosmetics
  isCurrentUser?: boolean
}>(), {
  isCurrentUser: false,
})

const effectClass = computed(() => {
  if (props.cosmetics?.name_effect) {
    return `effect-name-${props.cosmetics.name_effect}`
  }
  return ''
})

const fontClass = computed(() => {
  if (props.cosmetics?.name_font) {
    switch (props.cosmetics.name_font) {
      case 'serif': return 'font-serif'
      case 'mono': return 'font-mono'
      case 'cursive': return 'font-cursive'
      case 'slab': return 'font-slab'
      case 'tag': return 'font-tag'
      case 'comic': return 'font-comic'
    }
  }
  return ''
})

const fontWeightClass = computed(() => {
  return props.cosmetics?.name_bold ? 'font-black' : 'font-medium'
})
</script>

<template>
  <span
    :class="[
      effectClass,
      fontClass,
      fontWeightClass,
      isCurrentUser ? 'text-primary' : 'text-white',
    ]"
  >
    <span v-if="cosmetics?.name_emoji" class="mr-0.5">{{ cosmetics.name_emoji }}</span>
    {{ name }}
  </span>
</template>
