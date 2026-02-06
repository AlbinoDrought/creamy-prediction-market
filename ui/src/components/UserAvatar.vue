<script setup lang="ts">
import { computed } from 'vue'
import type { UserCosmetics } from '@/types/users'

const props = withDefaults(defineProps<{
  name: string
  cosmetics?: UserCosmetics
  size?: 'sm' | 'md'
  rankTop3?: boolean
}>(), {
  size: 'md',
  rankTop3: false,
})

const initial = computed(() => props.name.charAt(0).toUpperCase())

const gradientStyle = computed(() => {
  if (props.cosmetics?.avatar_color) {
    const [from, to] = props.cosmetics.avatar_color.split(',')
    return { background: `linear-gradient(135deg, ${from}, ${to})` }
  }
  return {}
})

const hasCustomColor = computed(() => !!props.cosmetics?.avatar_color)

const displayChar = computed(() => {
  return props.cosmetics?.avatar_emoji || initial.value
})

const effectClass = computed(() => {
  if (props.cosmetics?.avatar_effect) {
    return `effect-${props.cosmetics.avatar_effect}`
  }
  return ''
})

const sizeClass = computed(() => {
  return props.size === 'sm' ? 'w-8 h-8 text-sm' : 'w-10 h-10'
})

const hatEmoji = computed(() => props.cosmetics?.hat || '')

const hatSizeClass = computed(() => {
  return props.size === 'sm' ? 'text-xs -top-2.5' : 'text-sm -top-3'
})

const avatarItemEmoji = computed(() => props.cosmetics?.avatar_item || '')

const avatarItemSizeClass = computed(() => {
  return props.size === 'sm' ? 'text-xs' : 'text-sm'
})
</script>

<template>
  <div class="relative shrink-0" :class="hatEmoji ? (size === 'sm' ? 'mt-2.5' : 'mt-3') : ''">
    <span
      v-if="hatEmoji"
      class="absolute left-1/2 -translate-x-1/2 z-10 pointer-events-none leading-none"
      :class="hatSizeClass"
    >{{ hatEmoji }}</span>
    <div
      class="rounded-full flex items-center justify-center font-bold select-none"
      :class="[
        sizeClass,
        effectClass,
        hasCustomColor || rankTop3
          ? 'text-dark'
          : 'bg-dark-lighter text-gray-400',
        !hasCustomColor && rankTop3
          ? 'bg-gradient-to-br from-primary to-secondary'
          : '',
      ]"
      :style="hasCustomColor ? gradientStyle : {}"
    >
      {{ displayChar }}
    </div>
    <span
      v-if="avatarItemEmoji"
      class="absolute z-10 pointer-events-none leading-none -bottom-0.5 -right-1"
      :class="avatarItemSizeClass"
    >{{ avatarItemEmoji }}</span>
  </div>
</template>
