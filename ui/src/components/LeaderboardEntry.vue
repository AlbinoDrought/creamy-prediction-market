<script setup lang="ts">
import { computed } from 'vue'
import type { LeaderboardUser } from '@/types/users'
import { useAchievementsStore } from '@/stores/achievements'
import UserAvatar from '@/components/UserAvatar.vue'
import UserName from '@/components/UserName.vue'

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
      <UserAvatar
        :name="user.name"
        :cosmetics="user.cosmetics"
        :rank-top3="user.rank <= 3"
      />
      <div>
        <p class="font-medium">
          <UserName
            :name="user.name"
            :cosmetics="user.cosmetics"
            :is-current-user="isCurrentUser"
          />
          <span v-if="isCurrentUser" class="text-xs text-primary/70 ml-1">(you)</span>
        </p>
        <p v-if="user.cosmetics?.title" class="text-xs text-gray-400">{{ user.cosmetics.title }}</p>
        <p v-if="achievementIcons.length > 0" class="text-sm mt-0.5">
          <span v-for="(icon, index) in achievementIcons" :key="index" class="mr-0.5">{{ icon }}</span>
        </p>
      </div>
    </div>

    <!-- Score -->
    <div class="text-right">
      <p class="font-bold" :class="user.score > 0 ? 'text-success' : user.score < 0 ? 'text-error' : 'text-gray-400'">
        {{ user.score > 0 ? '+' : '' }}{{ user.score }}
      </p>
      <p class="text-xs text-gray-400">score</p>
    </div>
  </div>
</template>
