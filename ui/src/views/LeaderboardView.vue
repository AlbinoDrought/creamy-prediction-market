<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useLeaderboardStore } from '@/stores/leaderboard'
import AppHeader from '@/components/AppHeader.vue'
import BottomNav from '@/components/BottomNav.vue'
import LeaderboardEntry from '@/components/LeaderboardEntry.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'

const authStore = useAuthStore()
const leaderboardStore = useLeaderboardStore()

onMounted(async () => {
  await leaderboardStore.fetchLeaderboard()
})

const currentUserRank = computed(() => {
  if (!authStore.user) return null
  return leaderboardStore.leaderboard.find(u => u.id === authStore.user!.id)
})
</script>

<template>
  <div class="min-h-screen bg-bg pb-20">
    <AppHeader />

    <main class="p-4">
      <h1 class="text-2xl font-bold text-white mb-6">Leaderboard</h1>

      <div v-if="leaderboardStore.loading" class="flex justify-center py-12">
        <LoadingSpinner size="lg" />
      </div>

      <div v-else-if="leaderboardStore.error" class="bg-error/20 border border-error rounded-lg px-4 py-3 text-error">
        {{ leaderboardStore.error }}
      </div>

      <template v-else>
        <!-- Current user's rank if not in top -->
        <div v-if="currentUserRank && currentUserRank.rank > 10" class="mb-4">
          <p class="text-sm text-gray-400 mb-2">Your Position</p>
          <LeaderboardEntry
            :user="currentUserRank"
            :is-current-user="true"
          />
        </div>

        <!-- Top users -->
        <div class="space-y-3">
          <LeaderboardEntry
            v-for="user in leaderboardStore.leaderboard"
            :key="user.id"
            :user="user"
            :is-current-user="authStore.user?.id === user.id"
          />
        </div>

        <div v-if="leaderboardStore.leaderboard.length === 0" class="text-center py-12">
          <div class="w-16 h-16 mx-auto mb-4 rounded-full bg-dark-light flex items-center justify-center">
            <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
          </div>
          <p class="text-gray-400">No users yet</p>
        </div>
      </template>
    </main>

    <BottomNav />
  </div>
</template>
