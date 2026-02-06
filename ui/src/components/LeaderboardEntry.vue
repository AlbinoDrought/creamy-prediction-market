<script setup lang="ts">
import { computed } from 'vue'
import type { LeaderboardUser } from '@/types/users'
import { useAchievementsStore } from '@/stores/achievements'

const props = defineProps<{
  user: LeaderboardUser
  isCurrentUser: boolean
}>()

const achievementsStore = useAchievementsStore()

const achievementIcons = computed(() => {
  if (!props.user.achievements || props.user.achievements.length === 0) return []
  // Get icons for each achievement, limit to 6 most recent
  return props.user.achievements
    .slice(-6)
    .map(id => achievementsStore.getAchievementById(id)?.icon)
    .filter(Boolean)
})

const medalEmoji = computed(() => {
  switch (props.user.rank) {
    case 1: return '1st'
    case 2: return '2nd'
    case 3: return '3rd'
    default: return null
  }
})

const rankClass = computed(() => {
  switch (props.user.rank) {
    case 1: return 'text-yellow-400'
    case 2: return 'text-gray-300'
    case 3: return 'text-amber-600'
    default: return 'text-gray-500'
  }
})
</script>

<template>
  <div
    class="flex items-center gap-4 p-4 rounded-xl transition-colors"
    :class="isCurrentUser ? 'bg-primary/10 border border-primary/30' : 'bg-dark-light'"
  >
    <!-- Rank -->
    <div class="w-10 text-center">
      <span v-if="medalEmoji" class="text-lg font-bold" :class="rankClass">
        {{ medalEmoji }}
      </span>
      <span v-else class="text-lg text-gray-500">
        {{ user.rank }}
      </span>
    </div>

    <!-- Avatar & Name -->
    <div class="flex-1 flex items-center gap-3">
      <div
        class="w-10 h-10 rounded-full flex items-center justify-center font-bold text-dark"
        :class="user.rank <= 3
          ? 'bg-gradient-to-br from-primary to-secondary'
          : 'bg-dark-lighter text-gray-400'"
      >
        {{ user.name.charAt(0).toUpperCase() }}
      </div>
      <div>
        <p class="font-medium" :class="isCurrentUser ? 'text-primary' : 'text-white'">
          {{ user.name }}
          <span v-if="isCurrentUser" class="text-xs text-primary/70 ml-1">(you)</span>
        </p>
        <p v-if="achievementIcons.length > 0" class="text-sm mt-0.5">
          <span v-for="(icon, index) in achievementIcons" :key="index" class="mr-0.5">{{ icon }}</span>
        </p>
      </div>
    </div>

    <!-- Tokens -->
    <div class="text-right">
      <p class="font-bold" :class="isCurrentUser ? 'text-primary' : 'text-white'">
        {{ user.tokens }}
      </p>
      <p class="text-xs text-gray-400">tokens</p>
    </div>
  </div>
</template>
