<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

function goToLeaderboard() {
  router.push({ name: 'leaderboard' })
}
</script>

<template>
  <header class="bg-dark px-4 py-3 flex items-center justify-between border-b border-dark-lighter">
    <div class="flex items-center gap-3">
      <div class="w-10 h-10 rounded-full bg-gradient-to-br from-primary to-secondary flex items-center justify-center font-bold text-dark">
        {{ authStore.user?.name.charAt(0).toUpperCase() }}
      </div>
      <div>
        <p class="font-medium">{{ authStore.user?.name }}</p>
        <p class="text-sm text-gray-400">Player</p>
      </div>
    </div>
    <button
      @click="goToLeaderboard"
      class="bg-dark-light hover:bg-dark-lighter px-4 py-2 rounded-full flex items-center gap-2 transition-colors"
    >
      <span class="text-primary font-bold">{{ authStore.user?.tokens ?? 0 }}</span>
      <span class="text-gray-400 text-sm">tokens</span>
    </button>
  </header>
</template>
