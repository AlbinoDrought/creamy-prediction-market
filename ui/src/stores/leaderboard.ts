import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { LeaderboardUser } from '@/types/users'
import { api } from '@/api/client'

export const useLeaderboardStore = defineStore('leaderboard', () => {
  const leaderboard = ref<LeaderboardUser[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function swapLeaderboard() {
    error.value = null
    try {
      leaderboard.value = await api.getLeaderboard()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load leaderboard'
    }
  }

  async function fetchLeaderboard() {
    loading.value = true
    try {
      await swapLeaderboard()
    } finally {
      loading.value = false;
    }
  }

  return {
    leaderboard,
    loading,
    error,
    swapLeaderboard,
    fetchLeaderboard,
  }
})
